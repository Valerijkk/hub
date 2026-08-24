# Self-Hosted Jitsi Deployment and Provider Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` to implement this plan task-by-task. Each task uses a fresh Terra/XHigh atomic writer and an independent Terra/High reviewer.

**Goal:** Add an immutable self-hosted Jitsi shadow deployment and a provider-neutral Calls seam while LiveKit remains the active, reversible provider.

**Architecture:** Decouple the Hub call domain from LiveKit with a narrow `ConferenceProvider` implemented first by the current LiveKit behavior. Add a separate Jitsi Compose overlay using the official `stable-11146-2` core images pinned by OCI index digest, private component networks, local JWT auth, and a dedicated `meet.<Hub domain>` edge origin. Phase 150 does not issue Jitsi join credentials or change the public Calls DTO/UI; those changes are Phase 152, while recording and LiveKit removal remain Phase 154.

**Tech Stack:** Go 1.25, LiveKit server SDK (shadow adapter only), Jitsi Meet Docker distribution, Prosody, Jicofo, JVB, Docker Compose, Caddy, Hub configuration package, Docker-only Go units.

**Spec:** `.planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-SPEC.md`

## Global Constraints

- Binding decisions: `.planning/decisions/ADR-COMMS-OSS-REPLACEMENT-2026-08-24.md`, `150-CONTEXT.md`, and immutable receipt `150-RESEARCH.md`.
- Official upstream release is exactly `stable-11146-2`; core image references must equal the four `name@sha256` values in `150-RESEARCH.md`.
- No JaaS, `meet.jit.si`, cloud API, floating `stable`/`latest`, external hosting, shared Hub database, or deep upstream source fork.
- Jitsi public origin is exactly `https://meet.<Hub deployment domain>`; configuration relation is `JITSI_HOST=meet.<Hub deployment domain>` and `JITSI_PUBLIC_URL=https://${JITSI_HOST}`.
- Phase 150 keeps `CALLS_ACTIVE_PROVIDER=livekit`. `jitsi` as active provider fails closed until Phase 152.
- Do not delete or change the compatibility meaning of public `token`, `livekit_url`, LiveKit webhook, `/lk/*`, recording/Egress code, historical recordings, or LiveKit dependencies.
- Do not edit Mail/Calendar or `frontend/src/modules/calls/create-call-dialog.tsx`. Do not read or edit `pro-qa-ai/**`.
- Tests/build/typecheck/lint run only through accepted Docker services. No Windows-host Go/Node/validator execution.
- Do not bring up or pull Jitsi images during initial implementation unless a task-owned infrastructure run is explicitly authorized. Digest lookup metadata already caused no layer pull; runtime remains `NOT_RUN` until executed.
- Docs sync and exact-path commits are mandatory; no AI attribution trailers.

## File and Interface Map

| File | Responsibility in this phase |
|---|---|
| `backend/internal/calls/provider.go` | Provider-neutral interface and typed capability/session primitives. |
| `backend/internal/calls/livekit_provider.go` | Current room/token/metadata behavior as the shadow adapter. |
| `backend/internal/calls/provider_factory.go` | Config-driven provider selection; Phase 150 can construct only LiveKit and rejects Jitsi. |
| `backend/internal/calls/service.go` | Depend on `ConferenceProvider`, not LiveKit config. |
| `backend/internal/calls/service_room.go` | Deleted after its implementation moves without behavioral change. |
| `backend/internal/calls/export_hooks.go` | Test constructors for explicit provider injection. |
| `backend/cmd/server/wiring_integrations.go` | Construct LiveKit adapter and inject it; recording/webhook wiring remains LiveKit-specific. |
| `backend/internal/platform/config/config.go` | Add `CallsConfig`/`JitsiConfig` and fail-closed validation. |
| `backend/tests/calls/provider_contract_test.go` | Provider injection and rollback units. |
| `backend/tests/platform/config/calls_provider_config_test.go` | Provider/Jitsi configuration units. |
| `deployments/docker-compose.jitsi.yml` | Opt-in Jitsi shadow topology and pinned images. |
| `deployments/jitsi/.env.example` | Non-secret local/prod environment contract. |
| `deployments/jitsi/THIRD-PARTY-NOTICES.md` | Upstream source, release, license, digest, and SBOM evidence boundary. |
| `deployments/Caddyfile.prod` | Dedicated Jitsi origin; existing Hub and `/lk/*` blocks remain. |
| `deployments/.env.example` | Hub/backend provider and Jitsi shadow variables. |
| `docs/wiki/DEPLOY.md`, `docs/TECHNICAL.md` | Product/deployment contract. |
| `docs/runbooks/jitsi-shadow-deployment.md` | Bring-up, health, rollback, upgrade, and evidence boundaries. |

