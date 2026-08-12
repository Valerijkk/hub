# W2 агент A2 — tms services

**Дата:** 2026-07-28  
**Режим:** READ-ONLY (кроме этого отчёта)  
**Задача:** TASK-008 / волна W2 / агент A2  
**Перепроверка:** `.planning/audits/grok-2026-07/03-tracking.md` (A3-004/007/008/009/012/017 закрыты на service+repo; A3-014/020 всё ещё открыты)

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/tms/service.go` | 64 | 0 | чисто (FSM прогонов + parseUUIDs) |
| `backend/internal/tms/service_converters.go` | 389 | 1 | byte-truncate / silent json в snapshot |
| `backend/internal/tms/service_import.go` | 230 | 2 | Allure без tx; epoch только в ImportAllureResults |
| `backend/internal/tms/service_reports.go` | 496 | 5 | env IDOR; N+1 flaky/coverage/defects; silent continue |
| `backend/internal/tms/service_runs.go` | 388 | 3 | TC без project; UpdateTestResult unscoped; `_ = result` |
| `backend/internal/tms/service_suites.go` | 114 | 3 | IDOR Update/Delete suite/component; ParentID |
| `backend/internal/tms/service_test_cases.go` | 383 | 4 | refs без project; A3-014; Conflict→Internal; err leak |
| `backend/internal/tms/events.go` | 56 | 1 | subjects ≠ SubjectTestRunCompleted; Event* не публикуются |
| `backend/testframework/tms/service_cases_extra_test.go` | 167 | 1 | pagination assert на сломанном mock Limit |
| `backend/testframework/tms/service_converters_test.go` | 208 | 0 | чисто |
| `backend/testframework/tms/service_reports_test.go` | 265 | 0 | чисто (happy-path; нет denial env IDOR) |
| `backend/testframework/tms/service_runs_extra_test.go` | 155 | 0 | чисто (happy-path start/detail/assign/env) |
| `backend/testframework/tms/service_suites_test.go` | 96 | 0 | чисто (CRUD; **нет** foreign-project denial) |
| `backend/testframework/tms/service_test.go` | 1165 | 1 | mock UpdateTestResult слабый run-guard |
| `backend/testframework/tms/service_wave113_test.go` | 169 | 0 | чисто (validation / NotFound paths) |
| `backend/testframework/tms/service_wave114_test.go` | 96 | 0 | чисто (ExecuteTestResult errors; mock-only) |
| `backend/testframework/tms/service_wave115_test.go` | 67 | 0 | чисто |
| `backend/testframework/tms/service_wave116_test.go` | 47 | 0 | чисто |
| `backend/testframework/tms/service_wave120_test.go` | 66 | 0 | чисто (env update / empty reports) |
| `backend/testframework/tms/service_wave122_extra_test.go` | 289 | 0 | чисто (nil-TC bug path; ImportAllure mock) |
| `backend/testframework/tms/service_wave124_test.go` | 163 | 0 | чисто (Ent rollback/link happy-path) |
| `backend/testframework/tms/create_bug_nil_tc_a3009_test.go` | 21 | 0 | чисто (A3-009 Ent + mutation hint) |

Файлов в списке: 22. Прочитано: 22. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W2-A2-001 | P1 | `service_suites.go:50-57`, `69-70` (+ handler передаёт только `suiteId`) | `UpdateSuite` / `DeleteSuite` без `projectID`; repo `UpdateOneID` / `DeleteOneID` | Member проекта A с `PermTestCaseEdit` меняет/удаляет suite проекта B по UUID (IDOR). Контраст с `CreateSuite(ctx, projectID, …)` / `ListSuites(projectID)` | Сигнатуры `Update/DeleteSuite(ctx, projectID, suiteID, …)`; repo `Where(ID, HasProjectWith(projectID))` → 404 |
| W2-A2-002 | P1 | `service_suites.go:106-107` | `DeleteComponent` только по `componentID` | Тот же класс: чужой component удаляется при RBAC своего `:projectId` | `DeleteComponent(ctx, projectID, id)`; Ent `HasProjectWith` |
| W2-A2-003 | P1 | `service_reports.go:419-426` (+ handler `env-comparison`) | `CompareEnvironments`: `GetEnvironment(envID)` без сверки с `projectID` | Утечка имён/label чужих окружений; далее `GetLatestTestRunForEnvironment(projectID, envID)` даёт пустую карту, но Env1/Env2 в ответе уже раскрыты | `GetEnvironmentInProject(projectID, envID)` или compare `Edges.Project.ID`; mismatch → 404 |
| W2-A2-004 | P1 | `service_suites.go:22-28`; `service_test_cases.go:35-55`, `152-176` | `ParentID` / `SuiteID` / `Labels` / `Components` парсятся, но не проверяются на ∈ `projectID` | Можно привязать TC/suite к папке/метке/компоненту чужого проекта (cross-project graph / data corruption) | Перед create/update: suite/label/component `HasProjectWith(projectID)`; иначе 400/404 |
| W2-A2-005 | P1 | `service_runs.go:147-155`, `228-234` | `CreateTestResult` / `AddTestCasesToRun`: `TestCaseID`/`CaseIDs` без проверки принадлежности проекту | В прогон своего проекта можно воткнуть UUID кейса чужого проекта (если FK позволяет) или получить неочевидный DB error | `FindTestCaseInProject` / `Where(IDIn, HasProjectWith)` для всех ID; иначе 404 |
| W2-A2-006 | P2 | `service_reports.go:256-257` (A3-020 **открыт**) | Flaky: `ListTestResults` на каждый run в окне 5 | N+1 к PG на каждый отчёт flaky | Batch: `Where(HasRunWith(IDIn(...)))` одним запросом |
| W2-A2-007 | P2 | `service_reports.go:40-41`, `346-347` | Coverage / Traceability: `GetLastResultForTestCase` на каждый TC | N+1 при матрице/покрытии с десятками связей | Batch last-result по `TestCaseID IN (...)` или join |
| W2-A2-008 | P2 | `service_reports.go:372-373` | Defect categories: `ListTestResults` на каждый из 10 runs | Тот же N+1-класс, что A3-020 | Один запрос по `run IDs` |
| W2-A2-009 | P2 | `service_test_cases.go:127-132`, `178` (A3-014 **открыт**) | Snapshot `CreateTestCaseVersion` до `UpdateTestCase`; не одна tx | Ошибка update → orphan version; гонка → дубликаты снимков при успешном VersionEQ у одного из writers | Одна tx: snapshot + `VersionEQ` update; rollback всей tx при 409 |
| W2-A2-010 | P2 | `service_test_cases.go:231-234` | `RollbackTestCase`: любой `err` update → Internal; `errors.As(AppError)` нет (в отличие от `UpdateTestCase:179-184`) | Client VersionEQ conflict (409) маскируется как 500 Internal | Как в Update: `errors.As` → проброс AppError |
| W2-A2-011 | P2 | `service_import.go:47-125` | `ImportAllureResults`: create run → mongo → batch results → status без общей tx | Fail на шаге 5–7 оставляет orphan run / частичные results | Ent tx на PG-часть; при fail — abort/cleanup; mongo best-effort ок |
| W2-A2-012 | P2 | `service_import.go:177-183` | `ImportAllureToExistingRun` пишет `DurationMs`/`UnixMilli` без epoch-валидации (есть в `ImportAllureResults:72-84`) | Мусорные start/stop → отрицательная duration / некорректные timestamps | Скопировать тот же `maxEpochMs` guard |
| W2-A2-013 | P2 | `service_test_cases.go:375` | `CreateBugFromFailedResult`: `"ошибка создания бага: " + err.Error()` | Сырой Ent/DB текст может уйти клиенту | `NewInternal("ошибка создания бага")` без wrap message |
| W2-A2-014 | P2 | `events.go:9-11` vs `service_import.go:21` / `service_runs.go:349` | Константы `EventTestRunCompleted="tms.testrun.completed"` не используются; publish идёт в `tms.run.completed.%s` | Два subject в NATS/notification; `EventTestRunCreated` мёртв — drift контракта событий | Единый канон subject; publish через константы; удалить/использовать Event* |
| W2-A2-015 | P2 | `service_runs.go:164-191` | `UpdateTestResult(ctx, resultID, …)` без `projectID`/`runID` в сигнатуре сервиса (HTTP идёт через Execute) | Слабый API: lookup только `GetTestResult`; project-scope нет на service-слое | Убрать или требовать `(projectID, runID, resultID)` + `GetTestRunInProject` |
| W2-A2-016 | P2 | `service_test.go:409-428` | mock `UpdateTestResult`: если `runResults[runID]` нет и `Edges.Run==nil` — membership не проверяется | Service-тесты Execute (wave114/122) зелёные без реального `HasRunWith`; регрессия A3-008 на mock не ловится | Всегда требовать `Edges.Run.ID==runID` **или** membership в `runResults`; иначе NotFound |
| W2-A2-017 | P2 | `service_cases_extra_test.go:163-166` | `TestListTestRuns_Pagination`: `Limit:1` → assert `Len(items)==2` с комментарием «mock ignores» | Тест зелёный при полностью сломанной пагинации сервиса/repo | Ent-тест: Limit=1 → Len==1; mock либо пагинировать, либо не тестировать Limit здесь |
| W2-A2-018 | P3 | `service_runs.go:292` | `_ = result` после успешного Update | Мёртвый код; результат сразу перечитывается | Убрать `_ = result` / использовать updated edges |
| W2-A2-019 | P3 | `service_converters.go:315-317`, `385-387` | `buildSnapshot`: `_ = json.Marshal/Unmarshal`; `normalizeErrorPattern` режет по байтам | Тихий битый snapshot steps; UTF-8 pattern режется mid-rune | Проверять err Marshal; `utf8.RuneCount` / `[]rune` truncate |
| W2-A2-020 | P3 | `service_reports.go:258-259`, `374-375` | Ошибка `ListTestResults` → `continue` без лога | Частичный flaky/defect отчёт без сигнала в ops | `slog.Warn` + metric; опционально fail-closed |

## Чисто

- **A3-003 (caseKey):** все TC lookups через `FindTestCaseByKey(ctx, projectID, caseKey)` (get/update/delete/rollback/versions/links).
- **A3-004 (VersionEQ):** service передаёт `req.Version` → repo `Where(IDEQ, VersionEQ)` + 409 (`ent_repository.go:263-328`); Ent-тест `update_optimistic_a3004_ent_test.go` вне этого списка, но guard не в service.
- **A3-007 (rollback ownership):** `GetTestCaseVersion(ctx, existing.ID, versionID)` + комментарий A3-007; Ent `HasTestCaseWith`.
- **A3-008 (ExecuteTestResult):** `GetTestRunInProject` + `UpdateTestResult(ctx, runID, resultID, …)` с `HasRunWith`; CreateBug проверяет run∈project.
- **A3-009 (nil TC):** service title fallback; repo `if testCaseID != uuid.Nil`; regression `create_bug_nil_tc_a3009_test.go`.
- **A3-012 (run scope):** Get/List/Import/Execute/Detail/Compare используют `GetTestRunInProject` / `*InProject`.
- **A3-017:** `ExecutedByID: &userID` в ExecuteTestResult.
- **Link issue scope:** `FindIssueInProject` на link/unlink.
- **Soft-delete TC:** service → `DeleteTestCase` repo `SetDeletedAt` (не hard).
- **FSM прогонов:** `isValidRunTransition` + Start/Complete/Reopen валидируют статус.
- **NATS nil-safe:** `PublishJSON` допускает nil client (тесты с `nats=nil` ок).
- **Converters:** DTO mapping nil-safe по edges; round-trip snapshot покрыт `service_converters_test.go`.

### Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 5 |
| P2 | 12 |
| P3 | 3 |
| **Всего** | **20** |

**Файлов:** 22/22/0  

**Топ-3:**
1. **W2-A2-001/002** — IDOR Update/Delete Suite и Delete Component без `projectID`
2. **W2-A2-003** — CompareEnvironments читает Environment чужого проекта
3. **W2-A2-004/005** — SuiteID/ParentID/CaseIDs без project-scope (cross-project attach)

**Фокус A3 (статус на service-слое):** A3-004/007/008/009/012/017 — **закрыты**; A3-014/020 — **открыты** (P2).
