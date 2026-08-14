# G0 Assets and Runner: Pre-Genesis Bootstrap Plan

**Status:** planning only. This plan does not accept a Phase 129 batch, the two recovered bulk moves, or the 241 staged legacy deletions.

## Goal and recovery boundary

G0 creates a static source contract first, then one immutable, auditable binding input for the Option-A successor-authority recovery. G0 activation is a **pre-genesis input**, never a recovery event and never evidence that a v1 manifest batch was applied or verified.

The initial state remains:

- `STATIC_PATH_MAPPING=PASS`;
- `CUTOVER_ACCEPTANCE=FAIL/BLOCKED`;
- 241 legacy paths are staged and uncommitted; they must remain untouched;
- package and launcher are `NOT_READY`; all runtime evidence is `NOT_RUN`.

`pro-qa-ai/**` is excluded: do not read, enumerate, mount, modify, or name it in image inputs. Python package bootstrap and CI binding are out of scope. This plan creates neither `pyproject.toml`, `uv.lock`, `py.typed`, Python suites, nor CI workflow changes.

## Immutable authority and causal order

The required order is closed and forward-only:

1. Commit this G0 plan.
2. Package A0 authors and independently accepts the static source contract without dependency or provenance values.
3. Only after A0 independent acceptance, Package B may be authored and statically reviewed in parallel with pending A1. Package B and A1 have no overlapping file ownership.
4. A founder-approved offline-ingress handoff closes one handoff ID and its fixed physical repository-root-derived handoff root. That root supplies exact local bytes, digests, licenses, and provenance; Package A1 binds only those supplied values, receives independent review, and commits no code. The recovery source-tree baseline is permanently `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`; later plan, A0, A1, B, release, C1, evidence, and activation commits are control/repository history, not a substitute source baseline.
5. Only after accepted A0, A1, and B commits and a separate explicit founder approval, Package R writes and commits the sole closed pre-C repository authority artifact `.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json`. The artifact commit is the release; it is repository authority plus an explicit founder gate, not a cryptographic identity, signature, or assertion that cryptographic authentication occurred.
6. Package C derives authority only from that accepted artifact at committed `HEAD` and its commit, then performs one local-only `hub.qa-bootstrap/v1` transaction.
7. That transaction builds and attests four images, writes one task-owned candidate lock at `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`, runs only `g0-planning-selftest`, and performs exact cleanup.
8. Package C1 independently compares that candidate's raw bytes with Package C receipts and local attestations, then commits only `testframework/docker/images.lock.v3.json` without regenerating or editing it.
9. An independent evidence reviewer verifies the C/C1 transaction and creates the already-existing evidence acceptance commit. All later evidence acceptance and activation use the C1 commit as `repository_commit`.
10. A separate later commit writes the G0 activation record, whose `acceptance_commit` names step 9. The activation file cannot attest its own commit.
11. Only then may Recovery Package A consume activation as its pre-genesis input.

There is no repository lock prerequisite before bootstrap: the only lock produced by step 7 is the task-owned candidate at `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`. Package C1 is the only later path by which identical candidate bytes may enter the repository. Source contract tests are written failing-first but are `NOT_RUN` before bootstrap. The explicit founder approval recorded by Package R authorizes this exception to normal red-run sequencing. After image creation, the bootstrap profile runs the source contract fixtures through runner v2 and retains the required negative-red and positive-green proof before activation. It never runs a SUT target.

## Global safety constraints

- Preserve the existing index exactly: no staging, unstaging, deletion, restore, reset, checkout, clean, rebase, merge, or commit that includes any of the 241 legacy paths.
- Every source package is committed only with its literal `git commit --only` allowlist below. No package may modify an unlisted path.
- No host test, validator, linter, build, vet, Vite, pytest, Playwright, Locust, direct `docker`, direct `docker compose`, legacy `hubtest`, `run-task.ps1`, or direct `run-task-v2.ps1` invocation is a fallback.
- The only pre-run authority is static independent review. The only runtime authority is the separately founder-approved and Package-R-committed, local-only `hub.qa-bootstrap/v1` transaction. Until that artifact commit is accepted, all runtime states are `NOT_READY/NOT_RUN`.
- Bootstrap is not `/qa-gate`, a public runner route, direct Docker/Compose, or a host test fallback. Its outer launcher owns the complete build, admission, attestation, run, evidence publication, and cleanup lifecycle.
- No network, registry login, pull, push, remote lookup, cache lookup, package installation, or dependency resolution is permitted. If exact offline dependency bytes or locally preloaded base identities are absent, the result is `BLOCKED`.
- This plan does not authorize network ingress. Network, registry, package, or other external ingress needs later explicit founder approval outside Package R; A0 must not invent values while that approval is pending.
- Every bootstrap container is `linux/amd64`, network-disabled, read-only except its single task-owned report mount, `cpus: 2.0`, `mem_limit: 2g`, `pids_limit: 256`, `cap_drop: [ALL]`, and `no-new-privileges:true`.

## Canonical identifiers and ordered JSON bytes

The G0 bootstrap task identity is exactly `g0-assets-runner`; no alternate task
spelling, alias, or caller-selected task exists in Packages R, B, C, C1, D, or
E. The shared schemas use the one forward-compatible task grammar
`^[a-z][a-z0-9-]{2,31}$` and the one run grammar
`^[a-z0-9][a-z0-9-]{7,63}$` (8--64 characters). G0 copies its fixed task and
grammar-valid run unchanged from the raw Package-R artifact into the bootstrap
owner marker, local-base receipt, private context, repository snapshot, run
envelope, report root, planning-selftest record, and activation record. The
unique G0 report root is therefore exactly
`testframework/reports/g0-assets-runner-<run>/`; the placeholder `<task-run>`
elsewhere in this plan means that literal joined form, never a second delimiter,
escaped component, or caller path. Recovery may use the same schemas only with
a task/run matching these same grammars and the recovery-genesis role rules; it
must not fork a shorter identifier form. The historical v1 `run-task.ps1`
short identifier grammar is not a migration source, compatibility bridge, or
exception. Widening either identifier grammar requires a later contract version
and is not authorized here.

The bootstrap contract, repository-snapshot schema, run-envelope schema,
founder-release artifact, evidence-acceptance artifact, and activation record
all use one ordered canonical JSON encoding for every raw-byte SHA-256 used by
these contracts:

1. bytes are UTF-8 without BOM, strict RFC 8259 JSON, with no bytes before or
   after the JSON value and no insignificant whitespace;
2. object keys occur once and exactly in the order declared in the applicable
   closed schema below; arrays retain the stated producer order; unknown,
   duplicate, reordered, or omitted keys are invalid;
3. `null`, `true`, and `false` are lowercase literals; integers are base-10
   without sign, decimal point, exponent, or leading zero except `0`; strings
   contain literal valid Unicode scalar UTF-8 except `"`, `\\`, and controls,
   which use only `\"`, `\\`, `\b`, `\f`, `\n`, `\r`, `\t`, or uppercase
   `\u00XX` escapes; and
4. `sha256` values are lowercase `sha256:` plus 64 hexadecimal characters;
   a raw-byte digest always hashes the exact no-BOM bytes named by its paired
   path, never a parsed/re-emitted object. A Git commit/tree/blob SHA-1 is
   lowercase 40 hexadecimal characters. A repository-relative path uses `/`,
   names one physical regular file below the repository root, and has no `.`
   segment, `..` segment, empty segment, drive prefix, reparse traversal, or
   absolute form.

No producer may silently normalize a received artifact. Every consumer
duplicate-aware parses, checks this encoding and ordered closed keys before
using values, then compares raw-byte digests. These byte rules deliberately
apply to new G0 v1/v2 records only; they do not rewrite, normalize, or migrate
historical v1 runner records.

## Closed bootstrap contract

`testframework/runner/contracts/hub.qa-bootstrap.v1.json` defines the only bootstrap outer contract, `hub.qa-bootstrap/v1`. Its implementation is `testframework/runner/invoke-qa-bootstrap-v1.ps1`; its control topology is `testframework/docker/compose.g0-bootstrap.yaml` and `testframework/docker/bootstrap-control.v1.json`.

Its ordered canonical top-level keys are exactly `schema`, `contract`, `status`,
`task`, `run_pattern`, `targets`, `internal_role`, `report_root`,
`owner_marker`, `base_receipts`, `source_tree_gate`,
`post_build_attestation`, `release_authority`, `cleanup_contract`, and
`direct_invocation`. `schema` is `hub.qa.bootstrap-contract/v1`; `contract` is
`hub.qa-bootstrap/v1`; `status` is `NOT_READY` until a later accepted C
transaction; `task` is `g0-assets-runner`; and `run_pattern` is exactly
`^[a-z0-9][a-z0-9-]{7,63}$`. Every nested object is closed and has these
ordered keys:

| Key | Closed child fields and exact values |
|---|---|
| `targets` | `all`, `bootstrap_only`; `all` is exactly `backend-unit`, `frontend-unit`, `backend-openapi`, `g0-planning-selftest` in that order and `bootstrap_only` is exactly `g0-planning-selftest`. |
| `internal_role` | `claim`, `transport`, `context_keys`, `direct_invocation_error`; values are `hub.qa-bootstrap/v1`, `in-memory-private-context`, the eight keys in `Internal bootstrap role capability`, and `E_B0_DIRECT_INVOCATION_FORBIDDEN`. |
| `report_root` | `template`, `owner_marker_path`, `retention`; values are `testframework/reports/g0-assets-runner-<run>`, `testframework/reports/g0-assets-runner-<run>/bootstrap/owner-marker.v1.json`, and `retain`. |
| `owner_marker` | `schema`, `path`, `keys`, `nonce_pattern`; values are `hub.qa.bootstrap-owner/v1`, the same fixed `bootstrap/owner-marker.v1.json` path, exactly `schema`, `task`, `run`, `owner_nonce`, and `^[0-9a-f]{32}$`. |
| `base_receipts` | `admission_path`, `admission_schema`, `admission_context`, `image_path`, `image_schema`, `image_context`; values are `bootstrap/local-base-inspect.v1.json`, `hub.qa.local-base-inspect/v1`, `PrivateLocalBaseReceipt`/`PrivateLocalBaseReceiptSha256`/`PrivateLocalBaseReceiptNonce`, `bootstrap/local-image-inspect.v1.json`, `hub.qa.local-image-inspect/v1`, and `RunnerInspectionRawBytes`/`RunnerInspectionSha256`. They are distinct raw, create-new, owner/nonce-bound receipts: the admission receipt exists before helper admission, while the image receipt exists only after four target builds. |
| `source_tree_gate` | `function`, `receipt_path`, `receipt_schema`, `failure_code`; values are `Get-HubQaGitObjectSourceTree`, `bootstrap/source-tree-verification.v1.json`, `hub.qa.source-tree-verification/v1`, and `E_B0_SOURCE_TREE_PREBUILD`. |
| `post_build_attestation` | `function`, `verdict`, `failure_code`; values are `Invoke-HubQaBootstrapImageAttestation`, `READY_FOR_G0_BOOTSTRAP_ONLY`, and `E_B0_IMAGE_ATTESTATION_POSTBUILD`. |
| `release_authority` | `path`, `schema`, `task`, `network`; values are `.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json`, `hub.qa.g0-bootstrap-founder-release/v1`, `g0-assets-runner`, and `disabled`. |
| `cleanup_contract` | literal `hub.qa.exact-cleanup/v1`. |
| `direct_invocation` | `public_bootstrap=false`, `public_g0_planning_selftest=false`, `exit_code=64`. |

