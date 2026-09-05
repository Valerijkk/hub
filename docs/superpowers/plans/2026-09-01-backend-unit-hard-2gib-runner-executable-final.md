# Backend Unit Hard-2-GiB Runner Executable Final Implementation Plan

> **2026-09-05 execution supersession:** The founder requires all validators in Docker and rejects a Windows-host exception. Preserve this plan as historical source/review evidence. Its host `root_handoff` / observer execution sequence is replaced by `2026-09-05-docker-only-backend-packet.md`; do not retry it or fabricate its `SEAL`.

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` task-by-task. Checkboxes are the execution ledger.

**Goal:** Deliver the first observer-sealed, pre-start-capped 2-GiB/no-swap `backend-unit` candidate while keeping the one public backend QA command unchanged.

**Architecture:** The zero-argument image-baked POSIX runner is the sole writer of an owner-bound `.partial` packet. A root-managed stdlib-only observer starts only after build and GREEN, acquires and retains a cooperative source lease, uses read-only Docker discovery until its one guarded stop, then after every container writer has exited validates the payload and alone writes final receipt, observer receipt, and `SEAL`, followed by the only final rename. A cooperative lease detects ordinary source drift but does not claim hostile-host immutability.

**Tech Stack:** Docker Compose and read-only Docker Engine inspection; POSIX `sh`; Go 1.25; Python 3.12.x standard library; Docker-contained `pytest` and `jsonschema`; Draft 2020-12 JSON Schema.

**Spec:** `docs/superpowers/specs/2026-09-01-backend-unit-hard-2gib-runner-design.md`, `docs/superpowers/specs/2026-09-01-backend-unit-hard-2gib-runner-review-amendment.md`, `docs/superpowers/specs/2026-09-01-backend-unit-hard-2gib-runner-review-resolution.md`, and `docs/superpowers/specs/2026-09-01-backend-unit-hard-2gib-runner-final-errata.md`; later documents override earlier conflicts.

## Global Constraints

- Record `worktree=worktree-master`, initial `HEAD`, `approval_commit=1347d9f7`, accepted `C0/I0/M0`, `user_approval=recorded`, and a clean status before dispatch. Exact-path staging only; never use broad staging.
- Do not read or modify `pro-qa-ai/**`, product sources, frozen CALL-DUR-08C2 oracle sources, generated host Ent output, Mail/Calendar, `testframework/runner/**`, v1/v2/v3/final predecessor plans, or unrelated files.
- Docker-only applies to tests, builds, linters, validators, and QA. No Docker command occurs in Test Wave or Implementation Wave. Until observed evidence, retain `RED_PENDING`, `NOT_RUN`, `STATIC_ACCEPTED`, or `EXECUTABLE_NOT_RUN` as applicable.
- The sole public backend command is exactly `docker compose -f docker-compose.local.yml run --rm backend-unit`; it accepts zero arguments. No selector, `--build`, `--pull`, socket, privileged mode, nested container, new Compose service, host launcher, dependency install, prune, broad cleanup, or AI attribution trailer is allowed.
- Compose fixes `cpus: 6.0`, `mem_limit: 2g`, `memswap_limit: 2g`, `GOFLAGS=-p=1`, `GOENV=off`, `GOWORK=off`, `GOTOOLCHAIN=local`, and `GOPROXY=off`. The runner unsets inherited `GOMAXPROCS`, uses explicit `-p=1`, and never adds `go test -parallel`.
- The only writable bind is exactly `./testframework/reports:/reports` (writable). The six ordered source binds are exactly `backend/cmd`, `backend/internal`, `backend/tests`, `backend/ent/schema`, `backend/migrations`, and `backend/testframework`, all exact and read-only. Remove only `backend/docs`; prohibit broad `backend` root, documentation, observer, and every other writable bind. Generated Ent output remains image authority.
- Tmpfs ceilings are `/tmp=64 MiB`, `GOCACHE=768 MiB`, `GOTMPDIR=384 MiB`. The only cleanup is `go clean -cache`, after an owned child exits, at a `512 MiB` watermark and only after owned-path/UID/non-symlink proof. Cache, temporary data, and build outputs never enter reports.
- The hard report budget is `64 MiB`, reserving `8 MiB` metadata. Stream stdout/stderr caps are build `2/2`, vet `2/2`, unit `16/4`, race `24/4` MiB. A cap stops and reaps only the owned Go child group, marks truncation, and is infrastructure evidence, never source failure.
- The three and only three **HARD2G** top-level schemas are `testframework/schemas/backend-unit-package-plan.schema.v1.json`, `testframework/schemas/backend-unit-container-receipt.schema.v1.json`, and `testframework/schemas/backend-unit-observer-receipt.schema.v1.json`; this does not constrain unrelated repository schemas. The observer schema owns `$defs.finalReceipt`. Fixtures exist only under `testframework/selftests/architecture/observers/fixtures/backend_unit/**`.
- The two frozen subject mappings are `backend/internal/calls/call_durable_credential_allocation.go -> sha256:be6ff058468c4e288e0b0d7060fd17cf3118229d009f28dc056d89f281e99d4d` and `backend/tests/calls/call_durable_credential_allocation_test.go -> sha256:0ae525c86a75369bdb5790f26b1308f42574fb545883328da24011d4150b03a9`.
- Existing `testframework/README.md`, `testframework/schemas/README.md`, and `testframework/docker/README.md` package metadata/retained-asset copies are allowed human metadata copied by existing schema-directory/Docker-directory COPY rules. They are never read or tested as an image, static, or semantic oracle; no selective copy is needed. Only HARD2G human prose must not become an image, static, or semantic oracle. Executable source, policies, the three HARD2G schemas, fixtures, and selftests are the oracle.
- The exact read-only Docker allowlist additionally includes `docker image inspect --format '{{json .}}' worktree-master-python-contract` and `docker image inspect --format '{{json .}}' worktree-master-backend-unit`; no dynamic image name is allowed. There is no new attestation/helper mode. The canonical build-input path manifest is exactly Compose (`docker-compose.local.yml`), Dockerfile (`testframework/docker/Dockerfile.qa`), entrypoint (`testframework/docker/entrypoints/backend-unit.sh`), four helpers (`testframework/docker/backend_unit/{runtime.sh,plan.sh,execute.sh,receipt.sh}`), observer package (`testframework/observers/**`), three HARD2G schemas, two new policies (`backend-unit-observer.v1.json`, `backend-unit-phase150-08c2-subject.v1.json`), root policy (`root-compose-testframework.v1.json`), and root selftests/fixtures (`testframework/selftests/architecture/{docker,observers}/**`). Root hashes the same manifest before build and after GREEN; the observer recomputes it after lease and binds manifest bytes/hash plus both fixed image inspect IDs into observer/final receipt; bound backend container `.Image` must equal inspected backend image ID. No writes between snapshots, no rebuild/pull, mismatch no-seal `75`.
- Every atom dispatch explicitly sets implementer `model: gpt-5.6-terra`, `reasoning_effort: xhigh`; its distinct reviewer and re-reviewer explicitly set `model: gpt-5.6-terra`, `reasoning_effort: high`. `inherit`, omitted routing, fallback, and Sol subagents are prohibited.

## Closed Classification and Observer Contract

The policy's sealed enum is exactly these 15 values, and no other sealed result may be emitted:

```text
SEALED_EXECUTABLE_PASS
SEALED_EXECUTABLE_FAIL_DISCOVERY
SEALED_EXECUTABLE_FAIL_STAGE
SEALED_TIMEOUT_UNTRIAGED
SEALED_INFRA_ABORTED_OOM
SEALED_INFRA_ABORTED_RESOURCE_CONTRACT
SEALED_INFRA_ABORTED_REPORT_BUDGET
SEALED_INFRA_ABORTED_TMPFS
SEALED_INFRA_ABORTED_HOST_RAM_FLOOR
SEALED_INFRA_ABORTED_DISK_PRESSURE
SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT
SEALED_INTERRUPTED
SEALED_INVALID_QA_OVERLAP
SEALED_INVALID_SOURCE_MUTATED
SEALED_INVALID_PROVENANCE
```

`ClassificationFacts` has one and only one priority rule:

```text
source mutation > invalid provenance > earliest explicit stop
(tie: interrupt > QA > disk > RAM > deadline; QA => SEALED_INVALID_QA_OVERLAP)
> OOM > resource > budget > tmpfs > timeout > discovery > stage > pass
```

The sealed classification enum deliberately remains `SEALED_TIMEOUT_UNTRIAGED`; the immutable outcome field is `TIMEOUT_UNTRIAGED`. Later triage is a separate non-mutating record and never rewrites either. There is no second or duplicate QA bucket. Every sealed classification exits `0`; malformed usage exits `64`; no authoritative candidate/seal exits `75`; an interrupt before bind with no seal exits `130`.

The observer is invoked exactly as `python testframework/observers/backend_unit/observe.py --compose-project worktree-master`. It accepts only a project matching `^[a-z0-9][a-z0-9_-]{0,62}$`, self-bootstraps the exact `testframework` root, requires Python `3.12.x` and stdlib-only imports, acquires the exclusive regular non-symlink lease after disk preflight and before `READY`, retains it through terminal verification and post snapshots, and never launches/builds Compose. Lease uncertainty, dirty status, `HEAD`, manifest, or frozen-digest change yields nonauthoritative source-mutated evidence; it releases the lease only afterward. This remains a cooperative no-ACL guarantee, not hostile-host immutability.

Its literal `shell=False` argv allowlist is exactly:

```text
events  [docker, events, --since, <validated RFC3339Nano start>, --filter, type=container, --filter, event=create, --filter, event=start, --filter, event=die, --filter, event=destroy, --format, {{json .}}]
inspect [docker, inspect, --type, container, --format, {{json .}}, <bound64id>]
ps      [docker, ps, --no-trunc, --format, {{json .}}]
stats   [docker, stats, --no-stream, --format, {{json .}}, <bound64id>]
stop    [docker, stop, --timeout, 5, <bound64id>]
image   [docker, image, inspect, --format, {{json .}}, worktree-master-python-contract]
image   [docker, image, inspect, --format, {{json .}}, worktree-master-backend-unit]
```

Only the typed RFC3339Nano `since` and a previously bound full 64-lower-hex Engine ID are dynamic. The adapter streams events line-by-line, checks Compose project/service labels locally, retains at most 16 records and 64 KiB, and terminates/reaps the stream; each one-shot read has output cap 1 MiB and timeout 10 seconds. Overflow, timeout, extra token, foreign result, malformed local labels, or stream reap uncertainty is no-seal exit `75`. The only Docker mutation is exactly one stop after complete candidate binding.

Candidate binding is exactly one candidate only when Compose project/service labels, start window, full 64-lower-hex Engine ID, owner-marker hostname prefix, inspected fixed backend image ID, and container `.Image` all match. The identical `run_id` and `nonce` repeat through observer receipt, final receipt, and `SEAL`. Missing, ambiguous, or mismatched candidate evidence is `INVALID_EVIDENCE`, unsealed/`75`, and is never stoppable or sealable. Candidate deadline is 120 seconds, samples are every 5 seconds, and post-`die` grace is 30 seconds.

Host-safety policy is executable: preflight requires RAM `>=2 GiB`, C: free `>=250 GiB`, and three disk samples two seconds apart; defer when 2/3 samples have `PercentDiskTime>=70` or `CurrentDiskQueueLength>=2`. During runtime sample every 5 seconds and stop on two consecutive RAM `<2 GiB` or the rolling 2/3 disk predicate, prohibited QA overlap, or the six-hour deadline. No other `backend-unit`, no Mail/Calendar QA/test/build, and idle BuildKit are required. Receipts retain raw samples and predicate evaluations.

The source lease uses fixed read-only argv `git rev-parse HEAD` and `git status --porcelain=v1 --untracked-files=all`, plus before/after manifests for all six roots, the canonical build-input path manifest, and both frozen subject digests. Root additionally establishes a cooperative no-write window before FG-01, captures clean canonical build-input manifest bytes/hashes before build, rechecks after build and GREEN before observer, and the observer recomputes after lease and binds the manifest bytes/hash plus both fixed image inspect IDs; this does not move the observer before the approved build/GREEN sequence.

## File Map

| Area | Exact files |
|---|---|
| Existing static contract/root policy | `testframework/selftests/architecture/docker/test_root_compose_contract.py`; `testframework/policies/root-compose-testframework.v1.json` |
| Selftests/fixtures | `testframework/selftests/architecture/observers/test_backend_unit_receipts.py`; `testframework/selftests/architecture/observers/test_backend_unit_observer.py`; `testframework/selftests/architecture/observers/fixtures/backend_unit/**` |
| Policies/schemas | `testframework/policies/backend-unit-observer.v1.json`; `testframework/policies/backend-unit-phase150-08c2-subject.v1.json`; the three HARD2G schema paths above |
| Observer | `testframework/observers/__init__.py`; `testframework/observers/backend_unit/{__init__.py,contracts.py,safe_paths.py,lease.py,resources.py,docker_adapter.py,publication.py,observe.py}` |
| Runner | `testframework/docker/backend_unit/{runtime.sh,plan.sh,execute.sh,receipt.sh}`; `testframework/docker/entrypoints/backend-unit.sh` |
| Integration/docs | `docker-compose.local.yml`; `testframework/docker/Dockerfile.qa`; `testframework/PACKAGE-CONTRACT.md`; `testframework/README.md`; `testframework/schemas/README.md`; `testframework/docker/README.md`; `docs/TESTING.md`; `docs/test_plan.md`; `docs/TROUBLESHOOTING.md`; `docs/stack_and_tools.md` |
| Observed evidence only | `.planning/STATE.md`; `.planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-VERIFICATION.md` |

## Global SDD Review Package Contract

Every implementer, reviewer, and re-reviewer receives a bounded, no-history dispatch package containing: the exact task brief; the review package; the conclusion-blind TaskContract and governing-spec precedence; exact allowed and denied paths; current-code `BASE` and full `HEAD`; raw check facts limited to `PASS`, `FAIL`, or `NOT_RUN`; requested and actual model/effort plus telemetry, or literal `UNKNOWN` when unavailable; and no inherited routing or fallback. The reviewer is read-only. Its only output is a conclusion-blind finding package; a separate writer-only remediation handoff receives accepted findings and may change only exact allowed paths. The changed bounded package is then sent to a distinct read-only re-reviewer. At most two repair/re-review cycles are permitted for an atom or whole branch; unresolved, ambiguous, unavailable, or repeated failure escalates to root for further decomposition and `BLOCKED`, never silent continuation. `UNKNOWN` is not PASS.

## Review Protocol for Every Atom

Each Test Wave and Implementation Wave atom follows the Global SDD Review Package Contract and this mandatory sequence: (1) an implementer dispatched with explicit `model: gpt-5.6-terra`, `reasoning_effort: xhigh` makes exact-path implementation and provisional commit; (2) a distinct reviewer dispatched with explicit `model: gpt-5.6-terra`, `reasoning_effort: high` conducts both specification-conformance and quality/security review of `SDD BASE..HEAD` plus the atomic diff; (3) every accepted finding receives a smallest exact-path repair commit named `testframework: address hard-2gib <atom> review findings` (or `docs: address hard-2gib <atom> review findings`); (4) the changed bounded atom receives a distinct read-only re-reviewer dispatched with explicit `model: gpt-5.6-terra`, `reasoning_effort: high`, with zero unresolved critical, important, or minor findings before the next atom. At most two repair/re-review cycles are permitted; otherwise root decomposes or marks `BLOCKED`. `UNKNOWN` is not PASS. `inherit`, omitted routing, and fallback are forbidden. This is in addition to the later whole-branch review; do not substitute one for the other.

### Task 0: Authority

**Files:** No repository edit.

- [ ] Record the approval facts and initial clean lease status using `git status --porcelain=v1 --untracked-files=all` in the execution ledger.
- [ ] Record the limited exception: only the file map's QA policies, schemas, tests, observer, image runner, integration, documentation, and observed evidence may change.
- [ ] Freeze the Global SDD Review Package Contract for every later dispatch: task brief/review package, conclusion-blind TaskContract/spec precedence, allowed/denied paths, current-code `BASE`/full `HEAD`, raw check facts, requested/actual model/effort telemetry or `UNKNOWN`, bounded no-history routing, read-only reviewer/writer-only remediation handoff, and the two-cycle `BLOCKED` ceiling.

## Test Wave

### Task 1: TW-01 Receipt Fixtures and Semantic Contract

**Files:** Create `testframework/selftests/architecture/observers/test_backend_unit_receipts.py`; create `testframework/selftests/architecture/observers/fixtures/backend_unit/{valid,invalid}/**`.

- [ ] Add valid bytes for all three HARD2G schema families and `$defs.finalReceipt`; add one-rule-invalid fixtures for unknown key, uppercase digest, nonce, discovery argv, duplicate import path, count/set hash, non-bijection, inventory/hash, foreign run, engine-ID, canonical build-input manifest/hash, fixed image ID, bound backend `.Image`, and CRLF/missing/extra/reordered `SEAL`.
- [ ] Assert `parse_seal(b"hub.qa.backend-unit-seal/v1\\r\\n")` raises `ReceiptContractError`; assert all four `validate_*` functions reject the semantic fixtures, including immutable `TIMEOUT_UNTRIAGED` paired only with sealed `SEALED_TIMEOUT_UNTRIAGED`.
- [ ] Commit: `git add testframework/selftests/architecture/observers/test_backend_unit_receipts.py testframework/selftests/architecture/observers/fixtures/backend_unit && git commit -m "testframework: add hard-2gib receipt contract fixtures"`.
- [ ] Apply the Review Protocol; retain `RED_PENDING` and run no Docker command.

### Task 2: TW-02 Observer Security and Lifecycle Selftests

**Files:** Create `testframework/selftests/architecture/observers/test_backend_unit_observer.py`; modify `testframework/selftests/architecture/observers/test_backend_unit_receipts.py` only for indispensable shared fixtures.

- [ ] Test the literal Docker argv allowlist, including both fixed `docker image inspect --format '{{json .}}'` commands and rejection of dynamic image names; typed start, one-shot cap/timeout, local label checks, events line-by-line cap 16/64 KiB/termination/reap/exit 75, exact candidate binding (labels/start-window/full ID/owner hostname), identical `run_id`/`nonce` through observer/final/`SEAL`, full ID before one stop, 120/5/30 lifecycle, exact `READY`, no pre-validation write, and publication order.
- [ ] Test safe path failures (absolute/parent, symlink/reparse, hard link, nonregular file, identity/size race, existing output/cross-directory rename), lease acquisition after disk preflight before `READY`, retention through post snapshots, lease before/after drift, root no-write rechecks, canonical build-input manifest byte/hash snapshots before build, after GREEN, and after lease, bound image identity, and no-ACL hostile-host limitation.
- [ ] Test RAM/disk/overlap/deadline policy, raw-sample/predicate receipt retention, every literal enum, the immutable outcome `TIMEOUT_UNTRIAGED` versus sealed `SEALED_TIMEOUT_UNTRIAGED`, the sole priority rule, exits `0/64/75/130`, and QA mapping only to `SEALED_INVALID_QA_OVERLAP`.
- [ ] Commit: `git add testframework/selftests/architecture/observers/test_backend_unit_observer.py testframework/selftests/architecture/observers/test_backend_unit_receipts.py && git commit -m "testframework: specify hard-2gib observer security contract"`.
- [ ] Apply the Review Protocol; retain `RED_PENDING` and run no Docker command.

### Task 3: TW-03 Existing Compose/Docker Static Contract

**Files:** Modify `testframework/selftests/architecture/docker/test_root_compose_contract.py`; modify `testframework/policies/root-compose-testframework.v1.json`.

- [ ] Assert fixed cgroup/resources/Go/tmpfs/budget/zero argv, literal only writable bind `./testframework/reports:/reports`, and the six ordered read-only source binds. Mutate broad root/docs bind, any writable source or other writable bind, socket/privileged mode, new service, host launcher, permissive Docker argv, wrong resource values, container-written observer/final/SEAL, and wrong publication order to prove rejection.
- [ ] Assert Dockerfile placement: only the `python-tests` stage has `COPY testframework/observers/ ./observers/`, `ENV PYTHONPATH=/opt/hub/testframework/src:/opt/hub/testframework`, and `ENV PYTHONSAFEPATH=1`; `backend-unit` has `COPY testframework/docker/backend_unit/ /usr/local/lib/hub/backend-unit/`, CRLF stripping, helper files mode `0555`, and existing `/usr/local/bin/backend-unit` references those exact helper paths. Assert the observer is not copied to the Go stage and image import `observers.backend_unit` occurs only where permitted.
- [ ] Assert no observer bind is added: the TDD topology deliberately uses image-owned RED, then the one admitted rebuild, then GREEN. Assert no HARD2G prose image/static oracle, while existing `testframework/README.md`, `testframework/schemas/README.md`, and `testframework/docker/README.md` retained-asset/package-metadata copies remain allowed human metadata, copied by existing directory COPY without selective copy, and are never read/tested as an oracle.
- [ ] Extend retained Docker assets with all four runner helpers and retain `testframework/docker/README.md` without reading README prose as an oracle. Assert fixed image-inspect allowance, canonical build-input path manifest, manifest snapshot equality, receipt image-ID bindings, and backend container `.Image` equality to the inspected fixed backend image ID.
- [ ] Commit: `git add testframework/selftests/architecture/docker/test_root_compose_contract.py testframework/policies/root-compose-testframework.v1.json && git commit -m "testframework: specify hard-2gib compose launcher contract"`.
- [ ] Apply the Review Protocol; retain `RED_PENDING` and run no Docker command.

## Mandatory RED Gate

### Task 4: RED-01 One Root-Owned Full RED

**Files:** No source edit.

- [ ] Root runs exactly once: `docker compose -f docker-compose.local.yml run --rm python-contract`.
- [ ] Continue only on one nonzero RED that names either new observer/schema missing-import/path failures, or the expected TW-03 static `RootComposeContractError("root compose contract is invalid")` caused by the pre-implementation Compose/Docker mismatch. Record UTC timestamp, image identity, exit, and failure names.
- [ ] On any other, unrelated, or ambiguous RED, permanently stop: no retry/second RED, no production asset, no reinterpretation. Request newly approved authority only after documented triage.

## Implementation Wave

One root-owned implementation session spans Tasks 5–18. Every item below is one atom, runs no Docker command, commits only listed paths, then applies the Review Protocol before the next atom. Task 18 documentation is complete before whole-branch review or any runtime sequence.

### Task 5: IW-01 Policies and Three Structural Schemas

**Files:** Create `testframework/policies/backend-unit-observer.v1.json`, `testframework/policies/backend-unit-phase150-08c2-subject.v1.json`, and all three HARD2G schema paths.

- [ ] Create closed Draft 2020-12 structures (`additionalProperties:false`); structural shape only, with `$defs.finalReceipt` only in observer schema and no fourth HARD2G schema.
- [ ] Policy contains literal argv allowlist including the two fixed image-inspect commands, candidate/image binding, canonical build-input manifest and snapshot policy, host-safety thresholds, exits, immutable triage relation (`SEALED_TIMEOUT_UNTRIAGED` / `TIMEOUT_UNTRIAGED`), all 15 enums, the one `ClassificationFacts` priority rule, and exact QA mapping. Subject policy contains only the two frozen prefixed mappings.
- [ ] Commit: `git add testframework/policies/backend-unit-observer.v1.json testframework/policies/backend-unit-phase150-08c2-subject.v1.json testframework/schemas/backend-unit-package-plan.schema.v1.json testframework/schemas/backend-unit-container-receipt.schema.v1.json testframework/schemas/backend-unit-observer-receipt.schema.v1.json && git commit -m "testframework: add hard-2gib policies and schemas"`.

### Task 6: IW-02 Pure Receipt Semantics

**Files:** Create `testframework/observers/__init__.py`, `testframework/observers/backend_unit/__init__.py`, `testframework/observers/backend_unit/contracts.py`.

- [ ] Export exactly `ReceiptContractError`, `parse_seal`, `validate_package_plan`, `validate_container_receipt`, `validate_observer_receipt`, and `validate_final_receipt`; canonical JSON only, no I/O/Docker.
- [ ] Enforce LF-finalized four-line `SEAL`, discovery/order/count/set/result/artifact/identity/final bindings, canonical build-input manifest bytes/hash, both fixed image inspect IDs, bound backend `.Image` equality, identical `run_id`/`nonce`, immutable `TIMEOUT_UNTRIAGED` paired with sealed `SEALED_TIMEOUT_UNTRIAGED`, and the closed classification relation.
- [ ] Commit: `git add testframework/observers/__init__.py testframework/observers/backend_unit/__init__.py testframework/observers/backend_unit/contracts.py && git commit -m "testframework: validate hard-2gib receipt semantics"`.

### Task 7: IW-03 Fail-Closed Safe Paths

**Files:** Create `testframework/observers/backend_unit/safe_paths.py`.

- [ ] Implement `open_regular_relative(root: pathlib.Path, relative: str) -> int` and `atomic_write_same_dir(path: pathlib.Path, payload: bytes) -> None`; reject path escape, reparse/symlink, nonregular/multilink, instability, and existing output.
- [ ] Use Windows handle-anchored `CreateFileW` traversal and POSIX directory-fd `openat`/`O_NOFOLLOW`; uncertainty produces no seal/`75`.
- [ ] Commit: `git add testframework/observers/backend_unit/safe_paths.py && git commit -m "testframework: enforce hard-2gib safe paths"`.

### Task 8: IW-04 Cooperative Source Lease

**Files:** Create `testframework/observers/backend_unit/lease.py`.

- [ ] Implement `SourceLease.acquire(path: pathlib.Path) -> SourceLease`, `snapshot() -> dict`, and `verify_unchanged(before: dict, after: dict) -> None` around exclusive regular non-symlink `testframework/reports/.backend-unit-source-mutation.lease`.
- [ ] Acquire after disk preflight and before `READY`; retain through terminal verification and post snapshots, then release. Snapshot six roots, the canonical build-input path manifest bytes/hash, exact two fixed image-inspect argv results, the exact two Git argv commands, and the two frozen subject digests before/after. Dirty status or uncertainty, `HEAD`, manifest, image ID, or digest change is source-mutated/non-authoritative; do not change ACLs.
- [ ] Commit: `git add testframework/observers/backend_unit/lease.py && git commit -m "testframework: add hard-2gib cooperative source lease"`.

### Task 9: IW-05 Resource Sampling and Classification

**Files:** Create `testframework/observers/backend_unit/resources.py`.

- [ ] Implement `collect_host_sample() -> dict`, `preflight(samples: tuple[dict, dict, dict]) -> None`, and `classify(samples: tuple[dict, ...], policy: dict) -> str` using policy constants and the fixed noninteractive PowerShell metric argv.
- [ ] Enforce preflight RAM `>=2 GiB`, C: free `>=250 GiB`, three samples two seconds apart, defer on 2/3 `PercentDiskTime>=70` or `CurrentDiskQueueLength>=2`; runtime five-second samples, two-consecutive RAM floor, rolling disk 2/3, prohibited QA overlap, idle BuildKit, and six-hour deadline. Retain raw samples/predicates and use sole classification priority; sealed `0`, usage `64`, unsealable `75`, unbound interrupt `130`.
- [ ] Commit: `git add testframework/observers/backend_unit/resources.py && git commit -m "testframework: classify hard-2gib resources"`.

### Task 10: IW-06 Docker Read Adapter

**Files:** Create `testframework/observers/backend_unit/docker_adapter.py`.

- [ ] Implement `DockerReadAdapter(project: str)`, `events()`, `inspect(container_id: str)`, `ps()`, `stats(container_id: str)`, `inspect_fixed_images()`, and `stop_bound_once(container_id: str) -> dict`; validate project before construction and permit inspect/stats/stop only for the unique bound 64-lower-hex ID.
- [ ] Use only the literal `shell=False` `events`, `inspect`, `ps`, `stats`, two fixed `image inspect`, and one stop argv; only typed RFC3339Nano since and bound64id are dynamic. Events follow the frozen template; local labels, start window, owner-marker hostname, inspected backend image ID, 16/64KiB streaming cap/reap, and 1MiB/10s one-shot bounds are fail-closed. Any foreign/extra/ambiguous/mismatched result remains unsealed `75` and cannot be stopped.
- [ ] Commit: `git add testframework/observers/backend_unit/docker_adapter.py && git commit -m "testframework: add hard-2gib docker adapter"`.

### Task 11: IW-07 Observer Publication

**Files:** Create `testframework/observers/backend_unit/publication.py`.

- [ ] Implement `publish(run_dir: pathlib.Path, final: dict, observer: dict) -> pathlib.Path`; only after container exit/stability write final, observer, `SEAL`, then same-directory rename. Any absence/activity/instability/identity/manifest/image mismatch stays unsealed `75`.
- [ ] Commit: `git add testframework/observers/backend_unit/publication.py && git commit -m "testframework: publish hard-2gib receipts"`.

### Task 12: IW-08 Observer State Machine

**Files:** Create `testframework/observers/backend_unit/observe.py`.

- [ ] Implement `main(argv: collections.abc.Sequence[str]) -> int` for the literal `--compose-project worktree-master` command; bootstrap, root source checks, disk preflight, acquire lease, inspect both fixed images, recompute/bind the canonical build-input manifest bytes/hash, print one `READY observer_session=<32-lower-hex>\n`, bind exactly one candidate, monitor, terminal verification/post snapshots, validate, publish, then release. Never launch/build Compose.
- [ ] Preserve candidate state only when labels, start window, full Engine ID, and owner-marker hostname prefix, and `.Image == worktree-master-backend-unit` inspected ID all match; repeat one `run_id`/`nonce` through observer/final/`SEAL`. Missing/ambiguous/mismatch is `INVALID_EVIDENCE`, no seal/`75`, never stop/seal.
- [ ] Commit: `git add testframework/observers/backend_unit/observe.py && git commit -m "testframework: add hard-2gib observer state machine"`.

### Task 13: IW-09 Runtime Helper

**Files:** Create `testframework/docker/backend_unit/runtime.sh`.

- [ ] Add `assert_cgroup`, `assert_owned_tmpfs`, `write_atomic`, and `check_budget` for cgroup v1/v2, exact no-swap limit, budget, tmpfs, and child-group-only termination.
- [ ] Commit: `git add testframework/docker/backend_unit/runtime.sh && git commit -m "testframework: add hard-2gib runner runtime helper"`.

### Task 14: IW-10 Package Plan Helper

**Files:** Create `testframework/docker/backend_unit/plan.sh`.

- [ ] Discover packages with exactly these three commands, where each template contains one literal backslash byte followed by `t` and output delimiter byte `0x09`:

```text
go list -f '{{printf "%s\t%s" .ImportPath .Name}}' ./...
go list -tags hubtest -f '{{printf "%s\t%s" .ImportPath .Name}}' ./...
go list -tags hubtest -f '{{printf "%s\t%s" .ImportPath .Name}}' ./tests/...
```

- [ ] Map build to the first command, vet to the second, and both unit and race to the third; sort under `LC_ALL=C` lexical order. Add a byte/exact-argv selftest requiring output delimiter `0x09` and rejecting bytes `0x5c 0x74`; reject malformed/literal-backslash-t, duplicate, empty, subset, reconstruction, and `eval` paths. Use owned `$GOTMPDIR/bin/<six-digit-ordinal>-<sha256-import-path>` targets and unlink after recorded result.
- [ ] Commit: `git add testframework/docker/backend_unit/plan.sh && git commit -m "testframework: add hard-2gib package plan helper"`.

### Task 15: IW-11 Stage Executor

**Files:** Create `testframework/docker/backend_unit/execute.sh`.

- [ ] Implement `run_stage(stage, plan) -> stage-result.<stage>.json`; use exactly `go build -p=1 -o <owned-output> <quoted-import-path>`, `go vet -tags hubtest -p=1 <quoted-import-path>`, `go test -tags hubtest -json -p=1 -count=1 -timeout=150s <quoted-import-path>`, and `go test -tags hubtest -json -race -p=1 -count=1 -timeout=150s <quoted-import-path>` for build, vet, unit, and race respectively. Preserve target coverage after ordinary failure, produce five logical artifacts, bounded FIFOs/readers, post-child watermark cleanup, and infrastructure event handling.
- [ ] Commit: `git add testframework/docker/backend_unit/execute.sh && git commit -m "testframework: add hard-2gib stage executor"`.

### Task 16: IW-12 Container Receipt and Entrypoint

**Files:** Create `testframework/docker/backend_unit/receipt.sh`; modify `testframework/docker/entrypoints/backend-unit.sh`.

- [ ] Make `finalize_container() -> int` the sole container terminal writer: owner/plans/stages/inventory, empty `CONTAINER_COMPLETE`, then `container-receipt.json`; never final/observer/SEAL/rename. Entrypoint calls only `/usr/local/lib/hub/backend-unit/runtime.sh`, `/usr/local/lib/hub/backend-unit/plan.sh`, `/usr/local/lib/hub/backend-unit/execute.sh`, and `/usr/local/lib/hub/backend-unit/receipt.sh`, and owns no whole-run deadline.
- [ ] Commit: `git add testframework/docker/backend_unit/receipt.sh testframework/docker/entrypoints/backend-unit.sh && git commit -m "testframework: finalize hard-2gib container receipt"`.

### Task 17: IW-13 Compose, Dockerfile, Root Policy Integration

**Files:** Modify `docker-compose.local.yml`, `testframework/docker/Dockerfile.qa`, `testframework/policies/root-compose-testframework.v1.json`.

- [ ] Apply frozen limits/environment/tmpfs, six read-only roots, literal only writable `./testframework/reports:/reports`, zero arguments, and baked runner assets. Only the `python-tests` stage gets `COPY testframework/observers/ ./observers/`, `ENV PYTHONPATH=/opt/hub/testframework/src:/opt/hub/testframework`, and `ENV PYTHONSAFEPATH=1`. The `backend-unit` stage gets `COPY testframework/docker/backend_unit/ /usr/local/lib/hub/backend-unit/`, CRLF stripping, helper files mode `0555`, and existing `/usr/local/bin/backend-unit` references exact helper paths; do not copy observer to Go stage.
- [ ] Add no observer bind: the TDD topology deliberately uses image-owned RED, then the one admitted rebuild, then GREEN. Preserve permitted existing README copies as human metadata/retained assets, including `testframework/schemas/README.md` copied by existing schema-directory COPY without selective copy, while no HARD2G human prose becomes image/static/semantic oracle. Admit only the two fixed image-inspect commands and enforce the canonical build-input manifest/hash and bound backend `.Image` receipt contract.
- [ ] Commit: `git add docker-compose.local.yml testframework/docker/Dockerfile.qa testframework/policies/root-compose-testframework.v1.json && git commit -m "testframework: integrate hard-2gib runner image contract"`.

### Task 18: IW-14 Human Operational Metadata

**Files:** Modify `testframework/PACKAGE-CONTRACT.md`, `testframework/README.md`, `testframework/schemas/README.md`, `testframework/docker/README.md`, `docs/TESTING.md`, `docs/test_plan.md`, `docs/TROUBLESHOOTING.md`, `docs/stack_and_tools.md`.

- [ ] Document command; root build/GREEN/observer choreography; Docker stage placement; six roots and only report bind; resource contract; budget; candidate identity/image binding; canonical manifest snapshots; host preflight/runtime stop rules; publication; exits; immutable later non-mutating triage; lease limitation; and `NOT_RUN`. State that Tasks 5–18 are one root-owned implementation session and this documentation is complete before whole-branch review or runtime; do not make HARD2G prose an oracle.
- [ ] Commit: `git add testframework/PACKAGE-CONTRACT.md testframework/README.md testframework/schemas/README.md testframework/docker/README.md docs/TESTING.md docs/test_plan.md docs/TROUBLESHOOTING.md docs/stack_and_tools.md && git commit -m "docs: describe hard-2gib backend-unit evidence"`.

## Whole-Branch Review

### Task 19: RV-01 Independent Terra/High Whole-Branch Review

**Files:** No source edit until an accepted finding.

- [ ] Submit the Global SDD Review Package Contract for `SDD BASE..HEAD` after every atom has passed its own review/re-review to an independent reviewer dispatched with explicit `model: gpt-5.6-terra`, `reasoning_effort: high`. Review policies/schemas, safe paths/lease/authority, image runner, Compose/Dockerfile/root policy, fixtures/selftests, and docs against all governing specifications and frozen values.
- [ ] Require raw explicit `PASS`, `FAIL`, or `NOT_RUN` for architecture/provenance, security/authority, test semantics, and documentation factuality. requested/actual model+effort telemetry is recorded or `UNKNOWN`, which is not PASS. A writer-only exact-path remediation handoff repairs each accepted finding, then a distinct read-only re-reviewer receives the changed whole branch with explicit `model: gpt-5.6-terra`, `reasoning_effort: high`. At most two repair/re-review cycles; otherwise root decomposes or marks `BLOCKED`. Continue only with zero unresolved critical, important, or minor findings. No Docker command in review/fix/re-review.

## Final Executable Sequence

### Task 20: FG-01 One Admitted Build

- [ ] Root establishes the cooperative no-write window, records clean lease/preflight (RAM `>=2 GiB`, C: free `>=250 GiB`, three disk samples two seconds apart and 2/3 disk predicate), canonical build-input manifest bytes/hashes, no other `backend-unit`, no Mail/Calendar QA/test/build, and idle BuildKit; then runs exactly once `docker compose -f docker-compose.local.yml build python-contract backend-unit`. Any failure stops without retry. No pull or rebuild is allowed.

### Task 21: FG-02 One Full Contract GREEN

- [ ] Root rechecks the same clean canonical build-input manifest bytes/hashes and no-write window after build, then runs exactly once `docker compose -f docker-compose.local.yml run --rm python-contract`. Proceed only on exit `0` with all static, fixture, semantic, observer/security, Compose/Docker/launcher cases green. Any other result stops before observer/backend invocation.

### Task 22: FG-03 Ready Observer and Canonical Backend Run

- [ ] Root rechecks the same clean canonical build-input manifest bytes/hashes and no-write window after GREEN, starts exactly `python testframework/observers/backend_unit/observe.py --compose-project worktree-master` in a managed session, waits at most 30 seconds for `READY observer_session=<32-lower-hex>\n`, then separately runs exactly once `docker compose -f docker-compose.local.yml run --rm backend-unit`, and waits for observer exit. The observer has acquired its exclusive non-symlink lease after disk preflight and before `READY`, recomputed/bound manifest bytes/hash and both fixed image-inspect IDs, and verifies bound backend container `.Image` equals the inspected backend image ID.
- [ ] Accept only a sealed packet with exactly one candidate whose labels, start window, full 64-lower-hex Engine ID, and owner-marker hostname prefix match; require identical `run_id` and `nonce` through observer/final/`SEAL`, lifecycle/resource/identity/manifest/image facts, one-stop compliance, semantic validation, and final rename. Otherwise preserve `UNSEALED_ORPHAN / NONAUTHORITATIVE` and exit `75`.

### Task 23: FG-04 Evidence Closure

**Files:** Modify `.planning/STATE.md`, `.planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-VERIFICATION.md` only after observed facts.

- [ ] Obtain independent final architecture/provenance, security/authority, test-semantics, and publication/evidence reviews, each dispatched with explicit `model: gpt-5.6-terra`, `reasoning_effort: high`; record facts, never manufacture a PASS.
- [ ] Record commands, timestamps, fixed image IDs, raw samples/predicates, canonical manifest bytes/hashes, container ID and bound `.Image`, receipt hashes, classification and immutable outcome, exits, review verdicts, requested/actual model+effort telemetry or `UNKNOWN`, and hostile-host lease limitation. Promote only `SEALED_EXECUTABLE_PASS` plus four PASS reviews; otherwise retain truthful non-GREEN status.
- [ ] Commit: `git add .planning/STATE.md .planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-VERIFICATION.md && git commit -m "docs: record observed hard-2gib runner evidence"`.

## Self-Audit Checklist

- [ ] The three and only three HARD2G top-level schemas are literal and exhaustive without claiming control over unrelated repository schemas; the README exception includes `testframework/schemas/README.md` as existing-COPY human metadata, never an oracle.
- [ ] All 15 sealed enums are literal, exhaustive, and no duplicate QA priority exists. `SEALED_TIMEOUT_UNTRIAGED` remains the sealed enum and immutable `TIMEOUT_UNTRIAGED` is never later rewritten.
- [ ] `ClassificationFacts` is source mutation > invalid provenance > earliest explicit stop with the stated tie mapping > OOM > resource > budget > tmpfs > timeout > discovery > stage > pass.
- [ ] Every implementation atom has the Global SDD Review Package Contract, explicit Terra/xhigh implementer routing, distinct read-only Terra/high spec-and-quality review, writer-only repair commit if needed, distinct read-only explicit Terra/high re-review, a maximum of two repair/re-review cycles, and later whole-branch review; no inherit or fallback.
- [ ] The observer uses literal `--compose-project worktree-master`, all fixed `shell=False` argv vectors including both fixed image inspect commands, local label checks, one unique candidate binding, line streaming/termination/reap, 16-record/64-KiB events and 1MiB/10s one-shot bounds, and every failure no-seal `75`.
- [ ] The observer acquires the exclusive regular non-symlink lease after disk preflight before `READY`, retains it through terminal verification/post snapshots, and root establishes/rechecks the cooperative no-write window before FG-01, after build, and after GREEN; hostile-host no-ACL limitation remains explicit.
- [ ] Exactly one writable bind is `./testframework/reports:/reports`; exactly six source binds are read-only. Dockerfile observer placement, Go helper placement/CRLF/0555/entrypoint paths, frozen digest mappings, paths, commands, and README exception are preserved.
- [ ] The canonical build-input manifest has exactly Compose, Dockerfile, entrypoint, four helpers, observer package, three HARD2G schemas, two new policies, root policy, and root selftests/fixtures; root snapshots it before build and after GREEN, observer recomputes it after lease, both fixed image IDs are bound, backend `.Image` equals inspected backend image ID, and any mismatch is no-seal `75`.
- [ ] Task 14 contains exactly three discovery commands: build first, vet second, unit/race third, `LC_ALL=C` lexical order, a literal `0x09` delimiter and rejection of `0x5c 0x74`; Task 15 exact argv profiles preserve `-race` before `-p=1`.
- [ ] Exactly one RED, one build, one post-build `python-contract` GREEN, one READY observer, and one canonical backend run are allowed; all unobserved outcomes remain non-GREEN.



