# Contract 144-05C — live mail-thread PostgreSQL FTS

**Status:** `ROOT_AUTHORITY_CONTRACT / R2_SCHEMA_ACCEPTED / FTS_PENDING / RUNTIME_NOT_RUN`.

This document is the sole authority for the first Mail thread FTS atom and its
OpenAPI correction. The historical schema-lineage blocker is closed; this
amendment authorizes source-first tests and review plus bounded future
implementation. It grants no behavior, runtime, retained-database, or product
acceptance.

## Historical prerequisite diagnosis — superseded by accepted R2

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

## Prerequisite acceptance amendment (2026-08-29)

This amendment fully supersedes that historical blocker. R2 is accepted at
`b939553c` by the documentation receipt: exactly nine SQL migrations plus
`atlas.sum`, latest `20260829000001`; `lint9`; 9 applied / 0 pending;
snapshot 1 executed / 0 pending; both semantic diffs empty; equal 242505-byte
artifacts; and a raw SHA mismatch that is diagnostic-only. It authorizes only
subsequent FTS migration, tests, repository, OAS, checksum, and verification
atoms. R1 and R2 are immutable. No retained, production, or product-runtime
authority follows.

The prior seven-file diagnosis is historical only and is superseded in full for
current scope, inventory, and acceptance by this amendment's `b939553c` facts.
No historical file count, migration identifier, or conclusion is a current
acceptance input.

## Frozen eventual decision and narrow scope

With the R2 prerequisite accepted, root may select the following atomic FTS packets.
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
5. Docker-only Python integration and API tests for the DB and public behavior boundaries frozen below.

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
and any host execution. Docker execution remains limited to the accepted exact
Compose services and separately reviewed schema-maintenance operations.

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
| `testframework/tests/integration/mail/test_mail_threads_fts.py` | Real PostgreSQL-backed token membership through the authenticated public list route and deterministic QA fixture; no direct database connection and no planner-use assertion. |
| `testframework/tests/api/mail/test_list_mail_threads_search.py` | `GET /mail/threads` search owner scope, starred conjunction, stable order, and absence of `/mail/search`. |
| `backend/internal/qamailcalendar/config.go` | Admit one QA-only `thread-search-v1` scenario under the existing build tags and control-token boundary. |
| `backend/internal/qamailcalendar/http.go` | Route create/delete by the stored scenario and return a dedicated exact search-fixture response; do not weaken the existing closed request parser or expose a release route. |
| `backend/internal/qamailcalendar/controller_thread_search.go` | Own the isolated in-memory lifecycle and exact transaction boundary for the deterministic search fixture. |
| `backend/internal/qamailcalendar/fixture_thread_search_plan.go` | Derive all IDs and values from the reserved nonce and freeze three same-owner live threads plus their messages/folder relations. |
| `backend/internal/qamailcalendar/fixture_thread_search_persist.go` | Persist and verify only that scenario's organization, owner, domain, account, folder, three threads, three messages and three folder relations. |
| `backend/internal/qamailcalendar/controller_thread_search_delete.go` | Delete exactly the three folder relations, messages and threads, then the fixture folder/account/domain/user/organization in one transaction; retain a retryable record on failure. |
| `testframework/src/hub_testkit/domains/mail_calendar_qa/models.py` | Add one strict immutable search-fixture snapshot and strict thread-list projection models; do not relax existing exact payload validation. |
| `testframework/src/hub_testkit/domains/mail_calendar_qa/client.py` | Add create/delete search fixture, authenticated thread-list query, and raw status observation for the absent legacy route; retain bounded no-leak response parsing. |
| `backend/migrations/atlas/mail-calendar-v1/20260829000002_mail_threads_live_subject_snippet_fts.sql` | New forward-only migration containing exactly the frozen single `CREATE INDEX` statement. |
| `backend/migrations/atlas/mail-calendar-v1/atlas.sum` | Add only the accepted FTS migration hash after its approved schema-maintenance operation, on top of the accepted nine-migration R2 state. |
| `backend/docker/entrypoints/mail-calendar-schema.sh` | Explicit FTS transition: 10 SQL / 11 entries / latest `20260829000002` / `20260829000002_current_snapshot.sql` / membership in `run_hash` and `run_verify`; `run_diff` unchanged. |
| `backend/internal/mail/mailbox_projection_handler.go` | Read-only receipt: existing parser accepts one `search`, normalizes via `mailboxThreadProjectionNormalizeSearch`, rejects unknown/duplicate fields. No handler/route mutation. |
| `backend/cmd/server/wiring_calendar_mail.go` | Read-only receipt: existing projection handler registers `/mail/threads` once under authenticated API wiring. No wiring mutation. |
| `frontend/src/api/mail.ts` | Read-only receipt: existing `getMailThreads` emits normalized optional `search` on `/mail/threads`. No new method/endpoint/DTO/field. |

