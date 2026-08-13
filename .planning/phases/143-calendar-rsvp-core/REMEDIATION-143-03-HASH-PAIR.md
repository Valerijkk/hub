# Phase 143 schema remediation — canonical hash pair checks

> **Detected:** 2026-08-13 root deep static audit
> **Severity:** P1 persistence invariant
> **Status:** PACKET_A_STATIC_CONFORMANT · PACKET_C_PENDING · RUNTIME_NOT_RUN

## Finding

Mail 0.1.7 requires every stored canonical payload hash to be paired with its version.
`PLAN.md` Task 143-03 additionally requires schema-level annotations/checks supported by
current Ent patterns. Packet A now remediates the `CalendarEvent` pair statically; Packet C
for `CalendarEventException` remains pending:

- `backend/ent/schema/calendarevent.go`
- `backend/ent/schema/calendareventexception.go`

Before the CHECK, per-field minimum/length validators rejected an invalid supplied value but
did not stop one column being `NULL` while the other was present. That half-pair would make
canonical dedupe state ambiguous and violates the persistence contract.

## Frozen minimal remediation

The exact tests-first sequential execution packet is
`docs/superpowers/plans/2026-08-13-calendar-hash-pair-remediation.md`. It adds one external
Ent-metadata unit oracle beside each sole product schema; this does not authorize generated
output, a migration or runtime execution.

Apply two sequential atomic packets so the original sole-file ownership remains intact. Packet A
has the exact CHECK and external metadata oracle and is independently `STATIC_CONFORMANT`
(Critical/Important/Minor `0/0/0`). Packet C remains pending.
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

Packet A was dispatched on exact `gpt-5.6-terra` / `xhigh` with
`provenance_source=dispatcher-enforced` and actual telemetry `UNKNOWN`; no fallback. Its
independent reviewer `/root/calendar_hash_pair_a_review` used exact `gpt-5.6-terra` / `high`
with the same accepted dispatcher provenance and actual telemetry `UNKNOWN`. The source diff and
external Ent-metadata oracle are independently `STATIC_CONFORMANT`; code review
Critical/Important/Minor is `0/0/0`. The two Packet A paths and this evidence form the
indivisible pending Task 1 commit boundary. Before the root commit they remain pending; afterward
the containing commit is authoritative. Packet C is still pending.

Generation, schema compilation, unit execution and runtime remain `NOT_RUN` because the accepted
launcher is `NOT_READY`. This static result is not a Phase 143, product, migration or runtime
PASS.
