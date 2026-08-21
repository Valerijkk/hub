# Autonomous Mail + Calendar questions — 2026-08-21

These questions require founder authority and are intentionally deferred while autonomous work
continues on unblocked pure-domain packets. They do not authorize source, migration, launcher, or
runtime work by themselves.

## Pending founder questions

### AQ-01 — Ent migration and launcher authority

Choose the canonical production migration mechanism and task-owned Docker launcher authority for
Calendar `143-03E` and Mail `144-02E`. Until this is decided and accepted, Ent generation,
migrations, concrete Ent adapters, and the final integrated Docker acceptance remain
`BLOCKED / NOT_READY / NOT_RUN`.

### AQ-02 — existing recurring ScheduledCall interpretation

Confirm the canonical timezone/recurrence interpretation for already persisted recurring
`ScheduledCall` rows (founder question 18), including DST gap/fold handling and backfill behavior.
This blocks production backfill and concrete persistence integration only. It does not block the
pure `143-04J` ScheduledCall projection policy, which will receive normalized caller-owned inputs.

## Root decisions made autonomously

- Proceed with a separate docs-only ABI/readiness freeze for `143-04J`, then a test-first exact2
  pure policy packet if independent reviews pass.
- Do not implement Calendar `143-04D` in its current two-file form: a separate optional RSVP
  transaction-backend ABI and expanded exact manifest are required to preserve authorization,
  receipt replay, atomic attendee mutation, and outbox boundaries.
- Do not start Mail `144-04H` or Calendar `143-04F-*` before the accepted `02E` / `03E` generation
  gates; no handwritten generated Ent or migration fallback is allowed.
