# Docker-only Backend Packet Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development with frozen ownership and independent review. Root owns Docker execution and publication.

**Goal:** Resume backend verification without executing project helpers on Windows.

**Architecture:** Keep the existing capped `backend-unit` runner and its read-only source mounts. Validate its completed artifact packet in the existing `python-contract` container using portable schema, semantic and safe-path primitives. Do not port the Windows observer or introduce Docker socket access.

**Tech Stack:** Existing Compose services, POSIX shell, Python, pytest, jsonschema.

**Spec:** The frozen design below implements the founder's 2026-09-05 instruction: all checks run only in Docker; find a compliant solution, not a host exception.

## Frozen design and constraints

- This supersedes the Windows `root_handoff` / `observe.py` prerequisite in the final execution sequence of `2026-09-01-backend-unit-hard-2gib-runner-executable-final.md`. Historical source reviews and RED evidence remain historical evidence; no old host seal is fabricated.
- Host activity is limited to ordinary file read/edit, Git and Docker CLI. No project Python/Go/test/validator process on Windows.
- Backend stays at 2 GiB memory, no swap, 6 CPUs, serial package execution, tmpfs caches and a 64 MiB report budget. Run QA services sequentially, without overlapping another worktree's heavy QA.
- No Docker socket, privileged mode, host PID mounts, new dependencies, deletion of existing reports, or changes to Mail/Calendar or `pro-qa-ai/**`.
- Keep source read-only binds to avoid rebuilding the Go image for every source atom. Do not describe these mutable host-backed sources as image-immutable.
- `validate_container_packet(reports_root: Path, run_dir: str) -> None` raises `ReceiptContractError` for an unusable packet. Select one explicit `backend-unit-[0-9a-f]{12,64}.partial` directory, never latest/glob discovery. Its hostname must match the receipt.
- Use descriptor-based bounded reads and directory enumeration from `safe_paths.py`; reject traversal, symlinks, hardlinks, non-regular files, extra/missing entries, duplicate JSON keys, non-finite JSON and incorrect schema/semantics.
- Require the exact flat inventory tree plus `container-receipt.json`; COMPLETE has exactly 22 non-receipt artifacts (owner, marker and five artifacts per stage). Match every artifact's SHA-256 and size, the state-specific terminal marker, and the cumulative 64 MiB unique-packet bound. Count the receipt once; its final reread is separately bounded by its original size. Recheck receipt bytes and directory names after reads. Input must be quiescent after the owning container exits; this is not adversarial host attestation.
- The only new success label is `DOCKER_ARTIFACT_PACKET_PASS`. It proves packet integrity, not passing tests: a valid failed or aborted run can have an intact packet. It does not produce `SEAL`, `SEALED_*`, a Windows resource receipt, an image-identity claim or product readiness.
- Actual backend acceptance still needs the matching Docker run's exit status, complete build/vet/unit/race results, enforced resource invariants and unchanged intended subject/oracle. Missing evidence remains `NOT_RUN` / `NOT_ACCEPTED`.

## Task 1: Portable packet validation

**Files:** Create `testframework/observers/backend_unit/container_packet.py` and `testframework/selftests/architecture/observers/test_backend_unit_container_packet.py`.

**Consumes:** `validate_container_receipt`, existing container receipt schema, `read_regular_relative`, `list_directory_relative`.

**Produces:** The frozen `validate_container_packet` API above.

- [x] Write genuine temporary-directory selftests: intact COMPLETE/ABORTED packets; changed, missing and extra artifacts; mismatched marker/hostname; traversal/symlink; duplicate JSON and oversized reads.
- [x] Observe expected missing-module RED in the exact `python-contract` Docker service before implementation. Unrelated collection and CRLF launch errors were not counted as the expected RED.
- [x] Implement bounded validation with existing primitives; valid packet returns `None`, any invalid input raises `ReceiptContractError`.
- [x] Run Docker GREEN (14 packet cases) and independent review (C/I/M 0/0/0).

## Task 2: Fixed container entrypoint

**Files:** Modify `testframework/docker/entrypoints/python-qa.sh`; add the module CLI in `container_packet.py` and its selftests.

