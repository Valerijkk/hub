# Phase 148 — merge-readiness gate matrix

> **Status:** GATE_CONTRACT_PREPARED · IMPLEMENTATION_NOT_STARTED
> **Scope:** Mail/Calendar/CRM feature branch only
> **Evidence rule:** exact current SHA or no PASS
> **Execution rule:** accepted versioned Docker launcher only; host/direct Compose evidence
> is inadmissible

This matrix separates a reviewable branch from an operating mail service. Passing a source
or merge gate never means the unfinished Hub MVP, mail module, or production cutover is
complete/stable/ready.

## 1. Verdict vocabulary

| Verdict | Meaning |
|---|---|
| `PASS` | All required evidence exists for the exact reviewed SHA and has no unresolved P0/P1 |
| `FAIL` | Evidence ran/reviewed and proves a contract violation |
| `BLOCKED` | A required authority, dependency, model, launcher or external system is absent |
| `NOT_RUN` | Executable evidence was not attempted through an accepted path |
| `N/A` | Contract proves the gate does not apply; reason and approving artifact are cited |
| `STALE` | Evidence targets an ancestor/different config/image and cannot support current SHA |

`PARTIAL` is a report-level summary only; it is never a gate result and never upgrades a
missing gate. `PASS_WITH_WARNINGS` is forbidden. P2 observations may remain as explicitly
accepted backlog only after every P0/P1 is closed.

## 2. Release levels

| Level | Required result | What may be claimed |
|---|---|---|
| Contract ready | G148-01..03 PASS | frozen reviewed contract; no implementation/runtime claim |
| Source ready | G148-01..06 PASS | current diff is review-clean and model/provenance compliant |
| Local merge ready | G148-01..16 PASS, G148-17 authorized or explicitly awaiting user action | branch can be handed to founder for push/PR/merge |
| Operational evidence complete | Local merge ready + G148-11/12/13 external/runtime cases applicable to selected mode PASS | bounded environment behaved as specified on that SHA/config |

No level in this document grants production cutover. Existing project-wide ops blockers,
secrets, DNS/operator approval and CTO/founder sign-off remain separate.

## 3. Gate table

| ID | Required evidence | Current state | Blocks |
|---|---|---|---|
| G148-01 baseline | clean isolated branch; base/current SHA; worktree path; changed-file and protected-zone diff; no untracked product artifact | PASS for docs snapshot only; repeat after code | every later gate |
| G148-02 contract | ADR, Mail spec, OpenAPI, event schema, UI contract, Phase 147 contract and QA handoff independently agree; machine refs/examples closed | Mail 0.1.5 root/independent current-provenance review PENDING; 0.1.4 `46cfa26c` is STALE, so gate remains open | source work/claim |
| G148-03 founder decisions | signed answers for exact Spark runtime, migration mechanism, Stalwart license, L1 include/defer, recurring-call semantics, pause/resume, legacy POST idempotency break, linked-call guest access, broadcast disposition, rebase signal; push/PR separately authorized | BLOCKED/UNRESOLVED | affected packets and final gate |
| G148-04 model provenance | every production/test-code packet records actual `gpt-5.3-codex` + `xhigh`; no silent fallback; review packets record allowed Terra effort | BLOCKED: exact Spark unavailable | source acceptance |
| G148-05 migrations/generation | reviewed mechanism, preflight, generated Ent twice clean, complete ScheduledCall mapping + recurrence oracle + 100% bridge cutover dry-run on empty/realistic snapshot, counts/hashes and rollback | NOT_RUN | backend/runtime |
| G148-06 independent source review | backend, frontend, architecture, security, product and QA lenses on current diff; all P0/P1 fixed and re-reviewed | NOT_RUN | merge-ready |
| G148-07 launcher integrity | versioned outer launcher has task owner marker, CPU/RAM/PID limits, wall timeout, isolated report root and unconditional exact cleanup | NOT_READY | all executable evidence |
| G148-08 units/build/contracts | applicable backend/frontend feature units, generation, build/type/lint and schema/source-contract jobs green through launcher | NOT_RUN | merge-ready |
| G148-09 API/integration | PostgreSQL constraints/transactions, public API/authz, NATS/outbox, Meili, SeaweedFS and mailbox fake/adapter integration green | NOT_RUN | merge-ready |
| G148-10 security | owner isolation, admin RBAC/audit, HTML/CSP, image SSRF, attachment scan, rate/loop, iTIP spoof/order, feed SSRF if included, CRM cross-user denial | NOT_RUN | merge-ready |
| G148-11 engine/runtime | accepted license/version/digest/config; health/JMAP; internal delivery; optional approved external delivery; persistence, resource and backup/restore evidence | BLOCKED/NOT_RUN | operational evidence; selected engine acceptance |
| G148-12 recovery/reconciliation | all ambiguity, duplicate, out-of-order, cursor rebuild, outage and projection recovery cases pass | NOT_RUN | merge-ready where engine-independent; operational where engine-dependent |
| G148-13 user flows/quality | E2E smoke/full as applicable, keyboard/a11y, responsive UI, bounded load and no existing-flow regression | NOT_RUN | merge-ready |
| G148-14 broadcast governance | route removed/disabled and verified, or founder-approved superseding consent/audit/rate/unsubscribe/abuse contract fully implemented | BLOCKED_BY_FOUNDER_13 | merge-ready |
| G148-15 docs/QA | system/architecture/security/test/runbook/state/roadmap/API docs match current code; protected QA handoff updated without editing its program | docs only; repeat after code | merge-ready |
| G148-16 final rebase/current SHA | founder confirms protected test-rebase is in master; mail branch rebased; only mail-owned conflicts; all applicable gates rerun on resulting SHA | BLOCKED_BY_FOUNDER_9 | final handoff |
| G148-17 external repository action | founder explicitly authorizes push and draft/ready PR separately; emitted artifacts/URL match actual action | NOT_AUTHORIZED | push/PR only |

