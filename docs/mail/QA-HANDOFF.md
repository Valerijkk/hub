# Hub Mail + Calendar Sync — QA handoff

> **144-02C source receipt (2026-08-20):** `STATIC_ACCEPTED / EXACT_SHA_PASS` applies only to `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`) and the exact 13 blobs in `IMPLEMENTATION-144-02C.md`. Tests/oracles first and all named reviews PASS C/I/M `0/0/0`. Former candidate/source-unaccepted/lifecycle-next/implied-next routes are `CONSUMED/HISTORICAL`; no next Mail code packet is selected or authorized, and root must choose separately. QA must not infer executable evidence: Ent generation, migration, direct SQL, services, API/UI, product, phase and release remain unauthorized; 02F-D/02E, deletion port, physical User/GDPR intent, launcher `NOT_READY` and executable evidence `NOT_RUN` remain current. Provenance Terra/medium, dispatcher-enforced/no fallback, telemetry `UNKNOWN`.

> **144-02B projection source-authority acceptance:** combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS`; Contract/high, Security/xhigh and Lifecycle/high corrected-child exact reviews PASS C/I/M `0/0/0`. Root may dispatch only this frozen test-first exact13: new `backend/ent/schema/mailfolder.go`, `backend/ent/schema/mailthread.go`, `backend/ent/schema/mailmessage.go`, `backend/ent/schema/mailmessagefolder.go`, `backend/ent/schema/maillabel.go`, `backend/ent/schema/mailmessagelabel.go`, `backend/tests/mail/mail_projection_schema_test.go`; edits `backend/ent/schema/mailaccount.go`, `backend/ent/schema/calendarevent.go`, `backend/ent/schema/mailmutationreceipt.go`, `backend/tests/mail/mail_account_identity_schema_test.go`, `backend/tests/mail/mail_mutation_receipt_schema_test.go`, `backend/tests/calendar/calendar_event_schema_test.go`.
>
> Static compatibility coverage must prove PERSISTENCE 02B tables/checks/indexes/JSON; only 02F-R exception is `MailMutationReceipt.folders` inverse for `MailFolder.control_receipt_id -> receipt RESTRICT`; actor/message scalar-only and other receipt edges/checks unchanged; `MailAccount` inverses folders/threads/messages/labels; `CalendarEvent.mail_messages` inverse with Message FK `SET NULL`; rule targets service-only, no DB FK/normalized target. Source implementation is unaccepted and still requires test-first authoring, independent review, atomic commit, exact-SHA review and status sync. Candidate/precommit/correction/failure-next and lower 02B eligibility/no-dispatch clauses are `CONSUMED/HISTORICAL`; immutable `45fd3369` failure remains history. former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; 02F-D/02E/Ent generation/migration/runtime/services/API/UI/release BLOCKED; launcher `NOT_READY`, executables `NOT_RUN`; Calendar accepted/no-next unchanged. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02D source receipt (2026-08-15):** `2242a4e7846944498eb5ed701ab8f47c986bd804` (parent `e9a1764b001e28074e4074a07ca8ac982e5345ba`, tree `8129303399e627ca5cedac3e70648955c74cfed9`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for blobs `55b9ccfb4e19a190fc7c083a61350ce7ba8c7e5b`, `6b8614c651ee70a70bdb46c00456362c8edad796`, `a826916965ab4875a9c61da83fa0a68a8d5d3d4d`, `fba258559fe6a8557b045fa10757330be29b5c08`, `7643892d0af8934892d23713e4d35b07b0bc455c`, `b2216313dfdf492315a377b3567fdc4f71d05a0`, user `b595855cbe9cd39f522ef8263eb72ea4db4f9b55`, A3 compatibility test `74599da63c180942e092a909ee53eb0ca6b7e51f`, and authored-first 02D oracle `b94c686c35e3a5cca2d7838424f7e255012437d8`. Tests first; Terra/xhigh author and Terra/high precommit+exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`; launcher `NOT_READY`, executable/generation/migration/runtime evidence `NOT_RUN`. Former 02D-next/unimplemented/source-blocked clauses are `CONSUMED/HISTORICAL`; old docs receipts remain. Former 02B-waits-for-02F-R, 02F-R-root-dispatch and 02B-separate-authority/02C-predecessor clauses are `CONSUMED/HISTORICAL`; 02F-R is now `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`; 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation is unaccepted, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; other 02F incl 02F-D unauthorized. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-03C-Attachment static acceptance (2026-08-15):** `7b2ea5721684e43559d6f6b4b5fc462d8daafb8b` (parent `4e7c609713983a8f6db6fbfeb7474abd73451e8c`, tree `4c3fe4d27e170c5d3f42b116c39887036ce936c6`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for deterministic attachment policy: source blob `3f67ce1b9f2f056b70bfd3070f1c34f98b5c71c1`, test blob `bf7eda125d5aba1d0cc2594746d14e7bcfd06aa7`; precommit+exact-SHA reviews PASS C/I/M `0/0/0`. Terra/xhigh authoring, dispatcher-enforced/no fallback, telemetry `UNKNOWN`; tests were authored first. The static oracle covers per-attachment count/size, NFC output/NFKC structural+extension and closed MIME/type defense, inline CID, SHA/scan gates and clean-only send/read. RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime are `NOT_RUN`, launcher `NOT_READY`. No scanner/storage/network/project/tenant auth/schema/migration/service/API/UI/product/phase/release authority. Existing 144-01B/02/Rule/Sieve/Phase145 blockers and Calendar 143-03E launcher+migration blocker unchanged. Its former no-further-source/root-only-selection routing is `CONSUMED/HISTORICAL`. Founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker; every exact packet still requires its own test-first/static review/commit. No Ent generation/migration/runtime authority follows.

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` remain accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; unrelated gates remain current. No source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Base/correction lifecycle `CONSUMED/HISTORICAL`; no plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness keep Rule Task 1/145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **144-02F corrective static oracle:** inspect literal NULL-negative matrices for every receipt/deployment named check and all three full hash-v1 preimage hex/digest vectors in the persistence contract (empty 22 bytes, one 77, two 178). Verify separately folder p10 no-I/O terminal 409/30d replay and `message_bulk p10` manual-only no schedule/reclaim/I/O; same-key replay returns its persisted 202 snapshot, while fresh-key same/different targets repeatedly receive 423/no row/no expansion/no I/O under the account lock, including late-owner contention. Also verify the 16-byte header plus `0x00||0x01`, target uniques `(receipt_id,request_index)`/`(receipt_id,target_type,target_id)`, and account-first CAS/lease fencing. They are independently computed V8 pure-JS/.NET SHA256 future oracles, not test/runtime evidence. Correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle is `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executable evidence `NOT_RUN`.

> **144-02F-R QA source authority:** `DISPATCH-144-02F-RECEIPT.md` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` at `a877848e`; Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh, and Lifecycle/Dependency/Boundary Terra/high reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. It authorizes root dispatch of only the exact eight-path, test-first static receipt-schema packet and no execution or source implementation acceptance. Former candidate/precommit/commit/review-next, 02F-R-root-dispatch and predecessor `02B requires separate docs authority` clauses are `CONSUMED/HISTORICAL`; 02F-R is now `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`; current authority is 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 only, implementation unaccepted, and former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; other 02F incl 02F-D unauthorized. Launcher stays `NOT_READY`; executables, generation, migration, and runtime stay `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02D/02F current:** the 02D base remains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus `e563a3d3be8898371e5c8a9861c58f11b72167bc`, while affected User/RecentContact lifecycle clauses are reconciled by the accepted exact14 docs representation at `576f19231e37f635d5b7187a1ada121ed14ff8dc`. 02F `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` remains docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and Rule-plan-next instructions are `CONSUMED/HISTORICAL`. No source/schema/runtime authority follows; accepted Rule plan `669ef761` leaves Rule Task 1 `NOT_AUTHORIZED/BLOCKED` on independent Phase-145 carrier/G145-5B/G145-5C/launcher readiness. Runtime remains `NOT_RUN`.
> **Contract version:** 0.1.8
> **Prepared:** 2026-08-13
> **Contract review status:** five-document 0.1.8 entry amendment PASS at `5e314fb080d420696642058ae47397e251c23207`; Thread contract remains accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461`; implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`, reviews C/I/M `0/0/0`), static only; broader Phase 142/cross-contract review remains pending
> **Implementation status:** IN_PROGRESS — 144-02A1, 144-03C-Send and 144-03C-Attachment are `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, static only. The former no-further-source/root-only-selection routing is `CONSUMED/HISTORICAL`. Founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker; each exact packet still requires its own test-first/static review/commit. No accepted packet authorizes scanner/storage/network/project auth, Ent generation/migration, or runtime. 144-01B/02/Rule/Sieve/Phase145 blockers and G145-5C remain open; launcher `NOT_READY`, executable verification `NOT_RUN`.
> **Runtime status:** NOT_RUN
> **QA owner:** founder-designated Claude test program
> **Product-code owner:** ChatGPT mail feature branch

> **144-02A1 DNS static handoff (2026-08-15):** only `backend/internal/mail/domain_dns.go` blob `e00586637f209f85c52626c1ec8d96116af934b3` and authored-first `backend/tests/mail/domain_dns_test.go` blob `8a43140964d12cca7cdc09e34675a074ee164d7b` are `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `5892a90b7582ff4ac35db218428400ac9608c4e3` (parent `29a5c624c0bacc2de3cc4b5a93e3e9176320b5bd`, tree `c78d82ef8ea55a4656d33b82530b2fbb72582201`). Terra/high static precommit and immutable exact-SHA review PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN` unless exposed. RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime remain `NOT_RUN`; launcher `NOT_READY`. This is no executable QA, Ent/schema generation/migration/service/API/UI/product/phase/release acceptance. In combined A1/A2/A3 tracking, all three are `STATIC_ACCEPTED`; A2-unimplemented/A3-next clauses are `CONSUMED/HISTORICAL`, and A3 is accepted at `e9a1764b001e28074e4074a07ca8ac982e5345ba`. Founder root-only source sequencing before final integrated Docker does not authorize Ent generation/migrations/runtime; all existing blockers remain.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

> **144-02A/B/C/D 2026-08-15 acceptance:** canonical receipt `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) accepts the exact 14 content documents as docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high exact-SHA reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Candidate/precommit/root-commit/exact-review/status-sync-next clauses are `CONSUMED/HISTORICAL`. 02F is unchanged except dependency trace. Launcher `NOT_READY`; executable checks `NOT_RUN`; no source/schema/generation/migration/service/test/runtime authority. Engine deletion-port and User deletion-intent remain separate blockers.

