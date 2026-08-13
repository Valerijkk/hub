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
| 145-01A | Mailbox runtime config parsing/validation only | `backend/internal/mail/mailbox_config.go`, `backend/tests/mail/mailbox_config_test.go` | disabled safe default; exclusive-folder-writer false by default; one fixed internal origin; distinct management-key and mailbox-auth references; enabled mail auth rejected until G145-5B; no secret output | G145-2 + accepted 144-01A; enabled path also G145-5B |
| 145-01B | Redacting bounded dual-plane HTTP/JMAP client primitive | `backend/internal/mail/stalwart/client.go`, `errors.go`, focused units | fixed origin; management `/api` and mail JMAP auth providers cannot cross; no arbitrary redirect/URL; auth redaction; body/time/size limits | G145-1/2/4/5/5B + 145-01A |
| 145-02 | Version-specific management compatibility plus mail JMAP session discovery | `stalwart/jmap.go`, `stalwart/config.go`, focused units | fixed `/api` management object contract separate from `/.well-known/jmap`/`/jmap`; required mail capabilities explicit; unsupported/missing capability fails closed | G145-5/5B + 145-01B |
| 145-03C | Domain desired-state mapping | `stalwart/domain.go`, focused units | Ensure/Get by stable key/generation; bounded public expected DNS/DKIM; delete absence; `verified` only from the accepted authority, never domain existence/zone text; no private key | G145-5C + 145-02 + Phase 144 port |
| 145-03A | Account and identity mapping | `stalwart/account.go`, `mapper.go`, focused units | management account/address ensure plus mailbox `Identity/get|set`; exactly one stable opaque id per desired address; Get-before-retry; no fabricated id/password/token | G145-5B + 145-02 + Phase 144 port |
| 145-03B | Folder mapping | `stalwart/folder.go`, focused units | system roles/custom hierarchy; account scoping; exact state token; reviewed exclusive-writer capability and fail-closed recovery signal | G145-5A + 145-02 |
| 145-04 | Message/blob/list mapping | `stalwart/message.go`, `stalwart/blob.go`, focused units | bounded hydrate; raw MIME/body only in result value; BCC policy; pagination; no engine-search path | 145-02 |
| 145-05A | Hydrated incremental changes and invalid-cursor signal | `stalwart/changes.go`, focused units | composite opaque old/new cursor; created/updated/destroyed snapshots; typed invalid/expired; bounded page | 145-04 |
| 145-05B | Base-anchored resumable projection enumeration | `stalwart/enumeration.go`, focused units | composite object-state base; stable unfiltered ID-anchor; mandatory Changes catch-up; anchor/base-history loss typed restart | 145-03B/04/05A |
| 145-06A | Draft create/update and reconciliation with opaque expected version | `stalwart/draft.go`, focused units | Hub mutation id propagation; `ReconcileDraft` applied/absent/indeterminate; no save inside reconcile | 145-04/05A + Phase 144 draft contract |
| 145-06B | Submit and explicit reconciliation | `stalwart/submission.go`, focused units | stable engine key; sent/definitive absent/indeterminate mapping; no blind retry | 145-06A + Phase 144 submission contract |
| 145-07A | Flag/keyword/move/delete mapping | `backend/internal/mail/stalwart/mutation.go`, `backend/tests/mail/stalwart_mutation_test.go` | account-scoped ids; reserved keyword typed separately; partial failure safe; hard-delete policy | 145-03B/04 |
| 145-07B | Generation-reconciling RFC 9661 per-mailbox Sieve mapping | `backend/internal/mail/stalwart/sieve.go`, `backend/tests/mail/stalwart_sieve_test.go` | standard `SieveScript/*` + bounded blob upload on mail JMAP; fixed Hub script; generation/hash marker; lost response observable; management `x:Sieve*` objects forbidden | 145-02 + Phase 144 rule compiler |
| 145-07C | Explicit spam/ham classifier feedback mapping | `backend/internal/mail/stalwart/spam.go`, `backend/tests/mail/stalwart_spam_test.go` | selected-version message-to-feedback and idempotency/reconciliation evidence; accepted/rejected/indeterminate per message; move-only never trained | G145-5 remains blocked + 145-02/04 |
| 145-08 | Port-complete adapter composition | `stalwart/adapter.go`, adapter conformance units | compile assertion; every port method including domain/account/Changes/EnumerateProjection/Sieve/classifier delegated once; redacted health | 145-03A/B/C + 04 + 05A/B + 06A/B + 07A/B/C |
| 145-09A | Mailbox adapter/workers composition and exact root call | `backend/cmd/server/wiring_mailbox.go`, `backend/cmd/server/main.go` | disabled creates no adapter/worker/probe; enabled starts domain/account/combined projection-recovery/rule workers on one root context with reverse cleanup; Resend unchanged | 145-08 + accepted Phase 144 constructors/04I/K/N/Q/R |
| 145-09B | Generic mailbox health mapping for the wired adapter | `backend/internal/mail/mailbox_admin_service.go`, `mailbox_admin_handler.go`, `backend/tests/mail/mailbox_admin_health_test.go` | disabled not_configured; enabled unavailable degraded/503; redacted and admin-first | 145-09A + accepted 144-06B/C |
| 145-10 | Optional bounded Compose profile/config assets | `docker-compose.local.yml`, approved `deployments/mail/**`, names-only env example/docs | digest pin, named volumes for `/etc/stalwart` and `/var/lib/stalwart`, UID/capability review, healthcheck, SMTP-only opt-in public transport; management/mail credential separation; no recovery-admin steady-state credential; no user IMAP/JMAP/ManageSieve/admin mailbox mutation; CPU/RAM/PID limits, no secret | G145-1/3/4/5/5A/5B/5C + 145-09A/B |
| 145-11A | Dedicated engine backup/restore runbook | `docs/runbooks/mail-engine-backup-restore.md` | exact volume/config/digest prerequisites; restore isolation and cleanup | G145-4/5 + 145-10 |
| 145-11B | DNS and deliverability operator runbook | `docs/runbooks/mail-domain-deliverability.md` | manual MX/PTR/SPF/DKIM/DMARC/TLS/rollback; no external-success claim | G145-4/5 + founder domain/operator decision |
| 145-12 | Accepted launcher runtime verification | task-owned report paths only | all GATE-PLAN manifest cases, actual metrics, exact cleanup | G145-1..11 |

