# Hub — Testing Guide (канон для авторов автотестов)

> **Назначение:** практический executable-гайд «как писать автотесты в Hub».  
> **Не дублирует** эталонный playbook целиком — сжимает канон и даёт deep-links.  
> **Проза:** русский · **код / идентификаторы:** английский.

| Документ | Роль |
|----------|------|
| **Этот гайд** | Правила написания, паттерны, DoD перед `qa-auto-ver` |
| [`.cursor/playbooks/testing.md`](../.cursor/playbooks/testing.md) | Полный эталон (техники, Allure, triage, cookbook доменов) |
| [`docs/testframework_strategy.md`](../docs/testframework_strategy.md) | Стратегия: детерминизм, пирамида, TF-PY roadmap |
| [`PIPELINE.md`](PIPELINE.md) | Карта CI: стадия → job → путь |
| [`docs/TESTING.md`](../docs/TESTING.md) | TC-матрица + колонка «Автотест» |
| [`docs/test_plan.md`](../docs/test_plan.md) | Тест-план / регрессия |

**Deep-guides (детали по слоям):**

| # | Файл |
|---|------|
| 01 | [guides/01-principles.md](guides/01-principles.md) — SOLID / DRY / KISS / YAGNI / OOP |
| 02 | [guides/02-page-object.md](guides/02-page-object.md) — POM канон Hub |
| 03 | [guides/03-backend-go.md](guides/03-backend-go.md) — `backend/testframework` |
| 04 | [guides/04-frontend-vitest.md](guides/04-frontend-vitest.md) — Vitest + RTL + MSW |
| 05 | [guides/05-e2e-playwright.md](guides/05-e2e-playwright.md) — Playwright E2E |
| 06 | [guides/06-edge-corner-coverage.md](guides/06-edge-corner-coverage.md) — happy / edge / corner → тесты |

---

## Оглавление

