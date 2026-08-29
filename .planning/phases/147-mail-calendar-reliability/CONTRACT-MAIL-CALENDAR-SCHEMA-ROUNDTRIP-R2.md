# Mail/Calendar Atlas roundtrip reconciliation R2 contract

**Status:** FROZEN IMPLEMENTATION CONTRACT. This is an append-only correction
to the accepted R1 migration. It authorizes no apply by itself, and is not
Docker evidence, product acceptance, retained-database approval, merge, or
production runbook.

**Provenance:** `provenance_source: dispatcher-enforced`; contract author
`gpt-5.6-terra` / `high`; no fallback. Any unavailable, omitted, rejected, or
mismatched routing evidence is `BLOCKED`. Runtime evidence remains `UNKNOWN`
until the required fresh disposable Docker verification completes.

## Recorded R1 failure and bounded diagnosis

R1 migration plus checksum were statically accepted and committed. The fresh
Docker verification later applied eight migrations, reported current version
`20260829000000`, zero pending migrations, and one current snapshot. Its final
snapshot semantic comparison still failed, but only for these two CHECK
constraints:

1. `calendar_invite_deliveries_state_shape`;
2. `calendar_reply_deliveries_state_shape`.

The direct Atlas verify-to-compare diff confirmed that those are the only two
differences. There is no remaining foreign-key difference. In particular, R1's
foreign-key reconciliation must not be repeated, renamed, dropped, or otherwise
changed by R2.

The cause is a PostgreSQL parse-tree canonicalization difference limited to the
pending retry branch. R1 writes `attempts BETWEEN 1 AND 8`; the live PostgreSQL
parse tree preserves the nested `AND` shape, while snapshot replay canonicalizes
the equivalent comparison as `attempts >= 1 AND attempts <= 8` inside that
branch. The compared `pg_get_expr` lengths differ by two characters. This is a
roundtrip representation issue, not a new business rule, foreign-key drift, or
Ent-schema defect.

## Scope and immutable boundaries

R2 is migration-only at the product-schema level. It makes no Ent schema,
generated Ent, application, queue, provider, materializer, data, or foreign-key
change. It must never rewrite, reorder, regenerate, squash, or amend R1 or any
earlier migration/checksum lineage member.

The future R2 implementation atom may own only:

| File | Permitted change |
| --- | --- |
| `backend/migrations/atlas/mail-calendar-v1/20260829000001_mail_calendar_check_roundtrip_flattening.sql` | Add exactly the append-only SQL frozen below. |
| `backend/docker/entrypoints/mail-calendar-schema.sh` | Extend explicit admission/status constants from eight SQL files/nine entries to nine SQL files/ten entries and the new latest version. |

`backend/migrations/README.md` evidence wording is a separate implementation
documentation update; it is not a license to alter migration semantics, the
launcher beyond this contract, or historical evidence. The later checksum atom
may change only `backend/migrations/atlas/mail-calendar-v1/atlas.sum`, and only
after the required Docker hash output and two independent byte reviews.

Do not edit `atlas.sum` in the R2 implementation atom; any existing migration;
Ent schemas or generated files; Compose; tool pins; locks; tests; snapshot
generator; `run_baseline`, `run_diff`, or `run_generate`; or `pro-qa-ai/**`.

## Required append-only migration

Create exactly
`20260829000001_mail_calendar_check_roundtrip_flattening.sql`. Its complete SQL
is frozen here:

```sql
ALTER TABLE "calendar_invite_deliveries"
  DROP CONSTRAINT "calendar_invite_deliveries_state_shape";

ALTER TABLE "calendar_invite_deliveries"
  ADD CONSTRAINT "calendar_invite_deliveries_state_shape"
    CHECK (
      COALESCE((
        (((state)::text = 'pending'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND terminal_at IS NULL AND next_attempt_at IS NOT NULL AND (((attempts = 0) AND last_error_code IS NULL) OR ((attempts >= 1) AND (attempts <= 8) AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('internal_error'::character varying)::text])))))
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
        (((state)::text = 'pending'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND (((attempts = 0) AND last_error_code IS NULL) OR ((attempts >= 1) AND (attempts <= 8) AND next_attempt_at IS NOT NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('internal_error'::character varying)::text])))))
        OR (((state)::text = 'submitting'::text) AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND mail_submission_id IS NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'unknown'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND ((last_error_code)::text = ANY (ARRAY[('mailbox_unavailable'::character varying)::text, ('engine_unavailable'::character varying)::text, ('submission_outcome_unknown'::character varying)::text])))
        OR (((state)::text = 'sent'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'superseded'::text) AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL)
        OR (((state)::text = 'failed_permanent'::text) AND lease_token IS NULL AND lease_until IS NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND ((last_error_code)::text = ANY (ARRAY[('submission_rejected'::character varying)::text, ('invalid_recipient'::character varying)::text, ('materialization_invalid'::character varying)::text, ('policy_denied'::character varying)::text, ('retry_exhausted'::character varying)::text])) AND (((last_error_code)::text <> 'retry_exhausted'::text) OR attempts = 8))
      ), false)
    );
```

