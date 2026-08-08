# TASK-232 — 🔒 Cross-suite security/privacy audit

> Created: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **139**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

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
- No real egress/secrets; Docker-only, unexpected skip=0 and cleanup PASS.
