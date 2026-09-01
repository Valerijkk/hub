# Contract 144-06 — Inbound fixture cleanup and restart reconciliation

**Status:** `FROZEN_ARCHITECTURE / DOCS_ONLY`. Runtime remains `RED` / `NOT_RUN`; this contract authorizes neither implementation, migration, deployment, nor product acceptance.

## Frozen decision D

Cleanup is canonical exact-ID cleanup of the published inbound fixture. Restart recovery is map-less reconciliation from a known nonce. `newInboundRequestFixturePlan(nonce)` deterministically reconstructs the exact fixture **roots**: `Organization`, `User`, `MailDomain`, `MailAccount`, `MailIdentity`, `MailFolder`, `MailThread`, `MailMessage`, and envelope event ID, plus their immutable root values and relations. The canonical plan contains the deterministic Ent `mailIdentityID`; the engine fixture supplies the already-proven nonempty engine primary-identity ID. Those plan facts are the only ownership proof for roots.

The plan does **not** contain deterministic IDs for every Calendar-derived child. A derived `CalendarInboundMessageReceipt`, `CalendarInboundObjectState`, `CalendarEvent`, `CalendarAttendee`, `CalendarReplyDelivery`, or `CalendarMutationReceipt` may be enumerated only through unique, exact, constrained relations starting at the known `mailMessageID`, `envelopeEventID`, or the one proven event relation. Before cleanup, each derived child and each relation used to reach it must be recorded and validated against the exact root proof. A restart with no in-memory map repeats only this same constrained traversal; it may not search by name, nonce prefix, organization/user ownership inference, or any broad query.

Do not add a durable descriptor, ownership marker, schema change, namespace/prefix deletion, inferred ownership rule, per-invocation Compose teardown, database/volume reset, or collision retry convention. The retained QA substrate is intentional.

`Fake.NextID` is boot-scoped collision tolerance only. Its boot-specific generated IDs are not canonical cleanup identities, do not alter the fixture manifest, and currently remain runtime `NOT_RUN` evidence.

## Closed graph and required proof

The closed allow-list is exactly the following graph. Deletion or rehydration is authorized only after every present record has the stated exact proof and every expected edge is unique.

| Entity | Required exact proof |
|---|---|
| `Organization` | deterministic plan root and immutable root values |
| `User` | deterministic plan root; exact `Organization` relation |
| `MailDomain` | deterministic plan root; exact `Organization` relation |
| `MailAccount` | deterministic plan root; exact `User` and `MailDomain` relations |
| `MailIdentity` | deterministic plan `mailIdentityID`; exact `MailAccount` relation; normalized owner address; primary, active, and `send_allowed=true`; engine identity exactly equals the proven nonempty engine primary-identity ID; fixed timestamps |
| `MailFolder` | deterministic plan root; exact `MailAccount` relation |
| `MailThread` | deterministic plan root; exact `MailAccount` relation |
| `MailMessage` | deterministic plan root; exact `MailAccount`, `MailThread`, and `envelopeEventID` values |
| `MailMessageFolder` | deterministic relation ID or exact constrained `MailMessage`–`MailFolder` relation |
| `CalendarInboundMessageReceipt` | exact constrained relation from the proven `MailMessage`/envelope path |
| `CalendarInboundObjectState` | exact constrained relation from the proven inbound-message and object path |
| `CalendarEvent` | the single exact event relation reached from the proven envelope/message path |
| `CalendarAttendee` | exact constrained relation from the proven `CalendarEvent` |
| `CalendarReplyDelivery` | exact constrained relation from proven event/attendee/message records |
| `CalendarMutationReceipt` | exact constrained relation from proven event/inbound-message records |

Before the first delete, the implementation records the complete proven dependent set and validates every selected node's immutable fields and endpoints. It proves exactly one `MailIdentity` for the exact account before deleting that exact identity before the account. It also checks absence guards for every other dependent relation: a missing expected relation, an extra record, an extra endpoint, a foreign field value, a non-unique traversal result, or any unrecognized dependent is unsafe. It returns `409`; it never widens a query to make the graph fit.

## Lifecycle and response semantics

