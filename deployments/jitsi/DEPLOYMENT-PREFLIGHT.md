# Hub self-hosted Jitsi deployment preflight

## Current boundary

- Phase 150 status: **PARTIAL — STATIC_FOUNDATION_ONLY**.
- LiveKit is active; Jitsi remains unavailable until Phase 152.
- The Jitsi overlay is **NOT_CREATED / UNBOUND**.
- Task 8 is **BLOCKED_BY_AUTHORIZED_LINUX_DISCOVERY**.
- Runtime and media evidence are **NOT_RUN**.

This packet is a static deployment/configuration boundary only. It does not
authorize image pulls, Compose rendering, service startup, or any use of an
external API or hosted meeting provider.

## Pinned candidate release metadata

Candidate release: `stable-11146-2`.

The following references are registry metadata only. They do **not** verify
the image contents, licence, SBOM, runtime behaviour, or suitability for use.

| Component | Immutable reference |
| --- | --- |
| web | `ghcr.io/jitsi/web@sha256:4231d5ff7318ff71d12ffd41afc36adae3c319b40cf79cde39b009b0c68295dc` |
| prosody | `ghcr.io/jitsi/prosody@sha256:9e341b10134d947e49ca338b9efe8d70ac9d6e8d89d621da2246dfa79fe50363` |
| jicofo | `ghcr.io/jitsi/jicofo@sha256:a8d64e46eaa340751a2b5530a4abd4175adf05bff07e9775fc48daae6364cfe9` |
| jvb | `ghcr.io/jitsi/jvb@sha256:d7dc646d21af07530ae6acf1ace23ea4782445d0362004621a32a5a632a964e9` |

No claim is made here that Apache-2.0 has been verified, that any licence has
been accepted, or that an SBOM is complete. `THIRD-PARTY-NOTICES` belongs to
later Task 10.

## Evidence matrix

| Boundary | Evidence status |
| --- | --- |
| Overlay creation and binding | UNBOUND |
| Container users / UID / GID | NOT_RUN |
| Writable paths and configuration ownership | NOT_RUN |
| Rootless execution | NOT_RUN |
| Branding | NOT_RUN |
| Effective `external_api.js` | NOT_RUN |
| SBOM | NOT_RUN |
| Compose render, startup, and health | NOT_RUN |
| Caddy / TLS | NOT_RUN |
| Hub-to-Jitsi JWT integration | UNBOUND |
| JVB advertise address, UDP, and media | NOT_RUN |
| Restart and rollback | NOT_RUN |

## Authorized continuation gate

The overlay may proceed only with explicit infrastructure authority and a
numeric disk budget. Discovery must pull no more than one digest-pinned image
at a time, record disk space before and after each pull, and never run parallel
pulls or builds. Until that authorization exists, this preflight remains static
and all runtime/media checks remain `NOT_RUN`.
