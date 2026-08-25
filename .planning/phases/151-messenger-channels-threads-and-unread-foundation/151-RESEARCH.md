# Phase 151 Research: Messenger channels, threads, and unread foundation

**Status:** `ROOT_FROZEN_MONGO_DURABLE_ARCHITECTURE`
**Confidence:** high for current seams and frozen behaviour; medium for exact new document/index names until implementation review.

## Decision

PostgreSQL/Ent remains the authority for tenant, project, channel, membership, eligibility, and role facts. MongoDB messages remain the message system of record and own the new durable unread and thread-follow state.

This replaces the earlier PostgreSQL/Ent cursor and `ThreadFollow` proposal. Phase 151 introduces no Ent unread/follow schema, Ent code generation, PostgreSQL migration, Redis-count conversion, distributed transaction, or outbox. This is materially smaller for the disk-sensitive environment and keeps message tuples and read cursors in one durable store.

The access boundary is strict: every Mongo message, state, follow, or count lookup/mutation is preceded by fresh PostgreSQL authorization. Mongo state never grants membership, project access, tenant access, or a role.

## Frozen product rules

- Public channel metadata/history is readable by an authenticated same-tenant user before join; posting requires membership.
- Private resources are member-only and non-disclosive.
- Project channels remain public/private plus `project_id`; join/read eligibility is capped by active project membership and membership is opt-in.
- `group` preserves the current group-DM meaning. New DM/group-DM lifecycle UX remains Phase 153.
- Read-only is orthogonal: ordinary members cannot post, owner/admin can.
- A thread is a live root plus direct replies in the same channel, rendered in the native right side-panel. Nested, cross-channel, or deleted-root replies fail closed.
- A reply is one Mongo document. `also_send_to_channel=true` makes that same document visible in its thread and channel timeline.
- Root authors, repliers, and persisted resolved mentions auto-follow. `manual_unfollow` suppresses future auto-follow until manual follow.
- Unread categories are exactly `channel`, `dm`, `mention`, and `followed_thread`; they are independent and may overlap.
- Join, rejoin, and cutover baseline at the current authorized target head. Redis is not read or converted.
- Realtime is invalidation only. Reconnect subscribes before authoritative refetch; clients perform no counter arithmetic.
- Archive/retention delivery remains Phase 153; Phase 151 only preserves no-leak behaviour.

## Durable Mongo collections

### `read_states`

Represents `channel`, `dm`, and `mention` positions:

| Field | Rule |
|---|---|
| `tenant_id`, `user_id` | Required authority scope. |
| `category` | `channel`, `dm`, or `mention`. |
| `target_kind` | `channel` or `dm`. |
| `target_id` | Channel/conversation identity. |
| `cursor_created_at`, `cursor_message_id` | Full monotonic tuple. |
| `created_at`, `updated_at` | Durable timestamps. |

Unique index: `(tenant_id, user_id, category, target_kind, target_id)`.

Cursor acknowledgement is one conditional atomic upsert: create when absent; update only when the candidate tuple is later than the stored tuple. Equal/stale retries are successful no-ops. No unread count is persisted.

### `thread_follows`

Represents the `followed_thread` category:

| Field | Rule |
|---|---|
| `tenant_id`, `user_id`, `channel_id`, `root_message_id` | Required identity/scope. |
| `mode` | `auto`, `manual_follow`, or `manual_unfollow`. |
| `cursor_created_at`, `cursor_message_id` | Consumed thread-activity tuple. |
| `created_at`, `updated_at` | Durable timestamps. |

Indexes:

- unique `(tenant_id, user_id, root_message_id)`;
- access cleanup `(tenant_id, user_id, channel_id)`;
- user list `(tenant_id, user_id, mode)`.

`manual_unfollow` is retained, not deleted. Auto-follow may create a missing record or update only an existing `auto` record; it never overwrites either `manual_follow` or `manual_unfollow`. Manual follow changes the mode to `manual_follow` and baselines from the current authorized thread head.

## Message document and indexes

Persist `also_send_to_channel` and immutable `resolved_mention_user_ids` before the durable message create. Group mentions are resolved to concrete eligible user IDs at send time so later unread computation does not depend on changed membership.

