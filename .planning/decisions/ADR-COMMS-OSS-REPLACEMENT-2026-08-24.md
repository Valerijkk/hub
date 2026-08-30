# ADR: Communications and OSS replacement strategy

- **Status:** Accepted (founder lock)
- **Date:** 2026-08-24
- **Scope:** Messenger, Calls, and the queued self-hosted OSS replacements
- **Current delivery scope:** Messenger and Calls only

## Context

Hub is a self-hosted corporate super-app. The communications module must become a complete workplace messenger, while calls should use a mature conferencing core instead of maintaining a weak custom implementation. Accepted third-party products must run inside infrastructure controlled by the Hub owner; Hub must not depend on a SaaS API, JaaS, `meet.jit.si`, or hosting operated by another vendor.

"No external API" in this decision means no cloud/SaaS control plane or externally hosted runtime. Local service interfaces between Hub and self-hosted components remain allowed and are required for authentication, authorization, lifecycle, storage, audit, and observability.

## Decisions

### 1. Messenger is a native Hub product surface

Messenger will provide the workplace capabilities users reasonably expect from Slack without copying Slack's visual design one-to-one:

- public, private, project, and read-only channels with topic, description, membership, archive, and retention controls;
- direct messages and group direct messages;
- a side-panel thread for any message; thread replies do not pollute the channel timeline unless the sender explicitly selects "also send to channel";
- automatic thread follow for authors, repliers, and mentioned users plus manual follow/unfollow;
- separate unread state for channels, direct messages, mentions, and followed threads;
- reactions, edits, deletes, pins, bookmarks/saved items, search, files, mentions, presence, typing indicators, notification preferences, and moderation/audit controls;
- persist-first realtime delivery, reconnect cursor/refetch, `client_nonce` idempotency, and explicit event routing through `channel:<id>` and `user:<id>` subjects.

Hub owns the Messenger UI, message/channel data, RBAC, audit, notifications, and search contracts.

### 2. Calls move from LiveKit to a self-hosted Jitsi distribution

The current LiveKit-based conference runtime will be replaced with a ready Jitsi Meet build configured and branded for Hub. Hub will not recreate the conferencing client feature-by-feature and will not vendor a deep fork of every Jitsi repository.

Hub owns:

- deployment manifests, local configuration, branding, provider adapters, monitoring, backups, and runbooks;
- pinned upstream versions and immutable container image digests;
- call authorization, invite policy, scheduling/domain records, audit, notifications, and navigation;
- local JWT/session issuance and the boundary between Hub identities and Jitsi rooms;
- recording metadata and import into Hub-controlled storage.

All Jitsi services, including recording workers when enabled, run on Hub-controlled infrastructure. Calls must cover audio, video, screen sharing, device selection, roster, reconnect/error states, raise hand, reactions, in-call chat, guest policy, and recording. LiveKit remains available only during a reversible shadow/cutover window and is removed after parity evidence, rollback readiness, and active-call/recording drain are confirmed.

### 3. Accepted OSS modules are self-hosted replacements

The following modules are accepted for later delivery after Messenger and Calls:

| Hub capability | Accepted upstream | Replacement boundary |
|---|---|---|
| Whiteboards | Excalidraw | Upstream canvas embedded in a native Hub shell; Hub owns board lifecycle, ACL, storage, and audit. |
| Bases/forms/no-code | Baserow | Full replacement of native `Bases`; no coexistence period beyond controlled migration/rollback. |
| Automations/connectors | Activepieces Community Edition | Replacement of the corresponding native automation engine; Hub owns policy, approvals, audit, and navigation. |
| BI/analytics | Apache Superset | Advanced BI over read-only reporting projections; never direct writes to Hub operational databases. |
| Collaborative documents | ONLYOFFICE Docs | Upstream editor surface with Hub-owned files, ACL, lifecycle, and local callbacks. |
| Electronic signatures | Documenso | Self-hosted signing workflow; signed artifacts and evidence are imported into Hub-controlled storage. |
| External helpdesk | Chatwoot | Customer-support capability only; it does not replace the internal workplace Messenger. |

The normal UX is native Hub list/detail/settings/navigation. Specialized upstream surfaces remain upstream where rewriting them would destroy the cost advantage: Jitsi conference, Excalidraw canvas, and ONLYOFFICE editor. Upstream admin consoles are operations/admin-only.

### 4. Shared integration rules

- Every service is self-hosted in the Hub deployment contour; external SaaS dependencies and vendor-hosted control planes are forbidden.
- Hub remains the authority for identity, tenant boundary, RBAC, navigation, audit, and product lifecycle.
- Each component uses its own supported data store. Sharing the Hub application database or writing directly into another component's database is forbidden.
- Every business object has one writable system of record; synchronization uses explicit local adapters and an outbox/idempotent event contract where applicable.
- Runtime egress is denied by default and opened only for a founder-approved connector or operational dependency.
- Versions and image digests are pinned; provenance, SBOM, license notices, upgrade and rollback procedures are mandatory.
- AGPL components, especially ONLYOFFICE Docs and Documenso, require a distribution/deployment legal review before release.

## Delivery order and protected work

