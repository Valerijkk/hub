# Phase 147 — reliability and integration contract

> **Contract addendum:** Mail 0.1.6
> **Status:** ROOT_STATIC_REVIEW_0_1_6_PENDING · 0_1_5_SNAPSHOT_STALE · INDEPENDENT_REVIEW_PENDING · PRODUCT_CODE_NOT_IMPLEMENTED
> **Depends on:** accepted Phase 143 calendar core, Phase 144 mailbox domain, and the
> engine adapter portions named by each packet
> **Runtime evidence:** NOT_RUN

This addendum closes the behavior that was intentionally left after the mailbox and RSVP
core. It does not authorize production/test code under the wrong model, an unreviewed
migration, direct Docker/Compose execution, network access, or L1 feed implementation
before its founder gate.

## 1. Non-negotiable invariants

1. A received calendar part is never state-changing merely because it parses. The engine
   must supply a trusted internal-submission result or an externally authenticated and
   organizer-aligned result.
2. Incoming REQUEST/CANCEL order is decided per calendar owner, UID, and optional
   occurrence key. Lower sequence never rolls state backward; an equal sequence with a
   different method or canonical payload hash is quarantined.
3. A CANCEL received before its REQUEST creates a safe tombstone. A delayed older REQUEST
   cannot resurrect it.
4. Durable consumer receipts contain safe identifiers, hashes, outcome and retry metadata;
   they never persist raw MIME, message/calendar body, a feed URL, credentials, or an
   authentication header.
5. An outward REPLY is a mailbox submission with stable draft-create and submission keys.
   An indeterminate send is reconciled before any retry and can never become a blind second
   submit.
6. Availability authorizes all requested users before querying any source and returns no
   title, event id, leave reason, project, participant, or other private detail.
7. Reminder scheduling is durably idempotent. The in-app row is deduplicated persistently;
   Telegram/push retain their existing best-effort boundary and are never described as
   exactly-once. Mail/Resend is not added as a notification channel.
8. CRM compose creates one ordinary idempotent Mail draft and uses the owned mailbox
   composer/send receipt; recipient data never travels in a URL. Manual CRM logging links
   an already authorized message projection; CRM never stores message body, raw MIME,
   recipients, tracking state, or a second subject copy.
9. L1 iCal feed subscription is contract-ready but code-gated by founder question 5. L2
   external mailbox ingest and L3 OAuth remain out of scope.
10. Every new producer stores the complete shared v1 envelope through the unchanged outbox
    with `event_type == envelope.subject`. New consumers validate broker/JSON subject
    equality and deduplicate on envelope application `event_id`; the separate outbox-row
    `Hub-Event-Id` is transport diagnostics only. Legacy outbox payloads are never globally
    wrapped or reinterpreted.

## 2. Incoming REQUEST/CANCEL persistence

### 2.1 CalendarInboundObjectState

One current ordering record exists for `(calendar_owner_id, uid, recurrence_id)`:

| Field | Contract |
|---|---|
| `id`, `organization_id`, `calendar_owner_id` | required UUIDs; owner is server-resolved |
| `uid` | normalized bounded iCalendar UID |
| `recurrence_id` | nullable normalized occurrence key; null means series |
| `organizer_email` | normalized authenticated organizer, max 320 |
| `sequence` | highest accepted non-negative sequence for this object |
| `method` | `request` or `cancel` |
| `payload_hash` | exact parser-provided canonical SHA-256 |
| `state` | `applied`, `tombstone`, or `deferred_missing_master` |
| `calendar_event_id` | nullable; absent is valid for an early cancel/deferred occurrence |
| `expires_at` | nullable; required only for `deferred_missing_master` |
| timestamps | required |

