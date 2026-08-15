# Phase 144-02B preimplementation review — projection schema readiness

> **Candidate review gate:** `ROOT_SOURCE_AUTHORITY_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED` / `IMPLEMENTATION_BLOCKED` / `RUNTIME_NOT_RUN`, no source dispatch/acceptance. Review frozen test-first exact13: new `backend/ent/schema/mailfolder.go`, `backend/ent/schema/mailthread.go`, `backend/ent/schema/mailmessage.go`, `backend/ent/schema/mailmessagefolder.go`, `backend/ent/schema/maillabel.go`, `backend/ent/schema/mailmessagelabel.go`, `backend/tests/mail/mail_projection_schema_test.go`; edits `backend/ent/schema/mailaccount.go`, `backend/ent/schema/calendarevent.go`, `backend/ent/schema/mailmutationreceipt.go`, `backend/tests/mail/mail_account_identity_schema_test.go`, `backend/tests/mail/mail_mutation_receipt_schema_test.go`, `backend/tests/calendar/calendar_event_schema_test.go`.
>
> Require PERSISTENCE 02B tables/checks/indexes/JSON; sole 02F-R exception is `MailMutationReceipt.folders` inverse for `MailFolder.control_receipt_id -> receipt RESTRICT`. Actor/message scalar-only and all other receipt edges/checks unchanged; `MailAccount` folders/threads/messages/labels; `CalendarEvent.mail_messages` inverse with Message FK `SET NULL`; rule targets service-only, no DB FK/normalized target; compatibility tests mandatory. Contract/high, Security/xhigh and Lifecycle/high each PASS C/I/M `0/0/0`, then exact-nine-doc commit, three fresh exact-SHA reviews and separate docs-only acceptance precede root dispatch. 02C waits accepted 02B; 02F-D/02E/Ent/generation/migration/runtime/services/API/UI/release BLOCKED; launcher `NOT_READY`, executables `NOT_RUN`.

> **Postcommit correction lifecycle:** `45fd3369` is committed-unaccepted: exact-SHA reviews FAIL C/I/M Contract `0/1/0`, Security `0/1/1` for manifest-path and historical-wording defects only. The current exact2 correction candidate is `PRECOMMIT_REVIEW_REQUIRED`; no source dispatch follows unless its fresh review lifecycle passes.

> **2026-08-15 lifecycle correction acceptance:** exact14 is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` (parent `0ef3df9eedfb3683fe70a7f8fc502ca76278facb`, tree `b28904b96c084dd88c4cec37c8e1daa62fde4c5d`) after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high exact-SHA PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback, telemetry `UNKNOWN`. Prior candidate/precommit/review clauses only are `CONSUMED/HISTORICAL`; the current 02B exact13 gated authority candidate remains live. Engine deletion-port and User deletion-intent remain separately `BLOCKED` / `NOT_DISPATCHABLE`; unchanged source/generation/migration/launcher blockers remain current.

> **Detected:** 2026-08-13 root deep static audit
> **Status:** exact14 docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` · IMPLEMENTATION/SOURCE/SCHEMA/RUNTIME `BLOCKED` · launcher `NOT_READY` · executables `NOT_RUN`
> **Scope:** folder, thread, message, folder/message and label/message projection schemas

> `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` supplies the accepted exact14 docs representation.
> The lower discovery verdict and its former closure/open-until-review framing and all
> P1/outcome/gate-impact clauses are `CONSUMED/HISTORICAL` as readiness adjudication; they
> preserve discovery provenance only and grant no implementation, source, schema or runtime authority.

> **2026-08-13 source-only follow-up:** a requested Terra/high review whose actual runtime
> provenance was unavailable found that `EngineMessageProjection.Version` had no durable
> `MailMessage` field. The root amendment now requires `engine_version` 1..1,024 bytes and
> same-projection atomic persistence; this lower source-only remediation narrative is
> `CONSUMED/HISTORICAL`, not a current acceptance or implementation route.

## Outcome

The lifecycle and ownership model are coherent, but the packet is not yet deterministic at
the database boundary. It requires exact partial uniqueness and bounded values that cannot
be represented from the currently listed fields without choosing additional normalized
columns, storage types and relational checks.

