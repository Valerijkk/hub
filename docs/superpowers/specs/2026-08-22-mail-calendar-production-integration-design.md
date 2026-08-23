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
and rejects a lease end not strictly later than it. After the lock and fresh clock but before any
lease mutation, Claim rejects `SyncAttempts < 0 || SyncAttempts >= mailboxEntMaxSyncAttempts` through
the existing content-free invalid error with zero writes and no engine I/O. `request.Now` is
structural only. No-active Claim retains `sync_lease_epoch` (pristine first claim is `1`; a released
lease is already advanced), live same-token renewal retains it, and expired retained lease increments
it as before. Prepare passes the value to the unchanged `mailboxEntPrepareInput`; Resolve uses it for
all lease gates and `SyncScheduleUpdatedAt`, but persists and returns exact `request.NextAttemptAt`
without clamp or rebase. A next attempt at or before the DB clock is intentional immediate due; the
DB clock owns only lease/due decisions and the schedule marker. Repository code never performs
provider I/O. A sync step claims durable state, releases the transaction, performs one bounded
`MailEngine` call, then opens a new transaction that verifies the exact fence before applying results.
Handled `transient`, `rejected`, `invalid_cursor`, and `restart_required` Resolve failures all become
`degraded`. Tasks 6Q2/6V are `STATIC_ACCEPTED / QUALITY_APPROVED / RUNTIME_NOT_RUN` through source
commits `57b815a1`, `43a6546d`, and `89c47581`; `43a6546d` remains state-machine-correct for decoder
reconstruction. The only source manifest remains `backend/internal/mail/mailbox_ent_sync_claim.go`;
it changes the decoder plus Claim/Resolve retry-cap and lease-generation behavior only, with no
schema, public API, or migration change. Acceptance is static only and does not establish compile,
generation, database, Docker, test, or runtime evidence.
`CURSOR_INVALID` and `RESTART_REQUIRED` require nil rebuild fields (`sync_base_cursor` and
`sync_checkpoint`) and reconstruct empty bootstrap `Enumeration`. For `transient`/`rejected`, no
base/no checkpoint plus valid `sync_cursor` reconstructs `Incremental` with that cursor; the same
shape with nil cursor reconstructs empty bootstrap `Enumeration`; valid base/no checkpoint
reconstructs `CatchUp` with `Cursor=base`; valid base plus valid checkpoint reconstructs
`Enumeration` with that pair. Checkpoint without base or any invalid opaque cursor is the existing
content-free invalid error; other degraded codes retain the prior incremental-only guard.

After the complete live fence, Resolve computes `mailboxEntNextEpoch(account.SyncLeaseEpoch)` and
rejects overflow through the existing content-free invalid error with zero writes. Its one atomic
update calls `SetSyncLeaseEpoch(next)` together with exactly one `sync_attempts` increment, exact
`request.NextAttemptAt` scheduling without clamp/rebase, safe error, lease release, and applied-proof
clear; this prevents same-`LeaseID` ABA. A second Resolve cannot increment again; only successful
Apply resets attempts. Worker fault rescheduling owns only its separate failure increment. The
production-only test lock remains in force; exact-scope static source review is the present gate and
executable evidence is `NOT_RUN`. A forward schema prerequisite requires the complete proof CHECK to include
`sync_lease_applied_next_at < sync_lease_until`.
Apply writes `sync_lease_applied_next_at = NextAttemptAt` and checked exact-second
`sync_lease_until = NextAttemptAt + 1s`; this is the minimum same-lease replay grace,
and a new claim is permitted after lease expiry. A matching same-lease proof replays
without writes. A same-lease mismatch of hash, hash version, phase, next attempt, or
position returns the existing content-free invalid repository error with zero writes;
it is never overwritten or re-applied. Prior v1 source range `8935..52190` is
`SUPERSEDED_BY_6C0V2 / NOT_RUNTIME_ACCEPTED` and cannot be consumed. 6C0V2 is
`STATIC_ACCEPTED / SPEC_PASS / QUALITY_APPROVED / RUNTIME_NOT_RUN` through source commits
`41851638` and `805ac226`. It is the single normative v2 SHA-256 wire registry: fixed-width
big-endian TLV framing, outer domain `hub.mail.sync.apply.result.v2`, format/proof version `2`, exact
nested-record tag layout and canonical collection ordering. `Upsert` tag `8` is optional
source-authoritative `CalendarMethod` enum string and tag `9` is the ordered `ChangedFields`
enum-string collection. No alternate uvarint, RFC3339, JSON, or implementation-order encoding is
admissible, except the registered immutable `EventEnvelope.FrozenJSON` bytes. A stored v1 proof is a
mismatch and fails closed; no migration, proof overwrite, or event-JSON authority is introduced.
This acceptance is static/spec-only and does not establish compile, generation, database, Docker,
test, or runtime evidence.

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
Canonical `PATCH` and `DELETE /api/calendar/v2/events/{eventId}` return only
`CalendarEventMutationResult`, mirroring the accepted `CalendarEventUpdateResult` fields
`event_id`, `version`, `sequence`, `status`, and `changed_fields`. The handler does not hydrate
or fabricate canonical event detail; its caller invalidates and refetches that detail.
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

