# TASK-198 — 🧪 Retain genuine frontend units

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **129**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Executable evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Only genuine Vitest/RTL units and their support move to
`frontend/tests/<domain>/**` / `frontend/tests/support/**`; static/source/docs
guards are deleted and later redesigned in Python contract lane.

## [ROLE]

- Owners: `fe1`, `fe2`; verifier: `fe-ver`.
- Depends on: Phase 128 `retain_frontend_units` manifest.

## [OWNED TARGET PATHS]

sealed `frontend/tests/<domain>/**` and `frontend/tests/support/**` targets plus only exact affected consumers listed by the Phase 129 batches. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Move set exactly equals manifest; no TS Playwright/static-contract rows retained.
- Vitest, TypeScript, Biome and build discover new path only and run in Docker.
- Non-zero collection, unexpected skip=0, temporary failure sentinel RED→GREEN.
- Old copies/link aliases have zero consumers before TASK-219 deletion.
- Evidence records counts, limits, image digest and resource cleanup.

## [OPERATING RULES]

No path-based duplicate compatibility tree. Domain-sized atomic commits;
Docker-only execution, no push/AI attribution.
