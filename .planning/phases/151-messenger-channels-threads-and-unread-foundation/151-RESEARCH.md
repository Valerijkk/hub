# Phase 151 Research: Messenger channels, threads, and unread foundation

**Status:** `ROOT_FROZEN_MONGO_DURABLE_ARCHITECTURE`
**Confidence:** high for current seams and frozen behaviour; medium for exact new document/index names until implementation review.

## Decision

PostgreSQL/Ent remains the authority for tenant, project, channel, membership, eligibility, and role facts. MongoDB messages remain the message system of record and own the new durable unread and thread-follow state.

This replaces the earlier PostgreSQL/Ent cursor and `ThreadFollow` proposal. Phase 151 introduces no Ent unread/follow schema, Ent code generation, PostgreSQL migration, Redis-count conversion, distributed transaction, or outbox. This is materially smaller for the disk-sensitive environment and keeps message tuples and read cursors in one durable store.

The access boundary is strict: every Mongo message, state, follow, or count lookup/mutation is preceded by fresh PostgreSQL authorization. `AuthorizedUnreadScope` applies only to unread/read-state/follow/count/cutover work; timeline/thread/parent/message access instead requires fresh `AuthorizeChannelAccess` with `CanRead`, allowing same-tenant public pre-join history while still remaining PG-before-Mongo. Every authority, scope, grant, and event-recipient check includes an active account; a disabled account after its status commit is immediately denied with zero Mongo access. Mongo state never grants membership, project access, tenant access, or a role.

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

- unique `(tenant_id, user_id, channel_id, root_message_id)`;
- access cleanup `(tenant_id, user_id, channel_id)`;
- user list `(tenant_id, user_id, mode, channel_id)`.

`manual_unfollow` is retained, not deleted. Auto-follow may create a missing record or update only an existing `auto` record; it never overwrites either `manual_follow` or `manual_unfollow`. Manual follow changes the mode to `manual_follow` and baselines from the current authorized thread head. A newly auto-followed user is instead initialized at the triggering root/reply tuple, so that triggering activity is not unread; an existing eligible follower retains its cursor, so later activity may overlap with its other unread categories.

## Message document and indexes

Persist `also_send_to_channel`, immutable `resolved_mention_user_ids`, and canonical `post_create_effects_state` (`pending` or `complete`) on `MessageDocument` before the durable message create. `post_create_effects_completed_at` is recorded when it becomes `complete`; a legacy document with the state absent is treated as `complete`. Group mentions are resolved to concrete eligible user IDs at send time so later unread computation does not depend on changed membership.

Every head, unread selection, and cursor comparison uses ascending `(created_at, _id)` ordering. `_id` remains the canonical string message ID and breaks equal UTC timestamp ties.

| Query | Required Mongo index |
|---|---|
| Channel head/timeline/category unread | `(channel_id, created_at, _id)` |
| Thread replies/head | `(channel_id, thread_id, created_at, _id)` |
| Mention unread | `(channel_id, resolved_mention_user_ids, created_at, _id)` |
| Bounded pending post-create recovery | `(channel_id, post_create_effects_state, created_at, _id)` |

Timeline queries return root messages plus replies with `also_send_to_channel=true`. Thread queries return direct replies for the exact root. The same document is counted once inside each qualifying category, even if visible on two surfaces. Own messages are excluded from every category query and unread map. There is no separate effect ledger: every new canonical message, including a root and a direct-root reply, is created with `post_create_effects_state=pending`. Root effects include root-author/resolved-mention follow and recipient invalidation; a reply additionally runs idempotent `ReconcileReplyCount`. Both use `EnsureAutoThreadFollow` as applicable and `CompletePostCreateEffects` CAS to transition the canonical document to `complete`. The send retry path and bounded recovery use PostgreSQL-only, stable paginated, unique `ListPendingRecoveryTargets` proofs, then scan only those authorized channel-scoped pending documents through the recovery index. It has one global page/rate budget and performs neither per-member repeated scans nor a queue/outbox. The CAS precedes recipient-only NATS invalidation; a crash between them can lose invalidation, so reconnect/refetch remains required. No outbox or exactly-once claim is introduced.

