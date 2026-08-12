# A2 · W2 агент A1 — tms handlers (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Зеркало Stage A: [A2-1-tms-handlers.md](A2-1-tms-handlers.md) · TASK-015/016/036/037 W18 PASS  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/tms/handler.go` | 113 | 0 | чисто |
| `backend/internal/tms/handler_import.go` | 148 | 0 | чисто |
| `backend/internal/tms/handler_reports.go` | 203 | 0 | чисто |
| `backend/internal/tms/handler_runs.go` | 369 | 0 | чисто |
| `backend/internal/tms/handler_suites.go` | 167 | 0 | FIXED-OK IDOR |
| `backend/internal/tms/handler_test_cases.go` | 324 | 1 | RESIDUAL CreateBug |
| `backend/internal/tms/handler_webhook.go` | 214 | 2 | RESIDUAL swagger zip + NEW body cap |
| `backend/internal/tms/dto.go` | 416 | 0 | FIXED-OK validator wired |
| `backend/testframework/tms/handler_flow_test.go` | 571 | 0 | mock shape; Ent-neg вне файла |
| `backend/testframework/tms/handler_import_flow_test.go` | 150 | 0 | чисто |
| `backend/testframework/tms/handler_routes_test.go` | 181 | 0 | чисто |
| `backend/testframework/tms/handler_wave108_test.go` | 64 | 0 | RESIDUAL gap закрыт `project_scope_*` |
| `backend/testframework/tms/handler_wave110_test.go` | 113 | 0 | чисто |
| `backend/testframework/tms/handler_wave113_test.go` | 63 | 0 | чисто |
| `backend/testframework/tms/handler_wave114_test.go` | 34 | 0 | чисто |
| `backend/testframework/tms/handler_wave115_test.go` | 40 | 0 | чисто |
| `backend/testframework/tms/handler_wave120_test.go` | 76 | 0 | чисто |
| `backend/testframework/tms/handler_wave122_test.go` | 81 | 0 | чисто |
| `backend/testframework/tms/handler_wave124_test.go` | 45 | 1 | RESIDUAL mock без version |
| `backend/testframework/tms/handler_wave88_test.go` | 131 | 1 | RESIDUAL CreateBug mock |
| `backend/testframework/tms/handler_webhook_flow_test.go` | 94 | 0 | чисто |
| `backend/testframework/tms/handler_webhook_rbac_test.go` | 149 | 0 | чисто |
| `backend/testframework/tms/handler_webhook_wave99_test.go` | 26 | 0 | чисто |
| `backend/testframework/tms/helpers_test.go` | 119 | 0 | чисто |

Файлов в списке: **24**. Прочитано: **24**. Пропущено: **0**.

## FIXED-OK (Stage A → закрыто после B/C)

| Stage A ID | Evidence |
|------------|----------|
| W2-A1-001/002/003 | `handler_suites.go:72–189` парсит `projectID` → `Update/DeleteSuite/DeleteComponent(ctx, projectID, …)`; Ent `HasProjectWith`; W18-C2 TASK-015 PASS |
| W2-A1-005 | `wiring_fiber.go` `StructValidator: middleware.NewStructValidator()`; DTO `validate` теги исполняются на prod Bind |
| W2-A1-007/008 | Ent-negative: `project_scope_idor_p1be10_12_test.go`, `authz_ent_negative_task037_test.go` (вне списка A1, но закрывают gap) |
| W2-A1-011 | `handler_webhook.go:109–112`, `204–208` — фиксированные `"невалидный JSON"` / `"невалидный JUnit XML"` + slog |

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W2-1-001 | RESIDUAL (W2-A1-004) | P2 | `handler_test_cases.go:342–361` | `CreateBugFromResult` не парсит `:runId`; service `CreateBugFromFailedResult(ctx, projectID, userID, resultID, …)` | URL-прогон и result не связаны; в рамках проекта баг по любому `resultId` при произвольном `runId` | Парсить `runID`; service+repo `HasRunWith(IDEQ(runID))` |
| A2W2-1-002 | RESIDUAL (W2-A1-006) | P2 | `handler_webhook.go:41` (+ `:109`) | Swagger/описание обещает zip allure-results; код всегда `json.Unmarshal` | CI с zip → 400 «невалидный JSON» | Распаковать zip **или** убрать zip из swagger |
| A2W2-1-003 | RESIDUAL (W2-A1-009) | P2 | `handler_wave124_test.go` | PUT case без `version` на mock → успех | Закрепляет обход validator на mock-app без StructValidator | Подключать validator в test app **или** Ent-тест version required → 400 |
| A2W2-1-004 | RESIDUAL (W2-A1-010) | P2 | `handler_wave88_test.go` | CreateBug mock: нет кейса «чужой runId в URL» | Не ловит A2W2-1-001 | Assert result ∈ URL-run → 404 |
| A2W2-1-005 | NEW | P3 | `handler_webhook.go:98–101`, `195–197` | Multipart: `LimitReader(…, maxAllureFileBytes+1)` (5 MB); raw JSON/XML body — `c.Body()` без того же cap (только Fiber BodyLimit 25 MB) | Асимметрия: один и тот же webhook принимает до 25 MB JSON vs 5 MB file | Cap body: `if len(c.Body()) > maxAllureFileBytes → 413` |

## Чисто (перепроверено)

- **handler.go** — Fiber v3 RBAC after handler; env routes; webhook `PermTestRunCreate`.
- **handler_import/reports/runs** — `ParseProjectID` + run/result scope на CRUD/execute/compare.
- **handler_suites** — project scope на всех мутациях suite/component.
- **Webhook RBAC/flow** — outsider 403; empty/invalid asserts.

## Сводка

| | Count |
|--|------:|
| Files | 24/24/0 |
| FIXED-OK classes | 4 |
| NEW | 1 |
| REGRESS | 0 |
| RESIDUAL | 4 |
| Open P0/P1/P2/P3 | 0 / 0 / 4 / 1 |
