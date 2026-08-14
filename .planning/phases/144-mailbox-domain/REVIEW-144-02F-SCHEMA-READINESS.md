# Phase 144-02F preimplementation review — receipt/deployment schema readiness

> **Canonical transition evidence (2026-08-14):** corrective commit `fcb5b84fed4a417e736ddb4053256e494a0c25f6` (parent `55e36b7a1a3516d07b2baf390a5742a24ec8fbe5`, tree `5a0637f2306ff42295adea5560793e2a8f4a294a`, subject `docs(mail): correct receipt and rule invariants`) changed the exact same 28 docs; worktree was clean at review. It is `COMMITTED_UNACCEPTED_AT_FCB5B84F` / `EXACT_SHA_TRANSITION_REQUIRED` / `IMPLEMENTATION_BLOCKED`. Contract/Spec/Oracle Terra/high PASS C/I/M `0/0/0`; Security/Adversarial Terra/xhigh PASS `0/0/0`; Lifecycle/Dependency/Boundary Terra/high FAIL `0/1/0` solely because committed precommit status/instructions became stale after commit, while Git object, boundary, blockers and exclusions PASS. All dispatcher-enforced, fallback none, telemetry `UNKNOWN`. Prior precommit/root-commit-next sequence is `CONSUMED/HISTORICAL`. Next: static review of this transition candidate → one exact 28-doc docs-only transition commit (SHA intentionally not predeclared) → fresh Lifecycle exact-SHA re-review of `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition; trace-only Contract/Security rechecks as required → only after all exact reviews PASS, separate 02F docs-only `STATIC_ACCEPTED` sync → separate Rule plan acceptance and Phase-145 carrier/G145-5B/G145-5C. No source/schema/generation/migration/service/runtime authority; launcher `NOT_READY`, executables `NOT_RUN`.

> **Current prerequisite:** 144-02D exact docs boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after three exact PASS reviews C/I/M `0/0/0`. 02F itself remains unaccepted. The former sole-next closure instruction is `CONSUMED/HISTORICAL`; the live sequence is transition-review candidate → one exact 28-doc transition commit without self-SHA → fresh Lifecycle exact-SHA re-review of `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition with trace Contract/Security rechecks as needed → all exact reviews PASS → separate 02F docs-only acceptance sync. No implementation/runtime authority.
> **Detected:** 2026-08-13 root deep static audit
> **Historical status:** prior `ROOT_CORRECTION_CANDIDATE` / precommit/root-commit-next sequence is `CONSUMED/HISTORICAL`; 02F is not accepted.

> **02F normative mirror:** receipt exhaustion uses NULL-safe `safe_code IS DISTINCT FROM 'MAIL_RECONCILIATION_EXHAUSTED'`; folder `p=10` due unknown is no-I/O failed 409 with 30-day replay, while `message_bulk p=10` is manual-only safe unknown with both leases and schedule null. Same-key/same-request resolves first as persisted 202; a fresh key is account-lock guarded by that exact live manual-only row and returns `423 MAIL_ACCOUNT_BUSY` plus bounded `Retry-After`, with no row/snapshot/expansion/I/O. Target uniques are `(receipt_id,request_index)` and `(receipt_id,target_type,target_id)`; member aggregates are transaction-owned. Deployment uses desired/attempted/deployed tuples, bounded 0..10 budget, exact lease/CAS fencing and a full fail-closed state CASE; only outcome-unknown is `unknown`. Canonical v1 hash header includes `0x00||0x01`; frames are stored-order and length-exact, and fixed vectors are owned by the persistence contract. This review neither authorizes source/schema/generation/migration/service/runtime nor alters launcher `NOT_READY` / executable `NOT_RUN`.
> **Scope:** `MailMutationReceipt`, receipt target/member rows and `MailRuleDeployment`

> `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` now supplies the root candidate state/operation/
> hash matrices; this discovery verdict remains open until independent exact-SHA review.

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
- no production/test source, dependency, generator, validator, migration or Docker command was changed or run.

## 144-03B Rule dependency clarification

