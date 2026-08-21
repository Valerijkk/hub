# 144-04I-A - bounded mailbox-sync worker source receipt

> **Status:** `SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED / RUNTIME_NOT_RUN`.
> **Current status:** `STATIC_ACCEPTED / EXACT_SHA_PASS` at source `091f2ef286a71f506e1b048f5386fd37cefa06bc`; exact10 receipt `03674eace728642fc2c9b39d2e8e2c81bf016033` is ancestral to HEAD. The conditional receipt status above is `CONSUMED/HISTORICAL` only for 04I-A. PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback, telemetry `UNKNOWN`; no next source or runtime acceptance follows.
> This dirty exact10 documentation candidate grants no acceptance. It records authored-first source
> commit `091f2ef286a71f506e1b048f5386fd37cefa06bc` (parent
> `6b27643fcde3b702fa13c5fcc9b4bfefcb9fc87a`, tree
> `b01be92acf01033935b85446454eaee16b84e01b`, subject
> `feat(mail): add bounded sync worker`): test blob
> `79c2691971287e3a62c830a5d12c78699414bb86`, then source blob
> `6c20fadf563d4d146b2fc1cbacdddea485266eb7`. Contract, Security and Lifecycle precommit PASS
> C/I/M `0/0/0`; immutable Contract plus Security/Lifecycle exact-SHA PASS C/I/M `0/0/0`.
> Author Terra/xhigh; Contract Terra/high; Security/Lifecycle Terra/xhigh;
> dispatcher-enforced/no fallback; telemetry `UNKNOWN`.
> Only one atomic exact10 receipt-sync commit equal to the manifest below and its immutable receipt
> review may mark source `STATIC_ACCEPTED / EXACT_SHA_PASS`. Then lower 04I-A
> candidate/source-unaccepted/current-next clauses are `CONSUMED/HISTORICAL`; no next source is
> selected. Source/runtime acceptance is not granted by this dirty documentation tree.

## Ordered exact10 source receipt-sync manifest

The atomic receipt-sync commit must change exactly these paths, in this order, with no
substitution or extra path; its diff must equal this manifest exactly:

1. `.planning/STATE.md`
2. `.planning/decisions/specs/MAIL-SPEC.md`
3. `.planning/phases/144-mailbox-domain/PLAN.md`
4. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`
5. `.planning/phases/144-mailbox-domain/DISPATCH-REMAINING.md`
6. `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-04I-A.md`
7. `.planning/phases/144-mailbox-domain/REVIEW-144-04I-A-SYNC-WORKER-READINESS.md`
8. `docs/mail/CONTRACT-CHANGELOG.md`
9. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`
10. `docs/mail/QA-HANDOFF.md`

This receipt authorizes no automatic next source. It recognizes only the immutable exact2 above as
`STATIC_ACCEPTED / EXACT_SHA_PASS` after the atomic exact10 commit and its immutable review; no
other path, composition/wiring change, executable result or runtime acceptance is implied.

## Frozen Go ABI

The declarations, field order, constants and signatures below are exact. `MailboxSyncWorker` is
opaque; it owns no exported fields or methods other than `Run`.

```go
type MailboxSyncWorkerDueAccount struct {
    AccountID     uuid.UUID
    DueAt         time.Time
    UpdatedAt     time.Time
    SyncEpoch     int64
    MutationEpoch int64
    LeaseEpoch    int64
    SyncAttempts  int32
}

type MailboxSyncWorkerScanRequest struct {
    Limit int
}

type MailboxSyncWorkerScanResult struct {
    SelectedAt time.Time
    Accounts   []MailboxSyncWorkerDueAccount
}

type MailboxSyncWorkerFaultKind string

const (
    MailboxSyncWorkerFaultTimeout       MailboxSyncWorkerFaultKind = "timeout"
    MailboxSyncWorkerFaultPanic         MailboxSyncWorkerFaultKind = "panic"
    MailboxSyncWorkerFaultUnavailable   MailboxSyncWorkerFaultKind = "unavailable"
    MailboxSyncWorkerFaultInvalidResult MailboxSyncWorkerFaultKind = "invalid_result"
)

type MailboxSyncWorkerFaultRequest struct {
    Account    MailboxSyncWorkerDueAccount
    SelectedAt time.Time
    Failure    MailboxSyncWorkerFaultKind
    RetryDelay time.Duration
}

type MailboxSyncWorkerFaultResult struct {
    Rescheduled   bool
    RecordedAt    time.Time
    NextAttemptAt time.Time
    SyncAttempts  int32
}

type MailboxSyncWorkerRepository interface {
    ScanDueMailboxSyncAccounts(context.Context, MailboxSyncWorkerScanRequest) (MailboxSyncWorkerScanResult, error)
    RescheduleMailboxSyncWorkerFault(context.Context, MailboxSyncWorkerFaultRequest) (MailboxSyncWorkerFaultResult, error)
}

type MailboxSyncWorkerStepper interface {
    Step(context.Context, uuid.UUID) (MailboxSyncStepResult, error)
}

type MailboxSyncWorkerTimer interface {
    Wait(context.Context, time.Duration) error
    WithTimeout(context.Context, time.Duration) (context.Context, context.CancelFunc)
}

type MailboxSyncWorkerConfig struct {
    Enabled      bool
    ScanInterval time.Duration
    StepTimeout  time.Duration
    RetryMin     time.Duration
    RetryMax     time.Duration
    BatchSize    int
    Concurrency  int
}

type MailboxSyncWorkerErrorKind string

const (
    MailboxSyncWorkerErrorInvalid     MailboxSyncWorkerErrorKind = "invalid"
    MailboxSyncWorkerErrorTooLarge    MailboxSyncWorkerErrorKind = "too_large"
    MailboxSyncWorkerErrorUnavailable MailboxSyncWorkerErrorKind = "unavailable"
)

var (
    ErrMailboxSyncWorkerInvalid     = errors.New("mailbox sync worker invalid")
    ErrMailboxSyncWorkerTooLarge    = errors.New("mailbox sync worker exceeds bounds")
    ErrMailboxSyncWorkerUnavailable = errors.New("mailbox sync worker unavailable")
)

type MailboxSyncWorkerError struct { Kind MailboxSyncWorkerErrorKind }
type MailboxSyncWorker struct { /* unexported */ }

func NewMailboxSyncWorker(MailboxSyncWorkerRepository, MailboxSyncWorkerStepper, MailboxSyncWorkerTimer, MailboxSyncWorkerConfig) (*MailboxSyncWorker, error)
func (*MailboxSyncWorker) Run(context.Context) error
func (*MailboxSyncWorkerError) Error() string
func (*MailboxSyncWorkerError) Unwrap() error
```

