# Hub Backend — Architectural Audit (post-Phase 14)

Дата: 2026-05-15. Контекст: после декомпозиции god-objects и выноса `*ent.Client` из 3 сервисов через Repository. Цель — проверить целостность слоёв и отсутствие drift.

## TL;DR

- **Score: 72 / 100** — архитектура цельная, нет циклов и фатальных нарушений; основной долг — недо-доведённая миграция на Repository pattern (3 из 17 сервисов отрефакторены, 10 ещё держат `*ent.Client`).
- `go vet ./...` — clean (exit 0). `go build ./...` — clean (нет циклов импорта).
- Build-теги `//go:build release` / `!release` корректно расставлены только в `internal/dev/` (handler.go, seed.go, seed_data.go, seed_test.go, handler_release.go) — где это и требовалось.
- **Repository pattern coverage:** 11 / 17 сервисных пакетов = **~65%**. Полностью на Repository: messenger, tms, tracker, project, sprint, calls, search, audit, notification, github, gitlab. Без Repository: auth, dashboard, backup, imex, storage, apigateway, rbac. (rbac имеет свой `RBACService`+`PermissionCache`, но `*ent.Client` хранит напрямую.)

## 1. Layering integrity — нарушения

### 1.1 Handlers → DB напрямую (CRITICAL)
Два хендлера обращаются к `ent.Client` напрямую, минуя сервис:
- `internal/auth/handler.go:82` — `h.db.User.Query().Where(user.EmailEQ(req.Email)).Exist(ctx)` (проверка существования email при регистрации).
- `internal/auth/handler.go:503` — `h.db.User.Get(ctx, userID)` (GetMe / loadUser).
- `internal/dev/handler.go:102` — `h.db.User.Query()...Only(ctx)` (dev-login, под `//go:build !release` — менее критично).

Это единственные настоящие layering violations. `ent.IsNotFound(err)` в других хендлерах (`calls/handler_recording.go`, `storage/handler.go`, `dev/handler.go`) — это sentinel-error check, не direct DB access, считать нарушением не стоит.

### 1.2 Service → service imports (LOW)
Сервисы не импортируют друг друга «как сервисы»; импорты ограничены:
- `audit/service.go` → `auth.InvalidateUserStatusCache(ctx, redis, userID)` — это одна утилитная функция-инвалидатор Redis-кэша (не вызов AuthService). Допустимо, но в идеале вынести в `platform/authcache`.
- `notification/consumer.go` → `tracker.IssueAssignedEvent`, `tms.TestRunCompletedEvent`, `auth.SessionCreatedEvent` — это импорт event-DTO для разбора NATS-сообщений, не вызовы сервисов. Архитектурно OK, но событийные DTO желательно вынести в общий `platform/events`.
- `ws/handler.go` → `auth` (RequireAuth middleware). OK.
- `*/handler*.go` → `rbac` — middleware-импорт permission-guard'ов, корректно.

### 1.3 Handler → service binding — концретный struct, не интерфейс
Все handler'ы держат сервис как **указатель на конкретный struct** (`*MessengerService`, `*TrackerService`, `*CallService`, `*ProjectService` и т. д.). Интерфейс между handler и service не публикуется ни в одном модуле. Это не критично (handler — тонкий шов), но усложняет mock'и в handler-тестах. На уровне service→repo интерфейсы есть (см. §5).

## 2. Ent leakage — где `*ent.Client` живёт за пределами repository.go

После Phase 14 утечка осталась в **10 сервисных файлах** (плюс auth-handler — см. §1.1):

| Пакет | Файл | Держит `*ent.Client`? | Доступ `s.db.X.Query()` |
|---|---|---|---|
| auth | service.go, service_login.go, service_session.go, service_profile.go, service_registration.go | да (по копии) | да |
| dashboard | service.go | да | да (20 мест) |
| backup | service.go | да | да |
| imex | service.go | да | да |
| storage | service.go, orphan_cleanup.go | да | да |
| rbac | service.go, middleware.go | да | да |
| apigateway | apikey.go, webhook.go (×2 структуры) | да | да |
| notification | service.go, push.go | да (3 места) | да |
| project | service.go | да (4 места — issue_link orphan cleanup) | да |
| github | service.go (`entDB`, vararg) | да (2 места — TMS cross-link) | да |
| gitlab | service.go (`entDB`, vararg) | хранится, но 0 использований | — |
| audit | retention.go (отдельный сервис, не AuditService) | да | да |

