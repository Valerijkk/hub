# Consolidated QA container assets

This directory owns immutable execution assets consumed by the root Compose
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
profiles dispatch only to literal pytest or Locust commands.

## `backend-unit` low-disk execution

Run the backend unit gate only through the canonical root-Compose command:

```text
docker compose -f docker-compose.local.yml run --rm backend-unit
```

The service uses a **partial source overlay**: these six current checkout roots
are read-only binds at matching `/opt/hub/source/backend/**` targets:
`backend/cmd`, `backend/internal`, `backend/tests`, `backend/ent/schema`,
`backend/migrations`, and `backend/testframework`. It does not bind the
`backend/` or `backend/ent/` root; the image-generated Ent packages therefore
remain authoritative. `/reports` remains the host bind for compact receipts and
logs, while the image still supplies the resolved Go module cache.

For each disposable run, `GOCACHE` uses a separate 3 GiB `noexec` tmpfs at
`/var/cache/hub-go`; `GOTMPDIR` uses a separate executable 4 GiB tmpfs at
`/var/tmp/hub-go`. Both are owned by UID/GID 10001 with mode 1770.
`backend-unit` alone has a 6.0 CPU and 10 GiB memory ceiling; tmpfs ceilings
remain subject to that cgroup memory limit.

This avoids an image rebuild for ordinary source and unit-test edits within the
six overlaid roots. `backend/go.mod`, `backend/go.sum`, generated Ent output
(including a schema change requiring regeneration), and edits outside those
roots require an explicitly, separately authorized image rebuild; the normal
command never adds `--build`.

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
current `testframework/src` and `testframework/selftests` at
`/opt/hub/testframework/src` and `/opt/hub/testframework/selftests`. It also
binds the root Compose file at `/opt/hub/docker-compose.local.yml` and the
Phase 130 plan at `/opt/hub/.planning/phases/130-tf-traceability-gates/PLAN.md`.
These are the exact image targets used by the framework architecture selftests.
The resolved virtual environment, package metadata/lock, policies, schemas,
Docker assets, and other image-owned configuration remain in the image; no SUT
product source is mounted.

An explicit, separately authorized rebuild is required for
`testframework/pyproject.toml`, `testframework/uv.lock`,
`testframework/.python-version`, `testframework/policies/**`,
`testframework/schemas/**`, `testframework/docker/**`, or another image-owned
configuration change. Ordinary edits under `src` or `selftests` do not add
`--build`. On 2026-08-24, the canonical `python-contract` command executed
this exact partial overlay from an existing local image (no build, pull, or
package install), exited 0, and passed 182 framework selftests in 9.53s. This
is limited to the framework contract target and topology; it does not imply
product-SUT, deployed-runtime, API/integration, or Jitsi evidence.

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
