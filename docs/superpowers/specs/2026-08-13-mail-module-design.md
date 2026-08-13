# Hub Mail + Calendar Sync — product and interaction design

> **Date:** 2026-08-13
> **Status:** Approved for planning by autonomous founder mandate
> **Architecture:** `.planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md`
> **Contract:** `.planning/decisions/specs/MAIL-SPEC.md`

## Outcome

Add an Outlook-class mail workspace to Hub while keeping the product self-hosted and
consistent with existing Hub navigation, identity, storage, search, realtime, calendar,
notification, and audit systems. The differentiating flow is not only reading mail: an
invitation arriving in Inbox must update the same canonical event and RSVP state shown in
Hub Calendar, and a calendar change must generate standards-compliant mail updates.

The first useful vertical slice is:

1. Create a Hub event with internal attendees.
2. Generate deterministic `METHOD:REQUEST` MIME and submit it through the engine port.
3. Show the invitation in Inbox with Accept/Tentative/Decline actions.
4. Apply an optimistic RSVP transition.
5. Push the new RSVP to the organizer and both calendar/mail UIs in realtime.

The module is not called complete until external instance-domain delivery, mailbox
workflow, search, attachments, recurrence, and security gates also pass.

## Users and jobs

### Team member

- Read and triage personal work mail without leaving Hub.
- Compose rich internal/external mail with attachments and a stable draft.
- Understand a conversation as one thread.
- Accept, tentatively accept, or decline a meeting from the message itself.
- See calendar conflicts and leave before proposing a time.

### Organizer

- Create or reschedule a meeting and know who has responded.
- Trust that sequence changes, cancellation, and recurrence-instance updates converge.
- Receive realtime RSVP changes without refreshing.

### Administrator

- Configure instance domains and mail limits.
- See engine health and queue state without receiving engine credentials.
- Keep DKIM keys and operational secrets outside Git.
- Disable external mail while preserving internal/product access.

## Information architecture

Desktop uses three panes:

```text
┌───────────────┬──────────────────────┬──────────────────────────────────┐
│ Folders       │ Message list         │ Reader / thread                  │
│ 240–260 px    │ 360–400 px           │ min 480 px                       │
│               │                      │                                  │
│ Compose       │ Search/filter        │ Thread header                    │
│ Inbox      12 │ Bulk toolbar         │ Message cards                    │
│ Sent          │ Sender / subject     │ ICS invitation card              │
│ Drafts      2 │ Snippet / date       │ Reply actions                    │
│ Archive       │ Label / attachment   │                                  │
│ Trash         │                      │                                  │
│ Junk          │                      │                                  │
│ Custom        │                      │                                  │
└───────────────┴──────────────────────┴──────────────────────────────────┘
```

- The main sidebar gains **Mail** between Messenger and Calendar.
- Mobile bottom navigation is exactly **Dashboard → Messenger → Mail → Calendar → More**.
  The `More` panel is exactly **Tracker → Calls → Settings → Trash**. This replaces the
  previous bottom-level Tracker/Calls slots without removing either destination and keeps
  Mail between Messenger and Calendar on every navigation surface.
- Mobile shows one screen at a time: folders → list → thread. Browser history and URL
  route state preserve back navigation.
- Medium layouts keep list + reader; folders open in a Sheet.
- Composer is a right-side Sheet on desktop and full-screen on mobile.

## Routes and URL state

- `/mail` redirects to `/mail/inbox`.
- `/mail/:folderKey` renders folder + list.
- `/mail/:folderKey/thread/:threadId` selects a thread.
- `/mail/search?q=...` keeps shareable search/filter state.
- `/mail/drafts/:draftId` opens one owner-scoped saved draft in the ordinary composer.
- `/mail/message/:messageId` opens one owner-scoped message in the ordinary safe reader.
- `/calendar/event/:eventId` deep-links the canonical event.

