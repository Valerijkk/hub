# Phase 143 — executable packet contracts not split into standalone files

> **143-04I inbound iTIP REPLY acceptance sync candidate:** `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-self-activating dirty exact9 docs-only sync records immutable candidate `fa69484cde0fff77b5947a204a42f8566808ed73` (sole parent `6e54eb8c1764946826115036b90646b412db4792`, tree `d0a0971258dc34b09c8aa51c2ea71c2c231cb64b`, subject `docs(calendar): propose inbound reply authority`) and final manifest in `REVIEW-143-04I-INBOUND-REPLY-AUTHORITY.md`. Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Only one atomic manifest-equal exact9 acceptance commit and immutable acceptance-commit triple PASS can activate it; until then fa694 remains current `ROOT_SOURCE_AUTHORITY_CANDIDATE` and source is blocked. After activation only the ordered exact3 test → `event_repository.go` → new service is source-authorized, still `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` pending its own lifecycle. Only after that atomic manifest-equal exact9 acceptance commit and immutable Contract/Security/Lifecycle PASS do lower 04I candidate/precommit/root-commit/exact-SHA-next routes become `CONSUMED/HISTORICAL`; 04F-I, Ent/generated source, 03E, generation/migration, HTTP/API/OpenAPI/UI, NATS/iTIP transport, testframework and runtime remain blocked/not-ready/not-run; no next packet follows.

> **143-04E occurrence RSVP source receipt-sync:** `SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT2_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact2 documentation receipt is non-authoritative and cannot self-activate. The accepted linear source chain is: repository seam `63cbd8f48a3cd7beefbfc4ca734537da6c07ed89` (parent `7ba1bd2e732522c258063513c0134c49bf41561e`, tree `70531f6f011fa33896cfab23ff5302334cc87983`, subject `feat(calendar): add occurrence RSVP repository seam`; repository `92f718bc2eeda8edfa85367d9d32294d9aa1d9b5`, test `3a4301ec8807d68d127276bf5e1ba90f6fed7404`); identity-mask correction `37ec9e2136f3bc5662ca20280a4589a60a2c4a95` (parent `63cbd8f48a3cd7beefbfc4ca734537da6c07ed89`, tree `3133687b5b808762cc9ed7f10d2163970fd3ef60`, subject `fix(calendar): mask occurrence RSVP identity`; repository `1f93796b854ed196963837de9efc71aec5e7b27a`, test `4dee0fa8c814430ec34b91852c5c26f1bfa2f7db`); bounded-resolution correction `528d1a6de8182009674ab4101aad7764bb79573e` (parent `37ec9e2136f3bc5662ca20280a4589a60a2c4a95`, tree `4cf1fb52ae9572ff8c66d53dcb589e4007de2b5f`, subject `fix(calendar): bound occurrence RSVP resolution`; repository `13c125a58ab76ade948bf194fdaab58878dcc24c`, test `333f21367f79ac15412e96300d5e9f578cf75597`); then service `f4d3d341bdb463b065cc4b962939e76e25ce4d0a` (parent `528d1a6de8182009674ab4101aad7764bb79573e`, tree `b359cbf5f2d46f448dbe313c3628d30d5482734f`, subject `feat(calendar): add occurrence RSVP service`; source `e3017aa30ba3a510cbf5fec6ca8a1cdd5afcba1a`, test `1e52529b58128391ad1d30a2dd16d2f3ef6c2f9e`). Contract, Security and Lifecycle precommit plus immutable reviews PASS C/I/M `0/0/0`; author Terra/xhigh, reviewers Terra/high, Terra/xhigh, Terra/high; provenance `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Only one atomic exact2 commit of these documents and its immutable triple review can establish `STATIC_ACCEPTED / EXACT_SHA_PASS`; then the prior 04E future/unimplemented/source-absent route is `CONSUMED/HISTORICAL` only for 04E. No next Calendar packet is selected. 04F-E Ent adapter/generation/migration, HTTP handlers/API/OpenAPI/UI, NATS/iTIP, launcher and runtime remain blocked `NOT_READY`/`NOT_RUN` pending the external framework handoff.

> **143-04H current source receipt:** `STATIC_ACCEPTED / EXACT_SHA_PASS` at `73940810046bb3186f2746705a83565ab29368c5`; its exact16 receipt `f9f5d436c08737eeb5a3e965137f329f3ae7f627` is ancestral to HEAD. The former conditional route below is `CONSUMED/HISTORICAL` only for 04H. PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback, telemetry `UNKNOWN`; no next packet is selected and all generation/migration/adapter/API/UI/launcher/runtime blockers remain.

> **143-04H source receipt-sync:** `SOURCE_RECEIPT_SYNC_CANDIDATE /
> ATOMIC_EXACT16_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. Dirty exact16 grants no acceptance. Immutable
> source `73940810046bb3186f2746705a83565ab29368c5` (parent `9a638f75745aa0fb516a1177a33cbd1c657432fe`,
> tree `9f994a8805b43a84c00ed50b8e5544739425edaa`, subject `feat(calendar): add occurrence cancellation
> service`) records authored-first test `80e19d3b918ae9ddbe3b3ea8803c6062c2f8a713`, repository
> `e0da66bf95d4b8193eb5d4034053a353bd154b66`, service `2b500e5cdda34d7d80d0ccc0ba9c94427167716f`.
> Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high precommit plus immutable PASS
> C/I/M `0/0/0`; author Terra/xhigh; dispatcher-enforced/no fallback; telemetry `UNKNOWN`. One
> atomic exact16 manifest commit plus immutable triple PASS alone activates `STATIC_ACCEPTED /
> EXACT_SHA_PASS`; then lower 04H authority/precommit/source-unaccepted/current-next clauses are
> `CONSUMED/HISTORICAL`. No next packet or runtime action follows.

