# Phase 144-02D preimplementation review — user configuration schema readiness

> **Detected:** 2026-08-13 root deep static audit
> **Status:** CONTRACT_AMENDMENT_REQUIRED · IMPLEMENTATION_BLOCKED · RUNTIME_NOT_RUN
> **Scope:** `MailUserSettings`, `MailSignature`, `MailRule` and `RecentContact`

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
- no production/test source, dependency, generator, validator, migration or Docker command
  was changed or run.
