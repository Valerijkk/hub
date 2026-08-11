# TASK-200 — ⚡ Safe load/performance/resilience baseline

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **140**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Root Python project contains a production-safe Locust harness and reproducible
non-production hot-path/resource/recovery baselines.

## [ROLE]

- Owners: `qa-auto2`, `devops-engineer`; verifier: `qa-auto-ver`.
- Depends on: TASK-205 and stable hot paths from Phases 134–135.

## [OWNED TARGET PATHS]

`testframework/load/**`,
`testframework/docker/task-200-load-resilience.compose.yml`,
`testframework/reports/<task-run>/**`. This card authorizes no path beyond its
live PLAN/manifest and routes no legacy non-unit row to Python. TASK-200 may not
create or modify any launcher API or entrypoint under `testframework/runner/**`
or any launcher under `testframework/docker/**`; only a separately approved
runner slice may do so. It may not modify `testframework/docker/images.lock.json`;
a separate exact devops-owned immutable-asset contract change and independent
review are required.

## [ACCEPTANCE]

- Hard-fail guard rejects production/public/unapproved target before users spawn.
- Profiles pin users/ramp/duration/warm-up/data/CPU/RAM/stop limits.
- p50/p95/p99, throughput, errors and resource saturation are reported for named SHA/image.
- Burst, slow dependency and restart/recovery scenarios have observable oracle.
- Secret/PII scan and task-resource cleanup pass; execution Docker-only.
