# SearchDialog scoped request lifecycle

Status: ROOT_CONTRACT_FROZEN; oracle and implementation NOT_RUN.
Baseline calls-feature-codex f1691e34; concurrent read-receipts changes are
task-owned and must be preserved. This is one search interaction atom, not
message deep-link navigation or full Messenger acceptance.

## Outcome and boundaries

Explicit search shows only the result of the current submitted query in the
current authenticated dialog session. Successful empty is distinct from
failure. Old success/error cannot reappear after draft/scope/channel/user
changes or closing and reopening. Existing channel-only navigation remains.

Ownership for the future writer:
- frontend/src/components/messenger/search-dialog.tsx only production file;
- frontend/tests/messenger/search-dialog.lifecycle.test.tsx new genuine unit;
- root owns this plan and relevant same-session domain/status docs.
Keep existing a11y-names-p2.test.tsx unchanged. No API, hook, store, shared UI,
locale, backend, routes, dependencies, test launcher or Mail/Calendar changes.
Aim <=200-300 changed production lines; return BLOCKED for decomposition if
the atom requires a new framework, shared cache or broader feature.

## Frozen design

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

## Deferred, recorded rather than silently included

Search result-to-message/thread locator needs a separate route/server contract.
Backend search authorization and real Docker SUT/browser acceptance are NOT_RUN
in this atom. Join authorization, unread recovery and Jitsi durable/deployment
blockers remain in 2026-09-05-communications-completion-queue.md.
