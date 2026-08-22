# Mail and Calendar Production Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist and compose the accepted Mail and Calendar domain services, then connect them through durable iTIP event flows.

**Architecture:** The main Hub backend owns authenticated HTTP, Ent, NATS, outbox, and worker lifecycle. Ent adapters implement the existing domain ports without provider I/O; `hub-mail` remains an outbound-alert edge. Versioned SQL under `backend/migrations` is generated and verified by Atlas in the accepted Docker launcher.

**Tech Stack:** Go 1.25, Ent, PostgreSQL, Atlas versioned migrations, Fiber v3, NATS, Hub durable outbox.

**Spec:** `docs/superpowers/specs/2026-08-22-mail-calendar-production-integration-design.md`

## Global Constraints

- Root orchestrates only; every atomic source packet is authored with explicit `gpt-5.6-terra` / `xhigh`, no fallback.
- All tests, build, generation, formatting, lint, migration, and runtime commands run only through the accepted task-owned Docker launcher.
- Until that launcher is ready, executable evidence is `NOT_RUN`; static review must not be described as PASS runtime evidence.
- Do not read or edit `pro-qa-ai/**` or the separately owned `testframework/**`.
- Generated `backend/ent/**` code remains uncommitted; only `backend/ent/schema/**`, versioned SQL, and checksums are source artifacts.
- Repository transactions never span MailEngine, HTTP, NATS, or other network I/O.
- Use exact-path staging and one independently reviewable commit per task.
- No AI attribution or `Co-authored-by` trailers.

---

### Task 1: Atlas versioned-migration control plane

**Files:**
- Create: `backend/atlas.hcl`
- Create: `backend/migrations/README.md`
- Modify: `docs/stack_and_tools.md`
- Modify: `.planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: Ent schemas at `ent://ent/schema`, `DATABASE_URL`, and a task-owned `ATLAS_DEV_URL`.
- Produces: one `local` Atlas environment with `file://migrations` as the only versioned directory.

- [ ] **Step 1: Add the exact Atlas environment**

```hcl
variable database_url {
  type    = string
  default = getenv(DATABASE_URL)
}

variable dev_url {
  type    = string
  default = getenv(ATLAS_DEV_URL)
}

env local {
  url = var.database_url
  dev = var.dev_url
  schema {
    src = ent://ent/schema
  }
  migration {
    dir = file://migrations
  }
}
```

Both variables fail closed when absent in the launcher. Do not embed credentials or a host Docker URL.

- [ ] **Step 2: Document the production workflow**

`backend/migrations/README.md` must state: Ent is desired state; Atlas produces reviewed SQL plus `atlas.sum`; production applies only checked-in SQL; `Schema.Create` is development-only; backfill is a separate forward-only artifact; direct host commands are forbidden.

- [ ] **Step 3: Record static-only status**

Update stack/Phase/STATE docs with `MIGRATION_CONTROL_PLANE_STATIC`, launcher `NOT_READY`, and every executable field `NOT_RUN`.

- [ ] **Step 4: Static review**

Inspect exact files, run `git diff --check -- <exact paths>`, and obtain Contract/Security/Lifecycle precommit reviews. Do not invoke Atlas yet.

- [ ] **Step 5: Commit**

```text
chore(db): define Atlas migration control plane
```

### Task 2: Phase 147 inbound and reply-delivery schemas

