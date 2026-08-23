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

## Deferred test ledger — LOCKED pending framework cutover (2026-08-23)

> **LOCKED — 2026-08-23.** Until the new `testframework` skeleton is merged
> into `master` **and this branch is updated from that master**, Mail work
> writes or modifies **no tests at all**: backend, frontend, and Python test
> files are out of scope. Implementation packets contain production source
> only; existing tests are not edited; every executable claim remains
> `NOT_RUN`. The current gate is exact-scope static source review. After that
> sync, a separate dedicated test plan backfills coverage against the
> then-current contracts.

The following current test ownership is deferred rather than removed from the
behavior contracts. No row below is executable evidence before the cutover and
the dedicated test plan.

| Packet | Deferred test paths / coverage |
|---|---|
| 6Q / 6V | `backend/tests/mail/mailbox_ent_sync_prepare_contract_test.go`; `backend/tests/mail/mailbox_ent_sync_coordination_contract_test.go` |
| 6C0 | `backend/tests/mail/mailbox_sync_apply_canonical_test.go` |
| 6C1 | `backend/tests/mail/mailbox_ent_sync_apply_projection_contract_test.go` |
| 6C2 | `backend/tests/mail/mailbox_ent_sync_apply_contract_test.go` |
| 7A | `backend/tests/mail/stalwart_engine_contract_test.go` |
| 7B | `backend/tests/mail/mailbox_handler_test.go` |
| 7C | `backend/tests/calendar/event_handler_test.go` |
| 10 | `backend/tests/calendar/mail_inbound_service_test.go` |
| 11 | `backend/tests/mail/calendar_delivery_service_test.go` |
| 11B | `backend/tests/calendar/scheduled_call_backfill_test.go` |
| Mail frontend | Route, accessibility, RSVP, card-invalidation, and delivery-verification coverage; no frontend test file is newly owned before the cutover. |

Task 6W is historical: production schema source completed at `e0c2762b`;
its test line was restored but net-unchanged at `4c757048`. That history is
not a test write, not a test review, and not executable coverage.

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
- Historical pre-lock evidence only; no current test ownership; executable NOT_RUN: `backend/tests/calendar/calendar_mail_bridge_schema_test.go`
- Modify: `.planning/phases/147-mail-calendar-reliability/CONTRACT.md`
- Modify: `.planning/STATE.md`

**Status:** `HISTORICAL_PRE_LOCK_STATIC_ACCEPTED / RUNTIME_NOT_RUN` at
commits `55b2f2f4..f9e070a7`. All numbered steps below are historical static
record only, not active implementation or test-authoring work.

**Interfaces:**
- Consumes: exact field/state/retention rules in Phase 147 sections 2.1, 2.2, and 3.1.
- Produces: durable ordering state, leased inbound receipt, and immutable outward reply intent.

- [x] **Historical pre-lock static schema-contract evidence only (non-actionable)**

The historical evidence imports `backend/ent/schema` only and records exact fields, enum values, immutable identities, checks, edges, and indexes, including the two partial unique indexes for nullable/non-null `recurrence_id`, unique message plus locator hash, and unique attendee plus event sequence plus attendee version. It establishes no current test ownership or executable coverage.

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
- Historical pre-lock evidence only; no current test ownership; executable NOT_RUN: `backend/tests/calendar/event_ent_repository_contract_test.go`
- Modify: `.planning/phases/143-calendar-rsvp-core/PLAN.md`
- Modify: `.planning/STATE.md`

**Status:** `HISTORICAL_PRE_LOCK_STATIC_ACCEPTED / RUNTIME_NOT_RUN` at
commits `d03bfee9..dafd86b4`. All numbered steps below are historical static
record only, not active implementation or test-authoring work.

**Interfaces:**
- Consumes: `EventRepository`, mandatory `EventTransactionBackend`, and `RunEventTransaction`.
- Produces: `NewEventEntRepository(*ent.Client) *EventEntRepository` and compile-time `var _ EventRepository = (*EventEntRepository)(nil)`.

- [x] **Historical pre-lock static transaction-contract evidence only (non-actionable)**

The historical evidence records nil client/callback, begin failure, callback-error rollback, commit failure, transaction-lifetime closure, one transaction per call, and no callback reuse. It establishes no current test ownership or executable coverage.

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
- Historical pre-lock evidence only; no current test ownership; executable NOT_RUN: `backend/tests/calendar/event_ent_capabilities_contract_test.go`
- Modify: `.planning/phases/143-calendar-rsvp-core/PLAN.md`
- Modify: `.planning/STATE.md`

**Status:** `HISTORICAL_PRE_LOCK_STATIC_ACCEPTED / RUNTIME_NOT_RUN` at
commits `64c20b11..92cf944c`. All numbered steps below are historical static
record only, not active implementation or test-authoring work.

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
- Historical pre-lock evidence only; no current test ownership; executable NOT_RUN: `backend/tests/calendar/event_call_repository_contract_test.go`
- Historical pre-lock evidence only; no current test ownership; executable NOT_RUN: `backend/tests/calendar/event_call_policy_test.go`
- Modify: `docs/superpowers/plans/2026-08-22-mail-calendar-production-integration.md`
- Modify: `.planning/phases/143-calendar-rsvp-core/PLAN.md`
- Modify: `.planning/STATE.md`

