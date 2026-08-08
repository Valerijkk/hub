# TASK-214 — 🏢 Business-domain greenfield cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **135**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase backend slug is retained for
> link stability; current purpose is greenfield business-domain behavior.

## [OUTCOME]

Python API/integration cases cover calendar, intranet, accounting, CRM, HR,
bases, dashboard, AI, search and storage from current requirements/risks.

## [ROLE]

- Owner: `qa-auto1`; verifiers: `qa-auto-ver`, `product-ver`.
- Depends on: TASK-205 and applicable Phase 133 fixtures.

## [OWNED TARGET PATHS]

`testframework/tests/api/business/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Domain lifecycle/authz plus invalid/boundary/edge/timezone/currency/precision/
  duplicate/concurrent/failure-recovery scenarios.
- AI/search/storage have deterministic oracle and privacy/redaction checks.
- P0/P1 trace gaps get owner/date; legacy counts are not acceptance criteria.
- Docker-only, sentinel, unexpected skip=0 and cleanup PASS.
