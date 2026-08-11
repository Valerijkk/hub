---
phase: 129-tf-e2e-shared
plan: "00"
type: execute
wave: 0
depends_on:
  - 128-tf-python-scaffold
files_modified:
  - .planning/phases/129-tf-e2e-shared/129-01-PLAN.md
  - .planning/phases/129-tf-e2e-shared/129-02-PLAN.md
  - .planning/phases/129-tf-e2e-shared/129-03-PLAN.md
  - .planning/phases/129-tf-e2e-shared/129-04-PLAN.md
  - .planning/phases/129-tf-e2e-shared/129-05-PLAN.md
  - .planning/phases/129-tf-e2e-shared/129-06-PLAN.md
  - .planning/phases/129-tf-e2e-shared/129-07-PLAN.md
autonomous: false
requirements: [TASK-193, TASK-197, TASK-198, TASK-201, TASK-204, TASK-219, TASK-234]
must_haves:
  truths:
    - "Phase 129 remains blocked until Phase 128 has sealed the 1,049-row authority, the descriptor-relative no-follow executor, its negative sentinels, and an independent HIGH-free acceptance."
    - "The sealed ledger identity is 1,049 rows at SHA-256 72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511."
    - "Exactly 507 initially actionable moves, 245 initially blocked hook-dependent Go unit releases/moves, 297 ledger deletes, 28 non-ledger deletes, and 43 hook-file deletes are represented by 173 explicit planned batch tasks; IDs become authoritative only after the Phase-128 seal."
    - "Every planned batch has at most 56 literal operations and maps to one task; it remains non-executable until exact Phase-128 seal equality and separate future runner implementation plus independent acceptance."
    - "Every batch uses one sealed apply-and-verify transaction: an owner-authorized Linux writer applies the exact Phase-128 batch, emits a mutation receipt and rollback handle, then a fresh conclusion-blind read-only verifier records separate repository-root and focused-gate evidence."
    - "At completion, genuine Go and frontend units live only in their target test roots, the legacy non-unit/support closure has zero consumers, and all 43 hook files are deleted only after zero-consumer proof."
  prohibitions:
    - "No legacy non-unit test is ported, translated, parity-matched, or dual-run; historical Go coverage is not a migration gate."
    - "No cutover mutation or validator runs on the Windows host; future orchestration authority belongs only to a separately accepted API under testframework/runner/**, while testframework/docker/** contains immutable assets only. Current launcher/runtime status is NOT_READY / NOT_RUN."
    - "No operation reads or writes pro-qa-ai/**."
    - "Plan 129-01 never creates, regenerates, or repopulates cutover-manifest.json."
    - "No batch task hides multiple batch IDs in a loop or stages paths outside its manifest allowlist."
  artifacts:
    - "testframework package roots, src/hub_testkit/** reusable code, tests/{contract,api,integration,e2e}/**, selftests/**, cases/**, schemas/**, policies/**, load/**, immutable docker/{Dockerfile.core,images.lock.json}, future runner/** ownership declarations, and reports/<task-run>/** boundaries; package, launcher, suites, and runtime remain separately NOT_READY / NOT_RUN"
    - "backend/tests/<domain>/** containing the 350 hook-free and subsequently released Go units"
    - "frontend/tests/<domain>/** plus frontend/tests/support/render.tsx containing 156 frontend units and one render helper"
    - ".planning/audits/test-rebase/cutover-manifest.json with per-batch transitions and independent evidence"
    - "Updated root consumer registry, CI, scripts, and test-layout documentation with no stale legacy path"
  key_links:
    - "founder-greenfield-cutover-manifest.json row partitions -> sealed cutover-manifest batches -> apply_cutover_batch.py literal operations"
    - "apply_cutover_batch.py per-operation identity/no-follow checks -> atomic Git owner commit -> independent verifier evidence"
    - "each batch mutation -> repository-root manifest validation in its sealed validation mode -> focused Docker-native tests"
    - "hook consumer refactor -> regenerated authority with zero links -> hook-file deletion"
---

<objective>
Execute the founder-locked greenfield test-architecture cutover as 173 small, manifest-authorized batches after Phase 128 has sealed the authority and destructive executor.

Purpose: relocate only genuine native units, establish the root Python QA scaffold, delete the complete legacy non-unit/support closure, eliminate private test hooks through production seams, and keep every destructive step reversible and independently auditable.

Output: seven ordered execution plans containing 176 explicit planned task definitions: three non-destructive/scaffold tasks and 173 literal manifest-batch tasks. Runtime remains NOT_READY until their stated dependencies and independent receipts exist.
</objective>

