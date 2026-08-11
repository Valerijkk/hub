# TASK-215 — 🔌 External adapter and data-transfer cases

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **135**

> Runtime/traceability gate: package, target suites and the future launcher are `NOT_READY`; runtime is `NOT_RUN`. The future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. No supported command, fallback or runtime evidence exists until a separate implementation and independent acceptance. Any future evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.
> **Deprecated filename alias:** the pre-rebase backend slug is retained for
> link stability; current purpose is external-adapter/data-transfer behavior.

## [OUTCOME]

Python tests cover GitHub/GitLab/1C/mail/marketplace/import/export/backup
contracts and failure/recovery without real egress/secrets.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `qa-auto-ver`, `security-auditor`.
- Depends on: TASK-205 and applicable Phase 133 fixtures.

## [OWNED TARGET PATHS]

`testframework/tests/integration/adapters/**`, `testframework/reports/<task-run>/**`. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Local deterministic doubles cover timeout, 4xx/5xx, rate-limit, malformed,
  partial, replay, retry and recovery.
- Import/export/backup cover oversized/duplicate/corrupt data and rollback integrity.
- Outbound requests, redaction and tenant ownership asserted.
- Docker egress denied where possible; sentinel and cleanup PASS.
