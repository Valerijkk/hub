# Phase 144 — executable packet contracts not split into standalone files

> **144-02B source-authority acceptance:** combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` is `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS`; Contract/high, Security/xhigh and Lifecycle/high corrected-child exact reviews PASS C/I/M `0/0/0`. Root may dispatch only the fully-qualified frozen test-first exact13 manifest in `DISPATCH-INDEX.md`; no other source is implied. Source implementation remains unaccepted and requires test-first authoring, independent review, atomic commit, exact-SHA review and status sync. Prior candidate/precommit/correction/failure-next clauses are `CONSUMED/HISTORICAL`; immutable `45fd3369` failure remains history. 02C waits accepted 02B source; 02F-D/02E/Ent generation/migration/runtime/services/API/UI/release BLOCKED; launcher `NOT_READY`, executables `NOT_RUN`; Calendar accepted/no-next and unrelated blockers unchanged.

> **144-02D source receipt:** `2242a4e7846944498eb5ed701ab8f47c986bd804` remains `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 02F-R source is accepted at `ce542cb5`; former 02B eligibility clause is `CONSUMED/HISTORICAL`, 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, 02C waits accepted 02B source, and 02F-D remains blocked. Launcher `NOT_READY`; executables/generation/migration/runtime `NOT_RUN`.

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` stay accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; no source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness block Rule Task 1/145-07B; launcher `NOT_READY`, executables `NOT_RUN`.

> **Current dependency truth:** 144-02D retains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 144-02F correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and Rule-plan authoring/review are `CONSUMED/HISTORICAL`; no packet gains source/runtime eligibility. Rule TDD/production awaits all accepted non-marker Phase-145 carrier/G145-5B/G145-5C and launcher readiness.
> **Exact14 acceptance precedence:** correction-scoped 144-02A/B/C/D documentation plus the 02F dependency/status trace is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Prior candidate/precommit/root-commit/exact-review/status-sync-next clauses are `CONSUMED/HISTORICAL`. Engine deletion-port and physical User/GDPR deletion-intent contracts remain separate `BLOCKED` / `NOT_DISPATCHABLE`; no source/schema/generation/migration/service/test/runtime authority.
> **Assembly rule:** dispatch the Common execution envelope plus exactly one packet or one

> **02F dispatch invariant:** reviewers must reject NULL-vulnerable receipt/deployment predicates, any target unique other than `(receipt_id,request_index)` plus `(receipt_id,target_type,target_id)`, a fictitious composite FK, a `message_bulk p=10` folder-exhaustion path, incomplete literal state CASEs, or a hash header missing `0x00||0x01`. The receipt oracle separates folder p10 terminal 409/30d from manual-only bulk p10 and requires same-key persisted-202 replay before the account-locked fresh-key `423 MAIL_ACCOUNT_BUSY` guard (no row/snapshot/expansion/I/O); it also requires stored-order length-exact nested hash frames and late-owner contention handling. The persistence contract is the normative byte/state/CAS source; this dispatch authorizes neither source/schema/generation/migration/service/runtime nor any executable while launcher is `NOT_READY` and executable evidence `NOT_RUN`.
> explicitly named variant below. Combining variants is invalid.
> **Runtime:** `gpt-5.6-terra` / `xhigh`, fallback none.
> **Status:** TERRA_CODE_ROUTE_AUTHORIZED; accepted native `dispatcher-enforced` provenance
> permits actual telemetry `UNKNOWN` when telemetry is unavailable. Any exposed mismatch,
> rejection, omitted model/effort or fallback is `BLOCKED`; dependencies are necessary but
> not sufficient: each packet requires its own explicit root authorization. Executable
> generation/runtime evidence remains launcher-blocked and `NOT_RUN`.

Standalone complete prompts exist for 144-01A, 144-03D, 144-04D, 144-04E, 144-05A,
144-05B and 144-06A. This file completes every other index row.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

> **144-02F-R accepted source receipt:** `ce542cb5de839a1fb5200295f710b0ae45bcdbe6` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; exact eight blobs in `IMPLEMENTATION-144-02F-R.md`. Tests first; Contract/Security/Lifecycle PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Authority-only/source-unaccepted and 02B-separate-authority/02C-predecessor route `CONSUMED/HISTORICAL`; 02B `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, implementation unaccepted, 02C waits accepted 02B source, 02F-D blocked; launcher `NOT_READY`, executables/generation/migration/runtime `NOT_RUN`.

## Common execution envelope

- Role Hub `codex-micro`; TDD external-package genuine unit first, then minimal code. One
  behavior, exact allowed files, no commit/push/PR. Root supplies exact clean HEAD.
- You are not alone: no revert; dirty overlap/missing frozen dependency/model mismatch =>
  `BLOCKED` before write. Russian exported comments; Go 1.25; stable safe errors; bounded
  collections; no new dependency/global state/wall clock/network/listener/real DB/broker.
- Never read/change protected `testframework/**`, `backend/testframework/**`, test-rebase
  audits/phases 128–141/TASK-193–234, `pro-qa-ai/**`, `.env.local`, `.git/index.lock`.
- Existing Resend alert files/routes/flat DTO remain unchanged unless a packet explicitly
  lists a narrow composition call. No raw body/MIME/key/credential/provider response in
  PostgreSQL, error, audit or event.
- Host tests/build/lint/validators/generation forbidden; direct Docker/Compose not fallback.
  Accepted launcher absent => `NOT_RUN`. Real DB/engine proofs route to Python handoff, not
  fake component Go tests in `backend/tests/**`.
- Bounds: depth/fan-out 0; two retries/root cause; 45m (60 service, 75 handler). Response in
  Russian with files/symbols/oracles/verification/risks/requested+actual model/source SHA.

## 144-01B — deterministic engine fake — STATIC_ACCEPTED / EXACT_SHA_PASS

> **Immutable receipt:** `0c064363b5ccb657a7c444802e94075c4ca743e5`; parent `48a7273d100c07788def88e2f1b8d69c7173580f`; tree `9c568a0c58c6bf7b8c3171cfd934760b839db623`; subject `feat(mail): add deterministic engine fake`; fake blob `12bcd921668e406f07a98f7f3b0618ad8d44e81d`; authored-first test blob `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f`. Terra/xhigh author; Spec Terra/high, Security Terra/xhigh, ABI Terra/high precommit and exact-SHA Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Previous Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`; launcher `NOT_READY`, executables `NOT_RUN`, and no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority follows.

- Dependency: accepted 01A plus, under stricter Task3 authority, applicable supported current-SHA
  Docker evidence. Until both close, `BLOCKED`. Allowed only:
  `backend/internal/mail/mailbox_fake.go`,
  `backend/tests/mail/mailbox_fake_test.go`.
- Behavior: mutex-protected in-memory accounts/folders/messages/blobs/drafts/submissions,
  fixed Sieve state and spam-training outcomes;
  injected ids/clock; scripted typed errors and lost responses; strict account scoping;
  `ReconcileDraft` and `ReconcileSubmission` never mutate/retry. Script bounded projection-only
  incremental pages plus base-anchored enumeration pages/checkpoints, invalid cursor,
  anchor restart and transient failure without leaking cursor values.
- Oracles: compile conformance; same stable submit key one visible result; definitive
  absent vs indeterminate; sent returns two bounded non-UUID opaque ids; account/message-
  keyed `GetMessage` can script exact and mismatched projections for submission-finalization
  oracles; draft expected-version conflict; changes cursor pages; cross-
  account ids zero disclosure/mutation; deterministic snapshots; no secret/body in errors.

## 144-02A1/A2/A3 — typed DNS then domain/policy then account/identity schemas

- 144-02A1/A2/A3 remain `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. 02F-R source is accepted at `ce542cb5`. Former 02B eligibility/not-auto-dispatched clause is `CONSUMED/HISTORICAL`; 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch only, 02C waits accepted 02B source; 02F-D and 02E remain blocked. Launcher `NOT_READY`; executables/generation/migration/runtime `NOT_RUN`.

