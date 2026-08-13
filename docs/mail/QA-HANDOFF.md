# Hub Mail + Calendar Sync — QA handoff

> **Contract version:** 0.1.5
> **Prepared:** 2026-08-13
> **Implementation status:** NOT_IMPLEMENTED
> **Runtime status:** NOT_RUN
> **QA owner:** founder-designated Claude test program
> **Product-code owner:** ChatGPT mail feature branch

## Boundary

This handoff gives the test program stable inputs without modifying its protected zone.
The mail feature branch must not add or change anything under `testframework/**`,
`.planning/audits/test-rebase/**`, phases 128–141, or TASK-193–234.

Feature-owned genuine units are permitted only in:

- `backend/tests/mail/**`
- `backend/tests/calendar/**`
- `frontend/tests/mail/**`

No new test belongs in legacy `backend/testframework/**`.

## Contract artifact map

| Artifact | Path | Purpose |
|---|---|---|
| Architecture decision | `.planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md` | engine/source-of-truth/scope/security decisions |
| Domain + traceability spec | `.planning/decisions/specs/MAIL-SPEC.md` | ERD, state machines, requirements and named scenarios |
| Machine-readable HTTP API | `docs/mail/openapi.yaml` | all v1 public/admin endpoints and schemas |
| Machine-readable events | `docs/mail/events-v1.schema.json` | NATS envelope and payload union |
| Reliability addendum | `.planning/phases/147-mail-calendar-reliability/CONTRACT.md` | inbound ordering, reply recovery, availability, reminders, CRM/L1 seams |
| Product/UI design | `docs/superpowers/specs/2026-08-13-mail-module-design.md` | three-pane, composer and RSVP interaction |
| Founder decision queue | `docs/mail/FOUNDER-QUESTIONS.md` | defaults that do not authorize external action |

## Required test layers after accepted launcher availability

| Layer | Primary coverage | Dependencies |
|---|---|---|
| Contract | OpenAPI schema/examples, event schemas, source rules, authorization shape | no SUT runtime where possible |
| API | folder/label/thread/draft/send/event/RSVP/admin behavior | backend + fake or accepted engine profile |
| Integration | PostgreSQL constraints/outbox, JMAP adapter, NATS consumers, Meili, SeaweedFS, SMTP/iTIP | task-owned Docker profile |
| E2E smoke | internal send → inbox → invite RSVP → organizer realtime | full bounded stack |
| E2E full | external-domain flows, bulk, search, recurrence, TZ/DST, error recovery | full bounded stack + deterministic fixtures |
| Security | mailbox isolation, HTML sanitization, SSRF, attachment policy, spoofed iTIP, rate/loop | security fixture services |
| Load | list/search/unread/submit/RSVP bounded profiles | approved Locust profile |

The test program decides exact Python cases and identifiers, but must map each case to one
or more `FR-MAIL-*` / `FR-CALSYNC-*` rows and named `TC-*` scenarios in `MAIL-SPEC.md`.

## P0 scenario handoff

### Mail

1. Owner isolation prevents cross-user message/folder/attachment/engine-id access.
2. Internal delivery produces exactly one Sent and one recipient Inbox view through the
   engine path and emits idempotent NATS events.
3. Retrying a timed-out submission reconciles `Idempotency-Key` without duplicate mail.
4. Draft stale `If-Match` returns conflict and never overwrites newer content.
5. Pending/rejected attachment prevents submission; clean attachment preserves hash/name.
6. Active HTML, script handlers, unsafe SVG, CSS URL fetches, and tracking images do not
   execute or load directly.
7. Image proxy blocks private/loopback/link-local/metadata/rebinding targets before and
   after connection.
8. Search never leaks another mailbox through hit count, ids, snippets, or timing-oriented
   hydration differences in normal API semantics.
9. A lost draft-save response reconciles the opaque engine version before another write;
   PostgreSQL contains no recipient, subject, rich/plain body, or MIME copy.
10. Mail uploads use an owner/account-bound `MailObject` namespace and cannot inherit
    authorization from project-scoped file metadata.
