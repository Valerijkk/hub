# Phase 149: Messenger correctness and realtime foundation - Context

**Gathered:** 2026-08-24
**Status:** Ready for planning
**Mode:** Founder decisions accepted; no remaining product question for this phase

<domain>
## Phase Boundary

Make the existing Messenger persistence and realtime contracts correct before adding broader workplace features. This phase covers thread-parent integrity, send idempotency, authorized event subjects, missing event production, and reconnect recovery. It does not redesign the Messenger UI or deliver the full Slack-class feature set; those belong to Phases 151 and 153.

</domain>

<decisions>
## Implementation Decisions

### Message and thread integrity

- Any message may be the root of a side-panel thread.
- A thread reply must reference an existing parent in the same channel.
- Invalid, missing, cross-channel, or otherwise unusable parent references are rejected before persistence and publication.
- A reply is not duplicated into the main channel timeline unless a later phase explicitly supplies the accepted "also send to channel" behavior.

### Idempotent send

- The client supplies `client_nonce` for user-originated sends.
- Retries with the same authenticated sender, channel, and nonce resolve to one durable message and one logical publication.
- Concurrent duplicates must be prevented by storage-level enforcement, not only a process-local check.
- Backward compatibility for existing clients is allowed only as an explicit transitional path with tests and documentation.

### Realtime subjects

- Content/lifecycle events use `channel:<channel-id>`.
- Per-user membership, sidebar, unread, mention, and notification events use `user:<user-id>`.
- Subscription authorization must accept only subjects the authenticated user is allowed to observe; arbitrary user or channel subjects are forbidden.
- Persistence succeeds before publication. Publish failure cannot fabricate a successful delivered state; reconnect/refetch must recover durable data.

### Reconnect

- Reconnect correctness is based on durable cursors or deterministic refetch, not on assuming the websocket delivered every event.
- Separate unread semantics are completed in Phase 151; this phase provides the safe transport/recovery foundation.

### Agent discretion

- Exact Mongo index shape, backward-compatible DTO field optionality, event envelope additions, and cursor endpoint placement follow existing Hub patterns unless those patterns violate the locked behavior above.

</decisions>

<code_context>
## Existing Code Insights

- Current send flow accepts `ThreadID` without proving the parent exists in the same channel.
- Current frontend subscriptions reference channel lifecycle/notification events for which backend producers are incomplete.
- Current frontend uses a per-user channel subject shape that does not match the websocket authorization contract.
- Current send contracts do not expose a durable `client_nonce` idempotency key.
- Existing Messenger domain data and native Hub UI remain the system of record and product surface.

</code_context>

<specifics>
## Specific Ideas

- Start with failing backend unit tests for cross-channel/missing parent and concurrent duplicate send.
- Freeze the subject grammar once and share it between producer, authorizer, and frontend consumer tests.
- Prefer a reconnect cursor if current event ordering exposes a durable sequence; otherwise specify a deterministic refetch receipt and make the limitation explicit.
- Run all test/build/lint evidence only through the accepted root Compose services.

</specifics>

<deferred>
## Deferred Ideas

- Channel types, thread side panel, follow/unfollow, and separate unread categories: Phase 151.
- Full workplace feature parity and UI polish: Phase 153.
- Mail/Calendar notifications and scheduling integration: Phase 164 after the protected merge.

</deferred>
