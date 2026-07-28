# A2-W8 SUMMARY — ent/schema (re-audit)

> TASK-008 · Stage A2 · 2026-07-28 · READ-ONLY  
> HEAD: `8d9c09502a94e59620858e2057e970cc62f2cd57` (`8d9c0950`)  
> Baseline Stage A: [W8-SUMMARY.md](W8-SUMMARY.md) · Assignments: [W8-ASSIGNMENTS.md](W8-ASSIGNMENTS.md)  
> Отчёты: [A2W8-1](A2W8-1-schemas.md) · [A2W8-2](A2W8-2-schemas.md) · [A2W8-3](A2W8-3-schemas.md) · [A2W8-4](A2W8-4-schemas.md) · [A2W8-5](A2W8-5-schemas.md)

## Покрытие

| Агент | Зона | В списке | Прочитано | Пропущено | Отчёт |
|-------|------|--------:|----------:|----------:|-------|
| A1 | schemas part 1 | 17 | 17 | 0 | A2W8-1 |
| A2 | schemas part 2 | 17 | 17 | 0 | A2W8-2 |
| A3 | schemas part 3 | 17 | 17 | 0 | A2W8-3 |
| A4 | schemas part 4 | 17 | 17 | 0 | A2W8-4 |
| A5 | schemas part 5 | 15 | 15 | 0 | A2W8-5 |
| **Σ** | | **83** | **83** | **0** | |

## Исход A1 → A2 (теги)

| Tag | Count (A1 IDs / classes) | Комментарий |
|-----|-------------------------:|-------------|
| FIXED-OK | **25** | Soft-delete TMS/webhook/apikey (TASK-039); unique/CHECK/O2O (TASK-025/046); Sensitive secrets; MaxLen/mentions/mojibake (TASK-068) |
| RESIDUAL | **~33 A1 → 35 open A2 rows** | VersionMixin gap; OnDelete=0; indexes/trash на части схем |
| REGRESS | **0** | регрессий по закрытым A1 не найдено |
| NEW | **0** | новых классов сверх A1 не выявлено |

## Активные находки A2 (сырые open)

| Severity | Count | По агентам |
|----------|------:|------------|
| **P0** | **0** | — |
| **P1** | **14** | A1:2 · A2:3 · A3:1 · A4:3 · A5:5 |
| **P2** | **21** | A1:4 · A2:6 · A3:6 · A4:3 · A5:2 |
| **P3** | **0** | — |
| **Σ** | **35** | |

| Агент | NEW | REGRESS | RESIDUAL open | FIXED-OK | Σ open |
|-------|----:|--------:|--------------:|---------:|-------:|
| A1 | 0 | 0 | 6 | 4 | 6 |
| A2 | 0 | 0 | 9 | 6 | 9 |
| A3 | 0 | 0 | 7 | 7 | 7 |
| A4 | 0 | 0 | 6 | 4 | 6 |
| A5 | 0 | 0 | 7 | 4 | 7 |
| **Σ** | **0** | **0** | **35** | **25** | **35** |

### Дедуп / приоритет (классы)

| Класс | Tag | Sev | Суть |
|-------|-----|-----|------|
| VersionMixin gap | RESIDUAL | P1 | Mutable схемы без `version` (Comment/Project/User/Sprint/Channel/…); нужен ADR P1-BE-U06 |
| OnDelete / cascade | RESIDUAL | P1 | **0** `entsql.OnDelete` во всём `ent/schema/` (Project/Org/Channel/Call/Poll) |
| Soft-delete gaps | RESIDUAL | P1–P2 | LeaveRequest/Sprint/Recording/Poll/Label/Kudo/Team/WorkflowStatus |
| Composite/unique indexes | RESIDUAL | P2 | CRM email/owner; Hub* table+deleted_at; IssueLink link_type; Git*Link external |
| FileMetadata / Notification IDs | RESIDUAL | P2 | `entity_id` String; polymorphic без FK |
| GitLab webhook Optional | RESIDUAL | P2 | Sensitive FIXED; пустой секрет схемой всё ещё разрешён |

## Сравнение с Stage A W8

| Метрика | Stage A (W8-SUMMARY) | Stage A2 |
|---------|---------------------:|---------:|
| Файлов | 83 | 83 |
| Сырые находки | **71** | **35** активных |
| P0 | 0 | **0** |
| P1 | 19 | **14** |
| P2 | 40 | **21** |
| P3 | 12 | **0** |

P3 hygiene (mojibake/MaxLen/mentions) и часть P1/P2 constraints (coffee unique, notif partial unique, journal O2O, XOR CHECK, TMS/webhook soft-delete, Sensitive) — **FIXED-OK**. Системный VersionMixin + OnDelete остаются доминирующими residual.

## Handoff

- A2-W8 **COMPLETE** (5 отчётов + сводка; код не менялся).
- Кандидаты backlog после полного A2: VersionMixin ADR · OnDelete ADR · оставшиеся soft-delete/indexes (не заводить до сводки всех волн A2, если оркестратор не решит иначе).
