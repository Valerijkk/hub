# Backend Unit Hard-2-GiB Runner Review Resolution

**Date:** 2026-09-01
**Applies to:** the base design and normative amendment dated 2026-09-01
**Status:** normative review resolution; written-spec review pending

This document closes the first independent architecture, evidence, and
low-disk review. Its rules take precedence over conflicting text in the base
design or earlier amendment.

## Concurrency preservation

The service retains `cpus: 6.0`. Reducing it to two CPUs would change the
effective Go 1.25 `GOMAXPROCS`, default `go test -parallel` behavior,
`t.Parallel` scheduling, and race exposure. The accepted resource envelope is:

```yaml
cpus: 6.0
mem_limit: 2g
memswap_limit: 2g
```

Only cross-package execution changes: every Go command uses explicit `-p=1`.
The design claims exact package-set and flag preservation, not scheduling
identity with the prior cross-package-parallel launcher.

## Exact package plan

The image-baked launcher runs these literal discovery commands without `eval`,
caller input, filtering, or shell command reconstruction:

```text
go list -f '{{.ImportPath}}\t{{.Name}}' ./...
go list -tags hubtest -f '{{.ImportPath}}\t{{.Name}}' ./...
go list -tags hubtest -f '{{.ImportPath}}\t{{.Name}}' ./tests/...
```

The third command supplies both unit and race plans. Output is UTF-8/LF,
lexically sorted by import path under `LC_ALL=C`, and must be nonempty and
duplicate-free. Each row has exactly one tab, a nonempty import path accepted
as one quoted argv element, and a package name accepted by the Go tool. Shell
metacharacters are data and are never evaluated. A malformed row, duplicate,
list failure, empty plan, omitted result, extra result, or count/hash mismatch
is fail-closed.

Discovery failure is `SEALED_EXECUTABLE_FAIL_DISCOVERY` unless cgroup,
filesystem, signal, or observer evidence attributes it to infrastructure. It
enters normal test triage and is not automatically called an infrastructure
abort.

Each planned package appears exactly once as a stage target. Build targets use
one collision-safe output each, including non-main packages:

```text
$GOTMPDIR/bin/<six-digit-ordinal>-<sha256-of-import-path>
```

The ordinal comes from the frozen sorted plan. The digest is lowercase SHA-256
of the exact import-path bytes. The launcher creates `GOTMPDIR/bin` only after
proving that `GOTMPDIR` is the dedicated expected tmpfs mount root, is not a
symlink, and is owned by the runner UID. It invokes:

```text
go build -p=1 -o <owned-output> <quoted-import-path>
go vet -tags hubtest -p=1 <quoted-import-path>
go test -tags hubtest -json -p=1 -count=1 -timeout=150s <quoted-import-path>
go test -tags hubtest -json -race -p=1 -count=1 -timeout=150s <quoted-import-path>
```

Build outputs and Go temporary files are ephemeral work, never evidence. They
are excluded from the report manifest and disappear with the disposable
container.

## Resource revalidation and cache containment

The launcher reads and records the effective cgroup family, limit files, swap
files, and memory failure counters before discovery, before every Go package,
and during finalization.

For cgroup v1 it requires exact memory and memsw limits of `2147483648`, records
`memory.failcnt` plus available OOM-control fields, and treats any fail-count
increase as resource-contract abort. For cgroup v2 it requires
`memory.max=2147483648` and `memory.swap.max=0`, records `memory.events`, and
treats any `oom` or `oom_kill` increase as resource-contract abort. A limit
change, unreadable counter, unsupported layout, or inconsistent family is
`INFRA_ABORTED_RESOURCE_CONTRACT`.

Before cleanup the runner proves that `GOCACHE` and `GOTMPDIR` are their exact
dedicated tmpfs mount roots, are not symlinks, resolve under no other path, and
are owned by the runner UID. `go clean -cache` is the only cache cleanup. It
runs only after the owned Go child exits and only when GOCACHE usage is at least
the frozen 512 MiB watermark. Cleanup failure is infrastructure abort. No
recursive source/report deletion, reparse-point following, Docker cleanup,
volume cleanup, or host-cache cleanup exists.

Eviction can materially increase execution time and the probability of an
unchanged 150-second per-package test timeout. Such a timeout is sealed as
`TIMEOUT_UNTRIAGED` and diagnosed before source or infrastructure attribution.
The six-hour whole-run deadline is an infrastructure stop, not source failure.

## Source mutation lease

An accepted run holds a root-owned exclusive mutation lease for this worktree.
No agent, formatter, generator, editor, or user process may write the mounted
source roots until the observer releases the lease. Root records clean Git
status, HEAD, frozen subject hashes, and the mounted-source manifest hash before
start and after termination. Any write observation, dirty state, hash change,
HEAD change, lease uncertainty, or failed post-check yields
`SOURCE_MUTATED / INVALID_EVIDENCE`.

The manifest hashes the exact regular files under the read-only mounted roots
once before and once after the run in lexical relative-path order. It excludes
`.git`, reports, caches, generated container-only Ent output, and every path not
mounted into the container. Hashing is admitted only after the existing disk
preflight and is itself subject to the disk-pressure stop rule.

