# ADR-TEST-ARCHITECTURE-REBASE-2026 — unit extraction and greenfield Python QA

- **Status:** Accepted
- **Date:** 2026-08-07
- **Amended:** 2026-08-08 — approved hybrid package/suite physical layout
- **Decision owner:** founder
- **Architecture owner:** `arch`
- **Verifier:** `arch-ver`
- **Scope:** TASK-204; rebased phases 128–141
- **Supersedes:** the proposed row-by-row translation, dual-run parity and Go
  coverage-preservation design from the earlier revision of this ADR

The founder has selected a clean cutover. Hub preserves only genuine isolated
native units. Legacy component, static-contract and Playwright suites are
deleted by an exact sealed allowlist; the non-unit QA system is then designed
and implemented from requirements and risk in one root Python project. Old
tests are evidence about history, not specifications to translate.

## Machine-readable decision markers

```text
ADR_STATUS=ACCEPTED
BASELINE_E2E=DOCKER:53_COLLECTED:48_PASS:5_EXPECTED_SKIP:0_FAIL
MIGRATION_MODE=GREENFIELD_REBUILD
LEDGER_HEAD=5708671f00411c7c9e9dbee88b725de2092cbacd
LEDGER_SHA256=72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511
CUTOVER_PARTITION=MOVE:507:KEEP_BLOCKED:245:DELETE:297:NONLEDGER_DELETE:28
PARTIAL_CUTOVER_ALLOWED=true
PHASE129_EXIT=HOOK_DEPENDENT_REFACTORED_AND_MOVED
ROOT_PYTHON_ONLY=testframework/**
PYTHON_PACKAGE_ROOT=testframework/src/hub_testkit/**
PYTHON_SUITE_ROOT=testframework/tests/**
PYTHON_COMPOSITION=testframework/tests/conftest.py:COMPOSITION_ONLY
PYTHON_SELFTEST_ROOT=testframework/selftests/**
GENERIC_CORE=FORBIDDEN
PROGRESSIVE_TREE=FIRST_CONSUMER:NO_EMPTY_DIRS
BACKEND_UNIT_ROOT=backend/tests/<domain>/**
FRONTEND_UNIT_ROOT=frontend/tests/<domain>/**
KEEP_ONLY=ISOLATED_NATIVE_UNITS
NO_ROW_PARITY=true
NO_DUAL_RUN=true
GO_COVERAGE_MIGRATION_BLOCKER=false
DELETE_ALLOWLIST=TASK-202:SEALED_SOURCE_PATH_SHA256_GROUPS
TRACEABILITY=TASK:REQUIREMENT:TEST_CASE:RISK
SCENARIO_DESIGN=HAPPY:NEGATIVE:EDGE:CORNER:SECURITY
DOCKER_ONLY=true
ROLLBACK=ATOMIC_COMMIT_REVERT
PROTECTED_ZONE_DEPENDENCY=FORBIDDEN
```

The executable companion is
`.planning/audits/test-rebase/test-architecture-contract.json`. Its stdlib-only
validator treats both the contract and TASK-202 ledger as untrusted input. It
uses constant rules, exact objects and SHA-256 digests of sorted source paths;
the JSON cannot replace its own validation rules.

## Acceptance evidence

The former TASK-203 blocker is closed. The full canonical Playwright baseline
ran in a task-owned, CPU/RAM-limited Docker environment on 2026-08-07:

| Collected | Passed | Expected skips | Failed | Environment skips |
|---:|---:|---:|---:|---:|
| 53 | 48 | 5 | 0 | 0 |

Four expected skips are provider-dependent AI scenarios and one is the explicit
health negative sentinel. Missing infrastructure, readiness, seed or mandatory
credentials did not become a skip. This baseline proves the application and
auth seed were runnable before deletion; it is not a parity target for the new
Python suite.

## Final ownership roots

After cutover, executable suites and reusable Python modules have separate
ownership surfaces:

