# TASK-210 — 🧪 Backend auth/rbac/audit/ws/calls wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **133**

## [OUTCOME]

All classified Go-unit/component rows for auth, rbac, audit, ws and calls run from `backend/tests/**` with dual-run parity, race safety and preserved coverage.

## [ROLE]

- Owner: `be1` — five-domain migration and hook reconciliation.
- Verifier / handoff: `be-ver` — independent race/coverage/parity/mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/{auth,rbac,audit,ws,calls}/**`, shared support и read-only hook ledger; hook-файлы не меняются.
- Read-only: matching legacy tests/product.
- Do not touch: legacy deletion, CI, other domains, product behavior.
- Python-classified rows remain legacy until Python parity; do not relabel them Go unit for convenience.
- Depends on: TASK-209.

## [REFERENCES]

- TASK-197 ledger; TASK-203/204 baseline/contract
- matching backend legacy/internal directories; `.planning/codebase/TESTING.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/auth/... ./tests/rbac/... ./tests/audit/... ./tests/ws/... ./tests/calls/...`; exit code 0 и заявленные counts обязательны.

- Every in-scope ledger row has target/parity state; old/new names, counts and assertions reconcile; missing/duplicate = 0.
- Target and legacy domain commands pass `-tags hubtest -race -count=1`; affected coverage under frozen method has no unexplained regression.
- Hooks processed symbol-by-symbol with zero-consumer evidence and `retain|candidate-for-TASK-219`; release build remains green, hooks retained.
- Mutation proof: one security/lifecycle guard per domain REDs corresponding target test; restore all GREEN/diff-clean.
- Regression: auth pilot and all five legacy packages green; unexpected skips = 0.
- Evidence: commands/counts/timing/coverage, ledger/hook tables, five mutations.
- Stop when: any row needs real infrastructure/Python replacement not yet proven; retain legacy row and block cleanup.

## [OUTPUT SCHEMA]

1. Verdict. 2. Domain/parity/hook totals. 3. Race/coverage/mutation evidence. 4. Regression/gaps. 5. Handoff `be-ver`; skill-log.

## [OPERATING RULES]

- Domain batches; originals retained; test triage and repeat until old+new green.
- No silent hook/coverage/TC removal; public behavior preferred over wider test seam.
- Docs-sync ledger; atomic commits with TASK link, no AI attribution.
