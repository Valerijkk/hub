# Bookmark removal lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` task-by-task. Root alone owns canonical Docker QA, docs sync, commits, and push; record checkboxes as work is performed.

**Goal:** Each visible bookmark can be removed by one current actor-scoped request, with row-local pending UI, exact refresh, current error/retry, and no stale UI or cache effect.

**Architecture:** Keep lifecycle state inline in `BookmarksPanel`; do not introduce a hook or framework. A committed actor-scope object and request objects stored in a per-bookmark-ID `Map` make same-ID work single-flight, permit distinct IDs concurrently, and reject stale A-B-A/unmounted completions; a `Set` supplies row-local pending rendering.

**Tech Stack:** Existing React 19, TypeScript, Zustand auth store, TanStack Query, Vitest/RTL, and `toastApiError`; no dependency or API additions.

**Spec:** Frozen root design in the 2026-09-06 communications completion queue and current source contracts: `frontend/src/components/messenger/bookmarks-panel.tsx`, `removeBookmark(bookmarkId)` in `frontend/src/api/messenger.ts`, and `bookmarksQueryKey(userId)`.

## Global constraints

Current execution: UNIT_ACCEPTED / PUBLISHED; composed Docker GREEN below.
Commit49b0f1330c9d365749aeea3f7824ca96f505f94f immediately pushed to the
approved origin calls-feature-codex; remote SHA equality verified2026-09-06.
Independent source and final receipt review C/I/M0/0/0.
Runtime assignment paragraphs below are chronology,
not instructions to repeat failed dispatches or rewrite the frozen oracle.

Current assignment decision: after the panel oracle correction finishes,
root will reuse that available Terra/xhigh writer in a separate turn for
this disjoint oracle. This supersedes the unsuccessful Astra reuse below;
no concurrent dispatch retries. Root reviews both oracles before the shared
RED. Source and independent review routes remain Terra/xhigh and Terra/high.

Runtime rebind 2026-09-06: the new Terra/xhigh oracle dispatch was rejected
by the live agent-thread limit before any work. Root explicitly assigns
only oracle authoring to the existing idle Astra/high worker under current
model-routing; this supersedes the oracle route in historical text below,
not the frozen behavior. Source remains separately gated Terra/xhigh;
independent review remains Terra/high. No unavailable runtime is claimed
to have run; observed runtimes remain UNKNOWN. This is an explicit root
dispatch decision, not an inherited/silent fallback.

Live outcome: the explicit Astra/high reuse was also rejected by the same
thread limit before work. Bookmark oracle is NOT_WRITTEN / WAITING_FOR_SLOT;
no runtime fallback was executed. Only the panel oracle writer is dispatched.
Retry assignment only after capacity changes; do not poll/spawn-loop.

Root freeze 2026-09-06: behavior accepted. Independent plan review found one
stale runtime reference in the final handoff; it is now reconciled with the
global routing (oracle/source Terra/xhigh; reviewer Terra/high). C/I remaining:
0/0. Forward prerequisite published as 74c5c492a41998213fcf7cf1556fdf6333c7ee42,
immediate push and exact remote SHA equality verified on 2026-09-06. Oracle
application is bound to this clean base (201 files / 871 tests PASS).
Root batching ruling: this independent oracle and panel-unpin's 34-case oracle
may share canonical RED/GREEN; each keeps disjoint ownership and per-file
triage. Both oracles yield expected 203 files / 915 tests. No editing tests
during a running gate. Source still waits for reviewed oracle and accepted RED.