## P1 — case-insensitive uniqueness has no storage representation

`MailFolder` requires custom sibling names to be unique case-insensitively, and `MailLabel`
requires the same rule per account. Neither entity lists a normalized-name column, while the
contract also does not approve PostgreSQL `citext` or a functional lower-expression index.
An ordinary Ent field index cannot express the promised canonical uniqueness by itself.

The safe implementation-neutral amendment is:

- add a private required `normalized_name` column produced by the one accepted Unicode/name
  normalizer;
- keep the display `name` separately bounded;
- use two folder partial unique indexes for non-deleted custom rows: root
  `(mail_account_id, normalized_name)` where `parent_id IS NULL`, and child
  `(mail_account_id, parent_id, normalized_name)` where `parent_id IS NOT NULL`;
- use one label unique index `(mail_account_id, normalized_name)`;
- never depend on database/session collation or an unapproved extension.

The exact predicates must also say whether system-role folders participate. The contract
says custom sibling names, so the recommended predicate excludes non-null `role`; a
separate partial unique `(mail_account_id, role)` should enforce at most one live folder per
system role.

## P1 — bounded projection values lack exact hard limits

The following fields are called bounded but have no complete machine/storage ceiling:

- thread normalized subject and participant-address JSON;
- message `internet_message_id`, `in_reply_to`, `references`, structured address JSON,
  engine version, auxiliary engine state and snippets;
- folder engine ids/control links and label engine keywords;
- folder/thread/message/join rebuild markers where relevant.

The accepted engine contract already supports 1,024-byte opaque ids and mandatory versions,
4,096-byte state/cursor values, 200 addresses, 100 folders and 50 attachments. The write API also freezes a 998-byte
subject transport ceiling. The amendment should bind stored values to those limits, cap
references explicitly, and combine the 400-grapheme snippet rule with an independent UTF-8
byte ceiling. Grapheme count alone is not a memory bound because one grapheme may contain an
arbitrary combining sequence.

Recommended technical defaults are 200 unique participant/address items, 100 reference
Message-IDs, 998 bytes per Message-ID/subject, 4,096 UTF-8 bytes plus 400 graphemes per
snippet, 1,024 bytes per engine id/keyword and 4,096 bytes per opaque engine state.

## P1 — join and state relational checks are not frozen

- `MailMessageFolder` and `MailMessageLabel` do not specify an explicit UUID id versus an
  implicit surrogate/composite key, nor exact delete behavior. Ent still needs one stable
  schema representation even though the id is not public.
- `MailMessageLabel` requires `manual OR engine_rule`; this must be a named database CHECK,
  not only service validation.
- folder `engine_folder_id` is nullable for an unresolved create, while `creating`,
  `unknown` and `error` can all represent that lifecycle. The exact allowed
  `control_state`/engine-id/receipt combinations are not tabulated.
- system-role uniqueness, non-negative counters/epochs and immutable
  `received_event_id` need explicit index/check decisions. A stable application event marker
  should be globally unique if it is relied on for one-event admission.
- same-account joins cannot be enforced from the two foreign keys alone. The contract must
  continue to state that service/repository scope proves account equality before insert;
  adding duplicated account ids solely for a CHECK is not justified without a reviewed
  consistency mechanism.

Recommended representation: UUID surrogate ids for both join schemas, unique business-key
indexes, cascade only from a hard-deleted join owner, and no new duplicated account column.
Soft-deleted projections remain filtered by their owning rows.

## Dependencies already open

This packet also inherits the unresolved account schema, public list quotas/pagination and
thread-detail response bounds recorded by 144-02A and the Phase 142 OpenAPI response audit.
Those are not duplicated as new findings here.

## Gate impact

- 144-02B `BLOCKED_CONTRACT` is a `CONSUMED/HISTORICAL` discovery verdict, superseded by the current exact13 gated authority candidate; no current source dispatch is authorized until its required lifecycle completes;
- message/folder/label repositories, sync apply, bulk mutation and public DTO packets remain
  blocked on the reviewed field/index/check matrix;
- generation/migration and runtime remain independently blocked by founder/launcher gates;
- no production/test source, dependency, generator, validator, migration or Docker command
  was changed or run.
