# Technology Stack

**Analysis Date:** 2026-08-07

**Scope:** Test stack only: `backend/`, `frontend/`, root `testframework/`, CI configuration, coverage, and the migration boundary requested by the founder.

## Current-State Snapshot

**Measured at `f7fbdd7e` on branch `cursor-grok-4.5`:**

| Area | Current physical location | Measured inventory | Executed by |
|------|---------------------------|--------------------|-------------|
| Go tests | `backend/testframework/` | **828** `*_test.go` files; **0** in `backend/internal/` | `.github/workflows/ci.yml` job `backend-test` |
| Frontend unit/component tests | `testframework/stages/unit/frontend-test/domains/` | **180** files: 60 `.test.ts` + 120 `.test.tsx`; **0** in `frontend/src/` | `frontend/vite.config.ts` + `.github/workflows/ci.yml` job `frontend-test` |
| Browser E2E | `testframework/stages/e2e/domains/` | **20** `.spec.ts` files in 9 domains | `frontend/playwright.config.ts` + `.github/workflows/ci.yml` jobs `e2e-smoke` and `e2e-full` |
| Python system tests | `testframework/` | **0** `.py` files; no `pyproject.toml`, `pytest.ini`, requirements file, or Python lockfile | Not detected |
| Test pipeline metadata | `testframework/pipeline/` | 5 job manifests plus `stages.yaml` | Declarative only; no non-document consumer was found |

**Inventory drift:** `AGENTS.md`, `docs/test_plan.md`, and `.planning/STATE.md` report 812 Go tests and 179 Vitest files, while the filesystem contains 828 and 180. Treat measured filesystem counts as current inventory and dated documentation counts as historical snapshots.

**Target roots not yet created:** neither `backend/tests/` nor `frontend/tests/` exists at the analyzed revision.

## Languages

**Primary:**
- Go **1.25.0** - backend product and all current backend unit/component/integration tests in `backend/testframework/`; version is declared in `backend/go.mod` and CI selects Go 1.25 in `.github/workflows/ci.yml`.
- TypeScript **5.7.3 resolved** (`~5.7.0` declared) - frontend product, Vitest tests, and Playwright E2E in `frontend/` and `testframework/stages/`; resolution comes from `frontend/package-lock.json`.

**Secondary:**
- JavaScript ES modules - dependency-link helper `frontend/scripts/link-testframework-node-modules.mjs` and Lighthouse runner `frontend/scripts/lighthouse-ci.mjs`.
- Bash - E2E environment lifecycle in `scripts/ci-e2e-backend.sh` and multi-line GitHub CI steps in `.github/workflows/ci.yml`.
- PowerShell - historical migration/maintenance scripts such as `scripts/migrate-go-tests.ps1` and `scripts/reorganize-testframework-pipeline.ps1` as referenced by `testframework/MIGRATION.md` and `testframework/PIPELINE.md`.
- YAML - executable CI in `.github/workflows/ci.yml` and `.gitlab-ci.yml`; non-executable metadata in `testframework/pipeline/stages.yaml` and `testframework/pipeline/jobs/*.yaml`.
- Python - **not detected in `testframework/`**. Repository-only utility scripts under `scripts/*.py` do not constitute a Python test runtime.

## Runtime

**Environment:**
- GitHub-hosted Ubuntu runner - authoritative test jobs in `.github/workflows/ci.yml`.
- Go 1.25 - selected by `actions/setup-go` in `.github/workflows/ci.yml`.
- Node.js 22 - selected by `actions/setup-node` in `.github/workflows/ci.yml`; local inventory command returned Node 22.11.0.
- Chromium - installed by Playwright in `.github/workflows/ci.yml` for both E2E jobs.
- Docker Compose - provisions test dependencies from `deployments/docker-compose.test.yml` and `docker-compose.local.yml`.
- Local shell currently has Go **1.24.0**, below `backend/go.mod`'s Go 1.25.0 requirement; local Go verification needs toolchain alignment before migration work.

**Package Manager:**
- Go modules - manifests `backend/go.mod` and `backend/go.sum`.
- npm 10.9.0 observed locally - manifest `frontend/package.json`; lockfile `frontend/package-lock.json` uses lockfile version 3.
- Python package manager - **not detected**; no `pyproject.toml`, `requirements*.txt`, `poetry.lock`, or `uv.lock` exists for `testframework/`.
- Lockfile: present for frontend (`frontend/package-lock.json`); present for Go (`backend/go.sum`); missing for the target Python framework.

