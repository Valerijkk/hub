# W8 агент A5 — schemas part 5

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: UUID v4 PK · `field.Int("version")` optimistic lock · soft-delete `deleted_at` (корзина 30д) · `helpers.go` (вне списка A5 — у A2).  
> Сверка с `grok-2026-07` A8: A8-002 индекс TestCase **FIXED**; A8-004 / A8-015 **живут** в этом срезе.

## Сводка

| Метрика | Значение |
|--------:|----------|
| Файлов в списке | 15 |
| Прочитано | 15 |
| Пропущено | 0 |
| Находок | **11** |
| P0 / P1 / P2 / P3 | 0 / **6** / **2** / **3** |
| UUID PK `field.UUID("id")` | **15/15** |
| Mutable с `version` | 1/8 целевых (`TestCase`) |
| Soft-delete `deleted_at` | 1/15 (`TestCase`); junction/append-only — N/A |
| `helpers.go` в scope A5 | нет (A2); схемы A5 используют `nowUTC()` |

### Топ-3

1. **W8-A5-008** (P1) — `user.go`: mutable User без `version` (A8-004 residual).
2. **W8-A5-006** (P1) — `testsuite.go`: нет `version`/`deleted_at`; hard `DeleteOneID` (A8-015 residual).
3. **W8-A5-010** (P1) — `webhookendpoint.go`: нет `version`/`deleted_at`; hard delete.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/ent/schema/systemsetting.go` | 30 | 1 | замечания |
| `backend/ent/schema/team.go` | 38 | 1 | замечания |
| `backend/ent/schema/teammember.go` | 48 | 0 | чисто |
| `backend/ent/schema/testcase.go` | 130 | 1 | замечания |
| `backend/ent/schema/testcaseversion.go` | 54 | 1 | замечания |
| `backend/ent/schema/testresult.go` | 92 | 0 | чисто |
| `backend/ent/schema/testrun.go` | 90 | 2 | замечания |
| `backend/ent/schema/testsuite.go` | 57 | 1 | замечания |
| `backend/ent/schema/timelog.go` | 59 | 0 | чисто |
| `backend/ent/schema/user.go` | 153 | 1 | замечания |
| `backend/ent/schema/userblock.go` | 56 | 0 | чисто |
| `backend/ent/schema/webhookdelivery.go` | 58 | 0 | чисто |
| `backend/ent/schema/webhookendpoint.go` | 55 | 1 | замечания |
| `backend/ent/schema/workflowstatus.go` | 66 | 1 | замечания |
| `backend/ent/schema/workflowtransition.go` | 47 | 0 | чисто |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Перепроверка прошлого аудита (A8 / grok-2026-07)

| Старый ID | Статус в текущем коде | Комментарий |
|-----------|----------------------|-------------|
| A8-002 | **исправлено** — не дублировать | `testcase.go:127-128` — `index.Edges("project").Fields("deleted_at")` |
| A8-004 | **живо** → W8-A5-008 (+ соседние mutable без version) | User и др. без `field.Int("version")` |
| A8-015 | **живо** → W8-A5-006 | `testsuite.go` без `deleted_at`; hard-delete в `tms/ent_repository.go:445` |
| A8-001 / A8-003 / A8-005 / A8-006 | вне scope A5 | Issue/Project/Channel/Comment — другие агенты W8 |

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W8-A5-001 | P2 | `backend/ent/schema/systemsetting.go:17-28` | Mutable SystemSetting без `version` | TZ optimistic lock; concurrent upsert settings → LWW | `version` + VersionEQ; soft-delete — ADR-исключение или `deleted_at` |
| W8-A5-002 | P1 | `backend/ent/schema/team.go:18-27` | Team без `version` / `deleted_at` / `updated_at` | Mutable user-visible entity против TZ | Добавить три поля + soft-delete в project layer |
| W8-A5-003 | P3 | `backend/ent/schema/testcase.go:63-66` | Mojibake в Comment: `Рмя класса/метода` | UTF-8 порча (TASK-007) | Переписать «Имя…» |
| W8-A5-004 | P2 | `backend/ent/schema/testcaseversion.go:50-53` | Нет unique `(test_case, version_number)` | Дубликаты снимков при гонке CreateTestCaseVersion | `index.Edges("test_case").Fields("version_number").Unique()` |
| W8-A5-005 | P1 | `backend/ent/schema/testrun.go:19-59` | TestRun без `version` (и без `deleted_at`) | Concurrent counter/status update без CAS (`UpdateOneID` в tms repo) | `version` + VersionEQ; политика delete — ADR |
| W8-A5-006 | P1 | `backend/ent/schema/testsuite.go:19-34` | TestSuite без `version`/`deleted_at` (A8-015); hard DeleteOneID | TZ soft-delete нарушен | Soft-delete + version; фильтр `DeletedAtIsNil` |
| W8-A5-007 | P3 | `backend/ent/schema/testrun.go:47` | Mojibake Comment allure_run_id (`Рдентификатор…`) | TASK-007 | «Идентификатор…» |
| W8-A5-008 | P1 | `backend/ent/schema/user.go:19-77` | User без `version` (A8-004); lifecycle через `status`, не `deleted_at` | Lost update профиля/custom_fields | `version` + VersionEQ на UpdateUserProfile |
| W8-A5-009 | P1 | `backend/ent/schema/workflowstatus.go:19-39` | WorkflowStatus без `version`/`deleted_at`; hard DeleteOneID | Concurrent rename/WIP; FK risk при issues | version + soft-delete или запрет delete при issues |
| W8-A5-010 | P1 | `backend/ent/schema/webhookendpoint.go:19-35` | WebhookEndpoint без `version`/`deleted_at`; hard DeleteOneID | TZ; необратимый revoke | version + soft-delete (рядом с `is_active`) |
| W8-A5-011 | P3 | `backend/ent/schema/testcase.go:94` (+ `testcaseversion.go:36`, `testrun.go:63`, `userblock.go:34`, `webhookdelivery.go:42`, `workflowstatus.go:43`) | Godoc edges: «ибра» вместо «Рёбра» | Косметика | Унифицировать `// Рёбра (edges)…` |

