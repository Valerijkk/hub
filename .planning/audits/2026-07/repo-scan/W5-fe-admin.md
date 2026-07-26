# W5-fe-admin — Frontend admin

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** fe-ver (readonly) · **Дата:** 2026-07-23
> **Scope:** `modules/admin`, `api/admin*.ts`, backup UI · **Код не менялся**

## Summary

Admin UI бьёт в `/api/admin/*` (BE requireAdmin PASS). Риск — FE не скрывает пункты для non-admin до 403; backup destructive без второго подтверждения.

**Вердикт:** `verified: false` · **historical_self_check:** PASS (file scope) с оговорками MEDIUM UX (BE gate держит).

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | BE audit/backup | Admin middleware | — |
| **MEDIUM** | admin nav | Пункты видны non-admin → cascade 403 | Гейт по `system_role` |
| **MEDIUM** | backup restore UI | Destructive без confirm/type-name | Confirm modal + audit note |
| **LOW** | GDPR delete UI | Нужен double-confirm | — |
| **INFO** | E2E admin-* | 3 specs (ai/audit/metrics) | — |
| **PASS** | `api/admin.ts` / backup | Typed | — |
