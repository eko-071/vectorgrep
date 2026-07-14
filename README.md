# vectorgrep

Semantic search for Linux man pages. Given a natural language query, returns
the most relevant commands and their documentation excerpts.

## Requirements

- Go 1.26+
- Python 3.12+
- [uv](https://docs.astral.sh/uv/)

```bash
cd python-service && uv sync
```

## Running

The Python service handles embedding and search. The Go server exposes an HTTP
API and proxies requests to Python. The CLI is the user-facing tool.

```bash
# Terminal 1
make run-python

# Terminal 2
make run-go

# Terminal 3
make build-cli
./go-api/vgrep "compress a directory"
./go-api/vgrep -n 3 "find processes by memory"
```

## Docker

Build and run both services with a single command. Your existing `data/` is
bind-mounted in — no re-ingest needed.

```bash
# Build and start
make docker-up

# Check they're healthy
curl localhost:8080/health
curl localhost:8001/health

# Search via Go API
curl -s "localhost:8080/search?q=compress+a+directory&top_k=3"

# Use the CLI from host
GO_API_URL=http://localhost:8080 ./go-api/vgrep "compress a directory"

# View logs
make docker-logs

# Stop
make docker-down
```

## Indexing

`go-api/commands.txt` contains 5,873 commands across sections 1-8.

```bash
make ingest
```

Progress is tracked in `data/.indexed_state`. Interrupted runs resume on the
next `make ingest`. Pass `--force` to re-index everything, or delete `data/`
to reset.

## API

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Service status and vector count |
| `/search?q=<query>&top_k=N` | GET | Semantic search |
| `/ingest` | POST | Index a single command |
| `/embed` | POST | Embed arbitrary text |

## Architecture

Man pages are parsed with `man <command> | col -b`, split into ~400-word
chunks by section header, and embedded with BGE-small-en-v1.5 (384
dimensions). Vectors are stored in a FAISS IndexIDMap backed by JSON metadata.

On search, the query is embedded identically and FAISS returns nearest
neighbors above the configured score threshold. A single query takes roughly
28ms end to end.

Section coverage:
- 2, 4, 5, 7, 8: all entries included
- 1: filtered to remove GUI apps and other non-relevant tools
- 3: curated set of POSIX library functions

Regenerate the command list with `python scripts/generate_commands.py`.

## Project structure

```
├── python-service/
│   ├── Dockerfile
│   ├── .dockerignore
│   ├── main.py
│   ├── embedder.py
│   ├── parser.py
│   ├── vector_store.py
│   └── pyproject.toml
├── go-api/
│   ├── Dockerfile
│   ├── cmd/
│   │   ├── cli/main.go
│   │   ├── ingest-batch/main.go
│   │   └── server/main.go
│   ├── internal/
│   │   ├── embedder/client.go
│   │   └── env/env.go
│   ├── cli                    search binary
│   ├── ingest-batch           batch ingestion binary
│   ├── server                 server binary
│   ├── commands.txt
│   ├── go.mod
│   └── go.sum
├── data/
├── scripts/
│   └── generate_commands.py
├── docker-compose.yml
├── Makefile
├── .env
└── README.md
```
