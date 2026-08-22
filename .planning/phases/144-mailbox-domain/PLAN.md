# Phase 144 — Mailbox Domain

## Task 6S — Mail sync lease generation and due-scan schema prerequisite (2026-08-22)

> **Narrow source status:** `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `c8274bd2`.
> This prerequisite owns only the test-first `MailAccount` schema contract and source changes for
> durable `sync_lease_epoch` plus the bounded active due-scan index; it does not claim static
> acceptance, adapter implementation, sync-page receipt design, or any product/runtime authority.

`sync_lease_epoch` is required `Int64`, default `1`, minimum `1`, immediately after `sync_epoch`.
It is the monotonic generation carried later by `MailboxSyncFence` and
`MailboxSyncWorkerDueAccount`: a never-claimed due account is valid at `1`, and a future adapter
keeps `1` for its first lease then increments only when distinct durable lease ownership replaces
the prior lease. `mail_accounts_non_negative` preserves every existing clause and additionally
requires `sync_lease_epoch >= 1`. The appended
`mail_accounts_sync_due_idx` has fields `(next_sync_at, id)`, storage key
`mail_accounts_sync_due_idx`, and exactly
`status = 'active' AND next_sync_at IS NOT NULL`.

Task 6 depends on Task 6S and Task 6P. No sync-page receipt is invented here; the bounded
inline active-lease proof is frozen as the separate Task 6P prerequisite before its apply adapter.
Independent static review found no issues in this narrow manifest. The launcher is `NOT_READY`;
RED/GREEN, Ent generation, Atlas diff/migration, compile, SQL/DB, Docker, formatting/lint, and
runtime are all `NOT_RUN`. No migration or cutover authority follows.

## Task 6P — Mail sync-lease snapshot and apply-proof schema prerequisite (2026-08-22)

> **Narrow source status:** `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `7309c52a`.
> This packet owns only the test-first `MailAccount` schema contract for one bounded inline active
> sync-lease snapshot and its apply proof. It does not add an unbounded receipt table, retention
> work, Ent generation, adapter implementation, migration, cutover, or product/runtime authority.

Immediately after `sync_lease_token` and before `mutation_lease_token`, nullable snapshot fields
carry the claimed sync/mutation epochs, `incremental|enumeration|catch_up` phase, and bounded
opaque cursor/base/checkpoint state; nullable proof fields carry only a positive versioned
32-byte result hash, applied phase, next attempt and application time. No field has a default.
No message/calendar/provider body, raw error, or JSON payload is persisted.

The portable fail-closed `mail_accounts_sync_lease_snapshot` CHECK requires absent-token snapshots
to be entirely null; a present token and expiry require non-negative fences and the exact
incremental/catch-up cursor or initial/resumed enumeration shapes. The separate portable
fail-closed `mail_accounts_sync_lease_apply_proof` CHECK requires an all-null proof or a complete
active-lease proof with a 32-byte hash, `applied_at < sync_lease_until`,
`applied_next_at >= applied_at`, and only incremental→incremental,
catch_up→catch_up|incremental, or enumeration→enumeration|catch_up. Keeping this proof inline
preserves same-lease claim/replay even after the account cursor advances, without receipt
retention cleanup.

Independent static review found no issues in the bounded snapshot/proof checks. Task 6 cannot
start before both Task 6S and Task 6P are present. The
launcher is `NOT_READY`; executable RED/GREEN, tests, Ent generation, Atlas diff/migration,
compile, SQL/DB, Docker, gofmt, lint, and runtime are all `NOT_RUN`. No migration or cutover
authority follows.

## Task 5 — Mail Ent repository root and owner-resolution candidate (2026-08-22)

