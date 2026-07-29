# 🔒 C-W19: soft-delete project гасит webhook endpoints

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
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
