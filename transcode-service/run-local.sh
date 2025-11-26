#!/usr/bin/env bash
set -euo pipefail

# Build and run transcode-service in Docker for local testing
# Usage: bash transcode-service/run-local.sh

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
IMAGE_TAG="transcode-service:local"
CONTAINER_NAME="transcode-service-local"

echo "[local] Stopping/removing old container (if any) ..."
docker rm -f ${CONTAINER_NAME} >/dev/null 2>&1 || true

echo "[local] Building image ${IMAGE_TAG} ..."
docker build -t ${IMAGE_TAG} -f "${SCRIPT_DIR}/Dockerfile" "${ROOT_DIR}"

echo "[local] Starting ${CONTAINER_NAME} ..."
docker run -d \
  --name ${CONTAINER_NAME} \
  -p 8083:8083 \
  -p 9092:9092 \
  -e CONFIG_PATH=/app/configs/config.dev.yaml \
  --add-host=host.docker.internal:host-gateway \
  -v "${SCRIPT_DIR}/configs":/app/configs:ro \
  ${IMAGE_TAG}

echo "[local] Transcode service running:"
echo "  HTTP:   http://localhost:8083/health"
echo "  gRPC:   localhost:9092"
