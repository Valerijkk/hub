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
│   └── testframework/                 # 828 Go tests + 4 support .go files
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
│       ├── <domain>/                  # package <domain>_test; imports internal/<domain>
│       └── shared/                    # Go-only builders/factories used by units
├── frontend/
│   ├── src/                           # Production TS/React remains unchanged
│   └── tests/                         # Developer-owned Vitest tree
│       ├── <module>/                  # *.test.ts / *.test.tsx by system under test
│       └── shared/                    # setup/render/factories
└── testframework/                     # QA-owned standalone Python project
    ├── pyproject.toml                 # pytest, Playwright Python, Locust
    ├── conftest.py                    # Root fixtures
    ├── fixtures/                      # Data/auth/client factories
    ├── integration/                   # DB, Mongo, NATS, external service flows
    ├── api/                           # Whole HTTP API scenarios/contracts
    ├── e2e/                           # Playwright Python + POM/steps
    ├── full/                          # Cross-domain/system flows
    ├── load/                          # Locust profiles; local/scheduled only
    └── README.md                      # One-command entry points
```

`backend/tests/**` remains a separate Go package tree: it can import `backend/internal/**` because it is inside the module rooted at `backend/go.mod`, but it cannot access private parent-package symbols directly. Keep required `backend/internal/**/export_hooks*.go` under `hubtest` until their consumers are removed.

## Directory Purposes

**`backend/testframework/`:**
- Purpose: Current executable backend test root.
- Contains: 832 tracked Go files: 828 `*_test.go` and four reusable support files.
- Key files: `backend/testframework/env/builder.go`, `backend/testframework/factory/user.go`, `backend/testframework/mongotest/setup.go`, `backend/testframework/steps/http.go`.
- Placement rule: During migration, classify each test before moving. Mock/in-memory unit tests target `backend/tests/**`; DB/broker/external/whole-API tests require Python parity under `testframework/{integration,api}/` before removal.

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
- Contains: 180 tests, shared renderer, two fixture `.ts` files, two task-specific tsconfig files and README.
- Key files: `testframework/stages/unit/frontend-test/render.tsx`, `testframework/stages/unit/frontend-test/domains/auth/client-api.test.ts`, `testframework/stages/unit/frontend-test/domains/v4/`.
- Placement rule: Map by imported system under test into `frontend/tests/**`; split cross-domain aggregates instead of blindly preserving current folders.

**`frontend/src/test/`:**
- Purpose: Current frontend test runtime setup.
- Contains: `frontend/src/test/setup.ts` importing jest-dom matchers; no test cases.
- Key files: `frontend/src/test/setup.ts`.
- Placement rule: Move or re-export setup under `frontend/tests/shared/` only together with the `frontend/vite.config.ts` update.

**`testframework/stages/e2e/`:**
- Purpose: Current TypeScript Playwright suite.
- Contains: 20 specs under `testframework/stages/e2e/domains/`, 13 shared files under `testframework/stages/e2e/shared/`, active generated node_modules junction and one obsolete config shim.
- Key files: `testframework/stages/e2e/shared/fixtures.ts`, `testframework/stages/e2e/shared/helpers/auth.ts`, `testframework/stages/e2e/shared/pages/base.page.ts`, `testframework/stages/e2e/shared/steps/auth.steps.ts`.
- Placement rule: Preserve until Playwright Python parity. Port from this path, not from `frontend/e2e/`.

**`testframework/pipeline/`:**
- Purpose: Human-readable mirror of pipeline topology.
- Contains: `testframework/pipeline/stages.yaml` and five manifests in `testframework/pipeline/jobs/`.
- Key files: `testframework/pipeline/jobs/backend-test.yaml`, `testframework/pipeline/jobs/frontend-test.yaml`, `testframework/pipeline/jobs/e2e-full.yaml`.
- Placement rule: Treat `.github/workflows/ci.yml` as executable truth; delete/replace legacy manifests only in final cleanup after CI parity.

**`frontend/e2e/`:**
- Purpose: Compatibility documentation only.
- Contains: tracked `frontend/e2e/README.md`; no `*.spec.ts`.
- Key files: `frontend/e2e/README.md`.
- Placement rule: Do not use as migration source. Its pointer to `testframework/e2e/` is stale relative to `frontend/playwright.config.ts`, whose actual `testDir` is `testframework/stages/e2e/domains/`.

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

Snapshot: branch `cursor-grok-4.5`, HEAD `f7fbdd7e`, measured 2026-08-07. Test paths were clean at measurement time. Generated/vendor/node_modules paths are excluded.

| Test class | Path | Count |
|---|---|---:|
| Go tests total | `backend/**/*_test.go` | 829 |
| Go legacy root | `backend/testframework/**/*_test.go` | 828 |
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
| E2E shared TS | `testframework/stages/e2e/shared/**/*` | 13 |
| Root Python | `testframework/**/*.py` | 0 |
| Root tracked framework files | `testframework/**` via `git ls-files` | 241 |
| Root tracked stage files | `testframework/stages/**` via `git ls-files` | 224 |

### Backend domain counts

| Path | Tests | Path | Tests |
|---|---:|---|---:|
| `backend/testframework/accounting/` | 13 | `backend/testframework/ai/` | 44 |
| `backend/testframework/apigateway/` | 19 | `backend/testframework/audit/` | 37 |
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

The table counts all 828 test files. `backend/testframework/factory/`, `backend/testframework/mongotest/` and `backend/testframework/steps/` contain support code but no `*_test.go`.

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
- Shared frontend test utilities: descriptive filenames such as `render.tsx`; target these under `frontend/tests/shared/`.
- Python target: use pytest discovery names `test_*.py` under `testframework/{integration,api,e2e,full}/`; keep Locust entry files under `testframework/load/` as defined when TASK-200 is implemented.

**Directories:**
- Current test directories group by product/CI domain: `backend/testframework/<domain>/`, `testframework/stages/**/domains/<domain>/`.
- Founder target keeps language-native trees under `backend/tests/<domain>/` and `frontend/tests/<module>/`; QA-owned non-unit layers stay under root `testframework/<layer>/`.
- Preserve nested production mapping where needed: current `backend/testframework/onec/` maps to production `backend/internal/integrations/onec/`, and current `backend/testframework/platform/<subsystem>/` maps to `backend/internal/platform/<subsystem>/`.

## Where to Add New Code

**New Go Unit During Migration:**
- Primary code: `backend/tests/<domain>/` only after TASK-197 paths/CI are synced to the founder target.
- Until that CI path exists, add or mirror the regression in the currently executed `backend/testframework/<domain>/`; never leave the only test in an unexecuted target path.
- Shared helpers: `backend/tests/shared/` in target; migrate `backend/testframework/{env,factory}/` only after all consumers use the new import path.

**New Frontend Unit During Migration:**
- Primary code: `frontend/tests/<module>/` only together with a `frontend/vite.config.ts` include and `frontend/tsconfig.json`/test-tsconfig update.
- Until the target path is executed, keep the gate-bearing test under `testframework/stages/unit/frontend-test/domains/<domain>/`.
- Shared helpers: target `frontend/tests/shared/`; migrate `testframework/stages/unit/frontend-test/render.tsx` and `frontend/src/test/setup.ts` atomically with consumers.

**New Python Integration/API/E2E:**
- Integration: `testframework/python/integration/<domain>/` after `testframework/python/pyproject.toml` and root fixtures exist.
- API: `testframework/python/api/<domain>/`.
- Browser E2E: `testframework/python/e2e/<domain>/`, with shared POM/steps/fixtures inside `testframework/python/e2e/` or `testframework/python/fixtures/`.
- Cross-domain full flow: `testframework/python/full/`.
- Load: `testframework/load/`; execute only against the approved local/scheduled target from `.planning/backlog/TASK-200-2026-07-30-tf-locust-load.md`.

**New Component/Module:**
- Go domain unit: mirror its production domain name under `backend/tests/` and keep imports directed toward `backend/internal/<domain>`.
- Frontend module unit: mirror the production area under `frontend/tests/`, but place by system under test rather than preserving broad legacy buckets such as `testframework/stages/unit/frontend-test/domains/v4/`.

**Utilities:**
- Go unit helpers: target `backend/tests/shared/`; avoid importing Python framework code.
- Frontend unit helpers: target `frontend/tests/shared/`; dependencies resolve naturally through ancestor `frontend/node_modules` without root junctions.
- Python QA helpers: `testframework/fixtures/` and layer-local `conftest.py`; do not import TypeScript helper files from `testframework/stages/`.

## Migration Mapping Rules

| Current path | Target path | Rule |
|---|---|---|
| `backend/testframework/<domain>/**/*_test.go` | `backend/tests/<domain>/` | Only true units; retain external package imports and required hooks |
| `backend/testframework/platform/<subsystem>/` | `backend/tests/platform/<subsystem>/` | Preserve nested package mapping |
| `backend/testframework/onec/` | `backend/tests/integrations/onec/` | Match production path `backend/internal/integrations/onec/` |
| `backend/testframework/ent/` | classify between `backend/tests/ent/` and `testframework/python/integration/` | Schema/source checks may stay Go; real DB checks belong to Python integration |
| `backend/testframework/env/`, `factory/` | `backend/tests/shared/` | Go unit support only |
| `backend/testframework/mongotest/` | `testframework/python/integration/fixtures/` equivalent | Real Mongo dependency is integration, not unit support |
| `backend/testframework/steps/` | classify | Fiber-in-memory unit helper may stay Go; whole-API steps belong to Python `api/` |
| `testframework/stages/unit/frontend-test/domains/<domain>/` | `frontend/tests/<module>/` | Re-map by SUT import; split cross-domain files |
| `testframework/stages/unit/frontend-test/render.tsx` | `frontend/tests/shared/render.tsx` | Move with alias/config consumers |
| `testframework/stages/e2e/domains/<domain>/*.spec.ts` | `testframework/python/e2e/<domain>/test_*.py` | Preserve scenario/assertion parity before deleting TS source |
| `testframework/stages/e2e/shared/` | `testframework/python/e2e/` + `testframework/python/fixtures/` | Recreate POM/Steps/auth/data fixtures in Python |
| `testframework/{stages,guides,pipeline}/` | delete only after parity | Cleanup belongs to TASK-201, not scaffold |

## Special Directories

**`frontend/node_modules/`:**
- Purpose: Installed Node dependencies for frontend, Vitest and Playwright.
- Generated: Yes.
- Committed: No.

**Generated junctions under `testframework/`:**
- Purpose: Allow Node resolution from external root TypeScript test files.
- Paths: `testframework/stages/e2e/node_modules`, `testframework/stages/unit/frontend-test/node_modules`; stale local junctions also exist at `testframework/e2e/node_modules` and `testframework/frontend/node_modules`.
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
# 828 current backend tests
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

Do not reuse the stale 812/824 backend or 179 frontend counts in `AGENTS.md`, `docs/test_plan.md`, `.planning/STATE.md` and `.planning/backlog/TASK-197-2026-07-30-tf-backend-units-colocated.md`; recount at the start of each migration wave.

---

*Structure analysis: 2026-08-07*
