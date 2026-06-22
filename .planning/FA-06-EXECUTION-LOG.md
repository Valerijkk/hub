# FA-06 — PG-03 Secret Rotation Execution Log

> **Blank template** for manual execution during Wave C final audit. Fill during rotation; **never** paste secret values here.  
> Runbook: `.planning/PG-03-SECRET-ROTATION.md` · Ref: `docs/security_guidelines.md` · Task: FA-06

---

## Run metadata

| Field | Value |
|-------|-------|
| Executed by | AI-assisted scan (Cursor) + ops TBD |
| Date started (UTC) | 2026-06-22 |
| Date completed (UTC) | Phase 0 only |
| Environment | prod / staging |
| Maintenance window announced? | N |
| Repo HEAD at scan time (`git log -1 --format=%H`) | `a9f931c65ae47d274c7dfb88c76045945874bd4` |
| Cluster / compose target | K3s namespace `hub` / Docker Compose prod |

---

## Phase 0 — Pre-flight git scan

Run all commands from **repo root** before generating new secret values. Record results below; attach commit SHAs if any command returns hits.

### Pickaxe search (`git log -S`)

| # | Command | Result (clean / N commits) | Commit SHAs (if any) |
|---|---------|----------------------------|----------------------|
| 1 | `git log -S 'JWT_SECRET' --oneline --all` | 12 commits (env/docs refs) | multiple — env var name only |
| 2 | `git log -S 'INTEGRATION_ENCRYPTION_KEY' --oneline --all` | 6 commits | Wave 6 prod hardening |
| 3 | `git log -S 'RESEND_API_KEY' --oneline --all` | 13 commits | **INF-03** — plaintext before `859ecef` |
| 4 | `git log -S 're_' --oneline --all -- '*.yml' '*.yaml' '*.env*'` | hits in compose history | removed in `859ecef` |
| 5 | `git log -S 'TELEGRAM_BOT_TOKEN' --oneline --all` | 13 commits | **INF-03** — plaintext before `859ecef` |
| 6 | `git log -S 'POSTGRES_PASSWORD' --oneline --all` | 11 commits | compose examples |
| 7 | `git log -S 'REDIS_PASSWORD' --oneline --all` | env refs | |
| 8 | `git log -S 'MONGO_PASSWORD' --oneline --all` | env refs | |
| 9 | `git log -S 'S3_SECRET_KEY' --oneline --all` | env refs | |
| 10 | `git log -S 'S3_ACCESS_KEY' --oneline --all` | env refs | |
| 11 | `git log -S 'AWS_SECRET_ACCESS_KEY' --oneline --all` | clean / minimal | |
| 12 | `git log -S 'LIVEKIT_API_SECRET' --oneline --all` | 15 commits | **SEC-03** livekit-prod.yaml history |
| 13 | `git log -S 'LIVEKIT_API_KEY' --oneline --all` | history hits | livekit-prod.yaml removed from tree |
| 14 | `git log -S 'devkey' --oneline --all` | 17 commits | dev template + removed prod yaml |
| 15 | `git log -S 'livekit-prod.yaml' --oneline --all` | 4 commits | file removed; history remains |
| 16 | `git log -S 'HUB_MAIL_INTERNAL_TOKEN' --oneline --all` | Wave 6+ | |
| 17 | `git log -S 'ALERT_WEBHOOK_SECRET' --oneline --all` | observability wave | |

### Content grep (working tree + history)

| # | Command | Result (clean / N hits) | Notes |
|---|---------|-------------------------|-------|
| 18 | `git grep -E '(re_[a-zA-Z0-9]{20,}|[0-9]{8,}:[A-Za-z0-9_-]{30,})' -- ':!*.md' ':!.planning/*'` | 18 hits | **test mocks only** (`testframework/notification`, `auth`) |
| 19 | `git log --all --full-history -- '**/livekit-prod.yaml' '**/.env' '**/.env.local'` | 4+ commits | `livekit-prod.yaml` in history; not in working tree |
| 20 | `git log -S 'DEV_MODE=true' --oneline -- '.github/workflows/deploy.yml'` | 2 commits | fixed `2b3ae80` — no DEV_MODE in prod deploy |

