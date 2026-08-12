# Private QA runner boundary

## Status

This is documentation for a candidate private implementation. The package and
launcher are `NOT_READY`; PowerShell parsing is `PS5.1_PARSE_NOT_RUN`, private
contract execution is `PRIVATE_CONTRACT_NOT_RUN`, and runtime is `NOT_RUN`.
Static inspection of this document or its sibling source cannot promote any
of those statuses.

## Authority boundary

`/qa-gate` is the sole public execution authority. Its private implementation
detail is `testframework/runner/run-task.ps1`, which is not a supported direct
command or alternate entrypoint.

The candidate contract literal is `hub.qa-runner/v1`. It has exactly three
fixed target identities:

- `backend-unit`
- `frontend-unit`
- `backend-openapi`

Declared future Python jobs remain outside this candidate and return
`NOT_READY` without entering the private runner. Sentinel, selftest, fixture,
unknown, and caller-selected target names are not normal targets.

## Closed request and dependency surfaces

The private request contains only these fields:

- `Contract`
- `Target`
- `Task`
- `Run`

It cannot carry an executable, path, argument vector, image or Compose
override, timeout profile, dependency object, resource inventory, fixture,
selftest, arbitrary command, or selector.

The internal dependency seam has exactly four semantic operations:

1. `InspectLockedImage`
2. `RunFixedTarget`
3. `ListOwnedResources`
4. `CleanupOwnedResources`

The frozen candidate contract requires exactly one internal mutable
`PSCustomObject` context with these properties, in this order: `Images`,
`Project`, `Container`, `RepositoryRoot`, `ComposePath`, `ReportPath`,
`SourceSha`, and `ComposeEnvironment`. The contract permits only the four
dependency closures above to consume it. It remains a private implementation
detail, not a request field or caller-configurable surface.

`C1 Cycle-6 scoped static review: STATIC_SCOPED_REVIEWED; whole-slice review:
PENDING`, pinned to Cycle-6 source
`D0932A7943367DFB07660B0937E79C1A90EB347D40FCFB892D280778F2B433DB` and
Cycle-6 package
`8E00ADFA4E5C54AD023B76C267F6CA2DBBE34573C6ED0875AD9811EF0ECB32CA`.

All identities for the project, container, report root, Compose definition,
image lock, and executable are derived internally from the canonical
repository, fixed target, and validated task/run identifiers. Callers cannot
override any of them.

There is one physical `System.Diagnostics.Process` construction primitive. It
is reachable only through those four closed dependency operations; it is not a
claim that `RunFixedTarget` is the only process-construction site. A normal
run has at most ten Docker CLI calls: one inspect, one fixed target, three
initial inventory calls, at most two exact-ID cleanup calls, and three
residual-inventory calls.

## Validation, ownership, and evidence

The private supervisor completes duplicate-aware image-lock validation before
creating a report directory, owner marker, or dependency call. The lock has
only the three fixed target keys and immutable digest references. Invalid,
duplicate, missing, malformed, escaped, or reparsed input fails before report
mutation.

After that pre-mutation boundary, a run receives an exclusive task-owned
report root and owner marker. Its descriptive evidence shape is limited to:

- `.owner.json`
- `run.json`
- `commands.json`, the actual ordered per-command receipt sequence, which may
  stop at the first failed or timed-out command
- fixed-command stdout, stderr, and result artifacts
- target artifacts
- `cleanup.json`

Each command artifact uses the closed candidate convention
`commands/<command-id>.stdout.log`, `commands/<command-id>.stderr.log`, and
`commands/<command-id>.result.json`. This documents the private contract only;
it does not assert that separately owned image-entrypoint parity was executed.

`G0-B baked-entrypoint filename/parity: NOT_VERIFIED` until an independent G0-B
review supplies that compatibility evidence.

`run.json` is the final run receipt: it records the actual contract,
identities, source revision, selected immutable image identity, deadline and
limit configuration, final artifact manifest, immutable first exit, final
outcome, result counts where applicable, relative artifact paths, and cleanup
outcome. Its `commands` field is the fixed expected command identifier list,
not receipt objects; the actual receipt objects belong to `commands.json`. It
must not contain a host environment dump or secret value. Evidence records
observed status; static inspection cannot convert it into runtime, cleanup,
timeout, isolation, or target-result evidence.

## Bounded supervision

| Layer | Owner | Fixed bound | Meaning |
|---|---|---:|---|
| Fixed command | Digest-baked entrypoint | 180 seconds | Image-owned inner command bound, not supervisor scheduling. |
| Fixed target | Digest-baked entrypoint | 780 seconds | Image-owned inner target bound, not supervisor scheduling. |
| Locked-image inspect/attestation | Private supervisor | 30 seconds | Fixed inspect bound before Compose target start. |
| Fixed Compose target | Private supervisor | 810 seconds | Fixed outer `RunFixedTarget` process bound. |
| Inventory, exact cleanup, and zero proof | Private supervisor | 30 seconds aggregate | One fixed cleanup window for the complete owned-resource route. |

The supervisor validates baked receipt transitions but does not schedule or
terminate an inner command or target. There is no post-deadline grace period
and no retry at any layer. At the first observed process-deadline expiry, the
contract freezes exit `124` for that layer, permits only best-effort termination
of the exact handle, and requires disposal without any further wait, drain,
scheduling, read, progress, EOF, or child-exit observation.

After report ownership, an inspect expiry freezes immutable
`first_exit_code = 124` with `deadline_layer = inspect` and skips
`RunFixedTarget`, Compose target start, and Compose-environment start. It does
not skip the sole aggregate 30-second cleanup `finally`: the route remains
`ListOwnedResources` -> typed exact owned/foreign selection ->
`CleanupOwnedResources` only when the complete initial inventory is clean ->
`ListOwnedResources` residual zero-owned proof. A dependency-construction
failure is categorical; it is not an inspect expiry and is not represented as
exit `124` or deadline layer `inspect`. Cleanup records its own evidence and
outcome but never overwrites the frozen `124` or any earlier first exit. The
selector preserves every foreign record. Broad cleanup, pruning, and deletion
of foreign resources are forbidden.

## Supply-chain and host boundary

Only immutable, approved image identities may be used by the future private
implementation. Runtime pulls, downloads, installation, host-environment
forwarding, mutable image selection, and secret-bearing reports are forbidden.
The container contract and private runner do not create a public fallback or a
separate execution route.

Any future runtime evidence requires the public authority, approved immutable
assets, task-owned evidence, and independent acceptance. Until then this
boundary remains package `NOT_READY`, launcher `NOT_READY`, PowerShell parsing
`PS5.1_PARSE_NOT_RUN`, private contract `PRIVATE_CONTRACT_NOT_RUN`, and runtime
`NOT_RUN`.