1. [Аудитория и конвейер](#1-аудитория-и-конвейер)
2. [Пирамида тестов Hub](#2-пирамида-тестов-hub)
3. [Принципы и антипаттерны](#3-принципы-и-антипаттерны)
4. [OOP в тестах](#4-oop-в-тестах)
5. [Именование и Allure](#5-именование-и-allure)
6. [Детерминизм](#6-детерминизм)
7. [Трассируемость TC ↔ автотест](#7-трассируемость-tc--автотест)
8. [Happy / negative / edge / corner](#8-happy--negative--edge--corner)
9. [DoD перед qa-auto-ver](#9-dod-перед-qa-auto-ver)
10. [Команды прогона](#10-команды-прогона)

---

## 1. Аудитория и конвейер

| Кто | Пишет | Путь |
|-----|-------|------|
| `be1` / `be2` | **Unit** Go рядом с фичей (обязанность разработчика) | логика в `internal/`; тесты — `backend/testframework/{domain}/` |
| `fe1` / `fe2` | **Unit** TS/Vitest рядом с фичей | `stages/unit/frontend-test/domains/{domain}/` |
| `qa-auto1` | Integration/API Go + E2E (auth, navigation, tracker, tms, messenger) | `backend/testframework/`, `stages/e2e/domains/` |
| `qa-auto2` | Vitest-компонентные + E2E (admin, ai, platform, v4) | `stages/unit/frontend-test/`, `stages/e2e/domains/` |
| `qa-manual` | TC в `docs/TESTING.md` | не пишет автотесты |
| `qa-auto-ver` | Верификация автотестов | не пишет тесты (кроме контр-примеров) |

**Канонический конвейер** (юниты — вне него, пишут `be`/`fe`):

```
BA/SA (US/AC/FR) → qa-manual (TC) → qa-manual-ver (verified)
  → qa-auto1/2 (автотест + колонка «Автотест») → qa-auto-ver
```

Автотесты **только** по TC, прошедшим `qa-manual-ver`. Подробно: playbook §1.

**Calls / LiveKit:** SFU-миграция отклонена — тесты звонков остаются на **LiveKit** (не mediasoup).

---

## 2. Пирамида тестов Hub

```
        ┌─────────────┐
        │     E2E     │  Playwright + POM — критические user paths
        │  ~20 spec   │  e2e-smoke (3) / e2e-full
        ├─────────────┤
        │ Integration │  Go handler/API, Vitest-компонентные — по TC
        ├─────────────┤
        │    Unit     │  be/fe вместе с кодом (mockRepo / MSW / store)
        └─────────────┘
```

| Слой | Стек | Где | Кто |
|------|------|-----|-----|
| Unit BE | Go + testify, `//go:build hubtest` | `backend/testframework/{domain}/` | `be*` (+ `qa-auto1` integration) |
| Unit FE | Vitest + RTL + MSW | `stages/unit/frontend-test/domains/` | `fe*` (+ `qa-auto2`) |
| E2E | Playwright TS + POM | `stages/e2e/domains/` | `qa-auto1` / `qa-auto2` по доменам |
| Coverage | `go tool cover` | gate ≥82%, security ≥80% | `coverage-gate` |

Go физически в `backend/testframework/` (запрет импорта `internal/` из другого модуля). Логически — job `backend-test`. См. [`PIPELINE.md`](PIPELINE.md).

**Антипаттерн «мороженое»:** вся логика в E2E → медленно, флейки, жжёт CI (~3000 мин/мес). Логику — в unit; E2E — критический путь.

---

## 3. Принципы и антипаттерны

Кратко; развёрнуто — [guides/01-principles.md](guides/01-principles.md).

| Принцип | В тестах |
|---------|----------|
| **KISS** | Один сценарий = один тест; без «универсальных» хелперов на всё |
| **DRY** | Общее — в `env.Builder`, `factory`, POM, `renderWithProviders`; не копипастить setup |
| **YAGNI** | Не писать E2E «на будущее»; не тащить testcontainers в каждый unit |
| **SOLID** | SRP у Page/Steps/Factory; DIP через `Repository` + `mockRepo`; не god-fixture |

**Топ антипаттернов (полный список — playbook §16):**

| Bad | Good |
|-----|------|
| `time.Sleep` / `waitForTimeout` | web-first `expect`, fake clock |
| Мок `return nil, nil` | эмуляция контракта (`NotFound`, `Conflict`) |
| Общий мутабельный fixture | свой sqlite / QueryClient / сброс store на тест |
| CSS `div > .x:nth-child` | `getByRole` / `getByLabel` |
| Автотест без TC-ID | комментарий `// TC-…` + колонка в `TESTING.md` |
| Coverage-профиль в `backend/` | профиль в TMP + cleanup (rule 45) |

---

## 4. OOP в тестах

Паттерны Hub — не теория, а код в репо.

### 4.1. Page Object Model + Steps (E2E)

```
BasePage → LoginPage / AppShellPage → AuthSteps → login.spec.ts
```

| Слой | Путь | Ответственность |
|------|------|-----------------|
| BasePage | `stages/e2e/shared/pages/base.page.ts` | `goto`, общий locator |
| Page Object | `…/pages/login.page.ts`, `app-shell.page.ts` | локаторы + действия **одной** страницы |
| Steps | `…/steps/auth.steps.ts` | бизнес-сценарий (login as admin) |
| Spec | `…/domains/auth/login.spec.ts` | сценарий + ассерты, без сырых селекторов |
| Fixtures | `…/shared/fixtures.ts` | `skipUnlessBackend`, `skipUnlessDevMode` |

Подробно: [guides/02-page-object.md](guides/02-page-object.md), [guides/05-e2e-playwright.md](guides/05-e2e-playwright.md).

### 4.2. Builder + Factory (backend)

```go
// backend/testframework/env/builder.go — fluent Builder, уникальный in-memory DSN
e := env.New(t, "tracker").Build()

// backend/testframework/factory/user.go — Factory, уникальный email
uf := factory.NewUserFactory(t, e.Client, e.Ctx)
alice := uf.Create("alice")
```

Реальный consumer: `backend/testframework/tracker/repository_test.go` (`newTrackerRepoTestEnv`).

### 4.3. mockRepo (unit service)

Ручной мок интерфейса `Repository` (map in-memory) — изоляция сервиса от БД. Паттерн: playbook §4.7; примеры в `backend/testframework/ai/`, `project/`.

### 4.4. TestRenderer (Vitest)

```typescript
// stages/unit/frontend-test/render.tsx
renderWithProviders(<Component />); // свой QueryClient на тест, retry: false
```

Подробно: [guides/03-backend-go.md](guides/03-backend-go.md), [guides/04-frontend-vitest.md](guides/04-frontend-vitest.md).

---

## 5. Именование и Allure

### 5.1. Файлы

| Слой | Паттерн | Пример |
|------|---------|--------|
| Go | `{layer}_{feature}_test.go`, `service_wave<NN>_test.go` | `handler_magiclink_runes_test.go`, `service_wave126_test.go` |
| Vitest | `{subject}.test.ts(x)` | `auth-store.test.ts`, `use-auth.test.ts` |
| E2E | `{flow}.spec.ts` в `domains/{domain}/` | `login.spec.ts`, `project-lifecycle.spec.ts` |
| POM | `{page}.page.ts` | `login.page.ts` |
| Steps | `{domain}.steps.ts` | `auth.steps.ts` |

Пакет Go: `package {domain}_test` (black-box) + `export_hooks.go` с `//go:build hubtest` при white-box.

### 5.2. Имена тестов

| Слой | Формат | Пример |
|------|--------|--------|
| Go | `Test{Subject}_{Scenario}` | `TestVerifyTelegramAuth_InvalidHash` |
| Vitest | `it("…behavior…")` | `it("logout clears tokens and isAuthenticated")` |
| E2E | `test("…user-visible outcome…")` | `test("dev login panel authenticates and redirects to projects")` |

### 5.3. TC-ID

Формат: `TC-<MODULE>-<NN>` где MODULE = `AUTH` / `TRCK` / `MSG` / `TMS` / `CALL` / …  
В Allure/TMS: `HUB-TC-<MODULE>-<NN>` (label `AS_ID` или tms-link).

### 5.4. Allure labels (минимум)

| Label | Значение |
|-------|----------|
| `AS_ID` / tms-link | `HUB-TC-TRCK-03` — **обязателен** для матчинга в Hub TMS |
| `epic` / `feature` / `story` | иерархия отчёта |
| `severity` | `blocker` / `critical` / `normal` / `minor` |

Go: минимум комментарий `// TC-…` над функцией; JSON Allure — playbook §8.5.  
TS: `allure-js-commons` — playbook §8.4.  
**Не прикреплять** JWT / PII в attachments.

---

## 6. Детерминизм

| Правило | Как |
|---------|-----|
| **No sleep** | Playwright auto-wait / `expect(…).toBeVisible()`; Go — каналы, fake clock |
| **Isolation** | Go: DSN с `uuid.New()`; Vitest: `beforeEach` сброс store + свой `QueryClient`; E2E: уникальный суффикс данных |
| **Race** | `go test -race -count=1` обязателен |
| **No real network in unit** | `mockRepo` / MSW / `vi.spyOn`; MSW `onUnhandledRequest: "error"` |
| **Parallel** | `t.Parallel` только при полной изоляции; E2E workers — без shared project/channel |
| **Skip** | только с причиной (`skipUnlessBackend`); не молчаливый `t.Skip` |

Стратегия: [`docs/testframework_strategy.md`](../docs/testframework_strategy.md) §1.1. Флейки: playbook §13.

---

## 7. Трассируемость TC ↔ автотест

```
US/AC / FR  →  TC-<MODULE>-NN (docs/TESTING.md)  →  автотест  →  Allure/TMS
```

**Двусторонняя связь обязательна:**

1. **Код → MD:** комментарий / Allure `AS_ID` с TC-ID в тесте.
2. **MD → код:** точное имя функции / spec в колонке «Автотест».

```go
// TC-TRCK-03, TC-TRCK-04: optimistic lock — успех и конфликт версий.
func TestUpdateIssueOptimisticLocking(t *testing.T) { … }
```

Автотест без TC и без строки в `TESTING.md` — **reject** от `qa-auto-ver`.  
Подробно: playbook §7.

---

## 8. Happy / negative / edge / corner

| Тип | Что проверяем | Где в первую очередь |
|-----|---------------|----------------------|
| **Happy** | Основной успешный путь | unit + 1 smoke E2E |
| **Negative** | 400/401/403/404/409/429, доменные ошибки | unit/integration (не дублировать всё в E2E) |
| **Edge** | Границы (лимиты, пустые коллекции, max length) | unit table-driven |
| **Corner** | Редкие комбинации, гонки, double-submit | targeted unit + точечный E2E при user-visible риске |

Каталог кейсов (источник backlog): `.planning/EDGE-CORNER-CATALOG.md` (если есть).  
Как превращать кейсы в тесты: [guides/06-edge-corner-coverage.md](guides/06-edge-corner-coverage.md).

**Правило слоя:** чем экзотичнее кейс — тем ниже в пирамиде (дешевле и стабильнее).

---

## 9. DoD перед qa-auto-ver

Чек-лист автора автотеста (сжатый; полный — playbook §15.2 / §17):

- [ ] TC верифицирован `qa-manual-ver`; слой/домен не конфликтует с другим `qa-auto`
- [ ] Паттерн Hub: `mockRepo` / `env.Builder` / factory / POM+Steps / `renderWithProviders` / MSW
- [ ] Happy + negative + границы (где применимо); ассерты на **поведение**, не на мок
- [ ] Детерминизм: нет sleep / реальной сети / shared state; `-race` зелёный
- [ ] TC-ID в коде + имя в колонке «Автотест»; Allure `AS_ID` при заливке в TMS
- [ ] Прогон зелёный (команды §10); wall-time известен; артефакты удалены (rule 45)
- [ ] Docs: `docs/TESTING.md` (+ `COVERAGE.md` при coverage-задаче)
- [ ] При FAIL — triage (`ROOT=code|test`), не «зеленить вслепую» (skill `test-triage`)

`verified: true` — только после green regression + лога.

---

## 10. Команды прогона

Источник: [`PIPELINE.md`](PIPELINE.md), [`AGENTS.md`](../AGENTS.md).

```bash
# Backend (job: backend-test)
cd backend && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...
cd backend && go test -tags hubtest -race -count=1 ./testframework/auth/...

# Frontend unit (job: frontend-test)
cd frontend && node scripts/link-testframework-node-modules.mjs
cd frontend && npx tsc --noEmit && npm run test:run
cd frontend && npx vitest run ../testframework/stages/unit/frontend-test/domains/auth

# E2E
cd frontend && npm run test:e2e:smoke   # e2e-smoke
cd frontend && npm run test:e2e         # e2e-full

# Canonical Docker backend (как в CI)
docker compose -f deployments/docker-compose.test.yml up -d
# затем go test -tags hubtest … в контейнере golang (см. AGENTS.md)
```

Полный gate: `/qa-gate` (останавливается на первой FAIL).  
Coverage-профили — **только в TMP**, не в `backend/` (rule 45).

---

## Быстрая навигация по playbook

| Тема | Playbook |
|------|----------|
| Роли / разделение qa-auto1/2 | §1 |
| Техники тест-дизайна | §3 |
| Backend Go | §4 |
| Frontend Vitest | §5 |
| E2E POM | §6 |
| Traceability | §7 |
| Allure / TMS | §8 |
| Замер времени | §9 |
| Test triage | §12 |
| Флейки | §13 |
| Coverage | §14 |
| DoD / чек-лист ver | §15–§17 |
| Cookbook доменов | Приложение A |

---

*Канон testframework · 2026-07 · владельцы: qa-auto1 / qa-auto2 · верификация: qa-auto-ver*
