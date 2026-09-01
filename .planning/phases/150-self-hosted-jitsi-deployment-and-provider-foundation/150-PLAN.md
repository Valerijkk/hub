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
| CALL-DUR-03 | create neutral `call_participant_connection_key.go`, `provider_participant_identity.go` and one unit test | allocation-safe `hub.call-connection.v1/...` plus provider `hub.call-participant.v1/...`; every zero/malformed component rejects; no `Stringer` or logging |
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

The table above is a dependency roadmap, not authority to dispatch a broad row. Only `JITSI-02A`, `CALL-DUR-01` and `CALL-DUR-03` have frozen implementation inputs below; documentation-only `JITSI-01` is fulfilled by this reviewed plan correction. Every other row stays non-dispatchable until root creates a separate exact TaskContract; the decompositions below are current frozen examples, not an exhaustive claim for every future row.

Exact `CALL-DUR-01` ownership is `backend/ent/schema/call.go`, `backend/migrations/20260830_calls_durable_presence_expand.sql`, and `backend/tests/calls/phase150_call_provider_presence_schema_test.go`. RED and GREEN both run exactly `docker compose -f docker-compose.local.yml run --rm backend-unit`, with no selector, pull or build. Before each run, repeat the full resource preflight: no active buildx/buildkit work, Mail/Calendar build container or other `backend-unit`; C: at least `250 GiB` free; at least `2 GiB` free physical RAM for the current serialized backend-unit series per the founder directive of 2026-09-01; and three disk readings two seconds apart, deferring when two readings have `PercentDiskTime >= 70` or `CurrentDiskQueueLength >= 2`. An idle BuildKit daemon may remain untouched. Any failed preflight defers the run. Expected RED is only missing `provider`/`presence_mode` schema descriptors; expected GREEN is target/build/vet/unit/race PASS. The unit proves descriptors/defaults/immutability intent only; SQL application/backfill remains `NOT_RUN` until integration.

The frozen descriptor contract adds only `field.Enum("provider").Values("livekit", "jitsi").Default("livekit").Immutable()` and `field.Enum("presence_mode").Values("legacy", "durable").Default("legacy").Immutable()` to `Call.Fields()`. Both are required/non-nillable, have no update default, and add no index, edge, generated-client edit, router, provider switch or inferred durable backfill. The external `calls_test` oracle is a genuine deterministic in-process unit: it imports `backend/ent/schema`, calls `(schema.Call{}).Fields()`, requires exactly one descriptor for each target name, and inspects each for non-nil `Info` with `Info.Type == field.TypeEnum`, exact enum values/order, default, immutability, non-optional/non-nillable shape, nil update default and nil descriptor error. It must not read source or migration files, use `enttest`, import generated Call APIs, touch a filesystem/provider boundary or claim code generation.

The migration is a one-shot additive owner-applied transaction for an existing PostgreSQL instance: four explicit `ALTER TABLE calls` statements add fail-fast `provider VARCHAR NOT NULL DEFAULT 'livekit'`, `presence_mode VARCHAR NOT NULL DEFAULT 'legacy'`, `CONSTRAINT calls_provider_allowed CHECK (provider IN ('livekit', 'jitsi'))`, and `CONSTRAINT calls_presence_mode_allowed CHECK (presence_mode IN ('legacy', 'durable'))`, bracketed by `BEGIN`/`COMMIT`. It deliberately uses no `IF NOT EXISTS`, catalog predicate comparison or constraint drop: any pre-existing same-name column/constraint fails the transaction for owner triage rather than silently accepting an unknown definition, and PostgreSQL transactional DDL prevents a partial commit. It contains the repository `needs-owner-apply` marker, no `UPDATE`, trigger, index, provider activation or durable-state inference. Database defaults are only the unverified compatibility design for old inserts; compatibility requires separately authorized migration/integration evidence. Rollback design reverts code and leaves an accepted additive expansion in place. Ent/application immutability is not a database trigger and does not claim to prevent owner/direct-SQL updates. Direct script replay is intentionally fail-closed, not idempotent; migration-ledger exactly-once application, SQL parse/apply/failure atomicity and clean-database codegen parity remain `NOT_RUN` until their separately authorized gates.

CALL-DUR-01 may be prepared only outside test discovery now. Its test may become the sole discovered behavior oracle only after accepted MPR-01 RED/GREEN and accepted JITSI-02A RED/GREEN receipts; it must be accepted before CALL-DUR-03 is restored. The exact resulting queue is MPR-01 -> JITSI-02A -> CALL-DUR-01 -> CALL-DUR-03. No queued draft changes the current sole MPR-01 oracle.

Exact `CALL-DUR-03` ownership is new `backend/internal/calls/call_participant_connection_key.go`, new `backend/internal/calls/provider_participant_identity.go` and `backend/tests/calls/phase150_provider_participant_identity_test.go`. It remains blocked until the accepted MPR-01 RED/GREEN receipt is followed by accepted JITSI-02A and CALL-DUR-01 RED/GREEN receipts in that order. Only then may its test become the sole test-discovered missing-symbol oracle. RED and GREEN both run exactly `docker compose -f docker-compose.local.yml run --rm backend-unit`, with no selector, pull or build. Expected RED is only the missing neutral key and canonical identity symbols; expected GREEN is target/build/vet/unit/race PASS with valid round-trip, zero/malformed rejection and no `Stringer`.

The frozen public contract adds `ErrCallParticipantConnectionKeyInvalid`; an opaque neutral `CallParticipantConnectionKey` constructed only by `NewCallParticipantConnectionKey(callID,userID,epoch)` with `IsValid`, `Canonical`, and UUID accessors; and the separate sentinel `ErrProviderParticipantIdentityInvalid`. `ProviderParticipantIdentity` has only a private neutral key, accepts it through `NewProviderParticipantIdentity(key CallParticipantConnectionKey)`, parses only the provider grammar, and exposes `IsValid`, `Canonical`, and `ConnectionKey() CallParticipantConnectionKey`. The neutral key canonical is `hub.call-connection.v1/<call>/<user>/<epoch>` and means correlation/allocation only; it is not the existing `ConfirmedParticipantConnectionKey`, which remains unchanged and reserved for a connection fact already trusted upstream. Every construction or parse failure returns its exact sentinel and zero value. Invalid values expose no component and render empty. Neither type is authn/authz, confirmation evidence, `Stringer`, text/JSON marshaler or logger-safe raw data.

The parser accepts only the exact 134-byte ASCII language `hub.call-participant.v1/<call>/<user>/<epoch>`: the exact namespace and slash positions followed by three lowercase canonical hyphenated UUID strings whose parsed value is non-zero and whose `uuid.UUID.String()` equals the original segment. It rejects case variants, braces, URNs, compact UUIDs, whitespace, NUL/non-ASCII bytes, URL escaping or decoding, query/fragment suffixes, empty/missing/extra segments and every wrong namespace/version. It performs no trim, case fold, Unicode normalization or URL decode and adds no UUID-version/variant policy or distinctness requirement. `Canonical()` is derived from the wrapped key and returns the same exact language. The identity is correlation data only, never authentication or authorization evidence; raw identity values must not enter errors or logs. No `String`, `Stringer`, marshal, logging, Jitsi transport, duration, persistence, token or authz surface belongs to this atom.

