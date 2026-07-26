# Hub — Эталонный Playbook по тестированию и Allure

> **Статус:** эталонный документ QA-практик Hub. На него ссылаются агенты `qa-manual`, `qa-auto1`, `qa-auto2`, `qa-auto-ver`, `qa-manual-ver`, а также `be`/`fe`.
> **Источники истины:** `docs/TESTING.md` (TC-матрица), `docs/test_plan.md` (стратегия), `testframework/PIPELINE.md` (карта CI), `.cursor/rules/40-testing.mdc`, `.cursor/rules/45-artifacts-cleanup.mdc`, `.cursor/rules/56-regression-verify.mdc`, `.cursor/skills/test-triage/SKILL.md`, `.cursor/skills/qa-gate/SKILL.md`, `.cursor/skills/coverage-audit/SKILL.md`.
> **Правило:** проза — русская, код — английский (Go/TS как в репозитории). Комментарии в коде объясняют ПОЧЕМУ, без эмодзи.

Этот playbook — не пересказ теории, а рабочая инструкция «как мы тестируем именно Hub». Все примеры используют реальные пакеты (`tracker`, `tms`, `auth`, `messenger`, `calls`), реальные паттерны (`mockRepo`, env `Builder`, `factory`, POM, Steps, `renderWithProviders`, MSW) и реальные команды CI. Каждый раздел содержит блоки good/bad и чек-листы, чтобы агент мог свериться перед сдачей работы верификатору.

---

## Оглавление

