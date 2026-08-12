# A3 W2 агент 2 — tms services (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`)  
> Baseline: [A2W2-2-tms-services.md](A2W2-2-tms-services.md) · Stage A: [A2-2-tms-services.md](A2-2-tms-services.md)  
> Remediation: TASK-091 (batch reports) · TASK-111 (tx/import/events/API/tests)  
> Периметр: 22 файла (8 prod + 14 tests) · Прочитано: 22 · Пропущено: 0

## Сверка с A2 (A2W2-2-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W2-2-001 | FIXED-OK | закрыт | Coverage/flaky/traceability/defects → `GetLastResultsForTestCases` / `ListTestResultsByRunIDs` (`service_reports.go:28-38`, `:270-275`, `:368-371`, `:407-412`); TASK-091 |
| A2W2-2-002 | FIXED-OK | закрыт | `snapshotAndUpdateTestCase` → Ent `UpdateTestCaseWithSnapshot` одна tx (`ent_repository_tx.go:40-154`); TASK-111 |
| A2W2-2-003 | FIXED-OK | закрыт | Rollback: `errors.As(AppError)` → 409 (`service_test_cases.go:275-280`) |
| A2W2-2-004 | FIXED-OK | закрыт | `CreateAutomaticImportRun` Ent tx + mock abort-cleanup (`service_import.go:58-88`) |
| A2W2-2-005 | FIXED-OK | закрыт | `maxEpochMs` в `buildAllureResultInputs` для обоих import-путей (`:159-170`) |
| A2W2-2-006 | FIXED-OK | закрыт | CreateBug: `NewInternal("ошибка создания бага")` без `err.Error()` (`:420-422`) |
| A2W2-2-007 | FIXED-OK | закрыт | Publish `EventTestRunCompleted` (`events.go:10-11`; `service_runs.go:359`; `service_import.go:97`) |
| A2W2-2-008 | FIXED-OK | закрыт | `UpdateTestResult(ctx, projectID, runID, resultID, …)` + `GetTestRunInProject` (`service_runs.go:169-194`) |
| A2W2-2-009 | FIXED-OK | закрыт | mock `ListTestRuns` пагинирует; assert `Len==1`; Ent Limit=1 (`service_cases_extra_test.go:163-166`; TASK-111) |
| A2W2-2-010 | FIXED-OK | закрыт | `buildSnapshot` возвращает err Marshal/Unmarshal (`service_converters.go:305-326`) |
| A2W2-2-011 | FIXED-OK | закрыт | фиксированное «некорректный allure JSON» + slog (`service_import.go:35-37`, `:118-120`) |
| W2-A2-001…005,019 partial,020 | FIXED-OK | без регрессии | suites/components IDOR, ParentID/labels/components scope, env InProject, rune-safe normalize, slog.Warn на batch fail |

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/tms/service.go` | 64 | 0 | чисто |
| `backend/internal/tms/service_converters.go` | 398 | 1 | snapshotToUpdateInput silent steps |
| `backend/internal/tms/service_import.go` | 226 | 1 | ExistingRun без PG-tx |
| `backend/internal/tms/service_reports.go` | 538 | 0 | FIXED-OK batch N+1 |
| `backend/internal/tms/service_runs.go` | 399 | 2 | AssignTesters membership; FileIDs scope |
| `backend/internal/tms/service_suites.go` | 120 | 0 | FIXED-OK IDOR+ParentID |
| `backend/internal/tms/service_test_cases.go` | 430 | 1 | raw err Link/Delete |
| `backend/internal/tms/events.go` | 56 | 0 | FIXED-OK subject canon |
| `backend/testframework/tms/service_cases_extra_test.go` | 167 | 0 | FIXED-OK pagination |
| `backend/testframework/tms/service_converters_test.go` | 209 | 0 | чисто |
| `backend/testframework/tms/service_reports_test.go` | 265 | 0 | чисто |
| `backend/testframework/tms/service_runs_extra_test.go` | 155 | 0 | happy assign (без membership denial) |
| `backend/testframework/tms/service_suites_test.go` | 96 | 0 | чисто |
| `backend/testframework/tms/service_test.go` | 1368 | 1 | RESIDUAL mock UpdateTestResult |
| `backend/testframework/tms/service_wave113_test.go` | 169 | 0 | чисто |
| `backend/testframework/tms/service_wave114_test.go` | 96 | 0 | чисто |
| `backend/testframework/tms/service_wave115_test.go` | 67 | 0 | чисто |
| `backend/testframework/tms/service_wave116_test.go` | 47 | 0 | чисто |
| `backend/testframework/tms/service_wave120_test.go` | 66 | 0 | чисто |
| `backend/testframework/tms/service_wave122_extra_test.go` | 292 | 0 | чисто (зависит от слабого mock) |
| `backend/testframework/tms/service_wave124_test.go` | 163 | 0 | чисто |
| `backend/testframework/tms/create_bug_nil_tc_a3009_test.go` | 21 | 0 | чисто |

Файлов в списке: 22. Прочитано: 22. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W2-2-001 | FIXED | P2 | `service_runs.go` + `CountProjectMembers` | `AssignTestersToRun`: membership ∈ project | non-member → 400 | TASK-139 · `a3w2_2_tms_services_task139_test.go` |
| A3W2-2-002 | FIXED | P2 | `service_runs.go` + `CountFilesInProject` | `ExecuteTestResult` FileIDs project-scope | foreign FileID → 404 | TASK-139 · Ent+mock denial |
| A3W2-2-003 | FIXED | P2 | `ImportResultsToExistingRun` tx + mock compensating delete | ExistingRun import без half-written | inject fail → cleanup | TASK-139 |
| A3W2-2-004 | FIXED | P2 | mock `UpdateTestResult` | require Edges.Run / runResults | без scope → NotFound | TASK-139 (mock guard + тест) |
| A3W2-2-005 | FIXED | P3 | `snapshotToUpdateInput` | Marshal/Unmarshal steps fail-closed | corrupt steps → error | TASK-139 |
| A3W2-2-006 | FIXED | P3 | `Delete`/`Link`/`Unlink` | wrap → AppError | raw Ent text не утекает | TASK-139 |

## Чисто

- Suite/component IDOR + ParentID/suite/label/component project-scope на Create/UpdateTestCase — без регрессии.
- Reports N+1 batch (TASK-091) + slog.Warn при batch fail — FIXED-OK.
- Snapshot+update Ent tx, Rollback 409, ImportAllure create-path tx, epoch guard, Event subject, UpdateTestResult(project,run,result), CreateBug fixed msg, pagination mock+Ent — FIXED-OK (TASK-111).
- A3-003/004/007/008/009/012/017 service-path invariants держатся; nil-TC CreateBug + Ent mutation hint живы.
- FSM Start/Complete/Reopen; NATS nil-safe PublishJSON.

### Сводка counts (после TASK-139)

| Severity | Count open | FIXED |
|----------|------:|------:|
| P0 | 0 | — |
| P1 | 0 | — |
| P2 | 0 | 4 |
| P3 | 0 | 2 |
| **Total open** | **0** | **6** |

A3 active → **6/6 FIXED** (TASK-139, be2, 2026-07-29). Mutation red/green 001–006 PASS.
