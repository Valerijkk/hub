# Фаза 6: Уведомления, поиск, дашборды и аудит -- Бизнес-документация

## Обзор фазы

Фаза 6 реализует сквозные функции платформы Hub, которые связывают воедино все ранее созданные модули (трекер, мессенджер, TMS, звонки). Четыре подсистемы фазы:

1. **Уведомления** -- оповещение пользователей о событиях из всех модулей через In-App (WebSocket) и Telegram-бот
2. **Глобальный поиск** -- полнотекстовый поиск по задачам, тест-кейсам, пользователям, проектам, каналам, файлам и сообщениям
3. **Дашборды** -- персональный, проектный и административный дашборды с настраиваемыми виджетами
4. **Аудит и администрирование** -- журнал аудита, ресурсные лимиты, корзина с восстановлением, управление пользователями (GDPR)

Фаза покрывает **14 требований**: NOTF-01..04, SRCH-01..02, DASH-01..03, AUDT-01..04 и UI-05.

---

## Требования и реализация

### NOTF-01: Триггеры уведомлений

**Описание:**
Система генерирует уведомления при ключевых событиях: назначение задачи, обновление статуса, @упоминание в комментарии, блокировка задачи, личное сообщение, начало звонка, завершение прогона тестов.

**Реализация:**
- JetStream durable consumer `notification-dispatcher` подписан на стримы: `project.>`, `messenger.channel.>`, `tms.>`, `calls.channel.>`
- Consumer обрабатывает типы событий: `assigned`, `status_changed`, `comment_added` (с извлечением @mentions), `blocker_added`, `message_received` (только DM), `call_started`, `run_completed`
- Для каждого события определяются получатели, формируется объект `Notification` и сохраняется в PostgreSQL через Ent ORM
- 2-секундное in-memory debounce окно (ключ: `userID:type:entityID`) предотвращает дублирование уведомлений при быстрых изменениях

**Ограничения и допущения:**
- Уведомления о сообщениях приходят только для DM (личных сообщений), не для каналов -- чтобы избежать шума
- Упоминания в каналах обрабатываются через событие `comment_added`, а не `message_received`
- Debounce работает в памяти -- при перезагрузке сервиса окно сбрасывается

**Модули:** `backend/internal/notification/consumer.go`, `backend/internal/notification/events.go`

---

### NOTF-02: Каналы доставки уведомлений

**Описание:**
Уведомления доставляются по двум каналам: внутри Hub (иконка колокольчика в topbar) и через Telegram-бот.

**Реализация:**
- **In-App:** Уведомление публикуется в NATS-субъект `notification.user.{userID}`, WS bridge маршрутизирует его через WebSocket в браузер пользователя с событием `notification.new`
- **Telegram:** Буферизованный канал (cap=100) с worker-горутиной, rate limiting 35мс между отправками. Формат MarkdownV2. Используется telego библиотека
- **Frontend:** Zustand store хранит `notifications[]`, `unreadCount`. Sonner toast при получении нового уведомления. NotificationBell с красным бейджом в topbar

**Ограничения:**
- Telegram-бот переиспользуется из модуля авторизации (один бот для входа и уведомлений)
- Если botToken пустой, TelegramSender создается как nil (nil-safe: Send() на nil sender -- no-op)
- Буфер очереди Telegram ограничен 100 сообщениями -- при переполнении новые отбрасываются

**Endpoints:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/notifications` | Список уведомлений (page, per_page, is_read) |
| GET | `/api/notifications/unread-count` | Количество непрочитанных |
| PATCH | `/api/notifications/:id/read` | Пометить прочитанным |
| PATCH | `/api/notifications/read-all` | Пометить все прочитанными |

**Модули:** `backend/internal/notification/telegram.go`, `backend/internal/notification/handler.go`, `frontend/src/components/notifications/`

---

### NOTF-03: Настройки уведомлений

**Описание:**
Пользователь может включить/отключить уведомления по типам (assigned, mentioned, status_changed и т.д.) и по каналам (in_app, telegram), в том числе для конкретного проекта.

**Реализация:**
- Ent-схема `NotificationPreference` с полями: `notification_type`, `channel`, `enabled`, `project_id` (optional) и unique constraint (user_id + notification_type + channel + project_id)
- Проверка предпочтений в `NotificationService.Dispatch()`: сначала ищется per-project настройка, затем глобальная. По умолчанию уведомления включены (если записи нет)
- Frontend: таблица чекбоксов 9 типов x 2 канала в компоненте `NotificationSettings`

**Endpoints:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/notifications/preferences` | Получить настройки |
| PUT | `/api/notifications/preferences` | Создать/обновить настройку |

