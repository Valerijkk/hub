# 144-04G sync-service source-authority readiness

## 144-04G receipt-coherence sync candidate (2026-08-22)

> **Coherence gate:** `COHERENCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-self-activating dirty exact12 docs-only coherence candidate records immutable source `3031c53d399accc84fe10cd0c5421a85c089ac7c` and prior acceptance receipt `2a2975056c3fb8001d88863a041310271282989d`; their full metadata, blobs and ordered manifests are in `IMPLEMENTATION-144-04G.md`. Source Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Only after an atomic manifest-equal exact12 coherence commit and immutable coherence C/S/L triple PASS is 144-04G `STATIC_ACCEPTED / EXACT_SHA_PASS`; former 04G authority/acceptance/source-unimplemented/current-next routes then become `CONSUMED/HISTORICAL`. Historical coherence selects no next Mail packet and grants no runtime/downstream authority. 04F remains accepted; 04H/02E, Ent/SQL/concrete adapter, generation/migration, API/UI, launcher/testframework and runtime remain blocked `NOT_READY`/`NOT_RUN`.

## Source-authority acceptance-sync candidate (2026-08-20; predecessor route)

Status: `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED /
SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This dirty exact9 working tree grants no
authority. The `6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` metadata and the three immutable
exact-SHA PASS results below are proposed receipt facts only; they activate only after one atomic
commit of this exact9 acceptance sync. Until activation, the lower candidate/precommit/root-commit/
exact-SHA-next lifecycle remains current and root cannot dispatch source; the separately false
accepted-repository premise remains `CONSUMED/HISTORICAL`.

After that atomic exact9 sync commit, this docs authority is
`SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` only at immutable commit
`6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` (parent
`2b6837926c91d652ff7bf30447bd482740261bd0`, tree
`bdd59cfcb31376c95f5128370a75333cad78bc14`, subject
`docs(mail): propose sync service authority`). Its exact9 blobs are:

- `.planning/STATE.md` `af3d1fe0e7f60501dc05964f227eebf6a12b7acf`;
- `.planning/decisions/specs/MAIL-SPEC.md` `0a64630a4080872395ab29de9adac4cbf566bbba`;
- `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
  `35d9b0c21e641d695c23dbac3970684cd1c7f1e6`;
- `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
  `da673fe1a669ae09a81d063e79e7b03d26b351cf`;
- `.planning/phases/144-mailbox-domain/PLAN.md`
  `d12820f246e9f6a44b321e3e4e399c2757b3726c`;
- `.planning/phases/144-mailbox-domain/REVIEW-144-04G-SYNC-SERVICE-READINESS.md`
  `a886f6f01365194888dfc41e4b98a3be452d372c`;
- `docs/mail/CONTRACT-CHANGELOG.md` `40c9ffe5eeb8396f336187015fd86eb26ab0f8da`;
- `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` `4fa416ea3e2888b4ed34fb544c0be33aedb60e5b`;
- `docs/mail/QA-HANDOFF.md` `837c060cdb156a657ccbff15cf46f1a09289649c`.

Contract, Security and Lifecycle immutable exact-SHA reviews PASS C/I/M `0/0/0`;
dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Only after that atomic
exact9 sync commit is every candidate, precommit, root-commit and exact-SHA-next clause below
`CONSUMED/HISTORICAL` only for that completed docs-authority lifecycle. Only then may root
dispatch only the frozen ordered exact3 named below.
Its source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own authored-first
review, atomic commit and immutable exact-SHA acceptance. 04F remains accepted; 04H/02E,
Ent/SQL/concrete adapter, generation, migration, OpenAPI and runtime remain blocked or outside
authority. Launcher `NOT_READY`; executables/runtime `NOT_RUN`.

Status: `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN` is `CONSUMED/HISTORICAL` predecessor evidence only. Before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, 04G has no current static, source or downstream authority.

Date: 2026-08-20
Authoring route: `gpt-5.6-terra` / `xhigh`, `dispatcher-enforced`, no fallback;
telemetry `UNKNOWN` unless exposed.

This exact document is design/readiness evidence only. It does not authorize source, tests,
commit, Ent, a concrete database adapter, generation, migration, service wiring, API/UI,
product, phase, release, Docker, or runtime work. Mail 144-04F remains independently
`STATIC_ACCEPTED / EXACT_SHA_PASS` at
`4715f939f24e5b5c2d3f80a38d204babe68bdd34`. Calendar 143-04C keeps its current exact3
dispatch state; this candidate neither owns nor changes Calendar source or authority.

