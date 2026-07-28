# A3-W7 SUMMARY — remaining backend (FINAL)

> TASK-008 · Stage A3 FINAL · 2026-07-29 · READ-ONLY  
> HEAD: `ee353287` · Assignments: [W7-ASSIGNMENTS.md](W7-ASSIGNMENTS.md)  
> Отчёты: [A3W7-1](A3W7-1-notification-mail.md) · [A3W7-2](A3W7-2-v4-business.md) · [A3W7-3](A3W7-3-integrations.md) · [A3W7-4](A3W7-4-storage-dash-backup.md) · [A3W7-5](A3W7-5-mp-status-dev.md)

## Покрытие

| Агент | Зона | В списке | Σ open |
|-------|------|--------:|-------:|
| A1 | notification+mail | 66 | 2 |
| A2 | v4 business | 72 | 3 |
| A3 | integrations | 71 | 3 |
| A4 | storage+dash+backup | 48 | 5 |
| A5 | mp+status+dev | 27 | 4 |
| **Σ** | | **284** | **17** |

## Severity

| P0 | P1 | P2 | P3 |
|---:|---:|---:|---:|
| 0 | 0 | **7** | **10** |

## High-signal

| ID | Sev | Суть |
|----|-----|------|
| A3W7-2-001 | P2 REGRESS | Coffee notify `_=` вместо slog.Warn |
| A3W7-2-002 | P2 | UpdateNewsPoll без VersionEQ |
| A3W7-3-001…003 | P2 | trash filter sprint PDF / FindIssue / link-preview membership |
| A3W7-5-001 | P2 | review_note на public catalog |

A2 closes (NF-07, soft-delete GH/GL, TASK-121…) — mostly FIXED-OK. REGRESS: coffee notify.
