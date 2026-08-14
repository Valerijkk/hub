# G0 Assets and Runner: Pre-Genesis Bootstrap Plan

**Status:** planning only. This plan does not accept a Phase 129 batch, the two recovered bulk moves, or the 241 staged legacy deletions.

## Goal and recovery boundary

G0 creates one immutable, auditable input for the Option-A successor-authority recovery. G0 activation is a **pre-genesis input**, never a recovery event and never evidence that a v1 manifest batch was applied or verified.

The initial state remains:

- `STATIC_PATH_MAPPING=PASS`;
- `CUTOVER_ACCEPTANCE=FAIL/BLOCKED`;
- 241 legacy paths are staged and uncommitted; they must remain untouched;
- package and launcher are `NOT_READY`; all runtime evidence is `NOT_RUN`.

`pro-qa-ai/**` is excluded: do not read, enumerate, mount, modify, or name it in image inputs. Python package bootstrap and CI binding are out of scope. This plan creates neither `pyproject.toml`, `uv.lock`, `py.typed`, Python suites, nor CI workflow changes.

## Immutable authority and causal order

The required order is closed and forward-only:

1. Commit this G0 plan.
2. Founder supplies external offline ingress and separately releases one local-only bootstrap exception.
3. Source Packages A and B are authored, independently statically reviewed, and committed.
4. The one authorized `hub.qa-bootstrap/v1` transaction admits assets, builds and attests four images, writes one task-owned candidate lock at `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`, runs only `g0-planning-selftest`, and performs exact cleanup.
5. Package C1 independently compares that candidate's raw bytes with Package C receipts and local attestations, then commits only `testframework/docker/images.lock.v3.json` without regenerating or editing it.
6. An independent evidence reviewer verifies the C/C1 transaction and creates the already-existing evidence acceptance commit. All later evidence acceptance and activation use the C1 commit as `repository_commit`.
7. A separate later commit writes the G0 activation record, whose `acceptance_commit` names step 6. The activation file cannot attest its own commit.
8. Only then may Recovery Package A consume activation as its pre-genesis input.

There is no repository lock prerequisite before bootstrap: the only lock produced by step 4 is the task-owned candidate at `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`. Package C1 is the only later path by which identical candidate bytes may enter the repository. Source contract tests are written failing-first but are `NOT_RUN` before bootstrap. The founder release explicitly authorizes this exception to normal red-run sequencing. After image creation, the bootstrap profile runs the source contract fixtures through runner v2 and retains the required negative-red and positive-green proof before activation. It never runs a SUT target.

## Global safety constraints

- Preserve the existing index exactly: no staging, unstaging, deletion, restore, reset, checkout, clean, rebase, merge, or commit that includes any of the 241 legacy paths.
- Every source package is committed only with its literal `git commit --only` allowlist below. No package may modify an unlisted path.
- No host test, validator, linter, build, vet, Vite, pytest, Playwright, Locust, direct `docker`, direct `docker compose`, legacy `hubtest`, `run-task.ps1`, or direct `run-task-v2.ps1` invocation is a fallback.
- The only pre-run authority is static independent review. The only runtime authority is a separately founder-released, local-only `hub.qa-bootstrap/v1` transaction. Until that release, all runtime states are `NOT_READY/NOT_RUN`.
- Bootstrap is not `/qa-gate`, a public runner route, direct Docker/Compose, or a host test fallback. Its outer launcher owns the complete build, admission, attestation, run, evidence publication, and cleanup lifecycle.
- No network, registry login, pull, push, remote lookup, cache lookup, package installation, or dependency resolution is permitted. If exact offline dependency bytes or locally preloaded base identities are absent, the result is `BLOCKED`.
- Every bootstrap container is `linux/amd64`, network-disabled, read-only except its single task-owned report mount, `cpus: 2.0`, `mem_limit: 2g`, `pids_limit: 256`, `cap_drop: [ALL]`, and `no-new-privileges:true`.

## Closed bootstrap contract

`testframework/runner/contracts/hub.qa-bootstrap.v1.json` defines the only bootstrap outer contract, `hub.qa-bootstrap/v1`. Its implementation is `testframework/runner/invoke-qa-bootstrap-v1.ps1`; its control topology is `testframework/docker/compose.g0-bootstrap.yaml` and `testframework/docker/bootstrap-control.v1.json`.

It accepts exactly one founder release record plus `task=g0-assets-runner` and a grammar-valid unique `run`; all other caller input is rejected before report creation. It creates `testframework/reports/<task>-<run>/` with an owner marker using create-new semantics, then owns this fixed sequence:

1. verify the explicit founder release and local-only prerequisites;
2. admit exact offline dependency bytes and local preloaded base image IDs/digests/licenses/provenance;
3. build all four fixed targets from those local bases;
4. attest their local image IDs and generate the candidate `testframework/reports/<task-run>/bootstrap/images.lock.v3.json` once;
5. invoke only the fixed `g0-planning-selftest` entrypoint through runner v2;
6. publish closed task receipts and preserve the first non-zero exit;
7. enforce timeout and execute exact owned-resource cleanup in `finally`;
8. emit report and cleanup receipts, retaining failed evidence.

It has no arbitrary command, image, path, environment, retry, install, network, pull, or SUT-target capability. A timeout uses exit `124`; an earlier failure always wins over cleanup failure. Cleanup selects only resources carrying the exact owner marker from the initial owned-resource inventory; foreign resources and final reports are never adopted or deleted.

External ingress must provide, before bootstrap, every byte and evidence item in the closed dependency bundle plus all locally preloaded pinned base image IDs/digests, SPDX licenses, license-evidence digests, and provenance digests. Missing, mutable, remote, tagged-only, or nonlocal input is `BLOCKED_EXTERNAL_INGRESS`.

## Fixed source deliverables

All static source packages below are authored and independently accepted before the founder may release bootstrap. Existing historical `source-policy.v1.json` and `build-manifest.v2.json` remain byte-for-byte unchanged.

| Concern | New or changed files |
|---|---|
| Dependency admission | `testframework/docker/dependency-bundle.schema.v1.json`, `testframework/docker/dependency-bundle.manifest.v1.json`, `testframework/docker/verify-local-dependency-bundle.ps1`, `testframework/docker/admit-local-dependency-assets.ps1`, `testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1` |
| Source/build provenance | `testframework/docker/source-policy.v2.json`, `testframework/docker/build-manifest.v3.json`, `testframework/docker/source-image-attestor.ps1`, `testframework/selftests/runner/g0-image-contract.ps1` |
| Four image targets | `testframework/docker/Dockerfile.backend-unit`, `testframework/docker/Dockerfile.frontend-unit`, `testframework/docker/Dockerfile.backend-openapi`, `testframework/docker/Dockerfile.g0-planning-selftest`, `testframework/docker/entrypoints/backend-unit.sh`, `testframework/docker/entrypoints/frontend-unit.sh`, `testframework/docker/entrypoints/backend-openapi.sh`, `testframework/docker/entrypoints/g0-planning-selftest.sh`, `testframework/docker/vitest.frontend-unit.config.ts`, `testframework/docker/compose.g0-bootstrap.yaml`, `testframework/docker/bootstrap-control.v1.json` |
| Bootstrap and v2 runner | `testframework/runner/contracts/hub.qa-bootstrap.v1.json`, `testframework/runner/invoke-qa-bootstrap-v1.ps1`, `testframework/runner/invoke-qa-gate-v2.ps1`, `testframework/runner/run-task-v2.ps1`, `testframework/runner/new-repository-snapshot.ps1`, `testframework/runner/new-run-envelope.ps1`, `testframework/runner/write-g0-activation.ps1`, `testframework/schemas/repository-snapshot.schema.v1.json`, `testframework/schemas/run-envelope.schema.v1.json`, `testframework/selftests/runner/g0-runner-v2-contract.ps1`, `testframework/runner/README.md`, `.cursor/skills/qa-gate/SKILL.md` |

The four immutable targets are exactly `backend-unit`, `frontend-unit`, `backend-openapi`, and `g0-planning-selftest`. The first three are never executed by G0. `g0-planning-selftest` is an image-baked runner-fixture target, not an end-user or SUT QA target.

### Dependency admission contract

`dependency-bundle.manifest.v1.json` is strict UTF-8 without BOM and schema `hub.qa.dependency-bundle/v1`. Its closed top-level keys, in this order, are `schema`, `bundle_id`, `source_commit`, `platform`, `assets`, `bases`, and `bundle_sha256`. `source_commit` equals the recovery baseline named by the committed G0 plan. `platform` is exactly `{"os":"linux","architecture":"amd64","variant":null}`.