Known worker errors expose only the matching sentinel text and unwrap only the matching sentinel.
Nil receiver and unknown kind return the content-free `mailbox sync worker error`, unwrap nil, and
match no sentinel. No raw provider, cursor, account, panic, context or fault-write text is exposed.

## Constructor, scheduling and fault state

- Disabled construction is valid only with nil repository/stepper/timer and an otherwise zero
  config with `Enabled=false`; `Run` on a nonnil context returns nil with zero port calls. Any other
  disabled dependency/configuration shape, any enabled nil or typed-nil dependency, and a nil
  `Run` context are `invalid`. Concurrent `Run` on one worker is `invalid`.
- Enabled bounds are exact: `BatchSize 1..50`, `Concurrency 1..4`, `ScanInterval 1s..5m`,
  `StepTimeout 1s..30s`, and `1s <= RetryMin <= RetryMax <= 5m`. Structural bad duration/ordering
  is `invalid`; a value beyond a frozen cap is `too_large`.
- `Run` is total: nil receiver, nil context and typed-nil context are `invalid` before state. Exactly
  one concurrent `Run` is allowed; reuse after it returns is valid. Its run-guard may protect only
  the running flag, never a port call. It scans immediately, then calls `Timer.Wait` only after the
  entire selected batch has drained.
  It starts exactly a fixed pool of `Concurrency`, uses a bounded FIFO of at most 50 accounts, makes
  one `Stepper.Step` call per selected account, never overlaps an account locally, and never starts
  a new scan while the prior batch is active. The scanner gets no clock: repository `SelectedAt` is
  authoritative DB time and eligible order is `DueAt`, then raw UUID bytes.
- Every `MailboxSyncWorkerScanRequest.Limit` equals `config.BatchSize`. Before any clone or queue
  use, a scan result must have nonzero valid semantic UTC `SelectedAt`, a concrete nonnil account
  slice with at most request `Limit`, strict due/raw-UUID order, no duplicate or zero account ID,
  nonzero valid semantic UTC `DueAt`/`UpdatedAt`, both not later than `SelectedAt`, nonnegative
  sync/mutation epochs, `LeaseEpoch > 0`, and nonnegative `SyncAttempts`. Persisted timestamps may
  be fractional and are preserved byte/instant-exact: worker code never truncates, rounds or
  normalizes them. Malformed output is `invalid`; oversized output is `too_large`; neither reaches
  a step or fault write. A concrete nonnil empty slice waits `ScanInterval`. Scanner raw error or
  recovered panic discards its value; if parent is canceled, `Run` returns nil, otherwise it waits
  `RetryMin` and retries with no account fault write. Malformed/oversize scan exits typed
  invalid/too-large with no wait, step or fault.
- Each account invokes `Timer.WithTimeout` once in a recovery-safe wrapper. A panic creates a safe
  `panic` fault with no `Step`; a nil/typed-nil child context or nil cancellation function creates
  `invalid_result` with no `Step`; a valid cancellation function is called exactly once. After
  `Step` and cancellation, parent cancellation has priority and writes no fault; otherwise any step
  or cancellation panic records `panic`, child deadline records `timeout`, raw step failure records
  `unavailable`, and only then is the 04G result validated. No raw text or panic value is retained.
  A failed or panicking fault write is recovered with its value discarded, contained, and followed by
  the rest of the batch; it never exposes raw data. A `Timer.Wait` panic is likewise recovered and
  discarded: if parent cancellation is observed, `Run` returns nil; otherwise it returns content-free
  `unavailable`, with no account fault write. The same parent-cancel priority prevents a new fault
  attempt where applicable. All workers are joined before `Run` returns. Timer and port cancellation
  responsiveness is a port precondition, not a claim of forced cancellation.