## Preflight Interaction Table

| Producer task | Consumer task | Shared interface | Ruling |
|---|---|---|---|
| Task 1 | Task 2 | `ConferenceProvider` versus `CallsConfig.ActiveProvider` | Config selects only a provider factory; domain service never reads provider strings. |
| Task 2 | Task 3 | `JitsiConfig` and deployment env | Names/validation are frozen before Compose uses them: `JITSI_ENABLED`, `JITSI_PUBLIC_URL`, `JITSI_APP_ID`, `JITSI_APP_SECRET`, `JITSI_ISSUER`, `JITSI_AUDIENCE`. |
| Task 3 | Task 4 | topology, secrets, ports, image receipts | Runbook repeats exact contracts and distinguishes static receipt from runtime evidence. |
| Every task | Phase 152 | active-provider gate | No task sets `CALLS_ACTIVE_PROVIDER=jitsi` or changes public Calls DTO/UI. |
| Every task | Phase 154 | LiveKit/recording assets | No deletion, dependency cleanup, recording migration, or `/lk/*` removal. |

### Task 1: Introduce `ConferenceProvider` with a behavior-preserving LiveKit adapter

**Files:**

- Create: `backend/internal/calls/provider.go`
- Create: `backend/internal/calls/livekit_provider.go`
- Delete: `backend/internal/calls/service_room.go`
- Modify: `backend/internal/calls/service.go`
- Modify: `backend/internal/calls/export_hooks.go`
- Modify: `backend/cmd/server/wiring_integrations.go`
- Create: `backend/tests/calls/provider_contract_test.go`
- Modify: applicable existing genuine units under `backend/tests/calls/**`

**Interfaces:**

Create this exact domain interface:

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
```

`LiveKitConferenceProvider` implements it with the existing SDK calls and errors. Change constructor to:

```go
func NewCallService(
    provider ConferenceProvider,
    repo *EntRepository,
    nc *natsclient.Client,
    chatRepo *ChatRepository,
) *CallService
```

The `CallService` field is `provider ConferenceProvider`. Public `CallResponse.Token` continues to carry the returned compatibility credential in this phase.

- [ ] **Step 1: Write failing provider injection and rollback units**

In `provider_contract_test.go`, define a local fake implementing all five methods and recording calls. Add:

```go
func TestCallServiceUsesInjectedProviderForCreateAndCredential(t *testing.T)
func TestCallServiceDeletesProvisionedConferenceWhenPersistenceFails(t *testing.T)
func TestCallServiceUsesProviderForGuestCredential(t *testing.T)
func TestCallServiceUsesProviderForParticipantState(t *testing.T)
```

Tests assert the exact room name flows through provision, typed credential, state, and rollback without a network client. The fake returns `JoinCredential{Provider: ProviderLiveKit, RoomName: roomName, ServiceURL: "http://livekit:7880", Token: "test-token"}`; public compatibility responses still expose only the existing token/LiveKit URL fields. Existing LiveKit-specific webhook/recording tests remain unchanged.

- [ ] **Step 2: Run the Docker red gate**

Run:

```powershell
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: compile/test FAIL because `ConferenceProvider` and provider-injected constructor do not exist.

- [ ] **Step 3: Move behavior without changing product semantics**

Move the LiveKit SDK implementation into `livekit_provider.go` and map:

