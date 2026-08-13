# External Integrations

> **Superseded inventory note (2026-08-07):** legacy counts и parity wording
> ниже не управляют cutover. Текущий authority — founder manifest на revision
> `b9f690da` (ротация 2026-08-13): legacy non-units delete-not-port, новые non-unit suites greenfield
> Python, 245 genuine hook-dependent units сохраняются через TASK-234.

**Analysis Date:** 2026-08-07

**Scope:** Historical execution/reporting map. Target boundary: Go units in
`backend/tests/**`, frontend units in `frontend/tests/**`, and greenfield Python
system framework in `testframework/`; no parity acceptance requirement.

## Current Integration Map

| Boundary | Current owner | Current transport | Current gap |
|----------|---------------|-------------------|-------------|
| Go test execution | legacy root + server | `go test` from `.github/workflows/ci.yml` | 830 files / 3,667 tests: root 829/3,665 + server 1/2; 594 unit + 236 component |
| Frontend unit execution | `testframework/stages/unit/frontend-test/` | Vitest borrowed from `frontend/node_modules` | 180 files / 519 cases: 156 unit + 24 static contract; 3 support rows |
| Browser E2E | `testframework/stages/e2e/` | Playwright borrowed from `frontend/package.json` | 20 specs / 53 cases + 14 support rows; no independent runtime contract |
| System integration | Split across `backend/testframework/` and `testframework/stages/e2e/` | Go integration helpers plus TypeScript Playwright | No pytest orchestration layer in `testframework/` |
| Test reporting | GitHub artifacts, Playwright HTML, console output | GitHub Actions artifact upload | No cross-runner JUnit/Allure merge or TMS publication |
| CI metadata | `testframework/pipeline/jobs/*.yaml` | None | Metadata is not consumed by `.github/workflows/ci.yml` or `.gitlab-ci.yml` |

## APIs & External Services

**CI providers:**
- GitHub Actions - authoritative executable pipeline in `.github/workflows/ci.yml`; runs on push/pull request to `master` and manual dispatch.
- GitLab CI - intended mirror in `.gitlab-ci.yml`, but commands and gates differ materially from GitHub Actions.
- GitHub Container Registry/deploy - downstream of successful CI in `.github/workflows/deploy.yml`; test mapping does not modify deployment.

**Browser automation:**
- Chromium via Playwright - installed during `e2e-smoke` and `e2e-full` in `.github/workflows/ci.yml`.
  - SDK/Client: `@playwright/test` resolved to 1.61.0 in `frontend/package-lock.json`.
  - Config: `frontend/playwright.config.ts`.
  - Auth: application dev-session flow through the backend dev endpoint; URL is selected with `PLAYWRIGHT_API_URL` in `testframework/stages/e2e/shared/helpers/backend.ts`.
- Axe accessibility - `@axe-core/playwright` resolved to 4.11.3 and used by `testframework/stages/e2e/domains/platform/a11y.spec.ts`.

**Application external services exercised indirectly:**
- LiveKit - `scripts/ci-e2e-backend.sh` starts the `livekit` service from `docker-compose.local.yml`, but the 20-spec E2E inventory in `testframework/stages/e2e/domains/` has no dedicated voice-call spec; `docs/test_plan.md` classifies voice E2E as partial.
- Resend, Telegram, GitHub, and GitLab - application clients exist, but `.github/workflows/ci.yml` does not configure live provider credentials for tests. Coverage is through local Go doubles/handler tests under paths such as `backend/testframework/auth/`, `backend/testframework/notification/`, `backend/testframework/github/`, and `backend/testframework/gitlab/`.
- Hub TMS result import - CLI support exists at `cli/cmd/push_results.go`, but `.github/workflows/ci.yml`, `.gitlab-ci.yml`, and `testframework/pipeline/` do not invoke it.

**Allure distinction:**
- Allure parsing/storage is application functionality under `backend/internal/tms/allure_parser.go`, `backend/internal/tms/allure_matcher.go`, and `backend/internal/tms/allure_types.go`.
- Allure is **not** configured as a Go/Vitest/Playwright reporter: no runner adapter appears in `frontend/package.json`, `frontend/playwright.config.ts`, or `.github/workflows/ci.yml`.

## Data Storage

