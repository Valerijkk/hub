# A2 W8 агент 5 — ent/schema part 5 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `8d9c0950` · Baseline: [A8-5-schemas.md](A8-5-schemas.md)  
> Периметр: 15 схем · Прочитано: 15 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с Stage A (W8-A5-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W8-A5-001 | RESIDUAL | открыт | `systemsetting.go` — mutable KV без `version` |
| W8-A5-002 | RESIDUAL | открыт | `team.go` — нет version/deleted_at/updated_at |
| W8-A5-003 | FIXED-OK | закрыт | testcase Allure comments UTF-8 «Имя…» (TASK-068) |
| W8-A5-004 | FIXED-OK | закрыт | `testcaseversion.go:53` unique `(test_case, version_number)` |
| W8-A5-005 | RESIDUAL | открыт | `testrun.go` — нет `version` (mutable counters/status) |
| W8-A5-006 | RESIDUAL | частично | `deleted_at` FIXED (TASK-039); `version` всё ещё нет |
| W8-A5-007 | FIXED-OK | закрыт | testrun Allure comment «Идентификатор…» |
| W8-A5-008 | RESIDUAL | открыт | `user.go` — mutable User без `version` (A8-004) |
| W8-A5-009 | RESIDUAL | открыт | WorkflowStatus без version/deleted_at; hard Delete |
| W8-A5-010 | RESIDUAL | частично | `deleted_at` FIXED (TASK-039); `version` нет |
| W8-A5-011 | FIXED-OK | закрыт | godoc «Рёбра» / mojibake (TASK-068); grep `ибра` = 0 |

**REGRESS:** 0 · **RESIDUAL (open):** 7 · **NEW:** 0 · **FIXED-OK:** 4

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/systemsetting.go` | 30 | 1 | RESIDUAL version |
| `backend/ent/schema/team.go` | 38 | 1 | RESIDUAL OL/trash |
| `backend/ent/schema/teammember.go` | 48 | 0 | чисто (junction) |
| `backend/ent/schema/testcase.go` | 130 | 0 | эталон + mojibake FIXED |
| `backend/ent/schema/testcaseversion.go` | 55 | 0 | unique FIXED |
| `backend/ent/schema/testresult.go` | 92 | 0 | чисто (append) |
| `backend/ent/schema/testrun.go` | 90 | 1 | RESIDUAL version |
| `backend/ent/schema/testsuite.go` | 63 | 1 | soft FIXED; RESIDUAL version |
| `backend/ent/schema/timelog.go` | 59 | 0 | чисто (append) |
| `backend/ent/schema/user.go` | 153 | 1 | RESIDUAL version |
| `backend/ent/schema/userblock.go` | 56 | 0 | чисто |
| `backend/ent/schema/webhookdelivery.go` | 58 | 0 | чисто (append log) |
| `backend/ent/schema/webhookendpoint.go` | 61 | 1 | soft+Sensitive FIXED; RESIDUAL version |
| `backend/ent/schema/workflowstatus.go` | 66 | 1 | RESIDUAL version/trash |
| `backend/ent/schema/workflowtransition.go` | 47 | 0 | чисто (junction) |

Файлов в списке: 15. Прочитано: 15. Пропущено: 0.

## Находки (открытые A2)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W8-5-001 | RESIDUAL (W8-A5-002) | P1 | `team.go:16-28` | Team без version/deleted_at/updated_at | Mutable user-visible против TZ | Три поля + soft-delete в project layer |
| A2W8-5-002 | RESIDUAL (W8-A5-005) | P1 | `testrun.go:19-59` | TestRun без `version` | Concurrent counter/status без CAS | version + VersionEQ |
| A2W8-5-003 | RESIDUAL (W8-A5-006∩version) | P1 | `testsuite.go:19-38` | Soft-delete есть, `version` нет | Rename/sort concurrent LWW | version + VersionEQ |
| A2W8-5-004 | RESIDUAL (W8-A5-008) | P1 | `user.go:17-77` | User без `version` | Lost update профиля/custom_fields | version + VersionEQ на UpdateUserProfile |
| A2W8-5-005 | RESIDUAL (W8-A5-009) | P1 | `workflowstatus.go:17-40` | Нет version/deleted_at; hard Delete | Concurrent rename/WIP; FK issues | version + soft-delete **или** запрет delete |
| A2W8-5-006 | RESIDUAL (W8-A5-010∩version) | P2 | `webhookendpoint.go:17-41` | Soft-delete есть, `version` нет | Concurrent update events/url без CAS | version + VersionEQ |
| A2W8-5-007 | RESIDUAL (W8-A5-001) | P2 | `systemsetting.go:15-28` | Mutable setting без `version` | Concurrent upsert LWW | version + VersionEQ |

## Матрица фокуса (UUID / version / soft-delete)

| Схема | UUID PK | version | deleted_at | Класс | Вердикт |
|-------|:-------:|:-------:|:----------:|-------|---------|
| SystemSetting | ✓ | ✗ | ✗ | mutable KV | A2W8-5-007 |
| Team | ✓ | ✗ | ✗ | mutable | A2W8-5-001 |
| TeamMember | ✓ | — | — | junction | чисто |
| TestCase | ✓ | ✓ | ✓ | эталон | FIXED |
| TestCaseVersion | ✓ | — | — | snapshot | unique FIXED |
| TestResult | ✓ | — | — | append | чисто |
| TestRun | ✓ | ✗ | ✗ | mutable | A2W8-5-002 |
| TestSuite | ✓ | ✗ | ✓ | mutable+trash | A2W8-5-003 |
| TimeLog | ✓ | — | — | append | чисто |
| User | ✓ | ✗ | — (status) | mutable | A2W8-5-004 |
| UserBlock | ✓ | — | — | junction | чисто |
| WebhookDelivery | ✓ | — | — | append | чисто |
| WebhookEndpoint | ✓ | ✗ | ✓ | mutable+trash | A2W8-5-006 |
| WorkflowStatus | ✓ | ✗ | ✗ | mutable | A2W8-5-005 |
| WorkflowTransition | ✓ | — | — | junction | чисто |

## Чисто

| Файл | Проверено |
|------|-----------|
| `teammember.go` / `workflowtransition.go` / `userblock.go` | junction Unique OK |
| `testresult.go` / `timelog.go` / `webhookdelivery.go` | append-only |
| `testcase.go` | UUID+version+deleted_at + composite (A8-002) |
| `testcaseversion.go` | unique version_number FIXED |
| soft TestSuite / WebhookEndpoint | TASK-039 FIXED |

## Сводка counts (открытые)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 5 |
| P2 | 2 |
| P3 | 0 |
| **Total open** | **7** |

FIXED-OK: 4 · NEW: 0 · REGRESS: 0
