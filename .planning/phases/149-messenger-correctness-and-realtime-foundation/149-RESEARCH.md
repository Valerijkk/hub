# Phase 149 research: current Messenger correctness seams

**Date:** 2026-08-24
**Method:** static code inspection by a delegated explorer
**Runtime/tests:** NOT_RUN

## Current paths

| Concern | Current path / symbol |
|---|---|
| HTTP send | `backend/internal/messenger/handler_messages.go` - `(*Handler).SendMessage` |
| DTOs | `backend/internal/messenger/dto.go` - `SendMessageRequest`, `MessageResponse`, `MessageDocument` |
| Service | `backend/internal/messenger/service_messages.go` - `(*MessengerService).SendMessage` |
| Repository contract | `backend/internal/messenger/repository.go` - `MessageRepository` |
| Mongo persistence/indexes | `backend/internal/messenger/mongo_repository.go` - `InsertMessage`, `FindMessageByID`, `EnsureIndexes` |
| Channel membership | `backend/internal/messenger/ent_repository.go` - `(*EntRepository).IsMember` |
| WS authorization | `backend/internal/ws/authorizer.go` - `CanSubscribe` |
| WS relay | `backend/internal/ws/bridge.go` - `handleMessengerEvent` |
| Event models/producers | `backend/internal/messenger/events.go` and Messenger services |
| FE API | `frontend/src/api/messenger.ts` |
| FE send owner | `frontend/src/components/messenger/message-composer.tsx` |
| FE message subscriptions | `frontend/src/hooks/use-messages.ts` |
| Socket lifecycle | `frontend/src/stores/ws-store.ts` |

## Verified static gaps

1. `SendMessage` copies `ThreadID` and increments the referenced parent without proving that the parent exists, belongs to the URL channel, or is a root rather than a nested reply.
2. Messages are Mongo documents; `client_nonce` therefore needs a Mongo unique partial index and an atomic create-or-return repository operation, not an Ent migration.
3. Unread mutation, parent reply count, NATS publication, mentions, and hooks currently follow every successful insert. They must run only for the `created=true` branch of an idempotent operation.
4. `MessageSentEvent` is a partial message view. The current frontend cache either needs a canonical payload or must invalidate/refetch on receipt.
5. `ReactionToggledEvent` payload and the current frontend cache handler disagree: the backend publishes a delta, while the hook expects a complete reactions array.
6. Websocket reconnection and resubscription exist, but no reconnect-driven invalidation/refetch recovers events missed while disconnected. NATS websocket delivery is not replayed.
7. Existing direct channel subject authorization is membership-aware and fails closed for malformed subjects. Phase 149 must extend the same rigor to the canonical per-user subject rather than inventing an unauthorizable subject spelling.

## Frozen implementation direction

- Parent validation loads the referenced message before insert, checks same channel and usable root, and returns a non-disclosing not-found for a foreign parent.
- Repository gains an atomic idempotent creation result with `saved`, `created`, and `error` semantics.
- The unique partial index is scoped by authenticated sender and `client_nonce`; a reused nonce with incompatible channel/content returns conflict.
- Startup/readiness must fail closed if the mandatory uniqueness index cannot be established.
- The composer generates a nonce once per logical submit and retains it across retry for text, GIF, sticker, and voice sends.
- Reconnect triggers deterministic refetch of active channel and thread queries.

## Verification routes

- `docker compose -f docker-compose.local.yml run --rm backend-unit`
- `docker compose -f docker-compose.local.yml run --rm frontend-unit`
- Mongo index/upsert integration evidence belongs to `python-integration`; until the Phase-owned case exists and runs, this remains NOT_RUN.