## Bounded report budget

The host-bound report root has an exact 64 MiB per-run budget. The runner checks
the aggregate regular-file byte count after every package without recursively
walking cache/tmp. A single package may exceed the budget only in the still
owned `.partial` directory; after detection the runner stops before the next
package and records `INFRA_ABORTED_REPORT_BUDGET` when possible. It never
truncates a log and never seals an incomplete log set.

The bounded five-artifact-per-stage shape in the normative amendment remains.
Cache, tmp, and build outputs are never children of the report directory.

## Observer and final seal

The container writes an owner marker, package plans, stage artifacts, and a
container-terminal manifest only inside its `.partial` directory. It may write
`CONTAINER_COMPLETE`, but it does not publish the final `SEAL` and does not
rename the directory to final.

A narrow host observer is admitted as evidence infrastructure, not a test
runner and not a second public QA API. It starts before the unchanged canonical
Compose command and may perform only:

- read-only host RAM, C: free-space, disk/queue, Git, Docker event, stats, ps,
  and inspect operations;
- writes under the uniquely owner-bound `.partial` report root;
- exactly one `docker stop --timeout 5 <owned-container>` when two consecutive
  free-RAM samples are below 2 GiB, prohibited QA overlap appears, the accepted
  disk-pressure predicate holds, the six-hour deadline expires, or the user
  interrupts the owned run.

The observer cannot run Compose, start/create/update/build/pull/prune/remove a
container or image, stop any foreign container, execute inside a container,
change Git/source, accept arbitrary commands, or select more than one candidate.
Its state machine binds exactly one container through Compose project/service
labels plus the container ID in the owner marker. It captures creation-time
HostConfig and start/die/destroy events before `--rm` removes inspection state.

Only the observer validates the container manifest and exhaustive artifact
hashes, writes the observer receipt and final receipt, writes detached `SEAL`,
and atomically renames `.partial` to final. `SEAL` states integrity and complete
binding, not hostile-host immutability or success.

If OOM/SIGKILL occurs and the observer has complete engine evidence, it may
seal `SEALED_INFRA_ABORTED_OOM`. If container/event continuity or required
metadata is missing, the directory remains `UNSEALED_ORPHAN /
NONAUTHORITATIVE`. An unsealed packet alone is never assigned a trusted OOM,
source, or GREEN verdict.

## Closed receipt contracts

Implementation adds three Draft 2020-12, `additionalProperties: false`
contracts plus valid/invalid fixtures and Python selftests:

1. `testframework/schemas/backend-unit-package-plan.schema.v1.json` for exact
   discovery rows, ordering, set/hash/count, and target-result partition;
2. `testframework/schemas/backend-unit-container-receipt.schema.v1.json` for
   owner, cgroup baselines/finals, stage artifacts, package results, terminal
   state, and exhaustive container artifact hashes;
3. `testframework/schemas/backend-unit-observer-receipt.schema.v1.json` for
   Git/source/image/Compose/container binding, creation-time limits,
   concurrency/resources, lifecycle events, stop authority, final
   classification, and detached seal hash.

The observer protocol may be implemented under
`testframework/observers/backend_unit/`; it cannot live under or revive the
retired `testframework/runner/**`. The root policy and selftests must prove its
non-launching command allowlist and reject Compose execution, Docker mutation
outside the one guarded stop, arbitrary command construction, path escape,
symlinks, stale ownership, and foreign-container selection.

## Image rebuild topology

TDD begins with the changed Python selftests and schemas and one expected
`python-contract` RED. The implementation changes image-owned Docker assets.
Executable static and backend GREEN therefore require one admitted, monitored
local rebuild covering both Compose services:

```text
docker compose -f docker-compose.local.yml build python-contract backend-unit
```

No run adds `--build` or `--pull`. The rebuilt `python-contract` image supplies
the current Docker assets to architecture selftests; the rebuilt
`backend-unit` image supplies the exact launcher. Both image IDs and input
hashes enter the observer receipt. This single rebuild is scheduled only after
the implementation atom is complete and no Mail/Calendar build or QA runner is
active.

## Resolved status and scope

The first use is a candidate for CALL-DUR-08C2 executable GREEN, not an
already accepted GREEN. Until the full canonical receipt passes, 08C2 remains
`STATIC_ACCEPTED` and 08C3 remains blocked.

Implementation-plan ownership expands to the exact QA-only paths required by
the closed contracts:

- root Compose, backend launcher, and affected Dockerfile image stages;
- root Compose/launcher architecture selftest;
- the three schemas, schema README, fixtures, and schema selftests;
- the narrow observer implementation and its allowlist/security selftests;
- Docker/testframework/testing/troubleshooting documentation;
- artifact cleanup policy for owner-bound partial/final receipts;
- Phase-150 STATE/evidence only after observed runs.

Product source, the frozen 08C2 oracle, generated host Ent output,
Mail/Calendar, and `pro-qa-ai/**` remain outside ownership.
