# ADR-ENT-CHANNELMEMBER-LWW-2026 — ChannelMember preference updates: LWW accepted

- Статус: **Accepted**
- Дата: 2026-07-29
- Автор: be1 (TASK-167) · Верифицирует: arch-ver / be-ver · Approval: owner lock
- Связанные: NFR-24, ADR-ENT-VERSION-MIXIN-2026, A3W8-1-002, TASK-167,
  `.planning/audits/total-2026-07/A3W8-1-schemas.md`

## Контекст

`ChannelMember` в ADR-ENT-VERSION-MIXIN отнесён к **immutable_link** (create+delete,
без meaningful PATCH). Фактически prefs обновляются in-place:

- `is_starred` — `ToggleStar`
- `muted_until` — `MuteChannel` / `UnmuteChannel`
- `notify_mode` / `role` — нет user-facing PATCH на момент ADR (role только на AddMembers)

Concurrent mute/star без `version` = last-write-wins (LWW). A3W8-1-002 требует
либо VersionMixin+VersionEQ, либо явное owner accept LWW.

Полный `ent generate` для VersionMixin на `channel_members` переписывает весь
derived Ent-слой и пересекается с параллельным schema-wave TASK-171
(`team`/`workflowstatus`/…). Owner lock на TASK-167: путь **LWW accepted**,
без DDL `version` и без codegen в этой задаче.

## Рассмотренные варианты

1. **VersionMixin + VersionEQ на mute/star** — корректный NFR-24 RMW; отложен:
   codegen/migrate approve + изоляция от TASK-171.
2. **LWW accepted для preference updates** — **выбрано** (owner lock 2026-07-29).
3. **Убрать in-place updates (только recreate membership)** — отклонено: ломает UX star/mute.

## Решение

1. `ChannelMember` остаётся в exclusion **immutable_link** (ADR-ENT-VERSION-MIXIN §3)
   для junction identity (`channel_id`+`user_id`); **preference fields** явно
   допускают LWW concurrent updates без `version`.
2. Finding A3W8-1-002 → FIXED docs+PO (не schema gap).
3. Re-open VersionMixin, если появится admin PATCH `role`/`notify_mode` с
   multi-writer conflict risk **или** когда schema-wave TASK-171 смержен и
   owner даст migrate/codegen approve — отдельная задача.

## Последствия

- (+) Нет DDL/codegen drift с TASK-171; mute/star API без `version` в body.
- (−) Concurrent star/mute остаются LWW (приемлемо для per-user prefs).
- (−) NFR-24 для ChannelMember prefs — осознанное исключение, не «забытый Mixin».

## Approval record

- Owner lock / approve LWW path (TASK-167 A3W8-1-002): ✅ Accepted 2026-07-29
- Migrate/codegen: N/A (колонка `version` не добавляется)
