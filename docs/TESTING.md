# Hub — Тестовая документация

> **Обновлено:** 2026-07-28 · Стратегия: `docs/test_plan.md` · Coverage: `.planning/COVERAGE.md` · Quality audit: `.planning/audits/2026-06/AUDIT-2026-06-22-quality-processes.md`  
> **Edge/Corner backlog:** [`.planning/archive/july-2026/EDGE-CORNER-CATALOG.md`](../.planning/archive/july-2026/EDGE-CORNER-CATALOG.md) — Happy/Neg/Edge/Corner по доменам (источник будущих TC; CAL-SYNC C1–C15)

---

## Стратегия тестирования

### Уровни

| Уровень | Инструменты | Расположение | CI job |
|---------|-------------|--------------|--------|
| Unit (Backend) | Go testing + testify | `backend/testframework/{domain}/` | `backend-test` |
| Unit (Frontend) | Vitest + RTL | `testframework/stages/unit/frontend-test/domains/` | `frontend-test` |
| API / handler | httptest | `backend/testframework/` | в unit |
| E2E | Playwright + POM | `testframework/stages/e2e/domains/` | `e2e-smoke`, `e2e-full` |
| Build | tsc + vite | `frontend/` | `build` |
| Lint | Biome | `frontend/src` | `frontend-test` (gate) |

### Принципы

1. **Backend CI:** `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...`
2. **Coverage floor:** statement coverage `internal/` **83.0%** (2026-06-22 Docker baseline; ≥82% gate); security packages **≥80%**
3. **Auth тесты:** Magic Link only (v2.5) — Telegram login endpoints **удалены**
4. **E2E:** 20 spec files; config `frontend/playwright.config.ts` (см. `docs/test_plan.md`)
5. **Triage:** spec-first — `.cursor/skills/test-triage/SKILL.md`

### Актуальные счётчики (2026-07-29)

| Слой | Метрика |
|------|---------|
| Backend test files (`backend/testframework/`) | **805** `_test.go` (0 в `internal/` после TF-M08; measured 2026-07-29) |
| Backend `export_hooks.go` | **39** (`//go:build hubtest`; measured 2026-07-29) |
| Backend `internal/*_test.go` | **0** |
| Frontend Vitest | **167** files (`stages/unit/frontend-test/`; measured 2026-07-29; 61 `.ts` + 106 `.tsx`) |
| Playwright | **20** spec files (`stages/e2e/domains/`) |
| Overall `internal/` coverage | **83.0%** (2026-06-22 Docker baseline) |

---

## Backend-тесты (Go)

Все Go-тесты живут в `backend/testframework/{domain}/` (белый ящик через `export_hooks.go` с `//go:build hubtest`). В `backend/internal/` **нет** `*_test.go`. Каталог доменов и карта CI — `testframework/PIPELINE.md` / `docs/test_plan.md`.

Покрытие `internal/` меряется флагом `-coverpkg=./internal/...` при прогоне `./testframework/...` (job `coverage-gate`), не путём `./internal/...`.

### Запуск тестов

```bash
# Backend — как CI (только testframework; в internal/ тестов нет)
cd backend
go generate ./ent
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Конкретный модуль
go test -tags hubtest ./testframework/auth/...
go test -tags hubtest -v ./testframework/tms/...

# С покрытием кода internal (профиль вне repo — rule 45)
go test -tags hubtest -coverpkg=./internal/auth/... -coverprofile="${TMPDIR:-/tmp}/hub-cov.out" ./testframework/auth/...

# Prod build
go build -tags release -o bin/server ./cmd/server
```

### Frontend

```bash
cd frontend
npx tsc --noEmit          # типы
npm run lint              # Biome gate
npx vitest run            # unit
npm run build             # полная сборка
npm run test:e2e          # Playwright (нужен backend для full)
```

Vite build проверяет: импорты, tree-shaking, Tailwind 4, TanStack Router codegen.

### CI jobs (кратко)

| Job | Ключевые шаги |
|-----|---------------|
| `backend-test` | Go 1.25, Mongo 8, `-tags hubtest -race` |
| `frontend-test` | Node 22, tsc, vitest, biome |
| `e2e-smoke` | login + a11y + critical-gate (без backend) |
| `e2e-full` | `scripts/ci-e2e-backend.sh` + 20 specs |

---

## API-тесты по модулям

### Формат: Ожидаемые HTTP-коды

#### Auth (Magic Link only — v2.5+)

> Telegram login (`/auth/telegram*`) **удалён** (ADR-AUTH-01). Telegram — только notify bind.