Each `assets` item has exactly `asset_id`, `targets`, `relative_path`, `sha256`, `bytes`, `license_spdx`, `license_evidence_sha256`, and `provenance_sha256`. The only asset identities are `go-module-cache`, `go-race-toolchain`, `frontend-node-modules`, and `swag-binary`; their closed target mappings are respectively `["backend-unit","backend-openapi"]`, `["backend-unit"]`, `["frontend-unit"]`, and `["backend-openapi"]`. Each `bases` item has exactly `base_id`, `targets`, `local_image_id`, `manifest_digest`, `license_spdx`, `license_evidence_sha256`, and `provenance_sha256`; it is valid only when its `local_image_id` is already present locally and agrees with its pinned digest. All digest values use `sha256:` plus 64 lowercase hexadecimal characters.

`Test-HubQaLocalDependencyBundle` verifies exact keys, duplicate-aware JSON, strict decoding, path containment below `assets/`, byte counts, SHA-256, licenses, provenance, reparse points, and all four bases before derived output. It returns exactly `{ Verdict; ManifestSha256; Assets; Bases; ReasonCode }`; allowed reason codes are `E_MANIFEST_UTF8`, `E_MANIFEST_SCHEMA`, `E_DUPLICATE_ASSET_ID`, `E_ASSET_PATH_ESCAPE`, `E_ASSET_MISSING`, `E_ASSET_BYTES_MISMATCH`, `E_ASSET_SHA256_MISMATCH`, `E_LICENSE_EVIDENCE_MISSING`, `E_PROVENANCE_EVIDENCE_MISSING`, `E_BASE_MISSING`, `E_BASE_ID_MISMATCH`, and `E_SOURCE_COMMIT_MISMATCH`.

## Four-target image and lock contract

`source-policy.v2.json` is `hub.qa.local-source-policy/v2`; `build-manifest.v3.json` is `hub.qa.local-images-build-manifest/v3`. They bind the committed recovery source revision, sorted source tree digest, four target definitions, entrypoint digests, the dedicated Vitest config digest, bundle digest, and no network or registry capability.

The frontend target must use only `testframework/docker/vitest.frontend-unit.config.ts`, whose include is exactly `/opt/hub/source/frontend/tests/**/*.test.{ts,tsx}`. It has no exclusion that can hide a failure. The three SUT entrypoints are fixed source contracts only: backend unit runs the declared isolated `backend/tests/...` sequence, frontend unit runs its fixed TypeScript/Biome/Vite/Vitest sequence, and OpenAPI compares two independent fixed generations. Their exact commands are frozen in `build-manifest.v3.json`; none may reference legacy test roots, `testframework/stages`, a network client, or installation command.

`Dockerfile.g0-planning-selftest` and `entrypoints/g0-planning-selftest.sh` execute only injected runner-fixture contract cases: report staging/publication, owner marker, snapshot/envelope validation, controlled negative-red cases, positive-green case, first-exit, timeout, and exact cleanup. It does not invoke backend, frontend, OpenAPI, Python, Playwright, legacy tests, or application source commands.

The bootstrap transaction generates exactly one candidate lock at `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`; it never hand-authors a repository lock as a prerequisite. The candidate has exactly four `targets` records keyed `backend-unit`, `frontend-unit`, `backend-openapi`, and `g0-planning-selftest`. Every target child has exactly `local_image_id`, `local_image_id_sha256`, `entrypoint_path`, `entrypoint_sha256`, `command_sha256`, `compose_service`, and `runtime_status`; those are the identical seven target-child names later required by activation. The planning record binds the fixed `g0-planning-selftest` command through `command_sha256`. The lock root binds `linux/amd64`; all target children bind locally inspected immutable image IDs and `runtime_status=READY_FOR_G0_BOOTSTRAP_ONLY` until a later non-G0 authority says otherwise. Only Package C1 may copy those already-generated bytes, unchanged, to `testframework/docker/images.lock.v3.json` after independent comparison.

`compose.g0-bootstrap.yaml` contains exactly four named services, one per target, all `pull_policy: never`; only the planning service can be selected by bootstrap. Its only writable mount is the task-owned reports root. `bootstrap-control.v1.json` binds Dockerfile, entrypoint, compose service, local-image lock field, and fixed command for every target.

## Shared snapshot and envelope contracts

There is one shared recovery-compatible schema at `testframework/schemas/repository-snapshot.schema.v1.json`, `$id` `hub.qa.repository-snapshot/v1`. It has no additional top-level properties and requires, in this exact order:

`schema`, `producer`, `snapshot_id`, `captured_for`, `task`, `run`, `envelope_id`, `role_claim`, `self_attested`, `source_revision`, `head`, `index`, `worktree`, `commits`, `ancestry`, `control`, `evidence`, `lifecycle`.