`hub.qa-bootstrap.v1.json` is produced only by Package B, committed only in
the B allowlist, and consumed by B's outer bootstrap, C, C1/D/E review, and
Recovery Package A through the activation mapping. It must not embed dynamic
report digests, image IDs, ingress data, a release, a C receipt, or deletion
state. Runtime facts belong only to C/C1/D/E, so a static B contract cannot be
mistaken for a bootstrap receipt.

It accepts no release record, release path, handoff path, JSON, environment override, or other authority input from its caller. It derives the sole release only from the fixed artifact path `.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json` at committed `HEAD`, requires that `HEAD` equal that artifact's dedicated Package-R commit, and derives `task=g0-assets-runner` and the grammar-valid unique `run` only from its closed raw bytes. It creates exactly `testframework/reports/<task>-<run>/bootstrap/owner-marker.v1.json` with create-new semantics, retains its exact raw bytes and raw SHA-256, and then owns this fixed sequence:

1. verify the committed Package-R repository authority, its exact A0/A1/B/plan commit bindings, expiry, one-shot nonce, and local-only prerequisites;
2. create the owner marker, inspect **all** required local bases, and create the nonce/digest-bound local-base receipt;
3. verify the A1-bound `bootstrap-helper-python` identity against that receipt before any build;
4. build and attest the ephemeral helper from that verified local Python base;
5. copy and verify the founder's fixed offline ingress into the task-owned report subtree, retaining an in-memory digest map and a closed ingress-copy receipt;
6. invoke the fixed Windows-host admission launcher with only the immutable task-owned copied ingress plus runner-held private copy receipt and digest map; the launcher invokes the helper to verify/admit only those copied bytes;
7. remove and receipt the helper before four-target image creation;
8. bind the closed A1 controls, call exactly `Get-HubQaGitObjectSourceTree -Controls $binding.Controls`, and create/verify the one source-tree receipt before any target build;
9. build all four fixed targets from their admitted bytes and local bases, then create the post-build local-image inspection receipt;
10. construct the distinct post-build image-attestation context from that receipt, call the full A0 `Invoke-HubQaBootstrapImageAttestation -BootstrapPrivateContext $ImageAttestationPrivateContext`, and accept its one candidate `testframework/reports/<task-run>/bootstrap/images.lock.v3.json` result;
11. invoke only the fixed `g0-planning-selftest` entrypoint through runner v2;
12. publish closed task receipts and preserve the first non-zero exit;
13. enforce timeout and execute exact owned-resource cleanup in `finally`;
14. emit report and cleanup receipts, retaining failed evidence.

It has no arbitrary command, image, path, environment, retry, install, network, pull, or SUT-target capability. A timeout uses exit `124`; an earlier failure always wins over cleanup failure. Cleanup selects only resources carrying the exact owner marker from the initial owned-resource inventory; foreign resources and final reports are never adopted or deleted.

External ingress must provide, before A1 and bootstrap, every byte and evidence item in the closed dependency bundle plus all locally preloaded pinned base image IDs/digests, SPDX licenses, license-evidence digests, and provenance digests. Missing, mutable, remote, tagged-only, nonlocal, unreleased, or nonphysical input is `BLOCKED_EXTERNAL_INGRESS`. A1 is the only package allowed to bind those exact values into repository data. Bootstrap itself, rather than a caller, creates `testframework/reports/<task-run>/bootstrap/local-base-inspect.v1.json` from closed local Docker inspection before helper build or asset admission. It generates a cryptographically random receipt nonce, writes the raw receipt once with the exact task, run, owner-marker raw SHA-256, nonce, and inspected IDs/digests, then retains its no-BOM raw bytes, raw-byte SHA-256, and nonce only in the separate pre-helper admission context. The fixed helper invocation receives those admission-only values only from the outer launcher and must match them to the raw local-base receipt; it cannot choose a receipt path or substitute JSON. No later post-build image-attestation field may reuse this raw receipt. Caller-provided `LocalBaseInventory`, repository root, candidate path, target-attestation records, receipt path, receipt digest, receipt nonce, offline-ingress path, handoff JSON, asset root, or equivalent JSON are rejected.

### Separate pre-helper admission and post-build attestation contexts

Package B creates two non-interchangeable in-memory ordered hashtables. Neither
is serialized into a CLI, JSON file, environment variable, Docker label, mount,
or report. The only initial owner bytes are the create-new exact raw bytes at
`testframework/reports/<task-run>/bootstrap/owner-marker.v1.json`; every owner
digest in both contexts is the SHA-256 of those same bytes. There is no
alternate owner-marker filename or second owner marker.

1. Before helper admission, `$AdmissionPrivateContext` is passed exactly once as
   the `-BootstrapPrivateContext` argument to accepted A0
   `Invoke-HubQaG0AdmissionFromBootstrapContext`. Its exact A0 field order is
   `Task`, `Run`, `OwnerMarkerSha256`, `PrivateReportRoot`,
   `PrivateIngressRoot`, `PrivateIngressCopyReceipt`,
   `PrivateIngressCopyReceiptSha256`, `PrivateIngressDigestMap`,
   `PrivateIngressDigestMapSha256`, `PrivateIngressOrderedPaths`,
   `PrivateIngressByteCounts`, `PrivateIngressSourceRoot`,
   `PrivateIngressDestinationRoot`, `PrivateOfflineIngressHandoffId`,
   `PrivateOfflineIngressOwnerMarkerSha256`,
   `PrivateDependencyBundleManifest`, `PrivateDependencyBundleManifestSha256`,
   `PrivateSourceCommit`, `PrivateLocalBaseReceipt`,
   `PrivateLocalBaseReceiptSha256`, `PrivateLocalBaseReceiptNonce`,
   `PrivateDockerClientPath`, `PrivateHelperBuildTag`,
   `PrivateHelperImageAttestation`, `PrivateHelperImageAttestationSha256`,
   `PrivateHelperLocalImageId`, `PrivateBootstrapHelperPythonLocalImageId`, and
   `PrivateBootstrapHelperPythonLocalImageIdSha256`. In particular,
   `PrivateLocalBaseReceipt` is exactly the raw bytes at
   `bootstrap/local-base-inspect.v1.json`, and its paired SHA-256 and nonce are
   used only for the admission/helper sequence. This context has no
   `RoleClaim`, `OwnerMarkerRawBytes`, `RunnerInspectionRawBytes`, or
   `RunnerInspectionSha256` fields and must never be passed to the A0 image
   attestor.
2. Only after all four target images have been built and the runner has
   create-new written the exact raw
   `testframework/reports/<task-run>/bootstrap/local-image-inspect.v1.json`, it
   creates `$ImageAttestationPrivateContext` for accepted A0
   `Invoke-HubQaBootstrapImageAttestation`. Its exact field order is `Task`,
   `Run`, `RoleClaim`, `OwnerNonce`, `OwnerMarkerRawBytes`,
   `OwnerMarkerSha256`, `RunnerInspectionRawBytes`, and
   `RunnerInspectionSha256`. `OwnerMarkerRawBytes` and `OwnerMarkerSha256` are
   the exact bytes/digest of the fixed `bootstrap/owner-marker.v1.json` file;
   `RunnerInspectionRawBytes` and `RunnerInspectionSha256` are exclusively the
   exact bytes/digest of the post-build `bootstrap/local-image-inspect.v1.json`
   file. It has no `PrivateLocalBaseReceipt*` field and cannot be passed to the
   admission launcher. The A0 source attestor reads that same fixed local-image
   path and rejects any byte/digest mismatch before candidate publication.

### Internal bootstrap role capability

`hub.qa-bootstrap/v1` is an internal process-routing capability, not a public
credential. The sole valid `role_claim` value is the literal string
`hub.qa-bootstrap/v1`. It is transported only in the in-memory ordered
PowerShell hashtable `BootstrapPrivateContext`; it is never accepted from a
CLI argument, `stdin`, JSON file, environment variable, report, Docker label,
mount, or container command. Only the post-build image-attestation context,
after the local-image receipt is create-new written, carries this role and
constructs exactly these eight context keys in this order:

```text
Task, Run, RoleClaim, OwnerNonce, OwnerMarkerRawBytes, OwnerMarkerSha256,
RunnerInspectionRawBytes, RunnerInspectionSha256
```

`Task` is the fixed task, `Run` uses the global grammar, `RoleClaim` is the
literal above, and `OwnerNonce` is exactly 32 lowercase hexadecimal characters
generated once from the platform cryptographic RNG after Package-R authority
validation. `OwnerMarkerRawBytes` are the create-new raw bytes of the fixed
`bootstrap/owner-marker.v1.json`; `OwnerMarkerSha256` is their raw-byte
SHA-256. `RunnerInspectionRawBytes` and `RunnerInspectionSha256` are the
post-build `bootstrap/local-image-inspect.v1.json` raw bytes/digest, never the
earlier local-base receipt. Consumers compare every field byte-for-byte against
the marker and image receipt before use; a context with an extra, missing,
reordered, wrong-type, or mismatched field is rejected.

`invoke-qa-gate-v2.ps1` may expose only the normal closed v2 request
`target`, `task`, `run`; it never accepts `BootstrapPrivateContext`, a role
claim, an owner nonce, or report-path override. The bootstrap launcher calls a
non-exported `Invoke-HubQaBootstrapPlanningSelftestV2` closure in the same
PowerShell process; it receives the private context, derives target
`g0-planning-selftest` and the report root itself, and has no target/task/run
arguments. `run-task-v2.ps1` as a file has no executable parameter set:
`pwsh -File testframework/runner/run-task-v2.ps1`, a direct
`g0-planning-selftest` request, or any serialized/private-context attempt
fails before Docker use with `E_B0_DIRECT_INVOCATION_FORBIDDEN` and exit `64`.
The Package-R one-shot nonce, create-new owner marker, exact report-root
absence, and the single closure invocation make a duplicate/continuation/retry
attempt fail as `E_B0_ONE_SHOT_REJECTED`; a new transaction needs a new
founder release and run.

These bindings prove only that the accepted code observed byte-consistent
repository/report inputs within one local transaction. They are not a MAC,
signature, authenticated identity, proof of a human/founder/model/agent, or
proof that a hostile equal-trust host process did not forge in-memory state.
The independent Package-R/D reviews and committed raw artifacts remain the
authority boundary; no role claim is presented as cryptographic proof.

## Fixed source deliverables

Package A0 is a static source contract and may proceed without external ingress. It creates no dependency binding data and must not invent byte, digest, license, image-ID, or provenance values. Package A1 is the later binding-data-only package; it is blocked until founder-approved offline ingress supplies the exact values. Existing historical `source-policy.v1.json` and `build-manifest.v2.json` remain byte-for-byte unchanged.