### Pre-flight checklist

| Item | Done |
|------|------|
| All Phase 0 commands executed | [x] |
| Results recorded in tables above | [x] |
| If any history hit: incident note opened in audit log | [ ] ops |
| If any history hit: assume compromise — rotate **all** affected secrets | [x] P1 Resend+TG |
| If filter-repo / BFG needed: explicit CTO approval obtained | [ ] not requested |
| INF-03 priority confirmed (`RESEND_API_KEY`, `TELEGRAM_BOT_TOKEN`) | [x] |

**Phase 0 summary:** exposure found (describe without secret values):

```
Working tree: clean (test mock tokens only).
History (INF-03): RESEND_API_KEY + TELEGRAM_BOT_TOKEN in docker-compose.local.yml before commit 859ecef.
History (SEC-03): livekit-prod.yaml with devkey/secret in git history (file removed from tree).
P1 rotate before prod: RESEND_API_KEY, TELEGRAM_BOT_TOKEN.
P2 after prod cutover: JWT_SECRET, LIVEKIT keys, INTEGRATION_ENCRYPTION_KEY (re-encrypt).
Phase 1–3: manual ops — see PG-03-SECRET-ROTATION.md § Phase 1–3.
```

---

## Phase 1 — Secret rotation execution

Update K3s secret `hub-app-secrets` (namespace `hub`) or `deployments/.env` (never commit). Mark **Rotated** when new value is applied; **Verified** when post-deploy check confirms the service uses the new secret.

| Secret | Env var(s) | Rotated (Y/N) | Verified (Y/N) | Notes |
|--------|------------|---------------|----------------|-------|
| JWT signing | `JWT_SECRET` | | | All sessions invalidated — announce maintenance |
| Integration token encryption | `INTEGRATION_ENCRYPTION_KEY` | | | Re-encrypt GitHub/GitLab tokens after rotate |
| Transactional email | `RESEND_API_KEY` | | | INF-03 priority; revoke old key in Resend |
| Telegram notify | `TELEGRAM_BOT_TOKEN` | | | INF-03 priority; revoke via @BotFather |
| PostgreSQL | `POSTGRES_PASSWORD` | | | Update DB user + `DATABASE_URL`; rolling restart |
| Redis | `REDIS_PASSWORD` | | | Update Redis config + `REDIS_URL` |
| MongoDB | `MONGO_PASSWORD` | | | Update Mongo user + `MONGODB_URL` |
| SeaweedFS S3 access | `S3_ACCESS_KEY` | | | Create new keys; revoke old in SeaweedFS |
| SeaweedFS S3 secret | `S3_SECRET_KEY` | | | |
| LiveKit API key | `LIVEKIT_API_KEY` | | | From `livekit-prod.yaml.example` template |
| LiveKit API secret | `LIVEKIT_API_SECRET` | | | |
| hub-mail internal | `HUB_MAIL_INTERNAL_TOKEN` | | | Backend + hub-mail both redeployed |
| Alertmanager webhook | `ALERT_WEBHOOK_SECRET` | | | Update Alertmanager bearer + backend env |
| GitHub OAuth client secret | (provider) | | | N/A if not configured |
| GitLab OAuth client secret | (provider) | | | N/A if not configured |
| Webhook HMAC (per-integration) | (DB / UI) | | | Regenerate in project settings UI |
| CI deploy SSH | `SSH_HOST`, `SSH_USER`, `SSH_KEY` | | | GitHub Actions repo secrets |

### Deploy & rollout

| Step | Done | Notes |
|------|------|-------|
| K3s secrets patched (no plaintext in git) | [ ] | |
| `kubectl rollout restart deployment/hub-backend -n hub` | [ ] | |
| `kubectl rollout restart deployment/hub-mail -n hub` | [ ] | |
| `kubectl rollout status deployment/hub-backend -n hub --timeout=300s` | [ ] | |
| **Or** compose: `deployments/.env` updated + force-recreate | [ ] | |
| Old provider keys revoked (Resend, Telegram, S3, LiveKit) | [ ] | |
| Post-rotation re-scan: `git log -S '<old_prefix>'` clean | [ ] | |