Rule contract is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`.
This 02F review is not closed or accepted by Rule-contract acceptance. Rule TDD/production requires explicit 02F closure, candidate and
later-plan exact-SHA acceptance, and the separate Phase-145 engine-port carrier.

## Correction-candidate adjudication — 2026-08-14

> **Historical status:** prior `ROOT_CORRECTION_CANDIDATE` / precommit/root-commit-next sequence is `CONSUMED/HISTORICAL`; 02F is not accepted.

| Prior P1 | Candidate adjudication | Remaining gate |
|---|---|---|
| in-flight deployment generation crash | `CLOSED_IN_CANDIDATE`: immutable attempted tuple is persisted before I/O and only confirmed reconciliation copies it to deployed | independent exact-SHA contract/spec, security/adversarial and lifecycle/evidence reviews |
| receipt lifecycle/compaction state machine | `CLOSED_IN_CANDIDATE`: named digest/pair/state/replay/compaction predicates define `prior_attempts p`, live-tenth contention and post-expiry no-I/O exhaustion | same three independent exact-SHA reviews |
| operation, target and safe-code grammar | `CLOSED_IN_CANDIDATE`: closed union, exhaustive named operation-target-status-safe-code CASE, separate key-source/account predicate and folder matrices are published | same three independent exact-SHA reviews |
| bulk limits/outcomes | `CLOSED_IN_CANDIDATE`: row predicates are separated from locked-parent aggregate/cardinality invariants | same three independent exact-SHA reviews |
| deployment state/nullability | `CLOSED_IN_CANDIDATE`: tuple, lease, public/internal error separation, `0..10` budget, exhaustion terminal and ready/non-ready rules are explicit | same three independent exact-SHA reviews |

External/static oracle matrix: inspect every named receipt predicate (including attempts,
lease, unknown, exhaustion terminal, replay window and safe-code/operation/state), target
index/count/outcome/training alternatives, member non-row-CHECK disclaimer, deployment tuple/
state predicates, and the three named OpenAPI folder/idempotency 409 components. No test or
runtime oracle has run; launcher stays `NOT_READY` and all runtime evidence stays `NOT_RUN`.

Historical lifecycle: initial eight-document candidate
`63e419897aa0683e1d3dfcd442aca78eea1b0727` is `CONSUMED/HISTORICAL`. Its 02F receipt/deployment
normative facts remain frozen, but it is not a closure or acceptance; its 02D RecentContact
portion is superseded. Lifecycle transition `6c424ed2d6703ceccbb269c37d9abc03e75a0410` is
`CONSUMED/HISTORICAL` and its former next-review instruction is superseded.

**CONSUMED/HISTORICAL pre-acceptance record:** corrected 02D commit `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` was
`COMMITTED_UNACCEPTED` / `EXACT_SHA_CLOSURE_IN_PROGRESS`. Its Contract and Security exact-SHA
reviews PASS C/I/M `0/0/0`; Lifecycle exact-SHA FAIL C/I/M `0/1/0` solely on stale lifecycle
wording, with content/object boundary PASS. This transition record consumes the stale instruction;
all review provenance is dispatcher-enforced, fallback none, telemetry `UNKNOWN`.

**Consumed/historical next step:** fresh Lifecycle exact-SHA re-review of `7d` plus this transition
record, with fresh Contract/Security trace-only rechecks permitted. After every exact review PASS,
a separate docs-only 02D `STATIC_ACCEPTED` sync is required. Only then may an independent 02F
exact-SHA closure review occur. Its routes remain Contract/Spec (`gpt-5.6-terra`/high),
Security/Adversarial (`gpt-5.6-terra`/xhigh), and Lifecycle/Dependency
(`gpt-5.6-terra`/high), each dispatcher-enforced with no fallback. 02F remains unclosed;
no source, schema, generation, migration, service or runtime authority is granted. Launcher
stays `NOT_READY`; executable evidence stays `NOT_RUN`.
> **Current Rule acceptance:** the three Terra routes PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Rule contract acceptance does not close this review or authorize schemas.
