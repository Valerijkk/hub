# ADR-MAIL-01 — Mail engine boundary and calendar RSVP synchronization

> **Status:** Accepted with operational gates
> **Date:** 2026-08-13
> **Deciders:** founder product contract, Hub team-lead
> **Scope:** Outlook-class Hub mail client, instance-domain mail, RFC 5545/5546 calendar synchronization
> **Supersedes:** no previous mailbox ADR
> **Coexists with:** `ADR-NOTIF-01-transactional-email.md` (Resend transactional alerts)

## Context

Hub needs a first-class mail client without becoming a mail transport product. The existing
`backend/internal/mail` and `backend/cmd/hub-mail` implement a narrow Resend-based alert
control plane; they do not provide user mailboxes, SMTP receiving, IMAP/JMAP storage,
deliverability, anti-spam, or calendar invitation semantics.

The founder contract fixes the boundary:

- Hub owns product policy, tenant/user identity mapping, UX, audit, search projection,
  labels, calendar links, and notification integration.
- A mature adapter-backed engine owns SMTP/IMAP/JMAP transport, raw MIME, mailbox
  storage, delivery queues, DKIM/SPF/DMARC, and anti-spam.
- We do not build an SMTP server.
- OAuth synchronization with Google or Microsoft remains outside this ADR.

The selected engine must fit an optional Docker profile with a total hard budget of at
most 4 CPU and 4 GiB RAM. No runtime measurement has been performed because the accepted
project launcher is currently `NOT_READY`; direct Compose or ad-hoc Docker execution is
not an allowed substitute.

## Decision summary

1. **Select Stalwart as the preferred engine**, conditionally gated by a founder/legal
   license verdict and a reproducible resource/API spike through the accepted launcher.
2. Define a replaceable `MailEngine` port inside `backend/internal/mail`; keep the
   existing Resend alert behavior intact in the same package and isolate the production
   Stalwart implementation in `backend/internal/mail/stalwart`.
3. Route **internal and external delivery through the engine**. Hub may optimize UX and
   indexing, but does not write directly into two mailbox stores.
4. Treat the engine as source of truth for messages and mailbox state; PostgreSQL stores
   Hub-specific projections and relations only.
5. Treat Hub PostgreSQL as source of truth for calendar events, attendees, RSVP state,
   recurrence policy, and links to scheduled calls. Hub generates and validates iTIP;
   the engine transports and stores the MIME message.
6. Implement L0 completely. Implement instance-domain SMTP receive/send as part of the
   mailbox product. Defer external-provider mailbox ingest (L2) and OAuth two-way sync
   (L3). Keep L1 iCal feeds as a later phase after L0 reliability.
7. Treat the owner-scoped mailbox projection loop as an L0 product component, not adapter
   glue: one leased worker atomically applies each change page with its cursor/outbox,
   fences concurrent Hub operations by an account mutation lease plus epochs, and performs resumable no-clear
   base-anchored enumeration and mandatory change catch-up after an invalid cursor.

## Candidate comparison

Research was refreshed from primary project sources on 2026-08-13. Version and digest
must be resolved again immediately before an approved runtime run.

| Criterion | Stalwart | Maddy | docker-mailserver |
|---|---|---|---|
| Current release observed | `v0.16.15` | `v0.9.5` | `v15.1.0` |
| Hub adapter surface | Native JMAP for mail and Sieve; management APIs | SMTP/IMAP composition; no documented JMAP client surface | Postfix/Dovecot operational bundle; no JMAP/REST mailbox surface |
| Protocols | JMAP, IMAP, SMTP, ManageSieve | SMTP, IMAP; storage documentation still describes IMAP storage conservatively | SMTP, IMAP, Sieve through bundled components |
| Deliverability/security | Built-in DKIM/SPF/DMARC/ARC, throttling, filtering, spam/phishing features | Composable checks; Rspamd integration available | Mature Postfix/Dovecot stack; Rspamd or legacy anti-spam components |
| Migration | JMAP import/export plus standard protocols | Standard SMTP/IMAP paths | Standard SMTP/IMAP paths and file-oriented operations |
| License | Dual AGPL-3.0 / SELv2 according to repository; legal choice required | GPL-3.0 | MIT |
| Operational shape | Single modern server, best API fit | Smaller Go composition but weaker API contract and recent security-fix history | Mature bundle, but Hub would need a larger custom adapter around IMAP/admin scripts |
| Resource evidence in Hub | `NOT_RUN` | `NOT_RUN` | `NOT_RUN` |
| Verdict | **Preferred, gated** | Rejected for first implementation | Fallback if Stalwart licensing is rejected |