**Source-bound launcher amendment:** The available Python image predates the observer code. Modify only `python-contract` in `docker-compose.local.yml`: add read-only binds for `testframework/observers`, `schemas`, `policies` and `docker` at their matching `/opt/hub/testframework/` paths; set `memswap_limit: 2g`; use fixed entrypoint `[/bin/sh, /opt/hub/testframework/docker/entrypoints/python-qa.sh]`. Keep existing limits, mounts, command and isolation. Preserve LF with exact `.gitattributes` rules for `python-qa.sh`, `frontend-unit.sh`, `backend_unit/receipt.sh`, `vitest.frontend-unit.config.ts`, `Dockerfile.qa` and `policies/root-compose-testframework.v1.json`. No wildcard or reader-side byte normalization. Test the launcher in `testframework/selftests/architecture/test_backend_unit_packet_launcher.py`; synchronize the existing root-Compose topology/dispatcher oracle without weakening negative mutations. This avoids an uncapped image rebuild and does not claim image-baked source identity.

**Interface:** `backend-unit-receipt` mode reads required `HUB_BACKEND_UNIT_RUN_DIR`; no shell command or arbitrary path is evaluated. Reports root is fixed to `/reports`.

```sh
docker compose -f docker-compose.local.yml run --rm -e HUB_BACKEND_UNIT_RUN_DIR=backend-unit-0123456789ab.partial python-contract backend-unit-receipt
```

The example ID must be replaced with the actual task-owned container hostname. Missing input, invalid packets or unsupported arguments fail nonzero. Print only `DOCKER_ARTIFACT_PACKET_PASS` after successful validation, with the explicit limitation that this is not a test verdict.

- [x] Add missing-input, rejected-packet and success-path CLI tests inside Docker.
- [x] Add the fixed shell case; keep ordinary `contract` mode unchanged.
- [x] Verify the new behavior in Docker and independently review the source diff.
- [ ] Close full framework regression: the latest run is 480 passed / 71 failed, not GREEN.

## Task 3: Real execution and handoff

**Task 3A source-binding contract (September 5 product continuation):** mount
current `testframework/docker/backend_unit` read-only at
`/usr/local/lib/hub/backend-unit` and current
`testframework/docker/entrypoints/backend-unit.sh` read-only at
`/usr/local/bin/backend-unit`. Use literal entrypoint
`[/bin/sh, /usr/local/bin/backend-unit]`; the target stays where the receipt
hashes it. Retain the seven existing mounts, empty argv and every resource/
security field. Add exact LF attributes for the newly bound entrypoint and
runtime/plan/execute helpers. One focused Python launcher test goes RED before
Compose changes; synchronize the existing topology oracle, then GREEN/review.
No rebuild, Go implementation, generated source or Windows observer changes.

Task 3A executable evidence: the focused test first failed on the two missing
mounts (1 failed, 0.73s), then passed after the frozen patch (1 passed, 0.76s).
`backend-source-runner-regression-20260905.xml` records 23 passed / 529
deselected in 5.44s, exit 0, covering the new launcher, packet/dispatcher,
shell finalizer and five current Compose/security tests. This focused result
does not close the previous full-framework 71 failures or prove a Go PASS.
Independent source/contract/security review: C/I/M 0/0/0. Task 3A accepted;
actual Go execution and exact-packet verification remain the next steps.

Task 3B runtime repair (frozen after actual execution): checkpoint
`48332ed735cfebac23ec70fc8193ef7d226680da` was committed/pushed with a clean
worktree. Canonical `backend-unit` container
`worktree-master-backend-unit-run-5293b10252d4` exited 1 before compilation
with an awk syntax error. In `runtime.sh::_assert_tmpfs_mount`, the loop
variable `index` collides with awk's built-in function. Ownership is only
that loop variable renamed to `mount_field` and new
`testframework/selftests/architecture/test_backend_unit_tmpfs_mount.py`.
The test extracts the exact source awk program and executes real awk inside
`python-contract`: canonical/optional-field mount success and rejection of
missing/duplicate/non-tmpfs/malformed mount records. It must first reproduce
the syntax error, then pass after the variable-only fix. All resource,
ownership, size and mount-option checks are retained. No Go PASS or accepted
container packet exists from the bootstrap failure.

Task 3C digest repair (independent frozen atom): read-only review confirmed
`_receipt_sha256_file` and `_receipt_sha256_text` each reject their own valid
SHA-256 output because their shell case has 32, not 64, hex positions. Scope
is these two functions in `receipt.sh` plus new
`testframework/selftests/architecture/test_backend_unit_receipt_digest.py`.
Use an exact 64-character lowercase-hex check and a portable awk length/hex
predicate; preserve argument/file/symlink/error guards and output format.
No new helper inventory or broader receipt change. Real shell/hashlib Docker
selftests first reproduce valid-digest rejection, then cover valid file/text,
arity, missing/symlink files and malformed digest-producer outputs. Both
source atoms can be authored independently; root runs their RED/GREEN and
the actual backend service sequentially, without concurrent QA or rebuild.

