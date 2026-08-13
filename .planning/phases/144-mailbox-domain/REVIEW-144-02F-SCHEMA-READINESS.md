# Phase 144-02F preimplementation review — receipt/deployment schema readiness

> **Detected:** 2026-08-13 root deep static audit
> **Status:** CONTRACT_AMENDMENT_REQUIRED · IMPLEMENTATION_BLOCKED · RUNTIME_NOT_RUN
> **Scope:** `MailMutationReceipt`, receipt target/member rows and `MailRuleDeployment`

## Outcome

The narrative correctly separates client replay, internal recovery and asynchronous rule
deployment, but the schema packet cannot preserve those guarantees across a process crash.
Most importantly, rule deployment has no durable record of the generation/hash that was
actually sent to the engine.

## P1 — rule deployment loses the in-flight generation across crash

The worker must capture desired generation/hash, perform engine I/O outside the database
transaction, reconcile an unknown result, and finalize only the captured generation. The
listed row stores only current desired and last deployed pairs. If desired generation 6 is
written while generation 5 is applying, or the worker dies after installing 5, an expired
lease/reconciler cannot prove which pair reached the engine.

The row needs an explicit immutable-per-attempt pair:

- nullable `attempted_generation`, `attempted_hash_version` and exact 32-byte
  `attempted_hash`, all-present or all-null;
- `applying|unknown` require that attempted tuple; a new desired generation may advance
  independently without rewriting it;
- the engine script marker includes hash version + attempted generation + hash;
- successful reconciliation copies exactly the attempted tuple to deployed, then becomes
  `ready` only if it also equals current desired; otherwise it returns to `pending`;
- definitive rejection records only an allowlisted code and never labels current desired
  deployed; indeterminate state never uploads again before marker reconciliation.

Desired, attempted and deployed hashes also require a canonical v1 definition: active rules
ordered by `(priority,id)`, exact typed fields and preserved condition/action array order,
length-delimited encoding, SHA-256, and no comparison across hash versions.

## P1 — receipt lifecycle and compaction are not a database state machine

The contract lists many nullable fields but does not state which are legal for
`in_progress|unknown|completed|failed`, nor how a compacted client tombstone is distinguished
from live replay metadata independently of cleanup-job timing.

The amendment should freeze named CHECKs for:

- key and request digests are immutable exact 32-byte values and fingerprint version is
  positive; raw UUID keys are structurally absent;
- `in_progress|unknown` have no terminal/replay timestamps or public status and are never
  compacted; unknown carries only a reconciliation-safe code/schedule;
- the first `completed|failed` transition atomically stores terminal time, immutable original
  HTTP status and `replay_until = terminal_at + 30 days`; neither retry nor cleanup extends it;
- client rows at or after `replay_until` behave as expired immediately even if physical child
  cleanup is delayed, and persist a tombstone/optional `compacted_at` until account or
  organization deletion;
- internal folder rows may hard-delete only after the same terminal instant and are never
  addressable by an HTTP key;
- target kind/id, safe argument kind/id, status and safe error code follow an exact
  operation/state matrix rather than arbitrary nullable combinations;
- lease token/until and retry fields are paired and bounded.

## P1 — operation, target and safe-code grammars are not frozen

`operation`, target/argument kinds and persisted safe codes are currently prose. Two schema
implementations can expose different replay behavior while both claim conformance. Publish
one closed enum covering exactly folder create plus internal update/delete, label/signature/
rule/attachment-presign/domain/alias create and message bulk. For each operation, freeze:

- `client|internal` key source;
- whether account is required or forbidden;
- target/argument kind and terminal target requirement;
- success/failure HTTP status set and route-specific safe error-code allowlist;
- whether engine reconciliation is possible and which state remains non-terminal.

Strings must have explicit byte caps, original status must remain in the HTTP 100..599 range,
attempt counters must be non-negative/bounded, and no response JSON, body, address, filename,
engine id, provider error or raw Sieve field may exist.

## P1 — bulk child limits and outcome checks need exact enforcement

The 200-member limit is aggregate per receipt, not per target, and therefore cannot be
implemented by an Ent slice validator alone. Reservation must lock the parent, insert a
complete owner-scoped expansion and reject before any engine I/O if the distinct total would
exceed 200.

Minimum keys/checks:

- target UUID surrogate id; unique `(receipt_id, request_index)` and
  `(receipt_id, target_type, target_id)`;
- member UUID surrogate id; unique `(receipt_id, message_id)` plus required target parent;
- request indexes start at zero, are contiguous for the normalized request and remain below
  200; message requests store one matching member per target;
- affected/succeeded/failed counts are 0..200 and obey an exact status/count relation;
  training status and nullable code obey their closed combinations;
- target/member rows cascade only from receipt compaction/deletion and never from later
  deletion of the projected message/thread, because replay snapshots must survive;
- compaction removes all children transactionally while retaining the client parent hashes.

## P1 — deployment state/nullability checks remain incomplete

In addition to the attempted tuple, the one-row-per-account deployment needs:

- positive desired generation/hash-version/hash; nullable deployed tuple all-or-none and no
  deployed generation greater than desired;
- `ready` requires deployed = desired, no attempted tuple, no lease, no error/retry;
- `applying` requires attempted tuple plus a paired expiring lease and no terminal error;
- `unknown` preserves attempted tuple and one exact unknown code until marker reconciliation;
- `pending|error` never claim current desired deployed; error codes come only from the
  OpenAPI rule deployment allowlist;
- non-negative bounded attempts, nullable next retry, immutable account uniqueness and exact
  account-delete behavior;
- no engine script id/body or raw compiler output in PostgreSQL.

## Gate impact

- 144-02F is `BLOCKED_CONTRACT`; no receipt/deployment Ent packet should be authored;
- 144-02E generation, generic mutation services, rule CRUD/deployment and every dependent
  handler remain blocked on the exact matrices above and the earlier 02A/02D amendments;
- generation/migration and runtime remain independently blocked by founder/launcher gates;
- no production/test source, dependency, generator, validator, migration or Docker command
  was changed or run.
