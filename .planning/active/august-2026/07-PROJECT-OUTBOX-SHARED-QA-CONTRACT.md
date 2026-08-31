# Project Outbox — shared disposable QA substrate contract

> **Status:** `P0_ACCEPTED / P1A_CLEARED_FOR_ATOMIC_ADMISSION / SOURCE_CONTRACT_ONLY`  
> **Decision:** Path B — shared disposable `qa-*` substrate with exact server-side fixture ownership  
> **Scope:** `TC-OUTBOX-0001`…`TC-OUTBOX-0005` and their future Python protocol/test implementation. No Compose, service, database, stream, or product-runtime ownership is created here.

## Decision and rationale

The project-outbox integration cases use the existing, fixed launcher-compatible disposable `qa-*` substrate. They do not provision an exclusive server, PostgreSQL database, NATS/JetStream stream, MongoDB, Redis, or infrastructure service.

Path B is selected because it preserves fixed-launcher compatibility, reduces disk and container churn, keeps isolation at exact server-issued IDs, avoids Python database/NATS dependencies, and gives one maintainable fixture/cleanup boundary.

## Ownership boundary

Server-side fixture setup creates and records immutable fixture organization UUID and owner UUID. Each case additionally records only its exact project-side-effect IDs in one server manifest and owns one durable consumer created for that case.

The manifest is authoritative for cleanup. It may contain the fixture organization UUID, fixture owner UUID, created project UUIDs, their enumerated project-side-effect IDs, and the durable-consumer identity. Immutable raw UUID authority survives cleanup and tombstoning; tombstones retain the original exact target list and completion evidence. It must never be synthesized from names, emails, request nonces, prefixes, broad searches, received receipts, or cleanup outcomes.

The shared backend, database instances and contents outside exact manifest IDs, JetStream stream, broker, caches, and Compose services are shared substrate. A case neither owns nor deletes them.

## Observation and negative-or-zero evidence

The observation maximum wait is **3 seconds**, within the existing **5-second HTTP read budget**. Expiry of that observation window is evidence only as specified by the case; it does not expand fixture discovery or cleanup authority.

Authoritative server projection queries may use only the immutable fixture organization/owner UUIDs and recorded project UUIDs. Receipt assertions must compare exact outbox payload owner UUID and project UUID equality. They must not query, scan, or select by name, email, nonce, or prefix.

For negative/zero cases, zero is established only by these authoritative server projections over the fixture UUIDs and recorded project UUIDs. A missing receipt, an observer result, a timeout, a cleanup result, or the absence of a broad scan is never proof of zero state.

## Cleanup protocol

Cleanup quiesces only the exact case consumer and deletes only exact IDs enumerated in the server manifest: the durable consumer, fixture records within the stated fixture boundary, and recorded project-side-effect IDs. Cleanup must be idempotent for a lost-response or partial-cleanup attempt, but it must not discover targets dynamically or broaden ownership after an uncertain result.

It must not remove the shared backend, any shared database or database-wide contents, the JetStream stream, broker, cache, or infrastructure services. No cleanup receipt or empty result may be used to infer that all owned state was absent before cleanup.

## Packet admission and acceptance order

### P0 — protocol and test acceptance (mandatory before P1–P7)

Before any P1–P7 implementation or test packet is accepted, its frozen protocol/test design must prove all of the following:

1. Server fixture setup returns and records immutable organization/owner UUIDs and a single manifest of exact IDs.
2. Every project create records its project UUID and enumerated side-effect IDs before they can be used by a projection or cleanup.
3. Exactly one durable consumer is created per case and is named/addressed only through its manifest identity.
4. All observation stops at 3 seconds while remaining within the existing 5-second HTTP read budget.
5. Positive assertions use exact project/owner UUID equality in the outbox payload; zero assertions use only the constrained server projections described above.
6. Cleanup takes targets only from the manifest and has explicit negative tests demonstrating that shared server, database, stream, and services are not cleanup targets.

**P0 admission:** `ACCEPTED / CLEARED_FOR_P1A`. The independent P0 review accepted the remediated frozen protocol. Canonical Docker evidence is `python-contract` **280/280 PASS** (seed `4057944629`, `5.39s`). This clears only P1a's stated atomic admission; it is not evidence that P1a has run.

