# A2-W2 SUMMARY — tms + bases (Stage A2 re-audit)

> TASK-008 · Stage A2 · 2026-07-28 · **read-only**  
> Отчёты: [A2W2-1](A2W2-1-tms-handlers.md) · [A2W2-2](A2W2-2-tms-services.md) · [A2W2-3](A2W2-3-tms-repos.md) · [A2W2-4](A2W2-4-tms-import.md) · [A2W2-5](A2W2-5-bases.md)  
> Назначения: [W2-ASSIGNMENTS.md](W2-ASSIGNMENTS.md) · Stage A: [W2-SUMMARY.md](W2-SUMMARY.md)

## Покрытие

| Агент | Зона | В списке | Прочитано | Пропущено |
|-------|------|--------:|----------:|----------:|
| A1 | tms handlers | 24 | 24 | 0 |
| A2 | tms services | 22 | 22 | 0 |
| A3 | tms repos | 22 | 22 | 0 |
| A4 | tms import/parsers | 25 | 25 | 0 |
| A5 | bases | 16 | 16 | 0 |
| **Σ** | | **109** | **109** | **0** |

## Открытые находки по severity + tag

| Severity | NEW | REGRESS | RESIDUAL | Σ open |
|----------|----:|--------:|---------:|-------:|
| **P0** | 0 | 0 | 0 | **0** |
| **P1** | 0 | 0 | 0 | **0** |
| **P2** | 0 | 0 | 23 | **23** |
| **P3** | 2 | 0 | 3 | **5** |
| **Σ** | **2** | **0** | **26** | **28** |

> Дедуп: сырые строки в слотах = 28; классы Suite/Env IDOR/FK/import limits из Stage A **не** входят в open (FIXED-OK).

### По слотам (open)

| Slot | NEW | REGRESS | RESIDUAL | P0 | P1 | P2 | P3 |
|------|----:|--------:|---------:|---:|---:|---:|---:|
| A2W2-1 handlers | 1 | 0 | 4 | 0 | 0 | 4 | 1 |
| A2W2-2 services | 1 | 0 | 10 | 0 | 0 | 9 | 2 |
| A2W2-3 repos | 0 | 0 | 3 | 0 | 0 | 2 | 1 |
| A2W2-4 import | 0 | 0 | 7 | 0 | 0 | 6 | 1 |
| A2W2-5 bases | 0 | 0 | 2 | 0 | 0 | 2 | 0 |

## FIXED-OK (главные классы Stage A P1)

| Класс | Stage A IDs | Задача / evidence |
|-------|-------------|-------------------|
| Suite/Component IDOR | W2-A1-001…003, A2-001/002, A3-002/003 | TASK-015 · W18-C2 PASS |
| Environment IDOR + Compare | W2-A3-001, A2-003 | TASK-015 |
| Cross-project FK | W2-A2-004/005 | TASK-015 |
| Import LimitReader + steps + depth (importer) | W2-A4-001/002/003† | TASK-016 · W18-C2 PASS |
| Soft-delete suite/component/env | W2-A3-008 | P2-BE-08 |
| Bases CreateBase tx / merge / OL / has_more / ColType | W2-A5-001/002/004/005/007–012 | TASK-038+ |

† Depth/cap закрыт для **importer**; webhook/`ParseAllure`/`ParseJUnit` — **RESIDUAL** (A2W2-4-001).

## Топ открытых (handoff)

1. **N+1 отчёты TMS** (A2W2-2-001) + unbounded ListResults/Mongo (A2W2-3-001/002)  
2. **CreateBug ignores `:runId`** (A2W2-1-001) + err.Error на CreateBug (A2W2-2-006)  
3. **UpdateTestCase snapshot/rollback без tx / 409→500** (A2W2-2-002/003)  
4. **Allure matcher swallows DB err** на automation_id/case_key (A2W2-4-003)  
5. **Bases SoftDelete без cascade** (A2W2-5-001)

## REGRESS

**0** — P1 IDOR/FK/import limits и `authorizeBase` не откатились.

## NEW (2×P3)

| ID | Суть |
|----|------|
| A2W2-1-005 | Webhook: multipart 5 MB cap, raw body — только Fiber 25 MB |
| A2W2-2-011 | `ImportAllureResults` 400 с `err.Error()` парсера |

## Вердикт

Stage A W2 P1 (**14 сырых / 5 классов**) закрыты и подтверждены кодом + W18 mutation. Остаток — **P2/P3 RESIDUAL** (техдолг lists/tx/contract) и **2 мелких NEW**. Продуктовый код на A2 **не менялся**.