One internal canonical message-write/effects pipeline is mandatory for `SendMessage`, `ForwardMessage`, `SendBotMessage`, the poll wrapper, and promotion backlink creation. It persists every root/reply as `pending`, resolves mentions, applies root/reply effects and CAS completion through this one path; direct callers cannot bypass it. A trusted internal producer receives an explicit, operation-scoped PostgreSQL authority proof for its exact tenant/channel/action, never a generic authorization bypass. Idempotent retry observes the canonical result and performs no duplicate effects.

## Authorization and linearization

Every membership grant uses one shared primitive. This includes `CreateChannel` owner/initial members, `JoinChannel`, `AddMembers`, `AddMembersInternal`, `AcceptInvite`, and `CreateOrGetDM`:

```text
fresh PostgreSQL tenant/public/private/project/DM eligibility and active-account status -> `AuthorizeMembershipGrant` returns immutable `AuthorizedGrantBaselineScope` proof for this exact user/target/kind
-> delete/reset stale Mongo read_states and access-bearing thread_follows (`auto`/`manual_follow`) for this user/target; retain `manual_unfollow` as a non-disclosive suppression preference
-> snapshot the current authorized Mongo channel/DM head tuple only with that proof
-> conditionally upsert channel-or-DM and mention baselines only with that proof
-> conditionally revalidate the same authority facts and active-account status inside the PostgreSQL membership transaction
-> commit PostgreSQL membership only if that revalidation still succeeds
-> publish success/invalidation
```

The logical grant point is the captured head. `AuthorizeMembershipGrant` produces the immutable `AuthorizedGrantBaselineScope` proof required for every pre-membership Mongo head/baseline operation; it conveys no enduring authorization, and the final PostgreSQL transaction must recheck tenant/channel/project/DM eligibility plus active-account status before writing membership. Messages after that tuple are eligible unread even if publication is delayed. If the final recheck fails, no membership is committed. Any pre-commit Mongo baseline is non-authoritative, compensation is best effort, and reconciliation removes orphans; PostgreSQL denial still prevents disclosure. Public browsing before membership creates no channel unread state or count. A newly created empty channel/DM uses the empty head baseline.

Revoke/leave/project-deactivation order is exact:

```text
revoke PostgreSQL access first (or commit disabled-account/project-status denial)
-> deny before every Mongo lookup/mutation/response
-> best-effort delete Mongo `read_states` plus access-bearing `auto`/`manual_follow` records and cursors; retain `manual_unfollow` as non-disclosive suppression
```

Cleanup failure cannot disclose anything because PostgreSQL denial is the boundary. Rejoin runs the shared grant primitive: it deletes/resets stale `read_states` plus access-bearing `auto`/`manual_follow` records and cursors before creating fresh current-head baselines. A `manual_unfollow` record is retained as a non-disclosive suppression preference: it grants no access, contributes no unread, is never returned without fresh PostgreSQL authorization, and continues suppressing auto-follow until manual follow. Cutover pages only PostgreSQL-authorized unread scopes and then follows only roots bound to those scopes; it never discovers scope from Mongo. An `AuthorizedUnreadScope` is PG-only: `TenantID`, `UserID`, and `Targets[]` of `{ChannelID, Kind}`; it applies only to unread/read-state/follow/count/cutover operations and requires active-account status. Empty/denied scope performs zero Mongo reads, writes, aggregates, or counts. Timeline/thread/parent/message resources do not use it: each performs fresh `AuthorizeChannelAccess(..., CanRead)` before Mongo so public pre-join reads remain legal and private/revoked/disabled reads fail closed. For each authorized workspace-channel membership, DM/group-DM target, mention target, and eligible persisted followed root, cutover baselines the current authorized head and never imports Redis counters.

