# TASK-215 — 🧪 Backend AI/search/storage/integrations wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **135**

## [OUTCOME]

Classified Go-unit/component rows for ai, search, storage, backup, imex, github, gitlab, onec, marketplace and mail run from `backend/tests/**`, while real-integration rows remain explicitly mapped for Python.

## [ROLE]

- Owner: `be2` — ten-area migration, special path mapping and hooks.
- Verifier / handoff: `be-ver` — taxonomy, provider isolation, race/coverage/mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: corresponding `backend/tests/**` directories including `integrations/onec`, shared support и read-only hook ledger; hook-файлы не меняются.
- Read-only: matching legacy/product. Do not touch: legacy deletion, CI, other domains, behavior.
- No real external provider/network in unit lane; provider/DB/object-storage/full-API rows stay legacy for TASK-216.
- Depends on: TASK-212, TASK-213.

## [REFERENCES]

- TASK-197/203/204; matching directories; `.planning/codebase/{STRUCTURE,TESTING,INTEGRATIONS}.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/ai/... ./tests/search/... ./tests/storage/... ./tests/backup/... ./tests/imex/... ./tests/github/... ./tests/gitlab/... ./tests/integrations/onec/... ./tests/marketplace/... ./tests/mail/...`; exit code 0 и заявленные counts обязательны.

- All in-scope Go rows `PARITY_PROVEN`, integration rows linked to Python destination; no missing/duplicate.
- Target+legacy race and coverage evidence green; path-sensitive guards work from clean checkout; hooks have symbol-level `retain|candidate-for-TASK-219` proof.
- Mutation proof: representative AI policy, search scope, storage/backup integrity and provider adapter guard RED target tests; restore GREEN.
- Regression: previous waves plus all in-scope legacy packages green; no live network/secrets or unexpected skips.
- Evidence: mapping/taxonomy totals, commands/counts/timing/coverage, hook/mutation logs.
- Stop when: deterministic unit isolation is impossible; preserve row for TASK-216, do not add live dependency.

## [OUTPUT SCHEMA]

1. Verdict. 2. Area/taxonomy/hook table. 3. Race/coverage/mutations. 4. Regression/Python handoff. 5. Handoff `be-ver`; risks/skill-log.

## [OPERATING RULES]

- Evidence-based batch migration, originals retained, test triage/re-run.
- No live providers, secret leakage, broad hook removal or coverage weakening.
- Docs-sync ledger; TASK-linked atomic commits, no AI attribution.