### Task 10 P0 prerequisite contract

This prerequisite track is `STATIC_SOURCE_ONLY / RUNTIME_NOT_RUN`, not runtime readiness.
P0 contract is independently finally accepted C/I/M `0/0/0` at
`0d4fa053d569631234c5a7f902341bb6a73a7d46`, with pairing fix
`5a253d166f83ce37e695566f1396ee16fe00b224`; this is static/docs-only provenance.
P1 `CalendarInboundMessageReceipt` schema is independently approved C/I/M `0/0/0` at
`4b7f322b629dd269cac7e05ba61f7ca03e45d96c`. P2 makes `MailMessage`
`authentication_class` non-null with default `untrusted` and `calendar_part_locator_hash`
nullable at exactly 32 bytes; existing rows backfill as `untrusted`/null with no reparse or
historical re-emission. The P2 durable schema is approved C/I/M `0/0/0` at
`fc5dfe195b89963601e3f365e593469f73bed94f`; its projection carrier is approved at
`1ebbb129d62aeca695a2b028822101af9c77687f`; and event/durable/V2/V3 proof is approved at
`a61edbb086df8535349429438155fad015b29615`. The engine projection supplies only the class
and domain-separated locator hash, never raw locator, provider ID, or body.

`mail.received.v1` keeps `engine_message_id` solely for legacy v1 compatibility. A new Calendar
consumer ignores it; therefore it is incorrect to claim that no provider IDs exist anywhere in
NATS. The new fields are required: authentication enum `trusted_internal|dmarc_aligned|untrusted`,
and method plus locator pair (both null for non-candidate; otherwise a permitted method and a
64-lowercase-hex locator). Public APIs/logs receive no authentication evidence; the envelope carries
only the enum.

6C0V2 is retained unchanged for matching stored-v2 replay. 6C0V3 is the new-request proof:
domain `hub.mail.sync.apply.result.v3`, version 3, tags 1..9 unchanged, tag 10 authentication
class enum string, and tag 11 optional raw 32-byte locator hash using the registered presence
encoding. v1 remains fail-closed; a v2 proof is neither overwritten nor reinterpreted as v3.
Canonical validation proves projection/carrier equality and method/locator pairing.

P3 remains future: no Calendar consumer, repository, service, server wiring, NATS subscription,
or worker is accepted. Generated Ent artifacts are absent/uncommitted; migration, SQL,
compile/build, test, lint/gofmt, parser/validator, Docker, and runtime remain `NOT_RUN`.

### P3-A sealed inbound source freeze

P3 is contract-frozen only, `UNIMPLEMENTED / PENDING / RUNTIME_NOT_RUN`. Its owner is
`backend/internal/mail` (`package mail`): Calendar may use exported sealed Mail port/types and Mail
never imports Calendar. `CalendarPartReadRequest`/Result are opaque structs with exported validated
constructor/accessors. Calendar receives only Hub org/mailbox/message UUIDs, `[32]byte` locator hash,
and exact `maxBytes=1..1_048_576`, returning only owner ID, normalized sender, method, authentication
class, and defensive-copy iTIP. The source owns five seconds and whole-result `payload_too_large`.
`MailEngine.ReadCalendarPart` is Go-exported only for interface implementation and engine-private:
only Mail source calls it with engine account/message IDs, expected engine version/hash/method/max bytes.
The adapter canonicalizes private MIME locators, recomputes hashes, and requires exactly one match;
raw locator/version never leaves Mail. The V1 formula and snapshot/recheck are normative in CONTRACT §2.6.

Historic 147-02D `GetMessage`/full-body/MIME parsing and legacy event `engine_message_id` wording is
superseded for Calendar ingestion only; none is a fallback. P3-B factors existing REPLY application
into one EventTransaction under receipt-first lock/proof, preserving reply receipt/outbox behavior;
P3-C/D await the respective Mail/Calendar generation. P3-E is Mail-owned consumer code: it validates
compatibility `mail.received.v1`, ignores/drops `engine_message_id`, and sends only sanitized
`CalendarInboundDelivery` to an injected Calendar-implemented sink; it does not make Calendar subscribe
to the full envelope. The frozen consumer has EVENTS `mail.>`, durable `calendar-mail-inbound-v1`, exact
filter `mail.received.v1`, explicit ACK, AckWait 90s, hard timeout 45s, MaxAckPending 16, fetch batch
8, concurrency 4, and MaxDeliver 20. Semantic durable terminal/retry/defer dispositions ACK. Only a
no-durable-disposition failure uses `NakWithDelay(65s)`; exhaustion retains the message and raises one
content-free degraded/manual-replay signal, never ACK/delete/drop.

