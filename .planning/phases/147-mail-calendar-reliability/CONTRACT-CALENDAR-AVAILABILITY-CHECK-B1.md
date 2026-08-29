# Calendar Availability Check B1 — Frozen Atomic Contract

**AUTHORITY_STATUS:** `FROZEN`
**Phase:** 147 Mail–Calendar Reliability
**Date:** 2026-08-29
**Provenance:** `dispatcher-enforced` — `gpt-5.6-terra` / `high`; actual runtime telemetry is not exposed.
**Evidence status:** `SOURCE_CONTRACT_AUTHORED / STATIC_NOT_RUN / RUNTIME_NOT_RUN`.

## 1. Outcome and fixed boundary

B1 adds exactly one authenticated, redacting predicate:

`POST /api/calendar/v2/availability/check`

It accepts a proposed time range and 1..49 attendee user UUIDs, injects the
authenticated actor as the first private query user, delegates unchanged to
the existing `AvailabilityService`, and returns only whether every resulting
private busy list is empty. It is advisory; it creates no hold, reservation,
event, call, Mail record, membership rule, transaction, or side effect.

This is a public Task-4 HTTP atom. `docs/mail/QA-HANDOFF.md` and
`docs/system_specification.md` require one reviewed three-way receipt for each
such atom: normative OAS, authenticated handler plus route wiring, and a
matching handwritten typed client. Therefore B1 owns all three below. It
MUST NOT be split into a backend route followed later by OAS/client work: that
would expose an undocumented public route and violate the authority gate.

The existing `AvailabilityHandler` is already instantiated and registered by
`backend/cmd/server/wiring_calendar_mail.go`. B1 adds its route in that
handler's existing `RegisterRoutes`; no command-server wiring change is
needed or allowed for this atom. B1 supersedes the Calendar availability
registration/wiring portion formerly assigned to scheduling-bridge B2. B2,
if it remains necessary, later owns only Mail A3 instantiation/registration;
it owns no Calendar availability registration.

## 2. Exact owned files

| File | Change and sole responsibility |
|---|---|
| `backend/internal/calendar/availability_check.go` (new) | B1-only closed request/response DTOs, actor-injected query constructor, redacting handler method, and its safe 400 mapper. |
| `backend/internal/calendar/availability_handler.go` | Add only `calendarV2.Post("/availability/check", handler.CheckAvailability)` beside the existing availability route. No change to `QueryAvailability` or shared error mapping. |
| `backend/internal/calendar/event_handler.go` | Extend only `calendarHTTPValidateExactKeys` to recognise `*calendarHTTPAvailabilityCheckRequest`; delegate its allowed-key predicate to B1. This preserves the established duplicate-key, unknown-key, nesting, token, body-size, and trailing-value guard. |
| `docs/mail/openapi.yaml` | Add the sole normative OAS operation, schema, authz, and safe responses for `/calendar/v2/availability/check`. Do not modify legacy Swag files. |
| `frontend/src/api/calendar.ts` | Add B1 typed request/response interfaces and its transport-only wrapper. Keep the existing detailed `/calendar/v2/availability` client intact. |

No other file is owned. In particular B1 MUST NOT edit
`backend/internal/calendar/availability_service.go`,
`availability_repository.go`, `availability_policy.go`, Ent/schema/migration
files, `backend/cmd/server/wiring_calendar_mail.go`, Mail, meeting/event
creation, or UI modules. It creates no tests in this atom.

## 3. Backend shape

`availability_check.go` owns these unexported transport types and behavior:

```go
type calendarHTTPAvailabilityCheckRequest struct {
    AttendeeUserIDs calendarHTTPField[[]uuid.UUID] `json:"attendee_user_ids"`
    From            calendarHTTPField[time.Time]   `json:"from"`
    To              calendarHTTPField[time.Time]   `json:"to"`
    Timezone        calendarHTTPField[string]      `json:"timezone"`
}

type calendarHTTPAvailabilityCheckResponse struct {
    AllAvailable bool `json:"all_available"`
}

func (handler *AvailabilityHandler) CheckAvailability(c fiber.Ctx) error
func (request calendarHTTPAvailabilityCheckRequest) queryForActor(actorID uuid.UUID) (AvailabilityQuery, error)
func calendarHTTPAvailabilityCheckKeyAllowed(key string) bool
```