| Concern | New or changed files |
|---|---|
| A0 dependency-admission source contract | `testframework/docker/dependency-bundle.schema.v1.json`, `testframework/docker/admission/verify_local_dependency_bundle.py`, `testframework/docker/admission/admit_local_dependency_assets.py`, `testframework/docker/Dockerfile.g0-admission-helper`, `testframework/docker/entrypoints/g0-admission-helper.sh`, `testframework/docker/admit-local-dependency-assets.ps1`, `testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1` |
| A0 source/build-attestation source contract | `testframework/docker/source-image-attestor.ps1`, `testframework/selftests/runner/g0-image-contract.ps1` |
| A0 four image targets and bootstrap control | `testframework/docker/Dockerfile.backend-unit`, `testframework/docker/Dockerfile.frontend-unit`, `testframework/docker/Dockerfile.backend-openapi`, `testframework/docker/Dockerfile.g0-planning-selftest`, `testframework/docker/entrypoints/backend-unit.sh`, `testframework/docker/entrypoints/frontend-unit.sh`, `testframework/docker/entrypoints/backend-openapi.sh`, `testframework/docker/entrypoints/g0-planning-selftest.sh`, `testframework/docker/runner-fixtures/g0-planning-selftest.py`, `testframework/docker/vitest.frontend-unit.config.ts`, `testframework/docker/compose.g0-bootstrap.yaml`, `testframework/docker/bootstrap-control.v1.json` |
| A1 binding data only | `testframework/docker/dependency-bundle.manifest.v1.json`, `testframework/docker/source-policy.v2.json`, `testframework/docker/build-manifest.v3.json` |
| Bootstrap and v2 runner | `testframework/runner/contracts/hub.qa-bootstrap.v1.json`, `testframework/runner/invoke-qa-bootstrap-v1.ps1`, `testframework/runner/invoke-qa-gate-v2.ps1`, `testframework/runner/run-task-v2.ps1`, `testframework/runner/new-repository-snapshot.ps1`, `testframework/runner/new-run-envelope.ps1`, `testframework/runner/write-g0-activation.ps1`, `testframework/schemas/repository-snapshot.schema.v1.json`, `testframework/schemas/run-envelope.schema.v1.json`, `testframework/selftests/runner/g0-runner-v2-contract.ps1`, `testframework/runner/README.md`, `.cursor/skills/qa-gate/SKILL.md` |

The four immutable targets are exactly `backend-unit`, `frontend-unit`, `backend-openapi`, and `g0-planning-selftest`. The first three are never executed by G0. `g0-planning-selftest` is an image-baked runner-fixture target, not an end-user or SUT QA target. Its Dockerfile and the ephemeral helper Dockerfile both use the A1-bound `bootstrap-helper-python` base. The current local inventory observation is `python:3.12-slim@sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36`, locally inspected as image ID `sha256:54a0b2beae90fe14c2131f1e880e1388c9df77390bc260d7317e589765e064e8`; it is neither a hard-coded production assertion nor pull authority. A1 must bind the exact approved identity/provenance and C must verify it through the nonce/digest-bound local receipt.

### Closed founder offline-ingress handoff

Offline ingress is a founder-approved, local-only handoff; it is not a caller argument, a discovery scan, a Docker build context, or repository source. The committed Package-R artifact is the only bootstrap release record and binds the already-approved handoff through the closed fields defined below, including `offline_ingress_handoff_id`, `offline_ingress_owner_marker_sha256`, `dependency_bundle_manifest_sha256`, and `source_commit`. `offline_ingress_handoff_id` is grammar-valid lowercase ASCII `[a-z0-9][a-z0-9-]{7,63}` and selects exactly one physical root derived from the canonical repository root: `.qa-local-ingress/g0-assets/<offline_ingress_handoff_id>`. The artifact must name `source_commit=7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`; it cannot name a later plan, package, evidence, or activation commit. The bootstrap derives this path itself after validating the committed artifact; no API, environment variable, CLI switch, JSON, or runtime caller can provide or override any portion of it.

That root is an untracked local ingress zone and is never committed, added to a source image build context, mounted as source, copied by a Dockerfile, retained by a task, or deleted by task cleanup. Its exact regular-file tree is closed: `handoff-owner.v1.json`, `dependency-bundle.manifest.v1.json`, `assets/<relative_path>` for every ordered `assets[].relative_path` in the manifest, `evidence/licenses/assets/<asset_id>.txt`, `evidence/provenance/assets/<asset_id>.json`, `evidence/licenses/bases/<base_id>.txt`, and `evidence/provenance/bases/<base_id>.json`. No directory, file, alternate stream, reparse point, symlink, junction, hidden foreign child, or path outside this set is accepted. `handoff-owner.v1.json` is UTF-8/no-BOM, create-new, and has exactly `schema`, `handoff_id`, `source_commit`, `dependency_bundle_manifest_sha256`, `owner`, and `released_at_utc`; it must name the same handoff ID and source baseline as the founder release and its raw SHA-256 must equal `offline_ingress_owner_marker_sha256`. The raw manifest bytes must equal the release's `dependency_bundle_manifest_sha256` and later be byte-identical to the only A1 repository file `testframework/docker/dependency-bundle.manifest.v1.json`. Each asset and evidence file must be a physical regular file below this root, have no ADS/reparse traversal, and match its A1-bound byte length and SHA-256 or the associated manifest hash (`license_evidence_sha256` / `provenance_sha256`) exactly. Absent or mismatched handoff root, marker, manifest, child, raw bytes, digest, baseline, or physical-path proof is `BLOCKED_EXTERNAL_INGRESS` before helper construction or admission.

During C, after the base receipt and helper attestation but before admission, the bootstrap alone copies this closed handoff with create-new semantics into `testframework/reports/<task-run>/ingress/dependency-bundle`. It preserves the handoff layout below that fixed destination, retains a runner-private in-memory ordered digest map for every copied raw file, and writes exactly one create-new owner-bound `testframework/reports/<task-run>/bootstrap/ingress-copy-receipt.v1.json`. That receipt names task, run, owner marker, founder handoff ID, source baseline, source owner-marker and manifest digests, fixed source/destination roots, ordered relative paths, raw byte counts, raw SHA-256 values, and the digest-map SHA-256. The fixed host admission launcher receives only the immutable task-owned copied ingress and runner-held private copy receipt/digest map; it consumes them to compare copied bytes, receipt, A1 manifest bytes, and base receipt before invoking the helper. It cannot read, write, copy, mount, adopt, enumerate, or otherwise access the founder handoff root or ingress; it cannot accept a caller-selected path or JSON or rebuild the digest map from untrusted input. Bootstrap/B alone creates the copy and receipt. Cleanup removes only the task-owned copied ingress subtree and records it in `cleanup.json`; it never removes the untracked founder handoff root.

### Package R — committed founder bootstrap authority

**Files:** create only `.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json`. Package R starts only after independent `ACCEPT` verdicts and already-existing commits for A0, A1, and B, and only after the founder explicitly approves this exact local-only one-shot transaction. It is committed separately, after that approval, with no runtime operation before or during its creation:

```powershell
git commit --only .planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json -m "docs(test-rebase): release one G0 bootstrap transaction"
```

The artifact uses the global ordered canonical JSON bytes, has schema `hub.qa.g0-bootstrap-founder-release/v1`, and has exactly these top-level keys in order: `schema`, `release_id`, `g0_plan_path`, `g0_plan_commit`, `a0_commit`, `a1_commit`, `b_commit`, `task`, `run`, `local_only`, `network`, `offline_ingress_handoff_id`, `offline_ingress_owner_marker_sha256`, `dependency_bundle_manifest_sha256`, `source_commit`, `expires_at_utc`, `one_shot_nonce`, `founder_approval`, `created_at_utc`. `task` is exactly `g0-assets-runner`; `run` matches exactly `^[a-z0-9][a-z0-9-]{7,63}$` and is unique; `local_only` is JSON `true`; `network` is exactly `"disabled"`; `source_commit` is exactly `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`; every named commit is an existing full lowercase 40-hex commit and must be an ancestor of the dedicated Package-R commit; `g0_plan_path` is this plan's repository-relative path; `offline_ingress_handoff_id` and both SHA-256 fields obey the closed handoff contract; `expires_at_utc` is a future UTC instant at Package-R review and C must reject expiry; `one_shot_nonce` is a unique 32-lowercase-hex value never accepted by an earlier successful or failed bootstrap; and `founder_approval` is exactly `EXPLICIT_LOCAL_ONLY_ONE_SHOT`.

This artifact is deliberately not a cryptographic signature, identity proof, or substitute for founder authentication. Its authority is the explicit founder approval followed by this separately reviewed repository commit. C must read only the fixed raw artifact at committed `HEAD`, require `HEAD` to be its dedicated commit, and verify raw bytes, closed schema, exact plan/A0/A1/B commit allowlist, release expiry, nonce non-reuse, and all handoff bindings before report creation. It must reject a caller-supplied artifact/path/environment/JSON, an artifact from any other commit, a dirty replacement, a detached historical artifact, or a release whose named commit is not already accepted. Until this artifact exists and is independently accepted, `G0_C=BLOCKED`, `BOOTSTRAP=NOT_RUN`; it authorizes neither network nor any other ingress.

**Independent acceptance:** Terra/high independently verifies the founder-approved intent against the exact raw artifact bytes and schema; `git rev-parse HEAD` and the artifact-containing dedicated commit; A0/A1/B/plan commit existence, ancestry, accepted verdicts, and literal allowlists; fixed source baseline; release task/run/expiry/nonce uniqueness; local-only and no-network fields; and fixed handoff ID/marker/manifest hashes. It rejects ambiguous founder approval, altered raw bytes, an extra file in Package R, a self-reference, a later activation/evidence/C1 commit substituted for A0/A1/B, or any claim of cryptographic identity. A HIGH finding blocks C.

### Dependency admission contract

Package A0 freezes the schema and verifier/admission implementation only. It must test only structural fixtures and explicit missing/unbound cases, state `NOT_RUN`, and write no `dependency-bundle.manifest.v1.json`. Package A1 later creates the manifest only from approved ingress; it changes no A0 code.

`dependency-bundle.manifest.v1.json` is strict UTF-8 without BOM and schema `hub.qa.dependency-bundle/v1`. Its closed top-level keys, in this order, are `schema`, `bundle_id`, `source_commit`, `platform`, `assets`, `bases`, and `bundle_sha256`. `source_commit` is exactly `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`: it binds the recovered source tree and must never be changed to a later control/repository commit. `platform` is exactly `{"os":"linux","architecture":"amd64","variant":null}`.

Each `assets` item has exactly `asset_id`, `targets`, `relative_path`, `sha256`, `bytes`, `license_spdx`, `license_evidence_sha256`, and `provenance_sha256`. The only asset identities are `go-module-cache`, `go-race-toolchain`, `frontend-node-modules`, and `swag-binary`; their closed target mappings are respectively `["backend-unit","backend-openapi"]`, `["backend-unit"]`, `["frontend-unit"]`, and `["backend-openapi"]`. There is no `powershell-runtime` asset, Linux PowerShell archive, package, interpreter, `PATH` lookup, `COPY`, extraction, or install in G0. Each `bases` item has exactly `base_id`, `consumers`, `local_image_id`, `manifest_digest`, `license_spdx`, `license_evidence_sha256`, and `provenance_sha256`. Its closed `consumers` values may name one or more of the four immutable targets and the sole non-target consumer `g0-admission-helper`; `g0-admission-helper` is valid only in `bases`, never in `assets`, Compose, bootstrap control, a candidate lock, or activation. The required helper base is `base_id=bootstrap-helper-python` with consumers exactly `["g0-admission-helper","g0-planning-selftest"]`; A1 binds its exact local ID/digest/provenance. This replaces and retires the legacy permanent `Assert-BaseBinding` branch: no A0 implementation may preserve a separate hard-coded base assertion. Every base is valid only when its `local_image_id` is already present locally and agrees with its pinned digest in the runner-produced `local-base-inspect.v1.json` receipt. All digest values use `sha256:` plus 64 lowercase hexadecimal characters.

