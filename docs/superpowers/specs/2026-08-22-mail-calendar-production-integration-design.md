# Mail and Calendar Production Integration Design

**Status:** founder-approved for implementation on `codex/mail-module-20260813`

**Scope:** Ent schemas and production migrations, concrete repository adapters,
HTTP/NATS/worker composition, and the Mail-to-Calendar iTIP path. Frontend work,
the protected `pro-qa-ai/**` zone, and the separately owned `testframework/**`
implementation are outside this design.

> **LOCKED — 2026-08-23 testframework cutover.** Until the new
> `testframework` skeleton is merged into `master` **and this branch is updated
> from that master**, Mail work writes or modifies no backend, frontend, or
> Python tests. Implementation packets own production source only; existing
> tests are not edited; every executable claim is `NOT_RUN`; and exact-scope
> static source review is the current gate. A separate dedicated post-sync test
> plan backfills coverage against its then-current contracts.

## Goals

- Make the accepted Mail and Calendar domain services usable from production
  composition without weakening their owner, transaction, replay, sequence, or
  byte-boundary contracts.
- Keep persistence and transport concerns replaceable behind the existing domain
  ports.
- Establish one production migration mechanism instead of relying on development
  auto-migration.
- Preserve Docker-only executable verification after the locked framework
  cutover. Until then, exact-scope static source review is the only gate and
  all build, generation, migration, test, and runtime evidence remains
  `NOT_RUN`.

## Selected architecture

### Process ownership

The authenticated mailbox and canonical calendar domains run in the main Hub
backend. That process already owns the Ent client, authenticated Fiber API, NATS
connection, durable outbox worker, root cancellation context, and graceful
shutdown registry.

`backend/cmd/hub-mail` remains the narrow outbound delivery edge used for Hub
alerts. It does not become a second owner of mailbox state and does not receive a
second Ent client, migration runner, or Calendar repository. Provider-specific
mailbox engine access is supplied to the main backend through a `MailEngine`
adapter; the domain services continue to depend only on the existing port.

### Persistence boundaries

Calendar receives a new `EventEntRepository` implementing `EventRepository`.
Each `WithEventTransaction` call opens exactly one Ent transaction and creates a
transaction backend that implements only the optional capabilities needed by the
invoked service. Authorization resolution, row locks, receipt reservation,
domain mutations, optimistic-version checks, and outbox append stay in that same
transaction. The legacy Calendar aggregate repository remains read-only during
the compatibility window.

Mailbox persistence is split by responsibility around one shared Ent-backed
repository root:

- owner resolution;
- bounded sync claim/prepare/apply/resolve;
- due-account scan and fault rescheduling;
- later receipt-backed mailbox mutations.

The split keeps files small while reusing the same account lock order, lease/fence
validation, transaction helper, error mapping, and clock policy. Lease expiry is decided only by
`mailboxEntDatabaseWallClock(ctx, *ent.Tx)` in the owned Ent transaction. After a successful account
`FOR UPDATE`, Claim, Prepare, and Resolve each call it exactly once. Its transaction dialect `Query`
runs exactly `SELECT date_trunc('second', clock_timestamp())`, reads/scans exactly one row,
normalizes UTC, and validates through `mailboxEntValidTime`; caller/process time, a new transaction,
and the repository client are not expiry authorities. `CURRENT_TIMESTAMP` is rejected because it is
fixed at PostgreSQL transaction start and can be stale by the time the account lock is acquired.
Claim uses the fresh value for due, live-lease, schedule-marker, and returned non-claim authority,
and rejects a lease end not strictly later than it; `request.Now` is structural only. Prepare passes
the value to the unchanged `mailboxEntPrepareInput`; Resolve uses it for all lease gates and its
schedule marker. Repository code never performs provider I/O. A sync step claims durable state,
releases the transaction, performs one bounded `MailEngine` call, then opens a new transaction that
verifies the exact fence before applying results.
Handled `transient`, `rejected`, `invalid_cursor`, and `restart_required` Resolve failures all become
`degraded`. Source `43a` is state-machine-correct but `FINAL_QUALITY_PENDING_6Q2 / RUNTIME_NOT_RUN`;
it cannot be statically accepted until the fresh database-clock prerequisite has independent review.
The only 6V source manifest remains `backend/internal/mail/mailbox_ent_sync_claim.go` and only
`mailboxEntSyncClaimSnapshot` changes; there is no schema, public API, or migration change.
`CURSOR_INVALID` and `RESTART_REQUIRED` require nil rebuild fields (`sync_base_cursor` and
`sync_checkpoint`) and reconstruct empty bootstrap `Enumeration`. For `transient`/`rejected`, no
base/no checkpoint plus valid `sync_cursor` reconstructs `Incremental` with that cursor; the same
shape with nil cursor reconstructs empty bootstrap `Enumeration`; valid base/no checkpoint
reconstructs `CatchUp` with `Cursor=base`; valid base plus valid checkpoint reconstructs
`Enumeration` with that pair. Checkpoint without base or any invalid opaque cursor is the existing
content-free invalid error; other degraded codes retain the prior incremental-only guard.