Every head, unread selection, and cursor comparison uses ascending `(created_at, _id)` ordering. `_id` remains the canonical string message ID and breaks equal UTC timestamp ties.

| Query | Required Mongo index |
|---|---|
| Channel head/timeline/category unread | `(channel_id, created_at, _id)` |
| Thread replies/head | `(channel_id, thread_id, created_at, _id)` |
| Mention unread | `(channel_id, resolved_mention_user_ids, created_at, _id)` |

Timeline queries return root messages plus replies with `also_send_to_channel=true`. Thread queries return direct replies for the exact root. The same document is counted once inside each qualifying category, even if visible on two surfaces.

## Authorization and linearization

Every membership grant uses one shared primitive. This includes `CreateChannel` owner/initial members, `JoinChannel`, `AddMembers`, `AddMembersInternal`, `AcceptInvite`, and `CreateOrGetDM`:

```text
fresh PostgreSQL tenant/public/private/project/DM eligibility
-> delete/reset stale Mongo read_states and access-bearing thread_follows (`auto`/`manual_follow`) for this user/target; retain `manual_unfollow` as a non-disclosive suppression preference
-> snapshot the current authorized Mongo channel/DM head tuple
-> conditionally upsert channel-or-DM and mention baselines
-> commit PostgreSQL membership
-> publish success/invalidation
```

The logical grant point is the captured head. Messages after that tuple are eligible unread even if publication is delayed. A failed PostgreSQL membership commit leaves only non-authoritative baseline state, which reconciliation removes; PostgreSQL denial still prevents disclosure. Public browsing before membership creates no channel unread state or count. A newly created empty channel/DM uses the empty head baseline.

Revoke/leave/project-deactivation order is exact:

```text
revoke PostgreSQL access first
-> deny before every Mongo lookup/mutation/response
-> best-effort delete Mongo `read_states` plus access-bearing `auto`/`manual_follow` records and cursors; retain `manual_unfollow` as non-disclosive suppression
```

Cleanup failure cannot disclose anything because PostgreSQL denial is the boundary. Rejoin runs the shared grant primitive: it deletes/resets stale `read_states` plus access-bearing `auto`/`manual_follow` records and cursors before creating fresh current-head baselines. A `manual_unfollow` record is retained as a non-disclosive suppression preference: it grants no access, contributes no unread, is never returned without fresh PostgreSQL authorization, and continues suppressing auto-follow until manual follow. Cutover enumerates each authorized workspace-channel membership, DM/group-DM conversation, mention target, and eligible followed root derived from persisted messages; it baselines every category target at its current authorized head and never imports Redis counters.

No cross-store transaction, outbox, or exactly-once protocol is added. Refetch and idempotent reconciliation are the recovery contract. Reconciliation may remove orphan state or repair an incomplete baseline, but never grants membership or access.

## Current exact seams

### Backend

- Ent/PostgreSQL authority: `backend/ent/schema/channel.go`, `channelmember.go`, `projectmember.go`, `backend/internal/messenger/ent_repository.go`.
- Mongo message/state implementation: `backend/internal/messenger/dto.go`, `mongo_repository.go`, `repository.go`, `combined_repository.go`.
- Message/thread/idempotency: `service_messages.go`, `message_idempotency.go`, `handler_messages.go`, optional new compact `handler_threads.go`.
- Channel/join/revoke paths: `service_channels.go`, `service_channel_invites.go`, `service_dm.go`, `service.go`.
- Redis retirement: `unread.go` and `backend/cmd/server/wiring_messenger.go`; Redis may remain unrelated cache/transport but not unread authority.
- Realtime invalidation: `events.go`, `backend/internal/ws/bridge.go`; canonical user room remains `user:<uuid>`.
- Existing routes include `GET /messenger/unread`, `POST /channels/:channelId/read`, and thread reply listing. Exact additive route/DTO names must be frozen by the plan.

### Frontend

