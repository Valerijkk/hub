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

## Retained files

The closed retained set is `Dockerfile.qa`, this README, three entrypoints, and
`vitest.frontend-unit.config.ts`. Other files currently present in this
directory are migration inputs awaiting the separate retirement task and are
not alternative authority.

## Readiness

The root Compose services, dependency lock, and maintenance-service binding are
not part of this source slice. Package bootstrap and container execution remain
`NOT_RUN_BOOTSTRAP`; no build, test, or readiness claim is made here.
