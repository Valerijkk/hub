# Phase 144 — exact Terra dispatch index

> **Status:** FIVE_DOCUMENT_DEFAULT_VS_HARD_CAP_144_01A_ENTRY_AMENDMENT_PASS `5e314fb0` (144-01A entry only; broader Phase 142/cross-contract pending) · 144_01A_STATIC_ACCEPTED `e3e8c774` / EXACT_SHA_PASS · 144_03A_ADDRESS_STATIC_ACCEPTED `41564d00` / EXACT_SHA_PASS · 144_03A_THREAD_STATIC_ACCEPTED `9903f4df` / EXACT_SHA_PASS / CONTRACT `cbc12577` / C_I_M_0_0_0 · Folder source+oracle STATIC_ACCEPTED `f8cb7643986fc591cae16764575337b72ef3a692` and Folder docs accepted `ca4e2854` · 144-03B-Rule-Compiler docs-only STATIC_ACCEPTED `f199546513b758ca7b95328b5ccd7680009b732e` / EXACT_SHA_PASS, transition `6735ae746c7acde657359e8a709cd897c4dce61c`; next action 144-02D/02F adjudication/closure, then separate Rule plan and Phase-145 carrier · 144_01B_BLOCKED_ON_APPLICABLE_SUPPORTED_CURRENT_SHA_DOCKER_EVIDENCE_BY_TASK3 · 144_02A_02B_02C_02D_02F_ROOT_FROZEN_NOT_ACCEPTED · RUNTIME_NOT_RUN
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

## Dependency graph

