# TASK-089 — 🔒 Messenger reaction + CRM contact + invoice serial

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Исправлена  
> Источник: A2W3/W7 · Stage A2 dedup · TASK-008

## [OUTCOME]

Concurrent AddReaction не плодит дубли; CRM deal→contact owner-scoped; invoice number не Count+1 race.

## [ROLE]

- Role: be2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/messenger/mongo_repository.go
  - backend/internal/crm/service.go
  - backend/internal/accounting/invoice_repository.go
  - testframework messenger/crm/accounting
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - FR-INTR U03/U05 greenfield
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W3-3-004 · W7-A2-005 · W7-A2-007
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Atomic reaction dedup
  - validateDealRefs GetContact owner scope
  - serial/unique invoice number
  - denial/race тесты + мутации
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