- Recipient channel DTO/API/cache: `frontend/src/api/messenger.ts`, `frontend/src/hooks/use-channels.ts`.
- Durable unread projection: `frontend/src/hooks/use-unread.ts`, `frontend/src/stores/messenger-store.ts`.
- Browse/project/category badges: `frontend/src/components/messenger/channel-list.tsx`.
- Side-panel/follow/also-send: `thread-panel.tsx`, `message-composer.tsx`, `message-item.tsx`.
- Capability-driven posting and access loss: `chat-view.tsx`, `frontend/src/routes/_authenticated/messenger/$channelId.tsx`.
- Reconnect invalidation: `use-messenger-channels-realtime.ts`, `use-messages.ts`.

Current `useMessages` already excludes `thread_id` replies from the root cache; preserve it. Current thread parent fallback scans only the latest 50 root messages and must be replaced by an exact authorized parent/thread resource. Current client unread state is a flat `channelId -> count` map and must not survive as authority.

## API and DTO direction

The plan should freeze additive contracts for:

- recipient channel projection/capabilities: membership, role, `can_read`, `can_post`, `can_join`, project eligibility, read-only state;
- public/project discovery and explicit join/leave;
- send request `also_send_to_channel` valid only with a direct root reply;
- exact thread resource including authorized parent and follow mode;
- follow/unfollow and followed-thread read receipt;
- channel/DM/mention receipt using server-validated message ID or server-selected current head;
- unread response with separate `channels`, `dms`, `mentions`, and `threads` maps;
- recipient-only `unread_changed` invalidation with identifiers only, never content/count.

A stale/revoked request must produce zero Mongo message/state/count lookup or mutation. Idempotent `Created=false` send retries must not repeat mention resolution persistence, follow mutation, reply count, unread invalidation, or publication.

## TDD and Docker evidence

High-value backend RED cases:

- public prejoin read/no post/no unread; private/project/cross-tenant deny before Mongo access;
- read-only member denied, owner/admin allowed;
- root-only thread integrity and one-document dual placement;
- persisted effective mentions and auto/manual follow suppression;
- four independent/overlapping categories, sender exclusion, equal-time tuple order, forward-only upsert;
- every membership-grant entrypoint, join/send race, failed-grant orphan, revoke cleanup failure, rejoin stale-state reset, category-complete cutover, Redis loss, and manual-unfollow -> access loss with both successful and failed cleanup -> rejoin -> auto-trigger suppression until manual follow;
- strict `unread_changed` subject/payload and no duplicate retry side effects.

High-value frontend RED cases:

- Browse public/project sections and capability-driven composer/actions;
- right side-panel focus management and mobile modal behaviour;
- unchecked reply stays thread-only; checked `also_send_to_channel` updates both surfaces exactly once;
- distinct accessible channel/DM/mention/followed-thread badges;
- opening a destination acknowledges only its category;
- reconnect subscribes before refetch and never increments locally;
- access loss evicts channel/detail/root/thread/parent caches and closes the panel.

Canonical executable gates are Docker-only and sequential to protect disk:

1. one consolidated `backend-unit` RED after the full backend test packet;
2. one `backend-unit` GREEN after backend implementation/review;
3. one `frontend-unit` RED after frontend contract tests;
4. one `frontend-unit` GREEN after frontend implementation/review;
5. one phase-owned `python-integration` gate for Mongo index/upsert/join-revoke persistence, only after the plan binds fixtures and infrastructure.

Mongo/NATS/WS/browser evidence remains `NOT_RUN` unless its exact Docker service executes. Phase 149 integration evidence also remains `NOT_RUN`; do not infer runtime compatibility.

## Primary risks

- PG-before-Mongo ordering omitted on one handler/service path creates a disclosure vulnerability.
- Join/revoke cross-store windows are not transactions; failure compensation/reconciliation must be idempotent and non-authoritative.
- Timestamp-only cursors lose equal-time messages.
- Late mention resolution changes history.
- Client local arithmetic reintroduces duplicate/reconnect drift.
- Broad cache invalidation can leak stale private/thread data after access loss.
- Claiming Mongo index/aggregate correctness from mocks is invalid; integration evidence is required.

## Conclusion

Durable unread and follow state belongs beside Mongo message history. PostgreSQL decides access; deterministic tuples, forward-only atomic upserts, exact join/revoke ordering, authorization before Mongo disclosure, and explicit reconciliation bound the unavoidable cross-store residual without Ent codegen or Redis-led truth.
