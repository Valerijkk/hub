# Consolidated QA container assets

This directory owns execution assets consumed by the root Compose
contract. It does not own a launcher, host fallback, service discovery, or
runtime authority.

## Image targets

`Dockerfile.qa` defines four public build targets:

- `backend-unit` installs the closed backend unit entrypoint;
- `frontend-unit` installs the closed frontend unit entrypoint;
- `python-tests` installs the fixed Python profile dispatcher;
- `python-lock` provides the isolated maintenance boundary for creating the
  future dependency lock.

All public targets use pinned base digests and UID/GID 10001. Dependency
resolution occurs only during image construction. Backend and frontend test
sequences execute only from their fixed container entrypoints, while Python
profiles dispatch only to literal pytest, Locust or the fixed offline packet
verifier command. Source-bound assets are not image-immutable inputs.

## `backend-unit` low-disk execution

The September 5 Docker-only packet plan supersedes the Windows observer
choreography. Run backend verification only through the exact `backend-unit`
Compose service; inspect the completed packet through `python-contract`.
The backend candidate remains **STATIC_ACCEPTED / EXECUTABLE_NOT_RUN** until
its actual build/vet/unit/race run is accepted. Packet integrity alone cannot
close that gate.

The authoritative candidate topology, resource envelope, rebuild boundary, and
observer sequence are defined once in [hard2g `backend-unit`
candidate](#hard2g-backend-unit-candidate). This overview intentionally does
not duplicate them.

## `frontend-unit` current-source, low-disk execution

Run the frontend unit gate only through the canonical root-Compose command:

```text
docker compose -f docker-compose.local.yml run --rm frontend-unit
```

The service replaces the common report shorthand with an explicit `/reports`
bind, then overlays current `frontend/src` and `frontend/tests` read-only at
the matching `/opt/hub/source/frontend/**` targets. The image continues to
supply the resolved npm dependencies and all non-overlaid frontend inputs.

`HOME` and `TMPDIR` already point under the frontend report directory. The
common 64 MiB `/tmp` tmpfs is retained first; the service mounts
`/reports/frontend-unit/home` as a 512 MiB RAM tmpfs and
`/reports/frontend-unit/tmp` as a 4 GiB RAM tmpfs. Both mounts are
`noexec,nosuid,nodev`, owned by UID/GID 10001, and mode 0700. `frontend-unit`
has a 6.0 CPU and 10 GiB memory ceiling; the tmpfs ceilings remain subject to
that cgroup. On 2026-08-24, the canonical `frontend-unit` command executed
this exact source overlay from an existing local image (no build, pull, or
package install) and exited 0. Its receipt records TypeScript `PASS` and
Vitest `PASS`: 161 files / 419 tests in 45.83s. This is limited to the frontend
unit target and topology; it does not imply deployed-runtime, API/integration,
or Jitsi evidence.

## `python-contract` current-framework source execution

Run the framework contract gate only through the canonical root-Compose
command:

```text
docker compose -f docker-compose.local.yml run --rm python-contract
```

The service uses an explicit writable `/reports` bind and read-only binds for
current `testframework/src`, `selftests`, `observers`, `schemas`,
`policies` and `docker` at their matching `/opt/hub/testframework/` paths. It also
binds the root Compose file at `/opt/hub/docker-compose.local.yml` and the
Phase 130 plan at `/opt/hub/.planning/phases/130-tf-traceability-gates/PLAN.md`.
These are the exact image targets used by the framework architecture selftests.
The resolved virtual environment, package metadata/lock and other non-overlaid
configuration remain in the image; no SUT product source is mounted.
The entrypoint is fixed to `/bin/sh` and the read-only current `python-qa.sh`.
`PYTHONPATH` contains only the two fixed container package roots. Memory and
memory-plus-swap are both capped at 2 GiB; no Docker socket is exposed.

An explicit, separately authorized rebuild is required for
`testframework/pyproject.toml`, `testframework/uv.lock`,
`testframework/.python-version` or another image-owned configuration change.
Changes to overlaid Python helpers, policies, schemas and dispatcher bytes do
not add `--build`; Dockerfile/dependency changes still require an image rebuild
before claiming those image changes are executed. On 2026-08-24, the canonical
`python-contract` command executed
this exact partial overlay from an existing local image (no build, pull, or
package install), exited 0, and passed 182 framework selftests in 9.53s. This
is limited to the framework contract target and topology; it does not imply
product-SUT, deployed-runtime, API/integration, or Jitsi evidence.

### Offline backend packet check

After the owning backend container exits, select its exact report directory:

```text
docker compose -f docker-compose.local.yml run --rm -e HUB_BACKEND_UNIT_RUN_DIR=backend-unit-0123456789ab.partial python-contract backend-unit-receipt
```

Replace the example hostname with the actual task-owned run. There is no
latest-run discovery, arbitrary command or host helper. The fixed verifier
checks schema/semantics, safe paths, exact inventory, artifact sizes/hashes and
terminal marker under a 64 MiB unique-packet budget. It rejects extra arguments
or missing input. `DOCKER_ARTIFACT_PACKET_PASS` means artifact integrity only:
failed and aborted test runs can also have intact packets. It is not a passing
Go gate, host seal, source/image attestation or product-readiness claim.

September 5 evidence: all 17 new packet/launcher/shell selftests pass in Docker.
The latest full framework run is **519 passed / 71 failed** (590 collected,
20.41s); full framework acceptance remains **NOT_ACCEPTED**. The remaining
observer API/fixture and lexical/AST contract failures are recorded in the
[active plan](../../docs/superpowers/plans/2026-09-05-docker-only-backend-packet.md).

## hard2g `backend-unit` candidate

The candidate is Docker-only and **STATIC_ACCEPTED / EXECUTABLE_NOT_RUN**; no
new executable `PASS` is claimed and earlier dated baseline evidence remains
historical. Its fixed container envelope is 2 GiB memory with `MemorySwap` also
2 GiB (no swap), 6 CPU, read-only rootfs, `/tmp` tmpfs 64 MiB, `GOCACHE` tmpfs
768 MiB, `GOTMPDIR` tmpfs 384 MiB, and a required 512 MiB clean threshold.

It binds exactly six current product source roots read-only:
`backend/cmd`, `backend/internal`, `backend/tests`, `backend/ent/schema`,
`backend/migrations`, and `backend/testframework`; `/reports` is the only
writable bind. It does not bind the `backend/` or `backend/ent/` root, so the
image-generated Ent packages and resolved Go module cache remain authoritative.
Ordinary edits within those six roots do not add `--build`; `backend/go.mod`,
`backend/go.sum`, generated Ent output (including a schema change requiring
regeneration), and edits outside the documented source overlays require a
separately authorized rebuild.
Two additional read-only binds supply current runner helpers:
`testframework/docker/backend_unit` at `/usr/local/lib/hub/backend-unit`
and `testframework/docker/entrypoints/backend-unit.sh` at
`/usr/local/bin/backend-unit`. The fixed Compose entrypoint is
`[/bin/sh, /usr/local/bin/backend-unit]`, matching the launcher path hashed
by the receipt. These nine mounts (one report, eight read-only) let reviewed
shell fixes execute without rebuilding the image; they do not attest
image-immutable source identity. The September 5 focused launcher test has
Docker RED/GREEN; its current-launcher/packet regression is 23 passed.
The first actual backend attempt exited before Go on an awk built-in-name
collision. The variable-only fix and correction of SHA-256 helper length
checks have real shell/awk RED/GREEN (53 selected tests pass including
previous launcher/packet regressions), with independent review. Only a new
actual backend run can establish post-fix build/vet/unit/race results.
The subsequent owner-publication attempt exposed rejection of GNU stat's
empty regular-file label. The identity helper now retains the same strict
ownership/mode/type guard and normalizes only that empty label, preserving
identity through write/truncate. Its actual-file Docker RED/GREEN and the
combined 56 selected passing checks do not replace the pending Go gate.
The next backend run produced owner.json and four package plans but aborted
without a completed stage. Its Dash-incompatible kill syntax is fixed in
five calls, with real owned-process RED/GREEN and unchanged identity guards.
The cause of entering abort remains under investigation; plans and packet
integrity are not stage verdicts.
Six fixed stderr abort markers now distinguish child start, stdout/stderr
reader identity, monitor, reader exit and reaping failures without printing
dynamic data or changing cleanup. Their source-contract and the earlier
targeted regression total 63 passing checks. The shell-child diagnostic
confirms possible auto-reaping before explicit wait in its test topology;
it is not proof of the actual Go abort cause.
The old Windows `root_handoff.py` / `observe.py` choreography is superseded
and is not an executable fallback. Historical observer receipts remain
historical; do not fabricate their `SEAL` from an offline packet result.
The current receipt states are `CONTAINER_COMPLETE` and `CONTAINER_ABORTED`.
No automatic stop/cleanup authority is added by the packet verifier. Do not
rename or delete candidate containers or existing reports to obtain a PASS.

## Retained files

The closed retained set is `Dockerfile.qa`, this README, three entrypoints, and
`vitest.frontend-unit.config.ts`. Other files currently present in this
directory are migration inputs awaiting the separate retirement task and are
not alternative authority.

## Readiness

The dated `frontend-unit` and `python-contract` evidence above is limited to
their exact Compose services and source-overlay topologies. Dependency-lock
resolution, the maintenance-service boundary, and every other QA target remain
`NOT_RUN`; no broader build, test, or readiness claim is made here.