The two new definitions must be normal validated `ADD CONSTRAINT` statements,
not `NOT VALID`. There is no backfill, data rewrite, preflight query, FK rename,
FK drop/add, or data mutation. PostgreSQL must validate each replacement in the
normal transaction; an incompatible retained row is an operations blocker, not
a reason to weaken the constraint.

The only intentional syntactic difference from R1 is in each `pending` branch's
retry alternative:

```sql
((attempts >= 1) AND (attempts <= 8) AND ...)
```

All remaining alternatives and both constraints retain R1's business semantics
and exact forms, including every other `attempts BETWEEN 1 AND 8`, casts,
literal ordering, `ANY` array element form, `materialization_invalid`, and the
reply `retry_exhausted` terminal condition. No cleanup, reformatting, or
canonicalization outside those two pending retry alternatives is authorized.

## Required launcher synchronization

After the migration is accepted, the launcher must admit an explicit ordered
nine-SQL lineage plus `atlas.sum`, with no glob or directory-derived authority:

```text
20260823000000_mail_calendar_baseline.sql
20260823000001_mail_calendar_core.sql
20260826000000_calendar_deferred_receipt_expiry.sql
20260826000001_calendar_inbound_message_receipt_event_message_idx.sql
20260827000000_mail_account_password_enrollment.sql
20260827000001_calendar_delivery_materialization_invalid.sql
20260827000002_calendar_invite_sequence_non_negative.sql
20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql
20260829000001_mail_calendar_check_roundtrip_flattening.sql
```

`run_hash` and `run_verify` must use that same explicit sequence, preserve all
existing regular-file, symlink, unexpected-entry, pin, network, database,
materializer, review-size, semantic-diff, and status guards, and set exactly:

```sh
readonly EXPECTED_LINEAGE_SQL_FILE_COUNT=9
readonly EXPECTED_LINEAGE_ENTRY_COUNT=10
readonly EXPECTED_LATEST_VERSION=20260829000001
readonly SNAPSHOT_FILENAME=20260829000001_current_snapshot.sql
```

The copied/reviewed snapshot must therefore be the one current snapshot at
`20260829000001`. `run_diff` remains its historical baseline/core-only mode;
do not widen it into a reconciliation or repair path.

## Ordered checksum and verification gates

The checksum is a separate metadata-only atom, never a guessed value:

1. independently review and accept the exact R2 migration and launcher atom;
2. run only `docker compose -f docker-compose.local.yml run --rm schema-maintenance-hash`;
3. preserve the emitted candidate `atlas.sum` review record;
4. obtain two independent exact-byte reviews of the candidate, final newline,
   and exact nine-regular-SQL-plus-`atlas.sum` directory membership;
5. in a separate atom, replace only `atlas.sum` with the accepted candidate;
6. run only `docker compose -f docker-compose.local.yml run --rm schema-maintenance-verify`;
7. update README evidence separately and only with the actually observed
   accepted verification result.

The required fresh Docker acceptance is all-or-nothing:

| Check | Required result |
| --- | --- |
| Atlas lint | `lint9` success for the explicit nine-version lineage |
| Lineage status | nine applied, current `20260829000001`, zero pending |
| Snapshot | exactly one `20260829000001_current_snapshot.sql` |
| Semantic compare | forward diff empty and reverse diff empty |
| Inspection parity | compared snapshot inspection size and SHA-256 equal |

This proves only clean disposable Docker lineage replay and bidirectional
current-schema convergence. It does not prove retained/shared/production
upgrade safety, backend build, product tests, external delivery, merge, or
deployment.

## Static acceptance and blockers

Before the Docker phase, reviewers must establish all of the following:

- exactly two constraints are dropped and re-added, and no other constraint or
  foreign key is touched;
- the full SQL matches this contract after LF normalization;
- only the two pending retry alternatives use the flattened comparison;
- prior migrations, including R1, are byte-identical;
- no Ent schema or generated artifact changes;
- launcher counts, explicit membership, latest version, and snapshot filename
  are exactly `9`, `10`, `20260829000001`, and
  `20260829000001_current_snapshot.sql` respectively;
- `atlas.sum` is absent from the implementation diff and remains reserved for
  the later checksum atom.

`BLOCKED` includes an old-migration rewrite; extra or missing lineage member;
manual checksum; any R2 FK operation; a changed non-pending branch; `NOT VALID`;
nonempty diff in either semantic direction; unequal inspection size/hash;
retained-data failure; production/shared target; or unavailable/mismatched
provenance. Each needs a new reviewed decision. None may be bypassed by
relaxing this contract or changing R1.
