# W10 агент A1 — frontend routes+modules part 1

> TASK-008 · Стадия A · read-only · 2026-07-28  
> Фокус FE: TanStack Router only · Query invalidation · zustand vs query dup · auth guards · loading/error

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| frontend/src/modules/accounting/accounting-page.tsx | 383 | 1 | замечания |
| frontend/src/modules/admin/backups.tsx | 216 | 1 | замечания |
| frontend/src/modules/bases/bases-page.tsx | 463 | 1 | замечания |
| frontend/src/modules/calendar/unified-calendar.tsx | 462 | 0 | чисто |
| frontend/src/modules/calls/calendar-agenda-view.tsx | 101 | 0 | чисто |
| frontend/src/modules/calls/calendar-day-view.tsx | 125 | 0 | чисто |
| frontend/src/modules/calls/calendar-month-view.tsx | 155 | 0 | чисто |
| frontend/src/modules/calls/calendar-types.ts | 213 | 0 | чисто |
| frontend/src/modules/calls/calendar-week-view.tsx | 170 | 0 | чисто |
| frontend/src/modules/calls/call-calendar.tsx | 505 | 1 | замечания |
| frontend/src/modules/calls/create-call-dialog.tsx | 511 | 3 | замечания |
| frontend/src/modules/calls/recording-controls.tsx | 129 | 1 | замечания |
| frontend/src/modules/calls/recordings-list.tsx | 209 | 1 | замечания |
| frontend/src/modules/calls/scheduled-calls.tsx | 314 | 2 | замечания |
| frontend/src/modules/crm/crm-page.tsx | 373 | 1 | замечания |
| frontend/src/modules/crm/deals-pipeline.tsx | 150 | 0 | чисто |
| frontend/src/modules/hr/leave-requests.tsx | 259 | 1 | замечания |
| frontend/src/modules/intranet/intranet-hub.tsx | 57 | 1 | замечания |
| frontend/src/modules/intranet/kudos-feed.tsx | 176 | 1 | замечания |
| frontend/src/modules/intranet/news-feed.tsx | 178 | 0 | чисто |
| frontend/src/modules/intranet/org-directory.tsx | 267 | 0 | чисто |
| frontend/src/modules/intranet/random-coffee.tsx | 129 | 1 | замечания |
| frontend/src/modules/marketplace/marketplace-page.tsx | 472 | 2 | замечания |
| frontend/src/modules/messenger/link-preview.tsx | 238 | 0 | чисто |
| frontend/src/modules/messenger/message-bubble.tsx | 212 | 1 | замечания |
| frontend/src/modules/settings/ai-privacy.tsx | 93 | 1 | замечания |
| frontend/src/modules/settings/api-keys.tsx | 203 | 2 | замечания |
| frontend/src/modules/settings/appearance.tsx | 139 | 0 | чисто |
| frontend/src/modules/settings/project-templates.tsx | 164 | 1 | замечания |

