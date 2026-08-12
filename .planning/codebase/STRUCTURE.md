# Codebase Structure

**Analysis Date:** 2026-08-07

## Directory Layout

```text
hub/
├── backend/
│   ├── go.mod                         # Go module boundary; permits imports of backend/internal/**
│   ├── internal/                      # Production domains; 0 *_test.go
│   │   └── **/export_hooks*.go        # Test-only symbols under //go:build hubtest
│   ├── cmd/server/
│   │   └── wiring_fiber_recover_test.go # Only Go test outside backend/testframework
│   ├── ent/                           # Generated Ent client + committed schema sources
│   └── testframework/                 # 829 Go tests + 4 support .go files
│       ├── {auth,tracker,tms,...}/     # External packages <domain>_test
│       ├── platform/<subsystem>/       # Nested platform tests
│       ├── ent/                        # Ent/schema tests
│       └── {env,factory,mongotest,steps}/ # Shared Go test support
├── frontend/
│   ├── src/                           # Production TS/React; src/test/setup.ts only, 0 test files
│   ├── e2e/README.md                  # Stale compatibility pointer; no spec files
│   ├── vite.config.ts                 # Vitest external include/aliases
│   ├── playwright.config.ts           # Active Playwright config; external testDir
│   ├── package.json                   # Vitest/Playwright entry scripts
│   └── scripts/
│       └── link-testframework-node-modules.mjs # Generates dependency junctions
├── testframework/                     # 241 tracked legacy framework files; 0 Python files
│   ├── {README,TESTING_GUIDE,PIPELINE,MIGRATION,AUDIT}.md
│   ├── guides/                        # 6 legacy guides
│   ├── pipeline/                      # 6 YAML stage/job manifests
│   └── stages/                        # 224 tracked files
│       ├── unit/
│       │   ├── backend-test/README.md # Pointer to backend/testframework
│       │   └── frontend-test/
│       │       ├── render.tsx         # Shared React Query renderer
│       │       └── domains/           # 180 Vitest files in 13 domain dirs
│       ├── e2e/
│       │   ├── domains/               # 20 Playwright TS spec in 9 domain dirs
│       │   └── shared/                # 13 fixtures/helpers/POM/steps files
│       ├── build/                     # Stage README
│       └── quality/                   # Coverage gate pointer
├── .github/workflows/ci.yml           # Executable canonical test pipeline
├── .gitlab-ci.yml                     # Secondary pipeline with path/tag drift
└── Makefile                           # Secondary local entry point with path/tag drift
```

The protected `pro-qa-ai/**` tree is intentionally absent from this map and from all counts.

## Approved Target Layout

Founder target for this mapping overrides the older `internal/<domain>/unit` and `src/<module>/unit` paths still present in `.planning/active/august-2026/05-AUGUST-DETAILED.md` and TASK-197/198:

```text
hub/
├── backend/
│   ├── internal/                      # Production Go remains unchanged
│   └── tests/                         # Developer-owned Go unit tree
│       └── <domain>/                  # package <domain>_test; domain-local helpers only when explicitly approved
├── frontend/
│   ├── src/                           # Production TS/React remains unchanged
│   └── tests/                         # Developer-owned Vitest tree
│       ├── <domain>/                  # *.test.ts / *.test.tsx by system under test
│       └── support/                   # setup/render/factories
└── testframework/                     # QA-owned standalone Python project
    ├── pyproject.toml                 # pytest, Playwright Python, Locust
    ├── conftest.py                    # Root fixtures
    ├── core/                          # Settings/clients/factories/shared fixtures
    ├── integration/                   # DB, Mongo, NATS, external service flows
    ├── api/                           # Whole HTTP API scenarios/contracts
    ├── e2e/                           # Playwright Python + POM/steps
    ├── contract/                      # Schemas/events/static source contracts
    ├── load/                          # Locust profiles; local/scheduled only
    ├── docker/                        # Test-service profiles/orchestration
    ├── reports/                       # Sanitized artifact metadata
    └── README.md                      # One-command entry points
```

`backend/tests/**` remains a separate Go package tree: it can import `backend/internal/**` because it is inside the module rooted at `backend/go.mod`, but it cannot access private parent-package symbols directly. Keep required `backend/internal/**/export_hooks*.go` under `hubtest` until their consumers are removed.

