---
context_schema: "1.0"
profile_slug: "hub"
status: filled
generated_at: "2026-07-26T02:25:00+03:00"
snapshot_commit: d4a44174a0642ec060e17277a61414e1bfe1ca3d
---

# Test Architecture

This file describes observed test structure, not a desired scaffold. Each populated fact cites `SOURCE-MAP.md`; absence of evidence is `UNKNOWN`.

## Test topology

| Level/suite | Runner | Test root | Selection mechanism | Approximate count | Evidence |
|---|---|---|---|---:|---|
| Backend unit + integration | Go `testing` + testify | `backend/testframework` | `go test ... ./testframework/...`; packages grouped by domain | 499 test files | SRC-001, SRC-009, SRC-016, SRC-050 |
| Frontend unit/component | Vitest + React Testing Library | `testframework/stages/unit/frontend-test/domains` | Vite include `../testframework/stages/unit/frontend-test/domains/**/*.test.{ts,tsx}` | 37 files / 99 tests | SRC-002, SRC-003, SRC-010, SRC-022 |
| E2E smoke subset | Playwright `chromium-smoke` | `testframework/stages/e2e/domains` | Three explicit `testMatch` entries plus CI grep | 3 spec files | SRC-004, SRC-011, SRC-006 |
| E2E full | Playwright `chromium-full` | `testframework/stages/e2e/domains` | `**/*.spec.ts` | 20 spec files | SRC-004, SRC-012, SRC-022 |

## Organization

| Dimension | Observed convention | Exceptions | Evidence |
|---|---|---|---|
| Domain/layer grouping | Backend packages and frontend/E2E directories are grouped by product domain; shared support has dedicated directories | Some platform support is nested below `backend/testframework/platform/*` | SRC-009, SRC-010, SRC-012, SRC-007 |
| File naming | Go uses `*_test.go`; frontend uses `*.test.ts(x)`; E2E uses `*.spec.ts` | UNKNOWN | SRC-003, SRC-004, SRC-017 |
| Case naming | Go style is `Test<Function>_<Scenario>`; sampled Vitest and Playwright cases use descriptive strings within `describe`/`test` | UNKNOWN | SRC-017, SRC-027, SRC-033, SRC-042 |
| Unit vs integration boundary | Go unit, handler, SQLite, Redis, MongoDB, and embedded-NATS tests share the backend test root and CI job | MongoDB tests may skip when the service is unavailable; one sampled NATS integration file is excluded by `short` build tag | SRC-016, SRC-025, SRC-028, SRC-030 |

## Framework components

| Component | Implementation | Path | Evidence |
|---|---|---|---|
| Fixtures/factories | Fluent Ent/SQLite environment builder, user factory, and Playwright availability fixtures | `backend/testframework/env/builder.go`; `backend/testframework/factory/user.go`; `testframework/stages/e2e/shared/fixtures.ts` | SRC-023, SRC-024, SRC-036 |
| Shared helpers | Fiber HTTP Act/Assert helpers, React Query renderer, and E2E business steps | `backend/testframework/steps/http.go`; `testframework/stages/unit/frontend-test/render.tsx`; `testframework/stages/e2e/shared/steps/` | SRC-026, SRC-032, SRC-041 |
| Mocks/fakes/stubs | miniredis; Vitest module mocks/spies; embedded NATS server; handler-local Fiber apps | Representative files under auth, frontend tracker/AI, platform NATS, and bases | SRC-028, SRC-034, SRC-035, SRC-030, SRC-029 |
| API clients | E2E helpers use bounded `fetch` calls for health, dev seed, and dev login | `testframework/stages/e2e/shared/helpers/backend.ts`; `.../helpers/auth.ts` | SRC-037, SRC-038 |
| Page/screen objects | Base page plus feature page objects; business steps compose page objects | `testframework/stages/e2e/shared/pages/`; `testframework/stages/e2e/shared/steps/` | SRC-039, SRC-040, SRC-041 |
| Reporting | Playwright list + HTML reporters; trace on first retry; screenshots on failure; CI uploads failure reports/logs | `frontend/playwright.config.ts`; `.github/workflows/ci.yml` | SRC-004, SRC-006 |

