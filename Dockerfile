# Stage 1: Go Builder (Backend)
FROM golang:1.24.2-alpine AS go-builder

# Install git, as it might be needed by go mod download or go build
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum for the entire server project
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy the entire server source code
COPY server/ ./server/

# Build backend (assuming main package is at the root of 'server/' content)
# Using -ldflags to create smaller binaries
RUN cd server && go build -ldflags="-w -s" -o /app/olake-server .

# Stage 2: Frontend Builder
FROM node:20-alpine AS node-builder
WORKDIR /app/ui

# Install pnpm globally
RUN npm install -g pnpm

# Copy package files
COPY ui/package.json ui/pnpm-lock.yaml ./
# Install dependencies
RUN pnpm install

# Copy the rest of the UI code
COPY ui/ ./
# Build the UI
RUN pnpm build

# Stage 3: Final Runtime Image
FROM alpine:latest

# Install docker-cli
RUN apk update && apk add --no-cache docker-cli

# Set working directory
WORKDIR /app/olake-ui

# Create directories for applications
RUN mkdir -p conf /opt/frontend/dist

# Copy built artifacts from builder stages
COPY --from=go-builder /app/olake-server ./olake-server
COPY server/conf/app.conf ./conf/app.conf
COPY --from=node-builder /app/ui/dist /opt/frontend/dist

# Expose the Go backend port (which serves both API and frontend)
EXPOSE 8000

# Run the olake-ui app
CMD ["./olake-server"]
