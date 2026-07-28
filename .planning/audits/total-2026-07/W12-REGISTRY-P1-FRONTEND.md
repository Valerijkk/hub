# W12 — реестр P1 frontend (стадия B, дедуп)

> TASK-008 · 2026-07-28 · Источники: `A9-*.md`, `A10-*.md`, `A11-*.md` + `W{9,10,11}-SUMMARY.md`  
> Режим: **только реестр** — `TASK-*.md` не заводятся этим артефактом.  
> Product code не менялся.

## Сводка

| Метрика | Значение |
|---------|---------:|
| Сырых P1 FE (W9+W10+W11) | **42** |
| Классов (готовы к постановке) | **10** |
| Source ID в классах | **38** |
| «Требует уточнения» | **4** |
| P0 FE | 0 |

Разбивка сырых: W9=11 · W10=14 · W11=17.

---

## Классы

| ClassID | Severity | Title | Source IDs | Primary file:line | Mutation criterion | Suggested TASK slug |
|---------|----------|-------|------------|-------------------|--------------------|---------------------|
| FE-P1-API-CLIENT | P1 | Обход `api`-клиента / бесконечный refresh на 401 | W11-A1-001, W11-A1-002, W10-A2-001 | `frontend/src/api/client.ts:79` | Убрать флаг one-retry после refresh → второй 401 крутит refresh до hang; вернуть сырой `fetch` без Bearer в import-export/calendar → 401 на авторизованном сценарии | `fe-api-client-auth-retry` |
| FE-P1-CALLS-CHAT | P1 | CallChat: нет WS-комнаты `call:` + мусор в ленте | W9-A1-001, W9-A1-002 | `frontend/src/components/calls/CallChat.tsx:55` | Убрать `subscribe("call:"+id)` → чужие сообщения не появляются; убрать фильтр `event==="chat.message"` → в ленту попадают `call.joined`/`left` без text | `fe-callchat-ws-subscribe` |
| FE-P1-CALLS-LIFECYCLE | P1 | Join/create/leave/active: сломанный сценарий звонка | W9-A2-001, W10-A1-001, W10-A1-002, W10-A1-003, W11-A3-002, W11-A3-004 | `frontend/src/components/dashboard/upcoming-calls.tsx:79` | «Войти» → `startCall`/`no-op` вместо `joinCall(id)` → новый звонок или тишина; `checkActiveCall` всегда `null` на ошибке → второй create; fail `endCall` + local reset → UI «вышел», сервер active | `fe-calls-join-create-lifecycle` |
| FE-P1-MESSENGER | P1 | Reply/thread/voice/preview/unread — сломанные сценарии чата | W9-A3-001, W9-A3-002, W9-A3-003, W10-A2-002, W11-A2-001, W11-A4-001 | `frontend/src/components/messenger/message-composer.tsx:320` | `sendMessage` без `replyToMessage.id` → сообщение без reply; `limit:1` parent → пустая шапка треда; `stop()` до `onstop` → Promise зависает; link-preview читает `image` при API `image_url` → нет картинки; unread poll без replace map → badge >0 после MarkRead на другом клиенте | `fe-messenger-reply-thread-unread` |
| FE-P1-BOARD-DND | P1 | Канбан: двойной `moveIssue` + константный lexorank | W10-A4-001, W10-A4-002 | `frontend/src/routes/_auth/board/index.lazy.tsx:124` | Оставить оба `onDrop` → два mutate на один DnD (двойной optimistic / 409); `rank` всегда `"0|aaaaaa:"` → коллизии порядка в колонке | `fe-board-dnd-single-path-rank` |
| FE-P1-SPRINT-UI | P1 | Sprint UI: loading/error = empty / вечный спиннер | W10-A3-002, W10-A4-003, W11-A3-005 | `frontend/src/modules/tracker/planning-poker.tsx:101` | `pokerSession===null` до ответа → ложный «Сессия не найдена»; `getSprint` без catch → вечный «Загрузка»; `fetchSprints` catch → `[]` как «нет спринтов» | `fe-sprint-loading-error-states` |
| FE-P1-TRASH-WEBHOOKS | P1 | Корзина пуста (неверный API) + webhooks с `projectId=""` | W10-A5-001, W10-A5-003 | `frontend/src/routes/_auth/trash.tsx:36` | `listProjects` вместо `/projects/trash` → soft-deleted не видны; `WebhooksPage projectId=""` → запросы `/projects//webhooks` 4xx, секция Settings мёртвая | `fe-trash-list-and-webhooks-project` |
| FE-P1-ERROR-AS-EMPTY | P1 | Fail/сеть = empty; ErrorBoundary не сбрасывается; auth race | W9-A4-002, W9-A5-001, W11-A3-001, W11-A3-003 | `frontend/src/hooks/use-auth.ts:27` | Глотнуть error activity-feed → «Нет активности»; не сбрасывать boundary при смене route → sidebar nav оставляет error UI; late `setUser` после logout → чужой профиль; dashboard project/admin без `error` → пустые виджеты | `fe-error-empty-auth-race` |
| FE-P1-A11Y-CONTROLS | P1 | Nested interactive / нет accessible name в chrome | W9-A2-002, W9-A2-003, W9-A4-001 | `frontend/src/components/layout/sidebar-user.tsx:46` | button⊃Link / button⊃button → axe nested-interactive FAIL, клик/фокус ломают сценарий; collapsed `Link` без `aria-label` → name пустой у nav | `fe-a11y-nested-sidebar-notifications` |
| FE-P1-TMS-DOCS | P1 | TMS stale comparison + docs test-path drift | W10-A2-003, W11-A5-001, W11-A5-002, W11-A5-003, W11-A5-004, W11-A4-002, W11-A4-003 | `frontend/src/modules/tms/environment-comparison.tsx:224` | Смена env без сброса cache → таблица старого сравнения; вернуть в docs `./internal/...` как канон регрессии → команда расходится с `ci.yml` (0 тестов в internal); `architecture.md` main.go ~820 LOC / мёртвые line-refs → ложные якоря аудита | `fe-tms-compare-and-docs-testpath` |

