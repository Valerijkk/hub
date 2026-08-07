# TASK-214 — 🧪 Backend business-domain wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **135**

## [OUTCOME]

Только genuine Go units calendar, intranet, accounting, crm, hr, bases и
dashboard работают из `backend/tests/**`; DB/API components получают Python
parity TASK-229.

## [ROLE]

- Owner: `be1` — seven-domain migration and hooks.
- Verifier / handoff: `be-ver` — taxonomy/parity/race/coverage/mutation review.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/{calendar,intranet,accounting,crm,hr,bases,dashboard}/**`, shared support и read-only hook ledger; hook-файлы не меняются.
- Read-only: matching legacy/product. Do not touch: legacy deletion, CI, other domains, behavior.
- Real PostgreSQL/transaction/whole-API rows remain legacy until TASK-229
  parity; SQLite may not substitute engine-specific assertions.
- Depends on: TASK-212, TASK-213.

## [REFERENCES]

- TASK-197/203/204; matching directories; `.planning/codebase/{STRUCTURE,TESTING,CONCERNS}.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/calendar/... ./tests/intranet/... ./tests/accounting/... ./tests/crm/... ./tests/hr/... ./tests/bases/... ./tests/dashboard/...`; exit code 0 и заявленные counts обязательны.

- All seven-domain Go rows `PARITY_PROVEN`; Python rows explicitly pending; missing/duplicate = 0.
- Target+legacy race runs and frozen coverage comparison green; hook decisions are `retain|candidate-for-TASK-219` with consumer proof.
- Mutation proof: one meaningful guard/constraint across each of the seven domains REDs target test; restore GREEN.
- Regression: all earlier backend target/legacy waves green; skip policy satisfied.
- Evidence: ledger totals, commands/counts/timing/coverage, hook/mutation logs.
- Stop when: engine-specific or cross-domain behavior would be weakened; retain row for Python.

## [OUTPUT SCHEMA]

1. Verdict. 2. Seven-domain parity/hook table. 3. Race/coverage/mutations. 4. Regression/pending/risks. 5. Handoff `be-ver`; skill-log.

## [OPERATING RULES]

- Atomic domain batches; originals retained; triage/re-run old+new.
- No test semantic, TC, hook or coverage loss.
- Docs-sync ledger; TASK-linked commits, no AI attribution.
