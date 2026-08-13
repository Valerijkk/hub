# Task 143-02B — static acceptance evidence

> **Status:** STATIC_ACCEPTED · RUNTIME_NOT_RUN · not product-accepted
> **Initial implementation:** `0201303c703a0dccf1500ed7d7a89360124962b0`
> **Hardening commit:** `2e7f7fa427b3943922bcd38c641f505de9175bbb`
> **Accepted head:** `efa7814d347e533c916a0ce14e694ccc1be8f98a`
> **Reviewed semantic contract:** `b3fb6f483b1335c4ef9ec3efbb0ff49c6d94ac3d`

## Scope and provenance

- Authoring packet: `/root/mail_143_02b_terra` and its scoped remediation turns.
- Requested and reported actual authoring route: exact `gpt-5.6-terra` / `xhigh`; no
  fallback.
- Independent review packet: `/root/mail_143_02b_review` in a distinct context.
- Requested and reported actual review route: exact `gpt-5.6-terra` / `high`; no fallback.
- Owned product/test files:
  - `backend/internal/calendar/itip_parse.go`
  - `backend/internal/calendar/itip_limits.go`
  - narrow additions to `backend/internal/calendar/itip_types.go`
  - `backend/tests/calendar/itip_parse_test.go`

## Accepted behavior

- bounded RFC 5545/5546 `VCALENDAR`/single-`VEVENT` parsing with strict CRLF unfolding,
  logical-line/property/parameter/count/byte limits and a closed property grammar;
- `REQUEST`, `CANCEL` and `REPLY` cardinality, organizer, attendee, temporal, sequence,
  recurrence and RRULE validation;
- UTF-8 validation after logical unfolding so a folded multi-byte rune is accepted without
  admitting malformed logical text;
- safe email validation that returns only stable sentinels and never includes attacker
  controlled addresses in errors;
- canonical CVE-v1 identity using the reviewed domain prefix, version, length-delimited
  frames and deterministic ordering;
- direct and parsed identity paths share the same timezone rule: explicit `UTC` and real
  IANA names are accepted, while Go's host-dependent pseudo-zone `Local` is rejected;
- no public sequence-comparison API beyond the frozen parser and canonical identity surface.

## Review lifecycle

The initial independent review found four issues: UTF-8 was checked before unfolding; two
email failure paths leaked input through a wrapped error; two extra public sequence symbols
expanded the frozen surface; and parameter maps could allocate before the property-aware
allowlist/cap. Commit `2e7f7fa4` closed all four and added adversarial external-package
oracles.

A full repeat review then found that `time.LoadLocation("Local")` is accepted by Go but
depends on the process host timezone. That made one wire payload capable of producing
different UTC instants and CVE hashes across hosts. Commit `efa7814d` introduced the shared
canonical-location validator and parser/direct-identity fixtures. The final independent
review returned PASS with no P0, P1 or P2.

## Exact accepted source blobs

| Path | Blob |
|---|---|
| `backend/internal/calendar/itip_parse.go` | `23a9c8e2ff651ab3f1a8d6e3e64848926ac59772` |
| `backend/internal/calendar/itip_limits.go` | `a67807e0ff2b9cb84af3153269d4a2445eddc13b` |
| `backend/internal/calendar/itip_types.go` | `ea6b0c123275c9ffb82d5706b088138755c7c747` |
| `backend/tests/calendar/itip_parse_test.go` | `63c1337b780265145372af4ba75a265226386db5` |

## Verification verdict

- exact allowed-path review: PASS;
- staged diff whitespace check before each commit: PASS;
- protected-path inventory: zero changes;
- hostile bounds, folded UTF-8, safe-error, timezone and CVE-v1 fixtures: present and
  independently reviewed;
- executable tests, build, lint, validators, generation and Docker jobs: `NOT_RUN`;
- accepted Docker launcher: `NOT_READY`.

Docker Desktop being available does not authorize direct Docker/Compose execution. This
packet is statically accepted only; it is not runtime-verified, complete, stable or
production-ready.
