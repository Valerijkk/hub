# Фаза 4: TMS и Allure -- Тестовая документация

## 1. Обзор

Документ содержит тест-кейсы для верификации всех требований Фазы 4. Каждый тест-кейс привязан к конкретному requirement ID и содержит предусловия, шаги и ожидаемый результат.

**Формат ID тест-кейса:** `TC-{REQ_ID}-{NN}` (например, TC-TMS-01-01).

---

## 2. Тест-кейсы по требованиям

### TMS-01: Тест-кейсы (классический + BDD)

**TC-TMS-01-01: Создание классического тест-кейса с шагами и ожидаемым результатом**
- Предусловия: пользователь аутентифицирован, имеет право PermTestCaseCreate в проекте
- Шаги:
  1. `POST /api/projects/:projectId/tms/cases` с телом `{"title": "Авторизация", "case_type": "classic", "priority": "high", "steps": [{"order": 1, "action": "Открыть /login", "expected_result": "Форма входа отображается"}]}`
- Ожидаемый результат: HTTP 201, тело содержит `case_key` вида "PRJ-TC-1", `case_type: "classic"`, `steps` с одним элементом, `version: 1`

**TC-TMS-01-02: Создание BDD тест-кейса с Given/When/Then**
- Предусловия: пользователь аутентифицирован, имеет право PermTestCaseCreate
- Шаги:
  1. `POST /api/projects/:projectId/tms/cases` с телом `{"title": "Вход по email", "case_type": "bdd", "steps": [{"order": 1, "action": "пользователь на странице входа", "keyword": "Given"}, {"order": 2, "action": "вводит email и нажимает Войти", "keyword": "When"}, {"order": 3, "action": "получает код на email", "expected_result": "код 6 цифр", "keyword": "Then"}]}`
- Ожидаемый результат: HTTP 201, `case_type: "bdd"`, шаги содержат поля `keyword`

**TC-TMS-01-03: Переключение типа тест-кейса classic <-> bdd**
- Предусловия: существует тест-кейс с `case_type: "classic"`
- Шаги:
  1. `PUT /api/projects/:projectId/tms/cases/:caseKey` с телом `{"case_type": "bdd", "version": 1}`
- Ожидаемый результат: HTTP 200, `case_type: "bdd"`, `version: 2`, создана версия-снимок предыдущего состояния

---

### TMS-02: Папки, теги, компоненты, версионирование

**TC-TMS-02-01: Создание иерархии папок и перемещение тест-кейсов**
- Предусловия: проект существует, пользователь с правами
- Шаги:
  1. `POST /api/projects/:projectId/tms/suites` с `{"name": "Smoke"}`
  2. `POST /api/projects/:projectId/tms/suites` с `{"name": "Login", "parent_id": "<smoke_id>"}`
  3. `GET /api/projects/:projectId/tms/suites`
- Ожидаемый результат: дерево папок содержит "Smoke" с дочерней "Login"

**TC-TMS-02-02: Добавление тегов и компонентов к тест-кейсу**
- Предусловия: существует тест-кейс и компонент "API"
- Шаги:
  1. `POST /api/projects/:projectId/tms/components` с `{"name": "API"}`
  2. `POST /api/projects/:projectId/tms/cases` с `{"title": "Test", "components": ["<component_id>"], "labels": ["<label_id>"]}`
  3. `GET /api/projects/:projectId/tms/cases/:caseKey`
- Ожидаемый результат: тест-кейс содержит `components` и `labels` в ответе

**TC-TMS-02-03: Версионирование: изменение кейса создает версию**
- Предусловия: существует тест-кейс с version=1
- Шаги:
  1. `PUT /api/projects/:projectId/tms/cases/:caseKey` с `{"title": "Новое название", "version": 1}`
  2. `GET /api/projects/:projectId/tms/cases/:caseKey/versions`
- Ожидаемый результат: тест-кейс имеет `version: 2`, список версий содержит 1 элемент с `version_number: 1` и полным снимком предыдущего состояния

**TC-TMS-02-04: Откат тест-кейса к предыдущей версии**
- Предусловия: тест-кейс с version=2, существует версия с version_number=1
- Шаги:
  1. `POST /api/projects/:projectId/tms/cases/:caseKey/rollback/:versionId`
- Ожидаемый результат: HTTP 200, тест-кейс восстановлен из снимка версии 1, `version` увеличен (version=3)

---

### TMS-03: Связь с GitHub

**TC-TMS-03-01: Связь тест-кейса с GitHub через external_id (@TestCase аннотацию)**
- Предусловия: тест-кейс создан
- Шаги:
  1. `PUT /api/projects/:projectId/tms/cases/:caseKey` с `{"external_id": "HUB-TC-42", "version": 1}`
  2. `GET /api/projects/:projectId/tms/cases/:caseKey`
