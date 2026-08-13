[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только отмену одного
recurrence instance canonical calendar event.

- User outcome: организатор отменяет одну встречу серии, не отменяя остальные; участники
  получают occurrence-scoped CANCEL ровно один раз.
- Completion: failing external-package units precede minimal service behavior; canonical
  exception, master version/sequence, iTIP and outbox are one accepted command.
- Non-goals: whole-series cancellation/update, RSVP, parser, HTTP handler/DTO, schema/
  generation, migration, mail transport, frontend, legacy calendar aggregate.

[OWNERSHIP]

- Entry: accepted 143-02A, 143-04A/C/G and matching repository interfaces on clean HEAD.
- Allowed mutations only:
  - `backend/internal/calendar/event_service_occurrence_cancel.go`
  - `backend/tests/calendar/event_service_occurrence_cancel_test.go`
  - narrow method signature additions to accepted `event_repository.go` only if the
    frozen command cannot be expressed; no implementation in that file
- Do not modify master update/RSVP/iTIP serializer logic, Ent adapter, handlers/wiring,
  NATS list or legacy `calendar` Service/Repository.
- `package calendar_test`; in-memory deterministic fakes; no DB/network/broker/clock.

[FROZEN COMMAND]

Input is authenticated actor, event id, decoded canonical recurrence id and expected
master event version. The service:

1. resolves organization/organizer authority before exposing event existence;
2. loads active recurring master and validates recurrence id belongs to its temporal kind;
3. compares expected master version;
4. writes/updates one `CalendarEventException(status=cancelled)` for event + recurrence;
5. increments master `sequence` and `version` exactly once;
6. serializes one `METHOD:CANCEL` with same UID, new sequence and `RECURRENCE-ID`;
7. persists one `calendar.invite.cancelled.v1` through outbox in the same transaction;
8. returns event id, canonical recurrence id, cancelled status, sequence and event version.

It never mutates RRULE, other exceptions, attendee historical RSVP or master status.
Already-cancelled same occurrence with current post-command state is an idempotent read;
stale pre-command `If-Match` conflicts before any second outbox/notification.

[REQUIRED ORACLES]

1. One occurrence cancellation leaves master confirmed and RRULE/other occurrences intact.
2. Exception uniqueness is event + recurrence; one row only.
3. Master sequence/version each increment exactly one; attendee status unchanged.
4. Serialized iTIP is `METHOD:CANCEL`, stable UID, new sequence, exact recurrence id and
   cancelled status; no unrelated occurrence.
5. Outbox subject/payload is `calendar.invite.cancelled.v1` with recurrence id; exactly one
   row in the same fake transaction.
6. Retrying stale expected version returns conflict with zero mutation/iTIP/outbox.
7. Already-cancelled canonical state with matching current observation returns the same
   cancellation view without duplicate side effects.
8. Non-recurring event, invalid/out-of-policy recurrence id, cancelled master or unknown
   occurrence fails safely before mutation.
9. Non-organizer/cross-organization actor receives non-disclosing error and zero lookup
   hydrate/mutation beyond owner-scoped query.
10. Transaction failure rolls back exception, master counters and outbox together.

[EXECUTION / PROVENANCE]

- TDD first; no dependency/generated code/global state.
- Host verification/direct Docker forbidden; accepted launcher absent => `NOT_RUN`.
- No commit/push/PR/network. max_depth 0; fan_out 0; retries 2; wall clock 45m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none; mismatch => `BLOCKED`
  before mutation. Orchestrator records exact clean source-contract SHA.

[ОТВЕТ]

На русском: transition; files/symbols; oracle matrix; atomicity/iTIP evidence;
verification; model provenance; risks.
