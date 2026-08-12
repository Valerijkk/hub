# W11 агент A3 — fe lib+docs part 3 (stores + hooks)

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус FE: hooks/lib · useEffect leaks · hook deps · Query keys/invalidate · zustand↔query dup · any/ts-ignore · loading/empty/error · docs drift (в комментариях/контрактах этих файлов; `docs/*.md` — у A4/A5)

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/stores/call-store.ts | 204 | 3 | замечания |
| frontend/src/stores/command-store.ts | 13 | 0 | чисто |
| frontend/src/stores/dashboard-store.ts | 134 | 2 | замечания |
| frontend/src/stores/messenger-store.ts | 132 | 0 | чисто |
| frontend/src/stores/sidebar-store.ts | 27 | 0 | чисто |
| frontend/src/stores/sprint-store.ts | 120 | 2 | замечания |
| frontend/src/stores/timer-store.ts | 64 | 1 | замечания |
| frontend/src/stores/tms-store.ts | 83 | 1 | замечания |
| frontend/src/stores/ws-store.ts | 183 | 1 | замечания |
| frontend/src/hooks/use-auth.ts | 70 | 1 | замечания |
| frontend/src/hooks/use-channels.ts | 126 | 0 | чисто |
| frontend/src/hooks/use-composer-ai-suggest.ts | 43 | 1 | замечания |
| frontend/src/hooks/use-debounced-value.ts | 13 | 0 | чисто |
| frontend/src/hooks/use-file-upload.ts | 136 | 2 | замечания |
| frontend/src/hooks/use-invalidate.ts | 27 | 0 | чисто |
| frontend/src/hooks/use-messages.ts | 227 | 1 | замечания |
| frontend/src/hooks/use-notifications.ts | 298 | 1 | замечания |
| frontend/src/hooks/use-presence.ts | 60 | 1 | замечания |
| frontend/src/hooks/use-typing.ts | 68 | 1 | замечания |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W11-A3-001 | P1 | frontend/src/hooks/use-auth.ts:27 | `getMe().then(setUser)` без cancelled/generation guard: после `storeLogout()` (или смены токена) ответ всё ещё вызывает `setUser` | Logout чистит `user: null`, затем поздний `setUser` возвращает профиль. Следующий login (`setTokens`) не триггерит refetch (`!user` ложно) → UI показывает **чужой/устаревший** профиль при новых токенах | AbortController / `let cancelled` + cleanup; в `.then` проверять `getState().accessToken === tokenAtStart`; при logout инкрементировать generation |
| W11-A3-002 | P1 | frontend/src/stores/call-store.ts:174 | `checkActiveCall`: любой catch → `return null` (нет различия «нет звонка» / «сеть/403») | `CallButton` трактует null как «Начать звонок» → при ошибке API пользователь создаёт **второй** звонок в канале | Пробрасывать error / возвращать `{ call, error }`; UI: toast + disabled start при fail |
| W11-A3-003 | P1 | frontend/src/stores/dashboard-store.ts:75 | `fetchProject` / `fetchAdmin` (стр. 87): catch только `console.error`, **не** пишут `error`; `fetchPersonal` (стр. 67) — пишет | Project/admin dashboard: `loading=false` + старые/`null` data → empty widgets неотличимы от fail (`dashboard.lazy.tsx` рендерит `WidgetGrid` при `!loading`) | `set({ error: "…" })` как в `fetchPersonal`; UI — banner/retry; сбрасывать `error` в начале fetch |
| W11-A3-004 | P1 | frontend/src/stores/call-store.ts:148 | `leaveCall` / `endCall` (стр. 160): API fail глотается, локально всегда `set(initialState)` | Host думает, что завершил звонок, на сервере call ещё active; участники остаются в комнате | Не reset при fail `endCall` (или retry + toast); leave — хотя бы toast при fail; писать `error` |
| W11-A3-005 | P1 | frontend/src/stores/sprint-store.ts:73 | `fetchSprints` catch: только `loading: false`, без `error`; список остаётся `[]` | `sprint-board.tsx` показывает empty «нет спринтов» при сетевом сбое | Поле `error` + set в catch; UI empty только при успехе с `[]` |
| W11-A3-006 | P2 | frontend/src/hooks/use-file-upload.ts:100 | `onDragOver` инкрементирует `dragCounter` на **каждый** dragover (частый event), не dragenter | Счётчик уходит в огромные значения → `onDragLeave` не обнуляет → `isDragging` залипает true | Инкремент только в `onDragEnter`; dragover — только `preventDefault` |
| W11-A3-007 | P2 | frontend/src/hooks/use-file-upload.ts:76 | Глобальный `document` paste listener на всё время mount Dropzone | Ctrl+V с файлом в любом поле страницы (чат, формы) уходит в upload сущности Dropzone | Слушать paste только на drop-zone / при focus зоны; или игнорировать, если target — editable |
| W11-A3-008 | P2 | frontend/src/stores/call-store.ts:182 | `toggleHand`: catch пустой — UI не меняется и без feedback | Пользователь жмёт «рука» — ничего; неясно, сеть или запрет | set `error` / проброс; CallControls — toast |
| W11-A3-009 | P2 | frontend/src/stores/dashboard-store.ts:123 | `saveLayoutDebounced`: fail только `console.error`; локальный layout уже обновлён | Пользователь видит сохранённую раскладку; после reload — откат без предупреждения | toast/error flag; опционально rollback local layout |
| W11-A3-010 | P2 | frontend/src/hooks/use-messages.ts:51 | WS handler: `payload as Record` без null/object guard (в отличие от `use-channels.ts:81`) | `payload == null` → TypeError в listener → остальные WS-обработчики того же тика страдают | Guard `if (!payload \|\| typeof payload !== "object") return` |
| W11-A3-011 | P2 | frontend/src/hooks/use-typing.ts:30 | То же: typing payload без null-check | Битый/пустой WS typing → исключение в общем listener | Тот же object-guard |
| W11-A3-012 | P2 | frontend/src/hooks/use-presence.ts:55 | Возврат только `statuses` / `isLoading`; нет `isError` | Fail poll → все «offline»; путает сеть и реальный offline | Прокинуть `isError`/`error`; UI — «статус недоступен» |
| W11-A3-013 | P2 | frontend/src/stores/ws-store.ts:102 | Access JWT в query string `/ws?token=` | Токен попадает в access/proxy logs, history, Referer при редких кейсах | Предпочтительно subprotocol / first-message auth (если бэкенд позволит); минимум — не логировать query |
| W11-A3-014 | P2 | frontend/src/stores/sprint-store.ts:59 | Server-state (sprints/burndown/velocity) в zustand без Query keys; нет error; мутации без централизованного invalidate | Дубль паттерна «imperative store» vs TanStack в мессенджере/TMS; silent fail + stale между экранами | Мигрировать на `useQuery`/`useMutation` + invalidate; или добавить error + явный refresh contract |
| W11-A3-015 | P2 | frontend/src/hooks/use-notifications.ts:286 | WS `notification.new`: `payload as NotificationDTO` без проверки полей | Мусорный payload → toast с undefined + порча InfiniteQuery cache | Guard `id`/`title` (как channel_created в use-channels) |
| W11-A3-016 | P3 | frontend/src/stores/tms-store.ts:26 | Поля `testCases`/`suites`/`testRuns`/`selected*` + setters; потребители (`cases.lazy.tsx`) используют **только** `filters` | Мёртвый server-state API в store при живом TanStack Query — drift контракта store↔UI | Удалить неиспользуемые поля или документировать deprecated; оставить filters-only |
| W11-A3-017 | P3 | frontend/src/hooks/use-composer-ai-suggest.ts:1 | Файл в `hooks/`, но React-хука нет — только pure helpers | Docs/naming drift: ожидание Rules-of-Hooks от `use-*` в hooks/ | Перенести в `lib/` или добавить реальный `useComposerAiSuggest` |
| W11-A3-018 | P3 | frontend/src/stores/timer-store.ts:56 | `tick`: битый `startedAt` из persist → `NaN` в `elapsedSeconds` | UI таймера показывает NaN после повреждения localStorage | Guard `Number.isFinite`; иначе `stopTimer()` |

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 5 |
| P2 | 10 |
| P3 | 3 |
| **Σ** | **18** |

