whoami=$(shell whoami)
home=$(shell echo $$HOME)
GIT_VERSION=$(shell git describe --tags `git rev-list --tags --max-count=1`)
GIT_COMMITSHA=$(shell git rev-list -1 HEAD)
LDFLAGS="-w -s -X github.com/datazip/olake-server/constants.version=${GIT_VERSION} -X github.com/datazip/olake-ui/constants.commitsha=${GIT_COMMITSHA} -X github.com/datazip/olake-ui/constants.releasechannel=${RELEASE_CHANNEL}"
GOPATH = $(shell go env GOPATH)


## Lint check.
golangci:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest;
	cd server; $(GOPATH)/bin/golangci-lint run

frontend-lint:
	cd ui; pnpm run lint

frontend-lint-fix:
	cd ui; pnpm run lint:fix

frontend-format:
	cd ui; pnpm run format

frontend-format-check:
	cd ui; pnpm run format:check

build:
	gofmt -l -s -w .
	cd server; go build -ldflags=${LDFLAGS} -o olake-server main.go

gofmt:
	gofmt -l -s -w .
	
run:
	cd server; go mod tidy; \
	bee run;

run-build:
	./olake-server

restart: build run-build

pre-commit:
	chmod +x $(shell pwd)/.githooks/pre-commit
	chmod +x $(shell pwd)/.githooks/commit-msg
	git config core.hooksPath $(shell pwd)/.githooks

gosec:
	cd server; $(GOPATH)/bin/gosec -exclude=G115 -severity=high -confidence=medium ./...

trivy:
	trivy fs  --vuln-type  os,library --severity HIGH,CRITICAL .

# Create a user with specified username, password and email (e.g. make create-user username=admin password=admin123 email=admin@example.com)
create-user:
	@curl -s -X POST http://localhost:8000/signup -H "Content-Type: application/json" -d "{\"username\":\"$(username)\",\"password\":\"$(password)\",\"email\":\"$(email)\"}" | grep -q "\"success\": true" && echo "User $(username) created successfully" || echo "Failed to create user $(username)"
	
# Build, start server, and create frontend user in one command
setup: build pre-commit
	@echo "Starting server and setting up frontend user..."
	@$(MAKE) run
	@sleep 5
	@$(MAKE) create-user

# Variables
SERVER_DIR := $(PWD)/server
FRONTEND_DIR := $(PWD)/ui

# Create or update .env file in the frontend
create-ui-env:
	@echo "Ensuring .env exists in $(FRONTEND_DIR)..."
	@mkdir -p $(FRONTEND_DIR)
	@echo "VITE_IS_DEV=true" > $(FRONTEND_DIR)/.env
	@echo ".env created/updated in $(FRONTEND_DIR)"

# Create or update app.conf file in the backend
create-backend-conf:
	@echo "Ensuring app.conf exists in $(SERVER_DIR)/conf..."
	@mkdir -p $(SERVER_DIR)/conf
	@echo "appname = olake-server" > $(SERVER_DIR)/conf/app.conf
	@echo "httpport = 8000" >> $(SERVER_DIR)/conf/app.conf
	@echo "runmode = dev" >> $(SERVER_DIR)/conf/app.conf
	@echo "copyrequestbody = true" >> $(SERVER_DIR)/conf/app.conf
	@echo "postgresdb = postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" >> $(SERVER_DIR)/conf/app.conf
	@echo "logsdir = ./logger/logs" >> $(SERVER_DIR)/conf/app.conf
	@echo "sessionon = true" >> $(SERVER_DIR)/conf/app.conf
	@echo "TEMPORAL_ADDRESS=localhost:7233" >> $(SERVER_DIR)/conf/app.conf
	@echo "app.conf created/updated in $(SERVER_DIR)/conf"

# Initialize frontend and backend configuration to default values
init-config: create-ui-env create-backend-conf
	@echo "Frontend .env and Backend app.conf initialized."

# Start Temporal services using Docker Compose
start-temporal:
	cd $(SERVER_DIR) && docker compose down && docker compose up -d

# Start Temporal Go worker
start-temporal-server:
	cd $(SERVER_DIR) && go run ./cmd/temporal-worker/main.go

# Start backend server with live reload (dev mode)
start-backend:
	cd $(SERVER_DIR) && bee run

# Start frontend dev server
start-frontend:
	cd $(FRONTEND_DIR) && pnpm install && pnpm run dev





