# TASK-093 — 🧹 FE A2: error≠empty residual + a11y delete focus

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: Открыта  
> Источник: A2W9/W10/W11 · Stage A2 dedup · TASK-008

## [OUTCOME]

Оставшиеся A2 FE residual из TASK-047/048/049: isError≠empty, catch→toast, presence offline≠error; notification delete keyboard-visible.

## [ROLE]

- Role: fe1
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - frontend/src/modules/tracker/sprint-backlog.tsx
  - frontend/src/modules/** (по списку A2W10)
  - frontend/src/components/**/notification-panel.tsx
  - frontend/src/hooks/use-presence.ts
  - testframework unit frontend
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - новые фичи UI
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W10 · A2W9-4-001/002 · A2W11-3-001 · TASK-047/048
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - Чеклист ID из A2-DEDUP закрыт или WONTFIX с причиной
  - notification-panel group-focus-within для delete
  - ≥3 regression RTL
  - Мутация: silent catch → тест краснеет
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