## Directory Purposes

**`backend/testframework/`:**
- Purpose: Current executable backend test root.
- Contains: 833 tracked Go files: 829 `*_test.go` and four reusable support files.
- Key files: `backend/testframework/env/builder.go`, `backend/testframework/factory/user.go`, `backend/testframework/mongotest/setup.go`, `backend/testframework/steps/http.go`.
- Placement rule: Phase 128 classification sends only proven genuine units to
  `backend/tests/**`. DB/broker/external/whole-API rows are deleted by the exact
  Phase 129 manifest without Python parity; new Python coverage is designed
  independently from current requirements/risks/TC.

**`backend/internal/`:**
- Purpose: Production Go domains and platform packages.
- Contains: No `*_test.go`; 39 `export_hooks.go` and six additional `hubtest`-tagged support files bridge external tests to private code.
- Key files: `backend/internal/auth/export_hooks.go`, `backend/internal/tracker/export_hooks.go`, `backend/internal/platform/middleware/export_requestid_hooks.go`, `backend/internal/tms/export_hooks_repo.go`.
- Placement rule: Do not add founder-target tests here. Do not delete hooks merely because tests moved from `backend/testframework/` to `backend/tests/`; both are external package trees.

**`backend/cmd/server/`:**
- Purpose: Backend composition root.
- Contains: `backend/cmd/server/wiring_fiber_recover_test.go`, the only Go test outside `backend/testframework/`.
- Key files: `backend/cmd/server/wiring_fiber_recover_test.go`.
- Placement rule: Decide explicitly whether composition-root tests remain co-located or move to `backend/tests/server/`; current GitHub `backend-test` does not execute this file because it targets only `./testframework/...` in `.github/workflows/ci.yml`.

**`testframework/stages/unit/frontend-test/`:**
- Purpose: Current executable Vitest root outside `frontend/`.
- Contains: 180 tests / 519 cases (156 unit + 24 static contract), plus 3 ledger support rows (renderer + two fixture files).
- Key files: `testframework/stages/unit/frontend-test/render.tsx`, `testframework/stages/unit/frontend-test/domains/auth/client-api.test.ts`, `testframework/stages/unit/frontend-test/domains/v4/`.
- Placement rule: Move only the sealed 156 genuine units by imported SUT and move `render.tsx` to `frontend/tests/support/`. Delete 24 static contracts and the other two support rows; do not split/port them into new destinations.

**`frontend/src/test/`:**
- Purpose: Current frontend test runtime setup.
- Contains: `frontend/src/test/setup.ts` importing jest-dom matchers; no test cases.
- Key files: `frontend/src/test/setup.ts`.
- Placement rule: Do not move or re-export this legacy setup implicitly. Its keep/edit/delete action follows the exact Phase 129 consumer allowlist; any target setup beyond the sealed `render.tsx` move is created greenfield.

**`testframework/stages/e2e/`:**
- Purpose: Current TypeScript Playwright suite.
- Contains: 20 specs / 53 cases, plus 14 ledger support rows (13 shared files + `frontend/playwright.config.ts`).
- Key files: `testframework/stages/e2e/shared/fixtures.ts`, `testframework/stages/e2e/shared/helpers/auth.ts`, `testframework/stages/e2e/shared/pages/base.page.ts`, `testframework/stages/e2e/shared/steps/auth.steps.ts`.
- Placement rule: Delete this TypeScript suite and its support/config closure by
  exact Phase 129 manifest. Do not port files or preserve 20/53 counts; create
  Python E2E greenfield.

**`testframework/pipeline/`:**
- Purpose: Human-readable mirror of pipeline topology.
- Contains: `testframework/pipeline/stages.yaml` and five manifests in `testframework/pipeline/jobs/`.
- Key files: `testframework/pipeline/jobs/backend-test.yaml`, `testframework/pipeline/jobs/frontend-test.yaml`, `testframework/pipeline/jobs/e2e-full.yaml`.
- Placement rule: Treat these manifests as legacy inventory. Phase 129 switches
  exact consumers and deletes the closure without CI parity; target execution
  becomes available only through the independently accepted outer launcher.