- This is one frontend removal lifecycle only. Do not change role/channel authority, `removeBookmark` API shape, backend cancellation/idempotency, automatic retry, cache-key shape, query policy, or server behavior.
- Production ownership is only `frontend/src/components/messenger/bookmarks-panel.tsx`; production sum of added plus deleted lines must be at most 150. Preserve all existing tests and fixtures.
- Create only `frontend/tests/messenger/bookmark-removal-lifecycle.test.tsx`, at most 600 physical lines. It must use the real panel, auth store, `QueryClient`, and visible controls; mock only HTTP, router, and the canonical error formatter.
- Capture immutable actor ID and bookmark ID before the first `await`. Use committed actor-scope object identity, not equality of user strings alone: committed A -> B -> A must not revive A's old completion.
- The request lock is `Map<bookmarkId, request>` and pending UI is `Set<bookmarkId>`. Same ID in the same committed scope is one flight; separate bookmark IDs may run at once. A request remains pending until its exact `bookmarksQueryKey(capturedActorId)` invalidation resolves.
- Preserve the current formatter and literal Russian fallback: `toastApiError(error, "Не удалось удалить закладку")`. Current error leaves the row retryable; no automatic retry and no success toast.
- An obsolete success/error after actor ABA or unmount has no toast, cache invalidation, or UI effect. Its ownership-only `finally` must never erase a newer request or unlock its row.
- No `pro-qa-ai/**`, Mail/Calendar, Docker execution, host tests/validators/install, Git mutation, or additional writers. This proposal needs root acceptance and independent review before production work.
- Proposal-author routing was `gpt-5.6-terra` / `medium`; actual runtime: `UNKNOWN` unless telemetry records it. The frozen downstream routing is `gpt-5.6-terra` / `xhigh` for the oracle stage and atomic implementation, then `gpt-5.6-terra` / `high` for independent read-only review. Do not silently substitute a runtime.

## File structure and confirmed contracts

| Path | Responsibility |
| --- | --- |
| `frontend/src/components/messenger/bookmarks-panel.tsx` | Existing user-scoped query and remove button; gain only inline scope/request/pending lifecycle and `disabled` on the affected row button. |
| `frontend/tests/messenger/bookmark-removal-lifecycle.test.tsx` | Standalone deferred-promise lifecycle oracle; no edits to identity/error/preview tests. |

Current source facts that the implementation must retain:

- `bookmarksQueryKey(userId)` is `['messenger', 'bookmarks', userId]`; a successful current removal must call `queryClient.invalidateQueries({ queryKey: capturedKey, exact: true })` and await it.
- `removeBookmark(bookmark.id)` is the existing DELETE boundary; no optimistic list update is permitted.
- The panel's query is disabled without an actor and the existing no-user cache is deliberately not rendered.
- Current code already awaits exact invalidation after a current DELETE and verifies the auth store, but has no pending UI span, per-ID lock, unmount guard, or ABA-safe committed scope identity. This atom covers only committed actor A -> B -> A transitions. A batched/noncommitted actor-loss latch is a residual outside this atom, not a full actor-switch guarantee.

## Frozen inline algorithm

Use `useLayoutEffect`, `useRef`, and `useState` imports in the panel. The effect is the only producer of the committed scope; it must run for the subscribed `userId`, create a fresh object every committed actor transition, and invalidate its own identity on cleanup. Do not write these refs during render.

```ts
type RemovalScope = { actorId: string | null };
type RemovalRequest = { scope: RemovalScope; bookmarkId: string };

const [pendingIds, setPendingIds] = useState<Set<string>>(() => new Set());
const scopeRef = useRef<RemovalScope | null>(null);
const requestsRef = useRef(new Map<string, RemovalRequest>());

useLayoutEffect(() => {
  const scope = { actorId: userId };
  scopeRef.current = scope;
  requestsRef.current = new Map();
  setPendingIds(new Set());
  return () => {
    if (scopeRef.current === scope) {
      scopeRef.current = null;
      requestsRef.current = new Map();
    }
  };
}, [userId]);

const isCurrent = (request: RemovalRequest) =>
  scopeRef.current === request.scope &&
  Boolean(request.scope.actorId) &&
  useAuthStore.getState().user?.id === request.scope.actorId &&
  requestsRef.current.get(request.bookmarkId) === request;

const handleRemove = useCallback(async (bookmark: Bookmark): Promise<void> => {
  const scope = scopeRef.current;
  const bookmarkId = bookmark.id;
  if (!scope?.actorId || !bookmarkId ||
      useAuthStore.getState().user?.id !== scope.actorId ||
      requestsRef.current.has(bookmarkId)) return;

  const request = { scope, bookmarkId };
  requestsRef.current.set(bookmarkId, request);
  setPendingIds((prior) => new Set(prior).add(bookmarkId));
  try {
    await removeBookmark(bookmarkId);
    if (isCurrent(request)) {
      await queryClient.invalidateQueries({
        queryKey: bookmarksQueryKey(scope.actorId), exact: true,
      });
    }
  } catch (error) {
    if (isCurrent(request)) toastApiError(error, "Не удалось удалить закладку");
  } finally {
    if (scopeRef.current === scope && requestsRef.current.get(bookmarkId) === request) {
      requestsRef.current.delete(bookmarkId);
      setPendingIds((prior) => {
        const next = new Set(prior); next.delete(bookmarkId); return next;
      });
    }
  }
}, [queryClient]);
```

