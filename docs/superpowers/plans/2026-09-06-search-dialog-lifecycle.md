# SearchDialog scoped request lifecycle

Status: UNIT_ACCEPTED in the composed2026-09-06 Docker frontend gate.
Intended RED retained; two un-applied whole-file payloads rejected and closed
at fix bound. Decomposed local hook/UI and oracles independently reviewed
C/I/M0/0/0;23 real UI plus12 genuine hook cases now pass.
Baseline calls-feature-codex f1691e34; concurrent read-receipts changes are
task-owned and must be preserved. This is one search interaction atom, not
message deep-link navigation or full Messenger acceptance.

## Outcome and boundaries

Explicit search shows only the result of the current submitted query in the
current authenticated dialog session. Successful empty is distinct from
failure. Old success/error cannot reappear after draft/scope/channel/user
changes or closing and reopening. Existing channel-only navigation remains.

Ownership after root decomposition (sequential writers,2026-09-06):
- frontend/src/components/messenger/search-dialog-session.ts: new component-
  local session hook,150-180lines target; no shared infrastructure;
- frontend/tests/messenger/search-dialog-session.test.tsx: targeted hook
  regression for exact attempt cleanup and retained selection callbacks;
- frontend/src/components/messenger/search-dialog.tsx: owner/presentation
  integration only, <=100changedlines target, preserve original markup;
- frontend/tests/messenger/search-dialog.lifecycle.test.tsx new genuine unit;
- root owns this plan and relevant same-session domain/status docs.
Keep existing a11y-names-p2.test.tsx unchanged. No API, shared hook, store, shared UI,
locale, backend, routes, dependencies, test launcher or Mail/Calendar changes.
Aim <=200-300 changed production lines; return BLOCKED for decomposition if
the atom requires a new framework, shared cache or broader feature.

## Frozen design

### Accepted two-atom decomposition

Atom1 exports useSearchDialogSession({userId:string|null,channelId?:string}).
Its result is draft:string, scope:channel|global, setDraft(string),
setScope(scope), submit(), retry(), pending:boolean, successEmpty:boolean,
error:boolean, results:readonly Message[], and currentSelectionGuard:
((message:Message)=>boolean)|null. A retained guard captures the attempt
snapshot and must return false after discard, newer attempt, identity/revoke
change or disposal, even if a new result has the same message ID.

The hook owns the unique per-mounted-open-session ID, immutable attempt
snapshot, synchronous active/in-flight refs, Query data and tiny commit/layout
mounted/revoke fence. Key shape remains the original frozen contract.
Discard/replace cancels/removes only the exact prior full attempt key; final
session cleanup uses its unique prefix. No per-attempt cleanup effect may
cancel/remove a newer attempt by prefix. An internal disabled idle Query
record is allowed, but it must never fetch from typing or mounting alone.
All existing no-auto retry/refetch and UI authority guarantees still apply.

Atom2 preserves the current markup and date formatter and wires the hook.
Root explicitly selects a keyed identity/channel ScopedDialog owner with
conditional per-open child (no passive openEpoch increment). Committed
identity/channel change remounts the owner; reopening mounts a fresh child.
Keep scope Buttons, sender_id/content_text and original row styles.

The original23-case real UI oracle is frozen and its18 intended RED failures
are the existing behavioral failure evidence for this decomposition. Hook
regressions must be authored before its implementation and stay genuine
runtime tests; root will evaluate the composed UI at the next full gate.
No claim that an intermediate hook alone fixes the SearchDialog. The prior
writer is stopped, not re-dispatched for another whole-file rewrite.

Root implementation preflight clarification (first un-applied payload):
per-SearchDialog counters that restart at mount and passive-effect session
increments are rejected. Key the owner by the identity/channel tuple so a
committed change discards its open state; each conditional open session uses
a unique ephemeral ID that cannot collide across mounts/owners. A tiny commit
mounted fence rejects callbacks after disposal. A submitted-attempt child
may own useQuery to avoid an idle placeholder query. Keep explicit retry and
submit admission shared. This is local component decomposition, not a new
shared framework. Initial oversized payload was not applied or tested;
writer remediation attempt1 is active. Existing oracle remains unchanged.

Server results belong TanStack Query, not duplicated Message[] useState.
Keep draft input/scope/open/submission metadata local. Separate immutable
submitted search from editable draft. Existing searchMessages(trimmedQuery,
effectiveChannelId?) returns Message[]; do not change its API or pretend it
supports AbortSignal. An unknown identity or revoked current channel denies
submit and navigation. Check actual store state again inside event handlers.

Use a mounted identity/channel-scoped search session. Each open session has
a unique ephemeral identifier; closing unmounts/discards it. A committed
user/channel A-B-A transition must create a distinct session, including after
logout. Query keys start [messenger,search-dialog,userId,sessionId] and include
immutable submitted query/scope/effective channel plus explicit attempt ID.
Neither user-independent keys nor a counter restarting at the same identity
and remount position alone may reconnect a prior in-flight result.

