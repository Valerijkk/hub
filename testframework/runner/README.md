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
