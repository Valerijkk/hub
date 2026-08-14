# Mail persistence contract 0.1.8

> **Status:** five-document 0.1.8 entry amendment independently PASS at `5e314fb080d420696642058ae47397e251c23207`, authorizing only 144-01A entry; Thread contract is accepted at `cbc12577a6444df9c660e43fedfd0c8133c6e461`; implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS` (source `c221d547`, oracle `b21bd6fd`), static only; Phase 144 schemas and broader cross-contract review remain pending
> **Runtime/generation/migration:** NOT_RUN / NOT_RUN / NOT_AUTHORIZED
> **Scope:** Phase 144-02A/B/C/D/F Ent schemas only
> **Precedence:** this document supplies exact storage representation for the corresponding
> domain sections in `MAIL-SPEC.md`. It does not authorize generated code or migration SQL.

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
- Soft-delete uniqueness always includes `deleted_at IS NULL`. Hard account/organization
  deletion cascades owned Mail rows. Historical snapshot ids that must survive projection
  deletion are scalar UUIDs, not cascading edges.
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
| `status` | required enum, default `provisioning` |
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

- status: `provisioning|active|suspended|error`;
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
- list/control indexes cover status+next due time and projection state+next sync time.

All control workers persist the captured attempted generation/version/hash on the owner row
before I/O; a newer current generation cannot rewrite it or be finalized by the older result.
Successful completion copies exactly attempted to deployed, clears attempted/lease/error,
and reschedules when deployed is older than current. The account edge cascades from hard
organization/user deletion; its selected domain edge restricts domain deletion.

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
message. Organization deletion cascades. Enabling external mail still requires the
service-level verified-domain/engine gate.

### 144-02A Ent representation freeze

This subsection closes the remaining Go/Ent representation choices for 144-02A. The later
five-document entry amendment PASS at `5e314fb080d420696642058ae47397e251c23207` authorizes
only 144-01A entry; this schema subsection remains pending its own applicable review. It does
not authorize generation or migration.

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
  `(projection_state,next_sync_at)`.
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
  status = 'provisioning' OR (status = 'error' AND provision_error_code IS NOT NULL) OR (status = 'active' AND desired_enabled = TRUE AND engine_account_id IS NOT NULL AND deployed_generation IS NOT NULL) OR (status = 'suspended' AND desired_enabled = FALSE AND engine_account_id IS NOT NULL AND deployed_generation IS NOT NULL)
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
  `ON DELETE CASCADE`; a required unique edge to `MailDomain` with `ON DELETE NO ACTION`;
  and a to-many `identities` inverse.
- `MailIdentity` has one required unique `account` edge on `mail_account_id` with
  `ON DELETE CASCADE`.
- `MailDomain` has one required unique `organization` edge with `ON DELETE CASCADE` and a
  to-many `accounts` inverse. The account foreign key is the database RESTRICT/NO ACTION
  proof; confirmed engine absence and soft-delete finalization remain service checks.
- `MailboxPolicy` has one required unique `organization` edge with `ON DELETE CASCADE`.
- Existing `Organization` edges are preserved byte-semantically and append only
  `mail_accounts`, `mail_domains`, and unique `mailbox_policy`. Existing `User` edges are
  preserved byte-semantically and append only unique `mail_account`.

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
`normalized_name` (1..480 bytes), optional `color` (<=32 chars/128 bytes), required int32
`sort_order` default 0, required int32 `unread_count`/`total_count` default 0,
`control_state`, optional `control_receipt_id`, non-negative `last_seen_sync_epoch`,
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
`MAIL_FOLDER_IN_USE`. Hard account deletion cascades; parent delete is service-ordered and
restricted until children move/delete. The 100-row live quota counts reserved non-deleted
control rows, not only rows that already have an engine id. Projection/persistence maps opaque
engine folder IDs to Hub UUID rows before the pure Folder policy; it never passes engine IDs
to that policy.

### `MailThread`

Fields: `id`, `mail_account_id`, required immutable int32 `thread_key_version` default 1,
exact 32-byte `thread_key`, `normalized_subject`
(<=998 UTF-8 bytes), required concrete `participant_addresses` typed JSON array default empty
(exact `[]string` seed snapshot of unique canonical normalized addresses, live 1..401 items,
<=1,048,576 serialized UTF-8 bytes under default HTML-escaping JSON; no truncation/preview),
`last_message_at`, required int32 `message_count`/`unread_count`, required booleans
`is_pinned`/`has_attachments`, snippet
(<=1,200 chars, <=400 graphemes, <=4,096 bytes), required non-negative
`last_seen_sync_epoch` default 0, timestamps and optional `deleted_at`. Unique
`(mail_account_id,thread_key_version,thread_key)`; read index
`(mail_account_id,is_pinned,last_message_at,id)` on live rows. Account deletion cascades.
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
`flags_synced_at`, optional `calendar_method`, optional `calendar_event_id`, required
non-negative `last_seen_sync_epoch`, optional immutable `received_event_id`, timestamps and
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

Fields: UUID id/account, display name 1..60 chars/240 bytes, private normalized name <=240
bytes, color 1..32, immutable server-generated non-public `engine_keyword`: exactly 57 ASCII
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
UUID `draft_client_id`, upload state default `pending`, immutable expiry, timestamps. Exact owner
fields are `organization_id`, `mail_account_id`, `uploader_user_id`; size fields are
`expected_size_bytes` and nullable `actual_size_bytes`; hash is nullable `sha256` bytes(32).

Unique `(bucket,object_key)` and one object per id-derived key. State is
`pending|confirmed|orphaned|deleted`. Actual size/hash is a nullable pair; pending requires
both null; confirmed requires both present and actual=expected; orphaned/deleted retain any
known pair but cannot become pending. Expiry never changes. Account/org/uploader/staging
scope equality is transactionally validated. No scan field or project/file edge exists.

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
Reply/forward links set null on projected-message hard delete. Account deletion cascades;
individual draft deletion is the terminal status transition, not row deletion.

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

## 144-02D — user configuration

### `MailUserSettings`

One UUID/account row with required `default_identity_id`, optional
`default_signature_id`, required `remote_images` `blocked|trusted_senders` default
`blocked`, required `compose_format` `rich|plain` default `rich`, timestamps and
`VersionMixin`. Unique `mail_account_id`. The two defaults are canonical; public identity/signature
`is_default` is derived. Same-account, active/send-allowed identity and same-account live
signature are checked under a locked settings row. Selected identity deletion is restricted;
selected signature deletion clears the pointer and increments settings version atomically.
Account deletion cascades; identity/signature edges otherwise restrict until those ordered
service transitions complete.

### `MailSignature`

UUID/account, name 1..80 chars/320 bytes, body format, optional sanitized rich body and
required derived plain body, each <=100,000 chars/400,000 bytes, timestamps, optional
`deleted_at`, `VersionMixin`. Format is `rich|plain`; named CHECK requires rich non-null for
rich and null for plain. Exact fields are `id`, `mail_account_id`, `name`, `body_format`,
`rich_body`, `plain_body`, `created_at`, `updated_at`, `deleted_at` plus `VersionMixin`.
No default boolean exists. At most 20 live rows/account; the quota failure is terminally
stored in its receipt. Account deletion cascades; the settings pointer is cleared before a
selected signature is soft-deleted.

### `MailRule`

UUID/account, name 1..120 chars/480 bytes, enabled, priority 0..10,000, exact typed condition
array 1..20 and action array 1..10, combined serialized UTF-8 <=65,536 bytes, timestamps,
optional `deleted_at`, `VersionMixin`. JSON uses concrete structs matching OpenAPI and rejects
unknown fields; no `map[string]any`, raw Sieve, engine keyword/script id. Active canonical
ordering is `(priority,id)`. At most 100 live rows/account. Rule writes and referenced folder/
label deletes lock the account before validating JSON target UUIDs. Exact JSON fields are
stored as `conditions` and `actions`; each `move` owns one folder UUID, each `add_label` one
label UUID, and `mark_read` owns neither. A live index on
`(mail_account_id,enabled,priority,id)` drives hashing/deployment. Account deletion cascades;
soft-deleted targets are invalid and remain protected by the shared lock. Quota failure is
terminally stored in its receipt before a rule or desired deployment generation is written.

### `RecentContact`

UUID/account, normalized email 1..320 bytes, optional display name <=160 chars/640 bytes,
optional user id, `last_used_at`, non-negative saturating use count, timestamps. Unique
account+email and partial eviction index `(mail_account_id,last_used_at,id) WHERE user_id IS
NULL`. At most 1,000 rows with null `user_id` exist per account. Before admitting a new
external row, the same account-locked transaction evicts the oldest null-user row; Hub-linked
rows do not count toward the external-retention cap. If user deletion sets links null, that
transaction deterministically evicts oldest newly-external rows until the cap is restored.
`use_count` saturates at 2,147,483,647. Account deletion cascades.

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
| `reconcile_attempts` | required int32, default 0, bounded non-negative |
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
| `folder_create` | `folder` required after row reservation | none | 201 | `MAIL_FOLDER_NAME_CONFLICT|MAIL_FOLDER_TREE_CONFLICT|MAIL_FOLDER_LIMIT_EXCEEDED` / 409 or 422 as applicable | `MAIL_FOLDER_RECONCILIATION_REQUIRED` |
| `folder_update` | `folder` always | none; desired fields live on folder | 200 | none (preconditions occur before internal receipt) | `MAIL_FOLDER_RECONCILIATION_REQUIRED` |
| `folder_delete` | `folder` always | none | 204 | none (preconditions occur before internal receipt) | `MAIL_FOLDER_RECONCILIATION_REQUIRED` |
| `label_create` | `label` only on success | none | 201 | `MAIL_LABEL_NAME_CONFLICT|MAIL_LABEL_LIMIT_EXCEEDED` / 409 or 422 | forbidden |
| `signature_create` | `signature` only on success | none | 201 | `MAIL_SIGNATURE_LIMIT_EXCEEDED` / 422 | forbidden |
| `rule_create` | `rule` only on success | none | 201 | `MAIL_RULE_LIMIT_EXCEEDED` / 422 | forbidden |
| `attachment_presign` | `attachment` only after object reservation | none | 201 | none | forbidden |
| `domain_create` | `domain` only on success | none | 202 | `MAIL_DOMAIN_UNAVAILABLE|MAIL_DOMAIN_LIMIT_EXCEEDED` / 409 or 422 | forbidden on HTTP path; domain worker owns later state |
| `alias_create` | `identity` only on success | none | 202 | `MAIL_ALIAS_ADDRESS_UNAVAILABLE|MAIL_ALIAS_ACCOUNT_NOT_ACTIVE|MAIL_IDENTITY_LIMIT_EXCEEDED` / 409 or 422 | forbidden on HTTP path; account worker owns later state |
| `message_bulk` | parent target pair null; ordered child targets required | exact `bulk_action`; argument folder only for `move`, label only for add/remove-label, null otherwise | 200 | `MAIL_BULK_EXPANSION_LIMIT` / 422 | `MAIL_ENGINE_OUTCOME_UNKNOWN` only for reconcilable non-classifier engine state |

The persisted parent `safe_code` enum is exactly the union named by that matrix:
`MAIL_FOLDER_NAME_CONFLICT|MAIL_FOLDER_TREE_CONFLICT|MAIL_FOLDER_LIMIT_EXCEEDED|MAIL_FOLDER_RECONCILIATION_REQUIRED|MAIL_LABEL_NAME_CONFLICT|MAIL_LABEL_LIMIT_EXCEEDED|MAIL_SIGNATURE_LIMIT_EXCEEDED|MAIL_RULE_LIMIT_EXCEEDED|MAIL_DOMAIN_UNAVAILABLE|MAIL_DOMAIN_LIMIT_EXCEEDED|MAIL_ALIAS_ADDRESS_UNAVAILABLE|MAIL_ALIAS_ACCOUNT_NOT_ACTIVE|MAIL_IDENTITY_LIMIT_EXCEEDED|MAIL_BULK_EXPANSION_LIMIT|MAIL_ENGINE_OUTCOME_UNKNOWN`.
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
replay time and never accept an HTTP key. Lease pair is both-null or both-present. Account/
organization deletion cascades. Actor deletion never removes or blocks an organization
receipt because `actor_user_id` is a historical scalar snapshot rather than a user edge.

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
nullable `last_error_code`, nullable `next_retry_at`, required int32 `attempt_count` default
0, nullable `lease_token`/`lease_until`, timestamps. Hashes are bytes(32), generations int64
>=1, versions int32 >=1. Unique account; tuples and lease are independently all-or-none.

| State | Required | Forbidden / additional rule |
|---|---|---|
| `pending` | desired tuple; deployed null or older/different | attempted, lease and error null; retry optional |
| `applying` | attempted tuple and lease; attempted generation <= desired | error/retry null |
| `ready` | deployed tuple exactly equals desired | attempted, lease, error and retry null |
| `unknown` | attempted tuple, retry and `MAIL_RULE_OUTCOME_UNKNOWN|MAIL_RULE_ENGINE_UNAVAILABLE` | lease null while waiting; deployed does not advance |
| `error` | attempted tuple and one definitive/retryable error | lease null; deployed never equals desired unless state immediately becomes ready |

Error allowlist is exactly
`MAIL_RULE_CAPABILITY_UNAVAILABLE|MAIL_RULE_ENGINE_UNAVAILABLE|MAIL_RULE_APPLY_REJECTED|MAIL_RULE_OUTCOME_UNKNOWN|MAIL_RULE_DRIFT`.
Capability-unavailable/apply-rejected errors have null retry; engine-unavailable/drift errors
have non-null retry. A successful attempted tuple copies exactly to deployed, then clears
attempted: it becomes ready only when it equals current desired, otherwise pending. A newer
desired tuple never rewrites applying/unknown attempted state.

Canonical hash v1 is SHA-256 of `ASCII("hub.mail.ruleset") || 0x00 || 0x01 || U32BE(count)`
followed by active non-deleted rules ordered `(priority,id)`. Each rule is a
`U32BE(record_length)` frame containing RFC-4122 UUID bytes, `U32BE(priority)`, `str(name)`,
then count + length-delimited condition records and count + length-delimited action records;
`str` is `U32BE(UTF-8 byte length)||bytes`. Condition field codes are
from=1/to=2/subject=3/has_attachment=4, operator codes equals=1/contains=2/
matches_domain=3/is=4, followed by value tag string=1 or boolean=2 and its canonical value.
Action codes are move=1/add_label=2/mark_read=3 followed by the 16 target UUID bytes only
for the first two. Stored array order is preserved; JSON encoding, map order, locale and
timestamps never participate. Account deletion cascades; no engine script id/body is stored.
For equal priority, id order is unsigned lexicographic order of the 16 RFC-4122 UUID bytes;
database text/locale collation never determines the hash order.

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
not close 144-02D/02F or authorize schemas, generation, Rule TDD or production. The Rule contract is
docs-only `COMMITTED_UNACCEPTED_AT_f1995465`; Lifecycle/Dependency exact-SHA re-review remains required.