No cross-store transaction, outbox, or exactly-once protocol is added. Refetch and idempotent reconciliation are the recovery contract. Reconciliation may remove orphan state or repair an incomplete baseline, but never grants membership or access.

## Current exact seams

### Backend

- Ent/PostgreSQL authority: `backend/ent/schema/channel.go`, `channelmember.go`, `projectmember.go`, `backend/internal/messenger/ent_repository.go`.
- Mongo message/state implementation: `backend/internal/messenger/dto.go`, `mongo_repository.go`, `repository.go`, `combined_repository.go`.
- Message/thread/idempotency: `service_messages.go`, `message_idempotency.go`, `handler_messages.go`, optional new compact `handler_threads.go`.
- Channel/join/revoke paths: `service_channels.go`, `service_channel_invites.go`, `service_dm.go`, `service.go`.
- Redis unread retirement: `unread.go`, `service.go`, `service_channels.go`, `service_bot.go`, `service_messages.go`, forward/message paths, `handler_reactions.go`, and `backend/cmd/server/wiring_messenger.go`. Redis may remain unrelated cache/transport, and presence remains unrelated, but no path may read, increment, decrement, publish, hydrate, or recover unread truth from Redis; Redis loss therefore has no unread-truth effect.
- Realtime invalidation: `events.go`, `backend/internal/ws/bridge.go`; inject a PostgreSQL-backed `UserChannelEventAuthorizer` plus `ChannelEventAudienceResolver`, retain canonical user room `user:<uuid>`, and update every bridge production constructor/wiring and test callsite to pass the non-nil dependencies explicitly. There is no legacy bridge constructor/default audience fallback.
- Cross-module search: `backend/internal/search/service.go` and its constructor/server wiring consume the same non-nil fail-closed `AuthorizedMessageSearchScope` factory before any Mongo candidate query or goroutine launch.
- Promotion: replace/disable the message-ID-only promotion route with `POST /messenger/channels/:channelId/messages/:messageId/promote-to-issue`; route/service obtains fresh channel-scoped `ChannelActionPromoteToIssue` proof before any Mongo lookup.
- Bot/typing external seams: `backend/internal/ai/messenger_bot.go`, every caller, and marketplace integration wiring; each carries an exact `AuthorizedBotChannelScope` proof including allowlisted `BotKind`, channel, integration, and action. `GetChannelEntity`, `GetChannelMemberIDsInternal`, and `PublishTyping` are covered alongside bot send/update/find/list.
- Existing routes include `GET /messenger/unread`, `POST /channels/:channelId/read`, and thread reply listing. Exact additive route/DTO names must be frozen by the plan.

### Frontend

- Recipient channel DTO/API/cache: `frontend/src/api/messenger.ts`, `frontend/src/hooks/use-channels.ts`; update their legacy local unread and realtime-test coupling so authoritative maps/refetch replace client-side arithmetic.
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
- unread response with separate `channels`, `dms`, `mentions`, and `threads` maps, plus server-authoritative boolean `any_unread` for the global sidebar;
- channel-scoped promotion only at `POST /messenger/channels/:channelId/messages/:messageId/promote-to-issue`, with the legacy message-ID-only route disabled/replaced;
- recipient-only `unread_changed` invalidation with identifiers only, never content/count.

Unread maps are `map[string]int64` and category-scoped exactly as `channels[channel_id]`, `dms[dm_conversation_id]`, `mentions[target_id]`, and `threads[root_message_id]`; absent keys mean zero and maps never imply access. The global sidebar receives/derives only boolean `any_unread = any(non-zero entry in any of the four authoritative maps)` and renders a single dot; it never sums map values, computes client-side increments, or treats overlapping category entries as multiple global badges. The generic receipt endpoint accepts only `channel`, `dm`, or `mention`; `followed_thread` is accepted only by the exact followed-thread receipt route. Invalid category/route/target combinations return `400` before any Mongo access. A channel receipt advances only its `channel` state, a DM receipt only its `dm` state, a mention receipt only its `mention` state for the authorized target, and a thread receipt only the matching eligible follow cursor. Every receipt validates access and its candidate message tuple server-side; own messages never increase any map. Root authors, reply authors, and resolved mention recipients create an auto-follow record at the triggering tuple when missing, while existing eligible follows are not reset and may therefore receive overlapping later unread.