Every CALL-DUR implementation dispatch must state `model: gpt-5.6-terra`, `reasoning_effort: xhigh`; its independent reviewer must state `model: gpt-5.6-terra`, `reasoning_effort: high`. Each receipt records `actual_model: UNKNOWN` and `actual_reasoning_effort: UNKNOWN` unless trustworthy runtime telemetry is available. Omission, inheritance, Sol downstream use or silent fallback blocks dispatch.

#### Exact provider-routing decomposition

- `CALL-DUR-08A` owns only `backend/internal/calls/models.go`, `repository.go`, and `backend/tests/calls/call_provider_repository_test.go`: closed persisted provider data and exact `(provider, room_name)` lookup. Its isolated unit proves generated query/mapping/error behavior only; historical default/backfill persistence remains integration `NOT_RUN`.
- `CALL-DUR-08B` owns only `provider.go`, `provider_factory.go`, `livekit_provider.go`, new `jitsi_provider.go`, and `backend/tests/calls/provider_registry_capability_test.go`: registry plus credential-only base capability. Jitsi cannot satisfy provision/delete/participant-metadata capabilities.
- `CALL-DUR-08C` is `FROZEN_NONDISPATCHABLE`, decomposed below in strict sole-oracle order 08C1 durable Jitsi issuer -> 08C2 pure allocation coordinator -> 08C3 Ent transaction adapter. It preserves the legacy issuer, signs only an epoch-bearing provider identity, holds zero credential until commit, and creates no participant/presence/refcount/outbox effect. 08C1 requires accepted CALL-DUR-03 and 08B; 08C2 requires accepted 08C1; 08C3 requires accepted 08C2, generated 02A APIs and CALL-DUR-02G `SOURCE_ACCEPTED`. Persisted provider/mode mapping, guest denial, service/router/wiring and legacy parity remain 08D with no fallback.
- `CALL-DUR-08D` owns only `service.go`, `service_room.go`, `repository.go`, `handler.go`, and `backend/tests/calls/call_presence_mode_router_test.go`: create/join/end/callback routing by persisted mode, exactly one writer, legacy LiveKit parity and no provider flip for active Calls.
- `CALL-DUR-08E` is the post-08D PostgreSQL credential/GDPR acceptance gate. It remains `BLOCKED_REDECOMPOSITION / BLOCKED_QA_CAPABILITY`: exact SUT bridge, `TASK/requirement/risk/TC` allocation, bounded PostgreSQL client/lock, runner DSN, source/image receipt and integration ownership require a separate TaskContract. It must exercise production 02G plus 08C/08D flows; a direct-SQL simulation cannot satisfy it.

`ENT-CALL-GEN-01` is the single required `BLOCKED_AUTHORITY_AND_RESOURCES` image-refresh gate after accepted `CALL-DUR-02M` and before `CALL-DUR-08A`. CALL-DUR-01 plus 02A/02B/02C remain `BLOCKED_GENERATOR`, and 08A cannot start, until this shared receipt. `backend-unit` generated Ent packages are image-contained: `testframework/docker/Dockerfile.qa` runs `go generate ./ent` during build, while the normal service bind-mounts only `backend/ent/schema`. Therefore the canonical run cannot refresh or prove generated provider or durable-entity APIs. No host `go generate`, raw-SQL workaround, ignored generated-file staging, `--build`, pull or image creation is authorized. After explicit founder authority and the normal resource preflight, one separately receipted backend-unit image build followed by the canonical run must cover every accepted 01/02 schema source. Before build, the receipt proves host `backend/ent` contains only tracked `generate.go` and `schema/**`, then binds the source commit and hashes of all schema, generator, module and QA-Dockerfile inputs to generation sentinels, local image ID and later backend-unit evidence. Generated output remains image-only and no ignored-output Git manifest is claimed.

The separate `CALL-DUR-08A` TaskContract is `SOURCE_ACCEPTED` in commit `eae39e7df9f6160b591e3dd36a798564a19bd7c6` after accepted `ENT-CALL-GEN-01` and `CALL-DUR-02G`. The sole oracle `backend/tests/calls/call_provider_repository_test.go` is frozen at SHA-256 `E4622E9BEE982FBB9024EAFD3095E829E273FC41D66CD2DD96229504A0C6AAD3` after fresh full-file compile/contract and security reviews at `C0/I0/M0`. Initial SHA `AB4D41642ED3B0445FDCF96FBCBB44C7E1833542CC271E5688DB6E53DDF7EE2B` and its RED are rejected because two parser error branches returned one value from a `([]string,error)` function. The arity-only correction produced `DB7DEE16D68E80EFAB07407FCD042BE3AD37CF8BC5624BFAE1579AE5D490208A`; two later fail-closed padded-provider cases produced the final frozen SHA. Ownership is exactly `backend/internal/calls/models.go`, `backend/internal/calls/repository.go`, and that one new oracle; no other production, test, generated, schema, migration, service, handler, provider, routing, deployment, frontend, Messenger or Mail/Calendar file belongs to 08A. The additive public API is exact:

```go
type PersistedCallProvider string

const (
    PersistedCallProviderLiveKit PersistedCallProvider = "livekit"
    PersistedCallProviderJitsi   PersistedCallProvider = "jitsi"
)

var ErrCallProviderLookupInvalid = errors.New("invalid call provider lookup")

func (*EntRepository) GetCallByProviderAndRoomName(
    context.Context,
    PersistedCallProvider,
    string,
) (*ent.Call, error)
```

`PersistedCallProvider` is record authority distinct from the legacy credential/provider adapter string in `provider.go`. The method returns exact `ErrCallProviderLookupInvalid` before query for nil/typed-nil context, nil receiver/client, a provider other than the two exact constants, blank room, or `strings.TrimSpace(roomName) != roomName`. A valid lookup converts the validated provider only at the generated Ent predicate boundary, uses exact provider AND room predicates with `.Only(ctx)`, and has no room-only fallback. It includes active and ended Calls, returns Ent not-found/not-singular and driver/dependency errors without wrapping, adds no tenant/channel/deleted/status predicate, and preserves the existing globally unique `room_name`. No validation helper or extra public method is added.

The 08A oracle is a genuine isolated `hubtest` unit using an in-test fake Ent Postgres dialect driver: no real/SQLite database, filesystem, network, `enttest`, or new dependency. It proves exact conjunctive provider/room SQL and args, generated row mapping, no-row `ent.IsNotFound`, multi-row `ent.IsNotSingular`, unwrapped driver-error identity, exact invalid-input sentinel with no query, and closed provider values. It does not prove CALL-DUR-01 PostgreSQL migration/default/backfill, DB-trigger immutability, planner/index behavior, callbacks, authentication, tenant isolation, provider selection, routing or runtime; those remain integration/08B-08D `NOT_RUN`. Existing room-only callback methods are unchanged by 08A. The later trusted ingress must authenticate source/mTLS/HMAC and delivery identity before this resolver; the tuple itself is never authn/authz.

#### CALL-DUR-08B immutable capability registry implementation plan

**Spec:** `150-SPEC.md` CALL150-13. **Goal:** add a dormant immutable registry that exposes only capabilities a provider truly implements, while current factory/config/server wiring keep Jitsi unavailable. **Exact files:** modify `backend/internal/calls/provider.go`, `provider_factory.go`, `livekit_provider.go`; create `jitsi_provider.go` and `backend/tests/calls/provider_registry_capability_test.go`. No other production, test, config, wiring, service, handler, Ent, deployment or frontend path is owned.

