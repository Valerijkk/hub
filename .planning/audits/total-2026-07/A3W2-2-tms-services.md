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
| A3W2-2-001 | NEW | P2 | `service_runs.go:312-325` | `AssignTestersToRun`: `userIDs` только parse UUID; нет membership ∈ project | Repo `AddAssigneeIDs` без scope (`ent_repository.go:1178-1183`) → любой существующий user UUID в org можно назначить на прогон проекта A (зеркало tracker assignee / TASK-088) | Перед assign: `IsProjectMember` / count membership; mismatch → 404/400; Ent denial-тест |
| A3W2-2-002 | NEW | P2 | `service_runs.go:286-294` | `ExecuteTestResult`: `FileIDs` → `AddFilesToResult` без project-scope файлов | `AddFileIDs` без `HasProjectWith` (`ent_repository.go:1222-1236`) → M2M к чужому file UUID; detail отдаёт metadata чужого файла | Count/Find files in project; mismatch → 404; иначе не Add |
| A3W2-2-003 | NEW | P2 | `service_import.go:126-139` | `ImportAllureToExistingRun`: Mongo dump → `BatchCreateTestResults` → `recomputeRunStats` без общей Ent tx / abort | Sibling A2W2-2-004 закрыл только create-path; fail batch после Mongo / fail recompute → частичные results + stale stats | Batch+stats в tx; при fail — не оставлять half-written results (или compensating delete) |
| A3W2-2-004 | RESIDUAL (W2-A2-016) | P2 | `service_test.go:522-541` | mock `UpdateTestResult`: если `Edges.Run==nil` и `runResults[runID]` отсутствует — membership не проверяется | Wave122 Execute happy-path зелёный без `HasRunWith`; регрессия A3-008 на mock-слое не ловится (prod Ent покрыт `result_scope_a3008_*`) | Требовать `Edges.Run.ID==runID` **или** entry в `runResults`; иначе NotFound |
| A3W2-2-005 | NEW | P3 | `service_converters.go:367-372` | `snapshotToUpdateInput`: `_ = json.Marshal` + ignore Unmarshal err для steps | `buildSnapshot` (A2W2-2-010) уже fail-closed; rollback path молча теряет steps → откат без шагов | Проверять err; при fail — return error из Rollback / не применять битый snapshot |
| A3W2-2-006 | NEW | P3 | `service_test_cases.go:243`, `:333`, `:352` | `DeleteTestCase` / `Link` / `Unlink` возвращают сырой `repo` err | Non-AppError (Ent/constraint) может утечь наверх без `{"error","request_id"}` контракта | Wrap → AppError; AppError пробрасывать |

## Чисто

- Suite/component IDOR + ParentID/suite/label/component project-scope на Create/UpdateTestCase — без регрессии.
- Reports N+1 batch (TASK-091) + slog.Warn при batch fail — FIXED-OK.
- Snapshot+update Ent tx, Rollback 409, ImportAllure create-path tx, epoch guard, Event subject, UpdateTestResult(project,run,result), CreateBug fixed msg, pagination mock+Ent — FIXED-OK (TASK-111).
- A3-003/004/007/008/009/012/017 service-path invariants держатся; nil-TC CreateBug + Ent mutation hint живы.
- FSM Start/Complete/Reopen; NATS nil-safe PublishJSON.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 4 |
| P3 | 2 |
| **Total** | **6** |

A2 baseline open: 11 → FIXED-OK 11 · REGRESS 0 · RESIDUAL 1 (W2-A2-016) · NEW 5.