## Exact lifecycle and manifest

The only future 144-04G source packet is test-first and exact3, in this order:

1. NEW `backend/tests/mail/mailbox_sync_service_test.go`;
2. NEW `backend/internal/mail/mailbox_repository.go` — interface and value types only;
3. NEW `backend/internal/mail/mailbox_sync_service.go`.

The lower-precedence predicate that there is no accepted `mailbox_repository.go` today, and the
phrase that 04G adds "minimal new sync repository signatures in accepted
`mailbox_repository.go`", are `CONSUMED/HISTORICAL`. The immutable source receipt records the
new abstract port, but grants no current authority before exact12 coherence activation. The exact3
packet is historical source receipt evidence only. It must not add Ent, SQL, generated code, a concrete adapter, storage/network
clients, worker wiring, or another path. 144-04H and 144-02E remain blocked.

The recorded exact9 lifecycle is `CONSUMED/HISTORICAL` predecessor evidence. Before the
manifest-equal exact12 coherence commit and immutable C/S/L triple PASS, implementation has no
current static, source or downstream authority. After activation, 04G is `STATIC_ACCEPTED /
EXACT_SHA_PASS`, but historical coherence still selects no next packet and grants no downstream
authority.

## Frozen Go ABI

The new interface/types-only file owns exactly the following public-in-package shapes. Integer
enums start at one; their zero value is invalid. No type below is an HTTP/public API code.

```go
type MailboxSyncPhase uint8

const (
    MailboxSyncPhaseIncremental MailboxSyncPhase = iota + 1
    MailboxSyncPhaseEnumeration
    MailboxSyncPhaseCatchUp
)

type MailboxSyncFence struct {
    AccountID     uuid.UUID
    LeaseID       uuid.UUID
    LeaseEpoch    int64
    SyncEpoch     int64
    MutationEpoch int64
    Phase         MailboxSyncPhase
    Cursor        string
    BaseCursor    string
    Checkpoint    string
}

type MailboxSyncClaimRequest struct {
    AccountID  uuid.UUID
    LeaseID    uuid.UUID
    Now        time.Time
    LeaseUntil time.Time
}

type MailboxSyncClaimResult struct {
    Claimed         bool
    Fence           MailboxSyncFence
    EngineAccountID string
    EventCutoff     time.Time
    NextAttemptAt   time.Time
}

type MailboxSyncEventCandidate struct {
    EngineMessageID string
    ReceivedEventID uuid.UUID
    ChangedEventID  uuid.UUID
    Destroyed       bool
}

type MailboxSyncPrepareRequest struct {
    Fence           MailboxSyncFence
    EngineAccountID string
    Folders         []EngineFolder
    Page            MailboxSyncPolicyPage
    EventCandidates []MailboxSyncEventCandidate
    OccurredAt      time.Time
    CorrelationID   string
}

type MailboxSyncPreparedPage struct {
    Prepared bool
    Input    MailboxSyncPolicyInput
}

type MailboxSyncApplyRequest struct {
    Fence        MailboxSyncFence
    Result       MailboxSyncPolicyResult
    Now          time.Time
    NextAttemptAt time.Time
}

type MailboxSyncApplyResult struct {
    Applied       bool
    Replayed      bool
    Phase         MailboxSyncPhase
    NextAttemptAt time.Time
}

type MailboxSyncFailureKind uint8

const (
    MailboxSyncFailureInvalidCursor MailboxSyncFailureKind = iota + 1
    MailboxSyncFailureRestartRequired
    MailboxSyncFailureTransient
    MailboxSyncFailureRejected
)

type MailboxSyncResolveRequest struct {
    Fence        MailboxSyncFence
    Failure      MailboxSyncFailureKind
    Now          time.Time
    NextAttemptAt time.Time
}

type MailboxSyncResolveResult struct {
    Resolved      bool
    Phase         MailboxSyncPhase
    NextAttemptAt time.Time
}

type MailboxSyncRepository interface {
    ClaimMailboxSyncStep(context.Context, MailboxSyncClaimRequest) (MailboxSyncClaimResult, error)
    PrepareMailboxSyncPage(context.Context, MailboxSyncPrepareRequest) (MailboxSyncPreparedPage, error)
    ApplyMailboxSyncPage(context.Context, MailboxSyncApplyRequest) (MailboxSyncApplyResult, error)
    ResolveMailboxSyncStep(context.Context, MailboxSyncResolveRequest) (MailboxSyncResolveResult, error)
}
```

The new service file owns exactly these additional shapes:

```go
type MailboxSyncServiceConfig struct {
    Now           func() time.Time
    NewLeaseID    func() uuid.UUID
    NewEventID    func() uuid.UUID
    Jitter        func(time.Duration) time.Duration
    LeaseDuration time.Duration
    PollInterval  time.Duration
    RetryInterval time.Duration
    PageSize      int
}

type MailboxSyncStepOutcome uint8

const (
    MailboxSyncStepIdle MailboxSyncStepOutcome = iota + 1
    MailboxSyncStepApplied
    MailboxSyncStepRebuildScheduled
    MailboxSyncStepRetryScheduled
    MailboxSyncStepLeaseLost
)

type MailboxSyncStepResult struct {
    Outcome       MailboxSyncStepOutcome
    NextAttemptAt time.Time
}

type MailboxSyncServiceErrorKind string

const (
    MailboxSyncServiceErrorInvalid     MailboxSyncServiceErrorKind = "invalid"
    MailboxSyncServiceErrorTooLarge    MailboxSyncServiceErrorKind = "too_large"
    MailboxSyncServiceErrorUnavailable MailboxSyncServiceErrorKind = "unavailable"
)

type MailboxSyncServiceError struct {
    Kind MailboxSyncServiceErrorKind
}

var (
    ErrMailboxSyncServiceInvalid     = errors.New("mailbox sync service invalid")
    ErrMailboxSyncServiceTooLarge    = errors.New("mailbox sync service exceeds bounds")
    ErrMailboxSyncServiceUnavailable = errors.New("mailbox sync service unavailable")
)

type MailboxSyncService struct { /* exact unexported engine, repository, config fields */ }

func NewMailboxSyncService(MailEngine, MailboxSyncRepository, MailboxSyncServiceConfig) (*MailboxSyncService, error)
func (*MailboxSyncService) Step(context.Context, uuid.UUID) (MailboxSyncStepResult, error)
func (*MailboxSyncServiceError) Error() string
func (*MailboxSyncServiceError) Unwrap() error
```

`Error()` returns only the matching sentinel text; nil/unknown returns the literal
`mailbox sync service error`. `Unwrap()` returns only the matching sentinel, otherwise nil. There is no cause,
cursor, provider text, account, address, subject, secret, token, body, or arbitrary detail in an
error or step result. Raw engine/repository errors are never returned or logged by this packet.

## Constructor, bounds and deterministic inputs

The constructor rejects nil engine/repository/functions, non-positive durations, `PageSize`
outside `1..MaxEnginePageItems`, and any config combination that can overflow time arithmetic.
There are no defaults and no global clock/randomness. `Step` calls `Now` exactly once and
`NewLeaseID` exactly once before claim; both outputs must be nonzero/valid. `Jitter` is invoked
outside repository calls with exactly `PollInterval` after successful apply and exactly
`RetryInterval` for a retry; it must return `0 <= value < input`, otherwise the step fails
`invalid`. `NextAttemptAt = now + interval + jitter` is computed with checked arithmetic.

Before copying or allocating candidate IDs, the service caps folders at
`MaxEngineFolderItems`, an incremental page at combined
`len(CreatedOrUpdated)+len(DestroyedIDs) <= MaxEnginePageItems`, and an enumeration page at
`len(Messages) <= MaxEnginePageItems`. The accepted 04F classifier remains the authority for
the exact page union, nonnil slices, account/cursor/checkpoint, returned-count, aggregate-byte,
projection and all-and-only reservation rules.

Volatile event candidates are allocated outside every repository call after the bounded engine
page returns: sort created/updated engine message IDs bytewise and call `NewEventID` first for
received, then changed; then sort destroyed IDs bytewise and call it once for changed. A live
candidate has both IDs and `Destroyed=false`; a destroyed candidate has zero received ID, one
changed ID and `Destroyed=true`. Every generated ID is nonzero and unique within the step.
Candidates are not reservations and are never persisted unless `PrepareMailboxSyncPage`
selects them. Prepare must return the exact 04F input with all-and-only required reservations;
04F independently rejects a missing, extra, committed-new or unused reservation. This preserves
the accepted cutoff rule: first-seen inbound/unknown strictly after `EventCutoff` gets one
received ID/event/marker; before/equal gets zero reservation/no event/nil marker; confirmed
draft/submission provenance never gets received.

The prepare request uses `OccurredAt=now.UTC()` from the one Step clock sample and
`CorrelationID=leaseID.String()` in canonical lowercase UUID form. Prepare obtains the mailbox
and organization IDs from the fenced account, sets `ActorID=nil`, and must return those exact
OccurredAt/CorrelationID values. The service rejects unless the prepared input's account and
`EventCutoff` exactly equal the claim result and every event-context value matches this rule.

