# A2 W11 агент 2 — fe lib+docs part 2 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W11-ASSIGNMENTS.md` A2 · Stage A: `A11-2-fe-lib-docs.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/api/github.ts` | 140 | 0 | FIXED-OK link-preview дубль убран |
| `frontend/src/api/gitlab.ts` | 121 | 0 | чисто |
| `frontend/src/api/hr.ts` | 82 | 0 | чисто |
| `frontend/src/api/intranet.ts` | 185 | 0 | FIXED-OK `updateNews` |
| `frontend/src/api/issue-message-links.ts` | 57 | 0 | чисто |
| `frontend/src/api/marketplace.ts` | 136 | 0 | FIXED-OK явные `<T>` |
| `frontend/src/api/messenger.ts` | 533 | 0 | FIXED-OK `image_url` + WS guard |
| `frontend/src/api/notifications.ts` | 160 | 0 | чисто |
| `frontend/src/api/projects.ts` | 129 | 0 | FIXED-OK `updateMemberRole` |
| `frontend/src/api/search.ts` | 103 | 0 | чисто |
| `frontend/src/api/sprint.ts` | 170 | 0 | чисто |
| `frontend/src/api/status.ts` | 18 | 0 | чисто |
| `frontend/src/api/tms.ts` | 838 | 0 | FIXED-OK create/deleteEnvironment |
| `frontend/src/api/tracker.ts` | 673 | 0 | чисто (cursor list) |
| `frontend/src/api/types.ts` | 19 | 0 | FIXED-OK dual pagination comment |
| `frontend/src/api/webhooks.ts` | 52 | 0 | чисто |
| `frontend/src/stores/ai-store.ts` | 104 | 0 | FIXED-OK loaded/modelsError/rollback |
| `frontend/src/stores/auth-store.ts` | 154 | 1 | RESIDUAL JWT в sessionStorage |
| `frontend/src/stores/board-store.ts` | 49 | 0 | чисто |

Файлов в списке: **19**. Прочитано: **19**. Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A2W11-2-001 | RESIDUAL | P2 | `auth-store.ts:161` (+ `:13-15`) | Access + refresh JWT по-прежнему в `sessionStorage` (`hub-auth`) | W11-A2-005 / P2-FE-U01: XSS → кража refresh (30d). Фикс B не трогал persist (нужен ADR cookie/memory) | httpOnly refresh **или** memory-only access + не persist refresh в `partialize` |

## FIXED-OK (Stage A → закрыто после B/C)

| Stage A ID | Evidence |
|------------|----------|
| A2W11-2-002 | TASK-104 · `ADR-PAGINATION-CURSOR-ISSUES-2026` · TZ/rule-60/NFR-14 = cursor · `TestListIssuesCursorPagination_Task104` + HTTP |
| W11-A2-001 | `LinkPreviewData.image_url`/`domain`; consumers на каноне |
| W11-A2-002 | `ai-store`: `loaded` только на success; `botLoadError` |
| W11-A2-003 | `modelsError` + `saveSelectedModel` rollback |
| W11-A2-004 | `clearSessionLocal` без broadcast |
| W11-A2-006 | `tms.createEnvironment` / `deleteEnvironment` |
| W11-A2-007 | `intranet.updateNews` |
| W11-A2-008 | `projects.updateMemberRole` |
| W11-A2-009 | `types.ts` комментарий dual pagination (остаток — TZ, см. RESIDUAL) |
| W11-A2-010 | `buildMessageFromWsEvent(unknown)` + required id/sender_id |
| W11-A2-011 | `setUser` broadcast с `userId`; setTokens → null до me |
| W11-A2-012 | `broadcastAuth` зеркалит localStorage event |
| W11-A2-013 | `github.ts` без дубля `/link-preview` |
| W11-A2-014 | marketplace `api.get`/`post` с явными generics |

## Чисто

gitlab/hr/issue-links/notifications/search/sprint/status/webhooks/board-store — typed wrappers, version где нужно, без async leaks в board-store.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 19/19/0 |
| Открытых находок | **1** |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 1 (P2×1) |
| FIXED-OK | 14 |
