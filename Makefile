include .env
export

PYTHON_PORT ?= 8001

.PHONY: run-python run-go build-cli build-ingest ingest

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