## Матрица фокуса (UUID / version / soft-delete)

| Схема | UUID PK | version | deleted_at | Класс | Вердикт фокуса |
|-------|:-------:|:-------:|:----------:|-------|----------------|
| SystemSetting | ✓ | ✗ | ✗ | mutable KV | W8-A5-001 |
| Team | ✓ | ✗ | ✗ | mutable | W8-A5-002 |
| TeamMember | ✓ | — | — | junction | чисто |
| TestCase | ✓ | ✓ | ✓ | mutable + trash | эталон (+ mojibake P3) |
| TestCaseVersion | ✓ | — (`version_number`) | — | append snapshot | W8-A5-004 unique |
| TestResult | ✓ | — | — | append | чисто |
| TestRun | ✓ | ✗ | ✗ | mutable | W8-A5-005 |
| TestSuite | ✓ | ✗ | ✗ | mutable | W8-A5-006 |
| TimeLog | ✓ | — | — | append (нет Delete API) | чисто |
| User | ✓ | ✗ | — (`status`) | mutable | W8-A5-008 |
| UserBlock | ✓ | — | — | junction | чисто |
| WebhookDelivery | ✓ | — | — | append log | чисто |
| WebhookEndpoint | ✓ | ✗ | ✗ | mutable | W8-A5-010 |
| WorkflowStatus | ✓ | ✗ | ✗ | mutable | W8-A5-009 |
| WorkflowTransition | ✓ | — | — | junction | чисто |

## Чисто

| Файл | Что проверялось |
|------|-----------------|
| `teammember.go` | UUID PK; Unique(team,user); Required edges; junction — version/soft-delete N/A |
| `testresult.go` | UUID PK; append-only; status enum; индексы status/allure/flaky |
| `timelog.go` | UUID PK; Required issue+user; Positive minutes; Delete API нет — soft-delete N/A |
| `userblock.go` | UUID PK; Unique(blocker,blocked); Field edges; unblock = hard remove пары — ок |
| `webhookdelivery.go` | UUID PK; Required endpoint; append-only audit log |
| `workflowtransition.go` | UUID PK; Unique(from,to); junction |
| `testcase.go` (ядро) | UUID + `version` + `deleted_at` + composite index project+deleted_at (A8-002 FIXED) |

## helpers.go

Не в списке A5 (агент **A2**). Timestamp-поля A5 через `nowUTC()` из `backend/ent/schema/helpers.go` — UTC default соблюдён. Отдельных находок по helpers нет.

## Evidence (repo, read-only)

- Soft-delete TestCase: `backend/internal/tms/ent_repository.go:334-338` `SetDeletedAt`
- Hard-delete TestSuite: `…:443-445` `DeleteOneID`
- Hard-delete WorkflowStatus: `backend/internal/tracker/repo_workflow.go:53-55`
- Hard-delete WebhookEndpoint: `backend/internal/apigateway/repository.go:168-169`
