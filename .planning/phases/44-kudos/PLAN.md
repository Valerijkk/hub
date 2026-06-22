# Phase 44 — Kudos / thanks (MOD-03)

**Status:** COMPLETE  
**Branch:** `feat/v2.5-trust` (uncommitted)  
**Requirement:** MOD-03 — Send kudos → recipient notif + optional public feed

## Goal

«Спасибо» коллеге: in-app уведомление получателю + публичная лента на intranet.

## Deliverables

### Ent
- `Kudo` — message (max 500), is_public, sender/recipient edges

### Backend (`internal/intranet/` extended)
- `POST /intranet/kudos` — send (no self-kudos)
- `GET /intranet/kudos/feed` — public feed
- `GET /intranet/kudos/received` — my received
- `intranetKudosNotifier` → NotificationService

### Frontend
- Intranet hub tabs: Новости | Спасибо
- `kudos-feed.tsx` — send form + feed/received tabs
- Deep link `/intranet?tab=kudos` from push

## Verification

```bash
cd backend && go test ./internal/intranet/...
cd frontend && npx tsc --noEmit && npx vitest run src/modules/intranet/
```

## Next

Phase 45 — MOD-08 Internal marketplace or MOD-07 1C connector (P3 backlog)
