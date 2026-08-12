# TASK-219 — 🧹 Immediate exact-manifest legacy deletion

> Revision: **2026-08-07 founder authorization** · Priority: **P1** · Status: Open · Phase: **129**

> Runtime/traceability gate: package, target suites and the future launcher are `NOT_READY`; runtime is `NOT_RUN`. The future versioned launcher API belongs only to `testframework/runner/**`; `testframework/docker/**` contains immutable execution assets only. No supported command, fallback or runtime evidence exists until a separate implementation and independent acceptance. Any future evidence resolves requirement `TASK-219`, stable deletion-guard TC IDs and over-delete/protected-row/consumer-drift risks; broad globs, direct Docker/Compose, host test/validator, mutable-image and ad-hoc cleanup fallbacks are forbidden.

## [OUTCOME]

All legacy non-unit tests and the independently proven zero-consumer support,
configs, linkers and runners are deleted immediately after the hook-free units,
frontend units and scaffold exist. Export hooks/support still consumed by the
245 protected genuine units are removed only after TASK-234. Legacy counts do not block deletion.

## [ROLE]

- Owner: `devops-engineer`; verifier: `code-reviewer` + `qa-auto-ver` for
  per-batch manifest/entrypoint evidence.
- Depends on: Phase 128 manifests and TASK-193/197/198; hook-pruning tail also
  depends on TASK-234.

## [OWNED TARGET PATHS]

only exact `delete_legacy` and independently proven zero-consumer operations, affected-consumer paths and records in `.planning/audits/test-rebase/cutover-manifest.json`; no broad directory ownership. This card authorizes no path beyond its live PLAN/manifest and routes no legacy non-unit row to Python.

## [ACCEPTANCE]

- Deleted set equals reviewed `delete_legacy` exact paths; extra/missing=0.
- Each deletion batch has a real TASK-219 batch ID in
  `.planning/audits/test-rebase/cutover-manifest.json`; fictional task IDs or
  untracked subcards are forbidden.
- None of the 245 `KEEP-GO-UNIT-HOOK-DEPENDENT` rows or their live dependencies
  enters a delete batch; negative sentinel proves this guard.
- Path guard rejects traversal/absolute/UNC/backslash/symlink/protected targets.
- Old copies absent; obsolete consumer refs=0; retained unit/Python collections non-zero.
- TASK-201 switches every CI/script/docs consumer affected by the batch in the
  same commit.  `removed_path_reference_count` must be zero before the batch is
  independently marked `verified`.
- Small atomic batches; Docker gate after each; Git revert is tested rollback plan.
- No broad glob deletion. All task resources cleaned/TTL-handoff recorded.

## [OPERATING RULES]

Founder authorized this strategy, but dirty overlap/unmanifested path stops that
batch. Destructive work remains blocked until the Phase 128 seal, exact
manifest, immutable execution assets, separately implemented and independently
accepted launcher API under `testframework/runner/**`, and independent
verification are all present. Do not recover deleted tests as worktree copies;
Git history is recovery. No push/AI attribution. The final hook/support closure
remains within the real TASK-219 card and depends on TASK-234; no synthetic task
card is allowed.
