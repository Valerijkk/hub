# Phase 150 research: Jitsi release, immutable pins, and Hub seams

**Date:** 2026-08-24
**Release selected:** `stable-11146-2`
**Release date:** 2026-08-17
**Runtime/tests:** NOT_RUN; registry manifest metadata inspected without pulling image layers

## Official sources

- Jitsi Meet repository: <https://github.com/jitsi/jitsi-meet>
- Official Docker distribution: <https://github.com/jitsi/docker-jitsi-meet>
- Selected release: <https://github.com/jitsi/docker-jitsi-meet/releases/tag/stable-11146-2>
- Docker deployment guide: <https://jitsi.github.io/handbook/docs/devops-guide/devops-guide-docker/>
- Architecture: <https://jitsi.github.io/handbook/docs/architecture/>
- Jibri requirements: <https://jitsi.github.io/handbook/docs/devops-guide/devops-guide-requirements/>

The upstream release Compose uses a shared release tag and does not publish a release-wide immutable digest lockfile. Hub therefore resolves and records each component digest independently.

## Verified multi-architecture image pins

Resolved on 2026-08-24 with `docker buildx imagetools inspect` (registry metadata only):

| Component | Immutable OCI index reference |
|---|---|
| Web | `ghcr.io/jitsi/web@sha256:4231d5ff7318ff71d12ffd41afc36adae3c319b40cf79cde39b009b0c68295dc` |
| Prosody | `ghcr.io/jitsi/prosody@sha256:9e341b10134d947e49ca338b9efe8d70ac9d6e8d89d621da2246dfa79fe50363` |
| Jicofo | `ghcr.io/jitsi/jicofo@sha256:a8d64e46eaa340751a2b5530a4abd4175adf05bff07e9775fc48daae6364cfe9` |
| JVB | `ghcr.io/jitsi/jvb@sha256:d7dc646d21af07530ae6acf1ace23ea4782445d0362004621a32a5a632a964e9` |

Do not replace these with floating `stable`, `latest`, or tag-only image references. Re-resolution is an explicit upgrade change and must inspect the new official release first.

## Deterministic refresh procedure

For every enabled component, inspect the official release tag without pulling layers:

```powershell
docker buildx imagetools inspect ghcr.io/jitsi/web:stable-11146-2
docker buildx imagetools inspect ghcr.io/jitsi/prosody:stable-11146-2
docker buildx imagetools inspect ghcr.io/jitsi/jicofo:stable-11146-2
docker buildx imagetools inspect ghcr.io/jitsi/jvb:stable-11146-2
```

Record the top-level OCI index digest, release, retrieval date, and official source. Deployment uses `name@sha256:...`, while the release tag stays only as provenance metadata.

## Current Hub seams

- `backend/internal/calls/service.go` constructs a LiveKit-specific room service.
- `backend/internal/calls/service_room.go` owns LiveKit room/token behavior.
- `backend/cmd/server/wiring_integrations.go` creates LiveKit integration directly.
- `backend/internal/platform/config/config.go` exposes LiveKit-only calls configuration.
- Public Calls DTOs and `frontend/src/components/calls/CallRoom.tsx` remain LiveKit-shaped and are intentionally unchanged until Phase 152.
- `backend/internal/calls/recording.go`, LiveKit webhook handling, and `/lk/*` routing remain intact until Phase 154.

## Frozen topology decision

- Use a dedicated Jitsi origin derived as `https://meet.<Hub deployment domain>`; do not host Jitsi under an arbitrary Hub subpath.
- Core services are Web, Prosody, Jicofo, and JVB on dedicated Hub-owned networks/volumes.
- Hub does not share its Postgres or Mongo database with Jitsi.
- LiveKit and `/lk/*` remain available as the active/rollback provider in Phase 150.
- Jibri is deferred to Phase 154. One Jibri handles one simultaneous recording; it must run on separate capacity with fast storage. Official guidance requires at least 8 GB RAM for one 1280x720 recording and 12 GB for 1280x1024; the Docker overlay reserves 2 GB shared memory.

