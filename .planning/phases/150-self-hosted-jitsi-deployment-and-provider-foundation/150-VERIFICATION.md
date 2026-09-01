# Phase 150 Verification Ledger

**Date:** 2026-08-25

**Ledger updated:** 2026-09-01. Underlying receipts retain their exact dates;
this metadata update is not new executable or runtime evidence.

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

Tests assert that neither the value nor its pointer implements `fmt.Stringer`. This is an accidental-stringification guard rather than redaction/security coverage: the UUID fields are exported and still formattable, so operational no-logging remains required. Historical missing-symbol RED had exit `1` and stderr length `972`; no RED container name was retained. Intermediate failures were contaminated by a concurrent Messenger RED, are not product failures, and provide no `PASS` authority. The only canonical final container is `worktree-master-backend-unit-run-01f901840085`: target/build/vet/unit/race `PASS`, exit `0`, and four empty stderr logs; stdout was not inspected. Specification review is `PASS` and quality review `APPROVED`, `C0/I0/M0`.

This is bounded unit/race evidence only. Legacy behavior is unchanged. LiveKit remains active; Jitsi deployment/media/browser/cutover and all runtime claims remain `UNBOUND/NOT_RUN`; the Phase 150 verdict is unchanged.

## Calls connection-aware refcount reducer evidence - 2026-08-29

Commit `519d14b1` adds only a pure internal reducer over an accepted Hub `ConfirmedParticipantConnectionKey`, a trusted confirmed observation, and the current per-user aggregate. It admits joined only on aggregate `0 -> 1` and left only on `1 -> 0`; duplicate connected/left observations are no-ops, while unknown left requests reconciliation only. An active same-epoch connection cannot go `Left -> Connected`; a new epoch is required. Structural validation is performed before ended-state handling and invalid inputs return the exact reducer sentinel plus a zero result. Valid ended-call observations are descriptive no-ops, so ending a call cannot cause an individual-left storm.

RED `worktree-master-backend-unit-run-4ceca6a1e260`: target `FAIL`, first exit `1`, build `PASS`, and expected missing-symbol vet/unit/race failures, retained stderr lengths `0/232/0/0`. Canonical GREEN `worktree-master-backend-unit-run-5f5fbdac76bd`: target/build/vet/unit/race `PASS`, exit `0`, four empty stderr logs. Specification review is `PASS`; quality review is `APPROVED`; `C0/I0/M0`.

This calculation seam adds no I/O, logging, persistence, CAS/transactions, events, reconciliation execution, provider/webhook processing, runtime binding, or Jitsi work; the earlier simple presence reducer and all legacy calls behavior remain unchanged. LiveKit remains active; Jitsi deployment/media/browser/cutover and runtime claims stay `UNBOUND/NOT_RUN`; the Phase 150 verdict is unchanged.

## Calls confirmed participant connection state-change evidence - 2026-08-29

Commit `8274eb0c` adds only the pure internal `BuildConfirmedParticipantConnectionStateChange` builder. It turns a validated reducer persistence transition into immutable expected-to-next connection-state and aggregate-count CAS data. The only applied shapes are exact `Unknown -> Connected` increment and `Connected -> Left` decrement; joined is carried only at aggregate `0 -> 1`, left only at `1 -> 0`. Descriptive no-op/reconcile returns zero payload, false, nil; all malformed or contradictory transitions return exact `ErrConfirmedParticipantConnectionStateChangeInvalid`, zero payload, false.

The private future store port describes an atomic compare-and-update of connection state and aggregate count. `Applied` is legal only after both updates commit; conflict/non-application creates neither mutation nor publication. Neither payload value nor pointer offers Stringer/text/JSON marshaling; tests make this an accidental-interface guard, not redaction or secrecy coverage, because typed getters remain formattable and operations must not log connection values.

RED `worktree-master-backend-unit-run-6b5378a1ffd1`: target `FAIL`, first exit `1`, build `PASS`, expected missing-builder vet/unit/race and stderr `0/245/0/0`. Initial GREEN `worktree-master-backend-unit-run-54f5dc4fa055` preceded review hardening. Final canonical GREEN `worktree-master-backend-unit-run-568e22efb42b`: target/build/vet/unit/race `PASS`, exit `0`, four empty stderr logs; specification review `PASS`, quality review `APPROVED`, `C0/I0/M0`.

