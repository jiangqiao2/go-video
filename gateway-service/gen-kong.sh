#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEMPLATE="${SCRIPT_DIR}/kong.yml.tmpl"
OUTPUT="${SCRIPT_DIR}/kong.yml"

# Optional: load env file automatically when present
if [ -n "${ENV_FILE:-}" ] && [ -f "$ENV_FILE" ]; then
  set -a
  . "$ENV_FILE"
  set +a
elif [ -f "${SCRIPT_DIR}/.env.local" ]; then
  set -a
  . "${SCRIPT_DIR}/.env.local"
  set +a
elif [ -f "${SCRIPT_DIR}/.env" ]; then
  set -a
  . "${SCRIPT_DIR}/.env"
  set +a
fi

if ! command -v envsubst >/dev/null 2>&1; then
  echo "envsubst is required (install gettext)." >&2
  exit 1
fi

: "${UPLOAD_TARGET:=upload-service.go-video.svc:8082}"
: "${USER_TARGET:=user-service.go-video.svc:8081}"

export UPLOAD_TARGET USER_TARGET

envsubst '${UPLOAD_TARGET} ${USER_TARGET}' < "$TEMPLATE" > "$OUTPUT"

echo "Generated ${OUTPUT}" >&2
echo "  UPLOAD_TARGET=${UPLOAD_TARGET}" >&2
echo "  USER_TARGET=${USER_TARGET}" >&2
