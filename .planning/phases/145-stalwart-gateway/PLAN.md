# Phase 145 — Stalwart Gateway and Docker Mail Profile

> Task: TASK-238 · Status: BLOCKED_BY_144_CONTRACT_AND_EXTERNAL_GATES · Depends on: Phase 144

## Hard entry gates

0. Phase 144 engine projection/content/blob byte boundary and admin hard ceilings are
   amended and independently accepted.
1. Founder/legal chooses AGPL-3.0 or SELv2 terms.
2. Accepted Docker-only launcher exists for this task.
3. Candidate image tag and digest provenance are refreshed immediately before use.
4. Separate management `/api` and mailbox `/jmap` credential planes are reviewed; a
   management API key is never treated as a mail credential.
5. A domain-verification authority and bounded DNS projection are reviewed. Selected
   Stalwart `x:Domain` does not expose manual-DNS verification, so domain existence or
   `dnsZoneFile` cannot satisfy Hub `verified`.
6. Account provisioning is reviewed as two-plane reconciliation: management `x:Account`
   owns the address set, while separately authenticated mailbox `Identity/get|set` supplies
   stable RFC 8621 ids. This path remains disabled until gate 4/G145-5B.

No runtime implementation/claim proceeds by silent fallback.

## Ordered tasks after gates

1. Implement two SSRF-safe, redacting Stalwart transport planes under one trusted internal
   origin: scoped management API-key calls to fixed `/api`, and separately authenticated
   mailbox discovery/JMAP calls to `/.well-known/jmap` plus same-origin session URLs in
   `backend/internal/mail/stalwart/**`.
2. Implement domain/account/identity/folder/message/draft/submission/flags/fixed-Sieve
   mappings and, only after its separate evidence gate, explicit spam/ham
   classifier-feedback mapping,
   including opaque draft versions, stable submission keys and explicit submission
   reconciliation; prove adapter contract conformance.
   Folder lost-response recovery is enabled only by a reviewed capability proving Hub is
   the sole custom-folder writer; false/unknown or alternate writer remains degraded.
   Domain/account mappings use selected-version management `x:*` objects. Before the
   domain-verification gate, domain mapping returns at most `dns_required` and account
   provisioning remains ineligible. Account mapping obtains identity ids only from mailbox
   `Identity/get|set` after management provisioning; it never fabricates ids. Per-mailbox
   Sieve uses RFC 9661 `SieveScript/*` on mail JMAP, never the global management
   `x:SieveSystemScript` or `x:SieveUserScript` objects. Spam/ham stays omitted until an
   exact message mapping and idempotency/reconciliation contract is evidenced.
3. Implement bounded projection-only incremental Changes plus a resumable base-state/ID-anchored
   projection enumeration whose composite base cursor is accepted by Changes. Query-state
   drift converges through catch-up; missing anchor, unstable enumeration or expired base
   history returns typed restart-required and never resets silently.
4. Compose exactly one Phase 144 domain-control, account-control, combined leased
   projection/recovery and rule-deployment worker
   only for an enabled adapter on the shared root context, with bounded reverse-order
   cleanup; extend admin health without exposing credentials, cursors or checkpoints.
5. Add optional `mail` profile, named volumes, healthcheck, restart and explicit CPU/RAM/
   PID limits; add names-only env examples. Public transport is SMTP-only in wave 1;
   JMAP/admin stay internal to Hub and user IMAP/JMAP/ManageSieve remain unpublished.
6. Treat invalid/expired change cursors as degraded state followed by a bounded,
   owner-scoped projection rebuild; never silently reset and present stale data as fresh.
7. Through accepted launcher prove health, authenticated JMAP, internal delivery,
   instance-domain SMTP, accepted domain-verification authority, enumeration/catch-up plus
   folder lost-response recovery, replica
   lease exclusion, four-worker cancellation, restart persistence, exact cleanup and actual
   resource use.
8. Document dedicated engine-volume backup/restore and DNS/deliverability prerequisites.

External Internet delivery remains a separate operator acceptance even after local
transport passes.
