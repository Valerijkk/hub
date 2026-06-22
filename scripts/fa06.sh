#!/usr/bin/env bash
# FA-06 / PG-03 — единая точка входа для runbook ротации секретов Hub v2.5.
# Использование: bash scripts/fa06.sh <команда>
#   phase0-log   — пути к журналу выполнения и runbook
#   p1-compose   — Phase 1: чеклист Docker Compose
#   p1-k3s       — Phase 1: чеклист K3s (kubectl patch)
#   p2-smoke     — Phase 2: curl smoke после ротации
#   deploy       — preflight SSH-деплоя
#   release-gate — чеклист перед git tag v2.5.0
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
FA-06 / PG-03 — ротация секретов (Hub v2.5 Trust)

  bash scripts/fa06.sh p1-compose    Phase 1 — Docker Compose (Resend + Telegram)
  bash scripts/fa06.sh p1-k3s        Phase 1 — K3s patch hub-app-secrets
  bash scripts/fa06.sh p2-smoke      Phase 2 — curl smoke (HUB_BASE, MAGIC_LINK_EMAIL)
  bash scripts/fa06.sh deploy         Preflight SSH-деплоя (SSH_HOST, SSH_KEY)
  bash scripts/fa06.sh release-gate   Чеклист перед git tag v2.5.0
  bash scripts/fa06.sh phase0-log     Пути к FA-06-EXECUTION-LOG.md и PG-03

Runbook: .planning/PG-03-SECRET-ROTATION.md
EOF
    ;;
esac
