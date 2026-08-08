# TASK-230 — 🖥️ Frontend/static compatibility contracts in Python

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **136**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

New Python contracts protect frontend routes, bundle/link integrity, i18n,
docs links and static a11y prerequisites from current specs. Legacy guards are not copied.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `fe-ver`, `qa-auto-ver`.
- Depends on: TASK-206 and current frontend build artifact.

## [OWNED TARGET PATHS]

`testframework/tests/contract/frontend/**`, `testframework/reports/<task-run>/**`. Reusable validators receive exact task paths under `testframework/src/hub_testkit/contract/**`.

## [ACCEPTANCE]

- Observable artifact contracts preferred over brittle copied source strings.
- Deliberate route/locale/link/a11y fixture break fails the lane.
- TC trace and source/image SHA present; Docker-only and cleanup PASS.
