# Backend Unit Hard-2-GiB Runner Design

**Date:** 2026-09-01
**Status:** approved design direction; written-spec review pending
**Runtime truth:** `NOT_RUN`

## Context

Hub permits backend verification only through the repository-root Compose API:

```text
docker compose -f docker-compose.local.yml run --rm backend-unit
```

The current `backend-unit` service overrides the common resource envelope with
`cpus: 6.0` and `mem_limit: 10g`. Its fixed entrypoint executes build, vet,
unit, and race stages sequentially, but its 3 GiB Go-cache tmpfs and 4 GiB Go
temporary tmpfs share the same cgroup with the Go processes.

CALL-DUR-08C2 cannot obtain executable evidence under the founder-required
hard 2 GiB ceiling. Four post-create 2 GiB diagnostics, including variants
with package serialization and reduced Go runtime settings, were OOM-killed or
contained compiler kills. A later ordinary 10 GiB run passed build, vet, and
unit, but root stopped it before race after two consecutive host-free-RAM
samples below 2 GiB. That run is `INFRA_ABORTED_HOST_RAM_FLOOR`, not source
failure and not GREEN.

## Goal

Provide one canonical, fail-closed `backend-unit` service that:

- is capped before start at 2 GiB memory with no swap;
- preserves the complete build, vet, unit, and race package sets;
- preserves in-package test concurrency and race-detector exposure;
- keeps Go cache and temporary writes off the physical host disk;
- produces immutable, per-run evidence that cannot be confused with stale
  flat reports;
- requires no Docker socket, nested container, host test process, selector, or
  alternative public QA command.

The first accepted use is the blocked CALL-DUR-08C2 GREEN. The design is a QA
prerequisite and does not change Calls, Messenger, Jitsi, Mail, or Calendar
product behavior.

## Non-goals

- No focused-only or package-subset substitute for the full backend gate.
- No change to product source or the CALL-DUR-08C2 oracle.
- No new public Compose services or host-side test launcher.
- No Docker socket mount, privileged container, runtime dependency install,
  pull, or implicit image build.
- No mutable shared host Go cache, repository cache directory, Docker prune,
  or broad cleanup.
- No claim that the design fits 2 GiB until the canonical full run proves it.

## Locked public and authority boundary

The sole execution command remains exactly:

```text
docker compose -f docker-compose.local.yml run --rm backend-unit
```

`backend-unit` accepts zero user arguments. Its launcher is copied into the
local image, its bytes are reviewed and hashed before acceptance, and no host
launcher bind may override it. Source overlays remain read-only. Image-baked
generated Ent output remains authoritative. A launcher or Dockerfile change
requires one separately admitted and monitored image build before executable
acceptance; a stale image is `INVALID_EVIDENCE`.

## Resource envelope

The Compose service owns these pre-start controls:

```yaml
cpus: 2.0
mem_limit: 2g
memswap_limit: 2g
```

`memswap_limit == mem_limit` disables container swap. Package-level Go work is
serialized with explicit `-p=1`. The design does not set `GOMAXPROCS`, does not
set `go test -parallel`, and does not reduce the container below two CPUs;
tests therefore retain their existing in-process scheduling and race exposure.

The launcher verifies the effective cgroup before package discovery. On the
current Docker Desktop cgroup-v1 runtime it requires
`memory.limit_in_bytes == memory.memsw.limit_in_bytes == 2147483648`. A future
cgroup-v2 runtime is accepted only when `memory.max == 2147483648` and
`memory.swap.max == 0`. Unsupported, unreadable, unlimited, or mismatched
cgroup state exits before Go with `INFRA_ABORTED_RESOURCE_CONTRACT`.

The existing host observer remains acceptance-only and does not launch tests.
Before execution it requires at least 2 GiB free physical RAM, at least 250 GiB
free on C:, the accepted three-sample disk/queue window, no other
`backend-unit`, no Mail/Calendar QA/test/build, and only an idle BuildKit.
During execution, two consecutive free-host-RAM samples below 2 GiB, prohibited
QA overlap, or the accepted disk-pressure stop condition stops only the owned
task container and yields infrastructure abort.

## Low-memory, low-disk execution

`GOCACHE` and `GOTMPDIR` remain separate container tmpfs mounts. They are not
host binds, Docker volumes, repository directories, or `/reports` children.
Their exact ceilings and eviction watermark are implementation constants that
must satisfy all of these inequalities in the source contract:

```text
GOCACHE tmpfs ceiling <= 768 MiB
GOCACHE eviction watermark <= 512 MiB
GOTMPDIR tmpfs ceiling <= 384 MiB
/tmp tmpfs ceiling == 64 MiB
sum of tmpfs ceilings <= 1216 MiB
```

Tmpfs ceilings are maxima rather than reservations, but every used page counts
against the same 2 GiB cgroup. Cache eviction occurs only after a Go child has
exited and only through `go clean -cache` against the exact task-owned
`GOCACHE`. Temporary cleanup is limited to the validated task-owned
`GOTMPDIR`; it never touches reports, source, another run, a symlink target, a
Docker resource, or a host cache.

The launcher discovers an ordered package manifest for each stage inside the
same source and image context:

| Stage | Discovery scope | Per-package operation |
|---|---|---|
| build | default-tag equivalent of `./...` | `go build -p=1` |
| vet | `hubtest` equivalent of `./...` | `go vet -tags hubtest -p=1` |
| unit | `hubtest` equivalent of `./tests/...` | `go test -tags hubtest -json -p=1 -count=1 -timeout=150s` |
| race | `hubtest` equivalent of `./tests/...` | the unit command plus `-race` |

Every discovered import path executes exactly once as a stage target and is
recorded in deterministic order. The build operation uses a task-owned output
under `GOTMPDIR` when a single main package would otherwise write into the
read-only worktree. Package discovery failure, duplicate or malformed manifest
entry, missing result, or count mismatch fails the stage closed.

After each package exits, the launcher reads tmpfs usage using filesystem
metadata rather than recursively scanning cache files. If the eviction
watermark has been reached, it clears only the Go build cache before the next
package. This deliberately trades additional CPU and elapsed time for bounded
RAM and avoids physical cache writes. Source reads and compact report writes
remain unavoidable and are protected by the existing disk-pressure observer.

An ordinary package failure does not skip later declared packages or later
stages. The launcher retains the first nonzero exit in deterministic
stage/package order while collecting the remaining diagnostics. OOM, resource
contract loss, owned-path uncertainty, or external stop is infrastructure abort
and does not masquerade as a product failure.

## Per-run evidence protocol

Flat `testframework/reports/backend-unit/*` files are retired as acceptance
authority. A run derives a strict identifier from UTC start time plus the
container hostname and creates exactly one new directory:

```text
/reports/backend-unit-<UTC>-<container-id>.partial/
```

Creation fails if the path already exists. An owner marker is written before
package discovery and binds the target, run identifier, container identifier,
launcher SHA-256, start time, and observed cgroup contract. The container never
selects or mutates another report directory.

Each stage publishes:

- the ordered discovered-package manifest and its SHA-256;
- a package-result record for every manifest entry;
- complete stdout and stderr logs;
- stage start/end times, first exit, package totals, and log hashes;
- before/after cgroup counters, including the available OOM/failure counter.

Files are first written with `.tmp` names, closed and hashed, then atomically
renamed inside the same run directory. The overall manifest binds every stage
file, length, digest, result, timing, effective command contract, and final
cgroup counters.

An ordinary completed PASS or FAIL writes the final manifest and `SEALED`
marker, then atomically renames the directory without `.partial`. `SEALED`
means evidence completeness, not success. A caught TERM/INT/HUP records
`INFRA_ABORTED` when safe and possible. SIGKILL, OOM, or an interrupted writer
leaves an unsealed `.partial` directory, which is never acceptance evidence.
No compatibility `latest` pointer has authority.

The root observer separately binds the sealed packet to Git HEAD/status,
frozen source/oracle hashes, image ID/digest, resolved Compose hash, exact
container name/ID, HostConfig limits, resource/concurrency samples, terminal
state, and disposable-container removal. It remains read-only with respect to
Docker execution and cannot change the command, environment, package manifest,
or results.

## Failure classification