PostgreSQL uses two partial unique indexes because nullable recurrence values are not equal:
owner + UID where recurrence is null, and owner + UID + recurrence where non-null. The row
is locked while comparing and advancing sequence. It is not an HTTP idempotency receipt.
Known-event state follows calendar retention. An unbound CANCEL tombstone is retained for
the mailbox/account lifetime so an arbitrarily delayed older REQUEST cannot resurrect it;
it is removed only with the owning mailbox/account or after a newer admitted object binds
the state to an event. `deferred_missing_master` expires after seven days together with its
dead-lettered message receipt. A mailbox may hold at most 10,000 unbound object states; at
the cap every state-changing part for a previously unseen UID quarantines before event
mutation, while already-known UID state continues to converge. The mailbox calendar-import
health becomes explicitly degraded and emits one rate-limited, content-free operational
signal/audit summary. Cleanup never evicts a retained tombstone merely to admit another
object. Recovery requires an approved resource/quota increase or removal of the owning
mailbox/account; there is no automatic tombstone deletion/reset button.

### 2.2 CalendarInboundMessageReceipt

One receipt exists for a projected message plus hashed stable calendar-part locator:

| Field | Contract |
|---|---|
| `mail_message_id`, `part_locator_hash` | unique together; locator is SHA-256, not engine part id |
| `event_envelope_id` | original `mail.received.v1` event id for traceability |
| `state` | `pending`, `processing`, `applied`, `ignored_stale`, `duplicate`, `deferred_missing_master`, `quarantined`, `dead_letter` |
| `attempts`, `lease_until`, `next_attempt_at` | bounded claim/recovery metadata |
| `error_code` | nullable stable allowlisted code, no provider/body text |
| `calendar_event_id` | nullable resolved event |
| timestamps | required |

The worker claims with a lease. A transient engine/DB failure schedules bounded retry;
permanent malformed, untrusted, equal-conflict, unsupported-version, or exhausted input is
finalized without mutation. `deferred_missing_master` is rehydrated by the recovery job;
the receipt never stores the calendar bytes needed for that retry.
Deferred receipts stop retrying after seven days and become `dead_letter`; finalized
message receipts are retained for 30 days, then removed by a bounded cleanup job.

### 2.3 Authentication admission

The engine-facing message DTO exposes only a bounded calendar part plus an enum and safe
authentication facts:

- `trusted_internal`: engine confirms an authenticated local identity whose normalized
  From equals ORGANIZER;
- `dmarc_aligned`: external DMARC passes and RFC5322.From equals ORGANIZER;
- `untrusted`: missing/failing/unaligned results, forwarded calendar attachment, or an
  unverified delegation.

Only the first two classes may mutate. v1 has no organizer-delegation allowlist. Envelope
sender, header From and ORGANIZER are normalized before comparison; a null/empty sender is
untrusted. Once an object state exists, its organizer identity is immutable: even a newer
otherwise authenticated REQUEST/CANCEL from a different organizer is quarantined. This
prevents a sender who learns another organizer's UID from taking over the event.
Authentication details are not included in public API/event payloads.

### 2.4 REQUEST/CANCEL decision table

| Current object state | Incoming | Result |
|---|---|---|
| none | trusted REQUEST | create/import series or apply occurrence against an active master; otherwise defer occurrence; advance state |
| none | trusted CANCEL | create tombstone; no guessed event fields |
| any | lower sequence | `ignored_stale`; no event/outbox mutation |
| any | equal sequence + same method/hash | `duplicate`; no mutation |
| any | equal sequence + different method/hash | `quarantined`; no mutation |
| any | newer REQUEST | create/update/reactivate matching object, or keep an occurrence deferred while its master is absent/cancelled; advance state |
| any | newer CANCEL | cancel master/occurrence or advance tombstone |

The last five rows first require exact organizer match with the stored state; organizer
mismatch always quarantines regardless of sequence.