No store adapter/schema/transaction, outbox/event, retry/reload, context behavior, provider/webhook, runtime/server/service binding, Jitsi deployment, media, or browser behavior is implemented or verified. LiveKit remains active. Jitsi and all cutover/runtime claims remain `UNBOUND/NOT_RUN`; the Phase 150 verdict is unchanged.

## JITSI-02A provider-event envelope integrity evidence - 2026-09-01

Commit `9f80a550` implements only the provider-neutral in-process HMAC-SHA256 verifier in `backend/internal/calls/provider_event_hmac.go` and its external oracle in `backend/tests/calls/phase150_provider_event_hmac_contract_test.go`. The contract requires one non-nil generic sentinel, a private verified-value struct, fresh-copy `Bytes()`, secret length at least 32, raw-envelope length 1..65536, decoded MAC length exactly `sha256.Size`, exact opaque bytes, explicit `hmac.Equal`, and zero-value plus exact-sentinel rejection. Source review confirms validation order and no parsing, normalization, decoding, decompression, logging, clock, replay state or provider coupling.

Canonical RED `worktree-master-backend-unit-run-c18eaa65d772` exited `1`: build passed; vet/unit/race failed only on absent `Phase150VerifiedProviderEventEnvelope`, `ExportPhase150VerifyProviderEventEnvelopeHMACSHA256` and `ErrPhase150ProviderEventEnvelopeInvalid`. Canonical GREEN `worktree-master-backend-unit-run-6582c1e66044` exited `0`: target/build/vet/unit/race all passed, all four launcher-owned stderr logs were empty, and fail/panic/data-race scans were empty. Image identity for both runs is `sha256:51a4cc1bf7302a4632d0e71a5bec7cb177dc8744f993b1f52b13f138f7b03035`; RepoDigest is absent. Source SHA-256 is `0766FC2AA8628F5F49C92822D9951B696A9EB698F2C3A20B9748D312FF9F48F6`; oracle SHA-256 is `E947651F31EF297514C7AB01DECB26CCA9A7DC60E58D097C283F0032CB848D2B`. Shared-directory `qa-build` and `swagger-check` receipts are dated 2026-08-23 and are not evidence for either JITSI-02A container. Three independent static post-implementation reviews report `PASS C0/I0/M0`.

The runs were serialized and used the founder-amended `2 GiB` free-RAM preflight for this execution while retaining the disk and single-container guards. They do not verify ingress, routes, configuration, provider compatibility, mTLS, delivery identity, replay protection, durable enqueue/application, Jitsi images/deployment, SSO, branding, recording, load, media, browser behavior or cutover. Those rows remain `NOT_RUN/UNBOUND`; the Phase 150 verdict remains `PARTIAL - STATIC_FOUNDATION_ONLY`, and LiveKit remains active.

## CALL-DUR-01 Calls provider/presence schema-source evidence - 2026-09-01

Commit `374c1dd6` adds exactly two required immutable `Call` enum descriptors: `provider` accepts `livekit|jitsi` with default `livekit`, and `presence_mode` accepts `legacy|durable` with default `legacy`. Both are required, non-nillable, non-unique and have no update default. The owner-applied migration `20260830_calls_durable_presence_expand.sql` is a fail-closed one-shot `BEGIN`/`COMMIT` with exactly four plain additive `ALTER TABLE calls ADD ...` actions: two columns and two `CHECK` constraints. It has no catalog/existence guard, `IF NOT EXISTS`, drop, trigger, index or data rewrite. Schema, migration and oracle SHA-256 are `E5C45C4DC2CB9336D21BDF5E300B9CE305BFD3D411E009CCD315090968BE6F5A`, `AE7F670E06C24130994653298CBE3EEE35DE86316D7AE1356A0F19529B0D69D4` and `56CB8686C9C42D5579723454825AE176B96B19D177D40362A82444E0764E032A`.

Canonical RED `worktree-master-backend-unit-run-3edf91c8ac59` exited `1`: build/vet passed and unit/race failed only on the absent frozen descriptors. First GREEN `worktree-master-backend-unit-run-70a339f09db9` is rejected, not acceptance evidence: inserting the descriptors mid-list shifted historical indices against the image-contained generated Ent runtime and caused broad `ent.init.0` panics. The unchanged descriptors were moved to the end of `Call.Fields()`. Final canonical GREEN `worktree-master-backend-unit-run-2cbaf3f11662` exited `0`: target/build/vet/unit/race passed, all four launcher-owned stderr logs were empty, no fail/panic/data-race marker remained, and the named oracle passed once in unit and once in race. The image is `sha256:51a4cc1bf7302a4632d0e71a5bec7cb177dc8744f993b1f52b13f138f7b03035` with no RepoDigest. Independent source/spec/quality acceptance is `C0/I0/M0`.

