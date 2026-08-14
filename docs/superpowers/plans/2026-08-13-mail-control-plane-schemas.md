# Mail Control-Plane Schemas Implementation Plan

> **Current upstream lifecycle (2026-08-14):** The exact `55e` 144-02F closure attempt is `CONSUMED/HISTORICAL`: Contract/Spec/Oracle Terra/high FAIL C/I/M `0/1/0` (Rule hash v1 underdetermined), Security/Adversarial Terra/xhigh FAIL C/I/M `0/1/0` (PostgreSQL `NULL` bypass), Lifecycle/Dependency Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced, fallback none, telemetry `UNKNOWN`. The live exact 28-document correction is `ROOT_CORRECTION_CANDIDATE` / `PRECOMMIT_REVIEW_REQUIRED` / `INDEPENDENT_CLOSURE_NOT_DISPATCHABLE` / `IMPLEMENTATION_BLOCKED`. Exact next is precommit review ×3, root exact commit, exact-SHA review/transition, fresh 02F closure ×3. This plan remains blocked: no source/Ent/generation/migration/service/runtime authority and no Rule plan or Phase-145 carrier/G145-5B/G145-5C advancement.

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `superpowers:subagent-driven-development` to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the bounded typed DNS value and the four Phase 144 control-plane Ent
schemas without generating migrations or guessing lifecycle representation.

> **Representation receipt (not execution authority):** the frozen 144-02A A1/A2/A3
> representation at reviewed snapshot `b631866b2ba372101fd80eea30951eaa267c1931` is
> `STATIC_CONTRACT_READY` / docs-only accepted after Contract/Spec Terra/high,
> Security/Adversarial Terra/xhigh and Lifecycle/Dependency Terra/high PASS C/I/M `0/0/0`.
> Dispatcher-enforced, fallback none, telemetry `UNKNOWN` unless exposed. It does not authorize
> this plan's source/TDD execution, generated output, migration, launcher, tests/validators/build
> or runtime; founder migration authority, supported task-owned launcher, sequential explicit
> source dispatch and source-level review remain required. 144-02D is docs-only accepted; the sole
> former direct 02F-closure instruction is `CONSUMED/HISTORICAL`; active next action is the
> 28-doc correction precommit-review -> root-commit -> exact-SHA -> fresh-closure -> separate-sync
> sequence, not code.

**Architecture:** Three sequential Terra packets keep every intermediate source coherent.
The first packet adds a provider-neutral validated DNS JSON value. The second adds domain and
policy schemas that depend only on existing Organization. The third adds the mutually linked
account and identity schemas, then completes the narrow inverse edges. Row-local invariants
are named PostgreSQL CHECKs; cross-row quotas and exactly-one-primary remain explicit service
transaction responsibilities.

**Tech Stack:** Go 1.25, Ent v0.14.1, PostgreSQL CHECK/index annotations, standard library.

> **Lifecycle update:** 144-02D docs-only boundary `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition `e563a3d3be8898371e5c8a9861c58f11b72167bc` and docs-only acceptance sync `55e36b7a` is narrowly `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; direct 02F-closure instructions are `CONSUMED/HISTORICAL`. Active sole next action is the 28-doc correction precommit-review -> root-commit -> exact-SHA -> fresh-closure -> separate-sync sequence. This plan grants no source/schema/runtime authority; Phase145/Rule gates remain blocked.

## Global Constraints

- Contract base: `5f31a74f362f3d8f678fd3e3e5d403212dd119a1` plus the later docs-only
  plan commit containing this file and the nil/immutability clarification.
- The accepted `b631866b2ba372101fd80eea30951eaa267c1931` representation receipt consumes the
  historical independent contract-review prerequisite for this docs-only representation.
  Every product source packet remains gated by root-owned explicit source dispatch and
  independent `gpt-5.6-terra` / `high` source-level review with accepted per-dispatch provenance.
  The former direct-closure action is `CONSUMED/HISTORICAL`; the active global Mail action is the
  28-doc correction precommit-review -> root-commit -> exact-SHA -> fresh-closure -> separate-sync
  sequence, not 02A review or code.
- Worker runtime: exact `gpt-5.6-terra` / `xhigh`, no fallback. Reviewer runtime: exact
  `gpt-5.6-terra` / `high`, no fallback. For this branch/Phases 143–148, an explicit native
  dispatcher acceptance without rejection is `provenance_source: dispatcher-enforced`; actual
  telemetry, when exposed, must match.
