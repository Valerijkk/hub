# TASK-194 — 🎭 Python Playwright foundation

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **137**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Root Python framework has reusable browser/auth/data/POM/a11y/artifact fixtures
for new requirement-driven E2E tests.

## [ROLE]

- Owner: `qa-auto1`; verifier: `qa-auto-ver`.
- Depends on: TASK-193, TASK-205, TASK-206.

## [OWNED TARGET PATHS]

`testframework/src/hub_testkit/browser/**`, `testframework/src/hub_testkit/domains/<domain>/ui/pages/**`,
`testframework/docker/task-194-e2e-foundation.compose.yml`. This card
authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit
row to Python. TASK-194 may not create or modify any launcher API or entrypoint
under `testframework/runner/**` or any launcher under `testframework/docker/**`;
only a separately approved runner slice may do so. It may not modify
`testframework/docker/images.lock.json`; a separate exact devops-owned
immutable-asset contract change and independent review are required.

## [ACCEPTANCE]

- Pinned Python Playwright/browser image, per-run state and idempotent cleanup.
- Semantic locators/web-first assertions; no bare sleeps/shared developer state.
- Trace/screenshot/video/log redaction and retention policy with secret canary.
- Failing locator sentinel, repeated Docker run, unexpected skip=0 and cleanup PASS.