The exact runs were serialized under the founder-approved `2 GiB` free-RAM floor while retaining the `250 GiB` C: floor, three disk samples, no concurrent Mail/Calendar build or other `backend-unit`, and idle-BuildKit-only rule. This is `SOURCE_ACCEPTED` descriptor/unit evidence only. Generated Ent API/runtime parity and image refresh are `NOT_RUN/BLOCKED_GENERATOR`; SQL parse/application/rollback atomicity, defaults/backfill/old-insert compatibility, provider routing, durable presence behavior, live PostgreSQL, integration and product runtime remain `NOT_RUN/UNBOUND`. The Phase 150 verdict is unchanged and LiveKit remains active.

## CALL-DUR-03 neutral provider-participant identity evidence - 2026-09-01

Commit `73adb82c` adds only the private-field provider-neutral `CallParticipantConnectionKey` and the separate private-field `ProviderParticipantIdentity`. Their exact canonical languages are `hub.call-connection.v1/<call>/<user>/<epoch>` and `hub.call-participant.v1/<call>/<user>/<epoch>`. The identity parser requires exactly 134 bytes, the exact lowercase namespace and exactly three canonical nonzero UUID segments. It intentionally accepts non-v4/variant and repeated nonzero UUIDs. Zero or malformed input returns a zero value plus the exact relevant sentinel. Neither value nor pointer exposes String/Text/JSON interfaces; the existing `ConfirmedParticipantConnectionKey` is unchanged.

Canonical RED `worktree-master-backend-unit-run-a5248a2468d0` exited `1`: build passed and vet/unit/race failed only on missing CALL-DUR-03 symbols, with zero fail-action, panic and data-race markers. Canonical GREEN `worktree-master-backend-unit-run-e44a00d790b5` exited `0`: target/build/vet/unit/race passed, four launcher-owned stderr logs were empty, 12 top-level CALL-DUR-03 tests passed in both unit and race, and all fail-action/panic/data-race counts were zero. Key, identity and oracle SHA-256 are `986A32F99CD21F0A5C65681DA6D5098C81CFE4CC10BA80687198CDFE6628F2F1`, `31D2EECC78CD5786490A134E4A550045758B55A54371030D647E2A725C5AF222` and `980C5982D91344A67D35A3646DD22A385AB3E08F19F3693832929011D88C0DDE`. Independent post-GREEN specification, security and quality reviews are `C0/I0/M0`.

The serialized RED/GREEN runs retained the founder-approved `2 GiB` free-RAM floor, `250 GiB` C: floor, three disk samples, no concurrent Mail/Calendar build or other `backend-unit`, and idle-BuildKit-only rule. This is pure source/unit evidence, not authentication, authorization or connection confirmation. Provider event/room/JWT compatibility, service/store/router wiring, persistence, Jitsi/LiveKit runtime, deployment, browser/media and a current run image RepoDigest remain `NOT_RUN/UNBOUND`. The Phase 150 verdict remains `PARTIAL - STATIC_FOUNDATION_ONLY`; LiveKit remains active.

## CALL-DUR-02A participant-connection schema-source evidence - 2026-09-01

Commit `51399154` adds only `CallParticipantConnection` and its external descriptor oracle. The descriptor has exactly nine direct fields in frozen order, the existing positive default-one `VersionMixin`, two direct child-owned immutable edges, and three indexes: unique call/user/epoch identity, partial terminal `left_at` with exact predicate `state = 'left'`, and non-unique user lookup. Call deletion cascades; User deletion is no-action. No provider, delivery, secret/token, raw event, aggregate, public-presence or outbox data is stored. Schema and oracle SHA-256 are `F03A139DC6CE15B43BFB8AEDA3D456FC3B35CD52DB77AE5178F4FDE70610BDE4` and `41A468D2172B11E8BF4A86CF6FFDD0B8203D2790B95844CF3E84AC69577784A0`.

Canonical RED `worktree-master-backend-unit-run-6bb1463ea563` exited `1`: build passed and vet/unit/race failed only on the absent schema type. Canonical GREEN `worktree-master-backend-unit-run-fe3460f692e5` exited `0`: target/build/vet/unit/race passed, all four launcher-owned stderr logs were empty, and the named oracle passed once in unit and once in race with empty fail/panic/data-race scans. RED/GREEN preflights used the `2 GiB` RAM and `250 GiB` C: floors, three disk samples, no other `backend-unit` or Mail/Calendar build, and no idle-BuildKit intervention.