## Evidence boundary

The digest lookups are verified registry metadata. Compose rendering, container startup, JWT auth, TLS, UDP media, external IP/TURN, health recovery, and rollback remain NOT_RUN until the task-owned Docker environment is created.

## Official upstream discovery amendment - 2026-08-30

- The selected official release is `stable-11146-2`, commit `738058b`: <https://github.com/jitsi/docker-jitsi-meet/releases/tag/stable-11146-2>. The tagged Compose, not `master`, is the source template: <https://raw.githubusercontent.com/jitsi/docker-jitsi-meet/stable-11146-2/docker-compose.yml>.
- Official JWT configuration requires `ENABLE_AUTH=1`, `AUTH_TYPE=jwt`, `JWT_APP_ID`, and `JWT_APP_SECRET`. Hub's secret-store name `JITSI_APP_SECRET` must map explicitly to upstream `JWT_APP_SECRET`. Hub security policy requires an explicit distinct `JICOFO_COMPONENT_SECRET` because upstream documents the usable default `s3cr37`; `JICOFO_AUTH_PASSWORD` and `JVB_AUTH_PASSWORD` remain separately required for service startup: <https://jitsi.github.io/handbook/docs/devops-guide/devops-guide-docker/>.
- From `stable-11146`, official containers are rootless/read-only as UID/GID 1000 and separate config, durable `storage`, and regenerable `tmp` paths. Exact ownership and mounts remain a per-digest discovery gate.
- JVB normally publishes UDP 10000 and may use a public STUN default. Hub must set its own `JVB_ADVERTISE_IPS` and `JVB_DISABLE_STUN=true`; external TURN, analytics, JaaS, and provider hosts remain forbidden without a separate self-hosted decision.
- The tagged template proves that custom Prosody modules can be mounted through `/prosody-plugins-custom` and inserted on the MUC through `XMPP_MUC_MODULES`. Exact lifecycle hook names/order, retry behavior, and absence or presence of an equivalent stock durable egress remain `UNVERIFIED_FOR_SELECTED_DIGEST` until JITSI-D1; no stock facility is accepted as authority without that proof.
- The accepted candidate provider event source is a Hub-owned custom Prosody module plus an independently selected durable source spool. Before any provider-side Jitsi module/spool/sidecar code atom, JITSI-02B0 must freeze a persistent driver/store, retention, capacity, ownership, backup and crash behavior; `memory|none|null`, Hub application databases, MUC room state and unverified container filesystems are rejected. The provider-neutral JITSI-02A envelope/HMAC atom may precede B0 because it attaches to no Jitsi image or route. The module then persists a provider-deployment delivery ID and exact envelope bytes before send and retries those bytes. HMAC-SHA256 is supported by Prosody; documented `net.http` does not prove a client-certificate sender contract, so mTLS requires a separately verified private transport adapter/sidecar: <https://prosody.im/doc/developers/muc>, <https://prosody.im/doc/developers/moduleapi>, <https://prosody.im/doc/developers/util/hashes>, <https://prosody.im/doc/developers/net/http>.
- Public network remains Caddy HTTPS 443 plus JVB UDP 10000. Port 80/ACME is a separate certificate-provisioning decision. Prosody/XMPP, Jicofo health/control, JVB private REST, the event spool, and transport adapter remain private. Upstream JVB documents candidate health endpoint `/about/health`, but selected-digest compatibility is `UNVERIFIED_FOR_SELECTED_DIGEST` and belongs to JITSI-D1: <https://github.com/jitsi/jitsi-videobridge/blob/master/doc/rest.md>.

These findings are `STATIC_DISCOVERY_ONLY`. No image layer, Compose render/start, plugin, spool, transport, JWT, TLS, media, retry, or rollback behavior was executed.
