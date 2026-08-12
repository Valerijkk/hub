# Hub Backend — Code Review Report

**Reviewed:** 2026-05-14
**Scope:** `backend/internal/**` (all ~30 packages), `cmd/server/main.go`, `internal/platform/**`
**Depth:** Standard with deep dive into auth, calls, messenger, ws, apigateway, storage, webhooks
**Reviewer:** Claude (Opus 4.7 1M context)

---

## TL;DR

The codebase shows clear architectural intent (DDD-style packages, ent ORM, NATS event bus, RBAC), reasonable code quality, and decent test coverage in some areas — but there are **multiple CRITICAL security issues** that would be exploitable today by any authenticated user:

1. **Admin backup endpoints (`/admin/backups`) have NO admin guard** — any logged-in user can trigger backups, restore from arbitrary backups (full DB wipe), or delete backups.
2. **Recording download returns an unsigned, public S3 URL** — knowledge of a recording UUID allows anyone (no auth check on download) to grab the file forever.
3. **Webhook dispatcher has no SSRF protection** — any project member who can create a webhook can pivot the server into the internal network (169.254.169.254, RDS, etc.).
4. **JWT secret defaults to a weak hardcoded string** with no production guard, and `DEV_MODE=true` simply switches off the auth rate-limit and adds a no-password `/dev/login` endpoint.
5. **Most messenger read endpoints (`GetChannel`, `ListMessages`, `ListMembers`, `ListThreadReplies`) skip the membership check** — any authenticated user can read every channel in the system if they know/guess the UUID.

The biggest cross-cutting concern is **broken RBAC**: `rbac.RequirePermission` reads `c.Params("projectId")` — but several routes (most notably **all `/files/*` routes**) are registered without that path segment, which means the file upload/download/delete routes will reject every request with 400 "project ID required". This indicates the test coverage of the integration boundary between Fiber routing and the RBAC middleware is missing.

There is no Fiber recover/panic middleware anywhere. A single nil-deref in any handler (and there are real candidates — see issues below) crashes the request. Combined with `SaveX` panics in the outbox worker, a goroutine panic can take down the whole server.

---

## Severity Legend

- **CRITICAL** — exploitable auth bypass, RCE, data loss, or remote DoS
- **HIGH** — serious bug, broken functionality, exploitable IDOR within a tenant, missing auth check
- **MEDIUM** — bug that degrades quality / leaks information / racy data
- **LOW** — style, duplication, missing test, cosmetic

---

## CRITICAL

### CR-01 — Admin backup endpoints (`/api/admin/backups/*`) have NO admin role check

**File:** `internal/backup/handler.go:21-27`

```go
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
    backups := api.Group("/admin/backups", authMw)
    backups.Post("/", h.TriggerBackup)
    backups.Get("/", h.ListBackups)
    backups.Post("/:id/restore", h.RestoreBackup)
    backups.Delete("/:id", h.DeleteBackup)
}
```

Only `authMw` (JWT) is applied — there is **no `requireAdmin` / `RequireSystemRole` check**. Any authenticated user (including a brand new self-registered "member") can:

- `POST /api/admin/backups` — trigger a full `pg_dump` + `mongodump` on demand (resource exhaustion + leakage of all data once they read it back via `ListBackups`).
- `POST /api/admin/backups/:id/restore` — **rollback the entire platform to an arbitrary previous backup** = total data loss / time-travel attack. The only "protection" is `{"confirm": "RESTORE"}` in the body — that's not authorization, that's a confirmation dialog.
- `DELETE /api/admin/backups/:id` — delete a backup (anti-forensics).

Compare to `internal/audit/handler.go:25` which correctly applies `h.requireAdmin()` middleware.

**Fix:** add the admin guard, and prefer reusing an existing admin middleware:

```go
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, adminMw fiber.Handler) {
    backups := api.Group("/admin/backups", authMw, adminMw)
    // ...
}
```

In `cmd/server/main.go`, wire a `requireAdmin := rbac.RequireAdmin(db)` and pass it.

---

### CR-02 — `GetRecordingDownloadURL` returns a permanent, unsigned public S3 URL

**File:** `internal/calls/recording.go:206-229`

```go
url := fmt.Sprintf("%s://%s/%s/%s", protocol, s.s3Cfg.Endpoint, s.s3Cfg.Bucket, rec.FilePath)
return &RecordingDownloadResponse{ URL: url, ExpiresAt: expiresAt }, nil
```

The comment on line 217 honestly admits *"Для простоты возвращаем прямой путь — в production используется presigned URL"* — but that's exactly what's deployed. Issues:

1. The URL is **not presigned**: it bypasses S3 ACLs. If the bucket is public-readable (default for SeaweedFS in many deployments), anyone with the URL can download the call recording forever.
2. The endpoint URL points to `cfg.S3.Endpoint` which is the *internal* SeaweedFS service URL (default `http://localhost:8333`). External users will get a broken link, OR worse, if the endpoint is exposed publicly, the URL works for everyone on the internet who can guess `recordings/{projectId}/{callId}.mp4`.
3. `ExpiresAt` is fabricated for the response — the URL itself has no expiry.
4. The handler `GetRecordingDownloadURL` (`internal/calls/handler.go:60-61`, `handler_recording.go:137`) has **no RBAC check** — only `authMw`. Any logged-in user with a recording UUID can call it.

**Fix:** Use the `storage.StorageService.presignClient` (same as `GenerateDownloadURL`). Also add an authorization check that the caller is a member of the project the recording belongs to:

```go
// Verify caller has PermProjectView on rec.ProjectID
// Then presign:
ps, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
    Bucket: aws.String(s.s3Cfg.Bucket),
    Key:    aws.String(rec.FilePath),
}, s3.WithPresignExpires(1*time.Hour))
```

Apply the same fix to `ListProjectRecordings` (handler_recording.go:94) — it currently has no permission check either, so anyone can list recordings of any project.

---

### CR-03 — Webhook dispatcher has no SSRF protection — server-side request forgery to internal network

**File:** `internal/apigateway/webhook.go:180-206`, `internal/apigateway/webhook.go:238-255`

`CreateEndpoint` accepts any `req.URL` from the user. `WebhookDispatcher.sendWebhook` then calls `http.NewRequest(http.MethodPost, url, ...)` with `client.Do(req)` — no destination validation.

Combined with the fact that the webhook **POSTs the full event payload** (which may include sensitive project data), a project owner can:

- Point a webhook at `http://169.254.169.254/latest/meta-data/iam/security-credentials/` (AWS metadata service) — but since the dispatcher POSTs and doesn't read the response in detail, this is limited. **However**, `sendWebhook` does capture up to `4096` bytes of the response body and stores it in `webhook_delivery.response_body` (which the owner can later read via `ListDeliveries`). That's a full SSRF read primitive.
- Point at `http://localhost:6379` (Redis), `http://postgres:5432`, `http://livekit:7880` — internal services.
- Use it for amplification / port scanning of internal network.
- Combine with DNS rebinding to bypass naive IP allowlists.

The HTTP client has no `Transport` with a `DialContext` that blocks RFC1918 / link-local / loopback ranges.

**Fix:**

```go
client := &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        DialContext: blockPrivateNetDialer.DialContext,
    },
}
```

Where `blockPrivateNetDialer` is a `net.Dialer` wrapping `Control` that calls `getsockopt` and refuses RFC1918, 127.0.0.0/8, ::1, fe80::/10, 169.254.0.0/16. There are open-source helpers (`safedialer`, `go-safe-resolver`). Also validate the URL scheme is `https` (or http only on allowlisted hosts in dev), reject userinfo URLs, reject `file://`, etc.

The same fix applies to `internal/github/preview.go` (link preview — any URL fetched server-side) — please grep for any place we make an outbound HTTP request with a URL derived from user input.

---

### CR-04 — `/files/*` routes are broken AND lack proper project authorization

**File:** `internal/storage/handler.go:25-33` in combination with `internal/rbac/middleware.go:74-111`

```go
files := api.Group("/files")
files.Use(authMw)
files.Post("/upload-url", rbac.RequirePermission(rbac.PermFileUpload, rbacSvc), h.GenerateUploadURL)
files.Post("/:fileId/confirm", rbac.RequirePermission(rbac.PermFileUpload, rbacSvc), h.ConfirmUpload)
files.Get("/:fileId/url",      rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetDownloadURL)
files.Delete("/:fileId",       rbac.RequirePermission(rbac.PermFileDelete, rbacSvc), h.DeleteFile)
```

`rbac.RequirePermission` does:

```go
projectIDStr := utils.CopyString(c.Params("projectId"))
if projectIDStr == "" {
    return fiber.NewError(fiber.StatusBadRequest, "project ID required")
}
```

Routes under `/files` do **not** have `:projectId` in the path. So every file route returns **400 "project ID required"** — i.e. the entire file module is non-functional. Either:

- This is what's actually happening in prod (file upload is broken — easy to verify with `curl`), OR
- There's some shim I'm not seeing (would mean `RequirePermission` accepts a projectId from elsewhere, but the code shows it doesn't).

Even when the routes get fixed, **`GenerateUploadURL` does not check that the user is a member of `req.ProjectID`** — `RequirePermission` checks the projectId in the URL path, but the body contains a different projectId field (`req.ProjectID` at line 80 of storage/handler.go). And `GetDownloadURL` / `DeleteFile` / `ConfirmUpload` don't load the file's `ProjectID` and re-check membership — so once we put the routes under `/projects/:projectId/files/*` we'd still be vulnerable to IDOR (alice's projectId in the URL, bob's fileId — RBAC passes because alice is in her own project, but file belongs to bob).

**Fix:**

1. Move routes under `/projects/:projectId/files/*` OR change `RequirePermission` to also accept project lookups via fileID (e.g. a `RequireFilePermission` middleware that loads the file's project from DB).
2. In `GenerateUploadURL`, validate that the user has `PermFileUpload` on **`req.ProjectID`** (not whatever's in the path).
3. In `GetDownloadURL`, `DeleteFile`, `ConfirmUpload` — load the file, then check the caller has `PermProjectView` (or appropriate perm) on `file.ProjectID`.

---

### CR-05 — JWT secret has weak hardcoded fallback with no production guard

**Files:** `internal/platform/config/config.go:129`, `internal/auth/service.go:67-79`, `cmd/server/main.go:110-117`

```go
JWT: JWTConfig{
    Secret: getEnv("JWT_SECRET", "dev-secret-change-in-production"),
    ...
}
```

The docker-compose file ships with `JWT_SECRET="local-dev-secret-change-in-production"`. There is **no startup validation** that the secret is not equal to one of the well-known dev values, no minimum length check, no warning, no refusal to start in non-dev mode.

If anyone deploys the docker-compose unchanged (which is exactly what the `local-dev-secret-change-in-production` value invites), an attacker who knows the value (it's in your public repo / image) can mint admin JWTs against the deployment.

**Fix in `config.Load()`:**

```go
secret := getEnv("JWT_SECRET", "")
if !cfg.Server.DevMode {
    if secret == "" || strings.Contains(strings.ToLower(secret), "change-in-production") || len(secret) < 32 {
        return nil, fmt.Errorf("JWT_SECRET must be set to a strong value (>=32 bytes) when DEV_MODE=false")
    }
}
```

Same concern applies to:
- `S3_ACCESS_KEY=hub` / `S3_SECRET_KEY=hubsecret` defaults (config.go:122-123)
- `LIVEKIT_API_KEY=devkey` / `LIVEKIT_API_SECRET=devsecret` defaults (config.go:142-143)

All should be hard-fail when DevMode=false and the value matches the dev default.

---

### CR-06 — `DEV_MODE=true` exposes unauthenticated login endpoint

**File:** `internal/dev/handler.go:85-116`, gated only by `cfg.Server.DevMode` env flag.

```go
func (h *Handler) Login(c fiber.Ctx) error {
    var req DevLoginRequest
    // ... only requires email
    u, err := h.db.User.Query().Where(user.EmailEQ(req.Email)).Only(ctx)
    tokens, _ := h.authSvc.CreateSession(ctx, u.ID, "dev-login", c.IP())
    return c.JSON(...)
}
```

Anyone who reaches `POST /api/dev/login` with `{"email":"admin@hub.local"}` gets a full admin JWT, **including refresh token**, no password, no magic link, no Telegram. This is by design for dev — but two things make it dangerous:

1. The CI commit history (`80ccdef fix: remove DEV_MODE=true override in deploy, use env var` and the immediate revert `278e0f7 Revert "fix: remove DEV_MODE=true override in deploy..."`) shows the deployed environment has had `DEV_MODE=true` at least temporarily, possibly currently. If the deploy still hardcodes it, the staging/production has an unauthenticated admin login.
2. The auth rate-limit middleware is **skipped** when `DevMode=true` (`cmd/server/main.go:154-156`):
   ```go
   if !cfg.Server.DevMode {
       api.Use("/auth", middleware.RateLimit(middleware.AuthRateLimit))
   }
   ```
   So even setting aside dev/login, dev mode also removes brute-force protection from magic-link/Telegram endpoints.

**Fix:** beyond the obvious "don't deploy with DEV_MODE=true":

- In `main.go`, add a startup log line like `slog.Warn("DEV_MODE enabled — /api/dev/* exposed, auth rate-limit disabled")` on every boot if `DevMode`. Make it loud.
- Refuse to start at all if both `DevMode=true` AND the deploy is detected as non-local (e.g. binding to a public IP, or env var like `ENVIRONMENT=production`).
- Stop relying on a single boolean for two very different things (dev endpoints + rate-limit toggle) — split into `EXPOSE_DEV_ENDPOINTS` and `DISABLE_AUTH_RATE_LIMIT`. Both need explicit opt-in.

---

## HIGH

### HI-01 — Messenger read endpoints have no channel-membership check

**Files:** `internal/messenger/service_channels.go:65-74` (GetChannel), `service_messages.go:97-107` (ListMessages), `service_channels.go:259-278` (ListMembers), `service_messages.go:181-191` (ListThreadReplies)

```go
func (s *MessengerService) GetChannel(ctx context.Context, channelID uuid.UUID) (*ChannelResponse, error) {
    ch, err := s.repo.GetChannel(ctx, channelID)
    // ... returns the channel — NO membership check
}

func (s *MessengerService) ListMessages(ctx context.Context, channelID uuid.UUID, before *time.Time, limit int) ([]MessageResponse, error) {
    docs, err := s.repo.ListMessages(ctx, channelID, before, limit)
    // ... returns ALL messages — NO membership check
}
```

`SendMessage` (correctly) calls `s.repo.IsMember`. The read-side counterparts don't. Any authenticated user who learns a channel UUID can read every message in a private channel, including DMs (channel_type=`dm`).

The handler-level RBAC middleware doesn't apply because messenger routes don't use `RequirePermission` (see `handler.go:55-76`).

**Fix:** at the top of `GetChannel`, `ListMessages`, `ListThreadReplies`, `ListMembers`, `ListPins`, add:

```go
isMember, err := s.repo.IsMember(ctx, channelID, userID)
if err != nil { return nil, apperrors.NewInternal(...) }
if !isMember { return nil, apperrors.NewForbidden("not a channel member") }
```

This also means changing the method signatures to accept `userID uuid.UUID`. For DMs and private channels this is mandatory. For `public` channels you may want to relax it (anyone in the project can see), but the current code allows reading even DMs.

---

### HI-02 — `ForwardMessage` doesn't validate caller is a member of the SOURCE channel

**File:** `internal/messenger/service_messages.go:194-239`

```go
func (s *MessengerService) ForwardMessage(ctx context.Context, channelID, userID uuid.UUID, sourceMessageID string, targetChannelID string) (*MessageResponse, error) {
    targetID, err := uuid.Parse(targetChannelID)
    isMember, err := s.repo.IsMember(ctx, targetID, userID)  // checks TARGET only
    source, err := s.repo.FindMessageByID(ctx, sourceMessageID) // any message
    // ... copies source.ContentJSON into a new message
}
```

Authenticated user Alice in channel-A can forward any message UUID she's heard of (e.g. from logs, from a server-side bug, from URL leakage, brute-force) from Bob's private channel-B into Alice's channel-A — copying message content. Bypasses HI-01.

**Fix:** add a source-membership check:

```go
sourceIsMember, _ := s.repo.IsMember(ctx, source.ChannelID, userID)
if !sourceIsMember {
    return nil, apperrors.NewForbidden("cannot forward from a channel you are not a member of")
}
```

---

### HI-03 — `EndCall` allows any logged-in user to terminate any call

**File:** `internal/calls/handler_calls.go:220-232`, `internal/calls/service.go:215-240`

```go
func (h *Handler) EndCall(c fiber.Ctx) error {
    callID, err := parseCallID(c)
    if err := h.service.EndCall(c.Context(), callID); err != nil { ... }
}
```

No check that the caller is a participant, the creator, or a channel admin. Any authenticated user who learns a call UUID can kick everyone out of any meeting. Trivial griefing primitive.

Same issue with **`JoinCall`** (`service.go:167-193`) — no channel-membership check; anyone with a callID and the channelID of the call can join arbitrary calls, generate a LiveKit token, eavesdrop or disrupt.

**Fix:**

- `EndCall`: load call, get channel, verify user is `owner`/`admin` in that channel (or is the call creator).
- `JoinCall`: load call, get channel, verify caller is a channel member (or holds a valid invite). For instant calls (no channel), require either ownership or invite code.

---

### HI-04 — Magic link & Telegram-bot auth bypass via small code keyspace + lax rate-limit

**Files:** `internal/auth/magiclink.go:43-49`, `internal/auth/telegram_bot.go:160-166`

```go
n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
return fmt.Sprintf("%06d", n.Int64()), nil
```

Code space = 10^6 = ~20 bits of entropy. TTL = 5 minutes. Rate-limit on `SendMagicLink` = 3/10min/email (good for sending), but **`VerifyCode` has NO rate-limit at all** in `magiclink.go:99-115`. Combined with the global API rate-limit of 100 req/sec/IP, an attacker hitting `POST /auth/magic-link/verify` (any IP rotation) can try ~30k codes/minute = 150k codes / 5min — that's a 15% chance of hitting any code in the keyspace within the TTL window. Distributed across IPs / via residential proxies, it's a feasible brute force.

The auth rate-limit (`AuthRateLimit{Max:3, Window:10min}`) is by IP — trivially defeated by IP rotation, and disabled entirely when `DevMode=true`.

**Fix:**

1. Increase code length to 8 characters (alphanumeric, ~48 bits) or use a longer signed token.
2. Add a per-email verification rate-limit in `VerifyCode`: after 5 failed attempts on a given email/code, invalidate the code.
3. Track failures by `email` (not just IP). Lock the email out for N minutes on repeated failure.
4. Same applies to `telegram_bot.go:VerifyCode` — same 6-digit keyspace, same lack of attempt limiting.

---

### HI-05 — GitHub & GitLab access tokens / webhook secrets stored in plaintext

**File:** `internal/github/repository.go:78-91`, `internal/gitlab/repository.go` (analogous), `ent/schema/githubintegration.go`

```go
q := r.client.GitHubIntegration.Create().
    SetRepoOwner(input.RepoOwner).
    SetRepoName(input.RepoName).
    SetAccessToken(input.AccessToken).      // plain text
    SetWebhookSecret(input.WebhookSecret).  // plain text
    SetProjectID(projectID)
```

If the database is dumped, leaked via SQL injection, or read by any DBA/SRE, every GitHub/GitLab access token is recoverable — and these tokens often have repo write access. They should be encrypted at rest with a KMS-derived key (envelope encryption) or at minimum a server-held secret separate from the DB credential.

Worth noting that `apigateway.APIKeyService` does the right thing (line 39: `hash := hashKey(fullKey)` — SHA256, never stores the raw key) — so the pattern exists in the codebase.

**Fix:** add a small `internal/platform/crypto/aesgcm.go` helper using AES-256-GCM with a key from `INTEGRATION_ENCRYPTION_KEY` env var, and apply it via a `Hook` on the Ent schema before save (mutators) / after load. Same for the webhook secret. Webhook signatures only need a hash compare so the secret never needs to leave the server.

---

### HI-06 — Race condition in `imex.ImportIssuesCSV` and `ImportIssuesJSON` — non-atomic counter increments

**File:** `internal/imex/service.go:107-119` (CSV), and 187-200 (JSON) — same pattern

```go
proj, _ := s.db.Project.Get(ctx, projectID)
nextNum := proj.TaskCounter + 1
_, err = s.db.Project.UpdateOneID(projectID).
    SetTaskCounter(nextNum).
    Save(ctx)
// ... then
issueKey := fmt.Sprintf("%s-%d", proj.Key, nextNum)
```

Read-modify-write across multiple goroutines (or concurrent imports of the same project, or even one user double-clicking import) will collide:

- Two requests both read `TaskCounter=10`, both set `=11`, both create issue with key `PROJ-11` → unique constraint violation OR duplicate key in case there isn't one.
- The same pattern exists in any tracker `CreateIssue` flow — please audit `internal/tracker/service_issues.go` and `repo_issues.go`.

**Fix:** use a single UPDATE ... RETURNING transaction. With Ent this is awkward; the safer pattern is a Postgres sequence per project, or a `SELECT ... FOR UPDATE` in a transaction. At minimum wrap in a tx and use `UpdateOneID().AddTaskCounter(1).Save(ctx)` (Ent supports atomic `Add*` ops) and read the returned `TaskCounter` from the saved row:

```go
proj, err := tx.Project.UpdateOneID(projectID).AddTaskCounter(1).Save(ctx)
if err != nil { return err }
nextNum := proj.TaskCounter
issueKey := fmt.Sprintf("%s-%d", proj.Key, nextNum)
```

Same race in `JoinByInvite` (`internal/project/invite.go:59-89`) — check-then-increment on `UseCount` is not atomic; two parallel joins both pass `UseCount < MaxUses`, both succeed, exceeds `MaxUses`.

---

### HI-07 — WebSocket Hub overwrites multi-tab sessions and races on Register/Unregister

**File:** `internal/ws/hub.go:63-71`

```go
case client := <-h.register:
    h.clients.Store(client.UserID, client)  // overwrites previous *Client for same user
case client := <-h.unregister:
    h.removeClient(client)
```

Two problems:

1. **Multi-tab is broken**: keyed by `userID`, so Tab-A's `*Client` gets replaced by Tab-B's when B connects. Tab-A is still in some `rooms` sync.Maps but `clients` only points at B. `SendToUser` (line 142) finds only B. When Tab-A disconnects, `removeClient` deletes by `client.UserID` from `clients` — but it may delete B's entry instead of A's, leaking goroutines / breaking messaging to B.
2. **Race on quick reconnect**: register/unregister go through a channel handled by a single goroutine, but `Register()` and `Unregister()` callers don't wait. If the client disconnects and reconnects faster than `Run()` processes both events, the order is preserved by FIFO — but `Subscribe`/`Unsubscribe` modify `h.rooms` directly (not via the channel), bypassing the serialization (`hub.go:84-106`). So a disconnect's `removeClient` racing with a fresh `Subscribe` for the new connection can wipe the new connection's room membership.

**Fix:**

- Key `clients` by `sessionID` (or `userID + connectionID`) instead of `userID`. Maintain a separate `userID → []*Client` map (with mutex) for `SendToUser` to fan out.
- Funnel `Subscribe` and `Unsubscribe` through the same channel/goroutine as register/unregister so the state transitions are serialized.

This also relates to `internal/ws/handler.go:30`: `CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true }` — accepts any origin. JWT in query param is the only auth, which is fine, but a malicious site can still cause the browser to open a WS with the user's token if it ever leaks to the URL bar / referer / history. Enforce `Origin` header allowlist against `CORS_ORIGINS`.

---

### HI-08 — Webhook dispatcher fires DB calls unbounded — goroutine + connection-pool exhaustion

**File:** `internal/apigateway/webhook.go:116-122`, `125-177`

```go
for _, ep := range endpoints {
    if !containsEvent(ep.Events, eventType) { continue }
    go d.deliverWithRetry(ctx, ep, eventType, payload)  // unbounded
}
```

Every event spawns one goroutine per matching endpoint, each does up to 4 HTTP attempts and 4 DB writes (`WebhookDelivery.Create`). No semaphore, no worker pool. A high-volume event source (many tracker writes) + many endpoints = thousands of in-flight goroutines, each holding an Ent client connection. Postgres connection pool is exhausted, server stalls.

Also: each `deliverWithRetry` allocates `body, _ := json.Marshal(payload)` redundantly (already serialized in `handleMessage`).

**Fix:** use a buffered channel + worker pool (e.g. 16 workers) for delivery. Also pass through the already-serialized `msg.Data()` instead of re-marshaling.

---

### HI-09 — Outbox worker can panic and crash the server (no recover, no panic middleware anywhere)

**File:** `internal/platform/outbox/worker.go:56-70`

```go
w.client.OutboxEvent.UpdateOneID(evt.ID).
    SetAttempts(evt.Attempts + 1).
    SetLastError(err.Error()).
    SaveX(ctx)                      // panics on error
// ...
w.client.OutboxEvent.UpdateOneID(evt.ID).
    SetStatus(outboxevent.StatusPublished).
    SetPublishedAt(now).
    SaveX(ctx)                      // panics on error
```

`SaveX` panics if the update fails (e.g. transient DB error, deadlock). The worker goroutine has no `defer recover()`. The panic crashes the goroutine — `ProcessLoop` exits — outbox stops processing, but the rest of the server keeps running. Worse, if Go's runtime hits a panic in a goroutine spawned by `go d.deliverWithRetry(...)` (webhook), the whole process dies because there's no panic middleware.

Same applies to Fiber handlers: `cmd/server/main.go:121-126` instantiates Fiber without a recover middleware. A nil deref in any handler crashes the *request* (Fiber's `Listen` may also stop accepting depending on internal handling — but at minimum you lose the request and log nothing useful).

**Fix:**

```go
// In outbox/worker.go
defer func() {
    if r := recover(); r != nil {
        slog.Error("outbox worker panic", "panic", r, "stack", string(debug.Stack()))
    }
}()
```

And replace `SaveX(ctx)` with `Save(ctx)` + error logging. **And** add `app.Use(recover.New())` in `cmd/server/main.go` after `RequestID()`. Fiber ships a recover middleware in `gofiber/fiber/v3/middleware/recover`.

---

### HI-10 — In-memory rate limiter is per-instance and IP-bypassable via X-Forwarded-For

**File:** `internal/platform/middleware/ratelimit.go:38-53`

```go
key := c.IP()
```

Two problems:

1. `fiber.Ctx.IP()` returns the *remote* socket address by default. Behind a load balancer / reverse proxy, this is the LB's IP, so all traffic shares one rate-limit bucket and everyone gets throttled immediately on any traffic spike. Conversely, if Fiber is configured to trust `X-Forwarded-For`, that header is **trivially spoofable** by an attacker → infinite rate-limit bypass.
2. The limiter state is in-memory (`sync.Map` in the closure). Horizontally scaled deployment = each instance has its own bucket. With N instances, the effective limit is N × Max. The auth rate-limit (3 / 10min / IP) becomes 3N / 10min, which is exactly the brute-force enabler from HI-04.

**Fix:**

- For the global API rate-limit it's OK to keep in-memory but make sure Fiber's `EnableTrustedProxyCheck` is set with a known proxy list.
- For the auth rate-limit, the magic-link rate-limit (which is already Redis-based), and any other sensitive endpoint, move to Redis-based sliding-window — you already have Redis. Re-use the magic-link `Incr` + `Expire` pattern.
- Validate `c.IP()` against `Forwarded` / `X-Real-IP` / `X-Forwarded-For` only when behind a trusted proxy.

---

### HI-11 — `CORS()` middleware: AllowCredentials=true with potentially wildcard origins

**File:** `internal/platform/middleware/cors.go:14-25`

```go
origins := []string{"http://localhost:5173", "http://localhost:8080"}
if env := os.Getenv("CORS_ORIGINS"); env != "" {
    origins = strings.Split(env, ",")
}
return cors.New(cors.Config{
    AllowOrigins:     origins,
    AllowCredentials: true,
})
```

If anyone sets `CORS_ORIGINS=*` (and they will), Fiber will combine `Access-Control-Allow-Origin: *` with `Access-Control-Allow-Credentials: true` — modern browsers refuse this, but the misconfiguration goes silent. Worse, if they set `CORS_ORIGINS=https://attacker.com,https://hub.local`, both are allowed with credentials.

There's also no sanity check / trimming on the split. `CORS_ORIGINS=" http://localhost:5173 , http://localhost:8080"` → with spaces → mismatches strict origin compare.

**Fix:**

```go
origins := []string{...}
if env := os.Getenv("CORS_ORIGINS"); env != "" {
    for _, o := range strings.Split(env, ",") {
        o = strings.TrimSpace(o)
        if o == "" || o == "*" { continue }
        origins = append(origins, o)
    }
}
```

And fail-start if `len(origins) == 0` in production.

---

### HI-12 — `internal/search/SearchUsers` returns all matching users platform-wide with no scoping

**File:** `internal/search/service.go:257-295`

```go
func (s *SearchService) searchUsers(ctx context.Context, query string) ([]SearchResult, error) {
    rows, err := s.sqlDB.QueryContext(ctx,
        `SELECT id, name, email, display_name, ts_rank(...) AS rank
         FROM users WHERE search_vector @@ plainto_tsquery('simple', $1)
         ORDER BY rank DESC LIMIT 10`, query)
```

No filter by team or shared projects. Any authenticated user can enumerate every email/name in the platform via fuzzy queries. This is also a soft GDPR concern.

**Fix:** restrict to users sharing at least one project with the caller, or to the caller's team:

```sql
WHERE search_vector @@ plainto_tsquery('simple', $1)
  AND id IN (
    SELECT DISTINCT pm2.user_project_members
    FROM project_members pm1
    JOIN project_members pm2 ON pm1.project_members = pm2.project_members
    WHERE pm1.user_project_members = $2
  )
```

---

## MEDIUM

### ME-01 — `invalidateUserSessions` runs N+1 queries per session, ignored errors

**File:** `internal/auth/service.go:148-162`

```go
for _, sess := range sessions {
    s.deleteRefreshToken(ctx, sess.RefreshToken)   // Redis: no error check
    if err := s.db.Session.DeleteOneID(sess.ID).Exec(ctx); err != nil {
        slog.Error(...)  // continues anyway
    }
}
```

If `deleteRefreshToken` (which silently discards err — `service.go:140-142`) fails for one session, the refresh token stays in Redis. If `DeleteOneID` fails, the session row stays in Postgres but the JWT no longer matches. State drifts.

Also, the queried list and deletes are not in a transaction. If a new session is created mid-loop (concurrent login), it might or might not be invalidated depending on timing.

**Fix:**

```go
_, err := s.db.Session.Delete().Where(session.HasUserWith(user.ID(userID))).Exec(ctx)
// And pipeline the Redis DELs
```

Single delete + Redis pipeline.

---

### ME-02 — Refresh token rotation has TOCTOU between delete-old and store-new

**File:** `internal/auth/service.go:231-272`

```go
s.deleteRefreshToken(ctx, oldRefreshToken)
// ... generate access token ...
if err := s.storeRefreshToken(ctx, newRefresh, *payload); err != nil { ... }
// ... update Ent session
if _, saveErr := s.db.Session.UpdateOneID(sessionID).
    SetRefreshToken(newRefresh).
```

If the process crashes after `deleteRefreshToken` but before `storeRefreshToken`, the user is permanently locked out — their old token is gone, the new one isn't stored. The Ent session row still has the *old* token, which is now invalid in Redis but valid in DB.

**Fix:** store the new token in Redis *first*, then update Ent, then delete the old token last. This makes the refresh idempotent: a retried request still has the old token in Redis, the second rotation generates a new one, and you only delete after successful store.

Also, no protection against **refresh-token replay**: if an attacker steals an old refresh token, they can use it once and the legitimate user's next refresh will fail with "refresh token not found". That's the desired behavior (refresh-token theft detection), but there's no signal sent to the user — should at least log this case and consider invalidating the entire session.

---

### ME-03 — Logger middleware logs request paths but not status code body / error details, and the access logs include the full path with query string

**File:** `internal/platform/middleware/logging.go:11-30`

Path `c.Path()` includes `?token=<JWT>` for WS upgrades (since the JWT is in a query param — see HI-07 and ws/handler.go:42). Every WS upgrade attempt logs the JWT to slog → stdout → centralized log aggregation. Tokens-in-logs is a recurring source of session compromise.

**Fix:** redact query strings in the logger, or special-case `/ws` to strip query, or — better — move the WS auth to a temporary one-time-token (e.g. `POST /auth/ws-token` returns a short-lived 30-second token used only for the upgrade).

---

### ME-04 — `RegistrationMode` upsert is not atomic; first call may create a duplicate

**File:** `internal/auth/service.go:534-554`

```go
n, err := s.db.SystemSetting.Update().Where(systemsetting.KeyEQ("registration_mode")).SetValue(mode).Save(ctx)
if n == 0 {
    _, err = s.db.SystemSetting.Create().SetKey("registration_mode").SetValue(mode)...Save(ctx)
}
```

Two parallel admins setting the mode could both find `n=0`, both call Create, second one violates unique constraint on `key`. Minor — admins rarely race — but indicative of the pattern; same exists elsewhere.

**Fix:** Postgres `INSERT ... ON CONFLICT DO UPDATE`. Ent supports `UpsertOne()` API.

---

### ME-05 — `magic link rate-limit` Incr+Expire is not atomic

**File:** `internal/auth/magiclink.go:60-69`

```go
count, _ := s.redis.Incr(ctx, rateKey).Result()
if count == 1 { s.redis.Expire(ctx, rateKey, magicLinkRateTTL) }
```

If Redis crashes between Incr and Expire, the counter stays forever (no TTL) — first user gets rate-limited permanently. Also, between Incr returning `count=1` and Expire being applied, another concurrent Incr could increment to 2 with no TTL yet. Bug is small but real.

**Fix:** Use a Lua script (`SET key value EX ttl NX`-like pattern) or `SET key 1 EX ttl NX` for the first hit:

```go
ok, _ := s.redis.SetNX(ctx, rateKey, 1, magicLinkRateTTL).Result()
if !ok {
    count, _ := s.redis.Incr(ctx, rateKey).Result()
    if count > magicLinkRateMax { return "", ErrMagicLinkRateLimited }
}
```

---

### ME-06 — Bridge handlers `_ = json.Unmarshal(msg.Data, &payload)` ignore errors

**File:** `internal/ws/bridge.go:118-122`, repeated in every `handle*Event` function.

Malformed events from NATS get broadcast as `payload: null` to all subscribers. Better to drop the message and log:

```go
var payload any
if err := json.Unmarshal(msg.Data, &payload); err != nil {
    slog.Warn("ws bridge: malformed payload", "subject", msg.Subject, "error", err)
    return
}
```

---

### ME-07 — NATS `EnsureStreams` doesn't include all subjects actually used

**File:** `internal/platform/nats/client.go:57-77`

Configured subjects in stream EVENTS:

```
project.>, tracker.>, messenger.channel.>, tms.>, calls.channel.>, auth.session.>, sprint.>
```

But the codebase also publishes:

- `calls.chat.>` (calls/service.go:341) — NOT in stream → ephemeral only
- `calls.recording.*` (calls/recording.go:118, 153, 177) — NOT in stream
- `notification.user.>` (bridge subscribes to this in ws/bridge.go:80) — NOT in stream → notifications across restart are lost
- `auth.session.created` etc. — in stream (good)

The webhook dispatcher in `apigateway/webhook.go:67` subscribes to **all** subjects matched by the stream (it just consumes everything). So `calls.recording.ready` etc. cannot trigger webhooks even if subscribed.

**Fix:** add `notification.>`, `calls.>` to the stream Subjects list (or specifically `calls.recording.>`, `calls.chat.>`). Decide explicitly per subject whether it's durable (JetStream) or fire-and-forget (core NATS) — current state is incidental.

---

### ME-08 — Webhook secrets are queried in a tight loop for every webhook (no caching)

**File:** `internal/github/handler.go:106-119`, `internal/gitlab/handler.go:84-93`

`ListAllWebhookSecrets` issues a query for every webhook invocation. For high-frequency webhook sources (GitHub spamming push events on big repos) this is unnecessary load. Also: comparing all secrets is O(N) — fine for small N but if there are 1000 integrations, every webhook payload runs 1000 HMAC operations.

**Fix:** map webhook payloads to integrations via the repo identifier (GitHub sends repo owner/name in the body or in `X-GitHub-Hook-Installation-Target-ID`). Then only one HMAC compare is needed. Cache integrations in memory with a short TTL.

---

### ME-09 — Dev seeds and demo channels expose admin-token paths in dev mode

**Files:** `internal/dev/seed_data.go`, `internal/dev/handler.go:Login`

If `DEV_MODE=true` and someone calls `/api/dev/seed`, then `/api/dev/login` with `admin@hub.local`, they get an admin JWT. Re-iterating CR-06 here at MEDIUM as an architectural note: there's nothing preventing dev-mode from being toggled at runtime via env var change + restart. Consider compiling dev endpoints out via build tags (`//go:build dev`) so they cannot be enabled in a release build at all.

---

### ME-10 — Goroutine leak risk: cron tickers in `main.go` capture parent `ctx` (Background), not the cancel-wrapped one

**File:** `cmd/server/main.go:198-211, 358-371, 377-390, 393-408, 451-464, 502-516`

The pattern is consistent and correct:

```go
orphanCtx, orphanCancel := context.WithCancel(ctx)
defer orphanCancel()
go func() {
    ticker := time.NewTicker(24 * time.Hour)
    for { select { case <-ticker.C: ... ; case <-orphanCtx.Done(): return }}
}()
```

This is fine in theory. But `orphanCancel` is `defer`-ed inside `main()`, which only runs when `main` returns — which happens after `app.Shutdown()` blocks on the quit signal. So during shutdown, `orphanCancel` is called late (after the `app.Shutdown()` already drained in-flight HTTP requests). That's acceptable, but you'll have 7-8 background goroutines all calling their cleanup at the same time after shutdown. Verify they handle Postgres connection being closed gracefully — `s.db.Close()` deferred at line 70 happens via LIFO so it runs *after* all the cancels. OK.

The real bug: the inner closures call `orphanCleanupSvc.Cleanup(orphanCtx)` etc., but `Cleanup` may take longer than the shutdown grace period and the parent ctx isn't being cascaded. If the global `ctx` is needed to be canceled on shutdown signal, currently it's not — the SIGINT/SIGTERM handler at line 558-560 just blocks on the channel and calls `outboxCancel()` + `app.Shutdown()` — none of the other crons are signaled.

**Fix:** create one top-level `rootCtx, rootCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)` and pass it to *everything* (Hub, dispatcher, cron tickers, outbox). Drop the individual cancels for crons. One central shutdown path.

---

### ME-11 — `validateWebhookSignature` in calls handler logs API key on mismatch

**File:** `internal/calls/handler.go:236-241`

```go
if iss != h.cfg.APIKey {
    slog.Error("несовпадение API key в webhook", "expected", h.cfg.APIKey, "got", iss)
}
```

Logs the **expected** API key (= LiveKit API key) on mismatch, putting it into log streams that may be ingested by 3rd-party log aggregators. Don't log secrets even on the error path.

**Fix:** log only `got iss` (which is from the attacker's input) and a hash of expected, never the value.

---

### ME-12 — `parseProjectID` and similar helpers don't validate UUID version/non-nil consistently

**Files:** various (e.g. `internal/calls/handler.go:83-93`, `messenger/handler.go:127-137`)

`uuid.Parse("")` returns `uuid.Nil, nil` actually no — it returns an error for empty string. But `uuid.Parse("00000000-0000-0000-0000-000000000000")` returns `uuid.Nil, nil`. None of the handlers reject `uuid.Nil`, so a client sending the zero UUID gets through to DB queries that return "not found" (good) or *match the first row* in some pathological queries. Low risk in practice but worth a defensive check at the parse helper level.

---

### ME-13 — Auth middleware loads user from DB on **every** authenticated request

**File:** `internal/auth/middleware.go:38-43`

```go
uid, parseErr := uuid.Parse(claims.UserID)
if parseErr == nil {
    u, uErr := svc.db.User.Get(c.Context(), uid)
    if uErr == nil && u.Status == user.StatusDeactivated { ... }
}
```

Per-request DB read for `is user deactivated?`. For a high-traffic API this is a 1× overhead per request. Cache the user status in Redis with a short TTL (e.g. 30s) or piggyback on the RBAC cache.

Also: if `uuid.Parse` fails, the deactivation check is silently skipped — auth proceeds. Defensive coding would be to reject the request.

---

### ME-14 — Audit consumer uses background context; on shutdown, in-flight audit writes may be lost

**File:** `cmd/server/main.go:347-353` and `internal/audit/consumer.go` (implied)

```go
auditConsumer := auditmod.NewAuditConsumer(nc, auditRepo)
auditCtx, auditCancel := context.WithCancel(ctx)
defer auditCancel()
if err := auditConsumer.Start(auditCtx); err != nil { ... }
```

Same pattern as ME-10. Audit losses on shutdown are particularly bad because that's the very period when an incident might be happening.

**Fix:** consumer should `defer flush` on shutdown — explicitly drain pending events from JetStream before returning from `Stop()`.

---

### ME-15 — No request-body size limit configured on Fiber

**File:** `cmd/server/main.go:121-126`

```go
app := fiber.New(fiber.Config{
    ReadTimeout: cfg.Server.ReadTimeout,
    WriteTimeout: cfg.Server.WriteTimeout,
    Immutable: cfg.Server.Immutable,
    ErrorHandler: middleware.ErrorHandler,
})
```

Fiber default `BodyLimit` is 4MB which is fine for most endpoints but **CSV import** (`imex.ImportIssuesCSV`) and **avatar upload** could exceed it, and conversely chat / messenger endpoints could happily accept multi-MB JSON. Set explicit per-route limits.

---

## LOW

### LO-01 — Duplicated `escapeMarkdownV2` and `escapeMarkdownV2TG` are identical functions

**Files:** `internal/notification/telegram.go:119-126`, `internal/auth/telegram_bot.go:168-175`

```go
func escapeMarkdownV2(s string) string {
    special := []string{"_", "*", "[", "]", ...}
    ...
}
```

Move to `internal/platform/telegram/markdown.go` (or similar) — single source of truth.

### LO-02 — Duplicated `uuid.Parse(c.Params(...))` boilerplate in every handler

Pattern: `idStr := utils.CopyString(c.Params("xxx")); id, err := uuid.Parse(idStr); if err != nil { return 400 }` — appears in dozens of handlers across `internal/calls`, `internal/messenger`, `internal/tracker`, etc.

**Fix:** add `httputil.UUIDParam(c, "xxx") (uuid.UUID, error)` in `internal/platform/http/helpers.go`. The package already exists (`internal/platform/http/helpers.go`).

### LO-03 — Many service constructors take `*ent.Client` as well as a repository, breaking the repository abstraction

Examples: `messenger.NewMessengerService(repo Repository, ..., db *ent.Client)` (service.go:29), `audit.NewAuditService(db, auditRepo)` (audit/service.go), `search.NewSearchService(db, sqlDB, mongoDB)`. Services bypass their repo abstraction to query Ent directly. Pick one — either pass everything through the repo, or drop the Repository interface and let the service own ent. The hybrid means tests have to mock both, which is why test coverage at the service layer is shallow (see test coverage analysis below).

### LO-04 — Magic numbers hardcoded everywhere

Examples: `maxActiveRecordingsPerCall = 2`, `4096` (max webhook response body), `LIMIT 10` in search SQL, `MaxParticipants: 50` (calls), 5/8/etc minute TTLs. Pull these into named constants or config so they're discoverable and changeable without code surgery.

### LO-05 — `slog.Error` calls regularly include error in the message string and as a slog attribute — redundant

Example: `slog.Error("ошибка отправки telegram сообщения", "error", err, "chat_id", msg.chatID)` — fine. But also: `return fmt.Errorf("ошибка сохранения сообщения: %w", err)` after the slog.Error. Result: error logged twice through the chain. Pick a layer (typically the topmost) to log; lower layers just wrap and return.

### LO-06 — Missing tests on critical paths

Packages with NO `_test.go` files at all (excluding seed/dev tests):

- `internal/platform/database` — no tests for Mongo/Postgres connection setup
- `internal/platform/redis` — no tests
- `internal/platform/config` — no tests (no validation logic to test, but should be added once CR-05 fix lands)
- `internal/gitlab` — only `service_test.go` (which is ~minimal, see below)

Packages with shallow tests (file present but very thin):

- `internal/storage/service_test.go` — 25 lines, just verifies constants
- `internal/notification/service_test.go` — 132 lines, light coverage of Telegram path only
- `internal/dashboard/service_test.go` — 95 lines, no integration
- `internal/imex/service_test.go` — 85 lines, doesn't test the race condition in HI-06
- `internal/sprint/service_test.go`, `internal/tracker/handler` (no handler test file) — handler layer has no tests anywhere except auth & apigateway

**Critical paths with NO test coverage**:

- Webhook dispatcher (`internal/apigateway/webhook.go`) — SSRF, retry, signing
- Storage handler/service end-to-end (presigned URL generation, ConfirmUpload, OrphanCleanup)
- WS Hub register/unregister race (HI-07) and multi-tab handling
- RBAC middleware integration with the actual Fiber routes (CR-04 would have been caught by a single integration test)
- Backup service (`internal/backup/service.go`) — pg_dump / mongorestore are exec'd, but no test verifies error paths
- Outbox worker panic recovery (HI-09)
- Magic link brute force limits (HI-04) — code generation is tested, attempt limiting isn't
- Telegram bot auth code verification & expiry
- LiveKit webhook signature verification — `calls/handler.go:213-255` has zero tests

**Suggested minimal additions** (rough priority):

1. Integration test: spin up Fiber + sqlite (or testcontainers postgres) + miniredis, call every protected route as anonymous → expect 401; as member-of-other-project → expect 403; as project-member → expect 200. This single test would have caught CR-01, CR-04, HI-01, HI-02, HI-03 in one go.
2. Unit test: webhook delivery with mock HTTP server, verify HMAC signature, retry on 5xx, no retry on 4xx, max 3 attempts.
3. Unit test: WS Hub: register two clients for same userID → both should receive messages; unregister one → other should still receive.
4. Unit test: outbox worker survives DB error (SaveX → Save change in HI-09).
5. Race-condition tests with `go test -race` on `imex.Import*` and `project.JoinByInvite`.

### LO-07 — Hub `removeClient` mutates `client.rooms` under lock then iterates outside — minor races on Subscribe during disconnect

**File:** `internal/ws/hub.go:157-181`

```go
client.mu.Lock()
rooms := make([]string, 0, len(client.rooms))
for room := range client.rooms { rooms = append(rooms, room) }
client.rooms = make(map[string]bool)
client.mu.Unlock()
for _, room := range rooms {
    if val, ok := h.rooms.Load(room); ok {
        members := val.(*sync.Map)
        members.Delete(client.UserID)
    }
}
```

If a concurrent `Subscribe(client, room)` runs after `client.mu.Unlock()` but before the `members.Delete`, the client is added back into `h.rooms` but the Send/Broadcast loop has no `done` signal yet. Minor; the `closeOnce.Do` at line 177 handles the cleanup. But this is a code smell. See HI-07 for proposed fix.

### LO-08 — `fiber.Router` accepts both `requireAuth` (auth.RequireAuth) and `authMw fiber.Handler` — naming is inconsistent across packages

Some `RegisterRoutes` take `authMw fiber.Handler` (storage, messenger), others take a named middleware (auth.RequireAuth output). Inconsistent. Fine in practice but harder to grep.

---

## Notable DRY / SOLID Observations

- **DRY**: The handler-layer boilerplate (`parseChannelID`, `parseCallID`, `parseProjectID`, `parseIssueKey`, JSON binding error response, userID extraction) is nearly identical across `internal/messenger`, `internal/calls`, `internal/tracker`. Consolidate into `internal/platform/http`.
- **DRY**: Each module re-implements `toXResponse` converters. With Ent + DTOs this is unavoidable to a point, but the patterns (nil-check, optional pointer copy, etc.) are uniform — a code-gen pass or a small generic helper would help.
- **SOLID (DIP)**: `MessengerService` directly imports `ent.Client`, `ent/project`, `ent/projectmember` for resolving `@role` mentions (`service.go:184-204`). Pushes ent leakage into the service layer; should go through the repo interface (`Repository.ListProjectMembersByRole` etc.).
- **SOLID (SRP)**: `CallService` mixes concerns: LiveKit RPC, DB persistence, NATS publish, chat repository, webhook event handling all in one struct (`internal/calls/service.go`). Splitting `CallService` into `CallRoomService` (LiveKit), `CallRepository`, `CallEventPublisher` would simplify mocking.
- **SOLID (OCP)**: Permission system uses bitmask `Permission(uint64)` — neat. Custom roles store permissions as `int64` in DB. As soon as you exceed 64 permission bits, you're stuck. Today there's no immediate concern, but add a comment / TODO and consider switching to a `[]string` of permission names in the schema for future-proofing.
- The `internal/auth.AuthService` is 580 lines and is rapidly approaching god-object status: JWT, refresh tokens, sessions, magic link, Telegram (3 flavors), profile, onboarding, registration mode, NATS publish. Split into `SessionService`, `LoginService` (magic+telegram), `ProfileService`, `RegistrationModeService`.

---

## Summary by Severity

| Severity | Count | Notable |
|----------|-------|---------|
| CRITICAL | 6 | Backup auth bypass, recording IDOR, webhook SSRF, broken file RBAC, JWT secret, dev mode |
| HIGH     | 12 | Channel read IDOR, forward IDOR, call termination, brute force, plaintext tokens, races, ws hub, etc. |
| MEDIUM   | 15 | Various data integrity / observability / minor races |
| LOW      | 8  | DRY, SOLID, test gaps |

**Total findings:** ~41 (capped at top-impact ones per scope budget). Many small style issues omitted.

---

## Top-5 Fix-First List (impact vs. effort)

1. **CR-01** — add `requireAdmin` middleware to `/admin/backups/*` (5 min change, prevents total data loss).
2. **CR-04** — fix file route registration AND add explicit per-file RBAC. Either way, this is currently broken.
3. **CR-03** — add SSRF protection to webhook dispatcher (one helper file + Transport config; ~30 minutes).
4. **HI-01** — wrap messenger reads with `IsMember` check (a few lines per method, very high impact on data confidentiality).
5. **HI-09** — add Fiber recover middleware + replace `SaveX` with `Save` in outbox (`gofiber/fiber/v3/middleware/recover` is a one-liner).

---

_Reviewed: 2026-05-14_
_Reviewer: Claude (Opus 4.7 1M context)_
_Mode: deep, scoped to top ~25 findings (expanded to 41 due to severity density)_
