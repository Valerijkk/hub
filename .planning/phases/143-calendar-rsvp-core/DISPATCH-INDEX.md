# Phase 143 — exact Spark dispatch index

> **Status:** CONTRACT_READY · CODE_RUNTIME_BLOCKED
> **Required runtime:** `gpt-5.3-codex` / `xhigh`, no fallback
> **Execution:** strictly sequential where owned files overlap
> **Runtime evidence:** `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md`

This index decomposes Phase 143 into one-behavior implementation packets. Every dispatch
must include the full Hub prompt contract, exact clean HEAD, current allowed files, external-
package tests first, no commit, and `BLOCKED` on scope or runtime mismatch. Root reviews
each result before the next overlapping task.

Complete standalone prompts exist for 143-01, 02A, 02B, 04G and 04H. For every other row,
assemble the Common envelope plus exactly one matching section from
`DISPATCH-REMAINING.md`; combining sections is invalid.

## Dependency graph

```text
143-01 ── 143-02A ── 143-02B
   └───── 143-03A ── 143-03B ── 143-03F
                  ├── 143-03C
                  └── 143-03D
143-03A/B/C/D/F ── 143-03E ── 143-04A ── 143-04G

143-04A/G ──────────────── 143-04B create
143-04A/G ──────────────── 143-04C update/master cancel
143-04A/G ──────────────── 143-04D series RSVP ── 143-04E occurrence RSVP
143-02A + 143-04C/G ────── 143-04H occurrence cancel
143-02B + 143-04A/G ────── 143-04I inbound REPLY

143-03E + 143-04A/G ───── 143-04F-Base
each accepted 143-04B/C/D/E/H/I ── its matching 143-04F-* transaction

143-04B/C/D/E/H + matching F variants ── 143-05A ── 143-05B ── 143-05C ── 143-05D

143-06 production migration stays blocked until founder authority + launcher gates pass.
```

## Atomic packets