**Модули:** `backend/ent/schema/notificationpreference.go`, `frontend/src/components/notifications/notification-settings.tsx`

---

### NOTF-04: Тихие часы

**Описание:**
Пользователь может настроить "тихие часы" (quiet hours), в течение которых несрочные уведомления не доставляются. Срочные (is_urgent=true) пробивают тихие часы.

**Реализация:**
- Поля `quiet_hours_start` и `quiet_hours_end` (формат HH:MM, nullable) в схеме User
- Алгоритм проверки учитывает timezone пользователя и перенос через полночь (start > end)
- Срочные уведомления (is_urgent=true) обходят проверку тихих часов
- Проверка выполняется на уровне `NotificationService` перед dispatch

**Ограничения:**
- Часовой пояс берется из профиля пользователя; если не задан -- используется UTC
- Формат только HH:MM, без поддержки дней недели

**Модули:** `backend/internal/notification/service.go`, `backend/ent/schema/user.go`

---

### SRCH-01: Сохраненные и расшариваемые фильтры

**Описание:**
Пользователь может сохранять фильтры трекера с именем и конфигурацией, а также расшаривать их для всех участников проекта.

**Реализация:**
- Ent-схема `SavedFilter` с полями: `name`, `filter_config` (JSON), `shared` (bool), edges к `owner` (User) и `project` (Project)
- CRUD операции: создание, получение (свои + расшаренные проекта), удаление (только владелец), переключение shared
- Frontend: компонент `SavedFilters` с UI для создания, применения и удаления фильтров

**Endpoints:**
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/filters` | Создать фильтр |
| GET | `/api/filters?project_id=` | Список фильтров (свои + shared) |
| DELETE | `/api/filters/:id` | Удалить фильтр (владелец) |
| PATCH | `/api/filters/:id/share` | Переключить расшаривание |

**Модули:** `backend/ent/schema/savedfilter.go`, `backend/internal/search/service.go`, `frontend/src/components/search/saved-filters.tsx`

---

### SRCH-02: Глобальный поиск

**Описание:**
Глобальный поиск по всем сущностям платформы: задачи (issues), тест-кейсы (test_cases), пользователи (users), проекты (projects), каналы (channels), файлы (files) и сообщения (messages).

**Реализация:**
- **PostgreSQL (5 таблиц):** tsvector GENERATED ALWAYS AS ... STORED колонка + GIN индекс на таблицах issues, test_cases, users, projects, file_metadata. Конфигурация `'simple'` (language-agnostic, работает для русского и английского)
- **MongoDB:** $text индекс на коллекции messages (поле `content_text`)
- **Параллельный поиск:** горутины для каждого типа сущности с `sync.Mutex`-protected агрегацией результатов
- **Безопасность:** WHERE clause фильтрует по project membership (кроме users -- глобально доступны)
- **Ранжирование:** `ts_rank()` для PostgreSQL, `textScore` для MongoDB (x0.8 коэффициент), сортировка по убыванию score
- Каналы ищутся через ILIKE (нет tsvector)

**Endpoint:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/search?q=...&type=issue,test_case&limit=20` | Глобальный поиск |

**Модули:** `backend/internal/search/service.go`, `backend/internal/search/indexer.go`, `backend/internal/search/handler.go`

---

### UI-05: Горячие клавиши

**Описание:**
Ctrl+K открывает глобальный поиск (command palette), Ctrl+N создает новую задачу.

**Реализация:**
- **Ctrl+K:** Command palette на базе библиотеки cmdk. `shouldFilter=false` (серверная фильтрация). Группировка результатов по типу, иконки по типу сущности, localStorage для недавних поисков
- **Ctrl+N:** Dispatch кастомного события `hub:create-issue`, на которое подписан диалог создания задачи
- Регистрация хоткеев в `AppShell` через `useEffect` с cleanup

