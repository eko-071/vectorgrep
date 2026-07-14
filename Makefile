include .env
export

PYTHON_PORT ?= 8001

.PHONY: run-python run-go build-cli build-ingest ingest docker-up docker-down docker-logs

run-python:
	cd python-service && uv run uvicorn main:app --port $(PYTHON_PORT)

run-go:
	cd go-api && go run ./cmd/server

build-cli:
	cd go-api && go build -o vgrep ./cmd/cli

build-ingest:
	cd go-api && go build -o ingest-batch ./cmd/ingest-batch

ingest: build-ingest
	cd go-api && ./ingest-batch commands.txt

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
