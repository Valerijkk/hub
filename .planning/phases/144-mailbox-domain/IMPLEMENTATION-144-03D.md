# 144-03D implementation evidence receipt

## Git boundary

Source/oracle is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
`42c6b9f2bd4383040fc116630727564fcf17dd78` (parent
`d60dab6b39a4642f8c7ca29e377cce9a6ba6d704`, tree
`d711e0c0c9d56d082a3352c4f4bedafc8ce2c18d`; `feat(mail): add deterministic event mapping`).
The exact two-path boundary is `backend/internal/mail/mailbox_events.go`
(`792a4f460f756c4541003bf1ddc0391be637052c`) and
`backend/tests/mail/mailbox_events_test.go`
(`8947577115f60f4c3b026dcb686014444660f815`).

## Behavior and oracle matrix

| Behavior | Static oracle result |
|---|---|
| `mail.received.v1`, `mail.sent.v1`, `mail.message.changed.v1` map to complete `EnvelopeV1` | Exact schema-shaped deterministic envelopes |
| Bounded identifiers, normalized addresses, delivery/change values and deleted-state rules | Valid and negative fixtures; invalid input rejected before construction |
| No clock, randomness, I/O, body or secret leakage | Pure mapping and serialized forbidden-content checks |

## Provenance and limits

Authoring was `gpt-5.6-terra` / `xhigh`, dispatcher-enforced, no fallback; telemetry is
`UNKNOWN` unless exposed. Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh,
and ABI/Compile-shape/Lifecycle Terra/high reviews are all `EXACT_SHA_PASS`, C/I/M `0/0/0`.

This is static acceptance only. Runtime, tests, build, gofmt, lint, validators, Docker and
schema runtime remain `NOT_RUN`; launcher is `NOT_READY`. It grants no schema, migration,
service, UI, product, phase or release authority. Existing 02A-F, Rule, Sieve and Phase 145
blockers are unchanged. Next mainline permitted packet is `144-03C-Send` (send/rate/loop pure
policy); `144-03C-Attachment` follows separately. Calendar RSVP/service remains blocked on
143-03E launcher and migration authority.
