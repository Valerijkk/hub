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
| Messenger search lifecycle | Current-only explicit search, successful-empty/error separation and scoped retry/navigation | UNIT_ACCEPTED:35 new cases in composed TypeScript/728-test Docker PASS; channel-only navigation |
| Messenger navigation | Exact result navigation to message/thread | Route/API locator contract remains UNBOUND; current search only opens channel |
| Messenger delete | Single-flight, handled failure and stale-completion isolation | UNIT_ACCEPTED:20 new cases in same composed gate; existing own/delete-any policy preserved |
| Messenger thread actions | Current capabilities passed from ChatView through ThreadPanel to root and replies | UNIT_ACCEPTED:7 new boundary cases in same composed gate; no new rights |
| Messenger edit | Matched channel admission, pending/re-entry and current-scope error feedback | Frozen2026-09-06-message-edit-lifecycle.md;24-case oracle applied and reviewed C/I/M0/0/0; source/RED NOT_RUN |
| Messenger failed-send recovery | Per-nonce Retry single-flight, retained real Retry and working explicit Cancel | Frozen2026-09-06-composer-retry-lifecycle.md;2new mock plus4real-Toaster cases applied/reviewed C/I/M0/0/0; source/RED NOT_RUN |
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

Follow-up root source review: ChannelActionJoin does not yet exist in the
closed vocabulary. Public nonmember facts currently allow only ReadTimeline
and project CanJoin=true; an existing member normally has CanJoin=false.
Therefore checking CanJoin alone would break authorized idempotent Join.
Future decomposition must separate the closed fact/action contract from
the JoinChannel production binding. Root security direction is fail-closed
for missing Join authority, not a silent legacy nil fallback; other actions'
legacy behavior is outside that atom. Preserve a freshly authorized public
existing-member no-write/no-event result without setting the UI CanJoin flag
for existing members. Exact error/result contract and affected old nil-
constructor test manifest still need freezing before implementation.
No code or Go tests have been changed/run for this discovery.

## Operating constraints

Resource checkpoint2026-09-06T00:56UTC: adjacent Mail builder again has
active qa-runtime/backend-unit builds. Our combined edit/composer RED is
deferred; no competing run/build/pull was started. Free RAM14818716KiB,
disk1%/queue0 do not override the no-concurrent-heavy-build boundary.
Continue read-only reaction-flow discovery and durable bookkeeping while
waiting. Production edit/composer sources remain at published d10f1dda.

### Autonomous continuation instruction (2026-09-06 local time)

Founder requested uninterrupted autonomous work on the agreed scope: record
new questions/blockers here and continue independent unblocked work. This
does not approve the outstanding Go-only4GiB exception, external infrastructure,
image pulls, broader Git publication or host QA. Do not repeat the failed2GiB
Go compilation. No architecture re-questionnaire is required for bounded
implementation decisions inside the already agreed product contract.

Read-through source and36-case oracle now UNIT_ACCEPTED after review and
canonical Docker666/666GREEN; committed/pushed d65da37d with remote equality
verified. Next unblocked source
contract: SearchDialog successful-empty/error separation and scoped request
lifecycle, preserving channel-only result navigation. Frozen plan:
2026-09-06-search-dialog-lifecycle.md; oracle payload in progress. The adjacent
Mail builds finished before the accepted gate; root did not stop/change them.

## Next-flow discoveries (read-only,2026-09-06)

### Accepted checkpoint — 2026-09-06

Resource budget 75dcdb27 and composer recovery 5057c54d are committed,
pushed and exact remote SHA verified. Edit lifecycle is UNIT_ACCEPTED and
publishing next. The composed canonical frontend-unit gate passed TypeScript
and 197 files / 758 tests in 114.95s under 2 GiB/no-swap with two observed
Vitest workers. All 24 edit, 11 mock-composer and four real-Toaster cases
pass. Earlier RED and first post-source failures remain archived, not
relabeled. Exact receipts and source hashes are in the three 2026-09-06
edit/composer/worker-budget plans. Intermediate commits were not tested
individually. Next unblocked atom: shared per-row reaction request ownership.
Jitsi selected-image discovery and Go compilation remain separately blocked;
this frontend acceptance does not cover either.