11. `GET /api/mail/submissions/{id}` and a same-key send retry converge from `unknown` to
    the engine-confirmed result; a still-indeterminate result never triggers blind submit.
12. A draft is immutable while submission is queued/submitting/unknown; confirmation
    marks it submitted, while definitive non-delivery restores the same content for edit.

### Calendar synchronization

1. Legacy `GET /api/calendar/events` keeps `{events,from,to}` and all existing kinds;
   canonical CRUD/RSVP lives under `/api/calendar/v2/**`.
2. REQUEST create/update, CANCEL, and REPLY round-trip UID/SEQUENCE/METHOD correctly.
3. Whole-series RSVP and occurrence RSVP with `RECURRENCE-ID` remain distinct.
4. Time/location/recurrence changes reset non-organizer RSVP; silent description edit
   preserves it.
5. Concurrent stale RSVP returns canonical `{error,request_id}`, 409,
   `X-Hub-Error-Code: ATTENDEE_VERSION_CONFLICT`, current attendee-version `ETag`, and
   canonical event `Location`; the client refetches and reconciles before retry.
6. Duplicate equal UID/sequence/payload for REQUEST/CANCEL is idempotent; conflicting
   equal sequence is quarantined; lower sequence cannot revert state.
7. Reused HTTP idempotency key with identical input replays the original response; changed
   input conflicts and never emits a second mutation/outbox event.
8. REPLY changes only the normalized envelope/authenticated sender's attendee row. It
   accepts only the current event sequence and persists a bounded durable receipt for
   every distinct validated SHA-256(normalized sender + canonical payload hash). Exact
   redelivery is a no-op even after a later PARTSTAT; future/spoofed and the ninth distinct
   current-sequence reply are quarantined; no HTTP idempotency receipt is used. A distinct
   current-sequence PARTSTAT may advance that attendee once with one receipt/outbox, while
   a distinct same-status hash writes only an `observed_no_change` receipt.
9. External guest remains email-only and cannot gain Hub authorization.
10. Timed events require `starts_at/ends_at`, null date fields and `end > start`; all-day
    events require date/exclusive-end, null instants and no midnight conversion.
11. Timed events retain instants across organizer/viewer timezones and DST transitions.
12. Busy/free combines canonical events, scheduled calls, and approved HR leave. It
    authorizes the complete requested user set before source queries: self or active same-
    organization users sharing a non-deleted project/active channel, without admin bypass,
    returns only `{from,to,source}` busy data plus free complements, and never leaks title,
    ids, leave reason, project or participants. Missing/inaccessible users share one `404`;
    expansion cap returns `422` + `AVAILABILITY_LIMIT_EXCEEDED` without partial users.
    Approved leave stays viewer/shared-project scoped, expands in the employee timezone,
    and invalid legacy timezone returns `422` + `AVAILABILITY_SOURCE_INVALID`.
13. Cancelling one recurrence instance increments master sequence/version, materializes a
    cancelled exception and emits only occurrence-scoped iTIP/outbox; the series remains.
14. For both series and occurrence attendees, deliver `declined(hash A)`, then
    `accepted(hash B)`, then redeliver A at the same event sequence: the final RSVP remains
    accepted and the redelivery creates no attendee/outbox mutation. Repeat with a
    same-status distinct hash and prove its receipt prevents a later rollback; the ninth
    distinct hash is quarantined without mutation.
15. A trusted CANCEL received before REQUEST creates an owner/UID/occurrence ordering
    tombstone. A delayed lower or equal REQUEST is stale/conflicting and cannot resurrect
    it; a strictly newer authenticated REQUEST follows the frozen transition table only
    for the same organizer. Another authenticated organizer cannot take over the UID with
    a higher sequence.
16. An occurrence REQUEST received before its master is durably deferred without calendar
    bytes in PostgreSQL. Master arrival rehydrates and rechecks it; a newer occurrence
    tombstone wins. A different organizer cannot attach an occurrence to a known UID
    series. Whole-series CANCEL makes all materialized exceptions cancelled, and a later
    occurrence-only REQUEST stays deferred until a same-organizer active master exists.
