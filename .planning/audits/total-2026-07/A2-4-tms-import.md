# W2 агент A4 — tms import/parsers

> Стадия A · read-only · 2026-07-28 · TASK-008  
> Фокус: LimitReader vs ReadAll, path traversal, Allure matcher project scope (A3-010), XML/JSON bomb, import run scope (A3-012), TestRail secrets in logs.

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/internal/tms/allure_matcher.go | 176 | 1 | замечания |
| backend/internal/tms/allure_parser.go | 65 | 1 | замечания |
| backend/internal/tms/allure_types.go | 75 | 0 | чисто |
| backend/internal/tms/exporter.go | 288 | 1 | замечания |
| backend/internal/tms/importer.go | 534 | 4 | замечания |
| backend/internal/tms/junit_parser.go | 160 | 1 | замечания |
| backend/internal/tms/testrail_import.go | 418 | 1 | замечания |
| backend/internal/tms/export_hooks_converters.go | 61 | 0 | чисто |
| backend/internal/tms/export_hooks_mongo.go | 11 | 0 | чисто |
| backend/internal/tms/export_hooks_repo.go | 11 | 0 | чисто |
| backend/internal/tms/export_hooks_testrail.go | 64 | 0 | чисто |
| backend/testframework/tms/allure_matcher_wave99_test.go | 61 | 1 | замечания |
| backend/testframework/tms/allure_parser_test.go | 616 | 1 | замечания |
| backend/testframework/tms/coverage_wave_test.go | 139 | 0 | чисто |
| backend/testframework/tms/environment_wave105_test.go | 212 | 0 | чисто |
| backend/testframework/tms/execute_by_a3017_test.go | 50 | 0 | чисто |
| backend/testframework/tms/exporter_test.go | 51 | 0 | чисто |
| backend/testframework/tms/exporter_wave95_test.go | 164 | 0 | чисто |
| backend/testframework/tms/import_run_scope_a3012_test.go | 39 | 0 | чисто |
| backend/testframework/tms/importer_wave96_test.go | 55 | 1 | замечания |
| backend/testframework/tms/junit_parser_test.go | 153 | 1 | замечания |
| backend/testframework/tms/rollback_a3007_ent_test.go | 39 | 0 | чисто |
| backend/testframework/tms/rollback_a3007_test.go | 54 | 0 | чисто |
| backend/testframework/tms/run_scope_a3012_ent_test.go | 53 | 0 | чисто |
| backend/testframework/tms/testrail_import_wave96_test.go | 213 | 1 | замечания |

Файлов в списке: 25. Прочитано: 25. Пропущено: 0.

## Перепроверка прошлого аудита (A3 / grok-2026-07)

