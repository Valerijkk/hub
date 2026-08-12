# TASK-207 — 🔐 Auth/session/account greenfield cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **131**

> Runtime/traceability gate: package, target suites and the future launcher are `NOT_READY`; runtime is `NOT_RUN`. The future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. No supported command, fallback or runtime evidence exists until a separate implementation and independent acceptance. Any future evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase frontend-domain slug is kept
> only for link stability. Current purpose is auth/session/account cases.

## [OUTCOME]

Python API cases cover current authentication, session and account requirements
and risks from scratch; legacy case counts are not inputs.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `qa-auto-ver`, `security-auditor`.
- Depends on: TASK-205.

## [OWNED TARGET PATHS]

`testframework/tests/api/auth/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Happy login/session/logout plus malformed/expired/replayed/revoked tokens.
- Normalization/duplicate/boundary/rate-limit/deactivated user and concurrent session cases.
- Dependency timeout/partial failure/recovery where supported.
- TC metadata complete, unexpected skip=0, sentinel RED→GREEN, Docker resources cleaned.
