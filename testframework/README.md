# Hub Python QA project

`testframework/` is the separately extractable Python QA project boundary for
Hub.

## Current framework slice (2026-08-23)

`pyproject.toml`, `uv.lock`, `src/hub_testkit/py.typed`,
`docker/Dockerfile.qa`, and `docker/entrypoints/python-qa.sh` are committed.
The only QA entrypoint is root Compose:
`docker compose -f docker-compose.local.yml run --rm <exact-service>`, with
exact services `backend-unit`, `frontend-unit`, `python-contract`,
`python-api`, `python-integration`, `python-e2e-smoke`, `python-e2e-full`, and
`python-load`. `python-contract` ran 181/181 framework selftests in Docker on
2026-08-23. That result does not cover SUT behavior, Mail/Calendar, API,
integration, e2e, load, hosted CI, or production; all of those remain `NOT_RUN`.

`testframework/runner/**` and `qa-gate` are retired/absent and are never a
fallback. The five SUT contract tests/cases were removed by founder direction.

## Historical source-authoring record

The source-review receipts below are retained for provenance. Any statement in
this historical record that package bootstrap, `uv.lock`, Docker assets, or a
private runner is absent is superseded by the current framework slice above.
It does not supersede the `NOT_RUN` SUT/runtime boundaries.

F0 package/config/runtime/traceability/public-plugin/profile and health
sources are authored and independently source-reviewed. F1-01 injected-only
HTTP policy, transport, and redacted diagnostics source is independently
source-reviewed at `a5f3b728`; it creates no network client or runtime
evidence. F1-02 identity/session source is independently static-accepted at
`2b037dd2`: four active identity cases/nodes require a launcher-injected
`IdentityHarnessFactory` that produces a fresh harness and ownership ledger per
test with exact cleanup. No concrete provisioner or harness implementation
exists. F1-03 authorization and platform-abuse source is independently accepted
at `f184487a` and `3932ac78`: four TASK-208 and two TASK-209 active nodes/cases
require launcher-injected opaque fresh factories, per-test nonce-bound ledgers,
and exact cleanup; the platform contract records only four exact Redis ownership
entries. F1-04 source is independently static-accepted across `bc5feef8`,
`630bda9b`, `8ff13097`, and `1868cf71`: its sealed v3 manifest contains 14
bounded source blobs and its closed ledger contains five ordered critical/high
risk edges for three OpenAPI, one privacy, and one auth-session-event active
source contract. F2 135-F files-lifecycle source is independently static-
accepted at `af077c18`: five active TASK-214 cases/nodes require a
launcher-injected `FilesHarnessFactory`, a fresh opaque `FilesRuntime` per test,
and exact cleanup. The existing `files-api` source profile targets exactly
`tests/api/files`, with future launcher case root exactly `cases/files`. The
references in this historical receipt to a missing delegate, A1 admission,
profile binding, executable launcher, or private runner candidate describe the
pre-merge state only. They do not alter the committed package/root-Compose
boundary or promote any SUT runtime beyond `NOT_RUN`; all global critical/high
outcomes remain `NOT_PROVEN`.

Package/bootstrap boundaries are defined in
[`PACKAGE-CONTRACT.md`](PACKAGE-CONTRACT.md). The current 181/181 receipt is
framework-only Docker evidence, not a SUT or product-runtime assertion.

## Legacy quarantine

The existing `stages/**`, `pipeline/**`, `guides/**`, `TESTING_GUIDE.md`, and
related pipeline and guide content are quarantined legacy. They remain intact
until exact Phase 129 manifest handling and must not be imported, copied, or
used as authority by new work.

The current direct `testframework/e2e/` directory has no tracked suite content
and is also quarantined legacy. It is not a target suite root, must not be
populated, renamed, reused, or copied, and requires an exact closure/removal
decision before Phase 129 may treat the target tree as ready. The planned
Python E2E root is the currently absent `testframework/tests/e2e/`, created
progressively only by its first accepted consumer.

## Target ownership tree

