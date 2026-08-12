# A3 W2 агент 3 — tms repositories (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W2-3-tms-repos.md](A2W2-3-tms-repos.md) · A1: [A2-3-tms-repos.md](A2-3-tms-repos.md)  
> Периметр: 22 файла (7 prod + 15 tests) · Прочитано: **22/22** · Пропущено: **0**  
> Remediation: TASK-111 (`A2W2-3-*`), TASK-015/038/039 (A1 closes)

## Сверка с A2 (A2W2-3-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W2-3-001 | FIXED-OK | закрыт | `StoreAllureResults` `len > maxMongoAllureResults` → err; `GetAllureResultsByRunID` `SetLimit(maxMongoAllureResults)`; const в `ent_repository.go:52` |
| A2W2-3-002 | RESIDUAL | частично | TASK-111 закрыл `ListTestResults` / `ListSuites` / `ListComponents` / `ListTestCaseVersions` / `ListTestRunResults` / `BatchCreateTestResults`; **остались** unbounded batch/report paths — см. A3W2-3-001 |
| A2W2-3-003 | FIXED-OK | закрыт | `DeleteTestCase(ctx, projectID, id)` + `testCaseProjectIDEQ` + `DeletedAtIsNil`; тест `TestTask111_DeleteTestCase_ProjectScoped` (вне слота) |

**REGRESS:** 0

### Сверка с A1 (через A2 FIXED-OK) — подтверждение на HEAD

| A1 / A2 FIXED-OK | Tag на A3 |
|------------------|-----------|
| W2-A3-001 env Update/Delete projectID | FIXED-OK — `environment.go` + repo `HasProjectWith` + soft `SetDeletedAt` |
| W2-A3-002/003 Suite/Component scoped | FIXED-OK — `HasProjectWith` + soft delete |
| W2-A3-004 UpdateTestCase DeletedAtIsNil | FIXED-OK — `ent_repository.go:284-285` + VersionEQ |
| W2-A3-005 Find* soft-delete | FIXED-OK — External/Automation/FullName/ClassMethod |
| W2-A3-008 soft Suite/Component/Environment | FIXED-OK — не `DeleteOneID` |
| W2-A3-009 MoveSuite projectID | FIXED-OK |
| W2-A3-010 GetProjectIssuesWithTestCases soft-delete | FIXED-OK на Issue; list cap — см. residual |
| A3-003/004/007/008 | FIXED-OK без регрессии (FindByKey / VersionEQ / GetTestCaseVersion / UpdateTestResult HasRunWith) |

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/tms/repository.go` | 337 | 0 | чисто — ISP; `DeleteTestCase(projectID,id)`; scoped signatures |
| `backend/internal/tms/ent_repository.go` | 1348 | 3 | caps на части List*/Batch; residual unbounded + label/link soft-delete gaps |
| `backend/internal/tms/mongo_repository.go` | 133 | 0 | FIXED-OK InsertMany cap + Find SetLimit |
| `backend/internal/tms/combined_repository.go` | 22 | 0 | чисто — embedding Ent+Mongo |
| `backend/internal/tms/environment.go` | 251 | 0 | FIXED-OK Update/Delete + `parseTMSProjectID` |
| `backend/internal/tms/provisioner.go` | 27 | 0 | чисто — CreateDefaultSuite(projectID) |
| `backend/internal/tms/export_hooks.go` | 38 | 0 | чисто — hubtest re-exports |
| `backend/testframework/tms/ent_repository_batch_test.go` | 122 | 0 | чисто — BatchCreate + ListTestRuns filter + coverage |
| `backend/testframework/tms/ent_repository_extended_test.go` | 263 | 0 | чисто — versions/results/env/link/bug/suite |
| `backend/testframework/tms/ent_repository_list_update_test.go` | 110 | 0 | чисто — List filters + Update all fields |
| `backend/testframework/tms/ent_repository_test.go` | 185 | 0 | чисто — CRUD + Find* + suite/component + run |
| `backend/testframework/tms/ent_repository_wave122_extra_test.go` | 105 | 0 | чисто — AddCases/CreateBug/all fields |
| `backend/testframework/tms/ent_repository_wave122_test.go` | 72 | 0 | чисто — MoveSuite + Reopen + stats |
| `backend/testframework/tms/ent_repository_wave124_test.go` | 48 | 0 | чисто — CreateBug + Reopen |
| `backend/testframework/tms/mongo_repository_test.go` | 53 | 0 | lifecycle OK; **нет** assert на maxMongo cap (gap evidence) |
| `backend/testframework/tms/postgres_ent_helper_test.go` | 99 | 0 | чисто — Skip/Fail без PG; isolated DB |
| `backend/testframework/tms/find_by_key_a3003_test.go` | 48 | 0 | чисто — A3-003 foreign project NotFound |
| `backend/testframework/tms/link_issue_scope_a3003_ent_test.go` | 91 | 0 | чисто — FindIssueInProject + Link Postgres |
| `backend/testframework/tms/result_scope_a3008_ent_test.go` | 44 | 0 | чисто — HasRunWith Postgres |
| `backend/testframework/tms/result_scope_a3008_test.go` | 74 | 0 | чисто — mock wrong-run + foreign bug |
| `backend/testframework/tms/update_optimistic_a3004_ent_test.go` | 47 | 0 | чисто — VersionEQ Postgres 409 |
| `backend/testframework/tms/update_optimistic_a3004_test.go` | 43 | 0 | чисто — mock service 409 |

Файлов в списке: 22. Прочитано: 22. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|----------|-------------------|------------|
| A3W2-3-001 | FIXED | P2 | `ent_repository.go` | Unbounded ByRunIDs / LastResults / AddCases / WithResults / env·issues | OOM / долгий All | TASK-140: Limit / First / batch err — см. backlog |
| A3W2-3-002 | FIXED | P3 | `ent_repository.go` CountLabels | без `label.DeletedAtIsNil()` | trash labels в count | TASK-140: DeletedAtIsNil |
| A3W2-3-003 | FIXED | P3 | `ent_repository.go` linked edges | `WithLinked*` без soft-delete | trash в traceability | TASK-140: edge DeletedAtIsNil |

Remediation: `.planning/backlog/archive/2026-07-29/TASK-140-2026-07-29-a3-2-3-tms-repos.md` · статус **На проверке** · hubtest `TestTask140_*` PASS.

## Чисто (подтверждено после remediation)

- IDOR/scopes: FindByKey / FindIssueInProject / UpdateTestResult HasRunWith / Suite·Component·Environment Update/Delete·MoveSuite — без регрессии; Ent twins sqlite+PG живы.
- Optimistic lock: `UpdateTestCase` VersionEQ + DeletedAtIsNil + 409.
- Soft-delete: TC / Suite / Component / Environment — `SetDeletedAt`, не hard DeleteOneID.
- Mongo Allure: InsertMany len-cap + Find SetLimit (A2W2-3-001).
- List caps (частично): ListTestResults / Suites / Components / Versions / ListTestRunResults / BatchCreateTestResults.
- ListTestCases / ListTestRuns — limit 1..100, default 50.
- A3-003/004/007/008 mutation-ready тесты в слоте зелёные по контракту.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

A2 baseline: 3 активных → FIXED-OK 2 · RESIDUAL 1 (partial A2W2-3-002) · NEW 2 · REGRESS 0.

**Топ ID:** A3W2-3-001 (unbounded ByRunIDs/LastResults/AddCases/WithResults) · A3W2-3-002 (CountLabels soft-delete) · A3W2-3-003 (linked edges trash).
