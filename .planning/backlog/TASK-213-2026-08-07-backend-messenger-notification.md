# TASK-213 — 🧪 Backend messenger/notification wave

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **134**

## [OUTCOME]

Только genuine Go units messenger и notification работают из
`backend/tests/**`; Mongo/NATS/provider components получают Python parity
TASK-228 в этой же фазе.

## [ROLE]

- Owner: `be2` — two-domain migration and hook reconciliation.
- Verifier / handoff: `be-ver` — parity/race/coverage/mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/{messenger,notification}/**`, shared support и read-only hook ledger; hook-файлы не меняются.
- Read-only: matching legacy/product source. Do not touch: legacy deletion, CI, other domains, product behavior.
- Mongo/NATS/provider/whole-API rows stay legacy until TASK-228 parity; no
  external live providers и никакого component-under-backend.
- Depends on: TASK-210, TASK-211.

## [REFERENCES]

- TASK-197/203/204; matching backend directories; `.planning/codebase/{TESTING,INTEGRATIONS}.md`

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/messenger/... ./tests/notification/...`; exit code 0 и заявленные counts обязательны.

- All in-scope Go rows `PARITY_PROVEN`; Python rows linked/pending; missing/duplicate = 0.
- Target+legacy packages pass `-tags hubtest -race -count=1`; frozen coverage shows no unexplained regression.
- Hook decisions are symbol-level `retain|candidate-for-TASK-219` with zero-consumer/release-build proof; no removal in this wave.
- Mutation proof: messenger membership/authz guard and notification delivery/error boundary each RED target test; restore GREEN.
- Regression: prior backend waves and both legacy packages green; no unexpected required skips.
- Evidence: row totals, commands/counts/timing/coverage, hook and mutation logs.
- Stop when: real provider/infrastructure semantics would be mocked away; retain for TASK-216.

## [OUTPUT SCHEMA]

1. Verdict. 2. Domain/parity/hook table. 3. Race/coverage/mutation evidence. 4. Regression/pending rows. 5. Handoff `be-ver`; risks/skill-log.

## [OPERATING RULES]

- Domain batches, originals retained, FAIL triage/re-run old+new.
- No live providers, secret artifacts or silent coverage/assertion loss.
- Docs-sync ledger; atomic commits with TASK link, no AI attribution.
