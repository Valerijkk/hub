# FIX-W15 — Perf: indexes / Kanban / lazy / bundle / waterfalls

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** fe2 (+ non-schema BE query rewrite)  
**Правила:** no Ent schema/index churn without approval · no push master

Сфокус: **A8-001/002** (indexes under real queries), **A13-014** Kanban virtualization, **A13-015/016/017** lazy shell/i18n/routes + bundle, **A13-011/012** startup waterfalls, **A13-018** channel-list virtualization. Amplifying **A13-001** board cap already on branch.

---

## Итог

| ID | Status | Notes |
|----|--------|-------|
| A8-001 | PARTIAL / NEEDS_APPROVAL | Query rewrite FIXED: `issueProjectIDEQ` on FK `project_issues` (ListIssues / FindIssueInProject / cursor). **Index** `index.Edges("project").Fields("deleted_at")` (+ optional `(project, rank)`) → **NEEDS_APPROVAL** (Ent schema + Atlas migrate). |
| A8-002 | PARTIAL / NEEDS_APPROVAL | Query rewrite FIXED: `testCaseProjectIDEQ` on FK `project_test_cases` (ListTestCases). **Index** `(project_test_cases, deleted_at)` → **NEEDS_APPROVAL**. |
| A13-001 | FIXED (prior) | `MaxBoardIssuesPerColumn=200` in `repo_boards.go` (`456636f`) — non-schema amp for A8-001. |
| A13-011 | FIXED | Board: `effectiveBoardId` from list → `getBoardData` without effect waterfall. |
| A13-012 | FIXED | Dashboard: `Promise.all([fetchPersonal(), loadLayout("personal")])`. |
| A13-014 | FIXED | Kanban `BoardColumn`: `@tanstack/react-virtual` when ≥24 cards; motion path kept for small cols. Test: `board-column-virtual.test.tsx`. |
| A13-015 | FIXED | AppShell: lazy `AIChatPanel` / `CommandPalette` / `ShortcutsDialog` mount only when open. |
| A13-016 | FIXED | i18n: eager `common`+`help`; other ns via dynamic `import()` backend. |
| A13-017 | FIXED | Lazy routes: crm, accounting, marketplace, gantt, board (`*.lazy.tsx` + stubs). `routeTree.gen.ts` regenerated. |
| A13-018 | FIXED | Channel list: virtualize when ≥40 flat rows; section UI kept below threshold. |
| Bundle | FIXED (via 015–017) | Existing `manualChunks` retained; cold graph drops AI/search/help + heavy modules + most locale JSON. |

---

## NEEDS_APPROVAL (schema)

### A8-001 — `issues`

```go
// ent/schema/issue.go Indexes():
index.Edges("project").Fields("deleted_at"),
// optional Kanban:
// index.Edges("project").Fields("rank"),
```

Atlas migrate required. Real query after rewrite:

```sql
WHERE project_issues = $1 AND deleted_at IS NULL
-- (+ filters / ORDER BY rank|issue_key / LIMIT)
```

### A8-002 — `test_cases`

```go
// ent/schema/testcase.go Indexes():
index.Edges("project").Fields("deleted_at"),
```

Real query after rewrite:

```sql
WHERE project_test_cases = $1 AND deleted_at IS NULL
```

---

## Tests run

```
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit                          → PASS
npx vitest run                            → 59 files / 147 tests PASS (~12s)
  incl. board-column-virtual, channel-list-error, shortcuts-dialog, context-help-tip

cd backend
go build ./internal/tracker/ ./internal/tms/ → PASS
```

A13-001 board-cap test already on branch: `backend/testframework/tracker/repository_board_cap_a13001_test.go`.

---

## Out of scope / deferred

| ID | Why |
|----|-----|
| A13-004 Gantt N+1 | Parallel agent (links batch) — not this handoff |
| A13-002/003/005… BE hot paths | Other be waves |
| A13-024 issues table window | P3 |
| Full i18next-http-backend CDN | Avoided new dep; dynamic import sufficient |
