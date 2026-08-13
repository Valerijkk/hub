# Phase 145 — exact Terra dispatch index

> **Status:** ALL_CODE_PACKETS_BLOCKED_BY_144_CONTRACT_AND_EXTERNAL_GATES
> **Required runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Entry contract:** amended and accepted Phase 144 `MailEngine` byte boundary + `GATE-PLAN.md`

No packet may begin until its listed gates are satisfied. A dispatch prompt carries the
exact selected Stalwart version/digest/API evidence; an executor must return `BLOCKED`
rather than infer current JMAP/admin wire semantics from memory.
Executable packet bodies are in `DISPATCH-PACKETS.md`: concatenate its Common envelope
with exactly one packet after all named gates pass.

## Atomic packets

| ID | Single owned behavior | Exact primary files | Acceptance | Gates / dependency |
|---|---|---|---|---|
| 145-01A | Mailbox runtime config parsing/validation only | `backend/internal/mail/mailbox_config.go`, `backend/tests/mail/mailbox_config_test.go` | disabled safe default; exclusive-folder-writer false by default; required fields only when enabled; no secret output; fixed internal URL | G145-2 + accepted 144-01A |
| 145-01B | Redacting bounded HTTP/JMAP client primitive | `backend/internal/mail/stalwart/client.go`, `errors.go`, focused units | fixed base; no arbitrary redirect/URL; auth redaction; body/time/size limits | G145-1/2/4/5 + 145-01A |
| 145-02 | Version-specific JMAP session/capability discovery | `stalwart/jmap.go`, `stalwart/config.go`, focused units | required capabilities explicit; unsupported/missing capability fails closed | G145-5 + 145-01B |
| 145-03C | Domain desired-state mapping | `stalwart/domain.go`, focused units | Ensure/Get by stable key/generation; bounded public DNS/verification/DKIM; delete absence; no private key | 145-02 + Phase 144 port |
| 145-03A | Account and identity mapping | `stalwart/account.go`, `mapper.go`, focused units | idempotent EnsureAccount by stable key/local part/domain; exact primary/aliases; opaque ids; no password/token | 145-02 + Phase 144 port |
| 145-03B | Folder mapping | `stalwart/folder.go`, focused units | system roles/custom hierarchy; account scoping; exact state token; reviewed exclusive-writer capability and fail-closed recovery signal | G145-5A + 145-02 |
| 145-04 | Message/blob/list mapping | `stalwart/message.go`, `stalwart/blob.go`, focused units | bounded hydrate; raw MIME/body only in result value; BCC policy; pagination; no engine-search path | 145-02 |
| 145-05A | Hydrated incremental changes and invalid-cursor signal | `stalwart/changes.go`, focused units | composite opaque old/new cursor; created/updated/destroyed snapshots; typed invalid/expired; bounded page | 145-04 |
| 145-05B | Base-anchored resumable projection enumeration | `stalwart/enumeration.go`, focused units | composite object-state base; stable unfiltered ID-anchor; mandatory Changes catch-up; anchor/base-history loss typed restart | 145-03B/04/05A |
| 145-06A | Draft create/update and reconciliation with opaque expected version | `stalwart/draft.go`, focused units | Hub mutation id propagation; `ReconcileDraft` applied/absent/indeterminate; no save inside reconcile | 145-04/05A + Phase 144 draft contract |
| 145-06B | Submit and explicit reconciliation | `stalwart/submission.go`, focused units | stable engine key; sent/definitive absent/indeterminate mapping; no blind retry | 145-06A + Phase 144 submission contract |
| 145-07A | Flag/keyword/move/delete mapping | `backend/internal/mail/stalwart/mutation.go`, `backend/tests/mail/stalwart_mutation_test.go` | account-scoped ids; reserved keyword typed separately; partial failure safe; hard-delete policy | 145-03B/04 |
| 145-07B | Generation-reconciling fixed Sieve mapping | `backend/internal/mail/stalwart/sieve.go`, `backend/tests/mail/stalwart_sieve_test.go` | get/upsert fixed Hub script; generation/hash marker; lost response observable; no public raw Sieve | 145-02 + Phase 144 rule compiler |
| 145-07C | Explicit spam/ham classifier feedback mapping | `backend/internal/mail/stalwart/spam.go`, `backend/tests/mail/stalwart_spam_test.go` | selected-version feedback API; accepted/rejected/indeterminate per message; move-only never trained | 145-02/04 |
| 145-08 | Port-complete adapter composition | `stalwart/adapter.go`, adapter conformance units | compile assertion; every port method including domain/account/Changes/EnumerateProjection/Sieve/classifier delegated once; redacted health | 145-03A/B/C + 04 + 05A/B + 06A/B + 07A/B/C |
| 145-09A | Mailbox adapter/workers composition and exact root call | `backend/cmd/server/wiring_mailbox.go`, `backend/cmd/server/main.go` | disabled creates no adapter/worker/probe; enabled starts domain/account/combined projection-recovery/rule workers on one root context with reverse cleanup; Resend unchanged | 145-08 + accepted Phase 144 constructors/04I/K/N/Q/R |
| 145-09B | Generic mailbox health mapping for the wired adapter | `backend/internal/mail/mailbox_admin_service.go`, `mailbox_admin_handler.go`, `backend/tests/mail/mailbox_admin_health_test.go` | disabled not_configured; enabled unavailable degraded/503; redacted and admin-first | 145-09A + accepted 144-06B/C |
| 145-10 | Optional bounded Compose profile/config assets | `docker-compose.local.yml`, approved `deployments/mail/**`, names-only env example/docs | digest pin, named volume, healthcheck, SMTP-only opt-in public transport; no user IMAP/JMAP/ManageSieve/admin mailbox mutation; CPU/RAM/PID limits, no secret | G145-1/3/4/5/5A + 145-09A/B |
| 145-11A | Dedicated engine backup/restore runbook | `docs/runbooks/mail-engine-backup-restore.md` | exact volume/config/digest prerequisites; restore isolation and cleanup | G145-4/5 + 145-10 |
| 145-11B | DNS and deliverability operator runbook | `docs/runbooks/mail-domain-deliverability.md` | manual MX/PTR/SPF/DKIM/DMARC/TLS/rollback; no external-success claim | G145-4/5 + founder domain/operator decision |
| 145-12 | Accepted launcher runtime verification | task-owned report paths only | all GATE-PLAN manifest cases, actual metrics, exact cleanup | G145-1..11 |

