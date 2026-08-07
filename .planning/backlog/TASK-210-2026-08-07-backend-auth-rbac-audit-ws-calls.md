# TASK-210 — 🧪 Backend auth/rbac/audit/ws/calls wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **133**

## [OUTCOME]

Только доказанные service-free Go-unit rows auth, rbac, audit, ws и calls
работают из `backend/tests/**`. Все component rows принадлежат TASK-227 и
получают Python parity в этой же фазе.

## [ROLE]

- Owner: `be1` — five-domain migration and hook reconciliation.
- Verifier / handoff: `be-ver` — independent race/coverage/parity/mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/{auth,rbac,audit,ws,calls}/**`, shared support и read-only hook ledger; hook-файлы не меняются.
- Read-only: matching legacy tests/product.
- Do not touch: legacy deletion, CI, other domains, product behavior.
- DB/broker/network/process/HTTP component rows запрещено relabel как Go unit;
  они остаются legacy до принятого TASK-227 Python replacement.
- Depends on: TASK-209.

## [REFERENCES]

- TASK-197 ledger; TASK-203/204 baseline/contract
- matching backend legacy/internal directories; `.planning/codebase/TESTING.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/auth/... ./tests/rbac/... ./tests/audit/... ./tests/ws/... ./tests/calls/...`; exit code 0 и заявленные counts обязательны.

- Every in-scope unit row has target/parity state; component rows linked to
  TASK-227; missing/duplicate/component-under-backend = 0.
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
