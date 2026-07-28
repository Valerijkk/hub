# Hub Backend — Final Security Audit

**Audited:** 2026-05-15
**Auditor:** Claude (Opus 4.7, 1M context)
**Scope:** post-wave-6 state, OWASP Top 10 / ASVS L2
**Method:** static code read of `backend/` against the fix log in
  `.planning/audits/code-review/REVIEW-FIX.md`. No active fixing — report only.

Severity scale is CVSS-like: **CRITICAL** (≥9.0) / **HIGH** (7-9) /
**MEDIUM** (4-7) / **LOW** (<4).

---

## 1. Confirmed CLOSED (from waves 1-6)

The following findings have a present, working mitigation in the
current tree. No re-test was performed beyond reading the relevant files.

| ID         | Area                                | Evidence in repo |
|------------|-------------------------------------|------------------|
| CR-01..06  | admin guard, presign+RBAC, SSRF, /files RBAC, fail-start, DEV warn | per REVIEW-FIX |
| HI-01..04  | channel membership, forward auth, JoinCall RBAC, magic-link lockout | `messenger/service_messages.go`, `auth/magiclink.go` |
| HI-05      | token at-rest encryption (AES-GCM)  | `platform/crypto/aesgcm.go`, `github/repository.go` |
| HI-06..12  | atomic counters, WS multi-tab, bounded webhook pool, Redis rate-limit, CORS, scoped user search | per REVIEW-FIX |
| ME-01..15  | bulk delete, query log redaction, Incr+EX race, JSON unmarshal checks, body limit, refresh rotation, etc. | per REVIEW-FIX |
| LO-01..08  | dedup + ParseUUIDParam + const cleanup | per REVIEW-FIX |

---

## 2. NEW findings (post-wave-6)

### NF-01 — `APIKeyAuth` middleware never executes [CRITICAL, ~9.1]

**File:** `backend/cmd/server/main.go:554`
`app.Use(apigatewaymod.APIKeyAuth(apiKeySvc))` is registered AFTER every
route group is mounted. In Fiber v3 (as in v2), `app.Use` only applies
to handlers registered *after* it. Result: the X-API-Key header is
silently ignored on every endpoint that was meant to accept it. Any
caller relying on API-key auth instead hits the JWT-only middleware and
fails — *or worse, the route requires no auth and the key is moot*.
Net effect: API-key auth is dead code; documented capability is absent.

**Fix:** move `app.Use(APIKeyAuth(...))` above `api := app.Group("/api")`
or attach it to the `/api` group before any handler mounting.

### NF-02 — WebSocket has no origin check; rooms have no AuthZ [CRITICAL, ~9.0]

**Files:** `backend/internal/ws/handler.go:29-31`, `ws/hub.go:142`.
`CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true }` — any
origin can upgrade as long as it presents a JWT (CSWSH risk for any
state-changing message via WS). Worse: `Hub.Subscribe(client, room)` is
called from `readPump` based purely on the client-supplied room string;
there is no membership/RBAC check on the room name. An authenticated
user can subscribe to `project:<any-uuid>`, `channel:<any-uuid>`,
`admin:<...>` and receive every event the server fans out to that room.
This is broadcast-layer IDOR.

**Fix:** (a) validate origin against CORS whitelist before upgrade; (b)
gate `Subscribe` by parsing the room prefix and verifying the user has
membership/permission for the underlying entity.

### NF-03 — Mongo full-text search leaks every channel's messages [HIGH, ~8.1]

**File:** `backend/internal/search/service.go:535-587` (`searchMessages`).
Filter is `{"$text": {"$search": query}}` with no channel/project
restriction. Any authenticated user can `GET /api/search?q=...&types=message`
and get text snippets from channels they are not a member of. Bypasses
the HI-01 channel-membership rework on the messenger side.

**Fix:** intersect with the caller's `channel_members` set (Mongo
`$in: [channelIDs]`), mirroring how `searchIssues` already filters by
`projectIDs`.

### NF-04 — LiveKit webhook: body-hash check is optional [HIGH, ~8.0]

**File:** `backend/internal/calls/handler.go:256-265`.
```go
expectedHash, _ := claims["sha256"].(string)
if expectedHash != "" {  // <-- skipped if claim absent
    ...
}
return true
```
A forged JWT signed with the LiveKit API secret but **without** a
`sha256` claim passes the body verification trivially. Combined with
NF-05 (the LiveKit "secret" is `secret`), an attacker who reaches the
webhook endpoint can mint arbitrary call/recording events.

**Fix:** require `expectedHash != ""` and reject otherwise.

### NF-05 — `livekit-prod.yaml` ships `devkey: secret` [CRITICAL, ~9.4]

**File:** `deployments/livekit-prod.yaml:8-11`.
Despite the `-prod` filename:
```yaml
keys:
  devkey: secret
webhook:
  api_key: devkey
```
This is the LiveKit-side trust anchor for room JWTs and the webhook
signing key. Anyone who reads the repo can sign LiveKit tokens for the
production cluster and call NF-04. `config.validate()` only checks the
backend's `LIVEKIT_API_KEY/SECRET`; it does not read this YAML.

