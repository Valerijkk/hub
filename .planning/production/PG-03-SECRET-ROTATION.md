# PG-03 — Secret Rotation Checklist

> Manual ops runbook for **FA-06** (Wave C final audit). Secrets must never be committed; rotate if exposed in git history.  
> Ref: `docs/security_guidelines.md` · `.planning/tasks/TASKS-PHASES-121-FINAL-AUDIT.md` (FA-06)

## When to rotate

- Secret appeared in a commit, PR, CI log, or chat
- Team member with access leaves
- Quarterly rotation (recommended for production)
- **FA-06 gate:** execute before squash PR `feat/v2.5-trust` → `main` (INF-03 Resend/Telegram history)

## Secrets inventory

| Secret | Env var(s) | Location | Rotation steps |
|--------|------------|----------|----------------|
| JWT signing | `JWT_SECRET` | K3s secret / `.env` | Generate 64+ char random; redeploy backend; **all sessions invalidated** |
| Integration token encryption | `INTEGRATION_ENCRYPTION_KEY` | K3s / compose | **Re-encrypt** GitHub/GitLab tokens after rotate; export integrations first |
| Transactional email | `RESEND_API_KEY` | K3s / compose | Regenerate in Resend dashboard; update secret; redeploy |
| Telegram notify | `TELEGRAM_BOT_TOKEN` | K3s / compose | Revoke via @BotFather; new token; update secret (notify-only, not login) |
| PostgreSQL | `POSTGRES_PASSWORD` | compose / K3s | Change in DB + `DATABASE_URL`; rolling restart |
| Redis | `REDIS_PASSWORD` | compose / K3s | Update Redis config + `REDIS_URL`; restart |
| MongoDB | `MONGO_PASSWORD` | compose / K3s | Update user in Mongo + `MONGODB_URL` |
| SeaweedFS S3 | `S3_ACCESS_KEY` / `S3_SECRET_KEY` | compose / K3s | Create new keys in SeaweedFS; update backend; revoke old |
| LiveKit | `LIVEKIT_API_KEY` / `LIVEKIT_API_SECRET` | compose / K3s | Regenerate in LiveKit; update env (from `livekit-prod.yaml.example` template) |
| hub-mail internal | `HUB_MAIL_INTERNAL_TOKEN` | K3s / compose | New shared secret on backend + hub-mail; redeploy both |
| Alertmanager webhook | `ALERT_WEBHOOK_SECRET` | K3s / monitoring | Regenerate; update Alertmanager bearer + backend env |
| GitHub/GitLab OAuth | client secret (provider) | Provider settings | Rotate client secret; update Hub integration env |
| Webhook HMAC | per-integration | DB | Regenerate in project settings UI |
| CI deploy | `SSH_HOST`, `SSH_USER`, `SSH_KEY` | GitHub Actions secrets | Rotate deploy key; update repo secrets |

---

## FA-06 Phase 0 — Pre-rotation git scan

Run from repo root **before** generating new values. Record commit SHAs in the FA-06 sign-off section if any hit.

### Pickaxe search (`git log -S`)

Searches commits that **added or removed** the literal string (use known prefixes from old values when available).

```bash
# Core auth & encryption
git log -S 'JWT_SECRET' --oneline --all
git log -S 'INTEGRATION_ENCRYPTION_KEY' --oneline --all

# Email & Telegram (INF-03 history — priority for FA-06)
git log -S 'RESEND_API_KEY' --oneline --all
git log -S 're_' --oneline --all -- '*.yml' '*.yaml' '*.env*'
git log -S 'TELEGRAM_BOT_TOKEN' --oneline --all

# Databases
git log -S 'POSTGRES_PASSWORD' --oneline --all
git log -S 'REDIS_PASSWORD' --oneline --all
git log -S 'MONGO_PASSWORD' --oneline --all

# Object storage
git log -S 'S3_SECRET_KEY' --oneline --all
git log -S 'S3_ACCESS_KEY' --oneline --all
git log -S 'AWS_SECRET_ACCESS_KEY' --oneline --all

# LiveKit (SEC-03 — livekit-prod.yaml removed; scan history anyway)
git log -S 'LIVEKIT_API_SECRET' --oneline --all
git log -S 'LIVEKIT_API_KEY' --oneline --all
git log -S 'devkey' --oneline --all
git log -S 'livekit-prod.yaml' --oneline --all

# hub-mail & observability
git log -S 'HUB_MAIL_INTERNAL_TOKEN' --oneline --all
git log -S 'ALERT_WEBHOOK_SECRET' --oneline --all
```

### Content grep (working tree + history)

```bash
# Current tree — must return 0 real secrets (placeholders OK)
git grep -E '(re_[a-zA-Z0-9]{20,}|[0-9]{8,}:[A-Za-z0-9_-]{30,})' -- ':!*.md' ':!.planning/*'

# Known-bad filenames (should not exist with secrets)
git log --all --full-history -- '**/livekit-prod.yaml' '**/.env' '**/.env.local'

# CI must not set DEV_MODE in prod deploy
git log -S 'DEV_MODE=true' --oneline -- '.github/workflows/deploy.yml'
```