| Endpoint | 200 | 400 | 401 | 404 | 429 |
|----------|-----|-----|-----|-----|-----|
| POST /auth/magic-link | Код отправлен | Невалидный email | — | — | Rate limit (3/10m) |
| POST /auth/magic-link/verify | Токены выданы | Неверный код | — | — | Rate limit |
| POST /auth/refresh | Новый access token | Невалидный refresh | 401 expired | — | — |
| GET /auth/me | Профиль | — | Не авторизован | — | — |
| PUT /auth/me | Обновлён | Невалидные данные | Не авторизован | — | — |
| POST /auth/logout | Выход | — | Не авторизован | — | — |
| GET /auth/sessions | Список сессий | — | 401 | — | — |
| DELETE /auth/sessions/:id | Сессия отозвана | — | 401 | — | — |

#### Projects

| Endpoint | 200/201 | 400 | 401 | 403 | 404 |
|----------|---------|-----|-----|-----|-----|
| POST /projects | Создан | Невалидные данные | Нет токена | - | - |
| GET /projects | Список | - | Нет токена | - | - |
| GET /projects/:id | Проект | - | Нет токена | Не участник | Не найден |
| PUT /projects/:id | Обновлён | Невалидные данные | Нет токена | Нет прав | Не найден |
| DELETE /projects/:id | Удалён | - | Нет токена | Нет прав (Owner) | Не найден |
| POST /projects/:id/members | Добавлен | Невалидные данные | Нет токена | Нет прав | Не найден |

#### Tracker

| Endpoint | 200/201 | 400 | 401 | 403 | 404 | 409 |
|----------|---------|-----|-----|-----|-----|-----|
| POST /projects/:id/issues | Создана | Невалидные данные | Нет токена | RBAC | Проект не найден | - |
| PUT /projects/:id/issues/:key | Обновлена | Невалидные данные | Нет токена | RBAC | Задача не найдена | Version conflict |
| DELETE /projects/:id/issues/:key | Удалена | - | Нет токена | RBAC | Не найдена | - |

#### Messenger

| Endpoint | 200/201 | 400 | 401 | 403 |
|----------|---------|-----|-----|-----|
| POST /messenger/channels | Создан | Невалидные данные | Нет токена | - |
| POST /messenger/channels/:id/messages | Отправлено | Пустое сообщение | Нет токена | Не участник |
| POST /messenger/channels/:id/join | Присоединён | - | Нет токена | Приватный канал |
| POST /messenger/dm | DM создан/получен | Невалидный userId | Нет токена | - |

#### TMS

| Endpoint | 200/201 | 400 | 401 | 403 | 404 |
|----------|---------|-----|-----|-----|-----|
| POST /projects/:id/tms/cases | Создан | Невалидные данные | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/runs | Создан | Невалидные данные | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/allure/import | Импортировано | Невалидный JSON | Нет токена | RBAC | Проект |
| POST /projects/:id/tms/webhook/allure | Принято | Невалидный формат | API Key | - | Проект |

#### Calls

| Endpoint | 200/201 | 400 | 401 | 404 |
|----------|---------|-----|-----|-----|
| POST /calls | Создан | Нет channelId | Нет токена | Канал не найден |
| POST /calls/instant | Создан | - | Нет токена | - |
| POST /calls/:id/join | LiveKit token | - | Нет токена | Звонок не найден |
| POST /calls/guest-join | LiveKit token | Нет invite_code/guest_name | - | Звонок не найден |
| POST /calls/:id/recording/start | Запись начата | - | Нет токена | Звонок не найден |

---

## Тест-кейсы по модулям

> Каждый тест-кейс (TC-*) связан с автоматизированными unit-тестами. Колонка **Автотест** указывает Go test function, покрывающую данный сценарий.

