# Message bookmark command lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Root alone executes canonical Docker QA and Git; use the checkbox steps.

**Goal:** Saving a message bookmark has one scoped in-flight command, visible current failure and exact current-user list refresh.

**Architecture:** Add a local message-bookmark hook beside MessageItem. It owns only AddBookmark request lifecycle; MessageItem and MessageActions add thin wiring and bookmark-only pending UI. Reuse the existing bookmarksQueryKey, not a new global key or action framework.

**Tech Stack:** Existing React 19, TypeScript, Zustand, TanStack Query, Vitest/RTL and toastApiError; no new dependencies.

**Spec:** Existing user-scoped bookmark contract in docs/system_specification.md:715 and docs/TECHNICAL.md, plus the source finding and exact callsites in 2026-09-05-communications-completion-queue.md. Frozen AddBookmark behavior below supplements error/lifecycle UX without modifying backend rights.

Status: ROOT_CONTRACT_FROZEN; independent source-only review C/I/M 0/0/0,
READY_FOR_ORACLE. New oracle source-reviewed READY_FOR_RED C/I/M 0/0/0:
26 concrete cases / 372 physical lines, SHA256
9E2C85CD86A6F23840C184E26BB2E1DDCD6341EFCAF15DEB4FFE589A1B492E4B.
The foreign-author witness is merged into a stronger deferred success/cache
case, with an independent structuredClone snapshot before seeding. Both
revoke and unmount outcomes are present; prop-only transitions preserve the
same item. An unapplied same-reference snapshot payload was rejected.
Root applied only the author's corrected test payload after helper ACL
write failures. Production candidate now applied; independent source/oracle
review C/I/M0/0/0 READY_FOR_COMPOSED_GREEN. Footprint160 added/deleted
production lines including132-line hook; no existing oracle/assertion changes.
Baseline c0900e622964bd48541e8929b680801e82f550d0 on calls-feature-codex
is pushed. Reaction oracle is separately in progress. Root must serialize any
production writers touching MessageItem, MessageActions or messenger locales.

Latest publication checkpoint: reaction0d562ee855abfe1180646fe81a5a1ab8008cd475
is committed/pushed/remote-verified. This bookmark160-line candidate and26-case
oracle are UNIT_ACCEPTED in the disclosed composed809PASS gate; root docs
synced, exact bookmark commit/push next. No new runtime execution is inferred.

Publication completed:1877984e15462a4cc05ecb8cead2ca5810ac8ce2 committed,
immediately pushed to origin/calls-feature-codex and exact remote SHA
equality verified. Product/test tree remains the composed809PASS snapshot;
only the next pin plan was untracked after publication. No source acceptance
or branch mutation is inferred for that future pin atom.

## Global constraints

- This is ADD bookmark only. No removal, toggle state, new bookmark panel UX,
  forward/pin/reaction/edit/delete/SEND behavior or backend mutation changes.
- No optimistic bookmark list insertion, generic query invalidation, automatic
  retry, global command abstraction or cross-row/tab idempotency claim.
- Preserve POST payload {message_id, channel_id}, addBookmark(messageId,channelId),
  current prop AND live store can_bookmark policy and server authority.
