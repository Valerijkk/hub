# Phase 151 — Messenger Channels, Threads, and Unread Foundation

**Status:** `FROZEN_FOR_PLANNING`
**Scope:** Planning contract only. This document makes no runtime, test, Docker, migration, or delivery claim.

## Objective

Define the tenant-safe behavioural and API foundation for Messenger channels, one-level message threads, durable follow state, and server-authoritative unread state. Phase 151 supplies contracts that later implementation plans may realize; it does not include retention/archive delivery (Phase 153), broad Phase 153 work, Jitsi, Mail, Calendar, or `pro-qa-ai/**`.

## Frozen product rules

| Area | Decision |
|---|---|
| Channel visibility | Authenticated users of the same tenant may read metadata and history of a public channel before joining. Posting requires membership. Private channels are member-only for metadata, history, thread resources, follow state, unread state, and mutation. |
| Project channels | Project linkage is explicit opt-in through `project_id`; public/private visibility remains as declared by the channel. A project-linked channel is capped by active `ProjectMember` status. Inactive/non-member project users cannot obtain disclosure through that linkage. |
| Existing groups | `group` remains the current group-DM semantic. Phase 151 creates no new DM UX or reclassification. |
| Posting authority | Read-only is an orthogonal channel mode: ordinary members cannot post, while channel owners and admins may post. |
| Threads | A live root belongs to the same channel as every reply. Only root-reply is permitted: no reply-to-reply, cross-channel reply, or reply to a deleted root. Default channel timeline excludes replies. A reply with explicit `also_send_to_channel=true` remains one document and is intentionally visible in both the thread and channel timeline. |
| Follow | Automatically follow a root author, a permitted replier, and a resolved permitted mention. A manual unfollow suppresses later auto-follow until that user manually follows again. Lost access blocks both disclosure and state mutation. |
| Unread | The server persists forward-only position `(created_at, message_id)`; Redis/cache is never the authority. Joining a channel and a cutover establish the then-current-head baseline rather than retroactively marking history unread. Categories are independent and may overlap. |
| Reconnect | Client subscribes first, then refetches server state. It performs no local unread arithmetic. |
| Durable storage | PostgreSQL/Ent remains the authority for tenant, channel, membership, project membership, and roles. MongoDB stores messages plus durable `read_states` and `thread_follows`; Redis is not authoritative. Phase 151 adds no Ent schema, generated code, or migration. Every Mongo message, state, or count read/write is preceded by fresh PostgreSQL authorization. |
| Recovery | The original channel/thread/unread feature slice introduces no transactional outbox. The bounded 2026-08-30 amendment below permits the existing platform outbox only for canonical/recovery delivery types; residual delivery risk remains explicit. |

## Resource and capability contract

PostgreSQL authorization is resolved before any Mongo read-state, follow-state, message, or count lookup and before any response. Authorization loss therefore remains non-disclosive even if best-effort Mongo cleanup is delayed. Mongo cursor and follow updates are idempotent conditional upserts; they never persist counters or move a tuple backward.

Every recipient-facing channel DTO must provide the capability fields required to render actions without client inference. At minimum this includes immutable identifiers, tenant-scoped visibility, `project_id` where opted in, effective membership/active-project-membership state, effective read-only state, and booleans for read, post, reply, follow/unfollow, and unread-category visibility. Absence of a capability is denial; a client must not derive permission from role labels, cached history, or a project id.

Thread responses must name the exact parent root resource and the thread resource, each with a channel identity. A reply returned through a timeline, root, or thread endpoint must be demonstrably bound to that same channel and root. Deleted/inaccessible parent state must not be replaced with an observable surrogate that leaks identity, existence, membership, timestamp, follow, or unread data.

## Unread position and categories

Each unread category has an independently persisted position that advances only forward by the ordered tuple `(created_at, message_id)`; ties are resolved by `message_id`. Every position is tenant/user/category scoped plus its category target: workspace channel, DM/group-DM conversation, or followed root. It can be safely refetched after disconnect or process restart. A join/cutover baseline is the current accessible head of that target at that moment, not a read of earlier history.

