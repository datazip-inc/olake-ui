FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies and tidy modules
RUN go mod download && go mod tidy

# Copy source code
COPY . .

# Build the worker binary
RUN go build -o olake-workers ./main.go

# Final stage - minimal runtime image
FROM alpine:3.18

# Install only timezone data (minimal)
RUN apk --no-cache add tzdata

# Create non-root user
RUN addgroup -g 1001 -S olake && \
    adduser -u 1001 -S olake -G olake

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/olake-workers .

# Change ownership to non-root user
RUN chown -R olake:olake /app

# Expose health check port
EXPOSE 8090

# No health check (let Kubernetes handle it via HTTP probes)
CMD ["./olake-workers"]
