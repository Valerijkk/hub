# Task 143-03C — hash-pair remediation static evidence

> **Status:** STATIC_CONFORMANT · INDEPENDENT_REVIEW_C/I/M_0/0/0 · RUNTIME_NOT_RUN
> **Implementation commit:** the atomic Git commit that will contain this evidence document together
> with `backend/ent/schema/calendareventexception.go` and
> `backend/tests/calendar/calendar_event_exception_schema_test.go`; resolve it from Git history
> after root creates the pending commit
> **Accepted dependency:** Packet A `31bfff8f58ba83d12b59600135503f98a6c92fbd`

## Scope and result

- Remediation packet: `/root/calendar_hash_pair_c`.
- Implementation route: exact `gpt-5.6-terra` / `xhigh`; `provenance_source=dispatcher-enforced`;
  actual telemetry `UNKNOWN`; no fallback.
- Independent reviewer: `/root/calendar_hash_pair_c_review`, exact `gpt-5.6-terra` / `high`;
  `provenance_source=dispatcher-enforced`; actual telemetry `UNKNOWN`.
- Documentation sync route: exact `gpt-5.6-terra` / `medium`.
- Packet C paths: `backend/ent/schema/calendareventexception.go` and external oracle
  `backend/tests/calendar/calendar_event_exception_schema_test.go`.
- Packet A commit `31bfff8f58ba83d12b59600135503f98a6c92fbd` is unchanged.

The root audit finding for Packet C is remediated by the exact schema-level
`calendar_event_exceptions_payload_hash_pair` check with the portable nullable-pair predicate.
The external Ent-metadata oracle checks that annotation without requiring generated output. An
independent static review found Critical/Important/Minor `0/0/0`; the Packet C diff is therefore
`STATIC_CONFORMANT`.

Exact reviewed content SHA-256: schema
`a653e046e7c8247add965d22eb0c73d7f543b50f4ba6389a0c22af16ed53950d`; external oracle
`23554a0379c34c16853ec564fdd7bd76498b0505108919a5a5c29b8984ef0cd5`.

After Packet C, both hash-pair P1 remediations are statically conformant. This does not accept
Phase 143, the Calendar product, a migration, generated Ent output, or runtime behavior;
downstream acceptance remains separate.

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the
accepted launcher is `NOT_READY`. No generated or migration file was changed. The code and
external-test paths and this evidence form the indivisible pending Task 2 commit boundary. Before
the root commit they remain pending; afterward the containing commit is authoritative.
