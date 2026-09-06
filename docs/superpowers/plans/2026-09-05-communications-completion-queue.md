# Messenger / Calls completion queue

Root continuation map, baseline f1691e34 (2026-09-05). This is a dependency
queue, not executable authorization for every row. Individual code atoms
require frozen exact files/oracles, independent review and Docker evidence.
Historical unit PASS or milestone completion is not product acceptance.

| Order / lane | Concrete remaining deliverable | Current gate |
|---|---|---|
| Messenger accepted atom | Visible reply-end read-through receipts in ThreadPanel | UNIT_ACCEPTED:36newcases, canonical TypeScript/666tests PASS; browser/API NOT_RUN |
| Shared Go gate | Complete canonical backend-unit after confirmed Ent compile OOM at 2GiB | Approved4GiB diagnostics: scratch384MiB and1GiB both infra ENOSPC; only first2build packages pass in attempt2; vet/unit/race NOT_RUN, exact defaults restored; no third attempt; scratch diagnostics and canonical observer profile need a separate contract |
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
| Messenger edit | Matched channel admission, pending/re-entry and current-scope error feedback | UNIT_ACCEPTED:24newcases in composed197file/758testPASS; c0900e62 pushed and remote-verified |
| Messenger failed-send recovery | Per-nonce Retry single-flight, retained real Retry and working explicit Cancel | UNIT_ACCEPTED:2new mock plus4real-Toaster cases in same composed gate;5057c54d pushed and remote-verified |
| Messenger reactions | Per-row shared picker/pill request ownership, current errors and stale-session isolation | UNIT_ACCEPTED:25newcases in composed199files/809testsPASS;0d562ee8 pushed and remote-verified |
| Messenger bookmarks | Scoped add single-flight/error and exact current-user list refresh | UNIT_ACCEPTED:26newcases in same composedGREEN;1877984e pushed and remote-verified |
| Messenger row pin | Scoped toggle/error and exact current-user/channel pins refresh | UNIT_ACCEPTED: 26 new cases; TypeScript + 200 files / 835 tests PASS; independent C0/I0/M0; ac918830 pushed and remote-verified |
| Messenger forward | One request per open session; current error/retry; source and captured-target authority fences | UNIT_ACCEPTED: TypeScript + 201 files / 871 tests PASS,36new; independent source/final receipt C0/I0; 74c5c492 pushed and remote-verified |
| Messenger panel unpin | Panel-owned single-flight, handled failure and exact pins refresh through invalidation | UNIT_ACCEPTED:34new; composed TypeScript+203files/915testsPASS; independent C/I/M0/0/0;25c58baa pushed/remote-verified |
| Messenger bookmark removal | Prevent same-bookmark duplicate DELETE and preserve current error/retry | UNIT_ACCEPTED:10new; per-ID lifecycle89lines; same composed GREEN and independent C/I/M0/0/0;49b0f133 pushed/remote-verified |
| Messenger links | Thread issue-link create/unlink | Actor policy, target proof and capabilities UNBOUND; read list acceptance does not authorize mutation |
| Final acceptance | Reconnect/revoke/unread isolation, file/message/thread user journeys; Jitsi media/rollback/load; docs | Real Docker SUT/browser/API/storage evidence required; no full module PASS yet |
| Legacy cutover | Drain then remove LiveKit by exact manifest | Only after replacement runtime/rollback exit gates, not before |

## Exact message navigation discovery — 2026-09-06

At published 74c5c492, SearchDialog.handleNavigate sends only channelId;
the channel route has no validateSearch/message locator and ChatView takes
only channelId. Its initial/bottom auto-scroll cannot identify a search hit.
Search returns full Message DTO including id/channel_id/thread_id/created_at,
but these values alone do not guarantee the target is loaded in the virtual
list. The current messages GET accepts before(timestamp)+limit, not an exact
message anchor; no GET /messages/:messageId route is registered. GetThread
accepts a root ID, not a general reply locator, and returns a paginated thread.

