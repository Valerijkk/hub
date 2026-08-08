# TASK-212 — 🧩 Work lifecycle API cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **134**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase backend slug is kept only for
> link stability. Current purpose is public work-lifecycle API behavior.

## [OUTCOME]

Greenfield Python cases cover tracker/project/sprint/TMS requirements and risks.

## [ROLE]

- Owner: `qa-auto1`; verifiers: `qa-auto-ver`, `product-ver`.
- Depends on: TASK-205; TASK-210 for persistence-sensitive cases.

## [OWNED TARGET PATHS]

`testframework/tests/api/work/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Lifecycle/state transitions, invalid/stale/duplicate/concurrent actions,
  boundaries, ordering/pagination and time/timezone covered.
- Role/tenant behavior and persisted business invariant asserted.
- Malformed/oversized/partial TMS input and storage failure/recovery covered.
- Traceability, sentinel, Docker repeat and cleanup PASS.
