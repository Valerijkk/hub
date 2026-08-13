# Hub Mail contract changelog

## 0.1.2 — 2026-08-13

Implementation-readiness addendum. No runtime or product-code claim.

- Added explicit PostgreSQL metadata models for message-folder projection, drafts, user
  settings, signatures, domains, mailbox policy and calendar mutation receipts.
- Added a mail-owned object-storage authorization/lifecycle model; project-scoped file
  metadata cannot authorize mailbox attachments.
- Froze draft save recovery through opaque engine versions and `ReconcileDraft`; message
  bodies, recipients, subject and raw MIME remain engine-only canonical data.
- Froze submission request hashing, stable engine keys, draft freezing, explicit
  `ReconcileSubmission`, `unknown` behavior and owner-scoped submission lookup.
- Removed the raw idempotency key from `MailSubmission` API responses.
- Added `GET /api/mail/submissions/{submissionId}` and explicit 409 idempotency responses
  for draft create/send.
- Added canonical deletion of one recurrence instance with a cancelled exception and
  occurrence-scoped iTIP/outbox behavior.
- Split HTTP occurrence RSVP from inbound iTIP REPLY. Inbound replies now use a
  bounded durable per-attendee/sequence/canonical-payload-hash ledger and never an HTTP
  receipt; this prevents delayed `A -> B -> A` redelivery from rolling RSVP backward.
- Made message folder ids and engine-authoritative flag snapshots required in message
  summaries.
- Required `calendar.>` admission in Phase 143 and `mail.>` admission in Phase 144 while
  preserving the outbound webhook filter.
- Added exact Spark dispatch indexes and gate plans for Phases 143–146.

Unchanged compatibility boundaries:

- mandatory NFR-13 error body `{error,request_id}`;
- legacy `GET /api/calendar/events` response `{events,from,to}`;
- canonical calendar CRUD/RSVP under `/api/calendar/v2/**`;
- legacy flat Resend `/api/admin/mail/settings` contract;
- deprecated broadcast remains a documented unresolved governance gap, not new scope.

## 0.1.1 — 2026-08-13

Post-review hardening of the initial frozen contract: legacy calendar compatibility,
canonical v2 calendar paths, timed/all-day validation, same-origin RSVP recovery headers,
persisted calendar idempotency receipts, admin RBAC/audit metadata and explicit legacy
Resend/broadcast boundaries.

## 0.1.0 — 2026-08-13

Initial mail engine, calendar synchronization, OpenAPI, event schema, UI and QA contract.
