# Mail workstream pre-implementation readiness

> **143-04J source-receipt sync:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT14_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. These proposed receipt facts grant no
> authority before one atomic exact14 docs commit. After it, immutable exact2 `e74efa827b52...`
> is `STATIC_ACCEPTED / EXACT_SHA_PASS`; its accepted ABI stays current while all lower
> implementation-blocked/source-unaccepted clauses become `CONSUMED/HISTORICAL` only for 04J.

> **143-04J acceptance-sync gate:** `ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT13_COMMIT_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`.
> These proposed receipt facts have no authority before one atomic exact13 docs commit. After it,
> immutable candidate `5841621720cc03682b3d1f1e2af9cba2953797cf` is
> `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS`, its three immutable review dimensions PASS C/I/M
> `0/0/0`, and lower 04J candidate routes are `CONSUMED/HISTORICAL`. Only the frozen ordered
> test-first exact2 becomes dispatchable; it remains implementation-unaccepted.

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

> **Calendar route (2026-08-15):** 143-04G and 143-04B are `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact 04B receipt is in `IMPLEMENTATION-143-04B.md`. Former next-route clauses are `CONSUMED/HISTORICAL`. No next Calendar code packet is selected; root must separately dispatch any later packet. Executables `NOT_RUN`, launcher `NOT_READY`; 03E stays blocked only for Ent generation/migration.
> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); three exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` stay accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; unrelated gates remain authoritative. No source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher readiness; carrier six-path source additionally awaits accepted Rule compiler source/ABI. 145-07B source/runtime blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness block Rule Task 1/145-07B; launcher `NOT_READY`, executables `NOT_RUN`.

> **Rule-plan acceptance receipt (2026-08-14):** standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole added plan blob `b86c0423a307fd2673a3d322435922aa2a398e8f`). Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. All former 02F and Rule-plan authoring/review/acceptance instructions are `CONSUMED/HISTORICAL`. No Rule source, Ent schema, generation, migration, service, runtime, product, phase or release authority follows. Rule Task 1 RED remains blocked on all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, unordered; post-task runtime `NOT_RUN`. 144-01B stays Docker-blocked; Phase 142/148 pending; launcher `NOT_READY`, executables `NOT_RUN`.

> **Current Mail gate:** 144-02D retains its narrow docs-only acceptance. 144-02F correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS `0/0/0`, dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Former lifecycle and Rule-plan authoring/review are `CONSUMED/HISTORICAL`. No implementation/runtime authority follows; Rule TDD/production awaits all accepted non-marker Phase-145 carrier/G145-5B/G145-5C and launcher readiness.
> **Snapshot:** 2026-08-14
> **Branch:** `codex/mail-module-20260813`
> **Product truth:** IMPLEMENTATION_IN_PROGRESS · 143-01/02A/02B/03A/03B/03C/03D/03F/04L/04A/04G_STATIC_ACCEPTED · RUNTIME_NOT_RUN · Phase/product/migration/runtime not accepted
> **Execution truth:** launcher `NOT_READY`; generation, migration, tests and runtime `NOT_RUN`.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`). Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan; former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance.

> **143-04B source receipt:** `75dafea889055126f768227a4ae62b28c91b7393` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact blobs in `IMPLEMENTATION-143-04B.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Old 04B `PRECOMMIT_FAIL`/`NOT_ACCEPTED`/current-next clauses `CONSUMED/HISTORICAL`. No next Calendar code packet selected; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`; 03E Ent generation/migration blocked.

## Prepared artifacts (root and independently reviewed)

- Architecture, source-of-truth, security and compatibility decisions include the exact
  five-document default-vs-hard-cap/144-01A entry amendment, which passed at
  `5e314fb080d420696642058ae47397e251c23207`. This is not a full Mail 0.1.8, Phase 142 or
  global cross-contract PASS. The bounded 144-01A engine DTO/error boundary is `STATIC_ACCEPTED` at
  `e3e8c774d734f926f77f8e600b58dc68dabbd993` after `EXACT_SHA_PASS`. The 0.1.7
  `b3fb6f48` and 0.1.6 `5a351d5b` evidence remain historical.
- `144-03A-Thread` callable/encoding/oracle is separately
  `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
  `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` after independent
  Spec/API/Persistence, Security and Lifecycle/Evidence PASS reviews with C/I/M `0/0/0`.
  Implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`
  (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`), static only. Folder source+oracle are
  narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
  `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
  C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
  `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately docs-only `STATIC_ACCEPTED` /
  `EXACT_SHA_PASS` at `f199546513b758ca7b95328b5ccd7680009b732e`, transition
  `6735ae746c7acde657359e8a709cd897c4dce61c`. Contract/Spec Terra/high, Security Terra/xhigh and
  Lifecycle Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless
  exposed. 144-02D and 02F docs/readiness are accepted; the former lifecycle and Rule-plan authoring/review are `CONSUMED/HISTORICAL`. Rule TDD/production awaits all accepted non-marker Phase-145 carrier/G145-5B/G145-5C and launcher readiness.
  No execution.