Reaction request failure is source-confirmed in two production entrypoints:
ReactionPicker.handleSelect awaits toggleReaction without catch and its
emoji button discards the promise; MessageItem.handleReaction does the same
for an existing reaction pill. Existing pre/post picker capability/revoke
checks and WS delta validation do not consume rejected HTTP promises.
Runtime rejection/notification acceptance is NOT_RUN. Freeze a separate
request-owner/error-feedback contract after the current edit/composer atoms;
do not change those active files or infer new backend reaction rights.
Duplicate non-idempotent toggles also lack a current single-flight guard;
this is a separate scope decision, not yet an implemented reaction fix.

Exact search-result locator remains UNBOUND: current Messenger route has no
validated message/thread search params; no public channel-scoped get-message/
context/around API exists. Existing getThread validates a live root and returns
only a bounded initial reply page. Opening that root is not proof of visibility
of an old target reply; unbounded timeline pagination is not a locator. Freeze
a separate route/API contract before claiming exact-message search navigation.

Delete-action discovery corrected a previous broad suspicion: hover-toolbar
visibility matches current own-member versus delete-any moderator semantics.
can_delete_message is delete-any, not the author's own-message permission.
The concrete UI gap is no pending guard or handled rejection in handleDelete
(message-item.tsx), not a license to broaden/narrow server rights. A future
frontend-only single-flight/error UX atom can preserve the existing gate.
Runtime/backend authorization acceptance was not performed by this discovery.

Search/delete/thread wire now UNIT_ACCEPTED in one composed canonical
frontend-unit gate:195files/728tests PASS,TypeScriptPASS,64.12s,
2GiB/no-swap,6CPU,readonly,Compose exit0. Detailed RED/GREEN, hashes and
uncaptured post-run lifecycle events are in the three2026-09-06 atom plans.
Thread wire1db4ad3f, deleteb78ddafc and searchd10f1dda published and remote
SHA verified. Tree clean before the next independent edit-save plan.
No browser/backend/full-product acceptance is inferred.

Jitsi D1 constraint rechecked against the frozen150 plan: selected-image
inspection is blocked while adjacent Mail services/builder exist; no pull,
build, overlay or activation attempted. Latest read-only C: free sample
268353069056bytes is also below its250GiB floor. No cleanup of user/Mail
data is authorized by that observation. Go-only4GiB exception remains
UNAPPROVED; no further2GiB Ent-compile OOM retry.

Resource refresh2026-09-06T00:21UTC: no Mail service/test containers listed;
only its existing BuildKit daemon remains. Root did not stop anything.
Builder history contains completed jobs; free RAM12692528KiB, disk0%/queue0,
but C: free262325129216bytes is below the Jitsi250GiB floor. No image pull,
render, build or container start was attempted for Jitsi.
An initial jitsi/* inventory was insufficient because pins use ghcr.io.
Follow-up local image inspect of all four exact150-RESEARCH.md ghcr.io
digests returned No such image. This read-only command downloads nothing.
One permission-review deadline timed out before execution; its single
read-only retry succeeded. D1 remains NOT_RUN, not a completed image audit.

SEND discovery: existing immutable nonce/snapshot and scope guards are
retained, but repeated Retry can run concurrent HTTP. Root also verified
Sonner2.0.7 upstream: current mock-object cancellation does not prove that
the real non-dismissible toast's Cancel works, and action dismissal loses
Retry after another failure. Exact root-frozen recovery contract and primary
source are in2026-09-06-composer-retry-lifecycle.md. Do not equate callback
mock acceptance with real notification-control acceptance.

Thread action wiring gap confirmed by scoped source review: ChatView passes
the server channel capabilities to timeline MessageItem but not ThreadPanel;
the latter renders both root and replies with undefined capabilities. The
existing fail-closed MessageItem guards then suppress available actions.
Separate prop-only contract: 2026-09-06-thread-capability-wiring.md. No new
permission/query/role policy; unknown and revoked remain denied. Runtime
NOT_RUN; discovery is not acceptance of action/server behavior.

- Root owns serial canonical Docker QA; keep FE at2GiB/no-swap and no host QA.
- Recheck Mail build/test workload before a run; do not stop its containers.
- No rebuild/install/pull loops, generated-tree churn or large report copying.
- Explicit Terra task routing/ownership; no duplicate implementers or reviewers.
- Every accepted atomic commit is followed by a push of calls-feature-codex
  only to origin https://github.com/Valerijkk/hub.git, then remote SHA check.
- Whole Python framework remains NOT_ACCEPTED (last529PASS/71FAIL snapshot).
  FE success must not be promoted into a backend/framework/product PASS.
