# Message reaction request lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. Follow the checkbox steps; root alone executes canonical Docker QA and Git.

**Goal:** Both reaction entrypoints share one per-row request owner, visibly handle current failures and ignore obsolete completions.

**Architecture:** Add one local hook beside MessageItem, not a global command framework. MessageItem owns it and passes its non-rejecting command through MessageActions to ReactionPicker. Existing reaction pills use that same command. Picker owns only its open-session UI fence.

**Tech Stack:** Existing React 19, TypeScript, Zustand, TanStack Query, Vitest 3.2.4/RTL and Sonner formatter; no dependencies added.

**Spec:** This document's frozen behavior contract implements the existing can_react and reaction HTTP behavior described in docs/TECHNICAL.md, docs/system_specification.md and the source-confirmed reaction finding in 2026-09-05-communications-completion-queue.md. It does not change backend rights.

Status: ROOT_CONTRACT_FROZEN; independent source-only review C/I/M 0/0/0,
READY_FOR_ORACLE. New oracle is source-reviewed READY_FOR_RED: 25 concrete
cases / 503 physical lines, SHA256 EE316695F754793AC8871915D3011DA1D324B6B4A564D0992568834B1C1C74C5.
Rereview C/I/M 0/0/1; the minor note is that local window rejection spies
are supplemental and cannot replace the canonical whole-run error gate.
Root rejected the initial mounted-pill review premise: MessageItem:635
returns null on revoke. Oracle now verifies actual row removal, not changed
UI behavior. Other corrections add exactly-once toast, independent deep
snapshots and distinct parameterized titles. Production candidate is now
applied after two root correction rounds; independent source review C/I/M
0/0/1 READY_FOR_COMPOSED_GREEN. Minor requested exact line count: root
measured210 production added/deleted lines including131-line hook, within250.
Baseline: calls-feature-codex c0900e622964bd48541e8929b680801e82f550d0,
clean and pushed with exact remote equality. Prior composed canonical gate:
TypeScript and 197 files / 758 tests PASS under 2 GiB/no-swap, two workers.

## Global constraints

- One behavior only: reaction toggle request lifecycle at one mounted message row.
- No modifications to accepted edit/delete, SEND, pin/bookmark/forward, WS,
  query-key/cache architecture, backend, API payload, permissions or emoji catalog.
- No optimistic reaction count/user-ID changes, cache invalidation or auto-retry.
  Existing server/WS projection remains authoritative.
- An already sent HTTP request is not aborted, rolled back or globally serialized.
  Different mounted rows/tabs are NOT covered by the per-row single-flight lock.
- Host reads/edits/Git only. Root uses exactly:
  docker compose -f docker-compose.local.yml run --rm frontend-unit
- Preserve service image, two-worker budget, 2 GiB/no-swap, 6 CPU, full tests,
  assertions and timeouts. Recheck adjacent Mail workload first; no build/pull.
