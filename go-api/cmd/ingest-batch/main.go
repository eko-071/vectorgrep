package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/eko-071/vectorgrep/internal/embedder"
	"github.com/eko-071/vectorgrep/internal/env"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: ingest-batch <commands-file>")
		os.Exit(1)
	}

	commands, err := readCommands(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read commands file: %v\n", err)
		os.Exit(1)
	}

	apiURL := env.GetEnv("GO_API_URL", "http://localhost:8080")
	client := embedder.NewClient(apiURL)

	var succeeded, failed int
	for _, cmd := range commands {
		result, err := client.Ingest(cmd)
		if err != nil {
			fmt.Printf("FAIL %s - %v\n", cmd, err)
			failed++
			continue
		}
		if result.Skipped {
			fmt.Printf("SKIP %s - already indexed (%d chunks)\n", cmd, result.ChunksIndexed)
		} else {
			fmt.Printf("PASS %s - %d chunks\n", cmd, result.ChunksIndexed)
		}
		succeeded++
	}

	fmt.Printf("\ndone: %d succeeded, %d failed\n", succeeded, failed)
}

func readCommands(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		commands = append(commands, line)
	}
	return commands, scanner.Err()
}


