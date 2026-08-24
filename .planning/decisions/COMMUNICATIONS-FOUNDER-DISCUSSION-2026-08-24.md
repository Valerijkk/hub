# Communications founder discussion log

**Captured:** 2026-08-24
**Status:** Decisions accepted; implementation authorized
**Binding synthesis:** `.planning/decisions/ADR-COMMS-OSS-REPLACEMENT-2026-08-24.md`

This log preserves the product discussion that led to the binding ADR. It is a decision record, not a verbatim transcript.

## Sequence of decisions

1. The current branch is assigned to finishing Messenger and Calls to full product readiness. The agent operates as team tech lead: discuss only material gray zones, then plan and execute autonomously in bounded cycles with maximum safe agent parallelism.
2. Mail and Calendar are being completed in another branch as a minimalist Apple-like mail/calendar experience with the necessary product functions. This branch must not edit that work. When it is merged to `master`, this branch will merge `master` and then finish cross-module integration.
3. Messenger is a workplace messenger inside the Hub platform. Channels and discussions must have the genuinely necessary Slack-class behavior, including side-panel threads, but the visual design must remain Hub-native rather than copying Slack one-to-one.
4. Calls should not continue as the existing weak custom implementation. A mature open-source conferencing solution should replace its media/conference runtime.
5. Jitsi Meet was selected. The accepted mode is a ready Jitsi distribution customized for Hub, not a cloud call to `meet.jit.si`, not JaaS, and not an external API dependency.
6. The Hub repository owns Jitsi deployment/configuration/branding/adapters and pins versions and image digests. Every Jitsi service runs on infrastructure controlled by the Hub owner.
7. A full deep fork/vendor copy of all Jitsi repositories is not required. The cost advantage comes from using the maintained upstream distribution with a narrow Hub integration boundary.
8. The same self-hosted-only filter applies to future open-source modules. Accepted candidates are Excalidraw, Baserow, Activepieces Community Edition, Apache Superset, ONLYOFFICE Docs, Documenso, and Chatwoot.
9. The chosen UX option for future modules is a native Hub product shell. Specialist upstream surfaces remain embedded where replacing them would amount to rebuilding the product: Jitsi conference, Excalidraw canvas, and ONLYOFFICE editor.
10. Baserow fully replaces native `Bases`; parallel long-term coexistence is rejected. Activepieces fully replaces the corresponding native automation engine. Chatwoot is only the external helpdesk and does not replace internal Messenger.
11. "No API and no other host" means no SaaS/cloud control plane and no runtime hosted by another company. Local internal interfaces between Hub-owned services are allowed and required.
12. The founder authorized saving these decisions and immediately starting Messenger and Calls implementation. The later OSS modules are queued behind the communications work and the shared integration boundary.

## Engineering rulings required to realize the decisions

- LiveKit is not deleted before Jitsi parity and rollback evidence exists. It remains a shadow provider during migration, then is removed in the cutover phase.
- Hub remains the authority for users, tenants, RBAC, audit, navigation, and business lifecycle. Upstream products do not become second identity or policy authorities.
- Components do not share application databases. Each uses its supported store, with explicit local adapters and one writable system of record per business object.
- Runtime egress is denied by default. Enabling any Internet connector is a separate founder-approved change.
- Mail/Calendar protected paths stay untouched until their merge gate.

## Delivery phases

- `149`: Messenger correctness and realtime foundation.
- `150`: self-hosted Jitsi deployment and provider foundation.
- `151-155`: Messenger/Calls completion, cutover, and communications acceptance.
- `156-163`: shared OSS boundary followed by the accepted replacements/integrations.
- `164`: cross-module acceptance after Mail/Calendar reaches `master`.

## Implementation clarifications frozen during plan review

- Messenger send idempotency is keyed by authenticated sender plus `client_nonce`; channel and normalized content are the fingerprint. A retry returns the same message, while nonce reuse for another channel or payload conflicts.
- A nonce retry is resolved before revalidating a thread parent, so a successfully stored reply remains retryable if the parent is later deleted. A new reply still requires a live same-channel root.
- Phase 149 suppresses duplicate side effects but does not pretend Mongo and the current best-effort Redis unread projection are one transaction. Durable read state and projection rebuild are explicit Phase 151 work.
- Every existing `message_sent` producer (normal, forward, bot, poll, promote) migrates to one canonical event payload; reconnect refetch remains the durable recovery path.
- Jitsi baseline is `stable-11146-2` with four immutable multi-architecture OCI index digests. Phase 150 keeps `CALLS_ACTIVE_PROVIDER=livekit`; Jitsi cannot become active until Phase 152.
- The provider boundary returns typed provider/room/service URL/token data while preserving current public LiveKit compatibility fields during the shadow phase.
- Jitsi uses the dedicated `meet.<Hub domain>` origin behind Hub Caddy, local JWT auth with empty/anonymous admission disabled, UDP 10000 media, no removed legacy TCP 4443, no JaaS, no external analytics, and no public STUN/TURN dependency.
- Jitsi runtime, media, rollback, and per-image SBOM evidence remain `NOT_RUN` until an explicitly authorized task-owned deployment; static inspection is not runtime acceptance.