**Databases:**
- SQLite in-process - Ent component tests use `github.com/mattn/go-sqlite3`; shared builder is `backend/testframework/env/builder.go`.
  - Connection: generated in-memory DSNs inside Go test helpers.
  - Client: Ent `enttest` from `backend/ent/enttest`.
- PostgreSQL 17 - real-engine tests and full application E2E.
  - Connection variable: `POSTGRES_TEST_URL` for Go integration tests in `backend/testframework/`.
  - Client: `database/sql`, `lib/pq`, and Ent in tests such as `backend/testframework/tms/postgres_ent_helper_test.go`.
  - Provisioning: `deployments/docker-compose.test.yml` for `coverage-gate`; `docker-compose.local.yml` through `scripts/ci-e2e-backend.sh` for `e2e-full`.
- MongoDB 8/8.0 - backend integration tests and application E2E.
  - Connection variable: `MONGODB_TEST_URL` for Go tests; application E2E uses its application Mongo configuration from `.github/workflows/ci.yml`.
  - Client: MongoDB Go driver v2; shared optional helper `backend/testframework/mongotest/setup.go` creates and drops an ephemeral database.
- Redis 7.4-alpine - sessions/cache/presence in coverage and full E2E; in-process unit doubles use Miniredis from `backend/go.mod`.
- NATS 2.11 - application messaging in coverage/full E2E; test doubles or embedded clients are used in `backend/testframework/platform/` and domain tests.

**Job-to-service matrix:**

| Job | Declared infrastructure | Evidence |
|-----|-------------------------|----------|
| `backend-test` | MongoDB only | service block in `.github/workflows/ci.yml` |
| `coverage-gate` | PostgreSQL, MongoDB, Redis, NATS | `deployments/docker-compose.test.yml` invoked by `.github/workflows/ci.yml` |
| `frontend-test` | None | `.github/workflows/ci.yml` |
| `e2e-smoke` | Vite frontend + Chromium; no backend | `.github/workflows/ci.yml` and `frontend/playwright.config.ts` |
| `e2e-full` | PostgreSQL, MongoDB, Redis, NATS, LiveKit + native backend + Vite frontend | `scripts/ci-e2e-backend.sh`, `docker-compose.local.yml`, `.github/workflows/ci.yml` |

**Current service mismatch:**
- At least five Go integration helpers hard-fail without PostgreSQL: `backend/testframework/accounting/pay_concurrent_a6002_test.go`, `backend/testframework/sprint/repository_get_in_project_a3001_test.go`, `backend/testframework/project/a3w1_a5_task137_test.go`, `backend/testframework/imex/pdf_scope_a5010_ent_test.go`, and `backend/testframework/tms/postgres_ent_helper_test.go`.
- The authoritative `backend-test` job in `.github/workflows/ci.yml` provisions only MongoDB. Unless its runner supplies PostgreSQL out of band, its service contract does not match the current test tree.
- Other helpers skip when services are unavailable, for example `backend/testframework/tracker/postgres_ent_helper_test.go` and `backend/testframework/mongotest/setup.go`; a green run can therefore represent a different executed test set across environments.

**File Storage:**
- SeaweedFS/S3-compatible storage is defined in `docker-compose.local.yml` and application integration code, but `scripts/ci-e2e-backend.sh` does not start the `seaweedfs` service for `e2e-full`.
- Test artifacts are local runner files: `frontend/playwright-report/`, `.ci-backend.log`, and Lighthouse JSON under `frontend/`, as configured in `.github/workflows/ci.yml`.

**Caching:**
- Application Redis is available to coverage/full E2E through `deployments/docker-compose.test.yml` and `docker-compose.local.yml`.
- GitHub dependency caches are managed by `actions/setup-go` and `actions/setup-node` in `.github/workflows/ci.yml`.
- Pytest cache/environment caching is not applicable because no Python framework exists in `testframework/`.

## Authentication & Identity

**Auth Provider:**
- Custom Hub Magic Link/session authentication in `backend/internal/auth/`; no external identity provider is configured for CI tests.
  - Implementation: full Playwright paths use a development-only application endpoint discovered through `testframework/stages/e2e/shared/helpers/backend.ts` and guarded by `skipUnlessDevMode` in `testframework/stages/e2e/shared/fixtures.ts`.
  - Smoke limitation: `e2e-smoke` in `.github/workflows/ci.yml` deliberately runs without backend infrastructure, so it validates frontend-only login/a11y/critical-page behavior rather than the production Magic Link exchange.
  - Environment sensitivity: many E2E specs call `skipUnlessBackend` and `skipUnlessDevMode` from `testframework/stages/e2e/shared/fixtures.ts`; CI's pre-test health/dev-mode checks reduce broad skips, while optional AI/provider cases still contain explicit skips under `testframework/stages/e2e/domains/ai/ai-hub.spec.ts`.