Only explicit submit creates a query attempt. No eager fetch on typing,
automatic retry, polling, focus/reconnect/refetch-on-mount, or background
invalidations that execute a disabled draft. Use retry:false and explicit
refetch controls. One synchronous submission guard plus Query dedup prevents
duplicate Enter/click before React rerenders. Retry is a new explicit attempt
for the still-current failed submission; it cannot replay obsolete input.

Draft edits and scope changes discard the submission immediately, including
changing away and back before a new submit. Mask obsolete pending/success/
error before rendering, not only after a passive cleanup. Current identity,
session, channel, scope and trimmed draft must still match the submitted
snapshot. A new explicit submission may proceed after discarding an older
logical attempt; late transport completion has no current UI effect.

Use transient Query lifetime (gcTime:0) and exact session-prefix cancellation/
removal on discard where needed, without clearing other users' or other
Messenger queries. Cancellation is a local Query lifecycle operation, NOT a
guarantee that the existing HTTP request/server work is aborted. No retained
search-history feature. Do not mutate lifecycle refs during render; use keyed
mount/commit cleanup and synchronous event invalidation as appropriate.

Render states:
- no current submitted attempt: no results, empty caption or error;
- current pending: existing search loading affordance;
- current rejected: inline role=alert with common.search.error/errorHint and
  keyboard-operable common.btn.retry Button type=button;
- current fulfilled []: existing Russian no-results caption;
- current fulfilled messages: existing result rows.
Use useTranslation(common) for the existing common keys; do not add raw
exception details, toasts, untranslated strings or new locale files. Preserve
existing trigger accessible name, scope buttons, layout and date formatting.

Before result selection recheck live user identity, current submitted/session
validity and both current-channel and result-channel revoke state. A revoked
result must not remain an actionable row. This is UI defense, not proof of
server authorization. Allowed selection still closes the dialog and navigates
to /messenger/$channelId using params, with no invented message/thread URL.

## Test-first acceptance

Render the REAL SearchDialog, QueryClient and auth/revoke stores. Mock only
searchMessages HTTP boundary and navigate plus necessary browser boundaries.
Use isolated clients and clean up stores/queries after each case. No source
text/AST contract tests, no mocked target logic, no missing-helper RED.

1. Initial/open and whitespace-only draft: no API and no false empty/error.
2. Trimmed explicit submit in channel and global scopes: exact API arguments.
3. Fulfilled messages and fulfilled [] render their separate success states.
4. Rejection displays localized error/retry, not empty; explicit retry sends
   exactly one new request and recovers successfully.
5. Repeated Enter plus click in the same pending submission sends once.
6. Deferred old success/error after editing draft (including away-back) stays
   hidden until a new explicit submission; new result is not overwritten.
7. Deferred old success/error after scope change stays hidden.
8. Close/reopen cannot restore old result/error/pending or trigger a fetch.
9. Committed channel A-B-A and user A-B-A ignore old deferred completion.
10. Unknown user and revoked current channel cannot submit; live revoke of a
    result channel cannot navigate. Valid selection preserves existing route.
11. Unmount and discarded failure settle without unhandled rejection or
    affecting unrelated Query keys; exact transient cleanup is scoped.
12. No automatic retries/focus/reconnect request; no test timeout inflation.

Independent Terra/high lifecycle preflight: C/I/M0/0/0, STATIC_ONLY. Oracle
author then caught an incorrect explorer key attribution: root read both
common.json files and corrected actions.retry to the actual btn.retry. The
visible Retry / Повторить copy and no-locale-edit contract are unchanged.

Root owns canonical Docker RED/GREEN and serial workload checks:
docker compose -f docker-compose.local.yml run --rm frontend-unit
Use existing image,2GiB/no-swap; no host QA, alternate command, build/pull or
parallel heavy job. Independent Terra/high preflight/review; Terra/xhigh
test-first atomic writer after root freeze. Review/fix max2 cycles per cause,
then decompose/escalate. Do not claim browser/API search authorization tested
by these units. Root alone commits accepted exact paths then immediately pushes only
calls-feature-codex to origin https://github.com/Valerijkk/hub.git and verify.

## Execution checkpoint (2026-09-06 local)

Initial oracle collection attempt was ROOT=test, NOT behavioral RED:
undeclared @testing-library/user-event import could not resolve. No package
install or image rebuild was used. Mechanical oracle-only repair uses the
existing RTL fireEvent and real change/click/Enter/document-Escape paths.
Independent targeted preflight C/I/M0/0/0; all first-request/cache witnesses
and assertions retained. Reports: frontend-unit-search-collection-red-20260906
under testframework/reports; old666tests passed; TypeScriptPASS,57.88s;
engine7ede6b034f4a5c934fdf3ed9eba6245e5519044d72d633f6b23078b9f66345ec,
started2026-09-05T22:15:58.683975631Z, exit1/74s, destroyed, no lifecycle OOM.

