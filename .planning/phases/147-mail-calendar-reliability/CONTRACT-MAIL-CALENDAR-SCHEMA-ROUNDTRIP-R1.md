# Mail/Calendar Atlas roundtrip reconciliation R1 contract

**Status:** FROZEN IMPLEMENTATION CONTRACT. This artifact authorizes one narrow
forward-only source packet after independent implementation and review. It is
not an apply authority, Docker evidence, product-readiness claim, retained
database upgrade approval, or production runbook.

**Provenance:** `provenance_source: dispatcher-enforced`; contract author
`gpt-5.6-terra` / `high`; no fallback. Runtime telemetry is `UNKNOWN` until the
separate permitted disposable Docker verification completes.

## Problem and source facts

The checksum-reconciled seven-file lineage applies cleanly on a fresh
verification database through `20260827000002`, but the independent current
Ent snapshot does not roundtrip. The bounded Atlas comparison reports exactly
four changes:

1. replacement of `calendar_invite_deliveries_state_shape`;
2. replacement of `calendar_reply_deliveries_state_shape`;
3. rename-equivalent foreign key
   `calendar_events_scheduled_calls_calendar_event` to
   `calendar_events_scheduled_calls_calendar_event_link`;
4. rename-equivalent foreign key
   `mail_attachments_mail_messages_attachments` to
   `mail_attachments_mail_messages_attachment_records`.

The two foreign-key names drifted after the baseline/core lineage: commit
`9c083b70` renamed the MailMessage edge to `attachment_records`, and commit
`23aaca14` finalized the ScheduledCall ref as `calendar_event_link`. Neither
commit changed the referenced columns or referential actions. The
`materialization_invalid` terminal error is already present in both current Ent
schemas and `20260827000001_calendar_delivery_materialization_invalid.sql`;
this packet introduces no new delivery-state behavior or data transformation.

## Scope and ownership

The single future implementation atom owns exactly these files:

| File | Permitted change |
|---|---|
| `backend/migrations/atlas/mail-calendar-v1/20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql` | Add the one append-only reconciliation migration defined below. |
| `backend/docker/entrypoints/mail-calendar-schema.sh` | Extend only hash/verify explicit lineage admission from seven SQL files to eight and synchronize version/status/snapshot constants. |
| `backend/migrations/README.md` | Document the eight-file bounded lineage and this reconciliation's limited proof boundary. |

The implementation atom must not edit `atlas.sum`, any of the prior seven SQL
files, Ent schemas, generated Ent, Compose, materializer code, `run_baseline`,
`run_diff`, `run_generate`, dependencies, locks, tests, or `pro-qa-ai/**`.

The separately reviewed checksum atom owns only
`backend/migrations/atlas/mail-calendar-v1/atlas.sum` after the new migration
has landed and `schema-maintenance-hash` emits its candidate. It is not part of
this implementation atom.

## Required migration

Create exactly
`20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql`. Do not
modify, rename, reorder, squash, or regenerate any earlier lineage member. Its
complete SQL is:

```sql
ALTER TABLE "calendar_invite_deliveries"
  DROP CONSTRAINT "calendar_invite_deliveries_state_shape";

ALTER TABLE "calendar_invite_deliveries"
  ADD CONSTRAINT "calendar_invite_deliveries_state_shape"
    CHECK (
      COALESCE((
        (((state)::text = 'pending'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND terminal_at IS NULL AND next_attempt_at IS NOT NULL AND (((attempts = 0) AND last_error_code IS NULL) OR ((attempts BETWEEN 1 AND 8) AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('internal_error'::character varying)::text])))))
        OR (((state)::text = 'submitting'::text) AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NULL)
        OR (((state)::text = 'unknown'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND terminal_at IS NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('submission_outcome_unknown'::character varying)::text])))
        OR (((state)::text = 'sent'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL)
        OR (((state)::text = 'superseded'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL)
        OR (((state)::text = 'failed_permanent'::text) AND lease_token IS NULL AND lease_until IS NULL AND next_attempt_at IS NULL AND ((last_error_code)::text = ANY (ARRAY[('submission_rejected'::character varying)::text, ('invalid_recipient'::character varying)::text, ('materialization_invalid'::character varying)::text, ('policy_denied'::character varying)::text, ('retry_exhausted'::character varying)::text])) AND terminal_at IS NOT NULL)
      ), false)
    );

ALTER TABLE "calendar_reply_deliveries"
  DROP CONSTRAINT "calendar_reply_deliveries_state_shape";

ALTER TABLE "calendar_reply_deliveries"
  ADD CONSTRAINT "calendar_reply_deliveries_state_shape"
    CHECK (
      COALESCE((
        (((state)::text = 'pending'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND (((attempts = 0) AND last_error_code IS NULL) OR ((attempts BETWEEN 1 AND 8) AND next_attempt_at IS NOT NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('internal_error'::character varying)::text])))))
        OR (((state)::text = 'submitting'::text) AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND mail_submission_id IS NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'unknown'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('submission_outcome_unknown'::character varying)::text])))
        OR (((state)::text = 'sent'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'superseded'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'failed_permanent'::text) AND lease_token IS NULL AND lease_until IS NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND ((last_error_code)::text = ANY (ARRAY[('submission_rejected'::character varying)::text, ('invalid_recipient'::character varying)::text, ('materialization_invalid'::character varying)::text, ('policy_denied'::character varying)::text, ('retry_exhausted'::character varying)::text])) AND (((last_error_code)::text <> 'retry_exhausted'::text) OR attempts = 8))
      ), false)
    );

ALTER TABLE "calendar_events"
  RENAME CONSTRAINT "calendar_events_scheduled_calls_calendar_event"
  TO "calendar_events_scheduled_calls_calendar_event_link";

ALTER TABLE "mail_attachments"
  RENAME CONSTRAINT "mail_attachments_mail_messages_attachments"
  TO "mail_attachments_mail_messages_attachment_records";
```

