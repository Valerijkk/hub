# Phase 143 — Calendar schema, ScheduledCall backfill and cutover plan

> **Status:** BOUNDED_CONTROL_PLANE_FROZEN · SB-A/SB-B_STATIC_IMPLEMENTED · EXECUTABLE_NOT_RUN
> **Runtime:** NOT_RUN
> **Owner decision:** founder instruction 2026-08-22; design:
> `docs/superpowers/specs/2026-08-22-mail-calendar-production-integration-design.md`
> **Bounded lineage parent:** `f21e3f7748eac55a4b16746b078e88801f11829c`

## Current repository truth

- Ent schemas are the checked-in model source. Generated Ent files are ignored, remain
  uncommitted, and may be produced only inside the accepted bounded Docker operation.
- Development may call `client.Schema.Create(ctx)` only with `DEV_MODE=true`.
- The SQL files directly under `backend/migrations/*.sql` are historical additive operator
  artifacts for already-existing databases. They predate Atlas control, have repeated date-only
  version prefixes, have no `atlas.sum`, and are not an Atlas-replayable migration lineage. They
  must not be renamed, re-hashed, replayed as a fresh-database history, or silently imported into
  the Mail/Calendar lineage.
- The only future Atlas lineage for this workstream is the bounded child directory
  `backend/migrations/atlas/mail-calendar-v1/`. It starts from the exact pre-feature master
  `f21e3f7748eac55a4b16746b078e88801f11829c` and contains only its full baseline, the reviewed
  additive Mail/Calendar delta, and `atlas.sum`.
- The previous migration-mechanism/owner mismatch is resolved by the founder instruction dated
  2026-08-22. It is historical and no longer blocks authoring the Atlas configuration or a
  reviewed migration packet.
- No generation, SQL diff, lint, apply command or database evidence has run yet. Those executable
  steps still require the accepted isolated `schema-maintenance` Compose profile and remain
  `NOT_RUN`. Ent generation and its dependent normal Docker build also remain `NOT_RUN`.
- `DEV_MODE=true` may continue to call `client.Schema.Create(ctx)` for development only;
  production uses the versioned migration directory.

## Static control-plane artifact

The `local` environment in `backend/atlas.hcl` retains the checked-in feature
source `ent://ent/schema`, fixed bounded lineage, and fail-closed profile-supplied
`DATABASE_URL` and task-owned `ATLAS_DEV_URL`. SB-A is the committed immutable
pre-feature source carrier for parent `f21e3f7748eac55a4b16746b078e88801f11829c`;
SB-B statically consumes it image-locally, verifies its exact archive identity and
exposes only `prefeature_baseline` at
`ent:///opt/hub/prefeature/backend/ent/schema`. No credential, host Docker URL,
Docker socket, bind mount, `.git` capability, or shared database URL is admitted.

Atlas remains the deterministic diff, lint, checksum and apply tool, never a
second model authority. Production may apply only the independently reviewed
checked-in bounded SQL plus its exact `atlas.sum`. Direct host Atlas, Ent,
database, build, validator, or test commands remain forbidden. SB-A/SB-B/SB-C are
static source-control implementation only: no Docker image build, Atlas inspect,
diff, lint, checksum or apply, generated SQL, database replay, build, tests,
validators or runtime validation has run; all remain `NOT_RUN`.

## Bounded Atlas lineage contract

The lineage is ordered and immutable after acceptance:

1. `backend/migrations/atlas/mail-calendar-v1/20260823000000_mail_calendar_baseline.sql`
   is the complete PostgreSQL desired schema generated from exact parent
   `f21e3f7748eac55a4b16746b078e88801f11829c`. A new empty database applies it normally. An
   existing database may mark this baseline as already applied only after an independent
   schema-equivalence report proves that every baseline object already exists with the same
   shape. A mismatch is a hard stop; `--baseline` is never a repair mechanism.
2. `backend/migrations/atlas/mail-calendar-v1/20260823000001_mail_calendar_core.sql`
   is the additive delta from that baseline to the exact reviewed feature schema. It may create
   only the Mail/Calendar tables, columns, edges, indexes and constraints selected below.
