# Build stage
FROM golang:1.24.2-alpine AS builder
WORKDIR /app/worker

# Copy go.mod and go.sum first to leverage Docker caching
COPY server/go.mod server/go.sum ./

RUN go mod download

# Copy the entire server directory (since the worker might depend on shared code)
COPY server/ ./

# Build the worker binary
RUN go build -o temporal-worker ./cmd/temporal-worker

# Runtime stage
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/worker/temporal-worker .
RUN mkdir -p ./conf
COPY server/conf/app.conf ./conf/app.conf
RUN apk update && apk add --no-cache docker-cli
RUN mkdir -p ./logger/logs
RUN mkdir -p /mnt/config && chmod -R 777 /mnt/config
ENV TEMPORAL_ADDRESS="temporal:7233"
CMD ["./temporal-worker"]