**Status (2026-09-01):** `SOURCE_ACCEPTED` in commit `b66dada8`. The frozen oracle SHA-256 is `E0913BECB2D46C00EC1E4EC467741CFAE8911869ECC6CFC7B9562525CC67D918`; canonical clean GREEN is `worktree-master-backend-unit-run-9af81ee5d32d`. This accepts only dormant registry source plus isolated unit/race behavior. Provider selection/routing, persistence mapping, guest denial/HTTP mapping, config activation, service/server wiring, integration, deployment, media and browser runtime remain `NOT_RUN/UNBOUND`.

The exact additive API is:

```go
const ProviderJitsi = `jitsi`

type ConferenceProviderID string
type ConferenceProviderCapability uint8
type ConferenceProviderCapabilities uint8

func (c ConferenceProviderCapabilities) Has(ConferenceProviderCapability) bool

const (
    CapabilityIssueCallJoinCredential ConferenceProviderCapability = 1 << iota
    CapabilityProvisionRoom
    CapabilityDeleteRoom
    CapabilityIssueGuestCredential
    CapabilityUpdateParticipantMetadata
)

type ConferenceRoomProvisioner interface {
    Provision(context.Context, string, uint32) error
}
type ConferenceRoomDeleter interface {
    Delete(context.Context, string) error
}
type ConferenceGuestCredentialIssuer interface {
    IssueGuestCredential(string, string) (GuestJoinCredential, error)
}
type ConferenceParticipantMetadataUpdater interface {
    UpdateParticipantState(context.Context, string, string, string) error
}

type RegisteredConferenceProvider interface {
    CallJoinCredentialIssuer
    ProviderID() ConferenceProviderID
    Capabilities() ConferenceProviderCapabilities
}

var (
    ErrConferenceProviderRegistryInvalid = errors.New(`invalid conference provider registry`)
    ErrConferenceProviderRegistryDuplicate = errors.New(`duplicate conference provider registration`)
    ErrConferenceProviderUnknown = errors.New(`unknown conference provider`)
    ErrConferenceProviderNotRegistered = errors.New(`conference provider is not registered`)
    ErrConferenceProviderCapabilityUnavailable = errors.New(`conference provider capability is unavailable`)
)

type ConferenceProviderRegistry struct {
    providers map[ConferenceProviderID]RegisteredConferenceProvider
}

func NewConferenceProviderRegistry(...RegisteredConferenceProvider) (*ConferenceProviderRegistry, error)
func (r *ConferenceProviderRegistry) Lookup(ConferenceProviderID) (RegisteredConferenceProvider, error)
func (r *ConferenceProviderRegistry) Require(ConferenceProviderID, ConferenceProviderCapability) (RegisteredConferenceProvider, error)
func NewJitsiConferenceProvider(config.JitsiConfig) (*JitsiConferenceProvider, error)
```

`ConferenceProvider` retains its exact legacy five-method reflection contract. `ConferenceProviderID` is adapter-registry identity, not the future 08A persisted-call type and not `JoinCredential.Provider string`; 08D owns their explicit fail-closed mapping. Registration is construction-time only: there is no mutating `Register`. The closed IDs are exact lowercase `livekit|jitsi`. Constructor input rejects empty registry, nil/typed-nil provider, zero/unknown/padded/case-drift ID, missing credential bit, unknown capability bits, duplicate ID, and capability/interface mismatch in either direction. It returns exact unwrapped sentinels `ErrConferenceProviderRegistryInvalid` or `ErrConferenceProviderRegistryDuplicate`. Before ID or capability validation, `Lookup` and `Require` return exact `ErrConferenceProviderRegistryInvalid` when `r == nil` or `r.providers == nil`; the zero-value registry is invalid. On a valid constructed registry, `Require` resolves and validates the provider ID before validating capability, so an invalid ID returns exact `ErrConferenceProviderUnknown` even when the capability is also invalid. A valid omitted ID returns `ErrConferenceProviderNotRegistered`, never LiveKit fallback. `Require` otherwise accepts exactly one known bit and returns `ErrConferenceProviderCapabilityUnavailable` when the registered provider lacks it; lookup/require perform no provider method call.

`LiveKitConferenceProvider` advertises all five bits and already implements all narrow interfaces plus `CallJoinCredentialIssuer`. `JitsiConferenceProvider` validates `JitsiConfig`, composes the existing `jitsiCredentialIssuer` with the server clock, delegates only `IssueCallJoinCredential`, advertises only `CapabilityIssueCallJoinCredential`, and deliberately has no legacy provision/delete/room-string join/guest/participant-state method. Constructor/registry errors and logs contain no secret, token, room, service URL, identity or raw config. They create no client, network call, goroutine, room, participant, event or runtime binding.

- [x] **Step 1 - restore the sole oracle.** After the accepted dependency queue and a fresh resource preflight, move the reviewed ignored draft to exactly `backend/tests/calls/provider_registry_capability_test.go`. Confirm no other missing-symbol oracle is discoverable.
- [x] **Step 2 - prove RED once.** Run only `docker compose -f docker-compose.local.yml run --rm backend-unit`. Valid RED has build PASS and vet/unit/race failure solely for the frozen missing 08B symbols; any unrelated or stale-generator failure requires triage and no production dispatch.
- [x] **Step 3 - add provider contracts.** In `provider.go`, add the exact IDs, narrow interfaces, capability bits, `Has`, registered interface and five exact non-diagnostic sentinels. Keep `ConferenceProvider` and `JoinCredential` unchanged.
- [x] **Step 4 - add honest adapters.** In `livekit_provider.go`, add only `ProviderID` and all-capability reporting. In new `jitsi_provider.go`, compose the existing issuer and implement only provider ID, credential capability and exact credential delegation.
- [x] **Step 5 - add immutable registry.** In `provider_factory.go`, add constructor, `Lookup`, `Require`, typed-nil detection and exact capability/interface congruence. Preserve the existing `NewConferenceProvider` branches and stable Jitsi-unavailable error byte-for-byte.
- [x] **Step 6 - prove GREEN once.** Run the same canonical backend-unit command. Require build/vet/unit/race PASS and no Mail/Calendar overlap; this proves only isolated compile/unit/race behavior.
- [x] **Step 7 - accept and commit.** Require a conclusion-blind `be-ver` spec/diff/oracle review and an independent `security-auditor` no-disclosure/no-activation/no-fallback review, both with `C0/I0/M0`; the implementer cannot self-set `verified`. Then sync these five canonical docs without changing readiness, exact-stage only the five 08B code/test files plus owned docs, and commit `feat(calls): add immutable provider capability registry`.

The sole oracle covers constructor validation, duplicate no-overwrite, order-independent deterministic lookup, known-omitted versus unknown identity, nil registry, exact sentinel identity, no fallback, no issuer invocation, capability/interface congruence, LiveKit full capability, Jitsi credential-only structure, and returned-error non-disclosure. It uses no database, filesystem, network, generated Call provider field or credential issuance. It does not prove config loading, activation, persisted routing, Ent migration, epoch allocation, guest HTTP denial, media, deployment or runtime.

#### CALL-DUR-02G transactional GDPR serialization implementation plan

