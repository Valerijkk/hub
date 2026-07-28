# Hub

## Что это

Hub — единая платформа для команд разработки и тестирования (10-50 человек), объединяющая трекер задач (Kanban+Scrum), мессенджер с каналами и тредами, систему управления тестированием (TMS) с Allure-интеграцией и голосовые созвоны с демонстрацией экрана. Все модули связаны между собой: задачи привязаны к тест-кейсам, результаты прогонов автотестов из CI попадают в TMS, уведомления о событиях приходят в мессенджер и Telegram.

## Основная ценность

Единая среда для команды, где трекер задач, тестирование, общение и созвоны связаны в одном интерфейсе — без переключения между десятком разрозненных инструментов.

## Требования

### Подтверждённые

(Пока нет — нужно выпустить для валидации)

### Активные

> **C6 reconcile 2026-07-27:** чекбоксы сверены с кодом (`file:line` / путь модуля). История формулировок сохранена; где scope сменился — пометка в скобках. Источник FR-матрицы: `.planning/audits/planning-audit-2026-07/SPEC-COVERAGE.md`.

**Авторизация и пользователи:**
- [x] Magic Link (email) без паролей — `backend/internal/auth/handler.go:35-36,72,116`. ~~Telegram login~~ → Telegram только notify bind (ADR-AUTH-01), не вход
- [x] Два режима регистрации (open / invite-only) — `backend/internal/auth/dto.go` (reg mode) + invite schema `backend/ent/schema/invite.go`
- [x] Профиль (имя, аватар, email, Telegram bind, должность, отдел, телефон, TZ, язык, …) — `backend/internal/auth/service_profile.go`, `backend/ent/schema/user.go`
- [x] Роли системные: founder / admin / member — `backend/ent/schema/user.go:48-49` (исторически «Admin, Member» — founder добавлен v2.5+)
- [x] Множественные сессии: список + revoke — `backend/internal/auth/handler.go:45-46,291,331`; `service_session.go:278-284`
- [x] Онбординг профиля (complete) — `backend/internal/auth/service_profile.go:34-39`. Intranet onboarding-templates (FR-INTR-04..) — ещё нет (см. SPEC-COVERAGE)

**Проекты:**
- [x] Создание проектов + автоканал/связанные модули — `backend/internal/project/handler.go`; auto-channel `backend/internal/messenger/auto_channels.go`
- [x] Project Owner / RBAC участников, workflow, команды — `backend/internal/project/`, `backend/internal/rbac/`
- [x] Роли проекта + кастомные permissions — `backend/internal/rbac/`
- [x] Приглашения + teams — `backend/ent/schema/invite.go`, team membership edges on User
- [x] Шаблоны проектов — `backend/ent/schema/projecttemplate.go`; hooks `backend/internal/project/export_hooks.go`
- [x] Архивирование проекта — `backend/internal/project/handler.go` (Archive*)
- [x] Корзина 30д (project among trash types) — `backend/internal/audit/service.go:203-204,234-235`. Полный project ZIP backup — через backup/imex модули

**Трекер задач:**
- [x] 9 типов issue — `backend/ent/schema/issue.go:38-39`
- [x] Кастомные поля — `backend/ent/schema/customfielddefinition.go`
- [x] Parent/child вложенность — issue edges / tracker service
- [x] Связи issue (blocks, relates, …) — `backend/internal/tracker/dto.go` + `backend/ent/schema/issuelink.go`
- [x] Нумерация PREFIX-N — project key + issue counter (tracker/project)
- [x] Настраиваемый workflow / статусы — `backend/ent/schema/workflowstatus.go` (`wip_limit` at :34)
- [x] Комментарии / mentions / реакции — tracker comments module
- [x] Лог изменений — `backend/ent/schema/activitylog.go`
- [x] Таймтрекинг (estimate/spent/timer) — `backend/ent/schema/issue.go:53-57`; timer store/API
- [x] Метки + группы — `backend/ent/schema/label.go`, `labelgroup.go` (задачи + test cases)