Файлов в списке: 29. Прочитано: 29. Пропущено: 0.

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W10-A1-001 | P1 | frontend/src/modules/calls/create-call-dialog.tsx:108 | `projectId` берётся как `projects[0].id` без выбора проекта; `isValid` (стр. 206) не требует `projectId` | Созвон всегда создаётся в первом проекте списка (или с `""` → silent fail API). Пользователь в другом проекте получает неверную привязку / ошибку без UI | Выбор проекта (или project из канала/контекста); включить `!!projectId` в `isValid`; показать ошибку если проектов нет |
| W10-A1-002 | P1 | frontend/src/modules/calls/create-call-dialog.tsx:175 | `useMutation` без `onError` / toast; при fail диалог остаётся без объяснения | Сценарий «создать созвон» ломается незаметно — нет feedback, кнопка только снимает `isPending` | `onError: () => toast.error(...)`; опционально оставить форму открытой с сообщением |
| W10-A1-003 | P1 | frontend/src/modules/calls/scheduled-calls.tsx:50 | Список созвонов в `useState` + ручной `loadCalls`, тогда как `call-calendar` / `CreateCallDialog` живут на Query `["calls","scheduled",…]` | Дубль источника правды: invalidate после create в диалоге (стр. 180 create-call-dialog) не обновляет `ScheduledCalls`; два UI показывают разный набор | Перевести `ScheduledCalls` на `useQuery` с тем же key-префиксом + invalidate |
| W10-A1-004 | P2 | frontend/src/modules/calls/call-calendar.tsx:155 | `useQuery` scheduled calls без `isLoading` / `isError` в UI | Ошибка сети = пустой календарь («нет событий»); нет скелетона загрузки | Деструктурировать `isLoading`/`isError` и показать status/alert (как `unified-calendar.tsx:361–369`) |
| W10-A1-005 | P2 | frontend/src/modules/accounting/accounting-page.tsx:32 | Пять `useQuery` без `isLoading`/`isError`; UI сразу рисует пустые таблицы/формы | Fail API неотличим от empty ledger; админ может создать проводку на «пустом» плане счетов | Loading + error/retry per section или общий banner |
| W10-A1-006 | P2 | frontend/src/modules/bases/bases-page.tsx:41 | Список баз: есть `isLoading`, нет `isError` | Ошибка listBases → пустой sidebar, как «нет баз» | `isError` + alert; не рендерить empty как успех |
| W10-A1-007 | P2 | frontend/src/modules/crm/crm-page.tsx:41 | Контакты/проекты без loading/error UI (`contacts.length === 0` = empty) | Сбой listContacts выглядит как «Нет контактов» | isPending/isError как в `deals-pipeline.tsx:52–64` |
| W10-A1-008 | P2 | frontend/src/modules/admin/backups.tsx:28 | `fetchBackups` без `catch`: ошибка → `finally` сбрасывает loading, `backups` остаётся `[]` | Empty и fail сливаются; create/restore/delete тоже без toast на ошибку | catch + error state/toast; не показывать empty при fail |
| W10-A1-009 | P2 | frontend/src/modules/settings/api-keys.tsx:20 | `fetchKeys` без catch (fail → «Нет API-ключей»); `handleCreate`/`handleRevoke` без try/catch → unhandled rejection | Пользователь не видит сбой list/create/revoke | toastApiError + error UI; try/catch на мутациях |
| W10-A1-010 | P2 | frontend/src/modules/settings/api-keys.tsx:52 | `setTimeout(() => setCopied(false), 2000)` без cleanup | Unmount до 2s → setState на размонтированном | clearTimeout в cleanup / ref |
| W10-A1-011 | P2 | frontend/src/modules/settings/ai-privacy.tsx:19 | После ошибки `getAIPreferences` `loading=false`, `allowExternal` остаётся default `true` | Форма показывает «разрешено облако», хотя сервер не ответил; Save может перезаписать реальные prefs | Error-state + disable save / retry; не показывать форму как валидные prefs |
| W10-A1-012 | P2 | frontend/src/modules/settings/project-templates.tsx:26 | `TemplateSelector`: `api.get` без `.catch`/loading/error | Fail → пустая сетка шаблонов при создании проекта | loading + error; catch с toast |
| W10-A1-013 | P2 | frontend/src/modules/calls/scheduled-calls.tsx:60 | `loadCalls` / delete / create: `catch { /* пусто */ }` | Ошибки загрузки/CRUD невидимы; empty после fail | toast + error flag; не глотать |
| W10-A1-014 | P2 | frontend/src/modules/calls/recordings-list.tsx:43 | `loadRecordings` / download: catch пустой → empty «Записей пока нет» / silent download fail | Путает empty и fail; download без feedback | toast + isError; сообщение при ошибке download |
| W10-A1-015 | P2 | frontend/src/modules/calls/recording-controls.tsx:69 | catch комментирует «ошибка в store», но store не получает error — только silent fail | Start/stop записи без UI-ошибки | toast / set error в call-store |
| W10-A1-016 | P2 | frontend/src/modules/intranet/intranet-hub.tsx:19 | `useState(tabFromSearch(search.tab))` — tab не синхронизируется с URL при смене search; клик по вкладке не пишет `?tab=` | Deep-link `?tab=kudos` ломается при back/forward; шаринг URL не отражает активную вкладку (TanStack Router) | Controlled tab через `useSearch` + `navigate({ search })` |
| W10-A1-017 | P2 | frontend/src/modules/hr/leave-requests.tsx:39 | `useQuery` my leave: loading есть, `isError` нет | Fail = «Заявок пока нет» | isError alert |
| W10-A1-018 | P2 | frontend/src/modules/intranet/kudos-feed.tsx:30 | feed/received queries без `isError` | Fail → empty empty-state | isError UI |
| W10-A1-019 | P2 | frontend/src/modules/intranet/random-coffee.tsx:39 | `enrollMutation` без `onError` | Toggle opt-in откатывается только после refetch; fail незаметен (checkbox уже flipped via mutate) | onError + toast; optimistic rollback или controlled checked от server |
| W10-A1-020 | P2 | frontend/src/modules/marketplace/marketplace-page.tsx:46 | browse/mine/pending без `isError`; empty при fail (стр. 445) | Сбой browse = «нет объявлений» | isError + retry |
| W10-A1-021 | P2 | frontend/src/modules/marketplace/marketplace-page.tsx:76 | `approveMutation` / `rejectMutation` / `soldMutation` / `withdrawMutation` без `onError` | Модерация/статус fail без toast (в отличие от create/update/delete) | onError toast на все мутации |
| W10-A1-022 | P3 | frontend/src/modules/messenger/message-bubble.tsx:117 | `extractUrls(contentText)`, а отображается `displayText` из `contentJson` | Превью ссылок могут не совпасть с видимым текстом Plate (URL только в JSON / только в text) | `extractUrls(displayText)` или объединить оба источника |

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 18 |
| P3 | 1 |
| **Σ** | **22** |

