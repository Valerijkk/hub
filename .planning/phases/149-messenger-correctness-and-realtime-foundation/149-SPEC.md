# Phase 149: Messenger correctness and realtime foundation - Specification

## Goal

Establish a persistence-first, tenant-safe, retry-safe Messenger foundation on which channels, threads, and unread behavior can be completed without data corruption or unrecoverable realtime gaps.

## Functional requirements

### MSG149-01 Thread parent validation

Before creating a reply, the backend loads the referenced parent and verifies that it exists, belongs to the requested channel, and is a valid thread root/reply target under the current deletion policy. Failure returns a stable client error and produces no message or event.

### MSG149-02 Send idempotency

User-originated send accepts `client_nonce`. The storage key is the authenticated sender plus nonce; the stored channel and normalized message payload are the idempotency fingerprint. Sequential and concurrent retries with the same fingerprint return the same durable message identity, while reuse for another channel or payload returns conflict. Storage has a unique constraint/index for the key, and publication is logically deduplicated.

### MSG149-03 Canonical subject grammar

The backend and frontend use exactly `channel:<uuid>` for channel content/lifecycle events and `user:<uuid>` for per-user membership/sidebar/unread/mention events. Legacy unsupported subject spellings are removed or handled by an explicitly tested transition adapter.

### MSG149-04 Subscription authorization

An authenticated user may subscribe to their own user subject and to a channel subject only while membership/channel visibility permits it. They may not subscribe to another user's subject or an inaccessible private channel.

### MSG149-05 Missing event producers

Every event consumed by the current Messenger frontend has a matching backend producer with an explicit payload contract, or the unused consumer is removed. Channel create/update/member changes and message notification behavior must not depend on phantom events.

### MSG149-06 Reconnect recovery

After an interrupted websocket connection, the client deterministically refetches or resumes from a durable cursor so messages, channel membership, and sidebar state converge to persisted truth. Duplicate events remain harmless.

## Security requirements

- Server derives sender and subscription user identity from authenticated context, never from a trusted client field.
- Channel/tenant authorization is checked before parent lookup results or event data can leak.
- Nonces are bounded and validated; indexes must not allow cross-tenant collision or unbounded abuse.
- Event payloads contain only data already authorized for the target subject.

## Compatibility requirements

- Existing persisted messages remain readable.
- DTO evolution is compatible with the frontend migration in the same phase.
- Phase 149 prevents duplicate projections on a retry but does not claim durable recovery of the current best-effort Redis unread projection after an infrastructure failure; durable read-state and projection rebuild are owned by Phase 151.
- No Mail/Calendar path or protected call-dialog path is modified.
- No legacy non-unit suite is revived; new executable non-unit coverage belongs to `testframework/**`.

## Verification

- TDD unit coverage for parent validation, nonce retry/concurrency, subject authorization, producer mapping, and reconnect reducer/refetch behavior.
- Fresh Docker-only `backend-unit` and `frontend-unit` evidence for applicable code.
- Static evidence is labelled separately from runtime PASS.
