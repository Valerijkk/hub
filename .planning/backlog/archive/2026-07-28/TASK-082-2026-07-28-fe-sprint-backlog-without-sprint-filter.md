# TASK-082 — 🐛 FE SprintBacklog: фильтр задач без спринта

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта · MUTATION-2026-07-29-W3
> Источник: A2W10-A3-001 · Stage A2 dedup · TASK-008

## [OUTCOME]

Счётчик «N задач без спринта» и список показывают только issues без sprint_id (не весь listIssues).

## [ROLE]

- Role: fe1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - frontend/src/modules/tracker/sprint-backlog.tsx
  - frontend/src/api/tracker.ts (если нужен query param)
  - testframework stages unit frontend tracker (если есть)
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - backend sprint API контракт без согласования
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W10-A3-001 · sprint-backlog.tsx:28-32,73
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - listIssues с фильтром sprint_id empty / client filter после ответа
  - error path не silent catch (toast/isError)
  - Мутация: убрать фильтр → UI снова показывает все issues
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
