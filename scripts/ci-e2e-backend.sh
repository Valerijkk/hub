#!/usr/bin/env bash
# CI E2E helper — start/stop infra and backend for Playwright (GitHub Actions).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
COMPOSE=(docker compose -f "${ROOT}/docker-compose.local.yml")
BACKEND_PID_FILE="${ROOT}/.ci-backend.pid"
BACKEND_LOG="${ROOT}/.ci-backend.log"

cmd="${1:-}"

start_infra() {
  echo "Starting postgres, mongo, redis, nats..."
  "${COMPOSE[@]}" up -d postgres mongo redis nats
}

wait_infra() {
  echo "Waiting for Postgres..."
  for _ in $(seq 1 30); do
    if "${COMPOSE[@]}" exec -T postgres pg_isready -U hub >/dev/null 2>&1; then
      echo "Postgres ready"
      return 0
    fi
    sleep 2
  done
  echo "Postgres timeout"
  exit 1
}

start_backend() {
  cd "${ROOT}/backend"
  go generate ./ent
  go build -o bin/server ./cmd/server

  export SERVER_PORT="${SERVER_PORT:-3000}"
  export FIBER_IMMUTABLE="${FIBER_IMMUTABLE:-true}"
  export DATABASE_URL="${DATABASE_URL:-postgresql://hub:hub@localhost:5432/hub?sslmode=disable}"
  export MONGODB_URL="${MONGODB_URL:-mongodb://hub:hub@localhost:27017/?authSource=admin}"
  export MONGODB_DATABASE="${MONGODB_DATABASE:-hub}"
  export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"
  export NATS_URL="${NATS_URL:-nats://localhost:4222}"
  export S3_ENDPOINT="${S3_ENDPOINT:-http://localhost:8333}"
  export S3_ACCESS_KEY="${S3_ACCESS_KEY:-hub}"
  export S3_SECRET_KEY="${S3_SECRET_KEY:-hubsecret}"
  export S3_BUCKET="${S3_BUCKET:-hub}"
  export S3_REGION="${S3_REGION:-us-east-1}"
  export S3_USE_SSL="${S3_USE_SSL:-false}"
  export JWT_SECRET="${JWT_SECRET:-local-dev-secret-change-in-production}"
  export DEV_MODE="${DEV_MODE:-true}"
  export ENVIRONMENT="${ENVIRONMENT:-development}"

  nohup ./bin/server >"${BACKEND_LOG}" 2>&1 &
  echo $! >"${BACKEND_PID_FILE}"
  echo "Backend started (pid $(cat "${BACKEND_PID_FILE}"))"
}

wait_backend() {
  echo "Waiting for /api/health and /api/dev/status..."
  for _ in $(seq 1 60); do
    if curl -sf "http://127.0.0.1:${SERVER_PORT:-3000}/api/health" >/dev/null 2>&1; then
      if curl -sf "http://127.0.0.1:${SERVER_PORT:-3000}/api/dev/status" | grep -q '"dev_mode":true'; then
        echo "Backend ready (DEV_MODE=true)"
        return 0
      fi
    fi
    sleep 2
  done
  echo "Backend failed to start. Log tail:"
  tail -n 50 "${BACKEND_LOG}" || true
  exit 1
}

stop_all() {
  if [[ -f "${BACKEND_PID_FILE}" ]]; then
    kill "$(cat "${BACKEND_PID_FILE}")" 2>/dev/null || true
    rm -f "${BACKEND_PID_FILE}"
  fi
  "${COMPOSE[@]}" down -v --remove-orphans 2>/dev/null || true
}

case "${cmd}" in
  start-infra) start_infra ;;
  wait-infra) wait_infra ;;
  start-backend) start_backend ;;
  wait-backend) wait_backend ;;
  stop) stop_all ;;
  *)
    echo "Usage: $0 {start-infra|wait-infra|start-backend|wait-backend|stop}"
    exit 1
    ;;
esac