## Monitoring & Observability

**Error Tracking:**
- External error-tracking service - not detected in the test pipeline files `.github/workflows/ci.yml`, `.gitlab-ci.yml`, or `testframework/pipeline/`.

**Logs:**
- Go/Vitest results are console-only in `.github/workflows/ci.yml`; no structured report is uploaded for these jobs.
- Full E2E uploads `frontend/playwright-report/` and `.ci-backend.log` only when the job fails, as configured in `.github/workflows/ci.yml`.
- Playwright captures trace on first retry and screenshots on failure through `frontend/playwright.config.ts`.
- Lighthouse uploads `frontend/lighthouse-report.json` and `frontend/lighthouse-summary.json` on non-PR runs; its job is `continue-on-error` and its script failure is explicitly tolerated in `.github/workflows/ci.yml`.
- Backend coverage is computed from a temporary profile in `.github/workflows/ci.yml`; the profile itself is not uploaded as an artifact.

**Missing unified evidence:**
- JUnit XML, Allure runner results, pytest reports, Vitest coverage, and a cross-runtime test summary are not configured in `.github/workflows/ci.yml`.
- `cli/cmd/push_results.go` can upload Allure JSON to Hub TMS, but no current CI job produces and sends that directory.

## CI/CD & Deployment

**Hosting:**
- GitHub Actions runners execute the canonical pipeline in `.github/workflows/ci.yml`.
- GitLab runners have a partial mirror in `.gitlab-ci.yml`.
- Docker Compose provides service dependencies through `deployments/docker-compose.test.yml` and `docker-compose.local.yml`.

**CI Pipeline:**
- `backend-test` - Go generate/vet, all current backend tests with `hubtest` and race detection, then release build; source `.github/workflows/ci.yml`.
- `frontend-test` - npm install, product TypeScript check, root dependency links, Vitest, and product-source Biome; source `.github/workflows/ci.yml`.
- `build` - waits for both unit jobs and builds backend/frontend; source `.github/workflows/ci.yml`.
- `e2e-smoke` - waits only for frontend tests and runs 3 frontend-only specs; source `.github/workflows/ci.yml` and `frontend/playwright.config.ts`.
- `e2e-full` - waits for backend/frontend tests, starts infrastructure and dev backend, then runs all 20 specs; source `.github/workflows/ci.yml` and `scripts/ci-e2e-backend.sh`.
- `coverage-gate` - non-PR only, starts the four-service test compose and enforces 82%; source `.github/workflows/ci.yml`.
- `lighthouse` - non-PR, non-blocking artifact job; source `.github/workflows/ci.yml`.

**Pipeline contract drift:**
- `.gitlab-ci.yml` backend job omits `-tags hubtest` and targets `./...`, so it does not intentionally execute the 829-test legacy root at `backend/testframework/`.
- `.gitlab-ci.yml` frontend job omits `frontend/scripts/link-testframework-node-modules.mjs`; its lint command is allowed to fail, unlike GitHub Actions.
- `.gitlab-ci.yml` smoke job refers to stale `frontend/e2e/*.spec.ts` paths rather than `testframework/stages/e2e/domains/`.
- `testframework/pipeline/stages.yaml` refers to `build` and `lighthouse`, but `testframework/pipeline/jobs/` contains only `backend-test.yaml`, `frontend-test.yaml`, `e2e-smoke.yaml`, `e2e-full.yaml`, and `coverage-gate.yaml`.
- No executable consumer of `testframework/pipeline/jobs/*.yaml` was found; `.github/workflows/ci.yml` and `.gitlab-ci.yml` are maintained separately.

## Environment Configuration

