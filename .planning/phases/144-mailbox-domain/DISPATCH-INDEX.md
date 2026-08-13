# Phase 144 — exact Terra dispatch index

> **Status:** CONTRACT_0_1_7_REOPENED_P1_ENGINE_SCHEMA_AND_OPENAPI_BOUNDS · 144_01A_SOURCE_PASS · 144_01B_CONTRACT_AND_MODEL_BLOCKED · 144_02A_CONTRACT_BLOCKED · RUNTIME_NOT_RUN
> **Required runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Contract:** Mail 0.1.7 calendar CVE-v1 review historically passed on `b3fb6f48`; engine, schema and public response bounds are now P1-reopened
> **Execution:** sequential for overlapping files; independent pure packets may run in parallel only after root freezes their shared types

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

144-02A account/identity/domain/policy schemas
144-02B folder/thread/message/membership/label schemas
144-02C object/attachment/draft/submission schemas
144-02D user-settings/signature/rule/contact schemas
144-02F mutation-receipt/rule-deployment schemas
   └── 144-02E accepted Docker Ent generation + generated diff review

144-03A address -> threading variants
144-03B folder tree -> rule compiler variants
144-03C rate/loop -> attachment policy variants
143-04L + 144-01A -> 144-03D deterministic mail event mapper

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
| 144-01A | Mailbox engine port, immutable DTOs and stable safe errors — source matches `43d421ef`, but P1 response-byte boundary reopened; evidence `IMPLEMENTATION-144-01A.md` and `REVIEW-144-01A-ENGINE-BOUNDS.md` | `backend/internal/mail/mailbox_engine.go`, `mailbox_types.go`, `mailbox_errors.go`, `backend/tests/mail/mailbox_engine_contract_test.go` | consumer compile oracle; bounded Changes/enumeration; provisioning/Sieve/classifier/draft/submission reconcile explicit | contract 0.1.7 |
| 144-01B | Deterministic in-memory fake for one account boundary — blocked by 01A contract amendment and unavailable actual model telemetry | `backend/internal/mail/mailbox_fake.go`, `backend/tests/mail/mailbox_fake_test.go` | same-key submit; scripted incremental/enumeration pages, invalid cursor/anchor restart; strict account isolation | accepted 144-01A amendment |
| 144-02A | Account, identity, domain and mailbox-policy schemas only — `BLOCKED_CONTRACT`; evidence `REVIEW-144-02A-SCHEMA-READINESS.md` | `backend/ent/schema/mailaccount.go`, `mailidentity.go`, `maildomain.go`, `mailboxpolicy.go`, narrow inverse edges in `user.go`/`organization.go` | normalized uniqueness; account/domain desired/deployed generations + leases; projection cursors/epochs/cutoff; identity status; no secrets | exact field/bound/cardinality amendment + founder migration gate |
| 144-02B | Folder, thread, message, message-folder, label and message-label schemas — `BLOCKED_CONTRACT`; evidence `REVIEW-144-02B-SCHEMA-READINESS.md` | `mailfolder.go`, `mailthread.go`, `mailmessage.go`, `mailmessagefolder.go`, `maillabel.go`, `mailmessagelabel.go` | pending/versioned folder intent; same-account edges; label marker/provenance; rebuild marks; projection origin/event marker; no body/raw MIME | accepted 144-02A + exact normalized-name/bound/join/check matrix |
| 144-02C | Mail object, attachment, draft and submission schemas only — `BLOCKED_CONTRACT`; evidence `REVIEW-144-02C-SCHEMA-READINESS.md` | `mailobject.go`, `mailattachment.go`, `maildraft.go`, `mailsubmission.go` | hashed key uniqueness; opaque versions; attachment backing XOR represented; no project ACL edge | accepted 144-02A + stable draft recovery key + staging/state/bound matrix |
| 144-02D | User settings, signature, rule and recent-contact schemas only — `BLOCKED_CONTRACT`; evidence `REVIEW-144-02D-SCHEMA-READINESS.md` | `mailusersettings.go`, `mailsignature.go`, `mailrule.go`, `recentcontact.go` | account ownership, bounded config, one defaults, no raw Sieve | accepted 144-02A + one default authority + rule-target/contact-retention matrix |
| 144-02F | Mail mutation receipt target/member and rule-deployment state schemas — `BLOCKED_CONTRACT`; evidence `REVIEW-144-02F-SCHEMA-READINESS.md` | receipt/target/member schemas, `mailruledeployment.go`, narrow inverse edges | client/internal key source; client hash tombstone + bounded bulk expansion/outcome snapshot, no response/body/raw key; desired/deployed/attempted generations and paired lease | accepted 144-02A/D + exact receipt/deployment matrices |
| 144-02E | Generate and inspect Ent output | generated Ent tree only through repository launcher workflow | clean generation/compile evidence; no manual generated edits; schema drift report | launcher + 144-02A..D + founder #12 |
| 144-03A | Address normalization then thread identity as two sequential variants | `address_policy.go`/address unit, then `threading.go`/thread unit | Unicode/case/domain first; message-id/ref chain/prefix/bounds second; no I/O | 144-01A |
| 144-03B | Folder tree then typed rule compilation as two sequential variants | `folder_tree.go`/folder unit, then `rule_compiler.go`/rule unit | max 100/depth 2/cycle/system roles; native actions + opaque label marker only; injection impossible | 144-01A |
| 144-03C | Send/rate/loop then attachment policy as two sequential variants | `send_policy.go`/send unit, then `attachment_policy.go`/attachment unit | recipient/byte caps and loop headers; fail-closed scan and filename/type bounds | 144-01A |
| 144-03D | Deterministic mapping of the three mail payloads into shared full v1 envelopes | `mailbox_events.go`, `backend/tests/mail/mailbox_events_test.go` | explicit event/time/org/actor/correlation; exact schema fields; no body/secret; no clock/random/outbox | 143-04L + 144-01A |
| 144-04A | Resolve authenticated user to one organization/account and expose repositories | `mailbox_repository.go`, `mailbox_ent_repository.go`, `mailbox_service.go`, focused owner units | owner binding; begin/finish/reclaim mutation lease+epoch fence; sync lease/page interfaces; zero engine on invalid owner | 144-01B + 144-02E |
| 144-04J | Deterministic account/identity desired-state service | `mailbox_account_control_service.go`, focused unit, narrow repository signatures | eligible user/domain scan; immutable address/key/cutoff; alias/user lifecycle generation; idempotent ensure | 144-01B + 144-02A/E |
| 144-04K | Bounded multi-replica account-control worker | `mailbox_account_control_worker.go`, focused unit | due scan/lease/backoff/context shutdown; captured generation only; disable/reactivate; pending aliases | 144-04J |
| 144-04B | Folder then label metadata commands as two sequential variants | `folder_service.go`/folder unit, then `label_service.go`/label unit | versioned pending folder intent/cap/reconcile; optimistic label CRUD and rule-reference delete guard; message assignment remains receipt-backed 04C only | 144-03B + 144-04A/O |
| 144-04C | Message list/read/bulk projection and label-assignment behavior | `message_service.go`, `message_projection.go`, focused units | post-filter before hydrate; receipt-captured expansion; add/remove label provenance recovery; no duplicate training; engine mutations fenced; per-target result; full changed envelope; explicit freshness | 144-03A/D + 144-04A/O |
| 144-04D | Draft create/save/read/delete with lost-response reconciliation | `draft_service.go`, `backend/tests/mail/draft_service_test.go` | raw body only in engine DTO; command lease + begin/end epochs; create replay; `If-Match`; unknown reconcile | 144-01B + 144-02C/E + 144-03C + 144-04A |
| 144-04E | Send receipt state machine and reconciliation | `submission_service.go`, `backend/tests/mail/submission_service_test.go` | command lease + begin/end epochs; hashed key/request; one engine submit; unknown reconcile; one full envelope | 144-04D + 144-03D |
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
| 144-06A | Public mailbox DTO/error mapping and routes | `mailbox_dto.go`, `mailbox_handler.go`, `backend/cmd/server/wiring_mailbox.go`, focused handler units | exact OpenAPI 0.1.7; account/folder/thread projection metadata; NFR-13; no client account/org; Resend unchanged | 144-04* + 144-05* |
| 144-06B | Mailbox policy/domain/health admin service | `mailbox_admin_service.go`, focused service unit | owner organization, policy/domain constraints, redacted engine + projection state/lag and exact audit | 144-04I + 144-05C-Audit + 144-06A |
| 144-06D | Admin per-user alias desired-state service/handler | `mailbox_alias_admin_service.go`, `mailbox_alias_admin_handler.go`, focused units, narrow wiring | same-org account; pending/deleting alias generation; idempotent create; no engine call/423; named audit | 144-04J/K/O + 144-06B/C |
| 144-06C | Separate mailbox admin handler and narrow wiring | `mailbox_admin_handler.go`, narrow wiring additions, focused handler unit | admin middleware + named audit; no flat Resend route/shape drift | 144-06B |
| 144-06E | Sensitive HTTP route metadata redaction | platform access logger, generic HTTP audit middleware, focused in-memory unit | matched route template only; unmatched constant; image token/search query/UUID absent; existing audit metadata preserved | 144-06A |
| 144-07 | Current-state docs sync and independent static review | routed domain docs + phase verification artifact only | implementation vs contract matrix; honest `NOT_RUN`; no protected QA paths | all accepted code packets |