## Hard review points

- No package outside `backend/internal/mail/stalwart` imports Stalwart/JMAP-specific DTOs.
- A 401/403 is never automatically treated as token-refresh success; administrator
  credential failure is unavailable and redacted.
- Management calls use only fixed `/api` with a separately scoped API key. Mail JMAP uses
  `/.well-known/jmap` and same-origin session URLs with a separately accepted mailbox
  credential/delegation. Neither authorization header can be sent on the other plane.
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
  model and immutable Hub provisioning key, then separately authenticated mailbox
  `Identity/get|set` obtains stable server ids. Missing identities are created only after
  Get and a lost response is reconciled by another Get; exactly one identity per normalized
  desired address is accepted. Duplicate/unexpected identities degrade safely. A timeout
  never triggers an unobserved fresh create.
- Sieve owns one fixed Hub per-mailbox RFC 9661 `SieveScript` and exposes read/apply
  generation/hash state for reconciliation. It never mutates management
  `x:SieveSystemScript` or `x:SieveUserScript`. Spam/ham uses explicit classifier feedback;
  moving to Junk without accepted feedback is not success.
- Reserved Hub label keywords use selected-version message keyword/Sieve flag mapping but
  remain structurally distinct from public flags. Unknown markers never become label ids;
  marker removal is account-scoped and reconciled like other engine mutations.
- Domain mapping uses the selected-version management object/API, returns only bounded
  public expected DNS/DKIM state, and reconciles unknown create/delete through Get before
  retry. Domain existence, enabled state and `dnsZoneFile` never imply `verified`; only the
  G145-5C authority may promote it. No private DKIM key or provider response crosses the
  adapter.
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
| Phase 144 engine byte boundary | 0.1.8 ROOT_FROZEN / INDEPENDENT_REVIEW_PENDING / NOT ACCEPTED |
| Founder/legal license | UNRESOLVED |
| Exact Terra runtime | ROUTE_AUTHORIZED / new packet actual telemetry unavailable |
| Accepted launcher | NOT_READY |
| Immutable image provenance | Docker Hub OCI index `sha256:a8108e19…`, linux/amd64 manifest `sha256:2bcbd379…` and SLSA attestation reference captured; signature/SBOM verification and approval NOT_RUN |
| Version-specific Stalwart config/API | PARTIAL_RESEARCH: endpoint/object split frozen; exact config and spam feedback NOT_FROZEN |
| Management/mail credential separation | MANAGEMENT TEN-PERMISSION REPLACE SCOPE ROOT_FROZEN / REVIEW PENDING / ENABLED MAIL AUTH DESIGN UNRESOLVED (G145-5B) |
| Management key permission scope | TEN-PERMISSION REPLACE ALLOWLIST ROOT_FROZEN / INDEPENDENT SECURITY REVIEW PENDING |
| Domain verification authority | UNRESOLVED / SELECTED `x:Domain` HAS NO MANUAL-DNS VERIFICATION (G145-5C) |
| Account identity mapping | EXACT `v0.16.17` IDENTITY GET/SET SOURCE FROZEN / ENABLED USE BLOCKED BY G145-5B MAIL AUTH |
| Exclusive custom-folder writer | NOT_RUN / capability defaults false |
| Runtime/resources/backup/external mail | NOT_RUN |