```text
backend/tests/<domain>/**       Go; genuine service-free units only
frontend/tests/<domain>/**      TypeScript; genuine Vitest/RTL units only
frontend/tests/support/**       support reachable only from retained units
testframework/tests/**          executable pytest suites by layer/domain
testframework/load/**           separate Locust lifecycle and measured clients
testframework/selftests/**      framework contracts only, never SUT behaviour
testframework/src/hub_testkit/** reusable installable modules by responsibility/domain
```

The Python project lives directly in `testframework/`, while its sole
installable import namespace is `hub_testkit.*` under
`testframework/src/hub_testkit/`. A nested `testframework/python/` project,
generic `core/` and direct root suite packages are forbidden:

```text
testframework/
├── pyproject.toml              single dependency/tool and pytest authority
├── uv.lock                     exact dependency resolution authority
├── src/hub_testkit/            reusable modules by narrow responsibility/domain
├── selftests/                  framework contracts only
├── tests/
│   ├── conftest.py             explicit plugin/fixture composition only
│   └── {api,integration,contract,e2e}/<domain-or-boundary>/
├── load/                       separate Locust lifecycle and measured client
├── cases/<domain>/TC-*.md      repository-owned business intent
├── schemas/                    test-case and run-contract schemas
├── policies/                   environment/marker/redaction/profile policies
├── docker/                     pinned images/topologies and sole outer launcher
├── runner/                     validated in-container entrypoint/profile loading
└── reports/<task-run>/         ignored owner-marked per-run evidence only
```

The tree is progressive: no empty directory or speculative module/layer is a
required artifact. `testframework/tests/conftest.py` is composition-only.
Reusable UI code lives only below
`testframework/src/hub_testkit/domains/<domain>/ui/{pages,components}/`; root or
suite-local `support/`, `pages/` and `steps/` warehouses are forbidden.

Executable helpers under `testframework/**` are Python. Dockerfiles,
YAML/TOML/JSON, Markdown and generated report data are declarative artifacts,
not alternative test stacks. Go, Vitest and TypeScript Playwright executables
are forbidden in the root after cutover.

## Exact move, keep and delete boundary

TASK-202 is a deletion inventory, not a translation map. The validator seals
each group by both count and SHA-256 over sorted canonical source paths.

| Disposition | Classification | Files / cases | Meaning |
|---|---|---:|---|
| move now | hook-free `go-unit` | 350 / 1,264 | move to `backend/tests/<domain>/**` |
| keep/block | hook-dependent `go-unit` | 245 / 1,428 | retain in place until public-seam refactor, then move |
| move now | `frontend-unit` | 156 / 409 | move to `frontend/tests/<domain>/**` |
| move now | unit-reachable `render.tsx` | 1 / 0 | move to `frontend/tests/support/render.tsx` |
| delete/no-port | Go API/integration component | 237 / 982 | rebuild coverage from requirements |
| delete/no-port | frontend static contract | 24 / 110 | redesign Python contracts |
| delete/no-port | compile-only frontend fixtures | 2 / 0 | delete with their task-specific tsconfig guards |
| delete/no-port | TypeScript Playwright spec | 20 / 53 | redesign Python E2E |
| delete/no-port | TypeScript Playwright support | 14 / 0 | redesign Python support |

The immediate move set is exactly **507 files / 1,673 cases**. The blocked set
is **245 / 1,428** and the ledger deletion set is **297 / 1,145**. These nine
groups partition all **1,049 rows / 4,246 cases** at ledger revision
`5708671f00411c7c9e9dbee88b725de2092cbacd`. A separate exact non-ledger
closure contains 28 obsolete runners, configs, helpers and stage documents.

A count-only match is insufficient: every group is sealed by SHA-256 over its
sorted canonical source paths. Any added, removed, renamed, duplicated,
traversing or reclassified source makes validation red. Of the three legacy
frontend-support rows, only `render.tsx` remains unit-reachable; the two compile
fixtures belong to the delete group.

