# Frontend Architecture

**Аудит:** READ-ONLY · зона A10 · 2026-07-26  
**Ветка:** `cursor-grok-4.5`  
**Область:** `frontend/src/{routes,modules,components,stores,hooks,api,lib}`

## Что проверено

| Область | Фокус |
|---------|--------|
| Effects / cleanup | `use-websocket`, `ws-store`, `use-messages`, `use-channels`, `use-typing`, `use-presence`, `CallRoom`, `DeviceSelector`, `gif-picker`, `app-shell`, `mobile-layout`, `audit-log`, `planning-poker`, `time-tracker`, `recording-controls`, `CallReactions`, `notification-bell` |
| Race / abort | `CallButton`, `call.$code.lazy`, `gif-picker` AbortController, `kanban-board`/`user-management`/`audit-log` manual fetch, issues cursor pagination |
| TanStack Query | keys/invalidation (`use-invalidate`, messenger/tracker/tms/crm), `QueryStateView` adoption |
| Zustand ↔ Query | `messenger-store` + `use-channels`/`use-unread`, `notification-store`, `board-store`, `timer-store`, `dashboard-store` |
| Lists / perf | `chat-view` virtualizer, `board-column`, issues table, channel-list, gantt, CRM pipeline |
| Types | `any` / `@ts-ignore` (кроме generated `routeTree.gen.ts`) |
| UX states | loading/error/empty в messenger, admin, gantt, notifications |
| a11y | message-actions, mention-autocomplete, BookmarksButton, CreateChannel icon-buttons |