> **Narrow source status:** `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at head `ea0808a7`.
> Independent review accepted only the test-first owner-repository contract,
> `mailbox_ent_repository.go`, and `mailbox_ent_owner.go`; it does not claim static
> acceptance outside this manifest, Ent generation, migration, SQL/DB proof, provider connectivity, HTTP/API/UI
> wiring, launcher readiness, runtime, product, phase, or release acceptance.

`NewMailboxEntRepository(*ent.Client, MailboxRepositoryClock)` keeps both dependencies private,
rejects nil and typed-nil dependencies, owns one rollback-safe Ent transaction, maps Ent failures
to existing content-free mailbox errors, and validates the private clock as non-zero UTC at exact
second precision before any durable use. Owner resolution locks Organization, active User,
owner-bound MailAccount, then primary MailIdentity. An active actor without an account receives
only the valid `not_configured` record; foreign, inactive, deleting, ambiguous, cross-linked, or
state-inconsistent rows fail closed. The returned record contains copied scalar account/address/
recovery/projection state only and performs no provider or network I/O in its transaction.
The fix review confirmed that actor-owned cross-organization state fails closed before the sole
valid no-account result, provisioning accepts its already validated optional recovery code, and
the global unique `mail_accounts.user_id` invariant prevents ambiguous multi-organization owners.

The author-first contract source is deliberately `NOT_RUN`: generated Ent Go is absent until Task
9, and the canonical launcher is `NOT_READY`. Therefore RED/GREEN, compile, Ent generation,
database, Docker, lint, formatting, migration, and runtime evidence all remain `NOT_RUN`.

## 144-04A owner-resolution seam immutable source receipt

> **Narrow receipt / precedence:** `STATIC_ACCEPTED / EXACT_SHA_PASS` applies only to immutable source `7531565879e0d7aa72a246407fd0e1d609a82c49` (parent `2932748df0d6efdcff4f51a3b0eae2b4b3dd3d7b`, tree `448a6db855ba1b3a6873253967c789def305cda4`, subject `feat(mail): add owner resolution seam`): ordered blobs repository `4520db07c57e80faf6318a1444eb5e2aaf51355b`, service `d33a5911fd8a9ec5d14bf2ecddaaa321c5c6b50a`, test `2b414cbd1edece218976699d2777a80f51bd057b`. Contract, Security and Lifecycle precommit plus immutable reviews PASS C/I/M `0/0/0`; provenance `dispatcher-enforced`, Terra high/xhigh/high as applicable, no fallback, telemetry `UNKNOWN`.

This receipt consumes only lower claims that abstract `mailbox_repository.go`, the owner port or the owner seam is absent or unimplemented. It does not consume current claims that concrete sync primitives, `mailbox_ent_repository.go`, an Ent adapter, SQL, 144-04H, 144-02E, generation or migration remain absent, unimplemented, blocked or separately gated.

This dirty exact6 documentation packet is `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT6_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`; it cannot self-activate. No HTTP/API/UI/wiring, runtime, product, phase or release authority follows. Launcher remains `NOT_READY`; all executables, generation, migration and runtime remain `NOT_RUN`; the external-framework handoff remains a gate.

## 144-04I-A bounded sync-worker source receipt

> **Current source receipt:** `STATIC_ACCEPTED / EXACT_SHA_PASS` at `091f2ef286a71f506e1b048f5386fd37cefa06bc`; exact10 receipt `03674eace728642fc2c9b39d2e8e2c81bf016033` is ancestral to HEAD. The conditional receipt block below is `CONSUMED/HISTORICAL` only for 04I-A. PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback, telemetry `UNKNOWN`; no next packet, launcher, runtime, or broader authority follows.

`SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED / RUNTIME_NOT_RUN`: this dirty
exact10 docs candidate grants no acceptance. Only its exact-manifest atomic commit and immutable
receipt review may accept `091f2ef286a71f506e1b048f5386fd37cefa06bc` (parent
`6b27643fcde3b702fa13c5fcc9b4bfefcb9fc87a`, tree `b01be92acf01033935b85446454eaee16b84e01b`,
subject `feat(mail): add bounded sync worker`) as `STATIC_ACCEPTED / EXACT_SHA_PASS` for
authored-first test `79c2691971287e3a62c830a5d12c78699414bb86`, then source
`6c20fadf563d4d146b2fc1cbacdddea485266eb7`. See the ordered exact10 manifest in
`IMPLEMENTATION-144-04I-A.md`. Precommit plus immutable Contract/Security/Lifecycle reviews PASS
C/I/M `0/0/0`; author Terra/xhigh; Contract Terra/high; Security/Lifecycle Terra/xhigh;
dispatcher-enforced/no fallback;
telemetry `UNKNOWN`. The former 04I-A candidate/source-unaccepted/current-next clauses then become
`CONSUMED/HISTORICAL`; no next packet or blocker lift follows.

04F remains accepted. Before the manifest-equal 144-04G exact12 coherence commit and immutable
C/S/L triple PASS, 04G has no current static, source or downstream authority. 04H/02E, concrete adapter/Ent, generation/migration, SQL/two-replica
proof, API/OpenAPI/UI, launcher `NOT_READY` and executable/runtime evidence `NOT_RUN` remain
blocked.

## 144-05D Content plain-only source receipt sync (2026-08-22)

> **Source-receipt lifecycle:** `SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This dirty exact9 docs-only sync is non-self-activating. It records source `acc593fe4f20669a21a6f5ba70a8a20293111017` (sole parent `98ebedc9dd0b6ad90dbe06e1faf5810376dce392`, tree `7de89d1237bc0c8f640729f607a2dbfd9189f3c4`, subject `feat(mail): add plain content normalization`) with authored-first test blob `37c29c421d21e7c173142647d894de0d890894de`, then source blob `b064a2f6f944d5ffb10651ccb068a77634a9267b`. Source Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` reviews PASS C/I/M `0/0/0`, Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Its receipt diff must set-equal the ordered exact9 manifest in `IMPLEMENTATION-144-05D.md`. Until atomic receipt plus immutable receipt C/S/L PASS, acc has no static acceptance and former 05D authority/acceptance/source-unimplemented/current-next routes remain current. Only after activation is acc `STATIC_ACCEPTED / EXACT_SHA_PASS`; lower routes become `CONSUMED/HISTORICAL`. No next packet follows.

This route owns only pure plain output: `NormalizeMailboxContent` accepts explicit plain/rich input, returns canonical NFC/newline/whitespace plain text and `RemoteImagesBlocked=true`, rejects invalid UTF-8/NUL/C0/C1/bidi and bounded size/rune/token/depth excess with safe typed `invalid`/`too_large` errors, and uses only existing `x/net/html` tokenizer for rich text extraction. It adds no sanitizer dependency and does not edit `go.mod` (`x/net` `v0.51.0` direct; `x/text` `v0.35.0` already present). Accepted 03C-Attachment and 04A are the only dependencies; 04L is not. Rich sanitizer/token/proxy, 05D-Remote and 05E remain blocked; no source outside exact2, schema/OpenAPI/persistence/Ent/API/UI/testframework/runtime authority or next packet follows.

The literal error contract is `MailboxContentErrorKind string`, `Invalid="invalid"`, `TooLarge="too_large"`, exact invalid/too-large sentinels, and `MailboxContentError{Kind}` pointer Error/Unwrap: known kinds exactly return/unwrap their sentinel; nil/unknown is `mailbox content error`/nil; every failure is zero Result. Preflight byte length precedes UTF-8/rune/allocation for both formats; only HT/LF/CR C0 are allowed, then exact bidi/control rejection and exact NFC/newline/horizontal/line-trim/LF-collapse/outer-trim order precede canonical caps. Exported `MaxMailboxContentHTMLTokenBytes=65_536` is the maximum accepted raw tokenizer-token bytes including delimiters/attributes/text. Since pinned `x/net` fails at `rawLen >= maxBuf`, Rich calls `tokenizer.SetMaxBuf(MaxMailboxContentHTMLTokenBytes + 1)` before first `Next`: raw 65_536 passes and 65_537 is `html.ErrBufferExceeded` -> `too_large`/zero Result; static source review owns every other non-EOF-to-`invalid` branch. It counts all non-Error tokens before processing, requires the exact lowercase strict stack/void/self-close/end/EOF behavior at 100000/100001 tokens and depth 64/65, rejects decoded entity controls, and uses the exact suppression/block sets with pre-append extraction cap frozen in the review. `"<!--" + strings.Repeat("x", MaxMailboxContentHTMLTokenBytes-7) + "-->"` is raw 65_536 and succeeds with empty plain output; `MaxMailboxContentHTMLTokenBytes-6` payload is raw 65_537 -> typed `too_large`/zero Result. Suppression precedes breaks: root/descendant/self-closing/void suppression tokens emit neither text nor LF, only non-suppressed block tokens may break, and `<span>a</span><form><p>x</p></form><span>b</span>` is `ab`. The separate unterminated quoted-attribute `"<a data-x=\"" + strings.Repeat("x", MaxMailboxContentHTMLTokenBytes)` remains below body cap yet produces only `too_large`/zero Result/no panic/no leak, while a small quoted attribute succeeds.

## 144-04G receipt-coherence sync candidate (2026-08-22)

> **Coherence gate:** `COHERENCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN`. This non-self-activating dirty exact12 docs-only coherence candidate records immutable source `3031c53d399accc84fe10cd0c5421a85c089ac7c` and prior acceptance receipt `2a2975056c3fb8001d88863a041310271282989d`; their full metadata, blobs and ordered manifests are in `IMPLEMENTATION-144-04G.md`. Source Contract/Security/Lifecycle precommit and immutable `EXACT_SHA_PASS` are C/I/M `0/0/0`; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Only after an atomic manifest-equal exact12 coherence commit and immutable coherence C/S/L triple PASS is 144-04G `STATIC_ACCEPTED / EXACT_SHA_PASS`; former 04G authority/acceptance/source-unimplemented/current-next routes then become `CONSUMED/HISTORICAL`. Historical coherence selects no next Mail packet and grants no runtime/downstream authority. 04F remains accepted; 04H/02E, Ent/SQL/concrete adapter, generation/migration, API/UI, launcher/testframework and runtime remain blocked `NOT_READY`/`NOT_RUN`.

## 144-04G source-authority acceptance-sync candidate (2026-08-20; predecessor route)

> **Historical predecessor boundary:** This exact9 route is historical receipt evidence only. Before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, it confers no current static, source or downstream authority.

Status: `ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED /
SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN`. This dirty exact9 working tree grants no
authority. The `6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` metadata and the three immutable
exact-SHA PASS results below are proposed receipt facts only; they activate only after one atomic
commit of this exact9 acceptance sync. Until activation, the lower candidate/precommit/root-commit/
exact-SHA-next lifecycle remains current and root cannot dispatch source; the separately false
accepted-repository premise remains `CONSUMED/HISTORICAL`.

Historical receipt fact only: after that atomic exact9 sync commit, Mail 144-04G docs authority was recorded as
`SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` only at immutable
`6ba3aed3d9a69a76f98978faf4a9e3362c4bc29b` (parent
`2b6837926c91d652ff7bf30447bd482740261bd0`, tree
`bdd59cfcb31376c95f5128370a75333cad78bc14`, subject
`docs(mail): propose sync service authority`) and the exact9 blobs recorded in the readiness
review. Contract, Security and Lifecycle immutable exact-SHA reviews PASS C/I/M `0/0/0`;
dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Only after that atomic exact9 sync commit is the lower 144-04G
candidate/precommit/root-commit/exact-SHA-next lifecycle `CONSUMED/HISTORICAL`.
Historical record only: the exact9 route identified the frozen test-first exact3:
`backend/tests/mail/mailbox_sync_service_test.go`, new interface/types-only
`backend/internal/mail/mailbox_repository.go`, then new
`backend/internal/mail/mailbox_sync_service.go`. Source remains
`UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own review, atomic commit and immutable
exact-SHA acceptance. 04F stays
accepted; 04H/02E, Ent/adapter, generation, migration, OpenAPI and runtime stay blocked or outside
authority. Launcher `NOT_READY`; executables/runtime `NOT_RUN`.

