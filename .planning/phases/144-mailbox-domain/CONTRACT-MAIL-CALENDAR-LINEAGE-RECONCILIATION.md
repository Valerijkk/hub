# Mail/Calendar Atlas lineage reconciliation contract

**Status:** FROZEN IMPLEMENTATION CONTRACT — source inspection only; Docker, Atlas, and database execution are `NOT_RUN`.

**Authority:** Phase 144 mailbox-domain schema maintenance. This contract is the narrow prerequisite before any FTS work. It creates checksum-reconciliation capability only; it does not authorize a schema change, an Ent generation, an application build, a database upgrade, or FTS implementation.

**Provenance:** `provenance_source: dispatcher-enforced`; contract author `gpt-5.6-terra` / `high`; no fallback. Runtime telemetry is `UNKNOWN` until a permitted Docker execution records it.

## 1. Problem statement and fixed source facts

`backend/migrations/atlas/mail-calendar-v1/` contains seven committed SQL migrations, but its committed `atlas.sum` has entries only through `20260826000000`. The four later committed SQL files are not checksum-authorized. The current `run_verify` therefore accepts exactly three SQL inputs plus `atlas.sum`, lints `--latest 3`, and requires current version `20260826000000` with three executed files. It cannot be used as proof for the seven-file directory and must fail closed until reconciliation is complete.

The only SQL inputs for this reconciliation are the following exact ordered allowlist. Order is their version order; no glob, directory iteration, root `backend/migrations/*.sql`, generated migration, or stale `atlas.sum` may become an input authority.

| Order | Exact filename | Existing source commit |
|---:|---|---|
| 1 | `20260823000000_mail_calendar_baseline.sql` | existing lineage |
| 2 | `20260823000001_mail_calendar_core.sql` | existing lineage |
| 3 | `20260826000000_calendar_deferred_receipt_expiry.sql` | `58654b63` |
| 4 | `20260826000001_calendar_inbound_message_receipt_event_message_idx.sql` | `ea1c0473` |
| 5 | `20260827000000_mail_account_password_enrollment.sql` | `b7c2e613` / `3a660e0a` |
| 6 | `20260827000001_calendar_delivery_materialization_invalid.sql` | `e3b54b27` |
| 7 | `20260827000002_calendar_invite_sequence_non_negative.sql` | `39088eac` |

Fixed post-reconciliation facts:

- SQL file count: `7`; lineage directory entry count including `atlas.sum`: `8`.
- Expected latest version: `20260827000002`.
- Verification status: `Migration Status: OK`, current version `20260827000002`, `Executed Files: 7`, `Pending Files: 0`.
- Current-schema snapshot filename: `20260827000002_current_snapshot.sql`; its independent snapshot status remains one executed file at version `20260827000002` with zero pending files.
- Existing SQL bytes are immutable for this packet. The only derived candidate is a replacement `atlas.sum` computed from those seven exact bytes.

## 2. Scope and non-goals

In scope is one bounded offline `hash` launcher operation, its Compose service, the exact verification constants updated to the facts above, and accurate migration documentation. A second, separately reviewed metadata action may replace `backend/migrations/atlas/mail-calendar-v1/atlas.sum` with the emitted candidate. The second action is conditional, not pre-authorized by the first implementation commit.

Out of scope: editing any of the seven SQL files; creating, renaming, reordering, merging, or deleting migrations; changing Ent schemas or generated Ent; `migrate diff`; FTS; materializer logic; `run_baseline`, `run_diff`, or `run_generate`; database access; a shared/production apply; host Atlas; broad rehashing; dependency/image/pin changes; and changes under `pro-qa-ai/**`.

`run_diff` intentionally retains its historical baseline/core-only input contract. It is not a recovery path for this matured lineage and must not be widened as part of this small reconciliation. Its fail-closed behavior with later lineage files is not evidence of a defect in the new offline hash operation.

## 3. File ownership and atomic boundaries