```text
144-01A engine DTO/error contract
   └── 144-01B deterministic fake + conformance oracle

144-02A1 typed DNS value -> 144-02A2 domain/policy schemas -> 144-02A3 account/identity schemas
144-02B folder/thread/message/membership/label schemas
144-02C object/attachment/draft/submission schemas
144-02D user-settings/signature/rule/contact schemas
144-02F mutation-receipt/rule-deployment schemas
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
   └── 144-04F sync-page policy ── 144-04G sync service
                                      └── 144-04H Ent page/rebuild transaction
                                               └── 144-04I bounded worker lifecycle

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

## Atomic packets

| ID | Single owned behavior | Exact primary files | Oracle / acceptance | Entry dependency |
|---|---|---|---|---|
| 144-01A | Mailbox engine port, immutable DTOs and stable safe errors — 0.1.8 projection/budget/closed-DTO/draft-key amendment is `STATIC_ACCEPTED` at `e3e8c774d734f926f77f8e600b58dc68dabbd993` after Terra/high `EXACT_SHA_PASS`; executable evidence remains `NOT_RUN` | `backend/internal/mail/mailbox_engine.go`, `mailbox_types.go`, `mailbox_errors.go`, `backend/tests/mail/mailbox_engine_contract_test.go` | projection-only pages; exact request/projection/body/attachment fields; variable-byte/item/page limits; budgeted whole message/blob and draft save/reconcile; reference-only submission; stable first-draft reconciliation | accepted `ENGINE-CONTRACT-0.1.8.md` |
| 144-01B | Deterministic in-memory fake for one account boundary — blocked by stricter Task3 authority pending applicable supported current-SHA Docker evidence | `backend/internal/mail/mailbox_fake.go`, `backend/tests/mail/mailbox_fake_test.go` | same-key submit with opaque result pair; exact/mismatched result hydration; scripted incremental/enumeration pages, invalid cursor/anchor restart; strict account isolation | accepted 144-01A amendment + applicable supported current-SHA Docker evidence |
| 144-02A1/A2/A3 | Typed DNS persistence value, then domain/policy schemas, then account/identity schemas — exact representation root-frozen; independent review pending | A1 `backend/internal/mail/domain_dns.go`; A2 `maildomain.go`/`mailboxpolicy.go` + narrow org inverses; A3 `mailaccount.go`/`mailidentity.go` + remaining narrow inverses | typed bounded DNS JSON; normalized uniqueness; exact named tuple/lease/state checks and delete edges; each intermediate commit compiles without a missing schema type | independently accepted persistence 02A representation + founder migration gate |
| 144-02B | Folder, thread, message, message-folder, label and message-label schemas — `CONTRACT_AMENDED_ROOT_REVIEW_PENDING` | `mailfolder.go`, `mailthread.go`, `mailmessage.go`, `mailmessagefolder.go`, `maillabel.go`, `mailmessagelabel.go` | exact normalized-name indexes, required engine message version, bounded values, state/join checks and delete behavior; no body/raw MIME | accepted 144-02A + independently accepted persistence 02B matrix |
| 144-02C | Mail object, attachment, draft and submission schemas only — `CONTRACT_AMENDED_ROOT_REVIEW_PENDING` | `mailobject.go`, `mailattachment.go`, `maildraft.go`, `mailsubmission.go` | stable draft key/client scope; maximum 50 scalar one-draft attachment owners; split engine-result ids/Hub UUID; scan/upload/submission checks | accepted 144-02A + independently accepted persistence 02C matrix |
| 144-02D | User settings, signature, rule and recent-contact schemas only — `CONTRACT_AMENDED_ROOT_REVIEW_PENDING` | `mailusersettings.go`, `mailsignature.go`, `mailrule.go`, `recentcontact.go` | sole default authority, concrete locked rule JSON and bounded external-contact retention | accepted 144-02A + independently accepted persistence 02D matrix |
| 144-02F | Mail mutation receipt target/member and rule-deployment state schemas — `CONTRACT_AMENDED_ROOT_REVIEW_PENDING` | receipt/target/member schemas, `mailruledeployment.go`, narrow inverse edges | exact operation/state/code/count matrices; attempted tuple and canonical rule hash; no response/body/raw key | accepted 144-02A/D + independently accepted persistence 02F matrix |
| 144-02E | Generate and inspect Ent output | generated Ent tree only through repository launcher workflow | clean generation/compile evidence; no manual generated edits; schema drift report | launcher + 144-02A..D + founder #12 |
| 144-03A | Address normalization is `STATIC_ACCEPTED` at `41564d003946ddb78df18efca166ca86e4055405`; Thread contract is accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461` | exact 19-path Thread boundary is `STATIC_ACCEPTED` at `9903f4df3e5ca8034ea3f043b0386170e83a5101` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`) | exact callable, strict bytewise Message-ID/ref chain, prefix/bounds, raw hash and literal oracle; no I/O | Folder static acceptance plus docs acceptance `ca4e2854`; separate Rule candidate lifecycle governs next gate |
| 144-03B | Folder source+oracle are narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` / `EXACT_SHA_PASS` (source `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`, C/I/M `0/0/0`) | exact two-path static-only boundary; no handler, schema, migration, engine or runtime acceptance | Hub UUID tree: live-slot 100/depth 2/cycle; system rename/reparent/delete protected; create custom; engine IDs mapped before policy | `144-03B-Rule-Compiler` docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `f199546513b758ca7b95328b5ccd7680009b732e`; transition `6735ae746c7acde657359e8a709cd897c4dce61c`; 144-02D/02F adjudication/closure, separate Rule plan and Phase-145 carrier precede TDD |
| 144-03C | Send/rate/loop then attachment policy as two sequential variants | `send_policy.go`/send unit, then `attachment_policy.go`/attachment unit | recipient/byte caps and loop headers; fail-closed scan and filename/type bounds | 144-01A |
| 144-03D | Deterministic mapping of the three mail payloads into shared full v1 envelopes — static authoring authorized | `mailbox_events.go`, `backend/tests/mail/mailbox_events_test.go` | explicit event/time/org/actor/correlation; exact schema fields; no body/secret; no clock/random/outbox | 143-04L and 144-01A satisfied |
| 144-04A | Resolve authenticated user to one organization/account and expose repositories | `mailbox_repository.go`, `mailbox_ent_repository.go`, `mailbox_service.go`, focused owner units | owner binding; begin/finish/reclaim mutation lease+epoch fence; sync lease/page interfaces; zero engine on invalid owner | 144-01B + 144-02E |
| 144-04J | Deterministic account/identity desired-state service | `mailbox_account_control_service.go`, focused unit, narrow repository signatures | eligible user/domain scan; immutable address/key/cutoff; alias/user lifecycle generation; idempotent ensure | 144-01B + 144-02A/E |
| 144-04K | Bounded multi-replica account-control worker | `mailbox_account_control_worker.go`, focused unit | due scan/lease/backoff/context shutdown; captured generation only; disable/reactivate; pending aliases | 144-04J |
| 144-04B | Folder then label metadata commands as two sequential variants | `folder_service.go`/folder unit, then `label_service.go`/label unit | versioned pending folder intent/cap/reconcile; optimistic label CRUD and rule-reference delete guard; message assignment remains receipt-backed 04C only | 144-03B + 144-04A/O |
| 144-04C | Message list/read/bulk projection and label-assignment behavior | `message_service.go`, `message_projection.go`, focused units | post-filter before hydrate; receipt-captured expansion; add/remove label provenance recovery; no duplicate training; engine mutations fenced; per-target result; full changed envelope; explicit freshness | 144-03A/D + 144-04A/O |
| 144-04D | Draft create/save/read/delete with lost-response reconciliation | `draft_service.go`, `backend/tests/mail/draft_service_test.go` | raw body only in engine DTO; command lease + begin/end epochs; create replay; `If-Match`; unknown reconcile | 144-01B + 144-02C/E + 144-03C + 144-04A |
| 144-04E | Send receipt state machine and reconciliation | `submission_service.go`, `backend/tests/mail/submission_service_test.go` | command lease + begin/end epochs; hashed key/request; one engine submit; split opaque result ids/Hub UUID; retained-pair projection-only recovery; exact result binding/origin; one full envelope | 144-04D + 144-03D |
| 144-04L | User settings then signatures as sequential service variants | `mailbox_settings_service.go`/unit, then `mailbox_signature_service.go`/unit | active owner; same-account defaults; optimistic writes/deletes; server sanitizer derives plain; no raw pair trust | 144-02D/E + 144-04A |
| 144-04M | Canonical rule CRUD and desired generation | `mailbox_rule_service.go`, focused unit | one Hub transaction; complete-set hash; no engine call/423; deployment state honest | 144-02D/F/E + 144-03B + 144-04A |
| 144-04N | Bounded generation-reconciling rule worker | `mailbox_rule_worker.go`, focused unit | mutation lease; fixed script; Get-before-retry; captured generation only; drift repair | 144-04M + 144-01B |
| 144-04O | Generic mail mutation receipt reserve/replay/tombstone | `mailbox_mutation_receipt.go`, focused unit | public v4/v7 key plus internal folder UUID mode; scoped key digest + raw-keyed canonical request HMAC; 30-day replay/recovery; client lifetime reject tombstone; unknown reconcile seam | 144-02F/E + 144-04A |
| 144-04P | Canonical domain desired-state/reconciliation service | `mailbox_domain_service.go`, focused unit, narrow repository signatures | create/delete intent; stable key/generation/hash; verified/default guards; no DNS mutation | 144-01B + 144-02A/E + 144-04O |
| 144-04Q | Bounded multi-replica domain-control worker | `mailbox_domain_worker.go`, focused unit | Ensure/Get-before-retry; public DNS projection; captured generation; deleting/absence; context shutdown | 144-04P |
| 144-04F | Pure incremental/enumeration page classification and projection commands | `mailbox_sync_policy.go`, `backend/tests/mail/mailbox_sync_policy_test.go` | bounds/order; provenance-safe inbound vs confirmed Sent; first-seen cutoff; changed/deleted event set; no I/O/clock | 144-01B + 144-03A/D |
| 144-04G | Lease/epoch-aware mailbox sync orchestration against fake ports | `mailbox_sync_service.go`, `backend/tests/mail/mailbox_sync_service_test.go`, narrow repository signatures | claim/fetch/apply; invalid cursor to rebuild; checkpoint/catch-up/finalize; stale epoch discards whole page | 144-04A/F |
| 144-04H | Atomic Ent incremental/rebuild projection adapter | `mailbox_ent_repository_sync.go`; no Go DB/component test | account-lock lease/cursor/epoch proof; projection/events/cursor atomic; no-clear rebuild/final cleanup | 144-02E + 144-04G |
| 144-04I | Bounded context-owned multi-replica sync worker | `mailbox_sync_worker.go`, `backend/tests/mail/mailbox_sync_worker_test.go` | max concurrency; due-account claims; lease renewal/expiry; backoff; shutdown; disabled no start | 144-04G/H |
| 144-04R | Generic mutation recovery step in the existing sync worker | `mailbox_mutation_recovery.go`, narrow worker edit, focused unit | oldest-due work; receipt+mutation leases; folder recovery only under proven exclusive-writer capability; captured bulk only; unknown training no repeat; no fifth worker | 144-04B/C/I/O |
| 144-05A | Mail-owned object storage authorization and lifecycle | `mailbox_storage.go`, `mailbox_storage_adapter.go`, `backend/tests/mail/mailbox_storage_test.go` | `mail/<org>/<account>/<uuid>` server key; receipt-safe same-object presign replay; confirm/read/delete reauthorize; no project metadata | 144-02C/E + 144-03C + 144-04A/O |
| 144-05B | Versioned projection consumption and mail subject admission | `mailbox_projection.go`, `backend/internal/platform/nats/client.go`, focused units | subject/envelope match; application-event dedupe; unknown/malformed dead-letter; add `mail.>` once; webhook filter byte-equivalent | 144-03D + 144-04C/E + 143-04G |
| 144-05C | Search, contacts, realtime and named audit as four sequential variants | one of `mailbox_search.go`, `mailbox_contacts.go`, `mailbox_realtime.go`, `mailbox_audit.go` plus its unit | owner filters before snippets; ranking; bounded invalidation; exact named audit | 144-04A..I as applicable |
| 144-05D | Safe body sanitization then remote-image proxy as sequential variants | `mailbox_content.go`/unit, then `mailbox_remote_image.go`/unit | reviewed allowlist or plain-only; token rewrite; every redirect/connect safe; no origin URL/browser leak | 144-04A + 144-04L |
| 144-05E | Owner search document hydration/consumer | `mailbox_search_projection.go`, focused unit | bounded sanitized body + filenames; no raw/BCC/URL; delete/update idempotent; checkpoint freshness | 144-05B/D + 144-05C-Search |
| 144-06A | Public mailbox DTO/error mapping and routes | `mailbox_dto.go`, `mailbox_handler.go`, `backend/cmd/server/wiring_mailbox.go`, focused handler units | exact OpenAPI 0.1.8; bounded pages/details/calendar; NFR-13; no client account/org; Resend unchanged | 144-04* + 144-05* |
| 144-06B | Mailbox policy/domain/health admin service | `mailbox_admin_service.go`, focused service unit | owner organization, policy/domain constraints, redacted engine + projection state/lag and exact audit | 144-04I + 144-05C-Audit + 144-06A |
| 144-06D | Admin per-user alias desired-state service/handler | `mailbox_alias_admin_service.go`, `mailbox_alias_admin_handler.go`, focused units, narrow wiring | same-org account; pending/deleting alias generation; idempotent create; no engine call/423; named audit | 144-04J/K/O + 144-06B/C |
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

