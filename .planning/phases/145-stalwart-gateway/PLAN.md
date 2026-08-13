# Phase 145 — Stalwart Gateway and Docker Mail Profile

> Task: TASK-238 · Status: BLOCKED_BY_EXTERNAL_GATES · Depends on: Phase 144

## Hard entry gates

1. Founder/legal chooses AGPL-3.0 or SELv2 terms.
2. Accepted Docker-only launcher exists for this task.
3. Candidate image tag and digest provenance are refreshed immediately before use.

No runtime implementation/claim proceeds by silent fallback.

## Ordered tasks after gates

1. Implement JMAP session/auth request client with SSRF-safe fixed internal base URL and
   redacted errors in `backend/internal/mail/stalwart/**`.
2. Implement account/identity/folder/message/draft/submission/flags/Sieve mappings,
   including opaque draft versions, stable submission keys and explicit submission
   reconciliation; prove adapter contract conformance.
3. Extend admin settings/health without exposing credentials.
4. Add optional `mail` profile, named volumes, healthcheck, restart and explicit CPU/RAM/
   PID limits; add names-only env examples.
5. Treat invalid/expired change cursors as degraded state followed by a bounded,
   owner-scoped projection rebuild; never silently reset and present stale data as fresh.
6. Through accepted launcher prove health, authenticated JMAP, internal delivery,
   instance-domain SMTP, restart persistence, exact cleanup and actual resource use.
7. Document dedicated engine-volume backup/restore and DNS/deliverability prerequisites.

External Internet delivery remains a separate operator acceptance even after local
transport passes.
