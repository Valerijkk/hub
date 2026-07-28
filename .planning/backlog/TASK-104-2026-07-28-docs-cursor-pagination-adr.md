# TASK-104 — 📐 Docs/API: cursor vs offset list-issues

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: **Исправлена**  
> Источник: A2W11 pagination · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыто: 2026-07-28 · be1+docs · ветка `fix/tracker/task-104-pagination`

## [OUTCOME]

TZ/rule-60 и фактический list-issues согласованы: **cursor** (+ limit) — один канон (`ADR-PAGINATION-CURSOR-ISSUES-2026`).

## [ROLE]

- Role: sa→be1+docs
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - docs/**
  - backend tracker list handlers
  - .cursor/rules/60*
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать chat cursor
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **N/A (owner lock 2026-07-28: апрув не требуется)** — канон = код; ADR зафиксирован как Accepted

## [REFERENCES]

- Finding IDs: A2W11-2-002, A2W11-5-002 (2) — **FIXED**
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- ADR: `.planning/decisions/ADR-PAGINATION-CURSOR-ISSUES-2026.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - ADR/decision ✅
  - docs+code sync ✅
  - тест пагинации ✅
  - Каждый ID из списка — FIXED с evidence ✅
- Stop when: green + статус «Исправлена» в README ✅

## [OUTPUT SCHEMA]

### Checklist ID → evidence

| ID | Status | Evidence |
|----|--------|----------|
| A2W11-2-002 | FIXED | ADR + TZ/rule-60/NFR-14 cursor; `TestListIssuesCursorPagination_Task104`; HTTP `TestHandler_ListIssues_CursorPagination_Task104` |
| A2W11-5-002 | FIXED | то же; audit `A2W11-5-fe-lib-docs.md` → FIXED-OK |

### Test commands + PASS

```text
go test -tags hubtest -race -count=1 -timeout=120s ./testframework/tracker/ -run "CursorPagination_Task104|ListIssuesPagination"
```

### Mutation note

- Убрать `Cursor` из `ListIssuesRequest` / вернуть offset-only → service/HTTP cursor tests FAIL.
- Сломать `HasMore` (не `limit+1`) → `HasMore` assertions FAIL на mid-pages.
- Chat cursor не менялся.

### README status update

Статус TASK-104 → **Исправлена**.

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код list-issues не менялся (уже cursor); docs выровнены под код