- `Provision` -> current `CreateRoom`;
- `Delete` -> current `DeleteRoom`;
- `IssueJoinCredential` -> current `GenerateToken` with unchanged 24-hour grant, wrapped in typed `JoinCredential` using `ProviderLiveKit` and configured LiveKit URL;
- `IssueGuestCredential` -> current guest UUID/token behavior wrapped in `GuestJoinCredential`;
- `UpdateParticipantState` -> current participant metadata update.

Change helper signatures (`provisionRoom`, `finalizeCreate`, `rollbackCreate`) to accept `ConferenceProvider`. `CallService.GenerateToken` unwraps `JoinCredential.Token`, and `GuestJoin` unwraps identity/token/service URL into the existing compatibility DTO. In Task 1 wiring, construct `callsmod.NewLiveKitConferenceProvider(cfg.LiveKit)` directly and pass it to `NewCallService`; Task 2 replaces that temporary construction with the config-driven factory. Keep `NewRecordingService`, `NewHandler`, webhook config, and frontend response unchanged.

- [ ] **Step 4: Run backend regression to green**

Run `backend-unit`. Expected: build/vet/unit/race PASS and existing Calls behavior preserved.

- [ ] **Step 5: Commit exact paths**

```powershell
git add backend/internal/calls/provider.go backend/internal/calls/livekit_provider.go backend/internal/calls/service.go backend/internal/calls/service_room.go backend/internal/calls/export_hooks.go backend/cmd/server/wiring_integrations.go backend/tests/calls/provider_contract_test.go backend/tests/calls
git commit -m "refactor(calls): isolate conference provider boundary"
```

### Task 2: Add fail-closed Calls/Jitsi configuration without enabling cutover

**Files:**

- Modify: `backend/internal/platform/config/config.go`
- Create: `backend/internal/calls/provider_factory.go`
- Modify: `backend/cmd/server/wiring_integrations.go`
- Create: `backend/tests/platform/config/calls_provider_config_test.go`
- Modify: `backend/tests/calls/provider_contract_test.go`
- Modify: `deployments/.env.example`
- Modify: `docker-compose.local.yml`
- Modify: `deployments/docker-compose.prod.yml`

**Interfaces:**

Add:

```go
const CallsProviderLiveKit = "livekit"

type CallsConfig struct {
    ActiveProvider string
    Jitsi          JitsiConfig
}

type JitsiConfig struct {
    Enabled   bool
    PublicURL string
    AppID     string
    AppSecret string
    Issuer    string
    Audience  string
}
```

`Config` gains `Calls CallsConfig` and retains `LiveKit LiveKitConfig`. Env mapping is exact:

```text
CALLS_ACTIVE_PROVIDER=livekit
JITSI_ENABLED=false
JITSI_PUBLIC_URL=https://meet.<Hub deployment domain>
JITSI_APP_ID=hub
JITSI_APP_SECRET=<secret, minimum 32 characters outside dev>
JITSI_ISSUER=hub
JITSI_AUDIENCE=jitsi
```

- [ ] **Step 1: Write failing table-driven config units**

Cases:

```go
func TestLoadCallsProviderDefaultsToLiveKit(t *testing.T)
func TestLoadRejectsUnknownCallsProvider(t *testing.T)
func TestLoadRejectsPrematureJitsiActiveProvider(t *testing.T)
func TestLoadRejectsIncompleteEnabledJitsiInProduction(t *testing.T)
func TestLoadRejectsNonHTTPSJitsiPublicURLInProduction(t *testing.T)
func TestLoadAcceptsCompleteJitsiShadowWithLiveKitActive(t *testing.T)
func TestConferenceProviderFactorySelectsLiveKit(t *testing.T)
func TestConferenceProviderFactoryRejectsJitsiBeforePhase152(t *testing.T)
```

The premature case must fail with a stable message containing `jitsi provider is not available before phase 152`.

- [ ] **Step 2: Run the Docker red gate**

Run `backend-unit`. Expected: compile/test FAIL because config types and validation do not exist.