Its child objects are also closed:

- `captured_for`: `kind`, `ordinal`, `subject_path`, `subject_sha256`, `subject_commit_sha`, `predecessor_event_path`, `predecessor_event_sha256`, `predecessor_event_commit_sha`, `pre_event_snapshot_path`, `pre_event_snapshot_sha256`.
- `source_revision`: `commit`, `tree`, `captured_at_utc`.
- `head`: `commit`, `tree`, `event` where `event` is exactly `event_id`, `event_type`, `captured_at_utc`.
- `index`: `paths`, `count`, `sha256`; every path capture is exactly `path`, `status`, `blob_sha256`.
- `worktree`: `scoped_paths`, `count`, `sha256`; every file capture is exactly `path`, `state`, `blob_sha256`.
- `commits`: an array of closed entries `commit`, `parent`, `tree`, `changed_paths`; each changed-path entry is exactly `path`, `status`, `blob_sha256`.
- `ancestry`: `edges`; each edge is exactly `child`, `parent`.
- `control`: `g0_plan_sha256`, `bootstrap_contract_sha256`, `runner_contract_sha256`, `dependency_bundle_manifest_sha256`, `source_policy_sha256`, `build_manifest_sha256`, `images_lock_sha256`.
- `evidence`: `report_path`, `report_sha256`, `cleanup_path`, `cleanup_sha256`.
- `lifecycle`: `state`, `started_at_utc`, `finished_at_utc`.

`captured_for.kind` is exactly one of `G0_BOOTSTRAP`, `GENESIS_BASELINE`, `OWNER_MUTATION_AFTER`, `OWNER_RECEIPT_AFTER`, `INDEPENDENT_VERDICT_AFTER`, `ROLLBACK_MUTATION_AFTER`, `ROLLBACK_RECEIPT_AFTER`, or `ROLLBACK_VERDICT_AFTER`; `ordinal` is a non-negative integer and is monotonic within one successor-authority ledger. Every non-null `*_path` is a repository-relative regular-file path and every non-null `*_sha256` is the SHA-256 of its no-BOM raw bytes at its named commit. Every non-null `*_commit_sha` is a full 40-lowercase-hex existing commit.

| `kind` | `ordinal` | `subject_path` / `subject_sha256` | `subject_commit_sha` | predecessor event triple | pre-event snapshot pair |
|---|---:|---|---|---|---|
| `G0_BOOTSTRAP` | `0` | task-owned bootstrap `run.json` path/digest | `null` (report not committed) | `null` | `null` |
| `GENESIS_BASELINE` | `0` | committed `recovery-successor-genesis.json` path/digest | genesis commit | `null` | `null` |
| `OWNER_MUTATION_AFTER` | batch event ordinal | `null` (mutation is a commit, not an event file) | deletion mutation commit | preceding event or `null` for first batch | genesis baseline snapshot for first batch or preceding event's after-snapshot thereafter |
| `OWNER_RECEIPT_AFTER` | current | owner-receipt event path/digest | containing commit | preceding event or `null` for first owner receipt | current `OWNER_MUTATION_AFTER` snapshot |
| `INDEPENDENT_VERDICT_AFTER` | current | independent-verdict event path/digest | containing commit | current owner-receipt event | `OWNER_RECEIPT_AFTER` snapshot |
| `ROLLBACK_MUTATION_AFTER` | batch event ordinal | `null` (mutation is a commit, not an event file) | rollback mutation commit | FAIL independent-verdict event | FAIL `INDEPENDENT_VERDICT_AFTER` snapshot |
| `ROLLBACK_RECEIPT_AFTER` | current | rollback-receipt event path/digest | containing commit | FAIL verdict event | `ROLLBACK_MUTATION_AFTER` snapshot |
| `ROLLBACK_VERDICT_AFTER` | current | rollback-verdict event path/digest | containing commit | rollback receipt event | `ROLLBACK_RECEIPT_AFTER` snapshot |

In the table, a `null` predecessor triple or pre-event snapshot pair means every child field in that triple or pair is JSON `null`. All non-null path, digest, and commit values follow the exact repository-relative path, no-BOM raw-bytes SHA-256, and lowercase 40-hex commit rules above. `G0_BOOTSTRAP` is produced by Package C before recovery and is neither a successor-authority event nor a pre-event snapshot. `GENESIS_BASELINE` begins the successor-authority ledger from the committed recovery-successor genesis. Every non-null pre-event snapshot already exists before the current event commit; an after-event snapshot is written only after the event commit and is then the lawful pre-event evidence for a later event. This prevents self-reference and future reference.

