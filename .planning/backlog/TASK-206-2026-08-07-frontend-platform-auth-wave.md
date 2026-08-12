# TASK-206 — 🧭 Requirement/risk/test-case catalog

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **130**

> Runtime/traceability gate: catalog review is static; the future versioned launcher API belongs only to `testframework/runner/**` and `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase frontend slug is kept only for
> link stability. Current purpose is the requirement/risk/TC catalog.

## [OUTCOME]

Machine-readable catalog links each `TC-*` to current requirement/spec, change
task, protected risk, layer/domain/priority/owner, oracle, data and cleanup.

## [ROLE]

- Owners: `ba1`, `qa-auto1`; verifiers: `ba-ver`, `qa-auto-ver`.
- Depends on: Phase 129.

## [OWNED TARGET PATHS]

`testframework/src/hub_testkit/domains/<domain>/ui/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Duplicate/orphan/unknown/stale links fail validation.
- Happy, negative, authn/authz/tenant, boundary, edge/corner,
  failure/recovery/idempotency, security/privacy, a11y/performance are required
  when applicable; each `N/A` has rationale.
- P0/P1 uncovered requirement/risk produces explicit backlog owner/date, not silent green.
- Docker-only validator negative mutations RED→GREEN; resources cleaned.
