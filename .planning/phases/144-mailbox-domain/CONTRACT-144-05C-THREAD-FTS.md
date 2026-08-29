# Contract 144-05C — live mail-thread PostgreSQL FTS

**Status:** `ROOT_AUTHORITY_CONTRACT / SCHEMA_LINEAGE_BLOCKED / RUNTIME_NOT_RUN`.

This document is the sole authority for the first Mail thread FTS atom and its
OpenAPI correction. It first blocks all FTS/OAS/repository implementation on
the Mail/Calendar Atlas-lineage reconciliation below. It is docs-only; it is
not source acceptance, schema-apply authority, a compatibility claim, product
readiness, or runtime evidence.

## Mandatory prerequisite — repair the authoritative lineage

The current `mail-calendar-v1` lineage is internally inconsistent: seven SQL
migrations are tracked, but `atlas.sum` contains only the three entries through
`20260826000000_calendar_deferred_receipt_expiry.sql`, while
`backend/docker/entrypoints/mail-calendar-schema.sh` currently exposes only
`run_verify`, which accepts exactly four directory files (three migrations
plus `atlas.sum`) and asserts current version `20260826000000` with three
executed files. That makes its current
schema-maintenance proof non-authoritative for the tracked tree.

**No FTS code, migration, OAS correction, repository predicate, test, Ent
change, or runtime operation may begin until this prerequisite is accepted.**
The first hard packet must reconcile exactly these already tracked post-20260826000000
files into `atlas.sum` and the guard:

1. `20260826000001_calendar_inbound_message_receipt_event_message_idx.sql`
2. `20260827000000_mail_account_password_enrollment.sql`
3. `20260827000001_calendar_delivery_materialization_invalid.sql`
4. `20260827000002_calendar_invite_sequence_non_negative.sql`

Its exclusive ownership is:

| Path | Required prerequisite result |
|---|---|
| `backend/migrations/atlas/mail-calendar-v1/atlas.sum` | Hash all seven tracked migrations in exact version order, with no omitted, invented, reordered or changed historical entry. |
| `backend/docker/entrypoints/mail-calendar-schema.sh` | Add a precise seven-migration plus `atlas.sum` exact-file helper for `run_hash` and `run_verify`, and make those modes' status/current-version/executed-file assertions converge only at `20260827000002` / seven executed / zero pending. Preserve `run_diff` as its intentional historical baseline/core bootstrap-and-diff-only mode; it is not seven-file verification or recovery. Update snapshot/semantic checks only as required for this mode-specific exact-lineage guarantee. |
| `backend/migrations/atlas/mail-calendar-v1/20260826000001_calendar_inbound_message_receipt_event_message_idx.sql` | Read-only existing lineage member; its bytes are re-hashed, not rewritten. |
| `backend/migrations/atlas/mail-calendar-v1/20260827000000_mail_account_password_enrollment.sql` | Read-only existing lineage member; its bytes are re-hashed, not rewritten. |
| `backend/migrations/atlas/mail-calendar-v1/20260827000001_calendar_delivery_materialization_invalid.sql` | Read-only existing lineage member; its bytes are re-hashed, not rewritten. |
| `backend/migrations/atlas/mail-calendar-v1/20260827000002_calendar_invite_sequence_non_negative.sql` | Read-only existing lineage member; its bytes are re-hashed, not rewritten. |

The prerequisite needs independent exact-byte review and the explicitly
approved Docker-only schema-maintenance semantic allowance. Its required
evidence is a source review that proves the seven-file `run_hash`/`run_verify`
set and expected `20260827000002` status, followed by permitted Docker schema
maintenance that proves Atlas hash/validate/apply/status convergence. It must
not relabel `run_diff` as seven-file verification or recovery. Until then all such
evidence is `NOT_RUN`, and the FTS route is `BLOCKED`.

## Frozen eventual decision and narrow scope

Only after the prerequisite is accepted may root select an atomic FTS packet.
Its public surface remains existing authenticated
`GET /mail/threads?search=`, existing
`MailboxThreadProjectionRepository.ListMailboxThreadProjection`, and existing
`MailboxThreadProjectionQuery.Search`. It replaces only the repository's
folded substring predicate with PostgreSQL FTS and corrects OAS to describe
that already-wired route. It creates no handler, DTO, search port, cursor
version/wire field, response field, ranking, or pagination ABI.

The eventual searchable document is exactly the existing live thread's
non-null `normalized_subject` plus non-null `snippet`. The index SQL is frozen
byte-for-byte as one transactional additive statement:

