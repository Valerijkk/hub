# TASK-218 — 🧪 Collection, skip и sentinel gates

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **140**

## [OUTCOME]

Каждый legacy и replacement CI lane fail-closed проверяет non-empty/expected collection, pass/fail/skip budget, ledger completeness и deliberate failing sentinel.

## [ROLE]

- Owner: `qa-auto2` — validators/gates and cross-runtime sentinel evidence.
- Verifier / handoff: `qa-auto-ver` — independent negative runs and real-PR gate review.

## [SCOPE AND CONSTRAINTS]

- Allowed: test count/skip/parity validators, test-only sentinels/fixtures, canonical scripts, minimal CI invocation changes coordinated with TASK-201.
- Do not touch: product behavior, thresholds, legacy deletion, load on PR.
- Expected counts derive from dated TASK-202 ledger and accepted transformations; skip allowlist has owner/reason/expiry.
- Depends on: TASK-201 and verifier evidence for TASK-193–217.

## [REFERENCES]

- TASK-202 ledger, TASK-204 contract, TASK-201 entrypoints
- `.planning/codebase/{TESTING,CONCERNS}.md`, `.cursor/playbooks/testing.md`

## [ACCEPTANCE]

- Completion command: `pwsh -NoProfile -File scripts/validate-test-gates.ps1`; exit code 0 и заявленные counts обязательны.

- Go legacy/target, FE legacy/target, Python API/integration/smoke/full lanes reject zero/short collection, failures and unexpected skips; ledger missing/duplicate/unmapped = 0.
- One deliberately failing sentinel per runtime/lane is added temporarily: expected job RED; remove/disable sentinel only via explicit test fixture and job returns GREEN; evidence from real PR when available.
- Tampered expected count and unexpected skip each make validator RED; restore GREEN.
- Regression: complete dual-run pipeline green without sentinels; load remains manual/scheduled; artifacts publish counts and ledger hash.
- Evidence: gate matrix, PR/job links or local equivalents, all RED→GREEN logs, counts/skips/artifacts.
- Stop when: any lane lacks trustworthy count/report or real PR approval; block cleanup TASK-219.

## [OUTPUT SCHEMA]

1. Verdict. 2. Lane→expected count/skip/sentinel table. 3. Negative/positive evidence. 4. Regression/remaining approvals. 5. Handoff `qa-auto-ver`; risks/skill-log.

## [OPERATING RULES]

- Prove failure sensitivity, then restore and prove diff-clean green.
- Do not hardcode stale counts without ledger hash/revision; no fail-open conditions.
- CI edits owner-reviewed; docs-sync gate contract; TASK-linked commits, no AI attribution.
