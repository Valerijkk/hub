# TASK-232 — 🔒 Cross-suite security/privacy audit

> Created: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **139**

> Runtime/traceability gate: launcher and runtime are `NOT_READY` / `NOT_RUN`.
> A future versioned launcher API may be separately implemented only under
> `testframework/runner/**` and used only after independent acceptance;
> `testframework/docker/**` supplies immutable execution assets only. No
> supported command, Docker-directory entrypoint, or direct Docker/Compose,
> host test/validator, mutable-image, or ad-hoc-cleanup fallback exists.
> Executable evidence must resolve TASK/requirement/TC/risk after acceptance.

## [OUTCOME]

Risk-driven Python cases and artifact scans close OWASP/privacy gaps across
API, integration and E2E layers.

## [ROLE]

- Owners: `security-auditor`, `qa-auto2`; verifier: `qa-auto-ver`.
- Depends on: TASK-227 and Phases 134–138.

## [OWNED TARGET PATHS]

`testframework/tests/api/security/**`, `testframework/tests/contract/security/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Applicable injection, unsafe URL/content/upload, CSRF/CORS/cookies/headers,
  authz and resource-abuse risks have TC/evidence or owner/date.
- Secret/PII canary proves log/trace/report scan sensitivity.
- No real egress/secrets; accepted-runner Docker-only execution, unexpected skip=0 and cleanup PASS.
