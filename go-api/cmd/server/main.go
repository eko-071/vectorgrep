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

	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing query param 'q'"})
			return
		}

		vector, err := embedClient.Embed(query)
		if err != nil {
			slog.Error("embedding failed", "error", err, "query", query)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "embedding service unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"query":          query,
			"embedding_size": len(vector),
		})
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