- Ожидаемый результат: поле `external_id` равно "HUB-TC-42"; при импорте Allure с аннотацией `@TestCase("HUB-TC-42")` результат привяжется к этому кейсу

---

### TMS-04: Прогоны тестов

**TC-TMS-04-01: Создание ручного тест-рана и добавление кейсов**
- Предусловия: существуют тест-кейсы в проекте
- Шаги:
  1. `POST /api/projects/:projectId/tms/runs` с `{"title": "Regression v2.1", "run_type": "manual"}`
  2. `POST /api/projects/:projectId/tms/runs/:runId/cases` с `{"case_ids": ["id1", "id2"]}`
  3. `GET /api/projects/:projectId/tms/runs/:runId`
- Ожидаемый результат: прогон создан со статусом `created`, содержит 2 результата со статусом `unknown`

**TC-TMS-04-02: Выполнение теста пошагово (passed/failed/skipped)**
- Предусловия: прогон в статусе `in_progress` с тест-кейсами
- Шаги:
  1. `POST /api/projects/:projectId/tms/runs/:runId/start`
  2. `PUT /api/projects/:projectId/tms/runs/:runId/results/:resultId` с `{"status": "passed", "comment": "Все шаги пройдены", "step_results": [{"step_order": 1, "status": "passed"}]}`
  3. `GET /api/projects/:projectId/tms/runs/:runId`
- Ожидаемый результат: результат обновлён, счётчик `passed_count` увеличился на 1

**TC-TMS-04-03: Переоткрытие завершённого прогона**
- Предусловия: прогон в статусе `completed`
- Шаги:
  1. `POST /api/projects/:projectId/tms/runs/:runId/reopen`
  2. `GET /api/projects/:projectId/tms/runs/:runId`
- Ожидаемый результат: статус прогона = `in_progress`, можно обновлять результаты

**TC-TMS-04-04: Назначение тестировщиков на прогон**
- Предусловия: прогон создан, пользователи существуют
- Шаги:
  1. `POST /api/projects/:projectId/tms/runs/:runId/assignees` с `{"user_ids": ["user1", "user2"]}`
  2. `GET /api/projects/:projectId/tms/runs/:runId`
- Ожидаемый результат: поле `assignees` содержит 2 пользователя

---

### TMS-05: Скриншот и комментарий к результату

**TC-TMS-05-01: Добавление скриншота к результату шага**
- Предусловия: результат теста существует, файл загружен в Storage
- Шаги:
  1. `PUT /api/projects/:projectId/tms/runs/:runId/results/:resultId` с `{"step_results": [{"step_order": 1, "status": "failed", "comment": "Кнопка не видна", "file_ids": ["<file_id>"]}]}`
- Ожидаемый результат: HTTP 200, step_results содержит file_ids с привязанным скриншотом

**TC-TMS-05-02: Добавление комментария к результату шага**
- Предусловия: результат теста существует
- Шаги:
  1. `PUT /api/projects/:projectId/tms/runs/:runId/results/:resultId` с `{"comment": "Баг в верстке, кнопка перекрыта", "step_results": [{"step_order": 1, "status": "failed", "comment": "Элемент отсутствует на странице"}]}`
- Ожидаемый результат: HTTP 200, общий `comment` и пошаговый `comment` сохранены

---

### TMS-06: Связь с трекером

**TC-TMS-06-01: Двусторонняя связь тест-кейса с задачей**
- Предусловия: существуют тест-кейс и задача (Issue) в проекте
- Шаги:
  1. `POST /api/projects/:projectId/tms/cases/:caseKey/links` с `{"issue_id": "<issue_id>"}`
  2. `GET /api/projects/:projectId/tms/cases/:caseKey/links`
- Ожидаемый результат: список содержит привязанную задачу с `issue_key` и `title`; NATS-событие `tms.testcase.linked` опубликовано

**TC-TMS-06-02: Создание бага из упавшего теста**
- Предусловия: прогон содержит результат со статусом `failed`
- Шаги:
  1. `POST /api/projects/:projectId/tms/runs/:runId/results/:resultId/bug` с `{"priority": "high"}`
- Ожидаемый результат: HTTP 201, создана задача типа `bug` с заголовком "FAILED: {название тест-кейса}", задача привязана к тест-кейсу

**TC-TMS-06-03: Отчет покрытия (матрица трассируемости)**
- Предусловия: тест-кейсы привязаны к задачам, прогоны выполнены
- Шаги:
  1. `GET /api/projects/:projectId/tms/reports/coverage`
- Ожидаемый результат: ответ содержит список задач, для каждой -- привязанные тест-кейсы с `last_status` последнего результата

---

### TMS-08: Отчёты

