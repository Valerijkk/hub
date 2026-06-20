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

*Hub mandatory for: qa, be, fe, debugger, *-ver*