- Founder migration authority and the task-owned Docker launcher remain separate gates.
  Schema source may not be executed until both are recorded.
- Do not read or change `testframework/**`, `backend/testframework/**`,
  `.planning/audits/test-rebase/**`, phases 128–141/TASK-193–234, `pro-qa-ai/**`,
  `.env.local`, `.git/index.lock`, generated Ent output, migration SQL or dependencies.
- Host tests/build/lint/validators/generation are forbidden. Direct Docker/Compose is not a
  fallback. Until the accepted launcher exists, all executable evidence is `NOT_RUN`.
- Each packet starts from a clean branch, writes the external genuine unit first, changes
  only its exact owned paths, returns an uncommitted handoff, and is root-reviewed/committed
  before the next packet starts.
- Exported Go comments are Russian. Commits contain no AI attribution.

---

## File Structure

- `backend/internal/mail/domain_dns.go` — persistence-safe DNS record value and validation.
- `backend/tests/mail/domain_dns_test.go` — external boundary tests for JSON shape and every
  count/byte/type/TTL/uniqueness limit.
- `backend/ent/schema/maildomain.go` — domain desired/deployed lifecycle and public DNS state.
- `backend/ent/schema/mailboxpolicy.go` — one bounded policy per organization.
- `backend/tests/mail/mail_domain_policy_schema_test.go` — exact field/index/check/edge
  metadata oracle for Packet A2.
- `backend/ent/schema/mailaccount.go` — account provisioning, projection and mutation fence.
- `backend/ent/schema/mailidentity.go` — primary/alias/group lifecycle.
- `backend/tests/mail/mail_account_identity_schema_test.go` — exact metadata oracle for
  Packet A3.
- `backend/ent/schema/organization.go`, `backend/ent/schema/user.go` — append-only inverse
  edges; all existing edge descriptors must remain byte-semantically unchanged.

---

### Task 1: 144-02A1 — typed required DNS value

**Files:**

- Create: `backend/tests/mail/domain_dns_test.go`
- Create: `backend/internal/mail/domain_dns.go`

**Public surface:**

```go
type RequiredDNSRecord struct {
	Type  EngineDNSRecordType `json:"type"`
	Name  string              `json:"name"`
	Value string              `json:"value"`
	TTL   int                 `json:"ttl"`
}

type RequiredDNSRecords []RequiredDNSRecord

func (records RequiredDNSRecords) Validate() error
```

- [ ] **Step 1: Prove the packet boundary before repository reads/writes**

Record:

```text
branch=codex/mail-module-20260813
contract_ancestor=5f31a74f362f3d8f678fd3e3e5d403212dd119a1
worktree=clean
requested_model=gpt-5.6-terra/xhigh
provenance_source=dispatcher-enforced
dispatcher_acceptance=accepted_without_rejection
actual_model=<matching telemetry|UNKNOWN>
actual_reasoning_effort=<matching telemetry|UNKNOWN>
independent_contract_review=PASS
allowed_paths=backend/internal/mail/domain_dns.go,backend/tests/mail/domain_dns_test.go
runtime=NOT_RUN
```

Any unavailable branch, ancestor, clean-state or independent-review line returns `BLOCKED`
before opening product files. A dispatcher rejection, omitted model/effort, fallback or
telemetry mismatch also returns `BLOCKED`; unavailable telemetry alone is `UNKNOWN` with
dispatcher-enforced provenance.

- [ ] **Step 2: Write the external test matrix first**

Package is `mail_test`. Use only `encoding/json`, `strings`, `testing`, and the public mail
package. The table must prove:

1. non-nil empty slice passes and marshals exactly `[]`;
2. one valid record of each `MX`, `TXT`, `CNAME` type passes;
3. marshaled object keys are exactly lower-case `type`, `name`, `value`, `ttl` and no others;
4. nil slice fails;
5. 33 records fail while 32 small unique records pass;
6. unknown and empty type fail;
7. TTL 59 and 86,401 fail while 60 and 86,400 pass;
8. 253-byte name passes and 254-byte name fails;
9. 2,048-byte value passes when aggregate JSON is below the cap and 2,049 bytes fails;
10. invalid UTF-8 in name or value fails;
11. equal `(type,name,value)` with a different TTL still fails as duplicate;
12. records with distinct triples pass;
13. a 32-record value whose canonical JSON exceeds 65,536 bytes fails.

