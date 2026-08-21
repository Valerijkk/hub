# Phase 144 — exact Terra dispatch index

## 144-04I-A source-authority candidate — non-authoritative

`ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / IMPLEMENTATION_BLOCKED /
RUNTIME_NOT_RUN`. The exact9 manifest and frozen ABI/state/safety/oracles are only in
`REVIEW-144-04I-A-SYNC-WORKER-READINESS.md`. Triple precommit PASS, one atomic exact9 commit,
triple immutable exact-SHA PASS and separate acceptance sync must precede only test-first NEW
`backend/tests/mail/mailbox_sync_worker_test.go`, then NEW
`backend/internal/mail/mailbox_sync_worker.go`. 04F/04G stay accepted; 04H/02E/Ent/generation/
migration/adapters/API/OpenAPI/UI and runtime remain blocked. No schema/event change and no auto-next.

## 144-04G source acceptance-sync candidate (2026-08-21)

> **Gate:** `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact12 working tree grants no authority. The immutable `3031c53d399accc84fe10cd0c5421a85c089ac7c` source receipt and its Contract/Security/Lifecycle precommit plus immutable `EXACT_SHA_PASS C/I/M 0/0/0` facts are proposed only; they activate only after one atomic commit of the exact12 manifest in `IMPLEMENTATION-144-04G.md`. Only after that commit is 144-04G `STATIC_ACCEPTED / EXACT_SHA_PASS`; then, and only then, lower 04G source-authority/source-unimplemented/current-next routes are `CONSUMED/HISTORICAL`. No next Mail packet is selected. 04F remains accepted; 04H/02E, Ent/SQL/concrete adapter, generation, migration, OpenAPI and runtime remain blocked or outside authority; launcher `NOT_READY`, executables/runtime `NOT_RUN`.

## 144-04G source-authority acceptance-sync candidate (2026-08-20; predecessor route)

Status: `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED /
SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This dirty exact9 working tree grants no
authority. The `6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` metadata and the three immutable
exact-SHA PASS results below are proposed receipt facts only; they activate only after one atomic
commit of this exact9 acceptance sync. Until activation, the lower candidate/precommit/root-commit/
exact-SHA-next lifecycle remains current and root cannot dispatch source; the separately false
accepted-repository premise remains `CONSUMED/HISTORICAL`.

After that atomic exact9 sync commit, Mail 144-04G docs authority is
`SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` only at immutable
`6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` (parent
`2b6837926c91d652ff7bf30447bd482740261bd0`, tree
`bdd59cfcb31376c95f5128370a75333cad78bc14`, subject
`docs(mail): propose sync service authority`), with exact9 blobs recorded in the readiness review.
Contract, Security and Lifecycle immutable exact-SHA reviews PASS C/I/M `0/0/0`;
dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Only after that atomic exact9 sync commit is all lower 144-04G
candidate/precommit/root-commit/exact-SHA-next routing `CONSUMED/HISTORICAL`.
Only then may root dispatch only the frozen ordered exact3:
`backend/tests/mail/mailbox_sync_service_test.go`, new interface/types-only
`backend/internal/mail/mailbox_repository.go`, then new
`backend/internal/mail/mailbox_sync_service.go`. That source is still
`UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and needs its own authored-first review, atomic commit and immutable
acceptance. 04F stays accepted; 04H/02E, Ent/adapter, generation, migration, OpenAPI and runtime
stay blocked or outside authority; launcher `NOT_READY`, executables `NOT_RUN`.

## 144-04G candidate precedence (2026-08-20; predecessor route)

Status: `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED /
IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. Exact3 future manifest, test first: NEW sync-service
test, NEW abstract `mailbox_repository.go` interface/types, NEW sync service. The old accepted-
repository premise is `CONSUMED/HISTORICAL` only for that phrase. Exact ABI, fences, bounded
one-page Step, enumeration/catch-up/rebuild, deterministic 04F event candidates, no-I/O-under-
lock and static oracles are frozen by `REVIEW-144-04G-SYNC-SERVICE-READINESS.md`. Source dispatch
requires triple precommit PASS -> atomic exact9 -> triple exact-SHA PASS -> separate acceptance
sync. 04F receipt, Calendar 04C state and all 04H/02E/launcher/runtime blockers are unchanged.
Fix1 requires DB-transaction-time lease expiry on Prepare/Apply/Resolve (expired means LeaseLost/
zero effects even without takeover) and exact configured PageSize in every Changes/enumeration
request; request.Now is never mutation-expiry authority.

## 143-03B-T immutable source acceptance and conditional 143-04C gate (2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; 144-04G still requires separate docs authority because its repository port is absent. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## 144-04F immutable source acceptance (2026-08-20)

