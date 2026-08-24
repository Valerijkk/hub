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

## Retained files

The closed retained set is `Dockerfile.qa`, this README, three entrypoints, and
`vitest.frontend-unit.config.ts`. Other files currently present in this
directory are migration inputs awaiting the separate retirement task and are
not alternative authority.

## Readiness

The root Compose services, dependency lock, and maintenance-service binding are
not part of this source slice. Package bootstrap and container execution remain
`NOT_RUN_BOOTSTRAP`; no build, test, or readiness claim is made here.