**Status:** `SOURCE_ACCEPTED` in commit `64e68da443ad0e842aeec300b4c206e6eb89e102`. **Dependencies:** accepted 02A/02B/02C source packets, accepted 02M and the shared `ENT-CALL-GEN-01` receipt. 02G now permits 08C3 source work; it does not provide the later 08E runtime/concurrency evidence. Exact production ownership is only `backend/internal/audit/repository.go` and `backend/internal/audit/service.go`; focused unit ownership is new `backend/tests/audit/call_durable_gdpr_serialization_test.go`, with mechanical no-op test-double compatibility edits limited to `backend/tests/audit/unit_support_test.go`. No Python integration, Calls service, handler, allocator, ingress, schema, generated, migration, Mongo, Messenger, frontend, deployment or Mail/Calendar file belongs to 02G.

The existing `Repository` gains exactly these transaction-capable methods; no constructor or new service dependency is added:

```go
LockUserForGDPR(context.Context, uuid.UUID) (user.SystemRole, error)
EraseUserDurableCallData(context.Context, uuid.UUID, time.Time) error
```

The existing pre-transaction founder lookup remains a fail-fast compatibility guard, but it is never transaction authority. Inside `WithTx`, the first repository call and first SQL statement must be `LockUserForGDPR(ctx,userID)`: one exact User query projecting only `system_role` under PostgreSQL `FOR UPDATE`, returning the role from the locked row. The narrow projection is part of the 02G contract and avoids reading unrelated profile data while retaining the row lock. The service immediately repeats the founder denial from that locked role before any mutation. It then captures one nonzero UTC `terminalAt` and invokes the single repository-owned durable erase before assignment or membership cleanup. That method deletes all 02A connections by exact `user_id`, deletes all 02B presences by exact `user_id`, and tombstones 02C deliveries; keeping those statements behind one repository method prevents callers from constructing a partial cleanup sequence.

Delivery tombstoning is two generated Ent bulk updates in this order. The first matches the target user and `result IN (pending,reconcile)` and atomically sets `result=ignored`, `terminal_at=terminalAt`, and `user_id=NULL`. The second matches the still-identifying target rows with `result IN (applied,ignored)` and clears only `user_id`; it does not rewrite their established result or terminal time. Both preserve `provider`, `delivery_id`, `kind`, `call_id`, `epoch`, and `received_at`, so `(provider,delivery_id)` remains the durable replay barrier. No raw SQL or host generation workaround is allowed. After those steps, the existing assignment/content/membership/audit/HardDelete order continues unchanged, and Mongo erasure remains strictly after the Ent commit.

User-first locking is the cross-flow invariant: active delivery admission takes `FOR KEY SHARE`, 08C credential allocation takes `FOR SHARE`, and GDPR takes `FOR UPDATE`; each later touches Call/Channel/member state only after the User lock. If admission wins, GDPR waits and erases/tombstones its committed rows. If GDPR wins, admission/allocation waits, observes no active User after commit, creates no connection/presence/credential, and ingress may persist only the already-frozen terminal ignored/null tombstone. Lock, founder recheck, either durable cleanup, any existing cascade step, audit creation or HardDelete failure rolls back every Ent mutation. A commit error is returned and suppresses Mongo, but does not claim a synthetic rollback after PostgreSQL has begun commit; its database outcome remains unknown to the caller. No delivery fact or replay barrier is deleted.

- [x] **Step 1 - restore one unit oracle.** Only after the predecessor queue and generator receipt, restore `backend/tests/audit/call_durable_gdpr_serialization_test.go` as the sole missing-behavior oracle. Its service recording repository proves the first in-transaction operation is the lock, then one durable erase before the existing cascade and HardDelete, locked-role founder denial, one nonzero UTC terminal time, stop-on-error, rollback ownership and no post-commit eraser on failure. Its generated-query subtests prove the erase method's exact connection -> presence -> pending/reconcile terminalization -> terminal user-clear statements and immutable replay columns. It creates no credential/presence side effect.
- [x] **Step 2 - prove canonical RED once.** Run only `docker compose -f docker-compose.local.yml run --rm backend-unit`. Build must pass; because `go vet` compiles the focused external test package, vet/unit/race may fail solely on the frozen missing 02G methods/order. Stale generation or unrelated failure is triaged, not repaired under this atom.
- [x] **Step 3 - implement the two owned production files.** Add the two exact repository methods and place the locked-role recheck plus composite durable cleanup at the start of the existing transaction callback. Test-double edits provide only deterministic compatibility behavior; they do not broaden assertions or product scope.
- [x] **Step 4 - prove canonical unit GREEN.** Repeat the same no-selector backend-unit command and require build/vet/unit/race PASS with no Mail/Calendar overlap.
- [x] **Step 5 - source-accept and commit.** Require conclusion-blind audit/backend review plus independent security/concurrency review at `C0/I0/M0`, exact-stage only the owned implementation/unit/docs, and commit atomically as `SOURCE_ACCEPTED`. This permits 08C3 source work only; it is not PostgreSQL conflict, migration, flow or activation evidence.
- [x] **Step 6 - preserve the runtime block.** `CALL-DUR-08E`, after implemented 08D, owns genuine production PostgreSQL credential/GDPR acceptance. Current `python-integration` has image-baked tests/cases, no PostgreSQL client/DSN/SUT binding, and mandatory traceability metadata has no allocated task chain; therefore no old-image or direct-SQL run is authorized or accepted here.

#### CALL-DUR-08C durable Jitsi allocation implementation plan

**Status:** `FROZEN_NONDISPATCHABLE / BLOCKED_GDPR_SERIALIZATION`. No 08C subatom is an early-dispatch exception: the global sole-oracle queue through accepted 08B must complete first. Semantic code dependencies are accepted CALL-DUR-03 identity and 08B Jitsi wrapper for 08C1, accepted 08C1 for 08C2, and accepted 08C2 plus generated 02A APIs and CALL-DUR-02G `SOURCE_ACCEPTED` for 08C3. Strict execution is 08C1 -> 08C2 -> 08C3. Each subatom restores exactly its one test as the sole discovered missing-behavior oracle, completes the canonical no-selector backend-unit RED/GREEN plus reviews, then removes that oracle from the queue before the next enters discovery. No service, handler, factory, registry, wiring, frontend, deployment or callback file belongs to 08C.

`CALL-DUR-08C1` modifies only `backend/internal/calls/call_join_credential_issuer.go`, `jitsi_credential_issuer.go`, future `jitsi_provider.go`, and adds `backend/tests/calls/call_durable_join_credential_issuer_test.go`. It adds, without changing the legacy interface:

```go
type DurableCallJoinCredentialIssuer interface {
    IssueDurableCallJoinCredential(
        call *ent.Call,
        identity ProviderParticipantIdentity,
        serverKnownUserName string,
    ) (JoinCredential, error)
}
```

The Jitsi implementation requires an active canonical server-loaded Call, a valid identity whose neutral `ConnectionKey().CallID()` equals `call.ID`, and a canonical nonempty server-owned name. It signs `identity.Canonical()` into JWT `context.user.id`. That provider language is exactly `hub.call-participant.v1/<call>/<user>/<epoch>` and is deliberately not the neutral internal key language `hub.call-connection.v1/...`. No raw identity string is accepted. The future Jitsi wrapper's 08C1 delta is limited to durable-facet delegation and its compile assertion; registry, config, factory, base capability and activation behavior remain byte-compatible. 08D later requires the facet only for persisted Jitsi+durable routing. The 08C1 oracle verifies exact JWT identity, call/key mismatch rejection, zero credential/error redaction and unchanged legacy UUID issuance, with no DB/network/filesystem.