Draft/message deep links carry only opaque UUIDs. Recipient, subject, body, engine id and
CRM authorization never appear in URL/query state or logs; each route loads through the
ordinary owner-scoped Mail API.

TanStack Query owns server data. URL state owns selected folder/thread/search. A local
reducer owns transient multi-select and composer interaction. Optional CRM deal-link
intent contains only a deal UUID in ephemeral navigation state and is never authorization;
losing it on reload safely removes the linking prompt. No second mailbox cache is
introduced in Zustand; only a small global unread aggregate may be exposed for navigation.

## Mail list behavior

- Unread rows use semantic emphasis and an accessible “Unread” label, not color alone.
- Sender identity resolves Hub users first and shows avatar/presence; external addresses
  use initials and the full address in the accessible name.
- A thread row shows participant summary, normalized subject, last snippet, last activity,
  message count, attachment marker, labels, priority, flag, and pin.
- Selection mode exposes Move, Labels, Read/Unread, Archive, Junk, Trash. The result is
  optimistic only for reversible presentation; server rejection restores the previous
  row state and announces the error.
- Empty, loading, partial-error, offline, and permission-revoked states are explicit.

## Reader and threads

- Messages are collapsed except the newest unread/latest message; keyboard and buttons
  expand each card.
- HTML renders inside the constrained sanitized reader. Plain text remains available.
- Remote images are blocked with sender/domain disclosure and an explicit “Show images”
  action. The raw remote URL is never requested by the browser.
- Reply, Reply all, and Forward are available from thread and individual-message menus.
- Attachments show name, type, size, scan state, download, and safe preview availability.
- A deleted/moved source message keeps a stable thread projection but cannot expose body
  content without current engine authorization.

## Composer

The composer reuses the existing Plate-based editor behind a mail-specific wrapper. It
adds deterministic HTML/plain-text serialization and quoting rather than forking the
editor implementation.

Fields:

- From alias;
- To, CC, BCC recipient chips with Hub-user/recent-contact autocomplete;
- Subject;
- priority (normal/high/low);
- rich body with plain-text fallback;
- attachments and inline images;
- selected signature.

Draft behavior:

- Debounced autosave after 800 ms of inactivity and on blur/navigation.
- At most one save in flight; subsequent edits coalesce and carry the newest revision.
- The API uses an optimistic draft revision. A conflict keeps both server and local
  versions and asks the user which content to retain.
- Closing with unsaved changes attempts one bounded save; failure leaves a local recovery
  token that contains editor state but never attachment bytes or engine credentials.
- Send uses an idempotency key. A timeout produces “delivery unknown”, not an automatic
  duplicate send.

## Calendar invitation card

Every recognized `text/calendar` part gets a card with:

- method/status badge (`Invitation`, `Updated`, `Cancelled`, `Response`);
- title, localized time, organizer, location, recurrence summary;
- timezone disclosure when organizer and viewer zones differ;
- current RSVP and version;
- Accept, Tentative, Decline;
- conflict warning and “Open in Calendar”.

States:

- `needs_action`: three primary actions enabled;
- `accepted`, `tentative`, `declined`: current choice selected; other choices remain
  available as updates;
- `cancelled`: actions disabled, cancellation prominent;
- stale sequence: action disabled and message “A newer invitation is available”;
- RSVP version conflict: refresh current state and announce that another response won;
- untrusted sender/invalid iTIP: informational attachment only, no state-changing actions;
- recurrence: ask “This event” versus “Entire series” when the payload permits both.

An RSVP click updates the canonical event first, then queues `METHOD:REPLY`. If transport
fails, the local RSVP remains recorded with a visible “Reply not delivered — retry” state;
the UI never lies by rolling back an accepted canonical response.

## Calendar integration

- Unified Calendar adds kind `event` and shows attendee response summaries.
- Meeting creation uses busy/free results from events, calls, and approved leave.
- Scheduled calls linked to events share title/time/recurrence but remain distinct
  entities. Opening a call from the calendar routes to the existing call experience.