## 144-04G sync-service authority candidate (2026-08-20; predecessor route)

> **Historical predecessor boundary:** This route is historical receipt evidence only. Before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, it confers no current static, source or downstream authority.

`ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / IMPLEMENTATION_BLOCKED /
RUNTIME_NOT_RUN` is `CONSUMED/HISTORICAL` predecessor evidence only: its future exact3 was test-first NEW `backend/tests/mail/mailbox_sync_service_test.go`,
NEW interface/types-only `backend/internal/mail/mailbox_repository.go`, then NEW
`backend/internal/mail/mailbox_sync_service.go`. The stale claim that the new signatures extend
an accepted repository file and the lower-precedence claim that no such file exists are both
`CONSUMED/HISTORICAL`. The immutable source receipt records the source file, but grants no current
authority before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS.
The self-contained abstract-port ABI and deterministic claim/prepare/apply/resolve state machine
are frozen in `REVIEW-144-04G-SYNC-SERVICE-READINESS.md`. Before the manifest-equal exact12
coherence commit and immutable C/S/L triple PASS, 04G has no current static, source or downstream
authority. 04F stays static accepted; 04H/02E, Ent/adapter, generation,
migration and runtime stay blocked/`NOT_RUN`; Calendar 04C exact3 state is unchanged.