| State | Required result |
|---|---|
| complete proven `created` or `published` graph | exact cleanup is allowed |
| known nonce and in-memory map missing | reconstruct roots, including the deterministic `MailIdentity`, from the plan; re-derive Calendar children only by the exact constrained relations above; rehydrate only after full proof |
| all exact roots absent after completed cleanup | idempotent empty `204` |
| `publishing`, `publish_unknown`, `cleanup_failed`, `locked` | fail closed with `409` |
| partial graph, foreign/mismatched proof, non-unique child traversal, or unexpected dependent | fail closed with `409` |

A `409` must not delete, overwrite, recreate, reset Compose, expand its selection, or fall back to process-local state. A complete pre-crash graph is recoverable from the nonce only after the same exact root and constrained-child proof; map absence alone is not a failure.

## Crash boundaries and packets

1. Residual state before a complete proven graph is `partial` or `publish_unknown` and returns `409`.
2. A complete created/published graph that predates map update is reconciled by the exact root plan, including its deterministic `MailIdentity`, plus constrained Calendar-child traversal, never by a global scan or account-identity scan.
3. A cleanup failure or lock is `cleanup_failed`/`locked` and returns `409`; later calls do not guess which records to remove.
4. After a confirmed complete cleanup, all exact roots absent returns `204`.

Implementation remains split into independently reviewed packets: exact root proof and constrained child enumeration; lifecycle cleanup/reconciliation; Python `TC-0002` `finally` cleanup preserving primary failure; retained-substrate runtime evidence. This contract is documentation only; every packet and its tests are still `RED`/`NOT_RUN` until separately authored and accepted.

## Docker evidence gate

Only the exact Docker service is an acceptance gate:

```powershell
docker compose -f docker-compose.local.yml run --rm python-integration
```

It retains `qa-postgres`, publishes the fixture, recreates the backend process, invokes known-nonce reconciliation, and must eventually prove exact root reconstruction, constrained child re-derivation, no duplicate canonical graph, boot-scoped Fake-ID collision tolerance, idempotent root-absent `204`, and non-mutating `409` for all fail-closed states. Host execution, Compose teardown, and disk-churning resets are forbidden. `backend-unit` is not a substitute. Runtime evidence is currently `RED`/`NOT_RUN`.

## Review ledger

Self-review classification: `C/I/M = 0/0/0`.

- **C:** 0 — exact root determinism is kept distinct from non-plan Calendar child enumeration.
- **I:** 0 — all 15 listed graph nodes, including exact `MailIdentity` lifecycle proof and deletion order, absence guards, restart semantics, `204`/`409`, Fake-ID role, and Docker-only gate are bounded.
- **M:** 0 — this patch changes documentation only; it authorizes no source, data, schema, Compose, or protected-zone mutation.

## Frozen remediation addendum R2 — published-only cleanup boundary

This addendum is `DOCS_ONLY`. It supersedes only the prior published-only remediation addendum. The original 15-node created-graph allow-list and its root/Calendar proof rules remain unchanged. No code, migration, production interface, deployment, or acceptance is authorized; all evidence remains `RED`/`NOT_RUN`.

### Current-scenario selector, manifest, slots, and writer boundary

For the current `inbound-request-v1` only, the sent-branch selector is the literal, reproducible plan fact `sent_finalized = false`. The current QA publish flow creates an inbound `mail.received.v1` event and `TC-MAIL-CALENDAR-0002` creates/publishes an invite and responds RSVP; it contains no compose, submission, or send action. The selector therefore cannot be inferred from a database row, asynchronous timing, Fake boot state, a client retry, or a later Mail worker result.

`newInboundRequestFixturePlan(nonce)` is extended only for a published fixture with canonical `QAInboundPublishedManifestV1`. It has `revision = "v1"` and a SHA-256 `manifestHash` over strict canonical JSON bytes. Those bytes include the ordered selected phase list `["published_mail_received","invite_materialized","rsvp_applied"]`, the literal `sent_finalized:false`, all four scheduled UUIDs and event types, and the complete canonical EnvelopeV1 bytes for the three required rows. The fourth slot is recorded as `expected:"absent"` and has no envelope bytes because no outbound-submission input exists in this scenario. The hash is never recomputed from database rows, payload scans, generated IDs, or a process-local map.

