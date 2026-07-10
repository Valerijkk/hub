# 01 — Принципы тестов Hub (SOLID / DRY / KISS / YAGNI / OOP)

> Часть [Testing Guide](../TESTING_GUIDE.md). Полный эталон: [`.cursor/playbooks/testing.md`](../../.cursor/playbooks/testing.md) §2, §16.

---

## KISS

Один тест = один сценарий. Имя отражает поведение. Не смешивать happy и negative в одном `t.Run` без нужды.

```go
// Good
func TestVerifyTelegramAuth_InvalidHash(t *testing.T) { … }

// Bad — «всё про Telegram» в одном тесте на 200 строк
func TestTelegram(t *testing.T) { /* valid + invalid + expired + missing */ }
```

## DRY

Общий setup выносится в:

| Слой | Переиспользование |
|------|-------------------|
| Go | `env.Builder`, `factory.UserFactory`, `newTestService()` / `newTrackerRepoTestEnv` |
| Vitest | `renderWithProviders`, MSW handlers, `resetAuthStore()` |
| E2E | POM + Steps + `shared/fixtures.ts` |

```go
// Good — backend/testframework/tracker/repository_test.go
func newTrackerRepoTestEnv(t *testing.T) trackerTestEnv {
	t.Helper()
	e := env.New(t, "tracker").Build()
	return trackerTestEnv{client: e.Client, repo: tracker.NewEntRepository(e.Client, nil), …}
}
```

Не DRY ради DRY: общий хелпер, который знает про 5 доменов — это god-helper (нарушение SRP).

## YAGNI

- Не писать E2E на каждый negative HTTP-код — это unit/handler.
- Не поднимать testcontainers Postgres в каждом unit — сначала `mockRepo`, потом `env.Builder` (sqlite).
- Не добавлять абстракции «на будущее» (generic BaseTestSuite без потребителей).

## SOLID в тестах

| Буква | Применение |
|-------|------------|
| **S** | `LoginPage` — только `/login`; `AuthSteps` — только вход; Factory — только создание сущности |
| **O** | Новый Page Object расширяет `BasePage`, не правит чужие локаторы |
| **L** | `mockRepo` соблюдает контракт `Repository` (те же ошибки, что Ent) |
| **I** | Не тащить fat-интерфейс в тест — partial mock через embedded interface (playbook §4.7) |
| **D** | Service зависит от `Repository` interface → подставляем `mockRepo` |

## OOP-паттерны (карта)

```
                    ┌─────────────┐
                    │   Spec/Test │  сценарий + assert
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
      ┌────────┐     ┌─────────┐     ┌──────────┐
      │ Steps  │     │ Factory │     │ Fixtures │
      └───┬────┘     └────┬────┘     └──────────┘
          ▼               ▼
     ┌─────────┐    ┌──────────┐
     │   POM   │    │  Builder │
     └────┬────┘    └──────────┘
          ▼
     ┌─────────┐
     │ BasePage│
     └─────────┘
```

| Паттерн | Где в репо |
|---------|------------|
| Page Object | `testframework/stages/e2e/shared/pages/` |
| Steps | `…/shared/steps/` |
| Builder | `backend/testframework/env/builder.go` |
| Factory | `backend/testframework/factory/user.go` |
| Fixture | `…/e2e/shared/fixtures.ts`; Go `t.Cleanup` |
| Test double | `mockRepo` в доменных `*_test.go` |

## Антипаттерны (сжато)

| Bad | Почему | Good |
|-----|--------|------|
| God Page Object на всё приложение | хрупкость, merge-конфликты | одна страница = один класс |
| Spec с `page.locator(".btn")` | дублирование, ломка при UI | Steps → POM |
| Shared package-level `var client` | race / порядок тестов | создать в тесте / `Builder.Build()` |
| Assert на private поля | ломается при рефакторинге | публичный контракт / HTTP / UI |

Полная таблица: playbook §16.
