# TASK-210 — 🗄️ PostgreSQL/Mongo integration cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **133**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase backend-domain slug is kept
> only for link stability. Current purpose is PostgreSQL/Mongo integration.

## [OUTCOME]

Greenfield Python integration tests protect persistence, transactions,
constraints and cross-store consistency.

## [ROLE]

- Owner: `qa-auto1`; verifiers: `qa-auto-ver`, `be-ver`.
- Depends on: TASK-205.

## [OWNED TARGET PATHS]

`testframework/tests/integration/postgres/**`, `testframework/tests/integration/mongo/**`,
`testframework/docker/task-210-postgres-mongo.compose.yml`. This card
authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit
row to Python. TASK-210 may not mutate `testframework/docker/run-task.ps1` or
`testframework/docker/images.lock.json`; either requires a separate exact
devops-owned contract change and independent review.

## [ACCEPTANCE]

- Commit/rollback, uniqueness/FK/version, soft-delete, ordering/pagination,
  boundary values and concurrent writes covered where applicable.
- Partial PostgreSQL/Mongo failure and recovery have observable integrity oracle.
- Ephemeral limited Docker services, pinned digests, per-run namespaces and cleanup.
- TC trace complete, sentinel RED→GREEN, no dependency skip.
