# Hub Mail + Calendar Sync — QA handoff

> **Contract version:** 0.1.8
> **Prepared:** 2026-08-13
> **Contract review status:** ROOT_FROZEN · INDEPENDENT_TERRA_REVIEW_PENDING
> **Implementation status:** IN_PROGRESS — 143-01/02A/02B static accepted historically; 0.1.8 amends the 144-01A boundary and freezes 144-02A/02B/02C/02D/02F schema candidates; the exact 144-02A1/A2/A3 execution split is planned but no dependent code is accepted; 143-03A/C still require the hash-pair CHECK fix; selected Stalwart has an unresolved domain-verification mismatch and enabled mailbox auth remains blocked; executable verification `NOT_RUN`
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
| Bounded engine amendment | `docs/mail/ENGINE-CONTRACT-0.1.8.md` | projection/content separation, byte budgets and stable draft identity |
| Exact persistence amendment | `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` | Phase 144 schema fields, checks, indexes and state matrices |
| Control-plane schema execution plan | `docs/superpowers/plans/2026-08-13-mail-control-plane-schemas.md` | Typed DNS then domain/policy then account/identity external-oracle packets |
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
4. Draft update, delete and new-key send preserve the exact last-confirmed `If-Match`.
   Stale/reconciliation/not-editable/submission-in-progress returns its named conflict and
   never overwrites, deletes or submits newer content; same-key send receipt replay wins
   before a now-stale version check.
5. Pending/rejected attachment prevents submission; clean attachment preserves hash/name.
6. Active HTML, script handlers, unsafe SVG, CSS URL fetches, and tracking images do not
   execute or load directly.
7. Image proxy blocks private/loopback/link-local/metadata/rebinding targets before and
   after connection.
8. Search never leaks another mailbox through hit count, ids, snippets, or timing-oriented
   hydration differences in normal API semantics.
9. A lost initial draft-create/save response reconciles the immutable engine draft key and
   opaque engine version before another write. Save and reconciliation use the same positive
   caller content budget; an adapter result one byte above it is rejected as a whole and does
   not replace canonical state. PostgreSQL contains no recipient, subject, rich/plain body,
   or MIME copy.
10. Mail uploads use an owner/account-bound `MailObject` namespace and cannot inherit
    authorization from project-scoped file metadata.
11. `GET /api/mail/submissions/{id}` and a same-key send retry converge from `unknown` to
    the engine-confirmed result; a still-indeterminate result never triggers blind submit.
12. A draft is immutable while submission is queued/submitting/unknown; confirmation
    marks it submitted, while definitive non-delivery restores the same content for edit.
13. One projection-only incremental page commits folder/message/thread projection, complete
    envelopes and the exact next composite cursor atomically. Rollback or any stale lease,
    prior cursor, sync epoch, mutation epoch or present command lease advances none of them.
14. Every synchronous engine-authoritative folder/message/draft/submission command or reconciliation
    acquires one expiring account mutation lease and advances the begin epoch before I/O;
    completion/unknown handoff clears the exact lease and advances the end epoch. Sync
    capture is forbidden while the lease is present. Active-lease contention makes zero
    engine call; sync claim or a later command can reclaim expiry only with an epoch bump,
    invalidating late work whose former owner must reconcile rather than finalize directly.
    The asynchronous rule-deployment worker uses the same lease/epoch fence; HTTP rule CRUD
    itself changes Hub desired state only and therefore never returns account-busy.
15. Invalid/expired cursor preserves current owner rows as visibly degraded, resumes a
    bounded base-state/ID-anchored enumeration without destructive pre-clear, catches up
    from its base cursor and only then tombstones unseen rows and marks ready. Query-state
    drift must converge through catch-up; missing anchor/expired base history restarts.
    Crash/restart in every enumeration/catch-up phase resumes without guessing.
16. Enumeration history at or before account event cutoff emits no inbound notification. A
    first-seen post-cutoff inbound-or-unknown message reserves one stable marker and emits
    one `mail.received.v1` even across page replay/rebuild. Only a stable mapped Hub
    submission/draft suppresses it; spoofed owner `From`, alias equality or mutable Sent
    membership never does.
    The cutoff is committed once before worker eligibility and comparison uses engine
    receipt time, not the message `Date` header or local sync observation time.
