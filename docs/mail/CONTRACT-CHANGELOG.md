# Hub Mail contract changelog

> **Current review status (2026-08-14): five-document 0.1.8 entry amendment PASS at `5e314fb080d420696642058ae47397e251c23207`; Thread contract remains accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461`; implementation+oracle are `STATIC_CONFORMANT` pre-commit on base `73060a4b` (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`) and are not `STATIC_ACCEPTED` until exact atomic commit plus independent exact-SHA review; broader Phase 142/cross-contract review remains pending.**

## Candidate lifecycle sync — 2026-08-14

The historical preimplementation entry below is preserved. The two implementation/test files and
external oracle are an atomic `STATIC_CONFORMANT` pre-commit candidate; the next action is one
exact atomic commit then independent exact-SHA review. No schemas, routes, migrations, services,
generation or executable evidence are added; launcher remains `NOT_READY`, runtime `NOT_RUN`,
and `144-03B-Folder` is not dispatched.
> The root amendment closes the recorded engine-memory, public-response and Phase 144
> persistence representation gaps. Its independent PASS authorizes only 144-01A entry, not
> product code beyond that packet, generated Ent output, migration or runtime claim. Executable evidence remains
> `NOT_RUN`; the accepted launcher remains `NOT_READY`.

## 0.1.8 preimplementation amendment — 2026-08-14

`144-03A-Thread` callable/encoding/oracle contract is statically accepted at
`cbc12577a6444df9c660e43fedfd0c8133c6e461` (parent `4071ad8120850a8bde2d45d56fdedb5779270047`,
tree `2ee25fa08d343201ee72f80070c3fb58346f223c`, subject `docs(mail): freeze deterministic thread identity`)
with `EXACT_SHA_PASS`: independent Spec/API/Persistence, Security and Lifecycle/Evidence reviews
PASS, C/I/M `0/0/0`. This is docs contract acceptance, not implementation acceptance. It is
separate from the `5e314fb0` five-document 144-01A entry-only amendment and preserves Address
`STATIC_ACCEPTED` `41564d00`. It freezes strict
byte-preserved case-sensitive Message-ID literals/equality and same-account live resolver
provenance, V1 versioned raw key framing and literal SHA oracles, seed-only 1..401 participant
snapshots, 1 MiB typed JSON bounds, and closed public email-only participants. Any x/text
change requires separately accepted version/rekey/dual-read planning.
That packet is now the `STATIC_CONFORMANT` pre-commit candidate; exact next action is one
atomic commit then independent exact-SHA review, with no 144-03B-Folder dispatch. No 0.1.8
version bump, schema/generation/migration or runtime acceptance exists.

## 0.1.8 — 2026-08-13

Bounded engine/API and deterministic mailbox persistence amendment. The exact five-document
entry amendment independently PASSed at `5e314fb080d420696642058ae47397e251c23207`, which
authorizes only 144-01A entry; it is not implementation or runtime acceptance and does not
close broader Phase 142/cross-contract review.

- Statically accepted the root-frozen 143-04L `EnvelopeV1` subject amendment: the new builder
  has a 255 UTF-8-byte Hub product/safety bound distinct from the NATS transport/control-line
  ceiling, with the corresponding schema extension and static contract oracles. Independent
  Terra/high review PASS is C/I/M `0/0/0`; 143-04L is `STATIC_ACCEPTED` at
  `4525c559da22729aa8bc51ff5336c29dad4f27e9`, and status-sync commit
  `1c0299bd720ad6f8521314d83e478dedc666eccd` records it. The launcher remains `NOT_READY`,
  runtime remains `NOT_RUN`, and this records no product acceptance.
