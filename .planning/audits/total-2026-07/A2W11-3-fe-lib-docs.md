# A2 W11 агент 3 — fe lib+docs part 3 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: `W11-ASSIGNMENTS.md` A3 · Stage A: `A11-3-fe-lib-docs.md`  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/stores/call-store.ts` | 221 | 0 | FIXED-OK error + no reset on fail |
| `frontend/src/stores/command-store.ts` | 11 | 0 | чисто |
| `frontend/src/stores/dashboard-store.ts` | 134 | 0 | FIXED-OK project/admin/layout error |
| `frontend/src/stores/messenger-store.ts` | 136 | 0 | чисто |
| `frontend/src/stores/sidebar-store.ts` | 25 | 0 | чисто |
| `frontend/src/stores/sprint-store.ts` | 126 | 0 | FIXED-OK `error` на fetch |
| `frontend/src/stores/timer-store.ts` | 66 | 0 | FIXED-OK `Number.isFinite` |
| `frontend/src/stores/tms-store.ts` | 30 | 0 | FIXED-OK filters-only |
| `frontend/src/stores/ws-store.ts` | 160 | 1 | RESIDUAL JWT в query |
| `frontend/src/hooks/use-auth.ts` | 77 | 0 | FIXED-OK cancelled + token guard |
| `frontend/src/hooks/use-channels.ts` | 115 | 0 | чисто |
| `frontend/src/hooks/use-composer-ai-suggest.ts` | — | 0 | FIXED-OK удалён → `lib/composer-ai-suggest.ts` |
| `frontend/src/hooks/use-debounced-value.ts` | 10 | 0 | чисто |
| `frontend/src/hooks/use-file-upload.ts` | 149 | 0 | FIXED-OK dragenter + zone paste |
| `frontend/src/hooks/use-invalidate.ts` | 27 | 0 | чисто |
| `frontend/src/hooks/use-messages.ts` | 216 | 0 | FIXED-OK payload object-guard |
| `frontend/src/hooks/use-notifications.ts` | 284 | 0 | FIXED-OK id/title guard |
| `frontend/src/hooks/use-presence.ts` | 60 | 2 | RESIDUAL + NEW |
| `frontend/src/hooks/use-typing.ts` | 61 | 0 | FIXED-OK payload guard |

Файлов в списке: **19**. Прочитано: **19** (composer — подтверждено отсутствие файла; канон в `lib/`). Пропущено: **0**.

## Находки (открытые)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A2W11-3-001 | RESIDUAL | P2 | `use-presence.ts:55-59` | Возврат только `statuses` / `isLoading` / `getStatus` — **нет** `isError` | W11-A3-012 остался в TODO TASK-047 при статусе «Исправлена» (ложное закрытие). Fail poll → UI «все offline» | Прокинуть `isError`/`error`/`refetch`; channel-list — «статус недоступен» |
| A2W11-3-002 | RESIDUAL | P2 | `ws-store.ts:101-102` | Access JWT в `/ws?token=` query string | W11-A3-013 / P2-FE-U02: токен в access/proxy logs. FE-only не закрывает | BE: subprotocol / first-message auth; FE перейти на новый handshake |
| A2W11-3-003 | NEW | P3 | `use-presence.ts:41-44` | `sendHeartbeat().catch(() => {})` — silent swallow | Сеть/403 heartbeat невидимы; смежно с отсутствием isError на poll | Логировать/счётчик fail; не маскировать длительный fail |

## FIXED-OK (Stage A → закрыто после B/C)

| Stage A ID | Evidence |
|------------|----------|
| W11-A3-001 | `use-auth`: `cancelled` + `accessToken === tokenAtStart` |
| W11-A3-002 | `checkActiveCall` set error + **rethrow** (не null) |
| W11-A3-003 | dashboard `fetchProject`/`fetchAdmin` пишут `error` |
| W11-A3-004 | `leaveCall`/`endCall`: fail → error + **return** без `initialState` |
| W11-A3-005 | sprint-store `error: "Не удалось загрузить спринты"` |
| W11-A3-006 | dragCounter только `onDragEnter` |
| W11-A3-007 | paste на drop-zone props, не `document` |
| W11-A3-008 | `toggleHand` пишет `error` |
| W11-A3-009 | `saveLayoutDebounced` → `error` flag |
| W11-A3-010 | use-messages object/array guard |
| W11-A3-011 | use-typing тот же guard |
| W11-A3-014 | sprint `error` (Query migration — не блокер после error flags) |
| W11-A3-015 | notification.new: `id`+`title` string guard |
| W11-A3-016 | tms-store = filters only |
| W11-A3-017 | helpers в `lib/composer-ai-suggest.ts`; hooks-файл удалён |
| W11-A3-018 | timer `Number.isFinite` → stop |

## Чисто

command/messenger/sidebar stores — UI-only; use-channels/invalidate/debounced — Query keys + cleanup OK.

## Сводка агента

| Метрика | Значение |
|---------|----------|
| Файлов | 19/19/0 |
| Открытых находок | **3** |
| NEW | 1 (P3) |
| REGRESS | 0 |
| RESIDUAL | 2 (P2×2) |
| FIXED-OK | 16 |
