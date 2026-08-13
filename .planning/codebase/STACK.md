# Technology Stack

> **Superseded inventory note (2026-08-07):** старые counts ниже — historical
> map. Текущий cutover authority: founder manifest, 1,049 rows / 4,246 declared,
> revision `b9f690da`, SHA `83b6a311…` (ротация 2026-08-13); без parity/dual-run/82% migration gate.

**Analysis Date:** 2026-08-07

**Scope:** Test stack only: `backend/`, `frontend/`, root `testframework/`, CI configuration, coverage, and the migration boundary requested by the founder.

## Current-State Snapshot

**Measured at HEAD `ee6d696c` (2026-08-07):**

| Area | Current physical location | Measured inventory | Executed by |
|------|---------------------------|--------------------|-------------|
| Go tests | legacy root + server | **830** total / **3,667** tests: root 829/3,665 + server 1/2; taxonomy 594 unit + 236 component | `.github/workflows/ci.yml` job `backend-test` |
| Frontend tests | legacy Vitest root | **180** files / **519** cases: 156 unit + 24 static contract; 3 support rows | `frontend/vite.config.ts` + `frontend-test` |
| Browser E2E | legacy Playwright root + config | **20** specs / **53** cases + 14 support rows | `frontend/playwright.config.ts` + E2E jobs |
| Python system tests | `testframework/` | **0** `.py` files; no `pyproject.toml`, `pytest.ini`, requirements file, or Python lockfile | Not detected |
| Test pipeline metadata | `testframework/pipeline/` | 5 job manifests plus `stages.yaml` | Declarative only; no non-document consumer was found |

**Canonical inventory:** ledger HEAD `ee6d696c` / SHA-256 `01dae4e2…` содержит 1,047 rows: 830 Go,
180 Vitest + 3 support, 20 E2E + 14 support и 0 Python. Старые
812/824/828/829-total и 179 claims — dated snapshots.

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
- Go standard `testing` - runner for 829 legacy files under `backend/testframework/`; отдельный server test доводит total до 830.
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
- PostgreSQL/MongoDB/Redis/NATS and Chromium profiles require the future
  versioned launcher API solely under `testframework/runner/**`, with immutable
  assets from `testframework/docker/**`, limits, isolated reports, and exact
  cleanup. Until separate implementation and independent acceptance, launcher
  status is `NOT_READY` and runtime is `NOT_RUN`.
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
- Place only genuine service-free Go units in `backend/tests/**`; delete legacy
  DB/broker/network/process/whole-API component rows by literal manifest.
- New `testframework/tests/{contract,api,integration,e2e}/**` suites are derived
  independently from requirements, risks and verified TC.
- Because `backend/tests/**` remains inside the `backend` Go module, it can legally import `backend/internal/*`. Keep black-box `package <domain>_test` where public contracts suffice; use same-package tests only when an explicitly colocated white-box exception is approved.

**Frontend-native unit tree:**
- Move exactly 156 TypeScript/React units into `frontend/tests/<domain>/**` and make `frontend/tsconfig.json`, `frontend/vite.config.ts`, and the Biome command include that tree directly.
- Move only legacy `render.tsx` into `frontend/tests/support/render.tsx`; delete the other two legacy support rows. Any new setup/fixtures are designed explicitly rather than copied from the old tree.

**Independent root system framework:**
- Reserve `testframework/src/hub_testkit/**` for reusable installed modules;
  planned `testframework/tests/{contract,api,integration,e2e}/**` for SUT
  suites; separate `testframework/{selftests,cases,schemas,policies,load}/**`
  roots for their accepted responsibilities; `testframework/docker/**` for
  immutable execution assets only; future `testframework/runner/**` for sole
  versioned launcher authority; and `testframework/reports/<task-run>/**` for
  owner-marked evidence. If required, `testframework/tests/conftest.py` is
  composition-only. Package, target suites, and launcher are `NOT_READY`;
  runtime is `NOT_RUN`.
- Exact Python version, environment manager, pytest plugin set, and load tool are **not specified by the current repository** and must be selected and pinned by the planner before implementation.
- Phase 129 deletes legacy non-units/static-contract/TypeScript-E2E paths by
  exact manifest and advances the structure-only scaffold toward a separately
  accepted package bootstrap. Only genuine native units move; no scenario-count
  or report parity is required, and launcher implementation remains a separate
  future slice under `testframework/runner/**`.

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
13. **Outer launcher missing** - the future versioned launcher API is owned
    only by `testframework/runner/**`; `testframework/docker/**` is
    immutable-assets-only. Until separate implementation and independent
    acceptance, launcher is `NOT_READY`, runtime is `NOT_RUN`, and inline
    Docker/legacy runners are not fallback.

## Prescriptive Migration Contract

- Seal the Phase 128 source ledger, operation partitions and exact cutover manifest before moving the first test; historical Docker coverage remains discovery evidence only and no direct coverage command is a cutover gate.
- Use the sealed classification only to select a literal operation: genuine native unit `move`, protected unit `keep-blocked`, otherwise legacy `delete`. Python is never the destination of an old source row.
- Move classified Go units in exact domain batches into `backend/tests/**`;
  delete component/API/integration rows in their manifest batches without
  one-to-one Python replacement.
- Move exactly 156 frontend units plus `render.tsx`; delete 24 static-contract files and the remaining support/config closure. Expand TypeScript/Biome scopes and remove `@testframework`/the linker in their exact atomic consumer-switch batches.
- Create an independent Python project under `testframework/` from current
  requirements/risks/verified TC. Future package authority owns environment,
  lockfile, and pytest/Playwright configuration; only future
  `testframework/runner/**` owns the versioned launcher API, artifacts,
  orchestration, first-exit preservation, and exact cleanup, while
  `testframework/docker/**` owns immutable execution assets only. Launcher is
  `NOT_READY` and runtime is `NOT_RUN` until separate implementation and
  independent acceptance.
- Do not dual-run or preserve historical row/case counts. Each new required job
  must instead prove non-zero collection, expected-skip policy, traceability,
  oracle sensitivity and cleanup.
- After separate implementation and independent acceptance, GitHub Actions and
  every retained CI mirror must route to the same future API under
  `testframework/runner/**`; `testframework/docker/**` remains
  immutable-assets-only. No provider-specific test command or fallback may
  become a second execution authority; current launcher is `NOT_READY` and
  runtime is `NOT_RUN`.
- Preserve the 82.0373% result only as dated discovery evidence. Publish a new
  native-unit baseline after cutover; historical coverage is not a migration
  gate and its denominator is not preserved.

## Historical mapping evidence — NON-EXECUTABLE

The commands below document the old mapper snapshot only. Do not execute them
as current test/build/validator guidance; target runtime requires the accepted
future API under `testframework/runner/**`; `testframework/docker/**` remains
immutable-assets-only, launcher is `NOT_READY`, and runtime is `NOT_RUN`.

```powershell
# Current inventories
(rg --files backend/testframework -g '*_test.go').Count                         # 829 legacy; +1 server = 830 total
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
