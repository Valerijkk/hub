# Phase 144-02C preimplementation review — object/draft schema readiness
> **Current narrow deletion-source authority (2026-08-23; highest precedence):** accepted plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` is `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` only for one future atomic no-test exact2 production packet: `backend/internal/mail/mailbox_deletion.go` and `backend/internal/mail/mailbox_deletion_fake.go`. Plan Contract/high, Security/xhigh and Lifecycle/high reviews are C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own commit and immutable C/S/L triple `0/0/0`. Adapter capability proof is not required only for this provider-neutral ABI plus two standalone fakes exception, but remains mandatory for every concrete provider adapter. Persistence/Ent/schema/generation/migration/SQL/worker/storage drain/account finalization/User-GDPR/HTTP/NATS/Docker/runtime stay `BLOCKED` / `NOT_RUN`. Old blanket deletion-source-blocked or plan-pending claims are `CONSUMED/HISTORICAL` only within this exact2 scope; unrelated gates are unchanged.
> **Engine deletion-port adjudication (2026-08-23):** `STATIC_ACCEPTED / DOCS_DESIGN_ONLY`; former deletion-port-next selection is `CONSUMED/HISTORICAL`, never source authority. Freeze separate opaque future ports: `EngineAccountDeletionPort` (`DeleteAccount`, `ReconcileAccountAbsence`) and `MailObjectDeletionPort` (`DeleteObject`, `ReconcileObjectAbsence`); do not widen normal `MailEngine` or combine account delete with drain. Requests durably retain nonzero `DeletionID` plus account provisioning key+engine account ID or object bucket+key; opaque IDs use `MaxEngineOpaqueIDBytes` and bucket/key existing storage bounds. Results are `deleted|authoritatively_absent`, observations `present|authoritatively_absent`; bare `NotFound` preserves intent unless scoped authoritative absence is proven. No provider/raw IDs or causes are public/logged. Private intent/control is not owner projection; short lease transactions keep I/O outside locks, use exact token+generation+scope CAS, reclaim expired lease only, and stale token/zero rows cause no I/O/compensation. Object drain starts after `upload_drain_until` and attachment/reference terminality; account removal is last after exact child/object/receipt/control/activity manifest; hard-delete only after closed evidence. No Stalwart route/OAuth/DNS/OpenAPI assumption. Adapter capability and independent contract/security/lifecycle reviews are mandatory; future source is bounded ABI/fake then separately planned persistence/worker. Ent/source/generated/migration/SQL/tests/Docker/runtime and physical User/GDPR remain `BLOCKED` / `NOT_RUN`.

> **144-02C source receipt (2026-08-20):** `STATIC_ACCEPTED / EXACT_SHA_PASS` applies only to immutable `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) and the exact 13 blobs in `IMPLEMENTATION-144-02C.md`. Tests/oracles first; stated review routes PASS C/I/M `0/0/0`. The lower 02C authority-candidate, source-unaccepted, lifecycle-next and implied-next-packet entries are `CONSUMED/HISTORICAL`. No next Mail code packet is selected or authorized; root must select it separately. Ent generation, migration, direct SQL, services, API/UI, runtime, product, phase and release remain outside this receipt; 02F-D/02E, deletion-port, physical User/GDPR intent, launcher `NOT_READY` and executable evidence `NOT_RUN` remain current. Provenance Terra/medium, dispatcher-enforced/no fallback, telemetry `UNKNOWN`.

> **2026-08-15 lifecycle correction acceptance:** exact14 is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Candidate/precommit/root-commit/exact-review/status-sync-next clauses are `CONSUMED/HISTORICAL`. Engine drain/final Account deletion remains separately `BLOCKED` / `NOT_DISPATCHABLE` pending the independent docs-only engine deletion-port contract/adjudication. No source/Ent/service/test/runtime authority.

> **144-02C historical authority-bridge candidate:** `CONSUMED/HISTORICAL`; it is superseded by the immutable source receipt `e565523f589fbd7536f6f8e4df7effd39469c1af` and its 13-blob manifest. It confers no current or next-packet authority.

## Immutable accepted test-first exact13 source manifest — receipt only

1. `backend/tests/mail/mail_delivery_schema_test.go`
2. `backend/ent/schema/mailobject.go`
3. `backend/ent/schema/mailattachment.go`
4. `backend/ent/schema/maildraft.go`
5. `backend/ent/schema/mailsubmission.go`
6. `backend/ent/schema/mailaccount.go`
7. `backend/ent/schema/organization.go`
8. `backend/ent/schema/user.go`
9. `backend/ent/schema/mailmessage.go`
10. `backend/tests/mail/mail_account_identity_schema_test.go`
11. `backend/tests/mail/mail_projection_schema_test.go`
12. `backend/tests/mail/mail_domain_policy_schema_test.go`
13. `backend/tests/calendar/calendar_mutation_receipt_schema_test.go`