- `src/hub_testkit/` owns reusable framework code.
  - `config/` owns immutable validated environment and configuration
    composition.
  - `runtime/` owns run identity, ownership, bounded timing, polling, and
    redaction primitives.
  - `traceability/` owns `TC-*` metadata loading and the risk-linked
    `task -> requirement -> TC/case` graph.
  - `http/` owns injected-only HTTP policy, bounded retry, request identity,
    and redacted diagnostics; it never creates a default client or network
    request.
  - `domains/<domain>/` owns that domain's models, facades, builders, and
    oracles. Optional UI abstractions live only under the owning domain in
    `ui/pages/` and `ui/components/`.
- `tests/api/health/` contains the first source-authored tracer, linked to
  `TC-HEALTH-0001`: a fake-transport liveness contract, not live Hub evidence.
  `tests/api/identity/` contains four active identity/session source nodes,
  linked to `TC-IDENTITY-0001`--`TC-IDENTITY-0004`; its `identity-api` profile
  target is exactly `tests/api/identity` and requires exactly
  `--hub-case-root cases/identity` when a launcher binding exists. It has no
  concrete provisioner or harness implementation.
  `tests/api/authz/` contains four active authorization source nodes, linked to
  `TC-AUTHZ-0001`--`TC-AUTHZ-0004` for TASK-208. Its future profile root is
  exactly `tests/api/authz` and requires exactly `--hub-case-root cases/authz`;
  every test requires an opaque fresh launcher-injected `AuthzHarnessFactory`, a
  nonce-bound ledger, and exact cleanup. `tests/api/platform/` contains two
  active platform-abuse source nodes, linked to `TC-PLATFORM-0001` and
  `TC-PLATFORM-0002` for TASK-209. Its future profile root is exactly
  `tests/api/platform` and requires exactly `--hub-case-root cases/platform`;
  every test requires an opaque fresh launcher-injected
  `PlatformAbuseHarnessFactory` and exact cleanup of only four recorded Redis
  ownership entries. Neither has a concrete provisioner or launcher binding.
- `tests/api/files/` contains five active TASK-214 files-lifecycle source nodes,
  linked to `TC-FILES-0001`--`TC-FILES-0005`. Each future execution requires a
  launcher-injected `FilesHarnessFactory`, one fresh opaque `FilesRuntime`, and
  exact cleanup in the fixture `finally` path. The existing `files-api` source
  profile targets exactly `tests/api/files`, with future launcher case root
  exactly `cases/files`; no concrete delegate or launcher binding exists.
- `tests/contract/openapi/` contains three active TASK-227 source contracts,
  linked to `TC-OPENAPI-0001`--`TC-OPENAPI-0003`; source profile
  `openapi-contract` targets exactly `tests/contract/openapi`, with future
  launcher case root exactly `cases/openapi`.
  `tests/contract/privacy/` contains the active `TC-PRIVACY-0001` source
  contract; source profile `privacy-security` targets exactly
  `tests/contract/privacy`, with future launcher case root exactly
  `cases/privacy`. `tests/contract/security/` contains the active
  `TC-SECURITY-0001` auth-session-event source contract; source profile
  `security-risk` targets exactly `tests/contract/security`, with future
  launcher case root exactly `cases/security`. The closed ledger stores the
  corresponding repo-relative case roots below
  `testframework/cases/{openapi,privacy,security}`. These are source contracts,
  not executed evidence. Planned `tests/{integration,e2e}/` remain absent.
- `selftests/` owns contracts of the framework itself.
- `load/` owns the separate measured Locust lifecycle.
- `cases/` owns repository-controlled `TC-*` intent, risk linkage, and
  observable outcomes in the `task -> requirement -> TC/case` chain.
- `schemas/` owns closed machine-readable case and run contracts.
- `policies/` owns marker, environment, redaction, and profile policies.
- `docker/` owns immutable pinned images, entrypoints, topologies, and
  resource/container contracts consumed by root Compose.
- `runner/` is retired and absent. Root Compose exact services are the sole QA
  entrypoint; no private runner or `qa-gate` fallback exists.
- `reports/<task-run>/` is the exact generated-artifact boundary. Every run
  directory has a task-run owner marker, and nothing may write outside its
  owned run directory.

Native units remain in `backend/tests/**` and `frontend/tests/**`; they never
belong in this Python project. Reusable design is domain-first: facades,
builders, oracles, pages, and components stay below their owning domain. A
generic helper warehouse, cross-domain dumping ground, or suite-local reusable
framework is forbidden.

