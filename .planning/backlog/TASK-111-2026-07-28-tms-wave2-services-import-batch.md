# TASK-111 — 🐛 TMS W2 residual: tx/N+1/import/bases

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Открыта  
> Источник: A2W2 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

Все open W2 вне TASK-091 закрыты (tx, N+1, parser caps, import limits, bases).

## [ROLE]

- Role: be2
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/tms/**
  - backend/internal/bases/**
  - testframework/tms bases
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать TASK-015/016 closures
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W2-1-002/003/005, A2W2-2-002..005/007..011, A2W2-3-*, A2W2-4-*, A2W2-5-* (24)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - чеклист 24 ID в отчёте
  - hubtest green
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
