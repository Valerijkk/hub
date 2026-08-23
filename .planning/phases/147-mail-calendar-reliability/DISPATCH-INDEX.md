# Phase 147 — exact Terra dispatch index

> **Status:** CALENDAR_CVE_SEMANTICS_HISTORICAL_PASS · MAIL_0_1_8_ROOT_FROZEN_REVIEW_PENDING · TERRA_CODE_ROUTE_AUTHORIZED · DEPENDENCIES_BLOCKED
> **Required runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Contract:** `CONTRACT.md` + Mail 0.1.8 root candidate
> **Execution:** one row/variant per executor; overlapping files are sequential

Every code packet starts from a root-reviewed clean SHA, writes its stated genuine unit
first, owns only its allowed files, makes no commit, and returns `BLOCKED` when a dependency,
model, generation or external gate is missing. Assemble the Common envelope plus exactly
one matching section from `DISPATCH-PACKETS.md`.

## P3 inbound source boundary — contract frozen, no dispatch

P0-P2 remain accepted. P3-A through P3-E are `UNIMPLEMENTED / PENDING / RUNTIME_NOT_RUN`; this
freeze selects no executable packet. Calendar's only future Mail ingress is sealed
`CalendarPartSource`: opaque Hub org/mailbox/message UUIDs, `[32]byte` locator hash and 1..1,048,576
bytes in; owner ID, normalized sender, method, authentication class and defensive-copy iTIP out.
The exact opaque `package mail` request/result/delivery/sink and `package calendar`
claim/scope/completion signatures are frozen in CONTRACT §2.6 and ENGINE-CONTRACT-0.1.8. V1 identity
binds canonical private locator and decoded bounded iTIP bytes; `engine_version` is private TOCTOU state
only, never receipt identity.
The Mail-owned five-second read cap, private `ReadCalendarPart`, V1 locator-hash formula, durable
preflight/recheck, exact safe errors, and receipt/ACK policy are normative in CONTRACT §2.6.

For Calendar ingestion, all earlier 147-02D/02E `GetMessage`, full-body/MIME parsing, direct engine,
provider/locator, and legacy `engine_message_id` paths are `CONSUMED/HISTORICAL`; they must not be
dispatched as a fallback. Future atom order is P3-A abstract Mail port/fake/engine/types/errors;
P3-B exact Calendar policy/port files are `event_repository.go`, `event_service_inbound_reply.go`, and
new `event_service_inbound_message.go`; P3-C exact Mail Ent files after generation are new
`mailbox_ent_calendar_part.go` and `mailbox_ent_repository.go`; P3-D exact Calendar generated-Ent
files are new `mail_inbound_repository.go`, `mail_inbound_service.go`, `mail_inbound_due_worker.go`,
`event_ent_transaction.go`, and new `event_ent_inbound_message.go`; P3-E exact files after concrete
engine/source are `platform/nats/client.go`, new Mail `calendar_inbound_consumer.go`, server `main.go`,
and `wiring_calendar_mail.go`. Stalwart remains blocked on Phase 145 and is not P3-C. Generated Ent,
migration, SQL, build, tests, lint/gofmt, parser/validator, Docker and runtime remain `NOT_RUN`.

## Dependency graph

```text
143 accepted + 144 accepted
  ├─ 147-01A inbound object state ─ 147-01B message receipt
  ├─ 147-01C reply delivery
  ├─ 147-01D reminder receipt + notification dedupe
  └─ 147-01E accepted Ent generation

143 parser/policy + 147-01A/B/E
  └─ 147-02A order policy ─ 147-02B invite service ─ 147-02C Ent transaction
145 message mapper + 144 mailbox read
[CONSUMED/HISTORICAL — superseded by P3-A..E; do not dispatch the 147-02D/02E branch below]
  └─ 147-02D authenticated calendar-part source ─ 147-02E consumer/recovery/wiring

143 RSVP + 144 draft/submission + 147-01C/E
  └─ 147-03A reply sender port/service ─ 147-03B mailbox adapter/worker

143 event schemas + 147-01E
  └─ 147-04A interval/recurrence core ─ 147-04B Ent query ─ 147-04C HTTP handler

147-01D/E ─ 147-05A1 user-policy guard ─ 147-05A2 notification dedupe
                                             └─ 147-05B scheduler/outbox
                                                  └─ 147-05C notification consumer/wiring

founder #5 + crypto/safe-dialer + 147 L0 accepted
  └─ 147-06A schema ─ 06B secret ─ 06C fetch ─ 06D parser ─ 06E service
                                                     └─ 06F Ent snapshot ─ 06G HTTP
                                                                        └─ 06H scheduler

144 owned message projection + CRM owner policy
  └─ 147-07A activity schema ─ 07B service/port ─ 07C HTTP ─ 07D typed client
                                                       ├─ 07E draft handoff ─ 07F post-send link
                                                       ├─ 07G activity timeline
                                                       └─ 07H existing-message reader link

all accepted rows ─ 147-08 docs sync/static review
```

