# Phase 47 — Random Coffee (V4-05)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** V4-05 — Random Coffee / kudos matching v4 intranet

## Goal

Opt-in участие → admin запускает matching → пары коллег → in-app уведомления.

## Deliverables

### Ent
- `CoffeeEnrollment`, `CoffeeRound`, `CoffeePair`

### Backend (`internal/intranet/` coffee_*)
- Enrollment, match query, admin run matching
- Shuffle + anti-repeat vs previous round
- `intranetNotifier.NotifyCoffeeMatch`

### Frontend
- Intranet tab «Coffee» — opt-in, partner card, admin run

## Verification

```bash
cd backend && go test ./internal/intranet/...
cd frontend && npx vitest run src/modules/intranet/random-coffee.test.tsx
```

## Next

Phase 48+ — MOD-07 1C connector (P3) · Lark Base (V4-04) · commit v4 block