**Files:**
- Create: `backend/ent/schema/calendarinboundobjectstate.go`
- Create: `backend/ent/schema/calendarinboundmessagereceipt.go`
- Create: `backend/ent/schema/calendarreplydelivery.go`
- Modify: `backend/ent/schema/calendarevent.go`
- Modify: `backend/ent/schema/calendarattendee.go`
- Modify: `backend/ent/schema/organization.go`
- Modify: `backend/ent/schema/user.go`
- Test: `backend/tests/calendar/calendar_mail_bridge_schema_test.go`
- Modify: `.planning/phases/147-mail-calendar-reliability/CONTRACT.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: exact field/state/retention rules in Phase 147 sections 2.1, 2.2, and 3.1.
- Produces: durable ordering state, leased inbound receipt, and immutable outward reply intent.

- [ ] **Step 1: Author the isolated schema test**

The test imports `backend/ent/schema` only and asserts exact fields, enum values, immutable identities, checks, edges, and indexes. It must include the two partial unique indexes for nullable/non-null `recurrence_id`, unique message plus locator hash, and unique attendee plus event sequence plus attendee version.

- [ ] **Step 2: Add `CalendarInboundObjectState`**

Use required UUID organization/owner identity, bounded UID, nullable recurrence key, normalized organizer email max 320, non-negative sequence, method `request|cancel`, positive hash version, exact 32-byte hash, state `applied|tombstone|deferred_missing_master`, optional event, optional expiry, and timestamps. Encode the state/expiry alternatives as a PostgreSQL CHECK.

- [ ] **Step 3: Add `CalendarInboundMessageReceipt`**

Persist message UUID and 32-byte locator hash as scalar Mail identities, application envelope UUID, bounded lease/retry fields, safe state/error enums, optional Calendar event, and timestamps. Do not add a Mail Ent edge or raw part/body/provider field.

- [ ] **Step 4: Add `CalendarReplyDelivery`**

Use required Calendar event/attendee edges, immutable event and attendee versions, optional recurrence key, RSVP `accepted|declined|tentative`, delivery state, optional scalar Mail submission UUID, bounded retry metadata, and timestamps.

- [ ] **Step 5: Add only inverse edges required for Calendar ownership**

Organization/User/Event/Attendee inverse edges must not introduce cascades that can erase receipts, tombstones, or delivery audit history. Mail submission/message references remain scalar cross-domain identities.

- [ ] **Step 6: Static review and commit**

Obtain Contract/Security/Lifecycle review, record generated/migration/runtime `NOT_RUN`, then exact-path commit:

```text
feat(calendar): add mail bridge persistence schemas
```

### Task 3: Calendar Ent transaction adapter

**Files:**
- Create: `backend/internal/calendar/event_ent_repository.go`
- Create: `backend/internal/calendar/event_ent_transaction.go`
- Create: `backend/internal/calendar/event_ent_create.go`
- Test: `backend/tests/calendar/event_ent_repository_contract_test.go`
- Modify: `.planning/phases/143-calendar-rsvp-core/PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: `EventRepository`, mandatory `EventTransactionBackend`, and `RunEventTransaction`.
- Produces: `NewEventEntRepository(*ent.Client) *EventEntRepository` and compile-time `var _ EventRepository = (*EventEntRepository)(nil)`.

- [ ] **Step 1: Freeze the transaction contract in an isolated test**

Cover nil client/callback, begin failure, callback error rollback, commit failure, transaction lifetime closure, one transaction per call, and no callback reuse.

- [ ] **Step 2: Implement transaction ownership**

`WithEventTransaction` begins `client.Tx(ctx)`, creates one transaction backend, calls `RunEventTransaction`, rolls back on every error/panic path, and commits only after the callback succeeds. Never expose `*ent.Tx`.

- [ ] **Step 3: Implement the mandatory create backend**

Resolve actor organization and organizer/attendee membership from locked durable rows; reserve/complete the mutation receipt; persist event, attendee, exception, and outbox rows in the same transaction. Map unique/version conflicts to existing typed domain errors without leaking SQL text.

- [ ] **Step 4: Static review and commit**

```text
feat(calendar): add base Ent event repository
```

Executable compile and database evidence remain `NOT_RUN` until Task 9.

### Task 4: Calendar update, RSVP, cancellation, and inbound-REPLY capabilities

**Files:**
- Create: `backend/internal/calendar/event_ent_update.go`
- Create: `backend/internal/calendar/event_ent_rsvp.go`
- Create: `backend/internal/calendar/event_ent_inbound_reply.go`
- Test: `backend/tests/calendar/event_ent_capabilities_contract_test.go`
- Modify: `.planning/phases/143-calendar-rsvp-core/PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: optional transaction backend interfaces already declared in `event_repository.go`.
- Produces: one transaction backend satisfying update, series/occurrence RSVP, occurrence cancel, and inbound reply without expanding `EventRepository`.

- [ ] **Step 1: Implement authorization resolvers**

Every resolver locks owner/master before children, masks foreign identity as not-found/unauthorized per the frozen contract, returns bounded deep-copied snapshots, and never accepts caller-supplied current state.

- [ ] **Step 2: Implement optimistic persistence**

Every write compares exact expected version/fence, advances sequence/version once, persists receipt/outbox atomically, and returns the immutable echo expected by the service. Replay returns the stored terminal outcome.

- [ ] **Step 3: Implement inbound REPLY persistence**

Resolve owner/account relationship, sender-bound attendee, series/occurrence state and canonical receipt. Enforce lower/equal/newer sequence behavior and quarantine conflicts without raw content persistence.

- [ ] **Step 4: Static review and commit**

```text
feat(calendar): implement Ent event capabilities
```

### Task 4B: Freeze and implement the event-to-call repository bridge

**Files:**
- Create: `backend/internal/calendar/event_call_repository.go`
- Create: `backend/internal/calendar/event_ent_call.go`
- Modify: `backend/internal/calendar/event_call_policy.go`
- Test: `backend/tests/calendar/event_call_repository_contract_test.go`
- Modify: `backend/tests/calendar/event_call_policy_test.go`
- Modify: `docs/superpowers/plans/2026-08-22-mail-calendar-production-integration.md`
- Modify: `.planning/phases/143-calendar-rsvp-core/PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: `EventCallProjectionInput`, `EventCallProjection`, `EventCallDueInput`, and `EventCallDueDecision`.
- Produces: a separate `EventCallRepository`; it does not expand `EventRepository`.

