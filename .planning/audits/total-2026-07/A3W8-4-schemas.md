# A3 W8 агент 4 — ent/schema part 4 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W8-4-schemas.md](A2W8-4-schemas.md) · A1: [A8-4-schemas.md](A8-4-schemas.md)  
> Assignment: [W8-ASSIGNMENTS.md](W8-ASSIGNMENTS.md) § A4 · Периметр: **17** · Прочитано: **17/17** · Пропущено: **0**  
> Фокус: UUID / version / soft-delete / indexes / OnDelete  
> Remediation: TASK-095 (VersionMixin) · TASK-096 (OnDelete) · TASK-097 (soft-delete sprint/recording)

## Сверка с A2 (A2W8-4-*)

| A2 ID | Tag | Статус | Комментарий |
|-------|-----|--------|-------------|
| A2W8-4-001 | FIXED-OK | закрыт | `project.go:55-59` `VersionMixin{}`; Update — `VersionEQ`+`AddVersion` (`project/repository.go:213-216`); migrate TASK-095 |
| A2W8-4-002 | FIXED-OK | закрыт | `project.go:66-98` Cascade на owned children; `organization.go:54-61` `NoAction` users/projects; ADR-ENT-ONDELETE + `schema_ondelete_task096_test.go` |
| A2W8-4-003 | FIXED-OK | закрыт | VersionMixin на org/poll/role/savedfilter/scheduledcall/sprint (+ projecttemplate); guard `schema_version_mixin_task095_test.go` |
| A2W8-4-004 | FIXED-OK | закрыт | `sprint.go:43-46,75-76` `deleted_at` + indexes; SoftDeleteSprint + list filters; `soft_delete_task097_test.go` |
| A2W8-4-005 | FIXED-OK | закрыт | `recording.go:56-60,90-91` `deleted_at` + indexes; SoftDeleteRecording + List `DeletedAtIsNil`; TASK-097 |
| A2W8-4-006 | FIXED-OK | закрыт | `poll.go:70-73` / `polloption.go:44-45` Cascade options/votes; soft-delete Poll **не** требовался после ADR (OnDelete закрывает finding) |

**REGRESS:** 0 · **A2 open residual:** 0

### Сверка Stage A (W8-A4-*) — FIXED-OK cascade

| A1 ID | Tag на A3 | Evidence |
|-------|-----------|----------|
| W8-A4-001…003, 012 | FIXED-OK | VersionMixin + Project Cascade / Org NoAction |
| W8-A4-004, 008 | FIXED-OK | sprint/recording `deleted_at` (TASK-097) |
| W8-A4-005 | FIXED-OK | `role.go:58-62` unique `(project, name)` |
| W8-A4-006 | FIXED-OK | `organization.go:33-36` `invite_token` Sensitive |
| W8-A4-007 | FIXED-OK | `pushsubscription.go:28-34` p256dh/auth Sensitive |
| W8-A4-009…010 | FIXED-OK | Poll Cascade; Org NoAction |
| W8-A4-011 | FIXED-OK | godoc «Рёбра» (без mojibake) |

Recording `version` — **ADR exclusion** (unclear-12 / ADR-ENT-VERSION-MIXIN §4) → не residual.

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/organization.go` | 69 | 0 | UUID + VersionMixin + NoAction FIXED |
| `backend/ent/schema/outboxevent.go` | 51 | 0 | чисто (append-only) |
| `backend/ent/schema/pinnedmessage.go` | 51 | 0 | чисто (unique channel+message) |
| `backend/ent/schema/poll.go` | 83 | 0 | VersionMixin + Cascade FIXED |
| `backend/ent/schema/polloption.go` | 54 | 0 | Cascade votes FIXED |
| `backend/ent/schema/pollvote.go` | 62 | 0 | чисто (junction unique) |
| `backend/ent/schema/project.go` | 107 | 0 | VersionMixin + Cascade + soft FIXED |
| `backend/ent/schema/projectmember.go` | 54 | 0 | чисто (junction) |
| `backend/ent/schema/projecttemplate.go` | 66 | 0 | VersionMixin OK |
| `backend/ent/schema/pushsubscription.go` | 64 | 0 | Sensitive FIXED; version exclusion ADR |
| `backend/ent/schema/reaction.go` | 52 | 0 | чисто (junction) |
| `backend/ent/schema/recording.go` | 93 | 0 | deleted_at FIXED; version exclusion ADR |
| `backend/ent/schema/role.go` | 63 | 0 | VersionMixin + unique name FIXED |
| `backend/ent/schema/savedfilter.go` | 74 | 0 | VersionMixin FIXED |
| `backend/ent/schema/scheduledcall.go` | 103 | 0 | VersionMixin + soft FIXED |
| `backend/ent/schema/session.go` | 52 | 0 | чисто (refresh Sensitive) |
| `backend/ent/schema/sprint.go` | 78 | 2 | soft+version FIXED; NEW index/OnDelete |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W8-4-001 | FIXED | P2 | `sprint.go` Indexes | Индекс `project+status` **не** Unique; нет partial unique на `status=active` (+ `deleted_at IS NULL`) | DB не запрещает ≥2 active sprint в проекте; app-guard `StartSprintAtomic` — единственная защита (W19 C-002 / W1-A4-009). Soft-delete усложняет полный Unique | **FIXED TASK-170:** `IndexWhere("status = 'active' AND deleted_at IS NULL").Unique()` + migrate `20260729_task170_…sql`; hubtest `TestConstraint_TASK170_*` |
| A3W8-4-002 | FIXED | P3 | `sprint.go` Edges | `edge.To("issues")` без `entsql.OnDelete(SetNull)` | ADR optional_ref: Issue.sprint optional; hard DELETE sprint при живых issues → FK trap (soft-delete путь OnDelete не вызывает; PermanentDelete project каскадит оба) | **FIXED TASK-170:** `Annotations(entsql.OnDelete(entsql.SetNull))` + migrate FK SET NULL; hubtest `TestCascade_TASK170_SprintHardDeleteSetsIssueSprintNull` |

> Remediation 2026-07-29 · TASK-170 · be1 → be-ver. Migrate SQL: `needs-owner-apply` (pre-migrate cleanup duplicate actives).

## Чисто (подтверждено после remediation)

| Файл / класс | Проверено |
|--------------|-----------|
| UUID PK | **17/17** `field.UUID("id").Default(uuid.New).Immutable()` |
| VersionMixin | org, poll, project, projecttemplate, role, savedfilter, scheduledcall, sprint |
| Soft-delete | project, scheduledcall, sprint, recording (+ indexes `deleted_at`) |
| OnDelete | Project Cascade owned_child; Org NoAction; Poll/Option Cascade; Sprint→issues SetNull (TASK-170) |
| Sensitive | org `invite_token`; push p256dh/auth; session refresh |
| Junction/append-only | outbox, pinnedmessage, projectmember, pollvote, reaction, session — version/soft N/A |
| Recording/Push version | ADR unclear-12 exclusion — не gap |

**Примечание (не finding слота):** TASK-095 сервисный pilot VersionEQ закрыт для project/sprint/channel/comment; `ClosePoll` / `UpdateCustomRole` / ScheduledCall Update ещё без `VersionEQ` — зона доменов messenger/rbac/calls, не schema part4.

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 0 |
| **Total open** | **0** |

FIXED-OK (A2): **6/6** · NEW→FIXED (A3): **2/2** · REGRESS: **0** · RESIDUAL A2: **0**
