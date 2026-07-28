# W2 агент A1 — tms handlers

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tms/handler.go | 134 | 0 | чисто |
| backend/internal/tms/handler_import.go | 170 | 0 | чисто |
| backend/internal/tms/handler_reports.go | 235 | 0 | чисто |
| backend/internal/tms/handler_runs.go | 419 | 0 | чисто |
| backend/internal/tms/handler_suites.go | 182 | 3 | замечания |
| backend/internal/tms/handler_test_cases.go | 367 | 1 | замечания |
| backend/internal/tms/handler_webhook.go | 245 | 2 | замечания |
| backend/internal/tms/dto.go | 470 | 1 | замечания |
| backend/testframework/tms/handler_flow_test.go | 665 | 1 | замечания |
| backend/testframework/tms/handler_import_flow_test.go | 174 | 0 | чисто |
| backend/testframework/tms/handler_routes_test.go | 214 | 0 | чисто |
| backend/testframework/tms/handler_wave108_test.go | 76 | 1 | замечания |
| backend/testframework/tms/handler_wave110_test.go | 134 | 0 | чисто |
| backend/testframework/tms/handler_wave113_test.go | 75 | 0 | чисто |
| backend/testframework/tms/handler_wave114_test.go | 40 | 0 | чисто |
| backend/testframework/tms/handler_wave115_test.go | 48 | 0 | чисто |
| backend/testframework/tms/handler_wave120_test.go | 91 | 0 | чисто |
| backend/testframework/tms/handler_wave122_test.go | 99 | 0 | чисто |
| backend/testframework/tms/handler_wave124_test.go | 54 | 1 | замечания |
| backend/testframework/tms/handler_wave88_test.go | 150 | 1 | замечания |
| backend/testframework/tms/handler_webhook_flow_test.go | 110 | 0 | чисто |
| backend/testframework/tms/handler_webhook_rbac_test.go | 172 | 0 | чисто |
| backend/testframework/tms/handler_webhook_wave99_test.go | 31 | 0 | чисто |
| backend/testframework/tms/helpers_test.go | 132 | 0 | чисто |

Файлов в списке: 24. Прочитано: 24. Пропущено: 0.

## Перепроверка прошлого аудита (A3 / grok-2026-07)

