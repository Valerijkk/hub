# Stalwart 0.16.17 research snapshot — 2026-08-13

> **Status:** `RESEARCH_SNAPSHOT_ONLY / NOT_GATE_PASS`
>
> This file records current official-source findings for Phase 145. It does not
> approve a license, container image, credential design, deployment, network
> egress or runtime command. No image was pulled and no runtime check was run.

## Candidate identity

| Item | Snapshot value | Acceptance state |
|---|---|---|
| Latest official release observed | `v0.16.17`, released 2026-08-10 | CANDIDATE_ONLY |
| Source commit | `08509ddccbe6f978a1ffebeb250cab8588d765e5` | SOURCE_REFERENCE_ONLY |
| Container image tag | `docker.io/stalwartlabs/stalwart:v0.16.17` | REGISTRY_METADATA_INSPECTED / NOT_PULLED / NOT_APPROVED |
| Multi-platform OCI index digest | `sha256:a8108e19bd927e172d4d8c128907b8dfc93fd180ae8ee07dccdd42cb97eb9dfa` | CAPTURED_NOT_APPROVED |
| `linux/amd64` image manifest | `sha256:2bcbd3797c8170bd261a533d54d3b25c1a49558aa69f3e42b0d37bdf2b106a28` | CAPTURED_NOT_APPROVED |
| `linux/amd64` attestation manifest | `sha256:30141f15afd19a4dab739a31be018b3874df7b8cc01ced79125e7b34c9cf5c24` | CONTAINS_SLSA_PROVENANCE_REFERENCE / NOT_VERIFIED |
| SLSA provenance layer | `sha256:f5c80070c2184e8caa6caa5f32e90fa42d7e15f6e2bde66a77775d6d9125dba6` | PREDICATE_REFERENCE_ONLY |
| Binary asset checksums/attestations | release page publishes them, but they are not a container manifest digest | NOT_CONTAINER_PROVENANCE |
| License | upstream is dual-licensed AGPL-3.0 / SELv2 | FOUNDER_LEGAL_DECISION_PENDING |
| Docker/runtime evidence | none | NOT_RUN |

The release is new enough to include the fixes published after `v0.16.9`, including the
JMAP `*/changes` shared-object id disclosure fix, and after the scoped API-key privilege
escalation fix listed in `v0.16.15`. This is only a release-floor observation. It is not a
security acceptance or a substitute for digest, attestation, SBOM and configuration review.

Read-only registry inspection was performed with `docker buildx imagetools inspect`; it did
not pull or run the image. The available Docker daemon reports `linux/x86_64`, which maps to
the captured `linux/amd64` manifest. Registry metadata ties that image to source revision
`08509ddccbe6f978a1ffebeb250cab8588d765e5`, runs as user `stalwart`, declares
`/etc/stalwart` and `/var/lib/stalwart`, and contains the expected built-in healthcheck.
The index exposes one attestation manifest for that platform whose only layer is labelled
as SLSA provenance. Signature verification and SBOM discovery/validation were not run, so
G145-4 remains open.

## Official evidence used

| Official source | Frozen finding for Hub |
|---|---|
| <https://github.com/stalwartlabs/stalwart/releases> | `v0.16.17` is the latest observed release; the page identifies commit `08509dd` and binary asset attestations/checksums. |
| <https://github.com/stalwartlabs/stalwart/commit/08509ddccbe6f978a1ffebeb250cab8588d765e5> | Full source commit for the candidate release. |
| <https://github.com/stalwartlabs/stalwart> | Upstream license choice is AGPL-3.0 or SELv2; Hub's legal choice remains external to this repository. |
| <https://hub.docker.com/r/stalwartlabs/stalwart/tags> plus read-only OCI registry metadata | Docker Hub candidate index/platform/attestation digests above; image labels point to the exact source revision. Search-page caches were not used for the digest. |
| <https://stalw.art/docs/install/platform/docker/> | Official images are published with full and moving tags. Volumes are `/etc/stalwart` and `/var/lib/stalwart`; bind mounts require UID 2000 ownership. Hub still pins the exact immutable digest rather than trusting a tag. |
| <https://stalw.art/docs/http/> | Main-server health endpoints are `/healthz/ready` and `/healthz/live`. |
| <https://stalw.art/docs/http/jmap/> | Mail JMAP is discovered at `/.well-known/jmap`; the mail endpoint is `/jmap`. |
| <https://www.stalw.art/docs/auth/authentication/api-key/> | API keys authenticate the management JMAP API only. They cannot authenticate JMAP mail, IMAP, POP3, SMTP submission, CalDAV, CardDAV or WebDAV. |
| <https://stalw.art/docs/auth/authorization/administrator/> | `impersonate` grants broad access and upstream recommends granting it sparingly. It is not silently accepted as Hub's mailbox credential design. |
| <https://stalw.art/docs/ref/object/domain/> | Domain provisioning uses `x:Domain/*`, `urn:stalwart:jmap` and the management `/api` endpoint. |
| <https://stalw.art/docs/ref/object/account/> | Account provisioning uses `x:Account/*`; a user has a local-part `name`, a `domainId`, credentials, roles, permissions and aliases. |
| <https://stalw.art/docs/sieve/jmap/> and <https://www.rfc-editor.org/rfc/rfc9661.html> | A mailbox's personal scripts use RFC 9661 `SieveScript/*` on mail JMAP with capability `urn:ietf:params:jmap:sieve`; script bytes are uploaded as a bounded blob before `SieveScript/set`. |
| <https://stalw.art/docs/ref/object/sieve-system-script/> | `x:SieveSystemScript` is a server system script under the management capability; it is not the per-mailbox Hub rule target. |
| <https://stalw.art/docs/ref/object/sieve-user-script/> | `x:SieveUserScript` is a global script available for user imports; it is not a user's active personal script. |
| <https://stalw.art/docs/ref/object/spam-training-sample/> | `x:SpamTrainingSample` exists on `/api`, but current docs do not freeze a client-supplied message/blob-to-spam-or-ham mapping or native idempotency/reconciliation for Hub feedback commands. |