**Status:** `HISTORICAL_PRE_LOCK_STATIC_ACCEPTED / RUNTIME_NOT_RUN` at
commit `7fb25100`. All numbered steps below are historical static record only,
not active implementation or test-authoring work.

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
- Historical pre-lock evidence only; no current test ownership; executable NOT_RUN: `backend/tests/mail/mailbox_ent_owner_contract_test.go`
- Modify: `.planning/phases/144-mailbox-domain/PLAN.md`
- Modify: `.planning/STATE.md`

**Status:** `HISTORICAL_PRE_LOCK_STATIC_ACCEPTED / RUNTIME_NOT_RUN` at
commits `8a8e3221..ea0808a7`. All numbered steps below are historical static
record only, not active implementation or test-authoring work.

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

### Task 6W: Correct bounded sync-lease proof retention prerequisite

**Files:**
- Modify: `backend/ent/schema/mailaccount.go`
- Historical test line (restored/net-unchanged; no current ownership): `backend/tests/mail/mail_account_identity_schema_test.go`

**Status:** `SOURCE_COMPLETE_STATIC / RUNTIME_NOT_RUN` at source commit
`e0c2762b`.

The production schema source completed at `e0c2762b`. Its test line was
restored but net-unchanged at `4c757048`; that historical line is not a test
write, does not establish test review, and is never executable coverage.

This is a forward correction to the bounded proof constraint in accepted Task 6P; it does not
alter Task 6P's recorded static acceptance. Extend
`mail_accounts_sync_lease_apply_proof` so a complete proof also requires
`sync_lease_applied_next_at < sync_lease_until`. The pre-lock schema-contract
line was restored net-unchanged at `4c757048`; it records that strict predicate
and equality/reversal rejection only, establishes no current test ownership or
executable coverage, and any future coverage is deferred to the LOCKED
deferred-test ledger. Task 6C2 writes
`sync_lease_applied_next_at = NextAttemptAt` and a checked exact-second
`sync_lease_until = NextAttemptAt + 1s`; the one-second interval is the minimum grace for durable
same-lease replay, and a new claim remains permitted after lease expiry. Overflow, non-exact-second
values, and any incomplete proof fail closed. No retention table, migration, or cutover authority is
added; launcher and all executable evidence remain `NOT_READY` / `NOT_RUN`.

### Task 6: Mail sync and worker Ent adapters

**Files:**
- Create: `backend/internal/mail/mailbox_ent_sync_claim.go`
- Create: `backend/internal/mail/mailbox_ent_sync_prepare.go`
- Create: `backend/internal/mail/mailbox_sync_apply_canonical.go`
- Create: `backend/internal/mail/mailbox_ent_sync_apply_projection.go`
- Create: `backend/internal/mail/mailbox_ent_sync_apply.go`
- Create: `backend/internal/mail/mailbox_ent_worker.go`
- Modify: `.planning/phases/144-mailbox-domain/PLAN.md`
- Modify: `.planning/STATE.md`

**Interfaces:**
- Consumes: `MailboxSyncRepository`, `MailboxSyncWorkerRepository`, accepted sync fence/page policy, and worker caps.
- Produces: the same `MailboxEntRepository` satisfying both interfaces.

**Dependency:** Task 6 starts only after Task 6S, Task 6P, Task 6T, and Task 6W persist and
independently statically review the durable sync-lease generation, active due-scan index, bounded
apply proof including its strict retention predicate, and exact-second schedule CAS marker.

**Atomic decomposition:** Task 6A owns claim/resolve plus due-scan/fault coordination; Task 6B owns
prepare-only projection reads and policy-input construction. The active prerequisite order is
6W -> 6Q2 -> 6C0V2 -> 6C1 -> 6C2. 6Q2 replaces the superseded lease-clock source, and 6C0V2
replaces the pending v1 canonical handoff; both modify `mailbox_ent_sync_prepare.go`, so they are
strictly sequential. Task 6V's retry decoder remains documented separately: source `43a` is
state-machine-correct, but its final quality decision waits for 6Q2. Each 6C packet has a separate
implementer and independent exact-scope static review. No packet spans provider I/O and a repository
transaction.

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

- [x] **Step 4: Implement Task 6B prepare** — `STATIC_ACCEPTED / EXACT_SHA_PASS / RUNTIME_NOT_RUN`

Prepare re-locks the account under the complete unexpired fence, resolves only bounded committed
projection and reservation state, and returns a deep-copied `MailboxSyncPolicyInput`. Existing
rows retain their immutable UUIDs. IDs for not-yet-persisted folders, messages and threads are
pure UUIDv8 values derived from SHA-256 over a versioned domain tag plus length-delimited raw
account UUID and exact engine identity/thread-key bytes; this introduces no clock/random callback
under the repository lock and remains stable across Prepare retries. Existing identity collisions
or cross-account matches fail closed. Labels are resolved only from existing live Hub labels;
unknown engine markers remain ignored by 04F.