**Fix:** template the YAML with `${LIVEKIT_API_KEY}/${LIVEKIT_API_SECRET}`
the way `docker-compose.prod.yml` already does (line 122-125), or
generate it at deploy time.

### NF-06 — `gitlab/service.go` uses `http.DefaultClient` for user URLs [HIGH, ~7.6]

**File:** `backend/internal/gitlab/service.go:265, 367, 453, 508`.
`integration.GitlabURL` is per-tenant configurable. The four
`http.DefaultClient.Do(req)` sites have no SSRF filter, no timeout
beyond Go's defaults, and no redirect cap. Same vector that CR-03
closed for `apigateway/webhook.go` — but the gitlab integration was
not migrated.

**Fix:** replace with `safehttp.NewSafeHTTPClient(10*time.Second)` and
call `safehttp.ValidateOutboundURL(integration.GitlabURL)` at
integration-create time.

### NF-07 — WebPush endpoint is a SSRF sink [HIGH, ~7.4]

**Files:** `backend/internal/notification/push.go:135-148`,
`notification/handler.go:295-305`.
`req.Endpoint` is taken from the user, stored, and later passed to
`webpush.SendNotification` which performs an HTTP POST. No validation
against `safehttp.ValidateOutboundURL`. A user can subscribe with
`http://169.254.169.254/...` or `http://internal-host/...` and trigger
an outbound request from the server every time a notification fires.

**Fix:** call `ValidateOutboundURL` in the registration handler; ideally
also pin the legitimate WebPush gateway hosts (fcm.googleapis.com,
updates.push.services.mozilla.com, etc.).

### NF-08 — No HTTP security headers (HSTS, CSP, X-Frame-Options) [HIGH, ~7.0]

**File:** `backend/cmd/server/main.go:179-187`.
The middleware stack is RequestID → recover → Logger → CORS →
RateLimit. There is no Helmet-style middleware adding HSTS, CSP,
X-Frame-Options, X-Content-Type-Options, Referrer-Policy or
Permissions-Policy. Combined with NF-09 (frontend served from the same
origin) this leaves the SPA exposed to clickjacking and stylesheet/JS
hijack.

**Fix:** add a small middleware that sets at minimum:
`Strict-Transport-Security`, `X-Content-Type-Options: nosniff`,
`X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`,
and a CSP appropriate for the SPA.

### NF-09 — Filename is concatenated into S3 key without sanitization [MEDIUM, ~5.5]

**File:** `backend/internal/storage/service.go:70`.
`s3Key := fmt.Sprintf("%s/files/%s/%s", projectID, fileID, filename)`.
`filename` is user-controlled. Containing `../` or NUL or
extremely long values currently goes through to the S3 driver. AWS SDK
v2 URL-encodes the key on PUT so directory escape is unlikely in
SeaweedFS, but downstream consumers (the recording exporter, the
backup ZIP, future log paths) might `filepath.Join` it.
Also no MIME validation server-side; `contentType` is whatever the
client claims. No size enforcement on the presigned PUT (S3 presign
without `Content-Length` constraint accepts arbitrarily-large uploads —
the 25 MB Fiber `BodyLimit` does not apply to the direct-to-S3 path).
No antivirus / sandbox scan before `ConfirmUpload`.

**Fix:** strip path separators / control chars from `filename`; enforce
an allow-list of `contentType`; add a `Content-Length` constraint to
the presign; (defense-in-depth) hook a ClamAV scan into
`ConfirmUpload` for binary content types.

### NF-10 — Magic-link 6-digit code, weak entropy [MEDIUM, ~5.0]

**File:** `backend/internal/auth/magiclink.go:46-52`.
10⁶ codespace. HI-04 enforces 5 fails / email / TTL — but attacker
controls neither the salt nor the rotation. With `magicLinkVerifyMax=5`
and 5-minute TTL, an attacker who can trigger a real send (by knowing
the email) has a 0.001% per-attempt chance, ~1 in 200,000 sessions
guessed — meaningful at internet scale. The per-IP cap is via the
Redis sliding-window limiter on `/auth`, but IP rotation defeats it.

**Fix:** widen to 8 digits (10⁸ codespace, ~1e-7 guess probability) or
switch to a 128-bit URL token sent in the magic link instead of a
typed code.

### NF-11 — No reverse-proxy / trusted-proxy configuration [MEDIUM, ~4.5]

**Files:** `backend/cmd/server/main.go` (fiber.Config),
`backend/internal/platform/middleware/ratelimit.go:244-253`.
Fiber's `EnableTrustedProxyCheck` / `TrustedProxies` are not set.
Behind a reverse proxy (Caddy, Nginx), `c.IP()` returns the proxy's
address, collapsing per-client rate limits into a single bucket. The
in-line comment in `ratelimit.go` already flags this; it has not been
wired.

