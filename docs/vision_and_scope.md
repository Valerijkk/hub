# Hub — Видение и границы (Vision and Scope)

> Единая база знаний. Обновляется BA при изменении бизнес-требований.
> Детальная документация: `docs/BUSINESS.md`, планирование: `.planning/PROJECT.md`

---

## Проблема и ценность

**Проблема:** типичная команда 10–50 человек использует 4–6 разрозненных инструментов (Jira, Slack, TestRail/Allure TestOps, Google Meet, GitHub). Контекст теряется при переключении, traceability размазана, Allure TestOps стоит $30k+/год, SaaS создаёт vendor lock-in.

**Решение:** Hub — self-hosted единая платформа с глубокой интеграцией модулей, единым UX и сквозной трассируемостью задача → тест → баг → PR → уведомление.

**Целевая аудитория:** средние команды разработки и тестирования, организации с требованием контроля данных и минимального бюджета инфраструктуры.

---

## Бизнес-цели (SMART)

| ID | Цель | Измеримость | Срок |
|----|------|-------------|------|
| G-01 | Заменить 4–6 разрозненных инструментов (Jira, Slack, TestRail, Meet) одной платформой | 1 инстанс Hub покрывает трекер + мессенджер + TMS + звонки + дашборды | v1.0 ✓ |
| G-02 | Обеспечить сквозную traceability: задача → тест → баг → PR → уведомление | Матрица трассируемости TMS, автопривязка GitHub/GitLab, Allure CI | v1.0 ✓ |
| G-03 | Self-hosted без SaaS lock-in для команд 10–50 человек | Деплой Docker Compose (dev) + K3s (prod), бюджет инфраструктуры минимальный | v1.0 ✓ |
| G-04 | Production-ready безопасность и стабильность | 0 CRITICAL findings, unit ≥ 80% security-critical, e2e критические пути | v2.5 |
| G-05 | Pixel-perfect UX после редизайна v2.0 | 11 страниц по макетам (v2.0 ✓); WCAG 2.1 AA + Lighthouse ≥ 90 — цель v2.5, верификация pending | v2.0 partial |
| G-06 | Снизить стоимость владения TMS vs Allure TestOps | Собственный Allure-парсер + TMS без лицензии TestOps | v1.0 ✓ |

## KPI

