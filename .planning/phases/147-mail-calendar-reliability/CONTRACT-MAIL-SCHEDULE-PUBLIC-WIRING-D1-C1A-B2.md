# Mail scheduling participant public wiring - D1, C1a, B2 frozen contract

**Authority:**
[`CONTRACT-MAIL-SCHEDULING-BRIDGE.md`](CONTRACT-MAIL-SCHEDULING-BRIDGE.md),
[`CONTRACT-MAIL-SCHEDULE-PARTICIPANTS-A1.md`](CONTRACT-MAIL-SCHEDULE-PARTICIPANTS-A1.md),
[`CONTRACT-MAIL-SCHEDULE-PARTICIPANTS-HTTP-A3.md`](CONTRACT-MAIL-SCHEDULE-PARTICIPANTS-HTTP-A3.md),
and the accepted A2 Ent-repository source receipt. This document only freezes
three later, separately reviewed atomic implementation packets.

**Status:** `FROZEN_FOR_SEPARATE_ATOMIC_IMPLEMENTATION`

**Provenance:** `dispatcher-enforced` - `gpt-5.6-terra` / `high`; no fallback.
Actual runtime telemetry is not exposed.

**Evidence:** `SOURCE_CONTRACT_AUTHORED / STATIC_NOT_RUN / RUNTIME_NOT_RUN`.

## 1. Outcome, ordering, and non-combination rule

The public Mail resolver becomes usable only when all of these independently
accepted atoms exist, in this order:

1. A2 makes `*MailboxEntRepository` the accepted implementation of the closed
   A1 repository port.
2. A3 provides the isolated authenticated HTTP handler, but does not register
   it or alter public API documentation/client code.
3. D1 publishes the normative OpenAPI operation and its closed wire contract.
4. C1a adds the handwritten typed Mail client for that exact operation.
5. B2 constructs the accepted A1 service from the existing `mailboxSvc` and
   `mailboxRepo`, constructs the accepted A3 handler, and registers it with
   the existing authenticated server API group.

D1 and C1a may be authored only after the A3 contract is frozen, but no atom
may assert a public or runtime endpoint until B2 is accepted. B2 is strictly
`BLOCKED` until both A2 and A3 have accepted source receipts and their public
constructor/port surfaces have not drifted. A2/A3 are therefore the required
predecessors for the public-wiring sequence.

The three implementation atoms below are deliberately not a combined patch.
Each receives one explicit Terra/xhigh implementation dispatch, one
independent Terra/high review, its own exact changed-path inspection, and a
separate commit only after static acceptance. This contract grants no test,
build, formatter, generator, migration, Docker, database, network, runtime,
merge, push, or deployment authority.

The operation is only the Mail side of the bridge:

`POST /api/mail/messages/{messageId}/scheduling-participants`

It must not add Calendar orchestration, a Mail-to-Calendar import, a server
schedule command, NATS subject, worker, engine/provider read, correlation
record, delivery promise, or UI. Calendar availability and event/meeting
creation remain the existing client-directed follow-up paths.

## 2. Cross-atom immutable wire contract

All three atoms use one closed request and response contract. The client never
supplies headers from the message; `addresses` are caller-selected strings
which A1/A2 must prove belong to stored `From`, `To`, or `Cc` under the
authenticated mailbox owner/account/organization scope.

### 2.1 Request

The path ID and body are exactly:

```http
POST /api/mail/messages/8cf3a321-0000-4000-8000-000000000001/scheduling-participants
Content-Type: application/json

{"addresses":["anna@example.com","outside@example.net"]}
```

`messageId` is a required, operation-local path parameter. It is a lower-case,
hyphenated RFC 4122 UUID string with exactly 36 bytes and must not be the nil
UUID. OpenAPI `format: uuid` is descriptive only; the normative requirement is
the exact canonical representation enforced by A3 and reused by C1a through
`normalizeMailMessageID`. The operation accepts no query parameters.

The body is an object with exactly the required `addresses` property and no
other property. It is an ordered, unique array of 1 through 200 strings; every
string is 3 through 320 bytes and is already byte-for-byte canonical under
`mail.NormalizeAddress`. This schema describes the public contract only: A3
enforces strict JSON/media/UUID transport rules, and A1 remains the authority
for cardinality, exact-byte canonicality, uniqueness, and source selection.
No D1 or C1a implementation may lowercase, trim, IDNA-convert, parse display
names, sort, deduplicate, or otherwise rewrite an email address.

