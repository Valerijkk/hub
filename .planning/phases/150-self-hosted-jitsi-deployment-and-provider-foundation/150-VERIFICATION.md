# Phase 150 Verification Ledger

**Date:** 2026-08-25

**Verdict:** `PARTIAL — STATIC_FOUNDATION_ONLY`

Phase 150 is not complete, production-ready, a deployed self-hosted Jitsi stack, or media-ready. LiveKit remains the only active and rollback provider.

## Executable evidence

| Boundary | Command / receipt | Result | Proven scope |
|---|---|---|---|
| Backend final gate | `docker compose -f docker-compose.local.yml run --rm backend-unit` | `PASS` / exit `0` | Provider-neutral contract, LiveKit adapter parity, service injection, fail-closed provider/config factory, and internal server-only Jitsi issuer. |
| Frontend RED | `docker compose -f docker-compose.local.yml run --rm frontend-unit` | Expected `FAIL` / exit `1`; TypeScript `PASS`, Vitest `FAIL` on unresolved `@/modules/calls/jitsi-external-api` at test line 3 | Genuine missing-module TDD RED. |
| Frontend final gate | `docker compose -f docker-compose.local.yml run --rm frontend-unit` | Overall `PASS`; TypeScript `PASS`; Vitest `PASS`; exit `0` | Pure canonical self-hosted origin to `/external_api.js` URL adapter only. |

The initial backend RED was genuine because the frozen provider/config/issuer APIs did not exist. The first intended GREEN run then failed only in the LiveKit test oracle: it decoded generic map claims using `room_join` and expected `iat`, while the pinned LiveKit contract uses typed `VideoGrant.RoomJoin` and `nbf` plus `exp`. Independent triage classified this as test-only. The oracle was changed to `auth.ParseAPIToken(...).Verify(...)`, independently reviewed, and the final backend gate passed.

The frontend unit does not prove that `external_api.js` exists inside a Jitsi image, that a script can be loaded, or that any Jitsi UI/provider path is active.

## Static evidence

| Artifact / review | Result | Boundary |
|---|---|---|
| Provider/config/factory/issuer reviews | `PASS` | Source and contract review only; no public Jitsi route or cutover. |
| `deployments/jitsi/DEPLOYMENT-PREFLIGHT.md` | `PASS C0/I0/M0` | Static preflight only. |
| Frontend adapter final review | `PASS C0/I0/M0` | Pure URL calculation only. |
| Phase 150 Task 10 docs review | `PASS C0/I0/M0` | Documentation truth and links only. |

## Registry metadata

The recorded release is `stable-11146-2` with these immutable registry references:

- `ghcr.io/jitsi/web@sha256:4231d5ff7318ff71d12ffd41afc36adae3c319b40cf79cde39b009b0c68295dc`
- `ghcr.io/jitsi/prosody@sha256:9e341b10134d947e49ca338b9efe8d70ac9d6e8d89d621da2246dfa79fe50363`
- `ghcr.io/jitsi/jicofo@sha256:a8d64e46eaa340751a2b5530a4abd4175adf05bff07e9775fc48daae6364cfe9`
- `ghcr.io/jitsi/jvb@sha256:d7dc646d21af07530ae6acf1ace23ea4782445d0362004621a32a5a632a964e9`

These are registry metadata records only. No Jitsi image was pulled, built, inspected, rendered, or started. Image content, licenses, SBOMs, and runtime properties were not verified.

## Unbound and blocked evidence

