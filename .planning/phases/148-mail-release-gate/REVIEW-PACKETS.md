# Phase 148 — independent review packets

> **Review runtime:** `gpt-5.6-terra`; effort is explicit per packet
> **No implementation:** reviewers do not edit product/test code or commit
> **Input:** exact current SHA, frozen contract set, changed-file manifest and evidence
> **Output:** findings only; fixes return to atomic Terra packets

Each reviewer reads the actual changed source in its scope, not merely plan summaries. The
root deduplicates findings, freezes exact fixes, dispatches each code fix to
`gpt-5.6-terra` / `xhigh`, then sends the changed SHA back to the relevant reviewer. A
review cannot self-certify its own fix.

## Common review envelope

- Confirm actual model/effort first. Missing/mismatched provenance returns `BLOCKED`.
- Read root `AGENTS.md`, `.planning/STATE.md`, routed domain docs, Mail 0.1.8 contract,
  Phase 147 contract and the supplied exact diff. Never read protected test-rebase or
  `pro-qa-ai` paths.
- Do not run host tests/build/lint/validators or direct Docker/Compose. Review accepted
  launcher reports if present; otherwise label runtime conclusions `NOT_RUN`.
- Findings use `P0` security/data-loss/cross-tenant/duplicate-delivery, `P1` contract/
  correctness/recovery/major UX, `P2` maintainability/minor UX. Every finding includes
  exact file/line/symbol, failing scenario, evidence, minimal remediation and gate impact.
- Distinguish observed fact, inference and recommendation. Absence of evidence is not a
  defect claim unless the contract requires that evidence.
- Return `PASS` only with no unresolved P0/P1 in scope. P2 is listed and explicitly routed.

## R148-A — architecture and source-of-truth review

> Model: `gpt-5.6-terra` · Effort: `high`

Review package boundaries, dependency direction, transactions and source of truth:

- engine owns raw MIME/body/mailbox membership/server flags/delivery queue;
- Hub owns tenant mapping, safe projections, calendar/RSVP, labels/search/audit;
- no calendar↔mail import cycle or Fiber/JMAP/Ent leakage into pure domain ports;
- selected Stalwart domain verification is represented by an explicitly accepted authority;
  domain existence/zone text cannot satisfy it. Account/address management plus mailbox
  `Identity/get|set` is a credential-separated, Get-before-retry mapping with no fabricated ids;
- receipt/state ledgers have distinct responsibilities and atomic boundaries;
- outbox write shares the state transaction; no post-commit durability claim;
- ScheduledCall↔CalendarEvent, CRM↔MailMessage and notification relations do not introduce
  destructive cascades or implicit authorization;
- linked meetings have one canonical writer, event-before-call lock order, bridge-only
  mutations, all-or-nothing cutover and single-count aggregate/availability readers;
- background workers have bounded leases/retries, context shutdown and one composition;
- provisioning intent/identity ownership, generic mutation receipt and desired/deployed
  rule generations have distinct atomic boundaries and recovery owners;
- mailbox Changes/enumeration adapter proves a composite catch-up base and resumable stable
  query anchor; invalid cursor rebuild never pre-clears or finalizes partial enumeration;
- deployment/backup separates engine volume from PostgreSQL/object storage.

Required cross-check: every implementation file maps to one accepted dispatch packet and
no packet silently widened a schema/public API/dependency.

## R148-B — backend correctness and recovery review

> Model: `gpt-5.6-terra` · Effort: `high`

Review state machines, concurrency, bounds and error semantics:

- draft/save/submission idempotency and unknown reconciliation;
- account provisioning, create/bulk receipt replay/compaction, folder pending-intent
  exact-match/definitive-absence recovery and rule-generation Get-before-retry/stale-owner
  recovery;
- typed message/thread bulk expansion and explicit spam/ham feedback distinct from move;
- mailbox projection page/event/cursor atomicity, command-lease begin/end fencing, expiry/
  late-owner recovery, invalid-cursor enumeration/catch-up recovery, provenance-safe
  received-event cutoff/marker and multi-replica worker lifecycle;
