# A3-W8 SUMMARY — ent/schema (FINAL)

> TASK-008 · Stage A3 FINAL · 2026-07-29 · READ-ONLY  
> HEAD: `ee353287` · Assignments: [W8-ASSIGNMENTS.md](W8-ASSIGNMENTS.md)  
> Отчёты: [A3W8-1](A3W8-1-schemas.md) · [A3W8-2](A3W8-2-schemas.md) · [A3W8-3](A3W8-3-schemas.md) · [A3W8-4](A3W8-4-schemas.md) · [A3W8-5](A3W8-5-schemas.md)

## Покрытие

| Агент | В списке | Σ open |
|-------|--------:|-------:|
| A1 | 17 | 2 |
| A2 | 17 | 6 |
| A3 | 17 | 2 |
| A4 | 17 | 2 |
| A5 | 15 | 7 |
| **Σ** | **83** | **19** |

## Severity

| P0 | P1 | P2 | P3 |
|---:|---:|---:|---:|
| 0 | **1** | **13** | **5** |

## High-signal

| ID | Sev | Суть |
|----|-----|------|
| A3W8-5-* | P1/P2 | Team soft-delete · WorkflowStatus hard-delete · OnDelete Cascades |
| A3W8-2-002…004 | P2 | OnDelete CoffeeRound/HubBase/Integration |
| A3W8-3-001…002 | P2 | HubTable/JournalEntry Cascade |
| A3W8-4-001 | P2 | partial unique active sprint |

VersionMixin TASK-095 — большинство FIXED-OK. REGRESS 0.
