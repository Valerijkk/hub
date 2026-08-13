# Phase 143 — Calendar schema, ScheduledCall backfill and cutover plan

> **Status:** CONTRACT_0_1_7_REMEDIATION_REVIEW_PENDING · PRODUCTION_ARTIFACT_BLOCKED_BY_FOUNDER_REVIEW
> **Runtime:** NOT_RUN
> **Owner decision:** `docs/mail/FOUNDER-QUESTIONS.md` #12

## Current repository truth

- Ent schemas are the checked-in model source. Generated Ent files are ignored and must be
  regenerated only inside the accepted Docker launcher.
- Development may call `client.Schema.Create(ctx)` only with `DEV_MODE=true`.
- Production code comments require `Atlas migrate apply`, but the repository has no
  executable `atlas.hcl`, migration-directory owner, or CI/deploy apply command.
- `backend/migrations/*.sql` contains reviewed historical SQL, but this does not establish
  the future production migration authority.
- Database migrations and this ADR require founder review before integration.

No SQL file or apply command is authorized until the owner resolves that mismatch. Ent
schema work may proceed independently under the exact founder-authorized Terra route, but it is
not production-migration evidence.

## Target schema invariants

1. One master `CalendarEvent` per `(calendar_owner_id, uid)`.
2. One occurrence exception per `(calendar_event_id, recurrence_id)`.
3. One series attendee per `(calendar_event_id, normalized_email)` where
   `recurrence_id IS NULL`.
4. One occurrence attendee per `(calendar_event_id, normalized_email, recurrence_id)`
   where `recurrence_id IS NOT NULL`.
5. At most one event per non-null `scheduled_call_id`.
6. Timed events have only `starts_at/ends_at` with `end > start`; all-day events have only
   `start_date/end_date_exclusive` with exclusive end after start.
7. Calendar idempotency receipts are unique by organization, authenticated actor,
   operation and hashed idempotency key.

The two attendee constraints must be partial unique indexes. A normal nullable composite
unique constraint is insufficient in PostgreSQL.

## Backfill scope and mapping

Every row satisfying `scheduled_calls.deleted_at IS NULL` enters the candidate set, including
paused and completed rows. Otherwise a later legacy `is_active=true` could resurrect an
unlinked second authority after cutover. Soft-deleted rows remain non-reactivatable and are
excluded with their counts recorded.

| Target | Source/mapping |
|---|---|
| organization | the call's `Project.organization`; it must exist and match creator/participant organizations |
| owner + organizer | `scheduled_calls.created_by`; the user must exist, be active, belong to the project and project organization, and be a member of the exact call channel |
| UID | `urn:uuid:<scheduled_call_id>` |
| title | existing scheduled-call title |
| description | existing scheduled-call description |
| timed start | existing `scheduled_at` instant |
| timed end | start plus existing duration |
| timezone | creator's current valid IANA `users.timezone` for non-recurring rows; recurring rows follow the separate recurrence gate below |
| recurrence | validated and normalized legacy RRULE only after its future occurrence set passes the recurrence gate |
| reminder | existing `reminder_minutes`, null or `0..40320` |
| location/exdates | empty; legacy ScheduledCall has no corresponding data |
| source/status/sequence/version | `hub` / `confirmed` / `0` / `1` |
| call execution state | preserve `scheduled_calls.is_active`; inactive does not cancel the historical/canonical event |
| organizer attendee | normalized creator email, role `required`, `accepted`, `responded_at=scheduled_calls.created_at` |
| participant attendees | distinct active same-project/same-organization M2M participant emails, role `required`, `accepted`, `responded_at=scheduled_calls.created_at` |
| reply delivery | `not_required` for backfilled internal attendees |
| payload hash version/hash | both null: backfill is an internal Hub source, not an admitted inbound iTIP payload; never synthesize persistence identity from outbound serializer bytes |
| audit timestamps | preserve source `created_at` and `updated_at`; migration evidence records apply time separately |
| future invitees | `needs_action` through normal service behavior |
| delivery side effects | none: no invitation, notification, NATS, or outbox rows |

Creator duplication in the participant edge collapses to one attendee. The write is
idempotent on `scheduled_call_id` and must not overwrite an already-linked canonical event.

The source relation must be coherent before any mapping: `ScheduledCall.project_id` equals
`Channel.project_id`, the project has exactly one organization, and creator plus every Hub
participant are active members of that project and belong to that organization; the creator
also belongs to the exact call channel. Missing, foreign, inactive or ambiguous identity
fails the row; the migration never drops a participant or infers an organization from a
different edge. Duration uses checked arithmetic and must produce a representable positive
interval. Title/description/RRULE/reminder values must fit the target bounds.

## Recurring-call semantic gate

Legacy recurrence has no trustworthy original local `DTSTART`: the live create path parses
the raw RRULE and replaces `scheduled_at` with a moving next occurrence, while the legacy
calendar aggregate injects that moving UTC value as `DTSTART`. Existing clients also emit
`BYHOUR/BYMINUTE/BYSECOND`. Reinterpreting the row in the creator's current IANA timezone
can therefore change future instants and DST behavior.

Founder question 18 chooses the product rule. Until it is answered, any non-deleted recurring
row blocks the whole production backfill; it is not skipped and does not receive a guessed
timezone. The accepted preflight must expand, with the same pinned recurrence library:

1. the current worker interpretation of the raw RRULE;
2. the current legacy aggregate interpretation anchored at `scheduled_at` UTC; and
3. the proposed canonical event interpretation, including target timezone.