P0 acceptance remains static/protocol admission plus the stated framework receipt only. It does not claim a running launcher, product behavior, migration/build/lint success, runtime readiness, merge, or deployment authority.

### P1–P7 — downstream work constraint

P1a is cleared for its stated atomic admission, while P1a–P7 execution remains `NOT_RUN`. Every downstream packet must preserve the P0 fixture/manifest, observation, projection, and cleanup invariants; a packet that requires exclusive server/database/stream ownership, Python DB/NATS access, a name/email/nonce/prefix scan, or cleanup-derived zero evidence is **BLOCKED** and requires a new root decision. P1a alone may add untagged authority schemas/tables to the existing `mail-calendar-v1` Atlas lineage; `qa && !release` protects authority access/writes/routes/wiring rather than schema build tags, and release contains only inert generated types/tables.

P1a freezes immutable manifest `id`, `scenario`, SHA-256 `nonce_hash`, SHA-256 `control_capability_hash`, `creation_generation`, roots `organization_id`/`owner_user_id`, `session_id`, consumer stream/durable/filter, and SHA-256 consumer configuration fingerprint. Mutable manifest state is closed: `state`; all-or-none cached receipt `project_id`/`created_by_id`/`cached_at`; per-root cleanup timestamps; `cleanup_error_class`; `tombstoned_at`; and timestamps. Effects freeze immutable `id`, `manifest_id`, `ordinal`, roots, project/member/outbox UUIDs, and `registered_at`; only per-target cleanup completion timestamps are mutable. Hashes/fingerprint are exactly 32 bytes, every UUID authority is non-zero, state/tombstone and receipt-triple consistency are checked, and raw UUID authority persists through tombstones. Uniqueness is manifest `nonce_hash`, `organization_id`, `owner_user_id`, `session_id` and effect `(manifest_id, ordinal)`, `(manifest_id, project_id)`, `(manifest_id, project_member_id)`, `(manifest_id, outbox_event_id)`. The only relation is required immutable `Manifest -> Effects` on `Effect.manifest_id` with `OnDelete(NoAction)`; no FK/edge targets `User`, `Organization`, `Session`, `Project`, `ProjectMember`, or `OutboxEvent`. No raw nonce, payload, name, key, email, or discovery data is stored; durable name is an exact handle, never lookup authority outside its manifest plus exact stream/name.

P1b adds a separate exact-ID-returning publisher API (a `PublishInTxWithID`-like path) for the project recorder; existing `PublishInTx` semantics and Calendar/Tracker/Mail callers are unchanged. All manifest/effect writers and queries in P2/P4/P5 are compiled only under `qa && !release`; ordinary project code receives only a nil/no-op observer interface and cannot write QA authority in release. Session composition is split: P5a adds an internal-auth capability accepting a caller-supplied non-zero session UUID while preserving existing single-session, Redis, database, NATS, and token semantics and using the existing auth parser for verification; P5b owns the QA route/wiring and performs no manual JWT signing or signing-secret access.

## Case mapping

| Cases | Shared-substrate interpretation |
|---|---|
| `TC-OUTBOX-0001`, `TC-OUTBOX-0002` | Fixture owner/organization, recorded project UUID, exact payload equality, one manifest-owned consumer. |
| `TC-OUTBOX-0003`, `TC-OUTBOX-0004` | No created project is presumed; zero is an authoritative constrained projection, never a missing receipt or cleanup inference. |
| `TC-OUTBOX-0005` | The pre-publication/reconnect observer is the sole manifest-owned durable consumer; replay remains bounded by the 3-second observation maximum. |

## Non-goals and status

This contract does not modify Python source/protocol, Compose, tests, launcher, `STATE`, or product documentation. The recorded P0 receipt is limited to `python-contract` **280/280 PASS** (seed `4057944629`, `5.39s`); P1a–P7 are `NOT_RUN`. P1a may now enter only its frozen atomic admission/review lifecycle, not product, runtime, merge, or deployment acceptance.
