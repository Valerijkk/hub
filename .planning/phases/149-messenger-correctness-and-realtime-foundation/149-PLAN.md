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
| Task 1 | Task 2C | `MessengerService.SendMessage`, `FindMessageByID` | After channel authorization, nonce lookup precedes parent validation; only a new nonce validates the parent before creation, preserving Task 1 errors. |
| Task 2A | Task 2B / Task 2C | persistence API, unique index, both repository mocks | The storage contract lands green before startup policy and service orchestration consume it. |
| Task 2C | Task 3 | `client_nonce` in Go DTO/response and TS API types | JSON name is exactly `client_nonce`; optional only for transition, required for all new composer sends. |
| Task 2C | Task 4 | `MessageSentEvent` / canonical response | Only `created=true` publishes; Task 4 may expand payload but never republish a retry. |
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
- Modify: `backend/internal/messenger/mock_repository.go`
- Create: `backend/internal/messenger/message_idempotency.go`
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

- Request, document, and response expose exactly `ClientNonce string`; request/response use `json:"client_nonce,omitempty"`, document uses `bson:"client_nonce,omitempty"`, request validation is `omitempty,min=8,max=128`, and `toMessageResponse` copies the persisted nonce.
- Unique index name is exactly `messages_sender_client_nonce_unique` on `{sender_id: 1, client_nonce: 1}` with `unique=true` and partial filter `{$and: [{client_nonce: {$type: "string"}}, {client_nonce: {$gt: ""}}]}`.

**Mandatory atomic execution split:**

- **Task 2A — persistence contract:** `dto.go`, `repository.go`, `mongo_repository.go`, both repository mocks, `backend/tests/messenger/message_idempotency_persistence_test.go`, `docs/TECHNICAL.md`, and `docs/TESTING.md`. It owns fields, mapper, repository API, Mongo upsert/read, named partial unique index, every mutex/counter/result-override seam later consumed by 2C, and incremental contract/evidence docs. Commit: `feat(messenger): add nonce persistence contract`.
- **Task 2B — startup policy:** `backend/cmd/server/wiring_messenger.go`, `docs/TECHNICAL.md`, and `docs/TESTING.md`; required-index failure becomes fail-startup through the existing startup fatal path and docs record build-only versus process `NOT_RUN` evidence. Commit: `fix(messenger): require messenger indexes at startup`.
- **Task 2C — service orchestration:** `backend/internal/messenger/message_idempotency.go`, `backend/internal/messenger/service_messages.go`, `backend/tests/messenger/service_messages_test.go`, `docs/TECHNICAL.md`, and `docs/TESTING.md`. It owns retry-before-parent behavior, fingerprint checks, conflict mapping, the live-canonical plus `Created` side-effect gate, and incremental behavior/evidence docs. It consumes test seams already installed by 2A and may not reopen either mock. Commit: `feat(messenger): suppress duplicate send side effects`.
- Each slice starts from the accepted HEAD of the previous slice and receives its own implementer report, BASE..HEAD review package, independent Terra/high review, and at most three fix cycles. A writer may not combine slices.

**Frozen persistence and fingerprint semantics:**

- `FindMessageByClientNonce` returns `(nil, nil)` only when no document exists; other storage errors propagate. Its query intentionally does not filter `deleted_at`, because the unique nonce remains reserved after soft deletion.
- If the stored canonical document has `DeletedAt != nil`, service returns conflict `client_nonce уже принадлежит удалённому сообщению`; it never recreates the message and never returns deleted content.
- Fingerprint includes exact `ChannelID`, `ThreadID`, `ContentText`, `VoiceURL`, and `VoiceDurationMs`.
- `ContentJSON` is compared through deterministic `encoding/json.Marshal`; map-key order and Mongo numeric concrete types that serialize to the same JSON number do not create false conflicts.
- `nil` and empty attachments, mentions, and mention-groups are equivalent. Attachment order remains significant because it is visible to users; attachment fields compare exactly. Mentions and mention-groups are semantic sets and compare after sort plus de-duplication.
- Repository code only creates or returns the canonical document. Compatibility and deleted-canonical decisions live in the service helper so pre-existing and concurrent retries follow one rule.

- [ ] **Step 1: Write the Task 2A persistence-contract red units**