`CALL-DUR-08C2` adds only `backend/internal/calls/call_durable_credential_allocation.go` and `backend/tests/calls/call_durable_credential_allocation_test.go`. Its exact pure API is:

```go
type DurableCallJoinCredentialAllocationSource interface {
    BeginDurableCallJoinCredentialAllocation(context.Context) (DurableCallJoinCredentialAllocationTx, error)
}
type DurableCallJoinCredentialAllocationTx interface {
    LockActiveUserName(context.Context, uuid.UUID) (string, error)
    LockActiveCall(context.Context, uuid.UUID) (*ent.Call, error)
    AuthorizeLockedCall(context.Context, *ent.Call, uuid.UUID) error
    CreateUnknownParticipantConnection(context.Context, CallParticipantConnectionKey) error
    Commit() error
    Rollback() error
}
type DurableCallJoinCredentialAllocator struct { /* private dependencies */ }
func NewDurableCallJoinCredentialAllocator(DurableCallJoinCredentialAllocationSource, DurableCallJoinCredentialIssuer) (*DurableCallJoinCredentialAllocator, error)
func (*DurableCallJoinCredentialAllocator) IssueForAuthenticatedUser(context.Context, uuid.UUID, uuid.UUID) (JoinCredential, error)
```

The exact logical order is `begin -> lock_active_user -> lock_active_call -> authorize_locked_call -> create_unknown_connection -> issue_durable_credential -> context check -> commit`; the credential stays local until commit succeeds. The fake-port trace has the seven dependency operations in that order, while a cancellation case proves the direct context check between issuer and commit. The allocator uses `uuid.New()` once, constructs neutral `CallParticipantConnectionKey(call,user,epoch)`, then constructs `ProviderParticipantIdentity`; tests capture the values and require valid matching call/user, nonzero epoch, neutral allocation semantics, and the distinct provider canonical namespace without asserting a random UUID. It never constructs `ConfirmedParticipantConnectionKey`; only a later trusted provider observation may do so. Loaded Call ID/status/room/origin and locked trimmed name are defensively validated. Typed-nil dependencies/context and zero IDs fail before begin. An already-cancelled context fails before begin; cancellation before commit rolls back.

New fixed sentinels are `ErrCallJoinCredentialUnavailable` and `ErrCallJoinCredentialCommitAmbiguous`. Structural drift returns exact existing `ErrCallJoinCredentialInvalid`; authorization denial remains exact `ErrCallForbidden`. Every dependency/signing/pre-commit failure collapses to exact `ErrCallJoinCredentialUnavailable`; commit failure collapses to exact `ErrCallJoinCredentialCommitAmbiguous`. The raw cause and rollback error are never stored in or reachable from the returned error chain, and ordinary wrapping/joining is forbidden. Bounded stage-only metrics may identify the failed operation but contain no raw error text or credential fact. Begin failure performs no rollback. Every later pre-commit failure invokes rollback exactly once and returns zero credential. Commit failure returns zero ambiguous error, invokes no rollback/retry/reissue, and may leave one unknown row. The isolated oracle covers every stop point, order/arguments, exact sentinel equality, `!errors.Is(returned,cause)`, fixed strings, zero-on-error and commit-before-exposure. Its fake ports expose no participant, presence, refcount, publish or outbox operation.

The fixed error strings are exactly `call join credential unavailable` and `call join credential commit outcome ambiguous`. A successful Begin that returns a nil or typed-nil transaction is treated as unavailable before transaction ownership: no rollback is attempted. An authorization error satisfying `errors.Is(err, ErrCallForbidden)` is collapsed to the exact `ErrCallForbidden` sentinel; every other authorization cause is collapsed to unavailable. These freezes close the final 08C2 oracle ambiguities without broadening the API or adding logs/metrics.

`CALL-DUR-08C3` adds only `backend/internal/calls/call_durable_credential_allocation_ent.go` and `backend/tests/calls/call_durable_credential_allocation_ent_test.go`. The PostgreSQL lock order is exact and uses the same transaction/context: active User `FOR SHARE`; active Call `FOR SHARE`; `AuthorizeLockedCall` itself owns, for a channel Call, live Channel `FOR SHARE` then matching ChannelMember `FOR KEY SHARE`; for an instant Call it owns exact immutable creator equality. User/Call/Channel shared locks allow parallel credential issuance but conflict with deactivation, Call end/delete and Channel soft-delete; membership key-share conflicts with removal. `CALL-DUR-02G` must separately make User `FOR UPDATE` the first operation of the existing GDPR transaction before membership deletion and durable cleanup; only then does User shared-lock serialize with GDPR. Any future transaction touching User plus Channel/Member must also start with User. The allocator supplies its already-built fresh key and the adapter inserts exactly one 02A row with default `state=unknown`; the adapter never generates another epoch, calls legacy `AddParticipant`, or changes aggregate presence, participant count, public sequence or outbox.

The 08C3 isolated fake-dialect oracle may prove generated query/transaction order and exact predicates only. Real PostgreSQL lock conflict, rollback, ambiguous commit, migration/FK/index behavior and concurrent issuance/lifecycle/GDPR cases remain CALL-DUR-08E after implemented 08D; the current Python runner cannot prove them. A successfully issued five-minute Jitsi JWT is not revoked by a later end/removal in 08C; lifecycle/revocation remains 08D/future scope. No 08C atom activates Jitsi, routes a persisted provider/mode, accepts guests, exposes HTTP, or processes callbacks.

#### Exact ordered-outbox decomposition

`CALL-DUR-06A` owns `backend/ent/schema/outboxevent.go`, the Calls expand migration clauses, generated Ent output and `backend/tests/platform/outbox_ordering_schema_test.go`. It requires all-or-nothing `ordering_key/ordering_seq`, a partial unique keyed sequence and a status/key/sequence claim index. `CALL-DUR-06B` then owns only `backend/internal/platform/outbox/publisher.go`, `worker.go`, and `backend/tests/platform/outbox_ordered_delivery_test.go`: null-pair rejection, duplicate keyed sequence, concurrent claim, leased/failed-earlier blocking and unordered-event non-regression.

#### Exact schema contract

`CALL-DUR-02` is decomposed into three source atoms, each owning exactly one new schema file plus one focused descriptor oracle, followed by one separately reviewed migration atom `CALL-DUR-02M`. Connection rows carry immutable call/user/UUID-epoch, state/timestamps/version and unique `(call_id,user_id,epoch)`; presence rows carry immutable call/user, nonnegative connected count/nonnegative sequence and unique `(call_id,user_id)`; delivery rows carry immutable verified provider/delivery ID/kind/call/epoch, a GDPR-erasable user identity, lifecycle timestamps/result and unique `(provider,delivery_id)`. Normal retention is 30 days after connection `left_at` or parent `Call.ended_at`, parent `Call.ended_at` for presence, and delivery `terminal_at`; no atom creates a retention worker. Owner hard-delete may cascade connection/presence earlier, while provider-delivery replay barriers deliberately have no Call/User FK or cascade and retain their normalized non-user identity after GDPR tombstoning. No connection, presence, epoch, delivery or refcount is inferred from historic `CallParticipant`.