### 2.2 Success

A successful response is exactly:

```json
{
  "participants": [
    {"email": "anna@example.com", "user_id": "8cf3a321-0000-4000-8000-000000000002"},
    {"email": "outside@example.net", "user_id": null}
  ]
}
```

`participants` is required, ordered exactly as the request, and has the same
1 through 200 cardinality. Each item has exactly required `email` and
`user_id` properties; `email` echoes the exact already-canonical request
bytes, and `user_id` is either a non-nil canonical UUID or explicit JSON
`null`. The response contains no name, role, organization, source/header,
self flag, status, match reason/count, membership, availability, Calendar
data, message metadata, or partial error.

A null `user_id` intentionally conflates an external, absent, inactive,
cross-organization, ambiguous, or legacy-noncanonical User row. It is not a
directory lookup result and must not be rendered as an authorization promise.

### 2.3 Authentication, authorization, and errors

The operation inherits the document-level bearer authentication and declares
`x-hub-authz: mailbox-owner`. The actor is obtained only by the existing
authenticated Mail middleware/context. Missing or invalid authentication stays
the existing generic `401 unauthorized` behavior. There is no request actor,
owner, account, organization, identity, or participant user-ID field.

After authentication, the public response vocabulary is closed:

| Condition | Status and ErrorEnvelope | Header |
|---|---|---|
| Invalid/noncanonical/nil message ID; any query; invalid strict transport/body; invalid address input | `400` with `error: "invalid mail scheduling request"` | none |
| Missing/foreign message; selected self/Bcc-only/non-header/absent address | `404` with `error: "mail message not found"` | none |
| Mail account deleting | `409` with `error: "mailbox account deleting"` | required `X-Hub-Error-Code: MAIL_ACCOUNT_DELETING` |
| Temporary Mail projection/repository failure | `503` with `error: "mail message unavailable"` | none |
| Unexpected safe failure or malformed internal success | `500` with `error: "mail message unavailable"` | none |

Every error uses the existing closed `ErrorEnvelope` with its required
`request_id`; no route-specific envelope is introduced. The operation must not
document or expose `403`, `415`, `422`, `USER_NOT_FOUND`, `CROSS_ORG`,
`NOT_MEMBER`, a per-address status, a distinguishing error code, or an
availability result. The 404 remains intentionally indistinguishable across
all source-selection and owner-scope misses.

## 3. D1 - normative OpenAPI contract atom

### Exact ownership

D1 may modify only:

`docs/mail/openapi.yaml`

It owns the sole normative public description of this Mail operation. It must
not edit legacy Swag, Go source, TypeScript source, UI, tests, migrations, or
another documentation file.

### Scoped authority-metadata exception

For only `POST /mail/messages/{messageId}/scheduling-participants`, this
frozen contract has task-local priority over the current
`info.x-hub-task-4-api-authority` sentence requiring one reviewed change set
to contain the OAS operation/schema, authenticated backend handler and wiring,
and matching handwritten client. D1 MUST update that authority metadata in
the same `docs/mail/openapi.yaml`-only atom to record this exact scoped
exception and all of the following immutable conditions:

- D1, C1a, and B2 remain three separate, independently reviewed commits;
- `docs/mail/openapi.yaml` remains the sole semantic authority for the
  operation, schemas, authz, errors, compatibility, and operation identity;
- landing or accepting D1 alone does not make the public endpoint accepted;
- public acceptance remains incomplete until D1, C1a, and B2 have all landed
  and one final joint receipt confirms their exact cross-commit alignment; and
- the exception applies to no other Task-4 operation or future atom.

This authority-metadata synchronization is mandatory D1 ownership, not a
fourth atom and not permission for D1 to edit C1a or B2 files. Until the final
joint receipt passes, the route may be individually source-accepted but its
public acceptance and runtime evidence remain incomplete/`NOT_RUN`.

### Required OpenAPI shape

D1 adds exactly one `post` operation under:

`/mail/messages/{messageId}/scheduling-participants`

The operation must have all of the following:

- a stable `operationId: resolveMailSchedulingParticipants`;
- Mail tags consistent with existing Mail operations;
- `x-hub-authz: mailbox-owner` and no auth override or public alias;
- one **local operation parameter** named `messageId`, `in: path`,
  `required: true`, with a closed canonical Mail message UUID schema. That
  schema must constrain the lower-case 8-4-4-4-12 hyphenated form and exclude
  `00000000-0000-0000-0000-000000000000`; its description states that A3
  rejects every noncanonical form before the service call;
- no query/header/cookie parameters;
- a required `application/json` request body with
  `additionalProperties: false`, `required: [addresses]`, a single
  `addresses` array property, `minItems: 1`, `maxItems: 200`,
  `uniqueItems: true`, and string items with `minLength: 3`,
  `maxLength: 320`, and mandatory `x-hub-max-utf8-bytes: 320`. The
  description states that exact canonical-byte validation and the UTF-8 byte
  ceiling are server-authoritative; no raw-body byte extension is added;
- a `200 application/json` response whose schema is closed, requires only
  `participants`, requires 1 through 200 participant objects, and whose closed
  item schema requires exactly `email` and nullable UUID `user_id`. The
  `email` string requires `minLength: 3`, `maxLength: 320`, and mandatory
  `x-hub-max-utf8-bytes: 320`, matching the same 3..320-byte backend
  invariant;
- explicit `400`, `404`, `409`, `503`, and `500` responses aligned verbatim to
  section 2.3. The `409` response alone requires
  `X-Hub-Error-Code` with the single enum value `MAIL_ACCOUNT_DELETING`; and
- no success/error example or description that adds untrusted headers, Bcc,
  body/provider data, user directory detail, Calendar data, a call link, or
  delivery semantics.

D1 may add narrowly named reusable schemas/responses inside the same OpenAPI
file only when they are used solely to represent this immutable contract. It
must use the existing `ErrorEnvelope` rather than duplicate its shape. A broad
generic `MailMessageID` change, mutation of unrelated response components, or
making this route a second authority for Calendar scheduling is out of scope.

### D1 static acceptance

Independent review must prove from the D1-only diff that the normative OAS
path, local canonical parameter, request/response shapes, bearer/authz model,
and all five error cases are exactly aligned to section 2; both address
transport schemas carry the mandatory `x-hub-max-utf8-bytes: 320` extension;
the `info.x-hub-task-4-api-authority` metadata contains only the scoped
single-operation exception above and preserves OAS as sole semantic authority;
no unsupported method/status/field is documented; and no path outside
`docs/mail/openapi.yaml` changed. D1 acceptance alone remains insufficient:
the final joint D1/C1a/B2 alignment receipt is still required. OAS
parsing/generation and all executable evidence remain `NOT_RUN`.

## 4. C1a - handwritten Mail API client atom

### Exact ownership

C1a may modify only:

`frontend/src/api/mail.ts`

It adds only typed DTOs and one typed Mail resolver function. It must not edit
Calendar API code, UI, routes, state, generated clients, OpenAPI, tests, or
server code.

### Required client surface

C1a introduces narrowly scoped immutable-facing types equivalent to:

```ts
export interface MailSchedulingParticipantsRequest {
  readonly addresses: readonly string[];
}

export interface MailSchedulingParticipant {
  readonly email: string;
  readonly user_id: string | null;
}

export interface MailSchedulingParticipantsResult {
  readonly participants: readonly MailSchedulingParticipant[];
}

export async function resolveMailSchedulingParticipants(
  messageId: string,
  request: MailSchedulingParticipantsRequest,
  options: ApiRequestOptions = {},
): Promise<MailSchedulingParticipantsResult>
```

The function calls `normalizeMailMessageID(messageId)` before constructing the
path, then sends exactly one authenticated JSON POST to:

```ts
`/mail/messages/${encodeURIComponent(normalizedMessageID)}/scheduling-participants`
```

It transports only `{ addresses }` and returns only the typed `participants`
receipt. It forwards the existing `ApiRequestOptions` unchanged according to
the established Mail API-client convention. It does not receive or synthesize
actor IDs, Calendar fields, user IDs in the request, headers from the message,
idempotency, a generic event, a meeting, RSVP, availability, or delivery.