17. Two backend replicas contend through one expiring account lease. Scan/page/concurrency,
    timeout and backoff caps hold; disabled configuration starts no worker/timer/probe, and
    root-context cancellation drains in-flight work without leaked goroutines. Command I/O
    deadline is shorter than its 60-second lease, and begin/during/end sync races are covered.
18. Folder/thread/search responses always carry `projection_state` and required nullable
    `projection_as_of`. UI labels bootstrapping/degraded/rebuilding data stale or in progress,
    derives mixed pages conservatively and never fabricates ready after a query-only retry.
    Search is ready only when both mailbox projection and owner search checkpoint are ready;
    its freshness is the older complete instant.
19. The 101st non-deleted engine-backed folder is atomically rejected with 422 and
    `MAIL_FOLDER_LIMIT_EXCEEDED`. An engine account already over cap leaves the last owner
    projection degraded; no truncated folder tree is marked ready.
19a. Folder create first inserts one unique `creating` row; PATCH/DELETE use `If-Match` and
     an internal injected unexposed UUID v4/v7 receipt to persist desired state/version before engine I/O.
     Busy/stale/protected requests create neither receipt nor control change. Lost response returns/retains safe 202
     state, prevents a second key/name or folder mutation, and the existing projection worker
     recovers the oldest due receipt without a fifth worker. Exact non-system match/
     definitive absence may finalize/retry once only when the selected profile proves Hub
     is the exclusive custom-folder writer. False/unknown capability, an exposed alternate
     writer, a system-role candidate or ambiguous state makes zero recovery mutation, stays
     degraded and never duplicates/overwrites.
     Ordinary sync/rebuild preserves that intent and cannot mark the account ready; internal
     receipts are never accepted from an HTTP header or kept as lifetime public tombstones.
20. Concurrent account command/reconciliation lease contention returns 423 plus
    `MAIL_ACCOUNT_BUSY` and bounded `Retry-After` before any engine call. Unknown engine
    outcome never maps to busy, and the UI never blindly replays it as though no call ran.
21. Domain create first commits only one stable desired intent; its worker reconciles
    DNS-required/delete state through Ensure/Get without changing DNS or exposing private
    DKIM. Domain existence, `isEnabled` and expected `dnsZoneFile` never produce verified;
    only the future G145-5C authority may promote it. With engine/policy/default verified domain enabled, two account workers
    scan the same active new/existing user and reserve exactly one deterministic primary address, provisioning key
    and cutoff. A lost ensure response repeats the same key and activates one logical engine
    account/address set; separately authenticated `Identity/get|set` returns exactly one
    stable opaque id per desired address and reconciles a lost create by Get-before-retry.
    Missing domain, disabled engine, suspended user and stale lease make zero engine call.
    `GET /mail/account` is read-only and leaks no reservation, engine id,
    lease or provider text. Alias create/delete and user deactivation/reactivation advance
    account generation; pending/deleting identities are never selectable, and stale workers
    cannot finalize an older set. A claimed 12-hex UUID fallback expands once to the full
    UUID; its namespace cannot be claimed by aliases, and no collision path uses randomness
    or calls the engine before a unique reservation commits.
22. Folder/label/signature/rule/presign/admin-domain/alias create and bulk message actions
    require one stable client-generated UUID v4/v7 `Idempotency-Key`. Same key/input reconstructs the original owned
    result; changed input returns `IDEMPOTENCY_KEY_REUSED`; a removed target returns
    `IDEMPOTENCY_REPLAY_UNAVAILABLE`; compacted metadata returns
    `IDEMPOTENCY_REPLAY_EXPIRED`. Unknown engine outcome reconciles before retry and none of
    those 409 states executes again. Presign replay mints a fresh URL only for the same
    pending/unexpired attachment id without extending the object lease, and URL expiry never
    crosses it; confirmed/expired/orphaned/deleted state is unavailable and never allocates
    another object. Raw keys and response/content data never persist.
    Key digest is scoped/length-delimited SHA-256; request fingerprint is raw-keyed
    HMAC-SHA-256 over fingerprint-v1 tagged/length-delimited normalized typed fields with
    request array order preserved and no maps/unknown fields. Invalid keys are rejected
    before reservation, and small payloads cannot be enumerated from the stored fingerprint.
    The 30-day clock begins at first terminal commit; in-progress/unknown never compact and
    remain safely reconcilable/observable rather than silently becoming expired.
