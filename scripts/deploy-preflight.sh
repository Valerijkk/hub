#!/usr/bin/env bash
# Pre-flight checklist for GitHub Actions SSH deploy (.github/workflows/deploy.yml).
# Read-only diagnostics — never prints secret values.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SSH_HOST="${SSH_HOST:-}"
SSH_PORT="${SSH_PORT:-22}"

cat <<'EOF'
================================================================================
 Hub — Deploy preflight (Build & Deploy workflow)
 Workflow: .github/workflows/deploy.yml · continue-on-error on deploy job
================================================================================

GitHub repository secrets (Settings → Secrets → Actions):
  [ ] SSH_HOST     — target server hostname or IP
  [ ] SSH_USER     — deploy user (docker group member)
  [ ] SSH_KEY      — private key (ed25519), public key in authorized_keys

Server layout (deploy script expects):
  [ ] /opt/hub — git clone of this repo
  [ ] /opt/hub/deployments/.env — prod env (gitignored, from .env.example)
  [ ] docker + docker compose v2
  [ ] GHCR pull access (docker login ghcr.io for deploy user if images private)

Deploy script actions:
  • docker pull ghcr.io/valerijkk/hub/backend:<sha>
  • docker pull ghcr.io/valerijkk/hub/frontend:<sha>
  • git reset --hard <sha> in /opt/hub
  • compose up backend hub-mail livekit + standalone frontend container
  • curl http://127.0.0.1:3000/api/health

EOF

echo "Local repo: ${ROOT}"
echo "Latest master: $(git -C "${ROOT}" log -1 --oneline origin/master 2>/dev/null || git -C "${ROOT}" log -1 --oneline)"
echo ""

if command -v gh >/dev/null 2>&1; then
  echo "Recent Build & Deploy runs:"
  gh run list --workflow deploy.yml --limit 3 2>/dev/null || echo "  (gh run list failed — check network/auth)"
  echo ""
else
  echo "  (install gh CLI to list workflow runs from this machine)"
  echo ""
fi

if [[ -n "${SSH_HOST}" ]]; then
  echo "TCP probe: ${SSH_HOST}:${SSH_PORT}"
  if command -v nc >/dev/null 2>&1; then
    if nc -z -w 5 "${SSH_HOST}" "${SSH_PORT}" 2>/dev/null; then
      echo "  ✓ Port ${SSH_PORT} reachable on ${SSH_HOST}"
    else
      echo "  ✗ Port ${SSH_PORT} NOT reachable — fix firewall / security group / SSH_HOST"
    fi
  elif command -v bash >/dev/null 2>&1; then
    if timeout 5 bash -c "echo >/dev/tcp/${SSH_HOST}/${SSH_PORT}" 2>/dev/null; then
      echo "  ✓ Port ${SSH_PORT} reachable on ${SSH_HOST}"
    else
      echo "  ✗ Port ${SSH_PORT} NOT reachable"
    fi
  else
    echo "  (install nc or use: ssh -o ConnectTimeout=5 ${SSH_USER:-deploy}@${SSH_HOST})"
  fi
else
  echo "Optional: export SSH_HOST=your.server.com to run TCP probe from this machine"
fi

echo ""
echo "On-server quick check (run via SSH):"
echo "  docker ps --format 'table {{.Names}}\t{{.Status}}' | head -20"
echo "  curl -sf http://127.0.0.1:3000/api/health | head -c 200"
echo "  test -f /opt/hub/deployments/.env && echo '.env present'"
