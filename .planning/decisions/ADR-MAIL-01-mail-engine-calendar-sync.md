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
    GetMessage(ctx context.Context, accountID, messageID string) (EngineMessage, error)
    SaveDraft(ctx context.Context, draft EngineDraft) (EngineMessage, error)
    Submit(ctx context.Context, message EngineSubmission) (EngineSubmissionResult, error)
    MoveMessages(ctx context.Context, accountID string, messageIDs []string, folderID string) error
    SetFlags(ctx context.Context, accountID string, messageIDs []string, patch FlagPatch) error
    DeleteMessages(ctx context.Context, accountID string, messageIDs []string, hard bool) error
    UpsertSieve(ctx context.Context, accountID string, script SieveScript) error
    Health(ctx context.Context) (EngineHealth, error)
}
```

The interface is a contract sketch; exact Go DTOs are frozen by `MAIL-SPEC.md`. Domain
services never expose engine credentials or raw JMAP objects to HTTP handlers.

### Source of truth

| Data | Authority | Hub PostgreSQL role |
|---|---|---|
| Raw MIME, body parts, protocol headers, immutable message id | Engine | Stores engine id, safe snippet, normalized searchable metadata |
| Mailbox membership and server flags (`seen`, `answered`, `draft`, junk) | Engine | Projection for fast Hub counters; reconciled from engine changes |
| Delivery queue and SMTP result | Engine | Submission status/audit projection |
| System/custom mailbox hierarchy | Engine | `MailFolder` maps Hub user to engine folder and adds color/order |
| Hub labels and pin | Hub PostgreSQL | Canonical `MailLabel`, label edges, pinned projection |
| Thread identity | Hub PostgreSQL | Canonical normalized thread projection derived from message headers |
| Attachment bytes | Engine MIME plus SeaweedFS for Hub uploads | `MailAttachment` stores storage object id, scan state, content metadata |
| Search document | Meilisearch projection | Rebuildable from engine + PostgreSQL metadata |
| Calendar event, attendee, RSVP, sequence | Hub PostgreSQL | Canonical |
| Resend alert templates/settings | Existing alert subsystem | Unchanged |

No message body is duplicated in PostgreSQL. Sanitized HTML may be cached only in a
bounded, invalidatable application cache; it is not canonical persistence.

## Internal delivery

Internal mail uses the same engine submission path as external mail. This decision avoids
split-brain delivery, ensures Sent/Inbox copies and RFC headers are produced once, and
keeps Sieve/spam/audit behavior consistent. “Instant” means Hub publishes the successful
engine change immediately to NATS/WebSocket and refreshes the recipient projection; it
does not mean bypassing the engine.

Idempotency is keyed by a Hub-generated submission id propagated to the engine-facing
request and outbox record. A retry must not create a second visible message.

## Calendar model and iTIP policy

### Authority and ScheduledCall relationship

`CalendarEvent` is first-class and canonical for RSVP. `ScheduledCall` remains the call
execution/scheduling entity and may link one-to-one to a `CalendarEvent`. A calendar event
can exist without a call; a call may temporarily exist without a calendar link during
migration. New meeting creation writes the event and optional call link in one transaction.

Existing active `ScheduledCall.participants` are backfilled as attendees with `accepted`
status because they were already explicit call members in Hub. The creator/organizer is
also `accepted`. Newly invited participants default to `needs_action`.

### UID, SEQUENCE, recurrence, and duplicate handling

- Hub-created events use a stable opaque UID scoped to the instance domain.
- `(calendar_owner_id, uid, recurrence_id)` is the logical idempotency key for imported
  instances. Redelivery updates the same event/instance when `SEQUENCE` is newer.
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

`CalendarAttendee.version` is compared on RSVP writes. A stale client receives `409` and
the current attendee representation. An incoming `METHOD:REPLY` may change only the
attendee whose normalized email matches the authenticated/envelope sender; organizer,
other attendees, event fields, UID, and sequence are immutable from a reply.

### iTIP ownership

Hub renders and parses RFC 5545/5546 payloads using deterministic clocks and stable
serialization. The engine only carries the MIME parts. Supported L0 methods:

- `REQUEST` for create/update;
- `CANCEL` for event/series cancellation;
- `REPLY` for accepted/declined/tentative responses.

Incoming external `REQUEST` creates or updates a `mail_import` event for the addressed
Hub user. External guests are allowed as nullable-user attendees and participate only by
email iTIP in v1; there is no guest Hub session.

## Events

All subjects use JSON envelopes with `schema_version: 1`, UUID `event_id`, RFC 3339
`occurred_at`, `organization_id`, `actor_id` when known, `correlation_id`, and a payload.

| Subject | Producer | Required payload | Consumers |
|---|---|---|---|
| `mail.received.v1` | mailbox ingestion | mailbox/message/thread ids, sender, recipients, received_at, folder ids, has_attachments | projection, search, notification, websocket |
| `mail.sent.v1` | submit outbox | submission/message/thread ids, sender, recipients, sent_at, delivery_class | projection, search, recent contacts, websocket |
| `mail.message.changed.v1` | mailbox service | message id, changed fields, folder ids, flags, labels | projection, search, counters, websocket |
| `calendar.invite.created.v1` | calendar service | event id, uid, sequence, attendee ids | iTIP sender, notification, websocket |
| `calendar.invite.updated.v1` | calendar service | event id, uid, sequence, reset RSVP flag, changed fields | iTIP sender, notification, websocket |
| `calendar.invite.cancelled.v1` | calendar service | event id, uid, sequence, recurrence id if any | iTIP sender, notification, websocket |
| `calendar.rsvp.changed.v1` | calendar service | event/attendee ids, previous/new status, version, recurrence id | organizer notification, websocket |

Consumers are idempotent on `event_id`. Producers write through the existing transactional
outbox pattern where state change and event publication must be atomic.

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
- Marketing, newsletters, digests, bulk campaign tooling, and tracking pixels.
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
