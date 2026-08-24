# Messenger Correctness and Realtime Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` to implement this plan task-by-task. Each task uses a fresh Terra/XHigh atomic writer and an independent Terra/High reviewer.

**Goal:** Make Messenger thread replies, retries, realtime subjects, and reconnect recovery converge to one authorized durable truth.

**Architecture:** Keep Hub Messenger as the only message/channel system of record. Enforce thread and idempotency invariants in the Go service plus Mongo repository, then migrate the React client to stable nonces and persistence-backed refetch. Realtime remains an acceleration path: channel content uses `channel:<id>`, per-user membership/sidebar events use `user:<id>`, and reconnect always reconciles from HTTP persistence.

**Tech Stack:** Go 1.25, Fiber, MongoDB Go driver v2, NATS, Hub websocket bridge, React 19, TypeScript, TanStack Query, Zustand, Vitest/RTL, Docker Compose.

**Spec:** `.planning/phases/149-messenger-correctness-and-realtime-foundation/149-SPEC.md`

## Global Constraints

- Binding decisions: `.planning/decisions/ADR-COMMS-OSS-REPLACEMENT-2026-08-24.md` and `149-CONTEXT.md`.
- Execute only in `calls-feature-codex`; this is an already isolated linked worktree.
- Tests/build/typecheck/lint run only with `docker compose -f docker-compose.local.yml run --rm backend-unit` or `frontend-unit`. Never run Go, Node, Vitest, or validators on the Windows host.
- Genuine units live only under `backend/tests/**` and `frontend/tests/**`. Do not revive legacy backend component tests, static-contract Vitest, or Playwright suites.
- Do not read or edit `pro-qa-ai/**`, Mail/Calendar paths, or `frontend/src/modules/calls/create-call-dialog.tsx`.
- Persist before publish. A duplicate logical send must not increment unread/reply counters, resolve mentions, publish NATS, or invoke hooks twice.
- Phase 149 guarantees duplicate suppression, not recovery of a failed best-effort Redis unread projection. That existing cross-storage gap is explicit `NOT_READY` work for Phase 151; realtime remains recoverable from HTTP persistence.
- Server derives sender/user identity from authenticated context. Cross-channel parent rejection must not disclose the foreign message.
- New runtime behavior requires same-session docs sync. Runtime status must be `PASS`, `FAIL`, `NOT_RUN`, or `BLOCKED`; static inspection is not runtime PASS.
- Use exact-path staging and no AI attribution trailers.

## File and Interface Map

| File | Responsibility in this phase |
|---|---|
| `backend/internal/messenger/dto.go` | Add bounded `client_nonce` to request, document, and response. |
| `backend/internal/messenger/repository.go` | Define atomic idempotent create contract. |
| `backend/internal/messenger/mongo_repository.go` | Unique partial nonce index and atomic create-or-return. |
| `backend/internal/messenger/service_messages.go` | Validate thread root and gate every side effect on `created`. |
| `backend/internal/messenger/events.go` | Canonical message and per-user event payloads/subjects. |
| `backend/internal/messenger/service_bot.go`, `service_polls.go`, `service_promote.go` | Migrate every existing `MessageSentEvent` producer to the canonical payload. |
| `backend/internal/messenger/service_channels.go` | Produce membership/channel events to affected users. |
| `backend/internal/ws/authorizer.go` | Authorize exact own `user:<id>` and accessible `channel:<id>`. |
| `backend/internal/ws/bridge.go` | Relay `messenger.user.>` to `user:<id>` in addition to channel events. |
| `backend/tests/messenger/mock_repository_test.go` | Implement the changed repository contract and observable side-effect counters. |
| `backend/tests/messenger/service_messages_test.go` | Thread and nonce service units. |
| `backend/tests/ws/authorizer_test.go` | Own-user versus foreign-user subscription units. |
| `frontend/src/api/messenger.ts` | Carry nonce and canonical message event fields. |
| `frontend/src/components/messenger/message-composer.tsx` | Own one nonce per logical text/GIF/sticker/voice submission. |
| `frontend/src/hooks/use-messages.ts` | Apply compatible deltas and refetch active queries after reconnect. |
| `frontend/src/hooks/use-channels.ts` | Subscribe to `user:<id>` and reconcile channel list after reconnect. |
| `frontend/tests/messenger/*` | Genuine nonce, cache event, reconnect, and user-subject units. |
| `docs/TECHNICAL.md`, `docs/TESTING.md` | Document the message/idempotency/subject contract and test evidence. |