**`frontend/e2e/`:**
- Purpose: Compatibility documentation only.
- Contains: tracked `frontend/e2e/README.md`; no `*.spec.ts`.
- Key files: `frontend/e2e/README.md`.
- **Historical/forbidden pointer inventory:** do not use this README as migration source or execution authority; its direct-layer pointer is retired relative to the legacy `frontend/playwright.config.ts` inventory.

## Key File Locations

**Entry Points:**
- `.github/workflows/ci.yml`: Canonical GitHub jobs for backend, coverage, frontend, smoke E2E and full E2E.
- `frontend/package.json`: Local Vitest and Playwright scripts.
- `frontend/vite.config.ts`: Current Vitest include/exclude, aliases and environment.
- `frontend/playwright.config.ts`: Current E2E testDir, projects, reporter and web server.
- `backend/go.mod`: Go module/internal import boundary.

**Configuration:**
- `frontend/tsconfig.json`: Type-checks only `frontend/src`; root Vitest files are outside its include.
- `frontend/scripts/link-testframework-node-modules.mjs`: Generates active dependency junctions for root TypeScript tests.
- `testframework/pipeline/stages.yaml`: Declarative stage order, not an executor.
- `testframework/pipeline/jobs/*.yaml`: Declarative command/domain inventory, not consumed by CI code.
- `.gitignore`: Excludes `node_modules/`, Python caches, coverage and Playwright artifacts.

**Core Logic:**
- `backend/testframework/env/builder.go`: Ent/SQLite test environment.
- `backend/testframework/mongotest/setup.go`: Ephemeral Mongo database with cleanup.
- `testframework/stages/unit/frontend-test/render.tsx`: Shared React renderer/query-client.
- `testframework/stages/e2e/shared/`: Current E2E fixtures, helpers, Page Objects and Steps.

**Testing:**
- `backend/testframework/<domain>/**/*_test.go`: Current backend suite.
- `backend/cmd/server/wiring_fiber_recover_test.go`: Current co-located composition test outside canonical GitHub test command.
- `testframework/stages/unit/frontend-test/domains/<domain>/**/*.test.ts(x)`: Current frontend suite.
- `testframework/stages/e2e/domains/<domain>/**/*.spec.ts`: Current browser suite.
- `backend/tests/**`, `frontend/tests/**`, `testframework/**/*.py`: Founder target; not present in this snapshot.

## Current Inventory

Snapshot: HEAD `ee6d696c`, measured 2026-08-07; generated/vendor/node_modules paths are excluded.

| Test class | Path | Count |
|---|---|---:|
| Go tests total | `backend/**/*_test.go` | 830 |
| Go legacy root | `backend/testframework/**/*_test.go` | 829 |
| Go production domain tree | `backend/internal/**/*_test.go` | 0 |
| Go other | `backend/cmd/server/wiring_fiber_recover_test.go` | 1 |
| Go support, non-test | `backend/testframework/{env,factory,mongotest,steps}/*.go` | 4 |
| `export_hooks.go` | `backend/internal/**/export_hooks.go` | 39 |
| All `hubtest`-tagged production Go | `backend/internal/**/*.go` | 45 |
| Vitest total | `testframework/stages/unit/frontend-test/domains/**/*.test.ts(x)` | 180 |
| Vitest `.test.ts` | `testframework/stages/unit/frontend-test/domains/**/*.test.ts` | 60 |
| Vitest `.test.tsx` | `testframework/stages/unit/frontend-test/domains/**/*.test.tsx` | 120 |
| Frontend target tree | `frontend/tests/**/*.test.ts(x)` | 0 |
| Frontend source tree tests | `frontend/src/**/*.test.ts(x)` | 0 |
| Playwright TS specs | `testframework/stages/e2e/domains/**/*.spec.ts` | 20 |
| E2E support TS | shared tree + `frontend/playwright.config.ts` | 14 (13 + 1) |
| Root Python | `testframework/**/*.py` | 0 |
| Root tracked framework files | `testframework/**` via `git ls-files` | 241 |
| Root tracked stage files | `testframework/stages/**` via `git ls-files` | 224 |

### Backend domain counts

