# Phase 147 — executable packet contracts
> **P3-D schema amendment (2026-08-26; frozen):** before D1/D2, add the exact
> nullable co-null receipt fields `deferred_object_state_id` and
> `deferred_expires_at`, with a many-to-one `NoAction` object-state FK,
> first-deferral write-once CAS, exact `+7d` database-clock deadline and all
> later due values capped to it. Expired or attempt-8 missing-master work is
> dead-lettered before Mail I/O. Only the receipt/object-state Ent schemas are
> authorized for the source atom; generated Ent and migration follow through
> the accepted Docker schema launcher.
> **P3-C source checkpoint (2026-08-26):** `da197b74` is
> `SOURCE_AUTHORED / STATIC_REVIEW_PASS / DOCKER_COMPILE_NOT_RUN` for the exact
> concrete Mail Ent calendar-part source files. It owns the five-second bound,
> short durable pre/post snapshots, one private engine read, verified
> attestation comparison, and closed safe errors without exposing provider IDs
> or the private locator. Contract Terra/high PASS C/I/M `0/0/0`; adversarial
> Terra/xhigh found no security blocker. Its former generation blocker is
> `CONSUMED/HISTORICAL`; P3-D/E and executable/runtime gates remain pending.
> **147-02A static acceptance (2026-08-23):** `3f4d174f888f08a119bbe381dc86e5e9b34e0301` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly the sole production file `backend/internal/calendar/itip_ingest_policy.go`; independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. Tests are intentionally absent under the current production-only constraint and executable/generation/runtime evidence remains `NOT_RUN`. The accepted scope is pure REQUEST/CANCEL order policy only: no product, merge, readiness, schema, generation, migration, SQL, Docker, or runtime authority. D0-next is `CONSUMED/HISTORICAL`; P3-C, 147-02B/147-02C, D1/D2, and P3-E retain their stated blockers; no next unblocked P3 source is selected.
> **P3-D0 static acceptance (2026-08-23):** immutable source `47faf3a705e5641f4ceb7ee365c3dfeeeace725d` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/ent/schema/calendarinboundmessagereceipt.go`. It adds required immutable `organization_id`/`mailbox_id` UUID scalars and database nonzero checks, with no Ent edges or raw data; existing uniqueness, state-shape, and due behavior are unchanged. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. The D0-next route is `CONSUMED/HISTORICAL`. P3-C remains blocked on Mail Ent generation; 147-02B/147-02C and D1/D2 remain blocked on active Calendar generation/dependencies; P3-E is conjunction-blocked. No next unblocked P3 source is selected. Generation, migration, SQL, build, tests, Docker, and runtime are `NOT_RUN`; no readiness, merge, or product claim.

> **Assembly rule:** Common envelope + exactly one numbered packet.
> **Runtime:** `gpt-5.6-terra` / `xhigh`, fallback none.
> **Current state:** all product/test packets are model-blocked; generation/runtime packets
> also require the accepted launcher and their named founder/external gates.

## P3-A/P3-B/P3-D0 accepted static source; P3-C historical route; no next unblocked atom — not an executable packet

P3-A is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` at `05f87f87f6dc30bd57648030026faf5912add3e7`
(`feat(mail): add sealed calendar part port`) for exactly `calendar_part_source.go`, `mailbox_engine.go`,
`mailbox_types.go`, `mailbox_errors.go`, and `mailbox_fake.go`; independent Contract/high and
Security/xhigh are APPROVED C/I/M `0/0/0`. It accepts sealed ABI/port only, not concrete source/I-O,
MIME/provider, Ent, NATS, wiring, or runtime. P3-B is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` at
`43a6c42af88524be2951966bdbfc83b422ecb7f9` for exactly
`backend/internal/calendar/event_repository.go`, `event_service_inbound_reply.go`, and new
`event_service_inbound_message.go`: transaction ABI only, with Contract/high and Security/xhigh APPROVED
C/I/M `0/0/0`. CONSUMED/HISTORICAL: P3-C was the sole next `UNIMPLEMENTED / PENDING / BLOCKED` atom, awaiting Mail Ent
generation for exactly new `backend/internal/mail/mailbox_ent_calendar_part.go` and
`mailbox_ent_repository.go`; P3-D/E remain pending/blocked. All generation, migration, SQL, build, test,
lint/gofmt, parser/validator, Docker, and runtime evidence remains `NOT_RUN`; no merge, product, or
readiness claim follows. This section
supersedes prior 147-02D/02E Calendar ingestion packets only: they must not use `GetMessage`, full
body/MIME parsing, direct MailEngine or Mail Ent, provider IDs/raw locator, or legacy event
`engine_message_id`. Do not dispatch a fallback.

P3-A accepted files are `backend/internal/mail/calendar_part_source.go` (new), `mailbox_engine.go`,
`mailbox_types.go`, `mailbox_errors.go`, and `mailbox_fake.go`; no concrete source/read behavior is
accepted. P3-B is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` at `43a6c42af88524be2951966bdbfc83b422ecb7f9` for exact
`backend/internal/calendar/event_repository.go`, `event_service_inbound_reply.go`, and new
`event_service_inbound_message.go`, policy/ports only with no generated Ent. CONSUMED/HISTORICAL: P3-C was the sole next pending/blocked atom after Mail generation: new
`backend/internal/mail/mailbox_ent_calendar_part.go` plus `mailbox_ent_repository.go`; it owns concrete
source/I-O, byte/five-second bound, and durable pre/post behavior. Stalwart is separately
Phase-145-blocked. P3-D remains pending after Calendar generation: new `mail_inbound_repository.go`,
`mail_inbound_service.go`, `mail_inbound_due_worker.go`, `event_ent_transaction.go`, and new
`event_ent_inbound_message.go`. P3-E remains pending after concrete engine/source: it is
`backend/internal/platform/nats/client.go`, new Mail `calendar_inbound_consumer.go`, server `main.go`,
and `wiring_calendar_mail.go`. Its future Mail decoder validates v1 and sends sanitized delivery to an
injected Calendar sink; Calendar never subscribes to the full envelope. On no durable disposition use
`NakWithDelay(65s)`, MaxDeliver20, retained EVENTS plus one content-free degraded/manual-replay signal;
semantic durable disposition ACKs. No source, test, generation, migration, SQL, build, lint/gofmt,
parser/validator, Docker, runtime, merge, product, or readiness authority.

