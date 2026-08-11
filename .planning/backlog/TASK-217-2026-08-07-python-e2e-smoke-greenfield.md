# TASK-217 — 🎭 Greenfield risk-based Python E2E smoke

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **137**

> Runtime/traceability gate: package, target suites and the future launcher are `NOT_READY`; runtime is `NOT_RUN`. The future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. No supported command, fallback or runtime evidence exists until a separate implementation and independent acceptance. Any future evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Critical Python Playwright smoke journeys are designed from current
requirements/risks; legacy 3/20/53 counts are not inputs or acceptance criteria.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `qa-auto-ver`, `security-auditor`.
- Depends on: TASK-194, TASK-205, TASK-206.

## [OWNED TARGET PATHS]

`testframework/tests/e2e/smoke/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Login/session, navigation, tenant boundary, critical create/read and a11y
  sentinel journeys with stable TC IDs.
- Semantic locators, observable waits, persisted oracle and data cleanup.
- Non-zero, unexpected skip=0, broken-locator sentinel RED→GREEN.
- Immutable Docker app/browser images, redacted artifacts and cleanup PASS.
