# Phase 143 packet 04L — shared deterministic event envelope

> **Runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Status:** CONTRACT_AMENDMENT_STATIC_ACCEPTED · READY_FOR_CODE_REDISPATCH
> **Independent review:** Terra / high PASS · C/I/M `0/0/0` · dispatcher-enforced
> provenance · actual telemetry `UNKNOWN` · exact docs commit pending
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
  UUIDs are non-zero, occurred-at is UTC, correlation id is trimmed non-empty and at most
  128 bytes, and payload cannot be nil. Validation errors are stable and contain no payload.
- The constructor does not know `OutboxEvent`. Domain repositories later call the existing
  `PublishInTx(ctx, tx, envelope.Subject, envelope)`, causing the full envelope to be stored
  and published unchanged. The outbox row UUID/`Hub-Event-Id` remains a separate transport
  id; this package never creates or aliases it.

## Required oracles

1. Fixed inputs serialize byte-deterministically to the required complete envelope.
2. A valid exact 255-ASCII-byte subject is accepted and 256 bytes are rejected. A multibyte
   subject with at most 255 code points but more than 255 UTF-8 bytes, invalid UTF-8, an empty
   token, leading/trailing dot, wildcard, Unicode whitespace or control character is rejected;
   every such error excludes payload text.
3. Every invalid version-equivalent input, zero UUID, non-UTC time, empty/oversize
   correlation and nil payload is rejected before serialization.
4. Actor null versus UUID is exact; payload remains nested and no undeclared wrapper field
   appears.
5. Synthetic body/MIME/credential/idempotency-key values supplied only inside a rejecting
   fixture never appear in error text.
6. Reusing the same value preserves identical application event id and bytes; the unit
   does not pretend to prove database/outbox behavior.

## Execution and provenance

Write the failing genuine unit first, then the minimum implementation. No real database,
broker, network, wall clock, random source or listener. Host tests/build/lint/validators and
direct Docker/Compose are forbidden; absent accepted launcher means `NOT_RUN` with static
evidence only. Contract acceptance claims no 04L product code; the launcher remains
`NOT_READY` and runtime remains `NOT_RUN`. No commit/push/PR. Report in Russian: exact
symbols/files, oracle matrix, verification status, requested/actual model, source SHA and
residual risks.