## Atomic packets

| ID | Single owned behavior | Exact primary files | Acceptance | Dependency/gate |
|---|---|---|---|---|
| 147-01A | Incoming UID/sequence/tombstone ordering schema | `backend/ent/schema/calendarinboundobjectstate.go`, narrow edges in `calendarevent.go`/`user.go` | two nullable-recurrence partial uniques; paired canonical hash version/hash; safe state; no MIME | accepted 143 schemas + founder migration mechanism |
| 147-01B | Per-message calendar-part processing receipt | `backend/ent/schema/calendarinboundmessagereceipt.go`, narrow edge in `mailmessage.go` | unique message+locator hash; lease/retry/outcome bounds; no part bytes | 147-01A + accepted 144 schemas |
| 147-01C | External-organizer REPLY delivery ledger | `backend/ent/schema/calendarreplydelivery.go`, narrow edges in `calendarattendee.go`/`mailsubmission.go` | immutable intent unique by attendee/sequence/version; unknown distinct | accepted 143/144 schemas |
| 147-01D | Reminder receipt and persistent notification dedupe fields | `backend/ent/schema/calendarreminderreceipt.go`, narrow edges in `calendarevent.go`/`user.go`, `backend/ent/schema/notification.go` | receipt unique intent + closed safe suppression reason; optional hashed notification dedupe partial unique; no content | accepted 143 schemas + notification review |
| 147-01E | Generate and inspect Ent output | generated Ent only through accepted launcher | clean generation/compile; all four entities/indexes/edges; second diff empty | 01A–D + founder migration + launcher |
| 147-02A | Pure REQUEST/CANCEL order classifier | `backend/internal/calendar/itip_ingest_policy.go`, `backend/tests/calendar/itip_ingest_policy_test.go` | complete decision table, no I/O/clock, master/occurrence distinct | accepted 143 parser + contract |
| 147-02B | Incoming REQUEST/CANCEL domain application | `event_service_inbound_invite.go`, `backend/tests/calendar/event_service_inbound_invite_test.go` | auth admission, tombstone/defer/reactivate, one state transition/outbox | 02A + 01E + accepted 143 service |
| 147-02C | Atomic Ent ingest/receipt transaction | `backend/internal/calendar/event_ent_repository_inbound_invite.go`; no Go DB test | claim/object/event/exception/outbox/final receipt atomic; Python integration handoff; lease recovery | 02B + accepted 143 repository base |
| 147-02D | `CONSUMED/HISTORICAL — superseded by P3-A..E` | historical source/GetMessage route retained only as record | non-dispatchable; no Calendar ingestion fallback | none |
| 147-02E | `CONSUMED/HISTORICAL — superseded by P3-A..E` | historical Calendar consumer/bare-NAK route retained only as record | non-dispatchable; no Calendar full-envelope subscription | none |
| P3-A | Sealed Mail calendar-part source | `backend/internal/mail/calendar_part_source.go` (new), `mailbox_engine.go`, `mailbox_types.go`, `mailbox_errors.go`, `mailbox_fake.go` | opaque request/result, closed errors, locator-plus-decoded-iTIP hash and private-version TOCTOU read | P0-P2; no source dispatch yet |
| P3-B | Calendar transaction-aware REPLY policy/ports | `backend/internal/calendar/event_repository.go`, `event_service_inbound_reply.go`, `event_service_inbound_message.go` (new) | receipt-first single transaction; no generated Ent | P3-A; no source dispatch yet |
| P3-C | Mail Ent calendar-part source | `backend/internal/mail/mailbox_ent_calendar_part.go` (new), `mailbox_ent_repository.go` | after Mail generated Ent; Stalwart excluded | P3-B + generation |
| P3-D | Calendar inbound repository/service/due worker | `backend/internal/calendar/mail_inbound_repository.go` (new), `mail_inbound_service.go` (new), `mail_inbound_due_worker.go` (new), `event_ent_transaction.go`, `event_ent_inbound_message.go` (new) | after Calendar generated Ent | P3-C + generation |
| P3-E | Mail sanitized ingress consumer and wiring | `backend/internal/platform/nats/client.go`, `backend/internal/mail/calendar_inbound_consumer.go` (new), `backend/cmd/server/main.go`, `backend/cmd/server/wiring_calendar_mail.go` | Mail validates v1 then sink; durable disposition ACK, 65s no-durable NAK | P3-D + concrete engine/source |
| 147-03A | External REPLY delivery state machine and sender port | `backend/internal/calendar/reply_delivery.go`, `backend/tests/calendar/reply_delivery_test.go` | stable draft/send keys; unknown reconciles; permanent/transient split | 01C/E + accepted 143 RSVP/serializer |
| 147-03B | Mailbox reply adapter and bounded worker | `backend/internal/mail/calendar_reply_adapter.go`, `backend/internal/calendar/reply_worker.go`, narrow `backend/cmd/server/wiring_calendar_mail.go`, `backend/tests/mail/calendar_reply_adapter_test.go`, `backend/tests/calendar/reply_worker_test.go` | genuine fake-port units; hidden draft + one submission; no blind retry; attendee status reflects ledger | 03A + accepted 144 draft/submission + engine availability |
| 147-04A | Pure busy/free interval and occurrence normalization | `backend/internal/calendar/availability_policy.go`, `backend/tests/calendar/availability_policy_test.go` | half-open range; DST/all-day; merge/complement/bounds; no private fields | accepted 143 temporal/recurrence policy |
| 147-04B | Owner-scoped availability source queries | `backend/internal/calendar/availability_repository.go`; no Go DB test | authorize all users first; canonical linked pairs once as event, unlinked calls only, policy-visible approved leave; Python integration handoff; bounded rows | 04A + accepted 143 bridge + 01E |
| 147-04C | Availability service/DTO/HTTP composition | `backend/internal/calendar/availability_service.go`, `availability_dto.go`, `availability_handler.go`, narrow `backend/internal/calendar/handler.go` and `backend/cmd/server/wiring_modules.go`, `backend/tests/calendar/availability_service_test.go` | genuine fake-repository service unit; exact OpenAPI; stable private result; API handoff | 04A/B |
| 147-05A1 | Stable user-policy guard for preference/quiet-hours serialization | `backend/internal/notification/repository.go`; no Go DB test | preference upsert transaction locks user before optional-row query/insert/update; ordinary user UPDATE is the quiet-hours side of the same guard; Python concurrency handoff | 01D/E |
| 147-05A2 | Persistent notification dedupe seam | `backend/internal/notification/service.go`, narrow `repository.go`, `events.go`, `backend/tests/calendar/notification_calendar_dedupe_test.go` | genuine fake-repository unit; optional hash; guarded strict policy resolution/no UTC fallback; safe reason; lookup/store failures retry; legacy 2s behavior retained when absent | 05A1 |
| 147-05B | Reminder scheduler, receipt and outbox | `backend/internal/calendar/reminder_service.go`, `event_ent_repository_reminder.go`, `backend/tests/calendar/reminder_service_test.go` | injected clock; timed/all-day recurrence; current offset/start/sequence/eligibility recheck; one receipt/event; expired/suppressed | 05A2 + accepted outbox |
| 147-05C | Reminder event mapping and server scheduling | `backend/internal/notification/consumer_calendar.go`, `calendar_reminder_repository.go`, narrow `consumer.go`, `backend/cmd/server/wiring_notification.go`, `backend/tests/calendar/reminder_consumer_test.go` | genuine fake-port unit; transactional last-chance state/config/authz recheck; transient lookup NAK; durable outcome before best-effort WS; no Resend; bounded cleanup | 05B |
| 147-06A | Conditional feed persistence schema | `backend/ent/schema/calendarfeedsubscription.go`, narrow `calendarevent.go`/`user.go` edges; no Go DB test | ciphertext-only URL fields; owner/source unique; conditional/freshness state | founder #5 + migration + exact runtime |
| 147-06B | Feed URL encryption envelope | `backend/internal/calendar/feed_secret.go`, `backend/tests/calendar/feed_secret_test.go` | genuine crypto unit; AES-256-GCM, org/owner/feed AAD, no plaintext fallback/log/string | 06A schema contract + accepted key source |
| 147-06C | SSRF-safe conditional fetch | `backend/internal/calendar/feed_fetcher.go`, `backend/tests/calendar/feed_fetcher_test.go` | genuine fake-transport unit; HTTPS/443; every redirect/IP checked; 304; byte/decompression/time caps | 06B + accepted safe dialer |
| 147-06D | Bounded iCalendar feed parser | `backend/internal/calendar/feed_parser.go`, `backend/tests/calendar/feed_parser_test.go` | genuine pure unit; VEVENT/VTIMEZONE/attendee/line/depth/recurrence caps; deterministic snapshot | accepted Phase 143 parser primitives |
| 147-06E | Feed ownership and snapshot policy | `backend/internal/calendar/feed_service.go`, `backend/tests/calendar/feed_service_test.go` | genuine fake-port unit; owner auth before decrypt/fetch; retain prior on failure; disable/delete semantics | 06B/C/D |
| 147-06F | Atomic Ent feed snapshot repository | `backend/internal/calendar/feed_repository.go`; no Go DB test | source-only transactional replace/tombstone, refresh lock, safe state; Python integration handoff | 06A/E + generation |
| 147-06G | Feed HTTP API and contract activation | `backend/internal/calendar/feed_dto.go`, `feed_handler.go`, narrow `backend/internal/calendar/handler.go` and `backend/cmd/server/wiring_modules.go`, routed docs; no Go HTTP test | owner CRUD/refresh; write-only URL; exact accepted OpenAPI; Python API handoff | 06E/F + founder #5 final API approval |
| 147-06H | Feed refresh scheduler composition | `backend/internal/calendar/feed_scheduler.go`, `backend/cmd/server/wiring_calendar_feed.go`, `backend/tests/calendar/feed_scheduler_test.go` | genuine fake-clock unit; hourly jitter, one-per-feed claim, manual rate/backoff, context shutdown | 06E/F/G |
| 147-07A | Immutable CRM mail activity link schema | `backend/ent/schema/crmmailactivity.go`, narrow edges in `crmdeal.go`/`mailmessage.go`/`user.go` | unique deal+message; no content/address/tracking; safe deletion relation | accepted 144 message schema + migration |
| 147-07B | CRM compose context and owned-message link service | `backend/internal/crm/mail_service.go`, `mail_port.go`, `backend/internal/mail/crm_message_adapter.go`, `backend/tests/mail/crm_mail_service_test.go` | genuine fake-port unit; deal and message owners checked separately; idempotent link; zero hydrate on denial | 07A + accepted mailbox projection |
| 147-07C | CRM mail DTO/handlers/routes | `backend/internal/crm/mail_dto.go`, `mail_handler.go`, narrow `backend/internal/crm/handler.go` and `backend/cmd/server/wiring_modules.go`; no Go HTTP test | exact 0.1.8 paths/NFR-13; bounded timeline; not-found privacy; Python API handoff | 07B |
| 147-07D | Typed CRM mail API client | `frontend/src/api/crm-mail.ts`, `frontend/tests/mail/crm-mail-api.test.ts` | exact compose/list/link/unlink methods; opaque cursor; 204 handling; no draft/send or mailbox-authorization shortcut | 07C + accepted shared client |
| 147-07E | Deal compose-to-draft handoff | `frontend/src/modules/crm/deal-mail-compose-action.tsx`, narrow `frontend/src/modules/crm/deals-pipeline.tsx`, `frontend/tests/mail/crm-mail-compose.test.tsx` | context + Mail default identity + one idempotent ordinary draft create; recipient only in body; navigate by draft UUID; no send/link | 07D + accepted Phase 146 Mail client/routes + frontend harness |
| 147-07F | Saved-draft CRM intent and explicit post-send link | `frontend/src/modules/mail/mail-crm-link-intent.ts`, narrow `frontend/src/modules/mail/mail-shell.tsx`, `frontend/tests/mail/mail-crm-draft-link.test.tsx` | deal id only in ephemeral navigation state; normal composer/submission; link enabled only for definitive owned message; reload safely loses hint | 07D/E + accepted Phase 146 composer |
| 147-07G | Deal mail activity timeline and owner-scoped open/unlink | `frontend/src/modules/crm/deal-mail-activity-list.tsx`, narrow `frontend/src/modules/crm/deal-mail-compose-action.tsx`, `frontend/tests/mail/crm-mail-activity.test.tsx` | bounded pages; error distinct from empty; `/mail/message/:messageId` reauthorizes through Mail; explicit unlink; no body/subject/address | 07D + accepted Phase 146 direct-message route |
| 147-07H | Existing-message manual deal link from ordinary reader | `frontend/src/modules/mail/mail-crm-link-action.tsx`, narrow `frontend/src/modules/mail/mail-message-reader.tsx`, EN/RU Mail locale additions, `frontend/tests/mail/mail-crm-link-action.test.tsx` | current-user owned-deal filter from existing pipeline; explicit idempotent PUT; server remains authority; no auto-link/content copy | 07D + accepted Phase 146 reader/i18n + existing CRM pipeline/auth store |
| 147-08 | Docs synchronization and independent static/runtime review | mail-owned phase/domain docs only | implementation matrix, current SHA, findings closed; missing runtime remains NOT_RUN | all applicable accepted packets |

