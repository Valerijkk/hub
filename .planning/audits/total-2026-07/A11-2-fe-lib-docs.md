# W11 агент A2 — fe lib+docs part 2

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус FE: stores, hooks leaks (async/side-effects в stores), docs/API accuracy.  
> Список: W11-ASSIGNMENTS.md §A2 (19 файлов).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `frontend/src/api/github.ts` | 187 | 1 | замечания |
| `frontend/src/api/gitlab.ts` | 139 | 0 | чисто |
| `frontend/src/api/hr.ts` | 82 | 0 | чисто |
| `frontend/src/api/intranet.ts` | 171 | 1 | замечания |
| `frontend/src/api/issue-message-links.ts` | 57 | 0 | чисто |
| `frontend/src/api/marketplace.ts` | 142 | 1 | замечания |
| `frontend/src/api/messenger.ts` | 561 | 2 | замечания |
| `frontend/src/api/notifications.ts` | 160 | 0 | чисто |
| `frontend/src/api/projects.ts` | 136 | 1 | замечания |
| `frontend/src/api/search.ts` | 117 | 0 | чисто |
| `frontend/src/api/sprint.ts` | 199 | 0 | чисто |
| `frontend/src/api/status.ts` | 18 | 0 | чисто |
| `frontend/src/api/tms.ts` | 818 | 1 | замечания |
| `frontend/src/api/tracker.ts` | 670 | 1 | замечания |
| `frontend/src/api/types.ts` | 19 | 1 | замечания |
| `frontend/src/api/webhooks.ts` | 61 | 0 | чисто |
| `frontend/src/stores/ai-store.ts` | 93 | 2 | замечания |
| `frontend/src/stores/auth-store.ts` | 133 | 3 | замечания |
| `frontend/src/stores/board-store.ts` | 55 | 0 | чисто |

Файлов в списке: 19. Прочитано: 19. Пропущено: 0.

## Counts

| Метрика | Значение |
|---------|--------:|
| Файлов | 19 |
| Файлов с находками | 10 |
| Файлов чисто | 9 |
| Находок всего | 14 |
| P0 | 0 |
| P1 | 1 |
| P2 | 9 |
| P3 | 4 |

## Топ-3

