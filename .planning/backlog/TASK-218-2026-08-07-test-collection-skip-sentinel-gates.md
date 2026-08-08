# TASK-218 — 🚦 Python quality/traceability gates

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **130**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Framework fails closed on zero collection, unexpected skip, invalid TC metadata,
orphan requirement/task/risk link, missing artifact and insensitive sentinel.

## [ROLE]

- Owner: `qa-auto2`; verifier: `qa-auto-ver`.
- Depends on: TASK-193, TASK-206.

## [OWNED TARGET PATHS]

`testframework/tests/conftest.py`, `testframework/pyproject.toml`, `testframework/selftests/**`, plus exact task-owned plugin modules under `testframework/src/hub_testkit/pytest_plugin/**`.

## [ACCEPTANCE]

- Strict markers and catalog validation run before execution.
- Required/optional profiles and exact skip reasons are explicit.
- Negative fixtures prove every gate fails then restores green.
- Reports publish TC/requirement/risk coverage, counts/duration/provenance/cleanup.
- Docker-only with resource limits; no leaked task resource.
