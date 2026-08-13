# Phase 143 packet 04L — shared deterministic event envelope

> **Runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Status:** `STATIC_ACCEPTED` at `4525c559da22729aa8bc51ff5336c29dad4f27e9`.
> **Independent review:** final Terra / high PASS · C/I/M `0/0/0` · dispatcher-enforced
> provenance · actual telemetry `UNKNOWN`. The earlier docs-only 255-byte subject sub-amendment
> and immutable-envelope/payload amendment are accepted historical inputs; the initial code
> review C/I/M `0/2/2` is remediated and historical. Global Mail review remains pending.
> **Scope:** one pure shared value/builder plus one genuine external-package unit

## Outcome

Implement the additive v1 envelope used by every new Mail, Calendar and linked-call
producer without changing the global outbox publisher, worker or any legacy event payload.

## Ownership

Allowed mutations only:

- `backend/internal/platform/events/envelope.go`
- `backend/tests/calendar/event_envelope_test.go`

Do not edit `backend/internal/platform/outbox/**`, NATS code, Ent/generated code, domain
services, JSON schema, protected test roots or dependencies. Stop on dirty overlap.

## Frozen contract

- Export one immutable `EnvelopeV1` JSON value and one deterministic constructor/validator.
  `EnvelopeV1` has no exported mutable fields or references. It snapshots a valid payload
  JSON object during construction, then stores one frozen complete-envelope representation.
  `MarshalJSON` returns a fresh copy of those complete bytes, so neither later mutation of
  an input map/slice/struct pointer nor mutation of bytes returned to a caller can alter the
  envelope. Public read access is copy-returning `Subject() string` and `EventID() uuid.UUID`.
- Inputs are explicit: exact `.v1` subject, application event UUID, UTC RFC3339-capable
  occurrence time, organization UUID, nullable actor UUID, correlation id and non-null
  typed payload. The package owns no clock, UUID generator, global state or I/O.
- JSON field order/shape follows `docs/mail/events-v1.schema.json`: `subject`,
  `schema_version`, `event_id`, `occurred_at`, `organization_id`, `actor_id`,
  `correlation_id`, `payload`. Schema version is exactly `1`.
- Subject validity for every new `EnvelopeV1` is: valid UTF-8; 1..255 bytes; dot-separated
  nonempty tokens; no leading/trailing dot or `..`; no `*`/`>`; no Unicode whitespace or
  control characters; suffix `.v1`. The 255-byte bound is a Hub product/safety contract,
  not a NATS transport/control-line ceiling. It applies only to the new `EnvelopeV1` builder
  and does not apply globally to legacy `OutboxEvent` or the publisher.
  UUIDs are non-zero; an RFC3339-capable occurrence time is accepted exactly when its actual
  zone offset is zero, then normalized/stored/rendered as `time.UTC`/`Z`; correlation id is
  trimmed non-empty and at most 128 bytes. Payload accepts only a successfully serializable,
  non-null JSON object, including a custom `json.Marshaler` whose result is such an object.
  Scalars, arrays/slices, funcs/chans/unsupported values, all typed nils, `null`, marshal
  errors and custom array/scalar/null marshal output reject with the stable, content-free
  `ErrInvalidPayload`. Snapshot the payload object bytes once before assembling the envelope.
- The constructor does not know `OutboxEvent`. Domain repositories later call the existing
  `PublishInTx(ctx, tx, envelope.Subject(), envelope)`, causing the full envelope to be stored
  and published unchanged. The outbox row UUID/`Hub-Event-Id` remains a separate transport
  id; this package never creates or aliases it.

## Required oracles

1. Fixed inputs serialize byte-deterministically to the required complete envelope; a valid
   struct, map object and custom `json.Marshaler` that returns a non-null JSON object are
   accepted. The custom-Marshaler object bytes are frozen: later source mutation and mutation
   of a slice returned by `MarshalJSON` cannot alter later envelope bytes.
2. A valid exact 255-ASCII-byte subject is accepted and 256 bytes are rejected. A multibyte
   subject with at most 255 code points but more than 255 UTF-8 bytes, invalid UTF-8, an empty
   token, leading/trailing dot, wildcard, Unicode whitespace or control character is rejected;
   every such error excludes payload text.
3. Every invalid version-equivalent input, zero UUID, nonzero-offset time, empty/oversize
   correlation and invalid payload is rejected before envelope serialization. A
   `time.FixedZone("UTC", 0)` value is accepted, normalizes to `Z` and produces schema-valid
   output; a nonzero offset rejects.
4. Actor null versus UUID is exact; payload remains nested and no undeclared wrapper field
   appears.
5. Synthetic body/MIME/credential/idempotency-key values supplied only inside a rejecting
   fixture never appear in error text.
6. Mutating a source map, nested slice or struct pointer after construction leaves bytes,
   subject and event id unchanged; mutating a byte slice returned by `MarshalJSON` leaves a
   later marshal unchanged.
7. Scalar, non-nil slice, func, chan, typed nil map/slice/pointer/function/channel/interface,
   marshal-error and custom array/scalar/null non-object payloads reject without panic and
   with no payload content in the error. Reusing the same value preserves identical
   application event id and bytes; the unit does not pretend to prove database/outbox behavior.

## Execution and provenance

Write the failing genuine unit first, then the minimum implementation. No real database,
broker, network, wall clock, random source or listener. Host tests/build/lint/validators and
direct Docker/Compose are forbidden; absent accepted launcher means `NOT_RUN` with static
evidence only. The remediated code/test boundary is `STATIC_ACCEPTED` at
`4525c559da22729aa8bc51ff5336c29dad4f27e9`; the historical initial C/I/M `0/2/2` does not
describe the current packet. This does not accept the Phase, product, global Mail review or
runtime: the launcher remains `NOT_READY`, and compilation/tests/runtime remain `NOT_RUN`.
No commit/push/PR. Report in Russian: exact
symbols/files, oracle matrix, verification status, requested/actual model, source SHA and
residual risks.
