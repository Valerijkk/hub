---
context_schema: "1.0"
profile_slug: "hub"
status: filled
generated_at: "2026-07-26T02:25:00+03:00"
snapshot_commit: d4a44174a0642ec060e17277a61414e1bfe1ca3d
---

# Commands

Only exact commands discovered in checked-out files may appear here. Do not infer, repair, or execute a command. Execution requires separate, explicit human approval; invoking bootstrap or refresh is not approval.

## Discovered commands

| ID | Purpose | Exact command | Working directory | Discovery source | Execution status |
|---|---|---|---|---|---|
| CMD-001 | Backend test job | `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...` | `backend` | SRC-009:5; SRC-006:57-61 | not-run |
| CMD-002 | Link frontend test dependencies | `node scripts/link-testframework-node-modules.mjs` | `frontend` | SRC-010:5; SRC-006:154-156 | not-run |
| CMD-003 | Frontend TypeScript check | `npx tsc --noEmit` | `frontend` | SRC-010:7; SRC-006:150-152 | not-run |
| CMD-004 | Frontend unit tests | `npx vitest run` | `frontend` | SRC-010:8; SRC-006:158-160 | not-run |
| CMD-005 | Frontend lint gate | `npm run lint` | `frontend` | SRC-010:9; SRC-006:162-164 | not-run |
| CMD-006 | Playwright smoke suite | VERBATIM-BLOCK below | `frontend` | SRC-006:212-216 | not-run |
| CMD-007 | Playwright full suite | `npx playwright test -c playwright.config.ts --project=chromium-full` | `frontend` | SRC-006:269-273 | not-run |
| CMD-008 | Backend coverage command | `go test -tags hubtest -coverprofile=/tmp/cov -coverpkg=./internal/... ./testframework/...` | `backend` | SRC-013:7 | not-run |
| CMD-009 | One frontend domain | `npx vitest run ../testframework/stages/unit/frontend-test/domains/auth` | `frontend` | SRC-007:59-60; SRC-047:10-11 | not-run |
| CMD-010 | One backend domain | `go test -tags hubtest ./testframework/auth/...` | `backend` | SRC-007:62-63 | not-run |
| CMD-011 | Generate Ent code | `go generate ./ent` | `backend` | SRC-006:51-55 | not-run |
| CMD-012 | Vet backend with test exports | `go vet -tags hubtest ./...` | `backend` | SRC-006:51-55 | not-run |
| CMD-013 | Build production backend | `go build -tags release -o bin/server ./cmd/server` | `backend` | SRC-006:63-67 | not-run |
| CMD-014 | Install and build frontend | `npm ci && npm run build` | `frontend` | SRC-006:187-189 | not-run |
| CMD-015 | Start E2E infrastructure | `bash scripts/ci-e2e-backend.sh start-infra` | project root | SRC-006:257-258 | not-run |
| CMD-016 | Wait for E2E infrastructure | `bash scripts/ci-e2e-backend.sh wait-infra` | project root | SRC-006:260-261 | not-run |
| CMD-017 | Build and start E2E backend | `bash scripts/ci-e2e-backend.sh start-backend` | project root | SRC-006:263-264 | not-run |
| CMD-018 | Wait for E2E backend | `bash scripts/ci-e2e-backend.sh wait-backend` | project root | SRC-006:266-267 | not-run |
| CMD-019 | Stop E2E services | `bash scripts/ci-e2e-backend.sh stop` | project root | SRC-006:291-293 | not-run |
| CMD-020 | Install frontend dependencies and Chromium | `npm ci && npx playwright install chromium --with-deps` | `frontend` | SRC-006:204-206 | not-run |

### CMD-006 exact command

```text
npx playwright test -c playwright.config.ts --project=chromium-smoke --grep "login page|full page has no critical|path 1"
```

Source: SRC-006:212-216. Execution status: `not-run`.

Allowed execution statuses: `not-run`, `approved-and-passed`, `approved-and-failed`, `UNKNOWN`.

## Required prerequisites

| Command ID | Prerequisite | Evidence |
|---|---|---|
| CMD-001, CMD-010, CMD-012 | Go 1.25 and `hubtest` export hooks; MongoDB is required by some packages and unavailable Mongo tests may skip | SRC-001, SRC-021, SRC-025 |
| CMD-002–CMD-005, CMD-009, CMD-014 | Node 22 in CI; dependencies installed under `frontend`; link step required for tests outside `frontend` | SRC-006, SRC-003, SRC-021 |
| CMD-006, CMD-007, CMD-020 | Chromium installed; frontend available; full E2E additionally expects backend/dev-mode infrastructure | SRC-006, SRC-004, SRC-036 |
| CMD-008 | Test infrastructure started and `POSTGRES_TEST_URL` plus `MONGODB_TEST_URL` supplied by the environment | SRC-006, SRC-013 |
| CMD-015–CMD-019 | Checked-out `scripts/ci-e2e-backend.sh`; exact internal commands assembled by that script are UNKNOWN because the script was not executed | SRC-006 |

## Environment variable names

Record names and purpose only. Never copy values.

| Name | Required by | Source | Secret-sensitive |
|---|---|---|---|
| `MONGODB_TEST_URL` | Backend MongoDB tests / coverage | SRC-006:59-61,119-123; SRC-025:18-35 | yes |
| `POSTGRES_TEST_URL` | Backend coverage and PostgreSQL-backed tests | SRC-006:117-123 | yes |
| `CI` | Playwright worker/retry/`forbidOnly` behavior | SRC-004:23-25; SRC-006:212-216,269-273 | no |
| `PLAYWRIGHT_BASE_URL` | Browser base URL | SRC-004:16 | no |
| `PLAYWRIGHT_SKIP_WEBSERVER` | Whether Playwright starts the frontend server | SRC-004:17,45-55 | no |
| `PLAYWRIGHT_API_URL` | E2E backend helper base URL | SRC-006:224-235; SRC-037:4-8 | no |
| `DEV_MODE` | Full E2E backend/dev-login availability | SRC-006:224-235; SRC-036:37-42 | no |
| `DATABASE_URL` | Full E2E backend database connection | SRC-006:224-235 | yes |
| `MONGODB_URL` | Full E2E backend MongoDB connection | SRC-006:224-235 | yes |
| `MONGODB_DATABASE` | Full E2E backend database name | SRC-006:224-235 | no |
| `REDIS_URL` | Full E2E backend Redis connection | SRC-006:224-235 | yes |
| `NATS_URL` | Full E2E backend NATS connection | SRC-006:224-235 | yes |
| `JWT_SECRET` | Full E2E backend signing configuration | SRC-006:224-235; SRC-018:104-113 | yes |

## Selection patterns

Patterns must be discovered verbatim from project sources, not synthesized.

| Scope | Command ID or exact pattern | Evidence |
|---|---|---|
| Single test | UNKNOWN | UNKNOWN |
| Single file/package | UNKNOWN | UNKNOWN |
| Single frontend domain | CMD-009 | SRC-007, SRC-047 |
| Single backend domain | CMD-010 | SRC-007 |
| Smoke E2E project | CMD-006; `chromium-smoke` | SRC-004, SRC-011, SRC-006 |
| Full E2E project | CMD-007; `chromium-full` | SRC-004, SRC-012, SRC-006 |

## Execution evidence

This section is changed only after explicit approval to run a discovered command.

<!-- BEGIN MANUAL -->
UNKNOWN
<!-- END MANUAL -->
