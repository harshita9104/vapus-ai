# Docs Service Makefile

.PHONY: help build run stop clean install dev

# Variables
CONTAINER_NAME = vapusai-docs
IMAGE_NAME = vapusai/docs
PORT = 3000

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	npm ci

dev: ## Start development server
	npm run start

build: ## Build the Docker image
	docker build -t $(IMAGE_NAME) .

run: ## Run the container
	docker run --rm -d --name $(CONTAINER_NAME) -p $(PORT):3000 $(IMAGE_NAME)

stop: ## Stop the container
	docker stop $(CONTAINER_NAME) || true

clean: ## Clean up containers and images
	docker stop $(CONTAINER_NAME) || true
	docker rm $(CONTAINER_NAME) || true
	docker rmi $(IMAGE_NAME) || true

build-docs: ## Build documentation
	npm run build

serve-local: ## Serve built documentation locally
	npm run serve

clear: ## Clear Docusaurus cache
	npm run clear
