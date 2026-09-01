# Hub — Техническая документация

> **Обновлено:** 2026-06-22 · Стек: Go 1.25, React 19, Fiber v3 · Auth: Magic Link only  
> Wiki: `docs/wiki/API.md` · Deploy: `docs/wiki/DEPLOY.md` · Tests: `docs/test_plan.md`
>
> **Current QA authority:** CI-строки ниже — historical inventory, не current PASS.
> Единственный QA API — `docker compose -f docker-compose.local.yml run --rm <exact-service>`.

---

## CI и контроль качества

| Target | Authority | Current status |
|-----|-------------------|---|
| `backend-unit` | root Compose exact service | PASS: build/vet/unit/race |
| `frontend-unit` | root Compose exact service | PASS: 338/338 suites, 419/419 tests |
| python-contract | root Compose exact service | PASS: 181/181 framework selftests |
| Python API/integration/E2E/load | root Compose exact services | NOT_RUN; outside minimal merge gate |

Historical coverage baseline `internal/`: **~82%** (`.planning/COVERAGE.md`);
это не текущий migration/release gate.

---

## Архитектура

### Диаграмма системы

```
                          +-----------+
                          |   Caddy   |  :80/:443
                          |  (proxy)  |
                          +-----+-----+
                                |
             +------------------+------------------+
             |                  |                  |
        /api/*              /ws             (static)
             |                  |                  |
      +------+------+   +------+------+   +-------+------+
      |   Backend   |   |  WebSocket  |   |   Frontend   |
      |  Go/Fiber   |   |   Handler   |   |  React/Vite  |
      |   :3000     |   |   :3000     |   |   :5173      |
      +------+------+   +------+------+   +--------------+
             |                  |
    +--------+--------+--------+--------+
    |        |        |        |        |
+---+---+ +--+--+ +--+--+ +---+---+ +--+---+
| PG 17 | |Mongo| |Redis| | NATS  | |SeedFS|
| :5432 | |  8  | | 7.4 | | 2.11  | | :8333|
+-------+ |:2701| |:6379| |:4222  | +------+
          +-----+ +-----+ +---+---+
                               |
                          +----+----+
                          |LiveKit  |
                          | 1.8     |
                          | :7880   |
                          +---------+
```

### Диаграмма модулей

```
backend/internal/
  +-- apigateway/    API Key аутентификация
  +-- auth/          Авторизация (Magic Link + JWT/сессии; Telegram — notify bind, не login)
  +-- project/       Проекты, участники, invite, команды
  +-- tracker/       Задачи, workflow, комментарии, метки, кастомные поля, доски
  +-- sprint/        Спринты, burndown, velocity, planning poker
  +-- messenger/     Каналы, сообщения, реакции, поллы, присутствие
  +-- tms/           Тест-кейсы, прогоны, результаты, Allure, отчёты
  +-- calls/         Звонки, LiveKit, записи, запланированные звонки
  +-- dashboard/     Личный/проектный/админский дашборды
  +-- notification/  Уведомления (in-app, Telegram, Web Push)
  +-- search/        Глобальный поиск, JQL, сохранённые фильтры
  +-- github/        GitHub интеграция (webhook, PR, branches)
  +-- gitlab/        GitLab интеграция (webhook, MR, pipelines)
  +-- storage/       Файлы (SeaweedFS S3, presigned URLs)
  +-- backup/        Бэкапы (PG + Mongo + S3)
  +-- audit/         Аудит-лог, системные лимиты, корзина, GDPR
  +-- rbac/          Кастомные роли и permissions
  +-- imex/          Импорт/Экспорт (CSV, JSON, ZIP, PDF)
  +-- ws/            WebSocket Hub, Bridge (NATS -> WS)
  +-- ai/            AI gateway, bot, JQL assist, BYOK (v3)
  +-- crm/           Контакты, сделки, pipeline (v4)
  +-- intranet/      Новости, справочник, kudos, random coffee (v4)
  +-- bases/         No-code DB: HubBase/Table/Record (v4)
  +-- accounting/    Бухгалтерия: ledger, journal, invoices (v4)
  +-- marketplace/   Внутренняя доска объявлений (v4)
  +-- calendar/      Единый календарь (v4)
  +-- hr/            Заявки на отсутствие (v4)
  +-- dev/           Dev-эндпоинты (seed, dev-login)
  +-- platform/      Общие компоненты:
      +-- config/      Конфигурация (env vars)
      +-- middleware/  Rate limiting, CORS, auth middleware
      +-- errors/      Стандартизированные ошибки
      +-- http/        HTTP-утилиты
      +-- nats/        NATS-клиент
```

### Потоки данных

```
[Клиент] --HTTP--> [Caddy] --proxy--> [Fiber Handler]
                                            |
                                   [Service Layer]
                                      |        |
                              [Ent/PG]    [MongoDB]
                                      |
                              [NATS Publish]
                                      |
                         [WS Bridge] ---> [WebSocket клиенты]
                                      |
                         [Telegram Bot] ---> [Telegram]
                                      |
                         [Web Push] ---> [Браузер]
```

### Стек технологий

| Компонент | Технология | Версия (факт) |
|-----------|-----------|---------------|
| Language | Go | **1.25** (`go.mod`) |
| Web Framework | Fiber | **v3.0.0-beta.3** |
| ORM | Ent | **v0.14.1** |
| Migrations | Atlas | via Ent |
| Frontend | React | **19.1** |
| TypeScript | TypeScript | **5.7** |
| Bundler | Vite | **6.x** |
| Router | TanStack Router | **1.x** |
| State | Zustand + TanStack Query | **5.x** |
| UI | Shadcn/ui + Tailwind | **4.x** |
| PostgreSQL | relational | **17** |
| MongoDB | document | **8.0** |
| Redis | cache/sessions | **7.4-alpine** |
| NATS JetStream | pub/sub | **2.11** |
| Object Storage | SeaweedFS | **3.97** |
| WebRTC | LiveKit | **1.8** |
| Email | Resend / hub-mail | — |
| Telegram | telego (notify) | **v0.31.0** |
| Proxy | Caddy | **2-alpine** |
| S3 SDK | aws-sdk-go-v2 | **v1.32.7** |

---

## База данных

### PostgreSQL — Таблицы (Ent Schema)

#### users

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK, auto-generated |
| email | STRING | Уникальный, not empty |
| name | STRING | Опциональный |
| display_name | STRING | Опциональный |
| avatar_url | STRING | Опциональный |
| telegram_id | STRING | Уникальный, nullable |
| phone | STRING | Опциональный |
| position | STRING | Должность |
| department | STRING | Отдел |
| timezone | STRING | По умолчанию "UTC" |
| language | STRING | По умолчанию "ru" |
| birthday | TIME | Nullable |
| system_role | ENUM(admin, member) | По умолчанию "member" |
| status | ENUM(active, deactivated) | По умолчанию "active" |
| onboarding_completed | BOOL | По умолчанию false |
| quiet_hours_start | STRING | Nullable, формат HH:MM |
| quiet_hours_end | STRING | Nullable, формат HH:MM |
| custom_fields | JSONB | Кастомные поля профиля |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** telegram_id
**Связи:** sessions, project_members, team_members, files, invites, comments, reactions, activity_logs, time_logs, created_channels, channel_memberships, bookmarks, created_test_cases, created_runs, executed_results, test_case_versions, created_calls, call_participations, saved_filters, notifications, notification_preferences, dashboard_layouts, api_keys, recordings, scheduled_calls, project_templates, polls, poll_votes, push_subscriptions, channel_invites, blocks_created, blocks_received

#### sessions

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| refresh_token | STRING | Уникальный, sensitive |
| device | STRING | User-Agent |
| ip_address | STRING | Опциональный |
| last_active_at | TIME | Последняя активность |
| expires_at | TIME | Срок действия |
| created_at | TIME | Immutable |

**Индексы:** expires_at
**Связи:** user (M:1)

#### projects

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | 2-100 символов |
| key | STRING | Уникальный, regex (например HUB, MOB) |
| description | STRING | До 2000 символов |
| avatar_url | STRING | Опциональный |
| status | ENUM(active, archived, deleted) | По умолчанию "active" |
| task_counter | INT | Автоинкремент для задач |
| test_case_counter | INT | Автоинкремент для тест-кейсов |
| deleted_at | TIME | Nullable, мягкое удаление |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** status, deleted_at
**Связи:** members, roles, teams, invites, files, issues, workflow_statuses, labels, label_groups, boards, custom_fields, channels, test_cases, test_runs, test_suites, components, saved_filters, github_integrations, gitlab_integrations, recordings, scheduled_calls, webhook_endpoints, environments, sprints

#### project_members

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| project_role | ENUM(owner, developer, tester, viewer) | По умолчанию "developer" |
| joined_at | TIME | Immutable |

**Индексы:** UNIQUE(user, project)
**Связи:** user (M:1), project (M:1), custom_role (M:1, optional)

