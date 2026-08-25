# Migration artifacts

Ent schemas in `backend/ent/schema/**` are the desired-state source of truth.
Atlas may derive candidate SQL, lint it, maintain checksums and apply an accepted
versioned lineage; it is not a second schema authority. Generated Ent Go remains
uncommitted and may be produced only inside the accepted bounded Docker operation.
The static `schema-maintenance-generate` operation first verifies the LF-normalized full
`go.mod`/`go.sum` lock and exact Atlas/Ent pins, requires `GOWORK=off`, and rejects
`replace`, `exclude`, symlinks and nonregular inputs. It copies only `go.mod`, `go.sum`,
`ent/generate.go` and `ent/schema` into two fresh isolated writable module roots and runs the
fixed offline `go run -mod=readonly entgo.io/ent/cmd/ent generate ./ent/schema`; repository
`go:generate` directives are never executed. Both all-source pre/post manifests and both
generated-output manifests must match before one bounded SHA-labelled manifest is emitted. It
permits no generated-source egress or staging and is a prerequisite proof, not the dependent
normal Docker compile; runtime evidence remains `NOT_RUN`.

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

## Static pre-feature source controls

SB-A is the committed immutable carrier for pre-feature parent
`f21e3f7748eac55a4b16746b078e88801f11829c`. SB-B statically consumes that exact
carrier in the schema-maintenance image: it verifies the fixed 307200-byte
archive hash, permits only the reviewed regular-file selection, rejects symlinks
and extra entries, and materializes only
`/opt/hub/prefeature/backend`. The fixed `prefeature_baseline` Atlas environment
uses that image-local Ent source and exports only a bounded baseline SQL candidate
plus `atlas.sum` for review output.

This is static source-control implementation, not SQL authoring or execution.
The exact two-file baseline packet remains required at
`backend/migrations/atlas/mail-calendar-v1/20260823000000_mail_calendar_baseline.sql`
and `backend/migrations/atlas/mail-calendar-v1/atlas.sum`; it needs separate
exact-byte review before it gains any migration authority.

Any ScheduledCall or Mail data backfill is a separate reviewed forward-only
artifact after the schema migration. It preserves user data and emits no provider,
notification, NATS or outbox side effect. Non-deleted recurring ScheduledCall rows
remain a hard backfill/cutover block until founder question 18 selects their
canonical recurrence semantics.

## Execution boundary

The future execution API is an isolated non-QA Compose profile named
`schema-maintenance`. Its fixed baseline/diff/generate/verify operations use pinned tool and
image digests, non-root read-only containers, CPU/RAM/PID/tmpfs/time limits, no published ports,
no Docker socket, no host or named volumes, no persistent runtime cache, and unconditional
container cleanup. Baseline/diff/verify alone use an internal-only disposable PostgreSQL peer;
generate has `network_mode: none`, no database environment or dependency, and only a separate
256 MiB RAM TMPDIR/GOTMPDIR is executable while `/work`, `/tmp`, source, HOME, XDG and GOCACHE
remain noexec. It does not extend the closed testframework QA catalog.

`client.Schema.Create(ctx)` remains development-only and may run only with
`DEV_MODE=true`; production never auto-migrates. Shared or production apply needs
a current backup, exact database identity, successful preflight/replay evidence,
independent database review and direct founder approval. Recovery after apply is
forward-only and never deletes user data automatically.

Direct host Atlas, Ent, database, build, validator and test commands are forbidden.
SB-A/SB-B/SB-C are statically implemented, but no Docker image build, offline Ent generation,
dependent compile, Atlas inspect, diff, lint, checksum or apply, generated SQL, database replay,
build, tests or runtime validation has run. All executable and SQL evidence remains `NOT_RUN`.