Before a writer is called, the manifest schedules exactly four named primary-key slots, each deterministically derived from the nonce and its literal slot name: `mailReceivedOutboxEventID` / `mail.received.v1`, `calendarInviteUpdatedOutboxEventID` / `calendar.invite.updated.v1`, `calendarRSVPChangedOutboxEventID` / `calendar.rsvp.changed.v1`, and `mailSentOutboxEventID` / `mail.sent.v1`. A sent slot is always a manifest fact so its required absence can be proved, but it is not an event to publish when the selector is false. UUID generation, collision retry, and Fake boot IDs never enter the schedule.

The only shared writer is the QA-private `PublishInTxWithID(ctx, tx, id, eventType, canonicalEnvelope)` seam. It creates exactly one `OutboxEvent` in the caller-provided already-open Ent transaction and returns only the created row or an error. It never begins, commits, rolls back, or retains `tx`; it has no production `MailEngine` interface change. Each selected phase keeps its own transaction boundary: its existing owning writer calls this seam while that writer's transaction is open. The contract explicitly forbids an assertion that the selected event writes share one transaction.

### Exact phase causality, closed proof, and absent sent projections

P3 writes only the scheduled `mail.received.v1` row in the inbound materialization writer and establishes `published_mail_received`. P4 is causally after P3: the Calendar invite materialization writer writes only the scheduled `calendar.invite.updated.v1` row and establishes `invite_materialized`. P5 is causally after P4: the RSVP writer writes only the scheduled `calendar.rsvp.changed.v1` row and establishes `rsvp_applied`. No selected phase may synthesize, replay, send, or create another phase's row.

The complete outbox proof is four direct, scheduled-primary-key lookups and nothing else. The first three IDs must each return exactly one row with the scheduled `id`, exact `event_type`, strict canonical EnvelopeV1 decoding, byte-for-byte equality to that slot's canonical envelope bytes, and decoded `event_id == scheduled id`. The `mailSentOutboxEventID` lookup must return no row. There is no `id NOT IN`, unbounded `OutboxEvent` query, payload search, aggregate/transaction field, relation endpoint, generated-ID, public replay/send, or arbitrary-unscheduled-row assertion. A missing required row, present sent row, duplicate/non-unique result, wrong ID/type/envelope/event ID, or any failed original root/Calendar proof is non-mutating `409`.

`MailSubmissionResult` does not exist and is not a permitted entity, table, edge, or cleanup target. With `sent_finalized=false`, `MailDraft` and `MailSubmission` are forbidden sent-only projections: the exact proven `MailAccount` has zero constrained `drafts` and zero constrained `submissions`, and the exact-account `MailMessage` set is exactly the one original inbound `mailMessageID`. Thus no additional result `MailMessage` may exist; there is no `MailSubmission` from which the scalar `result_mail_message_id` could refer to one. No sent-only node is selected, deleted, or rehydrated. The original 15-node deletion order remains authoritative; after full proof, cleanup deletes only the three present scheduled `OutboxEvent` IDs before that original closed graph, and proves all four scheduled IDs absent only after a successful cleanup commit.

P6 is `NOT_ADMITTED` for `inbound-request-v1` and remains `NOT_RUN`. It cannot be enabled by a retry or environment setting. A future scenario that truly invokes the existing outbound submission/finalization flow requires a separately approved replacement contract which freezes its selector, deterministic `MailDraft`/`MailSubmission`/result-`MailMessage` IDs, exact account/draft cardinalities and immutable fields, scalar `MailSubmission.result_mail_message_id` equality, its `mail.sent.v1` canonical envelope, and complete dependency-respecting deletion order. It may not reintroduce a nonexistent result entity.

### QA Fake binding V1, CAS closure, and mapless restart

