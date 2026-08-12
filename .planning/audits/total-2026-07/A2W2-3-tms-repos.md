# A2 · W2 агент A3 — tms repositories (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: [A2-3-tms-repos.md](A2-3-tms-repos.md) · TASK-015/038/039 W18

## Прочитанные файлы

| Файл | Строк | Open | Вердикт |
|------|------:|-----:|---------|
| `backend/internal/tms/repository.go` | 304 | 0 | чисто — ISP + scoped signatures |
| `backend/internal/tms/ent_repository.go` | 1138 | 2 | RESIDUAL lists/mongo-class; IDOR FIXED |
| `backend/internal/tms/mongo_repository.go` | 108 | 1 | RESIDUAL no Limit/cap |
| `backend/internal/tms/combined_repository.go` | 18 | 0 | чисто |
| `backend/internal/tms/environment.go` | 219 | 0 | FIXED-OK Update/Delete projectID |
| `backend/internal/tms/provisioner.go` | 22 | 0 | чисто |
| `backend/internal/tms/export_hooks.go` | 30 | 0 | чисто |
| `backend/testframework/tms/ent_repository_batch_test.go` | 105 | 0 | чисто |
| `backend/testframework/tms/ent_repository_extended_test.go` | 211 | 0 | чисто |
| `backend/testframework/tms/ent_repository_list_update_test.go` | 95 | 0 | чисто |
| `backend/testframework/tms/ent_repository_test.go` | 150 | 0 | чисто |
| `backend/testframework/tms/ent_repository_wave122_extra_test.go` | 86 | 0 | чисто |
| `backend/testframework/tms/ent_repository_wave122_test.go` | 57 | 0 | чисто |
| `backend/testframework/tms/ent_repository_wave124_test.go` | 37 | 0 | чисто |
| `backend/testframework/tms/mongo_repository_test.go` | 43 | 0 | без cap assert |
| `backend/testframework/tms/postgres_ent_helper_test.go` | 93 | 0 | чисто |
| `backend/testframework/tms/find_by_key_a3003_test.go` | 39 | 0 | чисто |
| `backend/testframework/tms/link_issue_scope_a3003_ent_test.go` | 78 | 0 | чисто |
| `backend/testframework/tms/result_scope_a3008_ent_test.go` | 35 | 0 | чисто |
| `backend/testframework/tms/result_scope_a3008_test.go` | 64 | 0 | чисто |
| `backend/testframework/tms/update_optimistic_a3004_ent_test.go` | 39 | 0 | чисто |
| `backend/testframework/tms/update_optimistic_a3004_test.go` | 36 | 0 | чисто |

Файлов: **22/22/0**.

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| W2-A3-001 | `environment.go` + `Update/DeleteEnvironment(…, projectID, …)` + `HasProjectWith` + soft `SetDeletedAt` |
| W2-A3-002/003 | `Update/DeleteSuite` / `DeleteComponent` — `HasProjectWith` + soft delete; TASK-015 mutation PASS |
| W2-A3-004 | `UpdateTestCase` Where + `DeletedAtIsNil` (`ent_repository.go:269–275`) |
| W2-A3-005 | FindByExternal/Automation/FullName/ClassMethod — `DeletedAtIsNil` |
| W2-A3-008 | Suite/Component/Environment — soft `SetDeletedAt` (не `DeleteOneID`); `soft_delete_p2be08_test.go` |
| W2-A3-009 (MoveSuite) | `MoveSuite(ctx, projectID, id, …)` + `HasProjectWith` |
| W2-A3-010 | `GetProjectIssuesWithTestCases` — `DeletedAtIsNil` на Issue |
| A3-003/004/007/008 | без регрессии (FindByKey / VersionEQ / GetTestCaseVersion / UpdateTestResult HasRunWith) |

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему | Фикс |
|----|-----|-----|-------------|----------|--------|------|
| A2W2-3-001 | RESIDUAL (W2-A3-006) | P2 | `mongo_repository.go:73–116` | `StoreAllureResults` InsertMany без cap; `GetAllureResultsByRunID` Find без `SetLimit` | Большой прогон → OOM / unbounded Mongo read | Cap len(results); `SetLimit` + pagination |
| A2W2-3-002 | RESIDUAL (W2-A3-007) | P2 | `ent_repository.go:417–422`, `:847–853`, `:1090–1097`, `:480–489`, `:585–592` | `ListTestResults` / `ListSuites` / `ListComponents` / versions — `All` без Limit | Unbounded на горячих путях отчётов/импорта | Limit/cursor; max batch на CreateBulk |
| A2W2-3-003 | RESIDUAL (W2-A3-009) | P3 | `ent_repository.go` `DeleteTestCase` | Soft-delete Where по `id` без `projectID` (caller уже scoped FindByKey) | Repo TOCTOU если caller ошибётся | `Where(ID, HasProjectWith)` / project param |

## Сводка

| | Count |
|--|------:|
| Files | 22/22/0 |
| FIXED-OK classes | 8 |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 3 |
| Open P0/P1/P2/P3 | 0 / 0 / 2 / 1 |
