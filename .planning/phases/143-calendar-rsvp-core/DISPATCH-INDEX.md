# Phase 143 — exact Terra dispatch index

> **143-04H source receipt-sync:** `SOURCE_RECEIPT_SYNC_CANDIDATE /
> ATOMIC_EXACT16_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact16 documentation receipt is
> non-authoritative. Source `73940810046bb3186f2746705a83565ab29368c5` (parent `9a638f75745aa0fb516a1177a33cbd1c657432fe`,
> tree `9f994a8805b43a84c00ed50b8e5544739425edaa`, subject `feat(calendar): add occurrence cancellation
> service`) has ordered test-first blobs `80e19d3b918ae9ddbe3b3ea8803c6062c2f8a713`,
> `e0da66bf95d4b8193eb5d4034053a353bd154b66`, `2b500e5cdda34d7d80d0ccc0ba9c94427167716f`.
> Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high precommit plus immutable
> exact-SHA PASS C/I/M `0/0/0`; author Terra/xhigh; dispatcher-enforced/no fallback; telemetry
> `UNKNOWN`. Only atomic exact16 manifest commit plus immutable triple PASS yields
> `STATIC_ACCEPTED / EXACT_SHA_PASS`; lower 04H authority/precommit/source-unaccepted/current-next
> routes are then `CONSUMED/HISTORICAL`. No next packet; all downstream/static runtime blocks persist.

> **143-04H authority acceptance-sync:** `AUTHORITY_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. Dirty exact15 is not
> a dispatch grant. Candidate `f2966bdd` has Contract Terra/high, Security Terra/xhigh and Lifecycle
> Terra/high immutable exact-SHA PASS C/I/M `0/0/0`; atomic sync commit plus immutable
> acceptance-commit review remain required. Only activation consumes lower
> routes and permits root to dispatch the already frozen ordered exact3; no other path is selected.


> **143-04H root authority candidate:** `ROOT_SOURCE_AUTHORITY_CANDIDATE /
> PRECOMMIT_REVIEW_REQUIRED / ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED /
> RUNTIME_NOT_RUN`. Dirty exact15 is non-authoritative. The only frozen future source order is
> NEW `backend/tests/calendar/event_service_occurrence_cancel_test.go`, MODIFY
> `backend/internal/calendar/event_repository.go`, NEW
> `backend/internal/calendar/event_service_occurrence_cancel.go`. Root dispatch remains forbidden
> until triple precommit PASS -> atomic exact15 candidate commit -> triple immutable exact-SHA PASS
> -> separate atomic exact15 acceptance sync and immutable review. Legacy raw iTIP/serializer/
> 143-02A/source-first routes are `CONSUMED/HISTORICAL`; 04F-H and every downstream/runtime blocker
> remain. Full controlling contract: `REVIEW-143-04H-OCCURRENCE-CANCEL-AUTHORITY.md`.


> **143-04J source-receipt sync:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT14_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. These proposed receipt facts grant no
> authority before one atomic exact14 docs commit. After it, source `e74efa827b52...` is
> `STATIC_ACCEPTED / EXACT_SHA_PASS`, and lower 04J authority/source-unaccepted routes become
> `CONSUMED/HISTORICAL`. No next packet is selected; the generic 04K row remains future-only.

> **143-04J acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT13_COMMIT_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> These proposed receipt facts have no authority before one atomic exact13 docs commit. After it,
> `5841621720cc03682b3d1f1e2af9cba2953797cf` is `SOURCE_AUTHORITY_ACCEPTED /
> EXACT_SHA_PASS`, the three immutable reviews are PASS C/I/M `0/0/0`, and lower 04J candidate
> routes are `CONSUMED/HISTORICAL`. Root may then dispatch only test
> `backend/tests/calendar/event_call_policy_test.go`, followed by source
> `backend/internal/calendar/event_call_policy.go`; neither source file is accepted yet.

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

