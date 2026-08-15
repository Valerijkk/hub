# Mail persistence contract 0.1.8

> **Phase-145 carrier-plan acceptance (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787`/sync `375891bb` remain accepted. Only obsolete carrier-plan-pending clauses are `CONSUMED/HISTORICAL`; unrelated gates remain current. No schema/source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher; carrier source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Phase-145 carrier receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Base/correction lifecycle `CONSUMED/HISTORICAL`; no plan/source/test/runtime authority. Separate carrier-plan acceptance + G145-5B/G145-5C/launcher readiness keep Rule Task 1/145-07B blocked; launcher `NOT_READY`, executables `NOT_RUN`.

> **Status:** 144-02A frozen A1/A2/A3 representation is historical docs-only accepted at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931`; the current exact-14 144-02A/B/C/D correction candidate below amends its deletion edges/lifecycle and requires new review. 02F remains docs/readiness-only accepted and is a dependency trace, not part of the correction. This is not source, TDD, generated output, migration, launcher, test/validator/build/runtime, product/module/phase/release acceptance.
> **144-02D/02F current:** the 02D base remains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus `e563a3d3be8898371e5c8a9861c58f11b72167bc`, while its affected User/RecentContact lifecycle is amended only by the current unaccepted A/B/C/D candidate. 02F `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` remains docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and Rule-plan-next instructions are `CONSUMED/HISTORICAL`. All source/Ent/generation/migration/runtime authority remains blocked; accepted Rule plan `669ef761` leaves Rule Task 1 `NOT_AUTHORIZED/BLOCKED` on independent Phase-145 carrier/G145-5B/G145-5C/launcher readiness.
> **Runtime/generation/migration:** NOT_RUN / NOT_RUN / NOT_AUTHORIZED
> **Scope:** Phase 144-02A/B/C/D correction candidate plus accepted-02F dependency trace only
> **Precedence:** this document supplies exact storage representation for the corresponding
> domain sections in `MAIL-SPEC.md`. It does not authorize generated code or migration SQL.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Common representation

- Every primary key is immutable UUID with application default `uuid.New`.
- Every owned foreign-key UUID is an explicit field used by its Ent edge. Service/repository
  code proves same-account/same-organization relations before insert when a row-local CHECK
  cannot traverse a foreign key.
- Strings use the stated character limit plus an application UTF-8 byte validator. Opaque
  engine values additionally use a named PostgreSQL `octet_length(...)` CHECK.
- Every SHA-256 value is `bytea` of exactly 32 bytes. A versioned hash is either one required
  `(version,hash)` tuple or one all-null optional tuple; cross-version equality is forbidden.
- Attempt/counter/epoch fields are 0..2,147,483,647 unless a narrower bound is stated.
  Generations are positive signed int64. Timestamps are UTC.
- An unqualified `string` below is an Ent string with both a rune/character validator where
  stated and an `octet_length` CHECK for the stated UTF-8 byte ceiling. `bytes(32)` is Ent
  bytes plus an exact-length CHECK. A nullable scalar uses `Optional().Nillable()`; a
  required boolean defaults only when the value is named below.
- `created_at` is required/default-now/immutable, `updated_at` is required/default-now with
  update-default-now, and `deleted_at` is nullable. Every UUID primary key is
  required/default-`uuid.New`/immutable; every stable engine or idempotency key is immutable.
- Versioned public resources use `VersionMixin`; no second hand-written `version` field.
- Soft-delete uniqueness always includes `deleted_at IS NULL`. Hard account deletion, engine
  removal, object drain and physical Organization completion are BLOCKED pending the separate
  accepted engine deletion-port amendment and separate accepted source plan; this candidate
  supplies only the fail-closed representation.
  Historical snapshot ids that must survive projection deletion are scalar UUIDs, not cascading edges.
- `normalize_mail_name_v1` is a private key normalizer; no database/session collation
  participates. For the Folder consumer, validate raw valid UTF-8 with no Unicode control/NUL
  and raw 1..120 runes/<=480 bytes; display name preserves exact raw bytes/runes. Its key
  collapses every `unicode.IsSpace` run to one ASCII space and trims, then applies NFC,
  `cases.Fold` using `golang.org/x/text v0.35.0`, then final NFC; the key is nonempty and
  <=480 bytes. Folder raw validation precedes its private key normalization; key validation
  follows it. Label consumers retain their separately documented raw bounds and use this
  private/no-collation normalization algorithm only after those bounds. Any Unicode/x-text
  version change requires an accepted migration/collision plan.
- In every canonical binary encoding below, `U8(n)` is one unsigned octet and `U32BE(n)` is
  one unsigned 32-bit big-endian integer. Every encoded length/count is range-checked before
  conversion; concatenation never uses native integer width, native endianness or signed casts.
- No schema below contains raw HTTP idempotency keys, password/token/credential/private DKIM,
  provider error text, raw Sieve, message body, raw MIME, attachment bytes or project ACL.

## 144-02A — control plane

### `MailAccount`

Exact fields (all unlisted defaults are absent):

| Fields | Storage / nullability / default |
|---|---|
| `id` | UUID primary |
| `organization_id`, `user_id`, `mail_domain_id` | required UUID; immutable owner/domain selection |
| `engine_provisioning_key` | required immutable string, 1..1,024 bytes |
| `local_part` | required immutable normalized ASCII string, 1..64 bytes |
| `engine_account_id` | nullable string, 1..1,024 bytes when present |
| `desired_enabled` | required bool, default `true`; canonical engine enablement intent |
| `status` | required internal enum, default `provisioning`; includes public metadata-only non-interactive `deleting` state |
| `control_generation`, `control_hash_version`, `control_hash` | required int64 >=1, int32 >=1, bytes(32); generation defaults 1 and hash version defaults 1 |
| `attempted_control_generation`, `attempted_control_hash_version`, `attempted_control_hash` | nullable int64/int32/bytes(32) tuple |
| `deployed_generation`, `deployed_hash_version`, `deployed_hash` | nullable int64/int32/bytes(32) tuple |
| `projection_state` | required enum, default `bootstrapping` |
| `sync_cursor`, `rebuild_checkpoint`, `rebuild_base_cursor` | nullable strings, each 1..4,096 bytes when present |
| `sync_epoch`, `mutation_epoch` | required int64, default 0, bounded non-negative |
| `provision_attempts`, `sync_attempts` | required int32, default 0, bounded non-negative |
| `provision_lease_token`, `sync_lease_token`, `mutation_lease_token` | nullable UUIDs |
| `provision_lease_until`, `sync_lease_until`, `mutation_lease_until` | nullable UTC timestamps paired with the corresponding token |
| `next_provision_at`, `next_sync_at`, `last_synced_at` | nullable UTC timestamps |
| `provision_error_code`, `sync_error_code` | nullable closed enums below |
| `event_cutoff_at` | required immutable UTC timestamp |
| `created_at`, `updated_at` | common timestamps |

Enums:

- status: `provisioning|active|suspended|error|deleting`;
- projection: `bootstrapping|ready|degraded|rebuilding`;
- provision error: `MAILBOX_ENGINE_DISABLED|MAILBOX_DOMAIN_REQUIRED|MAILBOX_PROVISIONING_RETRY|MAILBOX_SUSPENDED|MAILBOX_UNAVAILABLE`;
- sync error: `MAIL_SYNC_ENGINE_UNAVAILABLE|MAIL_SYNC_CURSOR_INVALID|MAIL_SYNC_RESTART_REQUIRED|MAIL_SYNC_PAGE_REJECTED|MAIL_SYNC_PROJECTION_LIMIT|MAIL_SYNC_CONTROL_CONFLICT`.

Indexes/checks:

- unique `user_id`; unique `engine_provisioning_key`; unique nullable
  `engine_account_id`; unique `(mail_domain_id,local_part)`;
- each lease token/until pair is both-null or both-present;
- attempted/deployed generation/version/hash tuples are each all-null or all-present,
  have positive version/generation and generation <= `control_generation`; no hash is
  compared across versions;
- `active` requires `desired_enabled=true`; `suspended` requires
  `desired_enabled=false`. Both require an engine id and a deployed tuple, but an active account may
  intentionally have `deployed_generation < control_generation` while an alias/lifecycle
  change is pending; exact equality means converged, not merely active;
- `deleting` is the blocked-pending-engine-port representation only. Its schema shape requires
  `desired_enabled=false`, no lease/attempt/confirmed absence and the existing due marker; it does
  not claim engine absence, authorize a drain, map to `suspended` publicly, or make a worker
  executable.
- when no provision lease/attempt is retained and deployed is absent or differs from current,
  `next_provision_at` is non-null. A worker that applies an older captured tuple while current
  has advanced clears the attempt, sets `status=provisioning`, and reschedules; it never marks
  the newer desired enablement active/suspended from the stale result;
- a live provision lease requires the attempted tuple. After indeterminate I/O the lease is
  cleared while the tuple is retained with `MAILBOX_PROVISIONING_RETRY|MAILBOX_UNAVAILABLE`
  and non-null `next_provision_at`; definitive failure clears attempted before another try;
- `status=error` requires a provision error. `projection_state=ready` requires non-null
  `sync_cursor` and `last_synced_at`, null rebuild fields and null sync error;
  `projection_state=rebuilding` requires `rebuild_base_cursor` while checkpoint is nullable
  before the first committed page;
- list/control indexes cover status+next due time and projection state+next sync time; exact
  blocked-observability-only index `mail_accounts_deleting_blocked_idx (next_provision_at,id)
  WHERE status = 'deleting'` exposes only blocked-pending rows on their existing due field.
  It is neither a claim/recovery/drain queue nor authority for destructive work.

All ordinary control workers persist the captured attempted generation/version/hash on the owner row
before I/O; a newer current generation cannot rewrite it or be finalized by the older result.
Successful completion copies exactly attempted to deployed, clears attempted/lease/error,
and reschedules when deployed is older than current. Ordinary provisioning, synchronization,
projection, mutation, configuration and reactivation workers require `status <> 'deleting'`
both at claim and exact token/generation/epoch completion. A late completion updates zero rows
and causes no compensating I/O. The deleting representation has no claim path: its exact
blocked-pending shape has no lease, attempt or confirmed-absence tuple, and no ordinary worker
may claim or finalize it. `Organization -> MailAccount` and `User -> MailAccount` are `NO ACTION`;
they do not authorize physical deletion, precleanup, drain, or partial mutation.

Account-control hash v1 is SHA-256 of
`ASCII("hub.mail.account.control") || 0x00 || 0x01 || bool(desired_enabled) ||
str(local_part) || str(normalized_domain) || U32BE(identity_count) || identities`.
`bool(false)=0x00`, `bool(true)=0x01`, and `str` is
`U32BE(UTF-8-byte-length)||UTF-8`. Each identity record is
`U32BE(record_length)||U8(kind)||str(normalized_address)`, with primary=1/alias=2.
The vector contains exactly one primary plus every non-deleting alias desired in the engine,
sorted first by kind then normalized-address bytes; group rows and display names/engine ids
are excluded. The generation is stored beside but not inside the hash, so identical desired
specs may have identical hashes at different generations. Hash version 1 means this exact
encoding; no JSON/map/locale/collation participates.

### `MailIdentity`

Fields: `id`, `mail_account_id`, `normalized_address` (1..320 bytes), optional
`display_name` (<=160 chars/640 bytes), required `kind`, required `send_allowed` default
`false`, required `status` default `pending`, optional `engine_identity_id` (<=1,024 bytes),
`created_at`, `updated_at`. `is_default` is not stored.

Enums: kind `primary|alias|group`; status `pending|active|error|deleting`.

Indexes/checks: global unique normalized address; unique nullable
`(mail_account_id,engine_identity_id)`; unique partial `(mail_account_id)` where
`kind='primary'`; exactly one primary is created with the account under the same transaction;
at most 200 rows/account is enforced under account lock. Only `status=active` may have
`send_allowed=true`; group rows always have it false and are not selectable. Primary
hard-delete is restricted until its account is deleted; only aliases may enter `deleting`,
and confirmed-absent aliases may hard-delete. Public default is derived from settings.

Selected Stalwart 0.16.17 management `x:Account` represents the address set inline, while
mailbox RFC 8621 `Identity/get|set` supplies the stable server identity ids persisted here.
Packet 145-03A must reconcile exactly one id per normalized desired address through
Get-before-create/Get-before-retry. Duplicate or unexpected identities are degraded rather
than selected by list position, and no adapter-private id is fabricated. Enabled use remains
blocked by G145-5B mailbox authentication.

### `MailDomain`

| Fields | Storage / nullability / default |
|---|---|
| `id`, `organization_id` | UUID primary and required immutable owner UUID |
| `domain` | required immutable normalized ASCII string, 1..253 bytes |
| `engine_provisioning_key` | required immutable string, 1..1,024 bytes |
| `engine_domain_id` | nullable string, 1..1,024 bytes when present |
| `desired_present` | required bool, default `true`; false only for canonical deletion intent |
| `enabled`, `is_default` | required bools, defaults `true` / `false` |
| `state` | required enum, default `pending` |
| `verification_status`, `dkim_status` | required enums, defaults `pending` / `unknown` |
| `required_dns_records` | required concrete JSON array, default empty, bounded below |
| `desired_generation`, `desired_hash_version`, `desired_hash` | required int64 >=1, int32 >=1, bytes(32); generation/version default 1 |
| `attempted_generation`, `attempted_hash_version`, `attempted_hash` | nullable int64/int32/bytes(32) tuple |
| `deployed_generation`, `deployed_hash_version`, `deployed_hash` | nullable int64/int32/bytes(32) tuple |
| `last_error_code`, `next_retry_at` | nullable closed enum / UTC timestamp |
| `attempt_count` | required int32, default 0, bounded non-negative |
| `lease_token`, `lease_until` | nullable UUID/UTC timestamp pair |
| `created_at`, `updated_at`, `deleted_at`, `VersionMixin` | common timestamps, nullable soft delete and positive version |

DNS JSON is a typed array of at most 32 unique `(type,name,value)` records: type
`MX|TXT|CNAME`, name <=253, value <=2,048, TTL 60..86,400; serialized UTF-8 <=65,536 bytes.

Selected Stalwart 0.16.17 exposes expected `dnsZoneFile` but no manual-DNS verification
result, and that zone may contain record classes outside this enum. Until
G145-5C/founder questions 43 and 44 freeze the authority and exact projection, the adapter
may persist only a provisional bounded supported subset in `dns_required`; it may never
enter `verified` or call account provisioning from domain existence, `isEnabled` or zone
text alone.

Checks/indexes: global unique live domain; unique provisioning key; unique nullable engine
id; one live verified+enabled default per organization; attempted/deployed tuples all-or-none,
positive and no newer than desired; lease pair all-or-none. The state matrix is:

| State | Required | Forbidden / additional rule |
|---|---|---|
| `pending` | `desired_present=true`, desired tuple | attempted tuple, lease and error are null; deployed may be older/null |
| `provisioning` | `desired_present=true`, attempted tuple equal to a captured desired tuple and live lease pair | error/retry null; engine id may still be null |
| `dns_required` | `desired_present=true`, engine id, deployed tuple equal to current desired, non-empty DNS records | attempted tuple/lease/error/retry null; verification is not `verified` |
| `verified` | `desired_present=true`, engine id, deployed tuple equal current desired, `verification_status=verified` | attempted tuple/lease/error/retry null |
| `error` | `desired_present=true`, `last_error_code`; unknown outcome additionally requires attempted tuple and retry time | lease null; a definitive error clears attempted; deployed never advances from an unproved result |
| `deleting` | `desired_present=false`, `enabled=false`; any in-flight engine delete follows the provisioning/unknown tuple rules | `is_default=false`; retry/error obey the same allowlist, and `deleted_at` remains null until confirmed absent |

`last_error_code` is exactly
`MAIL_DOMAIN_ENGINE_UNAVAILABLE|MAIL_DOMAIN_APPLY_REJECTED|MAIL_DOMAIN_OUTCOME_UNKNOWN|MAIL_DOMAIN_DNS_INVALID`;
only `MAIL_DOMAIN_OUTCOME_UNKNOWN` may retain an attempted tuple after lease release. A
non-null `deleted_at` requires deleting, no account reference, confirmed engine absence,
and null `engine_domain_id`, attempted tuple, lease, error and retry fields.
`is_default=true` additionally requires desired-present, enabled, verified state and verified
verification status. If current desired advances while an older apply succeeds, the row
returns to pending instead of claiming `dns_required|verified` for stale desired state.
Account references restrict deletion. At most 100 live domains/organization under the
organization lock; the quota failure is stored terminally in its client receipt before any
domain or engine side effect.

Domain-control hash v1 is SHA-256 of
`ASCII("hub.mail.domain.control") || 0x00 || 0x01 || bool(desired_present) ||
bool(enabled) || str(domain)`, using the same boolean/string encodings as account control.
Deletion sets `desired_present=false`, `enabled=false` and `is_default=false` in the same
versioned transaction. Generation is stored beside, not inside, the hash.

### `MailboxPolicy`

One `id`/`organization_id` row with required `external_mail_enabled`, `max_message_bytes`
1,048,576..52,428,800, `max_attachment_bytes` 1..26,214,400, `max_recipients` 1..300,
`send_rate_per_minute` 1..1,000, trash/junk retention 1..3,650 days, timestamps and
`VersionMixin`. All policy values are required with no schema default; the provisioning/admin
service supplies the configured initial values explicitly. MailboxPolicy creation supplies
`max_message_bytes=25,000,000` and `max_attachment_bytes=25,000,000` decimal bytes. These
product defaults are distinct from the engine adapter hard caps of 52,428,800 message bytes
and 26,214,400 attachment bytes. Administrators may configure each value only within its
stated inclusive bound, with attachment <= message. Unique organization; CHECK attachment <=
message. `MailboxPolicy` may cascade only after the Organization mail-deletion gate below
has proved zero remaining `MailAccount` and `MailObject` rows. Enabling external mail still
requires the service-level verified-domain/engine gate.

### 144-02A Ent representation freeze

This subsection closes the remaining Go/Ent representation choices for 144-02A and is narrowly
docs-only accepted at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931`. It does not
authorize Ent source or TDD, generation, migration, launcher execution, tests/validators/build
or runtime; founder migration authority, supported task-owned launcher, sequential explicit source
dispatch and source-level review remain required.

#### Scalar representation and validation

- Generations and epochs use `field.Int64`; hash versions and attempt counters use
  `field.Int32`. Policy byte ceilings use `field.Int64`; policy recipient/rate/retention
  values use `field.Int32`.
- Every SHA-256 field uses `field.Bytes(...).MinLen(32).MaxLen(32)`. Nullable hashes are
  additionally `Optional().Nillable()`.
- A string bound stated in bytes is implemented by Ent `MinLen`/`MaxLen`; Ent v0.14.1 uses
  Go `len(string)`, so the validator is byte-counted. `display_name` additionally rejects
  more than 160 Unicode code points and invalid UTF-8.
- Normalized domain, local-part and address columns accept only already-normalized lower-case
  ASCII. Validators reject rather than transform. A domain is one or more dot-separated
  labels of 1..63 bytes using `a-z`, `0-9` and interior `-`, with no trailing dot. A local
  part is one dot-atom of 1..64 bytes using lower-case RFC 5322 `atext`, with no leading,
  trailing or repeated dot. A normalized address contains exactly one `@` and reuses those
  two validators. IDNA conversion and user-input normalization remain Phase 144-03/service
  responsibilities; persistence never silently canonicalizes.
- `verification_status` is exactly `pending|verified|failed`; `dkim_status` is exactly
  `unknown|pending|valid|invalid`.
- `MailIdentity.mail_account_id`, `normalized_address` and `kind` are immutable; identity
  address/kind changes use create/delete lifecycle instead of in-place identity mutation.
  `MailboxPolicy.organization_id` is immutable. The account/domain owner fields already
  declared immutable remain Ent-immutable.
- `MailboxPolicy` has the exact fields `id`, `organization_id`,
  `external_mail_enabled`, `max_message_bytes`, `max_attachment_bytes`, `max_recipients`,
  `send_rate_per_minute`, `trash_retention_days`, `junk_retention_days`, `created_at`,
  `updated_at`, plus `VersionMixin`. None of its policy fields has an Ent default.

`required_dns_records` uses the non-provider persistence value
`mail.RequiredDNSRecords` from `backend/internal/mail/domain_dns.go`:

```go
type RequiredDNSRecord struct {
	Type  EngineDNSRecordType `json:"type"`
	Name  string              `json:"name"`
	Value string              `json:"value"`
	TTL   int                 `json:"ttl"`
}

type RequiredDNSRecords []RequiredDNSRecord
```

`RequiredDNSRecords.Validate` rejects a nil slice, more than 32 records, invalid UTF-8, a name over 253
bytes, a value over 2,048 bytes, a type outside `MX|TXT|CNAME`, TTL outside 60..86,400,
duplicate `(type,name,value)` triples regardless of TTL, JSON marshal failure, or canonical
serialized JSON over 65,536 bytes. Empty strings remain representable because the public
0.1.8 schema gives them no minimum; provider/service validation decides whether a concrete
record is actionable. The Ent field is required and defaults to a non-nil empty
`mail.RequiredDNSRecords{}`.

#### Exact indexes

The exact ordered index definitions are:

- `MailAccount`: unique `user_id`; unique `engine_provisioning_key`; unique
  `engine_account_id` with predicate `engine_account_id IS NOT NULL`; unique
  `(mail_domain_id,local_part)`; lookup `(status,next_provision_at)`; lookup
  `(projection_state,next_sync_at)`; `mail_accounts_deleting_blocked_idx (next_provision_at,id)
  WHERE status = 'deleting'` (blocked observability only, never a recovery/claim queue).
- `MailIdentity`: global unique `normalized_address`; unique
  `(mail_account_id,engine_identity_id)` with predicate
  `engine_identity_id IS NOT NULL`; unique `mail_account_id` with predicate
  `kind = 'primary'`; lookup `(mail_account_id,status)`.
- `MailDomain`: unique live `domain` with predicate `deleted_at IS NULL`; unique
  `engine_provisioning_key`; unique `engine_domain_id` with predicate
  `engine_domain_id IS NOT NULL`; unique `organization_id` with predicate
  `deleted_at IS NULL AND is_default = TRUE`; lookup `(state,next_retry_at)`.
- `MailboxPolicy`: unique `organization_id`.

The service transaction still owns exactly-one-primary creation, 200 identities/account,
100 live domains/organization, same-account/default eligibility and cross-row domain-use
checks; an Ent row CHECK cannot prove those cross-row facts.

#### Exact named row checks

`MailAccount.Annotations` contains these exact names and predicates:

```text
mail_accounts_control_hash:
  control_generation >= 1 AND control_hash_version >= 1 AND length(control_hash) = 32
mail_accounts_attempted_control_tuple:
  ((attempted_control_generation IS NULL) = (attempted_control_hash_version IS NULL)) AND ((attempted_control_generation IS NULL) = (attempted_control_hash IS NULL)) AND (attempted_control_generation IS NULL OR (attempted_control_generation >= 1 AND attempted_control_hash_version >= 1 AND length(attempted_control_hash) = 32 AND attempted_control_generation <= control_generation))
mail_accounts_deployed_control_tuple:
  ((deployed_generation IS NULL) = (deployed_hash_version IS NULL)) AND ((deployed_generation IS NULL) = (deployed_hash IS NULL)) AND (deployed_generation IS NULL OR (deployed_generation >= 1 AND deployed_hash_version >= 1 AND length(deployed_hash) = 32 AND deployed_generation <= control_generation))
mail_accounts_provision_lease_pair:
  (provision_lease_token IS NULL) = (provision_lease_until IS NULL)
mail_accounts_sync_lease_pair:
  (sync_lease_token IS NULL) = (sync_lease_until IS NULL)
mail_accounts_mutation_lease_pair:
  (mutation_lease_token IS NULL) = (mutation_lease_until IS NULL)
mail_accounts_non_negative:
  sync_epoch >= 0 AND mutation_epoch >= 0 AND provision_attempts >= 0 AND sync_attempts >= 0
mail_accounts_status_state:
  status = 'provisioning' OR (status = 'error' AND provision_error_code IS NOT NULL) OR (status = 'active' AND desired_enabled = TRUE AND engine_account_id IS NOT NULL AND deployed_generation IS NOT NULL) OR (status = 'suspended' AND desired_enabled = FALSE AND engine_account_id IS NOT NULL AND deployed_generation IS NOT NULL) OR status = 'deleting'
mail_accounts_deleting_blocked_pending_shape:
  status <> 'deleting' OR (desired_enabled = FALSE AND sync_lease_token IS NULL AND sync_lease_until IS NULL AND mutation_lease_token IS NULL AND mutation_lease_until IS NULL AND provision_lease_token IS NULL AND provision_lease_until IS NULL AND attempted_control_generation IS NULL AND attempted_control_hash_version IS NULL AND attempted_control_hash IS NULL AND provision_error_code IS NULL AND next_provision_at IS NOT NULL)
mail_accounts_ready_projection:
  projection_state <> 'ready' OR (sync_cursor IS NOT NULL AND last_synced_at IS NOT NULL AND rebuild_checkpoint IS NULL AND rebuild_base_cursor IS NULL AND sync_error_code IS NULL)
mail_accounts_rebuilding_projection:
  projection_state <> 'rebuilding' OR rebuild_base_cursor IS NOT NULL
mail_accounts_live_provision_attempt:
  provision_lease_token IS NULL OR attempted_control_generation IS NOT NULL
mail_accounts_retained_provision_attempt:
  attempted_control_generation IS NULL OR provision_lease_token IS NOT NULL OR (provision_error_code IS NOT NULL AND provision_error_code IN ('MAILBOX_PROVISIONING_RETRY','MAILBOX_UNAVAILABLE') AND next_provision_at IS NOT NULL)
mail_accounts_unconverged_schedule:
  provision_lease_token IS NOT NULL OR attempted_control_generation IS NOT NULL OR next_provision_at IS NOT NULL OR (deployed_generation IS NOT NULL AND deployed_generation = control_generation AND deployed_hash_version = control_hash_version AND deployed_hash = control_hash)
```

`MailIdentity.Annotations` contains:

```text
mail_identities_send_allowed:
  send_allowed = FALSE OR status = 'active'
mail_identities_group_not_sendable:
  kind <> 'group' OR send_allowed = FALSE
mail_identities_deleting_alias_only:
  status <> 'deleting' OR kind = 'alias'
```

`MailDomain.Annotations` contains:

```text
mail_domains_desired_hash:
  desired_generation >= 1 AND desired_hash_version >= 1 AND length(desired_hash) = 32
mail_domains_attempted_tuple:
  ((attempted_generation IS NULL) = (attempted_hash_version IS NULL)) AND ((attempted_generation IS NULL) = (attempted_hash IS NULL)) AND (attempted_generation IS NULL OR (attempted_generation >= 1 AND attempted_hash_version >= 1 AND length(attempted_hash) = 32 AND attempted_generation <= desired_generation))
mail_domains_deployed_tuple:
  ((deployed_generation IS NULL) = (deployed_hash_version IS NULL)) AND ((deployed_generation IS NULL) = (deployed_hash IS NULL)) AND (deployed_generation IS NULL OR (deployed_generation >= 1 AND deployed_hash_version >= 1 AND length(deployed_hash) = 32 AND deployed_generation <= desired_generation))
mail_domains_lease_pair:
  (lease_token IS NULL) = (lease_until IS NULL)
mail_domains_attempt_count:
  attempt_count >= 0
mail_domains_unknown_attempt:
  last_error_code IS NULL OR last_error_code <> 'MAIL_DOMAIN_OUTCOME_UNKNOWN' OR attempted_generation IS NOT NULL
mail_domains_pending_state:
  state <> 'pending' OR (desired_present = TRUE AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL)
mail_domains_provisioning_state:
  state <> 'provisioning' OR (desired_present = TRUE AND attempted_generation IS NOT NULL AND lease_token IS NOT NULL AND last_error_code IS NULL AND next_retry_at IS NULL)
mail_domains_dns_required_state:
  state <> 'dns_required' OR (desired_present = TRUE AND engine_domain_id IS NOT NULL AND deployed_generation IS NOT NULL AND deployed_generation = desired_generation AND deployed_hash_version = desired_hash_version AND deployed_hash = desired_hash AND jsonb_array_length(required_dns_records) > 0 AND verification_status <> 'verified' AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL AND next_retry_at IS NULL)
mail_domains_verified_state:
  state <> 'verified' OR (desired_present = TRUE AND engine_domain_id IS NOT NULL AND deployed_generation IS NOT NULL AND deployed_generation = desired_generation AND deployed_hash_version = desired_hash_version AND deployed_hash = desired_hash AND verification_status = 'verified' AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL AND next_retry_at IS NULL)
mail_domains_error_state:
  state <> 'error' OR (desired_present = TRUE AND last_error_code IS NOT NULL AND lease_token IS NULL AND ((last_error_code = 'MAIL_DOMAIN_OUTCOME_UNKNOWN' AND attempted_generation IS NOT NULL AND next_retry_at IS NOT NULL) OR (last_error_code <> 'MAIL_DOMAIN_OUTCOME_UNKNOWN' AND attempted_generation IS NULL)))
mail_domains_deleting_state:
  state <> 'deleting' OR (desired_present = FALSE AND enabled = FALSE AND is_default = FALSE AND ((lease_token IS NOT NULL AND attempted_generation IS NOT NULL AND last_error_code IS NULL AND next_retry_at IS NULL) OR (lease_token IS NULL AND attempted_generation IS NOT NULL AND last_error_code = 'MAIL_DOMAIN_OUTCOME_UNKNOWN' AND next_retry_at IS NOT NULL) OR (lease_token IS NULL AND attempted_generation IS NULL)))
mail_domains_deleted_state:
  deleted_at IS NULL OR (state = 'deleting' AND desired_present = FALSE AND enabled = FALSE AND is_default = FALSE AND engine_domain_id IS NULL AND attempted_generation IS NULL AND lease_token IS NULL AND last_error_code IS NULL AND next_retry_at IS NULL)
mail_domains_default_state:
  is_default = FALSE OR (deleted_at IS NULL AND desired_present = TRUE AND enabled = TRUE AND state = 'verified' AND verification_status = 'verified')
```

`MailboxPolicy.Annotations` contains:

```text
mailbox_policies_bounds:
  max_message_bytes BETWEEN 1048576 AND 52428800 AND max_attachment_bytes BETWEEN 1 AND 26214400 AND max_attachment_bytes <= max_message_bytes AND max_recipients BETWEEN 1 AND 300 AND send_rate_per_minute BETWEEN 1 AND 1000 AND trash_retention_days BETWEEN 1 AND 3650 AND junk_retention_days BETWEEN 1 AND 3650
```

The PostgreSQL-only `jsonb_array_length` predicate is intentional because Hub persistence is
PostgreSQL. JSON shape/serialized-size validation still runs in the typed Go value.

#### Exact edges and delete behavior

- `MailAccount` has required unique owner edges to `Organization` and `User`, both
  `ON DELETE NO ACTION` under the 144-02A/B/C/D correction below; it has a required unique
  edge to `MailDomain` with `ON DELETE NO ACTION`;
  and a to-many `identities` inverse.
- `MailIdentity` has one required unique `account` edge on `mail_account_id` with
  `ON DELETE CASCADE`.
- `MailDomain` has one required unique `organization` edge with `ON DELETE CASCADE` and a
  to-many `accounts` inverse. The account foreign key is the database RESTRICT/NO ACTION
  proof; confirmed engine absence and soft-delete finalization remain service checks.
- `MailboxPolicy` has one required unique `organization` edge with `ON DELETE CASCADE`.
- Existing `Organization` inverse edge collections append only `mail_accounts`, `mail_domains`,
  and unique `mailbox_policy`; existing `User` collections append only unique `mail_account`.
  The correction changes only the two MailAccount owner-FK delete actions to `NO ACTION`.

#### Sequential implementation packets

1. `144-02A1`: `backend/internal/mail/domain_dns.go` plus its external unit; no Ent schema.
2. `144-02A2`: `MailDomain`, `MailboxPolicy`, their metadata oracle, and only the two
   `Organization` inverse edges they need. `MailDomain.accounts` is deliberately deferred so
   this intermediate source remains compilable without `MailAccount.Type`.
3. `144-02A3`: `MailAccount`, `MailIdentity`, their metadata oracle, `MailDomain.accounts`,
   `Organization.mail_accounts`, and `User.mail_account`.

No packet edits generated Ent output or migrations. 144-02E remains the sole generation
gate after all schema packets, founder migration authority and the supported Docker launcher
are accepted.

## 144-02B — projection

### `MailFolder`

Fields: `id`, `mail_account_id`, optional `engine_folder_id` (<=1,024 bytes), optional
same-account `parent_id`, optional `role`, `name` (raw valid UTF-8/no Unicode control/NUL,
1..120 runes/<=480 bytes; exact raw display bytes/runes retained), private
`normalized_name` (1..480 bytes), optional `color` (1..32 chars/<=128 bytes), required int32
`sort_order` default 0, required int32 `unread_count`/`total_count` default 0,
`control_state`, optional `control_receipt_id`, required int64 `last_seen_sync_epoch` default 0
(0..2,147,483,647),
timestamps, optional `deleted_at`, and `VersionMixin`.

Unique indexes:

- nullable `(mail_account_id,engine_folder_id)`;
- live system `(mail_account_id,role)` where role is non-null;
- live custom root `(mail_account_id,normalized_name)` where role and parent are null;
- live custom child `(mail_account_id,parent_id,normalized_name)` where role is null and
  parent is non-null.

Counters are non-negative and `unread_count <= total_count`.

Role is `inbox|sent|drafts|archive|trash|junk`; control state is
`creating|active|unknown|deleting|error`. Active requires engine id/no receipt/no deletion;
creating requires null engine id plus a receipt; unknown/error require a receipt and permit
the engine id to be either known or unknown. A live deleting row requires engine id+receipt;
confirmed deletion atomically sets `deleted_at`, clears `control_receipt_id`, retains the
safe opaque engine id only as a tombstone, and remains `deleting`. Every non-active state has
zero counters. Only that deleted tombstone may be deleting without a receipt. A system role
is a root (`parent_id` null); custom rows
have null role. Parent cannot equal id; depth/cycle/same-account is transactionally validated.
The receipt edge restricts deletion while referenced. A referenced move rule returns
`MAIL_FOLDER_IN_USE`. Any final service-owned account deletion is a future prerequisite-only
description, BLOCKED until an accepted engine deletion-port amendment and separate source plan;
it authorizes no current cascade or drain. Parent delete is service-ordered and restricted until
children move/delete. The 100-row live quota counts reserved non-deleted
control rows, not only rows that already have an engine id. Projection/persistence maps opaque
engine folder IDs to Hub UUID rows before the pure Folder policy; it never passes engine IDs
to that policy.

### `MailThread`

Fields: `id`, `mail_account_id`, required immutable int32 `thread_key_version` default 1,
exact 32-byte `thread_key`, `normalized_subject`
(<=998 UTF-8 bytes), required concrete `participant_addresses` typed JSON array with no default
(exact `[]string` seed snapshot of unique canonical normalized addresses, live 1..401 items,
<=1,048,576 serialized UTF-8 bytes under default HTML-escaping JSON; no truncation/preview),
`last_message_at`, required int32 `message_count`/`unread_count`, required booleans
`is_pinned`/`has_attachments`, snippet
(<=1,200 chars, <=400 graphemes, <=4,096 bytes), required non-negative
`last_seen_sync_epoch` int64 default 0 (0..2,147,483,647), timestamps and optional `deleted_at`. Unique
`(mail_account_id,thread_key_version,thread_key)`; read index
`(mail_account_id,is_pinned,last_message_at,id)` on live rows. Any final service-owned account
deletion is future/BLOCKED and does not authorize a current cascade.
Every live thread has `message_count >= 1` and `0 <= unread_count <= message_count`; cleanup
soft-deletes a thread in the same transaction that would otherwise make the count zero.
Only V1 is valid in this packet; version/key is immutable as one identity tuple. This is
storage representation only and authorizes neither schema code nor migration.

### `MailMessage`

Fields: `id`, `mail_account_id`, `mail_thread_id`, required `engine_message_id` and required
`engine_version` (each 1..1,024 bytes),
optional byte-preserved strict case-sensitive ASCII `internet_message_id`/`in_reply_to` Message-ID literals (<=998 UTF-8 bytes), required concrete
`references` typed JSON strict byte-preserved case-sensitive ASCII Message-ID string array default empty (<=100 items, each <=998 bytes, <=1,048,576
serialized UTF-8 bytes; no silent truncation), required concrete `from` typed JSON object containing exactly one Address,
and required concrete `to`/`cc`/`bcc`/`reply_to` typed JSON arrays default empty (each <=200
addresses and <=65,536 serialized bytes),
subject <=998 UTF-8 bytes, snippet limits above, optional `sent_at`/`received_at`,
`size_bytes` 0..52,428,800, priority, `has_attachments`, `is_draft`, `is_seen`, `is_answered`,
`is_flagged`, `is_junk`, optional auxiliary `engine_state` <=4,096 bytes,
`projection_origin`, optional
`flags_synced_at`, optional `calendar_method`, optional `calendar_event_id`, required int64
`last_seen_sync_epoch` default 0 (0..2,147,483,647), optional immutable `received_event_id`, timestamps and
optional `deleted_at`.

Indexes: unique `(mail_account_id,engine_message_id)`; non-unique live same-account resolver
index/equality `(mail_account_id, internet_message_id COLLATE "C") WHERE deleted_at IS NULL` or
binary equivalent over the strict byte-preserved case-sensitive ASCII literal (no folded/IDNA/
shadow key). Resolver joins/requires a same-account live `MailThread` (`thread.deleted_at IS NULL`);
unique nullable `received_event_id`; account/thread/date live indexes. Thread/account equality
is transactionally proven. Priority is `low|normal|high`, projection origin is
`inbound_or_unknown|confirmed_owner_submission|confirmed_owner_draft`, and calendar method is nullable
`request|cancel|reply`. BCC is empty except confirmed owner draft/submission. Account and
thread hard deletion cascade the owned projection; calendar deletion sets the optional link
null. Every projection upsert stores `engine_message_id` and the
`EngineMessageProjection.Version` from the same accepted engine value atomically. The
optional auxiliary `engine_state` is not a version substitute and cannot satisfy optimistic
reconciliation. A received event marker is set once and never rewritten.

### `MailMessageFolder`

UUID id, required `mail_message_id`/`mail_folder_id`, optional `engine_state` <=4,096 bytes,
required non-negative `last_seen_sync_epoch` default 0 and required `projected_at`. Unique
`(mail_message_id,mail_folder_id)`. Message/folder hard delete cascades the join;
same-account is checked before insert.

### `MailLabel`

Fields: UUID id/account, display name 1..60 chars/240 bytes, private normalized name 1..240
bytes, color 1..32 chars/<=128 bytes, immutable server-generated non-public `engine_keyword`: exactly 57 ASCII
bytes, `hubk-` plus canonical lower-case unpadded RFC-4648 base32 of exactly 32 cryptographically
random bytes. Validation must decode the suffix to exactly 32 bytes then re-encode it to the
identical lower-case suffix; `[a-z2-7]{52}` alone is insufficient. The value is immutable and
unique per account and embeds no Hub/account/user/label identifier or label name; timestamps,
optional `deleted_at`, `VersionMixin`. Live unique account+normalized name; unique account+keyword;
the database guard enforces ASCII/octet length 57 and `hubk-` plus `[a-z2-7]{52}`; application
validation additionally requires canonical decode-to-32/re-encode. The private marker is never
in OpenAPI/public DTOs. Resolver snapshots supplied to compiler are complete all-and-only live
non-deleted same-account rows under the same account lock; lifecycle freshness stays 04M-owned.
At most 100 live labels/account under account lock. The quota failure is terminally retained
by its receipt. Active rule references restrict delete.

### `MailMessageLabel`

UUID id, required message/label ids, `manual` and `engine_rule` booleans, timestamps. Unique
message+label; CHECK `manual OR engine_rule`; both owners hard-delete cascade; same-account is
checked transactionally.

## 144-02C — objects, attachments and delivery

### `MailObject`

Fields: UUID id/org/account/uploader, immutable server `object_key` <=512 bytes, immutable
bucket 1..63 bytes, immutable expected size 1..26,214,400, optional actual size in same range,
filename/media type 1..255 chars/1,020 bytes, optional exact SHA-256, immutable
UUID `draft_client_id`, upload state default `pending`, immutable expiry and immutable
`upload_drain_until`, nullable paired
`delete_lease_token` UUID/`delete_lease_until` timestamp, nullable `next_delete_at`, and common
timestamps. Exact owner fields are `organization_id`, `mail_account_id`, `uploader_user_id`;
size fields are `expected_size_bytes` and nullable `actual_size_bytes`; hash is nullable `sha256` bytes(32).

Unique `(bucket,object_key)` and one object per id-derived key. State is
`pending|confirmed|orphaned|deleted`. Actual size/hash is a nullable pair; pending requires
both null; confirmed requires both present and actual=expected. Pending/confirmed require all
 delete schedule/lease fields null. Orphaned/deleted require `next_delete_at` and may carry the
paired lease; `upload_drain_until = expires_at + INTERVAL '11 minutes'` and every scheduled
delete is no earlier than that immutable boundary. URLs expire within 15 minutes; the selected
provider/edge capability must prove a hard 10-minute maximum in-flight PUT lifetime and 60-second
maximum trusted clock skew or is unavailable. `deleted` is durable delete intent, not evidence that
provider bytes are absent; orphaned/deleted preserve immutable bucket/key/org/account/uploader/
metadata and every known size/hash. Account/org/uploader/staging scope equality is transactionally
validated. No scan field or project/file edge exists.

This is schema-ready representation only. It authorizes no current object transition, worker
claim, provider `DeleteObject(bucket,key)` I/O, or row hard-delete. A FUTURE/NON-DISPATCHABLE
protocol, only after an accepted engine deletion-port amendment and a separate accepted source
plan, must move an object to `orphaned|deleted` only after every Attachment reference is removed
or terminalized; those states reject confirm, download, new association and presign replay with
no storage I/O. That future protocol must claim a fresh token/deadline after the immutable drain
boundary, call idempotent `DeleteObject(bucket,key)` outside locks, accept only success or
authoritative NotFound, then reacquire User -> Account -> Object and hard-delete only by exact
token/state/scope CAS. Its future failure, crash-reclaim and stale-completion rules are respectively
exact-lease clear plus bounded retry, expired-lease reclaim only, and zero-row/no-compensating-I/O.
Account-owned object drain remains blocked, not an active Account-to-Object exception.

### `MailAttachment`

Fields: UUID `id`; nullable `mail_message_id` and `mail_draft_id`; nullable immutable
UUID `draft_client_id`; nullable unique `storage_object_id`; nullable `engine_blob_id` string
1..1,024 bytes; nullable `content_id` <=255 chars/1,020 bytes; required normalized
`filename`/`media_type` each 1..255 chars/1,020 bytes; `size_bytes` 0..26,214,400
(storage uploads remain non-empty; an engine-backed MIME attachment may be empty);
`disposition` `attachment|inline`; `scan_status` `pending|clean|rejected|error`; nullable
`sha256` bytes(32); timestamps.

Named checks enforce exactly one backing path. Storage backing requires storage object plus
`draft_client_id`, permits message/draft to be null while unclaimed, and once claimed the
draft's immutable client id must equal the attachment scope. Engine backing forbids
`draft_client_id`/draft id and requires a message owner. Clean/rejected require hash; pending
may retain the confirmed object hash while scanning; error retains any known hash. Inline
requires a non-empty content id. Storage object deletion is restricted while attached.
Message hard deletion cascades its engine-backed metadata and is service-ordered with
storage-backed attachment/object retention. The message foreign key is `RESTRICT`: before
message hard-delete, the service deletes engine-backed metadata and clears the message link
from retained storage-backed rows in one ordered transaction, so the backing XOR can never be
violated by an automatic `SET NULL`. Draft rows are not hard-deleted independently, so their
foreign key is also `RESTRICT` and preserves lifetime staging consumption. Same-account
equality is proved under account lock before either optional owner edge is set.

### `MailDraft`

Fields: UUID `id`/`mail_account_id`; required immutable UUID `draft_client_id`; required immutable
stable `engine_draft_key` 1..1,024 bytes; nullable `engine_message_id`/`engine_version`, each
1..1,024 bytes; required int64 `revision` default 1; status; immutable
`create_idempotency_key_hash` and `create_request_hash` bytes(32); nullable
`reply_to_message_id`/`forward_message_id` UUID snapshots; nullable UUID
`last_mutation_id`; timestamps. No content/address/attachment bytes are stored.

Unique account+draft key, account+draft-client id, account+create-key hash and nullable
account+engine message id. The engine pair is both-null or both-present. Status is
`active|saving|unknown|submitting|submitted|deleted`: active/submitting/submitted require the
pair and null mutation id; saving/unknown require a mutation id and permit a null pair only
for initial create; deleted is terminal, forbids mutation id and may retain either valid pair
state. Every selected storage attachment has the same draft-client id and receives this one
draft edge atomically; an existing different draft edge conflicts before engine I/O. One
draft owns at most 50 selected storage-backed attachments. Create/save replacement locks the
draft, validates the complete distinct attachment-id set and resulting row count, and rejects
a 51st association before any engine I/O. The repository enforces the same transactional
invariant even when called outside HTTP; a corrupted over-limit draft is never submitted.
Reply/forward links set null on projected-message hard delete. Any final service-owned account
deletion is future/BLOCKED and does not authorize a current cascade; individual draft deletion is
the terminal status transition, not row deletion.

### `MailSubmission`

Fields: UUID id/account, optional draft id, exact key/request SHA-256 hashes, immutable stable
engine submission key 1..1,024 bytes, status, optional `engine_submission_id` and optional
`engine_result_message_id` (each 1..1,024 bytes), optional UUID
`result_mail_message_id`, optional allowlisted error, non-negative attempts, timestamps.

Unique account+key hash, account+engine key, nullable account+engine submission id and
nullable account+engine-result-message id and nullable account+Hub-result-message snapshot.
Across every status the two opaque engine result fields are a both-null or both-present
pair.
State checks mirror OpenAPI while preserving engine identity: queued/submitting have no
result ids or error; sent requires both opaque engine ids plus `result_mail_message_id` with
no error; unknown always has a null Hub result and one of
`MAIL_SUBMISSION_OUTCOME_UNKNOWN|MAIL_SUBMISSION_RECONCILIATION_REQUIRED|MAILBOX_UNAVAILABLE`.
An unknown outcome may have both opaque engine ids null; a delivery-confirmed but
projection-pending unknown has both opaque engine ids present and uses only
`MAIL_SUBMISSION_RECONCILIATION_REQUIRED|MAILBOX_UNAVAILABLE`. Failed has all three result
ids null and exactly one of
`MAIL_SUBMISSION_REJECTED|MAIL_SUBMISSION_POLICY|MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN|MAIL_SUBMISSION_IDENTITY_UNAVAILABLE|MAIL_SUBMISSION_RECIPIENT_LIMIT|MAIL_SUBMISSION_SIZE_LIMIT|MAILBOX_UNAVAILABLE`.
The exact field names are `idempotency_key_hash`, `request_hash`, `engine_submission_key`,
`engine_submission_id`, `engine_result_message_id`, `result_mail_message_id`,
`last_error_code`, `attempt_count`,
`created_at`, `updated_at`; hashes are immutable bytes(32), attempt count is int32 default 0.
Draft deletion is restricted while non-terminal, otherwise its edge may set null; account
deletion cascades. The result Hub UUID is an immutable scalar snapshot, not a foreign key;
message projection deletion therefore does not erase it. A service may enter `sent` only in
the same transaction that maps `engine_result_message_id` to an owned `MailMessage`
projection whose returned account and engine message id exactly equal the requested account
and persisted result id and whose version is non-empty/bounded. That transaction derives
`confirmed_owner_submission` from the confirmed result pair rather than trusting an address,
folder or adapter-supplied origin. A mismatch stores no Hub UUID/event and remains
reconciliation-required. Until finalization the receipt remains `unknown`: a null opaque
pair calls `ReconcileSubmission`, while a complete pair skips it and retries only bounded
message read/projection/finalization. The complete-pair branch cannot requeue or call Submit.

### 144-02A/B/C/D root correction candidate — 2026-08-15

**Status:** `ROOT_CORRECTION_CANDIDATE` · `PRECOMMIT_REVIEW_REQUIRED` · `IMPLEMENTATION_BLOCKED` · `RUNTIME_NOT_RUN`. This is the one authoritative 144-02A/B/C/D correction proposal with accepted-02F dependency trace and expected exact-14 manifest; it grants no Ent, source, generation, migration, service, test, Docker, or runtime authority. The `5e314fb0` result is historical behavior-only evidence, never schema-readiness acceptance.

**Cross-packet correction scope.** This candidate supersedes the accepted 02A Organization/User
account cascades, the 02C Organization/Account object cascade, and the affected 02D User/
RecentContact lifecycle clauses. `Organization -> MailAccount`, `User -> MailAccount`, and
`Organization|User|MailAccount -> MailObject` are `ON DELETE NO ACTION`; Attachment-to-Object
remains `RESTRICT`. The 02A and 02D lifecycle review artifacts therefore join the exact-14
candidate manifest. Accepted 02F receipt/lease semantics remain current dependencies rather
than correction scope; until exact review/acceptance, none of this is schema/source authority.

**Representation boundary.** Ingress normalizes; persistence validators reject and never transform. `MailThreadParticipants []string` is stored exactly as `field.JSON("participant_addresses", MailThreadParticipants{}).SchemaType(map[string]string{dialect.Postgres:"jsonb"})`, `NOT NULL`, no default. `MailStoredAddress{Email string json:"email"; Name string json:"name"}`, `MailStoredAddresses []MailStoredAddress`, and `MailMessageReferences []string` are the Message JSON types: `from` is a required object with no default; `references`, `to`, `cc`, `bcc`, and `reply_to` are concrete empty arrays by default. Default `encoding/json` HTML escaping precedes serialized-byte caps. DB JSON checks only outer type/cardinality/member base shape; application validators own unknown keys, canonical grammar, duplicates, per-member and serialized-byte bounds.

**Common fields.** All ten entities have immutable `id uuid NOT NULL DEFAULT uuid.New`, immutable `created_at timestamptz NOT NULL DEFAULT now()`, and update-now `updated_at timestamptz NOT NULL DEFAULT now()`. Only Folder, Thread, Message, and Label have `deleted_at`. “set-once” is a service rule, not `Ent.Immutable`.

**Scope rules and canonical User lifecycle lock set.** Same-account joins, folder cycle/depth/quota, received-marker set-once, and the 50 selected-attachment locked-draft cap remain service invariants. Object is upload-only; Attachment is the only scan authority. Every nonphysical User deactivation/organization/email mutation, `RecentContact` link/unlink, and any future-authorized `MailObject` uploader-linked lifecycle path uses one canonical affected-account union: (a) every owned or otherwise lifecycle-affected `MailAccount`, including the operation's target account; (b) every account containing a `RecentContact.user_id` link to the User; and (c) every account of every physically present `MailObject` whose `uploader_user_id` is the User, regardless of `upload_state`. The future protocol locks that union exactly once in ascending raw UUID order, then re-reads all three dependency families and proves the union complete. It revalidates User state, organization, account, permission, exact contact links, and uploader/object ownership before any row mutation. Object paths would then lock object/receipt/control rows; `attachment_presign` preserves User before its existing `MailAccount -> receipt` claim order without changing the accepted-02F relative account-before-control-row rule. A race loser performs no row mutation, receipt creation/mutation, or engine/storage I/O, and no engine/storage I/O occurs while locks are held. Physical User deletion and account-owned object drain have no current exception or execution path.

**MailAccount deleting boundary.** The exact blocked-pending-engine-port shape is schema-ready only: `desired_enabled=false`, `status='deleting'`, no lease, no attempted tuple, no confirmed-absence tuple, and the existing due marker. The current candidate may describe/fence that blocked representation only; it starts no admission worker, precleanup, drain, partial mutation, engine I/O, or finalization. Ordinary workers require `status <> 'deleting'`. Until a separately pre-reviewed and accepted engine-port amendment freezes stable request identity/idempotency, a desired-absence primitive, authoritative absent reconciliation, lost-response/crash behavior, token/generation fencing, exact errors, selected adapter capability and 04J/04K ownership, destructive account drain, engine removal, attachment/object drain and final Account hard-delete are NOT DISPATCHED/BLOCKED. This exact14 candidate does not amend `ENGINE-CONTRACT` and grants no source/runtime authority.

**Physical User and Organization boundary.** No exact14 field authorizes physical User deletion. Physical User deletion and every cross-domain GDPR mutation are NOT_DISPATCHABLE/BLOCKED until a separately accepted durable User-deletion-intent/job/route contract and an accepted engine deletion-port amendment freeze intent persistence, authorization/reassignment snapshot, reactivation fencing, preflight before any GDPR mutation, recovery and final full purge. No mail precleanup, drain, partial mutation, User wait, or User deletion starts before both contracts and a separate accepted source plan. Deactivation and email change retain the canonical three-family union; organization change still requires zero accounts and uploader objects. Physical Organization completion is likewise BLOCKED by the missing engine-port amendment; `NO ACTION` remains the fail-closed edge and no eventual completion is claimed.

**02B field/type/null/default/immutable table (authoritative).**

| Entity | Field | PostgreSQL type | Null/default | Immutable |
|---|---|---|---|---|
| Folder | id/mail_account_id/parent_id/control_receipt_id | uuid | id/account required; parent/receipt nullable; `uuid.New` id | id/account yes |
| Folder | `VersionMixin.version` | integer/int32 | required/default `1`; Ent `field.Int`; PostgreSQL `integer NOT NULL DEFAULT 1`; Positive | mutable optimistic counter; no hand-written duplicate field |
| Folder | engine_folder_id | text | nullable/no default | service set-once after null, not Ent immutable |
| Folder | role/name/normalized_name/color | text | role/color nullable; name/normalized required; no defaults | role yes; name/normalized no |
| Folder | control_state/sort_order/unread_count/total_count/last_seen_sync_epoch | text/int32/int32/int32/int64 | state required/no default; others `0`, epoch required `0` | no |
| Thread | id/mail_account_id | uuid | required, `uuid.New` id | yes |
| Thread | thread_key_version/thread_key/participant_addresses | int32/bytea/jsonb | required, version `1`; participant required/no default | key/version yes |
| Thread | normalized_subject/last_message_at/message_count/unread_count/is_pinned/has_attachments/snippet | text/timestamptz/int32/int32/bool/bool/text | required/no DB defaults | no |
| Thread | last_seen_sync_epoch | int64 | required/default `0` | no |
| Message | id/mail_account_id | uuid | required, `uuid.New` id | id/account yes |
| Message | mail_thread_id | uuid | required/no default | no |
| Message | engine_message_id/engine_version/internet_message_id/in_reply_to/engine_state | text/text/text/text/text | first two required/no default; remainder nullable | engine_message_id yes; engine_version no |
| Message | references/to/cc/bcc/reply_to | jsonb | required/default `[]` | no |
| Message | from | jsonb | required/no default | no |
| Message | subject/snippet | text | required/no defaults | no |
| Message | sent_at/received_at/flags_synced_at/calendar_method/calendar_event_id | timestamptz/timestamptz/timestamptz/text/uuid | nullable/no defaults | no |
| Message | size_bytes/priority/projection_origin/is_draft/is_seen/is_answered/is_flagged/is_junk/has_attachments | int64/text/text/bool/bool/bool/bool/bool/bool | required/no DB defaults | no |
| Message | last_seen_sync_epoch | int64 | required/default `0` | no |
| Message | received_event_id | uuid | nullable/no default | service set-once then never rewritten, not Ent immutable |
| MessageFolder | id/mail_message_id/mail_folder_id | uuid | required, `uuid.New` id | ids yes |
| MessageFolder | engine_state/last_seen_sync_epoch/projected_at | text/int64/timestamptz | state nullable; epoch `0`; projected required | no |
| Label | id/mail_account_id | uuid | required, `uuid.New` id | yes |
| Label | `VersionMixin.version` | integer/int32 | required/default `1`; Ent `field.Int`; PostgreSQL `integer NOT NULL DEFAULT 1`; Positive | mutable optimistic counter; no hand-written duplicate field |
| Label | name/normalized_name/color/engine_keyword | text/text/text/text | name/normalized/color/keyword required, no defaults | keyword yes |
| MessageLabel | id/mail_message_id/mail_label_id | uuid | required, `uuid.New` id | ids yes |
| MessageLabel | manual/engine_rule | boolean | required/no DB defaults | no |

**02B exact edge/delete/index table.**

| Entity | FK / delete behavior | Exact indexes |
|---|---|---|
| Folder | account `CASCADE`; receipt `RESTRICT`; parent is service-ordered/restricted until child move/delete; soft delete only | `UNIQUE mail_folders_account_engine_folder_uidx (mail_account_id,engine_folder_id) WHERE engine_folder_id IS NOT NULL`; `UNIQUE mail_folders_live_role_uidx (mail_account_id,role) WHERE deleted_at IS NULL AND role IS NOT NULL`; `UNIQUE mail_folders_live_custom_root_uidx (mail_account_id,normalized_name) WHERE deleted_at IS NULL AND role IS NULL AND parent_id IS NULL`; `UNIQUE mail_folders_live_custom_child_uidx (mail_account_id,parent_id,normalized_name) WHERE deleted_at IS NULL AND role IS NULL AND parent_id IS NOT NULL` |
| Thread | account `CASCADE`; soft delete accompanies final message cleanup | `UNIQUE mail_threads_account_version_key_uidx (mail_account_id,thread_key_version,thread_key)`; non-unique `mail_threads_live_pinned_idx (mail_account_id,is_pinned,last_message_at,id) WHERE deleted_at IS NULL` |
| Message | account/thread `CASCADE`; calendar link `SET NULL`; soft delete only | `UNIQUE mail_messages_account_engine_uidx (mail_account_id,engine_message_id)`; non-unique `mail_messages_live_internet_message_id_idx (mail_account_id,internet_message_id COLLATE "C") WHERE deleted_at IS NULL`; `UNIQUE mail_messages_received_event_uidx (received_event_id)` (standard PostgreSQL NULL semantics); non-unique `mail_messages_live_account_date_idx (mail_account_id,received_at,sent_at,id) WHERE deleted_at IS NULL`; non-unique `mail_messages_live_thread_date_idx (mail_thread_id,received_at,sent_at,id) WHERE deleted_at IS NULL` |
| MessageFolder | message/folder `CASCADE`; same-account is locked service | `UNIQUE mail_message_folders_message_folder_uidx (mail_message_id,mail_folder_id)` |
| Label | account `CASCADE`; active rule reference `RESTRICT`; soft delete only | `UNIQUE mail_labels_live_normalized_name_uidx (mail_account_id,normalized_name) WHERE deleted_at IS NULL`; `UNIQUE mail_labels_account_keyword_uidx (mail_account_id,engine_keyword)` |
| MessageLabel | message/label `CASCADE`; same-account is locked service | `UNIQUE mail_message_labels_message_label_uidx (mail_message_id,mail_label_id)` |

**02C field/type/null/default/immutable table (authoritative).**

| Entity | Field | PostgreSQL type | Null/default | Immutable |
|---|---|---|---|---|
| Object | id/organization_id/mail_account_id/uploader_user_id/draft_client_id | uuid | required; id `uuid.New`; no other defaults | all yes |
| Object | bucket/object_key/filename/media_type | text | required/no defaults | yes |
| Object | expected_size_bytes/upload_state/expires_at/upload_drain_until | int64/text/timestamptz/timestamptz | required; state `pending` | expected/expiry/drain yes; state no |
| Object | actual_size_bytes/sha256 | int64/bytea | nullable/no defaults | no |
| Object | delete_lease_token/delete_lease_until/next_delete_at | uuid/timestamptz/timestamptz | nullable/no defaults; token/until paired | no |
| Attachment | id/draft_client_id/storage_object_id/engine_blob_id | uuid/uuid/uuid/text | id required `uuid.New`; others nullable | id/client and selected backing identity yes |
| Attachment | mail_message_id/mail_draft_id | uuid/uuid | nullable/no defaults | no |
| Attachment | content_id/filename/media_type/size_bytes/disposition | text/text/text/int64/text | content nullable; others required/no defaults | yes |
| Attachment | scan_status/sha256 | text/bytea | status required/no default; hash nullable | no |
| Draft | id/mail_account_id/draft_client_id/reply_to_message_id/forward_message_id | uuid | first three required (`uuid.New` id); reply/forward nullable | id/account/client yes |
| Draft | engine_draft_key/create_idempotency_key_hash/create_request_hash | text/bytea/bytea | required/no defaults | yes |
| Draft | engine_message_id/engine_version/revision/status/last_mutation_id | text/text/int64/text/uuid | pair/mutation nullable; revision `1`; status required/no default | no |
| Submission | id/mail_account_id/idempotency_key_hash/request_hash/engine_submission_key | uuid/uuid/bytea/bytea/text | required (`uuid.New` id); no defaults except attempts below | id/account/hashes/key yes |
| Submission | mail_draft_id/engine_submission_id/engine_result_message_id/result_mail_message_id | uuid/text/text/uuid | nullable; draft may become null via FK | service set-once opaque/result, not Ent immutable |
| Submission | last_error_code/attempt_count/status | text/int32/text | error nullable; attempts `0`; status required/no default | no |

**02C exact edge/delete/index table.**

All 02C `id` fields are `uuid NOT NULL DEFAULT uuid.New` immutable; `created_at` is `timestamptz NOT NULL DEFAULT now()` immutable and `updated_at` is `timestamptz NOT NULL DEFAULT now()` with update-now. `MailObject.upload_state` is `text NOT NULL DEFAULT 'pending'` and mutable. Object `actual_size_bytes`/`sha256` and Attachment `sha256` are service write-once `NULL -> validated value`; they are not generic mutable fields. Object account/org/uploader/draft-client/bucket/key/expected-size/filename/media/expiry are immutable; object deletion lease/schedule fields are mutable lifecycle controls. Submission opaque ids/result snapshot are service set-once, never Ent-immutable.

| Entity | FK / delete behavior | Exact indexes |
|---|---|---|
| MailObject | Organization/account/uploader User all `NO ACTION`; draft-client is scalar; User lifecycle uses the canonical three-family union; account/User/Organization finalization requires zero objects | `UNIQUE mail_objects_bucket_object_key_uidx (bucket,object_key)`; non-unique `mail_objects_delete_due_idx (next_delete_at,id) WHERE upload_state IN ('orphaned','deleted')` |
| MailAttachment | storage object/message/draft FKs `RESTRICT`; no cross-account row check | `UNIQUE mail_attachments_storage_object_uidx (storage_object_id) WHERE storage_object_id IS NOT NULL`; non-unique `mail_attachments_message_idx (mail_message_id)`; non-unique `mail_attachments_draft_idx (mail_draft_id)`; no unique index on `engine_blob_id` |
| MailDraft | account `CASCADE`; reply/forward projected-message snapshots `SET NULL`; individual delete status-only | `UNIQUE mail_drafts_account_engine_key_uidx (mail_account_id,engine_draft_key)`; `UNIQUE mail_drafts_account_client_uidx (mail_account_id,draft_client_id)`; `UNIQUE mail_drafts_account_create_key_uidx (mail_account_id,create_idempotency_key_hash)`; `UNIQUE mail_drafts_account_engine_message_uidx (mail_account_id,engine_message_id) WHERE engine_message_id IS NOT NULL` |
| MailSubmission | account `CASCADE`; draft FK `SET NULL` only after terminal sent/failed, while queued/submitting/unknown require non-null draft; normal draft delete status-only, exceptional locked purge only terminal sent/failed; result Hub UUID has no FK; CalendarReplyDelivery absent/deferred | `UNIQUE mail_submissions_account_key_hash_uidx (mail_account_id,idempotency_key_hash)`; `UNIQUE mail_submissions_account_engine_key_uidx (mail_account_id,engine_submission_key)`; `UNIQUE mail_submissions_account_engine_submission_uidx (mail_account_id,engine_submission_id) WHERE engine_submission_id IS NOT NULL`; `UNIQUE mail_submissions_account_engine_result_uidx (mail_account_id,engine_result_message_id) WHERE engine_result_message_id IS NOT NULL`; `UNIQUE mail_submissions_account_result_snapshot_uidx (mail_account_id,result_mail_message_id) WHERE result_mail_message_id IS NOT NULL` |

**Literal SQL predicate forms.** Every line below is an intended named PostgreSQL `CHECK (...)` predicate. State/pair predicates are NULL-safe with `COALESCE(..., FALSE)`. JSON checks deliberately enforce only DB-representable outer types, cardinality, primitive/object fields, and member primitive types; unknown keys, canonical values/grammar, duplicate detection, per-member semantic bounds, and serialized-byte caps are reject-not-transform application-validator obligations.

```sql
CONSTRAINT mail_folders_engine_folder_id CHECK (COALESCE(engine_folder_id IS NULL OR octet_length(engine_folder_id) BETWEEN 1 AND 1024, FALSE))
CONSTRAINT mail_folders_text CHECK (COALESCE(char_length(name) BETWEEN 1 AND 120 AND octet_length(name) BETWEEN 1 AND 480 AND octet_length(normalized_name) BETWEEN 1 AND 480 AND (color IS NULL OR (char_length(color) BETWEEN 1 AND 32 AND octet_length(color) BETWEEN 1 AND 128)), FALSE))
CONSTRAINT mail_folders_role CHECK (COALESCE(role IS NULL OR role IN ('inbox','sent','drafts','archive','trash','junk'), FALSE))
CONSTRAINT mail_folders_control_shape CHECK (COALESCE((control_state = 'active' AND engine_folder_id IS NOT NULL AND control_receipt_id IS NULL AND deleted_at IS NULL) OR (control_state = 'creating' AND engine_folder_id IS NULL AND control_receipt_id IS NOT NULL AND deleted_at IS NULL) OR (control_state IN ('unknown','error') AND control_receipt_id IS NOT NULL AND deleted_at IS NULL) OR (control_state = 'deleting' AND engine_folder_id IS NOT NULL AND ((deleted_at IS NULL AND control_receipt_id IS NOT NULL) OR (deleted_at IS NOT NULL AND control_receipt_id IS NULL))), FALSE))
CONSTRAINT mail_folders_counters CHECK (COALESCE(unread_count >= 0 AND total_count >= 0 AND unread_count <= total_count, FALSE))
CONSTRAINT mail_folders_non_active_counts CHECK (COALESCE(control_state = 'active' OR (unread_count = 0 AND total_count = 0), FALSE))
CONSTRAINT mail_folders_role_parent CHECK (COALESCE(role IS NULL OR parent_id IS NULL, FALSE))
CONSTRAINT mail_folders_not_self_parent CHECK (COALESCE(parent_id IS NULL OR parent_id <> id, FALSE))
CONSTRAINT mail_folders_epoch CHECK (COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE))
CONSTRAINT mail_folders_version CHECK (COALESCE(version >= 1, FALSE))