**TC-TMS-08-01: Отчет трендов (pass/fail по прогонам)**
- Предусловия: выполнено 3+ прогонов
- Шаги:
  1. `GET /api/projects/:projectId/tms/reports/trends?limit=10`
- Ожидаемый результат: массив объектов TrendPoint с полями `run_id`, `title`, `total`, `passed`, `failed`, `skipped`, `broken`, `pass_rate`, `created_at`; отсортирован по дате

**TC-TMS-08-02: Определение нестабильных (flaky) тестов**
- Предусловия: тест-кейс участвовал в 10+ прогонах с чередующимися статусами
- Шаги:
  1. `GET /api/projects/:projectId/tms/reports/flaky`
- Ожидаемый результат: список тест-кейсов с `is_flaky: true`, отсортированный по степени нестабильности

**TC-TMS-08-03: Сравнение двух прогонов**
- Предусловия: два прогона с общими тест-кейсами, но разными результатами
- Шаги:
  1. `GET /api/projects/:projectId/tms/reports/comparison?run1=<id1>&run2=<id2>`
- Ожидаемый результат: объект RunComparison с `run1` и `run2` (summary), массив `changes` с полями `case_key`, `case_title`, `status_run1`, `status_run2` для каждого изменения

---

### TMS-09: Импорт тест-кейсов

**TC-TMS-09-01: Импорт тест-кейсов из CSV**
- Предусловия: CSV-файл с заголовками "Title,Description,Priority,Steps"
- Шаги:
  1. `POST /api/projects/:projectId/tms/import` multipart: `file=test-cases.csv`, `format=csv`
- Ожидаемый результат: HTTP 200, `{"created": N, "skipped": 0, "errors": 0}`, тест-кейсы доступны через `GET /tms/cases`

**TC-TMS-09-02: Импорт тест-кейсов из Excel с маппингом полей**
- Предусловия: XLSX-файл с русскими заголовками "Название,Описание,Приоритет"
- Шаги:
  1. `POST /api/projects/:projectId/tms/import` multipart: `file=test-cases.xlsx`, `format=excel`
- Ожидаемый результат: HTTP 200, маппинг определён автоматически, тест-кейсы созданы

**TC-TMS-09-03: Импорт тест-кейсов из JSON и XML**
- Предусловия: JSON-массив `[{"title": "T1", "priority": "high"}]` и XML `<testcases><testcase><title>T1</title></testcase></testcases>`
- Шаги:
  1. `POST /api/projects/:projectId/tms/import` multipart: `file=cases.json`, `format=json`
  2. `POST /api/projects/:projectId/tms/import` multipart: `file=cases.xml`, `format=xml`
- Ожидаемый результат: HTTP 200 для обоих, тест-кейсы созданы

---

### TMS-10: Экспорт

**TC-TMS-10-01: Экспорт тест-кейсов в CSV/JSON/XML/Excel**
- Предусловия: проект содержит тест-кейсы
- Шаги:
  1. `GET /api/projects/:projectId/tms/export?format=csv`
  2. `GET /api/projects/:projectId/tms/export?format=json`
  3. `GET /api/projects/:projectId/tms/export?format=xml`
  4. `GET /api/projects/:projectId/tms/export?format=excel`
- Ожидаемый результат: HTTP 200, Content-Type соответствует формату, Content-Disposition содержит имя файла, тело содержит все тест-кейсы проекта

**TC-TMS-10-02: Экспорт отчета прогона в PDF**
- Предусловия: прогон в статусе `completed` с результатами
- Шаги:
  1. `GET /api/projects/:projectId/tms/runs/:runId/export?format=pdf`
- Ожидаемый результат: HTTP 200, Content-Type: application/pdf, PDF содержит заголовок с названием прогона, таблицу статистики и таблицу результатов

---

### ALLU-01: Загрузка результатов Allure

**TC-ALLU-01-01: Загрузка allure-results JSON через API**
- Предусловия: JSON-массив результатов Allure с валидными данными
- Шаги:
  1. `POST /api/projects/:projectId/tms/allure/import` с телом `{"results": [{...}], "title": "CI Run #42"}`
- Ожидаемый результат: HTTP 200, создан прогон типа `automatic` с результатами, сырой JSON сохранён в MongoDB

**TC-ALLU-01-02: Автоматическое создание прогона из Allure данных**
- Предусловия: JSON-массив с 5 результатами
- Шаги:
  1. `POST /api/projects/:projectId/tms/allure/import` с `{"results": [...5 items...], "title": "Auto Run"}`
  2. `GET /api/projects/:projectId/tms/runs`
- Ожидаемый результат: в списке прогонов появился "Auto Run" с `run_type: "automatic"`, `total_count: 5`, статус `completed`

---

### ALLU-02: Сопоставление результатов