| Старый ID | Статус в текущем коде | Комментарий |
|-----------|----------------------|-------------|
| A3-010 | **исправлено** — не дублировать как IDOR | `MatchTestCase` передаёт `projectID` во все `Find*` (`allure_matcher.go:84–162`); Ent — `HasProjectWith` (`ent_repository.go:162–215`). Регрессия Ent: `find_by_key_a3003_test.go` (вне списка A4). |
| A3-012 | **исправлено** — не дублировать как IDOR | Export: `GetTestRunWithResultsInProject` (`importer.go:466`). Import path (service вне списка): `GetTestRunInProject`. Покрытие A4: mock `import_run_scope_a3012_test.go` + Ent `run_scope_a3012_ent_test.go`. |
| A3-003 | закрыт кодом (repo) | matcher зависит от scoped find; см. A3-010 |
| A3-007 | покрыт тестами в списке | `rollback_a3007_*.go` — Ent + mock; prod rollback вне scope A4 |

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W2-A4-001 | P1 | `backend/internal/tms/importer.go:268–316` | `handleImport` открывает `FormFile` и отдаёт поток в CSV/JSON/XML/Excel **без** `io.LimitReader` и без лимита числа записей | TestRail (`testrail_import.go:385`) и Allure/JUnit webhook используют `ReadAll(LimitReader(...))`. Здесь только глобальный Fiber `BodyLimit` 25 MB: внутри лимита CSV `ReadAll`, JSON/XML decode и `excelize.OpenReader` (ZIP) дают memory/CPU amplification (bomb). TZ: не полагаться на голый unbounded parse | Обернуть `io.LimitReader(f, max+1)`; константа как у TestRail/webhook; hard-cap `len(inputs)` (напр. 5k); для xlsx — отдельный лимит uncompressed; тест 413 / «слишком много строк» |
| W2-A4-002 | P1 | `backend/internal/tms/importer.go:45–46`, `108–136`, `357–372` | AutoDetect мапит `"steps"/"шаги"→steps`, JSON-структура имеет `Steps`, но `applyFieldToInput` **не** обрабатывает `steps`, а `ImportFromJSON` **не** копирует `item.Steps` в `CreateTestCaseInput` | Импорт «успешен», шаги молча теряются; `CreateTestCaseInput.Steps` остаётся nil. Ломает миграцию кейсов из CSV/JSON | Парсить steps (как `parseTestRailSteps`) → `[]schema.TestStep`; добавить `case "steps"` в `applyFieldToInput`; покрыть тестом «колонка Steps → непустой Steps» |
| W2-A4-003 | P1 | `backend/internal/tms/allure_parser.go:18–21`; `junit_parser.go:64–72`; `importer.go:119–122`, `158–161`; `testrail_import.go:348–352` | Парсеры Allure/JUnit/import XML/JSON/TestRail XML — `json.Unmarshal` / `xml.Unmarshal` / `Decoder.Decode` без лимита глубины/числа элементов | Даже при ограниченном байтовом входе nested steps/labels/testcase дают раздувание аллокаций (JSON/XML bomb). Нет `MaxBytesReader` на stream-путях import | Лимит размера входа + cap `len(results)` / глубины steps; отвергать oversized; regression-тест на глубоко вложенный payload → 400 |
| W2-A4-004 | P2 | `backend/internal/tms/exporter.go:251–254` | Обрезка `comment` через `comment[:30]` по **байтам**; рядом `truncate()` уже rune-safe (`275–281`) | Non-ASCII комментарий → невалидный UTF-8 / panic на многобайтной границе в PDF-экспорте | Использовать `truncate(comment, 30)` |
| W2-A4-005 | P2 | `backend/internal/tms/allure_matcher.go:113–145` | Стратегии `automation_id` и `case_key` при `err != nil` **не** отличают `NotFound` от прочих ошибок (в отличие от annotation/fullName/classMethod) | Сбой БД/таймаут глотается → ложный unmapped / неверный следующий матч вместо ошибки импорта | Как в стр. 89–91: `if !ent.IsNotFound(err) && err != nil { return nil, -1, err }` |
| W2-A4-006 | P2 | `backend/internal/tms/importer.go:337–352`; `testrail_import.go:210–269` | Создание кейсов/сьютов по одному без транзакции; ошибки create только `Errors++` / `continue` | Частичный импорт без отката: половина кейсов в БД, половина нет; нет идемпотентности | Batch в Tx или all-or-nothing + явный `partial` в ответе; не глотать err без ValidationErrors |
| W2-A4-007 | P2 | `backend/internal/tms/importer.go:414–415` | `ListTestCases(..., Limit: 1000)` при экспорте | Проект >1000 кейсов — тихая усечённая выгрузка без сигнала клиенту | Пагинация до конца / явный `truncated` / 409 если total > max |
| W2-A4-008 | P2 | `backend/testframework/tms/allure_parser_test.go:262–305`; `allure_matcher_wave99_test.go:29–50` | Mock `Find*` игнорирует `projectID` (`_ uuid.UUID`); тесты матчера не проверяют cross-project denial | A3-010 guard живёт в Ent; unit-матчер останется зелёным, если matcher перестанет передавать `projectID` | Negative-тест: mock, возвращающий TC только при совпадении projectID; либо Ent-интеграция Match→Find |
| W2-A4-009 | P2 | `backend/testframework/tms/importer_wave96_test.go`; `junit_parser_test.go`; `testrail_import_wave96_test.go:179–212` | Нет тестов на oversize / nested bomb / LimitReader для `handleImport` и JUnit/Allure parsers; TestRail handler — нет кейса >`maxTestRailImportBytes` → 413 | Регрессия W2-A4-001/003 не ловится; TestRail LimitReader есть в prod, но без negative-теста | Добавить 413 для TestRail; oversized/deep XML→error для JUnit/import |
| W2-A4-010 | P3 | `backend/internal/tms/testrail_import.go:279–294` | `parseTestRailCSV`: создаётся `csv.NewReader(bytes.NewReader(data))` и **не используется**; дальше только `tryReadCSV` | Мёртвый код / шум при чтении | Удалить неиспользуемый reader |
| W2-A4-011 | P3 | `backend/internal/tms/testrail_import.go:147–154` | `trimStepNumbering` компилирует два `regexp.MustCompile` на **каждый** вызов | Лишняя аллокация на больших CSV | Вынести `var` на уровень пакета (как `htmlTagRegexp`) |
| W2-A4-012 | P3 | `backend/internal/tms/importer.go:277–289` | Определение формата через `HasSuffix(name, ".csv")` и т.п. — **case-sensitive** | `cases.CSV` / `.XLSX` → «не удалось определить формат» при пустом `format` | `strings.ToLower(name)` перед суффиксами |
| W2-A4-013 | P3 | `backend/internal/tms/exporter.go:159–167` | `ExportRunResultsToCSV`: заголовки статистики через `_ = w.Write(...)` | Ошибка записи игнорируется; частичный/битый CSV без err | Проверять каждый `Write`, как в цикле результатов |

