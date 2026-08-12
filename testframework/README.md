# Hub Python QA project

`testframework/` is the future separately extractable Python QA project for
Hub. This slice establishes ownership boundaries only; it does not introduce
behavior, dependencies, a launcher, or runtime authority.

Package/bootstrap boundaries for the next slice are defined in
[`PACKAGE-CONTRACT.md`](PACKAGE-CONTRACT.md). That document is architecture,
not evidence that `pyproject.toml`, `uv.lock`, `py.typed`, dependencies,
executable suites or a launcher exist.

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
  - `domains/<domain>/` owns that domain's models, facades, builders, and
    oracles. Optional UI abstractions live only under the owning domain in
    `ui/pages/` and `ui/components/`.
- Planned `tests/{api,integration,contract,e2e}/` will own executable pytest
  suites grouped by layer and domain: API and E2E assert public observable
  behavior, contract covers API/schema/event/source contracts, and integration
  covers DB/broker/storage boundaries. These layer directories are currently
  absent; `tests/README.md` is the only suite-boundary scaffold artifact.
- `selftests/` owns contracts of the framework itself.
- `load/` owns the separate measured Locust lifecycle.
- `cases/` owns repository-controlled `TC-*` intent, risk linkage, and
  observable outcomes in the `task -> requirement -> TC/case` chain.
- `schemas/` owns closed machine-readable case and run contracts.
- `policies/` owns marker, environment, redaction, and profile policies.
- `docker/` owns immutable pinned images, topologies, and resource/container
  contracts consumed by the future launcher; it does not own launcher
  implementation.
- `runner/` owns the future reviewed versioned launcher code and entrypoint and
  is not created by this structure-only slice.
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

Structure-only scaffold: **COMMITTED** as `22035c45`. Package bootstrap:
**NOT_READY**. Target suites: **NOT_READY**. Future `runner/**` launcher:
**NOT_READY**. Runtime: **NOT_RUN**. There is no supported direct runner or
fallback command. The next accepted consumer must create the first narrow
behavioral module test-first under its owning boundary and must not infer
runtime authority from this directory tree.