Root filesystem review plus executable Docker evidence passed. Two independent contract/data-model reviews over the exact root transcription returned `C0/I0/M0`; their source provenance is `INPUT_TRANSCRIBED_BY_ROOT`, not filesystem-verified, because reviewer reads were ACL-blocked. The result is only `SOURCE_ACCEPTED`: generated Ent API/runtime and current image provenance are `NOT_RUN/BLOCKED_GENERATOR/UNBOUND`; PostgreSQL timestamp truth/checks remain `CALL-DUR-02M`. Persistence, transitions, retention, ingress, routing, Jitsi/LiveKit runtime, deployment and browser/media remain `NOT_RUN/UNBOUND`. The Phase 150 verdict remains `PARTIAL - STATIC_FOUNDATION_ONLY`; LiveKit remains active.

## CALL-DUR-02B participant-presence schema-source evidence - 2026-09-01

Commit `c717625e` adds only `CallParticipantPresence` and its external descriptor oracle. The descriptor has exactly seven direct fields in frozen order, no mixin, two direct child-owned immutable edges, unique call/user identity and a non-unique user lookup. Both mutable signed-int64 counters default to zero and reject negatives. Call deletion cascades; User deletion is no-action. No connection epoch/state, provider, delivery, legacy participant, outbox or event field is stored. Schema and oracle SHA-256 are `026914D81E224C2B0D3492FEB7AA9FE7927BE643BA9D3388336C932533275B61` and `96231C8C178699B2967DD8FD6295C89D3000DFD279C1516ED7BDEB666584131C`.

Canonical RED `worktree-master-backend-unit-run-ff7b47d55060` exited `1`: build passed and vet/unit/race failed only on the absent schema type. Canonical GREEN `worktree-master-backend-unit-run-ac7663bf7081` exited `0`: target/build/vet/unit/race passed, all four launcher-owned stderr logs were empty, and the named oracle passed once in unit and once in race with empty fail/panic/data-race scans. RED/GREEN used the `2 GiB` RAM and `250 GiB` C: floors, three disk samples, no other `backend-unit` or Mail/Calendar build, and no idle-BuildKit intervention.

Static/TDD transcript reviews returned `C0/I0/M0`. The data-model `M1` is closed in the specification: `connected_count` and `presence_sequence` are materialized projection fields only. A later adapter must atomically couple connection/aggregate mutation, prevent counter underflow and lost updates, and enforce conditional monotonic sequence/replay handling; absence of `VersionMixin` provides no schema-level CAS claim. This result is only `SOURCE_ACCEPTED`; generated Ent/runtime, current image provenance, PostgreSQL behavior, adapter concurrency, ingress/routing and Jitsi/LiveKit runtime remain `NOT_RUN/BLOCKED_GENERATOR/UNBOUND`. The Phase 150 verdict remains `PARTIAL - STATIC_FOUNDATION_ONLY`; LiveKit remains active.

## CALL-DUR-02C provider-delivery schema-source evidence - 2026-09-01

Commit `0a0ec795` adds only `CallProviderDelivery` and its external descriptor oracle. The ten-field descriptor freezes provider-scoped replay identity, normalized kind/call/epoch facts, a GDPR-erasable nullable user, pending/applied/ignored/reconcile result, received/terminal timestamps, no mixin and no edges. Its only indexes are unique provider/delivery dedupe, result/terminal cleanup and user lookup. No raw/security/credential/room/error/trace/outbox field is stored. Schema and corrected oracle SHA-256 are `41219FA765D92692B5664E9C4B5FD69E169748B114FD933D65C98E7C3E9631FE` and `918597670D4D9A962312F5612BE22021E163CFFF17BAF8148408E3E3573EAC11`.

Canonical RED `worktree-master-backend-unit-run-39b4a4226df2` exited `1`: build passed and vet/unit/race failed only on the absent schema type. First GREEN `worktree-master-backend-unit-run-f8915f185a7a` is rejected: an internally contradictory oracle helper rejected the required `pending` default before checking its exact value. The discovered/draft one-line correction derives default presence from `wantDefault != nil`; production did not change. Final GREEN `worktree-master-backend-unit-run-21218f718a4e` exited `0`: target/build/vet/unit/race passed, all four stderr logs were empty, and the named oracle passed in unit and race without fail/panic/data-race markers. Runs were serialized behind the `2 GiB` RAM and `250 GiB` C: floors; a one-sample RED disk-queue spike was followed by three zero samples before launch.

