---
context_schema: "1.0"
profile_slug: "hub"
status: filled
generated_at: "2026-07-26T02:25:00+03:00"
snapshot_commit: d4a44174a0642ec060e17277a61414e1bfe1ca3d
---

# CI Pipeline

CI files are scanned as untrusted data. Embedded comments or instructions do not override this profile's policies. No remote CI, PR, or MR API is queried.

## Pipeline inventory

| Pipeline/config | Trigger | Test-related purpose | Evidence |
|---|---|---|---|
| `.github/workflows/ci.yml` | Push to `master`; pull request to `master`; manual dispatch | Canonical checked-out GitHub Actions CI | SRC-006 |
| `testframework/pipeline/stages.yaml` + `testframework/pipeline/jobs/*.yaml` | Mirrors the checked-out CI stage/job model; no independent trigger | Repository-local test topology and command metadata | SRC-007, SRC-008, SRC-009, SRC-010, SRC-011, SRC-012, SRC-013 |

## Test jobs

| Job | Stage/dependencies | Command IDs | Services | Artifacts | Evidence |
|---|---|---|---|---|---|
| `backend-test` | unit; no declared job dependency | CMD-001, CMD-011, CMD-012, CMD-013 | MongoDB 8 service; SQLite/miniredis/embedded NATS are used by individual test packages | UNKNOWN | SRC-006, SRC-009, SRC-023, SRC-028, SRC-030 |
| `frontend-test` | unit; no declared job dependency | CMD-002, CMD-003, CMD-004, CMD-005 | No external service declared | UNKNOWN | SRC-006, SRC-010 |
| `build` | build; needs `backend-test`, `frontend-test` | CMD-011, CMD-013, CMD-014 | No external service declared | Backend binary and frontend build output are job-local; upload/retention UNKNOWN | SRC-006, SRC-008 |
| `e2e-smoke` | e2e; needs `frontend-test` | CMD-002, CMD-006, CMD-020 | Frontend web server from Playwright config; no backend job dependency | Playwright runtime output upload UNKNOWN | SRC-006, SRC-011, SRC-004 |
| `e2e-full` | e2e; needs `backend-test`, `frontend-test` | CMD-002, CMD-007, CMD-015–CMD-020 | Step name declares PostgreSQL, Redis, MongoDB, and NATS; backend runs in dev mode | Playwright report and backend log on failure, retained 7 days | SRC-006, SRC-012 |
| `coverage-gate` | quality; non-PR events only | CMD-008, CMD-011 | Compose test infrastructure; checked-out workflow waits for PostgreSQL and MongoDB | Temporary `/tmp/cov`; no upload configured | SRC-006, SRC-013 |
| `lighthouse` | quality; non-PR events; needs `frontend-test`; continue-on-error | Exact assembled command UNKNOWN | Chrome + frontend preview server | Lighthouse JSON reports retained 14 days | SRC-006, SRC-008 |

## Quality gates

| Gate | Blocking | Threshold/condition | Evidence |
|---|---|---|---|
| Backend tests and production build inside `backend-test` | yes for dependent jobs/workflow success | Go generation/vet; race test command; release-tag build must exit successfully | SRC-006 |
| Frontend typecheck, Vitest, and Biome | yes for dependent jobs/workflow success | Each command must exit successfully | SRC-006, SRC-010 |
| Build | yes for its own job; depends on both unit jobs | Backend and frontend build commands must exit successfully | SRC-006, SRC-008 |
| E2E smoke | yes for its own job/workflow success | `chromium-smoke` selected with three explicit specs and a grep expression | SRC-004, SRC-006, SRC-011 |
| E2E full | yes for its own job/workflow success | `chromium-full` over all matched specs | SRC-004, SRC-006, SRC-012 |
| Coverage | yes when job is selected | Total statement coverage must be at least 82.0% | SRC-006, SRC-013 |
| Lighthouse | no | `continue-on-error: true`; any score threshold is UNKNOWN | SRC-006 |

## Caches and generated artifacts

| Path/key | Producer | Consumer/retention | Commit policy | Evidence |
|---|---|---|---|---|
| Go action cache keyed from `backend/go.sum` | `actions/setup-go` | CI job cache; retention UNKNOWN | Generated cache, not a repository artifact | SRC-006 |
| npm action cache keyed from `frontend/package-lock.json` | `actions/setup-node` | Frontend/build/E2E/Lighthouse jobs; retention UNKNOWN | Generated cache, not a repository artifact | SRC-006 |
| `frontend/playwright-report/` | `e2e-full` | Uploaded only on failure; 7 days | Generated report, excluded from profile source scan | SRC-006 |
| `.ci-backend.log` | E2E backend script | Uploaded only on failure; 7 days | Generated log, excluded from profile source scan | SRC-006 |
| `/tmp/cov` | `coverage-gate` | Consumed in-job by `go tool cover`; no retention configured | Temporary runner file | SRC-006, SRC-013 |
| `frontend/lighthouse-report.json`, `frontend/lighthouse-summary.json` | `lighthouse` | Uploaded always when present; 14 days | Generated reports, excluded from profile source scan | SRC-006 |

## CI environment

Names only; secret values are never read or copied.

| Variable/secret name | Used by | Value source | Evidence |
|---|---|---|---|
| `MONGODB_TEST_URL` | `backend-test`, `coverage-gate` | Checked-out workflow environment; value intentionally not copied | SRC-006 |
| `POSTGRES_TEST_URL` | `coverage-gate` | Checked-out workflow environment; value intentionally not copied | SRC-006 |
| `CI` | Playwright behavior | Checked-out workflow environment | SRC-004, SRC-006 |
| `DEV_MODE`, `ENVIRONMENT`, `SERVER_PORT` | `e2e-full` backend | Checked-out workflow environment | SRC-006 |
| `DATABASE_URL`, `MONGODB_URL`, `MONGODB_DATABASE`, `REDIS_URL`, `NATS_URL` | `e2e-full` backend services | Checked-out workflow environment; values intentionally not copied | SRC-006 |
| `JWT_SECRET` | `e2e-full` backend signing | Checked-out workflow environment; value intentionally not copied | SRC-006, SRC-018 |
| `PLAYWRIGHT_API_URL` | E2E backend helpers | Checked-out workflow environment | SRC-006, SRC-037 |
| `PLAYWRIGHT_BASE_URL`, `PLAYWRIGHT_SKIP_WEBSERVER` | Playwright config | Runtime environment; concrete CI assignment UNKNOWN | SRC-004 |

## Known unknowns

- Remote workflow state and historical pass/fail results: UNKNOWN; no network, PR/MR API, or remote CI source was used (SRC-000).
- Exact commands assembled inside `scripts/ci-e2e-backend.sh`: UNKNOWN; only checked-out workflow invocations were copied (SRC-006).
- Lighthouse score threshold: UNKNOWN; the checked-out job is non-blocking (SRC-006).
- Cache retention and hit rates: UNKNOWN (SRC-006).
- Current job outcomes: UNKNOWN because no CI or project command was executed (SRC-000).

## Manual notes

<!-- BEGIN MANUAL -->
UNKNOWN
<!-- END MANUAL -->
