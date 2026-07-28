# W1 SUMMARY — tracker + sprint + project

> TASK-008 · Стадия A · 2026-07-28 · read-only  
> Агенты: [A1](A1-1-tracker-handlers.md) · [A2](A1-2-tracker-services.md) · [A3](A1-3-tracker-repos.md) · [A4](A1-4-sprint.md) · [A5](A1-5-project.md)  
> Назначения: [W1-ASSIGNMENTS.md](W1-ASSIGNMENTS.md)

## Покрытие файлов

| Агент | Зона | В списке | Прочитано | Пропущено | Отчёт |
|-------|------|--------:|----------:|----------:|-------|
| A1 | tracker handlers | 36 | 36 | 0 | A1-1 |
| A2 | tracker services | 33 | 33 | 0 | A1-2 |
| A3 | tracker repos+automation | 37 | 37 | 0 | A1-3 |
| A4 | sprint | 19 | 19 | 0 | A1-4 |
| A5 | project | 38 | 38 | 0 | A1-5 |
| **Σ** | | **163** | **163** | **0** | |

Каждый файл — строка в таблице «Прочитанные файлы» соответствующего отчёта.

## Находки (сырые + дедуп классов)

| Severity | Сырые (5 отчётов) | Уникальные классы (оценка) |
|----------|------------------:|---------------------------:|
| **P0** | 1 | 1 |
| **P1** | 19 | ~8 (workflow/custom-fields IDOR повторяется handler+service+repo) |
| **P2** | 28 | ~22 |
| **P3** | 18 | ~16 |
| **Σ** | **66** | **~47** |

### P0 (подтверждено оркестратором)

| ID | Файл:строка | Суть |
|----|-------------|------|
| W1-A5-001 | `project/handler.go:23-57` + `wiring_core.go:28` | `RegisterRoutes` только `authMw`, без RBAC. Service `Update/Delete/Archive/AddMember` без membership. Любой auth-юзер с UUID проекта может soft-delete / менять members / invites |

### Топ P1 (уникальные классы)

| Класс | IDs | Суть |
|-------|-----|------|
| Workflow CRUD IDOR | W1-A1-002…005, A2-002…004, A3-002…003 | status/transition mutate без `projectID` |
| Custom field delete IDOR | W1-A1-001, A2-001, A3-001 | `DeleteCustomField` только по `fieldID` |
| Project invite revoke IDOR | W1-A5-002 | `RevokeInvite` без scope project |
| Project teams/templates ACL | W1-A5-004, A5-005 | teams/templates только auth |
| Last-owner demote | W1-A5-003 | `UpdateMemberRole` без защиты последнего owner |
| CompleteSprint не в tx | W1-A4-001 | move issues + status не атомарны |

### Закрыто с прошлого аудита (не дублировать)

A3-001 sprint scope · A3-002 issue key scope · A3-005 timer · A3-011 comments · A3-015/016 WIP/Start fail-closed — **исправлены** (подтверждено A2/A3/A4). Остатки: A3-013 MoveIssue version, A3-019 poker CAS/TTL, A3-021 outbox `_=`, A3-022 JQL layering, A3-023 AddTeamToProject.

## Evidence тестов (с БД)

```
POSTGRES_TEST_URL=postgresql://hub:hub@localhost:5432/hub?sslmode=disable
MONGODB_TEST_URL=mongodb://hub:hub@localhost:27017/?authSource=admin
go test -tags hubtest -race -count=1 ./testframework/{sprint,project,tracker}/...
```

| Пакет | Результат | Лог |
|-------|-----------|-----|
| sprint | **ok** 2.929s | W1-test-sprint.log |
| project | **ok** 3.291s | W1-test-project.log |
| tracker | **ok** 7.978s | W1-test-tracker.log |

Docker: `hub-postgres-1`, `hub-mongo-1`, `deployments-redis-1`, `deployments-nats-1` — Up.

## Handoff

- Стадия A продолжается → **W2: tms + bases** (код не трогать).
- P0 W1-A5-001 — первый кандидат на задачу стадии B (после W11).
- Дубликаты workflow/custom-fields между A1/A2/A3 схлопнуть в одну задачу на цепочку handler→service→repo.