The exact opaque `package mail` request/result/delivery/sink and `package calendar`
claim/scope/completion transaction signatures are in CONTRACT §2.6 and ENGINE-CONTRACT-0.1.8. V1 identity
binds canonical private locator plus SHA-256 of decoded bounded iTIP bytes; `engine_version` is private
before/after TOCTOU state only. The engine adapter retains the private locator, recomputes and returns its
verified locator-hash attestation, and rejects a mismatch. P3-C accepts bytes only when that attestation
equals the durable requested hash and the independent durable post-I/O snapshot still matches; it never
receives or exposes the private locator.

## Common execution envelope

- Role: Hub `codex-micro`, atomic Go/TypeScript implementer. Write the stated external-
  package genuine unit first, then the minimum implementation. Do not invent component or
  network tests when the packet calls only for a pure unit.
- Root supplies exact clean branch/head and accepted dependency signatures. You are not
  alone: do not revert others. Dirty overlap, absent prerequisite, incompatible frozen
  type, actual-model mismatch, or broader behavior means `BLOCKED` before writes.
- Modify only Allowed mutations for the selected packet. No commit/push/PR, dependency,
  secret, network, listener, real DB/broker/filesystem, generated manual edit, or wall
  clock except an explicitly injected clock.
- Never read/change `testframework/**`, legacy `backend/testframework/**`, test-rebase
  audits/phases 128–141/TASK-193–234, `pro-qa-ai/**`, `.env.local`, `.git/index.lock`.
- Host tests/build/lint/validators/generation/processes are forbidden. Direct Docker/
  Compose is not accepted fallback. Missing launcher => verification `NOT_RUN`.
- Preserve the legacy calendar `{events,from,to}`, existing Resend settings/broadcast
  implementation, notification callers without a dedupe key, CRM pipeline behavior, and
  outbound webhook allowlist unless this packet explicitly owns a narrow additive seam.
- Every producer uses accepted 143-04L and passes the complete envelope to the unchanged
  transaction outbox with `event_type == envelope.subject`. Every consumer compares broker
  subject to JSON subject and deduplicates on envelope application `event_id`; the separate
  `Hub-Event-Id` transport header is diagnostic only. Do not edit/wrap the global outbox.
- Go 1.25; exported comments in Russian; bounded immutable DTOs; stable safe errors; no
  body/MIME/feed URL/token/address/provider response in logs or errors.
- Response in Russian: summary, exact files/symbols, oracle results, verification, risks,
  requested/actual model evidence, source SHA. Depth/fan-out 0; retries 2/root cause;
  wall clock 45m pure/schema, 60m service/handler, 75m composition/runtime.

## 147-01A — incoming object state schema

- Allowed only `backend/ent/schema/calendarinboundobjectstate.go` and the narrow inverse
  edges in `backend/ent/schema/calendarevent.go` and `backend/ent/schema/user.go`.
- Implement CONTRACT §2.1 exactly. Use two PostgreSQL partial unique indexes for nullable
  recurrence. Hash length, nullable expiry and enum/state bounds are structural; cross-
  field/order/10,000-unbound cap policy remains in service. A nullable event edge permits
  early CANCEL/deferred occurrence.
- Oracles: series/occurrence uniqueness predicates; owner+UID scope; no raw MIME/message
  body/engine id; no generated/migration edit; pre/post inverse edge inventory.

## 147-01B — inbound message receipt schema

- Allowed only `backend/ent/schema/calendarinboundmessagereceipt.go` and the narrow inverse
  edge in `backend/ent/schema/mailmessage.go`.
- Implement CONTRACT §2.2. The unique key is message UUID + fixed SHA-256 locator hash.
  Bound attempts/error code; processing lease/retry timestamps nullable; finalized state
  does not retain the calendar part or authentication detail. Service cleanup later keeps
  deferred work at most seven days and finalized receipts 30 days.