**Kanban:**
- [x] Колонки = workflow statuses — board + workflowstatus
- [x] WIP лимиты — `backend/ent/schema/workflowstatus.go:34`
- [x] Swimlanes assignee/priority/type — `backend/ent/schema/board.go:29-31`
- [x] Несколько досок на проект — `backend/ent/schema/board.go` edge from Project
- [x] Фильтры на доске — `board.go:24-26` (`filters` JSON)

**Scrum:**
- [x] Спринты CRUD / backlog / start — `backend/internal/sprint/handler.go`
- [x] Burndown + Velocity — `backend/internal/sprint/service.go:176,240`; repository GetBurndown/GetVelocity
- [x] Sprint retrospective notes — `backend/internal/sprint/handler.go:34,249`; `UpdateRetroNotes`
- [x] Автоперенос незавершённых при complete — `backend/internal/sprint/service.go:120-138`
- [x] Planning Poker (скрытые голоса до reveal) — `backend/internal/sprint` PokerService; FE `frontend/src/modules/tracker/planning-poker.tsx`

**Фильтры и поиск:**
- [x] Сохранённые фильтры на проекте/доске — `backend/ent/schema/project.go` / board filters
- [x] Ctrl+K глобальный поиск — `frontend/src/components/search/command-palette.tsx:72-73` (UI 2026-06-20)
- [x] JQL-подобный язык — `backend/internal/search/jql.go`; FE `frontend/src/components/search/jql-editor.tsx`

**Мессенджер:**
- [x] DM + групповые чаты / каналы — `backend/internal/messenger/`
- [x] Публичные + приватные каналы — messenger channel types
- [x] Автоканалы проектов — `backend/internal/messenger/auto_channels.go`
- [x] Read-only каналы — `backend/internal/messenger/dto.go` (read_only)
- [x] Треды — `ListThreadReplies` in `backend/internal/messenger/repository.go:73-74`
- [x] Реакции — messenger reactions
- [x] Форматирование rich-text — Plate / message content
- [x] Edit/delete своих + moderation — messenger service/handlers
- [x] Пересылка, pin, bookmarks — `AddPin` / bookmarks in `repository.go:97,111`
- [x] Голосовые сообщения — `backend/internal/messenger/service_messages.go` (voice)
- [x] Превью ссылок (issue/PR/OG) — `backend/internal/github/preview.go`; platform OpenGraph helper
- [x] Поиск по сообщениям — messenger search handlers
- [x] Presence / typing — `backend/internal/messenger/handler_presence.go`; events presence

**TMS (Test Management System):**
- [x] Classic + BDD cases — `backend/internal/tms/dto.go`; `backend/ent/schema/testcase.go`
- [x] Папки/теги/версионирование — tms module + testcase schema
- [x] Связь с GitHub TestCase annotations / webhooks — github + tms integration paths
- [x] Прогоны manual/Allure/mixed — `backend/internal/tms/`
- [x] Вложения/комментарии к результатам — tms result attachments
- [x] Traceability / связь с трекером — `backend/internal/tms/dto.go` (traceability)
- [x] Окружения — `backend/ent/schema/environment.go`
- [x] Отчёты / Allure-oriented analytics — tms reports + Allure parser
- [x] Импорт кейсов (CSV/JSON/…) — `backend/internal/imex/`
- [x] Экспорт — imex / tms export paths

**Интеграция с Allure:**
- [x] CI Allure HTML (внешний) + ingest в Hub — documented pipeline
- [x] Парсер allure-results → TMS API — `backend/internal/tms/allure_parser.go`
- [x] Матчинг по ID/имени — `backend/internal/tms/allure_matcher.go`
- [x] История прогонов + Telegram notify path — tms + notification Telegram bind

