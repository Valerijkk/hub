# A3 · W2 агент A4 — tms import/parsers (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W2-4-tms-import.md](A2W2-4-tms-import.md) · remediation: [TASK-111](../../backlog/TASK-111-2026-07-28-tms-wave2-services-import-batch.md)  
> Периметр: 25 файлов (11 prod + 14 tests) · Прочитано: **25/25** · Пропущено: **0**

## Сверка с A2 (A2W2-4-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W2-4-001 | FIXED-OK | закрыт | `ParseAllureResults` / `ParseAllureResult` / `ParseJUnitXML`: `assert*MaxDepth` + `maxAllureParseResults=5000`; `TestTask111_ParseAllure_DepthCap` / `_ParseJUnit_DepthCap` |
| A2W2-4-002 | FIXED-OK | закрыт | PDF `truncate(r.Comment, 30)` rune-safe (`exporter.go:261–262`); helper `truncate` на runes; `TestTask111_PDFTruncateRuneSafe` |
| A2W2-4-003 | FIXED-OK | закрыт | automation_id / case_key: `!ent.IsNotFound(err) && err != nil` → fail (`allure_matcher.go:120–123`, `:133–135`, `:151–153`); `TestTask111_MatcherAutomationID_DBError` |
| A2W2-4-004 | FIXED-OK | закрыт | `ImportResult.Partial` в DTO + `handleImport` / `importTestRailCases` (`dto.go:454`, `importer.go:420–421`, `testrail_import.go:273`) — явный partial без скрытого «успеха» |
| A2W2-4-005 | FIXED-OK | закрыт | `handleExportTestCases` пагинирует `ListTestCases` pageSize=100 до `offset >= total` (`importer.go:534–549`) |
| A2W2-4-006 | FIXED-OK | закрыт | mock `Find*` уважает `projectID` при `Edges.Project`; `TestTask111_MatcherCrossProjectDenial` |
| A2W2-4-007 | FIXED-OK | закрыт | TestRail handler: фиксированное `"некорректный файл TestRail"` + slog (`testrail_import.go:411–414`) |

**Ранее FIXED (W2-A4 / A3-010/012, без регрессии):** LimitReader+413, steps mapping, importer JSON/XML depth, case-insensitive filename, CSV Write+Flush (run results), dead `csv.NewReader`, package-level regexp, matcher projectID prod, run-scope import/export.

