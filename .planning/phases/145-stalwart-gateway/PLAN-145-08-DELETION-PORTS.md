# Phase 145-08 — Mail deletion-port implementation plan

> **Current narrow source-authority receipt (2026-08-23; highest precedence):** immutable exact15 `8bc23a47ea9c511fa8c54e9189189f8168a69c41` accepts immutable plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` as `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` for one future atomic no-test exact2 production packet only: `backend/internal/mail/mailbox_deletion.go` and `backend/internal/mail/mailbox_deletion_fake.go`. Plan Contract/high, Security/xhigh, and Lifecycle/high reviews are C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. This status-alignment carrier records that already-granted authority and is not itself an authority receipt. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own exact commit and immutable Contract/Security/Lifecycle triple C/I/M `0/0/0`. Within this exact2 scope only, the prior `PLAN_CANDIDATE`, `SOURCE_BLOCKED`, and no-dispatch clauses below are `CONSUMED/HISTORICAL`; all other blockers and `NOT_RUN` claims remain current. Adapter-capability proof is waived only for these provider-neutral ports and standalone fakes, and remains required before any concrete provider adapter may be designed.

> **For agentic workers:** REQUIRED SUB-SKILL: use `superpowers:subagent-driven-development` or `superpowers:executing-plans` when this plan is explicitly authorized. Track the two tasks below with the checkbox syntax; do not broaden the manifest.

**Goal:** Add standalone, provider-neutral Mail account and object deletion-port ABIs plus deterministic in-memory fakes, without turning deletion intent into a provider, persistence, worker, or HTTP implementation.

**Architecture:** `mailbox_deletion.go` owns the closed public ABI, safe typed error surface, and production-local request/state validation. `mailbox_deletion_fake.go` owns two separately constructed mutex-protected fakes: `FakeEngineAccountDeletionPort` and `FakeMailObjectDeletionPort`. Each fake implements only its matching port and owns separate configuration, seeds, maps, and constructor; neither may be attached to `MailEngine` or `FakeMailEngine`. Each derives every result only from its own seeded maps and the production validators; no provider, Ent, storage, clock, random source, callback, goroutine, or I/O is introduced.

**Tech Stack:** Go, `context`, `errors`, `unicode/utf8`, `sync`, `github.com/google/uuid`; existing `EngineError`, `EngineErrorKind`, and `MaxEngineOpaqueIDBytes` in `backend/internal/mail`.

**Spec:** `docs/mail/ENGINE-CONTRACT-0.1.8.md` (frozen deletion-port ABI through `3900b331`); the applicable deletion adjudication carriers in Phase 145 are design inputs only.

## Global constraints

- Create exactly two production files and modify no existing file:
  - `backend/internal/mail/mailbox_deletion.go`
  - `backend/internal/mail/mailbox_deletion_fake.go`
- Do not modify `MailEngine`, `FakeMailEngine`, `mailbox_engine.go`, `mailbox_fake.go`, an Ent schema, generated Ent source, repository, service, worker, provider adapter, storage adapter, HTTP route, NATS handler, migration, SQL, Docker/Compose asset, test, testframework file, or documentation carrier in the atomic source packet.
- The public ABI is exactly the ABI quoted below. Do not add methods, fields, aliases, result values, error kinds, error details, `Unwrap`, provider identifiers, raw object keys in error text, or a generic deletion interface.
- Production validation is local to `mailbox_deletion.go`. It must not depend on the fake. The fake may use the same package-local production validators; the dependency direction is production ABI → fake, never fake → production ABI.
- Values are validated before a map lookup or mutation. `ProvisioningKey` and `EngineAccountID` are valid only as UTF-8 strings of 1..`MaxEngineOpaqueIDBytes` bytes. `Bucket` is valid only as UTF-8 1..63 bytes. `ObjectKey` is valid only as UTF-8 1..512 bytes. `DeletionID` must be nonzero.
- Invalid account requests return `&EngineError{Kind: EngineErrorRejectedPolicy}`. Invalid object requests return a `MailObjectDeletionError` whose validated kind is `MailObjectDeletionErrorRejectedPolicy`; they must not return a generic error or leak input.
- A zero or unknown result/observation state is never emitted. It is invalid at the boundary and must preserve any future durable deletion intent as indeterminate; this packet has no durable intent and therefore has no finalization action.
- Tests are intentionally out of scope by founder direction. Do not create, edit, move, run, plan a test command, or claim test coverage. Build, lint, gofmt, parser, validator, Docker, network, runtime, and provider verification are also `NOT_RUN` for this source-only packet.
- Use an exact-path stage and one commit with no AI attribution. After the commit, perform immutable Contract/high, Security/xhigh, and Lifecycle/high reviews of the exact source SHA before any docs receipt is authored.

---

## Frozen file structure

| File | Responsibility | Must not own |
|---|---|---|
| `backend/internal/mail/mailbox_deletion.go` | The two standalone port interfaces; requests, aliases, closed states/results, content-free object error surface, and local request/state validation | Provider calls, storage calls, fake state, Ent, deletion queue/control, transaction, retry, worker, HTTP/NATS, schema, migration, logging payload |
| `backend/internal/mail/mailbox_deletion_fake.go` | Two separately constructed deterministic mutex-protected fakes, each explicitly seeded and implementing only its matching port; exposes only the seed/configuration helpers required to establish its own local state | `MailEngine`/`FakeMailEngine` widening, shared fake state, global fake mutation, time, random values, provider simulation, persistence authority, asynchronous work |

## Frozen ABI and error surface

`mailbox_deletion.go` must declare these public types with these method signatures and exported request fields, byte-for-byte in meaning:

```go
type EngineAccountDeletionPort interface {
	DeleteAccount(context.Context, EngineAccountDeleteRequest) (EngineAccountDeleteResult, error)
	ReconcileAccountAbsence(context.Context, EngineAccountAbsenceRequest) (EngineAccountAbsenceResult, error)
}

