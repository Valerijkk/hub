# Inbound published fixture external-closure manifest

Status: `FROZEN` / implementation `NOT_RUN` / Docker `NOT_RUN`.

This document freezes the complete direct Ent/Atlas closure that must be read
and locked before deleting the deterministic Organization and User roots of a
published inbound Mail/Calendar QA fixture. It is a compile-time manifest, not
permission to discover dependencies dynamically from PostgreSQL catalogs.

## Canonical policy representation

The canonical line format is:

```text
<root>|<storage-table>|<storage-column>|<mode>|<expected-row-id-source>|<expected-count>
```

Rules:

- `root` is exactly `Organization` or `User` and determines whether the scalar
  predicate is `storage-column = rootProof.organizationID` or
  `storage-column = rootProof.ownerUserID`.
- `storage-table` and `storage-column` are the current Ent edge storage names;
  each implementation query is the typed Ent equivalent of that exact scalar
  equality plus `FOR UPDATE` in the admitted `SERIALIZABLE` transaction.
- `forbidden|none|0` requires the exact selected set to be empty.
- `allowlisted|<source>|1` requires exactly one row and its row ID to equal the
  named proof value. Missing, duplicate, extra, foreign, partial, unobserved or
  query-failed results fail closed before the first delete.
- Sort is bytewise ascending by `root`, then `storage-table`, then
  `storage-column`. Encoding is UTF-8, LF-only, with one final LF.

Frozen six-field policy metadata, validated independently:

- line count: `82`
- canonical byte count: `4767`
- SHA-256: `889c14127f47bdb73108bfef07573ff4b0bde7fd3b78b47122718a1f5b0a36b7`
- authority: current Ent schemas plus the committed Atlas
  `mail-calendar-v1` migration chain; no runtime catalog scan

The schema authority is a separate derived three-field projection:

```text
<root>|<storage-table>|<storage-column>
```

Generated Atlas desired-state and full committed migration replay derive and
compare only that projection. The schema checker must never infer `mode`,
`expected-row-id-source` or `expected-count` from DDL.

## Exact canonical manifest

```text
Organization|calendar_availability_reservations|organization_id|forbidden|none|0
Organization|calendar_events|organization_id|allowlisted|calendarProof.calendarEventID|1
Organization|calendar_inbound_object_states|organization_id|allowlisted|calendarProof.inboundObjectStateID|1
Organization|calendar_invite_deliveries|organization_id|forbidden|none|0
Organization|calendar_mutation_receipts|organization_id|allowlisted|calendarProof.mutationReceiptID|1
Organization|calendar_reply_deliveries|organization_id|allowlisted|calendarProof.replyDeliveryID|1
Organization|mail_accounts|organization_id|allowlisted|rootProof.mailAccountID|1
Organization|mail_domains|organization_id|allowlisted|rootProof.mailDomainID|1
Organization|mail_mutation_receipts|organization_id|forbidden|none|0
Organization|mail_objects|organization_id|forbidden|none|0
Organization|mail_send_rate_buckets|organization_id|forbidden|none|0
Organization|mailbox_policies|organization_id|forbidden|none|0
Organization|projects|organization_projects|forbidden|none|0
Organization|users|organization_users|allowlisted|rootProof.ownerUserID|1
User|activity_logs|user_activity_logs|forbidden|none|0
User|ai_providers|user_ai_providers|forbidden|none|0
User|ai_usage_dailies|user_ai_usage_daily|forbidden|none|0
User|ai_user_preferences|user_ai_preference|forbidden|none|0
User|api_keys|user_api_keys|forbidden|none|0
User|automation_rules|created_by_id|forbidden|none|0
User|bookmarks|user_id|forbidden|none|0
User|calendar_attendees|user_id|allowlisted|calendarProof.rsvpAttendeeID|1
User|calendar_availability_reservations|user_id|forbidden|none|0
User|calendar_events|calendar_owner_id|allowlisted|calendarProof.calendarEventID|1
User|calendar_events|organizer_id|forbidden|none|0
User|calendar_inbound_object_states|calendar_owner_id|allowlisted|calendarProof.inboundObjectStateID|1
User|calendar_mutation_receipts|actor_user_id|allowlisted|calendarProof.mutationReceiptID|1
User|call_participants|user_id|forbidden|none|0
User|calls|created_by|forbidden|none|0
User|channel_invites|creator_id|forbidden|none|0
User|channel_members|user_id|forbidden|none|0
User|channels|created_by|forbidden|none|0
User|coffee_enrollments|user_coffee_enrollment|forbidden|none|0
User|coffee_pairs|user_coffee_pairs_a|forbidden|none|0
User|coffee_pairs|user_coffee_pairs_b|forbidden|none|0
User|coffee_rounds|user_coffee_rounds_triggered|forbidden|none|0
User|comments|user_comments|forbidden|none|0
User|crm_contacts|user_crm_contact_link|forbidden|none|0
User|crm_contacts|user_crm_contacts_created|forbidden|none|0
User|crm_deals|user_crm_deals_owned|forbidden|none|0
User|dashboard_layouts|user_id|forbidden|none|0
User|file_metadata|uploader_id|forbidden|none|0
User|hub_bases|user_hub_bases|forbidden|none|0
User|hub_records|user_hub_records|forbidden|none|0
User|invites|user_invites|forbidden|none|0
User|invoices|user_invoices_created|forbidden|none|0
User|issue_message_links|created_by_id|forbidden|none|0
User|issues|issue_assignee|forbidden|none|0
User|issues|issue_reporter|forbidden|none|0
User|journal_entries|user_journal_entries|forbidden|none|0
User|kudos|user_kudos_received|forbidden|none|0
User|kudos|user_kudos_sent|forbidden|none|0
User|leave_requests|user_leave_requests|forbidden|none|0
User|leave_requests|user_leave_reviews|forbidden|none|0
User|mail_accounts|user_id|allowlisted|rootProof.mailAccountID|1
User|mail_objects|uploader_user_id|forbidden|none|0
User|mail_send_rate_buckets|user_id|forbidden|none|0
User|marketplace_listings|user_marketplace_listings|forbidden|none|0
User|marketplace_listings|user_marketplace_reviews|forbidden|none|0
User|news_poll_votes|user_id|forbidden|none|0
User|news_posts|user_news_posts|forbidden|none|0
User|notification_preferences|user_id|forbidden|none|0
User|notifications|user_id|forbidden|none|0
User|poll_votes|user_id|forbidden|none|0
User|polls|creator_id|forbidden|none|0
User|project_members|user_project_members|forbidden|none|0
User|project_templates|user_project_templates|forbidden|none|0
User|push_subscriptions|user_id|forbidden|none|0
User|reactions|user_reactions|forbidden|none|0
User|recent_contacts|user_id|forbidden|none|0
User|recordings|started_by|forbidden|none|0
User|saved_filters|owner_id|forbidden|none|0
User|scheduled_calls|created_by|forbidden|none|0
User|sessions|user_sessions|forbidden|none|0
User|team_members|user_team_members|forbidden|none|0
User|test_case_versions|user_test_case_versions|forbidden|none|0
User|test_cases|user_created_test_cases|forbidden|none|0
User|test_results|user_executed_results|forbidden|none|0
User|test_runs|user_created_runs|forbidden|none|0
User|time_logs|user_time_logs|forbidden|none|0
User|user_blocks|blocked_id|forbidden|none|0
User|user_blocks|blocker_id|forbidden|none|0
```

