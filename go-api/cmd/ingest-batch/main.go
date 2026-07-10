package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eko-071/vectorgrep/internal/embedder"
	"github.com/eko-071/vectorgrep/internal/env"
)

const stateFilePath = "../data/.indexed_state"
const indexPath = "../data/index.faiss"

func main() {
	force := flag.Bool("force", false, "re-ingest all commands even if already indexed")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: ingest-batch [--force] <commands-file>")
		os.Exit(1)
	}

	commands, err := readCommands(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read commands file: %v\n", err)
		os.Exit(1)
	}

	state := readState()
	if *force {
		state = map[string]bool{}
		os.Remove(stateFilePath)
	}

	apiURL := env.GetEnv("GO_API_URL", "http://localhost:8080")
	client := embedder.NewClient(apiURL, 120*time.Second)

	var succeeded, failed int
	for _, cmd := range commands {
		if state[cmd] {
			fmt.Printf("SKIP %s\n", cmd)
			continue
		}

		result, err := client.Ingest(cmd)
		if err != nil {
			fmt.Printf("FAIL %s - %v\n", cmd, err)
			failed++
			continue
		}
		fmt.Printf("PASS %s - %d chunks\n", cmd, result.ChunksIndexed)
		succeeded++
		appendState(cmd)
	}

	fmt.Printf("\ndone: %d succeeded, %d failed\n", succeeded, failed)
}

func readState() map[string]bool {
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		os.Remove(stateFilePath)
		return map[string]bool{}
	}

	f, err := os.Open(stateFilePath)
	if err != nil {
		return map[string]bool{}
	}
	defer f.Close()

	state := map[string]bool{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		state[strings.TrimSpace(scanner.Text())] = true
	}
	return state
}

func appendState(command string) {
	f, err := os.OpenFile(stateFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintln(f, command)
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


