FROM golang:1.21-alpine AS builder

# Install git and ca-certificates for building
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the worker binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o olake-k8s-worker \
    ./main.go

# Final stage - minimal runtime image
FROM alpine:3.18

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S olake && \
    adduser -u 1001 -S olake -G olake

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/olake-k8s-worker .

# Change ownership to non-root user
RUN chown -R olake:olake /app

# Switch to non-root user
USER 1001

# Expose health check port
EXPOSE 8090

# Set default environment variables
ENV TEMPORAL_ADDRESS="temporal.default.svc.cluster.local:7233"
ENV LOG_LEVEL="info"
ENV HEALTH_CHECK_PORT="8090"

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Run the worker
CMD ["./olake-k8s-worker"]
