# Phase 45 — Internal marketplace (MOD-08)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** MOD-08 — Post listing; mod queue; optional #marketplace channel

## Goal

Внутренняя барахолка: объявления с модерацией admin/founder, публичная лента, уведомления mod queue и автору.

## Deliverables

### Ent
- `MarketplaceListing` — title, description, category, status, price_cents, review fields, version

### Backend (`internal/marketplace/`)
- CRUD + approve/reject/sold/withdraw
- Re-moderation on edit of approved listing
- `marketplaceNotifierAdapter` → admins on submit, author on decision
- `marketplaceChannelBroadcaster` → `#marketplace` channel via SendBotMessage (optional)

### Frontend
- `/marketplace` — browse/mine tabs, create form, admin mod queue
- Sidebar + Ctrl+K

## Verification

```bash
cd backend && go test ./internal/marketplace/...
cd frontend && npx tsc --noEmit && npx vitest run src/modules/marketplace/
```

## Next

Phase 46 — Public status page (V4-06) or MOD-07 1C connector (P3)