Both Python tests are source-authored before production mutation. Executable
RED/GREEN evidence is Docker-only after candidate infrastructure is available.
The accepted schema-maintenance `run_hash`/`run_verify` path owns the exact
migration/index catalogue receipt. The integration test owns real
PostgreSQL-backed token membership through the public API without a direct DB
driver or planner-use assertion. The API test owns route scope, starred
conjunction, stable order, and absence of `/mail/search`. Static
OAS-parser/wiring/client consistency is a review receipt only.

### Frozen `thread-search-v1` QA fixture

The FTS tests use a dedicated QA-only scenario rather than widening
`inbound-request-v1`. It reuses only the existing token-guarded
`POST /api/_qa/mail-calendar/fixtures` and
`DELETE /api/_qa/mail-calendar/fixtures/:nonce` lifecycle. The scenario owns
its own controller map and never creates a fake-engine message, publishes
NATS, materializes Calendar data, or mutates the inbound invitation fixture.

One creation produces exactly one organization, owner, mail domain, account,
folder and these six live single-message threads in the same account. Values
are deterministic from the returned 32-hex nonce. The two equal-time pinned
IDs are assigned `high` and `low` only after comparing their UUID bytes, so
the expected PostgreSQL `id DESC` order is independent of insertion order:

| Role | normalized subject | snippet | pinned | last message time |
|---|---|---|---:|---|
| pinned-high | `qa <nonce> pinned-high` | `calendar alpha schedule token` | true | fixed base + 2 minutes |
| pinned-low | `calendar <nonce> pinned-low` | `qa beta schedule token` | true | fixed base + 2 minutes |
| recent | `qa <nonce> recent calendar` | `gamma schedule token` | false | fixed base + 3 minutes |
| older | `qa <nonce> older calendar` | `delta schedule token` | false | fixed base |
| qa-only decoy | `qa <nonce> only` | `epsilon schedule token` | false | fixed base + 5 minutes |
| calendar-only decoy | `calendar <nonce> only` | `zeta schedule token` | false | fixed base + 4 minutes |

The response exposes the nonce, scenario, created state, owner email,
organization/owner/account/folder IDs, the four matching thread IDs in exact
order `pinned-high`, `pinned-low`, `recent`, `older`, and the two decoy IDs.
Each message and message-folder ID remains
private to cleanup. Each thread has one live message owned by the same account
and organization; no label, attachment, deletion or Calendar correlation is
created.

The source-first RED query is exactly `search=qa calendar`: no individual
subject or snippet contains that contiguous substring, while `simple`
`plainto_tsquery` matches both tokens across the frozen subject-plus-snippet
document. The `qa`-only and `calendar`-only decoys must be absent, rejecting
token-OR and split-ILIKE fallbacks. The unfiltered search must return the four
matching fixture thread IDs in response order `pinned-high`, `pinned-low`,
`recent`, `older` under the unchanged
`is_pinned DESC, last_message_at DESC, id DESC` contract. The unpinned
`recent` row is newer than both pinned rows, so the assertion also detects a
lost pinned ordering clause. Adding `starred=true` must return exactly the
two pinned IDs in descending UUID order. A second independently
created fixture must not appear in the first owner's authenticated results.
The API route `/api/mail/search` must return `404`; redirects and another
success status are failures.

