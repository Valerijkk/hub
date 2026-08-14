# Hub testkit package contract

This is a documentation-only architecture contract for the future Python
package bootstrap. It does not install dependencies, create executable suites,
define a runnable command, or grant runtime authority.

## Current truth

| Surface | Status | Evidence |
|---|---|---|
| Structure and ownership boundaries | `COMMITTED` | commit `22035c45` |
| Installable `hub_testkit` package bootstrap | `NOT_READY` | package metadata, Python pin, typed marker, and package-contract selftests are source-authored; `uv.lock`, installation, build, and Docker evidence remain absent |
| Framework selftests | `NOT_READY` | package-contract selftest source exists under the registered architecture profile root; it is not executed evidence |
| SUT suite layer directories | `NOT_READY` | only `tests/README.md` defines the boundary; `tests/{contract,api,integration,e2e}/` are planned and absent |
| Versioned launcher API | `NOT_READY` | `runner/run-task.ps1` is a private static source candidate only; `PS5.1_PARSE_NOT_RUN`, `PRIVATE_CONTRACT_NOT_RUN`, no accepted public entrypoint |
| Runtime | `NOT_RUN` | structure and documentation are not execution evidence |

The status values above are independent. A committed directory tree does not
make the package, launcher, suites, or runtime ready.

## Founder maximum-coverage source-authoring authority (2026-08-15)

The founder has authorized the F0--F3 source waves in
`.planning/active/august-2026/07-TESTFRAMEWORK-MAX-COVERAGE-IMPLEMENTATION.md`.
That authority replaces the former `G129-C`/`G130-H`/per-feature activation
stops for authoring and independent source review only. It does **not** change
the table above: package execution, launcher execution, and runtime remain
`NOT_READY` / `NOT_RUN` until one exact revision has separately accepted A1
versioned-runner admission, R Docker resolver admission, immutable image
evidence, generated `uv.lock`, and Docker-only evidence.

Therefore source review may say that a bootstrap artifact is authored or
closed; it must never call it installed, resolved, runnable, PASS, or green.
The current task must not produce a lock, image, report, test execution, or
runtime claim.

## Physical ownership

The repository contains one Python project directly in `testframework/`, with
package code and executable suites on different roots:

| Path | Sole responsibility |
|---|---|
| `src/hub_testkit/**` | reusable installable framework modules imported as `hub_testkit.*` |
| Planned `tests/{contract,api,integration,e2e}/**` | future executable SUT pytest suites grouped by layer and domain; layer directories are currently absent |
| `selftests/**` | contracts of the framework/package itself; never SUT behavior |
| `cases/**` | repository-owned new-case intent and risk linkage; IDs match `^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$` |
| `schemas/**` | closed machine-readable case/run/ownership contracts |
| `policies/**` | marker, environment, redaction, and profile policy data |
| `load/**` | separate bounded Locust lifecycle |
| `docker/**` | immutable execution assets and container contracts only |
| `runner/**` | sole reviewed versioned launcher boundary; current private source candidate is static-only and not runtime authority |
| `reports/<task-run>/**` | generated owner-marked evidence for exactly one run |

The suite tree is progressive. `tests/README.md` is currently its only
boundary artifact; `tests/contract/`, `tests/api/`, `tests/integration/`, and
`tests/e2e/` do not exist yet and are created only by their first accepted
consumer. This documentation slice must not add empty layer directories.

Direct root `core/`, `api/`, `integration/`, `contract/`, or `e2e/` package or
suite trees are not target architecture. A nested `testframework/python/`
project, a second import namespace, and a generic helper warehouse are
forbidden.

`TC-HEALTH-0001` is a canonical new record. Old documentation examples such as
`TC-AUTH-03` and inventory identifiers such as `HUB-TC-*` are historical only;
they do not satisfy the new schema and may not be used for new
`testframework/cases/**` files.

