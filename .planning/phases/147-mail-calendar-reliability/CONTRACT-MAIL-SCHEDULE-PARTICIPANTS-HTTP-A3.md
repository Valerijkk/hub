# Mail scheduling participants HTTP A3 - frozen atomic contract

**Authority:** derives only from
[`CONTRACT-MAIL-SCHEDULING-BRIDGE.md`](CONTRACT-MAIL-SCHEDULING-BRIDGE.md)
and the accepted A1 domain API in
[`CONTRACT-MAIL-SCHEDULE-PARTICIPANTS-A1.md`](CONTRACT-MAIL-SCHEDULE-PARTICIPANTS-A1.md).

**Atom:** A3 (`backend/internal/mail/mailbox_schedule_participants_handler.go` only)

**Status:** `FROZEN_FOR_ATOMIC_IMPLEMENTATION`

**Provenance:** `dispatcher-enforced` - `gpt-5.6-terra` / `high`; actual runtime telemetry is not exposed.

**Evidence:** `SOURCE_CONTRACT_AUTHORED / STATIC_NOT_RUN / RUNTIME_NOT_RUN`

## 1. Outcome and exact boundary

A3 defines one isolated authenticated Mail HTTP adapter for the already closed
A1 service API:

`POST /api/mail/messages/{messageId}/scheduling-participants`

The handler accepts only a canonical non-nil message UUID in the path and one
closed JSON object containing selected address strings. It obtains the actor
only from the established authenticated Mail context, calls A1 exactly once,
and publishes only the ordered canonical email plus nullable internal user ID
result already authorized and validated by A1/A2.

A3 owns no persistence, source-selection, user lookup, owner logic, provider
or engine read, Calendar call, availability check, meeting/event creation,
cross-domain orchestration, correlation, worker, NATS subject, migration,
OpenAPI, TypeScript client, UI, production composition, or test.

A3 is deliberately source-only and unbound after implementation. It MUST NOT
be registered in production in this atom. Public OAS D1, handwritten Mail
client C1a, and production construction/registration B2 remain separately
frozen atoms with their own exact ownership and review. This split must not be
collapsed by editing those files opportunistically.

## 2. Sole owned production file

The future implementation may add exactly:

`backend/internal/mail/mailbox_schedule_participants_handler.go`

No existing Mail handler is edited. In particular A3 MUST NOT add this route
to `mailbox_message_handler.go`, widen `mailboxMessageReader`, reuse the
message-detail service, or expose A1 through a broad aggregate handler. The
new handler depends only on `MailboxScheduleParticipantsServiceAPI` as its
runtime application dependency.

The file may use standard-library packages required for bounded strict JSON,
Fiber, `uuid`, and existing same-package authenticated Mail helpers. It must
not import Ent, Calendar, an engine/provider package, server composition, or
any unrelated domain.

## 3. Handler and route shape

The new file declares the isolated shape:

```go
const mailboxScheduleParticipantsHTTPJSONBodyMaxBytes = 512 * 1024

type MailboxScheduleParticipantsHandler struct {
    service MailboxScheduleParticipantsServiceAPI
}

func NewMailboxScheduleParticipantsHandler(
    service MailboxScheduleParticipantsServiceAPI,
) (*MailboxScheduleParticipantsHandler, error)

func (handler *MailboxScheduleParticipantsHandler) RegisterRoutes(
    api fiber.Router,
    authMiddleware fiber.Handler,
) error
```

The constructor rejects nil and typed-nil service values with the closed A1
`internal` error. `RegisterRoutes` rejects a nil/invalid handler, service,
router, or auth middleware without registering anything. It creates the
existing authenticated Mail group and adds exactly:

```go
routes := api.Group("/mail", authMiddleware)
routes.Post(
    "/messages/:messageId/scheduling-participants",
    handler.resolve,
)
```