## Frozen service semantics

- Engine ids are opaque and are always paired with the server-resolved account id.
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
  operation first compares/retrieves engine state and never overwrites blindly.
- Submission receipts persist only hashes of HTTP keys. `unknown` invokes
  `ReconcileSubmission`; indeterminate stays unknown and never calls `Submit` again.
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
| Contract 0.1.7 root/independent static review | Historical PASS on `b3fb6f48`; later root deep audit reopened P1 engine response/admin byte bounds |
| 144-02A schema readiness | BLOCKED: exact operational field matrix, persisted safe-code allowlists, DNS TTL shape and domain/identity cardinality are not frozen |
| 144-02B schema readiness | BLOCKED: case-insensitive uniqueness representation, bounded projection values and join/state relational checks are not frozen |
| 144-02C schema readiness | BLOCKED: initial draft-create recovery identity, attachment staging cardinality, upload/scan authority and conditional state checks are not frozen |
| 144-02D schema readiness | BLOCKED: identity/signature defaults have duplicate authorities; rule target concurrency/JSON bounds and recent-contact retention are not frozen |
| 144-02F schema readiness | BLOCKED: rule deployment lacks durable attempted generation/hash; receipt lifecycle/operation grammar and aggregate child checks are not frozen |
| OpenAPI response boundedness | BLOCKED: list/detail collections and public hydrated values do not consistently encode accepted page/item/length/aggregate limits |
| Exact Terra implementation runtime | 144-01A authoring route recorded; 144-01B is contract-blocked and its attempted actual telemetry was unavailable |
| Accepted Docker launcher | NOT_READY |
| Ent generation | NOT_RUN |
| Production migration mechanism | FOUNDER_DECISION_PENDING |
| Product code/tests/build | NOT_RUN |
