# TASK-236 — Calendar RSVP core

> Date: **2026-08-13** · Priority: **P0** · Status: **Open** · Phase: **143**

## Outcome

Introduce first-class calendar events/attendees/occurrence exceptions, deterministic
RFC 5545/5546 generation/parsing, optimistic RSVP, and the optional ScheduledCall link.

## Owned paths

- `backend/ent/schema/calendar*.go`
- mail-owned generated Ent/migration artifacts required by those schemas
- `backend/internal/calendar/**`
- `backend/tests/calendar/**`
- calendar API additions in backend wiring/docs

## Acceptance

- Public constructors support fake clock and pure unit testing.
- Existing active call creator/participants backfill to `accepted` per ADR policy.
- REQUEST/UPDATE/CANCEL/REPLY, SEQUENCE, RECURRENCE-ID, UID idempotency, TZ/DST and
  all-day behavior match `MAIL-SPEC.md`.
- RSVP is version-checked, sender-bound for inbound REPLY, outbox-backed, and realtime-ready.
- No database, broker, listener, filesystem, Docker, or private export hook is used by
  genuine units under `backend/tests/calendar/**`.
- Relevant calendar/system/architecture docs are synchronized in the same commits.

## Blockers

No runtime blocker for pure/domain work. Migration/runtime verification needs the accepted
Docker launcher. A schema-generation command may run only in Docker.
