# Message delete request lifecycle

Status: UNIT_ACCEPTED in the composed2026-09-06 Docker frontend gate.
Independent plan/source review C/I/M0/0/0 after formatter clarification.
20-case oracle frozen; intended Docker RED preserved and all cases now pass.
One frontend behavior following the accepted read-through atom d65da37d and
current SearchDialog work. No new server permission, confirmation flow or
message-cache semantics are introduced.

## Source truth and outcome

system_specification.md documents two-stage delete authorization: fresh
membership/read admits ownership lookup; ordinary members delete their own,
owner/admin/moderator may delete any. The projected can_delete_message flag
is delete-any, not the author's own right. Existing MessageActions visibility
and MessageItem.canDeleteMessage reflect these distinct branches; preserve
them rather than treating broad toolbar visibility as an authorization bug.

Current handleDelete awaits deleteMessage with no caught rejection/pending
guard. One deliberate user action must issue at most one concurrent DELETE;
failure is handled through the existing canonical toastApiError and the user
can explicitly retry. Old-scope completion must not toast/reset a new row or
user's pending state. Existing WebSocket-authoritative success flow stays.

## Exact future ownership

- frontend/src/components/messenger/message-item.tsx: only delete lifecycle,
  required local imports/hooks and MessageActions glue;
- frontend/src/components/messenger/message-actions.tsx: optional deletePending
  prop/defaultfalse and disabled Button plumbing for Delete only;
- frontend/src/i18n/locales/en/messenger.json and ru/messenger.json: only
  messageDelete.error (EN Unable to delete message; RU Не удалось удалить сообщение);
- frontend/tests/messenger/message-delete-lifecycle.test.tsx: new genuine unit;
- root owns this plan and same-session domain/status docs.
No API/hook/store/backend/shared-ui/other message actions/deps/runner edits,
Mail/Calendar or pro-qa-ai access. Aim <=200 changed production lines total;
return BLOCKED for decomposition if this needs broader infrastructure.

## Frozen behavior

Keep existing own/member versus delete-any policy and recheck the current
canDeleteMessage guard immediately before transport. Additionally require a
nonempty current message ID, same message/channel, authenticated captured user
matching live auth, live mounted scope and non-revoked channel. No role-string
inference, stale prop-only permission or unknown-user allowance.

Use one synchronous in-flight token/ref and local pending state for this row.
Repeated click/keyboard activation before rerender must be a no-op, not merely
a disabled visual. While pending, keep the otherwise permitted Delete button
mounted and disabled with type=button; do not disable other actions or hide
the whole toolbar. Respect existing revoke/permission hiding.

Capture user/channel/message plus committed lifecycle generation before
DELETE. Tiny commit/layout activation and invalidating cleanup fence changes,
revoke and unmount; committed A-B-A cannot resurrect an old completion. No
lifecycle ref mutation during render. After await/rejection only a still-current
mounted scope may update its pending state or show error; finally cannot clear
another request token. Revocation releases local UI ownership but is not a
claim that sent HTTP/server work was canceled.

Catch transport rejection and call toastApiError(error,
t('messageDelete.error')) only for current authorized scope. Consume obsolete
rejections silently; no unhandled promise, automatic retry, timers or count
arithmetic. A current failure releases pending so another deliberate click can
retry. On success preserve existing server/WebSocket update semantics: no
optimistic tombstone, new cache invalidation, extra success toast or undo flow.

Root clarification after independent preflight I1: the common formatter
prioritizes API/Error.message (and special rate-limit/request-id handling)
over the passed fallback. This atom supplies a translated fallback but does
NOT guarantee every returned server/transport message is translated. Keep
the common formatter unchanged; no new cross-app error translation policy.

## Test-first acceptance

Render real MessageItem and MessageActions with QueryClient/auth/messenger
state; mock deleteMessage and toastApiError boundaries plus unrelated heavy
children only. Do not mock target callback, canDeleteMessage or own/store
authorization. Existing capability helper fixtures may be reused read-only.

1. Own member and foreign delete-any positive cases: exact DELETE IDs once.
2. Foreign ordinary member has no Delete; missing props/store permission,
   mismatched channel, unknown actor and live revoke send no request.
3. Pending duplicate click/keyboard is one transport; otherwise visible
   Delete remains mounted/disabled and other action controls are unchanged.
