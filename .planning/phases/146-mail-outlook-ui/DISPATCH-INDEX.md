# Phase 146 — exact Terra dispatch index

> **Status:** CONTRACT_0_1_8_ROOT_FROZEN_REVIEW_PENDING · TERRA_CODE_ROUTE_AUTHORIZED · BACKEND_DEPENDENCIES_AND_TEST_HARNESS_PENDING
> **Required runtime:** `gpt-5.6-terra` / `xhigh`, no fallback
> **Contract:** OpenAPI/Mail 0.1.8 root candidate + accepted UI spec; dependent UI code waits for independent contract and backend acceptance

The UI uses TanStack Query as server-state authority, TanStack Router/URL as selection
authority, and a local reducer only for the active composer. It must not create a second
Zustand mailbox cache. Every packet owns a small behavior and stops on dirty overlap.

Complete standalone prompts exist for 146-01A, 04C and 04D. For every other row/variant,
assemble the Common envelope plus exactly one section from `DISPATCH-REMAINING.md`.

## Entry gates

1. Phase 143 canonical calendar API is accepted before invitation RSVP UI.
2. Phase 144 public mailbox API is accepted before live mail queries/mutations.
3. Exact founder-authorized Terra runtime is available for every source/test packet.
4. **AVAILABLE_FROM_MASTER_INTEGRATION:** the frontend unit harness arrived at
   `f21e3f77` and was integrated into this branch by merge `6dc6a066`. Current
   `frontend/tests/**` is the genuine-unit tree and `frontend/vite.config.ts` configures
   Vitest to include `tests/**/*.test.{ts,tsx}` with the repository setup file. This
   removes the protected-rebase harness blocker; it does not authorize a packet: 146-01A
   retains its exact test-first two-file manifest and the current founder no-test
   constraint leaves all frontend source/test packets unselected. This phase must not
   read/change protected testframework or test-rebase suites, or claim unrun tests.
5. No new MSW dependency is added by default. Repository-native typed client/fetch fakes
   are used unless founder approves a reviewed dependency/provenance change.

## Route contract

```text
/mail                                  -> redirect /mail/inbox
/mail/:folderKey                       -> folder + thread list, no selected thread
/mail/:folderKey/thread/:threadId      -> folder + thread + reader
/mail/search                           -> search state in URL
/mail/search/thread/:threadId          -> search + reader
/mail/drafts/:draftId                  -> owner-scoped saved draft in ordinary composer
/mail/message/:messageId               -> owner-scoped direct message reader
```

Calendar remains `/calendar` with additive leaf/detail routes coordinated in its module.
Generated `frontend/src/routeTree.gen.ts` is never manually edited; generation is a
separate accepted Docker/build step.

## Atomic packets

