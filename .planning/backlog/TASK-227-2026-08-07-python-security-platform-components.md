# TASK-227 — 🛡️ Security/platform risk closure

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **132**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Greenfield Python security/platform suite closes current P0/P1 authz, tenant,
audit and abuse risks. It does not replace legacy rows one-for-one.

## [ROLE]

- Owners: `qa-auto1`, `qa-auto2`; verifiers: `qa-auto-ver`, `security-auditor`.
- Depends on: TASK-208, TASK-209.

## [OWNED TARGET PATHS]

`testframework/tests/api/security/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- P0/P1 risk→TC coverage complete or explicit owner/date backlog.
- Happy/negative/authz/boundary/edge/failure-recovery/security/privacy applied.
- No real egress/secrets; canary/sentinel and repeated Docker run PASS.
- Unexpected skip=0 and resources cleaned.