> **143-04H authority acceptance-sync:** `AUTHORITY_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. Dirty exact15 grants
> no dispatch. Candidate `f2966bdd` has Contract Terra/high, Security Terra/xhigh and Lifecycle
> Terra/high immutable exact-SHA PASS C/I/M `0/0/0`; atomic sync commit plus immutable
> acceptance-commit review remain required. On activation only the frozen
> exact3 becomes dispatchable; source and downstream work remain unaccepted/blocked.


> **143-04H root authority candidate:** `ROOT_SOURCE_AUTHORITY_CANDIDATE /
> PRECOMMIT_REVIEW_REQUIRED / ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED /
> RUNTIME_NOT_RUN`. This dirty exact15 carrier grants no dispatch. Its only future source is the
> frozen test-first exact3 in `REVIEW-143-04H-OCCURRENCE-CANCEL-AUTHORITY.md`, and it activates only
> after triple precommit PASS, atomic exact15 candidate commit, triple immutable exact-SHA PASS and
> a separate accepted exact15 sync. Legacy `METHOD:CANCEL`, `EventSerializer`, 143-02A and
> source-first clauses are `CONSUMED/HISTORICAL`; source/runtime/downstream blocks remain current.


> **143-04J source-receipt sync:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT14_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. These proposed receipt facts grant no
> authority before one atomic exact14 docs commit. After it, immutable exact2 `e74efa827b52...`
> is `STATIC_ACCEPTED / EXACT_SHA_PASS`, and lower 04J authority/source-unaccepted routes become
> `CONSUMED/HISTORICAL`. 04K still needs its own docs authority and root dispatch.

> **143-04J acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT13_COMMIT_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> These proposed receipt facts have no authority before one atomic exact13 docs commit. After it,
> immutable candidate `5841621720cc03682b3d1f1e2af9cba2953797cf` is
> `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS`, the three immutable reviews are PASS C/I/M
> `0/0/0`, and lower 04J candidate routes are `CONSUMED/HISTORICAL`. Only the frozen ordered
> test-first exact2 may then be dispatched; 04K/04F and all persistence/runtime routes stay closed.

> **143-04C acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT12_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact12 working tree grants no
> authority. The receipt below activates only after one atomic commit of exactly these 12 documents;
> until then the prior 04C lifecycle remains current. Activation selects no next Calendar packet.

## 143-04C immutable source acceptance (2026-08-20)

Calendar 143-04C is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `7fd061b5099b4608e0c33a6d99edbe1a66587d34` (parent `bb84a47e12516c19682ff08783a5844306fc5766`, tree `0fb65fc72705bb145a1f2950c6efbe16607e6a04`, subject `feat(calendar): add update and cancel service`) for its exact3 test-first blobs: `backend/tests/calendar/event_service_update_test.go` `4421204141719fa1af586b3d6272aa130f185af2`, `backend/internal/calendar/event_repository.go` `737207280cc23adb873ae8d36952a3e62a0a3a4d`, then `backend/internal/calendar/event_service_update.go` `094b68dd8369bdc9345d9348016b3bdc377aa7a9`. Receipt: `IMPLEMENTATION-143-04C.md`. Contract/Spec/Oracle, Security and Lifecycle precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; author Terra/xhigh, dispatcher-enforced/no fallback, telemetry `UNKNOWN` unless exposed.

Every lower 143-04C candidate, `ROOT_SOURCE_AUTHORITY_CANDIDATE`, precommit/source-unaccepted/uncommitted/current-next, or blocked-pending-source-acceptance clause is `CONSUMED/HISTORICAL` only for 04C. No next Calendar code packet is selected or authorized; root must separately select any later packet. 03E/Ent generation/migration, adapter/API/UI, raw iTIP, launcher `NOT_READY`, and all executable/runtime evidence `NOT_RUN` remain authoritative.

## 143-03B-T immutable source acceptance and former conditional 143-04C gate (CONSUMED/HISTORICAL as to 04C; 2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; 144-04G still requires separate docs authority because its repository port is absent. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## 144-04F immutable source acceptance (2026-08-20)

Mail 144-04F is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `4715f939f24e5b5c2d3f80a38d204babe68bdd34` (parent `7ac8714e99db2935ce290ad67836e1309cc05c78`, tree `6461894a90c4340bf96118ed1a242516c02b48c7`, subject `feat(mail): add sync page policy`). Its exact4 ordered test-first blobs are recorded in `IMPLEMENTATION-144-04F.md`: events test `50a41cefb539d9b840a43da9e1ff34cce6c8653f`, sync-policy test `b6410709e4fb8166c28a5c502cfba535eef08d06`, events source `a3017b0a6064d8fc1017e37e2d4f675ee8732945`, sync-policy source `acd642fcdb6290af79b7fa8e5aa36a5036576f47`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file asserting that Mail exact4 is dirty, `UNCOMMITTED`, `UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is `CONSUMED/HISTORICAL`; the immutable source receipt above is current. No automatic Mail next source follows: 144-04G remains `NOT_AUTHORIZED / BLOCKED` pending separate docs authority because its required repository port is absent.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted. 143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; 143-04A/B/G receipts remain accepted. Launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. 143-03E, 144-02E, Ent generation and migration stay blocked. No raw iTIP or service/API/UI/product/phase/release authority follows.

## Combined exact17 conditional acceptance sync (CONSUMED/HISTORICAL; 2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> **Assembly rule:** a valid dispatch consists of the Common execution envelope below plus
> exactly one numbered packet. Never send two packets in one executor turn.
> **Runtime:** `gpt-5.6-terra` / `xhigh`, fallback none.
> **Status:** the exact route is founder-authorized. An accepted native `dispatcher-enforced`
> record permits actual telemetry `UNKNOWN` when unavailable; exposed mismatch, rejection,
> omitted model/effort or fallback fails closed. Generation/runtime packets additionally require
> the accepted Docker launcher.

