# Phase 148 — merge-readiness gate matrix

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No source/plan/test/runtime/product/phase/release authority changes. Root-only separate carrier-plan acceptance, G145-5B/G145-5C and launcher readiness keep Rule Task 1/145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Current upstream lifecycle (2026-08-15):** Carrier contract `1daef78768048d7672c38a65b78ef0f119f3c207` is docs/design-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; its base/correction lifecycle is `CONSUMED/HISTORICAL`. No plan/source/runtime/release authority follows. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending a separately accepted carrier implementation plan, G145-5B, G145-5C and launcher readiness. Launcher `NOT_READY`; executables `NOT_RUN`.

> **Status:** G148_02_BLOCKED_BROADER_CROSS_CONTRACT_CURRENT_SOURCE_EXECUTABLE_GAPS · FIVE_DOCUMENT_144_01A_ENTRY_AMENDMENT_PASS_5E314FB0 · 144_01A_STATIC_ACCEPTED_E3E8C774 · 144_03A_THREAD_STATIC_ACCEPTED_9903F4DF_EXACT_SHA_PASS · IMPLEMENTATION_IN_PROGRESS · RUNTIME_NOT_RUN
> **Scope:** Mail/Calendar/CRM feature branch only
> **Evidence rule:** exact current SHA or no PASS
> **Execution rule:** accepted versioned Docker launcher only; host/direct Compose evidence
> is inadmissible

This matrix separates a reviewable branch from an operating mail service. Passing a source
or merge gate never means the unfinished Hub MVP, mail module, or production cutover is
complete/stable/ready.