## Preflight Interaction Table

| Producer task | Consumer task | Shared file/interface | Ruling |
|---|---|---|---|
| Task 1 | Task 2 | `MessengerService.SendMessage`, `FindMessageByID` | Parent validation stays before idempotent creation; Task 2 must preserve the same errors. |
| Task 2 | Task 3 | `client_nonce` in Go DTO/response and TS API types | JSON name is exactly `client_nonce`; optional only for transition, required for all new composer sends. |
| Task 2 | Task 4 | `MessageSentEvent` / canonical response | Only `created=true` publishes; Task 4 may expand payload but never republish a retry. |
| Task 4 | Task 5 | channel versus user subject routing | Message content remains channel-scoped; membership/sidebar changes are user-scoped. |
| Every task | Final gate | Docker services and docs evidence | No task claims full Phase PASS before both unit services and remaining integration boundary are classified. |

### Task 1: Enforce same-channel live root threads

**Files:**

- Create: `backend/tests/messenger/service_messages_test.go`
- Modify: `backend/internal/messenger/service_messages.go`
- Modify: `backend/tests/messenger/mock_repository_test.go`

**Interfaces:**

- Consumes: `MessageRepository.FindMessageByID(ctx, messageID)` and `MessageDocument.{ChannelID,ThreadID,DeletedAt}`.
- Produces: private `validateThreadParent(ctx context.Context, channelID uuid.UUID, threadID string) error`; `SendMessage` calls it after channel authorization and before constructing/inserting the reply.

- [ ] **Step 1: Write failing service units**

Add tests using existing `newMockRepository`, `setupChannelWithMembers`, and `messenger.ExportNewMessengerServiceTest`:

```go
func TestSendMessageRejectsForeignThreadParent(t *testing.T)
func TestSendMessageRejectsMissingThreadParent(t *testing.T)
func TestSendMessageRejectsDeletedThreadParent(t *testing.T)
func TestSendMessageRejectsNestedThreadParent(t *testing.T)
func TestSendMessageAcceptsLiveRootInSameChannel(t *testing.T)
```

Each rejection asserts `require.Error`, zero new documents, and unchanged parent `ReplyCount`. The foreign parent and missing parent must expose the same not-found class. The nested-parent case asserts invalid-input. The valid case asserts one reply with the requested root ID and `ReplyCount == 1`.

- [ ] **Step 2: Run the Docker red gate**

Run:

```powershell
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: FAIL in the new `TestSendMessageRejects*` cases because the current service inserts before verifying the parent.

- [ ] **Step 3: Implement the minimal invariant**

Add and call:

```go
func (s *MessengerService) validateThreadParent(
    ctx context.Context,
    channelID uuid.UUID,
    threadID string,
) error
```

Behavior is exact:

1. Empty `threadID` returns `nil`.
2. Load with `FindMessageByID`.
3. Repository error or `parent.ChannelID != channelID` returns `apperrors.NewNotFound("сообщение не найдено")`.
4. `parent.DeletedAt != nil` returns the same not-found class. The production Mongo lookup already filters deleted documents; this explicit check keeps mocks and future repositories honest.
5. `parent.ThreadID != ""` returns `apperrors.NewBadRequest("ответ должен ссылаться на корневое сообщение треда")` using the repository's existing invalid-input constructor name if it differs.
6. Validation completes before `InsertMessage`; reply increment stays after successful creation.

- [ ] **Step 4: Run new behavior and regression to green**

Run `backend-unit` again. Expected: all backend build/vet/unit/race stages PASS. If it fails, perform `/test-triage` against `149-SPEC.md` before changing code or tests.

- [ ] **Step 5: Commit exact paths**

```powershell
git add backend/internal/messenger/service_messages.go backend/tests/messenger/service_messages_test.go backend/tests/messenger/mock_repository_test.go
git commit -m "fix(messenger): bind replies to channel thread roots"
```

### Task 2: Make message creation storage-idempotent

**Files:**

- Modify: `backend/internal/messenger/dto.go`
- Modify: `backend/internal/messenger/repository.go`
- Modify: `backend/internal/messenger/mongo_repository.go`
- Modify: `backend/internal/messenger/service_messages.go`
- Modify: `backend/cmd/server/wiring_messenger.go`
- Modify: `backend/tests/messenger/mock_repository_test.go`
- Modify: `backend/tests/messenger/service_messages_test.go`

**Interfaces:**

- Produces:

```go
type CreateMessageResult struct {
    Message *MessageDocument
    Created bool
}

CreateMessageIdempotent(
    ctx context.Context,
    doc *MessageDocument,
) (*CreateMessageResult, error)

