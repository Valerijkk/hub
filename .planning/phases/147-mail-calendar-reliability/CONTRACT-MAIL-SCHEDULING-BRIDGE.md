# Mail Scheduling Bridge — Frozen Authority Contract

**AUTHORITY_STATUS:** `FROZEN`
**Phase:** 147 Mail–Calendar Reliability
**Date:** 2026-08-29
**Provenance:** `native-apply_patch`; read-only architecture revision after independent P0 review.
**Evidence status:** `SOURCE_CONTRACT_AUTHORED / RUNTIME_NOT_RUN`.

## 1. Scope and non-goals

This contract lets the Mail UI turn selected, visible message-header addresses into
opaque internal user UUIDs when that is safe, ask Calendar for one redacted
availability predicate, and then call an existing Calendar create endpoint.

It does **not** add a Mail-to-Calendar import, correlation record, server-side
schedule command, provider call, engine body read, Bcc exposure, or new
production/test code. The client orchestrates the three independent API calls.
`pro-qa-ai/**` is outside scope.

## 2. Normative P0 identity and source rules

1. The only candidate address source is the persisted, owner-scoped message
   header projection `From ∪ To ∪ Cc`. Bcc, raw provider/engine snapshots,
   body text, Reply-To, contacts, directory search, and client-supplied header
   values are never candidates.
2. A self address is every active Mail identity (primary and aliases) for the
   owner account **and** the actor's canonical `User.email`. Self addresses are
   removed from the server-derived candidate set before client selection.
3. The request supplies 1..200 address strings. Each must already equal the
   byte-for-byte output of `mail.NormalizeAddress(EngineAddress{Email: value})`;
   normalization is validation, not a rewrite. The set is unique by those exact
   canonical bytes.
4. Before any `User` lookup, the service rejects with one identical generic 404
   only when the message is absent or foreign, or a selected address fails
   source-selection membership: self, Bcc-only, non-header, or absent from the
   server-derived candidate set. It must not reveal which rule failed.
5. A deleting Mail account remains the existing `409` case; a temporary
   projection/repository failure remains the existing generic `503` case. They
   are not folded into the pre-lookup `404`.
6. The resolver performs one bounded set lookup for active users in the actor's
   exact organization. A mapping is emitted only where the requested canonical
   address exactly equals `NormalizeAddress(user.Email).Email` **and** exactly
   equals the stored `user.Email` bytes. Any legacy/noncanonical user row maps
   to `null`, never through `LOWER`, IDNA-only, fuzzy, display-name, CRM, or
   directory matching.
7. Resolver output conveys neither project/channel membership nor availability.
   It has no per-address error/status/reason, match count, organization, name,
   role, source, or timing detail.

The exact-byte `User.email` condition is intentionally fail-closed. A shared
canonical user-email write invariant/backfill is optional for safe fail-closed
A1--B2 deployment and a prerequisite only for complete internal recognition;
until then a legacy row is an email-only guest.

## 3. Endpoint A — Mail participant resolution

### Request

`POST /api/mail/messages/{messageId}/scheduling-participants`

Authentication: the existing authenticated `/api/mail` group. `messageId` is a
canonical non-nil UUID. The body is exactly:

```json
{
  "addresses": ["anna@example.com", "outside@example.net"]
}
```

Schema:

```yaml
type: object
additionalProperties: false
required: [addresses]
properties:
  addresses:
    type: array
    minItems: 1
    maxItems: 200
    uniqueItems: true
    items:
      type: string
      minLength: 3
      maxLength: 320
      description: Already canonical byte-for-byte under mail.NormalizeAddress.
```

`uniqueItems` is only a transport aid; the server repeats exact canonical-byte
uniqueness validation. Names and all user identifiers are forbidden.

### Success response

`200 application/json`:

```json
{
  "participants": [
    {"email": "anna@example.com", "user_id": "8cf3a321-0000-4000-8000-000000000001"},
    {"email": "outside@example.net", "user_id": null}
  ]
}
```

