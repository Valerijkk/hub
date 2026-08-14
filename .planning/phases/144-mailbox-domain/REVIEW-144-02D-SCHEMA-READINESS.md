# Phase 144-02D preimplementation review — user configuration schema readiness

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No carrier plan/source/test/runtime authority. Rule Task 1/145-07B await separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness; launcher `NOT_READY`, executables `NOT_RUN`.

> **Detected:** 2026-08-13 root deep static audit
> **Historical status:** ROOT_CORRECTION_CANDIDATE · INDEPENDENT_CLOSURE_PENDING · IMPLEMENTATION_BLOCKED · RUNTIME_NOT_RUN; superseded by the current corrective packet below.
> **Historical downstream lifecycle (2026-08-14):** 02D retains its narrow docs-only `STATIC_ACCEPTED` at the `7d`/`e563` chain. 02F correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact 28-doc transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed. Former unaccepted/transition-review lifecycle and its Rule-plan acceptance instruction are `CONSUMED/HISTORICAL`. No 02D source/schema/runtime authority changes.
> **Scope:** `MailUserSettings`, `MailSignature`, `MailRule` and `RecentContact`

> **CONSUMED/HISTORICAL discovery record:** `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md` supplied
> the root candidate matrix. This former discovery verdict was superseded by the accepted 02F
> docs/readiness receipt above and does not reopen a review or authorize source/schema/runtime work.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept/review/commit instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. The Rule-plan acceptance receipt is the documentation sync; it is not a future Rule Task 4. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

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

The former `f1995465`/`6735ae74` Rule acceptance transition is `CONSUMED/HISTORICAL` and superseded
by the current `669ef761` sole-plan receipt above. This 02D review is not broadened by Rule-plan
acceptance. Rule TDD/production remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated
carrier, G145-5B, G145-5C and launcher-readiness prerequisites, without an implied order.

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
at `fcb5b84f` plus `2f4381a2` is separately `STATIC_ACCEPTED` / `EXACT_SHA_PASS`. The former
Rule-plan authoring/adjudication/acceptance instruction is `CONSUMED/HISTORICAL`. Rule TDD/source
remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated carrier, G145-5B, G145-5C and
launcher-readiness prerequisites, without an implied order.

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
> **Current Rule acceptance:** the current `669ef761` sole-plan receipt above is the only Rule acceptance; it does not close this review or authorize schemas.
> **144-02F sync (historical):** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition text and Rule-plan acceptance instruction are `CONSUMED/HISTORICAL`; no implementation authority. Rule TDD remains `NOT_AUTHORIZED/BLOCKED` under the current receipt's stated prerequisites; 144-01B Docker-blocked, Phase 142/148 pending, launcher `NOT_READY`, executables `NOT_RUN`.