Every slice/byte field crossing engine, repository, policy, or result boundaries is deep-copied.
The service never mutates dependency-owned values. No cursor/provider/secret/raw message content
is persisted except through the already frozen 04F projection intents, and none is logged.

## One-step state machine

`ClaimMailboxSyncStep` is one short DB transaction/lock boundary. It atomically proves the
account is active and due, its immutable nonzero event cutoff is committed, no live mutation
lease exists, and the supplied sync lease can be acquired/renewed. An expired mutation lease is
reclaimed only by incrementing `MutationEpoch`. The returned fence carries the exact account,
lease token/version, sync epoch, mutation epoch, phase and committed position. A nonclaim returns
`Claimed=false`, a safe `NextAttemptAt`, no engine account/cutoff/fence material, and produces
`MailboxSyncStepIdle` with zero engine calls.

For a claimed step the service performs `ListFolders` exactly once and, only if it succeeds,
exactly one page call: `Changes` for `Incremental` or `CatchUp`, or `EnumerateProjection` for
`Enumeration`. It never performs both page calls and never loops pages in one step.

Every Changes request and every initial or resumed EnumerateProjection request sets exactly
`Page: EnginePageRequest{Limit: config.PageSize}`. No phase substitutes a constant, accepted
maximum, engine-returned bound or repository value for the validated config value.

Phase shapes are exact:

- `Incremental`: nonempty `Cursor`; empty base/checkpoint. Call `Changes(FromCursor=Cursor)`.
- initial `Enumeration`: cursor/base/checkpoint all empty. The first page captures one nonempty
  `BaseCursor`.
- resumed `Enumeration`: cursor empty; base and checkpoint both nonempty and echoed exactly.
- `CatchUp`: nonempty cursor, empty base/checkpoint. It begins at the completed enumeration
  page's `BaseCursor` and calls `Changes`.

`PrepareMailboxSyncPage` is a second short DB transaction. Under the exact fence it resolves
Hub message/thread/label IDs, previous state, owner provenance and committed event reservation
state, and returns one deep-copied `MailboxSyncPolicyInput`. The service then calls
`ClassifyMailboxSyncPage` outside any repository call. `ApplyMailboxSyncPage` is a third short
transaction that atomically rechecks the complete fence and applies every 04F projection/event/
count command plus the position transition. A guard mismatch applies nothing and returns
`Applied=false`; the service returns `MailboxSyncStepLeaseLost`. `Replayed=true` is permitted
only for the same previously committed full fence+advance and must not duplicate projection,
count or outbox effects.

Prepare, Apply and Resolve each read persisted `sync_lease_until` and compare it inside the same
transaction against the database's authoritative stable transaction clock (`CURRENT_TIMESTAMP`
or dialect-equivalent). Caller `request.Now` is never expiry authority, and matching epochs
without takeover are insufficient. Every mutation additionally requires
`sync_lease_until > authoritative_tx_time`. Equal/earlier returns `Prepared=false` with zero
Input, `Applied=false`, or `Resolved=false`, respectively; each surfaces
`MailboxSyncStepLeaseLost` with zero effects even when no takeover occurred. The service never
reads a DB clock and no DB-clock query occurs outside a repository transaction.

Change advancement is literal and avoids manufacturing provider cursors: require
`Result.Advance.Change.FromCursor == Fence.Cursor`; when `NextCursor` is nonempty it becomes the
next committed cursor, and on a terminal page (`NextCursor==""`) the committed cursor remains
the returned `FromCursor`. Incremental remains incremental. Catch-up remains catch-up while a
nonempty next cursor exists; its terminal page atomically finalizes the current sync epoch and
transitions to incremental at that returned `FromCursor`. This exact from/next rule matches the
accepted terminal change-page contract and prevents an empty stored cursor.

Enumeration advancement requires the exact input base/checkpoint. A noncomplete page persists
the returned base and nonempty next checkpoint. A complete page requires empty next checkpoint,
persists its base as the catch-up cursor and transitions to `CatchUp`. Enumeration never clears
previous good rows; epoch tagging and catch-up terminal finalization determine stale projection
cleanup in the future 04H adapter.

## Failure, fencing and no-I/O rules

