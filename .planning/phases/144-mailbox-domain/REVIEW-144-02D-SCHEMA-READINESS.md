# Phase 144-02D preimplementation review — user configuration schema readiness

> **Detected:** 2026-08-13 root deep static audit
> **Historical status:** ROOT_CORRECTION_CANDIDATE · INDEPENDENT_CLOSURE_PENDING · IMPLEMENTATION_BLOCKED · RUNTIME_NOT_RUN; superseded by the current corrective packet below.
> **Current downstream lifecycle (2026-08-14):** 02D retains its narrow docs-only `STATIC_ACCEPTED` at the `7d`/`e563` chain. 02F correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact 28-doc transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Former unaccepted/transition-review lifecycle is `CONSUMED/HISTORICAL`. No 02D source/schema/runtime authority changes. Root's sole next action is separately accepting the Rule implementation plan; Rule TDD/production also awaits accepted Phase-145 carrier/G145-5B/G145-5C.
> **Scope:** `MailUserSettings`, `MailSignature`, `MailRule` and `RecentContact`

> **CONSUMED/HISTORICAL discovery record:** `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` supplied
> the root candidate matrix. This former discovery verdict was superseded by the accepted 02F
> docs/readiness receipt above and does not reopen a review or authorize source/schema/runtime work.

## Outcome

Public request bounds are mostly present, but the persistence packet still has two competing
default authorities, unprotected rule target references inside JSON, and no storage-retention
bound for recent contacts. Those choices change fields, indexes and transactional behavior.

## P1 — default identity/signature has two canonical authorities

`MailIdentity.is_default` and `MailSignature.is_default` are stored booleans, while
`MailUserSettings` also stores required `default_identity_id` and nullable
`default_signature_id`. PostgreSQL cannot express that each pointer and every per-row flag
always identify the same record. Signature create/update can change `is_default` under the
signature version alone, while settings update uses a different resource version. A crash or
concurrent write can therefore expose two defaults, no default, or a pointer/flag mismatch.

Use one authority. The recommended normalized representation is:

- `MailUserSettings.default_identity_id/default_signature_id` are canonical;
- identity/signature `is_default` is derived in repository DTOs and is not persisted;
- one settings row is unique by account and created atomically when the account activates;
- changing a default locks the settings row, checks same-account/active/send-allowed state,
  changes the pointer and increments settings version in one transaction;
- signature create/update with `is_default=true` performs that same atomic pointer update and
  version increment; a stale settings patch then conflicts rather than overwriting it;
- deleting the selected signature sets the pointer null atomically; deleting the selected
  identity is restricted until the service selects another active send-allowed identity.

Keeping persisted booleans instead requires a database trigger or a complete reviewed
cross-row consistency mechanism; partial unique indexes alone do not make the pointers agree.

## P1 — rule JSON hides referential and concurrency invariants

The OpenAPI freezes 1..20 typed conditions and 1..10 typed actions, but the schema packet
only says bounded typed JSON. `move` and `add_label` targets stored inside JSON have no foreign
key. The already-required label-in-use guard, and the equivalent folder guard needed for a
move rule, can race a concurrent rule write/delete unless all operations share one account
serialization boundary.

The amendment must choose one of two complete representations:

1. normalized action child rows with exact nullable folder/label foreign keys and a named
   action-kind CHECK; or
2. canonical typed JSON plus an account-row lock acquired by rule writes and folder/label
   deletion before same-account target validation.

The second is the smaller v1 default, but it must also freeze:

- exact JSON structs matching the closed OpenAPI union; no `map[string]any`, unknown fields,
  raw Sieve, engine keyword or engine script id;
- serialized byte ceilings in addition to item/value counts;
- stable stored array order and whole-set ordering by `(priority, id)` for deployment hashes;
- soft-delete predicate/indexes and the rule quota from founder question 35;
- `MAIL_FOLDER_IN_USE` as the folder-delete analogue of `MAIL_LABEL_IN_USE`, or an equally
  explicit atomic disable/rewrite policy. Silent dangling targets are forbidden.

## P1 — recent-contact persistence can grow without limit

The suggestion response accepts `limit <= 30`, but each distinct external address can create
a permanent `RecentContact`; neither a per-account cap nor retention/eviction rule is frozen.
Response boundedness does not bound storage, indexes or privacy retention.