FindMessageByClientNonce(
    ctx context.Context,
    senderID uuid.UUID,
    clientNonce string,
) (*MessageDocument, error)
```

- DTO/document field is exactly `ClientNonce string` with `json:"client_nonce,omitempty"`, `bson:"client_nonce,omitempty"`, and request validation `omitempty,min=8,max=128`.
- Unique index name is exactly `messages_sender_client_nonce_unique` on `{sender_id: 1, client_nonce: 1}` with `unique=true` and partial filter `{$and: [{client_nonce: {$type: "string"}}, {client_nonce: {$gt: ""}}]}`.

- [ ] **Step 1: Extend the mock and write failing nonce units**

Add tests:

```go
func TestSendMessageRetryReturnsCanonicalMessageWithoutSideEffects(t *testing.T)
func TestSendMessageRetryAfterParentDeletionReturnsCanonical(t *testing.T)
func TestSendMessageConcurrentRetryCreatesOneMessage(t *testing.T)
func TestSendMessageNonceReuseWithDifferentChannelConflicts(t *testing.T)
func TestSendMessageWithoutNonceUsesTransitionalCreate(t *testing.T)
```

Extend `mockRepository` with a mutex, a `senderID + "\x00" + clientNonce` lookup, `createMessageCalls`, and the exact `CreateMessageIdempotent` method. Tests assert one stored ID and one reply increment for identical retries; a retry still returns the stored reply after its parent is later deleted; incompatible reuse returns conflict; legacy empty nonce still creates a message but is never considered idempotent.

- [ ] **Step 2: Run the Docker red gate**

Run `backend-unit`. Expected: compile/test FAIL because the repository contract and DTO do not yet expose the idempotent operation.

- [ ] **Step 3: Implement atomic Mongo create-or-return**

Implement `CreateMessageIdempotent` as follows:

- assign `doc.ID` and `CreatedAt` before persistence;
- empty nonce uses `InsertOne` and returns `Created=true`;
- non-empty nonce uses `UpdateOne(filter, {$setOnInsert: doc}, options.UpdateOne().SetUpsert(true))` with filter `{sender_id, client_nonce}`;
- `Created` is exactly `result.UpsertedCount == 1`; after the update, read the canonical document by the same key and return it;
- when a concurrent upsert reports `mongo.IsDuplicateKeyError`, read the canonical document and return `Created=false`; no other error is converted to a retry;
- compare stored `ChannelID`, `ThreadID`, `ContentJSON`, `ContentText`, attachments, voice fields, and mentions to the retry request; mismatch returns a typed conflict sentinel mapped by the service to `apperrors.NewConflict("client_nonce уже использован для другого сообщения")`;
- never use an unlocked find-then-insert as the creation decision;
- update `EnsureIndexes` with the named unique partial index;
- change `wiring_messenger.go` to fail startup instead of logging-and-continuing when this mandatory index cannot be established.

For a non-empty nonce, first call `FindMessageByClientNonce(senderID, nonce)`. An existing compatible document returns before thread-parent validation, so a retry remains valid after later parent deletion; an incompatible fingerprint conflicts. If no document exists, validate the parent and call `CreateMessageIdempotent`; its unique upsert resolves the concurrent race. When `Created=false`, immediately return `toMessageResponse(result.Message)`. Only `Created=true` may increment reply/unread, resolve mentions, publish NATS, or emit hooks. Existing best-effort projection failure semantics are documented rather than mislabelled as recovered in this phase.

- [ ] **Step 4: Run backend regression to green**

Run `backend-unit`. Expected: PASS. Record Mongo driver integration as `NOT_RUN`; do not claim the real unique-index/upsert path passed until a Phase-owned `python-integration` case exists.

- [ ] **Step 5: Commit exact paths**

```powershell
git add backend/internal/messenger/dto.go backend/internal/messenger/repository.go backend/internal/messenger/mongo_repository.go backend/internal/messenger/service_messages.go backend/cmd/server/wiring_messenger.go backend/tests/messenger/mock_repository_test.go backend/tests/messenger/service_messages_test.go
git commit -m "feat(messenger): make sends idempotent by client nonce"
```

### Task 3: Give every logical composer submit a stable nonce

**Files:**

- Modify: `frontend/src/api/messenger.ts`
- Modify: `frontend/src/components/messenger/message-composer.tsx`
- Create: `frontend/tests/messenger/message-composer-retry-nonce.test.tsx`

**Interfaces:**

- Consumes/produces: `SendMessageRequest.client_nonce?: string` and `Message.client_nonce?: string`.
- Logical submit state retains one `crypto.randomUUID()` through retry; a later distinct submit gets a new UUID.

- [ ] **Step 1: Write failing text/GIF/sticker/voice nonce tests**

Mock `sendMessage` and `crypto.randomUUID`. Cover all four send routes. Force the first request to reject, invoke retry through the existing error/retry UI path, and assert both calls carry the same nonce. Then complete a new submit and assert a different nonce.

- [ ] **Step 2: Run the Docker red gate**

Run `frontend-unit`. Expected: FAIL because current requests contain no `client_nonce`.

- [ ] **Step 3: Implement one nonce per logical submission**

Add the optional TS fields and a composer-owned pending-send object containing the payload plus nonce. Generate the nonce before the first API call, reuse the same object for retry, and clear it only after success or explicit discard. Do not generate the nonce inside the API client.

- [ ] **Step 4: Run frontend regression to green**

Run `frontend-unit`. Expected: TypeScript and Vitest PASS. Triage any failure against the spec before editing.

- [ ] **Step 5: Commit exact paths**

```powershell
git add frontend/src/api/messenger.ts frontend/src/components/messenger/message-composer.tsx frontend/tests/messenger/message-composer-retry-nonce.test.tsx
git commit -m "feat(messenger): preserve send nonce across retries"
```

### Task 4: Reconcile message caches from canonical events and reconnect

**Files:**

- Modify: `backend/internal/messenger/events.go`
- Modify: `backend/internal/messenger/service_messages.go`
- Modify: `backend/internal/messenger/service_reactions.go`
- Modify: `backend/internal/messenger/service_bot.go`
- Modify: `backend/internal/messenger/service_polls.go`
- Modify: `backend/internal/messenger/service_promote.go`
- Create: `backend/tests/messenger/message_sent_producers_test.go`
- Modify: `frontend/src/api/messenger.ts`
- Modify: `frontend/src/hooks/use-messages.ts`
- Modify: `frontend/tests/messenger/use-messages-cache-events.test.tsx`
- Modify: `frontend/tests/messenger/reaction-toggled-guard.test.tsx`

**Interfaces:**

- `MessageSentEvent` carries `Message MessageResponse` as `json:"message"` plus `MentionedUserIDs`; consumers stop reconstructing a partial message from unrelated top-level fields.
- `ReactionToggledEvent` remains a delta `{message_id,channel_id,user_id,emoji,added}`; the frontend reducer applies that delta idempotently.

- [ ] **Step 1: Write failing canonical-event and reconnect tests**

Assert normal send, forward, bot send, poll-message creation, and promote-to-channel each emit the same canonical nested `message` payload. Assert a `message_sent` event inserts the full attachments/voice/content/reactions/nonce response into the correct channel or thread query exactly once. Assert duplicate reaction delta is harmless. Simulate `connected: false -> true` and assert active channel and thread queries are invalidated after resubscription.

- [ ] **Step 2: Run both Docker red gates**

Run `backend-unit`, then `frontend-unit`. Expected: backend/FE tests fail on the old partial event and missing reconnect invalidation.

- [ ] **Step 3: Publish canonical response and make recovery persistence-backed**

Add one constructor `newMessageSentEvent(doc, mentionedUserIDs)` and use it from normal send, forward, bot, poll, and promote producers. Build `response := toMessageResponse(doc)` once and publish it inside `MessageSentEvent.Message`; normal HTTP send returns the same response. Update `buildMessageFromWsEvent` and hooks to consume the nested response. Apply reaction deltas by adding/removing `user_id` under the matching emoji without requiring a complete reactions array. In both `useMessages` and `useThreadMessages`, detect a reconnection generation and invalidate the active query after subscription; initial connection must not cause a redundant recovery fetch.

- [ ] **Step 4: Run backend and frontend regression to green**

Run `backend-unit` then `frontend-unit`. Expected: both PASS.

- [ ] **Step 5: Commit exact paths**

```powershell
git add backend/internal/messenger/events.go backend/internal/messenger/service_messages.go backend/internal/messenger/service_reactions.go backend/internal/messenger/service_bot.go backend/internal/messenger/service_polls.go backend/internal/messenger/service_promote.go backend/tests/messenger/message_sent_producers_test.go frontend/src/api/messenger.ts frontend/src/hooks/use-messages.ts frontend/tests/messenger/use-messages-cache-events.test.tsx frontend/tests/messenger/reaction-toggled-guard.test.tsx
git commit -m "fix(messenger): reconcile realtime message caches"
```

### Task 5: Add authorized per-user channel lifecycle events

**Files:**

- Modify: `backend/internal/messenger/events.go`
- Modify: `backend/internal/messenger/service_channels.go`
- Modify: `backend/internal/ws/authorizer.go`
- Modify: `backend/internal/ws/bridge.go`
- Modify: `backend/cmd/server/wiring_ws.go`
- Modify: `backend/tests/ws/authorizer_test.go`
- Create: `backend/tests/messenger/user_channel_events_test.go`
- Modify: `frontend/src/hooks/use-channels.ts`
- Modify: `frontend/tests/messenger/use-channels-ws.test.tsx`
- Modify: `frontend/tests/messenger/use-channels-cache.test.ts`
- Modify: `docs/TECHNICAL.md`
- Modify: `docs/TESTING.md`
- Create: `.planning/phases/149-messenger-correctness-and-realtime-foundation/149-VERIFICATION.md`

**Interfaces:**

- NATS subjects are exactly `messenger.user.<user-id>.channel_created`, `.channel_updated`, `.channel_removed`, and `.new_message_notification`.
- Websocket bridge maps them to room `user:<user-id>` and event suffix after the UUID.
- `CanSubscribe` accepts `user:<uuid>` only when the UUID equals the authenticated user; channel authorization remains membership-backed.
- `new_message_notification` carries stable `{channel_id,message_id}` and causes persistence-backed channel-list invalidation; the client never performs duplicate-sensitive local unread arithmetic.

- [ ] **Step 1: Write failing backend and frontend subject tests**

Backend authorizer tests cover own user PASS, foreign user DENY, malformed UUID DENY, `user:<id>:channels` DENY, accessible channel PASS, inaccessible private channel DENY. Messenger tests assert create/invite/update/remove publish only to affected users with complete channel payload, stable `{channel_id}` removal payload, and stable `{channel_id,message_id}` notification payload. Frontend tests assert subscription is exactly `user:<id>`, duplicate notifications are harmless, and reconnection invalidates `CHANNELS_QUERY_KEY` before processing new deltas.

- [ ] **Step 2: Run both Docker red gates**

Run `backend-unit` then `frontend-unit`. Expected: FAIL on unsupported user subject, missing bridge subscription/producers, and old FE room spelling.

- [ ] **Step 3: Implement subject authorization, bridge, and producers**

Add user subject constants and typed payloads. Subscribe the bridge to `messenger.user.>` alongside `messenger.channel.>` and parse exactly four segments with UUID validation before broadcasting. Publish create/invite/update/remove events only after repository mutation; resolve affected member IDs from the repository. Publish new-message notification only for members other than the sender and include the durable message ID, while message content remains on the channel subject. Update `useChannels` to one `user:<id>` subscription, invalidate `CHANNELS_QUERY_KEY` on notification instead of incrementing unread locally, and perform the same persistence refetch after reconnect.

- [ ] **Step 4: Run full Phase 149 unit gates and classify integration**

Run:

```powershell
docker compose -f docker-compose.local.yml run --rm backend-unit
docker compose -f docker-compose.local.yml run --rm frontend-unit
```

Expected: PASS. Record `python-integration` as `NOT_RUN` unless this task also adds and runs a greenfield Mongo/NATS case under `testframework/tests/integration/**` through the existing `python-integration` service.

- [ ] **Step 5: Sync docs and commit exact paths**

Update the websocket subject table, idempotency/index contract, explicit Phase-151 unread recovery boundary, reconnect behavior, all Task 1-5 evidence receipts, new/legacy test status, and protected scope in `docs/TECHNICAL.md` and `docs/TESTING.md`. Create `149-VERIFICATION.md` with separate static, backend-unit, frontend-unit, and Mongo/NATS integration rows; unavailable executable evidence stays `NOT_RUN`/`PARTIAL`.

```powershell
git add backend/internal/messenger/events.go backend/internal/messenger/service_channels.go backend/internal/ws/authorizer.go backend/internal/ws/bridge.go backend/cmd/server/wiring_ws.go backend/tests/ws/authorizer_test.go backend/tests/messenger/user_channel_events_test.go frontend/src/hooks/use-channels.ts frontend/tests/messenger/use-channels-ws.test.tsx frontend/tests/messenger/use-channels-cache.test.ts docs/TECHNICAL.md docs/TESTING.md .planning/phases/149-messenger-correctness-and-realtime-foundation/149-VERIFICATION.md
git commit -m "feat(messenger): deliver authorized user channel events"
```

## Final Phase Gate

1. Run `backend-unit`, then `frontend-unit`, twice only if the first pass exposed a real flaky/infra classification; never loop blindly.
2. Search the changed frontend for `user:${userId}:channels` and the changed backend for direct `InsertMessage` use in user send paths; expected product hits: zero.
3. Independent review receives `149-SPEC.md`, exact BASE..HEAD diff, raw Docker outputs, and no generator conclusion.
4. Update `149-VERIFICATION.md` with separate static/runtime verdicts. If Mongo/NATS integration was not executed, the phase remains `PARTIAL` with `python-integration: NOT_RUN`; do not relabel it PASS.
5. Phase 151 retains the explicit `NOT_READY` durable read-state/projection rebuild gap; Phase 149 PASS never implies that cross-storage recovery exists.