Root direction for future decomposition: use an actor-authorized bounded
anchor resource and explicit route locator, not unbounded history loading or
injecting stale search DTOs into the live timeline. Define deleted/foreign/
revoked concealment, root-versus-reply resolution, stable ordering including
timestamp ties, forward/backward page continuity, current-actor query keys,
and one-shot virtual scroll/highlight. First freeze the backend DTO/error/
pagination contract and its test acceptance; then bind route/ChatView/thread.
This is a discovery direction, not a new API implementation or security
integration approval. Go/SUT gates remain blocked; no source was changed.

## Current Messenger SUT QA boundary — 2026-09-06

Read-only discovery at c1990858: `testframework/tests/api/` currently contains
authz, files, health, identity and platform suites, but no Messenger suite;
`testframework/tests/e2e/` does not exist. The canonical Python entrypoint and
Compose service names exist, which is not evidence of a working two-user
channel/message/thread journey. The e2e services declare isolated QA storage
dependencies, not a bound task-owned Hub backend/frontend pair. Messenger
SUT acceptance is therefore NOT_READY / NOT_RUN. A separate bounded contract
must bind the SUT lifecycle, actor fixtures, verified cases and resource budget
before writing/running these journeys. No host or legacy-runner workaround,
unowned container startup, or unrelated Python-suite repair is authorized by
this discovery. Frontend unit acceptance remains useful but narrower.

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

Exact legacy-oracle manifest (read-only Terra/medium, c0900e62): one path,
two test cases and three direct JoinChannel statements, all with nil
ChannelAccessResolver. In backend/tests/messenger/user_channel_events_test.go:
TestUserChannelLifecycleJoinPublishesCreatedOnlyForNewMember at line176
calls JoinChannel at185 and186 (new member then idempotent retry); the
"join channel membership mutation failure" subtest at631 calls it at739.
These are isolated event/coordinator oracles, not production authorization
proof. Their constructor ExportNewMessengerServiceWithEventPublisher
(export_hooks.go:29-34; service.go:99-108) does not set authority. A future
binding atom must supply explicit authorized fixtures without reclassifying
these tests. New authorization/adversarial cases belong in a separate oracle.

Current Join errors: missing channel404, private/DM/unknown type403, repository
get/IsMember/AddMembers failure500 (service_channels.go:357-395). Nil-authority
error and authorized-existing-member action facts still need root freezing.
phase151ChannelAccessFixture/Allow/Deny are reusable scaffolding, but Allow
currently models owner/member, not public nonmember. Public prejoin examples
currently authorize ReadTimeline, not a nonexistent Join action. Runtime
remains NOT_RUN; no additional Go OOM attempt or memory-limit increase.

## Operating constraints

Current2026-09-06 Go resource outcome supersedes the earlier UNAPPROVED /
NOT_RUN notes below: temporary4GiB/no-swap was explicitly approved. Root ran
two bounded single-variable experiments sequentially, only after Mail QA
released the resource window. Both infra-aborted on RAM scratch ENOSPC;
attempt2 completed two build packages but no full stage or tests. Both unique
partial directories and terminal evidence remain preserved. Temporary
Compose/runtime edits are restored byte-exact; no third attempt, repeated
2GiB OOM, SSD scratch fallback or observer weakening is authorized. See
[diagnostic evidence and remaining contract](2026-09-06-backend-unit-4g-exception.md).
The outstanding gate is not permission for the already-approved4GiB ceiling:
it is a bounded scratch diagnostic and separately accepted canonical profile.

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

### Historical row-pin checkpoint — 2026-09-06

Checkpoint at that stage: row pin UNIT_ACCEPTED, TypeScript + 200 files / 835 tests
PASS in 117.97s; all 26 new pin cases, unchanged oracle/source/Compose hashes.
Independent final review C/I/M 0/0/0. Exact receipts are preserved under
testframework/reports/frontend-unit-pin-green-20260906/. Actual 2 GiB/no-swap,
6 CPU, readonly, two Vitest workers observed; samples are not peaks and
engine completion/OOM events NOT_CAPTURED. Published ac918830e2aa0d7082297d9913be0fd750013fb5,
immediately pushed to the authorized origin branch; exact remote SHA verified.