## Runtime behavior

| Concern | Observed behavior | Evidence |
|---|---|---|
| Setup/teardown | Go helpers register `t.Cleanup`; Mongo databases are dropped and clients disconnected; Vitest store state is reset in `beforeEach`; Playwright can start a frontend web server | SRC-023, SRC-025, SRC-033, SRC-004 |
| Isolation | SQLite builders use unique in-memory DSNs; user factory emails are UUID-suffixed; Mongo helper creates a unique database; each frontend render gets a new QueryClient | SRC-023, SRC-024, SRC-025, SRC-032 |
| Parallelism | Playwright is fully parallel locally but CI uses one worker; Go guidance allows `t.Parallel` only with an isolated environment; repository-wide Go parallelism is UNKNOWN | SRC-004, SRC-046 |
| Time/async waits | miniredis `FastForward` is used for magic-link TTL; sampled WS test polls with a bounded deadline; Playwright assertions use configured timeouts | SRC-028, SRC-045, SRC-040 |
| Retry/quarantine/skip | Playwright retries twice in CI and not locally; fixtures conditionally skip unavailable backend/dev/frontend; Mongo integration helper skips unavailable MongoDB | SRC-004, SRC-036, SRC-025 |
| Failure artifacts | Playwright HTML report, first-retry trace, failure screenshots, backend E2E log, and Lighthouse reports are configured | SRC-004, SRC-006 |

## Representative source locations

Paths here are pointers, not approved pattern cards.

| Purpose | Source path and lines | Why representative | Evidence |
|---|---|---|---|
| Backend environment builder | `backend/testframework/env/builder.go:18-96` | Shows unique in-memory Ent setup and cleanup | SRC-023 |
| Backend HTTP/RBAC integration | `backend/testframework/bases/rbac_test.go:22-88` | Shows Fiber handler testing over Ent/SQLite with explicit authorization actors | SRC-029 |
| Redis/time control | `backend/testframework/auth/magiclink_test.go:19-31,115-145` | Shows miniredis lifecycle and deterministic TTL advancement | SRC-028 |
| External-service integration | `backend/testframework/mongotest/setup.go:18-56` | Shows Mongo connection, conditional skip, unique DB, and cleanup | SRC-025 |
| Frontend provider renderer | `testframework/stages/unit/frontend-test/render.tsx:7-49` | Shows per-test QueryClient and provider wrapper | SRC-032 |
| Frontend mocks and async assertions | `testframework/stages/unit/frontend-test/domains/tracker/issue-form.test.tsx:1-80` | Shows module doubles, RTL queries, and `waitFor` | SRC-034 |
| E2E fixture policy | `testframework/stages/e2e/shared/fixtures.ts:9-50` | Shows availability fixtures and conditional skips | SRC-036 |
| E2E POM + steps | `testframework/stages/e2e/shared/pages/login.page.ts:8-40`; `.../steps/auth.steps.ts:16-97` | Shows page-object and business-step composition | SRC-040, SRC-041 |

## Risks and known unknowns

- Backend test-file count is 499 for the dedicated `**/*_test.go` filesystem inventory. The previous profile recorded a historical 495-versus-499 documentation disagreement; corrected `docs/test_plan.md` now agrees with the live count (SRC-016, SRC-050).
- Some integration and E2E cases skip when dependencies or dev mode are unavailable; a green run can therefore contain skipped coverage (SRC-025, SRC-036).
- Time handling is mixed: deterministic Redis clock advancement exists, while sampled concurrent tests also use bounded wall-clock polling (SRC-028, SRC-045).
- The source sample covered all declared suite types and shared support, but it was not an exhaustive semantic review of all 499 backend test files; unsampled conventions remain UNKNOWN (SRC-000, SRC-050).
- Test and coverage execution state is NOT_RUN; duration and flakiness remain UNKNOWN because no project command was executed (SRC-000).

## Manual notes

<!-- BEGIN MANUAL -->
UNKNOWN
<!-- END MANUAL -->