`Controller.fixtures` is the sole owner of the QA-only binding; `mail.FakeMailEngine` keeps only its existing boot-scoped engine objects and receives no durable descriptor or production-port expansion. The private binding record is named `QAInboundFixtureBindingV1` and contains exactly `nonce`, `manifestRevision`, `manifestHash`, `engineFixture`, and `state`. Its legal states are `prepared`, `writing_mail_received`, `mail_received_commit_unknown`, `published_mail_received`, `writing_invite`, `invite_commit_unknown`, `invite_materialized`, `writing_rsvp`, `rsvp_commit_unknown`, `rsvp_applied`, reserved `writing_sent`, `sent_commit_unknown`, `sent_finalized`, `cleanup_claimed`, `cleanup_commit_unknown`, `cleanup_engine_failed`, and `cleaned`. The three sent-writing states are unreachable while this manifest selector is false.

The only binding API is private controller code: `storeQAInboundFixtureBindingV1`, `advanceQAInboundFixtureBindingV1CAS`, and `claimQAInboundFixtureBindingV1CleanupCAS`. Every compare-and-swap compares nonce, revision, hash, and prior state while `Controller.mu` is held; all database, engine, wait, and HTTP work is outside the mutex. A CAS into `writing_*` is the per-phase execution fence. The complete allowed transition table is closed:

| Source state | Allowed CAS result and required prior proof |
|---|---|
| `prepared` | `writing_mail_received` only; P3 is the sole writer. |
| `writing_mail_received` | successful committed exact P3 proof -> `published_mail_received`; known pre-commit failure after rollback -> `prepared`; commit error -> `mail_received_commit_unknown`. |
| `published_mail_received` | `writing_invite` only; P4 is the sole writer. |
| `writing_invite` | successful committed exact P4 proof -> `invite_materialized`; known pre-commit failure after rollback -> `published_mail_received`; commit error -> `invite_commit_unknown`. |
| `invite_materialized` | `writing_rsvp` only; P5 is the sole writer. |
| `writing_rsvp` | successful committed exact P5 proof -> `rsvp_applied`; known pre-commit failure after rollback -> `invite_materialized`; commit error -> `rsvp_commit_unknown`. |
| `rsvp_applied` | full original-plus-published proof -> `cleanup_claimed`. It is also the only terminal source for this selector. |
| `sent_finalized` | full original-plus-published proof -> `cleanup_claimed`; reserved for a separately admitted selected-sent scenario only. |
| `writing_sent` | successful committed exact P6 proof -> `sent_finalized`; known pre-commit failure after rollback -> `rsvp_applied`; commit error -> `sent_commit_unknown`; all three are forbidden for this selector. |
| `*_commit_unknown` | exact scheduled-PK reconciliation only: a complete exact row and all lower-phase proofs -> its target stable phase; an exact absent row plus the complete source-phase proof -> its source stable phase; any partial, foreign, non-unique, or unavailable result -> unchanged state and `409`. |
| `cleanup_claimed` | known pre-commit cleanup failure after rollback and renewed complete proof -> the manifest-selected terminal source (`rsvp_applied` here, `sent_finalized` only in a future admitted manifest); commit error -> `cleanup_commit_unknown`; confirmed database deletion followed by engine-destroy success -> `cleaned`; known engine-destroy failure -> `cleanup_engine_failed`. |
| `cleanup_commit_unknown` | all original roots and required scheduled rows still exactly present -> selected terminal source; all original roots and all four scheduled IDs exactly absent -> continue only the exact recorded engine destruction then `cleaned`; every other result -> unchanged state and `409`. |
| `cleanup_engine_failed` | no destructive retry, reconstruction, or deletion is authorized; return `409`. |
| `cleaned` | only the existing exact root-absent and four-slot-absent `204` proof. |

No other state authorizes deletion. `prepared`, `published_mail_received`, and `invite_materialized` cannot claim cleanup. A full original-plus-published graph proof and a successful CAS from only `rsvp_applied` or `sent_finalized` to `cleanup_claimed` are required before engine destruction or the first database delete. Failed/lost CAS is `409`.

`writing_*` is observed lock-free for at most three seconds using a monotonic clock, then gets exactly one fixture-local locked recheck. A commit-unknown state never retries a writer: it first takes the table's exact known-PK reconciliation path. A mismatch of nonce/revision/hash, illegal state, lost CAS, timeout, incomplete graph, or a binding that cannot be read as one exact record is non-mutating `409`.