The test oracle is authored first. The final two compatibility tests are mandatory because they mechanically assert the precise appended Organization and User inverses and delete metadata. The source packet must preserve the accepted `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` 02C fields/defaults/immutability, named checks and index names; Organization/User/Account -> Object `NO ACTION`; Attachment -> Object/Message/Draft `RESTRICT`; Draft account `CASCADE`, reply/forward Message `SET NULL`; Submission account `CASCADE`, terminal-only Draft `SET NULL`. It also freezes paired actual-size/SHA, delete-lease fields, immutable 11-minute drain boundary and partial due index; Attachment backing XOR/owner exclusion/scan-hash rules/unique nullable storage object/no FileMetadata or project edge; Draft stable key/client scope, nullable engine id/version pair, status/mutation matrix and scalar one-draft attachment ownership (the 50-item cap is repository/service-only); Submission 32-byte hashes, opaque result pair versus scalar Hub snapshot, exact status/error/nullability and no FK from `result_mail_message_id`. This is a future review oracle, not authorization.

> **Detected:** 2026-08-13 root deep static audit
> **Lifecycle:** exact14 docs/readiness representation is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc`; implementation remains separately blocked.
> **Status:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS` docs representation · former implementation/source/schema route CONSUMED/HISTORICAL; runtime `NOT_RUN` · launcher `NOT_READY` · executables `NOT_RUN`
> **Scope:** `MailObject`, `MailAttachment`, `MailDraft` and `MailSubmission`

> `docs/mail/ENGINE-CONTRACT-0.1.8.md`, OpenAPI 0.1.8 and the 02C persistence matrix supply
> the accepted exact14 docs representation at `576f19231e37f635d5b7187a1ada121ed14ff8dc`.
> The lower re-review-pending/unaccepted candidate narrative and all P1/outcome/gate-impact
> clauses are `CONSUMED/HISTORICAL` as readiness adjudication; they preserve discovery provenance
> only and grant no implementation, source, schema or runtime authority.

> **2026-08-13 source-only follow-up:** a requested Terra/high review whose actual runtime
> provenance was unavailable found two remaining P1s: one field conflated the opaque engine
> result message id with the public Hub UUID, and the durable draft association lacked the
> 50-attachment cap. The root candidate now separates both identities, requires projection
> finalization before `sent`, and enforces the cap under a locked draft. This lower re-review-pending
> statement is `CONSUMED/HISTORICAL`, not a current acceptance or implementation route.

> A second source-only pass found that generic unknown reconciliation could still requeue a
> confirmed result and that hydration lacked exact account/message/origin proof. The root
> candidate now splits no-pair reconciliation from retained-pair projection-only recovery,
> validates exact returned identity/version and derives confirmed submission origin in the
> domain. This lower follow-up is `CONSUMED/HISTORICAL`, not a current unaccepted route.

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

The accepted scalar relation is additionally capped at 50 selected attachment rows per
draft. Replacement validates the entire distinct id set and resulting count while the draft
is locked; a 51st association is rejected before any engine I/O and cannot partially alter
the prior set.

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
- `MailboxPolicy` creation defaults are independently accepted at `25,000,000` decimal bytes
  for both message and attachment, distinct from the hard caps of `52,428,800` message bytes
  and `26,214,400` attachment bytes; administrators configure only within those bounds.

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
- persist opaque `engine_submission_id` and `engine_result_message_id` separately from the
  scalar Hub UUID snapshot `result_mail_message_id`;
- enforce submission states: `queued|submitting` have no result ids/error; `sent` requires
  both opaque ids plus the Hub UUID and no error; `unknown` has no Hub UUID and either no
  opaque result pair or a complete pair awaiting projection finalization; `failed` has no
  result ids and one terminal allowlist code;
- make stable engine submission key and both nullable result identities unique per account;
  the Hub result UUID is a receipt snapshot, not a deletion-sensitive foreign key;
- require a bounded same-account result read and atomic `MailMessage` id/version projection
  upsert before `sent`; returned account/message must exactly equal the requested/persisted
  result and confirmed submission origin is domain-derived;
- split unknown recovery: no opaque pair calls `ReconcileSubmission` and may requeue only on
  definitive absence; a complete pair retries only message hydration/finalization and can
  never requeue or call Submit;
- freeze `OnDelete` behavior: account hard-delete cascades owned rows; draft/message result
  links use restrict or set-null only where the retention contract explicitly permits it.

## Gate impact

- 144-02C contract amendment is independently accepted, but former schema-authoring block is CONSUMED/HISTORICAL; the accepted source receipt selects no next packet and
  generation/migration gates remain unaccepted/blocked and launcher remains `NOT_READY`;
- 144-04D/04E and attachment/storage packets remain blocked on the stable draft key, staging
  cardinality, state matrix and hard byte ceilings;
- 144-01A also requires the narrow draft-create reconciliation amendment in addition to its
  previously recorded response-byte boundary;
- generation/migration and runtime remain independently blocked by founder/launcher gates;
- no production/test source, dependency, generator, validator, migration or Docker command
  was changed or run.