Task 6B is statically accepted at effective immutable HEAD `431022ed` after source introduction
`6e788cc7` (`feat(mail): add Ent sync prepare adapter`) and correction
`431022ed` (`fix(mail): exclude deleted sync threads`). Its exact effective manifest is
`backend/internal/mail/mailbox_ent_sync_prepare.go` plus authored source-contract
`backend/tests/mail/mailbox_ent_sync_prepare_contract_test.go`. The one short read-only Ent
transaction takes one repository clock read, locks the active desired account `FOR UPDATE`, and
requires the complete current+persistent fence and unexpired lease. It has no provider/NATS/HTTP,
randomness, or writes. Initial/resumed enumeration and incremental/catch-up require exact matching;
an empty valid page is accepted. It deterministically derives stable length-delimited SHA-256 UUIDv8
folder/message/thread IDs, reuses existing live UUIDs with collision guards, normalizes untrusted
origin from live-only confirmed sent submission/stable draft provenance, resolves live
ancestors/threads, rejects duplicate labels within one projection and deduplicates them across the
page, and returns all-and-only unused event candidates plus a deep-copied policy input.

A live soft-deleted reappearance reuses its message UUID and tombstone previous
state/`ReceivedEventID`, so it cannot duplicate `mail.received.v1`; tombstone folders/flags are zero
and only still-live labels remain. Already-deleted or unknown destroyed rows remain
`Projected=false` and are excluded from state/provenance work. Precommit review findings for current
epoch fencing and retained folder links were corrected; the first immutable review then found the
deleted-thread resolver gap, corrected by `431022ed`. Independent immutable Terra/high rereview of
effective HEAD passed C/I/M `0/0/0`; provenance is `dispatcher-enforced`, telemetry `UNKNOWN`.
Generated Ent client/predicates are absent and the launcher remains `NOT_READY`; generation,
compile, tests, formatting/lint, Atlas/migration, SQL/DB, Docker, and runtime are `NOT_RUN`. Task 6C
apply remains pending; this accepts neither Task 6, Phase 144, the Mail module, nor product.

- [ ] **Step 5: Historical Task 6Q lease-time source -- superseded/pending correction**

**Files:**
- Modify: `backend/internal/mail/mailbox_ent_repository.go`
- Modify: `backend/internal/mail/mailbox_ent_sync_prepare.go`
- Modify: `backend/internal/mail/mailbox_ent_sync_claim.go`

Source `6477` is `SUPERSEDED / PENDING_CORRECTION / RUNTIME_NOT_RUN`. It is retained only to
explain the defect and must not be accepted, reimplemented, or used to discharge a later dependency.
Its `CURRENT_TIMESTAMP` helper is superseded by Task 6Q2 below.

The historical source froze this private helper in `mailbox_ent_repository.go`:

```go
func mailboxEntTransactionTime(ctx context.Context, transaction *ent.Tx) (time.Time, error)
```

It uses the transaction-bound Ent dialect driver only: `transaction.Driver().Query` executes exactly
`SELECT date_trunc('second', CURRENT_TIMESTAMP)`, reads and scans exactly one row, normalizes the
result to UTC, and admits it only through `mailboxEntValidTime`. Query, row-count, scan, UTC, and
validation failures map through the existing content-free `mailboxEntSafeError` path. It never uses
`repository.client`, opens a new transaction, samples the injected clock, or accepts request/caller
time as the lease-expiry authority.

Prepare and Resolve each call the helper exactly once per transaction they own. Prepare feeds the
value into the otherwise unchanged `mailboxEntPrepareInput`; Resolve uses the same value for both
`SyncLeaseUntil.After(transactionTime)` and its schedule timestamp. Claim is unchanged. Apply uses
the same authority only in its separately frozen 6C2 work. Exact-scope static source review
asserts the exact query/one-time owned-transaction behavior, unchanged Prepare input semantics, and
Resolve expiry/schedule use. The deferred-test ledger reserves executable coverage for the dedicated
post-sync test plan; generated Ent, runtime, compile, SQL/DB, Docker, formatting/lint, and all
executable evidence remain `NOT_RUN`. This forward correction preserves the recorded immutable
acceptance of Tasks 6A and 6B.

- [ ] **Step 5Q2: Implement fresh database-wall-clock prerequisite**

**Files:**
- Modify: `backend/internal/mail/mailbox_ent_repository.go`
- Modify: `backend/internal/mail/mailbox_ent_sync_claim.go`
- Modify: `backend/internal/mail/mailbox_ent_sync_prepare.go`

Freeze the replacement private helper exactly as:

```go
func mailboxEntDatabaseWallClock(ctx context.Context, transaction *ent.Tx) (time.Time, error)
```

It uses only the supplied transaction dialect driver to execute exactly
`SELECT date_trunc('second', clock_timestamp())`, reads and scans exactly one row, converts the
returned instant to UTC, and accepts it only as an exact-second `mailboxEntValidTime` instant.
Nil context/transaction, query, row-count, scan, close, UTC conversion, and validation failures map
through the existing content-free safe mapping; it never starts another transaction, reaches the
repository client, or samples the injected process clock. `CURRENT_TIMESTAMP` is not acceptable:
PostgreSQL fixes it at transaction start, so a value obtained after a lock can already be stale and
cannot prove the current lease boundary.