`ResolveMailboxSyncStep` is the fourth and only failure transaction. `InvalidCursor` from a
change call and `RestartRequired` from enumeration recheck the full fence, increment
`SyncEpoch`, set degraded/rebuild-enumeration with empty base/checkpoint, preserve the last good
projection, cutoff and freshness, and return `RebuildScheduled`; enumeration begins on the next
`Step`, never as a second page call in the failing step. `Transient` retains the committed
position and schedules bounded retry. `Rejected` records only the typed internal reason and does
not manufacture success. A resolve fence mismatch changes nothing and returns lease lost.

Context cancellation stops immediately; if claim already committed, the lease expires naturally.
A crash before apply leaves no projection/cursor/event effects. A crash or ambiguous repository
return after apply is recovered from the next claim: the committed fence/advance is replay-safe,
and no late owner can apply after lease, lease epoch, sync epoch or mutation epoch changes. No
engine call, object/storage I/O, callback (`Now`, ID generators, jitter), classifier call, or log
occurs inside a repository method/DB transaction/lock. Repository values contain no callback or
engine handle, so a future adapter cannot hide such I/O behind the port.

Error mapping is closed. Engine `ErrEngineInvalidCursor` maps only to
`MailboxSyncFailureInvalidCursor`; `ErrEngineRestartRequired` only to
`MailboxSyncFailureRestartRequired`; `ErrEngineUnavailable` and `ErrEngineIndeterminate` map to
`Transient`. Engine/policy payload-too-large resolves `Rejected` then returns the service
too-large sentinel. Engine rejected-policy/unauthorized-account and 04F invalid/duplicate/
account-mismatch resolve `Rejected` then return service invalid. Unknown/raw engine or repository
errors return only service unavailable; an ambiguous Apply error is never resolved over because
the commit may have succeeded. Context cancellation returns only the context sentinel. If
Resolve itself errors, only service unavailable is returned. `Replayed` implies `Applied`; false
Prepared, false Applied and false Resolved are lease loss, never dependency failure.

## Exact authored-first static oracle

The new test file must use deterministic in-memory engine/repository spies only; no DB, broker,
network, filesystem, storage, Ent, goroutine worker, or clock. It must enumerate at least:

1. ABI reflection for every exported field, enum, sentinel, method and exact3 path/order.
2. Constructor nil/zero/overflow/page-size failures and content-free `Error`/`Unwrap`/`errors.Is` behavior.
3. not-due/disabled/lease-lost/mutation-lease-live claim paths: zero engine calls and zero apply.
4. expired mutation lease reclaim with epoch bump; late old lease/epoch/mutation owner applies
   zero commands.
5. successful incremental: one ListFolders, one Changes with exact configured Page Limit, no
   Enumerate; exact prepare, one 04F call result, one atomic apply, poll scheduling and deep-copy
   isolation.
6. terminal incremental with empty next retains exact from cursor; nonterminal advances to exact
   next cursor. Neither stores empty or synthesizes a cursor.
7. initial/resumed enumeration echo with exact configured Page Limit, checkpoint persistence,
   terminal transition to catch-up at base, then a later Step does one Changes from base with the
   same exact configured limit; no same-step second page call.
8. catch-up nonterminal next advancement and terminal epoch finalization/transition to
   incremental at from; prior good rows are not cleared during rebuild.
9. invalid change cursor and enumeration restart schedule next-step rebuild with sync-epoch bump;
   transient retains position; resolve guard loss mutates nothing.
10. deterministic candidate order, nonzero/duplicate generator rejection, all-and-only prepare
    reservations, accepted 04F before/equal/after cutoff and confirmed-owner provenance oracles.
11. cap-before-copy/allocation for folders/change/enumeration, aggregate/policy errors, nil/empty
    page collections, account mismatch and hostile deep-copy mutation.
12. crash before apply, same-fence apply replay, ambiguous apply, cancellation, lease expiry and
    idempotent retry with no duplicate outbox/count/projection effects. A transaction-clock spy
    makes an engine call cross persisted lease expiry without takeover: subsequent Prepare,
    direct Apply and direct Resolve each compare DB transaction time, return false and surface
    LeaseLost with zero effects; equality expires and request.Now never rescues the lease.
13. hostile raw provider/repository errors never appear in returned error text, results, logs or
    persisted failure; callbacks/classifier/engine/storage are asserted absent inside each
    repository transaction marker.

All test/build/gofmt/lint/validator/Docker/runtime evidence remains `NOT_RUN`; launcher remains
`NOT_READY`. These are future static oracles, not current PASS evidence.

## Candidate self-review

- Critical: 0
- Important: 0
- Minor: 0

This self-review does not replace any required independent precommit or immutable exact-SHA
review and is not source-authority acceptance.