Tasks 3B/3C evidence: `backend-bootstrap-red-20260905.xml` has 10 failed /
20 passed (seven real awk syntax failures and three valid-digest rejections,
with 20 negative digest cases already passing). After their exact patches,
`backend-bootstrap-green-20260905.xml` has 53 passed / 529 deselected in
5.50s, exit 0. Both independent reviews return C/I/M 0/0/0.
`backend-bootstrap-full-framework-20260905.xml` records 511 passed /
71 failed in 20.06s (582 total); the complete framework is NOT_ACCEPTED.
The root delayed these runs while the Mail worktree's Go QA was active,
without stopping/modifying its container. These are source/runtime-selftest
checkpoints; post-fix actual Go acceptance is still pending.

The prior 71 framework failures stay recorded as `NOT_ACCEPTED`, not deleted
or suppressed. Current-launcher security and shared packet regressions must
pass before actual Go execution; obsolete Windows observer attestation is not
an active prerequisite of this Docker-only path. Any shared-code failure still
needs intent review. A passing Go run does not turn the full Python suite green.

- [x] Check Docker running workloads; use one bounded Python QA slot with no heavy build/pull. Mail/Calendar containers were not stopped or modified.
- [x] Repair COMPLETE-finalizer arity and the shell quote defects exposed by real `/bin/sh` regressions; both shell tests pass inside Docker.
- [ ] Run canonical `backend-unit`, record exact container/run identity and actual exit, then validate that exact report in `python-contract`.
- [ ] Review actual stage results before any executable acceptance. A packet PASS alone cannot close CALL-DUR08C2 or authorize claiming Messenger/Jitsi complete.
- [x] Sync `.planning/STATE.md`, Phase 150 continuation and QA READMEs with factual results and outstanding work.
- [ ] Exact-path commit and push only `calls-feature-codex` to `origin https://github.com/Valerijkk/hub.git`; verify local/remote revision and clean status.

## September 5 checkpoint evidence and remaining work

- Planning checkpoint `f58476b0` is already pushed. This implementation is a
  bounded source/runtime-selftest checkpoint, not an accepted full framework
  baseline or a completed backend/product gate.
- `docker-packet-final-green-20260905.xml`: 15 passed (14 packet + one launcher).
  `docker-finalizer-green-20260905.xml`: two actual shell tests passed.
  Reports are local under `testframework/reports/`; do not stage generated
  report directories or replace them with a fabricated PASS.
- The first broad source-bound run was 473 passed / 76 failed (549 cases).
  After two review regressions, topology/dispatcher sync and exact LF repair,
  `docker-only-framework-checkpoint-20260905.xml` is **480 passed / 71 failed**
  (551 cases, 19.94s). All 17 new tests pass in this full run.
- Docker resource inspection of the earlier full run recorded memory and
  memory-plus-swap both `2147483648`, `NanoCpus=2000000000`, read-only rootfs
  and `PidsLimit=256`; observed usage was 105.2 MiB. The later disposable
  checkpoint container exited before a second inspect; do not invent its
  container ID or a second resource sample. No host test/validator ran.
- The real fixed `backend-unit-receipt` mode with missing run selection exited
  1 and printed `DOCKER_ARTIFACT_PACKET_INVALID`. No stale report was selected.
  Validation of an actual newly completed backend run remains **NOT_RUN**.
- Remaining failures include historical observer adapter API mismatches,
  publication/safe-path test doubles, observer identity fixtures and stale
  seal expectations, and lexical/AST scanner contracts. Do not classify every
  failure as pre-existing without before/after evidence or disable these tests.
  The new packet path does not execute the Windows observer.
- The identity-fixture repair should be a separate frozen atom: update the
  valid observer and six named-invalid fixtures' hostname/run-ID baseline to
  the full engine ID prefix, preserving each deliberate invalidity; then sync
  the valid final receipt, seal digest and its direct test consumer. Do not
  alter the schema-invalid nonce fixture or weaken the identity validator.
- Next executable work: decompose remaining contract failures, close required
  regression, bind the actual backend runner helpers to their reviewed bytes,
  run build/vet/unit/race once in its capped container, and validate that exact
  packet. Shell fixes in source are not evidence that an old backend image
  already contains them. CALL-DUR08C2, Jitsi deployment/joining and Messenger
  runtime acceptance remain unchanged.