CONSTRAINT mail_threads_key CHECK (COALESCE(thread_key_version = 1 AND octet_length(thread_key) = 32, FALSE))
CONSTRAINT mail_threads_subject CHECK (COALESCE(octet_length(normalized_subject) <= 998, FALSE))
CONSTRAINT mail_threads_participants CHECK (CASE WHEN jsonb_typeof(participant_addresses) = 'array' THEN COALESCE(jsonb_array_length(participant_addresses) BETWEEN 1 AND 401 AND NOT jsonb_path_exists(participant_addresses, '$[*] ? (@.type() != "string")'), FALSE) ELSE FALSE END)
CONSTRAINT mail_threads_counts CHECK (COALESCE(message_count >= 1 AND unread_count BETWEEN 0 AND message_count, FALSE))
CONSTRAINT mail_threads_snippet CHECK (COALESCE(char_length(snippet) <= 1200 AND octet_length(snippet) <= 4096, FALSE))
CONSTRAINT mail_threads_epoch CHECK (COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE))

CONSTRAINT mail_messages_engine_ids CHECK (COALESCE(octet_length(engine_message_id) BETWEEN 1 AND 1024 AND octet_length(engine_version) BETWEEN 1 AND 1024, FALSE))
CONSTRAINT mail_messages_message_ids CHECK (COALESCE((internet_message_id IS NULL OR octet_length(internet_message_id) BETWEEN 1 AND 998) AND (in_reply_to IS NULL OR octet_length(in_reply_to) BETWEEN 1 AND 998), FALSE))
CONSTRAINT mail_messages_references CHECK (CASE WHEN jsonb_typeof(references) = 'array' THEN COALESCE(jsonb_array_length(references) <= 100 AND NOT jsonb_path_exists(references, '$[*] ? (@.type() != "string")'), FALSE) ELSE FALSE END)
CONSTRAINT mail_messages_from_shape CHECK (COALESCE(jsonb_typeof("from") = 'object' AND jsonb_typeof("from"->'email') = 'string' AND jsonb_typeof("from"->'name') = 'string', FALSE))
CONSTRAINT mail_messages_address_arrays CHECK (CASE WHEN jsonb_typeof("to") = 'array' AND jsonb_typeof(cc) = 'array' AND jsonb_typeof(bcc) = 'array' AND jsonb_typeof(reply_to) = 'array' THEN COALESCE(jsonb_array_length("to") <= 200 AND jsonb_array_length(cc) <= 200 AND jsonb_array_length(bcc) <= 200 AND jsonb_array_length(reply_to) <= 200 AND NOT jsonb_path_exists("to", '$[*] ? (@.type() != "object" || !exists(@.email) || !exists(@.name) || @.email.type() != "string" || @.name.type() != "string")') AND NOT jsonb_path_exists(cc, '$[*] ? (@.type() != "object" || !exists(@.email) || !exists(@.name) || @.email.type() != "string" || @.name.type() != "string")') AND NOT jsonb_path_exists(bcc, '$[*] ? (@.type() != "object" || !exists(@.email) || !exists(@.name) || @.email.type() != "string" || @.name.type() != "string")') AND NOT jsonb_path_exists(reply_to, '$[*] ? (@.type() != "object" || !exists(@.email) || !exists(@.name) || @.email.type() != "string" || @.name.type() != "string")'), FALSE) ELSE FALSE END)
CONSTRAINT mail_messages_text CHECK (COALESCE(octet_length(subject) <= 998 AND char_length(snippet) <= 1200 AND octet_length(snippet) <= 4096 AND (engine_state IS NULL OR octet_length(engine_state) <= 4096), FALSE))
CONSTRAINT mail_messages_size CHECK (COALESCE(size_bytes BETWEEN 0 AND 52428800, FALSE))
CONSTRAINT mail_messages_epoch CHECK (COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE))
CONSTRAINT mail_messages_priority CHECK (COALESCE(priority IN ('low','normal','high'), FALSE))
CONSTRAINT mail_messages_projection_origin CHECK (COALESCE(projection_origin IN ('inbound_or_unknown','confirmed_owner_submission','confirmed_owner_draft'), FALSE))
CONSTRAINT mail_messages_calendar_method CHECK (COALESCE(calendar_method IS NULL OR calendar_method IN ('request','cancel','reply'), FALSE))
CONSTRAINT mail_messages_bcc_origin CHECK (CASE WHEN jsonb_typeof(bcc) = 'array' THEN COALESCE(projection_origin IN ('confirmed_owner_submission','confirmed_owner_draft') OR jsonb_array_length(bcc) = 0, FALSE) ELSE FALSE END)
CONSTRAINT mail_message_folders_engine_state CHECK (COALESCE(engine_state IS NULL OR octet_length(engine_state) <= 4096, FALSE))
CONSTRAINT mail_message_folders_epoch CHECK (COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE))
CONSTRAINT mail_labels_text CHECK (COALESCE(char_length(name) BETWEEN 1 AND 60 AND octet_length(name) BETWEEN 1 AND 240 AND octet_length(normalized_name) BETWEEN 1 AND 240 AND char_length(color) BETWEEN 1 AND 32 AND octet_length(color) BETWEEN 1 AND 128 AND octet_length(engine_keyword) = 57 AND engine_keyword ~ '^hubk-[a-z2-7]{52}$', FALSE))
CONSTRAINT mail_labels_version CHECK (COALESCE(version >= 1, FALSE))
CONSTRAINT mail_message_labels_manual_or_engine_rule CHECK (COALESCE(manual OR engine_rule, FALSE))