| Packet | Exact files | Permitted change |
|---|---|---|
| A — capability + docs | `backend/docker/entrypoints/mail-calendar-schema.sh`, `docker-compose.local.yml`, `backend/migrations/README.md` | Add bounded hash capability; synchronize seven-file verification constants and documentation. No `atlas.sum` edit. |
| B — reviewed derived metadata | `backend/migrations/atlas/mail-calendar-v1/atlas.sum` | Replace only with the exact bytes emitted by Packet A's `schema-maintenance-hash` run after independent exact-byte review. |
| C — evidence only | a new phase evidence receipt, if required by the maintainer | Record the commands, image identity, candidate review hashes, and Docker result. It must not claim production readiness. |

Packet A is one atomic implementation/review unit. Packet B is deliberately separate: it cannot be staged, committed, or applied merely because Packet A merged. Packet C follows a successful Packet B Docker verification and is not a reason to change SQL or loosen guards.

## 4. Packet A implementation contract

### 4.1 `backend/docker/entrypoints/mail-calendar-schema.sh`

Keep POSIX `/bin/sh`, `set -eu`, current fixed Atlas binary/pin verification, non-root filesystem assumptions, `EX_USAGE=64`, `EX_CONFIG=78`, and `REVIEW_FILE_LIMIT_BYTES=33554432`. Do not add Bash arrays, command substitution of untrusted filenames, `eval`, network calls, a database URL, or a generic migration-directory copy.

Add these explicit immutable constants near the existing lineage constants:

```sh
readonly INBOUND_MESSAGE_RECEIPT_EVENT_MESSAGE_IDX_FILENAME=20260826000001_calendar_inbound_message_receipt_event_message_idx.sql
readonly PASSWORD_ENROLLMENT_FILENAME=20260827000000_mail_account_password_enrollment.sql
readonly DELIVERY_MATERIALIZATION_INVALID_FILENAME=20260827000001_calendar_delivery_materialization_invalid.sql
readonly INVITE_SEQUENCE_NON_NEGATIVE_FILENAME=20260827000002_calendar_invite_sequence_non_negative.sql
readonly EXPECTED_LINEAGE_SQL_FILE_COUNT=7
readonly EXPECTED_LINEAGE_ENTRY_COUNT=8
readonly EXPECTED_LATEST_VERSION=20260827000002
readonly SNAPSHOT_FILENAME=20260827000002_current_snapshot.sql
```

The existing baseline, core, and deferred-receipt-expiry filename constants remain unchanged. Add a single helper which calls the existing `require_exact_files` with `EXPECTED_LINEAGE_ENTRY_COUNT` and all seven explicit SQL filename constants followed by `atlas.sum`. It is the sole source-directory admission control for `run_hash` and `run_verify`. The helper must reject an absent required entry, symlink/non-regular entry, or any extra directory entry through the existing exact-file guard.

Add one helper which copies only the seven SQL filenames, in the order in section 1, from `LINEAGE_DIR` into a newly created caller-owned destination. It must use seven explicit `cp "$LINEAGE_DIR/$name" "$destination/$name"` operations (a fixed `for` list of the constants is acceptable). It must never copy `atlas.sum`; source `atlas.sum` is checked only to prove the source directory has no unexpected entry and is never an input to Atlas hashing.

Add `run_hash()` with this exact behavioral contract:

1. Call `require_exact_atlas_cli` and the seven-SQL lineage-input helper; do not call `require_internal_database_urls`, `start_materializer_job`, `run_materializer`, `schema inspect`, `migrate apply`, `migrate validate`, `migrate lint`, or `migrate diff`.
2. Create a fresh `/work/mail-calendar-schema/lineage-hash` directory. A pre-existing directory is a configuration failure; do not merge into it.
3. Copy only the seven allowlisted SQL files into that directory, invoke the pinned `"$ATLAS_CLI" migrate hash --dir "file://$candidate_dir"`, and fail with `EX_CONFIG` on any copy or Atlas failure. Suppress successful Atlas chatter so stdout consists solely of the reviewed candidate below; preserve failure diagnostics on stderr.
4. Require exactly eight regular candidate entries: the same seven SQL filenames and newly generated `atlas.sum`. A copied stale checksum, a missing/extra file, a symlink, or a nonregular entry is a hard failure.
5. Call `emit_review_file "$candidate_dir/atlas.sum"` exactly once and emit no SQL, database status, schema, or generated-source output. Its established SHA-256 and byte bound delimit the candidate for exact-byte review.

