#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${ZERO_API_APP_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
CONFIG_FILE="${ZERO_API_CONFIG:-etc/zero-api.yaml}"
SERVICE_NAME="${ZERO_API_SERVICE:-zero-api}"
HEALTH_CHECK_URL="${ZERO_API_HEALTH_CHECK_URL:-http://127.0.0.1:8888/api/v1/health}"
RUN_TESTS="${ZERO_API_RUN_TESTS:-1}"
GIT_PULL="${ZERO_API_GIT_PULL:-1}"

cd "$APP_DIR"

if [ ! -f "zero.go" ]; then
  echo "Missing zero.go. APP_DIR is wrong: $APP_DIR" >&2
  exit 1
fi

if [ ! -f "$CONFIG_FILE" ]; then
  echo "Missing config file: $CONFIG_FILE" >&2
  exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
  echo "Missing required command: systemctl" >&2
  exit 1
fi

if ! systemctl cat "$SERVICE_NAME" >/dev/null 2>&1; then
  echo "Missing systemd service: $SERVICE_NAME" >&2
  exit 1
fi

if ! systemctl cat "$SERVICE_NAME" | grep -F "go run zero.go -f" >/dev/null 2>&1; then
  echo "Warning: $SERVICE_NAME does not appear to use: go run zero.go -f ..." >&2
  echo "Current service definition:" >&2
  systemctl cat "$SERVICE_NAME" >&2
fi

if [ "$GIT_PULL" = "1" ]; then
  git pull --ff-only
fi

if [ "$RUN_TESTS" = "1" ]; then
  go test ./...
fi

systemctl restart "$SERVICE_NAME"
systemctl status "$SERVICE_NAME" --no-pager

if [ -n "$HEALTH_CHECK_URL" ]; then
  for _ in $(seq 1 20); do
    if curl -fsS "$HEALTH_CHECK_URL" >/dev/null; then
      echo "Health check passed: $HEALTH_CHECK_URL"
      exit 0
    fi
    sleep 1
  done

  echo "Health check failed: $HEALTH_CHECK_URL" >&2
  journalctl -u "$SERVICE_NAME" -n 80 --no-pager >&2
  exit 1
fi
