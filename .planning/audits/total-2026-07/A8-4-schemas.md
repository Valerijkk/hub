# W8 агент A4 — schemas part 4

> READ-ONLY · TASK-008 · фокус: UUID+version, soft-delete, cascade delete risks · 2026-07-28

## Прочитанные файлы
| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| backend/ent/schema/organization.go | 57 | 3 | замечания |
| backend/ent/schema/outboxevent.go | 51 | 0 | чисто |
| backend/ent/schema/pinnedmessage.go | 51 | 0 | чисто |
| backend/ent/schema/poll.go | 72 | 2 | замечания |
| backend/ent/schema/polloption.go | 51 | 1 | замечания |
| backend/ent/schema/pollvote.go | 62 | 1 | замечания |
| backend/ent/schema/project.go | 96 | 3 | замечания |
| backend/ent/schema/projectmember.go | 54 | 0 | чисто |
| backend/ent/schema/projecttemplate.go | 59 | 1 | замечания |
| backend/ent/schema/pushsubscription.go | 62 | 1 | замечания |
| backend/ent/schema/reaction.go | 52 | 0 | чисто |
| backend/ent/schema/recording.go | 87 | 1 | замечания |
| backend/ent/schema/role.go | 48 | 2 | замечания |
| backend/ent/schema/savedfilter.go | 67 | 1 | замечания |
| backend/ent/schema/scheduledcall.go | 96 | 1 | замечания |
| backend/ent/schema/session.go | 52 | 0 | чисто |
| backend/ent/schema/sprint.go | 65 | 2 | замечания |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Перепроверка прошлого аудита (A8 / grok-2026-07)

| Старый ID | Статус в текущем коде | Комментарий |
|-----------|----------------------|-------------|
| A8-003 | **живо** → W8-A4-001 | `project.go` всё ещё без `field.Int("version")` |
| A8-004 | **живо** (часть A4) → W8-A4-003 | Org/Poll/Role/SavedFilter/ScheduledCall/Sprint без `version` |
| A8-010 | **живо** → W8-A4-005 | `role.go` без `Indexes()` / unique `(project, name)` |
| A8-015 | **живо** (sprint) → W8-A4-004 | `sprint.go` без `deleted_at` |
| A8-005 / A8-006 | вне scope A4 | Channel/Comment — другие агенты; soft-delete уже в схемах |

