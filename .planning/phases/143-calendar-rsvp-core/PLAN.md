# Phase 143 — Calendar RSVP Core

> Task: TASK-236 · Status: CONTRACT_0_1_6_ROOT_REVIEW_PENDING_CODE_ROUTE_AUTHORIZED · Depends on: Phase 142

## Goal

Add the canonical calendar event/attendee model and deterministic iTIP/RSVP domain, then
connect it to the existing calendar and ScheduledCall surfaces without two write
authorities, duplicate meetings or dependence on the real mail engine.

## Task 143-01 — Freeze public domain types and failing policy units

**Create:**

- `backend/internal/calendar/event_types.go`
- `backend/internal/calendar/event_policy.go`
- `backend/tests/calendar/event_policy_test.go`

**Steps:**

1. Write external-package table tests for validation of timed/all-day shapes, valid IANA
   timezone, RSVP transitions, material vs non-material edits, sequence increments,
   reminder-only version behavior, and attendee reset/preserve policy.
2. Add public immutable input/output types and pure policy functions with injected clock
   only where time is required.
3. Run the focused Go unit job only through the accepted Docker launcher. If unavailable,
   leave evidence `NOT_RUN` and perform static review only.
4. Commit `test(calendar): зафиксировать политику событий и RSVP` including docs sync.

## Task 143-02 — Deterministic iTIP codec with fixtures

**Create:**

- `backend/internal/calendar/itip.go`
- `backend/internal/calendar/itip_types.go`
- `backend/tests/calendar/itip_test.go`
- `backend/tests/calendar/testdata/*.ics`

**Modify:** `backend/internal/calendar/ical.go` only to reuse safe shared primitives without
breaking existing export behavior.

**Steps:**

1. Add failing tests for REQUEST, update, CANCEL, REPLY, sequence, UID, recurrence id,
   folded lines, attendee parameters, timezone/DST, all-day exclusive end, duplicate
   delivery key, and parser bounds.
2. Implement deterministic serialize/parse with injected clock and explicit limits.
3. Keep METHOD and content method consistent; never use `time.Now` internally.
4. Focused Docker unit verification when available.
5. Commit `feat(calendar): добавить детерминированный iTIP-кодек`.

## Task 143-03 — Ent schemas and relation contracts

**Create:**

- `backend/ent/schema/calendarevent.go`
- `backend/ent/schema/calendarattendee.go`
- `backend/ent/schema/calendareventexception.go`
- `backend/ent/schema/calendarmutationreceipt.go`
- `backend/ent/schema/calendarinboundreplyreceipt.go`

**Modify:**

- `backend/ent/schema/user.go`
- `backend/ent/schema/scheduledcall.go`

**Steps:**

1. Add UUID/timestamp/soft-delete/version fields, enums, indexes, owner/organizer/attendee
   edges, master UID unique constraint, exception recurrence unique constraint, the two
   nullable-attendee partial unique constraints, optional one-to-one call link, and a
   bounded create/RSVP HTTP idempotency receipt. Add the separate bounded inbound-REPLY
   receipt ledger unique by attendee + event sequence + canonical sender-bound payload
   hash; it never stores raw MIME or an HTTP idempotency key.
2. Add schema-level annotations/checks supported by current Ent patterns.
3. Run Ent generation only in accepted Docker execution; inspect generated diff before
   staging. Never generate on host.
4. Add/adjust schema contract units only in `backend/tests/calendar/**` if they remain
   genuine service-free units; otherwise hand integration checks to QA.
5. Commit `feat(calendar): добавить схемы событий и участников`.

## Task 143-04 — Repository and transactional RSVP service

**Create:**

- `backend/internal/calendar/event_repository.go`
- `backend/internal/calendar/event_service.go`
- `backend/internal/calendar/event_errors.go`
- `backend/internal/calendar/event_call_policy.go`
- `backend/internal/calendar/event_call_due_service.go`
- `backend/tests/calendar/event_service_test.go`
- focused ScheduledCall projection/due genuine units under `backend/tests/calendar/**`

**Modify:** existing calendar DTO/repository/service files only through narrow composition.

**Steps:**

1. Define repository and outbox/notifier ports usable by pure fakes.
2. Write failing service tests for create, material/silent update, cancellation, series/
   occurrence RSVP, stale version returning the current version/refetch metadata, same-key
   response replay, changed-payload key conflict and spoofed reply. Cover series and
   occurrence `REPLY A -> B -> delayed A`, same-status receipt and the bounded ninth-hash
   quarantine so delivery order cannot roll RSVP backward.