`*ent.Tx` встречается только внутри repository-файлов и в `outbox/publisher.go` (корректно).

`rbac/middleware.go` принимает `*ent.Client` в `RequireSystemRole` / `RequireAdmin`, что вызывает создание ещё одной ent-зависимости в `cmd/server/main.go:559` (`backupHandler.RegisterRoutes(api, requireAuth, rbac.RequireAdmin(db))`). После рефакторинга rbac в Repository эта точка тоже исчезнет.

## 3. Dependency cycles

`go build ./...` exit 0 — циклов нет. Ручная проверка обратных импортов:
- `tracker` импортирует `search` (для JQL-парсинга) — `search` не импортирует `tracker`. OK.
- `notification/consumer.go` импортирует `tracker`, `tms`, `auth` (event DTOs) — обратных импортов из этих пакетов в `notification` нет.
- `audit` → `auth` (1 функция), обратной зависимости нет.

Граф направленный, ацикличный.

## 4. Singleton / DI patterns

`cmd/server/main.go` (629 строк) — единственная точка композиции. Все инфраструктурные клиенты создаются ровно по одному разу:
- `db := database.NewPostgresClient` (1 экземпляр, передаётся всем сервисам)
- `rdb := redisclient.NewClient` (1 экземпляр)
- `nc := natsclient.NewClient` (1 экземпляр)
- `mongoDB := database.NewMongoClient` (1 экземпляр)
- `rawDB := database.OpenRawDB` (отдельный `*sql.DB` для full-text search, оправдано — Ent не даёт raw ts_query)
- `integrationCipher` (1 экземпляр)

Дублей `http.Client` / `redis.Client` нет. Каждый cron/worker запускается через `go func() { ticker := ... }` с общим `rootCtx` — паттерн ME-10, повторён ~7 раз; стоит вынести в `platform/cron` helper.

## 5. Interface boundaries

На границе **service → repository** интерфейсы публикуются во всех 11 модулях с Repository:
- `messenger.Repository` — composed из 11 sub-интерфейсов (ChannelRepository, MessageRepository, PinRepository, …). Чисто.
- `tracker.Repository` — composed из 10 sub-интерфейсов. Чисто.
- `tms.Repository` — composed из 8 sub-интерфейсов. **Pitfall:** `tms.Repository` экспортирует `EntClient() *ent.Client`, что снова утекает ent наружу для `tracker.repository.go:153` (`GetDB() *ent.Client`) и `tms.repository.go:161` — это намеренные escape-hatch'и для legacy кода, но они отменяют пользу абстракции.

На границе **handler → service** интерфейсов нет — везде конкретные `*Service`-указатели (см. §1.3).

## 6. Repository pattern coverage

11 / 17 сервисных пакетов = **64.7%**.

С Repository:
audit, calls, github, gitlab, messenger, notification, project, search, sprint, tms, tracker.

Без Repository (держат `*ent.Client` напрямую):
auth (5 сервис-файлов), dashboard, backup, imex, storage, apigateway (apikey + 2 webhook-сервиса), rbac.

Особый случай: `audit/retention.go::RetentionService` — отдельный сервис в том же пакете, что и `AuditService`; держит `*ent.Client` и `*mongo.Database` напрямую, не использует `audit.Repository`. Кандидат на следующий Phase.

## 7. Top-10 раздутых файлов (LOC, без `_test.go`)

