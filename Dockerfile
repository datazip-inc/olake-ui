# This image only packages prebuilt artifacts. Build them first with `make dist`:
#   ui/dist                          -> /opt/frontend/dist
#   dist/<os>/<arch>/olake-server    -> /app/olake-ui/olake-server
# No RUN step runs for the target platform, so multi-arch builds need no QEMU.

# docker CLI for the target platform. The server runs `docker images` against the mounted
# /var/run/docker.sock. The official image ships a static binary built with a current Go
# toolchain; Alpine's docker-cli package can lag on Go security fixes.
FROM docker:29-cli AS docker-cli

FROM alpine:3.24

ARG APP_VERSION=unknown
ARG TARGETOS
ARG TARGETARCH

ENV APP_VERSION=${APP_VERSION}

WORKDIR /app/olake-ui

COPY --from=docker-cli /usr/local/bin/docker /usr/local/bin/docker
COPY server/conf/app.yaml ./config/app.yaml
COPY server/internal/services/optimization/resources/spec.json ./server/internal/services/optimization/resources/spec.json
COPY ui/dist /opt/frontend/dist
# --chmod: workflow artifacts drop the exec bit.
COPY --chmod=0755 dist/${TARGETOS}/${TARGETARCH}/olake-server ./olake-server

# Expose the Go backend port (which serves both API and frontend)
EXPOSE 8000

# Runs as root: the server needs the mounted docker.sock.
CMD ["./olake-server"]
