# Инструментарий: Allure, время прогона, артефакты

> Часть плейбука тестирования. Канон и оглавление — [`../testing.md`](../testing.md).
> Все измерения, reporters, validators и test commands выполняются только
> внутри task-owned Docker containers с resource limits. Host хранит лишь Git
> metadata и явно выбранные артефакты через task-owned mount.
> Historical Go coverage не является migration gate founder cutover; основной
> quality evidence нового Python слоя — traceability, risk coverage, JUnit,
> Allure, forced-failure sentinels и отсутствие unexpected skips.

Оформление отчётов, замер производительности прогонов, логи и артефакты.
Читается при настройке и разборе, не при написании теста.

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