The four categories below are independently computed and may overlap for one message; clearing or advancing one category must not silently clear another category unless its own rule selects the same message.

| Category | Meaning |
|---|---|
| `channel` | New accessible workspace-channel timeline messages after the channel baseline/position. |
| `dm` | New accessible direct-message or group-DM timeline messages after the DM baseline/position. |
| `mention` | New accessible resolved mentions of the recipient. |
| `followed_thread` | New accessible replies on followed eligible roots after the thread baseline/position. |

Exact API field names, Mongo document fields and indexes, event topic names, and category-clear endpoint design remain implementation-plan decisions; the storage authority and PostgreSQL-before-Mongo authorization boundary above are frozen.

## Requirement IDs and acceptance criteria

| ID | Requirement | Measurable acceptance evidence required before implementation acceptance |
|---|---|---|
| `MSG151-CHN-01` | Enforce tenant, public/private, membership, project opt-in, active `ProjectMember`, and read-only/post rules. | A permission matrix covers public pre-join read versus post denial, private non-member non-disclosure, active/inactive project membership, cross-tenant denial, and ordinary-member read-only post denial and owner/admin read-only post permission. No endpoint/event response exposes forbidden metadata or history. |
| `MSG151-CHN-02` | Return recipient DTO and capabilities as the authority for UI actions. | Contract assertions show allowed and denied capabilities for the matrix above; frontend rendering has no permission derivation from roles or cached data. |
| `MSG151-THR-01` | Permit only live same-channel root replies. | Negative cases reject reply-to-reply, cross-channel parent, and deleted-root parent; a valid root reply returns its exact parent and thread identity. |
| `MSG151-THR-02` | Keep replies out of default channel timeline and represent `also_send_to_channel` deliberately. | Timeline assertions exclude ordinary replies; a reply with `also_send_to_channel=true` is returned in the thread and timeline exactly as one document identity, with no duplicate unread arithmetic. |
| `MSG151-FOL-01` | Maintain automatic follow and manual-unfollow suppression. | Scenario sequence verifies root author/replier/resolved mention auto-follow, manual unfollow, continued suppressing auto triggers, and manual follow restoring eligibility. |
| `MSG151-FOL-02` | Prevent follow-related disclosure after access loss. | Access revocation scenarios deny read, mutation, follow state, and unread/thread information, including reconnect/refetch paths. |
| `MSG151-UNR-01` | Persist forward-only authoritative positions and baselines. | Equal-time ordering tests prove `(created_at,message_id)` ordering and independent persisted positions for every category target; restart/refetch returns persisted server state; join and cutover start at current target head without historical unread backfill; Redis loss cannot alter truth. |
| `MSG151-UNR-02` | Preserve four independent, overlapping unread categories. | Matrix includes a message qualifying for multiple categories and proves category-specific advancement does not clear another qualifying category. |
| `MSG151-STO-01` | Enforce the PostgreSQL authorization boundary and Mongo durable-state semantics. | Rejected, stale, or revoked requests perform zero Mongo message/state/count lookup or mutation. Allowed duplicate/concurrent conditional upserts are idempotent and cannot regress `(created_at,message_id)`. No unread counter is persisted, and static acceptance confirms no Phase 151 Ent schema, generated code, or migration. |
| `MSG151-RT-01` | Reconnect safely with subscribe-then-refetch and no client arithmetic. | Client/integration evidence records subscribe before refetch and confirms rendered counts come from refetched server response after reconnect. |
| `MSG151-RT-02` | Preserve recovery semantics without pretending an outbox exists. | Failure/retry evidence shows state can persist/refetch safely; documentation records that event delivery is not transactionally coupled and therefore may require refetch/reconciliation. |

## Frontend acceptance contract

The frontend must render a Browse path for public channels and project-channel opt-in, send mutations only when recipient capabilities permit them, render category badges without assuming exclusivity, and recover via subscribe-then-refetch. Opening a thread uses the native right side-panel thread surface and keeps the channel timeline context available. Keyboard focus order, visible focus treatment, accessible names/status announcements, focus return, and responsive desktop/mobile layouts must remain usable for Browse, channel, side-panel thread, badges, and follow controls. Accessibility and responsive assertions belong to later executable acceptance; no such evidence is claimed here.