23. Rich draft/signature input is sanitized on the server and its plain form is derived
    there; clients cannot supply two independent canonical bodies. Without the approved
    sanitizer, rich mode fails closed to plain-only. Received hostile HTML cannot retain
    active/fetch primitives. Remote images remain blocked until explicit action/trusted
    policy and use only owner/message/expiry-bound tokens; every redirect/connect is
    SSRF-validated and origin URLs never reach DOM, API output, search, logs or errors. A
    successful proxy response is bytes with exact private five-minute cache, `nosniff`,
    `no-referrer` and restrictive CSP headers; it never redirects to an origin/cache URL.
24. Rule CRUD commits only the complete Hub desired generation/hash and returns deployment
    `pending|applying|ready|unknown|error`. The bounded worker owns one fixed script, compares
    generation/hash through Get-before-retry after lost response, cannot finalize a stale
    generation, detects/repairs external drift, and stores/logs no raw Sieve.
    `add_label` compiles only an opaque reserved engine keyword; sync maps it to a same-
    account provenance join. Unknown markers never become labels and visible removal clears
    an engine-rule marker through the fenced keyword command before clearing its source.
    The receipt-backed bulk action is the only public message-label mutation path, including
    a single message; no weaker replace-label endpoint exists.
25. Bulk selection sends exact `target_type=message|thread`; thread targets are expanded and
    deduplicated server-side under the 200-message cap. The receipt commits ordered request
    targets and exact member UUID snapshot before engine I/O, then normalized safe outcome;
    replay never re-expands changed threads. Per-target results preserve the original type.
    `junk`/`not_junk` separately report spam/ham training accepted/rejected/unknown; a folder
    move alone never appears as trained. Without proven engine idempotency/reconciliation,
    indeterminate training is terminal unknown and receipt replay cannot double-train.
26. Search finds bounded From/To/subject, server-sanitized plain body and normalized
    attachment filenames. Index documents exclude raw HTML/MIME/BCC/remote URLs/bytes;
    consumer replay/update/delete is idempotent, checkpoints drive freshness, and each hit
    is reauthorized after owner filters before id/snippet return.
27. Folder/label managers, rules deployment status, blocked-image action, provisioning
    status and admin aliases have loading/error/empty/conflict/unknown states, keyboard/focus
    coverage and stable create keys. Folder controls preserve version and render creating/
    unknown/deleting/error without claiming completion. Primary/group identities cannot be
    mutated from alias UI.
28. Settings, signature, label, rule, mailbox-policy and domain edits preserve the last
     confirmed quoted `If-Match` across one auth retry. Stale versions mutate nothing and
     refetch before user retry; signature/rule/label deletes are versioned too. Deleting a
     label referenced by an active rule returns `MAIL_LABEL_IN_USE` without removing joins,
     markers or rules. Every new mailbox 409 has its exact OpenAPI `X-Hub-Error-Code`;
     label name, policy prerequisite, domain verified/default/in-use and alias state are
     distinguishable. Cross-tenant domain/address collisions expose only `*_UNAVAILABLE`.
29. Attachment confirm with changed hash/size or non-confirmable state returns the named
    409 and schedules no scan; expired pending returns named 410. Non-clean download is 409.
    Clean download redirects only to the fixed configured HTTPS storage origin with
    required `no-store`/`no-referrer`; client/provider origin input is impossible.
30. New mailbox request objects reject unknown properties before service invocation.
     Forged organization/account/engine fields and invalid rule condition/action unions make
     zero repository, engine, classifier, storage or audit-success mutation.
    Draft recipient inputs also reject response-only `user_id`; Hub-user enrichment is
    resolved server-side from the normalized address and never grants send authorization.
    Bulk move/label/pin/no-argument variants reject every missing or irrelevant target field.
31. Application access log and generic HTTP audit record the matched route template, never
    concrete remote-image token, UUID path or search query; unmatched uses one constant.
    Proxy evidence separately proves token-path and query redaction. OTel/metrics labels
    remain template-only, and named domain audit keeps only its explicit bounded target id.