Every snapshot report location is deterministic and task-owned: `testframework/reports/<task-run>/snapshots/<ordinal>-<kind>.repository-snapshot.v1.json`, where `<kind>` is the literal `captured_for.kind` and `<ordinal>` is zero-padded decimal. The producer creates the path once with create-new semantics, writes only after the required commit already exists, records `lifecycle.state=CAPTURED`, and never overwrites, moves, or regenerates it. `G0_BOOTSTRAP` is therefore `.../snapshots/000-G0_BOOTSTRAP.repository-snapshot.v1.json`; `GENESIS_BASELINE` is `.../snapshots/000-GENESIS_BASELINE.repository-snapshot.v1.json` in its successor task run. An event-after snapshot uses that event's ordinal, is published after the event commit but outside that commit, and becomes the sole pre-event snapshot candidate for the next event. Failed capture retains a distinct report with `lifecycle.state=FAILED`; it creates no event reference and blocks the next event. No snapshot is deleted by cleanup because reports are retained evidence.

The sole producer is `testframework/runner/new-repository-snapshot.ps1`; the activation record binds both that literal path and its SHA-256. It accepts no caller-selected source path and rejects invalid raw bytes, reparse points, changed closed controls, or an unexpected source revision before any container dependency operation.

`testframework/schemas/run-envelope.schema.v1.json`, `$id` `hub.qa.run-envelope/v1`, is a closed schema with top-level keys `schema`, `producer`, `envelope_id`, `contract`, `target`, `task`, `run`, `owner_nonce`, `repository_snapshot_sha256`, `report_root`, `limits`, `cleanup_contract`, `created_at_utc`. `contract` is `hub.qa-runner/v2`; target is one of the four fixed targets. `limits` has exactly `command_seconds`, `target_seconds`, `inspect_seconds`, `outer_seconds`, `cleanup_seconds` and values `180`, `780`, `30`, `810`, `30`; `cleanup_contract` is `hub.qa.exact-cleanup/v1`. Its only producer is `testframework/runner/new-run-envelope.ps1` and the activation record binds that path and digest.

## Activation record contract

`.planning/audits/test-rebase/g0-activation.v1.json` uses schema `hub.qa.g0-activation/v1`, is written only after bootstrap and an independent post-run evidence acceptance commit, and has exactly these top-level fields in this order:

`schema`, `activation_id`, `g0_plan_path`, `required_plan_commit`, `repository_commit`, `bootstrap_contract`, `bootstrap_contract_path`, `bootstrap_contract_sha256`, `runner_contract`, `runner_contract_path`, `runner_contract_sha256`, `dependency_bundle_manifest_path`, `dependency_bundle_manifest_sha256`, `source_policy_path`, `source_policy_sha256`, `build_manifest_path`, `build_manifest_sha256`, `images_lock_path`, `images_lock_sha256`, `target_entrypoints`, `planning_selftest`, `repository_snapshot_schema_path`, `repository_snapshot_schema_sha256`, `run_envelope_schema_path`, `run_envelope_schema_sha256`, `repository_snapshot_producer_path`, `repository_snapshot_producer_sha256`, `run_envelope_producer_path`, `run_envelope_producer_sha256`, `report_cleanup_contract`, `acceptance_verdict`, `acceptance_commit`, `created_at_utc`.

Path/digest semantics are exact: each `*_path` is a repository-relative regular file path; its companion `*_sha256` is the SHA-256 of that file's no-BOM raw bytes from `repository_commit`, except `g0_plan_path` which is bound by `required_plan_commit`. `bootstrap_contract` is `hub.qa-bootstrap/v1`; `runner_contract` is `hub.qa-runner/v2`; `report_cleanup_contract` is `hub.qa.exact-cleanup/v1`; and `acceptance_verdict` is exactly `ACCEPT`.

`target_entrypoints` is a closed object with exactly the four fixed target names. Each child has exactly `local_image_id`, `local_image_id_sha256`, `entrypoint_path`, `entrypoint_sha256`, `command_sha256`, `compose_service`, and `runtime_status`; every field must exactly equal the corresponding Package C1 repository-lock child and the bootstrap-control child. `planning_selftest` has exactly `target`, `task`, `run`, `run_receipt_sha256`, `cleanup_receipt_sha256`, `snapshot_sha256`, `envelope_sha256`, and `verdict`; target is `g0-planning-selftest`, task is `g0-assets-runner`, and verdict is `ACCEPT`.