## Gate status

| Gate | Status |
|---|---|
| Five-document default-vs-hard-cap/144-01A entry amendment | PASS at `5e314fb080d420696642058ae47397e251c23207`; historical 0.1.7 remains non-authoritative; this narrow PASS authorizes only 144-01A entry and does not accept implementation/runtime or broader Phase 142/cross-contract scope. |
| 144-02A schema readiness | ROOT_FROZEN_REVIEW_PENDING: exact fields, tuples, states, limits and delete rules published in the 0.1.8 persistence companion |
| 144-02B schema readiness | ROOT_FROZEN_RE_REVIEW_PENDING: normalized indexes, required engine version, bounded projection values and join/state checks published |
| 144-02C schema readiness | ROOT_FROZEN_RE_REVIEW_PENDING: stable draft identity, maximum-50 scalar staging ownership, split engine/Hub result identity, single scan authority and state checks published |
| 144-02D schema readiness | ROOT_FROZEN_REVIEW_PENDING: settings pointers are sole defaults; locked concrete rule JSON and external-contact retention published |
| 144-02F schema readiness | ROOT_FROZEN_REVIEW_PENDING: attempted deployment tuple, receipt grammar/state and bulk child matrices published |
| OpenAPI response boundedness | ROOT_FROZEN_REVIEW_PENDING: count/item/string/cursor/aggregate/calendar limits encoded in 0.1.8 |
| Exact Terra implementation runtime | 144-01A is `STATIC_ACCEPTED` at `e3e8c774`; native dispatcher-enforced provenance with actual telemetry `UNKNOWN` is accepted unless exposed telemetry mismatches. 144-01B remains Docker-blocked by Task3 |
| Accepted Docker launcher | NOT_READY |
| Ent generation | NOT_RUN |
| Production migration mechanism | FOUNDER_DECISION_PENDING |
| Product code/tests/build | NOT_RUN |
| 144-03B Rule compiler contract | Docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `f199546513b758ca7b95328b5ccd7680009b732e`; transition `6735ae746c7acde657359e8a709cd897c4dce61c`; Contract/Spec Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Next action: 144-02D/02F adjudication/closure; separate Rule plan and Phase-145 carrier remain before TDD |
> **Current Rule acceptance (2026-08-14):** `144-03B-Rule-Compiler` is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `f199546513b758ca7b95328b5ccd7680009b732e`, transition `6735ae746c7acde657359e8a709cd897c4dce61c`. Contract/Spec Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`; provenance is dispatcher-enforced with no fallback, while telemetry is `UNKNOWN` unless exposed. Next action is 144-02D/02F adjudication/closure; then a separate Rule plan and Phase-145 carrier. Rule TDD/production remain unauthorized; launcher `NOT_READY`, executable evidence `NOT_RUN`, and Phase 148 remains blocked.
