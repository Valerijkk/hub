# Phase 144 packet 03D — deterministic mail event mapping

> **Runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Status:** TERRA_CODE_ROUTE_AUTHORIZED · BLOCKED_BY_143_04L_AND_144_01A
> **Dependency:** accepted 143-04L and 144-01A

## Outcome

Map the three frozen Mail event payloads into the shared complete `EnvelopeV1` before any
transactional Mail producer is implemented. This packet is pure mapping only.

## Ownership

Allowed mutations only:

- `backend/internal/mail/mailbox_events.go`
- `backend/tests/mail/mailbox_events_test.go`

Do not edit services, repositories, outbox/NATS, schemas, generated code, consumers,
handlers or protected test roots. Stop on dirty overlap.

## Frozen behavior

- Define bounded typed inputs/payloads for `mail.received.v1`, `mail.sent.v1` and
  `mail.message.changed.v1` exactly as `docs/mail/events-v1.schema.json` and Mail spec
  section 6 require.
- Each mapper receives application event UUID, UTC occurred-at, organization UUID,
  nullable actor UUID, correlation id and domain payload explicitly, then calls the
  accepted shared 143-04L builder. It owns no clock, id generator, I/O or global state.
- `mail.received.v1` contains the owner-safe projection identifiers, normalized sender and
  recipients, received time, folder ids, attachment flag and nullable calendar method.
  `mail.sent.v1` contains confirmed submission/message/thread identifiers, normalized
  sender/recipients, sent time and delivery class. `mail.message.changed.v1` contains
  message id, a non-empty bounded changed-field set from exactly `folder_ids`, `seen`,
  `answered`, `flagged`, `junk`, `labels`, `deleted`; its current flags are drawn only from
  `seen`, `answered`, `flagged`, `junk`, `draft`. A `deleted` change has empty engine
  folder/flag state while preserving the current bounded Hub label ids supplied by the
  authoritative projection command.
- Reject missing/zero/cross-account identifiers, duplicate/unbounded addresses or ids,
  unknown field/change/delivery values and non-normalized inputs with stable safe errors.
- No event contains body, subject line, snippet, raw MIME/header, attachment bytes/name,
  upload URL, provider response/error, credential, token or raw idempotency key.

## Required oracles

1. Fixed inputs for every subject serialize to exact deterministic complete envelopes and
   conform field-for-field to the checked-in JSON schema.
2. Mapper never silently repairs normalized addresses, deduplicates ambiguous input or
   reads current time/randomness.
3. Every required field and bound has a negative fixture; errors contain no input payload.
4. Synthetic forbidden content cannot appear anywhere in serialized envelope bytes.
5. Reusing an application event id produces identical envelope bytes. A distinct event id
   changes only the expected field when all other inputs are fixed.
6. Unknown changed-field/flag values, `deleted` mixed with live engine state, duplicate
   values and an empty changed-field set are rejected before envelope construction.

## Execution and provenance

Write genuine external-package units first, then minimum mapping. No database, broker,
network, wall clock, random source or listener. Host tests/build/lint/validators and direct
Docker/Compose are forbidden; absent accepted launcher means `NOT_RUN` plus static evidence.
No commit/push/PR. Respond in Russian with exact files/symbols, mapping/oracle matrix,
verification, requested/actual model, exact source SHA and residual risks.