`participants` preserves requested order. Its `email` is the already canonical
request value. `user_id` is a UUID only under rule 6; `null` intentionally
conflates external, absent, inactive, cross-organization, ambiguous, and
legacy-noncanonical addresses.

### Authorization and error contract

The handler obtains actor identity solely from auth context, resolves the active
mailbox owner and organization, and reads the message only under that exact
owner/account/organization scope. It derives candidates from stored headers and
only then makes the bounded active same-organization lookup.

| Condition | HTTP response |
|---|---|
| Invalid UUID, JSON, unknown key, 0/>200 values, duplicate/noncanonical address | `400` with generic `invalid mail scheduling request` |
| Missing/foreign message; self, Bcc-only, non-header, or absent selected address | identical `404` with body `mail message not found`; no distinguishing error code |
| Mail account deleting | existing `409`, `X-Hub-Error-Code: MAIL_ACCOUNT_DELETING` |
| Mail projection/repository temporarily unavailable | `503` generic `mail message unavailable` |
| Unexpected safe failure | `500` generic `mail message unavailable` |

No User lookup occurs in the 404 path. The endpoint never returns `403`,
`USER_NOT_FOUND`, `CROSS_ORG`, `NOT_MEMBER`, a per-address failure, or a timing
classification.

## 4. Endpoint B — Calendar availability predicate

### Request

`POST /api/calendar/v2/availability/check`

```json
{
  "attendee_user_ids": ["8cf3a321-0000-4000-8000-000000000001"],
  "from": "2026-09-01T10:00:00Z",
  "to": "2026-09-01T10:30:00Z",
  "timezone": "Europe/Moscow"
}
```

Schema:

```yaml
type: object
additionalProperties: false
required: [attendee_user_ids, from, to, timezone]
properties:
  attendee_user_ids:
    type: array
    minItems: 1
    maxItems: 49
    uniqueItems: true
    items: {type: string, format: uuid}
  from: {type: string, format: date-time}
  to: {type: string, format: date-time}
  timezone: {type: string, minLength: 1, maxLength: 64}
```

The request accepts no public `slot_minutes` field. The wrapper sets private
server-side `AvailabilityQuery.SlotMinutes = AvailabilityMinSlotMinutes`.

The handler rejects nil UUIDs, duplicates, and a raw actor UUID in
`attendee_user_ids`. It prepends the authenticated actor internally, producing
exactly `[actor] + attendees`; total cardinality is therefore 2..50. It calls
the existing `AvailabilityService` and existing source authorization unchanged.
This endpoint is only a redacting wrapper; it adds no membership rule,
reservation, transaction, or Calendar-to-Mail behavior.

### Success response

`200 application/json` exactly:

```json
{"all_available": true}
```

`all_available` is `true` only when every returned private `Busy` list is empty;
it is `false` otherwise. No interval, user ID, source, free slot, attendee
status, or per-user result is serialized.

### Error contract

| Condition | HTTP response |
|---|---|
| Invalid JSON/key/UUID/actor duplication/cardinality/timezone/interval shape | `400` generic `calendar availability request is invalid` |
| Any target inactive, cross-org, absent, or lacking existing shared project/channel authorization | identical existing `404` `calendar availability not found` |
| Existing bounded-source limit or invalid source | existing `422` with only `AVAILABILITY_LIMIT_EXCEEDED` or `AVAILABILITY_SOURCE_INVALID` |
| Internal source failure | existing generic `500` `calendar operation failed` |

The check is advisory. It is not a slot hold and does not replace create-time
authorization or reservation.

## 5. Client-only orchestration and creation paths

1. Client reads an already authorized message, chooses visible non-self headers,
   calls Endpoint A, and may call Endpoint B using only non-null resolved UUIDs.
   If none resolve, external availability is explicitly unknown and the check is
   skipped.
2. If **every** selected participant has non-null `user_id` and the user chooses
   an authorized project/channel, client calls existing
   `POST /api/calendar/v2/meetings`. Calendar remains authoritative for active
   same-org project/channel membership, availability reservations, idempotency,
   and `409 SLOT_UNAVAILABLE` races.