Primary sources:

- [RFC 8620 — JMAP core query, anchor and changes semantics](https://www.rfc-editor.org/rfc/rfc8620.html)
- [RFC 8621 — JMAP for Mail](https://www.rfc-editor.org/rfc/rfc8621.html)
- [Stalwart repository and protocol/security feature list](https://github.com/stalwartlabs/stalwart)
- [Stalwart Docker installation](https://stalw.art/docs/install/platform/docker/)
- [Stalwart JMAP overview](https://stalw.art/docs/http/jmap/)
- [Stalwart releases](https://github.com/stalwartlabs/stalwart/releases)
- [Maddy releases](https://github.com/foxcpp/maddy/releases)
- [Maddy project](https://github.com/foxcpp/maddy)
- [docker-mailserver introduction](https://docker-mailserver.github.io/docker-mailserver/latest/introduction/)
- [docker-mailserver releases](https://github.com/docker-mailserver/docker-mailserver/releases)

### Operational gates for Stalwart

Stalwart is not considered runnable or production-approved until all gates pass:

1. Founder/legal explicitly accepts AGPL-3.0 use for an unmodified sidecar or selects
   SELv2 terms. Hub must not copy Stalwart code into the repository.
2. The immutable image reference is pinned by digest after provenance verification.
3. The accepted versioned launcher proves health, persistence, JMAP session discovery,
   authenticated mailbox access, SMTP delivery, and restart survival.
4. Idle and representative-load measurements prove the entire mail profile stays within
   4 CPU / 4 GiB RAM. Until then, resource status is `NOT_RUN`, not “within budget”.
5. Backup and restore steps for the named volume are documented and exercised.

If gate 1 fails, use docker-mailserver behind the same `MailEngine` port and accept the
additional IMAP/admin adapter work. If gates 2–5 fail, stop the runtime phase and keep the
domain implementation on the fake adapter; do not silently switch images or engines.

## Adapter boundary

The existing `backend/internal/mail` package gains explicitly named mailbox-domain files
without rewriting its alert sender. The production adapter is a child package wired at the
composition root, so the domain owns a port intentionally narrower than JMAP:

```go
type MailEngine interface {
    Capabilities(ctx context.Context) (EngineCapabilities, error)
    EnsureAccount(ctx context.Context, account AccountSpec) (EngineAccount, error)
    ListFolders(ctx context.Context, accountID string) ([]EngineFolder, error)
    CreateFolder(ctx context.Context, accountID string, spec FolderSpec) (EngineFolder, error)
    RenameFolder(ctx context.Context, accountID, folderID, name string) error
    DeleteFolder(ctx context.Context, accountID, folderID string) error
    ListMessages(ctx context.Context, query EngineMessageQuery) (EngineMessagePage, error)
    SearchMessages(ctx context.Context, query EngineSearchQuery) (EngineMessagePage, error)
    GetMessage(ctx context.Context, accountID, messageID string) (EngineMessage, error)
    GetBlob(ctx context.Context, accountID, blobID string) (EngineBlob, error)
    Changes(ctx context.Context, accountID, since string) (EngineChangePage, error)
    EnumerateProjection(ctx context.Context, accountID, checkpoint string) (EngineEnumerationPage, error)
    SaveDraft(ctx context.Context, draft EngineDraft) (EngineMessage, error)
    ReconcileDraft(ctx context.Context, draft EngineDraftReconciliation) (EngineDraftResult, error)
    Submit(ctx context.Context, message EngineSubmission) (EngineSubmissionResult, error)
    ReconcileSubmission(ctx context.Context, accountID, submissionKey string) (EngineSubmissionResult, error)
    MoveMessages(ctx context.Context, accountID string, messageIDs []string, folderID string) error
    SetFlags(ctx context.Context, accountID string, messageIDs []string, patch FlagPatch) error
    DeleteMessages(ctx context.Context, accountID string, messageIDs []string, hard bool) error
    UpsertSieve(ctx context.Context, accountID string, script SieveScript) error
    Health(ctx context.Context) (EngineHealth, error)
}
```

The interface is a contract sketch; exact Go DTOs are frozen by `MAIL-SPEC.md`. Domain
services never expose engine credentials or raw JMAP objects to HTTP handlers.
`EngineDraft` carries the previous opaque engine version and a stable Hub mutation id;
`ReconcileDraft` distinguishes applied, absent/unchanged, and indeterminate state without
issuing another save. `EngineSubmission` carries a stable engine submission key. A lost
response is reconciled through engine state before another mutation is attempted. The
domain must not translate an indeterminate timeout into a blind save or send retry.
`Changes` is a bounded hydrated incremental page. `EnumerateProjection` is a resumable
bounded base-state-anchored enumeration page, not a promise that standard JMAP freezes a mailbox
across requests. The adapter captures a composite Mailbox/Email object-state cursor before
enumeration and must prove stable unfiltered ID-anchor traversal in which unchanged
surviving base objects cannot cross the resume
boundary. It then catches up from that base through `Changes`. Query-state drift alone is
expected; missing anchor, unstable enumeration or expired base change history is typed
restart-required, never empty-cursor success. Opaque cursor/checkpoint values stay inside
the adapter/repository boundary and never enter logs, APIs or event payloads.

### Source of truth

| Data | Authority | Hub PostgreSQL role |
|---|---|---|
| Raw MIME, body parts, protocol headers, immutable message id | Engine | Stores engine id, safe snippet, normalized searchable metadata |
| Mailbox membership and server flags (`seen`, `answered`, `draft`, junk) | Engine | Leased owner projection; page + full events + cursor commit atomically; mutation lease plus begin/end epochs prevent stale overwrite |
| Delivery queue and SMTP result | Engine | Submission status/audit projection |
| System/custom mailbox hierarchy | Engine | `MailFolder` maps Hub user to engine folder and adds color/order |
| Hub labels and pin | Hub PostgreSQL | Canonical `MailLabel`, label edges, pinned projection |
| Thread identity | Hub PostgreSQL | Canonical normalized thread projection derived from message headers |
| Attachment bytes | Engine MIME plus SeaweedFS for Hub uploads | Mail-owned `MailObject` authorization/lease plus `MailAttachment` scan and content metadata; project file ACLs are not reused |
| Search document | Meilisearch projection | Rebuildable from engine + PostgreSQL metadata |
| Calendar event, attendee, RSVP, sequence | Hub PostgreSQL | Canonical |
| Resend alert templates/settings | Existing alert subsystem | Unchanged |

No message body is duplicated in PostgreSQL. Sanitized HTML may be cached only in a
bounded, invalidatable application cache; it is not canonical persistence.

The projection worker is context-owned and starts only with an enabled engine. Multiple
backend replicas claim accounts by expiring leases. Engine I/O occurs outside the database
transaction; final page apply locks the account and verifies lease, exact prior cursor,
sync epoch, captured mutation epoch and absence of an active mutation lease. Mismatch
discards the whole page. A user command acquires a separate expiring account mutation
lease and increments the epoch before engine I/O; its completion/reconciliation clears the
lease and increments the epoch again in the same projection transaction. Sync capture is
forbidden while that lease is present. Reclaiming an expired mutation lease also increments
the epoch, and a late former owner must reconcile instead of finalizing directly. Invalid cursor
keeps the old owner-scoped projection visibly degraded, applies a resumable base-anchored
enumeration without destructive pre-clear, catches up from its base cursor, and only then
tombstones unseen engine rows and reports ready. This recovery path is required for L0
instance-domain mail even though unrelated external-account L2 ingest remains deferred.

## Internal delivery

Internal mail uses the same engine submission path as external mail. This decision avoids
split-brain delivery, ensures Sent/Inbox copies and RFC headers are produced once, and
keeps Sieve/spam/audit behavior consistent. “Instant” means Hub publishes the successful
engine change immediately to NATS/WebSocket and refreshes the recipient projection; it
does not mean bypassing the engine.

The sender transaction owns `mail.sent.v1`. The recipient projection worker owns exactly
one `mail.received.v1` for a first-seen post-cutoff inbound-or-unknown engine message.
Suppressing it requires a stable mapped Hub submission/draft provenance; `From`/alias or
mutable Sent membership alone is never confirmed-owner provenance. Rebuild history at/before the
cutoff is projected without an event storm; the persisted message marker prevents replay/
rebuild duplication. Incremental/rebuild
state is always surfaced as projection freshness rather than silently labelled current.

Idempotency is keyed by a Hub-generated submission id propagated to the engine-facing
request and outbox record. The raw HTTP key is stored only as a hash. A timed-out submit
enters `unknown` and calls `ReconcileSubmission` with the same stable engine key before any
resubmit. A retry must not create a second visible message.

## Durable event admission

Before enabling calendar or mailbox producers, the shared JetStream EVENTS configuration
must explicitly admit `calendar.>` and `mail.>`. These subjects remain excluded from the
outbound webhook filter by default: mailbox/calendar metadata is not a general webhook
surface. Adding an individual webhook event requires its own privacy, authorization and
payload review.

The canonical call bridge replaces the current unconsumed best-effort
`calls.scheduled.due` publish with transactional outbox delivery on the existing
`calls.channel.{channelId}.scheduled_due.v1` subject. `calls.channel.>` is already admitted to
EVENTS and routed to the authorized channel WebSocket room, while `calls.*` remains excluded
from the webhook filter. The outbox makes the intent durable; JetStream/consumer delivery is
still at-least-once and the payload occurrence key is mandatory for deduplication.

## HTTP compatibility boundary

The existing `GET /api/calendar/events` aggregate keeps its public
`{events, from, to}` shape and existing call/sprint/deadline/leave behavior. Canonical
first-class event CRUD, RSVP and availability use `/api/calendar/v2/**`; the aggregate may
later expose a new `event` kind only additively. The existing flat
`GET/PUT /api/admin/mail/settings` Resend contract also remains unchanged. Mailbox-engine
policy is a separate `GET/PUT /api/admin/mail/mailbox-settings` resource. Every new admin
mail operation requires the existing admin middleware plus a named audit action.

Repository inspection also found the live `POST /api/admin/mail/broadcast` route. It sends
arbitrary admin-authored content to all active users and therefore conflicts with
ADR-NOTIF-01's template whitelist and no-marketing/bulk decision. This ADR does not bless
or reuse that route. Before mailbox acceptance it must be disabled/removed, or the founder
must explicitly supersede ADR-NOTIF-01 and approve consent, audit, rate, unsubscribe and
abuse controls. Until then OpenAPI marks it deprecated as a current governance gap.

## Calendar model and iTIP policy

### Authority and ScheduledCall relationship

`CalendarEvent` is first-class and canonical for meeting title/description, temporal and
recurrence state, exceptions, reminder, attendee/organizer state and RSVP. `ScheduledCall`
remains a distinct call-execution entity and may link one-to-one to a `CalendarEvent`; for a
linked pair its duplicated schedule fields are a legacy/read/execution projection, never a
second write authority. A calendar event can exist without a call. An active call may exist
without a calendar link only before the all-or-nothing migration cutover.

New creation through the legacy scheduled-call API writes the call projection, canonical
event, attendees and invitation outbox atomically. Calendar-v2 event creation may attach an
existing active, unlinked call only when the authenticated organizer owns it and the server
proves the same organization/project/channel boundary; v1 does not accept a nested call
creation object. The normalized create request must exactly match the locked call's title,
description, start/duration, recurrence, reminder, organizer and current participant set;
attach never silently chooses a winner or edits the pre-link call. The organizer must be an
active exact-channel member; organizer and every internal participant must be active members
of the same project and organization. Ordinary
CalendarEvents still allow email-only guests, but the safe unattended default rejects an
external guest on a linked-call event until founder question 23 accepts a join-code/link
delivery, expiry, revocation and audit contract. The call link is immutable after creation.
Calendar-v2 and legacy edits
lock the event before the call, apply event version/sequence policy, then mirror the call in
the same transaction. Because the legacy wire contract has no `If-Match`, its mutation is
serialized last-write-wins against the locked current event rather than pretending to offer
optimistic locking. Existing response fields remain compatible.

The due worker treats `scheduled_at` as the next effective execution occurrence only. It
derives recurrence from the canonical event plus EXDATE/cancelled exceptions, writes one
durable occurrence-keyed `calls.channel.{channelId}.scheduled_due.v1` outbox intent and
advances/deactivates the projection in the same event-before-call transaction. It never
advances the event master,
changes RSVP/SEQUENCE, or creates another reminder. NATS delivery stays at-least-once and
the deterministic occurrence key is the consumer dedupe key. Cancelling the next occurrence
advances the projection atomically; completing a one-time call deactivates only the call and
preserves the historical event. Legacy DELETE cancels the event and soft-deletes the call.
The provisional `is_active=false/true` mapping pauses/resumes call execution without
cancelling/resurrecting the event, pending founder question 19.

Compatibility readers never double count a linked pair. `GET /api/calendar/events` emits
it once as the existing `call` kind, populated from canonical temporal truth, and emits only
unlinked first-class events as additive `event`. Availability emits every canonical event
once as `event` and active unlinked calls only behind the pre-cutover transition flag. Once
the bridge is active, any unlinked non-deleted call is an invariant failure. The old
`/api/calls/conflicts` detail surface is not an authorization source for Mail meeting UX;
new UI uses privacy-preserving v2 availability, and legacy conflict details require caller
membership in the exact conflicting channel.

The ScheduledCall participant M2M is a series-level projection only. It contains distinct
internal non-organizer series attendees whose state is `needs_action`, `accepted`, or
`tentative`, and removes `declined` or deleted attendees in the same bridge transaction.
An occurrence-specific RSVP never mutates that M2M and never changes access to another
occurrence.

Every non-deleted `ScheduledCall`, including paused/completed rows, is backfilled; otherwise
legacy resume could recreate an unlinked authority. Its existing participants become
attendees with `accepted`
status because they were already explicit call members in Hub. The creator/organizer is
also `accepted`. For this backfill, `calendar_owner_id` and `organizer_id` are the call's
`created_by`, `organization_id` is the call project's required organization, UID is the
deterministic `urn:uuid:<scheduled_call_id>`, and title, description, instant, duration,
reminder and supported recurrence are mapped explicitly. Project/channel mismatch, missing
organization, inactive or cross-organization creator/participant, invalid normalized email,
timezone, interval, reminder or recurrence fails preflight and is not guessed or silently
dropped. Creator and participants must also be active members of the mapped project; the
creator must belong to the exact call channel. Newly invited participants default to
`needs_action`.

Legacy recurring calls need a separate founder decision because their stored
`scheduled_at` is a moving next occurrence and existing code interprets raw RRULE and an
injected UTC DTSTART differently. Until question 18 chooses exact-UTC preservation versus
creator-local reinterpretation, any non-deleted recurring row blocks the whole apply. The
backfill writes no invitation/outbox/notification records and sends no historical mail.

### UID, SEQUENCE, recurrence, and duplicate handling

- Hub-created events use a stable opaque UID scoped to the instance domain.
- `(calendar_owner_id, uid)` identifies the imported master. An occurrence is uniquely
  identified by `(calendar_event_id, recurrence_id)` in `CalendarEventException`.
  Redelivery updates the same event/instance when `SEQUENCE` is newer.
- Lower sequence is ignored. Equal sequence is idempotent when the canonical payload hash
  matches and rejected/quarantined when it conflicts.
- `SEQUENCE` increments on invitation-relevant changes: time, duration, timezone,
  all-day state, location, recurrence, organizer, attendee set, or cancellation.
- Description-only changes preserve attendee RSVP but still send an update only when the
  organizer explicitly chooses “notify participants”; that notified update increments
  sequence. Silent description edits do not.
- Time, timezone, all-day, location, or recurrence changes reset all non-organizer
  attendees to `needs_action`. Attendee-list-only changes preserve existing attendees and
  initialize only new attendees.
- Series response and occurrence response are distinct. An occurrence response includes
  `RECURRENCE-ID`; without it, the response targets the series.

### Optimistic locking and spoof defense

`CalendarAttendee.version` is compared on RSVP writes. A stale client receives `409` with
the canonical Hub NFR-13 body `{error,request_id}`, `X-Hub-Error-Code:
ATTENDEE_VERSION_CONFLICT`, the current attendee-version `ETag`, and `Location` of the
canonical event. The client must refetch that event and reconcile before retry; the error
body is not expanded because Hub forbids endpoint-specific wire error shapes. An incoming
`METHOD:REPLY` may change only the attendee whose normalized email matches the
authenticated/envelope sender; organizer, other attendees, event fields, UID, and sequence
are immutable from a reply.

Create and RSVP request idempotency is persisted in the Hub transaction as a bounded
receipt keyed by organization, authenticated actor, operation and a hash of
`Idempotency-Key`. The receipt also stores a canonical request hash and the original
successful response. A key reused for different input is a conflict; the same completed
input replays the original response without a second mutation or outbox event.

### iTIP ownership

Hub renders and parses RFC 5545/5546 payloads using deterministic clocks and stable
serialization. The engine only carries the MIME parts. Supported L0 methods:

- `REQUEST` for create/update;
- `CANCEL` for event/series cancellation;
- `REPLY` for accepted/declined/tentative responses.

Incoming external `REQUEST` creates or updates a `mail_import` event for the addressed
Hub user. External guests are allowed as nullable-user attendees and participate only by
email iTIP in v1; there is no guest Hub session.

### Reliability and integration addendum

Incoming REQUEST/CANCEL is admitted only when the engine proves an authenticated local
identity or an external DMARC-aligned From that equals ORGANIZER. Parsing alone, forwarding
and unverified delegation are not trust. Ordering is stored per calendar owner + UID +
optional recurrence id: lower sequence is stale, exact equal method/hash is duplicate,
equal conflict quarantines, and newer input advances. A valid CANCEL before REQUEST creates
a tombstone rather than an incomplete event, so delayed mail cannot resurrect it. After
the first admitted object, organizer identity is immutable; a different authenticated
sender cannot take over a known UID merely by presenting a higher sequence.
Unbound CANCEL tombstones remain for the mailbox/account lifetime and count against a hard
10,000-row per-mailbox unbound-state cap; they are never expired merely to admit a new UID.
Only a missing-master deferred occurrence expires after seven days with its dead-lettered
processing receipt.
[RFC 5546 §2.1.5](https://www.rfc-editor.org/rfc/rfc5546.html#section-2.1.5)
orders revisions for the same UID plus optional RECURRENCE-ID. Hub preserves those distinct
component keys but additionally locks the series admission record before an occurrence:
the organizer must match, an absent/cancelled master keeps an occurrence REQUEST deferred,
and whole-series CANCEL makes all materialized exceptions effectively cancelled. Thus an
occurrence-only message cannot attach to another organizer's UID or resurrect a cancelled
series.

External-organizer RSVP delivery is an immutable ledger intent created with the attendee
transition. It builds a hidden engine draft and `MailSubmission` from stable derived keys;
indeterminate results call reconciliation before any retry. The iTIP body remains request-
lifetime engine data.

Availability authorizes the full user set first under the existing self/shared-project-or-
channel boundary (never admin org-wide bypass), then combines canonical events, active
scheduled calls and policy-visible approved leave into busy/free intervals without names,
ids, reasons, projects or participant detail. Calendar reminder scheduling persists one intent and uses
a persistent notification dedupe key for one terminal in-app outcome: exactly one row when
policy permits or durable suppression when it does not. Telegram/Web Push preserve their
existing best-effort boundary; the decision does not add Resend/email notifications.
Transient user/policy/repository lookup failure remains pending and retriable rather than
being recorded as suppression; post-commit WebSocket publication is also best-effort.
Both the due claim and notification transaction revalidate the current effective start,
sequence, non-null `reminder_minutes`, derived due time and owner/attendee eligibility.
Reminder-relevant mutations serialize with those checks in event → optional linked
ScheduledCall → optional exception → optional attendee → user-policy guard → receipt →
notification lock order. The guard is the stable user row;
notification-preference upsert locks it before resolving/upserting optional preference
rows, and a quiet-hours profile update locks that row before changing it. Thus a committed
in-app channel disable, quiet-hours change, reminder disable/reschedule or decline cannot
leak a stale in-app reminder. Telegram/Web Push attempts happen only after the canonical
transaction commits, remain outside it, and retain their best-effort race boundary; a
rollback produces no external attempt. All-day reminder scheduling uses the first
representable instant of the local date in the event IANA timezone: a midnight overlap
selects the earlier instant and a gap/skipped date selects the first instant after the gap,
without changing the date-only event representation. The keyed calendar path rejects
legacy fail-open policy interpretation:
invalid timezone or incomplete/malformed quiet hours durably suppresses with a safe reason
and bounded content-free signal, never UTC/delivery fallback.

CRM compose creates one ordinary idempotent Mail draft and opens its owner-scoped UUID in
the normal mailbox composer; recipient data never travels in the URL. It is not a second
sender. A manual deal timeline link references an already authorized message projection
and stores no mail content or tracking. Draft/message deep links reauthorize through Mail;
CRM/admin access never delegates another user's mailbox. Full field/state rules are frozen
in Phase 147 `CONTRACT.md`.

## Events

All subjects use JSON envelopes with the exact broker `subject`, `schema_version: 1`, UUID
application `event_id`, RFC 3339 `occurred_at`, `organization_id`, `actor_id` when known,
`correlation_id`, and a payload. A shared pure builder receives every value explicitly;
it owns no clock, random source, persistence or publish side effect.

| Subject | Producer | Required payload | Consumers |
|---|---|---|---|
| `mail.received.v1` | mailbox ingestion | mailbox/message/thread ids, sender, recipients, received_at, folder ids, has_attachments | projection, search, notification, websocket |
| `mail.sent.v1` | submit outbox | submission/message/thread ids, sender, recipients, sent_at, delivery_class | projection, search, recent contacts, websocket |
| `mail.message.changed.v1` | mailbox service | message id, changed fields, folder ids, flags, labels | projection, search, counters, websocket |
| `calendar.invite.created.v1` | calendar service | event id, uid, sequence, attendee ids | iTIP sender, notification, websocket |
| `calendar.invite.updated.v1` | calendar service | event id, uid, sequence, reset RSVP flag, changed fields | iTIP sender, notification, websocket |
| `calendar.invite.cancelled.v1` | calendar service | event id, uid, sequence, recurrence id if any | iTIP sender, notification, websocket |
| `calendar.rsvp.changed.v1` | calendar service | event/attendee ids, previous/new status, version, recurrence id | organizer notification, websocket |
| `calendar.reminder.due.v1` | calendar reminder scheduler | receipt/event/user ids, sequence, recurrence id, due_at | notification consumer |

Consumers are idempotent on envelope `event_id`. Producers write the complete envelope
through the existing transactional outbox pattern as
`PublishInTx(ctx, tx, envelope.Subject, envelope)`, so state change and event persistence
remain atomic. `OutboxEvent.event_type`, the NATS subject and envelope `subject` must be
identical; a mismatch is dead-lettered before mutation. The existing outbox worker remains
unchanged and publishes its stored bytes as-is. Its row UUID/`Hub-Event-Id` header is a
separate transport identifier used only for diagnostics by these new consumers; it does
not replace the application event id. This additive rule deliberately avoids wrapping or
changing payloads for any legacy outbox subject.

## Security model

- Every mailbox query is scoped from authenticated Hub user to exactly one owned engine
  account. Future delegation requires a separate permission relation and ADR amendment.
- Admin engine/domain/settings endpoints use existing RBAC middleware and emit audit
  records. Bulk message operations also emit audit summaries without message body data.
- HTML is sanitized server-side with an allowlist; scripts, event handlers, forms,
  embedded objects, dangerous URLs, CSS URL fetches, and active SVG are removed.
- The reader uses a restrictive CSP/sandbox boundary. Remote images are blocked by
  default. User-authorized display goes through a size/time/type-limited SSRF-safe proxy
  that rejects private, loopback, link-local, metadata, and rebinding targets.
- SeaweedFS uploads use existing presign/confirm flow. Attachment metadata enforces total
  and per-file limits, safe filename handling, blocked executable types, and a pluggable
  `AttachmentScanner`; unscanned/quarantined data cannot be submitted.
- Send operations have per-user and per-instance token buckets. Automatic replies carry
  loop-prevention headers and never answer another automatic/bulk response.
- Engine/admin secrets and DKIM material live only in secrets/env/volumes, never Git or
  API responses.
- iTIP sender matching, UID/sequence checks, attendee-version checks, MIME depth/size
  bounds, and calendar parser limits are mandatory before state mutation.
- Search indexes store only the minimum authorized projection and are always filtered by
  organization + mailbox owner before returning identifiers.

## Product scope

### Included

- L0 invitations, updates, cancellation, replies, recurrence instance replies, timezone,
  DST, all-day, duplicate UID, optimistic lock, busy/free including leave, reminders.
- Inbox/Sent/Drafts/Archive/Trash/Junk plus nested custom folders and labels.
- Threads, flags, pin, unread counters, bulk actions, drafts/autosave, signatures,
  aliases, CC/BCC, priority, reply/reply-all/forward, SeaweedFS attachments.
- Instance-domain inbound/outbound SMTP through the engine.
- Meilisearch projection indexed asynchronously from NATS.
- Hub users and recent external recipients for autocomplete.

### Deferred

- Read receipts: deferred because of privacy, interoperability, and spoofing cost.
- Out-of-office: schema/loop-prevention seam only; no v1 UI or sending behavior.
- Group addresses and aliases beyond per-user aliases: modeled for wave 2.
- User Sieve rules: adapter supports them, but UI/rule compiler follows the base mailbox.
- L1 external iCal feeds: after L0 reliability and ownership controls.
- L2 ingest from unrelated external IMAP/JMAP mailboxes.
- L3 Google/Microsoft OAuth two-way synchronization.
- Antivirus engine implementation: only a fail-closed scanner hook in v1.

### Explicitly excluded

- Bespoke SMTP/IMAP/JMAP servers.
- New marketing, newsletters, digests, bulk campaign tooling, and tracking pixels. The
  current legacy admin broadcast is a known policy violation routed to the release gate,
  not an accepted exception.
- External analytics/tracking beacons.
- Changes to the Resend alert whitelist from ADR-NOTIF-01.

## Consequences

### Positive

- One canonical mailbox store and one canonical calendar store.
- Replaceable engine implementation and deterministic unit tests with a fake adapter.
- Fast Hub UX without pretending that Hub itself is an MTA.
- Calendar behavior is explicit enough for API, contract, integration, and E2E tests.

### Costs and risks

- Stalwart license acceptance is an external decision gate.
- JMAP change reconciliation, outbox idempotency, and projection repair add operational
  complexity.
- Real deliverability still depends on DNS, public IP reputation, PTR, and operator
  configuration; local Docker health cannot prove Internet delivery.
- Full Outlook parity spans multiple phases and must not be labelled complete when only a
  slice is implemented.

## Verification state

| Gate | Status | Evidence |
|---|---|---|
| Primary-source engine research | PASS | linked sources, refreshed 2026-08-13 |
| Architecture decision | PASS | this ADR |
| Legal license acceptance | PENDING | founder/legal decision required |
| Docker image provenance/digest | NOT_RUN | runtime phase not started |
| Resource budget | NOT_RUN | accepted launcher unavailable |
| JMAP/SMTP/persistence spike | NOT_RUN | accepted launcher unavailable |