31a. Remote-image path token accepts only unpadded base64url of 32..1024 ASCII characters.
     Padding, slash, percent-escape, Unicode, punctuation and boundary overflow fail before
     token decode, repository lookup, authorization or outbound network work.
32. Submission, rule-deployment, domain-control and bulk per-target `error_code` values
    reject every string outside their OpenAPI allowlists; raw SMTP/JMAP/provider text cannot
    enter API JSON, persisted safe code, audit or client analytics.
    Submission state also enforces its nullable shape: only `sent` has `message_id`, only
    `unknown|failed` have their state-appropriate non-null safe code.
33. Every stored `MailMessage` has a non-empty bounded `engine_version` from the same engine
    projection as `engine_message_id`; optional auxiliary engine state cannot substitute for
    it. A same-id newer-version page updates both projection content and version atomically.
34. Submission success stores the opaque engine submission/message ids separately from the
    Hub UUID returned as public `message_id`. The engine message id is never parseable as a
    UUID oracle. `sent` is committed only after a bounded same-account result read and atomic
    `MailMessage` projection upsert; failure after delivery confirmation remains
    reconciliation-required. With no opaque result pair retry calls submission reconciliation;
    with a complete pair it skips reconciliation and retries only result hydration/finalization,
    never requeue or a second Submit.
35. Draft create/save accepts at most 50 distinct attachment ids. The 51st id and a direct
    repository attempt to create a 51st draft association both fail before engine I/O and
    leave the previous complete attachment set unchanged; an over-limit persisted state is
    fail-closed at submission.
36. Result hydration accepts only exact requested account, exact confirmed engine message id
    and non-empty bounded version. A mismatched account/id or spoofed adapter origin stores no
    Hub binding, returns no `sent`, publishes no outbox and never resubmits. The domain derives
    `confirmed_owner_submission` only from the persisted confirmed result pair.

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
   The duplicate key is CVE-v1, not serialized/folded wire bytes: property order, folding,
   parameter case/quoting, attendee/EXDATE order and RRULE component order normalize to
   equal hashes. Mutating each included semantic field changes the hash. Stored event,
   exception, object-state and inbound-REPLY hashes include canonical version `1`; another
   version is never compared as equal. Malformed/unknown TEXT or caret escapes, decoded
   controls other than the explicitly allowed LF, and multibyte strings exceeding the
   post-normalization UTF-8 byte bounds reject before hashing.
7. Reused HTTP idempotency key with identical input replays the original response; changed
   input conflicts and never emits a second mutation/outbox event.
8. REPLY changes only the normalized envelope/authenticated sender's attendee row. It
   accepts only the current event sequence and persists a bounded durable receipt for
   every distinct validated Mail 0.1.8 domain-separated SHA-256 over canonical version,
   length-delimited normalized sender and canonical payload hash. Exact
   redelivery is a no-op even after a later PARTSTAT; future/spoofed and the ninth distinct
   current-sequence reply are quarantined; no HTTP idempotency receipt is used. A distinct
   current-sequence PARTSTAT may advance that attendee once with one receipt/outbox, while
   a distinct same-status hash writes only an `observed_no_change` receipt.
9. External guest remains email-only and cannot gain Hub authorization.
10. Timed events require `starts_at/ends_at`, null date fields and `end > start`; all-day
    events require date/exclusive-end, null instants and no midnight conversion.
    Inbound all-day `VALUE=DATE` uses the fixed CVE-v1 `UTC` semantic marker because it has
    no trustworthy TZID. An occurrence REQUEST is accepted and hashed inbound even though
    the outbound 143-02A serializer does not emit that shape.
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
23. Concurrent v2 and legacy mutations lock event before call. V2 master PATCH/DELETE with a
    stale `If-Match` returns canonical `{error,request_id}`, `409` +
    `CALENDAR_EVENT_VERSION_CONFLICT`, the current quoted event-version `ETag` and a
    same-origin canonical event `Location`, with zero mutation. A terminal event update or
    cancellation returns `CALENDAR_EVENT_NOT_MUTABLE`. Occurrence DELETE returns
    `CALENDAR_EVENT_VERSION_CONFLICT` with the same recovery headers when the master version
    is stale, or `CALENDAR_OCCURRENCE_NOT_MUTABLE` when the instance cannot be cancelled.
    Clients never parse localized `error` text. Legacy PUT still serializes against the
    locked current version. After every commit, canonical content/temporal/attendee fields
    and the call projection agree.
