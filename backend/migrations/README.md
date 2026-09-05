# Migration artifacts

## Historical three-file deferred-receipt expiry checkpoint (2026-08-26)

At the time of this checkpoint, the accepted mail-calendar-v1 lineage contained
20260826000000_calendar_deferred_receipt_expiry.sql after the baseline and
core migrations. A broad Atlas-generated candidate was rejected because it
included unrelated Mail constraint/index drop-and-recreate noise. The accepted
file is an exact additive migration: one PostgreSQL DO statement acquires an
ACCESS EXCLUSIVE receipt-table lock, aborts when legacy
deferred_missing_master rows exist, and then adds the nullable object-state
link, exact expiry, four checks, NO ACTION foreign key, and traversal index in
the same transaction.

During that historical three-file checkpoint, schema-maintenance-verify passed
in Docker on 2026-08-26: Atlas lint covered three versions and 260 schema changes;
all three migrations (555 statements)
applied at version 20260826000000 with zero pending files; the one-file
current-schema snapshot converged to the same version; forward and reverse
semantic comparisons were empty. This proves the historical three-file schema
lineage only; it is not evidence for the seven-file lineage. The
run used the disposable schema-maintenance PostgreSQL instance; it is not
evidence of an upgrade over retained production receipts. A live rollout still
requires a verified backup and successful execution of the same fail-closed
preflight on the target database. The
P3-D repository/service/worker, 147-02B/02C, D1/D2, full application runtime,
and product readiness remain NOT_RUN or pending.

Ent schemas in `backend/ent/schema/**` are the desired-state source of truth.
Atlas may derive candidate SQL, lint it, maintain checksums and apply an accepted
versioned lineage; it is not a second schema authority. Generated Ent Go remains
uncommitted and may be produced only inside the accepted bounded Docker operation.
The static `schema-maintenance-generate` operation first verifies the LF-normalized full
`go.mod`/`go.sum` lock and exact Atlas/Ent pins, requires `GOWORK=off`, and rejects
`replace`, `exclude`, symlinks and nonregular inputs. It copies only `go.mod`, `go.sum`,
`ent/generate.go` and `ent/schema` into two fresh isolated writable module roots and runs the
fixed offline `go run -mod=readonly entgo.io/ent/cmd/ent generate --feature sql/lock ./ent/schema`; repository
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
  20260826000000_calendar_deferred_receipt_expiry.sql
  20260826000001_calendar_inbound_message_receipt_event_message_idx.sql
  20260827000000_mail_account_password_enrollment.sql
  20260827000001_calendar_delivery_materialization_invalid.sql
  20260827000002_calendar_invite_sequence_non_negative.sql
  20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql
  20260829000001_mail_calendar_check_roundtrip_flattening.sql
  20260831000000_qa_project_outbox_authority.sql
  20260831000001_outbox_event_origin.sql
  atlas.sum