- [ ] **Step 1: Freeze the exact port**

```go
type EventCallRepository interface {
    WithEventCallTransaction(context.Context, func(*EventCallTransaction) error) error
}

type EventCallTransactionBackend interface {
    ResolveEventCallProjection(context.Context, EventCallProjectionRequest) (EventCallProjectionInput, error)
    PersistEventCallProjection(context.Context, EventCallProjectionRequest, EventCallProjection) (EventCallProjectionSnapshot, error)
    ResolveEventCallDue(context.Context, EventCallDueRequest) (EventCallDueInput, error)
    PersistEventCallDue(context.Context, EventCallDueRequest, EventCallDueDecision) (EventCallDueSnapshot, error)
}
```

`EventCallProjectionRequest` contains only scheduled-call ID, action, boundary, expected event version, and expected call version. `EventCallDueRequest` contains only scheduled-call ID, exact UTC `Now`, envelope input, expected event version, and expected call version. Constructors validate non-nil IDs, positive versions, exact-second UTC, and action/boundary combinations.

`EventCallProjectionSnapshot.DeletedAt` is nullable and policy-owned: only `delete` receives an isolated exact-second UTC copy of the validated boundary instant. All other dispositions retain `nil`; the Ent delete writes that timestamp with `is_active=false` and never calls a clock.

- [ ] **Step 2: Implement transaction lifetime**

Mirror the guarded Calendar transaction pattern: one Ent transaction, lifetime-closed wrapper, rollback on error/panic, commit only after callback success, compile-time interface assertion.

- [ ] **Step 3: Implement projection and due persistence**

Resolve all `EventCallProjectionInput` and `EventCallDueInput` state from locked event-before-call rows. Persist only the returned decision snapshot, compare both expected versions, reject direct linked-row divergence, and append due outbox atomically. Do not enable cutover.

- [ ] **Step 4: Static review and commit**

```text
feat(calendar): add event call Ent bridge
```

### Task 5: Mail Ent repository root and owner resolution

**Files:**
- Create: `backend/internal/mail/mailbox_ent_repository.go`
- Create: `backend/internal/mail/mailbox_ent_owner.go`
- Test: `backend/tests/mail/mailbox_ent_owner_contract_test.go`
- Modify: `.planning/phases/144-mailbox-domain/PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: `MailboxOwnerRepository` and accepted MailAccount/MailIdentity schema.
- Produces: `NewMailboxEntRepository(*ent.Client, MailboxRepositoryClock) (*MailboxEntRepository, error)`.

- [ ] **Step 1: Define shared repository primitives**

Keep Ent client and injected UTC clock private. Centralize transaction begin/rollback, account lock order, safe constraint mapping, checked time math, and deep-copy helpers.

- [ ] **Step 2: Implement owner resolution**

Resolve actor to one active owner-bound account/identity inside durable state. Foreign, inactive, deleting, or ambiguous rows fail closed with existing typed safe errors.

- [ ] **Step 3: Static review and commit**

```text
feat(mail): add Ent mailbox owner repository
```

### Task 6S: Persist Mail sync lease generation and due-scan schema prerequisite

**Files:**
- Modify: `backend/ent/schema/mailaccount.go`
- Test: `backend/tests/mail/mail_account_identity_schema_test.go`
- Modify: `.planning/phases/144-mailbox-domain/PLAN.md`
- Modify: `.planning/STATE.md`

**Status:** `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `c8274bd2`.

This prerequisite persists `sync_lease_epoch` immediately after `sync_epoch` as required
`Int64`, default `1`, minimum `1`, and appends
`mail_accounts_sync_due_idx` over `(next_sync_at, id)` where exactly
`status = 'active' AND next_sync_at IS NOT NULL`. `MailboxSyncFence` and
`MailboxSyncWorkerDueAccount` will carry the generation: a never-claimed due account starts at
`1`, and a later adapter increments it only when distinct durable lease ownership replaces the
prior lease. The existing non-negative CHECK retains every prior clause and additionally requires
`sync_lease_epoch >= 1`.

Task 6 also depends on Task 6P and Task 6T. This packet does not create a sync-page receipt; the
inline-proof and exact schedule-clock prerequisites are frozen below. Independent static review found no issues in
this narrow manifest. Launcher `NOT_READY`; Ent generation, Atlas diff/migration, compile, SQL/DB,
Docker, formatting/lint, and runtime are all `NOT_RUN`. No migration or cutover authority follows.

