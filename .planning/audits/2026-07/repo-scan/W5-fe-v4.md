# W5-fe-v4 — Frontend v4 modules

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23
> **Scope:** bases/crm/hr/intranet/calendar/accounting/marketplace UI · **Код не менялся**

## Summary

v4 pages существуют; Vitest smoke слабый (W6). CRM FE наследует BE CRITICAL IDOR. Bases UI после owner-scope BE — OK путь.

**Вердикт:** `verified: false`.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **CRITICAL** | CRM UI + BE (W3) | Любой member через UI правит чужие contact/deal | BE authorize; FE hide mutate без роли |
| **HIGH** | Vitest v4 smoke | heading + `toHaveBeenCalled` без error/RBAC | empty/error/deny cases |
| **MEDIUM** | accounting UI | Admin-only BE — FE должен гейтить route | Redirect non-admin |
| **PASS** | bases | Owner-scope BE + API | — |
| **INFO** | calendar leave | Org-wide leave в UI (W3 HIGH) | Согласовать privacy |
| **LOW** | marketplace | Owner checks BE OK | — |
