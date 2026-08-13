# Mail module — founder questions for the next check-in

> **Created:** 2026-08-13
> **Rule:** these questions do not pause autonomous local work. Until answered, the
> documented safe default applies. No push, PR, merge, deployment, DNS, or license
> purchase is performed from these defaults.

| # | Question | Working default while unattended | What changes after an answer |
|---:|---|---|---|
| 1 | Is unmodified Stalwart under AGPL-3.0 acceptable for Hub, or should Hub use SELv2/commercial terms? | Keep Stalwart as conditional preferred adapter; no runtime adoption claim | Approves/rejects Phase 145 engine gate; rejection selects docker-mailserver fallback |
| 2 | Which real instance mail domain and DNS operator will be used? | Use non-routable examples only; config remains empty/disabled | Enables MX/SPF/DKIM/DMARC/PTR runbook and external acceptance |
| 3 | Should external SMTP be disabled by default even after engine setup? | Disabled by default; internal product paths may be enabled independently | Changes admin default and rollout flag |
| 4 | Confirm migration policy: existing ScheduledCall creator and participants become `accepted`; only new invitees start `needs_action`? | Apply this policy in schema/migration design | Changes backfill migration and user-visible RSVP history |
| 5 | Is L1 external iCal feed subscription required before the first merge, or may it follow complete L0 reliability? | L1 remains Phase 147 / P2 after L0 | Changes critical path and security workload |
| 6 | Confirm external email-only guests in v1 despite the older seed wording that excluded guests? | Support nullable-user attendees and iTIP-only guests; enable with external gateway | Could remove guest runtime behavior while retaining forward-compatible schema |
| 7 | Keep read receipts and out-of-office deferred? | Read receipts excluded; OOO only has future loop-prevention seam | Adds privacy/interop UX or auto-response implementation if overridden |
| 8 | Should group addresses (`team@`) ship in the first merge? | Schema-ready only; per-user aliases ship first | Adds group authorization, expansion and bounce semantics |
| 9 | When Claude's test-rebase branch is merged, should I rebase immediately and prepare a local merge-ready report? | Wait for explicit notice that master includes it; never guess | Authorizes the required final rebase/conflict pass |
| 10 | After local verification, may I push the branch and open a draft PR? | No push and no PR | Grants external repository mutation authority |
| 11 | The exact locked `gpt-5.3-codex` runtime is unavailable to both the native dispatcher and this ChatGPT-account CLI. Will you expose that runtime or explicitly amend the routing lock? | Do not write production/test code with Sol or Terra; continue contracts, reviews and dispatch preparation | Unblocks Phase 143+ atomic code execution |
| 12 | Which production migration mechanism is authoritative now: establish versioned Atlas migrations, or temporarily continue reviewed SQL under `backend/migrations/`? | Freeze schemas and the backfill plan only; do not create/apply production migration SQL | Unblocks the founder-reviewed Phase 143 migration artifact and realistic integration oracle |
| 13 | The repository has a live admin broadcast to every active user, conflicting with ADR-NOTIF-01's whitelist/no-marketing rule. Remove/disable it, or explicitly re-authorize bulk mail with consent/audit/unsubscribe controls? | Treat it as deprecated governance gap; do not reuse it; plan disable/removal before mailbox acceptance | Resolves current product/security contradiction and Phase 148 gate |
| 14 | May the mail UI add MSW as a reviewed frontend development dependency, or must it stay on repository-native client/fetch mocks? | Add no dependency; use typed in-process mocks and fixtures | Changes the unit-test harness only, not production behavior |