One fenced transition increments `sync_attempts` exactly once or rejects overflow with zero partial
writes while atomically scheduling, recording a safe error, releasing the lease, and clearing
applied proof. A second Resolve cannot increment again; only successful Apply resets attempts.
Worker fault rescheduling owns only its separate failure increment. The production-only test lock
remains in force; exact-scope static source review is the present gate and executable evidence is
`NOT_RUN`. A forward schema prerequisite requires the complete proof CHECK to include
`sync_lease_applied_next_at < sync_lease_until`.
Apply writes `sync_lease_applied_next_at = NextAttemptAt` and checked exact-second
`sync_lease_until = NextAttemptAt + 1s`; this is the minimum same-lease replay grace,
and a new claim is permitted after lease expiry. A matching same-lease proof replays
without writes. A same-lease mismatch of hash, hash version, phase, next attempt, or
position returns the existing content-free invalid repository error with zero writes;
it is never overwritten or re-applied. Task 6C0 through source `52190` is
`PENDING_PROVENANCE_V2 / RUNTIME_NOT_RUN` and cannot be consumed. 6C0V2 in the implementation plan
is the single normative v2 SHA-256 wire registry: fixed-width big-endian TLV framing, outer domain
`hub.mail.sync.apply.result.v2`, format/proof version `2`, exact nested-record tag layout and
canonical collection ordering. `Upsert` tag `8` is optional source-authoritative `CalendarMethod` enum
string and tag `9` is the ordered `ChangedFields` enum-string collection. No alternate uvarint,
RFC3339, JSON, or implementation-order encoding is admissible, except the registered immutable
`EventEnvelope.FrozenJSON` bytes. A stored v1 proof is a mismatch and fails closed; no migration,
proof overwrite, or event-JSON authority is introduced.

`EngineMessageProjection.CalendarMethod` is source-authoritative only when a future provider adapter
derives it from bounded server-side MIME/calendar parsing, validates, copies, and budgets it. It is
never inferred from folders, provider event JSON, caller input, or a persisted row. Prepare must not
use the persisted `MailMessage.CalendarMethod` as policy input. Classification carries the current
projection method to `MailboxSyncPolicyMessageIntent`, and carries a concrete non-nil ordered policy
delta `ChangedFields` (empty for first-seen/no-delta). Canonical received events require exact
projection=intent=event method equality; live changed events require the exact ordered intent delta.
Deletion remains `[deleted]`; upserts permit only ordered `folder_ids`, `seen`, `answered`,
`flagged`, `junk`, `labels`, at most once each and never `deleted`. 6C1/6C2 later persist the intent
method to the existing column.

### Migration ownership

`backend/ent/schema/**` remains the model source. Generated Ent Go remains
uncommitted and is produced only inside the accepted task-owned Docker launcher.

Existing `backend/migrations` becomes the single versioned production migration
directory. Atlas is the deterministic diff, lint, checksum, and apply tool; it is
not a second schema authority. This founder decision supersedes the former
`PRODUCTION_ARTIFACT_BLOCKED_BY_FOUNDER_REVIEW` mechanism question in Phase 143;
the phase migration plan must record the decision before any executable migration
artifact is authored. The workflow is:

1. Generate Ent code in a disposable bounded container.
2. Produce an Atlas SQL diff from the checked-in Ent schemas against the current
   versioned migration state.
3. Inspect and commit the SQL plus `atlas.sum`.
4. Run lint and empty-database replay.
5. Run the Calendar legacy `ScheduledCall` preflight against a realistic snapshot.
6. Apply schema changes, then the separately reviewable forward-only backfill.
7. Verify postconditions before enabling canonical writers/workers.

`DEV_MODE=true` may continue to use `Schema.Create`; production never does.
Backfill emits no notification, NATS, or outbox side effect. Recovery after apply
is forward-only and preserves user data.

### HTTP and lifecycle composition

New mailbox and canonical-calendar handlers are registered under the authenticated
main `/api` group. Handlers translate transport DTOs and stable safe errors only;
they do not contain SQL, Ent queries, provider calls, or cross-domain branching.
Both canonical RSVP mutations are `POST /api/calendar/v2/events/{eventId}/rsvp` and
`POST /api/calendar/v2/events/{eventId}/instances/{recurrenceId}/rsvp`. Each returns only the
canonical `CalendarAttendee`; the caller invalidates and refetches its event/card rather than
receiving a full event or link state.
Legacy Calendar read endpoints remain available until the migration postconditions
authorize cutover. The current legacy Calendar handler registers only
`GET /api/calendar/events` and `GET /api/calendar/export.ics`; those reads remain
valid during compatibility. The cutover flag guards every existing
`ScheduledCall` writer and due worker, not the read handler: before cutover they
remain legacy-owned; after cutover they must delegate linked rows to the
event-before-call bridge and reject any direct linked-row schedule mutation. The
flag cannot enable until all non-deleted calls are linked and reconciliation reports
zero mismatch.

