# W2 — TMS (`backend/internal/tms/`)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Агент:** be-ver (Grok 4.5) · **Дата:** 2026-07-23 · **Режим:** readonly repo-scan
> **Scope:** layering, RBAC/Fiber v3, project-scope IDOR, Allure/Mongo, optimistic lock, NATS, тесты `testframework/tms`
> **Код не менялся.**

## Вердикт (кратко)

- **Layering:** OK — `*ent.Client` только в `ent_repository.go` (+ hubtest `export_hooks_repo.go`); service зависит от `Repository`.
- **RBAC route-level (FA-02-02):** OK — `RegisterRoutes` монтирует `rbac.RequirePermission` после handler; вебхуки закрыты `PermTestRunCreate`; есть denial-тесты.
- **Project-scope на ресурсных Get/Update/Delete:** **FAIL** — системный IDOR (аналог NF-03 / bases `authorizeBase`).
- **Allure/Mongo:** присутствует; лимиты multipart OK; raw dump best-effort.
- **Тесты:** **46** файлов в `backend/testframework/tms/` (широкое coverage-волновое покрытие); **нет** suite на cross-project IDOR.
- **Регрессия в этом прогоне:** не запускалась (scan-only; Ask/readonly). Доменный канон: `go test -tags hubtest -race ./testframework/tms/...`.
- **verified:** `false` — блокируют CRITICAL/HIGH ниже.

---

## Findings

| Sev | Файл:строка | Evidence | Fix |
|-----|-------------|----------|-----|
| **CRITICAL** | `service_test_cases.go:64-74`, `116-194`; `handler_test_cases.go:88-99` | `Get/Update/DeleteTestCase` ищут кейс **только** по `caseKey` (`FindTestCaseByKey` — глобальный query без `projectID`). Handler парсит `projectId` для Create/List, но **не передаёт** его в Get/Update/Delete. RBAC проверяет право на **project из URL**, не принадлежность кейса. Участник проекта A с `PermProjectView`/`PermTestCaseEdit` может читать/менять/удалять кейс проекта B, зная `caseKey` (`PROJ-TC-N`). | Прокинуть `projectID` во все case-методы; после load — `if tc.ProjectID != projectID { return NotFound/Forbidden }`. Эталон: `bases/service.go` `authorizeBase`. Тест: member A + кейс B → 403/404. |
| **CRITICAL** | `service_runs.go:37-123`, `194-317`; `handler_runs.go:83-95` | `Get/Update/Reopen/Start/Complete/AddCases/Assign/Detail` работают по `runID` **без** сверки `run.ProjectID` с `projectId` из пути. Тот же класс IDOR на мутациях прогона. | `authorizeRun(ctx, projectID, runID)` до любой мутации/чтения; 404 при чужом проекте. |
| **CRITICAL** | `service_suites.go:49-114`; `environment.go:67-94`, `181-223` | `Update/Delete Suite|Component|Environment` — только по UUID ресурса. `UpdateEnvironmentHandler`/`DeleteEnvironmentHandler` **даже не парсят** `projectId`. | Load + `entity.ProjectID == pathProjectID` (или edge `HasProjectWith`); для env — передавать `projectID` в service. |
| **CRITICAL** | `service_import.go:139-209`; `service_reports.go:94-120`; `handler_reports.go:128-149` | `ImportAllureToExistingRun` не сверяет `run` с `projectID` — можно дописать Allure/результаты в чужой прогон при праве Create на «своём» project в URL. `GetRunComparison` игнорирует `projectId` (не парсит), сравнивает любые `run1`/`run2` при праве View на любом доступном проекте в пути. | Проверка `run.ProjectID`; comparison — оба прогона принадлежат `projectID`. |
| **HIGH** | `ent_repository.go:246-310`; `repository.go:223`; `service_test_cases.go:148` | Поле `UpdateTestCaseInput.Version` / `req.Version` (**required** в DTO) **не используется** в update: нет `Where(testcase.VersionEQ(...))`, только слепой `AddVersion(1)`. Optimistic lock TZ **не enforced** — тихая перезапись при конкурентных PUT (эталон tracker: `repo_issues.go` `VersionEQ`). | `Update().Where(VersionEQ(input.Version)).AddVersion(1)`; при 0 rows → `ErrVersionConflict` + тест. |
| **HIGH** | `service_test_cases.go:197-228` | `RollbackTestCase`: `GetTestCaseVersion(versionID)` без проверки, что версия принадлежит загруженному кейсу — можно применить чужой snapshot к другому кейсу (при знании UUID). | `version.TestCaseID == existing.ID` иначе 404. |
| **HIGH** | `service_runs.go:213-267` | `ExecuteTestResult`: обновляет `resultID` без проверки принадлежности к `runID` (и без project-scope). | Verify `result.RunID == runID` (+ project). |
| **HIGH** | `testframework/tms/` | Есть `handler_webhook_rbac_test.go` (outsider → 403 на create/webhook) — **закрыт только membership на path-project**. Нет тестов cross-project resource IDOR (кейсы/раны/сьюты/env/allure-into-run/comparison). | Denial-suite по образцу `testframework/bases/rbac_test.go`. |
| **MEDIUM** | `service_import.go:128`; `service_runs.go:309`; `events.go:7-15` | Прямой `nats.PublishJSON` с subject `tms.run.completed.%s` — не outbox, не канон `hub.<domain>.<event>`; константы `events.go` (`tms.testrun.completed`) **не используются** публикаторами. Потеря атомарности / drift контракта. | Outbox + единый subject; выровнять `events.go` ↔ publish. |
| **MEDIUM** | `service_import.go:52-56`, `159-162` | Ошибка Mongo `StoreAllureResults` только логируется — импорт идёт дальше (осознанно). Риск «дыр» в raw-аудите без сигнала клиенту. | Метрика/флаг в ответе `mongo_stored: false` или fail policy по ADR. |
| **MEDIUM** | `ent_repository.go:420-421`, `468`, `976` | Hard delete suite/component/environment (`DeleteOneID`) vs soft-delete кейсов (`SetDeletedAt` UTC). Расхождение с TZ soft-delete (корзина 30д), если сущности user-visible. | Soft-delete или явный ADR «hard OK для справочников». |
| **MEDIUM** | `service_test_cases.go:56-59` (и др.) | Много `NewInternal("...")` **без** `%w` исходной ошибки — теряется triage/observability на границе service. | `fmt.Errorf("...: %w", err)` + map в AppError. |
| **MEDIUM** | `mongo_repository.go:78` | `time.Now()` без `.UTC()` для `CreatedAt` документов Allure (кейсы — UTC). | `time.Now().UTC()`. |
| **LOW** | `handler_webhook.go:17-19` vs `handler.go:109-117` | Комментарий про `X-Hub-Token`/query `token`; фактически путь = `authMw` + RBAC (JWT/API-key как в остальном API). Документационный drift. | Синхронизировать swagger/комментарии с реальной auth-цепочкой. |
| **LOW** | `combined_repository.go:21-22` | Dummy `var _ = schema.TestStep{}` для подавления unused import. | Убрать/перенести import. |

