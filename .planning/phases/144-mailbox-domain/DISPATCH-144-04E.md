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
POST send(draft, raw key, confirmed If-Match, canonical request)
  hash key; hash canonical request
  absent -> validate active draft + exact revision, then reserve queued with stable engine_submission_key
         -> validate identity/recipients/clean attachments
         -> atomically freeze draft as submitting
         -> submission submitting -> engine Submit once
         -> confirm opaque engine submission/result-message ids
         -> bounded same-account result read + atomic MailMessage id/version projection upsert
         -> sent + Hub result UUID + draft submitted + full mail.sent.v1 envelope in same Hub transaction
  same key + different request hash -> 409, zero engine call
  same key + sent/failed -> replay same receipt before current draft-version check
  same key + queued/submitting/unknown, no opaque result pair -> ReconcileSubmission first
  same key + unknown, complete opaque result pair -> GetMessage + projection finalize only

Reconcile result
  sent -> persist complete opaque pair -> validate/hydrate/finalize once + one outbox
  definitive absent (no-pair branch only) -> transition queued and permit exactly one
                                             controlled Submit while the draft remains frozen
  indeterminate -> remain unknown; draft remains frozen; never Submit
  definitive rejected -> failed safe code; restore draft active; never retry automatically
```

Raw idempotency key is request-scoped only. Persistence stores key hash, request hash and
stable engine key. The receipt response never exposes any of those values. A submission
may consume an active draft exactly once; concurrent keys for the same draft serialize and
only one can reach engine Submit.

`EngineSubmissionSent.SubmissionID` and `.MessageID` are opaque engine strings, never Hub
UUIDs. Persist them as `engine_submission_id`/`engine_result_message_id`. Public
`message_id` comes only from the separately stored `result_mail_message_id` snapshot after
the bounded projection upsert. If delivery is confirmed but read/upsert/finalization fails,
the receipt remains reconciliation-required with the opaque pair retained and no second
Submit is permitted.

The hydrate result must have exact requested account id, exact persisted
`engine_result_message_id` and a non-empty bounded version. The service derives
`confirmed_owner_submission` from the persisted confirmed pair; it does not trust From,
Sent membership or the adapter-provided origin as that proof. Any mismatch produces no Hub
message binding, `sent` transition or outbox and remains reconciliation-required.

The first submission/reconciliation workflow acquires one owning-account mutation lease and
increments `mutation_epoch` before engine I/O. Its completion/unknown handoff verifies the
token, finalizes the receipt/projection, clears the lease and increments the epoch again in
one transaction. A replay of an already final receipt performs no engine I/O or lease.
Recovery of an unknown receipt acquires a new lease. A row with no opaque result pair invokes
`ReconcileSubmission`; if it proves definitive absence, the controlled Submit may continue
under that same live lease. A row with the complete pair skips `ReconcileSubmission` and can
only repeat bounded hydrate/validate/upsert/finalize; it can never requeue or Submit.
Active-lease busy makes zero engine call. Expiry/reclaim forbids a late former owner from
finalizing directly, and no database transaction remains open across engine I/O.

[REQUIRED ORACLES]

1. First request performs one Submit, one receipt finalization and one complete 03D
   `mail.sent.v1` envelope persisted through `PublishInTx(subject, envelope)`.
2. Same key/input after success replays byte-equivalent receipt; zero Submit/outbox.
3. Same key/different canonical input returns idempotency conflict; zero Submit/outbox.
4. Timeout/indeterminate with no result pair produces `unknown`; retry and receipt read call
   `ReconcileSubmission` first. Definitive absence may requeue only in this branch.
5. Reconcile sent finalizes once; repeated read/retry produces no second outbox. Use
   deliberately non-UUID opaque engine ids and prove the public receipt contains only the
   separately resolved Hub message UUID; result hydration/upsert failure stays unknown and
   the retained-pair branch skips `ReconcileSubmission` and retries hydration, not Submit.
6. Reconcile definitive absence permits one Submit with the original stable engine key.
7. Reconcile indeterminate never calls Submit even across multiple retries.
8. Two different keys racing on one draft result in at most one engine Submit; loser gets a
   stable `MAIL_DRAFT_SUBMISSION_IN_PROGRESS` conflict without leaking winner metadata.
9. Draft is immutable while queued/submitting/unknown, returns active only after definitive
   failure/absence policy, and reaches submitted atomically with confirmed delivery.
10. Pending/rejected/error attachment, unauthorized From identity, over-limit recipients,
   more than 50 selected attachment rows, suspended account or stale/deleted draft fail
   before engine call. A new send key requires
   exact quoted `If-Match`; stale returns `MAIL_DRAFT_VERSION_CONFLICT`, while matching-key
   terminal replay does not fail merely because the draft is now submitted.
11. Cross-account draft/submission id makes zero engine/repository mutation.
12. Repository snapshots and response values contain no raw HTTP key, body, MIME,
    credential or raw SMTP/provider error.
13. Context cancellation after engine confirmation but before Hub finalization converges
    through reconciliation to one sent receipt and one outbox event.
14. Event id/time/org/actor/correlation are explicit stable inputs; retry/reconcile never
    regenerates a second envelope id. Stored event type equals JSON subject, and the fake
    outbox receives the whole envelope rather than only inner payload.
15. Every submission/reconciliation engine workflow has exact mutation-lease begin/end
    epoch edges; only a final stored-receipt replay skips the lease. Active-lease busy makes
    zero engine call, expiry forbids stale finalization, and a sync page cannot capture
    during the workflow or regress confirmed submission/draft projection.
16. Hydrated result account/message must exactly equal the request/persisted opaque result
    and version must be non-empty/bounded. Mismatch and spoofed adapter origin produce no Hub
    binding, `sent`, outbox or second Submit; confirmed-owner origin is domain-derived from
    the persisted result pair.

[EXECUTION / PROVENANCE]

- TDD first; deterministic fake engine/repository/outbox; no real network/DB/clock.
- Host verification/direct Docker/Compose forbidden; missing accepted launcher =>
  `NOT_RUN`, never fabricated PASS.
- No dependency/generated code/commit/push/PR/network. max_depth 0; fan_out 0; retries 2;
  wall clock 60m.
- requested_model `gpt-5.6-terra`; reasoning `xhigh`; fallback none. Mismatch => `BLOCKED`
  without file changes.

[ОТВЕТ]

На русском: state transition matrix; exact changed files/symbols; outbox/idempotency
evidence; verification; requested/actual model; remaining risks.
