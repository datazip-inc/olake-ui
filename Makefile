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
	@curl -s -X POST http://localhost:8080/signup -H "Content-Type: application/json" -d "{\"username\":\"$(username)\",\"password\":\"$(password)\",\"email\":\"$(email)\"}" | grep -q "username" && echo "User $(username) created successfully" || echo "Failed to create user $(username)"

# Build, start server, and create frontend user in one command
setup: build pre-commit
	@echo "Starting server and setting up frontend user..."
	@$(MAKE) run
	@sleep 5
	@$(MAKE) create-user

clean-docker-server:
	docker compose down && \
	cd server && \
	docker compose down -v && \
	docker compose up -d

run-temporal-worker:
	cd server && \
	go mod tidy && \
	go run $(PWD)/server/cmd/temporal-worker/main.go

run-backend:
	cd server && \
	go mod tidy && \
	go run $(PWD)/server/main.go

run-frontend:
	cd ui && \
	pnpm i && \
	pnpm run dev