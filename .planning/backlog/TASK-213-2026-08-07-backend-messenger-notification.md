# TASK-213 — 💬 Realtime/comms API and integration cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **134**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase backend slug is kept only for
> link stability. Current purpose is realtime/comms API and integration.

## [OUTCOME]

Greenfield Python cases cover messenger, notification, calls and realtime
delivery/recovery behavior.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `qa-auto-ver`, `product-ver`.
- Depends on: TASK-205, TASK-211.

## [OWNED TARGET PATHS]

`testframework/tests/api/realtime/**`, `testframework/tests/integration/realtime/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Offline/reconnect, order, duplicate, acknowledgement, retry, provider outage,
  permission/membership change and recovery covered with bounded waits.
- Message/file/input boundaries and cross-tenant privacy asserted.
- No real providers; deterministic local fakes/contracts.
- Traceability, sentinel, repeated Docker run and cleanup PASS.
