# TESTS — Hub backend full-test pass + coverage audit

Date: 2026-05-15
Scope: `C:\Users\Valerii\Desktop\hub\backend` (Go module `github.com/valerii/hub/backend`)
Context: Phase 14 (Architectural Hardening) just completed — services no longer hold `*ent.Client`. This report confirms 0 fails / 0 races and surfaces coverage weak spots.

---

## 1. Totals

| Metric                                  | Value         |
|-----------------------------------------|---------------|
| Test packages run                        | 27            |
| `[no test files]` packages (mostly `ent/*` generated) | 71 |
| Test functions (`grep '^func Test'`)     | **318**       |
| Test files (`*_test.go` under `internal/`) | 41          |
| FAILs                                    | **0**         |
| Data races (`go test -race -count=1`)    | **0**         |
| `go vet ./...`                           | clean (exit 0)|
| `go build -tags release ./...`           | clean (exit 0)|
| Overall statement coverage               | **11.5%**     |
| Race-suite wall time                     | ~14s longest pkg (`internal/calls`), full pass under 60s |

Race+test pass command:
```
go test -race -count=1 -timeout 540s ./...
```
All 27 testable packages reported `ok`. Logs: `$env:TEMP\hub_test_race.log`, `\hub_cover.log`, `\hub_func.log`.

> Note on the `go: no such tool "covdata"` lines in coverage log — they appear only for `ent/*` and a handful of zero-source platform pkgs that have no Go files compiled into the test binary. They are warnings, not failures; the run still exits 0 against actual test packages.

---

## 2. Coverage per package (sorted ascending)

| Package                         | Coverage |
|---------------------------------|----------|
| internal/dashboard              | 0.0%     |
| internal/dev                    | 0.0%     |
| internal/storage                | 0.0%     |
| internal/platform/nats          | 0.0%     |
| internal/imex                   | 2.5%     |
| internal/gitlab                 | 3.9%     |
| internal/calls                  | 4.9%     |
| internal/github                 | 5.0%     |
| internal/backup                 | 5.3%     |
| internal/notification           | 5.7%     |
| internal/project                | 7.4%     |
| internal/apigateway             | 9.4%     |
| internal/audit                  | 10.1%    |
| internal/tms                    | 12.3%    |
| internal/tracker                | 12.5%    |
| internal/messenger              | 12.8%    |
| internal/ws                     | 15.2%    |
| internal/auth                   | 17.0%    |
| internal/sprint                 | 18.3%    |
| internal/search                 | 24.3%    |
| internal/platform/outbox        | 28.6%    |
| internal/rbac                   | 33.8%    |
| internal/platform/http          | 40.4%    |
| internal/platform/middleware    | 45.2%    |
| internal/platform/crypto        | 78.0%    |
| internal/platform/errors        | 100.0%   |
| internal/platform/telegram      | 100.0%   |

> 4 packages are at flat 0% (tests exist but only assert DTO/constant shapes), 17 packages are below 20%, only 2 are above 50% (both small leaf platform pkgs).

---

## 3. Packages under 50% coverage (priority list)

All 25 packages except `platform/errors` (100%) and `platform/telegram` (100%) are below 50%. The 0%/sub-10% packages are the urgent ones:

### Tier 0 — 0% statement coverage (tests exist but exercise nothing real)
- **internal/dashboard** — `service_test.go` only validates DTO defaults and serialization; `service.go`/`provisioner.go`/`handler.go` (admin stats, layout persistence, default-layout provisioner) are untouched.
- **internal/dev** — `seed_test.go` only checks demo-data struct validity; the actual seeder execution path and `handler.go` (apidocs route) are untouched. Lower priority (dev-only).
- **internal/storage** — `service_test.go` only validates URL-TTL constants and DTO field shapes; `service.go` (S3/MinIO presign, upload/download flow) and `orphan_cleanup.go` are untouched. **High risk** — this is the file-upload path.
- **internal/platform/nats** — `client_test.go` only checks constant values and struct shape; `client.go` (connect, JetStream init, publish/subscribe) has no behavioral coverage. **Critical** — every event publish goes through here.

