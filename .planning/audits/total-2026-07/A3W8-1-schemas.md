# A3 W8 агент 1 — ent/schema part 1 (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W8-1-schemas.md](A2W8-1-schemas.md) · Stage A: [A8-1-schemas.md](A8-1-schemas.md)  
> Assignment: [W8-ASSIGNMENTS.md](W8-ASSIGNMENTS.md) § A1 · Периметр: **17** · Прочитано: **17/17** · Пропущено: **0**  
> Фокус: UUID v4 PK · version / VersionMixin · soft-delete · indexes · OnDelete  
> Remediation: TASK-095 (VersionMixin) · TASK-096 (OnDelete) · TASK-107 (board/AIUserPreference FK)

## Сверка с A2 (A2W8-1-*)

| A2 ID | Tag | Статус | Evidence в HEAD |
|-------|-----|--------|-----------------|
| A2W8-1-001 | FIXED-OK | закрыт | `channel.go:60-65` `VersionMixin{}` + messenger `VersionEQ` (TASK-095 / ADR-ENT-VERSION-MIXIN) |
| A2W8-1-002 | FIXED-OK | закрыт | Mixin на `aiprovider`/`automationrule`/`aiuserpreference`/`board`/`channel`; **ChannelMember** → ADR `immutable_link`; **APIKey/Call/ChannelInvite/CoffeeEnrollment** → unclear-12 exclusion (docs+PO) |
| A2W8-1-003 | FIXED-OK | закрыт | `channel.go:69-96` + `call.go:66-81` `entsql.OnDelete(Cascade\|SetNull)` (TASK-096 / ADR-ENT-ONDELETE) |
| A2W8-1-004 | FIXED-OK | закрыт | `board.go:37-43,65-70` `updated_at`/`deleted_at` + `index.Edges("project").Fields("deleted_at")` (TASK-107) |
| A2W8-1-005 | FIXED-OK | закрыт | `aiuserpreference.go:22-25,52-54` FK `default_provider` → `AIProvider`; migrate `ON DELETE SET NULL` (TASK-107) |
| A2W8-1-006 | FIXED-OK | закрыт | `apikey.go` soft-delete есть; `version` — unclear-12 exclusion ADR §4 (не schema gap) |

**REGRESS:** 0 · **FIXED-OK (A2 actives):** 6 · **RESIDUAL (A2):** 0 · **NEW:** 2

### Cascade FIXED-OK (Stage A / A2 уже закрытые в периметре)