24. Deleting a linked legacy call cancels the event and call atomically. Pausing/resuming
    execution follows founder question 19 and cannot cancel or resurrect the event; one-time
    due completion deactivates only the call projection. Any terminal/invalid legacy pause,
    resume, update or delete returns NFR-13 `409` + `SCHEDULED_CALL_NOT_MUTABLE` and mutates
    neither event nor call; legacy has no optimistic recovery headers because it has no
    `If-Match`.
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
3. Mailbox policy uses versioned `GET/PUT /api/admin/mail/mailbox-settings` with `If-Match`
   and remains separate from Resend.
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
`mail.received.v1` and `mail.sent.v1` admit at most 200 normalized `to` addresses, matching
the engine projection-address ceiling; 201 rejects before outbox insertion.
`mail.message.changed.v1` additionally accepts only the frozen changed-field/flag enums;
`deleted` has no live engine folder/flag state, preserves Hub-owned labels/verified links,
and removes search/counter membership idempotently.

Required `EnvelopeV1` subject contract fixtures accept an exact 255-ASCII-byte subject and
reject 256 bytes; a multibyte subject with at most 255 code points but more than 255 UTF-8
bytes; invalid UTF-8; an empty token or leading/trailing dot; wildcard, Unicode whitespace
or control characters. Every rejecting-subject error excludes payload text. The product
contract is valid UTF-8, 1..255 bytes, dot-separated nonempty tokens, no leading/trailing
dot or `..`, no `*`/`>`, no Unicode whitespace/control characters, and suffix `.v1`; this
is the Hub product/safety bound, not a NATS transport/control-line ceiling. It applies only
to the new `EnvelopeV1` builder, never globally to legacy `OutboxEvent` or the publisher.

For every new subject, the producer passes a complete deterministic envelope to the
existing outbox as `PublishInTx(ctx, tx, envelope.Subject(), envelope)`. Contract oracles must
prove that stored `event_type`, published NATS subject and JSON `subject` are identical;
the stored/published bytes contain the full envelope rather than only inner `payload`.
For `calls.channel.{channelId}.scheduled_due.v1`, the schema's
`CALLS_CHANNEL_SUBJECT_PAYLOAD_MATCH` invariant captures the UUID from `/subject`, normalizes
it as lowercase UUID text and requires equality with `/payload/channel_id`. A producer
mismatch is rejected before outbox insertion; a consumer mismatch is dead-lettered with zero
calendar, call, notification or receipt mutation. Fixtures must cover equal mixed-case UUID
text, a different valid UUID and a malformed subject. Calendar invite payloads admit at most
200 unique `attendee_ids`; 201 is rejected before outbox insertion.
Envelope `event_id` is stable across transaction retry/reconciliation and consumer
redelivery and is the new-consumer dedupe key. The distinct outbox row id in the
`Hub-Event-Id` header is diagnostic only. Subject mismatch, malformed envelope and unknown
version must reach dead-letter with zero projection/domain mutation. Existing legacy
outbox payloads and worker behavior must remain byte-semantically unchanged.

