# TASK-238 — Stalwart gateway and Docker mail profile

> Date: **2026-08-13** · Priority: **P0** · Status: **Blocked by gates** · Phase: **145**

## Outcome

After explicit license acceptance and accepted launcher availability, implement the
Stalwart JMAP adapter, account/domain provisioning, instance-domain SMTP send/receive,
health/queue status, persistence, and optional bounded Docker profile.

## Owned paths

- `backend/internal/mail/stalwart/**`
- mail-owned backend config/admin wiring
- `docker-compose.local.yml`
- `.env.local.example`
- mail-owned deployment/runbook additions

## Acceptance

- Founder/legal records the AGPL-3.0 or SELv2 choice; no silent assumption.
- Image is pinned by verified digest at execution time.
- Every added service has profile `mail`, healthcheck, named volume where needed, restart
  policy, explicit CPU/RAM/PID limits, minimal exposure, and env-only secrets.
- Added mail profile totals at most 4 CPU and 4 GiB RAM by configuration and measured
  evidence; shared-service usage is reported separately.
- Accepted launcher proves JMAP discovery/auth, mailbox operations, internal + external
  instance-domain transport, restart persistence, health, and cleanup.
- Failure preserves the `MailEngine` boundary and stops; docker-mailserver fallback needs
  an explicit recorded switch, not silent fallback.

## Blockers

- Founder/legal engine license choice.
- Versioned Docker-only launcher is currently `NOT_READY`.
- Real external delivery additionally needs domain, DNS, PTR, and network authority.
