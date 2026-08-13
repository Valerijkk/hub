# Phase 145 — exact Spark dispatch index

> **Status:** ALL_CODE_PACKETS_GATE_BLOCKED
> **Required runtime:** `gpt-5.3-codex` / `xhigh`, no fallback
> **Entry contract:** accepted Phase 144 `MailEngine` port + `GATE-PLAN.md`

No packet may begin until its listed gates are satisfied. A dispatch prompt carries the
exact selected Stalwart version/digest/API evidence; an executor must return `BLOCKED`
rather than infer current JMAP/admin wire semantics from memory.
Executable packet bodies are in `DISPATCH-PACKETS.md`: concatenate its Common envelope
with exactly one packet after all named gates pass.

## Atomic packets

| ID | Single owned behavior | Exact primary files | Acceptance | Gates / dependency |
|---|---|---|---|---|
| 145-01A | Mailbox runtime config parsing/validation only | `backend/internal/mail/mailbox_config.go`, `backend/tests/mail/mailbox_config_test.go` | disabled safe default; required fields only when enabled; no secret output; fixed internal URL | G145-2 + accepted 144-01A |
| 145-01B | Redacting bounded HTTP/JMAP client primitive | `backend/internal/mail/stalwart/client.go`, `errors.go`, focused units | fixed base; no arbitrary redirect/URL; auth redaction; body/time/size limits | G145-1/2/4/5 + 145-01A |
| 145-02 | Version-specific JMAP session/capability discovery | `stalwart/jmap.go`, `stalwart/config.go`, focused units | required capabilities explicit; unsupported/missing capability fails closed | G145-5 + 145-01B |
| 145-03A | Account and identity mapping | `stalwart/account.go`, `mapper.go`, focused units | idempotent EnsureAccount; opaque ids; no password/token persisted | 145-02 + Phase 144 port |
| 145-03B | Folder mapping | `stalwart/folder.go`, focused units | system roles/custom hierarchy; account scoping; exact state token | 145-02 |
| 145-04 | Message/blob/list/search mapping | `stalwart/message.go`, `stalwart/blob.go`, focused units | bounded hydrate; raw MIME/body only in result value; BCC policy; pagination | 145-02 |
| 145-05A | Hydrated incremental changes and invalid-cursor signal | `stalwart/changes.go`, focused units | composite opaque old/new cursor; created/updated/destroyed snapshots; typed invalid/expired; bounded page | 145-04 |
| 145-05B | Base-anchored resumable projection enumeration | `stalwart/enumeration.go`, focused units | composite object-state base; stable unfiltered ID-anchor; mandatory Changes catch-up; anchor/base-history loss typed restart | 145-03B/04/05A |
| 145-06A | Draft create/update and reconciliation with opaque expected version | `stalwart/draft.go`, focused units | Hub mutation id propagation; `ReconcileDraft` applied/absent/indeterminate; no save inside reconcile | 145-04/05A + Phase 144 draft contract |
| 145-06B | Submit and explicit reconciliation | `stalwart/submission.go`, focused units | stable engine key; sent/definitive absent/indeterminate mapping; no blind retry | 145-06A + Phase 144 submission contract |
| 145-07A | Flag/move/delete mapping | `backend/internal/mail/stalwart/mutation.go`, `backend/tests/mail/stalwart_mutation_test.go` | account-scoped ids; partial failure safe; explicit hard-delete policy | 145-03B/04 |
| 145-07B | Typed Sieve mapping | `backend/internal/mail/stalwart/sieve.go`, `backend/tests/mail/stalwart_sieve_test.go` | compiler-produced script only; bounded upload/activate; no public raw Sieve | 145-02 + Phase 144 rule compiler |
| 145-08 | Port-complete adapter composition | `stalwart/adapter.go`, adapter conformance units | compile assertion; every port method including Changes/EnumerateProjection delegated once; redacted health | 145-03A/B + 04 + 05A/B + 06A/B + 07A/B |
| 145-09A | Mailbox adapter/worker composition and exact root call | `backend/cmd/server/wiring_mailbox.go`, `backend/cmd/server/main.go` | disabled creates no adapter/worker/probe; enabled injects one adapter and starts one context-owned bounded worker with LIFO cleanup; legacy Resend wiring unchanged | 145-08 + accepted Phase 144 constructors/04I |
| 145-09B | Generic mailbox health mapping for the wired adapter | `backend/internal/mail/mailbox_admin_service.go`, `mailbox_admin_handler.go`, `backend/tests/mail/mailbox_admin_health_test.go` | disabled not_configured; enabled unavailable degraded/503; redacted and admin-first | 145-09A + accepted 144-06B/C |
| 145-10 | Optional bounded Compose profile/config assets | `docker-compose.local.yml`, approved `deployments/mail/**`, names-only env example/docs | digest pin, named volume, healthcheck, explicit ports, CPU/RAM/PID limits, no secret | G145-1/3/4/5 + 145-09A/B |
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
- Enabled composition starts exactly one Phase 144 bounded projection worker on the shared
  root context and registers its bounded `Stop` cleanup. Disabled composition creates no
  scheduler, timer, adapter, probe or goroutine.
- Engine volume backup is distinct from existing PostgreSQL/object-storage backups.
- External SMTP success is never inferred from local JMAP/internal delivery.
- Packet 145-09A owns the only `main.go` root-call edit. Packet 145-09B owns only generic
  MailEngine health-to-HTTP semantics and must not import Stalwart-specific DTOs.

## Gate status

| Gate | Status |
|---|---|
| Founder/legal license | UNRESOLVED |
| Exact Spark runtime | UNAVAILABLE |
| Accepted launcher | NOT_READY |
| Immutable image provenance | NOT_RUN |
| Version-specific Stalwart config/API | NOT_FROZEN |
| Runtime/resources/backup/external mail | NOT_RUN |