<execution_context>
@C:/Users/Valerii/.codex/gsd-core/workflows/execute-plan.md
@C:/Users/Valerii/.codex/gsd-core/templates/summary.md
@AGENTS.md
@.cursor/REGULATION.md
@.cursor/rules/07-agent-safety.mdc
@.cursor/rules/70-pro-qa-ai.mdc
@.cursor/playbooks/team-workflow.md
@.cursor/playbooks/prompt-contract.md
</execution_context>

## Execution gate

Phase 129 runtime is **NOT_READY** merely because these plans exist. Plan 129-01 declares non-executable requirements only; it does not create a launcher, runner entrypoint, API schema/version, or supported invocation. The future API belongs solely under absent `testframework/runner/**`; `testframework/docker/**` retains only immutable `Dockerfile.core` and `images.lock.json`. Runtime remains **NOT_RUN** until a separate runner slice is specified, implemented, independently accepted, and the affected authority/receipt graph is atomically resealed. The descriptor-relative executor, manifest equality, negative sentinels, and independent HIGH-free TASK-204 acceptance remain mandatory before any batch mutation.

Native Windows junction/reparse metadata remains unproven. Therefore every destructive operation is authorized only inside a constrained Linux container. There is no host-validator fallback.

## Locked cutover arithmetic

| Cohort | Rows/paths | Batch tasks |
|---|---:|---:|
| Hook-free Go unit moves | 350 | 32 |
| Frontend unit moves | 156 | 12 |
| Frontend render-support move | 1 | 1 |
| Go component/API/integration ledger deletes | 237 | 33 |
| Backend non-ledger support deletes | 4 | 1 |
| Static-contract deletes | 24 | 7 |
| Compile-fixture deletes | 2 | 2 |
| TypeScript Playwright spec deletes | 20 | 9 |
| TypeScript Playwright support deletes | 14 | 1 |
| Remaining non-ledger support/config/script/doc deletes | 24 | 5 |
| Hook-dependent Go unit releases and moves | 245 | 31 |
| Zero-consumer hook-file deletes | 43 | 39 |
| **All Phase-129 operations** | **832 initially actionable row/path operations + 245 blocked-unit releases/moves + 43 hook-file deletes = 1,120** | **173** |

The 245 hook-dependent units are initially `keep-blocked`; their 31 tasks refactor public/injected seams, regenerate current authority, and move only rows proven to have zero hook consumers. The baseline of 43 hook files / 582 symbols / 892 links and the protected subset of 42 distinct hook files / 434 referenced symbols / 653 links are classification facts, not drift.

## Wave and dependency graph

| Wave | Plan | Explicit tasks | Depends on | Outcome |
|---:|---|---:|---|---|
| 1 | 129-01 | 3 | Phase 128 seal | Declare future runner/scaffold requirements and two independent preflight gates; no implementation or invocation |
| 2 | 129-02 | 32 | 129-01 | Move 350 hook-free Go units |
| 3 | 129-03 | 13 | 129-02 | Move 156 frontend units and render helper |
| 4 | 129-04 | 34 | 129-03 | Delete 237 Go non-units and four backend support files |
| 5 | 129-05 | 24 | 129-04 | Delete 60 frontend/TypeScript ledger files and 24 remaining non-ledger files |
| 6 | 129-06 | 31 | 129-05 | Remove private-hook dependencies and move 245 Go units |
| 7 | 129-07 | 39 | 129-06 | Delete 43 zero-consumer production hook files and close CI/docs/manifest |

All plans share the manifest and consumer registry, so the sequence is intentionally linear. Within a plan, batch tasks are also ordered because each changes the same manifest. No same-wave parallelism is safe.

## TASK ownership

| Plan | Requirement(s) | Owner | Independent verifier |
|---|---|---|---|
| 129-01-A declarative runner/scaffold requirements | TASK-193; TASK-201 | devops-engineer | qa-auto-ver |
| 129-01-B read-only preflight | TASK-204 | qa-auto-ver | gsd-plan-checker |
| 129-01-C independent preflight verification | TASK-204 | tech-lead | gsd-plan-checker |
| 129-02 | TASK-197 | be1 | be-ver |
| 129-03 | TASK-198 | fe1 | fe-ver |
| 129-04 | TASK-219 | devops-engineer | code-reviewer + qa-auto-ver |
| 129-05 | TASK-219 primary deletion authority; TASK-201 secondary same-commit consumer/CI cleanup on four support batches | devops-engineer | code-reviewer + qa-auto-ver |
| 129-06 | TASK-234 | be1 | be-ver |
| 129-07 hook batches | TASK-219 primary deletion authority; TASK-234 prerequisite zero-consumer/refactor evidence | devops-engineer | code-reviewer + qa-auto-ver |
| 129-07 final hook batch / runner-CI closure | TASK-219 primary deletion authority; TASK-234 prerequisite evidence; TASK-201 secondary same-commit consumer/CI cleanup | devops-engineer | code-reviewer + qa-auto-ver |