| Evidence row | Status |
|---|---|
| Compose overlay | `NOT_CREATED / UNBOUND` |
| Task 8 authorized Linux discovery | `BLOCKED_BY_AUTHORIZED_LINUX_DISCOVERY` |
| Container users / UID / GID | `UNBOUND` |
| Writable paths and configuration ownership | `UNBOUND` |
| Rootless/read-only/tmpfs posture | `UNBOUND` |
| Branding result | `NOT_RUN` |
| Effective `external_api.js` asset | `NOT_RUN` |
| SBOM and license verification | `NOT_RUN` |
| Compose render / start / health | `NOT_RUN` |
| Caddy / TLS routing | `NOT_RUN` |
| Jitsi JWT integration | `NOT_RUN` |
| JVB advertise IP / UDP / media | `NOT_RUN` |
| Security runtime checks | `NOT_RUN` |
| Restart and rollback rehearsal | `NOT_RUN` |

No external meeting API, SaaS host, external STUN/TURN, analytics endpoint, Jitsi image pull, image build, Compose render, or Jitsi service start was used.

## Continuation gate

Any deployment continuation requires explicit infrastructure authority, an authorized Linux discovery packet, and a numeric disk budget. Images may then be handled one immutable digest at a time with disk measurements before and after; parallel pulls and builds remain forbidden.

## Calls Ent projection source evidence - 2026-08-28

`*EntRepository` now satisfies `CallJoinCredentialProjectionSource` through exact active-call and active-user queries, with trimmed `DisplayName` to `Name` fallback and fail-closed receiver/client/plain-or-typed-nil-context/UUID guards. The interface RED kept build `PASS` and failed vet/unit/race only on missing methods. A review-driven typed-nil RED kept build/vet `PASS` and reproduced unit/race Ent-client panics. The final fresh isolated Docker run passed target/build/vet/unit/race with first exit `0`; both named suites passed in unit and race without Mail overlap.

This is source/unit evidence only. The adapter is production-unwired; positive Ent predicate/fallback semantics and live PostgreSQL are `NOT_RUN`. Service/HTTP/factory/provider/client, Jitsi deployment/cutover, browser/media runtime, and the active LiveKit path are unchanged. Calls/Jitsi readiness is not established by this atom.
## Calls configured credential-service composition evidence - 2026-08-28

Commit `07f2fb5c` statically composes the existing server Calls service with `NewCallServiceWithConfiguredProvider`: configuration selects the factory provider, the provider supplies the typed issuer, the issuer and `callsRepo` construct the projector, and the projector enters the additive facade. Server wiring passes `callsRepo` twice by design: the first use is the Calls repository and the second is the projector source. A configuration/factory/projector error exits before `callsChatRepo.EnsureIndexes`.

Valid RED `46b078978d66` preceded GREEN `4d2f71546884` at `2026-08-28T22:39:54+03:00`; target/build/vet/unit/race are root-read `PASS`, exit `0`, and stderr is empty. `qa-build` and swagger check `PASS` are retained implementer receipts, not independently refreshed or root-verified receipts; code review found the implementation conformant and required only this documentation sync.

This does not change public `JoinCall`, routes, DTOs, frontend, recording, or the legacy token path. `cmd/server` runtime is `NOT_RUN`; Jitsi deployment/media/browser and provider cutover remain `UNBOUND/NOT_RUN`. LiveKit remains active and Jitsi remains rejected before Phase 152; this atom does not advance the Phase 150 verdict.


## Calls confirmed participant-presence reducer evidence - 2026-08-28

Commit `f405304f` adds a pure future canonical presence reducer for already authenticated, provider-normalized connection observations. Canonical lifecycle remains `authorized -> credential_issued -> connected -> left`; issuance is not presence and never publishes joined. `Unknown|Left -> Connected` persist/publish joined, duplicate connected is a no-op; `Connected -> Left` persists/publishes left, while `Unknown -> Left` and duplicate left are no-ops. An unrecognized/out-of-range previous enum, zero IDs, and invalid observation kind return `ErrConfirmedParticipantObservationInvalid` plus zero transition. It provides state-based idempotence only, not durable dedupe or exactly-once behavior.