## 4. Evidence manifest

Every accepted launcher report and review aggregate records:

- report id, task/packet id, exact branch, repository SHA, merge-base SHA and clean/dirty
  state before/after;
- requested and actual model/reasoning provenance for authored code and each review;
- launcher version/SHA, profile, start/end UTC, wall timeout, host-independent tool/image
  versions and artifact hashes;
- immutable container image tags + digests, signature/SBOM/provenance verdict and selected
  engine/license decision reference;
- redacted configuration checksum, feature flags and names-only secret inputs; no values;
- exact task-owned container/network/volume/report names, CPU/RAM/PID limits and observed
  idle/representative CPU/RSS;
- commands/jobs/cases, pass/fail/skip counts, duration, retry count and full safe report
  paths; expected skip needs a cited gate;
- DB migration revision/preflight counts, generated-diff result, fixture/data snapshot id and
  rollback outcome;
- cleanup inventory in `finally`: every exact task resource absent, report owner marker
  retained, unrelated resources untouched;
- changed/protected path inventory, secret scan, AI/coauthor trailer scan and doc-sync map;
- unresolved findings with severity, owner, disposition and blocking gate.

Missing fields make the evidence `FAIL` or `STALE`, never assumed.

## 5. Mandatory recovery matrix

| Case | Required observable result |
|---|---|
| draft save loses response | draft becomes unknown; next operation reconciles opaque engine version; no blind overwrite |
| send loses response | same `MailSubmission` reconciles; exactly one visible sent message |
| outbox publish redelivers | consumers remain idempotent; no second state mutation/in-app row |
| versioned envelope transport | shared builder is deterministic; outbox event type, NATS subject and JSON subject match; full envelope bytes are published unchanged; application `event_id` deduplicates under same/different `Hub-Event-Id`; mismatch/malformed/unknown version dead-letters before mutation; legacy outbox payloads remain unchanged |
| invalid mailbox cursor | mailbox projection is degraded, bounded owner rebuild occurs, freshness is not forged |
| engine unavailable | mailbox mutations return safe unavailable/unknown semantics; cached projections are visibly stale; Resend readiness unchanged |
| projection/search lag | authorized engine result may appear with indexing state; no unauthorized hit/id/snippet; rebuild converges |
| REQUEST/CANCEL reorder | lower/equal conflict cannot roll state back; early CANCEL tombstone survives delayed mail |
| REPLY A→B→A | attendee remains B; durable reply receipts/cap/outbox match contract |
| outward REPLY timeout | hidden draft/submission reconciles; no duplicate external reply |
| occurrence before master | deferred source rehydrates; newer tombstone/order is rechecked |
| unbound calendar-state cap | previously unseen UID mutates nothing and quarantines; known UID continues; degraded health/ops signal is bounded and content-free; no tombstone eviction |
| reminder duplicate/policy-store outage | one durable intent plus one permitted in-app row or one durable suppression; policy is not re-evaluated after terminal outcome; transient lookup/store failure stays pending and retries; post-commit WS failure preserves the row; Telegram/Web Push start only after commit, rollback starts none, and their guarantee remains best-effort |
| reminder setting/state race | scheduler and consumer each recheck current sequence, effective start, non-null offset, recomputed due and eligibility; consumer resolves channel/quiet policy under the stable user-row guard; mutation-first suppresses with no row, consumer-first preserves only its committed valid row |
| reminder receipt coherence | pending has no outbox/notification; emitted requires outbox; delivered requires notification/no reason; suppressed has no notification/one closed reason; expired uses suppressed + `expired`; no terminal tuple reopens |
| all-day reminder boundary | ordinary midnight, overlap, midnight gap and wholly skipped date resolve to the frozen first-representable-instant rule in the event IANA zone without changing date-only storage/API/iTIP |
| reminder invalid policy | invalid timezone or incomplete/malformed quiet hours yields durable `policy_invalid`, no row, bounded content-free signal and no UTC/delivery fallback; transient lookup/store remains pending/retriable |
| CRM compose double-click/lost response/navigation | one normal draft and one stable idempotency key/body while unknown; recipient appears only in request body; failed navigation reuses returned draft UUID; no send/link until explicit normal-Mail actions |
| CRM foreign message | generic not-found/forbidden boundary, zero engine/body hydration, no activity row |
| CRM existing-message link | ordinary reader offers only current-user owned deals; explicit duplicate PUT converges; forged/admin-visible unowned deal remains server-denied with no content copy |
| legacy call backfill | full non-deleted set, including paused/completed rows, maps org/project/channel/identity/content/reminder/recurrence exactly; one bad row blocks all; soft-deleted counts are fixed/non-reactivatable; no historical outbox |
| recurring call semantic drift | raw worker, injected-UTC aggregate and selected canonical future occurrence oracles agree; mismatch/cap/unsupported row blocks apply |
| paired meeting create/edit/delete | event, attendees, call projection and invitation outbox commit together; v2/legacy concurrency locks event before call and leaves one authority |
| legacy create retry/attach | both shipped clients reuse one key through ambiguous retry; exact replay returns one pair; changed input conflicts; attach requires exact locked-call equality and never overwrites a mismatch |
| linked-call membership/guest | creator is an active exact-channel member; internal attendees are active same-project/same-org members; external attendee fails closed pending guest join credentials while an ordinary unlinked event remains valid |
| series/occurrence participant projection | series `needs_action|accepted|tentative` adds/keeps the internal participant; declined/removed removes it; occurrence RSVP never changes the call edge |
| occurrence cancel/call due | cancelled next occurrence never fires; projection advances atomically; duplicate worker/outbox delivery uses one occurrence key and never changes event/reminder state; `call.scheduled_due.v1` causes authorized canonical refetch without another reminder |
| linked-pair readers | legacy aggregate emits one canonical-backed `call`; availability emits one `event`; no duplicate busy interval or ghost occurrence |
| legacy call detail privacy | title/channel detail requires caller membership in the exact conflicting channel; Mail UI uses private v2 availability |
| feed refresh failure | if L1 included, last complete snapshot remains, safe stale/error, no cross-source rewrite |
| restart | engine volume, Hub projections, receipts and workers recover without duplicate mutation |
| backup/restore | dedicated engine volume plus Hub stores restore in isolation and reconcile |

