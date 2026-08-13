# Phase 143 — executable packet contracts not split into standalone files

> **Assembly rule:** a valid dispatch consists of the Common execution envelope below plus
> exactly one numbered packet. Never send two packets in one executor turn.
> **Runtime:** `gpt-5.3-codex` / `xhigh`, fallback none.
> **Status:** every code packet is blocked until that exact runtime is available; generation
> and runtime packets additionally require the accepted Docker launcher.

Standalone complete prompts already exist for 143-01, 143-02A, 143-02B, 143-04G and
143-04H. This file makes every remaining row in `DISPATCH-INDEX.md` independently
dispatchable without duplicating the common safety contract.

## Common execution envelope

- Role: Hub `codex-micro`, atomic Go/TypeScript implementer. TDD: write the stated
  external-package genuine unit first, then minimal implementation. If the packet has no
  genuine isolated-unit seam, do not invent a component test; record the future Python
  integration oracle from `QA-HANDOFF.md`.
- Orchestrator supplies exact clean branch/head. You are not alone: do not revert others;
  dirty overlap, missing dependency or changed frozen signature means `BLOCKED` before
  writes. Modify only Allowed mutations for the chosen packet; no commit/push/PR.
- Never read/change `testframework/**`, `backend/testframework/**`, test-rebase audit/
  phases 128–141/TASK-193–234, `pro-qa-ai/**`, `.env.local`, `.git/index.lock`.
- No new dependency, generated-file manual edit, secret, network, listener, real DB,
  broker, filesystem or wall clock unless the chosen packet explicitly is a launcher-owned
  generation/migration operation after its gates.
- Host tests/build/lint/validators/generation are forbidden. Direct Docker/Compose is not
  accepted fallback. Missing launcher => verification `NOT_RUN`, static review only.
- Go 1.25 style, Russian exported comments, stable typed errors, bounded values. No raw
  request body/secret/PII in errors/logs. Preserve existing calendar service/export and
  exact legacy `{events,from,to}` envelope unless packet 05C owns its additive behavior.
- Bounds: depth/fan-out 0, retries per root cause 2, wall clock 45 minutes (60 for service,
  75 for handler/migration). Response in Russian: summary; exact files/symbols; oracle
  matrix; verification; risks; requested/actual model evidence; exact source SHA.

## 143-03A — CalendarEvent schema

- Dependency: accepted 143-01. Allowed: only
  `backend/ent/schema/calendarevent.go`.
- Behavior: fields/types/defaults/bounds from Mail spec §3.11; owner + UID unique at master
  level; nullable call link unique; no recurrence id; soft delete/audit timestamps; edges
  only to already existing schema names or forward declarations allowed by Ent.
- Oracles: static schema table covers every field; UID uniqueness includes owner; temporal
  union is also enforced in service (do not pretend Ent alone proves cross-field end>start);
  no message body/secret; no migration/generated changes.

## 143-03B — CalendarAttendee schema

- Dependency: reviewed 03A. Allowed: only
  `backend/ent/schema/calendarattendee.go`.
- Behavior: fields/bounds/enums from §3.12; version positive; nullable Hub user; event edge;
  two PostgreSQL partial unique indexes using existing `entsql.IndexWhere`: series where
  recurrence is null, occurrence where non-null. Do not use a nullable composite unique as
  substitute. The inverse receipt edge is owned later by 03F.
- Oracles: exact predicates and columns; external guest accepted; sender identity not an
  authorization edge; no cascade capable of deleting audit history silently.

## 143-03C — CalendarEventException schema

- Dependency: reviewed 03A. Allowed: only
  `backend/ent/schema/calendareventexception.go`.
- Behavior: event + normalized recurrence unique; cancelled/materialized occurrence;
  nullable typed override fields, sequence/hash/timestamps; no master recurrence id field.
- Oracles: unique index; all-day/timed override representation; cancelled exception can
  exist without rewriting RRULE; service retains cross-field validation responsibility.

## 143-03D — receipt and narrow inverse edges

- Dependencies: reviewed 03A–C. Allowed only:
  `backend/ent/schema/calendarmutationreceipt.go`,
  `backend/ent/schema/user.go`, `backend/ent/schema/scheduledcall.go`.
- Behavior: §3.14 receipt with actor/org/operation, key/request hashes, in-progress/
  completed, bounded response, expiry and unique tuple; optional event edge. Add only the
  inverse calendar edges and optional one-to-one ScheduledCall link; preserve all existing
  fields/edges.