CONSTRAINT mail_objects_text CHECK (COALESCE(char_length(bucket) BETWEEN 1 AND 63 AND octet_length(bucket) BETWEEN 1 AND 63 AND char_length(object_key) BETWEEN 1 AND 512 AND octet_length(object_key) BETWEEN 1 AND 512 AND char_length(filename) BETWEEN 1 AND 255 AND octet_length(filename) BETWEEN 1 AND 1020 AND char_length(media_type) BETWEEN 1 AND 255 AND octet_length(media_type) BETWEEN 1 AND 1020, FALSE))
CONSTRAINT mail_objects_sizes CHECK (COALESCE(expected_size_bytes BETWEEN 1 AND 26214400 AND (actual_size_bytes IS NULL OR actual_size_bytes BETWEEN 1 AND 26214400), FALSE))
CONSTRAINT mail_objects_actual_sha_pair CHECK (COALESCE((actual_size_bytes IS NULL) = (sha256 IS NULL) AND (sha256 IS NULL OR octet_length(sha256) = 32), FALSE))
CONSTRAINT mail_objects_upload_enum CHECK (COALESCE(upload_state IN ('pending','confirmed','orphaned','deleted'), FALSE))
CONSTRAINT mail_objects_delete_lease_pair CHECK (COALESCE((delete_lease_token IS NULL) = (delete_lease_until IS NULL), FALSE))
CONSTRAINT mail_objects_upload_state CHECK (COALESCE((upload_state = 'pending' AND actual_size_bytes IS NULL AND sha256 IS NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL) OR (upload_state = 'confirmed' AND actual_size_bytes = expected_size_bytes AND sha256 IS NOT NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL) OR (upload_state IN ('orphaned','deleted') AND next_delete_at IS NOT NULL), FALSE))
CONSTRAINT mail_objects_upload_drain_boundary CHECK (COALESCE(upload_drain_until = expires_at + INTERVAL '11 minutes' AND (next_delete_at IS NULL OR next_delete_at >= upload_drain_until), FALSE))

