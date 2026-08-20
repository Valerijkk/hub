# 143-03B-T — immutable CalendarAttendee tombstone source receipt

## Status

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable source commit
`68cbf7cb227799304ce679937b519b66cb6abf9b`. This uncommitted documentation working tree is not
source or dispatch authority; the conditional 143-04C dispatch gate below requires its own exact19
commit and immutable review.

| Field | Immutable value |
|---|---|
| Commit | `68cbf7cb227799304ce679937b519b66cb6abf9b` |
| Parent | `4715f939f24e5b5c2d3f80a38d204babe68bdd34` |
| Tree | `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e` |
| Subject | `feat(calendar): add attendee tombstones` |
| Source author | Terra/xhigh |
| Provenance | dispatcher-enforced, no fallback; telemetry `UNKNOWN` unless exposed |
| Reviews | Contract/high, Security/xhigh, ABI/Lifecycle high; precommit and immutable exact-SHA PASS |
| Review findings | C/I/M `0/0/0` |
| Runtime | generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; launcher `NOT_READY` |

## Exact test-first manifest and blobs

No substitute, additional path or reordering belongs to this source receipt:

1. `backend/tests/calendar/calendar_attendee_schema_test.go` — `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc`
2. `backend/ent/schema/calendarattendee.go` — `3c38799f3e56089e2e264b000ee0951a144b0257`

## Accepted static schema contract and limits

- `CalendarAttendee` has exactly the 13 frozen fields listed by the authored-first oracle. The only
  new field is `removed_at`: Ent time, optional and nillable, with no default, update-default,
  immutable or unique flag, storage-key override, schema-type override, validator or annotation.
- The schema retains exactly two unique attendee indexes. The series/master index fields are
  `calendar_event_id, normalized_email` with predicate
  `recurrence_id IS NULL AND removed_at IS NULL`. The occurrence index fields are
  `calendar_event_id, normalized_email, recurrence_id` with predicate
  `recurrence_id IS NOT NULL AND removed_at IS NULL`. Each index has no edge, no storage-key
  override and exactly one `IndexWhere` annotation.
- Existing scalar limits remain unchanged, including normalized email max 320, display name max 160
  and recurrence ID max 255. No service filtering, re-invite identity, replacement semantics,
  adapter, generated Ent code, migration or runtime behavior is accepted by this exact2 receipt.

## Exact19 conditional 143-04C dispatch gate

Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED /
IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Every clause in the prior exact18 packet saying 143-03B-T is
only source-authorized, not accepted, dirty, uncommitted or unaccepted, or saying 143-04C is blocked
pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. The immutable
03B-T source receipt above is current.

Root dispatch of only the frozen 143-04C exact3 becomes effective after one atomic exact19 docs
commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that exact commit:

1. `backend/tests/calendar/event_service_update_test.go`
2. `backend/internal/calendar/event_repository.go`
3. new `backend/internal/calendar/event_service_update.go`

Until that exact19 gate passes, 143-04C exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED`
and root dispatch is not effective. No alternate, extra or reordered path is authorized.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` only at
`4715f939f24e5b5c2d3f80a38d204babe68bdd34`; its exact4 receipt remains
`IMPLEMENTATION-144-04F.md`. Mail 144-04G still requires separate docs authority because its
repository port is absent. Calendar 143-04A/B/G receipts remain accepted; 143-03E, 144-02E, Ent
generation, migration, launcher readiness and all executable/runtime evidence remain blocked,
`NOT_READY` or `NOT_RUN`. No adapter, raw iTIP, service/API/UI, product, phase or release authority
follows from this receipt.