> **Status:** 143_01_02A_02B_03A_03B_03C_03D_03F_04L_04A_04C_04G_STATIC_ACCEPTED · RUNTIME_NOT_RUN
> **Next packet:** none selected. 143-04B is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; root must separately dispatch any later Calendar packet. No Ent/schema migration/runtime authority follows.
> **Required runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Execution:** strictly sequential where owned files overlap
> **Runtime evidence:** `.planning/phases/143-calendar-rsvp-core/RUNTIME-BLOCKER.md`

This index decomposes Phase 143 into one-behavior implementation packets. Every dispatch
must include the full Hub prompt contract, exact clean HEAD, current allowed files, external-
package tests first, no commit, and `BLOCKED` on scope or runtime mismatch. Root reviews
each result before the next overlapping task.

Complete standalone prompts exist for 143-01, 02A, 02B, 04G, 04H and 04L. The 04H prompt is a
non-authoritative exact15 candidate and is not dispatchable before its separate acceptance sync. For every other row,
assemble the Common envelope plus exactly one matching section from
`DISPATCH-REMAINING.md`; combining sections is invalid.

> **143-04B accepted source receipt:** `75dafea889055126f768227a4ae62b28c91b7393` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact blobs in `IMPLEMENTATION-143-04B.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Old 04B `PRECOMMIT_FAIL`/`NOT_ACCEPTED`/current-next clauses `CONSUMED/HISTORICAL`. No next Calendar code packet selected; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`; 03E Ent generation/migration blocked.

## Dependency graph

```text
143-01 ── 143-02A ── 143-02B
   └───── 143-03A ── 143-03B ── 143-03F
                  ├── 143-03C
                  └── 143-03D
contract 0.1.7 ── 143-04L shared versioned envelope
143-04L ── 143-04A ── 143-04G
143-03A/B/C/D (03F satisfied) ── 143-03E ── Ent generation/migration only

143-04A/G ──────────────── 143-04B create
143-04A/G ──────────────── 143-04C update/master cancel
143-04A/G ──────────────── 143-04D series RSVP ── 143-04E occurrence RSVP
143-04A/B/C/G/L ─────────── 143-04H safe occurrence cancel (candidate; no 02A serializer)
143-02B + 143-04A/G ────── 143-04I inbound REPLY
143-01/02A + 143-04A/B/G ─── 143-04J call-projection policy ── 143-04K due policy

143-03E + 143-04A/G ───── 143-04F-Base
each accepted 143-04B/C/D/E/H/I ── its matching 143-04F-* transaction
143-03D + 143-04B/C/H/J ── 143-04F-J call bridge transaction
143-04F-J + 143-04K ────── 143-04F-K call due transaction

143-04B/C/D/E/H/J/K + matching F variants ── 143-05A ── 143-05B
143-05B + 143-04F-J ───── 143-05E legacy write bridge
143-05B/E ──────────────── 143-05C aggregate dedupe ── 143-05D
143-05E + 143-04F-K ───── 143-05F legacy call readers/worker ── 143-05G call calendar

143-06 production migration stays blocked until founder authority + launcher gates pass.
```

## Atomic packets

