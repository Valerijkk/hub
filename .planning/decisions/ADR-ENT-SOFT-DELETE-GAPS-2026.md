# ADR-ENT-SOFT-DELETE-GAPS-2026 — Ent soft-delete gaps (TASK-097)

- Статус: **Accepted**
- Дата: 2026-07-28
- Автор: be2 · Верифицирует: be-ver · Реализация: be2 (TASK-097)
- Связанные: TZ soft-delete / корзина 30д, A2W8-3-001/004/005, A2W8-4-004/005,
  A2W8-2-004, W7-A3-010, TASK-097

## Контекст

A2W8/W7 фиксируют сущности без `deleted_at` при TZ-инварианте «soft delete →
корзина 30д; hard delete только Admin/GDPR». Hard `Delete().Exec` на Label,
CustomField, GitHub/GitLab Integration необратим (токены, каталог, PII).

## Рассмотренные варианты

1. **Добавить `deleted_at` + `DeletedAtIsNil` в list/get + soft-delete в delete-путях** — выбрано для самостоятельных сущностей
2. ADR-исключение hard-delete для credentials — отклонено: W7-A3-010 (токены/links)
3. Только join-filter без колонки — только для child-строк журнала (см. ниже)

## Решение

### A. Самостоятельные сущности → `deleted_at` + index

| Schema | Finding | Delete path |
|--------|---------|-------------|
| `LeaveRequest` | A2W8-3-001 | `SoftDeleteLeave` (GDPR/trash); list/get `DeletedAtIsNil` |
| `Label`, `LabelGroup` | A2W8-3-005 | `DeleteLabel` → `SetDeletedAt`; lists filter |
| `CustomFieldDefinition` | A2W8-2-004 | `DeleteCustomField` → soft; list filter |
| `Sprint` | A2W8-4-004 | `SoftDeleteSprint`; list/get/active filter |
| `Recording` | A2W8-4-005 | `SoftDeleteRecording`; list/get filter |
| `Kudo` | A2W8-3-004 | `SoftDeleteKudo` (GDPR); feed/received filter |
| `GitHubIntegration`, `GitLabIntegration` | W7-A3-010 | `DeleteIntegration` → soft; list/get/find filter |

`version` / VersionMixin на Label/LabelGroup/CustomFieldDefinition/Sprint/Leave —
зона **TASK-095** (ADR-ENT-VERSION-MIXIN); этот ADR не дублирует Mixin, чтобы
не конфликтовать с веткой `fix/ent/task-095-version-mixin`.

Poll soft-delete / OnDelete children — зона **TASK-096** (A2W8-4-006), не здесь.

### B. JournalLine → join-filter + ADR (без собственной колонки)

`JournalLine` — append-only child проводки. Soft-delete родителя
`JournalEntry.deleted_at` уже есть; list/trial-balance фильтруют
`HasEntryWith(…, DeletedAtIsNil())`. Отдельный `deleted_at` на line ломал бы
двойную запись (часть строки «исчезла», баланс XOR).

**Политика:** видимость line = видимость entry; hard-delete line запрещён вне
GDPR cascade родителя (OnDelete — TASK-096).

### C. Migrate

Additive SQL: `backend/migrations/20260728_task097_soft_delete_gaps.sql`.
Dev/test: Ent `Schema.Create` / enttest подхватывает schema.

## Finding IDs (TASK-097)

A2W8-3-001, A2W8-3-004, A2W8-3-005, A2W8-4-004, A2W8-4-005, A2W8-2-004, W7-A3-010.

## Последствия

- (+) TZ-корзина / GDPR soft-path; токены интеграций не уничтожаются мгновенно
- (−) Unique `(project, name)` / `(project, repo)` может блокировать recreate
  одноимённой сущности после soft-delete (как у Component/Environment — accepted)
- (−) ent codegen + регрессия затронутых доменов

## Approval record

- Owner approve ADR: ✅ Accepted 2026-07-28 (owner: апрув ADR/migrate НЕ требуется для lock 254 A2)
- Owner approve migrate apply (test/prod): ✅ Accepted 2026-07-28
