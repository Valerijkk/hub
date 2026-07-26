# W4-fe-auth-shell — Auth + App shell

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** auth routes, `auth-store`, `api/client.ts`, app-shell · **Код не менялся**

## Summary

Magic-link UI only; `ApiClient` single-flight refresh (FE-H1). Pending-user с JWT (W1 HIGH) — FE shell пускает в app. `!important` только в `app.css` prefers-reduced-motion (не ui/).

**Вердикт:** `verified: false` из-за pending session + BE auth HIGH.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **HIGH** | BE pending→session (W1) | Shell/`_authenticated` пускает pending как active | BE block; FE: status gate → waiting-approval |
| **MEDIUM** | `api/client.ts:33+` | Refresh single-flight — OK; access JWT 15m after logout | Accept TTL **или** broadcast logout |
| **PASS** | auth UI | Нет password/SSO forms | Держать до ADR-AUTH-SSO |
| **PASS** | `ApiError.request_id` | Клиент понимает request_id | — |
| **LOW** | `app.css:1359-1362` | `!important` в reduced-motion — a11y exception | Не добавлять в `components/ui` |
| **INFO** | Telegram | Нет login UI | PASS ADR-AUTH-01 surface |