| # | LOC | Файл | Комментарий |
|---|---|---|---|
| 1 | 932 | `internal/tms/ent_repository.go` | Ent CRUD по 8 сущностям TMS — нормально для repo |
| 2 | 734 | `internal/dev/seed.go` | dev-only (`//go:build !release`), не попадает в prod-бинарь |
| 3 | 676 | `internal/project/handler.go` | **Раздут** — кандидат на split |
| 4 | 673 | `internal/search/service.go` | Поиск+JQL+глобальный — можно дробить |
| 5 | 642 | `internal/dashboard/service.go` | Aggregations; нет Repository — оба фактора в одном файле |
| 6 | 629 | `cmd/server/main.go` | DI-кишки, ожидаемо; cron-хелперы вынести |
| 7 | 626 | `internal/notification/consumer.go` | NATS subjects → DTO разбор по 6 типам событий |
| 8 | 572 | `internal/messenger/ent_repository.go` | Repo, норм |
| 9 | 569 | `internal/gitlab/service.go` | GitLab API + webhooks в одном файле — split |
| 10 | 566 | `internal/messenger/handler_channels.go` | Уже разбит на handler_*.go; OK |

`cmd/server/main.go` после Phase 14 — 629 LOC, не раздут аномально (для DI-точки на 21 модуль это плотно). Cron-горутины (`outbox`, `orphan-cleanup`, `purge-expired`, `retention`, `deadlines`, `scheduled-calls`, `auto-backup`) занимают ~80 LOC шаблонного `for { select { case <-ticker.C: ... }}` — вынести в `platform/cron.RunInterval(ctx, dur, fn)`.

## 8. Top-пакеты по LOC

1. tms — 7476 (29 файлов)
2. messenger — 4972 (25 файлов)
3. tracker — 4723 (34 файла)
4. calls — 2898 (14 файлов)
5. project — 2466 (9 файлов)
6. notification — 2139
7. auth — 2094 (13 файлов, разбиение Phase 14 на 5 service_*.go сработало)
8. github — 1593
9. audit — 1549
10. search — 1533

## 9. `go.mod` / build tags

- `go.mod`: 1 module, Go 1.25.0, 30 direct deps, ~150 indirect. Все direct — production-релевантны. Unused-deps аудит не выполнен (`go mod why` для каждой не запускался, но `go build` clean).
- `//go:build release` / `!release` корректно расставлены в `internal/dev/` (5 файлов). Других build-тегов в проекте нет — для бинаря Hub это нормально (нет mock/integration-разделения).

## 10. Score 0–100

| Категория | Вес | Балл | Обоснование |
|---|---|---|---|
| Layering integrity | 20 | 14 | 3 handler-violations (auth, dev) |
| Ent leakage | 20 | 10 | 10 сервисов всё ещё держат ent напрямую |
| Dep cycles | 15 | 15 | Циклов нет |
| DI / singletons | 10 | 9 | Cron-boilerplate, иначе чисто |
| Interface boundaries | 10 | 7 | repo-интерфейсы есть, handler→service — нет |
| Repository coverage | 15 | 10 | 65%, ещё 6+ пакетов на очереди |
| Package/file size | 5 | 4 | Топ-файлы оправданы кроме project/handler.go |
| go.mod / build tags | 5 | 5 | OK |
| **Итого** | **100** | **72** | Цельная архитектура с явным остатком долга |

## Что я бы добавил в Phase 15 (без фикса сейчас)

1. **Repository для auth** — самый «вонючий» остаток: 5 сервис-файлов держат `*ent.Client`, плюс хендлер ходит в `h.db.User`. Это даст автоматический фикс layering violation §1.1.
2. **Repository для apigateway** (3 структуры: `APIKeyService`, `WebhookService`, `WebhookDispatcher` — все на `*ent.Client`).
3. **Repository для dashboard** — 20 мест прямого `s.db.X.Query()`, самый «грязный» по count.
4. **Repository для rbac** — устранит `*ent.Client` в middleware и в `cmd/server/main.go:559`.
5. **Убрать `EntClient()` / `GetDB()` escape-hatch'и** из `tms.Repository` и `tracker.Repository` — они отменяют выгоду от Repository.
6. **Вынести cron-boilerplate** из `cmd/server/main.go` в `platform/cron`.
7. **Вынести event-DTO** (`IssueAssignedEvent`, `TestRunCompletedEvent`, `SessionCreatedEvent`) в `platform/events`, чтобы `notification.consumer.go` не импортировал `tracker`/`tms`/`auth`.