No domain imports or queries the other domain's Ent tables. The single explicit exception is
Calendar importing the exported sealed `package mail` port/types; Mail never imports Calendar, and
neither domain receives or queries the other's Ent client. Shared identifiers and sanitized envelopes
cross the boundary; state ownership does not.

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
- Raw MIME, calendar body, credentials, provider errors, authentication evidence, and newly
  introduced provider identifiers are absent from receipts, logs, public errors, and new
  envelope fields. Legacy v1 `engine_message_id` is explicitly retained as an ignored
  compatibility field.
- Consumer poison messages are terminally quarantined/dead-lettered using safe
  codes and bounded retry metadata.
- Worker concurrency, page sizes, retry attempts, and wall-clock slots retain the
  existing hard caps.

## Sync projection persistence boundary

- One `ListFolders` result is a complete bounded account snapshot. 6C1 validates
  its hierarchy and live uniqueness before any write, updates or safely revives
  provider-owned rows, and soft-deletes only absent active provider rows. It never
  overwrites a folder carrying an in-flight control receipt.
- One message upsert is a complete provider projection. The canonical non-nil
  `ChangedFields` sequence is 6C0V2-validated event provenance; 6C1 never
  derives or compares it to Ent state and it is not a selective-write mask.
  Metadata, thread, provider version, origin,
  Calendar method, attachments, and sync bookkeeping therefore remain durable
  even when that sequence is empty.
- A provider tombstone soft-deletes the live message and removes its folder
  memberships while retaining labels, attachments, thread/projection history,
  Calendar linkage, and received-event provenance. Reappearance revives the same
  account-scoped message identity.
- 6C1 alone owns exact folder counters and live-message thread aggregates. Live
  means a non-deleted message and, for folder counts, an active non-deleted
  folder joined by a physical membership row; unread additionally means
  `is_seen=false`.
  Effective thread time is the later provider timestamp, otherwise the persisted
  message `updated_at` for an unchanged row or the single transaction time for a
  new/changed row; the lexicographically greatest message UUID wins the final tie.
  Zero-live-message threads are soft-deleted while retaining their last
  schema-valid non-empty aggregate values.
- 6C1 consumes only a passed `*ent.Tx` and exact-second UTC transaction time.
  6C2 remains the sole owner of the account lock/fence, database-clock read,
  transaction lifecycle, position/schedule/proof, and outbox writes.

## Delivery order

1. Migration mechanism and schema diff authority.
2. Calendar base Ent transaction adapter, followed by RSVP/update/cancel/inbound
   reply capability files.
3. Mail owner and sync/worker Ent adapters: 6Q2/6V are `STATIC_ACCEPTED / QUALITY_APPROVED /
   RUNTIME_NOT_RUN` through `57b815a1`, `43a6546d`, and `89c47581`; 6C0V2 is
   `STATIC_ACCEPTED / SPEC_PASS / QUALITY_APPROVED / RUNTIME_NOT_RUN` through `41851638` and
   `805ac226`. Prior v1 `8935..52190` is superseded by v2 and not runtime accepted. 6Q2 and 6C0V2
   share Prepare but are complete; the 6C1 persistence contract is frozen and independently
   reviewed PASS in the implementation plan. Its source is split into folder, membership,
   attachment, message/thread/tombstone, aggregate, and orchestrator atoms; folder persistence is
   statically accepted through `749bfad8`, message-link persistence through `559dade8`, and
   attachment persistence through `bfa07ae7`. Message/thread/tombstone persistence is statically
   accepted through `8d2bbb3c` and `5f28ad37`, aggregates through `0bb0a7e8`, and the 6C1
   orchestrator through `7ad57e85`. The fenced 6C2 transaction/proof/outbox composition is
   statically accepted through `e20fa860`, `fc25ca93`, `f7996092`, and `d6184d7a`.
   Executable evidence remains `NOT_RUN`; Task 7B and Task 7C now have only the static acceptance
   recorded below, while Task 7A remains blocked.
4. Main-server composition and authenticated HTTP surfaces: Task 7B is statically accepted through
   `f02c3be1`, `ebfe750d`, and `af92452d` for authenticated `GET /api/mail/account` only; Task 7C
   is independently statically accepted through `cb1c59e8`, `6a9276ba`, `fbeb7419`, `bc66cdcb`, and
   `0eacbc2d` for the six additive canonical mutation routes; and Task 8A route composition is
   statically accepted through `1627b46d`. Its private UTC/UUID/iTIP adapters register legacy
   Calendar GET, canonical mutations, and the Mail account route once. Task 7A stays blocked on
   Phase 145 founder auth/gateway gates and inaccessible private engine interfaces; Task 8B provider
   and worker lifecycle remains pending/blocked. No item in this paragraph is runtime evidence.
5. P0-P2 Mail-to-Calendar prerequisite sources are statically accepted; P3 remains the later
   sealed-source/transaction/repository/worker/NATS sequence (P3-A through P3-E), followed by
   the separate Calendar-to-Mail delivery consumer.
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
