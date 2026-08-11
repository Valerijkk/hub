# Hub testkit package contract

This is a documentation-only architecture contract for the future Python
package bootstrap. It does not install dependencies, create executable suites,
define a runnable command, or grant runtime authority.

## Current truth

| Surface | Status | Evidence |
|---|---|---|
| Structure and ownership boundaries | `COMMITTED` | commit `22035c45` |
| Installable `hub_testkit` package bootstrap | `NOT_READY` | `pyproject.toml`, `uv.lock` and `py.typed` are intentionally absent |
| Framework selftests | NOT_READY | only selftests boundary documentation is accepted; no accepted framework-selftest suite or package-bootstrap evidence exists. Untracked or other-task artifacts are outside this docs-only contract and cannot promote package, launcher, or runtime status |
| SUT suite layer directories | `NOT_READY` | only `tests/README.md` defines the boundary; `tests/{contract,api,integration,e2e}/` are planned and absent |
| Versioned launcher API | `NOT_READY` | future owner is `runner/**`; no entrypoint exists |
| Runtime | `NOT_RUN` | structure and documentation are not execution evidence |

The status values above are independent. A committed directory tree does not
make the package, launcher, suites, or runtime ready.

## Physical ownership

The repository contains one Python project directly in `testframework/`, with
package code and executable suites on different roots:

| Path | Sole responsibility |
|---|---|
| `src/hub_testkit/**` | reusable installable framework modules imported as `hub_testkit.*` |
| Planned `tests/{contract,api,integration,e2e}/**` | future executable SUT pytest suites grouped by layer and domain; layer directories are currently absent |
| `selftests/**` | contracts of the framework/package itself; never SUT behavior |
| `cases/**` | repository-owned `TC-*` intent and risk linkage |
| `schemas/**` | closed machine-readable case/run/ownership contracts |
| `policies/**` | marker, environment, redaction, and profile policy data |
| `load/**` | separate bounded Locust lifecycle |
| `docker/**` | immutable execution assets and container contracts only |
| `runner/**` | future reviewed versioned launcher API and entrypoints |
| `reports/<task-run>/**` | generated owner-marked evidence for exactly one run |

The suite tree is progressive. `tests/README.md` is currently its only
boundary artifact; `tests/contract/`, `tests/api/`, `tests/integration/`, and
`tests/e2e/` do not exist yet and are created only by their first accepted
consumer. This documentation slice must not add empty layer directories.

Direct root `core/`, `api/`, `integration/`, `contract/`, or `e2e/` package or
suite trees are not target architecture. A nested `testframework/python/`
project, a second import namespace, and a generic helper warehouse are
forbidden.

## Future bootstrap artifacts

The following artifacts are one future implementation slice. They are named
here to reserve authority, not to claim that they currently exist.

### `pyproject.toml`

The root `testframework/pyproject.toml` will be the single authority for:

- project and supported-Python metadata;
- `src/` package discovery for the sole import namespace `hub_testkit`;
- build-system metadata and inclusion of `hub_testkit/py.typed`;
- runtime, test, browser, load, lint, and type-check dependency groups without
  duplicating a second requirements authority;
- root pytest discovery/configuration, registered markers, and strict handling
  of unknown markers;
- tool configuration that genuinely belongs to this standalone Python project.

It must not add Hub application dependencies to the reusable package merely to
make suite collection convenient. Optional execution concerns remain explicit
dependency groups rather than ambient machine requirements.

### `uv.lock`

The root `testframework/uv.lock` will be the sole exact Python dependency
resolution derived from `pyproject.toml`. It is generated, reviewed, and
committed with the bootstrap slice; it is never hand-maintained as an
independent dependency list. A lock that is stale relative to project metadata
fails package acceptance.

### `src/hub_testkit/py.typed`

The PEP 561 marker will be added only with typed public package surfaces and a
selftest proving that the built distribution includes it. Its presence alone
must never be treated as proof that the package is type-safe.

### `tests/conftest.py`

If required by the first executable suite, the root suite `conftest.py` is
composition-only: it registers explicit plugins/fixtures and may not grow a
second reusable framework. Reusable implementation belongs under the owning
`hub_testkit` responsibility or domain.

## Import and dependency rules

1. SUT suites import reusable behavior through installed `hub_testkit.*`.
   `sys.path` mutation, direct imports from `src/`, and copied suite-local
   framework modules are forbidden.
2. `hub_testkit` never imports `tests`, `selftests`, `cases`, generated reports,
   or quarantined legacy content.
3. `selftests` may import the installed package and validate declarative
   contracts, but may not call a deployed Hub instance or assert product
   behavior.
4. `tests/**` may consume validated cases/schemas/policies through narrow
   package adapters; it may not implement independent metadata parsers in each
   suite.
5. `load/**` has a separate measured lifecycle. It may reuse an installed
   client from `hub_testkit`, but it is not collected as a pytest layer.
6. Domain facades, builders, models, oracles, pages, and components remain
   below `src/hub_testkit/domains/<domain>/**`. Cross-domain catch-all modules
   are forbidden.

## Required framework selftests for the bootstrap slice

The future package slice starts with failing contracts in `selftests/**`, then
adds the minimum package artifacts needed to satisfy them. At minimum, the
accepted slice must prove:

- clean installed import of `hub_testkit` without repository-root path hacks;
- package discovery includes only the intended namespace;
- distribution metadata and the resolved lock are coherent;
- `py.typed` is present in the built distribution when the marker is claimed;
- forbidden package-to-suite/legacy import directions are absent;
- root suite configuration is composition-only and registered markers are
  closed;
- a framework selftest does not require Hub services, credentials, network
  egress, or generated report state.

These are framework contracts, not a substitute for SUT tests. Their future
execution must use the accepted launcher and evidence boundary; no host or
direct command is defined by this document.

## Launcher and container boundary

The future `runner/**` owns the versioned launcher API, profile resolution,
bounded orchestration, exit preservation, task/run ownership, and exact cleanup
flow. `docker/**` owns only immutable images, topologies, and resource/container
contracts consumed by that launcher. Neither directory may expose a second or
implicit fallback. Until the runner is implemented and independently accepted,
launcher status remains `NOT_READY` and runtime remains `NOT_RUN`.

## Package-bootstrap acceptance

The future implementation slice is not accepted until all of the following are
true on one exact revision:

- `pyproject.toml`, generated `uv.lock`, `py.typed`, and the minimal failing-
  first selftests are reviewed together;
- package and suite roots remain physically distinct and import directions
  satisfy this contract;
- dependency groups and marker/profile authorities are unique, explicit, and
  closed;
- no executable legacy code is copied or ported into the new project;
- the package contains no secret, machine-specific path, mutable global config,
  or hidden network dependency;
- independent verification records package, launcher, and runtime statuses
  separately and does not turn file existence into `PASS`.

Until that slice exists and is independently verified, the only valid package
verdict is `NOT_READY` and the only valid runtime statement is `NOT_RUN`.