**TC-ALLU-02-01: Сопоставление по аннотации @TestCase (приоритет)**
- Предусловия: тест-кейс с `external_id: "TC-42"`, Allure результат с `links: [{"type": "tms", "name": "TC-42"}]`
- Шаги:
  1. Импортировать Allure-результат
- Ожидаемый результат: результат привязан к тест-кейсу, `match_strategy: "annotation"`

**TC-ALLU-02-02: Сопоставление по fullName (fallback)**
- Предусловия: тест-кейс с `full_name: "com.app.LoginTest.testLogin"`, результат с `fullName: "com.app.LoginTest.testLogin"`, без tms-аннотации
- Шаги:
  1. Импортировать Allure-результат
- Ожидаемый результат: `match_strategy: "fullname"`

**TC-ALLU-02-03: Сопоставление по suite+name (fallback)**
- Предусловия: тест-кейс с `class_name: "LoginTest"`, `method_name: "testLogin"`, результат с `labels: [{"name": "suite", "value": "LoginTest"}]`, `name: "testLogin"`, без fullName
- Шаги:
  1. Импортировать Allure-результат
- Ожидаемый результат: `match_strategy: "classmethod"`

**TC-ALLU-02-04: Несопоставленные результаты отображаются как "unmapped"**
- Предусловия: Allure результат без links, labels, fullName, которые бы совпали с тест-кейсами
- Шаги:
  1. Импортировать Allure-результат
- Ожидаемый результат: `match_strategy: "unmapped"`, `test_case_key` пустой

---

### ALLU-03: История и уведомления

**TC-ALLU-03-01: История прогонов с статистикой**
- Предусловия: выполнено 3+ прогонов (ручных и автоматических)
- Шаги:
  1. `GET /api/projects/:projectId/tms/reports/history`
- Ожидаемый результат: массив TestRunSummary с `run_id`, `title`, `total`, `passed`, `failed`, `pass_rate`, `created_at`, отсортированный по дате убывания

**TC-ALLU-03-02: NATS событие при завершении прогона (для Telegram)**
- Предусловия: прогон в статусе `in_progress`
- Шаги:
  1. `POST /api/projects/:projectId/tms/runs/:runId/complete`
  2. Проверить NATS-подписку на `tms.testrun.completed`
- Ожидаемый результат: опубликовано событие TestRunCompletedEvent с полями `run_id`, `project_id`, `title`, `total`, `passed`, `failed`, `skipped`, `broken`

---

### IMEX-02: Импорт из Excel

**TC-IMEX-02-01: Импорт тест-кейсов из Excel с автоопределением колонок**
- Предусловия: XLSX-файл с заголовками "Название", "Описание", "Приоритет", "Шаги"
- Шаги:
  1. `POST /api/projects/:projectId/tms/import` multipart: `file=test-cases.xlsx`
  2. Не передавать `mapping` -- маппинг определяется автоматически
- Ожидаемый результат: HTTP 200, тест-кейсы созданы с корректными полями `title`, `description`, `priority`; русские заголовки корректно маппятся на поля TMS

---

## 3. Сводка тест-кейсов

| Требование | Кол-во TC | ID тест-кейсов |
|------------|-----------|----------------|
| TMS-01 | 3 | TC-TMS-01-01, TC-TMS-01-02, TC-TMS-01-03 |
| TMS-02 | 4 | TC-TMS-02-01, TC-TMS-02-02, TC-TMS-02-03, TC-TMS-02-04 |
| TMS-03 | 1 | TC-TMS-03-01 |
| TMS-04 | 4 | TC-TMS-04-01, TC-TMS-04-02, TC-TMS-04-03, TC-TMS-04-04 |
| TMS-05 | 2 | TC-TMS-05-01, TC-TMS-05-02 |
| TMS-06 | 3 | TC-TMS-06-01, TC-TMS-06-02, TC-TMS-06-03 |
| TMS-08 | 3 | TC-TMS-08-01, TC-TMS-08-02, TC-TMS-08-03 |
| TMS-09 | 3 | TC-TMS-09-01, TC-TMS-09-02, TC-TMS-09-03 |
| TMS-10 | 2 | TC-TMS-10-01, TC-TMS-10-02 |
| ALLU-01 | 2 | TC-ALLU-01-01, TC-ALLU-01-02 |
| ALLU-02 | 4 | TC-ALLU-02-01, TC-ALLU-02-02, TC-ALLU-02-03, TC-ALLU-02-04 |
| ALLU-03 | 2 | TC-ALLU-03-01, TC-ALLU-03-02 |
| IMEX-02 | 1 | TC-IMEX-02-01 |
| **Итого** | **34** | |

Все 13 требований покрыты 34 тест-кейсами с форматом TC-{REQ_ID}-{NN}.