`bundle_sha256` is non-self-referential: it is `sha256:` plus the SHA-256 of the UTF-8/no-BOM bytes of the canonical projection of the manifest **without** `bundle_sha256`. Canonical projection preserves the declared key order at every closed object, preserves array order, emits no insignificant whitespace, and uses JSON literals only (`null`, decimal integer, JSON string escaping); it may not hash the received raw manifest bytes or a representation containing the field being checked. The verifier computes this projection itself and rejects mismatch as `E_BUNDLE_SHA256_MISMATCH`.

`license_spdx` is an offline SPDX-2.3 simple-license-ID field, not arbitrary text and not a compound expression. It accepts only `0BSD`, `Apache-2.0`, `Artistic-2.0`, `BSD-2-Clause`, `BSD-3-Clause`, `CC0-1.0`, `EPL-2.0`, `ISC`, `MIT`, `MPL-2.0`, `PSF-2.0`, `Python-2.0`, `Unicode-DFS-2016`, or `Zlib`; it rejects `LicenseRef-*`, `NOASSERTION`, `NONE`, `AND`, `OR`, `WITH`, whitespace-padded values, and every value outside that closed set. License evidence and provenance remain independently required SHA-256 evidence, never a substitution for this check.

`verify_local_dependency_bundle.py` and `admit_local_dependency_assets.py` are the only bundle parser and admission implementation. They use only the Python standard library and have no caller-selectable manifest, asset root, source commit, base receipt, repository root, Docker inspection, or target-record input. They resolve their fixed in-container layout from their own physical file paths: controls at `/opt/hub/controls/dependency-bundle.manifest.v1.json`, ingress at `/work/ingress/dependency-bundle`, and raw base receipt at `/work/bootstrap/local-base-inspect.v1.json`. The outer bootstrap creates that task-owned `/work` mapping; it never mounts the source tree. The public verifier checks exact keys, duplicate-aware JSON, strict decoding, path containment below `assets/`, byte counts, SHA-256, canonical bundle hash, closed SPDX policy, licenses, provenance, reparse/symlink escape, the fixed source commit, and every A1 base before derived output. It returns exactly the JSON object `{ "verdict", "manifest_sha256", "assets", "bases", "reason_code" }`; allowed reason codes are `E_MANIFEST_UTF8`, `E_MANIFEST_SCHEMA`, `E_DUPLICATE_ASSET_ID`, `E_ASSET_PATH_ESCAPE`, `E_ASSET_MISSING`, `E_ASSET_BYTES_MISMATCH`, `E_ASSET_SHA256_MISMATCH`, `E_BUNDLE_SHA256_MISMATCH`, `E_LICENSE_SPDX_INVALID`, `E_LICENSE_EVIDENCE_MISSING`, `E_PROVENANCE_EVIDENCE_MISSING`, `E_BASE_MISSING`, `E_BASE_ID_MISMATCH`, `E_SOURCE_COMMIT_MISMATCH`, and `E_RECEIPT_BINDING_INVALID`.

`admit-local-dependency-assets.ps1` is a Windows-host, bootstrap-private admission consumer only. It is never copied into, executed by, or structurally substituted for the Linux helper or `g0-planning-selftest`; it does not parse the bundle. It receives only the immutable task-owned ingress copy and runner-held private ingress-copy receipt/digest map created by bootstrap/B, verifies/consumes those held bindings, invokes the fixed `g0-admission-helper` image, and consumes its raw receipt. It cannot read, write, copy, mount, adopt, enumerate, or otherwise access the founder handoff root or ingress; bootstrap/B alone performs the create-new copy. It has no public CLI mode, caller-provided path/JSON, direct test mode, arbitrary Docker command, or standalone runtime claim. The behavior of that host PowerShell launcher is evidenced only by the actual founder-released Package C transaction.

## Four-target image and lock contract

Package A0 freezes `source-image-attestor.ps1`, four Dockerfiles, four entrypoints, dedicated Vitest config, Compose topology, bootstrap control, and the two failing-first static selftest sources. It creates no `source-policy.v2.json` or `build-manifest.v3.json`; those A1 data files bind the exact approved ingress and source values only after ingress is available. A0 source contracts must cover unbound/missing binding data as `NOT_RUN`, never manufacture a plausible digest or image identity.

`source-policy.v2.json` is `hub.qa.local-source-policy/v2`; `build-manifest.v3.json` is `hub.qa.local-images-build-manifest/v3`. They bind the committed recovery source revision, the mandatory sorted source-tree digest, four target definitions, entrypoint digests, the dedicated Vitest config digest, bundle digest, and no network or registry capability. `source-policy.v2.json` contains the closed ordered allowlist of every source `path` and its Git `blob_sha1` expected at source commit `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`; it admits only mode `100644` regular blobs. `build-manifest.v3.json` repeats the exact source commit and source-tree digest and binds the raw SHA-256 of `source-policy.v2.json`. Neither file may infer source entries from the worktree, index, glob, Docker context, or a later control commit.

`source-image-attestor.ps1` has no caller-selectable `RepositoryRoot`, `CandidatePath`, `TargetAttestations`, record collection, source-policy path, build-manifest path, or image-inspection input. It resolves the canonical repository root only from its own physical `PSScriptRoot`, verifies every ancestor and consumed path is non-reparse and contained, reads the fixed raw A1 controls `testframework/docker/source-policy.v2.json`, `testframework/docker/build-manifest.v3.json`, and `testframework/docker/dependency-bundle.manifest.v1.json`, and derives the candidate lock path only as `testframework/reports/<task-run>/bootstrap/images.lock.v3.json` from the internally validated post-build image-attestation context. It accepts only the runner-produced, closed projection of `docker image inspect` rooted at that same task report and recomputes the four target records, entrypoint hashes, and Docker `Cmd` hashes itself using JCS canonical bytes plus fixed Dockerfile/entrypoint/control bytes. It consumes, but never creates or adopts, the runner-created exact raw `bootstrap/owner-marker.v1.json` file, then creates only its task-owned reservation and staging paths with create-new semantics before candidate publication.

Before any build or candidate publication it must verify the source tree directly from Git object data at exact commit `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`, never from the worktree or index (the 241 staged legacy deletions are therefore irrelevant). It enumerates only the closed `source-policy.v2.json` allowlist and requires a one-to-one match: every path occurs exactly once, has Git mode `100644`, resolves to the declared Git blob SHA-1, and has raw content SHA-256 equal to the independently read blob bytes. It rejects every symlink (`120000`), submodule/gitlink (`160000`), tree, nonregular entry, duplicate/collision after UTF-8 byte encoding, unlisted entry, missing entry, or blob/digest disagreement. It canonicalizes each admitted source blob as raw UTF-8/no-BOM line bytes `path<TAB>100644<TAB>blob_sha1<TAB>sha256(raw)<LF>`, sorts those complete lines by the UTF-8 byte order of `path`, concatenates them with no separators or trailing bytes beyond each required LF, and computes SHA-256. That value must equal both `source-policy.v2.json.source_tree_sha256` and `build-manifest.v3.json.source_tree_sha256`; the build manifest's source-policy raw SHA-256 and source commit must also equal the fixed raw policy values. A missing/invalid closed source entry is `E_SOURCE_TREE_INVALID`; a computed digest or cross-control binding mismatch is `E_SOURCE_TREE_DIGEST_MISMATCH`; either gate is `G0_C=FAIL/BLOCKED` before build or candidate. It rejects a missing, duplicate, nonlocal, foreign-root, reparse, mismatched-source-baseline, or caller-forged record before writing a candidate. The source baseline comparison is exactly `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`; later control/repository commits are recorded separately and cannot satisfy it.

The frontend target must use only `testframework/docker/vitest.frontend-unit.config.ts`, whose include is exactly `/opt/hub/source/frontend/tests/**/*.test.{ts,tsx}`. It has no exclusion that can hide a failure. The three SUT entrypoints are fixed source contracts only: backend unit runs the declared isolated `backend/tests/...` sequence, frontend unit runs its fixed TypeScript/Biome/Vite/Vitest sequence, and OpenAPI compares two independent fixed generations. `build-manifest.v3.json` separately binds (a) each image's literal Docker `Entrypoint` and `Cmd`, whose JCS hash is the target `command_sha256`, and (b) its target-action contract, which may contain human-readable purpose text but is not a Docker `Cmd` and is never hashed as one. This distinction is mandatory for `backend-openapi`. Neither representation may reference legacy test roots, `testframework/stages`, a network client, or installation command.

`Dockerfile.g0-admission-helper` and `Dockerfile.g0-planning-selftest` are offline Python-stdlib images on the same A1-bound `bootstrap-helper-python` local base. The helper is built first from its closed A0 context, before dependency admission; it copies only `dependency-bundle.schema.v1.json`, `admission/verify_local_dependency_bundle.py`, `admission/admit_local_dependency_assets.py`, `entrypoints/g0-admission-helper.sh`, and its one A1 control `controls/dependency-bundle.manifest.v1.json` to `/opt/hub/controls/dependency-bundle.manifest.v1.json`. Package C stages that control, but never an asset, at `testframework/reports/<task-run>/bootstrap/build-context/g0-admission-helper/controls/dependency-bundle.manifest.v1.json` with create-new, owner-marker, physical-path, and raw-byte checks before the build. It has no source mount, network, pull, install, package manager, build argument, Compose service, bootstrap-control entry, candidate-lock record, or activation record. It executes only its fixed entrypoint against the task-owned report mount.

Before dependency admission, Package C uses the runner-private in-memory bootstrap context to inspect all locally required bases and create the local-base receipt, verifies the A1 `bootstrap-helper-python` identity against that receipt, builds this helper offline, and inspects the resulting local helper image. It creates exactly two create-new owner-bound raw receipts: `testframework/reports/<task-run>/bootstrap/g0-admission-helper.local-image-attestation.v1.json` and `testframework/reports/<task-run>/bootstrap/g0-admission-helper.cleanup.v1.json`. The first binds the helper image ID, its ID digest, the A1 `bootstrap-helper-python` base ID/digest, the helper Dockerfile and entrypoint digests, task, run, owner marker, and the bootstrap receipt nonce. The second records only the exact initially owned helper resources and their removal result. Only after the helper attestation may C make the closed create-new ingress copy and call the host launcher/helper; helper cleanup follows admission and precedes four-target construction. The helper is an ephemeral bootstrap implementation detail, not a fifth QA target; it is removed before four-target image attestation and is excluded from `images.lock.v3.json`. A missing local base, failed helper-base verification or attestation, missing/mismatched ingress-copy receipt, non-empty unexpected container state, nonzero helper result, or ambiguous cleanup is `G0_C=FAIL/BLOCKED`.

`Dockerfile.g0-planning-selftest` has a separate closed A0 copy set: `dependency-bundle.schema.v1.json`, `admission/verify_local_dependency_bundle.py`, `admission/admit_local_dependency_assets.py`, `entrypoints/g0-planning-selftest.sh`, and `runner-fixtures/g0-planning-selftest.py`. Its only A1 build-context inputs are the three immutable controls at `controls/dependency-bundle.manifest.v1.json`, `controls/source-policy.v2.json`, and `controls/build-manifest.v3.json`; C alone stages them under `testframework/reports/<task-run>/bootstrap/build-context/g0-planning-selftest/` with create-new, owner-marker, physical-path, and raw-byte checks before the first build. It never copies or executes `admit-local-dependency-assets.ps1`, `source-image-attestor.ps1`, any PowerShell archive, or any Package-B PowerShell source. No Package-B file is copied into this image, so A0 has no A0/B build cycle.