The safe unattended default is at most 1,000 recent external contacts per account. Upsert by
`(mail_account_id, normalized_email)`, then evict the least-recently-used non-Hub-linked row
under the same account lock; order eviction by `(last_used_at, id)` for determinism. Account
hard-delete cascades, user deletion sets nullable `user_id` to null, `use_count` is
non-negative/saturating, and email/display name use the accepted 320/160 limits.

## P1 — exact field and conditional storage matrix is incomplete

Minimum remaining schema decisions:

- settings: UUID primary/account/default fields, the two closed enums, timestamps and
  `VersionMixin`; same-account/default eligibility remains a transaction invariant;
- signature: name max 80, required `body_format`, nullable sanitized rich value, required
  server-derived plain value, timestamps and `VersionMixin`; a named CHECK requires rich
  content only for `rich` source and null rich content for `plain` source;
- signature rich/plain values need an explicit UTF-8 byte ceiling in addition to OpenAPI's
  100,000-character limit so combining/multibyte input cannot make storage unbounded;
- rule: name max 120, priority 0..10,000, exact typed/capped condition/action JSON,
  `deleted_at`, timestamps and `VersionMixin`; active ordering is deterministic even when
  priorities match;
- recent contact: required normalized email, bounded nullable display name/user link,
  `last_used_at`, non-negative `use_count`, deterministic eviction index and unique
  account/email index;
- all four account/user/identity/signature/target relations define exact `OnDelete` behavior;
  cross-account ids are rejected before mutation and never repaired by a broad query.

## Gate impact

- 144-02D is `BLOCKED_CONTRACT`; no four-file Ent packet should be authored;
- settings/signature/rule/contact services, rule compiler/deployment and 144-02F remain
  blocked on one default authority, rule-target serialization and contact retention;
- 144-02A must be amended consistently because persisted identity `is_default` participates
  in the same duplicate-authority problem;
- generation/migration and runtime remain independently blocked by founder/launcher gates;
- no production/test source, dependency, generator, validator, migration or Docker command was changed or run.

## 144-03B Rule dependency clarification

Rule contract is docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` accepted at `f199546513b758ca7b95328b5ccd7680009b732e`; acceptance transition/review target is recorded in `6735ae746c7acde657359e8a709cd897c4dce61c`.
This 02D review is not closed or accepted by Rule-contract acceptance. Rule TDD/production requires explicit 02D closure, candidate and
later-plan exact-SHA acceptance, and the separate Phase-145 engine-port carrier.

## Correction-candidate adjudication — 2026-08-14

> **Historical status:** `ROOT_CORRECTION_CANDIDATE` · `INDEPENDENT_CLOSURE_PENDING` · `IMPLEMENTATION_BLOCKED` · `RUNTIME_NOT_RUN`; superseded by the current corrective packet below.

| Prior P1 | Candidate adjudication | Remaining gate |
|---|---|---|
| dual default-signature authority | `CLOSED_IN_CANDIDATE`: settings pointers are sole authorities; identity response derives from `default_identity_id`, signature response post-commit from `default_signature_id`, and only Upsert is a locked command | independent exact-SHA contract/spec, security/adversarial and lifecycle/evidence reviews |
| rule JSON/referential concurrency ambiguity | `CLOSED_IN_CANDIDATE`: flat conjunction and transaction-only cross-row limits are explicit | same three independent exact-SHA reviews |
| field/conditional storage matrix | `CLOSED_IN_CANDIDATE`: four named signature predicates use exact character/`octet_length` bounds | same three independent exact-SHA reviews |

External/static oracle matrix: inspect the four literal signature CHECK names and predicates;
assert create/update/delete pointer truth table, same-account/live and stale-If-Match gates;
assert settings version changes iff pointer changes; assert no second settings PATCH. These are
static candidate oracles only. RecentContact additionally requires exact FK actions, unique/index
and bounds/checks; 999/1,000 boundary; linked exemption; timestamp/raw-UUID LRU tie; concurrent
distinct admission and same-email upsert; INT32 saturation/display bounds; cross-org/inactive/
email-mismatch link rejection; linked-name clearing; multi-account user-delete global lock order,
cap restoration and per-account global LRU; account-delete race; and public projection/privacy/
max-30 separation. Exact oracles include: at 1,000 external plus one linked row, unlink produces
exactly 1,000 and evicts the exact per-account minimum including possibly the just-unlinked row;
user-first then ascending-account lock order over the complete deduplicated owned+linked account
union; reciprocal concurrent delete where U owns A/links B and V owns B/links A locks `{A,B}` in
each transaction before cascade without deadlock/orphan/duplicate/cap breach; link-versus-delete
in a previously unlinked account; delete-first and mutation-first standalone-account races;
deactivate/org/email invalidation; and safe
display-name rejection of invalid UTF-8, CR/LF, Unicode Cc/header injection while preserving
permitted whitespace/Cf/content without trim/NFC/casefold. No schema, generated output, migration,
source, test, launcher or runtime result exists or is authorized.

## Docs-only STATIC_ACCEPTED receipt — 2026-08-14

The exact 144-02D representation/readiness boundary is narrowly docs-only `STATIC_ACCEPTED` /
`EXACT_SHA_PASS`: content commit `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb`
(parent `6c424ed2d6703ceccbb269c37d9abc03e75a0410`, tree
`da3fa6e19316bcf2c7c318a63ec09e1b175f68df`, exact nine documents) and lifecycle transition
`e563a3d3be8898371e5c8a9861c58f11b72167bc` (parent
`7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb`, tree
`b268efdf950609c6d3fa66efe188005d9a39cf40`, exact same nine documents). Post-transition
Contract/Spec/Oracle (`gpt-5.6-terra`/high), Security/Adversarial (`gpt-5.6-terra`/xhigh), and
Lifecycle/Dependency (`gpt-5.6-terra`/high) exact reviews all PASS C/I/M `0/0/0`;
`provenance_source: dispatcher-enforced`, fallback none, telemetry `UNKNOWN` for every route.
The earlier Lifecycle FAIL C/I/M `0/1/0` is `CONSUMED/HISTORICAL` only.

This accepts only the 144-02D static documentation representation/readiness boundary. It grants
no source, Ent schema, generation, migration, service, test/build/lint/validator/Docker/runtime,
phase, product or release authority. Launcher remains `NOT_READY`; executable evidence `NOT_RUN`.
The former direct-closure instruction is `CONSUMED/HISTORICAL`. The 144-02F docs/readiness boundary
at `fcb5b84f` plus `2f4381a2` is separately `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. The active sole
next action is root's separate Rule-plan authoring/adjudication and independent acceptance. Rule
TDD/source additionally requires that plan and Phase-145 engine-port/carrier gates G145-5B/G145-5C.