ForwardDialog confirmed gap: state-only duplicate guard, bare try/finally
without current-error feedback, no committed actor/message/open-session
token; late close/finally can affect a later dialog. Existing fresh source
can_forward and target can_post/revoke checks must remain. A request-specific
target authority subscription can retain any loss/restore invalidation
without resetting on unrelated channel-list updates. Root must freeze this
separate command contract and tests before any implementation.

Earlier chronological checkpoints below are historical, not current state:

Panel-unpin follow-up: root read PinnedMessages.handleUnpin and confirmed
bare await togglePin with no catch/synchronous request guard. Preserve both
prop/live can_list_pins AND can_pin, actor/channel and pin.channel_id binding,
exact current-user/channel key and existing query loading/error/empty UI.
Separate local panel hook proposal is appropriate; row hook only can_pin
must not silently acquire panel policy. One pending request per open panel
is the bounded candidate because every unpin refreshes the same pins query.
Future root freeze must separately cover list and pin authority loss, panel
channel/open ABA, unmount and pending through actual query invalidation.
Analyst's 18-case proposal is not yet an accepted test inventory; no panel
source or oracle edits and no executable panel-command claim.

Root has now frozen the separate 34-case panel plan (2026-09-06-panel-unpin-lifecycle.md)
and independent review returned C/I/M 0/0/0. Row hook remains unchanged;
only two production paths are authorized after its own RED. The current
forward publication must finish before panel oracle application. Existing
panel test counts are 5+4+3=12; identity/query-state API mocks export only
getChannelPins/togglePin, so the future hook uses direct capability fields.
Current frontend exposes togglePin only: lost-response retry ambiguity is
a separate backend product risk, not solved by the local single-flight guard.

Bookmark removal discovery (read-only Terra/medium): BookmarksPanel.handleRemove
already catches current-user errors and exact-invalidates bookmarksQueryKey
after success, but has no request lock or disabled control. Two same-tick
clicks can dispatch duplicate removeBookmark DELETE calls; the second may
produce a misleading error after first success. Source anchors panel:50-73,
140 and api/messenger.ts:510-511. Existing identity-boundary:174-211 covers
one A-to-B removal case; bookmarks-error:26-49 is GET error, not mutation.
Minimal next candidate is per-bookmark-ID dedupe plus explicit retry, no
query-wide blocking. A-to-B-to-A ID equality may refresh a valid current A
cache; whether additional generation fencing is required needs root contract
decision, not an assumed bug. No source/test changes or runtime proof yet.
Initial agent default read hit helper ACL failure; scoped escalated read
succeeded without ACL mutation. No broader access or model fallback used.

Both atoms published and remote-verified: reaction0d562ee8,bookmark1877984e.
Pin RED completed:TypeScriptPASS,200files199pass1fail,835tests821pass14fail,
6unhandled; only new pin oracle failed. Independent triage C/I/M0/0/0
opens the frozen production writer. Source remains unchanged until root
applies its reviewed bounded payload. In parallel, a read-only analyst
checks ForwardDialog command lifecycle for the next independent queue item;
that is discovery only, not authorization to implement or change permissions.
Pin baseline now1877984e15462a4cc05ecb8cead2ca5810ac8ce2; author authorized to
apply only the corrected26-case oracle. Root independent review/hash/RED
will precede pin source. Preserve the accepted809-case receipt separately.

Reaction0d562ee855abfe1180646fe81a5a1ab8008cd475 published and exact remote
verified. Bookmark160-line source/26-case oracle now publishing separately,
without changing the accepted composed source/test tree. Next unblocked
runtime atom is row-pin; its reviewed plan and corrected draft oracle are
not yet part of any executable gate. Do not apply that oracle before this
publication finishes and its own exact prerequisite HEAD is bound.

Latest composed GREEN:TypeScript PASS,199files/809testsPASS,118.81s,
all25reaction+26bookmark cases; independent final review C/I/M0/0/2
UNIT_ACCEPTED at genuine frontend-unit boundary. No changed assertions or
timeouts. Exact receipts/source hashes in reaction/bookmark plans; sampled
resource and missing engine-event limitations retained. Reaction-only index
is frozen for first commit; bookmark shared-file differences stay unstaged
until that commit is pushed and remote-verified. Pin oracle still payload
only, not part of this gate. Browser/API/media/full-module gates unchanged.