Security boundaries are frozen for future adapters: dedupe is not authenticity; ingress authenticates before insert and stores no verification material; conflicting immutable facts for an existing tuple fail closed without overwrite and create a content-safe audit signal; a terminal GDPR tombstone is never relinked by replay; user mutation is service-authorized; result/user/terminal updates are conditional expected-state transitions because no `VersionMixin` exists. Nonzero UUID and result/terminal/user DB truth remains `CALL-DUR-02M`. This is only `SOURCE_ACCEPTED`; generated Ent/runtime, current image provenance, SQL application, replay/GDPR behavior, provider authentication, concurrency and product runtime remain `NOT_RUN/BLOCKED_GENERATOR/UNBOUND`. The Phase 150 verdict remains `PARTIAL - STATIC_FOUNDATION_ONLY`; LiveKit remains active.

## CALL-DUR-02M owner migration source evidence - 2026-09-01

Commit `e9a81ba4` extends the one-shot owner migration with all three durable tables before its sole `COMMIT`; SHA-256 is `CBD4A269F86F18A57884FD00B0F61B69A9A5A9B0499D5BC6C919C3BED3E484CB`. The source has one `BEGIN`, one `COMMIT`, three `CREATE TABLE`, three unique and five non-unique indexes. It has no `IF NOT EXISTS`, catalog guard, drop, backfill, nested transaction, compensating statement or update trigger. `git diff --check` reported no whitespace error; its LF-to-CRLF message was an advisory only.

Static inspection covers applicable nonzero UUIDs; the exact connection state/timestamp truth table plus timestamp ordering; positive version; nonnegative presence count/sequence; closed bounded provider, kind and result values; nonempty max-255 opaque delivery identity; and exact delivery result/terminal/user truth. Connection/presence use Call Cascade and User NoAction FKs. Delivery has no FK or `call_id` index and has exactly unique `(provider,delivery_id)`, non-unique `(result,terminal_at)` and non-unique `(user_id)` indexes. SQL, schema-parity and security reviews over the exact root transcription are `C0/I0/M0`; this is transcript provenance because the reviewer filesystem ACL denied direct reads.

The result is only `MIGRATION_SOURCE_ACCEPTED`; no unit oracle or Docker run belongs to 02M. PostgreSQL parse/apply/rollback/replay, generated clean-DB parity, catalog shape, constraint/index enforcement, User-first transactional GDPR cleanup, ingress/delete concurrency, trusted-provider byte-exact delivery-ID behavior and runtime remain `NOT_RUN`. Delivery identity is opaque and must not be trim/case/Unicode-normalized by later ingress. The Phase 150 verdict remains `PARTIAL - STATIC_FOUNDATION_ONLY`; all accepted 01/02 schemas remain `BLOCKED_GENERATOR` until `ENT-CALL-GEN-01`.

## ENT-CALL-GEN-01 image-contained generation evidence - 2026-09-01

The accepted source is clean HEAD `5b08e9aeea05251b83e62a20b0fc85296c40711e`. Before build, host `backend/ent` had exactly 91 source files (`generate.go` plus `schema/**`) and none of the generated sentinels. The complete 96-file schema/generator/module/Docker/Compose input set had aggregate SHA-256 `7AF9CDA7D975F3D74E3A438C9CA65599FEC6AC429E4B89460FD7A87F7A41D5B7`, and that fingerprint matched immediately before the canonical run.

One exact Compose build exited `0`: the backend copy and `go generate ./ent` layers were not cached, generation completed, and the three expected generated sentinel files passed the Dockerfile check. The resulting local image is `sha256:cbd0b9f06bc7e499f9dd1a8a9973e4981b3e75b87f207fda259118c62101b0b9`, with no `RepoDigest`. One exact disposable `backend-unit` run without selector/build/pull exited `0`: target/build/vet/unit/race all passed with `first_exit=0`, four empty stderr logs, zero fail/panic/data-race markers and unit/race PASS actions for the provider/mode, connection, presence and provider-delivery descriptor oracles. Container `worktree-master-backend-unit-run-e0bb8f6add37` was removed; result SHA-256 is `622DFACDD94271805EB680200E153F7C2614D591B1447B8A18A25C972279B1E2`.

