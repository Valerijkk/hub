# TASK-241 — Mail security, reconciliation and merge-readiness gate

> Date: **2026-08-13** · Priority: **P0** · Status: **Open** · Phase: **148**

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
- Projection rebuild/reconciliation and engine-unavailable behavior are documented and
  proven through accepted Docker-only runs.
- All feature-owned units are green in Docker; QA handoff covers API/integration/E2E/load.
- `docs/system_specification.md`, `docs/architecture.md`, security/test docs, runbooks and
  the mail-owned `.planning/STATE.md` section match current code and evidence.
- After founder confirms test-rebase merge, branch rebases onto current master and handles
  moved/deleted unit/CI paths without touching protected artifacts.
- No “complete/stable/production-ready” claim is made unless every listed gate has current
  evidence. Push/PR/merge/deploy remain founder-authorized actions.