Add compile/behavior coverage in `backend/tests/messenger/message_idempotency_persistence_test.go` for `ClientNonce` propagation, `(senderID, nonce)` mock lookup, empty-nonce transitional insert, and concurrent mock create-or-return. Both `backend/internal/messenger/mock_repository.go` and `backend/tests/messenger/mock_repository_test.go` must implement the widened repository interface. Add mutex protection around every nonce-index/message-map read and write involved in create/find, plus `createMessageCalls` and a deterministic returned-result override used later by 2C tests. Run `docker compose -f docker-compose.local.yml run --rm backend-unit`; expected RED is compile/test failure on the missing contract.

- [ ] **Step 2: Implement and accept Task 2A**

Implement `CreateMessageIdempotent` as follows:

- assign `doc.ID` and `CreatedAt` before persistence;
- empty nonce uses `InsertOne` and returns `Created=true`;
- non-empty nonce uses `UpdateOne(filter, {$setOnInsert: doc}, options.UpdateOne().SetUpsert(true))` with filter `{sender_id, client_nonce}`;
- `Created` is exactly `result.UpsertedCount == 1`; after the update, read the canonical document by the same key and return it;
- when a concurrent upsert reports `mongo.IsDuplicateKeyError`, read the canonical document and return `Created=false`; no other error is converted to a retry;
- never use an unlocked find-then-insert as the creation decision;
- add the named unique partial index to `EnsureIndexes` using `{$and: [{client_nonce: {$type: "string"}}, {client_nonce: {$gt: ""}}]}`;
- add `ClientNonce` to request, document, response, and `toMessageResponse`.

Run `docker compose -f docker-compose.local.yml run --rm backend-unit` to green, record real Mongo unique-index/upsert execution as `NOT_RUN`, update the two owned docs, then exact-path stage `backend/internal/messenger/dto.go backend/internal/messenger/repository.go backend/internal/messenger/mongo_repository.go backend/internal/messenger/mock_repository.go backend/tests/messenger/mock_repository_test.go backend/tests/messenger/message_idempotency_persistence_test.go docs/TECHNICAL.md docs/TESTING.md` and commit Task 2A.

- [ ] **Step 3: Implement and accept Task 2B fail-startup policy**

Change `wiring_messenger.go` so failure to establish required Messenger indexes follows the existing startup fatal path instead of log-and-continue. Do not add a fake static-contract unit: no genuine isolated seam exists in this slice. Run `docker compose -f docker-compose.local.yml run --rm backend-unit` for build/vet/unit/race evidence, record a real failing-startup process test as `NOT_RUN`, update the two owned docs, then exact-path stage `backend/cmd/server/wiring_messenger.go docs/TECHNICAL.md docs/TESTING.md`, commit Task 2B, and obtain independent review.

- [ ] **Step 4: Write RED units, implement, and accept Task 2C**

Add tests:

```go
func TestSendMessageRetryReturnsCanonicalMessageWithoutSideEffects(t *testing.T)
func TestSendMessageRetryAfterParentDeletionReturnsCanonical(t *testing.T)
func TestSendMessageConcurrentRetryCreatesOneMessage(t *testing.T)
func TestSendMessageNonceReuseWithDifferentChannelConflicts(t *testing.T)
func TestSendMessageNonceReuseWithDifferentPayloadConflicts(t *testing.T)
func TestSendMessageNonceOfDeletedMessageConflictsWithoutContent(t *testing.T)
func TestSendMessageCreatedFalseSuppressesEverySideEffect(t *testing.T)
func TestSendMessageCreatedTrueDeletedCanonicalSuppressesSideEffects(t *testing.T)
func TestSendMessageWithoutNonceUsesTransitionalCreate(t *testing.T)
```

For non-empty nonce, preserve membership/channel authorization, then call `FindMessageByClientNonce(senderID, nonce)` before thread-parent validation. A live compatible canonical document returns immediately; a deleted canonical or incompatible fingerprint conflicts. When no canonical exists, validate the thread parent and call `CreateMessageIdempotent`. Apply the same deleted/compatibility helper to every returned canonical document regardless of `Created`; side effects require both `Created=true` and a live compatible canonical. Run genuine RED and GREEN gates with `docker compose -f docker-compose.local.yml run --rm backend-unit`, update the two owned docs, then exact-path stage `backend/internal/messenger/message_idempotency.go backend/internal/messenger/service_messages.go backend/tests/messenger/service_messages_test.go docs/TECHNICAL.md docs/TESTING.md`, commit Task 2C, and obtain independent review.