`acceptance_commit` is an existing full 40-lowercase-hex independent evidence acceptance commit. It cannot equal the later activation-record commit. `repository_commit` is exactly the already committed Package C1 lock commit; it is the baseline for all later evidence acceptance and activation comparisons. `required_plan_commit` identifies the earlier plan commit. `Write-HubQaG0Activation` uses create-new semantics, validates all values, and cannot create a second activation.

## Execution packages and gates

### Package 0 — commit the plan

**Files:** only `.planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md`.

**STOP:** Any HIGH finding from independent Terra/high plan review blocks source authoring.

```powershell
git commit --only .planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md -m "docs(test-rebase): plan G0 bootstrap pre-genesis"
```

### Package A — static bootstrap admission and source contracts

**Files:**

- `testframework/docker/dependency-bundle.schema.v1.json`
- `testframework/docker/dependency-bundle.manifest.v1.json`
- `testframework/docker/verify-local-dependency-bundle.ps1`
- `testframework/docker/admit-local-dependency-assets.ps1`
- `testframework/docker/source-policy.v2.json`
- `testframework/docker/build-manifest.v3.json`
- `testframework/docker/source-image-attestor.ps1`
- `testframework/docker/Dockerfile.backend-unit`
- `testframework/docker/Dockerfile.frontend-unit`
- `testframework/docker/Dockerfile.backend-openapi`
- `testframework/docker/Dockerfile.g0-planning-selftest`
- `testframework/docker/entrypoints/backend-unit.sh`
- `testframework/docker/entrypoints/frontend-unit.sh`
- `testframework/docker/entrypoints/backend-openapi.sh`
- `testframework/docker/entrypoints/g0-planning-selftest.sh`
- `testframework/docker/vitest.frontend-unit.config.ts`
- `testframework/docker/compose.g0-bootstrap.yaml`
- `testframework/docker/bootstrap-control.v1.json`
- `testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1`
- `testframework/selftests/runner/g0-image-contract.ps1`

**Failing-first source contracts:** author literal valid/duplicate/hash/path-escape/missing-asset fixtures for admission and local-ID/four-target/config/fixed-command fixtures for images. They must state `NOT_RUN` pre-bootstrap. No claim of red or green execution is permitted.

**Static acceptance:** independent Terra/high verifies closed keys, no network/install/source bind/arbitrary command, historical v1/v2 policy files unchanged, exact four target mapping, and no invented dependency or image digest. A HIGH finding blocks the package.

**STOP:** Without the founder's offline ingress and separate bootstrap release, do not build, attest, execute a selftest, or write a lock-derived runtime receipt. Source package status is `STATIC_ACCEPTED` only after review; runtime is `NOT_READY/NOT_RUN`.

```powershell
git commit --only testframework/docker/dependency-bundle.schema.v1.json testframework/docker/dependency-bundle.manifest.v1.json testframework/docker/verify-local-dependency-bundle.ps1 testframework/docker/admit-local-dependency-assets.ps1 testframework/docker/source-policy.v2.json testframework/docker/build-manifest.v3.json testframework/docker/source-image-attestor.ps1 testframework/docker/Dockerfile.backend-unit testframework/docker/Dockerfile.frontend-unit testframework/docker/Dockerfile.backend-openapi testframework/docker/Dockerfile.g0-planning-selftest testframework/docker/entrypoints/backend-unit.sh testframework/docker/entrypoints/frontend-unit.sh testframework/docker/entrypoints/backend-openapi.sh testframework/docker/entrypoints/g0-planning-selftest.sh testframework/docker/vitest.frontend-unit.config.ts testframework/docker/compose.g0-bootstrap.yaml testframework/docker/bootstrap-control.v1.json testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1 testframework/selftests/runner/g0-image-contract.ps1 -m "feat(qa): add G0 bootstrap assets"
```

### Package B — static bootstrap and runner v2 controls

**Files:**

- `testframework/runner/contracts/hub.qa-bootstrap.v1.json`
- `testframework/runner/invoke-qa-bootstrap-v1.ps1`
- `testframework/runner/invoke-qa-gate-v2.ps1`
- `testframework/runner/run-task-v2.ps1`
- `testframework/runner/new-repository-snapshot.ps1`
- `testframework/runner/new-run-envelope.ps1`
- `testframework/runner/write-g0-activation.ps1`
- `testframework/schemas/repository-snapshot.schema.v1.json`
- `testframework/schemas/run-envelope.schema.v1.json`
- `testframework/selftests/runner/g0-runner-v2-contract.ps1`
- `testframework/runner/README.md`
- `.cursor/skills/qa-gate/SKILL.md`