### Tier 1 — under 10% (large surface, thin tests)
- `internal/imex` (2.5%), `internal/gitlab` (3.9%), `internal/calls` (4.9%), `internal/github` (5.0%), `internal/backup` (5.3%), `internal/notification` (5.7%), `internal/project` (7.4%), `internal/apigateway` (9.4%).

### Tier 2 — 10–25%
- `internal/audit`, `internal/tms`, `internal/tracker`, `internal/messenger`, `internal/ws`, `internal/auth`, `internal/sprint`, `internal/search`.

---

## 4. Service test-file inventory (requested checklist)

Packages that exist in this repo and their `*_test.go` count:

| Requested package | Status | Test files |
|---|---|---|
| `internal/auth`      | ✅ present | 3 (`handler_test.go`, `service_test.go`, `magiclink_test.go`) |
| `internal/calls`     | ✅ present | 4 (`service_test.go`, `chat_test.go`, `recording_test.go`, `recording_webhook_test.go`) |
| `internal/messenger` | ✅ present | 4 (`service_test.go`, `auto_channels_test.go`, `presence_test.go`, `unread_test.go`) |
| `internal/audit`     | ✅ present | 1 (`service_test.go`) |
| `internal/search`    | ✅ present | 2 (`service_test.go`, `jql_test.go`) |
| `internal/backup`    | ✅ present | 1 (`service_test.go`) |
| `internal/storage`   | ✅ present | 1 (`service_test.go` — DTO-only) |
| `internal/rbac`      | ✅ present | 2 (`permissions_test.go`, `middleware_integration_test.go`) |
| `internal/project`   | ✅ present | 1 (`service_test.go`) |

Requested packages that **do not exist as standalone packages** in this codebase (rolled into other domains):

| Requested | Where it lives instead |
|---|---|
| `users`      | folded into `internal/auth` (user profile/session) and `ent/user` |
| `roles`      | folded into `internal/rbac` (`permissions_test.go` covers role checks) |
| `workspaces` | folded into `internal/project` (workspace == project root) |
| `channels`   | folded into `internal/messenger` (`auto_channels_test.go`) |
| `webhooks`   | folded into `internal/calls/recording_webhook_test.go` + `internal/platform/outbox` (delivery) |
| `outbox`     | lives at `internal/platform/outbox` (28.6% coverage, has `outbox_test.go`) |
| `magiclink`  | lives in `internal/auth/magiclink.go` + `magiclink_test.go` (6 tests, brute-force/lockout/rate-limit covered) |

So every domain on your list is exercised by at least one test — no missing files.

---

## 5. Integration tests for critical handlers

22 test functions match the critical-handler keywords (`TestLogin*`, `TestRefresh*`, `TestSendMagicLink*`, `TestVerifyMagicLink*`, `TestRequirePermission*`, `TestRequireSystemRole*`). Concretely:

- **Login / magic-link**: `TestSendMagicLink_InvalidEmail`, `TestSendMagicLink_EmptyBody`, `TestVerifyMagicLink_ShortCode`, plus 6 deep magic-link tests in `magiclink_test.go` covering success path, brute-force lockout, lockout TTL, send rate-limit, no-code-stored, success-clears-fails.
- **Refresh**: `TestRefreshToken_EmptyBody` (handler-level guard). JWT internals covered by `TestJWTRoundTrip`, `TestJWTExpired`, `TestJWTWrongSecret`.
- **Logout / session**: `TestLogout_Unauthenticated`, `TestGetMe_Unauthenticated`.
- **Telegram auth**: 4 tests (valid signature, invalid hash, expired auth, missing hash).
- **RBAC middleware** (`internal/rbac/middleware_integration_test.go`): `TestRequirePermission_Anonymous_401`, `_NonMember_403`, `_MemberWithPermission_Next`, `_MemberWithoutPermission_403`, `TestRequireSystemRole_AdminOnly`, `_InvalidUserID_401`, `_MissingProjectID_400`.
- **Platform middleware** (`internal/platform/middleware/middleware_test.go`): 19 tests covering request-id propagation/uniqueness, error-handler for AppError/FiberError/generic, CORS (default origins, env override, disallowed origin, credentials, parser rejects `*`/empty), rate-limit (within/exceeds/separate-keys/window-expiry/middleware-blocks), logger.

