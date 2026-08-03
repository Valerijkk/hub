# PW1-A6 — Inventory recount (measured 2026-07-27)

> Documentation-only measurement · **no** `go test` · PowerShell counts on workspace `c:\Users\Valerii\Desktop\hub`

## Measured NOW (2026-07-27)

| Metric | Measured | Method |
|--------|----------|--------|
| Frontend route files | **72** `.tsx` | `Get-ChildItem frontend\src\routes -Recurse -Filter *.tsx` (+ 1 `.gitkeep`, not a route) |
| of which `*.lazy.tsx` | **10** | TanStack lazy companions (same path as eager twin) |
| Eager route modules | **62** | total `.tsx` − `*.lazy.tsx` |
| Backend test files | **515** | `backend\testframework\**\*_test.go` |
| Ent schema `.go` files | **83** | `backend\ent\schema\*.go` |
| Ent entity schemas | **82** | exclude `helpers.go` (shared regexp/UTC helpers, not an entity) |
| Zustand stores | **13** | `frontend\src\stores\*.ts` each `import { create } from "zustand"` + one `create<…>` export |
| Go version | **1.25.0** | first `go` directive in `backend/go.mod` |

### TanStack note

Hub uses **file-based TanStack Router**. Canonical inventory for “how many route files” = **72** `.tsx` under `frontend/src/routes/`. Historical “39 routes” understated file count; unique URL modules without counting lazy twins ≈ **62**.

### Zustand store list (13)

`ai-store`, `auth-store`, `board-store`, `call-store`, `command-store`, `dashboard-store`, `messenger-store`, `notification-store`, `sidebar-store`, `sprint-store`, `timer-store`, `tms-store`, `ws-store`.

## Deltas vs prior claims

| Claim (source examples) | Was | Now (2026-07-27) | Δ |
|-------------------------|-----|------------------|---|
| TanStack routes | **39** (rules/README “39+”) | **72** route `.tsx` | **+33** files |
| Backend `*_test.go` | **495** (README) / **499** (AGENTS/COVERAGE 2026-07-26) | **515** | **+20** vs 495 · **+16** vs 499 |
| Go | **1.24** (`.cursor/rules/00-project.mdc`) | **1.25.0** (`go.mod`; `docs/stack_and_tools.md` already 1.25.0) | version bump claim |
| Zustand stores | **11** (rules) | **13** | **+2** |
| Ent schemas | **82** | **82** (+ `helpers.go`) | **0** (schemas OK) |

## Docs updated in companion commit (A6)

Canonical live claims corrected to measured numbers with note **«measured 2026-07-27»**. Historical audit/smoke reports under `.planning/audits/` and `pro-qa-ai/evals/` left as dated snapshots (not rewritten wholesale).

## Re-measure (PowerShell)

```powershell
(Get-ChildItem frontend\src\routes -Recurse -Filter *.tsx).Count
(Get-ChildItem backend\testframework -Recurse -Filter *_test.go).Count
(Get-ChildItem backend\ent\schema -Filter *.go | Where-Object Name -ne 'helpers.go').Count
(Get-ChildItem frontend\src\stores -Filter *.ts).Count
Select-String -Path backend\go.mod -Pattern '^go\s'
```
