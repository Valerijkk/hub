# Hub Python QA project

`testframework/` is the future separately extractable Python QA project for
Hub. This slice establishes ownership boundaries only; it does not introduce
behavior, dependencies, a launcher, or runtime authority.

## Legacy quarantine

The existing `stages/**`, `pipeline/**`, `guides/**`, `TESTING_GUIDE.md`, and
related pipeline and guide content are quarantined legacy. They remain intact
until exact Phase 129 manifest handling and must not be imported, copied, or
used as authority by new work.

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
- `tests/{api,integration,contract,e2e}/` owns executable pytest suites grouped
  by layer and domain: API and E2E assert public observable behavior, contract
  covers API/schema/event/source contracts, and integration covers
  DB/broker/storage boundaries.
- `selftests/` owns contracts of the framework itself.
- `load/` owns the separate measured Locust lifecycle.
- `cases/` owns repository-controlled `TC-*` intent, risk linkage, and
  observable outcomes in the `task -> requirement -> TC/case` chain.
- `schemas/` owns closed machine-readable case and run contracts.
- `policies/` owns marker, environment, redaction, and profile policies.
- `docker/` owns pinned images, topologies, and resource/container contracts
  consumed by the future launcher; it does not own launcher implementation.
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

Structure only: **NOT_READY**. Runtime: **NOT_RUN**. There is no supported
direct runner or fallback command. The next accepted consumer must create the
first narrow behavioral module test-first under its owning boundary.