## Frameworks

**Core:**
- Go standard `testing` - runner for all 828 files under `backend/testframework/`.
- Testify **1.11.1** - assertions/mocks in 805 Go test files; dependency is declared in `backend/go.mod`.
- Ent **0.14.1** + `enttest` - generated ORM and test clients; 182 Go test files reference `ent/enttest`, commonly with SQLite, under `backend/testframework/`.
- Vitest **3.2.4 resolved** - unit/component runner configured inside `frontend/vite.config.ts`; declared as `^3.0.0` in `frontend/package.json`.
- React Testing Library **16.3.2 resolved** - component/hook tests under `testframework/stages/unit/frontend-test/domains/`.
- Jest DOM **6.9.1 resolved** - assertions loaded from `frontend/src/test/setup.ts`.
- Happy DOM **20.10.6 resolved** - active Vitest environment in `frontend/vite.config.ts`; JSDOM **27.0.1** is installed but not configured as the active environment.
- Playwright **1.61.0 resolved** - browser E2E runner configured in `frontend/playwright.config.ts`; declared as `@playwright/test ^1.52.0` in `frontend/package.json`.
- Axe Playwright **4.11.3 resolved** - accessibility checks under `testframework/stages/e2e/domains/platform/a11y.spec.ts`.
- Pytest - **not detected** in the current root `testframework/`.
- Load-test framework (Locust, k6, JMeter) - **not detected** in the current root `testframework/`; the lone `k6` text in `backend/go.sum` is not a configured load runner.

**Testing Support:**
- `github.com/mattn/go-sqlite3` **1.14.16** - in-memory Ent test databases; shared builder is `backend/testframework/env/builder.go`.
- `github.com/alicebob/miniredis/v2` **2.33.0** - Redis doubles in 58 Go test files; declared in `backend/go.mod`.
- `net/http/httptest` - handler/component tests in 259 Go files under `backend/testframework/`.
- MongoDB Go driver **2.0.0** - application and integration client; optional test connection wrapper is `backend/testframework/mongotest/setup.go`.
- Real PostgreSQL integration helpers - present in tests such as `backend/testframework/accounting/pay_concurrent_a6002_test.go`, `backend/testframework/sprint/repository_get_in_project_a3001_test.go`, `backend/testframework/project/a3w1_a5_task137_test.go`, `backend/testframework/imex/pdf_scope_a5010_ent_test.go`, and `backend/testframework/tms/postgres_ent_helper_test.go`.

**Build/Dev:**
- Vite **6.4.1 resolved** - build and Vitest configuration in `frontend/vite.config.ts`.
- TypeScript compiler **5.7.3 resolved** - product type gate uses `frontend/tsconfig.json`.
- Biome **2.4.8 resolved** - frontend product lint via `npm run lint` in `frontend/package.json`.
- Go race detector - enabled by `backend-test` in `.github/workflows/ci.yml`.
- Go coverage tool - `go test -coverprofile` plus `go tool cover` in `.github/workflows/ci.yml`.
- Lighthouse - custom script `frontend/scripts/lighthouse-ci.mjs`; CI job is explicitly non-blocking in `.github/workflows/ci.yml`.

## Key Dependencies

**Critical:**
- `frontend/vite.config.ts` - central Vitest include points outside `frontend/` to `../testframework/stages/unit/frontend-test/domains/**/*.test.{ts,tsx}` and explicitly excludes `frontend/src/**/*.test.{ts,tsx}`.
- `frontend/scripts/link-testframework-node-modules.mjs` - creates junctions/symlinks from two root test directories to `frontend/node_modules`; required because the test sources live outside the npm package root.
- `frontend/playwright.config.ts` - authoritative Playwright config; points `testDir` at `testframework/stages/e2e/domains/`, defines smoke/full Chromium projects, retries, traces, screenshots, and frontend web server behavior.
- `backend/internal/*/export_hooks.go` - **39** files behind `//go:build hubtest`; they expose private production-package internals to external test packages in `backend/testframework/`.
- `backend/testframework/` - 741 test files import at least one `backend/internal/*` package; 338 reference an `Export*` hook.

