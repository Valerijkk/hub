# Self-Hosted Jitsi Static Foundation and Provider Boundary Plan

> **Execution rule:** Implement with `superpowers:subagent-driven-development`. Production tasks are sequential Terra/XHigh atomic writes with Terra/High independent review. No two writers own the same file concurrently.

**Phase goal:** Establish a provider-neutral Calls boundary, preserve LiveKit as the only active provider, add a server-only Jitsi credential primitive, and freeze a self-hosted Jitsi deployment preflight. This phase does not by itself deliver an accepted or production-ready Jitsi deployment.

**Current phase verdict:** `PARTIAL — STATIC_FOUNDATION_ONLY`. The overlay is `NOT_CREATED / UNBOUND` until the separately authorized post-discovery gate proves image users, UID/GID requirements, writable config paths, and branding/asset paths on Linux. Compose render, image pull/start, TLS, JWT integration, UDP media, restart, rollback, effective `external_api.js`, and SBOM remain `NOT_RUN`. Planning, env contracts, or draft YAML must not increase Jitsi deployment readiness beyond the already reported static-foundation baseline.

**Active provider:** `CALLS_ACTIVE_PROVIDER=livekit`. Selecting `jitsi` fails closed until Phase 152. LiveKit public DTO fields, webhook, `/lk/*`, recording/Egress, dependencies, and rollback assets remain intact.