- [ ] **Step 3: Implement exact defaults and validation**

- default `CALLS_ACTIVE_PROVIDER` to `livekit`;
- reject any value other than `livekit` in Phase 150, including `jitsi` with the stable premature-cutover error;
- when `JITSI_ENABLED=false`, allow empty Jitsi values;
- when enabled, require all five string fields;
- outside `DevMode`, require HTTPS public URL, minimum 32-character secret, and reject known example/dev secret values;
- keep existing LiveKit production validation because it remains active/rollback;
- add the variables to backend environment in local/prod Compose without adding Jitsi services to the default Hub stack.

Provider-name and enabled-field validation runs before the existing `DevMode` early return, so development cannot silently accept an unknown or premature active provider. Production-only URL/secret-strength checks remain after that boundary. Implement `NewConferenceProvider(callsCfg, liveKitCfg) (ConferenceProvider, error)`: `livekit` returns the LiveKit adapter, `jitsi` returns the locked premature-cutover error, and every other value is rejected. Wiring calls this factory and exits through the existing fail-closed startup pattern on error; it no longer constructs LiveKit directly.

- [ ] **Step 4: Run backend regression to green**

Run `backend-unit`. Expected: PASS.

- [ ] **Step 5: Commit exact paths**

```powershell
git add backend/internal/platform/config/config.go backend/internal/calls/provider_factory.go backend/cmd/server/wiring_integrations.go backend/tests/platform/config/calls_provider_config_test.go backend/tests/calls/provider_contract_test.go deployments/.env.example docker-compose.local.yml deployments/docker-compose.prod.yml
git commit -m "feat(calls): validate self-hosted jitsi shadow config"
```

### Task 3: Add immutable opt-in Jitsi shadow deployment

**Files:**

- Create: `deployments/docker-compose.jitsi.yml`
- Create: `deployments/jitsi/.env.example`
- Modify: `deployments/Caddyfile.prod`
- Modify: `deployments/docker-compose.prod.yml`

**Interfaces:**

Compose is invoked only as an overlay with the existing production file:

```powershell
docker compose -f deployments/docker-compose.prod.yml -f deployments/docker-compose.jitsi.yml --env-file deployments/.env up -d
```

Core image references are exact:

```yaml
jitsi-web: ghcr.io/jitsi/web@sha256:4231d5ff7318ff71d12ffd41afc36adae3c319b40cf79cde39b009b0c68295dc
jitsi-prosody: ghcr.io/jitsi/prosody@sha256:9e341b10134d947e49ca338b9efe8d70ac9d6e8d89d621da2246dfa79fe50363
jitsi-jicofo: ghcr.io/jitsi/jicofo@sha256:a8d64e46eaa340751a2b5530a4abd4175adf05bff07e9775fc48daae6364cfe9
jitsi-jvb: ghcr.io/jitsi/jvb@sha256:d7dc646d21af07530ae6acf1ace23ea4782445d0362004621a32a5a632a964e9
```

- [ ] **Step 1: Write the static deployment acceptance checklist before YAML**

At the top of `deployments/jitsi/.env.example`, document required variables with non-secret examples:

```text
JITSI_HOST=meet.yourdomain.com
JITSI_PUBLIC_URL=https://meet.yourdomain.com
JITSI_XMPP_DOMAIN=meet.jitsi
JITSI_XMPP_AUTH_DOMAIN=auth.meet.jitsi
JITSI_XMPP_MUC_DOMAIN=muc.meet.jitsi
JITSI_XMPP_INTERNAL_MUC_DOMAIN=internal-muc.meet.jitsi
JITSI_APP_ID=hub
JITSI_APP_SECRET=CHANGE_ME_MINIMUM_32_CHARS_RANDOM_STRING
JITSI_ISSUER=hub
JITSI_AUDIENCE=jitsi
JICOFO_AUTH_PASSWORD=CHANGE_ME_RANDOM_SECRET
JVB_AUTH_PASSWORD=CHANGE_ME_RANDOM_SECRET
JVB_ADVERTISE_IPS=203.0.113.10
```