Fix1: Prepare/Apply/Resolve compare persisted sync lease expiry to stable authoritative DB
transaction time, not request.Now, and expiry returns LeaseLost/zero effects without requiring a
takeover. All Changes and initial/resumed Enumerate requests use exact configured PageSize.

## 143-03B-T immutable source acceptance and conditional 143-04C gate (2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, 144-04G has no current static, source or downstream authority. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## 144-04F immutable source acceptance (2026-08-20)

Mail 144-04F is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `4715f939f24e5b5c2d3f80a38d204babe68bdd34` (parent `7ac8714e99db2935ce290ad67836e1309cc05c78`, tree `6461894a90c4340bf96118ed1a242516c02b48c7`, subject `feat(mail): add sync page policy`). Its exact4 ordered test-first blobs are recorded in `IMPLEMENTATION-144-04F.md`: events test `50a41cefb539d9b840a43da9e1ff34cce6c8653f`, sync-policy test `b6410709e4fb8166c28a5c502cfba535eef08d06`, events source `a3017b0a6064d8fc1017e37e2d4f675ee8732945`, sync-policy source `acd642fcdb6290af79b7fa8e5aa36a5036576f47`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file asserting that Mail exact4 is dirty, `UNCOMMITTED`, `UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is `CONSUMED/HISTORICAL`; the immutable source receipt above is current. No automatic Mail next source follows: before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, 144-04G has no current static, source or downstream authority.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted. 143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; 143-04A/B/G receipts remain accepted. Launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. 143-03E, 144-02E, Ent generation and migration stay blocked. No raw iTIP or service/API/UI/product/phase/release authority follows.

## Combined exact17 conditional acceptance sync (2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> **144-02C source receipt (2026-08-20):** `STATIC_ACCEPTED / EXACT_SHA_PASS` is recorded only for `e565523f589fbd7536f6f8e4df7effd39469c1af` (parent `df2dcbc2151c4b2df5c57774c3114b066b183d24`, tree `9fabe45429dfdda49625fe53cd6645e08ed63d52`, subject `feat(mail): add delivery schemas`) and the exact 13 blobs in `IMPLEMENTATION-144-02C.md`. Tests/oracles first; all stated reviews PASS C/I/M `0/0/0`. The former 02C candidate/source-unaccepted/lifecycle-next/implied-next routes are `CONSUMED/HISTORICAL`. No next Mail code packet is selected or authorized; root must select it separately. Launcher `NOT_READY`; executable evidence `NOT_RUN`; no generation/migration/service/API/UI/product/phase/release authority.

> **144-02B projection source-authority acceptance:** combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS`; Contract/high, Security/xhigh and Lifecycle/high corrected-child exact reviews PASS C/I/M `0/0/0`. Root may dispatch only this frozen test-first exact13: new `backend/ent/schema/mailfolder.go`, `backend/ent/schema/mailthread.go`, `backend/ent/schema/mailmessage.go`, `backend/ent/schema/mailmessagefolder.go`, `backend/ent/schema/maillabel.go`, `backend/ent/schema/mailmessagelabel.go`, `backend/tests/mail/mail_projection_schema_test.go`; edits `backend/ent/schema/mailaccount.go`, `backend/ent/schema/calendarevent.go`, `backend/ent/schema/mailmutationreceipt.go`, `backend/tests/mail/mail_account_identity_schema_test.go`, `backend/tests/mail/mail_mutation_receipt_schema_test.go`, `backend/tests/calendar/calendar_event_schema_test.go`.
>
> Accepted PERSISTENCE 02B only: 02F-R allows exactly `MailMutationReceipt.folders` inverse for `MailFolder.control_receipt_id -> receipt RESTRICT`; actor/message scalar-only, other receipt edges/checks unchanged. `MailAccount` folders/threads/messages/labels; `CalendarEvent.mail_messages` inverse with Message FK `SET NULL`; rule targets service-only locked, no DB FK/normalized target invention; compatibility tests mandatory. Source implementation is not accepted and still requires test-first authoring, independent review, atomic commit, exact-SHA review and status sync. Prior candidate/precommit/correction/failure-next clauses are `CONSUMED/HISTORICAL`; immutable `45fd3369` failure remains history. former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; 02F-D/02E/Ent generation/migration/runtime/services/API/UI/release BLOCKED; launcher `NOT_READY`, executables `NOT_RUN`; Calendar accepted/no-next unchanged. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02D source receipt:** `2242a4e7846944498eb5ed701ab8f47c986bd804` remains `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 02F-R source is accepted at `ce542cb5`; former 02B eligibility clause is `CONSUMED/HISTORICAL`, 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, and 02F-D remains blocked. Launcher `NOT_READY`; executables/generation/migration/runtime `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, sole blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787` and sync `375891bb` remain accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; unrelated gates remain current. No source/test/runtime/adapter/schema/service/worker/product/phase/release authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier six-path RED/source additionally awaits accepted Rule compiler source/ABI. 145-07B remains blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No carrier plan/source/test/runtime authority. Rule Task 1/145-07B await separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness; launcher `NOT_READY`, executables `NOT_RUN`.

> **144-02A1 DNS static acceptance (2026-08-15):** typed DNS source is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `5892a90b7582ff4ac35db218428400ac9608c4e3` (parent `29a5c624c0bacc2de3cc4b5a93e3e9176320b5bd`, tree `c78d82ef8ea55a4656d33b82530b2fbb72582201`; source `backend/internal/mail/domain_dns.go` blob `e00586637f209f85c52626c1ec8d96116af934b3`, test `backend/tests/mail/domain_dns_test.go` blob `8a43140964d12cca7cdc09e34675a074ee164d7b`). Tests were authored first. Terra/high static precommit and immutable exact-SHA review PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN` unless exposed. RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime are `NOT_RUN`; launcher `NOT_READY`. Narrow static source acceptance only: no Ent/schema generation/migration/service/API/UI/product/phase/release authority. In A1/A2/A3, all three are `STATIC_ACCEPTED`; A2-unimplemented/A3-next clauses are `CONSUMED/HISTORICAL`, and A3 is accepted at `e9a1764b001e28074e4074a07ca8ac982e5345ba`. Founder root-only source sequencing before final integrated Docker does not authorize Ent generation/migrations/runtime. Existing blockers remain unchanged.

