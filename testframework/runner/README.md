# Hub QA runner

## Current status

The versioned public runner contract is `hub.qa-runner/v2`. It is static
source only: package and launcher are `NOT_READY`, private-contract execution
is `PRIVATE_CONTRACT_NOT_RUN`, and all runtime evidence is `NOT_RUN`.

The repository image lock, the three A1 dependency/provenance controls, and
the founder-approved Package-R one-shot authority do not yet exist. A1 remains
`BLOCKED_EXTERNAL_INGRESS`; Package R and the only permitted bootstrap
transaction remain `BLOCKED`. The A0.2/B source present in the tree is only a
static candidate until its independent package gates are recorded. Static
source review is not Docker, image, cleanup, or target acceptance.

## Sole public request

`/qa-gate` is the sole public route. Its current adapter accepts exactly six raw
arguments in this order, with no PowerShell parameter-binding fallback:

```text
-Target <target> -Task <task> -Run <run>
```

Those arguments form exactly these versioned request values:

- `target`
- `task`
- `run`

`task` follows `^[a-z][a-z0-9-]{2,31}$`; `run` follows
`^[a-z0-9][a-z0-9-]{7,63}$`. The public route injects
`hub.qa-runner/v2`; callers cannot provide a contract, path, command,
arguments, image, Compose file, report root, timeout, role claim, owner nonce,
private context, environment override, or bootstrap authority. Missing,
malformed, private, unknown, sentinel, test, or extra input fails closed with
intended exit `64` before a private runner or Docker operation.

The only fixed G0 target identities are `backend-unit`, `frontend-unit`,
`backend-openapi`, and `g0-planning-selftest`. The public route never exposes
`g0-planning-selftest`; it is a bootstrap-private fixture target. Until the
later accepted C1 authority, all ordinary fixed targets return `NOT_READY`.

The declared Python jobs `python-contract`, `python-api`,
`python-integration`, `python-e2e-smoke`, `python-e2e-full`, and `python-load`
remain future `NOT_READY` jobs. They do not enter a private runner or Docker
and no compatibility runner, legacy test route, or direct command is a
fallback.

## Bootstrap is internal only

`hub.qa-bootstrap/v1` is not a public gate target. Its one-time internal
`g0-planning-selftest` closure receives only a non-serialized private context
from the founder-released bootstrap transaction. Direct script invocation,
public planning-selftest dispatch, serialized role transport, continuation,
or duplicate closure use rejects before Docker with the closed direct-call or
one-shot error. The private runner library has no executable file parameter
surface.

When all future gates are accepted, bootstrap derives the fixed task
`g0-assets-runner` and one grammar-valid run solely from the committed
Package-R artifact. It creates one task-owned report root,
`testframework/reports/g0-assets-runner-<run>`, and may run only the internal
fixture. It never runs an ordinary backend, frontend, Python, E2E, load, or
legacy suite during bootstrap.

The separate pre-helper admission context has exactly 28 private fields. Each
A1 dependency asset has exactly 12 ordered fields: `asset_id`, `targets`,
`relative_path`, `sha256`, `bytes`, `archive_format`, `member_count`,
`expanded_bytes`, `layout_sha256`, `license_spdx`,
`license_evidence_sha256`, and `provenance_sha256`. The post-build
image-attestation context has exactly 11 distinct private fields and
binds the owner marker, local-image inspection, and earlier local-base receipt
as raw bytes/digests/nonce. Neither context is a CLI, JSON, environment,
Docker label, mount, report, or public input.

The Package-B candidate imports each accepted private source while its exact
bytes remain held through a read-only/no-write/no-delete Windows handle. After
all imports and before the first admission or attestor entry point, it replaces
and verifies the active `Get-HubQaDockerClientPath`,
`Get-HubQaGitBlobBytes`, `Get-HubQaGitTreeEntry`,
`Get-HubQaGitObjectSourceTree`, and
`Invoke-HubQaG0AdmissionFromBootstrapContext` functions with B-owned wrappers.
Those wrappers reach Git and Docker only through the bounded Job Object process
runner, which creates each child suspended, assigns it to a kill-on-close Job,
and only then resumes it. The historical A0.2/A0.1 helper bodies remain
immutable archive input; their old process snippets are not executable
authority in the bootstrap.

The envelope producer is imported before the snapshot producer; the snapshot
producer has no dot-source, Git, or process path. Bootstrap captures the exact
16-field Git/control snapshot evidence through the bounded native runner before
sealing `command-receipts.v1.json`; captured HEAD must equal the release commit
both before and after the ordered index/worktree capture. After exact cleanup,
it constructs the 24-field snapshot private context entirely from those
captured values and the three retained receipt byte arrays, then calls the pure
snapshot producer.

## Docker-only future execution

All tests, validators, linters, builds, Vite, Playwright, pytest, Locust, and
target processes run only in the task-owned Docker transaction after the
required authority is accepted. Host Windows is limited to reading/editing,
Git, and the approved launcher/Docker CLI boundary; it is not a test or
runtime fallback.