17. An external-organizer REPLY uses one immutable delivery, hidden draft intent and
    `MailSubmission`; a lost result enters unknown and reconciles without a second submit.
    Public delivery status stays pending for pending/submitting/unknown, becomes sent only
    after definitive success, and becomes failed only after permanent failure.
18. Reminder lookahead creates/finds pending work; only the due-time recheck emits or
    suppresses it. Redelivery/concurrency yields one receipt and one reminder event. When
    policy permits, exactly one in-app row exists; preferences/quiet hours may instead
    produce zero. No Resend/email fallback is called. Telegram/push are asserted only as
    existing best-effort channels, invoked after canonical commit; rollback invokes none,
    and their result cannot change the committed receipt/notification outcome.
    A transient user/event/preferences/repository failure leaves the in-app outcome
    pending and NAKs for bounded retry; it is not recorded as suppression. A failed
    post-commit WebSocket publish leaves the durable row intact and is repaired by normal
    notification refetch without an exactly-once realtime claim.
    The due claim and consumer finalization each verify current sequence, effective
    occurrence start, non-null `reminder_minutes`, recomputed `due_at`, active status and
    owner/internal-attendee eligibility. Event, optional linked call, optional exception,
    optional attendee, stable user-policy guard, receipt and notification use the frozen
    lock order; preference upsert locks that
    user row before resolving an optional override. In-app channel/reminder disable,
    quiet-hours/offset change, reschedule/cancel/delete or decline/remove committed before
    finalization produces one durable suppression and no in-app row. The opposite
    transaction order preserves only the already committed valid row; Telegram/Web Push
    remain best-effort. Timed due uses the effective instant. All-day due uses the first
    representable instant of local `start_date` in the event's IANA timezone without
    changing date-only storage: midnight overlap picks the earlier instant, midnight gap
    picks the first instant after the gap, and a skipped date picks the first instant after
    that date's gap.
    Both-absent quiet-hour fields mean disabled; an incomplete/malformed pair or invalid
    IANA timezone creates one durable `policy_invalid` suppression, no notification row and
    a bounded content-free signal. It never falls back to UTC or delivery.
    Assert every receipt-state tuple: pending has no outbox/notification; emitted requires
    outbox; delivered requires notification and no reason; suppressed has no notification
    and one closed safe reason; scheduler-expired uses reason `expired`. No terminal tuple
    reopens.
19. A flood of authenticated unique unknown CANCELs cannot grow unbound state past 10,000
    per mailbox; excess input quarantines, CANCEL tombstones remain for mailbox/account
    lifetime, deferred work expires/dead-letters after seven days, and cleanup never removes
    pending/retained-event state or evicts a tombstone to admit a new UID. At cap, known UID
    updates still converge, every previously unseen UID state-changing part mutates no
    event, health is visibly degraded, and the content-free ops signal is rate-limited.
20. Backfill maps every non-deleted call, including paused/completed rows, to the project
    organization, exact-channel creator, title, description,
    instant, duration, reminder and distinct active same-project/same-organization
    participants exactly; preserves source
    audit timestamps; and emits no historical outbox/notification/mail. Any orphan,
    inactive, organization mismatch, invalid email/timezone/interval/reminder or target-bound
    overflow blocks the entire apply. Soft-deleted counts are recorded and cannot be
    reactivated; cutover requires 100% linkage of the non-deleted set.
21. Every non-deleted recurring source row compares the current raw-worker, injected-UTC
    aggregate and proposed canonical ordered occurrence sets with the pinned recurrence
    library. Parse/subset/cap/non-equality or an unanswered founder recurrence policy blocks
    migration; no row is skipped or reinterpreted silently.