- Oracles: raw key absent; uniqueness includes operation/actor/org; response bounded; no
  receipt used for mail submission; exact pre/post edge inventory proves no removal.

## 143-03F — inbound-REPLY receipt ledger schema

- Dependency: reviewed 03B. Allowed only
  `backend/ent/schema/calendarinboundreplyreceipt.go` and the narrow inverse edge in
  `backend/ent/schema/calendarattendee.go`.
- Behavior: fields from §3.15; required attendee edge, non-negative event sequence,
  fixed SHA-256 payload hash, `applied|observed_no_change` outcome and created timestamp;
  unique attendee + sequence + hash. No sender address, MIME, HTTP key/response, engine id
  or organizer/event mutation edge. Retention/cap enforcement remains service behavior.
- Oracles: exact unique columns and bounds; two attendees may hold the same hash; one
  attendee may hold multiple hashes at the same sequence; exact pre/post attendee edge
  inventory proves no removal; no generated/migration edits.

## 143-03E — Ent generation inspection

- Dependencies: accepted 03A–D/F, founder migration decision and accepted launcher. Allowed:
  no manually edited tracked source; task-owned ignored generated Ent output/report only.
- Behavior: run repository-approved Ent generation inside bounded launcher; capture exact
  SHA/tool version, inspect generated predicates/edges/types and schema compile; clean exact
  resources unconditionally. Do not stage ignored output or create migration SQL.
- Oracles: clean generation, no unexpected schema deletion, all five new types/edges
  including inbound receipt, and partial-index metadata retained; second generation diff
  empty. If launcher absent,
  `BLOCKED/NOT_RUN`, no direct command.

## 143-04A — event service ports and constructor

- Dependency: accepted 03E. Allowed only:
  `backend/internal/calendar/event_repository.go`, `event_service.go`, `event_errors.go`,
  `backend/tests/calendar/event_service_contract_test.go`.
- Behavior: immutable commands/results; repository transaction callback owns receipt,
  event/attendee/exception and outbox operations; actor-to-org resolution; injected clock/
  id/serializer ports; no Ent/JMAP/Fiber type in public domain contract.
- Oracles: external fake compiles; transaction API cannot publish post-commit; receipt
  reserve/replay primitives explicit; not-found/authz/conflict/invalid errors safe; no
  constructor side effect.

## 143-04B — create with persisted receipt

- Dependencies: 04A/G. Allowed only:
  `backend/internal/calendar/event_service_create.go`,
  `backend/tests/calendar/event_service_create_test.go`.
- Behavior: auth-derived owner/org; canonical validation; receipt key hash + request hash;
  event/attendees/outbox/completed response atomically; organizer accepted, new attendees
  needs_action; deterministic UID/sequence 0.
- Oracles: same key/input byte-equivalent replay and one event/outbox; changed input 409;
  rollback leaves no receipt/event; client owner ignored; external guest safe; no raw key.

## 143-04C — update and whole-master cancel

- Dependencies: 04A/G. Allowed only:
  `backend/internal/calendar/event_service_update.go`,
  `backend/tests/calendar/event_service_update_test.go`.
- Behavior: If-Match master version; complete merged temporal validation; exact §4.2
  sequence/RSVP/notification table; whole cancel preserves historical RSVP and emits one
  master-scoped cancel event/iTIP request descriptor.
- Oracles: every change class; no-op; silent/notified description; attendee add/remove;
  stale/conflict zero mutation; rollback atomic; master cancel cannot masquerade as one
  occurrence.

## 143-04D — series RSVP

- Dependencies: 04A/G. Allowed only:
  `backend/internal/calendar/event_service_rsvp.go`,
  `backend/tests/calendar/event_service_rsvp_test.go`.
- Behavior: actor resolves exactly one series attendee; persisted idempotency receipt;
  attendee-version compare; one status/version/responded_at change + outbox; external
  organizer yields pending REPLY transport descriptor.
- Oracles: all transitions; stale returns canonical recovery metadata for handler; same-key
  replay; changed-key input conflict; cancelled event blocked; another attendee unchanged;
  one outbox/version increment.

## 143-04E — HTTP occurrence RSVP

- Dependencies: 04D/G. Allowed only:
  `backend/internal/calendar/event_service_occurrence.go`,
  `backend/tests/calendar/event_service_occurrence_test.go`.