- [ ] **Step 5: Classify Task 2 evidence**

Record each slice independently in the SDD ledger. Task 2 is accepted only when 2A, 2B, and 2C each have C0/I0 independent review and fresh Docker build/vet/unit/race PASS. Mongo driver concurrency/index execution and startup-process behavior remain explicitly `NOT_RUN` until Phase-owned `python-integration` coverage exists; do not convert compilation or mocks into integration PASS.


### Task 3: Give every logical composer submit a stable nonce

**Files:**

- Modify: `frontend/src/api/messenger.ts`
- Modify: `frontend/src/components/messenger/message-composer.tsx`
- Create: `frontend/tests/messenger/message-composer-retry-nonce.test.tsx`
- Modify: `docs/TECHNICAL.md`
- Modify: `docs/TESTING.md`

**Interfaces:**

- Consumes/produces: `SendMessageRequest.client_nonce?: string` and `Message.client_nonce?: string`.
- Logical submit state retains one `crypto.randomUUID()` through retry; a later distinct submit gets a new UUID.
- Every text, attachment, GIF, sticker, and voice submit creates an immutable payload snapshot before the first request. Text/attachment/mention/reply editor state is cleared optimistically at that boundary; a later retry must never clear or overwrite a newer draft.
- A failed snapshot is kept only in the mounted composer and exposed by one persistent, non-dismissible Sonner error toast with primary `Повторить` and secondary `Отменить отправку` actions. Retry reuses the exact snapshot and nonce; discard deletes it. Route change/reload/offline replay remains `NOT_READY` and must not be described as an outbox.

- [ ] **Step 1: Write failing text/GIF/sticker/voice nonce tests**

Mock `sendMessage`, `crypto.randomUUID`, and Sonner. Cover text/attachment, GIF, sticker, and voice routes. Force the first request to reject, invoke the toast `Повторить` action, and assert both calls use the exact same payload snapshot and nonce. Type a newer draft before the retry resolves and assert retry success does not clear it. Exercise `Отменить отправку`, then complete a new logical submit and assert it receives a different nonce. Assert route unmount dismisses its owned pending toasts.

- [ ] **Step 2: Run the Docker red gate**

Run `frontend-unit`. Expected: FAIL because current requests contain no `client_nonce`.

- [ ] **Step 3: Implement one nonce per logical submission**

Add the optional TS fields and a composer-owned `Map` of pending immutable send snapshots keyed by nonce. Route all composer send handlers through one submit/attempt path. Generate the nonce before the first API call, register the snapshot, optimistically reset only the editor state captured by that new submit, and never mutate current editor state from a later retry. On failure render `toast.error(..., {duration: Infinity, dismissible: false, action, cancel})`; retry calls the same snapshot, success invalidates messages and dismisses/deletes only its toast, and discard deletes only that snapshot. Dismiss all composer-owned pending toasts on unmount. Do not generate a nonce in the API client and do not add reconnect/reload replay in this task.

- [ ] **Step 4: Run frontend regression to green**

Run `frontend-unit`. Expected: TypeScript and Vitest PASS. Triage any failure against the spec before editing.

- [ ] **Step 5: Commit exact paths**

```powershell
git add frontend/src/api/messenger.ts frontend/src/components/messenger/message-composer.tsx frontend/tests/messenger/message-composer-retry-nonce.test.tsx docs/TECHNICAL.md docs/TESTING.md
git commit -m "feat(messenger): preserve send nonce across retries"
```

### Task 4: Reconcile message caches from canonical events and reconnect

**Files:**

- Modify: `backend/internal/messenger/events.go`
- Modify: `backend/internal/messenger/service.go`
- Modify: `backend/internal/messenger/export_hooks.go`
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

- Export `EventPublisher` from `backend/internal/messenger` with the exact contract `PublishJSON(subject string, payload any)`. `MessengerService` depends on that interface; production continues to pass `*natsclient.Client`, which satisfies it without an adapter.
- Export the hubtest-only constructor `ExportNewMessengerServiceWithEventPublisher(repo Repository, publisher EventPublisher)`. Backend units inject an in-memory capture publisher through this constructor and assert published typed payloads and JSON shape. The existing nil-NATS test constructor is not evidence for event publication, and static source scans are explicitly forbidden as false-green substitutes.
- `MessageSentEvent` carries `Message MessageResponse` as `json:"message"` plus `MentionedUserIDs`; consumers stop reconstructing a partial message from unrelated top-level fields.
- `ReactionToggledEvent` remains a delta `{message_id,channel_id,user_id,emoji,added}`; the frontend reducer applies that delta idempotently.

