# This image only packages prebuilt artifacts. Build them first with `make dist`:
#   ui/dist                          -> /opt/frontend/dist
#   dist/<os>/<arch>/olake-server    -> /app/olake-ui/olake-server
# Every RUN step runs on the build platform, so multi-arch builds need no QEMU.

# docker CLI for the target platform. The server runs `docker images` against the mounted
# /var/run/docker.sock. The official image ships a static binary built with a current Go
# toolchain; Alpine's docker-cli package can lag on Go security fixes.
FROM docker:29-cli AS docker-cli

# The official binary keeps debug info (~14 MB). Strip it on the build platform with
# binutils for the target arch, so this also runs natively (no QEMU).
FROM --platform=$BUILDPLATFORM alpine:3.24 AS docker-cli-strip
ARG TARGETARCH
ARG BUILDARCH
COPY --from=docker-cli /usr/local/bin/docker /docker
RUN set -eux; \
    case "$TARGETARCH" in \
        amd64) arch=x86_64 ;; \
        arm64) arch=aarch64 ;; \
        *) echo "unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    if [ "$TARGETARCH" = "$BUILDARCH" ]; then \
        apk add --no-cache binutils; strip_bin="strip"; \
    else \
        apk add --no-cache "binutils-$arch"; strip_bin="$arch-alpine-linux-musl-strip"; \
    fi; \
    "$strip_bin" --strip-all -o /docker-stripped /docker

FROM alpine:3.24

ARG APP_VERSION=unknown
ARG TARGETOS
ARG TARGETARCH

ENV APP_VERSION=${APP_VERSION}

WORKDIR /app/olake-ui

COPY --from=docker-cli-strip /docker-stripped /usr/local/bin/docker
COPY server/conf/app.yaml ./config/app.yaml
COPY server/internal/services/optimization/resources/spec.json ./server/internal/services/optimization/resources/spec.json
COPY ui/dist /opt/frontend/dist
# --chmod: workflow artifacts drop the exec bit.
COPY --chmod=0755 dist/${TARGETOS}/${TARGETARCH}/olake-server ./olake-server

# Expose the Go backend port (which serves both API and frontend)
EXPOSE 8000

# Runs as root: the server needs the mounted docker.sock.
CMD ["./olake-server"]