**Fix:** add `EnableTrustedProxyCheck: true`, `TrustedProxies: [...]`,
`ProxyHeader: "X-Forwarded-For"` to `fiber.Config` and document the
expected deployment topology.

### NF-12 — Refresh tokens stored as plaintext in Postgres [MEDIUM, ~4.5]

**Files:** `backend/ent/schema/session.go` (RefreshToken field),
`auth/service_session.go:158`. The refresh-token value lives in
Postgres unhashed. If the DB is dumped (backup S3 leak, SQL injection
elsewhere, support engineer access) every active session can be
hijacked until `refreshTTL` elapses. Redis copy is also plaintext.

**Fix:** hash refresh tokens with SHA-256 (no need for slow KDF —
they are 128-bit random) and store only the hash; lookup-by-hash in
Refresh.

### NF-13 — WebSocket message size 4 KiB / no per-conn write rate [LOW, ~3.8]

**File:** `backend/internal/ws/handler.go:25`. `maxMessageSize = 4096`
on read. There is no per-connection write rate limit or send-buffer
cap; a malicious client that subscribes to a chatty room and then
stops reading can OOM the server (writePump blocks on the channel,
hub publishers continue to `select` into `default` which drops, so
this is partially mitigated — but the drop drops *for everyone* in
the room broadcast, not just the slow client).

**Fix:** when `client.send` is full, close that client only.

### NF-14 — No CSRF token for cookie-less, same-origin flow [INFORMATIONAL]

The API is consumed via `Authorization: Bearer` JWT; no auth cookie is
issued. As long as the SPA never depends on cookies for auth, CSRF is
moot. If a future change introduces auth cookies (refresh in cookie,
session cookie), Fiber's `csrf` middleware is required. Flagging only
for monitoring; not currently exploitable.

---

## 3. Residual gaps from REVIEW-FIX (informational)

These are flagged here for completeness; they are documented as
"partial" or "skipped" in `REVIEW-FIX.md`.

- **HI-05 backfill** — existing GitHub/GitLab tokens stored before the
  AES-GCM rollout are still plaintext until they happen to be re-saved.
- **HI-07 follow-through** done in wave 5; only LOW residual: stale
  room entries on close vs. send-to-default — not exploitable.
- **`opengraph.Fetch`** — link-preview library bypasses SafeDialer;
  `ValidateOutboundURL` is the only gate. Replace with an in-house
  OG parser or fork the library.
- **ME-09 Dockerfile** — dev endpoints excluded from binary requires
  `make build-prod`; Dockerfile must pass `-tags release` to be safe in
  CI/CD.

---

## 4. Top-5 next-step recommendations

1. **NF-01 + NF-05 + NF-04 — close the LiveKit webhook trust chain in
   one PR.** Move `APIKeyAuth` above `app.Group`, template
   `livekit-prod.yaml`, require the `sha256` claim. These three
   together are an unauthenticated RCE-equivalent against the calls
   subsystem.

2. **NF-02 — WebSocket Origin + room authorization.** Add `CheckOrigin`
   matching the CORS list and gate `Subscribe(channel:<uuid>)` /
   `Subscribe(project:<uuid>)` through `IsMember` / `GetEffectivePermissions`.
   This closes the broadcast-layer IDOR that survived HI-01.

3. **NF-03 + NF-06 + NF-07 — finish the SSRF and search-scoping pass.**
   Migrate `gitlab/service.go` and `notification/push.go` to
   `safehttp.NewSafeHTTPClient`; add the channel-IDs filter to
   `searchMessages`. Mechanical changes, low risk.

4. **NF-08 — security headers middleware.** One small middleware
   (HSTS, CSP, X-Frame-Options, X-Content-Type-Options,
   Referrer-Policy, Permissions-Policy). Cheap, large ASVS payoff.

5. **NF-12 — hash refresh tokens at rest.** SHA-256 hash on store,
   lookup-by-hash on refresh. Cuts the blast radius of any future DB
   leak and aligns with the AES-GCM work already done for integration
   secrets.

Honourable mentions (do next quarter):
- NF-09 storage hardening (filename sanitization + MIME allow-list +
  presign size constraint).
- NF-10 magic-link entropy upgrade.
- NF-11 trusted-proxy wiring + per-deployment IP-source documentation.

---

## 5. ASVS L2 verdict

After waves 1-6, the codebase meets most of the L2 controls for
auth, session, crypto-at-rest, input validation, error handling and
logging. **Three L2 controls fail** in the current tree:

- **V4.3 (Authorization) — fail** due to NF-02 (WS rooms).
- **V12.6 (SSRF) — partial** due to NF-06, NF-07.
- **V14.4 (Security headers) — fail** due to NF-08.

NF-01 and NF-05 are configuration/deployment defects with crit-level
blast radius and should block production cutover regardless of ASVS
target. The rest are addressable in one focused 2-3 day pass.