- OpenAPI covers mailbox, submission recovery, legacy calendar, canonical calendar and
  separate legacy/new admin resources.
- Event subjects and webhook privacy boundary are frozen.
- Phase 143 calendar, Phase 144 mailbox, Phase 145 gateway and Phase 146 UI work are split
  into exact-file atomic dispatch indexes; high-risk behaviors have full prompts.
- Historical independent contract/security and dispatch atomicity rereviews returned static
  `PASS` after occurrence, schema-closure, inbound-REPLY-ledger, gateway ownership and
  mobile-navigation remediation. They target a pre-0.1.4 ancestor and remain `STALE` as
  independent current-contract evidence; the new root review does not replace them.
- Phase 145 has a nine-gate legal/provenance/runtime/ops plan.
- Founder questions have safe unattended defaults and authorize no external action.
- Founder question 11 is answered: exact `gpt-5.6-terra` / `xhigh`, no fallback, is the
  task-local atomic-code route for this Mail branch; mandatory first-read `AGENTS.md` records
  the exact narrow exception and links its provenance record.

## Hard blockers

| Blocker | Evidence | Unblocks |
|---|---|---|
| Accepted versioned Docker launcher not implemented | root `AGENTS.md`, `STATE.md` | Ent generation, build/test/validator/runtime evidence |
| Accepted versioned Docker launcher not implemented; separate founder-approved production migration mechanism/gate still required | root `AGENTS.md`, `STATE.md`, `MIGRATION-PLAN.md`, all-A non-acceptance boundary | 143-03E generation/inspection evidence; all 03A/B/C/D/F schema packets are already static accepted |
| 144-01B requires applicable supported current-SHA Docker evidence under stricter Task3 authority | Phase 144 `IMPLEMENTATION-144-01A.md`, `REVIEW-144-01A-ENGINE-BOUNDS.md`, `DISPATCH-REMAINING.md` | 144-01B deterministic fake only; Folder source/oracle and docs are static-only accepted; Rule plan is separately docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761` (parent `77eb`, tree `1b554d`, plan blob `b86c0423`), reviews PASS C/I/M `0/0/0`; former authoring/review is `CONSUMED/HISTORICAL`. Rule TDD/production awaits all accepted non-marker Phase-145 carrier/G145-5B/G145-5C and launcher readiness |
| 144-02A representation is docs-only accepted at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931`, but source schemas are unimplemented and source review absent; generation remains launcher-gated | Phase 144 `REVIEW-144-02A-SCHEMA-READINESS.md`, `DISPATCH-INDEX.md` | 144-02A source and every dependent mailbox schema/service packet remain BLOCKED pending founder migration authority, launcher and sequential source dispatch/review |
| Calendar DTO/handler implementation is unimplemented; accepted 0.1.8 OpenAPI page/expansion bounds still lack executable evidence | five-document 144-01A entry amendment PASS `5e314fb080d420696642058ae47397e251c23207`; accepted 0.1.8 OpenAPI page/expansion bounds are a contract fact, Phase 142 `REVIEW-OPENAPI-RESPONSE-BOUNDS.md` | calendar list/legacy aggregate DTO, handler and UI transport packets |
| Production migration mechanism unresolved | `MIGRATION-PLAN.md`, founder question 12 | schema generation/migration/backfill acceptance |
| Recurring ScheduledCall semantics unresolved | `MIGRATION-PLAN.md`, founder question 18 | recurring backfill and all-or-nothing cutover |
| Legacy POST idempotency compatibility break unaccepted | OpenAPI, founder question 22 | enabling the ScheduledCall bridge cutover; both shipped clients must move atomically |
| Stalwart license not selected | founder question 1, Phase 145 `GATE-PLAN.md` | gateway/image adoption |
| Immutable image/config/API provenance not refreshed | Phase 145 `GATE-PLAN.md` | gateway wire implementation and Docker profile |
| Protected frontend test cutover not handed off/merged | Phase 146 `DISPATCH-INDEX.md` | executable genuine frontend units |
| Legacy broadcast governance unresolved | founder question 13 | Phase 148 mailbox acceptance |

## First authorized execution sequence

1. ~~Dispatch 143-01 as exact `gpt-5.6-terra` / `xhigh` on the recorded clean HEAD; root
   records actual provenance and reviews the allowed-path diff.~~ `STATIC_ACCEPTED` on
   `e7d5f450`; evidence: `IMPLEMENTATION-143-01.md`; runtime `NOT_RUN`.