- Oracles: exact unique columns; safe state enum; receipt can exist without event; raw
  locator/provider error absent; no generated/migration edit.

## 147-01C — reply delivery schema

- Allowed only `backend/ent/schema/calendarreplydelivery.go` and narrow inverse edges in
  `backend/ent/schema/calendarattendee.go` and `backend/ent/schema/mailsubmission.go`.
- Implement CONTRACT §3. Unique attendee + sequence + attendee version; immutable intent
  fields; optional submission; explicit unknown; bounded retry metadata.
- Oracles: series/occurrence both valid; two later attendee versions create two intents;
  duplicate same intent conflicts; no body/organizer/from/key; no generated edit.

## 147-01D — reminder receipt and notification dedupe schema

- Allowed only `backend/ent/schema/calendarreminderreceipt.go`, narrow inverse edges in
  `backend/ent/schema/calendarevent.go`/`user.go`, and
  `backend/ent/schema/notification.go`.
- Implement CONTRACT §5. Receipt unique columns include event, internal user, normalized
  occurrence key, event sequence and due-at; use partial indexes if recurrence is nullable.
  Receipt also has nullable in-app pending/delivered/suppressed outcome and notification
  edge plus the closed safe suppression-reason enum from CONTRACT §5. Notification gains
  an optional fixed hash field and a partial unique
  user+type+hash index.
- Oracles: legacy notifications with null key remain legal; keyed duplicates conflict;
  suppressed requires reason and pending/delivered forbid it; receipt contains no title/
  details; state/outbox/outcome/notification/reason combinations match the CONTRACT table;
  exact pre/post Notification field/index inventory;
  retention fields permit 90-day final cleanup without pending deletion; no generated edit.

## 147-01E — Ent generation/inspection

- Dependencies: reviewed 01A–D, founder migration mechanism, accepted launcher. Allowed
  mutation is task-owned ignored generated output/report only; do not manually edit or
  stage generated Ent or SQL.
- Generate twice inside the bounded launcher. Inspect all four types, partial predicates,
  enum types and inverse edges; compile current schema; second generation must be empty.
- Missing launcher or migration authority => `BLOCKED/NOT_RUN`, no direct command.

## 147-02A — pure ingest order policy

- Allowed only `backend/internal/calendar/itip_ingest_policy.go` and
  `backend/tests/calendar/itip_ingest_policy_test.go`.
- Model the complete CONTRACT §2.4 table as typed input/result. Normalize nothing and do no
  I/O here; accept already canonical method/hash/sequence/current object state plus the
  occurrence's series admission state. Series and occurrence remain distinct RFC component
  scopes; the latter may return defer because its master is absent/cancelled.
- Oracles: none/request/cancel current states × lower/equal same/equal conflict/newer;
  early cancel tombstone; deferred occurrence; different series organizer quarantine;
  cancelled/absent master cannot be reactivated by occurrence alone; reactivation only
  newer within the same component key; deterministic zero value cannot mean apply.

## 147-02B — inbound REQUEST/CANCEL service

- Allowed only `backend/internal/calendar/event_service_inbound_invite.go` and
  `backend/tests/calendar/event_service_inbound_invite_test.go`.
- Accept only the Phase 143 bounded parser result plus server-created authentication enum.
  Resolve calendar owner from mailbox source, never input. Apply/defer/tombstone through
  the 02A result and repository transaction; emit the accepted calendar invite event only
  on material event/exception change.
- Oracles: internal and aligned external admit; untrusted/forwarded zero mutation; delayed
  lower/equal conflict; different-organizer higher-sequence takeover quarantines; CANCEL
  before REQUEST; missing/cancelled-master occurrence; whole-series cancel covers existing
  exceptions; one atomic call; no public/body types in state.

## 147-02C — Ent inbound transaction

- Allowed only `backend/internal/calendar/event_ent_repository_inbound_invite.go`. Do not
  add a Go DB/component test; hand transaction/concurrency cases to Python integration.
- Claim/finalize message receipt, lock series state before occurrence state, compare/recheck,
  mutate event or exception, advance tombstone/state and write outbox in one transaction. A transient
  failure releases/advances bounded retry; permanent quarantine finalizes without event or
  outbox. Deferred master recovery retains only source references/hashes. Enforce the
  10,000 unbound-state cap, mailbox-lifetime CANCEL tombstones, seven-day deferred expiry
  and 30-day finalized-receipt cleanup without deleting live/pending work. At cap, a
  previously unseen UID quarantines before mutation, already-known UID state still
  converges, and one rate-limited content-free degraded-health/audit signal is recorded;
  no cleanup path evicts a tombstone to recover capacity.
- Oracles: concurrent equal/newer ordering; rollback; lease expiry; one outbox; no partial
  event; series CANCEL covers materialized exceptions; later master wakes only matching
  owner/UID/organizer current deferred occurrences; cap behavior for known/unseen UID;
  approved quota increase only, with no automatic tombstone reset.

## 147-02D — authenticated calendar-part source