- Valid 04G results are: `Idle`, `Applied`, `RebuildScheduled`, and `RetryScheduled` with a valid
  nonzero `NextAttemptAt`; `LeaseLost` has zero `NextAttemptAt`. Any other shape is
  `invalid_result`. Step success or lease loss writes no worker fault.
- `RetryDelay` is exactly `min(RetryMax, RetryMin * 2^SyncAttempts)`: attempt zero is `RetryMin`.
  It is O(1), computed with checked shift/clamp rather than a loop proportional to attempts; every
  overflow or exponent beyond representability clamps to `RetryMax`. It derives only from the
  captured attempt count; no clock, random jitter, global state or uncontrolled goroutine is allowed.
- `RescheduleMailboxSyncWorkerFault` is the sole worker write. Its repository transaction must use
  DB time and a full captured `(AccountID, DueAt, UpdatedAt, SyncEpoch, MutationEpoch, LeaseEpoch,
  SyncAttempts)` CAS plus active/eligible and no-live-lease guards. `DueAt` and `UpdatedAt` remain
  exact persisted values. `Rescheduled=false` returns a fully zero result and performs no projection,
  cursor, epoch, event, outbox or lease-clear change. `Rescheduled=true` requires valid semantic UTC
  `RecordedAt >= SelectedAt`, checked `NextAttemptAt == RecordedAt + RetryDelay`, and
  `SyncAttempts == captured SyncAttempts + 1`; timestamps preserve raw persisted precision. If
  captured `SyncAttempts == MaxInt32`, a needed fault makes no repository call, other accounts still
  drain and join, and `Run` returns content-free `too_large`. MaxInt32 remains valid scanner input
  so a successful step can recover. A later claim, including an expired/no-live lease that advances
  only `LeaseEpoch`, makes the stale fault CAS false with an exact-zero result. It must never clear
  leases or issue a page command. Timeout/panic/unavailable map only to safe
  `MAIL_SYNC_ENGINE_UNAVAILABLE`; invalid result maps only to `MAIL_SYNC_PAGE_REJECTED`.

No scheduler lock may cover a repository, stepper or timer call. The worker deep-copies scanner
accounts before enqueue/fault use and retains no caller-owned alias. It contains no direct engine,
Ent, SQL, NATS, HTTP, filesystem, log, clock, random or callback dependency.

## Mandatory external-package unit oracles

The authored-first `backend/tests/mail/mailbox_sync_worker_test.go` must be a genuine isolated unit
and reflection-freeze every exported record field/type/order, interface method signature, constructor,
`Run`, constants, sentinels and known/nil/unknown error behavior. It must cover:

1. disabled zero-call behavior; typed-nil dependencies; concurrent run; all exact config boundaries
   and deterministic/saturating backoff including overflow;
2. scan `Limit == BatchSize`, DB-selected due/UUID ordering, concrete-empty wait, raw-error/panic
   retry, fractional semantic-UTC `SelectedAt`/`DueAt`/`UpdatedAt` accepted and preserved exactly,
   and nil/zero/non-UTC/future/duplicate/negative/malformed and oversize results rejected before
   clone/queue, plus 50-account fixed-four FIFO,
   no local overlap and complete drain before wait/rescan;
3. every valid 04G outcome plus malformed result, `WithTimeout` panic/nil child/nil cancel/exactly
   once cancel, timeout, root-cancel priority, deadline, step/cancel panic isolation, unavailable
   step error, failed/panicking fault write, wait error, wait panic and root-cancel priority; other
   accounts continue and no raw error/panic data escapes;
4. exact fault CAS including LeaseEpoch, DB-time/precision/result arithmetic, active/eligible/live-
   lease guards, later-claim expired-lease LeaseEpoch race, false exact-zero result, MaxInt32 no-write
   too-large, two-replica race and protection from ambiguous prior apply; and
5. deep-copy/hostile port alias mutation, no I/O under worker callbacks/locks, no leaked timer or
   goroutine observable under the fake timer, and safe fault-code mapping.

These units do not substitute for 04H's Ent adapter, a real two-replica SQL/DB-time test, launcher
or runtime evidence; all remain later `NOT_RUN` work.

## Dependencies and non-goals

144-04F and 144-04G remain accepted static prerequisites. 144-04H, 144-02E, Ent/concrete adapter,
generation, migration, API/OpenAPI/UI, composition/wiring and runtime remain blocked. This candidate
does not add a schema, event, delivery, direct engine, raw body or release claim. Launcher is
`NOT_READY`; all executable evidence is `NOT_RUN`.