No test imports Ent or relies on an implementation-private constant.

- [ ] **Step 3: Record static RED**

The two public types and method do not exist before implementation. Record
`TDD_RED=STATIC_EXPECTED`, `RUNTIME_RED=NOT_RUN`, and the launcher reason. Do not run Go.

- [ ] **Step 4: Implement the value exactly**

Use `encoding/json`, `errors`, and `unicode/utf8`. Define one unexported safe sentinel with
text `invalid required DNS records`; every invalid branch returns that same sentinel without
embedding a record value. `Validate` performs checks in this order:

1. nil slice;
2. count `<= MaxEngineDNSRecordItems`;
3. for each record: allowed type, UTF-8, name/value byte limits, TTL bounds, duplicate tuple;
4. `json.Marshal(records)` error;
5. serialized byte limit 65,536.

The duplicate key is `[3]string{string(record.Type), record.Name, record.Value}`; no joined
delimiter is allowed. Do not trim, normalize, resolve DNS or perform I/O.

- [ ] **Step 5: Static inspection and handoff**

Inspect only both owned paths with `rg`, `git diff --check`, `git diff --stat`, and a forbidden
API scan for filesystem/network/time/random/global mutable state. Return start SHA,
per-dispatch provenance (plus actual telemetry when exposed), both content hashes, diff summary, and
`compile/unit/runtime=NOT_RUN`. Do not commit.

- [ ] **Step 6: Root review and atomic commit**

Root confirms every test is non-tautological, JSON keys are exact, errors contain no input,
and only two paths changed. Commit with:

```text
feat(mail): add bounded DNS persistence value
```

Evidence status is `STATIC_CONFORMANT`, not runtime PASS.

---

### Task 2: 144-02A2 — domain and mailbox policy schemas

**Files:**

- Create: `backend/tests/mail/mail_domain_policy_schema_test.go`
- Create: `backend/ent/schema/maildomain.go`
- Create: `backend/ent/schema/mailboxpolicy.go`
- Modify: `backend/ent/schema/organization.go`

**Dependency:** accepted Task 1 commit.

- [ ] **Step 1: Prove a clean sequential boundary**

Require exact Terra/xhigh dispatcher-enforced provenance, Task 1 as an ancestor, clean status
and only the four owned paths. Capture the ordered pre-change `Organization.Edges()` descriptor
inventory.

- [ ] **Step 2: Write the external schema metadata oracle first**

Package is `mail_test`, importing `ent`, `entsql`, schema field/index/edge descriptors, and
`entschema`. Helpers must compare literal expected metadata to production descriptors; they
must not derive expectations from contract constants in production.

Assert exact `MailDomain` field order:

```text
id, organization_id, domain, engine_provisioning_key, engine_domain_id,
desired_present, enabled, is_default, state, verification_status, dkim_status,
required_dns_records, desired_generation, desired_hash_version, desired_hash,
attempted_generation, attempted_hash_version, attempted_hash,
deployed_generation, deployed_hash_version, deployed_hash,
last_error_code, next_retry_at, attempt_count, lease_token, lease_until,
created_at, updated_at, deleted_at
```

Assert exact `MailboxPolicy` field order:

```text
id, organization_id, external_mail_enabled, max_message_bytes, max_attachment_bytes,
max_recipients, send_rate_per_minute, trash_retention_days, junk_retention_days,
created_at, updated_at
```

The oracle also verifies:

- UUID/default/immutability/nullability and every scalar type/default/enum from the contract;
- all required/nullable SHA descriptors have exact byte validators and sizes;
- `required_dns_records` has the `mail.RequiredDNSRecords` Go type, is required, and has a
  non-nil empty default;
- Domain uses `VersionMixin`; Policy uses `VersionMixin`;
- the exact ordered indexes and partial predicates in the contract;
- `MailDomain.Annotations().Checks` contains exactly the 14 named predicates in the contract;
- `MailboxPolicy.Annotations().Checks` contains exactly `mailbox_policies_bounds`;
- A2 Domain edges are only required unique `organization` on `organization_id`, inverse
  ref `mail_domains`, with `entsql.Cascade`; `accounts` is absent until A3;
- Policy has only required unique `organization` on `organization_id`, inverse ref
  `mailbox_policy`, with `entsql.Cascade`;