Standalone complete prompts already exist for 143-01, 143-02A, 143-02B, 143-04G and
143-04H. The 143-04H prompt is frozen but remains non-dispatchable until its exact15 candidate and
separate acceptance lifecycle completes. Subject to each row's current authority gate, this file
makes the remaining rows independently dispatchable without duplicating the common safety contract.

> **143-04B accepted source receipt:** `75dafea889055126f768227a4ae62b28c91b7393` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact blobs in `IMPLEMENTATION-143-04B.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Old 04B `PRECOMMIT_FAIL`/`NOT_ACCEPTED`/next-route clauses `CONSUMED/HISTORICAL`. No next Calendar code packet selected; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`; 03E Ent generation/migration blocked.

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
  level; nullable call link unique; nullable paired canonical payload-hash version/hash;
  no recurrence id; soft delete/audit timestamps; edges
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
  nullable typed override fields, sequence, paired canonical payload-hash version/hash and
  timestamps; no master recurrence id field.
- Oracles: unique index; all-day/timed override representation; cancelled exception can
  exist without rewriting RRULE; service retains cross-field validation responsibility.

## 143-03D — receipt and narrow inverse edges

- Dependencies: reviewed 03A–C. Allowed only:
  `backend/ent/schema/calendarmutationreceipt.go`,
  `backend/ent/schema/user.go`, `backend/ent/schema/scheduledcall.go`.
- Behavior: §3.14 receipt with actor/org/operation, key/request hashes, in-progress/
  completed, bounded response, expiry and unique tuple; optional event edge. Add only the
  inverse calendar edges and optional one-to-one ScheduledCall link; preserve all existing
  fields/edges. The link supports creation-time attachment and remains immutable by service
  contract after event creation.
- Oracles: raw key absent; uniqueness includes operation/actor/org; response bounded; no
  receipt used for mail submission; exact pre/post edge inventory proves no removal.

## 143-03F — inbound-REPLY receipt ledger schema

- Dependency: reviewed 03B. Allowed only
  `backend/ent/schema/calendarinboundreplyreceipt.go` and the narrow inverse edge in
  `backend/ent/schema/calendarattendee.go`.
- Behavior: fields from §3.15; required attendee edge, non-negative event sequence,
  required canonical payload-hash version plus fixed SHA-256 sender-bound hash,
  `applied|observed_no_change` outcome and created timestamp; unique attendee + sequence +
  version + hash. No sender address, MIME, HTTP key/response, engine id
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

- Dependencies: accepted 03E and 04L. Allowed only:
  `backend/internal/calendar/event_repository.go`, `event_service.go`, `event_errors.go`,
  `backend/tests/calendar/event_service_contract_test.go`.
- Behavior: immutable commands/results; repository transaction callback owns receipt,
  event/attendee/exception and outbox operations; actor-to-org resolution; injected clock/
  id/serializer ports; every outbox descriptor carries the complete accepted 04L envelope,
  never only inner payload; no Ent/JMAP/Fiber type in public domain contract.
- Oracles: external fake compiles; transaction API cannot publish post-commit; receipt
  reserve/replay primitives explicit; not-found/authz/conflict/invalid errors safe; no
  constructor side effect.

## 143-04B — create with persisted receipt

- Dependencies: 04A/G. Allowed only:
  `backend/internal/calendar/event_service_create.go`,
  `backend/tests/calendar/event_service_create_test.go`.
- Status: `PRECOMMIT_FAIL` / `NOT_ACCEPTED`. The existing 04B source candidate is not accepted: it requires this 04A ABI amendment, a 04B corrective source packet and fresh independent review. This document grants no runtime, Ent, migration or schema authority.
- ABI: the full proposed 04A `CalendarEventCreateAttendeeCandidate`, authorization request,
  resolved organizer/attendee and authorization field lists plus the exact backend method and
  guarded transaction wrapper are frozen only in `MAIL-SPEC.md` section 3.14. They are not an
  accepted code ABI; a separate 04A amendment packet must add and independently review them.
- Frozen behavior supersedes the candidate's stale create wording: within the same guarded repository transaction, first derive scope, then reserve or replay/conflict the receipt, then call the 04A `ResolveCalendarEventCreateAuthorization`, then atomically write event/attendees/exdates/outbox/completed response. `CalendarEventCreateAuthorizationRequest` carries candidate organizer/attendees plus canonical Temporal, RRULE, EXDATE, reminder and optional `ScheduledCallID` values; `CalendarEventCreateAuthorization` is resolved output only. The resolver alone resolves organizer/internal users; nonnil invalid/inactive/foreign/email-mismatched candidates reject rather than downgrade to guests. Client owner/org is ignored. Organizer is accepted, distinct new attendees are needs_action, UID is deterministic and sequence is 0.
- ScheduledCall: a nonnil ID is authorized in that same transaction only when its organizer and
  accessible call owner both equal `scope.ActorID()`, it is active, unlinked and in the same
  organization/project/channel context, with exact normalized mapping and no external guest.
  No implicit in-organization impersonation/delegation exists. Nonexistent, not-owned or
  otherwise inaccessible IDs return content-free `ErrEventNotFound` / 404; only accessible
  actor-owned inactive/already-linked/context-or-mapping-mismatched/external-guest cases use
  `ErrEventConflict`. Request nonnil means resolved/persisted/returned ID is exactly equal;
  request nil means resolved ID is nil; substitution/drop/insertion rejects before writes.
- EXDATE: timed `YYYY-MM-DDTHH:MM:SSZ` UTC whole seconds; all-day Gregorian `YYYY-MM-DD` in
  event IANA zone; nil equals empty; max 1000; sort canonical and reject duplicates. Use that
  vector in request hash, persistence and response. RRULE follows the exact 3.11 grammar and
  order/valid-invalid vectors; names are <=160 runes and <=640 bytes; timezone is <=64 UTF-8
  bytes and valid IANA.
