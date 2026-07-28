# TASK-119 — 🔒 Storage/backup/dashboard residual hygiene

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Открыта  
> Источник: W7 storage residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Sanitize backup errors; dashboard fail-visible; handler 404; argv secrets; layering — W7-A4 residual.

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/{storage,backup,dashboard}/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - NF-09/MaintenanceGate ADR (099/100)
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: W7-A4-007..009/011..013/017/018 (8)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 8 ID + tests
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
