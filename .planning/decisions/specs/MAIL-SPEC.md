# MAIL-SPEC — Hub mailbox and calendar synchronization contract

> **Version:** 0.1.2
> **Status:** Frozen for implementation
> **Date:** 2026-08-13
> **Architecture:** `../ADR-MAIL-01-mail-engine-calendar-sync.md`
> **OpenAPI:** `../../../docs/mail/openapi.yaml`
> **QA handoff:** `../../../docs/mail/QA-HANDOFF.md`

## 1. Scope and terms

This document freezes the domain, HTTP, event, security, and traceability contracts for
the Hub mail feature. It does not claim that the described behavior is implemented.

- **Mailbox engine:** adapter-backed Stalwart candidate; owns raw mail and mailbox state.
- **Mailbox:** one engine account owned by one Hub user in one organization.
- **Identity:** a permitted `From` address attached to a mailbox.
- **Message:** immutable RFC message content plus mutable engine mailbox/flag state.
- **Thread:** Hub projection grouping messages by references and normalized subject.
- **Event:** canonical Hub calendar object with stable iCalendar UID.
- **Occurrence:** one recurrence instance identified by UTC-normalized `RECURRENCE-ID`.
- **Projection:** rebuildable Hub metadata derived from engine state.

## 2. Domain invariants

1. Every HTTP mailbox operation derives `organization_id` and owner `user_id` from the
   authenticated principal; clients never choose them.
2. Every engine identifier is opaque. An id from mailbox A is invalid in mailbox B.
3. Raw MIME and message body are never persisted in PostgreSQL.
4. A mail submission is idempotent on `(mailbox_id, hash(idempotency_key))`; the raw key
   is request-scoped only.
5. A canonical calendar master is unique on `(calendar_owner_id, uid)`. Imported
   occurrences are unique on `(calendar_event_id, recurrence_id)` in
   `CalendarEventException`; `CalendarEvent` itself never carries `recurrence_id`.
6. Calendar create/RSVP HTTP retries are serialized through a persisted
   `CalendarMutationReceipt`; a reused key with a different request hash is a conflict,
   while a matching completed request replays the original status and response.
7. RSVP updates compare attendee `version`; exactly one successful transition increments
   it by one.
8. Incoming iTIP replies may mutate only the sender-matching attendee row.
9. Soft-deleted projection rows never authorize engine access.
10. Search results are post-filtered by current organization + mailbox ownership before
   any snippet or identifier is returned.
11. All externally observable timestamps are RFC 3339. Event timezone is a valid IANA
    name. All-day values use dates and never infer midnight instants.
12. Hub-uploaded mail bytes use a mail-owned storage namespace and authorization port;
    project-scoped `FileMetadata` is never reused as an implicit mailbox ACL.
13. Raw idempotency keys, engine credentials, DKIM private keys, message bodies, and raw
    MIME are never persisted in Hub PostgreSQL or returned by read APIs.
14. Durable `mail.*` and `calendar.*` outbox events are admitted to the shared EVENTS
    stream explicitly; neither subject family is added to the outbound webhook filter by
    default because payloads contain mailbox/calendar metadata.
15. Inbound iTIP REPLY delivery is deduplicated by a bounded durable ledger of every
    processed sender-bound payload hash for attendee + event sequence; storing only the
    last hash is forbidden because delayed redelivery could roll a later RSVP backward.

## 3. ERD

```mermaid
erDiagram
    Organization ||--o{ MailDomain : configures
    Organization ||--o| MailboxPolicy : governs
    User ||--o| MailAccount : owns
    MailAccount ||--o{ MailIdentity : exposes
    MailAccount ||--o{ MailFolder : maps
    MailAccount ||--o{ MailThread : projects
    MailThread ||--o{ MailMessage : groups
    MailMessage ||--o{ MailMessageFolder : projects
    MailFolder ||--o{ MailMessageFolder : contains
    MailMessage }o--o{ MailLabel : tagged
    MailMessage ||--o{ MailAttachment : has
    MailAccount ||--o{ MailObject : authorizes
    MailObject o|--o| MailAttachment : backs
    MailAccount ||--o{ MailDraft : edits
    MailDraft }o--o{ MailAttachment : stages
    MailAccount ||--o{ MailLabel : owns
    MailAccount ||--o{ MailRule : owns
    MailAccount ||--o{ RecentContact : remembers
    MailAccount ||--o| MailUserSettings : configures
    MailAccount ||--o{ MailSignature : owns
    MailAccount ||--o{ MailSubmission : submits
    MailDraft o|--o{ MailSubmission : sends
    MailSubmission }o--o| MailMessage : resolves_to
    User ||--o{ CalendarEvent : organizes
    CalendarEvent ||--o{ CalendarAttendee : invites
    CalendarAttendee ||--o{ CalendarInboundReplyReceipt : deduplicates
    CalendarEvent ||--o{ CalendarEventException : overrides
    User ||--o{ CalendarMutationReceipt : initiates
    CalendarEvent o|--o{ CalendarMutationReceipt : resolves
    CalendarEvent }o--o| ScheduledCall : links
    MailMessage }o--o| CalendarEvent : carries
    User o|--o{ CalendarAttendee : identifies
```