## Required endpoint and credential separation

The earlier Phase 145 shorthand of one “JMAP/admin URL” and one administrator secret is
invalid for this candidate. The accepted adapter must model two non-interchangeable planes:

| Plane | Fixed endpoint | Capability / objects | Credential class | Hub use |
|---|---|---|---|---|
| Management | trusted origin + `/api` | `urn:stalwart:jmap`; `x:Domain`, `x:Account` and other reviewed management objects | separately scoped management API key | domain/account/identity provisioning and only explicitly reviewed administrative objects |
| Mailbox | `/.well-known/jmap` discovery, then same-origin `/jmap`/session URLs | core, mail, submission, blob and optional RFC 9661 sieve | account password/app password/OAuth, or a separately reviewed least-privilege delegation design | messages, mailboxes, drafts, submission, changes and personal Sieve |
| Health | trusted origin + `/healthz/ready` and `/healthz/live` | main server health | deployment-specific; no credential assumption frozen here | bounded readiness/liveness observation only |

Hard consequences:

1. A management API key must never be sent to `/jmap`, an upload/download URL or a mail
   protocol port.
2. A mailbox credential must never be reused as the management API key.
3. Management `/api` is a fixed reviewed path, not a URL learned from a user or mailbox
   JMAP session.
4. Mail session, upload and download URLs are accepted only under the same-origin policy
   rooted at the trusted configured origin.
5. Enabled mailbox mode remains blocked until the least-privilege mailbox delegation and
   secret lifecycle are accepted. Hub does not auto-generate or persist one plaintext
   password per user as a guessed fallback.

## Packet corrections

- **145-01A:** the disabled configuration shape can remain safe, but enabled parsing must
  require distinct management and mailbox credential references. The old
  `MAILBOX_ENGINE_ADMIN_USER` / `MAILBOX_ENGINE_ADMIN_SECRET` pair is not an accepted
  enabled contract.
- **145-01B:** transport code must keep auth providers separated by plane. Redaction and
  same-origin policy are shared invariants; an authorization header is not.
- **145-02:** mail JMAP discovery verifies core/mail/submission/blob and optional sieve
  capabilities. Selected-version management compatibility is checked separately against
  fixed `/api` object/method evidence; it is not inferred from the mail session.
- **145-03A/03C:** account/domain operations use the management plane and exact
  `x:Account/*` / `x:Domain/*` schemas.
- **145-07B:** fixed per-mailbox rules use RFC 9661 `SieveScript/get`, bounded blob upload,
  `SieveScript/set` and activation on the mailbox plane. `x:SieveSystemScript` and
  `x:SieveUserScript` are forbidden for this behavior.
- **145-07C:** classifier feedback remains blocked. `x:SpamTrainingSample` alone does not
  establish the required explicit message mapping or safe lost-response retry semantics.

## Still-open acceptance gates

1. Founder/legal license selection.
2. Independent verification/approval of the captured `v0.16.17` index and `linux/amd64`
   manifest digests, provenance predicate/signature, SBOM/source relationship immediately
   before any authorized pull. The current registry read is discovery, not approval.
3. Exact selected-version configuration export/checksum and listener/access-control rules.
4. Least-privilege management API-key permissions and separate mailbox authentication or
   delegation design, rotation, revocation and audit.
5. Proof that no user-facing IMAP/JMAP/ManageSieve or alternate direct mailbox writer can
   mutate Hub-owned custom folders/scripts in the accepted profile.
6. A documented classifier-feedback mapping with idempotency/reconciliation, or explicit
   omission of feedback from the adapter.
7. Accepted task-owned Docker launcher, bounded runtime evidence, backup/restore and exact
   cleanup.

Until all applicable gates pass, the truthful state is `candidate researched`,
`implementation blocked`, and `runtime NOT_RUN`.