### AUTH — Авторизация (Magic Link)

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-AUTH-01 | Magic Link: запрос кода | Email в org | POST /auth/magic-link {email} | 200, код отправлен | handler validation tests |
| TC-AUTH-02 | Magic Link: верификация | Код из email | POST /auth/magic-link/verify | 200, access + refresh | `TestVerifyMagicLink_*` |
| TC-AUTH-03 | Magic Link: неверный код | — | verify с wrong code | 400/401 | `ConstantTimeCompare` path |
| TC-AUTH-04 | Refresh: обновление токена | Валидный refresh | POST /auth/refresh | 200, новый access | JWT round-trip tests |
| TC-AUTH-05 | Refresh: истёкший токен | Просроченный refresh | POST /auth/refresh | 401 | `TestJWTExpired` |
| TC-AUTH-06 | Профиль: получение | Авторизован | GET /auth/me | 200 | handler tests |
| TC-AUTH-07 | Профиль: обновление | Авторизован | PUT /auth/me | 200 | handler validation |
| TC-AUTH-08 | Сессии: список | Авторизован | GET /auth/sessions | 200 | — (E2E/manual) |
| TC-AUTH-09 | Сессии: отзыв | 2+ сессии | DELETE /auth/sessions/:id | 200 | — (E2E/manual) |
| TC-AUTH-10 | Rate limit | — | 4 auth запроса за 10 мин | 429 на 4-м | manual / middleware test |
| TC-AUTH-11 | Онбординг | onboarding=false | POST /auth/me/onboarding | 200 | `onboarding-profile.spec.ts` |
| TC-AUTH-12 | Logout | Авторизован | POST /auth/logout | 200, refresh revoked | `logout.spec.ts` |
| TC-AUTH-13 | Magic Link: OTP не в slog (A14-006) | Dev/Resend paths | SendMagicLink + VerifyCode; capture slog | raw OTP и `"code"` отсутствуют; email masked | `TestMagicLink_OTPNotLogged_A14_006`, `TestMaskEmailMinimized_A14_006` |

### PROJ — Проекты

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-PROJ-01 | Создание проекта | Авторизован | POST /projects {name, key} | 201, проект + workflow + доска + канал | `TestCreateProject_Success` |
| TC-PROJ-02 | Дубликат ключа | Проект HUB существует | POST /projects {key: "HUB"} | 400/409, ключ занят | `TestCreateProject_DuplicateKey` |
| TC-PROJ-03 | Список проектов | Участник 2 проектов | GET /projects | 200, 2 проекта | — (ручной) |
| TC-PROJ-04 | Архивирование | Owner проекта | POST /projects/:id/archive | 200, status=archived | `TestArchiveProject_SetsStatusArchived` |
| TC-PROJ-05 | Удаление в корзину | Owner проекта | DELETE /projects/:id | 200, deleted_at set | `TestDeleteProject_SoftDelete` |
| TC-PROJ-06 | Восстановление | Проект в корзине | POST /projects/:id/restore | 200, deleted_at null | — (ручной) |
| TC-PROJ-07 | Invite: создание | Owner/Developer | POST /projects/:id/invites | 201, код + ссылка | `TestJoinByInvite_ValidCode` |
| TC-PROJ-08 | Invite: присоединение | Валидный invite-код | POST /invites/join {code} | 200, участник добавлен | `TestJoinByInvite_ValidCode`, `TestJoinByInvite_ExpiredCode` |
| TC-PROJ-09 | Участник: смена роли | Owner | PUT /projects/:id/members/:userId {role: "tester"} | 200 | — (ручной) |
| TC-PROJ-10 | Участник: удаление | Owner | DELETE /projects/:id/members/:userId | 200, удалён | `TestRemoveMember_LastOwner` |

### TRCK — Трекер задач

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-TRCK-01 | Создание задачи | Участник проекта | POST issues {title, type, priority} | 201, issue_key = HUB-1 | `TestCreateIssue` |
| TC-TRCK-02 | Автоинкремент | HUB-1 существует | POST issues {title, type} | 201, issue_key = HUB-2 | `TestCreateIssueAllTypes` |
| TC-TRCK-03 | Обновление с version check | version=1 | PUT issues/:key {title, version: 1} | 200, version=2 | `TestUpdateIssueOptimisticLocking` |
| TC-TRCK-04 | Конфликт версий | version=2 в БД | PUT issues/:key {title, version: 1} | 409, conflict | `TestUpdateIssueOptimisticLocking` |
| TC-TRCK-05 | Комментарий с @mention | Задача существует | POST comments {body, mentions: [userId]} | 201, уведомление отправлено | — (ручной) |
| TC-TRCK-06 | Связь задач | 2 задачи | POST issues/:key/links {target, type: "blocks"} | 201 | `TestCreateIssueLink` |
| TC-TRCK-07 | Time tracking: manual | Задача существует | POST issues/:key/time {minutes: 60} | 201, time_spent += 60 | — (ручной) |
| TC-TRCK-08 | Time tracking: timer | Задача существует | POST timer/start -> POST timer/stop | 201, лог с длительностью | — (ручной) |
| TC-TRCK-09 | JQL поиск | Задачи существуют | POST issues/jql {query: "type = bug AND priority = high"} | 200, отфильтрованные задачи | `TestParseJQLSimple`, `TestParseJQLMultipleConditions`, `TestToEntPredicates` |
| TC-TRCK-10 | Workflow: смена статуса | Переход разрешён | PUT issues/:key {status_id} | 200, статус изменён | `TestMoveIssueValidTransition` |
| TC-TRCK-11 | Workflow: запрещённый переход | Переход не задан | PUT issues/:key {status_id} | 400, переход не разрешён | `TestMoveIssueInvalidTransition` |
| TC-TRCK-12 | Мягкое удаление | Задача существует | DELETE issues/:key | 200, deleted_at set | — (ручной) |

