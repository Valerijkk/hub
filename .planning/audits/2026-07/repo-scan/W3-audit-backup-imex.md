# W3-audit-backup-imex — audit / backup / imex admin scope

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** security-auditor (readonly) · **Дата:** 2026-07-23  
> **Код не менялся**

## Summary

**Audit** и **backup** смонтированы под `/api/admin/*` + `requireAdmin` / `adminMw` — **PASS**. **Imex** — только `authMw`, **без** `RequirePermission` / membership (тот же класс, что W2-project CRITICAL): любой JWT может import/export по чужому `projectId`.

**Вердикт:** `verified: false` — **1 CRITICAL** (imex IDOR).

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `audit/handler.go:24-36` | Group `/admin` + `requireAdmin()` на audit/trash/ban/GDPR | — |
| **PASS** | `audit/repository.go:510+` | `IsSystemAdmin` для admin gate | — |
| **PASS** | `backup/handler.go:21-23` | `/admin/backups` + `authMw, adminMw` | — |
| **CRITICAL** | `imex/handler.go:26-35` · `wiring_integrations.go:119-121` | Import/export issues/jira/ZIP/PDF — только JWT. Нет RBAC/`IsMember`. Любой auth → bulk read/write чужого проекта | `RequirePermission(PermProjectEdit/View/…)` + membership; denial-suite |
| **HIGH** | `imex` ExportRunPDF / ExportSprintPDF | Path `:projectId` не сверяется с run/sprint ownership (если service не фильтрует) | Authorize project + entity.ProjectID |
| **MEDIUM** | `imex/handler.go:69` | `ReadAll(LimitReader)` — OK лимит 25MB | — |
| **INFO** | audit Ban/GDPR | Admin-only destructive — OK; audit log обязателен | — |

## Checklist

| Модуль | Admin/RBAC | Статус |
|--------|------------|--------|
| audit | requireAdmin | PASS |
| backup | adminMw | PASS |
| imex | auth only | **FAIL CRITICAL** |
