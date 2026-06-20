# PG-03 — Secret Rotation Checklist

> Manual ops runbook. Secrets must never be committed; rotate if exposed in git history.

## When to rotate

- Secret appeared in a commit, PR, CI log, or chat
- Team member with access leaves
- Quarterly rotation (recommended for production)

## Secrets inventory

| Secret | Location | Rotation steps |
|--------|----------|----------------|
| `JWT_SECRET` | K3s secret / `.env` | Generate 64+ char random; redeploy backend; all sessions invalidated |
| `INTEGRATION_ENCRYPTION_KEY` | K3s / compose | **Re-encrypt** GitHub/GitLab tokens after rotate; export integrations first |
| `RESEND_API_KEY` | K3s / compose | Regenerate in Resend dashboard; update secret; redeploy |
| `TELEGRAM_BOT_TOKEN` | K3s / compose | Revoke via @BotFather; new token; update secret |
| `POSTGRES_PASSWORD` | compose / K3s | Change in DB + connection strings; rolling restart |
| `REDIS_PASSWORD` | compose / K3s | Update Redis config + app env; restart |
| `MONGO_PASSWORD` | compose / K3s | Update user in Mongo + connection URI |
| `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` | SeaweedFS S3 | Create new IAM-style keys in SeaweedFS; update backend; revoke old |
| `LIVEKIT_API_KEY` / `LIVEKIT_API_SECRET` | LiveKit config | Regenerate in LiveKit; update `livekit-prod.yaml` **from template** |
| GitHub/GitLab OAuth app secrets | Provider settings | Rotate client secret; update Hub integration env |
| Webhook HMAC secrets | Per-integration DB | Regenerate in project settings UI |

## Procedure (production)

1. **Announce** maintenance window if JWT or DB passwords change (users re-login).
2. **Generate** new values: `openssl rand -base64 48` (JWT, encryption keys).
3. **Update** K3s secrets or sealed secrets — never commit plaintext.
4. **Deploy** backend with rolling update; verify `/health`.
5. **Verify** magic link, Telegram login, file upload, calls, webhooks smoke.
6. **Revoke** old keys at provider (Resend, Telegram, S3, LiveKit).
7. **Audit** `git log -S '<old_prefix>'` — if found, consider `git filter-repo` + force-push policy per security guidelines.

## Git history purge (if committed)

```bash
# Example — run only with explicit approval
git filter-repo --path livekit-prod.yaml --invert-paths
# Or BFG for known secret strings
```

Document incident in audit log; notify team.

## CI/CD

- GitHub Actions: rotate `DEPLOY_*` repository secrets annually
- Never log env in deploy workflow (deploy.yml uses masked secrets only)
- Confirm `DEV_MODE` is **not** set in production deploy

## Verification after rotation

- [ ] Login (Magic Link + Telegram)
- [ ] Create issue + move on Kanban
- [ ] Upload file (presigned PUT)
- [ ] Voice call join
- [ ] GitHub/GitLab webhook delivery
- [ ] Backup job completes (S3 credentials)

*Owner: devops-engineer | Ref: docs/security_guidelines.md, IMPROVEMENT-MASTER-PLAN Wave C C-05*
