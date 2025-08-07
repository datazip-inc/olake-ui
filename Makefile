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

build:
	gofmt -l -s -w .
	cd server; go build -o olake-server main.go

frontend-format-check:
	cd ui; pnpm run format:check

gofmt:
	gofmt -l -s -w .

pre-commit:
	chmod +x $(shell pwd)/.githooks/pre-commit
	chmod +x $(shell pwd)/.githooks/commit-msg
	git config core.hooksPath $(shell pwd)/.githooks

gosec:
	cd server; $(GOPATH)/bin/gosec -exclude=G115 -severity=high -confidence=medium ./...

trivy:
	trivy fs  --vuln-type  os,library --severity HIGH,CRITICAL .

# Variables
SERVER_DIR := $(PWD)/server
FRONTEND_DIR := $(PWD)/ui

# Backend environment variables
BACKEND_ENV_VARS = \
      APP_NAME=olake-server \
      HTTP_PORT=8000 \
      RUN_MODE=localdev \
      COPY_REQUEST_BODY=true \
      POSTGRES_DB=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable \
      LOGS_DIR=./logger/logs \
      SESSION_ON=true \
      TEMPORAL_ADDRESS=localhost:7233

# Frontend environment variables
FRONTEND_ENV_VARS = \
      VITE_IS_DEV=true

# Start frontend dev server with env vars
start-frontend:
	cd $(FRONTEND_DIR) && \
	  pnpm install && \
	$(FRONTEND_ENV_VARS) pnpm run dev

# Start backend server with env vars
start-backend:
	cd $(SERVER_DIR) && \
	$(BACKEND_ENV_VARS) bee run


# Start Temporal services using Docker Compose
start-temporal:
	cd $(SERVER_DIR) &&  docker compose down && $(BACKEND_ENV_VARS) docker compose up -d

# Start Temporal Go worker
start-temporal-server:
	cd $(SERVER_DIR) && $(BACKEND_ENV_VARS)  go run ./cmd/temporal-worker/main.go

# Create a user with specified username, password and email (e.g. make create-user username=admin password=admin123 email=admin@example.com)
create-user:
	@curl -s -X POST http://localhost:8000/signup -H "Content-Type: application/json" -d "{\"username\":\"$(username)\",\"password\":\"$(password)\",\"email\":\"$(email)\"}" | grep -q "\"success\": true" && echo "User $(username) created successfully" || echo "Failed to create user $(username)"