Claim, Prepare, and Resolve lock the requested account with `FOR UPDATE` first, then call
`mailboxEntDatabaseWallClock` exactly once in the successful owned transaction. Claim uses that one
fresh DB value as the only due, lease, schedule-marker, and returned non-claim authority. After the
lock and fresh clock, but before any lease mutation, Claim rejects
`SyncAttempts < 0 || SyncAttempts >= mailboxEntMaxSyncAttempts` through the existing content-free
invalid error, with zero writes and no engine I/O. It rejects `request.LeaseUntil` unless it is
strictly after that value. `request.Now` remains structural input validation only; it is never
compared as the authority for due, lease expiry, schedule, or a returned time. A no-active-lease
Claim retains `sync_lease_epoch` (pristine first claim is `1`; a released lease is already advanced),
a live same-token renewal retains it, and an expired retained lease increments it as before. Prepare
passes its one post-lock value to the unchanged `mailboxEntPrepareInput`. Resolve uses its one
post-lock value for every lease gate and for `SyncScheduleUpdatedAt` only; it persists and returns the
exact `request.NextAttemptAt` without clamp or rebase. A next attempt at or before the DB wall clock
is intentional immediate due; the DB clock remains authority only for lease/due decisions and the
schedule marker.

Future 6C2 follows the same post-lock rule and uses its one fresh DB value for every live-lease gate,
`appliedAt`, projection timestamp, and schedule marker. Before a proof write it requires
`NextAttemptAt` to be strictly after that fresh wall clock and checks that
`NextAttemptAt + 1s` is still a valid exact-second instant strictly later than it. This is not an
authorization to implement 6C2 early.

**Static-review gate only:** independently inspect the exact three-file source manifest for the
literal query, one-row safe mapping, account-lock-before-single-call order, Claim authority split,
and Prepare/Resolve use. No test, build, generation, format, lint, SQL/DB, Docker, or runtime action
is authorized; all executable evidence remains `NOT_RUN`.

- [ ] **Step 6: Implement Task 6V Resolve projection-state and retry-attempt correction**

**Files:**
- Modify: `backend/internal/mail/mailbox_ent_sync_claim.go`

Sources `57b` and `43a` are `PENDING_RETRY_CAP_AND_FENCE_GENERATION / RUNTIME_NOT_RUN`, not static
accepted. `43a` remains state-machine-correct for decoder reconstruction, but final quality cannot
close until this one-file Claim/6Q2 correction is independently reviewed. The exact source manifest
remains `backend/internal/mail/mailbox_ent_sync_claim.go`; it changes the decoder plus Claim/Resolve
retry-cap and lease-generation behavior only, with no schema, public API, or migration change. The
locked production-only test rule remains in force.

After a handled failure, `mailboxEntSyncClaimSnapshot` reconstructs the degraded policy snapshot
exactly as follows. `CURSOR_INVALID` and `RESTART_REQUIRED` require nil rebuild fields
(`sync_base_cursor` and `sync_checkpoint`) and yield an empty bootstrap `Enumeration`. For
`transient` or `rejected`: no base/no checkpoint plus a valid `sync_cursor` yields `Incremental`
with that same cursor; no base/no checkpoint plus nil `sync_cursor` yields empty bootstrap
`Enumeration`; valid base/no checkpoint yields `CatchUp` with `Cursor=base`; and valid base plus
valid checkpoint yields `Enumeration` with that pair. A checkpoint without a base, or any invalid
opaque cursor, returns the existing content-free invalid error. Other degraded codes retain the
prior incremental-only guard.

Each of the four handled kinds remains `degraded`: `transient`, `rejected`, `invalid_cursor`, and
`restart_required`. After the complete live fence, Resolve computes
`mailboxEntNextEpoch(account.SyncLeaseEpoch)` and rejects an overflow through the existing
content-free invalid error with zero writes. In its one atomic update it calls
`SetSyncLeaseEpoch(next)` together with exactly one `sync_attempts` increment, the exact
`request.NextAttemptAt` schedule (no clamp or rebase), safe error, projection state, lease release,
and proof clear. This prevents same-`LeaseID` ABA. A second Resolve has no lease and therefore
performs no second increment. A `ready` account cannot retain `sync_error_code`; only successful
Apply may reset attempts to zero, while worker fault reschedule owns its separate failure increment
outside this path. Exact-scope static source review asserts the degraded snapshot shapes plus
retry-cap, epoch-overflow/atomic-generation, exact-next-attempt, lease/proof-clear, and
idempotent-second-Resolve rules. The deferred-test ledger reserves executable coverage for the
dedicated post-sync test plan; generated Ent, runtime, compile, SQL/DB, Docker, formatting/lint, and
all executable evidence remain `NOT_RUN`.

- [ ] **Step 7: Historical Task 6C0 canonical v1 handoff -- pending provenance v2**

**Files:**
- Create: `backend/internal/mail/mailbox_sync_apply_canonical.go`

Current 6C0 through source `52190` is `PENDING_PROVENANCE_V2 / RUNTIME_NOT_RUN`; it is not static
accepted and cannot satisfy 6C1 or 6C2. The following v1 registry is preserved as historical context
only and is superseded by 6C0V2.

The historical handoff froze the exported-for-internal-package implementation API:

