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
    - "Every planned batch has at most 64 literal operations and maps to one task; it becomes executable only after exact ID/count/owned-path equality with the Phase-128-sealed manifest and launcher contract v1 acceptance."
    - "Every batch uses one sealed apply-and-verify transaction: an owner-authorized Linux writer applies the exact Phase-128 batch, emits a mutation receipt and rollback handle, then a fresh conclusion-blind read-only verifier records separate repository-root and focused-gate evidence."
    - "At completion, genuine Go and frontend units live only in their target test roots, the legacy non-unit/support closure has zero consumers, and all 43 hook files are deleted only after zero-consumer proof."
  prohibitions:
    - "No legacy non-unit test is ported, translated, parity-matched, or dual-run; historical Go coverage is not a migration gate."
    - "No cutover mutation or validator runs on the Windows host; the sole runtime boundary is launcher contract hub.qa-runner/v1, and the host is limited to reads, text edits, Git, and its one canonical launcher call."
    - "No operation reads or writes pro-qa-ai/**."
    - "Plan 129-01 never creates, regenerates, or repopulates cutover-manifest.json."
    - "No batch task hides multiple batch IDs in a loop or stages paths outside its manifest allowlist."
  artifacts:
    - "testframework/pyproject.toml, lock, pytest configuration, conftest.py, README.md, core/, api/, integration/, contract/, e2e/, load/, docker/, and reports/ scaffold artifacts, including launcher v1, immutable images.lock.json, and launcher contract/negative-cleanup tests"
    - "backend/tests/<domain>/** containing the 350 hook-free and subsequently released Go units"
    - "frontend/tests/<domain>/** plus frontend/tests/support/render.tsx containing 156 frontend units and one render helper"
    - ".planning/audits/test-rebase/cutover-manifest.json with per-batch transitions and independent evidence"
    - "Updated root consumer registry, CI, scripts, and test-layout documentation with no stale legacy path"
  key_links:
    - "founder-greenfield-cutover-manifest.json row partitions -> sealed cutover-manifest batches -> apply_cutover_batch.py literal operations"
    - "apply_cutover_batch.py per-operation identity/no-follow checks -> atomic Git owner commit -> independent verifier evidence"
    - "each batch mutation -> repository-root validate_cutover_manifest.py --check -> focused Docker-native tests"
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

Phase 129 runtime is **NOT_READY** merely because these plans exist. Plan 129-01 must find, without changing, a Phase-128-sealed `.planning/audits/test-rebase/cutover-manifest.json`, Phase-128-owned `apply_cutover_batch.py`, and `validate_cutover_manifest.py`. Its first, explicitly non-runtime bootstrap substep creates `testframework/docker/run-task.ps1`, `testframework/docker/images.lock.json`, and their contract/negative-cleanup tests; only then may qa-auto-ver perform the sole canonical bootstrap-acceptance invocation. The executor must be descriptor-relative, use no-follow directory descriptors, revalidate manifest state and source identity immediately before every operation, reject link count other than one, and fail closed on every path/identity/collision ambiguity. Launcher/scaffold acceptance plus executor negative sentinels and an independent HIGH-free TASK-204 acceptance must be recorded before any batch mutation.

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
| 1 | 129-01 | 3 | Phase 128 seal | Statically bootstrap and accept launcher/scaffold, then independently revalidate authority/executor |
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
| 129-01 launcher/scaffold bootstrap | TASK-193; TASK-201 | qa-auto1 + devops-engineer with disjoint file partitions | qa-auto-ver |
| 129-01 preflight | TASK-204 | arch | arch-ver + qa-auto-ver mandatory; security-auditor input is optional and its absence is nonblocking |
| 129-02 | TASK-197 | be1 | be-ver |
| 129-03 | TASK-198 | fe1 | fe-ver |
| 129-04 | TASK-219 | devops-engineer | code-reviewer + qa-auto-ver |
| 129-05 | TASK-219 primary deletion authority; TASK-201 secondary same-commit consumer/CI cleanup on four support batches | devops-engineer | code-reviewer + qa-auto-ver |
| 129-06 | TASK-234 | be1 | be-ver |
| 129-07 hook batches | TASK-219 primary deletion authority; TASK-234 prerequisite zero-consumer/refactor evidence | devops-engineer | code-reviewer + qa-auto-ver |
| 129-07 final hook batch / runner-CI closure | TASK-219 primary deletion authority; TASK-234 prerequisite evidence; TASK-201 secondary same-commit consumer/CI cleanup | devops-engineer | code-reviewer + qa-auto-ver |

These are real ROADMAP task IDs, not plan-local aliases. Every numbered task repeats its exact requirement, owner, and verifier.

## Versioned outer launcher contract

The only orchestration authority is `testframework/docker/run-task.ps1` with contract version `hub.qa-runner/v1`. Plan 129-01 Task 1 owns their exact static bootstrap under TASK-201; runtime remains **NOT_READY** until `run-task.ps1`, immutable `images.lock.json`, launcher contract/negative-cleanup tests, and independent qa-auto-ver bootstrap acceptance exist. Direct `docker run`, Compose, host test, or ad-hoc cleanup fallback is forbidden.

A canonical batch invocation supplies `-ContractVersion hub.qa-runner/v1`, exact `-PlanId`, caller-supplied unique nonce/attempt-bearing `-TaskRunId`, real `-TaskId`, sealed `-BatchId`, `-ExpectedOperations`, explicit `-OwnerRole` and `-VerifierRoles`, `-Manifest`, `-ImageLock`, `-Mode apply-and-verify`, an independent `-InnerTimeoutSeconds`, and a longer outer `-TimeoutSeconds`. The launcher must:

