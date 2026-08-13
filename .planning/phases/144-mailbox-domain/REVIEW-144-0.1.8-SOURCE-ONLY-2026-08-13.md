# Phase 144 Mail 0.1.8 source-only review — 2026-08-13

> **Reviewed SHA:** `90229b2bedfefd3495b3b8dc54e9bd170d30ad60`
> **Requested runtime:** `gpt-5.6-terra` / `high`, no fallback
> **Actual runtime evidence:** model `UNKNOWN`, effort `UNKNOWN`
> **Gate verdict:** `NOT_ACCEPTED_FOR_GATE`
> **Execution:** static source review only; Docker/tests/build/generation `NOT_RUN`

The requested reviewer could inspect the exact clean source snapshot but the worker runtime
did not expose actual model/effort telemetry. Under founder question 29 this pass cannot
accept the contract or authorize product/test writes. Its technical findings remain useful
review evidence and were independently checked by root before amendment.

## P1 findings and root remediation

1. `EngineMessageProjection.Version` had no durable `MailMessage` field. The root candidate
   now requires `engine_version` 1..1,024 bytes from the same projection as
   `engine_message_id`; auxiliary engine state cannot substitute.
2. `MailSubmission.result_message_id` conflated the opaque engine message id with the Hub
   UUID exposed by OpenAPI. The candidate now stores `engine_result_message_id` separately
   from scalar `result_mail_message_id`, hydrates/upserts the bounded result projection and
   records `sent` only after its Hub UUID exists.
3. OpenAPI capped a draft request at 50 attachment ids but persistence permitted unlimited
   sequential associations. The candidate now requires a full-set/count check under the
   draft lock and rejects a 51st durable association before engine I/O.

## Follow-up source-only findings and root remediation

The first remediation was re-read under the same unavailable-provenance condition and found
two additional P1s:

1. A confirmed-delivery/projection-pending `unknown` row could still enter unconditional
   `ReconcileSubmission` and a definitive-absence requeue. Recovery is now disjoint: only a
   row with no opaque result pair reconciles the submission and may requeue on definitive
   absence; a row with the complete pair retries only bounded result hydration/upsert/
   finalization and can never requeue or submit.
2. Result hydration did not explicitly bind the returned projection to the confirmed result.
   The candidate now requires exact returned account/message equality and non-empty bounded
   version, derives `confirmed_owner_submission` in the domain from the persisted result pair,
   and makes mismatch a no-binding/no-sent/no-outbox reconciliation-required outcome.

These root edits close the reported mechanics but remain `NOT_ACCEPTED_FOR_GATE` until a new
exact-SHA review can prove actual reviewer provenance.

## Final source-only re-review

The final re-read of the complete dirty docs-only remediation found no remaining P0/P1/P2.
It confirmed all five findings are mechanically closed and the OpenAPI public shape is
unchanged beyond documenting that `message_id` is the Hub UUID. The verdict nevertheless
remains `NOT_ACCEPTED_FOR_GATE` solely because actual model/effort evidence is unavailable;
runtime/build/tests/generation remain `NOT_RUN`.

## Remaining gates

- independent exact-SHA review with observable actual Terra/high provenance;
- 144-01A implementation of the 0.1.8 engine amendment and its independent acceptance;
- founder migration authority plus accepted Docker-only generation launcher for schemas;
- runtime tests/build/generation through that launcher only.

No product code, test source, dependency, generated output or migration was changed by this
review/remediation packet.