type MailObjectDeletionPort interface {
	DeleteObject(context.Context, MailObjectDeleteRequest) (MailObjectDeleteResult, error)
	ReconcileObjectAbsence(context.Context, MailObjectAbsenceRequest) (MailObjectAbsenceResult, error)
}

type EngineAccountDeleteRequest struct {
	ProvisioningKey string
	EngineAccountID string
	DeletionID      uuid.UUID
}

type MailObjectDeleteRequest struct {
	Bucket     string
	ObjectKey  string
	DeletionID uuid.UUID
}

type EngineAccountAbsenceRequest = EngineAccountDeleteRequest
type MailObjectAbsenceRequest = MailObjectDeleteRequest

type EngineAccountDeletionState string

const (
	EngineAccountDeletionStateDeleted               EngineAccountDeletionState = "deleted"
	EngineAccountDeletionStateAuthoritativelyAbsent EngineAccountDeletionState = "authoritatively_absent"
)

type EngineAccountAbsenceState string

const (
	EngineAccountAbsenceStatePresent                EngineAccountAbsenceState = "present"
	EngineAccountAbsenceStateAuthoritativelyAbsent  EngineAccountAbsenceState = "authoritatively_absent"
)

type MailObjectDeletionState string

const (
	MailObjectDeletionStateDeleted                  MailObjectDeletionState = "deleted"
	MailObjectDeletionStateAuthoritativelyAbsent    MailObjectDeletionState = "authoritatively_absent"
)

type MailObjectAbsenceState string

const (
	MailObjectAbsenceStatePresent                   MailObjectAbsenceState = "present"
	MailObjectAbsenceStateAuthoritativelyAbsent     MailObjectAbsenceState = "authoritatively_absent"
)

type EngineAccountDeleteResult struct { State EngineAccountDeletionState }
type EngineAccountAbsenceResult struct { State EngineAccountAbsenceState }
type MailObjectDeleteResult struct { State MailObjectDeletionState }
type MailObjectAbsenceResult struct { State MailObjectAbsenceState }

type MailObjectDeletionErrorKind string

const (
	MailObjectDeletionErrorUnavailable       MailObjectDeletionErrorKind = "unavailable"
	MailObjectDeletionErrorIndeterminate     MailObjectDeletionErrorKind = "indeterminate"
	MailObjectDeletionErrorConflict          MailObjectDeletionErrorKind = "conflict"
	MailObjectDeletionErrorRejectedPolicy    MailObjectDeletionErrorKind = "rejected_policy"
	MailObjectDeletionErrorUnauthorizedScope MailObjectDeletionErrorKind = "unauthorized_scope"
)

type MailObjectDeletionError struct { kind MailObjectDeletionErrorKind }

