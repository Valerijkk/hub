# Hub Python QA project

`testframework/` is the separately extractable Python QA project boundary for
Hub. F0 package/config/runtime/traceability/public-plugin/profile and health
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
entries. No concrete provisioner, A1 admission, or profile-to-launcher binding
exists. A private runner source candidate also has static review only. None is
accepted package execution, launcher runtime, or runtime authority.

Package/bootstrap boundaries for the next slice are defined in
[`PACKAGE-CONTRACT.md`](PACKAGE-CONTRACT.md). That document is architecture,
not evidence that `pyproject.toml`, `uv.lock`, `py.typed`, or dependencies are
resolved, that a suite has executed, or that the runner candidate is ready.

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
  Planned `tests/{integration,contract,e2e}/` will own executable pytest suites
  grouped by layer and domain: API and E2E assert public observable behavior,
  contract covers API/schema/event/source contracts, and integration covers
  DB/broker/storage boundaries. Those layer directories remain absent.
- `selftests/` owns contracts of the framework itself.
- `load/` owns the separate measured Locust lifecycle.
- `cases/` owns repository-controlled `TC-*` intent, risk linkage, and
  observable outcomes in the `task -> requirement -> TC/case` chain.
- `schemas/` owns closed machine-readable case and run contracts.
- `policies/` owns marker, environment, redaction, and profile policies.
- `docker/` owns immutable pinned images, topologies, and resource/container
  contracts consumed by the sole runner boundary; it does not own launcher
  implementation.
- `runner/` owns the reviewed versioned launcher boundary. The private
  `runner/run-task.ps1` source candidate exists and has static C1/P2/P4 review
  only; it is not an accepted public command or executable runtime.
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

## Current status

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
and target-suite execution: **NOT_READY**. F1-04 and later source remain pending.
Private runner source candidate:
**STATIC_SOURCE_REVIEWED**; launcher: **NOT_READY**; `PS5.1_PARSE_NOT_RUN`;
`PRIVATE_CONTRACT_NOT_RUN`. Runtime: **NOT_RUN**. `uv.lock`, `images.lock.v3.json`,
and required local dependency bytes are absent. There is no supported direct runner
or fallback command. No file/source review promotes installation, execution, or
runtime to `PASS`.
