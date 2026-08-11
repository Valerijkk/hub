# TASK-229 — 🏢 Business/integration integrity closure

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **135**

> Runtime/traceability gate: launcher and runtime are `NOT_READY` / `NOT_RUN`.
> A future versioned launcher API may be separately implemented only under
> `testframework/runner/**` and used only after independent acceptance;
> `testframework/docker/**` supplies immutable execution assets only. No
> supported command, Docker-directory entrypoint, or direct Docker/Compose,
> host test/validator, mutable-image, or ad-hoc-cleanup fallback exists.
> Executable evidence must resolve TASK/requirement/TC/risk after acceptance.

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
- Sentinel, combined repeated accepted-runner execution, unexpected skip=0 and cleanup PASS.
