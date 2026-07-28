# TASK-097 — 📐 Ent soft-delete gaps: deleted_at + list filters

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: A2W8 soft-delete · ADR/migrate · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Схемы LeaveRequest/Label/Sprint/Recording/Poll/… из A2W8/W7 имеют deleted_at и list-пути их фильтруют.

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
- Approval record: требуется ADR/owner approve перед кодом/migrate

## [REFERENCES]

- Finding IDs: A2W8-3-001/004/005, A2W8-4-004/005, A2W8-2-004, W7-A3-010 (7)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - deleted_at + DeletedAtIsNil в list
  - migrate owner approve
  - тесты trash/list
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record
