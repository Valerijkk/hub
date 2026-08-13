[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только mailbox event mapping,
projection application seam и admission двух subject families в shared EVENTS stream.

- User outcome: mailbox/calendar mutations становятся durable и idempotently consumable,
  но приватные payloads не начинают автоматически уходить во внешние webhooks.
- Completion: external-package/unit oracles precede minimal mapping; shared stream subjects
  add the mail wildcard after accepted calendar admission, webhook filter stays unchanged,
  payloads match schema.
- Non-goals: service business logic, HTTP/UI, NATS server runtime, webhook feature,
  Stalwart, schema/migration, search implementation, Resend alerts.

[OWNERSHIP]

- Entry: accepted event-producing services 143/144 and stable
  `docs/mail/events-v1.schema.json`.
- Allowed mutations only:
  - `backend/internal/mail/mailbox_events.go`
  - `backend/internal/mail/mailbox_projection.go`
  - `backend/internal/platform/nats/client.go`
  - `backend/tests/mail/mailbox_events_test.go`
  - a focused genuine unit for the NATS subject lists only if repository layout requires it
- Do not modify `WebhookFilterSubjects`, webhook allowlists/handlers, event JSON schema or
  unrelated subject ordering. Stop on dirty overlap.

[FROZEN BEHAVIOR]

- Entry requires accepted Phase 143 packet 143-04G, which already added `calendar.>`.
- Add `mail.>` exactly once to `StreamSubjects`; preserve the existing `calendar.>` entry.
- Preserve `WebhookFilterSubjects` byte-for-byte. These wildcard families are private
  internal events, not opt-in webhook events.
- Map only subjects in `events-v1.schema.json`: `mail.received.v1`, `mail.sent.v1`,
  `mail.message.changed.v1`, calendar invite create/update/cancel and RSVP change.
- Envelope version/id/time/org/actor/correlation are explicit inputs; mapper never calls
  wall clock or creates random ids.
- Payload contains normalized bounded addresses/ids/state only; no body, raw MIME,
  attachment bytes, upload URL, credential, raw engine response or idempotency key.
- Projection application accepts event id/version and is duplicate-idempotent; unknown
  major version returns the existing dead-letter/error classification, not best-effort
  parsing.
- Domain services persist outbox event in their transaction. This packet does not add a
  post-commit direct publish shortcut.

[REQUIRED ORACLES]

1. Exact StreamSubjects set equals its accepted Phase 143 set plus `mail.>`; `calendar.>`
   remains exactly once.
2. Exact WebhookFilterSubjects snapshot equals its pre-packet value.
3. Every frozen event maps to schema-shaped JSON with deterministic envelope.
4. Synthetic body/MIME/secret/key/provider error never appears in serialized output.
5. Duplicate event id applies projection once; same domain ids with a new event id follow
   domain version rules without double counters.
6. Unknown schema version is rejected/dead-letter classified before projection mutation.
7. Missing org/mailbox/message/thread/submission fields returns safe validation error.
8. Mail sent event is emitted only from confirmed `sent`, never `unknown`/`submitting`.

[EXECUTION / PROVENANCE]

- TDD first; in-memory only; no broker/listener/network/DB/wall clock.
- Host tests/build/validators and direct Docker/Compose forbidden. Accepted launcher absent
  => `NOT_RUN` and exact static diff.
- No dependency/generated code/commit/push/PR. max_depth 0; fan_out 0; retries 2;
  wall clock 45m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none; mismatch => `BLOCKED`.

[ОТВЕТ]

На русском: exact subject diff, immutable webhook evidence, event mapping matrix,
verification, model provenance, risks.
