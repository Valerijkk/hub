# TASK-212 — 🧪 Backend tracker/project/sprint/tms wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **134**

## [OUTCOME]

Classified Go-unit/component rows for tracker, project, sprint and tms run from `backend/tests/**` with dual-run parity, race and coverage preservation.

## [ROLE]

- Owner: `be1` — four-domain migration and hooks.
- Verifier / handoff: `be-ver` — independent parity/race/coverage/mutation review.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/{tracker,project,sprint,tms}/**`, shared support и read-only hook ledger; hook-файлы не меняются.
- Read-only: matching legacy/product source.
- Do not touch: legacy deletion, CI, other domains or product behavior.
- PostgreSQL/Mongo/NATS/whole-API rows follow TASK-204 taxonomy and remain legacy until TASK-199/216 parity.
- Depends on: TASK-210, TASK-211.

## [REFERENCES]

- TASK-197 ledger, TASK-203 baseline, TASK-204 contract
- matching backend directories; `.planning/codebase/TESTING.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/tracker/... ./tests/project/... ./tests/sprint/... ./tests/tms/...`; exit code 0 и заявленные counts обязательны.

- All in-scope Go rows `PARITY_PROVEN`; pending Python rows explicitly linked; no missing/duplicate rows.
- Target+legacy commands pass `hubtest`, race and frozen coverage comparison; package/test counts reconcile.
- Hook decisions are symbol-level `retain|candidate-for-TASK-219` with zero-consumer/release-build proof; no removal in this wave.
- Mutation proof: meaningful guard/transaction behavior in each domain REDs target test; restore GREEN.
- Regression: prior backend waves and four legacy domains green; unexpected skips = 0 for required scope.
- Evidence: row totals, commands/counts/timing/coverage, hook and four mutation logs.
- Stop when: migration would weaken DB/transaction semantics or coverage; leave legacy/pending Python.

## [OUTPUT SCHEMA]

1. Verdict. 2. Domain/parity/hook table. 3. Race/coverage/mutations. 4. Regression/pending rows/risks. 5. Handoff `be-ver`; skill-log.

## [OPERATING RULES]

- Domain batches, originals retained, triage/re-run old+new.
- No silent assertion, TC, hook or coverage loss.
- Docs-sync ledger; atomic commits with TASK link, no AI attribution.