| Path | Tests | Path | Tests |
|---|---:|---|---:|
| `backend/testframework/accounting/` | 13 | `backend/testframework/ai/` | 44 |
| `backend/testframework/apigateway/` | 19 | `backend/testframework/audit/` | 38 |
| `backend/testframework/auth/` | 29 | `backend/testframework/backup/` | 15 |
| `backend/testframework/bases/` | 17 | `backend/testframework/calendar/` | 6 |
| `backend/testframework/calls/` | 35 | `backend/testframework/crm/` | 14 |
| `backend/testframework/dashboard/` | 11 | `backend/testframework/dev/` | 6 |
| `backend/testframework/ent/` | 13 | `backend/testframework/env/` | 1 |
| `backend/testframework/github/` | 22 | `backend/testframework/gitlab/` | 15 |
| `backend/testframework/hr/` | 10 | `backend/testframework/imex/` | 15 |
| `backend/testframework/intranet/` | 21 | `backend/testframework/mail/` | 7 |
| `backend/testframework/marketplace/` | 14 | `backend/testframework/messenger/` | 57 |
| `backend/testframework/notification/` | 44 | `backend/testframework/onec/` | 5 |
| `backend/testframework/platform/` | 49 | `backend/testframework/project/` | 39 |
| `backend/testframework/rbac/` | 13 | `backend/testframework/search/` | 37 |
| `backend/testframework/sprint/` | 23 | `backend/testframework/status/` | 2 |
| `backend/testframework/storage/` | 21 | `backend/testframework/tms/` | 77 |
| `backend/testframework/tracker/` | 80 | `backend/testframework/ws/` | 17 |

The table counts all 829 legacy-root test files. `backend/testframework/factory/`, `backend/testframework/mongotest/` and `backend/testframework/steps/` contain support code but no `*_test.go`.

### Frontend unit domain counts

| Path | Tests | Path | Tests |
|---|---:|---|---:|
| `testframework/stages/unit/frontend-test/domains/admin/` | 9 | `testframework/stages/unit/frontend-test/domains/ai/` | 6 |
| `testframework/stages/unit/frontend-test/domains/auth/` | 9 | `testframework/stages/unit/frontend-test/domains/calls/` | 10 |
| `testframework/stages/unit/frontend-test/domains/dashboard/` | 5 | `testframework/stages/unit/frontend-test/domains/hr/` | 1 |
| `testframework/stages/unit/frontend-test/domains/messenger/` | 27 | `testframework/stages/unit/frontend-test/domains/notifications/` | 1 |
| `testframework/stages/unit/frontend-test/domains/platform/` | 39 | `testframework/stages/unit/frontend-test/domains/search/` | 1 |
| `testframework/stages/unit/frontend-test/domains/tms/` | 10 | `testframework/stages/unit/frontend-test/domains/tracker/` | 44 |
| `testframework/stages/unit/frontend-test/domains/v4/` | 18 | — | — |

### E2E domain counts

| Path | Specs | Path | Specs |
|---|---:|---|---:|
| `testframework/stages/e2e/domains/admin/` | 3 | `testframework/stages/e2e/domains/ai/` | 2 |
| `testframework/stages/e2e/domains/auth/` | 3 | `testframework/stages/e2e/domains/messenger/` | 1 |
| `testframework/stages/e2e/domains/navigation/` | 3 | `testframework/stages/e2e/domains/platform/` | 5 |
| `testframework/stages/e2e/domains/tms/` | 1 | `testframework/stages/e2e/domains/tracker/` | 1 |
| `testframework/stages/e2e/domains/v4/` | 1 | — | — |

## Naming Conventions

**Files:**
- Go tests: `*_test.go`; domain package declarations are external (`package auth_test`, `package tracker_test`) throughout `backend/testframework/`.
- Test-only Go bridges: `export_hooks.go` and `export_hooks_<area>.go` under `backend/internal/**`, guarded by `//go:build hubtest`.
- Vitest: `*.test.ts` for non-React logic and `*.test.tsx` for React/UI under `testframework/stages/unit/frontend-test/domains/`.
- Playwright TS: `*.spec.ts` under `testframework/stages/e2e/domains/`.
- Shared frontend test utilities: descriptive filenames such as `render.tsx`; target these under `frontend/tests/support/`.
- Python target: use pytest discovery names `test_*.py` under `testframework/tests/{api,integration,contract,e2e}/`; keep Locust entry files under `testframework/load/` as defined when TASK-200 is implemented.