### MSG — Мессенджер

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-MSG-01 | Создание публичного канала | Авторизован | POST channels {type: "public", name} | 201 | `TestCreateChannel_WithProjectID`, `TestCreateChannel_ReadOnlyFlag` |
| TC-MSG-02 | DM дедупликация | DM между A и B существует | POST /dm {userId: B} | 200, тот же канал | `TestCreateDM_Dedup` |
| TC-MSG-03 | Отправка сообщения | Участник канала | POST messages {content} | 201, NATS event | `TestSendMessage_MemberCheck` |
| TC-MSG-04 | Реакция toggle | Сообщение существует | POST reactions {emoji: "thumbsup"} x2 | Добавлена, затем удалена | `TestToggleReaction_AddRemove` |
| TC-MSG-05 | Pin сообщения | Участник канала | POST pin | 200, pinned | — (ручной) |
| TC-MSG-06 | Тред | Сообщение существует | POST messages {reply_to: msgId} | 201, привязано к треду | `TestThreadReply_IncrementsReplyCount`, `TestDeleteMessage_ThreadReplyDecrementsCount` |
| TC-MSG-07 | Read-only канал | is_readonly=true, не Owner | POST messages | 403 | `TestSendMessage_ReadOnlyBlocked`, `TestSendMessage_ReadOnlyBlocked_RegularUser`, `TestSendMessage_ReadOnlyAllowed_Admin` |
| TC-MSG-08 | Mute канала | Участник | POST mute {until: "2026-04-01"} | 200, muted_until set | — (ручной) |
| TC-MSG-09 | Unread counts | Новые сообщения | GET unread | 200, count > 0 | `TestIncrement_AllMembers`, `TestIncrement_SkipsSender`, `TestMarkRead_ResetsToZero`, `TestGetAll_ReturnsAllChannels` |
| TC-MSG-10 | Поиск сообщений | Сообщения существуют | GET search?q=keyword | 200, результаты | `TestSearchMessages_TextQuery` |

### TMS — Управление тестированием

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-TMS-01 | Создание тест-кейса (classic) | Участник проекта | POST cases {title, case_type: "classic", steps} | 201, case_key = HUB-TC-1 | `TestCreateTestCase_Classic` |
| TC-TMS-02 | Создание тест-кейса (BDD) | Участник проекта | POST cases {case_type: "bdd", steps: [{keyword: "Given"}]} | 201 | `TestCreateTestCase_BDD` |
| TC-TMS-03 | Версионирование | Кейс существует | PUT cases/:key {title: "new"} -> GET versions | Версия 2 создана | `TestUpdateTestCase_CreatesVersion` |
| TC-TMS-04 | Откат версии | Версия 1 и 2 | POST rollback/1 | 200, данные версии 1 | `TestRollbackTestCase` |
| TC-TMS-05 | Создание прогона | Кейсы существуют | POST runs {title} + POST runs/:id/cases | 201, кейсы добавлены | `TestAddTestCasesToRun` |
| TC-TMS-06 | Выполнение результата | Прогон in_progress | PUT results/:id {status: "passed"} | 200, passed_count++ | `TestExecuteTestResult_UpdatesStats` |
| TC-TMS-07 | Allure import | Allure JSON | POST allure/import {multipart} | 200, прогон создан, matched | `TestParseAllureResults_ValidJSON`, `TestMatchTestCase_ByAnnotation`, `TestMatchTestCase_ByASIDLabel` |
| TC-TMS-08 | JUnit webhook | JUnit XML | POST webhook/junit | 200, результаты импортированы | — (ручной) |
| TC-TMS-09 | Создание бага из failed | Результат failed | POST results/:id/bug | 201, задача Bug создана | `TestCreateBugFromFailedResult` |
| TC-TMS-10 | Flaky report | Тесты с разными результатами | GET reports/flaky | 200, список flaky тестов | `TestGetCoverageReport` |
| TC-TMS-11 | Матрица трассируемости | Кейсы привязаны к задачам | GET reports/traceability | 200, матрица | `TestLinkTestCaseToIssue` |
| TC-TMS-12 | Экспорт PDF | Прогон completed | GET runs/:id/export | 200, PDF файл | `TestExportToCSV` |