| ID | Single owned behavior | Exact primary files | Test oracle | Entry dependency |
|---|---|---|---|---|
| 143-01 | Timed/all-day validation and sequence/RSVP reset classification | `backend/internal/calendar/event_types.go`, `event_policy.go`, `backend/tests/calendar/event_policy_test.go` | table units for every policy row, including reminder-only version change | contract 0.1.4 |
| 143-02A | Deterministic RFC 5545/5546 serialization for REQUEST/CANCEL/REPLY | `itip_types.go`, `itip_serialize.go`, `backend/tests/calendar/itip_serialize_test.go` | exact in-memory CRLF/folding/escaping output; injected DTSTAMP | 143-01 |
| 143-02B | Bounded iTIP parsing and canonical duplicate key | `itip_parse.go`, `itip_limits.go`, `backend/tests/calendar/itip_parse_test.go` | in-memory malformed/depth/size/count/TZ/UID/SEQUENCE/spoof fixtures | 143-02A |
| 143-03A | `CalendarEvent` Ent schema only | `backend/ent/schema/calendarevent.go` | static schema review; generation NOT_RUN until launcher | 143-01 |
| 143-03B | `CalendarAttendee` schema with two partial unique indexes | `backend/ent/schema/calendarattendee.go` | exact NULL/non-NULL index predicates | 143-03A |
| 143-03C | `CalendarEventException` occurrence schema | `backend/ent/schema/calendareventexception.go` | event+recurrence unique, timed/all-day override fields | 143-03A |
| 143-03D | `CalendarMutationReceipt` plus narrow User/ScheduledCall edges | `backend/ent/schema/calendarmutationreceipt.go`, `user.go`, `scheduledcall.go` | hashed key uniqueness, bounded response, optional O2O call | 143-03A..C |
| 143-03F | Durable bounded inbound-REPLY receipt schema and attendee edge | `backend/ent/schema/calendarinboundreplyreceipt.go`, `calendarattendee.go` | unique attendee/sequence/hash; safe outcome; no raw MIME/key | 143-03B |
| 143-03E | Generate/inspect Ent output in accepted Docker job | generated Ent only, never manually edited or committed | clean generation and schema compile evidence | launcher + 143-03A..D/F |
| 143-04A | Public event repository/outbox/receipt ports and service constructor | `event_repository.go`, `event_service.go`, `event_errors.go` | compile-only consumer fakes in next behavior; no persistence | 143-03E |
| 143-04G | Admit calendar events to shared EVENTS stream without webhook exposure | `backend/internal/platform/nats/client.go`, focused genuine unit | add `calendar.>` once; exact webhook filter unchanged | 143-04A |
| 143-04B | Create event with auth-derived owner and response replay receipt | `event_service_create.go`, `backend/tests/calendar/event_service_create_test.go` | same key/input replays; changed input conflicts; one outbox | 143-04A + 143-04G |
| 143-04C | Material/silent/reminder update and master cancellation | `event_service_update.go`, `backend/tests/calendar/event_service_update_test.go` | sequence/reset/version table, If-Match conflict, one master cancel event | 143-04A + 143-04G |
| 143-04D | Series RSVP optimistic transition | `event_service_rsvp.go`, `backend/tests/calendar/event_service_rsvp_test.go` | one version increment/outbox; NFR-13 recovery metadata | 143-04A + 143-04G |
| 143-04E | HTTP occurrence RSVP with persisted mutation receipt | `event_service_occurrence.go`, `backend/tests/calendar/event_service_occurrence_test.go` | materialized override; replay/conflict; one receipt/outbox | 143-04D + 143-04G |
| 143-04H | Cancel one recurrence instance | `event_service_occurrence_cancel.go`, `backend/tests/calendar/event_service_occurrence_cancel_test.go` | cancelled exception, sequence/event version +1, occurrence iTIP CANCEL and one outbox | 143-02A + 143-04C + 143-04G |
| 143-04I | Sender-bound inbound iTIP REPLY apply/quarantine | `event_service_inbound_reply.go`, `backend/tests/calendar/event_service_inbound_reply_test.go` | A→B→A no rollback; same-status receipt; stale/future/spoof/cap quarantine | 143-02B + 143-04A + 143-04G |
| 143-04F-Base | Ent adapter shell and shared transaction primitive | `event_ent_repository.go`, `backend/tests/calendar/event_ent_repository_contract_test.go` | interface assertion; injected DB/outbox; no network/post-commit publish | 143-03E + 143-04A/G |
| 143-04F-B | Create-event Ent transaction | `event_ent_repository_create.go` | receipt/event/attendees/outbox/completion in one transaction | 143-04F-Base + 143-04B |
| 143-04F-C | Update/cancel-master Ent transaction | `event_ent_repository_update.go` | event-first lock/version/update/reset/outbox atomic | 143-04F-Base + 143-04C |
| 143-04F-D | Series-RSVP Ent transaction | `event_ent_repository_rsvp.go` | event-before-attendee lock; compare/update/receipt/outbox atomic | 143-04F-Base + 143-04D |
| 143-04F-E | Occurrence-RSVP Ent transaction | `event_ent_repository_occurrence.go` | event/exception/attendee lock order; inherited override/receipt/outbox atomic | 143-04F-Base + 143-04E |
| 143-04F-H | Occurrence-cancel Ent transaction | `event_ent_repository_occurrence_cancel.go` | event→exception lock; version+sequence/outbox atomic | 143-04F-Base + 143-04H |
| 143-04F-I | Inbound-REPLY Ent transaction | `event_ent_repository_inbound_reply.go` | event/exception/attendee lock order; receipt ledger + optional one-attendee/outbox mutation | 143-04F-Base + 143-04I |
| 143-05A | Canonical v2 HTTP DTO binding/validation/error metadata | `event_dto.go`, `backend/tests/calendar/event_handler_contract_test.go` | owner ignored from client; temporal pairs; 409 headers; occurrence cancel | 143-04B/C/D/E/H + accepted 143-04F-B/C/D/E/H |
| 143-05B | `/api/calendar/v2/**` handlers and composition | `event_handler.go`, `backend/internal/calendar/handler.go`, `backend/cmd/server/wiring_modules.go` | exact OpenAPI paths; legacy handler untouched | 143-05A |
| 143-05C | Add first-class event to legacy backend aggregate without envelope drift | `dto.go`, `repository.go`, `service.go`, focused genuine backend units | `{events,from,to}` unchanged; existing filters/export preserved | 143-05B |
| 143-05D | Add the new aggregate event kind to the existing frontend calendar client | `frontend/src/api/calendar.ts`, focused genuine frontend unit | existing kinds byte-shape compatible; additive event decoder/type only | 143-05C + frontend feature-test harness |
| 143-06 | Founder-reviewed production migration/backfill artifact | mechanism chosen by founder; `MIGRATION-PLAN.md` is authority meanwhile | empty + realistic snapshot, counts, no historical outbox | founder #12 + launcher |

## Cross-packet hard rules

- Existing `Service`, `Repository`, `GET /api/calendar/events`, and `.ics` export remain
  operational; new canonical types use distinct names until additive integration.
- `CalendarEvent` never stores an occurrence `recurrence_id`; exceptions do.
- Nullable attendee uniqueness requires two partial PostgreSQL indexes.
- Event create and RSVP use `CalendarMutationReceipt`; mail submission uses
  `MailSubmission`. Outbox is not an idempotency response store.
- Inbound iTIP REPLY uses the bounded `CalendarInboundReplyReceipt` ledger, never a
  `CalendarMutationReceipt`; packet 143-04F-I is reviewed independently.
- Reminder-relevant update/cancel/RSVP transactions lock event → optional exception →
  optional attendee so Phase 147 can serialize last-chance reminder revalidation without
  deadlock or a stale notification row.
- Phase 143 admits `calendar.>` to the shared EVENTS stream before any calendar producer;
  Phase 144 later adds `mail.>`. Neither wildcard enters `WebhookFilterSubjects`.
- NFR-13 body remains exactly `{error,request_id}`. Machine conflict codes use
  `X-Hub-Error-Code`; stale RSVP additionally returns current-version `ETag` and a same-
  origin canonical-event `Location`.
- No generated Ent file is manually edited or committed. No schema migration is invented.
- No host test/build/lint/generate/validator command and no direct Docker/Compose fallback.
- No change enters protected QA/test-rebase paths.

## Gate status

| Gate | Status |
|---|---|
| Contract root/independent review | ROOT PASS on `46cfa26c`; independent PENDING |
| Exact Spark implementation runtime | UNAVAILABLE |
| Accepted Docker launcher | NOT_READY |
| Ent generation | NOT_RUN |
| Production migration authority | UNRESOLVED |
| Runtime/tests/build | NOT_RUN |