func NewMailObjectDeletionError(MailObjectDeletionErrorKind) (*MailObjectDeletionError, error)
func (*MailObjectDeletionError) Error() string
func (*MailObjectDeletionError) Kind() MailObjectDeletionErrorKind
func MailObjectDeletionErrorKindOf(error) (MailObjectDeletionErrorKind, bool)
```

The constructor accepts only the five listed object error kinds. Its invalid-kind error is the generic safe error `"invalid mail object deletion error kind"`. `Error()` always returns exactly `"mail object deletion failed"`. `Kind()` is nil-safe and returns `""` for nil, zero, or unknown stored kinds. `MailObjectDeletionErrorKindOf` must use `errors.As`, return `("", false)` for a typed nil target or a zero/unknown kind, and reveal no wrapped cause. `MailObjectDeletionError` declares no `Unwrap` method.

Account methods use the existing `*EngineError` surface only. Their allowed emitted kinds are `EngineErrorNotFound`, `EngineErrorConflict`, `EngineErrorUnavailable`, `EngineErrorIndeterminate`, `EngineErrorRejectedPolicy`, and `EngineErrorUnauthorizedAccount`. This packet emits `EngineErrorNotFound`, `EngineErrorConflict`, and `EngineErrorRejectedPolicy`; the other allowed kinds remain available for a future concrete adapter and are not invented by the fake. Object methods use only `MailObjectDeletionError` values with the five frozen kinds; this packet emits `indeterminate`, `conflict`, and `rejected_policy` only.

## Exact behavioral matrix

| Operation | Valid exact seeded scope | Valid absent exact scope | Valid scope mismatch | Invalid request |
|---|---|---|---|---|
| `DeleteAccount` | remove only the exact pair and return `EngineAccountDeletionStateDeleted` | return existing `*EngineError{Kind: EngineErrorNotFound}`; do not treat it as final absence | return existing `*EngineError{Kind: EngineErrorConflict}`; do not mutate either indexed account entry | return existing `*EngineError{Kind: EngineErrorRejectedPolicy}`; no lookup/mutation |
| `ReconcileAccountAbsence` | return `EngineAccountAbsenceStatePresent`; no mutation | return `EngineAccountAbsenceStateAuthoritativelyAbsent`; the fake's complete local map is the only proof | return existing `*EngineError{Kind: EngineErrorConflict}`; never call it absent | return existing `*EngineError{Kind: EngineErrorRejectedPolicy}`; no lookup |
| `DeleteObject` | remove only the exact `(Bucket, ObjectKey)` and return `MailObjectDeletionStateDeleted` | return a safe object error of kind `indeterminate`; it never returns `authoritatively_absent` | return a safe object error of kind `conflict`; no mutation | return a safe object error of kind `rejected_policy`; no lookup/mutation |
| `ReconcileObjectAbsence` | return `MailObjectAbsenceStatePresent`; no mutation | return `MailObjectAbsenceStateAuthoritativelyAbsent`; the fake's complete local map is the only proof | return a safe object error of kind `conflict`; never call it absent | return a safe object error of kind `rejected_policy`; no lookup |

For account scope, an exact pair is present only when both lookup indexes resolve to that same pair. A request is `authoritatively_absent` only when neither the provisioning-key index nor the engine-account-ID index has a record. Any one or both nonmatching index hits, including a cross-pair such as requested `P1/E2`, return `EngineErrorConflict` with no mutation. For object scope, the fake must retain an `ObjectKey` secondary index in addition to the exact `(Bucket, ObjectKey)` map. A requested object key that exists only under another bucket is a mismatch; an object key absent from the secondary index is absent. No result permits a caller to infer a provider-side deletion or physical object state.

## Atomic execution order

### Task 1: Create the production-only deletion ABI and validation

**Files:**

- Create: `backend/internal/mail/mailbox_deletion.go`
- Do not modify any other path.

**Ownership:** One Terra/xhigh atomic implementer owns only this file. The fake task cannot start until this task's contract is frozen and the source file is present in the same atomic packet.

**Implementation steps:**

- [ ] Declare the exact public interfaces, request structs, aliases, named state types/constants, and four distinct result structs from the frozen ABI. Do not add these methods to `MailEngine` or a type assertion to it.
- [ ] Implement `MailObjectDeletionError`, its constructor, `Error`, `Kind`, and `MailObjectDeletionErrorKindOf` exactly as frozen. Import `errors` only for `errors.As`; do not add a cause field or `Unwrap`.
- [ ] Add package-local pure helpers that validate account requests, object requests, the five object error kinds, and every named result/observation state. Validation must use UTF-8 byte length, reject `uuid.Nil`, have no I/O, and return/construct only the safe errors stated above.
- [ ] Keep validation entirely local to this production file. It must not import, instantiate, or refer to either fake type, test types, a provider, Ent, storage, or worker state.

**Static acceptance:** Inspect the staged diff for the one intended file and confirm all exported types match the frozen ABI. Do not run any executable command. Commit only after Task 2 is complete, because the approved atomic manifest is exactly two source files.

### Task 2: Create the standalone deterministic deletion-port fake

**Files:**

- Create: `backend/internal/mail/mailbox_deletion_fake.go`
- Depends on: Task 1 in the same atomic commit.
- Do not modify any other path.

**Ownership:** The same Terra/xhigh atomic implementer owns only this file and must accommodate concurrent unrelated changes without reverting them.

**Frozen fake shape:**

```go
type FakeEngineAccountDeletionSeed struct {
	ProvisioningKey string
	EngineAccountID string
}

