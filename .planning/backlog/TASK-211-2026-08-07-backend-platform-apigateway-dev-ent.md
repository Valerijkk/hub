# TASK-211 — 📨 Redis/NATS/jobs integration cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **133**

> Runtime/traceability gate: package, target suites and the future launcher are `NOT_READY`; runtime is `NOT_RUN`. The future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. No supported command, fallback or runtime evidence exists until a separate implementation and independent acceptance. Any future evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase backend-domain slug is kept
> only for link stability. Current purpose is Redis/NATS/jobs integration.

## [OUTCOME]

Greenfield Python integration tests protect cache/event/job delivery,
idempotency and recovery behavior.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `qa-auto-ver`, `be-ver`.
- Depends on: TASK-205.

## [OWNED TARGET PATHS]

`testframework/tests/integration/redis/**`, `testframework/tests/integration/nats/**`,
`testframework/docker/task-211-redis-nats-jobs.compose.yml`. This card
authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit
 row to Python. TASK-211 may not implement or modify any launcher API or
 entrypoint under `testframework/runner/**`, or place any launcher or entrypoint
 under `testframework/docker/**`; only a separately approved runner slice may do
 so. This card cannot claim runtime evidence.

## [ACCEPTANCE]

- Miss/stale/expiry, duplicate/out-of-order event, reconnect/restart, timeout,
  retry/dead-letter/backpressure and recovery cases where applicable.
- Required outage is asserted, never skipped; waits are bounded/observable.
- Ephemeral limited Docker services, pinned digests and full resource cleanup.
- TC trace, repeated run and sentinel PASS.