Mail 144-04F is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `4715f939f24e5b5c2d3f80a38d204babe68bdd34` (parent `7ac8714e99db2935ce290ad67836e1309cc05c78`, tree `6461894a90c4340bf96118ed1a242516c02b48c7`, subject `feat(mail): add sync page policy`). Its exact4 ordered test-first blobs are recorded in `IMPLEMENTATION-144-04F.md`: events test `50a41cefb539d9b840a43da9e1ff34cce6c8653f`, sync-policy test `b6410709e4fb8166c28a5c502cfba535eef08d06`, events source `a3017b0a6064d8fc1017e37e2d4f675ee8732945`, sync-policy source `acd642fcdb6290af79b7fa8e5aa36a5036576f47`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file asserting that Mail exact4 is dirty, `UNCOMMITTED`, `UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is `CONSUMED/HISTORICAL`; the immutable source receipt above is current. No automatic Mail next source follows: 144-04G remains `NOT_AUTHORIZED / BLOCKED` pending separate docs authority because its required repository port is absent.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted. 143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; 143-04A/B/G receipts remain accepted. Launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. 143-03E, 144-02E, Ent generation and migration stay blocked. No raw iTIP or service/API/UI/product/phase/release authority follows.

## Combined exact17 conditional acceptance sync (2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> **144-02C source receipt (2026-08-20):** `STATIC_ACCEPTED / EXACT_SHA_PASS` applies only to `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) and the exact 13 blobs in `IMPLEMENTATION-144-02C.md`. All lower 02C candidate/source-unaccepted/lifecycle-next/implied-next dispatch clauses are `CONSUMED/HISTORICAL`; no next Mail code packet is selected or authorized. Root must make a separate selection. Launcher `NOT_READY`; executable evidence `NOT_RUN`.

> **144-02B projection source-authority acceptance:** combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS`; Contract/high, Security/xhigh and Lifecycle/high corrected-child exact reviews PASS C/I/M `0/0/0`. Root may dispatch only this frozen test-first exact13: new `backend/ent/schema/mailfolder.go`, `backend/ent/schema/mailthread.go`, `backend/ent/schema/mailmessage.go`, `backend/ent/schema/mailmessagefolder.go`, `backend/ent/schema/maillabel.go`, `backend/ent/schema/mailmessagelabel.go`, `backend/tests/mail/mail_projection_schema_test.go`; edits `backend/ent/schema/mailaccount.go`, `backend/ent/schema/calendarevent.go`, `backend/ent/schema/mailmutationreceipt.go`, `backend/tests/mail/mail_account_identity_schema_test.go`, `backend/tests/mail/mail_mutation_receipt_schema_test.go`, `backend/tests/calendar/calendar_event_schema_test.go`.
>
> Only accepted PERSISTENCE 02B tables/checks/indexes/JSON: accepted 02F-R permits exactly `MailMutationReceipt.folders` inverse for `MailFolder.control_receipt_id -> receipt RESTRICT`; actor/message receipt edges stay scalar-only, other receipt edges/checks unchanged. `MailAccount` gains folders/threads/messages/labels; `CalendarEvent.mail_messages` inverse uses Message FK `SET NULL`; rule targets are service-only locked, no DB FK/normalized target invention. Compatibility tests mandatory. Source implementation remains unaccepted and needs test-first authoring, independent reviews, atomic commit, exact-SHA review and status sync. Prior candidate/precommit/correction/failure-next clauses are `CONSUMED/HISTORICAL`; immutable `45fd3369` failure remains history. former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; 02F-D/02E/Ent/generation/migration/runtime/services/API/UI/release BLOCKED; launcher `NOT_READY`, executables `NOT_RUN`; Calendar accepted/no-next/unrelated blockers unchanged. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02F-R current source:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`; authority precursor `a877848e` is `CONSUMED/HISTORICAL`. Former 02B eligibility/not-auto-dispatched clause is `CONSUMED/HISTORICAL`; 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, and 02F-D remains blocked. 02E/generation/migration/runtime/service/API/UI/release remain blocked; launcher `NOT_READY`, executables `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02D source receipt (2026-08-15):** `2242a4e7846944498eb5ed701ab8f47c986bd804` remains `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 02F-R is now source accepted; former 02B eligibility clause is `CONSUMED/HISTORICAL`, 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, and 02F-D remains blocked. Launcher `NOT_READY`; executables/generation/migration/runtime `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` stay accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; no source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness block Rule Task 1/145-07B; launcher `NOT_READY`, executables `NOT_RUN`.

> **Exact14 acceptance:** docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Prior candidate/precommit/root-commit/exact-review/status-sync-next clauses are `CONSUMED/HISTORICAL`. No source/schema/generation/migration/service/test/runtime authority. Engine deletion-port and User deletion-intent remain separately `BLOCKED` / `NOT_DISPATCHABLE`; 02F remains unchanged except dependency trace. Launcher `NOT_READY`, executables `NOT_RUN`.
> **Status:** historical 144-02D/02F accepted facts remain clause-scoped; current exact14 lifecycle controls correction scope · 144_01B_BLOCKED · PHASE145_BLOCKED · RUNTIME_NOT_RUN
> **144-02D/02F current:** 02D retains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 02F correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract/Spec/Oracle, Security/Adversarial and Lifecycle/Dependency/Boundary Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Prior lifecycle and Rule-plan authoring/review are `CONSUMED/HISTORICAL`; implementation/runtime stay blocked. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`.
> **144-03C-Attachment:** deterministic attachment policy is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7b2ea5721684e43559d6f6b4b5fc462d8daafb8b` (parent `4e7c609713983a8f6db6fbfeb7474abd73451e8c`, tree `4c3fe4d27e170c5d3f42b116c39887036ce936c6`; source `3f67ce1b9f2f056b70bfd3070f1c34f98b5c71c1`, test `bf7eda125d5aba1d0cc2594746d14e7bcfd06aa7`). Terra/xhigh authoring; Contract Terra/high, Security Terra/xhigh and ABI/Lifecycle Terra/high precommit+exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Tests authored first. Per-attachment bounds, NFC output/NFKC structural+extension defense, closed MIME/type deny tables, inline CID and scan/SHA clean-only gates are static-only; RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime are `NOT_RUN`, launcher `NOT_READY`. No scanner/storage/network/project/tenant auth/schema/migration/service/API/UI/product/phase/release authority. Its no-further-source/root-only-selection routing is `CONSUMED/HISTORICAL`. Founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker; every exact packet requires its own test-first/static review/commit. No Ent generation/migration/runtime authority follows.
> **Required runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Contract:** exact five-document default-vs-hard-cap/144-01A entry amendment passed at `5e314fb0`; it authorizes only 144-01A entry, while broader Phase 142/cross-contract review remains pending; 0.1.7 evidence is historical
> **Execution:** sequential for overlapping files; independent pure packets may run in parallel only after their shared types are root-frozen and independently accepted