CONSTRAINT mail_attachments_text CHECK (COALESCE((engine_blob_id IS NULL OR octet_length(engine_blob_id) BETWEEN 1 AND 1024) AND (content_id IS NULL OR (char_length(content_id) BETWEEN 1 AND 255 AND octet_length(content_id) BETWEEN 1 AND 1020)) AND char_length(filename) BETWEEN 1 AND 255 AND octet_length(filename) BETWEEN 1 AND 1020 AND char_length(media_type) BETWEEN 1 AND 255 AND octet_length(media_type) BETWEEN 1 AND 1020 AND size_bytes BETWEEN 0 AND 26214400 AND (sha256 IS NULL OR octet_length(sha256) = 32), FALSE))
CONSTRAINT mail_attachments_backing_xor CHECK (COALESCE((storage_object_id IS NOT NULL) <> (engine_blob_id IS NOT NULL), FALSE))
CONSTRAINT mail_attachments_owner_exclusive CHECK (COALESCE(NOT (mail_message_id IS NOT NULL AND mail_draft_id IS NOT NULL), FALSE))
CONSTRAINT mail_attachments_storage_shape CHECK (COALESCE(storage_object_id IS NULL OR (draft_client_id IS NOT NULL AND size_bytes >= 1), FALSE))
CONSTRAINT mail_attachments_engine_shape CHECK (COALESCE(engine_blob_id IS NULL OR (mail_message_id IS NOT NULL AND mail_draft_id IS NULL AND draft_client_id IS NULL), FALSE))
CONSTRAINT mail_attachments_inline_content CHECK (COALESCE(disposition <> 'inline' OR content_id IS NOT NULL, FALSE))
CONSTRAINT mail_attachments_disposition CHECK (COALESCE(disposition IN ('attachment','inline'), FALSE))
CONSTRAINT mail_attachments_scan_enum CHECK (COALESCE(scan_status IN ('pending','clean','rejected','error'), FALSE))
CONSTRAINT mail_attachments_hash_scan CHECK (COALESCE(scan_status NOT IN ('clean','rejected') OR sha256 IS NOT NULL, FALSE))