### Task 6P: Persist bounded Mail sync-lease snapshot and apply-proof prerequisite

**Files:**
- Modify: `backend/ent/schema/mailaccount.go`
- Test: `backend/tests/mail/mail_account_identity_schema_test.go`
- Modify: `.planning/phases/144-mailbox-domain/PLAN.md`
- Modify: `.planning/STATE.md`

**Status:** `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `7309c52a`.

This prerequisite adds one bounded inline proof to the active `MailAccount` sync lease rather
than an unbounded sync receipt table. It has no retention-cleanup dependency, stores no
message/calendar/provider body, raw error, or JSON payload, and keeps the original claimed
snapshot after the account cursor advances so the same lease can safely claim/replay its apply
result.

Immediately after `sync_lease_token` and before `mutation_lease_token`, persist nullable
non-negative snapshot `sync_lease_sync_epoch` and `sync_lease_mutation_epoch`, nullable phase
`incremental|enumeration|catch_up`, nullable bounded opaque cursor/base/checkpoint strings
(`1..4096`), and nullable proof hash-version/32-byte SHA-256 hash/applied phase/applied-next-at/
applied-at fields. All fields have no defaults.

`mail_accounts_sync_lease_snapshot` is a portable fail-closed `COALESCE` CHECK: no lease token
requires every snapshot value to be null; an active token plus expiry requires both epoch fences
and the exact `validMailboxSyncFence` shapes. Incremental and catch-up require only a cursor;
enumeration permits either initial no-base/no-checkpoint or resumed paired base/checkpoint.
`mail_accounts_sync_lease_apply_proof` is a separate portable fail-closed `COALESCE` CHECK:
all proof fields are null, or an active token plus expiry carries every proof field, positive hash
version, 32-byte hash, `applied_at < sync_lease_until`, `applied_next_at >= applied_at`, and only
the transitions incremental→incremental, catch_up→catch_up|incremental, and
enumeration→enumeration|catch_up.

Independent static review found no issues in the bounded snapshot/proof checks. Task 6 starts only
after Task 6S, Task 6P, and Task 6T. The launcher is `NOT_READY`; executable RED/GREEN,
Ent generation, Atlas diff/migration, compile, SQL/DB, Docker, gofmt, lint, and runtime are all
`NOT_RUN`. This source candidate provides no migration or cutover authority.

### Task 6T: Persist exact-second Mail sync schedule clock prerequisite

**Files:**
- Modify: `backend/ent/schema/mailaccount.go`
- Test: `backend/tests/mail/mail_account_identity_schema_test.go`
- Modify: `.planning/phases/144-mailbox-domain/PLAN.md`
- Modify: `.planning/STATE.md`

**Status:** `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `12f59cc8`.

This prerequisite adds required `sync_schedule_updated_at` immediately after `next_sync_at`.
Its creation default is a private UTC clock truncated to an exact second; it deliberately has no
`UpdateDefault`. The fail-closed
`mail_accounts_sync_schedule_updated_at_second_precision` CHECK requires exact PostgreSQL
second precision. `MailboxSyncWorkerDueAccount.UpdatedAt` maps from this dedicated marker rather
than generic Ent `updated_at`, whose default retains sub-second precision.

Every Task 6 adapter mutation that changes sync eligibility, `next_sync_at`, sync/lease/mutation
epochs, lease snapshot/proof, attempts, claim, apply, resolve, or fault scheduling must set
`sync_schedule_updated_at` in the same transaction from the repository's exact-second UTC clock.
Existing-row backfill and the final `NOT NULL` migration are owned by Task 9 and require separate
review. The launcher is `NOT_READY`; executable RED/GREEN/tests, Ent generation, Atlas
diff/migration, compile, SQL/DB, Docker, gofmt, lint, and runtime are all `NOT_RUN`. This source
prerequisite grants no migration or cutover authority. Independent static review approved the
exact five-file source boundary with C/I/M `0/0/0`.

### Task 6: Mail sync and worker Ent adapters

