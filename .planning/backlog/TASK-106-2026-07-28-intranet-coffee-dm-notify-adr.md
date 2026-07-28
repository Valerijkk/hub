# TASK-106 — 📐 Intranet FR-INTR-11: auto-DM vs notify

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Открыта  
> Источник: W7-A2-008 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

FR-INTR-11 реализован по ADR U05 (auto-DM или notify-only) — поведение не «полусделано».

## [ROLE]

- Role: product→be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/intranet/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - спам DM без ADR
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: требуется ADR/owner approve перед кодом/migrate

## [REFERENCES]

- Finding IDs: W7-A2-008 U05 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR U05
  - код+тест выбранного варианта
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
