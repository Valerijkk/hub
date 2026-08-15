# Docker Compose Test Framework Consolidation Design

**Date:** 2026-08-15
**Status:** approved architecture; implementation not started
**Runtime truth:** `NOT_RUN`

## Goal

Make the repository-root `docker-compose.local.yml` the only test execution
manifest and give every supported test stage the same command shape:

```text
docker compose run --rm <service>
```

The change removes the unfinished versioned PowerShell/bootstrap runner and
duplicate Compose control plane. It does not remove the greenfield Python
framework, test cases, policies, schemas, or genuine backend/frontend units.
No test, validator, linter, or build may run on the Windows host.

## Preserved Source Boundary

The following are product test assets, not launcher clutter, and remain under
`testframework/`:

- `src/hub_testkit/**` reusable Python package code;
- `tests/{contract,api,integration,e2e}/**` executable pytest suites;
- `cases/**` traceability records;
- `selftests/**` package, architecture, collection, and source-contract tests
  that do not belong solely to the retired runner;
- `policies/**` and `schemas/**` used by the Python framework;
- `load/**`, `pyproject.toml`, the future resolved `uv.lock`, and `py.typed`;
- package and usage documentation after it is rewritten to the Compose API;
- genuine Go units in `backend/tests/**` and Vitest units in
  `frontend/tests/**`.

The current uncommitted project-outbox source slice is protected user work. It
must be independently reviewed and committed before consolidation begins;
consolidation may not silently delete, absorb, or rewrite it.

## Retired Control Plane

The following families are removed after an exact reference and ownership
audit:

- `testframework/runner/**`;
- `testframework/docker/compose.unit.yaml` and
  `testframework/docker/compose.g0-bootstrap.yaml`;
- G0 admission, activation, snapshot, envelope, bootstrap-control, and planning
  selftest code and schemas that have no consumer after the runner removal;
- PowerShell launcher/admission/attestation scripts;
- `testframework/selftests/runner/**` and other tests that validate only the
  retired control plane;
- stale documentation, skills, CI snippets, and policy entries that advertise
  `/qa-gate`, Package A/R/B/C/D/E, or direct use of removed scripts.

Deletion is manifest-driven. Every candidate must be classified as retained,
moved, rewritten, or deleted, and a retained-reference scan must be clean
before its deletion is committed. Dockerfiles, shell entrypoints, and immutable
runtime configuration that are consumed by the new root Compose remain.

## Root Compose Topology

`docker-compose.local.yml` remains the development stack and gains QA services
behind Compose profiles, so ordinary `docker compose up` does not start test
jobs. Public test services align with the repository CI taxonomy:

| Service | Runtime | Source target |
|---|---|---|
| `backend-unit` | Go | `backend/tests/**` |
| `frontend-unit` | Node/Vitest | `frontend/tests/**` |
| `python-contract` | Python/pytest | `testframework/tests/contract/**` |
| `python-api` | Python/pytest | `testframework/tests/api/**` |
| `python-integration` | Python/pytest | `testframework/tests/integration/**` |
| `python-e2e-smoke` | Python/pytest | e2e tests selected by `smoke` |
| `python-e2e-full` | Python/pytest | full e2e profile |
| `python-load` | Python/Locust | `testframework/load/**` |

Go, Node, and Python use separate minimal images. Python services share one
locked image definition through a Compose extension and differ only by an
allowlisted entrypoint target. No all-in-one toolchain image, Docker socket,
nested Compose, dynamically evaluated shell command, host interpreter, or
PowerShell launcher is allowed.

The public commands are deliberately uniform:

```text
docker compose run --rm backend-unit
docker compose run --rm frontend-unit
docker compose run --rm python-contract
docker compose run --rm python-api
docker compose run --rm python-integration
docker compose run --rm python-e2e-smoke
docker compose run --rm python-e2e-full
docker compose run --rm python-load
```

Entrypoints accept no arbitrary command string. Each service has a fixed
command/target, fixed working directory, non-root user, read-only root
filesystem where practical, dropped capabilities, `no-new-privileges`, PID and
memory/CPU limits, bounded health checks, and a wall-clock timeout enforced
inside the container. Test reports are written only to a dedicated repository
report mount.

