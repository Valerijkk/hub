# Task 143-02A — static acceptance evidence

> **Status:** STATIC_ACCEPTED · RUNTIME_NOT_RUN · not product-accepted
> **Dispatch head:** `2ddabe905de5fc2e7c3e36d8f600c9c1e4eeefcb`
> **Implementation commit:** `dbf39ef9839e112f41b96d28aca18c42f1c4e523`
> **Fix commit / accepted head:** `d73a9380a567ffd17acd7ea04b0e9fb1fded3e70`
> **Reviewed semantic contract:** `5a351d5bd51bd215775f3828bc26979cf985a4ac`
> **Accepted dependency:** 143-01 `e7d5f450d112158369de60738f95e0f5e29ace53`

## Scope and provenance

- Initial authoring packet: `/root/mail_143_02a_terra`.
- Fix packet: `/root/mail_143_02a_fix1`.
- Requested and actual authoring route for both: exact `gpt-5.6-terra` / `xhigh`; no
  fallback.
- Independent task review: `/root/mail_143_02a_review`.
- Scoped fix re-review: `/root/mail_143_02a_rereview1`.
- Requested and actual review route for both: `gpt-5.6-terra` / `high`; no fallback.
- Owned product/test files:
  - `backend/internal/calendar/itip_types.go`
  - `backend/internal/calendar/itip_serialize.go`
  - `backend/tests/calendar/itip_serialize_test.go`

## Accepted behavior

- pure `SerializeITIP(message ITIPMessage) ([]byte, error)` for RFC 5545/5546
  `REQUEST`, `CANCEL` and `REPLY`;
- explicit caller-supplied non-zero UTC `DTSTAMP`, canonical UID/sequence/status,
  organizer and participant values;
- strict timed/all-day union, IANA/DST conversion, exclusive all-day end date and optional
  occurrence `RECURRENCE-ID`;
- deterministic RRULE/EXDATE and attendee ordering;
- RFC text/parameter escaping, CRLF-only output and 75-octet UTF-8-safe folding with one-
  space continuation;
- stable validation sentinels for method/status/participant/temporal/recurrence/stamp/
  sequence/email failures;
- rejection of disallowed control characters before TEXT or CN output;
- external-package literal oracles for exact REQUEST output, CANCEL, one-attendee REPLY,
  DST, all-day, recurrence, escaping, UTF-8 folding, errors and byte-identical repeats.

## Test-first evidence

The external test file was created while both production files and the public entry point
were absent. Production code was added only after that static RED condition. During fix
round 1, the control-character behavior table was likewise added before its validation
branch. The accepted launcher is not ready, so neither RED nor GREEN was executed; no
runtime TDD claim is made. The evidence is test-first edit order plus static review only.

## Review lifecycle

The first independent review returned three Important findings and one Minor:

1. it asked the serializer to prove that CANCEL sequence was incremented;
2. it questioned canonical lowercasing of normalized email identity;
3. it treated runtime IANA database resolution as forbidden filesystem/global I/O;
4. it found that non-NUL control characters could reach TEXT/CN output.

Fix round 1 added the control-character oracle and minimal guard. The other three were
resolved against frozen ownership rather than by inventing public fields: MAIL-SPEC §4.2
and accepted 143-01 own sequence increment before serialization; §§3.11–3.12 own normalized
email identity; OpenAPI explicitly requires the runtime IANA timezone database and accepted
143-01 owns the same validation boundary. A new previous-sequence field, email identity
policy or timezone resolver would have changed the reviewed contract.

Scoped independent re-review marked all four findings `ADDRESSED`, found no new Critical or
Important breakage and approved task quality. Runtime remained `NOT_RUN` throughout.

## Exact accepted source blobs

| Path | Blob |
|---|---|
| `backend/internal/calendar/itip_types.go` | `d7408dff2d5a6b796d67a8515049a895e7f15b8d` |
| `backend/internal/calendar/itip_serialize.go` | `138d0dbe5760632f8ba85c410fa6a8133a4f5d5a` |
| `backend/tests/calendar/itip_serialize_test.go` | `0e6fc624e4ed3b7dab759d5d30c1cf2d57089020` |

## Verification verdict

- exact task range contains only the three owned files: PASS;
- `git diff --cached --check` before implementation and fix commits: PASS;
- protected-path inventory: zero changes;
- secret and forbidden AI/coauthor-trailer scan: zero findings;
- forbidden explicit clock/application I/O/global mutable state scan: zero findings;
- original task review plus scoped fix re-review: PASS after one fix round;
- executable tests, build, lint, validators, generation and Docker jobs: `NOT_RUN`;
- accepted Docker launcher: `NOT_READY`.

Docker Desktop availability does not authorize direct Docker/Compose fallback. This packet
is `STATIC_ACCEPTED`, not runtime-verified, merge-ready, complete, stable or production-
ready. The next dependency-safe packet is 143-02B.
