# Migration artifacts

Ent schemas in `backend/ent/schema/**` are the desired-state source of truth.
Atlas may derive candidate SQL, lint it, maintain checksums and apply an accepted
versioned lineage; it is not a second schema authority. Generated Ent Go remains
uncommitted and may be produced only inside the accepted bounded Docker operation.

## Historical root SQL

The SQL files directly under `backend/migrations/*.sql` are historical additive
operator artifacts for databases that already existed when each task was shipped.
They predate Atlas control, use repeated date-only version prefixes, and have no
`atlas.sum`. Consequently, the root collection is not an Atlas-replayable migration
directory and must not be used as the `migrate diff`, lint, checksum, fresh-replay or
apply lineage for Mail/Calendar.

Preserve these files byte-for-byte. Do not rename them, assign synthetic Atlas
versions, re-hash them into a new `atlas.sum`, or replay them as a fresh-database
history. Their presence is historical evidence, not proof of current database state.

## Bounded Mail/Calendar Atlas lineage

The sole future Atlas lineage for this workstream is:

```text
backend/migrations/atlas/mail-calendar-v1/
  20260823000000_mail_calendar_baseline.sql
  20260823000001_mail_calendar_core.sql
  atlas.sum
```

The baseline is the complete PostgreSQL desired schema from exact pre-feature
master `f21e3f7748eac55a4b16746b078e88801f11829c`. The second file is an additive,
allowlisted delta to the reviewed Mail/Calendar Ent schemas. It may not contain
`DROP`, `TRUNCATE`, implicit data deletion, destructive type conversion,
constraint weakening, an unreviewed cascade, or unrelated schema change. Both SQL
files and `atlas.sum` require independent exact-byte review before they become
apply authority.

An empty database applies the baseline normally. An existing database may mark
the baseline as applied only after a separate schema-equivalence report proves it
already matches every baseline object. A mismatch blocks the operation; baseline
marking is not a repair or drift-suppression mechanism.

Any ScheduledCall or Mail data backfill is a separate reviewed forward-only
artifact after the schema migration. It preserves user data and emits no provider,
notification, NATS or outbox side effect. Non-deleted recurring ScheduledCall rows
remain a hard backfill/cutover block until founder question 18 selects their
canonical recurrence semantics.

## Execution boundary

The future execution API is an isolated non-QA Compose profile named
`schema-maintenance`. Its fixed baseline/diff/verify operations use pinned tool and
image digests, non-root read-only containers, an internal-only disposable
PostgreSQL peer, CPU/RAM/PID/tmpfs/time limits, no published ports, no Docker
socket, no host or named volumes, no persistent runtime cache, and unconditional
container cleanup. It does not extend the closed testframework QA catalog.

`client.Schema.Create(ctx)` remains development-only and may run only with
`DEV_MODE=true`; production never auto-migrates. Shared or production apply needs
a current backup, exact database identity, successful preflight/replay evidence,
independent database review and direct founder approval. Recovery after apply is
forward-only and never deletes user data automatically.

Direct host Atlas, Ent, database, build, validator and test commands are forbidden.
The profile and SQL lineage are not implemented yet. Ent generation, Atlas diff,
lint, checksum, apply, database replay, build, tests and runtime validation remain
`NOT_RUN`.
