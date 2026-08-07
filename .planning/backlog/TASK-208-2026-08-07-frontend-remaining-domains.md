# TASK-208 — 🧪 Frontend remaining-domain wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **131**

## [OUTCOME]

Все оставшиеся frontend ledger rows (admin, ai, dashboard, hr, notifications, search, v4 и residual/shared) имеют target parity in `frontend/tests/**`.

## [ROLE]

- Owner: `fe2` — remaining domains, residual reconciliation and full target suite.
- Verifier / handoff: `fe-ver` — 180-row completeness, static gates, coverage and mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/tests/{admin,ai,dashboard,hr,notifications,search,v4,shared}/**`.
- Read-only: matching legacy tests/SUT.
- Do not touch: legacy deletion/linker/CI, domains owned by TASK-206/207, product behavior.
- `v4` rows remap by SUT; broad bucket is not preserved when module evidence points elsewhere.
- Depends on: TASK-198, TASK-207.

## [REFERENCES]

- TASK-202 ledger; remaining legacy frontend directories
- `.planning/codebase/{STRUCTURE,TESTING,CONCERNS}.md`

## [ACCEPTANCE]

- Completion command: `cd frontend && npx vitest run tests`; exit code 0 и заявленные counts обязательны.

- All 180 frontend source rows have target path/status; `PARITY_PROVEN` total = 180, missing/duplicate/unexplained = 0.
- Full target suite passes test typecheck/Biome/Vitest and reports dated coverage comparable to TASK-198 method.
- Mutation proof: representative guards in at least admin/ai/v4 and one data-state domain cause target RED; restore GREEN.
- Regression: legacy 180 and TASK-206/207 targets green, counts/skips reconciled.
- Evidence: full row validator, commands/counts/timing/coverage and mutation logs.
- Stop when: any legacy row/TC/assertion is unmapped or coverage method differs; do not claim full parity.

## [OUTPUT SCHEMA]

1. Verdict. 2. 180-row reconciliation by domain. 3. Static/test/coverage/mutation evidence. 4. Regression/gaps. 5. Handoff `fe-ver`; skill-log.

## [OPERATING RULES]

- Batch by target SUT, preserve originals; triage/re-run until new and old green.
- No legacy cleanup, silent consolidation, weakened assertion or skip.
- Docs-sync ledger; atomic commits with TASK link, no AI attribution.