22. Legacy scheduled-call create requires one stable `Idempotency-Key` from both shipped
    clients and commits event, accepted organizer/new attendee state, call projection and
    one invitation outbox together, or none. The creator is an active exact-channel member;
    creator and at most 50 participants are active same-project/same-organization members.
    V2 create can
    link only an owned active unlinked same-organization call, and its normalized mapped
    fields plus organizer/participants must exactly match that locked call. Mismatch returns
    `409` + `SCHEDULED_CALL_LINK_MISMATCH`; inactive/deleted/already-linked owned state
    returns `SCHEDULED_CALL_NOT_ATTACHABLE`; inaccessible/nonexistent stays `404`. The
    link is immutable.
    Ordinary events allow email-only guests; under founder question 23's default a linked
    call with an external attendee is rejected before any mutation.
23. Concurrent v2 and legacy mutations lock event before call. V2 stale `If-Match` mutates
    nothing; legacy PUT serializes against the locked current version. After every commit,
    canonical content/temporal/attendee fields and the call projection agree.
24. Deleting a linked legacy call cancels the event and call atomically. Pausing/resuming
    execution follows founder question 19 and cannot cancel or resurrect the event; one-time
    due completion deactivates only the call projection.
25. Cancelling the next occurrence advances the projection in the same exception/event/call
    transaction. The due worker validates canonical RRULE/EXDATE/exception state, appends one
    deterministic occurrence-keyed `calls.channel.{channelId}.scheduled_due.v1` outbox row,
    and advances or
    deactivates exactly once. Redelivery is consumer-deduplicated; event sequence/version and
    calendar reminder state remain unchanged.
26. Legacy aggregate returns one linked pair as its existing `call` kind from canonical
    temporal truth and only unlinked canonical events as `event`. V2 availability returns
    the same linked pair once as `event`; active unlinked calls are a pre-cutover-only
    fallback. After activation any unlinked non-deleted call returns
    `AVAILABILITY_SOURCE_INVALID` with no partial availability.
27. A cancelled/excluded occurrence never reappears in the legacy backend aggregate, call
    calendar UI, availability result or due worker. Only distinct internal non-organizer
    series attendees in `needs_action|accepted|tentative` mirror to call participants;
    declined/removed series attendees leave that projection, while occurrence RSVP never
    changes it. External guests never become call participants.
28. Legacy `/api/calls/conflicts` reveals title/channel only when the caller belongs to that
    exact conflicting channel. The Mail meeting UI obtains conflict state from v2 private
    availability and never treats the legacy detail reader as authorization. Any malformed,
    duplicate or over-limit participant UUID input preserves legacy validation status 400;
    it is never silently
    omitted from the authorization/conflict set.
29. The exact-channel call calendar invalidates its scheduled-call query on authorized
    `call.scheduled_due.v1`, then refetches canonical occurrences. Mount/reconnect subscribes
    to every currently returned authorized channel room; channel removal/unmount cleans it
    up. It does not create a second reminder/toast or trust payload ids without normal API
    authorization, and missed best-effort WS recovers on ordinary HTTP refetch.
30. Every exact-channel call list sends paired RFC3339 UTC `[from,to)` bounds. The visual
    view's `23:59:59.999` end becomes the exact next millisecond, including local DST day,
    week, month-grid and agenda boundaries; the server never receives an inclusive end.

### CRM mail seam

1. Only the deal owner receives compose context for a linked contact with a valid email.
   One click obtains the ordinary Mail default send-enabled identity and creates one normal
   draft with one stable idempotency key, putting the recipient only in the draft request
   body. It navigates to `/mail/drafts/{draftId}`; URL/history/query keys/logs contain no
   address, subject or body. Missing recipient is `409` +
   `CRM_MAIL_RECIPIENT_REQUIRED`; missing/unavailable Mail identity creates no draft. A
   lost create response repeats only the same key/body, and failed navigation reopens the
   returned draft id rather than creating another draft.
2. Linking the same owned projected message to the same deal is idempotent. The activity
   stores direction/timestamps/link ids only, never subject, recipient, body, MIME, engine
   id or tracking fields.
3. Deal/admin access cannot grant another user's mailbox access. A foreign message id is
   indistinguishable from not found and causes zero engine/body hydration.
