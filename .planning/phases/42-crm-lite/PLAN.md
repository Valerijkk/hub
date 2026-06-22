# Phase 42 — CRM-lite (MOD-05)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** MOD-05 — Kanban deals; link to project; no invoicing

## Goal

CRM-lite для IT-команды: контакты, pipeline сделок, привязка к проектам Hub.

## Deliverables

### Ent
- `CrmContact` — name, email, phone, company, optional `linked_user`
- `CrmDeal` — title, stage enum, amount_cents, contact, project, owner

### Backend (`internal/crm/`)
- Contacts CRUD + soft delete
- Pipeline GET + deals CRUD
- Project link validates membership

### Frontend
- `/crm` — contacts sidebar + kanban pipeline (drag-drop stage move)
- Sidebar + Ctrl+K

## Verification

```bash
cd backend && go test ./internal/crm/...
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

## Next

Phase 43 — HR-lite (MOD-06) or Kudos (MOD-03)