The checklist states: exactly four distinct component-to-OCI-index mappings, `ENABLE_AUTH=1`, `AUTH_TYPE=jwt`, `ENABLE_GUESTS=0`, `JWT_ALLOW_EMPTY=0`, `ENABLE_JAAS_COMPONENTS=0`, no Hub DB credentials, Prosody/Jicofo/JVB administrative endpoints private, only Caddy HTTPS plus JVB UDP media exposed, and LiveKit untouched.

- [ ] **Step 2: Implement the overlay topology**

Define `jitsi-web`, `jitsi-prosody`, `jitsi-jicofo`, and `jitsi-jvb`. Use named config volumes per component; `jitsi-internal` with `internal: true` for XMPP/control; `jitsi-edge` with `internal: true` shared only by Caddy and web; and a separate `jitsi-media` network attached only to JVB for explicitly firewalled media traffic. Attach existing `caddy` to `jitsi-edge` and inject `JITSI_HOST=${JITSI_HOST:?required}` into the Caddy service through the overlay. Expose only JVB UDP `10000:10000/udp`; do not add legacy TCP `4443`, removed upstream, and do not publish Prosody/Jicofo/JVB REST ports. Set `PUBLIC_URL=${JITSI_PUBLIC_URL}`, `DISABLE_HTTPS=1`, `ENABLE_AUTH=1`, `AUTH_TYPE=jwt`, `ENABLE_GUESTS=0`, `ENABLE_JAAS_COMPONENTS=0`, `JWT_APP_ID=${JITSI_APP_ID}`, `JWT_APP_SECRET=${JITSI_APP_SECRET}`, `JWT_ACCEPTED_ISSUERS=${JITSI_ISSUER}`, `JWT_ACCEPTED_AUDIENCES=${JITSI_AUDIENCE}`, `JWT_ALLOW_EMPTY=0`, `JWT_AUTH_TYPE=token`, and `JWT_TOKEN_AUTH_MODULE=token_verification`; pass the four frozen XMPP domains/users/passwords following the release-tagged upstream Compose mapping. Set `RTCSTATS_ENABLED=0`, `ENABLE_JAAS_COMPONENTS=0`, `JVB_DISABLE_STUN=true`, and no TURN/STUN URL; any TURN addition requires a separately approved self-hosted Hub component. Add bounded CPU/RAM/pids limits, read-only roots/tmpfs matching upstream, and restart policies. Healthchecks are exact: web uses `curl --fail http://127.0.0.1:8000/`; Prosody, Jicofo, and JVB use the release-provided `/usr/local/bin/healthcheck.sh`; dependents use `condition: service_healthy`.

- [ ] **Step 3: Add the dedicated Caddy origin**

Keep the existing Hub host and `/lk/*` block byte-for-byte except surrounding context required by the patch. Add a second site:

```caddyfile
{$JITSI_HOST} {
    header Strict-Transport-Security "max-age=31536000; includeSubDomains"
    reverse_proxy jitsi-web:8000
}
```

Do not add a path-based fallback and do not reference `meet.jit.si`.

- [ ] **Step 4: Record verification honestly**

Perform static inspection for the four distinct digest mappings, networks, ports, auth flags, Caddy environment injection, upstream env mapping, healthchecks, and absence of SaaS endpoints. Do not run Compose parsing on the Windows host. Do not pull/start images without explicit infrastructure authorization. Record the operator continuation commands `docker compose -f deployments/docker-compose.prod.yml -f deployments/docker-compose.jitsi.yml --env-file deployments/.env config --quiet` and the overlay `up -d` command above; render/start/TLS/JWT/UDP/TURN/restart stay `NOT_RUN` until an explicitly authorized task-owned Linux deployment produces artifacts.

- [ ] **Step 5: Commit exact paths**

```powershell
git add deployments/docker-compose.jitsi.yml deployments/jitsi/.env.example deployments/Caddyfile.prod deployments/docker-compose.prod.yml
git commit -m "feat(deploy): add pinned self-hosted jitsi shadow"
```