**Files:**
- Create: `backend/internal/mail/mailbox_ent_sync_claim.go`
- Create: `backend/internal/mail/mailbox_ent_sync_prepare.go`
- Create: `backend/internal/mail/mailbox_ent_sync_apply.go`
- Create: `backend/internal/mail/mailbox_ent_worker.go`
- Test: `backend/tests/mail/mailbox_ent_sync_coordination_contract_test.go`
- Test: `backend/tests/mail/mailbox_ent_sync_prepare_contract_test.go`
- Test: `backend/tests/mail/mailbox_ent_sync_apply_contract_test.go`
- Modify: `.planning/phases/144-mailbox-domain/PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: `MailboxSyncRepository`, `MailboxSyncWorkerRepository`, accepted sync fence/page policy, and worker caps.
- Produces: the same `MailboxEntRepository` satisfying both interfaces.

**Dependency:** Task 6 starts only after Task 6S, Task 6P, and Task 6T persist and independently
statically review the durable sync-lease generation, active due-scan index, inline apply proof,
and exact-second schedule CAS marker.

**Atomic decomposition:** Task 6A owns claim/resolve plus due-scan/fault coordination; Task 6B owns
prepare-only projection reads and policy-input construction; Task 6C owns deterministic atomic
projection/count/outbox apply and bounded same-lease result-hash replay. No packet spans provider
I/O and a repository transaction.

- [x] **Step 1: Implement Task 6A claim/resolve and worker coordination**

Claim/resolve use one repository clock sample inside each transaction, enforce the complete
lease/sync/mutation/phase/position fence, and update the exact schedule CAS marker with every
schedule mutation. Due scan is bounded, ordered `(next_sync_at, id)`, active/eligible and
`FOR UPDATE SKIP LOCKED`; fault reschedule uses the full captured schedule/epoch/attempt CAS and
never clears leases or changes projection state.

Task 6A is `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `24d2873d` after independent
immutable review C/I/M `0/0/0`.

- [x] **Step 2: Correct trusted projection-origin normalization before 6B**

`EngineMessageProjection.Origin` is an untrusted adapter hint, not ownership proof. Prepare is
the only repository boundary that can resolve durable submission/draft provenance, so it returns
the engine page byte-identically except that each live projection origin is normalized from the
matching proof: submission, draft, or `inbound_or_unknown` when neither proof exists. Dual proof
is invalid. The service must compare against exactly that proof-derived normalized page and reject
any other repository rewrite. This keeps BCC and received-event admission under the accepted 04F
policy instead of trusting Sent-folder/address/provider metadata.

Task 6R is `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `744d0288` after independent
review C/I/M `0/0/0`.

- [x] **Step 3: Correct unknown destroyed-message handling before 6B**

JMAP `/changes` may return a destroyed ID for an object created and destroyed between client
states. `MailboxSyncPolicyDestroyed.Projected` therefore distinguishes a real persisted Hub
projection from an unknown destroy. `Projected=false` requires zero message/thread identities,
empty labels and zero reservation, and produces no tombstone, count or event. Every hybrid shape
fails closed. A projected destroy retains the existing exact delete-event behavior.

Task 6U is `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at immutable commit `de1e5723` after independent
Terra/high review C/I/M `0/0/0`; launcher and all executable evidence remain
`NOT_READY` / `NOT_RUN`.

- [ ] **Step 4: Implement Task 6B prepare**

Prepare re-locks the account under the complete unexpired fence, resolves only bounded committed
projection and reservation state, and returns a deep-copied `MailboxSyncPolicyInput`. Existing
rows retain their immutable UUIDs. IDs for not-yet-persisted folders, messages and threads are
pure UUIDv8 values derived from SHA-256 over a versioned domain tag plus length-delimited raw
account UUID and exact engine identity/thread-key bytes; this introduces no clock/random callback
under the repository lock and remains stable across Prepare retries. Existing identity collisions
or cross-account matches fail closed. Labels are resolved only from existing live Hub labels;
unknown engine markers remain ignored by 04F.

- [ ] **Step 5: Implement Task 6C apply**

Apply re-locks and compares account, lease, sync, mutation, phase, cursor, base cursor, and
checkpoint before any projection write, then persists projection/count/outbox/position and its
versioned fence-plus-advance hash in one transaction. Matching same-lease proof replays without
effects; projection/event candidates are never re-applied during that proof short-circuit.

- [ ] **Step 6: Keep provider I/O outside transactions**

Repository methods only persist state. `MailboxSyncService.Step` remains the sole orchestration point around one bounded `MailEngine` call.

- [ ] **Step 7: Confirm worker selection and fault reschedule**

Select due rows by `due_at, account_id` with bounded batch and skip-locked semantics. Fault reschedule verifies the selected snapshot, advances attempts once, caps delay, and loses safely on a changed lease/fence.
The worker maps `MailboxSyncWorkerDueAccount.UpdatedAt` from `sync_schedule_updated_at`, and every
schedule-affecting mutation advances that marker from the same repository clock read.

- [ ] **Step 8: Static review and commit each atomic packet**

```text
feat(mail): persist mailbox sync and worker state
```

### Task 7A: Stalwart MailEngine adapter