> **CONSUMED/HISTORICAL — superseded by P3-A..E; non-dispatchable.** The prior direct
> Stalwart/GetMessage/test route is preserved only as historical intent. It is replaced by
> package-mail sealed source, internal exact-one private MIME locator-plus-decoded-iTIP hash;
> no raw locator/provider/content reaches Calendar, and no source is authorized here.

## 147-02E — ingest consumer/recovery/wiring

> **CONSUMED/HISTORICAL — superseded by P3-A..E; non-dispatchable.** The prior Calendar
> full-envelope consumer and bare transient-NAK route are retained only as history. P3-E is a Mail
> consumer that validates legacy v1, drops `engine_message_id`, invokes a sanitized Calendar sink,
> ACKs only after durable semantic disposition, and uses `NakWithDelay(65s)` solely for no durable
> disposition; no source is authorized here.

## 147-03A — reply delivery state machine

- Allowed only `backend/internal/calendar/reply_delivery.go` and
  `backend/tests/calendar/reply_delivery_test.go`.
- Define a narrow sender port that accepts immutable serialized REPLY intent and stable
  delivery UUID, and returns sent/definitive-failure/indeterminate. State transitions are
  pending→submitting→sent|unknown|failed; unknown invokes reconcile and never send again.
- Oracles: stable draft-create/submission keys; timeout/crash states; bounded transient
  retry; newer attendee response creates a new delivery without changing the old one;
  cancelled/deleted/identity-invalid safe permanent result.

## 147-03B — mailbox REPLY adapter and worker

- Allowed only `backend/internal/mail/calendar_reply_adapter.go`,
  `backend/internal/calendar/reply_worker.go`, the narrow existing
  `backend/cmd/server/wiring_calendar_mail.go`,
  `backend/tests/mail/calendar_reply_adapter_test.go`, and
  `backend/tests/calendar/reply_worker_test.go` using fake mailbox/domain ports only.
- Implement the 03A port through accepted mailbox draft/reconcile/submission services.
  Build a hidden draft request from deterministic iTIP bytes, owner account and allowed
  identity. Persist/link the existing MailSubmission result; on unknown call reconcile.
- Oracles: exactly one draft intent/submission; owner/org/from checks before engine; engine
  unavailable retry; policy rejection permanent; crash recovery; attendee public delivery
  status maps pending/submitting/unknown to pending and permanent failure to failed; no body
  persistence/log.

## 147-04A — availability pure core

- Allowed only `backend/internal/calendar/availability_policy.go` and
  `backend/tests/calendar/availability_policy_test.go`.
- Implement half-open clipping, same-source merge, all-source union, free complement and
  minimum-gap filtering from CONTRACT §4. Use already expanded typed intervals; no DB/user
  lookup. Accept injected IANA locations/occurrence expander and return explicit cap error.
- Oracles: touching/overlap/nesting; source attribution; range edges; empty/full busy;
  5/240-minute gap filters; DST gap/overlap; all-day exclusive end; deterministic ordering;
  1,000/user and 5,000 total caps.

## 147-04B — availability source repository

- Allowed only `backend/internal/calendar/availability_repository.go`. Do not add a Go
  database/component test; route query/auth/transaction coverage to Python integration.
- Resolve requester's organization, validate all requested active users in one bounded
  query plus bounded shared-project/channel relation query, then load only relevant
  canonical events/attendees/exceptions, active unlinked scheduled calls/participants only
  while the pre-cutover transition flag is enabled, and approved non-deleted leave through
  the existing viewer/shared-project policy for `[from,to)`. A linked pair is returned once
  as event, never again as call; an unlinked non-deleted row after activation fails the
  entire source load as `AVAILABILITY_SOURCE_INVALID`.
  Return domain records, including each target's timezone, not public text DTOs.
- Oracles: foreign/deleted/missing/unrelated user causes no source queries; admin has no
  bypass; attendees/creator scope; declined/cancelled/deleted filters; approved leave only
  under viewer/shared-project scope; row/order bounds; no broad org-wide leave query.

## 147-04C — availability HTTP composition

- Allowed only `backend/internal/calendar/availability_service.go`,
  `availability_dto.go`, `availability_handler.go`, narrow additions to
  `backend/internal/calendar/handler.go` and `backend/cmd/server/wiring_modules.go`, and
  `backend/tests/calendar/availability_service_test.go` as a genuine fake-repository
  service unit. Do not add a Fiber/API Go test.
- Bind exact 0.1.8 request, enforce ≤31 days/IANA/unique 1–50/slot bounds, invoke 04B then
  04A, and serialize stable request-order `{user_id,busy,free}`. Use generic not-found for
  inaccessible users and NFR-13 errors.
- Oracles: no source IDs/title/reason/project; timed instants and all-day zones; same input
  stable output; inaccessible user is generic 404; cap is 422 +
  AVAILABILITY_LIMIT_EXCEEDED; invalid employee timezone is 422 +
  AVAILABILITY_SOURCE_INVALID with no UTC fallback; legacy `/calendar/events` byte-shape
  untouched.

## 147-05A1 — stable user-policy guard

- Allowed only `backend/internal/notification/repository.go`. This is a PostgreSQL/Ent
  transaction behavior, so add no Go DB/component test; route its concurrency proof to the
  Python integration handoff.
