# W3-v4-intranet-calendar — intranet / calendar / accounting / marketplace

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** security-auditor (readonly) · **Дата:** 2026-07-23
> **Scope:** IDOR/RBAC spot-check · **Код не менялся**

## Summary

**Accounting / intranet admin / marketplace mutate** — `requireAdmin` / owner checks на месте. **Calendar ListEvents** scoped по `ListUserProjectIDs`, но **approved leave** тянется org-wide. CRM/bases — см. `W3-v4-bases-crm-hr` (crm CRITICAL).

**Вердикт:** `verified: false` — 1 HIGH (calendar leave privacy) + зависимость от crm CRITICAL в соседнем отчёте.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `intranet/service.go:62-187` | News CRUD admin: `requireAdmin` / `IsSystemAdmin` | — |
| **PASS** | `accounting/service.go:33-161` + invoice | Все операции `requireAdmin` | — |
| **PASS** | `marketplace/service.go:107-316` | Update/Delete/status: owner **или** admin | — |
| **PASS** | `calendar/service.go:37-47` | Events: calls/sprints/issues через `projectIDs` membership | — |
| **HIGH** | `calendar/service.go:77-83` + repo `ListApprovedLeaveInRange` | Approved leave **без** user/org-privacy filter — любой JWT видит все отпуска org за range | Scope: self + team peers **или** HR-role; иначе ADR «org-visible leave» |
| **MEDIUM** | `intranet` kudos public lists | Public news/kudos — by design; проверить PII в DTO | Минимизировать поля |
| **INFO** | marketplace Create | Listing create без admin — OK (moderation queue `ListPending`) | — |
| **INFO** | CRM IDOR | См. W3-v4-bases-crm-hr CRITICAL | Не дублировать |

## Checklist

| Модуль | AuthZ mutate | Scope read |
|--------|--------------|------------|
| intranet news | PASS admin | PASS |
| accounting | PASS admin | PASS |
| marketplace | PASS owner/admin | PASS |
| calendar projects | PASS member projects | **FAIL leave org-wide** |