## Infrastructure Isolation

Unit services have no infrastructure dependencies. Integration and e2e
services depend on QA-only PostgreSQL, MongoDB, Redis, NATS JetStream, and
SeaweedFS services defined in the same root Compose behind QA profiles.

QA infrastructure:

- publishes no host ports;
- does not reuse development named volumes or development databases;
- uses tmpfs or disposable anonymous storage;
- has explicit health checks and bounded resource limits;
- receives no cloud credentials and performs no provider egress;
- uses per-test nonce-bound databases, buckets, streams, consumers, keys, and
  principals recorded by the injected harness delegate;
- is safe for sequential local runs without relying on cleanup as the test
  oracle.

Each test fixture cleans only its exact ledger-owned resources in `finally`.
Infrastructure containers may remain available between local sequential runs,
but their data is disposable and isolated from development. Stopping the root
Compose project removes the disposable infrastructure; no additional custom
cleanup command or runner script is introduced.

## Docker Asset Layout

After consolidation, `testframework/docker/` contains only execution assets:

- one multi-stage `Dockerfile.qa` with separate named Go, Node, and Python final
  targets selected by root Compose;
- fixed POSIX shell entrypoints;
- immutable tool configuration directly copied into an image, such as the
  Vitest config;
- a concise README documenting that root Compose is the sole caller.

Build admission schemas, activation receipts, Git snapshot/envelope producers,
PowerShell scripts, and internal Compose manifests do not remain in this
directory. Reusable Python behavior belongs in `src/hub_testkit/**`, not in a
Docker entrypoint.

## Failure Semantics

- A missing lock, image build failure, unhealthy dependency, invalid target, or
  cleanup failure is a non-zero test-service exit.
- Entrypoints never fall back to a host tool or another test service.
- Secrets, DSNs, bearer tokens, presigned URLs, raw messages, and response
  bodies are redacted from command output and reports.
- A test failure remains distinguishable from infrastructure startup and
  cleanup failure in JUnit/JSON metadata and the container exit status.
- Compose configuration errors fail before any test process starts.
- Until the Docker images, lock, root Compose services, and container-only
  verification are complete, CI remains fail-closed and documentation states
  `NOT_READY` / `NOT_RUN`.

## Migration Sequence

1. Freeze an exact retained/deleted/moved/rewrite manifest and resolve the
   current project-outbox draft without losing user work.
2. Author source-contract tests for the root Compose service list, security
   options, fixed commands, dependency isolation, and absence of retired public
   entrypoints.
3. Create the multi-stage `Dockerfile.qa` and its fixed shell entrypoints.
4. Add the QA services and QA-only infrastructure to
   `docker-compose.local.yml`.
5. Rewrite package docs and CI to the root Compose commands.
6. Delete the retired runner, internal Compose manifests, bootstrap assets, and
   runner-only selftests only after the replacement source contract is green in
   static review.
7. Resolve `uv.lock` and image inputs inside an approved Docker build; no host
   dependency installation.
8. Run the narrow source/contract matrix in Docker, then unit stages, then
   contract/API/integration/e2e/load stages. Record failures honestly and apply
   functional test triage before changing tests.

## Acceptance Criteria

- Root `docker-compose.local.yml` is the sole Compose execution manifest for
  tests.
- All eight public services use the exact `docker compose run --rm <service>`
  interface and fixed targets.
- No tracked documentation or CI path invokes the retired runner, nested
  Compose files, PowerShell launchers, direct host test commands, or Docker
  socket orchestration.
- `testframework/docker/` contains only currently consumed Docker execution
  assets.
- Framework source, cases, policies, schemas, and genuine units are preserved.
- QA data is isolated from development data and no QA service publishes an
  infrastructure port.
- Static source contracts and then Docker-only runtime stages pass before any
  readiness or merge claim.
- No runtime result is inferred from source review; current status remains
  `NOT_RUN`.
