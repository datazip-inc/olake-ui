FROM node:20-alpine AS frontend-builder

WORKDIR /app/ui

RUN npm install -g pnpm

COPY ui/package.json ui/pnpm-lock.yaml ./

RUN pnpm install

COPY ui/ ./

RUN pnpm run build


FROM golang:1.23-alpine AS backend-builder

WORKDIR /app/server

COPY server/go.mod server/go.sum ./

RUN go mod download

COPY server/ ./

RUN go build -o olake-server .

######################################################################
# Final image
######################################################################
FROM alpine:latest

ARG ENVIRONMENT
ARG APP_VERSION

WORKDIR /app

RUN apk add --no-cache nodejs npm

# Create directories
RUN mkdir -p /app/ui/dist /app/server/logger/logs

# Copy UI build from frontend-builder
COPY --from=frontend-builder /app/ui/dist /app/ui/dist

# Copy server binary from backend-builder
COPY --from=backend-builder /app/server/olake-server /app/server/
COPY --from=backend-builder /app/server/conf /app/server/conf
COPY --from=backend-builder /app/server/views /app/server/views

# Create startup script
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'cd /app/server && ./olake-server & cd /app/ui && exec node /app/ui/server.js' >> /app/start.sh && \
    chmod +x /app/start.sh

EXPOSE 5173 8080

# Set environment variables
ENV ENVIRONMENT=${ENVIRONMENT} \
    APP_VERSION=${APP_VERSION}

# Run both server and UI
CMD ["/app/start.sh"] 