#### issues

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| issue_number | INT | Positive, immutable |
| issue_key | STRING | Уникальный, immutable (HUB-1) |
| title | STRING | До 500 символов |
| description_json | JSONB | Plate editor JSON |
| description_text | TEXT | Plain text для поиска |
| issue_type | ENUM(bug, task, story, epic, sub_task, improvement, tech_debt, spike, release) | |
| priority | ENUM(critical, high, medium, low, trivial) | По умолчанию "medium" |
| rank | STRING | LexoRank для Kanban |
| version | INT | Оптимистичная блокировка, по умолчанию 1 |
| story_points | INT | Nullable, Fibonacci |
| estimate_minutes | INT | Nullable |
| time_spent_minutes | INT | По умолчанию 0 |
| checklist | JSONB | [{id, text, checked}] |
| custom_fields | JSONB | custom_field_def_id -> value |
| due_date | TIME | Nullable |
| deleted_at | TIME | Nullable, корзина 30 дней |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** issue_key, issue_number, rank, created_at, priority, due_date, deleted_at
**Связи:** project (M:1), status/WorkflowStatus (M:1), assignee/User (M:1), reporter/User (M:1), parent/children (self-referential), comments, labels, time_logs, activities, files, outgoing_links, incoming_links, linked_test_cases, github_links, gitlab_links, sprint (M:1)

#### workflow_statuses

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | До 100 символов |
| category | ENUM(todo, in_progress, done) | |
| color | STRING | По умолчанию "#6B7280" |
| sort_order | INT | Порядок отображения |
| is_default | BOOL | По умолчанию false |
| wip_limit | INT | 0 = без лимита |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(project, name), (project, sort_order)
**Связи:** project (M:1), issues (1:M), from_transitions (1:M), to_transitions (M:1)

#### workflow_transitions

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(from_status, to_status)
**Связи:** from_status (M:1), to_status (M:1)

#### comments

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| body_json | JSONB | Plate editor JSON |
| body_text | TEXT | Plain text для поиска/уведомлений |
| is_internal | BOOL | Видимость по ролям |
| mentions | JSON([]string) | User IDs |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Связи:** issue (M:1), author/User (M:1), reactions (1:M)

#### reactions

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| emoji | STRING | До 32 символов |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(comment, user, emoji)
**Связи:** comment (M:1), user (M:1)

#### issue_links

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| link_type | ENUM(blocks, blocked_by, relates_to, duplicates, duplicated_by, clones, cloned_from, causes, caused_by) | Иерархия epic/sub_task — только Issue.self-edge parent/children (ADR-ENT-ONDELETE-CASCADE-2026) |
| custom_type | STRING | Для пользовательских типов связей |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(source, target)
**Связи:** source/Issue (M:1), target/Issue (M:1)

#### labels

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | До 50 символов |
| color | STRING | По умолчанию "#3B82F6" |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(project, name)
**Связи:** project (M:1), group/LabelGroup (M:1, optional), issues (M:M)

#### label_groups

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | Название группы |
| created_at | TIME | Immutable |

**Связи:** project (M:1), labels (1:M)

#### boards

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | До 100 символов |
| filters | JSONB | Конфигурация фильтров |
| is_default | BOOL | По умолчанию false |
| swimlane_mode | ENUM(none, assignee, priority, type) | По умолчанию "none" |
| created_at | TIME | Immutable |

**Связи:** project (M:1)

#### custom_field_definitions

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | Название поля |
| field_type | ENUM(text, number, date, checkbox, select, multi_select, user, url) | |
| options | JSON([]string) | Варианты для select/multi_select |
| is_required | BOOL | По умолчанию false |
| created_at | TIME | Immutable |

**Связи:** project (M:1)

#### time_logs

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| minutes | INT | Positive |
| description | STRING | До 500 символов |
| started_at | TIME | Nullable (для таймера) |
| ended_at | TIME | Nullable |
| created_at | TIME | Immutable |

**Связи:** issue (M:1), user (M:1)

#### activity_logs

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| action | STRING | created/updated/commented/status_changed/assigned |
| changes | JSON([]FieldChange) | [{field, old_value, new_value}] |
| created_at | TIME | Immutable |

**Индексы:** (issue, created_at)
**Связи:** issue (M:1), user (M:1)

#### channels

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| channel_type | ENUM(dm, group, public, private) | |
| name | STRING | До 100 символов, опциональный для DM |
| description | STRING | До 500 символов |
| is_readonly | BOOL | По умолчанию false |
| project_id | UUID | Nullable |
| created_by | UUID | ID создателя |
| dm_key | STRING | Unique, nullable. Дедупликация DM |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** channel_type, project_id
**Связи:** members (1:M), project (M:1), creator/User (M:1), pinned_messages (1:M), calls (1:M), scheduled_calls (1:M), polls (1:M), invites (1:M)

#### channel_members

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| channel_id | UUID | FK |
| user_id | UUID | FK |
| role | ENUM(owner, admin, member) | По умолчанию "member" |
| joined_at | TIME | |
| is_starred | BOOL | По умолчанию false |
| muted_until | TIME | Nullable |
| notify_mode | ENUM(all, mentions, nothing) | По умолчанию "all" |

**Индексы:** UNIQUE(channel_id, user_id)
**Связи:** channel (M:1), user (M:1)

#### pinned_messages

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| channel_id | UUID | FK |
| message_id | STRING | MongoDB message ID |
| pinned_by | UUID | FK |
| created_at | TIME | Immutable |

**Связи:** channel (M:1)

#### bookmarks

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| user_id | UUID | FK |
| message_id | STRING | MongoDB message ID |
| channel_id | UUID | FK |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(user_id, message_id)
**Связи:** user (M:1)

#### polls

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| question | STRING | Вопрос опроса |
| channel_id | UUID | FK |
| created_by | UUID | FK |
| is_closed | BOOL | По умолчанию false |
| is_anonymous | BOOL | По умолчанию false |
| message_id | STRING | MongoDB message ID |
| created_at | TIME | Immutable |

**Связи:** channel (M:1), creator/User (M:1), options (1:M), votes (1:M)

#### poll_options

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| text | STRING | Текст варианта |
| sort_order | INT | Порядок |

**Связи:** poll (M:1), votes (1:M)

#### poll_votes

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(user, poll)
**Связи:** poll (M:1), option (M:1), user (M:1)

#### channel_invites

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| token | STRING | Уникальный токен |
| max_uses | INT | 0 = unlimited |
| use_count | INT | По умолчанию 0 |
| expires_at | TIME | Nullable |
| created_at | TIME | Immutable |

**Связи:** channel (M:1), created_by/User (M:1)

#### user_blocks

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| blocker_id | UUID | FK |
| blocked_id | UUID | FK |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(blocker_id, blocked_id)
**Связи:** blocker/User, blocked/User

#### test_cases

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| case_number | INT | Positive, immutable |
| case_key | STRING | Уникальный (HUB-TC-42) |
| title | STRING | До 500 символов |
| description | TEXT | Опциональный |
| precondition | TEXT | Предусловия |
| postcondition | TEXT | Постусловия |
| case_type | ENUM(classic, bdd) | По умолчанию "classic" |
| priority | ENUM(critical, high, medium, low, trivial) | По умолчанию "medium" |
| status | ENUM(draft, active, deprecated) | По умолчанию "draft" |
| steps | JSON([]TestStep) | [{order, action, expected_result, keyword}] |
| full_name | STRING | Allure matching |
| class_name | STRING | Allure matching |
| method_name | STRING | Allure matching |
| external_id | STRING | @TestCase annotation |
| automation_id | STRING | Явный ID для автотестов |
| version | INT | По умолчанию 1 |
| sync_status | ENUM(synced, outdated, unlinked) | По умолчанию "unlinked" |
| deleted_at | TIME | Nullable, корзина 30 дней |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** case_key, external_id, automation_id, full_name, (class_name, method_name), status, priority, deleted_at
**Связи:** project (M:1), suite/TestSuite (M:1), created_by/User (M:1), versions (1:M), results (1:M), labels (M:M), linked_issues (M:M), components (M:M)

#### test_case_versions

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| version_number | INT | Positive |
| snapshot | JSONB | Полный снимок тест-кейса |
| change_description | TEXT | Описание изменений |
| created_at | TIME | Immutable |

**Индексы:** version_number
**Связи:** test_case (M:1), created_by/User (M:1)

#### test_suites

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | До 200 символов |
| description | TEXT | Опциональный |
| sort_order | INT | По умолчанию 0 |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** name
**Связи:** project (M:1), test_cases (1:M), parent/children (self-referential)

#### components

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | Название компонента |
| description | TEXT | Опциональный |
| created_at | TIME | Immutable |

**Связи:** project (M:1), test_cases (M:M)

#### test_runs

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| title | STRING | До 500 символов |
| description | TEXT | Опциональный |
| run_type | ENUM(manual, automatic, mixed) | По умолчанию "manual" |
| status | ENUM(created, in_progress, completed, aborted) | По умолчанию "created" |
| total_count | INT | По умолчанию 0 |
| passed_count | INT | По умолчанию 0 |
| failed_count | INT | По умолчанию 0 |
| skipped_count | INT | По умолчанию 0 |
| broken_count | INT | По умолчанию 0 |
| unknown_count | INT | По умолчанию 0 |
| allure_run_id | STRING | ID прогона Allure |
| started_at | TIME | Nullable |
| completed_at | TIME | Nullable |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** status, run_type, allure_run_id, created_at
**Связи:** project (M:1), created_by/User (M:1), results (1:M), assignees (M:M/User), environment (M:1)

