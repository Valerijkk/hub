# A2 · W2 агент A4 — tms import/parsers (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало: [A2-4-tms-import.md](A2-4-tms-import.md) · TASK-016 W18 PASS

## Прочитанные файлы

| Файл | Строк | Open | Вердикт |
|------|------:|-----:|---------|
| `backend/internal/tms/allure_matcher.go` | 159 | 1 | RESIDUAL NotFound на automation/case_key |
| `backend/internal/tms/allure_parser.go` | 58 | 1 | RESIDUAL no depth (webhook path) |
| `backend/internal/tms/allure_types.go` | 66 | 0 | чисто |
| `backend/internal/tms/exporter.go` | 255 | 1 | RESIDUAL PDF byte truncate |
| `backend/internal/tms/importer.go` | 579 | 2 | LimitReader FIXED; export/tx residual |
| `backend/internal/tms/junit_parser.go` | 140 | 1 | RESIDUAL no depth |
| `backend/internal/tms/testrail_import.go` | 363 | 1 | RESIDUAL partial + err.Error handler |
| `backend/internal/tms/export_hooks_converters.go` | 47 | 0 | чисто |
| `backend/internal/tms/export_hooks_mongo.go` | 7 | 0 | чисто |
| `backend/internal/tms/export_hooks_repo.go` | 7 | 0 | чисто |
| `backend/internal/tms/export_hooks_testrail.go` | 50 | 0 | чисто |
| `backend/testframework/tms/allure_matcher_wave99_test.go` | 50 | 1 | RESIDUAL mock ignores projectID |
| `backend/testframework/tms/allure_parser_test.go` | 557 | 0 | mock project gap (см. matcher) |
| `backend/testframework/tms/coverage_wave_test.go` | 116 | 0 | чисто |
| `backend/testframework/tms/environment_wave105_test.go` | 190 | 0 | чисто |
| `backend/testframework/tms/execute_by_a3017_test.go` | 41 | 0 | чисто |
| `backend/testframework/tms/exporter_test.go` | 43 | 0 | чисто |
| `backend/testframework/tms/exporter_wave95_test.go` | 148 | 0 | чисто |
| `backend/testframework/tms/import_run_scope_a3012_test.go` | 32 | 0 | чисто |
| `backend/testframework/tms/importer_wave96_test.go` | 48 | 0 | oversize покрыт др. тестом |
| `backend/testframework/tms/junit_parser_test.go` | 138 | 0 | нет deep-bomb assert |
| `backend/testframework/tms/rollback_a3007_ent_test.go` | 30 | 0 | чисто |
| `backend/testframework/tms/rollback_a3007_test.go` | 45 | 0 | чисто |
| `backend/testframework/tms/run_scope_a3012_ent_test.go` | 41 | 0 | чисто |
| `backend/testframework/tms/testrail_import_wave96_test.go` | 188 | 0 | LimitReader prod OK |

Файлов: **25/25/0**.

## FIXED-OK

| Stage A ID | Evidence |
|------------|----------|
| W2-A4-001 | `importer.go:34–39`, `:350–355` — `LimitReader` + 413; records cap 5000; TASK-016 PASS |
| W2-A4-002 | `applyFieldToInput` `case "steps"` (`:432`); CSV/JSON mapping |
| W2-A4-003 (importer) | `assertJSONMaxDepth` / `assertXMLMaxDepth` на ImportFromJSON/XML |
| W2-A4-010 | мёртвый `csv.NewReader` удалён (`testrail_import.go:282`) |
| W2-A4-011 | regexp package-level (`stepNumberPrefixRegexp`) |
| W2-A4-012 | `strings.ToLower(filename)` перед HasSuffix (`importer.go:313`) |
| W2-A4-013 | CSV Write errors checked + Flush (`exporter.go:169–183`) |
| A3-010/012 | matcher projectID + export/import run scope — без регрессии |

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему | Фикс |
|----|-----|-----|-------------|----------|--------|------|
| A2W2-4-001 | RESIDUAL (W2-A4-003) | P2 | `allure_parser.go:18–21`; `junit_parser.go:64–72`; webhook callers | Depth/cap на **importer** есть; `ParseAllureResults` / `ParseJUnitXML` / webhook CI path — без `assert*MaxDepth` / max results | Nested bomb в пределах 5–25 MB всё ещё amplifies alloc | Depth+len cap в парсерах или перед ними в webhook/service |
| A2W2-4-002 | RESIDUAL (W2-A4-004) | P2 | `exporter.go:261–264` | PDF comment `comment[:30]` по байтам; рядом `truncate()` rune-safe | Non-ASCII → invalid UTF-8 / mid-rune | `truncate(comment, 30)` |
| A2W2-4-003 | RESIDUAL (W2-A4-005) | P2 | `allure_matcher.go:113–145` | automation_id / case_key: при `err != nil` нет ветки `!ent.IsNotFound` (в отличие от annotation/fullName) | DB timeout → ложный unmapped / следующий матч | Как стр. 89–91 / 154 |
| A2W2-4-004 | RESIDUAL (W2-A4-006) | P2 | `importer.go` create loop; `testrail_import.go:210–269` | Кейсы/сьюты по одному; err → `Errors++` / continue | Частичный импорт без отката | Tx / явный `partial` в ответе |
| A2W2-4-005 | RESIDUAL (W2-A4-007) | P2 | `importer.go:532` | Export `ListTestCases(Limit: 1000)` без сигнала truncate | Проект >1000 кейсов — тихая усечённая выгрузка | Пагинация до конца / `truncated` / 409 |
| A2W2-4-006 | RESIDUAL (W2-A4-008) | P2 | `allure_matcher_wave99_test.go`; mock Find* | Mock игнорирует `projectID`; нет cross-project denial | Снятие передачи projectID в matcher не краснеет unit | Negative mock / Ent Match→Find |
| A2W2-4-007 | RESIDUAL (testrail) | P3 | `testrail_import.go:411` | Handler отдаёт `err.Error()` на 400 | Internals парсера клиенту | Фиксированное сообщение |

## Фокус-чеклист

| Тема | Вердикт A2 |
|------|------------|
| LimitReader handleImport | **FIXED-OK** |
| Steps mapping | **FIXED-OK** |
| JSON/XML depth importer | **FIXED-OK**; webhook/parsers — **RESIDUAL** |
| A3-010 matcher project | **FIXED-OK** prod; test gap RESIDUAL |
| A3-012 run scope | **FIXED-OK** |
| Path traversal | чисто (filename только суффикс) |
| TestRail secrets in logs | чисто |

## Сводка

| | Count |
|--|------:|
| Files | 25/25/0 |
| FIXED-OK classes | 8 |
| NEW | 0 |
| REGRESS | 0 |
| RESIDUAL | 7 |
| Open P0/P1/P2/P3 | 0 / 0 / 6 / 1 |