| ID | Single owned behavior | Exact primary files | Test oracle | Entry dependency |
|---|---|---|---|---|
| 143-01 | Timed/all-day validation and sequence/RSVP reset classification — `STATIC_ACCEPTED` on `e7d5f450`; evidence: `IMPLEMENTATION-143-01.md`; runtime `NOT_RUN` | `backend/internal/calendar/event_types.go`, `event_policy.go`, `backend/tests/calendar/event_policy_test.go` | table units for every policy row, including reminder-only version change | contract 0.1.7 (policy unchanged) |
| 143-02A | Deterministic RFC 5545/5546 serialization for REQUEST/CANCEL/REPLY — `STATIC_ACCEPTED` through `d73a9380`; evidence: `IMPLEMENTATION-143-02A.md`; runtime `NOT_RUN` | `itip_types.go`, `itip_serialize.go`, `backend/tests/calendar/itip_serialize_test.go` | exact in-memory CRLF/folding/escaping output; injected DTSTAMP | 143-01 |
| 143-02B | Bounded iTIP parsing and versioned CVE-v1 canonical duplicate key — `STATIC_ACCEPTED` through `efa7814d`; evidence `IMPLEMENTATION-143-02B.md` | `itip_parse.go`, `itip_limits.go`, `backend/tests/calendar/itip_parse_test.go` | in-memory malformed/depth/size/count/TZ/UID/SEQUENCE/spoof/equivalence/version fixtures | 143-02A + reviewed Mail 0.1.7 |
| 143-03A | `CalendarEvent` Ent schema — `STATIC_ACCEPTED` on `31bfff8f58ba83d12b59600135503f98a6c92fbd`; exact CHECK and external Ent-metadata evidence recorded; runtime `NOT_RUN` | `backend/ent/schema/calendarevent.go` | static schema review; generation NOT_RUN until launcher | 143-01 |
| 143-03B-T | attendee tombstone schema correction — `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED`; prior no-next Calendar packet clause is CONSUMED/HISTORICAL only for this sequential bridge | `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go` | only `removed_at` representation, two live partial-index predicates, and schema-metadata oracle | accepted 143-03B |
| 143-03C | `CalendarEventException` occurrence schema — `STATIC_ACCEPTED` on `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`; exact CHECK and external Ent-metadata evidence recorded; runtime `NOT_RUN` | `backend/ent/schema/calendareventexception.go` | event+recurrence unique, timed/all-day override fields | 143-03A |
| 143-03D | `CalendarMutationReceipt` plus narrow User/ScheduledCall edges — `STATIC_ACCEPTED` on `10b4ceead119cb74a16b325b33f9121f60d7b1b5` after final independent C/I/M `0/0/0`; evidence: `IMPLEMENTATION-143-03D.md`; runtime `NOT_RUN` | `backend/ent/schema/calendarmutationreceipt.go`, `user.go`, `scheduledcall.go`, `backend/tests/calendar/calendar_mutation_receipt_schema_test.go` | hashed key uniqueness, bounded response, optional O2O call | 143-03A..C |
| 143-03F | Durable bounded inbound-REPLY receipt schema and attendee edge — `STATIC_ACCEPTED` on `846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final docs/whole-boundary review C/I/M `0/0/0`; the intermediate C/I/M `0/0/1` docs-sync finding is superseded/closed | `backend/ent/schema/calendarinboundreplyreceipt.go`, `calendarattendee.go`, `backend/tests/calendar/calendar_inbound_reply_receipt_schema_test.go` | external Ent metadata: exact named SHA-256 length CHECK and owned-child Cascade; runtime RED/GREEN `NOT_RUN` | 143-03B |
| 143-03E | Generate/inspect Ent output in accepted Docker job — 03A/B/C/D/F are accepted; blocked/`NOT_RUN` on the accepted launcher `NOT_READY` and the still-required founder-approved production migration mechanism/gate. The all-A product decision record does not itself establish that technical mechanism or execute/accept a migration. | generated Ent only, never manually edited or committed | clean generation and schema compile evidence | launcher + founder-approved mechanism/gate + 143-03A..D/F |
| 143-04L | Shared deterministic v1 event envelope over the unchanged outbox payload seam — `STATIC_ACCEPTED` at `4525c559da22729aa8bc51ff5336c29dad4f27e9` after the historical initial C/I/M `0/2/2`, independently accepted immutable/payload amendment, and final Terra/high PASS C/I/M `0/0/0`. Terra/xhigh implementation and Terra/high reviews are dispatcher-enforced; actual telemetry `UNKNOWN`. Evidence: `IMPLEMENTATION-143-04L.md`. Freeze: no exported mutable fields/references; JSON-object snapshot and frozen complete bytes; copy-returning `Subject()`/`EventID()` and `MarshalJSON`; semantic zero-offset UTC normalization; stable content-free `ErrInvalidPayload`. Launcher `NOT_READY`; compile/tests/runtime `NOT_RUN`; no Phase/product/global Mail acceptance. | `backend/internal/platform/events/envelope.go`, `backend/tests/calendar/event_envelope_test.go` | explicit inputs; strict subject/version/id/time/org/correlation/payload validation; complete-envelope immutability oracles; no I/O/random/clock | contract 0.1.7 |
| 143-04A | Public event repository/outbox/receipt ports and service constructor — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `aa85f833a806a36f6d6ef2d8e7ae36dca2621574` (parent `ac9ed8e62edb9064f985a0ad7abf3a2ec42d7304`, tree `1e305267cf47f98000ce89d76f9204d9cd84df93`): `event_errors.go` `660578a29c24fe610d3c97072d922c05e876ae83`, `event_repository.go` `591608e87092ea80308a05923b22d201d4d124a9`, `event_service.go` `585fe73ecdc5b8b45d41c6f16bebc4cd7b69d5e0`, tests-first `event_service_contract_test.go` `713cb9599baeaaf16c1b95e833e93ee69a50cf08`; Terra/xhigh author, Terra/high static precommit+exact-SHA PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; executable evidence `NOT_RUN`, launcher `NOT_READY` | `event_repository.go`, `event_service.go`, `event_errors.go`, `backend/tests/calendar/event_service_contract_test.go` | contract tests authored first; no persistence/runtime acceptance | 143-04L; former 03E/no-further-source block is `CONSUMED/HISTORICAL` |
| 143-04G | Admit calendar events to shared EVENTS stream without webhook exposure — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `48a7273d100c07788def88e2f1b8d69c7173580f` (parent `6386f3eeb4025fc0aca55f49aaa8083ff856f8a9`, tree `65c5074a9793d24bac5f9b18aa214fbc942ac922`): `client.go` `e838704e6b4f5ebb9d1bb52d7d36b776c4453738`, tests-first `calendar_subjects_test.go` `71f5f68280eeb6b64e590b809c6df6b618695c9b`; Terra/xhigh author, Terra/high static precommit+exact-SHA PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; executables `NOT_RUN`, launcher `NOT_READY`; former next-route is `CONSUMED/HISTORICAL` | `backend/internal/platform/nats/client.go`, `backend/tests/calendar/calendar_subjects_test.go` | add `calendar.>` once; exact webhook filter unchanged | 143-04A |
| 143-04B | `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `75dafea889055126f768227a4ae62b28c91b7393`; receipt in `IMPLEMENTATION-143-04B.md` | exact accepted source/test blobs | Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` | former current-next route `CONSUMED/HISTORICAL`; no next Calendar code packet selected |
| 143-04C | material/silent/reminder update and master cancellation — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7fd061b5099b4608e0c33a6d99edbe1a66587d34`; former candidate/current-next route `CONSUMED/HISTORICAL` | test first `event_service_update_test.go`, then `event_repository.go`, then new `event_service_update.go`; exact blobs in `IMPLEMENTATION-143-04C.md` | presence patch; authoritative locked organizer resolver; inactive-reader filtering/fresh re-invite/new ID/normalized-email role-resolved-identity rejection; opaque result echo/all-and-only rollback; typed recurrence-exception conflict | accepted 143-03B-T + 143-04A/G; no next Calendar packet selected |
| 143-04D | Series RSVP optimistic transition | `event_service_rsvp.go`, `backend/tests/calendar/event_service_rsvp_test.go` | one version increment/outbox; NFR-13 recovery metadata | 143-04A + 143-04G |
| 143-04E | HTTP occurrence RSVP with persisted mutation receipt | `event_service_occurrence.go`, `backend/tests/calendar/event_service_occurrence_test.go` | materialized override; replay/conflict; one receipt/outbox | 143-04D + 143-04G |
| 143-04H | Safe receipt-free cancellation of one recurrence instance — `ROOT_SOURCE_AUTHORITY_CANDIDATE / IMPLEMENTATION_BLOCKED` | test first `backend/tests/calendar/event_service_occurrence_cancel_test.go`, then `backend/internal/calendar/event_repository.go`, then new `backend/internal/calendar/event_service_occurrence_cancel.go` | capability before backend; nondisclosing masking; bounded 04B membership; idempotent cancelled exception; sequence/version +1; one safe occurrence-cancel EnvelopeV1; exact echo/rollback; zero serializer/iTIP | accepted 143-04A/B/C/G/L; separate exact15 acceptance required |
| 143-04I | Sender-bound inbound iTIP REPLY apply/quarantine | `event_service_inbound_reply.go`, `backend/tests/calendar/event_service_inbound_reply_test.go` | A→B→A no rollback; same-status receipt; stale/future/spoof/cap quarantine | 143-02B + 143-04A + 143-04G |
| 143-04J | Pure canonical ScheduledCall mapping, recurrence and next-effective-occurrence policy | `event_call_policy.go`, `backend/tests/calendar/event_call_policy_test.go` | exact-channel/project/org membership; linked-guest rejection; series participant projection; DTSTART/RRULE/EXDATE; pause/resume; no I/O | 143-01 + 143-02A + 143-04A/B/G |
| 143-04K | Pure due-call source receipt — `SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN` only | authored-first `event_call_due_service_test.go` blob `437db3af`, then `event_call_due_service.go` blob `376f46c1` in immutable source `572f9112`; no dirty-tree authority | frozen pure 04J-derived due decision, opaque descriptor; no I/O/clock/CAS/outbox append/NATS | accepted 143-04J + 143-04A + 143-04L |
| 143-04F-Base | Ent adapter shell and shared transaction primitive | `event_ent_repository.go`, `backend/tests/calendar/event_ent_repository_contract_test.go` | interface assertion; injected DB/outbox; no network/post-commit publish | 143-03E + 143-04A/G |
| 143-04F-B | Create-event Ent transaction | `event_ent_repository_create.go` | receipt/event/attendees/outbox/completion in one transaction | 143-04F-Base + 143-04B |
| 143-04F-C | Update/cancel-master Ent transaction | `event_ent_repository_update.go` | event-first lock/version/update/reset/outbox atomic | 143-04F-Base + 143-04C |
| 143-04F-D | Series-RSVP Ent transaction | `event_ent_repository_rsvp.go` | event→optional call→attendee; compare/update/series participant projection/receipt/outbox atomic | 143-03D + 143-04F-Base + 143-04D |
| 143-04F-E | Occurrence-RSVP Ent transaction | `event_ent_repository_occurrence.go` | event/exception/attendee lock order; inherited override/receipt/outbox atomic | 143-04F-Base + 143-04E |
| 143-04F-H | Occurrence-cancel Ent transaction | `event_ent_repository_occurrence_cancel.go` | event→optional call→exception lock; version+sequence/projection/outbox atomic | 143-04F-Base + 143-04H |
| 143-04F-I | Inbound-REPLY Ent transaction | `event_ent_repository_inbound_reply.go` | event→optional call→optional exception→attendee; receipt ledger + series-only participant projection | 143-03D + 143-04F-Base + 143-04I |
| 143-04F-J | Paired event/ScheduledCall create-link-edit-delete transaction | `event_ent_repository_call_bridge.go` | event→call lock; org/auth/link invariant; canonical mutation + projection + invite outbox atomic | 143-03D + 143-04F-Base + 143-04B/C/H/J |
| 143-04F-K | Canonical due projection transaction | `event_ent_repository_call_due.go` | event→call lock; current occurrence recheck; one due outbox + advance/deactivate atomic | 143-04F-J + 143-04K |
| 143-05A | Canonical v2 HTTP DTO binding/validation/error metadata | `event_dto.go`, `backend/tests/calendar/event_handler_contract_test.go` | owner ignored from client; temporal pairs; exact route-specific 409 codes/recovery headers; occurrence cancel | 143-04B/C/D/E/H + accepted 143-04F-B/C/D/E/H |
| 143-05B | `/api/calendar/v2/**` handlers and composition | `event_handler.go`, `backend/internal/calendar/handler.go`, `backend/cmd/server/wiring_modules.go` | exact OpenAPI paths; legacy handler untouched | 143-05A |
| 143-05C | Add first-class event to legacy backend aggregate without envelope drift or linked-pair duplicate | `dto.go`, `repository.go`, `service.go`, focused genuine backend units | linked pair once as canonical-backed `call`; unlinked event as `event`; existing filters/export preserved | 143-05B + 143-05E |
| 143-05D | Add the new aggregate event kind to the existing frontend calendar client | `frontend/src/api/calendar.ts`, focused genuine frontend unit | existing kinds byte-shape compatible; additive event decoder/type only | 143-05C + frontend feature-test harness |
| 143-05E | Delegate legacy ScheduledCall create/update/delete to canonical bridge | `backend/internal/calls/calendar_bridge.go`, narrow `scheduled.go`/`handler_scheduled.go`/`dto.go`, `backend/cmd/server/wiring_integrations.go`, both shipped create clients, genuine units in allowed roots | compatible success shape; stable key/replay; exact-match attach; paired atomicity; last-write-wins; exact `SCHEDULED_CALL_NOT_MUTABLE`; pause/resume vs delete | 143-05B + 143-04F-J |
| 143-05F | Canonicalize ScheduledCall list/conflict/due readers | narrow `backend/internal/calls/repository.go`/`scheduled.go`, `backend/cmd/server/wiring_integrations.go`, genuine policy/service unit | linked recurrence/exceptions; exact-channel detail auth; due uses 04F-K; pre-cutover unlinked bounded; post-cutover unlinked fails | 143-05E + 143-04F-K |
| 143-05G | Remove client-side dual recurrence authority from exact-channel call calendar | `frontend/src/api/calls.ts`, `frontend/src/modules/calls/call-calendar.tsx`, narrow `calendar-types.ts`, `frontend/tests/mail/**` | paired from/to; canonical occurrences; channel-room lifecycle; due.v1 refetch without toast; no raw RRULE | 143-05F + frontend feature-test harness |
| 143-06 | Founder-reviewed production migration/backfill and all-or-nothing cutover artifact | mechanism chosen by founder; `MIGRATION-PLAN.md` is authority meanwhile | full mapping, recurrence oracle, counts/hashes, no historical outbox, bridge postconditions | founder #12/#18 + launcher + 143-05E/F |

