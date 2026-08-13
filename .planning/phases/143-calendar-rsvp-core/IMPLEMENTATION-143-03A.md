# Task 143-03A — hash-pair remediation static evidence

> **Status:** STATIC_ACCEPTED · INDEPENDENT_REVIEW_C/I/M_0/0/0 · RUNTIME_NOT_RUN
> **Implementation commit:** `31bfff8f58ba83d12b59600135503f98a6c92fbd`
> **Accepted dependency:** 143-01 `e7d5f450d112158369de60738f95e0f5e29ace53`

## Scope and result

- Remediation packet: `/root/calendar_hash_pair_a`.
- Implementation route: exact `gpt-5.6-terra` / `xhigh`; `provenance_source=dispatcher-enforced`;
  actual telemetry `UNKNOWN`; no fallback.
- Independent reviewer: `/root/calendar_hash_pair_a_review`, exact `gpt-5.6-terra` / `high`;
  `provenance_source=dispatcher-enforced`; actual telemetry `UNKNOWN`.
- Packet A paths: `backend/ent/schema/calendarevent.go` and external oracle
  `backend/tests/calendar/calendar_event_schema_test.go`.
- The schema represents the reviewed timed/all-day union, status/sequence/version,
  recurrence/reminder values, immutable optional ScheduledCall link and event ownership.
- Fields, enum values, lengths, nullable pairs, `VersionMixin`, indexes and edge delete
  behavior were checked against Mail 0.1.7 and the Phase 143 packet.

The root audit finding for Packet A is remediated by the exact schema-level
`calendar_events_payload_hash_pair` check with the portable nullable-pair predicate. The
external Ent-metadata oracle checks that annotation without requiring generated output. An
independent static review found Critical/Important/Minor `0/0/0`; the Packet A diff is therefore
`STATIC_CONFORMANT`.

Exact reviewed content SHA-256: schema
`13682c6b1601961a12876431abf346485732b6b2b7aa5eb9bbcb833b565f0ec2`; external oracle
`da318bfb1fab24129c340f2b656b3cd93cf1c195bdf1970929427539071987ed`.

This is only a Packet A static result. It does not accept Phase 143, the Calendar product, a
migration, generated Ent output, or runtime behavior. Packet C is separately `STATIC_ACCEPTED`
on `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`; 143-03B is `STATIC_ACCEPTED` on clean
`dd346e31`. Packet 143-03F is `STATIC_ACCEPTED` on
`846cd00ecb243402877e18239c89a62bf07e8d14` after `SPEC PASS`/`QUALITY PASS` and final
docs/whole-boundary review C/I/M `0/0/0`; the intermediate C/I/M `0/0/1` docs-sync finding is
superseded/closed. Packet 143-03D is technically unblocked by accepted 03A–C, with no 03D work
claimed; 143-03E remains blocked on accepted 03D plus the launcher, with 03F satisfied.

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the
accepted launcher is `NOT_READY`. No generated or migration file was changed. The code,
external-test paths and this evidence form the accepted Task 1 boundary at
`31bfff8f58ba83d12b59600135503f98a6c92fbd`.
