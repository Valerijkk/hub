# A3-W1 SUMMARY — tracker + sprint + project (FINAL)

> TASK-008 · Stage A3 FINAL · 2026-07-29 · READ-ONLY  
> HEAD: `ee353287` (`cursor-grok-4.5`)  
> Plan: [A3-PLAN.md](A3-PLAN.md) · Assignments: [W1-ASSIGNMENTS.md](W1-ASSIGNMENTS.md)  
> Отчёты: [A3W1-1](A3W1-1-tracker-handlers.md) · [A3W1-2](A3W1-2-tracker-services.md) · [A3W1-3](A3W1-3-tracker-repos.md) · [A3W1-4](A3W1-4-sprint.md) · [A3W1-5](A3W1-5-project.md)

## Покрытие

| Агент | Зона | В списке | Прочитано | Пропущено | Отчёт |
|-------|------|--------:|----------:|----------:|-------|
| A1 | tracker handlers | 36 | 36 | 0 | A3W1-1 |
| A2 | tracker services | 33 | 33 | 0 | A3W1-2 |
| A3 | tracker repos+automation | 37 | 37 | 0 | A3W1-3 |
| A4 | sprint | 19 | 19 | 0 | A3W1-4 |
| A5 | project | 38 | 38 | 0 | A3W1-5 |
| **Σ** | | **163** | **163** | **0** | |

## Активные находки A3-W1 (сырые)

| Severity | Count |
|----------|------:|
| P0 | **0** |
| P1 | **1** |
| P2 | **6** |
| P3 | **12** |
| **Σ open** | **19** |

| Слот | P0 | P1 | P2 | P3 | Σ |
|------|---:|---:|---:|---:|--:|
| A1 handlers | 0 | 0 | 0 | 0 | 0 |
| A2 services | 0 | 1 | 2 | 3 | 6 |
| A3 repos | 0 | 0 | 1 | 2 | 3 |
| A4 sprint | 0 | 0 | 1 | 2 | 3 |
| A5 project | 0 | 0 | 2 | 5 | 7 |

## High-signal (до дедупа)

| ID | Sev | Tag | Суть |
|----|-----|-----|------|
| A3W1-2-001 | P1 | NEW | `UpdateIssue` labels без `CountLabelsInProject` (Create уже scoped) |
| A3W1-A5-001 | P2 | NEW | Archive/Unarchive воскрешает soft-deleted project (`ClearDeletedAt`) |
| A3W1-A5-002 | P2 | NEW | `PermanentDelete` без trash + без system admin |
| A3W1-A3-001 | P2 | REGRESS | Automation `FilterSubject` → Nak storm на updated/comment/link |
| A3W1-A4-001 | P2 | NEW | CompleteSprint тащит soft-deleted issues |
| A3W1-2-002/003 | P2 | NEW | CreateLabel groupID / automation channel_id без project scope |

## A2 → A3

| Tag | Комментарий |
|-----|-------------|
| FIXED-OK | A2 assignee/JQL, sprint VersionMixin, project cleanup tx, AddTeam tx — закрыты remediation |
| REGRESS | 1 (automation Nak) |
| RESIDUAL | last team member; automation action err swallow; sprint retro OL |
| NEW | см. high-signal |

## Следующий шаг

A3-W2 (tms+bases, 109 файлов) → 5 слотов по `W2-ASSIGNMENTS.md`.