Render `disabled={pendingIds.has(bm.id)}` on the existing remove `Button`; retain its accessible Russian name and all markup/navigation behavior. The `finally` deliberately checks only ownership identity before releasing; it must not call the broader `isCurrent`, because a stale scope must not mutate the new scope and may already be unmounted. A query-invalidation rejection is caught as a current error exactly like a DELETE rejection; a stale invalidation completion cannot release a newer request because `Map` ownership differs.

## Task 1: Test-first per-bookmark removal lifecycle

**Files:**

- Create: `frontend/tests/messenger/bookmark-removal-lifecycle.test.tsx`
- Modify after accepted RED only: `frontend/src/components/messenger/bookmarks-panel.tsx:5-16,36-76,134-141`
- Preserve: `frontend/tests/messenger/bookmarks-error.test.tsx`, `bookmarks-preview.test.tsx`, `phase151-bookmark-identity-boundary.test.tsx` and all their fixtures.

**Interfaces:**

- Consumes: `removeBookmark(bookmarkId: string): Promise<void>`, `BookmarksPanel`, `bookmarksQueryKey(userId: string)`, `useAuthStore`, `QueryClient.invalidateQueries`, and `toastApiError(error, fallback)`.
- Produces: no exported interface. The existing remove button is disabled iff its bookmark ID is in the current scope's pending `Set`.

- [x] **Step 1: Create a standalone failing oracle on the frozen baseline.**

  Reuse the test setup idioms from `phase151-bookmark-identity-boundary.test.tsx`: `QueryClient` with retries disabled, real `renderWithProviders`, a `deferred<T>()`, two authenticated user fixtures, two bookmarks, and mocked `getBookmarks`, `removeBookmark`, router, and `toastApiError`. Do not import a future helper or add props to the panel.

  Include these ten concrete cases (parameterize success/error pairs only where each outcome remains observed):

  1. Same-tick double click of bookmark A makes exactly one `removeBookmark(A.id)` while A's button is disabled.
  2. With A pending, bookmark B starts `removeBookmark(B.id)` and both row buttons are disabled independently; no global lock occurs.
  3. Current success keeps A disabled while a `QueryObserver` makes exact current-user invalidation await a deferred refetch; it unlocks only after that promise resolves and invalidates no other user key.
  4. Current rejected DELETE calls `toastApiError(error, 'Не удалось удалить закладку')` once, does not invalidate, unlocks A, and one later click is the only retry.
  5. Actor A -> B -> A through separate `act` commits, then stale A success: no invalidation and no release of a newly started current A request.
  6. The same committed actor ABA with stale A rejection: no toast and no release of the newer A request.
  7. Unmount while current DELETE is pending, then resolve: no invalidation and no unhandled rejection.
  8. Unmount while current DELETE is pending, then reject: no toast and no unhandled rejection.
  9. No authenticated actor: the query remains disabled; neither rows nor remove controls render, `getBookmarks` and `removeBookmark` are never called, and cached old/no-user rows stay hidden. A control captured before an actor transition may be used only as a supplementary stale-callback witness, not as the public no-actor assertion.
  10. A rendered bookmark with `id: ''`: clicking cannot issue DELETE or create pending UI.

  Representative current-error/retry oracle (the final test may use helpers but must retain these observations):

  ```tsx
  const failure = new Error('delete failed');
  vi.mocked(removeBookmark)
    .mockRejectedValueOnce(failure)
    .mockResolvedValueOnce(undefined);
  renderPanelWith([bookmarkA]);
  const remove = await screen.findByRole('button', { name: /удалить из закладок/i });
  fireEvent.click(remove);
  await waitFor(() => expect(toastApiError).toHaveBeenCalledWith(
    failure, 'Не удалось удалить закладку',
  ));
  expect(invalidate).not.toHaveBeenCalled();
  expect(remove).toBeEnabled();
  fireEvent.click(remove);
  await waitFor(() => expect(removeBookmark).toHaveBeenCalledTimes(2));
  ```

