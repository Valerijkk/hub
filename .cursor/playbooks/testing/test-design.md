# Тест-дизайн и трассируемость

> Часть плейбука тестирования. Канон и оглавление — [`../testing.md`](../testing.md).
> Founder greenfield cutover: legacy tests не задают обязательный набор или
> counts. Каждый новый Python module проектируется заново от task + FR/NFR +
> verified TC и risk matrix: happy, negative, authn/authz, boundaries,
> edge/corner, timeout/retry, partial failure/recovery и applicable
> security/privacy/a11y/performance scenarios.

Техники проектирования кейсов и связь тест-кейса с автотестом.
Читается при проектировании покрытия, а не при каждой правке теста.

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