Generated Ent output is image-contained mechanical evidence only. All accepted CALL-DUR-01 and prepared CALL-DUR-02A/02B/02C schema sources must enter one shared `ENT-CALL-GEN-01` build context so a single authorized image rebuild generates both Call provider fields and all durable entities. Before that rebuild, root must prove host `backend/ent` contains only tracked `generate.go` and `schema/**`; any ignored, untracked or extra path blocks execution. The receipt freezes all schema/source/generator hashes, build-time generation sentinels and rebuilt image identity. No host Ent output, Git diff manifest or generated-file staging is accepted.

The normal `backend-unit` run overlays only `backend/ent/schema` and consumes generated Ent packages baked into the image; it cannot refresh or prove schema-dependent generated APIs. A source atom may obtain only `SCHEMA_SOURCE_ACCEPTED` from its descriptor oracle and remains `BLOCKED_GENERATOR`; no old-image result is generated-API GREEN or atom completion. The common gate requires a separate founder-authorized rebuild of the existing Docker target whose receipt binds exact source/schema/generator inputs, successful build-time generation and sentinels, rebuilt image identity, before/after resource evidence, and the later approved `run --rm backend-unit` result. Generated output remains image-contained; host generation, ignored-file staging and a Git manifest of those outputs are forbidden. No generator/rebuild command is authorized by this packet.

##### CALL-DUR-02A connection schema source

Exact ownership is new `backend/ent/schema/callparticipantconnection.go` and `backend/tests/calls/phase150_call_participant_connection_schema_test.go`; no existing schema, migration, generated, adapter or runtime file is owned. The descriptor fields are: immutable generated UUID `id`; immutable required UUID `call_id`, `user_id`, `epoch` with no defaults; mutable enum `state=unknown|connected|left` default `unknown`; optional nillable `connected_at` and `left_at`; immutable defaulted `created_at`; default/update-default `updated_at`; and the existing `VersionMixin` with positive default-one `version`. Exact child-owned unidirectional edges are `call` bound to `call_id`, required/unique/immutable with `ON DELETE CASCADE`, and `user` bound to `user_id`, required/unique/immutable with `ON DELETE NO ACTION`; neither adds a reverse parent edge. Indexes are exact unique `(call_id,user_id,epoch)`, partial `left_at` for `state=left`, then `user_id` for hard-delete/GDPR lookup. It defines no legacy `CallParticipant`, provider, delivery, token, raw-event or public-presence field.

The exact timestamp truth table is `unknown=(NULL,NULL)`, `connected=(NOT NULL,NULL)`, and `left=(connected_at nullable, left_at NOT NULL)`. A left row with null `connected_at` is permitted only when later reconciliation resolves a never-confirmed epoch; the direct unknown-left observation remains reconciliation-only and persists no transition. `left_at >= connected_at` is required when both exist.

The external descriptor oracle calls only the schema `Fields`, `Mixin`, `Edges` and `Indexes` APIs. It proves exact types/order/defaults/immutability/nillability, zero-free application intent via required no-default tuple fields, exact state vocabulary, timestamp shapes, VersionMixin identity, the two exact field-bound edge descriptors/on-delete annotations, tuple uniqueness, partial retention index, user lookup index and forbidden-field absence. It uses no source read, generated API, database, filesystem, network or `enttest`; generated FK DDL, SQL checks/index behavior, transitions, CAS, retention and runtime remain `NOT_RUN`.

##### CALL-DUR-02B presence schema source

Exact ownership is new `backend/ent/schema/callparticipantpresence.go` and `backend/tests/calls/phase150_call_participant_presence_schema_test.go`. Fields are immutable generated UUID `id`; immutable required UUID `call_id` and `user_id`; mutable signed-int64 `connected_count` and `presence_sequence`, both default zero with nonnegative validators; immutable defaulted `created_at`; and default/update-default `updated_at`. Exact child-owned unidirectional edges match 02A: required/unique/immutable Call cascade and User no-action, with no reverse parent edit. Indexes are exact unique `(call_id,user_id)` then `user_id`. It has no state, epoch, provider, delivery, legacy `CallParticipant`, outbox or public-event field. `presence_sequence` is per-user public ordering, never provider delivery identity.

For both 02A and 02B, direction is literal child-owned `edge.To("call", Call.Type).Field("call_id").Unique().Required().Immutable()` and `edge.To("user", User.Type).Field("user_id").Unique().Required().Immutable()`, with only the exact `entsql.OnDelete(Cascade|NoAction)` annotation. `edge.From`, `Ref`, reverse parent edits and implicit mutable edges are forbidden.

The presence schema has no mixin; all mutation/CAS semantics belong to later atoms.

The descriptor oracle proves that exact field/edge/index shape and forbidden-field absence only. Count `0 -> 1`, `1 -> 0`, sequence advancement, connection CAS, joined/left publication, generated/SQL FK parity, migration checks, GDPR cleanup, retention and runtime belong to later atoms and remain `NOT_RUN`.

##### CALL-DUR-02C provider-delivery schema source

Exact ownership is new `backend/ent/schema/callproviderdelivery.go` and `backend/tests/calls/phase150_call_provider_delivery_schema_test.go`. Fields are immutable generated UUID `id`; immutable enum `provider=livekit|jitsi`; immutable nonempty max-255 `delivery_id`; immutable enum `kind=connected|disconnected`; immutable required UUID `call_id` and `epoch`; mutable optional nillable UUID `user_id` with no default; mutable enum `result=pending|applied|ignored|reconcile` default `pending`; immutable defaulted `received_at`; and optional nillable `terminal_at`. Normal active-user admission must insert a nonzero user; null is reserved for any terminal GDPR tombstone, while missing-user admission is specifically terminal ignored/null. Indexes are exact unique `(provider,delivery_id)`, `(result,terminal_at)` for terminal cleanup, and `(user_id)` for bounded GDPR lookup. `epoch` is UUID and must never be inferred from delivery order.

The exact lifecycle truth table is `pending|reconcile => terminal_at IS NULL AND user_id IS NOT NULL/nonzero` and `applied|ignored => terminal_at IS NOT NULL AND (user_id IS NULL OR user_id is nonzero)`. Any nonnull user is nonzero. Reconcile is unresolved work and cannot enter retention until it resolves.

This Hub inbox stores only normalized facts after trusted transport/HMAC admission. It has no Call/User/Connection Ent edge or FK, so a Call hard-delete cannot erase the replay barrier early. Admission and erasure must use one User-first lock order: admission takes a PostgreSQL `FOR KEY SHARE` lock on the present User before inserting a nonzero identity; if a concurrent or completed deletion's `FOR UPDATE` wins and the User is then absent, admission applies no presence effect and stores the delivery directly as terminal ignored with null user and one transaction time. The existing GDPR transaction takes `FOR UPDATE` on the User as its first database operation before deleting matching connection/presence rows, changing matching pending/reconcile deliveries to ignored with that transaction time, clearing `user_id` on every matching delivery, preserving provider/delivery ID/kind/call/epoch/received facts, and only then hard-deleting the User. This serialization prevents UUID reintroduction after cleanup; any failure rolls back every step. It stores no raw envelope, payload, digest, HMAC/MAC, key/version, certificate, room, raw participant identity, JWT, token, error text, trace or outbox ID. The source-side spool remains exact-byte retry authority. The descriptor oracle proves only schema shape, the sole dedupe identity and forbidden-field absence; HMAC, parsing, replay, apply, result transition, locking/GDPR transaction behavior, retention and runtime remain `NOT_RUN`.