- REQUEST/CANCEL lower/equal/newer, tombstone and missing-master occurrence handling;
- inbound REPLY sender-bound receipt cap and A→B→A protection;
- outward REPLY delivery ledger and stable mailbox keys;
- event/attendee optimistic versions and sequence/reset policy;
- ScheduledCall mapping, recurring semantic oracle, paired create/edit/delete atomicity,
  mandatory legacy-create idempotency in both clients, exact-match attach, series-only
  participant projection, occurrence-cancel projection advancement and versioned
  occurrence-keyed due-worker recovery;
- availability half-open intervals, recurrence, all-day, IANA/DST and caps;
- reminder claim/receipt/outbox/dedupe, current offset/start/eligibility recheck at both
  scheduler and consumer, all-day due derivation, strict no-fallback quiet policy and
  serialized mutation race;
- feed snapshot/conditional refresh if included;
- CRM link idempotency, existing-message owned-deal picker and no message hydration on denial;
- NFR-13 body plus exact headers/statuses and legacy shape compatibility.
- exact route-specific mailbox 409 codes, including non-disclosing global domain/address
  unavailable semantics; no UI depends on localized error text.

Mutation/transaction failures must leave no partial receipt, event, attendee, projection,
outbox, CRM link or falsely completed delivery.

## R148-C — security and privacy review

> Model: `gpt-5.6-terra` · Effort: `xhigh`

Threat-review every trust boundary:

- organization/mailbox owner derivation and cross-user opaque ids;
- admin middleware before probes/details plus named audit actions;
- body/MIME/recipient/BCC/secret/idempotency/feed-URL storage/log/error boundaries;
- server-side HTML sanitization, reader CSP/sandbox and no browser original-URL fetch;
- server-derived plain content, reviewed-sanitizer-or-plain-only behavior and no client
  rich/plain dual authority;
- remote-image proxy pre/post-connect SSRF, DNS rebinding and size/time/type caps;
- attachment namespace, hash, type, filename, scanner fail-closed and presign reauth;
- send rate/recipient/message limits and auto-response loop headers;
- iTIP parser bounds, auth alignment, spoof/forward/delegation, UID/sequence/attendee rules;
- feed encryption/AAD, redirects, safe dialer, decompression and conditional headers;
- availability HR/calendar detail minimization;
- ScheduledCall attach/cutover active project/organization and creator exact-channel checks,
  linked-external-guest fail-closed behavior, strict conflict UUID parsing and exact-channel
  authorization before any legacy conflict title/channel disclosure;
- CRM permission not becoming mailbox delegation;
- event/webhook payloads and dead-letter/audit redaction;
- Stalwart license/image/SBOM/signature/config/secret and network exposure gates, including
  the exact management permission allowlist, separate mailbox auth, authoritative DNS
  verification and absence of recovery-admin credentials/listeners in steady state.
- application and ingress logging never persist remote-image token paths, search queries or
  concrete UUID paths; unmatched route labels are constant rather than attacker input.

Any cross-tenant read/write, arbitrary URL, active-content execution, raw credential/body
leak, unauthenticated calendar mutation or duplicate send path is P0.

## R148-D — frontend workflow/accessibility review

> Model: `gpt-5.6-terra` · Effort: `high`

Review the actual UI against the accepted design and API:

- TanStack Query is server truth; URL is selection truth; composer reducer is local only;
- three-pane/deep-link/back-forward/mobile behavior and exact navigation order;
- provisioning status, folder/label management, rule deployment state and admin aliases;
- 401 retry preserves idempotency/If-Match; versioned folder/settings/signature/label/rule/
  policy/domain conflicts refetch without overwrite; pending folder controls never claim
  completion; same-origin canonical Location validation;
- draft autosave single-flight/conflict/unknown recovery and send receipt polling;
- safe reader/remote image/attachment state boundaries, same-object presign replay and
  byte-only proxy headers/no redirect;
- invitation sequence/occurrence/refetch/reply-failed states;
- call calendar consumes canonical-backed occurrences and does not expand a second browser
  RRULE or show cancelled/duplicate linked meetings; `call.scheduled_due.v1` invalidates the
  authorized query without a duplicate reminder/toast;
