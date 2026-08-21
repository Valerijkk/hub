# Hub Python QA project

`testframework/` is the separately extractable Python QA project boundary for
Hub. The tree contains a statically reviewed private runner source candidate,
but no accepted package, target suites, launcher runtime, or runtime authority.

Package/bootstrap boundaries for the next slice are defined in
[`PACKAGE-CONTRACT.md`](PACKAGE-CONTRACT.md). That document is architecture,
not evidence that `pyproject.toml`, `uv.lock`, `py.typed`, dependencies, or
executable suites exist, and does not promote the runner candidate to readiness.

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

Structure-only scaffold: **COMMITTED** as `22035c45`. Package bootstrap:
**NOT_READY**. Target suites: **NOT_READY**. Private runner source candidate:
**STATIC_SOURCE_REVIEWED**; launcher: **NOT_READY**; `PS5.1_PARSE_NOT_RUN`;
`PRIVATE_CONTRACT_NOT_RUN`. Runtime: **NOT_RUN**. `images.lock.v3.json` and
required local dependency bytes are absent. There is no supported direct runner or
fallback command. The next accepted consumer must create the first narrow
behavioral module test-first under its owning boundary and must not infer
runtime authority from this directory tree.

### Task 2A launcher receipt

Static launcher source was committed as
`e357e71ae445c797c686787e8cde2d643566bb68` (source
`cc0f3fcbf4e9a0adc70e946e739b81d21ac7da6c`, tree
`b16311222d3093ae4d2bb462a2e4d2835a3f38cf`, subject
`feat(qa): добавить закрытый public launcher`). The exact seven control blobs
are recorded in `runner/README.md`; Contract (Terra/high), Security
(Terra/xhigh), and Lifecycle (Terra/high) precommit and immutable reviews are
PASS C/I/M `0/0/0`, dispatcher-enforced, no fallback, telemetry `UNKNOWN`.

`runner/invoke-qa-gate.ps1` accepts only raw `-Target <target> -Task <task>
-Run <run>`. Invalid, extra, or unknown input exits `64` before effects. Every
known target remains literal `NOT_READY`, yielding one safe `hub.qa-gate/v1`
JSON line and exit `69`, with zero private runner, Docker, or report effect.
Portable roots, reparse rejection, linked-worktree support, and fixed-sibling
private handoff are static-only controls. `PS5.1_PARSE_NOT_RUN`,
`PRIVATE_CONTRACT_NOT_RUN`, `G0-B_PARITY_NOT_VERIFIED`, and `NOT_RUN` remain.
This receipt becomes current only after atomic Task 2B exact7 docs commit and
immutable review; it grants no image, dependency, Compose, Dockerfile, target,
runtime, or Task 3 authority.