The delivery schema has no mixin.

##### CALL-DUR-02M owner-applied PostgreSQL expansion

After 02A/02B/02C source acceptance, `CALL-DUR-02M` alone extends `backend/migrations/20260830_calls_durable_presence_expand.sql`; it has no unit oracle and requires independent schema/security review plus later integration evidence. It is the one-shot owner-applied DDL realization for existing PostgreSQL, while generated 02A/02B edges own the same clean-database relation contract. It creates the three tables with descriptor enums/indexes, insert-only `now()` defaults for created/updated/received timestamps, no update trigger, nonzero UUID checks, positive connection version, nonnegative presence count/sequence, connection state/timestamp consistency, and delivery result/terminal/user consistency. Connection and presence receive `call_id -> calls(id) ON DELETE CASCADE`, `user_id -> users(id) ON DELETE NO ACTION`, plus explicit user lookup indexes; delivery deliberately receives no Call/User FK but has a nullable user lookup column/index for GDPR tombstoning. Normal cleanup uses left-or-ended, ended, and terminal timestamps respectively; no purge worker is owned. SQL parse/apply/rollback/replay, generated clean-DB parity and live PostgreSQL remain `NOT_RUN`.

02M appends inside the same single `BEGIN ... COMMIT` transaction created by CALL-DUR-01, after its four Call alterations and before `COMMIT`. Every table, constraint and index statement is plain fail-closed DDL: no `IF NOT EXISTS`, catalog guard, constraint drop, nested transaction or compensating statement. Any collision or replay aborts and rolls back the complete 01+02M transaction for owner triage.

Literal tables are `call_participant_connections`, `call_participant_presences` and `call_provider_deliveries`. IDs and business IDs are `UUID` with primary-key/required/nonzero checks as applicable; delivery `user_id UUID NULL` accepts only null or nonzero; state/provider/kind/result and delivery ID use bounded `VARCHAR` plus exact checks; connection `version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)`; presence counters `BIGINT NOT NULL DEFAULT 0 CHECK (value >= 0)`; timestamps are `TIMESTAMPTZ`, with `now()` only on required insert timestamps. Exact indexes are connection unique `(call_id,user_id,epoch)`, partial `(left_at) WHERE state='left'`, and `(user_id)`; presence unique `(call_id,user_id)` and `(user_id)`; delivery unique `(provider,delivery_id)`, `(result,terminal_at)`, and `(user_id)`. Delivery gets no `call_id` index or FK in this atom.

Those consistency checks implement the exact connection timestamp truth table plus ordering check and the exact delivery result/terminal/user truth table above. Delivery cleanup applies only to terminal `applied|ignored`; unresolved `reconcile` rows are retained.

After `ENT-CALL-GEN-01`, a separate PostgreSQL integration/parity gate must compare generated clean-DB and owner-migration tables by columns, FK targets/actions and index semantics, not accidental constraint names. Before any durable-mode activation, admission and erasure must implement the User-first serialization above, and the existing GDPR transaction must atomically delete connection/presence rows, terminalize matching pending/reconcile deliveries as ignored, null every matching delivery `user_id`, and then hard-delete the User. All operations share one transaction timestamp and any failure rolls back; until this exact behavior exists, User `NO ACTION` intentionally fails closed instead of silently erasing linked history. Concurrent PostgreSQL integration cases must prove ingress-before-delete is observed and erased, ingress-after-delete-lock applies no effect and stores only ignored/null, and replay of an existing tombstone neither restores the UUID nor reapplies an effect.

The serial source queue after accepted CALL-DUR-03 is `CALL-DUR-02A -> CALL-DUR-02B -> CALL-DUR-02C -> CALL-DUR-02M -> ENT-CALL-GEN-01`. The three 02A/02B/02C descriptor tests stay ignored outside discovery until each becomes the sole accepted missing-symbol oracle. Their RED may contain only the missing schema type, and old-image GREEN can yield only `SCHEMA_SOURCE_ACCEPTED`. 02M has no unit oracle and awaits its independent schema/security review plus later integration evidence. The common founder-authorized build is the first generated-API gate and must include every accepted 01/02 schema source in one image receipt.

#### Jitsi discovery and surface ordering

- `JITSI-D1` is a hard prerequisite for any provider-hook attachment or executable overlay. It inspects one exact digest at a time and freezes the supported hook, image user, mounts, writable paths, health, JWT component, effective assets, SBOM and license data.
- `JITSI-02A` may define only the provider-neutral envelope/HMAC verifier and negative `backend-unit` tests before D1; it attaches to no Jitsi image or route. Exact ownership is `backend/internal/calls/provider_event_hmac.go` plus `backend/tests/calls/phase150_provider_event_hmac_contract_test.go`. The only production API is `ExportPhase150VerifyProviderEventEnvelopeHMACSHA256(secret, rawEnvelope, receivedMAC []byte) (Phase150VerifiedProviderEventEnvelope, error)` plus `Phase150VerifiedProviderEventEnvelope.Bytes() []byte`; the verified type is a struct with no exported fields, keeps its byte slice private, and both the successful result and `Bytes()` are defensive copies. One non-nil sentinel `ErrPhase150ProviderEventEnvelopeInvalid` covers every rejected input or authentication failure. Validation order is secret length at least 32 bytes, raw envelope length `1..65536` bytes, already-decoded binary MAC length exactly 32 bytes, then constant-time HMAC-SHA256 comparison over the exact unmodified raw envelope bytes. Every failure returns the zero verified value and zero-value `Bytes()` returns nil without panic. No trim, hex/base64/header decode, normalize, decompress, JSON parse, delivery-ID extraction, clock/skew decision, replay state or log is allowed. Neither the value nor its pointer may implement `fmt.Stringer`, `encoding.TextMarshaler` or `json.Marshaler`.
- `JITSI-02B0` is a design/evidence gate, not code. Its target is a dedicated self-hosted provider-side PostgreSQL database reached through a Prosody named SQL store, with separate database/schema/role/credentials/lifecycle from Hub application PostgreSQL. It freezes retention, capacity, ownership, backup and crash/restart behavior. D1 proves only that the pinned image contains the required SQL/archive driver and hook/config surface. A separate founder-authorized JITSI-02D fault-injection gate must prove that successful append acknowledgement follows a durable database commit. Failure of either gate leaves B0 blocked. `memory|none|null`, Hub application databases, MUC room state, Prosody internal files, SQLite, and any unverified container filesystem are rejected as the source of truth.
- `JITSI-02B1`, after D1/B0, owns only the selected MUC hook filter/map and durable append. It rejects focus/system/recorder/JVB identities, anonymous/guest principals, and rooms not exactly mapped to persisted `Call.RoomName`. The exact selected-image mechanism must disable breakouts for the baseline 1:1 contour or a separate main/breakout mapping is required. Room destruction never implies Hub Call end; CALL-DUR-09 remains authority.
- `JITSI-02B2` owns only spool dispatch/retry/restart and must replay the same delivery ID and exact stored bytes without JSON re-encoding.
- `JITSI-02C1` owns only the separately verified mTLS sidecar byte-forwarding and client-certificate policy; `JITSI-02C2` owns only Hub ingress mTLS/HMAC/replay admission.
- `JITSI-02D` owns the isolated durable-write-before-send, restart/replay, byte-identity, wrong-cert, skew/size/content-type/tamper and crash matrix. No sub-atom may claim stock Jitsi webhook durability.