- Closed the post-freeze machine-contract review gaps: all public HTTP objects are closed;
  summary/detail composition uses internal field bases plus closed wrappers; every non-admin
  operation declares one default-deny `x-hub-authz` policy; mail received/sent addresses and
  calendar invite attendees use their canonical 200-item ceilings, and channel-due events
  carry a structured subject/payload UUID equality invariant. Added standard
  cursor/recovery/idempotent-replay examples, bounded the still-blocked legacy broadcast
  compatibility shape, and made mailbox policy's maximum 300 recipients reachable from the
  three independently capped 100-item draft collections.
  The narrow five-document entry review is accepted at `5e314fb080d420696642058ae47397e251c23207`
  only for 144-01A entry. Broader cross-contract rereview and executable Docker validation remain pending.
- Split list/change/enumeration from hydrated message content. Projection pages now have
  independent count, per-item and aggregate retained-byte ceilings with an exact
  representation-independent variable-byte measure; authorized message/blob reads and draft
  save/reconciliation carry caller budgets and reject whole values on overflow with one safe
  error.
- Closed every request/page/projection/body/blob/draft/submission DTO field. Projection
  headers and attachment metadata are fixed rather than map-shaped, draft attachments carry
  stable keys and bounded bytes, and submission is reference-only with a final-MIME ceiling.
- Added a stable Hub-derived engine draft key for lost initial-create reconciliation and a
  required one-draft `draft_client_id` carried by draft API/storage rows; attachment reuse is
  scalar one-draft ownership rather than an ambiguous many-to-many edge.
- Froze MailboxPolicy creation defaults at 25,000,000 decimal bytes for both message and
  attachment, separately from the engine adapter hard caps of 52,428,800 and 26,214,400
  bytes. Administrator configuration remains bounded by the respective cap and enforces
  attachment <= message.
- Completed OpenAPI list pagination, nested collection/string bounds, 8 MiB cursor-page
  budgets, compact calendar summaries, 366-day ranges and non-truncating 16 MiB legacy
  calendar/call envelopes. Creation quotas now have replay-stable terminal errors.
- Froze exact Phase 144-02A/B/C/D/F storage names, tuple/nullability checks, state matrices,
  normalized-name indexes, delete behavior, quotas and transactional same-owner guards in
  `PERSISTENCE-CONTRACT-0.1.8.md`.
- Required every `MailMessage` projection to persist the engine version supplied with its
  opaque engine id; optional auxiliary state is no longer allowed to stand in for version.
- Split confirmed submission identity into opaque engine submission/result-message ids and
  an immutable Hub message UUID snapshot. Public `sent` now requires a bounded result read
  and atomic projection upsert. Retained-pair recovery retries only exact account/message
  hydration and finalization, derives confirmed submission origin in the domain, and can
  never requeue or resubmit.
- Enforced the 50-attachment maximum both on the public draft value and on the locked durable
  draft-to-attachment association, so sequential saves cannot manufacture an over-limit row.
- Made settings pointers the sole persisted identity/signature default authority, made
  attachment scan state single-source, bounded recent external contacts at 1,000, and added
  immutable attempted tuples for account/domain/rule engine work.
- Froze the generic mutation receipt operation/status/code grammar, bulk target/member
  matrices and canonical rule-set hash v1. No raw key, body, MIME, provider error, secret or
  Sieve source enters PostgreSQL.
- Historical 0.1.7 CVE-v1 semantics remain included; its earlier independent PASS does not
  review these 0.1.8 changes.

## 0.1.7 — 2026-08-13

Canonical inbound iTIP identity amendment. Product runtime remains `NOT_RUN`; 143-01 and
143-02A source acceptance is unchanged. Root and independent Terra/high static review PASS
on exact semantic SHA `b3fb6f48` with no remaining P0/P1/P2, which authorized 143-02B at
the time of this contract amendment. Current implementation status belongs to `STATE.md`
and the phase evidence rather than this historical entry. Review of
initial SHA `7a6ffc1c` found three P1 ambiguities; escape, UTF-8-byte-limit and status-sync
remediation closed them. The 0.1.6 review on `5a351d5b` is historical `STALE`.

- Separated inbound duplicate identity from the accepted outbound serializer. Inbound
  occurrence REQUEST is valid without expanding or weakening the 143-02A outbound subset.
