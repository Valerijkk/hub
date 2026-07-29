# 🧹 C-audit: calls calendar/dialog error≠empty + a11y close

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **На проверке**  
> Источник: [`AUDIT-FE-messenger-calls-2026-07-29.md`](../audits/total-2026-07/AUDIT-FE-messenger-calls-2026-07-29.md)  
> Исполнитель: **fe2** · Finding IDs: `FE-MC-001`…`FE-MC-003` (+ optional P3 `FE-MC-004`…`006`)

## [OUTCOME]

Fail `getChannels` / `listProjects` в call calendar и create-call dialog показывает alert/retry, а не пустой календарь / «нет проектов»; кнопка закрытия popover имеет accessible name.

## [ROLE]

- Role: fe2 · Handoff: fe-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Allowed: `frontend/src/modules/calls/call-calendar.tsx`, `create-call-dialog.tsx`; optional `calendar-week-view.tsx` a11y labels; Vitest calls domain
- Do not touch: messenger modules LinkPreview (TASK-180); recordings noopener (TASK-124)
- Commits EN/RU; no AI co-author

## [REFERENCES]

- FE-MC-001: channels query без isError → пустая сетка
- FE-MC-002: projects/channels без isError → «нет проектов/каналов»
- FE-MC-003: popover `×` без aria-label
- Audit report above

## [ACCEPTANCE]

- 001/002: isError + alert/retry; empty только success+len0
- 003: close control имеет accessible name
- Optional 004–006 a11y labels в том же PR
- Mutation: убрать isError → empty-on-fail RED
- Stop: 3/3 FIXED · «На проверке»

## Progress (fe2 · 2026-07-29)

| ID | Status | Evidence |
|----|--------|----------|
| FE-MC-001 | FIXED | `call-calendar.tsx`: `isError`/`refetch` channels + `QueryErrorAlert`; grid hidden while error/loading |
| FE-MC-002 | FIXED | `create-call-dialog.tsx`: projects/channels `isError` → alert/retry; empty copy only success+len0 |
| FE-MC-003 | FIXED | popover `×` `aria-label="Закрыть"` |
| FE-MC-004 | FIXED | channel `SelectTrigger` `aria-label="Выбор канала"` |
| FE-MC-005 | FIXED | remove participant `aria-label={`Удалить участника ${name}`}` |
| FE-MC-006 | FIXED | week slot label includes weekday+date+hour |

**Mutation:** force `channelsError = false` → `call-calendar-error-empty` RED (alert missing); restored → GREEN.

**Vitest:** `call-calendar-error-empty` + `create-call-dialog-error-empty` + `call-calendar-a11y` + `recordings-download-noopener` — **9/9 PASS**.

**OOS:** README (scope ban); TASK-008 not closed; messenger LinkPreview (TASK-180).
