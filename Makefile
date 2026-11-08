# The ONLY verbs this project uses. Do not add more without an ADR.
# Every target must work with just `go` on PATH (no network at run time).

BINARY      := jobcrawler
BINARY_ARM  := jobcrawler-arm64
BIN_DIR     := bin
VERSION     = $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT      = $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE        = $(shell date -u +%Y-%m-%d)
LDFLAGS     = -s -w -X github.com/Esemby/JobCrawler/internal/version.Version=$(VERSION) -X github.com/Esemby/JobCrawler/internal/version.Commit=$(COMMIT) -X github.com/Esemby/JobCrawler/internal/version.Date=$(DATE)

.PHONY: build test lint run build-arm64 check clean

build:            ## compile for the dev machine
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/jobcrawler

test:             ## run all tests
	go test ./...

lint:             ## vet + staticcheck (if installed) + fmt check
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipped"
	@out=$$(gofmt -l . ); if [ -n "$$out" ]; then echo "NOT GOFMT: $$out"; exit 1; fi

run:              ## run locally on the dev machine
	go run ./cmd/jobcrawler

build-arm64:      ## cross-compile for the Heart (Raspberry Pi 4B)
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_ARM) ./cmd/jobcrawler
	@file $(BIN_DIR)/$(BINARY_ARM) | grep -q aarch && echo "OK: arm64 binary in $(BIN_DIR)/"

check: lint test  ## THE gate. Green before every commit. No exceptions.

clean:
	rm -rf $(BIN_DIR)