Traceability is always explicit and risk-linked:
`task -> requirement -> TC/case`. No case or executable suite may bypass that
chain.

## Historical detailed source status

The following pre-merge source-status detail is retained as provenance. Its
former package/lock/image/runner `NOT_READY` wording is historical; the current
framework status is the 181/181 Docker receipt stated above. SUT execution and
Mail/Calendar runtime remain `NOT_RUN`.

Structure-only scaffold: **COMMITTED** as `22035c45`. F0 package/config/runtime/
traceability/public-plugin/profile and health source: **SOURCE_AUTHORED** and
**SOURCE_REVIEWED** through `92e9fec7`; the health test is fake-transport source
contract only, not live evidence. F1-01 HTTP policy and diagnostics source:
**SOURCE_AUTHORED** and **SOURCE_REVIEWED** at `a5f3b728`; it is injected-only,
creates no network client or runtime evidence. F1-02 identity/session source:
**SOURCE_AUTHORED** and independently **STATIC_ACCEPT** at `2b037dd2`, with four
active cases/nodes and a future launcher-injected `IdentityHarnessFactory` that
creates a fresh harness and ownership ledger per test with exact cleanup. There
is no concrete provisioner or harness implementation. The `identity-api` target
is exactly `tests/api/identity` and requires exactly `--hub-case-root
cases/identity`; accepted A1 admission and profile-to-launcher binding remain
absent. F1-03 authorization source is independently **ACCEPTED** at `f184487a`:
four active TASK-208 nodes/cases require an opaque fresh launcher-injected
`AuthzHarnessFactory`, a nonce-bound ownership ledger, and exact cleanup per
test. F1-03 platform-abuse source is independently **ACCEPTED** at `3932ac78`:
two active TASK-209 nodes/cases require an opaque fresh launcher-injected
`PlatformAbuseHarnessFactory`, exact cleanup, and only four exact recorded Redis
ownership entries per test. Their future profile roots are exactly
`tests/api/authz` and `tests/api/platform`, with case roots exactly
`cases/authz` and `cases/platform`; concrete provisioners, A1 admission, and
profile-to-launcher bindings remain absent. `TC-AUTHZ-0005` is
project-cross-scope blocked by the current backend middleware order: protected
handlers terminate before `rbac.RequirePermission` can run. Package bootstrap
and target-suite execution: **NOT_READY**. F1-04 is independently **STATIC
ACCEPT** across `bc5feef8`, `630bda9b`, `8ff13097`, and `1868cf71`: sealed
source-bundle manifest v3 has exactly 14 bounded blobs, and the closed risk
ledger has five ordered edges for three active OpenAPI, one active privacy, and
one active auth-session-event source contract. Every global critical/high
outcome remains **NOT_PROVEN**.

F2 135-F files-lifecycle source is independently **STATIC ACCEPT** at
`af077c18`: five active TASK-214 cases/nodes require a launcher-injected
`FilesHarnessFactory`, one fresh opaque `FilesRuntime` per test, and exact
cleanup through the fixture `finally` path. The existing `files-api` source
profile targets exactly `tests/api/files`, with future launcher case root
exactly `cases/files`. The concrete delegate, A1 admission, and accepted
profile-to-launcher binding remain absent.

The current generated Swagger is known RED: security-definition `in` values
are malformed, paths retain a second `/api` below `basePath`, magic verification
omits implemented 403/429 responses, TMS webhook docs still expose legacy
`X-Hub-Token`/query-token auth instead of BearerAuth-or-APIKeyAuth alternatives,
and public error responses reference internal `AppError` artifacts. Backend
repair and Swagger regeneration are pending. External OpenAPI-validator
admission/execution and accepted SUT profile bindings remain pending. The
references in this historical record to absent `uv.lock`, image evidence,
dependency bytes, or an executable/private runner describe the pre-merge state
only: `uv.lock` and the framework Docker assets are now committed, and root
Compose is the sole QA API. Source profile records do not grant SUT execution
authority. Remaining F2 and later source is pending; SUT runtime remains
`NOT_RUN`. No file/source review promotes SUT execution or product runtime to
`PASS`.
