# Phase 144-02A preimplementation review — schema contract readiness

> **Detected:** 2026-08-13 root deep static audit
> **Status:** CONTRACT_AMENDMENT_REQUIRED · IMPLEMENTATION_BLOCKED · RUNTIME_NOT_RUN
> **Scope:** `MailAccount`, `MailIdentity`, `MailDomain`, `MailboxPolicy` only

## Outcome

The packet cannot yet satisfy its own requirement to implement the exact section 3
fields, enums and bounds. The existing documents describe the lifecycle correctly, but
several persistence shapes and cardinality limits remain non-machine-readable or
representation-ambiguous. No schema implementation should guess them.

## P1 — operational persistence matrix is incomplete

The contract does not freeze enough detail for a deterministic Ent schema:

- `MailAccount` calls engine ids/cursors/checkpoints bounded but does not bind their stored
  columns to the accepted `MaxEngineOpaqueIDBytes=1024` and
  `MaxEngineCursorBytes=4096` limits;
- `provision_error_code` and `sync_error_code` require closed safe allowlists, but no exact
  persisted enums exist;
- account and domain deployed generation/hash values are described together without an
  explicit both-null-or-both-present constraint;
- `MailDomain` names no exact columns/nullability for its stable provisioning key, engine
  id, desired/deployed hashes, retry schedule/attempts, error, lease pair or deletion
  finalization;
- `EngineDNSRecord` contains `TTL`, while the public `MailDomain.required_dns_records`
  shape omits it; the persistence projection therefore has no exact canonical shape;
- `MailboxPolicy` still lacks hard maxima for message/attachment bytes and does not state
  whether `max_attachment_bytes <= max_message_bytes` is mandatory.

The minimum amendment must publish one field table with exact names, types, nullability,
byte/item limits, defaults, immutable fields, enums and relational checks. At minimum the
database must express all lease token/until pairs and all deployed generation/hash pairs as
both present or both absent, require exact SHA-256 length, and forbid a `ready` projection
without a committed freshness instant.

## P1 — account/domain/identity cardinality is not bounded

The API permits repeated domain and alias creation, but neither the spec nor OpenAPI defines
an organization domain cap or account identity/alias cap. Both admin list responses return an
unpaginated array without `maxItems`, and `AccountSpec.Aliases` therefore also has no
authoritative bound. An administrator or compromised admin session can grow storage,
response memory and one account-control engine request without a contract ceiling.

A decision must freeze either:

1. hard creation caps plus bounded single-page responses; or
2. hard creation caps plus cursor pagination, with the account-control adapter still
   receiving a separately bounded exact alias set.

The existing engine constants provide a coherent default basis: at most 100 domains per
organization, at most 200 total identities per account including the primary identity,
and list pagination default 50 / maximum 100. These values are recommendations only until
the contract is amended.

## Representation ambiguity — normalized identity uniqueness

`MailIdentity.address` is specified as unique within an organization, but the listed entity
has no `organization_id` column. The schema must choose between a denormalized organization
foreign key with consistency enforcement or a globally unique normalized address. Because
mail domains are already globally unique and every identity must belong to its account's
organization domain, global normalized-address uniqueness is the safer non-denormalized
default, but the contract must state that strengthening explicitly.

## Deterministic details already available

- normalized domain wire length: 253 characters;
- normalized address/display name: 320 / 160 characters;
- opaque engine id and cursor/checkpoint limits: 1024 / 4096 bytes;
- public DNS record count/name/value limits: 32 / 253 / 2048;
- generation/version values are positive; attempts/epochs are non-negative;
- no password, token, administrator secret, private DKIM material or provider text belongs
  in any of the four schemas.

## Gate impact

- 144-02A is `BLOCKED_CONTRACT`; no Ent schema or inverse edge should be authored yet;
- 144-02B/C/D/F, generation 144-02E and services/workers depending on 02A remain blocked;
- the separate founder migration decision and accepted Docker launcher remain required even
  after this contract gap is closed;
- runtime, tests, validators, generators, migrations and Docker commands remain `NOT_RUN`.

No production/test source was changed. The review used the routed Mail contract, OpenAPI,
Phase 144 packet documents and the already accepted engine-port/schema conventions only.
