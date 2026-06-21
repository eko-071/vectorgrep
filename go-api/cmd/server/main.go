package main

import (
	"log/slog"
	"net/http"
	"os"

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
			slog.Error("search failed", "error", err, "query", query)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	slog.Info("starting server", "port", port)
	r.Run(":" + port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