RED `worktree-master-backend-unit-run-3234f1a2a70b`: target `FAIL`, first exit `1`, build `PASS`, vet/unit/race missing exported symbols; retained implementer stderr lengths `0/204/0/0`. Root-read GREEN `worktree-master-backend-unit-run-af958086c4a2`: target/build/vet/unit/race `PASS`, first exit `0`, four stderr lengths `0`. Specification and quality reviews are `PASS/APPROVED`, `C0/I0/M0`.

This unit-only pure seam does not bind raw webhook verification, provider selection, persistence, NATS, inbox/outbox, transactional application, clocks, delivery IDs, multi-session, guest, refresh/revoke, API, service, repository, schema, server, or frontend. Legacy `/join`, DTO/token/frontend/recording remain unchanged. LiveKit remains active; Jitsi deployment/media/browser/cutover is `UNBOUND/NOT_RUN`; the Phase 150 verdict is unchanged.

## Calls trusted-provider participant-presence mapper evidence - 2026-08-29

Commits `2b63761c` and `a0d3c832` add the pure `NormalizeTrustedProviderParticipantConnection` seam and its trust-boundary Godoc. Its input is already provider-event-authenticated and normalized, with room resolved to call and identity resolved to Hub user. Trusted connected/disconnected kinds map to confirmed Connected/Left observations. Zero call/user IDs or a zero/unknown kind return the zero observation with the existing invalid-observation sentinel.

RED `worktree-master-backend-unit-run-e8e44e800277` was target `FAIL`, first exit `1`, build `PASS`, and vet/unit/race missing exports; retained stderr lengths were `0/222/0/0`. Root-read GREEN `worktree-master-backend-unit-run-72b909a17791`, followed by fresh Godoc-fix GREEN `worktree-master-backend-unit-run-80fe0c688984`, records target/build/vet/unit/race `PASS`, first exit `0`, and four empty stderr logs. Final specification and quality review is `PASS/APPROVED`, `C0/I0/M0`.

The mapper does not accept raw events, room, identity, JWT, provider, token, or delivery ID; invoke the reducer; verify webhooks; apply/publish presence; or use context, clock, network, repository, NATS, membership, provider selection, or session identity. It creates no wiring or side effects. Legacy `/join`, DTO/token/frontend/recording are unchanged. LiveKit remains active; mapper runtime/wiring, reducer invocation, webhook application, persistence/NATS, durable identity/order/dedupe, and Jitsi deployment/media/browser/cutover remain `UNBOUND/NOT_RUN`; the Phase 150 verdict is unchanged.

## Calls confirmed participant connection key evidence - 2026-08-29

Commit `b7695ddc` adds only pure `ConfirmedParticipantConnectionKey`, the Hub-owned `(CallID, UserID, Epoch)` value for a connection fact already trusted upstream. Its exact canonical grammar is `hub.call-connection.v1/<call>/<user>/<epoch>`. A zero component yields `ErrConfirmedParticipantConnectionKeyInvalid` plus the zero key; an invalid key has an empty canonical value. No epoch generator, parser, serializer, provider/webhook, persistence/refcount, service/server/runtime, or Jitsi wiring is included.

Tests assert that neither the value nor its pointer implements `fmt.Stringer`. This is an accidental-stringification guard rather than redaction/security coverage: the UUID fields are exported and still formattable, so operational no-logging remains required. Historical missing-symbol RED `worktree-master-backend-unit-run-01f901840085` had exit `1` and stderr length `972`; intermediate failures were contaminated by a concurrent Messenger RED and are not product failures. The canonical final run under that receipt records target/build/vet/unit/race `PASS`, exit `0`, and four empty stderr logs; stdout was not inspected. Specification review is `PASS` and quality review `APPROVED`, `C0/I0/M0`.

This is bounded unit/race evidence only. Legacy behavior is unchanged. LiveKit remains active; Jitsi deployment/media/browser/cutover and all runtime claims remain `UNBOUND/NOT_RUN`; the Phase 150 verdict is unchanged.