**Модули:** `frontend/src/components/search/command-palette.tsx`, `frontend/src/components/layout/app-shell.tsx`

---

### DASH-01: Персональный дашборд

**Описание:**
Персональный дашборд отображает данные текущего пользователя: мои задачи, непрочитанные сообщения, предстоящие звонки, дедлайны, таймтрекинг. Виджеты настраиваются drag-and-drop.

**Реализация:**
- Backend: `DashboardService.GetPersonalDashboardData()` агрегирует данные из трекера, мессенджера, звонков и таймлогов
- Frontend: react-grid-layout v2 с `useContainerWidth` hook. Виджеты: MyTasks, UnreadMessages, UpcomingCalls, Deadlines, TimeTracking
- Layout сохраняется в Ent-схеме `DashboardLayout` (JSON) с debounce 500мс после drag/resize

**Endpoint:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/dashboard/personal` | Данные персонального дашборда |
| GET | `/api/dashboard/layout?type=personal` | Получить раскладку |
| PUT | `/api/dashboard/layout` | Сохранить раскладку |

**Модули:** `backend/internal/dashboard/service.go`, `frontend/src/components/dashboard/widget-grid.tsx`, `frontend/src/routes/_authenticated/dashboard.tsx`

---

### DASH-02: Проектный дашборд

**Описание:**
Дашборд проекта показывает статистику: распределение задач по статусам, burndown, velocity, тренды тестов, активность команды. Доступны шаблоны (Dev, QA, Manager).

**Реализация:**
- Backend: `DashboardService.GetProjectStats()` с Redis-кешем (TTL 5 минут). Агрегирует: статусы задач, burndown данные, velocity по спринтам, тренды тест-прогонов, активность команды
- Frontend: виджеты TaskStatusChart (PieChart), BurndownChart (AreaChart), VelocityChart (BarChart), TestTrends, TeamActivity
- Шаблоны: JSON-пресеты `projectDev`, `projectQA`, `projectManager` с преднастроенной раскладкой виджетов

**Endpoint:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/dashboard/project/:projectId` | Статистика проекта |

**Модули:** `backend/internal/dashboard/service.go`, `frontend/src/components/dashboard/dashboard-templates.ts`, `frontend/src/routes/_authenticated/projects/$projectId/dashboard.tsx`

---

### DASH-03: Административный дашборд

**Описание:**
Дашборд администратора: количество пользователей, проектов, задач, DAU/WAU/MAU, использование хранилища, здоровье сервисов. Доступен только роли Admin.

**Реализация:**
- Backend: `DashboardService.GetAdminStats()` агрегирует данные из всех модулей. Проверка роли через `c.Locals("systemRole")`
- Frontend: KPI-карточки (пользователи, проекты, задачи) + виджеты SystemHealth, StorageUsage
- Guard: `beforeLoad` проверяет роль admin, редирект при отсутствии прав

**Endpoint:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/dashboard/admin` | Статистика администратора (403 для не-admin) |

**Модули:** `backend/internal/dashboard/handler.go`, `frontend/src/routes/_authenticated/admin/dashboard.tsx`

---

### AUDT-01: Журнал аудита

**Описание:**
Полный журнал аудита: все действия пользователей (логин/логаут, CRUD задач/проектов/кейсов, мессенджер). Администратор просматривает с фильтрами по действию, типу сущности, пользователю, проекту, дате.

**Реализация:**
- Ent-схема `AuditLog` с полями: `action`, `entity_type`, `entity_id`, `user_id`, `project_id`, `details` (JSON), `ip_address`, `user_agent`, `created_at`. Индексы на action, entity_type, user_id, project_id, created_at
- JetStream durable consumer `audit-logger` на стриме AUDIT, потребляющий события от всех модулей (project, tracker, messenger, tms, calls, auth)
- Пагинированный API с фильтрами
- Frontend: AuditLogViewer с CSV-экспортом и 30-секундным auto-refresh

**Endpoint:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/admin/audit` | Список записей аудита (admin-only) |