- Behavior: authenticated HTTP occurrence response materializes/inherits one attendee
  override while the series remains unchanged. Reserve/replay the occurrence-scoped
  `CalendarMutationReceipt`, compare attendee version and complete receipt + attendee +
  outbox atomically.
- Oracles: series vs occurrence; same key/input replay; changed input conflict; stale
  attendee zero mutation; exactly one receipt/outbox/version increment; rollback leaves no
  partial override. Do not parse or apply inbound iTIP in this packet.

## 143-04I — inbound REPLY apply/quarantine

- Dependencies: 02B/04A/G. Allowed only:
  `backend/internal/calendar/event_service_inbound_reply.go`,
  `backend/tests/calendar/event_service_inbound_reply_test.go`.
- Behavior: consume only the bounded canonical parser result. Resolve the addressed
  owner/event by UID and optional normalized recurrence id, then bind normalized
  authenticated/envelope sender to exactly one attendee. Require REPLY sequence equal to
  the current event sequence; lower is a typed stale no-op, higher/spoof/malformed identity
  is a typed quarantine with zero mutation. Derive the sender-bound receipt hash as SHA-256
  over a length-delimited normalized sender plus the parser's fixed canonical payload hash;
  then look up the durable ledger before state comparison. An existing hash is a no-op even
  if a later reply changed the status. For a
  new hash below the per-attendee/sequence cap of eight, same-status writes only an
  `observed_no_change` receipt; a distinct PARTSTAT may update only that attendee,
  increment its version, write an `applied` receipt and one
  `calendar.rsvp.changed.v1` outbox event. The ninth distinct hash quarantines with zero
  mutation.
- Oracles: series and occurrence keys; normalized sender case/Unicode; unknown attendee;
  organizer-field injection; other-attendee mutation; lower/higher sequence; exact replay;
  `declined(A) -> accepted(B) -> redelivered A` for series and occurrence; distinct-hash
  same-status receipt followed by later state change/redelivery; eight accepted hashes and
  ninth quarantine; rollback. No HTTP idempotency key, `CalendarMutationReceipt`, outbound
  REPLY transport, event sequence change or organizer/event-field mutation is allowed.

## 143-04F-Base — Ent adapter shell

- Dependencies: 03E/04A/G. Allowed only
  `backend/internal/calendar/event_ent_repository.go` and
  `backend/tests/calendar/event_ent_repository_contract_test.go`.
- Behavior/oracles: define the Ent-backed repository type/constructor, injected `*ent.Client`
  and outbox publisher, interface assertion and one shared transaction callback primitive.
  Constructor performs no query/network call; callback cannot publish after commit; safe
  not-found/unique/conflict translation only. The external-package test is compile/
  constructor behavior only and does not open a DB. No command method is implemented here.

## 143-04F-B — create-event Ent transaction

- Dependencies: 04F-Base/04B. Allowed only
  `backend/internal/calendar/event_ent_repository_create.go`.
- Behavior/oracles: implement only create receipt reserve/replay, event/attendees/outbox and
  receipt completion under one `*ent.Tx`; changed hash conflicts, exact completion replays,
  rollback leaves zero durable rows. Real constraint/rollback proof is routed to the Python
  integration handoff; do not add a Go DB component test.

## 143-04F-C — update/cancel-master Ent transaction

- Dependencies: 04F-Base/04C. Allowed only
  `backend/internal/calendar/event_ent_repository_update.go`.
- Behavior/oracles: implement only master lock/version read, event/attendee reset or audit
  tombstone mutations and outbox under one transaction. Stale/no-op paths publish nothing;
  no post-commit publish or Go DB component test.

## 143-04F-D — series-RSVP Ent transaction

- Dependencies: 04F-Base/04D. Allowed only
  `backend/internal/calendar/event_ent_repository_rsvp.go`.
- Behavior/oracles: implement only series attendee version compare/update, HTTP receipt and
  outbox atomically. Another attendee cannot be selected; replay and stale paths publish
  nothing; no Go DB component test.

## 143-04F-E — occurrence-RSVP Ent transaction

- Dependencies: 04F-Base/04E. Allowed only
  `backend/internal/calendar/event_ent_repository_occurrence.go`.
- Behavior/oracles: implement only inherited occurrence attendee materialization/update,
  HTTP receipt and outbox atomically; the series row remains unchanged. Duplicate/stale/
  rollback paths create no partial override; no Go DB component test.