The planning fixture imports and calls the same two Python admission modules directly with fixed test-owned fixture roots. It proves valid, duplicate-ID, SHA mismatch, path-escape, missing-asset, and unbound-manifest cases against the public Python verifier and verifies that the admission wrapper publishes no success receipt on a rejected fixture. It does not test Windows-host launcher behavior or future Package-B PowerShell behavior; those are static-reviewed in A0/B and evidenced only in the founder-released Package C transaction. The selftest image runs only through the internal G0 v2 profile during C, never as a public target. It installs no packages and does not invoke backend, frontend, OpenAPI, application Python suites, Playwright, legacy tests, or application source commands. Before C all fixture evidence remains `NOT_READY/NOT_RUN`.

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

`testframework/schemas/run-envelope.schema.v1.json`, `$id` `hub.qa.run-envelope/v1`, is a closed schema with top-level keys `schema`, `producer`, `envelope_id`, `contract`, `target`, `task`, `run`, `owner_nonce`, `repository_snapshot_sha256`, `report_root`, `limits`, `cleanup_contract`, `created_at_utc`. `contract` is `hub.qa-runner/v2`; target is one of the four fixed G0 targets or the schema-only `recovery-successor-authority` target. The latter has no Package-B/C runner route and remains `NOT_READY` until its separately accepted recovery package; it exists only so Recovery consumes this one schema rather than forking it. `limits` has exactly `command_seconds`, `target_seconds`, `inspect_seconds`, `outer_seconds`, `cleanup_seconds` and values `180`, `780`, `30`, `810`, `30`; `cleanup_contract` is `hub.qa.exact-cleanup/v1`. Its only producer is `testframework/runner/new-run-envelope.ps1` and the activation record binds that path and digest.

### Closed snapshot/envelope value and link matrix

The schema files are executable closed contracts, not descriptive examples.
`repository-snapshot.schema.v1.json` itself has top-level ordered keys `$schema`,
`$id`, `title`, `type`, `additionalProperties`, `required`, `properties` and
requires `$schema=https://json-schema.org/draft/2020-12/schema`,
`$id=hub.qa.repository-snapshot/v1`, `type=object`, and
`additionalProperties=false`. `run-envelope.schema.v1.json` has the same
schema-document key order and requires
`$schema=https://json-schema.org/draft/2020-12/schema`,
`$id=hub.qa.run-envelope/v1`. Their
runtime payloads use the ordered keys already listed above with no aliases.
The only payload producers are respectively
`testframework/runner/new-repository-snapshot.ps1` with
`producer=hub.qa.repository-snapshot-producer/v1`, and
`testframework/runner/new-run-envelope.ps1` with
`producer=hub.qa.run-envelope-producer/v1`; their paths and raw-byte SHA-256
are later copied to the activation record. Bootstrap, C1, D, E, and Recovery
Package A are consumers. Recovery creates no parallel schema or producer and
may only use the `captured_for` kinds/causality already fixed in this plan and
the recovery plan.

| Payload / field group | Exact value, type, and nullability | Mandatory link rule |
|---|---|---|
| Both payloads | `schema` is the matching `$id`; `task` and `run` use the global grammars; `created_at_utc` is a UTC RFC-3339 instant with `Z`; strings/integers/booleans obey the ordered-byte rules. G0 additionally requires `task=g0-assets-runner`. | G0 values exactly match Package R, the owner marker, local-base receipt, private context, and deterministic report root. Recovery values must match its immutable genesis and use no different grammar. |
| Snapshot IDs and role | `snapshot_id` matches `^[a-z][a-z0-9-]{2,31}-snapshot-[a-z0-9][a-z0-9-]{7,63}-[0-9]{3}-[A-Z_]+$`; `envelope_id` matches `^[a-z][a-z0-9-]{2,31}-envelope-[a-z0-9][a-z0-9-]{7,63}$`; `role_claim` is a nonempty ASCII string of at most 128 characters; `self_attested` is JSON `true` for `G0_BOOTSTRAP`. | G0 must use `g0-assets-runner-snapshot-<run>-<ordinal>-<kind>`, `g0-assets-runner-envelope-<run>`, and literal role `hub.qa-bootstrap/v1`. A recovery snapshot preserves the same field names and can use only the exact role/self-attestation required by its frozen genesis. |
| `captured_for` | `kind` is one of the eight literals in `Shared snapshot and envelope contracts`; `ordinal` is a non-negative base-10 integer; every path/digest/commit member is either all `null` in its declared group or a valid typed value | The all-null rules and the per-kind predecessor/pre-event matrix in that section are authoritative. `G0_BOOTSTRAP` is exactly ordinal `0`, task-owned `run.json` path/digest, `subject_commit_sha=null`, and every predecessor/pre-event field `null`. |
| `source_revision`, `head` | `source_revision.commit` is exactly `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`; `source_revision.tree`, `head.commit`, and `head.tree` are lower 40-hex Git values. `head.event` has exactly `event_id`, `event_type`, `captured_at_utc`; all three are `null` for `G0_BOOTSTRAP`. | Recovery may make `head.event` all non-null only for its already-committed event; a partial null triplet is invalid. Later control commits never replace `source_revision.commit`. |
| `index`, `worktree`, `commits`, `ancestry` | Every listed child object is closed as declared in `Shared snapshot and envelope contracts`. `count` equals array length. `sha256` is the raw ordered canonical projection digest defined by that producer. `paths` and `changed_paths` are ordered lexicographically by UTF-8 path bytes. | A record cannot infer a live worktree/index when consumed by Recovery; its raw snapshot bytes and bound digest are the evidence. |
| `control` | All seven named digest fields are typed SHA-256 strings except `images_lock_sha256`, which is JSON `null` only in the pre-C1 `G0_BOOTSTRAP` snapshot and non-null in a post-C1/recovery snapshot. | Non-null control values equal the exact raw bytes of the named committed source controls. A C1 lock digest cannot be invented from the candidate report lock. |
| `evidence`, `lifecycle` | All four evidence fields are non-null valid report-relative paths/digests for a successfully captured snapshot. `lifecycle` has exactly `state`, `started_at_utc`, `finished_at_utc`; successful state is `CAPTURED` and both times are non-null UTC instants, failed state is `FAILED` and blocks progression. | Evidence paths are contained in the deterministic report root and their digests bind raw bytes. Retained reports are never cleanup targets. |
| Envelope identity | G0 `envelope_id` is exactly `g0-assets-runner-envelope-<run>`; G0 `target` is one fixed target. A recovery envelope can use only `recovery-successor-authority`. `owner_nonce` is exactly the private context nonce for G0 and the genesis-bound nonce for recovery; `report_root` is its deterministic root; `cleanup_contract` is `hub.qa.exact-cleanup/v1`. | The envelope is written once after its snapshot is available, has no caller-selected report path, and links the snapshot raw-byte digest exactly. |
| Envelope limits | `limits` is the exact five-integer object declared immediately before this matrix; no null limit, alias, unit conversion, or extra member is allowed. | `outer_seconds=810` is greater than every nested budget plus fixed cleanup; `timeout` terminates with `124` and preserves the first earlier non-zero exit. |

The G0 bootstrap creates exactly one `G0_BOOTSTRAP` snapshot and one matching
run envelope after the ordered C receipts exist and before planning-selftest
publication; the snapshot is linked in `run.json`, and the envelope digest is
linked back in the snapshot. C1/D/E only consume those immutable raw bytes.
They must reject `E_B0_SNAPSHOT_ENVELOPE_LINK` on a schema/version/producer
drift, wrong task/run/role/nonce/report root, noncanonical bytes, forbidden
null, or either direction of the raw-byte digest link.

## Activation record contract

`.planning/audits/test-rebase/g0-activation.v1.json` uses schema `hub.qa.g0-activation/v1`, is written only after bootstrap and an independent post-run evidence acceptance commit, and has exactly these top-level fields in this order:

`schema`, `activation_id`, `g0_plan_path`, `required_plan_commit`, `founder_release_path`, `founder_release_sha256`, `founder_release_commit`, `repository_commit`, `bootstrap_contract`, `bootstrap_contract_path`, `bootstrap_contract_sha256`, `runner_contract`, `runner_contract_path`, `runner_contract_sha256`, `dependency_bundle_manifest_path`, `dependency_bundle_manifest_sha256`, `source_policy_path`, `source_policy_sha256`, `build_manifest_path`, `build_manifest_sha256`, `images_lock_path`, `images_lock_sha256`, `target_entrypoints`, `planning_selftest`, `repository_snapshot_schema_path`, `repository_snapshot_schema_sha256`, `run_envelope_schema_path`, `run_envelope_schema_sha256`, `repository_snapshot_producer_path`, `repository_snapshot_producer_sha256`, `run_envelope_producer_path`, `run_envelope_producer_sha256`, `report_cleanup_contract`, `acceptance_verdict`, `acceptance_commit`, `created_at_utc`.

Path/digest semantics are exact: each `*_path` is a repository-relative regular file path; its companion `*_sha256` is the SHA-256 of that file's no-BOM raw bytes from `repository_commit`, except `g0_plan_path` which is bound by `required_plan_commit`, and `founder_release_path`/`founder_release_sha256` which are bound by `founder_release_commit`. `founder_release_path` is exactly `.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json`; `founder_release_commit` is the earlier dedicated Package-R commit and must equal the authority commit used by C. `bootstrap_contract` is `hub.qa-bootstrap/v1`; `runner_contract` is `hub.qa-runner/v2`; `report_cleanup_contract` is `hub.qa.exact-cleanup/v1`; and `acceptance_verdict` is exactly `ACCEPT`.

`target_entrypoints` is a closed object with exactly the four fixed target names. Each child has exactly `local_image_id`, `local_image_id_sha256`, `entrypoint_path`, `entrypoint_sha256`, `command_sha256`, `compose_service`, and `runtime_status`; every field must exactly equal the corresponding Package C1 repository-lock child and the bootstrap-control child. `planning_selftest` has exactly `target`, `task`, `run`, `run_receipt_sha256`, `cleanup_receipt_sha256`, `snapshot_sha256`, `envelope_sha256`, and `verdict`; target is `g0-planning-selftest`, task is `g0-assets-runner`, and verdict is `ACCEPT`.

`acceptance_commit` is an existing full 40-lowercase-hex independent evidence acceptance commit. It cannot equal the later activation-record commit. `repository_commit` is exactly the already committed Package C1 lock commit; it is the baseline for all later evidence acceptance and activation comparisons. `required_plan_commit` identifies the earlier plan commit. `Write-HubQaG0Activation` uses create-new semantics, validates all values, and cannot create a second activation; it must reject a missing or byte-mismatched Package-R authority even though it makes no claim that the artifact cryptographically identifies its founder.

### Package-D evidence-acceptance record contract

Package D may create only
`.planning/audits/test-rebase/g0-bootstrap-evidence-acceptance.v1.json`. It is
strict UTF-8/no-BOM duplicate-aware parsed canonical JSON under the global
ordered-byte rules, uses schema `hub.qa.g0-bootstrap-evidence-acceptance/v1`,
and has exactly these top-level keys in this order:

