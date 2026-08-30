# Phase 150: Self-hosted Jitsi deployment and provider foundation - Specification

## Goal

Add a reproducible self-hosted Jitsi foundation and provider-neutral Hub calls contract without changing the active production provider or deleting the rollback path.

## Functional requirements

### CALL150-01 Provider-neutral session contract

The calls domain requests a conference session through a provider interface that returns a typed provider-neutral join credential containing provider, room, service URL, and token. Call ACL, membership, lifecycle, scheduling, chat, notifications, and audit remain outside the media provider.

### CALL150-02 Backward-safe provider selection

Configuration selects the calls provider explicitly. A missing selector defaults to LiveKit for backward compatibility, while an explicit unknown, unavailable, or partially configured provider fails closed with actionable diagnostics. Existing installations retain LiveKit behavior until a deliberate Jitsi cutover.

### CALL150-03 Self-hosted Jitsi topology

Hub deployment assets define the required Jitsi web, XMPP/control, conference focus, and video bridge services on Hub-controlled networks. Public web/media ports, internal-only services, health checks, persistence, and secrets are explicit.

### CALL150-04 Immutable upstream pinning

Every upstream Jitsi image uses an accepted release and immutable digest. The repository records the source release, image names, resolved digests, retrieval date, and deterministic refresh/verification procedure. Unverified digests are never committed as if confirmed.

### CALL150-05 Local authentication and room policy

Jitsi join is limited to an authenticated Hub principal after Hub ACL/membership authorization and a short-lived Hub-issued credential. Guest and anonymous/public rooms, JaaS, `meet.jit.si`, and external provider/API fallback are forbidden.

### CALL150-06 Operability and rollback

Configuration documents resource expectations, health/readiness, logging/metrics, restart behavior, backup scope, upgrade, and rollback. LiveKit remains intact and selectable throughout this phase.

## 2026-08-30 bounded durable-calls amendment

This amendment makes the accepted provider, presence, event-authority, and single-surface contracts explicit. It supersedes only conflicting Phase-150 scope wording; it does not claim that any roadmap atom, runtime path, provider cutover, media smoke, or product-parity gate has passed.

### CALL150-07 Immutable provider and presence identity

`Call.ID` remains the Hub business identity. Persisted, Hub-generated `Call.RoomName` is the opaque provider session and the exact Jitsi JWT `room`; it is never accepted from a client as authorization. Every Call records immutable `provider=livekit|jitsi` and `presence_mode=legacy|durable`. Existing Calls remain `livekit/legacy`, and the active-provider selector affects new Calls only. An active Call never changes provider or presence writer.

### CALL150-08 Durable connection, presence edges, and publication

A provider connection is identified by `(CallID, UserID, Epoch)`. Credential issuance creates a new nonzero epoch in unknown state but does not create presence. Public `joined` is emitted only for aggregate refcount `0 -> 1`, public `left` only for `1 -> 0`, and ending a Call emits one ended event without a per-user left storm; an unknown or reconciled leave never creates a public presence edge. Public durable event payloads contain only event ID, Call ID, User ID, and presence sequence; they never expose epoch, provider/delivery identity, token, or raw envelope. Durable identities and counts are unique and nonnegative; historical `CallParticipant` rows never imply durable connection or presence state. A verified provider delivery is applied by one PostgreSQL transaction owning inbox insertion, connection/refcount and compatibility projection updates, and platform-outbox insertion. Publication is post-commit and at-least-once; within an ordered Call stream an earlier leased or failed row blocks later rows.

### CALL150-09 Authenticated durable event authority

Durable deduplication accepts only a verified provider-side `(provider, delivery_id)`; timestamps, payload hashes, room or participant values, browser events, and Hub-generated ingress IDs are not delivery identities. LiveKit durable ingress remains disabled until the pinned version proves a signed, nonempty delivery ID stable across retries. Jitsi authority is limited to a Hub-owned custom Prosody module loaded through the selected release's supported custom-plugin/MUC-module path. Exact lifecycle hooks/order and any equivalent stock durable egress are `UNVERIFIED_FOR_SELECTED_DIGEST`; no stock facility is accepted without JITSI-D1 proof. Before send, the custom module persists a provider-deployment delivery ID and exact envelope through a Prosody named SQL store into a dedicated self-hosted provider-side PostgreSQL database that shares no Hub application database/schema/role/credentials. D1 proves only selected-image driver/config capability; separate authorized JITSI-02D fault injection must prove that successful append acknowledgement follows a completed durable database commit. Memory/null/internal-file/SQLite storage, MUC state, Hub application databases, and unverified container filesystems are forbidden. A separately verified private transport adapter/sidecar provides mutually authenticated TLS because direct Prosody client-certificate support is not assumed. Before parse/apply, Hub verifies that peer identity and an independent HMAC-SHA256 signature over the exact stored raw envelope. Browser iframe events, public Jitsi, JaaS, and other provider APIs are not state authority. Hub/Jitsi JWT, XMPP component/auth, mTLS, and event-signing credentials remain distinct.

