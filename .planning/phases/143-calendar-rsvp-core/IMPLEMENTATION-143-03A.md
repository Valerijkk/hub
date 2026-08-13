# Task 143-03A — superseded static acceptance evidence

> **Status:** P1_REMEDIATION_REQUIRED · MODEL_PROVENANCE_BLOCKED · NOT_STATIC_ACCEPTED
> **Implementation commit:** `c6f5ec8b54b7e35327125dd4445634c5d1764491`
> **Accepted dependency:** 143-01 `e7d5f450d112158369de60738f95e0f5e29ace53`

## Scope and result

- Authoring packet: `/root/mail_143_03a_impl`.
- Requested and reported actual route: exact `gpt-5.6-terra` / `xhigh`; no fallback.
- Sole owned file: `backend/ent/schema/calendarevent.go`.
- The schema represents the reviewed timed/all-day union, status/sequence/version,
  recurrence/reminder values, immutable optional ScheduledCall link and event ownership.
- Fields, enum values, lengths, nullable pairs, `VersionMixin`, indexes and edge delete
  behavior were checked against Mail 0.1.7 and the Phase 143 packet.

The initial independent Terra/high static review returned PASS with no P0, P1 or P2. A later
root deep audit superseded that acceptance: the nullable canonical hash version/hash fields
lack the schema-level pair check required by Mail 0.1.7 and `PLAN.md` Task 143-03. Per-field
validators still permit exactly one column to be null, so this is a P1 persistence invariant.
The frozen remediation is recorded in `REMEDIATION-143-03-HASH-PAIR.md`. Its first exact
Terra/xhigh attempt stopped before reads/writes because actual model telemetry was
unavailable.

The unchanged source blob is `8cece971a3c1c3b1f76bef72d01e247ae5b5230d` (file SHA-256
`61D63102FE819C3CE20562065A8A56F2635909824A75677536B4D564D4D8A6EF`).

Ent generation, schema compilation, migrations and executable tests remain `NOT_RUN`; the
accepted launcher is `NOT_READY`. No generated or migration file was changed. This packet
must not unlock dependents until the P1 is fixed and independently reviewed.
