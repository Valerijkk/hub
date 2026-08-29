# Mail scheduling participants A1 - frozen domain contract

**Authority:** derives only from
[`CONTRACT-MAIL-SCHEDULING-BRIDGE.md`](CONTRACT-MAIL-SCHEDULING-BRIDGE.md)
at `d1c41bc3`.

**Atom:** A1 (`backend/internal/mail/mailbox_schedule_participants.go` only)

**Status:** `FROZEN_FOR_ATOMIC_IMPLEMENTATION`

**Evidence:** `SOURCE_CONTRACT_AUTHORED / RUNTIME_NOT_RUN`

## 1. Scope

A1 defines the closed Mail-domain request/result/error vocabulary, its service,
and the opaque repository port used to resolve selected scheduling-participant
addresses. It does not implement the port, a handler, Ent queries, wiring,
OpenAPI, frontend code, or tests.

The service does not import Calendar and does not read a provider/engine,
message body, MIME, raw snapshot, Bcc, Reply-To, contacts, or headers. In
particular, it must not reuse `MailboxMessageRepository`: that port deliberately
returns a snapshot containing private Bcc and engine state, which this operation
does not need.

## 2. Closed public-domain vocabulary

The one owned file declares exactly the following closed API (unexported helper
names are implementation detail):

```go
const MaxMailboxScheduleParticipantAddresses = 200

type MailboxScheduleParticipantsInput struct {
	MessageID uuid.UUID
	Addresses []string
}

type MailboxScheduleParticipant struct {
	Address string
	UserID  *uuid.UUID
}

type MailboxScheduleParticipantsResult struct {
	Participants []MailboxScheduleParticipant
}

type MailboxScheduleParticipantsRepositoryRequest struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
	MessageID      uuid.UUID
	Addresses      []string
}

type MailboxScheduleParticipantsRepositoryResult struct {
	Participants []MailboxScheduleParticipant
}

type MailboxScheduleParticipantsRepository interface {
	ResolveMailboxScheduleParticipants(
		context.Context,
		MailboxScheduleParticipantsRepositoryRequest,
	) (MailboxScheduleParticipantsRepositoryResult, error)
}

type MailboxScheduleParticipantsServiceAPI interface {
	ResolveMailboxScheduleParticipants(
		context.Context,
		uuid.UUID,
		MailboxScheduleParticipantsInput,
	) (MailboxScheduleParticipantsResult, error)
}
```

`MailboxScheduleParticipant` deliberately carries only the requested canonical
address and its nullable internal user UUID. It has no display name,
organization, role, source/header field, membership, availability, reason,
count, timing, or per-address error/status.

`UserID == nil` has the bridge's required conflated meaning: external, absent,
inactive, cross-organization, ambiguous, or legacy-noncanonical user row. A
non-nil UUID must not be `uuid.Nil`.

## 3. A1 constructor and dependencies

The implementation type is:

```go
type MailboxScheduleParticipantsService struct {
	owners     mailboxOwnerResolver
	repository MailboxScheduleParticipantsRepository
}

func NewMailboxScheduleParticipantsService(
	owners mailboxOwnerResolver,
	repository MailboxScheduleParticipantsRepository,
) (*MailboxScheduleParticipantsService, error)
```

The constructor rejects nil and typed-nil dependencies with the closed
`internal` service error. Before any `ctx.Err()` call, a service call checks
the context with the existing Mail nil-helper semantics: nil or typed-nil
context is `internal`; a non-nil context with `ctx.Err() != nil` is
`unavailable`. Neither context failure is `invalid_request`.

After that context classification, a nil actor/message UUID or invalid address
request is `invalid_request`; invalid service or dependency state is
`internal`. These cases call neither dependency. The service must never retain
or return a caller-owned slice or UUID pointer.

The service resolves the actor's mailbox owner through `owners`, derives only
`ActorID`, `OrganizationID`, and `AccountID`, and submits those values plus the
validated message ID and address copies to the repository. It exposes no owner
record, engine-account ID, mailbox state, identity, or header to its caller.

## 4. Canonical input invariant

`Addresses` is a required ordered set with cardinality **1..200**. For every
input string `value`, A1 calls:

```go
canonical, err := NormalizeAddress(EngineAddress{Email: value})
```

and accepts it only if `err == nil` and `canonical.Email == value`. Thus
normalization is validation, never a rewrite. The address has no display name;
only `EngineAddress.Email` participates. A1 rejects a duplicate by exact
canonical bytes, preserves the initial order, and makes a fresh `strings.Clone`
copy of each accepted address.

An empty slice, more than 200 entries, non-canonical value, malformed/oversize
address, or duplicate is `invalid_request`. A1 must not call `owners` or the
repository for that case. It may not use lowercasing, IDNA-only matching,
display-name parsing, fuzzy matching, `LOWER`, CRM, contacts, or directory
search as a substitute.

