# ADR-TZ-REWRITE-2026 — Product Scope Expansion

> **Status:** Accepted · **Date:** 2026-06-21 · **Deciders:** PO  
> **Supersedes:** parts of `.cursor/rules/60-tz-constraints.mdc` FORBIDDEN list

## Context

PO directed **MAX scope** within self-hosted Hub evolution. Original TZ forbade Gantt, webcams, email marketing, no-code DB, built-in accounting, etc.

## Decision

### Remains OUT

- SSO / LDAP / OAuth login (Magic Link only — ADR-AUTH-01)
- Multi-tenant multi-org on one instance until **v5**
- Native App Store / Play apps (**PWA sufficient**)
- SaaS billing / subscriptions until **Later** milestone

### Now IN scope (by phase)

| Feature | Phase |
|---------|-------|
| Gantt charts | 16 (defer v4 if overload) |
| Webcam in calls | 16 |
| Call recording | audio+screen (unchanged) |
| Email transactional + marketing | v2.5 alerts; v2.6 Mailchimp-lite via hub-mail |
| Full JQL in automation | 17 |
| Meilisearch global search | v2.6+ |
| AI multi-provider + BYOK | 17 |
| Intranet / Calendar / CRM / HR | v4 |
| Full built-in accounting | v4 (no 1C connector required) |
| Lark-style no-code database | v4 |
| Public status page | v4 |
| Triage inbox | 16 |

### Organization model (v2.5)

- `organization` entity; 1 org per instance (v5: multi-org)
- Jira Cloud sites pattern: User → Org → Projects
- `system_role founder` (one per org)

## Consequences

- Update `docs/vision_and_scope.md`, `60-tz-constraints.mdc`, FEATURE-BACKLOG
- Security review required for webcam, marketing email, accounting modules
- Solo-dev timeline: planner sequences phases; no fixed deadline Phase 15

## Compliance

Code patterns and security non-negotiables (SSRF, WS RBAC, etc.) **unchanged**.