The 04L envelope immutability/payload amendment is `CONTRACT_AMENDMENT_STATIC_ACCEPTED`, ready
for code remediation redispatch after final Terra/high PASS (C/I/M `0/0/0`), with
dispatcher-enforced provenance, actual telemetry `UNKNOWN` and docs commit pending. The prior
implementation candidate remains review-failed/unaccepted (C/I/M `0/2/2`) until remediated;
the independently accepted 255-byte subject sub-amendment remains in force. Required 04L unit oracles also
prove that `EnvelopeV1` exposes no mutable exported field/reference; construction snapshots
only a successfully serializable non-null JSON object and stores frozen complete-envelope
bytes; `MarshalJSON` returns a fresh caller-owned copy; public reads are `Subject() string`
and `EventID() uuid.UUID`. Mutating source map/nested slice/struct-pointer values after
construction, and mutating returned marshal bytes, cannot change later bytes, subject or
event id. Valid struct/map objects pass. Scalars, non-nil arrays/slices, funcs/chans,
typed-nil map/slice/pointer/function/channel/interface, `null`, marshal-error and custom
array/scalar/null marshal payloads reject without panic as stable, content-free
`ErrInvalidPayload`. A custom `json.Marshaler` whose source value successfully returns a
non-null JSON object must pass; its object bytes are frozen, so later source mutation and
mutation of a `MarshalJSON`-returned slice cannot alter later envelope bytes. An
RFC3339-capable `time.FixedZone("UTC", 0)` occurrence passes, normalizes to `Z` and produces
schema-valid output; nonzero offset rejects. Launcher is `NOT_READY`; runtime is `NOT_RUN`;
global Mail review remains pending and no product/runtime acceptance is implied.

`mail.>` and `calendar.>` admission remains the 04G/Phase configuration gate, not 04L.
Before producers are enabled, the shared EVENTS stream must contain them; linked-call due
reuses the already admitted `calls.channel.>`. The generic outbound webhook filter must not
gain any of these wildcards as part of this work; any future webhook exposure needs a
separately reviewed payload allowlist.

## Machine contract conformance

1. Every non-admin OpenAPI operation has exactly one `x-hub-authz` value declared in
   `x-hub-authz-vocabulary` v1. An absent/unknown policy fails the contract gate. Every admin
   operation retains `x-hub-rbac: admin` and its named audit action.
2. Mailbox ids resolve only inside the authenticated user's mailbox; calendar reads require
   organizer/internal-attendee visibility, writes require organizer ownership, RSVP requires
   the matching internal attendee, and CRM link/unlink requires both deal and mailbox-message
   ownership. A missing scope relationship and a nonexistent id return the same NFR-13 `404`
   before engine/body hydration or mutation and do not disclose another organization. The
   read-only `GET /api/mail/account` probe is authorized against the user's mail namespace,
   requires no existing account resource and may return the safe `not_configured` state.
3. Every public request/response object and `ErrorEnvelope` rejects undeclared properties.
   The only intentionally open OpenAPI object components are internal thread/message field
   bases, and they are exposed only through public wrappers with
   `unevaluatedProperties: false`.
4. Standard OpenAPI examples must validate against their adjacent schemas and cover: a page
   stopped before the next complete item with a non-null cursor; calendar version `409` with
   `X-Hub-Error-Code`, quoted `ETag` and same-origin `Location`; mailbox `423` with bounded
   integer `Retry-After`; and a calendar create whose identical key/request replays the
   original `201` and byte-equivalent body without another mutation.
5. Draft `to`, `cc` and `bcc` each admit at most 100 normalized unique recipients. Their
   cross-collection normalized count must be at most policy `max_recipients`, whose persisted
   and public maximum is 300. Exactly 300 is accepted when each list fits; 301 rejects before
   draft save, submit, receipt success or engine I/O.
6. The deprecated legacy admin broadcast remains blocked for mailbox acceptance and is not
   re-authorized by this contract. While the compatibility route still exists, unknown input
   fields reject, subject/body have explicit UTF-8 ceilings, counters are non-negative int32,
   `sent + failed <= recipients`, and dry-run requires zero sent/failed. Phase 148 still
   requires disable/removal unless the founder separately re-authorizes the feature.

## Docker mail profile contract

The eventual profile is optional and invoked only through the accepted versioned outer
launcher. The direct Compose command is operator documentation, not accepted test evidence:

`docker compose -f docker-compose.local.yml --profile mail up -d`