**Созвоны:**
- [x] Voice + screen share (без webcam в текущем коде), LiveKit, лимит участников — `backend/internal/calls/`; NFR-03
- [x] Запланированные + RRULE recurring — `backend/ent/schema/scheduledcall.go:41-43`; `backend/internal/calls/scheduled.go:46`
- [x] Mute / hand raise / screen share controls — `backend/internal/calls/handler.go` (HandRaise); LiveKit FE
- [x] Запись через LiveKit Egress → S3-compatible (**SeaweedFS**, не MinIO) — `backend/internal/calls/recording.go:86-149`
- [x] Записи привязаны к project path `recordings/{projectID}/…` — `recording.go:117-118`
- [x] Чат звонка — `backend/internal/calls/chat.go`

**Дашборды:**
- [x] Личный / виджеты — `backend/internal/dashboard/`; FE dashboard routes
- [x] Проектный (burndown/velocity/…) — dashboard + sprint metrics
- [x] Админский KPI / audit surface — admin module + dashboard admin widgets

**Уведомления:**
- [x] Триггеры domain events → inbox — `backend/internal/notification/`
- [x] In-app + Telegram bot delivery — notification + telego bind (не login)
- [x] Prefs по типам/каналам/проектам — notification preferences schema/service
- [x] Quiet hours — `backend/internal/notification/service.go` / repository QuietHours
- [x] Web Push — `backend/internal/notification/push.go`

**Интеграция с GitHub:**
- [x] PR/commit preview cards — `backend/internal/github/preview.go`
- [x] Автопривязка issue key из коммита — github webhook handlers
- [x] Статус PR на задаче — github integration
- [ ] Автосмена статуса workflow при merge PR — **не подтверждено как полный auto-transition** (есть PR review path `pr_review.go`; оставить open)
- [x] Создание ветки из задачи — `backend/internal/github/dto.go` (CreateBranch)

**API и интеграции:**
- [x] REST API + Swagger — `backend/docs/docs.go`; 286+ endpoints
- [x] API ключи пользователя — `backend/internal/apigateway/handler.go:42-45,110`; schema `apikey.go`
- [x] Outbound webhooks — `backend/internal/apigateway/webhook.go` + handler CRUD
- [x] hub-cli — `cli/`

**Импорт/экспорт:**
- [x] CSV/JSON импорт задач + mapping — `backend/internal/imex/`
- [x] Импорт тест-кейсов — imex/tms
- [x] Экспорт / project backup ZIP — imex + `backend/internal/backup/`

**Аудит и безопасность:**
- [x] Audit log + admin filters — `backend/internal/audit/`
- [x] Resource limits (Admin) — audit GetResourceLimits / admin settings
- [x] Корзина 30д: issue / test_case / file / project — `backend/internal/audit/service.go:197-204,226-237` (newspost/wiki — gaps, SPEC-COVERAGE)
- [x] GDPR / user deactivation paths — `backend/internal/audit/handler.go` (GDPR handlers)

**Бэкапы:**
- [ ] Еженедельный автоматический cron (PG + Mongo + object storage) — **Admin trigger есть**, отдельный weekly scheduler в коде не найден; storage = SeaweedFS S3, не MinIO
- [x] Хранить последние N (cleanup) — `backend/internal/backup/service.go:363` (`cleanupOldBackups`)
- [x] Восстановление через Admin — `backend/internal/backup/handler.go:83` (`RestoreBackup`); routes `:22`

**UI/UX:**
- [x] Sidebar + topbar shell — `frontend/src/components/layout/app-shell.tsx`
- [x] Mobile bottom nav — `frontend/src/components/mobile-nav.tsx`
- [x] Dark / light / system + accent — theme tokens / settings (v2 redesign)
- [x] RU + EN — `frontend/src/components/language-switcher.tsx:1-28`
- [x] Hotkeys Ctrl+K (+ Ctrl+N где wired) — command-palette
- [x] PWA manifest + Web Push client — `frontend/public/manifest.webmanifest`; `frontend/src/lib/push-notifications.ts`
- [x] Error UX (toast / request_id / error pages) — platform middleware NFR-13 + FE error boundaries
- [x] Вложения / presigned storage — `backend/internal/storage/handler.go`
- [x] Пагинация offset tasks / cursor chat — NFR-14/15 в handlers