| ID | Single owned behavior | Exact primary files | Oracle / acceptance | Dependency |
|---|---|---|---|---|
| 146-01A | Safe request options and recovery response headers | `frontend/src/api/client.ts`, `frontend/tests/mail/api-client-mail-headers.test.ts` | idempotency/If-Match survive one auth retry; caller cannot override auth/content type; same-origin Location only; ApiError has code/etag/location/retry-after | runtime + test harness |
| 146-01B | OpenAPI 0.1.8 mailbox types/client/query keys | `frontend/src/api/mail.ts`, `frontend/src/modules/mail/mail-query-keys.ts`, focused unit | exact bounded paths/body/result; stable create/bulk key; account/projection state; submission GET; no account/org input | 146-01A + Phase 144 |
| 146-02A | Mail route files and shell layout only | `frontend/src/routes/_authenticated/mail*.tsx` as router convention requires, `frontend/src/modules/mail/mail-shell.tsx`, focused route unit | exact redirect/deep links including owned draft/message resources; back-forward; URL truth; no routeTree manual edit | 146-01B |
| 146-02B | Responsive folder pane + counts | `mail-folder-pane.tsx`, `mail-folder-tree.ts`, focused unit | system/custom hierarchy; keyboard/focus; mobile sheet; loading/error/empty | 146-02A |
| 146-02C | Virtualized thread list + selection | `mail-thread-list.tsx`, `mail-thread-row.tsx`, focused unit | cursor pages, stable keys, unread/flag/attachment, no duplicate cache | 146-02A |
| 146-02D | Projection freshness derivation and owner-visible status banner | `mail-projection-status.ts`, `mail-projection-banner.tsx`, narrow `mail-shell.tsx`, focused unit | conservative mixed-page state; stale/as-of/rebuilding explicit; retry only invalidates queries; accessible | 146-02A/B/C |
| 146-02E | Safe mailbox account/provisioning status | `mail-account-status.tsx`, narrow `mail-shell.tsx`, focused unit | read-only state; not-configured/provisioning/error actions; no engine/reservation leak | 146-02A |
| 146-02F | Folder management then label management as sequential variants | `mail-folder-manager.tsx`/unit, then `mail-label-manager.tsx`/unit | versioned pending/recovery folder CRUD depth/cap/conflict; then versioned label CRUD/color/in-use recovery; stable create keys | 146-02B + Phase 144 |
| 146-03A | Isolated safe message reader boundary | `mail-message-reader.tsx`, `mail-safe-content.ts`, focused security unit | sanitized server HTML only; sandbox/CSP boundary; no direct remote fetch; plain fallback | 146-02A + Phase 144 |
| 146-03C | Explicit remote-image consent/action | `mail-remote-images.tsx`, narrow reader integration, focused unit | blocked default; token-only same-origin fetch after action/trusted policy; original URL absent | 146-03A |
| 146-03B | Attachment states then calendar invite card as two sequential variants | `mail-attachments.tsx`/unit, then `mail-invite-card.tsx`/unit | same-object presign replay + pending/rejected/clean; then refetch-on-409 and occurrence/series | 146-03A + Phase 143 |
| 146-04A | Local composer reducer then Plate wrapper as two sequential variants | `mail-composer-state.ts`/unit, then `mail-composer-editor.tsx`/unit | deterministic dirty/revision state; then rich/plain editor boundary | 146-01B |
| 146-04B | Address chips then composition policy as two sequential variants | `mail-address-field.tsx`/unit, then `mail-compose-policy.ts`/unit | user-first suggestions/aliases; then CC/BCC/priority/signature/quoting/bounds | 146-04A |
| 146-04C | Autosave conflict/unknown recovery | `use-mail-draft.ts`, focused hook unit | debounce cancellation; one in-flight save; If-Match; unknown refetch/reconcile; no blind overwrite | 146-04A/B + Phase 144 draft API |
| 146-04D | Send receipt and unknown polling/retry | `use-mail-submit.ts`, focused hook unit | stable key per intent; duplicate clicks one request; unknown polls receipt; retry same key; no second send | 146-04C + Phase 144 submission API |
| 146-05 | Bulk actions/toolbar then search as sequential variants | action hook/toolbar units, then `mail-search.tsx`/unit | typed message/thread targets + stable key; training outcome/partial rollback; then URL filters/cursors/body+filename scope | 146-02B/C |
| 146-06A-Settings | User mailbox settings page and exact settings-route entry | `frontend/src/modules/mail/mail-settings-page.tsx`, `frontend/src/modules/mail/mail-user-settings.tsx`, `frontend/src/routes/_authenticated/settings/mail.tsx`, `frontend/src/routes/_authenticated/settings.tsx`, `frontend/tests/mail/mail-user-settings.test.tsx`, `frontend/tests/mail/mail-settings-route.test.tsx` | versioned defaults/options/errors; one `/settings/mail` entry; legacy settings items preserved | 146-01B |
| 146-06A-Signatures | Signature CRUD and exact mail-settings page integration | `frontend/src/modules/mail/mail-signatures.tsx`, `frontend/src/modules/mail/mail-settings-page.tsx`, `frontend/tests/mail/mail-signatures.test.tsx` | sanitized bounded versioned CRUD/default; existing user settings remain mounted | 146-06A-Settings |
| 146-06A-Rules | Typed rule CRUD and deployment state | `frontend/src/modules/mail/mail-rules.tsx`, narrow settings-page integration, focused unit | no raw Sieve; versioned update/delete; pending/applying/ready/unknown/error; stable create key; no false installed | 146-06A-Settings + Phase 144 |
| 146-06B-API | Separate mailbox admin client | `frontend/src/api/admin-mailbox.ts`, `frontend/tests/mail/admin-mailbox-api.test.ts` | versioned policy/domain/health shapes cannot alias legacy flat Resend client | Phase 144 admin API |
| 146-06B-UI | Mailbox admin panel and exact `/admin/mail` tab integration | `frontend/src/components/admin/mailbox-admin-settings.tsx`, `frontend/src/routes/_authenticated/admin/mail.tsx`, `frontend/tests/mail/mailbox-admin-settings.test.tsx` | distinct Alerts/Mailbox tabs; redacted health/domains/policy; legacy panel untouched | 146-06B-API |
| 146-06B-Aliases | Per-user alias admin UI | `frontend/src/components/admin/mailbox-aliases.tsx`, narrow mailbox admin integration, focused unit | same-org user picker; list/create/delete alias; stable create key; primary/group protected | 146-06B-API |
| 146-07 | WebSocket/query invalidation mapping | `use-mail-realtime.ts`, focused unit | invalidates smallest keys; duplicate event safe; no payload body cache | 146-01B/02 |
| 146-08-I18n | Register Mail namespace and exact EN/RU navigation/user strings | `frontend/src/i18n/index.ts`, `frontend/src/i18n/locales/en/mail.json`, `frontend/src/i18n/locales/ru/mail.json`, `frontend/src/i18n/locales/en/common.json`, `frontend/src/i18n/locales/ru/common.json`, `frontend/tests/mail/mail-i18n.test.ts` | identical keys; escaped interpolation; no hard-coded Mail UI strings | 146-02A |
| 146-08-Nav | Desktop/mobile Mail navigation and mailbox unread selector | `frontend/src/components/layout/sidebar-nav.tsx`, `frontend/src/components/layout/sidebar-link.tsx`, `frontend/src/components/mobile-nav.tsx`, `frontend/src/modules/mail/use-mail-unread-count.ts`, `frontend/tests/mail/mail-navigation.test.tsx` | Mail between Messenger/Calendar; query-derived unread; no Messenger count reuse | 146-02A/07 + 146-08-I18n |
| 146-09 | Route generation, Docker-only build/unit/a11y verification | generated route tree + task report only via accepted launcher | current SHA evidence; keyboard/resize/reduced motion/projection states; exact cleanup | all + harness/launcher |

