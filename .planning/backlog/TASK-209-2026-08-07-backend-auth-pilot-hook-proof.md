# TASK-209 — 🧪 Backend auth pilot и hook proof

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **132**

## [OUTCOME]

Один репрезентативный auth batch работает из `backend/tests/auth/**`, доказывает parity и формализует безопасное consumer-by-consumer обращение с `hubtest` export hooks.

## [ROLE]

- Owner: `be2` — auth pilot, imports/hooks and before/after evidence.
- Verifier / handoff: `be-ver` — package semantics, hook consumers, race/coverage/mutation verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/tests/auth/**`, target shared support и read-only auth hook ledger; hook-файлы не меняются в этой задаче.
- Read-only: `backend/testframework/auth/**`, auth production code and TASK-197 ledger.
- Do not touch: legacy tests, CI, other domains, product behavior.
- Originals remain; target external package uses `-tags hubtest`. Hook retained unless exact symbol has zero legacy+target consumers and public-contract replacement is proven.
- Depends on: TASK-197.

## [REFERENCES]

- `backend/testframework/auth/**`, `backend/internal/auth/export_hooks.go`
- `.planning/codebase/{ARCHITECTURE,TESTING,CONVENTIONS}.md`, backend testing playbook

## [ACCEPTANCE]

- Completion command: `cd backend && go test -tags hubtest -race -count=1 ./tests/auth/... ./testframework/auth/...`; exit code 0 и заявленные counts обязательны.

- Pilot ledger rows map old→new tests/names/assertions; both target and legacy auth packages pass `-race -count=1` on same revision.
- Hook table lists symbol, old/new consumers and decision `retain|candidate-for-TASK-219`; zero-consumer search является evidence, а не разрешением на удаление.
- Mutation proof: break one auth production guard visible to pilot — target RED; restore GREEN. Hook-файлы и symbols остаются неизменными до TASK-219.
- Regression: legacy auth and target pilot green; frozen coverage method shows no unexplained loss for affected packages.
- Evidence: commands/counts/coverage, consumer searches, mutation RED→GREEN.
- Stop when: hook consumer cannot migrate safely or pilot changes product contract; retain hook and escalate.

## [OUTPUT SCHEMA]

1. Verdict. 2. Pilot rows/hook decision table. 3. Race/coverage/mutation evidence. 4. Regression/risks. 5. Handoff `be-ver`; skill-log.

## [OPERATING RULES]

- Copy/port, do not delete originals; hook changes one symbol at a time.
- Test triage then rerun target+legacy; coverage uses TASK-203 method.
- Docs-sync ledger; atomic commit with TASK link, no AI attribution.