### Покрытие классов (38 source ID)

| ClassID | n |
|---------|--:|
| FE-P1-API-CLIENT | 3 |
| FE-P1-CALLS-CHAT | 2 |
| FE-P1-CALLS-LIFECYCLE | 6 |
| FE-P1-MESSENGER | 6 |
| FE-P1-BOARD-DND | 2 |
| FE-P1-SPRINT-UI | 3 |
| FE-P1-TRASH-WEBHOOKS | 2 |
| FE-P1-ERROR-AS-EMPTY | 4 |
| FE-P1-A11Y-CONTROLS | 3 |
| FE-P1-TMS-DOCS | 7 |
| **Σ** | **38** |

---

## Требует уточнения

Находки **не** входят в классы выше: нет самодостаточного критерия мутации / нужен BE или решение PO.

| Source ID | Severity | Файл:строка | Почему не в класс | Чего не хватает |
|-----------|----------|-------------|-------------------|-----------------|
| W10-A3-001 | P1 | `frontend/src/modules/tracker/sprint-backlog.tsx:28` | UI обещает «без спринта», API/`IssueResponse` без `sprint_id` / фильтра | Подтвердить BE: `sprint_id=null` (или backlog endpoint) + поле в DTO; иначе FE-only guard неполон → кросс-задача be+fe |
| W10-A5-002 | P1 | `frontend/src/routes/_auth/trash.tsx:439` | Fake success «Удалить навсегда» без API | Есть ли admin permanent-delete для project trash и какая роль; иначе критерий = только убрать кнопку (UX), не «починить hard delete» |
| W11-A5-005 | P1 | `docs/user_needs_analysis_2026.md:23` | Claims «webcams = Production» vs TZ Phase 16 | Решение PO: правка статуса в docs **или** признать partial (конфликт с `BUSINESS.md` / W11-A4-010 P2) |
| W11-A5-006 | P1 | `docs/vision_and_scope.md` G-07 | G-07 «US-WIKI реализованы» ↔ KPI wiki «планируется» | **FIXED** TASK-105 (G-07 Partial; A2W11-5-001) |

---

## Инвентарь сырых P1 (трассировка)

### W9 — components (11)