4. Unlink removes only CRM metadata and is audited; it never deletes the mailbox message.
5. All inaccessible/nonexistent deal or message ids use the same `404` envelope. The
   10,001st active link returns `409` + `CRM_MAIL_ACTIVITY_LIMIT` without eviction.
6. The saved-draft route independently authorizes the draft through Mail. Only a
   definitively reconciled submission with an owned `message_id` enables an explicit link;
   queued/submitting/unknown/failed never auto-link. Losing the optional deal-id navigation
   hint on reload safely removes the prompt.
7. Timeline `Open mail` uses `/mail/message/{messageId}` and reauthorizes through Mail;
   the CRM activity row or admin CRM access is never sufficient. Loading/error/empty and
   unlink failure are distinct and accessible.
8. An existing owned inbound or outbound message can be manually linked from the ordinary
   reader. The picker uses the authenticated user id and existing authorized CRM pipeline,
   offers only deals with matching `owner_id`, and exposes no mutation when user/pipeline
   state is unavailable. The server PUT still independently reauthorizes deal and message;
   client filtering is not accepted as security evidence.

### Conditional L1 feed

These cases are required only if founder question 5 admits L1 into the first merge:

1. HTTPS fetch blocks private/loopback/link-local/metadata/rebinding targets and rechecks
   every redirect; non-443 ports, fourth redirect, >10-second, >5/10 MiB compressed/
   decompressed and ratio>10 inputs fail; URL/token never appears in log/error/API.
2. ETag/Last-Modified 304 advances freshness without event rewrite; failed refresh retains
   the last complete snapshot and marks it stale/error.
3. Owner/source uniqueness, bounded recurrence, disable/delete and cross-feed isolation
   hold. Max 10 active feeds, 5,000 VEVENTs, 500 VTIMEZONEs, 500 attendees/event and
   hourly/manual-refresh rate/backoff boundaries are enforced. No feed can overwrite Hub-
   created or mail-import events.

### Compatibility and admin

1. Existing `GET/PUT /api/admin/mail/settings` remains flat and supports
   `hub_mail_url`, `from_alerts`, `alerts_enabled`, `daily_limit`, and read-only
   `resend_ready` without frontend changes.
2. Existing `POST /api/admin/mail/broadcast` is documented as a deprecated governance gap:
   it currently reaches active users but conflicts with ADR-NOTIF-01. Mailbox acceptance
   requires removal/disablement or explicit founder re-authorization; the new mailbox must
   never invoke or expand it implicitly.
3. Mailbox policy uses `/api/admin/mail/mailbox-settings` and remains separate from Resend.
4. Every admin mail operation rejects a non-admin with 403 and writes its named audit
   action without secret values.
5. Existing scheduled-call success response fields remain compatible while bridge-owned
   behavior is additive. No client-supplied call id can bypass server-side organization,
   project, channel, owner or one-to-one-link checks.

## Event subjects

- `mail.received.v1`
- `mail.sent.v1`
- `mail.message.changed.v1`
- `calendar.invite.created.v1`
- `calendar.invite.updated.v1`
- `calendar.invite.cancelled.v1`
- `calendar.rsvp.changed.v1`
- `calendar.reminder.due.v1`
- existing channel realtime family gains
  `calls.channel.{channelId}.scheduled_due.v1` through durable outbox with a stable occurrence
  dedupe key; the obsolete unconsumed `calls.scheduled.due` publisher is removed

Every consumer must prove duplicate-event idempotency. Unknown schema versions go to the
existing dead-letter/error policy and are not heuristically parsed.
Contract fixtures must also prove every envelope/payload rejects undeclared top-level or
payload properties, including `body`, `description`, raw MIME, attachment content, token
and secret-shaped fields.

For every new subject, the producer passes a complete deterministic envelope to the
existing outbox as `PublishInTx(ctx, tx, envelope.Subject, envelope)`. Contract oracles must
prove that stored `event_type`, published NATS subject and JSON `subject` are identical;
the stored/published bytes contain the full envelope rather than only inner `payload`.
Envelope `event_id` is stable across transaction retry/reconciliation and consumer
redelivery and is the new-consumer dedupe key. The distinct outbox row id in the
`Hub-Event-Id` header is diagnostic only. Subject mismatch, malformed envelope and unknown
version must reach dead-letter with zero projection/domain mutation. Existing legacy
outbox payloads and worker behavior must remain byte-semantically unchanged.

