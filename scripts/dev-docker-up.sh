#!/usr/bin/env bash
# Hub — локальный dev-стек в Docker
# Usage: bash scripts/dev-docker-up.sh [--build] [--down] [--seed]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${ROOT}"

BUILD=0
DOWN=0
SEED=0
for arg in "$@"; do
  case "${arg}" in
    --build) BUILD=1 ;;
    --down)  DOWN=1 ;;
    --seed)  SEED=1 ;;
  esac
done

ensure_file() {
  local example="$1" target="$2"
  if [[ ! -f "${target}" ]]; then
    [[ -f "${example}" ]] || { echo "Missing template: ${example}" >&2; exit 1; }
    cp "${example}" "${target}"
    echo "Created ${target} from ${example}"
  fi
}

ensure_file ".env.local.example" ".env.local"
ensure_file "docker-compose.local.yml.example" "docker-compose.local.yml"

COMPOSE=(docker compose --env-file .env.local -f docker-compose.local.yml)

if [[ "${DOWN}" -eq 1 ]]; then
  "${COMPOSE[@]}" down -v
  exit 0
fi

ARGS=(up -d)
[[ "${BUILD}" -eq 1 ]] && ARGS+=(--build)

echo "Starting Hub local stack..."
"${COMPOSE[@]}" "${ARGS[@]}"

echo "Waiting for backend health..."
for _ in $(seq 1 60); do
  if curl -sf http://localhost:3000/api/health >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if [[ "${SEED}" -eq 1 ]]; then
  echo "Seeding demo data..."
  curl -sf -X POST http://localhost:3000/api/dev/seed >/dev/null
fi

cat <<EOF

Hub local dev ready:
  UI:   http://localhost:8080
  API:  http://localhost:3000
  Seed: curl -X POST http://localhost:3000/api/dev/seed
  Logs: docker compose --env-file .env.local -f docker-compose.local.yml logs -f
EOF