Update `run_verify()` to use the seven-SQL lineage-input helper, use `--latest "$EXPECTED_LINEAGE_SQL_FILE_COUNT"`, and interpolate `EXPECTED_LATEST_VERSION` plus `EXPECTED_LINEAGE_SQL_FILE_COUNT` in the verification status regex checks. Update the independent snapshot status check to require `EXPECTED_LATEST_VERSION` while retaining `Executed Files: 1`. Do not relax `Migration Status`, `Already at latest version`, zero-pending, bounded review-file, semantic diff, or current-schema roundtrip checks.

Extend the single-argument usage string and `case` dispatch to include `hash` exactly: `mail-calendar-schema {baseline|diff|generate|hash|verify}`. Unknown or extra arguments still exit `64`.

### 4.2 `docker-compose.local.yml`

Add only this service adjacent to the existing schema-maintenance operations:

```yaml
  schema-maintenance-hash:
    <<: *schema-maintenance-common
    command: ["hash"]
    network_mode: none
```

It must inherit the common pinned `linux/amd64` schema-maintenance image, UID/GID `10001:10001`, `read_only: true`, `no-new-privileges`, `cap_drop: [ALL]`, `init`, PID limit `128`, CPU limit `1.0`, memory limit `1g`, and the existing noexec `/work` (512 MiB) and `/tmp` (64 MiB) tmpfs mounts. Do not add `depends_on`, `networks`, database environment variables, ports, volumes, Docker socket, named volume, executable `/run/go-run-exec`, or an increased resource limit. `network_mode: none` is mandatory even though the operation does not declare network consumers.

### 4.3 `backend/migrations/README.md`

Correct the bounded lineage listing to all seven exact SQL filenames and `atlas.sum`. Add a concise "checksum reconciliation" subsection that says:

- `schema-maintenance-hash` is offline, no-database, pinned-Atlas candidate generation; it copies the reviewed seven SQL files, never source `atlas.sum`, and emits only a bounded candidate checksum for review.
- A candidate checksum is not migration authority until independent exact-byte review and Packet B apply; a successful hash alone is not a schema verification or database-apply claim.
- After Packet B, `schema-maintenance-verify` is expected to lint seven versions and converge at `20260827000002` / seven executed / zero pending; until that Docker run, this is an expectation, not an evidence claim.

Retain dated 2026-08-26 three-version Docker evidence as historical evidence, explicitly bounded to the old three-file lineage. Do not rewrite it as seven-file evidence.

## 5. Packet A static acceptance

Before requesting implementation review, inspect only the three Packet A files and require all of the following:

- The only new runtime operation is `hash`; it has no database or materializer call path and no network capability.
- Source admission and candidate output checks name all seven SQL files literally through constants and reject any eighth SQL/non-regular/extra input.
- The source stale `atlas.sum` cannot be copied into `/work`; only Atlas may create the candidate `atlas.sum` there.
- `run_verify` requires eight lineage entries, seven linted versions, version `20260827000002`, seven executed files, zero pending files, and a one-file latest-version snapshot.
- Existing baseline/diff/generate behavior, Atlas pin, exact URL checks, destructive-operation prohibitions, review output size bound, and Compose service resources are unchanged except for the explicit additions above.
- `docker compose -f docker-compose.local.yml config` is reviewed for the new service's `network_mode: none`, inherited `read_only`, tmpfs, 1 CPU/1 GiB/PID limits, and absence of dependencies, networks, ports, volumes, and database environment. This is configuration inspection, not a host test substitute.