> Task: TASK-237 · Status: IN_PROGRESS · 144-02A1 typed DNS `STATIC_ACCEPTED` / `EXACT_SHA_PASS` `5892a90b`, C/I/M `0/0/0`, static-only · Previous no-further-source/root-only-selection routing is `CONSUMED/HISTORICAL` · Founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker; every exact packet requires its own test-first/static review/commit · 144-01B/02/Rule/Sieve/Phase145 blockers unchanged · no Ent generation/migration/runtime authority · launcher `NOT_READY`, executables/runtime `NOT_RUN` · Depends on: Phase 142

> **Current 144-02D/02F acceptance:** 02D remains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus `e563a3d3be8898371e5c8a9861c58f11b72167bc`. The exact-28 correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high all PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. The former 02F lifecycle and Rule-plan author/adjudication are `CONSUMED/HISTORICAL`. No source/Ent/generation/migration/services/runtime authority follows. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase-145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; Phase 145 remains blocked.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

> **144-02A/B/C/D plus 02F trace acceptance (2026-08-15):** the exact 14 content documents are docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. The previous candidate/precommit/root-commit/exact-review/status-sync-next clauses are `CONSUMED/HISTORICAL`. This accepts no source, Ent/schema, generation, migration, service, test or runtime work. Engine deletion-port and physical User deletion-intent contracts remain separately `BLOCKED` / `NOT_DISPATCHABLE`; 02F is unchanged except dependency tracing. Launcher `NOT_READY`, executables `NOT_RUN`; 144-01B, Phase 142/148, carrier/Rule and G145-5B/C unchanged.

