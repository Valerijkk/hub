# A2-W1 SUMMARY — tracker + sprint + project (re-audit)

> TASK-008 · Stage A2 · 2026-07-28 · READ-ONLY  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc` (`0d6a659f`)  
> Baseline Stage A: [W1-SUMMARY.md](W1-SUMMARY.md) · Assignments: [W1-ASSIGNMENTS.md](W1-ASSIGNMENTS.md)  
> Отчёты: [A2W1-1](A2W1-1-tracker-handlers.md) · [A2W1-2](A2W1-2-tracker-services.md) · [A2W1-3](A2W1-3-tracker-repos.md) · [A2W1-4](A2W1-4-sprint.md) · [A2W1-5](A2W1-5-project.md)

## Покрытие

| Агент | Зона | В списке | Прочитано | Пропущено | Отчёт |
|-------|------|--------:|----------:|----------:|-------|
| A1 | tracker handlers | 36 | 36 | 0 | A2W1-1 |
| A2 | tracker services | 33 | 33 | 0 | A2W1-2 |
| A3 | tracker repos+automation | 37 | 37 | 0 | A2W1-3 |
| A4 | sprint | 19 | 19 | 0 | A2W1-4 |
| A5 | project | 38 | 38 | 0 | A2W1-5 |
| **Σ** | | **163** | **163** | **0** | |

## Исход A1 → A2 (теги)

| Tag | Count (уникальные A1 IDs / классы) | Комментарий |
|-----|-----------------------------------:|-------------|
| FIXED-OK | **62 / ~66 сырых** | P0/P1 W1 закрыты (TASK-009/013/014…); большинство P2/P3 W21 |
| RESIDUAL | **4** | automation Ack; Sprint version; cleanup links; JQL GetDB |
| REGRESS | **0** | регрессий по A1 P0–P1 не найдено |
| NEW | **4** | assignee∈project; AddTeamToProject tx; unknown subject Ack; last team member |

## Активные находки A2 (сырые)

| Severity | Count | IDs |
|----------|------:|-----|
| **P0** | **0** | — |
| **P1** | **0** | — |
| **P2** | **5** | A2W1-A2-002 · A2W1-A3-001 · A2W1-A4-001 · A2W1-A5-001 · A2W1-A5-002 |
| **P3** | **3** | A2W1-A2-001 · A2W1-A3-002 · A2W1-A5-003 |
| **Σ** | **8** | |

### Дедуп / приоритет

| Класс | Tag | Sev | Суть |
|-------|-----|-----|------|
| Assignee без membership | NEW | P2 | Create/UpdateIssue `AssigneeID` не ∈ project |
| Automation Ack on engine fail | RESIDUAL | P2 | W1-A3-007; Handle* void |
| Sprint без `version` | RESIDUAL | P2 | W1-A4-008; TZ optimistic lock; нужен approval миграции |
| DeleteProject cleanup best-effort | RESIDUAL | P2 | W1-A5-008 |
| AddTeamToProject без tx | NEW | P2 | AddTeamIDs до AddMember |
| JQL via GetDB | RESIDUAL | P3 | остаток W1-A1-008 в service |
| Unknown NATS subject → Ack | NEW | P3 | automation_consumer |
| RemoveTeamMember last member | NEW | P3 | orphan team |

## Сравнение с Stage A W1

| Метрика | Stage A (W1-SUMMARY) | Stage A2 |
|---------|---------------------:|---------:|
| Файлов | 163 | 163 |
| Сырые находки | 66 | **8** активных |
| P0 | 1 | **0** |
| P1 | 19 | **0** |
| P2 | 28 | **5** |
| P3 | 18 | **3** |

P0 project RBAC (W1-A5-001) и все P1 IDOR/tx классы W1 — **FIXED-OK** на HEAD.

## Handoff

- A2-W1 **COMPLETE** (отчёты написаны; код не менялся).
- Следующий: A2-W2…W11 по `A2-PLAN.md`.
- Кандидаты в backlog после полного A2: P2×5 + P3×3 выше (не заводить задачи до сводки всех волн A2, если оркестратор не решит иначе).