#### 4A: Publisher test seam

- [ ] Change `MessengerService` and its production constructor to depend on the exported `EventPublisher`, retaining direct `*natsclient.Client` production compatibility.
- [ ] Add `ExportNewMessengerServiceWithEventPublisher` behind the existing hubtest export boundary. The backend test fixture uses only an in-memory publisher that captures `(subject, payload)` calls; it must never open NATS or rely on a nil publisher.
- [ ] Add RED coverage in `backend/tests/messenger/message_sent_producers_test.go` that proves the capture sees the exact message subject and a nested `MessageSentEvent.Message`; marshal the captured event to confirm the `message` JSON field. Do not inspect source text as a proxy for execution.

#### 4B: Canonical message event plus normal and forward producers

- [ ] Add one `newMessageSentEvent(doc, mentionedUserIDs)` constructor in `events.go`; it constructs `MessageSentEvent{Message: toMessageResponse(doc), MentionedUserIDs: ...}`.
- [ ] Route normal send and forward in `service_messages.go` through that constructor. Normal send returns the same single computed `MessageResponse` used by its event.
- [ ] Extend the producer unit fixture to assert both paths publish exactly one matching `message_sent` event with the complete nested response, including nonce, thread, rich content, attachments, voice, reactions, mentions, and forwarded fields where applicable.

#### 4C: Bot, poll, and promote producers

- [ ] Route bot send, poll-message creation, and the promote backlink message through the same canonical constructor in `service_bot.go`, `service_polls.go`, and `service_promote.go`.
- [ ] Extend the owned producer fixture with the existing repository fixture plus `SetSystemBotID`, valid poll options, and fake injected promotion/permission collaborators. Filter poll assertions to its `message_sent` subject because poll creation emits an additional event.
- [ ] Assert exact message subject, a single matching publication per path, and the complete nested response rather than only message ID or text. Assert an idempotent retry cannot duplicate a canonical publication.

#### 4D: Frontend canonical cache reconciliation

- [ ] Update `buildMessageFromWsEvent` in `frontend/src/api/messenger.ts` and cache reducers in `frontend/src/hooks/use-messages.ts` to consume only `event.message` for `message_sent`.
- [ ] RED and then satisfy `use-messages-cache-events.test.tsx`: insert the full nested response exactly once by durable message ID, route replies only to their exact thread query, and preserve the root channel query for non-thread messages. Client nonce is metadata, never the cache dedupe key.

#### 4E: Reaction delta reconciliation

- [ ] Apply `ReactionToggledEvent` as an idempotent add/remove of `user_id` below the matching emoji without requiring a complete reaction list.
- [ ] RED and then satisfy `reaction-toggled-guard.test.tsx` for duplicate add/remove, removal of an empty emoji bucket, and unrelated-message non-mutation.

#### 4F: Reconnect invalidation

- [ ] In both `useMessages` and `useThreadMessages`, distinguish the initial connection from a `connected: false -> true` transition and invalidate only the active query after resubscription.
- [ ] Extend frontend cache-event coverage for active channel and thread invalidation after reconnect; initial connection must not produce a redundant recovery fetch.

#### Joint Docker gates and commit

- [ ] Run the new RED gates through `backend-unit` and `frontend-unit`, then run both Docker services again after implementation. No host test, build, or lint command is valid evidence.
- [ ] Commit only after both services PASS; Mongo/NATS integration remains outside Task 4 and is recorded separately as `NOT_RUN` unless a later scoped task adds it.

```powershell
git add backend/internal/messenger/events.go backend/internal/messenger/service.go backend/internal/messenger/export_hooks.go backend/internal/messenger/service_messages.go backend/internal/messenger/service_reactions.go backend/internal/messenger/service_bot.go backend/internal/messenger/service_polls.go backend/internal/messenger/service_promote.go backend/tests/messenger/message_sent_producers_test.go frontend/src/api/messenger.ts frontend/src/hooks/use-messages.ts frontend/tests/messenger/use-messages-cache-events.test.tsx frontend/tests/messenger/reaction-toggled-guard.test.tsx
git commit -m "fix(messenger): reconcile realtime message caches"
```

