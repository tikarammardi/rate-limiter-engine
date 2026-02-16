.PHONY: all build run test lint generate ensure-tools clean \
        docker-build docker-run docker-up docker-down help

# Application
APP_NAME   := guard
CMD_PATH   := ./cmd/guard
BIN_DIR    := bin
BIN        := $(BIN_DIR)/$(APP_NAME)

# Docker
IMAGE_NAME := rguard
COMPOSE    := docker compose

# Proto generation
PROTOC       ?= protoc
PROTO_DIRS   ?= proto
PROTO_FILES  := $(shell find $(PROTO_DIRS) -name '*.proto')
GO_OUT       ?= .
GO_GRPC_OUT  ?= .

# Default target
all: build

##@ Build & Run

build: ## Build the guard binary locally
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BIN) $(CMD_PATH)

run: build ## Build and run guard locally
	$(BIN)


lint: ## Run golangci-lint (install if missing)
	@command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

##@ Proto Generation

generate: ensure-tools ## Generate protobuf Go code
	$(PROTOC) --go_out=$(GO_OUT) --go-grpc_out=$(GO_GRPC_OUT) $(PROTO_FILES)

ensure-tools:
	@command -v $(PROTOC) >/dev/null 2>&1 || { echo "$(PROTOC) not found. Please install protoc."; exit 1; }
	@command -v protoc-gen-go >/dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

##@ Docker

docker-build: ## Build Docker image
	docker build -t $(IMAGE_NAME) .

docker-run: docker-build ## Run Docker container (standalone, uses host Redis)
	docker run --rm -p 50051:50051 $(IMAGE_NAME)

docker-up: ## Start all services with docker compose
	$(COMPOSE) up --build -d

docker-down: ## Stop and remove docker compose services
	$(COMPOSE) down

docker-logs: ## Tail logs from docker compose services
	$(COMPOSE) logs -f

##@ Cleanup

clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)

##@ Help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)
