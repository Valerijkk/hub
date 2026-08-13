[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только versioned mailbox
projection application seam и admission `mail.>` в shared EVENTS stream.

- User outcome: mailbox/calendar mutations становятся durable и idempotently consumable,
  но приватные payloads не начинают автоматически уходить во внешние webhooks.
- Completion: external-package/unit oracles precede minimal consumer logic; shared stream
  subjects add the mail wildcard after accepted calendar admission, webhook filter stays
  unchanged, and malformed/mismatched envelopes mutate no projection.
- Non-goals: service business logic, HTTP/UI, NATS server runtime, webhook feature,
  Stalwart, schema/migration, search implementation, Resend alerts.

[OWNERSHIP]

- Entry: accepted 143-04L, 143-04G, 144-03D and event-producing 144-04C/E on a stable
  `docs/mail/events-v1.schema.json`.
- Allowed mutations only:
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
- Consume only the three complete envelopes produced by accepted 144-03D. Require broker
  subject to equal JSON `subject`, schema version 1, non-zero application `event_id` and
  all required bounded payload fields before mutation.
- Projection application deduplicates on envelope application `event_id`. The distinct
  `Hub-Event-Id` transport header is diagnostic and cannot bypass or replace that ledger.
  Same domain ids with a new application event id still obey domain version/state rules.
- Subject mismatch, malformed envelope, unknown major version or invalid payload returns
  the existing dead-letter/error classification before projection mutation; never parse
  heuristically.
- `mail.message.changed.v1.changed_fields` accepts only `folder_ids`, `seen`, `answered`,
  `flagged`, `junk`, `labels`, `deleted`; flags accept only `seen`, `answered`, `flagged`,
  `junk`, `draft`. `deleted` removes the engine-authoritative projection state without
  deleting Hub-owned labels or independently verified calendar/CRM edges.
- Payload contains normalized bounded addresses/ids/state only; no body, raw MIME,
  attachment bytes, upload URL, credential, raw engine response or idempotency key.
- Domain services already persist full envelopes transactionally. This packet neither maps
  producers nor adds a post-commit direct publish shortcut or changes global outbox code.

[REQUIRED ORACLES]

1. Exact StreamSubjects set equals its accepted Phase 143 set plus `mail.>`; `calendar.>`
   remains exactly once.
2. Exact WebhookFilterSubjects snapshot equals its pre-packet value.
3. Each accepted full envelope applies its projection; broker/JSON subject mismatch and
   inner-payload-only JSON dead-letter with zero mutation.
4. Synthetic body/MIME/secret/key/provider error never appears in projection/error output.
5. Duplicate application event id applies projection once even with a different transport
   header; same domain ids with a new application event id follow
   domain version rules without double counters.
6. Unknown schema version is rejected/dead-letter classified before projection mutation.
7. Missing org/mailbox/message/thread/submission fields returns safe validation error and
   zero ledger/projection mutation.
8. Mail sent event is emitted only from confirmed `sent`, never `unknown`/`submitting`.
9. Deleted change removes owner projection/search/counter membership idempotently while
   retaining Hub-owned labels and verified cross-domain links until their own retention
   policy acts; mixed deleted/live engine state is rejected.

[EXECUTION / PROVENANCE]

- TDD first; in-memory only; no broker/listener/network/DB/wall clock.
- Host tests/build/validators and direct Docker/Compose forbidden. Accepted launcher absent
  => `NOT_RUN` and exact static diff.
- No dependency/generated code/commit/push/PR. max_depth 0; fan_out 0; retries 2;
  wall clock 45m.
- requested_model `gpt-5.6-terra`; reasoning `xhigh`; fallback none; mismatch => `BLOCKED`.

[ОТВЕТ]

На русском: exact subject diff, immutable webhook evidence, event mapping matrix,
verification, model provenance, risks.
