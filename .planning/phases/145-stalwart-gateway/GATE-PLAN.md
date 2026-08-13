# Phase 145 — external and runtime gate plan

> **Status:** BLOCKED_BY_144_ACCEPTANCE_AND_EXTERNAL_GATES
> **No implementation authorization:** this plan does not approve a license, image,
> network egress, DNS change, secret, deployment or runtime command.

## Gate order

| Gate | Evidence required | Owner / authority | Current state | Failure action |
|---|---|---|---|---|
| G145-0 engine boundary | independently accepted projection-only pages, bounded content/blob request shape, hard admin maxima and stable over-limit behavior | founder + architecture/security review | 0.1.8 ROOT_FROZEN / INDEPENDENT_REVIEW_PENDING | no fake/adapter implementation against an unaccepted port |
| G145-1 license | written choice: unmodified sidecar under AGPL-3.0 or accepted SELv2/commercial terms | founder/legal | UNRESOLVED | select docker-mailserver fallback behind the same port; re-plan adapter |
| G145-2 runtime model | exact `gpt-5.6-terra` / `xhigh` requested and actual provenance for each code packet | founder/runtime operator | ROUTE_AUTHORIZED / new packet actual telemetry unavailable | reject any mismatch or fallback |
| G145-3 launcher | versioned task-owned outer launcher implementing limits, timeout, report owner marker and exact cleanup | test architecture owner | NOT_READY | no direct Docker/Compose fallback; runtime remains NOT_RUN |
| G145-4 provenance | refreshed official release, immutable digest, signature/SBOM/source and approval record | DevOps/security/founder | NOT_RUN | do not pull or reference mutable image |
| G145-5 config contract | reviewed Stalwart config/API/session/health/volume paths for the selected immutable version | adapter reviewer | NOT_RUN | stop rather than guess version-specific fields |
| G145-5A folder writer exclusivity | reviewed config, internal-network policy and published-port evidence prove only Hub can create/rename/delete custom folders; system roles are distinguishable | adapter + security reviewer | NOT_RUN | capability stays false; automatic folder recovery makes zero mutation and remains degraded |
| G145-6 local profile | bounded health, authenticated JMAP, internal delivery, restart persistence and exact cleanup | accepted launcher | NOT_RUN | gateway remains disabled/not_configured |
| G145-7 resource budget | observed whole added profile ≤4 CPU and ≤4 GiB under idle + representative load | accepted launcher | NOT_RUN | reduce/rebalance or reject candidate |
| G145-8 backup/restore | dedicated engine-volume backup and restore exercise with current SHA/digest | ops + accepted launcher | NOT_RUN | no production readiness claim |
| G145-9 external mail | approved domain/DNS/operator, PTR coordination, SPF/DKIM/DMARC/MX and controlled deliverability exercise | founder/operator | UNRESOLVED | keep external mode disabled |

Gates are sequential where evidence depends on an earlier choice. G145-1 never derives
from repository text. G145-4 must be refreshed immediately before an authorized pull/run;
the 2026-08-13 research snapshot is not an immutable-image approval.

## Disabled/default behavior

- `MAILBOX_ENGINE_ENABLED=false` produces `not_configured` and does not affect existing
  Resend alert readiness.
- Enabled but unreachable/misconfigured engine produces mailbox API `503`/degraded health;
  server readiness policy must be explicitly wired and tested, not silently reported ready.
- `MAILBOX_EXTERNAL_ENABLED=false` remains the default even after local engine setup.
- `MAILBOX_EXCLUSIVE_FOLDER_WRITER=false` remains the safe default. It may be true only in
  an accepted profile that publishes no user-facing IMAP/JMAP/ManageSieve or direct
  engine-admin mailbox mutation surface.
- Missing domain or secret refuses external mode without printing values.
- Invalid/expired JMAP cursor marks the affected mailbox projection degraded and triggers
  a bounded owner-scoped rebuild. It does not silently reset freshness.
- Enabled startup owns exactly one bounded domain-control, account-control, combined projection/recovery and rule-deployment
  worker on the process root context; disabled startup owns none. Shutdown cancels and
  drains them through registered reverse-order cleanup.

## Secret and endpoint rules

- JMAP/admin base URLs come only from trusted process configuration and resolve to the
  fixed internal service; no request parameter or domain record can redirect them.
- Administrator credentials, DKIM private keys and generated tokens are secret-provider
  or engine-volume state. Example files contain names/placeholders only.
- Logs, health and API responses redact authorization headers, session capability URLs,
  account ids if credential-like, raw provider bodies and SMTP recipient/body content.
- Selected reverse-proxy/access-log configuration logs no query string for mail search and
  no concrete remote-image token path (constant route label or disabled path logging).
  Application route-template logging is necessary but not accepted as proxy proof.
- Production transport ports are not published by default. Wave 1 may explicitly publish
  SMTP delivery ports after G145-9; it never publishes user-facing IMAP/JMAP/ManageSieve or
  direct engine-admin mailbox mutation. Internal JMAP/admin exposure is service-network
  scoped to Hub. Any alternate writer keeps the exclusivity capability false.

## Runtime evidence manifest

When authorized, the accepted launcher report must record exact repository SHA, launcher
version, image tag + digest, config checksum without secrets, container/resource names,
limits, timestamps, wall-clock timeout, cases, observed CPU/RSS and unconditional cleanup.
Required cases:

1. disabled/not_configured startup and unchanged Resend alert routes;
2. engine health and authenticated JMAP session discovery;
3. account provisioning and default folders;
3b. domain create/DNS-required/verification/delete desired state reconciles through
    Ensure/Get without automatic DNS change or private DKIM exposure;
3a. provisioning lost response repeats the same stable key, two replicas create one logical
    account/primary identity, and disabled/missing-domain paths make zero engine calls;
4. internal submit creates one Sent and one recipient Inbox view;
5. lost-submit response reconciles without duplicate;
6. draft version conflict and lost-save reconciliation;
7. change cursor advance, invalid cursor degraded state, base-state/ID-anchored enumeration,
   query-state drift catch-up, typed restart on anchor/base-history loss, and command-lease
   begin/during/end exclusion from sync capture/apply;
8. crash/resume during enumeration and catch-up, two-replica lease exclusion, worker context
   cancellation/drain, expired command-lease reclaim/late-owner reconciliation,
   engine-volume persistence and Hub projection recovery;
8a. fixed Hub Sieve generation/hash apply, lost-response Get-before-retry, stale-generation
    non-finalization and external script drift repair; explicit spam and ham feedback proves
    accepted/rejected/indeterminate outcomes separately from Junk folder movement;
8b. selected-version feedback idempotency/reconciliation capability is evidenced; if
    absent, a lost response stays terminal unknown and the stable command is never retried;
8c. folder create persists one pending intent before engine I/O; with proven exclusive
    writer capability a lost response survives restart and exact non-system match/
    definitive-absence recovery cannot duplicate the folder. Rename/delete recover only the
    persisted desired intent. False/unknown capability, an exposed alternate writer,
    system-role candidate or ambiguous state makes zero recovery mutation and remains
    degraded;
8d. enabled startup owns exactly four worker schedulers (domain, account, combined
    projection/recovery, rule), while disabled startup owns zero; recovery and sync make
    bounded progress without holding both account leases or starving either work class;
9. backup, destructive-in-container data loss simulation and restore;
10. exact task-resource cleanup even after timeout/failure;
11. idle and representative load resource measurement;
12. optional external SMTP only after G145-9, reported separately.