**Required env vars:**
- Go integration: `POSTGRES_TEST_URL`, `MONGODB_TEST_URL` - referenced by `.github/workflows/ci.yml` and `backend/testframework/`.
- Playwright routing: `PLAYWRIGHT_BASE_URL`, `PLAYWRIGHT_API_URL`, `PLAYWRIGHT_SKIP_WEBSERVER`, `CI` - referenced by `frontend/playwright.config.ts`, `testframework/stages/e2e/shared/helpers/backend.ts`, and `.github/workflows/ci.yml`.
- Full application E2E: application database, cache, messaging, authentication, storage, and LiveKit variable names are declared in `.github/workflows/ci.yml` and `scripts/ci-e2e-backend.sh`; values must remain CI/local secrets or non-production fixtures and must not be copied into planning documents.
- Target Python framework: variable schema not detected; define typed/defaulted profiles under `testframework/` rather than reading arbitrary root environment state.

**Secrets location:**
- `.env.local`, `.env.local.example`, `frontend/.env.example`, and `deployments/.env.example` are present; existence only was inspected, not contents.
- Workflow-scoped test configuration is declared in `.github/workflows/ci.yml`; production deployment secrets belong to GitHub Actions/deployment configuration, not the test framework.
- The target `testframework/` must accept secret names/references only and keep values out of tracked config, reports, screenshots, traces, and logs.

## Webhooks & Callbacks

**Incoming:**
- The test harness itself exposes no callback endpoint under `testframework/`.
- Application webhook handlers are covered by Go tests under `backend/testframework/apigateway/`, `backend/testframework/tms/`, `backend/testframework/github/`, and `backend/testframework/gitlab/`.

**Outgoing:**
- GitHub artifact upload occurs through `actions/upload-artifact` in `.github/workflows/ci.yml`.
- Hub TMS result upload is available manually through `cli/cmd/push_results.go` but is not wired into CI.
- No automatic external notification or test-result webhook was detected in `.github/workflows/ci.yml`, `.gitlab-ci.yml`, or `testframework/pipeline/`.

## Target Integration Boundary

**Native unit jobs:**
- `backend/tests/` and `frontend/tests/` do not yet exist at the analyzed revision; all target job/path definitions are prospective.
- `backend/tests/<domain>/**` owns only service-free Go unit tests and remains inside the
  `backend` Go module so it can legally import `backend/internal/*`; legacy
  component/integration rows are literal manifest deletes. New Python coverage
  is designed independently from requirements, risks and verified TC.
- Exactly 156 legacy Vitest/RTL units move to `frontend/tests/<domain>/**` and
  one `render.tsx` moves to `frontend/tests/support/**`. The 24 static contracts
  and remaining two support rows are deleted; any later helpers are greenfield.

**Independent Python system framework:**
- `testframework/` owns pytest-based system/integration/load orchestration and Playwright integration, with its own `pyproject.toml`, dependency lock, test configuration, fixtures, service profiles, markers, artifacts, and CI entry point.
- Exact Python version, package manager, pytest plugins, Playwright binding, and load tool are not present in the current repository and require explicit architecture decisions before scaffolding.
- Delete current TypeScript Playwright and its support/config entrypoints in
  Phase 129 by exact manifest. Design the Python E2E layer independently from
  current requirements, risks and verified TC; historical rows are not parity
  authority.

**Exact-manifest removal rule:**
- Phase 129 moves only proven genuine native units and deletes the complete
  non-unit/static/TypeScript-E2E closure without port or dual-run.
- A consumer/config/hook file is deleted only by its literal manifest batch
  after zero-consumer and identity checks. The 245 hook-dependent units and 43
  hook files remain protected until TASK-234 evidence releases them.

## Migration Blockers and Integration Risks

1. **No runnable Python environment** - neither `python` nor the Windows `py` launcher is available in the current PATH, and `testframework/` has no Python project metadata.
2. **Historical daemon observation only** - the mapping snapshot could not reach
   Docker. The future versioned launcher API belongs only to
   `testframework/runner/**`; `testframework/docker/**` contains only immutable
   execution assets. Launcher status is `NOT_READY` and runtime is `NOT_RUN`
   until separate implementation and independent acceptance.
3. **Backend job/service mismatch** - current `backend-test` in `.github/workflows/ci.yml` lacks PostgreSQL for hard-fail tests under `backend/testframework/`.
4. **Different jobs execute different suites** - service-sensitive Go tests may
   hard-fail or skip depending on infrastructure; they cannot move to the
   service-free `backend/tests/**` lane and require Python profiles.