**Модули:** `backend/ent/schema/auditlog.go`, `backend/internal/audit/consumer.go`, `backend/internal/audit/repository.go`, `frontend/src/components/admin/audit-log.tsx`

---

### AUDT-02: Ресурсные лимиты

**Описание:**
Администратор настраивает лимиты: максимальный размер файла, количество проектов на пользователя, объем хранилища на проект.

**Реализация:**
- Лимиты хранятся в таблице `SystemSetting` (key-value). Ключи: `limit.file.max_size_mb`, `limit.user.max_projects`, `limit.project.max_storage_mb`
- Enforcement points: проверка лимитов при загрузке файлов и создании проектов
- Frontend: ResourceLimits -- числовые поля с кнопкой сохранения

**Endpoints:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/admin/limits` | Получить текущие лимиты |
| PUT | `/api/admin/limits/:key` | Обновить лимит |

**Модули:** `backend/internal/audit/service.go`, `frontend/src/components/admin/resource-limits.tsx`

---

### AUDT-03: Корзина (soft-delete) и восстановление

**Описание:**
Удаленные задачи, тест-кейсы и файлы перемещаются в корзину на 30 дней, после чего автоматически очищаются. Администратор или владелец проекта может восстановить элемент.

**Реализация:**
- Soft delete через поле `deleted_at` (nullable Time) на схемах Issue, TestCase, FileMetadata
- Все запросы автоматически фильтруют `DeletedAtIsNil()` -- удаленные сущности скрыты из обычных выборок
- Cron-горутина `PurgeExpired` запускается каждые 24 часа, удаляет записи старше 30 дней
- Frontend: TrashManager с вкладками (issues/test-cases/files/projects), индикатор оставшихся дней, кнопка восстановления

**Endpoints:**
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/api/admin/trash?type=issue` | Список корзины |
| POST | `/api/admin/trash/:type/:id/restore` | Восстановить элемент |

**Модули:** `backend/internal/audit/service.go`, `frontend/src/components/admin/trash-manager.tsx`

---

### AUDT-04: Управление пользователями и GDPR

**Описание:**
Администратор может деактивировать пользователя (запретить вход), переназначить его задачи другому пользователю и полностью удалить данные пользователя (GDPR-удаление).

**Реализация:**
- **Деактивация:** поле `status` (enum: active/deactivated) в User schema. Auth middleware проверяет `user.Status` и отклоняет деактивированных
- **Переназначение:** bulk-обновление assignee у незакрытых задач на другого пользователя
- **GDPR-удаление:** каскадное удаление из PostgreSQL (clear assignee -> delete sessions -> delete user record) + анонимизация в MongoDB
- Frontend: UserManagement с кнопками деактивации, диалогом переназначения и мульти-шаговым подтверждением GDPR-удаления