> **144-03C-Attachment static acceptance (2026-08-15):** deterministic attachment policy is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7b2ea5721684e43559d6f6b4b5fc462d8daafb8b` (parent `4e7c609713983a8f6db6fbfeb7474abd73451e8c`, tree `4c3fe4d27e170c5d3f42b116c39887036ce936c6`; source blob `3f67ce1b9f2f056b70bfd3070f1c34f98b5c71c1`, test blob `bf7eda125d5aba1d0cc2594746d14e7bcfd06aa7`). Tests were authored first. Terra/xhigh authoring; Contract Terra/high, Security Terra/xhigh and ABI/Lifecycle Terra/high precommit+exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. It is pure deterministic per-attachment policy: count<=50, NFC output/NFKC structural+extension defense, closed MIME/type deny tables, inline CID, SHA/scan gates, validate-versus-send/read clean-only behavior, typed content-free errors and deep copies. RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime are `NOT_RUN`, launcher `NOT_READY`. No scanner/storage/network/project/tenant auth/schema/migration/service/API/UI/product/phase/release authority. 144-01B/02/Rule/Sieve/Phase145 blockers and Calendar 143-03E launcher+migration blocker remain unchanged. Its prior no-further-source/root-only-selection routing clause is `CONSUMED/HISTORICAL`; founder now directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker, each with its own test-first/static review/commit. No Ent generation/migration/runtime authority follows.

> **Historical 144-03D receipt remains valid:** deterministic event mapping is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `42c6b9f2bd4383040fc116630727564fcf17dd78`; its exact source/test blobs and C/I/M `0/0/0` evidence are retained in `IMPLEMENTATION-144-03D.md`.

> **144-01B static acceptance (2026-08-15):** immutable `0c064363b5ccb657a7c444802e94075c4ca743e5` (parent `48a7273d100c07788def88e2f1b8d69c7173580f`, tree `9c568a0c58c6bf7b8c3171cfd934760b839db623`, subject `feat(mail): add deterministic engine fake`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for fake `12bcd921668e406f07a98f7f3b0618ad8d44e81d` and authored-first test `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f`. Terra/xhigh author; Spec Terra/high, Security Terra/xhigh, ABI Terra/high precommit and exact-SHA Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 01B Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executables `NOT_RUN`; no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority.

> **144-02A2 static acceptance:** `65c91b4824e8a1948fc97c149305aa7f009de061` remains `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 02F-R source is accepted; former 02B eligibility clause is `CONSUMED/HISTORICAL`, 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, and 02F-D/02E remain blocked. Launcher `NOT_READY`, executables `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02A3 static acceptance:** `e9a1764b001e28074e4074a07ca8ac982e5345ba` remains `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 02F-R source is accepted; former 02B eligibility clause is `CONSUMED/HISTORICAL`, 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, and 02F-D/02E remain blocked. Launcher `NOT_READY`, executables `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

> **144-02F-R source acceptance:** `ce542cb5de839a1fb5200295f710b0ae45bcdbe6` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact eight blobs in `IMPLEMENTATION-144-02F-R.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Authority-only/source-unaccepted and 02B-separate-authority/02C-predecessor route `CONSUMED/HISTORICAL`; 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation unaccepted, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL, 02F-D blocked; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