Each case is strict UTF-8 Markdown with exactly one leading TOML frontmatter
block delimited by `+++` lines and parsed with CPython 3.12 `tomllib`. YAML
frontmatter and a separate parser dependency are forbidden. The loaded mapping
is validated by the closed JSON Schema. `automation` is a closed tagged object:
active pytest-layer cases bind one unique repo-relative pytest node ID; load
cases bind a Locust entrypoint; manual cases bind an evidence owner. Fields for
another automation kind are forbidden, and pytest collection proves a
bijection between active case, node ID, and sole `case_id` marker.

## Bootstrap source and deferred resolver artifact

The source bootstrap artifacts below are authored under F0-01. Their presence
does not mean that dependencies are resolved, the package is installed, or any
selftest has run. `uv.lock` remains a separate resolver-owned artifact.

### `pyproject.toml`

The root `testframework/pyproject.toml` is the single source authority for:

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

Its fixed bootstrap contract is:

- `[build-system]` uses Hatchling only;
- `requires-python = ">=3.12,<3.13"`, and `.python-version` is exactly
  `3.12.13`;
- project runtime dependencies are `pydantic` and `jsonschema`;
- dependency group `framework` contains `pytest`, `pytest-timeout`, and
  `pytest-randomly`;
- dependency group `quality` contains `ruff`, `mypy`, and `import-linter`;
- dependency group `http` contains `httpx`; `contract` contains
  `openapi-spec-validator`; `browser` contains `playwright` and
  `pytest-playwright`; and `load` contains `locust`;
- browser and load dependencies are not bootstrap imports and are consumed
  only by their F3 source slices.

Version resolution, transitive hashes, and the final exact dependency set are
not invented in this document. They become authoritative only when R generates
the lock in a task-owned Docker resolver.

### `uv.lock`

The root `testframework/uv.lock` will be the sole exact Python dependency
resolution derived from `pyproject.toml`. It is generated, reviewed, and
committed only after R Docker resolver admission; before that it may and should
remain absent. It is never hand-maintained as an independent dependency list.
A lock that is stale relative to project metadata fails package acceptance.

### `src/hub_testkit/py.typed`

The PEP 561 marker is present with a typed public package boundary and a
source-authored selftest that will verify its installed distribution location.
Neither file presence nor source review is proof that the package is installed
or type-safe.

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

The source-authoring package slice starts with negative contracts in
`selftests/**`, then adds the minimum package artifacts needed to satisfy them.
At minimum, source review must establish that the future accepted slice can
prove:

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
direct command is defined by this document. F0 also adds source-level negative
contracts for strict config/environment admission, nonce-bound ownership and
cleanup, recursive idempotent secret redaction, closed test-case schema links,
and public-hook-only collection behavior.

## Launcher and container boundary

`runner/**` owns the versioned launcher API, profile resolution,
bounded orchestration, exit preservation, task/run ownership, and exact cleanup
flow. `docker/**` owns only immutable images, topologies, and resource/container
contracts consumed by that launcher. Neither directory may expose a second or
implicit fallback. The current private source candidate has only static review;
until binding, parser/private-contract and Docker evidence are independently
accepted, launcher status remains `NOT_READY` and runtime remains `NOT_RUN`.

## Source closure and package-bootstrap acceptance

F0 source closure requires the exact Python/build/group contract above,
`src/hub_testkit/py.typed`, framework selftests, closed schema/policies, the
literal 47-profile registry, public-hook collection policy, and
`TC-HEALTH-0001` as a source tracer for implemented `GET /api/live`. It also
requires independent package, traceability, architecture, API-contract, and
security review. It cannot yield a runtime result.

The installable/executable package slice is not accepted until all of the
following are true on one exact revision:

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

Until A1, R, immutable image evidence, the generated lock, and Docker-only
evidence exist and are independently verified, the only valid package verdict
is `NOT_READY` and the only valid runtime statement is `NOT_RUN`.