#### test_results

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| status | ENUM(passed, failed, broken, skipped, unknown) | По умолчанию "unknown" |
| comment | TEXT | Опциональный |
| step_results | JSON([]StepResult) | [{step_order, status, comment, file_ids}] |
| duration_ms | INT64 | Nullable, миллисекунды |
| allure_uuid | STRING | UUID из Allure |
| error_message | TEXT | Сообщение об ошибке |
| error_trace | TEXT | Stack trace |
| is_flaky | BOOL | По умолчанию false |
| match_strategy | ENUM(annotation, fullname, classmethod, unmapped) | Стратегия привязки |
| started_at | TIME | Nullable |
| completed_at | TIME | Nullable |
| created_at | TIME | Immutable |

**Индексы:** status, allure_uuid, is_flaky
**Связи:** run/TestRun (M:1), test_case (M:1), executed_by/User (M:1), files (1:M)

#### environments

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | Название (dev, staging, prod) |
| description | TEXT | Опциональный |
| config | JSONB | Конфигурация (browser, OS) |
| created_at | TIME | Immutable |

**Связи:** project (M:1), test_runs (1:M)

#### sprints

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | До 200 символов |
| goal | STRING | До 1000 символов |
| status | ENUM(planning, active, completed) | По умолчанию "planning" |
| start_date | TIME | Дата начала |
| end_date | TIME | Дата окончания |
| retro_notes | JSONB | {went_well, to_improve, actions} |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** (project, status), (project, start_date)
**Связи:** project (M:1), issues (M:M)

#### calls

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| room_name | STRING | Уникальное имя комнаты LiveKit |
| channel_id | UUID | Nullable (nil для instant) |
| invite_code | STRING | Unique, до 12 символов |
| title | STRING | До 255 символов |
| created_by | UUID | FK |
| status | ENUM(active, ended) | По умолчанию "active" |
| started_at | TIME | Nullable |
| ended_at | TIME | Nullable |
| participant_count | INT | По умолчанию 0 |
| created_at | TIME | Immutable |

**Индексы:** (channel_id, status), room_name, invite_code
**Связи:** channel (M:1), creator/User (M:1), participants (1:M), recordings (1:M)

#### call_participants

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| call_id | UUID | FK |
| user_id | UUID | FK |
| joined_at | TIME | |
| left_at | TIME | Nullable |

**Индексы:** UNIQUE(call_id, user_id)
**Связи:** call (M:1), user (M:1)

#### recordings

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| egress_id | STRING | ID egress в LiveKit |
| status | ENUM(recording, processing, ready, failed) | |
| file_path | STRING | S3 key |
| file_size | INT64 | Байты |
| duration_seconds | INT | Секунды |
| started_at | TIME | |
| ended_at | TIME | Nullable |
| call_id | UUID | FK |
| project_id | UUID | FK |
| started_by | UUID | FK |
| created_at | TIME | Immutable |

**Индексы:** (project_id, created_at), call_id, egress_id
**Связи:** call (M:1), project (M:1), user (M:1)

#### scheduled_calls

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| title | STRING | Название |
| scheduled_at | TIME | Время начала |
| duration_minutes | INT | Планируемая длительность |
| recurrence | STRING | Повторение (daily, weekly, none) |
| channel_id | UUID | FK |
| project_id | UUID | FK |
| created_by | UUID | FK |
| created_at | TIME | Immutable |

**Связи:** channel (M:1), project (M:1), creator/User (M:1)

#### notifications

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| user_id | UUID | FK получателя |
| type | STRING | assigned, mentioned, status_changed и т.д. |
| title | STRING | Заголовок |
| body | STRING | Тело |
| entity_type | STRING | issue, channel, testrun |
| entity_id | STRING | ID сущности |
| project_id | UUID | Nullable |
| actor_id | UUID | Nullable, инициатор |
| is_read | BOOL | По умолчанию false |
| is_urgent | BOOL | По умолчанию false |
| created_at | TIME | Immutable |

**Индексы:** (user_id, is_read), (user_id, created_at), (user_id, type, entity_id, created_at)
**Связи:** user (M:1)

#### notification_preferences

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| user_id | UUID | FK |
| notification_type | STRING | Тип уведомления |
| channel | STRING | in_app, telegram |
| enabled | BOOL | По умолчанию true |
| project_id | UUID | Nullable (nil = глобальная) |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** UNIQUE(user_id, notification_type, channel, project_id)

#### push_subscriptions

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| user_id | UUID | FK |
| endpoint | STRING | Push endpoint URL |
| auth_key | STRING | Ключ аутентификации |
| p256dh_key | STRING | Ключ шифрования |
| created_at | TIME | Immutable |

**Связи:** user (M:1)

#### file_metadata

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| filename | STRING | Имя файла |
| content_type | STRING | MIME тип |
| size | INT64 | Размер в байтах |
| s3_key | STRING | Ключ в S3 |
| s3_bucket | STRING | Bucket |
| entity_type | STRING | task, message, testcase, avatar |
| entity_id | STRING | |
| status | ENUM(pending, confirmed, deleted) | |
| deleted_at | TIME | Nullable, корзина 30 дней |
| created_at | TIME | Immutable |

**Индексы:** (entity_type, entity_id), status, s3_key, deleted_at
**Связи:** uploader/User (M:1), project (M:1)

#### invites

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| code | STRING | Уникальный код |
| max_uses | INT | 0 = unlimited |
| use_count | INT | |
| project_role | ENUM(developer, tester, viewer) | |
| expires_at | TIME | Nullable |
| created_at | TIME | Immutable |

**Индексы:** expires_at
**Связи:** project (M:1), created_by/User (M:1)

#### roles

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | Название роли |
| description | STRING | Опциональный |
| permissions | UINT64 | Bitmask |
| is_system | BOOL | По умолчанию false |
| created_at | TIME | Immutable |

**Связи:** project (M:1), members/ProjectMember (1:M)

#### teams

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | |
| description | STRING | Опциональный |
| created_at | TIME | Immutable |

**Связи:** projects (M:M), members/TeamMember (1:M)

#### team_members

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(team, user)
**Связи:** team (M:1), user (M:1)

#### saved_filters

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | До 200 символов |
| filter_config | JSONB | Конфигурация фильтра |
| shared | BOOL | По умолчанию false |
| owner_id | UUID | FK |
| project_id | UUID | FK |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** (owner_id, project_id), (project_id, shared)
**Связи:** owner/User (M:1), project (M:1)

#### dashboard_layouts

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| user_id | UUID | FK |
| dashboard_type | ENUM(personal, project, admin) | |
| project_id | UUID | Nullable |
| items | JSON([]{position, size, widget_type}) | Виджеты |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** UNIQUE(user_id, dashboard_type, project_id)
**Связи:** user (M:1)

#### api_keys

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | До 255 символов |
| key_hash | STRING | SHA-256 хеш, уникальный |
| prefix | STRING | Первые 8 символов |
| last_used_at | TIME | Nullable |
| expires_at | TIME | Nullable |
| created_at | TIME | Immutable |

**Индексы:** key_hash, expires_at
**Связи:** user (M:1)

#### audit_logs

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| action | STRING | login, logout, create, update, delete |
| entity_type | STRING | issue, project, test_case, channel, file, user |
| entity_id | STRING | UUID или ключ |
| user_id | UUID | Nullable |
| project_id | UUID | Nullable |
| details | JSONB | Дополнительные данные |
| ip_address | STRING | |
| user_agent | STRING | |
| created_at | TIME | Immutable |

**Индексы:** action, entity_type, user_id, project_id, created_at

#### system_settings

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| key | STRING | Уникальный ключ настройки |
| value | STRING | Значение |
| updated_at | TIME | Auto-update |

#### backup_records

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| s3_key | STRING | Ключ в S3 |
| file_size | INT64 | Байты |
| status | ENUM(in_progress, completed, failed) | |
| error_message | TEXT | Nullable |
| created_at | TIME | Immutable |

**Индексы:** status, created_at

#### github_integrations

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| repo_owner | STRING | |
| repo_name | STRING | |
| access_token | STRING | Sensitive |
| webhook_secret | STRING | |
| merge_status_id | UUID | Nullable, целевой статус при merge |
| created_at | TIME | Immutable |

**Индексы:** UNIQUE(repo_owner, repo_name, project)
**Связи:** project (M:1), github_links (1:M)

#### github_links

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| link_type | ENUM(commit, pull_request) | |
| external_id | STRING | PR number или commit SHA |
| title | STRING | До 500 символов |
| url | STRING | URL на GitHub |
| status | ENUM(open, merged, closed, none) | |
| author | STRING | |
| branch | STRING | |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Индексы:** issue, (external_id, link_type)
**Связи:** issue (M:1), integration/GitHubIntegration (M:1)

#### gitlab_integrations

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| project_url | STRING | URL проекта GitLab |
| access_token | STRING | Sensitive |
| webhook_token | STRING | |
| merge_status_id | UUID | Nullable |
| created_at | TIME | Immutable |

**Связи:** project (M:1), gitlab_links (1:M)