### За рамками проекта

> Исторический список v1 сохранён. Факт на 2026-07-27 в скобках.

- Видеозвонки с вебкамерами — v1: только голос + screen share *(webcam = Phase 16 / later; в текущем calls-коде webcam enable не найден)*
- Мобильные нативные приложения (App Store/Google Play) — PWA достаточно *(manifest есть)*
- Email уведомления — исторически «Telegram покрывает»; *(hub-mail / Resend для Magic Link — отдельно; marketing email — later)*
- Биллинг/подписки — внутренний инструмент, не SaaS
- SSO/LDAP/OAuth — v1 out; *(2026-07-03 scope одобрен, реализация только после ADR-AUTH-SSO)*
- Gantt-диаграммы — v1 out; *(FE route существует: `frontend/src/routes/_authenticated/projects/$projectId/gantt.tsx` — shipped later)*
- Видео/аудио сообщения в мессенджере (кроме голосовых) — избыточно для v1
- Remote control (управление чужим экраном) — сложная фича, не критична

## Контекст

Hub создаётся как внутренний инструмент для средних команд разработки (10-50 человек). Основная мотивация — объединить в одном месте трекер (Jira), мессенджер (Slack), TMS (Allure TestOps/TestRail) и созвоны (Google Meet) с глубокой интеграцией между модулями.

Ключевая ценность — traceability: задача связана с фичей, фича покрыта тест-кейсами, тест-кейсы привязаны к автотестам в GitHub, результаты прогонов из CI попадают в TMS с Allure-отчётами, уведомления приходят в мессенджер и Telegram.

Деплой: локально для тестирования (Docker Compose), продакшн на CloudForBox (K3s).

Бюджет: минимальный (self-hosted решения, бесплатные инструменты, студенческий GitHub пакет ~3000 мин CI).

## Ограничения