3. `backend/migrations/atlas/mail-calendar-v1/atlas.sum` covers both immutable SQL files. Any
   reviewed SQL correction is followed by a fresh checksum and another exact-byte review. An
   accepted migration is never edited; later corrections are forward-only migrations.

The schema-source allowlist is exact:

- Calendar: `calendarattendee.go`, `calendarevent.go`, `calendareventexception.go`,
  `calendarinboundmessagereceipt.go`, `calendarinboundobjectstate.go`,
  `calendarinboundreplyreceipt.go`, `calendarmutationreceipt.go`, and
  `calendarreplydelivery.go`;
- Mail: `mailaccount.go`, `mailattachment.go`, `mailboxpolicy.go`, `maildomain.go`,
  `maildraft.go`, `mailfolder.go`, `mailidentity.go`, `maillabel.go`, `mailmessage.go`,
  `mailmessagefolder.go`, `mailmessagelabel.go`, `mailmutationreceipt.go`,
  `mailmutationreceiptmember.go`, `mailmutationreceipttarget.go`, `mailobject.go`,
  `mailrule.go`, `mailsignature.go`, `mailsubmission.go`, `mailthread.go`, and
  `mailusersettings.go`;
- bounded shared additions: `recentcontact.go` plus only the reviewed Mail/Calendar edge or field
  changes in `organization.go`, `user.go`, and `scheduledcall.go`.

The delta rejects every table or shared-field change outside that allowlist, every `DROP`,
`TRUNCATE`, destructive type conversion, implicit data deletion, constraint weakening, unreviewed
cascade, and non-null alteration that cannot be proven safe for existing rows. Atlas output is a
candidate, not authority: any unexpected object stops the packet and returns ownership to a new
schema review. The schema SQL contains no ScheduledCall data backfill, feature enablement, provider
I/O, NATS/outbox emission, or runtime cutover.

## Generated Ent and dependent adapters

Generated files below `backend/ent/` remain image-local and uncommitted. Static SB-C verifies the
LF-normalized full `go.mod`/`go.sum` lock and exact Atlas/Ent pins, requires `GOWORK=off`, rejects
`replace`, `exclude`, symlinks and nonregular inputs, then copies only `go.mod`, `go.sum`,
`ent/generate.go` and `ent/schema` into two fresh isolated writable module roots. Each pass runs
the fixed offline `go run -mod=readonly entgo.io/ent/cmd/ent generate ./ent/schema`; repository
`go:generate` directives are audited but never executed. Canonical all-source manifests must be
unchanged before/after each pass, and both generated-output manifests must match exactly before
one bounded SHA-labelled manifest is emitted. No generated source may egress or be staged; this
prerequisite is not a dependent compile, whose normal Docker build remains `NOT_RUN`.

P3-C (`backend/internal/mail/mailbox_ent_calendar_part.go` and the required narrow change in
`backend/internal/mail/mailbox_ent_repository.go`) and D2
(`backend/internal/calendar/event_ent_transaction.go` and
`backend/internal/calendar/event_ent_inbound_message.go`) compile against that same image-local
generation. This dependency map grants no source authority by itself; each P3-C/D2 packet retains
its existing contract, ordering and independent review gates.

## Isolated schema-maintenance execution contract

Generation and Atlas work belong to a non-QA Compose profile named `schema-maintenance`; it does
not expand or modify the closed `qa` and `qa-maintenance` testframework catalogs. Its public
operations are fixed baseline, diff, generate and verify entrypoints rather than an arbitrary
command shell. `generate` is a prerequisite determinism proof only and emits no source files.
The exact Atlas module/version must match the reviewed Go module pin, and the Go and PostgreSQL
images must use reviewed immutable digests recorded in a checked-in tool lock.