1. **W11-A2-001 (P1)** — `messenger.ts` DTO `LinkPreviewData` не совпадает с backend `/link-preview` (`image`/`site_name` вместо `image_url`/`domain`) → active UI (`components/messenger/link-preview`) не показывает OG-картинки.
2. **W11-A2-002 (P2)** — `ai-store.ts` при ошибке `loadBotInfo` ставит `loaded: true` навсегда → бот не узнаётся до reload, retry невозможен.
3. **W11-A2-006 (P2)** — `tms.ts` нет `createEnvironment`/`deleteEnvironment` при наличии BE → UI ходит через голый `api.post`/`api.delete` в обход типизированного клиента.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W11-A2-001 | P1 | `messenger.ts:419-433` | `LinkPreviewData` / `fetchLinkPreview`: поля `image`, `site_name`; нет `type`/`url`/`image_url`/`domain` | Backend (`github.LinkPreviewResponse`) и канон в `github.ts:36-48` отдают `image_url`/`domain`. Active chat (`components/messenger/link-preview.tsx`) читает `data.image` → картинка всегда пустая; title/description работают случайно | Удалить дубль: реэкспорт `getLinkPreview`/`LinkPreviewResponse` из `github.ts` (или общий `api/link-preview.ts`); поправить consumer |
| W11-A2-002 | P2 | `ai-store.ts:37-48` | `loadBotInfo`: в `catch` — `set({ loaded: true })` без флага ошибки | Сбой сети/403 фиксирует «загружено»; `if (get().loaded) return` блокирует повтор; `isBotMessage` всегда false | `loaded` только на success; отдельно `error`/`retry`; не глотать fail |
| W11-A2-003 | P2 | `ai-store.ts:50-68`; `:71-77` | `loadModels` / `saveSelectedModel`: пустой `catch` | Ошибка провайдеров/prefs неотличима от «список пуст»; optimistic `selectedModel` без rollback при fail save | Toast/`error` в store; rollback model на fail; лучше Query+mutation |
| W11-A2-004 | P2 | `auth-store.ts:121-128` | `logout` всегда `broadcastAuth`; нет silent apply для remote | `_authenticated` при remote logout вызывает `storeLogout()` → повторный broadcast (комментарий «без повторного broadcast» ложен относительно API стора) | Добавить `clearSessionLocal()` без broadcast; remote handler — только его |
| W11-A2-005 | P2 | `auth-store.ts:12-15`; `:131` | Access + refresh JWT в `sessionStorage` (`hub-auth`) | XSS → кража refresh (30d); sessionStorage не защищает от script | httpOnly cookie для refresh (ADR) или memory-only access + refresh rotation; минимум — не класть refresh в persist partial |
| W11-A2-006 | P2 | `tms.ts:773-775` | Есть только `listEnvironments`; нет create/delete | BE: `POST/DELETE …/environments`; UI (`environments.tsx`) бьёт в `api.post`/`api.delete` в обход DTO — drift и риск расхождения путей | Добавить `createEnvironment`/`deleteEnvironment` (+ типы) и перевести UI |
| W11-A2-007 | P2 | `intranet.ts:36-42` | Есть `createNews`/`deleteNews`, нет `updateNews` | BE: `PUT /intranet/news/:id` + `UpdateNewsRequest` с `version` — FE-клиент не покрывает контракт | Добавить `updateNews(id, {…, version})` |
| W11-A2-008 | P2 | `projects.ts:118-126` | Есть add/remove member, нет смены роли | BE: `PUT /projects/:id/members/:userId` (`UpdateMemberRoleRequest`) | Добавить `updateMemberRole(projectId, userId, role)` |
| W11-A2-009 | P2 | `types.ts:2-3`; `tracker.ts:268-277` | `PaginatedResponse` и комментарий обещают offset; `ListIssuesParams` — только `cursor` | TZ (`tz_constraints`): задачи — offset/limit; FE+BE list issues — cursor. Docs/API drift внутри клиента и относительно ТЗ | Либо выровнять tracker на offset, либо поправить ТЗ/комментарий `types.ts` под факт (cursor для issues) |
| W11-A2-010 | P2 | `messenger.ts:94-120` | `buildMessageFromWsEvent(evt: Record<string, unknown>)` + цепочка `as` | Битый WS payload → runtime с пустым `sender_id`/дефолтами; обход типизации | Zod/type-guard на MessageSentEvent; не принимать произвольный Record |
| W11-A2-011 | P3 | `auth-store.ts:112-118` | `setTokens` шлёт `{ type: "login", userId: null }` | Другие вкладки всегда `reload` без знания userId; нельзя отличить same-user re-login | Передавать `userId` после `/auth/me` или отдельный `setUser` broadcast |
| W11-A2-012 | P3 | `auth-store.ts:87-101` | Fallback `storage` слушает `hub-auth-event`, но `broadcastAuth` пишет только в BroadcastChannel | Dead defence-in-depth; комментарий вводит в заблуждение | Писать mirror в `localStorage` при broadcast **или** удалить fallback |
| W11-A2-013 | P3 | `github.ts:183-187` | Дубль клиента `/link-preview` рядом с `messenger.fetchLinkPreview` | Два контракта на один endpoint → W11-A2-001 | Один модуль-источник правды |
| W11-A2-014 | P3 | `marketplace.ts:62`; `:66`; `:72`; `:82`; `:96` | Вызовы `api.get`/`post`/`put` без явного `<T>` (тип только с return-annotation) | Хрупко: снятие annotation → `unknown`; расходится со стилем остальных api-файлов | Явные дженерики: `api.get<{ items: MarketplaceListing[]; total: number }>(…)` |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `gitlab.ts` | Пути integrations/pipelines/branch; типы status union; без секретов в response DTO |
| `hr.ts` | Leave CRUD + approve/reject/cancel с `version`; labels RU; пагинация N/A |
| `issue-message-links.ts` | CRUD + `listThreadIssueLinks` с `encodeURIComponent`; DTO с `channel_id` |
| `notifications.ts` | List/prefs/telegram/push; unsubscribe DELETE+body `{endpoint}` совпадает с BE |
| `search.ts` | `/search` + filters; BE default limit 20 — omit `limit` ок |
| `sprint.ts` | Sprint/poker/burndown/velocity; зеркалит BE DTO (без version — как на бэке) |
| `status.ts` | Overall `operational\|degraded` = BE aggregate; component `outage\|unknown` учтён |
| `webhooks.ts` | Create под project; update/delete/deliveries `/webhooks/:id` — как BE (RBAC на id) |
| `board-store.ts` | Только UI state (filters/drag); нет async/timers/listeners → leak-класса нет |
| Stores vs hooks | В §A2 нет `hooks/*`; утечки useEffect проверялись косвенно по store side-effects (BroadcastChannel singleton ок; ai/auth async) |

## Notes

- Active messenger path: `components/messenger/*` → сломанный `fetchLinkPreview`. `modules/messenger/link-preview.tsx` уже на каноне `github.getLinkPreview` + cleanup в `useEffect` — эталон для фикса W11-A2-001.
- `marketplace`/`hr` optimistic `version` на мутациях — хорошо; `intranet` delete без version согласован с soft-delete BE.