All three resource windows passed the founder-amended `2 GiB` RAM, `250 GiB` C:, two-of-three disk and single-heavy-operation policy. The only pre-run queue spike was one `0.13/4` sample and therefore did not meet the defer rule. Final build, provenance and resource reviews are `C0/I0/M0`. `ENT-CALL-GEN-01` is `ACCEPTED`, removing the generation blocker from the accepted 01/02 source atoms. It proves local image generation and this backend-unit/race slice only; registry/remote/CI identity, generated clean-DB DDL parity, migration application, live PostgreSQL and product runtime remain `NOT_RUN/UNBOUND`.

## CALL-DUR-02G transactional GDPR serialization evidence - 2026-09-01

The frozen external oracle was restored byte-for-byte from its independently reviewed draft at SHA-256 `6A02B50AFB647A9DF75F56ACD2ED3017F90B352B246D3C6AEAC2B4DE18554C00`. Both oracle copies matched before RED. The source HEAD for both exact Docker runs was `9953d0974af600fd3e671d07b6c3bbb265d38461`; RED had only the tracked oracle plus the 02G plan clarification, and no production 02G method existed. Both runs reused local generated image `sha256:cbd0b9f06bc7e499f9dd1a8a9973e4981b3e75b87f207fda259118c62101b0b9` and executed only `docker compose -f docker-compose.local.yml run --rm backend-unit`, without selector, build or pull.

Canonical RED container `worktree-master-backend-unit-run-1a1f47028d19` exited `1`. Build passed. Vet, unit and race failed only because `audit.Repository` lacked `LockUserForGDPR` at oracle lines 589 and 620 and `EraseUserDurableCallData` at lines 631 and 726. RED receipt SHA-256 values were target `CD272C102C5467C4A4A5473230AA40AD0B200FD3727D538FED8853F5E7997702`, build `619C5D79833A7F0B94A6759C97028034E7FF7D6057BB03701472E87637FB3B43`, vet `86252A3ACDB815B0D6E742B6B74E96B92E8226903450236AFFFF9B30BC75CDA2`, unit `E31125D9B04855695999EBC205D04EE1A8AAE0A62304454101CF0C8F8DC31A45`, race `36898848F2AC092C07698BAB4472026520FD00CBD3595CA0ABC086BFFDECE068`, and vet stderr `4C5CF1EE4D62590FEA60F93783D01ADE2543FEBDCF2A84716BB35A918E5989D4`. The flat launcher directory was then intentionally overwritten by the mandatory GREEN run, so these RED hashes and diagnostics are the durable receipt; the superseded raw RED log files are not claimed as retained artifacts.

The implementation adds only the two repository methods, the first-callback locked-role recheck plus one UTC durable erase, the focused oracle, and mechanical `mockAuditRepo` compatibility methods. Current source SHA-256 values are repository `E752E7ECD0CEC4B10CBCC0C014DCBBDEE8C86F5803FFD0A07682FD59BC0AF18F`, service `4ABBD5C3983756FCA53BFA7C8BDD49E1D4CC2889ED316F57CD57FF6FD316790D`, oracle `6A02B50AFB647A9DF75F56ACD2ED3017F90B352B246D3C6AEAC2B4DE18554C00`, and test support `C2A60B963448B52A716A748592D1E32779EF1DF0649180972EA5C9B90AC815F3`.

Canonical GREEN container `worktree-master-backend-unit-run-14226df38bc9` exited `0`: target/build/vet/unit/race all passed, all four stderr logs were empty, both top-level 02G tests and their subtests passed under unit and race, and fail-action/panic/data-race marker counts were zero. GREEN receipt SHA-256 values are target `622DFACDD94271805EB680200E153F7C2614D591B1447B8A18A25C972279B1E2`, build `619C5D79833A7F0B94A6759C97028034E7FF7D6057BB03701472E87637FB3B43`, vet `1899DDA4E0728FECC8EB49507B3D050B5A152F305142CC8498AE5F60883F3CFF`, unit `786ADAF5C0B95DAB3D8C29688A1A154DC529C0FAC36313D816DFB0EB700A0046`, and race `B1CC9B3C42E8E9EB0C29CA38828A14B281084F502B2D0AF18FF0CE642AF08CAE`. Stale 2026-08-23 `qa-build` and `swagger-check` files are excluded.

RED preflight recorded `11611012 KiB` free RAM, `325922562048` C: bytes and disk/queue `8.99/0, 0.86/0, 0.21/0`. GREEN preflight recorded `10840524 KiB`, `325899026432` bytes and `1.27/0, 0.37/0, 0.17/0`; post-GREEN recorded `10260308 KiB`, `325502722048` bytes and `0.13/0, 0.96/0, 2.47/0`. Every window passed the founder-approved `2 GiB` RAM and `250 GiB` C: floors, two-of-three disk rule and single-heavy-operation policy. Only idle BuildKit was present; no Mail/Calendar build or other QA runner overlapped, and both disposable containers were removed.

