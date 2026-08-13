# Mail 0.1.7 deep review — OpenAPI response bounds

> **Detected:** 2026-08-13 root deep static audit
> **Severity:** P1 machine-contract boundary
> **Status:** 0.1.8_ROOT_AMENDMENT_AVAILABLE · INDEPENDENT_CLOSURE_PENDING · DOWNSTREAM_BLOCKED · RUNTIME_NOT_RUN

> **Superseding candidate:** OpenAPI 0.1.8 adds list/detail/count/string/cursor/aggregate
> bounds and compact/non-truncating calendar behavior. This finding remains open until an
> independent exact-SHA review and eventual launcher validation.

## Outcome

Mail 0.1.7 states that pagination uses an opaque cursor with default 50 / maximum 100 and
repeatedly requires bounded collections. The machine contract does not consistently encode
that promise. Several list and detail responses can grow without `maxItems`, pagination or
an aggregate byte ceiling, and some returned strings omit the limits already required on
their corresponding write or persistence path.

## P1 — list routes do not implement the global pagination rule

Only thread/search and CRM activity lists expose the common cursor/limit contract. The
following list responses are arrays without `maxItems` and without cursor pagination:

| Route | Current semantic evidence | Missing machine boundary |
|---|---|---|
| `GET /mail/folders` | at most 100 engine-backed folders per account | `maxItems: 100` |
| `GET /mail/labels` | no creation quota | quota plus bounded page/list |
| `GET /mail/contacts/suggest` | query limit max 30 | response `maxItems: 30` |
| `GET /mail/identities` | no identity quota | quota plus bounded page/list |
| `GET /mail/signatures` | each item may contain two 100,000-character values; no row quota | quota plus pagination/response budget |
| `GET /mail/rules` | each item owns bounded trees, but row count is unbounded | quota plus pagination |
| `GET /admin/mail/domains` | no organization-domain quota | quota plus pagination |
| `GET /admin/mail/accounts/{userId}/identities` | no account identity/alias quota | quota plus pagination |

An admin or ordinary mailbox owner can therefore grow several PostgreSQL-backed resources
and force a single unpaginated response to hydrate all rows. Request item limits elsewhere
do not constrain these reads.

## P1 — nested detail collections and values are not fully bounded

- `MailThreadPage.items` has no `maxItems` despite the request's maximum 100.
- `GET /mail/threads/{threadId}` returns every message summary in one `messages` array with
  no cap or continuation contract; a long-lived thread is unbounded.
- thread participants/labels, message To/CC/folder ids, message attachments, draft
  To/CC/BCC/attachment ids and calendar-event attendees omit response `maxItems` even where
  write-side or domain caps already exist.
- `MailMessageSummary.subject`, `MailMessageDetail.plain_body`/`sanitized_html`, draft
  response bodies and several attachment/calendar response strings omit the corresponding
  maximum length. The engine byte-bound P1 therefore remains visible at the public API.
- `MailThreadPage.next_cursor` omits the common 512-character bound.

For machine-verifiable boundedness, every response collection/value must mirror its
authoritative domain limit. A page count alone is insufficient when each item can contain a
large body; list projections must exclude hydrated content or have an explicit aggregate
byte budget.

## P1 — calendar list expansion has no response ceiling

`GET /calendar/v2/events` accepts an arbitrary `from`/`to` interval and returns full
`CalendarEvent` objects, each potentially containing a 200,000-character description and
200 attendees, in an unpaginated array. The legacy aggregate has an implementation-only
366-day range check but no total event/occurrence ceiling. The legacy scheduled-call list
also has an unbounded top-level call array, although its additive occurrence expansion is
already capped at 1,000.

This permits a large date range, dense recurrence and accumulated rows to produce
unbounded query work and serialization memory. The v2 route should return a compact event
summary page and hydrate one event through its existing detail route. Legacy envelopes must
remain shape-compatible, but need explicit range/cardinality failure semantics rather than
silent truncation.

## Already bounded and not findings

- common request cursor: 512 characters; limit: default 50 / maximum 100;
- bulk mutation request/result: 200 items;
- contact suggestion request: maximum 30;
- draft write recipients: 100 per To/CC/BCC collection; attachments: 50;
- canonical event write attendees: 200; EXDATE values: 1,000;
- scheduled-call additive occurrences: 1,000 across the response by prose contract;
- availability: 31-day/user/interval/global expansion limits in the Phase 147 contract;
- CRM activity response: 100 items plus a 512-character cursor.

## Required amendment shape

1. Add `maxItems`/`maxLength` everywhere an accepted semantic bound already exists.
2. Freeze quotas and cursor pagination for labels, identities, signatures, rules and admin
   domains; enforce create caps before persistence or engine I/O.
3. Paginate thread messages independently; never return an unbounded historical thread.
4. Split canonical calendar list summary from full event detail, bound the date range/page,
   and define non-truncating overflow behavior for legacy aggregate/call envelopes.
5. Add conformance fixtures proving maximum, maximum+1, cursor length and no partial-success
   behavior. OpenAPI examples must validate against the amended schemas.

## Gate impact

- Phase 142 global contract readiness remains reopened;
- affected Phase 143 calendar HTTP/legacy aggregate packets and Phase 144/146 list/detail
  packets must not freeze DTOs against the current shapes;
- G148-02 remains blocked until the amended exact SHA is independently reviewed;
- no production/test source, dependency, validator, runtime or Docker command was changed
  or run by this review.