Historical lifecycle: initial eight-document candidate
`63e419897aa0683e1d3dfcd442aca78eea1b0727` is `CONSUMED/HISTORICAL`; its 02D RecentContact
portion is superseded, never accepted. Lifecycle transition
`6c424ed2d6703ceccbb269c37d9abc03e75a0410` is likewise `CONSUMED/HISTORICAL`; its former
next-review instruction is superseded. Corrective commit
`7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` (parent
`6c424ed2d6703ceccbb269c37d9abc03e75a0410`, tree
`da3fa6e19316bcf2c7c318a63ec09e1b175f68df`, exact nine docs) is
`COMMITTED_UNACCEPTED` / `EXACT_SHA_CLOSURE_IN_PROGRESS`, not accepted at that historical point. Postcommit
Contract/Spec/Oracle (`gpt-5.6-terra`/high) and Security/Adversarial
(`gpt-5.6-terra`/xhigh) exact-SHA reviews PASS C/I/M `0/0/0`. Lifecycle/Dependency
(`gpt-5.6-terra`/high) exact-SHA review FAIL C/I/M `0/1/0` solely because `7d` still described
precommit/commit-next work; content/object boundary passed. All routes are dispatcher-enforced,
fallback none, telemetry `UNKNOWN`. This transition record consumes that stale instruction and
does not claim its own commit SHA.

**Consumed/historical next step:** fresh independent Lifecycle/Dependency exact-SHA re-review of `7d`
plus this transition record; fresh Contract/Security trace-only rechecks are permitted. After all
exact reviews PASS, a separate docs-only 02D `STATIC_ACCEPTED` sync is required. Only then may 02F
closure reviews start. No source, schema, generation, migration, service or runtime authority is
granted; launcher stays `NOT_READY` and executable evidence stays `NOT_RUN`.
> **Current Rule acceptance:** the three Terra routes PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Rule contract acceptance does not close this review or authorize schemas.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text `CONSUMED/HISTORICAL`; no implementation authority. Root next: accept separate Rule plan only; Rule TDD also awaits Phase-145/G145-5B/G145-5C; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