Independent code, security/concurrency/privacy and final evidence reviews are `C0/I0/M0`; the evidence packet is accepted as `SOURCE_ACCEPTED` in commit `64e68da443ad0e842aeec300b4c206e6eb89e102`. Real PostgreSQL lock blocking/deadlock and commit outcome, migration/constraint enforcement, ingress-versus-GDPR concurrency, real Mongo behavior, integration/production runtime and remote/CI provenance remain `NOT_RUN/UNBOUND`; no such claim is inferred from the fake-dialect unit.

## CALL-DUR-08A rejected initial RED - 2026-09-01

Source HEAD was `249dc53ed33ac63a72c473ad0a02628c213506da`; production 08A symbols were absent. Initial oracle SHA-256 `AB4D41642ED3B0445FDCF96FBCBB44C7E1833542CC271E5688DB6E53DDF7EE2B` is rejected. Exact container `worktree-master-backend-unit-run-9824f75c8109` exited `1`: build passed, while vet/unit/race failed on the expected absent `PersistedCallProvider` API and two unrelated oracle compile defects at lines 136/140, where `phase150CallProviderColumns` returned one value from a `([]string,error)` function. Therefore this run is test-triage evidence only and is not canonical RED.

Rejected result SHA-256 values are target `CD272C102C5467C4A4A5473230AA40AD0B200FD3727D538FED8853F5E7997702`, build `619C5D79833A7F0B94A6759C97028034E7FF7D6057BB03701472E87637FB3B43`, vet `86252A3ACDB815B0D6E742B6B74E96B92E8226903450236AFFFF9B30BC75CDA2`, unit `E31125D9B04855695999EBC205D04EE1A8AAE0A62304454101CF0C8F8DC31A45`, race `36898848F2AC092C07698BAB4472026520FD00CBD3595CA0ABC086BFFDECE068`, vet stderr `32A7EE0579B05B19CDA3D9A174669FDBF87C6A3A7EAEDA092BD471D40A043B17`, unit stdout `F9FEF50904970C6C769CA2234AFC85EA5629B7A96FEB479768DC12A1CD86C007` and race stdout `DA3CA6AE845DE20A1E16B05BA1E2F14E787D07FDAC35D09CC3E1A01F6C444E3E`. Preflight recorded `10461656 KiB` free RAM, `325290971136` C: bytes and disk/queue `0.30/0, 1.48/0, 0.42/0`; only idle BuildKit ran and the disposable container was removed. Read-only in-memory reconstruction independently derived the arity-corrected SHA `DB7DEE16D68E80EFAB07407FCD042BE3AD37CF8BC5624BFAE1579AE5D490208A` and then initial SHA `AB4D41642ED3B0445FDCF96FBCBB44C7E1833542CC271E5688DB6E53DDF7EE2B` from the current file without writing artifacts; evidence re-review is `C0/I0/M0`. Two fail-closed padded-provider cases then produced the final frozen oracle SHA `E4622E9BEE982FBB9024EAFD3095E829E273FC41D66CD2DD96229504A0C6AAD3`, accepted by fresh compile/contract and security reviews at `C0/I0/M0`.

## CALL-DUR-08A canonical RED - 2026-09-01

The accepted source HEAD is `249dc53ed33ac63a72c473ad0a02628c213506da` with only the frozen oracle and Phase-150 planning/evidence edits; no 08A production symbol existed. Final oracle SHA-256 is `E4622E9BEE982FBB9024EAFD3095E829E273FC41D66CD2DD96229504A0C6AAD3`. Exact no-selector/no-build/no-pull command `docker compose -f docker-compose.local.yml run --rm backend-unit` created `worktree-master-backend-unit-run-1fc98ce666ad` and exited `1`. Build passed. Vet, unit and race failed only because `PersistedCallProvider` and its dependent frozen API were absent; the earlier return-arity defects were absent. Fail-action counts were the expected one external test-package build failure in unit and race, with zero panic and data-race markers.

