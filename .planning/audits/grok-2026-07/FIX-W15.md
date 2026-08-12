# FIX-W15 — Perf: indexes / Kanban / lazy / bundle / waterfalls

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** fe2 (+ non-schema BE query rewrite)  
**Правила:** no Ent schema/index churn without approval · no push master

Сфокус: **A8-001/002** (indexes under real queries), **A13-014** Kanban virtualization, **A13-015/016/017** lazy shell/i18n/routes + bundle, **A13-011/012** startup waterfalls, **A13-018** channel-list virtualization. Amplifying **A13-001** board cap already on branch.

---

## Итог

| ID | Status | SHA | Notes |
|----|--------|-----|-------|
| A8-001 | FIXED | `6031985` | Query rewrite `0ce8cd7` + schema `index.Edges("project").Fields("deleted_at")` on Issue. Ent generate local (gitignored); `Schema.Create` applies at runtime. |
| A8-002 | FIXED | `6031985` | Query rewrite `443a569` + schema `index.Edges("project").Fields("deleted_at")` on TestCase. Same generate / migrate path. |
| A13-001 | FIXED (prior) | `456636f` | `MaxBoardIssuesPerColumn=200` — non-schema amp for A8-001. |
| A13-011 | FIXED | `2ae0f1f` | Board: `effectiveBoardId` from list → no effect waterfall. |
| A13-012 | FIXED | `2ae0f1f` | Dashboard: `Promise.all([fetchPersonal(), loadLayout])`. |
| A13-014 | FIXED | `015cc4c` + WEAK `3536e0e` | Kanban virtualize ≥24 cards + Vitest; **WEAK close 2026-07-28:** threshold assertion in `board-column-virtual.test.tsx`. |
| A13-015 | FIXED | `7510125` | Lazy AI/Command/Shortcuts in AppShell. |
| A13-016 | FIXED | `1de2e13` | Eager common+help; other ns dynamic import. |
| A13-017 | FIXED | `94760b7` (+ board lazy in `2ae0f1f`) | crm/accounting/marketplace/gantt/board lazy routes. |
| A13-018 | FIXED | `cf4dbd5` + WEAK `53d72c8` | Channel list virtualize ≥40 rows; **WEAK close 2026-07-28:** `channel-list-virtual.test.tsx` threshold. |
| Bundle | FIXED | via `7510125`…`94760b7` | Cold graph drops shell panels + heavy modules + most locale JSON. |
| Report | docs | `4621d9c` | This file + STATE session note. |

---

## Schema indexes — FIXED (2026-07-28, user-approved)

> A8-001 / A8-002 indexes **FIXED**. Generated Ent under `backend/ent/*` is gitignored; only `ent/schema` is committed. Runtime migrate via `client.Schema.Create` (no Atlas SQL migrations in repo).

### A8-001 — `issues` — FIXED

```go
// ent/schema/issue.go Indexes():
index.Edges("project").Fields("deleted_at"),
```

Generated index name: `issue_deleted_at_project_issues` (Ent: Fields then Edges columns).

```sql
WHERE project_issues = $1 AND deleted_at IS NULL
-- (+ filters / ORDER BY rank|issue_key / LIMIT)
```

### A8-002 — `test_cases` — FIXED

```go
// ent/schema/testcase.go Indexes():
index.Edges("project").Fields("deleted_at"),
```

Generated index name: `testcase_deleted_at_project_test_cases`.

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