Every service is non-root and `read_only`, drops all capabilities, enables
`no-new-privileges`, uses `init`, publishes no port, and mounts no Docker socket or host/named
volume. Baseline/diff/verify join only the dedicated internal database network; generate instead
has `network_mode: none` and no database environment or dependency. The schema tool is bounded to
`pids_limit: 128`, `cpus: 1.0`, `mem_limit: 1g`, a 512 MiB noexec `/work` tmpfs and a 64 MiB
noexec `/tmp` tmpfs. Generate alone adds a 256 MiB RAM-only executable tmpfs split into separate
per-pass TMPDIR/GOTMPDIR paths; source, HOME, XDG and GOCACHE remain noexec. Its disposable
PostgreSQL peer is bounded to `pids_limit: 128`, `cpus: 1.0`, `mem_limit: 1g` and a 512 MiB data
tmpfs. Each operation has a 15-minute wall-clock limit, no automatic retry, no persistent runtime
cache and unconditional `run --rm` cleanup. Generated SQL/checksum candidates leave the container
only as bounded, hashed review output and become repository files only in a separate exact-path
authoring packet.

## Implemented source controls and next packets

- **SB-A, immutable source carrier:** committed static input at
  `192f23686a69c036ddc03d35ea04192b30780936`; it supplies only the reviewed
  pre-feature `go.mod`, `go.sum`, and Ent-schema archive for parent
  `f21e3f7748eac55a4b16746b078e88801f11829c`.
- **SB-B/SB-C, source-carrier and Ent-generation control plane, exact seven files:** the static
  control plane is implemented in `backend/docker/Dockerfile.schema-maintenance`,
  `backend/docker/entrypoints/mail-calendar-schema.sh`,
  `backend/docker/schema-tool.lock.json`, `docker-compose.local.yml`, `backend/atlas.hcl`, this
  plan, and `backend/migrations/README.md`. It neither authors nor accepts SQL or generated
  source; `generate` uses the separate networkless non-database Compose boundary and accepts no
  arbitrary command.

SB-A/SB-B/SB-C are `STATIC_IMPLEMENTED` only. Docker, Ent generation, dependent compile, Atlas,
SQL, database and runtime evidence are `NOT_RUN`; the following exact SQL packets remain separately authored,
committed and independently reviewed. They contain no test changes.

1. **Immutable baseline, exact two files:** create
   `backend/migrations/atlas/mail-calendar-v1/20260823000000_mail_calendar_baseline.sql` and
   `backend/migrations/atlas/mail-calendar-v1/atlas.sum` from the exact parent above. Before
   authoring, the operator approves the parent SHA, tool lock, image digests and empty output
   directory; after authoring, Contract/Security/Lifecycle review the exact bytes and source
   manifest.
2. **Additive feature delta, exact two files:** create
   `backend/migrations/atlas/mail-calendar-v1/20260823000001_mail_calendar_core.sql` and modify
   only its sibling `atlas.sum`. The operator must approve the allowlist report and a destructive-
   SQL report showing zero forbidden statements before this packet can be accepted.
3. **Replay and cutover evidence:** after both SQL packets are accepted and executable
   permission is explicit, run checksum/lint, empty replay, realistic pre-feature replay,
   zero-diff comparison, ScheduledCall preflight and postconditions in the isolated profile.
   Shared or production apply additionally requires a fresh backup, exact database identity,
   independent database review and direct founder approval.

ScheduledCall backfill is a later separately reviewed forward-only artifact. It is not owned by
the Atlas schema delta. Founder question 18 remains unresolved: any non-deleted recurring row
blocks backfill and canonical-writer cutover; it is never skipped, coerced or assigned a guessed
timezone. This does not block producing or reviewing the additive schema migration itself.

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
2. Apply the independently reviewed bounded Atlas schema lineage before any backfill; the schema
   SQL contains no data rewrite or feature enablement.
3. Stop/guard legacy scheduled-call writers and the due worker; verify the guard before the
   first backfill write.
4. Insert events and attendees in bounded transactions; record exact affected counts.
5. Run target invariants, source/target hashes and event/call coherence postconditions.
6. Enable only bridge-owned writers/readers/worker after 100% non-deleted-call linkage passes.
7. Before production apply, rollback is removal of the unapplied artifact. After apply,
   recovery is forward-only: disable the feature flag, preserve rows for audit, and ship a
   reviewed corrective migration. Never delete user calendar data automatically.

## Required future evidence

- accepted `schema-maintenance` image/tool-lock identity and bounded Docker database target;
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