## Goal

Implement the engine-independent domain and HTTP behavior against a deterministic fake. `144-02A1`, `144-03C-Send`, and `144-03C-Attachment` are static-only accepted packets. The former no-further-source/root-only-selection routing is `CONSUMED/HISTORICAL`. Founder directs root to continue remaining Mail+Calendar atomic source packets before final integrated Docker; every exact packet still requires its own test-first/static review/commit. This does not authorize Ent generation, migrations, or runtime; launcher is `NOT_READY` and executables remain `NOT_RUN`.

`143-04L` and `144-01A` are satisfied for `144-03D`; Address is `STATIC_ACCEPTED` at
`41564d003946ddb78df18efca166ca86e4055405`. Thread contract remains accepted at `cbc12577` and
its exact 19-path implementation/oracle boundary is `STATIC_ACCEPTED` at `9903f4df` /
`EXACT_SHA_PASS`, C/I/M `0/0/0`; this remains static-only evidence. Folder source+oracle are
narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
C/I/M `0/0/0`), static only; Folder docs acceptance is `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
The distinct Rule contract is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; its acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`. The separately accepted Rule plan is at full `669ef761` (parent `77eb`, tree `1b554d`, added plan blob `b86c0423`) and authorizes neither source nor TDD. The future implementation boundary remains exactly `rule_compiler.go` plus `rule_compiler_test.go`, but only after all accepted non-marker Phase-145 engine-port/typed-IR/script carrier, G145-5B, G145-5C and launcher readiness.

> **Local 01B clause reconciliation:** within this ordered-task record, every older assertion that 144-01B is Docker-blocked, must not start, or awaits supported current-SHA Docker evidence is `CONSUMED/HISTORICAL` only. It changes no 02E/Ent/generation/migration/service/API/UI/product/phase/release blocker.

## Ordered tasks

1. `144-01` — public `MailEngine` port, DTOs, error taxonomy and fake; write adapter
   contract units under `backend/tests/mail/**`.
2. `144-02` — Ent schemas for account, identity, domain/policy, folder, thread, message,
   message-folder membership, label, mail object, attachment, draft, user settings,
   signature, rule/deployment, mutation receipt/target/member, contact and submission;
   Docker-only generation.
3. `144-03` — pure address normalization, threading, reply/forward quoting, folder tree,
   typed rule compiler, rate/loop and attachment policies plus deterministic mapping of
   all three mail events through the shared 143-04L envelope, with failing units first.
4. `144-04` — repositories/services for folders, labels, list/read/bulk, draft revisions,
   unknown-save reconciliation, send/generic mutation idempotency/reconciliation, owned
   provisioning, settings/signatures and desired rule generation; then the
   separately packetized incremental/enumeration sync policy, epoch-fenced service, atomic Ent
   apply plus domain-control/account-control/projection-with-mutation-recovery/rule bounded worker lifecycles
   against fake ports.
5. `144-05` — dedicated mail-storage presign/confirm adapter, scanner interface,
   plain-only content normalization followed only by separately blocked rich sanitizer/remote-image
   proxy work, versioned event projection consumption/subject
   admission, bounded body/filename search hydration, recent contacts, notifications/ws/
   audit seams; project file ACLs are not reused.
6. `144-06` — Fiber handlers/routes matching `docs/mail/openapi.yaml`, owner-visible
   account/projection state, admin policy/health/alias DTOs, error mapping and genuine units.
7. `144-07` — after Phase 143 admits `calendar.>`, add `mail.>` to EVENTS without changing
   the webhook filter, then system/architecture/security docs sync and independent review.

Each item is one or more atomic commits split further when generated schema, service,
handler, or docs cannot be reviewed independently. Existing Resend alert behavior stays
separate and green.

