# TASK-205 — 🐍 Python core, health and readiness foundation

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **131**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase `health-auth-slice` slug is kept
> for link stability. Current purpose is Python core plus health/readiness;
> auth/session cases belong to TASK-207.

## [OUTCOME]

Typed settings/client, isolation, redaction, data factories and health/readiness
cases establish the greenfield Python API foundation.

## [ROLE]

- Owner: `qa-auto1`; verifier: `qa-auto-ver`.
- Depends on: TASK-193, TASK-206, TASK-218.

## [OWNED TARGET PATHS]

`testframework/src/hub_testkit/domains/health/api.py`, `testframework/tests/api/health/**`,
`testframework/docker/task-205-core-health.compose.yml`. This card authorizes
no path beyond its live PLAN/manifest and routes no legacy non-unit row to
Python. TASK-205 may not create or modify any launcher API or entrypoint under
`testframework/runner/**` or any launcher under `testframework/docker/**`; only
a separately approved runner slice may do so. It may not modify
`testframework/docker/images.lock.json`; a separate exact devops-owned
immutable-asset contract change and independent review are required.

## [ACCEPTANCE]

- Fail-closed Docker readiness and per-run user/data namespace with idempotent cleanup.
- Status+schema+business-oracle helpers; tokens/cookies/PII redacted.
- Liveness and dependency readiness/failure/recovery are distinguished.
- All cases have TC/requirement/task/risk metadata and scenario checklist.
- Two clean Docker runs, sentinel RED→GREEN, unexpected skip=0 and resources cleaned.