| Старый ID | Статус в текущем коде | Комментарий |
|-----------|----------------------|-------------|
| A3-003 | **исправлено** — не дублировать | Handlers передают `projectID`; `FindTestCaseByKey(ctx, projectID, caseKey)` + `HasProjectWith` (`ent_repository.go:161-174`) |
| A3-007 | **исправлено** — не дублировать | `GetTestCaseVersion(ctx, existing.ID, versionID)` + `HasTestCaseWith` (`ent_repository.go:369-378`); service комментарий A3-007 |
| A3-008 | **частично** | `ExecuteTestResult` + `UpdateTestResult(runID, resultID)` scoped — OK. `CreateBugFromResult` **не парсит `:runId`** → W2-A1-004 (остаток на handler) |
| A3-010 | **исправлено** — не дублировать | `allure_matcher.go:139` вызывает `FindTestCaseByKey(ctx, projectID, key)` |
| A3-012 | **исправлено** — не дублировать | Handler передаёт `projectID`; service `GetTestRunInProject` (`service_import.go:140-148`) |
| A3-004 | вне handler-scope, **исправлено в repo** | `VersionEQ` в `UpdateTestCase` (`ent_repository.go:262-268`) — для A2 services/repos |

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W2-A1-001 | P1 | `backend/internal/tms/handler_suites.go:72-88` | `UpdateSuite` не парсит `:projectId`, вызывает `UpdateSuite(ctx, suiteID, req)` | RBAC смотрит только membership URL-проекта. Зная UUID папки чужого проекта, member с `PermTestCaseEdit` переименует её (`repo.UpdateSuite` → `UpdateOneID` без project scope) | Парсить `projectID`; `UpdateSuite(ctx, projectID, suiteID, req)` + `Where(ID, HasProjectWith(...))`; mismatch → 404 |
| W2-A1-002 | P1 | `backend/internal/tms/handler_suites.go:98-108` | `DeleteSuite` — только `suiteID`, без project scope | Hard-delete чужой папки по UUID (`DeleteOneID`) | Scope delete: `HasProjectWith(projectID)`; 404 при mismatch |
| W2-A1-003 | P1 | `backend/internal/tms/handler_suites.go:171-181` | `DeleteComponent` — только `componentID` | Hard-delete чужого компонента по UUID | Аналогично: delete scoped по проекту |
| W2-A1-004 | P2 | `backend/internal/tms/handler_test_cases.go:342-361` | `CreateBugFromResult` парсит `resultId`, **игнорирует `:runId`**; service `CreateBugFromFailedResult(..., resultID)` без `runID` | Остаток A3-008: URL-прогон и результат не связаны. В рамках проекта баг создаётся по любому `resultId` при произвольном `runId` в пути (контракт/целостность; cross-project закрыт через `GetTestRunInProject` на run результата) | Парсить `runID`; `CreateBugFromFailedResult(ctx, projectID, runID, userID, resultID, req)` + `Where(result.ID, HasRunWith(IDEQ(runID)))` |
| W2-A1-005 | P2 | `backend/internal/tms/dto.go:14-133` (+ handlers `Bind().JSON` без validator) | Теги `validate:"required,…"` на DTO не исполняются | Пустые title/status/version принимаются. Доказательство: `handler_wave124_test.go:34` шлёт update без `version` и ждёт успех на mock | Fiber StructValidator или `validate.Struct` в handlers; отсекать нулевой `Version` в service |
| W2-A1-006 | P2 | `backend/internal/tms/handler_webhook.go:106-111` (+ swagger:39-41) | Swagger обещает multipart **zip** allure-results; код всегда `json.Unmarshal` тела/файла | Клиент CI, отправивший zip, получает 400 «невалидный JSON» — сломан заявленный контракт | Либо распаковать zip и собрать results JSON, либо убрать zip из swagger/описания |
| W2-A1-007 | P2 | `backend/testframework/tms/handler_flow_test.go:26-90` (+ `269-321`) | Handler-flow на `mockRepo`; suite/component update/delete без project ownership | Тесты зелёные при живом IDOR (W2-A1-001…003); нет negative «чужой projectId + чужой suiteId → 404» | Ent/sqlite + cross-project denial (как `handler_webhook_rbac_test`); mock — только HTTP shape |
| W2-A1-008 | P2 | `backend/testframework/tms/handler_wave108_test.go:66-76` | `DeleteComponent_InvalidID` бьёт random UUID и ждёт «не 200» | NotFound на отсутствующий ID ≠ проверка project scope; IDOR не ловится | Foreign-project component delete → 404; assert на scoped repo |
| W2-A1-009 | P2 | `backend/testframework/tms/handler_wave124_test.go:34-38` | PUT case без `version` → ожидается успех (mock) | Закрепляет отсутствие валидации (W2-A1-005); на Ent с `VersionEQ` путь должен быть 400/409 | Требовать `version` в теле; на Ent — conflict без совпадения |
| W2-A1-010 | P2 | `backend/testframework/tms/handler_wave88_test.go:60-105` | `CreateBugFromResult` на mock: run/result согласованы вручную; нет кейса «чужой runId в URL» | Не ловит W2-A1-004 | Assert: result ∈ URL-run + project; mismatch → 404 |
| W2-A1-011 | P3 | `backend/internal/tms/handler_webhook.go:109-110`, `204-205` | В 400 клиенту отдаётся `err.Error()` парсера JSON/XML | Утечка внутренних деталей парсера во внешний API (не `{"error","request_id"}` канон) | Фиксированное сообщение клиенту; детали — в `slog` |

## Чисто

Файлы без находок — что проверялось:

- **handler.go** — Fiber v3 RBAC после handler (FA-02-02); webhook под `PermTestRunCreate`; маршруты cases/runs/reports с permission; `parseCaseKey`; нет `_ = err`. Env-роуты регистрируются здесь, тела — в `environment.go` (список A3 / A2-3) — там же проверить IDOR Update/Delete env (паттерн как suite).
- **handler_import.go** — `ParseProjectID` + `UserID` / `runId`; Bind errors → 400; export через `SendStream`; project+run в ExportTestRunReport.
- **handler_reports.go** — projectID на всех отчётах; UUID parse для run1/run2/env1/env2; limit нормализуется в service (≤100).
- **handler_runs.go** — projectID+runID(+resultID+userID) на CRUD/execute/list/detail/assignees/start/complete; Bind → 400; нет проглоченных error.
- **handler_import_flow_test.go** — invalid body / export formats / allure success; asserts есть.
- **handler_routes_test.go** — RegisterRoutes + Ent sqlite + RBAC owner; lifecycle/export/env list.
- **handler_wave110/113/114/115/120/122** — parse/validation/empty happy-path; asserts есть; не claim’ят project-scope guards.
- **handler_webhook_flow_test.go** / **wave99** — empty/invalid/success multipart missing file.
- **handler_webhook_rbac_test.go** — outsider 403 на webhook/junit/cases/runs (реальный RBAC) — хороший образец.
- **helpers_test.go** — converters/tree/snapshot; asserts на каждом тесте.

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 7 |
| P3 | 1 |
| **Total** | **11** |

**Топ-3:** W2-A1-001/002/003 (IDOR suite update/delete + component delete без project scope) → W2-A1-004 (CreateBug игнорирует `:runId`) → W2-A1-005/007 (мёртвые validate-теги + mock-тесты, не ловящие IDOR).

**Handoff A2/A3:** `service_suites.go` / `ent_repository.go` UpdateSuite/DeleteSuite/DeleteComponent без project; `environment.go` Update/Delete handlers — тот же класс IDOR; CreateBug service signature без `runID`.
