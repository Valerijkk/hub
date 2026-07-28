# A3 W2 агент 1 — tms HTTP handlers (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Mirror: [A2W2-1-tms-handlers.md](A2W2-1-tms-handlers.md) · Assignment: [W2-ASSIGNMENTS.md](W2-ASSIGNMENTS.md) § A1  
> Периметр: 24 файла (8 prod + 16 tests) · Прочитано: **24/24** · Пропущено: **0**

## Сверка с A2 (A2W2-1-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W2-1-001 | FIXED-OK | закрыт | `CreateBugFromResult` парсит `:runId` → `CreateBugFromFailedResult(ctx, projectID, runID, …)`; service `GetTestRunInProject` + `GetTestResultInRun` — `handler_test_cases.go:347–366`; TASK-091 |
| A2W2-1-002 | FIXED-OK | закрыт | Swagger: «ZIP не поддерживаются» — `handler_webhook.go:41–42`; только JSON unmarshal |
| A2W2-1-003 | FIXED-OK | закрыт | wave124 шлёт `version`; StructValidator PUT без version → 400 — `handler_wave124_test.go:34–41`; `task111_w2_residual_test.go` (вне слота) |
| A2W2-1-004 | FIXED-OK | закрыт | Foreign run denied — `task091_createbug_reports_test.go` (service); prod binding закрыт A2W2-1-001. wave88 остаётся happy-path mock |
| A2W2-1-005 | FIXED-OK | закрыт | Raw body cap 5 MB → 413 — `handler_webhook.go:101–105`, `201–205`; `TestTask111_WebhookRawBodyCap` |

### Cascade FIXED-OK (Stage A / W18, без регрессии)

| Stage A / prior | Evidence |
|-----------------|----------|
| W2-A1-001/002/003 suite/component IDOR | `handler_suites.go:72–193` + `projectID` в Update/Delete |
| W2-A1-005 StructValidator | DTO `validate` теги; wiring вне слота |
| W2-A1-011 error leak webhook | фиксированные `"невалидный JSON"` / `"невалидный JUnit XML"` + slog |

**REGRESS:** 0 · **RESIDUAL (этот слот):** 0 · **NEW (этот слот):** 1

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/tms/handler.go` | 134 | 0 | чисто — Fiber v3 RBAC after handler; webhook `PermTestRunCreate` |
| `backend/internal/tms/handler_import.go` | 170 | 0 | чисто — ParseProjectID; run scope на export/allure-to-run |
| `backend/internal/tms/handler_reports.go` | 235 | 0 | чисто — projectID; UUID run/env; limit clamp в service ≤100 |
| `backend/internal/tms/handler_runs.go` | 419 | 0 | чисто — project+run+result scope на CRUD/execute |
| `backend/internal/tms/handler_suites.go` | 194 | 1 | NEW — UpdateSuite без OL (см. dto) |
| `backend/internal/tms/handler_test_cases.go` | 372 | 0 | FIXED-OK CreateBug runId |
| `backend/internal/tms/handler_webhook.go` | 257 | 0 | FIXED-OK zip docs + body cap + safe errors |
| `backend/internal/tms/dto.go` | 471 | 1 | NEW — `UpdateSuiteRequest`/`SuiteResponse` без `version` |
| `backend/testframework/tms/handler_flow_test.go` | 665 | 0 | mock happy-path; Ent-neg вне файла |
| `backend/testframework/tms/handler_import_flow_test.go` | 174 | 0 | чисто |
| `backend/testframework/tms/handler_routes_test.go` | 214 | 0 | RegisterRoutes + RBAC owner path |
| `backend/testframework/tms/handler_wave108_test.go` | 76 | 0 | invalid JSON/UUID |
| `backend/testframework/tms/handler_wave110_test.go` | 134 | 0 | чисто |
| `backend/testframework/tms/handler_wave113_test.go` | 75 | 0 | чисто |
| `backend/testframework/tms/handler_wave114_test.go` | 40 | 0 | чисто |
| `backend/testframework/tms/handler_wave115_test.go` | 48 | 0 | чисто |
| `backend/testframework/tms/handler_wave120_test.go` | 91 | 0 | чисто |
| `backend/testframework/tms/handler_wave122_test.go` | 99 | 0 | чисто |
| `backend/testframework/tms/handler_wave124_test.go` | 56 | 0 | FIXED-OK version на PUT |
| `backend/testframework/tms/handler_wave88_test.go` | 150 | 0 | FIXED-OK gap закрыт TASK-091 вне файла |
| `backend/testframework/tms/handler_webhook_flow_test.go` | 110 | 0 | empty/invalid/success |
| `backend/testframework/tms/handler_webhook_rbac_test.go` | 172 | 0 | outsider 403 |
| `backend/testframework/tms/handler_webhook_wave99_test.go` | 31 | 0 | multipart missing file |
| `backend/testframework/tms/helpers_test.go` | 132 | 0 | converters |

Файлов в списке: **24**. Прочитано: **24**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W2-1-001 | NEW | P3 | `dto.go:73–78`, `251–261`; `handler_suites.go:72–92` | После TASK-095 `TestSuite` имеет `VersionMixin`, но API: `UpdateSuiteRequest` без `version`, `SuiteResponse` не отдаёт `version`; repo `UpdateSuite` без `VersionEQ`/`AddVersion` | NFR-24 / ADR-ENT-VERSION-MIXIN: mutable entity с version в схеме, update — last-write-wins; клиент не может сделать OL (контраст с `UpdateTestCaseRequest.Version`) | `version` в Request+Response; validate required; service→repo `VersionEQ`+`AddVersion` → 409; regression concurrent PUT |

## Чисто (перепроверено)

- **RBAC Fiber v3:** handler затем `RequirePermission`; webhook = `PermTestRunCreate` (FA-02-02).
- **IDOR:** все мутации/чтения с `ParseProjectID`; suites/components/runs/results/CreateBug — scoped service calls.
- **Secrets in logs:** webhook slog только `error` парсера, не body/token.
- **Soft-delete:** DeleteTestCase/Suite/Component делегированы service→repo soft path (корзина).
- **Context:** `c.Context()` / UserIDFromCtx на мутациях с actor.
- **OL test-cases:** DTO `Version` required; update path VersionEQ (service/repo вне слота).
- **N+1 / limit:** report `limit` нормализуется в service (≤100); list DTO max=100.
- **Webhook:** multipart LimitReader 5 MB + raw body cap; zip убран из контракта.
- **Тесты:** RBAC outsider; CreateBug foreign-run (TASK-091); StructValidator version; body 413.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 1 |
| **Total** | **1** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active) | 5 |
| REGRESS | 0 |
| RESIDUAL | 0 |
| NEW | 1 (`A3W2-1-001`) |

**Топ ID:** A3W2-1-001 (suite Update без OL при VersionMixin).