CONSTRAINT mail_drafts_key CHECK (COALESCE(octet_length(engine_draft_key) BETWEEN 1 AND 1024, FALSE))
CONSTRAINT mail_drafts_engine_pair CHECK (COALESCE((engine_message_id IS NULL AND engine_version IS NULL) OR (octet_length(engine_message_id) BETWEEN 1 AND 1024 AND octet_length(engine_version) BETWEEN 1 AND 1024), FALSE))
CONSTRAINT mail_drafts_revision CHECK (COALESCE(revision >= 1, FALSE))
CONSTRAINT mail_drafts_hashes CHECK (COALESCE(octet_length(create_idempotency_key_hash) = 32 AND octet_length(create_request_hash) = 32, FALSE))
CONSTRAINT mail_drafts_status_enum CHECK (COALESCE(status IN ('active','saving','submitting','submitted','unknown','deleted'), FALSE))
CONSTRAINT mail_drafts_state CHECK (COALESCE((status IN ('active','submitting','submitted') AND engine_message_id IS NOT NULL AND last_mutation_id IS NULL) OR (status IN ('saving','unknown') AND last_mutation_id IS NOT NULL) OR (status = 'deleted' AND last_mutation_id IS NULL), FALSE))

CONSTRAINT mail_submissions_key CHECK (COALESCE(octet_length(engine_submission_key) BETWEEN 1 AND 1024, FALSE))
CONSTRAINT mail_submissions_hashes CHECK (COALESCE(octet_length(idempotency_key_hash) = 32 AND octet_length(request_hash) = 32, FALSE))
CONSTRAINT mail_submissions_opaque_pair CHECK (COALESCE((engine_submission_id IS NULL AND engine_result_message_id IS NULL) OR (octet_length(engine_submission_id) BETWEEN 1 AND 1024 AND octet_length(engine_result_message_id) BETWEEN 1 AND 1024), FALSE))
CONSTRAINT mail_submissions_attempts CHECK (COALESCE(attempt_count >= 0, FALSE))
CONSTRAINT mail_submissions_status_enum CHECK (COALESCE(status IN ('queued','submitting','sent','unknown','failed'), FALSE))
CONSTRAINT mail_submissions_state CHECK (COALESCE((status IN ('queued','submitting') AND mail_draft_id IS NOT NULL AND engine_submission_id IS NULL AND engine_result_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL) OR (status = 'sent' AND engine_submission_id IS NOT NULL AND engine_result_message_id IS NOT NULL AND result_mail_message_id IS NOT NULL AND last_error_code IS NULL) OR (status = 'unknown' AND mail_draft_id IS NOT NULL AND result_mail_message_id IS NULL AND last_error_code IN ('MAIL_SUBMISSION_OUTCOME_UNKNOWN','MAIL_SUBMISSION_RECONCILIATION_REQUIRED','MAILBOX_UNAVAILABLE') AND NOT (engine_submission_id IS NOT NULL AND last_error_code = 'MAIL_SUBMISSION_OUTCOME_UNKNOWN')) OR (status = 'failed' AND engine_submission_id IS NULL AND engine_result_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IN ('MAIL_SUBMISSION_REJECTED','MAIL_SUBMISSION_POLICY','MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN','MAIL_SUBMISSION_IDENTITY_UNAVAILABLE','MAIL_SUBMISSION_RECIPIENT_LIMIT','MAIL_SUBMISSION_SIZE_LIMIT','MAILBOX_UNAVAILABLE')), FALSE))
```

Every `*_uidx` in the index inventory is a `CREATE UNIQUE INDEX`; every `*_idx` is non-unique. In particular, `mail_attachments_storage_object_uidx` is a unique partial index on `(storage_object_id) WHERE storage_object_id IS NOT NULL`; there is no uniqueness constraint or index on `engine_blob_id`. `Organization -> MailObject`, `User -> MailObject`, and `MailAccount -> MailObject` are `ON DELETE NO ACTION`; `MailAttachment -> MailObject` is `ON DELETE RESTRICT`. A direct delete is guaranteed to fail only while its named blocker survives; no universal arbitrary-direct-SQL safety is claimed.

## 144-02D — user configuration

### `MailUserSettings`

One UUID/account row with required `default_identity_id`, optional
`default_signature_id`, required `remote_images` `blocked|trusted_senders` default
`blocked`, required `compose_format` `rich|plain` default `rich`, timestamps and
`VersionMixin`. Unique `mail_account_id`. The two defaults are canonical; public identity/signature
`is_default` is derived. Same-account, active/send-allowed identity and same-account live
signature are checked under a locked settings row. Selected identity deletion is restricted;
selected signature deletion clears the pointer and increments settings version atomically.
Any final service-owned account deletion is future/BLOCKED until the accepted engine deletion-port
amendment and separate source plan; it authorizes no current cascade. Identity/signature edges
otherwise restrict until their ordered service transitions complete.

### `MailSignature`

UUID/account, name 1..80 chars/320 bytes, body format, optional sanitized rich body and
required derived plain body, each <=100,000 chars/400,000 bytes, timestamps, optional
`deleted_at`, `VersionMixin`. Format is `rich|plain`; named CHECK requires rich non-null for
rich and null for plain. Exact fields are `id`, `mail_account_id`, `name`, `body_format`,
`rich_body`, `plain_body`, `created_at`, `updated_at`, `deleted_at` plus `VersionMixin`.
No default boolean exists. At most 20 live rows/account; the quota failure is terminally
stored in its receipt. Any final service-owned account deletion is future/BLOCKED and authorizes
no current cascade; the settings pointer is cleared before a selected signature is soft-deleted.

### `MailRule`

UUID/account, name 1..120 chars/480 bytes, enabled, priority 0..10,000, exact typed condition
array 1..20 and action array 1..10, combined serialized UTF-8 <=65,536 bytes, timestamps,
optional `deleted_at`, `VersionMixin`. JSON uses concrete structs matching OpenAPI and rejects
unknown fields; no `map[string]any`, raw Sieve, engine keyword/script id. Active canonical
ordering is `(priority,id)`. At most 100 live rows/account. Rule writes and referenced folder/
label deletes lock the account before validating JSON target UUIDs. Exact JSON fields are
stored as `conditions` and `actions`; each `move` owns one folder UUID, each `add_label` one
label UUID, and `mark_read` owns neither. A live index on
`(mail_account_id,enabled,priority,id)` drives hashing/deployment. Any final service-owned account
deletion is future/BLOCKED and authorizes no current cascade; soft-deleted targets are invalid and
remain protected by the shared lock. Quota failure is
terminally stored in its receipt before a rule or desired deployment generation is written.

### `RecentContact`

Required `mail_account_id` FK `ON DELETE CASCADE`; nullable `user_id` FK `ON DELETE SET NULL`.
Canonical normalized email is required 1..320 bytes; nullable display name uses the safe
address-name policy and is <=160 runes/640 bytes; `last_used_at` is required; `use_count` is
int32 `0..2147483647`; timestamps apply.
The database enforces unique `(mail_account_id, normalized_email)` and partial eviction index
`(mail_account_id,last_used_at,id) WHERE user_id IS NULL`. The storage cap is exactly <=1,000
rows per account where `user_id IS NULL`; linked rows are exempt.

Only an active same-organization user with exactly equal canonical normalized email may link;
linked rows have `display_name NULL`, linking clears it, and unlink/nonphysical User lifecycle
changes never restore a stale external name. Cap restoration is transaction/service logic, not a
row CHECK.

External-only admission/upsert locks the account row first and revalidates live ownership. Same-email
upsert refreshes injected UTC `last_used_at` and saturates `use_count` at `INT32_MAX`; distinct
admission at 999 inserts without eviction, while at 1,000 it evicts the exact per-account minimum
`(last_used_at ASC,id ASC raw UUID)` before insert. Every operation that creates, changes or
consumes `RecentContact.user_id`, including link/manual unlink and nonphysical User
deactivate/organization/email mutation, uses the canonical User-lifecycle three-family union above: User
first, one ascending raw-UUID account-lock pass, then a re-read of owned/lifecycle accounts,
exact contact links, and every persisted uploader object. The user lock prevents new link or
invalidating user mutation races. After account locks, link/unlink re-reads exact rows and revalidates account/user
organization, status and exact canonical email. Explicit unlink clears `user_id`+`display_name`,
then repeatedly deletes the global minimum within that locked account among all its current
null-user rows until <=1,000. The just-unlinked row participates and may be
evicted; unlink is never rejected or truncated to preserve it.

Nonphysical User deactivation/organization/email mutation holds the User lock and uses that
canonical union; this section never defines a narrower complete set. While the User is locked, no
link, owned-account lifecycle, or uploader-object lifecycle may change outside this protocol.
Physical User deletion and every cross-domain GDPR route are NOT_DISPATCHABLE/BLOCKED: they do not
adjudicate blockers, commit accounts to `deleting`, wait, clean contacts, drain objects, or delete
the User unless the separate accepted durable User intent/job/route contract, accepted engine
deletion-port amendment, and separate accepted source plan exist. Organization change requires
zero owned accounts and zero physically present uploader objects and otherwise aborts before
contact mutation. Deactivation and email change do not reject solely because objects exist; after
the canonical re-read they clear invalid `user_id`+`display_name` and restore each cap. This
RecentContact cleanup itself never rejects/truncates a permitted mutation or prefers only newly
external rows.

The blocked account representation is account-scoped and never requests a User lock, but it does
not commit a runnable deletion workflow. `User -> MailAccount NO ACTION` and `User -> MailObject
NO ACTION` fail closed only while those named blockers survive. `RecentContact.user_id ON DELETE
SET NULL` is referential fallback only, never cap restoration, so arbitrary direct User deletion
without a named account/object blocker is not claimed safe.

External `display_name` ingestion reuses safe address-name validation: valid UTF-8, no Unicode
`Cc` or header-injection controls, <=160 runes/640 bytes. Permitted whitespace, `Cf`, rune and
byte content is preserved exactly without trim/NFC/casefold. Linking clears the name; unlink and
user invalidation never restore a stale value.

## 144-02F — receipts and rule deployment

### `MailMutationReceipt`

| Fields | Storage / nullability / default |
|---|---|
| `id`, `organization_id`, `actor_user_id` | UUID primary plus required immutable scope UUID and historical actor UUID snapshot; actor has no foreign-key edge |
| `mail_account_id` | nullable immutable UUID; required by the operation matrix below |
| `operation`, `key_source`, `state` | required closed enums; state defaults `in_progress` |
| `idempotency_key_hash`, `request_hash` | required immutable bytes(32) |
| `fingerprint_version` | required int32 >=1, default 1 |
| `target_kind`, `target_id` | nullable `folder|label|signature|rule|attachment|domain|identity` plus UUID pair, both-null or both-present |
| `bulk_action` | nullable closed bulk-action enum |
| `argument_kind`, `argument_id` | nullable `folder|label` plus UUID pair, both-null or both-present |
| `original_http_status` | nullable int16, 100..599 |
| `safe_code` | nullable closed enum below, <=128 bytes |
| `reconcile_attempts` | required int32, default 0, exact range 0..10 |
| `next_reconcile_at` | nullable UTC timestamp |
| `reconcile_lease_token`, `reconcile_lease_until` | nullable UUID/time pair |
| `terminal_at`, `replay_until`, `compacted_at` | nullable UTC timestamps |
| `created_at`, `updated_at` | common timestamps |

Unique `(organization_id,actor_user_id,operation,idempotency_key_hash)`.

Operations are exactly `folder_create|folder_update|folder_delete|label_create|signature_create|rule_create|attachment_presign|domain_create|alias_create|message_bulk`.
Only folder update/delete use `internal`; every other operation uses `client`. Account is
required for mailbox operations and null for organization-scoped domain create. Exact
operation grammar is:

| Operation | Target | Action / argument | Terminal success | Terminal failure code/status | Non-terminal unknown |
|---|---|---|---|---|---|
| `folder_create` | `folder` required after row reservation | none | 201 | `MAIL_FOLDER_NAME_CONFLICT|MAIL_FOLDER_TREE_CONFLICT` / 409; `MAIL_FOLDER_LIMIT_EXCEEDED` / 422; `MAIL_RECONCILIATION_EXHAUSTED` / 409 | `MAIL_FOLDER_RECONCILIATION_REQUIRED` |
| `folder_update` | `folder` always | none; desired fields live on folder | 200 | `MAIL_RECONCILIATION_EXHAUSTED` / 409 only (preconditions occur before internal receipt) | `MAIL_FOLDER_RECONCILIATION_REQUIRED` |
| `folder_delete` | `folder` always | none | 204 | `MAIL_RECONCILIATION_EXHAUSTED` / 409 only (preconditions occur before internal receipt) | `MAIL_FOLDER_RECONCILIATION_REQUIRED` |
| `label_create` | `label` only on success | none | 201 | `MAIL_LABEL_NAME_CONFLICT|MAIL_LABEL_LIMIT_EXCEEDED` / 409 or 422 | forbidden |
| `signature_create` | `signature` only on success | none | 201 | `MAIL_SIGNATURE_LIMIT_EXCEEDED` / 422 | forbidden |
| `rule_create` | `rule` only on success | none | 201 | `MAIL_RULE_LIMIT_EXCEEDED` / 422 | forbidden |
| `attachment_presign` | `attachment` only after object reservation | none | 201 | none | forbidden |
| `domain_create` | `domain` only on success | none | 202 | `MAIL_DOMAIN_UNAVAILABLE|MAIL_DOMAIN_LIMIT_EXCEEDED` / 409 or 422 | forbidden on HTTP path; domain worker owns later state |
| `alias_create` | `identity` only on success | none | 202 | `MAIL_ALIAS_ADDRESS_UNAVAILABLE|MAIL_ALIAS_ACCOUNT_NOT_ACTIVE|MAIL_IDENTITY_LIMIT_EXCEEDED` / 409 or 422 | forbidden on HTTP path; account worker owns later state |
| `message_bulk` | parent target pair null; ordered child targets required | exact `bulk_action`; argument folder only for `move`, label only for add/remove-label, null otherwise | 200 | `MAIL_BULK_EXPANSION_LIMIT` / 422 | `MAIL_ENGINE_OUTCOME_UNKNOWN` only for reconcilable non-classifier engine state |

The persisted parent `safe_code` enum is exactly the union named by that matrix:
`MAIL_FOLDER_NAME_CONFLICT|MAIL_FOLDER_TREE_CONFLICT|MAIL_FOLDER_LIMIT_EXCEEDED|MAIL_FOLDER_RECONCILIATION_REQUIRED|MAIL_RECONCILIATION_EXHAUSTED|MAIL_LABEL_NAME_CONFLICT|MAIL_LABEL_LIMIT_EXCEEDED|MAIL_SIGNATURE_LIMIT_EXCEEDED|MAIL_RULE_LIMIT_EXCEEDED|MAIL_DOMAIN_UNAVAILABLE|MAIL_DOMAIN_LIMIT_EXCEEDED|MAIL_ALIAS_ADDRESS_UNAVAILABLE|MAIL_ALIAS_ACCOUNT_NOT_ACTIVE|MAIL_IDENTITY_LIMIT_EXCEEDED|MAIL_BULK_EXPANSION_LIMIT|MAIL_ENGINE_OUTCOME_UNKNOWN`.
No idempotency reuse/expiry code or target-child code is a member of this enum.

`bulk_action` is exactly
`move|add_label|remove_label|mark_read|mark_unread|flag|unflag|archive|junk|not_junk|trash|restore|pin|unpin`.
The idempotency reuse/replay codes are derived from hashes, target availability and
`replay_until`; they are never stored as `safe_code`. Busy/stale/protected checks that occur
before receipt reservation likewise leave no receipt.

State is `in_progress|unknown|completed|failed` with named checks:

| State | Required | Forbidden |
|---|---|---|
| `in_progress` | operation grammar; optional live reconcile lease | safe code, status and all terminal timestamps |
| `unknown` | one allowed unknown code, `next_reconcile_at`; captured target/children | original status and terminal timestamps; lease is null while waiting and paired while a reconciler owns it |
| `completed` | route success status, `terminal_at`, `replay_until=terminal_at+30 days` | safe code, retry and lease |
| `failed` | route failure status/code from the table, terminal/replay timestamps | retry and lease |

The first terminal transition makes status/code/times immutable. Client rows at or after
replay time behave expired independent of cleanup; compaction deletes all target/member rows
atomically and sets `compacted_at >= replay_until` while retaining the parent hashes and
route metadata as the lifetime tombstone. Internal terminal rows may hard-delete after
replay time and never accept an HTTP key. Lease pair is both-null or both-present. Any final
service-owned account or organization deletion is future/BLOCKED and authorizes no current receipt
cascade. Actor deletion never removes or blocks an organization receipt because `actor_user_id` is
a historical scalar snapshot rather than a user edge.

### `MailMutationReceiptTarget`

UUID id/receipt, request index 0..199, target type `message|thread`, target UUID snapshot,
status `succeeded|partial|failed|unknown`, affected/succeeded/failed counts 0..200, training
status `not_requested|accepted|rejected|unknown`, optional exact OpenAPI bulk code, timestamps.
Unique receipt+request index and receipt+target type+target id. Counts obey
`succeeded+failed <= affected`. Request indexes are contiguous from zero in request order,
enforced while the locked receipt is reserved. Exact outcome checks are:

| Status | Counts | Code |
|---|---|---|
| `succeeded` | affected >=1; succeeded=affected; failed=0 | null |
| `partial` | affected >=2; succeeded>=1; failed>=1; succeeded+failed=affected | `MAIL_ACTION_REJECTED` or `MAIL_CLASSIFIER_REJECTED` |
| `failed` unavailable | all counts 0 | `MAIL_TARGET_UNAVAILABLE|MAIL_FOLDER_UNAVAILABLE|MAIL_LABEL_UNAVAILABLE` |
| `failed` executed | affected>=1; succeeded=0; failed=affected | `MAIL_ACTION_REJECTED|MAIL_CLASSIFIER_REJECTED` |
| `unknown` | affected>=1; succeeded+failed<affected | `MAIL_ENGINE_OUTCOME_UNKNOWN|MAIL_CLASSIFIER_OUTCOME_UNKNOWN` |

For `junk|not_junk`, training is `accepted` only with succeeded/null-code, `rejected` only
with partial/failed plus `MAIL_CLASSIFIER_REJECTED`, and `unknown` only with unknown plus
`MAIL_CLASSIFIER_OUTCOME_UNKNOWN`; pre-execution unavailable rows use `not_requested`.
Every other action requires `not_requested` and forbids classifier codes. Receipt
delete/compaction cascades; projected target deletion does not.

### `MailMutationReceiptMember`

UUID id/target plus denormalized immutable receipt id and message UUID snapshot. Unique
`(receipt_id,message_id)` and `(target_id,message_id)`. The service locks the parent and
enforces at most 200 distinct members across the receipt before engine I/O; message targets
with nonzero affected count have exactly their one matching member. Each member's receipt id
must equal its target's receipt id. Thread members are the immutable owner-scoped expansion;
unavailable zero-count targets have none. Target/receipt deletion cascades; projected message
deletion does not.

### `MailRuleDeployment`

One UUID/account row has required `desired_generation`, `desired_hash_version`,
`desired_hash`; nullable `attempted_generation`, `attempted_hash_version`, `attempted_hash`;
nullable `deployed_generation`, `deployed_hash_version`, `deployed_hash`; required `state`;
nullable public `last_error_code`, nullable internal-only `internal_terminal_code`, nullable
`next_retry_at`, required int32 `attempt_count` default 0 and bounded `0..10`, nullable
`lease_token`/`lease_until`, timestamps. `internal_terminal_code` is nullable or exactly
`MAIL_RULE_DEPLOYMENT_EXHAUSTED`; it is never an API, audit, event, log or error output.
Hashes are bytes(32), generations int64
>=1, versions int32 >=1. Unique account; tuples and lease are independently all-or-none.

| State | Required | Forbidden / additional rule |
|---|---|---|
| `pending` | desired tuple; deployed null or older/different; `attempt_count=0` | attempted, lease, public/internal error null; retry optional |
| `applying` | attempted tuple and lease; attempted generation <= desired; `attempt_count=1..10` | public/internal error and retry null |
| `ready` | deployed tuple exactly equals desired; `attempt_count=0` | attempted, lease, public/internal error and retry null |
| `unknown` | attempted tuple, retry and `MAIL_RULE_OUTCOME_UNKNOWN|MAIL_RULE_ENGINE_UNAVAILABLE`; `attempt_count=1..9` | lease null; internal code null; deployed does not advance |
| `error` | attempted tuple and one definitive/retryable public error with `attempt_count=1..9`, or exhaustion with `attempt_count=10` | lease null; deployed never equals desired unless state immediately becomes ready |

Error allowlist is exactly
`MAIL_RULE_CAPABILITY_UNAVAILABLE|MAIL_RULE_ENGINE_UNAVAILABLE|MAIL_RULE_APPLY_REJECTED|MAIL_RULE_OUTCOME_UNKNOWN|MAIL_RULE_DRIFT`.
Capability-unavailable/apply-rejected errors have null retry; engine-unavailable/drift errors
have non-null retry. The exact deployment row checks are
`mail_rule_deployments_attempt_count_bounds` (`attempt_count BETWEEN 0 AND 10`),
`mail_rule_deployments_internal_terminal_code_allowlist`
(`internal_terminal_code IS NULL OR internal_terminal_code = 'MAIL_RULE_DEPLOYMENT_EXHAUSTED'`),
`mail_rule_deployments_exhaustion_terminal`
(`internal_terminal_code <> 'MAIL_RULE_DEPLOYMENT_EXHAUSTED' OR (state = 'error' AND attempt_count = 10 AND last_error_code IS NULL AND next_retry_at IS NULL AND lease_token IS NULL AND lease_until IS NULL AND attempted_generation IS NOT NULL)`),
`mail_rule_deployments_unknown_pre_cap` (`state <> 'unknown' OR attempt_count BETWEEN 1 AND 9`),
and state updates: applying is `1..10`; unknown/retryable public error is `1..9`; pending/ready
are zero; ready has internal code null. Existing desired/attempted/deployed tuple
equality/inequality constraints remain locked.

Define `prior_attempts p` per immutable attempted tuple/current desired. For `0 <= p < 10`, a
claim persists attempted generation/hash, lease/fence and `p + 1` before at most one bounded
read/install cycle; `p = 9` is the tenth cycle. A tenth owner may confirm and complete. If it
crashes or remains unknown, no eleventh cycle is allowed: `p = 10` does no I/O after expiry and,
only when the attempted tuple still equals current desired, transitions to `error` with the
internal terminal code, null public `last_error_code`, null retry/lease, retained attempted tuple
and unchanged deployed tuple. If desired advanced, the old owner cannot terminalize the new desired;
after its outcome or expiry the new tuple begins at budget zero. Late owners CAS deployment with the
account token/epoch. A successful attempted tuple copies exactly to deployed, then clears attempted:
it becomes ready only when it equals current desired, otherwise pending. A newer desired tuple never
rewrites applying/unknown attempted state.

Canonical hash v1 is SHA-256 of `ASCII("hub.mail.ruleset") || 0x00 || 0x01 || U32BE(count)`
followed by active non-deleted rules ordered `(priority,id)`. Each rule is a
`U32BE(record_length)` frame containing RFC-4122 UUID bytes, `U32BE(priority)`, `str(name)`,
then count + length-delimited condition records and count + length-delimited action records;
`str` is `U32BE(UTF-8 byte length)||bytes`. Condition field codes are
from=1/to=2/subject=3/has_attachment=4, operator codes equals=1/contains=2/
matches_domain=3/is=4, followed by value tag string=1 or boolean=2 and its canonical value.
Action codes are move=1/add_label=2/mark_read=3 followed by the 16 target UUID bytes only
for the first two. Stored array order is preserved; JSON encoding, map order, locale and
timestamps never participate. Any final service-owned account deletion is future/BLOCKED and
authorizes no current cascade; no engine script id/body is stored.
For equal priority, id order is unsigned lexicographic order of the 16 RFC-4122 UUID bytes;
database text/locale collation never determines the hash order.

#### 144-02D exact named row checks and signature-pointer transaction

`MailSignature` owns no default column. `is_default` is only a command over the canonical `MailUserSettings.default_signature_id`, and every response derives it after commit by equality with that pointer. The schema packet must emit these named row CHECKs with these literal SQL-like predicates:

| Name | Literal predicate |
|---|---|
| `mail_signatures_body_format` | `((body_format = 'rich' AND rich_body IS NOT NULL) OR (body_format = 'plain' AND rich_body IS NULL))` |
| `mail_signatures_name_bounds` | `(char_length(name) BETWEEN 1 AND 80 AND octet_length(name) BETWEEN 1 AND 320)` |
| `mail_signatures_plain_body_length` | `(char_length(plain_body) <= 100000 AND octet_length(plain_body) <= 400000)` |
| `mail_signatures_rich_body_length` | `(rich_body IS NULL OR (char_length(rich_body) <= 100000 AND octet_length(rich_body) <= 400000))` |

Create/update/delete lock the live signature (when it exists), the new live signature (when selected) and the unique settings row in one transaction. Create `is_default=false` leaves the pointer unchanged; create `true` selects it. Update `false` clears only when this signature is currently selected, and is otherwise a no-op; update `true` selects it and is a no-op when already selected. Deleting a selected signature clears the pointer before its soft delete. Increment settings `version` exactly iff the pointer changes. All selected signatures must be same-account and live; stale signature `If-Match` conflicts. No second caller settings PATCH is used to perform any of these commands.

#### 144-02F exact named row checks, recovery and deployment transaction invariants

The following are row CHECKs; each literal predicate is part of the candidate schema contract. Temporal, aggregate and cross-row guarantees following this table are service-transaction invariants, not claimed current Ent CHECKs.

| Name | Literal predicate |
|---|---|
| `mail_mutation_receipts_digest_version` | `(octet_length(idempotency_key_hash) = 32 AND octet_length(request_hash) = 32 AND fingerprint_version >= 1)` |
| `mail_mutation_receipts_target_pair` | `((target_kind IS NULL AND target_id IS NULL) OR (target_kind IS NOT NULL AND target_id IS NOT NULL))` |
| `mail_mutation_receipts_argument_pair` | `((argument_kind IS NULL AND argument_id IS NULL) OR (argument_kind IS NOT NULL AND argument_id IS NOT NULL))` |
| `mail_mutation_receipts_http_status_range` | `(original_http_status IS NULL OR original_http_status BETWEEN 100 AND 599)` |
| `mail_mutation_receipts_safe_code_closed_union` | `(safe_code IS NULL OR safe_code IN ('MAIL_FOLDER_NAME_CONFLICT','MAIL_FOLDER_TREE_CONFLICT','MAIL_FOLDER_LIMIT_EXCEEDED','MAIL_FOLDER_RECONCILIATION_REQUIRED','MAIL_RECONCILIATION_EXHAUSTED','MAIL_LABEL_NAME_CONFLICT','MAIL_LABEL_LIMIT_EXCEEDED','MAIL_SIGNATURE_LIMIT_EXCEEDED','MAIL_RULE_LIMIT_EXCEEDED','MAIL_DOMAIN_UNAVAILABLE','MAIL_DOMAIN_LIMIT_EXCEEDED','MAIL_ALIAS_ADDRESS_UNAVAILABLE','MAIL_ALIAS_ACCOUNT_NOT_ACTIVE','MAIL_IDENTITY_LIMIT_EXCEEDED','MAIL_BULK_EXPANSION_LIMIT','MAIL_ENGINE_OUTCOME_UNKNOWN'))` |
| `mail_mutation_receipts_reconcile_attempts_bounds` | `(reconcile_attempts BETWEEN 0 AND 10)` |
| `mail_mutation_receipts_reconcile_lease_pair` | `((reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL) OR (reconcile_lease_token IS NOT NULL AND reconcile_lease_until IS NOT NULL))` |
| `mail_mutation_receipts_terminal_replay_pair` | `((terminal_at IS NULL AND replay_until IS NULL) OR (terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days'))` |
| `mail_mutation_receipts_compaction` | `(compacted_at IS NULL OR (terminal_at IS NOT NULL AND compacted_at >= replay_until))` |
| `mail_mutation_receipts_in_progress_shape` | `(state <> 'in_progress' OR (safe_code IS NULL AND original_http_status IS NULL AND terminal_at IS NULL AND replay_until IS NULL AND compacted_at IS NULL))` |
| `mail_mutation_receipts_unknown_shape` | `(state <> 'unknown' OR (safe_code IN ('MAIL_FOLDER_RECONCILIATION_REQUIRED','MAIL_ENGINE_OUTCOME_UNKNOWN') AND next_reconcile_at IS NOT NULL AND original_http_status IS NULL AND terminal_at IS NULL AND replay_until IS NULL AND compacted_at IS NULL))` |
| `mail_mutation_receipts_completed_shape` | `(state <> 'completed' OR (safe_code IS NULL AND original_http_status BETWEEN 200 AND 299 AND terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days' AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL))` |
| `mail_mutation_receipts_failed_shape` | `(state <> 'failed' OR (safe_code IS NOT NULL AND original_http_status BETWEEN 400 AND 599 AND terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days' AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL))` |
| `mail_mutation_receipts_reconciliation_exhausted_terminal` | `COALESCE(safe_code IS DISTINCT FROM 'MAIL_RECONCILIATION_EXHAUSTED' OR (state = 'failed' AND original_http_status = 409 AND terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days' AND reconcile_attempts = 10 AND compacted_at IS NULL AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL), FALSE)` |
| `mail_mutation_receipts_operation_bulk_action` | `(operation = 'message_bulk' OR bulk_action IS NULL)` |
| `mail_mutation_receipts_operation_target_status_safe_code` | `COALESCE(CASE operation WHEN 'folder_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND (target_kind IS NULL OR target_kind = 'folder')) OR (state = 'unknown' AND target_kind = 'folder' AND safe_code = 'MAIL_FOLDER_RECONCILIATION_REQUIRED' AND original_http_status IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'folder' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind = 'folder' AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code IN ('MAIL_FOLDER_NAME_CONFLICT','MAIL_FOLDER_TREE_CONFLICT','MAIL_RECONCILIATION_EXHAUSTED') AND original_http_status = 409) OR (safe_code = 'MAIL_FOLDER_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'folder_update' THEN (target_kind = 'folder' AND argument_kind IS NULL AND argument_id IS NULL AND ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL) OR (state = 'unknown' AND safe_code = 'MAIL_FOLDER_RECONCILIATION_REQUIRED' AND original_http_status IS NULL) OR (state = 'completed' AND safe_code IS NULL AND original_http_status = 200) OR (state = 'failed' AND safe_code = 'MAIL_RECONCILIATION_EXHAUSTED' AND original_http_status = 409))) WHEN 'folder_delete' THEN (target_kind = 'folder' AND argument_kind IS NULL AND argument_id IS NULL AND ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL) OR (state = 'unknown' AND safe_code = 'MAIL_FOLDER_RECONCILIATION_REQUIRED' AND original_http_status IS NULL) OR (state = 'completed' AND safe_code IS NULL AND original_http_status = 204) OR (state = 'failed' AND safe_code = 'MAIL_RECONCILIATION_EXHAUSTED' AND original_http_status = 409))) WHEN 'label_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'label' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code = 'MAIL_LABEL_NAME_CONFLICT' AND original_http_status = 409) OR (safe_code = 'MAIL_LABEL_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'signature_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'signature' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND safe_code = 'MAIL_SIGNATURE_LIMIT_EXCEEDED' AND original_http_status = 422 AND argument_kind IS NULL AND argument_id IS NULL)) WHEN 'rule_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'rule' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND safe_code = 'MAIL_RULE_LIMIT_EXCEEDED' AND original_http_status = 422 AND argument_kind IS NULL AND argument_id IS NULL)) WHEN 'attachment_presign' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'attachment' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL)) WHEN 'domain_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'domain' AND safe_code IS NULL AND original_http_status = 202 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code = 'MAIL_DOMAIN_UNAVAILABLE' AND original_http_status = 409) OR (safe_code = 'MAIL_DOMAIN_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'alias_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'identity' AND safe_code IS NULL AND original_http_status = 202 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code IN ('MAIL_ALIAS_ADDRESS_UNAVAILABLE','MAIL_ALIAS_ACCOUNT_NOT_ACTIVE') AND original_http_status = 409) OR (safe_code = 'MAIL_IDENTITY_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'message_bulk' THEN (target_kind IS NULL AND target_id IS NULL AND ((bulk_action = 'move' AND argument_kind = 'folder' AND argument_id IS NOT NULL) OR (bulk_action IN ('add_label','remove_label') AND argument_kind = 'label' AND argument_id IS NOT NULL) OR (bulk_action IN ('mark_read','mark_unread','flag','unflag','archive','junk','not_junk','trash','restore','pin','unpin') AND argument_kind IS NULL AND argument_id IS NULL)) AND ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL) OR (state = 'unknown' AND safe_code = 'MAIL_ENGINE_OUTCOME_UNKNOWN' AND original_http_status IS NULL) OR (state = 'completed' AND safe_code IS NULL AND original_http_status = 200) OR (state = 'failed' AND safe_code = 'MAIL_BULK_EXPANSION_LIMIT' AND original_http_status = 422))) ELSE FALSE END, FALSE)` |
| `mail_mutation_receipts_operation_key_source_account` | `((operation IN ('folder_update','folder_delete') AND key_source = 'internal' AND mail_account_id IS NOT NULL) OR (operation = 'domain_create' AND key_source = 'client' AND mail_account_id IS NULL) OR (operation IN ('folder_create','label_create','signature_create','rule_create','attachment_presign','alias_create','message_bulk') AND key_source = 'client' AND mail_account_id IS NOT NULL))` |

The named operation predicate is exhaustive across operation, state, target, HTTP status, safe code and argument; `mail_mutation_receipts_operation_bulk_action` makes its non-bulk `bulk_action IS NULL` requirement explicit, and it is deliberately separate from `mail_mutation_receipts_operation_key_source_account`. For the three folder operations, create may lack its target only while pre-reservation `in_progress`; unknown, completed and exhausted terminal rows require a folder target. `MailFolderCreateConflict` and `MailFolderMutationConflict` are the folder-only OpenAPI 409 components that expose `MAIL_RECONCILIATION_EXHAUSTED` as the generic NFR-13 code with no internal detail; `MailMutationIdempotencyConflict` excludes it.

Target CHECKs are: `request_index BETWEEN 0 AND 199`; `affected_count`, `succeeded_count` and `failed_count` each `BETWEEN 0 AND 200`; `succeeded_count + failed_count <= affected_count`; and the exact status/outcome and training alternatives in the target matrix above. Child equality, request-index contiguity, aggregate distinct-members `<= 200`, and one-member cardinality are not row CHECKs: the service enforces them under the locked parent transaction (or only a later accepted migration may add a composite FK). No current Ent CHECK is claimed for them.

Define `prior_attempts p` as the persisted `reconcile_attempts` observed by the recovery claimant. For `0 <= p < 10`, it atomically claims the receipt plus account lease/epoch and persists `p + 1` before exactly one bounded recovery I/O slot; `p = 9` is the tenth slot. Preclaim, poll, busy contention and pre-engine failure do not increment. A live tenth lease is contention/no-op. `p = 10` grants no I/O: after its lease expires, a takeover terminalizes without I/O as `failed`/`MAIL_RECONCILIATION_EXHAUSTED`/HTTP 409, sets `terminal_at` and the exact 30-day replay window, clears receipt and account leases, and advances the account end epoch. The tenth owner may complete normally; its crash or unknown result exhausts and permits no eleventh I/O. Only an authorized matching same-key/same-request replay inside that window receives generic NFR-13 error; after it expires it receives `IDEMPOTENCY_REPLAY_EXPIRED`. No attempt, receipt, lease, engine or provider detail is exposed. A fresh/different key cannot bypass unresolved folder control/error; it never re-expands, retrains or resubmits. This rule excludes drafts, submissions, authorization/not-found, and classifier feedback.

`MailRuleDeployment` emits these row CHECKs: desired/attempted/deployed tuples are each all-or-none; present generation/version are positive and every present hash satisfies `octet_length(hash) = 32`; deployed generation is `<= desired_generation`; lease is paired; `attempt_count BETWEEN 0 AND 10`; and `last_error_code`/`next_retry_at` follow the stated allowlist. State predicates are: `pending` has no attempted tuple or lease; `applying` has attempted tuple and lease; `ready` has no attempted tuple/lease/error/retry and deployed exactly equals desired; `unknown` retains attempted tuple and its unknown code; `error` retains attempted tuple and allowlisted error. In every non-ready state deployed is null or not equal to current desired. Applying/unknown/error retain immutable attempted data; a newer desired never rewrites it. The claim transaction persists attempted before I/O; only confirmed reconciliation copies attempted to deployed, becoming ready only if it equals current desired, otherwise pending. No raw Sieve, script id/body or provider text is persisted.

### 144-02F root correction — NULL-safe receipt/deployment and rule-set hash v1

This correction is normative persistence design only. It replaces every earlier shorthand or
NULL-vulnerable receipt/deployment predicate in this section; no Ent schema, generated output,
migration, service, executable test, or runtime authority follows.

Correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition
`2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` /
`EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M
`0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle is
`CONSUMED/HISTORICAL`; launcher/executable status remains `NOT_READY` / `NOT_RUN`. This lifecycle banner
does not modify any normative predicate, state block, vector, or QA oracle row below.

The earlier 900–930 named-check table and its recovery/deployment paragraphs are
`SUPERSEDED/HISTORICAL` wherever they differ from this correction, including their NULL-sensitive
comparison, `(receipt_id,target_id)` unique, abbreviated state wording, retry/unknown caps and
old lifecycle. They are retained only as audit history; the following literal predicates and
state transitions are the sole active 02F contract.

`MailMutationReceipt` digest/version validity is exactly
`COALESCE(octet_length(idempotency_key_hash) = 32 AND octet_length(request_hash) = 32 AND fingerprint_version >= 1, FALSE)`.
The paired terminal form is exactly `COALESCE((terminal_at IS NULL AND replay_until IS NULL) OR (terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days'), FALSE)`;
compaction is exactly `COALESCE(compacted_at IS NULL OR (terminal_at IS NOT NULL AND replay_until IS NOT NULL AND compacted_at >= replay_until), FALSE)`.
All state shape validation is one exhaustive fail-closed check:

```sql
mail_mutation_receipts_state_shape:
COALESCE(CASE state
  WHEN 'in_progress' THEN safe_code IS NULL AND original_http_status IS NULL
    AND terminal_at IS NULL AND replay_until IS NULL AND compacted_at IS NULL
  WHEN 'unknown' THEN original_http_status IS NULL
    AND terminal_at IS NULL AND replay_until IS NULL AND compacted_at IS NULL
    AND ((reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL)
      OR (reconcile_lease_token IS NOT NULL AND reconcile_lease_until IS NOT NULL))
    AND ((operation IN ('folder_create','folder_update','folder_delete')
          AND safe_code='MAIL_FOLDER_RECONCILIATION_REQUIRED'
          AND reconcile_attempts BETWEEN 1 AND 10 AND next_reconcile_at IS NOT NULL)
      OR (operation='message_bulk' AND safe_code='MAIL_ENGINE_OUTCOME_UNKNOWN'
          AND reconcile_attempts BETWEEN 1 AND 9 AND next_reconcile_at IS NOT NULL)
      OR (operation='message_bulk' AND safe_code='MAIL_ENGINE_OUTCOME_UNKNOWN'
          AND reconcile_attempts=10 AND next_reconcile_at IS NULL
          AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL))
  WHEN 'completed' THEN safe_code IS NULL AND original_http_status BETWEEN 200 AND 299
    AND terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days'
    AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL
  WHEN 'failed' THEN safe_code IS NOT NULL AND original_http_status BETWEEN 400 AND 599
    AND terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days'
    AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL
  ELSE FALSE END, FALSE)
```

`reconcile_attempts` is `0..10`; an unknown recovery row is exactly `1..10`. The earlier
NULL-sensitive receipt-exhaustion spelling is superseded: it is NULL-vulnerable
and is not a valid 02F predicate. The only terminal predicate is:

```sql
mail_mutation_receipts_reconciliation_exhausted_terminal:
COALESCE(safe_code IS DISTINCT FROM 'MAIL_RECONCILIATION_EXHAUSTED' OR
  (state='failed' AND original_http_status=409 AND reconcile_attempts=10
   AND terminal_at IS NOT NULL AND replay_until=terminal_at + INTERVAL '30 days'
   AND compacted_at IS NULL AND next_reconcile_at IS NULL
   AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL), FALSE)
```

For a folder receipt, a claimant with `p < 10` atomically claims/fences, stores `p + 1`
and its lease before one bounded recovery I/O slot; `p=9` is the tenth slot. `p=10` with
`state=unknown`, `safe_code=MAIL_FOLDER_RECONCILIATION_REQUIRED` and a due/expired claim is
no-I/O only: it becomes failed/409 `MAIL_RECONCILIATION_EXHAUSTED`, writes the 30-day
terminal/replay pair, clears both leases and advances the account epoch. A live tenth lease
is contention/no-op; its owner may finish, but crash/unknown permits no eleventh I/O.

For `message_bulk`, `p=1..9` `MAIL_ENGINE_OUTCOME_UNKNOWN` remains reconcilable with a due
`next_reconcile_at` and the ordinary paired lease while owned or both-null while waiting.
`p=10` unknown is never folder exhaustion: it remains manual-repair-only
`unknown` with `safe_code=MAIL_ENGINE_OUTCOME_UNKNOWN`, `reconcile_attempts=10`, both leases
NULL and `next_reconcile_at=NULL`; the scheduler neither reclaims it nor performs I/O. Exact
same-key/same-request replay resolves first and returns the persisted `202` snapshot; same key
with a different request remains standard `409`. For a fresh key, while `MailAccount` is locked
and before creating a receipt, expansion, snapshot or I/O, query a live manual-only row exactly
as `operation='message_bulk'`, `state='unknown'`, `safe_code='MAIL_ENGINE_OUTCOME_UNKNOWN'`,
`reconcile_attempts=10`, `next_reconcile_at IS NULL`, terminal/replay/compaction all NULL and
both receipt leases NULL. Its existence returns existing OpenAPI `423 MAIL_ACCOUNT_BUSY` with a
bounded `Retry-After`, creates no receipt/snapshot and performs no expansion/I/O. The account
lock makes the race deterministic; the guard persists until manual repair moves that receipt out
of the exact shape. No lease leak, bypass, or eleventh I/O is permitted.

`operation_bulk_action` is fail-closed:
`COALESCE((operation='message_bulk' AND bulk_action IN ('move','add_label','remove_label','mark_read','mark_unread','flag','unflag','archive','junk','not_junk','trash','restore','pin','unpin')) OR (operation<>'message_bulk' AND bulk_action IS NULL), FALSE)`.
`operation_key_source_account` is also fail-closed:
`COALESCE((operation IN ('folder_update','folder_delete') AND key_source='internal' AND mail_account_id IS NOT NULL) OR (operation='domain_create' AND key_source='client' AND mail_account_id IS NULL) OR (operation IN ('folder_create','label_create','signature_create','rule_create','attachment_presign','alias_create','message_bulk') AND key_source='client' AND mail_account_id IS NOT NULL), FALSE)`.
`operation_target_status_safe_code` is the complete literal named predicate published in the
named-check table above, with every operation arm written there; that table is superseded below only where this
correction expressly changes an arm: folder `p=10` exhaustion and manual-only `message_bulk`
unknown. Its nonbulk arms require `bulk_action IS NULL`. The only public exhaustion is the
frozen folder-create/folder-mutation 409 component; generic idempotency excludes it.

`MailMutationReceiptTarget` is a physical row: `id`, `receipt_id`, `request_index`,
`target_type`, `target_id`, `status`, `affected_count`, `succeeded_count`, `failed_count`,
`training_status`, `created_at` and `updated_at` are required/not-null; `safe_code` alone is
optional and maps to public `code`. Its only target uniques are `(receipt_id,request_index)`
and `(receipt_id,target_type,target_id)`, never `(receipt_id,target_id)`. Its literal
fail-closed row checks are the bounds and exhaustive status/outcome/training literal matrix
published above; no abbreviated branch stands as active contract. `MailMutationReceiptMember.id`, `receipt_id`, `target_id`, `message_id`,
`created_at` and `updated_at` are required/not-null, with unique `(receipt_id,message_id)` and
`(target_id,message_id)`. Parent/action relation, child equality, ordered contiguity and the
200-member aggregate are transaction-owned under the receipt lock; no fake composite FK is
claimed.

`MailRuleDeployment` desired tuple is required/non-null. Attempted and deployed triples are each
all-null or all-non-null, present hash length is 32, all generations/versions are positive,
`deployed_generation <= desired_generation`, lease fields are paired, `attempt_count BETWEEN 0 AND 10`,
and internal terminal code is exactly `MAIL_RULE_DEPLOYMENT_EXHAUSTED` when non-null. Its full
public retry shape is:

```sql
mail_rule_deployments_public_error_retry_shape:
COALESCE(CASE
 WHEN last_error_code IS NULL THEN TRUE
 WHEN last_error_code IN ('MAIL_RULE_CAPABILITY_UNAVAILABLE','MAIL_RULE_APPLY_REJECTED') THEN
   state='error' AND internal_terminal_code IS NULL AND attempt_count BETWEEN 1 AND 10 AND next_retry_at IS NULL
 WHEN last_error_code IN ('MAIL_RULE_ENGINE_UNAVAILABLE','MAIL_RULE_DRIFT') THEN
   state='error' AND internal_terminal_code IS NULL AND ((attempt_count BETWEEN 1 AND 9 AND next_retry_at IS NOT NULL) OR (attempt_count=10 AND next_retry_at IS NULL))
 WHEN last_error_code='MAIL_RULE_OUTCOME_UNKNOWN' THEN
   state='unknown' AND internal_terminal_code IS NULL AND attempt_count BETWEEN 1 AND 10 AND next_retry_at IS NOT NULL
 ELSE FALSE END,FALSE)
```

The sole full literal state predicate is:

```sql
mail_rule_deployments_state_shape:
COALESCE(CASE state
 WHEN 'pending' THEN attempted_generation IS NULL AND lease_token IS NULL
   AND last_error_code IS NULL AND internal_terminal_code IS NULL AND attempt_count=0
   AND NOT (deployed_generation IS NOT DISTINCT FROM desired_generation AND deployed_hash_version IS NOT DISTINCT FROM desired_hash_version AND deployed_hash IS NOT DISTINCT FROM desired_hash)
 WHEN 'applying' THEN attempted_generation IS NOT NULL AND lease_token IS NOT NULL
   AND last_error_code IS NULL AND internal_terminal_code IS NULL AND next_retry_at IS NULL
   AND attempt_count BETWEEN 1 AND 10
   AND NOT (deployed_generation IS NOT DISTINCT FROM desired_generation AND deployed_hash_version IS NOT DISTINCT FROM desired_hash_version AND deployed_hash IS NOT DISTINCT FROM desired_hash)
 WHEN 'ready' THEN attempted_generation IS NULL AND lease_token IS NULL
   AND last_error_code IS NULL AND internal_terminal_code IS NULL AND next_retry_at IS NULL
   AND attempt_count=0 AND deployed_generation IS NOT DISTINCT FROM desired_generation
   AND deployed_hash_version IS NOT DISTINCT FROM desired_hash_version AND deployed_hash IS NOT DISTINCT FROM desired_hash
 WHEN 'unknown' THEN attempted_generation IS NOT NULL AND lease_token IS NULL
   AND last_error_code='MAIL_RULE_OUTCOME_UNKNOWN' AND internal_terminal_code IS NULL
   AND attempt_count BETWEEN 1 AND 10 AND next_retry_at IS NOT NULL
   AND NOT (deployed_generation IS NOT DISTINCT FROM desired_generation AND deployed_hash_version IS NOT DISTINCT FROM desired_hash_version AND deployed_hash IS NOT DISTINCT FROM desired_hash)
 WHEN 'error' THEN attempted_generation IS NOT NULL AND lease_token IS NULL
   AND ((last_error_code IN ('MAIL_RULE_CAPABILITY_UNAVAILABLE','MAIL_RULE_APPLY_REJECTED') AND internal_terminal_code IS NULL AND attempt_count BETWEEN 1 AND 10 AND next_retry_at IS NULL)
     OR (last_error_code IN ('MAIL_RULE_ENGINE_UNAVAILABLE','MAIL_RULE_DRIFT') AND internal_terminal_code IS NULL AND ((attempt_count BETWEEN 1 AND 9 AND next_retry_at IS NOT NULL) OR (attempt_count=10 AND next_retry_at IS NULL)))
     OR (last_error_code IS NULL AND internal_terminal_code='MAIL_RULE_DEPLOYMENT_EXHAUSTED' AND attempt_count=10 AND next_retry_at IS NULL))
   AND NOT (deployed_generation IS NOT DISTINCT FROM desired_generation AND deployed_hash_version IS NOT DISTINCT FROM desired_hash_version AND deployed_hash IS NOT DISTINCT FROM desired_hash)
 ELSE FALSE END,FALSE)
```

Unknown is only `MAIL_RULE_OUTCOME_UNKNOWN`. At `p=10` unknown/crash, the takeover deadline
permits no I/O: if attempted equals current desired it becomes internal
`MAIL_RULE_DEPLOYMENT_EXHAUSTED`; if desired advanced it clears the old attempted tuple and
leaves the newer desired `pending` at attempt zero. A desired update changes only desired;
it never rewrites live attempted/deployed/lease/budget, so an old completion may store
deployed-old and leave pending-new at zero. Known capability/apply rejection is terminal public
error at counts 1..10 without retry; engine unavailable/drift retry at 1..9 and are terminal
public error at 10. No eleventh read/install cycle exists.

Every claim uses one injected DB time `T`, fresh UUIDs `A` (account), `R` (receipt), `D`
(deployment), locks `MailAccount` first then one control row (receipt before deployment when
both) and children ascending `(request_index,raw UUID)`; no engine I/O occurs in the DB
transaction. Account claim guards id/live/expected `mutation_epoch` and absent-or-expired lease,
using common `T+60s` for account and control-row deadline. Receipt claim additionally guards
id/account/state/`p`/terminal+replay+compaction/due/expired lease, persists `p+1,R,T+60s,A` and
`begin_epoch` before one I/O. Completion guards state/count/R/A/begin_epoch and both unexpired
leases, atomically stores result, clears leases and increments epoch; zero rows drop a late
result. Deployment claim/completion analogously guard state/count/attempted tuple/D/A/begin_epoch.
Pre-I/O compensation uses those exact guards only. Stale token, expired lease and ABA
epoch/token reuse are explicit no-write/no-I/O oracles.

#### Rule-set canonical hash v1

The header is exactly the 16 ASCII bytes `hub.mail.ruleset` then `0x00 || 0x01 ||
U32BE(active_count)`. Active means `enabled=true AND deleted_at IS NULL`; sort priority
ascending then raw UUID16 unsigned lexicographic. Every count, frame length, UTF-8 string byte
length and priority is U32BE. A rule frame is `UUID16||U32BE(priority)||str(name)||U32BE(condition_count)||condition_frames||U32BE(action_count)||action_frames`, nested under `U32BE(frame_len)`. `condition_frames` is the concatenation in stored array order of `U32BE(condition_frame_len)||condition_frame`; `action_frames` is the same stored-order concatenation of `U32BE(action_frame_len)||action_frame`. Every nested frame consumes exactly its stated length and permits no trailing byte.
A condition frame is `U8(field)||U8(operator)||U8(value_tag)||value`: fields
from=1,to=2,subject=3,has_attachment=4; operators equals=1,contains=2,matches_domain=3,is=4;
tags string=1 (`str`) or bool=2 (`00|01`). Only closed field/operator/value combinations are
valid. An action frame is `U8(action)||U8(target_present)||target?`: move=1 and add_label=2
require `01||UUID16(nonzero)`, mark_read=3 requires `00` and no following bytes. Stored arrays
remain ordered. No JSON, locale, time or provider value participates.

Literal future static-oracle vectors (independently recomputed by V8 pure-JS and .NET SHA256;
not runtime/test evidence; launcher `NOT_READY`, executable evidence `NOT_RUN`) are:

| fixture | full preimage hex | bytes | SHA-256 |
|---|---|---:|---|
| empty | `6875622e6d61696c2e72756c65736574000100000000` | 22 | `04886ba512485ccfa2a66d5769ff8772ade648dec3c216293bc226f5c51b4e68` |
| one: id `00112233-4455-6677-8899-aabbccddeeff`, priority 2, name A, subject contains x, mark_read | `6875622e6d61696c2e72756c657365740001000000010000003300112233445566778899aabbccddeeff0000000200000001410000000100000008030201000000017800000001000000020300` | 77 | `5d08cfaba9849d4886666cdc3752588426fb96603c82a49ffee483594907c871` |
| two canonical UUID order: low id `00112233-4455-6677-8899-aabbccddeeff`, priority 7, name B, has_attachment is true + subject contains x, move `aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee` + add_label `11111111-2222-3333-4444-555555555555`; high id `10213243-5465-7687-8899-aabbccddeeff`, priority 7, name C, subject equals z, mark_read | `6875622e6d61696c2e72756c657365740001000000020000006100112233445566778899aabbccddeeff00000007000000014200000002000000040404020100000008030201000000017800000002000000120101aaaaaaaabbbbccccddddeeeeeeeeeeee000000120201111111112222333344445555555555550000003310213243546576878899aabbccddeeff0000000700000001430000000100000008030101000000017a00000001000000020300` | 178 | `80cd4ae7cff1967c8a25853e4786392cfbc562d86bec90fbd046e96877baf49a` |

The final two-rule literal is the authoritative 178-byte preimage supplied above; its semantic
fixture intentionally presents reverse input order but hashes low UUID then high UUID.

## Generation gate

The schema packet is eligible for implementation only after this document and OpenAPI 0.1.8
receive independent Terra/high review with accepted per-dispatch provenance. For this branch
and Phases 143–148, `dispatcher-enforced` is sufficient only for an exact native dispatch
accepted without rejection; actual telemetry, when exposed, must match. Ent generation,
schema validation, build/tests and migration artifacts remain `NOT_RUN` until the accepted
task-owned launcher and founder migration decision exist.

## 144-03B Rule compiler candidate dependency

Compiler neither calculates nor returns 04M-owned canonical hash v1; input action order remains
hash material. Labels use server-generated immutable unique keywords: `hubk-` plus canonical
lower-case unpadded RFC-4648 base32 of 32 random bytes, decode-to-32/re-encode exact. This does
does not authorize schemas, generation, Rule TDD or production. The Rule contract is
The pre-acceptance `f199546513b758ca7b95328b5ccd7680009b732e` / `6735ae746c7acde657359e8a709cd897c4dce61c` lifecycle is `CONSUMED/HISTORICAL`; broad `63e419897aa0683e1d3dfcd442aca78eea1b0727` and `6c424ed2d6703ceccbb269c37d9abc03e75a0410` are historical only. The standalone Rule plan is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4`, sole-path blob `b86c0423a307fd2673a3d322435922aa2a398e8f`, after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. It does not authorize schemas, generation, Rule TDD or production: Rule Task 1 is `NOT_AUTHORIZED/BLOCKED` until the independent accepted Phase-145 carrier/G145-5B/G145-5C/launcher prerequisites, without implied order; post-task runtime is `NOT_RUN`. Launcher `NOT_READY`; executable evidence `NOT_RUN`.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan-next text is `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. Accepted Rule plan `669ef761` does not authorize Task 1, which awaits independent carrier/G145-5B/G145-5C/launcher readiness; Task4 is post-future-Task3 source acceptance only. 144-01B remains Docker-blocked, Phase 142/148 and release pending, launcher `NOT_READY`, executables `NOT_RUN`.