A stale/revoked request must produce zero Mongo message/state/count lookup or mutation. `UserChannelEventAuthorizer` and `ChannelEventAudienceResolver` are injected into the WebSocket bridge with PostgreSQL implementations/wiring; every channel-scoped user event — `channel_created`, `channel_updated`, `new_message_notification`, and `unread_changed` — performs fresh active-account/project/channel authorization before audience selection, publish, and delivery. The sole post-revoke exception is a content-free `channel_removed` invalidation addressed to the just-removed user so their client can evict local state; it must not perform or imply renewed access. Idempotent `Created=false` send retries must not repeat mention resolution persistence, follow mutation, reply count, unread invalidation, or publication.

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
- global sidebar derives one boolean `any_unread` dot only from server-authoritative non-zero entries across the four maps, never a numeric sum; overlapping categories still show one dot;
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

## Implementation addendum: audited authorization seams

This addendum only hardens the frozen contract with findings from the completed backend seam audit; it introduces no new product behaviour.

### PG-before-Mongo rule

Create one centralized, fresh `AuthorizeChannelAccess` boundary in the messenger service/repository path. It resolves authorization from PostgreSQL/Ent for every request and must run before any Mongo read, aggregation, count, write, or response assembly. In addition to the Phase 151 unread/thread paths, harden these existing Mongo-touching operations:

- `ListMessages`, including same-tenant public history;
- legacy and exact thread/parent resource lookup before parent/reply Mongo access;
- `ListThreadReplies`;
- `EditMessage`;
- `ForwardMessage`, with separate fresh source and target channel proofs;
- `ToggleReaction`;
- `DeleteMessage` (Stage 1 `ChannelActionDeleteMessage` validates active account/member/project/channel eligibility in PostgreSQL; authority-denied public/non-member/read-only actors perform zero Mongo access. Stage 2 performs one exact authorized Mongo lookup to determine the author; the author or owner/admin/moderator may delete, while an ordinary other member is forbidden with no further Mongo operation or resource-derived response. No Ent ownership index/migration is introduced; validate any affected follow root is an eligible live root in the same channel);
- `PromoteToIssue` only through `POST /messenger/channels/:channelId/messages/:messageId/promote-to-issue`: the route resolves fresh `ChannelActionPromoteToIssue` for the route `channelId` before any Mongo lookup, requires the message to belong to that exact channel, and disables/replaces every message-ID-only route;
- `ListBookmarks`;
- pin and list-pin operations, because pinned-message reads and mutations can disclose channel history and therefore use the same PG-before-Mongo hardening scope rather than expanding Phase 151 product behaviour.

The relevant implementation seams are `backend/internal/messenger/service_messages.go`, `message_idempotency.go`, `service_channels.go`, `handler_messages.go`, optional compact `handler_threads.go`, `dto.go`, `mongo_repository.go`, `combined_repository.go`, `ent_repository.go`, `backend/internal/search/service.go`, `backend/internal/ai/messenger_bot.go`, marketplace integration wiring, and server/WS bridge wiring. Recipient capability projection is produced from fresh PG authority at channel-list/detail call sites and carried through recipient DTO/handler/cache seams; the exact thread service/resource resolves the authorized parent then direct replies, replacing the 50-root fallback. Global Search is separate: `AuthorizedMessageSearchScope` is PG-only, active-account/project-aware, and must never use Mongo discovery or `AuthorizedUnreadScope` to derive candidates. `AuthorizeChannelAccess` is an access check, not a Mongo-derived shortcut: stale/revoked/private/project-ineligible requests fail before any Mongo operation and return no resource-derived data.