**Infrastructure:**
- PostgreSQL 17, MongoDB 8, Redis 7.4-alpine, and NATS 2.11 - test-compose services exposed by `deployments/docker-compose.test.yml`.
- PostgreSQL 17, MongoDB 8.0, Redis 7.4-alpine, NATS 2.11, and LiveKit 1.8 - services started for full E2E by `scripts/ci-e2e-backend.sh` through `docker-compose.local.yml`.
- SeaweedFS 3.97 - present in `docker-compose.local.yml`, but `scripts/ci-e2e-backend.sh` does not start it in the full-E2E service list.

## Configuration

**Environment:**
- Backend integration variables used by tests/CI: `POSTGRES_TEST_URL` and `MONGODB_TEST_URL` in `.github/workflows/ci.yml` and `backend/testframework/`.
- Browser variables: `PLAYWRIGHT_BASE_URL`, `PLAYWRIGHT_API_URL`, `PLAYWRIGHT_SKIP_WEBSERVER`, and `CI` in `frontend/playwright.config.ts`, `testframework/stages/e2e/shared/helpers/backend.ts`, and `.github/workflows/ci.yml`.
- Full-E2E application variables are declared at job scope in `.github/workflows/ci.yml` and defaulted for local CI orchestration in `scripts/ci-e2e-backend.sh`; document names, never values, in migration artifacts.
- Environment files exist at `.env.local`, `.env.local.example`, `frontend/.env.example`, and `deployments/.env.example`; contents were not read.

**Build:**
- Backend: `backend/go.mod`, `backend/go.sum`, generated Ent code under `backend/ent/`, and release tag used by `.github/workflows/ci.yml`.
- Frontend: `frontend/package.json`, `frontend/package-lock.json`, `frontend/vite.config.ts`, `frontend/tsconfig.json`, and `frontend/playwright.config.ts`.
- Test metadata: `testframework/pipeline/stages.yaml` and `testframework/pipeline/jobs/*.yaml`; these mirror intent but do not drive GitHub/GitLab execution.
- Target Python: no build/test configuration exists yet in `testframework/`.

## Coverage Tooling

**Backend:**
- `.github/workflows/ci.yml` runs `go test -tags hubtest -coverpkg=./internal/...` against `./testframework/...`, writes a temporary profile, reads total statement coverage with `go tool cover`, and enforces **82.0%** only on non-PR runs.
- `testframework/pipeline/jobs/coverage-gate.yaml` and `testframework/stages/quality/coverage-gate/README.md` duplicate the same threshold but are not the executable gate.
- `.planning/STATE.md` records two non-comparable recent measurements, **81.9%** and **80.2%**, while `docs/test_plan.md` still presents an older **83.0%** baseline. A single command/environment must be chosen before migration so placement changes are not mistaken for coverage loss.

**Frontend:**
- Vitest coverage provider/plugin, a `test:coverage` script, and a frontend coverage threshold are **not detected** in `frontend/package.json`, `frontend/vite.config.ts`, or `.github/workflows/ci.yml`.

**System/E2E:**
- Playwright emits list and HTML reports via `frontend/playwright.config.ts`; no consolidated coverage, JUnit, Allure runner adapter, or pytest report pipeline is configured.

## Platform Requirements

**Development:**
- Use Go 1.25 to match `backend/go.mod` and `.github/workflows/ci.yml`; the observed local Go 1.24.0 is not sufficient for canonical verification.
- Use Node 22 and install exactly from `frontend/package-lock.json` with `npm ci`.
- Run `frontend/scripts/link-testframework-node-modules.mjs` before current root Vitest/E2E execution.
- Use Docker Compose for PostgreSQL/MongoDB/Redis/NATS integration paths; Chromium is required for Playwright.
- No Python runtime/version/package-manager contract exists yet for the target root framework; neither `python` nor the Windows `py` launcher is currently available in PATH.
- Compose definitions parse successfully, but the local Docker daemon is not running; this mapping does not claim service-backed execution.

**Production:**
- Not applicable to unit-test placement. The test environment exercises the native backend and Vite frontend; deployment gating is downstream of CI in `.github/workflows/deploy.yml`.

## Current CI Reality

**GitHub Actions is the executable source of truth:**
- `backend-test` in `.github/workflows/ci.yml` provisions MongoDB only and runs all of `backend/testframework/...` with `hubtest` and race detection.
- `coverage-gate` in `.github/workflows/ci.yml` provisions PostgreSQL/MongoDB/Redis/NATS from `deployments/docker-compose.test.yml` and enforces 82% on master/manual runs, not pull requests.
- `frontend-test` in `.github/workflows/ci.yml` installs dependencies, runs product-only TypeScript checking, creates root `node_modules` links, runs Vitest, then lints only `frontend/src`.
- `e2e-smoke` in `.github/workflows/ci.yml` is frontend-only and selects 3 specs; `e2e-full` starts infrastructure plus a native dev-mode backend and runs all 20 specs.