[RFC 5546 sequencing](https://www.rfc-editor.org/rfc/rfc5546.html#section-2.1.5)
orders revisions only for the same UID plus optional RECURRENCE-ID; Hub does not compare a
series sequence and an occurrence sequence as if they were the same component. It does lock
the series state before an occurrence state and adds these safety rules:

- an occurrence must match the immutable organizer of an existing series even when no
  occurrence row exists; a pre-master occurrence records its candidate organizer and only a
  matching master may wake it;
- an occurrence REQUEST cannot create or reactivate an event while the master is absent or
  cancelled; it remains deferred until a same-organizer master REQUEST is active;
- a series CANCEL cancels the master and all materialized exceptions without deleting their
  ordering high-watermarks; a later occurrence-only message cannot resurrect the series;
- a master REQUEST wakes only same-owner/UID/organizer deferred occurrences. Each source is
  rehydrated and rechecked against both current series admission and its own object state.

Within an occurrence's own key, an early CANCEL remains effective until a strictly newer
authenticated occurrence REQUEST is eligible under the active-master rule. Event/exception
mutation, object state, finalized message receipt and any calendar outbox event commit
atomically.

## 3. External-organizer REPLY delivery

### 3.1 CalendarReplyDelivery

Every RSVP that requires external transport creates one delivery in the same transaction:

| Field | Contract |
|---|---|
| `id`, `calendar_event_id`, `calendar_attendee_id` | required UUIDs |
| `event_sequence`, `attendee_version` | exact RSVP revision being transported |
| `recurrence_id` | nullable occurrence key copied for immutable delivery intent |
| `rsvp_status` | `accepted`, `declined`, or `tentative` |
| `state` | `pending`, `submitting`, `unknown`, `sent`, `failed_permanent` |
| `mail_submission_id` | nullable link after mailbox reservation |
| `attempts`, `next_attempt_at`, `last_error_code` | bounded safe recovery metadata |
| timestamps | required |

Unique `(calendar_attendee_id, event_sequence, attendee_version)` makes one immutable
delivery intent. The attendee's public `reply_delivery_status` reflects its latest delivery
without erasing older audit rows: no delivery intent is `not_required`; `pending`,
`submitting`, and `unknown` map to `pending`; `sent` maps to `sent`; and
`failed_permanent` maps to `failed`.

The sender adapter creates/reconciles a hidden engine draft using a stable key derived from
the delivery UUID, then creates/reconciles `MailSubmission` using another stable key. It
uses the attendee owner's active account and allowed identity; a caller never supplies an
account, From identity, organizer address, or engine id. `unknown` always calls mailbox
reconciliation. Definitive policy/identity rejection becomes `failed_permanent`; transient
unavailability remains bounded pending retry. The iTIP serializer output is request-lifetime
data only.

## 4. Availability

`POST /api/calendar/v2/availability` uses a half-open `[from,to)` window, at most 31 days,
1–50 unique active users eligible under the existing collaboration scope, a valid IANA
response timezone, and `slot_minutes` between 5 and 240.

- Authorization is all-or-nothing before source queries. An eligible target is the caller
  or an active same-organization user sharing at least one non-deleted project or active
  channel with the caller, matching the existing HR/calendar and scheduled-call privacy
  boundary. An unrelated, inactive or foreign user produces the same generic not-found
  response and no partial result. Org-wide admin/founder listing does not bypass this
  availability rule.
- Busy event policy: organizer plus internal attendees are busy for `accepted`,
  `tentative`, and `needs_action`; `declined`, cancelled and deleted instances are absent.
- Canonical events include linked-call meetings once as source `event`; source `call` is
  restricted to active unlinked calls while the pre-cutover transition flag is enabled.
  Linked call projection rows are not loaded as a second busy interval. After bridge
  activation, an unlinked non-deleted call is `AVAILABILITY_SOURCE_INVALID` with no partial
  response. Pre-cutover unlinked active calls include creator and participants.
  Approved, non-deleted leave is
  loaded only through the existing viewer/shared-project scope. Its stored inclusive UTC
  date labels are interpreted as whole local dates in the target employee's stored IANA
  timezone, converted to an exclusive end, then to instants/response timezone. An invalid
  legacy employee timezone fails the entire response with `422` +
  `AVAILABILITY_SOURCE_INVALID`; UTC is not silently substituted.
- Timed events retain instants. All-day and recurrence expansion use the event timezone;
  DST gaps/overlaps are resolved by the frozen calendar policy, never by fixed offsets.
- Expansion/output is bounded to the request range, 1,000 busy intervals per user and
  5,000 total across the response.
  Limit exhaustion is `422` + `X-Hub-Error-Code: AVAILABILITY_LIMIT_EXCEEDED`, not a
  truncated success.
- `busy` entries expose only `{from,to,source}` where source is `event`, `call`, or `leave`.
  Same-source overlapping/touching entries are merged deterministically; different-source
  entries may overlap so attribution remains truthful.
- `free` is the complement of the union of all busy intervals, clipped to the requested
  range. Gaps shorter than `slot_minutes` are omitted. User order matches request order.

No authorization decision is made from the legacy aggregate or a client-provided project.

## 5. Reminder scheduling and notification delivery

### 5.1 CalendarReminderReceipt

One receipt is unique on event, internal user, normalized occurrence key, event sequence,
and derived `due_at`. Timed occurrences derive `due_at` from their effective start instant.
For an all-day occurrence, the scheduling contract resolves the effective `start_date` at
the event-local civil-day boundary and subtracts `reminder_minutes`. That boundary is the
first representable instant of the date in the event's valid IANA timezone: a midnight
overlap chooses the earlier instant, a midnight gap chooses the first instant after the
gap, and a wholly skipped civil date chooses the first instant after that date's gap. This
scheduling-only instant never replaces the date-only event value in persistence, API or
iTIP. It stores `reminder_minutes`, state (`pending`, `emitted`,
`suppressed`, `expired`), nullable outbox id, nullable in-app outcome (`pending`,
`delivered`, `suppressed`), nullable safe outcome reason (`policy_disabled`, `quiet_hours`,
`stale_event`, `ineligible_user`, `configuration_changed`, `policy_invalid`, `expired`),
nullable notification id, and timestamps. A suppressed terminal outcome requires one
reason; pending/delivered require none. It stores no title, description, location or
recipient address. A bounded lookahead creates/finds `pending`;
only at/after `due_at` does a claim recheck current state and atomically become `emitted`
with one outbox row, `suppressed`, or, when more than 24 hours late, `expired`. The claim
must prove that the current event/exception start, event sequence and non-null
`reminder_minutes` still derive exactly the receipt's sequence, offset and `due_at`, and
that the user is still the owner or an eligible internal attendee. A disabled or changed
offset, moved occurrence, removal, decline, cancellation, deletion or superseding sequence
suppresses the old receipt. A changed offset creates a separate receipt only for its newly
derived due time. A transient pre-due RSVP change does not finalize the receipt; due-time
state is authoritative. Final receipts are retained for 90 days after `due_at`; cleanup is
bounded and never removes pending work.

State coherence is mandatory:

| Receipt state | Outbox | In-app outcome | Notification | Reason |
|---|---|---|---|---|
| `pending` | null | `pending` | null | null |
| `emitted` before consumer | required | `pending` | null | null |
| `emitted` after consumer delivery | required | `delivered` | required | null |
| `emitted` after consumer suppression | required | `suppressed` | null | required |
| scheduler `suppressed` | null | `suppressed` | null | required |
| scheduler `expired` | null | `suppressed` | null | `expired` |

`configuration_changed` covers offset/due mismatch; `stale_event` covers cancellation,
deletion, superseding sequence or moved occurrence; `ineligible_user` covers removed/
declined/inactive recipients. `policy_disabled`, `quiet_hours`, `policy_invalid` and
`expired` retain their literal meanings. No terminal state returns to pending.

The scheduler uses an injected clock, bounded lookahead and row claims. It selects event
owners and internal attendees, expands recurrence within the scan window, and at due time
rechecks current event/exception/attendee/configuration state before the terminal
transition. Reminder-relevant mutations and both reminder stages use one lock order: event,
optional linked ScheduledCall, optional exception, optional attendee, stable user-policy
guard, reminder receipt, then optional notification. The guard is the user row:
notification-preference upsert locks
it before resolving/upserting optional preference rows, a quiet-hours profile update locks
it before mutation, and the reminder consumer resolves both only after acquiring it. This
also serializes an absent default preference with a concurrent first override. Event
reschedule/cancel/delete, RSVP/remove, reminder configuration and notification-policy
changes therefore have one observable order. Cancelled, superseded, declined,
configuration-mismatched, already
emitted, or more-than-24-hours-overdue reminders are suppressed/expired without delivery.

The event payload contains only `reminder_receipt_id`, `event_id`, `user_id`, optional
`recurrence_id`, `sequence`, and `due_at`. The notification consumer refetches the owned
event, resolves existing `calendar_reminder` channel preferences and quiet hours, and uses
the receipt/event id as a persistent notification dedupe key. In one PostgreSQL transaction
and the same lock order, it locks the current event/optional linked call/exception/attendee,
user-policy guard and a still-pending in-app outcome, then performs the last-chance
authorization, policy and state check. Current sequence, effective start, non-null
`reminder_minutes`, derived
`due_at`, owner/attendee eligibility and active event/occurrence must still match the
receipt and event payload before a notification row may be created. A definitive mismatch
records one durable suppressed in-app outcome and acknowledges the event; policy
permission creates/finds the unique notification row and stores delivered. A duplicate
event observes that terminal outcome instead of re-evaluating policy. External channel
attempts remain outside this
exactly-once boundary and begin only after the canonical transaction commits. A rollback
produces no Telegram/Web Push attempt; a later external result never changes or rolls back
the receipt/notification outcome.

A transient failure while reading the user, event, preferences, quiet-hours policy, or
notification repository is not a policy decision: the consumer leaves the in-app outcome
`pending`, does not create a suppression, and NAKs/retries through the existing bounded
delivery policy. Only a definitive current authorization/configuration/policy result may
become `delivered` or `suppressed`; a definitively stale, deleted or ineligible event is
durably suppressed without a notification row. Realtime WebSocket publication happens
only after the durable transaction and remains best-effort. A missed socket publish is
repaired by the ordinary notification list/query refetch; no exactly-once realtime claim
is made.

For the keyed calendar path, quiet hours are disabled only when both fields are absent.
When present, both values must be strict `HH:MM` and the user timezone must load as a valid
IANA name; there is no UTC or "not quiet" fallback. An incomplete/invalid pair or timezone
is a definitive `policy_invalid` suppression with no notification row plus a rate-limited,
content-free metric/ops signal. It is not retried as a transient repository failure.

The transaction boundary defines concurrent user intent: when a reminder- or policy-
relevant mutation commits first, the old receipt cannot create an in-app row; when the
reminder transaction commits first, that row represents the previously valid configuration
and is not recalled. This ordering guarantee covers the canonical in-app outcome only;
Telegram/Web Push attempts after commit may still race a later channel preference change
or be missed/duplicated at a crash boundary and retain their explicit best-effort semantics.

Guarantees are deliberately narrow:

- one reminder intent per eligible due occurrence; when current notification policy permits
  in-app delivery, exactly one in-app row per receipt; when it suppresses delivery, zero;
- duplicate JetStream delivery is harmless for both the one-row and suppressed outcomes;
- Telegram and Web Push use existing best-effort senders and may be duplicated or missed
  at a crash boundary; their outcomes must not be claimed exactly-once;
- no Resend/email channel, template, broadcast, digest, or marketing behavior is added.

## 6. Conditional L1 iCal feed contract

This section is design-complete but `BLOCKED_BY_FOUNDER_QUESTION_5`. Its endpoints are not
part of OpenAPI 0.1.6 until accepted.

- CRUD proposal: `/api/calendar/v2/feeds` and `/feeds/{feedId}` plus explicit
  `/feeds/{feedId}/refresh`; every resource is owner-scoped, with at most 10 active feeds,
  URL length 2,048 and display name length 120.
- The full URL is encrypted with existing AES-256-GCM using organization/owner/feed id as
  AAD and is write-only. Public responses expose only normalized display host/path mask.
- Only HTTPS on port 443 is accepted; userinfo, fragments, non-public resolved addresses,
  link-local, loopback, private/metadata targets and DNS rebinding are blocked. Redirects
  are at most three and every hop is revalidated by the safe dialer.
- Total timeout is 10 seconds; compressed response is at most 5 MiB, decompressed content
  at most 10 MiB and compression ratio at most 10. A snapshot has at most 5,000 VEVENTs,
  500 VTIMEZONEs and 500 attendees per event; the Phase 143 line/depth/recurrence bounds
  apply before allocation or expansion.
- Conditional GET uses bounded ETag/Last-Modified. A 304 advances freshness without
  rewriting events. Credentials/tokens/query values never appear in log, audit or error.
- Source rows are unique by subscription + UID. Refresh applies one complete parsed
  snapshot transactionally; failure retains the prior snapshot and marks stale/error.
- Disable stops fetch and keeps the last snapshot visibly stale. Delete soft-deletes the
  subscription and its source-owned event projections. No feed can overwrite a Hub/mail-
  import event or another owner's feed.
- Scheduled refresh is hourly with jitter; manual refresh is rate-limited to one per five
  minutes. Consecutive failures use bounded exponential backoff capped at 24 hours.

## 7. CRM compose and manual mail activity

The CRM seam follows the locked GZ-V4-08/09/10 decision without creating campaigns.

1. `GET /api/crm/deals/{dealId}/mail-compose-context` authorizes the deal owner, loads its
   linked contact, and returns only the normalized recipient plus deal/contact identifiers.
   Missing contact/email is `409` + `CRM_MAIL_RECIPIENT_REQUIRED`. It does not send.
2. One `Write email` intent first loads the ordinary Mail settings/default send-enabled
   identity and creates a normal saved draft with `POST /api/mail/drafts`. The CRM recipient
   is present only in that request body. The client uses one stable idempotency key for the
   click intent. A lost create response is retried only with the same key and byte-equivalent
   body; the client cannot start a new-key draft while that intent is unknown. After `201`
   it navigates to `/mail/drafts/{draftId}`; a navigation failure reopens that returned UUID
   rather than creating again. URL, query, browser history, query keys, logs and errors
   contain no recipient, subject or body. The route contains only the opaque draft UUID and
   refetches it through the owner-scoped Mail API.
3. The saved-draft route opens the same ordinary Mail composer used from the mailbox. A
   local navigation-state hint may contain only `deal_id`; losing that hint on reload is
   safe and merely removes the post-send linking prompt. CRM has no alternate draft/send
   client, and the normal mailbox API remains the only send path enforcing account,
   identity, attachment, rate and submission rules.
4. After the ordinary submission has definitively resolved with an owned `message_id`, the
   UI may offer an explicit `Link to deal` action. `queued`, `submitting`, `unknown` or
   failed delivery never appears linked and never triggers an automatic PUT. After an owned
   message exists, `PUT
   /api/crm/deals/{dealId}/mail-activities/{messageId}` idempotently links it. The service
   authorizes the deal and independently proves the message belongs to the same actor's
   active mailbox. A system admin is not granted another user's mailbox access by CRM.
5. `GET /api/crm/deals/{dealId}/mail-activities` returns a bounded newest-first timeline;
   DELETE on the keyed link performs an audited unlink. Cross-user/cross-org ids are
   indistinguishable from not found and cause zero engine hydration.
6. `Open mail` uses `/mail/message/{messageId}`. That route independently authorizes and
   loads the ordinary Mail reader; the CRM activity never acts as an access token. Manual
   linking from an owned message uses the same explicit PUT and never auto-logs mail. The
   reader may offer only deals whose `owner_id` matches the authenticated user, derived from
   the already authorized CRM pipeline; that client filter is usability/privacy defense,
   never authorization, so the PUT still independently validates both owners. Missing user
   state, pipeline failure or an empty owned-deal set exposes no link mutation.

`CrmMailActivity` is immutable link metadata: id, deal id, mail message id, linked-by user,
direction (`inbound`/`outbound`), message occurred-at, created-at and nullable deleted-at.
Unique
`(deal_id,mail_message_id)`. It stores no content, subject, address, engine id, tracking,
open/click state or campaign relation. GDPR deletion removes/tombstones the link according
to CRM policy without deleting the canonical mailbox message. Relinking restores the same
owned link rather than creating duplicate history.
Each deal may hold at most 10,000 non-deleted mail links; reaching the cap returns
`409` + `CRM_MAIL_ACTIVITY_LIMIT` and never evicts history silently. Public lists use an opaque
newest-first cursor and return at most 100.
Deal deletion follows the existing CRM retention policy and tombstones its links. Mailbox
projection hard-delete first tombstones matching CRM links; no FK cascade may delete a deal
or retain a navigable link to removed personal mail.

## 8. Required static/runtime evidence

Feature units later live only in `backend/tests/calendar/**`, `backend/tests/mail/**`, and
`frontend/tests/mail/**`. Higher-layer scenarios are handed to the protected Python QA
program without editing it. Required named additions:

- `TC-CALSYNC-015`: trusted CANCEL before REQUEST blocks delayed lower/equal resurrection;
- `TC-CALSYNC-016`: occurrence REQUEST waits for master and rechecks a newer tombstone;
- `TC-CALSYNC-017`: external REPLY timeout reconciles one MailSubmission;
- `TC-CALSYNC-018`: availability returns private busy/free across event/call/leave;
- `TC-CALSYNC-019`: reminder outbox redelivery creates one permitted in-app row or one
  durable suppression, never re-evaluating policy;
- `TC-CALSYNC-020`: quiet hours/preferences suppress without email fallback;
- `TC-CALSYNC-020A`: transient user/preferences/repository failure leaves the reminder
  outcome pending and retriable; a post-commit WebSocket failure does not erase the
  canonical in-app row or claim realtime exactly-once;
- `TC-CALSYNC-020C`: invalid IANA timezone or incomplete/malformed quiet hours never falls
  back to UTC/delivery; it records one `policy_invalid` suppression and one bounded
  content-free health signal;
- `TC-CALSYNC-020B`: disabling the in-app channel/reminder, changing quiet hours/offset,
  moving/cancelling/deleting the occurrence, or declining/removing the attendee between
  scheduler emission and consumer handling produces one durable suppression and no in-app
  row; the opposite serialization order preserves only the already committed valid row,
  while Telegram/Web Push remain best-effort;
- `TC-CALSYNC-021`: an occurrence from another organizer cannot attach to or mutate a known
  UID series even when the occurrence key has no prior row;
- `TC-CALSYNC-022`: a whole-series CANCEL cancels materialized exceptions and a later
  occurrence-only REQUEST remains deferred until a same-organizer master is active;
- `TC-CRM-MAIL-001`: deal owner creates one ordinary idempotent Mail draft without putting
  the recipient in URL/history, sends through Mail, then explicitly links the owned sent
  message;
- `TC-CRM-MAIL-002`: admin/foreign message id cannot grant mailbox access;
- `TC-CRM-MAIL-003`: an owned existing inbound/outbound message offers only current-user
  owned deals and links exactly one through the same server-authorized PUT;
- conditional `TC-CALSYNC-023..025`: feed SSRF, conditional refresh, and snapshot failure.

Static review can accept this contract. Runtime/product status stays `NOT_IMPLEMENTED` /
`NOT_RUN` until exact Terra implementation, reviewed generation/migration and the accepted
Docker launcher produce current-SHA evidence.