**Directories:**
- Current test directories group by product/CI domain: `backend/testframework/<domain>/`, `testframework/stages/**/domains/<domain>/`.
- Founder target keeps language-native trees under `backend/tests/<domain>/` and `frontend/tests/<domain>/`; QA-owned non-unit layers stay under root `testframework/<layer>/`.
- Preserve nested production mapping where needed: current `backend/testframework/onec/` maps to production `backend/internal/integrations/onec/`, and current `backend/testframework/platform/<subsystem>/` maps to `backend/internal/platform/<subsystem>/`.

## Where to Add New Code

**New Go Unit During Cutover:**
- Primary code: `backend/tests/<domain>/` only, as a genuine service-free Go
  unit with TASK/requirement/TC/risk metadata.
- Until target CI/launcher exists, mark runtime `NOT_READY`; never mirror a new
  test into frozen `backend/testframework/**`.
- Shared helpers belong to an explicitly accepted target support seam and must
  not retain service-backed legacy coupling.

**New Frontend Unit During Cutover:**
- Primary code: `frontend/tests/<domain>/` together with Vitest/TypeScript/Biome
  target configuration.
- Until the target launcher exists, mark runtime `NOT_READY`; never add or
  mirror tests under frozen `testframework/stages/unit/frontend-test/**`.
- Shared helpers live in `frontend/tests/support/` and switch atomically with
  their exact consumers.

**New Python Integration/API/E2E:**
- Integration: `testframework/tests/integration/<domain>/` after root
  `testframework/pyproject.toml` and fixtures exist.
- API: `testframework/tests/api/<domain>/`.
- Browser E2E: `testframework/tests/e2e/<domain>/`; browser lifecycle/artifacts live at exact task paths under `testframework/src/hub_testkit/browser/**`, and domain UI helpers under `testframework/src/hub_testkit/domains/<domain>/ui/{pages,components}/**`.
- Cross-domain full flow: pytest `full` marker over root Python tests.
- Load: `testframework/load/`; execute only against the approved local/scheduled target from `.planning/backlog/TASK-200-2026-07-30-tf-locust-load.md`.

**New Component/Module:**
- Go domain unit: mirror its production domain name under `backend/tests/` and keep imports directed toward `backend/internal/<domain>`.
- Frontend module unit: mirror the production area under `frontend/tests/`, but place by system under test rather than preserving broad legacy buckets such as `testframework/stages/unit/frontend-test/domains/v4/`.

**Utilities:**
- Go unit helpers: четыре legacy support-файла удаляются вместе с legacy closure. Если genuine unit требует helper, новый минимальный helper создаётся рядом с конкретным доменом по отдельному принятому плану; общего migrated `backend/tests/support/` нет.
- Frontend unit helpers: target `frontend/tests/support/`; dependencies resolve naturally through ancestor `frontend/node_modules` without root junctions.
- Python QA helpers: exact task-owned modules under `testframework/src/hub_testkit/**`; composition at `testframework/tests/conftest.py`; do not import TypeScript helper files from `testframework/stages/`.

## Migration Mapping Rules