`hub.qa-runner/v2` accepts only validated `target`, `task`, and `run`; it rejects extras with exit `64`, has no v1 bridge, no public direct invocation, and no general command operation. It supports the four fixed target names but reports all future Python jobs `NOT_READY`. The G0 selftest is callable only from the bootstrap launcher with a valid internal role claim; `/qa-gate` documents v2 but is not a bootstrap route.

**Failing-first source contracts:** author fixtures for extra input, wrong task, duplicate report root, snapshot mismatch, envelope lock mismatch, first-exit preservation, timeout `124`, foreign-resource exclusion, and cleanup. They are `NOT_RUN` until bootstrap.

**Static acceptance:** independent Terra/high checks closed schemas, producer-only paths, no arbitrary inputs or v1/v2 bridge, owner marker before any dependency operation, no activation before accepted evidence, and that the bootstrap exception remains founder-released.

**STOP:** A HIGH finding, an absent Package A independent verdict, or a request to execute any source contract without founder release blocks Package B. Do not create a runtime receipt.

```powershell
git commit --only testframework/runner/contracts/hub.qa-bootstrap.v1.json testframework/runner/invoke-qa-bootstrap-v1.ps1 testframework/runner/invoke-qa-gate-v2.ps1 testframework/runner/run-task-v2.ps1 testframework/runner/new-repository-snapshot.ps1 testframework/runner/new-run-envelope.ps1 testframework/runner/write-g0-activation.ps1 testframework/schemas/repository-snapshot.schema.v1.json testframework/schemas/run-envelope.schema.v1.json testframework/selftests/runner/g0-runner-v2-contract.ps1 testframework/runner/README.md .cursor/skills/qa-gate/SKILL.md -m "feat(qa): add G0 bootstrap runner controls"
```

### Package C — founder-released one-time local bootstrap

**Repository source files:** none. This runtime package consumes the accepted Package A and B commits and produces only task-owned reports, including the candidate `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`. It never creates, edits, stages, or commits `testframework/docker/images.lock.v3.json`.

**Founder prerequisites:** one explicit release naming the exact plan commit, source package commits, task, unique run, local-only execution authority, accepted ingress manifest, all dependency assets, all local base IDs/digests/licenses/provenance, and permission for this one bootstrap exception. An absent item is `G0_C=BLOCKED`.

**Only permitted operation:** call `hub.qa-bootstrap/v1` once. It admits assets, builds all four targets, locally attests them, writes the candidate v3 lock at its fixed task-owned path, invokes only `g0-planning-selftest` through v2, then runs its source contract fixtures through that same internal v2 profile for the required negative-red and positive-green proof. It does not run `backend-unit`, `frontend-unit`, `backend-openapi`, a Python target, or any legacy suite.

**Evidence:** report root, owner marker, admission receipt, immutable local image-ID attestations, candidate v3-lock receipt and raw candidate bytes, closed snapshot, closed envelope, negative-red receipt, positive-green receipt, ordered command receipts, `run.json`, and `cleanup.json`. Candidate lock bytes remain under the task-owned report root for review; they are retained with failed evidence and are never regenerated for C1. The transaction preserves the first exit, applies fixed timeouts, and unconditionally cleans exactly the initially owned resources.

**STOP:** Any missing receipt, unexpected negative pass, failed positive case, nonlocal image, leak, timeout, nonzero exit, digest disagreement, report collision, or cleanup ambiguity is `G0_C=FAIL/BLOCKED`; preserve evidence and do not activate.

### Package C1 — independent candidate-lock admission

**Input:** exactly one successful Package C report root, its immutable candidate `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`, the candidate SHA-256 receipt, all four immutable local image-ID attestations, `bootstrap-control.v1.json`, and the accepted Package A/B commit SHAs. **Output:** either `G0_C1=REJECT/BLOCKED` with no repository mutation, or one repository file `testframework/docker/images.lock.v3.json` whose raw no-BOM bytes are byte-for-byte identical to the candidate. C1 never rebuilds, reattests, regenerates, normalizes, or edits a lock.

