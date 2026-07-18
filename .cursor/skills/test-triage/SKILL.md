---
name: test-triage
description: При падении теста или баге — сначала проверить корректность функционала, потом решать чинить код или тест. Обязателен для QA, BE, FE, debugger.
---

# Test Triage — Code vs Test

> **Принцип Hub:** тест — это гипотеза о поведении. Падающий тест не означает автоматически баг в продакшн-коде.

## Когда применять

- `go test` / `vitest` / Playwright FAIL
- Регрессия после рефакторинга
- Агент «чинит» тесты чтобы стали зелёными
- Пользователь сообщает о баге, а тесты зелёные

## Порядок (строго)

### 1. Зафиксировать симптом

```
Тест:     TestBanUser_EmptyReason
Пакет:    internal/audit
Ошибка:   expected error, got nil
```

### 2. Воспроизвести ФУНКЦИОНАЛ (не тест)

- Прочитать требование в `docs/system_specification.md` / User Story / PLAN.md
- Трассировать production path: handler → service → repository
- Ручная проверка или минимальный repro (curl, integration fixture)
- Ответить: **«Как ДОЛЖНО работать по спецификации?»**

### 3. Классификация

| Ситуация | Действие |
|----------|----------|
| Код нарушает спецификацию | **Fix code** → обновить тесты если контракт уточнился |
| Код соответствует спецификации, тест устарел/неверен | **Fix test** (mock, assertion, fixture, flaky timing) |
| Спецификация неясна | **Stop** → уточнить у SA/BA, обновить docs, потом код/тест |
| Flaky (race, timing, network) | Стабилизировать тест (deterministic mock, `-race`, fake clock) |

### 4. Запрещено

- Менять assertion чтобы «прошло», не проверив шаг 2
- Удалять тест без обоснования в skill-логе
- `t.Skip()` без ticket/плана
- Чинить продакшн-код только потому что «тест так ожидает», если тест противоречит docs

### 5. После фикса

```bash
# Минимум затронутый пакет
go test -race -count=1 ./internal/<pkg>/...

# Перед handoff
/qa-gate  # или полный gate по REGULATION §8
```

Записать в skill-лог:

```
TRIAGE: TestX — ROOT=test (устаревший mock) | ROOT=code (BanUser не проверял reason)
FIX: <файл:строка>
VERIFIED: manual + go test green
```

## GSD связь

- Сложный баг → `gsd-debug` (научный метод, `.planning/debug/`)
- Массовые падения после аудита → `gsd-audit-fix --dry-run` сначала
- Новые тесты после фазы → `gsd-add-tests`

## Примеры Hub

### Пример A — чинить ТЕСТ

`TestRefreshRedisKey_UsesHashNotPlaintext` падает после NF-12: тест ожидал plaintext key в Redis, но код правильно хеширует → обновить ожидание в тесте.

### Пример B — чинить КОД

`TestBanUser_EmptyReason` падает с `got nil`: спецификация требует reason → добавить validation в `BanUser`, тест оставить.

### Пример C — чинить ОБА

GDPR delete: тест ожидал 200 при partial failure, спецификация требует error → код `errors.Join` + тест обновить на `want err`.

---

## Дополнительные сценарии по доменам (`backend/testframework/{domain}/`)

> Полный набор с таблицами и обоснованием: `.cursor/REGULATION.md` §14, §18. Здесь — быстрый справочник по каждому домену для повторного использования.

### `auth` (`backend/internal/auth/`: `magiclink.go`, `service_login.go`, `service_session.go`, `middleware.go`, `telegram.go`)

| Тест | Вердикт | Почему |
|------|---------|--------|
| `TestMagicLink_RateLimit_Exceeded` | ROOT=code | `tz_constraints.md`: 3 попытки/10 мин — не реализовано |
| `TestRefreshToken_TTL` (ожидал 7д) | ROOT=test | Требование поменялось на 30д, тест устарел |
| `TestTelegramBind_AsLogin` (зелёный, но не должен быть) | ROOT=code | ADR-AUTH-01: Telegram — bind/notify только, тест слишком слабый |
| `TestSessionMiddleware_ExpiredJWT` (flaky) | Flaky | Заменить `time.Now()` на fake clock |

### `tracker` (`backend/internal/tracker/`: `handler_issues.go`, `automation_engine.go`, `repo_time.go`, `service_workflow.go`)

| Тест | Вердикт | Почему |
|------|---------|--------|
| `TestAutomationEngine_TriggerOnStatusChange` | ROOT=code | Consumer не подписан на `hub.tracker.issue.updated` |
| `TestIssueHandler_Pagination` (ждёт cursor) | ROOT=test | Задачи — offset/limit по TZ, не cursor |
| `TestRepoTime_Overlap` | Stop → clarify | Спецификация не покрывает overlap явно — эскалация `sa` |
| `TestWorkflowTransition_SkipGuard` | ROOT=test | Устаревший mock после рефакторинга интерфейса |

### `bases` (no-code БД, `backend/internal/bases/`: `service.go`, `repository.go`)

| Тест | Вердикт | Почему |
|------|---------|--------|
| `TestBasesRBAC_ViewerCannotEdit` | ROOT=code | Нет role-guard перед `UpdateRow` |
| `TestBasesService_SoftDelete` | ROOT=code | Использует hard `Delete()` вместо `SetDeletedAt()` — критично, TZ mandatory |
| `TestBasesService_OptimisticLock` | ROOT=code | Нет проверки `version` перед update |
| `TestBasesService_ColumnTypeCast` | Stop → clarify | Неясно ожидаемое поведение при invalid type cast |

### `messenger` (`backend/internal/messenger/`)

| Тест | Вердикт | Почему |
|------|---------|--------|
| `TestChannelPagination_Cursor` (ждёт offset/limit) | ROOT=test | Чат — cursor по TZ, тест скопирован из tracker-паттерна |
| `TestPresence_WSRoomJoin_NoMembershipCheck` (зелёный, не должен быть) | ROOT=code | NF-02: WS room без RBAC — forbidden, дыра скрыта слабым тестом |

### `tms`

| Тест | Вердикт | Почему |
|------|---------|--------|
| `TestTestRailImport_DuplicateCase` | ROOT=code | Импорт должен делать upsert по `external_id`, создаёт дубликат |
| `TestRunsExtra_StatusTransition` | ROOT=test | Сигнатура сервиса изменилась, поведение — нет; обновить mock |

### `calls` (LiveKit)

| Тест | Вердикт | Почему |
|------|---------|--------|
| `TestChatMongo_MessagePersist` (после reconnect) | ROOT=code | Consumer не переподписывается на JetStream после reconnect |
| `TestSearch_Wave103Mongo_NoUserScope` (зелёный, не должен быть) | ROOT=code | NF-03: search без user-scope — forbidden |

## Правило для «Stop → clarify»

Если ни docs, ни ADR, ни TZ-constraints не дают однозначного ответа — **не гадать**. Записать вопрос в `.planning/STATE.md` §Blockers, эскалировать `sa1`/`sa2` или `product`, дождаться уточнения (обновление `docs/system_specification.md`), и только потом классифицировать code/test.

---

*Hub mandatory for: qa, be, fe, debugger, *-ver*