- Representation dependency is satisfied narrowly by the docs-only `STATIC_CONTRACT_READY`
  acceptance of reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931`: Contract/Spec
  Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency Terra/high PASS C/I/M
  `0/0/0`, dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. It authorizes
  neither source/TDD nor generation/migration/runtime. Founder migration authority, supported
  task-owned launcher, sequential explicit source dispatch and source-level review remain gates. Execute
  sequentially from clean accepted commits:
  - `144-02A1`: only `backend/internal/mail/domain_dns.go` and
    `backend/tests/mail/domain_dns_test.go`;
  - `144-02A2` is accepted/historical; do not redispatch its four exact paths;
  - `144-02A3` is accepted/historical; do not redispatch its exact seven paths.
- Behavior: exact spec §3 fields/enums/bounds; unique org+user, opaque engine account;
  nullable engine id, immutable account provisioning key/domain/local part, desired/deployed
  account generations/hashes, paired account-control lease/retry/error; identity pending/
  active/error/deleting; domain stable key/state/desired+deployed generation/hash/paired lease/
  retry/public DNS fields; projection state, incremental/rebuild cursor, sync/mutation epoch, separate paired sync
  and mutation leases,
  retry/error, immutable injected activation cutoff committed before first worker eligibility
  and freshness fields; normalized identity/domain uniqueness;
  partial one-default indexes; one policy/org; no password/admin/DKIM private secret.
  Preserve all existing edges byte-semantically.
- Oracles: typed DNS value/JSON bounds first; schema field/index/edge table; each lease token/until pair and non-negative
  epochs/attempts; bounded opaque cursor/error; active/provisioning and projection enums; exact
  pre/post existing-edge inventory; no generated/migration edit.

## 144-02B — folder/thread/message/membership/label schemas

> Current state: 144-02A3 and 02F-R source are `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. Former A3/02F-R dispatch routes are `CONSUMED/HISTORICAL`. Combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` makes 02B `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS`; root may dispatch only its exact13 manifest while implementation remains unaccepted. 02C waits accepted 02B source; 02F-D and 02E remain blocked. Launcher `NOT_READY`, executables `NOT_RUN`.

- Dependency: reviewed 02A. Allowed only new schema files
  `mailfolder.go`, `mailthread.go`, `mailmessage.go`, `mailmessagefolder.go`, `maillabel.go`,
  `mailmessagelabel.go`.
- Behavior: same-account relations; engine folder/message ids unique per account;
  folder engine id nullable only for a creating intent; folder control state/receipt/version;
  message-folder join unique; non-negative rebuild epoch mark on folder/thread/message/join;
  server-derived projection-origin enum and nullable immutable received-event UUID on
  message; required bounded engine version from the same projection as its message id;
  auxiliary state/flags/freshness rebuildable;
  label has opaque random engine keyword; unique same-account message-label join carries
  manual/engine_rule source bits with at least one true; labels/pin/verified calendar/CRM relations Hub-owned; bounded headers/snippet only, never
  raw body/MIME; custom root/child sibling uniqueness with nullable-parent partial indexes.
- Oracles: exact indexes/predicates; no folder cycle claim at schema alone; soft-delete
  fields; BCC projection policy; no cross-account edge accepted by service contract.

## 144-02C — object/attachment/draft/submission schemas

> Current state: combined 144-02A/B/C/D correction with 02F dependency/status trace is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc`; candidate/precommit status is `CONSUMED/HISTORICAL` and this dispatch remains unauthorized. Relevant scope includes 02C Account->MailObject `NO ACTION` plus durable deleted-object lease/schedule lifecycle. Object authority is owner/account/object-key/upload-state/hash/expiry; attachment is sole scan authority.

- Dependency: reviewed 02A. Allowed only new schema files
  `mailobject.go`, `mailattachment.go`, `maildraft.go`, `mailsubmission.go`.
- Behavior: mail object owner/account/object-key/upload-state/hash/expiry; attachment is the sole scan authority and owns backing metadata;
  draft opaque engine id/version/revision/states/hashes, no content and at most 50 locked
  scalar attachment associations; submission hashed key, request hash, stable engine
  key/states, separate opaque engine result ids and scalar Hub result UUID. Express XOR/check through selected
  migration only when authority exists; service invariants remain explicit.
- Oracles: raw key/body absent; unique account+hash and account+engine ids; 51st draft
  association rejected atomically; exact result-id/nullability matrix; draft submitting
  state; no project/FileMetadata edge; expiry/audit; no generated/migration edits.

## 144-02D — settings/signature/rule/contact schemas

> Current state: the 02D RecentContact/User lifecycle slice participates in the docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` exact14 receipt `576f19231e37f635d5b7187a1ada121ed14ff8dc`; candidate/precommit status is `CONSUMED/HISTORICAL` and no dispatch is authorized. The canonical complete lock set is the deduplicated union of (a) owned or otherwise lifecycle-affected accounts including the operation target, (b) accounts containing an exact `RecentContact.user_id` link, and (c) accounts containing any physically persisted `MailObject` uploaded by the User. Physical User deletion and every cross-domain GDPR mutation are blocked until their separate accepted durable intent/job/route contract; this is not async owned-account-deletion authority. Target-organization mismatch remains a blocker. Unrelated accepted 02D facts remain authoritative.

- Dependency: accepted 02A representation and independently accepted 02D contract review.
  Corrective docs `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition
  `e563a3d3be8898371e5c8a9861c58f11b72167bc` are docs-only `STATIC_ACCEPTED` /
  `EXACT_SHA_PASS` after three post-transition exact PASS reviews C/I/M `0/0/0`. This packet
  remains `IMPLEMENTATION_BLOCKED`; no
  source authorization. Allowed only new schema files
  `mailusersettings.go`, `mailsignature.go`, `mailrule.go`, `recentcontact.go`.
- Behavior: one settings/account; required same-account default identity for active
  account, nullable signature; one default signature; bounded sanitized config; typed rule
  JSON/version/soft delete but no raw Sieve or engine script id; RecentContact has canonical
  account/email uniqueness, <=1,000 null-user external rows/account, linked-row exemption,
  User-first then one ascending raw-UUID lock/re-read of the complete deduplicated three-family
  owned/lifecycle-account + exact contact-link + physically persisted uploader-object union
  for user mutation, user-first link/manual-unlink, account-only external admission and standalone
  account deletion, exact per-account LRU cap restoration, safe exact-content display-name validation,
  and no public persistence-history route.