---

## Phase 2 — Post-rotation smoke results

Set once per run (use env vars; avoid secrets in shell history):

```bash
export HUB_BASE="https://yourdomain.com/api"
export HUB_ORIGIN="https://yourdomain.com"
export HUB_ACCESS_TOKEN="<access_jwt>"   # after magic-link verify
```

| # | Check | Pass (Y/N) | HTTP / outcome | Notes |
|---|-------|------------|----------------|-------|
| 1 | Health (Postgres, Redis, Mongo) — `GET /health` | | | Expected: `"status":"ok"` |
| 2 | Magic Link send — `POST /auth/magic-link` | | | Resend key OK |
| 3 | Magic Link verify — `POST /auth/magic-link/verify` | | | `access_token` + `refresh_token` |
| 4 | Authenticated `/me` — `GET /auth/me` | | | `200`, user object |
| 5 | Presigned upload URL — `POST /files/upload-url` | | | `upload_url` present (S3 OK) |
| 6 | Search (scoped) — `GET /search?q=smoke` | | | `200`, no 500 |
| 7 | Telegram bind status — `GET /auth/telegram` | | | Bot token valid at Telegram API |
| 8 | LiveKit token (calls) — UI or `POST /calls/...` | | | Room token issued |
| 9 | GitHub/GitLab webhook — test delivery from provider | | | `200` + delivery logged |
| 10 | Backup job — scheduled or manual trigger | | | No S3 auth error |
| 11 | hub-mail alert path — test alert webhook | | | Email received; bearer accepted |
| 12 | Session invalidation (JWT rotate only) — old refresh token | | | Expected: `401` on refresh |

### Manual UI smoke (recommended)

| Action | Pass (Y/N) | Notes |
|--------|------------|-------|
| Create issue → Kanban move | | |
| Messenger send | | |
| Voice call join | | |
| Admin audit log visible | | |

**Smoke evidence link** (CI run / screenshot / ticket — no secrets):

```
```

---

## Phase 3 — CTO sign-off

Complete after Phases 0–2. Blocks Wave C closure and squash PR `feat/v2.5-trust` → `main`.

### Gate checklist

| Item | Owner | Done |
|------|-------|------|
| Phase 0 git scan executed; results attached (or "clean") | devops | [ ] |
| INF-03 priority: `RESEND_API_KEY` + `TELEGRAM_BOT_TOKEN` rotated if any history hit | devops | [ ] |
| All secrets from inventory rotated or explicitly N/A documented | devops | [ ] |
| K3s / compose secrets updated; **no plaintext in git** | devops | [ ] |
| `kubectl rollout status` / compose recreate green | devops | [ ] |
| Post-rotation curl smoke (#1–12) all pass | devops + qa | [ ] |
| Old provider keys revoked (Resend, Telegram, S3, LiveKit) | devops | [ ] |
| Incident note in audit log if git exposure confirmed | devops | [ ] |
| **CTO approval for FA-06 / PG-03 complete** | CTO | [ ] |
| **CTO approval for squash PR `feat/v2.5-trust` → `main`** | CTO | [ ] |

### Sign-off record

| Field | Value |
|-------|-------|
| Executed by | |
| Date (UTC) | |
| Environment | prod / staging |
| Git scan commit range / HEAD at scan | |
| Secrets rotated (keys only, not values) | |
| Secrets marked N/A (with reason) | |
| Smoke test evidence | |
| **CTO name** | |
| **CTO signature / approval date (UTC)** | |
| **FA-06 / PG-03 approved** | Y / N |
| **Squash PR approved** | Y / N |

### CTO comments

```
```

---

*Template version: 2026-06-22 · Owner: devops-engineer · Task: FA-06 · Runbook: `.planning/PG-03-SECRET-ROTATION.md`*