- **Бэкенд**: Go + Fiber — выбрано для производительности и конкурентности (горутины для real-time)
- **Фронтенд**: React 19 + TypeScript + Vite — знакомый стек, максимум экосистемы
- **Базы данных**: PostgreSQL (задачи, TMS, юзеры) + MongoDB (чаты, Allure-отчёты) — гибридный подход
- **ORM**: Ent + Atlas миграции — кодогенерация, типобезопасность, граф сущностей для иерархий
- **Real-time**: NATS + JetStream — pub/sub + фоновые задачи без дублирования Redis
- **Кеш**: Redis — только кеш, сессии, magic link токены, онлайн-статусы
- **Созвоны**: LiveKit (self-hosted) — Go-native, WebRTC из коробки, запись через Egress
- **Хранилище**: SeaweedFS (self-hosted S3) — бесплатно, MinIO архивирован Feb 2026, SeaweedFS S3-compatible замена
- **Telegram**: telego — 100% Bot API, кодогенерация, типобезопасность
- **CSS/UI**: Shadcn/ui + Tailwind CSS 4 — компоненты копируются, полный контроль
- **Редактор**: Plate (Shadcn-native rich-text) — бесплатный, @упоминания, slash-команды
- **DnD**: pragmatic-drag-and-drop + motion (ex framer-motion) — от Atlassian, tree adapter для иерархий
- **Роутинг**: TanStack Router — type-safe, file-based routes, единая экосистема с TanStack Query
- **WebSocket**: Fiber встроенный middleware — один стек с бэкендом, без смешивания fasthttp/net-http
- **S3 SDK**: aws-sdk-go-v2 — стандарт, миграция на любой S3-compatible провайдер
- **Email**: Resend — 3000 писем/мес бесплатно для Magic Link
- **Сессии**: JWT access (15 мин) + refresh в Redis (30 дней) — мгновенный отзыв сессий
- **Репозиторий**: Монорепо (backend/ + frontend/ + cli/ + deployments/)
- **Rich-text**: Plate JSON в БД + plain text колонка для поиска/уведомлений
- **API**: Без версии сейчас, /api/v1/ при появлении публичного API (Phase 7)
- **Кастомные поля**: текст, число, дата, чекбокс, select, multi-select, user, URL
- **Часовые пояса**: UTC в БД, фронтенд конвертирует (автоматически из браузера + переопределение в профиле)
- **Конкурентность**: Оптимистичная блокировка (version field) + автосохранение черновика
- **Rate limiting**: По уровням — auth 3/10мин, API 100/сек, файлы 10/мин, сообщения 60/мин
- **Retention**: Настраиваемый Admin (чат 2 года, Allure 1 год, записи 6 мес по умолчанию)
- **Orphaned файлы**: Cron ежесуточно, удаление непривязанных старше 24ч
- **Тенанси**: Single-tenant (один инстанс = одна команда), multi-tenant не нужен для v1
- **Статусы доставки сообщений**: Нет (как Slack), только unread counts
- **Групповые упоминания**: @all, @online, @role (по ролям проекта)
- **Кросс-проектные связи при удалении**: Связи удаляются + уведомление затронутым
- **Документация**: Код + тесты + Swagger в каждом плане; бизнес + тестовая + техническая дока — отдельный план в конце каждой фазы
- **Графики**: Tremor — Tailwind-native дашборд компоненты
- **Прокси**: Caddy — автоHTTPS, WebSocket из коробки, 15 строк конфига
- **CI/CD**: GitHub Actions — студенческий пакет ~3000 мин, self-hosted runner как fallback
- **Оркестрация**: Docker Compose (dev) + K3s (prod) — лёгкий Kubernetes, 500MB RAM
- **Мониторинг**: Slog + Loki + Prometheus + Grafana — логи + метрики в одном UI
- **Тестирование бэкенда**: testing + testify + testcontainers + mockery
- **Тестирование фронтенда**: Vitest + RTL + Playwright + MSW
- **Бюджет**: Минимальный — все self-hosted, бесплатные инструменты

## Ключевые решения

