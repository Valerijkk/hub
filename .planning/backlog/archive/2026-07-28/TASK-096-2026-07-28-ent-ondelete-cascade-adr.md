# TASK-096 — 📐 Ent OnDelete/cascade: ADR + FK policy

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта · MUTATION-2026-07-29-W6  
> Источник: A2W8 OnDelete · ADR-first · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008
> Закрыто: 2026-07-28 · be1 · ADR-ENT-ONDELETE-CASCADE-2026 Accepted (owner: апрув migrate не требуется)

## [OUTCOME]

Parent→child FK имеют явную OnDelete/purge-политику по ADR; нет «тихих» orphan rows на удалении родителя.

## [ROLE]

- Role: arch→be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/ent/schema/**
  - migrate + service purge где app-level
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - слепой CASCADE без ADR
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **Owner 2026-07-28: апрув ADR/migrate НЕ требуется** (выполнено)

## [REFERENCES]

- Finding IDs: A2W8-1-003, A2W8-4-002/006, A2W8-3-003 (4)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR cascade policy
  - entsql.OnDelete или документированный app purge + тесты
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