## Чисто

Файлы без находок — что проверялось:

- **allure_types.go** — только DTO Allure; path/секреты/IO нет; рекурсивные `Steps` — риск проявляется в парсере (W2-A4-003), не в типах.
- **export_hooks_converters.go / export_hooks_mongo.go / export_hooks_repo.go / export_hooks_testrail.go** — hubtest re-export; без IO/authz/parse логики; layering OK (`*ent.Client` только в test hook).
- **coverage_wave_test.go** — маппинг TestRail priority/type, stripHTML/steps, CombinedRepository/provisioner, traceability; asserts есть; не assert-free.
- **environment_wave105_test.go** — validation/404/compare env; вне фокуса import-bomb; mocks с asserts.
- **execute_by_a3017_test.go** — ExecutedByID на mock; mutation-комментарий есть; не scope A4 security.
- **exporter_test.go / exporter_wave95_test.go** — CSV/JSON/XML/Excel/PDF happy-path; ImportFromJSON/XML smoke без steps (дыра закрывается W2-A4-002).
- **import_run_scope_a3012_test.go** — service+mock: чужой `Project` edge → error; mock `GetTestRunInProject` сверяет project (см. `service_test.go:351–359`).
- **run_scope_a3012_ent_test.go** — Postgres `HasProjectWith` на `GetTestRunInProject` / `GetTestRunWithResultsInProject`; mutation documented.
- **rollback_a3007_ent_test.go / rollback_a3007_test.go** — version ownership Ent+mock; не import-фокус.

### Фокус-чеклист (итог)

| Тема | Вердикт |
|------|---------|
| LimitReader vs ReadAll | TestRail OK (`testrail_import.go:385–390`). **handleImport** — без LimitReader (W2-A4-001). Парсеры Allure/JUnit принимают уже `[]byte` — лимит на caller (webhook вне списка; там LimitReader есть). |
| Path traversal | **Чисто** в scope: `file.Filename` только для суффикса формата; имена экспорта хардкод (`test-cases.csv` и т.д.); `sheetName` в handleImport = `""`. `AllureAttachment.Source` не открывается как FS path в этих файлах. |
| A3-010 matcher project scope | **Исправлено** в prod. Остаток — mock-тесты не ловят снятие `projectID` (W2-A4-008). |
| XML/JSON bomb | **Живо** (W2-A4-003); + xlsx zip amplification через excelize (W2-A4-001). |
| A3-012 import/export run scope | **Исправлено**; тесты A4 покрывают mock+Ent. |
| TestRail secrets in logs | **Чисто**: `slog.Error` — section/title/row/error (`testrail_import.go:225–266`); API-токенов/паролей в модуле нет (файл CSV/XML, не TestRail API). |

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | 25/25/0 |
| Находок | **13** |
| P0 | 0 |
| P1 | 3 |
| P2 | 6 |
| P3 | 4 |

### Топ-3

1. **W2-A4-001** — `handleImport` без LimitReader / cap записей + excelize ZIP (P1 DoS).
2. **W2-A4-002** — steps CSV/JSON молча отбрасываются (P1 потеря данных при импорте).
3. **W2-A4-003** — unbounded JSON/XML unmarshal → bomb amplification (P1).
