package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eko-071/vectorgrep/internal/embedder"
)

const maxTextLength = 200

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: vgrep <query>")
		os.Exit(1)
	}

	query := strings.Join(os.Args[1:], " ")
	apiURL := getEnv("GO_API_URL", "http://localhost:8080")

	client := embedder.NewClient(apiURL)
	result, err := client.Search(query, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "search failed: %v\n", err)
		os.Exit(1)
	}

	if len(result.Results) == 0 {
		fmt.Println("no matches found")
		return
	}

	fmt.Printf("\n query: %s\n\n", result.Query)
	for i, r := range result.Results {
		command, _ := r.Metadata["command"].(string)
		section, _ := r.Metadata["section"].(string)

		fmt.Printf("%d. %s (%s) — score: %.2f\n", i+1, command, section, r.Score)
		fmt.Printf("%s\n\n", truncate(r.Text, maxTextLength))
	}
}

func truncate(text string, max int) string {
	cleaned := strings.ReplaceAll(text, "\n", " ")
	if len(cleaned) <= max {
		return cleaned
	}
	return cleaned[:max] + "..."
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
