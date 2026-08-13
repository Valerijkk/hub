# Phase 144-02C preimplementation review — object/draft schema readiness

> **Detected:** 2026-08-13 root deep static audit
> **Status:** 0.1.8_ROOT_AMENDMENT_AVAILABLE · INDEPENDENT_CLOSURE_PENDING · IMPLEMENTATION_BLOCKED · RUNTIME_NOT_RUN
> **Scope:** `MailObject`, `MailAttachment`, `MailDraft` and `MailSubmission`

> `docs/mail/ENGINE-CONTRACT-0.1.8.md`, OpenAPI 0.1.8 and the 02C persistence matrix now
> supply the root candidate; this verdict remains open until independent exact-SHA review.

## Outcome

The four-entity packet cannot yet be implemented deterministically. The documents freeze the
intended recovery outcomes, but they do not freeze the durable identity needed to recover an
initial draft create, the attachment staging/consumption relationship, or the database state
matrix for uploaded objects and submissions.

## P1 — initial draft-create reconciliation has no durable engine identity

The frozen draft flow reserves a Hub row in `saving`, invokes `SaveDraft(create)`, and on a
lost response calls `ReconcileDraft`. The accepted engine value requires `MessageID`,
`PreviousVersion` and `MutationID`, while the first successful engine response is the source
of the opaque message id and version. A row whose first response was lost therefore cannot
populate two of the inputs that the current reconciliation value requires. Persisting only
`engine_message_id`, `engine_version` and `last_mutation_id` does not close this gap.

The contract must add one adapter-independent stable draft key before any schema/service
packet proceeds. Recommended shape:

- derive an immutable bounded `engine_draft_key` from the Hub draft UUID, never from the raw
  HTTP idempotency key;
- carry that key in both `EngineDraft` and `EngineDraftReconciliation` and require adapters to
  find-or-prove-absent before another create;
- persist it on `MailDraft`, unique per account;
- keep `engine_message_id` and `engine_version` as a nullable pair only while the initial
  create is `saving|unknown`; require both after an applied create;
- require `last_mutation_id` for `saving|unknown`, and never treat an empty engine id as a
  provider-specific magic lookup rule.

Without that key, the advertised lost-create replay can either duplicate a draft or remain
permanently indeterminate.

## P1 — attachment staging and ownership cardinality are contradictory

Presign requires `draft_client_id`; `MailObject` and `MailAttachment` persist it, but draft
create/update accepts only attachment ids and `MailDraft` has no corresponding scope field.
The ERD simultaneously declares a many-to-many draft/attachment edge. The contract therefore
does not say whether `draft_client_id` is authorization, cleanup-only upload grouping, an
immutable draft identity, or whether one uploaded attachment may be reused across drafts.

That choice changes the Ent edges, unique indexes and delete behavior. The safe working
default is an immutable owner/account-bound staging scope that can be consumed by one live
draft and its eventual sent message; autosave keeps the same association, while another
draft must clone/restage it. A deleted draft releases only still-unconfirmed/unreferenced
objects to expiry cleanup. If cross-draft reuse is intended instead, deletion and reference
counting must be specified before the many-to-many edge is generated.

Every ownership path must remain explicit:

- a storage-backed attachment obtains account ownership through its unique `MailObject`;
- an engine-blob attachment must have an owning `MailMessage` and is account-scoped through
  that message;
- draft/message/attachment links are same-account service invariants before insert;
- no `FileMetadata`, project or project-membership edge is permitted.

## P1 — upload and scan state have two authorities and no conditional CHECK matrix

`MailObject` and `MailAttachment` both list `scan_status`, although the ERD makes the backing
object optional and the public resource reads attachment status. No rule identifies the
canonical field or prevents them from drifting. The object also lists expected/actual size,
hash, expiry and `pending|confirmed|orphaned|deleted` without exact nullability by state.

Recommended normalized representation:

- `MailObject` owns upload lifecycle only: immutable bucket/key/expected metadata/expiry,
  nullable actual size plus SHA-256 as a pair, and `upload_state`;
- `MailAttachment` is the sole scan-state authority for both storage and engine backing;
- `pending` upload has no actual-size/hash pair; `confirmed` requires the pair and exact
  actual = expected size; terminal cleanup states retain any already-known pair for audit;
- `MailAttachment` has an always-enforced backing XOR, a unique nullable
  `storage_object_id`, and an engine blob requires a message owner;
- SHA-256 fields are exactly 32-byte values, sizes are non-negative and policy-capped, and
  object `(bucket, object_key)` is immutable and unique;
- the final maximum size remains gated by founder question 31.

Cross-row account equality and storage-state-to-scan transitions must be enforced in the
repository transaction because a row-local PostgreSQL CHECK cannot traverse foreign keys.

## P1 — draft/submission fields and terminal states are not machine-frozen

The packet still lacks exact byte/storage types for every hash and opaque id, plus the named
database checks that mirror the already-frozen public submission state machine.

Minimum amendment:

- store idempotency/request hashes as immutable 32-byte values with unique
  `(mail_account_id, idempotency_key_hash)` indexes; raw keys and body material remain
  structurally absent;
- bound every engine id/key/version by the accepted 1,024-byte opaque-id ceiling using an
  application validator and a database octet-length check, not character-count alone;
- make draft revision positive; make attempt counts non-negative and bounded; use UUID for
  Hub mutation ids and exact allowlisted enums/codes;
- enforce the draft engine id/version nullable pair and the stable-key rules above;
- enforce submission states: `queued|submitting` have no result ids/error, `sent` requires
  both engine submission id and result message id with no error, `unknown` has no result ids
  and one reconciliation allowlist code, and `failed` has no result ids and one terminal
  allowlist code;
- make stable engine submission key unique per account, and make a confirmed result message
  resolve from at most one submission in that account;
- freeze `OnDelete` behavior: account hard-delete cascades owned rows; draft/message result
  links use restrict or set-null only where the retention contract explicitly permits it.

## Gate impact

- 144-02C is `BLOCKED_CONTRACT`; no four-file Ent packet should be authored;
- 144-04D/04E and attachment/storage packets remain blocked on the stable draft key, staging
  cardinality, state matrix and hard byte ceilings;
- 144-01A also requires the narrow draft-create reconciliation amendment in addition to its
  previously recorded response-byte boundary;
- generation/migration and runtime remain independently blocked by founder/launcher gates;
- no production/test source, dependency, generator, validator, migration or Docker command
  was changed or run.