- [x] **Step 2: Root runs the canonical RED, without editing production.**

  Root command only:

  ```powershell
  docker compose -f docker-compose.local.yml run --rm frontend-unit
  ```

  Expected: the new lifecycle cases fail because baseline permits duplicate same-ID deletes, lacks the new row-pending UI span, and can expose stale completion effects. Baseline already awaits exact current-user invalidation; the new oracle proves the additional requirement that visible pending remains through that awaited refetch. Type/fixture failures, old-suite failures, timeout changes, missing resources, and unhandled rejections are not intended RED; root classifies them before authorization to implement. Preserve receipts and mark execution `NOT_RUN` if the exact service cannot be started.

- [x] **Step 3: Have the designated single production writer implement only the frozen inline algorithm.**

  Add the committed scope, `Map`, pending `Set`, current predicate, and guarded `finally` exactly at the panel boundary. Replace only `handleRemove`; add the row's `disabled` prop. Capture `bookmark.id` and `scope.actorId` before awaits. Await exact invalidation only after a current DELETE success. Keep literal current error fallback and let catch consume current invalidation rejection. Do not alter query enablement, navigation, labels, API/store contracts, authorities, existing error view/retry, or other files.

- [x] **Step 4: Independently review before GREEN.**

  A read-only reviewer verifies the diff uses object identity for scope and request ownership, does not use user-ID comparison as the sole ABA test, only disables a matching row, awaits the exact actor key, preserves the formatter/Russian fallback, and remains within `<=150` production added+deleted lines. Review the new oracle for all ten cases, including both stale outcomes and both unmount outcomes; reject UI/internal-state mocking, weakened existing fixtures, broad invalidation, or speculative authority changes.

- [x] **Step 5: Root runs the canonical GREEN.**

  Run the same exact Docker command from Step 2. Expected only on successful execution: TypeScript and the complete `frontend-unit` suite pass, all ten new observed cases pass, and no whole-run unhandled errors/rejections are present. Do not claim GREEN from static review, a focused host run, or an earlier composed candidate; record Docker infrastructure failure as `NOT_RUN`/`BLOCKED_INFRA`.

- [x] **Step 6: Root performs docs-sync and publication only after GREEN and review.**

  Root updates the completion queue/`STATE.md` and any required acceptance docs, preserving actual RED/GREEN receipts and limitations. Root alone measures the exact diff, stages exact owned paths, creates the authorized conventional commit without AI attribution trailers, pushes only the approved branch, and verifies remote SHA equality. This plan itself does not authorize Git mutation.

## Self-review / acceptance map

| Frozen requirement | Plan evidence |
| --- | --- |
| Per-ID single-flight; distinct rows concurrent | Cases 1-2; `Map` keyed by captured `bookmarkId` and `Set` rendering. |
| Pending includes exact invalidation | Case 3; await inside the current-success branch. |
| Current formatter/fallback and explicit retry | Case 4; unchanged `toastApiError` literal. |
| Actor ABA, success and error safe | Cases 5-6; committed scope object and request ownership. |
| Unmount, success and error safe | Cases 7-8; effect cleanup invalidates scope. |
| No actor / empty ID fail closed | Cases 9-10; before-flight admission. |
| No authority/API/retry/framework expansion | Global constraints and single-file implementation boundary. |

