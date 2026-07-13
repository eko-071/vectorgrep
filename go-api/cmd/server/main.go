package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/eko-071/vectorgrep/internal/embedder"
	"github.com/eko-071/vectorgrep/internal/env"
	"github.com/gin-gonic/gin"
)

func main() {
	port := env.GetEnv("GO_PORT", "8080")
	embedServiceURL := env.GetEnv("EMBEDDING_SERVICE_URL", "http://localhost:8001")

	timeoutSec, _ := strconv.Atoi(env.GetEnv("HTTP_TIMEOUT_SEC", "30"))
	embedClient := embedder.NewClient(embedServiceURL, time.Duration(timeoutSec)*time.Second)
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
			sendError(c, err, "ingestion failed", "command", req.Command)
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

		topK, _ := strconv.Atoi(c.DefaultQuery("top_k", "5"))
		scoreThreshold, _ := strconv.ParseFloat(c.DefaultQuery("score_threshold", "0"), 64)
		result, err := embedClient.Search(query, topK, scoreThreshold)
		if err != nil {
			sendError(c, err, "search failed", "query", query)
			return
		}

		c.JSON(http.StatusOK, result)
	})

	if err := waitForPython(embedServiceURL, 15); err != nil {
		slog.Error("could not reach python service", "error", err)
		os.Exit(1)
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
	}
}

func sendError(c *gin.Context, err error, msg string, extra ...any) {
	var statusErr *embedder.StatusError
	if errors.As(err, &statusErr) {
		args := append([]any{"error", statusErr.Message, "status", statusErr.StatusCode}, extra...)
		slog.Error(msg, args...)
		c.JSON(statusErr.StatusCode, gin.H{"error": statusErr.Message})
		return
	}
	args := append([]any{"error", err}, extra...)
	slog.Error(msg, args...)
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
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