## Hard review points

- No package outside `backend/internal/mail/stalwart` imports Stalwart/JMAP-specific DTOs.
- A 401/403 is never automatically treated as token-refresh success; administrator
  credential failure is unavailable and redacted.
- Redirects, capability URLs and upload/download URLs are accepted only when derived from
  the already trusted fixed internal session and same-origin policy frozen for the chosen
  version.
- Response bodies, header counts, object counts, page sizes and recursion are bounded
  before allocation/hydration.
- `ReconcileSubmission` is a real adapter operation. Returning “not found” for timeout or
  calling `Submit` from it violates the contract.
- Invalid cursor is observable degraded state plus bounded rebootstrap orchestration; it
  cannot become an empty cursor and “fresh” status in one step.
- Rebuild captures selected-version Mailbox/Email object states before enumeration, uses
  unfiltered stable ID-anchor traversal rather than positional offsets and then calls
  Changes from that composite base. JMAP query-state drift alone is expected; missing
  anchor, unstable ordering or expired base history is typed restart-required. Partial
  enumeration is never finalized as complete.
- Enabled composition starts exactly one Phase 144 bounded domain-control, account-control,
  combined projection/recovery and rule-deployment worker on the shared root context and registers bounded reverse-order
  cleanup. Disabled composition creates no scheduler, timer, adapter, probe or goroutine.
- Account mapping uses the selected-version Account API's exact EmailLocalPart/domain/alias
  model and immutable Hub provisioning key. A timeout repeats ensure, never fresh create.
- Sieve owns one fixed Hub script and exposes read/apply generation/hash state for
  reconciliation. Spam/ham uses explicit classifier feedback; moving to Junk without
  accepted feedback is not success.
- Reserved Hub label keywords use selected-version message keyword/Sieve flag mapping but
  remain structurally distinct from public flags. Unknown markers never become label ids;
  marker removal is account-scoped and reconciled like other engine mutations.
- Domain mapping uses the selected-version management object/API, returns only public DNS/
  verification/DKIM state, and reconciles unknown create/delete through Get before retry;
  no private DKIM key or provider response crosses the adapter.
- Folder exact-match adoption or absence retry is forbidden unless the reviewed profile and
  runtime exposure evidence prove Hub-exclusive custom-folder writing. System-role folders,
  false/unknown capability or any alternate writer keep recovery degraded with zero write.
- Engine volume backup is distinct from existing PostgreSQL/object-storage backups.
- External SMTP success is never inferred from local JMAP/internal delivery.
- Packet 145-09A owns the only `main.go` root-call edit. Packet 145-09B owns only generic
  MailEngine health-to-HTTP semantics and must not import Stalwart-specific DTOs.

## Gate status

| Gate | Status |
|---|---|
| Phase 144 engine byte boundary | P1 REOPENED / NOT ACCEPTED |
| Founder/legal license | UNRESOLVED |
| Exact Terra runtime | ROUTE_AUTHORIZED / new packet actual telemetry unavailable |
| Accepted launcher | NOT_READY |
| Immutable image provenance | NOT_RUN |
| Version-specific Stalwart config/API | NOT_FROZEN |
| Exclusive custom-folder writer | NOT_RUN / capability defaults false |
| Runtime/resources/backup/external mail | NOT_RUN |