On a mapless process restart, `Controller.fixtures` and therefore the Fake binding are absent. Only a fully proven deterministic original-plus-published database graph yields the distinct `engine_absent` result; it authorizes exact database cleanup without an engine delete and does not recreate, infer, or replace a binding. Absence alone is never ownership proof.

### Python `TC-MAIL-CALENDAR-0002` lifecycle

The exact public Python cleanup API is `MailCalendarQAClient.delete_published_fixture(published_fixture: FixtureSnapshot) -> None`. It validates the immutable published snapshot, sends the existing nonce-scoped DELETE request with the QA token, and accepts only empty `204`; it does not expose manifest revision/hash, worker control, engine control, database reset, Compose teardown, volume reset, or collision retry.

`test_published_invite_projects_bidirectional_mail_calendar_correlation_and_rsvp` must use the exact shape `created: FixtureSnapshot | None = None`, `published: FixtureSnapshot | None = None`, and a `try/finally`: create and publish inside `try`; in `finally`, call `delete_published_fixture(published)` when published is non-null, otherwise `delete_created_fixture(created)` when created is non-null. If the body already failed, cleanup failure is suppressed so the original exception remains primary; if the body passed and cleanup fails, the cleanup failure is raised. `TC-MAIL-CALENDAR-0002` metadata must replace its disposable-profile-teardown statement with this same public cleanup lifecycle.

### Packet graph and runtime status

Exactly ten independently reviewed packets are required. P1-P5 and P7-P10 are `NOT_RUN`; P6 is `NOT_ADMITTED` for this selector and therefore also `NOT_RUN`:

1. P1 canonical `QAInboundPublishedManifestV1`, literal selector, deterministic four-slot schedule, revision/hash, required strict envelopes, and sent-slot absence marker.
2. P2 QA-private `PublishInTxWithID` writer; depends on P1.
3. P3 inbound materialization `mail.received.v1`; depends on P1/P2.
4. P4 Calendar invite materialization `calendar.invite.updated.v1`; depends on P3.
5. P5 RSVP `calendar.rsvp.changed.v1`; depends on P4.
6. P6 reserved sent finalization `mail.sent.v1`; `NOT_ADMITTED` for `inbound-request-v1`, and would require a separately approved selected-sent contract before depending on P5.
7. P7 closed original-plus-published database proof, four known-PK proof, and sent-only absence guards; depends on P3/P4/P5.
8. P8 `QAInboundFixtureBindingV1`, lock-free wait, locked recheck, exact-PK commit reconciliation, CAS, and `engine_absent`; depends on P1/P7.
9. P9 Python published cleanup `finally` lifecycle, primary-failure preservation, and TC metadata sync; depends on P7/P8.
10. P10 retained-substrate Docker evidence; depends on P3-P5/P7-P9.

No packet after P2 is acceptable until P1 and P2 pass independent review; no runtime proof is authorized before P10. The sole eventual runtime command remains:

```powershell
docker compose -f docker-compose.local.yml run --rm python-integration
```

It must eventually prove the three required outbox rows, the exact absent sent slot and sent-only projections, separate writer transaction boundaries, exact-ID one-row writes, closed proof, V1 revision/hash/CAS and commit-unknown outcomes, timeout then locked recheck, mapless `engine_absent` cleanup, `finally` primary-failure preservation, root-and-slot-absent `204`, and non-mutating `409`.

### R3 shared OutboxOrigin receipt dependency

This R3 subsection supersedes only published-R2 statements about scheduled
physical OutboxEvent IDs, the QA-private PublishInTxWithID seam, four-slot
primary-key proof, absent sent outbox ID, and known-PK reconciliation. It
preserves the original 15-node created graph, all original root and Calendar
relation proof, literal sent_finalized=false selector, response codes, binding
CAS, private QA surface, no-retry rule, P1-P10 structure, and NOT_RUN status.

R2 now depends on the frozen shared contract
.planning/active/august-2026/09-OUTBOX-ORIGIN-RECEIPT-CONTRACT.md and its
execution plan .planning/active/august-2026/10-OUTBOX-ORIGIN-RECEIPT-EXECUTION.md.
No R2 writer, proof, cleanup, controller, test, migration, or runtime packet
is admitted until Project Outbox P1a accepts its reserved
20260831000000_qa_project_outbox_authority.sql migration and the shared next
serial migration 20260831000001_outbox_event_origin.sql is independently
accepted. P1a QA tables remain separate from this R2 authority.