1. [Кто и когда использует playbook (роли и разделение доменов)](#1-кто-и-когда-использует-playbook)
2. [Философия тестирования (пирамида, антипаттерны)](#2-философия-тестирования)
3. [Техники тест-дизайна](#3-техники-тест-дизайна)
4. [Backend unit/integration (Go)](#4-backend-unitintegration-go)
5. [Frontend unit (Vitest + RTL)](#5-frontend-unit-vitest--rtl)
6. [E2E (Playwright + POM)](#6-e2e-playwright--pom)
7. [Тестовая документация с привязкой к автотестам (traceability)](#7-тестовая-документация-с-привязкой-к-автотестам)
8. [Allure в деталях](#8-allure-в-деталях)
9. [Замер времени прогонов и производительность](#9-замер-времени-прогонов-и-производительность)
10. [Логирование в тестах и артефакты](#10-логирование-в-тестах-и-артефакты)
11. [Регрессия и qa-gate](#11-регрессия-и-qa-gate)
12. [Test triage (функционал → код vs тест)](#12-test-triage)
13. [Стабильность и флейки](#13-стабильность-и-флейки)
14. [Coverage](#14-coverage)
15. [Definition of Done для тест-задачи](#15-definition-of-done)
16. [Антипаттерны тестов](#16-антипаттерны-тестов)
17. [Чек-лист перед передачей верификатору](#17-чек-лист-перед-передачей-верификатору)

---

## 1. Кто и когда использует playbook

### 1.1. Карта ролей QA-команды Hub

Hub использует ролевую модель агентов (см. `AGENTS.md`). В тестировании участвуют пять специализированных агентов плюс исполнители `be`/`fe`, которые пишут тесты рядом с кодом.

| Агент | Основная зона | Пишет | Не делает |
|-------|---------------|-------|-----------|
| `qa-manual` | Ручные тест-кейсы, TC-матрица, исследовательское тестирование | TC-`<MODULE>`-NN в `docs/TESTING.md`, чек-листы, баг-репорты | Не пишет автотесты; не линкует Go-функции |
| `qa-auto1` | Автоматизация backend (integration/API) + часть E2E — **по верифицированным TC** | Go `_test.go` в `backend/testframework/{domain}/`, часть Playwright | Не переписывает ручные TC; не пишет unit-тесты за разработчиков; не стартует до verified тестовой доки |
| `qa-auto2` | Автоматизация frontend (компонентные) + часть E2E — **по верифицированным TC** | Vitest в `stages/unit/frontend-test/domains/{domain}/`, часть Playwright | Не трогает backend-домены `qa-auto1`; не пишет unit-тесты за разработчиков; не стартует до verified тестовой доки |
| `qa-auto-ver` | Верификатор автотестов | Прогон, проверка триажа, проверка привязки TC↔автотест | Не пишет тесты сам (только контр-примеры) |
| `qa-manual-ver` | Верификатор ручной документации | Проверка полноты TC, соответствия спецификации, покрытия acceptance criteria | Не пишет TC сам |
| `be` / `fe` | Разработчики | Unit-тесты рядом с фичей (обязательно по rule 56) | Не владеют TC-матрицей, но обязаны линковать новые автотесты |

### 1.2. Когда открывать этот playbook

- **Перед началом любой тест-задачи** — свериться со стандартом (структура теста, паттерны, DoD).
- **При падении теста** — раздел [12. Test triage](#12-test-triage) обязателен до любых правок.
- **Перед хендоффом верификатору** — раздел [17](#17-чек-лист-перед-передачей-верификатору).
- **При добавлении нового TC или автотеста** — раздел [7](#7-тестовая-документация-с-привязкой-к-автотестам).
- **При работе с Allure/TMS-импортом** — раздел [8](#8-allure-в-деталях).
- **При жалобе на медленный CI** — раздел [9](#9-замер-времени-прогонов-и-производительность).

### 1.3. Канонический QA-конвейер (порядок обязателен — решение пользователя 2026-07-03)

Поток «бизнес-дока → тест-кейс → верификация доки → автотест → верификация автотестов» строго последователен. **Юнит-тесты в этот конвейер не входят — их пишут разработчики (`be*`/`fe*`) вместе с кодом.**

```
BA/SA (бизнес-дока: vision_and_scope, system_specification — US/AC/FR)
      │
      ▼
qa-manual      ── анализирует бизнес-доку → пишет TC-<MODULE>-NN в docs/TESTING.md
      │            (шаги, ожидаемый результат; колонка «Автотест» = TODO/«ручной»)
      ▼
qa-manual-ver  ── ВЕРИФИЦИРУЕТ тестовую доку (полнота, техники, трассируемость)
      │            без verified:true автотесты НЕ стартуют
      ▼
qa-auto1/2     ── автоматизирует ВЕРИФИЦИРОВАННЫЕ TC → Go integration/API,
      │            Vitest-компонентные, Playwright E2E; обновляет колонку «Автотест»
      ▼
qa-auto-ver    ── верифицирует автотесты (детерминизм, привязка TC↔тест, green regression)

Параллельно (вне конвейера): be1/be2, fe1/fe2 пишут UNIT-тесты своего кода
вместе с фичей (rule 56) — это обязанность разработчика, не qa-auto.
```

**Ключевые правила:**
1. **Порядок жёсткий:** автотесты пишутся только по TC, прошедшим `qa-manual-ver`. Автоматизация неверифицированной доки = переделка.
2. **Юниты — зона разработчиков.** `qa-auto*` не пишут unit-тесты за `be`/`fe`; их слой — integration/API/компонентные/E2E поверх верифицированных TC.
3. Ручной TC и автотест — одна пара: `qa-manual` владеет «что и зачем», `qa-auto` — «как». Колонка «Автотест» в `docs/TESTING.md` — точка встречи (раздел 7).

> **Целевое состояние фреймворка (решение пользователя 2026-07-03, фаза TF-PY — после микросервисного дробления Track P M2–M4):** юниты остаются с кодом (Go — backend, TS/Vitest — frontend); **весь не-юнит слой (integration/API/contract/smoke/regression/E2E) переезжает на Python** (pytest + allure-pytest + httpx + playwright-python + pytest-xdist) в `testframework/` по доменам/сервисам; **нагрузочные — Locust** (`testframework/load/`, те же фикстуры), причём Locust-baseline заводится раньше остального (до выноса realtime). Отчётность — allure-pytest → Hub TMS (webhook), привязка TC через tms-link/ASID. До завершения TF-PY действует текущий стек этого playbook (Go/Vitest/Playwright-TS). План: `.planning/july-2026/ROADMAP-FINAL-2026-07-03.md` фаза TF-PY.

### 1.4. Разделение доменов между `qa-auto1` и `qa-auto2` без конфликтов

Конфликты в тестах возникают из-за: (а) правки одного файла двумя агентами; (б) дублирующихся `_test.go`; (в) гонок за общий fixture/seed. Hub разделяет ответственность **по слою пирамиды и домену**, а не по фиче.

**Правило разделения (по умолчанию):**

| Слой / стадия | Владелец | Путь |
|---------------|----------|------|
| Backend unit/integration (Go) | `qa-auto1` | `backend/testframework/{domain}/` |
| Frontend unit (Vitest + RTL) | `qa-auto2` | `testframework/stages/unit/frontend-test/domains/{domain}/` |
| E2E (Playwright) | делят по доменам (см. ниже) | `testframework/stages/e2e/domains/{domain}/` |

**E2E делится по доменам, а не по файлам:**

| Домен E2E | Владелец |
|-----------|----------|
| auth, navigation, tracker, tms, messenger | `qa-auto1` |
| admin, ai, platform, v4 | `qa-auto2` |

**Протокол бесконфликтной работы:**

1. **Один файл — один владелец в рамках задачи.** Если нужно тронуть чужой файл, это хендофф: оставь заметку в `.planning/STATE.md` (session note) и договорись, кто мержит.
2. **Новые файлы именуются по домену + волне:** `service_wave<NN>_test.go`, `handler_<feature>_flow_test.go` (см. реальные имена в `backend/testframework/tracker/`). Это исключает коллизии имён между `qa-auto1` и `qa-auto2`.
3. **Общие фикстуры не форкать.** Env `Builder` (`backend/testframework/env/`), `factory` и E2E `shared/` — общий код. Изменения в них — через `be`/`fe` или согласованный хендофф, потому что они влияют на оба потока.
4. **Seed изолирован по тесту.** Каждый Go-тест поднимает свой in-memory sqlite (уникальный DSN с `uuid.New()`), каждый Vitest — свой `QueryClient`. Никакого общего состояния между тестами (см. разделы 4-5).
5. **При параллельном прогоне** (`t.Parallel`, Playwright workers) домены не должны шарить внешние ресурсы (один и тот же проект в БД, один и тот же канал). Используй уникальные суффиксы: `Date.now().toString(36)` в E2E, `uuid.New()` в Go.

**Good — разделение по слою:**

```text
qa-auto1: backend/testframework/tracker/service_wave126_test.go   (Go)
qa-auto2: testframework/stages/unit/frontend-test/domains/tracker/issue-card.test.tsx  (Vitest)
```

Оба покрывают домен tracker, но не пересекаются по файлам и слоям — merge без конфликтов.

**Bad — оба правят один файл:**

```text
qa-auto1 и qa-auto2 одновременно добавляют тесты в
backend/testframework/tracker/service_test.go
→ git conflict, дублирование mockRepo, гонка за newTestService()
```

### 1.5. Чек-лист «перед стартом тест-задачи»

- [ ] Прочитал требование/US в `docs/vision_and_scope.md` или `docs/system_specification.md`.
- [ ] Понял, к какому TC-модулю относится задача (`docs/TESTING.md`).
- [ ] Определил свой слой и домен (раздел 1.4) — нет ли пересечения с другим `qa-auto`.
- [ ] Проверил, существует ли уже автотест/TC для сценария (не дублировать).
- [ ] Знаю команду прогона своего слоя (раздел 11).

---

## 2. Философия тестирования

### 2.1. Пирамида тестирования Hub

Hub придерживается классической пирамиды: много быстрых unit-тестов, меньше интеграционных, ещё меньше медленных E2E. Форма закреплена в `docs/test_plan.md`.

```
        ┌─────────────┐
        │     E2E     │  Playwright + POM — критические user paths
        │  20 spec    │  jobs e2e-smoke (3) / e2e-full (20)
        ├─────────────┤
        │ Integration │  httptest handlers, enttest/sqlite, Mongo в CI
        │             │  (в составе job backend-test)
        ├─────────────┤
        │    Unit     │  Go testify (-tags hubtest) + Vitest/RTL
        │ 499 BE + 99 │  jobs backend-test / frontend-test
        └─────────────┘
```

| Уровень | Инструмент | Расположение | CI job | Скорость | Цель |
|---------|------------|--------------|--------|----------|------|
| Unit (BE) | Go testing + testify | `backend/testframework/{domain}/` | `backend-test` | мс | логика сервисов, конвертеры, валидация |
| Unit (FE) | Vitest + RTL | `testframework/stages/unit/frontend-test/domains/{domain}/` | `frontend-test` | мс | компоненты, хуки, сторы, api-клиенты |
| Integration | httptest, enttest, miniredis | `backend/testframework/` | в составе `backend-test` | десятки мс | handler↔service↔repo, реальная схема Ent |
| E2E | Playwright + POM + Steps | `testframework/stages/e2e/domains/{domain}/` | `e2e-smoke`, `e2e-full` | секунды | критические сквозные пути |

### 2.2. Что тестировать на каждом уровне

**Unit (Go) — большинство тестов.** Быстрые, детерминированные, с мок-репозиторием:
- бизнес-логика сервисов (создание задачи, оптимистичная блокировка, workflow-переходы);
- валидация входных данных (невалидные UUID, пустые поля, границы);
- конвертеры `ent.*` → DTO (`toIssueResponse`, `toCommentResponse` — через `export_hooks.go`);
- error paths (что возвращается при not found, conflict, forbidden).

**Unit (Frontend) — компоненты и хуки:**
- рендер компонента с разными пропсами (`IssueCard` со story points / без);
- поведение хуков (`useAuth.logout` вызывает API до очистки стора);
- сторы Zustand (переходы состояния);
- api-клиенты (правильный URL и payload через `vi.spyOn(api, 'post')`).

**Integration (Go) — стыки слоёв на реальной схеме:**
- handler → service → repository с in-memory sqlite (env `Builder`);
- RBAC-проверки на реальных ролях;
- пагинация, фильтры, каскады Ent.

**E2E — только критические пути:**
- вход (dev-login), создание проекта, отправка сообщения, TMS-флоу, Ctrl+K, logout, a11y;
- не дублировать unit-проверки; E2E дорог и медленен.

### 2.3. Правило распределения усилий

**Good — правильная пирамида:**

```text
Новая фича «WIP-лимиты на колонке Kanban»:
- 6 Go unit-тестов (лимит достигнут, превышен, снят, дефолт) — mockRepo
- 2 integration-теста (реальный board через env.Builder)
- 1 Vitest (компонент колонки показывает счётчик N/limit)
- 0 новых E2E (покрыто существующим project-lifecycle, если критично)
```

**Bad — перевёрнутая пирамида (ice-cream cone):**

```text
Та же фича:
- 1 Go unit-тест «на всякий»
- 5 E2E-тестов, каждый логинится, создаёт проект, задачу, тащит на доску,
  проверяет лимит через UI
→ CI +90 секунд, флейки на auto-wait, лимит GitHub Actions горит,
  а логика лимита по-прежнему почти не покрыта unit-тестами
```

### 2.4. Антипаттерны на уровне стратегии

| Антипаттерн | Симптом | Чем плохо | Как правильно |
|-------------|---------|-----------|---------------|
| Ice-cream cone | Много E2E, мало unit | Медленный CI, флейки, дорогая отладка | Логику — в unit, E2E — только критический путь |
| Flaky-терпимость | «Перезапусти, обычно проходит» | Теряется доверие к CI, маскирует реальные баги | Раздел 13: детект, quarantine, фикс причины |
| Тест ради покрытия | Тест без ассертов / проверяет мок | Coverage растёт, баги не ловятся | Проверять поведение и контракт, а не строки |
| Зелёные вслепую | Assert подгоняется под текущий вывод | Тест фиксирует баг как «правильное» поведение | Раздел 12: сначала функционал по спеке |
| Общий мутабельный fixture | Тесты зависят от порядка | Гонки, падения при `-race`/parallel | Изоляция: свой sqlite/QueryClient на тест |
| Тестирование деталей реализации | Assert на приватные поля/порядок вызовов | Ломается при рефакторинге без изменения поведения | Assert на публичный контракт |

### 2.5. Принципы Hub (закреплены в rules)

1. **Backend CI-инвариант:** `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...`. `-race` всегда, `-count=1` отключает кэш.
2. **Coverage floor:** statement coverage `internal/` ≥ 82% (сейчас 83.0%); security-пакеты ≥ 80%.
3. **Тесты в модуле backend:** Go запрещает импорт `internal/*` из другого модуля, поэтому тесты живут в `backend/testframework/` и используют `export_hooks.go` для white-box доступа.
4. **Triage перед фиксом:** падающий тест ≠ баг в коде. Сначала спека (раздел 12).
5. **Docs sync:** новый автотест → обновить колонку «Автотест» в `docs/TESTING.md` в той же сессии (rule 55).
6. **Cleanup:** coverage-профили не коммитить, писать в TMP (rule 45).

---

## 3. Техники тест-дизайна

Тест-дизайн — это как из бесконечного пространства входов выбрать минимальный набор, ловящий максимум багов. Ниже — шесть техник с примерами на реальных фичах Hub.

### 3.1. Классы эквивалентности

Разбиваем вход на классы, внутри которых система ведёт себя одинаково. Тестируем по одному представителю класса.

**Пример: тип задачи (`issue_type`) в трекере.**

Валидный класс — 9 типов: `bug, task, story, epic, sub_task, improvement, tech_debt, spike, release`. Невалидный класс — всё остальное.

Реальный тест `TestCreateIssueAllTypes` берёт по представителю валидного класса (все 9 через `t.Run`):

```go
func TestCreateIssueAllTypes(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issueTypes := []string{
		"bug", "task", "story", "epic", "sub_task",
		"improvement", "tech_debt", "spike", "release",
	}

	for i, it := range issueTypes {
		t.Run(it, func(t *testing.T) {
			req := tracker.CreateIssueRequest{
				Title:     fmt.Sprintf("Issue type: %s", it),
				IssueType: it,
				Priority:  "medium",
			}
			resp, err := svc.CreateIssue(ctx, projectID, userID, req)
			require.NoError(t, err)
			assert.Equal(t, it, resp.IssueType)
		})
	}
}
```

**Классы для типовых полей Hub:**

| Поле | Валидные классы | Невалидные классы |
|------|-----------------|-------------------|
| `email` | `user@hub.local` | пустой, без `@`, без домена |
| `uuid` | валидный v4 | `not-uuid`, пустой, префикс |
| `issue_type` | 9 enum-значений | `foo`, пустой |
| `priority` | `low/medium/high/critical` | `urgent`, число |
| `case_type` (TMS) | `classic`, `bdd` | `unknown` |
| `status` (Allure) | `passed/failed/broken/skipped` | `weird_status` → `unknown` |

### 3.2. Граничные значения

Баги живут на границах. Для каждого числового/размерного ограничения тестируем: граница−1, граница, граница+1.

**Пример: rate limiting auth (3 попытки / 10 минут → 4-я = 429).**

| Вход | Ожидание |
|------|----------|
| 3-й запрос | 200/OK (граница) |
| 4-й запрос | 429 (граница+1) |

**Пример: лимит сообщений мессенджера 60/мин.**

| Вход | Ожидание |
|------|----------|
| 60-е сообщение | OK |
| 61-е сообщение | 429 (`MessageSendRateLimit`) |

**Пример: длина текста чата звонка (`TestSendMessage_TextTooLong`).** Тестируем max и max+1.

**Пример: файл 25 MB (Fiber BodyLimit).** 25 MB — OK, 25 MB + 1 байт — отклонён.

**Границы Hub, которые нельзя забыть:**

| Ограничение | Значение | Тест-граница |
|-------------|----------|--------------|
| Auth rate limit | 3 / 10 мин | 3 OK, 4 → 429 |
| API rate limit | 100 / сек | 100 OK, 101 → 429 |
| File upload rate | 10 / мин | 10 OK, 11 → 429 |
| Message rate | 60 / мин | 60 OK, 61 → 429 |
| Max file size | 25 MB | 25 MB OK, +1 → отклонён |
| JWT access TTL | 15 мин | до 15 — валиден, после — 401 |
| Refresh TTL | 30 дней | до — валиден, после — 401 |
| Pending approve | 7 дней | до — есть, после — delete |
| Soft delete корзина | 30 дней | до — restore, после — purge |
| Cursor pagination | limit ≤ 100, дефолт 50 | 0/отриц → 50, >100 → 50 |

**Пример из кода — граница пагинации в `mockRepo.ListIssues`:**

```go
limit := params.Limit
if limit <= 0 || limit > 100 {
	limit = 50 // дефолт при выходе за границы
}
```

Тест `TestListIssuesPagination` проверяет, что при `Limit: 2` возвращается ≤ 2, а `Total` = 5.

### 3.3. Попарное тестирование (pairwise)

Когда параметров много, полный перебор комбинаций невозможен. Pairwise покрывает все пары значений — эмпирически ловит большинство багов при кратном сокращении числа тестов.

**Пример: фильтр задач `status × priority × assignee`.**

Наивно: 4 статуса × 4 приоритета × 3 назначенца = 48 комбинаций. Pairwise сокращает до ~12-16, покрывая каждую пару (status,priority), (status,assignee), (priority,assignee).

```go
func TestIssueFilter_Pairwise(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	// Каждая строка покрывает ранее не встречавшиеся пары значений.
	cases := []struct {
		name     string
		status   string
		priority string
		assignee string
	}{
		{"todo+low+me", "todo", "low", "me"},
		{"todo+high+other", "todo", "high", "other"},
		{"inprogress+low+other", "in_progress", "low", "other"},
		{"inprogress+critical+me", "in_progress", "critical", "me"},
		{"done+medium+none", "done", "medium", "none"},
		// ... сгенерировано pairwise-инструментом
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := tracker.ListIssuesRequest{
				Status:   tc.status,
				Priority: tc.priority,
				Assignee: tc.assignee,
			}
			_, err := svc.ListIssues(ctx, uuid.New(), req)
			// Комбинация валидна → без ошибки; фокус на отсутствии паник и корректной фильтрации.
			require.NoError(t, err)
		})
	}
}
```

**Где pairwise полезен в Hub:** фильтры трекера, комбинации ролей RBAC × действий, настройки уведомлений (тип × канал × quiet hours), Allure matching (наличие link × label AS_ID × fullName × className).

### 3.4. Таблицы решений

Когда результат зависит от комбинации условий по чётким правилам — строим таблицу решений и покрываем каждую строку.

**Пример: workflow-переходы (разрешён/запрещён).**

| Условие: переход задан в workflow? | Действие |
|-----------------------------------|----------|
| Да (ToDo → In Progress) | 200, статус изменён |
| Нет (произвольный статус) | 400, переход не разрешён |

Реальные тесты — пара `TestMoveIssueValidTransition` / `TestMoveIssueInvalidTransition`:

```go
func TestMoveIssueInvalidTransition(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	created, err := svc.CreateIssue(ctx, uuid.New(), uuid.New(), tracker.CreateIssueRequest{
		Title: "Invalid move", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	// Переход в статус без определённого transition → 400.
	_, err = svc.MoveIssue(ctx, created.IssueKey, uuid.New(), tracker.MoveIssueRequest{
		StatusID: uuid.New().String(),
		Rank:     "aaa",
	})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 400, appErr.Code)
}
```

**Пример: RBAC-матрица (роль × действие).** Таблица решений — идеальный инструмент: для каждой пары (роль, endpoint) есть чёткое «разрешено/403».

**Пример: read-only канал мессенджера.**

| is_readonly | Роль | Результат отправки |
|-------------|------|--------------------|
| false | любой участник | 201 |
| true | Admin/Owner | 201 (`TestSendMessage_ReadOnlyAllowed_Admin`) |
| true | обычный участник | 403 (`TestSendMessage_ReadOnlyBlocked_RegularUser`) |

### 3.5. Переходы состояний (state machine)

Многие сущности Hub — конечные автоматы. Тестируем валидные переходы, запрещённые переходы и инварианты.

**Пример: жизненный цикл спринта.** `planning → active → completed`.

| Из | Действие | В | Тест |
|----|----------|---|------|
| planning | start | active | `TestStartSprint_Success` |
| active (другой) | start второго | ошибка «уже есть active» | `TestStartSprint_RejectsWhenActiveExists` |
| active | complete | completed (+ перенос незакрытых) | `TestCompleteSprint_MovesIncompleteIssues` |

**Пример: статус звонка.** `active → ended`; запись: `idle → recording → processing`.

**Пример: тест-ран TMS.** `created → in_progress → completed`, с пересчётом статистики при каждом результате (`TestExecuteTestResult_UpdatesStats`).

**Диаграмма состояний спринта:**

```
planning ──start──▶ active ──complete──▶ completed
   │                  │
   │                  └─ start(второй) ──▶ ERROR (уже есть active)
```

### 3.6. Use-case тестирование

Тестируем сквозной пользовательский сценарий целиком — обычно на уровне integration или E2E. Каждый шаг use-case превращается в шаг теста.

**Пример use-case «Тестировщик прогоняет тест-ран и заводит баг из падения»:**

1. Создать тест-кейс (`TestCreateTestCase_Classic`).
2. Создать прогон и добавить кейсы (`TestAddTestCasesToRun`).
3. Выполнить результат `failed` (`TestExecuteTestResult_UpdatesStats`).
4. Из failed-результата создать Bug (`TestCreateBugFromFailedResult`).

**Пример use-case E2E «Жизненный цикл проекта»** — реальный `project-lifecycle.spec.ts`: логин → создать проект → проверить доску → удалить → восстановить из корзины. Один тест покрывает весь путь.

### 3.7. Чек-лист тест-дизайна

- [ ] Выделены классы эквивалентности для каждого входного поля.
- [ ] Для каждого числового/размерного ограничения есть тесты границ (−1, граница, +1).
- [ ] Многопараметрические комбинации сокращены pairwise, а не полным перебором.
- [ ] Условная логика покрыта таблицей решений (каждая строка = тест).
- [ ] Автоматы покрыты по валидным и запрещённым переходам + инварианты.
- [ ] Ключевые use-case покрыты сквозным integration/E2E-тестом.
- [ ] Error paths покрыты (not found → 404, conflict → 409, forbidden → 403, bad request → 400).

---

## 4. Backend unit/integration (Go)

Backend-тесты — фундамент пирамиды Hub (499 `_test.go` файлов, filesystem count
2026-07-26). Из-за ограничения Go на импорт `internal/*` из другого модуля они
физически лежат в `backend/testframework/{domain}/`, но принадлежат модулю
`github.com/valerii/hub/backend`.

### 4.1. Обязательный каркас теста

Каждый backend-тестовый файл начинается со стандартного заголовка:

```go
// Package tracker содержит юнит-тесты для tracker.TrackerService.
package tracker_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)
```

Правила каркаса:

1. **Пакет с суффиксом `_test`** (`tracker_test`, `tms_test`) — black-box тестирование через публичный API пакета. White-box доступ к приватным функциям — только через `export_hooks.go` (см. 4.10).
2. **testify:** `require` для критичных проверок (останавливают тест), `assert` для не критичных (продолжают).
3. **Контекст всегда первым** аргументом сервисов: `ctx := context.Background()`.
4. **UUID через `uuid.New()`**, не строки.

### 4.2. Тег сборки и запуск

Тесты, использующие `export_hooks.go`, требуют тег `hubtest`. Канонический прогон (как в CI job `backend-test`):

```bash
cd backend
go generate ./ent
go vet -tags hubtest ./...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...
```

| Флаг | Зачем |
|------|-------|
| `-tags hubtest` | активирует `export_hooks.go` (`//go:build hubtest`) |
| `-race` | детектор гонок — обязателен всегда (rule 40) |
| `-count=1` | отключает кэш результатов тестов |
| `-timeout=600s` | верхняя граница на весь прогон |

Один домен: `go test -tags hubtest ./testframework/tracker/...`. Один тест: `go test -tags hubtest -run TestCreateIssue ./testframework/tracker/...`.

### 4.3. Table-driven tests

Основной паттерн Go. Таблица случаев + цикл с `t.Run`. Реальный пример `TestCreateIssue_InvalidUUIDs`:

```go
func TestCreateIssue_InvalidUUIDs(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		req  tracker.CreateIssueRequest
		err  string // подстрока ожидаемой ошибки
	}{
		{
			name: "invalid assignee_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", AssigneeID: "not-uuid"},
			err:  "assignee_id",
		},
		{
			name: "invalid parent_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", ParentID: "bad"},
			err:  "parent_id",
		},
		{
			name: "invalid status_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", StatusID: "bad"},
			err:  "status_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateIssue(ctx, projectID, userID, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}
```

**Good:** каждый случай — отдельный подтест с осмысленным `name`; при падении видно, какая строка упала. **Bad:** один тест с 10 последовательными `assert` без `t.Run` — при падении первого остальные не проверяются, а имя не говорит, что сломалось.

### 4.4. `require` vs `assert` — когда что

| Ситуация | Использовать | Почему |
|----------|--------------|--------|
| Проверка `err` перед использованием результата | `require.NoError(t, err)` | если ошибка есть, дальше `resp` — nil, будет паника |
| Проверка `resp != nil` перед разыменованием | `require.NotNil(t, resp)` | защита от nil-паники |
| Type assertion (`err.(*apperrors.AppError)`) | `require.True(t, ok)` | без валидного каста дальше некорректно |
| Проверка значения поля | `assert.Equal(t, want, got)` | не критично для продолжения, показать все расхождения |

**Good — require перед разыменованием:**

```go
resp, err := svc.CreateIssue(ctx, projectID, userID, req)
require.NoError(t, err)      // стоп, если ошибка
require.NotNil(t, resp)      // стоп, если nil
assert.Equal(t, "TEST-1", resp.IssueKey)  // дальше безопасно
assert.Equal(t, 1, resp.Version)
```

**Bad — assert перед разыменованием:**

```go
resp, err := svc.CreateIssue(ctx, projectID, userID, req)
assert.NoError(t, err)              // тест продолжится даже при ошибке
assert.Equal(t, "TEST-1", resp.IssueKey)  // паника: resp == nil
```

### 4.5. Проверка доменных ошибок (`apperrors.AppError`)

Hub возвращает типизированные ошибки с HTTP-кодом. Проверяем не текст, а код:

```go
func TestGetIssue_NotFound(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.GetIssue(ctx, "MISSING-1")
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 404, appErr.Code)
}
```

Матрица кодов, которую нужно покрывать (из `docs/TESTING.md`): 400 (bad request), 401 (unauth), 403 (RBAC), 404 (not found), 409 (version conflict), 429 (rate limit).

### 4.6. Subtests, `t.Parallel`, `t.Helper`, `t.Cleanup`

**`t.Run` — группировка и изоляция подслучаев** (см. 4.3).

**`t.Parallel` — параллельный прогон** независимых тестов. Ускоряет прогон, но требует полной изоляции состояния. В Hub env `Builder` даёт каждому тесту свой in-memory sqlite (уникальный DSN), поэтому параллелизм безопасен:

```go
func TestSomethingIndependent(t *testing.T) {
	t.Parallel() // безопасно: свой sqlite на тест
	e := env.New(t, "trk").Build()
	// ...
}
```

**Bad — `t.Parallel` с общим состоянием:**

```go
var sharedRepo = newMockRepo() // ПЛОХО: общий на все тесты

func TestA(t *testing.T) {
	t.Parallel()
	sharedRepo.CreateIssue(...) // гонка при -race
}
```

**`t.Helper` — помечает вспомогательную функцию**, чтобы стек ошибки указывал на вызывающий тест, а не на хелпер. Реальный пример из env `Builder`:

```go
func New(t *testing.T, prefix string) *Builder {
	t.Helper() // ошибки Build() укажут на тест, а не на эту строку
	if prefix == "" {
		prefix = "tf"
	}
	return &Builder{t: t, prefix: prefix}
}
```

**`t.Cleanup` — освобождение ресурсов** после теста (вместо `defer`, работает и в подтестах). Реальный пример — закрытие sqlite-клиента:

```go
func (b *Builder) Build() *Env {
	b.t.Helper()
	dsn := "file:" + b.prefix + "_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(b.t, "sqlite3", dsn)
	b.t.Cleanup(func() { _ = client.Close() }) // закрытие после теста
	// ...
}
```

### 4.7. Мок репозитория (паттерн `mockRepo`)

Основной способ изолировать сервис от БД в unit-тестах — ручной мок, реализующий интерфейс `Repository`. Пример из `tracker/service_test.go`: `mockRepo` хранит данные в `map` и реализует все методы интерфейса.

```go
// mockRepo реализует tracker.Repository для тестирования TrackerService.
type mockRepo struct {
	issues          map[uuid.UUID]*ent.Issue
	issuesByKey     map[string]*ent.Issue
	statuses        map[uuid.UUID]*ent.WorkflowStatus
	transitions     map[string]*ent.WorkflowTransition // ключ "fromID:toID"
	projectCounters map[uuid.UUID]int
	updateVersionError bool // флаг для эмуляции конфликта версий
}

func newMockRepo() *mockRepo {
	defaultStatusID := uuid.New()
	return &mockRepo{
		issues:      make(map[uuid.UUID]*ent.Issue),
		issuesByKey: make(map[string]*ent.Issue),
		statuses: map[uuid.UUID]*ent.WorkflowStatus{
			defaultStatusID: {ID: defaultStatusID, Name: "To Do", IsDefault: true},
		},
		transitions:     make(map[string]*ent.WorkflowTransition),
		projectCounters: make(map[uuid.UUID]int),
	}
}
```

Метод мока эмулирует поведение реального репозитория, включая ошибки:

```go
func (m *mockRepo) UpdateIssue(_ context.Context, issueID uuid.UUID, input tracker.UpdateIssueInput) (*ent.Issue, error) {
	i, ok := m.issues[issueID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	// Эмуляция оптимистичной блокировки: версия из запроса должна совпасть.
	if m.updateVersionError || i.Version != input.Version {
		return nil, apperrors.NewConflict("version mismatch")
	}
	if input.Title != nil {
		i.Title = *input.Title
	}
	i.Version++
	return i, nil
}
```

**Хелпер сборки сервиса** держим рядом:

```go
func newTestService() (*tracker.TrackerService, *mockRepo) {
	repo := newMockRepo()
	svc := tracker.NewTrackerService(repo, nil)
	return svc, repo
}
```

**Частичный мок через embedded interface** (для больших интерфейсов) — паттерн из `tms/allure_parser_test.go`. Встраиваем интерфейс и переопределяем только нужные методы; остальные дают nil-заглушки:

```go
type mockMatcherRepo struct {
	tms.Repository // embedded: неиспользуемые методы — заглушки
	externalIDMap  map[string]*ent.TestCase
	fullNameMap    map[string]*ent.TestCase
}

func (m *mockMatcherRepo) FindTestCaseByExternalID(_ context.Context, _ uuid.UUID, id string) (*ent.TestCase, error) {
	if tc, ok := m.externalIDMap[id]; ok {
		return tc, nil
	}
	return nil, &ent.NotFoundError{}
}
```

**Good:** мок эмулирует контракт (not found → `&ent.NotFoundError{}`, конфликт → `apperrors.NewConflict`). **Bad:** мок всегда возвращает `nil, nil` — тест «зелёный», но не проверяет реальное поведение сервиса на ошибках.

### 4.8. Env `Builder` + `Factory` (integration на реальной схеме)

Когда нужна реальная схема Ent (каскады, индексы, реальные запросы) — используем in-memory sqlite через env `Builder`. Он готовит user + project + дефолтный статус:

```go
func TestBoardData_RealSchema(t *testing.T) {
	e := env.New(t, "board").Build() // sqlite + user + project + status
	repo := tracker.NewRepository(e.Client)
	svc := tracker.NewTrackerService(repo, nil)

	_, err := svc.CreateIssue(e.Ctx, e.ProjectID, e.UserID, tracker.CreateIssueRequest{
		Title: "Real", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
}
```

**Factory — создание дополнительных сущностей** (паттерн из `factory/user.go`):

```go
func TestMultipleUsers(t *testing.T) {
	e := env.New(t, "usr").Build()
	uf := factory.NewUserFactory(t, e.Client, e.Ctx)

	alice := uf.Create("alice") // уникальный email через uuid
	bob := uf.Create("bob")
	require.NotEqual(t, alice.ID, bob.ID)
}
```

Factory гарантирует уникальность (`email = prefix + "-" + uuid[:8] + "@test.com"`), поэтому тесты не конфликтуют по unique-констрейнтам.

### 4.9. Мок vs env.Builder (testcontainers) — выбор

| Критерий | Мок (`mockRepo`) | env.Builder (sqlite) | testcontainers (Postgres/Mongo) |
|----------|------------------|----------------------|---------------------------------|
| Скорость | мкс | мс | секунды (старт контейнера) |
| Что проверяет | логику сервиса | логику + реальную схему Ent | + реальные особенности PG/Mongo |
| Когда | большинство unit | integration слоёв | только специфика PG/Mongo (JSONB, полнотекст, транзакции) |
| CI-стоимость | нулевая | низкая | высокая (тянет образы) |

**Правило Hub:** по умолчанию — мок. Нужна реальная схема — `env.Builder` (sqlite). testcontainers/реальный Mongo — только в CI job (Mongo 8 сервис) и только там, где sqlite неадекватен (реальные Mongo-репозитории `*_mongo_test.go`). Не тащить Docker в каждый unit-тест — это ломает быстрый локальный цикл и жрёт лимит GitHub Actions.

### 4.10. White-box доступ через `export_hooks.go`

Приватные функции пакета нельзя вызвать из `tracker_test`. Для их тестирования пакет `internal/tracker` содержит `export_hooks.go` с тегом `//go:build hubtest`, экспортирующий приватные символы:

```go
//go:build hubtest

package tracker

// ExportBuildFieldChanges строит список изменений полей задачи для hubtest.
func ExportBuildFieldChanges(existing *ent.Issue, req UpdateIssueRequest) []FieldChange {
	return buildFieldChanges(existing, req)
}

// ExportTrackerServiceWorkflow возвращает workflow-движок для hubtest.
func ExportTrackerServiceWorkflow(s *TrackerService) *WorkflowEngine {
	return s.workflow
}
```

Тест вызывает экспортированную обёртку:

```go
func TestWorkflowValidation(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	defaultID := repo.getDefaultStatusID()
	inProgressID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, defaultID)

	// Доступ к приватному workflow-движку через export hook.
	err := tracker.ExportTrackerServiceWorkflow(svc).ValidateTransition(ctx, defaultID, inProgressID)
	assert.NoError(t, err)

	err = tracker.ExportTrackerServiceWorkflow(svc).ValidateTransition(ctx, inProgressID, defaultID)
	assert.Error(t, err) // обратный переход не определён
}
```

**Правила `export_hooks.go`:**
- Обязательно `//go:build hubtest` — в prod-сборке (`-tags release`) этот файл не компилируется.
- Именование `Export<PrivateName>` — единый префикс.
- Комментарий на русском с суффиксом «для hubtest».
- Экспортировать минимально необходимое, не публичный API «на всякий».

### 4.11. HTTP handler-тесты (httptest)

Handler-слой тестируется через `httptest`/Fiber test app: проверяем статус-коды и тело. Матрица ожидаемых кодов — в `docs/TESTING.md` (раздел «API-тесты по модулям»). Пример логики:

```go
func TestCreateIssueHandler_Unauthorized(t *testing.T) {
	app := setupTestApp(t) // Fiber app с роутами tracker
	req := httptest.NewRequest("POST", "/api/projects/"+projID+"/issues",
		strings.NewReader(`{"title":"X","issue_type":"task"}`))
	req.Header.Set("Content-Type", "application/json")
	// без Authorization → ожидаем 401

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}
```

### 4.12. Покрытие и `-race`

```bash
# Профиль ТОЛЬКО в TMP (rule 45), никогда в backend/
go test -tags hubtest -race -coverprofile="$TMPDIR/hub-cov.out" ./testframework/...
go tool cover -func="$TMPDIR/hub-cov.out" | tail -1  # total
```

`-race` обязателен: Hub активно использует горутины (ws, nats, outbox). Гонки ловятся только с ним. Coverage-детали — раздел 14.

### 4.13. Чек-лист backend-теста

- [ ] Пакет `<domain>_test`, импорты сгруппированы, testify подключён.
- [ ] Прогон зелёный с `-tags hubtest -race -count=1`.
- [ ] `require` перед разыменованием, `assert` для значений.
- [ ] Доменные ошибки проверяются по `AppError.Code`, не по тексту.
- [ ] Table-driven для однотипных случаев, `t.Run` с осмысленными именами.
- [ ] Мок эмулирует контракт (ошибки, not found), а не `nil, nil`.
- [ ] Состояние изолировано (свой мок/sqlite на тест); `t.Parallel` только при изоляции.
- [ ] `t.Helper` в хелперах, `t.Cleanup` для ресурсов.
- [ ] White-box только через `export_hooks.go` с `//go:build hubtest`.
- [ ] Coverage-профиль в TMP, не закоммичен.

---

## 5. Frontend unit (Vitest + RTL)

Frontend-тесты (99 тестов, Vitest + React Testing Library) живут в `testframework/stages/unit/frontend-test/domains/{domain}/`, job `frontend-test`. Перед прогоном линкуются node_modules:

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run
# один домен:
npx vitest run ../testframework/stages/unit/frontend-test/domains/auth
```

### 5.1. Рендер через TestRenderer

Компоненты, которым нужны провайдеры (TanStack Query и т.п.), рендерятся через `renderWithProviders` из `render.tsx`. Он даёт каждому тесту свой `QueryClient` с отключёнными retry:

```typescript
import { renderWithProviders } from "../../render";
import { screen } from "@testing-library/react";
import { IssueList } from "@/components/tracker/issue-list";

describe("IssueList", () => {
	it("renders issues from query", async () => {
		renderWithProviders(<IssueList projectId="proj-1" />);
		expect(await screen.findByText("HUB-42")).toBeInTheDocument();
	});
});
```

Для простых компонентов без провайдеров допустим прямой `render` из RTL (как в `issue-card.test.tsx`).

> **(2026) Vitest 4 Browser Mode.** В Vitest 4 (дек 2025) Browser Mode стабилен: компонентные тесты можно гонять в реальном Chromium через provider `playwright()` — jsdom-моки (`matchMedia`, `ResizeObserver`) не нужны. Массово не мигрируем; применяем точечно для компонентов, страдающих от jsdom (DnD, виртуализация, Radix-поповеры), после планового апгрейда Vitest 3→4. Остальное остаётся в jsdom.

**Почему свой QueryClient на тест:** отключённые `retry: false` делают ошибки детерминированными; изоляция клиента исключает утечку кэша между тестами (иначе — флейки от порядка).

### 5.2. Queries: by role / by label — приоритет

RTL-запросы отражают то, как пользователь и скринридер видят UI. Приоритет (сверху вниз):

| Приоритет | Query | Когда |
|-----------|-------|-------|
| 1 | `getByRole` | кнопки, ссылки, заголовки, поля (`getByRole('button', { name: /edit/i })`) |
| 2 | `getByLabelText` | поля формы с label |
| 3 | `getByText` | некликабельный текст, бейджи |
| 4 | `getByTestId` | крайний случай, когда роль/текст недоступны |

**Good — by role/label (реальный `issue-card.test.tsx`):**

```typescript
it("renders issue key and title", () => {
	render(<IssueCard issue={baseIssue} projectId="proj-1" />);
	expect(screen.getByText("HUB-42")).toBeInTheDocument();
	expect(screen.getByText("Fix login bug")).toBeInTheDocument();
});
```

**Bad — хрупкие CSS-селекторы:**

```typescript
// container.querySelector завязан на структуру DOM, ломается при рефакторинге верстки
const { container } = render(<IssueCard issue={baseIssue} projectId="proj-1" />);
expect(container.querySelector(".issue-card__title span")).toHaveTextContent("Fix login bug");
```

### 5.3. Взаимодействие через userEvent

`userEvent` эмулирует реальные события (focus, keydown, click) точнее, чем `fireEvent`. Всегда `await`:

```typescript
import userEvent from "@testing-library/user-event";

it("calls onEdit when edit button clicked", async () => {
	const onEdit = vi.fn();
	render(<IssueCard issue={baseIssue} projectId="proj-1" onEdit={onEdit} />);
	await userEvent.click(screen.getByRole("button", { name: /edit/i }));
	expect(onEdit).toHaveBeenCalledWith(baseIssue.id);
});
```

**Bad:** синхронный `fireEvent.click` без await для сложных взаимодействий (ввод текста, комбо клавиш) — пропускает промежуточные события и даёт ложные результаты.

### 5.4. Мок API: MSW vs vi.spyOn

Два подхода к изоляции сети:

**vi.spyOn на api-клиент** — когда тестируем логику, вызывающую клиент (реальный `admin-api.test.ts`):

```typescript
import { api } from "@/api/client";
import { banUser } from "@/api/admin";

it("banUser posts reason to /admin/users/:id/ban", async () => {
	const postSpy = vi.spyOn(api, "post").mockResolvedValue({ success: true, message: "User banned" });
	const result = await banUser("user-42", "spam");
	expect(postSpy).toHaveBeenCalledWith("/admin/users/user-42/ban", { reason: "spam" });
	expect(result.success).toBe(true);
});
```

**MSW (Mock Service Worker)** — когда компонент делает fetch на уровне HTTP и мы хотим проверить весь путь query → render:

```typescript
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";

const server = setupServer(
	http.get("/api/projects/:id/issues", ({ params }) => {
		return HttpResponse.json({ data: mockIssues, total: mockIssues.length });
	}),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers()); // сброс переопределений между тестами
afterAll(() => server.close());
```

`onUnhandledRequest: "error"` ловит незамоканные запросы — защита от «случайно ходили в реальный бэкенд».

**Выбор:** тестируем чистую функцию/api-обёртку → `vi.spyOn`; тестируем компонент с data-fetching через TanStack Query → MSW.

### 5.5. Тестирование хуков

Хуки тестируются через `renderHook` + `act`. Реальный `use-auth.test.ts`:

```typescript
import { act, renderHook } from "@testing-library/react";
import * as authApi from "@/api/auth";
import { useAuth } from "@/hooks/use-auth";
import { useAuthStore } from "@/stores/auth-store";

it("logout calls POST /auth/logout before clearing store", async () => {
	const logoutSpy = vi.spyOn(authApi, "logout").mockResolvedValue(undefined);
	useAuthStore.getState().setTokens("access-token", "refresh-token");

	const { result } = renderHook(() => useAuth());
	await act(async () => {
		await result.current.logout();
	});

	expect(logoutSpy).toHaveBeenCalledOnce();
	expect(useAuthStore.getState().isAuthenticated).toBe(false);
});
```

`act` оборачивает вызовы, меняющие состояние, чтобы React применил обновления до ассертов.

### 5.6. Тестирование стора (Zustand)

Стор — синглтон, поэтому сбрасываем его в `beforeEach` (реальный паттерн из `use-auth.test.ts`):

```typescript
function resetAuthStore(): void {
	sessionStorage.clear();
	useAuthStore.setState({
		accessToken: null,
		refreshToken: null,
		user: null,
		isAuthenticated: false,
	});
}

describe("useAuth", () => {
	beforeEach(() => {
		resetAuthStore();
		vi.restoreAllMocks(); // снять все spy/mock между тестами
	});
	// ...
});
```

**Правило:** любой глобальный синглтон (стор, sessionStorage) сбрасывается в `beforeEach`; иначе тесты зависят от порядка → флейки.

### 5.7. Snapshot — осторожно

Snapshot-тесты допустимы **только** для стабильного, редко меняющегося вывода (например, форматтер ошибки). Для компонентов с активной вёрсткой они дают шум: любая правка класса ломает снепшот, и разработчики привыкают бездумно обновлять (`--update`), теряя ценность.

**Good — snapshot на детерминированную структуру данных:**

```typescript
it("formats api error to stable shape", () => {
	expect(formatApiError(raw)).toMatchInlineSnapshot(`{ "message": "...", "requestId": "..." }`);
});
```

**Bad — snapshot всего компонента:**

```typescript
// Ломается при любой правке className/структуры; ревьюер не понимает, баг это или ожидаемо
expect(container).toMatchSnapshot();
```

Предпочитай точечные ассерты (`getByRole`/`getByText`) вместо снепшота всего дерева.

### 5.8. Мок внешних зависимостей (vi.mock)

Тяжёлые/несущественные для теста зависимости мокаются на уровне модуля. Реальный `issue-card.test.tsx` мокает роутер и drag-and-drop:

```typescript
vi.mock("@tanstack/react-router", () => ({
	Link: ({ children, onClick, className }: {
		children: React.ReactNode; onClick?: (e: React.MouseEvent) => void; className?: string;
	}) => (
		<a href="#" onClick={onClick} className={className}>{children}</a>
	),
}));

vi.mock("@atlaskit/pragmatic-drag-and-drop/element/adapter", () => ({
	draggable: vi.fn(() => vi.fn()),
}));
```

**Правило:** мокать только то, что мешает изоляции (роутер, DnD, анимации). Не мокать сам тестируемый компонент.

### 5.9. Чек-лист frontend-теста

- [ ] node_modules слинкованы (`link-testframework-node-modules.mjs`) перед прогоном.
- [ ] Провайдеры — через `renderWithProviders`, свой QueryClient.
- [ ] Запросы by role/label, не CSS-селекторы.
- [ ] `userEvent` с `await`, не сырой `fireEvent` для сложных действий.
- [ ] API замокан (MSW с `onUnhandledRequest: "error"` или `vi.spyOn`) — нет реальной сети.
- [ ] Хуки — `renderHook` + `act`.
- [ ] Глобальные сторы/sessionStorage сброшены в `beforeEach`, `vi.restoreAllMocks()`.
- [ ] Snapshot — только на стабильные структуры, не на всё дерево.
- [ ] `npx tsc --noEmit` и Biome (`npm run lint`) зелёные.

---

## 6. E2E (Playwright + POM)

E2E-тесты (20 spec) проверяют критические сквозные пути в реальном браузере. Живут в `testframework/stages/e2e/domains/{domain}/`, общий код — в `shared/`. Два job: `e2e-smoke` (3 spec без backend) и `e2e-full` (20 spec с backend).

```bash
cd frontend
npm run test:e2e:smoke   # job e2e-smoke (--project=chromium-smoke)
npm run test:e2e         # job e2e-full (--project=chromium-full)
```

### 6.1. Архитектура: BasePage → Page Object → Steps → Spec

Слоистая структура, каждый слой со своей ответственностью (SRP):

```
BasePage (общая навигация)
   ▲
   │ extends
Page Object (локаторы + действия одной страницы)   ← shared/pages/
   ▲
   │ использует
Steps (бизнес-шаги поверх POM)                       ← shared/steps/
   ▲
   │ использует
Spec (сценарий + ассерты)                            ← domains/{domain}/
```

**BasePage** — общий предок (реальный `base.page.ts`):

```typescript
export abstract class BasePage {
	constructor(protected readonly page: Page) {}

	async goto(path: string): Promise<void> {
		await this.page.goto(path);
	}

	protected locator(selector: string): Locator {
		return this.page.locator(selector);
	}
}

// Хелпер этапа Assert (AAA).
export async function expectVisible(loc: Locator, timeout?: number): Promise<void> {
	await expect(loc).toBeVisible(timeout ? { timeout } : undefined);
}
```

**Page Object** — локаторы и действия конкретной страницы (реальный `login.page.ts`):

```typescript
export class LoginPage extends BasePage {
	readonly emailInput = this.page.getByLabel("Email");
	readonly hubTitle = this.page.getByText("Hub").first();
	readonly devPanel = this.page.getByText("Демо-доступ");

	async open(): Promise<void> {
		await this.goto("/login");
		await expect(this.page).toHaveURL(/\/login/);
	}

	async devLoginAs(buttonName: string): Promise<void> {
		await expectVisible(this.devPanel, 15_000);
		await this.page.getByRole("button", { name: buttonName }).click();
		await expect(this.page).toHaveURL(/\/projects/, { timeout: 15_000 });
	}
}
```

**Steps** — бизнес-шаги (реальный `auth.steps.ts`): `AuthSteps.loginAsAdmin`, `loginAsDev`, `logoutFromProfile`. Spec оперирует шагами, не локаторами.

**(2026) Бизнес-шаги оборачивай в нативный `test.step()`.** Шаги видны в HTML-отчёте, trace viewer и UI Mode, а `allure-playwright` конвертирует их в Allure-шаги автоматически — один API покрывает оба отчёта (вместо отдельного `allure.step` в Steps-слое):

```typescript
export class AuthSteps {
	async loginAsAdmin(page: Page): Promise<void> {
		await test.step("Login as admin (seed + dev-login)", async () => {
			const login = await devLogin("admin@hub.local");
			await authenticatePage(page, login);
		});
	}
}
```

### 6.2. Устойчивые локаторы

Приоритет локаторов (как в RTL, отражает пользовательское восприятие):

| Приоритет | Локатор | Пример |
|-----------|---------|--------|
| 1 | `getByRole` | `page.getByRole("button", { name: "Создать проект" })` |
| 2 | `getByLabel` | `page.getByLabel("Email")` |
| 3 | `getByText` | `page.getByText("Демо-доступ")` |
| 4 | `data-testid` / `#id` | `page.locator("#proj-name")` — когда роль/текст неоднозначны |

`data-testid`/`#id` уместны для полей форм без стабильного label и контейнеров (`#main-content`, `#proj-name`, `#proj-key` в реальном `project-lifecycle.spec.ts`). Не завязываться на CSS-классы и порядок в DOM.

**Good — роль + имя:**

```typescript
await page.getByRole("button", { name: "Создать проект" }).click();
await expect(page.getByRole("tab", { name: "Обзор" })).toBeVisible({ timeout: 15_000 });
```

**Bad — CSS-цепочка:**

```typescript
await page.locator("div.form > button.primary:nth-child(3)").click(); // ломается от любой правки вёрстки
```

### 6.3. Ожидания: auto-wait и web-first assertions, без sleep

Playwright сам ждёт актуальности (видимость, кликабельность) в `expect(...).toBeVisible()` и действиях. Никогда не использовать `page.waitForTimeout` (sleep) для синхронизации.

**Good — web-first assertions с явным таймаутом на медленные операции:**

```typescript
await page.getByRole("button", { name: "Создать проект" }).click();
// Ждём навигацию по URL, а не фиксированную паузу.
await expect(page).toHaveURL(/\/projects\/[0-9a-f-]+/, { timeout: 30_000 });
await expect(page.locator("#main-content")).toBeVisible({ timeout: 15_000 });
```

**Bad — sleep:**

```typescript
await page.getByRole("button", { name: "Создать проект" }).click();
await page.waitForTimeout(3000); // флейк: иногда мало, иногда медленно; всегда тратит время
expect(page.url()).toContain("/projects/");
```

Типовые ожидания: `toHaveURL`, `toBeVisible`, `toHaveText`, `toHaveCount`. Для навигации — `toHaveURL(regex)`.

### 6.4. Фикстуры и условный пропуск

E2E-фикстуры (реальный `fixtures.ts`) сообщают, доступен ли backend/dev-mode, и позволяют корректно пропускать тесты, требующие инфраструктуры:

```typescript
export const test = base.extend<HubFixtures>({
	backendAvailable: async ({}, use) => { await use(await isBackendReachable()); },
	devModeAvailable: async ({}, use) => { await use(await isDevModeAvailable()); },
});

export function skipUnlessBackend(backendAvailable: boolean): void {
	test.skip(!backendAvailable, "Backend not running — start with: docker compose ...");
}
```

Smoke-тесты (job `e2e-smoke`) идут без backend, поэтому spec с backend-зависимостью объявляют пропуск в начале:

```typescript
test("create project ...", async ({ page, backendAvailable, devModeAvailable }) => {
	skipUnlessBackend(backendAvailable);
	skipUnlessDevMode(devModeAvailable);
	// ...основной сценарий
});
```

### 6.5. Dev-login (аутентификация в E2E)

E2E не проходит magic-link через email. Вместо этого — dev-login (только при `DEV_MODE=true` на backend). Реальный `helpers/auth.ts` сидит демо-данные, логинится и инъектит сессию в sessionStorage:

```typescript
export async function devLogin(email: string, apiUrl = getApiUrl()): Promise<DevLoginResult> {
	await seedDevData(apiUrl); // POST /api/dev/seed
	const res = await fetch(`${apiUrl}/api/dev/login`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email }),
		signal: AbortSignal.timeout(10_000),
	});
	if (!res.ok) throw new Error(`dev login failed: ${res.status}`);
	return (await res.json()) as DevLoginResult;
}

export async function authenticatePage(page: Page, login: DevLoginResult): Promise<void> {
	await page.addInitScript((payload) => {
		sessionStorage.setItem("hub-auth", JSON.stringify(payload));
	}, { state: { accessToken: login.access_token, isAuthenticated: true, user: login.user }, version: 0 });
}
```

Через `AuthSteps` это выглядит одной строкой в spec:

```typescript
const auth = new AuthSteps(page);
await auth.loginAsAdmin(page); // seed + dev-login + inject session
```

**Важно (TZ-инвариант):** dev-login/DEV_MODE — только для локали и CI e2e-full. В prod `DEV_MODE=true` запрещён (rule 00, deploy без DEV_MODE).

**(2026) `storageState`: логин один раз вместо dev-login в каждом тесте.** Канон Playwright — auth setup-project: логинится один раз, сохраняет состояние в json; остальные projects декларируют его через `use` и стартуют уже залогиненными (экономия секунд на каждый spec):

```typescript
// auth.setup.ts (project "setup")
setup("authenticate", async ({ page }) => {
	const login = await devLogin("admin@hub.local");
	await authenticatePage(page, login);
	await page.goto("/projects");
	await page.context().storageState({ path: "playwright/.auth/admin.json" });
});

// playwright.config.ts
projects: [
	{ name: "setup", testMatch: /auth\.setup\.ts/ },
	{
		name: "chromium-full",
		dependencies: ["setup"],
		use: { storageState: "playwright/.auth/admin.json" },
	},
],
```

Нюанс Hub: сессия хранится в **sessionStorage**, который `storageState` из коробки не сохраняет (только cookies + localStorage) — нужна адаптация (инъекция sessionStorage через `addInitScript` из сохранённого json) или storageState-совместимое хранение для тестов. До адаптации per-test dev-login (выше) остаётся рабочим fallback; тесты, мутирующие сессию (logout), выполняют логин сами.

### 6.6. Изоляция E2E

- **Уникальные данные на прогон:** суффикс из времени, чтобы параллельные worker'ы и повторные прогоны не конфликтовали:

```typescript
const suffix = Date.now().toString(36).toUpperCase().slice(-6);
const projectName = `E2E Project ${suffix}`;
const projectKey = `E${suffix}`.slice(0, 10);
```

- **Каждый тест логинится сам** (в `beforeEach` или в начале теста) — не полагаться на состояние предыдущего.
- **Не шарить созданные сущности** между тестами; чистка — через восстановление/удаление внутри сценария (как в `project-lifecycle`: создать → удалить → восстановить).

### 6.7. Smoke vs full

| Job | Что | Backend | Когда |
|-----|-----|---------|-------|
| `e2e-smoke` | 3 spec: login, a11y, critical-gate | нет | на каждом PR (быстро, дёшево) |
| `e2e-full` | 20 spec: все домены | да (`scripts/ci-e2e-backend.sh`) | полный прогон |

Список e2e-full: `login, critical-gate, navigation, logout, health, admin-audit, admin-ai, admin-metrics, ai-hub, automation, metrics, command-palette, shortcuts, onboarding-profile, notifications, project-lifecycle, messenger, tms, a11y, v4-superapp`.

**Правило:** новый критический user-path → добавить в подходящий домен `e2e-full`; в smoke — только login/a11y/critical-gate (не раздувать, smoke должен быть быстрым).

### 6.8. Шардинг и merge-reports в CI (2026)

Для e2e-full (20 spec, 30 min timeout) канон CI-механики — шардинг по машинам + объединённый отчёт (сейчас `workers: 1` — wall-time можно сократить до ~10–12 мин):

```yaml
# ci.yml: матрица шардов, каждый job гонит свою долю
strategy:
  matrix: { shard: [1/2, 2/2] }
steps:
  - run: npx playwright test --project=chromium-full --shard=${{ matrix.shard }} --reporter=blob
  - uses: actions/upload-artifact@v4
    with: { name: blob-report-${{ strategy.job-index }}, path: blob-report }

# отдельный job после шардов: единый HTML-отчёт
- run: npx playwright merge-reports --reporter=html ./all-blob-reports
```

- Параллелизм внутри одного файла — `test.describe.configure({ mode: "parallel" })` (только для независимых тестов; изоляция данных — раздел 6.6 обязательна).
- **Бюджет CI:** студенческий лимит ~3000 мин/мес — шардинг умножает job-минуты незначительно, но сокращает wall-time; включать по образцу coverage-gate (только master push + label), не на каждый PR.

### 6.9. Чек-лист E2E-теста

- [ ] Логика вынесена в Page Object/Steps, spec читается как сценарий.
- [ ] Локаторы by role/label; `data-testid`/`#id` только при необходимости.
- [ ] Ожидания — web-first assertions (`toHaveURL`, `toBeVisible`), ноль `waitForTimeout`.
- [ ] Backend-зависимые spec вызывают `skipUnlessBackend/DevMode` в начале.
- [ ] Аутентификация через `AuthSteps`/dev-login (цель 2026 — setup-project + `storageState`), не ручной magic-link.
- [ ] (2026) Бизнес-шаги обёрнуты в `test.step()` — видны в HTML/trace и конвертируются в Allure.
- [ ] Уникальные данные на прогон (суффикс времени), тест самодостаточен.
- [ ] Новый критический путь добавлен в e2e-full, smoke не раздут.
- [ ] Прогон стабилен (нет флейков при повторе) — раздел 13.

---

## 7. Тестовая документация с привязкой к автотестам

Это **главное требование** playbook. Hub уже реализовал двустороннюю привязку тест-кейсов к автотестам в `docs/TESTING.md`: каждая строка TC-`<MODULE>`-NN имеет колонку «Автотест», указывающую на Go-функцию или spec. Этот раздел формализует, как поддерживать связь.

### 7.1. Зачем нужна traceability

Traceability отвечает на три вопроса аудита:

1. **Покрыто ли требование тестом?** (requirement → TC)
2. **Автоматизирован ли тест-кейс?** (TC → autotest)
3. **Что проверяет этот автотест с точки зрения бизнеса?** (autotest → TC → requirement)

Без этой цепочки невозможно доказать, что фича протестирована, и невозможно понять, что сломается при изменении кода. Формула:

```
requirement (US/FR)  →  TC-<MODULE>-NN  →  autotest (Go func / spec)
   docs/vision_*          docs/TESTING.md      *_test.go / *.spec.ts
```

### 7.2. Формат таблицы TC в `docs/TESTING.md`

Каждый модуль имеет таблицу тест-кейсов со строго фиксированными колонками:

```markdown
### TRCK — Трекер задач

| ID | Название | Предусловие | Шаги | Ожидаемый результат | Автотест |
|----|----------|-------------|------|---------------------|----------|
| TC-TRCK-03 | Обновление с version check | version=1 | PUT issues/:key {title, version: 1} | 200, version=2 | `TestUpdateIssueOptimisticLocking` |
| TC-TRCK-04 | Конфликт версий | version=2 в БД | PUT issues/:key {title, version: 1} | 409, conflict | `TestUpdateIssueOptimisticLocking` |
```

| Колонка | Правило заполнения |
|---------|--------------------|
| ID | `TC-<MODULE>-<NN>`, MODULE = AUTH/PROJ/TRCK/MSG/TMS/CALL/SPRINT/NOTIF/ADMIN…, NN — двузначный |
| Название | краткая суть сценария (что проверяем) |
| Предусловие | состояние системы до шагов (данные, права, флаги) |
| Шаги | конкретные действия (endpoint + payload или UI-действия) |
| Ожидаемый результат | наблюдаемый исход (код, сущность, событие) |
| Автотест | имя Go-функции / spec-файла в бэктиках, либо `— (ручной)` |

**Значения колонки «Автотест»:**
- `` `TestUpdateIssueOptimisticLocking` `` — конкретная Go-функция;
- `` `logout.spec.ts` `` — E2E-spec;
- «— (ручной)» — сценарий проверяется только руками (пока не автоматизирован);
- «manual / middleware test» — комбинированный (частично авто).

### 7.3. Двусторонняя связь: код ↔ MD

Связь должна быть навигируемой в обе стороны.

**Направление MD → код:** колонка «Автотест» содержит точное имя функции/spec. Поиск `rg "TestUpdateIssueOptimisticLocking" backend/testframework` мгновенно находит реализацию.

**Направление код → MD:** тест ссылается на TC-ID одним из способов (по приоритету для Hub):

1. **Комментарий с TC-ID над функцией** (минимальный, всегда доступен):

```go
// TC-TRCK-03, TC-TRCK-04: обновление с проверкой версии и конфликт версий.
func TestUpdateIssueOptimisticLocking(t *testing.T) {
	// ...
}
```

2. **Allure tms-link / label** (для тестов, чьи Allure-отчёты заливаются в TMS — раздел 8). В Go это метаданные результата; в TS — `allure.tms("HUB-TC-42", url)` или `allure.label("AS_ID", "HUB-TC-99")`.

3. **Имя функции, отражающее сценарий** — `TestSendMessage_ReadOnlyBlocked_RegularUser` явно соответствует TC-MSG-07.

**Good — двусторонняя связь:**

```go
// TC-MSG-07: read-only канал, обычный участник не может отправить сообщение.
func TestSendMessage_ReadOnlyBlocked_RegularUser(t *testing.T) { /* ... */ }
```

В `docs/TESTING.md`:

```markdown
| TC-MSG-07 | Read-only канал | is_readonly=true, не Owner | POST messages | 403 | `TestSendMessage_ReadOnlyBlocked_RegularUser` |
```

**Bad — связь только в одну сторону:**

```markdown
| TC-MSG-07 | Read-only канал | ... | 403 | ✅ автоматизировано |
```

«✅ автоматизировано» без имени функции нельзя навигировать: непонятно, какой именно тест, и жив ли он ещё.

### 7.4. Разделение обязанностей: кто что заполняет

| Шаг | Роль | Действие |
|-----|------|----------|
| 1 | `qa-manual` | Пишет строку TC: ID, название, предусловие, шаги, ожидаемый результат. Колонка «Автотест» = «— (ручной)» |
| 2 | `qa-auto1/2` | Автоматизирует TC → пишет `_test.go`/`.spec.ts`, добавляет комментарий с TC-ID |
| 3 | `qa-auto1/2` | Обновляет колонку «Автотест» именем функции/spec — в той же сессии (rule 55 docs-sync) |
| 4 | `qa-*-ver` | Проверяет: имя в колонке существует в коде; тест реально проверяет описанный сценарий |

**Правило Hub:** нельзя закрыть авто-задачу, не обновив колонку «Автотест». Код без обновления `docs/TESTING.md` = незавершённая работа (rule 55).

### 7.5. Поддержание статистики автоматизации

`docs/TESTING.md` содержит сводку автоматизации по модулям:

```markdown
| Модуль | TC всего | Автоматизировано | Только ручные | % автоматизации |
|--------|----------|-----------------|---------------|-----------------|
| MSG | 10 | 8 | 2 | 80% |
| TMS | 12 | 10 | 2 | 83% |
```

При добавлении автотеста к ранее ручному TC — обновить и строку TC, и эту сводку. Верификатор сверяет: число «Автоматизировано» = число строк с непустой колонкой «Автотест».

### 7.6. Матрица трассируемости requirement → TC → autotest

Полная трассируемость сводится в `.planning/COVERAGE.md` (создаётся `/coverage-audit`). Формат строки:

```markdown
| ID | Requirement | Code Location | TC | Tests | Status |
|----|-------------|---------------|----|-------|--------|
| AUTH-01 | Magic Link login | auth/magiclink.go | TC-AUTH-01, TC-AUTH-02 | `TestVerifyMagicLink_*` | ✅ |
```

Статусы: `✅ Implemented + Tested`, `⚠️ Implemented, No Tests`, `🔴 Not Implemented`.

Hub также ведёт трассируемость на уровне TMS-платформы: тест-кейсы связываются с задачами (`TestLinkTestCaseToIssue`), а отчёт traceability доступен через `GET reports/traceability` (TC-TMS-11).

### 7.7. Пример полного цикла привязки

Задача: автоматизировать ручной TC-TRCK-12 (мягкое удаление задачи).

1. **Исходное состояние в `docs/TESTING.md`:**

```markdown
| TC-TRCK-12 | Мягкое удаление | Задача существует | DELETE issues/:key | 200, deleted_at set | — (ручной) |
```

2. **`qa-auto1` пишет тест с комментарием TC-ID:**

```go
// TC-TRCK-12: мягкое удаление задачи выставляет deleted_at и убирает из выдачи.
func TestDeleteIssue_SoftDelete(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	created, err := svc.CreateIssue(ctx, uuid.New(), uuid.New(), tracker.CreateIssueRequest{
		Title: "To delete", IssueType: "task", Priority: "low",
	})
	require.NoError(t, err)

	err = svc.DeleteIssue(ctx, created.IssueKey, uuid.New())
	require.NoError(t, err)

	_, err = svc.GetIssue(ctx, created.IssueKey)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 404, appErr.Code) // после soft delete не доступна
}
```

3. **`qa-auto1` обновляет колонку и сводку:**

```markdown
| TC-TRCK-12 | Мягкое удаление | Задача существует | DELETE issues/:key | 200, deleted_at set | `TestDeleteIssue_SoftDelete` |
```

И `TRCK: 7 → 8 автоматизировано`.

4. **`qa-auto-ver` проверяет:** `rg TestDeleteIssue_SoftDelete backend/testframework` находит функцию; тест реально проверяет 404 после удаления; прогон зелёный.

### 7.8. Чек-лист traceability

- [ ] Каждый новый сценарий имеет строку TC-`<MODULE>`-NN в `docs/TESTING.md`.
- [ ] Автотест содержит комментарий/аннотацию с TC-ID (код → MD).
- [ ] Колонка «Автотест» содержит точное имя функции/spec в бэктиках (MD → код).
- [ ] Имя из колонки реально существует в коде (проверяемо через `rg`).
- [ ] Сводка «% автоматизации» модуля обновлена.
- [ ] Для критичных требований строка есть в `.planning/COVERAGE.md`.
- [ ] Ручные TC явно помечены «— (ручной)», не «✅ автоматизировано» без имени.

---

## 8. Allure в деталях

В Hub Allure присутствует в **двух смыслах**, и важно их не путать:

- **(а) Allure как формат отчётов автотестов**, который команда генерирует у себя и заливает в TMS-модуль Hub. Hub **сам парсит** Allure/JUnit-отчёты (`backend/internal/tms`, `allure_parser`, `allure_matcher`) и превращает их в тест-раны и результаты.
- **(б) Практики Allure-аннотаций** (epic/feature/story/severity/step/attachment/tms-link) в собственных автотестах команды.

### 8.1. Как Hub парсит и матчит Allure (реализовано в TMS)

Платформа принимает Allure-результаты двумя путями:

| Путь | Endpoint | Auth | TC |
|------|----------|------|----|
| Ручной импорт | `POST /projects/:id/tms/allure/import` (multipart) | JWT + RBAC | TC-TMS-07 |
| Webhook из CI | `POST /projects/:id/tms/webhook/allure` | API Key | TC-TMS-07/08 |

Парсер (`tms.ParseAllureResults`) разбирает массив Allure-результатов. Реальный формат результата (из `allure_parser_test.go`):

```json
[
  {
    "uuid": "abc-123",
    "historyId": "hist-1",
    "fullName": "com.example.LoginTest.testSuccess",
    "name": "testSuccess",
    "status": "passed",
    "statusDetails": { "message": "", "trace": "" },
    "start": 1700000000000,
    "stop": 1700000001000,
    "steps": [ { "name": "Open page", "status": "passed", "steps": [] } ],
    "labels": [
      { "name": "suite", "value": "LoginTest" },
      { "name": "severity", "value": "critical" }
    ],
    "links": [ { "name": "HUB-TC-42", "url": "https://hub.test/tc/42", "type": "tms" } ],
    "attachments": [ { "name": "screenshot.png", "source": "abc.png", "type": "image/png" } ]
  }
]
```

Парсер устойчив к неполным данным (реальные тесты): пустой массив → пусто без ошибки; отсутствующие поля → nil; malformed JSON → ошибка; null-таймстемпы → nil-указатели.

### 8.2. Стратегии матчинга (matching) тест-кейсов

`tms.Matcher.MatchTestCase` привязывает Allure-результат к существующему тест-кейсу Hub. Стратегии (реальные из `allure_parser_test.go`), в порядке приоритета:

| Приоритет | Стратегия | Константа | Источник в Allure | Метод репозитория |
|-----------|-----------|-----------|-------------------|-------------------|
| 1 | По аннотации/ссылке TMS | `MatchByAnnotation` | `links[].type == "tms"` (name = TC key), либо label `AS_ID` | `FindTestCaseByExternalID` |
| 2 | По automation ID | `MatchByAutomationID` | automation-идентификатор | `FindTestCaseByAutomationID` |
| 3 | По case key | `MatchByCaseKey` | ключ кейса | `FindTestCaseByKey` |
| 4 | По полному имени | `MatchByFullName` | `fullName` | `FindTestCaseByFullName` |
| 5 | По классу/методу | `MatchByClassName` | `suite` label + `name` | `FindTestCaseByClassMethod` |
| — | Нет совпадения | `MatchStrategy(-1)` → `"unmapped"` | — | — |

**Приоритет важен:** если есть и tms-link, и fullName — побеждает аннотация (реальный `TestMatchTestCase_PriorityOrder`):

```go
func TestMatchTestCase_PriorityOrder(t *testing.T) {
	repo := newMockMatcherRepo()
	annotationTC := makeMockTestCase(uuid.New(), "HUB-TC-10")
	fullNameTC := makeMockTestCase(uuid.New(), "HUB-TC-20")
	repo.externalIDMap["HUB-TC-10"] = annotationTC
	repo.fullNameMap["com.example.PrioTest.testPrio"] = fullNameTC

	matcher := tms.NewMatcher(repo)
	result := tms.AllureResult{
		Name:     "testPrio",
		Status:   "passed",
		FullName: "com.example.PrioTest.testPrio",
		Links:    []tms.AllureLink{{Name: "HUB-TC-10", Type: "tms"}},
	}
	matched, strategy, err := matcher.MatchTestCase(context.Background(), uuid.New(), &result)
	require.NoError(t, err)
	assert.Equal(t, annotationTC.ID, matched.ID)      // аннотация победила
	assert.Equal(t, tms.MatchByAnnotation, strategy)
}
```

**Вывод для команды:** чтобы результат гарантированно матчился на нужный TC, используй **явную tms-link или AS_ID label с ключом кейса Hub** (`HUB-TC-42`). Полагаться на fullName/className — хрупко (переименование метода рвёт связь).

### 8.3. Статусы Allure → TMS

`tms.AllureStatusToTMS` маппит статусы (реальный `TestAllureStatusToTMS`):

| Allure | TMS |
|--------|-----|
| `passed` | `passed` |
| `failed` | `failed` |
| `broken` | `broken` |
| `skipped` | `skipped` |
| `""` / неизвестный | `unknown` |

`passed` — тест прошёл; `failed` — упал ассерт; `broken` — упал не на ассерте (исключение/инфра); `skipped` — пропущен. Различие failed/broken важно для triage (broken часто = проблема окружения, не кода).

### 8.4. (б) Аннотации в собственных автотестах — TypeScript (allure-playwright)

Для Playwright-тестов, чьи отчёты заливаются в TMS Hub, добавляем Allure-метаданные. Ключевое — **tms-link/label с ключом кейса Hub** для надёжного матчинга.

```typescript
import { test } from "../../shared/fixtures";
import * as allure from "allure-js-commons";

test("create project lifecycle", async ({ page }) => {
	// Иерархия для навигации в отчёте.
	await allure.epic("Tracker");
	await allure.feature("Project lifecycle");
	await allure.story("Create and restore project");
	await allure.severity("critical");

	// Привязка к тест-кейсу Hub (matching by annotation — высший приоритет).
	await allure.label("AS_ID", "HUB-TC-PROJ-01");
	// либо явная tms-ссылка:
	await allure.link("https://hub.local/tms/HUB-TC-PROJ-01", "HUB-TC-PROJ-01", "tms");

	await allure.step("Login as admin", async () => {
		// действия шага; каждый step попадает в steps[] отчёта
	});

	await allure.step("Create project", async () => {
		// ...
	});

	// (2026) Нативный test.step() — предпочтительный способ в Steps-слое POM:
	// allure-playwright конвертирует его в Allure-шаг автоматически, и шаг виден
	// ещё и в HTML-отчёте/trace viewer Playwright. Отдельный allure.step нужен
	// только для чисто отчётных шагов вне page-действий.

	// Attachment (скриншот/лог) — прикрепляется к результату.
	await allure.attachment("page.png", await page.screenshot(), "image/png");
});
```

Соответствие аннотаций формату, который парсит Hub:

| allure-js вызов | Поле в Allure JSON | Использование в Hub |
|-----------------|--------------------|--------------------|
| `allure.label("AS_ID", "HUB-TC-99")` | `labels[] {name:"AS_ID"}` | `MatchByAnnotation` |
| `allure.link(url, name, "tms")` | `links[] {type:"tms"}` | `MatchByAnnotation` |
| `allure.epic/feature/story` | `labels[]` | иерархия отчёта |
| `allure.severity("critical")` | `labels[] {name:"severity"}` | приоритизация |
| `allure.step(...)` | `steps[]` | детализация в отчёте |
| `allure.attachment(...)` | `attachments[]` | скриншоты/логи |

### 8.5. (б) Метаданные в Go-автотестах

Go-тесты Hub — это `testing` + testify, без нативного Allure-раннера. Метаданные для TMS передаются иначе:

1. **Комментарий с TC-ID** над функцией (см. 7.3) — минимальный обязательный уровень.
2. **Генерация Allure-результата из Go** (если нужно залить в TMS): формируется JSON в формате `tms.AllureResult` — с `fullName`, `labels` (включая `AS_ID`), `links` (`type: "tms"`), `status`, `steps`, `attachments`. Ключевое поле для матчинга — `AS_ID` label или tms-link с ключом кейса Hub.

Пример структуры результата, пригодного для заливки (соответствует тому, что парсит `tms.ParseAllureResults`):

```go
// Формируем Allure-результат для заливки в TMS Hub (matching by AS_ID).
result := tms.AllureResult{
	UUID:     uuid.New().String(),
	Name:     "TestUpdateIssueOptimisticLocking",
	FullName: "tracker.TestUpdateIssueOptimisticLocking",
	Status:   "passed",
	Labels: []tms.AllureLabel{
		{Name: "AS_ID", Value: "HUB-TC-TRCK-03"}, // привязка к кейсу Hub
		{Name: "suite", Value: "tracker"},
		{Name: "severity", Value: "critical"},
	},
}
```

### 8.6. Генерация allure-results (клиентская сторона)

Прежде чем заливать в TMS Hub, команда генерирует `allure-results/` локально/в CI:

```bash
# Playwright: reporter allure-playwright пишет allure-results/
npx playwright test --reporter=line,allure-playwright
# результат: allure-results/*.json (+ attachments)

# Просмотр локально (не заливка в TMS, а человекочитаемый отчёт):
npx allure generate allure-results --clean -o allure-report
npx allure open allure-report
```

`allure-results/*.json` — это тот самый массив, который принимает `POST /tms/allure/import`.

**(2026) Allure Report 3:** ставится как `npm i -D allure` → команды через `npx allure`; `allure-playwright`/`allure-vitest` совместимы и с Report 2, и с 3 — формат results не меняется, миграция просмотра безболезненна.

**(2026) allure-vitest для frontend-unit.** Сейчас в TMS льются только e2e/Go — подключаем reporter `allure-vitest/reporter`, чтобы и Vitest-результаты попадали в TMS (осторожно с `describe.concurrent` в browser mode). Для `AS_ID`-привязки используем общую wrapper-функцию вместо ручного `allure.label` в каждом тесте (рекомендация Qameta):

```typescript
// testframework/.../allure-helpers.ts — единая точка привязки TC
export function tcTest(tcId: string, name: string, fn: TestFunction): void {
	test(name, async (ctx) => {
		await allure.label("AS_ID", tcId);
		await fn(ctx);
	});
}
```

### 8.7. Заливка в TMS Hub

**Ручной импорт (multipart):**

```bash
curl -sf -X POST "$HUB/api/projects/$PROJECT_ID/tms/allure/import" \
  -H "Authorization: Bearer $JWT" \
  -F "files=@allure-results/result-1.json" \
  -F "files=@allure-results/result-2.json"
# 200 → создан тест-ран, результаты сматчены на TC (по стратегиям 8.2)
```

**Webhook из CI (API Key):**

```bash
curl -sf -X POST "$HUB/api/projects/$PROJECT_ID/tms/webhook/allure" \
  -H "X-API-Key: $HUB_API_KEY" \
  -H "Content-Type: application/json" \
  --data-binary @allure-results/combined.json
```

**Важно (TZ-инвариант SEC-01/SEC-02):** webhook требует API Key; подпись/ключ нельзя пропускать. Тесты вебхука проверяют отказ без ключа (`handler_webhook_rbac_test.go`).

### 8.8. Чтение отчётов: flaky, trends, coverage, traceability

После заливки Hub строит отчёты (TMS):

| Отчёт | Endpoint / TC | Что показывает |
|-------|---------------|----------------|
| Flaky | `GET reports/flaky` (TC-TMS-10) | тесты с нестабильным результатом между прогонами |
| Coverage | `TestGetCoverageReport` | доля кейсов, покрытых прогонами |
| Traceability | `GET reports/traceability` (TC-TMS-11) | матрица кейс↔задача |
| Trends | по историям прогонов | динамика pass-rate |

**Bug from failed:** из failed-результата создаётся Bug-задача (`TestCreateBugFromFailedResult`, TC-TMS-09) — замыкает цикл «упавший автотест → задача в трекере».

### 8.9. Good / bad аннотирования

**Good — явная привязка к TC + осмысленная иерархия:**

```typescript
await allure.epic("Auth");
await allure.feature("Magic Link");
await allure.label("AS_ID", "HUB-TC-AUTH-02"); // надёжный матчинг
await allure.severity("critical");
await allure.step("Request code", async () => { /* ... */ });
```

**Bad — без привязки, матчинг «на удачу» по имени:**

```typescript
test("test login works", async ({ page }) => {
	// нет epic/feature/AS_ID; матчинг попытается по fullName/className,
	// переименуешь тест — связь с TC порвётся молча
});
```

**Bad — attachment с секретами:**

```typescript
// НЕЛЬЗЯ: JWT/пароли в attachment попадут в TMS-отчёт и логи
await allure.attachment("token.txt", jwtAccessToken, "text/plain");
```

### 8.10. Чек-лист Allure

- [ ] Автотест, заливаемый в TMS, имеет `AS_ID` label или tms-link с ключом кейса Hub (`HUB-TC-*`); в Vitest — через общий wrapper (2026).
- [ ] Заданы epic/feature/story/severity для навигации в отчёте.
- [ ] Ключевые действия обёрнуты в шаги: `test.step()` в Playwright (2026), `allure.step` — для чисто отчётных шагов.
- [ ] Attachments не содержат секретов (токены, пароли, PII).
- [ ] `allure-results/` генерируется в CI перед заливкой.
- [ ] Импорт/webhook возвращает 200; результаты сматчены (strategy ≠ `unmapped`).
- [ ] Webhook использует API Key (не пропускать — SEC-01/SEC-02).
- [ ] Различаем failed (баг ассерта) и broken (инфра) при triage.
- [ ] Из failed при необходимости заведён Bug (TC-TMS-09).

---

## 9. Замер времени прогонов и производительность

Это **явное требование**: мы измеряем длительность тестов, ловим медленные, укладываемся в бюджет CI (студенческий лимит GitHub Actions ~3000 мин/мес — экономить) и проверяем целевые метрики производительности продукта.

### 9.1. Зачем мерить время тестов

- **Экономия CI-минут.** Лимит ~3000 мин/мес на всё (backend-test, frontend-test, e2e, coverage). Медленный тест умножается на число прогонов (каждый PR, push).
- **Детект деградации.** Резкий рост длительности = регрессия производительности или новый неэффективный тест (реальная БД вместо мока, sleep).
- **Обратная связь разработчику.** Быстрый локальный цикл (`go test` домена < 10 c) держит фокус.

### 9.2. Тайминги Go-тестов

**`-v` печатает длительность каждого теста:**

```bash
go test -tags hubtest -v ./testframework/tracker/... 2>&1 | rg -- "--- (PASS|FAIL)"
# --- PASS: TestCreateIssue (0.00s)
# --- PASS: TestListIssuesPagination (0.01s)
```

**`-json` — машиночитаемо, для агрегации медленных:**

```bash
go test -tags hubtest -json ./testframework/... > "$TMPDIR/gotest.json"
# извлечь топ медленных (Action=="pass", поле Elapsed):
rg '"Action":"pass"' "$TMPDIR/gotest.json" | \
  node -e 'const fs=require("fs");const lines=fs.readFileSync(0,"utf8").trim().split("\n").map(JSON.parse).filter(x=>x.Test&&x.Elapsed).sort((a,b)=>b.Elapsed-a.Elapsed).slice(0,15);lines.forEach(x=>console.log(x.Elapsed.toFixed(2)+"s  "+x.Package.split("/").pop()+"."+x.Test))'
```

**Общее wall-time прогона** — фиксируется в skill-логе и `.planning/COVERAGE.md`:

```bash
# PowerShell (Windows, среда пользователя):
Measure-Command { go test -tags hubtest -race -count=1 ./testframework/... } | Select-Object TotalSeconds
```

### 9.3. Бенчмарки (`testing.B`, `b.ReportMetric`)

Для горячих участков (парсинг Allure на 1000 результатов, JQL, merge поиска) пишем бенчмарк:

```go
func BenchmarkParseAllureResults(b *testing.B) {
	data := []byte(largeAllureJSON) // ~1000 результатов
	b.ReportAllocs()               // отчёт по аллокациям
	b.ResetTimer()                 // не учитывать подготовку данных

	for i := 0; i < b.N; i++ {
		results, err := tms.ParseAllureResults(data)
		if err != nil {
			b.Fatal(err)
		}
		_ = results
	}

	// Кастомная метрика: результатов в секунду.
	b.ReportMetric(float64(1000*b.N)/b.Elapsed().Seconds(), "results/s")
}
```

Запуск:

```bash
go test -tags hubtest -run '^$' -bench BenchmarkParseAllureResults -benchmem ./testframework/tms/...
# BenchmarkParseAllureResults-8   1200   950000 ns/op   ... B/op   ... allocs/op   1052631 results/s
```

**Правило:** бенчмарки не входят в обычный прогон (`-run '^$'` отключает тесты). Их гоняем точечно при работе над производительностью, результат фиксируем в заметке, а не в CI (CI-минуты дороги).

### 9.4. Целевые метрики продукта (из `docs/TESTING.md`)

Performance-тесты проверяют бюджеты продукта:

| Метрика | Целевое значение |
|---------|-----------------|
| API response time (p95) | < 200 ms |
| WebSocket message latency | < 100 ms |
| Kanban board load (100 задач) | < 500 ms |
| Messenger: загрузка канала (1000 сообщений) | < 300 ms |
| Глобальный поиск | < 500 ms |
| Allure import (1000 результатов) | < 5 s |
| Одновременных WS-подключений | 100+ |
| Участников в звонке | до 50 |

Эти метрики измеряются нагрузочными инструментами (9.6), а не unit-тестами. В unit-бенчмарках проверяем локальную скорость критичных функций (парсер, JQL).

### 9.5. Тайминги Playwright

Playwright-репортер печатает длительность каждого теста; `--reporter=list` показывает время. Для анализа:

```bash
npx playwright test --reporter=json > "$TMPDIR/pw.json"
# durations в поле results[].duration (мс)
```

Медленные E2E — главный пожиратель CI-минут. Правила экономии:
- smoke (3 spec) без backend — на каждом PR;
- full (20 spec) с backend — реже/по необходимости;
- не дублировать unit-проверки в E2E;
- переиспользовать сессию через dev-login (быстрее реального magic-link).

### 9.6. Нагрузочное тестирование (k6 / vegeta)

Для проверки p95 и лимитов — отдельные нагрузочные прогоны (не в основном CI, по требованию). Пример k6 для API p95 < 200 ms:

```javascript
import http from "k6/http";
import { check } from "k6";

export const options = {
	vus: 50,
	duration: "30s",
	thresholds: {
		http_req_duration: ["p(95)<200"], // целевой бюджет Hub
	},
};

export default function () {
	const res = http.get(`${__ENV.HUB}/api/projects`, {
		headers: { Authorization: `Bearer ${__ENV.JWT}` },
	});
	check(res, { "status 200": (r) => r.status === 200 });
}
```

`vegeta` для быстрой проверки rate limit (100/сек → 101-й 429):

```bash
echo "GET $HUB/api/projects" | vegeta attack -rate=101/1s -duration=1s \
  -header "Authorization: Bearer $JWT" | vegeta report
```

**Правило:** нагрузочные тесты запускаются локально/в отдельном окружении, их результаты фиксируются в отчёте. Не гонять их в обычном CI — они дорогие и требуют развёрнутой инфры.

### 9.7. Детект и лечение медленных тестов

| Причина медленного теста | Признак | Лечение |
|--------------------------|---------|---------|
| Реальная БД вместо мока | тест > 100 ms на unit-логике | заменить на `mockRepo`, если реальная схема не нужна |
| `time.Sleep` / `waitForTimeout` | фиксированная пауза | заменить на fake clock / web-first assertion |
| testcontainers в unit | старт контейнера секунды | использовать sqlite `env.Builder` или мок |
| Избыточный E2E | логин+создание ради проверки строки | перенести проверку в unit |
| Нет `b.ResetTimer` в бенчмарке | подготовка учтена в замере | `b.ResetTimer()` после setup |

### 9.8. Фиксация времени в отчётах

После значимого прогона фиксируем wall-time:
- в **skill-логе** сессии (кратко: «backend-test full: 210s, топ-медленный TestX 4.2s»);
- в **`.planning/COVERAGE.md`** — при coverage-задачах (общее время + coverage);
- при деградации > 20% — заметка в `.planning/STATE.md` и, при необходимости, задача на оптимизацию.

### 9.9. Чек-лист производительности

- [ ] Знаю wall-time прогона своего слоя, зафиксировал при значимом изменении.
- [ ] Нет `time.Sleep`/`waitForTimeout` для синхронизации.
- [ ] Unit-тесты используют мок/sqlite, не поднимают контейнеры без нужды.
- [ ] Бенчмарки с `b.ResetTimer`/`b.ReportAllocs`, вне обычного CI.
- [ ] Медленные тесты (топ по `-json`) отревьюены и оправданы.
- [ ] Нагрузочные проверки (k6/vegeta) — вне основного CI, результат зафиксирован.
- [ ] Бюджет CI-минут учтён: smoke на PR, full реже.

---

## 10. Логирование в тестах и артефакты

### 10.1. Что логировать в тестах

**Правило:** тест должен при падении сам объяснять причину через сообщения ассертов, а не через россыпь `fmt.Println`/`t.Log`.

**Good — информативные ассерты:**

```go
assert.Equal(t, 409, appErr.Code, "обновление со старой версией должно давать conflict")
require.NoError(t, err, "создание задачи с валидными данными не должно падать")
```

**Bad — отладочный шум:**

```go
fmt.Println("resp:", resp)        // засоряет вывод CI, не помогает при падении
t.Log("got here")                 // бессмысленно
```

`t.Logf` уместен для контекста, который виден только при `-v`/падении:

```go
t.Logf("issue_key=%s version=%d", resp.IssueKey, resp.Version)
```

**Frontend:** не оставлять `console.log` в тестах (Biome это ловит). Для отладки — `screen.debug()` временно, удалить перед коммитом.

### 10.2. Артефакты Playwright: скриншоты, трейсы, видео

Playwright-конфиг Hub собирает диагностику при падении. Практика:
- **trace** — `on-first-retry` (трейс только при ретрае, экономия);
- **screenshot** — `only-on-failure`;
- **video** — `retain-on-failure`.

Просмотр трейса упавшего теста:

```bash
npx playwright show-trace test-results/<...>/trace.zip
```

Эти артефакты — в `test-results/` (в `.gitignore`), в CI выгружаются как job-артефакт. **Не коммитить** их в репозиторий.

### 10.3. Allure attachments

Attachments (скриншоты, логи, HTTP-дампы) прикрепляются к результату и попадают в TMS-отчёт (раздел 8). Правила:
- прикреплять только релевантное для диагностики;
- **никогда** не прикреплять секреты (JWT, пароли, API-ключи, PII);
- бинарь (png) — с корректным MIME.

### 10.4. Куда писать coverage-профили (rule 45)

**Только в TMP, никогда в `backend/` или корень:**

```bash
# Good
go test -tags hubtest -coverprofile="${TMPDIR:-/tmp}/hub-cov.out" ./testframework/...

# Bad — попадёт в git status, в .gitignore, но мусорит
go test -tags hubtest -coverprofile=coverage.out ./...
```

### 10.5. Cleanup артефактов (rule 45 — обязательно)

После coverage/race/bench-прогонов удалить локальные дампы **в той же сессии**, до хендоффа. Паттерны для удаления:

| Категория | Паттерн |
|-----------|---------|
| Go cover profiles | `*.out` в `backend/` |
| Text cover dumps | `backend/cover*`, `backend/cov_*` |
| Per-package reports | `backend/*_cov*` |
| Temp bench runs | `backend/tmp*` |
| Scratch scripts | `tmp_*.py` в корне |

Cleanup one-liner (PowerShell, из корня — среда пользователя):

```powershell
Get-ChildItem backend -File | Where-Object {
  $_.Name -match '^(cover|coverage|cov_|tmp|.*_cov)' -or $_.Extension -eq '.out'
} | Remove-Item -Force
Remove-Item -Force tmp_*.py -ErrorAction SilentlyContinue
```

После cleanup — `git status`: в diff не должно быть `cover*`, `cov_*`, `*.out`, `tmp_*`. Playwright `test-results/`, `allure-results/`, `allure-report/` — тоже не коммитить.

### 10.6. Чек-лист логирования и артефактов

- [ ] Ассерты с осмысленными сообщениями; нет `fmt.Println`/`console.log`.
- [ ] `t.Logf`/`screen.debug()` — только временно, удалены перед сдачей.
- [ ] Playwright trace/screenshot/video — on-failure, в `test-results/` (не в git).
- [ ] Allure attachments без секретов, корректный MIME.
- [ ] Coverage-профили в TMP.
- [ ] Артефакты удалены (rule 45), `git status` чист.

---

## 11. Регрессия и qa-gate

### 11.1. Обязательный цикл (rule 56) — качество всегда циклично

Проверка **новой фичи** и **старого функционала** — не one-shot, а циклы до green:

```
Новое поведение:
  implement → unit/TC/сценарий → FAIL? → triage → fix → повтор до green

Старый функционал (регрессия):
  домен / смежное → FAIL? → triage → fix → повтор до green
  интеграция волны → /qa-gate (цикл при FAIL)
```

```
1. Изменил код
2. Цикл качества НОВОГО поведения — до green
3. Цикл регрессии СТАРОГО (затронутый модуль) — до green
4. Полный /qa-gate перед handoff (волна: team-lead; одиночная работа: исполнитель)
5. Docs sync (rule 55) — включая колонку «Автотест»
6. *-ver / verifier для критичных изменений
7. Обновить .planning/STATE.md
```

`verified: true` ставится **только** при green обоих циклов + приложенный лог.
`team-lead` обязан включать оба контура в шаблон делегирования (`team-workflow.md` §5.3.1).

### 11.2. Минимальная регрессия по слою

Быстрая проверка затронутого слоя — до полного gate:

| Слой | Команда |
|------|---------|
| Backend | `cd backend && go build -tags release ./... && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...` |
| Один домен BE | `cd backend && go test -tags hubtest -race ./testframework/<domain>/...` |
| Frontend | `cd frontend && npx tsc --noEmit && npx vitest run` |
| Один домен FE | `npx vitest run ../testframework/stages/unit/frontend-test/domains/<domain>` |
| E2E smoke | `cd frontend && npm run test:e2e:smoke` |
| Infra | проверить diff в `ci.yml`, `deploy.yml`, `docker-compose.prod.yml` |

### 11.3. Полный qa-gate (порядок, стоп на первой FAIL)

`/qa-gate` (skill) выполняет проверки последовательно и **останавливается на первой FAIL**:

```bash
# Step 1: Go build + vet
cd backend && go build -tags release ./... && go vet -tags hubtest ./...

# Step 2: Go tests (race)
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Step 3: TypeScript
cd frontend && npx tsc --noEmit

# Step 4: Frontend lint (Biome gate)
npm run lint

# Step 5: Frontend tests
npx vitest run

# Step 6: Security scan (секреты в staged)
git diff --staged | rg -E "DEV_MODE=true|devkey: secret|livekit_api_secret|password|api_key"

# Step 7: API smoke (если сервер поднят)
curl -sf http://localhost:3000/api/health | jq .status   # ожидаем "ok"
```

**Важно:** при FAIL на шаге тестов — **не чинить сразу**. Сначала triage (раздел 12): воспроизвести функционал по спеке, классифицировать ROOT=code|test.

### 11.4. Порядок при большом изменении

1. Затронутый домен (минимальная регрессия) — быстрый цикл правок.
2. Смежные домены (если менялся общий код: env.Builder, shared/, платформенные пакеты).
3. Полный `/qa-gate`.
4. E2E full — если менялся критический user-path.
5. Coverage-gate — при push в master / coverage-задачах.

### 11.5. Результат gate

```
GATE PASSED — ready to commit
GATE FAILED at step N: <причина>
```

Лог прикладывается к хендоффу верификатору (rule 56: `verified: true` требует лог).

### 11.6. Чек-лист регрессии

- [ ] Цикл качества **нового** поведения зелёный (не one-shot).
- [ ] Цикл регрессии **старого** (затронутый слой) зелёный.
- [ ] Полный `/qa-gate` пройден, лог сохранён.
- [ ] При FAIL применён triage (раздел 12), а не слепой фикс; после фикса — повтор цикла.
- [ ] Docs sync выполнен (STATE + domain doc + колонка «Автотест»).
- [ ] Для security/auth/deploy вызван `security-auditor`.
- [ ] `.planning/STATE.md` обновлён.

---

## 12. Test triage

**Обязательный протокол** (REGULATION §10, skill `test-triage`). Принцип Hub: **тест — это гипотеза о поведении. Падающий тест не означает автоматически баг в продакшн-коде.**

### 12.1. Когда применять

- `go test` / `vitest` / Playwright FAIL;
- регрессия после рефакторинга;
- агент собирается «починить» тест, чтобы стал зелёным;
- пользователь сообщает о баге, а тесты зелёные.

### 12.2. Порядок (строго)

**Шаг 1 — зафиксировать симптом:**

```
Тест:   TestBanUser_EmptyReason
Пакет:  testframework/audit
Ошибка: expected error, got nil
```

**Шаг 2 — воспроизвести ФУНКЦИОНАЛ (не тест):**
- прочитать требование в `docs/system_specification.md` / US / PLAN.md;
- трассировать production path: handler → service → repository;
- ответить на вопрос: **«Как ДОЛЖНО работать по спецификации?»**

**Шаг 3 — классификация:**

| Ситуация | Действие |
|----------|----------|
| Код нарушает спецификацию | **Fix code** → обновить тесты, если контракт уточнился |
| Код соответствует спеке, тест устарел/неверен | **Fix test** (mock, assertion, fixture, timing) |
| Спецификация неясна | **Stop** → уточнить у SA/BA, обновить docs, потом код/тест |
| Flaky (race, timing, network) | стабилизировать тест (детерминированный мок, `-race`, fake clock) |

**Шаг 4 — запрещено:**
- менять assertion, чтобы «прошло», не выполнив шаг 2;
- удалять тест без обоснования в skill-логе;
- `t.Skip()` без ticket/плана;
- чинить прод-код только потому, что «тест так ожидает», если тест противоречит docs.

**Шаг 5 — после фикса:**

```bash
go test -tags hubtest -race -count=1 ./testframework/<pkg>/...
# перед handoff:
/qa-gate
```

### 12.3. Запись результата (обязательно)

В skill-лог записывается корень проблемы:

```
TRIAGE: TestBanUser_EmptyReason — ROOT=code (BanUser не проверял reason)
FIX: internal/audit/service.go:142 (добавлена валидация reason)
VERIFIED: manual repro + go test green
```

`ROOT=code` — баг в продакшене; `ROOT=test` — устаревший/неверный тест. Без явного ROOT triage не считается выполненным.

### 12.4. Примеры Hub

**Пример A — чинить ТЕСТ.** `TestRefreshRedisKey_UsesHashNotPlaintext` падает: тест ожидал plaintext-ключ в Redis, но код правильно хеширует → `ROOT=test`, обновить ожидание.

**Пример B — чинить КОД.** `TestBanUser_EmptyReason` падает с `got nil`: спецификация требует reason → `ROOT=code`, добавить validation, тест оставить.

**Пример C — чинить ОБА.** GDPR delete: тест ожидал 200 при partial failure, спека требует error → `ROOT=code` (`errors.Join`) + обновить тест на `want err`.

### 12.5. Запрет «зелёных вслепую»

**Bad — подгонка под текущий вывод:**

```go
// Было: assert.Equal(t, 200, code)  — падает, вернулось 409
assert.Equal(t, 409, code) // «починил», не проверив, что 409 корректен по спеке
```

Если спека говорит, что должно быть 200 (например, идемпотентный retry) — это баг в коде, а не в тесте. Менять assertion без шага 2 — фиксация бага как «правильного» поведения.

**Good — сначала спека:**

```
1. Спека (system_specification.md): повторный retry идемпотентен → 200
2. Код вернул 409 → ROOT=code
3. Fix: сервис должен возвращать существующую сущность при дубле
4. Тест остаётся assert.Equal(t, 200, code)
```

### 12.6. Связь с GSD

- Сложный баг → `gsd-debug` (научный метод, `.planning/debug/`).
- Массовые падения после аудита → `gsd-audit-fix --dry-run` сначала.
- Новые тесты после фазы → `gsd-add-tests`.

### 12.7. Чек-лист triage

- [ ] Зафиксирован симптом (тест, пакет, ошибка).
- [ ] Воспроизведён функционал по спеке (не только assert).
- [ ] Классифицирован ROOT=code | test | unclear | flaky.
- [ ] Починен минимальный слой (код ИЛИ тест, не оба «наугад»).
- [ ] Записан ROOT + FIX + VERIFIED в skill-лог.
- [ ] Нет `t.Skip`/удаления теста без обоснования.
- [ ] Регрессия затронутого пакета зелёная.

---

## 13. Стабильность и флейки

Флейк — тест, который то проходит, то падает без изменений кода. Флейки разрушают доверие к CI: команда начинает игнорировать красное, пропуская реальные баги. Hub держит флейки на нуле.

### 13.1. Причины флейков и лечение

| Причина | Проявление | Лечение |
|---------|------------|---------|
| Гонки (data race) | падение только под нагрузкой/иногда | `-race` всегда; исправить синхронизацию (mutex, channel) |
| Время / таймауты | `waitForTimeout`, `time.Sleep`, зависимость от «сейчас» | web-first assertions; fake clock; инъекция времени |
| Порядок тестов | падение при перемешивании | изоляция состояния (свой sqlite/QueryClient, сброс стора) |
| Общий внешний ресурс | конфликт за один проект/канал в БД | уникальные данные (uuid/суффикс времени) |
| Сеть / внешние сервисы | зависимость от реального API | мок (MSW/`vi.spyOn`), `onUnhandledRequest: "error"` |
| Недетерминированный вывод | map iteration order, случайные ID | сортировать перед сравнением; фиксировать seed |
| Auto-wait не дождался | элемент появляется асинхронно | увеличить таргетный timeout у конкретного `expect`, не глобальный sleep |

### 13.2. Детект флейков

**Go — повторный прогон подозрительного теста:**

```bash
# Прогнать 20 раз; -count отключает кэш, каждый прогон реальный.
go test -tags hubtest -race -run TestSuspect -count=20 ./testframework/<pkg>/...
# Перемешать порядок (Go 1.17+):
go test -tags hubtest -shuffle=on ./testframework/<pkg>/...
```

**Playwright — прогон N раз:**

```bash
npx playwright test path/to.spec.ts --repeat-each=10
```

Если тест падает хоть раз из N — он флейки, чинить причину (не ретраить).

### 13.3. Ретраи — осторожно

Playwright поддерживает `retries`. В Hub ретраи допустимы для смягчения инфраструктурной нестабильности в E2E, **но не как замена фиксу**:
- ретрай маскирует флейк, а не устраняет причину;
- трейс собирается `on-first-retry` — используем его для диагностики;
- unit-тесты (Go/Vitest) **не ретраим** — они должны быть детерминированы.

**Bad — ретрай вместо фикса:**

```typescript
// Тест флейки из-за sleep-синхронизации; накрутили retries=5 и забыли
test.describe.configure({ retries: 5 });
```

**Good — устранить причину, минимальный ретрай на инфру:**

```typescript
// web-first assertion вместо sleep + 1 ретрай на сетевую нестабильность CI
await expect(page).toHaveURL(/\/projects/, { timeout: 15_000 });
```

### 13.4. Quarantine (карантин)

Если флейк нельзя починить немедленно, а он блокирует CI:
1. Пометить тест с явной причиной и ticket-ссылкой (не голый `t.Skip`).
2. Занести в `.planning/STATE.md` как долг с owner и сроком.
3. Не оставлять в карантине «навсегда» — карантин это ticket, а не решение.

```go
func TestFlakyThing(t *testing.T) {
	t.Skip("QUARANTINE: флейк из-за гонки в ws pump, тикет HUB-1234, owner qa-auto1")
	// ...
}
```

Verifier обязан проверять: нет ли `t.Skip`/`test.skip` без причины и тикета.

### 13.5. Борьба с флейками — детерминизм по умолчанию

Принципы, делающие тесты детерминированными:
- **никакого реального времени** — инъекция clock или проверка через web-first assertions;
- **никакой реальной сети** в unit — только моки;
- **изоляция состояния** — свой ресурс на тест, сброс синглтонов;
- **уникальные данные** — uuid / суффикс времени;
- **`-race` и `-shuffle=on`** в арсенале при подозрении.

### 13.6. Чек-лист стабильности

- [ ] Нет `time.Sleep`/`waitForTimeout` для синхронизации.
- [ ] `-race` зелёный; подозрительный тест прогнан `-count=20`/`--repeat-each`.
- [ ] Состояние изолировано, синглтоны сброшены.
- [ ] Уникальные данные на прогон.
- [ ] Ретраи не маскируют неисправленный флейк.
- [ ] Карантин — только с причиной, тикетом и owner в STATE.md.

---

## 14. Coverage

### 14.1. Порог покрытия Hub

| Область | Порог | Текущее |
|---------|-------|---------|
| Statement coverage `internal/` | ≥ 82% (gate) | 83.0% |
| Security-пакеты (auth, ws, rbac, audit, calls) | ≥ 80% | 6/6 ✅ |

Gate живёт в CI job `coverage-gate` (push в master). Падение ниже порога блокирует.

### 14.2. Как считать покрытие

```bash
# Профиль в TMP (rule 45)
cd backend
go test -tags hubtest -race -coverprofile="${TMPDIR:-/tmp}/hub-cov.out" -covermode=atomic ./testframework/...
# Общий процент:
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | tail -1
# Пакеты с нулевым покрытием:
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | rg " 0.0%"
# HTML-отчёт (для локального анализа, не коммитить):
go tool cover -html="${TMPDIR:-/tmp}/hub-cov.out" -o "${TMPDIR:-/tmp}/cov.html"
```

`-covermode=atomic` обязателен вместе с `-race` (иначе счётчики покрытия сами дают гонку).

**Frontend:**

```bash
cd frontend
npx vitest run --coverage
```

### 14.3. Что покрывать в приоритете

| Пакет | Приоритет | Причина |
|-------|-----------|---------|
| auth, ws, rbac | Critical | безопасность, panic-risk, authz |
| audit, calls | Critical (≥80%) | security-инварианты |
| tracker, messenger, tms | High | ядро продукта |
| платформенные (errors, http, middleware) | High | используются везде |

Покрывать в первую очередь: error paths, граничные значения, security-инварианты (SEC-01..07, NF-02/03/07).

### 14.4. Что можно НЕ покрывать

- инфраструктурные пакеты подключений: `platform/config`, `platform/database`, `platform/redis` (тонкие обёртки, проверяются интеграционно/e2e);
- сгенерированный код (`ent/`), кроме кастомных хуков;
- тривиальные геттеры без логики.

**Но:** «не покрывать» ≠ «нет тестов на фичу». Если пакет содержит логику — она покрывается.

### 14.5. Coverage — не самоцель

**Bad — тест ради процента:**

```go
func TestGetterCoverage(t *testing.T) {
	i := &ent.Issue{Title: "x"}
	_ = i.Title // «покрыли» геттер, ничего не проверив
}
```

**Good — тест проверяет поведение:**

```go
func TestUpdateIssueOptimisticLocking(t *testing.T) {
	// проверяет реальный контракт: конфликт версий → 409
}
```

Высокий процент при слабых ассертах — ложное чувство безопасности. Coverage измеряет исполненные строки, не проверенное поведение.

### 14.6. Ведение `.planning/COVERAGE.md`

`/coverage-audit` создаёт/обновляет матрицу: требование → код → тест → статус. После coverage-задачи:
1. зафиксировать проценты и wall-time в `.planning/COVERAGE.md`;
2. обновить статусы (✅ / ⚠️ / 🔴);
3. удалить профили (rule 45);
4. `git status` — без coverage-мусора.

### 14.7. Чек-лист coverage

- [ ] Профиль считается с `-covermode=atomic` (+`-race`), пишется в TMP.
- [ ] Общий `internal/` ≥ 82%, security-пакеты ≥ 80%.
- [ ] Новый код покрыт: error paths + границы + инварианты.
- [ ] Нет тестов «ради процента» без ассертов поведения.
- [ ] `.planning/COVERAGE.md` обновлён, проценты зафиксированы.
- [ ] Профили удалены, git чист.

---

## 15. Definition of Done

DoD разделён для ручных и авто-задач. Верификаторы (`qa-manual-ver`, `qa-auto-ver`) проверяют соответствующий список.

### 15.1. DoD — ручная тест-задача (`qa-manual`)

- [ ] Требование/US прочитано, границы понятны.
- [ ] Для каждого сценария создана строка TC-`<MODULE>`-NN в `docs/TESTING.md`.
- [ ] Заполнены все колонки: ID, название, предусловие, шаги, ожидаемый результат.
- [ ] Применены техники тест-дизайна (классы, границы, таблицы решений) — покрыты позитив и негатив.
- [ ] Покрыты error paths (400/401/403/404/409/429 где применимо).
- [ ] Колонка «Автотест» = «— (ручной)» или TODO (если ещё не автоматизировано).
- [ ] Сводка автоматизации модуля обновлена.
- [ ] Баг-репорты (если найдены) в формате `BUG-<module>-NN` с severity.
- [ ] `.planning/STATE.md` обновлён.

### 15.2. DoD — авто-задача (`qa-auto1`/`qa-auto2`)

- [ ] Тест написан по паттернам Hub (mockRepo/env.Builder/POM/renderWithProviders/MSW).
- [ ] Тег/каркас корректны (`_test` пакет; `//go:build hubtest` для export hooks).
- [ ] Прогон зелёный: BE `-tags hubtest -race -count=1`; FE `tsc` + `vitest`; E2E стабилен.
- [ ] Покрыты позитив, негатив, границы; ассерты проверяют поведение, не мок.
- [ ] Тест детерминирован (нет sleep/сети/общего состояния); прогнан на флейк.
- [ ] Двусторонняя привязка: комментарий/аннотация с TC-ID в коде + имя в колонке «Автотест».
- [ ] Allure-метаданные (если заливается в TMS): AS_ID/tms-link, epic/feature/severity, без секретов.
- [ ] Coverage не упал ниже порога; для нового кода добавлены тесты.
- [ ] Артефакты удалены (rule 45), git чист.
- [ ] Docs sync: `docs/TESTING.md` + `.planning/STATE.md` (+ COVERAGE.md при coverage-задаче).
- [ ] `/qa-gate` пройден, лог приложен.

### 15.3. Что проверяет `qa-manual-ver`

- Полнота TC: все сценарии требования покрыты (нет дыр в позитиве/негативе/границах).
- Корректность техник тест-дизайна (классы/границы/таблицы применены осмысленно).
- Трассируемость: requirement → TC (каждое требование имеет TC).
- Тестируемость шагов: шаги конкретны и воспроизводимы, ожидаемый результат наблюдаем.
- Ручные TC явно помечены, не выданы за автоматизированные.

### 15.4. Что проверяет `qa-auto-ver`

- **Детерминизм:** нет sleep/flaky; прогон повторно зелёный (`-count`/`--repeat-each`).
- **Привязка TC↔автотест:** имя из колонки существует в коде; тест реально проверяет описанный сценарий.
- **Покрытие:** новый код покрыт, порог держится; ассерты проверяют поведение.
- **Замер времени:** прогон не деградировал; медленные тесты оправданы.
- **Регрессия:** приложен зелёный лог `/qa-gate`.
- **Triage при фиксах:** записан ROOT=code|test, нет «зелёных вслепую».
- **Data race:** `-race` зелёный.
- **Артефакты:** git чист, профили в TMP.

---

## 16. Антипаттерны тестов

Сводная таблица «нельзя → почему → как правильно». Верификаторы используют её как чек-лист отклонения.

| Нельзя | Почему плохо | Как правильно |
|--------|--------------|---------------|
| `time.Sleep`/`waitForTimeout` для синхронизации | флейк + трата времени CI | web-first assertions, fake clock |
| `assert` перед разыменованием результата | nil-паника, невнятная ошибка | `require.NoError`/`require.NotNil` сначала |
| Мок возвращает `nil, nil` на всё | тест зелёный, ничего не проверяет | эмулировать контракт (ошибки, not found) |
| Общий мутабельный fixture | гонки, зависимость от порядка | свой sqlite/QueryClient/сброс стора на тест |
| `key={index}` в тест-данных React | несоответствие реальному коду | стабильные ID (как в проде) |
| Подгонка assertion под вывод | фиксирует баг как «правильное» | triage: сначала спека (раздел 12) |
| `t.Skip`/`test.skip` без причины | молча теряется покрытие | причина + тикет + owner (карантин) |
| Тест на приватные детали реализации | ломается при рефакторинге | assert на публичный контракт |
| Snapshot всего дерева компонента | шум, бездумные `--update` | точечные `getByRole`/`getByText` |
| CSS-селекторы `div > .cls:nth-child` | хрупко к вёрстке | `getByRole`/`getByLabel` |
| E2E ради unit-логики (ice-cream cone) | медленно, флейки, жжёт CI-минуты | логику в unit, E2E — критический путь |
| Реальная сеть в unit | недетерминизм, зависимость от бэкенда | MSW/`vi.spyOn`, `onUnhandledRequest:"error"` |
| testcontainers в каждом unit | секунды на тест, лимит CI | мок или sqlite `env.Builder` |
| coverage-профиль в `backend/` | мусор в git (rule 45) | профиль в TMP, cleanup |
| Секреты в Allure attachment/логах | утечка JWT/PII в отчёты | не прикреплять секреты |
| Автотест без TC-ID и без строки в TESTING.md | нет traceability | комментарий TC-ID + колонка «Автотест» |
| Матчинг Allure только по fullName/className | рвётся при переименовании | AS_ID label / tms-link с `HUB-TC-*` |
| `fmt.Println`/`console.log` в тестах | шум CI, Biome-fail | сообщения ассертов, временный `t.Logf` |
| Webhook-тест без проверки API Key | пропуск SEC-01/SEC-02 | проверять отказ без ключа |
| `panic`/`log.Fatal` в тест-хелперах | обрывает прогон без диагностики | `t.Fatal`/`require` |

---

## 17. Чек-лист перед передачей верификатору

Финальная проверка перед хендоффом `qa-auto-ver` / `qa-manual-ver`. Пройди сверху вниз; любой не-отмеченный пункт = работа не готова.

### 17.1. Общее

- [ ] Задача соответствует требованию/US; границы понятны.
- [ ] Слой и домен не пересекаются с другим `qa-auto` (раздел 1.4).
- [ ] Не дублирует существующий TC/автотест.

### 17.2. Тесты

- [ ] Написаны по паттернам Hub (mockRepo / env.Builder / factory / POM / Steps / renderWithProviders / MSW).
- [ ] Позитив + негатив + границы + error paths.
- [ ] Ассерты проверяют поведение и контракт, не мок и не детали реализации.
- [ ] Детерминированы: нет sleep/реальной сети/общего состояния; прогнаны на флейк.
- [ ] BE: зелёно на `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...`.
- [ ] FE: `npx tsc --noEmit` + `npx vitest run` зелёные, Biome чист.
- [ ] E2E: стабильны, backend-зависимые с `skipUnless*`, уникальные данные.

### 17.3. Traceability и Allure

- [ ] Строка TC-`<MODULE>`-NN есть в `docs/TESTING.md`.
- [ ] Код → MD: комментарий/аннотация с TC-ID в тесте.
- [ ] MD → код: точное имя функции/spec в колонке «Автотест».
- [ ] Сводка «% автоматизации» обновлена.
- [ ] Allure (если TMS): AS_ID/tms-link `HUB-TC-*`, epic/feature/severity, без секретов.

### 17.4. Качество и процесс

- [ ] Coverage держит порог (≥82% / security ≥80%); новый код покрыт.
- [ ] Wall-time прогона известен, деградации нет; медленные тесты оправданы.
- [ ] При падениях применён triage с записью ROOT=code|test (нет «зелёных вслепую»).
- [ ] `/qa-gate` пройден, лог приложен.
- [ ] Docs sync: `docs/TESTING.md` + `.planning/STATE.md` (+ `.planning/COVERAGE.md` при coverage).
- [ ] Артефакты удалены (rule 45): `git status` без `*.out`, `cover*`, `cov_*`, `tmp_*`, `test-results/`, `allure-results/`.

### 17.5. Финальное

- [ ] `verified: true` ставится только после green regression + приложенного лога.
- [ ] Хендофф-заметка в `.planning/STATE.md`: что сделано, что осталось, ссылки на TC/тесты.

---

---

## Приложение A. Cookbook по доменам

Для каждого домена — что тестировать, какие паттерны применять, реальные имена тестов из `backend/testframework/` и `docs/TESTING.md`. Используй как отправную точку при работе над доменом.

### A.1. auth (Magic Link only, v2.5)

**Инварианты (TZ):** только Magic Link (без паролей/SSO); Telegram — только notify bind, не login; JWT access 15 мин; refresh Redis 30 дней с instant revocation; rate limit 3/10 мин; сравнение секретов через `subtle.ConstantTimeCompare`.

| Что тестировать | Слой | Тест / TC |
|-----------------|------|-----------|
| Запрос кода magic link | handler | TC-AUTH-01 |
| Верификация кода → токены | service | `TestVerifyMagicLink_*`, TC-AUTH-02 |
| Неверный код (constant-time) | service | TC-AUTH-03, SEC-07 |
| JWT round-trip, expiry | service | `TestJWTExpired`, TC-AUTH-05 |
| Refresh обновление/отзыв | service | TC-AUTH-04 |
| Rate limit 3/10 мин | middleware | TC-AUTH-10 |
| Онбординг, logout | e2e | `onboarding-profile.spec.ts`, `logout.spec.ts` |

**Ключевой паттерн — проверка constant-time сравнения (SEC-07):**

```go
// TC-AUTH-03: неверный код отклоняется; сравнение через ConstantTimeCompare (timing-safe).
func TestVerifyMagicLink_WrongCode(t *testing.T) {
	svc, _ := newAuthTestService()
	ctx := context.Background()

	_, err := svc.VerifyMagicLink(ctx, "user@hub.local", "000000")
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	// 400/401 по контракту; главное — не 500 и не паника.
	assert.Contains(t, []int{400, 401}, appErr.Code)
}
```

**Anti-regression:** endpoints `/auth/telegram*` **удалены** (ADR-AUTH-01). Тест не должен их воскрешать; если встречаешь — это устаревший тест (ROOT=test).

### A.2. tracker (задачи, workflow, доски)

**Инварианты:** optimistic locking через `version`; issue_key автоинкремент (`HUB-1`, `HUB-2`); soft delete; 9 типов задач; workflow-переходы валидируются.

| Что тестировать | Слой | Тест / TC |
|-----------------|------|-----------|
| Создание, автоинкремент | service | `TestCreateIssue`, `TestCreateIssueAllTypes`, TC-TRCK-01/02 |
| Optimistic locking, конфликт | service | `TestUpdateIssueOptimisticLocking`, TC-TRCK-03/04 |
| Workflow валидный/невалидный переход | service | `TestMoveIssueValidTransition`, `TestMoveIssueInvalidTransition`, TC-TRCK-10/11 |
| Связи задач | service | `TestCreateIssueLink`, TC-TRCK-06 |
| JQL поиск | service | `TestParseJQLSimple`, `TestToEntPredicates`, TC-TRCK-09 |
| Метки, доски, кастомные поля | service | `TestCreateLabel`, `TestBoardData`, `TestCustomFieldValues` |
| Пагинация (cursor) | service | `TestListIssuesPagination` |

Паттерн — `mockRepo` с `projectCounters` для автоинкремента (см. раздел 4.7). Optimistic locking — эталонный тест раздела 4.4.

### A.3. messenger (каналы, DM, треды, реакции)

**Инварианты:** DM дедупликация; read-only канал (403 для не-Owner); unread не инкрементит отправителя; WS room только для участника (NF-02).

| Что тестировать | Тест / TC |
|-----------------|-----------|
| DM дедуп | `TestCreateDM_Dedup`, TC-MSG-02 |
| Отправка с проверкой участника | `TestSendMessage_MemberCheck`, TC-MSG-03 |
| Реакция toggle | `TestToggleReaction_AddRemove`, TC-MSG-04 |
| Тред + счётчик ответов | `TestThreadReply_IncrementsReplyCount`, TC-MSG-06 |
| Read-only (таблица решений) | `TestSendMessage_ReadOnlyBlocked_RegularUser`, `TestSendMessage_ReadOnlyAllowed_Admin`, TC-MSG-07 |
| Unread (skip sender) | `TestIncrement_SkipsSender`, `TestMarkRead_ResetsToZero`, TC-MSG-09 |
| Presence | `TestHeartbeat_SetsScore`, `TestGetStatus_Online/Away/Offline` |

**Паттерн — read-only как таблица решений (раздел 3.4):**

```go
// TC-MSG-07: матрица (is_readonly × роль) → разрешение отправки.
func TestSendMessage_ReadOnly_DecisionTable(t *testing.T) {
	cases := []struct {
		name       string
		isReadonly bool
		role       string
		wantCode   int // 0 = успех
	}{
		{"open channel, member", false, "member", 0},
		{"readonly, admin allowed", true, "admin", 0},
		{"readonly, member blocked", true, "member", 403},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// arrange канал с tc.isReadonly, пользователь с tc.role
			// act: отправка
			// assert: tc.wantCode
		})
	}
}
```

### A.4. tms (тест-кейсы, прогоны, Allure)

**Инварианты:** версионирование кейсов; case_key `HUB-TC-N`; Allure matching по стратегиям (раздел 8.2); webhook требует API Key.

| Что тестировать | Тест / TC |
|-----------------|-----------|
| CRUD кейса (classic/BDD) | `TestCreateTestCase_Classic`, `TestCreateTestCase_BDD`, TC-TMS-01/02 |
| Версии + rollback | `TestUpdateTestCase_CreatesVersion`, `TestRollbackTestCase`, TC-TMS-03/04 |
| Прогон + результаты | `TestAddTestCasesToRun`, `TestExecuteTestResult_UpdatesStats`, TC-TMS-05/06 |
| Allure parse | `TestParseAllureResults_*`, TC-ALLURE-01..04 |
| Allure matching | `TestMatchTestCase_By*`, `TestMatchTestCase_PriorityOrder`, TC-ALLURE-05 |
| Bug from failed | `TestCreateBugFromFailedResult`, TC-TMS-09 |
| Webhook API Key | `handler_webhook_rbac_test.go`, SEC |

Allure — центральный домен playbook, см. раздел 8. Паттерн частичного мока (`mockMatcherRepo` с embedded interface) — раздел 4.7.

### A.5. calls (LiveKit, запись, webhook)

**Инварианты:** webhook signature обязательна (SEC-02, не skip); лимит активных записей; NATS nil-safe publish; эфемерный чат с лимитом длины.

| Что тестировать | Тест / TC |
|-----------------|-----------|
| LiveKit токен | `TestGenerateToken`, `TestGenerateToken_DifferentRooms`, TC-CALL-01/02 |
| Webhook подпись (reject invalid) | `TestVerifyWebhookSignature_EmptyAuth`, `TestVerifyWebhookSignature_InvalidJWT`, SEC-02 |
| Лимит записей | `TestMaxActiveRecordingsLimit`, TC-CALL-07 |
| Эфемерный чат + лимит длины | `TestSendMessage`, `TestSendMessage_TextTooLong`, TC-CALL-09 |

**Паттерн — webhook подпись (SEC-02, нельзя skip):**

```go
// SEC-02: webhook без валидной подписи отклоняется даже при отсутствии claim.
func TestVerifyWebhookSignature_EmptyAuth(t *testing.T) {
	svc := newCallsTestService()
	err := svc.VerifyWebhookSignature("", []byte(`{"event":"x"}`))
	require.Error(t, err) // пустая подпись → отказ, не пропуск
}
```

### A.6. sprint (state machine)

**Инварианты:** один active-спринт на проект; переходы planning→active→completed; перенос незакрытых при complete.

| Что тестировать | Тест / TC |
|-----------------|-----------|
| Создание, дата конца < начала | `TestCreateSprint`, `TestCreateSprint_EndDateBeforeStartDate`, TC-SPRINT-01 |
| Start, запрет второго active | `TestStartSprint_Success`, `TestStartSprint_RejectsWhenActiveExists`, TC-SPRINT-02/03 |
| Burndown, velocity | `TestGetBurndown`, `TestGetVelocity`, TC-SPRINT-04/05 |
| Complete + перенос | `TestCompleteSprint_MovesIncompleteIssues`, TC-SPRINT-07 |

Эталон техники «переходы состояний» — раздел 3.5.

### A.7. ws (real-time)

**Инварианты:** send after disconnect без паники (SEC-05); non-member subscribe отклоняется (NF-02); Fiber WS (не gorilla/net/http).

| Что тестировать | Тест |
|-----------------|------|
| Message-типы, JSON | `TestMessageTypeConstants`, `TestClientMessage_JSON`, TC-WS-01/02 |
| Ping без room | `TestClientMessage_PingHasNoRoom`, TC-WS-03 |
| Authz подписки (NF-02) | `authorizer_test.go`, `handler_auth_test.go` |
| Shutdown без паники (SEC-05) | `hub_shutdown_test.go`, `pump_test.go` |

`-race` здесь критичен — домен полон горутин.

### A.8. rbac, search, notification (кратко)

| Домен | Ключевые тесты | Инвариант |
|-------|----------------|-----------|
| rbac | `TestPermission_Has/Add/Remove`, `TestDefaultRolePermissions` | битовые операции идемпотентны; матрица ролей |
| search | `TestMergeResultsByScore`, JQL-тесты | user-scope фильтр (NF-03) — результаты только по проектам/каналам пользователя |
| notification | `TestSubjectToType`, `TestShouldDispatch_Debounce` | quiet hours; urgent пробивает; WebPush endpoint валидируется (NF-07) |

**Паттерн — user-scope фильтр поиска (NF-03):**

```go
// NF-03: результаты поиска ограничены проектами/каналами пользователя.
func TestSearch_UserScoped(t *testing.T) {
	// arrange: две сущности в разных проектах, пользователь — участник только одного
	// act: поиск
	// assert: возвращена только сущность из «своего» проекта
}
```

---

## Приложение B. Полный worked-example: фича end-to-end

Задача: добавить **WIP-лимит на колонку Kanban** (нельзя перенести задачу в колонку, где достигнут лимит). Показываем полный QA-цикл: TC → BE unit → FE unit → E2E → Allure → docs sync.

### B.1. Шаг 1 — `qa-manual` пишет TC

В `docs/TESTING.md`, модуль TRCK:

```markdown
| TC-TRCK-13 | WIP-лимит: перенос в полную колонку | Колонка "In Progress" wip_limit=2, в ней 2 задачи | Перетащить 3-ю задачу в колонку | 409/400, перенос отклонён, показана ошибка лимита | — (ручной) |
| TC-TRCK-14 | WIP-лимит: перенос при свободном месте | wip_limit=2, в колонке 1 задача | Перетащить 2-ю задачу | 200, задача перенесена | — (ручной) |
```

Границы (раздел 3.2): limit−1 (успех), limit (граница — ещё успех при перемещении внутри?), limit+1 (отказ).

### B.2. Шаг 2 — `qa-auto1` пишет BE unit

```go
// TC-TRCK-13, TC-TRCK-14: WIP-лимит колонки при перемещении задачи.
func TestMoveIssue_WIPLimit(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	defaultID := repo.getDefaultStatusID()
	inProgress := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, defaultID)
	repo.setWIPLimit(inProgress, 2) // хелпер мока: лимит колонки = 2

	// Заполняем колонку до лимита.
	for i := 0; i < 2; i++ {
		iss, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
			Title: fmt.Sprintf("Filler %d", i), IssueType: "task", Priority: "medium",
		})
		require.NoError(t, err)
		_, err = svc.MoveIssue(ctx, iss.IssueKey, userID, tracker.MoveIssueRequest{StatusID: inProgress.String(), Rank: "a"})
		require.NoError(t, err)
	}

	// TC-TRCK-13: третья задача — отказ по лимиту.
	third, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Third", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)
	_, err = svc.MoveIssue(ctx, third.IssueKey, userID, tracker.MoveIssueRequest{StatusID: inProgress.String(), Rank: "b"})
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Contains(t, []int{400, 409}, appErr.Code, "перенос сверх WIP-лимита должен отклоняться")
}
```

### B.3. Шаг 3 — `qa-auto2` пишет FE unit

```typescript
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { KanbanColumn } from "@/components/tracker/kanban-column";

describe("KanbanColumn WIP limit", () => {
	it("shows count vs limit and highlights when full", () => {
		render(
			<KanbanColumn
				name="In Progress"
				wipLimit={2}
				issues={[{ id: "1", issue_key: "HUB-1", title: "A", issue_type: "task", priority: "low", status: "in_progress" },
				         { id: "2", issue_key: "HUB-2", title: "B", issue_type: "task", priority: "low", status: "in_progress" }]}
			/>,
		);
		// Счётчик 2/2 виден.
		expect(screen.getByText("2/2")).toBeInTheDocument();
	});
});
```

### B.4. Шаг 4 — E2E (только если критический путь)

WIP-лимит критичен для UX доски → добавляем шаг в существующий сценарий доски, а не отдельный тяжёлый spec. Используем POM/Steps, dev-login, web-first assertions.

```typescript
// В домене tracker; проверяем сообщение об отказе при переносе сверх лимита.
test("WIP limit blocks over-fill", async ({ page, backendAvailable, devModeAvailable }) => {
	skipUnlessBackend(backendAvailable);
	skipUnlessDevMode(devModeAvailable);
	const auth = new AuthSteps(page);
	await auth.loginAsAdmin(page);
	// ... открыть доску с настроенным лимитом, перетащить сверх лимита,
	await expect(page.getByText(/лимит/i)).toBeVisible({ timeout: 10_000 });
});
```

### B.5. Шаг 5 — Allure-аннотация (при заливке в TMS)

```typescript
await allure.epic("Tracker");
await allure.feature("Kanban WIP limit");
await allure.label("AS_ID", "HUB-TC-TRCK-13");
await allure.severity("normal");
```

### B.6. Шаг 6 — docs sync и привязка

Обновить колонку «Автотест» и сводку:

```markdown
| TC-TRCK-13 | WIP-лимит: перенос в полную колонку | ... | 409/400 | `TestMoveIssue_WIPLimit` |
| TC-TRCK-14 | WIP-лимит: перенос при свободном месте | ... | 200 | `TestMoveIssue_WIPLimit` |
```

`TRCK: N → N+2 автоматизировано`. Заметка в `.planning/STATE.md`.

### B.7. Шаг 7 — регрессия и хендофф

```bash
cd backend && go test -tags hubtest -race -count=1 ./testframework/tracker/...
cd frontend && npx tsc --noEmit && npx vitest run ../testframework/stages/unit/frontend-test/domains/tracker
/qa-gate
```

Пирамида соблюдена: 1 BE unit (закрывает границы через один тест с наполнением), 1 FE unit, 0-1 E2E. Traceability двусторонняя. Артефакты вычищены.

---

## Приложение C. Каталог security-critical тестов

Эти тесты — приоритет №1 (из `docs/test_plan.md`). Их нельзя удалять/ослаблять; изменение — только с `security-auditor`.

| ID | Модуль | Требуемая проверка | Реальный тест |
|----|--------|--------------------|---------------|
| SEC-01 | apigateway | API key middleware order → 401 без ключа | `apikey_test.go`, `handler_webhook_rbac_test.go` |
| SEC-02 | calls | Webhook reject missing/tampered signature | `TestVerifyWebhookSignature_*` |
| SEC-05 | ws | send after disconnect — без паники | `hub_shutdown_test.go`, `pump_test.go` |
| SEC-06 | auth | concurrent register / first-admin lock | auth concurrency-тесты |
| SEC-07 | auth | `ConstantTimeCompare` на magic link | verify-тесты |
| NF-02 | ws | non-member subscribe → reject | `authorizer_test.go` |
| NF-03 | search | результаты в user-scope | search service-тесты |
| NF-07 | notification | WebPush invalid endpoint → 400 | notification-тесты |

**Правило:** при работе рядом с любым из этих тестов — прогнать security-пакеты и убедиться, что покрытие ≥80% держится. Изменение поведения security-инварианта требует `security-auditor` + `gsd-secure-phase`.

**Запрещённые паттерны (TZ FORBIDDEN), которые ловятся тестами/ревью:**

| Паттерн | Риск | Проверка |
|---------|------|----------|
| `http.DefaultClient` для user-URL | SSRF | `ValidateOutboundURL`/safehttp |
| `io.ReadAll` без `io.LimitReader` | memory DoS | лимитированное чтение тела |
| `!=` для секретов | timing attack | `subtle.ConstantTimeCompare` |
| `*ent.Client` в service/handler | нарушение layering | repository interface |
| `panic()` в prod path | краш | возврат ошибки |
| skip webhook signature | подделка события | обязательная проверка |

---

## Приложение D. Справочник команд и карта CI

### D.1. Команды прогона

```bash
# Backend — как CI (job backend-test)
cd backend
go generate ./ent
go vet -tags hubtest ./...
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Backend — один домен / один тест
go test -tags hubtest -race ./testframework/tracker/...
go test -tags hubtest -run TestCreateIssue ./testframework/tracker/...

# Backend — coverage (профиль в TMP, rule 45)
go test -tags hubtest -race -covermode=atomic -coverprofile="${TMPDIR:-/tmp}/hub-cov.out" ./testframework/...
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | tail -1

# Backend — тайминги / медленные тесты
go test -tags hubtest -v ./testframework/tracker/... 2>&1 | rg -- "--- (PASS|FAIL)"
go test -tags hubtest -json ./testframework/... > "${TMPDIR:-/tmp}/gotest.json"

# Backend — бенчмарк
go test -tags hubtest -run '^$' -bench . -benchmem ./testframework/tms/...

# Backend — детект флейка
go test -tags hubtest -race -run TestSuspect -count=20 ./testframework/<pkg>/...
go test -tags hubtest -shuffle=on ./testframework/<pkg>/...

# Prod build check
go build -tags release -o bin/server ./cmd/server

# Frontend — как CI (job frontend-test)
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
npm run lint
npx vitest run
# один домен
npx vitest run ../testframework/stages/unit/frontend-test/domains/auth

# E2E
npm run test:e2e:smoke   # job e2e-smoke (3 spec, без backend)
npm run test:e2e         # job e2e-full (20 spec, с backend)
npx playwright show-trace test-results/<...>/trace.zip

# Full gate
/qa-gate
```

### D.2. Карта CI job → путь → домены

| CI Job | Стадия | Путь | Домены |
|--------|--------|------|--------|
| `backend-test` | unit | `backend/testframework/{domain}/` | auth, tracker, tms, messenger, calls, ws, sprint, … |
| `frontend-test` | unit | `stages/unit/frontend-test/domains/{domain}/` | auth, admin, ai, tracker, tms, v4, … |
| `build` | build | — | только сборка (`tsc`, `vite`, `-tags release`) |
| `e2e-smoke` | e2e | `stages/e2e/domains/` (3) | auth, platform |
| `e2e-full` | e2e | `stages/e2e/domains/` (20) | все 9 доменов |
| `coverage-gate` | quality | `backend/testframework/` | все BE, ≥82% |

### D.3. Docker-прогон backend (канонический, как в CI)

```bash
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."
```

### D.4. Быстрая навигация по документам

| Нужно | Документ |
|-------|----------|
| TC-матрица, API-коды, manual checklist | `docs/TESTING.md` |
| Стратегия, пирамида, CI pipeline | `docs/test_plan.md` |
| Карта testframework | `testframework/PIPELINE.md` |
| Паттерны тестов (rule) | `.cursor/rules/40-testing.mdc` |
| Cleanup артефактов | `.cursor/rules/45-artifacts-cleanup.mdc` |
| Регрессия и верификаторы | `.cursor/rules/56-regression-verify.mdc` |
| Triage FAIL | `.cursor/skills/test-triage/SKILL.md` |
| Полный gate | `.cursor/skills/qa-gate/SKILL.md` |
| Coverage-аудит | `.cursor/skills/coverage-audit/SKILL.md` |
| Матрица покрытия | `.planning/COVERAGE.md` |
| Session state | `.planning/STATE.md` |

---

## Приложение E. FAQ и типовые ошибки агентов

**Q: Тест падает, я вижу почему — можно сразу поправить assertion?**
Нет. Сначала triage (раздел 12): воспроизведи функционал по спеке, определи ROOT=code|test, запиши. Правка assertion без этого = фиксация бага.

**Q: Где писать новый Go-тест — в `internal/` рядом с кодом?**
Нет. `internal/*_test.go` = 0 (TF-M08). Все тесты в `backend/testframework/{domain}/`. Приватные символы — через `export_hooks.go` (`//go:build hubtest`).

**Q: Нужен реальный Postgres для теста сервиса?**
Обычно нет. Мок (`mockRepo`) для логики; sqlite `env.Builder` для реальной схемы Ent. Реальный PG/Mongo — только для их специфики и только в CI.

**Q: Как гарантировать, что мой Allure-результат сматчится на нужный TC?**
Используй `AS_ID` label или tms-link с ключом `HUB-TC-*` (стратегия `MatchByAnnotation`, высший приоритет). Не полагайся на fullName/className.

**Q: E2E флейки — накрутить retries?**
Нет. Найди причину (обычно sleep вместо web-first assertion или общие данные). Ретрай — смягчение инфры, не замена фикса. Unit-тесты не ретраим вовсе.

**Q: Забыл обновить колонку «Автотест» — критично?**
Да. Без этого нет traceability (rule 55). Задача не готова к хендоффу.

**Q: qa-auto1 и qa-auto2 хотят тронуть один файл — как быть?**
Один файл — один владелец в рамках задачи. Делите по слою (BE/FE) и домену (раздел 1.4); пересечение = хендофф через STATE.md.

**Q: Куда писать coverage-профиль?**
В TMP (`${TMPDIR:-/tmp}`), никогда в `backend/`. После — cleanup (rule 45), `git status` чист.

**Q: Что делать, если спецификация неясна?**
Stop. Не угадывать поведение. Уточнить у SA/BA, обновить docs, потом писать код/тест (triage, шаг 3).

---

## Приложение F. Расширенные good/bad по частым сценариям

Быстрый справочник «как надо / как не надо» для ситуаций, которые встречаются чаще всего при работе над тестами Hub.

### F.1. Проверка ошибки создания сущности

**Bad — проверка по тексту ошибки (хрупко к переводам/рефакторингу):**

```go
_, err := svc.CreateIssue(ctx, projectID, userID, req)
assert.Equal(t, "issue validation failed: title required", err.Error())
```

**Good — по коду и подстроке поля:**

```go
_, err := svc.CreateIssue(ctx, projectID, userID, req)
require.Error(t, err)
appErr, ok := err.(*apperrors.AppError)
require.True(t, ok)
assert.Equal(t, 400, appErr.Code)
assert.Contains(t, err.Error(), "title") // указываем поле, не весь текст
```

### F.2. Тест с несколькими похожими случаями

**Bad — копипаста трёх почти одинаковых функций:**

```go
func TestCreateBug(t *testing.T)  { /* ... IssueType: "bug" ... */ }
func TestCreateTask(t *testing.T) { /* ... IssueType: "task" ... */ }
func TestCreateEpic(t *testing.T) { /* ... IssueType: "epic" ... */ }
```

**Good — table-driven + `t.Run` (см. реальный `TestCreateIssueAllTypes`):**

```go
for _, it := range []string{"bug", "task", "epic"} {
	t.Run(it, func(t *testing.T) {
		resp, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
			Title: "X", IssueType: it, Priority: "medium",
		})
		require.NoError(t, err)
		assert.Equal(t, it, resp.IssueType)
	})
}
```

### F.3. Ожидание асинхронного результата (FE)

**Bad — синхронный поиск сразу после рендера:**

```typescript
renderWithProviders(<IssueList projectId="p1" />);
expect(screen.getByText("HUB-42")).toBeInTheDocument(); // упадёт: данные ещё грузятся
```

**Good — `findBy*` (ждёт появления):**

```typescript
renderWithProviders(<IssueList projectId="p1" />);
expect(await screen.findByText("HUB-42")).toBeInTheDocument();
```

### F.4. Изоляция стора между тестами

**Bad — стейт течёт между тестами:**

```typescript
it("test A", () => { useAuthStore.getState().setTokens("a", "b"); /* ... */ });
it("test B", () => { /* isAuthenticated неожиданно true из test A */ });
```

**Good — сброс в `beforeEach`:**

```typescript
beforeEach(() => {
	sessionStorage.clear();
	useAuthStore.setState({ accessToken: null, refreshToken: null, user: null, isAuthenticated: false });
	vi.restoreAllMocks();
});
```

### F.5. E2E-навигация после действия

**Bad — sleep + проверка url строкой:**

```typescript
await createButton.click();
await page.waitForTimeout(2000);
expect(page.url()).toContain("/projects/");
```

**Good — web-first assertion с regex и таймаутом:**

```typescript
await createButton.click();
await expect(page).toHaveURL(/\/projects\/[0-9a-f-]+/, { timeout: 30_000 });
```

### F.6. Мок HTTP-клиента vs реальная сеть (FE)

**Bad — тест ходит в реальный backend (флейк, зависимость от окружения):**

```typescript
it("loads projects", async () => {
	const data = await fetchProjects(); // реальный fetch
	expect(data.length).toBeGreaterThan(0);
});
```

**Good — `vi.spyOn` на api-клиент (детерминизм):**

```typescript
const getSpy = vi.spyOn(api, "get").mockResolvedValue([{ id: "1", name: "Hub" }]);
const data = await fetchProjects();
expect(getSpy).toHaveBeenCalledWith("/projects");
expect(data).toHaveLength(1);
```

### F.7. Проверка события/побочного эффекта

**Bad — проверка, что метод «как-то отработал»:**

```go
err := svc.SendMessage(ctx, channelID, userID, "hi")
require.NoError(t, err) // но не проверили, что участник проверен и событие отправлено
```

**Good — проверка наблюдаемого контракта:**

```go
// TC-MSG-03: отправка проверяет членство и публикует NATS-событие.
msg, err := svc.SendMessage(ctx, channelID, memberID, "hi")
require.NoError(t, err)
assert.Equal(t, "hi", msg.Content)
// не-участник → отказ
_, err = svc.SendMessage(ctx, channelID, strangerID, "hi")
require.Error(t, err)
```

### F.8. Уникальность тестовых данных

**Bad — фиксированный ключ (конфликт при параллели/повторе):**

```go
proj, _ := client.Project.Create().SetKey("TST").Save(ctx) // второй тест упадёт на unique
```

**Good — уникальность через uuid (реальный env.Builder):**

```go
email := b.prefix + "-" + uuid.New().String()[:8] + "@test.com"
```

```typescript
const suffix = Date.now().toString(36).toUpperCase().slice(-6);
const projectKey = `E${suffix}`.slice(0, 10);
```

---

## Приложение G. Рубрика ревью для верификаторов

`qa-auto-ver` и `qa-manual-ver` используют эту рубрику для приёмки. Каждый пункт — PASS/FAIL; любой FAIL блокирует `verified: true`.

### G.1. Рубрика `qa-auto-ver`

| # | Критерий | PASS если | FAIL если |
|---|----------|-----------|-----------|
| 1 | Детерминизм | нет sleep/сети/общего состояния; повторный прогон зелёный | есть `waitForTimeout`/`time.Sleep`/реальный fetch |
| 2 | Race | `-race` зелёный | гонки или `-race` не запускался |
| 3 | Привязка TC | имя в колонке «Автотест» существует в коде | «✅» без имени, или имя не находится через `rg` |
| 4 | Осмысленность ассертов | проверяют поведение/контракт | проверяют мок или `nil,nil`; тест без ассертов |
| 5 | Покрытие | новый код покрыт, порог держится | падение ниже 82%/80% |
| 6 | Границы и негатив | есть позитив+негатив+границы | только happy path |
| 7 | Triage при фиксах | записан ROOT=code\|test | assertion подогнан под вывод |
| 8 | Время | нет деградации; медленные оправданы | новый тест тормозит без причины |
| 9 | Артефакты | git чист, профили в TMP | `*.out`/`cover*`/`test-results` в diff |
| 10 | Docs sync | TESTING.md + STATE.md обновлены | колонка не обновлена |

### G.2. Рубрика `qa-manual-ver`

| # | Критерий | PASS если | FAIL если |
|---|----------|-----------|-----------|
| 1 | Полнота TC | покрыты позитив/негатив/границы требования | дыры в сценариях |
| 2 | Техники тест-дизайна | классы/границы/таблицы применены осмысленно | случайный набор шагов |
| 3 | Трассируемость | каждое требование → TC | требование без TC |
| 4 | Воспроизводимость | шаги конкретны, результат наблюдаем | размытые «проверить, что работает» |
| 5 | Формат | все колонки заполнены, ID корректен | пропущены предусловие/ожидаемый результат |
| 6 | Честность статуса | ручные помечены «— (ручной)» | ручной выдан за авто |

### G.3. Как верификатор проверяет привязку (практика)

```bash
# 1. Извлечь имена из колонки «Автотест» в docs/TESTING.md и проверить их существование:
rg -o '`Test\w+`' docs/TESTING.md | tr -d '`' | sort -u | while read fn; do
  rg -q "func $fn" backend/testframework && echo "OK  $fn" || echo "MISS $fn"
done

# 2. Проверить обратную связь: тесты ссылаются на TC-ID
rg -c "TC-[A-Z]+-[0-9]+" backend/testframework | rg -v ":0"

# 3. Прогнать на флейк подозрительный тест
go test -tags hubtest -race -run <Name> -count=20 ./testframework/<pkg>/...
```

`MISS` в шаге 1 = битая привязка MD→код (FAIL критерия 3 рубрики G.1).

---

> **Итог.** Этот playbook — контракт качества QA Hub. `qa-manual` владеет описанием (TC), `qa-auto1/2` — реализацией (автотесты) и привязкой, верификаторы — приёмкой. Пирамида (много unit, мало E2E), детерминизм (без sleep/сети/общего состояния), traceability (requirement → TC → autotest) и Allure-матчинг по `HUB-TC-*` — четыре опоры, на которых держится доверие к зелёному CI. Экономь CI-минуты, чисти артефакты, и никогда не делай тест зелёным, не проверив функционал по спецификации.
