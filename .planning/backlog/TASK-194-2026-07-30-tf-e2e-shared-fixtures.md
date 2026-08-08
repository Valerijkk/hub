# TASK-194 — 🎭 Python Playwright foundation

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **137**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

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
row to Python. TASK-194 may not mutate `testframework/docker/run-task.ps1` or
`testframework/docker/images.lock.json`; either requires a separate exact
devops-owned contract change and independent review.

## [ACCEPTANCE]

- Pinned Python Playwright/browser image, per-run state and idempotent cleanup.
- Semantic locators/web-first assertions; no bare sleeps/shared developer state.
- Trace/screenshot/video/log redaction and retention policy with secret canary.
- Failing locator sentinel, repeated Docker run, unexpected skip=0 and cleanup PASS.
