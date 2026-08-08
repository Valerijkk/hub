# TASK-231 — ♿ Accessibility coverage

> Created: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **139**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Python E2E/contract suites protect applicable WCAG keyboard, focus, semantic,
error and contrast behavior on critical journeys.

## [ROLE]

- Owner: `qa-auto1`; verifiers: `qa-auto-ver`, `designer-ver`.
- Depends on: TASK-217; Phase 138 journeys.

## [OWNED TARGET PATHS]

`testframework/tests/e2e/accessibility/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Keyboard-only, focus order/return/trap, name/role/state, error association/
  live region and representative contrast cases are TC-traceable.
- Axe scope/exceptions explicit; manual-only gaps have owner/date.
- Deliberate a11y violation sentinel, Docker run and cleanup PASS.