The report compares ordered UTC instants through the earlier of the finite rule end or a
bounded oracle of 400 future occurrences and two years, and records parser/library version,
anchor, timezone and hashes. Any parse failure, unsupported canonical subset, cap hit or
non-identical sequence blocks the entire apply. A bounded match is migration evidence, not
permission to silently reinterpret an unbounded rule; the founder-selected rule remains
the authority.

## Canonical cutover and compatibility bridge

Backfill alone is insufficient. At the write cutover, `CalendarEvent` becomes the sole
business authority for title, description, start/end, timezone, recurrence/exdates/
exceptions, reminder, status, organizer and attendees. A linked `ScheduledCall` is an
execution/legacy-read projection containing channel/project identity and the next effective
call occurrence. Its duplicated schedule fields are mirrors, not an independent source.

- New legacy ScheduledCall creation writes one canonical event, attendees, call projection
  and invitation outbox in one transaction. A calendar-v2 create may attach only an active,
  unlinked call owned by the authenticated organizer after same-organization/project/channel
  authorization. The normalized request must equal the locked call's canonical-mapped title,
  description, start/duration, recurrence, reminder, organizer and participant set; attach
  never overwrites a mismatch. v1 does not synthesize a call from a nested calendar request.
  Ordinary CalendarEvents allow email-only guests, but a linked-call event rejects them under founder
  question 23's default until guest join credential delivery is specified.
- Calendar-v2 and legacy call edits lock `CalendarEvent` before `ScheduledCall`, mutate the
  event under canonical sequence/version policy, then mirror the projection atomically.
  The legacy request has no `If-Match`, so it retains serialized last-write-wins semantics
  by reading the locked current event version; the existing response shape is unchanged.
- `scheduled_call_id` is immutable after canonical event creation. A legacy `DELETE` cancels
  the linked canonical event and soft-deletes/deactivates the call in the same transaction.
  `is_active=false/true` is provisionally a call-execution pause/resume only, never an event
  cancellation or resurrection; founder question 19 owns that product decision.
- ScheduledCall participants mirror only distinct internal non-organizer series attendees in
  `needs_action|accepted|tentative`; declined/removed series attendees leave the projection.
  An occurrence-specific RSVP never mutates this edge.
- Event cancellation/deletion deactivates the projection. One-time execution completion
  deactivates only the projection and preserves the historical event. Cancelling the next
  recurrence instance atomically advances the projection to the next non-cancelled
  occurrence.
- The due worker derives membership and the next occurrence from the canonical event,
  exdates and exceptions. Under the same event-before-call lock it appends one durable
  `calls.channel.{channelId}.scheduled_due.v1` outbox row and advances/deactivates the
  projection atomically. The payload carries a deterministic occurrence key; NATS remains
  at-least-once and consumers
  deduplicate that key. It never advances the event master or emits a second reminder.
- The legacy aggregate emits a linked pair once as the existing `call` kind, using canonical
  temporal truth. It emits only unlinked canonical events as the additive `event` kind.
  Availability does the inverse for privacy/source truth: all canonical events once as
  `event`, plus active unlinked calls only while a pre-cutover transition flag is enabled.
  After activation, any unlinked non-deleted row is an invariant failure, not a fallback.

Cutover is all-or-nothing. It stays disabled until every non-deleted call is linked,
all bridge postconditions pass, and no application writer or worker can mutate a linked
projection outside the bridge. Reconciliation reports any event/call mismatch and fails
closed; it does not choose a winner automatically.

## Mandatory preflight report

Before an apply artifact can be accepted, a read-only report must enumerate and count:

- non-deleted source calls by active/inactive state, excluded soft-deleted rows and
  already-linked targets;
- missing/orphan/inactive creator, organization, project, channel or participant references;
- project/channel mismatch, creator exact-channel membership failure, or creator/participant
  project/organization membership mismatch;
- creators/participants without a normalized email;
- invalid or unloadable creator timezone;
- non-positive duration or impossible interval;
- out-of-range reminder, target-bound overflow, RRULE parse/subset/length failure and every
  recurring semantic-oracle comparison;
- duplicate scheduled-call links or target UID collisions;
- existing linked event/call field mismatch;
- expected source-to-event and distinct-attendee totals plus deterministic content hashes.

Any unresolved row fails closed. The migration must not silently choose a replacement
owner, email, timezone, UID, or duration.

## Apply and recovery contract

1. Take a database backup and record source revision plus preflight hashes/counts.
2. Create tables/constraints before backfill, inside the migration mechanism chosen by the
   founder.
3. Stop/guard legacy scheduled-call writers and the due worker; verify the guard before the
   first backfill write.
4. Insert events and attendees in bounded transactions; record exact affected counts.
5. Run target invariants, source/target hashes and event/call coherence postconditions.
6. Enable only bridge-owned writers/readers/worker after 100% non-deleted-call linkage passes.
7. Before production apply, rollback is removal of the unapplied artifact. After apply,
   recovery is forward-only: disable the feature flag, preserve rows for audit, and ship a
   reviewed corrective migration. Never delete user calendar data automatically.

## Required future evidence

- accepted launcher identity and bounded Docker database target;
- generated Ent diff inspection without committing generated files;
- migration lint/dry run from empty schema and realistic pre-feature snapshot;
- backfill preflight, apply and postcondition reports;
- recurrence semantic-oracle report and founder decision 18;
- duplicate/retry proof, no-historical-outbox proof and bridge atomicity proof;
- concurrent calendar-v2/legacy edit, occurrence-cancel, aggregate/availability dedupe and
  due-worker recovery proof;
- independent database review and founder approval.

Current evidence is static analysis only. Generation, migration, database integration,
rollback rehearsal, tests, build, lint and validators are `NOT_RUN`.