```text
schema, acceptance_id, g0_plan_path, g0_plan_commit, a0_commit, a1_commit,
b_commit, founder_release, repository_commit, task, run, c_receipts, c1_lock,
repository_snapshot, run_envelope, acceptance_verdict, accepted_at_utc
```

`acceptance_id` is exactly `g0-bootstrap-evidence-acceptance-<run>`;
`g0_plan_path` is this plan's repository-relative path; `task` is exactly
`g0-assets-runner`; `run` is the same Package-R run; and
`acceptance_verdict` is exactly `ACCEPT`. A rejected independent review creates
no Package-D artifact or commit and therefore cannot be input to E. The
accepted record names no Package-D commit and no Package-E commit: its later
containing commit is supplied only to `Write-HubQaG0Activation` as
`AcceptanceCommit`.

Every child object is closed, and each `path` is a physical regular file with a
repository-relative `/` path while each `raw_sha256` is the SHA-256 of that
file's exact no-BOM raw bytes. `founder_release` has exactly `path`,
`raw_sha256`, `commit`, whose values are respectively the fixed Package-R path
`.planning/audits/test-rebase/g0-bootstrap-founder-release.v1.json`, its raw
digest, and its dedicated R commit. `g0_plan_commit`, `a0_commit`,
`a1_commit`, `b_commit`, task, run, and all handoff/authority facts must equal
the exact raw Package-R artifact; none may be reconstructed from a caller or
from a mutable worktree.

`c_receipts` has exactly these keys in this order:

```text
owner_marker, local_base_inspect, helper_base_verification,
helper_image_attestation, ingress_copy, helper_admission, helper_cleanup,
source_tree_verification, local_image_inspection, candidate_lock,
planning_selftest_negative_red, planning_selftest_positive_green,
command_receipts, run, cleanup
```

Each `c_receipts` child has exactly `path`, `raw_sha256`. All paths are below
the Package-R-derived task report root. The fixed bootstrap names are
`bootstrap/owner-marker.v1.json`, `bootstrap/local-base-inspect.v1.json`,
`bootstrap/bootstrap-helper-python.verification.v1.json`,
`bootstrap/g0-admission-helper.local-image-attestation.v1.json`,
`bootstrap/ingress-copy-receipt.v1.json`,
`bootstrap/g0-admission-helper.admission.v1.json`,
`bootstrap/g0-admission-helper.cleanup.v1.json`,
`bootstrap/source-tree-verification.v1.json`,
`bootstrap/local-image-inspect.v1.json`, and `bootstrap/images.lock.v3.json`.
The five remaining paths are exactly `selftests/g0-planning-negative-red.v1.json`,
`selftests/g0-planning-positive-green.v1.json`,
`commands/command-receipts.v1.json`, `run.json`, and `cleanup.json`.
The owner-marker link must name the same exact raw
`bootstrap/owner-marker.v1.json` bytes used by both private contexts; the
local-base and local-image links remain distinct and cannot share bytes or a
digest. The source-tree link must bind the create-new pre-build receipt; the
candidate-link raw digest must be byte-for-byte equal to the candidate examined
by C1.

`c1_lock` has exactly `path`, `raw_sha256`, `commit`: path is
`testframework/docker/images.lock.v3.json`, its bytes and digest are exactly
the C candidate bytes/digest, and commit is exactly `repository_commit`, the
separate C1 commit. `repository_snapshot` and `run_envelope` each have exactly
`path`, `raw_sha256`; they are respectively the retained
`snapshots/000-G0_BOOTSTRAP.repository-snapshot.v1.json` and
`envelopes/g0-assets-runner-envelope-<run>.run-envelope.v1.json` records below
the same report root. Their raw values must satisfy the closed shared
snapshot/envelope bidirectional link rules. Thus R supplies the immutable
authority, C supplies only retained raw evidence, C1 supplies the committed
lock, D accepts those facts, and E can only read that completed causal chain.

The D parser/reviewer must reject duplicate, reordered, malformed, unknown,
missing, noncanonical, nonphysical, or raw-digest-mismatched values with one of
`E_B0_D_CANONICAL_JSON`, `E_B0_D_SCHEMA_CLOSED`, `E_B0_D_RECEIPT_LINK`,
`E_B0_D_R_BINDING`, `E_B0_D_C1_BINDING`,
`E_B0_D_SNAPSHOT_ENVELOPE_LINK`, `E_B0_D_VERDICT`, or
`E_B0_D_SELF_REFERENCE`. `Write-HubQaG0Activation` duplicate-aware parses this
same record and maps any such D failure, a non-`ACCEPT` verdict, or an absent
required C/C1/R link to `E_B0_ACTIVATION_D_ACCEPTANCE`; it must not treat an
unparsed or re-emitted D object as evidence.

### `Write-HubQaG0Activation` closed interface and causal inputs

`testframework/runner/write-g0-activation.ps1` has one named internal command
surface and no positional, environment, JSON, stdin, path, task, run, output,
or report-root override:

```powershell
Write-HubQaG0Activation `
  -RequiredPlanCommit <40-lowercase-hex> `
  -FounderReleaseCommit <40-lowercase-hex> `
  -RepositoryCommit <40-lowercase-hex> `
  -AcceptanceCommit <40-lowercase-hex>
```

The function is callable only when committed `HEAD` is exactly
`AcceptanceCommit`; an invocation against another `HEAD`, direct script-file
execution, a missing/extra parameter, or a dirty replacement of any input is
`E_B0_ACTIVATION_INPUT`. It resolves every path from its own physical script
root and uses Git object bytes, never caller paths. In particular it requires:

1. `RequiredPlanCommit` contains this G0 plan at its fixed repository path and
   is an accepted ancestor of `FounderReleaseCommit`;
2. `FounderReleaseCommit` is the dedicated Package-R commit and its fixed
   release raw bytes satisfy the global ordered JSON rules, have
   `g0_plan_commit=RequiredPlanCommit`, the fixed task/run/source baseline,
   no-network/local-only/one-shot values, and exact accepted A0/A1/B bindings;
3. `RepositoryCommit` is the Package-C1 commit, contains exactly the
   byte-identical C candidate at `testframework/docker/images.lock.v3.json`,
   and is neither R, D, E, an A package, nor a report-only commit; and
4. `AcceptanceCommit` is the separate Package-D commit containing exactly one
   canonical record satisfying `Package-D evidence-acceptance record contract`:
   `acceptance_verdict=ACCEPT`, `repository_commit=RepositoryCommit`, the raw
   Package-R path/digest/commit, complete closed C receipt links, the exact C1
   lock and shared snapshot/envelope links, and no self/future reference.

The function derives the report root from the verified release task/run, reads
the retained C `run.json`, `cleanup.json`, planning-selftest, repository
snapshot, and envelope raw bytes there, and rejects a missing/mismatched
receipt rather than accepting an argument. It produces only the create-new
ordered canonical file `.planning/audits/test-rebase/g0-activation.v1.json`
and returns this ordered result object:

```text
activation_path, activation_sha256, required_plan_commit,
founder_release_commit, repository_commit, acceptance_commit
```

The output's `activation_id` is exactly `g0-activation-<run>`; it copies the
verified Package-R path/raw digest/commit, uses `repository_commit` exactly as
the C1 SHA, and uses `acceptance_commit` exactly as D's SHA. It fills every
already-listed activation field from the Git object or retained C receipt that
owns it, including the two shared schema/producer paths and digests and every
planning-selftest digest. Package E commits the created file only afterwards;
therefore neither `acceptance_commit` nor `repository_commit` can point to the
future Package-E commit. Package A receives the whole activation record as
immutable pre-genesis input and must copy every activation field, including
`founder_release_path`, `founder_release_sha256`, and
`founder_release_commit`, rather than reconstructing a release mapping.

Stable failure codes are: `E_B0_ACTIVATION_INPUT` (closed interface, HEAD,
or ancestor failure), `E_B0_ACTIVATION_RELEASE_BINDING` (Package-R path/raw
bytes/fields), `E_B0_ACTIVATION_C1_BINDING` (wrong C1/lock/target mapping),
`E_B0_ACTIVATION_D_ACCEPTANCE` (any stable D parser/reviewer failure, wrong D
verdict, or missing/mismatched C/C1/R receipt linkage), and
`E_B0_ACTIVATION_SELF_COMMIT` (any E/self/future commit reference). A valid
write must be rejected if the destination already exists; it never overwrites,
normalizes, stages, or commits the output.

## Execution packages and gates

### Package 0 — commit the plan

**Files:** only `.planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md`.

**STOP:** Any HIGH finding from independent Terra/high plan review blocks source authoring.

```powershell
git commit --only .planning/phases/129-tf-e2e-shared/129-G0-ASSETS-RUNNER-PLAN.md -m "docs(test-rebase): plan G0 bootstrap pre-genesis"
```

### Package A0 — static source contract

**Files:**

- `testframework/docker/dependency-bundle.schema.v1.json`
- `testframework/docker/admission/verify_local_dependency_bundle.py`
- `testframework/docker/admission/admit_local_dependency_assets.py`
- `testframework/docker/Dockerfile.g0-admission-helper`
- `testframework/docker/entrypoints/g0-admission-helper.sh`
- `testframework/docker/admit-local-dependency-assets.ps1`
- `testframework/docker/source-image-attestor.ps1`
- `testframework/docker/Dockerfile.backend-unit`
- `testframework/docker/Dockerfile.frontend-unit`
- `testframework/docker/Dockerfile.backend-openapi`
- `testframework/docker/Dockerfile.g0-planning-selftest`
- `testframework/docker/entrypoints/backend-unit.sh`
- `testframework/docker/entrypoints/frontend-unit.sh`
- `testframework/docker/entrypoints/backend-openapi.sh`
- `testframework/docker/entrypoints/g0-planning-selftest.sh`
- `testframework/docker/runner-fixtures/g0-planning-selftest.py`
- `testframework/docker/vitest.frontend-unit.config.ts`
- `testframework/docker/compose.g0-bootstrap.yaml`
- `testframework/docker/bootstrap-control.v1.json`
- `testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1`
- `testframework/selftests/runner/g0-image-contract.ps1`

`testframework/docker/verify-local-dependency-bundle.ps1` is retired from this plan and is not an A0 deliverable or commit allowlist entry. A pre-amendment worktree draft of that path is unaccepted and must not be staged or committed under any package.

**Failing-first source contracts:** author both static PowerShell selftest sources and the exact Python-stdlib image-baked fixture. `g0-planning-selftest.py` must import the A0 Python public verifier/admission modules directly, not duplicate their parser or emit fixture metadata. It specifies literal valid, duplicate-ID, SHA mismatch, path-escape, missing-asset, and unbound-manifest fixtures, and asserts that a rejected fixture yields its exact reason code and no admission success receipt. `g0-bootstrap-attestor-contract.ps1` statically asserts the host launcher has only runner-private bootstrap context, the fixed helper image/receipt paths, no public CLI path/JSON/Docker input, and no Linux PowerShell invocation. `g0-image-contract.ps1` statically asserts the helper is absent from the four-target topology/Compose/control/lock while its Dockerfile/entrypoint and fixed base binding are present, and that all four actual target commands remain closed. The two PowerShell selftests and the host launcher are not copied into either Linux image. All contracts state `NOT_READY/NOT_RUN` before bootstrap. No claim of red or green execution is permitted before Package C.

