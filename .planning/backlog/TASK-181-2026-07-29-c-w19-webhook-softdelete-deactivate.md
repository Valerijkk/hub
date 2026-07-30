# 🔒 C-W19: soft-delete project гасит webhook endpoints

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`C-W19-fe-ai-platform.md`](../audits/total-2026-07/C-W19-fe-ai-platform.md) §C-W19-004  
> Исполнитель: **be2** · Finding ID: `C-W19-FE-004`

## [OUTCOME]

При soft-delete project webhook endpoints деактивируются (или эквивалент) в той же tx; list-by-project не отдаёт endpoints корзины без `project.DeletedAtIsNil`.

## [ROLE]

- Role: be2 · Handoff: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/internal/project/repository.go` SoftDelete*; `backend/internal/apigateway/repository.go` list-by-project filter; tests apigateway/project
- Do not touch: ослаблять ListActive `HasProjectWith(DeletedAtIsNil)` (TASK-152)
- Commits EN/RU; no AI co-author

## [REFERENCES]

- ListActive уже фильтрует soft project; SoftDeleteWithCrossLinkCleanup не трогает endpoints
- ListWebhookEndpointsByProject без DeletedAtIsNil на project
- Audit: C-W19-fe-ai-platform.md

## [ACCEPTANCE]

- Soft-delete project → endpoints inactive / не в active dispatch; list-by-project пуст для trash
- Mutation: убрать deactivate/filter → RED
- Stop: FIXED · «На проверке»

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 1 refill / be1 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified  
> Finding alias: audit `C-W19-004` ≡ task `C-W19-FE-004`

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| C-W19-FE-004-001 | `backend/internal/project/repository.go:691-700` — `SoftDeleteWithCrossLinkCleanup` tx: `WebhookEndpoint.Update().…SetIsActive(false)` before project `DeletedAt` | EntRepository | `TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181`; `TestSoftDeleteProject_WebhookDeactivate_SourceGuard_TASK181` · `backend/testframework/apigateway/task181_c_w19_004_test.go` | убрать `WebhookEndpoint.Update` / `SetIsActive(false)` из SoftDelete tx → `IsActive` остаётся true / source Contains RED |
| C-W19-FE-004-002 | `backend/internal/apigateway/repository.go:221-231` — `ListWebhookEndpointsByProject`: `HasProjectWith(…, project.DeletedAtIsNil())` | EntRepository | `TestListWebhookEndpointsByProject_SoftDeletedProject_Empty_TASK181`; `TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181` · `task181_c_w19_004_test.go` | убрать `project.DeletedAtIsNil()` из Where → list-by-project trash непуст → Empty-assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Ent SoftDelete tx + Ent WebhookEndpoint row (`IsActive`) + ListActive | **OK** — Ent |
| 002 | Ent ListWebhookEndpointsByProject + project `DeletedAt` fixture | **OK** — Ent |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A** — оба ID runtime Ent.
- **Do not touch** (TASK-152 lock): `ListActiveWebhookEndpoints` `HasProjectWith(project.DeletedAtIsNil())` (`apigateway/repository.go:170-177`) — source-guard asserts remain; не ослаблять в kill 001/002.

## [MUTATION RESULT] — 2026-07-30 W13 slot 1 refill (be1)

| ID | Verdict | RED |
|----|---------|-----|
| C-W19-FE-004-001 | **PROVEN** | `TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181`; `TestSoftDeleteProject_WebhookDeactivate_SourceGuard_TASK181` |
| C-W19-FE-004-002 | **PROVEN** | `TestListWebhookEndpointsByProject_SoftDeletedProject_Empty_TASK181`; `TestSoftDeleteProject_DeactivatesWebhookEndpoints_TASK181` |

- Report: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W13-be1-task181.md`
- Status → **Закрыта** (2/2 PROVEN).