**Authoritative inputs:** `150-SPEC.md`, immutable `150-RESEARCH.md`, `.planning/decisions/ADR-COMMS-OSS-REPLACEMENT-2026-08-24.md`, the [official Jitsi Docker guide](https://jitsi.github.io/handbook/docs/devops-guide/devops-guide-docker/), [official docker-jitsi-meet repository](https://github.com/jitsi/docker-jitsi-meet), and official Jitsi iframe API documentation. `150-RESEARCH.md` is read-only evidence and is not owned by any implementation task.

## Global constraints

- Official release is exactly `stable-11146-2`; any eventual core image reference must equal the four immutable OCI index digests already recorded in `150-RESEARCH.md`.
- No JaaS, `meet.jit.si`, external Calls API/host, floating image tag, shared Hub database, external STUN/TURN, external analytics, or upstream source fork.
- `JITSI_HOST` is required when enabled and follows exact ASCII DNS grammar: total length 1–253; regex `^meet\.(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?))*$`; every label is 1–63 characters. IP literals, underscore, wildcard, uppercase, leading/trailing hyphen, trailing dot, `meet.jit.si`, any value ending `.meet.jit.si`, explicit port, scheme, path, query, fragment, and userinfo are rejected. Canonical origin is exact string equality `JITSI_PUBLIC_URL="https://" + JITSI_HOST`; URL normalization must not hide a mismatch.
- One XMPP subject source exists: backend `JITSI_XMPP_DOMAIN`; eventual overlay passes `XMPP_DOMAIN=${JITSI_XMPP_DOMAIN:?required}`; JWT `sub` equals the same value. An independent `XMPP_DOMAIN` input is forbidden.
- `JITSI_ISSUER` and `JITSI_AUDIENCE` are deprecated/forbidden inputs: when either environment variable is nonempty, configuration fails closed instead of ignoring it. JWT `iss=JITSI_APP_ID`; `aud` is exactly the JSON array `["jitsi"]`.
- `JITSI_APP_SECRET`, `JICOFO_COMPONENT_SECRET`, `JICOFO_AUTH_PASSWORD`, `JVB_AUTH_PASSWORD`, and `JITSI_EVENT_HMAC_SECRET` are separate required secrets with no usable default. Hub `JITSI_APP_SECRET` maps explicitly to upstream `JWT_APP_SECRET`; the upstream default for `JICOFO_COMPONENT_SECRET` is forbidden. Private mTLS identities are separately provisioned and scoped. Examples contain only an explicitly rejected marker such as `__REQUIRED_UNUSABLE_NO_DEFAULT__`; empty, example, marker, reused, short, or known dev values fail acceptance. `JITSI_APP_SECRET` and `JITSI_EVENT_HMAC_SECRET` are each at least 32 bytes and are rejected when reused with Hub, LiveKit, XMPP, mTLS or each other.
- Do not edit Mail/Calendar, `frontend/src/modules/calls/create-call-dialog.tsx`, or `pro-qa-ai/**`.
- All tests/build/typecheck/lint are Docker-only through accepted exact services. No host tests or validators. No Jitsi image pull, Compose render, build, or container start before explicit infrastructure authorization.
- No invented container `user`, numeric UID/GID, `read_only`, `tmpfs`, writable path, or branding mount. Each must be backed by the selected-image Linux discovery receipt before it enters accepted YAML.
- Secrets/tokens must not appear in logs, health endpoints, browser configuration, public DTOs, snapshots, or test output.

## Frozen backend contracts

```go
const ProviderLiveKit = "livekit"

type JoinCredential struct {
    Provider   string
    RoomName   string
    ServiceURL string
    Token      string
}

type GuestJoinCredential struct {
    Identity string
    Join     JoinCredential
}

type ConferenceProvider interface {
    Provision(ctx context.Context, roomName string, maxParticipants uint32) error
    Delete(ctx context.Context, roomName string) error
    IssueJoinCredential(roomName, userID, userName string) (JoinCredential, error)
    IssueGuestCredential(roomName, guestName string) (GuestJoinCredential, error)
    UpdateParticipantState(ctx context.Context, roomName, identity, stateJSON string) error
}

type CallsConfig struct {
    ActiveProvider string
    Jitsi          JitsiConfig
}

type JitsiConfig struct {
    Enabled    bool
    Host       string
    PublicURL  string
    AppID      string
    AppSecret  string
    XMPPDomain string
}
```

When `JITSI_ENABLED=true`, dedicated lowercase `Host`, exact `PublicURL="https://"+Host`, nonempty `AppID`, non-placeholder `AppSecret` of at least 32 bytes, and nonempty `XMPPDomain` are mandatory. `Host` rejects `meet.jit.si` and all its subdomains, ports, paths, query, fragment, userinfo, uppercase, and trailing dots. `AppSecret` must differ from Hub `JWT_SECRET` and the LiveKit API secret. These checks are not weakened by development mode. Unknown provider values fail closed; explicit `jitsi` returns stable text containing `jitsi provider is not available before phase 152`.

The issuer method boundary is exact:

```go
Issue(call *ent.Call, authenticatedUserID uuid.UUID, serverKnownUserName string) (string, error)
```

`call` is a server-loaded Ent entity, not a client DTO. It must be nonnil, have a nonzero ID, and be active. `authenticatedUserID` must be nonzero; `serverKnownUserName` must be nonempty after trimming. The room claim is exactly `call.ID.String()` and is never accepted as an argument. Claims are HS256 only, `iss=JITSI_APP_ID`, `aud=["jitsi"]`, `sub=JITSI_XMPP_DOMAIN`, `iat`, `exp=iat+5m`, exact `room`, and `context.user.id/name`. No wildcard, moderator, features, guest policy, or arbitrary claims exist.

This issuer is unreachable in Phase 150: no HTTP route, WebSocket subject, Handler method, public DTO, frontend token request, or active-provider wiring. Its existence does not prove foreign-call membership. Caller-to-call membership, rate limiting, replay policy, and the public credential endpoint are Phase 152 gates.

The historical `call.ID.String()` room sentence above is superseded. The accepted and executable issuer contract requires a nonblank persisted `call.RoomName` and places exactly that value in the JWT `room`; `Call.ID` remains Hub business identity. All issuer RED assertions must use `call.RoomName`.

## Low-disk test cadence

The historical Tasks 0-6 batched cadence below is superseded as implementation acceptance authority. It is retained only as an old receipt description; every new atom follows the per-atom RED/GREEN rule in the 2026-08-30 amendment.

1. Write the complete backend RED contract packet once.
2. Run exactly one sequential `docker compose -f docker-compose.local.yml run --rm backend-unit` RED receipt.
3. Implement Tasks 1–6 sequentially; use static review between tasks, without repeated full-suite runs.
4. Run exactly one sequential `backend-unit` GREEN receipt after Task 6. If it fails, triage from that receipt before any rerun.
5. For the isolated frontend adapter, run one `frontend-unit` RED and one `frontend-unit` GREEN only.
6. Do not pull/build/render/start Jitsi while the infrastructure gate is closed.

## Task 0: Batched backend RED contract packet

**Exclusive files:**

- `backend/tests/calls/provider_interface_contract_test.go`
- `backend/tests/calls/livekit_provider_adapter_test.go`
- `backend/tests/calls/service_provider_injection_test.go`
- `backend/tests/platform/config/calls_provider_config_test.go`
- `backend/tests/calls/provider_factory_test.go`
- `backend/tests/calls/jitsi_credential_issuer_test.go`

Write all failing tests before production changes, then perform the single backend RED run. Coverage includes the exact typed provider methods; LiveKit behaviour parity including guest credential identity/join semantics; service injection for join-token delegation, create rollback, and participant-state delegation; provider defaults and fail-closed values; exact Host/PublicURL binding, deprecated-env rejection, and cross-secret rejection; factory rejection; and issuer output/binding. Rollback assertions freeze exact persistence arguments and order: provider `Provision` with the derived room/max occurs before repository create with the same room/channel/creator inputs; a repository failure triggers exactly one provider `Delete` for that room and no credential issue. `GuestJoin` service delegation is intentionally a static-review assertion because it requires a concrete repository and would otherwise force a broad production test hook.

Issuer product negatives include nil call, inactive call, zero call ID, zero authenticated user ID, blank server-known name, client-room impossibility, and absence of secret/token output. Valid different Call/user inputs are not rejected: tests prove the token binds exactly to the server values passed on that invocation. Preventing a caller from substituting another valid Call/user is the Phase 152 authenticated membership boundary, not issuer responsibility. Synthetic malformed/wrong-algorithm/wrong-secret/expired/non-five-minute tokens are strict inspection-oracle self-tests only; they validate the test parser and must not be described as incoming-JWT product verification. Product acceptance covers only tokens emitted by the issuer. Parsing tests never log credentials.

## Task 1: Provider interface contract

**Exclusive file:** `backend/internal/calls/provider.go`.

Create only provider-neutral types and the narrow `ConferenceProvider` interface. It contains no config reads, LiveKit construction, Jitsi construction, routing, or public DTO change. Review interface direction and typed credential semantics against `provider_interface_contract_test.go`. Do not run the full suite here.

## Task 2: LiveKit adapter extraction

**Exclusive file:** `backend/internal/calls/livekit_provider.go`.

Copy only the LiveKit SDK-facing behaviour currently owned by `CallRoomService`—room provision/delete, user/guest credential generation, service URL, and participant metadata—into `LiveKitConferenceProvider` without changing grants, expiry, errors, webhook/recording semantics, or compatibility output. Do not edit or delete `service_room.go` in Task 2: its provider-neutral orchestration helpers remain authoritative until Task 3 changes their argument types, so temporary SDK duplication is an intentional compile-safe transition. The adapter test owns executable guest semantics: generated guest identity plus nested typed `JoinCredential` provider/room/service URL/token. Do not edit service wiring. Review against the frozen adapter test; no full-suite run.

## Task 3: Call service provider injection

**Exclusive files:**

- `backend/internal/calls/service.go`
- `backend/internal/calls/service_room.go`
- `backend/internal/calls/export_hooks.go`

Make `CallService` depend on `ConferenceProvider`. Preserve `service_room.go` and change its provider-neutral orchestration helper arguments from `*CallRoomService` to `ConferenceProvider`; remove only the SDK method bodies duplicated into the adapter, and do not delete the file until those helpers are relocated in a later atomic task. Preserve persistence order and rollback, unwrap typed credentials only at the compatibility DTO boundary, and preserve chat/events/recording behaviour. Add exact hubtest constructor `ExportNewCallServiceWithProviderForTest`. During the compile transition update `ExportFinalizeCreate` to the provider argument and keep `ExportNewCallServiceForTest`, `ExportCallServiceRoom`, `ExportCallServiceSetRoom`, `ExportCallRoomServiceCfg`, and `ExportGenerateGuestToken` as explicitly temporary compatibility helpers or migrate their callers in the same Task 3 diff; no dangling signature is allowed. The RED service test intentionally executes join-token delegation, create rollback with exact repository arguments/order, and participant-state delegation. `GuestJoin` delegation is reviewed statically against the adapter contract; do not add a broad repository bypass/test hook merely to execute it. No config selection or server wiring is owned here. Review against injection tests; no full-suite run.

## Task 4: Calls and Jitsi configuration validation

**Exclusive files:**

- `backend/internal/platform/config/config.go`
- `deployments/.env.example`
- `docker-compose.local.yml`

Add exact `CallsConfig`/`JitsiConfig` parsing and validation. Default active provider to LiveKit. Enabled Jitsi requires `JITSI_HOST`, exact `JITSI_PUBLIC_URL="https://"+JITSI_HOST`, AppID, XMPP domain, and separate non-placeholder AppSecret of at least 32 bytes in every mode. Host must satisfy the exact global ASCII DNS regex/length/label grammar and rejects `meet.jit.si`, every `.meet.jit.si` suffix, IP, underscore, wildcard, port, path, query, fragment, userinfo, uppercase, and trailing dot. Reject AppSecret equality with Hub `JWT_SECRET` or the LiveKit API secret. Any nonempty deprecated `JITSI_ISSUER` or `JITSI_AUDIENCE` input fails closed; it is never silently ignored. Document only required backend variables; do not add Jitsi services. `JITSI_XMPP_DOMAIN` is the sole source later mapped to deployment `XMPP_DOMAIN` and JWT `sub`. No suite run.

## Task 5: Provider factory and server wiring

**Exclusive files:**

- `backend/internal/calls/provider_factory.go`
- `backend/cmd/server/wiring_integrations.go`

Create a factory which constructs LiveKit only. `jitsi` returns the frozen Phase-152 error; all unknown values fail closed. Wire the factory into the server while leaving LiveKit recording, webhook, `/lk/*`, and rollback assets intact. No issuer construction or transport is allowed. Review wiring statically; no suite run.

## Task 6: Server-only Jitsi credential issuer

**Exclusive files:**

- `backend/internal/calls/jitsi_credential_issuer.go`
- `backend/internal/calls/export_hooks.go` sequentially after Task 3, only if the external `hubtest` package requires the issuer hook

Implement the exact `Issue(*ent.Call, uuid.UUID, string)` boundary and claims above using `github.com/golang-jwt/jwt/v5`. `export_hooks.go` has sequential, nonconcurrent ownership: Task 3 first lands the service-provider exports, then Task 6 may reopen the same file only for the issuer hook. If the external `hubtest` package needs a hook, its exact name is `ExportNewJitsiCredentialIssuerForTest`. It exists only in the build-tagged test export surface, returns the concrete unexported issuer type, and that type exposes only the exported `Issue` method required by the test. The hook is not a transport or public product API and must not wire the issuer to server startup, handlers, routes, WebSocket, DTOs, or frontend.

After independent review, run the atom's exact backend GREEN gate. Record PASS/FAIL from the Docker receipt only. Membership authorization and Jitsi activation remain unimplemented by design.

## Task 7: Deployment preflight and static env contract

**Exclusive files:**

- `deployments/jitsi/.env.example`
- `deployments/jitsi/DEPLOYMENT-PREFLIGHT.md`

Create only the preflight/static contract. It lists required dedicated lowercase `JITSI_HOST`, exact `JITSI_PUBLIC_URL=https://${JITSI_HOST}`, `JITSI_XMPP_DOMAIN`, `JITSI_APP_ID`, mandatory Hub-owned `JVB_ADVERTISE_IPS`, `JVB_DISABLE_STUN=true`, XMPP component domains, and five separate required secrets: `JITSI_APP_SECRET` (mapped only to upstream `JWT_APP_SECRET`), `JICOFO_COMPONENT_SECRET`, `JICOFO_AUTH_PASSWORD`, `JVB_AUTH_PASSWORD`, and `JITSI_EVENT_HMAC_SECRET`. Private mTLS identities are separately provisioned and scoped. Secret examples use only the rejected `__REQUIRED_UNUSABLE_NO_DEFAULT__` marker and state `required/no default`.

Event signing, JWT, XMPP, mTLS, Hub and LiveKit credentials are never reused.

Record exact equality: overlay `XMPP_DOMAIN` is populated from `JITSI_XMPP_DOMAIN`, and issuer JWT `sub` uses the same backend value. Freeze the four digest receipts by reference to immutable `150-RESEARCH.md` without editing it.

The preflight matrix must leave these rows `UNBOUND / NOT_RUN`: effective image user, UID/GID, config volume ownership, writable directories, rootless compatibility, branding mount path, final effective `external_api.js`, image SBOM, Caddy/TLS, JVB advertise reachability, UDP media, restart, and rollback. At the end of Task 7, `deployments/docker-compose.jitsi.yml` remains `NOT_CREATED`; no overlay is claimed accepted.

## Task 8: GATED post-discovery overlay — not executable now

**Gate status:** `BLOCKED_BY_AUTHORIZED_LINUX_DISCOVERY`; do not dispatch during the current low-disk/static phase.

**Future exclusive files after gate opens:**

- `deployments/docker-compose.jitsi.yml`
- `deployments/Caddyfile.prod`
- `deployments/docker-compose.prod.yml`

Required gate input is an authorized task-owned Linux receipt for the exact four digest-pinned images proving effective users, required UID/GID handling, writable config directories, ownership transition, branding/static-asset path, healthcheck commands, and effective self-hosted `external_api.js`. Discovery must obey an explicit low-disk budget. Until that receipt exists, overlay YAML is `NOT_CREATED / UNBOUND`; a speculative file may only be labelled `DRAFT_NON_ACCEPTED` and cannot satisfy Phase 150.

After the gate opens, map `XMPP_DOMAIN=${JITSI_XMPP_DOMAIN:?required}` and all five secrets with `${VAR:?required}` and no fallback, including Hub `JITSI_APP_SECRET` only as upstream `JWT_APP_SECRET`. The default public boundary is Caddy `443/tcp` plus JVB `10000:10000/udp`; port 80 and any ACME flow remain absent until a separate certificate-provisioning contract is accepted. Require `JVB_ADVERTISE_IPS=${JVB_ADVERTISE_IPS:?required}` and selected-image proof that public STUN is disabled. Use explicit control/edge/media networks and the four frozen digests. Do not add `user`, `read_only`, `tmpfs`, config/branding mounts, or healthchecks unless each exact setting is proven by the discovery receipt.

Static YAML review after discovery still does not establish runtime readiness. Compose render, start, TLS, local JWT accept/reject, browser media, UDP, restart, rollback, effective asset, and SBOM remain `NOT_RUN` until separately authorized runtime execution.

## Task 9: Local self-hosted iframe adapter

**Exclusive files:**

- `frontend/src/modules/calls/jitsi-external-api.ts`
- `frontend/tests/calls/jitsi-external-api.test.ts`

First write a RED unit packet and run exactly one Docker `frontend-unit` RED. The pure adapter accepts only the validated canonical self-hosted HTTPS origin and returns `${origin}/external_api.js`. It rejects empty, HTTP, userinfo, query, fragment, non-root path, explicit port, and every `meet.jit.si` form. URL calculation performs no script injection or network call.

Implement the pure adapter with no fallback, provider activation, public DTO change, or import from protected `create-call-dialog.tsx`; then run exactly one Docker `frontend-unit` GREEN. The unit does not prove the asset exists in an image; effective-asset verification stays in the gated Linux/runtime matrix.

## Task 10: Documentation and evidence

**Exclusive files:**

- `docs/runbooks/jitsi-shadow-deployment.md`
- `deployments/jitsi/THIRD-PARTY-NOTICES.md`
- `.planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-VERIFICATION.md`
- `docs/wiki/DEPLOY.md`
- `docs/TECHNICAL.md`

Do not edit immutable `150-RESEARCH.md`. Document the active LiveKit rollback contract, provider seam, unreachable issuer, preflight-only deployment state, and gated overlay. Do not publish overlay bring-up instructions as executable/accepted until Task 8 is unblocked.

`150-VERIFICATION.md` separates registry metadata, backend/frontend Docker units, discovery, static overlay, and runtime evidence. It records overlay as `NOT_CREATED / UNBOUND` after Task 7, gated Task 8 as `BLOCKED_BY_AUTHORIZED_LINUX_DISCOVERY`, and every runtime row as `NOT_RUN`. No result is upgraded from a plan, source scan, or container-start assumption.

## Phase acceptance and reporting gate

- Backend provider/config/issuer work may be reported PASS only from the one final Docker `backend-unit` GREEN receipt; frontend adapter only from its Docker GREEN receipt.
- LiveKit remains the only active/rollback provider. No public Jitsi credential flow, membership authorization, provider cutover, UI conference activation, callback, Jibri recording, or LiveKit removal is delivered.
- After Task 7, the honest verdict is `PARTIAL — STATIC_FOUNDATION_ONLY`; overlay is `NOT_CREATED / UNBOUND`; runtime is `NOT_RUN`. Phase 150 cannot be called complete, production-ready, self-hosted Jitsi deployed, or media-ready.
- Even if Task 8 later creates statically reviewed YAML, production readiness does not advance until authorized Linux render/start/security/media/rollback/SBOM evidence passes. Planning/static artifacts alone must not increase the deployment percentage beyond the static-foundation baseline.
- Any floating image, external SaaS/API/STUN/TURN/analytics, divergent XMPP subject, usable secret default, guessed user/UID/GID/writable/read-only/tmpfs setting, exposed internal component, active `jitsi` provider, issuer transport exposure, secret leak, or false runtime PASS is a critical blocker.

## 2026-08-30 durable Calls and self-hosted Jitsi execution amendment

This amendment supersedes conflicting future-work prose but preserves every existing Phase-150 receipt and `NOT_RUN/UNBOUND` boundary. The tracked communications ADR amendment is canonical authority and this section is self-contained.

### Frozen identity, provider and transaction rules

- `Call.ID` is the Hub business identity. Persisted `Call.RoomName` is the opaque Hub-generated provider session and exact Jitsi JWT `room`; it is never client input or authorization.
- Every Call records immutable `provider=livekit|jitsi` and `presence_mode=legacy|durable`. Existing rows are `livekit/legacy`; the active-provider setting selects only new Calls. No active Call changes provider or writer.
- A Hub-issued connection is `(CallID, UserID, Epoch)`; a new credential receives a new nonzero epoch. Credential issuance creates unknown state but is not presence.
- Public `joined` is only aggregate `0 -> 1`, `left` only `1 -> 0`; ended Calls produce one ended event and no per-user left storm.
- Dedupe accepts only a verified provider-side `(provider, delivery_id)`; for Jitsi the provider-deployment source module creates and persists that ID before send. One PostgreSQL transaction owns `inbox -> connection/refcount/compatibility projection -> platform outbox`; publication is post-commit and at-least-once.
- LiveKit durable ingress remains disabled until its pinned version proves a signed nonempty delivery ID stable over retry. The current legacy LiveKit path stays active.
- Jitsi authoritative events may come only from a Hub-owned custom Prosody module installed through the selected release's supported custom-plugin/MUC-module path. Exact lifecycle hook names, ordering, retry semantics, and any equivalent stock durable egress are `UNVERIFIED_FOR_SELECTED_DIGEST` until JITSI-D1. No stock facility is accepted as durable authority without that proof. The custom module must persist the exact envelope plus a stable provider-deployment delivery ID in an accepted durable source spool before delivery and sign those stored bytes with a distinct HMAC-SHA256 credential. Browser iframe events, SaaS, JaaS and public provider APIs are not authority.

Private TLS in this plan means a separately verified transport adapter/sidecar that terminates mutually authenticated TLS for the provider-to-Hub route plus independent raw-body HMAC-SHA256 verification. Direct Prosody `net.http` client-certificate support is `UNKNOWN` and cannot satisfy this gate by assumption.

### Atomic TDD sequence

| Atom | Exclusive ownership and exact outcome | Acceptance focus |
| --- | --- | --- |
| CALL-DUR-01 | modify `backend/ent/schema/call.go`; add `backend/migrations/20260830_calls_durable_presence_expand.sql`; add focused `backend/tests/calls/` schema contract | additive immutable provider/presence declarations; historical `livekit/legacy`; no runtime/router change or inferred durable backfill |
| CALL-DUR-02 | add Ent schemas `CallParticipantConnection`, `CallParticipantPresence`, `CallProviderDelivery`, their generated output and focused tests | unique business keys, nonnegative counts, cascade/retention shape, no adapter |
| CALL-DUR-03 | create `backend/internal/calls/provider_participant_identity.go` and its unit test | canonical `hub.call-participant.v1/<call>/<user>/<epoch>`; every zero/malformed component rejects; no `Stringer` or logging |
| CALL-DUR-04 | add pure durable delivery/result and v1 event payload builders | payload contains event ID/call/user/presence sequence but never epoch/provider/delivery ID/token/raw body |
| CALL-DUR-05 | add `durable_presence_repository.go` behind the accepted CAS port | inbox duplicate, fixed lock order, exact expected-state/count CAS, 0->1/1->0 compatibility/count/outbox effects, rollback/no partial commit |
| CALL-DUR-06 | evolve `backend/ent/schema/outboxevent.go`, publisher and worker for nullable ordering key/sequence | same-key earliest-row claim, failed/leased earlier row blocks later, unordered non-regression, stable event ID, no exactly-once claim |
| CALL-DUR-07 | add LiveKit verifier/parser only after delivery-ID proof | signature before parse/apply; stable signed provider ID only; malformed/replay/duplicate fail closed; no legacy mutation |
| CALL-DUR-08 | add provider registry, immutable mode router and durable credential allocation | selector affects new Calls only; legacy stays LiveKit; unknown row exists before token; no dual writer/public DTO regression |
| CALL-DUR-09 | durable room-finish and unknown-left reconciliation request | one ended event, no left fanout, unknown-left no public edge |
| JITSI-01 | correct all issuer/plan contracts to persisted `Call.RoomName`; split credential issuance from optional provisioning capability | Jitsi does not fake LiveKit provision/delete/metadata and remains fail-closed |
| JITSI-02 | roadmap group only: freeze persistent source storage and selected hook semantics, then implement a custom Prosody event module, dispatcher, private transport and Hub ingress | provider-side durable delivery ID before send, identical retry bytes, system/guest/breakout filtering, HMAC verification, verified mTLS sidecar and crash/replay negatives |
| JITSI-03 | add provider-aware typed join DTO and one frontend Jitsi surface | exact validated self-hosted origin, one provider rendered, dispose/unmount cleanup, no token logging, no anonymous Jitsi guest |
| JITSI-04 | create discovery-backed Compose/Caddy/monitoring overlay | exact digests/users/mounts/health/assets/JWT/event hook only; private control plane; public Caddy HTTPS plus JVB UDP 10000 |
| JITSI-05 | operator shadow, two-user media/reconnect, rollback-new-call, coordinated load | immutable per-call provider, LiveKit regression-free rollback, capacity evidence before default selection |
| JITSI-06 | Phase-154 Jibri/importer and eventual LiveKit retirement | separate disk/capacity gate, least-privilege SeaweedFS import, historic recording compatibility, exact deletion manifest after drain |

The table above is a dependency roadmap, not authority to dispatch a broad row. Only `CALL-DUR-01` and `CALL-DUR-03` have frozen implementation inputs below; documentation-only `JITSI-01` is fulfilled by this reviewed plan correction. Every other row stays non-dispatchable until root creates a separate exact TaskContract; the decompositions below are current frozen examples, not an exhaustive claim for every future row.

Exact `CALL-DUR-01` ownership is `backend/ent/schema/call.go`, `backend/migrations/20260830_calls_durable_presence_expand.sql`, and `backend/tests/calls/phase150_call_provider_presence_schema_test.go`. RED and GREEN both run `docker compose -f docker-compose.local.yml run --rm backend-unit`; the approved entrypoint accepts no selector. Expected RED is only missing `provider`/`presence_mode` schema descriptors; expected GREEN is target/build/vet/unit/race PASS. The unit proves descriptors/defaults/immutability intent only; SQL application/backfill remains `NOT_RUN` until integration.

Exact `CALL-DUR-03` ownership is `backend/internal/calls/provider_participant_identity.go` and `backend/tests/calls/phase150_provider_participant_identity_test.go`. It remains blocked until the accepted MPR-01 RED/GREEN receipt is followed by the accepted JITSI-02A RED/GREEN receipt. Only then may its test become the sole test-discovered missing-symbol oracle. RED and GREEN both run exactly `docker compose -f docker-compose.local.yml run --rm backend-unit`, with no selector, pull or build. Expected RED is only the missing canonical identity type/parser symbols; expected GREEN is target/build/vet/unit/race PASS with valid round-trip, zero/malformed rejection and no `Stringer`.

The frozen public contract is one sentinel `ErrProviderParticipantIdentityInvalid`; an opaque `ProviderParticipantIdentity` whose only field is a private `ConfirmedParticipantConnectionKey`; `NewProviderParticipantIdentity(key ConfirmedParticipantConnectionKey) (ProviderParticipantIdentity, error)`; `ParseProviderParticipantIdentity(raw string) (ProviderParticipantIdentity, error)`; and the value methods `IsValid() bool`, `Canonical() string`, and `ConnectionKey() ConfirmedParticipantConnectionKey`. Construction accepts only an already valid existing connection key and does not create a second call/user/epoch validity authority. Every construction or parse failure returns the exact sentinel and zero identity. Invalid/zero values report false, an empty canonical string and a zero connection key.

The parser accepts only the exact 134-byte ASCII language `hub.call-participant.v1/<call>/<user>/<epoch>`: the exact namespace and slash positions followed by three lowercase canonical hyphenated UUID strings whose parsed value is non-zero and whose `uuid.UUID.String()` equals the original segment. It rejects case variants, braces, URNs, compact UUIDs, whitespace, NUL/non-ASCII bytes, URL escaping or decoding, query/fragment suffixes, empty/missing/extra segments and every wrong namespace/version. It performs no trim, case fold, Unicode normalization or URL decode and adds no UUID-version/variant policy or distinctness requirement. `Canonical()` is derived from the wrapped key and returns the same exact language. The identity is correlation data only, never authentication or authorization evidence; raw identity values must not enter errors or logs. No `String`, `Stringer`, marshal, logging, Jitsi transport, duration, persistence, token or authz surface belongs to this atom.

Every CALL-DUR implementation dispatch must state `model: gpt-5.6-terra`, `reasoning_effort: xhigh`; its independent reviewer must state `model: gpt-5.6-terra`, `reasoning_effort: high`. Each receipt records `actual_model: UNKNOWN` and `actual_reasoning_effort: UNKNOWN` unless trustworthy runtime telemetry is available. Omission, inheritance, Sol downstream use or silent fallback blocks dispatch.

#### Exact provider-routing decomposition

- `CALL-DUR-08A` owns only `backend/internal/calls/models.go`, `repository.go`, and `backend/tests/calls/call_provider_repository_test.go`: provider enum and `(provider, room_name)` lookup. It proves a persisted LiveKit Call remains LiveKit after the default changes.
- `CALL-DUR-08B` owns only `provider.go`, `provider_factory.go`, `livekit_provider.go`, new `jitsi_provider.go`, and `backend/tests/calls/provider_registry_capability_test.go`: registry plus credential-only base capability. Jitsi cannot satisfy provision/delete/participant-metadata capabilities.
- `CALL-DUR-08C` owns only `call_join_credential_issuer.go`, `call_join_credential_projection.go`, `call_join_credential_projection_ent.go`, `jitsi_credential_issuer.go`, and its focused credential-allocation test: new Hub epoch, unknown connection before token, persisted provider/room, server-loaded identity, no public event.
- `CALL-DUR-08D` owns only `service.go`, `service_room.go`, `repository.go`, `handler.go`, and `backend/tests/calls/call_presence_mode_router_test.go`: create/join/end/callback routing by persisted mode, exactly one writer, legacy LiveKit parity and no provider flip for active Calls.

#### Exact ordered-outbox decomposition

`CALL-DUR-06A` owns `backend/ent/schema/outboxevent.go`, the Calls expand migration clauses, generated Ent output and `backend/tests/platform/outbox_ordering_schema_test.go`. It requires all-or-nothing `ordering_key/ordering_seq`, a partial unique keyed sequence and a status/key/sequence claim index. `CALL-DUR-06B` then owns only `backend/internal/platform/outbox/publisher.go`, `worker.go`, and `backend/tests/platform/outbox_ordered_delivery_test.go`: null-pair rejection, duplicate keyed sequence, concurrent claim, leased/failed-earlier blocking and unordered-event non-regression.

#### Exact schema contract

`CALL-DUR-02` must be decomposed into one schema file plus one focused test per dispatch. Connection rows carry call/user/epoch/state/timestamps/version and unique `(call_id,user_id,epoch)`; presence rows carry call/user/nonnegative count/nonnegative sequence and unique `(call_id,user_id)`; delivery rows carry provider-side ID/kind/resolved identities/timestamps/result and unique `(provider,delivery_id)`. Retention is 30 days after terminal state. The expand migration adds nonnegative checks and cascade policy. No connection, presence, epoch, delivery or refcount is inferred from historic `CallParticipant`.

Generated Ent output is a mechanical exception only. After the approved Docker generator runs, root freezes the observed `git diff --name-only -- backend/ent` list into that sub-atom TaskContract before review/staging. No generated path outside the accepted manifest may be staged; absent an approved generator receipt, schema/codegen atoms remain blocked.

The normal `backend-unit` run overlays only `backend/ent/schema`; it neither regenerates nor proves checked-in host `backend/ent/**`. Therefore `CALL-DUR-02` and `CALL-DUR-06A` are `BLOCKED_GENERATOR`. Their gate requires a separate founder-authorized Docker generator handoff/launcher that updates host generated files, freezes the exact generated manifest, records before/after disk and rebuilt image identity, and then permits the normal approved `run --rm backend-unit` verification against the refreshed image. No generator/rebuild command is authorized or invented by this packet; no generated atom may use an old-image unit GREEN.

#### Jitsi discovery and surface ordering

- `JITSI-D1` is a hard prerequisite for any provider-hook attachment or executable overlay. It inspects one exact digest at a time and freezes the supported hook, image user, mounts, writable paths, health, JWT component, effective assets, SBOM and license data.
- `JITSI-02A` may define only the provider-neutral envelope/HMAC verifier and negative `backend-unit` tests before D1; it attaches to no Jitsi image or route. Exact ownership is `backend/internal/calls/provider_event_hmac.go` plus `backend/tests/calls/phase150_provider_event_hmac_contract_test.go`. The only production API is `ExportPhase150VerifyProviderEventEnvelopeHMACSHA256(secret, rawEnvelope, receivedMAC []byte) (Phase150VerifiedProviderEventEnvelope, error)` plus `Phase150VerifiedProviderEventEnvelope.Bytes() []byte`; the verified type keeps its byte slice private and both the successful result and `Bytes()` are defensive copies. One sentinel `ErrPhase150ProviderEventEnvelopeInvalid` covers every rejected input or authentication failure. Validation order is secret length at least 32 bytes, raw envelope length `1..65536` bytes, already-decoded binary MAC length exactly 32 bytes, then constant-time HMAC-SHA256 comparison over the exact unmodified raw envelope bytes. Every failure returns the zero verified value and zero-value `Bytes()` returns nil without panic. No trim, hex/base64/header decode, normalize, decompress, JSON parse, delivery-ID extraction, clock/skew decision, replay state, log or `Stringer` is allowed.
- `JITSI-02B0` is a design/evidence gate, not code. Its target is a dedicated self-hosted provider-side PostgreSQL database reached through a Prosody named SQL store, with separate database/schema/role/credentials/lifecycle from Hub application PostgreSQL. It freezes retention, capacity, ownership, backup and crash/restart behavior. D1 proves only that the pinned image contains the required SQL/archive driver and hook/config surface. A separate founder-authorized JITSI-02D fault-injection gate must prove that successful append acknowledgement follows a durable database commit. Failure of either gate leaves B0 blocked. `memory|none|null`, Hub application databases, MUC room state, Prosody internal files, SQLite, and any unverified container filesystem are rejected as the source of truth.
- `JITSI-02B1`, after D1/B0, owns only the selected MUC hook filter/map and durable append. It rejects focus/system/recorder/JVB identities, anonymous/guest principals, and rooms not exactly mapped to persisted `Call.RoomName`. The exact selected-image mechanism must disable breakouts for the baseline 1:1 contour or a separate main/breakout mapping is required. Room destruction never implies Hub Call end; CALL-DUR-09 remains authority.
- `JITSI-02B2` owns only spool dispatch/retry/restart and must replay the same delivery ID and exact stored bytes without JSON re-encoding.
- `JITSI-02C1` owns only the separately verified mTLS sidecar byte-forwarding and client-certificate policy; `JITSI-02C2` owns only Hub ingress mTLS/HMAC/replay admission.
- `JITSI-02D` owns the isolated durable-write-before-send, restart/replay, byte-identity, wrong-cert, skew/size/content-type/tamper and crash matrix. No sub-atom may claim stock Jitsi webhook durability.

`JITSI-02B0/B1/B2/C1/C2/D` are `BLOCKED_NONDISPATCHABLE`: the approved Hub QA services do not execute the custom Lua module, selected spool, or sidecar. Root may dispatch them only after JITSI-D1 and a separate founder-authorized task-owned Docker launcher/service with exact RED/GREEN oracles. No command is invented or authorized here.

`JITSI-02A` uses the exact approved launcher `docker compose -f docker-compose.local.yml run --rm backend-unit` with no selector, pull or build. Its RED is valid only when build/vet succeed and unit/race fail solely on the missing frozen symbols above. Its test proves valid exact-byte authentication; body boundaries `0/1/65536/65537`; secret boundaries `31/32`; MAC boundaries `31/32/33`; one-byte/whitespace/NUL mutation rejection; hex/base64 text rejection where a decoded 32-byte MAC is required; non-JSON byte acceptance when correctly signed; zero-value failure results; zero-value `Bytes()` safety; and both input-to-result and per-`Bytes()` alias isolation. HTTP header decoding, streaming size enforcement before allocation, content type, mTLS peer identity, envelope parsing, durable replay/dedupe and business mutation remain `JITSI-02C1/C2/D`; this atom cannot establish provider compatibility, delivery durability or runtime readiness.

Target provider-source production semantics are `at-least-once after successful durable enqueue`, not zero-loss observation, and may be claimed only after B1/B2/D evidence. A crash before hook append or database unavailability can lose a post-event observation because the MUC state change is already committed. Production acceptance therefore also requires a bounded reconciliation design or separately proven pre-event fail-closed gate; metrics alone cannot relabel this gap as durable delivery.
- `JITSI-03A` owns the backend provider-aware typed join DTO and authenticated handler/service tests while leaving LiveKit compatibility fields intact and Jitsi guests denied. `JITSI-03B` later owns `frontend/src/api/calls.ts`, `frontend/src/stores/call-store.ts`, one new `JitsiCallSurface.tsx`, `CallRoom.tsx`, and focused frontend tests for exact self-hosted origin, one-provider rendering, dispose/unmount and token non-disclosure.

Each production atom starts with one focused Docker RED for the expected missing behavior, then the minimum implementation, one focused Docker GREEN, independent review, docs sync and exact-path commit. The commands below are runners, not an accumulated substitute for per-atom RED/GREEN; no additional broad rerun occurs until a coherent milestone packet. No atom may silently pull a later schema, adapter, server wiring, provider activation or deployment change forward.

Planned accumulated unit gates are:

```powershell
docker compose -f docker-compose.local.yml run --rm backend-unit
docker compose -f docker-compose.local.yml run --rm frontend-unit
```

They are `NOT_RUN` here. For each atom the applicable runner executes sequentially, without `--build`, only when no Mail/Calendar build/container is active. API/integration/e2e/load use only their exact named Docker services after the relevant self-hosted dependencies exist.

### Low-disk discovery and runtime gate

No Jitsi pull/build/render/start occurs while another buildx builder or adjacent branch service is active. Before each exact digest, record free disk and active containers; inspect/pull one image at a time; stop if task-owned Docker growth would exceed 8 GiB or free space would fall below 250 GiB. Sample `Win32_PerfFormattedData_PerfDisk_PhysicalDisk` three times, two seconds apart; defer when two readings show `PercentDiskTime >= 70` or `CurrentDiskQueueLength >= 2`. Initial task limits are at most 4 CPUs and 8 GiB RAM for the Jitsi contour, one build/pull at a time, and no floating tag or `--build` during discovery. These are safety ceilings, not capacity acceptance.

The selected `stable-11146-2` digests remain registry metadata until discovery proves Linux users, writable/config/branding paths, effective `external_api.js`, health commands, supported JWT/event hook, SBOM and license inputs. YAML cannot guess these values. Runtime acceptance separately proves TLS, JWT positive/negative, UDP media, restart, observability, secret rotation, rollback and no vendor host.

### 2026-08-30 official upstream discovery receipt

- Official release `stable-11146-2` is tied to docker-jitsi-meet commit `738058b`; the tagged Compose contains only the four core images in the baseline. Existing OCI index pins in `150-RESEARCH.md` remain the repository authority and are not re-resolved in this packet.
- From `stable-11146`, official containers are rootless/read-only with UID/GID 1000 and separate immutable config, durable storage, and regenerable tmp paths. Exact host ownership/mount discovery remains required before YAML.
- Official JWT configuration consumes `JWT_APP_SECRET`; Hub exposes `JITSI_APP_SECRET` only as its secret-store name and must project it explicitly. `JICOFO_COMPONENT_SECRET` is a fifth independent required value.
- The official JVB default can use public STUN. A Hub overlay must set explicit `JVB_ADVERTISE_IPS` and disable STUN; self-hosted TURN is a separate future contract, not an implicit external fallback.
- The tagged template proves the custom plugin path and MUC-module insertion path. Exact lifecycle hook names/order, retry behavior, selected-digest health, and absence or presence of an equivalent stock durable egress remain `UNVERIFIED_FOR_SELECTED_DIGEST` until JITSI-D1. No stock facility is accepted as authority without proof. The accepted custom design therefore remains blocked behind `JITSI-02B0/B1/B2/C1/C2/D`. This receipt is static discovery, not implementation or runtime evidence.

### Rollback and exit

On Jitsi degradation, selection returns to LiveKit for future Calls only; existing Jitsi Calls drain on Jitsi. Expand-only schema stays, ingress/outbox continue for already durable Calls, and neither provider's active Calls or recordings are rewritten. LiveKit removal requires conference and recording parity, zero active/pending LiveKit work, retained historical reads, completed rollback window, operational sign-off and a separately reviewed exact deletion manifest.

Until all runtime gates pass the honest status remains `DESIGN_FROZEN / STATIC_FOUNDATION_ONLY / RUNTIME_UNBOUND`; a plan, source diff, image pull, container start or iframe render cannot increase deployment readiness.
