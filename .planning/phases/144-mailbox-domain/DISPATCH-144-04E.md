[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только idempotent mailbox
submission state machine после принятого draft service.

- User outcome: повтор после таймаута никогда не отправляет второе письмо, а UI получает
  owner-scoped receipt для polling/retry.
- Completion: failing units precede minimal service; каждый terminal/unknown transition и
  outbox cardinality доказаны.
- Non-goals: HTTP handler, Stalwart mapping, SMTP delivery policy, UI, draft save, schema,
  attachment upload, NATS transport, migration, legacy broadcast/Resend.

[OWNERSHIP]

- Entry: accepted 143-04L, 144-01B, 144-02C/E, 144-03D, 144-04A and 144-04D on
  clean reviewed HEAD.
- Allowed mutations only:
  - `backend/internal/mail/submission_service.go`
  - `backend/tests/mail/submission_service_test.go`
  - narrow submission repository/outbox signatures in accepted
    `backend/internal/mail/mailbox_repository.go` when strictly required.
- No other file. Stop on dirty overlap or missing port semantics.

[FROZEN STATE MACHINE]

```text
POST send(draft, raw key, canonical request)
  hash key; hash canonical request
  absent -> reserve queued with stable engine_submission_key
         -> validate active draft/identity/recipients/clean attachments
         -> atomically freeze draft as submitting
         -> submission submitting -> engine Submit once
         -> sent + result ids + draft submitted + full mail.sent.v1 envelope in same Hub transaction
  same key + different request hash -> 409, zero engine call
  same key + sent/failed -> replay same receipt
  same key + queued/submitting/unknown -> ReconcileSubmission first

Reconcile result
  sent -> finalize sent once + one outbox
  definitive absent -> transition queued and permit exactly one controlled Submit while
                        the draft remains frozen
  indeterminate -> remain unknown; draft remains frozen; never Submit
  definitive rejected -> failed safe code; restore draft active; never retry automatically
```

Raw idempotency key is request-scoped only. Persistence stores key hash, request hash and
stable engine key. The receipt response never exposes any of those values. A submission
may consume an active draft exactly once; concurrent keys for the same draft serialize and
only one can reach engine Submit.

[REQUIRED ORACLES]

1. First request performs one Submit, one receipt finalization and one complete 03D
   `mail.sent.v1` envelope persisted through `PublishInTx(subject, envelope)`.
2. Same key/input after success replays byte-equivalent receipt; zero Submit/outbox.
3. Same key/different canonical input returns idempotency conflict; zero Submit/outbox.
4. Timeout/indeterminate produces `unknown`; retry and receipt read call reconcile first.
5. Reconcile sent finalizes once; repeated read/retry produces no second outbox.
6. Reconcile definitive absence permits one Submit with the original stable engine key.
7. Reconcile indeterminate never calls Submit even across multiple retries.
8. Two different keys racing on one draft result in at most one engine Submit; loser gets a
   stable draft/submission conflict without leaking winner metadata.
9. Draft is immutable while queued/submitting/unknown, returns active only after definitive
   failure/absence policy, and reaches submitted atomically with confirmed delivery.
10. Pending/rejected/error attachment, unauthorized From identity, over-limit recipients,
   suspended account or stale/deleted draft fail before engine call.
11. Cross-account draft/submission id makes zero engine/repository mutation.
12. Repository snapshots and response values contain no raw HTTP key, body, MIME,
    credential or raw SMTP/provider error.
13. Context cancellation after engine confirmation but before Hub finalization converges
    through reconciliation to one sent receipt and one outbox event.
14. Event id/time/org/actor/correlation are explicit stable inputs; retry/reconcile never
    regenerates a second envelope id. Stored event type equals JSON subject, and the fake
    outbox receives the whole envelope rather than only inner payload.

[EXECUTION / PROVENANCE]

- TDD first; deterministic fake engine/repository/outbox; no real network/DB/clock.
- Host verification/direct Docker/Compose forbidden; missing accepted launcher =>
  `NOT_RUN`, never fabricated PASS.
- No dependency/generated code/commit/push/PR/network. max_depth 0; fan_out 0; retries 2;
  wall clock 60m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none. Mismatch => `BLOCKED`
  without file changes.

[ОТВЕТ]

На русском: state transition matrix; exact changed files/symbols; outbox/idempotency
evidence; verification; requested/actual model; remaining risks.