#### gitlab_links

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| link_type | ENUM(commit, merge_request) | |
| external_id | STRING | |
| title | STRING | |
| url | STRING | |
| status | ENUM(open, merged, closed, none) | |
| author | STRING | |
| branch | STRING | |
| created_at | TIME | Immutable |
| updated_at | TIME | Auto-update |

**Связи:** issue (M:1), integration/GitLabIntegration (M:1)

#### webhook_endpoints

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| url | STRING | Endpoint URL |
| secret | STRING | Для подписи payload |
| events | JSON([]string) | Список событий |
| is_active | BOOL | По умолчанию true |
| created_at | TIME | Immutable |

**Связи:** project (M:1), deliveries (1:M)

#### webhook_deliveries

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| event | STRING | Тип события |
| payload | JSONB | Тело запроса |
| response_status | INT | HTTP status |
| response_body | TEXT | Тело ответа |
| delivered_at | TIME | |
| created_at | TIME | Immutable |

**Связи:** endpoint/WebhookEndpoint (M:1)

#### project_templates

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name | STRING | Название шаблона |
| description | STRING | Опциональный |
| config | JSONB | Конфигурация (workflow, labels) |
| is_system | BOOL | Системный шаблон |
| created_at | TIME | Immutable |

**Связи:** created_by/User (M:1)

#### outbox_events

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| event_type | STRING | Тип события |
| payload | JSONB | Данные события |
| status | ENUM(pending, sent, failed) | |
| created_at | TIME | Immutable |

### v3/v4 — Дополнительные таблицы (Ent Schema)

> Полный перечень: **82** схемы в `backend/ent/schema/`. Ниже — выборка v3/v4 super-app (см. `docs/wiki/API.md`).

#### crm_contacts / crm_deals

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| name / title | STRING | Имя контакта / название сделки |
| email, phone, company | STRING | Контактные данные (contacts) |
| stage | ENUM(lead…lost) | Стадия pipeline (deals) |
| amount_cents | INT64 | Сумма сделки (deals) |
| project_id | UUID | Nullable, optional link к проекту |
| version | INT | Optimistic locking |
| deleted_at | TIME | Soft delete |

#### news_posts / kudos (intranet)

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| title / message | STRING / TEXT | Заголовок новости / текст kudos |
| body_json | JSONB | Rich-text новости |
| is_pinned / is_public | BOOL | Закрепление / публичность |
| author_id, recipient_id | UUID | FK на users |
| status | ENUM(draft, published) | Новости |
| version | INT | Optimistic locking |

#### hub_bases / hub_tables / hub_columns / hub_records

| Сущность | Ключевые поля |
|----------|---------------|
| hub_bases | name, description, version, deleted_at |
| hub_tables | base_id, name, sort_order |
| hub_columns | table_id, name, col_type (text/number/date/select/checkbox/url/user_ref), options JSON |
| hub_records | table_id, cells JSONB `{column_uuid: value}`, version, deleted_at |

#### ledger_accounts / journal_entries / journal_lines / invoices (accounting)

| Сущность | Ключевые поля |
|----------|---------------|
| ledger_accounts | code, name, account_type (asset/liability/equity/revenue/expense) |
| journal_entries | status (draft/posted), posted_at, description |
| journal_lines | entry_id, account_id, debit_cents, credit_cents |
| invoices | number, status (draft/issued/paid), amount_cents, crm_deal_id (optional) |

#### ai_providers / ai_usage_daily / ai_user_preferences

| Сущность | Ключевые поля |
|----------|---------------|
| ai_providers | name, provider_type, base_url, api_key_enc (AES-GCM), models JSON, enabled, priority |
| ai_usage_daily | user_id, date, tokens_used, requests_count |
| ai_user_preferences | user_id, default_model, allow_external_ai |

#### marketplace_listings

| Колонка | Тип | Описание |
|---------|-----|----------|
| id | UUID | PK |
| title | STRING | Заголовок объявления |
| description | TEXT | Описание |
| category | ENUM(equipment, books, services, furniture, other) | |
| price_cents | INT64 | Nullable |
| status | ENUM(pending, approved, rejected, sold, withdrawn) | |
| author_id | UUID | FK |
| version | INT | Optimistic locking |
| deleted_at | TIME | Soft delete |

### MongoDB — Коллекции

| Коллекция | Описание |
|-----------|----------|
| messages | Сообщения мессенджера (content, sender_id, channel_id, reply_to, files, reactions) |
| allure_reports | Allure JSON отчёты (сырые данные из CI) |

---

## Справочник API

Базовый путь: `/api`

### Auth (11 эндпоинтов)

> **v2.5 (Phase 15):** Login — **Magic Link only** (frontend `MagicLinkForm`). Telegram — **notify bind only** (`/notifications/telegram/*`), не login. Registration default: **invite-only** + admin Pending queue.

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /auth/magic-link | Отправить magic-link код на email | - |
| POST | /auth/magic-link/verify | Верифицировать код → JWT + refresh | - |
| POST | /auth/refresh | Обновить access token через refresh token | - |
| POST | /auth/logout | Выход из текущей сессии | JWT |
| GET | /auth/me | Получить профиль текущего пользователя | JWT |
| PUT | /auth/me | Обновить профиль | JWT |
| POST | /auth/me/onboarding | Завершить онбординг | JWT |
| GET | /auth/sessions | Список активных сессий | JWT |
| DELETE | /auth/sessions/:sessionId | Отозвать сессию | JWT |
| GET | /admin/registration-mode | Получить режим регистрации | JWT (Admin) |
| PUT | /admin/registration-mode | Установить режим регистрации | JWT (Admin) |

> **Removed v2.5:** `POST /auth/telegram`, `POST /auth/telegram-code` (AUTH-01).

### Admin mail / hub-mail (Phase 15.2)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /admin/mail/settings | Настройки transactional email / hub-mail | JWT (Admin) |
| PUT | /admin/mail/settings | Обновить from, URL hub-mail, лимиты | JWT (Admin) |

Микросервис: `cmd/hub-mail` — `GET /health`, `GET /v1/status`, `POST /v1/send` (internal).

### Projects (22 эндпоинта)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /projects | Создать проект | JWT |
| GET | /projects | Список проектов пользователя | JWT |
| GET | /projects/trash | Список удалённых проектов | JWT |
| GET | /projects/:projectId | Получить проект | JWT |
| PUT | /projects/:projectId | Обновить проект | JWT |
| POST | /projects/:projectId/archive | Архивировать проект | JWT |
| POST | /projects/:projectId/unarchive | Разархивировать проект | JWT |
| DELETE | /projects/:projectId | Удалить проект (мягкое) | JWT |
| POST | /projects/:projectId/restore | Восстановить из корзины | JWT |
| GET | /projects/:projectId/members | Список участников | JWT |
| POST | /projects/:projectId/members | Добавить участника | JWT |
| PUT | /projects/:projectId/members/:userId | Изменить роль участника | JWT |
| DELETE | /projects/:projectId/members/:userId | Удалить участника | JWT |
| POST | /projects/:projectId/invites | Создать invite-ссылку | JWT |
| GET | /projects/:projectId/invites | Список invite-ссылок | JWT |
| DELETE | /projects/:projectId/invites/:inviteId | Отозвать invite | JWT |
| POST | /invites/join | Присоединиться по invite-коду | JWT |
| POST | /projects/:projectId/teams | Создать команду | JWT |
| GET | /projects/:projectId/teams | Список команд | JWT |
| POST | /projects/:projectId/teams/:teamId/members | Добавить участника в команду | JWT |
| DELETE | /projects/:projectId/teams/:teamId/members/:userId | Удалить участника из команды | JWT |
| POST | /projects/:projectId/teams/:teamId/projects/:projectId | Добавить команду в проект | JWT |