type FakeEngineAccountDeletionPortConfig struct {
	Accounts []FakeEngineAccountDeletionSeed
}

type FakeMailObjectDeletionSeed struct {
	Bucket    string
	ObjectKey string
}

type FakeMailObjectDeletionPortConfig struct {
	Objects []FakeMailObjectDeletionSeed
}

type FakeEngineAccountDeletionPort struct {
	// private sync.Mutex and private account indexes only
}

type FakeMailObjectDeletionPort struct {
	// private sync.Mutex and private object indexes only
}

func NewFakeEngineAccountDeletionPort(FakeEngineAccountDeletionPortConfig) (*FakeEngineAccountDeletionPort, error)
func NewFakeMailObjectDeletionPort(FakeMailObjectDeletionPortConfig) (*FakeMailObjectDeletionPort, error)
```

Each constructor is deterministic and creates no global state. `NewFakeEngineAccountDeletionPort` returns `EngineErrorRejectedPolicy` for an invalid or duplicate account seed; `NewFakeMailObjectDeletionPort` returns `MailObjectDeletionErrorRejectedPolicy` for an invalid or duplicate object seed. The fakes have no exported mutation, fault scripting, time, random, callback, provider, or goroutine API. Each constructor config is that fake's sole seeding mechanism; the account fake's only operations are its two account-port methods, and the object fake's only operations are its two object-port methods.

**Implementation steps:**

- [ ] Give each fake its own private `sync.Mutex` guarding only that fake's maps. The account fake retains two account indexes — provisioning key → engine account ID and engine account ID → provisioning key. The object fake retains an exact object map and an object-key secondary index. All reads that determine a result and all deletes occur while the owning fake's mutex is held; no state, map, constructor, or error path is shared between fakes.
- [ ] Validate each constructor's seeds with the production-local validators. The account constructor rejects duplicate exact account pairs, duplicate provisioning keys, and duplicate engine account IDs. The object constructor rejects duplicate exact object tuples and duplicate object keys. This makes mismatch versus absence deterministic rather than dependent on map iteration.
- [ ] Implement the account delete/reconcile rows of the behavioral matrix exactly. Deleting a present account removes both account indexes atomically. A missing delete returns `EngineErrorNotFound`; reconcile returns `authoritatively_absent` only when both indexes miss. Any one or both nonmatching index hits, including cross-pair `P1/E2`, returns `EngineErrorConflict` with no mutation.
- [ ] Implement the object delete/reconcile rows of the behavioral matrix exactly. Deleting a present object removes the tuple and secondary index atomically. A missing delete returns the safe `indeterminate` error; only reconcile may emit `authoritatively_absent`.
- [ ] Add compile-time local assertions in the fake file only: `var _ EngineAccountDeletionPort = (*FakeEngineAccountDeletionPort)(nil)` and `var _ MailObjectDeletionPort = (*FakeMailObjectDeletionPort)(nil)`. Do not add an assertion involving `MailEngine` or `FakeMailEngine`.

**Static acceptance and commit:** The highest-precedence receipt for immutable plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` authorizes an implementer to inspect the textual staged diff, stage exactly the two named new files by path, and create one atomic commit without AI attribution. This status-alignment carrier does not self-authorize source. Immediately after the source commit, capture its SHA, parent SHA, tree SHA, both blobs, exact two-path manifest, dispatcher-enforced Terra/xhigh provenance, and `NOT_RUN` executable status as **review-dispatch metadata**, explicitly not as a docs receipt. No source authority extends beyond this exact2 packet.

## Mandatory review and receipt sequence

