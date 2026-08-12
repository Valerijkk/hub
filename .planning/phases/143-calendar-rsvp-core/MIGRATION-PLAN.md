# Phase 143 — Calendar schema and ScheduledCall backfill plan

> **Status:** CONTRACT_READY · PRODUCTION_ARTIFACT_BLOCKED_BY_FOUNDER_REVIEW
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
schema work may proceed independently once the exact Spark runtime is available, but it is
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

Only rows satisfying `scheduled_calls.is_active = true AND deleted_at IS NULL` enter the
candidate set.

| Target | Source/mapping |
|---|---|
| owner + organizer | `scheduled_calls.created_by` |
| UID | `urn:uuid:<scheduled_call_id>` |
| title | existing scheduled-call title |
| timed start | existing `scheduled_at` instant |
| timed end | start plus existing duration |
| timezone | creator's current valid IANA `users.timezone`; current schema default is `UTC` |
| organizer attendee | creator email, `accepted` |
| participant attendees | distinct M2M participant emails, `accepted` |
| future invitees | `needs_action` through normal service behavior |
| delivery side effects | none: no invitation, notification, NATS, or outbox rows |

Creator duplication in the participant edge collapses to one attendee. The write is
idempotent on `scheduled_call_id` and must not overwrite an already-linked canonical event.

## Mandatory preflight report

Before an apply artifact can be accepted, a read-only report must enumerate and count:

- active source calls and already-linked targets;
- missing/orphan creator, organization, channel or participant references;
- creators/participants without a normalized email;
- invalid or unloadable creator timezone;
- non-positive duration or impossible interval;
- duplicate scheduled-call links or target UID collisions;
- expected source-to-event and distinct-attendee totals.

Any unresolved row fails closed. The migration must not silently choose a replacement
owner, email, timezone, UID, or duration.

## Apply and recovery contract

1. Take a database backup and record source revision plus preflight hashes/counts.
2. Create tables/constraints before backfill, inside the migration mechanism chosen by the
   founder.
3. Insert events and attendees in bounded transactions; record exact affected counts.
4. Run postconditions matching every target invariant and compare pre/post counts.
5. Enable application writes only after the schema and backfill postconditions pass.
6. Before production apply, rollback is removal of the unapplied artifact. After apply,
   recovery is forward-only: disable the feature flag, preserve rows for audit, and ship a
   reviewed corrective migration. Never delete user calendar data automatically.

## Required future evidence

- accepted launcher identity and bounded Docker database target;
- generated Ent diff inspection without committing generated files;
- migration lint/dry run from empty schema and realistic pre-feature snapshot;
- backfill preflight, apply and postcondition reports;
- duplicate/retry proof and no-outbox proof;
- independent database review and founder approval.

Current evidence is static analysis only. Generation, migration, database integration,
rollback rehearsal, tests, build, lint and validators are `NOT_RUN`.
