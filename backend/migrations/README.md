# Versioned production migrations

Ent schemas in `backend/ent/schema/**` are the desired-state source of truth.
Atlas only derives deterministic SQL diffs, performs lint/checksum/apply work,
and writes `atlas.sum`; it is not a second schema authority.

`backend/migrations/` is the sole versioned production migration directory.
Review and check in every SQL migration together with its updated `atlas.sum`.
Production applies only those checked-in, reviewed SQL artifacts.

`client.Schema.Create(ctx)` is development-only and may run only with
`DEV_MODE=true`. Production never uses auto-migration. Generated Ent Go remains
uncommitted and is produced only through the accepted task-owned Docker launcher.

Any data backfill is a separate, reviewed, forward-only artifact after the
schema migration. It must preserve user data and cannot be rolled back by
deleting production rows. Direct host Atlas, Ent, Docker, or database commands
are forbidden; use the accepted launcher when it is ready.

## Current execution status

`MIGRATION_CONTROL_PLANE_STATIC`; launcher `NOT_READY`. Ent generation, Atlas
diff, lint, checksum, apply, database replay, and runtime validation are all
`NOT_RUN`.