Every case must identify which source of truth is checked: engine mail state, Hub PostgreSQL,
outbox/JetStream, search projection, object storage, notification row or UI query cache.

## 6. Runtime suites

Only the accepted launcher selects exact jobs. The expected layers are:

1. feature-owned Go/React genuine units;
2. OpenAPI/event/source contracts;
3. API authn/authz/idempotency/error-shape cases;
4. PostgreSQL/Ent constraints, transactions, migrations and backfill;
5. NATS outbox/consumer duplicate/dead-letter behavior;
6. fake engine plus selected adapter/JMAP/internal SMTP runtime;
7. SeaweedFS attachment authorization/scanner and Meili projection isolation;
8. browser E2E/a11y/responsive flows;
9. bounded recovery/outage/resource/load and exact cleanup;
10. conditional external SMTP/DNS and L1 feed suites only after their gates.

The protected Python QA owner chooses final file/case placement. This branch supplies the
requirements and evidence request; it does not edit that program.

## 7. Final rebase protocol

1. Obtain explicit founder confirmation that the protected test-rebase changes are in the
   target master. Do not infer from dates or local refs.
2. Record pre-rebase branch/base/current SHA and clean state.
3. Rebase only the mail branch. Do not modify or resurrect protected testframework,
   phases 128–141, TASK-193–234, legacy runners or deleted support paths.
4. Resolve only mail-owned conflicts against current product truth. A changed test layout,
   launcher API or generated policy returns to its owner instead of being guessed.
5. Record post-rebase SHA/diff and rerun G148-01, 02, 05–15 on that SHA. Ancestor evidence
   is `STALE`.
6. Prepare the merge report locally. Push/PR waits for G148-17.

## 8. Current honest aggregate

`CONTRACT_REVIEW_PENDING`: Mail 0.1.5 amended the ScheduledCall authority, migration and
reader/worker contracts, so the 0.1.4 pass on `46cfa26c` is stale. Root and independent
current-provenance reviews are required before G148-02 can close. Product code is not
implemented, the exact code-writing runtime is unavailable, the launcher is not ready,
engine/legal/migration/broadcast/rebase decisions are unresolved, and all executable evidence is
`NOT_RUN`.
