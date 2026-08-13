# Phase 144 — exact Spark dispatch index

> **Status:** CONTRACT_READY · CODE_RUNTIME_BLOCKED
> **Required runtime:** `gpt-5.3-codex` / `xhigh`, no fallback
> **Contract:** Mail 0.1.4 (mailbox surface unchanged; current root re-review required)
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
   └── 144-02E accepted Docker Ent generation + generated diff review

144-03A address -> threading variants
144-03B folder tree -> rule compiler variants
144-03C rate/loop -> attachment policy variants

144-01B + 144-02E
   ├── 144-04A account resolver/repository base
   ├── 144-04B folder -> label variants
   ├── 144-04C message projection/read/bulk
   ├── 144-04D draft save/reconcile
   └── 144-04E submission idempotency/reconcile

144-04* + 144-03*
   ├── 144-05A mail-owned storage port
   ├── 144-05B outbox/projection + EVENTS subject admission
   └── 144-05C search -> contact -> realtime -> audit variants
        └── 144-06A public handlers/routes
             └── 144-06B mailbox admin service
                  └── 144-06C mailbox admin handler/routes
                       └── 144-07 docs sync + independent review
```

## Atomic packets

| ID | Single owned behavior | Exact primary files | Oracle / acceptance | Entry dependency |
|---|---|---|---|---|
| 144-01A | Mailbox engine port, immutable DTOs and stable safe errors | `backend/internal/mail/mailbox_engine.go`, `mailbox_types.go`, `mailbox_errors.go`, `backend/tests/mail/mailbox_engine_contract_test.go` | consumer compile oracle; opaque account/message/blob/state ids; `SaveDraft` expected version + mutation id; `ReconcileDraft`; `Submit` stable key; `ReconcileSubmission` explicit | contract 0.1.4 |
| 144-01B | Deterministic in-memory fake for one account boundary | `backend/internal/mail/mailbox_fake.go`, `backend/tests/mail/mailbox_fake_test.go` | same-key submit never duplicates; timeout/unknown/reconcile scripted; owner/account mismatch makes zero mutation | 144-01A |
| 144-02A | Account, identity, domain and mailbox-policy schemas only | `backend/ent/schema/mailaccount.go`, `mailidentity.go`, `maildomain.go`, `mailboxpolicy.go`, narrow inverse edges in `user.go`/`organization.go` | normalized uniqueness, one default partial indexes, no secrets; no generated edit | contract + founder migration gate noted |
| 144-02B | Folder, thread, message, message-folder and label schemas only | `mailfolder.go`, `mailthread.go`, `mailmessage.go`, `mailmessagefolder.go`, `maillabel.go` | same-account edges, engine projection fields, no body/raw MIME | 144-02A reviewed |
| 144-02C | Mail object, attachment, draft and submission schemas only | `mailobject.go`, `mailattachment.go`, `maildraft.go`, `mailsubmission.go` | hashed key uniqueness; opaque versions; attachment backing XOR represented; no project ACL edge | 144-02A reviewed |
| 144-02D | User settings, signature, rule and recent-contact schemas only | `mailusersettings.go`, `mailsignature.go`, `mailrule.go`, `recentcontact.go` | account ownership, bounded config, one defaults, no raw Sieve | 144-02A reviewed |
| 144-02E | Generate and inspect Ent output | generated Ent tree only through repository launcher workflow | clean generation/compile evidence; no manual generated edits; schema drift report | launcher + 144-02A..D + founder #12 |
| 144-03A | Address normalization then thread identity as two sequential variants | `address_policy.go`/address unit, then `threading.go`/thread unit | Unicode/case/domain first; message-id/ref chain/prefix/bounds second; no I/O | 144-01A |
| 144-03B | Folder tree then typed rule compilation as two sequential variants | `folder_tree.go`/folder unit, then `rule_compiler.go`/rule unit | max depth 2/cycle/system roles; only frozen conditions/actions; injection impossible | 144-01A |
| 144-03C | Send/rate/loop then attachment policy as two sequential variants | `send_policy.go`/send unit, then `attachment_policy.go`/attachment unit | recipient/byte caps and loop headers; fail-closed scan and filename/type bounds | 144-01A |
| 144-04A | Resolve authenticated user to one organization/account and expose repositories | `mailbox_repository.go`, `mailbox_ent_repository.go`, `mailbox_service.go`, focused owner units | client never supplies account/org; suspended/deleted account makes zero engine call | 144-01B + 144-02E |
| 144-04B | Folder then label commands as two sequential variants | `folder_service.go`/folder unit, then `label_service.go`/label unit | engine folder reconciliation; Hub labels transactional; partial failures explicit | 144-03B + 144-04A |
| 144-04C | Message list/read/bulk projection behavior | `message_service.go`, `message_projection.go`, focused units | post-filter before hydrate; per-item bulk result; engine flags/folders authoritative | 144-03A + 144-04A |
| 144-04D | Draft create/save/read/delete with lost-response reconciliation | `draft_service.go`, `backend/tests/mail/draft_service_test.go` | raw body only in engine DTO; create key replay; `If-Match`; unknown save reconciles before write | 144-01B + 144-02C/E + 144-03C |
| 144-04E | Send receipt state machine and reconciliation | `submission_service.go`, `backend/tests/mail/submission_service_test.go` | hashed idempotency key/request hash; one engine submit; unknown never blind-retries; one outbox | 144-04D |
| 144-05A | Mail-owned object storage authorization and lifecycle | `mailbox_storage.go`, `mailbox_storage_adapter.go`, `backend/tests/mail/mailbox_storage_test.go` | `mail/<org>/<account>/<uuid>` server key; presign/confirm/read/delete reauthorize; no project metadata | 144-02C/E + 144-03C + 144-04A |
| 144-05B | Projection outbox and mail subject admission | `mailbox_events.go`, `mailbox_projection.go`, `backend/internal/platform/nats/client.go`, focused units | add `mail.>` exactly once after Phase 143 admitted `calendar.>`; webhook filter byte-equivalent; no body/secret in payload | 144-04C/E + 143-04G |
| 144-05C | Search, contacts, realtime and named audit as four sequential variants | one of `mailbox_search.go`, `mailbox_contacts.go`, `mailbox_realtime.go`, `mailbox_audit.go` plus its unit | owner filters before snippets; ranking; bounded invalidation; exact named audit | 144-04A..E |
| 144-06A | Public mailbox DTO/error mapping and routes | `mailbox_dto.go`, `mailbox_handler.go`, `backend/cmd/server/wiring_mailbox.go`, focused handler units | exact OpenAPI 0.1.4 paths; NFR-13; no client account/org; legacy Resend routes unchanged | 144-04* + 144-05* |
| 144-06B | Mailbox policy/domain/health admin service | `mailbox_admin_service.go`, focused service unit | owner organization, policy/domain constraints, redacted health and exact audit intent | 144-05C-Audit + 144-06A |
| 144-06C | Separate mailbox admin handler and narrow wiring | `mailbox_admin_handler.go`, narrow wiring additions, focused handler unit | admin middleware + named audit; no flat Resend route/shape drift | 144-06B |
| 144-07 | Current-state docs sync and independent static review | routed domain docs + phase verification artifact only | implementation vs contract matrix; honest `NOT_RUN`; no protected QA paths | all accepted code packets |

## Frozen service semantics

- Engine ids are opaque and are always paired with the server-resolved account id.
- Raw MIME, recipients, subject and rich/plain draft/message bodies are engine state. Hub
  PostgreSQL stores safe projections, Hub relations, hashes and opaque versions only.
- `MailMessageFolder` and flag fields are rebuildable snapshots. Mutations call the engine
  and reconcile the returned/new change state; the projection does not authorize access.
- Draft mutation is serialized per draft. A lost response becomes `unknown`; a later
  operation first compares/retrieves engine state and never overwrites blindly.
- Submission receipts persist only hashes of HTTP keys. `unknown` invokes
  `ReconcileSubmission`; indeterminate stays unknown and never calls `Submit` again.
- Hub uploads use the dedicated mail storage port. Existing project-bound `FileMetadata`
  is not adapted into a mailbox permission model.
- Every domain mutation that emits a durable event persists it through the current outbox
  transaction seam. Direct post-commit publish is not an accepted substitute.
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
| Contract 0.1.4 root static review | PASS on `46cfa26c`; independent PENDING |
| Exact Spark implementation runtime | UNAVAILABLE |
| Accepted Docker launcher | NOT_READY |
| Ent generation | NOT_RUN |
| Production migration mechanism | FOUNDER_DECISION_PENDING |
| Product code/tests/build | NOT_RUN |