Latest reaction/bookmark RED: TypeScript PASS,199files197pass2fail,
809tests781pass28fail,14 unhandled rejections. Both new oracles account for
all failures; older suites pass. Exact failure mapping and retained receipts
are in their two frozen plans; independent triage precedes production writes.
Do not read historical NOT_RUN discoveries below as the latest gate.
No production/config changes since c0900e62. Next: serialized reaction then
bookmark source ownership, independent reviews, one composed canonical GREEN.

Publication follow-up: edit c0900e622964bd48541e8929b680801e82f550d0
was committed, pushed and remote-verified; tree was clean. Reaction plan
2026-09-06-message-reaction-lifecycle.md is root-frozen for independent
review; its new source and oracle remain NOT_WRITTEN/NOT_RUN.

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

### Bookmark discovery — source-only

Next independent bookmark discovery (read-only Terra/medium): MessageItem
handleBookmark uses canAct("can_bookmark") followed by bare await addBookmark;
it lacks handled rejection, per-row pending and actor/message-channel fences.
The existing BookmarksPanel exports bookmarksQueryKey(userId) and exact-
invalidates that user key for remove, but AddBookmark does not refresh it.
Proposed later atom: current-scope AddBookmark single-flight/error handling
plus exact current-user success invalidation, preserving payload and server
permissions. Root must freeze its contract before writes; do not fold it
into reactions or refactor accepted edit/delete. Runtime and backend
idempotency/write-authorization remain unverified. Likely exact ownership:
new message-bookmark-lifecycle.test.tsx, focused local command module or
MessageItem-only handler, MessageActions bookmark-pending prop and two
messageBookmark.error locale entries. Source evidence: MessageItem:519,
canAct:312, api/messenger.ts:502, bookmarks-panel.tsx:18/50 and
docs/system_specification.md:715 at c0900e62. No bookmark code/test written.

### Reaction discovery — source-only

Execution checkpoint: reaction and bookmark plans independently reviewed
C/I/M 0/0/0; reaction oracle was written (claimed 25 concrete cases, source
review in progress), bookmark oracle has its own writer. Production stays
at c0900e62. Root serializes shared component/locales writers after accepted
RED, and may batch the two independent oracles in one canonical matrix.
An extra read-only dispatch hit the active thread limit; no repeated
capacity request or duplicate writer was started.

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

### Pin command discovery — source-only

Root froze a separate row-only plan:
2026-09-06-message-pin-lifecycle.md. It preserves can_pin-only row admission
(can_list_pins is not newly required), adds message-channel binding and
current-scoped command/error/pending fences, retains exact server refetch,
and excludes PinnedMessages unpin. Independent plan review is in progress.
Oracle/production NOT_WRITTEN; reaction/bookmark shared-file writers precede
its implementation. Localized fallback arguments are frozen without changing
canonical formatter precedence. No automatic retry or optimistic pin list.

Root read the two current command paths while test writers worked:
MessageItem.handlePin at488 and PinnedMessages.handleUnpin at131 both await
togglePin and exact-invalidate pinsQueryKey(channelId,userId) after current
identity/capability/revoke checks. These protections already exist and must
be retained. Neither command has caught mutation rejection or a synchronous
pending lock; GET query error/loading UI in PinnedMessages is separate and
already implemented. Existing phase151-message-row-pin-cache,
phase151-pinned-identity-boundary, phase151-pinned-capabilities and
phase151-pinned-query-state oracles cover identity/cache/permission/query
behavior, not accepted mutation-recovery UX. Any later pin atom must retain
server-refetch semantics and list-vs-row capability differences; do not
derive PinnedMessage from the toggle boolean response or mix both entrypoints
without a separately frozen owner contract. No pin source/test changed;
runtime failures and repeated-toggle behavior remain NOT_RUN.

Jitsi gate provenance rechecked in 150-PLAN.md: explicit infrastructure
authorization precedes image pull/render/build/start; D1 also stops below
250 GiB free or over 8 GiB task-owned growth. These are frozen preconditions,
not a disk-cleanup or permission grant. No gate was lowered and no image
was downloaded during this source-only check.

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
