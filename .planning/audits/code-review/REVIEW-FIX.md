# Hub Backend — Code Review Fix Report

**Fixed at:** 2026-05-14
**Source review:** `.planning/audits/code-review/REVIEW.md`
**Branch:** master
**Scope:** CRITICAL findings (CR-01 through CR-06)

## Summary

| Finding | Status | Commit | Note |
|---------|--------|--------|------|
| CR-01   | FIXED  | `96ae5fb` | Pre-applied (working-tree), committed as part of this run. |
| CR-02   | FIXED  | `e58cca9` | Presign + RBAC. |
| CR-03   | FIXED  | `6d9300d` | Webhook SSRF + URL validation; preview.go partial. |
| CR-04   | FIXED  | `988c93b` | /files/* RBAC rewired to use the file's actual project. |
| CR-05   | FIXED  | `37aff96` | Fail-start on dev defaults; config.Load now returns error. |
| CR-06   | FIXED  | `95f173e` | Loud DEV_MODE warning; hard fail enforced by CR-05 commit. |

Verification (per commit):
- `go build ./...` — clean after every commit.
- `go test ./internal/{affected}/...` — green for all touched packages
  (calls, storage, apigateway, platform/http, rbac, github, dev). No
  package had its test count regress.

---

## CR-01 — Admin guard for `/admin/backups/*`

**Status:** FIXED
**Commit:** `96ae5fb`
**Files touched:**
- `backend/internal/backup/handler.go` — `RegisterRoutes(api, authMw, adminMw fiber.Handler)` signature; wires `adminMw` into the group middleware chain.
- `backend/cmd/server/main.go` — call site passes `rbac.RequireAdmin(db)`.

**Note:** This fix already existed in the working tree at session start. It
was committed in this run so the rest of the fixes could be applied on a
clean tree.

---

## CR-02 — Recording download: presigned URL + RBAC

**Status:** FIXED
**Commit:** `e58cca9`
**Files touched:**
- `backend/internal/calls/recording.go`
  - Added AWS SDK v2 imports (`aws`, `awsconfig`, `credentials`, `s3`).
  - `RecordingService` gains `presignClient *s3.PresignClient`.
  - `NewRecordingService` builds the presign client from `S3Config` (same
    pattern as `storage.NewStorageService`). Endpoint is prefixed with
    `http(s)://` if the raw value omitted the scheme.
  - New `RecordingProjectID(ctx, recordingID) (uuid.UUID, error)` helper
    so the handler can authorize before any S3 work.
  - `GetRecordingDownloadURL` now calls `presignClient.PresignGetObject`
    with a 1-hour expiry; falls back to a clear error if the presign
    client failed to initialize.
- `backend/internal/calls/handler.go`
  - `Handler` gains `rbacSvc *rbac.RBACService`.
  - `NewHandler` accepts `rbacSvc` (last positional argument). Documented
    as optional (nil-safe for tests).
- `backend/internal/calls/handler_recording.go`
  - `ListProjectRecordings` requires `PermProjectView` via
    `rbacSvc.GetEffectivePermissions`; returns 403 when not a project
    member.
  - `GetRecordingDownloadURL` loads the recording's project via
    `RecordingProjectID`, then checks `PermProjectView`.
- `backend/cmd/server/main.go`
  - Updated `callsmod.NewHandler(callsSvc, recordingSvc, scheduledCallSvc, cfg.LiveKit, rbacSvc)`.

**Tests:** `go test ./internal/calls/...` — passes (10.9s; uses
testcontainers Mongo).

---

## CR-03 — SSRF protection for webhook dispatcher

**Status:** FIXED (preview.go partial — see note)
**Commit:** `6d9300d`
**Files touched:**
- `backend/internal/platform/http/safedialer.go` (new) — reusable helpers:
  - `SafeDialContext()` resolves the host and rejects RFC1918, 127.0.0.0/8,
    169.254.0.0/16, ::1, fe80::/10, plus CGNAT, multicast and TEST-NETs.
    The resolution happens inside the dialer, so DNS-rebinding is covered.
  - `ValidateOutboundURL(url)` rejects non-http(s) schemes, userinfo URLs,
    and empty hosts.
  - `NewSafeHTTPClient(timeout)` returns an `http.Client` with the safe
    transport and reasonable timeouts.
- `backend/internal/apigateway/webhook.go`
  - Shared package-level `webhookHTTPClient = safehttp.NewSafeHTTPClient(10s)`.
  - `sendWebhook` validates the URL up front and dispatches via the safe
    client.
  - `CreateEndpoint` and `UpdateEndpoint` reject obviously bad URLs at
    persist time (defense-in-depth, returns 400).
- `backend/internal/github/preview.go`
  - `resolveExternalPreview` calls `ValidateOutboundURL` before
    `opengraph.Fetch`.

**Note (preview.go):** The `otiai10/opengraph/v2` package uses its own
internal `http.Client`. Wiring `SafeDialContext` into it would require
either a fork or an upstream change, both out of scope for this fix. The
URL-scheme / userinfo / empty-host check still rejects the easiest SSRF
strings, and the comment in the file documents the residual gap.

**Tests:** apigateway, github, platform/http packages — all green.

---

## CR-04 — `/files/*` RBAC

**Status:** FIXED
**Commit:** `988c93b`
**Files touched:**
- `backend/internal/storage/service.go`
  - New `FileProjectID(ctx, fileID) (uuid.UUID, error)` that queries
    `FileMetadata.Query().Where(filemetadata.IDEQ(fileID), filemetadata.DeletedAtIsNil()).QueryProject().OnlyID(ctx)`.
    Verified the edge API by reading `ent/filemetadata.go` (has
    `QueryProject`) and `ent/filemetadata_query.go` (M2O edge to
    Project). Soft-deleted files (`deleted_at != nil`) return not-found.
- `backend/internal/storage/handler.go` — rewritten:
  - Removed all `rbac.RequirePermission` middlewares from routes (they
    required `:projectId` in the path that the routes never had — so the
    module was effectively returning 400 on every call).
  - `Handler` now stores `rbacSvc *rbac.RBACService` set in `RegisterRoutes`.
  - Helper `authorizeProject(c, userID, projectID, perm)` centralises the
    `GetEffectivePermissions` + `Has(perm)` check with proper 401/403/500
    mapping (uses `ent.IsNotFound` to return 403 instead of 500 when the
    caller is simply not a project member).
  - `GenerateUploadURL` parses `req.ProjectID` from the body and checks
    `PermFileUpload` against it.
  - `ConfirmUpload`, `GetDownloadURL`, `DeleteFile` look up the file's
    actual `projectID` via `FileProjectID` (closes the alice-projectId /
    bob-fileId IDOR), then check `PermFileUpload`, `PermProjectView`,
    `PermFileDelete` respectively.
  - Bonus: missing/soft-deleted files now return 404 (instead of 500
    after the presign call).

**Tests:** storage, rbac — green.

---

## CR-05 — Fail-start on dev secret defaults

**Status:** FIXED
**Commit:** `37aff96`
**Files touched:**
- `backend/internal/platform/config/config.go`
  - Signature change: `func Load() (*Config, error)`.
  - New `validate()` method on `*Config`:
    - In `DevMode=false`, refuses to boot when any of `JWT_SECRET`,
      `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `LIVEKIT_API_KEY`,
      `LIVEKIT_API_SECRET` matches its known dev default (case-insensitive
      string equality on the values from the original review).
    - `JWT_SECRET` is additionally rejected if it's empty, shorter than
      32 bytes, or contains `change-in-production` (covers stock
      docker-compose values).
    - All problems are collected and reported in a single error to
      avoid the "fix one secret, restart, get told about the next one"
      cycle.
  - New `Environment string` field on `ServerConfig`, read from
    `ENVIRONMENT` (default `development`).
  - `DEV_MODE=true` AND `ENVIRONMENT=production` is rejected outright.
- `backend/cmd/server/main.go`
  - `cfg, err := config.Load()` with `slog.Error` + `os.Exit(1)` on
    error, matching the existing fatal-error pattern in this file.
  - `db, err := ...` immediately after compiles because `db` is new
    (`:=` allows at least one new lvalue).

**Tests:** No existing config tests; build clean.

---

## CR-06 — Loud DEV_MODE warning

**Status:** FIXED
**Commit:** `95f173e`
**Files touched:**
- `backend/cmd/server/main.go`
  - After successful config load, when `cfg.Server.DevMode` is true,
    emit `slog.Warn("DEV_MODE enabled — /api/dev/* endpoints exposed, auth rate-limit DISABLED — DO NOT USE IN PRODUCTION", "environment", cfg.Server.Environment)`.
  - The hard guarantee (refuse to boot when
    `ENVIRONMENT=production AND DEV_MODE=true`) lives in `config.validate`
    and shipped in the CR-05 commit; this commit adds the visible warning
    for the non-production case where DEV_MODE is intentional.

**Tests:** dev package — green.

---

## Skipped / Not in scope

Nothing was skipped from the requested CR-02..CR-06 set. CR-01 and
`chat_test.go` were already-applied per the task brief and were
committed at the start of this session so the rest of the work happened
on a clean tree.

---

# Wave 2 — HIGH fixes

**Fixed at:** 2026-05-14
**Scope:** HI-01, HI-02, HI-03, HI-09 (per task brief)
**Branch:** master
**Build:** `go build ./...` clean after every commit.
**Tests:** every touched package re-run with `go test -count=1`, all green.

## Summary

| Finding | Status | Commit  | Note |
|---------|--------|---------|------|
| HI-01   | FIXED  | `f26b589` | Messenger read endpoints now require channel membership. |
| HI-02   | FIXED  | `22dfa60` | ForwardMessage validates source-channel membership. |
| HI-03   | FIXED  | `c25c010` | JoinCall + EndCall gated by channel membership/role. |
| HI-09   | FIXED  | `093b1cf` | Outbox per-event recover + Save (not SaveX) + Fiber recover middleware. |

---

## HI-01 — Channel membership required on read endpoints

**Status:** FIXED
**Commit:** `f26b589`
**Files touched:**
- `backend/internal/messenger/service_channels.go` — `GetChannel` and
  `ListMembers` now take `userID uuid.UUID` and call
  `s.repo.IsMember(ctx, channelID, userID)` first, returning
  `apperrors.NewForbidden(...)` on non-membership.
- `backend/internal/messenger/service_messages.go` — same change for
  `ListMessages`. `ListThreadReplies` adds a `FindMessageByID` lookup
  first so the membership check is anchored on the parent message's
  channel (the route lives under `/channels/:channelId/messages/...` but
  the service operates by `messageID`, so the channel must be resolved
  from the document, not assumed).
- `backend/internal/messenger/service_pins.go` — `ListPins` gains the
  same check.
- `backend/internal/messenger/handler_channels.go`,
  `handler_messages.go`, `handler_reactions.go` — call sites use
  `httputil.UserIDFromCtx(c)` and pass `userID` through.

**Notes:**
- Public channels are also gated by membership at this layer. Project
  RBAC continues to govern who can be added as a member; nothing
  changes about legitimate access.
- `repo.IsMember` is already used by `SendMessage`, so behaviour and
  performance characteristics match the write-side check.

**Tests:** `go test ./internal/messenger/...` — green (0.96s).

---

## HI-02 — Source-channel membership in ForwardMessage

**Status:** FIXED
**Commit:** `22dfa60`
**Files touched:**
- `backend/internal/messenger/service_messages.go` — after loading
  `source` via `FindMessageByID`, runs
  `s.repo.IsMember(ctx, source.ChannelID, userID)` and returns 403
  before copying `ContentJSON`/`ContentText`/`Attachments` into the new
  message.

**Notes:**
- Target-channel membership check already existed; only the source
  side was missing, which let HI-01 be bypassed by forwarding.
- `TestForwardMessage_CreatesNewMessage` already sets up the caller as
  a member of the source channel, so the existing test continues to
  pass without modification.

**Tests:** messenger green.

---

## HI-03 — Authorization on JoinCall and EndCall

**Status:** FIXED
**Commit:** `c25c010`
**Files touched:**
- `backend/internal/calls/repository.go` — adds two helpers on
  `EntRepository`:
  - `IsChannelMember(ctx, channelID, userID) (bool, error)` — queries
    the same `channel_members` table the messenger uses (`channelmember`
    Ent client is already imported for `FindConflictingCalls`, no new
    cross-package wiring needed).
  - `GetChannelMemberRole(ctx, channelID, userID) (string, error)` —
    returns `""` (not error) when the user is not a member, so the
    caller doesn't have to special-case `ent.IsNotFound`.
- `backend/internal/calls/service.go`:
  - New sentinel `ErrCallForbidden = errors.New("forbidden")` so the
    handler can map authorization failures to 403 (other errors stay
    500).
  - `JoinCall` — for channel calls, requires `IsChannelMember`. For
    instant calls (`ChannelID == nil`), allows only `CreatedBy ==
    userID`; invitees/guests use `JoinByInviteCode` and `GuestJoinCall`
    which have their own paths and don't go through `JoinCall`.
  - `EndCall` — accepts `userID`; allows when `CreatedBy == userID`,
    otherwise loads channel role and accepts `owner`/`admin`. Instant
    calls without a channel: creator only.
- `backend/internal/calls/handler_calls.go`:
  - `EndCall` now reads `userID` via `httputil.UserIDFromCtx` and
    passes it to `service.EndCall`.
  - Both `JoinCall` and `EndCall` translate `errors.Is(err,
    ErrCallForbidden)` to `fiber.StatusForbidden`; everything else
    stays 500.
  - `errors` added to imports.

**Instant-call regression check:**
- `CreateInstantCall` calls `repo.AddParticipant` directly and returns a
  token — it doesn't go through `JoinCall`, so creators still receive a
  working response on their own instant call.
- `JoinByInviteCode` calls `repo.AddParticipant` + `GenerateToken`
  directly without invoking `JoinCall`, so the invite flow is untouched.
- `GuestJoinCall` constructs a guest token without calling `JoinCall`.
- Only the explicit `/api/calls/:callId/join` endpoint goes through
  `JoinCall`, and the new check is exactly the behaviour the issue
  asked for (members of the channel for channel calls, creator for
  instant calls).

**Tests:** `go test ./internal/calls/...` — green (11.1s, uses
testcontainers Mongo).

---

## HI-09 — Outbox panic isolation + global Fiber recover

**Status:** FIXED
**Commit:** `093b1cf`
**Files touched:**
- `backend/internal/platform/outbox/worker.go`:
  - `processBatch` splits per-event work into a new `processEvent`
    method. Each call defers a `recover()` that logs panic + stack with
    `event_id`/`event_type` context and returns — one bad event no
    longer kills the loop. Recover is at the event-iteration scope,
    not the outer `ProcessLoop`, exactly as the brief requested.
  - Both Ent status updates (`SetAttempts(...).SetLastError(...)` on
    publish failure, `SetStatus(Published).SetPublishedAt(...)` on
    success) switched from `SaveX(ctx)` to `Save(ctx)` with explicit
    error logging. We accept at-least-once delivery: if the
    "mark-published" update fails, the event stays pending and may
    publish again next tick — preferable to crashing the worker.
- `backend/cmd/server/main.go`:
  - Imports `fiberrecover
    "github.com/gofiber/fiber/v3/middleware/recover"` — verified path
    exists at `gofiber/fiber/v3@v3.1.0/middleware/recover/recover.go`
    (the version pinned in `go.mod`).
  - `app.Use(fiberrecover.New(fiberrecover.Config{EnableStackTrace:
    true}))` registered immediately after `middleware.RequestID()` so a
    panic still carries the request_id in logs, but before the rest of
    the middleware (Logger, CORS, rate-limit) — handlers and any
    middleware-spawned goroutines that propagate panics back through
    `c.Next()` are protected.

**Tests:** outbox green, full build clean.

---

## Skipped / Not in scope

Nothing skipped from the requested HI-01/HI-02/HI-03/HI-09 set. The
remaining HIGH findings (HI-04 brute-force, HI-05 plaintext tokens,
HI-06 race, HI-07 WS hub, HI-08 webhook fan-out, HI-10
rate-limit-by-IP, HI-11 CORS, HI-12 user search scoping) are outside
this task brief.

---

_Wave 2 fixed: 2026-05-14_
_Fixer: Claude (Opus 4.7 1M context)_

---

# Wave 3 — remaining HIGH

**Fixed at:** 2026-05-14
**Scope:** HI-04, HI-05, HI-06, HI-07, HI-08, HI-10, HI-11, HI-12
**Branch:** master
**Build:** `go build ./...` clean after every commit.
**Tests:** every touched package re-run with `go test -count=1`, all green.

## Summary

| Finding | Status  | Commit    | Note |
|---------|---------|-----------|------|
| HI-04   | FIXED   | `39bd63e` | Per-email/IP fails-counter, 429 mapping. |
| HI-05   | PARTIAL | `1f7a524` | AES-256-GCM at-rest encryption + lazy migration; no backfill. |
| HI-06   | FIXED   | `b8b909d` | Atomic counter/invite increments via Ent Tx + AddX + Where. |
| HI-07   | PARTIAL | `71a62b6` | Multi-tab fixed (keyed by ConnectionID); Subscribe/Unregister race minimised, not fully serialised. |
| HI-08   | FIXED   | `defb815` | 16-worker bounded pool + buffered queue; redundant Marshal dropped. |
| HI-10   | FIXED   | `e85b80e` | Redis sliding-window for /auth and /auth/magic-link/verify; global APIRateLimit unchanged (noted). |
| HI-11   | FIXED   | `167a7dc` | CORS origin parse + trim + reject wildcard; prod fail-start when empty. |
| HI-12   | PARTIAL | `27d04d1` | User search scoped to shared-project members; admin-override deferred. |

---

## HI-04 — Per-email / per-IP failed-attempt limit on code verification

**Status:** FIXED
**Commit:** `39bd63e`
**Files touched:**
- `backend/internal/auth/magiclink.go`
  - New constants `magicLinkVerifyMax = 5`, `magicLinkFailsPrefix = "magiclink:fails:"`.
  - `VerifyCode` checks the counter before any code lookup; on
    failure (`redis.Nil` or mismatched stored code) calls
    `recordFailure` which `INCR`s and, on first failure, `EXPIRE`s
    the key to roughly the code TTL (5m). Success path `DEL`s both
    the code key and the fails counter.
  - New sentinel `ErrMagicLinkVerifyLimited`.
- `backend/internal/auth/telegram_bot.go`
  - Same pattern with `tgFailsPrefix` keyed by `clientKey` (the IP
    threaded in from the handler). `clientKey == ""` disables the
    check (tests).
  - `VerifyCode` signature now `(ctx, code, clientKey)`; new sentinel
    `ErrTelegramVerifyLimited`.
- `backend/internal/auth/handler.go`
  - `LoginTelegramCode` passes `c.IP()` to `VerifyCode` and maps the
    new sentinel to 429.
  - `VerifyMagicLink` maps `ErrMagicLinkVerifyLimited` to 429.

**Tests:** `go test ./internal/auth/...` — green.

---

## HI-05 — Encrypt GitHub/GitLab tokens & webhook secrets at rest

**Status:** PARTIAL
**Commit:** `1f7a524`
**Files touched:**
- `backend/internal/platform/crypto/aesgcm.go` (new)
  - `Cipher` over `cipher.AEAD` (AES-256-GCM).
  - Constructors: `NewCipherFromKey([32]byte)`,
    `NewCipherFromString(hex|base64)`. `DeriveDevKey()` returns a
    deterministic SHA-256 hash of a fixed string for DevMode-only use.
  - `Encrypt` / `Decrypt`: base64( nonce(12) || ciphertext || tag(16) ).
  - `MaybeDecryptString`: tries `Decrypt`; on failure returns the
    input as plaintext with `ok=false` — used for the migration
    window where existing rows are unencrypted.
- `backend/internal/platform/crypto/aesgcm_test.go` (new)
  - Roundtrip, nonce-uniqueness, tampered-ciphertext rejection,
    constructor-format coverage, legacy-plaintext passthrough.
- `backend/internal/platform/config/config.go`
  - New `IntegrationConfig.EncryptionKey` from
    `INTEGRATION_ENCRYPTION_KEY`.
  - `validate()` appends `INTEGRATION_ENCRYPTION_KEY не задан` to
    the production fail-start error list when the env var is empty.
- `backend/internal/github/repository.go`,
  `backend/internal/gitlab/repository.go`
  - `EntRepository` gains a `cipher *cryptopkg.Cipher` field.
  - `NewEntRepository(client, cipher)` — nil cipher = plaintext
    fallback for tests.
  - `CreateIntegration` encrypts `AccessToken` and `WebhookSecret`
    on write, then decrypts the returned row so callers see plaintext.
  - `GetIntegration`, `ListIntegrations`, `FindIntegrationByRepo` /
    `FindIntegrationByProjectPath`: decrypt on read using
    `MaybeDecryptString`, so existing plaintext rows pass through
    untouched. Re-encryption happens implicitly on the next write
    (Update is not currently used; documented as lazy migration).
  - `ListAllWebhookSecrets` decrypts at the point of use for HMAC
    compare and never persists plaintext.
- `backend/cmd/server/main.go`
  - Builds the global `integrationCipher` after config load. If
    `cfg.Integration.EncryptionKey != ""` we call
    `NewCipherFromString`; otherwise (DevMode only — production
    already failed in `validate`) we log a loud warning and use
    `DeriveDevKey`.
  - Threads the cipher into `githubmod.NewEntRepository` and
    `gitlabmod.NewEntRepository`.

**Partial — what is NOT done:**
1. No full backfill migration. Existing rows stay plaintext until
   they happen to be re-saved. `MaybeDecryptString` covers reads;
   re-encrypt on next save is lazy. A one-shot backfill job is
   feasible but considered too risky for a single commit.
2. Ent `Hook` integration was not used; encryption happens at the
   repository layer because that's where the codebase keeps its
   serialization boundary today.

**Tests:** crypto, github, gitlab, config — all green.

---

## HI-06 — Atomic counter increments (imex import, invite join)

**Status:** FIXED
**Commit:** `b8b909d`
**Files touched:**
- `backend/internal/imex/service.go`
  - New `allocateIssueKey(ctx, projectID) (nextNum, issueKey, statusID, err)`
    wraps `s.db.Tx(ctx)` + `tx.Project.UpdateOneID(projectID).AddTaskCounter(1).Save(ctx)`
    + workflow-status lookup, then `tx.Commit()`. Same pattern as
    `tracker.CreateIssue` (verified at `repo_issues.go:24-37`).
  - `ImportIssuesCSV` and `ImportIssuesJSON` both call the new
    helper instead of the old read-modify-write block. Both call
    sites now share the same code path, removing the duplicated
    `proj.TaskCounter + 1` race.
- `backend/internal/project/repository.go`
  - New `TryConsumeInvite(ctx, inviteID, maxUses) (ok bool, err error)`:
    `UpdateOneID(inviteID).AddUseCount(1)` plus `Where(invite.UseCountLT(maxUses))`
    when `maxUses > 0`. `ent.IsNotFound` → returns `ok=false` (slot
    exhausted) rather than error.
  - Method added to the `Repository` interface so mocks must
    implement it.
- `backend/internal/project/invite.go`
  - `JoinByInvite` keeps the cheap fast-path check, then calls
    `TryConsumeInvite` *before* `AddMember`. If reservation fails
    we return 400 immediately; if `AddMember` fails after consuming
    the slot we log and return 500 — at most one "wasted" use_count
    increment, acceptable tradeoff vs. exceeding `MaxUses`.
- `backend/internal/project/service_test.go`
  - `mockRepo.TryConsumeInvite` implementation.

**Tests:** imex, project — green.

---

## HI-07 — Multi-tab WS sessions + race minimisation

**Status:** PARTIAL
**Commit:** `71a62b6`
**Files touched:**
- `backend/internal/ws/hub.go` — rewritten:
  - `Client` gains `ConnectionID` (uuid v4, assigned in `NewClient`).
  - `Hub` keeps two maps:
    - `clients sync.Map` — connectionID → *Client (cleanup + by-id
      lookup);
    - `userClients sync.Map` — userID → `*sync.Map[connectionID → *Client]`
      (fan-out for `SendToUser`).
  - `rooms` inner map now keyed by `ConnectionID` — two tabs of the
    same user are independent room members.
  - `registerClient` inserts into both top-level maps. `removeClient`
    cleans both, scrubs every room the client was in, and prunes the
    `userClients[userID]` entry when it goes empty.
  - `SendToUser` iterates `userClients[userID]` and pushes to every
    open tab; previously only the most-recently-registered tab
    received messages.
- `backend/internal/ws/protocol_test.go`
  - New assertions: `ConnectionID` is non-empty; two clients of the
    same user receive distinct IDs; `SendToUser` fans out to all
    tabs; removing one connection leaves the sibling functional.

**Partial — what is NOT done:**
- Subscribe/Unsubscribe still mutate `h.rooms` directly rather
  than through the register/unregister channel. With per-connection
  keying the worst residual outcome is a stale room entry pointing
  at a closed connection — the closed-channel send selects
  `default` and the stale entry is cleaned the next time
  `removeClient` runs for that connection. Funnelling subscribe
  through the channel would require touching every send path and
  was explicitly out-of-scope per the task brief.

**Tests:** ws — green under `-race`.

---

## HI-08 — Bounded worker pool for webhook delivery

**Status:** FIXED
**Commit:** `defb815`
**Files touched:**
- `backend/internal/apigateway/webhook.go`
  - New constants `webhookWorkers = 16`, `webhookQueueSize = 256`.
  - New `webhookJob` struct carrying `ep, eventType, body, payload`.
  - `WebhookDispatcher` gains `jobs chan webhookJob`,
    `workerWG sync.WaitGroup`, `stopOnce sync.Once`.
  - `Start` spawns 16 `deliveryWorker` goroutines before mounting
    the JetStream consumer. Logs include `workers` and `queue`
    counts at startup.
  - `Stop` (idempotent via `sync.Once`) stops the consumer, closes
    the channel, and waits for the workers to drain.
  - `handleMessage` now passes `msg.Data()` through as `body` and
    unmarshals once into `payload` for the `WebhookDelivery`
    record. The previous per-delivery `json.Marshal(payload)` is
    gone.
  - `dispatchEvent` enqueues via a non-blocking `select` — if the
    queue is full we drop with a warn-log and rely on JetStream
    redelivery (`MaxDeliver=3`).
  - `deliverWithRetry` now takes a `webhookJob`; `signPayload` uses
    `job.body`; the DB row uses `job.payload`.

**Tests:** apigateway — green.

---

## HI-10 — Redis sliding-window rate limiter for auth endpoints

**Status:** FIXED
**Commit:** `e85b80e`
**Files touched:**
- `backend/internal/platform/middleware/ratelimit.go`
  - New `RedisRateLimitConfig{Max, Window, KeyPrefix}` and
    `RateLimitKeyExtractor func(c fiber.Ctx) string`.
  - New `RateLimitRedis(rdb, cfg, extractKey)` middleware using a
    Redis pipeline: `ZRemRangeByScore` (drop old) + `ZAdd` (current)
    + `ZCard` (count) + `Expire` (TTL = window + 1m). Fail-OPEN on
    Redis errors with a warn log — the alternative is a DoS-self
    when Redis is briefly unavailable.
  - `ExtractEmailFromBody(field)` parses JSON body and returns
    `"email:" + normalizedEmail`. Used for magic-link verify so
    the bucket is per-email rather than per-IP.
  - `hashIP(ip)` returns the first 16 hex of `sha256(ip)` for a
    stable, slightly PII-reduced fingerprint when no extractor or
    extractor returns empty.
- `backend/cmd/server/main.go`
  - In production (`!cfg.Server.DevMode`):
    - `api.Use("/auth/magic-link/verify", ...)` Redis limiter
      keyed by email-from-body, 10 / 10 min (registered FIRST so
      it runs before the broader limiter on the same path);
    - `api.Use("/auth", ...)` Redis limiter keyed by hashed IP,
      reusing `AuthRateLimit.Max=3, Window=10m`.
  - Global `app.Use(middleware.RateLimit(APIRateLimit))` left
    in-memory (followup: still single-instance only, but the
    budget is acceptable per-instance).

**Tests:** middleware — green.

---

## HI-11 — CORS origin parsing + production fail-start

**Status:** FIXED
**Commit:** `167a7dc`
**Files touched:**
- `backend/internal/platform/middleware/cors.go`
  - New pure helper `parseCORSOrigins(env) (origins, problems)`:
    splits on `,`, trims each entry, drops empty strings AND
    literal `"*"`. Each rejection is collected as a human-readable
    reason and emitted via `slog.Warn` from the caller.
  - `CORS()` now logs all parse problems, falls back to localhost
    defaults only when the env var is empty. If after filtering
    there are zero valid origins AND DEV_MODE is not truthy, logs
    `slog.Error` + `os.Exit(1)` — same fail-start pattern as
    `config.validate`.
  - `isDevMode()` reads `DEV_MODE` and returns true for
    `true/1/yes/on` (case-insensitive).
- `backend/internal/platform/middleware/middleware_test.go`
  - Existing CORS_ tests set `DEV_MODE=true` via `t.Setenv` so
    they don't trigger the prod fail-start.
  - New `TestParseCORSOrigins_TrimsAndRejectsWildcardAndEmpty`
    covers the pure parsing helper across five cases.

**Tests:** middleware — green.

---

## HI-12 — Scope user search to shared-project members

**Status:** PARTIAL
**Commit:** `27d04d1`
**Files touched:**
- `backend/internal/search/service.go`
  - `searchUsers(ctx, query)` is now `searchUsers(ctx, query, callerID)`.
  - SQL adds:
    ```
    AND (
      id = $2
      OR id IN (
        SELECT DISTINCT pm2.user_project_members
        FROM project_members pm1
        JOIN project_members pm2
          ON pm1.project_members = pm2.project_members
        WHERE pm1.user_project_members = $2
      )
    )
    ```
    so the result set is restricted to (caller) ∪ (users sharing
    at least one project membership with caller).
  - `Search()` already had `userID` — passes it directly into
    `searchUsers`; no handler change.

**Partial — what is NOT done:**
- No system-admin override path. The codebase doesn't currently
  thread a global-admin role through `Search()`; admins still get
  their own scoped list. Filed as follow-up — adding this would
  require touching `rbac.IsSystemAdmin` (or equivalent) and
  plumbing it through `SearchService`.

**Tests:** search — green.

---

_Wave 3 fixed: 2026-05-14_
_Fixer: Claude (Opus 4.7 1M context)_

---

# Wave 4 — MEDIUM

**Fixed at:** 2026-05-14
**Source review:** `.planning/audits/code-review/REVIEW.md`
**Branch:** master
**Scope:** selected MEDIUM findings (ME-01, ME-03, ME-05, ME-06, ME-07,
ME-11, ME-13, ME-15). Other MEDIUM findings deferred — none of the
selected fixes were skipped.

## Summary

| Finding | Status | Commit | Note |
|---------|--------|--------|------|
| ME-01   | FIXED  | `5afbb71` | Bulk `Session.Delete()` + Redis `Pipeline.Del`; reads tokens via `Select(session.FieldRefreshToken).Strings`. |
| ME-03   | FIXED  | `3b9dbd5` | `stripQuery` helper drops `?...` before logging path. |
| ME-05   | FIXED  | `4cb3075` | `SET NX EX` initializes counter+TTL atomically; falls back to Incr on collision. |
| ME-06   | FIXED  | `4d8c5f9` | All 6 `_ = json.Unmarshal` occurrences in `internal/ws/bridge.go` replaced with checked unmarshal + slog.Warn + return. |
| ME-07   | FIXED  | `714cbd0` | EVENTS stream gains `notification.>` and `calls.recording.>`. `calls.chat.>` intentionally left ephemeral. |
| ME-11   | FIXED  | `9f04961` | Webhook signature-mismatch log emits `expected_hash=sha256(APIKey)[:8]` instead of the raw key. |
| ME-13   | FIXED  | `804b2cd` | Redis cache `user:status:<uid>`, 30s TTL, with explicit invalidation in `audit.DeactivateUser`. |
| ME-15   | FIXED  | `95a92cc` | Fiber `BodyLimit: 25 * 1024 * 1024`. |

Verification per commit: `go build ./...` clean, plus `go test -count=1`
on the touched package(s). Final cross-package smoke run after ME-15 was
green for `auth`, `audit`, `ws`, `platform/nats`, `platform/middleware`,
`calls` (the only package families that were modified or that consume
the modified APIs).

---

## ME-01 — `invalidateUserSessions` N+1 → bulk delete + Redis pipeline

**Status:** FIXED
**Commit:** `5afbb71`
**Files touched:**
- `backend/internal/auth/service.go` — `invalidateUserSessions` rewritten:
  - One `Session.Query().Where(session.HasUserWith(user.ID(userID))).Select(session.FieldRefreshToken).Strings(ctx)` to collect tokens.
  - One `Session.Delete().Where(session.HasUserWith(user.ID(userID))).Exec(ctx)` bulk DELETE.
  - `s.redis.Pipeline()` over all refresh keys, single round-trip.

**Note:** Ent does not expose a direct `session.UserIDEQ` predicate
(the FK is modelled as an edge), so the edge-based predicate from the
original code is reused. Bulk delete still single-statement on the
Postgres side.

**Tests:** `go test ./internal/auth/...` — green.

---

## ME-03 — Logger leaks query strings (incl. WS `?token=`)

**Status:** FIXED
**Commit:** `3b9dbd5`
**Files touched:**
- `backend/internal/platform/middleware/logging.go`
  - Added unexported `stripQuery(string) string` helper.
  - Logger now passes `stripQuery(c.Path())` to the `path` slog attr.

**Note:** Only the access log is affected. Handlers continue to receive
the full `c.Path()` / `c.OriginalURL()` unchanged. This fixes the
specific WS upgrade case (`/ws?token=<JWT>`) plus any other endpoint
that ever passes a secret in the query string.

**Tests:** `go test ./internal/platform/middleware/...` — green.

---

## ME-05 — Magic-link rate-limit Incr+Expire race

**Status:** FIXED
**Commit:** `4cb3075`
**Files touched:**
- `backend/internal/auth/magiclink.go`
  - `SendMagicLink`: replaced `Incr` + `Expire-on-1` with
    `redis.SetNX(ctx, rateKey, 1, magicLinkRateTTL)`. If SetNX returns
    false (key exists) we `Incr` (TTL already set on the first hit) and
    compare against `magicLinkRateMax`.
- `recordFailure` (per-email failure counter) was left unchanged —
  its TTL is the longer code-TTL and isn't security-critical; a
  separate refactor would unify both paths.

**Tests:** `go test ./internal/auth/...` — green.

---

## ME-06 — `_ = json.Unmarshal` in WS bridge handlers

**Status:** FIXED
**Commit:** `4d8c5f9`
**Files touched:**
- `backend/internal/ws/bridge.go` — all six handlers
  (`handleProjectEvent`, `handleMessengerEvent`, `handleCallEvent`,
  `handleCallChatEvent`, `handleAuthSessionEvent`,
  `handleNotificationEvent`) now check the unmarshal error, `slog.Warn`
  the subject + error, and `return` instead of broadcasting `null`.

**Note:** Verified with a follow-up `Grep "_ = json.Unmarshal"` in
`bridge.go` — no matches remain.

**Tests:** `go test ./internal/ws/...` — green.

---

## ME-07 — NATS `EnsureStreams` missing subjects

**Status:** FIXED
**Commit:** `714cbd0`
**Files touched:**
- `backend/internal/platform/nats/client.go` — `EnsureStreams` stream
  `Subjects` extended with:
  - `notification.>` — uniformly durable so users don't miss pushes
    across a broker restart.
  - `calls.recording.>` — recording lifecycle events become durable
    (webhook dispatcher consumes from this stream).

**Note:** `calls.chat.>` deliberately omitted — call chat is real-time
only; replay would deliver stale chat messages after a restart. This
matches the comment in `handleCallChatEvent`.

**Tests:** `go test ./internal/platform/nats/...` — green
(test-container NATS spins up and the create-or-update call succeeds).

---

## ME-11 — Webhook handler logs the expected LiveKit API key

**Status:** FIXED
**Commit:** `9f04961`
**Files touched:**
- `backend/internal/calls/handler.go`
  - Added `encoding/hex` import.
  - `verifyWebhookSignature` mismatch branch now emits
    `expected_hash=hex(sha256(APIKey))[:8]` instead of the raw key.
    Attacker-controlled `got` value (the `iss` claim from the request)
    is preserved verbatim for diagnostics.

**Tests:** `go test ./internal/calls/...` — green
(testcontainers Mongo, ~11s).

---

## ME-13 — Per-request DB read in auth middleware

**Status:** FIXED
**Commit:** `804b2cd`
**Files touched:**
- `backend/internal/auth/middleware.go`
  - Added exported `UserStatusCacheKey(uid)` and
    `InvalidateUserStatusCache(ctx, rdb, uid)` so other packages can
    invalidate without re-implementing the key.
  - New unexported `userIsDeactivated(ctx, svc, uid)` performs:
    1. Redis `GET user:status:<uid>` (returns immediately on hit).
    2. On miss → `ent.User.Get(ctx, uid)` + Redis `SET` with 30s TTL.
  - `RequireAuth` calls the helper. Behaviour preserved on every
    failure path (DB error or Redis error both fall through to "not
    deactivated" — same fail-open semantics as before).
- `backend/internal/audit/service.go`
  - Added `redis.Client` import.
  - `AuditService` gains an optional `redis *redis.Client` field.
  - `NewAuditService` takes `rdb *redis.Client` (nil-safe — nil means
    no invalidation, deactivation still applies via 30s TTL).
  - `DeactivateUser` calls `auth.InvalidateUserStatusCache` after
    revoking sessions, so the new status is visible immediately.
- `backend/cmd/server/main.go`
  - `auditmod.NewAuditService(db, auditRepo, rdb)` — pass through the
    existing `rdb` client (no new wiring).

**Note:** Considered marking SKIP because the audit→auth import is a
new dependency direction (audit now imports auth). The cycle check is
clean — auth does not import audit. The wiring change is one extra
arg at one call site, which is well below the "complex wiring" bar.

**Tests:** `go test ./internal/auth/... ./internal/audit/...` — green.

---

## ME-15 — Explicit Fiber `BodyLimit`

**Status:** FIXED
**Commit:** `95a92cc`
**Files touched:**
- `backend/cmd/server/main.go`
  - `fiber.Config` gains `BodyLimit: 25 * 1024 * 1024`.

**Note:** 25 MB covers CSV import + reasonable avatar uploads. Endpoints
that need more can override per-route in their own `fiber.Config` /
group; none currently exceed 25 MB.

**Tests:** `go build ./...` + `go vet ./...` — clean.

---

## Wave 4 — Not in scope / deferred

The following MEDIUM findings from REVIEW.md were not selected for this
wave (per prompt). None were attempted in Wave 4:

- ME-02 — Refresh-token rotation TOCTOU (re-order store/delete).
- ME-04 — `RegistrationMode` upsert non-atomic.
- ME-08 — Webhook secret query in tight loop (no caching).
- ME-09 — Dev seeds expose admin-token paths.
- ME-10 — Cron tickers capture wrong context on shutdown.
- ME-12 — `uuid.Nil` not consistently rejected in parse helpers.
- ME-14 — Audit consumer loses in-flight events on shutdown.

---

_Wave 4 fixed: 2026-05-14_
_Fixer: Claude (Opus 4.7 1M context)_

---

# Wave 5 — follow-throughs + remaining MEDIUMs

**Fixed at:** 2026-05-15
**Source review:** `.planning/audits/code-review/REVIEW.md`
**Branch:** master
**Scope:** HI-07 follow-through, HI-12 follow-through, ME-02, ME-04,
ME-08, ME-09 (partial), ME-10 + ME-14 (unified), ME-12. HI-05
follow-through SKIPPED — see note below.

**Build:** `go build ./...` and `go build -tags release ./...` clean
after every commit.
**Tests:** every touched package re-run with `go test -count=1`, all
green.

## Summary

| Finding         | Status  | Commit    | Note |
|-----------------|---------|-----------|------|
| HI-07 follow-up | FIXED   | `e44c011` | Subscribe/Unsubscribe routed through new subscribe/unsubscribe channels; serialised with register/unregister. |
| HI-12 follow-up | FIXED   | `2130fee` | System-admin (`user.SystemRoleAdmin`) bypasses project-scoping via `ent.User.Get`. |
| HI-05 follow-up | SKIPPED | —         | Ent Hook integration not added — repo-layer encryption already in place; per brief "don't over-engineer". |
| ME-02           | FIXED   | `6f226ef` | Refresh-token rotation re-ordered: store new before delete old. |
| ME-04           | FIXED   | `b072c5c` | Race-safe SystemSetting upsert (Update → Create → on `IsConstraintError` retry Update). |
| ME-08           | FIXED   | `13d1174` | 60s in-memory TTL cache for webhook secrets (github + gitlab), invalidated on Create/Delete. |
| ME-09           | PARTIAL | `f62058a` | `//go:build !release` tags + `make build-prod`. Dockerfile not updated. |
| ME-10 + ME-14   | FIXED   | `de2faa8` | Single `signal.NotifyContext` rootCtx threaded into every goroutine; audit/dispatcher Stop() on defer. |
| ME-12           | FIXED   | `48fab28` | `uuid.Nil` rejected in 8 parse helpers. |

---

## HI-07 follow-through — funnel Subscribe/Unsubscribe through Hub channels

**Status:** FIXED
**Commit:** `e44c011`
**Files touched:**
- `backend/internal/ws/hub.go`
  - Added `subscribeReq{client, room}` and two new channels
    `subscribe chan subscribeReq` / `unsubscribe chan subscribeReq`
    (buffered at 256 to match register/unregister handling).
  - `Hub.Run` adds two `case req := <-h.subscribe / <-h.unsubscribe`
    branches that delegate to private `subscribeClient` /
    `unsubscribeClient` (these contain the actual `h.rooms` and
    `client.rooms` mutations that used to live in the public
    `Subscribe`/`Unsubscribe` methods).
  - The public API is unchanged — `Subscribe(c, room)` and
    `Unsubscribe(c, room)` simply post to the channel and return
    (send-and-forget per task brief).

**Effect:** all four state transitions (register, unregister,
subscribe, unsubscribe) now run in the same goroutine, in FIFO
order. The previous race (a quick disconnect+reconnect where
Subscribe for the new connection ran concurrently with the
disconnect's removeClient and could leave stale entries in
`h.rooms` or wipe membership of the fresh connection) is gone.
Per-connection keying from the original HI-07 commit is preserved.

**Tests:** `go test -count=1 ./internal/ws/...` — green (0.47s).
The existing TestHub_* fan-out / multi-tab regression tests
continue to pass with no modifications because they call
`registerClient` / `removeClient` directly (still synchronous
package-private), not the channel-backed public Register/Unregister.

---

## HI-12 follow-through — system-admin override for user search

**Status:** FIXED
**Commit:** `2130fee`
**Files touched:**
- `backend/internal/search/service.go`
  - Added import `ent/user` for the `SystemRoleAdmin` constant.
  - New unexported `isSystemAdmin(ctx, userID)` — `ent.User.Get` +
    role compare. Same pattern as `rbac.RequireAdmin`
    (`rbac/middleware.go:113-116`) and
    `rbac.GetEffectivePermissions` (`rbac/service.go:46`).
  - `searchUsers` short-circuits to a new unexported
    `searchUsersUnscoped(ctx, query)` when the caller is admin.
    `searchUsersUnscoped` drops the project-shared-member sub-query
    and runs a plain ranked tsvector search.

**Effect:** non-admins keep the project-scoped behaviour from Wave 3.
System admins (rare — `SystemRoleAdmin`) now see all matching users,
matching the directory-management use case.

**Why not cache the admin check:** one `ent.User.Get` per search
query is fine — admins are rare and the rest of the search already
does multi-second parallel queries to Postgres + Mongo. Caching
would add complexity without measurable benefit.

**Tests:** `go test -count=1 ./internal/search/...` — green.

---

## HI-05 follow-through — Ent Hook integration

**Status:** SKIPPED
**Note:** Per the task brief: *"If neither is clean, SKIP and leave
as-is (repo-layer is already secure). Don't over-engineer."*

Reading `ent/githubintegration_create.go` confirmed Ent in this repo
exposes Hook() at the schema level only via codegen. Wiring a global
cipher into a schema-level Hook would either:

1. require a package-level singleton accessor in `internal/platform/crypto`,
   making cipher initialisation a hidden import-time side-effect; OR
2. require regenerating Ent code with a custom template that allows
   `Hook` to receive runtime parameters,

neither of which is "one commit clean". The Ent Interceptor API does
not post-process field values on read in any version pinned here, so
decryption-on-load via an interceptor isn't viable either.

The repository-layer encryption from HI-05 already provides the
at-rest guarantee. Skipping per brief; documented as residual gap
in this report.

---

## ME-02 — Refresh-token rotation TOCTOU

**Status:** FIXED
**Commit:** `6f226ef`
**Files touched:**
- `backend/internal/auth/service.go`
  - `RefreshToken` re-ordered to:
    1. generate new access + refresh
    2. `storeRefreshToken(newRefresh, *payload)`
    3. `Session.UpdateOneID(...).SetRefreshToken(newRefresh)`
    4. `deleteRefreshToken(oldRefreshToken)` last
  - Step 3 failure now returns an error (was: log + continue) so
    the client can retry safely; the old refresh remains valid in
    Redis until a future retry completes the rotation.

**Idempotency:** A retry with the old refresh token after a crash
between any pair of steps simply restarts at step 1 and overwrites
the partial state. No more "user permanently locked out" window.

**Tests:** `go test -count=1 ./internal/auth/...` — green.

---

## ME-04 — Race-safe SystemSetting upsert

**Status:** FIXED
**Commit:** `b072c5c`
**Files touched:**
- `backend/internal/auth/service.go`
  - Verified Ent API: `ent/systemsetting_create.go` has no
    `OnConflict*` methods — Ent in this repo does not expose
    upsert. Confirmed via `Grep "OnConflict|Upsert"` on that file.
  - `SetRegistrationMode` now forwards to a new
    `upsertSystemSetting(ctx, key, value, description)` helper.
  - Helper algorithm: Update → on `n==0` Create → on `IsConstraintError`
    retry Update (which is now guaranteed to hit the row a
    concurrent caller created in between).

**Effect:** two admins racing on `SetRegistrationMode` no longer
produce a 500 on the second call.

**Tests:** auth — green.

---

## ME-08 — Webhook secret TTL cache

**Status:** FIXED
**Commit:** `13d1174`
**Files touched:**
- `backend/internal/github/repository.go`,
  `backend/internal/gitlab/repository.go`
  - Each `EntRepository` gains `secretsMu sync.RWMutex`,
    `secretsCache []string`, `secretsExpire time.Time`.
  - `ListAllWebhookSecrets` reads under RLock when cache is fresh
    (`time.Now().Before(secretsExpire)`); on miss, falls through
    to the existing `Select(FieldWebhookSecret)` + `decryptField`
    path, then writes the result back into the cache under Lock.
    A defensive copy is returned to every caller so the cache
    slice cannot be mutated externally.
  - TTL = 60 seconds (constant `webhookSecretsCacheTTL`).
  - `CreateIntegration` and `DeleteIntegration` call
    `invalidateSecretsCache()` so new / removed secrets take
    effect immediately. If invalidation is missed, worst-case
    staleness is the 60s TTL.

**Effect:** GitHub push-spam on big repos no longer issues N
identical Postgres SELECTs per second.

**Tests:** github, gitlab — green.

---

## ME-09 — Build tag for dev endpoints

**Status:** PARTIAL (Dockerfile not updated)
**Commit:** `f62058a`
**Files touched:**
- `backend/internal/dev/handler.go`, `seed.go`, `seed_data.go`,
  `seed_test.go` — prefixed with `//go:build !release`.
- `backend/internal/dev/handler_release.go` (new) —
  `//go:build release` stub providing a no-op `Handler`,
  `NewHandler(db, authSvc)` and `RegisterRoutes(router, devMode)`
  identical in signature to the dev build. main.go compiles
  unmodified under both build tags.
- `backend/internal/dev/apidocs.go` is intentionally untagged —
  it only renders the registered-routes catalogue and contains no
  privileged endpoints.
- `backend/Makefile` — new `make build-prod` target running
  `go build -tags release -o bin/server ./cmd/server`.

**Effect under `-tags release`:** `/api/dev/login` and
`/api/dev/seed` are *physically not in the binary*. DEV_MODE env
var can no longer toggle them on at runtime.

**Partial — what is NOT done:** Dockerfile (`deployments/Dockerfile`
or equivalent) was not updated to pass `-tags release`. Operators
must build via `make build-prod` for now.

**Tests:** dev — green under default build; no tests under
release build (the stub has no behaviour to test).

---

## ME-10 + ME-14 — Single root context for graceful shutdown

**Status:** FIXED
**Commit:** `de2faa8`
**Files touched:**
- `backend/cmd/server/main.go`
  - `rootCtx, rootCancel := signal.NotifyContext(context.Background(), SIGINT, SIGTERM)`
    at the top; `defer rootCancel()`.
  - Every per-goroutine `context.WithCancel(ctx)` + `defer xxxCancel()`
    deleted (12 cancels). The goroutines themselves now select on
    `<-rootCtx.Done()` instead of their local ctx.
  - Bottom-of-main signal handling (`signal.Notify(quit, …); <-quit`)
    replaced by `<-rootCtx.Done()`.
  - `defer auditConsumer.Stop()` and `defer webhookDispatcher.Stop()`
    added so the JetStream consumer drains in-flight Ack-s (ME-14)
    and the bounded webhook worker pool drains (works with HI-08).
- The `ctx` alias is kept as a friendly name for bootstrap-time
  calls (DB connect, EnsureStreams, EnsureIndexes); it points at
  rootCtx so even bootstrap aborts cleanly on SIGINT before
  Listen.

**Effect:** SIGTERM now propagates to every goroutine in the same
tick instead of staggering through `defer xxxCancel()` calls that
ran only after main() returned (which itself blocked on
app.Shutdown). Audit consumer's `Stop()` now blocks the deferred
shutdown sequence until the JetStream consumer has drained.

**Verified:** `go build ./...` and `go build -tags release ./...`
both clean. Per-package tests for outbox, audit, apigateway —
green.

---

## ME-12 — Reject uuid.Nil in parse helpers

**Status:** FIXED
**Commit:** `48fab28`
**Files touched:**
- `backend/internal/platform/http/helpers.go` — `ParseUUIDParam`
- `backend/internal/calls/handler.go` — `parseCallID`
- `backend/internal/messenger/handler.go` — `parseChannelID`
- `backend/internal/messenger/handler_reactions.go` — `parsePollID`
- `backend/internal/github/handler.go` — `parseProjectID`
- `backend/internal/gitlab/handler.go` — `parseProjectID`
- `backend/internal/imex/handler.go` — `parseProjectID`
- `backend/internal/tms/environment.go` — `parseTMSProjectID`

All eight helpers now return 400 when `uuid.Parse` succeeds but
yields `uuid.Nil` ("00000000-0000-0000-0000-000000000000"). The
zero UUID is never a legitimate domain identifier in this
codebase.

**Tests:** platform/http, messenger, calls, github, gitlab, imex,
tms — all green.

---

## Skipped / Not in this wave

- **HI-05 follow-through (Ent Hook)** — see note above. Repository-
  layer encryption is already in place from the original HI-05
  commit; an Ent Hook integration would require either a package-
  level cipher singleton or a custom codegen template, both
  rejected by the brief.

- **ME-09 Dockerfile update** — out of scope for this commit per
  brief; `make build-prod` is the documented path.

- Anything not listed in the Wave 5 task brief is unchanged.

---

_Wave 5 fixed: 2026-05-15_
_Fixer: Claude (Opus 4.7 1M context)_

---

# Wave 6 — LOW-severity + light DRY/SOLID

**Fixed at:** 2026-05-15
**Source review:** `.planning/audits/code-review/REVIEW.md`
**Branch:** master
**Scope:** LO-01, LO-02 (partial scope per brief), LO-04, LO-05, LO-06
(two guard tests only), LO-08. LO-07 explicitly skipped per brief.
LO-03 not in scope (architectural refactor, out of "light-touch" budget).

**Build:** `go build ./...` and `go build -tags release ./...` clean
after every commit.
**Tests:** every touched package re-run with `go test -count=1`, all
green. New tests in `internal/ws` and `internal/apigateway` run under
`-race`.

## Summary

| Finding | Status   | Commit    | Note |
|---------|----------|-----------|------|
| LO-01   | FIXED    | `9b39133` | New `internal/platform/telegram` package; both callers re-routed. |
| LO-02   | FIXED    | `1179cd1` | 12 handlers migrated to `httputil.ParseUUIDParam`. |
| LO-04   | FIXED    | `35c3b71` | 3 magic numbers → constants; `maxActiveRecordingsPerCall` was already a const. |
| LO-05   | FIXED    | `226c213` | One genuine double-log site fixed; rest of scan was false-positive. |
| LO-06   | FIXED    | `253cc53` | Two guard tests added (WS multi-tab + webhook SSRF). |
| LO-07   | SKIPPED  | —         | Already addressed in HI-07 Wave 5 (per brief). |
| LO-08   | FIXED    | `57ded74` | `requireAuth` (search) and `authMiddleware` (auth) renamed → `authMw`. |

---

## LO-01 — Duplicated `escapeMarkdownV2`

**Status:** FIXED
**Commit:** `9b39133`
**Files touched:**
- `backend/internal/platform/telegram/markdown.go` (new) —
  `EscapeMarkdownV2(s string) string`. Same special-character set as
  the two duplicates. Comment explains it's the single source of truth
  per LO-01.
- `backend/internal/platform/telegram/markdown_test.go` (new) —
  roundtrip + edge case (all-special) coverage.
- `backend/internal/notification/telegram.go` — drops local
  `escapeMarkdownV2`; imports the new helper. The `strings` import
  is dropped too (no other use).
- `backend/internal/auth/telegram_bot.go` — drops local
  `escapeMarkdownV2TG`; imports the new helper. `strings` import
  retained (still used by TrimSpace / SplitN / HasPrefix).
- `backend/internal/notification/service_test.go` — existing
  `TestEscapeMarkdownV2` now calls the new exported helper.

**Tests:** `go test ./internal/notification/... ./internal/auth/...
./internal/platform/telegram/...` — green.

---

## LO-02 — `uuid.Parse(c.Params(...))` boilerplate → `httputil.ParseUUIDParam`

**Status:** FIXED
**Commit:** `1179cd1`
**Files touched:**
- `backend/internal/dashboard/handler.go` — 1 call site (`projectId`).
- `backend/internal/rbac/handler.go` — full rewrite (8 call sites):
  `CreateRole`, `ListRoles`, `UpdateRole`, `DeleteRole`, `AssignRole`,
  `UnassignRole`. Uses both `ParseProjectID` and `ParseUUIDParam` for
  the per-role / per-user IDs. Direct `uuid` import dropped because
  no other usages remained in this file.
- `backend/internal/search/handler.go` — 2 call sites (filter ID for
  delete + share).
- `backend/internal/notification/handler.go` — 2 call sites
  (`MarkRead`, `Delete`).
- `backend/internal/storage/handler.go` — 3 call sites
  (`ConfirmUpload`, `GetDownloadURL`, `DeleteFile`).

`httputil.ParseUUIDParam` was added in Wave 5 (ME-12) and already
rejects empty params, malformed UUIDs and `uuid.Nil`. The migrated
handlers gain that defensive check "for free" (previously
`uuid.Parse("00000000-...")` would succeed and pass uuid.Nil down to
the service).

**Out of scope per brief:**
- `internal/calls/handler.go`, `internal/messenger/handler.go`,
  `internal/github/handler.go`, `internal/gitlab/handler.go`,
  `internal/imex/handler.go`, `internal/tms/environment.go` — these
  already have package-local helpers (`parseCallID`, `parseChannelID`,
  `parseProjectID`, `parsePollID`, etc.) that were rewritten in Wave
  5 ME-12 to centralise the uuid.Nil check. Migrating those to
  `ParseUUIDParam` is a follow-up cleanup, not a behavioural change,
  and was deliberately deferred to keep the diff scoped.
- `internal/tracker/...` — uses neither pattern; reads issue keys
  (`PROJ-123`), not raw UUIDs.

**Tests:** `go test ./internal/dashboard/... ./internal/rbac/...
./internal/search/... ./internal/notification/... ./internal/storage/...`
— green.

---

## LO-04 — Magic numbers → named constants

**Status:** FIXED
**Commit:** `35c3b71`
**Files touched:**
- `backend/internal/apigateway/webhook.go` —
  `maxWebhookResponseBodyBytes = 4096` (cap on `webhook_delivery.response_body`,
  CR-03 SSRF read primitive defence). Replaces the two `4096` literals
  in `sendWebhook`.
- `backend/internal/calls/service.go` —
  `defaultMaxCallParticipants = 50`. Replaces both `MaxParticipants: 50`
  literals (CreateCall + CreateInstantCall).
- `backend/internal/search/service.go` — `defaultSearchLimit = 10`.
  Applied via `fmt.Sprintf("... LIMIT %d", defaultSearchLimit)` to
  seven `ORDER BY rank DESC LIMIT 10` queries and one
  `ORDER BY name ASC LIMIT 10` fallback; mongo `SetLimit(10)` →
  `SetLimit(int64(defaultSearchLimit))`. The channels-search fallback
  using `name ILIKE '%' || $1 || '%'` had its `%` doubled to `%%` to
  survive the new `fmt.Sprintf` wrapper.

**Verified:** `maxActiveRecordingsPerCall = 2` in
`internal/calls/recording.go:24` was already a named constant — no
change.

**Out of scope (intentional):** dozens of other magic numbers across
the codebase (TTLs, page sizes, JWT durations, etc.). The four flagged
in the review are the ones with the highest "future tuning" cost.

**Tests:** `go test ./internal/apigateway/... ./internal/calls/...
./internal/search/...` — green.

---

## LO-05 — Double-log before wrapped error return

**Status:** FIXED
**Commit:** `226c213`
**Files touched:**
- `backend/internal/auth/service.go` — in `RefreshToken`, the
  session-update failure path had both `slog.Error(...)` and
  `return nil, fmt.Errorf("...: %w", saveErr)`. The handler logs the
  returned error already, so the same error was being emitted to log
  pipelines twice (once by service, once by handler). Dropped the
  `slog.Error` and added a comment explaining the
  log-only-at-topmost-layer convention.

**Why no other sites in auth/service.go / calls/service.go:**
A multi-line `Grep` for `slog\.Error[^\n]*\n\s+[^\n]*fmt\.Errorf[^\n]*%w`
found exactly one match (the one fixed above). Other `slog.Error`
calls in those files are either:
- Terminal logs (e.g. `invalidateUserSessions` returns `void`; logs
  are the only signal),
- Fire-and-forget event publish (`publishSessionCreated`,
  `OnEgressEnded` NATS publish),
- Inside NATS consumer handlers where the error doesn't propagate up.

The brief explicitly says "don't touch event handlers (NATS consumers
etc.) where the error doesn't propagate up", which matches.

**Tests:** `go test ./internal/auth/...` — green.

---

## LO-06 — Two new guard tests

**Status:** FIXED
**Commit:** `253cc53`
**Files touched:**
- `backend/internal/ws/hub_test.go` (new):
  - `TestHub_MultiTab_ChannelBackedRegisterAndSendToUser` — spins up
    a `NewHub()` with `Run(ctx)` in a goroutine, registers two
    clients of the same userID via the public `Register()` channel,
    waits for the Hub to observe both connections, calls
    `SendToUser(userID, msg)`, asserts both clients' `send` channels
    receive the message. Verifies the channel-routing introduced in
    Wave 5 HI-07 follow-through still fans out correctly.
  - Helpers: `Hub.connectionsForUser(userID) int` (test-only count
    of registered connections per user) and `waitFor(timeout, cond)`
    (poll-with-timeout).
- `backend/internal/apigateway/webhook_test.go` (new):
  - `TestSendWebhook_BlocksLoopbackSSRF` — calls
    `dispatcher.sendWebhook` on `http://127.0.0.1:9999/`, plus the
    other classic SSRF targets: `localhost`, AWS IMDS
    `169.254.169.254`, RFC1918 `10.x` / `192.168.x`, IPv6 `[::1]`.
    Asserts every one fails with `status=0`, empty body, and an
    error message containing `private` / `blocked` / `отклонён
    небезопасный url` — covering both rejection paths
    (ValidateOutboundURL pre-flight + SafeDialContext during dial).
  - `TestSendWebhook_RejectsBadScheme` — `file://`, `gopher://`,
    userinfo URL, empty host — all should fail at the
    `ValidateOutboundURL` gate.
  - `TestValidateOutboundURL_RejectsKnownBad` — direct check of the
    helper against `file://`, `javascript:`, `data:`, `ftp://`,
    `http://` (empty host).

**Why this scope:** the brief asked for *guard* tests so future
refactors don't silently regress the multi-tab fan-out (HI-07) or
webhook SSRF protection (CR-03). Both tests run via the public API
that production uses, so any change to the channel-routing or to
the dispatcher's HTTP client surfaces as a test failure.

**Tests:** `go test -count=1 -race ./internal/ws/...
./internal/apigateway/...` — green.

---

## LO-07 — Skipped

**Status:** SKIPPED
**Rationale:** per task brief, this finding is already addressed by
the HI-07 follow-through commit in Wave 5 (`e44c011`). That commit
funnelled `Subscribe` / `Unsubscribe` through the same channel as
`register` / `unregister`, so all four state transitions now run in
the same goroutine — exactly the fix LO-07 sketched at a smaller
scope.

No code change in Wave 6.

---

## LO-08 — Standardize `authMw` parameter name

**Status:** FIXED
**Commit:** `57ded74`
**Files touched:**
- `backend/internal/search/handler.go` — `requireAuth fiber.Handler`
  → `authMw fiber.Handler`. Two internal call sites updated.
- `backend/internal/auth/handler.go` — `authMiddleware fiber.Handler`
  → `authMw fiber.Handler`. Three internal call sites updated
  (auth.Group and admin.Group).

Audit of all `RegisterRoutes` signatures (19 handlers) shows the
remaining 17 already used `authMw fiber.Handler`. Pure parameter
rename — no behaviour change, no test change, no caller-site change
(call sites pass middlewares positionally).

**Tests:** `go test ./internal/search/... ./internal/auth/...` — green.

---

## Out of scope this wave (deliberately)

- **LO-03 (SOLID DIP):** "service constructors take both repo and
  `*ent.Client`". This is an architectural refactor that touches
  every service in the codebase. Not in the "light-touch" budget.

- **LO-02 (full repo migration):** `parseCallID`, `parseChannelID`,
  `parseProjectID`, etc. in messenger/calls/tracker still exist as
  package-private helpers (touched in Wave 5 ME-12). They could be
  removed in favour of `httputil.ParseUUIDParam` calls, but the
  current state is functionally identical and the migration is
  cosmetic-only.

- **LO-06 (more tests):** Only the two highest-value guard tests
  were added. The other suggestions in the review (RBAC integration
  test, outbox panic-recovery test, race-condition tests for
  imex/JoinByInvite, LiveKit webhook signature test) are still
  open follow-ups.

---

_Wave 6 fixed: 2026-05-15_
_Fixer: Claude (Opus 4.7 1M context)_

---

_Fixed: 2026-05-14_
_Fixer: Claude (Opus 4.7 1M context)_