`CheckAvailability` follows the established handler order: readiness;
`calendarHTTPActor(c)`; `calendarHTTPDecode(c, &request)`; request-to-query
validation; then `handler.service.QueryAvailability(c.Context(), actorID,
query)`. Decode or shape failures are remapped to B1's generic 400 rather
than returned verbatim. It calls no repository and implements no authz.

`queryForActor` requires every field to be present and non-null. It rejects a
nil actor, 0 or more than 49 submitted attendees, every nil attendee UUID,
duplicates, and any submitted UUID equal to the actor. It constructs a fresh,
ordered private `UserIDs` slice exactly as `[actorID] + attendee_user_ids`, so
the existing source-query limit is 2..50 and the existing service/repository
authorization applies to the complete set unchanged. It sets:

```go
AvailabilityQuery{
    UserIDs:     injectedIDs,
    From:        request.From.value,
    To:          request.To.value,
    Timezone:    request.Timezone.value,
    SlotMinutes: AvailabilityMinSlotMinutes,
}
```

There is no public `slot_minutes`; an absent, `null`, or unknown field is a
400. `validateAvailabilityQuery` remains the authority for canonical time,
range, timezone, slot, duplicate and source-query validation. B1 must not
copy or weaken it.

On a successful service result, B1 computes `all_available` only by checking
that `len(user.Busy) == 0` for every element of the private `result.Users`.
It never serializes `AvailabilityResult`, `Users`, `UserID`, `Busy`, `Free`,
intervals, sources, slot suggestions, per-user status, or count. The only
success representation is `calendarHTTPAvailabilityCheckResponse`.

The small exact-key extension in `event_handler.go` is required because
`calendarHTTPDecode` otherwise rejects an unknown DTO type. It must add the
B1 case only; B1's allowed top-level set is exactly
`attendee_user_ids`, `from`, `to`, and `timezone`. Existing strict JSON
validation already rejects duplicate keys and non-object/trailing payloads;
B1 must reuse it rather than add a looser decoder.

## 4. Closed public contract

The OAS operation is `checkCalendarAvailability`, tagged `CalendarEvents`,
with `x-hub-authz: calendar-availability-shared`. It is the only declaration
of this new route. Its request is exactly:

```json
{
  "attendee_user_ids": ["8cf3a321-0000-4000-8000-000000000001"],
  "from": "2026-09-01T10:00:00Z",
  "to": "2026-09-01T10:30:00Z",
  "timezone": "Europe/Moscow"
}
```

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
  timezone:
    type: string
    minLength: 1
    maxLength: 64
    x-hub-validation: iana-timezone
```

The `200` body is closed and exactly:

```json
{"all_available": true}
```

Its schema has `additionalProperties: false`, requires `all_available`, and
contains no other property. It declares no `slot_minutes`, source, interval,
free-slot, user-ID, or per-attendee field. Every non-`200` OAS response uses
the existing `ErrorEnvelope` through these exact response references:

```yaml
"400": {$ref: "#/components/responses/Error"}
"404": {$ref: "#/components/responses/NotFound"}
"422": {$ref: "#/components/responses/AvailabilityUnprocessable"}
default: {$ref: "#/components/responses/Error"}
```

`frontend/src/api/calendar.ts` adds only:

```ts
export interface CalendarAvailabilityCheckRequest {
  attendee_user_ids: string[];
  from: string;
  to: string;
  timezone: string;
}

export interface CalendarAvailabilityCheckResponse {
  all_available: boolean;
}