## 5. Closed error family and generic semantics

The file defines a private-field typed error, constructor, `Kind()` accessor,
and `MailboxScheduleParticipantsErrorKindOf(error)` recognizer following the
existing content-free Mail error pattern. Its only accepted kinds are:

```go
type MailboxScheduleParticipantsErrorKind string

const (
	MailboxScheduleParticipantsErrorInvalidRequest MailboxScheduleParticipantsErrorKind = "invalid_request"
	MailboxScheduleParticipantsErrorNotFound       MailboxScheduleParticipantsErrorKind = "not_found"
	MailboxScheduleParticipantsErrorDeleting       MailboxScheduleParticipantsErrorKind = "deleting"
	MailboxScheduleParticipantsErrorUnavailable    MailboxScheduleParticipantsErrorKind = "unavailable"
	MailboxScheduleParticipantsErrorInternal       MailboxScheduleParticipantsErrorKind = "internal"
)
```

Its `Error()` text is one fixed content-free string; no kind, selected address,
message ID, owner/account, header, identity, SQL/Ent, or wrapped cause may be
rendered. Unknown/typed-nil errors are not recognized and fail closed as
`internal` at the service boundary.

The later A3 handler maps these **only** as follows:

| A1 kind | HTTP contract |
|---|---|
| `invalid_request` | `400` generic `invalid mail scheduling request` |
| `not_found` | identical `404` body `mail message not found` |
| `deleting` | existing `409` with `X-Hub-Error-Code: MAIL_ACCOUNT_DELETING` |
| `unavailable` | `503` generic `mail message unavailable` |
| `internal` or unknown | `500` generic `mail message unavailable` |

`not_found` intentionally conflates absent/foreign message and every selected
self, Bcc-only, non-header, or otherwise non-candidate address. It carries no
secondary code. `deleting` and `unavailable` are not folded into this 404.

For the A1 repository boundary, A2 returns the same closed A1 error family:
only `not_found`, `deleting`, `unavailable`, or `internal` are valid results.
An A2 `invalid_request`, malformed result, unknown error, or typed-nil error is
an A1 `internal`, never a client validation response. This keeps client input
validation exclusively in A1 and avoids a repository-controlled 400 oracle.

## 6. Required service sequence

The sequence is normative and is the P0 fail-closed boundary:

1. Validate service and dependency state first; nil/typed-nil or otherwise
   invalid service/dependency state is `internal` and calls neither dependency.
   Before `ctx.Err()`, classify context with the existing Mail nil-helper
   semantics: nil/typed-nil context is `internal`; only a non-nil context with
   `ctx.Err() != nil` is `unavailable`. Neither context outcome is
   `invalid_request`.
2. Validate actor ID, message ID, and `1..200` exact canonical ordered-unique
   input. Only these invalid caller request values return `invalid_request`
   without I/O.
3. Call `owners.ResolveOwner` exactly once. Recognize only a non-nil
   `*MailboxServiceError`: `MailboxServiceErrorNotFound` maps to `not_found`,
   `MailboxServiceErrorDeleting` to `deleting`,
   `MailboxServiceErrorUnavailable` to `unavailable`, and
   `MailboxServiceErrorInvalid` to `internal`. Unknown error, typed-nil error,
   or an unrecognized kind is `internal`. This is error mapping only; A1 must
   not claim that `ResolveOwner` emits an owner state as an error.
4. On successful `ResolveOwner`, map the returned owner snapshot separately:
   `not_configured` is `not_found`; `deleting` is `deleting`; an active owner
   with a valid non-nil account/organization snapshot and
   `MailboxProjectionReady` continues. A valid provisioning, suspended, or
   error owner, or any otherwise valid non-ready projection, is `unavailable`.
   A malformed or unknown owner snapshot/state is `internal`. No non-success
   owner result calls A2.
5. Call A2 exactly once with the copied opaque scope and validated addresses.
   A1 makes no User lookup and has no header/identity inputs.
6. Map only the closed A2 failure vocabulary above. Validate A2's success
   result before publishing it: exactly one participant for each requested
   position, identical address bytes at each position, no extra/reordered rows,
   and only nil or non-zero user IDs. Any violation is `internal`.
7. Return a deep copy: a new participant slice, cloned address strings, and a
   fresh UUID pointer for every non-nil `UserID`.

The repository result is a proof of source selection and user resolution, not a
partial-success vector. If even one selected address is not a valid candidate,
A2 returns `not_found` and A1 returns no participants; it must not resolve the
remaining addresses.

## 7. Frozen A1/A2 responsibility split

**Self/header membership belongs to A2, not A1.** A1 deliberately receives no
raw headers, Bcc field, identity rows, or actor-email record; moving membership
into A1 would require widening its port or leaking those data. A2 is the sole
durable adapter allowed to read the minimal owner-scoped projection fields and
identity/user rows necessary to prove membership.