Original plan self-review: `RemovalScope`, `RemovalRequest`, `isCurrent`, all external calls, exact key, error fallback, tests, limits, and Docker command were defined before implementation. Current execution evidence follows.

## Canonical combined RED — 2026-09-06

Actual oracle author: sequential reuse of Terra/xhigh panel writer, not the
unsuccessfully requested Astra reuse.10cases/252physical lines,
SHA83350B577F43E039A98BFC40E19E21F5B2618016D3DBBA21C590B00237368AAB.
Root full read and independent READY_FOR_RED. Canonical TypeScript PASS;
203files201PASS/2FAIL,915tests887PASS/28FAIL,0pending/todo,126.19s,exit1;
bookmark2PASS/8FAIL, panel14PASS/20FAIL. All201oldfilesPASS. Eight bookmark
failures:duplicate DELETE,distinct-row pending,real-refetch pending,2ABA
pending witnesses,unmounted success invalidation,unmounted error toast,
emptyID admission. TwoABA cases sendrequest2 then fail disabled BEFORE old
settlement; later isolation assertions NOT_EXECUTED. Current error/retry
and noactor baseline protections pass. All9runner unhandled rejections
belong to the separate panel handler, not bookmarks. Source remains unchanged;
independent triage accepts ROOT=code, C/I0/0 READY_FOR_SOURCE. Only the frozen
single-file source stage is authorized, not product acceptance.

Shared eight receipts and full inspected2GiB/no-swap6CPU/two-worker provenance:
`testframework/reports/frontend-unit-panel-actions-red-20260906/` and the
panel-unpin plan's Canonical combined RED section. Same source/oracle/Compose
hashes after run; no build/pull/install or adjacent Mail changes. One approval
timeout occurred before execution, then one allowed retry launched the only
runner. Container absent after exit; peak/OOM NOT_CAPTURED. No claim of
API/browser/storage/build/lint acceptance.

## Execution handoff

Oracle and source stages are complete; do not repeat them. Root retains
docs/Git/publication. Author requested Terra/xhigh; independent reviewer
Terra/high; actual runtime UNKNOWN without telemetry.

## Canonical composed GREEN — 2026-09-06

Base74c5c492 plus both panel/bookmark atoms; root exact
`docker compose -f docker-compose.local.yml run --rm frontend-unit`:
TypeScript PASS,203 files /915 tests PASS, all10 bookmark and34 panel cases,
exit0,136.52s,start11:37:40UTC. Failed/pending/todo/whole-run unhandled0.
Both previously unreached ABA settlement tails and actual refetch pending
now pass. No oracle or old-fixture edits; this does not extend the contract
to batched/noncommitted actor ABA, API/browser or server cancellation.

Single-file production +67/-22=89/150; source SHA
BD9558EE0E2210AE8D24708C2D4BC890A4BA81A234FE34E6C7D0214C8A95AB71.
Frozen10case/252line oracle SHA83350B577F43E039A98BFC40E19E21F5B2618016D3DBBA21C590B00237368AAB.
All source/oracle/old-fixture/Compose hashes match before/after GREEN.
Independent source review spec/quality compliant, C/I/M0/0/0. Final receipt
review UNIT_ACCEPTED / READY_FOR_PUBLICATION, C/I/M0/0/0; counts and absence
of whole-run unhandled confirmed directly from raw receipts.

Eight raw files preserved without overwrite at
`testframework/reports/frontend-unit-panel-actions-green-20260906/`;
RED remains alongside. Full names/resource/runner provenance and observed
late Mail build overlap are in the panel-unpin plan's GREEN section and
`.superpowers/sdd/2026-09-06-panel-actions/green-receipt.md`.
Actual2GiB/no-swap6CPUreadonlytwo workers; no root build/pull/install or
Mail changes. Existing act/DialogContent/mock warnings remain; build/lint/
API/browser/storage NOT_RUN. Samples are not peaks and continuous workload
exclusivity is not claimed. Separate commits use composed-candidate evidence,
not independently executed intermediate revisions.