### If secrets found in history

1. Rotate **all** affected secrets (assume compromise).
2. Document incident in audit log.
3. Consider `git filter-repo` / BFG only with explicit CTO approval (see [Git history purge](#git-history-purge-if-committed)).
4. Do **not** force-push `main` without team coordination.

---

## FA-06 Phase 1 — Generate & patch (K3s)

**Namespace:** `hub` · **Secret name:** `hub-app-secrets` (adjust to your cluster manifest).

Generate values locally (never paste into tickets/chat):

```bash
# JWT, encryption keys, internal tokens (48+ bytes)
openssl rand -base64 48

# Base64-encode for K8s secret data fields
echo -n 'REPLACE_WITH_PLAINTEXT' | base64 -w0   # Linux
# PowerShell: [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes('REPLACE_WITH_PLAINTEXT'))
```

### Single-key JSON patch (template)

Replace `<KEY>`, `<BASE64_VALUE>`, and verify secret name/namespace.

```bash
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/<KEY>", "value": "<BASE64_VALUE>"}
]'
```

### Batch patch examples (no real values)

```bash
# JWT — invalidates all sessions; announce maintenance window
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/JWT_SECRET", "value": "<BASE64_JWT_SECRET>"}
]'

# Integration encryption — run re-encrypt workflow AFTER patch + deploy
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/INTEGRATION_ENCRYPTION_KEY", "value": "<BASE64_INTEGRATION_ENCRYPTION_KEY>"}
]'

# Resend + Telegram (INF-03 priority)
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/RESEND_API_KEY", "value": "<BASE64_RESEND_API_KEY>"},
  {"op": "replace", "path": "/data/TELEGRAM_BOT_TOKEN", "value": "<BASE64_TELEGRAM_BOT_TOKEN>"}
]'

# Database passwords — patch secret THEN update DB user passwords to match
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/POSTGRES_PASSWORD", "value": "<BASE64_POSTGRES_PASSWORD>"},
  {"op": "replace", "path": "/data/MONGO_PASSWORD", "value": "<BASE64_MONGO_PASSWORD>"},
  {"op": "replace", "path": "/data/REDIS_PASSWORD", "value": "<BASE64_REDIS_PASSWORD>"}
]'

# SeaweedFS S3
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/S3_ACCESS_KEY", "value": "<BASE64_S3_ACCESS_KEY>"},
  {"op": "replace", "path": "/data/S3_SECRET_KEY", "value": "<BASE64_S3_SECRET_KEY>"}
]'

# LiveKit
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/LIVEKIT_API_KEY", "value": "<BASE64_LIVEKIT_API_KEY>"},
  {"op": "replace", "path": "/data/LIVEKIT_API_SECRET", "value": "<BASE64_LIVEKIT_API_SECRET>"}
]'

# hub-mail shared token (backend + hub-mail deployment)
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/HUB_MAIL_INTERNAL_TOKEN", "value": "<BASE64_HUB_MAIL_INTERNAL_TOKEN>"}
]'

# Alertmanager webhook bearer
kubectl patch secret hub-app-secrets -n hub --type='json' -p='[
  {"op": "replace", "path": "/data/ALERT_WEBHOOK_SECRET", "value": "<BASE64_ALERT_WEBHOOK_SECRET>"}
]'
```

### Rolling restart (after patch)

```bash
kubectl rollout restart deployment/hub-backend -n hub
kubectl rollout restart deployment/hub-mail -n hub
kubectl rollout status deployment/hub-backend -n hub --timeout=300s
```

**Docker Compose prod:** update `deployments/.env` (never commit), then `docker compose -f deployments/docker-compose.prod.yml up -d --force-recreate backend hub-mail livekit`.

---

## Procedure (production)

1. **Announce** maintenance window if JWT or DB passwords change (users re-login).
2. **Run** [Phase 0 git scan](#fa-06-phase-0--pre-rotation-git-scan); log results.
3. **Generate** new values: `openssl rand -base64 48` (JWT, encryption keys, internal tokens).
4. **Update** K3s secrets or sealed secrets — never commit plaintext.
5. **Deploy** backend with rolling update; verify `/api/health`.
6. **Run** [Post-rotation curl smoke](#fa-06-phase-2--post-rotation-curl-smoke-checklist).
7. **Revoke** old keys at provider (Resend, Telegram, S3, LiveKit).
8. **Re-audit** `git log -S '<old_prefix>'` — if found, consider filter-repo per security guidelines.

---

## FA-06 Phase 2 — Post-rotation curl smoke checklist

Set once per run (no secrets in shell history if possible — use env vars):

```bash
export HUB_BASE="https://yourdomain.com/api"
export HUB_ORIGIN="https://yourdomain.com"
# After magic-link verify:
export HUB_ACCESS_TOKEN="<access_jwt>"
```

| # | Check | Command / action | Expected |
|---|-------|------------------|----------|
| 1 | Health (Postgres, Redis, Mongo) | `curl -sS "$HUB_BASE/health" \| jq .` | `"status":"ok"`, postgres/redis/mongo `"ok"` |
| 2 | Magic Link send | `curl -sS -X POST "$HUB_BASE/auth/magic-link" -H "Content-Type: application/json" -d '{"email":"ops@yourdomain.com"}'` | `200` / no error body (check Resend) |
| 3 | Magic Link verify | `curl -sS -X POST "$HUB_BASE/auth/magic-link/verify" -H "Content-Type: application/json" -d '{"email":"ops@yourdomain.com","code":"123456"}'` | `access_token` + `refresh_token` |
| 4 | Authenticated /me | `curl -sS "$HUB_BASE/auth/me" -H "Authorization: Bearer $HUB_ACCESS_TOKEN"` | `200`, user object |
| 5 | Presigned upload URL | `curl -sS -X POST "$HUB_BASE/files/upload-url" -H "Authorization: Bearer $HUB_ACCESS_TOKEN" -H "Content-Type: application/json" -d '{"filename":"smoke.txt","content_type":"text/plain","size_bytes":12}'` | `upload_url` present (S3 creds OK) |
| 6 | Search (scoped) | `curl -sS "$HUB_BASE/search?q=smoke" -H "Authorization: Bearer $HUB_ACCESS_TOKEN"` | `200`, no 500 (Mongo + scope OK) |
| 7 | Telegram bind status | `curl -sS "$HUB_BASE/auth/telegram" -H "Authorization: Bearer $HUB_ACCESS_TOKEN"` | `200` (bot token valid at Telegram API) |
| 8 | LiveKit token (calls) | Create/join call via UI or `POST /calls/...` with Bearer | Room token issued, no LiveKit auth error |
| 9 | GitHub/GitLab webhook | Trigger test delivery from provider UI | `200` + delivery logged in project |
| 10 | Backup job | Wait for scheduled backup or trigger manual | Completes without S3 auth error |
| 11 | hub-mail alert path | Fire test alert webhook (if monitoring enabled) | Email received; `ALERT_WEBHOOK_SECRET` accepted |
| 12 | Session invalidation (JWT rotate only) | Old refresh token after JWT rotation | `401` on refresh |

**Manual UI smoke (recommended):** create issue → Kanban move → messenger send → voice call join → admin audit log visible.

---

## FA-06 Phase 3 — CTO sign-off

Complete after Phases 0–2. Blocks Wave C closure and squash PR to `main`.

| Item | Owner | Done |
|------|-------|------|
| Phase 0 git scan executed; results attached (or "clean") | devops | [ ] |
| INF-03 priority: `RESEND_API_KEY` + `TELEGRAM_BOT_TOKEN` rotated if any history hit | devops | [ ] |
| All secrets from [inventory](#secrets-inventory) rotated or explicitly N/A documented | devops | [ ] |
| K3s / compose secrets updated; **no plaintext in git** | devops | [ ] |
| `kubectl rollout status` / compose recreate green | devops | [ ] |
| Post-rotation curl smoke (#1–12) all pass | devops + qa | [ ] |
| Old provider keys revoked (Resend, Telegram, S3, LiveKit) | devops | [ ] |
| Incident note in audit log if git exposure confirmed | devops | [ ] |
| **CTO approval for FA-06 / PG-03 complete** | CTO | [ ] |
| **CTO approval for squash PR `feat/v2.5-trust` → `main`** | CTO | [ ] |

**Sign-off record**

| Field | Value |
|-------|-------|
| Executed by | |
| Date (UTC) | |
| Environment | prod / staging |
| Git scan commit range | `git log -1 --format=%H` at scan time: |
| Secrets rotated (list keys only, not values) | |
| Smoke test evidence | link to CI / screenshot / ticket |
| CTO name + date | |

---

## Git history purge (if committed)

```bash
# Example — run only with explicit CTO approval
git filter-repo --path livekit-prod.yaml --invert-paths
# Or BFG for known secret strings
```

Document incident in audit log; notify team.

---

## CI/CD

- GitHub Actions: rotate `SSH_HOST`, `SSH_USER`, `SSH_KEY` repository secrets annually
- Never log env in deploy workflow (deploy.yml uses masked secrets only)
- Confirm `DEV_MODE` is **not** set in production deploy

---

## Verification after rotation (quick checklist)

- [ ] Login (Magic Link; Telegram notify bind only — not login)
- [ ] Create issue + move on Kanban
- [ ] Upload file (presigned PUT)
- [ ] Voice call join
- [ ] GitHub/GitLab webhook delivery
- [ ] Backup job completes (S3 credentials)

*Owner: devops-engineer · Task: FA-06 · Ref: `docs/security_guidelines.md`, IMPROVEMENT-MASTER-PLAN Wave C C-05*

*Last updated: 2026-06-22 — FA-06 manual execution sections*
