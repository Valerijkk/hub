# TASK-204 — 📐 Greenfield test architecture and cutover contract

> Revision: **2026-08-07 founder override** · Priority: **P1** · Status: Open · Phase: **128**

> Runtime/traceability gate: the future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. Launcher status remains `NOT_READY` and runtime remains `NOT_RUN` until a separate implementation and independent acceptance. Every executable case/evidence must resolve TASK/requirement/TC/risk; direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

ADR/validator encode immediate exact-allowlist cutover: retain only genuine
units, delete legacy non-units/runners and build new Python QA from requirements/risks.

## [ROLE]

- Owner: `arch`; verifier: `arch-ver` + `qa-auto-ver`.
- Depends on: TASK-202 current-SHA ledger. TASK-203 is diagnostic, not blocker.

## [OWNED TARGET PATHS]

`.planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md`, `.planning/audits/test-rebase/test-architecture-contract.json`, `.planning/audits/test-rebase/cutover-manifest.json` and their bounded validator/docs records only. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Create `.planning/audits/test-rebase/cutover-manifest.json` before Phase 129.
  It pins the current source revision, ledger/founder-manifest digests, ordered
  batch IDs, exact move/delete/edit paths, the CI/docs consumers switched in
  each batch, Docker evidence URI, verifier verdict and revert commit.  An
  empty or missing batch field fails closed.
- Contract target paths match root Python architecture and protected-zone exclusion.
- Move/keep/delete row partitions are exact, disjoint and exhaustive over the
  1,049-row pinned ledger; ambiguous rows have human rationale.
- Contract freezes 350 hook-free Go units as immediate move and 245
  hook-dependent genuine units as protected keep until TASK-234 refactor/move.
- Validator rejects traversal/absolute/UNC/backslash/symlink/protected paths,
  unknown categories and non-exhaustive sets.
- Legacy row/file/case comparison requirements are removed. Go ≥82% is explicitly non-blocking;
  post-cutover honest unit baseline remains required evidence.
- Positive and negative validator runs execute only in limited Docker and clean resources.
- The cutover manifest starts with every batch `planned`; Phase 129 owners may
  change only their batch to `applied` and the independent verifier alone may
  change it to `verified`.  Recalculation on a new SHA invalidates prior batch
  authority until the exact partitions are reviewed again.

## [OPERATING RULES]

Do not modify/delete tests in this task. No protected-zone dependency, no push.