Canonical intended RED command:
docker compose -f docker-compose.local.yml run --rm frontend-unit
Engine32744562cc9574c1df67509a54365bcc362a967d5cdfe01ec1455b6ab19dabca,
outer worktree-master-frontend-unit-run-baab676db7db,
start2026-09-05T22:27:07.028166457Z; inspected2GiB/no-swap,6CPU,readonly.
Image sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
TypeScriptPASS;671PASS18FAIL/689tests,192files,59.97s. All18 failures are new expected
missing lifecycle behavior; all previous666tests passed. ROOT=code: false
whitespace empty state, absent error/retry, duplicateEnter, stale draft/scope/
close/identity/channel completions, unknown/revoked submit, revoked result and
missing transient Query ownership. No test expectations were relaxed.
Compose/engineexit1,77s; create/start/die/destroy observed, no lifecycle OOM;
container absent afterward. Reports preserved as eight current small files:
testframework/reports/frontend-unit-search-lifecycle-red-20260906/.
Oracle SHA25670E4899F92BC1560D57BAC754BF8B27C207159F1C509A8BCF24B21F9CA73D7EA
unchanged before/after; production baseline SHA256
3797B572FDB55F1632DEE347DC9FECD4C6847C8E8817BACAE28A362671949BA3.
No build/pull/install/hostQA. Adjacent Mail builds completed before start;
resident Mail services were left untouched. Production writer owns only
search-dialog.tsx; oracle is frozen. Runtime model/provider telemetry UNKNOWN.

## Decomposed source checkpoint

search-dialog-session.ts and12 genuine hook cases now applied; independent
Terra/high source/oracle review C/I/M0/0/0. Before application root closed
raw-revoked-result projection, query-time live admission and never-cleared
inflight issues; a second narrow fix captures the FAILED render attempt in
retry instead of retargeting through activeRef. Added regression proves an
old retry callback cannot restart a later successful query. Old whole-file
rewrites remain unapplied. Root added only cancellation comments and full
mockReset between hook tests. Current hook SHA256
F2E42F4FBF3BCE0445FB250A2FD809340A1FB3074DD2B9ADA67FDA66B983B8A0;
oracle8DA7792DFB2E5F36AF150B53DCAC30C098EDCB35C771DA9E484C768DF8A5DF0E.
These are source-only candidates, not a composed SearchDialog fix or GREEN.
The independent reviewer corrected an invalid carried-over line reference;
that reference is not used as evidence. UI adapter owns only original
search-dialog.tsx and must preserve scope Buttons/date/row markup.

## Composed GREEN and acceptance (2026-09-06 local)

Canonical composed-tree gate (search + delete + thread wiring):
docker compose -f docker-compose.local.yml run --rm frontend-unit
TypeScript PASS;195 files/728 tests PASS,64.12s,Compose exit0. No whole-run
unhandled-error failure. Existing unrelated React act/description warnings
remain; this is not a clean-stderr claim. Observed engine
61238c7ed829fb461c15326296f38aa946cb1ebc330b4d99305b3a749a05536d,
outer worktree-master-frontend-unit-run-8768e627cfd5,
start2026-09-05T23:44:23.616980053Z, image
sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218.
Inspected2GiB/no-swap,6CPU,readonly; container absent after completion.
Post-run Docker event lookups returned no records, so separate die/destroy
timestamps and OOM-event history are NOT_CAPTURED, not invented.
Eight current reports preserved at
testframework/reports/frontend-unit-messenger-lifecycle-green-20260906/.
No source/oracle changes during execution; recorded pre/post hashes match.
No build/pull/install/hostQA or modification of resident Mail services.
This tests the composed working tree, not each intermediate publication
commit in isolation. Browser/API/storage/media and whole modules remain
NOT_RUN/UNBOUND.

Final original-markup UI adapter SHA256
6A9F0B9929CECC6ACFD5682B9B3DE9A290BCF73C98CD2C6C596BEE40C8B37B3E;
hook and both frozen search-oracle hashes above are unchanged. The accepted
adapter preserves scope Buttons, original sender/content/date/row styles,
uses keyed authenticated identity/channel ownership and mounts the unique
hook session only while open. Current error uses existing common keys and
explicit captured retry; a current selection guard precedes channel route
navigation. The35 new search cases include failed/empty distinction,
duplicate submission, draft/scope/identity/channel/open-session races,
current/revoked target navigation, scoped Query cleanup and retained stale
retry/selection callbacks. No exact-message/thread locator is claimed.

## Remaining runtime limits

Search result-to-message/thread locator needs a separate route/server contract.
Backend search authorization and real Docker SUT/browser acceptance are NOT_RUN
in this atom. Join authorization, unread recovery and Jitsi durable/deployment
blockers remain in 2026-09-05-communications-completion-queue.md.