## Integration and residuals

Phase 149 integration is a planning dependency/residual only: `NOT_RUN`. The Phase 151 plan must explicitly bind its resource/capability and reconnect assumptions to the Phase 149 boundary before implementation acceptance. Archive/retention delivery is deferred to Phase 153; Phase 151 must only preserve no-leak behaviour when archived/retained/deleted state is encountered. Until bounded MPR-05A-D lands, the original feature slice still lacks transactional outbox delivery; after it lands, delivery remains at-least-once and never implies exactly-once.

## 2026-08-30 bounded recovery-runtime amendment

This tracked amendment is controlled by `.planning/decisions/ADR-COMMS-OSS-REPLACEMENT-2026-08-24.md` and overrides only the Phase-151 recovery-runtime storage/delivery prohibitions above. All channel, thread, unread, follow, authorization, tenant, no-leak and protected-path rules remain unchanged.

- The original feature tasks still add no Ent schema, generated code or migration. Only the separately frozen MPR-03A/B and MPR-05A recovery sub-atoms may add `MessengerRecoveryLease`, `MessengerPostCreateDelivery`, the optional unique OutboxEvent `delivery_key`, their reviewed expand migrations, and the mechanically frozen generated Ent manifest.
- Recovery uses one global PostgreSQL leader lease/fence/checkpoint plus the independent Mongo per-message owner/fence/lease. Neither guard grants user/channel authority or replaces fresh PostgreSQL reauthorization.
- Mongo completion writes a replayable delivery intent. Current audience is re-resolved before creating a per-recipient delivery plan. A PostgreSQL materializer writes the deterministic delivery ledger row and platform `OutboxEvent` in one transaction.
- For canonical/recovery delivery types, the existing outbox worker is the sole NATS publisher. The isolated legacy publisher may coexist only before a tenant cutover, never for the same request, and must retire before default canonical enablement.
- Delivery is at-least-once. Cross-store atomicity and exactly-once are forbidden claims; unique delivery keys, idempotent consumers, checkpointing and refetch/reconciliation close the supported recovery model.
- The dedicated recovery worker is a separate self-hosted process, not a `cmd/server` goroutine. Writer cutover is allowlisted per tenant with exactly one writer; rollback changes future writes only and drains already canonical intents/outbox rows.

Additional acceptance requirements:

| ID | Requirement | Evidence boundary |
|---|---|---|
| `MSG151-REC-01` | Fence global recovery ownership and each Mongo message independently. | Focused Docker RED/GREEN units for MPR-01 and later lease adapters; live PostgreSQL/Mongo contention remains `NOT_RUN` until Python integration. |
| `MSG151-REC-02` | Recover Mongo completion through current-audience planning and one PostgreSQL ledger/outbox transaction. | Transaction/idempotency units plus later crash-boundary, NATS outage, revoke-before-fanout and rollback integration matrix. |

This amendment is planning authority only. Runtime, Docker, migration application, live PostgreSQL/Mongo, NATS, WebSocket, browser, canary and rollback evidence remain `UNBOUND/NOT_RUN`.

## 2026-08-31 MPR-02 scheduler contract

MPR-02 is a deterministic process-local scheduling controller over exactly one opaque attempt. It accepts only a closed immutable result kind: `standby`, `success`, `retryable_failure` or `leadership_lost`. It receives no repository, raw error, owner, lease, fence, checkpoint, tenant, channel or message fact. MPR-06 later owns the PostgreSQL leadership lifecycle plus recovery executor and maps acquisition denial to `standby`; MPR-07 later owns configuration binding, process wiring, health/readiness and metrics.

All waiting uses injected one-shot timers and one injected initial-jitter source. Frozen policy is run timeout 20s, initial jitter at most 5s once, normal cycle 30s, saturating retry `1,2,4,8,16,30...` seconds, and shutdown grace 10s. One scheduler instance is single-use and permits at most one active caller, one cadence timer and one attempt; while the attempt runs, no cadence timer exists and at most one timeout/grace control timer is armed. Parent cancellation wins every ready event, leadership loss wins run timeout, run timeout wins other simultaneous completion, and selected grace expiry wins a simultaneous completion. A timed-out or cancelled attempt receives cancellation and must return within grace; otherwise the scheduler returns a fixed shutdown-timeout sentinel, schedules nothing further and makes no claim that the external attempt stopped.