**Static acceptance:** independent Terra/high verifies the dependency schema, two Python-stdlib admission modules, helper Dockerfile/entrypoint, Windows-host launcher contract, source-image-attestor code, the four Dockerfiles and four entrypoints, dedicated Vitest config, bootstrap Compose/control, image-baked Python fixture, and both static selftest sources. It verifies closed trusted roots; the Package-R-derived, untracked handoff-root contract; fixed source baseline versus later control commits; the mandatory Git-object-only `100644` allowlist/tree-digest algorithm and its stable pre-build gates; the owner marker then all-local-base inspection then nonce/digest-bound receipt sequence; non-self-referential bundle digest; closed SPDX policy; A1-bound `bootstrap-helper-python` identity; retirement of `Assert-BaseBinding`; no caller-forged repository/path/record/receipt/base-inventory/ingress input; no network/install/source mount/arbitrary command; historical v1/v2 policy files unchanged; and exact four-target mapping. It verifies that helper build/attestation precedes the create-new verified ingress copy, helper admission, helper cleanup, mandatory source-tree verification, and only then four-target builds; that the helper is a pre-admission ephemeral implementation detail with fixed receipt and cleanup paths but no Compose/control/lock/activation record; and that `g0-planning-selftest` copies and calls only the shared Python logic plus its three fixed A1 controls. It rejects every PowerShell asset, Linux `pwsh`, PowerShell `COPY`, in-image host launcher, Package-B copy, A0/B cycle, caller-selected handoff path, or direct handoff-root mount/read by helper. A HIGH finding blocks the package.

**STOP:** A0 creates no `dependency-bundle.manifest.v1.json`, `source-policy.v2.json`, `build-manifest.v3.json`, `images.lock.v3.json`, lock, or receipt. Without the founder's offline ingress and separate bootstrap release, do not build, attest, execute a selftest, or write a lock-derived runtime receipt. A0 status is `STATIC_ACCEPTED` only after review; runtime is `NOT_READY/NOT_RUN`. This plan does not authorize any network ingress.

```powershell
git commit --only testframework/docker/dependency-bundle.schema.v1.json testframework/docker/admission/verify_local_dependency_bundle.py testframework/docker/admission/admit_local_dependency_assets.py testframework/docker/Dockerfile.g0-admission-helper testframework/docker/entrypoints/g0-admission-helper.sh testframework/docker/admit-local-dependency-assets.ps1 testframework/docker/source-image-attestor.ps1 testframework/docker/Dockerfile.backend-unit testframework/docker/Dockerfile.frontend-unit testframework/docker/Dockerfile.backend-openapi testframework/docker/Dockerfile.g0-planning-selftest testframework/docker/entrypoints/backend-unit.sh testframework/docker/entrypoints/frontend-unit.sh testframework/docker/entrypoints/backend-openapi.sh testframework/docker/entrypoints/g0-planning-selftest.sh testframework/docker/runner-fixtures/g0-planning-selftest.py testframework/docker/vitest.frontend-unit.config.ts testframework/docker/compose.g0-bootstrap.yaml testframework/docker/bootstrap-control.v1.json testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1 testframework/selftests/runner/g0-image-contract.ps1 -m "feat(qa): add G0 static source contract"
```

### Package A1 — binding data after approved offline ingress

**Files:** only:

- `testframework/docker/dependency-bundle.manifest.v1.json`
- `testframework/docker/source-policy.v2.json`
- `testframework/docker/build-manifest.v3.json`

**Binding rule:** A1 is blocked until a founder-approved offline ingress closes the exact physical `.qa-local-ingress/g0-assets/<handoff-id>` handoff root and its owner marker/manifest hashes. It reads no caller path and records only the byte-for-byte manifest, dependency bytes, local base image IDs/digests, SPDX licenses, license-evidence digests, and provenance digests from that fixed handoff into the three binding data files. It makes no code, Dockerfile, entrypoint, control, runner, selftest, or documentation change. The dependency schema value `source_commit` is the constant `7b5d5501cefab34f9aa5661e4f8c38e71474f9ce`; `bases` binds the required `bootstrap-helper-python` identity for both `g0-admission-helper` and `g0-planning-selftest`. `build-manifest.v3.json` binds the helper Dockerfile/entrypoint digests and excludes it from the four target definitions and `images.lock.v3.json`.

**Independent acceptance:** Terra/high verifies the exact data only against the separately explicit founder-approved fixed offline-ingress handoff, its repository-root-derived physical handoff root, owner marker, fixed source baseline, raw manifest bytes, A1 raw files, local availability/provenance evidence, schema, historical v1/v2 files, and A0's frozen contracts. This acceptance precedes Package R and must not read, require, or compare any Package-R bootstrap release artifact. It rejects a `powershell-runtime` asset or any asset/consumer not in the closed A0 mapping. Missing, mutable, remote, tagged-only, nonlocal, reparse/ADS-tainted, inferred, invented, caller-provided, or byte-mismatched values are `A1=BLOCKED_EXTERNAL_INGRESS`; no commit is made.

**STOP:** A1 has no network authority, no build, no attestation, no selftest execution, no lock, and no runtime receipt. It may not begin before approved ingress exists, and it may not change an A0 or B-owned path.