> **144-01B static acceptance (2026-08-15):** immutable `0c064363b5ccb657a7c444802e94075c4ca743e5` (parent `48a7273d100c07788def88e2f1b8d69c7173580f`, tree `9c568a0c58c6bf7b8c3171cfd934760b839db623`, subject `feat(mail): add deterministic engine fake`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for fake `12bcd921668e406f07a98f7f3b0618ad8d44e81d` and authored-first test `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f`. Terra/xhigh author; Spec Terra/high, Security Terra/xhigh, ABI Terra/high precommit and exact-SHA Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 01B Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executables `NOT_RUN`; no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority.

> **144-02A2 static handoff (2026-08-15):** `65c91b4824e8a1948fc97c149305aa7f009de061` (parent `5df6b47678364c438ade579252209c8fd238c43c`, tree `08583554939f597443566093f30b6ec7aa9373db`) accepts blobs `4de2ec0882d94db595aaabf8d4117566118ee472`, `a70e05122ffa126e9754d029ca00493faae2aeef`, `ef6e2e65430a1397b0f563f1f361ecd89660a39c`, authored-first test `65de04719a60a38d7fc456553e092bb106456957` as `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Terra/xhigh author, Terra/high precommit+exact-SHA PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Launcher `NOT_READY`, executable QA `NOT_RUN`; no generation/migration/runtime/service/API/UI/product/phase/release authority. A2-unimplemented/A3-next, 02B-waits-for-02F-R, 02F-R-root-dispatch and predecessor `02B requires separate docs authority` clauses are `CONSUMED/HISTORICAL`; 02F-R is now `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`; current authority is 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 only, implementation unaccepted, and former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; other 02F incl 02F-D unauthorized. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02A3 static handoff (2026-08-15):** immutable `e9a1764b001e28074e4074a07ca8ac982e5345ba` (parent `47f8987b269e1625022071e0de95a76da0d2f00a`, tree `c0b59b2ba1168076a4cc787cacb99fe5081cd301`, subject `feat(mail): add account and identity schemas`) accepts exactly seven paths/blobs: account `30e68ff52171e643e1361f3d6acfd2310289d8e1`, domain `16f5d20600d1359ba737417428d594ac73468360`, identity `1b06e04c9235e5c891a601261a2e400068155065`, organization `df85d2e2f61bcc8b62235f0557f774fb06eca1e3`, user `5594ca1c68b8f3c726ab90a77dcc26b7e7713e0c`, authored-first A3 test `0939618d620c7f9b503df50e4aff04fbf0b3a85d`, and A2 compatibility test `70b0c56f20043aeb8b4b719ca4ae6f3405cb5221`, as `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. Terra/xhigh author; Terra/high precommit+exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Launcher `NOT_READY`; all executable/generation evidence is `NOT_RUN`. No Ent generation/migration/runtime/service/API/UI/product/phase/release authority. A3-unimplemented/A3-next, 02B-waits-for-02F-R, 02F-R-root-dispatch and predecessor `02B requires separate docs authority` clauses are `CONSUMED/HISTORICAL`; 02F-R is now `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`; current authority is 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 only, implementation unaccepted, and former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; other 02F incl 02F-D unauthorized. A1/A2 facts and all blockers remain unchanged. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02F-R static receipt:** `ce542cb5de839a1fb5200295f710b0ae45bcdbe6` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact eight blobs and static-review evidence are in `.planning/phases/144-mailbox-domain/IMPLEMENTATION-144-02F-R.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Not executable QA: launcher `NOT_READY`, tests/build/generation/migration/runtime `NOT_RUN`.

## Boundary

This handoff gives the test program stable inputs without modifying its protected zone.
The mail feature branch must not add or change anything under `testframework/**`,
`.planning/audits/test-rebase/**`, phases 128–141, or TASK-193–234.

Feature-owned genuine units are permitted only in:

- `backend/tests/mail/**`
- `backend/tests/calendar/**`
- `frontend/tests/mail/**`

No new test belongs in legacy `backend/testframework/**`.

## Contract artifact map

| Artifact | Path | Purpose |
|---|---|---|
| Architecture decision | `.planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md` | engine/source-of-truth/scope/security decisions |
| Domain + traceability spec | `.planning/decisions/specs/MAIL-SPEC.md` | ERD, state machines, requirements and named scenarios |
| Machine-readable HTTP API | `docs/mail/openapi.yaml` | all v1 public/admin endpoints and schemas |
| Bounded engine amendment | `docs/mail/ENGINE-CONTRACT-0.1.8.md` | projection/content separation, byte budgets and stable draft identity |
| Exact persistence amendment | `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` | Phase 144 schema fields, checks, indexes and state matrices |
| Control-plane schema execution plan | `docs/superpowers/plans/2026-08-13-mail-control-plane-schemas.md` | Typed DNS then domain/policy then account/identity external-oracle packets |
| Machine-readable events | `docs/mail/events-v1.schema.json` | NATS envelope and payload union |
| Reliability addendum | `.planning/phases/147-mail-calendar-reliability/CONTRACT.md` | inbound ordering, reply recovery, availability, reminders, CRM/L1 seams |
| Product/UI design | `docs/superpowers/specs/2026-08-13-mail-module-design.md` | three-pane, composer and RSVP interaction |
| Founder decision queue | `docs/mail/FOUNDER-QUESTIONS.md` | defaults that do not authorize external action |

## Required test layers after accepted launcher availability

| Layer | Primary coverage | Dependencies |
|---|---|---|
| Contract | OpenAPI schema/examples, event schemas, source rules, authorization shape | no SUT runtime where possible |
| API | folder/label/thread/draft/send/event/RSVP/admin behavior | backend + fake or accepted engine profile |
| Integration | PostgreSQL constraints/outbox, JMAP adapter, NATS consumers, Meili, SeaweedFS, SMTP/iTIP | task-owned Docker profile |
| E2E smoke | internal send → inbox → invite RSVP → organizer realtime | full bounded stack |
| E2E full | external-domain flows, bulk, search, recurrence, TZ/DST, error recovery | full bounded stack + deterministic fixtures |
| Security | mailbox isolation, HTML sanitization, SSRF, attachment policy, spoofed iTIP, rate/loop | security fixture services |
| Load | list/search/unread/submit/RSVP bounded profiles | approved Locust profile |

The test program decides exact Python cases and identifiers, but must map each case to one
or more `FR-MAIL-*` / `FR-CALSYNC-*` rows and named `TC-*` scenarios in `MAIL-SPEC.md`.

## P0 scenario handoff

### Mail

1. Owner isolation prevents cross-user message/folder/attachment/engine-id access.
2. Internal delivery produces exactly one Sent and one recipient Inbox view through the
   engine path and emits idempotent NATS events.
3. Retrying a timed-out submission reconciles `Idempotency-Key` without duplicate mail.
4. Draft update, delete and new-key send preserve the exact last-confirmed `If-Match`.
   Stale/reconciliation/not-editable/submission-in-progress returns its named conflict and
   never overwrites, deletes or submits newer content; same-key send receipt replay wins
   before a now-stale version check.
5. Pending/rejected attachment prevents submission; clean attachment preserves hash/name.
6. Active HTML, script handlers, unsafe SVG, CSS URL fetches, and tracking images do not
   execute or load directly.
7. Image proxy blocks private/loopback/link-local/metadata/rebinding targets before and
   after connection.
8. Search never leaks another mailbox through hit count, ids, snippets, or timing-oriented
   hydration differences in normal API semantics.
9. A lost initial draft-create/save response reconciles the immutable engine draft key and
   opaque engine version before another write. Save and reconciliation use the same positive
   caller content budget; an adapter result one byte above it is rejected as a whole and does
   not replace canonical state. PostgreSQL contains no recipient, subject, rich/plain body,
   or MIME copy.
10. Mail uploads use an owner/account-bound `MailObject` namespace and cannot inherit
    authorization from project-scoped file metadata.
11. `GET /api/mail/submissions/{id}` and a same-key send retry converge from `unknown` to
    the engine-confirmed result; a still-indeterminate result never triggers blind submit.
12. A draft is immutable while submission is queued/submitting/unknown; confirmation
    marks it submitted, while definitive non-delivery restores the same content for edit.
13. One projection-only incremental page commits folder/message/thread projection, complete
    envelopes and the exact next composite cursor atomically. Rollback or any stale lease,
    prior cursor, sync epoch, mutation epoch or present command lease advances none of them.
14. Every synchronous engine-authoritative folder/message/draft/submission command or reconciliation
    acquires one expiring account mutation lease and advances the begin epoch before I/O;
    completion/unknown handoff clears the exact lease and advances the end epoch. Sync
    capture is forbidden while the lease is present. Active-lease contention makes zero
    engine call; sync claim or a later command can reclaim expiry only with an epoch bump,
    invalidating late work whose former owner must reconcile rather than finalize directly.
    The asynchronous rule-deployment worker uses the same lease/epoch fence; HTTP rule CRUD
    itself changes Hub desired state only and therefore never returns account-busy.
15. Invalid/expired cursor preserves current owner rows as visibly degraded, resumes a
    bounded base-state/ID-anchored enumeration without destructive pre-clear, catches up
    from its base cursor and only then tombstones unseen rows and marks ready. Query-state
    drift must converge through catch-up; missing anchor/expired base history restarts.
    Crash/restart in every enumeration/catch-up phase resumes without guessing.
16. Enumeration history at or before account event cutoff emits no inbound notification. A
    first-seen post-cutoff inbound-or-unknown message reserves one stable marker and emits
    one `mail.received.v1` even across page replay/rebuild. Only a stable mapped Hub
    submission/draft suppresses it; spoofed owner `From`, alias equality or mutable Sent
    membership never does.
    The cutoff is committed once before worker eligibility and comparison uses engine
    receipt time, not the message `Date` header or local sync observation time.
17. Two backend replicas contend through one expiring account lease. Scan/page/concurrency,
    timeout and backoff caps hold; disabled configuration starts no worker/timer/probe, and
    root-context cancellation drains in-flight work without leaked goroutines. Command I/O
    deadline is shorter than its 60-second lease, and begin/during/end sync races are covered.
18. Folder/thread/search responses always carry `projection_state` and required nullable
    `projection_as_of`. UI labels bootstrapping/degraded/rebuilding data stale or in progress,
    derives mixed pages conservatively and never fabricates ready after a query-only retry.
    Search is ready only when both mailbox projection and owner search checkpoint are ready;
    its freshness is the older complete instant.
19. The 101st non-deleted engine-backed folder is atomically rejected with 422 and
    `MAIL_FOLDER_LIMIT_EXCEEDED`. An engine account already over cap leaves the last owner
    projection degraded; no truncated folder tree is marked ready.
19a. Folder create first inserts one unique `creating` row; PATCH/DELETE use `If-Match` and
     an internal injected unexposed UUID v4/v7 receipt to persist desired state/version before engine I/O.
     Busy/stale/protected requests create neither receipt nor control change. Lost response returns/retains safe 202
     state, prevents a second key/name or folder mutation, and the existing projection worker
     recovers the oldest due receipt without a fifth worker. Exact non-system match/
     definitive absence may finalize/retry once only when the selected profile proves Hub
     is the exclusive custom-folder writer. False/unknown capability, an exposed alternate
     writer, a system-role candidate or ambiguous state makes zero recovery mutation, stays
     degraded and never duplicates/overwrites.
     Ordinary sync/rebuild preserves that intent and cannot mark the account ready; internal
     receipts are never accepted from an HTTP header or kept as lifetime public tombstones.
20. Concurrent account command/reconciliation lease contention returns 423 plus
    `MAIL_ACCOUNT_BUSY` and bounded `Retry-After` before any engine call. Unknown engine
    outcome never maps to busy, and the UI never blindly replays it as though no call ran.
21. Domain create first commits only one stable desired intent; its worker reconciles
    DNS-required/delete state through Ensure/Get without changing DNS or exposing private
    DKIM. Domain existence, `isEnabled` and expected `dnsZoneFile` never produce verified;
    only the future G145-5C authority may promote it. With engine/policy/default verified domain enabled, two account workers
    scan the same active new/existing user and reserve exactly one deterministic primary address, provisioning key
    and cutoff. A lost ensure response repeats the same key and activates one logical engine
    account/address set; separately authenticated `Identity/get|set` returns exactly one
    stable opaque id per desired address and reconciles a lost create by Get-before-retry.
    Missing domain, disabled engine, suspended user and stale lease make zero engine call.
    `GET /mail/account` is read-only and leaks no reservation, engine id,
    lease or provider text. Alias create/delete and user deactivation/reactivation advance
    account generation; pending/deleting identities are never selectable, and stale workers
    cannot finalize an older set. A claimed 12-hex UUID fallback expands once to the full
    UUID; its namespace cannot be claimed by aliases, and no collision path uses randomness
    or calls the engine before a unique reservation commits.
22. Folder/label/signature/rule/presign/admin-domain/alias create and bulk message actions
    require one stable client-generated UUID v4/v7 `Idempotency-Key`. Same key/input reconstructs the original owned
    result; changed input returns `IDEMPOTENCY_KEY_REUSED`; a removed target returns
    `IDEMPOTENCY_REPLAY_UNAVAILABLE`; compacted metadata returns
    `IDEMPOTENCY_REPLAY_EXPIRED`. Unknown engine outcome reconciles before retry and none of
    those 409 states executes again. Presign replay mints a fresh URL only for the same
    pending/unexpired attachment id without extending the object lease, and URL expiry never
    crosses it; confirmed/expired/orphaned/deleted state is unavailable and never allocates
    another object. Raw keys and response/content data never persist.
    Key digest is scoped/length-delimited SHA-256; request fingerprint is raw-keyed
    HMAC-SHA-256 over fingerprint-v1 tagged/length-delimited normalized typed fields with
    request array order preserved and no maps/unknown fields. Invalid keys are rejected
    before reservation, and small payloads cannot be enumerated from the stored fingerprint.
    The 30-day clock begins at first terminal commit; in-progress/unknown never compact and
    remain safely reconcilable/observable rather than silently becoming expired.
23. Rich draft/signature input is sanitized on the server and its plain form is derived
    there; clients cannot supply two independent canonical bodies. Without the approved
    sanitizer, rich mode fails closed to plain-only. Received hostile HTML cannot retain
    active/fetch primitives. Remote images remain blocked until explicit action/trusted
    policy and use only owner/message/expiry-bound tokens; every redirect/connect is
    SSRF-validated and origin URLs never reach DOM, API output, search, logs or errors. A
    successful proxy response is bytes with exact private five-minute cache, `nosniff`,
    `no-referrer` and restrictive CSP headers; it never redirects to an origin/cache URL.
24. Rule CRUD commits only the complete Hub desired generation/hash and returns deployment
    `pending|applying|ready|unknown|error`. The bounded worker owns one fixed script, compares
    generation/hash through Get-before-retry after lost response, cannot finalize a stale
    generation, detects/repairs external drift, and stores/logs no raw Sieve.
    `add_label` compiles only an opaque reserved engine keyword; sync maps it to a same-
    account provenance join. Unknown markers never become labels and visible removal clears
    an engine-rule marker through the fenced keyword command before clearing its source.
    The receipt-backed bulk action is the only public message-label mutation path, including
    a single message; no weaker replace-label endpoint exists.
25. Bulk selection sends exact `target_type=message|thread`; thread targets are expanded and
    deduplicated server-side under the 200-message cap. The receipt commits ordered request
    targets and exact member UUID snapshot before engine I/O, then normalized safe outcome;
    replay never re-expands changed threads. Per-target results preserve the original type.
    `junk`/`not_junk` separately report spam/ham training accepted/rejected/unknown; a folder
    move alone never appears as trained. Without proven engine idempotency/reconciliation,
    indeterminate training is terminal unknown and receipt replay cannot double-train.
26. Search finds bounded From/To/subject, server-sanitized plain body and normalized
    attachment filenames. Index documents exclude raw HTML/MIME/BCC/remote URLs/bytes;
    consumer replay/update/delete is idempotent, checkpoints drive freshness, and each hit
    is reauthorized after owner filters before id/snippet return.
27. Folder/label managers, rules deployment status, blocked-image action, provisioning
    status and admin aliases have loading/error/empty/conflict/unknown states, keyboard/focus
    coverage and stable create keys. Folder controls preserve version and render creating/
    unknown/deleting/error without claiming completion. Primary/group identities cannot be
    mutated from alias UI.
28. Settings, signature, label, rule, mailbox-policy and domain edits preserve the last
     confirmed quoted `If-Match` across one auth retry. Stale versions mutate nothing and
     refetch before user retry; signature/rule/label deletes are versioned too. Deleting a
     label referenced by an active rule returns `MAIL_LABEL_IN_USE` without removing joins,
     markers or rules. Every new mailbox 409 has its exact OpenAPI `X-Hub-Error-Code`;
     label name, policy prerequisite, domain verified/default/in-use and alias state are
     distinguishable. Cross-tenant domain/address collisions expose only `*_UNAVAILABLE`.
29. Attachment confirm with changed hash/size or non-confirmable state returns the named
    409 and schedules no scan; expired pending returns named 410. Non-clean download is 409.
    Clean download redirects only to the fixed configured HTTPS storage origin with
    required `no-store`/`no-referrer`; client/provider origin input is impossible.
30. New mailbox request objects reject unknown properties before service invocation.
     Forged organization/account/engine fields and invalid rule condition/action unions make
     zero repository, engine, classifier, storage or audit-success mutation.
    Draft recipient inputs also reject response-only `user_id`; Hub-user enrichment is
    resolved server-side from the normalized address and never grants send authorization.
    Bulk move/label/pin/no-argument variants reject every missing or irrelevant target field.
31. Application access log and generic HTTP audit record the matched route template, never
    concrete remote-image token, UUID path or search query; unmatched uses one constant.
    Proxy evidence separately proves token-path and query redaction. OTel/metrics labels
    remain template-only, and named domain audit keeps only its explicit bounded target id.
31a. Remote-image path token accepts only unpadded base64url of 32..1024 ASCII characters.
     Padding, slash, percent-escape, Unicode, punctuation and boundary overflow fail before
     token decode, repository lookup, authorization or outbound network work.
32. Submission, rule-deployment, domain-control and bulk per-target `error_code` values
    reject every string outside their OpenAPI allowlists; raw SMTP/JMAP/provider text cannot
    enter API JSON, persisted safe code, audit or client analytics.
    Submission state also enforces its nullable shape: only `sent` has `message_id`, only
    `unknown|failed` have their state-appropriate non-null safe code.
33. Every stored `MailMessage` has a non-empty bounded `engine_version` from the same engine
    projection as `engine_message_id`; optional auxiliary engine state cannot substitute for
    it. A same-id newer-version page updates both projection content and version atomically.
34. Submission success stores the opaque engine submission/message ids separately from the
    Hub UUID returned as public `message_id`. The engine message id is never parseable as a
    UUID oracle. `sent` is committed only after a bounded same-account result read and atomic
    `MailMessage` projection upsert; failure after delivery confirmation remains
    reconciliation-required. With no opaque result pair retry calls submission reconciliation;
    with a complete pair it skips reconciliation and retries only result hydration/finalization,
    never requeue or a second Submit.
35. Draft create/save accepts at most 50 distinct attachment ids. The 51st id and a direct
    repository attempt to create a 51st draft association both fail before engine I/O and
    leave the previous complete attachment set unchanged; an over-limit persisted state is
    fail-closed at submission.
36. Result hydration accepts only exact requested account, exact confirmed engine message id
    and non-empty bounded version. A mismatched account/id or spoofed adapter origin stores no
    Hub binding, returns no `sent`, publishes no outbox and never resubmits. The domain derives
    `confirmed_owner_submission` only from the persisted confirmed result pair.

### Calendar synchronization

> **143-04C candidate precedence:** the earlier no-next Calendar packet wording is
> `CONSUMED/HISTORICAL` only for the separately reviewed 143-03B-T → 143-04C sequence; this
> remains docs-only `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED`.

1. Legacy `GET /api/calendar/events` keeps `{events,from,to}` and all existing kinds;
   canonical CRUD/RSVP lives under `/api/calendar/v2/**`.
2. REQUEST create/update, CANCEL, and REPLY round-trip UID/SEQUENCE/METHOD correctly.
3. Whole-series RSVP and occurrence RSVP with `RECURRENCE-ID` remain distinct.
4. Time/location/recurrence changes reset non-organizer RSVP; silent description edit
   preserves it.
5. Concurrent stale RSVP returns canonical `{error,request_id}`, 409,
   `X-Hub-Error-Code: ATTENDEE_VERSION_CONFLICT`, current attendee-version `ETag`, and
   canonical event `Location`; the client refetches and reconciles before retry.
6. Duplicate equal UID/sequence/payload for REQUEST/CANCEL is idempotent; conflicting
   equal sequence is quarantined; lower sequence cannot revert state.
7. For an existing master with a materialized exception, PATCH RRULE, EXDATE, DTSTART/start-date,
   timezone or all-day fails with NFR-13 body and
   `X-Hub-Error-Code: CALENDAR_EVENT_RECURRENCE_CONFLICT`, typed server conflict and zero writes;
   end/duration-only patch remains eligible. Verify only the safe update/cancel EnvelopeV1
   notification-intent descriptor, never raw iTIP/body, queued/delivered transport, or a delivery
   claim; a separately accepted Mail/iTIP hydration-delivery packet may later reconstruct transport.
8. Update tests must prove resolver non-disclosure precedes terminal/source/If-Match checks;
   missing optional update backend causes `ErrEventInvalid` with no backend/receipt call; escaped
   transaction and caller/backend alias mutation fail to affect state; every event/version/sequence
   or missing/extra/duplicate/ID/version/removed attendee-result echo mismatch and any write failure
   rolls the entire event/attendee/outbox unit back. Test inactive reader
   exclusion, fresh re-invite ID, and rejected same normalized-email role/resolved-identity change
   through the 04C update service, not 03B-T schema metadata. The absent optional backend oracle
   proves `ErrEventInvalid` and zero backend calls including `ResolveActorOrganization`.
9. Constructor oracles cover deep-copy, zero absent wrapper rejection and present-null versus
   absent for temporal/RRULE/reminder; test timed/all-day transitions and merged final temporal
   validation. Cover 200 attendees accepted/201 rejected, exact update/cancel envelope JSON,
   canonical changed_fields order/duplicate rejection, persistence-result mismatch rollback and
   no-op zero persistence/outbox calls.
   The duplicate key is CVE-v1, not serialized/folded wire bytes: property order, folding,
   parameter case/quoting, attendee/EXDATE order and RRULE component order normalize to
   equal hashes. Mutating each included semantic field changes the hash. Stored event,
   exception, object-state and inbound-REPLY hashes include canonical version `1`; another
   version is never compared as equal. Malformed/unknown TEXT or caret escapes, decoded
   controls other than the explicitly allowed LF, and multibyte strings exceeding the
   post-normalization UTF-8 byte bounds reject before hashing.
7. Reused HTTP idempotency key with identical input replays the original response; changed
   input conflicts and never emits a second mutation/outbox event.
8. REPLY changes only the normalized envelope/authenticated sender's attendee row. It
   accepts only the current event sequence and persists a bounded durable receipt for
   every distinct validated Mail 0.1.8 domain-separated SHA-256 over canonical version,
   length-delimited normalized sender and canonical payload hash. Exact
   redelivery is a no-op even after a later PARTSTAT; future/spoofed and the ninth distinct
   current-sequence reply are quarantined; no HTTP idempotency receipt is used. A distinct
   current-sequence PARTSTAT may advance that attendee once with one receipt/outbox, while
   a distinct same-status hash writes only an `observed_no_change` receipt.
9. External guest remains email-only and cannot gain Hub authorization.
10. Timed events require `starts_at/ends_at`, null date fields and `end > start`; all-day
    events require date/exclusive-end, null instants and no midnight conversion.
    Inbound all-day `VALUE=DATE` uses the fixed CVE-v1 `UTC` semantic marker because it has
    no trustworthy TZID. An occurrence REQUEST is accepted and hashed inbound even though
    the outbound 143-02A serializer does not emit that shape.
11. Timed events retain instants across organizer/viewer timezones and DST transitions.
12. Busy/free combines canonical events, scheduled calls, and approved HR leave. It
    authorizes the complete requested user set before source queries: self or active same-
    organization users sharing a non-deleted project/active channel, without admin bypass,
    returns only `{from,to,source}` busy data plus free complements, and never leaks title,
    ids, leave reason, project or participants. Missing/inaccessible users share one `404`;
    expansion cap returns `422` + `AVAILABILITY_LIMIT_EXCEEDED` without partial users.
    Approved leave stays viewer/shared-project scoped, expands in the employee timezone,
    and invalid legacy timezone returns `422` + `AVAILABILITY_SOURCE_INVALID`.
13. Cancelling one recurrence instance increments master sequence/version, materializes a
    cancelled exception and emits only occurrence-scoped iTIP/outbox; the series remains.
14. For both series and occurrence attendees, deliver `declined(hash A)`, then
    `accepted(hash B)`, then redeliver A at the same event sequence: the final RSVP remains
    accepted and the redelivery creates no attendee/outbox mutation. Repeat with a
    same-status distinct hash and prove its receipt prevents a later rollback; the ninth
    distinct hash is quarantined without mutation.
15. A trusted CANCEL received before REQUEST creates an owner/UID/occurrence ordering
    tombstone. A delayed lower or equal REQUEST is stale/conflicting and cannot resurrect
    it; a strictly newer authenticated REQUEST follows the frozen transition table only
    for the same organizer. Another authenticated organizer cannot take over the UID with
    a higher sequence.
16. An occurrence REQUEST received before its master is durably deferred without calendar
    bytes in PostgreSQL. Master arrival rehydrates and rechecks it; a newer occurrence
    tombstone wins. A different organizer cannot attach an occurrence to a known UID
    series. Whole-series CANCEL makes all materialized exceptions cancelled, and a later
    occurrence-only REQUEST stays deferred until a same-organizer active master exists.
17. An external-organizer REPLY uses one immutable delivery, hidden draft intent and
    `MailSubmission`; a lost result enters unknown and reconciles without a second submit.
    Public delivery status stays pending for pending/submitting/unknown, becomes sent only
    after definitive success, and becomes failed only after permanent failure.
18. Reminder lookahead creates/finds pending work; only the due-time recheck emits or
    suppresses it. Redelivery/concurrency yields one receipt and one reminder event. When
    policy permits, exactly one in-app row exists; preferences/quiet hours may instead
    produce zero. No Resend/email fallback is called. Telegram/push are asserted only as
    existing best-effort channels, invoked after canonical commit; rollback invokes none,
    and their result cannot change the committed receipt/notification outcome.
    A transient user/event/preferences/repository failure leaves the in-app outcome
    pending and NAKs for bounded retry; it is not recorded as suppression. A failed
    post-commit WebSocket publish leaves the durable row intact and is repaired by normal
    notification refetch without an exactly-once realtime claim.
    The due claim and consumer finalization each verify current sequence, effective
    occurrence start, non-null `reminder_minutes`, recomputed `due_at`, active status and
    owner/internal-attendee eligibility. Event, optional linked call, optional exception,
    optional attendee, stable user-policy guard, receipt and notification use the frozen
    lock order; preference upsert locks that
    user row before resolving an optional override. In-app channel/reminder disable,
    quiet-hours/offset change, reschedule/cancel/delete or decline/remove committed before
    finalization produces one durable suppression and no in-app row. The opposite
    transaction order preserves only the already committed valid row; Telegram/Web Push
    remain best-effort. Timed due uses the effective instant. All-day due uses the first
    representable instant of local `start_date` in the event's IANA timezone without
    changing date-only storage: midnight overlap picks the earlier instant, midnight gap
    picks the first instant after the gap, and a skipped date picks the first instant after
    that date's gap.
    Both-absent quiet-hour fields mean disabled; an incomplete/malformed pair or invalid
    IANA timezone creates one durable `policy_invalid` suppression, no notification row and
    a bounded content-free signal. It never falls back to UTC or delivery.
    Assert every receipt-state tuple: pending has no outbox/notification; emitted requires
    outbox; delivered requires notification and no reason; suppressed has no notification
    and one closed safe reason; scheduler-expired uses reason `expired`. No terminal tuple
    reopens.
19. A flood of authenticated unique unknown CANCELs cannot grow unbound state past 10,000
    per mailbox; excess input quarantines, CANCEL tombstones remain for mailbox/account
    lifetime, deferred work expires/dead-letters after seven days, and cleanup never removes
    pending/retained-event state or evicts a tombstone to admit a new UID. At cap, known UID
    updates still converge, every previously unseen UID state-changing part mutates no
    event, health is visibly degraded, and the content-free ops signal is rate-limited.
20. Backfill maps every non-deleted call, including paused/completed rows, to the project
    organization, exact-channel creator, title, description,
    instant, duration, reminder and distinct active same-project/same-organization
    participants exactly; preserves source
    audit timestamps; and emits no historical outbox/notification/mail. Any orphan,
    inactive, organization mismatch, invalid email/timezone/interval/reminder or target-bound
    overflow blocks the entire apply. Soft-deleted counts are recorded and cannot be
    reactivated; cutover requires 100% linkage of the non-deleted set.
21. Every non-deleted recurring source row compares the current raw-worker, injected-UTC
    aggregate and proposed canonical ordered occurrence sets with the pinned recurrence
    library. Parse/subset/cap/non-equality or an unanswered founder recurrence policy blocks
    migration; no row is skipped or reinterpreted silently.
22. Legacy scheduled-call create requires one stable `Idempotency-Key` from both shipped
    clients and commits event, accepted organizer/new attendee state, call projection and
    one invitation outbox together, or none. The creator is an active exact-channel member;
    creator and at most 50 participants are active same-project/same-organization members.
    V2 create can
    link only an owned active unlinked same-organization call, and its normalized mapped
    fields plus organizer/participants must exactly match that locked call. Mismatch returns
    `409` + `SCHEDULED_CALL_LINK_MISMATCH`; inactive/deleted/already-linked owned state
    returns `SCHEDULED_CALL_NOT_ATTACHABLE`; inaccessible/nonexistent stays `404`. The
    link is immutable.
    Ordinary events allow email-only guests; under founder question 23's default a linked
    call with an external attendee is rejected before any mutation.
23. Concurrent v2 and legacy mutations lock event before call. V2 master PATCH/DELETE with a
    stale `If-Match` returns canonical `{error,request_id}`, `409` +
    `CALENDAR_EVENT_VERSION_CONFLICT`, the current quoted event-version `ETag` and a
    same-origin canonical event `Location`, with zero mutation. A terminal event update or
    cancellation returns `CALENDAR_EVENT_NOT_MUTABLE`. Master PATCH with a materialized-exception
    recurrence conflict returns `CALENDAR_EVENT_RECURRENCE_CONFLICT` with neither recovery header.
    Master DELETE remains only `CALENDAR_EVENT_VERSION_CONFLICT|CALENDAR_EVENT_NOT_MUTABLE`.
    Occurrence DELETE returns
    `CALENDAR_EVENT_VERSION_CONFLICT` with the same recovery headers when the master version
    is stale, or `CALENDAR_OCCURRENCE_NOT_MUTABLE` when the instance cannot be cancelled.
    Clients never parse localized `error` text. Legacy PUT still serializes against the
    locked current version. After every commit, canonical content/temporal/attendee fields
    and the call projection agree.
24. Deleting a linked legacy call cancels the event and call atomically. Pausing/resuming
    execution follows founder question 19 and cannot cancel or resurrect the event; one-time
    due completion deactivates only the call projection. Any terminal/invalid legacy pause,
    resume, update or delete returns NFR-13 `409` + `SCHEDULED_CALL_NOT_MUTABLE` and mutates
    neither event nor call; legacy has no optimistic recovery headers because it has no
    `If-Match`.
25. Cancelling the next occurrence advances the projection in the same exception/event/call
    transaction. The due worker validates canonical RRULE/EXDATE/exception state, appends one
    deterministic occurrence-keyed `calls.channel.{channelId}.scheduled_due.v1` outbox row,
    and advances or
    deactivates exactly once. Redelivery is consumer-deduplicated; event sequence/version and
    calendar reminder state remain unchanged.
26. Legacy aggregate returns one linked pair as its existing `call` kind from canonical
    temporal truth and only unlinked canonical events as `event`. V2 availability returns
    the same linked pair once as `event`; active unlinked calls are a pre-cutover-only
    fallback. After activation any unlinked non-deleted call returns
    `AVAILABILITY_SOURCE_INVALID` with no partial availability.
27. A cancelled/excluded occurrence never reappears in the legacy backend aggregate, call
    calendar UI, availability result or due worker. Only distinct internal non-organizer
    series attendees in `needs_action|accepted|tentative` mirror to call participants;
    declined/removed series attendees leave that projection, while occurrence RSVP never
    changes it. External guests never become call participants.
28. Legacy `/api/calls/conflicts` reveals title/channel only when the caller belongs to that
    exact conflicting channel. The Mail meeting UI obtains conflict state from v2 private
    availability and never treats the legacy detail reader as authorization. Any malformed,
    duplicate or over-limit participant UUID input preserves legacy validation status 400;
    it is never silently
    omitted from the authorization/conflict set.
29. The exact-channel call calendar invalidates its scheduled-call query on authorized
    `call.scheduled_due.v1`, then refetches canonical occurrences. Mount/reconnect subscribes
    to every currently returned authorized channel room; channel removal/unmount cleans it
    up. It does not create a second reminder/toast or trust payload ids without normal API
    authorization, and missed best-effort WS recovers on ordinary HTTP refetch.
30. Every exact-channel call list sends paired RFC3339 UTC `[from,to)` bounds. The visual
    view's `23:59:59.999` end becomes the exact next millisecond, including local DST day,
    week, month-grid and agenda boundaries; the server never receives an inclusive end.

### CRM mail seam

1. Only the deal owner receives compose context for a linked contact with a valid email.
   One click obtains the ordinary Mail default send-enabled identity and creates one normal
   draft with one stable idempotency key, putting the recipient only in the draft request
   body. It navigates to `/mail/drafts/{draftId}`; URL/history/query keys/logs contain no
   address, subject or body. Missing recipient is `409` +
   `CRM_MAIL_RECIPIENT_REQUIRED`; missing/unavailable Mail identity creates no draft. A
   lost create response repeats only the same key/body, and failed navigation reopens the
   returned draft id rather than creating another draft.
2. Linking the same owned projected message to the same deal is idempotent. The activity
   stores direction/timestamps/link ids only, never subject, recipient, body, MIME, engine
   id or tracking fields.
3. Deal/admin access cannot grant another user's mailbox access. A foreign message id is
   indistinguishable from not found and causes zero engine/body hydration.
4. Unlink removes only CRM metadata and is audited; it never deletes the mailbox message.
5. All inaccessible/nonexistent deal or message ids use the same `404` envelope. The
   10,001st active link returns `409` + `CRM_MAIL_ACTIVITY_LIMIT` without eviction.
6. The saved-draft route independently authorizes the draft through Mail. Only a
   definitively reconciled submission with an owned `message_id` enables an explicit link;
   queued/submitting/unknown/failed never auto-link. Losing the optional deal-id navigation
   hint on reload safely removes the prompt.
7. Timeline `Open mail` uses `/mail/message/{messageId}` and reauthorizes through Mail;
   the CRM activity row or admin CRM access is never sufficient. Loading/error/empty and
   unlink failure are distinct and accessible.
8. An existing owned inbound or outbound message can be manually linked from the ordinary
   reader. The picker uses the authenticated user id and existing authorized CRM pipeline,
   offers only deals with matching `owner_id`, and exposes no mutation when user/pipeline
   state is unavailable. The server PUT still independently reauthorizes deal and message;
   client filtering is not accepted as security evidence.

### Conditional L1 feed

These cases are required only if founder question 5 admits L1 into the first merge:

1. HTTPS fetch blocks private/loopback/link-local/metadata/rebinding targets and rechecks
   every redirect; non-443 ports, fourth redirect, >10-second, >5/10 MiB compressed/
   decompressed and ratio>10 inputs fail; URL/token never appears in log/error/API.
2. ETag/Last-Modified 304 advances freshness without event rewrite; failed refresh retains
   the last complete snapshot and marks it stale/error.
3. Owner/source uniqueness, bounded recurrence, disable/delete and cross-feed isolation
   hold. Max 10 active feeds, 5,000 VEVENTs, 500 VTIMEZONEs, 500 attendees/event and
   hourly/manual-refresh rate/backoff boundaries are enforced. No feed can overwrite Hub-
   created or mail-import events.

### Compatibility and admin

1. Existing `GET/PUT /api/admin/mail/settings` remains flat and supports
   `hub_mail_url`, `from_alerts`, `alerts_enabled`, `daily_limit`, and read-only
   `resend_ready` without frontend changes.
2. Existing `POST /api/admin/mail/broadcast` is documented as a deprecated governance gap:
   it currently reaches active users but conflicts with ADR-NOTIF-01. Mailbox acceptance
   requires removal/disablement or explicit founder re-authorization; the new mailbox must
   never invoke or expand it implicitly.
3. Mailbox policy uses versioned `GET/PUT /api/admin/mail/mailbox-settings` with `If-Match`
   and remains separate from Resend.
4. Every admin mail operation rejects a non-admin with 403 and writes its named audit
   action without secret values.
5. Existing scheduled-call success response fields remain compatible while bridge-owned
   behavior is additive. No client-supplied call id can bypass server-side organization,
   project, channel, owner or one-to-one-link checks.

## Event subjects

- `mail.received.v1`
- `mail.sent.v1`
- `mail.message.changed.v1`
- `calendar.invite.created.v1`
- `calendar.invite.updated.v1`
- `calendar.invite.cancelled.v1`
- `calendar.rsvp.changed.v1`
- `calendar.reminder.due.v1`
- existing channel realtime family gains
  `calls.channel.{channelId}.scheduled_due.v1` through durable outbox with a stable occurrence
  dedupe key; the obsolete unconsumed `calls.scheduled.due` publisher is removed

Every consumer must prove duplicate-event idempotency. Unknown schema versions go to the
existing dead-letter/error policy and are not heuristically parsed.
Contract fixtures must also prove every envelope/payload rejects undeclared top-level or
payload properties, including `body`, `description`, raw MIME, attachment content, token
and secret-shaped fields.
`mail.received.v1` and `mail.sent.v1` admit at most 200 normalized `to` addresses, matching
the engine projection-address ceiling; 201 rejects before outbox insertion.
`mail.message.changed.v1` additionally accepts only the frozen changed-field/flag enums;
`deleted` has no live engine folder/flag state, preserves Hub-owned labels/verified links,
and removes search/counter membership idempotently.

Required `EnvelopeV1` subject contract fixtures accept an exact 255-ASCII-byte subject and
reject 256 bytes; a multibyte subject with at most 255 code points but more than 255 UTF-8
bytes; invalid UTF-8; an empty token or leading/trailing dot; wildcard, Unicode whitespace
or control characters. Every rejecting-subject error excludes payload text. The product
contract is valid UTF-8, 1..255 bytes, dot-separated nonempty tokens, no leading/trailing
dot or `..`, no `*`/`>`, no Unicode whitespace/control characters, and suffix `.v1`; this
is the Hub product/safety bound, not a NATS transport/control-line ceiling. It applies only
to the new `EnvelopeV1` builder, never globally to legacy `OutboxEvent` or the publisher.

For every new subject, the producer passes a complete deterministic envelope to the
existing outbox as `PublishInTx(ctx, tx, envelope.Subject(), envelope)`. Contract oracles must
prove that stored `event_type`, published NATS subject and JSON `subject` are identical;
the stored/published bytes contain the full envelope rather than only inner `payload`.
For `calls.channel.{channelId}.scheduled_due.v1`, the schema's
`CALLS_CHANNEL_SUBJECT_PAYLOAD_MATCH` invariant captures the UUID from `/subject`, normalizes
it as lowercase UUID text and requires equality with `/payload/channel_id`. A producer
mismatch is rejected before outbox insertion; a consumer mismatch is dead-lettered with zero
calendar, call, notification or receipt mutation. Fixtures must cover equal mixed-case UUID
text, a different valid UUID and a malformed subject. Calendar invite payloads admit at most
200 unique `attendee_ids`; 201 is rejected before outbox insertion.
Envelope `event_id` is stable across transaction retry/reconciliation and consumer
redelivery and is the new-consumer dedupe key. The distinct outbox row id in the
`Hub-Event-Id` header is diagnostic only. Subject mismatch, malformed envelope and unknown
version must reach dead-letter with zero projection/domain mutation. Existing legacy
outbox payloads and worker behavior must remain byte-semantically unchanged.

The 04L envelope immutability/payload amendment is `STATIC_ACCEPTED` at
`4525c559da22729aa8bc51ff5336c29dad4f27e9`; status sync is commit `1c0299bd` and no redispatch
is required. Dispatcher-enforced provenance is recorded and actual telemetry remains `UNKNOWN`.
Required 04L unit oracles also
prove that `EnvelopeV1` exposes no mutable exported field/reference; construction snapshots
only a successfully serializable non-null JSON object and stores frozen complete-envelope
bytes; `MarshalJSON` returns a fresh caller-owned copy; public reads are `Subject() string`
and `EventID() uuid.UUID`. Mutating source map/nested slice/struct-pointer values after
construction, and mutating returned marshal bytes, cannot change later bytes, subject or
event id. Valid struct/map objects pass. Scalars, non-nil arrays/slices, funcs/chans,
typed-nil map/slice/pointer/function/channel/interface, `null`, marshal-error and custom
array/scalar/null marshal payloads reject without panic as stable, content-free
`ErrInvalidPayload`. A custom `json.Marshaler` whose source value successfully returns a
non-null JSON object must pass; its object bytes are frozen, so later source mutation and
mutation of a `MarshalJSON`-returned slice cannot alter later envelope bytes. An
RFC3339-capable `time.FixedZone("UTC", 0)` occurrence passes, normalizes to `Z` and produces
schema-valid output; nonzero offset rejects. Launcher is `NOT_READY`; runtime is `NOT_RUN`;
global Mail review remains pending and no product/runtime acceptance is implied.

`mail.>` and `calendar.>` admission remains the 04G/Phase configuration gate, not 04L.
Before producers are enabled, the shared EVENTS stream must contain them; linked-call due
reuses the already admitted `calls.channel.>`. The generic outbound webhook filter must not
gain any of these wildcards as part of this work; any future webhook exposure needs a
separately reviewed payload allowlist.

## Machine contract conformance

1. Every non-admin OpenAPI operation has exactly one `x-hub-authz` value declared in
   `x-hub-authz-vocabulary` v1. An absent/unknown policy fails the contract gate. Every admin
   operation retains `x-hub-rbac: admin` and its named audit action.
2. Mailbox ids resolve only inside the authenticated user's mailbox; calendar reads require
   organizer/internal-attendee visibility, writes require organizer ownership, RSVP requires
   the matching internal attendee, and CRM link/unlink requires both deal and mailbox-message
   ownership. A missing scope relationship and a nonexistent id return the same NFR-13 `404`
   before engine/body hydration or mutation and do not disclose another organization. The
   read-only `GET /api/mail/account` probe is authorized against the user's mail namespace,
   requires no existing account resource and may return the safe `not_configured` state.
3. Every public request/response object and `ErrorEnvelope` rejects undeclared properties.
   The only intentionally open OpenAPI object components are internal thread/message field
   bases, and they are exposed only through public wrappers with
   `unevaluatedProperties: false`.
4. Standard OpenAPI examples must validate against their adjacent schemas and cover: a page
   stopped before the next complete item with a non-null cursor; calendar version `409` with
   `X-Hub-Error-Code`, quoted `ETag` and same-origin `Location`; mailbox `423` with bounded
   integer `Retry-After`; and a calendar create whose identical key/request replays the
   original `201` and byte-equivalent body without another mutation.
5. Draft `to`, `cc` and `bcc` each admit at most 100 normalized unique recipients. Their
   cross-collection normalized count must be at most policy `max_recipients`, whose persisted
   and public maximum is 300. Exactly 300 is accepted when each list fits; 301 rejects before
   draft save, submit, receipt success or engine I/O.
6. The deprecated legacy admin broadcast remains blocked for mailbox acceptance and is not
   re-authorized by this contract. While the compatibility route still exists, unknown input
   fields reject, subject/body have explicit UTF-8 ceilings, counters are non-negative int32,
   `sent + failed <= recipients`, and dry-run requires zero sent/failed. Phase 148 still
   requires disable/removal unless the founder separately re-authorizes the feature.

## Docker mail profile contract

The eventual profile is optional and invoked only through the accepted versioned outer
launcher. The direct Compose command is operator documentation, not accepted test evidence:

`docker compose -f docker-compose.local.yml --profile mail up -d`

| Component | Intended image/config | Network exposure | Persistent data | Hard limit allocation | Current evidence |
|---|---|---|---|---|---|
| Mail engine | Stalwart `v0.16.17` source candidate; Docker Hub OCI index `sha256:a8108e19…`, linux/amd64 manifest `sha256:2bcbd379…`; signature/SBOM not verified and candidate not approved | Hub-only shared `/jmap` endpoint with separately authorized Registry and mailbox method contexts; exact OAuth helper exception only if accepted; optional SMTP delivery only; no user IMAP/JMAP/ManageSieve/admin mailbox mutation | named `/etc/stalwart` config and `/var/lib/stalwart` data volumes | ≤3 CPU / ≤3 GiB target, subject to measured rebalance | registry metadata only; runtime NOT_RUN |
| Hub mailbox domain/account/combined projection-recovery/rule workers + API | existing backend image/process | existing backend network only | PostgreSQL/outbox | within remaining mail budget when profiled | NOT_RUN |
| Existing NATS/Meili/SeaweedFS | shared stack, not duplicated | existing internal network | existing volumes | shared-service attribution must be reported separately | NOT_RUN |

The entire added mail profile must remain at or below 4 CPU and 4 GiB RAM. Target
allocations are not proof. The accepted run must capture actual idle and representative
load, health, restart persistence, JMAP session discovery, authenticated access, SMTP
delivery, incremental and base-anchored-enumeration recovery, two-replica lease exclusion,
worker cancellation/drain, and exact cleanup.

### Environment matrix (names only; never values in artifacts)

The 2026-08-13 exact-source review invalidated both the old single administrator-secret
shorthand and the interim separate-URL model. A Stalwart management API key authenticates
only reviewed Registry `x:*` methods on shared `/jmap`; it cannot authenticate standard
JMAP mail. Typed operation context, not URL, selects the credential, and a JMAP batch may
not mix Registry and mail methods. The following names express the required separation;
any enabled mode remains rejected until the mailbox authentication/delegation design in
founder question 42 passes security review.

| Variable | Required for | Secret | Safe default |
|---|---|---:|---|
| `MAILBOX_ENGINE_ENABLED` | mailbox runtime | no | `false` |
| `MAILBOX_EXTERNAL_ENABLED` | public SMTP send/receive | no | `false` |
| `MAILBOX_EXCLUSIVE_FOLDER_WRITER` | enables automatic unknown-folder recovery only for an accepted no-alternate-writer profile | no | `false` |
| `MAILBOX_INSTANCE_DOMAIN` | addresses/UID domain | no | empty; startup refuses external mode |
| `MAILBOX_ENGINE_ORIGIN` | trusted internal origin for shared `/jmap`, individually allowlisted `/api` helper path+method pairs, health and same-origin JMAP discovery/session URLs | no | empty; enabled startup refuses it |
| `MAILBOX_ENGINE_MANAGEMENT_API_KEY` | scoped provisioning through exact Registry methods on fixed `/jmap` only | yes | no repository value; never used for standard mail methods, mixed batches, upload/download or `/api/*` helpers |
| `MAILBOX_ENGINE_MAIL_AUTH_MODE` | mailbox JMAP authentication/delegation mode | no | `disabled`; no other mode accepted before G145-5B |
| `MAILBOX_ENGINE_MAIL_PRINCIPAL` | conditional mailbox JMAP identity selected by G145-5B | yes-like identity | no repository value |
| `MAILBOX_ENGINE_MAIL_SECRET` | conditional mailbox JMAP credential selected by G145-5B | yes | no repository value; never used for Registry `x:*` methods or mixed batches |
| `STALWART_PUBLIC_URL` | conditional single advertised OAuth/OIDC/JMAP origin selected by G145-5D | no | absent; never guessed from request headers or set before the OAuth topology is accepted |
| `MAILBOX_REMOTE_IMAGE_TOKEN_KEY` | encrypted/signed proxy tokens | yes | startup disables remote-image loading if absent |
| `MAILBOX_ENGINE_VOLUME` | local operator override | no | named volume |
| `MAILBOX_MAX_MESSAGE_BYTES` | policy | no | contract-defined bounded value at implementation |
| `MAILBOX_MAX_ATTACHMENT_BYTES` | policy | no | contract-defined bounded value at implementation |
| `MAILBOX_TRASH_RETENTION_DAYS` | retention | no | `30` |
| `MAILBOX_JUNK_RETENTION_DAYS` | retention | no | `30` |

DKIM private keys are generated/stored by the engine volume or secret provider and are
never environment examples containing key material.

The source snapshot and endpoint/object evidence are recorded in
`.planning/phases/145-stalwart-gateway/RESEARCH-2026-08-13-STALWART-0.16.17.md`.
It is not image provenance or runtime evidence. Personal per-mailbox rules must use RFC 9661
`SieveScript/*` on `/jmap`; management `x:SieveSystemScript` and `x:SieveUserScript` are not
substitutes. The documented `x:SpamTrainingSample` object does not yet freeze Hub's
message-to-spam/ham mapping or retry idempotency, so classifier feedback remains blocked.
The management API-key design is provisionally narrowed to `Replace` with only domain and
individual-account CRUD/list permissions, pending independent security review. Exact
candidate source confirms mailbox `Identity/get|set` supplies stable ids for configured
account addresses, so identity mapping is not a port mismatch; its enabled use remains
blocked by G145-5B mailbox authentication. Shared-account ACL is not a full-engine fallback:
selected-version Email/Mailbox methods admit shared access, while Identity,
EmailSubmission, Sieve and Blob require account membership. OAuth exposes Authorization
Code and Device Authorization, not client credentials. Until founder question 42 selects
per-user Authorization Code OAuth or explicitly accepts broad impersonation, config accepts
only `MAILBOX_ENGINE_MAIL_AUTH_MODE=disabled` and every other mail-auth mode fails before
network I/O. QA must also prove both typed contexts use fixed `/jmap`, wrong-context
credentials and mixed Registry/mail batches fail before I/O, and provisioning never uses
`/api`. If OAuth is selected, one `STALWART_PUBLIC_URL` also drives JMAP discovery, so
QA must prove one accepted HTTPS origin for browser and Hub, exact user-visible OAuth paths,
an exact-method/rate-limited `POST /api/auth` exception, and negative access to shared
management/mail `POST /jmap`, every unrelated `/api/*`, upload/download, admin/static UI and
telemetry from the user-facing boundary. Prefix-wide `/api`, default/allow-all
`Http.allowedEndpoints`, steady-state 8080, unused IMAP/POP3/submission/ManageSieve
listeners, permissive CORS and untrusted forwarded headers
are release blockers. The selected Domain object exposes expected zone
text but no manual-DNS verification result, so domain promotion/account provisioning remains
blocked by founder questions 43-44 and G145-5C. `STALWART_RECOVERY_ADMIN` is operator-only
bootstrap/recovery state and is absent from the steady-state Hub credential matrix.

## Fixtures and clocks

The 0.1.8 delta additionally requires maximum / maximum+1 oracles for projection item/page
bytes, message/blob budgets, every paginated response count/cursor/8 MiB budget, 16 MiB
legacy calendar/call rejection, stored-resource quotas and 50 MiB/25 MiB policy ceilings.
Compile/static oracles must also prove that projection pages structurally contain neither
body nor attachment bytes, every page/change/enumeration request carries one unambiguous
cursor/checkpoint plus limit, and `EngineSubmission` contains only the confirmed draft
reference/version/key and a bounded final-MIME ceiling rather than another content copy.
Body-union fixtures prove an inbound rich read exposes one untrusted rich source only and
cannot reach API/search before sanitization/plain derivation, while an outbound rich draft
contains the accepted sanitized rich value plus only its server-derived plain value.
Quota failures must replay from the same receipt without later creating a resource after
capacity is freed. Draft fixtures must prove lost first-create reconciliation by immutable
engine draft key and reject a second draft claiming the same staging scope/attachment.
Schema fixtures must exercise every tuple/nullability/state CHECK, the receipt operation and
bulk count/code matrices, and canonical rule-hash byte vectors. These are required future
oracles, not current executable evidence.

- All RFC fixtures use reserved domains (`example.test`, `.invalid`) and synthetic users.
- Calendar tests inject a clock. `DTSTAMP`, generated UID suffixes, and notification times
  are deterministic.
- DST fixtures must use real IANA zones and dates on both sides of a transition.
- MIME fixtures include folded headers, encoded words, multipart nesting, invalid depth,
  oversized values, duplicate headers, and malicious calendar properties.
- No test contacts real DNS, SMTP, remote image, or public mailbox service.

## 144-02A/B/C/D future persistence and deletion oracles

| Boundary | DB direct-write oracle | DB-only expected result |
|---|---|---|
| MailAccount delete shape | Write every status/desired/engine/lease combination and inspect `mail_accounts_deleting_blocked_idx (next_provision_at,id) WHERE status='deleting'` | named status/lease checks reject invalid tuples; deleting requires desired false; the partial index is blocked-observability-only and contains only deleting rows, never a recovery/claim/drain queue |
| Folder | invalid control tuple, bounds, role-parent/self-parent, counters/epoch | named NULL-safe CHECK/unique/FK rejects; cycle/depth/quota are not DB claims |
| Thread/Message | invalid outer JSON, cardinality, wrong member base type/presence, invalid scalar state/bounds | only the stated CHECKs reject |
| MessageFolder/Label/MessageLabel | negative epoch, invalid stored bounds, both flags false | named CHECK/unique/FK rejects |
| Object/Attachment/Draft/Submission | invalid object delete lease pair/state/schedule/expiry, backing/owner/state, `queued\|submitting\|unknown` submission with `mail_draft_id IS NULL`, and declared delete edge | named CHECKs reject; pending/confirmed have no delete controls, orphaned/deleted have due time no earlier than expiry; `mail_objects_delete_due_idx (next_delete_at,id)` contains only orphaned/deleted; Attachment→Object restricts; submission-state rejects nonterminal detach |
| FK action metadata | Inspect `pg_constraint.confdeltype` (or exact generated migration action) separately for Organization→Account, User→Account, Organization/User/Account→Object and Attachment→Object | first five are exact `NO ACTION`; Attachment→Object is exact `RESTRICT`. This independently proves Organization→Object without relying on a delete also blocked by Organization→Account. |

| Boundary | Application-validator oracle | Validator-only expected result |
|---|---|---|
| Thread/Message JSON | unknown member, noncanonical address/name, duplicate, per-member limit, serialized-byte ceiling | reject without transformation; default HTML escaping precedes byte measurement |
| Same-account and lifecycle | cross-account join, folder cycle/depth/quota, received marker rewrite, 51st selected attachment | locked service transaction rejects; exceptional terminal `sent\|failed` draft detach remains the specified `ON DELETE SET NULL` lifecycle behavior |
| Account deleting blocked shape/fencing | Future static contract oracle: direct-SQL rejects deleting with any sync/mutation/provision lease, attempted tuple, provision error, confirmed-absence tuple, or null due marker; exact blocked-pending shape passes. Separately inspect every ordinary-worker claim/finalize guard. | Schema-ready representation has desired false, no lease/attempt/confirmed absence and the due marker; ordinary workers cannot claim/finalize it. It starts no admission worker, precleanup, drain or partial mutation. No engine removal, object drain or hard-delete is runtime-tested or authorized until the accepted engine deletion-port amendment and separate source plan. |
| Deleting HTTP guard oracle | Exercise every account-scoped `/mail/**` operation except `GET /mail/account`, admin account identity list/create/delete, and CRM mail-content read/link/unlink with a resolved deleting account. | Each returns `409 MAIL_ACCOUNT_DELETING` before receipt creation/replay, expansion/snapshot, storage or engine I/O; mutation has no Retry-After. GET is metadata-only deleting with null error, all capabilities false and optional bounded poll hint. |
| Presign versus delete/expiry | Future static contract oracle only (not a runtime test): inspect the representation/checks and proposed future protocol against a PUT started before expiry, an overlong PUT and a delete boundary. | `upload_drain_until=expires_at+11 minutes`; the future protocol may never claim/I/O/hard-delete before it, replay never extends it, and provider/edge capability must prove the <=10m PUT plus <=60s skew bound. No current provider/delete worker is executable. |
| Object delete lease/CAS | Future static contract oracle only (not a runtime test): examine the future protocol's crash/lost-response/two-token/failure/success/authoritative-NotFound branches. | Only a future accepted engine deletion-port amendment plus separate source plan may define provider `DeleteObject` behavior. The required future invariant is exact current token/state/scope CAS, expired-lease-only recovery, stale-token zero rows/no compensating I/O, and at-most-one database hard-delete; this candidate authorizes none of those branches at runtime. |
| Blocked destructive execution | Static negative oracle: enumerate all state/authority edges from the blocked account/object representation and inspect the candidate for an edge to a worker claim, provider I/O or final hard-delete. | No such edge exists: a blocked state produces no worker claim, no storage/provider I/O and no final hard-delete. This is a non-runtime contract check, not a claim that a worker was executed. |
| Physical User deletion | Future static contract oracle: request physical User deletion or a cross-domain GDPR mutation at the current authority boundary. | NOT_DISPATCHABLE/BLOCKED before any GDPR mutation pending an accepted durable intent/job/route contract, accepted engine deletion-port amendment and separate source plan; no mail precleanup, drain, partial mutation, User wait or async finalization is inferred. |
| User organization/deactivation/email | Attempt organization change with an owned account or any physically present uploader object; separately deactivate/change email with objects present. | Organization change requires both zero sets and aborts before contacts; deactivation/email do not reject solely for objects and perform canonical-union contact cleanup. |
| Organization deletion/races | Request physical Organization completion with accounts/objects. | `NO ACTION` remains fail-closed; completion is BLOCKED by the missing engine-port amendment and no eventual drain/completion is claimed. |
| Named direct blockers | Direct-delete Organization/User/Account with each named account/object blocker surviving, then inspect a fixture without that blocker. | named `NO ACTION`/`RESTRICT` FK blocks exactly when its row survives; tests make no universal claim that arbitrary direct SQL is safe without a named blocker. |

No QA claim may treat every malformed nested JSON form as a DB rejection.

> **Local 01B clause reconciliation:** older 144-01B Docker-blocked/do-not-start/dependency wording in this blocker record is `CONSUMED/HISTORICAL` only; every launcher, 02E/Ent/generation/migration/service/API/UI/product/phase/release, Rule/Sieve and Phase145 blocker remains current.

## Current blockers and honesty

### Accepted-02D/02F external/static oracle matrix

The initial `63e`/`6c` candidate lifecycle and earlier Lifecycle FAIL are historical. Corrective commit
`7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` (parent
`6c424ed2d6703ceccbb269c37d9abc03e75a0410`, tree
`da3fa6e19316bcf2c7c318a63ec09e1b175f68df`, exact nine
documents) plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` is docs-only
`STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three post-transition exact reviews PASS C/I/M `0/0/0`.
All review provenance is dispatcher-enforced, fallback none, telemetry `UNKNOWN`. 02F
`fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is
docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` with Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`; former lifecycle and Rule-plan-next instructions are `CONSUMED/HISTORICAL`. The accepted Rule plan leaves Task 1 blocked on independent carrier/G145-5B/G145-5C/launcher readiness. Launcher remains `NOT_READY`; runtime stays `NOT_RUN`.

| Oracle | Static/external assertion | Runtime status |
|---|---|---|
| Signature pointer | identity derives only from `default_identity_id`; signature derives post-commit from `default_signature_id`; Upsert false/true lock/If-Match/no-second-PATCH truth table | NOT_RUN |
| Receipt recovery | Folder only: `p=9` tenth I/O, live-tenth no-op contention and p=10 expiry no-I/O terminal 409 with 30-day replay. Bulk: for a due unknown `message_bulk` receipt at each prior attempt `p=1..9`, a live receipt or account lease claimant is a no-op with zero I/O; an expired-lease takeover acquires `MailAccount` first, creates fresh account and receipt tokens, captures the mutation epoch, and in one transaction persists exactly `p→p+1` before exactly one bounded engine I/O. Two contenders yield exactly one slot. Completion with a stale receipt token, account token, or mutation epoch updates zero rows, drops the late result, and causes no compensating or repeated I/O. Separately, p10 stays manual-only with no schedule/reclaim/I/O; same-key replay returns persisted 202, then fresh-key account-locked guard returns 423 with no receipt/snapshot/expansion/I/O. | NOT_RUN |
| Receipt shapes | digest/pair/status/safe-union/lease/terminal/compaction plus exact operation-target-status-safe-code and folder matrices | NOT_RUN |
| Bulk rows | index/count/outcome/training checks; parent-transaction-only member aggregate/cardinality | NOT_RUN |
| Deployment | attempt `0..10`, internal-only exhaustion code, tuple/state/lease, tenth crash/no-11th and late owner CAS token/epoch invariants | NOT_RUN |
| OpenAPI | only `MailFolderCreateConflict`, `MailFolderMutationConflict`, `MailMutationIdempotencyConflict` as named folder/idempotency 409 components | NOT_RUN |
| RecentContact persistence | FK `mail_account_id ON DELETE CASCADE`, nullable `user_id ON DELETE SET NULL`, unique account/email, partial null-user eviction index, canonical/bounds/checks | NOT_RUN |
| RecentContact unlink cap | With 1,000 external rows plus one linked row, unlink clears link/name, yields exactly 1,000 and removes the exact per-account global minimum across all now-null rows; the just-unlinked row can be the victim | NOT_RUN |
| RecentContact lifecycle | 999/1,000 boundary, linked exemption, timestamp+raw-UUID LRU tie, concurrent distinct admit/same-email upsert, INT32 saturation, User-first then the single owned/lifecycle + contact-link + persisted-uploader-object sorted-account union for nonphysical deactivation/email/organization paths; physical User delete is a future static blocked oracle with no writes, while deactivate/email do not reject solely for objects | NOT_RUN |
| RecentContact user races | Link versus delete in a previously unlinked account plus deactivate/org/email invalidation: user lock prevents new/invalid links; exact rows re-read after account locks; no stale name, orphan or cap breach | NOT_RUN |
| RecentContact account-delete race | Account-deleting-first makes later mutation return deleting conflict/not-found; mutation-first commits. No final account cascade is authorized pending engine-port acceptance. | NOT_RUN |
| RecentContact reciprocal user delete | Future static blocked oracle: physical User deletion is NOT_DISPATCHABLE before mutation pending the separate accepted durable intent/job/route contract, accepted engine deletion-port amendment and separate source plan; canonical union behavior remains relevant only to nonphysical lifecycle paths. | NOT_RUN |
| RecentContact display name | Reject invalid UTF-8, CR/LF, Unicode Cc and header injection; enforce 160-rune/640-byte edges; preserve permitted whitespace/Cf/content exactly with no trim/NFC/casefold; never restore stale name | NOT_RUN |
| RecentContact projection | active same-org Hub user first, canonical-email Hub winner, external privacy projection, no hidden persistence fields/route, existing max 30 | NOT_RUN |

### 144-03A Thread accepted-contract future scenarios (not implemented or run)

- Assert the exact exported callable/error grammar including `ThreadKeyVersionV1`, versioned
  ancestor/identity tuple, unchanged NormalizeAddress propagation, no Base64 API, purity and
  no input alias/mutation. The external error oracle checks exact `ThreadPolicyError` mapping:
  nil receiver/unknown kind returns `mail thread policy error` and nil unwrap; invalid returns
  `mail thread invalid`/`Unwrap() == ErrThreadInvalid`; too-large returns
  `mail thread exceeds bounds`/`Unwrap() == ErrThreadTooLarge`; `errors.Is` matches only the
  corresponding sentinel and neither method
  leaks raw input/detail.
- Exercise all raw preflight maxima/+1: References 100, To/CC 200, Ancestors 202, subject and
  strict Message-ID 998 bytes, email 320 and name 640, before UTF-8/rune/Message-ID/IDNA/NFC/
  fold/prefix/group/NormalizeAddress processing. Invalid UTF-8 subject and control subject
  fail safely; malformed/oversize candidates skip while malformed/oversize/zero-key/non-V1
  ancestors fail safely. Verify strict ASCII byte/case equality and ambiguity collapse/skip.
- Prove References right-to-left unique resolution, then InReplyTo, then fallback; ancestor
  resolver receives only same-account live message+live-thread rows and copies exact
  `(MessageID,KeyVersion,Key)`; ancestor result returns its exact version+key, empty subject
  and concrete nonnil empty participants.
- Prove `AncestorLookupComplete=false` returns content-free `ThreadErrorInvalid` with zero
  result, no fallback and no partial result. A resolver/database/cancellation failure aborts
  before Derive/fallback/write as safe retriable projection failure; only a definitive successful
  lookup may skip a candidate, while complete true plus empty Ancestors permits fallback.
- Prove From→To→CC first-occurrence dedupe, bytewise ASCII sort, ReplyTo/BCC/name/user_id
  exclusion, 1..401 participant bounds and seed-only (non-growing) semantics.
- Cover normalized_subject_v1 NFC/fold/NFC, all Unicode-space collapse, longest-first prefixes,
  sixteen removals then preserved seventeenth, ASCII colon-only behavior and postnormalization
  bytes; V1 remains fixed to x/text v0.35.0, while any change needs accepted new version/rekey/
  dual-read.
- Check V1/V2/V3 preimage and SHA-256 literals exactly, version octet 01 and raw versioned
  32-byte key tuples only, and 1 MiB typed JSON participant/reference boundaries. Public
  participants are closed email-only objects with unique 1..401 items; pages/details stop before
  8 MiB with whole items only.

These remain non-executable required oracle scenarios. The exact 19-path Thread candidate is narrowly
`STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`. Thread implementation/test are narrowly
`STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`; Folder source+oracle are narrowly
`STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
`ca4e2854a319a020f7322ae0a9a98e3fb4a07374`. The historical pre-acceptance `f199546513b758ca7b95328b5ccd7680009b732e` /
`6735ae746c7acde657359e8a709cd897c4dce61c` Rule lifecycle is `CONSUMED/HISTORICAL`; the standalone
Rule plan is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4`,
sole-path blob `b86c0423a307fd2673a3d322435922aa2a398e8f`, after Contract Terra/high, Security Terra/xhigh and
Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Corrective `7d29d08c` plus transition `e563a3d3` is now docs-only
`STATIC_ACCEPTED` / `EXACT_SHA_PASS`; the earlier Lifecycle FAIL and pre-acceptance sequence are
`CONSUMED/HISTORICAL`. The following former sequence is a complete `CONSUMED/HISTORICAL` record,
not a current route. The live sequence was transition-review candidate → one exact 28-doc transition
commit without self-SHA → fresh Lifecycle exact-SHA re-review of fcb plus transition with trace Contract/Security rechecks as needed → all exact reviews PASS → separate 02F docs-only acceptance sync; later accepted Rule plan and Phase-145 carrier/G145-5B/G145-5C
remain required. Launcher remains
`NOT_READY` and runtime `NOT_RUN`.

### 144-03B Folder static-accepted source/oracle scenarios (executable evidence NOT_RUN)

- Reflection must prove the exact UUID-only Folder ABI, all 13 error literals/sentinels and
  pointer `FolderPolicyError` methods, no custom role or extra public API. Zero account/ID/
  action target is invalid; only zero parent is root; UUID equality and final ID sorting use
  16 raw bytes. Opaque engine IDs are mapped to Hub UUID projection rows before policy.
- Cover folder raw and postfold validation: valid UTF-8/no Unicode control/NUL, raw 1..120
  runes/<=480 bytes, whitespace-only, NFC, Straße/STRASSE, sigma, fold expansion, and an empty
  or >480-byte final private key. Display remains exact raw input; normalization is collapse/
  trim Unicode space, NFC, x/text v0.35.0 Fold, final NFC, private/no collation.
- Prove complete snapshot before action handling; then rename/reparent/delete exact target
  lookup and system-role outcome before candidate name/parent/reference/children; system
  delete precedes in_use. Create has no target. Verify overlay only, reparent subtree checks,
  0/100/101 slots, delete does not free a slot, decision key only create/rename and zero
  decision/no partial mutation on error.
- HTTP oracle: after nondisclosing target resolution, missing PATCH/DELETE target is 404 and
  system target is `MAIL_FOLDER_SYSTEM_ROLE`; only a resolved custom PATCH target with supplied
  `parent_id` returns temporary 409 `MAIL_FOLDER_TREE_CONFLICT` before receipt/engine I/O.
  Folder-create terminal receipts enumerate exactly `MAIL_FOLDER_NAME_CONFLICT` or
  `MAIL_FOLDER_TREE_CONFLICT` at 409 and `MAIL_FOLDER_LIMIT_EXCEEDED` at 422; none creates a
  row or calls the engine.

- The founder-required test runner/package is `NOT_READY`; no supported execution command
  exists in this branch.
- Direct Docker/Compose or host tests are prohibited fallback paths.
- Stalwart legal acceptance, independent digest/provenance approval, signature/SBOM,
  exact config, least-privilege mailbox auth, domain-verification authority, measured
  resource use, and runtime behavior are pending. `v0.16.17` is
  only a researched source/registry candidate; read-only digest capture did not pull or run it.
- The five-document 0.1.8 entry amendment independently PASSed at
  `5e314fb080d420696642058ae47397e251c23207`, authorizing only 144-01A; 144-01A is
  `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993`. Broader Phase 142/
  cross-contract and Phase 148 cross-discipline review remain pending. Historical 0.1.7
  review evidence is stale for its engine/API/persistence changes.
- Therefore this handoff is static-only: contract plus accepted source/oracle evidence. It must
  not become a global, product or executable PASS claim until supported current-SHA evidence and
  the remaining gates close.

## 144-02C source-authority handoff candidate

Status is `CONSUMED/HISTORICAL candidate lifecycle` only. The accepted 02B source `947e2a27650ff1c6db60af9ce13bb33b6bf19cb0` and receipt `b93c77ae771c6509be54fca8817fbca7617d6853` consume/historicize only the former predecessor-wait wording. The mandatory freeze is: independent Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high precommit reviews; one exact eight-document authority commit; the same three immutable exact-child-SHA reviews; a separate docs-only source-authority acceptance sync; only then root may dispatch the test-first exact13 schema packet. Until that later sync, QA must treat source, acceptance, Ent generation, migrations, services, API/UI, runtime, product, phase and release as not authorized. Preserve 02F-D/02E, engine deletion-port, physical User/GDPR deletion-intent, launcher `NOT_READY` and executable/runtime `NOT_RUN` blockers. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

## 144-03B Rule static-contract test handoff

The prior `f199546513b758ca7b95328b5ccd7680009b732e` / `6735ae746c7acde657359e8a709cd897c4dce61c`
lifecycle is `CONSUMED/HISTORICAL`. The standalone Rule plan is docs/plan-only `STATIC_ACCEPTED` /
`EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4`, blob
`b86c0423a307fd2673a3d322435922aa2a398e8f`, after Contract Terra/high, Security Terra/xhigh and
Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Future RED oracle
must cover ABI/errors, closed union/bounds/Unicode/targets/order/deep-copy/static purity and
reject noncanonical base32 (decode-to-32 plus identical re-encode). 144-02D is docs-only accepted;
no test before Rule Task 1's independent Phase-145 carrier/G145-5B/G145-5C/launcher prerequisites are all accepted, without implied order; 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after the three PASS `0/0/0` reviews, and its former lifecycle is `CONSUMED/HISTORICAL`; launcher `NOT_READY`.
Compiler errors are internal; HTTP uses 04M account-scoped lookup which masks foreign/missing
targets identically and never makes mismatch/ambiguity an existence oracle.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan-next text is `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. Accepted Rule plan `669ef761` does not authorize Task 1, which awaits independent carrier/G145-5B/G145-5C/launcher readiness; Task4 is post-future-Task3 source acceptance only. 144-01B remains Docker-blocked, Phase 142/148 and release pending, launcher `NOT_READY`, executables `NOT_RUN`.

> **02B precedence (2026-08-15):** source `947e2a27650ff1c6db60af9ce13bb33b6bf19cb0` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; `IMPLEMENTATION-144-02B.md` records exact13 blobs. Tests first; Contract/high, Security/xhigh and Lifecycle/high precommit and exact-SHA PASS C/I/M `0/0/0`. All prior 02B `SOURCE_AUTHORITY_ACCEPTED`, source-unaccepted and 02C-waits-02B-source clauses are `CONSUMED/HISTORICAL`. Root needs a separately authored/adjudicated narrow 02C docs/source-authority bridge; 02C source is not automatically authorized. Other blockers remain unchanged. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

## 144-04F sync-page policy source-authority receipt

Candidate `4e608e890ec12aeed8de965f9649ae503260b94e` (parent
`c53a5fe60e1283cab58be5efaec7ffa131145b31`, tree
`ed7e2bdab6a44254a7c677603ad0afc298aafed8`, subject `docs(mail): propose sync policy authority`)
is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` for immutable exact9 tree blobs. Contract,
Security and Lifecycle exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/
telemetry `UNKNOWN`. Candidate/precommit/commit-review/separate-acceptance-next clauses are
`CONSUMED/HISTORICAL` for that receipt. Its former immediate exact4 root-dispatch clause is now
`SUSPENDED / CONSUMED/HISTORICAL` for the cutoff correction route. The following remains only the
future exact test-first manifest:
`backend/tests/mail/mailbox_events_test.go`; `backend/tests/mail/mailbox_sync_policy_test.go`;
`backend/internal/mail/mailbox_events.go`; `backend/internal/mail/mailbox_sync_policy.go`.
`mailbox_events.go` changes only received folder IDs, changed folder IDs and changed label IDs from
50 to 100 (100 pass/101 reject), without changing field, subject or version; all new policy types
are confined to `mailbox_sync_policy.go`.

The exact callable is `ClassifyMailboxSyncPage(MailboxSyncPolicyInput) (MailboxSyncPolicyResult,
error)`. Input has `Page`, `Folders`, `Messages`, `Destroyed`, `Labels`, `SyncEpoch`,
`EventCutoff`, and `Event`; Page is exact-one Change/Enumeration. Message has the exact engine
message ID, Message, Thread resolution, Previous, Provenance, ReceivedEventID, ChangedEvent and
CalendarMethod; Destroyed has exact engine ID, Message, Thread, Labels and ChangedEvent. Thread
resolution has Thread, KeyVersion, Key, Ancestors and AncestorLookupComplete; previous has Thread,
FolderIDs, Flags, Labels and optional ReceivedEventID; provenance has optional confirmed
submission/draft `*MailboxSyncPolicyProvenanceProof`, where proof is Relation/EngineMessageID;
reservation is ID/Committed; label definition is Label/EngineKeyword; label state
is Label/Manual/EngineRule; event context is Mailbox/OrganizationID/ActorID/OccurredAt/CorrelationID.
Result has `Advance`, `Folders`, `Upserts`, `Tombstones`, `Counts`, `Events`; Advance is exact-one
Change/Enumeration plus SyncEpoch, with the frozen cursor/checkpoint fields; intents and count IDs
are exactly as in `REVIEW-144-04F-SYNC-POLICY-READINESS.md`. Exact error ABI is `type
MailboxSyncPolicyErrorKind string`, constants Invalid=`invalid`, TooLarge=`too_large`,
Duplicate=`duplicate`, AccountMismatch=`account_mismatch`, and `type MailboxSyncPolicyError
struct{Kind MailboxSyncPolicyErrorKind}`. Sentinels are respectively `errors.New("mailbox sync
policy invalid")`, `errors.New("mailbox sync policy exceeds bounds")`, `errors.New("mailbox sync
policy duplicate")`, and `errors.New("mailbox sync policy account mismatch")`. Content-free
nil-safe Error returns the matching sentinel text or `mailbox sync policy error` for unknown/nil;
Unwrap returns matching sentinel/nil, so errors.Is works.

The future external oracle must cover malformed/duplicate/mixed-account/union and missing/extra
state; change 99+1/100+0 pass, 100+1/overlap/Returned mismatch fail; enumeration/folder/label/event
100 pass/101 fail; deterministic raw-byte order; References threading and BCC; before/equal/after
cutoff, nil ReceivedAt, required post-cutoff reservation, and unused historical/equal reservation;
confirmed submission versus spoofed From/alias/mutable Sent; unknown provenance, replay and no-op;
changed fields/flag enum and destroy; counts; sizes/addresses; no secret leakage; deep copy; and no
panic. All state/proof/ancestor/reservation inputs are caller-provided; no query/generation or
callback. Normalized engine pages require Bounds.Requested 1..100, Bounds.Maximum==100, exact
Bounds.Returned, concrete nonnil slices even empty, and accepted ENGINE-CONTRACT aggregate byte
budgets (overflow too_large). Provenance has zero-or-one proof total (submission or draft, never
both), same-account, byte-equal to current projection EngineMessageID and exact Origin agreement;
BCC is nonempty only for matching confirmed
proof/origin and otherwise empty. Events use only existing constructors; cursors remain only in
Advance. Add proof dual/missing/engine-ID/cross-account/origin/BCC and bounds/nil-slice/byte-budget
rows to the listed oracle.

**ROOT_CORRECTION_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN — cutoff sentences only:** the proposed 144-04F cutoff rule is exact but not accepted authority. A first-seen inbound/unknown projection requires nonnil `ReceivedAt`; if it is strictly after `EventCutoff`, it requires nonzero `ReceivedEventID`, writes that marker and emits exactly one `mail.received.v1`. If it is before or equal to `EventCutoff`, it accepts only a zero reservation, writes a nil marker and emits no received event; a supplied historical/equal reservation is `invalid`. Stable confirmed submission/draft never emits received. This candidate changes no other 144-04F behavior or manifest path. Existing uncommitted exact4 source/test candidates remain `UNACCEPTED` and MUST NOT be committed until Contract/Security/Lifecycle precommit PASS, one atomic exact9 correction-candidate commit, all three immutable exact-SHA PASS, and a separate docs-only acceptance sync. This is not correction acceptance.

The four source/test candidates remain `UNACCEPTED`; the previously committed 144-04F authority acceptance sync is not
a new acceptance claim. No runtime claim: launcher `NOT_READY`, executables/runtime `NOT_RUN`; no 02E/Ent/generation/
migration/service/API/UI/product/phase/release, 02F-D, engine deletion or User/GDPR authority.