`JITSI-02B0/B1/B2/C1/C2/D` are `BLOCKED_NONDISPATCHABLE`: the approved Hub QA services do not execute the custom Lua module, selected spool, or sidecar. Root may dispatch them only after JITSI-D1 and a separate founder-authorized task-owned Docker launcher/service with exact RED/GREEN oracles. No command is invented or authorized here.

`JITSI-02A` uses the exact approved launcher `docker compose -f docker-compose.local.yml run --rm backend-unit` with no selector, pull or build. Its RED is valid only when production build passes and vet/unit/race fail solely on the missing frozen symbols above. Its test proves valid exact-byte authentication; nil-slice rejection through the same non-nil sentinel; body boundaries `0/1/65536/65537`; secret boundaries `31/32/33`; MAC boundaries `31/32/33`; one-byte/whitespace/NUL mutation rejection; hex/base64 text rejection where a decoded 32-byte MAC is required; correctly signed leading/trailing whitespace/NUL, JSON-looking invalid and gzip-looking opaque byte acceptance without parsing or normalization; zero-value failure results; zero-value `Bytes()` safety; private struct shape with no exported fields or value/pointer output interfaces; and both input-to-result and per-`Bytes()` alias isolation. HTTP header decoding, streaming size enforcement before allocation, content type, mTLS peer identity, envelope parsing, durable replay/dedupe and business mutation remain `JITSI-02C1/C2/D`; this atom cannot establish provider compatibility, delivery durability or runtime readiness.

Target provider-source production semantics are `at-least-once after successful durable enqueue`, not zero-loss observation, and may be claimed only after B1/B2/D evidence. A crash before hook append or database unavailability can lose a post-event observation because the MUC state change is already committed. Production acceptance therefore also requires a bounded reconciliation design or separately proven pre-event fail-closed gate; metrics alone cannot relabel this gap as durable delivery.
- `JITSI-03A` is `BLOCKED_NONDISPATCHABLE` until the sole-oracle queue and accepted dependencies complete in this order: `MPR-01 -> JITSI-02A -> CALL-DUR-01 -> CALL-DUR-03 -> CALL-DUR-02A -> CALL-DUR-02B -> CALL-DUR-02C -> CALL-DUR-02M -> ENT-CALL-GEN-01 -> CALL-DUR-02G(SOURCE_ACCEPTED) -> CALL-DUR-08A -> CALL-DUR-08B -> CALL-DUR-08C -> CALL-DUR-08D -> CALL-DUR-08E`. Root must then freeze a separate exact `TaskContract`, transfer exactly one discovered missing-behavior oracle, and define expected Docker RED/GREEN plus legacy `/join` parity before dispatch. Future ownership is exactly `backend/internal/auth/middleware.go`, `backend/internal/calls/handler.go`, `backend/internal/calls/handler_calls.go`, `backend/internal/calls/models.go`, `backend/cmd/server/wiring_integrations.go`, `backend/tests/auth/phase150_interactive_session_middleware_test.go`, `backend/tests/calls/phase150_join_credential_handler_test.go`, `backend/docs/docs.go`, `backend/docs/swagger.json`, `backend/docs/swagger.yaml`, and `docs/wiki/API.md`. It consumes the existing `IssueJoinCredentialForAuthenticatedUser` method; `service.go` is not edited without another frozen contract.

  The recommended additive API is `POST /calls/:callId/join-credential`. After global API-key/generic auth and `RequireInteractiveSession` succeed, any nonempty decoded body, nonzero `Content-Length`, or `Transfer-Encoding` is `400` before Calls service work. A complete success has no extra top-level fields: canonical UUID `call_id` equal to the path; closed `provider=livekit|jitsi` equal to the immutable persisted Call provider; server-derived nonempty `room_name` and exact validated `service_url`; nonempty opaque `token`; and nonzero RFC3339 UTC `expires_at`. Success sets `Cache-Control: no-store`. `CallResponse`, legacy `/join`, `/guest-join`, `/invite/:code`, the existing LiveKit guest flow, and LiveKit UI remain unchanged.

  Dedicated `RequireInteractiveSession` runs after global `APIKeyAuth` and the generic Calls `authMw`, independently re-parses and verifies the Bearer header, requires structurally valid non-nil UUID `uid` and `sid`, repeats the active-account check, and never treats Fiber Locals as authority. No active persisted-session lookup or revocation check is claimed; `sid` is structural only. Invalid or revoked API keys retain the global `401`; current API-key precedence makes a valid API-key-only or valid API-key-plus-Bearer request reach the dedicated boundary and receive `403` before Calls service. Without an API key, missing, malformed or expired Bearer and invalid `uid`/`sid` are `401`; account-status dependency failure is pre-service `503`. Malformed `callId` is `400`, Call authorization denial is `403`, and remaining Calls projection/repository/config/issuer failures are `500`; Calls-level `404`/`503` wait for a separately accepted non-enumerating taxonomy.

  The handler calls only `IssueJoinCredentialForAuthenticatedUser`. `CALL-DUR-08C` may allocate the required unknown connection epoch; credential issuance cannot call `AddParticipant`, publish joined/confirmed-presence/outbox effects, or return any credential field on error. Token, room and service URL are forbidden in errors, logs, audit and telemetry. `JITSI-03A` makes no guest-path change: accepted `CALL-DUR-08B/08D` must already deny a Jitsi-backed legacy guest request before provider credential issuance while retaining LiveKit guest parity. Three-agent architecture council alignment is recorded as decision input, but formal verifier provenance is `UNVERIFIED` and is not implementation evidence. Client/endpoint implementation remains `ABSENT/UNBOUND`; OpenAPI generation, execution, Docker and runtime are `NOT_RUN`. `JITSI-03B` later owns `frontend/src/api/calls.ts`, `frontend/src/stores/call-store.ts`, one new `JitsiCallSurface.tsx`, `CallRoom.tsx`, and focused frontend tests for exact self-hosted origin, one-provider rendering, dispose/unmount and token non-disclosure.

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

### CALL-DUR-08C execution checkpoint - 2026-09-01

`CALL-DUR-08C1` is `SOURCE_ACCEPTED` in commit `7b1d63fa`. The implementation is confined to the three frozen production files and the sole external oracle: an active canonical server-loaded Call, matching `ProviderParticipantIdentity`, and exact canonical server-known name are required; Jitsi signs only the provider participant canonical identity; the legacy UUID/padded-name issuer remains compatible; and the wrapper adds only pure delegation. No registry capability, provider activation, factory/config/service/server/HTTP/client binding, guest/lifecycle behavior, DB access or routing was added.

The accepted launcher receipt is deliberately qualified `USER_AUTHORIZED_CONCURRENT_MAIL_DEV_INFRA_GREEN`: the founder explicitly allowed execution beside the active Mail development stack, no Mail QA/test/build runner overlapped, and one below-floor RAM sample immediately recovered. The earlier exit-zero run with Mail `python-api` overlap is rejected. `CALL-DUR-08C2` remains the next serial atom, followed by 08C3; allocator/transaction/persistence and every runtime/deployment/browser gate remain `UNBOUND/NOT_RUN`. Therefore `CALL-DUR-08C` and Phase 150 are not complete.