| Решение | Обоснование | Результат |
|---------|-------------|-----------|
| Go + Fiber вместо Python + FastAPI | Максимальная производительность, горутины для real-time (мессенджер, 50 участников звонка) | Ожидает проверки |
| PostgreSQL + MongoDB гибрид | PG для реляционных данных (задачи, TMS), Mongo для документных (чаты, Allure JSON) | Ожидает проверки |
| NATS вместо RabbitMQ | Легче (10MB RAM), Go-native, JetStream покрывает и pub/sub и очереди | Ожидает проверки |
| Redis только кеш (не pub/sub) | NATS взял на себя pub/sub и очереди, Redis остался для кеша/сессий | Ожидает проверки |
| Ent вместо GORM | Кодогенерация, граф сущностей (иерархия задач), Privacy layer для RBAC | Ожидает проверки |
| LiveKit вместо Jitsi/mediasoup | Go-native, React SDK, запись из коробки, один Docker контейнер | Ожидает проверки |
| MinIO → SeaweedFS (S3-compatible) | MinIO archived Feb 2026; SeaweedFS + aws-sdk-go-v2 | ✅ shipped |
| Caddy вместо Nginx | АвтоHTTPS, WebSocket автоматически, конфиг 15 строк vs 60+ | Ожидает проверки |
| Shadcn/ui вместо Ant Design/MUI | Код копируется (полный контроль), Radix UI (доступность), единый Tailwind стиль | Ожидает проверки |
| Plate вместо Tiptap | Shadcn-native, всё бесплатно (таблицы у Tiptap платные), @упоминания из коробки | Ожидает проверки |
| pragmatic-dnd вместо dnd-kit | От Atlassian (Jira/Trello), tree adapter из коробки, 0 лишних рендеров | Ожидает проверки |
| Tremor вместо Recharts | Дашборд-фреймворк (не просто графики), Tailwind-native, KPI карточки | Ожидает проверки |
| Magic Link + Telegram (без паролей) | Современный UX, не нужно помнить пароли, Telegram бот и для входа и для нотификаций | Ожидает проверки |
| K3s вместо Docker Compose на проде | Zero downtime deploy (звонки/чаты), auto-healing, масштабирование replicas | Ожидает проверки |
| Allure Report + свой парсер вместо TestOps | TestOps $30k/год, свой парсер бесплатен, данные в своей TMS | Ожидает проверки |
| telego вместо telebot/go-telegram-bot-api | 100% Bot API кодогенерация, полная типизация, middleware | Ожидает проверки |
| Resend вместо своего SMTP | Бесплатно 3000/мес, настройка 5 минут, DKIM/SPF из коробки | Ожидает проверки |
| JWT access + Redis refresh вместо opaque sessions | Мгновенный отзыв сессий, список активных сессий, 15 мин автономности при падении Redis | Ожидает проверки |
| Кастомные поля: таблица + JSONB | Метаданные в реляционной таблице, значения в JSONB — гибко, нет EAV оверхеда | Ожидает проверки |
| Оптимистичная блокировка + черновик | Не блокируем, не теряем данные, версия в БД для конфликтов | Ожидает проверки |
| TanStack Router вместо React Router | Type-safe search params, единая экосистема с TanStack Query | Ожидает проверки |
| Fiber WebSocket вместо coder/websocket | Один стек fasthttp, совместимость с Fiber middleware (auth, RBAC) | Ожидает проверки |
| aws-sdk-go-v2 вместо minio-go | Стандарт S3, MinIO мёртв, миграция на любой S3 без изменений кода | Ожидает проверки |
| Шаги тест-кейса в JSONB + таблица версий | Шаги = единое целое с кейсом, версионирование снимками, UX как Allure TestOps | Ожидает проверки |
| Параллельный поиск PG + MongoDB | PG tsvector + MongoDB text index, без отдельного поискового движка для 50 юзеров | Ожидает проверки |
| Retention policy настраиваемый Admin | Дефолты: чат 2 года, Allure 1 год, записи 6 мес, Admin меняет | Ожидает проверки |

## Current Milestone: v2.0 Hub Redesign

**Goal:** Pixel-perfect редизайн всего фронтенда Hub на основе 11 HTML-макетов из папки `редизайн/` — без ломания существующего функционала, без новых фич.

**Target features:**
- Новая дизайн-система (Inter + JetBrains Mono, CSS variables, тёмная палитра #08090B→#1E2028, радиусы 6/8/12/16px, 4 уровня теней, spring/bounce анимации)
- Light + Dark mode (макеты dark-only → light адаптируется под ту же систему)
- Редизайн 11 страниц по макетам: Login, Dashboard, Messenger, Kanban, Issue Detail, Calls, TMS, Settings, Admin, Onboarding, Trash
- Страницы без макетов (Projects list, Sprint management, JQL search, Project settings, Project dashboard) — привести к общему стилю
- User-chosen accent color (сохранить текущее поведение)
- Компоненты переписываются поштучно для максимального качества
- `/gsd:ui-review` после каждой фазы для контроля качества

**Constraints:**
- Функционал НЕ трогаем — только визуал
- Качество > скорость
- Эталон — HTML-макеты из `редизайн/`

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Последнее обновление: 2026-07-27 — C6: Active checkboxes reconciled to code evidence (planning-audit)*
