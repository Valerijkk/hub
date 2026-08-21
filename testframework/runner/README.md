# Private QA runner boundary

## Status

This document describes an unexecuted local-only v3 private contract.
`testframework/docker/images.lock.v3.json` is absent, and the required local
dependency bytes are absent. The package and launcher are `NOT_READY`;
PowerShell parsing is `PS5.1_PARSE_NOT_RUN`, private contract execution is
`PRIVATE_CONTRACT_NOT_RUN`, and runtime is `NOT_RUN`.
`G0-B_PARITY_NOT_VERIFIED` remains the status for baked-entrypoint parity.
Static source reviews for C1, P2, and P4 are accepted, but they provide no
runtime `PASS`, Docker verification, generated-lock validation, or launch
readiness. P7 Docker-only verification remains blocked pending separately
approved READY inputs and evidence.

## Task 2A committed static launcher receipt

`e357e71ae445c797c686787e8cde2d643566bb68` (`feat(qa): добавить закрытый
public launcher`) is an immutable static-source receipt with sole parent/source
revision `cc0f3fcbf4e9a0adc70e946e739b81d21ac7da6c` and tree
`b16311222d3093ae4d2bb462a2e4d2835a3f38cf`. Its exact closed bundle is:

| Path role | Blob |
|---|---|
| runner contract | `31084540fc841476d80b9b37f09e151011aa6ecc` |
| runner fixtures | `8a47518dfcdf35352db905dc575a557e689d196b` |
| public launcher contract | `0a4872e9ab4646203c687eae060c455cc00137e2` |
| public `invoke-qa-gate.ps1` | `46e9dacfe87d7b2841bb11929a7d8a2dff61429a` |
| private `run-task.ps1` | `f3dbe32f8c1565adf27742d8829860e0d37ead9b` |
| build manifest | `da1524cca5f68a4962680b4b0c8272fd22359df9` |
| source policy | `04148a8a0d6ddb0836437bbf7b9c91d07b996fd2` |

Contract (Terra/high), Security (Terra/xhigh), and Lifecycle (Terra/high)
precommit and immutable exact-SHA reviews are all PASS C/I/M `0/0/0`;
provenance is dispatcher-enforced, with no fallback and telemetry `UNKNOWN`.
This record does not change readiness: static review is not PowerShell parsing,
private-contract, Docker, image, target, cleanup, or runtime evidence.

The public entrypoint raw ABI is exactly `-Target <target> -Task <task> -Run
<run>`; malformed, extra, and unknown input exit `64` before roots or effects.
Known targets emit one safe `hub.qa-gate/v1` JSON line and exit `69` while all
are literal `NOT_READY`, so no private runner, Docker, or report effect occurs.
Its roots are portable from the script location, reject reparse ancestry, accept
normal and linked-worktree `.git` forms without parsing a pointer, and use only
the fixed sibling `run-task.ps1` for a future ready handoff.

This documentation receipt is conditional: it becomes current only after the
one atomic seven-path Task 2B documentation commit and immutable exact-SHA
review. Until then earlier committed documentation governs. It does not
self-activate a lifecycle or select Task 3.

## Authority and fixed targets

`/qa-gate` is the sole public execution authority. Its private implementation
detail is `testframework/runner/run-task.ps1`; it is not a supported direct
command, alternate entrypoint, or fallback route.

The private request contract is `hub.qa-runner/v1` and has exactly three fixed
target identities:

- `backend-unit`
- `frontend-unit`
- `backend-openapi`

Future Python jobs, sentinel, selftest, fixture, unknown, and caller-selected
target names are outside this candidate and return `NOT_READY` without entering
the private runner.

The private request contains only `Contract`, `Target`, `Task`, and `Run`.
It cannot carry an executable, path, argument vector, image or Compose
override, timeout profile, dependency object, resource inventory, fixture,
selftest, arbitrary command, or selector. All identities for the project,
container, report root, Compose definition, image lock, and executable are
derived internally from the canonical repository, fixed target, and validated
task/run identifiers.

The internal dependency seam has exactly four semantic operations:

1. `InspectLockedImage`
2. `RunFixedTarget`
3. `ListOwnedResources`
4. `CleanupOwnedResources`

No fifth dependency operation, caller-controlled I/O surface, or public
execution route is permitted.

## Local image identity and v3 control provenance

The proposed image lock schema is exactly `hub.qa.local-images-lock/v3` in
`testframework/docker/images.lock.v3.json`. It is a local Docker-engine
contract: every target is selected only by its complete local `image_id`.
Target `repo_digest` members reject the lock. Observed target tags and
RepoDigests are ignored for identity and selection. Base `repo_digest` is
local inventory/admission evidence only and is never a selector, pull trigger,
or target identity field.

The lock has exactly the fixed bases `golang`, `playwright`, and `python`, and
the three fixed targets above. Its parser is strict and duplicate-aware:
duplicate, missing, additional, malformed, escaped, or reparsed members fail
before report creation, owner-marker creation, or dependency calls. Each exact
Entrypoint array, Cmd array, and complete labels projection is retained as its
own value. Their hashes are computed over no-BOM UTF-8 RFC 8785 JSON
Canonicalization Scheme (JCS) bytes; null substitution, array concatenation,
or partial label projection is forbidden.

