# A2 W8 агент 4 — ent/schema part 4 (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `8d9c0950` · Baseline: [A8-4-schemas.md](A8-4-schemas.md)  
> Периметр: 17 схем · Прочитано: 17 · Пропущено: 0  
> Теги: `NEW` | `REGRESS` | `RESIDUAL` | `FIXED-OK`

## Сверка с Stage A (W8-A4-*)

| A1 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| W8-A4-001 | RESIDUAL | открыт | `project.go` — нет `field.Int("version")` при updated_at/soft-delete |
| W8-A4-002 | RESIDUAL | открыт | Project 25+ `edge.To` без `entsql.OnDelete` (пакет: **0** OnDelete) |
| W8-A4-003 | RESIDUAL | открыт | Org/Poll/Role/SavedFilter/ScheduledCall/Sprint без `version` |
| W8-A4-004 | RESIDUAL | открыт | `sprint.go` — нет `deleted_at` |
| W8-A4-005 | FIXED-OK | закрыт | `role.go:52-55` unique `(project, name)` (P2-BE-22) |
| W8-A4-006 | FIXED-OK | закрыт | `organization.go:35` `invite_token` Sensitive (TASK-039) |
| W8-A4-007 | FIXED-OK | закрыт | `pushsubscription.go:30-34` p256dh/auth Sensitive |
| W8-A4-008 | RESIDUAL | открыт | `recording.go` — нет `deleted_at` |
| W8-A4-009 | RESIDUAL | открыт | Poll/Option/Vote без soft-delete / OnDelete |
| W8-A4-010 | RESIDUAL | открыт | Org `users`/`projects` без OnDelete policy |
| W8-A4-011 | FIXED-OK | закрыт | project/projecttemplate godoc «Рёбра» (TASK-068) |
| W8-A4-012 | RESIDUAL | открыт | ScheduledCall soft-delete есть, version нет (= W8-A4-003) |

**REGRESS:** 0 · **RESIDUAL (open):** 7 · **NEW:** 0 · **FIXED-OK:** 4

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/organization.go` | 58 | 1 | Sensitive FIXED; RESIDUAL version/OnDelete |
| `backend/ent/schema/outboxevent.go` | 51 | 0 | чисто (append-only) |
| `backend/ent/schema/pinnedmessage.go` | 51 | 0 | чисто |
| `backend/ent/schema/poll.go` | 72 | 1 | RESIDUAL version/trash |
| `backend/ent/schema/polloption.go` | 51 | 1 | RESIDUAL cascade (с poll) |
| `backend/ent/schema/pollvote.go` | 62 | 1 | RESIDUAL cascade (с poll) |
| `backend/ent/schema/project.go` | 96 | 1 | RESIDUAL version + OnDelete |
| `backend/ent/schema/projectmember.go` | 54 | 0 | чисто (junction) |
| `backend/ent/schema/projecttemplate.go` | 59 | 0 | чисто (mojibake FIXED) |
| `backend/ent/schema/pushsubscription.go` | 64 | 0 | Sensitive FIXED |
| `backend/ent/schema/reaction.go` | 52 | 0 | чисто |
| `backend/ent/schema/recording.go` | 87 | 1 | RESIDUAL deleted_at |
| `backend/ent/schema/role.go` | 56 | 1 | unique FIXED; RESIDUAL version |
| `backend/ent/schema/savedfilter.go` | 67 | 1 | RESIDUAL version |
| `backend/ent/schema/scheduledcall.go` | 96 | 1 | soft OK; RESIDUAL version |
| `backend/ent/schema/session.go` | 52 | 0 | чисто (refresh Sensitive) |
| `backend/ent/schema/sprint.go` | 65 | 1 | RESIDUAL version + deleted_at |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (открытые A2)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A2W8-4-001 | RESIDUAL (W8-A4-001) | P1 | `project.go:17-52` | Mutable Project без `version` | TZ OL; concurrent Update/UpdateStatus LWW | `version` + VersionEQ |
| A2W8-4-002 | RESIDUAL (W8-A4-002/010) | P1 | `project.go:55-87`; `organization.go:46-50` | Нет `entsql.OnDelete` на parent→child | PermanentDelete FK trap / silent data loss | ADR Cascade vs app purge + OnDelete |
| A2W8-4-003 | RESIDUAL (W8-A4-003/012) | P1 | org/poll/role/savedfilter/scheduledcall/sprint | Mutable без `version` | Системный P1-BE-U06 | VersionMixin + exclusion list |
| A2W8-4-004 | RESIDUAL (W8-A4-004) | P2 | `sprint.go:17-43` | Нет `deleted_at` | Issue/Project soft; Sprint без trash | deleted_at **или** ADR no-delete lifecycle |
| A2W8-4-005 | RESIDUAL (W8-A4-008) | P2 | `recording.go:18-56` | Нет soft-delete media metadata | Нет корзины/GDPR soft-path | deleted_at + List filter |
| A2W8-4-006 | RESIDUAL (W8-A4-009) | P2 | `poll.go` (+ option/vote) | Нет soft-delete / OnDelete children | User content + FK trap | soft Poll **или** tx delete + OnDelete |

## Чисто

| Файл | Проверено |
|------|-----------|
| `outboxevent.go` | UUID; append-only worker index |
| `pinnedmessage.go` | unique (channel, message_id Mongo) |
| `projectmember.go` / `reaction.go` | junction unique OK |
| `session.go` | refresh_token Sensitive; ephemeral revoke |
| `pushsubscription.go` | Sensitive FIXED |
| `role.go` Indexes | unique (project, name) FIXED |
| `organization.go` invite_token | Sensitive FIXED |

UUID PK: **17/17**. Soft-delete в A4: project, scheduledcall · нет у sprint/poll*/recording/org.

## Сводка counts (открытые)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 3 |
| P3 | 0 |
| **Total open** | **6** |

FIXED-OK: 4 · NEW: 0 · REGRESS: 0
