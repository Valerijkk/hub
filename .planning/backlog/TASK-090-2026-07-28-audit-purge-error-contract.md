# TASK-090 — 🔒 Audit: purge fail-closed + admin error contract

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Открыта  
> Источник: A2W6 · Stage A2 dedup · TASK-008

## [OUTCOME]

Purge возвращает ошибку наружу (не fail-open); admin handler не отдаёт err.Error() в 500 body.

## [ROLE]

- Role: be1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/audit/**
  - backend/testframework/audit/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - GDPR eraser TASK-079 без нужды
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W6-A4-001, A2W6-A5-002, A2W6-A4-007
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - purge error propagates
  - тест fail-open перевёрнут
  - fixed user messages + slog
  - Мутация: swallow purge err → тест краснеет
- Regression: затронутый `./testframework/...` или frontend unit
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Diff summary
- Test commands + PASS
- Mutation note
- Статус задачи обновлён

## [OPERATING RULES]

- Не закрывать соседние A2 ID «заодно» без evidence
- Docs sync только если контракт API изменился
