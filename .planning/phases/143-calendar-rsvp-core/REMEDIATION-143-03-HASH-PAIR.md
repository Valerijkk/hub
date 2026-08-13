# Phase 143 schema remediation — canonical hash pair checks

> **Detected:** 2026-08-13 root deep static audit
> **Severity:** P1 persistence invariant
> **Status:** CODE_FIX_BLOCKED_BY_MODEL_PROVENANCE · RUNTIME_NOT_RUN

## Finding

Mail 0.1.7 requires every stored canonical payload hash to be paired with its version.
`PLAN.md` Task 143-03 additionally requires schema-level annotations/checks supported by
current Ent patterns. Both of these schemas currently declare the fields nullable but do
not enforce the pair at the database-schema layer:

- `backend/ent/schema/calendarevent.go`
- `backend/ent/schema/calendareventexception.go`

Per-field minimum/length validators reject an invalid supplied value, but they do not stop
one column being `NULL` while the other is present. That half-pair would make canonical
dedupe state ambiguous and violates the persistence contract.

## Frozen minimal remediation

The exact tests-first sequential execution packet is
`docs/superpowers/plans/2026-08-13-calendar-hash-pair-remediation.md`. It adds one external
Ent-metadata unit oracle beside each sole product schema; this does not authorize generated
output, a migration or runtime execution.

Apply two sequential atomic packets so the original sole-file ownership remains intact.
For each schema:

1. add only the imports needed for `schema.Annotation` and `entsql.Checks`;
2. add an `Annotations()` method with a stable schema-specific check name;
3. use the portable predicate
   `(payload_hash_version IS NULL) = (payload_hash IS NULL)`;
4. change no field, edge, index, enum, comment, generated file or migration;
5. use the existing `backend/ent/schema/journalline.go` annotation style as the repository
   API oracle;
6. leave generation/compile/runtime `NOT_RUN` until the accepted launcher exists.

Suggested check names:

- `calendar_events_payload_hash_pair`
- `calendar_event_exceptions_payload_hash_pair`

The temporal timed/all-day union and end-after-start remain service-owned exactly as the
original dispatch states; this remediation does not add unrelated SQL checks.

## Dispatch result

An exact Terra/xhigh 143-03A fix attempt on clean head `f0b91f07` stopped before reading or
writing because actual model telemetry was unavailable. No fallback, diff, stage, commit,
test, generator or Docker command occurred. 143-03A is no longer statically accepted until
the check is authored and independently reviewed; 143-03C already remains unaccepted.
