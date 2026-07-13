package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eko-071/vectorgrep/internal/embedder"
	"github.com/eko-071/vectorgrep/internal/env"
)

const maxTextLength = 200

func main() {
	n := flag.Int("n", 5, "number of results")
	s := flag.Float64("s", 0, "score threshold (default: SCORE_THRESHOLD env var)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: vgrep [-n N] [-s S] <query>\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	query := strings.Join(flag.Args(), " ")
	apiURL := env.GetEnv("GO_API_URL", "http://localhost:8080")

	timeoutSec, _ := strconv.Atoi(env.GetEnv("HTTP_TIMEOUT_SEC", "15"))
	client := embedder.NewClient(apiURL, time.Duration(timeoutSec)*time.Second)
	result, err := client.Search(query, *n, *s)
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


