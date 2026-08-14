# Phase 145 — Stalwart Gateway and Docker Mail Profile

> Task: TASK-238 · Status: BLOCKED_BY_144_CONTRACT_AND_EXTERNAL_GATES · Depends on: Phase 144

> **Upstream lifecycle update:** 144-02D docs-only boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` is `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; contrary 02D/02F-next text below is historical. Sole Mail next action is three exact 02F closure reviews. Phase 145 remains blocked/not accepted; G145-5B/G145-5C and carrier/runtime gates remain open.

## Hard entry gates

0. Phase 144 engine projection/content/blob byte boundary and admin hard ceilings are
   amended and independently accepted.
1. Founder/legal chooses AGPL-3.0 or SELv2 terms.
2. Accepted Docker-only launcher exists for this task.
3. Candidate image tag and digest provenance are refreshed immediately before use.
4. Shared `/jmap` management Registry and mailbox method contexts are reviewed with
   separate credential providers and non-overlapping method allowlists; authentication is
   selected by typed operation, never URL, and mixed batches fail before I/O. A management
   API key is never treated as a mail credential. Shared-account ACL is not a
   full-engine substitute because exact candidate source keeps Identity, EmailSubmission,
   Sieve and Blob member-only. Founder/security selects per-user Authorization Code OAuth
   (root recommendation) or explicitly accepts broad dedicated-account impersonation for
   zero-touch operation; no OAuth client-credentials flow is assumed.
5. If Authorization Code OAuth is selected, one user-reachable HTTPS origin and exact
   default-deny listener/path/proxy boundary are accepted. `STALWART_PUBLIC_URL` controls
   OAuth/OIDC and JMAP discovery together; separate browser and Hub origins are not assumed.
6. A domain-verification authority and bounded DNS projection are reviewed. Selected
   Stalwart `x:Domain` does not expose manual-DNS verification, so domain existence or
   `dnsZoneFile` cannot satisfy Hub `verified`.
7. Account provisioning is reviewed as two-context reconciliation on shared `/jmap`:
   management `x:Account` owns the address set, while separately authenticated mailbox
   `Identity/get|set` supplies stable RFC 8621 ids. This path remains disabled until gate
   4/G145-5B.

No runtime implementation/claim proceeds by silent fallback.

## Ordered tasks after gates

1. Implement two SSRF-safe, redacting typed Stalwart contexts under one trusted internal
   origin and shared fixed `/jmap`: scoped API-key Registry calls and separately authenticated
   mailbox calls. Mail discovery uses `/.well-known/jmap`; `/api` is limited to individually
   allowlisted helper/auth path+method pairs. Reject mixed Registry/mail batches and wrong-
   context credentials before I/O in `backend/internal/mail/stalwart/**`.
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
   PID limits; add names-only env examples. Public mail transport is SMTP-only in wave 1;
   JMAP/admin stay Hub-only and user IMAP/JMAP/ManageSieve remain unpublished. A separately
   reviewed HTTPS surface may expose only OAuth paths if G145-5D selects Authorization Code.
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

## Rule compiler boundary blocker

`144-03B-Rule-Compiler` is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
`f199546513b758ca7b95328b5ccd7680009b732e`; the lifecycle transition/review target is
`6735ae746c7acde657359e8a709cd897c4dce61c`, not the acceptance commit. Contract/Spec
Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency Terra/high PASS C/I/M
`0/0/0`; provenance is dispatcher-enforced, fallback none and telemetry `UNKNOWN` unless
exposed. It accepts typed IR only, not an engine port or Sieve bytes. Phase 145 still needs
accepted 144-02D plus 144-02F's three exact closure reviews, then its later separate Rule plan and
a separate accepted engine-port carrier before Sieve work. Engine folder ID is not mailbox name;
attachment IR needs MIME gate. No Rule TDD/source/schema/runtime/product/release authority follows.
