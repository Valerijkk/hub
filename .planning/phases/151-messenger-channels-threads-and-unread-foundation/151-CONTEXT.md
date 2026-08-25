# Phase 151 — Frozen Context

**Status:** `FROZEN_FOR_PLANNING`
**Authority:** Founder decisions supplied for Phase 151 planning. This is scope/context, not implementation evidence.

## Problem boundary

Messenger needs a stable, tenant-safe base for channels, root-only threads, follow semantics, and unread recovery before richer messaging work. The contract must avoid accidental disclosure at all visibility, membership, project-membership, deletion, access-revocation, and reconnect boundaries.

## In-scope decisions

1. Public channel metadata/history are readable by an authenticated same-tenant user before joining; membership gates posting.
2. Private channel resources are member-only. This includes metadata, history, exact thread/parent resources, follow state, unread categories, and action results.
3. Project channels are an explicit `project_id` opt-in. Their declared public/private setting still applies, while project-linked access is limited to active `ProjectMember` users.
4. `group` preserves the existing group-DM meaning; Phase 151 does not add a new direct-message experience.
5. Read-only is an orthogonal channel mode: ordinary members cannot post, while channel owners and admins may post.
6. A thread opens in the native right side-panel while preserving channel timeline context, and consists of a live root plus direct replies in that root's channel only. No nested replies, cross-channel replies, or replies to deleted roots.
7. Default channel timeline omits replies. A reply with explicit `also_send_to_channel=true` remains one document rendered in both the thread and permitted channel timeline.
8. Root authors, permitted repliers, and resolved permitted mentions auto-follow; manual unfollow disables future auto-follow until manual follow.
9. Access loss prevents disclosure and follow/unread mutation, even through stale subscriptions or recovery paths.
10. Each unread category has an independently durable, forward-only `(created_at, message_id)` position scoped to tenant, user, category, and category target. MongoDB stores `read_states` and `thread_follows` beside authoritative messages; Redis is never the source of truth. PostgreSQL/Ent remains the tenant/channel/membership/project/role authority, and every Mongo message, state, or count lookup or mutation requires fresh PostgreSQL authorization. Join and cutover use the current accessible target head as baseline; Phase 151 adds no Ent schema/codegen/migration.
11. Four unread categories are independent and overlap-capable: `channel`, `dm`, `mention`, `followed_thread`. An also-sent reply may contribute to its channel/DM category and, independently, mention or followed-thread categories.
12. On reconnect the client subscribes and then refetches authoritative state; it never calculates unread deltas locally.
13. Recovery is persist/refetch safe. There is no transactional outbox; residual missed/duplicated delivery is reconciled by refetch rather than hidden by a guarantee.

## Explicit non-goals

- Archive and retention delivery; these belong to Phase 153. Phase 151 imposes only no-leak handling for such states.
- Broad Phase 153 Messenger features.
- Jitsi, Mail, Calendar, and any work under `pro-qa-ai/**`.
- New group/DM user experience, reply nesting, a general event-delivery guarantee, or Redis-led unread truth.
- Runtime validation claims, test execution, Docker work, schema migrations, implementation, or release status.

## Planning constraints

Implementation decomposition must retain separate authorization checks for tenant, channel visibility, channel membership, project link, active project membership, and read-only posting. Authorization must fail closed before any Mongo message/read/follow/count lookup or response; best-effort cleanup after revoke cannot be the disclosure boundary. It must emit exact parent/thread resources, not a client-reconstructed relationship. It must specify a capability-based recipient DTO before frontend action design, and ensure inaccessible/deleted resources are non-disclosive.

Later plans must trace all work to the following IDs:

| Family | IDs | Planning focus |
|---|---|---|
| Channels | `MSG151-CHN-01`, `MSG151-CHN-02` | Visibility, membership/project/read-only authority; DTO capabilities. |
| Threads | `MSG151-THR-01`, `MSG151-THR-02` | Root-only integrity, native side-panel context, parent/thread response, and timeline/also_send_to_channel placement. |
| Follow | `MSG151-FOL-01`, `MSG151-FOL-02` | Auto-follow, manual suppression, access-loss no-leak. |
| Unread | `MSG151-UNR-01`, `MSG151-UNR-02` | Independent durable category-target positions/baselines and overlap-capable categories. |
| Storage | `MSG151-STO-01` | PostgreSQL authorization boundary, Mongo durable state, and conditional forward-only upsert. |
| Realtime/recovery | `MSG151-RT-01`, `MSG151-RT-02` | Subscribe-then-refetch and explicit no-outbox residual. |

## Acceptance posture

Acceptance requires executable evidence against the measurable criteria in `151-SPEC.md`, including negative authorization/no-leak paths, cursor tie ordering, rejoin/cutover baseline, category overlap, and reconnect recovery. Planning status remains `FROZEN_FOR_PLANNING` until a later plan defines exact components and its approved Docker-only verification matrix.

## Dependency and known residual

Phase 149 integration is `NOT_RUN`; no compatibility/runtime assertion is authorized from this context. A later implementation plan must identify the precise Phase 149 interface/event/resource assumptions and test them. The absence of a transactional outbox is a documented residual: persist/refetch recovers authoritative state, but does not establish transactional publication or exactly-once delivery.