## Cross-packet hard rules

- Existing `Service`, `Repository`, `GET /api/calendar/events`, and `.ics` export remain
  operational; new canonical types use distinct names until additive integration.
- `CalendarEvent` never stores an occurrence `recurrence_id`; exceptions do.
- A linked `CalendarEvent` is the sole meeting authority. `scheduled_call_id` is immutable
  after create; no direct linked-call writer/worker or automatic mismatch winner is allowed.
- Nullable attendee uniqueness requires two partial PostgreSQL indexes.
- Event create and RSVP use `CalendarMutationReceipt`; mail submission uses
  `MailSubmission`. Outbox is not an idempotency response store.
- Every new producer builds one complete `events.EnvelopeV1` through 143-04L and passes it to
  the unchanged `PublishInTx(ctx, tx, envelope.Subject(), envelope)` seam. Stored event type,
  broker subject and JSON subject are identical. Envelope `event_id` is the application
  dedupe id; the outbox row UUID/`Hub-Event-Id` is a distinct diagnostic transport id.
  Legacy outbox payloads and the worker are not wrapped or rewritten.
- Inbound iTIP REPLY uses the bounded `CalendarInboundReplyReceipt` ledger, never a
  `CalendarMutationReceipt`; packet 143-04F-I is reviewed independently.
