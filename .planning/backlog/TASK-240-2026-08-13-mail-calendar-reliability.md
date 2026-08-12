# TASK-240 — Mail/calendar reliability and product integrations

> Date: **2026-08-13** · Priority: **P1** · Status: **Open** · Phase: **147**

## Outcome

Close recurrence/import/recovery behavior, busy/free with leave, reminders, external iCal
feed subscription, and the CRM send/manual-log seam after the P0 mailbox and RSVP paths.

## Acceptance

- Series and occurrence exceptions converge under redelivery and out-of-order sequence.
- Incoming REQUEST/CANCEL and outward REPLY are bounded, spoof-resistant and recoverable.
- Availability combines first-class events, scheduled calls and approved leave without
  disclosing private titles/details.
- Reminder events reuse notification preferences and do not broaden Resend templates.
- L1 feed ingestion has SSRF protections, owner controls, ETag/Last-Modified handling,
  bounded recurrence expansion, disable/delete semantics, and source attribution.
- CRM can compose/send as an authorized user identity and append a manual activity link;
  it does not become a campaign or tracking system.
- Units and docs trace to FR-CALSYNC-009..013 and relevant FR-MAIL requirements.

## Deferred beyond this card

External-provider IMAP/JMAP account ingest (L2) and Google/Microsoft OAuth (L3).
