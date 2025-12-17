#!/usr/bin/env bash

# End-to-end follow load test runner for WSL/Linux.
# 1) Creates a "star" user via gateway.
# 2) Logs in to get star user's UUID.
# 3) Creates many "fan" accounts and collects their access tokens into tokens.txt.
# 4) Runs k6 follow_stress.js against the star user via gateway.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Default configuration (override via env vars if needed).
# For k8s deployment, Kong gateway NodePort is 30080 by default.
GATEWAY="${GATEWAY:-http://117.50.33.177:30080}"
STAR_ACCOUNT="${STAR_ACCOUNT:-star_user_1}"
STAR_PASSWORD="${STAR_PASSWORD:-StarUser123}"
FAN_COUNT="${FAN_COUNT:-200}"
FAN_BASE_ACCOUNT="${FAN_BASE_ACCOUNT:-fan_}"
FAN_PASSWORD="${FAN_PASSWORD:-FanPass123}"
TOKENS_FILE="${TOKENS_FILE:-tokens.txt}"
RATE="${RATE:-2000}"
DURATION="${DURATION:-1h}"

echo "== follow load test config =="
echo "GATEWAY          = $GATEWAY"
echo "STAR_ACCOUNT     = $STAR_ACCOUNT"
echo "FAN_COUNT        = $FAN_COUNT"
echo "FAN_BASE_ACCOUNT = $FAN_BASE_ACCOUNT"
echo "RATE             = $RATE"
echo "DURATION         = $DURATION"
echo

# Dependency checks
for cmd in curl jq k6; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "ERROR: '$cmd' is required but not found in PATH."
    if [ "$cmd" = "jq" ]; then
      echo "       Install on Ubuntu/WSL with: sudo apt update && sudo apt install -y jq"
    fi
    exit 1
  fi
done

echo "== Step 1: create / login star user =="

star_body=$(printf '{"account":"%s","password":"%s"}' "$STAR_ACCOUNT" "$STAR_PASSWORD")

echo "Registering star user (may fail if already exists, this is OK)..."
set +e
curl -sS -X POST "$GATEWAY/api/user/v1/open/users/register" \
  -H "Content-Type: application/json" \
  -d "$star_body" >/dev/null 2>&1
set -e

echo "Logging in star user..."
star_login_json=$(curl -sS -X POST "$GATEWAY/api/user/v1/open/users/login" \
  -H "Content-Type: application/json" \
  -d "$star_body")

STAR_USER_UUID=$(echo "$star_login_json" | jq -r '.data.user_uuid // empty')
STAR_ACCESS_TOKEN=$(echo "$star_login_json" | jq -r '.data.access_token // empty')

if [ -z "$STAR_USER_UUID" ] || [ "$STAR_USER_UUID" = "null" ]; then
  echo "ERROR: failed to obtain star user UUID from login response:"
  echo "$star_login_json"
  exit 1
fi

echo "Star user UUID: $STAR_USER_UUID"
echo

echo "== Step 2: generate fan tokens ($FAN_COUNT accounts) into $TOKENS_FILE =="

rm -f "$TOKENS_FILE"

for i in $(seq 1 "$FAN_COUNT"); do
  account=$(printf "%s%04d" "$FAN_BASE_ACCOUNT" "$i")
  body=$(printf '{"account":"%s","password":"%s"}' "$account" "$FAN_PASSWORD")
  echo "Processing fan account: $account"

  # Register fan (ignore errors if already exists)
  set +e
  curl -sS -X POST "$GATEWAY/api/user/v1/open/users/register" \
    -H "Content-Type: application/json" \
    -d "$body" >/dev/null 2>&1
  set -e

  # Login fan and capture access_token
  fan_login_json=$(curl -sS -X POST "$GATEWAY/api/user/v1/open/users/login" \
    -H "Content-Type: application/json" \
    -d "$body")

  token=$(echo "$fan_login_json" | jq -r '.data.access_token // empty')
  if [ -z "$token" ] || [ "$token" = "null" ]; then
    echo "  WARNING: no access_token for $account, response:"
    echo "  $fan_login_json"
  else
    echo "$token" >>"$TOKENS_FILE"
  fi
done

fan_token_count=$(wc -l <"$TOKENS_FILE" 2>/dev/null || echo 0)
if [ "$fan_token_count" -eq 0 ]; then
  echo "ERROR: no fan tokens written to $TOKENS_FILE"
  exit 1
fi

echo "Fan tokens written to $TOKENS_FILE ($fan_token_count tokens)."
echo

echo "== Step 3: run k6 follow_stress.js =="

export GATEWAY
export TARGET_UUID="$STAR_USER_UUID"

echo "Running k6 with:"
echo "  GATEWAY     = $GATEWAY"
echo "  TARGET_UUID = $TARGET_UUID"
echo

# Note: rate/duration are configured inside follow_stress.js. To change them
# you can either edit that file, or define scenarios dynamically if needed.
k6 run "$SCRIPT_DIR/follow_stress.js"