- Reminders reuse the existing notification pipeline. Mail is not used as a new generic
  notification channel and does not expand the Resend whitelist. When preferences and
  quiet hours permit delivery, one in-app reminder is persistently deduplicated; otherwise
  the outcome is suppressed. Telegram/Web Push keep their existing best-effort boundary
  and the UI does not label them exactly-once.

### CRM mail seam

- “Write email” on an owned deal loads the authorized contact recipient and ordinary Mail
  default identity, creates one standard draft with the normal idempotent Mail endpoint,
  then navigates to `/mail/drafts/:draftId`. The recipient exists only in the draft request
  body. A lost create response retries the same key/body, and failed navigation reopens the
  returned draft UUID; CRM has no separate draft, sender, send or campaign workflow.
- After a send receipt resolves to a message, linking to the deal is an explicit user
  action. Queued/submitting/unknown/failed states never appear linked. Existing owned
  inbound/outbound messages may also be linked manually.
- The timeline shows direction/time and an “Open mail” action only for the mailbox owner.
  The action uses `/mail/message/:messageId`, whose ordinary Mail load reauthorizes the
  resource. It does not copy body, subject, recipients, tracking or engine identifiers
  into CRM.
- Unknown send state must reconcile before the UI offers a successful link. Admin CRM
  visibility never grants access to another user's message.

## Accessibility and keyboard contract

- Landmarks label folders, message list, reader, and composer.
- Folder tree uses tree semantics and supports arrows, Home/End, Enter, and collapse.
- Message list uses roving focus; `j/k` move, `x` selects, `Enter` opens, `r` replies,
  `a` replies all, `f` forwards, and `e` archives when focus is not in an editor/input.
- All shortcuts are discoverable and can be disabled.
- Focus returns to the initiating row after composer/Sheet close.
- Bulk result, save state, and RSVP changes use polite live regions.
- Minimum pointer target is 44 px on touch layouts; no status relies on color alone.

## Error and recovery policy

| Failure | User-visible behavior | Recovery |
|---|---|---|
| Engine unavailable | Read-only cached list labels data as stale; bodies not falsely cached | Retry with backoff; admin health link for admins |
| Send timeout | “Delivery status unknown” | Reconcile by idempotency key before retry |
| Projection/search delay | Newly sent/received item appears from engine response with indexing badge | Event consumer catches up; rebuild is supported |
| Attachment scan pending | Send disabled for that attachment | Remove or wait; no bypass |
| RSVP conflict | Current server RSVP replaces stale local state | User may choose a new response |
| Newer invite sequence | Older card disabled | Navigate to newest message/event |
| Remote content blocked | Placeholder with domain disclosure | Explicit proxied load |
| Reminder external channel uncertain | In-app state remains canonical; no false delivery badge | User keeps existing channel preferences; no email fallback |
| CRM send unknown | Deal link is not shown as complete | Poll/reconcile the same mail submission, then link explicitly |

## Rollout slices

1. Contract foundation: ADR, ERD, API, NATS, security, QA handoff.
2. Calendar RSVP core: models, deterministic iTIP, event CRUD/RSVP, scheduled-call link.
3. Mailbox domain: engine port/fake, folders/messages/threads/drafts, outbox projection.
4. Stalwart gateway: legal/provenance/runtime gates, instance-domain delivery, Docker
   profile, persistence and admin health.
5. Outlook workflow UI: three panes, composer, attachments, labels, search, realtime.
6. Calendar reliability: recurrence instances, imports/replies, busy/free/leave,
   reminders, L1 feeds, CRM send/log seam.
7. Security and release gate: authorization, sanitization, SSRF proxy, rate/loop limits,
   traceability, docs, independent review and Docker verification.

## Non-goals

- A general project management redesign.
- A new design system or editor framework.
- OAuth synchronization with Google/Microsoft.
- Marketing campaigns, newsletters, tracking pixels, or read receipts.
- A custom SMTP/IMAP/JMAP server.
