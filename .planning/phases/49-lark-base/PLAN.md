# Phase 49 — V4-04 Lark Base Clone (MVP)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust`  
**Requirement:** V4-04 — Lark-style no-code database

## Goal

Org-wide no-code tables: workspaces (bases), typed columns, grid records. MVP slice — no kanban views, automations, or formulas.

## Delivered

### Ent

- `HubBase`, `HubTable`, `HubColumn`, `HubRecord` schemas
- Column types: `text`, `number`, `date`, `select`, `checkbox`, `url`, `user_ref`
- Cells stored as JSON map keyed by column UUID

### Backend

- `backend/internal/bases/` — repository, service, handler, tests
- Routes under `/api/bases/*`
- New base auto-creates table «Main» + column «Name»
- Optimistic locking via `version` on base/table/column/record

### Frontend

- `/bases` — list bases, tab tables, spreadsheet grid, add columns/rows
- Sidebar + Ctrl+K navigation

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/bases` | List workspaces |
| POST | `/bases` | Create base (+ default table) |
| GET | `/bases/:baseId` | Base detail + tables |
| PUT | `/bases/:baseId` | Update base |
| DELETE | `/bases/:baseId` | Soft delete |
| POST | `/bases/:baseId/tables` | New table |
| GET | `/bases/:baseId/tables/:tableId` | Columns + records |
| POST | `/bases/:baseId/tables/:tableId/columns` | Add column |
| POST | `/bases/:baseId/tables/:tableId/records` | Add row |
| PUT | `/bases/:baseId/tables/:tableId/records/:recordId` | Update cells |
| DELETE | `/bases/:baseId/tables/:tableId/records/:recordId` | Soft delete |

## Verification

```bash
cd backend && go build ./... && go test ./internal/bases/...
cd frontend && npx tsc --noEmit && npx vite build
```

## Out of scope (future)

- Kanban/gallery views, linked records, formulas, RBAC per base, import CSV

## Next

Phase 50 — v4 Milestone Gate · commit phases 40–49 · PLAN-CLOSE-GAPS coverage wave