- Non-tautological oracles: byte-equivalent replay produces one event/outbox; changed canonical
  input conflicts; rollback leaves no completed receipt/event; each invalid nonnil identity
  class rejects; resolver output, not candidate ids, is persisted; actor A with organizer B
  returns `ErrEventConflict`, while actor A with ScheduledCall owner B returns
  `ErrEventNotFound` / 404; each has zero receipt, event, attendee, exception and outbox
  writes. An accessible actor-owned inactive/already-linked/context-or-mapping-mismatched or
  external-guest call returns `ErrEventConflict` before writes. A nonnil request must resolve,
  persist and return exactly its own ID; nil must resolve nil; substitution/drop/insertion
  rejects before writes. Timed/all-day EXDATE grammar, nil-empty equivalence, sorting,
  duplicate and 1001st-value rejection are observable; the canonical EXDATE vector changes
  the request hash and is returned/persisted unchanged; the 3.11 RRULE, name and timezone
  boundary vectors hold; no raw key or Hub JSON `payload_hash` persists.
- Behavior: auth-derived owner/org; canonical validation; receipt key hash + request hash;
  event/attendees/outbox/completed response atomically; organizer accepted, new attendees
  needs_action; deterministic UID/sequence 0.
- Oracles: same key/input byte-equivalent replay and one event/outbox; changed input 409;
  rollback leaves no receipt/event; client owner ignored; external guest safe; no raw key.

### 04B RRULE occurrence-cap correction candidate

- Status: `04B` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `75dafea889055126f768227a4ae62b28c91b7393`; the former correction-candidate/precommit/`NOT_ACCEPTED` record is `CONSUMED/HISTORICAL`. Its static-only limits remain: no generation, migration, launcher, runtime or broader authority.
- `GLOBAL_RRULE_OCCURRENCE_CAP=1000`. On create, every RRULE must contain exactly one terminal `COUNT=1..1000` or compatible `UNTIL`: neither and both reject. `COUNT` includes DTSTART; UNTIL is inclusive; `UNTIL < DTSTART` rejects. Apply the cap before EXDATE or exception, so a 1001st nominal occurrence rejects even if it would subsequently be excluded/cancelled.
- Expand as a stream, never by full series materialization. Reject on the 1001st nominal occurrence. Candidate work is bounded to 1001 nominal attempts and fails closed if the bound cannot be proved, including conservative rejection of sparse rules. DTSTART appears exactly once even when it mismatches a selector. Weekly semantics are `WKST=MO` with selector dates on/after DTSTART; missing monthly BYMONTHDAY dates are skipped. Timed recurrence is in event-timezone wall clock: a New York fold chooses the earlier occurrence. RFC 5545 section 3.3.5 pre-gap UTC-offset interpretation applies to standalone DATE-TIME values, while section 3.3.10 makes a recurrence-generated nonexistent local time ignored and not counted. All-day recurrence is date-based.
- Required exact oracles: `FREQ=DAILY;INTERVAL=1;COUNT=1000` accepts and `FREQ=DAILY;INTERVAL=1;COUNT=1001` rejects. With `DTSTART=2026-01-01`, daily `UNTIL=20280926` accepts and `UNTIL=20280927` rejects. Weekly interval-two Monday uses `FREQ=WEEKLY;INTERVAL=2;BYDAY=MO` with `WKST=MO`; monthly 31 uses `FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=31` and skips months without day 31. The New York (`America/New_York`) spring-gap oracle is `DTSTART=2026-03-01 02:30` with `FREQ=WEEKLY;INTERVAL=1;BYDAY=SU;COUNT=2`: recurrence-generated `2026-03-08 02:30` is nonexistent, ignored and not counted; `2026-03-15 02:30` is occurrence two. New York fold fixtures choose the earlier occurrence. Neither/both terminal-selector fixtures and `UNTIL < DTSTART` rejection are mandatory.

## 143-04C — update and whole-master cancel

- Status: `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7fd061b5099b4608e0c33a6d99edbe1a66587d34`;
  exact test-first three paths and review facts are recorded in `IMPLEMENTATION-143-04C.md`. The
  former 04A/G-only/two-file, candidate, and current-next clauses are `CONSUMED/HISTORICAL` only
  for 04C. No next Calendar code packet is selected or authorized by this receipt.
- Behavior: If-Match master version; complete merged temporal validation; exact §4.2
  sequence/RSVP/notification table; whole cancel preserves historical RSVP and emits one safe
  master-scoped `calendar.invite.cancelled.v1` EnvelopeV1 outbox descriptor. The former
  durable/raw/typed iTIP request-descriptor requirement is `CONSUMED/HISTORICAL` for this 04C
  exact-three source packet because no accepted transport port exists. It records notification
  intent only: no iTIP is queued or delivered, and no raw iTIP/body is stored. A future separate
  Mail/iTIP hydration-delivery packet may consume this safe event and reconstruct transport
  content. A stale master returns typed
  `CALENDAR_EVENT_VERSION_CONFLICT` recovery metadata; terminal state returns typed
  `CALENDAR_EVENT_NOT_MUTABLE`. Reminder-only change advances event version without iTIP
  sequence/RSVP/invitation mutation.
- Oracles: every change class; no-op; silent/notified description; reminder-only change;
  attendee add/remove;
  exact stale code/current event version/canonical location; exact terminal code;
  stale/conflict zero mutation; rollback atomic; master cancel cannot masquerade as one
  occurrence.

## 143-04D — series RSVP