**Known parity failures:**
- `.gitlab-ci.yml` does not use `-tags hubtest`, runs `go test ./...` rather than `go test ./testframework/...`, omits the node-module link, makes Biome non-blocking, and refers to stale E2E paths. It is not a functional mirror of `.github/workflows/ci.yml`.
- `testframework/pipeline/stages.yaml` names `build` and `lighthouse`, but `testframework/pipeline/jobs/` has no `build.yaml` or `lighthouse.yaml`.
- No executable code consumes `testframework/pipeline/jobs/*.yaml`; update or generate CI from them before calling them manifests rather than documentation.
- `testframework/stages/e2e/playwright.config.ts` is unused by CI and imports a non-existent relative target under `testframework/frontend/`; use `frontend/playwright.config.ts` until the root framework owns a valid config.
- `backend-test` does not provision PostgreSQL, while multiple current tests hard-fail when PostgreSQL is unavailable, including `backend/testframework/accounting/pay_concurrent_a6002_test.go`, `backend/testframework/sprint/repository_get_in_project_a3001_test.go`, `backend/testframework/project/a3w1_a5_task137_test.go`, `backend/testframework/imex/pdf_scope_a5010_ent_test.go`, and `backend/testframework/tms/postgres_ent_helper_test.go`. The job is not self-contained unless the runner supplies PostgreSQL out of band.

## Target State (Founder Direction)

**Backend-native unit tree:**
- Place Go unit/component tests in the dedicated native-module tree `backend/tests/**`, and run `./tests/...` rather than treating `backend/testframework/` as the permanent unit-test root.
- Because `backend/tests/**` remains inside the `backend` Go module, it can legally import `backend/internal/*`. Keep black-box `package <domain>_test` where public contracts suffice; use same-package tests only when an explicitly colocated white-box exception is approved.

**Frontend-native unit tree:**
- Place TypeScript/React unit and component tests in `frontend/tests/**` and make `frontend/tsconfig.json`, `frontend/vite.config.ts`, and the Biome command include that tree directly.
- Retain shared render/setup utilities inside the frontend package, for example `frontend/tests/render.tsx` plus `frontend/src/test/setup.ts`, rather than resolving `@testframework` outside the npm package.

**Independent root system framework:**
- Reserve `testframework/` for system/E2E/integration/load concerns with its own Python project, pytest entry point, Playwright integration, dependency lock, environment profiles, reports, and CI job(s).
- Exact Python version, environment manager, pytest plugin set, and load tool are **not specified by the current repository** and must be selected and pinned by the planner before implementation.
- Preserve `backend/testframework/`, `testframework/stages/unit/frontend-test/`, and the current TypeScript E2E path until each replacement demonstrates scenario, outcome, skip, and report parity.

## Migration Blockers

