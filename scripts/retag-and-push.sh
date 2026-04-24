#!/usr/bin/env bash

set -euo pipefail

SOURCE_TAG="${SOURCE_TAG:-v0.4.1}"
TARGET_TAG="${TARGET_TAG:-v0.4.2}"
LATEST_TAG="${LATEST_TAG:-latest}"

DOCKER_USERNAME="${DOCKER_USERNAME:-${DOCKER_LOGIN:-}}"
DOCKER_PASSWORD="${DOCKER_PASSWORD:-}"
DOCKER_IMAGE="olakego/ui"

if [[ -z "$DOCKER_USERNAME" || -z "$DOCKER_PASSWORD" ]]; then
  echo "DOCKER_USERNAME/DOCKER_LOGIN and DOCKER_PASSWORD are required."
  exit 1
fi

SOURCE_IMAGE="${DOCKER_IMAGE}:${SOURCE_TAG}"
TARGET_IMAGE="${DOCKER_IMAGE}:${TARGET_TAG}"
LATEST_IMAGE="${DOCKER_IMAGE}:${LATEST_TAG}"

echo "Logging in to Docker registry as ${DOCKER_USERNAME}"
echo "${DOCKER_PASSWORD}" | docker login --username "${DOCKER_USERNAME}" --password-stdin

echo "Inspecting source multi-arch manifest ${SOURCE_IMAGE}"
docker buildx imagetools inspect "${SOURCE_IMAGE}" >/dev/null

echo "Copying manifest to ${TARGET_IMAGE} and ${LATEST_IMAGE}"
docker buildx imagetools create \
  --tag "${TARGET_IMAGE}" \
  --tag "${LATEST_IMAGE}" \
  "${SOURCE_IMAGE}"

echo "Verifying target manifests"
docker buildx imagetools inspect "${TARGET_IMAGE}" >/dev/null
docker buildx imagetools inspect "${LATEST_IMAGE}" >/dev/null

echo "Retag and push completed successfully."