These are real ROADMAP task IDs, not plan-local aliases. Every numbered task repeats its exact requirement, owner, and verifier.

## Future runner acceptance boundary

No supported invocation exists. The future versioned outer API, profile resolution, orchestration, first-causal-exit preservation, task-owned report lifecycle, and exact cleanup belong only to a separately specified implementation under `testframework/runner/**`. `testframework/docker/**` owns immutable execution assets only, including `Dockerfile.core` and `images.lock.json`; it owns no launcher or entrypoint.

Every numbered task therefore carries a non-executable 11-field requirements block. Those declarations preserve plan/task/batch identity, expected operations, owner/verifier roles, manifest and image-lock dependencies, mode, and inner/outer deadlines, but grant no mutation or execution authority. Launcher status remains **NOT_READY**, runtime remains **NOT_RUN**, and direct Docker/Compose, host execution, aliases, shims, or fallback paths remain forbidden until a separate runner slice is implemented and independently accepted.

A future accepted runner must bind exact manifest-owned paths and immutable image provenance; create only task-owned reports after validating the owner marker; isolate the authorized writer from a fresh conclusion-blind verifier; preserve the first causal exit; enforce inner deadline shorter than outer deadline; and prove unconditional exact-resource cleanup on success, failure, timeout, and interruption. These are acceptance requirements, not an implemented command surface.

## Batch contract

For every batch task:

1. The task names one unique **planned** batch ID and one literal expected operation count. No task selects a domain with a runtime loop.
2. Before mutation, assert the Phase-128 receipt, current Git HEAD, inventory identity, ledger SHA, batch state, expected operation count, exact sealed owned paths, clean-worktree non-overlap, and unchanged protected-zone exclusion.
3. `<action>` contains no runtime command. The adjacent structured block preserves the task's declarative execution requirements; no host/manual mutation, executable span, or owner/verifier invocation is authorized while the future runner is absent.
4. A future independently accepted runner transaction must separate the owner-authorized writer from the fresh conclusion-blind verifier, limit writes/staging to the sealed allowlist, preserve one atomic owner commit and rollback handle, and emit independent evidence. This is a future acceptance condition, not current authority.
5. The future independent verifier must receive no writer conclusion, mount mutation outputs read-only, apply the sealed root and focused gates, and emit a separate result; current runtime remains `NOT_RUN`.
6. Any HIGH finding, incomplete declarative requirement/evidence, manifest/consumer drift, unexpected dirty overlap, operation mismatch, timeout, cleanup failure, or rollback ambiguity keeps the batch `NOT_READY` and halts advancement.

## Artifacts this phase produces

- Future files after separate authorization: the accepted Python scaffold and runner-owned implementation under `testframework/runner/**`; immutable `testframework/docker/{Dockerfile.core,images.lock.json}` assets; relocated unit files; per-batch evidence/state; and explicitly listed consumer/CI/docs edits. This rebase creates none of them.
- New symbols: typed fail-closed Python runtime configuration, marker/traceability registration, shared fixtures and redaction helpers; public or injected production seams explicitly replacing private test hooks in Plan 129-06.
- New CLI flags: none are invented by Phase 129. It consumes only the Phase-128-sealed executor modes and inputs as immutable declarative authority; this plan publishes no callable syntax.
- New manifest fields: none are invented by Phase 129. It transitions the Phase-128 schema's existing batch state, owner commit, verifier commit, evidence, consumer status, and rollback fields.
- Product UI artifacts: None.

## Source coverage audit

| Source | Item | Coverage |
|---|---|---|
| GOAL | Founder greenfield cutover and root Python QA project | Plans 129-01 through 129-07 |
| REQ | ROADMAP tasks | TASK-193, TASK-197, TASK-198, TASK-201, TASK-204, TASK-219, TASK-234; repeated in numbered frontmatter and tasks |
| SPEC | Product SPEC / requirement probe | No product SPEC applies; ROADMAP TASK IDs and the founder greenfield canon are the executable requirements |
| UI-SPEC | UI design contract | Not applicable: no product UI is designed or changed |
| RESEARCH | RESEARCH.md / CONTEXT.md | Intentionally skipped; ADR, founder manifest, current ledger, and backlog decisions are already pinned |
| CONTEXT | Founder locks in ADR/manifest/task artifacts | Fully mapped by the arithmetic, batch contract, and prohibitions above |