- existing Organization edge descriptors remain in the same order and unchanged; only
  to-many `mail_domains` and unique `mailbox_policy` are appended in that order.

- [ ] **Step 3: Record static RED**

Both schemas and new inverse edges are absent. Record static expected compile failure and
runtime `NOT_RUN`; do not execute Go/Ent.

- [ ] **Step 4: Implement `MailDomain` fields**

Use the exact order above. Required details:

- UUID id defaults to `uuid.New` and is immutable; `organization_id`, `domain`, and
  `engine_provisioning_key` are immutable;
- `domain` uses the frozen lower-case ASCII domain validator, `MinLen(1).MaxLen(253)`;
- opaque strings use `MinLen(1).MaxLen(1024)` when present;
- booleans default `true`, `true`, `false`; lifecycle enums/defaults are exact;
- DNS JSON uses `mail.RequiredDNSRecords{}` and that same non-nil default;
- desired generation/version default to 1 and are positive; desired hash is exactly 32 bytes;
- attempted/deployed triples and all optional scalars are `Optional().Nillable()`;
- attempt count is int32 default 0/non-negative; timestamps use `nowUTC`; deleted time is
  optional/nillable; `VersionMixin` is the only mixin.

Define the domain validator in this file. It rejects invalid UTF-8, uppercase/non-ASCII,
empty labels, labels over 63 bytes, leading/trailing hyphen and any character outside
`a-z0-9-`; it never transforms input.

- [ ] **Step 5: Add exact Domain checks, indexes and A2 edge**

Copy the 14 names/predicates literally from the accepted contract into one
`entsql.Checks(map[string]string{...})`. Add only the five ordered indexes frozen there.
Add only the Organization owner edge; do not reference `MailAccount.Type` yet.

- [ ] **Step 6: Implement `MailboxPolicy`**

All policy fields are required with no defaults. Use int64 for byte ceilings, int32 for the
remaining numbers, matching `.Min/.Max` validators and the exact named SQL CHECK. Add
timestamps, `VersionMixin`, one unique organization index, and the cascading owner edge.

- [ ] **Step 7: Append only two Organization inverse edges**

Do not reorder or rewrite existing edges/comments. Append:

```go
edge.To("mail_domains", MailDomain.Type),
edge.To("mailbox_policy", MailboxPolicy.Type).Unique(),
```

- [ ] **Step 8: Static inspection and handoff**

Compare pre/post Organization descriptors, exact field/index/check/edge literals, imports,
forbidden paths and diff ownership. `git diff --check` must be clean. Return four path hashes
and `generation/compile/unit/runtime=NOT_RUN`; do not commit.

- [ ] **Step 9: Root review and atomic commit**

Root rejects an `accounts` edge, missing named predicate, map/secret JSON, policy default,
generated file or migration. Commit:

```text
feat(mail): add domain and mailbox policy schemas
```

---

### Task 3: 144-02A3 — account and identity schemas

**Files:**

- Create: `backend/tests/mail/mail_account_identity_schema_test.go`
- Create: `backend/ent/schema/mailaccount.go`
- Create: `backend/ent/schema/mailidentity.go`
- Modify: `backend/ent/schema/maildomain.go`
- Modify: `backend/ent/schema/organization.go`
- Modify: `backend/ent/schema/user.go`

**Dependency:** accepted Task 2 commit.

- [ ] **Step 1: Prove the final sequential boundary**

Require clean status, Tasks 1–2 ancestors, exact Terra/xhigh dispatcher-enforced provenance,
and only the six owned paths. Capture pre-change ordered edge descriptors for Domain,
Organization and User.

- [ ] **Step 2: Write the external schema metadata oracle first**

Assert exact `MailAccount` field order:

```text
id, organization_id, user_id, mail_domain_id, engine_provisioning_key, local_part,
engine_account_id, desired_enabled, status, control_generation, control_hash_version,
control_hash, attempted_control_generation, attempted_control_hash_version,
attempted_control_hash, deployed_generation, deployed_hash_version, deployed_hash,
projection_state, sync_cursor, rebuild_checkpoint, rebuild_base_cursor, sync_epoch,
mutation_epoch, provision_attempts, sync_attempts, provision_lease_token,
sync_lease_token, mutation_lease_token, provision_lease_until, sync_lease_until,
mutation_lease_until, next_provision_at, next_sync_at, last_synced_at,
provision_error_code, sync_error_code, event_cutoff_at, created_at, updated_at
```

