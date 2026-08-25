# Self-Hosted Jitsi Static Foundation and Provider Boundary Plan

> **Execution rule:** Implement with `superpowers:subagent-driven-development`. Production tasks are sequential Terra/XHigh atomic writes with Terra/High independent review. No two writers own the same file concurrently.

**Phase goal:** Establish a provider-neutral Calls boundary, preserve LiveKit as the only active provider, add a server-only Jitsi credential primitive, and freeze a self-hosted Jitsi deployment preflight. This phase does not by itself deliver an accepted or production-ready Jitsi deployment.

**Current phase verdict:** `PARTIAL — STATIC_FOUNDATION_ONLY`. The overlay is `NOT_CREATED / UNBOUND` until the separately authorized post-discovery gate proves image users, UID/GID requirements, writable config paths, and branding/asset paths on Linux. Compose render, image pull/start, TLS, JWT integration, UDP media, restart, rollback, effective `external_api.js`, and SBOM remain `NOT_RUN`. Planning, env contracts, or draft YAML must not increase Jitsi deployment readiness beyond the already reported static-foundation baseline.

**Active provider:** `CALLS_ACTIVE_PROVIDER=livekit`. Selecting `jitsi` fails closed until Phase 152. LiveKit public DTO fields, webhook, `/lk/*`, recording/Egress, dependencies, and rollback assets remain intact.