> **Upstream lifecycle update:** 144-02D remains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former lifecycle is `CONSUMED/HISTORICAL`. Phase 148 remains pending/blocked; no release or runtime acceptance follows.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

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
| G148-02 contract | ADR, Mail spec, OpenAPI, event schema, UI contract, Phase 147 contract and QA handoff independently agree; machine refs/examples closed | BLOCKED: 144-02A representation alone is docs-only `STATIC_CONTRACT_READY` at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931`; its source schemas are unimplemented/source review absent, 02B/C/D/F remain pending, broader cross-contract review and executable schema/example validation remain absent, and G145-5C authoritative manual-DNS verification is open | source work/claim |
| G148-03 founder decisions | task-local Terra route is answered; signed answers remain required for migration mechanism, Stalwart license, L1 include/defer, recurring-call semantics, pause/resume, legacy POST idempotency break, linked-call guest access, broadcast disposition, primary-address policy, sanitizer dependency, Sieve ownership, engine hydration/stream shape, message/attachment hard ceilings, mailbox quotas/pagination, calendar overflow behavior, Stalwart mailbox auth, DNS verification authority/projection and rebase signal; push/PR separately authorized | PARTIAL / unresolved items remain | affected packets and final gate |
| G148-04 model provenance | every production/test-code packet records exact requested `gpt-5.6-terra` and `xhigh`, no fallback, and either matching actual telemetry or an accepted native `provenance_source: dispatcher-enforced` record; review packets record exact allowed Terra effort and distinct review context | PASS for the provenance rule: for this branch/Phases 143–148, an explicit native dispatcher acceptance without rejection is sufficient where telemetry is unavailable; any telemetry mismatch, rejection, omitted model/effort or fallback remains `BLOCKED` | source acceptance |
| G148-05 migrations/generation | reviewed mechanism, preflight, generated Ent twice clean, complete ScheduledCall mapping + recurrence oracle + 100% bridge cutover dry-run on empty/realistic snapshot, counts/hashes and rollback | NOT_RUN | backend/runtime |
| G148-06 independent source review | backend, frontend, architecture, security, product and QA lenses on current diff; all P0/P1 fixed and re-reviewed | BLOCKED: 143-03A/C and 144-01A have narrow static evidence; 144-02A representation is docs-only accepted but its source schemas are unimplemented and source review absent, 02B/C/D/F schemas remain unimplemented, and no full current-source multi-lens review exists | merge-ready |
| G148-07 launcher integrity | versioned outer launcher has task owner marker, CPU/RAM/PID limits, wall timeout, isolated report root and unconditional exact cleanup | NOT_READY | all executable evidence |
| G148-08 units/build/contracts | applicable backend/frontend feature units, generation, build/type/lint and schema/source-contract jobs green through launcher | NOT_RUN | merge-ready |
| G148-09 API/integration | PostgreSQL constraints/transactions, public API/authz, NATS/outbox, Meili, SeaweedFS and mailbox fake/adapter integration green | NOT_RUN | merge-ready |
| G148-10 security | owner isolation, admin RBAC/audit, HTML/CSP, image SSRF, attachment scan, rate/loop, iTIP spoof/order, feed SSRF if included, CRM cross-user denial | NOT_RUN | merge-ready |
| G148-11 engine/runtime | accepted license/version/digest/config; least-privilege credentials; authoritative domain verification; exact two-plane identity mapping; health/JMAP; internal delivery; optional approved external delivery; persistence, resource and backup/restore evidence | BLOCKED: selected `x:Domain` has no manual-DNS verification result and mailbox auth is unresolved; exact candidate `Identity/get|set` source is researched, but all runtime evidence is NOT_RUN | operational evidence; selected engine acceptance |
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
- requested model/effort and provenance source for authored code and each review; actual
  telemetry and its evidence when exposed, which must match the requested route;
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
| draft save/delete/send race | update/delete/new-key send require exact last-confirmed If-Match; lost save becomes unknown and reconciles; matching send receipt replays before current-version check; no blind overwrite/delete/second submit |
| send loses response | same `MailSubmission` reconciles; exactly one visible sent message |
| outbox publish redelivers | consumers remain idempotent; no second state mutation/in-app row |
| versioned envelope transport | shared builder is deterministic; outbox event type, NATS subject and JSON subject match; full envelope bytes are published unchanged; application `event_id` deduplicates under same/different `Hub-Event-Id`; mismatch/malformed/unknown version dead-letters before mutation; legacy outbox payloads remain unchanged |
| incremental mailbox page | hydrated projection, complete envelopes and exact composite next cursor commit atomically under live sync lease/prior cursor/sync+mutation epochs and absent command lease; any mismatch/rollback advances none |
| command/sync interlock | command or reconciliation acquires one account mutation lease and begin epoch before I/O, blocks sync capture, then clears the exact lease and advances the end epoch; active contention makes zero engine call, expired reclaim invalidates late finalization and forces reconciliation |
| domain/account control | domain intent uses Ensure/Get with public DNS only and no DNS mutation; eligible account scan reserves one immutable address/key/cutoff; two replicas/lost response yield one account/identity set; alias/user lifecycle generations fence stale workers; public status is read-only/redacted |
| generic create/bulk receipt | every named route requires one stable cryptographic v4/v7 UUID; scoped key digest + raw-keyed canonical request HMAC prevent small-payload enumeration; same input replays/reconstructs, changed input is `KEY_REUSED`, removed target is `REPLAY_UNAVAILABLE`, terminal metadata compacts 30 days after first terminal commit to `REPLAY_EXPIRED`, nonterminal never compacts and unknown reconciles; none executes again; presign replay refreshes only the same pending/unexpired object URL; no raw key/body/response persists |
| folder unknown recovery | unique pending row precedes create I/O; busy changes no receipt/control state; update/delete create an internal unexposed receipt only after lease + optimistic checks; combined projection/recovery worker may use exact non-system match/definitive absence and captured receipt only under proven Hub-exclusive custom-folder writing; false/unknown capability, alternate writer, system role or ambiguity makes zero recovery mutation and stays degraded; later command cannot duplicate/overwrite an unresolved folder |
| rule deployment | HTTP commits desired generation without engine I/O; worker owns fixed script under mutation lease, Get-before-retry reconciles lost response, stale generation cannot finalize, external drift is detected/repaired |
| junk/ham feedback | typed message/thread expansion snapshot stays bounded/persisted and per-target; replay never re-expands; classifier accepted/rejected/unknown is distinct from move/flag, stable command ids are reconciled only when capability is proven, and receipt replay never double-trains |
| invalid mailbox cursor | current owner rows remain visibly degraded; base-state/ID-anchored enumeration resumes without pre-clear, query-state drift converges through Changes, catches up from its base cursor and finalizes only when complete; anchor/base-history loss causes typed restart rather than forged freshness |
| received-event replay/provenance | pre-cutoff history and stably mapped Hub submission/draft emit none; spoofed owner From/alias or mutable Sent membership cannot suppress; first-seen post-cutoff inbound-or-unknown reserves one marker/event across page replay, rebuild and replica retry |
| mailbox worker lifecycle | exactly one domain, account, combined projection/recovery and rule worker obey independent scan/page/concurrency/timeout/backoff caps and multi-replica lease exclusion; recovery never holds sync+mutation leases together; disabled starts none; reverse cancellation drains without leak |
| engine unavailable | mailbox mutations return safe unavailable/unknown semantics; cached projections are visibly stale; Resend readiness unchanged |
| safe content/remote image | server derives plain from sanitized rich or fails plain-only; hostile fetch/active content is removed; path accepts only unpadded base64url 32..1024 and rejects all other input before decode/I/O; token binds owner/message/expiry; every redirect/connect is SSRF-safe; response is bytes with fixed private cache/nosniff/no-referrer/CSP and never redirects; browser/API/log/search never sees origin URL |
| sensitive HTTP metadata | app access log/generic audit use matched route templates and constant unmatched label; image token/UUID/query absent; ingress independently proves remote-token path and search-query redaction |
| attachment confirm/download | changed confirm metadata/state and expired lease are no-scan named conflicts; non-clean cannot download; clean 302 is fixed-storage HTTPS only with no-store/no-referrer and no arbitrary origin input |
| versioned mailbox resources | stale folder/label/settings/signature/rule/policy/domain update and versioned delete mutate nothing; one auth retry preserves exact If-Match; pending folder intent blocks overwrite and active-rule label delete returns `MAIL_LABEL_IN_USE` without join/marker/rule loss |
| projection/search lag | bounded sanitized body and normalized filenames index idempotently; no raw HTML/MIME/BCC/URL or unauthorized hit/id/snippet; current owner is rechecked and rebuild converges |
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
| calendar/legacy writer conflict | every 409 keeps NFR-13 body; stale RSVP/event/occurrence exposes its exact code plus applicable current ETag and same-origin event Location; event/occurrence terminal and legacy terminal transitions expose their exact not-mutable code; all paths mutate nothing and clients never parse error text |
| legacy create retry/attach | both shipped clients reuse one key through ambiguous retry; exact replay returns one pair; changed input conflicts; attach requires exact locked-call equality and never overwrites a mismatch |
| linked-call membership/guest | creator is an active exact-channel member; internal attendees are active same-project/same-org members; external attendee fails closed pending guest join credentials while an ordinary unlinked event remains valid |
| series/occurrence participant projection | series `needs_action|accepted|tentative` adds/keeps the internal participant; declined/removed removes it; occurrence RSVP never changes the call edge |
| occurrence cancel/call due | cancelled next occurrence never fires; projection advances atomically; duplicate worker/outbox delivery uses one occurrence key and never changes event/reminder state; `call.scheduled_due.v1` causes authorized canonical refetch without another reminder |
| linked-pair readers | legacy aggregate emits one canonical-backed `call`; availability emits one `event`; no duplicate busy interval or ghost occurrence |
| legacy call detail privacy | title/channel detail requires caller membership in the exact conflicting channel; Mail UI uses private v2 availability |
| feed refresh failure | if L1 included, last complete snapshot remains, safe stale/error, no cross-source rewrite |
| restart | engine volume, Hub projections, enumeration/catch-up checkpoints, command/sync leases, receipts and workers recover without duplicate mutation or stale late finalization |
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

`FIVE_DOCUMENT_144_01A_ENTRY_AMENDMENT_PASS_144_01A_STATIC_ACCEPTED_IMPLEMENTATION_IN_PROGRESS`:
the exact five-document default-vs-hard-cap/144-01A entry amendment passed at
`5e314fb080d420696642058ae47397e251c23207`; this is not a full Mail 0.1.8, Phase 142 or global
cross-contract PASS. The 144-01A bounded engine DTO/error boundary is `STATIC_ACCEPTED` at
`e3e8c774d734f926f77f8e600b58dc68dabbd993`. G148-02 remains blocked only for broader
cross-contract/current-source review, executable schema/example validation and Stalwart
authoritative manual-DNS verification. 144-01B remains blocked under stricter Task3 authority
pending applicable supported current-SHA Docker evidence. Thread is narrowly `STATIC_ACCEPTED`
  at `9903f4df` / `EXACT_SHA_PASS`; packet eligibility is status only. Folder source+oracle are
  narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
  `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
  C/I/M `0/0/0`), static only. Folder docs are also static-only accepted at
  `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`; Rule is separately docs-only
  `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4`, sole-path blob `b86c0423a307fd2673a3d322435922aa2a398e8f`, not a release acceptance. The prior `f199546513b758ca7b95328b5ccd7680009b732e` / `6735ae746c7acde657359e8a709cd897c4dce61c` lifecycle is `CONSUMED/HISTORICAL`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until all independent Phase-145 carrier/G145-5B/G145-5C/launcher prerequisites; no execution. The launcher is
`NOT_READY`, engine/legal/migration/broadcast/rebase decisions are unresolved, and all executable
evidence is `NOT_RUN`.

`144-03B-Rule-Compiler` plan is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
`669ef761715ec35a638bc87d8396e6f562b162c4` (blob `b86c0423a307fd2673a3d322435922aa2a398e8f`), not a release PASS; the prior candidate lifecycle is `CONSUMED/HISTORICAL`. Rule Task 1 remains blocked on the engine-port/145 carrier/G145-5B/G145-5C/launcher prerequisites before
Phase 148 acceptance. The launcher remains `NOT_READY` and
executable evidence `NOT_RUN`.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan-next text is `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. Carrier contract `1daef787` is docs/design-only accepted; Rule plan `669ef761` does not authorize Task 1. The separate carrier implementation-plan/source gate, G145-5B, G145-5C and launcher readiness remain open. 144-01B remains Docker-blocked, Phase 142/148 and release pending, launcher `NOT_READY`, executables `NOT_RUN`.
