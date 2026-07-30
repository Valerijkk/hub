# 🔒 Audit BE v4: calendar trash filter + draft poll IDOR

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: [`AUDIT-BE-v4-crm-hr-intranet-2026-07-29.md`](../audits/total-2026-07/AUDIT-BE-v4-crm-hr-intranet-2026-07-29.md)  
> Исполнитель: **be1** · Finding IDs: `BE-V4-001`…`BE-V4-003`  
> Mutation: [`MUTATION-2026-07-30-W13-be2-task189.md`](../audits/total-2026-07/MUTATION-2026-07-30-W13-be2-task189.md) · 3/3 PROVEN

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

## [MUTATION CRITERIA] — verifier-ready

> W13 slot 3 refill / be2 · criteria only (no mutate) · 2026-07-30 · HEAD guards verified

| ID | Production file:line (guard) | слой | Exact test RED | Kill recipe |
|----|------------------------------|------|----------------|-------------|
| BE-V4-001 | `backend/internal/calendar/repository.go:118-126` — `ListApprovedLeaveInRange`: `leaverequest.DeletedAtIsNil()` | EntRepository | `TestListApprovedLeaveInRange_ExcludesSoftDeleted_Task189` · `backend/testframework/calendar/leave_sprint_trash_task189_test.go` | убрать `leaverequest.DeletedAtIsNil()` из Where → soft-deleted leave в результате → NotEqual/Len assert RED |
| BE-V4-002 | `calendar/repository.go:78-85` — `ListSprintsInRange`: `sprintEnt.DeletedAtIsNil()` | EntRepository | `TestListSprintsInRange_ExcludesSoftDeleted_Task189` · `leave_sprint_trash_task189_test.go` | убрать `sprintEnt.DeletedAtIsNil()` → soft-deleted sprint в /events path → Len/NotEqual RED |
| BE-V4-003 | `backend/internal/intranet/service_polls.go:22-24` + `:49` + `requirePublishedNewsOrAuthorOrAdmin:216-223` — Get/Vote draft → `ErrNotFound` для non-author | service | `TestDraftNewsPoll_GetVoteHiddenFromNonOwner_Task189` · `backend/testframework/intranet/draft_poll_idor_task189_test.go` | убрать вызовы `requirePublishedNewsOrAuthorOrAdmin` (или всегда `return nil`) → member Get/Vote draft OK → ErrNotFound assert RED |

### Layer notes (mock vs Ent)

| ID | Тест бьёт | Flag |
|----|-----------|------|
| 001 | Ent `LeaveRequest` + `calendar.NewEntRepository` | **OK** — Ent |
| 002 | Ent `Sprint` + EntRepository | **OK** — Ent |
| 003 | Ent `NewsPost`/`NewsPoll` + intranet.Service | **OK** — Ent + service authz |

`layer_mismatch IDs: (none)` · `mock≠Ent IDs: (none)` · `docs-only IDs: (none)`

### Gaps

- Docs-only: **N/A** — все 3 ID runtime code.
- Do not touch: coffee/partners TASK-163; CRM P3 version mapping.

## [MUTATION EXECUTION — be2 2026-07-30 W13 slot 3]

| ID | Verdict | RED |
|----|---------|-----|
| BE-V4-001 | PROVEN | `TestListApprovedLeaveInRange_ExcludesSoftDeleted_Task189` |
| BE-V4-002 | PROVEN | `TestListSprintsInRange_ExcludesSoftDeleted_Task189` |
| BE-V4-003 | PROVEN | `TestDraftNewsPoll_GetVoteHiddenFromNonOwner_Task189` |

- BREAK: убрать `DeletedAtIsNil` leave → RED; sprint → RED; `requirePublishedNewsOrAuthorOrAdmin` always nil → RED; restore → GREEN all.
- leftover MUTATED: **нет** · `git diff` calendar/repository.go + intranet/service_polls.go empty.
- Статус задачи: **Закрыта** (0 NOT_PROVEN).