**Authoritative inputs:** `150-SPEC.md`, immutable `150-RESEARCH.md`, `.planning/decisions/ADR-COMMS-OSS-REPLACEMENT-2026-08-24.md`, the [official Jitsi Docker guide](https://jitsi.github.io/handbook/docs/devops-guide/devops-guide-docker/), [official docker-jitsi-meet repository](https://github.com/jitsi/docker-jitsi-meet), and official Jitsi iframe API documentation. `150-RESEARCH.md` is read-only evidence and is not owned by any implementation task.

## Global constraints

- Official release is exactly `stable-11146-2`; any eventual core image reference must equal the four immutable OCI index digests already recorded in `150-RESEARCH.md`.
- No JaaS, `meet.jit.si`, external Calls API/host, floating image tag, shared Hub database, external STUN/TURN, external analytics, or upstream source fork.
- Canonical origin is exactly `JITSI_PUBLIC_URL=https://${JITSI_HOST}`. It must be HTTPS with a nonempty host and no userinfo, non-root path, query, fragment, or explicit port.
- One XMPP subject source exists: backend `JITSI_XMPP_DOMAIN`; eventual overlay passes `XMPP_DOMAIN=${JITSI_XMPP_DOMAIN:?required}`; JWT `sub` equals the same value. An independent `XMPP_DOMAIN` input is forbidden.
- There is no `JITSI_ISSUER` or `JITSI_AUDIENCE`. JWT `iss=JITSI_APP_ID`; `aud` is exactly the JSON array `["jitsi"]`.
- `JITSI_APP_SECRET`, `JICOFO_AUTH_PASSWORD`, and `JVB_AUTH_PASSWORD` are separate required secrets with no usable default. Examples contain only an explicitly rejected marker such as `__REQUIRED_UNUSABLE_NO_DEFAULT__`; empty, example, marker, reused, short, or known dev values fail acceptance. `JITSI_APP_SECRET` is at least 32 bytes.
- Do not edit Mail/Calendar, `frontend/src/modules/calls/create-call-dialog.tsx`, or `pro-qa-ai/**`.
- All tests/build/typecheck/lint are Docker-only through accepted exact services. No host tests or validators. No Jitsi image pull, Compose render, build, or container start before explicit infrastructure authorization.
- No invented container `user`, numeric UID/GID, `read_only`, `tmpfs`, writable path, or branding mount. Each must be backed by the selected-image Linux discovery receipt before it enters accepted YAML.
- Secrets/tokens must not appear in logs, health endpoints, browser configuration, public DTOs, snapshots, or test output.

## Frozen backend contracts

```go
type CallsConfig struct {
    ActiveProvider string
    Jitsi          JitsiConfig
}

type JitsiConfig struct {
    Enabled    bool
    PublicURL  string
    AppID      string
    AppSecret  string
    XMPPDomain string
}
```

When `JITSI_ENABLED=true`, canonical `PublicURL`, nonempty `AppID`, non-placeholder `AppSecret` of at least 32 bytes, and nonempty `XMPPDomain` are mandatory. These checks are not weakened by development mode. Unknown provider values fail closed; explicit `jitsi` returns stable text containing `jitsi provider is not available before phase 152`.

The issuer method boundary is exact:

```go
Issue(call *ent.Call, authenticatedUserID uuid.UUID, serverKnownUserName string) (string, error)
```

`call` is a server-loaded Ent entity, not a client DTO. It must be nonnil, have a nonzero ID, and be active. `authenticatedUserID` must be nonzero; `serverKnownUserName` must be nonempty after trimming. The room claim is exactly `call.ID.String()` and is never accepted as an argument. Claims are HS256 only, `iss=JITSI_APP_ID`, `aud=["jitsi"]`, `sub=JITSI_XMPP_DOMAIN`, `iat`, `exp=iat+5m`, exact `room`, and `context.user.id/name`. No wildcard, moderator, features, guest policy, or arbitrary claims exist.

This issuer is unreachable in Phase 150: no HTTP route, WebSocket subject, Handler method, public DTO, frontend token request, or active-provider wiring. Its existence does not prove foreign-call membership. Caller-to-call membership, rate limiting, replay policy, and the public credential endpoint are Phase 152 gates.

## Low-disk test cadence

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

Write all failing tests before production changes, then perform the single backend RED run. Coverage includes typed provider methods; LiveKit behaviour parity; create/rollback/guest/participant-state injection; provider defaults and fail-closed values; canonical origin and secret validation; factory rejection; and issuer claims/negative binding.

Issuer negatives include nil call, inactive call, zero call ID, zero authenticated user ID, blank server-known name, substituted Call entity/ID, substituted authenticated user, client-room impossibility, wrong signing algorithm, wrong secret, expired token, non-five-minute lifetime, and absence of secret/token output. Parsing tests never log credentials.

## Task 1: Provider interface contract

**Exclusive file:** `backend/internal/calls/provider.go`.

Create only provider-neutral types and the narrow `ConferenceProvider` interface. It contains no config reads, LiveKit construction, Jitsi construction, routing, or public DTO change. Review interface direction and typed credential semantics against `provider_interface_contract_test.go`. Do not run the full suite here.

## Task 2: LiveKit adapter extraction

**Exclusive files:**

- `backend/internal/calls/livekit_provider.go`
- `backend/internal/calls/service_room.go` (source behaviour moves, then this file is deleted)

Move existing room provision/delete, user/guest credential, service URL, and participant metadata behaviour into `LiveKitConferenceProvider` without changing grants, expiry, errors, webhook/recording semantics, or compatibility output. Do not edit service wiring. Review against the frozen adapter test; no full-suite run.

## Task 3: Call service provider injection

**Exclusive file:** `backend/internal/calls/service.go`.

Make `CallService` depend on `ConferenceProvider`. Preserve persistence order and rollback, unwrap typed credentials only at the compatibility DTO boundary, and preserve chat/events/recording behaviour. No config selection or server wiring is owned here. Review against injection tests; no full-suite run.

## Task 4: Calls and Jitsi configuration validation

**Exclusive files:**

- `backend/internal/platform/config/config.go`
- `deployments/.env.example`
- `docker-compose.local.yml`

Add exact `CallsConfig`/`JitsiConfig` parsing and validation. Default active provider to LiveKit. Enabled Jitsi requires canonical HTTPS origin, AppID, XMPP domain, and separate non-placeholder AppSecret of at least 32 bytes in every mode. Document only required backend variables; do not add Jitsi services. `JITSI_XMPP_DOMAIN` is the sole source later mapped to deployment `XMPP_DOMAIN` and JWT `sub`. No suite run.

## Task 5: Provider factory and server wiring

**Exclusive files:**

- `backend/internal/calls/provider_factory.go`
- `backend/cmd/server/wiring_integrations.go`

Create a factory which constructs LiveKit only. `jitsi` returns the frozen Phase-152 error; all unknown values fail closed. Wire the factory into the server while leaving LiveKit recording, webhook, `/lk/*`, and rollback assets intact. No issuer construction or transport is allowed. Review wiring statically; no suite run.

## Task 6: Server-only Jitsi credential issuer

**Exclusive files:**

- `backend/internal/calls/jitsi_credential_issuer.go`
- `backend/internal/calls/export_hooks.go` only if the external `hubtest` package cannot exercise the issuer without a tagged test hook

Implement the exact `Issue(*ent.Call, uuid.UUID, string)` boundary and claims above using `github.com/golang-jwt/jwt/v5`. A required test hook must be build-tagged/test-only, constructor-shaped, and explicitly not a transport or public product API. It must not wire the issuer to server startup, handlers, routes, WebSocket, DTOs, or frontend.

After independent review, run the single batched backend GREEN gate. Record PASS/FAIL from the Docker receipt only. Membership authorization and Jitsi activation remain unimplemented by design.

## Task 7: Deployment preflight and static env contract

**Exclusive files:**

- `deployments/jitsi/.env.example`
- `deployments/jitsi/DEPLOYMENT-PREFLIGHT.md`

Create only the preflight/static contract. It lists canonical `JITSI_HOST`, `JITSI_PUBLIC_URL`, `JITSI_XMPP_DOMAIN`, `JITSI_APP_ID`, mandatory `JVB_ADVERTISE_IPS`, XMPP component domains, and the three separate required secret inputs. Secret examples use only the rejected `__REQUIRED_UNUSABLE_NO_DEFAULT__` marker and state `required/no default`.

Record exact equality: overlay `XMPP_DOMAIN` is populated from `JITSI_XMPP_DOMAIN`, and issuer JWT `sub` uses the same backend value. Freeze the four digest receipts by reference to immutable `150-RESEARCH.md` without editing it.

The preflight matrix must leave these rows `UNBOUND / NOT_RUN`: effective image user, UID/GID, config volume ownership, writable directories, rootless compatibility, branding mount path, final effective `external_api.js`, image SBOM, Caddy/TLS, JVB advertise reachability, UDP media, restart, and rollback. At the end of Task 7, `deployments/docker-compose.jitsi.yml` remains `NOT_CREATED`; no overlay is claimed accepted.

## Task 8: GATED post-discovery overlay — not executable now

**Gate status:** `BLOCKED_BY_AUTHORIZED_LINUX_DISCOVERY`; do not dispatch during the current low-disk/static phase.

**Future exclusive files after gate opens:**

- `deployments/docker-compose.jitsi.yml`
- `deployments/Caddyfile.prod`
- `deployments/docker-compose.prod.yml`

Required gate input is an authorized task-owned Linux receipt for the exact four digest-pinned images proving effective users, required UID/GID handling, writable config directories, ownership transition, branding/static-asset path, healthcheck commands, and effective self-hosted `external_api.js`. Discovery must obey an explicit low-disk budget. Until that receipt exists, overlay YAML is `NOT_CREATED / UNBOUND`; a speculative file may only be labelled `DRAFT_NON_ACCEPTED` and cannot satisfy Phase 150.

After the gate opens, map `XMPP_DOMAIN=${JITSI_XMPP_DOMAIN:?required}` and all three secrets with `${VAR:?required}` and no fallback. Preserve Caddy ports 80 and 443 for ACME HTTP-01/HTTPS. The only extra published port is JVB `10000:10000/udp`, with `JVB_ADVERTISE_IPS=${JVB_ADVERTISE_IPS:?required}`. Use explicit control/edge/media networks and the four frozen digests. Do not add `user`, `read_only`, `tmpfs`, config/branding mounts, or healthchecks unless each exact setting is proven by the discovery receipt.

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