```sql
CREATE INDEX "mail_threads_live_subject_snippet_fts_idx" ON "mail_threads" USING GIN (to_tsvector('simple'::regconfig, (("normalized_subject")::text || ' '::text) || ("snippet")::text)) WHERE ("deleted_at" IS NULL);
```

The repository predicate must use an expression structurally identical to the
index expression and then match with:

```sql
to_tsvector('simple'::regconfig, (("normalized_subject")::text || ' '::text) || ("snippet")::text) @@ plainto_tsquery('simple'::regconfig, $1)
```

`$1` is the existing normalized Search bound as a query parameter, never SQL
interpolation. The repository implementation may use Ent's bound-argument
builder, but it must emit this same cast/concatenation/`regconfig` structure.
There is no `COALESCE` (both fields are NOT NULL), `CONCURRENTLY` (the lineage
is transactional), `to_tsquery`, phrase/prefix syntax, caller dictionary,
ranking, highlight, score, or result snippet.

In scope after prerequisite acceptance:

1. One new forward-only FTS migration and its checksum in the reconciled sole
   `backend/migrations/atlas/mail-calendar-v1/` lineage.
2. One private predicate/helper change in the existing Ent projection
   repository, retaining all existing scope/filter/order clauses.
3. OAS removal of unimplemented `GET /mail/search` / `searchMail`, and
   reconciliation of already-wired optional `search` and `starred` on
   existing `GET /mail/threads` / `listMailThreads`.
4. A static OAS ↔ existing backend parser/wiring ↔ handwritten-client receipt.
5. Focused contract tests for the frozen SQL/binding/scope/order/exclusions.

The OAS `search` is optional, named exactly `search`, with `minLength: 1`
and `x-hub-max-utf8-bytes: 128`. That extension documents the existing
normalized UTF-8 **byte** bound of 1..128; OAS `maxLength` must not claim a
false 128-character limit. It describes only owner-scoped live-thread
normalized-subject plus snippet token search. OAS also declares the already
wired optional `starred` boolean alongside the existing list filters. It must
not expose `q`, `from`, `to`, `date_from`, `date_to`, body, filename,
hydrated-result, or owner-index semantics.

## Explicit exclusions and old routes

Forbidden: Meilisearch or another search store; NATS consumer/checkpoint/replay;
secondary projection; hydration; plain/rich body, HTML, MIME, raw data, BCC,
URL, attachment bytes, names or filenames; remote images; data backfill; new
route/handler, DTO, service, port, client method, OAS endpoint or cursor ABI;
and any host/Docker/runtime execution outside the prerequisite's separately
authorized schema-maintenance allowance.

The old `DISPATCH-REMAINING.md` `144-05C-Search` is non-authoritative wherever
it requires a typed search port, body/filename index document, consumer
checkpoint, hydration, foreign-hit discard, or freshness composition.

`144-05E` owner search-projection hydration remains `BLOCKED / NOT_AUTHORIZED`.
Its body/filename hydration is neither dependency nor implementation detail.
The plain-only 05D content route does not authorize body search.

The current OAS `/mail/search` / `searchMail` is non-authoritative and, after
the prerequisite, must be removed rather than implemented or retained as a
compatibility endpoint. It advertises unimplemented broad From/To/body/
filename/date behavior contradicting this atom. Removal is a safe pre-runtime
source correction, not a deployment or compatibility claim.

## Eventual exact FTS packet ownership

| Path | Ownership after prerequisite acceptance |
|---|---|
| `docs/mail/openapi.yaml` | Remove `/mail/search` GET and `searchMail` from operations. Add optional `search` with `minLength: 1` plus `x-hub-max-utf8-bytes: 128`, and the already-wired optional `starred` boolean, to `/mail/threads`; retain `operationId: listMailThreads`, existing list filters/responses, and state live owner-scoped subject/snippet token semantics. Do not add OAS `maxLength: 128`. |
| `backend/internal/mail/mailbox_ent_projection_repository.go` | Replace only `mailboxThreadProjectionSearchPredicate` and private helper required for the frozen bound FTS predicate. Do not alter scope, filters, mapping, cursor, response or ordering. |
| `backend/tests/mail/mailbox_ent_projection_repository_fts_contract_test.go` | New focused source/contract test: exact structural expression and bound parameter; no `ILIKE`/`COALESCE`/`CONCURRENTLY`/rank or prohibited data/capability expansion; preserved scope/order. |
| `backend/migrations/atlas/mail-calendar-v1/20260829000000_mail_threads_live_subject_snippet_fts.sql` | New forward-only migration containing exactly the frozen single `CREATE INDEX` statement. |
| `backend/migrations/atlas/mail-calendar-v1/atlas.sum` | Add only the accepted FTS migration hash after its approved schema-maintenance operation, on top of the accepted seven-entry prerequisite state. |
| `backend/internal/mail/mailbox_projection_handler.go` | Read-only receipt: existing parser accepts one `search`, normalizes via `mailboxThreadProjectionNormalizeSearch`, rejects unknown/duplicate fields. No handler/route mutation. |
| `backend/cmd/server/wiring_calendar_mail.go` | Read-only receipt: existing projection handler registers `/mail/threads` once under authenticated API wiring. No wiring mutation. |
| `frontend/src/api/mail.ts` | Read-only receipt: existing `getMailThreads` emits normalized optional `search` on `/mail/threads`. No new method/endpoint/DTO/field. |