Паттерны grep: `useEffect`, `addEventListener`/`setInterval`, `AbortController`/`cancelled`, `queryKey`/`invalidateQueries`, `useVirtualizer`, `: any`/`@ts-ignore`, `QueryStateView`, `aria-`/`role=`.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A10-001 | P1 | `hooks/use-channels.ts:51-74` (+ `channel-list.tsx:493`) | WS `channel_created`/`channel_updated` пишут только в zustand (`addChannel`/`updateChannel`); UI читает `useQuery` data | Каналы/переименования от других клиентов не появляются в списке до refetch; store↔query drift. Локальный create инвалидирует query (`useInvalidateChannels`), WS-путь — нет | В WS-handler: `queryClient.setQueryData` / `invalidateQueries(['messenger','channels'])`; либо ChannelList читать store после sync |
| A10-002 | P1 | `routes/.../issues/index.tsx:74-81` (+ `244`, `321`) | Cursor «Загрузить ещё»: в `allIssues` кладётся *текущая* страница, cursor двигается вперёд — новая страница в UI не попадает, пока не кликнут ещё раз | Off-by-one: после 1-го клика видна только page1, page2 в кэше Query, но `displayIssues = allIssues` без merge ответа | `useInfiniteQuery` или при успехе fetch с cursor аппендить `data.items` в `allIssues` (effect по `data`+cursor), не в click handler |
| A10-003 | P1 | `stores/ws-store.ts:63-65` (+ `hooks/use-websocket.ts:23-30`) | `onclose` обнуляет socket; reconnect только при смене `accessToken`/auth deps | Обрыв сети/proxy → UI «мёртвый» real-time до remount/logout; messenger/notifications/AI stream перестают обновляться | Backoff reconnect в `connect`; при `onclose` если ещё authenticated — schedule reconnect; сброс `socket` только после intentional `disconnect` |
| A10-004 | P1 | `hooks/use-presence.ts:24` | `userIds.sort().join(",")` мутирует входной массив | In-place `sort` меняет props/caller state → лишние ре-рендеры, нестабильный порядок у родителя, риск багов в React Strict Mode | `[...userIds].sort().join(',')` или `stableIds` без mutate |
| A10-005 | P2 | `components/messenger/channel-list.tsx:493-545` | Есть `isLoading` skeleton; `isError` не обрабатывается | Ошибка API → `channels ?? []` → пустой мессенджер как «нет каналов» | `QueryStateView` / ветка `isError` + retry (как в `deals-pipeline.tsx:59-64`) |
| A10-006 | P2 | `components/admin/user-management.tsx:186-202` | Manual fetch: `.catch(() => setUsers([]))`, нет `cancelled` | Сеть/403 выглядит как «пользователей нет»; unmount race → setState после размонтирования | `useQuery` + error UI; `let cancelled` / AbortSignal |
| A10-007 | P2 | `stores/notification-store.ts:54-68` (+ `115-167`) | Server state в zustand; ошибки только `console.warn`; нет race-guard на fetch/fetchMore | Дубль паттерна Query; UI не видит fail; параллельные fetchMore могут перепутать страницы | Перевести на `useInfiniteQuery` + toast; либо request-id/seq в store |
| A10-008 | P2 | `components/tracker/board-column.tsx:100-114` | Колонка рендерит все карточки + `AnimatePresence`/`motion` без виртуализации | Большие WIP-колонки (десятки–сотни) → тяжёлый DOM/layout (чат уже на `@tanstack/react-virtual`) | Virtualizer внутри scroll-container колонки; memo `IssueCard` уже есть — оставить |
| A10-009 | P2 | `components/tracker/time-tracker.tsx:42-56` | `const timer = useTimerStore()` без селектора; deps `[…, timer]` | Каждый `tick()` меняет state → новый snapshot → effect teardown/setup каждую секунду | Селекторы: `useTimerStore(s => s.tick)` + флаги; deps без всего store |
| A10-010 | P2 | `components/messenger/mention-autocomplete.tsx:200-228` | `role="listbox"`, пункты — `<button>` без `role="option"` / `aria-selected` | Screen reader не связывает активный пункт с listbox (клавиатура при этом есть) | `role="option"` + `aria-selected={index===selectedIndex}` на items |
| A10-011 | P2 | `components/query-state-view.tsx` (usage ~4 feature-файла) vs admin/gantt/notifications | `QueryStateView` есть, но почти не внедрён вне issues/board/projects/triage | Систематическое смешение error/empty (см. A10-005/006); gantt (`gantt-chart.tsx:60-62`) — loading text, без error | Rollout QueryStateView на admin/messenger/gantt/notifications |
| A10-012 | P2 | `modules/tracker/kanban-board.tsx:88-102` | Дубль Kanban: manual `useState`+fetch без cancel; канон — `routes/.../board/index.tsx` на Query | Dead/parallel path: race при смене boardId; нет optimistic/DnD как на board route (импортов `KanbanBoard` нет) | Удалить/свести к board route; не держать два источника |
| A10-013 | P3 | `channel-list.tsx:92-96` · `create-channel-dialog.tsx:22-26` | Icon-button только с `title`, без `aria-label` | Часть AT игнорирует `title` → безымянная кнопка | `aria-label="Закладки"` / `"Создать канал"` (как у Call/message-actions) |
| A10-014 | P3 | `modules/tracker/gantt-chart.tsx:21-25` | `limit: 100`, полный `.map` баров без virtual/window | Крупный проект: усечённый Gantt + тяжёлый DOM | Пагинация/фильтр + virtual row list; явный «показано N из M» |

## Чисто

- Messenger messages: cursor infinite query + WS `setQueryData` с dedupe по id (`use-messages.ts`); virtualizer в `chat-view.tsx`
- Effect cleanup в типовых местах: WS rooms, mediaDevices, matchMedia, keydown, IntersectionObserver (gif), CallReactions timers
- Race guards: `CallButton.tsx:40-52`, `call.$code.lazy.tsx:51-80`, gif `AbortController`
- Auth refresh single-flight (`api/client.ts:33-75`); rate-limit toast timer с clear (`lib/format-api-error.ts`)
- a11y сильный слой: `message-actions`, composer, CallControls, SkipLink в app-shell; `role="alert"` в QueryStateView/CRM
- Hand-written TS: нет `@ts-ignore` / `: any` вне generated `routeTree.gen.ts`
- Board page: Query keys `['board', projectId, activeBoardId]`, optimistic move + rollback
- `use-invalidate.ts` централизует messenger/projects invalidation для локальных мутаций

## Оценка зоны: 6.5/10

Зрелый каркас (Query + Zustand UI-state, virtual chat, Abort/cancelled в части flows, хороший a11y в composer/actions), но P1 по real-time channel list, issues pagination и WS reconnect ломают доверие к данным; QueryStateView и виртуализация колонок не доведены — зона «работает для happy path», не «production-hardened UX».