**Task boundary:** Task 4 changes only channel-scoped canonical message/reaction cache reconciliation and persistence-backed reconnect invalidation. It does not add user-room authorization or bridge subjects, `useChannels`/unread behavior, durable read-state or outbox/replay semantics, or Mongo/NATS integration; those remain Task 5, Phase 151, or explicitly `NOT_RUN` work.

### Task 5: Add authorized per-user channel lifecycle events

**Files:**

- Modify: `backend/internal/messenger/events.go`
- Modify: `backend/internal/messenger/service_channels.go`
- Modify: `backend/internal/messenger/service_channel_invites.go`
- Modify: `backend/internal/messenger/service_messages.go`
- Modify: `backend/internal/messenger/service_bot.go`
- Modify: `backend/internal/messenger/service_polls.go`
- Modify: `backend/internal/messenger/service_promote.go`
- Modify: `backend/internal/ws/authorizer.go`
- Modify: `backend/internal/ws/bridge.go`
- Modify: `backend/internal/ws/hub.go`
- Modify: `backend/internal/ws/export_hooks.go`
- Modify: `backend/cmd/server/wiring_ws.go`
- Modify: `backend/tests/ws/authorizer_test.go`
- Create: `backend/tests/ws/bridge_user_events_test.go`
- Create: `backend/tests/ws/bridge_channel_authz_test.go`
- Create: `backend/tests/messenger/user_channel_events_test.go`
- Modify: `backend/tests/messenger/message_sent_producers_test.go`
- Modify: `frontend/src/hooks/use-channels.ts`
- Modify: `frontend/tests/messenger/use-channels-ws.test.tsx`
- Modify: `docs/TECHNICAL.md`
- Modify: `docs/TESTING.md`
- Create: `.planning/phases/149-messenger-correctness-and-realtime-foundation/149-VERIFICATION.md`

**Frozen interfaces and security boundary:**

- User subjects are exactly `messenger.user.<user-id>.channel_created`, `.channel_updated`, `.channel_removed`, and `.new_message_notification`.
- Lifecycle payloads are invalidation-only and exactly `{channel_id}`; `new_message_notification` is exactly `{channel_id,message_id}`. They never carry message content, invite tokens, or recipient-local `is_starred`/mute fields.
- The websocket bridge accepts exactly four dot-separated user-subject segments, validates the UUID and event whitelist, then maps them to room `user:<user-id>` and the suffix event.
- `CanSubscribe` accepts `user:<uuid>` only when the UUID equals the authenticated user; `user:<uuid>:channels`, foreign, nil, and malformed rooms fail closed. Channel authorization remains membership-backed.
- Every Messenger and Calls channel-scoped delivery re-resolves current channel members through an injected `ChannelMembersResolver` before fanout. `Hub.BroadcastToRoomForUsers` serializes once and delivers only to connections whose authenticated `UserID` is in the resolved set. Malformed IDs, nil resolver, timeout, or repository error drop the event for everyone.
- A lost best-effort `channel_removed` event cannot authorize later channel delivery; it is a UI invalidation signal, not the security boundary. The delivery check is fail-closed but not globally linearizable with concurrent removal: an event already authorized/in flight before removal may arrive. Transactional outbox/ACL versioning is outside Task 5 and remains explicit later hardening work.
- Create targets all successfully persisted members; update targets current members; delete snapshots recipients before mutation; join and accepted invite target only the newly added user; leave targets only the removed user. Publication happens only after mutation succeeds.
- Normal, forward, bot, poll-message, and promoted-message producers notify current members other than the canonical sender. An idempotent nonce retry publishes nothing.
- Every lifecycle/notification event invalidates exactly `CHANNELS_QUERY_KEY`; the client never patches the channel list or performs duplicate-sensitive local unread arithmetic. On reconnect it subscribes to `user:<id>` first and then invalidates; initial connect only subscribes.

- [ ] **Step 1: Batch genuine failing backend and frontend tests**

