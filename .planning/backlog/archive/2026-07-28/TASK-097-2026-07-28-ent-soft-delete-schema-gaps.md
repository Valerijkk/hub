# TASK-097 — 📐 Ent soft-delete gaps: deleted_at + list filters

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта · MUTATION-2026-07-29-W6  
> Источник: A2W8 soft-delete · ADR/migrate · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-28 · be2 · ветка `fix/ent/task-097-soft-delete`

## [OUTCOME]

Схемы LeaveRequest/Label/Sprint/Recording/Kudo/CustomField/GitHub·GitLab Integration
имеют `deleted_at` (или ADR join-filter для JournalLine) и list-пути их фильтруют.

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/ent/schema/**
  - затронутые service/repository list
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - hard delete без GDPR path
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: ✅ Owner 2026-07-28 — апрув ADR/migrate НЕ требуется (lock 254 A2)

## [REFERENCES]

- Finding IDs: A2W8-3-001/004/005, A2W8-4-004/005, A2W8-2-004, W7-A3-010 (7)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-ENT-SOFT-DELETE-GAPS-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ✅ deleted_at + DeletedAtIsNil в list (JournalLine — join-filter ADR §B)
  - ✅ migrate SQL + owner approve record
  - ✅ тесты trash/list + schema evidence
  - ✅ Каждый ID FIXED с evidence
- Stop when: green + статус «Исправлена» в README — **достигнуто**

## [OUTPUT SCHEMA]

См. handoff be2 в коммите / отчёте исполнителя.

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record ✅

## [CHECKLIST FIXED]

| ID | Evidence |
|----|----------|
| A2W8-3-001 | `leaverequest.deleted_at` + `SoftDeleteLeave` + list filters · `hr/soft_delete_task097_test.go` |
| A2W8-3-004 | Kudo `deleted_at` + SoftDeleteKudo · JournalLine join-filter ADR · tests kudo + accounting |
| A2W8-3-005 | Label/LabelGroup `deleted_at` + soft DeleteLabel · `tracker/soft_delete_task097_test.go` |
| A2W8-4-004 | Sprint `deleted_at` + SoftDeleteSprint · `sprint/soft_delete_task097_test.go` |
| A2W8-4-005 | Recording `deleted_at` + SoftDeleteRecording · `calls/soft_delete_recording_task097_test.go` |
| A2W8-2-004 | CustomFieldDefinition soft-delete · `tracker/soft_delete_task097_test.go` (version → TASK-095) |
| W7-A3-010 | GitHub/GitLab DeleteIntegration → SetDeletedAt · github/gitlab soft_delete_task097_test.go |

## [MUTATION]

Убрать `DeletedAtIsNil` из List* / заменить SetDeletedAt на hard `Delete()` → соответствующий TASK-097 тест краснеет.