- `STATIC_ACCEPTED` / `EXACT_SHA_PASS` only after the conditional exact2 docs receipt and
  immutable review. The accepted linear source chain is repository seam
  `c1932c261f29ca4ae313d39cf299e0ade313837c` (parent
  `7531565879e0d7aa72a246407fd0e1d609a82c49`, tree
  `7d36d49f0edfbd03434526cc1974d9c83ef685a6`, subject
  `feat(calendar): add series RSVP repository seam`; repository blob
  `079f68d9a1a9432b65ee88010cbc92befce23d0d`, test blob
  `83154ef604ca725414aef3a5340e2d5c427dddf8`), amendment
  `9332516550f8a04fc9530a9b89dd55b93d30bf54` (parent `c1932c261f29ca4ae313d39cf299e0ade313837c`,
  tree `105f4fad3313d8db512d2bf502fdb05434c3304e`, subject
  `fix(calendar): allow initial RSVP sequence`; final repository blob
  `83d8b382415472b02b032dd024d34a134a6ad265`, test blob
  `5b001e7ea9ef5968e735a56d0b69610677f08a77`), then service
  `e9781ff6885d482b23b53455e6c8f3c07c3b8177` (parent
  `9332516550f8a04fc9530a9b89dd55b93d30bf54`, tree
  `e53b8f00e1c22f2ff16d7183fcf851a4e50ad1bf`, subject
  `feat(calendar): add series RSVP service`; source blob
  `9d27ef48aa68b7472cf4b22e427793428ab255ec`, test blob
  `0d5f156d9d7fa90c4a30d5dae3a5032120426d79`). Contract/Security/Lifecycle
  precommit and immutable reviews PASS C/I/M `0/0/0`; provenance
  `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Dirty docs cannot self-activate.
  No next Calendar packet is selected or authorized: 143-04E is future/unselected;
  143-04F-D is blocked behind 143-04F-Base/143-03E; Ent persistence/generation/migration,
  HTTP/API/OpenAPI/UI, launcher and runtime remain blocked `NOT_READY`/`NOT_RUN`. Former
  04D absent, unimplemented, and future-dispatch claims are `CONSUMED/HISTORICAL` only for
  implementation of 143-04D itself.

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
  over the Mail 0.1.7 domain prefix, parser canonical version, length-delimited normalized
  sender and the parser's fixed canonical payload hash; persist the parser version and look
  up version + hash in the durable ledger before state comparison. Different canonical
  versions are never equal. An existing hash is a no-op even
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

## 143-04J — pure ScheduledCall projection policy

- Dependencies: 01/02A/04A/B/G. Allowed only
  `backend/internal/calendar/event_call_policy.go` and
  `backend/tests/calendar/event_call_policy_test.go`.
- Behavior: pure bounded validation/mapping between canonical timed Hub meetings and a
  ScheduledCall projection. Require project/channel equality, resolved organization,
  active exact-channel creator, active same-project/same-organization creator and internal
  participants, normalized unique emails, positive checked duration, reminder
  `null|0..40320`, target string bounds and canonical RRULE subset. The event owns
  recurrence; compute next effective occurrence from DTSTART/RRULE, EXDATE and cancelled
  exceptions. Reject external attendees on a linked-call event under founder question 23's
  safe default. Project only series-level internal non-organizer attendees in
  `needs_action|accepted|tentative`; occurrence RSVP never mutates the participant edge.
  Define pause/resume/delete and one-time-complete outcomes without I/O.
- Oracles: every missing/cross-org/boundary case; creator dedupe; DST/UTC anchors; supported
  daily/weekly/monthly COUNT/UNTIL fixtures; EXDATE/exception removal; no-future result;
  pause/resume cannot cancel/resurrect event; deterministic projection/result. Do not encode
  founder question 18 as an assumed migration answer.

## 143-04K — pure call-due service

- **Source receipt-sync gate:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
  ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact15 documentation candidate
  grants no authority. Only after one atomic exact15 receipt-sync commit and immutable acceptance is
  `572f911227b9d50e21aa73b87c530ddcaf0f8934` (parent `26b377eb`, tree `ac96e2f5`, subject
  `feat(calendar): add call due policy`) `STATIC_ACCEPTED / EXACT_SHA_PASS` for test blob
  `437db3af0263cedb10dcf217cd53fbbcac03cf23` then source blob
  `376f46c1cc79956fab3479efc38750fd88161b16`. Precommit plus immutable Contract/Spec/Oracle,
  Security/Adversarial and Lifecycle/Dependency/Boundary reviews PASS C/I/M `0/0/0`; Terra/xhigh,
  dispatcher-enforced/no fallback, telemetry `UNKNOWN`. The dirty candidate selects no next packet.

- Prerequisites: accepted 04J `d7926a27`/`e74efa827`, 04A `EventOutboxDescriptor` and 04L
  `EnvelopeV1`. It explicitly does **not** derive Calendar-subject admission from 04G.
- Status: the authority-candidate/precommit/test-first-dispatch clauses are `CONSUMED/HISTORICAL`
  only after this source receipt activates. The immutable exact2 is recorded in
  `IMPLEMENTATION-143-04K.md`; it authorizes no next packet.
- Frozen behavior: pure explicit `Call`, UTC whole-second `Now`, raw 04J projection plus explicit
  envelope input; opaque `not_due|reschedule|advance|deactivate|noop` result and optional complete
  descriptor. An active call deep-clones projection then invokes exactly two fixed 04J reconcile
  inputs (`Call.ScheduledAt` inclusive and `Now` strict-after); inactive/deleted requires nil
  projection/zero envelope and immediately noops without 04J. It derives
  the domain-separated key from raw calendar-event/call UUID bytes plus nominal 20-byte key, builds
  the plural existing nine-field due envelope, and locally validates subject/payload/channel.
  Exact/overdue yields at most one descriptor; no catch-up burst. It has no clock lookup, I/O,
  CAS, transaction, outbox append or NATS publish and never mutates event version/sequence, RSVP
  or reminder state.
- Oracles: full state/output table; exact/overdue/early and `Now==OccurredAt`; fixed two-call
  projection boundary/no-caller-skip and 04J typed-error mapping; fixed key vector/different IDs/
  moved nominal/DST; payload null/subject invariant; deep-copy/repeated-call/observable purity.
  Repeat full input returns a byte-identical descriptor; payload/envelope/subject bindings reject.
  04F-K later candidate-calls then reserves/reloads EventID/OccurredAt/CorrelationID only after a
  due outcome (discarding non-due candidates) and separately owns event→call
  CAS/locking/rollback/same-tx append. AQ-02 is backfill-only; Ent/generation/migration,
  adapters/API/UI/raw iTIP, legacy `calls.scheduled.due` cutover and runtime remain blocked.

## 143-04F-Base — Ent adapter shell

### 143-04J controlling candidate clarification — not active

The exact13 documentation candidate freezes the future ABI only in
`REVIEW-143-04J-CALL-PROJECTION-AUTHORITY.md`: string actions and `ProjectEventCall`; separate
event/target/creator/attendee/exception/boundary input; opaque copied snapshot; series-only
participants; exact whole-second UTC key and inclusive/strict-after boundary. It requires Hub
source, event-target-creator identity equality, active project/channel creator and active
same-project/same-organization attendees. 1..1440 exact-minute duration, title 255,
description 2000, attendee 200 and RRULE/EXDATE/exception 1000 limits apply. The candidate is
not dispatch authority until atomic exact13 commit, three immutable exact-SHA PASS reviews and
separate docs acceptance. It leaves 04K/04F, Ent, adapter, migration, API/UI, raw iTIP and
runtime outside scope; AQ-02 remains backfill-only. At that later scoped acceptance, this exact
freeze consumes only the preceding generic 04J ABI/behavior/oracle wording; no other packet or
blocker clause is consumed.

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
- Behavior/oracles: implement only event-first master lock/version read, then optional linked
  call lock before event/attendee reset or audit tombstone mutations and outbox under one
  transaction. Stale/no-op paths publish nothing; no post-commit publish or Go DB component
  test. Actual projection mutation is owned by 04F-J.

## 143-04F-D — series-RSVP Ent transaction

- Dependencies: 03D/04F-Base/04D. Allowed only
  `backend/internal/calendar/event_ent_repository_rsvp.go`.
- Behavior/oracles: lock event → optional linked call → series attendee, then implement
  attendee version compare/update, series-level call-participant projection, HTTP receipt
  and outbox atomically. Another attendee cannot be selected; occurrence state is untouched;
  replay and stale paths publish nothing; no Go DB component test.

## 143-04F-E — occurrence-RSVP Ent transaction

- Dependencies: 04F-Base/04E. Allowed only
  `backend/internal/calendar/event_ent_repository_occurrence.go`.
- Behavior/oracles: lock event → exception → attendee, then implement only inherited
  occurrence attendee materialization/update, HTTP receipt and outbox atomically; the
  series row remains unchanged. Duplicate/stale/rollback paths create no partial override;
  no Go DB component test.

## 143-04F-H — occurrence-cancel Ent transaction

- Dependencies: 04F-Base/04H. Allowed only
  `backend/internal/calendar/event_ent_repository_occurrence_cancel.go`.
- Behavior/oracles: lock event → optional linked call → exception, then implement only
  cancelled exception upsert, master version/sequence compare/increment and occurrence
  cancel outbox atomically.
  RRULE and other instances are
  untouched; stale/retry paths publish nothing; no Go DB component test.

## 143-04F-I — inbound-REPLY Ent transaction

- Dependencies: `143-03F + 143-04F-Base + accepted 143-04I`. `143-03D` is forbidden and
  insufficient for the inbound receipt ledger. Allowed only
  `backend/internal/calendar/event_ent_repository_inbound_reply.go`.
- Behavior/oracles: implement only sender-bound attendee lookup plus atomic current event
  sequence, event → optional linked call → optional exception → attendee lock order,
  bounded receipt count and
  unique ledger lookup/insert. An
  accepted distinct
  PARTSTAT inserts `applied`, updates that one attendee/version and writes one outbox; a new
  same-status hash inserts only `observed_no_change`; existing hashes, stale/future/spoof
  and cap quarantine write nothing. Concurrent duplicate insert converges through the
  unique constraint without a second attendee/outbox mutation. Prove static paths for
  `A -> B -> redelivered A` and occurrence scope. A series reply updates the linked call
  participant projection; an occurrence reply never does. No HTTP receipt or Go DB component
  test.

## 143-04F-J — event/ScheduledCall bridge transaction

- Dependencies: 03D/04F-Base/04B/C/H/J. Allowed only
  `backend/internal/calendar/event_ent_repository_call_bridge.go`.
- Behavior/oracles: implement paired legacy create, calendar-v2 existing-call attach,
  canonical/legacy edit mirror, pause/resume, legacy delete and event/occurrence cancellation
  projection effects. Resolve project organization and all identities server-side. Existing
  call attachment requires authenticated creator ownership, active/non-deleted/unlinked
  state and same organization/project/channel. Normalize and compare every mapped business
  field plus organizer/participant set; mismatch is a stable no-mutation conflict rather
  than an overwrite. Inaccessible/nonexistent ids remain indistinguishable; link is
  immutable. For existing rows lock
  event → call → optional exception/attendees and mutate canonical event, attendees,
  projection and invitation outbox in one transaction. Legacy edits read the locked current
  event version; v2 still compares If-Match. No partial pair, direct NATS, automatic mismatch
  winner or Go DB component test.

## 143-04F-K — canonical call-due transaction

- Dependencies: 04F-J/04K. Allowed only
  `backend/internal/calendar/event_ent_repository_call_due.go`.
- Behavior/oracles: claim one due linked pair with event → call lock, reload canonical
  exceptions/participants, re-run 04K, insert one outbox row and advance/deactivate the call
  in the same transaction. Conditional current scheduled-at/version predicates make
  concurrent workers converge; rollback leaves both pending. The outbox payload contains
  only existing safe call metadata plus deterministic occurrence key. Delivery is
  at-least-once; downstream dedupe is explicit. No event/reminder mutation or direct publish.

## 143-05A — canonical v2 DTO/validation

- Dependencies: accepted 04B–E/H and matching 04F variants. Allowed only:
  `backend/internal/calendar/event_dto.go`,
  `backend/tests/calendar/event_dto_test.go`.
- Behavior: OpenAPI 0.1.7 request/response shapes; strict temporal union/bounds; quoted
  version parser; base64url recurrence decoder; NFR-13 metadata carrier; no owner/org input.
- Oracles: timed/all-day pairs/end order/IANA annotations; unknown/oversized fields;
  creation-only server-authorized `scheduled_call_id` absent from PATCH; RSVP version/status;
  occurrence cancel response; exact event/occurrence/legacy writer conflict-code unions;
  stale event-version ETag plus same-origin Location; error body has only error/request_id.

## 143-05B — v2 handlers and wiring

- Dependency: 05A. Allowed only:
  `backend/internal/calendar/event_handler.go`, the smallest additive registration in
  `backend/internal/calendar/handler.go`, `backend/cmd/server/wiring_modules.go`, and
  `backend/tests/calendar/event_handler_test.go`.
- Behavior: exact `/api/calendar/v2/**` methods including occurrence DELETE; current auth
  middleware; Idempotency-Key/If-Match; same-origin RSVP/event recovery Location; exact
  `CALENDAR_EVENT_VERSION_CONFLICT`, `CALENDAR_EVENT_NOT_MUTABLE`,
  `CALENDAR_OCCURRENCE_NOT_MUTABLE` and `SCHEDULED_CALL_NOT_MUTABLE` mappings; no change to
  legacy route handler/export.
- Oracles: route inventory, status/body/headers, malformed recurrence/version, cross-org
  non-disclosure, service fake call counts, every writer's exact conflict code/recovery
  headers/zero-mutation mapping, exact legacy route registration snapshot.

## 143-05C — additive legacy backend aggregate

- Dependency: 05B. Allowed only:
  `backend/internal/calendar/dto.go`, `repository.go`, `service.go`, and a focused genuine
  unit under `backend/tests/calendar/legacy_aggregate_test.go` when isolatable.
- Behavior: add canonical event as one new aggregate kind while preserving top-level
  `{events,from,to}`, existing calls/sprints/deadlines/leave and `.ics` behavior/filters. A
  linked active pair is emitted once as existing `call` using canonical temporal truth;
  only unlinked canonical events are emitted as `event`.
- Oracles: old fixture JSON remains byte-shape compatible; new unlinked-event kind additive;
  linked recurrence/exdate/cancelled-exception has no duplicate/ghost; time range/auth
  filters; no N+1 hydrate; export unchanged. If repository proof requires DB, route it to
  Python integration instead of a component Go test.

## 143-05D — additive frontend aggregate type

- Dependencies: 05C + accepted frontend feature-unit harness. Allowed only:
  `frontend/src/api/calendar.ts`, `frontend/tests/mail/calendar-legacy-event-kind.test.ts`.
- Behavior: type/decode one new `event` kind from the existing endpoint; preserve URL,
  response wrapper and all existing kinds. No canonical v2 mutations or UI.
- Oracles: old fixtures parse identically; new fixture parse; unknown kind safe; no API
  client/header/dependency change. Docker-only unit; harness missing => BLOCKED.

## 143-05E — legacy ScheduledCall write bridge

- Dependencies: 05B/04F-J. Allowed only
  `backend/internal/calls/calendar_bridge.go`, the narrow required changes in
  `backend/internal/calls/scheduled.go`, `handler_scheduled.go`, `dto.go`,
  `backend/cmd/server/wiring_integrations.go`, both shipped create clients
  `frontend/src/modules/calls/create-call-dialog.tsx` and `scheduled-calls.tsx`, the narrow
  API change in `frontend/src/api/calls.ts`, and genuine idempotency/bridge units under
  `backend/tests/calendar/**` and `frontend/tests/mail/**`.
- Behavior: inject a consumer-owned bridge interface. Create/update/delete delegate to the
  canonical transaction while preserving existing successful JSON fields/routes. POST now
  requires `Idempotency-Key`; the same key/request replays the original response and a
  changed request conflicts. Both shipped create UIs generate one stable key per submission
  attempt and reuse it through transport/auth retry. Existing
  creator-or-channel-owner/admin authorization remains explicit and is passed as actor;
  calendar authorization still requires creator exact-channel membership plus active
  same-project/same-organization creator and internal participants; no external guest path
  is invented. Founder question 22's default permits the mandatory header only with both
  shipped clients updated in the same cutover. Legacy PUT has no `If-Match`
  and is documented serialized last-write-wins. Map stable validation/conflict/not-found/
  forbidden errors without leaking another call/event; terminal or invalid pause/resume/
  update/delete maps exactly to `SCHEDULED_CALL_NOT_MUTABLE` with the NFR-13 body and zero
  mutation. Apply founder question 19 default: is_active toggles execution; DELETE cancels
  both.
- Oracles: one bridge call, response mapping, invalid/missing/foreign, rollback error, admin
  actor audit, invitation-relevant update, exact not-mutable code for terminal pause/resume/
  update/delete, and no direct old repository mutation after cutover. No handler listener or
  component test.

## 143-05F — canonical legacy readers and due worker

- Dependencies: 05E/04F-K. Allowed only the narrow necessary portions of
  `backend/internal/calls/repository.go`, `scheduled.go`,
  `backend/cmd/server/wiring_integrations.go`, plus genuine isolated policy/service units
  under `backend/tests/calendar/**`.
- Behavior: list reads the canonical-backed next execution projection; due cron delegates
  to 04F-K and never parses ScheduledCall RRULE or publishes directly. Conflict reads use
  canonical effective occurrences for linked calls and a bounded active-unlinked path only
  before cutover. After activation any unlinked non-deleted row fails the invariant rather
  than being resumed, processed or silently treated as canonical.
  A detail containing title/channel is returned only when the caller belongs to that exact
  channel; sharing another channel with a participant is insufficient. Keep the legacy
  success envelope and generic safe errors.
- Oracles: linked/unlinked, recurring exception, paused/cancelled/exhausted, concurrent due,
  exact-channel member/non-member, participant enumeration and zero title/channel leak.
  Repository query/transaction proof belongs to Python integration.

## 143-05G — canonical-backed exact-channel call calendar reader

- Dependencies: 05C/F and accepted frontend unit harness. Allowed only
  `frontend/src/api/calls.ts`, `frontend/src/modules/calls/call-calendar.tsx`, the narrow
  unused raw-expansion helper removal in `calendar-types.ts`, and
  `frontend/tests/mail/call-calendar-canonical-range.test.tsx`.
- Behavior: keep the existing exact-channel list authorization. Add paired `from/to` to
  every channel request as an RFC3339 UTC half-open range; convert the current view's
  inclusive visual end to the exact next-millisecond exclusive instant. Render additive
  canonical `occurrences`; stop expanding
  ScheduledCall RRULE independently in the browser. Preserve call route, colors and existing
  call interaction. A cancelled/excluded occurrence is absent because the server is
  authoritative. Create-call API accepts the stable idempotency key required by 05E. On
  mount/reconnect, subscribe to each channel room returned by the authorized channel list
  and unsubscribe on removal/unmount. On authorized `call.scheduled_due.v1`, invalidate only
  the scheduled-call query family; do not invent a second reminder/toast or trust payload
  fields as authorization. Missed best-effort WS is recovered by the normal HTTP query.
- Oracles: one-time/recurring/exception/range edges including local day/month and DST
  exclusive-end conversion, no duplicate linked event, no raw RRULE
  expansion path, connect/reconnect/add/remove/unmount room lifecycle, exact-event-only
  invalidation and unchanged navigation. Docker-only unit; missing harness => BLOCKED.

## 143-06 — production migration/backfill

- Dependencies: founder chooses mechanism, 03 schemas accepted, launcher ready. Allowed:
  only the founder-selected versioned migration path plus
  `.planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md` evidence update; no alternate
  migration system.
- Behavior: exact preflight, schema indexes/constraints, complete ScheduledCall field/org/
  identity mapping, recurring semantic oracle and all-or-nothing bridge cutover from
  `MIGRATION-PLAN.md`; no historical invitation outbox.
- Oracles: empty + realistic bounded snapshot in accepted launcher, before/after counts and
  hashes, deterministic UID/payload, accepted creator/participants, project/channel/org
  coherence, founder recurrence policy, 100% non-deleted-call linkage, direct-writer inventory,
  rollback/forward procedure and current SHA artifact. Until authority exists: `BLOCKED`,
  plan only.

## 143-04C source-authority candidate (CONSUMED/HISTORICAL)

The former candidate is `CONSUMED/HISTORICAL`: the sequential 143-03B-T → 143-04C bridge is
now statically accepted at `7fd061b5099b4608e0c33a6d99edbe1a66587d34`. Its test-first exact
three paths are `backend/tests/calendar/event_service_update_test.go`,
`backend/internal/calendar/event_repository.go`, and
`backend/internal/calendar/event_service_update.go`; exact immutable evidence is in
`IMPLEMENTATION-143-04C.md`. This receipt neither selects nor authorizes a subsequent Calendar
code packet.

143-04C freezes internal `EventMutationOperation` (the four receipt values plus `event_update`),
not a receipt enum change. Scope retains a nonempty receipt operation except `event_update`;
`ResolveCalendarEventUpdateScope` is separate and update scope cannot reserve/complete a receipt.
Before `ResolveActorOrganization`, it asserts active backend capability; absent capability is
`ErrEventInvalid` with zero backend calls. Guarded resolve/persist wrappers repeat that assertion;
no `event_ent` adapter exists.
Presence-preserving patch and locked organizer resolver are authoritative and non-disclosing:
missing/foreign/nonorganizer is `ErrEventNotFound` before terminal/source/ExpectedVersion.
The request has EventID plus only presence-gated copied attendee candidates; authorization returns
authoritative current, active/resolved attendees and exception state. Existing masters only:
no-op zero writes; otherwise one command does event/version plus add/reset/tombstone and immutable
outbox. Material changes order sequence, version, applicable non-organizer RSVP reset, then
canonical changed fields. A materialized exception makes RRULE,
EXDATE, DTSTART/start-date, timezone and all-day changes reject before writes with typed
content-free `recurrence_exception_conflict`, `errors.Is(..., ErrEventConflict)` and public
`CALENDAR_EVENT_RECURRENCE_CONFLICT`; end/duration-only remains possible. 04C persists only safe
`calendar.invite.updated.v1` / `calendar.invite.cancelled.v1` EnvelopeV1 outbox descriptors;
they express notification intent, never queued or delivered iTIP transport. No raw iTIP/body,
delivery claim, recurrence rebase or exception retention behavior. A future separate Mail/iTIP
hydration-delivery packet may consume a safe event and reconstruct transport content.

The authoritative review freezes the exact eight presence wrappers, full update input/command/
service/cancel signatures, full authorization attendee snapshot, `add|reset|remove` mutation and
persistence constructor/getter record declarations, 13 ordered changed fields, and two `events.NewEnvelopeV1`
update/cancel builders. Required tests include constructor/deep-copy/absent-null, transition
validation, 200/201 attendee cap, exact payload, changed-field order/duplicate, every
event/version/sequence/all-and-only attendee-result mismatch rollback and no-op zero persistence/outbox.