```go
type MailboxSyncCanonicalApply struct {
    Fence         MailboxSyncFence
    Result        MailboxSyncPolicyResult
    ExpectedPhase MailboxSyncPhase
    NextAttemptAt time.Time
    ResultHash    [sha256.Size]byte
}

func CanonicalizeMailboxSyncApplyRequest(MailboxSyncApplyRequest) (MailboxSyncCanonicalApply, error)
```

This is an implementation API inside `backend/internal/mail` for the genuine unit and Task 6C2,
not a new repository, service, or HTTP port. Its package-fixed proof hash version is `int32(1)`;
no caller supplies an alternative. It returns a deep-copied canonical result and is deterministic and
side-effect-free: no Ent query, transaction, projection mutation, outbox append, clock read,
network I/O, or random operation. The result hash is SHA-256 over the normative v1 stream below.

**Normative v1 TLV stream.** All multibyte unsigned and signed integers are fixed-width
big-endian; signed `i64` uses two's-complement. A TLV field is exactly `u16 tag BE || u64
payload_len BE || payload`, with no padding or trailer. The outer stream contains these fields once,
in this strictly ascending order:

| Tag | Payload |
|---:|---|
| 1 | the literal raw UTF-8 bytes `hub.mail.sync.apply.result.v1`, with no NUL, prefix, or suffix beyond its TLV |
| 2 | `u16 BE(1)` format version |
| 3 | complete `Fence` record |
| 4 | `NextAttemptAt` canonical UTC time |
| 5 | complete `Result` record, including canonical full event envelopes |

For a concrete normative prefix, tags 1 and 2 are exactly:

```text
0001000000000000001d6875622e6d61696c2e73796e632e6170706c792e726573756c742e7631
000200000000000000020001
```

Every nested record is concatenated TLVs in strictly ascending registered-tag order. Every registered
tag occurs exactly once; optional and union values retain their tag and use the optional payload
below. Missing, duplicate, or out-of-order registered fields are invalid; unknown fields are
unsupported under v1. Strings are their validated raw UTF-8 bytes, byte arrays are raw bytes,
booleans are exactly one byte `00` or `01`, UUIDs are raw 16 bytes, and UTC time is `i64 BE`
seconds followed by `u32 BE` nanos. An optional payload is one byte `00` when absent, or byte `01`
followed by the canonical value when present. A collection is `u32 BE count`, then for each element
`u64 BE element_len || canonical element bytes`. Enums use their canonical validated wire string
unless a row explicitly freezes a numeric code. No uvarint, RFC3339, implementation-order, or
JSON substitute is permitted for a record; the sole exception is `EventEnvelope` tag 3's immutable
frozen envelope bytes as registered below.

**Complete nested-record registry.** A `record` value below is the registered nested TLV stream;
`collection<T>` uses the collection framing above. All set-like collections use the stated stable
sort and canonical bytes as the final tie-break; semantic address and reference collections retain
input order.

| Record | Ordered tags and canonical payloads |
|---|---|
| `Fence` | `1 AccountID: uuid`; `2 LeaseID: uuid`; `3 LeaseEpoch: i64`; `4 SyncEpoch: i64`; `5 MutationEpoch: i64`; `6 Phase: u8` (`1=incremental`, `2=enumeration`, `3=catch_up`); `7 Cursor: string`; `8 BaseCursor: string`; `9 Checkpoint: string`. |
| `Result` | `1 Advance: record`; `2 Folders: collection<FolderIntent>` sorted by `Folder.ID`, then `Snapshot.FolderID`; `3 Upserts: collection<Upsert>` sorted by `Message.ID`, then `Projection.MessageID`; `4 Tombstones: collection<Tombstone>` sorted by `Message.ID`, then `EngineMessageID`; `5 Counts: record`; `6 Events: collection<EventEnvelope>` sorted by `EventID`, then full canonical event bytes. |
| `Advance` | `1 SyncEpoch: i64`; `2 Change: optional<ChangeAdvance>`; `3 Enumeration: optional<EnumerationAdvance>`. Exactly one of tags 2 and 3 is present (`01`); both absent or both present is invalid. |
| `ChangeAdvance` | `1 FromCursor: string`; `2 NextCursor: string`. |
| `EnumerationAdvance` | `1 BaseCursor: string`; `2 Checkpoint: string`; `3 NextCheckpoint: string`; `4 Complete: bool`. |
| `FolderIntent` | `1 Folder: MailboxEventID`; `2 Snapshot: EngineFolder`. |
| `EngineFolder` | `1 AccountID: string`; `2 FolderID: string`; `3 ParentID: string`; `4 Name: string`; `5 Role: enum string`. |
| `Upsert` | `1 Message: MailboxEventID`; `2 Thread: MailboxEventID`; `3 ThreadIdentity: record`; `4 Projection: record`; `5 FolderIDs: collection<MailboxEventID>` sorted by `ID`; `6 Labels: collection<LabelState>` sorted by `Label.ID`; `7 ReceivedEventID: optional<uuid>`. |
| `ThreadIdentity` | `1 KeyVersion: u8` (currently `1`); `2 Key: raw 32 bytes`; `3 Source: enum string`; `4 NormalizedSubject: string`; `5 CanonicalParticipants: collection<string>` sorted bytewise. |
| `Projection` | `1 AccountID: string`; `2 MessageID: string`; `3 Version: string`; `4 InternetMessageID: string`; `5 InReplyTo: string`; `6 References: collection<string>` in semantic reference order; `7 From: EngineAddress`; `8 To: collection<EngineAddress>` in semantic address order; `9 CC: collection<EngineAddress>` in semantic address order; `10 BCC: collection<EngineAddress>` in semantic address order; `11 ReplyTo: collection<EngineAddress>` in semantic address order; `12 Subject: string`; `13 Snippet: string`; `14 SentAt: optional<time>`; `15 ReceivedAt: optional<time>`; `16 SizeBytes: i64`; `17 Priority: enum string`; `18 FolderIDs: collection<string>` sorted bytewise; `19 Flags: Flags`; `20 ReservedKeywords: collection<string>` sorted bytewise; `21 Attachments: collection<Attachment>` sorted by `BlobID`, then `ContentID`; `22 Origin: enum string`. |
| `EngineAddress` | `1 Email: string`; `2 Name: string`. |
| `Flags` | `1 Seen: bool`; `2 Answered: bool`; `3 Draft: bool`; `4 Flagged: bool`; `5 Junk: bool`. |
| `Attachment` | `1 BlobID: string`; `2 ContentID: string`; `3 Filename: string`; `4 MediaType: string`; `5 SizeBytes: i64`; `6 Disposition: enum string`; `7 SHA256: raw bytes`. |
| `LabelState` | `1 Label: MailboxEventID`; `2 Manual: bool`; `3 EngineRule: bool`. |
| `Tombstone` | `1 EngineMessageID: string`; `2 Message: MailboxEventID`; `3 Labels: collection<LabelState>` sorted by `Label.ID`. |
| `Counts` | `1 Folders: collection<MailboxEventID>` sorted by `ID`; `2 Threads: collection<MailboxEventID>` sorted by `ID`. |
| `MailboxEventID` | `1 AccountID: uuid`; `2 ID: uuid`. |
| `EventEnvelope` | `1 Subject: string` from `events.EnvelopeV1.Subject()`; `2 EventID: uuid` from `EventID()`; `3 FrozenJSON: raw bytes` from `MarshalJSON()`. `FrozenJSON` is the complete validated envelope, not a partial payload or re-marshaled substitute. |

The canonicalizer validates every value before encoding, requires all result collections and nested
records above (including empty collections), and rejects a malformed/unknown enum or an envelope
that cannot yield its complete frozen JSON. `ExpectedPhase` is derived and returned only after this
full `Fence` and `Advance` validation; it is not an unregistered sixth outer field. The only excluded
inputs are request `Now` and database `appliedAt`.

The dedicated post-sync test plan must freeze this prefix, complete byte vectors and hashes spanning
every registry branch, deep-copy isolation, canonical sorting, semantic order preservation, and
rejection of malformed/missing/duplicate/out-of-order/unknown fields before 6C1/6C2 consume it.
Until then, exact-scope static source review is the only gate and executable evidence is `NOT_RUN`.

- [ ] **Step 7C0V2: Implement source-provenance canonicalization prerequisite**

**Files:**
- Modify: `backend/internal/mail/mailbox_types.go`
- Modify: `backend/internal/mail/mailbox_sync_policy.go`
- Modify: `backend/internal/mail/mailbox_ent_sync_prepare.go`
- Modify: `backend/internal/mail/mailbox_sync_service.go`
- Modify: `backend/internal/mail/mailbox_fake.go`
- Modify: `backend/internal/mail/mailbox_sync_apply_canonical.go`

`EngineMessageProjection` gains source-authoritative `CalendarMethod *MailCalendarMethod`. Only the
future provider adapter may derive it, from bounded server-side MIME/calendar parsing; it validates,
copies, and budgets that parsed result before exposing a projection. Provider event JSON, caller data,
and persisted rows are never a semantic authority for it. Prepare removes persisted-row
`CalendarMethod` as a policy source and preserves the projection value through its bounded deep-copy
boundary.

`MailboxSyncPolicyMessageIntent` gains `CalendarMethod *MailCalendarMethod` and a concrete non-nil
`ChangedFields []MailChangedField`. `syncClassifyMessage` copies `CalendarMethod` from the current
projection and fills `ChangedFields` from the policy delta; first-seen and no-delta upserts carry an
empty concrete slice, not `nil`. Service and fake clone paths copy both fields safely, including the
method pointer and changed-field slice. Future 6C1 and 6C2 persist the intent's `CalendarMethod` to
the existing column; this packet adds no schema migration.

The canonicalizer validates that every received event's `CalendarMethod` equals both the projection
and intent source, and that each live changed event carries exactly the intent's ordered
`ChangedFields`. A deletion remains the tombstone event with `[deleted]`. An upsert changed-field
sequence may contain only, in this exact canonical order, `folder_ids`, `seen`, `answered`, `flagged`,
`junk`, and `labels`; it has at most six entries, contains no duplicate, and forbids `deleted`.

Canonical format and proof hash version become `2`; the outer domain is exactly
`hub.mail.sync.apply.result.v2`. `Upsert` appends tag `8` as optional `CalendarMethod` enum string
and tag `9` as a `ChangedFields` collection of enum strings. A stored same-lease proof must have
hash version `2`; a v1 stored proof is a mismatch and fails closed, with no migration, overwrite,
projection write, event append, or alternate JSON/event authority.

Task 7A is the only future provider boundary allowed to add the bounded MIME/calendar derivation;
it may not synthesize or infer the method from folders, message state, event JSON, or a prior row.

**Static-review gate only:** independently inspect the exact six-file manifest for provenance source,
validation/budget/deep-copy boundaries, concrete delta semantics, v2 TLV tags/domain/version, and
fail-closed v1 proof handling. No test, build, generation, format, lint, migration, SQL/DB, Docker,
or runtime action is authorized; all executable evidence remains `NOT_RUN`.

- [ ] **Step 8: Implement Task 6C1 bounded projection and aggregate helpers**

**Files:**
- Create: `backend/internal/mail/mailbox_ent_sync_apply_projection.go`

Freeze the private transaction handoff (a repository receiver is permitted only if required by the
existing private helpers):

```go
func (r *MailboxEntRepository) mailboxEntApplySyncProjection(
    ctx context.Context,
    tx *ent.Tx,
    canonical MailboxSyncCanonicalApply,
    transactionTime time.Time,
) error
```

It starts only after independently reviewed 6Q2 and 6C0V2. It receives a checked exact-second
`transactionTime` explicitly and does not read a clock, open or
commit a transaction, or append an outbox row. It owns only bounded projection/count/thread
aggregate persistence. Thread effective latest time is the later non-nil `SentAt`/`ReceivedAt`; if
both are nil, an unchanged persisted message uses its persisted `updated_at`, while a changed or
new message persists and uses this single `transactionTime` as `updated_at`. UUID is the final
deterministic tie-break, so valid drafts with both provider timestamps nil remain admissible. The
dedicated post-sync test plan owns the time and tie-break coverage; until then, exact-scope static
source review is the only gate and executable evidence is `NOT_RUN`.

- [ ] **Step 9: Implement Task 6C2 transaction, outbox, and same-lease proof**

**Files:**
- Create: `backend/internal/mail/mailbox_ent_sync_apply.go`

After 6W, 6Q2, 6C0V2, and 6C1, this is the sole consumer of both implementation handoffs and the sole
owner of the repository transaction, outbox append, and same-lease proof. It re-locks and compares
account, lease, sync, mutation, phase, cursor, base cursor, and checkpoint before any projection
write. In one transaction, it obtains the authoritative transaction time, invokes the bounded
projection handoff, persists position/counts/outbox result, and writes
`sync_lease_applied_next_at = NextAttemptAt` plus checked exact-second
`sync_lease_until = NextAttemptAt + 1s`. It obtains the 6Q2 database wall clock only after the
successful account lock, once, and uses it for every live-lease/applied-at/projection/schedule value;
before proof write `NextAttemptAt` is strictly after that clock and its checked `+1s` value remains
strictly later. The latter is the minimum grace required by Task 6W; a new claim is permitted after
lease expiry.

A matching same-lease proof replays without writes or re-applying projection/event candidates. A
pre-existing same-lease proof whose hash, hash version, phase, next attempt, or position differs
from this canonical request returns the existing content-free invalid repository error with zero
writes; it is never overwritten or re-applied. Expose only the existing
`MailboxSyncRepository.ApplyMailboxSyncPage` ABI and perform no network I/O.

- [ ] **Step 10: Keep provider I/O outside transactions**

Repository methods only persist state. `MailboxSyncService.Step` remains the sole orchestration point around one bounded `MailEngine` call.

- [ ] **Step 11: Confirm worker selection and fault reschedule**

Select due rows by `due_at, account_id` with bounded batch and skip-locked semantics. Fault reschedule verifies the selected snapshot, advances attempts once, caps delay, and loses safely on a changed lease/fence.
The worker maps `MailboxSyncWorkerDueAccount.UpdatedAt` from `sync_schedule_updated_at`, and every
schedule-affecting mutation advances that marker from the same repository clock read.

- [ ] **Step 12: Static review and commit each atomic packet**

```text
feat(mail): persist mailbox sync and worker state
```

### Task 7A: Stalwart MailEngine adapter

**Files:**
- Create: `backend/internal/mail/stalwart/engine.go`
- Create: `backend/internal/mail/stalwart/client.go`
- Modify: `backend/internal/mail/config.go`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: the full `MailEngine` interface and accepted Phase 145 endpoint/auth contract.
- Produces: `stalwart.NewEngine(Config, *http.Client) (*Engine, error)` and `var _ mail.MailEngine = (*Engine)(nil)`.

- [ ] **Step 1: Implement transport and authentication**

Use separate management and JMAP base URLs/credentials, bounded request/response bodies, context deadlines, safe error mapping, and no provider text in logs/public errors.

- [ ] **Step 2: Implement the full engine surface**

Implement every current method, grouped exactly as follows: control (`Capabilities`, `GetDomain`, `EnsureDomain`, `EnsureAccount`); folders (`ListFolders`, `CreateFolder`, `RenameFolder`, `DeleteFolder`); projection/read (`ListMessages`, `GetMessage`, `GetBlob`, `Changes`, `EnumerateProjection`); drafts/submission (`SaveDraft`, `ReconcileDraft`, `Submit`, `ReconcileSubmission`); mutations (`MoveMessages`, `SetFlags`, `SetKeywords`, `DeleteMessages`); rules/training/health (`GetSieve`, `UpsertSieve`, `ClassifyMessages`, `Health`). On projection/read only, this future adapter is the sole source allowed to derive `EngineMessageProjection.CalendarMethod` through bounded server-side MIME/calendar parsing, then validate, copy, and budget it. It never takes the method from provider event JSON, folders, persisted rows, or caller input. Unsupported provider capability returns the existing typed unavailable/capability error and never silently succeeds.

- [ ] **Step 3: Record deferred HTTP-boundary coverage**

The dedicated post-sync test plan owns isolated HTTP-boundary coverage for request encoding, response bounds, cancellation/deadline, non-2xx safe mapping, malformed/oversized JSON, separate credentials, and redacted errors. Until then this packet writes no test and exact-scope static source review is the only gate; executable evidence is `NOT_RUN`.

- [ ] **Step 4: Static review and commit**

```text
feat(mail): add Stalwart engine adapter
```

### Task 7B: Mail account HTTP surface

**Files:**
- Create: `backend/internal/mail/mailbox_handler.go`
- Modify: `docs/mail/openapi.yaml`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: `MailboxService.ResolveOwner` only.
- Produces: authenticated `GET /api/mail/account` matching the existing `/mail/account` OpenAPI operation.

- [ ] **Step 1: Add the handler**

Resolve actor exclusively from auth context, call `ResolveOwner`, and return the existing bounded owner/account/projection/recovery representation. Map invalid auth, not configured, not found, suspended, deleting, and unavailable states to the frozen status and safe-code contract.

- [ ] **Step 2: Record deferred handler coverage**

The dedicated post-sync test plan owns fake-owner coverage for missing/invalid actor, each owner state, typed service errors, no secret/provider identifiers, and exact route registration. Until then this packet writes no test and exact-scope static source review is the only gate; executable evidence is `NOT_RUN`.

- [ ] **Step 3: Static review and commit**

```text
feat(mail): expose mailbox account status
```

All other `/api/mail` OpenAPI operations remain unregistered until their own accepted services and receipt-backed repositories exist.

### Task 7C: Canonical Calendar mutation HTTP surface

**Files:**
- Create: `backend/internal/calendar/event_handler.go`
- Modify: `docs/mail/openapi.yaml`
- Modify: `docs/mail/QA-HANDOFF.md`

**Interfaces:**
- Consumes: the six accepted actor-facing `EventService` methods.
- Produces: authenticated `POST /api/calendar/v2/events`, `PATCH|DELETE /api/calendar/v2/events/{eventId}`, `POST /api/calendar/v2/events/{eventId}/rsvp`, `POST /api/calendar/v2/events/{eventId}/instances/{recurrenceId}/rsvp`, and `DELETE /api/calendar/v2/events/{eventId}/instances/{recurrenceId}`.

- [ ] **Step 1: Implement exact route-to-service mapping**

Map routes respectively to `CreateCalendarEvent`, `UpdateCalendarEvent`, `CancelCalendarEvent`, `RespondToCalendarSeries`, `RespondToCalendarOccurrence`, and `CancelCalendarEventOccurrence`. Both RSVP POST routes return the canonical `CalendarAttendee`, never a full event or link state; their caller invalidates and refetches the event/card. Actor comes only from auth context; current owner/state is never accepted from the client.

- [ ] **Step 2: Record deferred handler coverage**

The dedicated post-sync test plan owns coverage for bounded DTO parsing, UUID/recurrence/version validation, typed safe status mapping, foreign-resource masking, canonical `CalendarAttendee` RSVP result encoding, caller invalidation/refetch, and exact POST route registration. Until then this packet writes no test and exact-scope static source review is the only gate; executable evidence is `NOT_RUN`. Keep legacy GET routes in the existing handler.

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

- [ ] **Step 5: Exact-scope static review and commit**

Docker verification and all test coverage are deferred to the dedicated post-sync test plan. Until that framework cutover, this production-source packet has only exact-scope static source review and all executable evidence is `NOT_RUN`.

```text
feat(calendar): consume inbound mail invitations
```

### Task 11: Calendar-to-Mail delivery consumer

**Files:**
- Create: `backend/internal/mail/calendar_delivery_consumer.go`
- Create: `backend/internal/mail/calendar_reply_worker.go`
- Create: `backend/internal/mail/mailbox_delivery_repository.go`
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

- [ ] **Step 4: Exact-scope static review and commit**

Docker verification and all test coverage are deferred to the dedicated post-sync test plan. Until that framework cutover, this production-source packet has only exact-scope static source review and all executable evidence is `NOT_RUN`.

```text
feat(mail): deliver calendar invitations and replies
```

### Task 11B: Bounded ScheduledCall backfill

**Files:**
- Create: `backend/internal/calendar/scheduled_call_backfill.go`
- Create: `backend/cmd/calendar-backfill/main.go`
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

- [ ] **Step 5: Exact-scope static review and commit**

The dedicated post-sync test plan owns dry-run, apply, idempotent-rerun, postcondition, and failure coverage through the then-current accepted launcher. Until that framework cutover, this production-source packet writes no test, uses exact-scope static source review only, and every executable claim is `NOT_RUN`.

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