- Linked and reminder-relevant transactions share one global order: event → optional
  ScheduledCall → optional exception → optional attendee → stable user-policy guard →
  receipt/outbox/notification. Skipped row types do not change the order. This lets Phase
  147 serialize last-chance reminder revalidation without deadlock or a stale call/reminder
  projection.
- Phase 143 admits `calendar.>` to the shared EVENTS stream before any calendar producer;
  linked-call due uses versioned `calls.channel.{channelId}.scheduled_due.v1` inside the
  existing admitted `calls.channel.>`; Phase 144 later adds `mail.>`.
  None enters `WebhookFilterSubjects`.
- NFR-13 body remains exactly `{error,request_id}`. Machine conflict codes use
  `X-Hub-Error-Code`; stale RSVP additionally returns current-version `ETag` and a same-
  origin canonical-event `Location`.
- No generated Ent file is manually edited or committed. No schema migration is invented.
- No host test/build/lint/generate/validator command and no direct Docker/Compose fallback.
- No change enters protected QA/test-rebase paths.

## 143-04J source-authority candidate — not active

The exact13 documentation candidate has one future ordered manifest: external test first
`backend/tests/calendar/event_call_policy_test.go`, then
`backend/internal/calendar/event_call_policy.go`. Its frozen ABI is in
`REVIEW-143-04J-CALL-PROJECTION-AUTHORITY.md`: no I/O, an opaque snapshot result, canonical
timed recurrence boundary and series-only participant projection. It is not current source
authority. Only atomic exact13 docs commit, Contract/Spec/Oracle, Security and Lifecycle immutable
exact-SHA PASS, and separate docs acceptance may authorize exact2. No 04K/04F, Ent/generation/
migration, adapter/API/UI, raw iTIP or runtime route follows. At scoped acceptance this review
consumes only the generic 04J row's ABI/behavior/oracle wording; every other route and blocker
remains current.

