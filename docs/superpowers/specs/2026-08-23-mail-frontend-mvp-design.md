# Hub Mail Frontend MVP Design

**Status:** proposed frontend design; requires spec review and a dedicated
implementation plan before code changes.

**Scope:** a native authenticated Hub Mail workspace at `/mail`, designed to
use the repository's TanStack Router, TanStack Query, existing UI primitives,
and the existing authenticated API client. This document defines frontend
boundaries and the additional backend HTTP contract it needs. It does not alter
the approved Mail/Calendar production-integration design.

**Evidence state:** static design only. Runtime, build, lint, tests, generated
route checks, and Docker validation are **NOT_RUN**; the accepted launcher is
currently **NOT_READY**.

## Product outcome and limits

Mail is a useful, compact workspace comparable in everyday flow to Apple Mail,
not a demo page: open a mailbox, triage a bounded list, safely read a message,
compose or respond, manage messages, and answer a calendar invitation without
leaving Hub. The first release supports exactly these system folders:
`inbox`, `sent`, `drafts`, `archive`, `trash`, and `junk`, plus server-defined
custom folders. Folder labels, ordering, and counts are display data; clients
must not infer ownership or authorization from them.

The route is `/mail`. Canonical deep links are `/mail?folder=<folder-id>` and
`/mail?folder=<folder-id>&thread=<thread-id>&message=<message-id>`. A message
link may omit `thread`; the server response resolves its authoritative thread.
The active selection and search criteria live in validated route search params
so refresh, browser navigation, and Calendar return links preserve context.
Transient compose content, multi-select state, expanded quote state, and list
scroll restoration are component-local state, not a new global store.

## Information architecture

`MailPage` is a file route at `/_authenticated/mail.tsx`, wrapped in the same
`RouteErrorBoundary` pattern as Calendar. It composes a `MailWorkspace` module:

1. `MailAccountGate` obtains the account snapshot. It renders provisioning,
   unavailable, or reconnect guidance before any folder/thread request.
2. `MailSidebar` renders the account selector only when multiple accounts are
   returned, system/custom folders, stable unread counts, and the compose
   action. `sidebar-nav.tsx` gets one normal Hub navigation entry to `/mail`;
   it does not reimplement a second navigation shell.
3. `MailToolbar` owns folder title, debounced search, refresh, list actions,
   pagination controls, and the selected-item action menu.
4. `ThreadListPane` is a bounded, virtualized list. Each `ThreadRow` shows
   sender/subject/snippet, latest timestamp, unread state, flag state, and an
   attachment indicator. It never fetches or renders body HTML.
5. `MessageReaderPane` loads the selected thread's bounded message summaries
   and only then one selected message body. It contains `MessageHeader`, safe
   `MessageBody`, `AttachmentList`, `CalendarMailCard`, and reply actions.
6. `ComposeSheet` is an accessible dialog/drawer for new mail, reply,
   reply-all, and forward. It owns autosave and dirty-close confirmation.

Desktop is a three-pane layout (folder rail, list, reader) that collapses to a
folder/list view and a reader view on narrow widths. On mobile, selecting a
thread moves focus into a full-width reader with an explicit Back-to-list
control; no horizontal three-pane overflow is allowed.

## Module and API-client boundaries

Create one transport-only `frontend/src/api/mail.ts`. It owns DTOs, URL/query
encoding, and authenticated client calls, following `frontend/src/api/calendar.ts`.
It does not sanitize HTML, navigate, retain UI state, create object URLs beyond
the narrowly scoped attachment download helper, or translate business rules.

`frontend/src/modules/mail/` owns view models and UI composition:

- `mail-route-search.ts`: Zod-backed route-search parsing/defaults for account,
  folder, thread, message, query, cursor, and sort. Invalid or unauthorized
  values are discarded to safe defaults, never sent as trusted identifiers.
- `mail-query-keys.ts`: the sole query-key factory; components never handwrite
  `['mail', ...]` arrays.
- `mail-workspace.tsx`, `mail-sidebar.tsx`, `thread-list-pane.tsx`,
  `message-reader-pane.tsx`, and `compose-sheet.tsx`: composition and UI.
- `message-body.tsx`, `calendar-mail-card.tsx`, and `mail-action-menu.tsx`:
  focused security/integration/action boundaries.