### Tracker (35 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /projects/:projectId/issues | Создать задачу | JWT + RBAC |
| GET | /projects/:projectId/issues | Список задач | JWT + RBAC |
| GET | /projects/:projectId/issues/:issueKey | Получить задачу | JWT + RBAC |
| PUT | /projects/:projectId/issues/:issueKey | Обновить задачу | JWT + RBAC |
| DELETE | /projects/:projectId/issues/:issueKey | Удалить задачу | JWT + RBAC |
| PUT | /projects/:projectId/issues/:issueKey/position | Изменить позицию (rank) | JWT + RBAC |
| POST | /projects/:projectId/issues/jql | Поиск задач по JQL | JWT + RBAC |
| POST | /projects/:projectId/issues/:issueKey/links | Создать связь | JWT + RBAC |
| GET | /projects/:projectId/issues/:issueKey/links | Список связей | JWT + RBAC |
| DELETE | /projects/:projectId/issues/:issueKey/links/:linkId | Удалить связь | JWT + RBAC |
| POST | /projects/:projectId/issues/:issueKey/comments | Создать комментарий | JWT + RBAC |
| GET | /projects/:projectId/issues/:issueKey/comments | Список комментариев | JWT + RBAC |
| PUT | /projects/:projectId/issues/:issueKey/comments/:commentId | Обновить комментарий | JWT + RBAC |
| DELETE | /projects/:projectId/issues/:issueKey/comments/:commentId | Удалить комментарий | JWT + RBAC |
| POST | /projects/:projectId/issues/:issueKey/comments/:commentId/reactions | Toggle реакция | JWT + RBAC |
| GET | /projects/:projectId/issues/:issueKey/activity | Лог изменений | JWT + RBAC |
| POST | /projects/:projectId/issues/:issueKey/time | Залогировать время | JWT + RBAC |
| POST | /projects/:projectId/issues/:issueKey/timer/start | Запустить таймер | JWT + RBAC |
| POST | /projects/:projectId/issues/:issueKey/timer/stop | Остановить таймер | JWT + RBAC |
| GET | /projects/:projectId/issues/:issueKey/time-logs | Список time-логов | JWT + RBAC |
| GET | /projects/:projectId/boards | Список досок | JWT + RBAC |
| POST | /projects/:projectId/boards | Создать доску | JWT + RBAC |
| GET | /projects/:projectId/boards/:boardId | Данные доски (Kanban) | JWT + RBAC |
| GET | /projects/:projectId/workflow/statuses | Список статусов | JWT + RBAC |
| POST | /projects/:projectId/workflow/statuses | Создать статус | JWT + RBAC |
| PUT | /projects/:projectId/workflow/statuses/:statusId | Обновить статус | JWT + RBAC |
| DELETE | /projects/:projectId/workflow/statuses/:statusId | Удалить статус | JWT + RBAC |
| GET | /projects/:projectId/workflow/transitions | Список переходов | JWT + RBAC |
| POST | /projects/:projectId/workflow/transitions | Создать переход | JWT + RBAC |
| DELETE | /projects/:projectId/workflow/transitions/:transitionId | Удалить переход | JWT + RBAC |
| GET | /projects/:projectId/labels | Список меток | JWT + RBAC |
| POST | /projects/:projectId/labels | Создать метку | JWT + RBAC |
| DELETE | /projects/:projectId/labels/:labelId | Удалить метку | JWT + RBAC |
| GET | /projects/:projectId/labels/groups | Список групп меток | JWT + RBAC |
| POST | /projects/:projectId/labels/groups | Создать группу меток | JWT + RBAC |
| GET | /projects/:projectId/custom-fields | Список кастомных полей | JWT + RBAC |
| POST | /projects/:projectId/custom-fields | Создать кастомное поле | JWT + RBAC |
| DELETE | /projects/:projectId/custom-fields/:fieldId | Удалить кастомное поле | JWT + RBAC |

### Messenger (38 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /messenger/channels | Создать канал | JWT |
| GET | /messenger/channels | Список каналов пользователя | JWT |
| GET | /messenger/channels/:channelId | Получить канал | JWT |
| PUT | /messenger/channels/:channelId | Обновить канал | JWT |
| DELETE | /messenger/channels/:channelId | Удалить канал | JWT |
| POST | /messenger/channels/:channelId/join | Присоединиться к каналу | JWT |
| POST | /messenger/channels/:channelId/leave | Покинуть канал | JWT |
| POST | /messenger/channels/:channelId/members | Добавить участников | JWT |
| GET | /messenger/channels/:channelId/members | Список участников | JWT |
| POST | /messenger/channels/:channelId/star | Toggle звёздочка | JWT |
| POST | /messenger/channels/:channelId/mute | Замьютить канал | JWT |
| POST | /messenger/channels/:channelId/unmute | Размьютить канал | JWT |
| POST | /messenger/channels/:channelId/typing | Отправить typing indicator | JWT |
| POST | /messenger/channels/:channelId/messages | Отправить сообщение | JWT |
| GET | /messenger/channels/:channelId/messages | Список сообщений | JWT |
| PUT | /messenger/channels/:channelId/messages/:messageId | Редактировать сообщение | JWT |
| DELETE | /messenger/channels/:channelId/messages/:messageId | Удалить сообщение | JWT |
| POST | /messenger/channels/:channelId/messages/:messageId/reactions | Toggle реакция | JWT |
| POST | /messenger/channels/:channelId/messages/:messageId/pin | Toggle pin | JWT |
| POST | /messenger/channels/:channelId/messages/:messageId/forward | Переслать сообщение | JWT |
| GET | /messenger/channels/:channelId/messages/:messageId/replies | Список тредовых ответов | JWT |
| GET | /messenger/presence | Статусы присутствия | JWT |
| PUT | /messenger/presence/status | Установить кастомный статус | JWT |
| POST | /messenger/presence/heartbeat | Heartbeat (обновление онлайн-статуса) | JWT |
| GET | /messenger/unread | Непрочитанные по каналам | JWT |
| POST | /messenger/channels/:channelId/read | Пометить канал прочитанным | JWT |
| GET | /messenger/channels/:channelId/pins | Список закреплённых сообщений | JWT |
| GET | /messenger/search | Поиск по сообщениям | JWT |
| GET | /messenger/bookmarks | Список закладок | JWT |
| POST | /messenger/bookmarks | Добавить закладку | JWT |
| DELETE | /messenger/bookmarks/:bookmarkId | Удалить закладку | JWT |
| POST | /messenger/dm | Создать или получить DM | JWT |
| POST | /messenger/channels/:channelId/polls | Создать опрос | JWT |
| POST | /messenger/polls/:pollId/vote | Проголосовать | JWT |
| POST | /messenger/polls/:pollId/close | Закрыть опрос | JWT |
| GET | /messenger/polls/:pollId | Результаты опроса | JWT |
| POST | /messenger/channels/:channelId/invites | Создать invite-ссылку канала | JWT |
| GET | /messenger/channels/:channelId/invites | Список invite-ссылок | JWT |
| DELETE | /messenger/channels/:channelId/invites/:inviteId | Отозвать invite | JWT |
| GET | /messenger/invites/:token | Валидация invite | JWT |
| POST | /messenger/invites/:token/accept | Принять invite | JWT |

### TMS (47 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /projects/:projectId/tms/cases | Создать тест-кейс | JWT + RBAC |
| GET | /projects/:projectId/tms/cases | Список тест-кейсов | JWT + RBAC |
| GET | /projects/:projectId/tms/cases/:caseKey | Получить тест-кейс | JWT + RBAC |
| PUT | /projects/:projectId/tms/cases/:caseKey | Обновить тест-кейс | JWT + RBAC |
| DELETE | /projects/:projectId/tms/cases/:caseKey | Удалить тест-кейс | JWT + RBAC |
| GET | /projects/:projectId/tms/cases/:caseKey/versions | История версий | JWT + RBAC |
| POST | /projects/:projectId/tms/cases/:caseKey/rollback/:versionId | Откат к версии | JWT + RBAC |
| POST | /projects/:projectId/tms/cases/:caseKey/links | Привязать к задаче | JWT + RBAC |
| DELETE | /projects/:projectId/tms/cases/:caseKey/links/:issueId | Отвязать от задачи | JWT + RBAC |
| GET | /projects/:projectId/tms/cases/:caseKey/links | Список связей с задачами | JWT + RBAC |
| POST | /projects/:projectId/tms/suites | Создать suite (папку) | JWT + RBAC |
| GET | /projects/:projectId/tms/suites | Список suites | JWT + RBAC |
| PUT | /projects/:projectId/tms/suites/:suiteId | Обновить suite (`version` OL) | JWT + RBAC |
| DELETE | /projects/:projectId/tms/suites/:suiteId | Удалить suite | JWT + RBAC |
| POST | /projects/:projectId/tms/components | Создать компонент | JWT + RBAC |
| GET | /projects/:projectId/tms/components | Список компонентов | JWT + RBAC |
| DELETE | /projects/:projectId/tms/components/:componentId | Удалить компонент | JWT + RBAC |
| POST | /projects/:projectId/tms/runs | Создать прогон | JWT + RBAC |
| GET | /projects/:projectId/tms/runs | Список прогонов | JWT + RBAC |
| GET | /projects/:projectId/tms/runs/:runId | Получить прогон | JWT + RBAC |
| PUT | /projects/:projectId/tms/runs/:runId/status | Обновить статус прогона | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/reopen | Переоткрыть прогон | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/cases | Добавить кейсы в прогон | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/assignees | Назначить тестировщиков | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/start | Запустить прогон | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/complete | Завершить прогон | JWT + RBAC |
| GET | /projects/:projectId/tms/runs/:runId/detail | Детализация прогона | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/results | Создать результат | JWT + RBAC |
| PUT | /projects/:projectId/tms/runs/:runId/results/:resultId | Обновить результат | JWT + RBAC |
| GET | /projects/:projectId/tms/runs/:runId/results | Список результатов | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/allure | Импорт Allure в существующий прогон | JWT + RBAC |
| POST | /projects/:projectId/tms/allure/import | Импорт Allure результатов (новый прогон) | JWT + RBAC |
| POST | /projects/:projectId/tms/runs/:runId/results/:resultId/bug | Создать баг из результата | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/history | История прогонов | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/flaky | Flaky-тесты | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/coverage | Покрытие | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/trends | Тренды | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/comparison | Сравнение прогонов | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/traceability | Матрица трассируемости | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/defect-categories | Категории дефектов | JWT + RBAC |
| GET | /projects/:projectId/tms/reports/env-comparison | Сравнение окружений | JWT + RBAC |
| POST | /projects/:projectId/tms/import | Импорт тест-кейсов (CSV/JSON/XML/Excel) | JWT + RBAC |
| GET | /projects/:projectId/tms/export | Экспорт тест-кейсов | JWT + RBAC |
| GET | /projects/:projectId/tms/runs/:runId/export | Экспорт отчёта прогона (PDF) | JWT + RBAC |
| POST | /projects/:projectId/tms/webhook/allure | Webhook Allure | API Key |
| POST | /projects/:projectId/tms/webhook/junit | Webhook JUnit | API Key |
| POST | /projects/:projectId/tms/environments | Создать окружение | JWT + RBAC |
| GET | /projects/:projectId/tms/environments | Список окружений | JWT + RBAC |
| PUT | /projects/:projectId/tms/environments/:envId | Обновить окружение | JWT + RBAC |
| DELETE | /projects/:projectId/tms/environments/:envId | Удалить окружение | JWT + RBAC |