5. **Cross-root Node dependency ownership** - both Vitest and Playwright currently borrow `frontend/node_modules` through `frontend/scripts/link-testframework-node-modules.mjs`; the target Python system framework must not silently retain this coupling.
6. **No unit-test static-analysis coverage** - `frontend/tsconfig.json` and `npm run lint` currently scope to `frontend/src`, while unit tests live under root `testframework/`; moving to `frontend/tests/**` must expand both gates.
7. **E2E uses dev-only identity** - `testframework/stages/e2e/shared/fixtures.ts` and `testframework/stages/e2e/shared/helpers/backend.ts` depend on backend dev mode; production-auth/system profiles need separate markers and credentials policy.
8. **Weak skip semantics in legacy inventory** - service/provider availability
   can turn scenarios into skips. New required jobs fail on unexpected skip and
   zero collection; legacy pass/fail/skip counts are not a migration gate.
9. **Incomplete environment fidelity** - `scripts/ci-e2e-backend.sh` starts LiveKit but not SeaweedFS, and waits only for PostgreSQL; system coverage of storage/media is not guaranteed by the current full-E2E label.
10. **No common report contract** - Go, Vitest, Playwright, and future pytest results cannot currently be merged or uploaded to TMS despite `cli/cmd/push_results.go`.
11. **Three-way CI drift** - `.github/workflows/ci.yml`, `.gitlab-ci.yml`, and `testframework/pipeline/jobs/*.yaml` cannot be treated as equivalent migration gates.
12. **Cutover atomicity** - legacy and target trees must not become competing
   authorities. Each Phase 129 batch switches exact consumers with its
   filesystem operation and remains independently revertible.

## Prescriptive Handoff for Planning

- Define the target job contract table for `backend-unit`, `frontend-unit`, `python-contract`, `python-api`, `python-integration`, `python-e2e-smoke`, `python-e2e-full`, and `python-load`; map each to paths, services, markers/tags, timeout, artifacts, and required status.
- Provision Python and Docker in the implementation environment before accepting a `testframework/` scaffold as runnable.
- Treat legacy service observations as discovery only; create fail-closed target
  profiles for required PostgreSQL/Mongo/Redis/NATS boundaries.
- Add structured per-task-run reports and fail on zero collection, unexpected
  skip, missing owner marker, insensitive sentinel or leaked resource.
- Reserve the future versioned launcher API solely for `testframework/runner/**`:
  it will own profile resolution, orchestration, first-exit preservation,
  task-run report ownership, and exact cleanup. `testframework/docker/**` stays
  immutable-assets-only. `/qa-gate` remains user-facing routing only and cannot
  execute a job while launcher status is `NOT_READY` and runtime is `NOT_RUN`;
  no direct command or alternative fallback is supported before separate
  implementation and independent acceptance.
- Wire TMS publication only after report format and redaction rules are stable; `cli/cmd/push_results.go` is an available integration point, not current evidence.

## Historical mapping evidence — NON-EXECUTABLE

Команды ниже сохранены только как снимок старой mapping-сессии. Они не являются
текущей инструкцией: validators/tests/builds/linters запускаются только через
будущий API под `testframework/runner/**`; `testframework/docker/**` immutable-assets-only,
launcher `NOT_READY`, runtime `NOT_RUN`.

```powershell
# Service definitions can be parsed without starting containers
docker compose -f deployments/docker-compose.test.yml config --services
# Outcome: mongo, nats, postgres, redis

docker compose -f docker-compose.local.yml config --services
# Outcome: mongo, nats, postgres, seaweedfs, redis, hub-mail, backend, frontend, livekit

# Runtime availability
python --version
py --version
# Outcome: neither command is available in PATH

docker info --format '{{.ServerVersion}}'
# Outcome: Docker daemon connection unavailable

# Historical CI and report integration inventory
rg -n 'go test|vitest|playwright|coverage|upload-artifact' .github/workflows/ci.yml .gitlab-ci.yml
rg -n 'push-results|allure-results' cli .github/workflows/ci.yml testframework/pipeline
# Outcome: TMS upload exists in CLI, not in CI

# Manifest consumers
rg -n --glob '!*.md' 'testframework/pipeline|pipeline/jobs|stages.yaml' .
# Outcome: no executable consumer found
```

No full regression or service-backed test run was executed for this map.

---

*Integration audit: 2026-08-07*