### 3.1 MailAccount

| Field | Type | Contract |
|---|---|---|
| `id` | UUID | primary key |
| `organization_id` | UUID | required, indexed |
| `user_id` | UUID | required, unique with organization |
| `engine_account_id` | string | opaque, encrypted-at-rest if credential-like, unique |
| `status` | enum | `provisioning`, `active`, `suspended`, `error` |
| `sync_cursor` | string nullable | opaque JMAP/adapter state token |
| `last_synced_at` | timestamp nullable | projection freshness |
| `created_at`, `updated_at` | timestamp | required |

No password, token, DKIM key, or engine administrator secret is stored on this entity.

### 3.2 MailIdentity

| Field | Type | Contract |
|---|---|---|
| `id` | UUID | primary key |
| `mail_account_id` | UUID | required |
| `address` | normalized email | unique within organization |
| `display_name` | string | max 160 |
| `kind` | enum | `primary`, `alias`, `group` |
| `send_allowed` | bool | server-authoritative |
| `is_default` | bool | at most one per account |
| `engine_identity_id` | string nullable | opaque mapping |
| `created_at`, `updated_at` | timestamp | required |

Group identity is schema-only in wave 1. The API must not expose it as selectable until
group authorization and fan-out are implemented.

### 3.3 MailFolder

| Field | Type | Contract |
|---|---|---|
| `id` | UUID | Hub projection id |
| `mail_account_id` | UUID | required |
| `engine_folder_id` | string | opaque, unique per account |
| `parent_id` | UUID nullable | same-account folder only; depth at most 2 in v1 |
| `role` | enum nullable | `inbox`, `sent`, `drafts`, `archive`, `trash`, `junk` |
| `name` | string | 1..120; custom sibling names unique case-insensitively |
| `color` | string nullable | design-token-safe color value |
| `sort_order` | int | stable user ordering |
| `unread_count`, `total_count` | non-negative int | rebuildable projection |
| `created_at`, `updated_at`, `deleted_at` | timestamp | soft delete for projection |

System-role folders cannot be deleted or reparented. Folder cycles are rejected.

### 3.4 MailThread

| Field | Type | Contract |
|---|---|---|
| `id` | UUID | primary key |
| `mail_account_id` | UUID | required |
| `thread_key` | SHA-256/base64url | unique per account |
| `normalized_subject` | string | prefix-normalized, Unicode-normalized |
| `participant_addresses` | JSON array | normalized unique addresses, bounded |
| `last_message_at` | timestamp | sort key |
| `message_count`, `unread_count` | non-negative int | projection |
| `is_pinned` | bool | Hub canonical |
| `has_attachments` | bool | projection |
| `snippet` | string | sanitized plain text, max 400 graphemes |
| `created_at`, `updated_at`, `deleted_at` | timestamp | projection lifecycle |

Thread key algorithm:

1. Prefer the nearest existing ancestor resolved from valid `References`, then
   `In-Reply-To`.
2. If no ancestor is present, derive a bounded fallback from normalized subject and
   canonical participant set.
3. Strip localized reply/forward prefixes iteratively with a bounded table; do not strip
   arbitrary colon-delimited text.
4. A later header-based match may merge fallback projections atomically; it never changes
   the engine message id.

### 3.5 MailMessage

| Field | Type | Contract |
|---|---|---|
| `id` | UUID | Hub projection id |
| `mail_account_id`, `mail_thread_id` | UUID | required |
| `engine_message_id` | string | opaque, unique per account |
| `internet_message_id` | string nullable | normalized RFC Message-ID, indexed |
| `in_reply_to` | string nullable | normalized RFC Message-ID |
| `references` | JSON array | bounded normalized Message-IDs |
| `from`, `to`, `cc`, `bcc`, `reply_to` | structured JSON | normalized addresses; BCC only for owner-sent copy |
| `subject` | string | max 998 octets at transport boundary |
| `snippet` | string | safe plain-text projection, max 400 graphemes |
| `sent_at`, `received_at` | timestamp nullable | protocol timestamps |
| `size_bytes` | int64 | non-negative |
| `priority` | enum | `low`, `normal`, `high` |
| `has_attachments`, `is_draft` | bool | projection |
| `is_seen`, `is_answered`, `is_flagged`, `is_junk` | bool | engine-authoritative projection snapshot |
| `engine_state` | string nullable | opaque bounded state/version token for reconciliation |
| `flags_synced_at` | timestamp nullable | snapshot freshness |
| `calendar_method` | enum nullable | `request`, `cancel`, `reply` |
| `calendar_event_id` | UUID nullable | verified Hub link only |
| `created_at`, `updated_at`, `deleted_at` | timestamp | projection lifecycle |

