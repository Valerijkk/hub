#!/usr/bin/env bash
# FA-06 Phase 1 — P1 secret rotation checklist for K3s (namespace hub).
# Prints steps only; never embed or echo secret values.
# Runbook: .planning/ops/production/PG-03-SECRET-ROTATION.md · Log: .planning/ops/production/FA-06-EXECUTION-LOG.md
set -euo pipefail

NS="${HUB_K8S_NAMESPACE:-hub}"
SECRET="${HUB_K8S_SECRET:-hub-app-secrets}"

cat <<EOF
================================================================================
 FA-06 Phase 1 — P1 rotation (K3s)
 Namespace: ${NS} · Secret: ${SECRET}
 Secrets: RESEND_API_KEY, TELEGRAM_BOT_TOKEN (INF-03 history exposure)
 NEVER paste secret values into tickets, chat, or git.
================================================================================

STEP 1 — Provider revocation (do BEFORE patching secret)
  • Resend dashboard → API Keys → revoke compromised key → create new key
  • Telegram @BotFather → /revoke → issue new bot token

STEP 2 — Base64-encode new values (run locally, not in shared shell history)
  Linux:
    echo -n 'REPLACE_PLAINTEXT' | base64 -w0
  macOS:
    echo -n 'REPLACE_PLAINTEXT' | base64

STEP 3 — Patch K8s secret (replace placeholders; keys must match your manifest)
  kubectl patch secret ${SECRET} -n ${NS} --type='json' -p='[
    {"op": "replace", "path": "/data/RESEND_API_KEY", "value": "<BASE64_RESEND_API_KEY>"},
    {"op": "replace", "path": "/data/TELEGRAM_BOT_TOKEN", "value": "<BASE64_TELEGRAM_BOT_TOKEN>"}
  ]'

  If keys use different names in your cluster, list them first:
    kubectl get secret ${SECRET} -n ${NS} -o jsonpath='{.data}' | jq 'keys'

STEP 4 — Rolling restart
  kubectl rollout restart deployment/hub-backend -n ${NS}
  kubectl rollout restart deployment/hub-mail -n ${NS}
  kubectl rollout status deployment/hub-backend -n ${NS} --timeout=300s

STEP 5 — Phase 2 smoke
  export HUB_BASE="https://yourdomain.com/api"
  export MAGIC_LINK_EMAIL="ops@yourdomain.com"
  bash scripts/fa06-rotate-p2-smoke.sh

STEP 6 — Sign-off
  • Mark RESEND + TELEGRAM in .planning/ops/production/FA-06-EXECUTION-LOG.md (Rotated/Verified)
  • Revoke old keys at Resend + BotFather after smoke passes

Optional pre-check (this script):
EOF

if command -v kubectl >/dev/null 2>&1; then
  if kubectl get secret "${SECRET}" -n "${NS}" >/dev/null 2>&1; then
    echo "  ✓ Secret ${SECRET} exists in namespace ${NS}"
    for key in RESEND_API_KEY TELEGRAM_BOT_TOKEN; do
      if kubectl get secret "${SECRET}" -n "${NS}" -o "jsonpath={.data.${key}}" 2>/dev/null | grep -q .; then
        echo "  ✓ data.${key} present (value hidden)"
      else
        echo "  ✗ data.${key} missing — verify secret key names"
      fi
    done
  else
    echo "  ✗ Secret ${SECRET} not found in ${NS} (check namespace / secret name)"
  fi
else
  echo "  (kubectl not available — run on machine with cluster access)"
fi
