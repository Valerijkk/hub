# TASK-208 — 🛡️ Admin/platform authorization matrix

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **132**

> Runtime/traceability gate: package, target suites and the future launcher are `NOT_READY`; runtime is `NOT_RUN`. The future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. No supported command, fallback or runtime evidence exists until a separate implementation and independent acceptance. Any future evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase frontend slug is kept only for
> link stability. Current purpose is the admin/platform authorization matrix.

## [OUTCOME]

Python API suite covers admin/platform/membership role×resource×action behavior
and tenant isolation from current requirements.

## [ROLE]

- Owner: `qa-auto1`; verifier: `qa-auto-ver`.
- Depends on: TASK-207.

## [OWNED TARGET PATHS]

`testframework/tests/api/security/**`, `testframework/tests/api/platform/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Allowed/denied matrix includes owner/admin/member/guest, last-owner/member,
  deactivated/banned actor, guessed/deleted/cross-tenant resource.
- State/version/duplicate/concurrency boundaries and audit event invariant covered.
- No existence/PII leakage in denied responses/artifacts.
- Docker-only repeated run, sentinel and cleanup PASS.