| Класс | Tag | Evidence |
|-------|-----|----------|
| CoffeeEnrollment O2O unique | FIXED-OK | `coffeeenrollment.go:44-45` `index.Edges("user").Unique()`; `user.go:143` `.Unique()` |
| Bookmark channel FK | FIXED-OK | `bookmark.go:44-48` `edge.From("channel").Field("channel_id")` |
| APIKey dup unique index | FIXED-OK | `apikey.go:62-68` — только `expires_at`/`deleted_at` |
| Mojibake «ибра» | FIXED-OK | «Рёбра» в apikey/backuprecord/board; grep `ибра` = 0 |
| UUID PK + `nowUTC` | FIXED-OK | все 17 схем |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/ent/schema/activitylog.go` | 55 | 0 | append-only; чисто |
| `backend/ent/schema/aiprovider.go` | 77 | 0 | VersionMixin FIXED-OK |
| `backend/ent/schema/aiusagedaily.go` | 45 | 0 | unique (user, date); чисто |
| `backend/ent/schema/aiuserpreference.go` | 58 | 1 | VersionMixin+FK OK; NEW OnDelete SetNull annotation |
| `backend/ent/schema/apikey.go` | 64 | 0 | soft-delete; unclear-12 version OK |
| `backend/ent/schema/auditlog.go` | 58 | 0 | append-only; чисто |
| `backend/ent/schema/automationrule.go` | 69 | 0 | VersionMixin FIXED-OK |
| `backend/ent/schema/backuprecord.go` | 51 | 0 | чисто (mojibake FIXED) |
| `backend/ent/schema/board.go` | 65 | 0 | soft+indexes+VersionMixin FIXED-OK |
| `backend/ent/schema/bookmark.go` | 53 | 0 | channel FK FIXED-OK |
| `backend/ent/schema/call.go` | 86 | 0 | OnDelete Cascade FIXED; unclear-12 version OK |
| `backend/ent/schema/callparticipant.go` | 54 | 0 | junction; чисто |
| `backend/ent/schema/channel.go` | 101 | 0 | VersionMixin+OnDelete+soft FIXED-OK |
| `backend/ent/schema/channelinvite.go` | 69 | 0 | unclear-12 version OK |
| `backend/ent/schema/channelmember.go` | 63 | 1 | NEW: mutable prefs vs immutable_link |
| `backend/ent/schema/coffeeenrollment.go` | 42 | 0 | unique OK; unclear-12 version OK |
| `backend/ent/schema/coffeepair.go` | 43 | 0 | unique (round,A,B); чисто |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Sev | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|-----|-------------|---------|---------------|------------|
| A3W8-1-001 | NEW | P3 | `aiuserpreference.go:52-54` | `edge.To("default_provider")` без `entsql.OnDelete(entsql.SetNull)`; migrate TASK-107 уже `ON DELETE SET NULL` | ADR-ENT-ONDELETE: optional_ref обязан быть **явным** в schema; codegen без annotation ≠ hand-migrate → drift при `ent generate` | `Annotations(entsql.OnDelete(entsql.SetNull))` на `default_provider` (+ guard-тест как TASK-096) |
| A3W8-1-002 | NEW | P3 | `channelmember.go:26-43` | ADR/`immutable_link` exclusion, но поля `role` / `is_starred` / `muted_until` / `notify_mode` обновляются in-place (`MuteChannel`/`ToggleStar`) без `version` | NFR-24: immutable_link = «create+delete, no meaningful PATCH»; mute/star/role — concurrent PATCH → LWW. A2W8-1-002 закрыт exclusion’ом при живом update-path | Reclassify → `mutable_needs_version` + `VersionMixin` + VersionEQ на mute/star/role **или** явное ADR «ChannelMember LWW accepted» |

## Чисто (подтверждено)

- **UUID PK** `Default(uuid.New).Immutable()` — все 17 схем.
- **VersionMixin** на mutable периметра: AIProvider, AIUserPreference, AutomationRule, Board, Channel.
- **Unclear-12 / immutable_link** (owner): APIKey, Call, ChannelInvite, CoffeeEnrollment без Mixin; ChannelMember — exclusion (см. A3W8-1-002).
- **Soft-delete:** Channel, Board, APIKey имеют `deleted_at` (+ индексы).
- **OnDelete (TASK-096):** Channel→members/invites/polls/bookmarks/… Cascade; Channel→calls/scheduled_calls SetNull; Call→participants/recordings Cascade.
- **Indexes:** Board project+deleted_at; CoffeeEnrollment user Unique; Bookmark (user,message) Unique; CallParticipant/CoffeePair composites.
- **Append-only / counters:** ActivityLog, AuditLog, AIUsageDaily, CoffeePair, BackupRecord — без version/soft по канону.
- **Project→automation_rules** Cascade — на стороне `project.go` (вне слота, без регресса).

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 0 |
| P3 | 2 |
| **Total open** | **2** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 actives 001–006) | 6 |
| RESIDUAL | 0 |
| NEW | 2 (`A3W8-1-001`, `A3W8-1-002`) |
| REGRESS | 0 |

A2 baseline open: 6 → **FIXED-OK 6** · RESIDUAL 0 · NEW 2 · REGRESS 0 · open **2** (P3=2).

Код на A3 **не менялся** — только отчёт.
