# TASK-102 — 📐 Telegram login surface: ADR-AUTH-01 cleanup

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: A2W6 Telegram · ADR-AUTH-01 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Telegram не является login: API/тесты login удалены или quarantine; остаётся только notification bind (ADR-AUTH-01).

## [ROLE]

- Role: be1
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/auth/**
  - testframework/auth/**
  - frontend auth telegram widgets если есть
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать notification bind
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: требуется ADR/owner approve перед кодом/migrate

## [REFERENCES]

- Finding IDs: A2W6-A2-001, A2W6-A1-001/002/005/006, A2W6-A2-002/003 (7)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - login endpoints/tests gone or 410
  - positive Telegram-login test removed
  - bind-only path green
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
