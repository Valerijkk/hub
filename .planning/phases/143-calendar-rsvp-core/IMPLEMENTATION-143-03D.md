# Task 143-03D — Calendar mutation receipt static acceptance

> **Status:** STATIC_ACCEPTED · FINAL_INDEPENDENT_REVIEW_C/I/M_0/0/0 · RUNTIME_NOT_RUN
> **Implementation commit:** `10b4ceead119cb74a16b325b33f9121f60d7b1b5`

## Scope, route and review history

- Requested authoring route: `gpt-5.6-terra` / `xhigh`, no fallback.
- Independent reviews: `gpt-5.6-terra` / `high`, no fallback.
- Provenance source: `dispatcher-enforced`; actual model/effort telemetry is `UNKNOWN`.
- Review chain: initial C/I/M `0/2/1`; first-fix re-review `0/1/0`; final independent review `PASS`, C/I/M `0/0/0`.
- Owned files: `backend/ent/schema/calendarmutationreceipt.go`, `backend/ent/schema/user.go`, `backend/ent/schema/scheduledcall.go`, and `backend/tests/calendar/calendar_mutation_receipt_schema_test.go`.

The static test was written first; expected static RED history is retained. Executable RED/GREEN,
Ent generation, schema compilation, migrations, tests and build are `NOT_RUN`; launcher `NOT_READY`.

## Static schema result

The §3.14 receipt has actor/org/operation identity, immutable SHA-256 key and request hashes,
in-progress/completed state, expiry, optional event reference, and the unique tuple; raw key absent.
Both hashes have Go validators and named 32-byte database checks. `response_json` requires
`utf8.Valid` and `json.Valid` and is 1 MiB-bounded; `response_status` is 100..599 in Go and DB.
Actor/org/event deletion is `Cascade`/`NoAction`/`SetNull`. User and ScheduledCall edges are
additive; optional one-to-one call link immutable by service contract. Service owns atomic state/response pairing.

## Authoritative committed content hashes

SHA-256 was recalculated by streaming each Git blob from the implementation commit:

| Path | verified SHA-256 |
|---|---|
| `backend/ent/schema/calendarmutationreceipt.go` | `2A07CB61D1AF5CCFAAD4F8F7E3998EB071FA91F635CA10B9ABDB28ECCB5C1808` |
| `backend/ent/schema/user.go` | `9CBE1C120ACCF8C7533910DA42B5EE5224B6E3FCE3A174EADED34B535FE72AB3` |
| `backend/ent/schema/scheduledcall.go` | `C085E698E1079D7EF2ED91B57038EA34180C327C730657DA35C4728FD02CCA6E` |
| `backend/tests/calendar/calendar_mutation_receipt_schema_test.go` | `89DE0503A4DC40DC507B4AD7B2A871EBA9EF80100686A6CD6F2118EBF80DF903` |

The authoritative values are the streamed Git blob digests above. Earlier filesystem digests
for `user.go` and `scheduledcall.go` were normalization-sensitive (CRLF) and are not provenance
or acceptance evidence.

## Downstream state

03A/B/C/D/F are static accepted. 03E is no longer blocked by 03D or 03F; it remains
`BLOCKED/NOT_RUN` pending launcher and the separate founder-approved migration mechanism/gate.
The all-A product decision does not establish that technical mechanism, execute, or accept migration.