1. bind the manifest's sealed revision and exact batch-owned source/target/consumer paths; a planned ID is rejected unless ID, count, state, and owned-path hash exactly match the seal;
2. accept only immutable image digests from the image lock and enforce CPU, RAM, swap, PID, network, read-only-root, capability, tmpfs, and writable-mount policy;
3. create only `testframework/reports/<task-run>/` after validating its owner marker containing task ID, authorized owner role, verifier roles, exact revision, and nonce;
4. inside the one transaction, start an isolated Linux writer child authorized only for the named owner and sealed allowlist; invoke the Phase-128 applier with `--apply`; stage and create exactly one allowlisted atomic owner commit; and emit a signed transaction receipt containing owner, nonce, revision, batch/count/path hash, writer exit, commit SHA, and rollback handle `git revert <commit-sha>`;
5. after the writer receipt, start a fresh read-only verifier child whose inputs exclude the writer's conclusion; the named verifier independently runs the Phase-128 root validator plus focused native gate and emits a separate conclusion-blind result;
6. enforce an inner job deadline strictly shorter than the outer wall-clock deadline, record both inner and outer timeout/outcome fields, preserve the first failing child exit code as the transaction exit, and perform unconditional exact-name container/network/task-volume cleanup in `finally`;
7. prove cleanup after success, failure, timeout, and interruption; shared phase caches are read-only during tasks and are removed only by a separate post-independent-acceptance cleanup step in the final transaction.

## Batch contract

For every batch task:

1. The task names one unique **planned** batch ID and one literal expected operation count. No task selects a domain with a runtime loop.
2. Before mutation, assert the Phase-128 receipt, current Git HEAD, inventory identity, ledger SHA, batch state, expected operation count, exact sealed owned paths, clean-worktree non-overlap, and unchanged protected-zone exclusion.
3. `<action>` contains no runtime command. It names the owner authorization and refers to the task's sole canonical nonce-bearing `apply-and-verify` invocation in `<automated>`; no host/manual mutation or second owner/verifier launcher invocation is authorized. Until launcher v1 and exact manifest equality are independently accepted, status remains `NOT_READY` and no mutation is authorized.
4. The single launcher transaction internally separates the owner-authorized writer from the fresh conclusion-blind verifier. Only sealed `operations[].source`, `operations[].target`, `affected_consumers[].path`, and that batch's manifest/evidence record may be staged. The writer invokes `apply_cutover_batch.py --apply`, creates one atomic allowlisted owner commit without AI attribution, and emits the transaction receipt and rollback handle before verifier startup.
5. The independent verifier receives no writer conclusion, mounts mutation outputs read-only, runs `validate_cutover_manifest.py --check` and the focused native gate, and emits a separate result. Its evidence commit, if any, is narrow and is not a second operation commit or a second launcher invocation.
6. Any HIGH finding, incomplete receipt/evidence, manifest/consumer drift, unexpected dirty overlap, operation mismatch, writer/verifier failure, either timeout, cleanup failure, or rollback ambiguity halts before the next batch. Rollback uses the receipt's `git revert <single owner mutation commit>` handle and a new nonce-bearing rollback-verification task; manual recursive deletion and host mutation are forbidden.

## Artifacts this phase produces

- New files: the root Python QA scaffold, `testframework/docker/run-task.ps1`, immutable `testframework/docker/images.lock.json`, and launcher contract/negative-cleanup tests listed in Plan 129-01; relocated unit files at manifest target paths; per-batch evidence/state in the sealed cutover manifest; consumer/CI/docs edits explicitly listed by each batch.
- New symbols: typed fail-closed Python runtime configuration, marker/traceability registration, shared fixtures and redaction helpers; public or injected production seams explicitly replacing private test hooks in Plan 129-06.
- New CLI flags: none are invented by Phase 129. It consumes the Phase-128-sealed `apply_cutover_batch.py` flags `--repo-root`, `--manifest`, `--batch-id`, `--check`, `--self-test`, and `--apply`.
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
| T-129-RESOURCE | Denial of service | Docker execution | medium | mitigate | Explicit CPU, RAM, swap, PID, network, read-only-root, tmpfs, name, and --rm limits on every command. |
| T-129-EVIDENCE | Repudiation | batch receipt | high | mitigate | Require command, exit code, duration, image digest, HEAD, counts, cleanup, owner/verifier identities, and independent verdict before advancement. |
| T-129-ROLLBACK | Repudiation | recovery | high | mitigate | One allowlisted owner mutation commit per batch; rollback only by ordered git revert and root revalidation. |
| T-129-SUPPLY | Tampering | package/image inputs | high | mitigate | Use pinned images and existing lock/checksum files only; reject dependency or lock drift and record image digests. |
</threat_model>

<verification>
The checker must confirm frontmatter/structure, 176 explicit tasks, 173 unique planned batch IDs, each batch at most 64 operations, 1,120 total operations, TASK/owner/verifier mappings, exact equality with the Phase-128 seal, launcher v1 availability and independent acceptance, source audit, and absence of port/parity/dual-run semantics.
</verification>

<success_criteria>
The seven numbered files are schema-complete but **NOT_READY** while the Phase-128 seal, launcher v1, immutable image lock, and independent launcher/guard acceptance are absent. Planning becomes execution-ready only after all 173 planned IDs/counts/owned-path hashes exactly match the seal.
</success_criteria>

<output>
Each numbered plan creates its corresponding `.planning/phases/129-tf-e2e-shared/129-NN-SUMMARY.md` when executed.
</output>