| Observation | Classification |
|---|---|
| All stages PASS, sealed packet, exact 2 GiB/no-swap contract, stable source, no prohibited overlap | executable GREEN candidate |
| Complete sealed packet with attributable build/vet/unit/race failure and no infrastructure fault | source/test `FAIL`, then mandatory test triage |
| OOM counter increase, exit 137 with corroboration, host safety stop, timeout, tmpfs exhaustion, unsupported cgroup, or unsealed packet | `INFRA_ABORTED / EXECUTABLE_NOT_RUN` |
| Stale image, launcher/Compose hash mismatch, missing package, foreign owner marker, flat/stale report, post-create cap, or changed command | `INVALID_EVIDENCE / NONAUTHORITATIVE` |
| Mail/Calendar QA/test/build overlap | `INVALID_MAIL_OVERLAP` |

No 10 GiB or post-create-limited receipt may be relabelled as a hard-2-GiB
result. A hard-2-GiB OOM is not evidence against CALL-DUR-08C2 source.

## Verification strategy

Implementation follows TDD and Docker-only execution:

1. Extend the existing Python architecture selftest first with the new Compose,
   launcher, package-set, cgroup, tmpfs, sealing, and negative-mutation
   contract.
2. Run the exact `python-contract` Compose service and retain the expected RED
   caused only by the absent runner contract.
3. Implement the minimal Compose and image-baked launcher changes.
4. Rebuild only the affected local QA image targets in one admitted,
   resource-monitored build window.
5. Run `python-contract` GREEN, then the exact canonical `backend-unit` command
   under continuous RAM/disk/concurrency monitoring.
6. Parse the sealed packet and obtain independent architecture, security,
   evidence, and test-semantics reviews.
7. Only a complete accepted GREEN upgrades CALL-DUR-08C2 from
   `STATIC_ACCEPTED` to `SOURCE_ACCEPTED` and unblocks 08C3.

The semantic oracle must independently compare discovered package sets with
the existing pattern scopes and prove that no package or test is skipped.
CALL-DUR-08C2 additionally requires the already frozen 36 focused PASS and zero
FAIL actions in both unit and race, zero global failure/panic/data-race markers,
and empty stage stderr where the current contract requires it.

## File ownership for implementation planning

Expected implementation scope is limited to:

- `docker-compose.local.yml`;
- `testframework/docker/entrypoints/backend-unit.sh`;
- `testframework/docker/Dockerfile.qa` only if the launcher-copy contract needs
  an explicit hash or build sentinel change;
- `testframework/selftests/architecture/docker/test_root_compose_contract.py`;
- `testframework/docker/README.md`;
- `docs/TESTING.md`, `docs/test_plan.md`, and `docs/TROUBLESHOOTING.md` where the
  executable contract or evidence path changes;
- `.planning/STATE.md` and Phase-150 evidence after observed execution.

Product source, product oracles, generated host Ent output, Mail/Calendar, and
`pro-qa-ai/**` are outside ownership.

## Rejected alternatives

### Keep the 10 GiB service and wait for host RAM

This preserves the current launcher but violates the hard 2 GiB container
requirement and already crossed the host safety floor. It cannot unblock 08C2.

### Change only Compose limits and Go environment

Earlier `-p=1`, `GOMAXPROCS`, and `GOMEMLIMIT` diagnostics still OOM-killed.
This guarantees containment but does not address cumulative tmpfs cache use or
stale flat reports.

### Disk-backed or shared prewarmed cache

This may reduce memory and runtime, but it introduces sustained Docker
Desktop/Windows disk activity, mutable-cache provenance, ownership, cleanup,
and cross-run contamination. It conflicts with the founder's low-disk
constraint.

### Stage services or child containers

They require a new host coordinator or Docker socket, expand authority and
cleanup surfaces, weaken the one-command contract, and do not reduce the peak
of an individual compiler or test process.

### Reduce in-package concurrency

`GOMAXPROCS=1` or `go test -parallel=1` can weaken concurrency and race
coverage. The accepted design serializes only package processes and retains the
current test runtime semantics.

## Acceptance boundary

The design is successful only when the unchanged public command completes the
full package manifest inside a pre-capped 2 GiB/no-swap container, produces a
sealed provenance packet, passes the architecture selftests, and survives
independent semantic/security/evidence review. Until then the runner is
`NOT_RUN` and CALL-DUR-08C2 remains `STATIC_ACCEPTED`.

If a single package still cannot execute inside the resource envelope after
cache isolation, the result is an honest `BLOCKED_INFRA_2GIB_OOM`. The suite is
not sharded away, focused-only evidence is not substituted, and limits are not
raised silently.