### Топ-3
1. **W10-A1-001** — CreateCallDialog всегда пишет созвон в `projects[0]` (или пустой id).
2. **W10-A1-002** — та же мутация без `onError` → silent fail создания.
3. **W10-A1-003** — `ScheduledCalls` на local state vs Query в календаре → рассинхрон после invalidate.

## Чисто
Проверено по фокусу W10 (Router / Query invalidate / zustand↔query / auth UI / loading-error):

| Файл | Что проверено |
|------|----------------|
| unified-calendar.tsx | TanStack `useNavigate`/`useQuery`; loading+error; invalidate не нужен (read-only+export toast) |
| calendar-agenda/day/month/week-view.tsx | Presentational; без Query/router-dup; empty agenda ок |
| calendar-types.ts | Чистые типы/хелперы RRULE; логики UI нет |
| deals-pipeline.tsx | invalidate `["crm","pipeline"]`; loading/error; navigate TanStack only |
| news-feed.tsx | invalidate news; isLoading+isError; admin UI via `isSystemAdmin` |
| org-directory.tsx | loading/error; navigate+search `user`; mailto внешний `<a>` ок |
| link-preview.tsx | cancelled flag; loading skeleton; error→null; external `<a rel=noopener>` |
| appearance.tsx | theme/accent local (не server query); a11y radiogroup |

**TanStack Router only:** в 29 файлах нет `react-router-dom` / `BrowserRouter` / `window.location` для навигации. Внешние `href` (mailto, OG links) допустимы. Route-level auth (`_authenticated` / `beforeLoad`) вне списка A1 — в модулях только UI-гейты `isSystemAdmin` + `useAuthStore` (ожидаемо; сервер остаётся источником правды).

**Query invalidation (OK):** accounting `["accounting"]`, bases, crm contacts/pipeline, hr leave (+calendar), intranet news/kudos/coffee, marketplace — после мутаций invalidate присутствует.

**zustand vs query:** `useAuthStore` / `useCallStore` — identity и live recording, не дубль list-query. Единственный явный dup — W10-A1-003 (`ScheduledCalls` state vs Query).

**Хорошие паттерны:** `deals-pipeline` loading/error; `org-directory`/`news-feed`/`unified-calendar` error UI; `link-preview` cancelled fetch; HR/marketplace invalidate + calendar cross-key.
