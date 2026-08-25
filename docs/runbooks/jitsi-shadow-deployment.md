# Jitsi shadow-deployment preflight

## Status and boundary

This is a gated preflight runbook, not an approved Jitsi deployment procedure.
Phase 150 is **PARTIAL — STATIC_FOUNDATION_ONLY**. LiveKit remains the sole
active and rollback conference provider. The Jitsi configuration is statically
validated but disabled; the provider factory fails closed for `jitsi`.

There is no accepted Jitsi compose overlay: its state is
`NOT_CREATED / UNBOUND`. Task 8 is `BLOCKED_BY_AUTHORIZED_LINUX_DISCOVERY`.
Runtime and media evidence is `NOT_RUN`.

## Preconditions for any continuation

Before any infrastructure work, obtain explicit infrastructure authority and a
numeric disk budget. The authorized Linux discovery must resolve the relevant
host facts and record the outcome against the Phase 150 evidence packet.

Use [the Jitsi deployment preflight](../../deployments/jitsi/DEPLOYMENT-PREFLIGHT.md)
as the controlling checklist. Only after its gates are accepted may an
authorized operator consider a one-digest-at-a-time continuation. This document
intentionally contains no overlay bring-up commands, proxy routes, or ports.

## Explicit non-scope

This preflight does not enable a public credential flow, UI activation,
callbacks, Jibri, recording, a cutover, or removal of LiveKit. The Jitsi issuer
remains internal and server-only: no route, DTO, or membership path exposes it.
