# TASK-197 — 🧪 Retain genuine backend units

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **129**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

The exact 350-row hook-free cohort of proven service-free Go units is moved to
`backend/tests/<domain>/**`. The separate 245-row hook-dependent cohort is
protected for TASK-234; component/API/integration/source guards are not retained.

## [ROLE]

- Owners: `be1`, `be2` by non-overlapping domains; verifier: `be-ver`.
- Depends on: Phase 128 founder manifest partition
  `MOVE-GO-UNIT-HOOK-FREE`.

## [OWNED TARGET PATHS]

sealed `backend/tests/<domain>/**` targets plus only the exact source/consumer/manifest paths listed by the Phase 129 batches. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Move set exactly equals 350 rows / 1264 declared cases in the pinned manifest;
  no DB/broker/network/container dependency.
- External-package semantics use only justified production API; no new test-only hook.
- Docker collection/run is non-zero and green; race/build where applicable.
- Post-move coverage is reported as new honest unit baseline; ≥82% is not a blocker.
- Old copies are absent and resource cleanup is recorded.

## [OPERATING RULES]

No semantic mass rewrite during this hook-free move. Triage failures by spec,
do not restore component tests for coverage. Do not touch/delete TASK-234 rows.
Domain-sized atomic commits; Docker-only, no push.
