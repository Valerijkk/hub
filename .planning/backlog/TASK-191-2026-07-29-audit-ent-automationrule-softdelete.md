# 🔒 Audit ent: AutomationRule soft-delete вместо hard Delete

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`AUDIT-ent-apigateway-2026-07-29.md`](../audits/total-2026-07/AUDIT-ent-apigateway-2026-07-29.md)  
> Исполнитель: **be1** · Finding ID: `EA-001`  
> Approval: migrate SQL `needs-owner-apply` — **BLOCKED waiting owner** (файл в репо, не применён на prod)

## [OUTCOME]

`AutomationRule` имеет soft-delete (`deleted_at`) и runtime Delete не делает hard `Delete()`; list/get фильтруют trash.

## [ROLE]

- Role: be1 · Handoff: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/ent/schema` AutomationRule (или аналог); `backend/internal/tracker` automation delete/list; migrate SQL `needs-owner-apply`; tests
- Do not touch: ослаблять webhook SEC-02 / VersionEQ (TASK-152)
- Optional P3 из того же отчёта (TestSuite OnDelete; soft-delete webhook VersionEQ) — отдельным коммитом или follow-up
- Commits EN/RU; no AI co-author
- Approval: migrate — **needs-owner-apply**

## [ACCEPTANCE]

- Schema `deleted_at` + soft Delete path; list `DeletedAtIsNil`
- Mutation: hard Delete / без filter → RED
- Stop: FIXED · «На проверке» · migrate BLOCKED waiting owner если SQL