**Independent review:** Terra/high compares each target's `local_image_id`, `local_image_id_sha256`, `entrypoint_path`, `entrypoint_sha256`, `command_sha256`, `compose_service`, and `runtime_status` against the candidate, raw C receipts, local attestations, and bootstrap control. It verifies all four target names, candidate digest, report containment, and that no SUT target ran. Its review decision must precede the C1 commit.

**Allowlist:** only `testframework/docker/images.lock.v3.json`.

**STOP:** Missing, mismatched, regenerated, nonlocal, extra, or unverifiable candidate data is `G0_C1=REJECT/BLOCKED`; do not write a repository lock, do not stage any other path, and do not continue to D. On success, the C1 commit is the only `repository_commit` allowed in D and E.

```powershell
git commit --only testframework/docker/images.lock.v3.json -m "feat(qa): admit G0 bootstrap image lock"
```

### Package D — independent evidence acceptance

**Files:** create only `.planning/audits/test-rebase/g0-bootstrap-evidence-acceptance.v1.json`.

The independent Terra/high reviewer checks the frozen contracts, the C1 `repository_commit`, and every Package C receipt against raw bytes, the closed shared snapshot, the four C1 lock records, exact resource cleanup, first-exit semantics, and proof that no SUT target ran. The acceptance record must name that exact C1 SHA as `repository_commit`; no A, B, C report, D, or E commit is valid in this field. Its closed verdict is `ACCEPT` or `REJECT`. `ACCEPT` is committed separately before activation; `REJECT` creates no activation record.

```powershell
git commit --only .planning/audits/test-rebase/g0-bootstrap-evidence-acceptance.v1.json -m "docs(test-rebase): accept G0 bootstrap evidence"
```

### Package E — activation record after acceptance

**Files:** create only `.planning/audits/test-rebase/g0-activation.v1.json`.

`Write-HubQaG0Activation` is invoked only after Package D's independent acceptance commit exists and must set `repository_commit` to the already-existing C1 lock commit. `acceptance_commit` points to that already-existing commit, never the Package E commit. Package E therefore records causal evidence without a self-commit cycle. It remains a pre-genesis input, not a recovery event.

**STOP:** Invalid digest, missing Package D `ACCEPT`, non-existing `acceptance_commit`, a self-reference, any batch-state claim, or any attempt to include the staged deletion set blocks activation.

```powershell
git commit --only .planning/audits/test-rebase/g0-activation.v1.json -m "docs(test-rebase): record G0 pre-genesis activation"
```

### Recovery Package A handoff

No file is changed by G0. Recovery Package A may read Package E only as immutable pre-genesis input. It must create a new successor authority and ordered handling for the 241 legacy deletions; it cannot modify the all-planned v1 cutover manifest, infer receipts, mark a historical batch applied or verified, or claim Phase 129 complete.

## Acceptance matrix

| Gate | Static evidence | Runtime evidence | Truthful precondition status |
|---|---|---|---|
| Plan | independent Terra/high plan review | none | `BLOCKED` on HIGH finding |
| A | dependency/source/build/image source contracts, four-target topology, static contracts | none | `NOT_READY/NOT_RUN` |
| B | bootstrap/v2/shared schemas and producers, static contracts | none | `NOT_READY/NOT_RUN` |
| C | founder release and all accepted static packages | one bootstrap transaction, planning selftest only, red/green fixtures, exact cleanup | `BLOCKED_EXTERNAL_INGRESS` or `BLOCKED` |
| C1 | independent candidate/receipt/local-attestation comparison; only lock allowlist | committed byte-identical repository lock | `REJECT/BLOCKED` unless exact candidate is present |
| D | independent Terra/high C/C1 receipt review | accepted evidence commit naming C1 as `repository_commit` | `REJECT` unless complete |
| E | activation record static validation against C1 plus prior D acceptance | references prior D acceptance only | `BLOCKED` until C1 and D exist |
| Recovery A | activation is immutable input | separate recovery evidence | `CUTOVER_ACCEPTANCE=FAIL/BLOCKED` until separately accepted |

## Rollback and handoff

Rollback is forward-only: revert only an exact named source-package commit after a separate review; never alter the 241 staged paths. A bootstrap failure retains its task-owned evidence and removes only its proven owned resources. No automatic retry, continuation, lock edit, or digest rewrite is permitted; a new founder release must name a new run.

Before source authoring, commit this corrected plan separately and obtain independent Terra/high review. Until all prerequisites exist, hand off exactly: `G0=NOT_READY`, `BOOTSTRAP=NOT_RUN`, `CUTOVER_ACCEPTANCE=FAIL/BLOCKED`.
