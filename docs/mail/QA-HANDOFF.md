# Hub Mail + Calendar Sync — QA handoff

> **Contract version:** 0.1.1
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
6. Duplicate equal UID/sequence/payload is idempotent; conflicting equal sequence is
   quarantined; lower sequence cannot revert state.
7. Reused HTTP idempotency key with identical input replays the original response; changed
   input conflicts and never emits a second mutation/outbox event.
8. REPLY changes only the envelope/authenticated sender's attendee row.
9. External guest remains email-only and cannot gain Hub authorization.
10. Timed events require `starts_at/ends_at`, null date fields and `end > start`; all-day
    events require date/exclusive-end, null instants and no midnight conversion.
11. Timed events retain instants across organizer/viewer timezones and DST transitions.
12. Busy/free combines canonical events, scheduled calls, and approved HR leave.

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

## Event subjects

- `mail.received.v1`
- `mail.sent.v1`
- `mail.message.changed.v1`
- `calendar.invite.created.v1`
- `calendar.invite.updated.v1`
- `calendar.invite.cancelled.v1`
- `calendar.rsvp.changed.v1`

Every consumer must prove duplicate-event idempotency. Unknown schema versions go to the
existing dead-letter/error policy and are not heuristically parsed.

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