- One mounted-row scope does not cancel/undo an already sent HTTP request.
- Do not read/edit pro-qa-ai/**, Mail/Calendar or their containers.
- Host reads/edits/Git only. Root uses exactly:
  docker compose -f docker-compose.local.yml run --rm frontend-unit
- Keep 2 GiB/no-swap, 6 CPU, two Vitest workers, same image/full tests/timeout.
  No build/pull/install; recheck adjacent Mail workload before serial QA.
- Explicit Terra/xhigh atomic writer, Terra/high independent reviewer;
  actual provider/model UNKNOWN absent telemetry, never silently fall back.
- Root owns exact-path commits and immediate push only calls-feature-codex
  to origin https://github.com/Valerijkk/hub.git with remote SHA equality.

## Confirmed source and root decision

MessageItem.handleBookmark checks canAct(can_bookmark), then bare await
addBookmark. It does not handle errors, block concurrent clicks or refresh
the user-scoped list. canAct checks prop/live capability and revoke, but not
actor identity or message-channel binding. MessageActions already supports
disabled on ActionButton; the bookmark action has no pending input.
BookmarksPanel exports bookmarksQueryKey(userId) and exact-invalidates the
captured current user's list after removal. Root chooses that existing key
for successful ADD refresh, without changing panel fetch/remove behavior.

## Frozen interface and effects

NEW frontend/src/components/messenger/use-message-bookmark-command.ts:

```ts
type BookmarkCommandOptions = {
  channelId: string;
  messageId: string;
  messageChannelId: string | undefined;
  canBookmark: boolean;
};
export function useMessageBookmarkCommand(options: BookmarkCommandOptions): {
  pending: boolean;
  submit: () => Promise<void>;
};
```

- Subscribe to current auth user ID, target channel live can_bookmark and
  revocation. Admit only nonempty actor/channel/message, matching message
  channel, prop canBookmark AND live can_bookmark===true, not revoked,
  mounted and exact committed scope.
- useLayoutEffect creates a fresh scope object for every committed actor,
  channel, message, messageChannelId, prop/live capability or revoke change.
  Invalidate on cleanup; A-B-A never revives the old object. Reset pending
  for the new scope. No mutable-ref writes during render.
- Synchronous requestRef claim before first await makes repeated/same-tick
  clicks no-ops. Disable only the bookmark action while its command is pending.
  Other row actions and other rows remain unaffected.
- Send exact captured IDs to addBookmark. On success, if exact request/scope
  are still current and live authority remains allowed, exact-invalidate
  bookmarksQueryKey(capturedUserId), exact:true. Do not touch another user,
  message timeline, channel list or alternate query prefix.
- Keep command pending through the awaited invalidation promise. TanStack
  default invalidation/refetch error policy stays unchanged; do not add
  throwOnError or treat a missing active observer as a failure.
- On current caught error, call toastApiError(error,t("messageBookmark.error"))
  once. RU fallback "Не удалось сохранить закладку"; EN "Unable to save bookmark".
  No success toast, hidden retry or arbitrary inline error classes.
- Stale success/error after committed user/channel/message change, capability
  loss/restore, revoke or unmount is consumed without toast/invalidation or
  unlocking a later command. Before side effects and final pending clear,
  check BOTH scope/request object identity and current auth/live capabilities.
- A later explicit click after current failure can retry. No transport abort,
  server duplicate/idempotency or rollback guarantee is claimed.

Command effect sketch:

```ts
const request = { scope };
requestRef.current = request;
setPending(true);
try {
  await addBookmark(scope.messageId, scope.channelId);
  if (isCurrent(request)) {
    await queryClient.invalidateQueries({
      queryKey: bookmarksQueryKey(scope.userId),
      exact: true,
    });
  }
} catch (error) {
  if (isCurrent(request)) toastApiError(error, t("messageBookmark.error"));
} finally {
  if (isCurrent(request)) {
    requestRef.current = null;
    setPending(false);
  }
}
```

Scope setup/cleanup invalidates obsolete refs separately without invoking
HTTP or cache effects. isCurrent uses object equality and live authority,
not just matching user/channel strings.

Adapters:
- MessageItem invokes the local hook with safeMessageId/message.channel_id
  and capabilities?.can_bookmark===true; replace only direct handleBookmark
  and remove only the addBookmark import. Pass submit as onBookmark and
  pending as bookmarkPending. Preserve in-flight reaction changes verbatim.
- MessageActions adds optional bookmarkPending:boolean default false to
  avoid changing existing capability-visibility fixture contracts; pass it
  only to the existing bookmark ActionButton.disabled.
- Locale additions are exactly messageBookmark.error in both messenger JSONs.
- Import the existing key from ./bookmarks-panel; do not modify that panel
  or extract unrelated shared modules in this atom.

## Exact ownership and bounds

Production:
- NEW frontend/src/components/messenger/use-message-bookmark-command.ts
- frontend/src/components/messenger/message-item.tsx: bookmark adapter only
- frontend/src/components/messenger/message-actions.tsx: bookmarkPending only
- frontend/src/i18n/locales/en/messenger.json: messageBookmark.error only
- frontend/src/i18n/locales/ru/messenger.json: messageBookmark.error only

Oracle:
- NEW frontend/tests/messenger/message-bookmark-lifecycle.test.tsx ONLY

Root docs:
this plan, .planning/STATE.md, completion queue, docs/TECHNICAL.md,
docs/TESTING.md, docs/test_plan.md, .planning/COVERAGE.md, docs/wiki/HOME.md.

Max 200 added/deleted production lines; one local command behavior.
Oracle max 550 lines and 26 concrete cases. Use real MessageItem/Actions,
Zustand stores, QueryClient and exported bookmarksQueryKey. Mock only
addBookmark/error formatter and unrelated rich children; do not mock owner,
force pending/internal state, weaken expectations or alter existing fixtures.
Other agents are not alone; exact ownership and root writer serialization
are mandatory. If the contract cannot fit, return BLOCKED for decomposition.

### Task 1: Scoped AddBookmark lifecycle

- [x] Independent plan review; C/I clear before oracle.
- [x] Write genuine test oracle against baseline without importing the
      future hook or using future props; production remains untouched.
- [x] Root full canonical RED; classify type/fixture/runtime issues separately
      from intended missing behavior. Preserve exact receipts and source hashes.
      May compose with the independent reaction oracle; disclose provenance.
- [x] Single production writer implements the hook and thin adapters AFTER
      reaction production writer has relinquished shared files, preserving it.
- [x] Independent oracle/source review C/I clear before GREEN.
- [x] Root full canonical GREEN, all tests and no unhandled errors. No
      exclusions/timeouts/assertion relaxation. Preserve failures honestly.
- [x] Root docs sync, exact commit/push/remote equality; continue unblocked work.

Oracle cases:
- Current successful add exact argument order and one exact captured-user
  bookmarks invalidation; seeded other-user and timeline keys untouched.
- Current RU/EN failure calls canonical formatter once, no success
  invalidation; button enabled afterward; explicit retry succeeds.
- Two same-tick clicks make one POST and only bookmark is disabled.
  Distinct row can still submit. No automatic retry.
- Unknown actor, empty/mismatched message channel, prop denied, absent or
  denied live capability and revoked channel produce no POST.
- Committed actor/channel/message A-B-A, both success and rejection:
  old result cannot notify/invalidate or unlock a new current request.
  Separate act commits required, not a batched A-B-A.
- Prop/live capability loss/restore, revoke and unmount ignore late success
  and failure. At least one test establishes a newer pending operation and
  proves the older completion cannot unlock it.
- Invalidation promise pending keeps current bookmark disabled; obsolete
  completion of that promise cannot unlock a newer scope.
- Preserve other-author bookmarking when current capability permits it.

Interaction sketch with helpers rendering the real production tree:

```ts
const save = screen.getByRole("button", { name: "Сохранить в закладки" });
act(() => { save.click(); save.click(); });
expect(api.addBookmark).toHaveBeenCalledTimes(1);
expect(api.addBookmark).toHaveBeenCalledWith("message-a", "channel-a");
expect(save).toBeDisabled();
await act(async () => pendingPost.resolve(bookmark));
expect(invalidate).toHaveBeenCalledWith({
  queryKey: bookmarksQueryKey("user-a"),
  exact: true,
});
```

## Composed baseline RED receipt — 2026-09-06

Same root canonical run and unchanged c0900e62 source as the reaction plan's
full receipt: TypeScript PASS;199 files197PASS/2FAIL;809 tests781PASS/28FAIL;
14 unhandled rejections; exit1; Vitest122.25s. Bookmark26cases13PASS/13FAIL;
reaction25cases10PASS/15FAIL. All old suites pass, not a clean-stderr claim.
Eight exact receipts are preserved without overwrite in
testframework/reports/frontend-unit-reaction-bookmark-red-20260906/.
Oracle SHA256 remains9E2C85CD86A6F23840C184E26BB2E1DDCD6341EFCAF15DEB4FFE589A1B492E4B.

Bookmark failures: missing exact invalidation in current success and awaited
invalidation cases; two current RU/EN notifications; repeated-click3POSTs
instead of2 across two rows; absent actor; message/channel mismatch; six
ABA pending-state preconditions. Those six stop before old promise settlement,
so this RED is not runtime evidence for their later stale-result assertions.
Six unhandled strings map to bookmark RU/EN, late prop/live capability,
revoke and unmount rejection; eight others belong to reaction. Source failure
triage independently confirms C/I/M0/0/2 READY_FOR_PRODUCTION; minor notes
preserve supplemental-window-spy and missing-resource-telemetry limitations.
Bookmark writer waits for reaction's shared-file relinquishment.

Outer worktree-master-frontend-unit-run-35071bc3e78d; engine
31e9575d2bdeb43f9b75c16a99bd62e6450133fd221190ad57cd0ba538ec6515;
same accepted image286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Engine events exit1/140s/no OOM event. Actual live resource/process telemetry
NOT_CAPTURED: inspect/top/stats happened after removal. Service declaration
remains2GiB/no-swap,6CPU/two workers; no live-observation claim for this RED.
Preflight idle Mail jobs,15636988KiB free RAM and low disk queue. No new
build/pull/install or source/config/assertion/timeout edits during the run.

## Acceptance and recovery limits

Composed canonical GREEN completed:TypeScript PASS,199files/809testsPASS,
all26bookmark and25reactioncases pass,0failed/pending/todo,118.81s,exit0.
Exact eight receipts, live telemetry and every composed source hash are
recorded under the reaction plan's Composed canonical GREEN section and in
testframework/reports/frontend-unit-reaction-bookmark-green-20260906/.
Whole-run error gate is exit0 plus0unhandled headings; JSON does not expose
an unhandled count. Existing warnings/mock stderr remain. Observed2GiB/
no-swap,6CPU,readonly,two workers;402.5/423.7MiB are samples,not peaks.
Engine completion/OOM events NOT_CAPTURED; no telemetry rerun. This is a
composed-tree receipt, not individual intermediate-commit runtime evidence.
Independent final receipt review C/I/M0/0/2 UNIT_ACCEPTED at genuine-unit
boundary; two minor evidence limitations are retained above. Publication pending.

Composed GREEN started2026-09-06T03:38UTC against this160-line candidate
plus independently reviewed reaction210-line candidate; no source/Compose/
oracle changes during execution. Hook SHA256
C587FD74389B8FF111C9A88D6D96E54A0416FFD54F09CB3F8D2B3F8BF4E6BB26.
Root staged the reaction-only nine-path source/fixture/oracle snapshot before
applying bookmark adapters; preserve this index for later separate commits.
Bookmarked shared-file changes remain unstaged until reaction publication.

Baseline TypeScript/197 files/758 tests PASS is not acceptance of this atom.
Browser/API/storage/end-to-end bookmark authorization is NOT_RUN; user list
refresh is proven here only at the genuine unit boundary. No module readiness
percentage or server idempotency claim derives from test count. Max one
transient runtime retry, three fixes/root cause, five review rounds. Record
blocked parts durably and continue unrelated approved work without new authority.
