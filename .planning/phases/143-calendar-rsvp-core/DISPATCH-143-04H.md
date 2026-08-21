# Calendar 143-04H occurrence-cancel source dispatch

> **Authority acceptance-sync precedence:** `AUTHORITY_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This dirty exact15
> sync is non-authoritative. Candidate `f2966bdd` has immutable Contract Terra/high, Security
> Terra/xhigh and Lifecycle Terra/high exact-SHA PASS C/I/M `0/0/0`, but may become
> `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` only after one atomic exact15 sync commit is created
> and that acceptance commit receives immutable PASS.
> Only then are the lower candidate/precommit/exact-review-next routes `CONSUMED/HISTORICAL` and
> root may dispatch the ordered exact3. Until then this dispatch remains blocked.


> **Predecessor candidate status:** `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED /
> ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> This dirty exact15 carrier is non-authoritative. Source dispatch is forbidden until the
> full candidate -> exact-SHA review -> separate acceptance-sync lifecycle in
> `REVIEW-143-04H-OCCURRENCE-CANCEL-AUTHORITY.md` is complete.

## Outcome and exact ownership

An authenticated Hub organizer may cancel exactly one canonical occurrence of an active
recurring event. The command creates or reuses one cancelled exception, increments the master
version and sequence once, and atomically persists one safe content-free
`calendar.invite.cancelled.v1` `EnvelopeV1`. It does not create raw iTIP or transport work.

Future source ownership is exactly this ordered test-first manifest:

1. NEW `backend/tests/calendar/event_service_occurrence_cancel_test.go`
2. MODIFY `backend/internal/calendar/event_repository.go`
3. NEW `backend/internal/calendar/event_service_occurrence_cancel.go`

No source currently exists or is authorized. Do not modify `event_service_update.go`,
`event_call_policy.go`, `event_call_due_service.go`, Ent, adapters, handlers, wiring, generated
code, migrations, OpenAPI implementation, frontend or the legacy calendar aggregate.

## Frozen boundary

- Add receipt-free `EventMutationCalendarOccurrenceCancel = "event_occurrence_cancel"`; never add
  a `MutationReceiptOperation` value.
- Command ingress checks recurrence byte length 1..255 before UTF-8/control/canonical scans and
  before every repository/clock/ID call; invalid UTF-8/control is rejected, accepted bytes are
  preserved, and temporal canonicalization waits for the authoritative master.
- The optional `CalendarEventOccurrenceCancelTransactionBackend` capability is checked before
  `ResolveActorOrganization` or any other backend call. `WithEventTransaction` is the only outer
  seam allowed before the capability check.
- After the owner/org-scoped resolve, bind only fixed-shape EventID/org/OwnerID/OrganizerID first.
  Any mismatch is the same non-disclosing `ErrEventNotFound` before validating, capping, allocating
  or traversing untrusted current strings/slices/attendees/exception. Only an identity-bound
  snapshot receives full validation; master mutability and recurrence membership precede version.
- Only active confirmed/tentative Hub masters with a valid normalized 04B recurrence are mutable.
  Non-recurring, external, cancelled, malformed, non-member and EXDATE instances fail closed as
  `not_mutable`.
- Membership is canonical and bounded independently of EXDATE: all-day recurrence ids are
  `YYYY-MM-DD`, timed ids are UTC-second `YYYY-MM-DDTHH:MM:SSZ`, and expansion/work is capped at
  1000 occurrences. Existing accepted 04B DAILY/WEEKLY/MONTHLY, INTERVAL,
  BYDAY/BYMONTHDAY, COUNT xor UNTIL and DST semantics are reused without broadening.
- Matching already-cancelled exception plus current expected version is an idempotent read with
  zero clock/ID/persist/outbox calls. A stale version still conflicts.
- New mutation changes only master `Version` and `Sequence` by checked `+1`; writes one cancelled
  exception with nil temporal override and the same new sequence; leaves master status, recurrence,
  EXDATE, content, links, reminders and attendee state unchanged.
- The occurrence payload contains only event id, stable `hub-<event-id>` UID, new sequence,
  canonical non-null recurrence id and sorted unique active attendee ids. The organizer's UserID
  is excluded; external attendees remain. The JSON array is non-null and capped at 200.
- Persistence is one expected-version CAS and must echo the complete master, exception and exact
  frozen envelope. Any mismatch or raw/unknown error becomes content-free `ErrEventInvalid` and
  rolls the transaction back.
- `EventSerializer` must receive zero calls. No `METHOD:CANCEL`, MIME, provider, delivery or raw
  iTIP data belongs to this packet.

## Required non-tautological oracles

The authored-first external-package unit must reflection/compile-freeze the complete ABI in the
review and prove: optional capability zero-backend-call; scope operation and forged/escaped/closed
rejection; nondisclosing identity masking before version; full recurrence/DST/boundary/cap matrix;
EXDATE not used as membership proof; stale versus idempotent cancelled behavior; checked overflow;
exact ID/clock call order; attendee projection and empty-array JSON; authorization/getter alias
resistance; every master/exception/envelope echo field tamper; exact EnvelopeV1 JSON and subject
bindings; foreign malformed/oversized identity masking; 1/255/256-byte and 1 MiB recurrence
preflight; raw-error masking; behavioral serializer zero-call; and rollback at every failure point.
Static Security/Lifecycle review owns source-token and no-I/O scans.

## Dependencies and blockers

Hard static dependencies: accepted 143-04A, 143-04B, 143-04C, 143-04G and 143-04L.
Accepted 143-04J/04K remain independent and are not invoked. Legacy 143-02A serializer dependency,
raw `METHOD:CANCEL`, `EventSerializer` and source-first wording are `CONSUMED/HISTORICAL`.
143-04F-H owns the eventual Ent transaction only and remains blocked. Launcher is `NOT_READY`;
tests/build/lint/validators/Docker/runtime are `NOT_RUN`; AQ migration/backfill blockers remain.

## Provenance and lifecycle

Authoring provenance is `gpt-5.6-terra` / `xhigh`, dispatcher-enforced, no fallback, telemetry
`UNKNOWN`. Required independent reviews are Contract/Spec/Oracle Terra/high,
Security/Adversarial Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high.

No downstream action is authorized until, in order: triple precommit PASS; one atomic exact15
candidate commit; triple immutable exact-SHA PASS; a separate exact15 docs-only acceptance sync;
and immutable review of that acceptance commit. Only then may root dispatch the ordered exact3.
