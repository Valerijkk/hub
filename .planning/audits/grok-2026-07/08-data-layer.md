# Ent Schema / CMD Wiring / Data Layer

**Дата:** 2026-07-26  
**Зона:** A8 — `backend/ent/schema` (+ generated `ent/migrate`), `backend/cmd/server`  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Ветка:** `cursor-grok-4.5`  
**Объём:** 82 Ent-схемы (+ `helpers.go`); 19 файлов `cmd/server` (`main.go` + `wiring_*.go`); сверка с `ent/migrate/schema.go` и query-паттернами в `tracker`/`tms`/`auth`/`messenger`/`apigateway`.

## Что проверено

| Область | Фокус |
|---------|--------|
| UUID PK | Все 82 схемы имеют `field.UUID("id", …).Default(uuid.New).Immutable()` |
| `version` / soft delete | Наличие `field.Int("version")`, `deleted_at`, соответствие TZ |
| Indexes / FK | Schema `Indexes()` vs generated `migrate/schema.go`; hot-path queries |
| Cascades | `OnDelete` в migrate (Cascade/SetNull/NoAction); app delete paths |
| Unique constraints | Label/ProjectMember/IssueLink vs Role/Board/Founder gaps |
| CMD wiring | `main` order, `rootCtx`/`addCleanup` LIFO, Listen/Shutdown, Mongo/NATS/DB |
| Outbox | `status+created_at` index, worker lifecycle vs shutdown |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A8-001 | P1 | `ent/schema/issue.go:117-127` (+ `ent/migrate/schema.go:1488-1523`) | Нет индекса на FK `project_issues` и нет композита `(project_issues, deleted_at)` | Hot-path `ListIssues` / board: `HasProjectWith` + `DeletedAtIsNil()` (`tracker/repo_issues.go:172-173`, `handler_issues.go:269-271`). В migrate индексы только на `issue_key`, `issue_number`, `rank`, `created_at`, `priority`, `due_date`, `deleted_at` — **без** project FK. Postgres не индексирует FK автоматически → риск seq scan / filter по слабоселективному `deleted_at` на больших таблицах | `index.Edges("project").Fields("deleted_at")`; опционально `(project, rank)` для Kanban. Atlas migrate |
| A8-002 | P1 | `ent/schema/testcase.go:116-127` (+ `ent/migrate/schema.go:2802-2842`) | Нет индекса на FK `project_test_cases` (+ `deleted_at`) | `ListTestCases`: `HasProjectWith` + `DeletedAtIsNil()` (`tms/ent_repository.go:203-205`). Indexes: `case_key`, matching fields, `status`, `priority`, `deleted_at` — **без** project | `index.Edges("project").Fields("deleted_at")` (+ status/priority по нужде) |
| A8-003 | P1 | `ent/schema/project.go:17-52` | У mutable `Project` нет `version` (есть `deleted_at`, `task_counter`, name/status updates) | TZ: «`version` в каждой изменяемой Ent-схеме». Concurrent update метаданных проекта = last-write-wins; счётчик через `AddTaskCounter` атомарный, но остальное — нет | Добавить `field.Int("version").Default(1)`; `VersionEQ` на UpdateProject |
| A8-004 | P1 | `ent/schema/user.go:17-78`; `channel.go:17-53`; `sprint.go:17-43`; `comment.go:17-38`; `organization.go:20-42`; `board.go:16-36`; `role.go:16-36`; `automationrule.go:17-43`; `call.go:19-61` (+ ~60 других без `version`) | Системный gap: **~14/82** схем с `version` (Issue/TestCase/Hub*/CRM/…); большинство mutable entity без optimistic lock | TZ mandatory нарушен на User/Channel/Sprint/Comment/Org/Board/Role/Call/AutomationRule и др. Гонки update → тихая потеря правок (уже проявлялось в TMS без `VersionEQ`, зона A3) | Mixin `VersionMixin` + rollout по изменяемым сущностям; junction/append-only (Outbox, AuditLog, ActivityLog) можно исключить явно в ADR |
| A8-005 | P1 | `ent/schema/channel.go:17-53`; `messenger/ent_repository.go:86-88`; `ent/migrate/schema.go:555-560` | Channel: нет `deleted_at`; `DeleteChannel` = hard `DeleteOneID`; FK `channel_members` → `OnDelete: NoAction` | TZ soft-delete (корзина 30д). Hard delete при существующих members → FK violation (комментарий «и связанные данные» врёт — members/pins/polls не чистятся). Нет Cascade и нет soft-delete | Добавить `deleted_at`; soft-delete в repo; либо tx: delete children + channel; либо `entsql.OnDelete(Cascade)` на members/invites/pins |
| A8-006 | P1 | `ent/schema/comment.go:17-38`; `tracker/repo_comments.go:91` | Comment без `deleted_at`; delete = hard `DeleteOneID` | Нарушение TZ soft-delete для user content; нет корзины/восстановления | `deleted_at` + soft delete; фильтр `DeletedAtIsNil` в list |
| A8-007 | P1 | `cmd/server/wiring_messenger.go:26-31` | Mongo connect без `addCleanup(Disconnect)` | PG/Redis/NATS закрываются в LIFO cleanup; Mongo connection pool остаётся до kill процесса — неполный graceful shutdown | `d.addCleanup(func() { _ = mongoDB.Client().Disconnect(context.Background()) })` |
| A8-008 | P1 | `cmd/server/wiring_run.go:16-22` | При ошибке `app.Listen` вызывается `os.Exit(1)` | `defer d.runCleanups()` **не выполняется** на `os.Exit` → NATS/Redis/PG/Mongo/consumers не закрываются | Логировать ошибку + `rootCancel()` / return из `runServer` без `os.Exit`; либо `os.Exit` только после явного cleanup |
| A8-009 | P2 | `ent/schema/board.go` (нет `Indexes()`); `ent/migrate/schema.go:317-329` | `boards` — FK на project **без** индекса | `ListBoards`: `HasProjectWith` (`tracker/repo_boards.go:37`). BoardsTable в migrate — только PK+FK, Indexes пуст | `index.Edges("project")`; unique partial на `is_default` per project при бизнес-правиле |
| A8-010 | P2 | `ent/schema/role.go` (нет `Indexes()` / unique name); `ent/migrate/schema.go:2507-2520` | Roles: нет индекса на project FK; нет unique `(project, name)` | List/lookup ролей по проекту без индекса; дубли имён ролей в одном проекте допустимы на уровне БД | `index.Edges("project").Fields("name").Unique()` |
| A8-011 | P2 | `ent/schema/webhookdelivery.go:52-57`; `ent/migrate/schema.go:3218-3228`; evidence `apigateway/repository.go:179-181` | Нет индекса на FK `webhook_endpoint_deliveries` | `ListWebhookDeliveries` фильтрует `HasEndpointWith(endpointID)` — только indexes `delivered_at`, `event_type` | `index.Edges("endpoint").Fields("delivered_at")` |
| A8-012 | P2 | `ent/schema/webhookendpoint.go:25-27` | `secret` без `.Sensitive()` (в отличие от GitHub/GitLab/AI keys) | Ent/логи/debug могут сериализовать HMAC secret | `field.String("secret").Sensitive().NotEmpty()` |
| A8-013 | P2 | `ent/schema/user.go:146-152`; evidence `auth/repository.go:185-188` | Нет индекса под purge pending: `status=pending` + `pending_since < cutoff` | `PurgeExpiredPending` / list pending сканируют по `status` index alone; `pending_since` не покрыт | `index.Fields("status", "pending_since")` |
| A8-014 | P2 | `ent/schema/user.go:48-50` | TZ: один `founder` на org — в схеме нет unique/partial constraint | Инвариант только в app (`CreateOrganizationWithFounder`); гонка/баг могут создать второго founder | Partial unique index `WHERE system_role = 'founder'` (один ряд на инстанс) или check в DB |
| A8-015 | P2 | `ent/schema/sprint.go`; `testsuite.go`; `filemetadata.go:17-46` | Soft-delete gap / inconsistency: Sprint/TestSuite без `deleted_at`; FileMetadata имеет soft-delete **без** `version` | Неединообразная корзина; sprint/suite hard-delete vs Issue/Project/TestCase soft | Выровнять политику: soft-delete для user-visible entities; version на FileMetadata если concurrent confirm/delete |
| A8-016 | P2 | `cmd/server/wiring_notification.go:21-23`, `53-55` | `notifConsumer.Start` / `automationConsumer.Start` при ошибке только `slog.Error`, процесс живёт дальше | Сервер «healthy», но outbox/NATS events не потребляются — silent data lag | `fatal` / `os.Exit` после cleanup, или readiness probe = consumers started |
| A8-017 | P2 | `cmd/server/wiring_notification.go:40-42`; `notification/telegram_bind.go:176-184` | Telegram bind polling без `addCleanup`; `GetUpdates` Timeout=30s | После SIGTERM горутина может до 30s держать процесс (если main ждёт только HTTP Shutdown + sync cleanups; long-poll не в cleanup) | Передавать короткий timeout на shutdown; cancel HTTP client; зарегистрировать Stop |
| A8-018 | P2 | `cmd/server/wiring_run.go:29-36` + `wiring_infra.go:54-55` | Нет явного wait outbox/cron после `rootCtx` cancel перед `db.Close()` | Outbox `ProcessLoop` выходит по ctx, но mid-`processBatch` может ещё писать в PG, пока cleanup уже закрывает DB (узкое окно) | `sync.WaitGroup` / `outboxWorker.Stop()` в cleanup **до** `db.Close`/`nc.Close` |
| A8-019 | P3 | `ent/schema/filemetadata.go:50-58` vs `ent/migrate/schema.go:952-977` | Schema Edges объявляют только `uploader`/`project`; migrate имеет ещё `issue_files`, `test_result_files` (от `Issue.To("files")` / `TestResult.To("files")`) | Односторонние edges работают, но schema-файл врёт maintainers; риск потерять FK при ручной правке | Явные `edge.From("issue"…)` / `From("test_result"…)` в FileMetadata |
| A8-020 | P3 | `ent/schema/issue.go:119-121` | Redundant non-unique index на `issue_key` при `field.Unique()` | Дублирующий index в PG (unique constraint уже даёт index) | Убрать `index.Fields("issue_key")`; оставить unique field |
| A8-021 | P3 | `ent/schema/issue.go:22-24`, `120-121` | `issue_number` индексирован глобально non-unique; нет unique `(project, issue_number)` | Корректность держится на `issue_key` unique + app counter; номер сам по себе не защищён constraint’ом per-project | `index.Edges("project").Fields("issue_number").Unique()` |

