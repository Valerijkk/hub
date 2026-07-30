# 🧹 C-audit: calls calendar/dialog error≠empty + a11y close

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
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

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe2 W7 slot 5). Слой **frontend**.

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| FE-MC-001 | `call-calendar.tsx:325-330` — убрать `channelsError && <QueryErrorAlert>`; `showCalendarGrid` (`:315`) всегда true при fail | describe `CallCalendar channels error≠empty (FE-MC-001 / TASK-186)` → `it("shows alert+retry when getChannels fails, not month grid")` | frontend |
| FE-MC-002 | `create-call-dialog.tsx:296-297` / `:336-337` — убрать `projectsError`/`channelsError` → alert; `projectsEmpty`/`channelsEmpty` (`:127-128`) без `!projectsError` | describe `CreateCallDialog error≠empty (FE-MC-002 / TASK-186)` → `it("shows projects alert+retry…")`; `it("shows channels alert+retry…")` | frontend |
| FE-MC-003 | `call-calendar.tsx:472` / `:485` — убрать `aria-label="Закрыть"` с popover close | describe `Calls calendar a11y (FE-MC-003/006 / TASK-186)` → `it("event popover close control has accessible name (FE-MC-003)")` | frontend |
| FE-MC-004 | `create-call-dialog.tsx:346` — убрать `aria-label="Выбор канала"` | same CreateCallDialog describe → `it("channel SelectTrigger has accessible name (FE-MC-004)")` | frontend |
| FE-MC-005 | `create-call-dialog.tsx:407` — убрать ``aria-label={`Удалить участника ${member.name}`}`` | **GAP** — нет отдельного `it` | frontend |
| FE-MC-006 | `calendar-week-view.tsx:106` — aria-label только hour, без weekday/date | describe `Calls calendar a11y` → `it("week slot aria-label includes weekday/date, not only hour (FE-MC-006)")` | frontend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| channels/projects isError | ✅ RTL | |
| popover close name + week slot | ✅ RTL | |
| channel SelectTrigger name | ✅ RTL | |
| remove-participant aria-label (FE-MC-005) | ❌ GAP | FIXED в коде, нет `it` |
| layer_mismatch | нет | |
| `npx tsc --noEmit` | ❌ | не primary |

**Confidence:** CONFIRMED для 001–004/006; FE-MC-005 PARTIAL/GAP.

## Mutation prove (fe1 · W9 slot 5 · 2026-07-30)

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W9-fe1.md`

| ID | Verdict |
|----|---------|
| FE-MC-001 | PROVEN |
| FE-MC-002 | PROVEN |
| FE-MC-003 | PROVEN |
| FE-MC-004 | PROVEN |
| FE-MC-005 | N/A (нет `it`) |
| FE-MC-006 | PROVEN |

**5/5 runnable PROVEN** · 005 N/A · статус **Закрыта**. Коммит не создавался.