Before producers are enabled, the shared EVENTS stream must contain `mail.>` and
`calendar.>`; linked-call due reuses the already admitted `calls.channel.>`. The generic
outbound webhook filter must not gain any of these wildcards as part of this work; any
future webhook exposure needs a separately reviewed payload allowlist.

## Docker mail profile contract

The eventual profile is optional and invoked only through the accepted versioned outer
launcher. The direct Compose command is operator documentation, not accepted test evidence:

`docker compose -f docker-compose.local.yml --profile mail up -d`

| Component | Intended image/config | Network exposure | Persistent data | Hard limit allocation | Current evidence |
|---|---|---|---|---|---|
| Mail engine | Stalwart version candidate; digest unresolved until provenance gate | internal JMAP/admin; SMTP/IMAP ports only when explicitly configured | named engine volume | ≤3 CPU / ≤3 GiB target, subject to measured rebalance | NOT_RUN |
| Hub mailbox worker/API | existing backend image/process | existing backend network only | PostgreSQL/outbox | within remaining mail budget when profiled | NOT_RUN |
| Existing NATS/Meili/SeaweedFS | shared stack, not duplicated | existing internal network | existing volumes | shared-service attribution must be reported separately | NOT_RUN |

The entire added mail profile must remain at or below 4 CPU and 4 GiB RAM. Target
allocations are not proof. The accepted run must capture actual idle and representative
load, health, restart persistence, JMAP session discovery, authenticated access, SMTP
delivery, and exact cleanup.

### Environment matrix (names only; never values in artifacts)

| Variable | Required for | Secret | Safe default |
|---|---|---:|---|
| `MAILBOX_ENGINE_ENABLED` | mailbox runtime | no | `false` |
| `MAILBOX_EXTERNAL_ENABLED` | public SMTP send/receive | no | `false` |
| `MAILBOX_INSTANCE_DOMAIN` | addresses/UID domain | no | empty; startup refuses external mode |
| `MAILBOX_ENGINE_URL` | JMAP/internal API | no | internal service URL |
| `MAILBOX_ENGINE_ADMIN_USER` | provisioning | yes-like identity | no repository value |
| `MAILBOX_ENGINE_ADMIN_SECRET` | provisioning | yes | no repository value |
| `MAILBOX_ENGINE_VOLUME` | local operator override | no | named volume |
| `MAILBOX_MAX_MESSAGE_BYTES` | policy | no | contract-defined bounded value at implementation |
| `MAILBOX_MAX_ATTACHMENT_BYTES` | policy | no | contract-defined bounded value at implementation |
| `MAILBOX_TRASH_RETENTION_DAYS` | retention | no | `30` |
| `MAILBOX_JUNK_RETENTION_DAYS` | retention | no | `30` |

DKIM private keys are generated/stored by the engine volume or secret provider and are
never environment examples containing key material.

## Fixtures and clocks

- All RFC fixtures use reserved domains (`example.test`, `.invalid`) and synthetic users.
- Calendar tests inject a clock. `DTSTAMP`, generated UID suffixes, and notification times
  are deterministic.
- DST fixtures must use real IANA zones and dates on both sides of a transition.
- MIME fixtures include folded headers, encoded words, multipart nesting, invalid depth,
  oversized values, duplicate headers, and malicious calendar properties.
- No test contacts real DNS, SMTP, remote image, or public mailbox service.

## Current blockers and honesty

- The founder-required test runner/package is `NOT_READY`; no supported execution command
  exists in this branch.
- Direct Docker/Compose or host tests are prohibited fallback paths.
- Stalwart legal acceptance, immutable digest, measured resource use, and runtime behavior
  are pending.
- Therefore this handoff is a contract deliverable only. It must not be converted into a
  PASS claim until current-SHA executable evidence exists.
