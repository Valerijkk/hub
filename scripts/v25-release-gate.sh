#!/usr/bin/env bash
# v2.5.0 release gate — print checklist before tagging (no auto-tag).
# Ref: .planning/STATE.md · FA-06: .planning/ops/production/FA-06-EXECUTION-LOG.md
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HEAD_SHA="$(git -C "${ROOT}" rev-parse HEAD 2>/dev/null || echo unknown)"
HEAD_MSG="$(git -C "${ROOT}" log -1 --oneline 2>/dev/null || echo unknown)"

pass=true

echo "================================================================================"
echo " Hub v2.5.0 release gate"
echo " HEAD: ${HEAD_MSG}"
echo "================================================================================"
echo ""

echo "── Code & CI (automated repo checks) ──"
test -f "${ROOT}/scripts/fa06-rotate-p1.sh" && echo "  ✓ fa06-rotate-p1.sh (compose)" || { echo "  ✗ fa06-rotate-p1.sh"; pass=false; }
test -f "${ROOT}/scripts/fa06-rotate-p1-k3s.sh" && echo "  ✓ fa06-rotate-p1-k3s.sh (k3s)" || { echo "  ✗ fa06-rotate-p1-k3s.sh"; pass=false; }
test -f "${ROOT}/scripts/fa06-rotate-p2-smoke.sh" && echo "  ✓ fa06-rotate-p2-smoke.sh" || { echo "  ✗ fa06-rotate-p2-smoke.sh"; pass=false; }
test -f "${ROOT}/scripts/deploy-preflight.sh" && echo "  ✓ deploy-preflight.sh" || { echo "  ✗ deploy-preflight.sh"; pass=false; }

if command -v gh >/dev/null 2>&1; then
  echo ""
  echo "── GitHub CI (latest master push) ──"
  latest="$(gh run list --branch master --limit 1 --json conclusion,status,databaseId -q '.[0]' 2>/dev/null || echo '{}')"
  if echo "${latest}" | grep -q '"conclusion":"success"'; then
    echo "  ✓ Latest master CI: success"
  else
    echo "  ✗ Latest master CI not success — run: gh run list --branch master"
    pass=false
  fi
  open_prs="$(gh pr list --state open --json number -q 'length' 2>/dev/null || echo "?")"
  if [[ "${open_prs}" == "0" ]]; then
    echo "  ✓ No open PRs"
  else
    echo "  ⚠ Open PRs: ${open_prs} (review before tag)"
  fi
else
  echo "  ○ Install gh CLI to verify CI status from this machine"
fi

cat <<'EOF'

── FA-06 / PG-03 (manual — blocks v2.5.0 tag) ──
  [ ] Phase 0 git scan recorded (.planning/ops/production/FA-06-EXECUTION-LOG.md)
  [ ] P1 rotated: RESEND_API_KEY, TELEGRAM_BOT_TOKEN (INF-03)
  [ ] Phase 2 smoke: scripts/fa06-rotate-p2-smoke.sh all pass
  [ ] Old provider keys revoked (Resend, Telegram)
  [ ] Phase 3 CTO sign-off in FA-06-EXECUTION-LOG.md

── Deploy (manual) ──
  [ ] scripts/deploy-preflight.sh — SSH :22 + secrets OK
  [ ] Build & Deploy workflow green or deploy verified on server

── Tag (after all above) ──
  git tag -a v2.5.0 -m "Hub v2.5 Trust — Wave C complete, FA-06 rotated"
  git push origin v2.5.0

EOF

echo ""
if [[ "${pass}" == true ]]; then
  echo "Automated checks: PASS (manual FA-06 + CTO still required before tag)"
  exit 0
fi
echo "Automated checks: FAIL — fix items above before tagging"
exit 1
