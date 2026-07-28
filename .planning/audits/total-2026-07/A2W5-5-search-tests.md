# A2-W5 агент A5 — search remaining tests (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc`  
> Зеркало: `W5-ASSIGNMENTS.md` §A5 · Stage A: `A5-5-search-tests.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/testframework/search/service_scope_test.go` | 199 | 1 | замечания |
| `backend/testframework/search/service_shard_errors_test.go` | 50 | 1 | замечания |
| `backend/testframework/search/service_test.go` | 373 | 0 | чисто |
| `backend/testframework/search/service_wave100_test.go` | 77 | 1 | замечания |
| `backend/testframework/search/service_wave103_mongo_test.go` | 166 | 0 | чисто (FIXED-OK Len) |
| `backend/testframework/search/service_wave106_test.go` | 111 | 1 | замечания |
| `backend/testframework/search/service_wave108_test.go` | 36 | 0 | чисто (дубль empty) |
| `backend/testframework/search/service_wave109_test.go` | 26 | 0 | чисто |
| `backend/testframework/search/service_wave114_test.go` | 29 | 0 | чисто (дубль empty) |
| `backend/testframework/search/service_wave115_test.go` | 52 | 1 | замечания |
| `backend/testframework/search/service_wave116_postgres_test.go` | 319 | 0 | чисто |
| `backend/testframework/search/service_wave92_test.go` | 55 | 0 | чисто |
| `backend/testframework/search/soft_delete_a4002_runtime_test.go` | 67 | 1 | замечания |
| `backend/testframework/search/soft_delete_a4002_test.go` | 62 | 1 | замечания |

Файлов в списке: 14. Прочитано: 14. Пропущено: 0.

## Находки

| ID | Sev | Tag | Файл:строка | Проблема | A1 / задача | Мутация |
|----|-----|-----|-------------|----------|-------------|---------|
| A2W5-A5-001 | P2 | FIXED | `service_scope_test.go` | hollow → Postgres TASK037 | TASK-113 | `TestSearch_UsesProjectScopeForIssues` |
| A2W5-A5-002 | P2 | FIXED | soft_delete runtime | + test_case/project | TASK-113 | `TestPostgresSearch_ExcludesSoftDeleted_A4_002` |
| A2W5-A5-003 | P2 | FIXED | static Contains secondary | primary = runtime | TASK-113 | comment + runtime suite |
| A2W5-A5-004 | P3 | FIXED | EmptyProjectIDs канон | дубли wave убраны | TASK-113 | `TestSearchScopedEntities_EmptyProjectIDsReturnNothing` |
| A2W5-A5-005 | P3 | FIXED | wave109 | один DeleteSavedFilter NotFound | TASK-113 | `TestDeleteSavedFilter_FilterMissing` |
| A2W5-A5-006 | P3 | FIXED | wave106/115 | sqlite Empty → ErrSearchUnavailable | TASK-113 | missing tables ≠ NF-03 |

## FIXED-OK

| A1 ID | Evidence |
|-------|----------|
| W5-A5-002 | wave103: `require.Len(results, 1)` + foreign channel excluded |
| assert-free | Строго assert-free `Test*` в 14 файлах не найдено |
| saved-filter | mock service + Ent IsProjectMember пара жива |
| wave116 | Postgres users/admin scope outsider excluded |

## Чисто

`service_test.go` Create/List NonMemberForbidden; GetUser*IDs в scope_test с exact UUID; soft_delete runtime issues/files mutation-ready.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **14 / 14 / 0** |
| Находок | **6** (NEW 0 · REGRESS 0 · RESIDUAL 6) |
| P0 | 0 |
| P1 | 0 |
| P2 | 3 |
| P3 | 3 |