### CALL — Созвоны

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-CALL-01 | Создание звонка | Канал существует | POST calls {channelId} | 201, room_name, LiveKit token | `TestGenerateToken` |
| TC-CALL-02 | Instant-звонок | Авторизован | POST calls/instant | 201, invite_code | `TestGenerateToken_DifferentRooms` |
| TC-CALL-03 | Join по invite-коду | Звонок active | GET calls/invite/:code | 200, информация + token | — (ручной) |
| TC-CALL-04 | Guest join (без аккаунта) | Звонок active, invite_code | POST calls/guest-join {invite_code, guest_name} | 200, LiveKit token (гостевой) | — (ручной) |
| TC-CALL-05 | Leave | В звонке | POST calls/:id/leave | 200, participant_count-- | — (ручной) |
| TC-CALL-06 | End call | В звонке | POST calls/:id/end | 200, status=ended | — (ручной) |
| TC-CALL-07 | Start recording | В звонке | POST calls/:id/recording/start | 200, egress started | `TestMaxActiveRecordingsLimit` |
| TC-CALL-08 | Stop recording | Запись идёт | POST calls/:id/recording/stop | 200, status=processing | `TestRecordingStatusValues` |
| TC-CALL-09 | Chat в звонке | В звонке | POST calls/:id/chat {text} | 200, NATS event | `TestSendMessage`, `TestSendMessage_TextTooLong`, `TestWSAuthorizer_CallRoom_CanSubscribe_A2_007`, `TestReadPump_CallAuthorizer_StrangerDenied_A2_007`, `TestReadPump_CallAuthorizer_ParticipantAllowed_A2_007` |
| TC-CALL-10 | Raise hand | В звонке | POST calls/:id/raise-hand | 200, NATS event | — (ручной) |
| TC-CALL-11 | Webhook: participant_joined | LiveKit event | POST calls/webhook | 200, participant_count++ | `TestVerifyWebhookSignature_EmptyAuth`, `TestVerifyWebhookSignature_InvalidJWT` |

### SPRINT — Спринты

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-SPRINT-01 | Создание спринта | Участник проекта | POST sprints {name, start_date, end_date} | 201, status=planning | `TestCreateSprint`, `TestCreateSprint_EndDateBeforeStartDate` |
| TC-SPRINT-02 | Запуск спринта | Спринт planning | POST sprints/:id/start | 200, status=active | `TestStartSprint_Success` |
| TC-SPRINT-03 | Двойной active | Спринт уже active | POST sprints/:id2/start | 400, уже есть active | `TestStartSprint_RejectsWhenActiveExists` |
| TC-SPRINT-04 | Burndown | Спринт active | GET sprints/:id/burndown | 200, точки по дням | `TestGetBurndown` |
| TC-SPRINT-05 | Velocity | 3+ спринтов completed | GET velocity | 200, planned vs completed | `TestGetVelocity` |
| TC-SPRINT-06 | Planning Poker | Спринт planning | POST poker -> vote -> reveal | 200, голоса раскрыты | — (ручной) |
| TC-SPRINT-07 | Завершение спринта | Спринт active | POST sprints/:id/complete | 200, status=completed | `TestCompleteSprint_MovesIncompleteIssues` |

### NOTIF — Уведомления

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-NOTIF-01 | Назначение задачи -> уведомление | Задача существует | PUT issue {assignee: userId} | Уведомление type=assigned | `TestSubjectToType`, `TestNotificationConsumer_NakOnBadJSON_A2_010`, `TestNotificationConsumer_AckOnSuccess_A2_010`, `TestNotificationConsumer_NakOnDispatchError_A2_010` |
| TC-NOTIF-02 | @mention -> уведомление | Комментарий с mention | POST comment {mentions: [userId]} | Уведомление type=mentioned | `TestSubjectToType` |
| TC-NOTIF-03 | Mark as read | Уведомление unread | PATCH notifications/:id/read | 200, is_read=true | — (ручной) |
| TC-NOTIF-04 | Mark all read | 5 unread | PATCH notifications/read-all | 200, все is_read=true | — (ручной) |
| TC-NOTIF-05 | Quiet hours | quiet_hours set, не urgent | Событие в quiet hours | Уведомление отложено | `TestShouldDispatch_Debounce` |
| TC-NOTIF-06 | Urgent пробивает quiet hours | quiet_hours set, is_urgent=true | Событие в quiet hours | Уведомление доставлено | — (ручной) |
| TC-NOTIF-07 | Web Push subscribe | Авторизован | POST push/subscribe {endpoint, keys} | 200 | — (ручной) |