Private attestation accepts one mandatory literal 40-lowercase-hex
`ControlCommit`. It reads exactly these six mode-`100644` Git blobs from that
commit, in this order, and derives both the ordered control-bundle digest and
the raw Compose-byte digest:

1. `testframework/docker/Dockerfile.backend-openapi`
2. `testframework/docker/Dockerfile.backend-unit`
3. `testframework/docker/Dockerfile.frontend-unit`
4. `testframework/docker/build-manifest.v2.json`
5. `testframework/docker/compose.unit.yaml`
6. `testframework/docker/source-policy.v1.json`

The control-bundle grammar is exactly
`path<TAB>100644<TAB>git-blob-sha1<TAB>sha256-of-exact-blob-bytes<LF>` for
each ordered row. `control.compose_sha256` is the SHA-256 of the exact raw
Compose blob bytes, not a YAML parse or normalized rendering. Worktree copies
of control files are non-authoritative. Product source is materialized only
from mode-`100644` Git blobs at the source policy's `source_commit` (recorded
in the lock as `source.commit`), never from a source bind, checkout,
working-tree copy, or Git metadata inference.

The private attestor is excluded from the control bundle and does not
self-attest. Its integrity is instead an out-of-band review and commit-admission
gate. Existing lock objects fail closed: the attestor must not overwrite,
replace, delete, or adopt an existing `images.lock.v3.json` object.

Every target uses its exact local `image_id` and binds
`control_commit` plus `control_bundle_sha256` to the top-level control pair.
Its v3 identity/provenance binding subset contains exactly these eight labels:
`hub.qa.target`, `hub.qa.base-image-id`, `hub.qa.control-commit`,
`hub.qa.control-bundle-sha256`, `hub.qa.source-commit`,
`hub.qa.source-policy-sha256`, `hub.qa.source-tree-manifest-sha256`, and
`hub.qa.build-manifest-sha256`. The separate ninth lifecycle/status label is
`hub.qa.runtime-status`. Attestor and runner equality for the full labels
projection includes all nine labels; the v3 binding subset remains exactly the
eight labels above. All required selected images must already exist locally as
`linux/amd64` image objects. A missing local base, target, platform, identity
match, or required dependency byte fails closed before staging, execution, or
publication. There is no registry login, pull, push, manifest lookup, remote
provenance lookup, remote cache, network download, package installation,
source bind, or fallback image acquisition.

## Compose boundary

The six control blobs above are the authoritative v3 private contract inputs.
The local Compose definition is an immutable raw-byte-pinned input and is
rechecked against `control.compose_sha256` before staging, execution, or
publication; no YAML substring parse or transformed representation substitutes
for that byte identity.

After all request, path, v3-lock, and raw Compose byte checks, the candidate
creates one nonce-bound reservation with create-new semantics. Its staging
directory may contain only `.owner.json` and `compose.unit.yaml`; both are
re-snapshotted before one atomic final-directory publication. The owned `raw/`
directory is created only after publication and verified reservation release,
before dependency construction. A collision or ambiguity retains every
reservation, staging, or final-report object and starts no dependency; there is
no adoption, recursive deletion, wildcard cleanup, or automatic janitor.

G0-B remains an inert, no-source-bind template and is `NOT_VERIFIED`; it
neither consumes the absent lock nor selects or runs an image. Once separately
accepted, the local Compose contract may use `pull_policy: never`, may bind
each service only to the exact locked local target `image_id`, and may mount
only the task-owned report path. It must contain no repository or source bind,
checkout, working-tree read, Git metadata inference, runtime package download,
or runtime installation.

## Evidence, ownership, and deadlines

After successful pre-mutation validation, a run has an exclusive task-owned
report root and owner marker. It can record the owner marker, final receipt,
fixed-command receipts, fixed-command output, target artifacts, and cleanup
outcome. Reports must not contain host environment dumps or secret values.
Raw Compose output is confined to its owned raw artifacts and is not imported
into structured receipts.

The candidate keeps fixed limits: an image-baked command bound of 180 seconds,
an image-baked target bound of 780 seconds, a private image-inspection bound of
30 seconds, a private outer fixed-target bound of 810 seconds, and one
aggregate owned-resource cleanup window of 30 seconds. The first observed
deadline freezes exit `124` for its layer; there is no retry or post-deadline
grace period.

Cleanup is a single exact owned-resource route: initial inventory, typed exact
owned/foreign selection, exact cleanup only when the complete initial inventory
is clean, then residual zero-owned proof. Foreign records are preserved.
Recursive deletion, wildcard deletion, pruning, automatic staging janitors,
and deletion of a final report are forbidden.

## Acceptance boundary

The local successor is not accepted until an independent review and separately
authorized Docker-only local-cache verification prove exact three-image local
identity, v3 ControlCommit and control-bundle equality, raw Compose equality,
complete nine-label full-projection parity (including `hub.qa.runtime-status`)
with its exact eight-label v3 binding subset, no registry/network path, no
repository/source bind, and the bounded ownership/cleanup route. Until then
the exact status is:
package and launcher `NOT_READY`; private contract
`PRIVATE_CONTRACT_NOT_RUN`; runtime `NOT_RUN`; and G0-B parity
`G0-B_PARITY_NOT_VERIFIED`. No static review promotes those statuses.
