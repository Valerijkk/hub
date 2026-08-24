# Phase 150: Self-hosted Jitsi deployment and provider foundation - Context

**Gathered:** 2026-08-24
**Status:** Ready for planning
**Mode:** Founder decisions accepted; no remaining product question for this phase

<domain>
## Phase Boundary

Introduce a self-hosted, pinned Jitsi deployment contour and a provider-neutral calls seam. All services run on infrastructure controlled by the Hub owner. This phase does not cut production calls over, remove LiveKit, deliver the final branded conference UI, or enable recording; those belong to Phases 152 and 154.

</domain>

<decisions>
## Implementation Decisions

### Distribution and ownership

- Use the maintained upstream Jitsi distribution/containers, not JaaS, `meet.jit.si`, or another vendor's host.
- Hub owns manifests, local configuration, secrets contract, branding inputs, provider adapters, monitoring, backups, runbooks, and upgrade/rollback receipts.
- Pin an official stable upstream release and immutable image digests. Floating tags are forbidden.
- Do not deep-fork or vendor all Jitsi source repositories into Hub.

### Provider migration

- Add a provider-neutral boundary around the existing Hub call domain.
- Hub remains responsible for call authorization, membership, invites, scheduling records, chat/domain events, audit, and notifications.
- LiveKit remains a shadow/rollback provider during Phases 150-153. Removal is forbidden before the Phase 154 exit gate.
- Existing recordings and historical records remain readable; they are not rewritten merely to erase provider names.

### Security and topology

- Jitsi authentication is local and Hub-issued; anonymous room creation is forbidden.
- Browser-visible configuration contains no service secrets.
- Internal component networks are private; required media ingress is explicit; administrative interfaces are not exposed as Hub product UI.
- Egress is denied by default except for documented operational dependencies approved for this deployment.

### Agent discretion

- Exact config object names and provider interface signatures follow the current calls package style, provided LiveKit-specific token/URL fields do not remain the canonical domain contract.
- The initial local topology may be single-bridge for bounded team size, but configuration must not prevent later bridge scaling.

</decisions>

<code_context>
## Existing Code Insights

- Current call room creation/join and server wiring are directly tied to LiveKit.
- Current frontend call components use LiveKit React packages directly.
- Current recording implementation is LiveKit Egress-specific and is intentionally outside this phase.
- Current Compose, Caddy, environment, dependency, and deploy files contain LiveKit-specific runtime configuration.
- Mail/Calendar and `frontend/src/modules/calls/create-call-dialog.tsx` are protected until the other workstream merges.

</code_context>

<specifics>
## Specific Ideas

- Land the provider interface and config validation behind a default that preserves current runtime behavior.
- Add Jitsi manifests in parallel rather than destructively replacing LiveKit assets in this phase.
- Resolve and record release tags and digests through official sources plus an approved Docker inspection command; never invent a digest when registry access is unavailable.
- Treat a container starting as infrastructure smoke only, not product acceptance.

</specifics>

<deferred>
## Deferred Ideas

- Branded ready Jitsi conference surface and final user controls: Phase 152.
- Jibri recording, SeaweedFS import, cutover, rollback proof, and LiveKit retirement: Phase 154.
- Calls/Messenger cross-module acceptance: Phase 155.
- Mail/Calendar scheduling integration: Phase 164 after merge to `master`.

</deferred>
