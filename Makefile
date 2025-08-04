# Makefile
include setup.mk
VERSION=$(shell git describe --tags --abbrev=0)

.PHONY: api
# generate apis by compiling proto files and generating relevant golang/python code
api:
	make -C ./apis api

.PHONY: build

.PHONY: build_dev
# build for development 
build-dev:
	make -C ./app/src/aigateway build
	make -C ./app/src/aistudio build
	make -C ./app/src/cli build
	make -C ./app/src/webapp build
	make -C ./app/src/aitools build
.PHONY: release_dev
# release worker images for development testing
release-dev:
	make -C ./app/src/cli releaseoci
	make -C ./app/src/aigateway release
	make -C ./app/src/webapp release
	make -C ./app/src/aistudio release
	make -C ./app/src/webapp release
# show help for managing the project in your local environment
help:
	@echo ''
	@echo "\033[1mUsage:\033[0m"
	@echo ' make [target]'
	@echo ''
	@echo "\033[1mTargets:\033[0m"
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
	@echo ''
	@echo "\033[1mExample:\033[0m"
	@echo '- make init-dev user=kshitij password=secret db=analytics'

.DEFAULT_GOAL := help

.PHONY: check-golangci-lint-tool
check-golangci-lint-tool:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint is not installed. Please run \"make init\" or install the tool manually."; \
		exit 1; \
	fi

.PHONY: check-buf-tool
check-buf-tool:
	@if ! command -v buf >/dev/null 2>&1; then \
		echo "buf is not installed. Please run \"make init\" or install the tool manually."; \
		exit 1; \
	fi

.PHONY: check-goimports-tool
check-goimports-tool:
	@if ! command -v goimports >/dev/null 2>&1; then \
		echo "goimports is not installed. Please run \"make init\" or install the tool manually."; \
		exit 1; \
	fi
.PHONY: check-poetry-tool
check-poetry-tool:
	@if ! command -v poetry >/dev/null 2>&1; then \
		echo "poetry is not installed. Please install poetry (https://python-poetry.org/docs/#installation)"; \
		exit 1; \
	fi

.PHONY: init-dev
# starts a TimescaleDB container with the given credentials and database
init-dev:
	@if [ -z "$(user)" ] || [ -z "$(password)" ] || [ -z "$(db)" ]; then \
		echo "\033[1;31mError:\033[0m Missing required variables."; \
		echo "\033[1mUsage:\033[0m make init-dev user=<username> password=<password> db=<database>"; \
		exit 1; \
	fi
	@echo "\033[1;32m🚀 Starting TimescaleDB container...\033[0m"
	sudo docker run -d \
		--name timescaledb \
		-p 5432:5432 \
		-e POSTGRES_USER=$(user) \
		-e POSTGRES_PASSWORD=$(password) \
		-e POSTGRES_DB=$(db) \
		timescale/timescaledb:latest-pg17
	@echo "\033[1;32m✅ Container started with user '\033[36m$(user)\033[32m' and database '\033[36m$(db)\033[0m'"