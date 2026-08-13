# Phase 143 packet 04L — shared deterministic event envelope

> **Runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Status:** CODE_RUNTIME_BLOCKED
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
- Subject is bounded, contains no wildcard/whitespace/control character and ends in `.v1`.
  UUIDs are non-zero, occurred-at is UTC, correlation id is trimmed non-empty and at most
  128 bytes, and payload cannot be nil. Validation errors are stable and contain no payload.
- The constructor does not know `OutboxEvent`. Domain repositories later call the existing
  `PublishInTx(ctx, tx, envelope.Subject, envelope)`, causing the full envelope to be stored
  and published unchanged. The outbox row UUID/`Hub-Event-Id` remains a separate transport
  id; this package never creates or aliases it.

## Required oracles

1. Fixed inputs serialize byte-deterministically to the required complete envelope.
2. Every invalid subject/version-equivalent input, zero UUID, non-UTC time, empty/oversize
   correlation and nil payload is rejected before serialization.
3. Actor null versus UUID is exact; payload remains nested and no undeclared wrapper field
   appears.
4. Synthetic body/MIME/credential/idempotency-key values supplied only inside a rejecting
   fixture never appear in error text.
5. Reusing the same value preserves identical application event id and bytes; the unit
   does not pretend to prove database/outbox behavior.

## Execution and provenance

Write the failing genuine unit first, then the minimum implementation. No real database,
broker, network, wall clock, random source or listener. Host tests/build/lint/validators and
direct Docker/Compose are forbidden; absent accepted launcher means `NOT_RUN` with static
evidence only. No commit/push/PR. Report in Russian: exact symbols/files, oracle matrix,
verification status, requested/actual model, source SHA and residual risks.