`CalendarInboundMessageReceipt` is intentionally absent: its
`organization_id`, `mailbox_id` and `mail_message_id` are scalar identity
fields, not direct Ent edges to Organization/User. It remains mandatory in the
closed Calendar graph proof and six-ID cleanup proof. Outbox origin columns are
also scalar provenance, not root dependents, and belong to O6.

## Six implementation shards

Shard boundaries are frozen and non-overlapping. Slot numbers refer to the
one-based order of the canonical block above.

| Shard | Exact slots | Root/range | File ownership |
|---|---:|---|---|
| `externalClosureOrganization` | `01-14` | all Organization slots | `controller_delete_published_dependents_organization.go` |
| `externalClosureUserA` | `15-27` | `activity_logs` through `calendar_mutation_receipts` | `controller_delete_published_dependents_user_a.go` |
| `externalClosureUserB` | `28-41` | `call_participants` through `dashboard_layouts` | `controller_delete_published_dependents_user_b.go` |
| `externalClosureUserC` | `42-55` | `file_metadata` through `mail_accounts` | `controller_delete_published_dependents_user_c.go` |
| `externalClosureUserD` | `56-69` | `mail_objects` through `reactions` | `controller_delete_published_dependents_user_d.go` |
| `externalClosureUserE` | `70-82` | `recent_contacts` through `user_blocks.blocker_id` | `controller_delete_published_dependents_user_e.go` |

Each shard owns only its typed locked reads and exact-set observations. The
manifest/policy/core/aggregate files own no shard reads. The coordinator owns
composition and the fail-closed decision; shards do not delete, commit, relax
another shard, or inspect catalogs.

## Drift gate

Implementation must contain a literal slot enum/array whose canonical emitted
lines match this block byte-for-byte. That literal-vs-document check is
necessary but insufficient: it cannot detect an added, removed or renamed
direct root FK in the schema authority.

Every disposable `schema-maintenance-generate`, `schema-maintenance-hash` and
`schema-maintenance-verify` must independently create generated desired-state
and full migration-replay inspect artifacts. A bounded stdlib Atlas-inspect SQL
parser derives only direct `Organization`/`User` three-field tuples, rejects
unknown or ambiguous DDL, requires desired/replay equality, then requires
equality to the three-field projection of this policy block. It must not read
PostgreSQL catalogs, use Ent reflection or source-grep.

`schema-maintenance-hash` prints frozen expected six-field count/bytes/full SHA
beside actual desired and replay three-field projection count/bytes/SHA.
`schema-maintenance-verify` recomputes independently and fails non-zero; it
must not consume another invocation's artifacts. Compose generate/hash uses
the schema-maintenance tool/internal DB, with no shared tmpfs/run-order state.

Static/unit acceptance requires all of:

1. exact policy grammar/order/slot count `82`;
2. every exact root/table/column/mode/source/count name present once;
3. canonical LF-terminated SHA-256
   `889c14127f47bdb73108bfef07573ff4b0bde7fd3b78b47122718a1f5b0a36b7`;
4. all six shard ranges complete, disjoint and ordered;
5. executable generated-desired-state and full-migration-replay three-field
   derivations agree with each other and with the frozen policy projection; and
6. no implementation fallback to dynamic `information_schema`, `pg_catalog`,
   reflection-derived predicates, broad organization/user scans, or
   best-effort omission.

Any Ent schema or migration change that adds, removes or changes a direct
Organization/User edge blocks cleanup implementation until this document,
the literal implementation manifest, expected count, hash, shards,
schema-maintenance derivation/binding and tests are reviewed and updated
together. An added, removed or renamed FK, or a count/hash/ordering mismatch,
is fail-closed and must never be interpreted as a clean closure.