Seen/answered/flagged/junk/mailbox membership stay authoritative in the engine and are
exposed through a projection snapshot. `bcc` is never reconstructed for received mail.

### 3.5.1 MailMessageFolder

`mail_message_id`, `mail_folder_id`, `engine_state`, `projected_at`. Unique message +
folder. Both rows must belong to the same `MailAccount`. This join is rebuildable from
engine state and must not be used to authorize an engine call without resolving the
authenticated owner's account again.

### 3.6 MailLabel

`id`, `mail_account_id`, `name` (1..60, unique case-insensitive), `color`, timestamps.
Many-to-many assignment to messages is Hub-canonical. Thread label aggregation is derived.

### 3.7 MailAttachment

| Field | Type | Contract |
|---|---|---|
| `id` | UUID | primary key |
| `mail_message_id` | UUID nullable | set when saved/submitted/ingested |
| `draft_client_id` | UUID nullable | temporary owner-bound draft scope |
| `storage_object_id` | UUID nullable | SeaweedFS-backed Hub upload |
| `engine_blob_id` | string nullable | received/engine MIME blob |
| `content_id` | string nullable | inline reference |
| `filename`, `media_type`, `size_bytes` | metadata | server-normalized |
| `disposition` | enum | `attachment`, `inline` |
| `scan_status` | enum | `pending`, `clean`, `rejected`, `error` |
| `sha256` | string | required after confirm/ingest |
| timestamps | timestamp | required |

Exactly one of `storage_object_id` and `engine_blob_id` is set after attachment creation.
Only `clean` attachments may be submitted. The scanner interface is fail-closed.

### 3.7.1 MailObject

`id`, `organization_id`, `mail_account_id`, server-generated `object_key`, bounded
`bucket`, `upload_state` (`pending`, `confirmed`, `orphaned`, `deleted`), expected and
actual sizes, normalized filename/media type, nullable `sha256`, uploader user id,
`draft_client_id`, `scan_status`, `expires_at`, and timestamps. Object keys are generated
only by the server under `mail/<organization>/<account>/<object-id>`; no client-supplied
path is accepted. A dedicated mail storage port performs presign/confirm/read/delete and
rechecks owner + account on every operation. Pending/orphaned objects expire and are
cleaned idempotently. Project-scoped `FileMetadata` and project membership never grant
mail access.

### 3.8 MailRule

`id`, `mail_account_id`, `name`, `enabled`, `priority`, a bounded typed condition tree,
typed actions, `engine_script_id`, `version`, timestamps. V1 supports safe conjunctions
over sender/recipient/subject/has-attachment and actions move/add-label/mark-read. Raw
user-provided Sieve is never accepted by the public API.

### 3.9 RecentContact

`id`, `mail_account_id`, normalized `email`, `display_name`, nullable `user_id`,
`last_used_at`, `use_count`. Unique by account + normalized email. Suggestions always
rank active Hub users before external recent contacts.

### 3.10 MailSubmission

`id`, `mail_account_id`, nullable `mail_draft_id`, `idempotency_key_hash`, canonical
`request_hash`, stable `engine_submission_key`, `status` (`queued`, `submitting`, `sent`,
`unknown`, `failed`), nullable `engine_submission_id`, nullable `result_message_id`,
`last_error_code`, `attempt_count`, timestamps. Unique account + key hash. The raw key is
never persisted or returned. The persisted error is a safe code, never raw SMTP content
containing addresses/body.

Before retrying an `unknown` or interrupted submission, the service calls the engine's
submission reconciliation operation with the stable engine key. A confirmed engine result
is finalized without another submit; a definitive absence may transition back to queued;
an indeterminate result remains `unknown` and is not blindly retried. Reusing a key with a
different request hash returns `409`; identical input returns the same receipt.

The canonical send request hash covers the account, draft id, Hub revision, opaque engine
version, authorized From identity, normalized recipients, subject/body content hash,
priority and ordered attachment ids/hashes. Only the final hash is persisted. The stable
engine submission key derives from the Hub submission id, never from the raw HTTP key.

### 3.10.1 MailDraft

`id`, `mail_account_id`, opaque `engine_message_id`, opaque `engine_version`, positive
Hub `revision`, `status` (`active`, `saving`, `unknown`, `submitting`, `submitted`,
`deleted`), hashed
create-idempotency key + canonical create-request hash, nullable reply/forward message ids,
nullable last mutation id, timestamps. Unique engine message id per account and unique
account + create-key hash.
Recipients, subject, rich/plain bodies and MIME remain canonical in the engine and are
hydrated only for an authorized draft read; PostgreSQL never stores a second body copy.