**Files:**
- Create: `backend/internal/mail/stalwart/engine.go`
- Create: `backend/internal/mail/stalwart/client.go`
- Test: `backend/tests/mail/stalwart_engine_contract_test.go`
- Modify: `backend/internal/mail/config.go`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: the full `MailEngine` interface and accepted Phase 145 endpoint/auth contract.
- Produces: `stalwart.NewEngine(Config, *http.Client) (*Engine, error)` and `var _ mail.MailEngine = (*Engine)(nil)`.

- [ ] **Step 1: Implement transport and authentication**

Use separate management and JMAP base URLs/credentials, bounded request/response bodies, context deadlines, safe error mapping, and no provider text in logs/public errors.

- [ ] **Step 2: Implement the full engine surface**

Implement and test every current method, grouped exactly as follows: control (`Capabilities`, `GetDomain`, `EnsureDomain`, `EnsureAccount`); folders (`ListFolders`, `CreateFolder`, `RenameFolder`, `DeleteFolder`); projection/read (`ListMessages`, `GetMessage`, `GetBlob`, `Changes`, `EnumerateProjection`); drafts/submission (`SaveDraft`, `ReconcileDraft`, `Submit`, `ReconcileSubmission`); mutations (`MoveMessages`, `SetFlags`, `SetKeywords`, `DeleteMessages`); rules/training/health (`GetSieve`, `UpsertSieve`, `ClassifyMessages`, `Health`). Unsupported provider capability returns the existing typed unavailable/capability error and never silently succeeds.

- [ ] **Step 3: Test the HTTP boundary**

Use an isolated `httptest.Server` to cover request encoding, response bounds, cancellation/deadline, non-2xx safe mapping, malformed/oversized JSON, separate credentials, and redacted errors. No database or real network is used.

- [ ] **Step 4: Static review and commit**

```text
feat(mail): add Stalwart engine adapter
```

### Task 7B: Mail account HTTP surface

**Files:**
- Create: `backend/internal/mail/mailbox_handler.go`
- Test: `backend/tests/mail/mailbox_handler_test.go`
- Modify: `docs/mail/openapi.yaml`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: `MailboxService.ResolveOwner` only.
- Produces: authenticated `GET /api/mail/account` matching the existing `/mail/account` OpenAPI operation.

- [ ] **Step 1: Add the handler**

Resolve actor exclusively from auth context, call `ResolveOwner`, and return the existing bounded owner/account/projection/recovery representation. Map invalid auth, not configured, not found, suspended, deleting, and unavailable states to the frozen status and safe-code contract.

- [ ] **Step 2: Add isolated handler tests**

Use a fake `MailboxOwnerRepository`; cover missing/invalid actor, each owner state, typed service errors, no secret/provider identifiers, and exact route registration.

- [ ] **Step 3: Static review and commit**

```text
feat(mail): expose mailbox account status
```

All other `/api/mail` OpenAPI operations remain unregistered until their own accepted services and receipt-backed repositories exist.

### Task 7C: Canonical Calendar mutation HTTP surface

**Files:**
- Create: `backend/internal/calendar/event_handler.go`
- Test: `backend/tests/calendar/event_handler_test.go`
- Modify: `docs/mail/openapi.yaml`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: the six accepted actor-facing `EventService` methods.
- Produces: authenticated `POST /api/calendar/v2/events`, `PATCH|DELETE /api/calendar/v2/events/{eventId}`, `PUT /api/calendar/v2/events/{eventId}/rsvp`, `PUT /api/calendar/v2/events/{eventId}/instances/{recurrenceId}/rsvp`, and `DELETE /api/calendar/v2/events/{eventId}/instances/{recurrenceId}`.

- [ ] **Step 1: Implement exact route-to-service mapping**

Map routes respectively to `CreateCalendarEvent`, `UpdateCalendarEvent`, `CancelCalendarEvent`, `RespondToCalendarSeries`, `RespondToCalendarOccurrence`, and `CancelCalendarEventOccurrence`. Actor comes only from auth context; current owner/state is never accepted from the client.

- [ ] **Step 2: Add isolated handler tests**

Cover bounded DTO parsing, UUID/recurrence/version validation, typed safe status mapping, foreign-resource masking, result encoding, and exact route registration. Keep legacy GET routes in the existing handler.

- [ ] **Step 3: Static review and commit**

```text
feat(calendar): expose canonical event mutations
```

### Task 8: Main-server composition and worker lifecycle

**Files:**
- Create: `backend/cmd/server/wiring_calendar_mail.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/cmd/server/wiring_modules.go`
- Modify: `backend/cmd/server/wiring_deps.go`
- Modify: `backend/cmd/server/wiring_run.go`
- Modify: `backend/internal/mail/config.go`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: Ent repositories, Stalwart engine, injected real timer, server root context, NATS, outbox, and cleanup stack.
- Produces: fail-closed startup and graceful shutdown for Mail sync plus canonical Calendar.

