#!/usr/bin/env bash
# FA-06 / PG-03 — entry point for secret rotation runbooks.
# Usage: bash scripts/fa06.sh <command>
#   phase0-log   — open execution log path
#   p1-compose   — Phase 1 Docker Compose checklist
#   p1-k3s       — Phase 1 K3s kubectl checklist
#   p2-smoke     — Phase 2 post-rotation curl smoke
#   deploy       — SSH deploy preflight
#   release-gate — v2.5.0 tag readiness
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cmd="${1:-help}"

case "${cmd}" in
  phase0-log)
    echo "${ROOT}/.planning/FA-06-EXECUTION-LOG.md"
    echo "${ROOT}/.planning/PG-03-SECRET-ROTATION.md"
    ;;
  p1-compose)  exec bash "${ROOT}/scripts/fa06-rotate-p1.sh" ;;
  p1-k3s)      exec bash "${ROOT}/scripts/fa06-rotate-p1-k3s.sh" ;;
  p2-smoke)    exec bash "${ROOT}/scripts/fa06-rotate-p2-smoke.sh" ;;
  deploy)      exec bash "${ROOT}/scripts/deploy-preflight.sh" ;;
  release-gate) exec bash "${ROOT}/scripts/v25-release-gate.sh" ;;
  help|*)
    cat <<EOF
FA-06 / PG-03 secret rotation (Hub v2.5 Trust)

  bash scripts/fa06.sh p1-compose    Phase 1 — Docker Compose (Resend + Telegram)
  bash scripts/fa06.sh p1-k3s        Phase 1 — K3s hub-app-secrets patch
  bash scripts/fa06.sh p2-smoke      Phase 2 — curl smoke (set HUB_BASE, MAGIC_LINK_EMAIL)
  bash scripts/fa06.sh deploy         SSH deploy preflight (optional SSH_HOST)
  bash scripts/fa06.sh release-gate   Checklist before git tag v2.5.0
  bash scripts/fa06.sh phase0-log     Paths to FA-06-EXECUTION-LOG.md + PG-03

Runbook: .planning/PG-03-SECRET-ROTATION.md
EOF
    ;;
esac