## Cross-packet hard rules

- `frontend/src/api/client.ts` exposes named request options; callers cannot pass arbitrary
  Authorization, Cookie, Host or Content-Type overrides.
- A 401 refresh retry preserves the exact idempotency key and `If-Match`; it never creates
  a new send intent key.
- RSVP conflict UI accepts `Location` only when it is a relative same-origin canonical
  `/api/calendar/v2/events/<uuid>` path, then refetches before offering retry.
- Server HTML is treated as already-sanitized but still rendered inside the frozen reader
  isolation boundary. `dangerouslySetInnerHTML` outside that boundary is forbidden.
- Remote images load only through the server-issued message/token route after user action
  or trusted-sender policy; original URLs never reach browser fetch/DOM.
- Every create and bulk request gets one stable random idempotency key per user intent,
  retained across auth retry/unknown response and regenerated only for a new explicit intent.
- Rich compose/signature requests send explicit format and rich or plain source; they do
  not fabricate a second client-authoritative plain body for server-rich content.
- Draft bodies live in local composer state and authorized API responses, not TanStack
  mutation metadata persisted beyond need, Zustand, localStorage or logs.
- Autosave and send are separate state machines. Send waits for or reconciles the current
  draft revision; unknown send polls/retries the same receipt key.
- `MAIL_ACCOUNT_BUSY` is shown as a bounded retryable state and honors validated
  `Retry-After`; it is never conflated with an unknown command outcome or auto-replayed
  without the same safe request/idempotency context.
- Folder/thread/search data is labelled fresh only when every visible source is `ready`.
  Bootstrapping, degraded and rebuilding remain owner-filtered but visibly stale/in-progress;
  client retry invalidates bounded Mail queries and never starts or simulates the worker.
- Accessibility includes logical focus on pane changes, keyboard list/reader navigation,
  visible focus, semantic labels/status announcements and mobile escape/back behavior.
- No manual edit of generated route tree and no host test/build/lint/validator.

## Gate status

| Gate | Status |
|---|---|
| UI design contract | READY |
| Mail/Calendar backend implementation | NOT_IMPLEMENTED |
| Exact Terra runtime | AUTHORIZED_BY_FOUNDER / first actual packet pending |
| Frontend unit harness | AVAILABLE_FROM_MASTER_INTEGRATION (`f21e3f77`; merge `6dc6a066`); 146-01A remains NOT_SELECTED under founder no-test constraint |
| Accepted launcher | NOT_READY |
| Build/unit/a11y/runtime | NOT_RUN |