- [ ] **Step 1: Build dependencies in one wiring function**

Construct repositories before services, services before handlers/workers, and consumers last. Enabled Mail sync with incomplete DB/engine/config is a startup error.

- [ ] **Step 2: Register routes without replacing legacy reads**

Move legacy Calendar registration into the new function, keep its two GET routes, and register canonical mutations separately. Avoid double registration.

- [ ] **Step 3: Own worker lifetime**

Run `MailboxSyncWorker.Run(d.rootCtx)` in one guarded goroutine, surface unexpected terminal errors through server shutdown/error reporting, and rely on root cancellation plus cleanup drain. Do not add process-local background contexts.

- [ ] **Step 4: Static review and commit**

```text
feat(mail-calendar): wire production services
```

### Task 9: Docker Ent generation and versioned migration packet

**Files:**
- Create: `backend/migrations/20260822000000_mail_calendar_core.sql`
- Create: `backend/migrations/atlas.sum`
- Modify: `.planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: accepted launcher, `backend/atlas.hcl`, all checked-in schemas.
- Produces: inspected generated Ent diff, replayable versioned SQL, checksum, and Docker evidence.

- [ ] **Step 1: Generate Ent through `/qa-gate`**

Use the accepted versioned outer launcher only. Generated Go must remain uncommitted. Expected result: generation succeeds and the diff is inspected for unexpected tables, columns, FKs, indexes, checks, or cascades.
Any schema/adapter discrepancy is a hard stop: record the exact diff, abandon this task's commit, and return to a new reviewed source packet. Task 9 authorizes no unspecified production-source edit.

- [ ] **Step 2: Generate and inspect the Atlas diff**

Use `ATLAS_DEV_URL` owned by the task container. Expected SQL contains only the reviewed Mail/Calendar schemas and Phase 147 bridge tables; no drop/destructive operation is accepted.

- [ ] **Step 3: Run migration evidence**

Through `/qa-gate`: checksum, lint, empty-database replay, realistic pre-feature snapshot replay, Calendar preflight, apply, and postconditions. Any unresolved ScheduledCall row blocks cutover.

- [ ] **Step 4: Run build/focused jobs**

Through `/qa-gate`: backend build plus applicable Calendar/Mail unit/contract/integration profiles. Record exact launcher identity, task-run report path, exit status, elapsed time, and cleanup evidence.

- [ ] **Step 5: Commit**

```text
feat(db): add mail calendar production migration
```

### Task 10: Inbound Mail-to-Calendar consumer

**Files:**
- Create: `backend/internal/calendar/mail_inbound_consumer.go`
- Create: `backend/internal/calendar/mail_inbound_repository.go`
- Create: `backend/internal/calendar/mail_inbound_service.go`
- Test: `backend/tests/calendar/mail_inbound_service_test.go`
- Modify: `backend/cmd/server/wiring_calendar_mail.go`
- Modify: `.planning/phases/147-mail-calendar-reliability/CONTRACT.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: `mail.received.v1`, bounded Mail read port, Phase 147 schemas, parser, and accepted Calendar services.
- Produces: leased claim, trusted-authentication admission, REQUEST/CANCEL ordering, inbound REPLY application, and atomic terminal receipt/outbox.

- [ ] **Step 1: Implement envelope admission**

Require broker/JSON subject equality, schema version 1, valid application event UUID, durable organization/account ownership, and authentication class `trusted_internal|dmarc_aligned`. Finalize all other input as safe quarantine before service dispatch.

- [ ] **Step 2: Implement the two-transaction lease protocol**

Claim transaction locks only the receipt and commits a bounded lease. Load/parse happens outside DB transactions. Terminalization locks receipt first, proves the lease token, then locks Calendar rows in canonical order; lease loss performs no mutation.

- [ ] **Step 3: Implement REQUEST/CANCEL state decisions**

Enforce organizer immutability, per-series/per-occurrence sequence and hash-version comparison, early-cancel tombstones, deferred occurrences, 10,000 unbound cap, seven-day dead-letter, and atomic event/object-state/receipt/outbox commit.

- [ ] **Step 4: Route REPLY through the accepted service**

Pass only canonical parsed identity. Never persist raw MIME/iTIP; duplicate and stale results replay stored outcomes.

- [ ] **Step 5: Review, Docker verification, and commit**

```text
feat(calendar): consume inbound mail invitations
```

### Task 11: Calendar-to-Mail delivery consumer