- Oracles: ownership/default partial indexes; signature/rule bounds; active Hub user rank
  not encoded as persistence; no message body authority or secret.

## 144-02F — mutation receipt and rule-deployment schemas

> **02F-R precedence (2026-08-15):** source is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `ce542cb5de839a1fb5200295f710b0ae45bcdbe6`; authority precursor `a877848e` is `CONSUMED/HISTORICAL`. Combined `45fd3369` + `4f3713590fee455fae1c94f2464d7de548e73e76` makes 02B `SOURCE_AUTHORITY_ACCEPTED` / `EXACT_SHA_PASS` for root exact13 dispatch only; source implementation remains unaccepted. 02C waits accepted 02B source; 02F-D remains blocked and nonblocking. 02E/generation/migration/runtime/service/API/UI/release and launcher blockers remain current.

> Current trace: 02F receipt/deployment normative `STATIC_ACCEPTED` / `EXACT_SHA_PASS` facts remain unchanged. Its broad 02D-unchanged/dependent writer-replay clause is dependency-traced by the docs/readiness-only exact14 receipt `576f19231e37f635d5b7187a1ada121ed14ff8dc`; candidate/precommit status is `CONSUMED/HISTORICAL` and it authorizes no source/schema/runtime work.

- **Legacy combined 02F source route — `CONSUMED/HISTORICAL`:** the former broad manifest that
  combined receipt/target/member with `mailruledeployment.go` and unspecified narrow inverses is
  not an active source route and authorizes no path. Its normative receipt/deployment facts remain
  only as the accepted docs/readiness contract.
- **Current routing:** 02F-R source is `STATIC_ACCEPTED` at `ce542cb5`; exact blobs are recorded
  in `IMPLEMENTATION-144-02F-R.md`. 02B is `SOURCE_AUTHORITY_ACCEPTED` for root exact13 dispatch
  only; source implementation remains unaccepted, and 02C waits accepted 02B source. 02F-D is separately `BLOCKED` /
  `NOT_AUTHORIZED`, has no source manifest, and is nonblocking.
- Behavior: receipt organization/actor/nullable account/operation/key source (`client` or
  `internal`)/key hash/request hash/positive fingerprint version/state/target/status/safe
  code; no response body/raw key/content/address. Client full replay metadata expires 30 days
  after first terminal commit while its compact key/request hash tombstone survives account/
  organization lifetime; internal folder update/delete receipts may hard-delete only after
  the same terminal recovery window. One rule deployment per account with non-negative desired/
  attempted/deployed generations, bounded hashes, pending/applying/ready/unknown/error, retry
  fields and paired lease. Claim persists immutable attempted tuple before I/O; only confirmed
  reconciliation copies it to deployed, and ready requires equality with current desired.
  `attempt_count` is exactly `0..10`; nullable internal-only `internal_terminal_code` is only
  `MAIL_RULE_DEPLOYMENT_EXHAUSTED` and never leaves persistence. Per immutable attempted tuple/
  current desired, `p<10` persists attempted tuple+lease/fence+`p+1` before at most one
  read/install cycle (`p=9` tenth); p=10 is no-I/O after expiry. An unchanged desired then
  terminalizes error with null public error/retry/lease and retained attempted tuple; an advanced
  desired cannot be terminalized by the old tuple and begins budget zero after old outcome/expiry.
  Late owner deploy CAS includes account token/epoch; tenth confirmed result may complete while
  crash/unknown exhausts with no eleventh cycle.
- Oracles: exact composite uniqueness and source-specific delete behavior; client UUID and
  internal UUID modes cannot be confused; target optionality; completed
  replay vs tombstone-only reject distinguishable; in-progress/unknown never compact;
  terminal/replay timestamps coherent; generation monotonic constraints; paired
  lease fields; no raw Sieve/provider/error/body; no generated/migration edit.
  Bulk targets preserve unique request index/type/id, bounded safe result counts/training;
  member UUID snapshots total at most 200/receipt, have no engine/content edge and compact
  with full replay metadata.

## 144-02F-D — rule-deployment schema route

> **Status:** `BLOCKED` / `NOT_AUTHORIZED` / `RUNTIME_NOT_RUN`. This is separate from 02F-R,
> has no current source manifest, and is nonblocking for 02B and 02C.

- The retained deployment tuple/lease/hash/state contract above is normative documentation only.
  It authorizes no schema, test, generated Ent, migration, service, worker, deployment, or runtime
  work. A future 02F-D packet requires its own separately accepted authority, exact manifest,
  test-first/static review, commit, and exact-SHA acceptance.

## 144-02E — Ent generation inspection

- Dependencies: accepted 02A–D/F, founder migration choice and launcher. Allowed: no manual
  tracked source; task-owned ignored generated output/evidence only.
- Behavior/oracles: repository Ent generator inside bounded launcher; exact SHA/tool;
  inspect all schema types/edges/index metadata and compile; second generation empty diff;
  exact cleanup. No migration invention/stage. Missing gate => BLOCKED/NOT_RUN.

## 144-03A-Address — address policy

- Dependency: 01A. Allowed only `backend/internal/mail/address_policy.go`,
  `backend/tests/mail/address_policy_test.go`.
- Behavior/oracles: parse normalized mailbox/domain/display name with IDNA/email bounds;
  case/domain policy and duplicate canonical set; CRLF/control/header-injection rejection;
  reserved synthetic fixtures; no DNS/network; stable safe errors.

## 144-03A-Thread — thread identity

> **Accepted exact boundary:** Thread `9903f4df`, Folder `f8cb7643`/docs `ca4e2854`, Rule docs `f1995465`, 02D docs `7d29d08c` + transition `e563a3d3`, and 02F correction `fcb5b84f` + transition `2f4381a2` are static/docs-readiness accepted. Rule plan is also docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761` (parent `77eb`, tree `1b554d`, blob `b86c0423`), exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. The former 02F closure, transition route and Rule-plan authoring/review are `CONSUMED/HISTORICAL`; Rule TDD/production awaits all accepted non-marker Phase-145 carrier/G145-5B/G145-5C and launcher readiness.

- Dependency: 03A-Address is `STATIC_ACCEPTED` at
  `41564d003946ddb78df18efca166ca86e4055405`, but dependencies are necessary and not
  sufficient. Status: `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
  `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS`; independent
  Spec/API/Persistence, Security and Lifecycle/Evidence reviews PASS with C/I/M `0/0/0`.
  The accepted contract freezes the exact callable,
  strict bytewise Message-ID lookup, bounds/preflight/purity, normalized_subject_v1, raw
  ThreadKey hash framing, persistence/OpenAPI participant representation, exact typed
  `ThreadPolicyError` Error/Unwrap mapping and three literal SHA-256 oracles in
  `MAIL-SPEC.md`; it does not accept implementation.
  The prior dispatcher-enforced `gpt-5.6-terra`/`xhigh`, no-fallback static-TDD authorization
  for `backend/internal/mail/threading.go` and `backend/tests/mail/threading_test.go` is
  **CONSUMED/HISTORICAL**. Thread is now narrowly `STATIC_ACCEPTED` at `9903f4df` /
  `EXACT_SHA_PASS`; it is not remaining code authorization. The accepted contract includes its callable
  API, exact Message-ID grammar and resolution semantics,
  participant and bounds policy, Unicode and prefix policy, raw 32-byte versioned hash
  encoding/preimage, safe errors, and literal oracle vectors. The required external error
  oracle asserts the exact literals and typed `errors.Is` mapping: nil receiver or unknown
  kind yields `mail thread policy error`/nil unwrap; invalid yields `mail thread invalid` /
  `Unwrap() == ErrThreadInvalid`; too-large yields `mail thread exceeds bounds` /
  `Unwrap() == ErrThreadTooLarge`;
  no case leaks raw input/detail or matches a noncorresponding sentinel. Executable
  evidence remains `NOT_RUN`.

