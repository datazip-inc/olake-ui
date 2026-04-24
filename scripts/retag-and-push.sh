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

echo "Pulling source image ${SOURCE_IMAGE}"
docker pull "${SOURCE_IMAGE}"

echo "Tagging ${SOURCE_IMAGE} as ${TARGET_IMAGE}"
docker tag "${SOURCE_IMAGE}" "${TARGET_IMAGE}"

echo "Tagging ${SOURCE_IMAGE} as ${LATEST_IMAGE}"
docker tag "${SOURCE_IMAGE}" "${LATEST_IMAGE}"

echo "Pushing ${TARGET_IMAGE}"
docker push "${TARGET_IMAGE}"

echo "Pushing ${LATEST_IMAGE}"
docker push "${LATEST_IMAGE}"

echo "Retag and push completed successfully."
