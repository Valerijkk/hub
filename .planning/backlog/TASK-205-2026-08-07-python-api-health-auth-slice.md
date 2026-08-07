# TASK-205 — 🧪 Python API health/auth vertical slice

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **129**

## [OUTCOME]

Python framework проверяет health и полный auth vertical slice через публичный HTTP contract, выдаёт structured report и не зависит от Go/TS internals.

## [ROLE]

- Owner: `qa-auto2` — API client, health/auth tests и traceability.
- Verifier / handoff: `qa-auto-ver` — clean-env, negative/mutation и legacy-parity verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/python/api/{health,auth}/**`, `testframework/python/core/**`, auth/API fixtures and reports.
- Read-only: `docs/wiki/API.md`, current auth Go tests and E2E helpers.
- Do not touch: product code, legacy tests, CI, E2E browser layer.
- Cover health, invalid/missing auth, permitted test login/Magic Link flow, session use/revoke and error envelope/request_id; never use production mail/secrets.
- Depends on: TASK-193, TASK-203, TASK-204.

## [REFERENCES]

- `backend/testframework/auth/**`, `testframework/stages/e2e/shared/helpers/auth.ts`
- `docs/wiki/API.md`, `docs/security_guidelines.md`, `.planning/codebase/TESTING.md`

## [ACCEPTANCE]

- Completion command: `cd testframework/python && python -m pytest api/health api/auth -m api --strict-markers -q`; exit code 0 и заявленные counts обязательны.

- `pytest -m api` collects non-zero health/auth tests and passes against documented isolated stack; unauthorized paths assert status and safe error contract.
- Ledger maps each new test to legacy behavior/TC; all artifacts redact tokens, codes, cookies and PII.
- Mutation proof: temporarily bypass one auth denial or alter expected revoke behavior in a reversible local production patch — relevant Python test RED; restore GREEN/diff-clean.
- Regression: Python scaffold smoke and legacy auth domain tests green; counts/skips recorded, unexpected skips = 0.
- Evidence: environment, commands, counts, timings, reports and mutation RED→GREEN.
- Stop when: only production credential/mailbox can complete auth or contract contradicts docs; classify and escalate.

## [OUTPUT SCHEMA]

1. Verdict/summary. 2. Endpoint→scenario→TC mapping. 3. Commands/counts/timing/mutation. 4. Regression/redaction/risks. 5. Handoff `qa-auto-ver`; skill-log.

## [OPERATING RULES]

- Implement→run→triage/fix until green, then legacy regression; no mocks of Hub API.
- Required service absence fails setup; no unexpected skips or secret-bearing artifacts.
- Preserve all legacy; docs-sync Python API guidance; atomic commit with TASK link, no AI attribution.