### Calls (20 эндпоинтов)

> **Гостевой доступ:** Эндпоинт `POST /calls/guest-join` позволяет присоединиться к звонку без аккаунта Hub — достаточно invite_code и имени.

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /calls | Создать звонок (привязка к каналу) | JWT |
| POST | /calls/instant | Создать instant-звонок (без канала) | JWT |
| GET | /calls/invite/:code | Присоединиться по invite-коду | JWT |
| POST | /calls/guest-join | Гостевой вход по invite_code + guest_name (без аккаунта) | - |
| GET | /calls/channel/:channelId | Получить активный звонок в канале | JWT |
| POST | /calls/:callId/join | Присоединиться к звонку | JWT |
| POST | /calls/:callId/leave | Покинуть звонок | JWT |
| POST | /calls/:callId/end | Завершить звонок | JWT |
| POST | /calls/:callId/raise-hand | Поднять руку | JWT |
| POST | /calls/:callId/chat | Отправить сообщение в чат звонка | JWT |
| GET | /calls/:callId/chat | Список сообщений чата звонка | JWT |
| POST | /calls/:callId/recording/start | Начать запись (channel member) | JWT |
| POST | /calls/:callId/recording/stop | Остановить запись (channel member + call ownership) | JWT |
| GET | /projects/:projectId/recordings | Список записей проекта | JWT |
| POST | /projects/:projectId/scheduled-calls | Создать запланированный звонок | JWT |
| GET | /recordings/:recordingId/download | Получить URL для скачивания записи | JWT |
| GET | /calls/conflicts | Проверить конфликты расписания | JWT |
| GET | /channels/:channelId/scheduled-calls | Список запланированных звонков | JWT |
| PUT | /scheduled-calls/:id | Обновить запланированный звонок | JWT |
| DELETE | /scheduled-calls/:id | Удалить запланированный звонок | JWT |
| POST | /calls/webhook | Webhook LiveKit (room/participant events) | Internal |

#### Настройки качества LiveKit

| Параметр | Значение |
|----------|----------|
| Audio sample rate | 48000 Hz |
| Audio channels | 2 (stereo) |
| Audio max bitrate | 256 kbps |
| Audio features | autoGainControl, echoCancellation, noiseSuppression |
| DTX (Discontinuous Transmission) | off |
| RED (Redundant Encoding) | on |
| Screen share max bitrate | 8 Mbps |
| Screen share max framerate | 60 fps |
| Guest token validity | 4 часа |

#### Гостевой доступ к звонкам

```
Участник Hub                          Гость (без аккаунта)
     |                                      |
     |-- POST /calls/instant -->            |
     |<-- { invite_code: "hub-abc123" }     |
     |                                      |
     |-- отправляет ссылку /call/abc123 --> |
     |                                      |
     |                POST /calls/guest-join|
     |                { invite_code, name } |
     |                                      |
     |              <-- { token: LiveKit }  |
     |                                      |
     |===== LiveKit WebRTC соединение ======|
```

### Sprint (13 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /projects/:projectId/sprints | Создать спринт | JWT + RBAC |
| GET | /projects/:projectId/sprints | Список спринтов | JWT + RBAC |
| GET | /projects/:projectId/sprints/:sprintId | Получить спринт | JWT + RBAC |
| PUT | /projects/:projectId/sprints/:sprintId | Обновить спринт | JWT + RBAC |
| POST | /projects/:projectId/sprints/:sprintId/start | Запустить спринт | JWT + RBAC |
| POST | /projects/:projectId/sprints/:sprintId/complete | Завершить спринт | JWT + RBAC |
| PUT | /projects/:projectId/sprints/:sprintId/retro | Обновить ретроспективу | JWT + RBAC |
| GET | /projects/:projectId/sprints/:sprintId/burndown | Burndown chart | JWT + RBAC |
| POST | /projects/:projectId/sprints/:sprintId/poker | Создать Poker-сессию | JWT + RBAC |
| POST | /projects/:projectId/sprints/:sprintId/issues | Добавить задачи в спринт | JWT + RBAC |
| GET | /projects/:projectId/velocity | Velocity chart | JWT + RBAC |
| POST | /poker/:sessionId/vote | Проголосовать (Planning Poker) | JWT |
| POST | /poker/:sessionId/reveal | Раскрыть голоса | JWT |
| GET | /poker/:sessionId | Состояние Poker-сессии | JWT |

### Notifications (11 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /notifications | Список уведомлений | JWT |
| GET | /notifications/unread-count | Количество непрочитанных | JWT |
| PATCH | /notifications/:id/read | Пометить прочитанным | JWT |
| PATCH | /notifications/read-all | Пометить все прочитанными | JWT |
| DELETE | /notifications/:id | Удалить уведомление | JWT |
| GET | /notifications/preferences | Настройки уведомлений | JWT |
| PUT | /notifications/preferences | Обновить настройки | JWT |
| GET | /notifications/push/vapid-key | Получить VAPID public key | JWT |
| POST | /notifications/push/subscribe | Подписаться на Web Push | JWT |
| DELETE | /notifications/push/unsubscribe | Отписаться от Web Push | JWT |

### Search (5 эндпоинтов)

> **Полнотекстовый поиск** реализован через PostgreSQL `tsvector` + GIN-индексы и MongoDB `$text`. Результаты объединяются по score.

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /search?q=...&type=...&limit=... | Глобальный поиск | JWT |
| POST | /filters | Создать сохранённый фильтр | JWT |
| GET | /filters | Список сохранённых фильтров | JWT |
| DELETE | /filters/:id | Удалить фильтр | JWT |
| PATCH | /filters/:id/share | Расшарить фильтр | JWT |

#### tsvector-индексы (PostgreSQL)

| Таблица | Колонка | Состав вектора |
|---------|---------|---------------|
| issues | search_vector | `issue_key` (weight A) + `title` (weight A) + `description_text` (weight B) |
| test_cases | search_vector | `case_key` (weight A) + `title` (weight A) + `description` (weight B) |
| users | search_vector | `name` (weight A) + `email` (weight A) + `display_name` (weight B) |
| projects | search_vector | `key` (weight A) + `name` (weight A) + `description` (weight B) |
| file_metadata | search_vector | `filename` (без весов) |
| channels | (inline) | `to_tsvector('simple', name)` с fallback на ILIKE |

Все tsvector-колонки создаются автоматически при старте сервера через `EnsureSearchIndexes`. Конфигурация: `plainto_tsquery('simple', query)` — language-agnostic поиск. Сообщения ищутся через MongoDB `$text` индекс.

### Dashboard (7 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /dashboard/personal | Личный дашборд | JWT |
| GET | /dashboard/project/:projectId | Проектный дашборд | JWT |
| GET | /dashboard/admin | Админский дашборд | JWT (Admin) |
| GET | /dashboard/layout | Получить раскладку | JWT |
| PUT | /dashboard/layout | Сохранить раскладку | JWT |
| GET | /dashboard/templates | Список шаблонов | JWT |
| POST | /dashboard/templates/apply | Применить шаблон | JWT |

### Admin / Audit (7 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /admin/audit | Аудит-лог с фильтрами | JWT (Admin) |
| GET | /admin/limits | Системные лимиты | JWT (Admin) |
| PUT | /admin/limits/:key | Обновить лимит | JWT (Admin) |
| GET | /admin/trash | Системная корзина | JWT (Admin) |
| POST | /admin/trash/:type/:id/restore | Восстановить из корзины | JWT (Admin) |
| POST | /admin/users/:id/deactivate | Деактивировать пользователя | JWT (Admin) |
| POST | /admin/users/:id/reassign | Переназначить задачи | JWT (Admin) |
| DELETE | /admin/users/:id/gdpr | GDPR-удаление пользователя | JWT (Admin) |

### GitHub (9 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /github/webhook | GitHub webhook handler | Signature |
| POST | /projects/:projectId/github/integrations | Создать интеграцию | JWT + RBAC |
| GET | /projects/:projectId/github/integrations | Список интеграций | JWT + RBAC |
| DELETE | /projects/:projectId/github/integrations/:integrationId | Удалить интеграцию | JWT + RBAC |
| GET | /projects/:projectId/github/workflows | Список workflow runs | JWT + RBAC |
| GET | /projects/:projectId/github/workflows/:runId | Детали workflow run | JWT + RBAC |
| GET | /issues/:issueKey/github-links | Список GitHub-ссылок задачи | JWT |
| POST | /projects/:projectId/issues/:issueKey/github/create-branch | Создать ветку из задачи | JWT + task.edit |
| GET | /link-preview | Preview URL (GitHub/external) | JWT |

