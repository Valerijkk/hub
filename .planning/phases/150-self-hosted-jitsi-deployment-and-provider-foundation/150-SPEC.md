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

Room creation/join depends on a locally verified Hub-issued credential or an explicitly allowed guest token. No public anonymous room creation, JaaS token, or `meet.jit.si` fallback exists.

### CALL150-06 Operability and rollback

Configuration documents resource expectations, health/readiness, logging/metrics, restart behavior, backup scope, upgrade, and rollback. LiveKit remains intact and selectable throughout this phase.

## Security requirements

- Service secrets remain server-side and use existing Hub secret handling patterns.
- Component networks and administrative endpoints are least-exposed.
- Provider callbacks/events must have an authentication and replay/idempotency design before they mutate Hub call state.
- Default-deny egress and explicit ingress/media port documentation are required.
- Supply-chain evidence includes license notice, source provenance, and digest verification.

## Non-goals

- No production provider cutover.
- No LiveKit dependency/config deletion.
- No Jibri recording implementation.
- No final Jitsi conference UI or broad Calls redesign.
- No Mail/Calendar edits and no change to `frontend/src/modules/calls/create-call-dialog.tsx`.

## Verification

- Provider configuration and adapter contract unit tests run in Docker `backend-unit`.
- Compose/config static validation and a task-owned isolated Jitsi smoke are separate evidence classes.
- A successful container start is not labelled product parity.
- Anything requiring registry/network/media infrastructure that is unavailable remains `NOT_RUN` or `BLOCKED` with the exact continuation command.
