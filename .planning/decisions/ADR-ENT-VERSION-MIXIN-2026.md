# ADR-ENT-VERSION-MIXIN-2026 — Ent VersionMixin + optimistic lock rollout

- Статус: **Accepted**
- Дата: 2026-07-28
- Автор: arch · Верифицирует: arch-ver · Реализация: be1 (TASK-095)
- Связанные: NFR-24, TZ optimistic lock, P1-BE-U06, TASK-095,
  `.planning/audits/planning-audit-2026-07/NFR-24-VERSION-CLASSIFICATION.md`

## Контекст

TZ/NFR-24 требуют `version` на изменяемых Ent. Сейчас часть `mutable_needs_version`
уже имеет `field.Int("version")`; Mixin отсутствует; A2W8 фиксирует gap-класс.
Без exclusion list задача бесконечна (P1-BE-U06).

## Рассмотренные варианты

1. **VersionMixin на все `mutable_needs_version`** — выбрано
2. Сузить NFR-24 до allow-list — отклонено: ослабляет TZ-инвариант
3. Только high-risk (project/user/channel/…) — отклонено: остаток A2W8 open; неоднородный паттерн

## Решение

1. Добавить `backend/ent/schema/mixin_version.go`:
   `VersionMixin` → `field.Int("version").Default(1).Positive()`
   Comment: `Optimistic locking (NFR-24 / ADR-ENT-VERSION-MIXIN-2026)`
2. Все `mutable_needs_version` используют Mixin (или уже имеют идентичное поле —
   не дублировать в `Fields`).
3. Exclusion (документировать здесь):
   - **append_only**: ActivityLog, AuditLog, OutboxEvent, и аналоги
   - **immutable_link**: ChannelMember, IssueLink, Reaction, и аналоги
     - **Amendment 2026-07-29 (TASK-167 / A3W8-1-002):** ChannelMember junction
       identity остаётся `immutable_link`; preference updates (`is_starred` /
       `muted_until`) — **LWW accepted** (см. ADR-ENT-CHANNELMEMBER-LWW-2026).
       VersionMixin на `channel_members` отложен (параллельный schema-wave /
       codegen approve).
4. **Unclear (12)** — exclusion до owner reclassify (не добавляем version):
   apikey, call, channelinvite, coffeeenrollment, coffeeround, filemetadata,
   invite, notification, pushsubscription, recording, session, timelog.
   Finding A2W8-1-002/006 для unclear → FIXED docs+PO exclusion, не schema.
5. Update-пути: `Where(VersionEQ(v)).AddVersion(1)`; conflict → 409;
   `version≤0` → 400 InvalidInput (паттерн CRM/bases).
6. Migrate: additive `ALTER TABLE … ADD COLUMN version INT NOT NULL DEFAULT 1` —
   SQL в `backend/migrations/20260728_task095_version_mixin.sql`.

## Finding IDs (TASK-095)

A2W8-1-001/002/006, A2W8-2-001..003/006, A2W8-3-007, A2W8-4-001/003,
A2W8-5-001..007, A2W1-A4-001, W7-A2-003 (19 rows).

## Последствия

- (+) Единый TZ-паттерн; новые схемы копируют Mixin
- (−) ~31 ALTER + ent codegen + сервисные VersionEQ wave
- (−) CI/регрессия по многим доменам

## Approval record

- Owner approve ADR: ✅ Accepted 2026-07-28 (owner decision: approve не блокирует)
- Owner approve migrate apply (test/prod): ✅ Accepted 2026-07-28
- Owner reclassify unclear (12): ✅ exclusion (не version) 2026-07-28