The existing QA control token is the sole principal for fixture creation and
cleanup; an end-user session is intentionally not a second control-plane
credential. The nonce is generated server-side from `crypto/rand`, encoded as
exactly 32 lower-hex characters, and reserved under one mutex across all three
fixture maps before persistence. The reservation is recorded as `creating`;
a failed create transaction rolls back and removes that exact reservation
under the same mutex before returning the error, while a successful commit
atomically publishes the complete immutable manifest as `created`. It is never
accepted from a create request.
Every route, registration and implementation file remains compiled only under
`qa && !release`; missing/invalid QA configuration registers no surface.

Creation and deletion are fail-closed transactions. The in-memory record owns
an immutable exact-ID manifest and explicit
`creating -> created -> deleting -> deleted` lifecycle. Because this scenario
has no external engine or broker side effect, it has no `cleanup_failed`
state: any database deletion failure rolls the single transaction back and
returns the retained manifest and record to `created` for caller retry;
successful commit removes the record. Unknown nonce, duplicate nonce, partial
persisted set, mismatched plan or wrong state returns conflict/internal error,
never false success. A completed deletion removes the map record; replay then
returns conflict, matching the existing fixture seam rather than hiding a bad
nonce. The controller mutex admits only one mutation per record at a time.

The immutable manifest therefore contains exactly 18 child-row IDs: six
message-folder IDs, six message IDs and six thread IDs. Cleanup deletes only
those recorded IDs in FK-safe order, followed by the exact recorded
folder/account/domain/user/organization IDs, all in one DB transaction. It
never reconstructs scope from nonce, email, subject, folder name or a broad
organization predicate. Delete dispatches by the immutable stored scenario;
it cannot accidentally invoke inbound or account-materialization cleanup.
Because the QA DB and controller are disposable and process-owned, this atom
does not add a persistent lifecycle ledger, TTL service, quota system or
production recovery protocol.

The 10/11 transition is a future conditional obligation after the FTS packet is
accepted; it is not performed, verified, or accepted by this amendment.

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

Hard gate zero is accepted R2 at `b939553c`: exactly nine SQL migrations plus
`atlas.sum`, latest `20260829000001`, `lint9`, 9 applied / 0 pending, snapshot
1 executed / 0 pending, both semantic diffs empty, equal 242505-byte artifacts,
and a diagnostic-only raw SHA mismatch. This permits the FTS atoms; `run_diff`
remains unchanged. The ten-file FTS transition, checksum and verification
remain pending/not run.

After that gate, static acceptance requires the exact migration/path/index
statement; structural identical bound repository predicate; unchanged scope,
optional filters (including already-wired `starred`), `limit + 1` and stable
order; a four-surface static receipt covering OAS, client, parser, and wiring;
Docker-only integration/API acceptance; updated checksum; and independent Contract, Security and Lifecycle review.
R2 schema reconciliation is accepted on the recorded evidence. Every FTS-dependent
requirement, including migration application, repository/query behavior,
OAS/client exposure, and Docker integration/API execution, remains
`PENDING/NOT_RUN`.

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
| Nine-file R2 lineage/guard reconciliation | `PASS (b939553c; Docker schema roundtrip accepted)` |
| Contract rebase | `STATIC_AUTHORED` |
| OAS/backend/wiring/client receipt | `PENDING` |
| Independent Contract/Security/Lifecycle review | `PASS (2026-08-29; 0 critical / 0 important / 0 minor after scoped re-review)` |
| FTS Atlas checksum/verification | `NOT_RUN` |
| FTS Docker schema maintenance, DB apply and plan | `NOT_RUN` |
| Compile, tests, format/lint and runtime | `NOT_RUN` |

The nine-file R2 launcher is accepted for its bounded schema-roundtrip gate.
The ten-file FTS transition, checksum and verification remain `PENDING/NOT_RUN`.
No host tool, retained/shared/production SQL apply, product runtime, merge,
push or deploy authority follows from this document.