The mailbox sync worker receives the server root context. Startup is fail-closed:
when mailbox sync is enabled, missing database, engine, timer, or configuration is
a startup error. Shutdown is driven by root-context cancellation and the existing
cleanup stack; no detached goroutine owns process lifetime.

### Mail-to-Calendar flow

Cross-domain integration uses the existing shared envelope and durable outbox:

1. Mail projection persistence commits `mail.received.v1` with the safe locator,
   authentication classification, and optional `request|cancel|reply` iTIP method.
2. A Calendar consumer validates broker subject equals envelope subject, validates
   schema version and organization/account ownership, and deduplicates by envelope
   application `event_id`.
3. Before any body load or parse, the consumer requires authentication class
   `trusted_internal` or `dmarc_aligned`. `untrusted`, absent, or inconsistent
   classification is terminally quarantined and can never dispatch a Calendar
   service. Parsing alone never grants mutation authority.
4. A short claim transaction locks the receipt identity, replays any terminal
   outcome, treats a live lease as contention, or records `processing`, a bounded
   lease token, attempt count, and retry metadata. It then commits before loading
   bounded message/calendar content through the Mail read port.
5. Parsed `REPLY` invokes the accepted inbound-reply service. `REQUEST/CANCEL`
   invokes a separately persisted Phase 147 object-state decision that enforces
   organizer identity, sequence, canonical hash version, tombstones, and deferred
   occurrence rules.
6. A terminalization transaction locks the same receipt and verifies the exact
   lease token before locking calendar owner, master, children, object state, and
   outbox. Lease loss returns a typed retry/no-op before any mutation. Calendar
   mutation, object state, terminal receipt outcome, and Calendar outbox append
   then commit atomically. Retryable Mail-read/parser failures only reschedule the
   receipt in this second short transaction; terminal poison input records a safe
   quarantine/dead-letter outcome there.
7. Outbound Calendar invitation and RSVP envelopes use
   `calendar.invite.created.v1`, `calendar.invite.updated.v1`,
   `calendar.invite.cancelled.v1`, and `calendar.rsvp.changed.v1`. Mail consumes
   them through durable delivery/submission state; unknown submission outcomes are
   reconciled before retry.

No domain imports or queries the other domain's Ent tables. Shared identifiers and
envelopes cross the boundary; state ownership does not.

## Failure and security rules

- Every actor/account/event/organization relationship is resolved from durable
  state inside the transaction; client-supplied ownership is never trusted.
- Lock order is stable: organization/account or calendar owner, then master event,
  child rows, receipt/object state, and outbox as defined by the frozen contracts.
  The inbound consumer is the explicit exception needed to make leased work safe:
  its claim transaction locks only the receipt; its terminalization transaction
  locks the receipt first to prove the lease, then the Calendar rows in canonical
  order, and never holds either transaction over Mail I/O.
- Unique constraints and receipt rows are the final idempotency authority; in-memory
  deduplication is only an optimization.
- Optimistic-version or fence mismatch produces a typed retry/lease-lost outcome,
  never a last-write-wins overwrite.
- Raw MIME, calendar body, credentials, provider errors, and engine identifiers are
  absent from receipts, logs, public errors, and NATS envelopes.
- Consumer poison messages are terminally quarantined/dead-lettered using safe
  codes and bounded retry metadata.
- Worker concurrency, page sizes, retry attempts, and wall-clock slots retain the
  existing hard caps.

## Delivery order

1. Migration mechanism and schema diff authority.
2. Calendar base Ent transaction adapter, followed by RSVP/update/cancel/inbound
   reply capability files.
3. Mail owner and sync/worker Ent adapters: required prerequisite chain 6W -> 6Q2 -> 6C0V2 ->
   6C1 -> 6C2. 6Q source `6477` is superseded/pending correction; 6V source `43a` remains
   state-machine-correct but its final quality waits for 6Q2. 6C0 through `52190` remains pending
   provenance v2. 6Q2 and 6C0V2 share Prepare and are strictly sequential.
4. Main-server composition and authenticated HTTP surfaces.
5. Outbox-backed Mail-to-Calendar consumer and Calendar-to-Mail delivery consumer.
6. Compatibility cutover and ScheduledCall backfill only after migration evidence.
7. Docker generation, build, focused verification, migration replay, and runtime
   smoke through the accepted launcher.

## Acceptance

- Concrete adapters satisfy their existing domain interfaces without expanding the
  public ABI.
- 6Q2 and 6C0V2 receive independent exact-scope static source review only; until
  the locked framework cutover every test, build, generation, format, lint,
  migration, SQL/DB, Docker, and runtime claim remains `NOT_RUN`.
- No database transaction spans a `MailEngine` or NATS network call.
- Domain writes and their outbox/receipt rows are atomic.
- Legacy Calendar remains read-compatible until explicit cutover.
- After the locked framework cutover, versioned migrations replay from empty
  state and from the accepted pre-feature snapshot in Docker.
- After that cutover, Docker build and applicable regression jobs pass before
  merge; until then this design has static-source-review evidence only and is
  not merge-ready.