| Component | Intended image/config | Network exposure | Persistent data | Hard limit allocation | Current evidence |
|---|---|---|---|---|---|
| Mail engine | Stalwart `v0.16.17` source candidate; Docker Hub OCI index `sha256:a8108e19…`, linux/amd64 manifest `sha256:2bcbd379…`; signature/SBOM not verified and candidate not approved | Hub-only shared `/jmap` endpoint with separately authorized Registry and mailbox method contexts; exact OAuth helper exception only if accepted; optional SMTP delivery only; no user IMAP/JMAP/ManageSieve/admin mailbox mutation | named `/etc/stalwart` config and `/var/lib/stalwart` data volumes | ≤3 CPU / ≤3 GiB target, subject to measured rebalance | registry metadata only; runtime NOT_RUN |
| Hub mailbox domain/account/combined projection-recovery/rule workers + API | existing backend image/process | existing backend network only | PostgreSQL/outbox | within remaining mail budget when profiled | NOT_RUN |
| Existing NATS/Meili/SeaweedFS | shared stack, not duplicated | existing internal network | existing volumes | shared-service attribution must be reported separately | NOT_RUN |

The entire added mail profile must remain at or below 4 CPU and 4 GiB RAM. Target
allocations are not proof. The accepted run must capture actual idle and representative
load, health, restart persistence, JMAP session discovery, authenticated access, SMTP
delivery, incremental and base-anchored-enumeration recovery, two-replica lease exclusion,
worker cancellation/drain, and exact cleanup.

### Environment matrix (names only; never values in artifacts)

The 2026-08-13 exact-source review invalidated both the old single administrator-secret
shorthand and the interim separate-URL model. A Stalwart management API key authenticates
only reviewed Registry `x:*` methods on shared `/jmap`; it cannot authenticate standard
JMAP mail. Typed operation context, not URL, selects the credential, and a JMAP batch may
not mix Registry and mail methods. The following names express the required separation;
any enabled mode remains rejected until the mailbox authentication/delegation design in
founder question 42 passes security review.

| Variable | Required for | Secret | Safe default |
|---|---|---:|---|
| `MAILBOX_ENGINE_ENABLED` | mailbox runtime | no | `false` |
| `MAILBOX_EXTERNAL_ENABLED` | public SMTP send/receive | no | `false` |
| `MAILBOX_EXCLUSIVE_FOLDER_WRITER` | enables automatic unknown-folder recovery only for an accepted no-alternate-writer profile | no | `false` |
| `MAILBOX_INSTANCE_DOMAIN` | addresses/UID domain | no | empty; startup refuses external mode |
| `MAILBOX_ENGINE_ORIGIN` | trusted internal origin for shared `/jmap`, individually allowlisted `/api` helper path+method pairs, health and same-origin JMAP discovery/session URLs | no | empty; enabled startup refuses it |
| `MAILBOX_ENGINE_MANAGEMENT_API_KEY` | scoped provisioning through exact Registry methods on fixed `/jmap` only | yes | no repository value; never used for standard mail methods, mixed batches, upload/download or `/api/*` helpers |
| `MAILBOX_ENGINE_MAIL_AUTH_MODE` | mailbox JMAP authentication/delegation mode | no | `disabled`; no other mode accepted before G145-5B |
| `MAILBOX_ENGINE_MAIL_PRINCIPAL` | conditional mailbox JMAP identity selected by G145-5B | yes-like identity | no repository value |
| `MAILBOX_ENGINE_MAIL_SECRET` | conditional mailbox JMAP credential selected by G145-5B | yes | no repository value; never used for Registry `x:*` methods or mixed batches |
| `STALWART_PUBLIC_URL` | conditional single advertised OAuth/OIDC/JMAP origin selected by G145-5D | no | absent; never guessed from request headers or set before the OAuth topology is accepted |
| `MAILBOX_REMOTE_IMAGE_TOKEN_KEY` | encrypted/signed proxy tokens | yes | startup disables remote-image loading if absent |
| `MAILBOX_ENGINE_VOLUME` | local operator override | no | named volume |
| `MAILBOX_MAX_MESSAGE_BYTES` | policy | no | contract-defined bounded value at implementation |
| `MAILBOX_MAX_ATTACHMENT_BYTES` | policy | no | contract-defined bounded value at implementation |
| `MAILBOX_TRASH_RETENTION_DAYS` | retention | no | `30` |
| `MAILBOX_JUNK_RETENTION_DAYS` | retention | no | `30` |

DKIM private keys are generated/stored by the engine volume or secret provider and are
never environment examples containing key material.