## Находки
| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|----------|-------------------|------------|
| W8-A4-001 | P1 | `backend/ent/schema/project.go:17-52` | Mutable `Project` без `field.Int("version")` при наличии `updated_at`, `status`, counters, soft-delete | TZ: optimistic lock в каждой изменяемой Ent-схеме. Concurrent Update/UpdateStatus = last-write-wins (A8-003) | `field.Int("version").Default(1)`; `VersionEQ` в `project` Update/UpdateStatus |
| W8-A4-002 | P1 | `backend/ent/schema/project.go:55-87` (+ app `PermanentDelete` → `DeleteOneID`) | 25+ `edge.To(...)` без `entsql.OnDelete`; hard-purge проекта не объявляет cascade/orphan policy | `PermanentDelete` бьёт голый `DeleteOneID`: при Ent default NO ACTION — FK failure при детях; при наивном Cascade — мгновенная потеря issues/channels/TMS/… | Явный hard-purge план: soft→30d→tx cleanup children **или** `OnDelete` per-edge + тест PermanentDelete с детьми; не полагаться на неявный cascade |
| W8-A4-003 | P1 | `organization.go:20-42`; `poll.go:17-45`; `role.go:16-36`; `savedfilter.go:18-42`; `scheduledcall.go:18-63`; `sprint.go:17-43` | Mutable сущности A4 без `version` (есть `updated_at` и/или update API) | Системный TZ-gap (A8-004): гонки UpdateOrg/UpdatePoll/UpdateRole/UpdateFilter/UpdateScheduled/UpdateSprint | VersionMixin на изменяемые; junction/append-only (Outbox, Session, Vote, Reaction, Member) — исключить ADR |
| W8-A4-004 | P2 | `backend/ent/schema/sprint.go:17-43` | Нет `deleted_at`; soft-delete policy отсутствует | TZ корзина 30д; Issue/Project soft, Sprint — нет пути trash (A8-015). Update без восстановления | `deleted_at` + фильтр в List/Get **или** ADR «sprint immutable lifecycle / no delete» |
| W8-A4-005 | P2 | `backend/ent/schema/role.go:39-47` | Нет `Indexes()`; нет unique `(project, name)` | List/lookup ролей без индекса на project FK; дубли имён ролей в проекте на уровне БД (A8-010) | `index.Edges("project").Fields("name").Unique()` |
| W8-A4-006 | P2 | `backend/ent/schema/organization.go:32-35` | `invite_token` без `.Sensitive()` | Ent Stringer/debug/логи могут утечь org-wide invite secret | `.Sensitive().Unique().NotEmpty()` |
| W8-A4-007 | P2 | `backend/ent/schema/pushsubscription.go:28-33` | `p256dh_key` / `auth_key` без `.Sensitive()` | Web Push shared secrets в plain field serialization | `.Sensitive().NotEmpty()` на обоих |
| W8-A4-008 | P2 | `backend/ent/schema/recording.go:18-56` | Нет `deleted_at` у user-visible media metadata | Нет корзины/GDPR soft-path для записей звонков; только hard lifecycle | `deleted_at` + List filter; hard purge S3+row по Admin/GDPR |
| W8-A4-009 | P2 | `poll.go:48-63`; `polloption.go:34-43`; `pollvote.go:34-52` | Poll/Option/Vote без soft-delete; O2M без `OnDelete` | User content без корзины; hard-delete poll при детях → FK trap (как бывш. Channel A8-005) | Soft-delete Poll **или** tx delete votes→options→poll; объявить `OnDelete` |
| W8-A4-010 | P2 | `backend/ent/schema/organization.go:45-49` | `edge.To("users"/"projects")` без OnDelete | Org delete (редко, но GDPR/reset) — cascade/orphan не специфицирован в схеме | Explicit `OnDelete` + app purge playbook; 1-org инстанс не отменяет FK-контракт |
| W8-A4-011 | P3 | `project.go:55`; `projecttemplate.go:44` | Комментарий «ибра» вместо «Рёбра» | Косметика / mojibake-след в godoc | Исправить на «Рёбра» |
| W8-A4-012 | P3 | `backend/ent/schema/scheduledcall.go:53-56` | Soft-delete есть, `version` нет (учтено в W8-A4-003) | Concurrent Update + soft-delete без CAS | Покрыть VersionMixin вместе с W8-A4-003 |

## Чисто

Файлы без находок — что проверялось (UUID / version / soft-delete / cascade):

- **outboxevent.go** — UUID PK; append-only (`status`/`attempts`); `version`/`deleted_at` не требуются; индекс `(status, created_at)` для worker; edges нет → cascade N/A.
- **pinnedmessage.go** — UUID PK; `message_id` string = Mongo id (не PK, допустимо); unique `(channel_id, message_id)`; O2O-from channel без OnDelete — ок при soft-delete канала (children retain, A8-005 policy).
- **projectmember.go** — UUID PK; junction; unique `(user, project)`; hard-remove membership ожидаем; `version` не нужен.
- **reaction.go** — UUID PK; unique `(comment, user, emoji)`; hard-delete реакции ок; soft-delete живёт на Comment.
- **session.go** — UUID PK; `refresh_token` `.Sensitive()`; ephemeral hard-revoke; `version`/soft-delete N/A.

Частично чисто по UUID: **все 17/17** имеют `field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable()`.

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 3 |
| P2 | 7 |
| P3 | 2 |
| **Total** | **12** |

Soft-delete в A4: **есть** у `project`, `scheduledcall` · **нет** у sprint/poll*/recording/organization (и append-only/junction — ок).  
`version` в A4: **0/17** схем · mutable без lock: project, org, poll, role, savedfilter, scheduledcall, sprint.

## Топ-3

1. **W8-A4-001** (P1) — `Project` без optimistic `version` (TZ + concurrent updates).
2. **W8-A4-002** (P1) — hard-purge `Project` без объявленной cascade/orphan стратегии на 25+ children edges.
3. **W8-A4-003** (P1) — шесть mutable схем A4 без `version` (Org/Poll/Role/SavedFilter/ScheduledCall/Sprint).