The two CHECK definitions are frozen in the Atlas-emitted current-snapshot
canonical form: every `state` and `last_error_code` comparison casts the column
to `text`, state literals are `'value'::text`, and every `ANY` array element is
`('value'::character varying)::text`. Do not replace that element-wise form
with an array-wide cast, omit casts, reorder alternatives, or otherwise
reformat the predicate. Apart from that canonicalization, the predicates are
the current Ent business expressions exactly, including
`materialization_invalid`.

The foreign keys must use `RENAME CONSTRAINT`, never a drop/add pair. This is a
metadata-only reconciliation preserving the same foreign-key columns and
referential actions.

The CHECK replacement is normal validated `ADD CONSTRAINT`, not `NOT VALID`.
No backfill, data rewrite, or custom preflight query belongs in this packet. On
a retained database, PostgreSQL validates the replacement against existing
rows; an incompatible row must fail the normal transactional migration and
roll back the preceding drop rather than leave a weakened or unvalidated
constraint. The existing `20260827000001` business rule already admits
`materialization_invalid`, so this packet is not expected to invalidate
conforming retained rows. Any actual retained-data failure is an operations
blocker requiring a separate reviewed repair, not a reason to weaken this
migration.

## Launcher synchronization

In `backend/docker/entrypoints/mail-calendar-schema.sh`, add the explicit
constant:

```sh
readonly ROUNDTRIP_RECONCILIATION_FILENAME=20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql
```

Retain the seven existing literal filename constants. Rename the seven-file
admission and copy helpers to eight-file equivalents, or otherwise make their
names truthful. Both `run_hash` and `run_verify` must admit exactly these eight
SQL files in version order plus `atlas.sum`, with no glob or directory-derived
authority:

```text
20260823000000_mail_calendar_baseline.sql
20260823000001_mail_calendar_core.sql
20260826000000_calendar_deferred_receipt_expiry.sql
20260826000001_calendar_inbound_message_receipt_event_message_idx.sql
20260827000000_mail_account_password_enrollment.sql
20260827000001_calendar_delivery_materialization_invalid.sql
20260827000002_calendar_invite_sequence_non_negative.sql
20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql
```

Set these exact values:

```sh
readonly EXPECTED_LINEAGE_SQL_FILE_COUNT=8
readonly EXPECTED_LINEAGE_ENTRY_COUNT=9
readonly EXPECTED_LATEST_VERSION=20260829000000
readonly SNAPSHOT_FILENAME=20260829000000_current_snapshot.sql
```

The eight-file copy helper must perform an explicit copy of the new migration
after the seven existing files and must never copy source `atlas.sum`. Preserve
all current regular-file, symlink, extra-entry, pin, network, database,
materializer, review-size, semantic-diff, and status guards. In particular,
`run_diff` remains its deliberate historical baseline/core-only mode and is not
widened into a repair path.

## Documentation and checksum order

Update `backend/migrations/README.md` to list the eight exact SQL files and
`atlas.sum`, distinguish the historical seven-file evidence from the pending
eight-file reconciliation, and state that a future clean verification proves
only bounded disposable lineage replay and current-schema convergence.

After the implementation atom is independently accepted, the required order is:

1. run only `docker compose -f docker-compose.local.yml run --rm schema-maintenance-hash`;
2. preserve the bounded emitted `atlas.sum` review record;
3. have two independent exact-byte reviewers compare the candidate with the
   proposed replacement, including final newline and exact eight-file regular
   directory membership;
4. in a separate metadata-only atom, replace only `atlas.sum` with the accepted
   candidate;
5. run only `docker compose -f docker-compose.local.yml run --rm schema-maintenance-verify`.

The clean verification acceptance requires Atlas validation and lint for eight
versions, lineage status at `20260829000000` with eight executed and zero
pending files, a one-file current snapshot at `20260829000000`, and empty
forward and reverse semantic diffs. It is not evidence for a shared or
production apply, backend build, product tests, external mail delivery, merge,
or deployment.

## Static acceptance and blockers

Before the separate runtime phase, source review must prove all of the
following:

- only the three implementation-owned files changed and `atlas.sum` did not;
- the previous seven SQL files remain byte-identical;
- the new migration contains exactly two validated CHECK replacements and two
  FK renames, with no data mutation, provider, queue, or application behavior;
- canonical CHECK expressions match this contract byte-for-byte after normal
  LF source normalization;
- hash/verify require eight explicit SQL inputs, nine directory entries, latest
  version `20260829000000`, eight executed files, and the new snapshot name;
- `run_diff`, baseline/generate behavior, tool pins, resource limits, and
  Compose topology remain unchanged.

`BLOCKED` conditions are a missing or extra lineage entry, a change to an old
SQL byte, a manually guessed checksum, a CHECK expression differing from this
canonical contract, use of `NOT VALID`, a retained-data validation failure, a
foreign-key drop/add substitution, any proposed shared/production target, or
any unavailable/mismatched model provenance. Each condition requires a new
reviewed decision; none may be bypassed by changing an old migration or
relaxing the launcher.