## 143-04F-H — occurrence-cancel Ent transaction

- Dependencies: 04F-Base/04H. Allowed only
  `backend/internal/calendar/event_ent_repository_occurrence_cancel.go`.
- Behavior/oracles: implement only cancelled exception upsert, master version/sequence
  compare/increment and occurrence cancel outbox atomically. RRULE and other instances are
  untouched; stale/retry paths publish nothing; no Go DB component test.

## 143-04F-I — inbound-REPLY Ent transaction

- Dependencies: 04F-Base/04I. Allowed only
  `backend/internal/calendar/event_ent_repository_inbound_reply.go`.
- Behavior/oracles: implement only sender-bound attendee lookup plus atomic current event
  sequence, attendee-row lock, bounded receipt count and unique ledger lookup/insert. An
  accepted distinct
  PARTSTAT inserts `applied`, updates that one attendee/version and writes one outbox; a new
  same-status hash inserts only `observed_no_change`; existing hashes, stale/future/spoof
  and cap quarantine write nothing. Concurrent duplicate insert converges through the
  unique constraint without a second attendee/outbox mutation. Prove static paths for
  `A -> B -> redelivered A` and occurrence scope; no HTTP receipt or Go DB component test.

## 143-05A — canonical v2 DTO/validation

- Dependencies: accepted 04B–E/H and matching 04F variants. Allowed only:
  `backend/internal/calendar/event_dto.go`,
  `backend/tests/calendar/event_dto_test.go`.
- Behavior: OpenAPI 0.1.2 request/response shapes; strict temporal union/bounds; quoted
  version parser; base64url recurrence decoder; NFR-13 metadata carrier; no owner/org input.
- Oracles: timed/all-day pairs/end order/IANA annotations; unknown/oversized fields; RSVP
  version/status; occurrence cancel response; error body has only error/request_id.

## 143-05B — v2 handlers and wiring

- Dependency: 05A. Allowed only:
  `backend/internal/calendar/event_handler.go`, the smallest additive registration in
  `backend/internal/calendar/handler.go`, `backend/cmd/server/wiring_modules.go`, and
  `backend/tests/calendar/event_handler_test.go`.
- Behavior: exact `/api/calendar/v2/**` methods including occurrence DELETE; current auth
  middleware; Idempotency-Key/If-Match; same-origin RSVP Location; no change to legacy
  route handler/export.
- Oracles: route inventory, status/body/headers, malformed recurrence/version, cross-org
  non-disclosure, service fake call counts, occurrence cancel mapping, exact legacy route
  registration snapshot.

## 143-05C — additive legacy backend aggregate

- Dependency: 05B. Allowed only:
  `backend/internal/calendar/dto.go`, `repository.go`, `service.go`, and a focused genuine
  unit under `backend/tests/calendar/legacy_aggregate_test.go` when isolatable.
- Behavior: add canonical event as one new aggregate kind while preserving top-level
  `{events,from,to}`, existing calls/sprints/deadlines/leave and `.ics` behavior/filters.
- Oracles: old fixture JSON remains byte-shape compatible; new kind additive; time range/
  auth filters; no N+1 hydrate; export unchanged. If repository proof requires DB, route it
  to Python integration instead of a component Go test.

## 143-05D — additive frontend aggregate type

- Dependencies: 05C + accepted frontend feature-unit harness. Allowed only:
  `frontend/src/api/calendar.ts`, `frontend/tests/calendar/legacy-event-kind.test.ts`.
- Behavior: type/decode one new `event` kind from the existing endpoint; preserve URL,
  response wrapper and all existing kinds. No canonical v2 mutations or UI.
- Oracles: old fixtures parse identically; new fixture parse; unknown kind safe; no API
  client/header/dependency change. Docker-only unit; harness missing => BLOCKED.

## 143-06 — production migration/backfill

- Dependencies: founder chooses mechanism, 03 schemas accepted, launcher ready. Allowed:
  only the founder-selected versioned migration path plus
  `.planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md` evidence update; no alternate
  migration system.
- Behavior: exact preflight, schema indexes/constraints and ScheduledCall backfill policy;
  missing creator/email/timezone reported, not guessed; no historical invitation outbox.
- Oracles: empty + realistic bounded snapshot in accepted launcher, before/after counts,
  deterministic UID, accepted creator/participants, rollback/forward procedure and current
  SHA artifact. Until authority exists: `BLOCKED`, plan only.