- Do not read/edit pro-qa-ai/**, Mail/Calendar or other branches' resources.
- Exact-path commits and immediate push only calls-feature-codex to
  origin https://github.com/Valerijkk/hub.git, then verify remote SHA.
- Root freezes contracts and accepts reviews. Explicit Terra/xhigh atomic
  writer, Terra/high independent review; no child agents or silent fallback.

## Source-confirmed gap

ReactionPicker.handleSelect directly awaits toggleReaction and discards the
returned handler promise at onClick. MessageItem.handleReaction does the
same for existing emoji pills. Neither consumes rejection or shares a
synchronous pending lock. Existing capability/revoke checks and WS projection
idempotency are not a non-idempotent HTTP toggle guard. Runtime defect
witness remains NOT_RUN until the oracle is applied.

## Frozen behavior and interface

Create frontend/src/components/messenger/use-message-reaction-command.ts:

```ts
type ReactionCommandOptions = {
  channelId: string;
  messageId: string;
  messageChannelId: string | undefined;
  canReact: boolean;
};
// The returned submit promise ALWAYS resolves boolean for transport outcomes.
// true: HTTP success while this exact request and committed scope remain current.
// false: denied, duplicate, stale, unmounted or handled transport failure.
export function useMessageReactionCommand(options: ReactionCommandOptions): {
  pending: boolean;
  submit: (emoji: string) => Promise<boolean>;
};
```

1. Hook subscribes to current auth user ID, target channel's live can_react and
   revocation. Admission requires nonempty actor/channel/message, exact
   messageChannelId===channelId, prop canReact AND live can_react===true,
   no revoke and a mounted/current committed scope.
2. useLayoutEffect setup creates a fresh object identity for every committed
   actor/channel/message/messageChannelId/prop-canReact/live-canReact/revoke
   transition; cleanup invalidates it. A-B-A cannot reuse an earlier token.
   Reset local pending for the new scope. No ref writes during render.
3. Before the first await, synchronously claim a request ref. If occupied,
   return false without transport. Capture immutable scope IDs; send the
   original emoji unchanged to toggleReaction(scope.channelId, scope.messageId, emoji).
4. Before effects/settlement, require exact scope AND request object identity,
   mounted lifecycle, live actor and live permission/revoke checks. Old success
   returns false; old error is consumed and returns false without notification.
5. Current error calls toastApiError(error, t("messageReaction.error")) once,
   then returns false. EN fallback "Unable to update reaction"; RU fallback
   "Не удалось изменить реакцию". No success toast or new error-class rules.
6. Finally clear pending only if this exact request still owns the current
   scope/ref. Old completion must not unlock a newer request.
7. A later explicit click can retry after current failure; no automatic retry.
   Reactions to other authors' messages remain permitted by can_react.

Adapter contracts:
- MessageItem instantiates the hook using safeMessageId and message.channel_id.
  Replace only direct handleReaction with submit; remove only toggleReaction
  import. Existing pills add pending to their existing disabled condition.
  Wire required onToggleReaction and reactionPending to MessageActions.
- MessageActions adds required onToggleReaction:(emoji:string)=>Promise<boolean>
  and reactionPending:boolean props. Pass to ReactionPicker as
  onToggleReaction and pending. Do not alter any other action.
- ReactionPicker keeps channelId/messageId/canReact/open/onOpenChange and adds
  required pending:boolean plus onToggleReaction. Remove direct toggleReaction.
  Retain existing canReactNow gate, capability/revoke UI and emoji markup.
- Emoji buttons disabled while pending; trigger remains openable so a user
  can close/reopen during transport. Do not infer a successful toggle from
  a resolved false. Failure leaves the still-current picker session open.
- useLayoutEffect creates a local open-scope object keyed by committed
  open/disabled/channelId/messageId; cleanup nulls only its own token.
  Pending changes and callback function identity do NOT reset this token.
  Capture token before awaiting submit; close only on true AND unchanged,
  non-null open-scope AND canReactNow. Closing/reopening, route change,
  capability loss or unmount prevents the old result closing a later popup.
  Auth A-B-A is denied by the shared owner even when popup stays mounted.
- Minimal direct fixture adaptation only: add required callback/pending props
  in a11y-names-p2.test.tsx and phase151-message-action-capabilities.test.tsx.
  Preserve every existing expectation and test count.

Implementation shape (not permission to broaden the module):

```ts
const request = { scope };
requestRef.current = request;
setPending(true);
try {
  await toggleReaction(scope.channelId, scope.messageId, emoji);
  return isCurrent(request);
} catch (error) {
  if (isCurrent(request)) toastApiError(error, t("messageReaction.error"));
  return false;
} finally {
  if (isCurrent(request)) {
    requestRef.current = null;
    setPending(false);
  }
}
// Scope setup/cleanup separately clears an obsolete ref without settling
// any HTTP operation. isCurrent checks both refs AND live auth/capability.
```

## Exact files and one task

Production writer owns only:
- NEW frontend/src/components/messenger/use-message-reaction-command.ts
- frontend/src/components/messenger/message-item.tsx: reaction adapter only
- frontend/src/components/messenger/message-actions.tsx: reaction prop wiring only
- frontend/src/components/messenger/reaction-picker.tsx: owner adapter/open fence
- frontend/src/i18n/locales/en/messenger.json: messageReaction.error only
- frontend/src/i18n/locales/ru/messenger.json: messageReaction.error only

Oracle writer owns:
- NEW frontend/tests/messenger/message-reaction-lifecycle.test.tsx
- The two named existing fixtures ONLY for final required-prop adaptation,
  AFTER baseline RED so the initial typecheck does not fail for future props.

Root owns this plan, .planning/STATE.md, completion queue and final docs sync:
docs/TECHNICAL.md, docs/TESTING.md, docs/test_plan.md, .planning/COVERAGE.md,
docs/wiki/HOME.md. No other writes.

Production change budget: new focused hook plus thin adapters, at most
250 added/deleted production lines total. If it cannot fit one behavior,
return BLOCKED for decomposition; no broad MessageItem extraction.
Oracle budget: at most 600 lines and 30 concrete cases. Helpers must keep
real production MessageItem/Actions/Picker/hook, Zustand stores and QueryClient.
Mock only HTTP boundary, canonical error formatter and unrelated rich children.
Do not mock the command owner or force internal pending/state.

### Task 1: Shared reaction owner and picker settlement

- [x] Independent contract review C/I/M; resolve C/I before oracle.
- [x] Write genuine UI oracle first against the published baseline, without
      importing the future hook or requiring its future props. Assertions
      observe actual accessible controls and exact API calls.
- [x] Root run full canonical frontend-unit; inspect TypeScript, all results
      and whole-run errors. Missing fixture/type errors are not intended RED.
      Preserve precise receipts; classify unhandled old HTTP rejection as
      source failure only with the corresponding focused case evidence.
- [x] Implement the frozen hook/adapters and required fixture props only.
- [x] Independent source + oracle review; unchanged old behavior must survive.
- [x] Root full canonical GREEN, unchanged assertions/timeouts/tests; whole
      run unhandled errors and missing cases preclude acceptance.
- [x] Root docs sync, exact-path commit/push/remote equality; continue queue.

Required executable cases (parameterization allowed, each observed case
must be concrete; trim redundant cases to stay at or below 30):
- Exact successful picker API payload and current-session close; no optimistic
  count/cache mutation. Exact existing-pill request for another author's message.
- Current rejection via each entrypoint produces one canonical error; RU/EN
  fallbacks; current picker stays open and later explicit retry succeeds.
- Same-tick/repeated click and cross-entrypoint clicks share one pending lock;
  only reaction controls disabled, other message actions remain enabled.
- Unknown actor, prop denial, absent/live-denied capability, revoked channel
  and mismatched message.channel_id produce no HTTP.
- Committed actor/channel/message A-B-A for success AND rejection: old scope
  cannot notify, close a new picker, alter current data or unlock its request.
  Use separate act commits for A->B and B->A, not one batched rerender.
- Capability loss/restore, revoke and unmount consume late rejection quietly.
- Close/reopen popup while HTTP pending; old success cannot close the new
  popup, but completing a request started in its own open session can close it.
- Distinct rows are not globally blocked; no automatic retry.

Oracle interaction sketch (helpers must render the real production tree):

```ts
fireEvent.click(screen.getByRole("button", { name: "Реакция" }));
fireEvent.click(screen.getByRole("button", { name: "👍" }));
expect(api.toggleReaction).toHaveBeenCalledTimes(1);
expect(api.toggleReaction).toHaveBeenCalledWith("channel-a", "message-a", "👍");
expect(screen.getByRole("button", { name: "👍" })).toBeDisabled();
await act(async () => pendingRequest.reject(new Error("network")));
expect(errorToast).toHaveBeenCalledTimes(1);
expect(screen.getByRole("button", { name: "👍" })).toBeEnabled();
```

## Composed baseline RED receipt — 2026-09-06

Root canonical command: docker compose -f docker-compose.local.yml run --rm frontend-unit.
Source/Compose remain c0900e622964bd48541e8929b680801e82f550d0; both frozen
oracle hashes above and in the bookmark plan are unchanged before/after.
TypeScript PASS;199 files197PASS/2FAIL;809 tests781PASS/28FAIL;14 whole-run
unhandled rejections; exit1; Vitest122.25s starting02:45:33UTC.
Reaction oracle25cases:10PASS/15FAIL; bookmark oracle26cases:13PASS/13FAIL.
All older suites pass. Existing React act warnings are not a clean-stderr claim.

Reaction failures map to five missing current-error notifications, one
same-tick/cross-entrypoint duplicate (3 requests instead of1), absent actor,
message/channel mismatch, six ABA pending-state preconditions and one old
success closing a newly reopened popup. The ABA cases stop before settling
the old promise; this RED does NOT establish that their later assertions ran.
Eight unhandled strings match reaction current/retry and late capability,
revoke/unmount HTTP rejections; six others belong to bookmark. Local window
spies are supplemental, not substitutes for the whole-run error gate.
Independent final triage C/I/M0/0/2: READY_FOR_PRODUCTION. The two minor
notes preserve supplemental-spy and NOT_CAPTURED-resource limitations above.
Root dispatched one Terra/xhigh reaction writer; bookmark production waits
for shared-file relinquishment. Production GREEN remains NOT_RUN.
Root rejected the writer's first UNAPPLIED payload: it widened the frozen
hook options instead of owning subscriptions, removed Picker.canReactNow,
omitted nonempty channel admission and added an unrequested emoji guard.
No source file changed. Fix round1 restores the exact interface/live guards
without touching frozen oracles; this is not a runtime failure or accepted code.
Root also rejected the second UNAPPLIED hook cleanup: clearing requestRef
alone left its scope usable by a retained callback after unmount. Round2
uses nullable scopeRef, clears both refs on cleanup and denies absent scope.
Only that corrected hook plus round1 adapters/fixtures were applied. Final
candidate210 production added/deleted lines (131-line hook), four fixture
prop lines; no old assertions or either new oracle changed. Hook SHA256
995B9FBB8564CBF7E5E23A9859DEFC3844E6439542620D8F6389DC4858B04D26;
Picker74BF01EBB4988F21B63E14A6F0948B8FDEA3AADF3C8D6AC12F89510E58BB74F5.
Independent source review C/I/M0/0/1 READY_FOR_COMPOSED_GREEN. Root measured
the exact210-line production budget, resolving the review's missing numstat
observation. Bookmark payload writer now owns the next command boundary;
no concurrent source writer remains. Composed GREEN remains NOT_RUN.

Eight preserved receipts, no overwrite:
testframework/reports/frontend-unit-reaction-bookmark-red-20260906/
{result.json,typescript.result.json,typescript.stdout.log,typescript.stderr.log,
vitest.result.json,vitest.report.json,vitest.stdout.log,vitest.stderr.log}.
Outer worktree-master-frontend-unit-run-35071bc3e78d; engine
31e9575d2bdeb43f9b75c16a99bd62e6450133fd221190ad57cd0ba538ec6515;
image sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218;
Compose config-hash3aab41e0124f93a444c46b9f0bc8d2554314525e0fdd28fd6e12c453db46bc19.
Engine started1788662715.778222068, died1788662855.995787677, exit1,
execDuration140s, removed1788662856.004405031; no OOM event in exact window.
Live inspect/top/stats were too late: actual memory/swap/CPU/worker telemetry
NOT_CAPTURED for this run. The unchanged accepted service declares2GiB,
no-swap,6CPU/two workers; do not reclassify declaration as live observation.
Preflight: no active Mail build job,15636988KiB free RAM, disk samples0/0,
0/0,2/0 percent/queue. No source/config/assertion/timeout/image changes,
build/pull/install or competing root QA. Do not rerun RED for telemetry alone.

## Verification limits / recovery

## Composed canonical GREEN — 2026-09-06

Root command: docker compose -f docker-compose.local.yml run --rm frontend-unit.
TypeScript PASS;199 files/809 tests PASS;0 failed/pending/todo; all25 reaction
and26 bookmark cases pass, including ABA assertions after their former RED
pending preconditions. Vitest118.81s, start03:38:52UTC; canonical step/target
receipts and unified Compose process exit0. Source/oracle hashes unchanged
before/after; no test exclusion, assertion/timeout relaxation or source edit.

Eight copied receipts, new directory without overwrite:
testframework/reports/frontend-unit-reaction-bookmark-green-20260906/
{result.json,typescript.result.json,typescript.stdout.log,typescript.stderr.log,
vitest.result.json,vitest.report.json,vitest.stdout.log,vitest.stderr.log}.
Vitest JSON schema has NO numUnhandledErrors property. Successful whole-run
exit and0 unhandled-error/rejection/exception headings in both logs establish
the error gate; do not invent a JSON count. Existing React act/Dialog warnings
and expected mocked-failure logging remain, so stderr is not described as clean.

Live outer worktree-master-frontend-unit-run-17763f3f03f8, engine
9fc0d913f531ef769a15fc90be3bea57842f15a09d3d5ca62ec45960de9ae529,
started2026-09-06T03:38:34.889487339Z. Image remains
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Inspected Memory=MemorySwap2147483648,NanoCpus6000000000,readonlyRootfs=true,
source/test binds readonly,MAX2/MIN1 threads/forks. docker top observed main
Vitest plus two named workers and esbuild. Samples402.5/423.7MiB and
CPU227.00/234.58percent,IO0B/0B; these are samples, NOT measured peaks.
Host queue0, freeRAM14374764KiB at preflight; Mail builds completed and its
six standing services had low CPU usage. None were stopped/changed.
No image build/pull/install or parallel root QA.

Engine completion/OOM events NOT_CAPTURED: first format was rejected and
two bounded read-only event retrievals returned no matching events. Do not
infer a measured engine OOM flag from missing events. Canonical receipts
and unified process independently establish exit0; no rerun for telemetry.

Exact composed working-tree hashes:
- Compose E8DA0EEC4F99FE3B437C885B3B19019F27D24958D477152176F7210CB4A56733
- MessageItem0AC755B5BB17B98A7B44A3A2DE67178D4B42EA2E1CD74A2DFD03DCAE921351AC
- MessageActionsC32FC4679CBD0BC193E0936E05D4788DEE4D689BFC7DB8C98828A76712419B40
- Picker74BF01EBB4988F21B63E14A6F0948B8FDEA3AADF3C8D6AC12F89510E58BB74F5
- ReactionHook995B9FBB8564CBF7E5E23A9859DEFC3844E6439542620D8F6389DC4858B04D26
- BookmarkHookC587FD74389B8FF111C9A88D6D96E54A0416FFD54F09CB3F8D2B3F8BF4E6BB26
- EN F03C8F8695F08FEEFE90A61C67804F82D8C8DB21FE7B138B62D6BF7DC10EEAC9
- RU0231B1040903C5BE35A6E669DEBD3F56C7BAA1FB4FA10E27EDB92C6F37EC4BA6
- ReactionOracleEE316695F754793AC8871915D3011DA1D324B6B4A564D0992568834B1C1C74C5
- BookmarkOracle9E2C85CD86A6F23840C184E26BB2E1DDCD6341EFCAF15DEB4FFE589A1B492E4B

This is a composed reaction+bookmark candidate based on c0900e62, not runtime
acceptance of either intermediate commit in isolation. Root index preserves
reaction-only source snapshot for separate publication; bookmark adapters
remain unstaged until the first commit. Browser/API/WS/storage/media and full
module readiness are NOT_RUN/UNBOUND. Independent final receipt review
C/I/M0/0/2 UNIT_ACCEPTED for genuine frontend units. The minor notes retain
the explicit non-existent JSON count and sampled/not-captured telemetry
limitations above; they are not source/runtime failures. Publication pending.
Publication completed:0d562ee855abfe1180646fe81a5a1ab8008cd475 committed,
immediately pushed to authorized calls-feature-codex/origin and exact remote
SHA equality verified. It contains only the frozen reaction source snapshot,
25-case oracle, two fixture props and reaction/root docs, not bookmark source.
The disclosed composed GREEN remains the runtime provenance.


Canonical prior GREEN is baseline evidence, not this atom's acceptance.
Browser, real API/storage/WS/media, Jitsi activation and complete modules
remain NOT_RUN/UNBOUND. Preserve failed receipts and source hashes; no
change to tests to hide an observed product defect. One retry per transient
runtime failure; at most three fixes per root cause and five review rounds.
No model/runtime fallback. Independent reviewers are read-only; one writer
owns each exact file boundary. Root records blockers and continues other
unblocked agreed work without expanding authority.
