# 🧹 C-W19: HR pending leave error≠empty

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **На проверке**  
> Источник: [`C-W19-fe-ai-platform.md`](../audits/total-2026-07/C-W19-fe-ai-platform.md) §C-W19-005  
> Исполнитель: **fe2** · Finding ID: `C-W19-FE-005`

## [OUTCOME]

Admin секция `listPendingLeave` показывает alert при `isError`, а не «нет заявок» при fail.

## [ROLE]

- Role: fe2 · Handoff: fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/src/modules/hr/leave-requests.tsx` (или актуальный path); Vitest HR
- Optional adjacent P3: `pipeline-status.tsx` sanitize `html_url` (если в том же PR удобно)
- Commits EN/RU; no AI co-author

## [REFERENCES]

- `myError` уже обработан; pending admin gate только `length > 0`
- Класс A3W10-1-004 marketplace pending

## [ACCEPTANCE]

- `isError` + QueryErrorAlert/retry на pending leave; empty только при success+len0
- Mutation: убрать isError → «нет заявок» при fail RED
- Stop: FIXED · «На проверке»