No legacy non-unit assertion, target path, case count or skip count is a
required mapping for new Python tests. Copying a legacy test is allowed only as
an explicitly reviewed implementation aid; it never satisfies traceability or
design acceptance by itself.

## Native unit eligibility

### Go

A movable Go unit is deterministic and in-process. It may use public
constructors, pure fakes, fake clocks and `httptest.ResponseRecorder`, but it
must not start an application/listener or require PostgreSQL, MongoDB, Redis,
NATS, Docker/testcontainers, filesystem/provider boundaries,
`testframework/env`, `mongotest`, shared global state or private export hooks.
The destination uses external-package semantics where practical. The 245 rows
that still consume 434 private symbols through 653 links across 42 hook files
are not admitted to the movable allowlist. They stay in the legacy location,
are refactored to public/injected seams, receive a regenerated zero-hook ledger
classification and then move as native units. They are neither deleted nor
silently treated as component replacements.

### Frontend

A retained frontend unit is Vitest/RTL behavior for a component, hook, store,
schema or deterministic client with local fakes/MSW. It cannot require a real
browser/application/service. Source scanning, docs drift, filename/token
assertions and other static guards are non-units and are deleted. Support is
retained only when the import graph proves that every consumer is a retained
frontend unit.

## Greenfield Python design contract

The source of truth is current product behavior and documentation, not legacy
test shape. Before implementing a Python slice, the owner builds a risk-based
test design from `docs/`, API/schema contracts, linked work items, incidents and
security constraints.

Every executable case must expose stable metadata for:

- one unique test-case ID (`TC-*`);
- at least one task or backlog link;
- at least one requirement, story, NFR or accepted contract link;
- risk ID and priority/severity;
- scenario kind and owning domain;
- deterministic preconditions, action and observable expected result.

Every feature matrix covers applicable happy path, negative/error handling,
boundary/edge, corner/state-transition, authorization/security and
recovery/idempotency/concurrency risks. “Applicable” exclusions require an
owner reason in the matrix; an empty dimension cannot silently pass. Generated
or parameterized cases retain individual IDs in JUnit/Allure. Tests assert
public behavior and externally observable state, never private implementation
details merely copied from the old suite.

Definition of done for a Python slice:

1. requirement/risk matrix reviewed and all planned IDs unique;
2. Docker-only collection is non-zero and matches the matrix;
3. expected pass/fail/skip totals and runtime are recorded;
4. at least one controlled negative or mutation sentinel proves detection;
5. JUnit and Allure artifacts are non-empty, sanitized and link back to IDs;
6. no live external egress or production/shared state is used;
7. independent verifier accepts the slice.

There is deliberately no row-by-row parity state machine and no same-SHA
legacy/candidate dual run. Progress is measured by accepted requirement and
risk coverage, not by recreating 3,000+ historical assertions.

## Coverage decision

The historical Docker baseline (82.0373% overall Go statement coverage) remains
useful diagnostic evidence. It is not a migration acceptance threshold and may
drop when component tests leave Go.

- Go unit jobs gate on compilation, race-safe execution, zero unexpected
  skips/retries and service-free isolation.
- Python black-box tests never claim Go line coverage.
- No overall 82% or per-package 80% threshold blocks unit extraction, legacy
  deletion or Python scaffolding.
- Future code-coverage policy may be introduced only by a separate
  founder-approved ADR; it cannot be smuggled back into this migration.

Quality is protected by requirement/risk traceability, critical-path gates,
security scenarios, mutation/negative sentinels and observed production risks.

## Canonical jobs

All execution is Docker-only in uniquely named task-owned containers with
explicit CPU/RAM limits. Host execution is noncanonical. Missing mandatory
services fail; they never skip.