Draft mutation is serialized per draft. `If-Match` compares Hub revision, then the engine
save receives the last engine version and a stable Hub mutation id. A successful engine
write stores the returned version and increments Hub revision exactly once. If the process
loses the response, state becomes `unknown`; the next read/save reconciles the engine
version before accepting another edit. It never overwrites a newer engine draft blindly.
Starting a submission atomically moves the draft to `submitting` and freezes its content;
edits are rejected until reconciliation proves definitive failure/absence and restores
`active`, or successful delivery moves it to `submitted`.

### 3.10.2 MailUserSettings

One row per active account: required default identity and nullable default signature ids
owned by the same account, `remote_images` (`blocked`, `trusted_senders`),
`compose_format` (`rich`, `plain`), version and timestamps. Provisioning/error accounts do
not expose this public resource. Defaults never bypass identity `send_allowed` checks.

### 3.10.3 MailSignature

`id`, `mail_account_id`, bounded name, sanitized rich body, server-normalized plain body,
`is_default`, version and timestamps. At most one default per account. Signatures are user
configuration rather than message bodies, but are still sanitized on write and render.

### 3.10.4 MailDomain

`id`, `organization_id`, normalized ASCII domain, `enabled`, `is_default`, verification
and DKIM status, bounded required-DNS projection, engine domain id, version and timestamps.
Domain is unique globally after normalization; at most one default per organization. DKIM
private material remains engine/secret-provider state and is never stored here.

### 3.10.5 MailboxPolicy

One versioned row per organization containing external-mail enablement, maximum message,
attachment and recipient limits, per-minute send rate, trash/junk retention, and
timestamps. The existing flat Resend alert setting remains a separate legacy resource.

### 3.11 CalendarEvent

| Field | Type | Contract |
|---|---|---|
| `id` | UUID | primary key |
| `organization_id`, `calendar_owner_id` | UUID | required, indexed |
| `uid` | string | stable, bounded, unique with owner; master-level only |
| `title`, `description`, `location` | string | bounded; description sanitized at render |
| `starts_at`, `ends_at` | timestamp nullable | required for timed events, end > start |
| `start_date`, `end_date_exclusive` | date nullable | required for all-day events |
| `all_day` | bool | selects date vs timestamp representation |
| `timezone` | IANA string | required |
| `rrule` | string nullable | RFC 5545 subset daily/weekly/monthly + until/count |
| `exdates` | JSON date/date-time array | same value type and timezone semantics as DTSTART |
| `sequence` | non-negative int | monotonic invitation revision |
| `source` | enum | `hub`, `mail_import`, `ical_feed` |
| `status` | enum | `confirmed`, `tentative`, `cancelled` |
| `organizer_id` | UUID nullable | null only for external organizer |
| `organizer_email` | normalized email | required |
| `scheduled_call_id` | UUID nullable unique | optional 1:1 link |
| `reminder_minutes` | int nullable | 0..40320 |
| `payload_hash` | SHA-256 | normalized iTIP idempotency check |
| `version` | int | optimistic event editing |
| timestamps + `deleted_at` | timestamp | soft delete/cancellation audit |

### 3.12 CalendarAttendee

| Field | Type | Contract |
|---|---|---|
| `id`, `calendar_event_id` | UUID | required |
| `user_id` | UUID nullable | null for external guest |
| `email`, `display_name` | normalized identity | email required |
| `role` | enum | `required`, `optional` |
| `rsvp_status` | enum | `needs_action`, `accepted`, `declined`, `tentative` |
| `recurrence_id` | string nullable | normalized occurrence key; null means series |
| `responded_at` | timestamp nullable | set outside `needs_action` |
| `version` | positive int | optimistic lock |
| `reply_delivery_status` | enum | `not_required`, `pending`, `sent`, `failed` |
| timestamps | timestamp | required |

Database enforcement uses two partial unique indexes because PostgreSQL does not consider
null values equal: `(event_id, normalized_email) WHERE recurrence_id IS NULL` and
`(event_id, normalized_email, recurrence_id) WHERE recurrence_id IS NOT NULL`. An
occurrence-specific attendee row inherits the series row until a response or override
materializes it.

### 3.13 CalendarEventException

`id`, `calendar_event_id`, `recurrence_id`, optional overridden start/end/date/timezone/
location/status, `sequence`, `payload_hash`, timestamps. Unique event + recurrence id.
Deleting one occurrence stores a cancelled exception rather than mutating the master rule.

### 3.14 CalendarMutationReceipt

