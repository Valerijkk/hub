# Hub — Technical Constraints (tracked mirror)

> **Canonical for agents (local):** `.cursor/rules/60-tz-constraints.mdc`  
> **Updated:** 2026-06-21 · **ADR:** `.planning/decisions/ADR-TZ-REWRITE-2026.md`

## MANDATORY

### Authentication
- **Magic Link (email) only** — no passwords, no SSO (ADR-AUTH-01)
- Telegram: **notification bind/delivery only**, not login
- JWT access: 15 min · Refresh: Redis 30 days · Auth rate limit: 3/10 min

### Organization (v2.5+)
- `organization` table; 1 org per instance (multi-org → v5)
- Registration via **invite**; admin **Pending queue**; 7d auto-delete
- `system_role founder` — one per org

### Data / API / Real-time / Storage / Frontend / Infra
- Unchanged from v2.0.2 (UTC, UUID v4, soft delete, Fiber WS, NATS, SeaweedFS, TanStack Router, Shadcn, Caddy, K3s prod)

## IN SCOPE (by phase — was forbidden)

| Feature | Phase |
|---------|-------|
| Gantt | 16 |
| Webcam calls | 16 |
| Email alerts + marketing (hub-mail) | 15.2 / 16 |
| Full JQL automation | 17 |
| AI BYOK + Ollama | 17 |
| Intranet, Calendar, CRM, HR | v4 |
| Full accounting (no 1C) | v4 |
| No-code DB (Lark-style) | v4 |

## STILL FORBIDDEN

- SSO / LDAP / OAuth providers
- Multi-tenant multi-org (until v5)
- Native App Store / Play (PWA OK)
- SaaS billing (until Later)
- Code anti-patterns: `http.DefaultClient`, `io.ReadAll`, `*ent.Client` in handlers, etc.
- Security: webhook skip, WS without RBAC, search without user scope

## Email

- Magic link (auth)
- Transactional alerts (opt-in) — ADR-NOTIF-01
- Marketing Mailchimp-lite (admin-only, v2.6) — ADR-TZ-REWRITE-2026
