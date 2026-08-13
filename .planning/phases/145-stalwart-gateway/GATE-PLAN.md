# Phase 145 — external and runtime gate plan

> **Status:** BLOCKED_BY_EXTERNAL_GATES
> **No implementation authorization:** this plan does not approve a license, image,
> network egress, DNS change, secret, deployment or runtime command.

## Gate order

| Gate | Evidence required | Owner / authority | Current state | Failure action |
|---|---|---|---|---|
| G145-1 license | written choice: unmodified sidecar under AGPL-3.0 or accepted SELv2/commercial terms | founder/legal | UNRESOLVED | select docker-mailserver fallback behind the same port; re-plan adapter |
| G145-2 runtime model | actual `gpt-5.3-codex` / `xhigh` provenance for each code packet | founder/runtime operator | UNAVAILABLE | no production/test code mutation |
| G145-3 launcher | versioned task-owned outer launcher implementing limits, timeout, report owner marker and exact cleanup | test architecture owner | NOT_READY | no direct Docker/Compose fallback; runtime remains NOT_RUN |
| G145-4 provenance | refreshed official release, immutable digest, signature/SBOM/source and approval record | DevOps/security/founder | NOT_RUN | do not pull or reference mutable image |
| G145-5 config contract | reviewed Stalwart config/API/session/health/volume paths for the selected immutable version | adapter reviewer | NOT_RUN | stop rather than guess version-specific fields |
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
- Missing domain or secret refuses external mode without printing values.
- Invalid/expired JMAP cursor marks the affected mailbox projection degraded and triggers
  a bounded owner-scoped rebuild. It does not silently reset freshness.

## Secret and endpoint rules

- JMAP/admin base URLs come only from trusted process configuration and resolve to the
  fixed internal service; no request parameter or domain record can redirect them.
- Administrator credentials, DKIM private keys and generated tokens are secret-provider
  or engine-volume state. Example files contain names/placeholders only.
- Logs, health and API responses redact authorization headers, session capability URLs,
  account ids if credential-like, raw provider bodies and SMTP recipient/body content.
- Production transport ports are not published by default. Every published SMTP/IMAP port
  is explicit, documented and scoped to the selected operating mode.

## Runtime evidence manifest

When authorized, the accepted launcher report must record exact repository SHA, launcher
version, image tag + digest, config checksum without secrets, container/resource names,
limits, timestamps, wall-clock timeout, cases, observed CPU/RSS and unconditional cleanup.
Required cases:

1. disabled/not_configured startup and unchanged Resend alert routes;
2. engine health and authenticated JMAP session discovery;
3. account provisioning and default folders;
4. internal submit creates one Sent and one recipient Inbox view;
5. lost-submit response reconciles without duplicate;
6. draft version conflict and lost-save reconciliation;
7. change cursor advance, invalid cursor degraded state and bounded rebuild;
8. restart persistence for engine volume and Hub projection recovery;
9. backup, destructive-in-container data loss simulation and restore;
10. exact task-resource cleanup even after timeout/failure;
11. idle and representative load resource measurement;
12. optional external SMTP only after G145-9, reported separately.