### GitLab (7 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /gitlab/webhook | GitLab webhook handler | Token |
| POST | /projects/:projectId/gitlab/integrations | Создать интеграцию | JWT + RBAC |
| GET | /projects/:projectId/gitlab/integrations | Список интеграций | JWT + RBAC |
| DELETE | /projects/:projectId/gitlab/integrations/:integrationId | Удалить интеграцию | JWT + RBAC |
| GET | /projects/:projectId/gitlab/pipelines | Список pipelines | JWT + RBAC |
| POST | /projects/:projectId/gitlab/pipelines/trigger | Запустить pipeline | JWT + RBAC |
| POST | /projects/:projectId/issues/:issueKey/gitlab/create-branch | Создать ветку из задачи | JWT + task.edit |

### Storage (4 эндпоинта)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /projects/:projectId/files/upload-url | Получить presigned upload URL | JWT + RBAC |
| POST | /projects/:projectId/files/:fileId/confirm | Подтвердить загрузку | JWT |
| GET | /projects/:projectId/files/:fileId/url | Получить download URL | JWT |
| DELETE | /projects/:projectId/files/:fileId | Удалить файл | JWT + RBAC |

### Backup (4 эндпоинта)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /admin/backups | Создать бэкап | JWT (Admin) |
| GET | /admin/backups | Список бэкапов | JWT (Admin) |
| POST | /admin/backups/:id/restore | Восстановить из бэкапа | JWT (Admin) |
| DELETE | /admin/backups/:id | Удалить бэкап | JWT (Admin) |

### RBAC (6 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /projects/:projectId/roles | Создать кастомную роль | JWT + RBAC |
| GET | /projects/:projectId/roles | Список ролей | JWT + RBAC |
| PUT | /projects/:projectId/roles/:roleId | Обновить роль | JWT + RBAC |
| DELETE | /projects/:projectId/roles/:roleId | Удалить роль | JWT + RBAC |
| POST | /projects/:projectId/roles/:roleId/assign/:userId | Назначить роль | JWT + RBAC |
| DELETE | /projects/:projectId/roles/:roleId/assign/:userId | Снять роль | JWT + RBAC |

### Импорт/Экспорт (6 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | /projects/:projectId/import/issues | Импорт задач (CSV/JSON) | JWT |
| POST | /projects/:projectId/import/jira | Импорт из Jira | JWT |
| GET | /projects/:projectId/export/issues | Экспорт задач (CSV/Excel) | JWT |
| GET | /projects/:projectId/export/full | Экспорт проекта (ZIP) | JWT |
| GET | /projects/:projectId/test-runs/:runId/export/pdf | Экспорт прогона (PDF) | JWT |
| GET | /projects/:projectId/sprints/:sprintId/export/pdf | Экспорт спринта (PDF) | JWT |

### Intranet (v4, ~18 эндпоинтов)

> Новости, справочник сотрудников, kudos, random coffee. Полный список: `docs/wiki/API.md`.

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /intranet/news | Опубликованные новости | JWT |
| POST | /intranet/news | Создать новость | JWT (Admin) |
| GET | /intranet/directory | Дерево отделов → сотрудники | JWT |
| GET | /intranet/kudos/feed | Лента благодарностей | JWT |
| POST | /intranet/kudos | Отправить kudos | JWT |
| PUT | /intranet/coffee/enrollment | Opt-in/out random coffee | JWT |
| POST | /intranet/coffee/run | Запуск matching (admin) | JWT (Admin) |

### CRM (v4, 8 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /crm/contacts | Список контактов (owner; admin=org) | JWT |
| POST | /crm/contacts | Создать контакт | JWT |
| PUT | /crm/contacts/:contactId | Обновить (version) | JWT |
| DELETE | /crm/contacts/:contactId | Soft delete | JWT |
| GET | /crm/pipeline | Kanban сделок (owner; admin=org) | JWT |
| POST | /crm/deals | Создать сделку | JWT |
| PUT | /crm/deals/:dealId | Обновить / переместить stage | JWT |
| DELETE | /crm/deals/:dealId | Soft delete | JWT |

### Hub Base / bases (v4, 10 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /bases | Список баз (workspaces) | JWT |
| POST | /bases | Создать базу | JWT |
| GET | /bases/:baseId/tables/:tableId | Колонки + записи (grid) | JWT |
| POST | /bases/:baseId/tables/:tableId/columns | Добавить колонку | JWT |
| POST | /bases/:baseId/tables/:tableId/records | Добавить строку | JWT |
| PUT | /bases/:baseId/tables/:tableId/records/:recordId | Обновить ячейки | JWT |

### Accounting (v4, 11 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /accounting/accounts | План счетов | JWT (Admin) |
| POST | /accounting/journal | Создать черновик проводки | JWT (Admin) |
| POST | /accounting/journal/:entryId/post | Провести | JWT (Admin) |
| GET | /accounting/trial-balance | ОСВ | JWT (Admin) |
| GET | /accounting/invoices | Список счетов | JWT (Admin) |
| POST | /accounting/invoices/:invoiceId/pay | Оплата → posted проводка | JWT (Admin) |
| GET | /accounting/income-statement | P&L | JWT (Admin) |

### AI Provider Hub (v3, ~18 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /ai/models | Доступные модели (system + BYOK) | JWT |
| PATCH | /ai/preferences | Default model, allow_external_ai | JWT |
| POST | /ai/jql/assist | NL → Hub JQL | JWT |
| POST | /ai/v1/chat/completions | OpenAI-compatible proxy | JWT |
| GET | /admin/ai/providers | Реестр провайдеров | JWT (Admin) |
| POST | /admin/ai/providers/:id/test | Ping провайдера | JWT (Admin) |

### Marketplace (v4, 10 эндпоинтов)

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| GET | /marketplace/listings | Опубликованные объявления | JWT |
| POST | /marketplace/listings | Создать объявление | JWT |
| GET | /marketplace/listings/pending | Очередь модерации | JWT (Admin) |
| POST | /marketplace/listings/:listingId/approve | Одобрить | JWT (Admin) |
| POST | /marketplace/listings/:listingId/sold | Отметить проданным | JWT |

**Итого:** OpenAPI **252** paths (`backend/docs/swagger.json`, 2026-07-28); исторический route inventory ~286+ — см. `docs/wiki/API.md`

---

## WebSocket

### Подключение

```
ws://host/ws  (+ Sec-WebSocket-Protocol: hub.ws.ticket.<ticket>; ?token= forbidden)
```

JWT передаётся через query-параметр, так как WebSocket не поддерживает кастомные заголовки при handshake.

### Константы

| Параметр | Значение |
|----------|----------|
| writeWait | 10 сек |
| pongWait | 60 сек |
| pingPeriod | 54 сек (90% от pongWait) |
| maxMessageSize | 4096 байт |

### Клиентские сообщения (Client -> Server)

```json
{
  "type": "subscribe" | "unsubscribe" | "ping",
  "room": "project:{projectId}" | "channel:{channelId}" | "call:{callId}",
  "payload": {}
}
```

### Серверные сообщения (Server -> Client)

```json
{
  "type": "event" | "pong",
  "event": "issue_created" | "message_new" | "call.started" | "notification.new",
  "room": "project:{projectId}" | "channel:{channelId}",
  "payload": {}
}
```

### NATS -> WebSocket Bridge

| NATS Subject | WS Room | События |
|-------------|---------|---------|
| project.{projectId}.* | project:{projectId} | issue_created, issue_updated, sprint_started и т.д. |
| messenger.channel.{channelId}.* | channel:{channelId} | message_new, message_edited, message_deleted, typing |
| calls.channel.{channelId}.* | channel:{channelId} | call.started, call.ended, call.joined, call.left |
| calls.chat.{callId} | call:{callId} | chat.message (эфемерный чат звонка) |
| auth.session.{userID}.* | -> конкретный user | new_login (уведомление о новом входе) |
| notification.user.{userID} | -> конкретный user | notification.new |

---

## Frontend-архитектура

### Структура маршрутов (TanStack Router)

```
routes/
  __root.tsx                    # Корневой layout
  index.tsx                     # Redirect на dashboard
  login.tsx                     # Страница авторизации
  onboarding.tsx                # Онбординг (4 шага: профиль → тур → тема → проект)
  _authenticated.tsx            # Layout для авторизованных (sidebar + topbar)
  _authenticated/
    dashboard.tsx               # Личный дашборд
    trash.tsx                   # Корзина
    calls.tsx                   # Активный звонок
    messenger.tsx               # Мессенджер layout
    messenger/
      $channelId.tsx            # Конкретный канал
    settings.tsx                # Настройки layout
    settings/
      ...                       # Профиль, тема, уведомления
    admin/
      ...                       # Админ-панель
    projects/
      $projectId/
        overview.tsx            # Обзор проекта
        board.tsx               # Kanban-доска
        issues/
          $issueKey.tsx         # Детальная страница задачи
        sprints/
          ...                   # Sprint management
        tms/
          ...                   # TMS pages
        settings/
          ...                   # Настройки проекта
        dashboard.tsx           # Проектный дашборд
```