4. Current rejection invokes canonical error helper with translated fallback
   once, clears pending,
   explicit second click can succeed; no automatic retry/unhandled failure.
5. Late success/error after user/channel/message A-B-A, unmount or revoke is
   silent and cannot unlock a newer in-flight row request.
6. Existing reply/edit/reaction/pin/bookmark/forward behavior is preserved by
   the full frontend regression; no new global shortcut/confirmation assumed.

Independent Terra/high plan+oracle preflight, Terra/xhigh test-first writer;
root runs only docker compose -f docker-compose.local.yml run --rm frontend-unit
after workload check, existing image2GiB/no-swap, no build/pull/host QA. Source
starts after intended RED, then independent source review and full GREEN.
max2 fix cycles per root cause, then decompose. Root docs-sync and exact commit
followed immediately by push only calls-feature-codex to authorized origin;
remote equality verified. Runtime/browser/server auth acceptance is separate.

## Combined Docker RED (2026-09-06 local)

Canonical command: docker compose -f docker-compose.local.yml run --rm frontend-unit
Existing image sha256:286783701fc3ef3cf25b7960eb14b6eb50e80adbbbb5c0d22e5c0799e68ce218;
outer worktree-master-frontend-unit-run-c1effb4f13de;
engine fe485fa4ac6fb90911537ec75ac1ca6e5c8eab5c7f500b2bc184917f25ee556c,
start2026-09-05T23:10:23.769260921Z. Inspected2GiB/no-swap,6CPU,readonly;
Compose/engineexit1,79s, create/start/die/destroy observed with no OOM event,
container absent afterward. TypeScriptPASS,194files/716tests:
684PASS32FAIL,62.36s,plus3 unhandled rejections.

32 failures are18 already-known search lifecycle,10 new delete lifecycle and
4 new thread capability wiring. All prior666 assertions passed, but this is
NOT a green run: the uncaught DELETE errors make whole-run acceptance fail.
Preserved only eight current report artifacts:
testframework/reports/frontend-unit-delete-wiring-red-20260906/.
No build/pull/install/hostQA. Mail frontend/backend and data services stayed
running; its last builds were completed, no other QA container was listed.
Initial Docker event formatting failed to quote a map key; the read-only
subscription was corrected, without retrying the QA process.

Frozen source/oracle hashes stayed unchanged through the RED:
- delete oracle09A734D7C376C7680B1E390D50CF6A4EBCA0D7BED9FD9D1F6B5ECD082C0930C2;
- parent wiring oracle8403CB9492EFB9D5DE167B54F0E6FEEEBCE157557E273325D9B4BA9BD9E91AEB;
- owner wiring oracleF9E544E838517DB3D6ECA3DCD85CB1BD4ABA6B8A4B23598D54BF24617D3D2BD2;
- search oracle70E4899F92BC1560D57BAC754BF8B27C207159F1C509A8BCF24B21F9CA73D7EA.

Delete ROOT=code: unknown actor and mismatched message/channel are sent;
three clicks send three requests; canonical error helper is never called;
pending control is absent in6ABA cases. Vitest independently catches exactly
transport failed / late revoke completion / late unmount completion from
the new deletion tests. No unhandled-error suppression was added. Two late
error case assertions can pass locally while Vitest flags their derived
handler promise globally; only zero whole-run unhandled errors can accept
GREEN. Oracle settlement witnesses were strengthened before RED, not after
a proven original async-act timing failure. Native browser keyboard NOT_RUN.

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

One matched lifecycle/request token now guards DELETE admission, pending,
failure toast and cleanup; committed user/channel/message changes invalidate
old work. Only Delete is disabled while its own request is pending. Existing
own-member versus delete-any semantics and WS-owned success remain unchanged.
Current errors use toastApiError with messageDelete.error RU/EN fallback;
the shared formatter may still prioritize a provided API error message.
No optimistic removal, success toast, auto retry or permission redesign.

Target source hashes:
- message-item.tsx 9F9B0EC52F9DA93F404E2EF217BBE41940A7C4F7D5F1F872CB48644A9789735A;
- message-actions.tsx F8E882F64B0CC22B2F62871215E44FD7E82ABF04B5C5A640E02D997F982F5FD3.
The frozen20-case oracle SHA above is unchanged. Three old uncaught error
conditions now settle without a whole-run unhandled-error failure.
