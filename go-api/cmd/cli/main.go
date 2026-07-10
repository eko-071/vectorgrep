package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eko-071/vectorgrep/internal/embedder"
	"github.com/eko-071/vectorgrep/internal/env"
)

const maxTextLength = 200

func main() {
	n := flag.Int("n", 5, "number of results")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: vgrep [-n N] <query>\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	query := strings.Join(flag.Args(), " ")
	apiURL := env.GetEnv("GO_API_URL", "http://localhost:8080")

	client := embedder.NewClient(apiURL, 15*time.Second)
	result, err := client.Search(query, *n)
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


