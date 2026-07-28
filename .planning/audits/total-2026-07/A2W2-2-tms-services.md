# A2 · W2 агент A2 — tms services (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: [A2-2-tms-services.md](A2-2-tms-services.md) · TASK-015 W18 PASS

## Прочитанные файлы

| Файл | Строк | Open | Вердикт |
|------|------:|-----:|---------|
| `backend/internal/tms/service.go` | 57 | 0 | чисто |
| `backend/internal/tms/service_converters.go` | 366 | 1 | RESIDUAL silent Marshal |
| `backend/internal/tms/service_import.go` | 199 | 3 | RESIDUAL tx/epoch/err leak |
| `backend/internal/tms/service_reports.go` | 458 | 3 | RESIDUAL N+1 |
| `backend/internal/tms/service_runs.go` | 342 | 1 | RESIDUAL UpdateTestResult API |
| `backend/internal/tms/service_suites.go` | 104 | 0 | FIXED-OK IDOR+ParentID |
| `backend/internal/tms/service_test_cases.go` | 387 | 3 | RESIDUAL tx/rollback/err |
| `backend/internal/tms/events.go` | 50 | 1 | RESIDUAL subject drift |
| `backend/testframework/tms/service_cases_extra_test.go` | 140 | 1 | RESIDUAL pagination mock |
| `backend/testframework/tms/service_converters_test.go` | 189 | 0 | чисто |
| `backend/testframework/tms/service_reports_test.go` | 228 | 0 | чисто |
| `backend/testframework/tms/service_runs_extra_test.go` | 132 | 0 | чисто |
| `backend/testframework/tms/service_suites_test.go` | 82 | 0 | happy; Ent-neg отдельно |
| `backend/testframework/tms/service_test.go` | 1117 | 0 | mock updated signatures |
| `backend/testframework/tms/service_wave113_test.go` | 146 | 0 | чисто |
| `backend/testframework/tms/service_wave114_test.go` | 81 | 0 | чисто |
| `backend/testframework/tms/service_wave115_test.go` | 59 | 0 | чисто |
| `backend/testframework/tms/service_wave116_test.go` | 41 | 0 | чисто |
| `backend/testframework/tms/service_wave120_test.go` | 55 | 0 | чисто |
| `backend/testframework/tms/service_wave122_extra_test.go` | 245 | 0 | чисто |
| `backend/testframework/tms/service_wave124_test.go` | 135 | 0 | чисто |
| `backend/testframework/tms/create_bug_nil_tc_a3009_test.go` | 17 | 0 | чисто |

Файлов: **22/22/0**.

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| W2-A2-001/002 | `service_suites.go:55–119` — `projectID` + NotFound |
| W2-A2-003 | `service_reports.go:426–431` — `GetEnvironmentInProject` |
| W2-A2-004 | `Create/UpdateTestCase` + `CreateSuite`: `GetSuiteInProject` / `CountLabelsInProject` / `CountComponentsInProject` |
| W2-A2-005 | `CreateTestResult` / `AddTestCasesToRun`: `CountTestCasesInProject` |
| W2-A2-019 (partial) | `normalizeErrorPattern` — rune-safe (`service_converters.go:385–390`) |
| W2-A2-020 | flaky/defect: `slog.WarnContext` перед `continue` |

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему | Фикс |
|----|-----|-----|-------------|----------|--------|------|
| A2W2-2-001 | RESIDUAL (W2-A2-006/007/008) | P2 | `service_reports.go:42`, `:259`, `:351`, `:378` | Coverage/flaky/traceability/defects: `GetLastResult*` / `ListTestResults` per item/run | N+1 на горячих отчётах; batch API нет | `Where(HasRunWith(IDIn))` / batch last-result |
| A2W2-2-002 | RESIDUAL (W2-A2-009) | P2 | `service_test_cases.go:147–222` | Snapshot `CreateTestCaseVersion` до `UpdateTestCase` вне одной tx | Fail update → orphan version | Одна Ent tx |
| A2W2-2-003 | RESIDUAL (W2-A2-010) | P2 | `service_test_cases.go:275–278` | Rollback: любой update err → Internal (нет `errors.As` AppError) | 409 VersionEQ маскируется 500 | Как в `UpdateTestCase:224–228` |
| A2W2-2-004 | RESIDUAL (W2-A2-011) | P2 | `service_import.go:47–125` | ImportAllure: create run → mongo → batch → status без общей tx | Orphan run / частичные results | Ent tx на PG-часть + cleanup |
| A2W2-2-005 | RESIDUAL (W2-A2-012) | P2 | `service_import.go:177–183` | `ImportAllureToExistingRun` пишет duration без `maxEpochMs` guard | Мусорные timestamps | Скопировать guard из `:72–84` |
| A2W2-2-006 | RESIDUAL (W2-A2-013) | P2 | `service_test_cases.go:419` | `"ошибка создания бага: " + err.Error()` | Сырой Ent/DB клиенту | `NewInternal("ошибка создания бага")` |
| A2W2-2-007 | RESIDUAL (W2-A2-014) | P2 | `events.go:9–11` vs `SubjectTestRunCompleted` | Константы `tms.testrun.*` не используются; publish `tms.run.completed.%s` | Drift NATS/notification | Единый subject через Event* |
| A2W2-2-008 | RESIDUAL (W2-A2-015) | P2 | `service_runs.go:170–198` | `UpdateTestResult(ctx, resultID, …)` без project/run в сигнатуре | Слабый service API (HTTP идёт через Execute) | Требовать `(projectID, runID, resultID)` |
| A2W2-2-009 | RESIDUAL (W2-A2-017) | P2 | `service_cases_extra_test.go` | Pagination assert на mock, игнорирующем Limit | Тест зелёный при сломанной пагинации | Ent Limit=1 → Len==1 |
| A2W2-2-010 | RESIDUAL (W2-A2-019) | P3 | `service_converters.go:315–317` | `buildSnapshot`: `_ = json.Marshal/Unmarshal` steps | Тихий битый snapshot | Проверять err |
| A2W2-2-011 | NEW | P3 | `service_import.go:30`, `:153` | `"некорректный allure JSON: " + err.Error()` | Internals парсера в 400 | Фиксированное сообщение + slog |

## Сводка

| | Count |
|--|------:|
| Files | 22/22/0 |
| FIXED-OK classes | 6 |
| NEW | 1 |
| REGRESS | 0 |
| RESIDUAL | 10 |
| Open P0/P1/P2/P3 | 0 / 0 / 9 / 2 |
