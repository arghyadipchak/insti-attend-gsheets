# ==============================================================================
# Configuration
# ==============================================================================

SHELL := /usr/bin/env bash

GO            := go
BINARY        := insti-attend-gsheets
BUILD_DIR     := bin
BIN           := $(BUILD_DIR)/$(BINARY)

PKG           := .
PKGS          := ./...

PORT          ?= 8090
HOST          ?= 0.0.0.0
DATA_DIR      ?= ./data

ENV_FILE      ?= .env

IMAGE         ?= arghyadipchak/insti-attend-gsheets:dev

VERSION       ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS       := -ldflags="-X github.com/pocketbase/pocketbase.Version=$(VERSION)"
RELEASE_FLAGS := -trimpath -ldflags="-s -w -X github.com/pocketbase/pocketbase.Version=$(VERSION)"

# ==============================================================================
# Reusable Snippets
# ==============================================================================

LOAD_ENV = set -a; [[ -f "$(ENV_FILE)" ]] && source "$(ENV_FILE)"; set +a;

# ==============================================================================
# Help & Default
# ==============================================================================

.PHONY: all help
all: build

help:
	@echo ""
	@printf "\033[1mUsage:\033[0m make \033[36m<target>\033[0m\n\n"
	@printf "\033[1;32m🚀 Development & Execution\033[0m\n"
	@printf "  \033[36m%-12s\033[0m %s\n" "run" "Run application using go run with .env"
	@printf "  \033[36m%-12s\033[0m %s\n" "dev" "Start development server with local data"
	@printf "  \033[36m%-12s\033[0m %s\n" "serve" "Build binary and start production server"
	@echo ""
	@printf "\033[1;32m🛠  Code Quality & Testing\033[0m\n"
	@printf "  \033[36m%-12s\033[0m %s\n" "fmt" "Format code with gofmt -s"
	@printf "  \033[36m%-12s\033[0m %s\n" "lint" "Run golangci-lint linter"
	@printf "  \033[36m%-12s\033[0m %s\n" "lint-fix" "Run golangci-lint with auto-fix"
	@printf "  \033[36m%-12s\033[0m %s\n" "fix" "Format code and auto-fix linter issues"
	@printf "  \033[36m%-12s\033[0m %s\n" "test" "Execute unit tests"
	@printf "  \033[36m%-12s\033[0m %s\n" "check" "Full verification (fmt + lint + test)"
	@echo ""
	@printf "\033[1;32m📦 Build & Packaging\033[0m\n"
	@printf "  \033[36m%-12s\033[0m %s\n" "build" "Build application binary in bin/"
	@printf "  \033[36m%-12s\033[0m %s\n" "release" "Build optimized, stripped release binary"
	@printf "  \033[36m%-12s\033[0m %s\n" "docker" "Build development Docker image (:dev)"
	@echo ""
	@printf "\033[1;32m🧹 Maintenance\033[0m\n"
	@printf "  \033[36m%-12s\033[0m %s\n" "deps" "Download Go dependencies"
	@printf "  \033[36m%-12s\033[0m %s\n" "clean" "Remove build directory and binary artifacts"
	@echo ""

# ==============================================================================
# Development & Execution
# ==============================================================================

.PHONY: run dev serve

run:
	@$(LOAD_ENV) $(GO) run . serve --http $(HOST):$(PORT) --dir $(DATA_DIR)

dev:
	@$(LOAD_ENV) $(GO) run . serve --dir $(DATA_DIR)

serve: build
	@$(LOAD_ENV) $(BIN) serve --http $(HOST):$(PORT) --dir $(DATA_DIR)

# ==============================================================================
# Code Quality & Testing
# ==============================================================================

.PHONY: fmt lint lint-fix fix test check

fmt:
	gofmt -s -w .

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

fix: fmt lint-fix

test:
	$(GO) test -v $(PKGS)

check: fmt lint test

# ==============================================================================
# Build & Packaging
# ==============================================================================

.PHONY: build release docker

$(BUILD_DIR):
	mkdir -p $@

build: $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN) $(PKG)

release: $(BUILD_DIR)
	$(GO) build $(RELEASE_FLAGS) -o $(BIN) $(PKG)

docker:
	docker build -t $(IMAGE) .

# ==============================================================================
# Maintenance
# ==============================================================================

.PHONY: deps clean

deps:
	$(GO) mod download

clean:
	$(GO) clean
	rm -rf $(BUILD_DIR)