**Endpoints:**
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/admin/users/:id/deactivate` | Деактивировать пользователя |
| POST | `/api/admin/users/:id/reassign` | Переназначить задачи |
| DELETE | `/api/admin/users/:id/gdpr` | GDPR-удаление данных |

**Модули:** `backend/internal/audit/service.go`, `backend/internal/auth/middleware.go`, `frontend/src/components/admin/user-management.tsx`

---

## Пользовательские сценарии

### Уведомления

**Сценарий 1: Получение уведомления о назначении задачи**
1. Менеджер назначает задачу HUB-42 разработчику Иванову
2. Трекер публикует событие `project.{projectID}.assigned` в NATS
3. Consumer создает уведомление для Иванова с заголовком "Вам назначена задача HUB-42"
4. Уведомление доставляется через WebSocket (колокольчик) и Telegram
5. Иванов видит красный бейдж на колокольчике, нажимает -- видит уведомление в панели
6. Нажатие на уведомление открывает задачу, уведомление помечается прочитанным

**Сценарий 2: Тихие часы**
1. Тестировщик Петрова установила тихие часы 22:00-08:00
2. В 23:00 трекер обновляет статус задачи -- уведомление не доставляется
3. В 23:30 приходит срочное уведомление (is_urgent=true) о блокировке -- пробивает тихие часы
4. Утром после 08:00 Петрова видит обычные уведомления в панели (сохранены в БД, но не были отправлены)

### Поиск

**Сценарий: Быстрый поиск через Ctrl+K**
1. Разработчик нажимает Ctrl+K в любом месте приложения
2. Открывается command palette с фокусом на поле ввода
3. Вводит "логин" -- через 300мс debounce отправляется запрос `/api/search?q=логин`
4. Результаты группируются: задачи (HUB-5: Исправить логин), тест-кейсы (TC-AUTH-01-01), сообщения
5. Клик на результат переходит на страницу сущности, palette закрывается

### Дашборды

**Сценарий: Настройка персонального дашборда**
1. Пользователь переходит на /dashboard
2. Если layout не сохранен -- применяется шаблон personalDefault
3. Пользователь перетаскивает виджет "Мои задачи" в верхний левый угол
4. Через 500мс после отпускания -- layout сохраняется через PUT /api/dashboard/layout
5. При следующем входе раскладка восстанавливается из базы

### Аудит

**Сценарий: Расследование инцидента**
1. Администратор открывает /admin/audit
2. Ставит фильтр: действие="delete", тип="issue", дата: последние 7 дней
3. Видит кто удалял задачи, когда, из какого IP
4. Переходит в /admin/trash, находит удаленную задачу
5. Нажимает "Восстановить" -- задача возвращается в проект

---

## Диаграмма потока уведомлений

```
Событие в модуле              NATS JetStream                Consumer                     Доставка
+------------------+     +---------------------+     +--------------------+     +-------------------+
| Трекер           |     |                     |     |                    |     | In-App (WS)       |
| - assigned       |---->| Stream:             |     | notification-      |---->| notification.     |
| - status_changed |     |   NOTIFICATIONS     |---->| dispatcher         |     | user.{userID}     |
| - comment_added  |     |                     |     |                    |     +-------------------+
+------------------+     | Subjects:           |     | 1. Parse event     |
                         |   project.>          |     | 2. Debounce check  |     +-------------------+
+------------------+     |   messenger.         |     | 3. Preferences     |     | Telegram Bot      |
| Мессенджер       |---->|     channel.>        |     | 4. Quiet hours     |---->| telego + rate     |
| - message_recv   |     |   tms.>              |     | 5. Save to DB      |     | limit 35ms        |
+------------------+     |   calls.channel.>    |     | 6. Dispatch        |     +-------------------+
                         +---------------------+     +--------------------+
+------------------+
| TMS              |
| - run_completed  |
+------------------+

+------------------+
| Звонки           |
| - call_started   |
+------------------+
```

---

## Интеграции между модулями

| Подсистема фазы 6 | Трекер (Phase 2) | Мессенджер (Phase 3) | TMS (Phase 4) | Звонки (Phase 5) |
|--------------------|------------------|----------------------|---------------|------------------|
| **Уведомления** | assigned, status_changed, comment, blocker | DM message_received | run_completed | call_started |
| **Поиск** | issues (tsvector) | messages (MongoDB $text) | test_cases (tsvector) | -- |
| **Дашборды** | задачи, статусы, burndown, velocity | unread count | тренды прогонов | активные/последние звонки |
| **Аудит** | soft-delete задач, лог изменений | -- | soft-delete кейсов | -- |

---

## RBAC-матрица фазы 6

| Функция | Admin | Member/Developer/Tester | Viewer |
|---------|-------|------------------------|--------|
| Получать уведомления | Да | Да | Да |
| Настраивать уведомления | Да | Да | Да |
| Глобальный поиск | Да | Да (только свои проекты) | Да (только свои проекты) |
| Сохранять фильтры | Да | Да | Нет |
| Персональный дашборд | Да | Да | Да |
| Проектный дашборд | Да | Да | Да |
| Админ-дашборд | Да | Нет (403) | Нет (403) |
| Журнал аудита | Да | Нет | Нет |
| Управление лимитами | Да | Нет | Нет |
| Корзина/восстановление | Да | Нет | Нет |
| Деактивация пользователей | Да | Нет | Нет |
| GDPR-удаление | Да | Нет | Нет |

---

*Фаза: 06-notifications-search-dashboards*
*Дата документации: 2026-03-21*
