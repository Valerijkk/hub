# Phase 129 recovery — Option A (2026-08-14)

## Decision and authority boundary

Founder selected **Option A** on 2026-08-14: preserve bulk commits
`794ae3cc` and `7b5d5501` on branch `claude/test-rebase-20260813`, then
reconcile forward only. This packet records observations; it is not a
replacement for, amendment to, or transition of the v1
`cutover-manifest.json`.

The v1 manifest remains byte-identical, SHA-256
`85dd85ca38923a0211eff6110bef797274289c2ef453ec66364118100dff7bd5`, with
its 173 batches all `planned`. It contains no batch receipts or independent
verifier evidence. This packet therefore cannot claim an old batch as
`applied`, `verified`, PASS, deleted, or complete.

## Status

| Field | Status |
|---|---|
| `STATIC_PATH_MAPPING` | `PASS` |
| `CUTOVER_ACCEPTANCE` | `FAIL/BLOCKED` |
| `PHASE_129` | `RECOVERY_BLOCKED` |
| `CUTOVER_MANIFEST` | `PLANNED_UNCHANGED` |
| `BATCH_RECEIPTS` | `ABSENT` |
| `INDEPENDENT_BATCH_VERDICTS` | `ABSENT` |
| `129-02_BULK_MOVE` | `OBSERVED_RECOVERY_UNACCEPTED` |
| `129-03_BULK_MOVE` | `OBSERVED_RECOVERY_UNACCEPTED` |
| `129-04_DELETE_SET` | `STAGED_UNCOMMITTED_UNACCEPTED` |
| `CI_ROOT_DISCOVERY` | `UNBOUND` |
| `LAUNCHER` | `NOT_READY` |
| `POST_CUTOVER_RUNTIME` | `NOT_RUN` |
| `RECOVERY_VERIFICATION` | `RECOVERY_VERIFIED_NOT_BATCH_ACCEPTED` |

## Observed Git/path provenance

Recovery observation was made at HEAD
`7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`.

- Commit `794ae3cc` contains exactly 350 backend moves mapped to the 32
  Phase 129-02 batches. All 350/350 are `R100` blob-identical moves.
- Commit `7b5d5501` contains exactly 157 frontend moves mapped to the 13
  Phase 129-03 batches. Ninety-nine are blob-identical. The remaining 58
  contain only authorized render-helper import rewrites, 116 changed lines in
  total; no unexplained source change was observed.
- The Git index contains exactly 241 staged deletions matching the 34 Phase
  129-04 batches: 237 legacy Go component/API/integration tests and four
  legacy supports (`backend/testframework/env/builder.go`,
  `backend/testframework/factory/user.go`,
  `backend/testframework/mongotest/setup.go`, and
  `backend/testframework/steps/http.go`). They must remain uncommitted and
  unaccepted during recovery. **Staged is not deleted.**

These facts establish path mapping only. They do not supply the missing
ordered manifest transitions, individual receipts, independent batch verdicts,
CI discovery binding, package/runner implementation, or Docker-only runtime
evidence.

## Required successor gates

Before any recovered change can become accepted authority, work must proceed
in this order:

1. Founder explicitly accepts a successor authority anchored to the recovered
   committed baseline.
2. The 241 staged deletions receive explicit ordered handling; they must never
   become one bulk acceptance.
3. A successor manifest/protocol defines truthful new transitions, receipts,
   verifier evidence and rollback boundaries.
4. CI root discovery is bound to the greenfield roots.
5. The planned `testframework` package and versioned runner are implemented.
6. Only then may Docker-only collection/runtime evidence be produced through
   the accepted launcher.

Until these gates are satisfied, Phase 129 is not complete and the runtime
status remains `NOT_RUN`.

## Independent recovery re-review — ACCEPT (2026-08-14)

The independent re-review accepts this recovery observation packet only.
W-01 was corrected before a fresh review; the fresh review found no
`HIGH` or `WARNING` findings. It reconfirmed the static facts recorded above:
the 350 backend and 157 frontend path mappings, the limited 58 frontend
render-helper rewrites, the 241 staged legacy deletions, and the unchanged
all-`planned` v1 manifest.

This ACCEPT is not a batch receipt or verifier verdict. It accepts neither a
v1 batch, either bulk move, the staged deletion set, CI root discovery,
launcher/package readiness, Docker-only runtime evidence, nor Phase 129
completion. `CUTOVER_ACCEPTANCE=FAIL/BLOCKED` and
`PHASE_129=RECOVERY_BLOCKED` remain in force.