- Froze CVE-v1 as an explicitly tagged, length-delimited, big-endian canonical value stream
  domain-separated by `hub.calendar.itip.payload`, NUL and version byte 1, streamed into
  SHA-256 without retaining another unbounded payload.
- Closed every included field, enum code, sort rule, RRULE normalization, physical-format
  exclusion and unsupported property/parameter rejection. No JSON, map order, locale,
  clock, random value, raw MIME or wire folding participates.
- Closed RFC TEXT and RFC 6868 decoding to explicit allowlists, rejecting terminal/unknown
  escapes and decoded controls except the contracted LF in text/CN. All decoded field
  maxima are post-normalization UTF-8 byte counts, including multibyte boundaries.
- Fixed bounded parser maxima/defaults: 1 MiB total, 8 KiB unfolded line, 4,096 properties,
  200 attendees, 1,000 EXDATE values and exactly one `VCALENDAR` -> `VEVENT` component path.
- Paired every persisted canonical event/exception/object-state/REPLY-receipt hash with its
  version and forbade cross-version equality. The sender-bound REPLY receipt hash now has a
  separate domain prefix and includes that version.
- Chose the unattended fail-closed default for all-day `VALUE=DATE`: canonical timezone
  marker `UTC`, date/exclusive-end preserved, no inferred midnight. Founder question 27
  records the possible trusted-context alternative as a future hash-version amendment.
- Added TC-CALSYNC-039..041 and synchronized Phase 143/147/148 and QA status. OpenAPI only
  receives the contract-version bump; no HTTP request/response shape changes.
- No precomputed digest from ad-hoc host research is accepted as an oracle. The literal
  canonical byte-vector oracle remains executable `NOT_RUN` until the approved launcher is
  available; static review can verify the grammar and test construction only.

## 0.1.6 — 2026-08-13

Mailbox control-plane, safe-content and mutation-recovery amendment. Product code/runtime
remain `NOT_IMPLEMENTED` / `NOT_RUN`; 0.1.5 is historical `STALE`. Root and independent
Terra/high static reviews PASS on exact semantic SHA `5a351d5b` with no P0/P1/P2.

- Added an explicit bounded provisioning service/worker: deterministic immutable primary
  address/key/cutoff reservation, idempotent `EnsureAccount`, exact-lease activation,
  read-only safe `/mail/account` state and no request-path engine creation.
- Made Hub rules canonical desired state with a generation/hash deployment row and separate
  mutation-leased worker. One fixed Hub Sieve script is reconciled through Get-before-retry;
  rule HTTP responses expose honest pending/applying/ready/unknown/error state and no longer
  return 423 for synchronous engine work they do not perform.
- Added explicit spam/ham classifier feedback to the engine port. Bulk requests now preserve
  typed message/thread targets, include label actions, report classifier outcome separately
  from move/flag state, and use one stable mutation receipt with normalized expansion/outcome
  children. Replay never re-expands a changed thread or retries indeterminate training.
- Added generic hashed receipts for folder/label/signature/rule/presign/admin-domain/alias
  creates and message bulk intents. Full replay metadata lasts 30 days and a compact
  key/request tombstone remains for the owning scope lifetime; raw keys/response/content do
  not persist. Generic replay-unavailable/expired conflicts are distinct from specialized
  draft/send/legacy-call receipts; presign replay refreshes only the same pending object URL.
  Generic intents use v4/v7 UUID keys and a raw-keyed canonical request HMAC so small
  payloads cannot be enumerated from database fingerprints.
- Reused the receipt schema in an explicit non-wire `internal` mode for versioned folder
  update/delete recovery. Its injected UUID key is created only after lease/optimistic checks,
  is never accepted from HTTP, and may be removed after the terminal recovery window.
- Froze server-derived plain content, reviewed-sanitizer-or-plain-only behavior, received
  remote-image token rewriting and owner/message-bound proxy checks on every redirect/
  connect. The proxy returns bounded image bytes with fixed private/no-referrer/nosniff/CSP
  headers and never redirects or returns the origin URL.
