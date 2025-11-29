#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEMPLATE="${SCRIPT_DIR}/kong.yml.tmpl"
OUTPUT="${SCRIPT_DIR}/kong.yml"
KONG_ENV="${KONG_ENV:-dev}"

# Load env: explicit ENV_FILE > .env.<env> > .env.local > .env
if [ -n "${ENV_FILE:-}" ] && [ -f "$ENV_FILE" ]; then
  set -a; . "$ENV_FILE"; set +a
elif [ -f "${SCRIPT_DIR}/.env.${KONG_ENV}" ]; then
  set -a; . "${SCRIPT_DIR}/.env.${KONG_ENV}"; set +a
elif [ -f "${SCRIPT_DIR}/.env.local" ]; then
  set -a; . "${SCRIPT_DIR}/.env.local"; set +a
elif [ -f "${SCRIPT_DIR}/.env" ]; then
  set -a; . "${SCRIPT_DIR}/.env"; set +a
fi

if ! command -v envsubst >/dev/null 2>&1; then
  echo "envsubst is required (install gettext)." >&2
  exit 1
fi

if [ "${KONG_ENV}" = "dev" ]; then
  : "${UPLOAD_TARGET:=host.docker.internal:8082}"
  : "${USER_TARGET:=host.docker.internal:8081}"
  : "${RUSTFS_TARGET:=host.docker.internal:9000}"
  : "${VIDEO_TARGET:=host.docker.internal:8085}"
else
  : "${UPLOAD_TARGET:=upload-service.go-video.svc:8082}"
  : "${USER_TARGET:=user-service.go-video.svc:8081}"
  : "${RUSTFS_TARGET:=rustfs.go-video.svc:9000}"
  : "${VIDEO_TARGET:=video-service.go-video.svc:8085}"
fi

export UPLOAD_TARGET USER_TARGET RUSTFS_TARGET VIDEO_TARGET

envsubst '${UPLOAD_TARGET} ${USER_TARGET} ${RUSTFS_TARGET} ${VIDEO_TARGET}' < "$TEMPLATE" > "$OUTPUT"

echo "Generated ${OUTPUT}" >&2
echo "  UPLOAD_TARGET=${UPLOAD_TARGET}" >&2
echo "  USER_TARGET=${USER_TARGET}" >&2
echo "  RUSTFS_TARGET=${RUSTFS_TARGET}" >&2
echo "  VIDEO_TARGET=${VIDEO_TARGET}" >&2
echo "  ENV=${KONG_ENV}" >&2