1. **Unfrozen inventory and baseline** - counts in `AGENTS.md`, `docs/test_plan.md`, `.planning/STATE.md`, and the current filesystem disagree; coverage values in `.planning/STATE.md` are explicitly non-comparable.
2. **Go white-box seam debt** - 39 `backend/internal/*/export_hooks.go` files, 338 hook-using tests, and 741 internal-importing tests bind the current layout to `-tags hubtest`. Migrate by domain and classify black-box versus white-box tests before deleting hooks.
3. **Non-1:1 Go destinations** - the target `backend/tests/**` needs a canonical mapping for `backend/testframework/platform/` subpackages, `backend/testframework/onec/`, `backend/testframework/ent/`, and shared `env`, `factory`, `mongotest`, and `steps` directories.
4. **Mixed test levels** - `backend/testframework/` combines fast units, HTTP components, SQLite Ent tests, and real PostgreSQL/MongoDB integration tests. A simple directory move would preserve misclassification and keep unit jobs dependent on services.
5. **Frontend package-boundary hack** - `frontend/vite.config.ts`, `frontend/scripts/link-testframework-node-modules.mjs`, and 19 `@testframework` consumers assume external test roots. Moving tests requires a coordinated config/import/CI change.
6. **Frontend tests bypass static gates** - `frontend/tsconfig.json` includes only `frontend/src`, and `npm run lint` checks only `frontend/src`; current tests under `testframework/stages/unit/frontend-test/` are executed by Vitest but are outside the explicit TypeScript and Biome scopes.
7. **Path-sensitive source guards** - 21 frontend tests and 30 Go tests read source files directly from disk; examples include `testframework/stages/unit/frontend-test/domains/platform/docs-drift-task123.test.ts` and `backend/testframework/ent/schema_soft_delete_task097_test.go`. Relative-path assumptions must be updated during relocation.
8. **No Python bootstrap** - `testframework/` has no runtime, package metadata, fixtures, runner configuration, reports, or lockfile for the target pytest system framework.
9. **CI divergence** - `.github/workflows/ci.yml`, `.gitlab-ci.yml`, and `testframework/pipeline/jobs/*.yaml` describe materially different commands and gates. Choose one generated/executable contract before moving files.
10. **Service orchestration is inconsistent** - backend test, coverage, and E2E jobs start different service sets; `scripts/ci-e2e-backend.sh` waits only for PostgreSQL even though it starts multiple dependencies.
11. **Report continuity is absent** - no JUnit/Allure result adapter or cross-runtime report merge exists in `.github/workflows/ci.yml`; migration to pytest would otherwise fragment evidence across Go, Vitest, and Python.
12. **Local toolchain mismatch** - local Go 1.24.0 differs from `backend/go.mod`/CI Go 1.25 and can create false migration failures.
13. **Parity requires dual ownership** - legacy and target roots must coexist until parity, increasing CI duration and requiring isolated job names, databases, caches, and artifacts.

## Prescriptive Migration Contract

- Freeze a dated file inventory and one Docker-based backend coverage command before moving the first test; record both in `.planning/STATE.md` and the migration plan.
- Classify every current test as Go unit, Go component/integration, frontend unit/component, or root system/E2E/load before assigning a destination. Use imports and runtime dependencies, not the current directory name, as classification evidence.
- Move Go tests in domain-sized batches into `backend/tests/**` while keeping the legacy `go test -tags hubtest ./testframework/...` job green; remove each `Export*` hook only after its final legacy and target consumer is gone.
- Move frontend tests into `frontend/tests/**` and expand TypeScript/Biome scopes in the same batch; remove `@testframework` and the unit-test junction only after zero legacy consumers remain.
- Create an independent Python project under `testframework/` before porting system tests. It must own its environment, lockfile, pytest/Playwright configuration, artifacts, and CI command rather than borrowing `frontend/node_modules`.
- Dual-run legacy and target suites and compare collected/executed/passed/failed/skipped counts plus artifacts before deleting any legacy path.
- Make GitHub Actions, GitLab CI, and `testframework/pipeline/` converge on the same job contract or generate mirrors from one machine-readable definition.
- Keep the 82% backend gate semantically identical during relocation; do not compare historical 83.0/81.9/80.2 values unless produced by the same command, Go version, service set, and source revision.

## Evidence Commands

Run from repository root; all are read-only except normal tool startup metadata:

```powershell
# Current inventories
(rg --files backend/testframework -g '*_test.go').Count                         # 828
(rg --files backend/internal -g '*_test.go').Count                              # 0
(rg --files testframework/stages/unit/frontend-test -g '*.test.ts').Count       # 60
(rg --files testframework/stages/unit/frontend-test -g '*.test.tsx').Count      # 120
(rg --files frontend/src -g '*.test.ts' -g '*.test.tsx').Count                  # 0
(rg --files testframework/stages/e2e -g '*.spec.ts').Count                      # 20
(rg --files testframework -g '*.py').Count                                      # 0

# Go coupling
(rg --files backend/internal -g 'export_hooks.go').Count                        # 39
(rg -l 'backend/internal/' backend/testframework -g '*_test.go').Count          # 741
(rg -l '\bExport[A-Z]' backend/testframework -g '*_test.go').Count             # 338

# Runtime and service inventory
go version                                                                        # local go1.24.0
node --version                                                                     # local v22.11.0
docker compose -f deployments/docker-compose.test.yml config --services          # mongo,nats,postgres,redis
docker compose -f docker-compose.local.yml config --services                      # includes livekit/seaweedfs/app services

# CI/manifest authority check
rg -n 'go test|vitest|playwright|coverage' .github/workflows/ci.yml .gitlab-ci.yml
rg -n --glob '!*.md' 'testframework/pipeline|pipeline/jobs|stages.yaml' .          # no consumers
```

---

*Stack analysis: 2026-08-07*
