# TASK-220 — ✅ Final greenfield test architecture acceptance

> Revision: **2026-08-07** · Priority: **P1** · Status: Open · Phase: **141**

> Runtime/traceability gate: package, target suites and the future launcher are `NOT_READY`; runtime is `NOT_RUN`. The future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. No supported command, fallback or runtime evidence exists until a separate implementation and independent acceptance. Any future evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

Independent clean-room verification proves the new units+Python architecture,
live quality evidence and synchronized planning/docs.

## [ROLE]

- Owners: `qa-auto-ver`, `docs-writer`; final verifier: `team-lead`.
- Depends on: Phases 132–140 and TASK-233.

## [OWNED TARGET PATHS]

`AGENTS.md`, `docs/**`, `.planning/STATE.md`, `.planning/COVERAGE.md`, `.planning/audits/test-rebase/**` for final read/verification/docs handoff only. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Clean Docker run green for backend/frontend units and required Python layers.
- Path/language audit: legacy executable tests/runners=0; target violations=0.
- TC/requirement/task/risk orphans and unexpected skips=0; lane sentinels proven.
- Honest new unit coverage and requirement/risk coverage published without
  legacy-count or ≥82% preservation claim.
- STATE/COVERAGE/AGENTS/test docs point to current SHA/image evidence and post-rebase audit.
- Task resources cleaned; no push/deploy.
