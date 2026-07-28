# A2 W8 агент 5 — ent/schema part 5 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659f` · Baseline: [A8-5-schemas.md](A8-5-schemas.md)  
> Периметр: 15 схем · Прочитано: 15 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с Stage A (W8-A5-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W8-A5-001 | RESIDUAL | открыт | SystemSetting без version |
| W8-A5-002 | RESIDUAL | открыт | Team без version/deleted_at/updated_at |
| W8-A5-003 | FIXED-OK | закрыт | Mojibake TestCase comments (TASK-068) |
| W8-A5-004 | FIXED-OK | закрыт | `testcaseversion.go:53` unique `(test_case, version_number)` |
| W8-A5-005 | RESIDUAL | открыт | TestRun без version (и без deleted_at) |
| W8-A5-006 | FIXED-OK / RESIDUAL | частично | `deleted_at` есть; `version` нет |
| W8-A5-007 | FIXED-OK | закрыт | Mojibake TestRun allure comment |
| W8-A5-008 | RESIDUAL | открыт | User без version (lifecycle через status) |
| W8-A5-009 | RESIDUAL | открыт | WorkflowStatus без version/deleted_at |
| W8-A5-010 | FIXED-OK / RESIDUAL | частично | `deleted_at` + secret Sensitive; `version` нет |
| W8-A5-011 | FIXED-OK | закрыт | «Рёбра» godoc (TASK-068) |

**REGRESS:** 0 · **RESIDUAL:** 7 · **NEW:** 1 · **FIXED-OK (полные):** 4 (+2 partial)

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/systemsetting.go` | 31 | 1 | RESIDUAL version |
| `backend/ent/schema/team.go` | 39 | 1 | RESIDUAL OL/trash |
| `backend/ent/schema/teammember.go` | 49 | 0 | чисто (junction; M2M Team↔Project intentional) |
| `backend/ent/schema/testcase.go` | 131 | 0 | эталон version+deleted_at |
| `backend/ent/schema/testcaseversion.go` | 56 | 0 | FIXED unique |
| `backend/ent/schema/testresult.go` | 93 | 0 | чисто (append) |
| `backend/ent/schema/testrun.go` | 91 | 1 | RESIDUAL version |
| `backend/ent/schema/testsuite.go` | 64 | 2 | RESIDUAL version + NEW no unique name |
| `backend/ent/schema/timelog.go` | 60 | 0 | чисто |
| `backend/ent/schema/user.go` | 154 | 1 | RESIDUAL version |
| `backend/ent/schema/userblock.go` | 57 | 0 | чисто |
| `backend/ent/schema/webhookdelivery.go` | 59 | 0 | чисто (append log) |
| `backend/ent/schema/webhookendpoint.go` | 62 | 1 | RESIDUAL version (deleted_at FIXED) |
| `backend/ent/schema/workflowstatus.go` | 67 | 1 | RESIDUAL version/trash |
| `backend/ent/schema/workflowtransition.go` | 48 | 0 | чисто (junction) |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки (открытые A2)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W8-5-001 | RESIDUAL (W8-A5-008) | P1 | `user.go:19-77` | Mutable User без `version` | Lost update профиля/custom_fields/quiet_hours | `version` + VersionEQ |
| A2W8-5-002 | RESIDUAL (W8-A5-002) | P1 | `team.go:18-27` | Нет version/deleted_at/updated_at | Mutable user-visible entity | Три поля + soft-delete |
| A2W8-5-003 | RESIDUAL (W8-A5-005) | P1 | `testrun.go:19-59` | Нет version (и deleted_at) | Concurrent counter/status UpdateOneID | version + VersionEQ; delete policy ADR |
| A2W8-5-004 | RESIDUAL (W8-A5-006) | P1 | `testsuite.go:19-39` | Soft-delete есть, version нет | Concurrent rename/sort без CAS | version |
| A2W8-5-005 | RESIDUAL (W8-A5-009/010) | P1 | `workflowstatus.go`; `webhookendpoint.go` | WS: no version/deleted_at; WE: deleted_at OK, no version | Hard delete status; concurrent webhook mutate | version + soft-delete status |
| A2W8-5-006 | RESIDUAL (W8-A5-001) | P2 | `systemsetting.go:17-28` | Mutable KV без version | Concurrent upsert LWW | version + VersionEQ |
| A2W8-5-007 | NEW | P2 | `testsuite.go:56-62` | Soft-delete добавлен, но нет unique `(project, name)` (в отличие от Component/Environment после BE-22) | Дубликаты имён папок в проекте на уровне БД | `index.Edges("project").Fields("name").Unique()` + partial WHERE deleted_at IS NULL |
| A2W8-5-008 | RESIDUAL (P1-BE-U06) | P2 | A5 mutable set | Системный VersionMixin gap | ADR | ADR |

## Матрица фокуса (A2)

| Схема | UUID | version | deleted_at | Вердикт |
|-------|:----:|:-------:|:----------:|---------|
| SystemSetting | ✓ | ✗ | ✗ | A2W8-5-006 |
| Team | ✓ | ✗ | ✗ | A2W8-5-002 |
| TeamMember | ✓ | — | — | чисто |
| TestCase | ✓ | ✓ | ✓ | эталон |
| TestCaseVersion | ✓ | — | — | FIXED unique |
| TestResult | ✓ | — | — | чисто |
| TestRun | ✓ | ✗ | ✗ | A2W8-5-003 |
| TestSuite | ✓ | ✗ | ✓ | A2W8-5-004/007 |
| TimeLog | ✓ | — | — | чисто |
| User | ✓ | ✗ | status | A2W8-5-001 |
| UserBlock | ✓ | — | — | чисто |
| WebhookDelivery | ✓ | — | — | чисто |
| WebhookEndpoint | ✓ | ✗ | ✓ | A2W8-5-005 |
| WorkflowStatus | ✓ | ✗ | ✗ | A2W8-5-005 |
| WorkflowTransition | ✓ | — | — | чисто |

## Чисто

- **testcase.go** — UUID+version+deleted_at+project composite (A8-002).
- **testcaseversion.go** — unique version_number FIXED.
- **testresult / timelog / userblock / webhookdelivery / workflowtransition / teammember** — junction/append OK.
- Mojibake A5 — FIXED (TASK-068).

## Сводка counts (открытые)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 5 |
| P2 | 3 |
| P3 | 0 |
| **Total open** | **8** |

FIXED-OK (полные): 4 · NEW: 1 · REGRESS: 0