**Files:**
- Create: `backend/internal/mail/calendar_delivery_consumer.go`
- Create: `backend/internal/mail/calendar_reply_worker.go`
- Create: `backend/internal/mail/mailbox_delivery_repository.go`
- Test: `backend/tests/mail/calendar_delivery_service_test.go`
- Modify: `backend/cmd/server/wiring_calendar_mail.go`
- Modify: `.planning/phases/147-mail-calendar-reliability/CONTRACT.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: frozen Calendar envelope subjects, `CalendarReplyDelivery`, Mail draft/submission services, and owner-bound active account/identity resolution.
- Produces: idempotent invitations/cancellations and reconciled outward RSVP replies.

- [ ] **Step 1: Admit and deduplicate Calendar envelopes**

Validate subject/envelope equality and application event UUID. Resolve destination from durable Calendar/Mail ownership; never trust caller/provider account or From identity.

- [ ] **Step 2: Implement stable delivery keys**

Derive hidden draft and submission idempotency keys from delivery UUID. `unknown` always reconciles before retry; no blind second submit is possible.

- [ ] **Step 3: Persist safe outcomes**

Update delivery state and attendee public reply status atomically where required. Provider error text, MIME, serializer output, and credentials remain request-lifetime only.

- [ ] **Step 4: Review, Docker verification, and commit**

```text
feat(mail): deliver calendar invitations and replies
```

### Task 11B: Bounded ScheduledCall backfill

**Files:**
- Create: `backend/internal/calendar/scheduled_call_backfill.go`
- Create: `backend/cmd/calendar-backfill/main.go`
- Test: `backend/tests/calendar/scheduled_call_backfill_test.go`
- Modify: `.planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: the applied Task 9 schema, frozen Calendar recurrence policy, and legacy `ScheduledCall` rows.
- Produces: `BackfillScheduledCalls(context.Context, BackfillRepository, BackfillOptions) (BackfillReport, error)` with explicit `DryRun`, `BatchSize`, and report identity.

- [ ] **Step 1: Define a bounded repository port and report**

The port pages non-deleted calls in stable UUID order, resolves organization/project/channel/creator/participants, writes one event plus deduplicated attendees plus linkage in a transaction, and records no outbox. The report contains source revision, candidate/excluded/linked/already-linked/failed counts, ordered row hashes, recurrence-oracle version, start/end instants, and no content bodies.

- [ ] **Step 2: Implement dry-run preflight**

Validate every identity, active membership, duration, title/description/RRULE/reminder bound, creator channel membership, recurrence equivalence, existing linkage, and target uniqueness without writes. Any unresolved row fails the whole preflight.

- [ ] **Step 3: Implement idempotent apply**

Require a byte-equivalent successful dry-run report identity. Process bounded transactions, preserve source timestamps, collapse duplicate creator participation, use `not_required` reply delivery, set no inbound payload hash, emit no notification/NATS/outbox side effect, and never overwrite an already-linked event.

- [ ] **Step 4: Implement recovery and evidence**

Before apply require backup identity and legacy-writer/due-worker guard. After apply verify counts, hashes, 100% non-deleted linkage, and event/call coherence. Failure disables cutover and preserves rows for forward-only correction; there is no automatic delete rollback.

- [ ] **Step 5: Docker verification and commit**

Run dry-run, apply, idempotent rerun, postconditions, and failure fixtures only through `/qa-gate`.

```text
feat(calendar): add ScheduledCall backfill
```

### Task 12: Cutover, final verification, and merge readiness

**Files:**
- Modify: `backend/internal/calls/scheduled.go`
- Modify: `backend/cmd/server/wiring_calendar_mail.go`
- Modify: `docs/mail/QA-HANDOFF.md`
- Modify: `.planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: 100% linkage/postcondition evidence and all prior Docker reports.
- Produces: guarded canonical Calendar writer ownership and a merge-readiness verdict.

- [ ] **Step 1: Add the compatibility flag**

Before cutover, existing ScheduledCall writers remain legacy-owned. After the verified flag activates, linked rows delegate through the event-before-call bridge and direct schedule mutation fails closed.

- [ ] **Step 2: Reconcile before enable**

Require all non-deleted calls linked, zero event/call mismatch, recurrence oracle pass, no unauthorized writer, and durable outbox/receipt recovery evidence.

- [ ] **Step 3: Run final Docker gates**

Use `/qa-gate` only: generation cleanliness, migration replay, backend build, applicable Mail/Calendar focused jobs, integration, race where applicable, and runtime smoke. Record any unrun job as `NOT_RUN`, never PASS.

- [ ] **Step 4: Independent final review**

Run Contract, Security, Lifecycle, code, and integration reviews against immutable HEAD. Confirm clean tree, exact branch ancestry, no `testframework/**` or `pro-qa-ai/**` changes, and no generated Ent files committed.

- [ ] **Step 5: Mark merge-ready only after all gates**

```text
docs(mail-calendar): record production integration acceptance
```

Do not merge to `master`; hand off the verified branch for founder approval.
