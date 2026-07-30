# 🧹 C-W19: HR pending leave error≠empty

> Заведено: 2026-07-29 · Автор: team-lead · Приоритет: **P3** · Статус: **Закрыта**  
> Источник: [`C-W19-fe-ai-platform.md`](../audits/total-2026-07/C-W19-fe-ai-platform.md) §C-W19-005  
> Исполнитель: **fe2** · Mutation: **qa-auto1** (W8 slot 5) · Finding ID: `C-W19-FE-005`

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
- Stop: FIXED · mutation PROVEN · «Закрыта»

## Progress note

См. Vitest `domains/hr/leave-pending-error.test.tsx` + guard в `leave-requests.tsx` (ниже criteria).

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe2 W7 slot 5). Слой **frontend**.

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| C-W19-FE-005 | `leave-requests.tsx:197-201` — убрать ветку `pendingError ? <QueryErrorAlert>`; `:185` — gate только `pendingItems.length > 0` (fail → секция скрыта = «нет заявок») | describe `LeaveRequestsPage pending error≠empty (C-W19-FE-005 / TASK-182)` → `it("shows approval queue alert when listPendingLeave fails, not silent empty")` | frontend |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| pending leave isError → alert | ✅ RTL | |
| success+empty hides queue | ✅ twin `it` | не mutation primary |
| optional pipeline-status sanitize | N/A | optional scope, не ID |
| layer_mismatch | нет | |
| `npx tsc --noEmit` | ❌ | не primary |

**Confidence:** CONFIRMED (1/1).

## Mutation prove (W8 slot 5 · qa-auto1 · 2026-07-30)

| ID | Verdict | Evidence |
|----|---------|----------|
| C-W19-FE-005 | **PROVEN** | [`MUTATION-2026-07-30-W8-qa-auto1.md`](../audits/total-2026-07/MUTATION-2026-07-30-W8-qa-auto1.md) |

Break→RED→restore→GREEN. Production restored. No commit.
