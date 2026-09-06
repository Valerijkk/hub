# Bookmark removal lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` task-by-task. Root alone owns canonical Docker QA, docs sync, commits, and push; record checkboxes as work is performed.

**Goal:** Each visible bookmark can be removed by one current actor-scoped request, with row-local pending UI, exact refresh, current error/retry, and no stale UI or cache effect.

**Architecture:** Keep lifecycle state inline in `BookmarksPanel`; do not introduce a hook or framework. A committed actor-scope object and request objects stored in a per-bookmark-ID `Map` make same-ID work single-flight, permit distinct IDs concurrently, and reject stale A-B-A/unmounted completions; a `Set` supplies row-local pending rendering.

**Tech Stack:** Existing React 19, TypeScript, Zustand auth store, TanStack Query, Vitest/RTL, and `toastApiError`; no dependency or API additions.

**Spec:** Frozen root design in the 2026-09-06 communications completion queue and current source contracts: `frontend/src/components/messenger/bookmarks-panel.tsx`, `removeBookmark(bookmarkId)` in `frontend/src/api/messenger.ts`, and `bookmarksQueryKey(userId)`.

## Global constraints

Root freeze 2026-09-06: behavior accepted. Independent plan review found one
stale runtime reference in the final handoff; it is now reconciled with the
global routing (oracle/source Terra/xhigh; reviewer Terra/high). C/I remaining:
0/0. Oracle application waits for exact forward publication and root base
binding. The current ForwardDialog GREEN must remain an isolated candidate.

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

- [ ] **Step 1: Create a standalone failing oracle on the frozen baseline.**

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

- [ ] **Step 2: Root runs the canonical RED, without editing production.**

  Root command only:

  ```powershell
  docker compose -f docker-compose.local.yml run --rm frontend-unit
  ```

  Expected: the new lifecycle cases fail because baseline permits duplicate same-ID deletes, lacks the new row-pending UI span, and can expose stale completion effects. Baseline already awaits exact current-user invalidation; the new oracle proves the additional requirement that visible pending remains through that awaited refetch. Type/fixture failures, old-suite failures, timeout changes, missing resources, and unhandled rejections are not intended RED; root classifies them before authorization to implement. Preserve receipts and mark execution `NOT_RUN` if the exact service cannot be started.

- [ ] **Step 3: Have the designated single production writer implement only the frozen inline algorithm.**

  Add the committed scope, `Map`, pending `Set`, current predicate, and guarded `finally` exactly at the panel boundary. Replace only `handleRemove`; add the row's `disabled` prop. Capture `bookmark.id` and `scope.actorId` before awaits. Await exact invalidation only after a current DELETE success. Keep literal current error fallback and let catch consume current invalidation rejection. Do not alter query enablement, navigation, labels, API/store contracts, authorities, existing error view/retry, or other files.

- [ ] **Step 4: Independently review before GREEN.**

  A read-only reviewer verifies the diff uses object identity for scope and request ownership, does not use user-ID comparison as the sole ABA test, only disables a matching row, awaits the exact actor key, preserves the formatter/Russian fallback, and remains within `<=150` production added+deleted lines. Review the new oracle for all ten cases, including both stale outcomes and both unmount outcomes; reject UI/internal-state mocking, weakened existing fixtures, broad invalidation, or speculative authority changes.

- [ ] **Step 5: Root runs the canonical GREEN.**

  Run the same exact Docker command from Step 2. Expected only on successful execution: TypeScript and the complete `frontend-unit` suite pass, all ten new observed cases pass, and no whole-run unhandled errors/rejections are present. Do not claim GREEN from static review, a focused host run, or an earlier composed candidate; record Docker infrastructure failure as `NOT_RUN`/`BLOCKED_INFRA`.

- [ ] **Step 6: Root performs docs-sync and publication only after GREEN and review.**

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

Placeholder/type scan: complete. `RemovalScope`, `RemovalRequest`, `isCurrent`, all external calls, exact key, error fallback, tests, limits, and Docker command are defined above. No production work, test execution, docs-sync, commit, or push has occurred under this plan.

## Execution handoff

Plan saved at `docs/superpowers/plans/2026-09-06-bookmark-removal-lifecycle.md`. Root behavior acceptance and plan review are recorded above; exact prerequisite publication binding is still required before oracle application. Use a Terra/xhigh oracle writer for the new test path, then a separately dispatched Terra/xhigh source writer for the single panel path after accepted RED; Terra/high owns independent read-only review. Root retains RED/GREEN/docs/Git; actual runtime remains UNKNOWN without telemetry.
