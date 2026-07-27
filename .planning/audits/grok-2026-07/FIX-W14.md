# FIX-W14 — Block D perf (A13-001…006)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Агент:** be1  
**Правила:** 1 commit / finding · measure before/after where feasible · no schema/indexes (→ W15) · no push master

Сфокус: unbounded board/bases/dashboard/sprint, Gantt N+1, notification fan-out.

---

## Итог

| ID | Status | SHA | Notes |
|----|--------|-----|-------|
| A13-001 | FIXED | `456636f` | `MaxBoardIssuesPerColumn=200` on `GetBoardData` `WithIssues` |
| A13-002 | FIXED | `f961fec` | `GetTableMeta` + `GetRecord` write-path; `MaxTableRecords=500` on `GetTable` |
| A13-003 | FIXED | `60149a0` | `AggregateProjectIssueStats` (Count) replaces `.All` in `GetProjectStats` |
| A13-004 | FIXED | `e76d5ef` | `GET /projects/:id/issue-links` batch + Gantt parent preload (no per-row N+1) |
| A13-005 | FIXED | `6d15338` | `ResolveChannelPrefs` + single `GetUser` in `Dispatch` |
| A13-006 | FIXED | `3b224e4` | `ListSprints` without `WithIssues`; `MaxSprintIssues=500` on GetSprint |

---

## Measure notes

| ID | Before | After |
|----|--------|-------|
| A13-001 | unbounded issues/column | ≤200/column (test creates 205 → loads 200) |
| A13-002 | full table on every mutation | meta-only write; view ≤500 records |
| A13-003 | `.All` all project issues on cache miss | O(statuses)+3 Count queries |
| A13-004 | N HTTP `listIssueLinks` per Gantt row | 1× `listProjectIssueLinks` (cap 2000) |
| A13-005 | ≤3×`IsEnabled` (≤6 SQL) + 1–2 `GetUser` per recipient | 1 prefs query + 1 `GetUser` |
| A13-006 | `ListSprints` + full issue graph | list edges.Issues=nil; Count for `issue_count` |

---

## Tests run

```
go test -tags hubtest -race -count=1 ./testframework/tracker/ -run A13_001|A13_004
go test -tags hubtest -race -count=1 ./testframework/bases/ -run A13_002
go test -tags hubtest -race -count=1 ./testframework/dashboard/ -run A13_003
go test -tags hubtest -race -count=1 ./testframework/notification/ -run A13_005|Dispatch_
go test -tags hubtest -race -count=1 ./testframework/sprint/ -run A13_006|ListSprints
npx vitest run …/tracker/gantt-batch-links-a13004.test.ts
→ PASS
```

## Docs

- `docs/wiki/API.md` — `GET /projects/:projectId/issue-links` (A13-004)

## Out of scope (W15 / later)

- Indexes A8-001/002 (project FK) — schema approval required
- A13-007+ P2/P3 perf items
