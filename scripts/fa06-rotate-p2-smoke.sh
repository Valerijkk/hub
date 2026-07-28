#!/usr/bin/env bash
# FA-06 Phase 2 — post-rotation HTTP smoke checks (Resend / Telegram / core stack).
# Requires: curl, jq (optional but recommended).
# Runbook: .planning/ops/production/PG-03-SECRET-ROTATION.md · Log: .planning/ops/production/FA-06-EXECUTION-LOG.md
#
# Usage:
#   export HUB_BASE="https://yourdomain.com/api"
#   export MAGIC_LINK_EMAIL="ops@yourdomain.com"
#   export HUB_ACCESS_TOKEN="<jwt after login>"   # optional — enables checks 4–7
#   bash scripts/fa06-rotate-p2-smoke.sh
set -euo pipefail

HUB_BASE="${HUB_BASE:-http://127.0.0.1:3000/api}"
MAGIC_LINK_EMAIL="${MAGIC_LINK_EMAIL:-}"
HUB_ACCESS_TOKEN="${HUB_ACCESS_TOKEN:-}"

pass=0
fail=0
skip=0

check() {
  local name="$1"
  shift
  if "$@"; then
    echo "  ✓ ${name}"
    pass=$((pass + 1))
  else
    echo "  ✗ ${name}"
    fail=$((fail + 1))
  fi
}

skip_check() {
  local name="$1"
  echo "  ○ ${name} (skipped)"
  skip=$((skip + 1))
}

curl_json() {
  curl -sS -f "$@"
}

echo "================================================================================"
echo " FA-06 Phase 2 — smoke checks"
echo " HUB_BASE=${HUB_BASE}"
echo "================================================================================"
echo ""

# 1 — Health
check "Health (postgres/redis/mongo)" bash -c '
  body="$(curl_json "${HUB_BASE}/health")"
  echo "${body}" | grep -q "\"status\":\"ok\"" || exit 1
  echo "${body}" | grep -q "postgres" || exit 1
'

# 2 — Magic Link send (Resend)
if [[ -n "${MAGIC_LINK_EMAIL}" ]]; then
  check "Magic Link send (Resend)" bash -c "
    code=\$(curl -sS -o /tmp/hub-magic-link.json -w '%{http_code}' \\
      -X POST \"\${HUB_BASE}/auth/magic-link\" \\
      -H 'Content-Type: application/json' \\
      -d '{\"email\":\"${MAGIC_LINK_EMAIL}\"}')
    [[ \"\${code}\" == \"200\" || \"\${code}\" == \"204\" ]]
  "
else
  skip_check "Magic Link send — set MAGIC_LINK_EMAIL"
fi

# 3 — Magic Link verify (manual code)
echo "  ○ Magic Link verify — manual: check email, then verify via UI or:"
echo "      curl -X POST \"\${HUB_BASE}/auth/magic-link/verify\" -H 'Content-Type: application/json' \\"
echo "        -d '{\"email\":\"<email>\",\"code\":\"<6-digit>\"}'"
skip=$((skip + 1))

if [[ -z "${HUB_ACCESS_TOKEN}" ]]; then
  skip_check "Authenticated checks — set HUB_ACCESS_TOKEN after login"
else
  # 4 — /me
  check "/auth/me" bash -c "
    curl_json -H 'Authorization: Bearer ${HUB_ACCESS_TOKEN}' \"${HUB_BASE}/auth/me\" | grep -q '\"id\"'
  "

  # 5 — Presigned upload URL (S3)
  check "Files presigned upload-url (S3)" bash -c "
    curl_json -X POST -H 'Authorization: Bearer ${HUB_ACCESS_TOKEN}' \\
      -H 'Content-Type: application/json' \\
      -d '{\"filename\":\"fa06-smoke.txt\",\"content_type\":\"text/plain\",\"size_bytes\":12}' \\
      \"${HUB_BASE}/files/upload-url\" | grep -q 'upload_url'
  "

  # 6 — Search scoped
  check "Search scoped" bash -c "
    code=\$(curl -sS -o /dev/null -w '%{http_code}' \\
      -H 'Authorization: Bearer ${HUB_ACCESS_TOKEN}' \\
      \"${HUB_BASE}/search?q=fa06\")
    [[ \"\${code}\" == \"200\" ]]
  "

  # 7 — Telegram status (bot token valid)
  check "Telegram bind status" bash -c "
    code=\$(curl -sS -o /dev/null -w '%{http_code}' \\
      -H 'Authorization: Bearer ${HUB_ACCESS_TOKEN}' \\
      \"${HUB_BASE}/notifications/telegram/status\")
    [[ \"\${code}\" == \"200\" ]]
  "
fi

echo ""
echo "Manual UI smoke (recommended): issue → Kanban → messenger → call → admin audit"
echo ""
echo "Summary: pass=${pass} fail=${fail} skipped=${skip}"
if [[ "${fail}" -gt 0 ]]; then
  echo "FA-06 Phase 2: FAIL — fix errors before CTO sign-off"
  exit 1
fi
echo "FA-06 Phase 2: automated checks OK (complete manual steps above)"
