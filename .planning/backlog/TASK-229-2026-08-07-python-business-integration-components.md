# TASK-229 — 🏢 Business/integration integrity closure

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **135**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Cross-domain review closes current P0/P1 boundaries, data-integrity and
provider failure/recovery gaps in business/integration Python suites.

## [ROLE]

- Owners: `qa-auto1`, `qa-auto2`; verifiers: `qa-auto-ver`, `product-ver`.
- Depends on: TASK-214, TASK-215.

## [OWNED TARGET PATHS]

`testframework/tests/integration/adapters/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Scenario checklist, integrity/rollback and provider matrices have no silent P0/P1 gap.
- No live provider/secret; egress control and canary scan proven.
- Sentinel, combined repeated Docker run, unexpected skip=0 and cleanup PASS.