1. Complete Messenger correctness and workplace parity.
2. Introduce self-hosted Jitsi in parallel, cut calls over safely, then remove LiveKit.
3. Run communications acceptance and regression gates.
4. Deliver the queued OSS replacements as separate phases after communications.
5. Mail and Calendar remain owned by their dedicated branch/worktree. This branch must not modify their protected paths. After that work merges to `master`, merge `master` into this branch and perform the planned cross-module integrations.

## Acceptance boundary

An integration is not accepted because an upstream container starts or an iframe renders. Acceptance requires Hub authentication and authorization, tenant isolation, native navigation, auditability, observability, backup/restore, upgrade/rollback, Docker-only executable evidence, and removal or quarantine of the replaced implementation after the rollback window.

## Consequences

- Hub avoids long-term ownership of commodity conferencing and specialist editor engines.
- Operations complexity moves into pinned self-hosted services and explicit adapters.
- Calls migration is intentionally staged instead of deleting the rollback path before production evidence exists.
- Future OSS modules cannot share databases or silently become a second identity/permission authority.

## Accepted implementation clarifications

- Messenger send idempotency uses `(authenticated_sender_id, client_nonce)` as the unique key and binds the stored channel plus normalized payload as the fingerprint; incompatible reuse is a conflict.
- Phase 149 suppresses duplicate projections on retry. Durable read state and repair of the current cross-storage Redis unread projection are intentionally owned by Phase 151 and are not implied by Phase 149 acceptance.
- Phase 150 freezes Jitsi `stable-11146-2` by immutable OCI index digest, uses a dedicated `meet.<Hub domain>` origin and local JWT, disables JaaS/external analytics/public STUN defaults, exposes only Caddy HTTPS and JVB UDP 10000, and keeps LiveKit active.
- Jitsi runtime/media/SBOM evidence is a separate deployment gate; static manifests and registry metadata alone cannot authorize cutover.

## Accepted architecture amendment - 2026-08-30

The following decisions close the Calls/Jitsi and Messenger runtime gray zones without changing the acceptance boundary above:

- Every `Call` permanently records its `provider` (`livekit|jitsi`) and `presence_mode` (`legacy|durable`). Configuration selects a provider only for a newly created call; an active call is never moved between providers or presence writers.
- `Call.ID` remains the Hub business identity. The persisted, Hub-generated `Call.RoomName` remains the opaque provider session name and is the exact Jitsi JWT `room` claim. It is not an authorization grant, tenant key, invite, or client input.
- Durable participant identity is a Hub-issued `(CallID, UserID, Epoch)` value. Credential issuance creates an unknown connection but is not proof of presence. Public `joined` is emitted only on aggregate `0 -> 1`, `left` only on `1 -> 0`, and call end never produces a per-user left storm.
- A provider observation is admitted only after provider authentication and a stable provider-issued delivery identity are proven. Timestamp, payload hash, room, participant identity, or a Hub-generated UUID cannot substitute for the delivery ID.
- One Calls database transaction owns `inbox -> connection/refcount/compatibility projection -> outbox`. NATS/WebSocket delivery remains at-least-once through the platform outbox and occurs only after commit. LiveKit durable ingress stays disabled until its pinned version proves a signed stable retry identity.
- Jitsi authoritative events come from a Hub-owned self-hosted bridge, never from browser iframe events or an external service. The bridge persists a stable delivery ID before delivery and signs the exact raw envelope with a distinct HMAC-SHA256 secret over private mTLS. Hub JWT, Jitsi AppSecret, LiveKit credentials, mTLS identities, and event-signing credentials are not reused.
- The Jitsi release remains `stable-11146-2` at the recorded immutable digests. Exact Linux image discovery, SBOM/license verification, numeric disk budget, supported mounts/users/health/event hooks, and isolated Docker evidence precede an executable overlay or provider activation. Only Caddy HTTPS and JVB UDP are public.
- Messenger recovery runs as a dedicated self-hosted worker. A PostgreSQL global leader lease/fence/checkpoint limits process ownership while the existing Mongo message owner/fence/lease protects each pending message. Neither lease replaces the other.
- Messenger completion records a replayable Mongo delivery intent. A materializer writes its deterministic unique delivery ledger and platform outbox event in one PostgreSQL transaction. For canonical/recovery delivery types, only the existing outbox worker publishes NATS after current-audience re-resolution. The isolated legacy publisher may coexist only before a tenant cutover, never for the same request, and must retire before default canonical enablement. Delivery is at-least-once, never exactly-once.
- Messenger writer cutover is staged per tenant with exactly one writer. Rollback changes future writes only and continues draining already canonical intents/outbox rows; no dual-write or state-rewrite rollback is permitted.
- Phase 149 unit/static acceptance may be used as a dependency, but its Mongo/NATS/WebSocket/browser and cross-store residuals remain `NOT_RUN` and must be closed by Phases 151/155. Phase 150 remains `STATIC_FOUNDATION_ONLY` until image, deployment, media, security, rollback, and operational evidence exist.

Implementation follows atomic RED -> GREEN -> independent review -> docs sync -> exact-path commit gates. Host tests/builds remain forbidden; Docker execution is accumulated and deferred while adjacent Mail/Calendar work or the disk/load budget makes another container unsafe.