The source snapshot and endpoint/object evidence are recorded in
`.planning/phases/145-stalwart-gateway/RESEARCH-2026-08-13-STALWART-0.16.17.md`.
It is not image provenance or runtime evidence. Personal per-mailbox rules must use RFC 9661
`SieveScript/*` on `/jmap`; management `x:SieveSystemScript` and `x:SieveUserScript` are not
substitutes. The documented `x:SpamTrainingSample` object does not yet freeze Hub's
message-to-spam/ham mapping or retry idempotency, so classifier feedback remains blocked.
The management API-key design is provisionally narrowed to `Replace` with only domain and
individual-account CRUD/list permissions, pending independent security review. Exact
candidate source confirms mailbox `Identity/get|set` supplies stable ids for configured
account addresses, so identity mapping is not a port mismatch; its enabled use remains
blocked by G145-5B mailbox authentication. Shared-account ACL is not a full-engine fallback:
selected-version Email/Mailbox methods admit shared access, while Identity,
EmailSubmission, Sieve and Blob require account membership. OAuth exposes Authorization
Code and Device Authorization, not client credentials. Until founder question 42 selects
per-user Authorization Code OAuth or explicitly accepts broad impersonation, config accepts
only `MAILBOX_ENGINE_MAIL_AUTH_MODE=disabled` and every other mail-auth mode fails before
network I/O. QA must also prove both typed contexts use fixed `/jmap`, wrong-context
credentials and mixed Registry/mail batches fail before I/O, and provisioning never uses
`/api`. If OAuth is selected, one `STALWART_PUBLIC_URL` also drives JMAP discovery, so
QA must prove one accepted HTTPS origin for browser and Hub, exact user-visible OAuth paths,
an exact-method/rate-limited `POST /api/auth` exception, and negative access to shared
management/mail `POST /jmap`, every unrelated `/api/*`, upload/download, admin/static UI and
telemetry from the user-facing boundary. Prefix-wide `/api`, default/allow-all
`Http.allowedEndpoints`, steady-state 8080, unused IMAP/POP3/submission/ManageSieve
listeners, permissive CORS and untrusted forwarded headers
are release blockers. The selected Domain object exposes expected zone
text but no manual-DNS verification result, so domain promotion/account provisioning remains
blocked by founder questions 43-44 and G145-5C. `STALWART_RECOVERY_ADMIN` is operator-only
bootstrap/recovery state and is absent from the steady-state Hub credential matrix.

## Fixtures and clocks

The 0.1.8 delta additionally requires maximum / maximum+1 oracles for projection item/page
bytes, message/blob budgets, every paginated response count/cursor/8 MiB budget, 16 MiB
legacy calendar/call rejection, stored-resource quotas and 50 MiB/25 MiB policy ceilings.
Compile/static oracles must also prove that projection pages structurally contain neither
body nor attachment bytes, every page/change/enumeration request carries one unambiguous
cursor/checkpoint plus limit, and `EngineSubmission` contains only the confirmed draft
reference/version/key and a bounded final-MIME ceiling rather than another content copy.
Body-union fixtures prove an inbound rich read exposes one untrusted rich source only and
cannot reach API/search before sanitization/plain derivation, while an outbound rich draft
contains the accepted sanitized rich value plus only its server-derived plain value.
Quota failures must replay from the same receipt without later creating a resource after
capacity is freed. Draft fixtures must prove lost first-create reconciliation by immutable
engine draft key and reject a second draft claiming the same staging scope/attachment.
Schema fixtures must exercise every tuple/nullability/state CHECK, the receipt operation and
bulk count/code matrices, and canonical rule-hash byte vectors. These are required future
oracles, not current executable evidence.

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
- Stalwart legal acceptance, independent digest/provenance approval, signature/SBOM,
  exact config, least-privilege mailbox auth, domain-verification authority, measured
  resource use, and runtime behavior are pending. `v0.16.17` is
  only a researched source/registry candidate; read-only digest capture did not pull or run it.
- The 0.1.8 root amendment still needs an independent exact-SHA Terra/high review. Its
  provenance may be `dispatcher-enforced` only when the native dispatcher accepted the exact
  model/effort without rejection; actual telemetry, when exposed, must match. Historical 0.1.7
  review evidence is stale for its engine/API/persistence changes.
- Therefore this handoff is a contract deliverable only. It must not be converted into a
  PASS claim until current-SHA executable evidence exists.