### ADMIN — Администрирование

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-ADMIN-01 | Аудит-лог | Admin | GET audit?action=login | 200, логи входов | `TestAuditLogFiltersDefaults` |
| TC-ADMIN-02 | Деактивация юзера | Admin | POST users/:id/deactivate | 200, status=deactivated | — (ручной) |
| TC-ADMIN-03 | GDPR удаление | Admin | DELETE users/:id/gdpr | 200, все данные удалены | — (ручной) |
| TC-ADMIN-04 | Бэкап | Admin | POST /admin/backups | 200, backup started | `TestCreateZIP` |
| TC-ADMIN-05 | Restore | Admin, backup exists | POST /admin/backups/:id/restore | 200, восстановлено | `TestCreateZIPEmptyData` |
| TC-ADMIN-06 | Restore/dump errors без секретов (A7-007) | Admin; pg_restore/mongorestore fail | Restore / ExportRestore* | ошибка без DSN/CombinedOutput | `TestRestorePostgres_DoesNotEmbedDumpOutput_A7_007`, `TestRestoreMongo_DoesNotEmbedDumpOutput_A7_007`, `TestRestoreHandler_DoesNotLeakSecrets_A7_007` |
| TC-ADMIN-07 | Mail broadcast только active (A7-008) | Admin; pending+deactivated users | POST broadcast dry-run | recipients = active only | `TestAdminService_Broadcast` |
| TC-HR-01 | Leave approve/cancel version≥1 (A6-006) | Member leave pending | approve/cancel version=0 / bad JSON | 400 | `TestHandler_ApproveLeave_VersionZero_A6_006`, `TestHandler_CancelLeave_BindAndVersion_A6_006` |

### Дополнительные модули