- Make `UpsertPreference` execute one transaction that locks the stable user row before
  querying, updating or first-inserting an optional preference row. The existing auth
  profile `UPDATE users` already takes the same row lock when quiet hours change; do not
  modify auth DTO/service/repository. Expose the narrow caller-owned transaction primitive
  needed by 05A2/05C to lock the guard and resolve quiet hours/preferences on the same Ent
  transaction. Do not change the HTTP wire API.
- Oracles: existing and absent preference in both serialization orders; quiet-hours UPDATE
  versus guarded read; user missing/disabled; rollback leaves no preference mutation;
  unrelated user does not block; exact pre/post methods; Python integration handoff; no
  generated/migration edit.

## 147-05A2 — persistent notification dedupe

- Allowed only `backend/internal/notification/service.go`, narrow additions to
  `repository.go`, `events.go`, and
  `backend/tests/calendar/notification_calendar_dedupe_test.go` as a genuine
  fake-repository unit.
- Add optional safe dedupe hash to `DispatchEvent`. When present, atomically create/find
  the one in-app row under the 01D unique index and finalize the locked reminder receipt's
  in-app outcome, or finalize suppressed without a row. A terminal outcome is not
  re-evaluated on duplicate delivery. When absent, preserve current callers and 2-second
  behavior exactly. Add
  `calendar_reminder` type and calendar push path; do not add mail/Resend channel.
- For the keyed calendar path, use 05A1's caller-owned transaction primitive to read quiet
  hours and resolve optional preference rows after the stable user-row guard. Evaluate
  quiet hours at an injected policy clock instant acquired under that guard. The unkeyed
  legacy dispatch path keeps its existing behavior.
- Oracles: keyed concurrency/redelivery; absent-key regression; prefs/quiet hours; in-app
  disabled; guarded policy fake call and policy change in both serialization orders;
  both-absent quiet hours disabled; incomplete/malformed pair or invalid IANA timezone
  creates `policy_invalid` suppression and bounded content-free signal with no UTC fallback;
  user/preferences/repository failure leaves outcome pending and returns a retriable error
  instead of suppression. For the keyed path, return/queue external channel work for
  caller-owned post-commit execution; never invoke Telegram/push before the canonical
  transaction commits. External result remains best effort and cannot mutate/rollback the
  receipt outcome. No raw caller key is persisted/logged.

## 147-05B — reminder scheduler/outbox

- Allowed only `backend/internal/calendar/reminder_service.go`,
  `event_ent_repository_reminder.go`, and
  `backend/tests/calendar/reminder_service_test.go`.
- With injected clock/claim lease, create/find pending occurrences in bounded lookahead for
  owners and internal attendees. Timed due derives from effective occurrence instant;
  all-day due derives from the first representable instant of its civil start date in the
  valid event IANA zone: overlap chooses the earlier instant, a midnight gap chooses the
  first instant after the gap, and a skipped date chooses the first instant after its gap.
  Date-only storage is unchanged. At/after due time lock event → optional ScheduledCall → optional exception →
  optional attendee → user-policy guard → receipt, then require current sequence,
  effective start, active user, non-null `reminder_minutes`, recomputed due and eligibility
  to match exactly before atomically marking emitted plus writing one complete 143-04L
  `calendar.reminder.due.v1` envelope through the unchanged transaction outbox. Its
  application event id/time/correlation are stable explicit claim inputs. Expire >24h overdue;
  suppress stale, configuration-mismatched, cancelled, deleted, removed or declined work.
  A changed offset uses a separate receipt; disabled offset creates no new work. Pre-due
  RSVP changes do not terminally suppress a pending receipt.
- Oracles: timed/all-day/IANA/DST, midnight overlap/gap and skipped civil date; recurrence;
  duplicate/concurrent scans; offset disable/change, update/cancel and attendee race in
  both serialization orders; owner-attendee
  dedupe; external guest absent; transaction rollback; no content in event.

## 147-05C — reminder consumer and scheduling composition

- Allowed only `backend/internal/notification/consumer_calendar.go`,
  `calendar_reminder_repository.go`, narrow additions to `consumer.go`,
  `backend/cmd/server/wiring_notification.go`, and
  `backend/tests/calendar/reminder_consumer_test.go` using fake ports.
- Validate the complete closed reminder envelope, including broker/JSON subject equality,
  and refetch owned current event metadata. Deduplicate on envelope application event id,
  never on the `Hub-Event-Id` transport header alone. In one
  PostgreSQL transaction lock event → optional ScheduledCall → optional exception → optional attendee → user-policy
  guard → receipt → optional notification. Resolve current quiet hours/preferences only
  after the guard; reauthorize and require current active state, user eligibility, sequence,
  effective start, non-null offset and recomputed due to match the receipt/payload.
  A definitive mismatch durably suppresses and acks with no row. Otherwise dispatch with
  receipt/event-derived dedupe hash and existing preferences/quiet hours, atomically
  persisting one delivered/suppressed in-app outcome before ack; duplicate events reuse it.
  A transient user/event/preferences/repository read or write failure leaves the outcome
  pending and NAKs through the existing bounded retry path. It is never converted to
  suppression. Only after commit invoke any permitted Telegram/Web Push work and publish
  WebSocket invalidation, all as best effort; a rollback invokes none. Clients recover a
  missed WebSocket publish through the ordinary notification list/query refetch.
  Wire one context-owned bounded scheduler; preserve existing notification/automation
  consumers.