The result/policy constructors have exact invalid sentinel behavior and the generic immutable policy requires positive durations, base no greater than cap and an integer multiplier of at least two; MPR-07 later binds the frozen operational values. A stable read-only Done signal closes only after Run is terminal and every attempt goroutine exits. Future lifecycle cannot replace a timed-out scheduler in-process while Done remains open. No ticker, queued catch-up, overlapping attempt, real-time sleep, raw-error observation, live-store behavior or distributed-ownership claim is allowed.

Acceptance is one external fake-clock oracle after accepted MPR-01. It proves the frozen exported API/sentinels, invalid/nil/typed-nil rejection, jitter bounds, no attempt before the initial tick, finish-anchored cadence, standby/success reset, capped backoff, observable late completion, causally ordered cases around each cancellation/deadline/grace priority, non-cooperative shutdown plus delayed Done closure, leadership-loss termination, stale-timer rejection, single-use `Run` and zero successor activity after return. Literal simultaneous select readiness and private channel capacity remain code-review evidence because no test-only synchronization hook is allowed. This is unit/source planning only; MPR-02 production, Docker RED/GREEN and runtime remain `ABSENT/NOT_RUN`.

The ignored external-oracle draft at `.superpowers/sdd/151-PLAN/phase151_pending_recovery_scheduler_test.go.draft` may be `STATIC_ACCEPTED` as source-contract review only and remains outside test discovery. Its accepted hash is `BC03B8CF8019102CFA74204BA7637BE6E8773C95162913D04C365394D45EFAC3`; MPR-02 source, Docker RED/GREEN, distributed-store behavior and runtime remain `ABSENT/NOT_RUN`.

## 2026-08-31 MPR-03A global recovery-leader schema-source contract

Future MPR-03A owns only new `backend/ent/schema/messengerrecoverylease.go` and external descriptor oracle `backend/tests/messenger/phase151_recovery_leader_schema_test.go`. The designated global row uses the Hub UUID-id convention plus an immutable unique `lease_key` whose default and only validator-accepted value is exact `global`; mutable state is nullable holder UUID, non-negative fence default zero, nullable lease expiry and a nullable scalar checkpoint triple `(tenant UUID, channel UUID, non-empty valid-UTF-8 kind)`. It intentionally has no audit timestamps, version mixin, edges, foreign keys, JSON checkpoint or kind enum.

Checkpoint kind remains forward-compatible rather than schema-closed to `channel|dm`; its schema-source UTF-8 validator represents only the durable text boundary. Cross-column integrity, enforced singleton cardinality and seed row belong to MPR-03B migration after accepted MPR-03A and the separate generator gate. MPR-03C then owns row-evaluation-time `clock_timestamp()` lease predicates, monotonic fencing, fail-closed seed/duration handling and collation-independent comparison matching the admitted generic cursor's UUID-raw-byte/UUID-raw-byte/kind-raw-byte order. All-null is initial and lower than a full tuple; release preserves fence/checkpoint, and clearing plus partial/equal/backward checkpoint updates are forbidden. Global fence gates new process work; server-time Mongo owner+fence+expiry guards every message write. The worker emits no direct external effect: deterministic intent then ledger/outbox guard admitted replay and publisher delivery remains at-least-once. Cross-store atomicity remains forbidden and its handoff race is MPR-09 evidence. Lifecycle is `FROZEN_NONDISPATCHABLE`, dependency state is `BLOCKED_MPR01_AND_SOLE_ORACLE_QUEUE`, and the exact prefix is `MPR-01 -> JITSI-02A -> CALL-DUR-01 -> CALL-DUR-03 -> MPR-03A`; schema/oracle, Docker, generated code, migration, adapter, live PostgreSQL and runtime are `ABSENT/NOT_RUN/UNBOUND`.
