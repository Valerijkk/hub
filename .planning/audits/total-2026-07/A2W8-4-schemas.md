# A2 W8 агент 4 — ent/schema part 4 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659f` · Baseline: [A8-4-schemas.md](A8-4-schemas.md)  
> Периметр: 17 схем · Прочитано: 17 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с Stage A (W8-A4-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W8-A4-001 | RESIDUAL | открыт | Project без `version` |
| W8-A4-002 | RESIDUAL | открыт | Project 25+ edges без OnDelete / purge policy |
| W8-A4-003 | RESIDUAL | открыт | Org/Poll/Role/SavedFilter/ScheduledCall/Sprint без version |
| W8-A4-004 | RESIDUAL | открыт | Sprint без deleted_at |
| W8-A4-005 | FIXED-OK | закрыт | `role.go:54` unique `(project, name)` |
| W8-A4-006 | FIXED-OK | закрыт | `organization.go:32-35` invite_token Sensitive |
| W8-A4-007 | FIXED-OK | закрыт | Push `p256dh_key`/`auth_key` Sensitive |
| W8-A4-008 | RESIDUAL | открыт | Recording без deleted_at |
| W8-A4-009 | RESIDUAL | открыт | Poll/Option/Vote без soft-delete / OnDelete |
| W8-A4-010 | RESIDUAL | открыт | Org→users/projects без OnDelete |
| W8-A4-011 | FIXED-OK | закрыт | Mojibake «Рёбра» (TASK-068) |
| W8-A4-012 | RESIDUAL | открыт | ScheduledCall soft-delete без version |

**REGRESS:** 0 · **RESIDUAL:** 7 · **NEW:** 1 · **FIXED-OK:** 4

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/organization.go` | 59 | 2 | RESIDUAL version/OnDelete + NEW dup index |
| `backend/ent/schema/outboxevent.go` | 52 | 0 | чисто (append worker) |
| `backend/ent/schema/pinnedmessage.go` | 52 | 0 | чисто |
| `backend/ent/schema/poll.go` | 73 | 1 | RESIDUAL version/trash |
| `backend/ent/schema/polloption.go` | 52 | 1 | RESIDUAL OnDelete/trash |
| `backend/ent/schema/pollvote.go` | 63 | 1 | RESIDUAL OnDelete |
| `backend/ent/schema/project.go` | 97 | 2 | RESIDUAL version + OnDelete |
| `backend/ent/schema/projectmember.go` | 55 | 0 | чисто (junction) |
| `backend/ent/schema/projecttemplate.go` | 60 | 0 | чисто (mojibake FIXED) |
| `backend/ent/schema/pushsubscription.go` | 65 | 0 | FIXED-OK Sensitive |
| `backend/ent/schema/reaction.go` | 53 | 0 | чисто |
| `backend/ent/schema/recording.go` | 88 | 1 | RESIDUAL trash |
| `backend/ent/schema/role.go` | 57 | 1 | FIXED unique; RESIDUAL version |
| `backend/ent/schema/savedfilter.go` | 68 | 1 | RESIDUAL version |
| `backend/ent/schema/scheduledcall.go` | 97 | 1 | RESIDUAL version |
| `backend/ent/schema/session.go` | 53 | 0 | чисто (refresh Sensitive) |
| `backend/ent/schema/sprint.go` | 66 | 1 | RESIDUAL version + trash |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (открытые A2)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W8-4-001 | RESIDUAL (W8-A4-001) | P1 | `project.go:17-52` | Mutable Project без `version` | Concurrent Update/UpdateStatus LWW (A8-003) | `version` + VersionEQ |
| A2W8-4-002 | RESIDUAL (W8-A4-002/010) | P1 | `project.go:55-87`; `organization.go:45-49` | Нет OnDelete / hard-purge policy на children | PermanentDelete FK trap или silent orphan | ADR staged purge **или** explicit OnDelete |
| A2W8-4-003 | RESIDUAL (W8-A4-003/012) | P1 | org/poll/role/savedfilter/scheduledcall/sprint | Mutable без `version` | P1-BE-U06 | VersionMixin |
| A2W8-4-004 | RESIDUAL (W8-A4-004) | P2 | `sprint.go:17-43` | Нет deleted_at | Issue soft, Sprint — нет trash | deleted_at **или** ADR lifecycle |
| A2W8-4-005 | RESIDUAL (W8-A4-008/009) | P2 | `recording.go`; `poll*.go` | User media/content без soft-delete / OnDelete | Нет корзины; FK trap при delete poll | Soft-delete Poll; recording deleted_at; OnDelete |
| A2W8-4-006 | RESIDUAL (W8-A4-003∩Role) | P2 | `role.go:16-36` | Unique name FIXED; version нет | Concurrent UpdateRole LWW | version |
| A2W8-4-007 | NEW | P3 | `organization.go:53-57` | `slug`/`invite_token` уже `Unique()` на field + повтор в `Indexes()` | Дублирующий unique index (класс W8-A1-009) | Оставить один источник Unique |
| A2W8-4-008 | RESIDUAL (P2-BE-U02/U03) | P2 | A4 mutable set | Системный gap version/OnDelete | ADR | ADR |

## Чисто

| Файл | Проверено |
|------|-----------|
| `outboxevent.go` | UUID; worker index (status, created_at); append-ish |
| `pinnedmessage.go` | UUID; unique (channel, message_id Mongo) |
| `projectmember.go` | UUID; unique (user, project) |
| `reaction.go` | UUID; unique (comment, user, emoji) |
| `session.go` | UUID; refresh_token Sensitive; ephemeral |
| `pushsubscription.go` | Sensitive keys FIXED |
| `projecttemplate.go` | UUID; godoc OK |

UUID PK: **17/17**. Soft-delete в A4: project + scheduledcall (sprint/poll/recording/org — нет).

## Сводка counts (открытые)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 4 |
| P3 | 1 |
| **Total open** | **8** |

FIXED-OK: 4 · NEW: 1 · REGRESS: 0