- Constrained the public remote-image path token to unpadded base64url 32..1024 so traversal,
  escaping, padding and malformed bytes fail before decode, authorization or network work.
- Made label, user-settings, signature, rule-delete, mailbox-policy and domain wire writes
  match their versioned models through `If-Match`; active-rule label deletion now has a
  machine-readable no-mutation conflict.
- Replaced generic new-mail 409s with route-specific header codes for version/name/in-use,
  policy prerequisites, domain invariants and alias state. Global domain/address collisions
  disclose only `*_UNAVAILABLE`, never another tenant's owner/resource.
- Replaced the remaining generic calendar/legacy writer 409s with exact event-version,
  event-not-mutable, occurrence-not-mutable and ScheduledCall-not-mutable codes. Stale v2
  writes expose only the current quoted event ETag and same-origin canonical event Location;
  every conflict retains NFR-13 and mutates nothing.
- Required `If-Match` for draft delete and send as well as save. Same-key send replay is
  resolved first; new stale/not-editable/competing transitions have exact no-mutation codes.
- Added machine-readable attachment confirm/expiry/non-clean conflicts and constrained clean
  download redirects to fixed HTTPS storage with required `no-store`/`no-referrer`.
- Closed new mailbox request objects against unknown properties and made rule conditions/
  actions exact typed unions, preventing client scope ids or ambiguous targets on the wire.
- Made bulk move/label/thread-pin/no-argument actions mutually exclusive and forbade every
  irrelevant folder/label target field instead of silently ignoring it.
- Closed public submission/rule/domain/bulk error codes to explicit allowlists so provider
  text cannot cross persistence, API, audit or client analytics boundaries.
- Made submission message/error fields required-nullable with state-conditional invariants;
  pending cannot masquerade as sent and failed/unknown cannot carry a message id.
- Split enriched response addresses from draft recipient input so client-supplied `user_id`
  is rejected and cannot become identity or authorization input.
- Removed the duplicate message-label replace wire; receipt-backed bulk add/remove now owns
  both single and multi-target label provenance/recovery.
- Added an atomic route-metadata redaction packet because current generic access/audit paths
  would otherwise expose a remote-image token; ingress must independently redact token paths
  and mail-search queries.
- Made folders versioned desired intents with owner-visible pending/unknown/delete states and
  bounded receipt recovery inside the existing projection worker, preventing duplicate
  create or overwrite after a lost engine response without adding a fifth worker.
- Made exact-match/absence folder recovery fail closed behind a reviewed Hub-exclusive
  custom-folder-writer capability. Wave 1 exposes only Hub-internal JMAP/admin plus optional
  SMTP delivery; user IMAP/JMAP/ManageSieve/direct-admin folder mutation stays unpublished.
- Froze Meilisearch hydration for bounded sanitized plain body plus normalized attachment
  filenames, with no raw HTML/MIME/BCC/remote URL and a second current-owner authorization
  before any id/snippet response.
- Added missing folder/label management, remote-image consent, rule-deployment, provisioning
  and admin-alias UI packets. Added founder questions 24–26 for address policy, sanitizer
  dependency and fixed-script ownership.
- Bumped the OpenAPI and routed phase/QA contracts to 0.1.6; no event payload changed.
- Recorded the founder's task-local implementation-route amendment: Mail atomic code uses
  exact `gpt-5.6-terra` / `xhigh`, no fallback; Docker verification remains separately
  gated by the canonical launcher. The narrow exception is also recorded in mandatory
  first-read `AGENTS.md`, while all non-Mail routing remains unchanged.

## 0.1.5 — 2026-08-13

ScheduledCall canonical-cutover and mailbox-projection amendment. Product code and runtime remain
`NOT_IMPLEMENTED` / `NOT_RUN`; the 0.1.4 review is historical `STALE` for this revision and
the current 0.1.6 semantic review PASS is recorded separately on `5a351d5b`.

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
