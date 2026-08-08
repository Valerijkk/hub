# TASK-209 — 🛡️ Security abuse and isolation cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **132**

> Runtime/traceability gate: `NOT_READY` until `testframework/docker/run-task.ps1` contract `hub.qa-runner/v1` and `testframework/docker/images.lock.json` are independently accepted. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase backend pilot slug is kept only
> for link stability. Current purpose is security abuse and isolation cases.

## [OUTCOME]

Risk-driven Python cases detect IDOR, privilege escalation, enumeration,
replay, rate-limit bypass and secret/PII leakage.

## [ROLE]

- Owner: `qa-auto2`; verifier: `security-auditor`.
- Depends on: TASK-205, TASK-208.

## [OWNED TARGET PATHS]

`testframework/tests/api/security/**`, `testframework/tests/api/platform/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Abuse matrix covers valid-looking foreign IDs, stale/replayed credentials,
  concurrency and boundary-rate behavior.
- Security oracle checks observable denial and non-disclosure, not status alone.
- Secret canary proves log/report/trace scanner sensitivity.
- No real egress/secrets; Docker-only, unexpected skip=0, cleanup PASS.