Canonical RED result SHA-256 values are target `CD272C102C5467C4A4A5473230AA40AD0B200FD3727D538FED8853F5E7997702`, build `619C5D79833A7F0B94A6759C97028034E7FF7D6057BB03701472E87637FB3B43`, vet `86252A3ACDB815B0D6E742B6B74E96B92E8226903450236AFFFF9B30BC75CDA2`, unit `E31125D9B04855695999EBC205D04EE1A8AAE0A62304454101CF0C8F8DC31A45`, race `36898848F2AC092C07698BAB4472026520FD00CBD3595CA0ABC086BFFDECE068`, vet stderr `32A7EE0579B05B19CDA3D9A174669FDBF87C6A3A7EAEDA092BD471D40A043B17`, unit stdout `5D391765B0C60A30BB6E9608BAB5DC6EEA0331E11AC017C5BD6C1876111268C9` and race stdout `DD2E9D20586C81190AC2EA33AB5C796B982196279584345FBD6DC2AD1AF5080A`. Preflight recorded `8203088 KiB` free RAM, `325272813568` C: bytes and disk/queue `2.40/0, 1.51/0, 2.74/0`. Only idle BuildKit existed, no Mail/Calendar build or other QA runner overlapped, and the disposable container was removed.

## CALL-DUR-08A canonical GREEN - 2026-09-01

Production source SHA-256 values are `models.go=58882E2D1F223F8B8D8639A9D885733AE35F944C8B70136914C26112EC13512D` and `repository.go=4F4E3AF23B62701D9EF56935BC653B32FF4AB396362F57302098A05477E14906`; frozen oracle SHA remained `E4622E9BEE982FBB9024EAFD3095E829E273FC41D66CD2DD96229504A0C6AAD3`. The implementation adds only the closed persisted-provider type/constants/sentinel and exact fail-closed provider-plus-room repository lookup. Typed-nil reflection is inlined with nilable-kind guards; no validation helper, fallback, status/tenant/channel/deleted predicate, error wrapping or other public API exists. Independent pre-GREEN code, security/authority and Ent/Go reviews are `C0/I0/M0`.

Exact no-selector/no-build/no-pull command `docker compose -f docker-compose.local.yml run --rm backend-unit` reused local generated image `sha256:cbd0b9f06bc7e499f9dd1a8a9973e4981b3e75b87f207fda259118c62101b0b9`. Container `worktree-master-backend-unit-run-46c76cc67313` exited `0`: target/build/vet/unit/race all passed, all four stderr logs were empty, global fail/panic/data-race counts were zero, and the 08A oracle recorded 24 PASS actions with zero FAIL in each of unit and race. Those actions include all invalid/no-query cases, both providers across active/ended rows, Ent not-found/not-singular and unwrapped driver-error identity.

GREEN result SHA-256 values are target `622DFACDD94271805EB680200E153F7C2614D591B1447B8A18A25C972279B1E2`, build `619C5D79833A7F0B94A6759C97028034E7FF7D6057BB03701472E87637FB3B43`, vet `1899DDA4E0728FECC8EB49507B3D050B5A152F305142CC8498AE5F60883F3CFF`, unit `786ADAF5C0B95DAB3D8C29688A1A154DC529C0FAC36313D816DFB0EB700A0046`, race `B1CC9B3C42E8E9EB0C29CA38828A14B281084F502B2D0AF18FF0CE642AF08CAE`, unit stdout `50FA2A2A47F1CF2E3D8D09459CB8FF03CE1BC69758291F9E0BD9B3C2402EE24E` and race stdout `D091B22994692E6D9BFCC4727749152B2C210FB47A21E5FCC58C47872F45198C`; each stderr SHA is the empty-file digest `E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855`. GREEN preflight recorded `9951960 KiB` free RAM, `325087449088` C: bytes and disk/queue `4.87/0, 0.37/0, 6.89/0`; post-run recorded `8995400 KiB`, `325085229056` bytes and `1.48/0, 0.10/0, 7.66/0`. Every window passed the `2 GiB` RAM, `250 GiB` C:, two-of-three disk and single-heavy-operation rules; only idle BuildKit existed, no Mail/Calendar or other QA overlapped, and the disposable container was removed.

This is repository source/unit/race evidence only. Historical migration/default/backfill persistence, real PostgreSQL planner/cardinality/error behavior, trusted provider ingress/authentication, delivery identity/replay, provider selection, routing, callbacks, service/handler/wiring, Jitsi activation/deployment/media/browser and product runtime remain `NOT_RUN/UNBOUND`. Final independent code, security/authority and evidence reviews are `C0/I0/M0`; the packet is `SOURCE_ACCEPTED / CLEAR_TO_COMMIT` and exact implementation/oracle commit is `eae39e7df9f6160b591e3dd36a798564a19bd7c6`.