**REGRESS:** 0

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/tms/allure_matcher.go` | 185 | 0 | FIXED-OK A2W2-4-003 |
| `backend/internal/tms/allure_parser.go` | 77 | 0 | FIXED-OK A2W2-4-001 |
| `backend/internal/tms/allure_types.go` | 74 | 0 | чисто |
| `backend/internal/tms/exporter.go` | 295 | 1 | CSV Flush без `w.Error` (кейсы) |
| `backend/internal/tms/importer.go` | 665 | 2 | Steps export gap; parse `err.Error` в 400 |
| `backend/internal/tms/junit_parser.go` | 167 | 0 | FIXED-OK A2W2-4-001 |
| `backend/internal/tms/testrail_import.go` | 418 | 2 | нет depth/records cap; suite fail silent |
| `backend/internal/tms/export_hooks_converters.go` | 60 | 0 | чисто (hubtest) |
| `backend/internal/tms/export_hooks_mongo.go` | 10 | 0 | чисто (hubtest) |
| `backend/internal/tms/export_hooks_repo.go` | 10 | 0 | чисто (hubtest) |
| `backend/internal/tms/export_hooks_testrail.go` | 63 | 0 | чисто (hubtest) |
| `backend/testframework/tms/allure_matcher_wave99_test.go` | 59 | 0 | happy-path; denial в task111 |
| `backend/testframework/tms/allure_parser_test.go` | 631 | 0 | mock projectID FIXED-OK |
| `backend/testframework/tms/coverage_wave_test.go` | 138 | 0 | чисто |
| `backend/testframework/tms/environment_wave105_test.go` | 220 | 0 | чисто (вне import-ядра) |
| `backend/testframework/tms/execute_by_a3017_test.go` | 49 | 0 | чисто |
| `backend/testframework/tms/exporter_test.go` | 50 | 0 | unit writers OK; service Steps gap вне |
| `backend/testframework/tms/exporter_wave95_test.go` | 168 | 0 | чисто |
| `backend/testframework/tms/import_run_scope_a3012_test.go` | 38 | 0 | FIXED-OK A3-012 |
| `backend/testframework/tms/importer_wave96_test.go` | 54 | 0 | Excel happy-path; limits в др. файле |
| `backend/testframework/tms/junit_parser_test.go` | 152 | 0 | depth; depth в task111 |
| `backend/testframework/tms/rollback_a3007_ent_test.go` | 38 | 0 | чисто |
| `backend/testframework/tms/rollback_a3007_test.go` | 53 | 0 | чисто |
| `backend/testframework/tms/run_scope_a3012_ent_test.go` | 52 | 0 | FIXED-OK A3-012 |
| `backend/testframework/tms/testrail_import_wave96_test.go` | 212 | 0 | LimitReader OK; caps/depth — gap prod |

Файлов в списке: 25. Прочитано: 25. Пропущено: 0.

## FIXED-OK (evidence)

| Stage ID | Evidence |
|----------|----------|
| A2W2-4-001 | `allure_parser.go:21–35`, `:45–47`; `junit_parser.go:64–68`, `:91–93`; `task111_w2_residual_test.go:97–109` |
| A2W2-4-002 | `exporter.go:261–262`, `:283–289`; `TestTask111_PDFTruncateRuneSafe` / `_ExportTruncateHelper` |
| A2W2-4-003 | `allure_matcher.go:120–123`, `:133–135`, `:151–153`; `TestTask111_MatcherAutomationID_DBError` |
| A2W2-4-004 | `dto.go:454`; `importer.go:420–421`; `testrail_import.go:273` |
| A2W2-4-005 | `importer.go:534–549` pagination loop |
| A2W2-4-006 | `allure_parser_test.go:312–317`; `TestTask111_MatcherCrossProjectDenial` |
| A2W2-4-007 | `testrail_import.go:411–414` |
| W2-A4-001/002/003/010–013 | LimitReader, steps, importer depth, ToLower filename, CSV run Flush, dead reader, package regexp — без регрессии |
| A3-010 / A3-012 | matcher project + import/export run scope tests green path |

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W2-4-001 | FIXED | P2 | `importer.go` handleExportTestCases | `formatTestStepsForExport` → `exportData.Steps` (TASK-141) | was: Steps always `""` | Evidence: `TestTask141_ExportTestCases_StepsFilled` + mutation |
| A3W2-4-002 | FIXED | P2 | `testrail_import.go` parseTestRailCSV/XML | depth + `maxImportRecords` (TASK-141) | was: unbounded nest/rows | Evidence: `TestTask141_ParseTestRailXML_DepthCap` / `_CSV_RecordsCap` |
| A3W2-4-003 | FIXED | P3 | `importer.go` handleImport | fixed `"некорректный файл импорта"` + slog (TASK-141) | was: `err.Error()` leak | Evidence: `TestTask141_HandleImport_FixedParseError` |
| A3W2-4-004 | FIXED | P3 | `exporter.go` ExportTestCasesToCSV | Flush + `w.Error()` (TASK-141) | was: silent Flush fail | Evidence: `TestTask141_ExportTestCasesToCSV_FlushError` |
| A3W2-4-005 | FIXED | P3 | `testrail_import.go` CreateSuite fail | `Errors++` + Partial (TASK-141) | was: silent continue | Evidence: `TestTask141_ImportTestRail_CreateSuiteFail_Partial` |

## Фокус-чеклист

| Тема | Вердикт A3 |
|------|------------|
| LimitReader handleImport / TestRail | **FIXED-OK** |
| Steps mapping (import) | **FIXED-OK** |
| JSON/XML depth allure/junit/importer | **FIXED-OK** |
| TestRail parser depth/records | **FIXED** A3W2-4-002 (TASK-141) |
| Matcher !IsNotFound + project mock | **FIXED-OK** |
| Export ListTestCases pagination | **FIXED-OK** |
| Export Steps populated | **FIXED** A3W2-4-001 (TASK-141) |
| ImportResult.Partial | **FIXED-OK** |
| TestRail handler err.Error | **FIXED-OK**; generic import — **FIXED** A3W2-4-003 (TASK-141) |
| A3-010 / A3-012 | **FIXED-OK** |
| Path traversal / secrets in logs | чисто |

## Чисто (подтверждено)

- `allure_types.go`, export_hooks_* — без дефектов в слоте.
- Rollback A3-007 / run-scope A3-012 / execute_by A3-017 — без регрессии.
- Filename только суффикс (`detectImportFormat` / TestRail ToLower); нет path traversal.
- Importer caps 25 MB / 5000 records / depth 32 — на месте.

### Сводка counts (после TASK-141)

| Severity open | Count |
|---------------|------:|
| P0–P3 open | **0** |
| FIXED (A3W2-4-001..005) | **5** |

| Tag | Count |
|-----|------:|
| FIXED-OK (A2W2-4 closes) | 7 |
| FIXED (A3 residual TASK-141) | 5 |
| NEW open | 0 |
| REGRESS | 0 |

A3 residual 5 → **0 open** · TASK-141 **На проверке** · hubtest + mutation RED_OK.

**Return:** `.planning/audits/total-2026-07/A3W2-4-tms-import.md` · remediation TASK-141 · **5/5 FIXED**.