## 143-04K source receipt-sync — conditional

> **Source receipt-sync gate:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. The dirty exact15 documentation candidate has
> no authority. After one atomic receipt-sync commit and immutable acceptance, source `572f9112`
> (parent `26b377eb`, tree `ac96e2f5`, subject `feat(calendar): add call due policy`) is
> `STATIC_ACCEPTED / EXACT_SHA_PASS` for test blob `437db3af0263cedb10dcf217cd53fbbcac03cf23`
> then source blob `376f46c1cc79956fab3479efc38750fd88161b16`. Contract/Spec/Oracle, Security/Adversarial
> and Lifecycle/Dependency/Boundary precommit plus immutable reviews PASS C/I/M `0/0/0`;
> Terra/xhigh, dispatcher-enforced/no fallback, telemetry `UNKNOWN`. No next packet is selected.

`REVIEW-143-04K-CALL-DUE-AUTHORITY.md` is the accepted authority predecessor. Its lower
authority-candidate/precommit/current-next clauses become `CONSUMED/HISTORICAL` only upon this
receipt's activation. Source consumes accepted 04J `d7926a27`/`e74efa827`, 04A descriptor and 04L
envelope values, but claims no Calendar-subject admission from 04G. The policy is pure: it deep-clones raw projection and calls
accepted 04J only through two fixed reconcile inputs, then validates the existing plural nine-field
due envelope including every payload/envelope/subject binding. 04F-K separately candidate-calls,
reserves/reloads EventID/OccurredAt/CorrelationID only after due (non-due candidates discard),
reruns byte-identically, and owns event→call CAS/locks/rollback and transactional
descriptor append. AQ-02 is backfill-only; Ent/generation/migration, adapter/API/UI, raw iTIP,
legacy cutover, launcher and runtime remain blocked or `NOT_RUN`.

## Gate status

| Gate | Status |
|---|---|
| Contract root/independent review | Calendar CVE-v1 identity semantics PASS on `b3fb6f48`; global Mail 0.1.7 later reopened for engine/schema bounds, and calendar list/legacy aggregate HTTP bounds are also P1-open |
| Requested-route provenance | 03B requested Terra/high independent re-review, dispatcher-enforced, actual telemetry `UNKNOWN`; 03F author Terra/xhigh and independent reviewer Terra/high, dispatcher-enforced, actual telemetry `UNKNOWN` |
| Accepted Docker launcher | NOT_READY |
| Ent generation | NOT_RUN |
| Production migration authority | UNRESOLVED |
| Runtime/tests/build | NOT_RUN |