There is no unauthenticated alias, alternate method, trailing route variant,
Calendar route, or handler-local auth bypass. Existing middleware remains the
authentication authority. The operation then obtains the actor only through
`mailboxActorID(ctx)`; it never accepts an actor, owner, account,
organization, user ID, or identity from the path, query, headers, or body.

A request with any query-string bytes is `400 invalid mail scheduling
request`. A3 accepts no query parameter, including an empty named parameter.

`messageId` is parsed as a UUID and accepted only when all are true:

- parsing succeeds;
- the UUID is not `uuid.Nil`; and
- the original path bytes equal `messageID.String()` exactly.

Thus upper-case, braced, URN, compact, non-hyphenated, nil, malformed, or
otherwise non-canonical UUID representations are all the same generic 400.
No service call occurs for them.

## 4. Closed request DTO and strict JSON

The only request transport type is unexported and has exactly one JSON key:

```go
type mailboxScheduleParticipantsHTTPRequest struct {
    Addresses *[]string `json:"addresses"`
}
```

A pointer is required so an absent or explicit `null` `addresses` value is
rejected before A1. The handler requires exactly one `Content-Type` field
value through the existing Mail helper, then parses only that value:

```go
value, one := mailboxSubmissionHTTPExactlyOneHeaderValue(
    ctx,
    fiber.HeaderContentType,
)
if !one {
    return mailboxScheduleParticipantsHTTPInvalidRequest()
}
mediaType, _, err := mime.ParseMediaType(
    strings.TrimSpace(string(value)),
)
if err != nil || !strings.EqualFold(mediaType, fiber.MIMEApplicationJSON) {
    return mailboxScheduleParticipantsHTTPInvalidRequest()
}
```

Parameters such as `charset=utf-8` remain allowed by `mime.ParseMediaType`.
A missing, duplicate, malformed, or different media type is the same generic
400 before the service call; A3 does not create a distinct 415 surface. It
must not use `ctx.Get` or any first/last-value behavior that could silently
accept one value from a duplicated `Content-Type` field.

The strict decoder must enforce all of the following before calling A1:

1. Raw body length is 1..524288 bytes. Oversize is the same generic 400; it
   must not be read into another unbounded buffer and does not create 413.
2. Raw bytes are valid UTF-8.
3. Outer JSON whitespace consists only of the RFC JSON whitespace bytes
   `SP` (`0x20`), `HTAB` (`0x09`), `LF` (`0x0A`), and `CR` (`0x0D`). A local
   bounded trim/view may remove only those four byte values before the object
   precheck. It must not use `bytes.TrimSpace`, Unicode whitespace trimming,
   or otherwise accept NBSP or any non-JSON whitespace byte around the value.
4. After that JSON-only outer-whitespace handling, the first token is an
   object, never an array, scalar, or `null`.
5. The sole allowed top-level key is exactly `addresses`; unknown keys and a
   duplicate `addresses` key are rejected.
6. The value decodes as a non-null JSON array of JSON strings. A missing,
   null, scalar, object, nested array, or null member is rejected.
7. `json.Decoder.DisallowUnknownFields()` is enabled for the DTO decode.
8. Exactly one JSON value is present. A second value, trailing token, or any
   non-whitespace bytes after the object are rejected; only `io.EOF` after
   the first complete object is accepted. Any surrounding whitespace admitted
   by this check is still restricted to the four JSON whitespace bytes above.

Duplicate-key detection must be local to this file and bounded by the same
raw-body cap; A3 must not import or widen Calendar's decoder. The decoder does
not retain raw bytes or caller-owned mutable state after the request.

A3 itself does not lowercase, trim, parse display names, IDNA-convert, sort,
deduplicate, or otherwise rewrite an address. It passes a fresh slice copy to
A1 as received. A1 remains the sole application authority for the required
1..200 cardinality, exact-byte uniqueness, `NormalizeAddress` equality, and
320-byte address ceiling. Every A1 validation failure maps to the same generic
400; the handler exposes no address index, value, field, or reason.

The only application call is:

```go
result, err := handler.service.ResolveMailboxScheduleParticipants(
    ctx.Context(),
    actorID,
    MailboxScheduleParticipantsInput{
        MessageID: messageID,
        Addresses: copiedAddresses,
    },
)
```

It executes at most once and only after auth, query, UUID, content-type, body,
object/key, and presence validation have succeeded.

## 5. Exact success representation

Domain structs are not serialized directly. A3 declares only these closed
transport response types:

```go
type mailboxScheduleParticipantsHTTPParticipant struct {
    Email  string     `json:"email"`
    UserID *uuid.UUID `json:"user_id"`
}

type mailboxScheduleParticipantsHTTPResponse struct {
    Participants []mailboxScheduleParticipantsHTTPParticipant `json:"participants"`
}
```

A `200 application/json` response therefore has exactly:

```json
{
  "participants": [
    {"email": "anna@example.com", "user_id": "8cf3a321-0000-4000-8000-000000000001"},
    {"email": "outside@example.net", "user_id": null}
  ]
}
```

`participants` is always present and contains exactly 1..200 entries. Every
entry always contains both `email` and `user_id`; `user_id` is JSON `null`
when the A1 pointer is nil and is never omitted. No `Address` Go field name is
allowed to leak onto the wire.

Before publishing, A3 defensively requires the result length to equal the
submitted address length, every result address to equal the request address at
the same index byte-for-byte, and every non-nil user ID to be non-zero. A
violation is generic 500 with no partial response. It constructs a fresh
response slice and fresh UUID value pointers, preserving order. It adds no
display name, organization, role, source/header, self flag, status, reason,
count, membership, availability, timing, message metadata, or Calendar field.

## 6. Closed HTTP error mapping

Authentication failures from the established middleware/`mailboxActorID`
retain the existing generic `401 unauthorized` handling. After authentication,
the A3 mapping is exactly:

| Condition | Status, public message, and headers |
|---|---|
| Any query string; invalid/non-canonical/nil UUID; missing/duplicate/malformed/different content type; empty/oversize/invalid UTF-8/non-object/malformed/multiple/trailing/duplicate-key JSON; non-JSON outer whitespace including NBSP; unknown key; absent/null/invalid `addresses`; A1 `invalid_request` | `400`, `invalid mail scheduling request`, no `X-Hub-Error-Code` |
| A1 `not_found` | `404`, `mail message not found`, no `X-Hub-Error-Code` |
| A1 `deleting` | `409`, `mailbox account deleting`, `X-Hub-Error-Code: MAIL_ACCOUNT_DELETING` |
| A1 `unavailable` | `503`, `mail message unavailable`, no `X-Hub-Error-Code` |
| A1 `internal`; unknown, typed-nil, or unrecognized error; malformed success result; encoding failure | `500`, `mail message unavailable`, no `X-Hub-Error-Code` |

All errors are returned as `fiber.Error` for the existing global error handler,
which remains responsible for the canonical `{ "error", "request_id" }`
envelope. A3 must not JSON-encode a second error envelope, log selected
addresses, return the A1 error text/kind, attach a per-address error, or expose
a wrapped cause.

The 404 is deliberately identical for a missing/foreign message and every
self, Bcc-only, non-header, or otherwise rejected source selection represented
by A1/A2. A3 does not reinterpret it and returns no secondary code. The only
machine header this atom may set is `MAIL_ACCOUNT_DELETING` for the exact 409.

## 7. Frozen adjacent atoms

A3 does not grant authority for any adjacent work:

| Later atom | Exact ownership and boundary |
|---|---|
| A2 | `backend/internal/mail/mailbox_ent_schedule_participants.go`; implements the frozen A1 repository port and all owner-scoped source-selection/User lookup. A3 does not depend on its concrete type. |
| D1 | `docs/mail/openapi.yaml`; adds the normative `/mail/messages/{messageId}/scheduling-participants` operation, strict request/response schemas, authz, canonical path UUID, and exact errors. No legacy Swag edit. |
| C1a | `frontend/src/api/mail.ts`; adds only typed request/result DTOs and `resolveMailSchedulingParticipants`, reusing `normalizeMailMessageID` and never rewriting email bytes. |
| B2 | `backend/cmd/server/wiring_calendar_mail.go`; only after A2 and A3 acceptance, constructs `NewMailboxScheduleParticipantsService(mailboxSvc, mailboxRepo)`, constructs the A3 handler, and registers it on `d.api` with `d.requireAuth`. It owns no Calendar availability registration. |
| C2 | Later Mail UI dialog/action/route packet; selects only already visible `From + To + Cc` candidates and orchestrates Mail resolver -> redacted Calendar check -> existing create route. |
| Q1 | Later dedicated Docker-only API/authz packet under `testframework/tests/api/**`. |

A3 may compile in isolation against the accepted A1 API before A2 exists,
because its sole service dependency is the interface. Production composition
B2 is blocked until `*MailboxEntRepository` implements the A1 repository port
through accepted A2. Until D1, C1a, and B2 are separately accepted, the route
must remain unregistered and there is no public/runtime endpoint claim.

## 8. Static acceptance

The future A3 implementation is statically acceptable only when independent
review establishes all of the following from the immutable diff:

- the only production path changed is the one new handler file;
- the handler field is typed only as `MailboxScheduleParticipantsServiceAPI`;
- there is no Ent, Calendar, engine/provider, repository, NATS, worker, or
  server-composition dependency/import;
- the only declared route is the authenticated POST route and no source file
  actually invokes `RegisterRoutes` in this atom;
- actor identity comes only from `mailboxActorID`; all query strings and every
  non-canonical/nil message UUID fail before the service call;
- exactly one `Content-Type` is required through
  `mailboxSubmissionHTTPExactlyOneHeaderValue`; its single value is parsed by
  `mime.ParseMediaType(strings.TrimSpace(string(value)))`, and only a
  case-insensitive `application/json` media type continues;
- JSON is capped at 512 KiB, valid UTF-8, object-only, one-value-only,
  trailing-free, duplicate/unknown-key-free, and requires non-null addresses;
  its outer whitespace accepts only SP/HTAB/LF/CR and never uses
  `bytes.TrimSpace` or admits NBSP/other Unicode whitespace;
- A3 performs no email normalization/rewrite and delegates exact canonical,
  uniqueness, and cardinality enforcement to A1;
- the service executes no more than once and only after complete transport
  validation;
- success has exactly `participants`, then exactly `email` and `user_id` per
  item, with ordered alignment, explicit nullability, and no extra field;
- every error has the exact generic mapping in section 6, with no oracle or
  raw cause, and only deleting sets a machine header; and
- no OAS, frontend, UI, wiring, A1/A2, test, migration, documentation outside
  this contract, or unrelated file is changed in the A3 implementation packet.

Required source evidence is exact changed-path inspection, complete diff
review, and `git diff --check` for the one production file. This contract and
its future implementation packet authorize no test, build, formatter, lint,
generator, Docker, database, network, runtime, merge, push, or deploy claim.
Those evidence classes remain `NOT_RUN`.

## 9. Blockers and stop conditions

A3 implementation is `BLOCKED` rather than broadened when any of these holds:

- the accepted A1 API or error vocabulary is absent or has drifted;
- implementing the route would require reading Bcc/body/raw provider state,
  adding a User/Ent query, or importing Calendar/engine code;
- strict duplicate/unknown/trailing JSON rejection cannot be kept within the
  one-file, 512-KiB-bounded handler;
- production registration, OAS, TypeScript, UI, test, or composition is
  requested inside the A3 packet; or
- required Terra/xhigh implementation provenance or independent review is
  unavailable.

A blocker requires a new root-owned decomposition or frozen amendment. It is
not permission to modify an adjacent atom, weaken strict transport handling,
expose a partial route, silently register it, or substitute runtime claims for
missing source authority.
