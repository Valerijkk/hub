# 🔒 Audit BE v4: calendar trash filter + draft poll IDOR

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`AUDIT-BE-v4-crm-hr-intranet-2026-07-29.md`](../audits/total-2026-07/AUDIT-BE-v4-crm-hr-intranet-2026-07-29.md)  
> Исполнитель: **be1** · Finding IDs: `BE-V4-001`…`BE-V4-003`

## [OUTCOME]

Calendar leave/sprint queries фильтруют soft-deleted; draft poll GET/vote недоступны non-author без published (нет IDOR draft).

## [ROLE]

- Role: be1 · Handoff: be-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `backend/internal/calendar/**`, `backend/internal/intranet/**` polls; tests calendar/intranet
- Do not touch: coffee/partners TASK-163; CRM P3 version mapping (optional separate)
- Commits EN/RU; no AI co-author

## [ACCEPTANCE]

- 001/002: `DeletedAtIsNil` на leave/sprint calendar sources; mutation RED
- 003: draft poll GET/vote → 404/403 for non-owner; mutation RED
- Stop: 3/3 FIXED · «На проверке»