A2 must implement its opaque port in
`backend/internal/mail/mailbox_ent_schedule_participants.go` with this order:

1. Revalidate the non-nil actor/organization/account/message scope and the
   exact canonical, ordered-unique `1..200` address input.
2. In one bounded read transaction/snapshot, establish that the account is
   still owned by the actor in the exact organization and is active/ready;
   map deletion to `deleting`, temporary projection/repository failure to
   `unavailable`, and invalid durable state to `internal`.
3. Select the target persisted message only within that account and only its
   persisted `From`, `To`, and `Cc` address projections. Do not select, decode,
   load, or derive candidates from Bcc, body, raw/provider/engine snapshot,
   Reply-To, contacts, or client-supplied headers.
4. In that same trusted scope, derive the self set from every active mailbox
   identity (primary plus aliases) and the actor's canonical stored `User.email`.
   Remove self addresses from the `From + To + Cc` canonical candidate set.
   If the message is absent/foreign or any requested address is absent after
   this subtraction, return the single `not_found` result **before any User
   lookup**.
5. Only after all requested addresses pass membership, execute one bounded set
   lookup for active users in the exact organization. Emit a non-nil UUID only
   where the requested canonical address exactly equals both
   `NormalizeAddress(EngineAddress{Email: user.Email}).Email` and the stored
   `user.Email` bytes. Every other case is `nil`; do not expose why.
6. Return the complete aligned closed result with copied values. A2 must not
   start engine I/O, Calendar work, or a Mail-to-Calendar import/correlation.

This split proves the bridge's required statement: no User lookup precedes a
source-selection `not_found`, while raw headers and identities never cross the
A2-to-A1 port.

## 8. Copy and validation invariants

- All exported input/result structs are value data; no exported raw header,
  owner, account, organization, identity, engine, or Calendar value is added.
- Input address slices are copied only after exact-canonical validation; output
  participant slices and non-nil UUID pointers are copied on every boundary.
- Neither service nor A2 mutates caller-owned input, retains aliases to it, or
  returns repository-owned slices/pointers.
- A1 must not synthesize, filter, sort, deduplicate, or rewrite an A2 success
  result. It accepts only exact positional echo of its validated request.
- A1 succeeds only if every requested address has a result. It never leaks a
  per-address rejection, match count, source field, self flag, or timing fact.
- A1 has no engine body/calendar import and no transactional side effect.

## 9. A1 static acceptance

The atomic implementation is acceptable only when static review demonstrates:

- the sole production file changed is
  `backend/internal/mail/mailbox_schedule_participants.go`;
- its constructor, service, closed error family, DTOs, and port exactly obey
  sections 2-8;
- input acceptance is exact canonical bytes, order-preserving, unique, and
  bounded `1..200`, without rewrite;
- invalid input causes no owner/repository call;
- nil/typed-nil context is classified as `internal` before any `ctx.Err()`
  evaluation; a non-nil cancelled/deadline context is `unavailable`; neither
  is permitted to become `invalid_request` or HTTP 400;
- `ResolveOwner` errors use the exact recognized `MailboxServiceError` mapping
  and a successful owner snapshot uses the separate state mapping in section 6;
- A1 contains no raw header/Bcc/body/engine/Calendar import or type and does no
  User lookup;
- all external data and results receive the specified validation/deep-copy
  treatment; malformed A2 output fails closed;
- the only cross-atom access is the single opaque A2 port call after owner
  scope validation; and
- no handler, Ent, wiring, OpenAPI, frontend, test, migration, or unrelated
  file is changed in the A1 packet.

No test, build, lint, generation, Docker, database, or runtime claim follows
from this contract. Those evidence classes are **NOT_RUN**.

## 10. Exact next handoff - A2

The next atom is A2 only:

```text
Owner file: backend/internal/mail/mailbox_ent_schedule_participants.go
Implements: MailboxScheduleParticipantsRepository
Consumes:   the frozen A1 request/result/error contract unchanged
May use:    MailboxEntRepository private Ent read machinery and only minimal
            MailAccount, MailMessage From/To/Cc, active MailIdentity, and User
            fields required by Section 7
Must not:   edit A1; add handler/wiring/OAS/frontend/tests; read Bcc/body/raw
            engine data/Reply-To; call Calendar or provider engine; perform a
            User lookup before complete source-membership proof
Acceptance: one bounded owner/org/account-scoped read, exact all-address
            pre-lookup membership gate, exact-byte user-email mapping, aligned
            copied nullable result, and closed error mapping
Evidence:   static only; Docker/runtime remain NOT_RUN
```

A2 may begin only under its own frozen atomic contract, explicit
`gpt-5.6-terra`/`xhigh` implementation dispatch, and independent review. This
A1 contract grants no authority to implement A2 or any later atom.