`id`, `organization_id`, authenticated `actor_user_id`, `operation`
(`event_create`, `rsvp_series`, `rsvp_occurrence`), `idempotency_key_hash`, canonical
`request_hash`, `state` (`in_progress`, `completed`), `response_status`, bounded
`response_json`, nullable `calendar_event_id`, timestamps and `expires_at`. Unique
`(organization_id, actor_user_id, operation, idempotency_key_hash)`. The raw key is never
persisted. Reservation, domain mutation, outbox write and completed response are committed
atomically; abandoned `in_progress` receipts reconcile by transaction rollback, not by a
second mutation. Completed responses are retained for at least 24 hours and replayed byte-
equivalently. The stored response is calendar metadata only and follows the same access and
retention policy as its event.

### 3.15 CalendarInboundReplyReceipt

`id`, `calendar_attendee_id`, non-negative `event_sequence`, canonical sender-bound
`payload_hash` (SHA-256), `outcome` (`applied`, `observed_no_change`), and `created_at`.
Unique `(calendar_attendee_id, event_sequence, payload_hash)`. The service derives this
sender-bound hash as SHA-256 over the normalized authenticated/envelope sender plus the
parser's canonical payload hash (which already covers UID, optional recurrence id,
sequence and normalized REPLY semantics); raw MIME is never stored. This ledger is
distinct from the HTTP
`CalendarMutationReceipt` and contains no idempotency key or response body.

The service accepts at most eight distinct validated REPLY hashes for one attendee and
event sequence. It retains all eight until that event sequence is obsolete; a ninth new
hash is quarantined without attendee/outbox mutation. Older-sequence receipts may be
removed only after the event sequence has advanced because every lower-sequence REPLY is
then unconditionally stale. Receipt insertion, optional attendee transition and optional
outbox write share the same transaction and a lock on the sender-matched attendee
serializes concurrent distinct replies and cap checks.

## 4. Calendar state machines

### 4.1 RSVP

All transitions between `needs_action`, `accepted`, `tentative`, and `declined` are valid
for an active event. A cancelled event accepts no RSVP mutation. Each transition:

1. validates event/occurrence and current attendee version;
2. updates status, `responded_at`, version in one transaction;
3. appends `calendar.rsvp.changed.v1` to the outbox;
4. queues external `METHOD:REPLY` when organizer is external;
5. returns the canonical attendee including delivery status.

Repeated request with the same idempotency key returns the original result. Repeated
request without the same key but with stale version returns `409`, canonical
`{error,request_id}`, `X-Hub-Error-Code: ATTENDEE_VERSION_CONFLICT`, the current attendee
version as `ETag`, and a `Location` to refetch the canonical event before retry.

### 4.2 Event updates

| Change | Sequence | Existing attendee RSVP | Notification |
|---|---:|---|---|
| description only, silent | unchanged | preserve | none |
| description only, notify | +1 | preserve | REQUEST update |
| title/agenda | +1 | preserve | REQUEST update |
| time, timezone, all-day | +1 | reset non-organizer | REQUEST update |
| location | +1 | reset non-organizer | REQUEST update |
| recurrence/exdate | +1 | reset affected non-organizer scope | REQUEST update |
| add attendee | +1 | preserve existing; new needs_action | REQUEST to affected/all per transport policy |
| remove attendee | +1 | removed row retained as audit tombstone | CANCEL to removed attendee |
| cancel event/occurrence | +1 | preserve historical response | CANCEL |

Cancelling one occurrence writes or updates a `CalendarEventException` with canonical
`recurrence_id` and `cancelled` status; it never rewrites the master RRULE or cancels the
series. `DELETE /calendar/v2/events/{eventId}/instances/{recurrenceId}` compares the master
event version from `If-Match`, increments master sequence and version once, returns the
cancelled occurrence metadata, writes one `calendar.invite.cancelled.v1` outbox event and
queues one iTIP `CANCEL` carrying `RECURRENCE-ID`. A retry after an ambiguous response
refetches the event/occurrence; a stale version conflicts before any second mutation.

### 4.3 Incoming iTIP

- MIME parser enforces total size, nesting depth, line length, component count, timezone,
  recurrence expansion, and attendee count bounds.
- `REQUEST`: validate addressed mailbox and organizer; upsert by owner + UID + recurrence;
  apply only newer sequence or an exact equal-sequence duplicate.
- `CANCEL`: same idempotency rules; cancel only matching event/occurrence.
- `REPLY`: match the normalized authenticated/envelope sender to exactly one attendee and
  never accept organizer/event mutations. Its `SEQUENCE` must equal the current event
  sequence: lower is stale, higher is quarantined. The service derives the sender-bound
  hash from the normalized sender plus parser-provided canonical payload hash and looks it
  up in `CalendarInboundReplyReceipt`; a hash already processed for that attendee and
  sequence is a no-op even if later replies changed the attendee again. A distinct
  current-sequence sender-bound PARTSTAT may advance only that attendee and atomically
  records an `applied` receipt plus one outbox event. A distinct hash whose PARTSTAT already
  equals current state records `observed_no_change` without attendee/outbox mutation, so a
  later redelivery cannot roll state backward. The bounded ninth distinct hash is
  quarantined. This inbound path never creates or consumes an HTTP
  `CalendarMutationReceipt`.
