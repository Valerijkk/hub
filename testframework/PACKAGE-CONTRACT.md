# Hub testkit package contract

## 2026-08-23 minimal merge-slice override

This dated section is the controlling runtime statement for the minimal
framework merge slice. Historical source-authoring and deferred-runtime text
below remains as provenance, but does not override these current facts:

- `hub_testkit` is installed from the committed `pyproject.toml` and `uv.lock`
  into the task-owned Docker image;
- `python-contract` runs framework selftests only and passed `181/181` in
  Docker on 2026-08-23;
- the five SUT contract tests and their five OpenAPI/privacy/security cases
  were removed by explicit founder direction; they are not a merge gate;
- API, integration, e2e, load and hosted CI evidence remains `NOT_RUN` and is
  not implied by the framework selftest result;
- PyYAML is permitted only for safe Compose parsing in framework selftests.
  Test-case frontmatter remains TOML and YAML frontmatter remains forbidden.

This accepts the package/framework skeleton gate only. It is not a product,
module, production-readiness, SUT-coverage, or hosted-CI acceptance claim.

This is a documentation-only architecture contract for the future Python
package bootstrap. It does not install dependencies, create executable suites,
define a runnable command, or grant runtime authority.

## Current truth

| Surface | Status | Evidence |
|---|---|---|
| Structure and ownership boundaries | `COMMITTED` | commit `22035c45` |
| Installable `hub_testkit` package bootstrap | `NOT_READY` | package metadata, Python pin, typed marker, and package-contract selftests are source-authored; `uv.lock`, installation, build, and Docker evidence remain absent |
| Framework selftests | `NOT_READY` | package-contract selftest source exists under the registered architecture profile root; it is not executed evidence |
| F0 framework/domain source | `SOURCE_AUTHORED` / `SOURCE_REVIEWED` | strict config/runtime/ownership/redaction, traceability, public pytest collection plugin, literal 47-profile policy, and fake-only health source are authored and independently source-reviewed; this is not execution evidence |
| F1-01 HTTP policy and diagnostics source | `SOURCE_AUTHORED` / `SOURCE_REVIEWED` | commit `a5f3b728`; immutable injected-only HTTP policy, transport, redaction-bound diagnostics, and health adapter source are independently reviewed. They create no client, network request, or runtime evidence |
| F1-02 identity/session source | `SOURCE_AUTHORED` / `STATIC_ACCEPT` | commit `2b037dd2`; four active `TC-IDENTITY-0001`--`TC-IDENTITY-0004` API nodes are independently static-accepted. They require a launcher-injected `IdentityHarnessFactory` to create a fresh harness and ownership ledger per test with exact cleanup; no concrete provisioner or harness implementation exists |
| F1-03 authorization and platform-abuse source | `SOURCE_AUTHORED` / `INDEPENDENTLY_ACCEPTED` | commits `f184487a` and `3932ac78`; four active TASK-208 authorization nodes/cases and two active TASK-209 platform-abuse nodes/cases use launcher-injected opaque fresh factories, per-test nonce-bound ownership ledgers, and exact cleanup. The platform contract records only its four exact Redis ownership entries; concrete provisioners, profile bindings, and a launcher remain absent |
| F1-04 sealed source-contract boundary | `SOURCE_AUTHORED` / `INDEPENDENT STATIC ACCEPT` | commits `bc5feef8`, `630bda9b`, `8ff13097`, and `1868cf71`; the sealed v3 manifest contains exactly 14 bounded source blobs and the closed ledger contains five ordered risk edges. Three active OpenAPI cases, one privacy case, and one auth-event case are source-authored; all critical/high outcomes remain `NOT_PROVEN` |
| F2 135-F files lifecycle source | `SOURCE_AUTHORED` / `INDEPENDENT STATIC ACCEPT` | commit `af077c18`; five active TASK-214 cases/nodes use a launcher-injected `FilesHarnessFactory` to create a fresh opaque `FilesRuntime` per test and require exact cleanup. The existing `files-api` source profile targets `tests/api/files`; its future launcher case root is `cases/files`. No concrete delegate, A1 admission, launcher binding, or runtime evidence exists |
| SUT suite layer directories | `NOT_READY` | API source tracers exist for health, identity, authorization, platform abuse, and files lifecycle. Contract source tracers now exist under `tests/contract/{openapi,privacy,security}/` with five active TASK-227 cases; `tests/{integration,e2e}/` remain absent and no suite execution evidence exists |
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
| `tests/api/health/**` | first source-authored API tracer: one fake-transport liveness contract linked to `TC-HEALTH-0001`; it is not live, installed, or execution evidence |
| `tests/api/identity/**` | four source-authored identity/session API nodes linked to `TC-IDENTITY-0001`--`TC-IDENTITY-0004`; the `identity-api` profile target is exactly `tests/api/identity` and requires exactly `--hub-case-root cases/identity` when a launcher binding exists |
| `tests/api/authz/**` | four source-authored authorization API nodes linked to `TC-AUTHZ-0001`--`TC-AUTHZ-0004` for TASK-208; the future profile root is exactly `tests/api/authz` and its future launcher binding requires exactly `--hub-case-root cases/authz` |
| `tests/api/platform/**` | two source-authored platform-abuse API nodes linked to `TC-PLATFORM-0001`--`TC-PLATFORM-0002` for TASK-209; the future profile root is exactly `tests/api/platform` and its future launcher binding requires exactly `--hub-case-root cases/platform` |
| `tests/api/files/**` | five active TASK-214 files-lifecycle source nodes linked to `TC-FILES-0001`--`TC-FILES-0005`; the existing `files-api` source profile targets exactly `tests/api/files`, with future launcher case root exactly `cases/files` |
| `tests/contract/openapi/**` | three active TASK-227 source contracts linked to `TC-OPENAPI-0001`--`TC-OPENAPI-0003`; source profile `openapi-contract` targets exactly `tests/contract/openapi`, with future launcher case root exactly `cases/openapi` |
| `tests/contract/privacy/**` | one active TASK-227 source contract linked to `TC-PRIVACY-0001`; source profile `privacy-security` targets exactly `tests/contract/privacy`, with future launcher case root exactly `cases/privacy` |
| `tests/contract/security/**` | one active TASK-227 auth-event source contract linked to `TC-SECURITY-0001`; source profile `security-risk` targets exactly `tests/contract/security`, with future launcher case root exactly `cases/security` |
| Planned `tests/{integration,e2e}/**` | future executable SUT pytest suites grouped by layer and domain; these layer directories are currently absent |
| `selftests/**` | contracts of the framework/package itself; never SUT behavior |
| `cases/**` | repository-owned new-case intent and risk linkage; IDs match `^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$` |
| `schemas/**` | closed machine-readable case/run/ownership contracts |
| `policies/**` | marker, environment, redaction, and profile policy data |
| `load/**` | separate bounded Locust lifecycle |
| `docker/**` | immutable execution assets and container contracts only |
| `runner/**` | sole reviewed versioned launcher boundary; current private source candidate is static-only and not runtime authority |
| `reports/<task-run>/**` | generated owner-marked evidence for exactly one run |

