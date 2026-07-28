# ADR-ENT-ONDELETE-CASCADE-2026 — Ent OnDelete / cascade FK policy

- Статус: **Accepted** (owner 2026-07-28: отдельный approve ADR/migrate не требуется)
- Дата: 2026-07-28
- Автор: arch · Реализация: be1 · Верифицирует: arch-ver / be-ver
- Связанные: P1-BE-U05, P2-BE-U03, A8-005, TASK-096,
  `.planning/audits/total-2026-07/A2W8-{1,3,4}-schemas.md`

## Контекст

Во всём `ent/schema/` не было явных `entsql.OnDelete`. Ent по умолчанию ставит
`NO ACTION` (Required) / `SET NULL` (Optional) — при hard-purge родителя либо
FK trap, либо «тихие» orphan rows. Soft-delete (корзина 30д) остаётся основным
user-path; hard-purge — Admin/GDPR / PermanentDelete / retention.

## Рассмотренные варианты

1. **Неявный default NO ACTION** — статус-кво; отклонено (FK trap на PermanentDelete).
2. **Только app purge без schema annotations** — отклонено: политика не видна в схеме.
3. **Явный OnDelete per-edge: Cascade на owned_child, SetNull на optional,
   NoAction на Org** — **выбрано**. Soft-delete не триггерит OnDelete;
   PermanentDelete = намеренный hard DELETE с DB CASCADE.

## Решение

### Классы рёбер

| Класс | OnDelete | Примеры | Почему |
|-------|----------|---------|--------|
| **owned_child** | `CASCADE` | Channel→members/polls/…; Call→participants; Poll→options/votes; Project→issues/channels/TMS/catalog; Issue→comments/links | Hard DELETE родителя обязан убрать детей; soft-delete `deleted_at` FK OnDelete **не** вызывает |
| **optional_ref** | `SET NULL` | Channel→calls/scheduled_calls (optional FK) | Звонок может жить без канала (instant) |
| **org_anchor** | `RESTRICT` (`NoAction`) + playbook | Org→users/projects | 1 org/instance; запрещён CASCADE wipe пользователей |

### Soft vs hard

1. **User delete** → soft-delete (`deleted_at`); строки и FK остаются; OnDelete не срабатывает.
2. **PermanentDelete / retention hard-purge** → `DELETE` строки → DB CASCADE по owned_child.
3. **Organization**: только GDPR/reset playbook; `NoAction` на users/projects.

### IssueLink (A2W8-3-003)

- Unique: `(source, target, link_type)` — разрешены blocks+relates на одной паре.
- Иерархия эпик/подзадача — **только** `Issue` self-edge `parent`/`children`.
- Значения `link_type` `parent`/`child` **удалены** из enum (dual model закрыт).

### Migrate

- Schema annotations → `ent generate` обновляет `ent/migrate/schema.go`.
- Additive SQL в `backend/migrations/20260728_task096_ondelete_cascade.sql`
  для живых Postgres (ALTER FK / unique). Owner-approve migrate **не требуется**
  (решение владельца 2026-07-28); применять в test/dev вместе с фичей.

## Finding IDs (TASK-096)

| ID | Closing action |
|----|----------------|
| A2W8-1-003 | Channel/Call `entsql.OnDelete(Cascade)` на owned children |
| A2W8-4-002 | Project Cascade owned_child; Org NoAction; PermanentDelete опирается на FK CASCADE |
| A2W8-4-006 | Poll/Option `Cascade` на options/votes |
| A2W8-3-003 | IssueLink unique(source,target,type); убрать parent/child |

## Последствия

- (+) Нет неявной FK-политики; PermanentDelete с детьми проходит через CASCADE
- (+) Soft-delete path не меняется (OnDelete не вызывается)
- (−) ALTER FK на живых БД; Schema.Create на чистой — из codegen
- (−) S3/Mongo cleanup после hard-purge остаётся ответственностью retention-сервиса

## Approval record

- Owner: **апрув ADR/migrate не требуется** (2026-07-28)
