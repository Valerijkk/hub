# TASK-083 — 🐛 FE Trash: «Удалить навсегда» вызывает hard-delete API

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Открыта  
> Источник: A2W10-A5-001 · Stage A2 dedup · TASK-008

## [OUTCOME]

Кнопка «Удалить навсегда» вызывает backend hard-delete и обновляет список; toast без API запрещён.

## [ROLE]

- Role: fe2
- Responsibility: фикс + regression/mutation-ready тест
- Handoff to: be-ver/fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - frontend/src/routes/_authenticated/trash.tsx
  - frontend/src/api/** (trash/hard-delete client)
  - при отсутствии API — согласовать с be, не фейковый toast
  - `.planning/backlog/README.md` + этот файл
- Do not touch:
  - admin listTrash path TASK-032 без нужды
- Constraints: Conventional Commits RU description; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A (кроме schema)

## [REFERENCES]

- A2W10-A5-001 · trash.tsx:455-460 — toast.success без fetch
- Dedup: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - onClick → API hard-delete + invalidate/refetch
  - ошибка API → toast.error, элемент остаётся
  - Мутация: stub API fail → нет success toast
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