2. ~~Accept or reject that packet before any dependent dispatch.~~ Independent Terra/high
   re-review PASS with no P0/P1/P2 after remediation.
3. ~~Dispatch 143-02A and 143-02B.~~ Both are `STATIC_ACCEPTED`; 143-02B accepted through
   `efa7814d` after parser hardening and host-local timezone rejection. Evidence:
   `IMPLEMENTATION-143-02A.md` and `IMPLEMENTATION-143-02B.md`; runtime `NOT_RUN`.
4. 143-03A is `STATIC_ACCEPTED` on `31bfff8f58ba83d12b59600135503f98a6c92fbd`;
   143-03C is `STATIC_ACCEPTED` on `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`;
   143-03B is `STATIC_ACCEPTED` on clean `dd346e31`. 143-03F is `STATIC_ACCEPTED` on
   `846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final
   docs/whole-boundary review C/I/M `0/0/0`; the intermediate C/I/M `0/0/1` docs-sync finding
   is superseded/closed. 143-03D is `STATIC_ACCEPTED` on
   `10b4ceead119cb74a16b325b33f9121f60d7b1b5` after final independent review C/I/M `0/0/0`.
   Do not start generation: 143-03E is `BLOCKED/NOT_RUN` on the accepted launcher `NOT_READY`
   and the separate founder-approved migration mechanism/gate; 03F is satisfied.
5. 143-04L is `STATIC_ACCEPTED` at `4525c559da22729aa8bc51ff5336c29dad4f27e9` after final
   independent Terra/high PASS C/I/M `0/0/0`; requested Terra/xhigh implementation and Terra/high
   review provenance are dispatcher-enforced, actual telemetry `UNKNOWN`. Its static RED only
   does not make the launcher ready.
6. 143-04A is `STATIC_ACCEPTED` at `aa85f833a806a36f6d6ef2d8e7ae36dca2621574`; tests were authored first and its Terra/xhigh author/Terra/high static precommit+exact-SHA reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Its exact four source/test blobs are recorded in `IMPLEMENTATION-143-04A.md`. Separate hygiene `6386f3ee` changed only the stale accepted 04L test import, with static PASS `0/0/0` and no behavior change. Former 03E/no-further-source restrictions on pure 04A are `CONSUMED/HISTORICAL`; 03E remains blocked for Ent generation/migration.
7. 143-04G and 143-04B are `STATIC_ACCEPTED`; their former next routes are `CONSUMED/HISTORICAL`. No next Calendar code packet is selected; root must separately dispatch any later packet. All executable evidence remains `NOT_RUN`.
8. Resolve the separate migration mechanism/gate and accepted launcher before Ent generation.
9. Admit `calendar.>` before calendar producers.
10. Execute Phase 143 service/HTTP plus ScheduledCall bridge packets; retain the legacy
   calendar envelope and count linked pairs once. Keep linked external guests fail-closed
   until founder question 23 has a separately reviewed join-credential contract.
11. Phase 144-01A is `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993` after
   exact-commit Terra/high `EXACT_SHA_PASS`, Spec/Quality/Security/Evidence PASS and C/I/M
   `0/0/0`. Exact native dispatch provenance is dispatcher-enforced; actual telemetry `UNKNOWN`
   is not a telemetry-only blocker. `144-03A-Address` is `STATIC_ACCEPTED` at
   `41564d003946ddb78df18efca166ca86e4055405` after exact-SHA PASS. `144-03A-Thread`
   callable/encoding/oracle is `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
   `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` with independent
   Spec/API/Persistence, Security and Lifecycle/Evidence PASS reviews and C/I/M `0/0/0`.
    Its implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`
    (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`), static only. Folder source+oracle are
    narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
    `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
    C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
    `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately docs-only `STATIC_ACCEPTED` /
    `EXACT_SHA_PASS` at `f199546513b758ca7b95328b5ccd7680009b732e`, transition
    `6735ae746c7acde657359e8a709cd897c4dce61c`. Contract/Spec Terra/high, Security Terra/xhigh and
    Lifecycle Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless
    exposed. 144-02D and 02F docs/readiness are accepted; the former lifecycle and Rule-plan authoring/review are `CONSUMED/HISTORICAL`. Rule TDD/production awaits all accepted non-marker Phase-145 carrier/G145-5B/G145-5C and launcher readiness.
    No execution.
   `144-01B` remains blocked by applicable
   supported current-SHA Docker evidence and executable status `NOT_RUN`.
12. Admit `mail.>` only after calendar admission; keep webhook filter unchanged.
13. Execute UI transport/routes/core only after backend contracts are accepted and the
   feature-test harness is available.
14. Begin Phase 145 only in the gate order; external SMTP remains separately disabled.

## Forbidden shortcuts

