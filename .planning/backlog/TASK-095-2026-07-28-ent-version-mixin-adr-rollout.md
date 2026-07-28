# TASK-095 — 📐 Ent VersionMixin: ADR + rollout version

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: A2W8 VersionMixin · ADR-first · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Все mutable Ent-схемы из класса A2W8 VersionMixin имеют field version + optimistic lock в update-путях (после принятого ADR P1-BE-U06).

## [ROLE]

- Role: arch→be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/ent/schema/**
  - Atlas migrate (только после owner approve)
  - `.planning/decisions/` ADR draft
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - генерированный ent/* кроме schema; прод без approve migrate
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: требуется ADR/owner approve перед кодом/migrate

## [REFERENCES]

- Finding IDs: A2W8-1-001/002/006, A2W8-2-001..003/006, A2W8-3-007, A2W8-4-001/003, A2W8-5-001..007, A2W1-A4-001, W7-A2-003 (19 rows)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR принят владельцем
  - version на схемах класса + VersionEQ в сервисах
  - migrate применён в test
  - мутация: убрать version check → тест краснеет
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