export function checkCalendarAvailability(
  request: CalendarAvailabilityCheckRequest,
): Promise<CalendarAvailabilityCheckResponse> {
  return api.post<CalendarAvailabilityCheckResponse>(
    "/calendar/v2/availability/check", request,
  );
}
```

It neither performs actor injection nor exposes a slot field. It does not
replace or alter `getCalendarAvailability`; the two endpoints intentionally
have different privacy surfaces.

## 5. Exact safe errors
Every non-`200` B1 outcome uses the standard `ErrorEnvelope` exactly as
`{"error": "...", "request_id": "..."}`. In the table, a stated message is
the exact `error` value only; it does not suppress the standard `request_id`
envelope property. The success response remains the sole one-property
representation: `{"all_available": boolean}`.

| Condition | Status and body/header |
|---|---|
| Empty/malformed/non-object/trailing/duplicate JSON; unknown key; missing/null field; invalid/nil/duplicate attendee UUID; actor submitted; 0/>49 attendees; invalid temporal range or timezone | `400`; `error` is exactly `calendar availability request is invalid`; no error code or field detail |
| Existing target absence, inactive/cross-organization status, or absent existing shared project/channel authorization | existing identical `404`; `error` is exactly `calendar availability not found`; no target distinction |
| Existing bounded-source limit | existing `422`; `error` is exactly `calendar availability cannot be resolved`; header `X-Hub-Error-Code: AVAILABILITY_LIMIT_EXCEEDED` only |
| Existing invalid source | existing `422`; `error` is exactly `calendar availability cannot be resolved`; header `X-Hub-Error-Code: AVAILABILITY_SOURCE_INVALID` only |
| Readiness failure or other internal failure | existing generic `500`; `error` is exactly `calendar operation failed`; no raw cause |

B1 reuses `calendarHTTPAvailabilityError` for service errors so its existing
404/422/500 mapping remains the sole mapping authority. The B1-specific 400
helper must not change `calendarHTTPBadRequest`, because existing Calendar
routes retain their `invalid calendar request` contract. The implementation
and OAS must preserve the ordinary `ErrorEnvelope.request_id` on every error.

## 6. Acceptance and evidence

Static acceptance for the future implementation must establish all of the
following from the immutable B1 diff:

- the only added route is authenticated `/calendar/v2/availability/check`;
  the existing handler's registration is the only wiring change;
- request JSON is strict and closed; it has 1..49 UUID attendees, rejects
  nil/duplicates/actor, and has no accepted `slot_minutes`;
- the actor is taken only from `calendarHTTPActor`, prepended privately, and
  every service invocation has exactly 2..50 distinct ordered IDs;
- `AvailabilityService`, its repository, and existing source authorization
  are unchanged;
- private `AvailabilityQuery.SlotMinutes` is exactly
  `AvailabilityMinSlotMinutes` and the response is derived only from every
  private `Busy` list being empty;
- source data, intervals, IDs, free spans, per-user results, and raw errors
  cannot reach the success response or `ErrorEnvelope.error`; the standard
  `request_id` remains permitted and required on errors;
- the same reviewed change contains the normative OAS operation and the
  matching handwritten frontend wrapper; no legacy Swag file is changed.

Required evidence at contract handoff is a read-only changed-path inspection
and `git diff --check` for the five owned paths. This contract itself runs no
test, generator, formatter, host validator, Docker command, migration, or
runtime process. Future executable behavior belongs to a separately frozen
Task-4 API packet under `testframework/tests/api/**`, Docker-only, covering
success, actor injection, all listed negative/authz cases, 422 boundaries,
and redaction. Until that packet runs, runtime status remains `NOT_RUN`.

## 7. Explicit non-goals and stop conditions

B1 does not create a meeting/event, reserve a slot, call Mail, resolve email,
alter existing detailed availability, add a frontend UI, or assert external
invite delivery. Any request to change `AvailabilityService` authorization,
the route authority, the OpenAPI source of truth, the 49+actor cardinality,
or the one-boolean response is outside this atom and requires a newly frozen
contract. An unavailable required dependency, contrary source authority, or
missing three-way review is `BLOCKED`, not an invitation to silently broaden
B1.
