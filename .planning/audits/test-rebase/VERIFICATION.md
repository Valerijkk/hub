---
task: PHASE-128-TASK-204
verdict: NOT_RUN
verified: false
verifier: pending-independent-docker-rerun
revision: null
verified_at: null
---

# Phase 128 test-rebase guard verification

## Current verdict

**NOT_RUN.** The current guard implementation has not received a fresh run of
the canonical Docker wrapper and has not been signed off by independent
`arch-ver` and `qa-auto-ver`. No PASS, verified-state transition, cutover
manifest, batch application, rollback execution, or Phase 128 completion is
claimed by this document.

The older TASK-202/TASK-203 observations are historical inputs only. They do
not authorize migration and are preserved separately in the sealed ledger and
baseline artifacts.

## Required acceptance evidence

The only supported runtime entrypoint is:

```powershell
pwsh -NoProfile -File .planning/audits/test-rebase/validate-contract.ps1
```

A future PASS requires all of the following from one exact task run:

- immutable inspected image ID and image-lock digest;
- exact live Git source revision;
- CPU, RAM, swap, PID, network-none and read-only limits;
- inner timeout plus an independent bounded host-side watchdog whose sentinel
  exercises the outer kill path;
- held-handle-relative, atomically claimed, identity-pinned, owner-marked
  `testframework/reports/<task-run>/` evidence boundary with exclusive
  no-reparse/single-link artifacts and bounded pipe-captured process output;
- raw stdout/stderr hashes, duration and exact first exit code;
- exact run-script and host execution-argument contract digests;
- exact committed Git blob plus SHA-256 for every task-owned, read-only pinned
  executable snapshot; held-handle proof through consumption and a task-only
  ABA rename/replace sentinel;
- forced-timeout sentinel and daemon-backed proof by exact names and task-run
  label enumeration that every task-owned container is absent after cleanup;
  cleanup must capture/recheck immutable container IDs and ownership labels,
  remove only by ID, then scan both exact names and labels;
- deterministic builder dry-run equality and continued absence of the final
  repository `cutover-manifest.json`;
- independent validator plus atomic apply/rollback/builder race,
  crash-window, final-name-safe pending-journal publication,
  staged-directory/mkdir/cleanup receipt recovery,
  retry-convergence, source-collision and cleanup self-tests;
- temp-only inventory publication idempotence, appearing-target and
  post-fsync/pre-publish crash sentinels without live ledger generation, plus
  pinned-Git-tree/worktree cross-file mutation and HEAD-drift sentinels;
- independent verifier identities recorded only after inspecting the artifacts.

## Verified-state rule

Manifest JSON cannot self-declare a verified state. Every verified batch must
reference an external canonical receipt whose SHA-256 is checked from the
repository. That receipt binds the exact evidence object, operations digest,
applied/verifier commits, verifier identity/timestamp, and transition-chain
digest. Gate evidence also binds the inspected immutable image, exact command
contract, owner marker, raw output artifacts, reconciled counts, duration and
parsed cleanup proof. Aggregate evidence and cleanup artifacts are parsed and
matched to task-run, image, source, exit code and declared digests. Transition
digests must be continuous, and descriptor-safe Git object reads must prove
live source/applied/verifier commit existence and ancestry. Packed-only commit
objects currently fail closed instead of using an unsafe fallback.
The receipt graph is one-way: the batch-local verified projection excludes all
sibling batch state plus its own downstream receipt/self-consumer references,
the receipt binds that projection, and a two-batch constructibility sentinel
must prove later sibling transitions cannot stale an earlier receipt. Each
batch first transition must authenticate the exact immediate global predecessor
from the sole parent commit of its applied commit. The first bootstrap artifact
comes from the declared source commit and names a strict ancestor; later
artifacts must contain the completed verified prefix, current planned batch and
untouched planned suffix, with the prior verifier commit in their ancestry.
Global states must be monotone, and the two-batch sentinel must reject reuse of
the stale original all-planned snapshot. Adjacent within-batch transitions must
preserve exact after/before continuity.
Receipt, raw streams, image/owner locks and cleanup/evidence artifacts must
byte-match the authenticated verifier-commit tree as well as their live
descriptor reads and declared SHA-256 values.

Until those checks run successfully, this frontmatter must remain
`verdict: NOT_RUN`, `verified: false`, and `verified_at: null`.

## G128-S static guard acceptance - NOT_READY (2026-08-09)

**Classification:** static source-and-worktree review only. This is neither
`STATIC_REVIEW_PASS` nor runtime `PASS`; no wrapper, Python tool, validator,
Docker command, test, or batch apply was invoked.

The required receipt cannot truthfully bind a committed full
HEAD/tool/image/authority envelope in the inspected worktree. The wrapper has
a declared immutable image digest and strong committed-blob checks for its
six execution sources, but those facts do not make the current dirty and
untracked input set a committed authority envelope.

### Literal failed assertions

| Assertion | Result | Static evidence |
|---|---|---|
| `G128-S/committed-execution-sources` | **FAIL** | Targeted `git status --porcelain=v1 -uall` reports ` M .planning/audits/test-rebase/inventory.py`, `?? .planning/audits/test-rebase/apply_cutover_batch.py`, and `?? .planning/audits/test-rebase/validate-contract.ps1`. `Add-PinnedExecutionSnapshot` compares each source worktree blob with `HEAD:<path>` and throws `Pinned execution source is not the exact committed Git blob` on mismatch, so a receipt cannot be produced from these current bytes ([validate-contract.ps1](validate-contract.ps1):365-410, 618-623). |
| `G128-S/committed-authority-inputs` | **FAIL** | The same targeted status reports ` M .planning/audits/test-rebase/inventory.json` and `?? .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json`. The wrapper snapshots only the six execution sources above; it passes the founder manifest as a live `/workspace` input to the validator and does not compare either authority JSON input with `$SourceRevision:<path>` ([validate-contract.ps1](validate-contract.ps1):365-410, 673-676). |
| `G128-S/full-HEAD-receipt-binding` | **FAIL** | `$SourceRevision` is only `git rev-parse --verify HEAD`; the inner guard repeats that commit-id comparison. Neither assertion establishes a clean worktree or rejects untracked authority inputs, and the execution-source blob guard does not cover the live JSON authority inputs ([validate-contract.ps1](validate-contract.ps1):614-623, 670-676). |
| `G128-S/runtime-image-and-receipt` | **NOT_RUN** | The source declares an expected immutable image digest and would compare Docker inspection output at runtime, but no such inspection, evidence artifact, or receipt exists in this review ([validate-contract.ps1](validate-contract.ps1):29, 645-656, 745-750). |

### Remediation boundary

No second launcher, tool edit, authority-JSON edit, batch apply, or runtime
guard is authorised by G128-S. Before the single named Docker guard may be
considered, a separately authorised change and review must make every
authority input consumed by the guard provably equal to a blob in the exact
`$SourceRevision` tree (and fail closed on modified, missing, or untracked
paths), then establish a fully committed reviewed envelope. Only a subsequent
one-shot execution through `validate-contract.ps1`, under the sealed plan
gate, may create runtime evidence; direct host/Docker fallback remains
forbidden. Until then, retain `NOT_READY`/`NOT_RUN` and do not represent this
static review as a runtime acceptance.
This receipt does not authorize execution: only G128-D may make one attempt
after an accepted G128-S; that one-shot window closes after the attempt and the
historical wrapper is forbidden after G129-L.
