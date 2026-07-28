# A2 W6 агент A5 — audit remaining tests

> TASK-008 · Stage A2 · read-only · 2026-07-28  
> Периметр: `W6-ASSIGNMENTS.md` §A5 (12 файлов)

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/testframework/audit/repository_extra_test.go | 197 | 2 | RESIDUAL |
| backend/testframework/audit/repository_newspost_trash_test.go | 113 | 0 | чисто |
| backend/testframework/audit/repository_scheduled_call_trash_test.go | 117 | 0 | чисто |
| backend/testframework/audit/repository_test.go | 115 | 0 | чисто |
| backend/testframework/audit/repository_trash_test.go | 203 | 1 | RESIDUAL |
| backend/testframework/audit/retention_test.go | 175 | 0 | чисто |
| backend/testframework/audit/service_deactivate_test.go | 117 | 1 | RESIDUAL |
| backend/testframework/audit/service_limits_test.go | 281 | 1 | RESIDUAL |
| backend/testframework/audit/service_purge_test.go | 183 | 2 | RESIDUAL (+ NEW cement) |
| backend/testframework/audit/service_security_test.go | 484 | 0 | FIXED-OK (SkipsOnCreateError) |
| backend/testframework/audit/service_test.go | 370 | 0 | чисто |
| backend/testframework/audit/service_trash_test.go | 69 | 0 | чисто |

Файлов: **12/12/0**.

## Перепроверка Stage A (W6-A5)

| Старый ID | Tag | Статус | Комментарий |
|-----------|-----|--------|-------------|
| W6-A5-001 | RESIDUAL | жив | `PurgeExpired_Succeeds` — только NoError; нет assert вызовов/cutoff |
| W6-A5-002 | RESIDUAL | жив | Retention-hide 30d только newspost/scheduled_call |
| W6-A5-003 | RESIDUAL | жив | SoftDelete/Restore AllTypes без call-recording |
| W6-A5-004 | RESIDUAL | жив | Purge test_case/file/project — только `n==1`, нет IsNotFound |
| W6-A5-005 | RESIDUAL | жив | ListTrash_WithProjectFilter — только StatusOK |
| W6-A5-006 | RESIDUAL | жив | SoftDelete+Restore issue/file без DeletedAt nil assert |
| W6-A5-007 | FIXED-OK | закрыт | `ListTrash_WithItems` — cutoff ≈ now−30d |
| W6-A5-008 | RESIDUAL | жив | ReassignTasks mock → Equal(0) без call counter |
| W6-A5-009 | FIXED-OK | закрыт | `SkipsOnCreateError` inject `createErr` |
| W6-A5-010 | RESIDUAL | жив | GDPR Success на mock без cascade evidence |

REGRESS: **0**. assert-free: **не найдено**.

## Находки (открытые)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|----------|---------------|------------|
| A2W6-A5-001 | RESIDUAL | P2 | `service_purge_test.go:79-94` | PurgeExpired_Succeeds hollow | Prod всегда nil — тест не ловит пропуск purge | Captured cutoffs + call counts ×6 |
| A2W6-A5-002 | NEW | P2 | `service_purge_test.go:96-103` | `PurgeExpired_LogsErrorsWithoutFailing` **требует** NoError при db fail | Цементирует W6-A4-004 fail-open после Stage B | Удалить/перевернуть: err != nil при purge fail |
| A2W6-A5-003 | RESIDUAL | P2 | repo trash/extra | Нет RetentionHides для issue/test_case/file/project | Guard `DeletedAtGT` в repo не mutation-proof | Зеркало NewsPost −40d |
| A2W6-A5-004 | RESIDUAL | P2 | `service_purge_test.go:145-167` | AllTypes SoftDelete/Restore без map[type]called | Switch collapse → зелёный | Call-recording mock |
| A2W6-A5-005 | RESIDUAL | P2 | `repository_extra_test.go` purge TC/files/projects | Нет IsNotFound после hard delete | `n==1` при soft-only возможен | Get → IsNotFound |
| A2W6-A5-006 | RESIDUAL | P2 | `service_deactivate_test.go:48-63` | project_id filter HTTP без captured projectID | Снятие filter → 200 green | Assert mock got projectID |
| A2W6-A5-007 | RESIDUAL | P3 | trash restore paths | Нет DeletedAt==nil после restore (issue/file) | Ложный PASS на list empty | Get + Nil(DeletedAt) |
| A2W6-A5-008 | RESIDUAL | P3 | `service_limits_test.go` Reassign | mock returns 0 без proof call | Service no-op → green | Counters assigned/reported |
| A2W6-A5-009 | RESIDUAL | P3 | `repository_extra_test.go` GDPR Success | mock wiring only | Не ловит пропуск cascade | Оставить mock; cascade — p1be34 suite |

## Counts

| | n |
|--|--:|
| NEW | 1 |
| REGRESS | 0 |
| RESIDUAL | 8 |
| FIXED-OK | 2 |
| P0 / P1 / P2 / P3 (open) | 0 / 0 / 6 / 3 |
| **Σ open** | **9** |