| Job | Ownership | Services | Required evidence |
|---|---|---|---|
| `backend-unit` | `backend/tests/**` | none | result log/JSON, race result |
| `frontend-unit` | `frontend/tests/**` | none | Vitest JSON, typecheck/lint log |
| `python-contract` | `testframework/tests/contract/**` | none | JUnit, Allure |
| `python-api` | `testframework/tests/api/**` | isolated application stack | JUnit, Allure, app log |
| `python-integration` | `testframework/tests/integration/**` | declared isolated dependencies | JUnit, Allure, service logs |
| `python-e2e-smoke` | `testframework/tests/e2e/**` | isolated full app + browser | JUnit, Allure, traces/screenshots |
| `python-e2e-full` | `testframework/tests/e2e/**` | isolated full app + browser | JUnit, Allure, traces/screenshots/app log |
| `python-load` | `testframework/load/**` | isolated full stack | Locust CSV/HTML, service logs |
| `test-architecture-contract` | ADR/contract/ledger | none | validator and negative-self-test logs |

Unit retries are zero. E2E retries are limited to two and only for classified
infrastructure instability; every attempt remains visible. Quarantine requires
an issue, owner, reason and expiry wave. Zero collection, a missing dependency,
an unexpected skip or an empty artifact fails closed.

## Cutover, deletion and rollback

The cutover is an exact, reviewable filesystem operation and may proceed in
safe partial batches:

1. regenerate TASK-202 inventory on the cutover revision;
2. validate all nine sealed source-path groups, the 1,049-row partition and
   the exact 28-path non-ledger closure against the founder manifest;
3. move the 507 immediately eligible native-unit/support rows and prove their
   isolation/import rules;
4. keep the 245 hook-dependent Go rows and all still-consumed hooks in place;
5. produce `cutover-manifest.json` containing every move, deletion, CI
   entrypoint change and zero-consumer hook removal;
6. run moved unit suites and architecture validators in limited Docker;
7. delete only the sealed 297-row non-unit set plus exact 28-path non-ledger
   closure; no glob-expanded or ad-hoc deletion is permitted;
8. switch affected docs/manifests/required jobs in the same atomic batch;
9. refactor and move all 245 blocked units before Phase 129 is complete.

The 507-row move and 297+28 deletion do not wait for the hook refactor. This is
the authorized partial cutover. Phase 129, however, cannot be marked complete
while any of the 245 rows remains hook-dependent or outside `backend/tests/**`.
Each batch is an allow-listed atomic commit; partial progress never broadens the
next batch's authority.

The new Python suites do not have to reproduce legacy rows before deletion.
Their acceptance proceeds domain-by-domain using the greenfield definition of
done. Production `export_hooks*.go` files are removed only when an exact symbol
consumer scan proves zero remaining product/unit consumers; test deletion alone
does not authorize production-code deletion.

Rollback is reverse-order `git revert` of the affected atomic batch commits.
Each batch manifest restores its moves, deletions and entrypoint changes
together. Cleanup may remove
only resources with the current task’s labels/names. Shared or production data,
networks, volumes and containers are never reset. Finished task-owned
containers are removed immediately unless a named downstream task has recorded
their continued ownership.

## Consequences

- The destination is simpler and avoids spending time porting obsolete test
  mechanics.
- Historical assertion counts and Go statement coverage are intentionally not
  preserved.
- Requirements and risks must be made explicit before Python implementation;
  missing product documentation becomes visible work instead of being hidden by
  copied tests.
- Exact sealed deletion and an atomic revert keep the aggressive cutover
  recoverable.
- The protected external zone is neither read nor used as an implicit source;
  any founder handoff must arrive as an explicit reviewed artifact outside it.

## Validation

From the repository root, the wrapper itself launches a limited read-only
Docker validator; it never executes Python on the Windows host:

```powershell
C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File .planning/audits/test-rebase/validate-contract.ps1
```

The wrapper runs four fail-closed checks in the same container: TASK-202
inventory provenance/check and negative self-test, then the contract/founder-
manifest validator and its negative self-test. The latter proves, among other
sentinels, that forcing a hook-dependent Go row into the movable set is RED.
Normal product/test changes may legitimately make the earlier ledger stale;
that is an expected RED until the ledger, founder manifest and sealed hashes
are deliberately regenerated on the stable cutover revision.
