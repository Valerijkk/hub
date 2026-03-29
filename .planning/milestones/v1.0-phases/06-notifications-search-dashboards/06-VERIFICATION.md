---
phase: 06-notifications-search-dashboards
verified: 2026-03-21T12:00:00Z
status: passed
score: 5/5 must-haves verified
re_verification: false
human_verification:
  - test: "Нажать Ctrl+K в браузере"
    expected: "Открывается палитра команд поверх всего приложения"
    why_human: "Поведение горячей клавиши не проверяется статическим анализом"
  - test: "Получить задачу — проверить колокольчик"
    expected: "Badge с числом на иконке колокольчика увеличивается без перезагрузки страницы"
    why_human: "WS-доставка требует живого окружения с NATS"
  - test: "Перетащить виджет на персональном дашборде"
    expected: "Виджет перемещается, layout сохраняется, после перезагрузки восстанавливается"
    why_human: "drag-and-drop поведение нельзя верифицировать статически"
  - test: "Проверить Telegram-уведомление"
    expected: "При наступлении события бот отправляет отформатированное сообщение в Telegram"
    why_human: "Требует живой бот с токеном и аккаунт с telegram_id"
  - test: "Перейти в /admin/audit"
    expected: "Страница показывает журнал аудита с фильтрами; non-admin получает 403"
    why_human: "RBAC-проверка требует браузерного тестирования"
---

# Phase 6: Notifications, Search & Dashboards Verification Report