| Current path | Target path | Rule |
|---|---|---|
| `backend/testframework/<domain>/**/*_test.go` | `backend/tests/<domain>/` | Only true units; retain external package imports and required hooks |
| `backend/testframework/platform/<subsystem>/` | `backend/tests/platform/<subsystem>/` | Preserve nested package mapping |
| `backend/testframework/onec/` | `backend/tests/integrations/onec/` | Match production path `backend/internal/integrations/onec/` |
| `backend/testframework/ent/` | exact Phase 129 `move` or `delete` operation | Только доказанные genuine units переносятся в `backend/tests/ent/`; остальные legacy rows удаляются, а Python contracts/integration создаются greenfield |
| `backend/testframework/env/`, `factory/` | delete by exact Phase 129 manifest | Оба legacy support-файла не переносятся и не становятся общей unit-support библиотекой |
| `backend/testframework/mongotest/` | delete by exact Phase 129 manifest | Legacy Mongo helper не портируется; Python integration support проектируется независимо |
| `backend/testframework/steps/` | delete by exact Phase 129 manifest | Legacy API steps не портируются; новый Python API layer выводится из requirements/risks/verified TC |
| 156 sealed unit rows under `testframework/stages/unit/frontend-test/domains/<domain>/` | `frontend/tests/<domain>/` | Move only literal unit allowlist; 24 static-contract rows delete, no Python port |
| `testframework/stages/unit/frontend-test/render.tsx` | `frontend/tests/support/render.tsx` | Move with alias/config consumers |
| `testframework/stages/e2e/domains/<domain>/*.spec.ts` | delete by exact Phase 129 manifest | Do not port TS files or preserve case counts; design Python journeys from current requirements/risks/TC |
| `testframework/stages/e2e/shared/` | delete by exact Phase 129 manifest | New Python POM/Steps/auth/data fixtures are greenfield artifacts, not migrated copies |
| `testframework/{stages,guides,pipeline}/` | delete by exact Phase 129 closure batches | TASK-219 owns immediate cleanup; no parity gate |

## Special Directories

**`frontend/node_modules/`:**
- Purpose: Installed Node dependencies for frontend, Vitest and Playwright.
- Generated: Yes.
- Committed: No.

**Generated junctions under `testframework/`:**
- Purpose: Allow Node resolution from external root TypeScript test files.
- **Historical/forbidden junction inventory:** `testframework/stages/e2e/node_modules`, `testframework/stages/unit/frontend-test/node_modules`, and two retired local junctions. None grants target ownership or execution authority.
- Generated: Yes; active two are created by `frontend/scripts/link-testframework-node-modules.mjs`.
- Committed: No.

**`frontend/{test-results,playwright-report,dist}/`:**
- Purpose: Runtime/build artifacts.
- Generated: Yes.
- Committed: No; excluded by `.gitignore` or untracked.

**`backend/ent/`:**
- Purpose: Ent generated client plus schema sources used by backend tests.
- Generated: Mostly; schema source is maintained under `backend/ent/schema/`.
- Committed: Partially/currently tracked; exclude generated Ent output from test inventory.

**`.gotmp/`:**
- Purpose: Local Go build/test cache workspace.
- Generated: Yes.
- Committed: No; excluded by `.gitignore`.

**`pro-qa-ai/`:**
- Purpose: Protected external founder toolkit, outside Hub test topology.
- Generated: Not assessed.
- Committed: Not assessed; directory was not read or counted.

## Reproducible Counts

Run from repository root. These commands exclude dependency/generated trees and do not traverse `pro-qa-ai/**`:

```powershell
# 829 current legacy backend tests; full tree adds 1 server test
@(rg --files backend/testframework -g '*_test.go' -g '!**/vendor/**' -g '!**/node_modules/**' -g '!**/generated/**').Count

# 0 production-tree tests; 1 other backend test
@(rg --files backend/internal -g '*_test.go' -g '!**/generated/**').Count
@(rg --files backend/cmd -g '*_test.go' -g '!**/generated/**').Count

# 180 Vitest files: 60 TS + 120 TSX
@(rg --files testframework/stages/unit/frontend-test/domains -g '*.test.ts' -g '*.test.tsx' -g '!**/node_modules/**').Count
@(rg --files testframework/stages/unit/frontend-test/domains -g '*.test.ts' -g '!**/node_modules/**').Count
@(rg --files testframework/stages/unit/frontend-test/domains -g '*.test.tsx' -g '!**/node_modules/**').Count

# 20 E2E TS; 0 Python
@(rg --files testframework/stages/e2e/domains -g '*.spec.ts' -g '!**/node_modules/**').Count
@(rg --files testframework -g '*.py' -g '!**/node_modules/**' -g '!**/generated/**').Count

# 39 canonical hook files; 241 tracked root framework files
@(rg --files backend/internal -g 'export_hooks.go' -g '!**/generated/**').Count
@(git ls-files -- testframework).Count
```

Do not reuse stale 812/824/828/829-total backend or 179 frontend counts; canonical full-tree total is 830 Go and 180 Vitest.

---

*Structure analysis: 2026-08-07*