3. Implement minimal transaction orchestration and version conflict mapping.
4. Freeze pure call projection mapping and next-effective-occurrence policy from canonical
   RRULE/EXDATE/exceptions, then implement separate event-before-call Ent transactions for
   paired create/link/edit/delete and occurrence-keyed due advancement. Require active exact-
   channel creator plus active same-project/same-organization internal participants. The
   call participant edge mirrors only series-level non-organizer
   `needs_action|accepted|tentative`; occurrence RSVP never changes it.
5. Ensure notification/NATS payload contains no description/body; the existing safe call
   title may remain in `calls.channel.{channelId}.scheduled_due.v1`, but the worker publishes
   only through outbox. The obsolete unconsumed `calls.scheduled.due` subject is removed.
6. Focused Docker unit verification when available.
7. Commit `feat(calendar): реализовать сервис RSVP и событий`.

## Task 143-05 — HTTP contract and aggregate integration

**Modify:**

- `backend/internal/calendar/dto.go`
- `backend/internal/calendar/handler.go`
- `backend/internal/calendar/service.go`
- `backend/internal/calendar/repository.go`
- `backend/cmd/server/wiring_modules.go`
- applicable route wiring file discovered before edit
- narrow `backend/internal/calls/{calendar_bridge,scheduled,handler_scheduled,dto,repository}.go`
- `backend/cmd/server/wiring_integrations.go`
- `frontend/src/api/calendar.ts` only for additive `event` DTO compatibility if needed
- `frontend/src/modules/calls/call-calendar.tsx` only in its later canonical-reader packet
- `docs/system_specification.md`
- `docs/architecture.md`

**Create:** `backend/tests/calendar/handler_contract_test.go` only if it stays an in-process
public-handler unit with fakes and no listener.

**Steps:**

1. Add failing handler units for auth-derived owner, timed/all-day validation, NFR-13 409
   plus error-code/ETag/Location headers, and recurrence id decoding.
2. Implement canonical endpoints exactly under `/api/calendar/v2/**` as frozen in
   `docs/mail/openapi.yaml`.
3. Delegate legacy ScheduledCall create/update/delete to the canonical bridge while keeping
   successful response fields/routes compatible. Legacy PUT serializes against the locked
   current event; `scheduled_call_id` is creation-only and server-authorized. Attach is an
   exact normalized match, not an overwrite; linked-call external guests fail closed under
   founder question 23. The legacy POST idempotency compatibility break activates only when
   both shipped clients use one stable key.
4. Preserve exact `GET /api/calendar/events` `{events,from,to}` and existing kinds. Emit a
   linked pair once as canonical-backed `call`, and only an unlinked canonical event as the
   additive `event` kind. The exact-channel call calendar consumes canonical server-expanded
   occurrences instead of parsing an independent RRULE and refetches on authorized
   `call.scheduled_due.v1` without a second reminder.
5. Move due processing to the canonical occurrence-keyed transaction and harden legacy
   conflict details to exact-channel membership. New Mail meeting UX uses v2 availability.
6. Sync system/architecture docs and record unrun integration evidence honestly.
7. Focused Docker unit/build verification when available.
8. Commit `feat(calendar): открыть API событий и RSVP`.

## Task 143-06 — Backfill migration plan and review

**Create/modify:** keep `.planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md` current;
create a production migration artifact only after founder review establishes the canonical
Atlas/apply mechanism. Do not invent a second migration system.

**Steps:**

1. Backfill every non-deleted call, including paused/completed rows, with project-derived
   organization, exact
   project/channel/creator/participant coherence, title/description/instant/duration/
   reminder/recurrence/audit mapping, deterministic UID/hash and accepted internal
   attendees. Emit no historical invitation/outbox records.
2. Block every recurring row until founder question 18 and the raw-worker/injected-UTC/
   proposed-canonical occurrence oracle agree. Never guess timezone or partially apply.
3. Cut over all-or-nothing only after 100% non-deleted-call linkage, bridge-only writer/worker
   inventory and reconciliation postconditions pass.
4. Define rollback/forward-only policy consistent with repository practice.
5. Require database integration evidence through the accepted QA/Docker layer; a unit mock
   is not migration proof.
6. Review source, generated artifacts, docs, secrets, and protected paths.
7. Commit `feat(calendar): подготовить миграцию участников звонков`.