Existing shadcn/Radix controls, `Button`, `ScrollArea`, dialogs/menus, Lucide
icons, `sonner`, and `@tanstack/react-virtual` are reused. No iframe, external
mail client, separate React application, provider SDK/credentials, raw MIME
viewer, new state framework, or unrelated layout redesign is introduced.

### Query keys and invalidation

The immutable key factory is:

```ts
mailKeys.account()                         // ["mail", "account"]
mailKeys.folders(accountId)                // ["mail", "folders", accountId]
mailKeys.threads(accountId, folderId, q, cursor, pageSize, sort)
mailKeys.thread(accountId, threadId)       // message metadata only
mailKeys.message(accountId, messageId)     // sanitized display payload
mailKeys.draft(accountId, draftId)
```

`threads` is cursor/page-size bounded (default and maximum limits are enforced
by the server), and virtualizes rows only within the received page. Infinite
unbounded client accumulation is forbidden. A mutation optimistically changes
only fields whose final state is unambiguous (read/unread and flag); it then
invalidates the affected thread page, `folders`, and selected `thread` as
needed. Move/archive/delete/junk, batch actions, send, draft save, and RSVP
invalidate authoritative lists rather than fabricating destination state.

## Required HTTP surface

The current `GET` account capability is insufficient for this user-facing
workspace. The backend must expose authenticated, organization/account-scoped
surfaces below, with server-derived actor and ownership; no request accepts
provider credentials or client-supplied authorization fields. Names are the
proposed stable frontend contract and must be confirmed by the backend plan.

| Capability | Required endpoint and response boundary |
|---|---|
| Gate and folders | `GET /api/mail/account` returns provisioning/status/error-safe account snapshot. `GET /api/mail/folders?account_id=` returns IDs, system kind/custom name, ordering, selectable state, and bounded counts. |
| Bounded triage | `GET /api/mail/threads?account_id=&folder_id=&q=&cursor=&limit=&sort=` returns a bounded ordered page of row-safe summaries, `next_cursor`, and no MIME/body. |
| Reader | `GET /api/mail/threads/:thread_id?account_id=` returns bounded message metadata. `GET /api/mail/messages/:message_id?account_id=` returns one display-safe body representation, headers allowed for UI, attachments, and optional calendar-card metadata. |
| Write/compose | `POST /api/mail/drafts`, `PATCH /api/mail/drafts/:id`, `POST /api/mail/drafts/:id/send`; reply/reply-all/forward are typed compose modes with server-resolved recipients/thread references. Draft save returns revision/version and canonical draft ID. |
| Message state | `POST /api/mail/messages/:id/actions` and `POST /api/mail/threads/actions` accept an allowlisted action plus server-validated IDs: `mark_read`, `mark_unread`, `flag`, `unflag`, `move`, `archive`, `delete`, `junk`, `not_junk`. Batch returns per-item accepted/rejected/unknown outcome; a transport success is not assumed to mean every item succeeded. |
| Attachments | `GET /api/mail/messages/:id/attachments/:attachment_id/download` authenticates and authorizes on every request and streams bytes with safe disposition/type headers. A list payload contains only display metadata and download URL/id, never provider URLs or raw MIME. |
| Calendar RSVP | `POST /api/calendar/events/:event_id/rsvp` accepts an allowlisted RSVP and returns canonical event/link state. Mail never interprets an iTIP body into a write request. |

Every mutation carries an idempotency key and expected revision where the
backend contract requires it. `409`/revision conflict, `401/403`, `404`,
`429`, and safe `5xx` codes are distinguishable. An ambiguous timeout,
connection loss, or `5xx` after submit is **unknown**, not a failure that may be
blindly retried: compose remains recoverable, action controls stop submitting,
the UI offers Refresh/Reconcile, and the next authoritative query decides state.

## Safe reader and attachment behavior

The reader receives only a server-produced, sanitized display representation:
either `plain_text` or a bounded `sanitized_html` fragment plus a declared
representation type. The client must render plain text as text. Sanitized HTML
is rendered in a dedicated component with no script execution, event handlers,
forms, embedded documents, active URL schemes, remote-image auto-load, or raw
MIME fallback. If the representation/type is unknown or validation fails,
render a safe text-only "content unavailable" state and preserve no unsafe
markup. Link activation requires normal safe-protocol handling and opens
external destinations with `noopener,noreferrer`.

Attachments show file name, byte size, and server-provided safe media category.
They are downloaded only after an explicit user action through the authenticated
Hub endpoint; there is no automatic preview/execution and no direct provider
URL. Download errors produce a local retry affordance without exposing provider
details. Client object URLs, if used for a completed blob download, are revoked
promptly.

