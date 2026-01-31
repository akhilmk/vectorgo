# Makefile for VectorGo development environment

# Load environment variables from docker/.env.dev if it exists
ifneq (,$(wildcard docker/.env.dev))
    include docker/.env.dev
    export
endif

# Configuration
COMPOSE_FILE ?= docker/compose.dev.yaml
DOCKER_COMPOSE ?= docker compose
IMAGE_NAME := akhilmk01/vectorgo
CONTAINER_NAME := vectorgo
NODE_IMAGE := node:24-alpine
USER_ID := $(shell id -u)
GROUP_ID := $(shell id -g)
COMPOSE_ENV := USER_ID=$(USER_ID) GROUP_ID=$(GROUP_ID)
DOCKER_EXEC_NODE := docker exec -i vectorgo-node-dev

.PHONY: dev-up dev-down dev-logs \
        frontend-install frontend-audit-fix frontend-build build-frontend build-backend build-all \
        docker run logs app-shell \
        clean docker-stop docker-clean help

# --- Development Environment Commands ---

dev-up:
	@echo "Starting VectorGo dev environment (Ollama + ChromaDB + Node builder)..."
	$(COMPOSE_ENV) $(DOCKER_COMPOSE) -f $(COMPOSE_FILE) --env-file docker/.env.dev up -d

dev-down:
	@echo "Stopping VectorGo dev environment..."
	$(COMPOSE_ENV) $(DOCKER_COMPOSE) -f $(COMPOSE_FILE) --env-file docker/.env.dev down -v

dev-logs:
	$(COMPOSE_ENV) $(DOCKER_COMPOSE) -f $(COMPOSE_FILE) --env-file docker/.env.dev logs -f

# --- Build Commands ---

frontend-install:
	@if [ $$(docker ps -q -f name=vectorgo-node-dev) ]; then \
		$(DOCKER_EXEC_NODE) npm install; \
	else \
		docker run --rm -v $(PWD)/frontend:/app -w /app --user $(USER_ID):$(GROUP_ID) $(NODE_IMAGE) npm install; \
	fi

frontend-audit-fix:
	@if [ $$(docker ps -q -f name=vectorgo-node-dev) ]; then \
		$(DOCKER_EXEC_NODE) npm audit fix; \
	else \
		docker run --rm -v $(PWD)/frontend:/app -w /app --user $(USER_ID):$(GROUP_ID) $(NODE_IMAGE) npm audit fix; \
	fi

frontend-build:
	@if [ $$(docker ps -q -f name=vectorgo-node-dev) ]; then \
		$(DOCKER_EXEC_NODE) npm run build; \
	else \
		docker run --rm -v $(PWD)/frontend:/app -w /app --user $(USER_ID):$(GROUP_ID) $(NODE_IMAGE) npm run build; \
	fi

build-frontend: frontend-build
	@echo "Preparing frontend artifacts..."
	mkdir -p bin/frontend
	rm -rf bin/frontend/dist
	cp -r frontend/dist bin/frontend/

build-backend:
	@echo "Building Go backend..."
	cd backend && CGO_ENABLED=0 GOOS=linux go build -o ../bin/vectorgo ./cmd/server
	@echo "✓ Backend built successfully"

build-all: build-backend build-frontend
	@echo "✓ All artifacts built in bin/"

docker: docker-stop docker-clean clean build-all
	@echo "Building Docker image $(IMAGE_NAME):latest..."
	docker build -t $(IMAGE_NAME):latest -f docker/Dockerfile .
	@echo "✓ Docker image built"

# --- Runtime Commands ---

run:
	@echo "Starting $(CONTAINER_NAME) container..."
	-docker rm -f $(CONTAINER_NAME) 2>/dev/null
	docker run -d \
		--name $(CONTAINER_NAME) \
		-p 8080:8080 \
		--network vectorgo-dev \
		-v $(PWD)/frontend/dist:/app/frontend/dist:ro \
		-e OLLAMA_URL=http://vectorgo-ollama:11434 \
		-e CHROMA_URL=http://vectorgo-chromadb:8000 \
		-e EMBEDDING_MODEL=$(EMBEDDING_MODEL) \
		-e COLLECTION_NAME=$(COLLECTION_NAME) \
		-e PORT=8080 \
		$(IMAGE_NAME):latest
	@echo "✓ VectorGo running at http://localhost:8080"

logs:
	docker logs -f $(CONTAINER_NAME)

app-shell:
	docker exec -it $(CONTAINER_NAME) sh

# --- Testing Commands ---

go-test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...

# --- Utility Commands ---

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf frontend/dist

docker-clean:
	@echo "Stopping and removing Docker images/containers..."
	-docker rm $(CONTAINER_NAME) 2>/dev/null
	-docker rmi $(IMAGE_NAME):latest 2>/dev/null

docker-stop:
	@echo "Stopping Docker container..."
	-docker stop $(CONTAINER_NAME) 2>/dev/null

help:
	@echo "VectorGo Development Commands"
	@echo ""
	@echo "Development Environment:"
	@echo "  dev-up          - Start dev environment (Ollama + ChromaDB + Node builder)"
	@echo "  dev-down        - Stop dev environment and remove volumes"
	@echo "  dev-logs        - Follow dev environment logs"
	@echo ""
	@echo "Build Commands:"
	@echo "  docker          - Full clean build and docker image creation"
	@echo "  build-all       - Build backend and frontend locally"
	@echo "  frontend-install - Install frontend dependencies"
	@echo "  frontend-audit-fix - Fix frontend dependency vulnerabilities"
	@echo "  build-frontend  - Build frontend and copy to bin folder"
	@echo "  build-backend   - Build Go backend"
	@echo ""
	@echo "Runtime Commands:"
	@echo "  run             - Run the VectorGo container locally"
	@echo "  logs            - Follow app container logs"
	@echo "  app-shell       - Open shell inside the app container"
	@echo ""
	@echo "Testing Commands:"
	@echo "  go-test         - Run Go backend tests"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean           - Remove local build artifacts"
	@echo "  docker-clean    - Remove app container and image"
	@echo "  docker-stop     - Stop app container"
