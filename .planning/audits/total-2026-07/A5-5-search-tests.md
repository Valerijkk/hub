# W5 агент A5 — search remaining tests

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: assert-free, mock vs real scope guard, soft_delete_a4002 mutation quality.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/testframework/search/service_scope_test.go` | 184 | 1 | замечания |
| `backend/testframework/search/service_shard_errors_test.go` | 55 | 1 | замечания |
| `backend/testframework/search/service_test.go` | 379 | 0 | чисто |
| `backend/testframework/search/service_wave100_test.go` | 112 | 1 | замечания |
| `backend/testframework/search/service_wave103_mongo_test.go` | 76 | 1 | замечания |
| `backend/testframework/search/service_wave106_test.go` | 118 | 1 | замечания |
| `backend/testframework/search/service_wave108_test.go` | 50 | 0 | чисто (дубль empty-scope — см. W5-A5-005) |
| `backend/testframework/search/service_wave109_test.go` | 45 | 0 | чисто (дубль NotFound — см. W5-A5-006) |
| `backend/testframework/search/service_wave114_test.go` | 54 | 0 | чисто (дубль empty-scope — см. W5-A5-005) |
| `backend/testframework/search/service_wave115_test.go` | 52 | 1 | замечания |
| `backend/testframework/search/service_wave116_postgres_test.go` | 262 | 0 | чисто |
| `backend/testframework/search/service_wave92_test.go` | 55 | 0 | чисто |
| `backend/testframework/search/soft_delete_a4002_runtime_test.go` | 67 | 1 | замечания |
| `backend/testframework/search/soft_delete_a4002_test.go` | 41 | 1 | замечания |

Файлов в списке: 14. Прочитано: 14. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W5-A5-001 | P2 | `service_scope_test.go:169-183` | `TestSearch_UsesProjectScopeForIssues` вставляет membership и ждёт только `assert.Empty` | Имя обещает NF-03 project-scope; таблицы issues нет → shard падает/пуст. Тест зелёный и при полном игноре `projectIDs` в `Search` | Postgres: active issue в allowed + foreign project; `Search`/`ExportSearchIssues` → только allowed ID; mutation: убрать `project_issues = ANY($2)` → FAIL |
| W5-A5-002 | P2 | `service_wave103_mongo_test.go:37-43,57-61` | Чужой канал с тем же query-hit засеян, assert только `NotEmpty` + `Contains("wave103")` | Оба текста матчат `"wave103 searchable"`. Снятие channel-membership фильтра → 2 hit, тест всё ещё зелёный | `require.Len(t, results, 1)`; `assert.Equal` своего `channelID` (Subtitle/URL); явно `assert.NotEqual` otherChannel |
| W5-A5-003 | P2 | `soft_delete_a4002_runtime_test.go:15-66` | Runtime A4-002 покрывает только **issues + files** | В prod `deleted_at IS NULL` также в `searchTestCases`/`searchProjects` (`service.go:276`, `443`). Снятие фильтра там → static краснеет, runtime stage-C для этих шардов нет | Расширить runtime: soft-deleted test_case + project с тем же query → `Len==1` только live; mutation: drop `deleted_at` в этих SQL → FAIL |
| W5-A5-004 | P2 | `soft_delete_a4002_test.go:16-40` | «Мутация» = `os.ReadFile` + `strings.Contains(..., "deleted_at IS NULL")` | Не исполняет SQL. Эквивалентный фильтр (`deleted_at IS NOT DISTINCT FROM NULL`) ломает тест при живой защите; мёртвая строка в chunk — ложный PASS. TASK-008: доказательство — поведение на слое guard | Primary = runtime (W5-A5-003); static оставить как lint или удалить как sole evidence |
| W5-A5-005 | P3 | `service_wave100_test.go:26-58`; `service_wave108_test.go:38-49`; `service_wave109_test.go:12-31`; `service_wave114_test.go:12-54`; `service_scope_test.go:123-149` | 5+ копий empty `projectIDs` → nil/empty | Не assert-free, но нулевой сигнал по scope/soft-delete; дубли раздувают suite | Один канонический empty-scope тест; бюджет — на foreign-project denial (W5-A5-001) |
| W5-A5-006 | P3 | `service_shard_errors_test.go:51-54`; `service_wave109_test.go:33-38` | Два почти одинаковых `DeleteSavedFilter` NotFound | Drift сообщений/ожиданий | Оставить один с `assert.Contains("не найден")` |
| W5-A5-007 | P3 | `service_wave106_test.go:12-49`; `service_wave115_test.go:12-51`; `service_test.go:233-289` | `Search_*` на sqlite без entity-таблиц → всегда `Empty` | Зелёные при сломанном soft-delete/scope SQL; проверяют только swallow missing-table / limit normalize | Не трактовать как NF-03/A4-002 proof; scope/soft-delete — только Postgres/Mongo с данными |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| **assert-free** | Во всех 14 файлах у каждого `Test*` есть `assert.*` и/или `require.Error`/`require.Len`/`require.NotEmpty`. Строго assert-free тестов **нет**. |
| `service_test.go` Create/List NonMemberForbidden | Guard `requireProjectMembership` живёт в **service** (`service.go:644-676`) → mock `memberOK=false` — правильный слой. Mutation-комменты A4-003/004 адекватны. `TestEntRepository_IsProjectMember` бьёт **реальный** Ent — пара mock-service + real-repo на месте. |
| `service_test.go` CRUD/Share/owner | Owner-check delete/share; asserts на error strings; Ent SavedFilter CRUD на sqlite Ent. |
| `service_scope_test.go` GetUser*IDs | Реальный SQL sqlite: `Len==1` + exact UUID — не mock membership для channel/project ID lists. Empty projectIDs early-return — OK. |
| `service_wave116_postgres_test.go` | Реальный Postgres: issues/projects/test_cases/files/channels; `UsersAdminAndScoped` исключает outsider (`assert.NotEqual` outsiderID) — хороший mock-admin-flag + real SQL scope. |
| `service_wave92_test.go` | Ent GetSavedFilter — asserts есть. |
| `service_shard_errors_test.go` (кроме дубля) | Closed DB error; shard recover Empty; messages no membership. |
| soft_delete runtime (issues/files) | Active vs deleted issue; confirmed vs deleted vs pending file; `Len==1` + ID — **mutation-ready** для этих двух шардов. |

### Вердикт по фокусу

| Чек-лист TASK-008 (тесты) | Статус |
|---------------------------|--------|
| assert-free тест | **Не найдено** |
| Тест на моке там, где guard в репозитории | **Не нарушено** для saved-filter membership (guard в service). Ctrl+K scope — SQL в service; слабые тесты не мок, а empty-assert без данных (W5-A5-001/002/007) |
| Тест, который пройдёт и на сломанном коде | **Да:** W5-A5-001, W5-A5-002, часть wave Empty-suite (W5-A5-007) |
| soft_delete_a4002 mutation quality | Runtime **сильный** для issues/files; **дырка** test_cases/projects; static-only — слабое sole evidence (W5-A5-003/004) |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | 14 / 14 / 0 |
| Находок | **7** |
| P0 | 0 |
| P1 | 0 |
| P2 | 4 |
| P3 | 3 |

### Топ-3

1. **W5-A5-001** — `UsesProjectScopeForIssues` не доказывает project-scope (пустой assert при отсутствии данных).  
2. **W5-A5-002** — Mongo message search не фиксирует исключение чужого канала (NotEmpty недостаточен).  
3. **W5-A5-003 / W5-A5-004** — soft_delete A4-002: runtime не покрывает test_cases/projects; static string-scan слабый как единственная мутация.

P0/P1 в remaining search tests **нет**. Saved-filter membership mock+Ent пара корректна; главные риски — ложное чувство NF-03/A4-002 от hollow/empty/static тестов.
