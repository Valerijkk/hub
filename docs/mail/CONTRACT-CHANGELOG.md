# Hub Mail contract changelog

## 0.1.5 — 2026-08-13

ScheduledCall canonical-cutover and mailbox-projection amendment. Product code and runtime remain
`NOT_IMPLEMENTED` / `NOT_RUN`; the 0.1.4 review is historical `STALE` for this revision and
current semantic SHA review is pending.

- Made `CalendarEvent` the only business authority for linked meeting content, time,
  recurrence/exceptions, reminder, status and attendees; `ScheduledCall` is a distinct
  execution/legacy projection with a next effective occurrence.
- Froze atomic legacy create/update/delete, calendar-v2 attach/edit, event-before-call lock
  order, occurrence-cancel projection advancement and occurrence-keyed durable call-due
  outbox behavior on `calls.channel.{channelId}.scheduled_due.v1`. Legacy PUT keeps
  serialized last-write-wins because its wire has no `If-Match`; the successful response
  shape stays compatible.
- Closed reader duplication: legacy aggregate shows a linked pair once as `call`, while v2
  availability shows it once as `event` and loads active unlinked calls only before cutover.
  The backfill now includes paused/completed non-deleted rows so legacy resume cannot restore
  an unlinked writer; any unlinked non-deleted row after activation is an invariant failure.
  New meeting conflict UX must use private v2 availability; legacy conflict detail requires
  membership in the exact conflicting channel.
- Expanded ScheduledCall backfill to organization/project/channel coherence, description,
  duration, reminder, recurrence, participant identity, audit timestamps and deterministic
  hashes. Recurring rows block apply until founder-selected UTC-versus-local semantics and
  bounded occurrence-oracle evidence agree; partial migration and guessing are forbidden.
- Made event `scheduled_call_id` an immutable creation-only, server-authorized link in
  OpenAPI. Attach requires exact normalized equality with the locked call, and the safe
  default rejects external guests until a join-credential delivery contract exists. Series
  attendee state alone drives the internal participant projection; occurrence RSVP cannot.
- Froze mandatory legacy-create idempotency as an explicit compatibility decision, strict
  conflict UUID parsing, exact-channel/project membership and canonical server-side
  occurrence delivery to the browser. Added founder questions 18–23 for recurrence,
  pause/resume, nested call creation, overdue recovery, idempotency compatibility and linked
  external guests.
- Closed the outbox/envelope ambiguity: a shared deterministic builder creates the complete
  versioned envelope, new producers store it unchanged through the existing `PublishInTx`
  seam with `event_type == subject`, and the global worker/legacy payload contract remains
  untouched. Envelope `event_id` is the consumer dedupe id; the outbox row
  `Hub-Event-Id` remains a distinct transport diagnostic id. Phase 144 event mapping now
  precedes transactional producers instead of depending on them.
- Closed the L0 mailbox ingestion/recovery gap: one leased bounded worker owns hydrated
  incremental pages, separate command-lease begin/end epoch fencing, atomic projection/
  event/cursor commits and a
  no-clear resumable base-anchored enumeration/catch-up rebuild. First-seen post-cutoff
  inbound mail emits one stable received event; history and only stably mapped Hub
  submission/draft copies do not. From/alias or mutable Sent membership never proves owner
  origin. Phase 145 now owns incremental and base-state/ID-anchored enumeration adapters
  plus root-context worker composition, while OpenAPI/UI expose bootstrapping/degraded/
  rebuilding freshness without false-ready state.
- Closed the command/sync overlap race: sync cannot capture while an account command lease
  is present; command completion advances a second epoch edge, expired-lease reclaim also
  advances it, and a late former owner must reconcile instead of finalizing stale state.
- Froze the 100-folder total account cap with engine-atomic enforcement, a machine-readable
  422 response and fail-closed degraded projection for an already over-cap engine account.

## 0.1.4 — 2026-08-13

Reminder/CRM race-closure amendment. Product code and runtime remain
`NOT_IMPLEMENTED` / `NOT_RUN`; the prior 0.1.3 root review stays historical `STALE`, while
the 0.1.4 root static review passed on exact semantic SHA `46cfa26c`. Independent review is
still pending.

- Defined timed and all-day reminder due derivation, including deterministic midnight
  overlap/gap/skipped-date behavior, and made both scheduler emission and notification
  finalization recheck current sequence, effective start, non-null offset, derived due and
  owner/attendee eligibility under one lock order. A stable user-row guard also serializes
  quiet-hours and absent/existing preference changes.
- Defined concurrent mutation semantics: mutation-first durably suppresses stale work;
  consumer-first preserves only the already committed valid row. Transient reads/writes
  remain pending/retriable. Telegram/Web Push are post-commit-only best effort, with no
  attempt on rollback and no authority to change the canonical outcome.
- Invalid timezone or incomplete/malformed quiet hours now suppresses with a safe bounded
  reason/signal instead of the legacy UTC/delivery fail-open behavior.
- Added the missing atomic UI packet and QA case for manually linking an existing owned
  inbound/outbound message from the ordinary reader, showing only current-user owned deals
  while preserving independent server authorization.
- Tightened RSVP conflict `Location` to a same-origin canonical UUID event path.
- Bumped OpenAPI metadata to 0.1.4; endpoint and event payload shapes are unchanged.

## 0.1.3 — 2026-08-13

Phase 147 reliability/integration addendum. Product code and runtime remain
`NOT_IMPLEMENTED` / `NOT_RUN`.

- Added durable owner/UID/occurrence ordering and per-message processing receipts for
  authenticated incoming REQUEST/CANCEL, including cancel-before-request tombstones and
  missing-master occurrence recovery. CANCEL tombstones persist for mailbox/account life
  under a hard cap; only deferred missing-master state expires.
- Added immutable external-organizer REPLY delivery reconciliation through stable mailbox
  draft/submission keys and froze its conservative public status mapping.
- Froze privacy-preserving availability: all-or-nothing collaboration-scoped authorization
  for self or same-organization shared-project-or-channel users, bounded event/call/
  policy-visible approved-leave intervals and free complements without private detail.
- Added durable reminder intents and `calendar.reminder.due.v1`; transient policy/store
  failures remain pending/retriable, in-app is persistently deduplicated, and post-commit
  WebSocket/Telegram/push remain explicitly best-effort while Resend stays unchanged.
- Added the CRM compose-context and idempotent manual deal/message link API. `Write email`
  creates one ordinary Mail draft and navigates by its opaque UUID without putting the
  recipient in URL/history. The normal mailbox remains the only send path; no body,
  campaign or tracking data enters CRM.
- Prepared but did not activate the L1 feed contract pending founder question 5; L2/L3
  remain deferred.
- Added exact Phase 147 Spark dispatch packets and named QA scenarios 015–025 plus CRM
  cases. No code packet is eligible while the locked runtime and launcher gates are absent.

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