Independent `gpt-5.6-terra` / `high` review must record C/I/M findings and reject any silent model/effort fallback. Packet A does not itself run Docker or alter `atlas.sum`.

## 6. Packet B candidate review and application

Only after Packet A has independent acceptance, run exactly:

```text
docker compose -f docker-compose.local.yml run --rm schema-maintenance-hash
```

This is the only allowed candidate-generation command. Do not run host Atlas, copy a source checksum into `/work`, or infer the checksum from individual files. Preserve the complete bounded stdout record containing `BEGIN REVIEW FILE atlas.sum`, its SHA-256 and byte count, candidate bytes, and the matching end marker.

Two independent exact-byte reviewers must compare the candidate enclosed by those markers against the proposed replacement at `backend/migrations/atlas/mail-calendar-v1/atlas.sum`. They must confirm all of these before Packet B is written:

1. The candidate was produced by the accepted `schema-maintenance-hash` service and contains exactly the one `atlas.sum` review file.
2. Its delimiter SHA-256 and byte count match the captured content.
3. The proposed repository file is byte-for-byte identical to the candidate, including final newline; no SQL filename or SQL byte changed.
4. The repository directory contains exactly the seven allowlisted SQL files and this one checksum, with only regular files.

If either reviewer rejects, differs, cannot reproduce provenance, or observes any extra output/input, mark `BLOCKED`; do not edit migrations to force a checksum, do not apply a guessed checksum, and do not invoke verify.

After both reviews pass, Packet B changes only `atlas.sum` and records the candidate stdout SHA-256/byte count in its review evidence. It has no production or shared-database side effect.

## 7. Docker verification, ordering, and acceptance

The order is mandatory:

1. Land/review Packet A.
2. Generate one candidate with `schema-maintenance-hash` and complete Packet B exact-byte review.
3. Apply only the reviewed candidate as Packet B.
4. Run the canonical disposable verification command:

```text
docker compose -f docker-compose.local.yml run --rm schema-maintenance-verify
```

5. Accept only if Atlas validation/lint succeeds, lint covers seven versions, apply/status reports latest `20260827000002`, seven executed and zero pending, the one-file snapshot reports latest `20260827000002` and zero pending, and both semantic schema diffs are empty. Record the exact command, image digest/pin, exit status, and bounded status output in Packet C.

The verification service uses its existing internal disposable PostgreSQL peer. Passing it proves only replayability and semantic convergence of this seven-file lineage; it does not prove upgrade safety for retained production data, application behavior, Ent generation, FTS, product readiness, merge, deploy, or production apply.

## 8. Fail-closed and rollback rules

| Condition | Required disposition |
|---|---|
| Missing, extra, symlink, or nonregular source/candidate entry | Exit `78`; emit no candidate authority; investigate the directory without changing SQL. |
| Atlas pin/hash failure or candidate output beyond 32 MiB | Exit `78`; preserve diagnostics; no Packet B. |
| Hash service has network, DB, dependency, volume, port, or resource expansion | Reject Packet A; remove the capability change before retrying review. |
| Candidate differs from either exact-byte review | `BLOCKED`; do not hand-edit/reorder/recreate SQL or checksum. |
| Packet B applied but disposable `schema-maintenance-verify` fails | Do not claim reconciliation. Revert only the Packet B `atlas.sum` commit if the branch must return to the pre-review state; retain the failure evidence and keep all seven SQL files unchanged. |
| Any shared/production target is proposed | Stop: requires a separate approved runbook, target identity, verified backup, current-state preflight, independent database review, and direct founder approval. |

There is no automatic database rollback. Packet A and Packet B operate only on source metadata and the disposable verification database; any eventual live migration remains forward-only under the existing execution boundary.

## 9. Contract completion criteria

This contract is complete when an implementer can execute Packets A/B/C without making an architectural decision: exact files, inputs, output, failure exits, resource limits, review gates, final status values, and rollback boundary are fixed above. It does not declare the reconciliation executed. At contract authoring time, all Docker/Atlas/database evidence for this new seven-file reconciliation remains `NOT_RUN`.