```powershell
git commit --only testframework/docker/dependency-bundle.manifest.v1.json testframework/docker/source-policy.v2.json testframework/docker/build-manifest.v3.json -m "feat(qa): bind G0 offline ingress data"
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

`hub.qa-runner/v2` accepts only validated `target`, `task`, and `run`; it rejects extras with exit `64`, has no v1 bridge, no public direct invocation, and no general command operation. It supports the four fixed target names but reports all future Python jobs `NOT_READY`. The G0 selftest is callable only from the bootstrap launcher with the exact internal role capability above; `/qa-gate` documents v2 but is not a bootstrap route.

**Mandatory A0 call chain:** Package B source must dot-source the accepted A0
`testframework/docker/source-image-attestor.ps1` into its bootstrap-private
scope, retaining no separate source-tree implementation. During C, after
helper cleanup and after A1 controls are bound but before the first of the four
target image builds, it calls exactly
`Get-HubQaGitObjectSourceTree -Controls $binding.Controls`. It creates exactly
one owner-bound receipt at
`testframework/reports/g0-assets-runner-<run>/bootstrap/source-tree-verification.v1.json`
with ordered keys `schema`, `producer`, `task`, `run`, `owner_marker_sha256`,
`source_commit`, `source_tree_sha256`, `entry_count`,
`source_policy_sha256`, `build_manifest_sha256`, `created_at_utc`; the schema
is `hub.qa.source-tree-verification/v1` and producer is
`hub.qa-bootstrap/v1`. The four builds are forbidden unless this receipt is
create-new, raw-digestable, and exactly matches the returned A0 source commit,
tree digest, entry count, private task/run/owner marker, and raw A1 controls;
a failed call/receipt is `E_B0_SOURCE_TREE_PREBUILD` and `G0_C=FAIL/BLOCKED`.
After all four immutable local image inspection records exist, B calls exactly
`Invoke-HubQaBootstrapImageAttestation -BootstrapPrivateContext
$ImageAttestationPrivateContext`. Its full A0 return must be
`READY_FOR_G0_BOOTSTRAP_ONLY`, name the same candidate path/digest and source
tree digest, and independently repeats the Git-object source-tree gate before
candidate publication. It may not substitute an inspection-only helper,
hand-authored candidate, or a partial A0 function; a wrong return/path/digest
is `E_B0_IMAGE_ATTESTATION_POSTBUILD` and blocks C/C1/D/E.

The post-build records are one create-new ordered canonical
`testframework/reports/g0-assets-runner-<run>/bootstrap/local-image-inspect.v1.json`
with exactly `schema`, `producer`, `task`, `run`, `source_commit`,
`owner_marker_sha256`, `receipt_nonce`, `controls`, `targets`. Its schema is
`hub.qa.local-image-inspect/v1`, its producer is `hub.qa-bootstrap/v1`, and
its task/run/source baseline/owner marker/nonce must equal the private context.
`controls` has exactly `dependency_bundle_sha256`, `source_policy_sha256`,
`build_manifest_sha256`; they are the raw A1 control digests. `targets` has
exactly the four fixed target keys in canonical target order. Each has exactly
`local_image_id`, `platform`, `entrypoint`, `cmd`, `labels`; A0 validates every
child against the static control and local inspect. The earlier
`local-base-inspect.v1.json` is a separate before-helper admission receipt and
must never be renamed, substituted, or re-used as the local-image receipt.

**Failing-first static source contracts:**
`testframework/selftests/runner/g0-runner-v2-contract.ps1` contains the
following fixed source-fixture table in its own ordered data (no new fixture
files and no runtime execution before C):

| Fixture label | Required assertion / stable code |
|---|---|
| `noncanonical-bootstrap-json` | wrong UTF-8/BOM, whitespace, duplicate, unknown, or reordered key is `E_B0_CANONICAL_JSON` or `E_B0_SCHEMA_CLOSED` |
| `wrong-version-or-producer` | bootstrap/snapshot/envelope schema or producer drift is `E_B0_SCHEMA_VERSION` |
| `short-or-foreign-run` | a run outside `^[a-z0-9][a-z0-9-]{7,63}$`, a changed task, or a noncanonical report root is `E_B0_TASK_RUN` |
| `serialized-role-claim` | role claim/context supplied through CLI/environment/JSON/report/mount is `E_B0_ROLE_TRANSPORT` |
| `direct-planning-selftest` | direct script, public v2 G0 target, or duplicate closure call is `E_B0_DIRECT_INVOCATION_FORBIDDEN` or `E_B0_ONE_SHOT_REJECTED` |
| `snapshot-envelope-drift` | wrong nonce/task/run/role/report root, nullability, schema digest, or either raw-byte link is `E_B0_SNAPSHOT_ENVELOPE_LINK` |
| `missing-prebuild-source-receipt` | missing/wrong A0 `Get-HubQaGitObjectSourceTree` call or receipt gates builds with `E_B0_SOURCE_TREE_PREBUILD` |
| `partial-postbuild-attestation` | a partial/wrong A0 `Invoke-HubQaBootstrapImageAttestation` return blocks with `E_B0_IMAGE_ATTESTATION_POSTBUILD` |
| `d-evidence-acceptance-contract` | duplicate/reordered/unknown D keys, a wrong C/C1/R raw link, non-`ACCEPT` verdict, or D self/future reference is the matching `E_B0_D_*` error and causes `Write-HubQaG0Activation` to reject with `E_B0_ACTIVATION_D_ACCEPTANCE` |
| `activation-r-c1-d-chain` | wrong Package-R raw digest, C1 lock commit, D acceptance, output collision, or self/future commit is the exact `E_B0_ACTIVATION_*` code above |
| `first-exit-cleanup` | first non-zero exit wins; timeout is `124`; foreign resource/report exclusion and exact cleanup remain closed |

The host PowerShell sources and this static selftest are reviewed as source
only before C; neither is copied to a Linux image as a false runtime proof.
Their executable behavior is evidenced only by the actual founder-released C
transaction. All fixtures are `NOT_RUN` until bootstrap.

**Static acceptance:** independent Terra/high checks the global canonical
JSON/identifier contract; all three closed schemas and their producer/consumer
links; the exact in-memory role capability and direct-invocation rejection; no
arbitrary inputs or v1/v2 bridge; owner marker before any dependency operation;
all-local-base inspect plus nonce/digest receipt before helper-base
verification/build; the mandatory pre-build and post-build full A0 call chain;
the runner-private ingress digest-map and ingress-copy receipt propagation; no
activation before accepted evidence; and that the bootstrap exception remains
founder-released. It verifies that B/C derive the handoff root only from the
accepted committed Package-R artifact, bootstrap/B alone copies it create-new
only after helper attestation, the Windows-host admission launcher only
consumes immutable task-owned copied ingress plus held private receipt/digest
map during C, cleanup removes only the copied ingress subtree, and static
PowerShell selftests cannot become runtime evidence.

**STOP:** A HIGH finding, an absent independently accepted A0 verdict, or a request to execute any source contract without founder release blocks Package B. B static source may be authored and accepted after A0 while A1 waits for ingress, because its allowlist does not overlap A0 or A1. B runtime remains blocked until A1 is independently accepted, the founder explicitly approves, and the dedicated Package-R artifact commit is independently accepted. Do not create a runtime receipt.

```powershell
git commit --only testframework/runner/contracts/hub.qa-bootstrap.v1.json testframework/runner/invoke-qa-bootstrap-v1.ps1 testframework/runner/invoke-qa-gate-v2.ps1 testframework/runner/run-task-v2.ps1 testframework/runner/new-repository-snapshot.ps1 testframework/runner/new-run-envelope.ps1 testframework/runner/write-g0-activation.ps1 testframework/schemas/repository-snapshot.schema.v1.json testframework/schemas/run-envelope.schema.v1.json testframework/selftests/runner/g0-runner-v2-contract.ps1 testframework/runner/README.md .cursor/skills/qa-gate/SKILL.md -m "feat(qa): add G0 bootstrap runner controls"
```

### Package C — founder-released one-time local bootstrap

**Repository source files:** none. This runtime package consumes the accepted Package A0, A1, and B commits and produces only task-owned reports, including the candidate `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`. It never creates, edits, stages, or commits `testframework/docker/images.lock.v3.json`.

**Founder prerequisites:** the one already-committed, independently accepted Package-R authority artifact at the fixed path and at current committed `HEAD`; no caller may supply, select, or override it. Its raw bytes name the exact plan/A0/A1/B commits, task, unique run, local-only/no-network authority, closed offline-ingress handoff ID, its owner-marker SHA-256, its dependency-manifest SHA-256, fixed source baseline, expiry, one-shot nonce, and explicit founder approval. C requires `HEAD` to equal the dedicated Package-R commit, derives the physical repository-root `.qa-local-ingress/g0-assets/<handoff-id>` tree only from those raw bytes, and rejects an expired/reused nonce or any other artifact/history state. An absent ingress item is `G0_C=BLOCKED_EXTERNAL_INGRESS`; an absent/rejected Package-R artifact or any other authority failure is `G0_C=BLOCKED`.

**Only permitted operation:** call `hub.qa-bootstrap/v1` once from the accepted committed Package-R authority. In exact order, it creates the owner marker; inspects all local bases and writes the raw nonce/digest-bound local-base receipt; verifies the A1 helper base identity from that receipt; builds and locally attests the ephemeral `g0-admission-helper`; copies and verifies the fixed founder handoff into the create-new task-owned ingress subtree while retaining its in-memory digest map and writing the closed ingress-copy receipt; gives `admit-local-dependency-assets.ps1` only the immutable copied ingress and runner-held private receipt/digest map; and lets the helper verify/admit only those copied staged bytes. The launcher neither reads nor copies/adopts the founder handoff/ingress. Bootstrap then removes and receipts the helper; binds the closed A1 controls; calls the full A0 `Get-HubQaGitObjectSourceTree -Controls $binding.Controls`; writes and validates the exact source-tree-verification receipt; and only then builds the four fixed targets. After all four immutable local inspection records exist it calls the full A0 `Invoke-HubQaBootstrapImageAttestation -BootstrapPrivateContext $ImageAttestationPrivateContext` and accepts no result except the exact post-build value/path/digests defined for Package B. Only then does it retain the candidate v3 lock at its fixed task-owned path and invoke `g0-planning-selftest` through the private v2 closure. The planning selftest executes the shared Python fixture negative-red and positive-green cases; the outer host launcher and B runner behavior are proven by this real transaction rather than a Linux structural imitation. It does not run `backend-unit`, `frontend-unit`, `backend-openapi`, a package Python job, or any legacy suite.

**Evidence:** accepted Package-R artifact raw digest/commit/expiry/nonce-use receipt; report root; owner marker; raw nonce/digest-bound all-local-base inspection receipt; helper-base identity verification receipt; helper local-image attestation; closed ingress-copy receipt with its digest-map SHA-256 and fixed source/destination roots; mandatory Git-object source-tree verification receipt with its canonical digest; helper admission receipt; helper cleanup receipt; immutable four-target local image-ID attestations; candidate v3-lock receipt and raw candidate bytes; closed snapshot; closed envelope; negative-red receipt; positive-green receipt; ordered command receipts; `run.json`; and `cleanup.json`. Candidate lock bytes remain under the task-owned report root for review; they are retained with failed evidence and are never regenerated for C1. The transaction preserves the first exit, applies fixed timeouts, and unconditionally cleans exactly the initially owned resources, including only its copied ingress subtree.

**STOP:** Any missing receipt, unexpected negative pass, failed positive case, nonlocal image, leak, timeout, nonzero exit, digest disagreement, report collision, or cleanup ambiguity is `G0_C=FAIL/BLOCKED`; preserve evidence and do not activate.

### Package C1 — independent candidate-lock admission

**Input:** exactly one successful Package C report root, its immutable candidate `testframework/reports/<task-run>/bootstrap/images.lock.v3.json`, the candidate SHA-256 receipt, the nonce/digest-bound local-base receipt, helper attestation/admission/cleanup receipts, all four immutable local image-ID attestations, `bootstrap-control.v1.json`, the accepted committed Package-R artifact/raw digest/commit, and the accepted Package A0/A1/B commit SHAs. **Output:** either `G0_C1=REJECT/BLOCKED` with no repository mutation, or one repository file `testframework/docker/images.lock.v3.json` whose raw no-BOM bytes are byte-for-byte identical to the candidate. C1 never rebuilds, reattests, regenerates, normalizes, or edits a lock.

**Independent review:** Terra/high verifies the committed Package-R raw authority digest/commit, exact plan/A0/A1/B bindings, expiry/non-reused nonce, and no-network/local-only fields; the raw receipt's create-new owner binding, in-memory nonce/digest linkage, A1 `bootstrap-helper-python` identity, mandatory source-tree verification evidence, and the helper's exact build/attest/admit/cleanup sequence before comparing each target's `local_image_id`, `local_image_id_sha256`, `entrypoint_path`, `entrypoint_sha256`, `command_sha256`, `compose_service`, and `runtime_status` against the candidate, raw C receipts, local attestations, and bootstrap control. It verifies the helper is absent from the lock, all four target names are present, the candidate digest/report containment are exact, and no SUT target ran. Its review decision must precede the C1 commit.

**Allowlist:** only `testframework/docker/images.lock.v3.json`.

**STOP:** Missing, mismatched, regenerated, nonlocal, extra, or unverifiable candidate data is `G0_C1=REJECT/BLOCKED`; do not write a repository lock, do not stage any other path, and do not continue to D. On success, the C1 commit is the only `repository_commit` allowed in D and E.

```powershell
git commit --only testframework/docker/images.lock.v3.json -m "feat(qa): admit G0 bootstrap image lock"
```

### Package D — independent evidence acceptance

**Files:** create only `.planning/audits/test-rebase/g0-bootstrap-evidence-acceptance.v1.json`.

The independent Terra/high reviewer checks every closed field/link in `Package-D evidence-acceptance record contract`: the frozen contracts, the accepted Package-R artifact/commit/raw digest and its exact plan/A0/A1/B bindings, the C1 `repository_commit`, and every Package C receipt against exact raw bytes, including the distinct local-base and local-image receipts, the mandatory source-tree receipt, helper build/attest/admit/cleanup receipts, the closed shared snapshot/envelope link, the four C1 lock records, exact resource cleanup, first-exit semantics, and proof that no SUT target ran. It rejects an in-container execution of a Windows host launcher as proof. The record must name the exact C1 SHA as `repository_commit`; no A, B, R, C report, D, or E commit is valid in this field. Only an `ACCEPT` record is committed separately before activation; a rejection writes no Package-D artifact and creates no activation record.

```powershell
git commit --only .planning/audits/test-rebase/g0-bootstrap-evidence-acceptance.v1.json -m "docs(test-rebase): accept G0 bootstrap evidence"
```

### Package E — activation record after acceptance

**Files:** create only `.planning/audits/test-rebase/g0-activation.v1.json`.

`Write-HubQaG0Activation` is invoked only after Package D's independent acceptance commit exists and must set `repository_commit` to the already-existing C1 lock commit. It also records the earlier Package-R artifact path, raw digest, and commit used by C; its raw bytes must match that committed artifact. `acceptance_commit` points to that already-existing commit, never the Package E commit. Package E therefore records causal evidence without a self-commit cycle. It remains a pre-genesis input, not a recovery event.

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
| A0 | dependency schema, shared Python admission modules, fixed physical founder-handoff contract, ephemeral helper contract, Windows-host launcher/source-image-attestor, four-target topology, Compose/control, and both static contracts | none | `NOT_READY/NOT_RUN` |
| A1 | exact fixed-root founder ingress binding data only; independent raw-byte/value/provenance review | none | `BLOCKED_EXTERNAL_INGRESS` until released ingress exists |
| B | bootstrap/v2/shared schemas and producers, closed release-to-root derivation, base-receipt/helper/ingress-copy ordering, static contracts; starts only after accepted A0 and has no A1 overlap | none | `NOT_READY/NOT_RUN` |
| R | explicit founder-approved, separately committed repository authority with exact plan/A0/A1/B allowlists, source/handoff bindings, expiry and one-shot nonce | none | `BLOCKED` until the artifact exists and is independently accepted |
| C | committed Package-R authority at `HEAD` with fixed handoff and all accepted A0/A1/B packages | one bootstrap transaction: all-base nonce-bound receipt, helper identity/build/attest, create-new ingress-copy receipt, mandatory source-tree verification, helper admission/cleanup, planning selftest only, red/green fixtures, exact cleanup | `BLOCKED_EXTERNAL_INGRESS` or `BLOCKED` |
| C1 | independent candidate/receipt/local-attestation comparison; only lock allowlist | committed byte-identical repository lock | `REJECT/BLOCKED` unless exact candidate is present |
| D | independent Terra/high C/C1 receipt review | accepted evidence commit naming C1 as `repository_commit` | `REJECT` unless complete |
| E | activation record static validation against C1 plus prior D acceptance | references prior D acceptance only | `BLOCKED` until C1 and D exist |
| Recovery A | activation is immutable input | separate recovery evidence | `CUTOVER_ACCEPTANCE=FAIL/BLOCKED` until separately accepted |

## Rollback and handoff

Rollback is forward-only: revert only an exact named source-package commit after a separate review; never alter the 241 staged paths. A bootstrap failure retains its task-owned evidence and removes only its proven owned resources. No automatic retry, continuation, lock edit, or digest rewrite is permitted; a new founder release must name a new run.

Before A0 source authoring, commit this corrected plan separately and obtain independent Terra/high review. Until all prerequisites exist, hand off exactly: `G0=NOT_READY`, `A1=BLOCKED_EXTERNAL_INGRESS`, `BOOTSTRAP=NOT_RUN`, `CUTOVER_ACCEPTANCE=FAIL/BLOCKED`.
