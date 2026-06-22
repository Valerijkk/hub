#!/usr/bin/env bash
# FA-06 Phase 1 — P1 secret rotation checklist for Docker Compose prod.
# Prints steps only; never embed or echo secret values.
# Runbook: .planning/PG-03-SECRET-ROTATION.md · Log: .planning/FA-06-EXECUTION-LOG.md
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${ROOT}/deployments/.env"
COMPOSE=(docker compose -f "${ROOT}/deployments/docker-compose.prod.yml" --env-file "${ENV_FILE}" -p hub)

cat <<'EOF'
================================================================================
 FA-06 Phase 1 — P1 rotation (Docker Compose prod)
 Secrets: RESEND_API_KEY, TELEGRAM_BOT_TOKEN (INF-03 history exposure)
 NEVER paste secret values into tickets, chat, or git.
================================================================================

STEP 1 — Provider revocation (do BEFORE updating .env)
  • Resend dashboard → API Keys → revoke compromised key → create new key
  • Telegram @BotFather → /revoke → issue new bot token

STEP 2 — Update deployments/.env (local on server only; file is gitignored)
  • RESEND_API_KEY=<new value from Resend>
  • TELEGRAM_BOT_TOKEN=<new value from BotFather>
  • Verify RESEND_FROM_EMAIL is still valid for your domain

STEP 3 — Recreate affected services
EOF

echo "  cd ${ROOT}/deployments"
echo "  docker compose -f docker-compose.prod.yml --env-file .env -p hub \\"
echo "    up -d --force-recreate backend hub-mail"

cat <<'EOF'

STEP 4 — Post-deploy verification (Phase 2 smoke)
  • curl -sf http://127.0.0.1:3000/api/health
  • Trigger Magic Link login → email arrives (Resend)
  • Settings → Notifications → Telegram bind → test notification

STEP 5 — Sign-off
  • Mark RESEND + TELEGRAM rows in .planning/FA-06-EXECUTION-LOG.md (Rotated/Verified)
  • Do NOT commit .env or log secret values

Optional pre-check (this script):
EOF

if [[ -f "${ENV_FILE}" ]]; then
  echo "  ✓ Found ${ENV_FILE}"
  for key in RESEND_API_KEY TELEGRAM_BOT_TOKEN; do
    if grep -q "^${key}=" "${ENV_FILE}" 2>/dev/null; then
      echo "  ✓ ${key} is set (value hidden)"
    else
      echo "  ✗ ${key} missing in .env — add before recreate"
    fi
  done
else
  echo "  ✗ ${ENV_FILE} not found — copy from deployments/.env.example first"
fi

echo ""
echo "Current compose project status (if docker available):"
if command -v docker >/dev/null 2>&1 && [[ -f "${ENV_FILE}" ]]; then
  "${COMPOSE[@]}" ps --status running 2>/dev/null || echo "  (compose ps failed — run on prod host)"
else
  echo "  (skipped — docker or .env unavailable on this machine)"
fi
