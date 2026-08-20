# Phase 143 — Calendar RSVP Core

## Combined exact17 conditional acceptance sync (2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> Task: TASK-236 · Status: IMPLEMENTATION_IN_PROGRESS_143_01_02A_02B_03A_03B_03C_03D_03F_04L_04A_04G_STATIC_ACCEPTED · Depends on: Phase 142

> **143-04G receipt:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `48a7273d100c07788def88e2f1b8d69c7173580f`; former 04G/04B next routes are `CONSUMED/HISTORICAL`. 143-04B is accepted at `75dafea889055126f768227a4ae62b28c91b7393`. No next Calendar code packet is selected; root must separately dispatch any later packet. Executables `NOT_RUN`, launcher `NOT_READY`; 03E remains blocked only for Ent generation/migration.

> **143-04B source receipt:** `75dafea889055126f768227a4ae62b28c91b7393` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact blobs in `IMPLEMENTATION-143-04B.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Old 04B `PRECOMMIT_FAIL`/`NOT_ACCEPTED`/next-route clauses `CONSUMED/HISTORICAL`. No next Calendar code packet selected; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`; 03E Ent generation/migration blocked.

> **143-04C authority candidate:** Clause-scoped precedence consumes only the preceding no-next-Calendar-packet clause for a separately reviewed 143-03B-T then 143-04C sequence. This is `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED`, not source acceptance; 03B-T owns only `removed_at`, two live partial-index predicates and schema metadata, while 04C owns later inactive/re-invite/replacement service behavior plus zero-call capability guarding and opaque all-and-only result echo rollback. All 04A/B/G receipts, 03E, launcher and `NOT_RUN` facts remain authoritative.

## Goal

Add the canonical calendar event/attendee model and deterministic iTIP/RSVP domain, then
connect it to the existing calendar and ScheduledCall surfaces without two write
authorities, duplicate meetings or dependence on the real mail engine.

## Task 143-01 — Freeze public domain types and failing policy units

> **Status:** `STATIC_ACCEPTED` on `e7d5f450`; executable verification `NOT_RUN`.
> Evidence: `IMPLEMENTATION-143-01.md`.

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

> **Status:** 143-02 codec and 143-03A/B/C/D/F are `STATIC_ACCEPTED`; 03D is committed at
> `10b4ceead119cb74a16b325b33f9121f60d7b1b5` after final independent review C/I/M `0/0/0`;
> 03F is committed at
> `846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final
> docs/whole-boundary review C/I/M `0/0/0` (the intermediate `0/0/1` docs-sync finding is
> superseded/closed). Generation, compilation, migration and executable verification remain `NOT_RUN`.
> Evidence: matching `IMPLEMENTATION-143-*.md` artifacts and Phase 147
> `STATIC-REVIEW-0.1.7.md`.

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
2. Keep accepted outbound serialization separate from bounded inbound parsing. The parser
   uses explicit limits and Mail 0.1.7 CVE-v1 identity, accepts occurrence REQUEST, and
   never hashes serialized/folded wire bytes.
3. Keep METHOD and content method consistent; never use `time.Now` internally.
4. Focused Docker unit verification when available.
5. Commit `feat(calendar): добавить детерминированный iTIP-кодек`.

## Task 143-04L — Shared deterministic event envelope

> **Status:** `STATIC_ACCEPTED` at `4525c559da22729aa8bc51ff5336c29dad4f27e9` after final
> independent Terra/high PASS C/I/M `0/0/0`; evidence: `IMPLEMENTATION-143-04L.md`.
> The initial C/I/M `0/2/2` and contract-amendment progression are historical only. The
> launcher remains `NOT_READY`; compile/tests/runtime are `NOT_RUN`.

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
   receipt ledger unique by attendee + event sequence + canonical hash version + sender-
   bound payload hash; it never stores raw MIME or an HTTP idempotency key. Every inbound
   canonical event/exception hash is paired with its version.
2. Add schema-level annotations/checks supported by current Ent patterns.
3. Run Ent generation only in accepted Docker execution; inspect generated diff before
   staging. Never generate on host.
4. Add/adjust schema contract units only in `backend/tests/calendar/**` if they remain
   genuine service-free units; otherwise hand integration checks to QA.
5. Commit `feat(calendar): добавить схемы событий и участников`.

## Task 143-04 — Repository and transactional RSVP service

> **04A/04G/04B status:** all are `STATIC_ACCEPTED` as recorded above. Their accepted boundaries neither generate Ent nor authorize migration or runtime. Former next-route clauses are `CONSUMED/HISTORICAL`; no next Calendar code packet is selected and root must separately dispatch any later packet.

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
create a production migration artifact only after the separate founder-approved migration
mechanism/gate establishes the canonical Atlas/apply authority. The all-A product decision
record does not establish that technical mechanism. Do not invent a second migration system.

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