| ID | Название | Автотест |
|----|----------|----------|
| TC-RBAC-01 | Проверка наличия permission | `TestPermission_Has` |
| TC-RBAC-02 | Добавление permission | `TestPermission_Add` |
| TC-RBAC-03 | Удаление permission | `TestPermission_Remove` |
| TC-RBAC-04 | Idempotent операции | `TestPermission_AddRemoveIdempotent` |
| TC-RBAC-05 | Дефолтные роли | `TestDefaultRolePermissions` |
| TC-GITLAB-01 | Issue key regex | `TestIssueKeyRegex` |
| TC-GITLAB-02 | Slugify | `TestSlugify` |
| TC-GITLAB-03 | Truncate | `TestTruncate` |
| TC-GITHUB-01 | Issue key regex | `TestIssueKeyRegex` |
| TC-GITHUB-02 | Webhook signature | `TestValidateWebhookSignature` |
| TC-GITHUB-03 | Slugify | `TestSlugify` |
| TC-GITHUB-04 | Truncate | `TestTruncate` |
| TC-IMEX-01 | Field mapping | `TestBuildFieldMap` |
| TC-IMEX-02 | Get field | `TestGetField` |
| TC-IMEX-03 | Bounds check | `TestGetField_OutOfBounds` |
| TC-STORAGE-01 | Upload URL TTL | `TestUploadURLTTL` |
| TC-STORAGE-02 | Download URL TTL | `TestDownloadURLTTL` |
| TC-STORAGE-03 | Result fields | `TestUploadResult_Fields` |
| TC-STORAGE-04 | ConfirmUpload already confirmed → 409 (A4-006) | `TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006`, `TestConfirmUpload_RejectsAlreadyConfirmed_A4_006` |
| TC-WS-01 | Message type constants | `TestMessageTypeConstants` |
| TC-WS-02 | Client/Server JSON | `TestClientMessage_JSON`, `TestServerMessage_JSON` |
| TC-WS-03 | Ping protocol | `TestClientMessage_PingHasNoRoom` |
| TC-WS-04 | NewClient | `TestNewClient` |
| TC-APIKEY-01 | Hash key | `TestHashKey`, `TestHashKeyDeterministic`, `TestHashKeyUniqueness` |
| TC-APIKEY-02 | Key generation | `TestKeyGeneration` |
| TC-APIKEY-03 | Event filtering | `TestContainsEvent`, `TestContainsEventWildcard` |
| TC-APIKEY-04 | Webhook signing | `TestSignPayload` |
| TC-NATS-01 | Stream config | `TestStreamNameConstant`, `TestStreamSubjects`, `TestClientStructFields` |
| TC-PLATFORM-01 | Fiber recover EnableStackTrace == DevMode (A14-007) | `TestFiberRecoverEnableStackTrace_FollowsDevMode_A14_007` |
| TC-SEED-01 | Demo data validity | `TestDemoUsersValid`, `TestDemoStatusesValid`, `TestDemoIssuesValid`, `TestDemoTestCasesValid`, `TestDemoChannelsValid` |
| TC-DASH-01 | Layout serialization | `TestLayoutItemSerialization`, `TestLayoutDTOEmptyItems` |
| TC-DASH-02 | Data defaults | `TestPersonalDataDefaults`, `TestProjectStatsDefaults` |
| TC-DASH-03 | Admin stats | `TestAdminStatsServiceHealth` |
| TC-DASH-04 | Save validation | `TestSaveLayoutRequestValidation` |
| TC-SEARCH-01 | Merge results | `TestMergeResultsByScore`, `TestMergeResultsWithLimit` |
| TC-SEARCH-02 | UUID array | `TestUUIDSliceToPgArray`, `TestUUIDSliceToPgArrayEmpty` |
| TC-SEARCH-03 | Saved filters | `TestSavedFilterToDTO` |
| TC-SEARCH-04 | Defaults | `TestSearchRequestDefaultLimit`, `TestSearchResultTypes` |
| TC-SEARCH-05 | Soft-delete exclusion Ctrl+K (A4-002) | `TestPostgresSearch_ExcludesSoftDeleted_A4_002`, `TestSearchSQL_ExcludesSoftDeleted_A4_002` |
| TC-JQL-01 | JQL parsing | `TestParseJQLSimple` .. `TestParseJQLOR` (13 тестов) |
| TC-JQL-02 | JQL to Ent predicates | `TestToEntPredicates`, `TestToEntPredicatesIN` |
| TC-ALLURE-01 | Parse valid JSON | `TestParseAllureResults_ValidJSON` |
| TC-ALLURE-02 | Missing/empty/malformed | `TestParseAllureResults_MissingFields`, `TestParseAllureResults_EmptyInput`, `TestParseAllureResults_MalformedJSON` |
| TC-ALLURE-03 | Nil pointers | `TestParseAllureResults_NilPointers` |
| TC-ALLURE-04 | Status mapping | `TestAllureStatusToTMS`, `TestExtractLabelValue` |
| TC-ALLURE-05 | Matching strategies | `TestMatchTestCase_ByAnnotation` .. `TestMatchTestCase_PriorityOrder` (7 тестов) |
| TC-PRESENCE-01 | Heartbeat | `TestHeartbeat_SetsScore` |
| TC-PRESENCE-02 | Status detection | `TestGetStatus_Online`, `TestGetStatus_Away`, `TestGetStatus_Offline` |
| TC-PRESENCE-03 | Cleanup | `TestCleanupStale` |
| TC-PRESENCE-04 | Custom status | `TestSetGetCustomStatus` |
| TC-PRESENCE-05 | Bulk status | `TestGetBulkStatus` |

---

## Статистика покрытия

### Автоматизация тест-кейсов

| Модуль | TC всего | Автоматизировано | Только ручные | % автоматизации |
|--------|----------|-----------------|---------------|-----------------|
| AUTH | 11 | 7 | 4 | 64% |
| PROJ | 10 | 6 | 4 | 60% |
| TRCK | 12 | 7 | 5 | 58% |
| MSG | 10 | 8 | 2 | 80% |
| TMS | 12 | 10 | 2 | 83% |
| CALL | 11 | 6 | 5 | 55% |
| SPRINT | 7 | 6 | 1 | 86% |
| NOTIF | 7 | 3 | 4 | 43% |
| ADMIN | 5 | 2 | 3 | 40% |
| Доп. модули | 36 | 36 | 0 | 100% |
| **Итого** | **121** | **91** | **30** | **75%** |

---

## Regression-тесты

### Критический путь (после каждого деплоя / CI e2e-full)

1. **Авторизация**: Magic Link или dev-login → GET /auth/me
2. **Проект**: Создать проект → workflow + канал
3. **Задача**: CRUD issue → смена статуса → комментарий
4. **Мессенджер**: Отправить сообщение → реакция → тред
5. **TMS**: Кейс → прогон → результат passed
6. **Ctrl+K**: Глобальный поиск
7. **Logout**: POST /auth/logout → redirect login
8. **a11y**: axe на login + authenticated routes