### Task 4: Document operation, rollback, and Phase boundaries

**Files:**

- Create: `docs/runbooks/jitsi-shadow-deployment.md`
- Create: `deployments/jitsi/THIRD-PARTY-NOTICES.md`
- Create: `.planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-VERIFICATION.md`
- Modify: `docs/wiki/DEPLOY.md`
- Modify: `docs/TECHNICAL.md`
- Modify: `.planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-RESEARCH.md`

**Interfaces:**

- Runbook uses the exact overlay, env names, image receipts, dedicated origin, and `CALLS_ACTIVE_PROVIDER=livekit` contract from Tasks 1-3.
- Runtime evidence table has columns `check`, `command`, `revision`, `status`, `artifact`; initial unavailable checks stay `NOT_RUN`.

- [ ] **Step 1: Write the runbook acceptance table**

Include: preflight secrets and DNS; registry/digest verification; host-firewall default-deny egress owned by the deployment operator with explicit Caddy ACME/DNS and JVB media exceptions; firewall/TLS/UDP requirements; start order; core health; local JWT rejection/acceptance; browser media smoke; restart recovery; logs/metrics; config-volume backup; upgrade with new digest receipt; rollback by stopping only the overlay; and explicit proof that LiveKit active calls remain unaffected.

Create `THIRD-PARTY-NOTICES.md` with the upstream repositories, exact release, Apache-2.0 notice links, four immutable OCI index references, retrieval date, and a clear statement that per-image SBOM generation/review is `NOT_RUN` until an authorized infrastructure run. Give exact digest-addressed SBOM continuation commands; release acceptance remains blocked until those rows are complete.

- [ ] **Step 2: Synchronize product/deployment docs**

Update current architecture wording from “Calls are LiveKit” to “LiveKit active shadow; self-hosted Jitsi foundation present; cutover not yet accepted.” Keep recording documentation LiveKit-labelled. State explicitly that Phase 150 creates no Jitsi-to-Hub callback/mutation endpoint; callback authentication and replay/idempotency remain a blocking Phase 152 contract. Link Phase 152 for UI/credential activation and Phase 154 for Jibri/cutover/removal.

- [ ] **Step 3: Run the applicable code regression gate**

Run `backend-unit` once more because docs describe the new config/provider code. Expected: PASS. Jitsi runtime remains `NOT_RUN` unless the task-owned environment was authorized and executed.

- [ ] **Step 4: Independent static review**

Create `150-VERIFICATION.md`, then give the reviewer `150-SPEC.md`, `150-RESEARCH.md`, exact BASE..HEAD diff, backend Docker output, and the static deployment checklist. Critical failures include any floating image, external host/API, shared DB, enabled `jitsi` active provider, missing auth, exposed internal component, or removed LiveKit rollback asset.

- [ ] **Step 5: Commit exact paths**

```powershell
git add docs/runbooks/jitsi-shadow-deployment.md deployments/jitsi/THIRD-PARTY-NOTICES.md docs/wiki/DEPLOY.md docs/TECHNICAL.md .planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-RESEARCH.md .planning/phases/150-self-hosted-jitsi-deployment-and-provider-foundation/150-VERIFICATION.md
git commit -m "docs(calls): add jitsi shadow operations contract"
```

## Final Phase Gate

1. `backend-unit` must PASS with provider/config units and existing Calls regressions.
2. Static scan must show exactly four distinct component-to-OCI-index mappings, no `meet.jit.si`/JaaS/external analytics or STUN endpoint, and no public/internal DB sharing.
3. `CALLS_ACTIVE_PROVIDER` remains `livekit`; public Calls API/UI/recording/webhook and `/lk/*` remain compatible.
4. Update `150-VERIFICATION.md` with registry metadata marked verified, backend Docker evidence marked PASS/FAIL, and Jitsi render/start/media/rollback/SBOM as `NOT_RUN` unless actually executed. Phase 150 cannot be called production-ready while those runtime rows are not run.
