# Phase 102 — Dev Routes + Notification Push

**Branch:** `feat/v2.5-trust`  
**Depends on:** Phase 101 (sub-75% package push)  
**Goal:** Close deferred gaps: `internal/dev` (0%) and `notification` (74.8% → ≥75%)

## Scope

| Package | Before | Target | Wave |
|---------|--------|--------|------|
| `internal/dev/` | 0% | ≥75% | apidocs + handler + seed |
| `internal/notification/` | 74.8% | ≥75% | consumer edges + telegram handlers |

## Tasks

### Wave 102a — dev/apidocs (unit)

- [x] `methodOrder`, `methodColor`, `guessModule`
- [x] `collectRoutes`, `BuildAPIDocsHTML`, `BuildAPIDocsJSON`

### Wave 102b — dev/handler (`//go:build !release`)

- [x] DevMode guard → 404 when disabled
- [x] Status, Login validation/success, Seed idempotent

### Wave 102c — notification consumer

- [x] `handleAssigned` invalid JSON/UUID + debounce
- [x] `handleStatusChanged` no assignee, self-move, invalid JSON
- [x] `handleCommentAdded` invalid JSON, author-is-assignee skip
- [x] `handleNewLogin` invalid JSON/UUID, empty device, debounce

### Wave 102d — notification telegram

- [x] `NewTelegramBindService` misconfiguration nil paths
- [x] `handleUnlinkCommand` not linked, `handleStartBind` ID taken
- [x] Handler unauthorized, linked test success, confirm invalid body

## Deferred (Phase 103+)

- `messenger/mongo_repository`, `search/searchMessages` (Mongo — needs `MONGODB_TEST_URL`)
- Commit + PR (on user request)

## Success criteria

- [x] `dev` ≥75%
- [x] `notification` ≥75%
- [x] overall COV ≥76.3% (maintain/improve)
- [x] race PASS on touched packages