```

## P1a project-outbox authority addition (2026-08-31)

The append-only P1a migration creates only qa_project_outbox_manifests and
qa_project_outbox_effects. It records exact UUID authority and 32-byte hashed
control material; it contains no raw nonce, payload, name, key, email, or
discovery field. The only foreign key is immutable effect.manifest_id to the
manifest with ON DELETE NO ACTION. No foreign key or Ent edge targets User,
Organization, Session, Project, ProjectMember, or OutboxEvent.

P1a advances the bounded mail-calendar-v1 lineage from nine SQL migrations and
ten total entries to ten SQL migrations and eleven total entries, ending at
20260831000000. The schema-maintenance launcher copies and allowlists that
exact file, uses the new version for its current-schema snapshot, and hashes
the complete ten-file SQL input.

## O1 OutboxEvent origin addition (2026-08-31)

The append-only O1 migration adds nullable immutable origin_kind and origin_id
columns to outbox_events, a both-null-or-both-set check, and the non-unique
origin_kind/origin_id/event_type lookup index. It adds no foreign key, Ent edge,
domain enum database constraint, trigger, backfill, data scan, or QA table.

O1 advances the bounded mail-calendar-v1 lineage from ten SQL migrations and
eleven total entries to eleven SQL migrations and twelve total entries, ending
at 20260831000001. The schema-maintenance launcher copies and allowlists that
exact file, uses the new version for its current-schema snapshot, and hashes
the complete eleven-file SQL input.

## P1a/O1 Docker evidence (2026-09-01)

All commands used the current source through `docker-compose.local.yml` and
completed serially:

- `schema-maintenance-generate --build`: PASS; two isolated Ent generations
  matched, generated manifest 79671 bytes with SHA-256
  `49602e8cf48d5e5682c0076e4f4202fbe2163f4997b329867ba1ce2ae3388b43`.
- `schema-maintenance-hash --build`: PASS; the candidate `atlas.sum` is 1159
  bytes with SHA-256
  `8c4aff6c69e2aa555a89e11859d83f4a939a9341c8e68880791a24fc6823a481`
  and header `h1:HK2ODb264K0k6PtJPE4xmiifzOprM5H3JUxiWA04z5Y=`.
- `schema-maintenance-verify --build`: PASS; 11/11 migrations applied, desired
  and replay root-FK projections both contained 82 rows at SHA-256
  `d594dbcd8308675a89875ab8e6ff3a0f60e97c8ecd7696aebf6a562c5341876c`,
  and both bounded Atlas semantic diffs were empty.
- `backend-unit --build`: PASS; build, vet, unit, and race receipts all exited
  zero.

## Occurrence RSVP schema-maintenance evidence (2026-09-05)

The additive receipt-scope migration advances the bounded lineage to 13 SQL migrations and current version `20260905000000`. The Docker-only maintenance operations passed:

- `schema-maintenance-generate --build`: two Ent generations matched; generated manifest was 79,671 bytes with SHA-256 `51f14255590757512584a2905f162edb232436289c4d04228b6038ab75a6c3b8`.
- `schema-maintenance-hash`: candidate `atlas.sum` was 1,375 bytes with SHA-256 `d810480fcbfd8fed08d06d4bb1dfd0d598579e7d0c4ee30e8e96088516e04cf7`, header `h1:W7JFK6OA6prf9t/wJMApKbiDOeiaj9p2FdUp3Lk7fIs=`, and 13 SQL lines.
- fresh `schema-maintenance-verify`: PASS; 13 migrations / 591 statements, current `20260905000000`, pending 0, desired and replay root-FK projections 82/82, and expected semantic diffs empty.

This is bounded disposable schema-maintenance evidence only. It is not backend broad-suite, dev-demo/runtime, browser, shared/production apply, commit, merge, release, or external-delivery evidence.

## Historical seven-file checksum and partial replay evidence (2026-08-29)

`schema-maintenance-hash` is an offline, no-database, pinned-Atlas candidate
generator. The pre-R1 checksum packet copied only the reviewed seven SQL files,
never source `atlas.sum`, and emitted only the bounded candidate checksum review
file. That seven-file packet ended at `20260827000002`; it is not evidence for
the pending eight-file reconciliation.

A candidate checksum is not migration authority until independent exact-byte
review and its separate metadata-only apply; a successful hash alone is not
schema-verification or database-apply evidence.

On 2026-08-29, the rebuilt schema-maintenance image and a fresh disposable
database provided bounded seven-file evidence: Atlas analyzed and linted all
seven versions through current version `20260827000002`, and the fresh lineage
apply reached seven executed files with zero pending.

The full `schema-maintenance-verify` operation did not pass. Its final
roundtrip comparison failed because the current Ent snapshot still differed by
exactly four changes: two canonical delivery CHECK replacements and two
foreign-key constraint-name drifts. This evidence proves only bounded
seven-file checksum handling, lint, and fresh disposable lineage replay. It is
not an accepted verify, current-schema convergence, retained-database upgrade
proof, product test, external mail delivery, deployment, or product-readiness
claim.

## Historical R1 reconciliation and accepted nine-file R2 roundtrip evidence (2026-08-29)

`20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql` is an
append-only metadata reconciliation: it replaces two validated delivery-state
CHECK constraints in the current Ent canonical form and renames two foreign-key
constraints without changing their columns or referential actions. It neither
rewrites data nor changes provider, queue, or application behavior.

R1 was the historical eight-file reconciliation step. The append-only
`20260829000001_mail_calendar_check_roundtrip_flattening.sql` then aligned the
two delivery CHECK expressions with PostgreSQL roundtrip canonicalization
without changing their business semantics.

On 2026-08-29, `schema-maintenance-verify` exited 0 against the bounded
disposable nine-file lineage. Atlas `lint9` succeeded; the lineage status was
nine applied at current `20260829000001` with zero pending; and the one-file
current-schema snapshot status was one executed with zero pending. Forward and
reverse Atlas semantic diffs were both empty. The two raw inspected SQL outputs
were both 242505 bytes. Their SHA-256 values were
`ad14eff7a0e6754425176a3d867698cea80c5abeaca3bab052ba8cf45654cf4a`
for the verification schema and
`66762ca84e296884592f0408f9b055f1c06f96f3adc7d0458cc23c90cc034e7d`
for the snapshot-roundtrip schema.

The raw SHA-256 difference is diagnostic serialization evidence, not a
semantic-equivalence gate: it is limited to column ordering in
`calendar_inbound_message_receipts`, `mail_accounts`, and `scheduled_calls`.
There is no type, default, CHECK, foreign-key, index, or bidirectional Atlas
semantic difference. This evidence proves only bounded disposable lineage
replay and current-schema convergence; it is not evidence for retained/shared
or production upgrade safety, backend build, product tests, external mail
delivery, merge, deployment, or production apply.

The baseline is the complete PostgreSQL desired schema from exact pre-feature
master `f21e3f7748eac55a4b16746b078e88801f11829c`. The core migration is the first
additive, allowlisted delta to the reviewed Mail/Calendar Ent schemas. The five
later migrations are separately reviewed forward-only additions; their reviewed
semantics, including any intentional constraint replacement, are not restated by
the core-delta guard. The seven historical SQL files, pending eighth
reconciliation SQL file, and `atlas.sum` each require independent exact-byte
review and authority appropriate to their packet before they become apply
authority.

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
`schema-maintenance`. Its fixed baseline/diff/generate/hash/verify operations use pinned tool and
image digests, non-root read-only containers, CPU/RAM/PID/tmpfs/time limits, no published ports,
no Docker socket, no host or named volumes, no persistent runtime cache, and unconditional
container cleanup. Baseline/diff/verify use an internal-only disposable PostgreSQL peer; generate
and hash use `network_mode: none` with no database environment or dependency. Only generate has
a separate 256 MiB RAM executable TMPDIR/GOTMPDIR; hash retains the common noexec `/work` and
`/tmp` tmpfs mounts and has no executable `/run/go-run-exec`. Source, HOME, XDG and GOCACHE
remain noexec. It does not extend the closed testframework QA catalog.

`client.Schema.Create(ctx)` remains development-only and may run only with
`DEV_MODE=true`; production never auto-migrates. Shared or production apply needs
a current backup, exact database identity, successful preflight/replay evidence,
independent database review and direct founder approval. Recovery after apply is
forward-only and never deletes user data automatically.

Direct host Atlas, Ent, database, build, validator and test commands are forbidden.

## Historical Docker evidence for the pre-reconciliation lineage (2026-08-26)

This dated execution covered only the baseline and core migrations. It is
historical two-file evidence and not three-file or seven-file reconciliation
evidence:

- double offline Ent generation produced 837 image-local files with identical manifest
  SHA-256 `9baabd28f85b00e3c8a808f11773e24e5b56ae37b4d599f68efdcdb9c63e8b83`;
- Atlas lint accepted 2 versions and 259 schema changes;
- fresh apply executed 2 migrations and 554 SQL statements;
- lineage status is version `20260823000001`, 2 executed, 0 pending;
- current-schema snapshot status is 1 executed, 0 pending;
- forward and reverse semantic diffs between the applied lineage and the independent
  round-trip snapshot are both empty.

This is not dependent backend compile, product-test, full-stack runtime, shared/production
apply, merge, push, deploy, or cutover evidence. Generated Ent remains uncommitted and is
still produced only inside the accepted Docker build/generation boundary.
