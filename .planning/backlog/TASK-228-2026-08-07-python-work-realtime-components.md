# TASK-228 — 🧩 Work/realtime edge and failure closure

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **134**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Cross-domain review closes current P0/P1 edge/corner and failure/recovery gaps
for work and realtime Python suites; legacy cases and counts are not inputs.

## [ROLE]

- Owners: `qa-auto1`, `qa-auto2`; verifiers: `qa-auto-ver`, `product-ver`.
- Depends on: TASK-212, TASK-213.

## [OWNED TARGET PATHS]

`testframework/tests/api/realtime/**`, `testframework/tests/integration/realtime/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Scenario checklist and lifecycle/delivery matrices have no silent P0/P1 gap.
- Cross-domain concurrency, time/order, partial failure and recovery cases included.
- Sentinel, combined repeated Docker run, unexpected skip=0 and cleanup PASS.
