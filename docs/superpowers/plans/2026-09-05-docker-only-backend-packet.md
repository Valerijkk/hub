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
- Require the exact flat inventory tree plus `container-receipt.json`; match every artifact's SHA-256 and size, the state-specific terminal marker, and the cumulative 64 MiB packet bound. Recheck receipt bytes and directory names after reads. Input must be quiescent after the owning container exits; this is not adversarial host attestation.
- The only new success label is `DOCKER_ARTIFACT_PACKET_PASS`. It proves packet integrity, not passing tests: a valid failed or aborted run can have an intact packet. It does not produce `SEAL`, `SEALED_*`, a Windows resource receipt, an image-identity claim or product readiness.
- Actual backend acceptance still needs the matching Docker run's exit status, complete build/vet/unit/race results, enforced resource invariants and unchanged intended subject/oracle. Missing evidence remains `NOT_RUN` / `NOT_ACCEPTED`.

## Task 1: Portable packet validation

**Files:** Create `testframework/observers/backend_unit/container_packet.py` and `testframework/selftests/architecture/observers/test_backend_unit_container_packet.py`.

**Consumes:** `validate_container_receipt`, existing container receipt schema, `read_regular_relative`, `list_directory_relative`.

**Produces:** The frozen `validate_container_packet` API above.

- [ ] Write genuine temporary-directory selftests: intact COMPLETE/ABORTED packets; changed, missing and extra artifacts; mismatched marker/hostname; traversal/symlink; duplicate JSON and oversized reads.
- [ ] Observe expected missing-module RED in the exact `python-contract` Docker service before implementation. Do not treat unrelated collection errors as the expected RED.
- [ ] Implement bounded validation with existing primitives; valid packet returns `None`, any invalid input raises `ReceiptContractError`.
- [ ] Run Docker GREEN and independent review. Source-only acceptance does not close this task.

## Task 2: Fixed container entrypoint

**Files:** Modify `testframework/docker/entrypoints/python-qa.sh`; add the module CLI in `container_packet.py` and its selftests.

**Source-bound launcher amendment:** The available Python image predates the observer code. Modify only `python-contract` in `docker-compose.local.yml`: add read-only binds for `testframework/observers`, `schemas`, `policies` and `docker` at their matching `/opt/hub/testframework/` paths; set `memswap_limit: 2g`; use fixed entrypoint `[/bin/sh, /opt/hub/testframework/docker/entrypoints/python-qa.sh]`. Keep existing limits, mounts, command and isolation. Add only `testframework/docker/entrypoints/python-qa.sh text eol=lf` to `.gitattributes` and preserve LF script bytes. Test this contract in `testframework/selftests/architecture/test_backend_unit_packet_launcher.py`. This avoids an uncapped image rebuild and does not claim image-baked source identity.

**Interface:** `backend-unit-receipt` mode reads required `HUB_BACKEND_UNIT_RUN_DIR`; no shell command or arbitrary path is evaluated. Reports root is fixed to `/reports`.

```sh
docker compose -f docker-compose.local.yml run --rm -e HUB_BACKEND_UNIT_RUN_DIR=backend-unit-0123456789ab.partial python-contract backend-unit-receipt
```

The example ID must be replaced with the actual task-owned container hostname. Missing input, invalid packets or unsupported arguments fail nonzero. Print only `DOCKER_ARTIFACT_PACKET_PASS` after successful validation, with the explicit limitation that this is not a test verdict.

- [ ] Add missing-input, rejected-packet and success-path CLI tests inside Docker.
- [ ] Add the fixed shell case; keep ordinary `contract` mode unchanged.
- [ ] Verify new behavior and existing framework regression in Docker, then independently review the diff.

## Task 3: Real execution and handoff

- [ ] Check Docker running workloads; reserve one bounded QA slot. Build only changed task-owned images when needed; avoid repeated heavyweight builds.
- [ ] Repair the independently confirmed COMPLETE-finalizer arity defect (`_receipt_build_container_json` requires two arguments but its sole caller passes one), with an actual Docker shell regression before any COMPLETE packet claim.
- [ ] Run canonical `backend-unit`, record exact container/run identity and actual exit, then validate that exact report in `python-contract`.
- [ ] Review actual stage results before any executable acceptance. A packet PASS alone cannot close CALL-DUR08C2 or authorize claiming Messenger/Jitsi complete.
- [ ] Sync `.planning/STATE.md` and Phase 150 continuation with factual results and outstanding work.
- [ ] Exact-path commit and push only `calls-feature-codex` to `origin https://github.com/Valerijkk/hub.git`; verify local/remote revision and clean status.