The published manifest keeps immutable selector/phase/root facts only. It does
not schedule, inject, persist, or assert a physical OutboxEvent ID. P3/P4/P5
use shared PublishInTxWithReceipt inside their existing open transaction and
validate its returned physical/envelope receipt before commit:

| Phase | event_type | immutable origin pair |
|---|---|---|
| P3 | mail.received.v1 | mail_message / exact proven MailMessage ID |
| P4 | calendar.invite.updated.v1 | calendar_event / exact proven CalendarEvent ID |
| P5 | calendar.rsvp.changed.v1 | calendar_attendee / exact proven RSVP CalendarAttendee ID |

The receipt is transient transaction proof, not a QA binding table or durable
descriptor. The existing QA binding remains the only process-local fence.
Mapless restart derives the physical row only from the shared pair after the
original graph proof.

R2 outbox proof replaces scheduled-PK lookup with exactly:

    origin_kind = required literal
    AND origin_id = exact already-proven domain row ID
    AND event_type = required literal

Each lookup returns exactly one row. Only that row is decoded and checked for
the immutable pair, event type, strict EnvelopeV1 identity and the already
proven MailMessage, CalendarEvent, or CalendarAttendee relation. No payload
scan, name/nonce/organization lookup, ID NOT IN, aggregate, unrestricted
OutboxEvent query, generated-ID schedule, payload predicate or ID injection is
allowed. Zero, duplicate, partial, foreign, unavailable or malformed results
retain the current fence and return non-mutating 409.

Traversal is fixed: roots and P3 from MailMessage; constrained inbound
relation to CalendarEvent and P4; constrained event/attendee relation to P5.
After all three proofs pass, cleanup deletes only the returned three physical
OutboxEvent IDs before the unchanged 15-node order. Duplicate selected rows
are never deduplicated or deleted.

The false sent branch has no outbox absence assertion. It must not look up
mail.sent.v1 by deterministic, generated, origin, or other outbox ID. Its
negative proof remains exactly the proven account's zero MailDraft and
MailSubmission sets and single original inbound MailMessage; no result
message, send action, or sent-only cleanup target exists.

The ten R2 packets remain NOT_RUN: P1 removes physical scheduling facts; P2
adopts the shared prerequisite and is blocked on P1a plus shared O1/O2; P3-P5
write/validate their listed origins; P6 remains NOT_ADMITTED; P7 performs
three origin/type cardinality proofs and sent-only guard; P8 replaces known-PK
reconciliation with exact origin/type proof; P9 keeps public finally cleanup;
P10 remains the only Docker evidence packet. Writer transactions stay
separate, and the publisher never opens, commits, rolls back, retries or
retains them.

### R4 authoritative shared-receipt dependency clarification

This append-only R4 clarification supersedes only R3 phrases that could be
read to give P4 origin/receipt ownership to a tombstone/cancelled branch, to
leave strict stored-envelope decoding or Origin copying implicit, or to alter
the O2/O4/O6 order. It preserves P1a's `20260831000000` prerequisite, the
shared `20260831000001` slot, immutable nullable origin pair, no payload scan,
false-sent boundary, P1-P10 graph, separate writer transactions and NOT_RUN.

R2 P4 is exclusively the `CalendarInboundIngestApplyRequest` emission of
`calendar.invite.updated.v1`, with `calendar_event` and the exact persisted
CalendarEvent ID. Inbound tombstone series/occurrence and cancelled-envelope
branches retain generic `PublishInTx`; they have no origin-bearing receipt,
R2 lookup or cleanup ownership.

R2 depends on additive O2 before P3/P4/P5: O2 defines public
`events.DecodeEnvelopeV1([]byte) (events.EnvelopeV1, error)` in
`backend/internal/platform/events/envelope.go` and the publisher receipt seam
in `backend/internal/platform/outbox/publisher.go`. Decode uses
`DisallowUnknownFields`, one top-level object/no trailing token,
schema/required-field validation through `NewEnvelopeV1`, and re-marshal
canonical-byte equality. A nil receipt Origin remains nil; a non-nil receipt
Origin is a fresh defensive copy and never an alias.