- Root Sol or any runtime other than exact `gpt-5.6-terra` / `xhigh` writing atomic
  Mail product/test code;
- silent model fallback;
- host test/build/lint/validator/generator/process execution;
- direct Docker/Compose as accepted evidence;
- manual edits to generated Ent/router artifacts;
- reading or changing protected test-rebase/QA/pro-qa-ai zones;
- raw draft/message body or idempotency-key persistence in PostgreSQL;
- project membership as mailbox attachment authorization;
- blind retry after unknown draft save or submission;
- push, PR, merge, deployment, DNS, license purchase or secret mutation without authority.

> **Current Rule acceptance:** docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
> `f199546513b758ca7b95328b5ccd7680009b732e`, transition `6735ae746c7acde657359e8a709cd897c4dce61c`.
> Contract/Spec Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced,
> no fallback, telemetry `UNKNOWN` unless exposed. 144-02D is docs-only accepted. The former 02F closure
> attempt and transition route are `CONSUMED/HISTORICAL`; correction `fcb5b84f` plus transition `2f4381a2`
> is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` with all three Terra routes PASS C/I/M
> `0/0/0`. Rule-plan authoring/review is `CONSUMED/HISTORICAL`; Rule TDD/production awaits all
> accepted non-marker Phase-145 carrier/G145-5B/G145-5C and launcher readiness.

> **143-04J docs-only candidate:** the exact13 documentation packet is
> `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED` only. Its full pure-policy ABI,
> two-file test-first manifest and non-I/O boundary are in
> `REVIEW-143-04J-CALL-PROJECTION-AUTHORITY.md`. Atomic exact13 commit, independent
> Contract/Spec/Oracle, Security and Lifecycle immutable exact-SHA PASS, then separate docs-only
> acceptance are mandatory before dispatch. AQ-02/backfill, 04K/04F, Ent/generation/migration,
> adapter/API/UI, raw iTIP, launcher and executable/runtime boundaries remain unchanged.

> **143-04K source receipt candidate:** the immutable exact2 is recorded in
> `IMPLEMENTATION-143-04K.md`; this dirty exact15 documentation candidate is not a source,
> persistence or admission decision. It consumes accepted 04J
> `d7926a27`/`e74efa827`, 04A and 04L, freezes a pure explicit UTC whole-second `Now`/raw-
> projection/envelope due decision with optional descriptor and the existing plural nine-field due
> payload. Active projection is deep-cloned and reconciled by exactly two fixed 04J calls;
> inactive/deleted requires nil projection/zero envelope and noops. It claims no 04G
> Calendar-subject admission. The former authority-candidate/precommit/test-first-dispatch route is
> `CONSUMED/HISTORICAL` only after this receipt activates. 04F-K alone later owns CAS/locks/rollback/same-tx outbox; AQ-02,
> Ent/generation/migration, adapter/API/UI, raw iTIP, legacy cutover, launcher and runtime remain
> blocked or `NOT_RUN`.

> **143-04K source receipt-sync gate:** `SOURCE_ACCEPTANCE_SYNC_CANDIDATE /
> ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact15 documentation candidate
> grants no authority. Only after one atomic receipt-sync commit and immutable acceptance is source
> `572f911227b9d50e21aa73b87c530ddcaf0f8934` (parent `26b377eb`, tree `ac96e2f5`, subject
> `feat(calendar): add call due policy`) `STATIC_ACCEPTED / EXACT_SHA_PASS` for test blob
> `437db3af0263cedb10dcf217cd53fbbcac03cf23` then source blob
> `376f46c1cc79956fab3479efc38750fd88161b16`. Contract/Spec/Oracle, Security/Adversarial and
> Lifecycle/Dependency/Boundary precommit plus immutable reviews PASS C/I/M `0/0/0`; Terra/xhigh,
> dispatcher-enforced/no fallback, telemetry `UNKNOWN`. No next code packet is selected.

> EventID/OccurredAt/CorrelationID are fresh candidates, not 04K allocation: non-due outcomes
> discard them; only after due may 04F-K reserve or reload them, then rerun 04K byte-identically on
> CAS retry/redelivery. 04K rechecks every payload/envelope/subject binding while retaining no
> reservation storage.

> **143-04C immutable source receipt:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
> `7fd061b5099b4608e0c33a6d99edbe1a66587d34`; the candidate/precommit/current-next lifecycle is
> `CONSUMED/HISTORICAL` only for 04C. `IMPLEMENTATION-143-04C.md` records the test-first exact
> three blobs and Contract/Security/Lifecycle PASS C/I/M `0/0/0`. No next Calendar code packet is
> selected. Launcher `NOT_READY`, all executable evidence `NOT_RUN`, and 03E generation/migration
> blocking remain authoritative.