### Tenant source and project-cap enforcement

The sole server-side tenant source is `Organization.ID`, resolved through Ent after authentication and access checks. The request body, route, client cache, message document, WebSocket payload, or caller-supplied tenant identifier must never select tenant scope. Missing or ambiguous organization resolution fails closed. Phase 151 does not introduce a multi-organization claim or tenant-selection model.

For a project-linked channel, `AuthorizeChannelAccess` and every membership grant/revoke path enforce an active `ProjectMember` cap. The same cap applies to WebSocket audience construction and project-member removal: remove/revoke access first, then emit only recipient-safe invalidation; never retain or publish a project-channel audience to a removed/ineligible user. Authority seams remain `backend/ent/schema/channel.go`, `channelmember.go`, `projectmember.go`, `backend/internal/messenger/ent_repository.go`, `service_channels.go`, `service_channel_invites.go`, `service_dm.go`, `service.go`, `events.go`, and `backend/internal/ws/bridge.go`.

### Evidence boundary

The phase-owned `python-integration` fixture for Mongo indexes, atomic upserts, membership grant/revoke, and project-removal no-leak checks is `NOT_BOUND` and therefore `NOT_RUN` until the implementation plan creates and binds its fixtures/infrastructure. No unit, mock, or static review result substitutes for that integration evidence.

### External callers and shared authorization contracts

`MessengerService` receives a non-nil `ChannelAccessResolver` explicitly in every production wiring and test constructor. There is no nil/default/legacy fallback: construction fails closed, and every operation below accepts only a fresh resolver proof.

**Bookmarks.** `ListBookmarks` first obtains the caller's PostgreSQL bookmark records and their candidate group-channel IDs, then applies fresh target `ChannelAccess` to each candidate before any Mongo enrichment. Denied targets are omitted; they cause zero target Mongo reads and return neither target identifiers nor resource-derived metadata.

**Search.** Messenger search and the external global `backend/internal/search/service.go` share the same PG-only `AuthorizedMessageSearchScope` and the same non-nil fail-closed scope factory in their constructors/server wiring; neither may derive candidates from Mongo or unread scope. The scope is resolved before any Mongo candidate query, aggregation, response assembly, or goroutine launch. It contains only freshly authorized channel targets and enforces same-tenant source, active account, public `CanRead` (including pre-join), private membership, and active `ProjectMember` cap. Ambiguous, cross-tenant, revoked, project-ineligible, nil, or construction-missing scope is empty and performs zero Mongo access and launches no search goroutine.

**GDPR erasure.** The retry-safe erasure workflow idempotently deletes the subject's messages, `read_states`, and `thread_follows`, and removes that user from every `resolved_mention_user_ids` array. It is deliberately cross-collection non-atomic: each bounded operation is independently repeatable, partial failure resumes safely, and authorization/search/list responses must not expose stale other-user roots, mentions, follow state, or counts while cleanup is incomplete.

**Bots and marketplace integration.** Introduce a narrow PG-only `AuthorizedBotChannelScope` with an allowlisted `BotKind`, exact `ChannelID`, exact integration identity, and exact action. Every operation and caller in `backend/internal/ai/messenger_bot.go`, plus marketplace integration wiring, carries this proof; `SendBotMessage`, bot update/find/list, `ListThreadReplies`, `GetChannelEntity`, `GetChannelMemberIDsInternal`, and `PublishTyping` all validate the supplied channel and integration against it before Mongo, Ent-derived member disclosure, NATS, or any other side effect. No generic trusted-producer or bot bypass exists, and a bot cannot expand tenant/channel/integration visibility through a caller-supplied ID.

