#!/usr/bin/env bash
set -euo pipefail

# Generate kong.yml using dev env and start kong via docker-compose
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# 防止之前 kong.yml 被错误创建成目录
if [ -d "${SCRIPT_DIR}/kong.yml" ]; then
  echo "[local] Detected kong.yml as directory, removing to regenerate..."
  rm -rf "${SCRIPT_DIR}/kong.yml"
fi

echo "[local] Generating kong.yml with KONG_ENV=dev ..."
KONG_ENV=dev ./gen-kong.sh

echo "[local] Restarting gateway (kong + konga)..."
docker-compose down -v --remove-orphans || true
docker-compose build --no-cache
docker-compose up -d

echo "[local] Kong admin: http://localhost:8001"
echo "[local] Kong proxy: http://localhost:8000"
echo "[local] Konga UI:  http://localhost:1337"
