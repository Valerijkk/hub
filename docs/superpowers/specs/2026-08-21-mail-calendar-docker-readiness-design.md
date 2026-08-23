# Mail + Calendar Docker readiness design

> **P3-D0 static acceptance (2026-08-23):** immutable source `47faf3a705e5641f4ceb7ee365c3dfeeeace725d` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/ent/schema/calendarinboundmessagereceipt.go`. It adds required immutable `organization_id`/`mailbox_id` UUID scalars and database nonzero checks, with no Ent edges or raw data; existing uniqueness, state-shape, and due behavior are unchanged. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. The D0-next route is `CONSUMED/HISTORICAL`. P3-C remains blocked on Mail Ent generation; 147-02B/147-02C and D1/D2 remain blocked on active Calendar generation/dependencies; P3-E is conjunction-blocked. No next unblocked P3 source is selected. Generation, migration, SQL, build, tests, Docker, and runtime are `NOT_RUN`; no readiness, merge, or product claim.

**Date:** 2026-08-21
**Branch:** `codex/mail-module-20260813`
**Status:** approved by founder direction for autonomous implementation; no merge authority
**Scope:** Phases 143–148 completion plus the canonical QA launcher and migration lineage needed to verify them

## Goal

Deliver a maintainable Mail and Calendar vertical that can be generated, migrated, composed, and
verified in task-owned Docker containers before this branch is considered for merge. Static source
acceptance is necessary but is not runtime evidence.

The finish line is one reproducible Docker evidence set on one exact commit:

- generated Ent code compiles without a committed generated tree;
- versioned Mail and Calendar migrations pass fresh-database and upgrade-fixture checks;
- canonical Calendar and Mail services use concrete transactional repositories;
- the server starts the required workers and exposes health/readiness;
- contract, API, integration, and smoke tests have non-zero collection and zero unexpected skips;
- all task-owned containers, networks, volumes, and reports are accounted for and cleaned;
- no merge, push, deploy, or shared/production migration is performed.

## Current constraints

1. `testframework/runner/run-task.ps1` is a reviewed private supervisor, but the public launcher,
   local image lock, admitted dependency assets, and executable target images are not ready.
2. The repository contains Ent schemas and `backend/ent/generate.go`, but generated Ent files are
   ignored and absent.
3. Production comments require Atlas, while the repository has no live `atlas.hcl`, migration
   directory contract, or bounded migration target.
4. Calendar create/update/occurrence and call policies exist behind abstract ports. Mail sync service
   and worker exist behind abstract ports. Concrete persistence, composition, and most public APIs
   do not exist.
5. Windows-host tests, generators, builds, linters, validators, and application processes are
   forbidden. Direct Docker/Compose is not a fallback for `/qa-gate`.
6. `pro-qa-ai/**` is outside scope and must not be read or changed.

## Architectural decisions

### D1. One public launcher, one private supervisor

Add one closed public entrypoint under `testframework/runner/**`. The caller supplies exactly
`target`, `task`, and `run`; the launcher selects the contract and target-catalog revisions only
from committed internal authority, performs side-effect-free admission, and delegates to the
private supervisor. It exposes no caller-controlled contract, catalog, command, path, Compose,
image, environment, timeout, or cleanup override.

The private runner remains the sole owner of image inspection, fixed-target execution, evidence,
timeouts, and exact cleanup. No alternate launcher or direct Compose command is documented.

### D2. Portable worktree-safe roots

Machine-specific `C:\Users\Valerii\Desktop\hub` paths are replaced by roots derived from the
canonical repository containing the launcher script. Inputs cannot redirect cache, ingress, or
reports elsewhere. Resolved roots must be regular non-reparse paths contained in the selected
worktree, except Docker's fixed installed client path.

This preserves task isolation and makes the launcher reusable from any linked Hub worktree.

### D3. Commit-materialized, offline target images

Target images are built from an exact committed Git tree and admitted local dependency assets. The
runtime container has no checkout bind, source bind, Git inference, package installation, registry
access, or network fallback. Base identities remain complete local `linux/amd64` image IDs and
reviewed digests.

Dependency admission may consume only dependencies already pinned by `go.sum`,
`package-lock.json`, or a reviewed tool lock. Missing bytes yield `NOT_READY`; they are not silently
downloaded during a QA run.

### D4. Closed, extensible target catalog

The runner does not accept arbitrary commands. A versioned catalog enumerates fixed image-baked
targets. Initial accepted targets remain:

- `backend-unit`;
- `frontend-unit`;
- `backend-openapi`.

Later reviewed catalog revisions add, in order:

- `python-selftest` and `python-contract`;
- `backend-migration`;
- `calendar-integration`;
- `calendar-api`;
- `mail-integration`;
- `mail-api` and `frontend-unit` Mail coverage;
- `mail-calendar-smoke`.

Every new target has a pinned image, fixed entrypoint, resource limits, deadline, evidence schema,
and cleanup policy. A catalog revision is an authority change and receives independent review.

### D5. Ent generation is image-local

`go generate ./ent` runs during the backend image build from the exact source commit, before
`build`, `vet`, units, or integration binaries. Generated files are inspected and compiled inside
the image but remain uncommitted. A deterministic double-generation oracle proves the second run
has no generated diff.

Application code never hand-writes generated Ent files.

### D6. Atlas owns versioned database migration

Use an explicitly admitted Atlas CLI whose exact version equals the current indirect module pin.
The CLI module/version/sums, image-local build path, binary digest, `atlas version` output, target
architecture, and containing image ID are recorded in a reviewed tool lock; an indirect `go.mod`
entry alone is not executable-tool provenance. Add one repository-owned `backend/atlas.hcl`, one
versioned pre-feature baseline fixture produced from an exact accepted parent schema, and one
ordered Mail/Calendar migration directory. SQL files are immutable once accepted; corrections are
forward-only migrations.

This is deliberately a bounded feature lineage, not a false claim that the new directory owns all
historical Hub migrations. A fresh target database first loads the immutable full pre-feature
baseline fixture, then applies the Mail/Calendar migration directory. Upgrade tests load the same
baseline plus realistic legacy rows. Ent comparison is allowlisted to the Mail/Calendar tables and
the shared columns, edges, indexes, and constraints those phases change (including
`scheduled_calls`); unrelated pre-feature tables are required to remain byte/schema equivalent but
are not re-authored by this packet.

The migration target performs:

1. lint/checksum verification;
2. load the exact pre-feature baseline into an empty PostgreSQL database and apply the feature
   lineage;
3. apply the same lineage to a versioned realistic pre-Mail/Calendar fixture;
4. schema/index/constraint comparison against generated Ent metadata;
5. a negative/failure fixture proving the transaction fails closed;
6. forward-recovery rehearsal without deleting user data.

Shared or production apply remains a separate founder-reviewed operation and is not authorized by
this branch.

### D7. Legacy recurring ScheduledCall backfill is fail-closed

There is no trustworthy original local `DTSTART` or timezone for an already persisted recurring
`ScheduledCall`. Automatic reinterpretation is forbidden.

- non-recurring, coherent rows may be migrated deterministically;
- any non-deleted recurring legacy row aborts automatic backfill and is reported without mutation;
- the operator must recreate/map that series through a later explicit per-row artifact before
  cutover;
- clean databases and fixtures prove both the successful no-recurring path and the blocking
  recurring path;
- no historical invitation, notification, NATS, or outbox event is emitted by backfill.

This rule is simpler to operate and safer to extend than maintaining heuristic recurrence parity.

### D8. Domain policy stays separate from persistence and composition

Existing pure Calendar and Mail policies remain free of Ent, SQL, clocks, network, and global
state. New concrete adapters implement their frozen ports in separate files and own transaction
ordering, locks, compare-and-swap predicates, outbox writes, and persistence echoes.

Server wiring owns dependency construction, worker lifecycle, health/readiness, and graceful
shutdown. Handlers own transport mapping only. This separation is mandatory for future engine,
storage, and API evolution.

### D9. Vertical completion order

Calendar:

1. Ent generation and migrations;
2. concrete event transaction adapter;
3. RSVP transaction/service;
4. canonical reads and `/calendar/v2/**` handlers;
5. event-to-call, occurrence-cancel, and due-job transaction bridges;
6. server wiring and legacy compatibility cutover.

Mail:

1. Ent generation and migrations;
2. concrete sync and worker repositories;
3. real MailEngine adapter with explicit authentication/DNS readiness;
4. provisioning/control-plane services;
5. owner mailbox services and handlers;
6. worker composition, health/readiness, outbox/storage/search/realtime seams.

No public route is enabled before its persistence and authorization path is integration-tested.

### D10. Test layering

The Python package, framework selftests, and closed `python-contract`, Calendar integration/API,
and Mail integration/API targets are accepted before the first persistence or route packet that
depends on them. Every such implementation packet authors its failing contract/integration/API
oracle first and can be verified immediately through the already accepted launcher; Task ordering
never permits direct pytest, Go, or Compose fallback.

- `backend/tests/{calendar,mail}/**`: genuine isolated Go units.
- `testframework/tests/contract/**`: OpenAPI, event, migration, source, and generated-schema
  contracts.
- `testframework/tests/api/**`: public status/body/header/authz/idempotency behavior.
- `testframework/tests/integration/**`: PostgreSQL, NATS, storage, engine, workers, races, rollback,
  and recovery.
- `testframework/tests/e2e/**`: bounded user journeys; `smoke` is a marker.

Expected values are literal or independently derived. Tests do not assert mocks merely exist and do
not grep implementation text as a behavior oracle.

### D11. Phase 143–148 traceability

| Phase | Deliverable in this program | Docker evidence | Completion boundary |
|---|---|---|---|
| 143 | Calendar schema migration, concrete transactional repositories, RSVP/v2 API, ScheduledCall bridges | `backend-migration`, `calendar-integration`, `calendar-api`, `mail-calendar-smoke` | Canonical events/RSVP/call compatibility green; no heuristic recurring-row backfill |
| 144 | Mail schemas, concrete sync/control/product repositories and services, owner API | `backend-migration`, `mail-integration`, `mail-api`, `mail-calendar-smoke` | Engine-neutral domain is persisted, authorized, recoverable, and wired |
| 145 | Admitted Stalwart image/license/provenance, real engine/provisioning adapters, bounded persistent worker profile | `mail-integration`, `mail-calendar-smoke` plus measured resource receipt | External delivery remains disabled until DNS/PTR/network gates; local persistent runtime is green |
| 146 | Responsive accessible Mail workspace, invitation/compose/conflict/degraded-state behavior | `frontend-unit`, `mail-api`, `mail-calendar-smoke` | Desktop/tablet/mobile journeys and honest recovery states are green |
| 147 | Recurrence/import/recovery, busy/free + leave, reminders, feeds/CRM seams or explicit bounded deferrals | `python-contract`, both integration targets, `mail-calendar-smoke` | Every roadmap criterion is implemented and green or recorded as an explicit founder-visible deferral that prevents phase-complete claims |
| 148 | Current-SHA security/recovery/docs/QA reconciliation | all selected targets on one exact commit | Clean merge-ready branch and evidence bundle; no rebase/merge/push until founder signal |

No phase is marked complete merely because an underlying packet is statically accepted. Phase 148
cannot pass while any non-deferred Phase 143–147 criterion lacks executable evidence.

## Failure and recovery rules

- Unknown target, malformed input, missing image/dependency, zero collection, unexpected skip,
  report collision, leaked resource, or cleanup ambiguity fails closed.
- Database writes use bounded transactions and content-free public errors.
- Worker failures never expose provider bodies, cursors, secrets, raw mail, or panic values.
- Migration rollback before apply is removal of the unapplied artifact. After apply, recovery is a
  feature disable plus reviewed forward migration; destructive automatic rollback is forbidden.
- Runtime evidence is tied to source SHA, control SHA, image ID, target, resource limits, duration,
  counts, artifact path, and cleanup result.

## Completion and merge boundary

The branch is only merge-ready after all selected Docker targets pass on one exact final revision,
independent code/security/lifecycle review passes, documentation is synchronized, and the worktree
is clean. Because this work introduces database migrations and QA/deploy infrastructure, the final
merge still requires founder review. This design grants no merge, push, deploy, or production
apply authority.
