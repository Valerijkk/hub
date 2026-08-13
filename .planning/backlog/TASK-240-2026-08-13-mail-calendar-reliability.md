# TASK-240 — Mail/calendar reliability and product integrations

> Date: **2026-08-13** · Priority: **P1** · Status: **Mail 0.1.4 root static pass on 46cfa26c; independent review pending; code blocked** · Phase: **147**

## Outcome

Close recurrence/import/recovery behavior, busy/free with leave, reminders, the CRM
send/manual-log seam, and, only if founder question 5 admits it, external iCal feed
subscription after the P0 mailbox and RSVP paths.

## Acceptance

- Series and occurrence exceptions converge under redelivery and out-of-order sequence.
- Incoming REQUEST/CANCEL and outward REPLY are bounded, spoof-resistant and recoverable.
- Availability combines first-class events, scheduled calls and approved leave without
  disclosing private titles/details.
- Reminder events reuse notification preferences, keep transient policy/store failures
  pending/retriable, recheck current offset/start/eligibility immediately before durable
  delivery, and do not broaden Resend templates.
- If founder question 5 admits L1, feed ingestion has SSRF protections, owner controls,
  ETag/Last-Modified handling, bounded recurrence expansion, disable/delete semantics, and
  source attribution; otherwise every L1 packet remains explicitly deferred.
- CRM creates one ordinary idempotent Mail draft without recipient-in-URL, then Mail owns
  compose/send/reconciliation and an explicit owned-message activity link, including from
  the ordinary reader; CRM does not become a sender, campaign or tracking system.
- Units and docs trace to FR-CALSYNC-009..013, FR-CRM-MAIL-001..002 and relevant FR-MAIL
  requirements.

## Deferred beyond this card

External-provider IMAP/JMAP account ingest (L2) and Google/Microsoft OAuth (L3).

## Prepared artifacts

- `.planning/phases/147-mail-calendar-reliability/CONTRACT.md`
- `.planning/phases/147-mail-calendar-reliability/DISPATCH-INDEX.md`
- `.planning/phases/147-mail-calendar-reliability/DISPATCH-PACKETS.md`
- `.planning/phases/147-mail-calendar-reliability/STATIC-REVIEW.md`

No implementation/test/runtime evidence exists. Exact Spark, accepted launcher, generation/
migration and the named engine/L1 gates remain mandatory.
