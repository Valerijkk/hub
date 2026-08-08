# TASK-216 — 📜 API/schema/config contract suite

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **136**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase integration slug is kept only
> for link stability. Current purpose is API/schema/config contracts.

## [OUTCOME]

Python contract tests protect public API/schema/route/config/database invariants
from current specs, not copied legacy source guards.

## [ROLE]

- Owner: `qa-auto1`; verifiers: `arch-ver`, `qa-auto-ver`.
- Depends on: TASK-206 and current specs/artifacts.

## [OWNED TARGET PATHS]

`testframework/tests/contract/api/**`, `testframework/tests/contract/schema/**`, `testframework/reports/<task-run>/**`. Reusable validators receive exact task paths under `testframework/src/hub_testkit/contract/**`.

## [ACCEPTANCE]

- Required/optional/null/enum/format/status/error/security compatibility covered.
- DB/config/migration ownership/index/unsafe-default invariants covered.
- Deliberate incompatible fixtures fail; source SHA recorded.
- Docker-only, no protected-zone/network discovery, cleanup PASS.
