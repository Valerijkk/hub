# 🔒 Audit ent: AutomationRule soft-delete вместо hard Delete

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`AUDIT-ent-apigateway-2026-07-29.md`](../audits/total-2026-07/AUDIT-ent-apigateway-2026-07-29.md)  
> Исполнитель: **be1** · Finding ID: `EA-001`  
> Mutation: [`MUTATION-2026-07-30-W13-be2-task191.md`](../audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task191.md) · 3/3 PROVEN · migrate SQL **N/A**  
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

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| EA-001-001 | `backend/internal/tracker/repo_automation.go:98-106` — `DeleteAutomationRule`: `Update().…SetDeletedAt` (не `.Delete()`) | EntRepository | `TestEntRepository_AutomationRuleSoftDelete_TASK191` · `backend/testframework/tracker/automation_soft_delete_task191_test.go` | заменить soft Update/`SetDeletedAt` на hard `AutomationRule.Delete()` → `client.Get` NotFound / source `.Delete().` assert RED |
| EA-001-002 | `repo_automation.go:15-19` (+ Find `:60`, ListEnabled `:122`) — `DeletedAtIsNil()` на list/get/engine | EntRepository | `TestEntRepository_AutomationRuleSoftDelete_TASK191` (+ Count: `TestEntRepository_AutomationRuleCountIgnoresTrash_TASK191`) · `automation_soft_delete_task191_test.go` | убрать `DeletedAtIsNil` из `ListAutomationRules` (и/или Find) → soft-deleted в list / Find OK → NotEqual/NotFound assert RED |
| EA-001-003 | `backend/ent/schema/automationrule.go:44-46` (+ indexes `:78-79`) — `field.Time("deleted_at")` | schema | `TestSchema_TASK191_AutomationRuleDeletedAt` · `automation_soft_delete_task191_test.go` (dup `backend/testframework/ent/schema_automationrule_task191_test.go`) | убрать `deleted_at` field/indexes из schema → Contains assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Ent soft Delete + raw Get row + source guard | **OK** — Ent |
| 002 | Ent List/Find/Count после soft Delete | **OK** — Ent |
| 003 | schema source Contains | **OK** — static schema (не Ent runtime DDL) |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A** — code/schema.
- Migrate SQL `backend/migrations/20260729_task191_automationrule_soft_delete.sql` (`needs-owner-apply` / BLOCKED waiting owner) — **N/A runtime mutate** (owner-apply gate; `TestMigrate_TASK191_*` = file marker only).
- Do not touch: webhook SEC-02 / VersionEQ (TASK-152).