- Oracles: duplicate JetStream event one permitted in-app row or one durable suppression;
  policy change after first handling cannot change outcome; offset disable/change,
  reschedule/cancel/delete, attendee decline/remove, in-app channel first-insert/disable and
  quiet-hours races in both transaction orders; disabled in-app channel/quiet suppression;
  external Telegram/push race is asserted best-effort only; rollback produces no external
  attempt and external failure cannot alter the committed outcome; transient lookup/store
  failure NAKs with no terminal outcome; invalid policy suppresses with its exact safe
  reason; title fetched after authorization;
  stale/deleted event durable suppression;
  post-commit WebSocket failure preserves the row and exposes no exactly-once claim;
  unknown schema existing dead-letter path; no email or broadcast call.

## 147-06A — conditional feed persistence schema

- Eligibility: founder question 5 affirmative, exact Terra and migration authority.
  Allowed only `backend/ent/schema/calendarfeedsubscription.go` and narrow inverse edges in
  `backend/ent/schema/calendarevent.go` and `backend/ent/schema/user.go`.
- Implement CONTRACT §6 storage fields: required ciphertext/nonce/key version, safe display
  host/path mask, conditional validators, freshness/error/state/version, owner and source
  relations. No plaintext URL, response body, credential or provider error field.
- Oracles: owner/source uniqueness; max/state/value bounds; no cascade into Hub/mail-import
  events; no Go DB/component test; no generated/migration manual edit.

## 147-06B — feed URL encryption envelope

- Allowed only `backend/internal/calendar/feed_secret.go` and
  `backend/tests/calendar/feed_secret_test.go` as a genuine pure crypto unit.
- AES-256-GCM uses the accepted instance key source and organization/owner/feed UUIDs as
  versioned AAD. Encrypt before persistence; missing/wrong key, AAD or version fails closed.
  Plaintext is request-lifetime only and no String/error/log method exposes it.
- Oracles: randomized ciphertext/nonces, correct round trip, every AAD component mismatch,
  truncated/tampered input, unknown key version and unavailable key. Never use network/DB.

## 147-06C — conditional SSRF-safe fetch

- Allowed only `backend/internal/calendar/feed_fetcher.go` and
  `backend/tests/calendar/feed_fetcher_test.go` using an injected fake transport/resolver/
  clock. Reuse the accepted platform safe-dial primitive; do not make real network calls.
- HTTPS port 443 only, no userinfo/fragment; at most three redirects with URL resolution and
  DNS/IP revalidation on every hop and dial. One 10-second total budget covers redirects and
  reads. Compressed bytes ≤5 MiB, decompressed ≤10 MiB, ratio ≤10; validators are bounded.
  A 304 is a distinct bodyless success.
- Oracles: loopback/private/link-local/metadata, mixed DNS/rebinding, redirect pivot,
  non-443, timeout, compressed/decompressed/ratio boundary, invalid content type, 200/304,
  ETag/Last-Modified and fully redacted URL/error.

## 147-06D — bounded iCalendar feed parser

- Allowed only `backend/internal/calendar/feed_parser.go` and
  `backend/tests/calendar/feed_parser_test.go` as a genuine pure unit. Reuse reviewed Phase
  143 lexical/temporal primitives without weakening their limits.
- Produce one deterministic source snapshot with at most 5,000 VEVENTs, 500 VTIMEZONEs and
  500 attendees/event, enforcing Phase 143 total bytes, unfolded-line, property/component
  depth and recurrence-expansion caps before allocation/expansion.
- Oracles: valid timed/all-day/recurring snapshots, duplicate UID resolution, malformed ICS,
  line/depth/count boundaries, invalid timezone, recurrence bomb, deterministic hashes and
  no source URL/body in errors.

## 147-06E — feed ownership and snapshot policy

- Allowed only `backend/internal/calendar/feed_service.go` and
  `backend/tests/calendar/feed_service_test.go` using fake secret/fetch/parser/repository
  ports. Authorize owner before decrypt/fetch, enforce max ten active feeds and one manual
  refresh per five minutes, then pass one complete snapshot to the repository.
- A fetch/parse/apply failure keeps the prior snapshot and records only safe stale/error
  metadata. Disable stops refresh and retains a visibly stale snapshot. Delete requests
  soft deletion of only this source's projections. Backoff is bounded at 24 hours.
- Oracles: cross-owner zero decrypt/fetch, missing key, 304 freshness, failure retention,
  max/rate/backoff boundary, disable/delete, cancellation and no URL in audit/error.

## 147-06F — atomic Ent feed snapshot repository

- Allowed only `backend/internal/calendar/feed_repository.go`. Do not add a Go database/
  component test; route transaction/concurrency cases to Python integration.