## Hard review points

- A forwarded or merely parseable invitation is attachment-only. No DTO may flatten
  authentication into a boolean supplied by a public caller.
- Object-state comparison and event/exception mutation occur under one transaction/lock.
  A tombstone is not represented by an invalid partially empty `CalendarEvent`.
- The message receipt is consumer idempotency; the object state is cross-message ordering;
  the inbound REPLY receipt is attendee-response history. They are not interchangeable.
- `CalendarReplyDelivery` is created with the RSVP transaction. Reconstructing delivery
  solely from a later scan of attendee state loses immutable intent and is rejected.
- Availability queries do not reuse the legacy aggregate authorization/filter as proof.
  They return no source ids or text and never partially answer an unauthorized user list.
- Notification persistent dedupe is optional for legacy callers. It must not silently
  change existing 2-second behavior when no key is supplied.
- Every linked-call/reminder-relevant mutation and both reminder stages use event → optional
  ScheduledCall → optional exception → optional attendee → user-policy guard → receipt →
  notification order.
  Preference upsert locks the same stable user row before optional rows. Neither scheduler
  nor consumer may trust the emitted payload instead of rechecking current offset/start/
  state; the consumer also re-resolves quiet hours/preferences under the guard.
- Feed packets remain ineligible until founder question 5 is answered affirmatively; an
  encrypted URL still requires SSRF validation on every fetch and redirect.
- CRM owner/admin permission never becomes mailbox delegation. The ordinary mail send API
  remains the only public send mutation. CRM recipient data is carried only in the normal
  draft request body; draft/message URLs contain opaque resource UUIDs only, and a CRM
  activity or navigation hint never authorizes Mail access.

## Gate status

| Gate | Status |
|---|---|
| Mail 0.1.8 root static review | ROOT_FROZEN engine/API/persistence amendment; historical CVE-v1 PASS on `b3fb6f48` remains semantic evidence only |
| Mail 0.1.8 independent static review | PENDING; 0.1.7 Terra/high evidence is STALE for this amendment |
| Exact Terra implementation runtime | AUTHORIZED_BY_FOUNDER / first actual packet pending |
| Accepted Docker launcher | NOT_READY |
| Ent generation/migration | NOT_RUN / FOUNDER_DECISION_PENDING |
| Engine authentication-result mapping | BLOCKED_BY_G145 |
| L1 feed founder decision | UNRESOLVED |
| Frontend unit harness | PENDING_PROTECTED_REBASE |
| Product code/tests/runtime | NOT_RUN |