---

## Чек-лист be-ver (сводка)

| Пункт | Статус |
|-------|--------|
| Layering `*ent.Client` вне repository | OK (`ent_repository.go:24`; export_hooks — hubtest) |
| Handler без прямого ent | OK |
| Fiber v3 RBAC после handler | OK (`handler.go:30-124`, комментарий FA-02-02) |
| Webhook RBAC | OK + тесты `handler_webhook_rbac_test.go` |
| Owner/project-scope Get+mutate | **FAIL** (CRITICAL cluster) |
| `io.ReadAll` + LimitReader (multipart Allure/JUnit) | OK (`handler_webhook.go:86`, `182`; TestRail similarly) |
| `http.DefaultClient` / SSRF | OK (нет исходящих user-URL в модуле) |
| `panic` в prod-пути | OK (не найдено) |
| Optimistic `version` | **FAIL** (HIGH) |
| Counter в Tx | OK (`CreateTestCase` Tx + `AddTestCaseCounter`) |
| Mongo Allure | OK структура (`mongo_repository.go`, `CombinedRepository`) |
| Matcher scoped by projectID | OK (`allure_matcher.go` — match в рамках project) |
| Тесты домена | OK наличие (~46 файлов, COV tms **82.0%** в COVERAGE.md); gap — IDOR suite |

---

## Тесты — наличие

| Область | Файлы (примеры) |
|---------|-----------------|
| Handler/routes/RBAC | `handler_routes_test.go`, `handler_webhook_rbac_test.go`, `handler_flow_test.go`, `handler_wave*.go` |
| Service | `service_test.go`, `service_suites_test.go`, `service_runs_extra_test.go`, `service_wave*.go` |
| Ent/Mongo repo | `ent_repository_*.go`, `mongo_repository_test.go` |
| Allure/JUnit/TestRail | `allure_parser_test.go`, `allure_matcher_wave99_test.go`, `junit_parser_test.go`, `testrail_import_wave96_test.go` |
| Import/export | `handler_import_flow_test.go`, `importer_wave96_test.go`, `exporter_*.go` |
| Environments | `environment_wave105_test.go` |

**Пробел:** нет явного теста «ресурс проекта B через URL проекта A».

---

## Воспроизведение CRITICAL (минимальный сценарий)

1. User U — member проекта A (`PermTestCaseEdit`), **не** member проекта B.
2. В B существует кейс `B-TC-1` (или известный UUID run/suite/env).
3. `GET|PUT|DELETE /api/projects/{A}/tms/cases/B-TC-1` с JWT U.
4. RBAC: View/Edit на A → **проходит**.
5. Service: `FindTestCaseByKey("B-TC-1")` → возвращает/мутирует кейс B.

Аналогично: `PUT .../A/tms/runs/{runB}/status`, `PUT .../A/environments/{envB}`, `POST .../A/tms/runs/{runB}/allure`.

---

## Вердикт

```
## Верификация (BE): internal/tms
- Воспроизведено: трассировка Handler→Service→Repo для cases/runs/suites/env/allure/comparison (код); runtime curl не гонялся (scan).
- Layering: OK — ent.Client только EntRepository.
- Security-инварианты: RBAC route OK; project-scope resource IDOR FAIL (CRITICAL); optimistic lock FAIL (HIGH).
- Allure/Mongo: OK с оговорками (best-effort Mongo, UTC).
- Регрессия: не запускалась в scan; наличие testframework/tms — да; gap — IDOR denial.
- Замечания: см. таблицу Findings.
- Вердикт: verified: false — 2026-07-23 00:56 UTC+3
```

**Handoff:** `be1`/`be2` — фикс authorize* + VersionEQ + тесты; затем повторный be-ver; при CRITICAL security — опционально `security-auditor`.
