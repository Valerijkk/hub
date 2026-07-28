# A3 W11 агент 3 — fe lib+docs part 3 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W11-3-fe-lib-docs.md](A2W11-3-fe-lib-docs.md)  
> Stage A: [A11-3-fe-lib-docs.md](A11-3-fe-lib-docs.md)  
> Периметр: 19 файлов (W11-ASSIGNMENTS A3) · Прочитано: 19 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с A2 (A2W11-3-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W11-3-001 | RESIDUAL P2 | **FIXED-OK** | `use-presence.ts:71-78` — наружу `isError`/`error`/`refetch`; `getStatus` учитывает `query.isError` (`:60-68`); consumer `channel-list.tsx:674,772` → `presenceUnavailable`; тест `use-presence-error.test.tsx`; TASK-093 |
| A2W11-3-002 | FIXED (TASK-103) | **FIXED-OK** | `ws-store.ts:105-110` — `openSocket(ticket)` → `/ws` + `hub.ws.ticket.*` subprotocol; **нет** `?token=`; `fetchWSTicket` POST `/api/auth/ws-ticket` (`:87-103`); ADR-WS-JWT-HANDSHAKE |
| A2W11-3-003 | NEW P3 | **FIXED-OK** | `use-presence.ts:46-53` — `heartbeatFailCount` + `console.warn("[presence] heartbeat failed", …)`; не silent `catch {}`; regression в `use-presence-error.test.tsx` |

**REGRESS:** 0 · **RESIDUAL (A2 actives):** 0 · **NEW (этот слот):** 2

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/stores/call-store.ts` | 221 | 0 | FIXED-OK error + leave/end no reset on fail + checkActiveCall rethrow |
| `frontend/src/stores/command-store.ts` | 13 | 0 | чисто (UI open/toggle) |
| `frontend/src/stores/dashboard-store.ts` | 145 | 1 | fetch*/saveLayout error OK; **loadLayout** silent → A3W11-3-001 |
| `frontend/src/stores/messenger-store.ts` | 136 | 0 | чисто (UI + unread + typing mirror) |
| `frontend/src/stores/sidebar-store.ts` | 27 | 0 | чисто (persist collapsed) |
| `frontend/src/stores/sprint-store.ts` | 140 | 0 | FIXED-OK `error` на fetch/burndown/velocity |
| `frontend/src/stores/timer-store.ts` | 72 | 0 | FIXED-OK `Number.isFinite` → stopTimer |
| `frontend/src/stores/tms-store.ts` | 34 | 0 | FIXED-OK filters-only |
| `frontend/src/stores/ws-store.ts` | 230 | 0 | FIXED-OK ticket subprotocol + reconnect backoff |
| `frontend/src/hooks/use-auth.ts` | 82 | 0 | FIXED-OK cancelled + `accessToken === tokenAtStart` |
| `frontend/src/hooks/use-channels.ts` | 126 | 0 | чисто (Query SoT + WS guards + cleanup) |
| `frontend/src/hooks/use-composer-ai-suggest.ts` | — | 0 | FIXED-OK удалён → `lib/composer-ai-suggest.ts` |
| `frontend/src/hooks/use-debounced-value.ts` | 13 | 0 | чисто (timeout cleanup) |
| `frontend/src/hooks/use-file-upload.ts` | 149 | 0 | FIXED-OK dragenter-only counter + zone paste |
| `frontend/src/hooks/use-invalidate.ts` | 27 | 0 | чисто (keys ↔ messages/channels/projects) |
| `frontend/src/hooks/use-messages.ts` | 235 | 1 | object-guard OK; reaction field soft → A3W11-3-002 |
| `frontend/src/hooks/use-notifications.ts` | 305 | 0 | FIXED-OK id/title guard + toastApiError |
| `frontend/src/hooks/use-presence.ts` | 79 | 0 | FIXED-OK isError + heartbeat warn |
| `frontend/src/hooks/use-typing.ts` | 72 | 0 | FIXED-OK payload object-guard + timer cleanup |

Файлов в списке: **19**. Прочитано: **19** (composer — отсутствие файла подтверждено; канон `frontend/src/lib/composer-ai-suggest.ts`). Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W11-3-001 | NEW | P3 | `dashboard-store.ts:101-109` | `loadLayout` catch только `console.error`, **не** пишет `error` и **не** rethrow | Callers (`dashboard.lazy.tsx` / project / admin) после `loadLayout().then(…)` видят пустой layout → подставляют default. Fail API неотличим от «раскладки нет» — пользователь теряет кастом без banner/retry (асимметрия с `saveLayoutDebounced` / `fetch*`, где `error` уже есть) | `set({ error: "Не удалось загрузить раскладку" })` и/или `throw`; UI — alert/retry; не затирать default поверх fail без явного UX |
| A3W11-3-002 | NEW | P3 | `use-messages.ts:128-144` | `reaction_toggled`: object-guard есть, но `reactions` без проверки типа — пишется как есть | Мусорный/`undefined` `reactions` затирает кэш реакций сообщения (в отличие от `message_edited`, где поля через `??` fallback) | Guard `Array.isArray(reactEvt.reactions)` (или schema); иначе skip update |

## FIXED-OK (Stage A W11-A3-* + A2 actives)

| Stage A / A2 ID | Evidence в HEAD |
|-----------------|-----------------|
| W11-A3-001 | `use-auth.ts:28-44` cancelled + tokenAtStart |
| W11-A3-002 | `call-store.ts:186-194` set error + **rethrow** |
| W11-A3-003 | dashboard `fetchProject`/`fetchAdmin` → `error` (`:75-98`) |
| W11-A3-004 | `leaveCall`/`endCall` fail → error + **return** без `initialState` (`:158-181`) |
| W11-A3-005 | sprint `fetchSprints` → `error` (`:76-80`) |
| W11-A3-006 | dragCounter только `onDragEnter` (`use-file-upload.ts:96-107`) |
| W11-A3-007 | paste на drop-zone props, не `document` (`:132-147`) |
| W11-A3-008 | `toggleHand` пишет `error` (`:197-207`) |
| W11-A3-009 | `saveLayoutDebounced` → error + rollback previous (`:127-135`) |
| W11-A3-010 | use-messages object/array guard (`:51-54`, `:194-197`) |
| W11-A3-011 | use-typing тот же guard (`:29-32`) |
| W11-A3-012 | ≡ A2W11-3-001 FIXED-OK (`isError`/`error`/`refetch`) |
| W11-A3-013 | ≡ A2W11-3-002 FIXED-OK (ticket subprotocol) |
| W11-A3-014 | sprint `error` на fetch/charts; Query migration — не блокер (как A2) |
| W11-A3-015 | notification.new: `id`+`title` string guard (`:286-295`) |
| W11-A3-016 | tms-store = filters only (`:16-34`) |
| W11-A3-017 | helpers в `lib/composer-ai-suggest.ts`; hooks-файл удалён |
| W11-A3-018 | timer `Number.isFinite` → `stopTimer` (`:56-66`) |
| A2W11-3-001…003 | см. таблицу сверки выше |

## Чисто

command / messenger / sidebar stores — UI-only; use-channels / invalidate / debounced — Query keys + WS cleanup OK; ws reconnect + generation cancel OK; presence heartbeat с логированием OK.

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 2 |
| **Total** | **2** |

| Метрика | Значение |
|---------|----------|
| Файлов | **19 / 19 / 0** |
| Активных | **2** (NEW 2 · REGRESS 0 · RESIDUAL 0) |
| FIXED-OK (A2 actives) | **3/3** |
| FIXED-OK (Stage A via A2) | **18** без регресса |

### Топ-3

1. **A3W11-3-001 (P3, NEW)** — `loadLayout` silent fail → default layout как «успех».
2. **A3W11-3-002 (P3, NEW)** — `reaction_toggled` без guard на `reactions`.
3. _(нет P2+)_ — A2 residuals presence/`?token=`/heartbeat — **FIXED-OK**.

---

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=0 · P3=2 · Total=2.  
Топ: A3W11-3-001, A3W11-3-002.