| Метрика | Целевое значение | Текущий статус |
|---------|------------------|----------------|
| REST API (`/api/*`) | 280+ production routes | **286** ✅ (+3 infra, +1 WS `GET /ws`; см. `docs/wiki/API.md`) |
| Frontend UI-маршруты | Покрытие всех модулей | **60+** страниц ✅ (`frontend/src/routeTree.gen.ts`, v4 super-app) |
| Ent schemas (PostgreSQL) | Полная модель домена | **82** сущности ✅ (`backend/ent/schema/`, без `helpers.go`) |
| Backend domain-модули | Модульный монолит | **31** пакет в `backend/internal/` (+ `platform/`, `testutil/`) |
| v2.0 UI requirements (DSFN–XTRA) | 85/85 | ✅ COMPLETE (Phases 8–14) |
| CRITICAL security blockers | 0 | **12/12 FIXED** в коде (v2.0.1); PG-03 ротация секретов в git history — manual ops |
| Unit coverage (security-critical) | ≥ 80% | **6/6 пакетов ≥80%** ✅ (auth, ws, storage, rbac, audit, calls) |
| Unit coverage (overall `internal/`) | ≥ 50% (gate), ≥ 82% (wave) | **83.0%** ✅ (Wave C, Docker `-tags hubtest`) |
| Frontend Vitest | Unit-тесты модулей | **99/99** ✅ |
| E2E critical paths | Playwright | **20 spec-файлов** ✅ (`testframework/e2e/specs/`) |
| Global search Ctrl+K (US-M07) | UI + user-scope backend | **UI ✅** (PR #15: Calls, Settings, Admin) |
| Lighthouse Performance/Accessibility | ≥ 90 | CI artifact job (PR #27, non-blocking); formal gate — backlog |
| WCAG 2.1 AA | 4.5:1 text, 3:1 UI | Реализовано (THME-04), E2E a11y spec ✅ |
| DAU/WAU/MAU | Админ-дашборд | ✅ Реализовано |
| Участники голосового звонка | До 50 | ✅ LiveKit |
| CI budget (GitHub Actions) | ~3000 мин/мес | В рамках лимита |

## Портреты пользователей

| Персона | Роль | Потребности | Ключевые сценарии |
|---------|------|-------------|-------------------|
| **Алексей, Developer** | Backend/Frontend dev | Задачи, PR-линки, таймтрекинг, код-ревью в одном месте | Kanban, комментарии, @упоминания, созвоны, GitHub-ветка из задачи |
| **Мария, QA Lead** | Тестировщик | TMS, Allure CI, traceability, баги из прогонов | Тест-кейсы (classic/BDD), прогоны, матрица покрытия, flaky-отчёты |
| **Дмитрий, Team Lead** | Тимлид | Спринты, burndown, velocity, дашборды | Scrum, Planning Poker, ретроспектива, проектный дашборд |
| **Виктор, Project Owner** | Владелец проекта | Workflow, RBAC, инвайты, интеграции | Шаблоны проектов, кастомные роли, GitHub/GitLab, архивирование |
| **Елена, Admin** | Системный администратор | Пользователи, бэкапы, аудит, лимиты, регистрация | Админ-панель, audit log, GDPR, resource limits, режим инвайтов |
| **Игорь, Member** | Участник команды | Мессенджер, уведомления, быстрый поиск | Каналы, треды, Ctrl+K, Telegram bot |
| **Сергей, DevOps** | CI/CD инженер | Allure в CI, hub-cli, webhooks | `hub-cli push-results`, JUnit webhook, #hub-ci read-only канал |

## User Stories (MoSCoW)

### Must Have

| ID | История | Acceptance Criteria |
|----|---------|---------------------|
| US-M01 | Как пользователь, хочу войти через Magic Link (email) без пароля | JWT 15 мин + refresh 30 дней; rate limit 3/10 мин; TG только notify bind (ADR-AUTH-01) |
| US-M02 | Как участник, хочу создать проект с автоматическим трекером, TMS, чатом | Проект создаёт workflow, Kanban, канал #project-{key}, дашборд; ключ HUB/MOBILE для нумерации |
| US-M03 | Как разработчик, хочу управлять задачами (9 типов) с кастомным workflow | CRUD, связи (blocks/relates/duplicates/clones/causes), комментарии, optimistic lock, чеклисты |
| US-M04 | Как команда, хочу общаться в каналах с тредами и реакциями | Публичные/приватные каналы, DM, треды, presence, typing indicator |
| US-M05 | Как QA, хочу вести тест-кейсы и получать результаты из CI (Allure) | Парсер Allure JSON, матчинг по ID/annotation, уведомления, создание бага из failed |
| US-M06 | Как участник, хочу созвоны с screen share (и webcam с Phase 16) до 50 человек | LiveKit; запись audio+screen; webcam Phase 16 (ADR-TZ-REWRITE) |
| US-M07 | Как пользователь, хочу глобальный поиск (Ctrl+K) по всем сущностям | User-scope фильтрация; задачи, кейсы, пользователи, проекты, каналы, файлы; JQL для задач |
| US-M08 | Как новый пользователь, хочу пройти онбординг | Профиль → тур (6 карточек) → тема/accent → создание/вступление в проект |
| US-M09 | Как Admin, хочу управлять режимом регистрации | Открытая регистрация ИЛИ только по инвайтам — настройка в админ-панели |
| US-M10 | Как Project Owner, хочу RBAC внутри проекта | Фиксированные роли (Owner/Developer/Tester/Viewer) + кастомные с bitmask permissions |
| US-M11 | Как команда, хочу Kanban с WIP-лимитами и swimlanes | Колонки = статусы workflow; несколько досок; фильтры на доске |
| US-M12 | Как разработчик, хочу таймтрекинг задач | Estimate + факт; таймер Start/Stop; ручной ввод с описанием |
| US-M13 | Как Admin, хочу автоматические бэкапы и восстановление | Еженедельный бэкап PG + Mongo + SeaweedFS; restore через админ-панель |

### Should Have

| ID | История | Acceptance Criteria |
|----|---------|---------------------|
| US-S01 | Как тимлид, хочу Scrum со спринтами и burndown | Спринты, velocity, Planning Poker, ретроспектива, автоперенос незавершённых |
| US-S02 | Как пользователь, хочу настраиваемые дашборды (личный/проектный) | Drag & drop виджеты, шаблоны Dev/QA/Manager |
| US-S03 | Как Admin, хочу audit log и корзину 30 дней | Фильтры по модулям; восстановление; GDPR-удаление; resource limits |
| US-S04 | Как команда, хочу интеграцию с GitHub (PR, ветки, webhook) | Автопривязка HUB-123, статус PR, автосмена статуса при merge, ветка из задачи |
| US-S05 | Как пользователь, хочу уведомления в Hub, Telegram и Web Push | Quiet hours; срочные пробивают; настройка по типам/каналам/проектам |
| US-S06 | Как команда, хочу интеграцию с GitLab | MR автопривязка, pipelines, создание ветки из задачи |
| US-S07 | Как разработчик, хочу REST API с ключами и исходящими webhook | Swagger; API key в профиле (SHA-256 hash); delivery tracking с retry |
| US-S08 | Как DevOps, хочу hub-cli для CI/CD | `push-results`, `create-bug`, `update-task`; JUnit webhook |
| US-S09 | Как Admin, хочу импорт/экспорт данных | CSV/JSON задачи; Excel тест-кейсы; ZIP бэкап проекта; Jira import |
| US-S10 | Как Project Owner, хочу шаблоны и архивирование проектов | Шаблоны: пустой, Software Dev, QA, Kanban Simple; архив = read-only |
| US-S11 | Как Admin, хочу админ-дашборд | DAU/WAU/MAU, хранилище, здоровье сервисов, топ проектов/пользователей |
| US-S12 | Как участник, хочу расширенный мессенджер | Голосовые сообщения, поллы, pin/bookmark, read-only #hub-ci, link preview |
| US-S13 | Как QA, хочу полноценный TMS | BDD-кейсы, версионирование с откатом, окружения, матрица трассируемости |
| US-S14 | Как участник, хочу запланированные и гостевые звонки | Recurring стендап; guest join по invite; записи в разделе проекта |

### Could Have

| ID | История | Acceptance Criteria |
|----|---------|---------------------|
| US-C01 | Как пользователь, хочу PWA с offline-кешем | Установка на телефон; offline-кеш статики (Web Push — см. US-S05) |
| US-C02 | Как QA, хочу flaky-тесты и сравнение прогонов | Отчёты уровня Allure TestOps; сравнение окружений |
| US-C03 | Как пользователь, хочу кастомный accent color и ru/en | Переключатель в профиле; Ctrl+N новая задача |
| US-C04 | Как пользователь, хочу сохранённые и расшариваемые JQL-фильтры | Фильтры по статусу/приоритету/исполнителю/типу |
| US-C05 | Как пользователь, хочу мобильный UX | Bottom navigation; упрощённый Kanban; мобильные дашборды и созвоны |
| US-C06 | Как разработчик, хочу просмотр GitHub Workflow runs | Статус CI в контексте задачи/PR |

### Won't Have (v1–v2) — updated v2.5 PO lock

- ~~Видеозвонки с вебкамерами~~ → **Phase 16** (ADR-TZ-REWRITE)
- Нативные мобильные приложения (PWA достаточно)
- ~~Email только magic link~~ → transactional + marketing via **hub-mail** (v2.5+)
- SSO/LDAP/OAuth, SaaS billing (later), ~~Gantt~~ → Phase 16, multi-org (v5)
- Remote control (управление чужим экраном)
- Видео/аудио сообщения в мессенджере (кроме голосовых)
- Route-level page transition animations
- Parallax/3D-анимации на карточках

См. также: [tz_constraints.md](tz_constraints.md) · `.planning/DISCUSS-ROADMAP-LOCKED.md`

### v3.0 AI Chat Hub — SHIPPED (`.planning/MILESTONE-3.md`)

| ID | История | Статус |
|----|---------|--------|
| US-F01 | Как участник, хочу писать `@hub-ai` в каналах и DM | ✅ `backend/internal/ai/messenger_bot.go` |
| US-F02 | Как Admin, хочу выбирать AI-провайдера (GigaChat/YandexGPT/DeepSeek/Ollama/OpenRouter) | ✅ `ai/admin_handler`, Ent `AIProvider` |
| US-F03 | Как QA, хочу генерацию кейсов из задачи | ⚠️ Partial — AI composer suggest, не slash-команда |
| US-F04 | Как пользователь, хочу AI-панель и выбор модели | ✅ `/settings/ai`, `GET /api/ai/models` |
| US-F05 | Как Admin, хочу rate limits и audit AI без утечки содержимого | ✅ `AIUsageDaily`, content policy |
| US-F06 | Как разработчик, хочу OpenAI-compatible API Hub | ⏳ v3.5 backlog |

### v4.0 Corporate Super-App — SHIPPED (`.planning/PLATFORM-VISION-v4.md`)

| ID | История | Статус |
|----|---------|--------|
| US-V01 | Как сотрудник, хочу ленту новостей компании | ✅ `intranet/news`, Ent `NewsPost` |
| US-V02 | Как сотрудник, хочу каталог коллег и org chart | ✅ `/intranet/directory` |
| US-V03 | Как сотрудник, хочу единый календарь (созвоны, спринты, отпуск) | ✅ `/calendar`, `backend/internal/calendar/` |
| US-V04 | Как PO, хочу CRM-lite (контакты, сделки) | ✅ `/crm`, Ent `CRMContact`, `CRMDeal` |
| US-V05 | Как сотрудник, хочу подать заявку на отпуск | ✅ `/hr`, Ent `LeaveRequest` |
| US-V06 | Как Admin, хочу встроенную бухгалтерию (без 1C) | ✅ `/accounting`, journal entries, invoices |

**Дополнительно в v4 (код):** no-code DB (`/bases`, `HubBase/HubTable/HubRecord`), marketplace (`/marketplace`), kudos и random coffee (`intranet`), status page (`/status`), hub-mail (`/admin/mail`), 1C connector (`integrations/onec`).

**Won't Have v4:** payroll/НДФЛ calc, multi-tenant SaaS (v5), SSO, native apps.

## Сценарии использования

### Основной: создание бага из failed тест-кейса

1. CI запускает автотесты → Allure JSON → `hub-cli push-results` в TMS API
2. Парсер матчит результат с тест-кейсом HUB-TC-42
3. QA видит failed в прогоне → «Создать баг»
4. Создаётся задача HUB-123 типа Bug, связь с кейсом
5. Уведомление в канал #hub-bugs и Telegram ответственному

### Основной: рабочий день разработчика

1. Личный дашборд → назначенные задачи и дедлайны
2. Открывает HUB-42, Start Timer
3. Коммит «HUB-42 fix login» → автопривязка PR
4. Stop Timer → время в логе задачи

### Альтернативный: вход через Magic Link

1. Пользователь открывает /login → вводит email
2. Получает код (Resend / hub-mail) → verify → JWT
3. При необходимости — привязка Telegram в Settings для notify

*(Telegram login removed — ADR-AUTH-01)*

### Альтернативный: настройка нового проекта (Project Owner)

1. Создаёт проект «Mobile App», ключ MOB
2. Настраивает workflow: To Do → In Progress → Code Review → QA → Done
3. Генерирует invite-ссылку, подключает GitHub и Allure webhook

### Альтернативный: ежедневный стендап

1. Recurring звонок в #dev-standup
2. Screen share Kanban-доски
3. Запись по кнопке → ссылка в канал и раздел «Записи» проекта

### Исключительный: конфликт optimistic lock

1. Два пользователя редактируют задачу одновременно
2. Второй save получает 409 Conflict (version mismatch)
3. UI показывает diff и предлагает перезагрузить или перезаписать

## Организация (v2.5+)

| Правило | Реализация |
|---------|------------|
| Одна org на инстанс | Ent `Organization`; multi-org → v5 |
| Регистрация | Invite-only или open — настройка Admin (`/admin/registration-mode`) |
| Pending queue | Новые пользователи без invite → approve/reject; 7 дней без approve → delete (cron) |
| Роль `founder` | Ровно один на org (`User.system_role`: founder/admin/member) |
| Org invite | `GET /admin/organization/invite` — ссылка для первичного онбординга |

## Текущий статус проекта

| Milestone | Phases | Статус | Дата |
|-----------|--------|--------|------|
| v1.0 MVP | 1–7 | ✅ SHIPPED | 2026-03-21 |
| v2.0 Hub Redesign | 8–14 | ✅ COMPLETE | 2026-05-15 |
| v2.5 Trust | 15 | ✅ COMPLETE | 2026-06-22 |
| v2.6 Automate + Arch + OBS | 16–38 | ✅ COMPLETE | 2026-06-24 |
| v3.0 AI Native | 17–19 | ✅ COMPLETE | 2026-06-22 |
| v4 Super-app | 39–58 | ✅ COMPLETE | 2026-06-25 |
| Coverage wave | 59–125 | ✅ COMPLETE | 2026-06-21 |

**Текущий фокус:** Wave C ✅ · FA-06 manual — единственный prod-блокер · ветка `master`. См. `.planning/STATE.md` и `.planning/AUDIT-2026-06-23.md`.

### Отложенный polish (не блокирует релиз)

- PLSH-01: Per-page ambient body gradients
- PLSH-02: Skeleton loading states для виджетов
- PLSH-03: Smooth theme toggle transition
- Phase 16: Gantt UI (`/projects/$id/gantt`), webcam в звонках, hub-mail transactional

## Связанные документы

| Документ | Назначение |
|----------|------------|
| `docs/BUSINESS.md` | Детальное описание модулей и ролей |
| `docs/system_specification.md` | Техническая спецификация |
| `.planning/PROJECT.md` | Активные требования и ключевые решения |
| `.planning/REQUIREMENTS.md` | v2.0 redesign requirements (85 шт., complete) |
| `.planning/ROADMAP.md` | Фазы и success criteria |
| `.planning/STATE.md` | Текущая позиция и blockers |
| `.planning/MILESTONE-3.md` | v3.0 AI Chat Hub (30 требований) |

---
*Последнее обновление: 2026-06-21 — RU-D01: верификация по коду (backend/internal, frontend/src, Ent schema, COVERAGE.md)*