- Untrusted or malformed calendar parts remain downloadable attachments with zero state
  mutation and a safe parse error code.

## 5. HTTP API contract

The machine-readable contract is `docs/mail/openapi.yaml`. General rules:

- Prefix: `/api`, matching the existing Fiber router and generated Swagger base path.
- Auth: existing bearer/session middleware.
- Pagination: opaque `cursor`, default 50, max 100.
- Mutation idempotency: `Idempotency-Key` required for send, event create, and RSVP.
- Calendar create/RSVP idempotency uses `CalendarMutationReceipt`; `MailSubmission`
  remains the send-specific receipt.
- Optimistic locking: `If-Match: "<version>"` for draft/event/rule updates; RSVP body
  includes `attendee_version` for explicit conflict responses.
- Errors retain mandatory Hub NFR-13 wire JSON `{error,request_id}`. Machine-readable
  calendar conflict codes use `X-Hub-Error-Code`; stale RSVP additionally returns current
  attendee-version `ETag` and canonical event `Location` for refetch.
- `403` does not reveal whether an engine/message id exists in another mailbox.

Endpoint groups frozen for v1:

| Group | Endpoints |
|---|---|
| Folders | `GET/POST /mail/folders`, `PATCH/DELETE /mail/folders/{id}` |
| Labels | `GET/POST /mail/labels`, `PATCH/DELETE /mail/labels/{id}`, `PUT /mail/messages/{id}/labels` |
| Threads/messages | `GET /mail/threads`, `GET /mail/threads/{id}`, `GET /mail/messages/{id}`, `POST /mail/messages/bulk` |
| Draft/send | `POST /mail/drafts`, `GET/PATCH/DELETE /mail/drafts/{id}`, `POST /mail/drafts/{id}/send`, `GET /mail/submissions/{id}` |
| Attachments | `POST /mail/attachments/presign`, `POST /mail/attachments/{id}/confirm`, `GET/DELETE /mail/attachments/{id}` |
| Search/contacts | `GET /mail/search`, `GET /mail/contacts/suggest` |
| User settings | `GET/PATCH /mail/settings`, `GET /mail/identities`, CRUD `/mail/signatures`, CRUD `/mail/rules` |
| Existing calendar aggregate | `GET /calendar/events` remains byte-shape compatible as `{events,from,to}`; first-class events are later added only as an additive aggregate kind |
| Canonical calendar | `GET/POST /calendar/v2/events`, `GET/PATCH/DELETE /calendar/v2/events/{id}`, `DELETE /calendar/v2/events/{id}/instances/{recurrenceId}`, v2 RSVP series/occurrence and v2 availability |
| Existing Resend admin | `GET/PUT /admin/mail/settings` retains the flat `hub_mail_url/from_alerts/alerts_enabled/daily_limit/resend_ready` response contract; deprecated `POST /admin/mail/broadcast` is documented as a current ADR-NOTIF-01 governance gap, not new scope |
| Mailbox admin | `GET/PUT /admin/mail/mailbox-settings`, domain CRUD, `GET /admin/mail/health`; all require admin RBAC and an audit action |

## 6. Event envelope and compatibility

Canonical JSON envelope:

```json
{
  "schema_version": 1,
  "event_id": "uuid",
  "occurred_at": "2026-08-13T12:00:00Z",
  "organization_id": "uuid",
  "actor_id": "uuid-or-null",
  "correlation_id": "opaque-bounded-string",
  "payload": {}
}
```

- Subject suffix and `schema_version` advance together for breaking changes.
- Additive optional fields do not require a new subject version.
- Consumers reject unknown major versions to dead letter; they never guess.
- PII-bearing payloads contain normalized addresses but no body, raw MIME, attachment
  content, auth tokens, or engine credentials.
- Retention follows existing event/audit policies; search consumers store only projections.
- The shared JetStream `EVENTS` subject configuration must include `mail.>` and
  `calendar.>` before any producer is enabled. The outbound webhook consumer allowlist is
  unchanged unless a separate privacy/security decision explicitly adds a safe event.

### 6.1 Payload schemas

`mail.received.v1`:

```json
{
  "mailbox_id": "uuid",
  "message_id": "uuid",
  "engine_message_id": "opaque",
  "thread_id": "uuid",
  "from": {"email": "sender@example.test", "name": "Sender"},
  "to": [{"email": "user@hub.example"}],
  "received_at": "2026-08-13T12:00:00Z",
  "folder_ids": ["uuid"],
  "has_attachments": true,
  "calendar_method": "request"
}
```

