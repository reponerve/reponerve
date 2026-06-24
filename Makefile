SHELL := /bin/sh

APP_NAME := reponerve
MODULE := github.com/reponerve/reponerve
MAIN := ./cmd/reponerve/main.go
BIN := ./bin/reponerve

GO ?= go
PKGS := ./...

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(BUILD_DATE)

.PHONY: help setup tidy verify fmt vet lint test test-race test-integration build install run clean check module-check release-check release-dry scan context mcp

help:
	@echo "RepoNerve Make Targets"
	@echo ""
	@echo "Required commands"
	@echo "  make setup           - download and verify dependencies"
	@echo "  make test            - run full test suite"
	@echo "  make lint            - run vet and format check"
	@echo "  make build           - build ./bin/reponerve binary"
	@echo "  make install         - install reponerve to \$$(go env GOPATH)/bin"
	@echo ""
	@echo "Common commands"
	@echo "  make run             - run CLI from source"
	@echo "  make scan            - build then run reponerve scan"
	@echo "  make context         - build then run reponerve context generate"
	@echo "  make mcp             - build then run reponerve mcp"
	@echo "  make clean           - remove local binary"
	@echo ""
	@echo "Release checks"
	@echo "  make check           - fmt check + vet + tests + module checks"
	@echo "  make release-check   - check + goreleaser check (if installed)"
	@echo "  make release-dry     - local goreleaser release --snapshot --skip=publish"

setup: tidy
	$(GO) mod download
	$(GO) mod verify

tidy:
	$(GO) mod tidy

verify:
	$(GO) mod verify

fmt:
	$(GO) fmt $(PKGS)

vet:
	$(GO) vet $(PKGS)

lint: fmt vet

module-check:
	$(GO) list $(PKGS) >/dev/null

test:
	$(GO) test $(PKGS)

test-race:
	$(GO) test -race $(PKGS)

test-integration:
	$(GO) test ./tests/integration/...

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) $(MAIN)

install:
	$(GO) install -ldflags "$(LDFLAGS)" $(MODULE)/cmd/reponerve

run:
	$(GO) run $(MAIN)

scan: build
	$(BIN) scan

context: build
	$(BIN) context generate

mcp: build
	$(BIN) mcp

check: fmt vet test verify module-check

release-check: check
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser check; \
	else \
		echo "goreleaser not found on PATH. Install goreleaser to run local release validation."; \
		exit 1; \
	fi

release-dry:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean --snapshot --skip=publish; \
	else \
		echo "goreleaser not found on PATH. Install goreleaser to run local dry release."; \
		exit 1; \
	fi

clean:
	rm -f $(BIN)