### Топ-3
1. **W11-A3-001** — гонка `getMe` → `setUser` после logout/смене сессии → чужой профиль.
2. **W11-A3-002** — `checkActiveCall` глотает ошибки как «нет звонка» → риск дубля звонка.
3. **W11-A3-003** — dashboard project/admin fail без `error` → пустые виджеты как успех.

## Чисто
Проверено по фокусу W11 A3 (hooks leaks/deps · Query · zustand↔query · loading/error · docs drift в этих файлах):

| Файл | Что проверено |
|------|----------------|
| command-store.ts | Локальный UI open/toggle; нет async/Query/leak |
| messenger-store.ts | UI+unread+typing; server list пишет `useChannels` в Query (A10-001) — зеркало осознанное, не dup-баг |
| sidebar-store.ts | persist collapsed; нет секретов |
| use-channels.ts | Query key `CHANNELS_QUERY_KEY`; WS room `user:{id}:channels` + cleanup; payload guards; sync в zustand только mirror |
| use-debounced-value.ts | setTimeout + clear в cleanup |
| use-invalidate.ts | Keys совпадают с `use-messages` / `use-channels` / projects |
| use-notifications.ts (кроме WS cast) | InfiniteQuery + mutations с toastApiError; prepend без дублей; listener cleanup |
| use-typing.ts (кроме null payload) | timers Map clear на unmount; skip self typing |

**Query invalidation (OK в scope):** `use-invalidate` ключи согласованы с `useMessages` / `useChannels`. Notifications патчат cache в mutations без silent warn.

**zustand vs query:** messenger channels — Query source of truth + zustand mirror (OK). sprint-store — полный server cache в zustand (W11-A3-014). tms — Query в routes, store filters-only фактически (W11-A3-016). call/ws/timer/command/sidebar — уместны как client/live state.

**Docs drift:** в списке A3 нет `docs/*.md` (A4/A5). Внутри среза: naming `hooks/use-composer-ai-suggest` без хука; tms-store API шире реального использования; JSDoc use-file-upload про presigned — соответствует `api/files.uploadFile`.

**Хорошие паттерны:** ws-store reconnect + backoff; use-channels upsert/patch helpers; use-notifications toastApiError; use-presence `presenceQueryKeyPart` без in-place sort; use-auth logout сначала API потом local clear.