`mail.sent.v1` adds `submission_id`, `sent_at`, and `delivery_class` (`internal`,
`instance_domain`, `external`). `mail.message.changed.v1` contains a non-empty bounded
`changed_fields` list plus current folder ids/flags/label ids.

Calendar invite events contain `event_id`, `uid`, `sequence`, optional `recurrence_id`,
attendee ids, and changed fields. `calendar.rsvp.changed.v1` contains attendee id/email,
previous/new status, new version, recurrence id, and reply delivery status.

## 7. Security acceptance contract

| Control | Required behavior | Failure mode |
|---|---|---|
| Owner isolation | all engine calls bind server-resolved account id | `404`/`403`, no engine call |
| HTML sanitization | server allowlist removes active content and fetch primitives | plain-text fallback |
| Reader isolation | CSP/sandbox; no direct remote network | content remains blocked |
| Image proxy SSRF | DNS/IP validation before and after connect; private ranges blocked; size/time/type caps | `REMOTE_CONTENT_BLOCKED` |
| Attachments | limits, filename normalization, type policy, hash, fail-closed scanner | quarantine/reject |
| Sending | per-user/instance token buckets, recipient/message caps | `429` with bounded retry hint |
| Auto-response | `Auto-Submitted`, precedence, null sender and loop headers honored | no automated response |
| iTIP reply | sender-attendee match, UID/sequence/version rules | attachment-only/quarantine |
| Admin | existing admin RBAC + audit | `403` |
| Search | owner/org filters before hydrate | empty/not found, audit anomaly |
| Secrets | engine/DKIM credentials never in repository/log/API | startup failure if absent |

Remote image retrieval uses only server-issued tokens bound to an owned message via
`GET /mail/messages/{messageId}/remote-images/{imageToken}`. The public API never accepts
an arbitrary target URL for proxying.

## 8. Requirements and traceability

Priority: P0 blocks merge-ready feature claim; P1 is required for full requested scope;
P2 is explicitly deferred. Risk uses Critical/High/Medium.

| ID | Requirement | Priority | Risk | Unit evidence owner | Higher-layer handoff |
|---|---|---:|---|---|---|
| FR-MAIL-001 | owned mailbox provisioning and isolation | P0 | Critical | Go adapter/service fakes | contract/API/authz |
| FR-MAIL-002 | system + depth-2 custom folders | P1 | High | Go folder invariants | API/E2E |
| FR-MAIL-003 | thread grouping from references + bounded fallback | P0 | High | Go pure table tests | API/E2E |
| FR-MAIL-004 | message/thread read, flag, pin and realtime unread | P1 | High | Go projection + FE reducer | API/E2E/realtime |
| FR-MAIL-005 | bulk move/label/archive/junk/trash | P1 | High | Go command tests | API/E2E/audit |
| FR-MAIL-006 | drafts with revisioned autosave and recovery | P0 | High | Go service + FE hook | API/E2E |
| FR-MAIL-007 | reply/reply-all/forward RFC headers and quoting | P0 | High | Go pure tests | contract/E2E |
| FR-MAIL-008 | SeaweedFS multi/inline attachments with scan gate | P0 | Critical | Go policy + FE state | API/integration/security |
| FR-MAIL-009 | internal and instance-domain external send/receive through engine | P0 | Critical | adapter mock | integration/E2E/runtime |
| FR-MAIL-010 | labels and nested folders | P1 | Medium | Go invariants + FE | API/E2E |
| FR-MAIL-011 | Meilisearch async search/filter projection | P1 | High | event mapping | integration/API |
| FR-MAIL-012 | Hub user + recent contact autocomplete | P1 | Medium | ranking pure tests | API/E2E |
| FR-MAIL-013 | signature, From alias, CC/BCC, priority | P1 | High | composition tests | API/E2E |
| FR-MAIL-014 | user rules compile to safe typed Sieve | P1 | High | compiler tests | integration/security |
| FR-MAIL-015 | admin domains/limits/retention/health/external toggle | P0 | Critical | policy/service tests | API/authz/runtime |
| FR-MAIL-016 | HTML/remote content/attachment/rate/loop security controls | P0 | Critical | pure policy tests | security/integration |
| FR-CALSYNC-001 | first-class event/attendee model and call backfill | P0 | Critical | state/policy units | migration/integration |
| FR-CALSYNC-002 | deterministic REQUEST create/update | P0 | Critical | RFC fixture tests | contract/E2E |
| FR-CALSYNC-003 | deterministic CANCEL event/occurrence | P0 | Critical | RFC fixture tests | contract/E2E |
| FR-CALSYNC-004 | incoming REPLY updates only matching attendee | P0 | Critical | spoof/state tests | API/integration/security |
| FR-CALSYNC-005 | incoming external REQUEST import + outward REPLY | P0 | Critical | parser/state tests | integration/E2E |
| FR-CALSYNC-006 | inline RSVP with optimistic lock and realtime organizer view | P0 | Critical | Go state + FE card | API/E2E/realtime |
| FR-CALSYNC-007 | reschedule sequence/reset policy | P0 | High | policy table tests | contract/E2E |
| FR-CALSYNC-008 | duplicate UID/sequence idempotency | P0 | Critical | parser/service tests | integration |
| FR-CALSYNC-009 | series vs RECURRENCE-ID response | P0 | Critical | recurrence fixtures | integration/E2E |
| FR-CALSYNC-010 | IANA TZ, DST and all-day correctness | P0 | Critical | fixed-zone fixtures | integration/E2E |
| FR-CALSYNC-011 | busy/free includes calls/events/leave | P1 | High | interval pure tests | API/E2E |
| FR-CALSYNC-012 | reminders through existing notification preferences | P1 | High | event mapping | integration/E2E |
| FR-CALSYNC-013 | L1 external iCal feed subscription | P2 | High | parser fixtures | integration/security |

