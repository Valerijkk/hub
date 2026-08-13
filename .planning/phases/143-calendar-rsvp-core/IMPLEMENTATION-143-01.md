# Task 143-01 — static acceptance evidence

> **Status:** STATIC_ACCEPTED · RUNTIME_NOT_RUN · not product-accepted
> **Implementation commit:** `e7d5f450d112158369de60738f95e0f5e29ace53`
> **Parent:** `c3430e7c75ca691973368a3d5c7d0fcf7770b0ea`
> **Reviewed semantic contract:** `5a351d5bd51bd215775f3828bc26979cf985a4ac`

## Scope and provenance

- Authoring packet: `/root/mail_143_01_terra`.
- Requested and actual authoring route: exact `gpt-5.6-terra` / `xhigh`; no fallback.
- Independent review packet: `/root/mail_143_01_review` in a distinct context.
- Requested and actual review route: `gpt-5.6-terra` / `high`; no fallback.
- Owned product/test files:
  - `backend/internal/calendar/event_types.go`
  - `backend/internal/calendar/event_policy.go`
  - `backend/tests/calendar/event_policy_test.go`

## Accepted behavior

- strict timed/all-day temporal union and exclusive all-day end-date validation;
- named IANA timezone validation;
- closed RSVP status set;
- pure event-update classification for event version, iTIP sequence, attendee RSVP reset,
  notifications and cancellation intent;
- distinct whole-event and removed-attendee cancellation intent;
- mixed-update sequence invariant: an accepted mutation increments sequence at most once;
- external-package table units with hostile temporal, enum, policy-row and mixed-flag cases.

The implementation follows the exact MAIL-SPEC §4.2 writer table for recurrence/EXDATE:
reset the affected non-organizer scope. The broader ADR prose is restored in this evidence
commit to the exact independently reviewed semantic blob `b7bb9535c466e3115b1ffafa19893e818af23bc7`;
that restoration changes no accepted source blob. `OrganizerChanged` was not invented because
the frozen PATCH surface has no organizer mutation and inbound organizer identity is immutable.

## Review lifecycle

The first independent review found one P0: a boolean cancellation output conflated whole-event
CANCEL with cancellation only for removed attendees. It also found P2 gaps in hostile-date and
mixed-update coverage. The author replaced that boolean with typed `CancelIntent`, added the
missing cases, and returned the same three owned files for re-review. The follow-up independent
review returned PASS with no P0, P1 or P2.

Two proposed review changes were rejected by root against the frozen contract: resetting every
attendee for recurrence edits contradicted the exact MAIL-SPEC writer table, and exposing an
organizer-change flag contradicted the immutable/public API surface.

## Exact reviewed source blobs

| Path | Blob |
|---|---|
| `backend/internal/calendar/event_types.go` | `6169cd1dce828c4b2bc29a80e4170cb61c171d38` |
| `backend/internal/calendar/event_policy.go` | `58216751d4656a28dc77acb191981686973986a4` |
| `backend/tests/calendar/event_policy_test.go` | `09b1dc84fc478d4bb7653cfa95bcacc3922b25e2` |

## Verification verdict

- allowed-path/source review: PASS;
- `git diff --cached --check`: PASS before the implementation commit;
- protected-path inventory: zero changes;
- secret and forbidden AI/coauthor-trailer scan: zero findings;
- forbidden I/O, wall-clock and global-state scan for the genuine units: zero findings;
- executable tests, build, lint, validators and generation: `NOT_RUN`;
- accepted Docker launcher: `NOT_READY`.

Docker Desktop availability does not make direct Docker/Compose an admissible fallback. This
packet is therefore `STATIC_ACCEPTED`, not runtime-verified, merge-ready, complete, stable or
production-ready. The next dependency-safe packet is 143-02A.