### CALL150-10 Bounded Hub-native Jitsi surface

The only Phase-150 Jitsi UI allowance is one later-gated provider-aware Hub conference surface driven by a typed join DTO. It accepts only the exact validated self-hosted `JITSI_PUBLIC_URL`, renders the provider persisted on the Call, creates and disposes provider resources within the Call room lifecycle, never logs or exposes tokens, and never enables anonymous Jitsi guests or an external-origin fallback. This allowance does not authorize a broad Calls redesign, provider parity claim, UI activation, or default provider cutover before its separate acceptance gates pass.

## Security requirements

- Service secrets remain server-side and use existing Hub secret handling patterns.
- Hub `JITSI_APP_SECRET` maps explicitly to upstream `JWT_APP_SECRET`; `JICOFO_COMPONENT_SECRET`, `JICOFO_AUTH_PASSWORD`, `JVB_AUTH_PASSWORD`, event HMAC, and mTLS identities are separate required secret domains with no usable defaults.
- Component networks and administrative endpoints are least-exposed.
- Provider callbacks/events must satisfy `CALL150-09` authentication, replay, stable-delivery-identity, and atomic-application rules before they mutate Hub call state.
- Default-deny egress and explicit ingress/media port documentation are required.
- Supply-chain evidence includes license notice, source provenance, and digest verification.

## Non-goals

- No production provider cutover.
- No LiveKit dependency/config deletion.
- No Jibri recording implementation.
- No broad Calls redesign, custom reimplementation of the upstream Jitsi conference feature set, multi-surface Jitsi experience, or final provider-parity UI beyond the single bounded surface in `CALL150-10`.
- No Mail/Calendar edits and no change to `frontend/src/modules/calls/create-call-dialog.tsx`.

## Verification

- Provider configuration and adapter contract unit tests run in Docker `backend-unit`.
- Provider/presence identity, connection/refcount edges, authenticated replay handling, and ordered outbox effects require focused Docker evidence before their corresponding roadmap atoms can pass.
- Missing, tampered, replayed, stale, oversized, wrong-content-type, or wrong-credential Jitsi bridge envelopes must fail closed without mutation or credential leakage at the applicable atom boundary.
- Before D1, `JITSI-02A` is limited to a provider-neutral HMAC-SHA256 integrity primitive over exact raw bytes. It returns only a package-owned populated verified-value wrapper with defensive byte copies, rejects secrets shorter than 32 bytes, bodies outside `1..65536` bytes and already-decoded binary MACs not exactly 32 bytes, and exposes one non-diagnostic sentinel for all invalid/authentication failures. It does not decode transport headers or hex/base64, parse JSON, inspect delivery IDs, decide freshness, persist replay state or mutate Calls; those remain later ingress/durability gates.
- `CALL-DUR-03` is only an opaque wrapper around the existing valid `ConfirmedParticipantConnectionKey` plus an exact parser/canonicalizer for `hub.call-participant.v1/<call>/<user>/<epoch>`. Its parser accepts exactly three non-zero lowercase canonical hyphenated UUID segments and performs no normalization, URL decoding, UUID-version filtering or authorization decision. All invalid inputs collapse to one exact non-diagnostic sentinel and a zero value; the identity is correlation data, not trust evidence.
- Source-spool crash/retry tests must prove that a delivery ID and exact envelope bytes are durable before first send and identical across retries; stock-hook observation alone is not accepted as at-least-once evidence.
- Source tests must reject focus/system/recorder/bridge identities, anonymous/guest principals, unmapped rooms, and unaccepted breakout semantics. Room destruction never ends a Hub Call without the separate CALL-DUR-09 decision.
- Target provider-source production semantics are only `at-least-once after successful durable enqueue` after B1/B2/D evidence. Production acceptance must close the pre-append/database-outage observation gap through bounded reconciliation or a separately proven pre-event fail-closed gate.
- Docker `frontend-unit` must prove the single-surface provider selection, exact self-hosted-origin validation, anonymous-Jitsi denial, token non-disclosure, and dispose/unmount cleanup. Browser/media parity remains a separate future gate.
- Compose/config static validation and a task-owned isolated Jitsi smoke are separate evidence classes.
- A successful container start is not labelled product parity.
- A typed DTO, static manifest, source diff, or iframe render does not establish provider cutover, deployment readiness, or product parity.
- Anything requiring registry/network/media infrastructure that is unavailable remains `NOT_RUN` or `BLOCKED` with the exact continuation command.