### Named critical scenarios

- `TC-CALSYNC-001`: accept whole recurring series.
- `TC-CALSYNC-002`: decline one occurrence using `RECURRENCE-ID`.
- `TC-CALSYNC-003`: time change increments sequence and resets non-organizer RSVP.
- `TC-CALSYNC-004`: silent description edit preserves sequence/RSVP.
- `TC-CALSYNC-005`: availability reports call/event conflict and approved leave.
- `TC-CALSYNC-006`: concurrent RSVP with stale attendee version returns 409.
- `TC-CALSYNC-007`: repeated same UID/sequence/payload is idempotent.
- `TC-CALSYNC-008`: same UID/sequence with conflicting payload is quarantined.
- `TC-CALSYNC-009`: external guest receives request and reply changes only that guest.
- `TC-CALSYNC-010`: organizer/attendee zones cross a DST boundary without instant drift.
- `TC-CALSYNC-011`: all-day event round-trips exclusive end date.
- `TC-CALSYNC-012`: spoofed REPLY cannot change another attendee or organizer.
- `TC-CALSYNC-013`: cancelling one recurrence instance creates a cancelled exception and
  occurrence-scoped iTIP/outbox without changing or cancelling the remaining series.
- `TC-CALSYNC-014`: for series and occurrence, accepted REPLY hashes A then B followed by
  delayed A cannot roll RSVP backward; same-status hashes are durably observed and a ninth
  distinct current-sequence hash quarantines without attendee/outbox mutation.
- `TC-MAIL-001`: retry after send timeout reconciles idempotency key without duplicate.
- `TC-MAIL-002`: mailbox-owner isolation prevents cross-user engine access.
- `TC-MAIL-003`: hostile HTML and SVG active content are removed server-side.
- `TC-MAIL-004`: image proxy blocks loopback/private/rebound targets.
- `TC-MAIL-005`: pending/rejected attachment cannot be sent.
- `TC-MAIL-006`: nested folder cycle and depth 3 are rejected.
- `TC-MAIL-007`: References grouping wins over misleading normalized subject.
- `TC-MAIL-008`: auto-response loop headers suppress generated reply.
- `TC-MAIL-009`: lost draft-save response reconciles before another write and PostgreSQL
  contains no body copy.
- `TC-MAIL-010`: mail attachment object authorization is account-scoped and cannot be
  obtained through project membership.
- `TC-MAIL-011`: submission key reuse with changed frozen draft input returns 409; an
  unknown receipt never causes blind submit.
- `TC-MAIL-012`: calendar/mail outbox subjects are admitted durably while the generic
  outbound webhook filter remains unchanged.
- `TC-MAIL-013`: draft content is frozen while its submission is queued/submitting/unknown
  and becomes editable again only after a definitive non-delivery result.

## 9. Delivery phases

| Phase | Backlog | Contract outcome |
|---:|---|---|
| 142 | TASK-235 | ADR, ERD, API/events/security/QA handoff frozen |
| 143 | TASK-236 | CalendarEvent/Attendee/Exception, iTIP and RSVP core |
| 144 | TASK-237 | Mailbox domain, fake engine, folders/threads/drafts/submission |
| 145 | TASK-238 | Stalwart adapter, instance domain, Docker profile and admin health |
| 146 | TASK-239 | Outlook three-pane UI, composer, attachments, search/realtime |
| 147 | TASK-240 | recurrence/import reliability, busy/free, reminders, L1/CRM seam |
| 148 | TASK-241 | security, reconciliation, docs, independent verification/rebase gate |

## 10. Status honesty

Freezing this document authorizes implementation but proves no runtime behavior. Until
the corresponding evidence exists, all implementation, migration, Docker, API, UI,
security, and end-to-end statuses remain `NOT_RUN` or `NOT_IMPLEMENTED`.
