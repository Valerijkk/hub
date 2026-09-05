# Messenger / Calls completion queue

Root continuation map, baseline f1691e34 (2026-09-05). This is a dependency
queue, not executable authorization for every row. Individual code atoms
require frozen exact files/oracles, independent review and Docker evidence.
Historical unit PASS or milestone completion is not product acceptance.

| Order / lane | Concrete remaining deliverable | Current gate |
|---|---|---|
| Messenger accepted atom | Visible reply-end read-through receipts in ThreadPanel | UNIT_ACCEPTED:36newcases, canonical TypeScript/666tests PASS; browser/API NOT_RUN |
| Shared Go gate | Complete canonical backend-unit after confirmed Ent compile OOM at 2GiB | Requested temporary Go-only 4GiB/no-swap exception UNAPPROVED; no further 2GiB retry |
| Messenger security | JoinChannel consumes fresh tenant/active-actor/project authority before membership effect | Source gap confirmed; separate atomic backend contract/test required |
| Messenger membership | Bind accepted membership lifecycle/baseline coordinator into production entrypoints | Pure coordinator unit evidence does not bind Join/Leave; no UI join acceptance before server gate |
| Messenger recovery | MPR-02 scheduler then remaining MPR production recovery chain | Accepted draft and prior reviews retained; executable RED waits on Go gate |
| Calls durable | CALL-DUR-08C2 executable acceptance, then 08C3 Ent transaction adapter | 08C2 STATIC_ACCEPTED only; no 08C3 dispatch before executable acceptance |
| Calls routing | CALL-DUR-08D persisted provider mode/router with guest denial and LiveKit parity | Split into exact single-behaviour atoms after 08C3 |
| Calls storage | CALL-DUR-08E real PostgreSQL locks, GDPR/concurrency/migration acceptance | SUT bridge, TC/DSN/runner/PG ownership need a separate frozen QA contract |
| Jitsi deployment | JITSI-D1 selected digest image discovery, self-hosted config/branding/overlay | NOT_RUN; separate bounded image/disk/infra authority before pull/start/attachment |
| Jitsi durable events | Separate provider PostgreSQL spool; Prosody durable append/retry, mTLS ingress/replay, crash matrix | D1 and B0 precede B1/B2/C1/C2/D; no memory-only spool or at-least-once claim before durable enqueue |
| Calls product | Authenticated public join, native Hub embed, task-owned containers and real two-user audio/video | Internal provider/credential/embed units are not this runtime gate |
| Messenger next UI | Explicit Browse Join, then Leave; owner metadata/read-only/member management | Join auth fix first; manage/delete capabilities must be server-defined, not role-inferred |
| Messenger navigation | Search successful-empty/error separation, identity-safe result navigation to message/thread | Exact error/locator/thread contract must be frozen before code; current search only opens channel |
| Messenger links | Thread issue-link create/unlink | Actor policy, target proof and capabilities UNBOUND; read list acceptance does not authorize mutation |
| Final acceptance | Reconnect/revoke/unread isolation, file/message/thread user journeys; Jitsi media/rollback/load; docs | Real Docker SUT/browser/API/storage evidence required; no full module PASS yet |
| Legacy cutover | Drain then remove LiveKit by exact manifest | Only after replacement runtime/rollback exit gates, not before |

## Confirmed Join gap and boundary

handler.go applies authMw; handler_channels.go calls JoinChannel directly.
service_channels.go:356 checks channel type and IsMember, then AddMembers,
without the existing ChannelAccessResolver proof. Discover and channel_access.go
do apply tenant, active actor and active project membership before can_join.
The returned capability cannot secure a later direct HTTP mutation. Root
verified the source body; live exploit and persistence behavior remain NOT_RUN.
The next backend contract must preserve authorized idempotency while denying
stale/foreign/inactive authority before effects, and define the legacy nil-
resolver branch explicitly. It must not silently reuse an unscoped fallback.

## Operating constraints

### Autonomous continuation instruction (2026-09-06 local time)

Founder requested uninterrupted autonomous work on the agreed scope: record
new questions/blockers here and continue independent unblocked work. This
does not approve the outstanding Go-only4GiB exception, external infrastructure,
image pulls, broader Git publication or host QA. Do not repeat the failed2GiB
Go compilation. No architecture re-questionnaire is required for bounded
implementation decisions inside the already agreed product contract.

Read-through source and36-case oracle now UNIT_ACCEPTED after review and
canonical Docker666/666GREEN; publication pending. Next unblocked source
contract: SearchDialog successful-empty/error separation and scoped request
lifecycle, preserving channel-only result navigation. Frozen plan:
2026-09-06-search-dialog-lifecycle.md; oracle payload in progress. The adjacent
Mail builds finished before the accepted gate; root did not stop/change them.

- Root owns serial canonical Docker QA; keep FE at2GiB/no-swap and no host QA.
- Recheck Mail build/test workload before a run; do not stop its containers.
- No rebuild/install/pull loops, generated-tree churn or large report copying.
- Explicit Terra task routing/ownership; no duplicate implementers or reviewers.
- Every accepted atomic commit is followed by a push of calls-feature-codex
  only to origin https://github.com/Valerijkk/hub.git, then remote SHA check.
- Whole Python framework remains NOT_ACCEPTED (last529PASS/71FAIL snapshot).
  FE success must not be promoted into a backend/framework/product PASS.