**Phase Goal:** Пользователи имеют единый центр уведомлений, могут мгновенно найти что угодно через глобальный поиск и видеть состояние проекта на дашбордах
**Verified:** 2026-03-21
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths (from Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|---------|
| 1 | Пользователь получает in-app уведомления (иконка колокольчика) о назначениях, @упоминаниях, сменах статуса и других триггерах | ✓ VERIFIED | `notification-bell.tsx` рендерит Bell + badge с `unreadCount`; consumer.go обрабатывает TypeAssigned/TypeMentioned/TypeStatusChanged/TypeCommentAdded/TypeBlockerAdded/TypeCallStarted/TypeRunCompleted; WS bridge подписан на `notification.user.>` |
| 2 | Пользователь может настраивать предпочтения уведомлений; тихие часы подавляют несрочные | ✓ VERIFIED | `service.go::isQuietHours` реализован с поддержкой midnight crossover и timezone; `repo.IsEnabled` проверяет per-type per-channel предпочтения; `notification-settings.tsx` рендерит checkbox-таблицу типов × каналов + поля тихих часов |
| 3 | Ctrl+K открывает глобальный поиск по задачам, тест-кейсам, пользователям, проектам, каналам, файлам | ✓ VERIFIED | `app-shell.tsx` регистрирует слушатель `keydown` на `Ctrl+K`/`Cmd+K` и вызывает `setOpen(true)`; `CommandPalette` монтируется в app-shell; `search/service.go::Search` выполняет параллельные запросы по 7 типам |
| 4 | Персональный дашборд показывает задачи, непрочитанные, звонки, прогоны, дедлайны с drag-and-drop виджетами | ✓ VERIFIED | `dashboard/service.go::GetPersonalDashboardData` агрегирует MyTasks/UpcomingCalls/RecentRuns/Deadlines/TimeToday; `WidgetGrid` использует `react-grid-layout` v2 с `onLayoutChange`; layout сохраняется через `saveLayoutDebounced` |
| 5 | Дашборды проекта и администратора показывают метрики; аудит-журнал фиксирует все действия | ✓ VERIFIED | `GetProjectStats` (Redis TTL 5min) и `GetAdminStats` реализованы в `dashboard/service.go`; `AuditConsumer` пишет в поток "AUDIT" из всех модулей; `AuditLogViewer` отображает с фильтрами |

**Score:** 5/5 truths verified

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `backend/ent/schema/notification.go` | Notification Ent schema | ✓ VERIFIED | Файл существует, содержит `type Notification struct` |
| `backend/ent/schema/notificationpreference.go` | NotificationPreference Ent schema | ✓ VERIFIED | Файл существует, содержит `type NotificationPreference struct` |
| `backend/internal/notification/consumer.go` | NATS JetStream durable consumer | ✓ VERIFIED | `func (c *Consumer) Start` реализован; создаёт stream "NOTIFICATIONS" + consumer "notification-dispatcher"; обрабатывает 7 типов событий |
| `backend/internal/notification/service.go` | Notification service с preference check и quiet hours | ✓ VERIFIED | `Dispatch`, `isQuietHours`, `publishToWS`, `sendTelegram` — все реализованы, не заглушки |
| `backend/internal/notification/handler.go` | 6 REST endpoints | ✓ VERIFIED | `RegisterRoutes` регистрирует GET/PATCH/PUT на `/api/notifications/*` |
| `backend/ent/schema/savedfilter.go` | SavedFilter Ent schema | ✓ VERIFIED | Файл существует (`ls` подтверждён) |
| `backend/internal/search/service.go` | Parallel PG tsvector + MongoDB $text search | ✓ VERIFIED | `Search` с `wg.Add/wg.Wait` для 7 горутин; `searchIssues` использует `plainto_tsquery('simple', $1)`; `searchMessages` использует `$text: {$search: query}` |
| `frontend/src/components/search/command-palette.tsx` | Ctrl+K command palette (cmdk) | ✓ VERIFIED | Использует `Command` из `cmdk`, `shouldFilter={false}`, debounce 300ms, группировка по типу, навигация через `useNavigate` |
| `backend/ent/schema/dashboardlayout.go` | DashboardLayout Ent schema | ✓ VERIFIED | Файл существует (`ls` подтверждён) |
| `backend/internal/dashboard/service.go` | Dashboard aggregation service | ✓ VERIFIED | `GetPersonalDashboardData` — реальные Ent-запросы; `GetProjectStats` — Redis cache 5min; `GetAdminStats` — системные метрики |
| `frontend/src/components/dashboard/widget-grid.tsx` | react-grid-layout wrapper | ✓ VERIFIED | Использует `GridLayout`, `useContainerWidth`, `onLayoutChange` → `saveLayoutDebounced`; реестр из 12 виджетов |
| `backend/ent/schema/auditlog.go` | AuditLog Ent schema | ✓ VERIFIED | Файл существует (`ls` подтверждён) |
| `backend/internal/audit/consumer.go` | NATS JetStream audit consumer | ✓ VERIFIED | `func (c *AuditConsumer) Start` реализован; поток "AUDIT" со всеми модульными subjects |
| `frontend/src/components/notifications/notification-bell.tsx` | Bell icon с badge | ✓ VERIFIED | `Bell` lucide-icon, badge с `unreadCount`, WS listener через `useRef` pattern |
| `frontend/src/components/admin/audit-log.tsx` | Admin audit log viewer | ✓ VERIFIED | `AuditLogViewer` экспортирован, подключён к маршруту `/admin/audit` |
| `docs/phase-06/business.md` | Business documentation (≥200 строк) | ✓ VERIFIED | 411 строк |
| `docs/phase-06/testing.md` | Test cases (≥300 строк) | ✓ VERIFIED | 798 строк, 44 тест-кейса по TC-{REQ_ID}-{NN} |
| `docs/phase-06/technical.md` | Technical architecture (≥300 строк) | ✓ VERIFIED | 847 строк |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `notification/consumer.go` | NATS JetStream (stream "NOTIFICATIONS") | `js.CreateOrUpdateStream` | ✓ WIRED | Subjects: `project.>`, `messenger.channel.>`, `tms.>`, `calls.channel.>` |
| `notification/consumer.go` | `ws/bridge.go` | NATS publish `notification.user.{userID}` | ✓ WIRED | `service.publishToWS` публикует на subject; bridge.go строка 80 подписана на `notification.user.>` и вызывает `SendToUser` |
| `notification-store.ts` | WS события | listener на `notification.new` | ✓ WIRED | `notification-bell.tsx` регистрирует WS handler через `addListener/removeListener` с `useRef` |
| `search/service.go` | PostgreSQL tsvector | `plainto_tsquery('simple', $1)` в raw SQL | ✓ WIRED | Используется в `searchIssues`, `searchTestCases`, `searchUsers`, `searchProjects`, `searchFiles` |
| `search/service.go` | MongoDB `$text` | `bson.M{"$text": bson.M{"$search": query}}` | ✓ WIRED | `searchMessages` использует `$text.$search` |
| `command-palette.tsx` | `api/search.ts` | `search(value)` с debounce 300ms | ✓ WIRED | `handleInputChange` → `search(value)` из store → `searchGlobal` API |
| `widget-grid.tsx` | `react-grid-layout` | `GridLayout` с `onLayoutChange` | ✓ WIRED | `onLayoutChange={handleLayoutChange}` → `saveLayoutDebounced` |
| `dashboard-store.ts` | `api/dashboard.ts` | `getLayout`/`saveLayout` | ✓ WIRED | `saveLayoutDebounced` вызывает `saveLayout` через store |
| `audit/consumer.go` | NATS JetStream (stream "AUDIT") | durable consumer "audit-logger" | ✓ WIRED | `CreateOrUpdateStream` "AUDIT" + `CreateOrUpdateConsumer` "audit-logger" |
| `audit/service.go` | `ent/schema/auditlog.go` | `r.client.AuditLog.Query()` | ✓ WIRED | `repository.go` строки 52+ используют `AuditLog.Query()` с фильтрами |
| `app-shell.tsx` | `command-palette.tsx` | Ctrl+K keydown listener | ✓ WIRED | `app-shell.tsx` импортирует и монтирует `CommandPalette`; keydown listener вызывает `setOpen` |
| `topbar.tsx` | `notification-bell.tsx` | import + render | ✓ WIRED | Строка 20 импортирует `NotificationBell`, строка 90 рендерит в topbar |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|---------|
| NOTF-01 | 06-01 | Триггеры: назначение, @упоминание, смена статуса, сообщение, звонок, тест-ран, блокировщик | ✓ SATISFIED | consumer.go обрабатывает 7 типов событий; все функции обработчиков реализованы |
| NOTF-02 | 06-01 | In-app колокольчик + Telegram бот | ✓ SATISFIED | WS bridge доставляет `notification.new`; `telegram.go` с telego + rate limiting 35ms |
| NOTF-03 | 06-01 | Настройки по типу, каналу, проекту | ✓ SATISFIED | `repo.UpsertPreference` с per-type/per-channel/per-project; `notification-settings.tsx` |
| NOTF-04 | 06-01 | Тихие часы; срочные проходят всегда | ✓ SATISFIED | `isQuietHours` с midnight crossover; `isUrgent` bypass в `Dispatch` |
| SRCH-01 | 06-02 | Сохраняемые и общие фильтры задач | ✓ SATISFIED | `SavedFilter` Ent-схема; CRUD в `search/service.go`; `saved-filters.tsx` компонент |
| SRCH-02 | 06-02 | Ctrl+K глобальный поиск по 6+ типам | ✓ SATISFIED | `SearchService.Search` охватывает issue, test_case, user, project, channel, file, message |
| DASH-01 | 06-03 | Персональный дашборд с drag-and-drop виджетами | ✓ SATISFIED | `GetPersonalDashboardData`; `WidgetGrid` с react-grid-layout; маршрут `/dashboard` |
| DASH-02 | 06-03 | Дашборд проекта с шаблонами Dev/QA/Manager | ✓ SATISFIED | `GetProjectStats` (Redis cache); `dashboard-templates.ts`; маршрут `/projects/$id/dashboard` |
| DASH-03 | 06-03 | Административный дашборд с системными метриками | ✓ SATISFIED | `GetAdminStats`; маршрут `/admin/dashboard` с role guard |
| AUDT-01 | 06-04 | Полный журнал аудита с фильтрами | ✓ SATISFIED | `AuditConsumer` + `AuditLogViewer` с фильтрами action/entity_type/user/project/date |
| AUDT-02 | 06-04 | Настраиваемые лимиты (файл, проекты, хранилище) | ✓ SATISFIED | `GetResourceLimits`/`UpdateResourceLimit` через SystemSetting; `resource-limits.tsx` |
| AUDT-03 | 06-04 | 30-дневная корзина с восстановлением | ✓ SATISFIED | `deleted_at` на Issue/TestCase/FileMetadata; `ListTrash`/`RestoreItem`/`PurgeExpired`; `trash-manager.tsx` |
| AUDT-04 | 06-04 | Деактивация пользователя, переназначение, GDPR удаление | ✓ SATISFIED | `DeactivateUser`/`ReassignTasks`/`DeleteUserGDPR` реализованы; `user-management.tsx` |
| UI-05 | 06-02 | Ctrl+K поиск, Ctrl+N новая задача | ✓ SATISFIED | `app-shell.tsx` регистрирует оба шортката в `useEffect` с cleanup |

---

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `frontend/src/components/notifications/notification-panel.tsx` | 98 | `TODO: навигация к сущности (issue, channel, testrun)` | ⚠️ Warning | Клик по уведомлению помечает его прочитанным, но не переходит к связанной сущности. Не блокирует базовую функциональность уведомлений. |

**Единственный найденный anti-pattern** — незавершённая навигация при клике по уведомлению. Это неполная фича (navigation deeplink), но не критично: уведомление создаётся, доставляется через WS, отображается с badge и панелью, может быть помечено прочитанным. Остальной pipeline полностью рабочий.

---

### Human Verification Required

#### 1. Ctrl+K Command Palette

**Test:** В браузере с запущенным приложением нажать Ctrl+K (или Cmd+K на macOS)
**Expected:** Открывается модальный оверлей с полем поиска "Поиск задач, кейсов, пользователей...", никакого browser-native поведения
**Why human:** Поведение горячей клавиши требует живого браузера

#### 2. Real-time Notification Bell

**Test:** Пользователь A назначает задачу на пользователя B; пользователь B наблюдает topbar
**Expected:** Badge на колокольчике у B появляется без перезагрузки страницы; toast-уведомление через sonner
**Why human:** WS-доставка требует живого NATS и запущенного backend

#### 3. Drag-and-drop Widget Grid

**Test:** На странице /dashboard перетащить виджет "Мои задачи" в другое место; перезагрузить страницу
**Expected:** Виджет остаётся на новом месте (layout сохранился в backend)
**Why human:** drag-and-drop UX требует браузерного тестирования

#### 4. Telegram Notification

**Test:** Настроить telegram_id в профиле; наступление события (назначение задачи)
**Expected:** Telegram-бот отправляет форматированное сообщение с MarkdownV2
**Why human:** Требует живого telegram bot token и аккаунта

#### 5. Admin Role Guard

**Test:** Member (не Admin) открывает /admin/audit
**Expected:** Получает 403 / редирект; Admin видит журнал с фильтрами
**Why human:** RBAC-поведение требует браузерного тестирования с разными ролями

---

### Summary

Все 5 наблюдаемых истин из Success Criteria фазы 6 подтверждены кодом.

**Notifications (NOTF-01..04):** Полный pipeline от NATS события до браузера реализован и связан. JetStream consumer "notification-dispatcher" обрабатывает 7 типов событий. WS bridge подписан на `notification.user.>`. Frontend bell + panel + settings — не заглушки, содержат реальную логику. Telegram sender с rate limiting реализован.

**Search (SRCH-01..02, UI-05):** tsvector + GIN индексы создаются через `EnsureSearchIndexes`. Параллельный поиск по 7 типам реализован с mutex-protected merge. CommandPalette с cmdk, debounce 300ms, навигация через TanStack Router. Ctrl+K / Ctrl+N шорткаты зарегистрированы в app-shell.

**Dashboards (DASH-01..03):** Все три типа дашбордов реализованы с реальной агрегацией данных. 12 виджетов в реестре. react-grid-layout v2 с drag-and-drop. Layout persistence через backend API. Redis cache 5min для проектного дашборда.

**Audit (AUDT-01..04):** JetStream consumer "audit-logger" пишет из всех модульных событий. Soft-delete на Issue/TestCase/FileMetadata с 30-дневной корзиной. Resource limits через SystemSetting. GDPR delete каскадно очищает данные.

**Единственное предупреждение:** TODO в notification-panel.tsx — отсутствует deeplink навигация при клике по уведомлению (только mark-as-read). Не блокирует цели фазы.

---

_Verified: 2026-03-21_
_Verifier: Claude (gsd-verifier)_
