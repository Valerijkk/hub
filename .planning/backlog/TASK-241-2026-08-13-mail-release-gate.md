# TASK-241 — Mail security, reconciliation and merge-readiness gate

> Date: **2026-08-13** · Priority: **P0** · Status: **Gate contract prepared** · Phase: **148**

## Outcome

Prove the implemented mail/calendar slices against contracts, close security and recovery
gaps, synchronize documentation, hand off non-unit QA, then rebase after the founder merges
Claude's test-rebase branch.

## Acceptance

- Independent architecture, backend, frontend, security, and product review findings are
  resolved or explicitly blocked.
- Mailbox owner isolation, admin RBAC/audit, HTML/CSP, remote image SSRF proxy, attachment
  scan gate, rate limits, auto-response loop prevention, iTIP spoofing, event idempotency,
  and search authorization are verified at applicable layers.
- The current `/admin/mail/broadcast` governance gap is removed/disabled, or a founder-
  approved superseding contract with consent/audit/unsubscribe/abuse controls is implemented
  and verified; mailbox code never inherits it silently.
- Atomic projection/event/cursor apply, command-lease begin/end fencing, expiry/late-owner
  recovery, base-state/ID-anchored no-clear rebuild/catch-up, received-event cutoff/marker,
  replica lease/shutdown, visible freshness and engine-unavailable behavior are documented
  and proven through accepted Docker-only runs.
- Reminder offset/state changes are serialized against scheduler and consumer finalization;
  existing-message CRM linking is reauthorized at both UI and server boundaries.
- ScheduledCall migration is all-or-nothing; linked writes/workers use event-before-call
  atomicity; recurrence/cancel/due recovery converges; aggregate/availability count one
  meeting once; legacy conflict details cannot leak without exact-channel membership.
  Both shipped legacy-create clients preserve one idempotency key, malformed conflict UUIDs
  fail closed, attach mismatches do not overwrite, and occurrence RSVP cannot mutate the
  series participant projection.
- All feature-owned units are green in Docker; QA handoff covers API/integration/E2E/load.
- `docs/system_specification.md`, `docs/architecture.md`, security/test docs, runbooks and
  the mail-owned `.planning/STATE.md` section match current code and evidence.
- After founder confirms test-rebase merge, branch rebases onto current master and handles
  moved/deleted unit/CI paths without touching protected artifacts.
- No “complete/stable/production-ready” claim is made unless every listed gate has current
  evidence. Push/PR/merge/deploy remain founder-authorized actions.

## Prepared artifacts

- `.planning/phases/148-mail-release-gate/GATE-MATRIX.md`
- `.planning/phases/148-mail-release-gate/REVIEW-PACKETS.md`

Current aggregate remains `BLOCKED/NOT_RUN`: code is absent, exact Spark and the accepted
launcher are unavailable, and founder/external/rebase decisions are unresolved.
