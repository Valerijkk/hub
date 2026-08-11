# TASK-230 — 🖥️ Frontend/static compatibility contracts in Python

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **136**

> Runtime/traceability gate: launcher and runtime are `NOT_READY` / `NOT_RUN`.
> A future versioned launcher API may be separately implemented only under
> `testframework/runner/**` and used only after independent acceptance;
> `testframework/docker/**` supplies immutable execution assets only. No
> supported command, Docker-directory entrypoint, or direct Docker/Compose,
> host test/validator, mutable-image, or ad-hoc-cleanup fallback exists.
> Executable evidence must resolve TASK/requirement/TC/risk after acceptance.

## [OUTCOME]

New Python contracts protect frontend routes, bundle/link integrity, i18n,
docs links and static a11y prerequisites from current specs. Legacy guards are not copied.

## [ROLE]

- Owner: `qa-auto2`; verifiers: `fe-ver`, `qa-auto-ver`.
- Depends on: TASK-206 and current frontend build artifact.

## [OWNED TARGET PATHS]

`testframework/tests/contract/frontend/**`, `testframework/reports/<task-run>/**`. Reusable validators receive exact task paths under `testframework/src/hub_testkit/contract/**`.

## [ACCEPTANCE]

- Observable artifact contracts preferred over brittle copied source strings.
- Deliberate route/locale/link/a11y fixture break fails the lane.
- TC trace and source/image SHA present; accepted-runner Docker-only execution and cleanup PASS.
