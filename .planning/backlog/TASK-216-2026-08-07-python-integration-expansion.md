# TASK-216 — 🧪 Python infrastructure integration expansion

> Заведено: 2026-08-07 · Приоритет: **P1** · Статус: Открыта · Фаза: **136**

## [OUTCOME]

All ledger rows classified as Python API/integration/full have replacement coverage for PostgreSQL, Mongo, Redis, NATS, storage and provider-boundary behavior with clean isolation and no unexpected skips.

## [ROLE]

- Owner: `qa-auto2` — infrastructure/API/full expansion and legacy mapping.
- Verifier / handoff: `qa-auto-ver` — real-service, parallel, mutation and completeness verification.

## [SCOPE AND CONSTRAINTS]

- Allowed: `testframework/python/{core,fixtures,api,integration,full}/**`, Python reports/config.
- Read-only: TASK-197/209–215 ledgers and corresponding legacy tests/public APIs.
- Do not touch: product code, legacy deletion, CI, E2E browser suite.
- Required services fail closed; external providers use approved local deterministic adapters, never live credentials. Parallel workers get isolated namespaces/databases.
- Depends on: TASK-199, TASK-209–215.

## [REFERENCES]

- TASK-202/204 mapping and contract; `.planning/codebase/{TESTING,INTEGRATIONS,CONCERNS}.md`
- `docs/wiki/API.md`, `docs/security_guidelines.md`

## [ACCEPTANCE]

- Completion command: `cd testframework/python && python -m pytest api integration full --strict-markers -q`; exit code 0 и заявленные counts обязательны.

- All Python-classified backend rows have new path/test/TC/parity evidence; unmapped/duplicate = 0.
- Layer marker commands pass on clean stack and in documented parallel mode; required service outage is setup FAIL, not skip.
- Mutation proof: reversible persist/publish/auth/provider-boundary mutations each make mapped Python test RED; restore GREEN/diff-clean.
- Regression: legacy source domains, Python pilots and frozen Go coverage remain green; no claim that Python contributes to Go coverprofile.
- Evidence: service matrix, commands/counts/skips/timing, row validator, mutation and cleanup proof.
- Stop when: stable public boundary or required service unavailable; keep legacy and block parity.

## [OUTPUT SCHEMA]

1. Verdict. 2. Layer/service/row reconciliation. 3. Run/parallel/mutation evidence. 4. Legacy/coverage regression. 5. Handoff `qa-auto-ver`; risks/skill-log.

## [OPERATING RULES]

- Expand layer/domain batches; new then legacy regression; triage every FAIL.
- No unexpected skip, live provider, secret/PII artifact or broad cleanup.
- Docs-sync mapping/readme; TASK-linked atomic commits, no AI attribution.
