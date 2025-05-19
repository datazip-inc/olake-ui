FROM golang:1.23-alpine

WORKDIR /app/worker

# Install Docker CLI
# hadolint ignore=DL3018
RUN apk update && apk add --no-cache docker-cli

# Copy go.mod and go.sum first to leverage Docker caching
COPY server/go.mod server/go.sum ./

RUN go mod download

# Copy the entire server directory (since the worker might depend on shared code)
COPY server/ ./

# Build the worker binary
RUN go build -o temporal-worker ./cmd/temporal-worker

# Create necessary directories
RUN mkdir -p ./logger/logs

# Environment variables
ENV TEMPORAL_ADDRESS="temporal:7233"

RUN mkdir -p /mnt/config
RUN chmod -R 777 /mnt/config

# Run the worker
CMD ["./temporal-worker"]