> **02A source gate:** the accepted b631 snapshot is representation-only and is not an
> implementation commit. It does not authorize Ent/product source or TDD, generated output,
> migrations, launcher execution, tests/validators/build or runtime. Founder migration authority,
> supported task-owned launcher, sequential explicit source dispatch and source-level review remain
> required. 144-02D is docs-only accepted; before any source work, 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 02F and Rule-plan authoring/review lifecycle is `CONSUMED/HISTORICAL`.
> The Rule plan is already `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; its former author/adjudication instruction is `CONSUMED/HISTORICAL`. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase-145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, with no implied order; post-task runtime is `NOT_RUN`.

> **02F-R current source:** `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`; authority precursor `a877848e` is `CONSUMED/HISTORICAL`. Former 02B eligibility/not-auto-dispatched clause is `CONSUMED/HISTORICAL`; 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, former 02C-waits-02B predecessor route is CONSUMED/HISTORICAL; 02F-D remains blocked/nonblocking. 02E/generation/migration/runtime/service/API/UI/release remain blocked with launcher `NOT_READY` and executables `NOT_RUN`. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**
> **Current Rule lifecycle (2026-08-14):** standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole added plan blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); exact Contract high, Security xhigh and Lifecycle high reviews PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 02F and Rule-plan authoring/review is `CONSUMED/HISTORICAL`. Rule Task 1 RED remains blocked until all accepted non-marker Phase-145 engine-port/typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, unordered. No Rule source/test/TDD/schema/runtime/product/release authority; launcher `NOT_READY`, executable evidence `NOT_RUN`, Phase 148 blocked.

> **Current 144-03B rule gate:** Rule-plan acceptance is complete and authorizes neither source nor TDD. 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, with Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; former lifecycle is `CONSUMED/HISTORICAL`. The future implementation boundary remains exactly `rule_compiler.go` plus `rule_compiler_test.go`; Rule Task 1 is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase-145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order. All other blockers remain in force.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text and Rule-plan author/adjudication are `CONSUMED/HISTORICAL`; no implementation authority. Rule Task 1 remains blocked on all Phase-145 carrier/G145-5B/G145-5C gates and launcher readiness; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.

> **02B precedence (2026-08-15):** `947e2a27650ff1c6db60af9ce13bb33b6bf19cb0` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact13 blobs are in `IMPLEMENTATION-144-02B.md`. Tests first; Contract/high, Security/xhigh and Lifecycle/high precommit and exact-SHA PASS C/I/M `0/0/0`. Earlier `SOURCE_AUTHORITY_ACCEPTED`, source-unaccepted and 02C-waits-02B-source clauses are `CONSUMED/HISTORICAL`. A separate narrow root-authored/adjudicated 02C docs/source-authority bridge is required; 02C source is not automatically authorized. Other blockers remain unchanged. **02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same paragraph is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.**

## 144-02C historical authority-bridge candidate lifecycle

This is a `CONSUMED/HISTORICAL` candidate record only. The authoritative state is the accepted `e565523f589fbd7536f6f8e4df7effd39469c1af` source receipt in `IMPLEMENTATION-144-02C.md`; it selects no subsequent packet. All non-02C blockers remain unchanged.

## 144-04F sync-page policy source-authority receipt

The candidate `4e608e890ec12aeed8de965f9649ae503260b94e` (parent
`c53a5fe60e1283cab58be5efaec7ffa131145b31`, tree
`ed7e2bdab6a44254a7c677603ad0afc298aafed8`, subject `docs(mail): propose sync policy authority`)
is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` for its exact9 tree blobs after immutable
Contract, Security and Lifecycle reviews PASS C/I/M `0/0/0`; `dispatcher-enforced`, no fallback,
telemetry `UNKNOWN`. Candidate/precommit/commit-review/separate-acceptance-next clauses are
`CONSUMED/HISTORICAL` for that receipt. Its former immediate exact4 root-dispatch clause is now
`SUSPENDED / CONSUMED/HISTORICAL` for this correction route.
**ROOT_CORRECTION_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN:** the following remains only the future immutable canonical test-first manifest:
`backend/tests/mail/mailbox_events_test.go`,
`backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, then
`backend/internal/mail/mailbox_sync_policy.go`; no substitute or extra path. `mailbox_events.go` changes only the
three existing folder/label caps from 50 to 100 (100 pass/101 reject; no field, subject or version
change); all new ABI types are only in `mailbox_sync_policy.go`. Existing uncommitted exact4
source/test candidates remain `UNACCEPTED` and MUST NOT be committed until triple independent
precommit PASS, one atomic exact9 correction-candidate commit, triple immutable exact-SHA PASS,
and a separate docs-only acceptance sync. This correction is not accepted. The previously committed
144-04F authority acceptance sync is not a new acceptance claim. No other
source packet is selected. 02E, Ent generation/migration, services/API/UI/product/
phase/release, 02F-D, engine deletion and physical User/GDPR remain blocked/out of scope;
launcher `NOT_READY`, executables/runtime `NOT_RUN`.