- partial bulk rollback and search cursor/error-vs-empty behavior;
- typed message/thread target preservation, stable create/bulk keys, classifier-training
  outcome and body/attachment-filename search without client content persistence;
- conservative folder/thread/search projection freshness, stale/rebuilding banner and
  query-only retry with no client-side worker or false-ready transition;
- CRM creates one ordinary Mail draft, never puts recipient/content in URL/history, and
  links only after a confirmed/owned message; draft/message deep links reauthorize through
  Mail and ephemeral deal intent is not authorization;
- loading/error/empty, focus movement, keyboard shortcuts, live regions, 44px touch target,
  dark/light/reduced motion/responsive layouts;
- no localStorage/Zustand/body leakage, second mailbox cache or generated route manual edit.

Visual/browser evidence is `NOT_RUN` unless produced by the accepted launcher/browser
workflow against the exact SHA.

## R148-E — product/scope/compatibility review

> Model: `gpt-5.6-terra` · Effort: `high`

Review whether the feature solves the frozen user outcome without scope drift:

- Outlook-class owned mailbox and inline calendar invitation loop are coherent;
- external guests remain email-only for ordinary events; a linked call explains its blocked
  guest-access state until secure join credentials exist; L2/L3/read-receipts/OOO/groups stay
  deferred;
- availability is useful without privacy leakage; reminder guarantee is communicated;
- CRM link is manual and the mailbox remains the sender;
- L1 is either explicitly included with all gates or explicitly deferred;
- no marketing/newsletter/digest/tracking behavior or implicit broadcast reuse;
- legacy calendar aggregate and flat Resend admin contract remain compatible;
- linked event/call pair appears once per reader, with event as business truth and the
  existing call experience retained;
- failure, stale, unknown, blocked and delivery states are honest to the user;
- wording never claims module/MVP/production complete or stable without all evidence.

Founder questions that materially change behavior remain blocking rather than guessed.

## R148-F — QA/traceability/evidence review

> Model: `gpt-5.6-terra` · Effort: `high`

Review FR/TC/API/event/source-to-evidence mapping and the gate report:

- every implemented FR-MAIL/FR-CALSYNC and CRM case maps to feature units plus the correct
  Python contract/API/integration/E2E/security/load handoff layer;
- no component/API/integration behavior is mislabeled as a genuine Go/React unit;
- protected testframework paths are unchanged and future cases are not invented there;
- every report matches exact SHA/launcher/images/config and has complete cleanup evidence;
- skips are contractually expected/cited, retries obey policy, stale evidence is rejected;
- contract schemas/examples and source surfaces agree;
- recovery matrix and failure injection cover every ambiguous side effect;
- documentation states PASS/FAIL/BLOCKED/NOT_RUN consistently.

## R148-G — final convergence review

> Model: `gpt-5.6-terra` · Effort: `xhigh`

Run only after A–F findings are fixed/re-reviewed and final rebase is complete. Inspect the
full current diff plus final evidence manifest. Confirm:

1. no unresolved P0/P1 and every P2 has explicit disposition;
2. no fix invalidated another review or contract artifact;
3. protected paths, existing Resend/legacy calendar and product-truth boundaries hold;
4. all required gates in `GATE-MATRIX.md` are current-SHA PASS or correctly BLOCKED;
5. the proposed report claims no more than the lowest satisfied release level.

Output one of `PASS_FOR_LOCAL_HANDOFF`, `BLOCKED`, or `FAIL`. It cannot authorize push,
PR, merge, deployment, DNS, license or production cutover.

## Finding format

```text
ID: R148-<lens>-NN
Severity: P0|P1|P2
Gate: G148-NN
Fact: exact observed behavior
Evidence: file:line/symbol or report artifact
Scenario: minimal failure/reproduction
Impact: user/data/security/recovery consequence
Required fix: smallest contract-compliant change
Re-review: lens and evidence that must be repeated
```

Final packet output also includes reviewed SHA, actual model/effort, files read, runtime
evidence status and explicit unknowns.