| ID | Кратко | → |
|----|--------|---|
| W9-A1-001 | CallChat нет subscribe `call:` | FE-P1-CALLS-CHAT |
| W9-A1-002 | CallChat любой event с `call_id` | FE-P1-CALLS-CHAT |
| W9-A2-001 | Upcoming «Войти» no-op / startCall | FE-P1-CALLS-LIFECYCLE |
| W9-A2-002 | sidebar-user button⊃Link | FE-P1-A11Y-CONTROLS |
| W9-A2-003 | sidebar-link collapsed без aria-label | FE-P1-A11Y-CONTROLS |
| W9-A3-001 | thread parent `limit:1` | FE-P1-MESSENGER |
| W9-A3-002 | reply без id в send | FE-P1-MESSENGER |
| W9-A3-003 | voice-recorder onstop race | FE-P1-MESSENGER |
| W9-A4-001 | notification nested buttons | FE-P1-A11Y-CONTROLS |
| W9-A4-002 | activity-feed error→empty | FE-P1-ERROR-AS-EMPTY |
| W9-A5-001 | ErrorBoundary не reset на nav | FE-P1-ERROR-AS-EMPTY |

### W10 — routes/modules (14)

| ID | Кратко | → |
|----|--------|---|
| W10-A1-001 | create-call `projects[0]` | FE-P1-CALLS-LIFECYCLE |
| W10-A1-002 | create-call без onError | FE-P1-CALLS-LIFECYCLE |
| W10-A1-003 | ScheduledCalls вне Query | FE-P1-CALLS-LIFECYCLE |
| W10-A2-001 | import-export без JWT | FE-P1-API-CLIENT |
| W10-A2-002 | issue→messenger без channelId | FE-P1-MESSENGER |
| W10-A2-003 | env comparison stale cache | FE-P1-TMS-DOCS |
| W10-A3-001 | backlog без фильтра спринта | **уточнение** |
| W10-A3-002 | poker null→empty | FE-P1-SPRINT-UI |
| W10-A4-001 | DnD double moveIssue | FE-P1-BOARD-DND |
| W10-A4-002 | rank константа | FE-P1-BOARD-DND |
| W10-A4-003 | sprint detail вечный spinner | FE-P1-SPRINT-UI |
| W10-A5-001 | trash ≠ ListTrash | FE-P1-TRASH-WEBHOOKS |
| W10-A5-002 | permanent delete fake | **уточнение** |
| W10-A5-003 | webhooks `projectId=""` | FE-P1-TRASH-WEBHOOKS |

### W11 — api/stores/hooks/lib + docs (17)

| ID | Кратко | → |
|----|--------|---|
| W11-A1-001 | calendar raw fetch | FE-P1-API-CLIENT |
| W11-A1-002 | refresh recurse без retry cap | FE-P1-API-CLIENT |
| W11-A2-001 | LinkPreview field mismatch | FE-P1-MESSENGER |
| W11-A3-001 | use-auth late setUser | FE-P1-ERROR-AS-EMPTY |
| W11-A3-002 | checkActiveCall swallow | FE-P1-CALLS-LIFECYCLE |
| W11-A3-003 | dashboard-store no error | FE-P1-ERROR-AS-EMPTY |
| W11-A3-004 | endCall/leave fail→reset | FE-P1-CALLS-LIFECYCLE |
| W11-A3-005 | sprint-store error→[] | FE-P1-SPRINT-UI |
| W11-A4-001 | unread sticky after HDel | FE-P1-MESSENGER |
| W11-A4-002 | architecture main.go LOC | FE-P1-TMS-DOCS |
| W11-A4-003 | architecture middleware refs | FE-P1-TMS-DOCS |
| W11-A5-001 | TESTING.md `./internal/...` | FE-P1-TMS-DOCS |
| W11-A5-002 | TESTING.md inventory stale | FE-P1-TMS-DOCS |
| W11-A5-003 | TECHNICAL.md CI path | FE-P1-TMS-DOCS |
| W11-A5-004 | QUICKSTART.md internal | FE-P1-TMS-DOCS |
| W11-A5-005 | webcams Production claim | **уточнение** |
| W11-A5-006 | G-07 wiki drift | **FIXED** TASK-105 |

---

## Handoff

Следующий шаг стадии B: завести `TASK-*.md` по **10** slug’ам из таблицы классов (нумерация после свободного ID в `backlog/`), обновить `backlog/README.md`. Четвёрку из «уточнение» — отдельно после BE/PO.