### Управление состоянием (Zustand stores)

| Store | Описание |
|-------|----------|
| auth-store | Пользователь, токены, login/logout |
| board-store | Состояние Kanban-доски |
| call-store | Активный звонок, участники, mute/hand |
| dashboard-store | Виджеты, раскладка |
| messenger-store | Каналы, сообщения, unread, typing |
| (removed) notification-store | A10-007: list/unread → TanStack Query (`hooks/use-notifications.ts`) |
| sidebar-store | Состояние сайдбара (collapsed/expanded) |
| sprint-store | Спринты, бэклог |
| timer-store | Таймтрекинг (активные таймеры) |
| tms-store | Тест-кейсы, прогоны |
| ws-store | WebSocket-соединение, подписки |

### Компоненты

```
components/
  ui/           # Shadcn/ui примитивы (Button, Card, Input, Badge, Dialog...)
  layout/       # Sidebar, Topbar, MobileNav
  auth/         # Login form, code input
  tracker/      # Issue card, board, filters
  messenger/    # Channel list, message, composer
  calls/        # Call UI, controls, participants
  dashboard/    # Widgets, layout grid
  notifications/# Bell, notification list
  search/       # Global search (Ctrl+K)
  files/        # Upload, preview
  admin/        # Admin tables, charts

modules/
  admin/        # Admin page logic
  calls/        # Call module logic
  messenger/    # Messenger module logic
  settings/     # Settings pages
  tms/          # TMS module logic
  tracker/      # Tracker module logic
```

### Дизайн-система (CSS-переменные)

```css
/* Поверхности (5 уровней) */
--color-surface-void     /* #08090B — самый тёмный фон */
--color-surface-base     /* #0F1117 — основной фон */
--color-surface-raised   /* #161820 — приподнятый */
--color-surface-card     /* #1A1D27 — карточки */
--color-surface-input    /* #1E2028 — поля ввода */

/* Текст (4 уровня) */
--color-text-t1          /* Основной текст */
--color-text-t2          /* Вторичный */
--color-text-t3          /* Третичный */
--color-text-t4          /* Самый бледный */

/* Границы (3 уровня) */
--color-border-b1        /* Тонкая граница */
--color-border-b2        /* Средняя */
--color-border-b3        /* Яркая */

/* Accent цвета (пользователь выбирает) */
--accent-color           /* red / blue / green / purple / orange */

/* Радиусы */
--radius-sm: 6px
--radius-md: 8px
--radius-lg: 12px
--radius-xl: 16px
--radius-pill: 9999px

/* Тени (4 уровня) */
--shadow-sh1 ... --shadow-sh4

/* Шрифты */
--font-sans: Inter
--font-mono: JetBrains Mono

/* Анимации */
--ease-spring            /* spring/bounce easing */
```

---

## Константы и конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|-----------|----------|-------------|
| SERVER_PORT | Порт HTTP-сервера | 3000 |
| FIBER_IMMUTABLE | Копирование строк из RequestCtx | true |
| DEV_MODE | Dev-эндпоинты (seed, dev-login) | false |
| MAILBOX_DEMO_ENABLED | Ровно `true` включает локальный Mail/Calendar demo provider только при `DEV_MODE=true`; unset/empty/ровно `false` отключает его; любое другое заданное значение невалидно и отклоняется startup wiring | false |
| DATABASE_URL | PostgreSQL connection string | postgresql://hub:hub@localhost:5432/hub?sslmode=disable |
| MONGODB_URL | MongoDB connection string | mongodb://hub:hub@localhost:27017 |
| MONGODB_DATABASE | Имя базы MongoDB | hub |
| REDIS_URL | Redis connection string | redis://localhost:6379/0 |
| NATS_URL | NATS connection string | nats://localhost:4222 |
| S3_ENDPOINT | SeaweedFS S3 endpoint | http://localhost:8333 |
| S3_ACCESS_KEY | S3 access key | hub |
| S3_SECRET_KEY | S3 secret key | hubsecret |
| S3_BUCKET | S3 bucket | hub |
| S3_REGION | S3 region | us-east-1 |
| S3_USE_SSL | SSL для S3 | false |
| JWT_SECRET | Секрет подписи JWT | dev-secret-change-in-production |
| RESEND_API_KEY | API ключ Resend | "" |
| RESEND_FROM_EMAIL | Email отправителя | noreply@hub.local |
| TELEGRAM_BOT_TOKEN | Токен Telegram бота | "" |
| LIVEKIT_URL | URL сервера LiveKit | http://localhost:7880 |
| LIVEKIT_API_KEY | API ключ LiveKit | devkey |
| LIVEKIT_API_SECRET | API секрет LiveKit | devsecret |
| VAPID_PUBLIC_KEY | VAPID public key (base64url) | "" |
| VAPID_PRIVATE_KEY | VAPID private key (base64url) | "" |
| VAPID_SUBJECT | VAPID subject (mailto:) | mailto:admin@hub.local |

### Dev Mail/Calendar composition

`MAILBOX_DEMO_ENABLED=true` действует только вместе с `DEV_MODE=true` и компилируется
только в обычной dev-сборке (`!qa && !release`). В ней Fake-провайдер обслуживает
локальную доставку без внешнего транспорта, но использует общие production-сервисы,
workers и correlation-поток Mail/Calendar. QA- и release-сборки остаются inert для
этой опции. Production-подключение Stalwart/JMAP — отдельная composition-конфигурация
и не заменяется dev Fake-провайдером.

### Лимиты запросов

| Лимит | Max | Window | Описание |
|-------|-----|--------|----------|
| AuthRateLimit | 3 | 10 мин | Авторизация (по IP) |
| APIRateLimit | 100 | 1 сек | Общий API (по IP) |
| FileUploadRateLimit | 10 | 1 мин | Загрузка файлов (по IP) |
| MessageSendRateLimit | 60 | 1 мин | Отправка сообщений (по user) |

Реализация: in-memory sliding window (без зависимости от Redis).

### Настройки JWT

| Параметр | Значение |
|----------|----------|
| Access Token TTL | 15 минут |
| Refresh Token TTL | 30 дней |
| Алгоритм | HS256 |
| Хранение refresh | Redis |

### Настройки WebSocket

| Параметр | Значение |
|----------|----------|
| Write timeout | 10 сек |
| Pong wait | 60 сек |
| Ping period | 54 сек |
| Max message size | 4096 байт |
| Send buffer | 256 сообщений |

### Темы NATS

| Subject Pattern | Описание |
|----------------|----------|
| project.{id}.* | События проекта (issue, sprint, label) |
| messenger.channel.{id}.* | События мессенджера |
| calls.channel.{id}.* | События звонков |
| calls.chat.{id} | Эфемерный чат звонка |
| auth.session.{userId}.* | Событие авторизации |
| notification.user.{userId} | Уведомление пользователю |

---

## Деплой

### Docker Compose (разработка)

```yaml
services:
  postgres:     # PostgreSQL 17          :5432
  mongo:        # MongoDB 8              :27017
  redis:        # Redis 7.4-alpine       :6379
  nats:         # NATS 2.11 + JetStream  :4222/:8222
  seaweedfs:    # SeaweedFS (S3)         :9333/:8333
  livekit:      # LiveKit 1.8            :7880/:7881/:50000-50100/udp
  caddy:        # Caddy 2                :80/:443
```

### Продакшен (CloudForBox + K3s)

```
Домен: 770470.cloud4box.ru

Caddy (prod):
  /api/*       -> backend:3000
  /ws          -> backend:3000 (WebSocket upgrade)
  /lk/*        -> livekit:7880 (strip prefix)
  /*           -> frontend:80 (static build)
```

### Конфигурация Caddy (dev)

```caddyfile
:80 {
  handle /api/*    { reverse_proxy localhost:3000 }
  handle /livekit/* { reverse_proxy localhost:7880 }
  handle /ws       { reverse_proxy localhost:3000 }
  handle           { reverse_proxy localhost:5173 }
}
```

### Конфигурация Caddy (prod)

```caddyfile
770470.cloud4box.ru {
  handle /api/*  { reverse_proxy backend:3000 }
  handle /ws     { reverse_proxy backend:3000 { transport http { versions 1.1 } } }
  handle /lk/*   { uri strip_prefix /lk; reverse_proxy livekit:7880 }
  handle         { reverse_proxy frontend:80 }
}
```

### Бэкап / Восстановление

- **Автоматический**: еженедельный cron (PG pg_dump + Mongo mongodump + файлы S3)
- **Хранение**: последние N бэкапов в S3 (backups/{timestamp}.zip)
- **Восстановление**: Admin панель -> Backups -> Restore
- **Статусы**: in_progress -> completed/failed
- **Ручной**: Admin -> POST /admin/backups

---

*Последнее обновление: 2026-07-28 — auth Magic Link + Telegram notify bind; OpenAPI 252 / inventory ~286; онбординг 4 шага*
