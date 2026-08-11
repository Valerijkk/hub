# TASK-227 — 🛡️ Security/platform risk closure

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **132**

> Runtime/traceability gate: launcher and runtime are `NOT_READY` / `NOT_RUN`.
> A future versioned launcher API may be separately implemented only under
> `testframework/runner/**` and used only after independent acceptance;
> `testframework/docker/**` supplies immutable execution assets only. No
> supported command, Docker-directory entrypoint, or direct Docker/Compose,
> host test/validator, mutable-image, or ad-hoc-cleanup fallback exists.
> Executable evidence must resolve TASK/requirement/TC/risk after acceptance.

## [OUTCOME]

Greenfield Python security/platform suite closes current P0/P1 authz, tenant,
audit and abuse risks. It does not replace legacy rows one-for-one.

## [ROLE]

- Owners: `qa-auto1`, `qa-auto2`; verifiers: `qa-auto-ver`, `security-auditor`.
- Depends on: TASK-208, TASK-209.

## [OWNED TARGET PATHS]

`testframework/tests/api/security/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- P0/P1 risk→TC coverage complete or explicit owner/date backlog.
- Happy/negative/authz/boundary/edge/failure-recovery/security/privacy applied.
- No real egress/secrets; canary/sentinel and repeated accepted-runner execution PASS.
- Unexpected skip=0 and resources cleaned.