**Typing.** `ChannelActionTyping` obtains fresh PostgreSQL proof of active sender account, channel membership, and active `ProjectMember` cap before NATS publish. Recipient audience resolution is independently fresh and PG-authorized at delivery time; denied/revoked/project-ineligible sender or recipient paths perform no NATS publish/delivery and disclose no audience.

**Retention compatibility.** Delivery is deferred to Phase 153, but every external deleted root/message path must preserve no-leak behaviour: deleted/ineligible roots produce no unread, follow, thread, or receipt disclosure. Tests cover deleted-root/message requests and stale other-user roots with zero unauthorized Mongo enrichment.

**WebSocket `unread_changed`.** The bridge uses a strict payload parser: only the agreed identifiers are accepted, with optional string `root_message_id`; unknown fields, content, and counts are rejected. Fresh delivery authorization runs before recipient selection and delivery, so an access change after enqueue cannot disclose the event. Every production bridge constructor/server wiring and every bridge test callsite must receive the explicit non-nil `UserChannelEventAuthorizer` and `ChannelEventAudienceResolver`; compatibility is an explicit callsite migration, not a nil/default legacy constructor fallback.

### Dashboard durable unread summary

The external Dashboard never reads the Redis `unread:<userID>` hash and does not reconstruct unread from Redis values. Redis may remain injected only for unrelated Dashboard cache/health behaviour; it is not unread truth, fallback, conversion input, or a soft-degradation source for the personal unread widget.

Inject a non-nil, fail-closed durable `UnreadSummaryProvider` into `DashboardService`. `GetPersonalDashboardData` first obtains the same fresh PostgreSQL-only active-account `AuthorizedUnreadScope` used by Messenger unread/read-state/follow/count work, then passes that immutable scope to the provider. A nil provider/factory, empty/denied scope, or authorization error yields no Mongo operation and fails closed for the Dashboard request; Dashboard must not synthesize a zero, a Redis-derived value, or an unauthorised partial count. The provider returns exactly:

```text
UniqueMessageCount int64
HasUnread          bool
```

It computes the server-side set union of accessible unread canonical message IDs selected under that scope across `channel`, `dm`, `mention`, and `followed_thread`. Each qualifying message ID contributes once even when it appears in multiple category maps, visible surfaces, or followed-thread activity. It may use bounded Mongo selectors/aggregation behind the fresh scope but never trusts caller category values, reads Redis, or sums category counts. `HasUnread` is true exactly when the deduplicated union is non-empty; it is not calculated from a client-side or Dashboard-side numeric sum.

The implementation seams are `backend/internal/messenger/unread.go` (durable provider, scope-bound selector/union), the PostgreSQL `AuthorizedUnreadScope` factory/resolver already required by Messenger, `backend/internal/dashboard/service.go` (replace only its personal unread Redis read), `backend/internal/dashboard/dto.go` (preserve a 64-bit count and expose the authoritative boolean if the widget contract needs it), `backend/cmd/server/wiring_modules.go` (explicit non-nil provider/factory injection), `backend/internal/dashboard/handler.go`, `frontend/src/api/dashboard.ts`, and `frontend/src/components/dashboard/widgets/unread-messages.tsx`. `backend/tests/dashboard/service_test.go`, Messenger unread unit tests, and wiring constructor tests must prove: no Redis unread call, denied/disabled/empty scope causes zero Mongo unread access, duplicate IDs across all four categories count once, and constructor missing dependencies fail closed. The Dashboard widget consumes this server response only and must not add the four maps or increment locally.

The OpenAPI source annotations and committed generated artifacts `backend/docs/swagger.yaml`, `backend/docs/swagger.json`, and `backend/docs/docs.go` must remain synchronized with the replacement channel-scoped promotion route and the four unread maps (`channels`, `dms`, `mentions`, `threads`), as well as any Dashboard summary DTO field exposed by this phase. The legacy message-ID-only promotion route and flat unread-map schema must not remain documented as supported.