- Claim one subscription refresh, atomically upsert subscription+UID source rows and
  tombstone source rows absent from the complete snapshot. Never match only on UID and never
  modify Hub/mail-import/another-feed events. Failure updates safe subscription state in a
  separate bounded transaction without altering the last complete snapshot.
- Oracles: unique source identity, concurrent claim, 200 replace, 304 no rewrite, rollback,
  deletion/tombstone and parent-retention behavior; static query review plus Python handoff.

## 147-06G — feed HTTP API and contract activation

- Allowed only `backend/internal/calendar/feed_dto.go`, `feed_handler.go`, narrow additions
  to `backend/internal/calendar/handler.go` and `backend/cmd/server/wiring_modules.go`, and
  the routed Mail/OpenAPI docs explicitly named after founder question 5 approval. Do not
  add a Go HTTP/component test; route API cases to Python.
- Activate owner-scoped list/create/get/patch/delete/refresh. URL is accepted only on create
  or explicit replacement and never returned; responses expose a safe display mask. All
  errors use NFR-13 with frozen max/rate/version codes.
- Oracles: auth before lookup/fetch, exact routes, invalid UUID/body, optimistic conflict,
  masked response, no secret in audit, max-ten/manual-rate boundary and no legacy route drift.

## 147-06H — feed refresh scheduler composition

- Allowed only `backend/internal/calendar/feed_scheduler.go`,
  `backend/cmd/server/wiring_calendar_feed.go`, the narrow call in
  `backend/cmd/server/wiring_modules.go`, and
  `backend/tests/calendar/feed_scheduler_test.go` using an injected fake clock/service.
- Claim due enabled feeds once, schedule hourly with deterministic bounded jitter, honor
  per-feed backoff capped at 24 hours, and stop all work on context cancellation. Missing or
  disabled dependencies create no goroutine/listener.
- Oracles: due/not-due ordering, jitter bounds, overlap claim, backoff, shutdown/restart,
  disabled feature and no wall clock/network/DB in the genuine unit.

## 147-07A — CRM mail activity schema

- Allowed only `backend/ent/schema/crmmailactivity.go` and narrow inverse edges in
  `crmdeal.go`, `mailmessage.go`, `user.go`.
- Implement CONTRACT §7 immutable metadata and unique deal+message link. Direction and
  occurred-at are server-derived; deletion follows CRM link policy without cascade-delete
  of mail. Relink restores the same soft-deleted row; service owns the 10,000 active-link
  cap.
- Oracles: exact fields/unique/edges; no content/address/subject/engine/tracking field;
  parent hard-delete is restricted until link tombstone cleanup; no cascade deletes deal or
  mail implicitly; pre/post existing CRM/mail edge inventory; no generated/migration edit.

## 147-07B — CRM mail service and ownership port

- Allowed only `backend/internal/crm/mail_service.go`, `mail_port.go`,
  `backend/internal/mail/crm_message_adapter.go`, and
  `backend/tests/mail/crm_mail_service_test.go` using fake CRM/mail ports.
- Compose context requires the deal owner and linked contact email. Link/unlink/list first
  authorize the deal, then independently resolve an owned message projection through the
  port without body hydration. PUT same deal/message is idempotent.
- Oracles: foreign deal/message/org/admin-without-mailbox-right produces generic not-found
  and zero engine call; inbound/outbound direction/server time; no contact email persisted
  on activity; same-row relink; missing compose recipient is 409 +
  CRM_MAIL_RECIPIENT_REQUIRED; 10,001st link is 409 + CRM_MAIL_ACTIVITY_LIMIT without
  eviction; GDPR link deletion does not delete message.

## 147-07C — CRM mail HTTP

- Allowed only `backend/internal/crm/mail_dto.go`, `mail_handler.go`, narrow additions to
  `backend/internal/crm/handler.go` and `backend/cmd/server/wiring_modules.go`. Do not add a
  Go HTTP/component test; route the public API cases to the Python API handoff.
- Implement exact compose-context and mail-activity GET/PUT/DELETE routes from OpenAPI
  0.1.8. Bound list to max 100 newest-first with an opaque cursor; actor/deal/message are
  path/auth derived; NFR-13 errors and not-found privacy.
- Oracles: route/method/path shapes; duplicate PUT; invalid UUID; auth ordering; no body/
  recipient/subject in activity response; exact 404/409 codes; existing contacts/pipeline
  unchanged.

## 147-07D — typed CRM mail API client

- Allowed only `frontend/src/api/crm-mail.ts` and
  `frontend/tests/mail/crm-mail-api.test.ts` as a genuine typed-client unit with
  in-process fetch fakes.
- Implement exact compose-context and activity list/PUT/DELETE functions from OpenAPI
  0.1.8. Use bounded `limit`, opaque cursor and shared safe error parsing. This client has
  no draft/save/send function and cannot accept actor, mailbox or organization ids.
- Oracles: exact methods/paths/query/body/204; 404/409 codes remain available to UI;
  cursor stays opaque; response types contain no content/address beyond compose context;
  no new auth headers, alternate send endpoint, campaign or tracking shape.

## 147-07E — deal compose-to-draft handoff