No deferred idea is planned. No source item is silently omitted.

<threat_model>
ASVS level: L1. Any unresolved HIGH threat blocks the current batch and every later plan.

## Trust boundaries

| Boundary | Description |
|---|---|
| Sealed descriptor -> executor | Untrusted path and identity data becomes filesystem authority |
| Windows host -> Linux cutover container | Host may orchestrate Docker but has no destructive authority |
| Writable repository bind -> protected/unrelated worktree | Only manifest allowlisted paths may change |
| Owner commit -> independent verifier | Executor evidence is not self-attestation |
| Batch mutation -> CI/docs/consumer registry | Layout changes must remain globally reachable and current |

## STRIDE threat register

| Threat ID | Category | Component | Severity | Disposition | Mitigation plan |
|---|---|---|---|---|---|
| T-129-TRAV | Tampering | path parser | high | mitigate | Reject traversal, dot segments, absolute/drive/UNC paths, ADS colons, trailing dot/space, control characters, and Windows reserved names before opening a descriptor. |
| T-129-LINK | Elevation of privilege | filesystem resolver | high | mitigate | Linux-only openat-style no-follow resolution; reject symlink, junction, mount/reparse ambiguity, and non-regular sources. |
| T-129-HARD | Tampering | source identity | high | mitigate | Require link count one and sealed size/SHA/device/inode identity immediately before every operation. |
| T-129-TOCTOU | Tampering | mutation boundary | high | mitigate | Revalidate the full manifest and reopen/recheck identity at each operation; fail closed on any change. |
| T-129-UNICODE | Spoofing | portable path key | high | mitigate | Enforce canonical NFC plus portable case/Unicode key uniqueness across sources, targets, consumers, and protected prefixes. |
| T-129-COLLIDE | Tampering | move destination | high | mitigate | Reject existing destinations, duplicate targets, parent aliases, and source/destination collisions before mutation. |
| T-129-DRIFT | Tampering | manifest/consumer registry | high | mitigate | Bind inventory SHA, source revision, HEAD, builder/validator hashes, batch count, and all affected consumers; root validator runs after every batch. |
| T-129-DIRTY | Tampering | dirty worktree | high | mitigate | Compute batch allowlist first and halt on any pre-existing dirty overlap; preserve all unrelated user changes. |
| T-129-PRO | Information disclosure | pro-qa-ai/** | high | mitigate | Protected-prefix rejection in builder, validator, and executor; zero read/write authority without founder handoff. |
| T-129-RESOURCE | Denial of service | Docker execution | medium | mitigate | Explicit CPU, RAM, swap, PID, network, read-only-root, tmpfs, name, and automatic exact-resource removal constraints on every future runner transaction. |
| T-129-EVIDENCE | Repudiation | batch receipt | high | mitigate | Require command, exit code, duration, image digest, HEAD, counts, cleanup, owner/verifier identities, and independent verdict before advancement. |
| T-129-ROLLBACK | Repudiation | recovery | high | mitigate | One allowlisted owner mutation commit per batch; rollback only by ordered git revert and root revalidation. |
| T-129-SUPPLY | Tampering | package/image inputs | high | mitigate | Use pinned images and existing lock/checksum files only; reject dependency or lock drift and record image digests. |
</threat_model>

<verification>
The checker must confirm frontmatter/structure, 176 explicit tasks, 173 unique planned batch IDs, maximum 56 operations, 1,120 total operations, exact TASK/owner/verifier mappings, one adjacent ordered 11-field non-executable block per task, exact command-free readiness body, future `runner/**` ownership, immutable-assets-only `docker/**`, exact Phase-128 seal equality, and absence of port/parity/dual-run or positive invocation semantics.
</verification>

<success_criteria>
The seven numbered files are schema-complete but **NOT_READY** and runtime is **NOT_RUN** while the Phase-128 seal, immutable image lock, separately implemented runner, independent acceptance, and atomic receipt reseal are absent. Declarative planning never by itself authorizes execution.
</success_criteria>

<output>
Each numbered plan creates its corresponding `.planning/phases/129-tf-e2e-shared/129-NN-SUMMARY.md` when executed.
</output>