Phase 144 adds a mailbox domain beside the existing Resend alert subsystem. It does not
rename, rewrite or route user mail through `Service`, `AdminService`, `HubHandler` or the
legacy flat `/api/admin/mail/settings` contract. Every packet starts from a root-reviewed
clean HEAD, owns only its listed files, writes external-package genuine units first, makes
no commit, and returns `BLOCKED` if the exact runtime or frozen surface is unavailable.

Complete standalone prompts exist for the high-risk packets named in this directory. For
every remaining row/variant, assemble the Common envelope plus exactly one section from
`DISPATCH-REMAINING.md`; combining variants is invalid.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

> **144-01B static acceptance (2026-08-15):** immutable `0c064363b5ccb657a7c444802e94075c4ca743e5` (parent `48a7273d100c07788def88e2f1b8d69c7173580f`, tree `9c568a0c58c6bf7b8c3171cfd934760b839db623`, subject `feat(mail): add deterministic engine fake`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for fake `12bcd921668e406f07a98f7f3b0618ad8d44e81d` and authored-first test `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f`. Terra/xhigh author; Spec Terra/high, Security Terra/xhigh, ABI Terra/high precommit and exact-SHA Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 01B Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executables `NOT_RUN`; no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority.

> **144-02F-R accepted source receipt:** `ce542cb5de839a1fb5200295f710b0ae45bcdbe6` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact eight blobs in `IMPLEMENTATION-144-02F-R.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Prior authority-only/source-unaccepted and 02B-separate-authority/02C-predecessor route `CONSUMED/HISTORICAL`; 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation unaccepted, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, 02F-D blocked; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

## Dependency graph

```text
144-01A engine DTO/error contract
   └── 144-01B deterministic fake + conformance oracle

144-02A1 typed DNS value -> 144-02A2 domain/policy schemas -> 144-02A3 account/identity schemas (accepted)
144-02B folder/thread/message/membership/label schemas (SOURCE_AUTHORITY_ACCEPTED; root exact13 dispatch authorized, implementation unaccepted)
144-02C object/attachment/draft/submission schemas (waits accepted 02B)
144-02D user-settings/signature/rule/contact schemas
144-02F-R mutation-receipt/target/member schemas (accepted exact authority)
144-02F-D rule-deployment schemas (separate; blocked/not authorized)
   └── 144-02E accepted Docker Ent generation + generated diff review

144-03A address -> threading variants
144-03B folder tree -> rule compiler variants
144-03C rate/loop -> attachment policy variants
143-04L and 144-01A satisfied -> 144-03D deterministic mail event mapper

144-01B + 144-02E
   ├── 144-04A account resolver/repository base
   ├── 144-04J account desired-state service ── 144-04K account-control worker
   ├── 144-04P domain desired-state service ── 144-04Q domain-control worker
   ├── 144-04B folder -> label variants
   ├── 144-04C message projection/read/bulk
   ├── 144-04D draft save/reconcile
   └── 144-04E submission idempotency/reconcile

144-04A + 144-02D/F
   ├── 144-04L user settings/signature service
   ├── 144-04M rule canonical service ── 144-04N rule deployment worker
   └── 144-04O mail mutation receipt service

144-01B + 144-02E + 144-03A/D + 144-04A
   └── 144-04F sync-page policy ── 144-04G sync service ── 144-04I-A bounded worker candidate
                                      └── 144-04H Ent page/rebuild transaction (blocked separately)

144-04* + 144-03*
   ├── 144-05A mail-owned storage port
   ├── 144-05B outbox/projection + EVENTS subject admission
   ├── 144-05C search -> contact -> realtime -> audit variants
   ├── 144-05D safe content/image proxy
   └── 144-05E search projection hydration consumer
        └── 144-06A public handlers/routes
             └── 144-06B mailbox admin service
                  └── 144-06C mailbox admin handler/routes
                       └── 144-07 docs sync + independent review
```

> **Local 01B clause reconciliation:** within this atomic-packet record, older 144-01B Docker-blocked/do-not-start/dependency assertions are `CONSUMED/HISTORICAL` only. The accepted exact boundary is `0c064363b5ccb657a7c444802e94075c4ca743e5`; 02E/Ent/generation/migration/service/API/UI/product/phase/release blockers stay current.

## Atomic packets

| ID | Single owned behavior | Exact primary files | Oracle / acceptance | Entry dependency |
|---|---|---|---|---|
| 144-02A/B/C/D correction plus 02F trace | Docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc`; no schema dispatch authorization. 02F remains unchanged except its dependency trace. Engine/physical deletion and User/GDPR deletion remain separately `BLOCKED` / `NOT_DISPATCHABLE`. | `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` | Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high exact-SHA PASS C/I/M 0/0/0; dispatcher-enforced/no fallback/telemetry UNKNOWN | Separate docs-only engine deletion-port contract/adjudication; no code |

> **02A/B/C/D plus 02F-trace row precedence:** only the correction-scoped 02A/B/C/D clauses and the broad 02F 02D-unchanged/dependency/writer-replay clause are `HISTORICAL/SUPERSEDED` where they conflict with this current row. Every unrelated accepted 02A/02D/02F fact remains authoritative. This row does not record acceptance or authorize dispatch.
| 144-01A | Mailbox engine port, immutable DTOs and stable safe errors — 0.1.8 projection/budget/closed-DTO/draft-key amendment is `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993` after Terra/high `EXACT_SHA_PASS`; executable evidence remains `NOT_RUN` | `backend/internal/mail/mailbox_engine.go`, `mailbox_types.go`, `mailbox_errors.go`, `backend/tests/mail/mailbox_engine_contract_test.go` | projection-only pages; exact request/projection/body/attachment fields; variable-byte/item/page limits; budgeted whole message/blob and draft save/reconcile; reference-only submission; stable first-draft reconciliation | accepted `ENGINE-CONTRACT-0.1.8.md` |
| 144-01B | Deterministic in-memory fake for one account boundary — blocked by stricter Task3 authority pending applicable supported current-SHA Docker evidence | `backend/internal/mail/mailbox_fake.go`, `backend/tests/mail/mailbox_fake_test.go` | same-key submit with opaque result pair; exact/mismatched result hydration; scripted incremental/enumeration pages, invalid cursor/anchor restart; strict account isolation | accepted 144-01A amendment + applicable supported current-SHA Docker evidence |
| 144-02A1/A2/A3 | A1 `5892a90b`, A2 `65c91b48`, A3 `e9a1764b` static accepted | exact blobs in implementation receipts | tests first; executables `NOT_RUN` | former 02C route `CONSUMED/HISTORICAL`; immutable `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for exact13 `IMPLEMENTATION-144-02C.md`; no next Mail packet selected; generation/migration/runtime blocked; launcher `NOT_READY` |
| 144-02B | Projection docs/readiness accepted | listed projection schema paths | checks/indexes/no raw MIME | former 02C route `CONSUMED/HISTORICAL`; immutable `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for exact13 `IMPLEMENTATION-144-02C.md`; no next Mail packet selected; generation/migration/runtime blocked; launcher `NOT_READY` |
| 144-02C | Delivery schemas source receipt accepted | `mailobject.go`, `mailattachment.go`, `maildraft.go`, `mailsubmission.go` | accepted immutable fields/checks | former 02C route `CONSUMED/HISTORICAL`; immutable `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for exact13 `IMPLEMENTATION-144-02C.md`; no next Mail packet selected; generation/migration/runtime blocked; launcher `NOT_READY` |
| 144-02D | User configuration receipt accepted | listed 02D paths | existing 02D semantics | former 02C route `CONSUMED/HISTORICAL`; immutable `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for exact13 `IMPLEMENTATION-144-02C.md`; no next Mail packet selected; generation/migration/runtime blocked; launcher `NOT_READY` |
| 144-02F-R | Receipt schema accepted | exact eight paths | existing receipt constraints | former 02C route `CONSUMED/HISTORICAL`; immutable `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for exact13 `IMPLEMENTATION-144-02C.md`; no next Mail packet selected; generation/migration/runtime blocked; launcher `NOT_READY` |
| 144-02F-D | Rule-deployment schema route — `BLOCKED` / `NOT_AUTHORIZED` | no current source manifest | retained normative deployment contract only; no dispatch authority | separate and nonblocking for 02B/02C; does not follow 02F-R and remains blocked pending a separately accepted authority packet |
| 144-02E | Generate and inspect Ent output | generated Ent tree only through repository launcher workflow | clean generation/compile evidence; no manual generated edits; schema drift report | launcher + 144-02A..D + founder #12 |
| 144-03A | Address normalization is `STATIC_ACCEPTED` at `41564d003946ddb78df18efca166ca86e4055405`; Thread contract is accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461` | exact 19-path Thread boundary is `STATIC_ACCEPTED` at `9903f4df3e5ca8034ea3f043b0386170e83a5101` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`) | exact callable, strict bytewise Message-ID/ref chain, prefix/bounds, raw hash and literal oracle; no I/O | Folder static acceptance plus docs acceptance `ca4e2854`; separate Rule candidate lifecycle governs next gate |
| 144-03B | Folder source+oracle are narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` / `EXACT_SHA_PASS` (source `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`, C/I/M `0/0/0`) | exact two-path static-only boundary; no handler, schema, migration, engine or runtime acceptance | Hub UUID tree: live-slot 100/depth 2/cycle; system rename/reparent/delete protected; create custom; engine IDs mapped before policy | Rule plan docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761` (parent `77eb`, tree `1b554d`, blob `b86c0423`), reviews PASS C/I/M `0/0/0`; former authoring/review `CONSUMED/HISTORICAL`; Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN` |
| 144-03C-Send | Send/rate/loop policy — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `f8f87b09c4fe65be612fcf6d15ba1f4bd5f8aa19`, C/I/M `0/0/0` | `send_policy.go` blob `07f2641f6efa846110e82240a72bc67b4a4f4509`, `send_policy_test.go` blob `ab47f81812f0aecb05ed0c1d17e7b0f7ee340e78` | PERSISTENCE-backed 1,048,576 floor; fail-closed token/HTAB; authored-first static tests, executable `NOT_RUN` | 144-01A |
| 144-03C-Attachment | Deterministic attachment policy — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7b2ea5721684e43559d6f6b4b5fc462d8daafb8b`, C/I/M `0/0/0` | `attachment_policy.go` blob `3f67ce1b9f2f056b70bfd3070f1c34f98b5c71c1`, `attachment_policy_test.go` blob `bf7eda125d5aba1d0cc2594746d14e7bcfd06aa7` | tests first; per-attachment count/size, NFKC structural+extension and MIME deny defense, inline CID, SHA/scan clean-only gates; executable `NOT_RUN` | prior no-further-source/root-only-selection routing is `CONSUMED/HISTORICAL`; founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker, each requiring test-first/static review/commit. No Ent generation/migration/runtime authority; existing blockers unchanged |
| 144-03D | Deterministic mapping of the three mail payloads into shared full v1 envelopes — `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `42c6b9f2bd4383040fc116630727564fcf17dd78`, C/I/M `0/0/0` | `mailbox_events.go` blob `792a4f460f756c4541003bf1ddc0391be637052c`, `backend/tests/mail/mailbox_events_test.go` blob `8947577115f60f4c3b026dcb686014444660f815` | exact schema envelopes; negative bounds/content/no-clock-or-random oracles; static-only, executable `NOT_RUN` | 143-04L and 144-01A satisfied |
| 144-04A | Resolve authenticated user to one organization/account and expose repositories | `mailbox_repository.go`, `mailbox_ent_repository.go`, `mailbox_service.go`, focused owner units | owner binding; begin/finish/reclaim mutation lease+epoch fence; sync lease/page interfaces; zero engine on invalid owner | 144-01B + 144-02E |
| 144-04J | Deterministic account/identity desired-state service | `mailbox_account_control_service.go`, focused unit, narrow repository signatures | eligible user/domain scan; immutable address/key/cutoff; alias/user lifecycle generation; idempotent ensure | 144-01B + 144-02A/E |
| 144-04K | Bounded multi-replica account-control worker | `mailbox_account_control_worker.go`, focused unit | due scan/lease/backoff/context shutdown; captured generation only; disable/reactivate; pending aliases | 144-04J |
| 144-04B | Folder then label metadata commands as two sequential variants | `folder_service.go`/folder unit, then `label_service.go`/label unit | versioned pending folder intent/cap/reconcile; optimistic label CRUD and rule-reference delete guard; message assignment remains receipt-backed 04C only | 144-03B + 144-04A/O |
| 144-04C | Message list/read/bulk projection and label-assignment behavior | `message_service.go`, `message_projection.go`, focused units | post-filter before hydrate; receipt-captured expansion; add/remove label provenance recovery; no duplicate training; engine mutations fenced; per-target result; full changed envelope; explicit freshness | 144-03A/D + 144-04A/O |
| 144-04D | Draft create/save/read/delete with lost-response reconciliation | `draft_service.go`, `backend/tests/mail/draft_service_test.go` | raw body only in engine DTO; command lease + begin/end epochs; create replay; `If-Match`; unknown reconcile | 144-01B + 144-02C/E + 144-03C + 144-04A |
| 144-04E | Send receipt state machine and reconciliation | `submission_service.go`, `backend/tests/mail/submission_service_test.go` | command lease + begin/end epochs; hashed key/request; one engine submit; split opaque result ids/Hub UUID; retained-pair projection-only recovery; exact result binding/origin; one full envelope | 144-04D + 144-03D |
| 144-04L | User settings then signatures as sequential service variants | `mailbox_settings_service.go`/unit, then `mailbox_signature_service.go`/unit | active owner; same-account defaults; optimistic writes/deletes; server sanitizer derives plain; no raw pair trust | 144-02D/E + 144-04A |
| 144-04M | Canonical rule CRUD and desired generation | `mailbox_rule_service.go`, focused unit | account-deleting guard returns stable 409 before receipt, transaction or I/O; otherwise no handler engine call and no transient lease/423 mapping; deployment state honest | 144-02D/F/E + 144-03B + 144-04A |
| 144-04N | Bounded generation-reconciling rule worker | `mailbox_rule_worker.go`, focused unit | mutation lease; fixed script; Get-before-retry; captured generation only; drift repair | 144-04M + 144-01B |
| 144-04O | Generic mail mutation receipt reserve/replay/tombstone | `mailbox_mutation_receipt.go`, focused unit | public v4/v7 key plus internal folder UUID mode; scoped key digest + raw-keyed canonical request HMAC; 30-day replay/recovery; client lifetime reject tombstone; unknown reconcile seam | 144-02F/E + 144-04A |
| 144-04P | Canonical domain desired-state/reconciliation service | `mailbox_domain_service.go`, focused unit, narrow repository signatures | create/delete intent; stable key/generation/hash; verified/default guards; no DNS mutation | 144-01B + 144-02A/E + 144-04O |
| 144-04Q | Bounded multi-replica domain-control worker | `mailbox_domain_worker.go`, focused unit | Ensure/Get-before-retry; public DNS projection; captured generation; deleting/absence; context shutdown | 144-04P |
| 144-04F | Pure incremental/enumeration page classification and projection commands | Original `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` receipt `4e608e890ec12aeed8de965f9649ae503260b94e`; its immediate exact4 root-dispatch clause is `SUSPENDED / CONSUMED/HISTORICAL`. `ROOT_CORRECTION_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`: existing exact4 remains only the future test-first manifest (`mailbox_events_test.go`, `mailbox_sync_policy_test.go`, `mailbox_events.go`, `mailbox_sync_policy.go`). Uncommitted candidates are `UNACCEPTED` and MUST NOT be committed until Contract/Security/Lifecycle precommit PASS, atomic exact9 correction-candidate commit, all three immutable exact-SHA PASS and separate docs-only acceptance sync; no correction acceptance | exact page union/bounds (Requested 1..100, Maximum=100, Returned exact, nonnil slices, ENGINE-CONTRACT byte budget; change combined 100; enumeration/folders 100); content-free typed error/sentinel ABI; exact cutoff: nonnil ReceivedAt, after requires marker+one received, before/equal zero reservation+nil marker+no received, confirmed owner never received; changed/deleted event set; event folder/label caps 50->100 only; no I/O/clock | 144-01B + 144-03A/D |
| 144-04G | Lease/epoch-aware mailbox sync orchestration against fake ports | `ROOT_SOURCE_AUTHORITY_CANDIDATE` only; future exact3: NEW `backend/tests/mail/mailbox_sync_service_test.go`, NEW interface/types-only `backend/internal/mail/mailbox_repository.go`, NEW `backend/internal/mail/mailbox_sync_service.go`; implementation blocked | exact claim/prepare/apply/resolve ABI; DB-transaction-time lease expiry on every mutation guard; full lease/sync/mutation fence; one ListFolders + one Changes/Enumerate with exact configured PageSize; base enumeration, catch-up/finalize; invalid cursor next-step rebuild; deterministic 04F IDs; no external I/O/callback under lock | 144-04A/F; triple precommit + exact9 + triple exact-SHA + separate acceptance sync |
| 144-04H | Atomic Ent incremental/rebuild projection adapter | `mailbox_ent_repository_sync.go`; no Go DB/component test | account-lock lease/cursor/epoch proof; projection/events/cursor atomic; no-clear rebuild/final cleanup | 144-02E + 144-04G |
| 144-04I-A | Bounded context-owned worker | `ROOT_SOURCE_AUTHORITY_CANDIDATE` only; exact2 is NEW `backend/tests/mail/mailbox_sync_worker_test.go`, then NEW `backend/internal/mail/mailbox_sync_worker.go` | DB-time ordered scan <=50 preserving fractional timestamps; fixed pool <=4/FIFO; total timed 04G step; LeaseEpoch full fault-CAS/checked result/O(1) retry; disabled zero calls; no I/O under local lock | accepted 144-04F/G; triple precommit -> exact9 -> triple exact-SHA -> separate acceptance sync; 04H remains blocked |
| 144-04R | Generic mutation recovery step in the existing sync worker | `mailbox_mutation_recovery.go`, narrow worker edit, focused unit | oldest-due work; receipt+mutation leases; folder recovery only under proven exclusive-writer capability; captured bulk only; unknown training no repeat; no fifth worker | 144-04B/C/I/O |
| 144-05A | Mail-owned object storage authorization and lifecycle | `mailbox_storage.go`, `mailbox_storage_adapter.go`, `backend/tests/mail/mailbox_storage_test.go` | `mail/<org>/<account>/<uuid>` server key; receipt-safe same-object presign replay; confirm/read/delete reauthorize; no project metadata | 144-02C/E + 144-03C + 144-04A/O |
| 144-05B | Versioned projection consumption and mail subject admission | `mailbox_projection.go`, `backend/internal/platform/nats/client.go`, focused units | subject/envelope match; application-event dedupe; unknown/malformed dead-letter; add `mail.>` once; webhook filter byte-equivalent | 144-03D + 144-04C/E + 143-04G |
| 144-05C | Search, contacts, realtime and named audit as four sequential variants | one of `mailbox_search.go`, `mailbox_contacts.go`, `mailbox_realtime.go`, `mailbox_audit.go` plus its unit | owner filters before snippets; ranking; bounded invalidation; exact named audit | 144-04A..I as applicable |
| 144-05D | Safe body sanitization then remote-image proxy as sequential variants | `mailbox_content.go`/unit, then `mailbox_remote_image.go`/unit | reviewed allowlist or plain-only; token rewrite; every redirect/connect safe; no origin URL/browser leak | 144-04A + 144-04L |
| 144-05E | Owner search document hydration/consumer | `mailbox_search_projection.go`, focused unit | bounded sanitized body + filenames; no raw/BCC/URL; delete/update idempotent; checkpoint freshness | 144-05B/D + 144-05C-Search |
| 144-06A | Public mailbox DTO/error mapping and routes | `mailbox_dto.go`, `mailbox_handler.go`, `backend/cmd/server/wiring_mailbox.go`, focused handler units | exact OpenAPI 0.1.8; bounded pages/details/calendar; NFR-13; no client account/org; Resend unchanged | 144-04* + 144-05* |
| 144-06B | Mailbox policy/domain/health admin service | `mailbox_admin_service.go`, focused service unit | owner organization, policy/domain constraints, redacted engine + projection state/lag and exact audit | 144-04I + 144-05C-Audit + 144-06A |
| 144-06D | Admin per-user alias desired-state service/handler | `mailbox_alias_admin_service.go`, `mailbox_alias_admin_handler.go`, focused units, narrow wiring | account-deleting guard returns stable 409 before receipt or I/O; otherwise no handler engine call; transient live-lease 423 remains only where already contracted and is not mixed with deleting; same-org account; pending/deleting alias generation; idempotent create; named audit | 144-04J/K/O + 144-06B/C |
| 144-06C | Separate mailbox admin handler and narrow wiring | `mailbox_admin_handler.go`, narrow wiring additions, focused handler unit | admin middleware + named audit; no flat Resend route/shape drift | 144-06B |
| 144-06E | Sensitive HTTP route metadata redaction | platform access logger, generic HTTP audit middleware, focused in-memory unit | matched route template only; unmatched constant; image token/search query/UUID absent; existing audit metadata preserved | 144-06A |
| 144-07 | Current-state docs sync and independent static review | routed domain docs + phase verification artifact only | implementation vs contract matrix; honest `NOT_RUN`; no protected QA paths | all accepted code packets |

## Frozen service semantics

- Engine ids are opaque and are always paired with the server-resolved account id. The
  144-03B Folder tree policy receives only projection/persistence-mapped Hub UUID rows, never
  an engine id; its UUID equality/order is raw 16-byte comparison.
- Raw MIME, recipients, subject and rich/plain draft/message bodies are engine state. Hub
  PostgreSQL stores safe projections, Hub relations, hashes and opaque versions only.
- `MailMessageFolder` and flag fields are rebuildable snapshots. Mutations call the engine
  and reconcile the returned/new change state; the projection does not authorize access.
  Every engine command/reconciliation owns one expiring account mutation lease and begin/
  finish epoch edges around I/O. Sync capture is forbidden during that lease. One leased
  sync page commits projection/full events/cursor atomically only if lease, prior cursor,
  epochs and mutation-quiescent state still match.
- Invalid cursor never clears rows. It exposes degraded/rebuilding state, resumes a
  base-anchored epoch-marked enumeration, catches up from its base cursor and only then deletes
  unseen projection rows/marks ready. First-seen event markers suppress rebuild storms.
- Draft mutation is serialized per draft. A lost response becomes `unknown`; a later
  operation first compares/retrieves engine state and never overwrites blindly. Save and
  reconciliation carry the same policy-bounded content budget and reject expanded results
  without committing them.
- Submission receipts persist only hashes of HTTP keys. `unknown` without an opaque result
  pair invokes `ReconcileSubmission`; indeterminate stays unknown. `unknown` with the pair
  retries only exact result hydration/finalization and can never requeue or call `Submit`.
- Domain/account provisioning is not a request-path side effect. Domain HTTP writes create
  desired intents; a domain worker reconciles Ensure/Get and public DNS status. The account
  worker reserves immutable address/key/cutoff, repeats idempotent `EnsureAccount` across
  lost responses, reconciles alias and user-lifecycle generations, and activates/suspends
  only by exact lease ownership. Public state never exposes engine/provider data.
- Rule HTTP writes are Hub-only desired-state transactions. A separate bounded worker owns
  the account command lease, reconciles one fixed script by generation/hash and never marks
  a newer desired generation deployed from an older result.
- `junk`/`not_junk` invoke explicit spam/ham classifier feedback. A move/flag result alone is
  neither trained nor audited as trained.
- Safe server content either passes the approved sanitizer and remote-image token rewrite or
  falls back to plain text. Search hydration indexes only bounded sanitized plain body and
  normalized filenames, then reauthorizes every hit before snippets.
- Mail create/bulk endpoints reserve a client-keyed hashed `MailMutationReceipt`; exact replay
  is 30 days and its compact key/request tombstone survives for the account/organization
  lifetime. Folder update/delete use an internal unexposed receipt only after lease +
  optimistic checks, and may remove it after the terminal recovery window.
- Hub uploads use the dedicated mail storage port. Existing project-bound `FileMetadata`
  is not adapted into a mailbox permission model.
- Every domain mutation that emits a durable event persists it through the current outbox
  transaction seam as the complete accepted 143-04L envelope. The existing publisher is
  called with `event_type == envelope.subject`; it and its worker are not changed or wrapped.
  Envelope event id is the consumer dedupe key; `Hub-Event-Id` remains transport-only.
  Direct post-commit publish is not an accepted substitute.
- Phase 143 first admits `calendar.>`; this phase then adds `mail.>`.
  `WebhookFilterSubjects` gains neither wildcard.
- Generic request logging is not proof of admin auditing. Each admin operation emits the
  exact `x-hub-audit-action` from OpenAPI without secret or raw content.

## Protected compatibility boundaries

- Existing alert files and behavior remain intact: `config.go`, `service.go`, `dto.go`,
  `admin_service.go`, `admin_handler.go`, `admin_broadcast.go`, `hub_handler.go`, and
  `/api/admin/mail/settings` are changed only by a separately reviewed narrow wiring need.
- `POST /api/admin/mail/broadcast` is not reused or expanded. Its removal/disablement is a
  Phase 148 founder decision.
- No file under `testframework/**`, legacy `backend/testframework/**`, phases 128–141,
  TASK-193–234 or `pro-qa-ai/**` may be read or changed by a packet.
- Genuine feature units live only under `backend/tests/mail/**`.

> **Local 01B clause reconciliation:** in this gate-status record only, prior 144-01B Docker-blocked Task3 wording is `CONSUMED/HISTORICAL`; its static source acceptance adds no executable or downstream authority.

## Gate status

| Gate | Status |
|---|---|
| Five-document default-vs-hard-cap/144-01A entry amendment | PASS at `5e314fb080d420696642058ae47397e251c23207`; historical 0.1.7 remains non-authoritative; this narrow PASS authorizes only 144-01A entry and does not accept implementation/runtime or broader Phase 142/cross-contract scope. |
| 144-02A schema readiness | `STATIC_CONTRACT_READY` docs-only acceptance at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931`; source schemas, generation and migration remain unimplemented/not authorized |
| 144-02B schema readiness | docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc`: normalized indexes, required engine version, bounded projection values and join/state checks are accepted as representation only; source/Ent/migration/runtime remain unimplemented and `BLOCKED` |
| 144-02C schema readiness | docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc`: stable draft identity, maximum-50 scalar staging ownership, split engine/Hub result identity, single scan authority and state checks are accepted as representation only; source/Ent/migration/runtime remain unimplemented and `BLOCKED` |
| 144-02D schema readiness | docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7d29d08c` plus transition `e563a3d3`; three post-transition exact reviews PASS C/I/M `0/0/0`; no source/Ent/runtime authority |
| 144-02F schema readiness | docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`: correction `fcb5b84f` plus exact-28 transition `2f4381a2`; Contract/Spec/Oracle, Security/Adversarial and Lifecycle/Dependency/Boundary Terra routes PASS C/I/M `0/0/0`; no source/Ent/runtime authority |
| OpenAPI response boundedness | ROOT_FROZEN_REVIEW_PENDING: count/item/string/cursor/aggregate/calendar limits encoded in 0.1.8 |
| Exact Terra implementation runtime | 144-01A is `STATIC_ACCEPTED` at `e3e8c774`; native dispatcher-enforced provenance with actual telemetry `UNKNOWN` is accepted unless exposed telemetry mismatches. 144-01B remains Docker-blocked by Task3 |
| Accepted Docker launcher | NOT_READY |
| Ent generation | NOT_RUN |
| Production migration mechanism | FOUNDER_DECISION_PENDING |
| Product code/tests/build | NOT_RUN |
| 144-03B Rule compiler contract | Docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `f199546513b758ca7b95328b5ccd7680009b732e`; 02F is accepted docs/readiness-only; Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN` |
> **Current Rule acceptance (2026-08-14):** Rule docs remain accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; 02D and 02F docs/readiness are accepted. The former pre-acceptance flow is `CONSUMED/HISTORICAL`. Rule Task 1 RED remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`; launcher `NOT_READY`, executable evidence `NOT_RUN`, Phase 148 blocked.
> **Rule-plan acceptance receipt (2026-08-14):** standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract high/Security xhigh/Lifecycle high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. All former 02F and Rule-plan authoring/review instructions are `CONSUMED/HISTORICAL`. No source, Ent schema, generation, migration, service, runtime, product, phase or release authority follows. Rule Task 1 RED remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`. 144-01B stays Docker-blocked; Phase 142/148 pending; launcher `NOT_READY`, executables `NOT_RUN`.

> **02B precedence (2026-08-15):** source `947e2a27650ff1c6db60af9ce13bb33b6bf19cb0` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; `IMPLEMENTATION-144-02B.md` records exact13 blobs. Tests first; Contract/high, Security/xhigh and Lifecycle/high precommit and exact-SHA PASS C/I/M `0/0/0`. All prior 02B `SOURCE_AUTHORITY_ACCEPTED`, source-unaccepted and 02C-waits-02B-source routing is `CONSUMED/HISTORICAL`. A separately authored/adjudicated narrow 02C docs/source-authority bridge is required; 02C source is not automatically authorized. Other blockers remain unchanged. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02C current route — candidate only:** `CONSUMED/HISTORICAL candidate lifecycle`. The former wait-for-accepted-02B-source clause is consumed/historical because source `947e2a27650ff1c6db60af9ce13bb33b6bf19cb0` and receipt `b93c77ae771c6509be54fca8817fbca7617d6853` are accepted. Before any exact13 dispatch: three independent precommit reviews (Contract/Spec/Oracle Terra/high; Security/Adversarial Terra/xhigh; Lifecycle/Dependency/Boundary Terra/high) -> one exact eight-document authority commit -> same three immutable exact-child-SHA reviews -> separate docs-only source-authority acceptance sync. This does not authorize source or source acceptance now, and does not open Ent generation, migration, services, API/UI, runtime, product, phase or release. 02F-D/02E, deletion-port and physical User/GDPR-intent blockers, launcher `NOT_READY` and executable/runtime `NOT_RUN` remain current. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**
