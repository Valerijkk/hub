# W2 SUMMARY — tms + bases

> TASK-008 · Стадия A · 2026-07-28 · read-only  
> Отчёты: [A2-1](A2-1-tms-handlers.md) · [A2-2](A2-2-tms-services.md) · [A2-3](A2-3-tms-repos.md) · [A2-4](A2-4-tms-import.md) · [A2-5](A2-5-bases.md)  
> Назначения: [W2-ASSIGNMENTS.md](W2-ASSIGNMENTS.md)

## Покрытие

| Агент | Зона | В списке | Прочитано | Пропущено |
|-------|------|--------:|----------:|----------:|
| A1 | tms handlers | 24 | 24 | 0 |
| A2 | tms services | 22 | 22 | 0 |
| A3 | tms repos | 22 | 22 | 0 |
| A4 | tms import/parsers | 25 | 25 | 0 |
| A5 | bases | 16 | 16 | 0 |
| **Σ** | | **109** | **109** | **0** |

## Находки

| Severity | Сырые | Комментарий |
|----------|------:|-------------|
| **P0** | 0 | — |
| **P1** | 14 | suite/component/env IDOR; import без LimitReader; CompareEnvironments |
| **P2** | 36 | cross-ref без scope, N+1, test gaps |
| **P3** | 16 | косметика / assert-weak |
| **Σ** | **66** | дубли suite/component между A1–A3 |

### Топ P1

| Класс | IDs | Суть |
|-------|-----|------|
| Suite/Component IDOR | W2-A1-001…003, A2-001…002, A3-002…003 | Update/Delete без `projectID` |
| Environment IDOR | W2-A3-001, A2-003 | Update/Delete/Compare env без project scope |
| Import memory DoS | W2-A4-001 (+ A4-003) | FormFile → parsers без `io.LimitReader` |
| Cross-ref без scope | W2-A2-004 | ParentID/SuiteID/Labels/Components не проверяются на project |

### Bases

Owner-scope `authorizeBase()` держится: **P0/P1 нет**. Остаток P2/P3 (caps, тесты, мелочь).

### Закрыто с прошлого аудита

A3-003 FindByKey scope · A3-004 VersionEQ · A3-007 rollback · A3-008 result scope · A3-009 nil TC · A3-010 matcher · A3-012 import run — в основном **закрыты** (подтверждено A2–A4); новые дыры — suites/components/env и import body limit.

## Evidence тестов

| Пакет | Результат |
|-------|-----------|
| `testframework/tms` | **ok** 7.029s |
| `testframework/bases` | **ok** 2.928s |

## Handoff

→ **W3: messenger + ws + calls**. Код не трогать до конца стадии A.