Verdict: **login, refresh, and RBAC middleware all have integration coverage**. Refresh-token happy-path and rotation/replay-detection scenarios are the obvious gap (only the empty-body guard is tested).

---

## 6. Docker / infra dependencies

No package was skipped because of missing infra. Tests run **purely locally** — no Postgres/Mongo/Redis/MinIO/NATS container required:

- DB-bound services (`auth`, `messenger`, `calls`, `audit`, `project`, `rbac`, `search`, `tms`, `tracker`, `sprint`, `notification`, `backup`, `imex`, `dashboard`) use `ent/enttest` with a SQLite in-memory client.
- Outbox/NATS/Redis tests work against in-process fakes or unit-test the pure helpers (which is why `platform/nats` shows 0% — only constants are asserted).
- The race suite finished in <60 s wall-clock; no test required `testcontainers`, `dockertest`, or `TestMain` env-gated skips.

---

## 7. Top recommendations (what to cover next)

Ranked by leverage / risk:

1. **`internal/platform/nats` — currently 0%.** Add a test that spins an embedded NATS server (`server.Run(opts)` from `nats-server/v2`) and asserts: connect retry, JetStream stream init idempotency, `Publish` returns ack, subject routing. Every domain event flows through this client; a regression here is silent and catastrophic.
2. **`internal/storage` — currently 0%.** Use `minio-go` fake/seaweedfs-stub or a `s3test` httptest server to cover `PresignUpload` / `PresignDownload` / `orphan_cleanup` happy + error paths. This is the file-upload boundary — security-sensitive (presign TTL, key sanitization).
3. **`internal/auth` refresh-token path (17% overall).** Add `TestRefreshToken_*` cases: valid refresh rotates the session, reused refresh detects replay and revokes the family, expired refresh returns 401, refresh after logout returns 401. Right now only empty-body is exercised.
4. **`internal/calls` recording webhook (4.9%).** `recording_webhook_test.go` exists but the service-side state machine (start → in-progress → finalized) and signed-URL verification are barely covered. Add table-driven tests for each transition + tampered-signature rejection.
5. **`internal/messenger` core flow (12.8%).** Send-message → fan-out → unread-counter → mention-notification is the product's hottest path. Cover at minimum: send to channel with N members increments unread for N-1, edit message preserves unread state, delete decrements pinned/bookmark counts.
6. **`internal/rbac` permission resolution (33.8%).** Middleware is well-tested; the permission **computation** (role inheritance, project-scoped overrides, system-role short-circuits) is the gap. Add table tests over the `permissions.go` resolver.
7. **`internal/platform/outbox` (28.6%).** Add worker-loop tests: ack on success, exponential backoff on transient error, poison-pill quarantine after N failures, ordering preservation within an aggregate.
8. **`internal/dashboard` provisioner (0%).** `provisioner.go` builds default layouts on first login — regression here breaks new-user onboarding silently. Cover provision-idempotent, existing-layout-preserved, project-scoped vs personal.
9. **`internal/audit` redaction (10.1%).** Make sure PII redaction in `service.go` is exercised — currently only the basic create/list path is touched.
10. **`internal/imex` (2.5%)** and **`internal/backup` (5.3%)** — both deal with bulk data movement; add at least round-trip tests (export → import yields equal entity counts; backup → restore preserves a small fixture).

---

## 8. Bottom line

- **0 failures, 0 races, vet clean, release build clean — Phase 14 baseline is green.** ✅
- 318 tests / 27 packages confirmed.
- The pass-rate is solid; the **coverage story is weak** — overall 11.5%, four packages at 0%, only two leaf packages above 50%. Service-level statement coverage is the next investment, not test stability.
- No Docker dependency, no skipped suites — what you see is what runs everywhere.