Only O6 in `backend/internal/qamailcalendar/controller_delete.go` calls the
decoder, and only after its exact `origin_kind/origin_id/event_type` predicate
returns exactly one physical OutboxEvent. It validates decoded EventID and the
already-proven relation before deleting that returned physical ID. O6 neither
uses the decoder as a selector nor scans another outbox row. The fixed order is
P1a/O0 -> O1 -> O2 -> O3 -> O4 (ApplyRequest only) -> O5 -> O6.

### R5 authoritative cleanup-lifecycle replacement

Status: `DOCS_FROZEN` / `NOT_RUN`. This R5 supersedes only R2-R4 rules for
binding-state granularity, P4/P5 phase callbacks, mapless-absent `204`, and
cleanup-proof retention. It preserves O1-O6 order and origin semantics, the
false-sent boundary, the closed 15-node graph, and the public `204`-or-`409`
contract. It authorizes no source, schema, deployment, test, or runtime work.

There are no production P4/P5 Controller observer hooks. The existing
process-local record becomes `QAInboundFixtureBindingV2`; it is the only
same-process lifecycle fence. `Published` means that P3 is committed, but it
does not itself authorize cleanup. Cleanup admission first proves, in one
transaction, all immutable 15-node roots and constrained Calendar closure,
the false-sent guard (zero `MailDraft`, zero `MailSubmission`, and exactly one
inbound `MailMessage`), and the three exact outbox proofs. Each proof has
cardinality one and the exact tuple: P3 `mail_message`/proven MailMessage ID/
`mail.received.v1`; P4 `calendar_event`/proven CalendarEvent ID/
`calendar.invite.updated.v1`; P5 `calendar_attendee`/proven RSVP attendee ID/
`calendar.rsvp.changed.v1`. Each returned row is strict-canonically decoded
and must match its immutable origin, type, EventID, and proven relation.

Only after that complete proof may a CAS move the same immutable V2 record
from `Published` to `CleanupClaimed`, attaching value copies of exactly six
Calendar domain IDs: receipt, objectState, event, RSVP attendee,
replyDelivery, and mutationReceipt. It never retains or schedules physical
OutboxEvent IDs. In the same transaction it re-proves the complete fence,
then deletes the three transient returned physical IDs in causal P3 -> P4 ->
P5 order, followed by the unchanged 15-node deletion order. No payload,
broad, name, nonce, `mail.sent`, or physical-ID retention/scheduling query is
permitted.

A known pre-commit failure rolls back and CASes V2 back to `Published`,
clearing only the attached cleanup proof. A commit error becomes
`CleanupCommitUnknown` retaining the domain proof: exact-present proof may
restore `Published`; exact all-known-ID absence may continue only to engine
destroy; partial proof returns `409`. After a confirmed commit, engine destroy
is outside the transaction. Its success CASes to retained `Cleaned`; failure
is terminal `CleanupEngineFailed` and returns `409`. A same-process `Cleaned`
retry returns `204` only after exact known-ID absence is re-proven.

On mapless restart, no binding is reconstructed. A full present graph may be
derived only from deterministic roots and cleaned database-only; absent,
partial, ambiguous, or commit-error mapless states are fail-closed `409`, never
guessed `204`. Mapless cleanup never invokes engine destroy. The atomic
execution decomposition is: roots proof; external-closure guard; Calendar
proof; O6 outbox proof; V2 lifecycle coordinator; Python `finally` and tests;
then Docker gates. Every packet remains `NOT_RUN` until separately accepted.

### R5 strict V2 lifecycle and reconciliation rules