1. Do not author a docs receipt before the source commit exists and its exact manifest/blob IDs are captured.
2. Dispatch three independent immutable reviews of that exact source SHA using the review-dispatch metadata:
   - Contract / ABI review at Terra/high: exported shape, validation, aliases, closed states, error surface, no `MailEngine` widening.
   - Security review at Terra/xhigh: no raw identifiers/cause leakage, no missing-delete finalization, no unsafe fake state, no provider/storage creep.
   - Lifecycle review at Terra/high: deterministic locking, exact deletion/reconcile semantics, mismatch handling, scope boundaries, and honest gate status.
3. Any C/I/M finding blocks docs receipt and all downstream work until a new exact source commit is produced and re-reviewed.
4. Only C/I/M `0/0/0` on the final immutable source SHA permits a separate docs receipt. That receipt repeats the review-dispatch metadata and records all three review outcomes; it must state `STATIC_SOURCE_ONLY`, not runtime, Docker, release, or product readiness.

## Explicitly deferred and blocked work

- No selected provider or Stalwart/JMAP route is proven for account deletion; adapter capability evidence is still required before a concrete provider adapter can be designed.
- No durable deletion-intent/control schema, Ent generation, migration, SQL, transaction/CAS implementation, worker, retry, lease, upload-drain proof, attachment/reference drain, hard-delete, or account child/receipt/control/activity manifest is authorized by this plan.
- No object-storage adapter, bucket policy, deletion marker, user-visible attachment behavior, repository/service composition, HTTP, NATS, background worker, or production composition is authorized.
- Physical User, GDPR, Organization, mailbox ownership, retention, legal hold, and audit deletion flows remain separately blocked. These ports do not accept user IDs, organization IDs, mail-account IDs, or policy authority.
- Tests, testframework work, Docker, build, lint, gofmt, parser/validator, runtime, and external verification are intentionally `NOT_RUN`. This plan adds no claim that Phase 145, Mail, Calendar synchronization, Docker launch, or merge readiness has been achieved.

## Plan self-review

- **Spec coverage:** the two-port split, exact ABI, safe object errors, nonzero/byte-bounded request identity, no `MailEngine` widening, missing/mismatch/reconcile semantics, deterministic fake, static review gates, and deferred lifecycle dependencies each map to an explicit section above.
- **Manifest check:** exactly two future source files are named; no current carrier/document is modified by this plan.
- **No-test check:** no test path, test command, or coverage claim is introduced; the founder prohibition is explicit.
- **Authority check:** the highest-precedence `8bc23a47ea9c511fa8c54e9189189f8168a69c41` receipt makes immutable plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` only for the exact2 production manifest above; this status-alignment carrier merely records that receipt. `PLAN_CANDIDATE`, `SOURCE_BLOCKED`, and no-dispatch wording retained elsewhere in this file is `CONSUMED/HISTORICAL` only within that exact2 scope. Adapter-capability evidence remains mandatory for every concrete provider adapter. Source stays `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own immutable Contract/Security/Lifecycle C/I/M `0/0/0` reviews; any finding ends this narrow authority.
> **Current deletion-port source receipt (2026-08-23; highest precedence):** immutable source `e4f81bf3d8b932f5ff2b9a8069d41bfa1d1a0e26` (parent `bd76d8469782e9bb7926b440b6b45495bc865e84`, tree `93bac3c6a5798cb75b0a6647fa938416c5e82ede`) is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/internal/mail/mailbox_deletion.go` blob `001b03c879aab36031feac07f0d3894f76e23abd` and `backend/internal/mail/mailbox_deletion_fake.go` blob `7d43b4dee20ad762d251e3c14e428f2f0ccda497`. `provenance_source: dispatcher-enforced`; Terra/xhigh implementation; Contract/high, Security/xhigh, and Lifecycle/high immutable reviews are C/I/M `0/0/0`. This accepts only the provider-neutral deletion ABI and two standalone deterministic fakes. Tests, build, gofmt, lint, parser, validator, generation, migration, SQL, Docker, and runtime are `NOT_RUN`. Within this exact2 packet, former `UNIMPLEMENTED`/`UNCOMMITTED`/`UNACCEPTED`/exact2-next/source-review-pending clauses are `CONSUMED/HISTORICAL`. Concrete provider adapter capability, persistence/Ent/schema, migrations, workers/drains/finalization/User-GDPR/HTTP/NATS remain `BLOCKED`; no next source is selected. Testframework merge facts are unchanged.
