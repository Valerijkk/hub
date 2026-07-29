# TASK-091 — 🐛 TMS: createBug runId binding + reports batch

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Закрыта · MUTATION-2026-07-29-W4
> Источник: A2W2 · Stage A2 dedup · TASK-008

## [OUTCOME]

createBug берёт runId из URL и сверяет result∈run; report endpoints без N+1 last-result на каждый case.

## [ROLE]

- Role: be2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - backend/internal/tms/handler_test_cases.go
  - backend/internal/tms/service_reports.go
  - backend/testframework/tms/**
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - TMS IDOR TASK-015 (уже закрыт)
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W2-1-001/004 · A2W2-2-001
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - URL :runId + ownership check
  - batch IDIn / join для reports
  - тесты + мутации
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
