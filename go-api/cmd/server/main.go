package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/eko-071/vectorgrep/internal/embedder"
	"github.com/gin-gonic/gin"
)

func main() {
	port := getEnv("GO_PORT", "8080")
	embedServiceURL := getEnv("EMBEDDING_SERVICE_URL", "http://localhost:8001")

	embedClient := embedder.NewClient(embedServiceURL)
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/ingest", func(c *gin.Context) {
		var req struct {
			Command string `json:"command"`
		}
		if err := c.BindJSON(&req); err != nil || req.Command == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid 'command' field"})
			return
		}

		result, err := embedClient.Ingest(req.Command)
		if err != nil {
			var statusErr *embedder.StatusError
			if errors.As(err, &statusErr) {
				slog.Error("ingestion failed", "error", statusErr.Message, "status", statusErr.StatusCode, "command", req.Command)
				c.JSON(statusErr.StatusCode, gin.H{"error": statusErr.Message})
				return
			}
			slog.Error("ingestion failed", "error", err, "command", req.Command)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ingestion failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing query param 'q'"})
			return
		}

		result, err := embedClient.Search(query, 5)
		if err != nil {
			var statusErr *embedder.StatusError
			if errors.As(err, &statusErr) {
				slog.Error("search failed", "error", statusErr.Message, "status", statusErr.StatusCode, "query", query)
				c.JSON(statusErr.StatusCode, gin.H{"error": statusErr.Message})
				return
			}
			slog.Error("search failed", "error", err, "query", query)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	if err := waitForPython(embedServiceURL, 15); err != nil {
		slog.Error("could not reach python service", "error", err)
		os.Exit(1)
	}

	slog.Info("starting server", "port", port)
	r.Run(":" + port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func waitForPython(url string, maxAttempts int) error {
	client := &http.Client{Timeout: 2 * time.Second}
	for i := range maxAttempts {
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			slog.Info("python service ready")
			return nil
		}
		slog.Info("waiting for python service...", "attempt", i+1, "max", maxAttempts)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("python service not ready after %d attempts", maxAttempts)
}
