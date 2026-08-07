# TASK-211 — 🧪 Backend platform/apigateway/dev/ent wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **133**

## [OUTCOME]

Classified Go-unit/component rows for platform, apigateway, dev and ent run from correct `backend/tests/**` targets with special-path and DB taxonomy preserved.

## [ROLE]

- Owner: `be2` — four-area migration, nested mapping and hook reconciliation.
- Verifier / handoff: `be-ver` — taxonomy, race/coverage, source-guard and mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/{platform,apigateway,dev,ent}/**`, shared support и read-only hook ledger; hook-файлы не меняются.
- Read-only: matching legacy/product source.
- Do not touch: legacy deletion, CI, other domains, product behavior.
- Preserve `platform/<subsystem>` nesting. Ent real-Postgres/schema integration rows stay legacy until Python integration parity; source guards use stable repo-root resolution.
- Depends on: TASK-209.

## [REFERENCES]

- TASK-197/204 classification; matching legacy directories
- `.planning/codebase/{STRUCTURE,TESTING,CONCERNS}.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/platform/... ./tests/apigateway/... ./tests/dev/... ./tests/ent/...`; exit code 0 и заявленные counts обязательны.

- In-scope Go rows map and pass in target; Python-classified rows explicitly remain pending, never duplicated/misclassified.
- Old/new target commands pass race; affected coverage and package counts reconcile with TASK-203 method.
- Hook symbols have `retain|candidate-for-TASK-219` evidence; nested packages and path-sensitive source guards work from clean checkout.
- Mutation proof: representative platform, apigateway and Ent constraint/source guard REDs target test; restore GREEN.
- Regression: all four legacy areas and release build green; unexpected skips explained/allowlisted per contract.
- Evidence: ledger totals, commands/counts/timing/coverage, hook and mutation logs.
- Stop when: real-engine semantics would be replaced by SQLite/mock; retain row for TASK-216.

## [OUTPUT SCHEMA]

1. Verdict. 2. Area/taxonomy/hook table. 3. Race/coverage/mutation evidence. 4. Regression/pending Python rows. 5. Handoff `be-ver`; skill-log.

## [OPERATING RULES]

- Evidence-based classification; no mechanical all-to-unit move.
- Originals retained; test triage and rerun; no broad hook deletion.
- Docs-sync ledger; atomic commits with TASK link, no AI attribution.