- Allowed only `frontend/src/modules/crm/deal-mail-compose-action.tsx`, a narrow mount in
  existing `frontend/src/modules/crm/deals-pipeline.tsx`, and
  `frontend/tests/mail/crm-mail-compose.test.tsx` using typed in-process fakes.
- On one explicit `Write email` action, fetch compose context, obtain the ordinary Mail
  default send-enabled identity, and call the accepted `createMailDraft` client once with
  a stable idempotency key for that click intent. Put the normalized recipient only in the
  normal draft body; initialize other required fields through accepted Mail composition
  policy. Navigate only after 201 to `/mail/drafts/$draftId`, passing at most `deal_id` in
  ephemeral router state. A lost create response retries only the same key and exact body;
  while unknown, no new-key intent is allowed. A navigation failure reopens the returned
  UUID and never creates again. Do not send or link in this packet.
- Oracles: repeated render does not create; double-click shares one intent/key/request;
  401 and lost-response retry preserve key/body; new-key retry while unknown is blocked;
  failed navigation reuses draft id; missing recipient/default identity/error is visible
  and stays in CRM; URL/history/query keys/logs contain no address/subject/body; keyboard/
  focus and 44px touch target; existing drag/drop and project navigation remain unchanged.

## 147-07F — saved-draft CRM intent and explicit post-send link

- Allowed only `frontend/src/modules/mail/mail-crm-link-intent.ts`, a narrow addition to
  `frontend/src/modules/mail/mail-shell.tsx`, and
  `frontend/tests/mail/mail-crm-draft-link.test.tsx`. The route file
  `frontend/src/routes/_authenticated/mail/drafts/$draftId.tsx` belongs to accepted
  Phase 146 packet 02A and is a dependency, not an extra write here.
- Validate an optional ephemeral `deal_id` navigation hint without treating it as
  authorization. Load the draft through the ordinary owner-scoped Mail API and use the
  ordinary composer/autosave/submission hooks. Only a definitively reconciled submission
  with owned `message_id` may show an explicit `Link to deal` action using 07D PUT.
  Reload/back-forward may lose the hint; that safely removes the prompt and never changes
  draft or delivery state.
- Oracles: no auto-link; queued/submitting/unknown/failed has no successful link action;
  same-key reconciliation; link 404/409 visible without changing Mail success; forged or
  stale hint grants no Mail/CRM access; no recipient/body in route or persisted state.

## 147-07G — CRM mail activity timeline

- Allowed only `frontend/src/modules/crm/deal-mail-activity-list.tsx`, a narrow composition
  in `frontend/src/modules/crm/deal-mail-compose-action.tsx`, and
  `frontend/tests/mail/crm-mail-activity.test.tsx` using typed in-process fakes.
- Render bounded newest-first activity pages with direction/time, explicit unlink and
  `Open mail`. The latter navigates only to `/mail/message/$messageId`; that accepted
  Phase 146 route independently loads/authorizes the ordinary Mail reader. CRM activity is
  never accepted as mailbox authorization. Error, loading and empty states are distinct.
- Oracles: cursor pagination/order; duplicate link collapse; unlink confirmation/error;
  404/permission revocation; open route contains UUID only and reauthorizes; no body,
  subject, address, engine id, campaign/tracking state or second mailbox cache.

## 147-07H — existing-message manual deal link

- Allowed only `frontend/src/modules/mail/mail-crm-link-action.tsx`, a narrow composition in
  accepted `frontend/src/modules/mail/mail-message-reader.tsx`, narrow matching key
  additions to `frontend/src/i18n/locales/en/mail.json` and
  `frontend/src/i18n/locales/ru/mail.json`, and
  `frontend/tests/mail/mail-crm-link-action.test.tsx` using typed in-process fakes. Do not
  modify `frontend/src/api/crm.ts` or the auth store.
- After the ordinary Mail reader has definitively loaded an owned inbound/outbound message,
  load the existing authorized `getPipeline()` result and current user from `useAuthStore`.
  Offer an explicit deal picker containing only deals with `owner_id === user.id`, then use
  07D's idempotent PUT. Client filtering is not authorization: forged/stale selections and
  admin-visible unowned deals must still fail at the server without body/engine hydration.
  Missing user, pipeline error or no owned deals exposes no mutation; never auto-link.
- Oracles: owned inbound/outbound happy path; same link double-click converges; unowned deals
  filtered even for admin-visible pipeline data; forged deal id 404; user/pipeline loading,
  error and empty states; keyboard/focus/44px target; no subject/address/body/engine id in
  CRM request, local persistence, query key, URL or log; reader security boundary unchanged.

## 147-08 — documentation and independent review

- Allowed only routed mail/calendar/CRM/notification docs and phase verification artifacts
  explicitly supplied by root. Compare current code to CONTRACT, OpenAPI, events and every
  accepted packet/TC. Record exact SHA and `NOT_RUN` for unavailable runtime.
- Required independent lenses: calendar ordering/transactions, mailbox reconciliation,
  privacy/authz, notification guarantees, CRM boundary, frontend recovery, migrations and
  docs consistency. P0/P1 findings block acceptance; no product-ready claim.