The client may make a fresh array/container copy for ownership safety, but it
must preserve every address string byte-for-byte and in order. It must not
normalize, lowercase, trim, parse, validate-as-canonical, deduplicate, sort,
filter, rewrite, or convert email values. The server remains authoritative for
the address acceptance rules and for every error in section 2.3. C1a must not
invent a divergent local error, fallback, or `null`-to-UUID interpretation.

### C1a static acceptance

Independent review must establish: exactly the one client file changed;
`normalizeMailMessageID` is called before path interpolation; the endpoint
path/method/body and nullable `user_id` align exactly to D1/A3; no email byte
is rewritten; no Calendar/UI orchestration leaked in; and no unrelated client
surface changed. Typecheck, unit tests, build, lint, browser, Docker, and
runtime evidence remain `NOT_RUN`.

## 5. B2 - production Mail construction and authenticated registration atom

### Exact ownership and dependency gate

B2 may modify only:

`backend/cmd/server/wiring_calendar_mail.go`

It is `BLOCKED` until accepted A2 proves that `mailboxRepo` satisfies
`MailboxScheduleParticipantsRepository`, and accepted A3 provides
`NewMailboxScheduleParticipantsHandler` plus its authenticated
`RegisterRoutes` surface. If either acceptance receipt, constructor signature,
or port conformance is absent/drifted, B2 returns `BLOCKED`; it must not edit
A1/A2/A3 to make construction convenient.

### Required construction sequence

Within existing `wireCalendarMail`, after the existing successful construction
of `mailboxSvc` and while the same `mailboxRepo` is in scope, B2 adds exactly:

```go
mailboxScheduleParticipantsService, err := mailmod.NewMailboxScheduleParticipantsService(mailboxSvc, mailboxRepo)
if err != nil {
    d.fatal("failed to wire calendar and mail routes", err)
}
mailboxScheduleParticipantsHandler, err := mailmod.NewMailboxScheduleParticipantsHandler(mailboxScheduleParticipantsService)
if err != nil {
    d.fatal("failed to wire calendar and mail routes", err)
}
if err := mailboxScheduleParticipantsHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
    d.fatal("failed to wire calendar and mail routes", err)
}
```

The exact local names may follow surrounding Go style, but the dependencies,
constructor order, `d.api`, `d.requireAuth`, and existing fatal message/pattern
are mandatory. The registration belongs before an early return that would make
the already-authenticated Mail endpoint disappear in the accepted QA/dev
runtime branch; it does not depend on a configured external Mail engine.

B2 must use the existing `mailmod` import and must not add a new import unless
the immutable A3 signature genuinely requires it. It introduces no service
global, feature flag, environment variable, worker, cleanup callback,
background goroutine, metrics, NATS message, engine/provider client,
repository query, Calendar service/handler, availability handler, event or
meeting construction, or cross-domain orchestration.

In particular B2 must not alter existing `AvailabilityHandler.RegisterRoutes`,
Calendar route registration, Calendar `eventSvc`/`meetingSvc` construction,
or `wireCalendarInboundRuntime`/`wireCalendarOutboundRuntime`. B1 owns the
redacted Calendar availability adapter; B2 owns only Mail A3 composition.

### B2 static acceptance

Independent review must show a one-file diff that constructs A1 exactly with
the existing `mailboxSvc` plus `mailboxRepo`, constructs A3 exactly with that
service, registers only through `RegisterRoutes(d.api, d.requireAuth)`, and
uses the surrounding fatal/error pattern for each failure. It must also prove
there is no Calendar, NATS, worker, engine, persistence, OAS, client, UI, or
test change and no new unauthenticated route. Compile, dependency injection
runtime, Docker, and endpoint evidence remain `NOT_RUN`.

## 6. Stop conditions and future work

A D1, C1a, or B2 implementer stops and returns `BLOCKED` rather than widening
its atom when any required behavior needs a second file, an A1/A2/A3 contract
amendment, a different error vocabulary, a changed endpoint schema, Calendar
wiring, external engine behavior, UI orchestration, or executable verification.
A new root-owned frozen amendment is required before crossing any boundary.

After all three atoms are independently accepted, C2 remains a separate Mail
UI dialog/action/route packet and Q1 remains a dedicated Docker-only API/authz
test packet. Neither follows automatically from this public wiring contract.