3. If **any** selected participant has `user_id: null`, client calls existing
   `POST /api/calendar/v2/events`, labels it **Calendar event**, and sends
   **every** selected attendee as email-only (`user_id` absent/null), including
   addresses that Endpoint A happened to resolve. It must not promise a Hub call,
   linked ScheduledCall, reservation, or external invitation delivery.
4. External delivery remains pending/`RUNTIME_NOT_RUN`; successful Calendar event
   creation is not evidence that an outside recipient received an invitation.

There is deliberately no Mail-to-Calendar Go import, no server-side cross-domain
command, and no persistent Mail↔Calendar correlation added by this bridge.

## 6. Planned atomic ownership (not authorization to implement)

| Atom | Exact owned files | Responsibility |
|---|---|---|
| A0 | canonical User-email decision/migration packet, separately approved | Optional for safe fail-closed A1--B2 deployment; prerequisite only for complete exact-byte internal-recipient recognition; no fuzzy fallback. |
| A1 | `backend/internal/mail/mailbox_schedule_participants.go` | Closed DTO, ports, service, generic safe errors. |
| A2 | `backend/internal/mail/mailbox_ent_schedule_participants.go` | Owner/message/header/identity set resolver under the P0 order. |
| A3 | `backend/internal/mail/mailbox_schedule_participants_handler.go` | Authenticated route and exact HTTP mapping. |
| B1 | `backend/internal/calendar/availability_check.go` | Redacting availability wrapper only. |
| B2 | `backend/cmd/server/wiring_calendar_mail.go` | Instantiate/register A3 and B1; no behavior changes elsewhere. |
| C1a | `frontend/src/api/mail.ts` | Typed Mail scheduling-participant resolver client and DTOs. |
| C1b | `frontend/src/api/calendar.ts` | Typed redacted availability-check client and DTOs; retain existing generic-event and meeting-creation clients. |
| C2 | `frontend/src/modules/mail/mail-scheduling-dialog.tsx` (new), `frontend/src/modules/mail/mail-message-actions.tsx`, `frontend/src/routes/_authenticated/mail/$folderKey/thread/$threadId.tsx` | The new dialog owns client orchestration: select header candidates, resolve, check redacted availability, then create a generic event or authoritative meeting. `mail-message-actions.tsx` opens it; the thread route supplies the visible message header projection. |
| D1 | `docs/mail/openapi.yaml` and affected Mail/Calendar docs | Public contract sync. |
| Q1 | later dedicated test packet only | Docker-only API/authz/negative/race coverage; not part of this source contract. |

Each production atom requires its own frozen task contract, explicit Terra/xhigh
implementation dispatch, independent review, and static acceptance before the
next atom. Tests remain a later Docker-only packet; this document grants no host
or Docker execution authority.

## 7. Acceptance conditions and evidence limits

Static acceptance must prove:

- no candidate path reads Bcc/raw snapshot/body or accepts client headers;
- all active mailbox identities/aliases plus actor canonical user email are
  excluded before User lookup;
- every invalid source-selection case reaches the identical pre-lookup 404;
- exact canonical-byte User email equality is enforced and legacy rows yield
  only `null`;
- availability request accepts no slot field and the private wrapper sets
  `SlotMinutes = AvailabilityMinSlotMinutes`; its response is exactly
  `all_available`, derived only from private `Busy` lists, and does not
  serialize raw `AvailabilityResult` data;
- create remains client-selected between existing meeting and generic-event
  routes, with no Mail→Calendar import/correlation/server command;
- mixed recipients always use email-only generic attendees and UI says
  `Calendar event` / delivery pending.

Required future runtime coverage: owner/foreign/self/Bcc/non-header/noncanonical
negative cases, same-org active mapping, legacy canonical mismatch, redacted
availability response, member authorization non-enumeration, create-time
`SLOT_UNAVAILABLE`, and mixed-external generic-event behavior. All executable,
generation, migration, database, Docker, and runtime evidence for this contract
is currently **NOT_RUN**.