`backend/internal/mail/mailbox_projection_repository.go` is unchanged: it owns
the existing 128-byte normalized/case-folded Search validation and cursor
filter signature. `backend/ent/schema/mailthread.go` is unchanged.

Ent v0.14.1 cannot represent this expression index. The index is a
**root/founder-approved migration-only exception**, not a speculative Ent
annotation or fake schema field. It additionally requires the schema-
maintenance semantic allowance before generation, hash, lint, application or
verification evidence may be claimed.

## Repository, privacy and order invariants

Retain the locked account proof for `AccountID`, authenticated `ActorID`, and
`OrganizationID`; retain account id, live-thread, owner/org and live-message
fences. FTS never relaxes or replaces these conditions and never triggers an
unscoped hydrate. Folder, label, unread, attachment and starred clauses remain
conjunctive and semantically unchanged. Cursor admission/signature, limit
`1..100` and Search normalization remain unchanged. Ordering remains exactly:

```text
is_pinned DESC, last_message_at DESC, id DESC
```

The existing `limit + 1` continuation and cursor shape are unchanged. No rank,
threshold, total, score or extra cursor value exists; FTS determines membership
before pinned/time/id order.

Caller Search/filter/cursor is untrusted. Bind the FTS value to prevent
injection; do not assemble SQL or accept tsquery operators. Existing validation
and account/thread/message fences remain fail-closed. Empty/nonmatching search
adds no metadata, ID, total, timing classification, hidden-document signal or
error echo. The index may contain only already persisted normalized subject and
snippet: it must not cause body/HTML/MIME/BCC/URL/attachment/filename storage,
logging, output or error leakage.

## Acceptance, rollback and evidence

Hard gate zero is the prerequisite: seven-entry `atlas.sum`, exact
seven-file `run_hash`/`run_verify` guard/status, independent byte review,
and authorized Docker-only semantic schema maintenance must all pass before the
eventual FTS source packet exists. `run_diff` remains its baseline/core
bootstrap-and-diff-only mode and is not this verification/recovery evidence.

After that gate, static acceptance requires the exact migration/path/index
statement; structural identical bound repository predicate; unchanged scope,
optional filters (including already-wired `starred`), `limit + 1` and stable
order; OAS/client/parser/wiring three-way receipt; focused exclusion tests;
updated checksum; and independent Contract, Security and Lifecycle review.
Every requirement remains `PENDING`.

The existing route, DTO, result shape and cursor ABI are retained. Search
matching intentionally changes from substring `ILIKE` to simple FTS under the
same `search` field. `/mail/search` is removed as an unimplemented OAS error,
not retained as compatibility.

Rollback is forward-only. If the new query semantics must be reverted,
repository matching may return to `ILIKE` while the additive FTS index remains.
Dropping `mail_threads_live_subject_snippet_fts_idx` requires a separate,
reviewed destructive forward migration; never rewrite Atlas history, delete
user data or mutate the historical root migration collection.

| Evidence | Status |
|---|---|
| Seven-file lineage/guard reconciliation | `BLOCKED / NOT_RUN` |
| Contract/source inspection | `STATIC_AUTHORED` |
| OAS/backend/wiring/client receipt | `PENDING` |
| Independent Contract/Security/Lifecycle review | `PENDING` |
| Atlas/Ent generation and checksum verification | `NOT_RUN` |
| Docker schema maintenance, DB apply and plan | `NOT_RUN` |
| Compile, tests, format/lint and runtime | `NOT_RUN` |

The accepted launcher remains `NOT_READY`. No host tool, Docker, Ent
generation, SQL apply or runtime authority follows from this document.
