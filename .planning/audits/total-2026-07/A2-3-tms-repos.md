# W2 агент A3 — tms repositories

> TASK-008 · стадия A · READ-ONLY · 2026-07-28  
> Фокус: Ent `VersionEQ`, `HasProjectWith` на FindByKey/Update/Delete, Mongo limits, soft delete, tx.  
> Перепроверка grok-2026-07: **A3-003 / A3-004 / A3-007 / A3-008**.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/tms/repository.go` | 321 | 0 | чисто — ISP-контракты; A3-003/007/008/012 в сигнатурах |
| `backend/internal/tms/ent_repository.go` | 1079 | 9 | IDOR suite/component/env; Update без DeletedAt; matchers soft-delete; lists/batch без limit |
| `backend/internal/tms/mongo_repository.go` | 127 | 1 | GetAllureResultsByRunID / Store без Limit/cap |
| `backend/internal/tms/combined_repository.go` | 22 | 0 | чисто — embedding Ent+Mongo |
| `backend/internal/tms/environment.go` | 241 | 1 | Update/Delete handler+service без projectID (IDOR) |
| `backend/internal/tms/provisioner.go` | 27 | 0 | чисто — CreateDefaultSuite через projectID |
| `backend/internal/tms/export_hooks.go` | 33 | 0 | чисто — hubtest re-exports parsers |
| `backend/testframework/tms/ent_repository_batch_test.go` | 122 | 0 | чисто — batch results + ListTestRuns filter + coverage |
| `backend/testframework/tms/ent_repository_extended_test.go` | 263 | 0 | чисто — versions/results/env/link/bug/suite CRUD |
| `backend/testframework/tms/ent_repository_list_update_test.go` | 110 | 0 | чисто — List filters + Update all fields + Version |
| `backend/testframework/tms/ent_repository_test.go` | 185 | 0 | чисто — CRUD + Find* + suite/component + run lifecycle |
| `backend/testframework/tms/ent_repository_wave122_extra_test.go` | 105 | 0 | чисто — AddCases/CreateBug/all fields |
| `backend/testframework/tms/ent_repository_wave122_test.go` | 72 | 0 | чисто — suite move + reopen + stats |
| `backend/testframework/tms/ent_repository_wave124_test.go` | 48 | 0 | чисто — CreateBug + Reopen |
| `backend/testframework/tms/mongo_repository_test.go` | 53 | 0 | чисто — Store/Get lifecycle + indexes (без cap assert) |
| `backend/testframework/tms/postgres_ent_helper_test.go` | 99 | 0 | чисто — Postgres env для VersionEQ/HasRunWith |
| `backend/testframework/tms/find_by_key_a3003_test.go` | 48 | 0 | чисто — Ent sqlite A3-003 mutation-ready |
| `backend/testframework/tms/link_issue_scope_a3003_ent_test.go` | 91 | 0 | чисто — FindIssueInProject + Link Postgres |
| `backend/testframework/tms/result_scope_a3008_ent_test.go` | 44 | 0 | чисто — HasRunWith Postgres A3-008 |
| `backend/testframework/tms/result_scope_a3008_test.go` | 74 | 0 | чисто — mock service wrong-run + foreign bug |
| `backend/testframework/tms/update_optimistic_a3004_ent_test.go` | 47 | 0 | чисто — VersionEQ Postgres A3-004 |
| `backend/testframework/tms/update_optimistic_a3004_test.go` | 43 | 0 | чисто — mock service 409 (не Ent) |

Файлов в списке: 22. Прочитано: 22. Пропущено: 0.

## Перепроверка grok-2026-07 (A3-003 / 004 / 007 / 008)

| ID | Было | Текущий код | Статус |
|----|------|-------------|--------|
| A3-003 | `FindTestCaseByKey` без project | `ent_repository.go:161-174` — `CaseKey` + `DeletedAtIsNil` + `HasProjectWith`; контракт `repository.go:22-23`; тест `find_by_key_a3003_test.go` | **FIXED** |
| A3-003-link | Link к чужому issue | `FindIssueInProject` `ent_repository.go:778-787` — `DeletedAtIsNil` + `HasProjectWith`; Ent Postgres `link_issue_scope_a3003_ent_test.go` | **FIXED** |
| A3-004 | `UpdateOneID` + `AddVersion` без `VersionEQ` | `ent_repository.go:262-331` — `Where(IDEQ, VersionEQ).AddVersion(1)`; `affected==0` → 409; Postgres Ent + mock | **FIXED** |
| A3-007 | `GetTestCaseVersion` без ownership | `ent_repository.go:369-378` — `ID` + `HasTestCaseWith(testcase.ID)`; Ent Postgres тест в A4 (`rollback_a3007_ent_test.go`, вне списка) | **FIXED** |
| A3-008 | result update без run | `ent_repository.go:678-710` — `HasRunWith(testrun.ID(runID))`; Postgres Ent + mock service | **FIXED** |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W2-A3-001 | P1 | `environment.go:181-198`, `68-73`; `ent_repository.go:1042-1058` | `UpdateEnvironment` / `DeleteEnvironment`: handler не парсит `projectId`, service/repo мутируют только по `envID` (`UpdateOneID` / `DeleteOneID`) | Member проекта A с `PermTestCaseEdit` меняет/удаляет окружение проекта B по UUID; RBAC смотрит `:projectId` URL, мутация — нет | `Update/Delete.Where(IDEQ, HasProjectWith(projectID))`; signature + service + handler с `parseTMSProjectID`; affected==0 → 404 |
| W2-A3-002 | P1 | `ent_repository.go:422-445`; handler `handler_suites.go:72-108` (контекст) | `UpdateSuite` / `DeleteSuite` — только `UpdateOneID` / `DeleteOneID` без project | Тот же класс IDOR: `:projectId` в URL не участвует в Where | `Where(IDEQ, HasProjectWith(projectID))`; прокинуть `projectID` через service/handler |
| W2-A3-003 | P1 | `ent_repository.go:490-492`; handler `handler_suites.go:171-178` (контекст) | `DeleteComponent` — `DeleteOneID` без project | Member A удаляет компонент проекта B | `Delete().Where(IDEQ, HasProjectWith(projectID))`; signature + service/handler |
| W2-A3-004 | P2 | `ent_repository.go:264-268` | `UpdateTestCase` — `IDEQ`+`VersionEQ`, **без** `DeletedAtIsNil` | Зная ID+version, можно править TC в корзине (аналог W1-A3-005) | Добавить `testcase.DeletedAtIsNil()` в Where update |
| W2-A3-005 | P2 | `ent_repository.go:177-215` | `FindTestCaseByExternalID` / `ByAutomationID` / `ByFullName` / `ByClassMethod` — project scope есть, soft-delete нет | Allure matcher (`allure_matcher.go`) может привязать результат к soft-deleted TC | Добавить `testcase.DeletedAtIsNil()` во все Find* (как в `FindTestCaseByKey`) |
| W2-A3-006 | P2 | `mongo_repository.go:96-116`, `73-93` | `GetAllureResultsByRunID` — `Find` без `SetLimit`; `StoreAllureResults` — `InsertMany` без cap | Большой Allure-прогон → OOM / неограниченная запись в Mongo | `SetLimit` (+ cursor/pagination); cap на len(results) до InsertMany |
| W2-A3-007 | P2 | `ent_repository.go:714-720`, `953-960`, `829-833`, `412-418`, `483-487`, `1034-1038`, `361-366`, `724-771`, `925-938` | List* результатов/сьютов/компонентов/env/версий/issues и bulk create без Limit/cap | `ListTestResults` / `ListTestRunResults` / `GetProjectIssuesWithTestCases` / `BatchCreateTestResults` / `AddTestCasesToRun` — unbounded на горячих путях импорта и отчётов | Limit/cursor на list; max batch size на CreateBulk/AddCases |
| W2-A3-008 | P2 | `ent_repository.go:444-445`, `491-492`, `1057-1058` | Hard delete Suite / Component / Environment | ТЗ: soft delete → корзина 30 дней; у TC есть `deleted_at`, у suite/component/env — нет, сразу hard delete | Schema `deleted_at` + soft delete (как TestCase); hard delete — Admin/GDPR |
| W2-A3-009 | P3 | `ent_repository.go:448-462`, `1006-1009`, `334-338` | `MoveSuite` / `AddFilesToResult` / `DeleteTestCase` — мутации только по UUID сущности | Repo-level TOCTOU/IDOR, если caller ошибётся; HTTP MoveSuite сейчас не смонтирован, Delete TC смягчён FindByKey на service | Где есть HTTP — Where + project/run; DeleteTestCase: soft-delete Where + project |
| W2-A3-010 | P3 | `ent_repository.go:829-833` | `GetProjectIssuesWithTestCases` без `DeletedAtIsNil` на Issue | Coverage/traceability отчёты включают задачи из корзины | `entissue.DeletedAtIsNil()` (+ опционально filter linked TC) |

## Чисто

- **A3-003 FindTestCaseByKey** — `HasProjectWith` + soft-delete; regression Ent sqlite.
- **A3-003-link FindIssueInProject** — project + soft-delete; Link/Unlink service path на Postgres Ent.
- **A3-004 VersionEQ** — Where + AddVersion + 409; Postgres Ent mutation-ready; mock отдельно документирован как не-Ent.
- **A3-007 GetTestCaseVersion** — `HasTestCaseWith`; код в этом срезе FIXED (Ent-тест у A4).
- **A3-008 UpdateTestResult** — `HasRunWith(runID)`; Postgres Ent + mock Execute/CreateBug.
- **A3-012 GetTestRunInProject / GetTestRunWithResultsInProject** — project scope на read-путях.
- **CreateTestCase / CreateBugIssue tx** — Rollback на ошибках, Commit в конце; counter+entity атомарны.
- **ListTestCases / ListTestRuns** — cap limit 1..100, default 50.
- **DeleteTestCase** — soft delete через `SetDeletedAt` (не hard).
- **combined / provisioner / export_hooks / repository interfaces** — без дефектов по чек-листу.
- **Тесты списка** — assert’ы есть; A3-004/008 закрывают mock-only gap Ent-слоем; gap на IDOR suite/component/env — тестов нет (ожидаемо при открытом дефекте).

## Сводка

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 5 |
| P3 | 2 |
| **Σ** | **10** |

Файлов: **22/22/0** (список / прочитано / пропущено).

**Топ-3:** W2-A3-001 (Environment Update/Delete IDOR) · W2-A3-002 (Suite Update/Delete IDOR) · W2-A3-003 (Component Delete IDOR).
