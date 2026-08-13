# Phase 143 packet 04L — implementation evidence

> **Status:** `STATIC_ACCEPTED`
> **Authoritative boundary:** `4525c559da22729aa8bc51ff5336c29dad4f27e9`
> **Scope:** shared deterministic v1 event envelope only

## Acceptance record

The implementation route was `gpt-5.6-terra` / `xhigh`; independent reviews were
`gpt-5.6-terra` / `high`. The requested native routes were accepted by the dispatcher, so
`provenance_source` is `dispatcher-enforced`. Actual model/effort telemetry is `UNKNOWN`.

The initial code review recorded C/I/M `0/2/2`. The immutable-envelope/payload contract
amendment was then independently accepted (including the earlier 255-byte subject
sub-amendment). The remediated authoritative boundary received final independent PASS, C/I/M
`0/0/0`; therefore 04L is `STATIC_ACCEPTED`.

## Exact accepted paths and canonical blobs

| Path | Git blob object | SHA-256 of canonical Git blob bytes |
|---|---|---|
| `backend/internal/platform/events/envelope.go` | `add892cde5b8fab0fb0c19d08a576d2c915c39b4` | `8891E2BEDC0DA40E4C77630374F0B3BDCEEA5FD1D9F281F49049706BC4D55D5A` |
| `backend/tests/calendar/event_envelope_test.go` | `d58055044769d1961005285c31a910cfcc66feae` | `4D11A3EDCD5041650D738698DA250DF3ADA769B6FB3D7106161B324EF3540750` |

These SHA-256 values are independently recorded from streamed Git blobs, not working-tree
normalization; Git blob bytes are canonical.

## Frozen API and static oracles

- `EnvelopeV1` has no exported mutable fields or references. Construction snapshots only a
  serializable, non-null JSON object payload, stores frozen complete-envelope bytes, and
  `MarshalJSON` returns a fresh byte copy. `Subject()` and `EventID()` return values only.
- The JSON shape and order are `subject`, `schema_version` (`1`), `event_id`, `occurred_at`,
  `organization_id`, `actor_id`, `correlation_id`, `payload`. The unchanged producer seam is
  `PublishInTx(ctx, tx, envelope.Subject(), envelope)`; legacy outbox payloads are not wrapped
  or rewritten.
- Subjects are valid UTF-8, 1..255 bytes, dot-tokenized, wildcard/space/control-free and end
  in exact `.v1`. Event and organization UUIDs must be non-nil; actor is nil or non-nil UUID;
  correlation is trimmed, non-empty and at most 128 bytes.
- `occurred_at` must be RFC3339-capable with semantic zero offset. A zero-offset fixed zone is
  normalized/rendered as UTC `Z`; non-zero offsets reject. Invalid payload classes return the
  stable, content-free `ErrInvalidPayload`.
- Static RED tests cover deterministic complete bytes, exact subject and object-only payload
  validation, no sensitive rejected-input leakage, complete-envelope immutability after source
  and returned-byte mutation, UUID/actor/correlation checks, and UTC normalization. The package
  owns no I/O, clock, random source or generated UUIDs.

## Execution boundary

The accepted versioned launcher is `NOT_READY`. No Docker bypass was used; compilation, tests
and runtime are `NOT_RUN`. This record does not accept Phase 143, any calendar product behavior,
the global Mail workstream, migration, or runtime behavior.