The suite tree is progressive. API source tracers exist at
`tests/api/health/test_health_api.py`, linked to `TC-HEALTH-0001`, and at
`tests/api/identity/test_identity_api.py`, linked to four active identity
cases. The health tracer uses a fake transport. The identity tracer has no
concrete provisioner or harness implementation; its future launcher must
inject a fresh `IdentityHarnessFactory` product for each test and perform exact
cleanup. The independently accepted TASK-208 authorization tracer has four
nodes/cases at `tests/api/authz/test_authz_api.py`; its future launcher must
inject an opaque fresh `AuthzHarnessFactory` product, own one nonce-bound ledger
per test, and perform exact cleanup. The independently accepted TASK-209
platform-abuse tracer has two nodes/cases at
`tests/api/platform/test_platform_abuse.py`; its future launcher must inject an
opaque fresh `PlatformAbuseHarnessFactory` product and clean up only its four
exact recorded Redis ownership entries. The future profile roots are exactly
`tests/api/authz` and `tests/api/platform`, with exact case roots `cases/authz`
and `cases/platform`. No concrete authorization or platform-abuse provisioner
exists. `TC-AUTHZ-0005` remains project-cross-scope blocked: current backend
middleware order reaches protected handlers before `rbac.RequirePermission`,
and those handlers terminate the request without `Next`. None of these tracers
proves live Hub behavior. F1-04 adds source-only contract tracers at the exact
future profile targets `tests/contract/openapi`, `tests/contract/privacy`, and
`tests/contract/security`, with exact case roots `cases/openapi`,
`cases/privacy`, and `cases/security`. The closed risk ledger records their
repo-relative case roots as `testframework/cases/{openapi,privacy,security}`.
F2 135-F adds five active TASK-214 files-lifecycle nodes/cases at
`tests/api/files/test_files_api.py` and `cases/files/TC-FILES-0001.md` through
`TC-FILES-0005.md`. Its future launcher must invoke the injected
`FilesHarnessFactory` once per test, expose only a fresh opaque `FilesRuntime`,
and perform exact delegate cleanup in the fixture `finally` path. The existing
`files-api` source profile targets exactly `tests/api/files`, with future case
root exactly `cases/files`. No concrete files delegate, A1 admission, or
profile-to-launcher binding exists.
`tests/integration/` and `tests/e2e/` remain absent and are created only by
their first accepted consumer. This documentation slice must not add empty
layer directories.

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
literal 47-profile registry, strict config/runtime/ownership/redaction and
traceability modules, the public-hook collection plugin, and
`TC-HEALTH-0001` with its fake-transport source tracer for `GET /api/live`.
Those F0 sources are independently source-reviewed. The health tracer is not
live Hub evidence. F1-01 HTTP policy and diagnostics source is independently
reviewed at `a5f3b728`; it remains injected-only, creates no network client or
request, and cannot yield runtime evidence. F1-02 identity/session source is
independently static-accepted at `2b037dd2`: four active cases/nodes require a
launcher-injected `IdentityHarnessFactory` that produces a fresh harness and
ownership ledger per test with exact cleanup. No concrete provisioner or
harness implementation exists. The `identity-api` target is exactly
`tests/api/identity` and its future launcher binding requires exactly
`--hub-case-root cases/identity`. A1 admission and accepted profile-to-launcher
binding remain absent. F1-03 authorization source is independently accepted at
`f184487a`: four active TASK-208 nodes/cases require an opaque fresh
`AuthzHarnessFactory` product per test, a nonce-bound ownership ledger, and
exact cleanup. F1-03 platform-abuse source is independently accepted at
`3932ac78`: two active TASK-209 nodes/cases require an opaque fresh
`PlatformAbuseHarnessFactory` product per test, exact cleanup, and only four
exact recorded Redis ownership entries. Their future profile roots are exactly
`tests/api/authz` and `tests/api/platform`, with case roots exactly
`cases/authz` and `cases/platform`; A1 admission, concrete provisioners, and
profile-to-launcher binding, `uv.lock`, and `images.lock.v3.json` remain absent.
`TC-AUTHZ-0005` remains
project-cross-scope blocked by the current backend middleware order: protected
handlers terminate before `rbac.RequirePermission` can run. F1-04 is
independently static-accepted across `bc5feef8`, `630bda9b`, `8ff13097`, and
`1868cf71`: the sealed source-bundle manifest is v3 with exactly 14 bounded
source blobs, and `f1-security-risk-ledger.toml` is closed to five ordered
critical/high edges. It binds three active OpenAPI nodes/cases, one active
privacy node/case, and one active auth-session-event node/case to the exact
future profile targets and case roots listed above. This is source-contract
coverage only; every global critical/high outcome remains `NOT_PROVEN`.
F2 135-F files-lifecycle source is independently static-accepted at
`af077c18`: five active TASK-214 nodes/cases require a launcher-injected
`FilesHarnessFactory`, one fresh opaque `FilesRuntime` per test, and exact
cleanup through that runtime. The existing `files-api` source profile targets
exactly `tests/api/files`, with future launcher case root exactly `cases/files`.
The concrete delegate, A1 admission, and profile-to-launcher binding remain
absent.

The current generated Swagger is known RED against those contracts: its
security-definition `in` values are malformed, generated paths retain a second
`/api` below `basePath`, magic verification omits the implemented 403 and 429
responses, TMS webhooks still advertise legacy `X-Hub-Token`/query-token auth
instead of the BearerAuth-or-APIKeyAuth alternatives, and public error responses
still reference internal `AppError` artifacts rather than the redacted wire
shape. Backend annotation/schema repair and Swagger regeneration remain
pending. No external OpenAPI-validator admission or execution exists. A1,
accepted profile-to-launcher bindings, concrete launcher execution, `uv.lock`,
and image evidence remain absent; the three source profile records do not grant
execution authority. Package and launcher stay `NOT_READY`, the concrete files
delegate and A1 admission remain absent, and runtime stays `NOT_RUN`. Remaining
F2 and later source is pending. F0/F1/F2 source closure cannot yield a runtime
result.

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