## 144-03B-Folder — folder tree policy

- Accepted exact two-path source+oracle boundary: `f8cb7643986fc591cae16764575337b72ef3a692` /
  `EXACT_SHA_PASS`, static-only (source `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle
  `efe0db0071cb22ce4b026c635157f0fb08885d64`, C/I/M `0/0/0`). Folder docs are accepted
  static-only at `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`. 02D is docs-only accepted and 02F
  correction `fcb5b84f` plus transition `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` /
  `EXACT_SHA_PASS` after all three Terra routes PASS C/I/M `0/0/0`. The former 02F transition route
  and Rule-plan authoring/review are `CONSUMED/HISTORICAL`. All accepted non-marker Phase-145 engine-port/
  typed-IR/script carrier/G145-5B/G145-5C plus launcher readiness remain before Rule TDD/production.
- Static behavior/oracle summary: Hub admission has at most 100 live non-deleted control slots
  including system/reserved states, separate from <=100 engine-backed enumeration; depth ≤2,
  no cycle/self-parent/cross-account parent; system role cannot rename/reparent/delete and
  create is custom; UUID-only policy receives projection-mapped Hub rows, normalized sibling
  collision and deterministic ordering; no engine call or executable evidence acceptance.

## 144-03B-Rule — typed rule compiler

- Future frozen-boundary proposal only; accepted 03B-Folder is necessary but not sufficient.
  The callable/error/escaping/oracle contract and plan must be root-frozen and independently
  accepted before any RED oracle, production/test authoring or dispatch; this section does not
  authorize Rule code.
- A future accepted boundary may name only `rule_compiler.go` and
  `backend/tests/mail/rule_compiler_test.go`, with frozen conjunction conditions/actions, escaped
  literals, engine-folder target resolution, native mark-read, server-resolved opaque engine
  keywords (never name/UUID), deterministic priority/order, and rejection of raw Sieve/new
  action/nesting/bounds and command-injection fixtures.

## 144-03C-Send — rate/recipient/loop policy

- Dependency: 01A. Allowed only `send_policy.go`,
  `backend/tests/mail/send_policy_test.go`.
- Behavior/oracles: message/recipient/rate bounds, external toggle, authorized identity;
  null sender/Auto-Submitted/precedence/loop headers suppress auto-response; injected clock
  and token-bucket state; explicit retry hint; no network/global clock.

## 144-03C-Attachment — attachment policy

- Dependency: accepted 03C-Send. Allowed only `attachment_policy.go`,
  `backend/tests/mail/attachment_policy_test.go`.
- Behavior/oracles: normalized filename/media/inline CID/size/count; clean-only send/read;
  pending/rejected/error fail closed; unsafe SVG/executable/double-extension policy;
  deterministic no I/O; no project authorization.

## 144-04A — account resolver/repository base

- Dependencies: 01B/02E. Allowed only `mailbox_repository.go`,
  `mailbox_ent_repository.go`, `mailbox_service.go`,
  `backend/tests/mail/mailbox_owner_test.go`.
- Behavior: authenticated user -> active User organization -> active MailAccount -> opaque
  engine account. Public commands never carry org/account. Repository interfaces expose
  `BeginEngineCommand`, `FinishEngineCommand`, expired-command reclaim, bounded sync claim/
  renew/release and page/rebuild transaction primitives for later services. Begin acquires
  one expiring account mutation lease and increments the epoch; finish verifies the token,
  clears it and increments again. Sync capture refuses a present mutation lease; reclaim
  clears an expired lease and increments before any engine fetch.
- Oracles: cross/suspended/deleted/no account zero engine; one command per account; active-
  lease busy; begin/finish/reclaim epoch edges; late owner cannot finalize; no client id;
  transaction lease contract; safe non-disclosure. Real DB constraints remain Python integration.

## 144-04B-Folder — folder commands

- Dependencies: 03B-Folder/04A/01B. Allowed only `folder_service.go`,
  `backend/tests/mail/folder_service_test.go`.
- Behavior/oracles: list/create/rename/delete through resolved account + engine; policy
  first. Existing client receipt replay is resolved before a new command claim. For an absent
  key, active lease contention returns 423 with no receipt/control mutation; one transaction
  under the acquired lease rechecks/reserves 04O, inserts a unique creating row before I/O
  and enforces normalized parent/name absence including pending rows. Name/tree/cap failure is a
  terminal safe receipt without a folder row or engine call;
  projected-count precheck is advisory and adapter/engine atomically enforces the
  100-folder cap with `MAIL_FOLDER_LIMIT_EXCEEDED`; `BeginEngineCommand` commits before the
  first engine call and finish clears the exact lease plus advances the second epoch edge;
  successful engine mutation reconciles projection/state. PATCH/DELETE first resolve the target
  nondisclosingly: missing is 404, system is `MAIL_FOLDER_SYSTEM_ROLE`, and only a resolved
  custom PATCH target with supplied `parent_id` is temporarily 409
  `MAIL_FOLDER_TREE_CONFLICT` before receipt/engine I/O until accepted reparent port work.
  PATCH/DELETE then require If-Match and generate an internal injected UUID v4/v7 receipt under
  the acquired command lease, increment
  version when atomically persisting desired state/receipt link, and reject a new mutation
  while that folder is creating/unknown/deleting. Unknown create lists exact-account
  folders only when the injected adapter capability proves Hub-exclusive custom-folder
  writing, and accepts only one exact non-system parent/name match before controlled
  same-receipt retry; rename/delete read current state before retry under the same gate.
  False/unknown exclusivity, a system-role candidate, multiple/mismatch or alternate writer
  stays degraded with zero recovery mutation, never guessed; pending state remains
  owner-visible with zero counters and due recovery. System
  roles protected; cross-account zero engine. Hub-only label mutation
  never increments the engine fence.

## 144-04B-Label — label commands

- Dependency: accepted 04A. Allowed only `label_service.go`,
  `backend/tests/mail/label_service_test.go`.
- Behavior/oracles: Hub-transactional label metadata CRUD, normalized uniqueness and create
  reserve/replay through 04O. Message assignment is owned only by receipt-backed 04C bulk,
  not a second replace route. Label delete rejects active rule references and cannot silently
  orphan a manual/engine provenance join. PATCH/DELETE require optimistic version; stale or active-rule
  delete conflicts mutate no label, join, marker or rule. Create/update normalized collision,
  stale version and active-rule delete map to their exact label conflict code. Thread
  aggregation derived.

## 144-04C — message projection/read/bulk

- Dependencies: 03A-Thread/03B-Folder/03D/04A/04O/01B/02E. Allowed only
  `message_service.go`, `message_projection.go`,
  `backend/tests/mail/message_service_test.go`.
- Behavior: owner filter before engine hydrate; opaque cursor; engine authoritative
  folders/flags, Hub labels/pin; typed message-or-thread bulk targets expand owner-scoped and
  deduplicate before the 200-message cap; per-target outcomes; projection freshness explicit.
  Every engine move/flag/delete batch owns one account command lease with begin/end epoch
  edges around I/O. A competing command returns the stable retryable busy result.
  A successful mutation that changes the durable projection writes one complete 03D
  `mail.message.changed.v1` envelope through the supplied transaction outbox seam.
- `junk`/`not_junk` use the explicit spam/ham classifier-feedback port and report training
  outcome separately from folder/flag mutation. The bulk request reserves/replays one 04O
  receipt so retries cannot double-train. Each message feedback command derives a stable
  receipt/message id; without an accepted engine reconcile/idempotency primitive,
  indeterminate training is terminal unknown and never called again.
- Message/thread targets are expanded from one owner-scoped consistent snapshot, deduplicated
  and persisted as ordered receipt target/member rows before engine I/O. More than 200
  distinct messages terminally records/replays safe `MAIL_BULK_EXPANSION_LIMIT` without
  member rows, mutation or classifier I/O. Pin/unpin requires thread; move and label actions
  require their typed target ids. Per-target results retain the request target type and
  exact stored succeeded/failed/training counts. Replay never re-expands current threads;
  state mutation may reconcile only the captured members.
- `add_label` and `remove_label` are the only public message-label mutation path, even for a
  single message. Add changes the manual source; remove clears manual and, if `engine_rule`
  exists, uses the fenced opaque keyword command/reconciliation before clearing that source.
  Unknown/deleted markers are ignored, never remapped to a future label.
- Oracles: no cross-mailbox hit/count/snippet/timing hydrate; flag/move failure reconcile;
  partial batch; BCC owner-sent only; body never repository; duplicate change idempotent;
  rollback writes neither projection change nor envelope.

## 144-04J — account/identity desired-state service

- Dependencies: 01B/02A/E. Allowed only
  `backend/internal/mail/mailbox_account_control_service.go`,
  `backend/tests/mail/mailbox_account_control_service_test.go`, and narrow account-control
  repository signatures in accepted `mailbox_repository.go`.
- Behavior: stable UUID-ordered eligible active-user scan only when engine/policy and a
  verified enabled default domain are ready. In one transaction reserve the login-email
  local part when valid/unclaimed, else `u-` + first 12 lowercase UUID hex, expanding to the
  full 32 UUID hex only on a reserved short-prefix collision. Both fallback shapes are
  alias-reserved; any further collision is typed failure, never random suffixing. Persist
  immutable domain/local part/provisioning key/event cutoff plus primary identity. A claimed step calls
  idempotent `EnsureAccount` outside DB with 30-second timeout and commits opaque account/
  identity ids + deployed generation/active only under the exact live lease. Admin alias
  create/delete and Hub user activation/deactivation atomically change the exact desired
  enabled/identity set, increment generation/hash and schedule reconciliation; pending/
  deleting aliases are not public From choices. Lost response repeats the same key/generation.
- Oracles: disabled/missing-domain/suspended/collision zero engine; deterministic fallback;
  two replicas one claim; stale generation/expired owner cannot activate/suspend aliases;
  deactivation disables engine and reactivation reuses identity; partial result safely
  retries; no password/provider text; public status redacts reservation/lease/engine id.

## 144-04K — bounded account-control worker

- Dependency: accepted 04J. Allowed only
  `backend/internal/mail/mailbox_account_control_worker.go`,
  `backend/tests/mail/mailbox_account_control_worker_test.go`.
- One context-owned scheduler scans at most 50 candidates per second, processes at most four
  concurrently and one account step per claim. Lease 60 seconds, I/O 30 seconds and capped
  exponential backoff 1 second..5 minutes are injected. Cancellation stops claims, cancels
  I/O and waits boundedly; disabled configuration starts nothing. Oracles cover due order,
  new/existing users, alias generations, deactivate/reactivate, two replicas, renewal/
  expiry, panic isolation and no leaked timers.

## 144-04L-Settings — user settings service

- Dependencies: 02D/E + 04A. Allowed only `mailbox_settings_service.go`,
  `backend/tests/mail/mailbox_settings_service_test.go`.
- Behavior/oracles: active owner only; same-account send-allowed default identity and
  nullable signature; versioned patch; blocked/trusted-senders and rich/plain enums; no
  provisioning/error resource, engine call, credential or cross-account disclosure.

## 144-04L-Signatures — signature service

- Dependencies: accepted 04L-Settings + 05D-Content sanitizer port. Allowed only
  `mailbox_signature_service.go`, `backend/tests/mail/mailbox_signature_service_test.go`.
- Behavior/oracles: bounded CRUD/default and optimistic version; create uses 04O receipt;
  sanitizer accepts rich or plain format and derives canonical plain server-side; hostile
  rich falls back/rejects by typed policy; no client rich/plain pair trust or raw render.
  `is_default` mutates only the locked settings pointer: false clears only self if selected,
  true selects self, delete clears before soft delete, and settings version changes iff pointer changes.

## 144-04M — canonical rule service

- Dependencies: 02D/F/E + 03B-Rule + 04A/O. Allowed only
  `mailbox_rule_service.go`, `backend/tests/mail/mailbox_rule_service_test.go`.
- Behavior: owner-scoped typed CRUD/soft delete. Each transaction validates all targets,
  applies optimistic version, deterministically orders the complete active set, increments
  desired generation and stores its hash/deployment pending. Create uses 04O receipt.
  The account-deleting guard returns stable 409 before receipt, transaction or I/O. Otherwise,
  it performs no handler engine call and no transient lease/423 mapping. Response includes
  honest deployment state.
- Oracles: concurrent generation monotonic; same key replay; changed key conflict; stale
  If-Match zero change; delete leaves deployable tombstone until newer generation ready;
  no raw Sieve/script id/provider text.

## 144-04N — rule deployment worker

- Dependencies: accepted 04M + 01B. Allowed only `mailbox_rule_worker.go`,
  `backend/tests/mail/mailbox_rule_worker_test.go`.
- Bounded context-owned worker claims at most 20 due accounts per second and two concurrent
  deployments. It captures and persists immutable attempted generation/hash before I/O, acquires the ordinary account mutation lease,
  compiles the whole set, and manages one fixed-name Hub script with marker. `unknown` first
  calls `GetSieve`; matching marker finalizes without upload, mismatch stays/retries safely.
  Only confirmed reconciliation may copy captured attempted data to deployed; newer generation remains pending. External
  fixed-script drift is degraded and repaired under founder default 26.
- Oracles: lost response, crash before/after upload, stale generation, active command busy,
  lease expiry/late owner, disabled capability, retry/backoff/cancel/panic; no raw script in
  DB/log/error and no overlapping local step.

## 144-04O — mail mutation receipt service

- Dependencies: 02F/E + 04A. Allowed only `mailbox_mutation_receipt.go`,
  `backend/tests/mail/mailbox_mutation_receipt_test.go`.
- Behavior/oracles: canonical scope+operation+request hashing; client reserve/replay/conflict;
  internal mode is callable only by folder update/delete after successful lease/version/state
  checks and generates an injected UUID v4/v7 that is never an HTTP key;
  response reconstructed from current owned target, never stored as JSON; in-progress/
  unknown owned takeover requires operation-specific reconcile; 30-day client full replay
  cleanup runs only after terminal replay-until and compacts to lifetime key/request
  tombstone, while an internal terminal folder receipt may then hard-delete because it has
  no public replay;
  stuck nonterminal remains safely visible/reconcilable; raw key/body/address absent from persistence,
  errors and audit. Presign reconstruction permits a fresh URL only for the same pending,
  unexpired object; terminal/expired targets are unavailable and allocate nothing. Cross-
  actor/account key reuse is isolated. Reject non-v4/v7 UUID keys before persistence;
  store a length-delimited scoped key SHA-256 and raw-keyed HMAC-SHA-256 canonical request
  fingerprint. V1 uses tagged length-delimited typed fields, normalized scalars, no maps and
  preserves array order; version mismatch never guesses. Raw key never leaves request scope
  after reservation or enters persistence/error/audit/log values.
  Recovery defines `prior_attempts p`: for `0<=p<10` it claims receipt plus account lease/epoch,
  persists `p+1`, then receives exactly one bounded I/O slot (`p=9` tenth); preclaim/poll/busy/
  pre-engine failure do not increment. A live tenth lease is no-op contention; `p=10` allows no
  I/O and post-expiry takeover terminalizes folder receipts as generic NFR-13
  `MAIL_RECONCILIATION_EXHAUSTED`/409 with leases cleared/end epoch advanced. Tenth completion is
  valid; tenth crash/unknown exhausts with no eleventh slot.
  Bulk reserve atomically persists bounded ordered target/member snapshot; terminal safe
  outcomes replay exactly even after target deletion, while compaction removes all children.

## 144-04P — domain desired-state service

- Dependencies: 01B/02A/E/04O. Allowed only `mailbox_domain_service.go`,
  `backend/tests/mail/mailbox_domain_service_test.go`, and narrow domain-control repository
  signatures in accepted `mailbox_repository.go`.
- Behavior/oracles: admin same-org create reserves stable provisioning key and generation/
  hash through 04O, returns pending with no engine/DNS mutation; update desired enabled/
  default only under version and verified/default invariants; delete marks deleting only
  when not default/referenced. Repeated/changed key, global normalized collision, stale
  version and cross-org are safe. DNS records are engine-projected public values only; no
  private DKIM/provider text.

## 144-04Q — bounded domain-control worker

- Dependency: accepted 04P. Allowed only `mailbox_domain_worker.go`,
  `backend/tests/mail/mailbox_domain_worker_test.go`.
- One context-owned scheduler scans at most 20 due domain intents per second, at most two
  concurrently. It captures generation/hash under a 60-second lease, calls `GetDomain`
  before retrying unknown and then idempotent `EnsureDomain` with a 30-second timeout. Only
  the exact live generation may store opaque engine id, bounded public required-DNS data,
  verification/DKIM/safe state or confirm deleting absence. It never changes DNS.
- Oracles: create/DNS-required/verified/error/delete; lost response; newer update wins;
  reference/default guard; two replicas/lease expiry/late owner/backoff/cancel/panic;
  disabled engine zero calls; no secret/private key/provider body/log value.

## 144-04F — pure mailbox sync-page policy

- Dependencies: 01B/03A-Address/03A-Thread/03D. Allowed only
  `backend/internal/mail/mailbox_sync_policy.go` and
  `backend/tests/mail/mailbox_sync_policy_test.go`.
- Accept only bounded normalized engine pages: at most 100 total folder snapshots, 100 message
  snapshots and 100 destroyed ids, unique per account, exact incremental from/next cursor
  or enumeration base/checkpoint metadata plus caller-supplied current sync epoch and
  immutable event cutoff. Produce deterministic upsert/tombstone/thread/count and
  complete-envelope intents without I/O, clock or ids of its own.
- A first-seen inbound-or-unknown message with engine arrival strictly after cutoff requests
  one `mail.received.v1`; enumeration history at/before cutoff and confirmed owner submission/
  Draft copies do not. Suppression requires a stable engine-submission/message mapping or
  accepted Hub submission relation; From/alias equality or mutable Sent membership alone
  is never proof. Existing messages emit `mail.message.changed.v1` only for the exact
  allowlisted projection delta; destroyed emits `changed_fields=[deleted]` with empty
  engine folders/flags while preserving current Hub label ids in the intent.
- Hydrated reserved keywords are mapped only by exact same-account `MailLabel.engine_keyword`
  into engine_rule join sources. Unknown markers are ignored. Public event label ids are the
  Hub join union; raw markers never enter event/search/API/log/error.
- Oracles: duplicate/mixed-account/malformed page rejected; deterministic order; References
  threading; BCC policy; first-seen boundary; spoofed-own-From vs confirmed submission/Sent;
  unknown provenance fails toward inbound; replay marker; no-op;
  changed field/flag enum; destroy; size/count/address bounds; no raw body/MIME/cursor/
  provider error in command/event/error.

## 144-04G — mailbox sync orchestration

- Dependencies: 04A/F and deterministic fake. Allowed only
  `backend/internal/mail/mailbox_sync_service.go`,
  `backend/tests/mail/mailbox_sync_service_test.go`, and the minimal new sync repository
  signatures in accepted `mailbox_repository.go`.
- One `Step(account)` claims/renews a bounded lease and captures exact cursor, sync epoch and
  mutation epoch and proves no mutation lease is present; engine I/O occurs outside the
  repository transaction. Incremental page
  apply succeeds only through one repository call carrying all captured guards/commands.
  Guard or new mutation-lease mismatch discards the entire page and retries the unchanged
  committed cursor. An expired mutation lease is atomically reclaimed with an epoch bump
  before capture; a late former command owner cannot finalize from its stale result.
- Typed invalid cursor enters degraded then resumable rebuilding without clearing rows:
  increment epoch, retain cutoff/old freshness, apply base-anchored enumeration pages/checkpoints,
  catch up through Changes from the base cursor, then request one atomic finalization.
  Typed anchor-restart starts a new epoch; transient error retains checkpoint and schedules
  bounded retry. Inject clock, event ids, lease ids and jitter; never log cursors/provider
  text or manufacture empty success.
- First bootstrap is eligible only after the account's immutable `event_cutoff_at` is
  committed. Event admission compares engine receipt/arrival time, never the spoofable
  message `Date` header or local observation time.
- Oracles: first bootstrap; incremental page/more-pages; rollback; stale lease/cursor/sync-
  epoch/mutation-epoch/active-mutation-lease; command begin/during/end races; expired
  command-lease reclaim and late finalizer; invalid cursor; crash/resume each rebuild phase; anchor restart;
  catch-up arrival; cutoff boundary; one received marker/event; transient/permanent account
  state; cancellation. No DB/broker/network in unit.

## 144-04H — atomic Ent sync adapter

- Dependency: accepted 02E/04G. Allowed only
  `backend/internal/mail/mailbox_ent_repository_sync.go`; no Go DB/component test.
- Claim/renew/release uses one expiring account lease. Incremental/rebuild page transaction
  locks MailAccount, verifies token/until/prior cursor/sync epoch/mutation epoch and absent
  mutation lease, upserts by
  account + opaque engine id while preserving stable Hub ids and Hub-owned labels/pin/
  verified calendar/CRM edges, marks the current rebuild epoch for every row seen during
  enumeration or catch-up, recomputes affected derived threads/
  counts, reserves first received-event id once, writes full envelopes and advances cursor/
  checkpoint/freshness atomically.
- Rebuild page never pre-clears or tombstones unseen rows. Finalization alone, after
  base-anchored enumeration plus catch-up, soft-deletes unseen engine-authoritative folder/message/
  membership/thread projections, emits required delete envelopes, installs final cursor,
  clears checkpoint/lease/error and marks ready. Retention/canonical link rules run before
  later hard delete. Real constraint/concurrency/rollback proof is a Python integration
  handoff.
- Incremental/rebuild upsert preserves folder desired fields, `control_state`,
  `control_receipt_id` and `version` while the folder is not `active`. Final cleanup excludes
  every pending/error control row, including a create intent with no engine id. An unresolved
  or ambiguous folder receipt keeps the account degraded; only exact receipt recovery may
  activate/delete its row and permit `ready` finalization.

## 144-04I — bounded sync worker lifecycle

- Dependencies: 04G/H. Allowed only `backend/internal/mail/mailbox_sync_worker.go` and
  `backend/tests/mail/mailbox_sync_worker_test.go`.
- One context-owned scheduler scans at most 50 due active accounts per second and runs at
  most four account steps concurrently; one step handles at most one bounded engine page.
  Default I/O timeout 30 seconds, lease 60 seconds and capped retry backoff 1 second to
  5 minutes are injected/config-bounded. No goroutine per mailbox, unbounded queue or
  overlapping local step. Multiple replicas rely on repository lease guards.
- Disabled/unavailable engine starts no worker or probe. Cancellation stops claims, cancels
  in-flight I/O, waits boundedly and leaves leases recoverable. Panic is isolated to the
  account step, recorded only as safe code and rescheduled; no cursor/provider/message data
  enters logs.
- Oracles: due ordering/batch/concurrency caps; two-worker lease race; active command blocks
  capture; expired command reclaim invalidates late work; sync renewal/expiry;
  transient backoff/reset; context cancel before/during page; panic isolation; no leaked
  goroutine/timer; disabled path zero calls. Wiring belongs to gated Phase 145 composition.

## 144-04R — bounded generic mutation recovery in the sync worker

- Dependencies: accepted 04B/C/I/O plus 04A repository. Allowed only
  `backend/internal/mail/mailbox_mutation_recovery.go`, a narrow sequential edit to
  `mailbox_sync_worker.go`, and `backend/tests/mail/mailbox_mutation_recovery_test.go`.
- The existing one scheduler chooses the oldest due sync-page or receipt-recovery work item;
  it does not start a fifth worker. One account step performs one item and never holds sync
  and mutation leases together. A recovery claim uses paired receipt lease then account
  mutation lease, with the same 60/30-second lease/I/O and capped backoff bounds.
- Folder recovery first checks the injected selected-adapter
  `ExclusiveCustomFolderWriter` capability. Only a proven true value permits exact account
  enumeration: one non-system desired match finalizes, none permits at most one controlled
  same-receipt create attempt, and rename/delete reconcile current vs persisted desired/
  prior state. False/unknown capability, a system-role candidate, alternate-writer signal,
  multiple or mismatch remains degraded and makes zero recovery mutation. Bulk uses only
  stored target/member snapshot; state mutations may read/reconcile those ids, while
  indeterminate training is terminal unknown unless an accepted adapter primitive proves
  idempotent reconciliation. Unknown operation/fingerprint version executes nothing.
- Finalization atomically updates receipt/target outcomes, folder control state/projection,
  required full envelope, receipt lease, account mutation lease and end epoch. Oracles cover
  crash/reclaim, two replicas, later thread change, target deletion, exclusive-writer true/
  false/unknown, system-role and ambiguous folder match,
  preservation through ordinary sync/rebuild/final cleanup, no sync or recovery starvation,
  cancel/panic/backoff and no raw key/content/provider data.

## 144-05C-Search — search seam

- Dependency: 04A/C/I + 05B. Allowed only `mailbox_search.go`,
  `backend/tests/mail/mailbox_search_test.go`.
- Behavior/oracles: typed search port/query; mandatory org/account filters embedded before
  query and again before hydrate; bounded cursor/filters/snippet; foreign hits discarded
  without leaking total/id; index document accepts only the 05E bounded sanitized body and
  normalized filename projection, never raw HTML/MIME/BCC/remote URL.
  Delete events remove owner documents; rebuilding/degraded state is reported and never
  relabelled fresh, while all hits remain owner-filtered during convergence. Search response
  state is the conservative composition of mailbox projection and owner search-consumer
  checkpoint; ready requires both and freshness is their older complete instant.

## 144-05D-Content — server body sanitizer

- Dependencies: 03C-Attachment/04A. Allowed only `mailbox_content.go`,
  `backend/tests/mail/mailbox_content_test.go`.
- Pure port/policy accepts explicit rich/plain format. Approved implementation removes
  script/event/form/meta/base/frame/object/embed/active SVG, unsafe URL schemes, CSS fetches
  and all remote fetches; derives bounded canonical plain text; remote images become opaque
  token intents bound to org/account/message/normalized URL/expiry. Without the founder-
  approved dependency it returns plain-only capability and never passes raw HTML through.
- Oracles: hostile HTML matrix, Unicode/size/depth/count bounds, deterministic plain,
  rewritten/removed images, no origin URL in safe errors/loggable DTO.

## 144-05D-Remote — remote-image proxy

- Dependency: accepted 05D-Content + existing `safehttp` primitives. Allowed only
  `mailbox_remote_image.go`, `backend/tests/mail/mailbox_remote_image_test.go`.
- Behavior/oracles: reject any path token outside exact unpadded base64url 32..1024 before
  decode/repository/network, then verify signed/encrypted token and owner/message/expiry before network;
  validate http(s) URL and resolved IP before every redirect/connect with injected safe
  dialer; block private/loopback/link-local/metadata/rebinding, credentials and non-image;
  max 3 redirects, 5 MiB, 10 seconds, no cookie/auth/referrer, restrictive response/cache
  headers. Browser gets Hub-controlled bytes with fixed private five-minute cache,
  `nosniff`, `no-referrer` and restrictive CSP; no redirect/cache URL. No arbitrary URL route, DNS/
  real network in unit, origin URL in logs/errors/search or cross-owner timing disclosure.

## 144-05E — owner search projection hydration consumer

- Dependencies: 05B + accepted 05D-Content/05C-Search. Allowed only
  `mailbox_search_projection.go`, `backend/tests/mail/mailbox_search_projection_test.go`.
- Behavior/oracles: admitted mail event resolves current owner/account then bounded engine
  hydrate; indexes normalized From/To/subject, at most 256 KiB sanitized UTF-8 plain body,
  at most 50 normalized 255-char filenames, allowed folder/label/flag/time metadata and no
  BCC/raw HTML/MIME/URL/bytes. Duplicate/update/delete idempotent; foreign/missing owner
  deletes/ignores safely; checkpoint commits only after index outcome and drives freshness.

## 144-06E — sensitive HTTP route metadata redaction

- Dependency: accepted 06A. Allowed only
  `backend/internal/platform/middleware/logging.go`,
  `backend/internal/audit/middleware.go`, and
  `backend/tests/mail/mail_sensitive_http_path_test.go`.
- Behavior/oracles: after `Next`, both application access logging and generic HTTP audit use
  the matched Fiber route template for every matched route and a fixed `unmatched` label for
  misses; query strings and concrete UUID/token segments never persist. Existing method,
  status, duration, request/user metadata and skip rules remain. The remote-image token,
  arbitrary 404 path and mail-search query are absent from captured slog/audit/panic values;
  named domain audit still carries its separate bounded target id. No request body/header,
  auth token or new async work is added.

## 144-05C-Contacts — contact ranking

- Dependency: 04A + 03A-Address. Allowed only `mailbox_contacts.go`,
  `backend/tests/mail/mailbox_contacts_test.go`.
- Behavior/oracles: active same-org Hub users first, then owner recent contacts; normalized
  dedupe; bounded query/result; disabled users removed; no address disclosure cross-org.

## 144-05C-Realtime — websocket/notification seam

- Dependencies: 05B + 04C/E. Allowed only `mailbox_realtime.go`,
  `backend/tests/mail/mailbox_realtime_test.go`.
- Behavior/oracles: internal event -> owner/session bounded invalidation/counter metadata;
  no body/recipient list; duplicate idempotent; unknown version ignored/error-routed; no
  direct websocket network in unit.

## 144-05C-Audit — named audit seam

- Dependency: 04A. Allowed only `mailbox_audit.go`,
  `backend/tests/mail/mailbox_audit_test.go`.
- Behavior/oracles: exact OpenAPI admin action enum plus bounded bulk summaries; actor/org/
  target/result/request id only; no recipient/body/key/secret; mutation audit participates
  in supplied transaction; generic HTTP audit is not substituted.

## 144-06B — mailbox admin service

- Dependencies: 05C-Audit/04A/02E. Allowed only `mailbox_admin_service.go`,
  `backend/tests/mail/mailbox_admin_service_test.go`.
- Behavior/oracles: admin principal + current org; policy get/update, domain list/create/
  update/delete desired-state and redacted health; domain create uses 04O receipt and no
  handler calls engine/DNS directly; normalized uniqueness/default safety; enabled/external
  preconditions; exact policy/domain conflict codes; global normalized collision is only
  `MAIL_DOMAIN_UNAVAILABLE`; exact audit intent; no secret/DKIM private material; non-admin
  zero mutate.

## 144-06C — mailbox admin handler/wiring

- Dependencies: 06A/B. Allowed only `mailbox_admin_handler.go`, its focused unit, and the
  narrow call in `backend/cmd/server/wiring_mailbox.go` needed to register it.
- Behavior/oracles: exact OpenAPI mailbox-settings/domain/health routes and x-hub audit;
  admin middleware; NFR-13; legacy `/admin/mail/settings` response/registration and
  `/broadcast` current state untouched; health redacted; no credential input/output.

## 144-06D — admin per-user aliases

- Dependencies: 04J/K/O + 06B/C + 05C-Audit. Allowed only
  `mailbox_alias_admin_service.go`, `mailbox_alias_admin_handler.go`, their focused units and
  the narrow route registration in `wiring_mailbox.go`.
- Behavior/oracles: admin/current org selects the user account server-side; list plus
  idempotent normalized alias create and non-primary delete change canonical identity state,
  increment account desired generation and return pending/deleting; verified enabled domain
  and global address uniqueness; the account-deleting guard returns stable 409 before receipt
  or I/O; otherwise no handler engine call. Transient live-lease 423 remains only where
  already contracted and is not mixed with deleting. Account worker performs exact Ensure
  reconciliation. Group/primary mutation forbidden; cross-org/non-admin zero
  disclosure; collision/reserved address is only `MAIL_ALIAS_ADDRESS_UNAVAILABLE`, while
  inactive account and non-mutable identity have distinct safe codes; no engine identity id/
  provider error/raw key. Exact OpenAPI 0.1.8/NFR-13.

## 144-07 — docs sync and independent static verification

- Dependencies: all accepted code packets. Allowed only affected current-truth docs named
  by `.cursor/rules/55-docs-sync.mdc` plus a Phase 144 verification artifact; no product/
  test/protected file.
- Behavior/oracles: implementation-to-OpenAPI/FR/TC matrix; exact diff proves legacy
  Resend untouched, bodies/keys absent from persistence, `mail.>` present once and webhook
  list unchanged; record current SHA and all Docker/test/build/generation statuses honestly.
  Must request independent Terra review; cannot mark runtime PASS from static review.

> **Local 01B clause reconciliation:** every older 144-01B Docker-blocked/do-not-start/dependency assertion in this Rule record is `CONSUMED/HISTORICAL` only. Rule, Sieve, Phase145, launcher and every 02E/Ent/generation/migration/service/API/UI blocker remain current.

## 144-03B-Rule-Compiler — accepted plan, no dispatch

> **Current Rule acceptance:** Rule docs remain `STATIC_ACCEPTED` at `f199546513b758ca7b95328b5ccd7680009b732e`; accepted standalone plan is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761` (parent `77eb`, tree `1b554d`, blob `b86c0423`), exact reviews PASS C/I/M `0/0/0`. 02D docs are accepted at `7d29d08c` + `e563a3d3`; 02F correction `fcb5b84f` plus transition `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after all three Terra routes PASS C/I/M `0/0/0`. The `55e36b7a` closure, former transition route and Rule-plan authoring/review are `CONSUMED/HISTORICAL`.

`CONTRACT-144-03B-RULE.md` is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; no source/oracle dispatch is authorized. Its standalone implementation plan is accepted at full `669ef761`, docs/plan-only, C/I/M `0/0/0`; former `55e36b7a` closure, transition route and Rule-plan authoring/review are `CONSUMED/HISTORICAL`. All accepted non-marker Phase-145 engine-port/typed-IR/script carrier, G145-5B, G145-5C and launcher readiness precede Rule TDD/production. Launcher remains `NOT_READY`, executable evidence `NOT_RUN`, Phase 148 blocked, and 144-01B remains blocked on applicable supported current-SHA Docker evidence under Task3.
> **Rule-plan acceptance receipt:** standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at full `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract high/Security xhigh/Lifecycle high reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan authoring/review are `CONSUMED/HISTORICAL`; no implementation authority. Rule Task 1 RED awaits all accepted non-marker Phase-145/G145-5B/G145-5C and launcher readiness, unordered; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