`QAInboundFixtureBindingV2` is one exact comparable private value: existing
nonce, immutable plan identity, engineFixture, persisted roots, and state,
plus `cleanupProof`. `cleanupProof` has an explicit presence bit and value
copies of exactly six UUIDs: CalendarInboundMessageReceipt,
CalendarInboundObjectState, CalendarEvent, RSVP CalendarAttendee,
CalendarReplyDelivery, and CalendarMutationReceipt. It contains no pointer,
envelope, row, or physical OutboxEvent ID. Controller map and Fake engine are
created or replaced together at boot; all record mutation is private; a
`Cleaned` record is retained and is never deleted.
The cleanupProof is absent in Creating, Created, Deleting, Publishing,
PublishUnknown, Published, and CleanupFailed; it is present and exact in
CleanupClaimed, CleanupCommitUnknown, CleanupEngineFailed, and Cleaned.
CleanupCommitUnknown exact-present -> Published clears cleanupProof in that CAS.

Legal V2 states retain `Creating`, `Created`, `Deleting` (created-only legacy),
`Publishing`, `PublishUnknown`, `Published`, and `CleanupFailed`
(created-only legacy), and add `CleanupClaimed`, `CleanupCommitUnknown`,
`CleanupEngineFailed`, and `Cleaned`. Existing create and P3 transitions are
unchanged. `Published -> CleanupClaimed` requires the complete proof below.
`CleanupClaimed` permits only confirmed rollback -> `Published`, commit error
-> `CleanupCommitUnknown`, or confirmed commit plus engine success -> `Cleaned`
/ engine failure -> terminal `CleanupEngineFailed`. `CleanupCommitUnknown`
permits exact-present -> `Published`, exact-absent -> the single engine step
then `Cleaned` or `CleanupEngineFailed`. `Cleaned` is same-process idempotent;
`CleanupEngineFailed` is terminal. Every CAS under `Controller.mu` compares
nonce, plan identity, persisted roots, engineFixture, prior state, and expected
cleanupProof presence/value; database, engine, waits, and HTTP remain outside
the mutex.

Published cleanup and reconciliation use Ent `BeginTx` at `SERIALIZABLE`; all
selected exact rows/sets are locked, and the full predicate plus external
closure reads occur in that transaction through commit. Normal
`READ COMMITTED` cleanup is forbidden. A serialization/deadlock failure is
pre-commit only after rollback is confirmed; only then may a renewed full proof
precede restoring `Published`. Otherwise return `409` or retain
`CleanupCommitUnknown`.

`exact-present` means the full closed 15-node graph, external-closure guard,
all three cardinality-one origin proofs, and false-sent guard. `exact-absent`
means every deterministic root/relation ID is absent; all six retained Calendar
IDs are absent; each exact origin triple has zero rows; MailDraft and
MailSubmission sets are zero; and the original inbound MailMessage is absent.
Physical OutboxEvent IDs are never retained or required.

Before the synchronous public RSVP has succeeded and the full P5 closure,
DELETE is non-mutating `409`. A successful public RSVP response after the
completed mutation receipt is the readiness signal. Python `finally` invokes
published cleanup only after that signal; an earlier body failure may make
cleanup fail while preserving the primary test exception.

After full proof, CAS `Published -> CleanupClaimed`, re-prove in the same
transaction, then delete transient selected physical OutboxEvent IDs in this
exact order: P3 -> P4 -> P5; then delete the unchanged closed graph in exact
order: ReplyDelivery -> MutationReceipt -> InboundReceipt -> ObjectState ->
Attendee -> Event -> messageFolder -> message -> thread -> folder -> identity
-> account -> domain -> user -> organization.

Restore `Published` and clear cleanupProof only after rollback success is
confirmed. Rollback error or unavailable rollback produces
`CleanupCommitUnknown` and retains the proof. After DB commit and engine
destroy success, a failed or lost CAS to `Cleaned` returns `409`; do not retry
the DB or engine operation and retain current claim/unknown evidence.

A mapless complete-present graph is admissible only while a monotonic
fresh-boot-unmutated Controller flag proves the Controller and its task-owned
Fake were constructed together and that Fake has accepted no seed or mutation
in this boot; it then performs DB-only cleanup. The flag starts true only in
that constructor path and becomes permanently false before the first QA seed
or engine mutation. A missing record with a false flag, absent/partial/
ambiguous graph, or mapless commit error
returns fail-closed `409`, never guessed `204`. No persistent QA table/schema,
payload/broad/name/nonce scan, `mail.sent` lookup, or physical-ID scheduling or
retention is allowed. Status remains `DOCS_FROZEN` / `NOT_RUN`.