## Compose, actions, and offline states

Compose uses controlled fields with recipient/subject/body validation and
server-provided reply context. Autosave is debounced, serialized per draft, and
uses the returned revision; a newer local edit is never overwritten by an older
save response. Closing with unsaved content asks to keep the local draft open
or discard; it does not silently lose text. Send disables duplicate submission.
On unknown submit outcome, it retains the draft/context and directs the user to
refresh Drafts/Sent before another send attempt.

The account gate has distinct states for `provisioning`, `ready`,
`degraded/error`, and `offline`. Provisioning explains that the mailbox is being
prepared and offers refresh; it never presents an empty Inbox as success.
Offline means no new mutating request is sent and cached/previously loaded
content is visibly stale. It does not imply a local outbox or background send
queue. Safe errors are actionable, contextual, and do not leak provider,
credential, MIME, or transport internals.

## Calendar cards and bidirectional links

When the server's trusted Mail-to-Calendar projection marks a message as an
eligible `REQUEST`, `CANCEL`, or `REPLY`, the reader renders `CalendarMailCard`
from its structured metadata. The card displays event summary, timing, current
RSVP if available, and one authoritative link to `/calendar` with the canonical
event ID/search selection. It does not parse body text, MIME, or untrusted
calendar attachment client-side.

For an eligible invitation, RSVP buttons call the Calendar RSVP endpoint and
invalidate the Mail message/thread and relevant Calendar event queries. A
successful response updates only returned authoritative card state. Cancel and
reply cards are informational unless the Calendar service explicitly returns a
permitted action. The Calendar event detail supplies the reverse deep link to
the exact Mail message; if it is deleted, inaccessible, or no longer projected,
Mail displays a safe unavailable state and Calendar remains usable. The flow
preserves the approved backend rule: only trusted internal or DMARC-aligned
mail can result in Calendar mutations, and all receipt/outbox authority remains
server-side.

## Accessibility and responsive acceptance

- Every icon-only control has an accessible name; selected folder/thread and
  unread/flag status have programmatic state, not color-only meaning.
- List rows are keyboard reachable with visible focus. Arrow-key roving focus,
  Enter/Space selection, Escape close/back behavior, and multi-select
  checkboxes are announced correctly. Virtualization preserves logical list
  semantics and does not make off-screen selection inaccessible.
- Dialog/drawer focus is trapped and restored to the invoking element. Reader
  headings, landmarks, live loading/status text, and error alerts use semantic
  elements consistent with existing Hub primitives.
- Desktop three panes preserve usable minimum widths. Tablet hides the folder
  rail behind a labelled control; mobile shows one task pane at a time with
  44px-capable touch targets and no required hover action.

## Acceptance and explicit exclusions

Static acceptance for the implementation plan is met when it maps every item
above to exact route/module/API/test ownership, retains the key factory as the
only query-key source, and includes safe/unknown-result/offline scenarios.
Executable acceptance requires the accepted task-owned Docker launcher only:
bounded route and interaction coverage for folder/thread pagination, reader
representation rejection, attachment download authorization failure, draft
autosave/version conflict, unknown send/action reconciliation, batch partial
outcomes, keyboard/mobile behavior, and Calendar RSVP/deep-link round trips.
Until that launcher is ready, all executable claims remain **NOT_RUN**.

Excluded from this MVP: provider onboarding/credential UI, IMAP/SMTP engine
implementation, mailbox sync-worker changes, raw MIME/source view, offline
outbox or background send, arbitrary HTML preview, full desktop-email parity
(rules, signatures, snooze, rich editor, contacts), advanced search language,
cross-account bulk operations, Calendar domain/object-state changes, and any
unrelated Hub navigation or visual redesign.

## Delivery and workstation constraint

Frontend work starts only after this design is reviewed and an implementation
plan freezes API DTOs, ownership, and test cases. The Mail frontend may proceed
independently of backend production composition only against an accepted
contract; it must not invent a mock protocol as production truth.

At most three disk-light subagents plus the root may work concurrently when
their ownership is independent. Do not overlap high-I/O repository scans, diff
generation, dependency work, code generation, builds, tests, Docker, or other
heavy commands. Host build/test/lint/validator/generation/application processes
are forbidden. Final executable verification is Docker-only through the
task-owned bounded launcher with CPU/RAM/PID/time limits, isolated reports and
volumes, and exact cleanup.
