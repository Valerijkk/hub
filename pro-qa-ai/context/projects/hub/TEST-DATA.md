---
context_schema: "1.0"
profile_slug: "hub"
status: filled
generated_at: "2026-07-26T02:25:00+03:00"
snapshot_commit: d4a44174a0642ec060e17277a61414e1bfe1ca3d
---

# Test Data

Document data mechanisms and safety properties only from observed sources. Never copy credentials, personal data, production samples, or `.env` values.

## Data builders and fixtures

| Mechanism | Path | Scope/lifetime | Override strategy | Evidence |
|---|---|---|---|---|
| Ent/SQLite environment builder | `backend/testframework/env/builder.go` | Per test; creates user, project, and workflow status; client closed through `t.Cleanup` | Prefix is configurable; `WithUser` arguments are currently discarded, so user-field override behavior is not implemented in the observed source | SRC-023 |
| Ent user factory | `backend/testframework/factory/user.go` | Per factory/test context | Name prefix supplied by caller; email receives a UUID suffix | SRC-024 |
| Mongo integration database helper | `backend/testframework/mongotest/setup.go` | Unique database per test helper call; dropped during cleanup | Database prefix supplied by caller; connection URI selected from environment or a built-in test fallback whose value is intentionally not copied | SRC-025 |
| Frontend provider renderer | `testframework/stages/unit/frontend-test/render.tsx` | New QueryClient per render by default | Caller may supply a QueryClient or wrapper | SRC-032 |
| E2E availability fixtures | `testframework/stages/e2e/shared/fixtures.ts` | Per Playwright test fixture | Availability is detected; test may conditionally skip | SRC-036, SRC-037 |
| E2E dev seed/session injection | `testframework/stages/e2e/shared/helpers/auth.ts` | Dev data seeded through backend; returned session is injected before page code runs | Caller selects a synthetic email and API URL; concrete values are intentionally not copied | SRC-038 |

## Backing services

| Service | Test implementation | Initialization | Cleanup | Evidence |
|---|---|---|---|---|
| SQLite / Ent | In-memory SQLite through `enttest.Open` | Unique DSN per builder/RBAC setup | Ent client closed through `t.Cleanup` | SRC-023, SRC-029 |
| Redis | In-process miniredis with Go Redis client | Started in helper/test setup | Server and client closed through `t.Cleanup` | SRC-028 |
| MongoDB | External MongoDB selected by `MONGODB_TEST_URL` or test fallback | Client connect/ping; unique database name | Database drop and client disconnect through `t.Cleanup`; test skips if unavailable | SRC-025, SRC-031 |
| NATS JetStream | Embedded NATS server with JetStream and temporary store | Server starts on an ephemeral port | Server shutdown and temp-dir cleanup through `t.Cleanup` | SRC-030 |
| Full CI/E2E infrastructure | Checked-out CI invokes Compose or `ci-e2e-backend.sh`; observed service names are PostgreSQL, Redis, MongoDB, and NATS | CMD-015/CMD-016 or Compose start/wait steps | CMD-019 or Compose `down -v` | SRC-006 |

## Isolation and cleanup

| Concern | Observed mechanism | Risk/exception | Evidence |
|---|---|---|---|
| Unique identifiers | UUID-backed SQLite DSNs, test emails, Mongo database names, and domain IDs | Repository-wide uniqueness is UNKNOWN outside sampled helpers | SRC-023, SRC-024, SRC-025, SRC-031 |
| Transaction/reset | Per-test SQLite clients; Mongo database drop; miniredis process teardown; frontend session state reset in sampled store tests | A global database reset strategy for all PostgreSQL-backed tests is UNKNOWN | SRC-023, SRC-025, SRC-028, SRC-033 |
| Parallel test safety | Shared guide requires isolated environments before `t.Parallel`; sampled builders use unique resources; Playwright CI uses one worker | Repository-wide compliance is UNKNOWN | SRC-046, SRC-023, SRC-025, SRC-004 |
| Time/randomness control | miniredis `FastForward` controls TTL; UUID/ObjectID values isolate data | Other sampled tests use wall-clock deadlines, so global deterministic clock control is UNKNOWN | SRC-028, SRC-023, SRC-025, SRC-045 |

## Sensitive-data controls

| Control | Status | Evidence |
|---|---|---|
| Synthetic data only | UNKNOWN | UNKNOWN |
| Secret values excluded from this profile | true | SRC-000, SRC-018 |
| Logged data redaction | Required by project guidance for tokens, passwords, magic-link codes, and JWTs | SRC-017, SRC-018 |
| Production data prohibition in tests | UNKNOWN | UNKNOWN |

## Environment variable names

| Name | Purpose | Secret-sensitive | Evidence |
|---|---|---|---|
| `MONGODB_TEST_URL` | Mongo integration connection | yes | SRC-025, SRC-006 |
| `POSTGRES_TEST_URL` | Coverage/PostgreSQL test connection | yes | SRC-006 |
| `DATABASE_URL` | Full E2E backend database connection | yes | SRC-006 |
| `MONGODB_URL` | Full E2E backend Mongo connection | yes | SRC-006 |
| `MONGODB_DATABASE` | Full E2E Mongo database name | no | SRC-006 |
| `REDIS_URL` | Full E2E backend Redis connection | yes | SRC-006 |
| `NATS_URL` | Full E2E backend NATS connection | yes | SRC-006 |
| `JWT_SECRET` | Full E2E signing configuration | yes | SRC-006, SRC-018 |
| `PLAYWRIGHT_API_URL` | E2E backend helper base URL | no | SRC-006, SRC-037 |

## Known unknowns

- Whether every test uses synthetic-only data: UNKNOWN.
- Repository-wide PostgreSQL transaction/reset and cleanup behavior: UNKNOWN.
- Logged-data redaction enforcement in executable test code: UNKNOWN; only project guidance was observed (SRC-017, SRC-018).
- Production-data access prohibition specific to the test framework: UNKNOWN.
- Backing-service state after an actual run: UNKNOWN because no services or commands were started (SRC-000).

## Manual notes

<!-- BEGIN MANUAL -->
UNKNOWN
<!-- END MANUAL -->