## Чисто

- **UUID v4 PK** на всех 82 схемах — единый паттерн `Default(uuid.New).Immutable()`.
- **UTC timestamps** через `nowUTC()` в `helpers.go:14`.
- **Issue / TestCase** имеют `version` + `deleted_at` + индексы на soft-delete; tracker `UpdateIssue` использует `VersionEQ` (`repo_issues.go:287-289`).
- **OutboxEvent** индекс `(status, created_at)` (`outboxevent.go:47-50`) совпадает с polling worker.
- **Notification** индексы `(user_id, is_read)`, `(user_id, created_at)` — адекватны inbox queries.
- **Уникальности junction**: `ChannelMember(channel,user)`, `ProjectMember(user,project)`, `IssueLink(source,target)`, `Label(project,name)`, `PollVote`, `UserBlock` — ок.
- **M2M Cascade** только на join-таблицах (`issue_labels`, `project_teams`, …) — корректный Ent default.
- **CMD order** (`main.go:22-33`): bootstrap → infra → fiber → auth → core → … → run; единый `signal.NotifyContext` (`wiring_bootstrap.go:70-79`).
- **Cleanup LIFO**: consumers/bridges регистрируются после infra → Stop до `nc.Close`/`db.Close` (`wiring_deps.go:124-127`).
- **Shared PG pool** для health: `rawDB` = `drv.DB()` (`database/postgres.go:31-35`) — без второго pool.
- **Fiber BodyLimit 25MB**, trusted proxies, recover, rate limits на rootCtx (`wiring_fiber.go`).

## Оценка зоны: 6/10

**Обоснование:** Фундамент (UUID, UTC, outbox index, частичный soft-delete Issue/Project/TestCase, аккуратный wiring order + LIFO cleanup consumers) выглядит зрело. Оценка снижена системными TZ-gap’ами: **~80% схем без `version`**, hot-path **отсутствие индексов на project FK** для issues/test_cases (с прямым evidence из List-запросов), **Channel/Comment hard-delete** против политики корзины + FK NoAction trap, **Mongo без Disconnect** и **Listen `os.Exit` минуя defer**. После индексов `(project, deleted_at)`, VersionMixin на mutable entities, soft-delete Channel/Comment и полного shutdown cleanup зона могла бы быть ~8/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 8 |
| P2 | 10 |
| P3 | 3 |
| **Total** | **21** |

**Score: 6/10**