That later transaction is local-only and network-disabled. It uses admitted
offline dependency archives at the fixed
`bootstrap/build-context/<target>/bindings/<relative_path>` topology, then
runs the Git-object source-tree gate before builds. It produces evidence in
one direction: snapshot first, envelope second (raw-linking the snapshot), and
`run.json` third (aggregating the two raw path/digest pairs). There is no hash
cycle, report replacement, source bind, direct Docker/Compose route, registry
operation, runtime installation, or network fallback.

Git and inspection waits are bounded by 30 seconds, an individual command by
180 seconds, the shared cumulative target budget by 780 seconds, cleanup by 30
seconds, and the sole outer budget by 840 seconds including 30 seconds of
orchestration slack. Each child timeout is clipped to the remaining applicable
budgets. Deadline expiry maps to `124`; the first non-zero exit remains
immutable through termination and cleanup.

The outer deadline rejects a missing or stopped runtime Stopwatch and compares
elapsed ticks against the exact 840-second limit. It brackets every final
stage before and after: snapshot Git evidence, command-receipt sealing, exact
cleanup, snapshot context and snapshot, envelope, completed-zero `run.json`,
and the C receipt chain. The intended order remains command receipts first and
exact cleanup second. A final success-only fence runs in `finally` before the
outer Stopwatch is stopped. The Stopwatch is then stopped and its elapsed ticks
are checked once more before runtime authority is cleared; either expiry
becomes exit `124`.

The shared 30-second cleanup Stopwatch accumulates cleanup work only. Helper
cleanup lazily creates, starts, and stops it in `finally`; later exact cleanup
restarts the same Stopwatch and stops it in `finally`. Builds and inspections
between those scopes do not consume the cleanup budget. Helper and exact
cleanup rethrow `E_B0_TIMEOUT_124` immediately; both the success-cleanup and
fallback-cleanup catches preserve it as exit `124` rather than general cleanup
exit `71`. Exact cleanup uses one combined outer-plus-cleanup guard around
resource enumeration and every iteration, owned file/directory deletion,
Docker probe/removal, report-tree removal, and cleanup-receipt creation. A
single unconditional post-delete guard may also close the iteration fence.
Its `finally` stops the cumulative cleanup clock, rechecks the stopped elapsed
ticks against 30 seconds, and then rechecks the still-live outer clock.

Immediately after Package-R validation and before any per-run report root or
runtime resource, bootstrap scans the fixed physical ledger root and writes the
create-new durable receipt
`testframework/reports/g0-assets-runner-nonce-consumptions/<release_id>.nonce-consumption.v1.json`.
Its exact nine fields bind release commit, raw release digest, task/run/release
ID, nonce, consumption time, and `CONSUMED`; cleanup never owns or removes it.

After positive-green and the bounded snapshot-evidence capture, bootstrap writes canonical
`commands/command-receipts.v1.json`. Its exact top-level fields are `schema`,
`producer`, `task`, `run`, `owner_marker_sha256`,
`release_one_shot_nonce`, `commands`, and `created_at_utc`. Each ordered command
entry is exactly `name`, `argv`, `exit_code`, `timed_out`, `started_at_utc`,
`finished_at_utc`, `stdout_sha256`, and `stderr_sha256`. Package C retains the
exact ordered keys `nonce_consumption`, `owner_marker`, `local_base_inspect`,
`helper_base_verification`, `helper_image_attestation`, `ingress_copy`,
`helper_admission`, `helper_cleanup`, `source_tree_verification`,
`local_image_inspection`, `candidate_lock`,
`planning_selftest_negative_red`, `planning_selftest_positive_green`,
`command_receipts`, `run`, and `cleanup`, each as `{path, raw_sha256}`.
Snapshot, envelope, and the ten-field completed-zero `run.json` remain a
one-way chain, and activation binds the same nonce receipt plus `a02_commit`
through C1/D.

## Authority sequence

1. A0.1 archive contract is independently accepted and committed.
2. A0.2 is independently accepted as the exact nine-path unit-runner successor:
   A0.1 must be in its first-parent history, its actual parent must descend from
   source baseline `61b38573682d7c01e6d9381babc5085537d7990f`, and current
   source selection is the exact seven A0.2-overlap plus six remaining A0.1
   blobs.
3. A1 binds a founder-provided offline ingress in exactly three data files and
   remains blocked until those exact local bytes and evidence exist.
4. Package B is independently accepted as static source only.
5. A separately accepted Package-R artifact binds ordered `a0_commit`,
   `a02_commit`, `a1_commit`, and `b_commit` and provides the explicit one-shot
   founder release.
6. Only then can Package C perform the Docker-only bootstrap transaction;
   C1, D, and activation later verify the candidate lock and evidence.

Until all of that exists, report `NOT_READY` / `NOT_RUN` rather than a pass.