### WebSocket Regression

1. Подключение /ws через ticket subprotocol (не ?token=JWT)
2. Subscribe to project:{id}
3. Создать задачу через API
4. Получить event issue_created через WS
5. Unsubscribe

---

## Performance-тесты

### Целевые метрики

| Метрика | Целевое значение |
|---------|-----------------|
| API response time (p95) | < 200ms |
| WebSocket message latency | < 100ms |
| Kanban board load (100 задач) | < 500ms |
| Messenger: загрузка канала (1000 сообщений) | < 300ms |
| Глобальный поиск | < 500ms |
| Allure import (1000 результатов) | < 5s |
| Одновременных WS-подключений | 100+ |
| Участников в звонке | до 50 |

### Rate Limiting

| Тест | Ожидание |
|------|----------|
| 4 запроса auth за 10 мин | 4-й -> 429 |
| 101 API запрос за 1 сек | 101-й -> 429 |
| 11 file uploads за 1 мин | 11-й -> 429 |
| 61 сообщение за 1 мин | 61-й -> 429 (лимит **60/мин**, `MessageSendRateLimit`) |

---

## Ручное тестирование — чеклист по фичам

### Авторизация
- [ ] Magic Link: email → код → вход
- [ ] Dev-login (только DEV_MODE)
- [ ] Профиль: имя, timezone, язык
- [ ] Сессии: список, отзыв
- [ ] Logout: POST /auth/logout
- [ ] Онбординг: 4 шага (профиль → тур → тема → проект)
- [ ] Telegram: **только** привязка уведомлений (не login)

### Проекты
- [ ] Создание проекта с ключом
- [ ] Invite-ссылка: генерация, присоединение
- [ ] Смена ролей участников
- [ ] Архивирование и восстановление
- [ ] Удаление (корзина) и восстановление

### Трекер задач
- [ ] Kanban-доска: drag & drop задач между колонками
- [ ] Создание задачи всех типов
- [ ] Комментарии с @mention
- [ ] Time tracking: таймер start/stop
- [ ] JQL поиск
- [ ] Workflow: настройка статусов и переходов

### Мессенджер
- [ ] Публичные/приватные каналы
- [ ] DM (личные сообщения)
- [ ] Отправка сообщений с форматированием
- [ ] Реакции, pin, forward
- [ ] Треды (ответы)
- [ ] Поллы
- [ ] Typing indicator
- [ ] Unread counts

### TMS
- [ ] Создание тест-кейса (classic + BDD)
- [ ] Прогон: создание, добавление кейсов, выполнение
- [ ] Allure import через API
- [ ] Создание бага из failed результата
- [ ] Отчёты: flaky, trends, traceability

### Созвоны
- [ ] Создание звонка из канала
- [ ] Instant-звонок с invite-кодом
- [ ] Гостевой вход по invite-коду (без аккаунта)
- [ ] Screen share
- [ ] Запись звонка (start/stop)
- [ ] Эфемерный чат звонка
- [ ] Raise hand

### Дашборды
- [ ] Личный дашборд: виджеты с данными
- [ ] Проектный дашборд: burndown, velocity
- [ ] Админский дашборд: статистика пользователей
- [ ] Drag & drop виджетов, изменение размеров

### Уведомления
- [ ] In-app уведомления (колокольчик)
- [ ] Telegram бот уведомления
- [ ] Web Push подписка
- [ ] Quiet hours
- [ ] Настройка по типам

---

## Accessibility-тесты

### Стандарт

WCAG 2.1 AA (целевой для v2.0 Redesign)

### Проверки

| Проверка | Метод | Статус |
|----------|-------|--------|
| Контраст текста (4.5:1) | Automated (axe-core) | Проверяется при UI review |
| Контраст UI элементов (3:1) | Automated (axe-core) | Проверяется при UI review |
| Keyboard navigation | Manual | Tab, Enter, Escape на всех interactive элементах |
| Screen reader labels | Manual | aria-label на кнопках без текста |
| Focus indicators | Visual | Все focusable элементы имеют видимый focus ring |
| Color не единственный индикатор | Visual | Иконки + текст в дополнение к цвету |

---

*Последнее обновление: 2026-07-29 — inventory **805** BE `*_test.go` / **39** export_hooks / Vitest **167** files; COV 83% остаётся baseline прогона 2026-06-22*