Assert exact `MailIdentity` field order:

```text
id, mail_account_id, normalized_address, display_name, kind, send_allowed, status,
engine_identity_id, created_at, updated_at
```

The oracle must also prove:

- exact types/defaults/immutability/nullability/enum allowlists and string/hash sizes;
- Account has no mixin; Identity has no mixin;
- Account has exactly the six ordered indexes and 13 exact named CHECK predicates;
- Identity has exactly the four ordered indexes and three exact named CHECK predicates;
- Account edges are organization/user/domain required unique inverses on their explicit FK
  fields with Cascade/Cascade/NoAction respectively, then to-many identities;
- Identity has only the required unique cascading account inverse;
- Domain appends only `accounts`; Organization appends only `mail_accounts` after the A2
  additions; User appends only unique `mail_account`; all prior descriptors are unchanged;
- the local-part/address/display validators accept exact maxima and reject non-ASCII,
  uppercase, invalid dot-atoms, invalid domain labels, invalid UTF-8, 161 code points or
  641 display-name bytes as applicable.

- [ ] **Step 3: Record static RED**

Account/Identity types and inverse edges do not exist. Record static expected failure and
runtime `NOT_RUN`; no Go/Ent command runs.

- [ ] **Step 4: Implement shared normalized mailbox validators**

Keep unexported pure validators in `mailaccount.go`/`mailidentity.go` without I/O or
normalization. Lower-case `atext` is `a-z`, `0-9`, and the RFC symbols
`!#$%&'*+-/=?^_`, ASCII backtick, `{`, `|`, `}` and `~`; dot is allowed
only between non-dot atoms. Address validation requires one `@` and reuses the local-part and
Domain validator. Display-name validation requires valid UTF-8, at most 160 code points and
at most 640 bytes. All validation errors are constant and contain no input.

- [ ] **Step 5: Implement `MailAccount` exactly**

Use the field order above and the contract defaults. Owner/domain/provisioning/local fields
and `event_cutoff_at` are immutable. All opaque/cursor fields use exact byte limits. Every
nullable tuple/lease/time/error uses `Optional().Nillable()`. Counters/epochs use exact
int32/int64 widths and non-negative validators. Add the exact 13 named CHECK predicates and
six ordered indexes. Add the four exact edges; no unrelated inverse is permitted.

- [ ] **Step 6: Implement `MailIdentity` exactly**

Make `mail_account_id`, `normalized_address`, and `kind` immutable. `display_name` is
optional/nillable, `send_allowed` defaults false, status defaults pending, and the opaque
engine id is optional/nillable. Add three named checks, four indexes, timestamps and only the
cascading account edge. Do not store `is_default`, organization id, password or body data.

- [ ] **Step 7: Complete only the three inverse edge additions**

Append without reordering prior edges:

```go
// MailDomain
edge.To("accounts", MailAccount.Type),

// Organization, after A2 additions
edge.To("mail_accounts", MailAccount.Type),

// User
edge.To("mail_account", MailAccount.Type).Unique(),
```

- [ ] **Step 8: Static inspection and handoff**

Perform exact metadata literal scan, pre/post edge inventory comparison, forbidden-path scan,
`git diff --check`, and path/content hashes. Return
`generation/compile/unit/runtime=NOT_RUN`; do not commit.

- [ ] **Step 9: Root review and atomic commit**

Root rejects any missing explicit FK, wrong OnDelete, nullable tuple gap, accidental default,
secret/content field, existing-edge rewrite, generated output or migration. Commit:

```text
feat(mail): add account and identity schemas
```

---

## Acceptance and downstream gate

After all three commits, an independent Terra/high reviewer compares the exact source SHA to
the contract and reports P0/P1/P2 findings. Static acceptance requires zero P0/P1, resolved
P2 or explicit disposition, clean ownership, exact three-commit ancestry and no protected
path changes. It does not claim generated validity.

Only after founder migration authority and the supported task-owned Docker launcher are both
accepted may 144-02E generate Ent output and run the current-SHA schema/build/unit job. A
running Docker Desktop daemon alone is not that launcher. Until then:

```text
schema_source=STATIC_CONFORMANT_or_BLOCKED
ent_generation=NOT_RUN
migration=NOT_RUN
compile=NOT_RUN
unit=NOT_RUN
runtime=NOT_RUN
```