Backend authorizer tests cover own user PASS, foreign user DENY, malformed UUID DENY, `user:<id>:channels` DENY, accessible channel PASS, and inaccessible private channel DENY. Messenger tests cover channel create/update/delete/join/leave, `AcceptInvite`, and all five canonical message-producing paths with exact subjects, recipient sets, invalidation-only payloads, post-success publication, sender exclusion, and retry suppression. Bridge tests exercise exact user-subject routing and malformed drops; active/former already-subscribed clients for both Messenger and Calls; resolver error/timeout drops for all; and a missed removal event cannot leak the next channel event. Frontend tests assert exact `user:<id>`, all four invalidations, duplicate notification harmlessness, no local unread arithmetic, and subscribe-before-invalidate reconnect ordering without initial redundant invalidation.

- [ ] **Step 2: Run one low-disk RED gate per stack**

Run `backend-unit` once and `frontend-unit` once after batching their RED oracles. Expected: FAIL on unsupported user rooms, missing exact bridge routing/fail-closed fanout, missing producers, and old frontend room/arithmetic behavior. Do not loop Docker per atomic edit and do not use host execution.

- [ ] **Step 3: Implement the frozen contract in atomic slices**

Implement **5A** typed event/auth contract; **5B** channel lifecycle producers; **5C** accepted-invite producer; **5D** normal/forward notifications; **5E** bot notification; **5F** poll-message notification; **5G** promoted-message notification; **5H** exact user bridge plus shared fail-closed Messenger/Calls channel fanout with a short membership-resolution timeout and a pure/testable subscription-definition seam; **5I** frontend exact subscription and persistence invalidation. Production wiring injects the existing Messenger repository membership query. Do not weaken existing channel/call/project authorization or use source scans as executable coverage.

- [ ] **Step 4: Run full Phase 149 unit gates and classify integration**

Run:

```powershell
docker compose -f docker-compose.local.yml run --rm backend-unit
docker compose -f docker-compose.local.yml run --rm frontend-unit
```

Expected: PASS. Record Mongo, NATS broker, live WebSocket, and deployed runtime as `NOT_RUN` unless greenfield integration coverage is added under `testframework/tests/integration/**` and executed through the existing `python-integration` service.

- [ ] **Step 5: Sync docs and commit exact paths**

Update the websocket subject/security table, lifecycle recipient/payload rules, idempotency/index contract, fail-closed fanout and its in-flight race boundary, Phase-151 unread recovery boundary, reconnect behavior, Task 1-5 receipts, new/legacy test status, and protected scope in `docs/TECHNICAL.md` and `docs/TESTING.md`. Create `149-VERIFICATION.md` with separate static, backend-unit, frontend-unit, and Mongo/NATS/live-WebSocket rows; unavailable executable evidence stays `NOT_RUN`/`PARTIAL`.

```powershell
git add backend/internal/messenger/events.go backend/internal/messenger/service_channels.go backend/internal/messenger/service_channel_invites.go backend/internal/messenger/service_messages.go backend/internal/messenger/service_bot.go backend/internal/messenger/service_polls.go backend/internal/messenger/service_promote.go backend/internal/ws/authorizer.go backend/internal/ws/bridge.go backend/internal/ws/hub.go backend/internal/ws/export_hooks.go backend/cmd/server/wiring_ws.go backend/tests/ws/authorizer_test.go backend/tests/ws/bridge_user_events_test.go backend/tests/ws/bridge_channel_authz_test.go backend/tests/messenger/user_channel_events_test.go backend/tests/messenger/message_sent_producers_test.go frontend/src/hooks/use-channels.ts frontend/tests/messenger/use-channels-ws.test.tsx docs/TECHNICAL.md docs/TESTING.md .planning/phases/149-messenger-correctness-and-realtime-foundation/149-VERIFICATION.md
git commit -m "feat(messenger): deliver authorized user channel events"
```

## Final Phase Gate

1. Run `backend-unit`, then `frontend-unit`, twice only if the first pass exposed a real flaky/infra classification; never loop blindly.
2. Search the changed frontend for `user:${userId}:channels` and the changed backend for direct `InsertMessage` use in user send paths; expected product hits: zero.
3. Independent review receives `149-SPEC.md`, exact BASE..HEAD diff, raw Docker outputs, and no generator conclusion.
4. Update `149-VERIFICATION.md` with separate static/runtime verdicts. If Mongo/NATS integration was not executed, the phase remains `PARTIAL` with `python-integration: NOT_RUN`; do not relabel it PASS.
5. Phase 151 retains the explicit `NOT_READY` durable read-state/projection rebuild gap; Phase 149 PASS never implies that cross-storage recovery exists.
