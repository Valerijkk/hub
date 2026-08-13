# Phase 146 — exact Spark dispatch index

> **Status:** CONTRACT_READY · CODE_RUNTIME_BLOCKED · TEST_HARNESS_PENDING_REBASE
> **Required runtime:** `gpt-5.3-codex` / `xhigh`, no fallback
> **Contract:** OpenAPI/Mail 0.1.2 + accepted UI spec

The UI uses TanStack Query as server-state authority, TanStack Router/URL as selection
authority, and a local reducer only for the active composer. It must not create a second
Zustand mailbox cache. Every packet owns a small behavior and stops on dirty overlap.

Complete standalone prompts exist for 146-01A, 04C and 04D. For every other row/variant,
assemble the Common envelope plus exactly one section from `DISPATCH-REMAINING.md`.

## Entry gates

1. Phase 143 canonical calendar API is accepted before invitation RSVP UI.
2. Phase 144 public mailbox API is accepted before live mail queries/mutations.
3. Exact Spark runtime is available for every source/test packet.
4. The protected test-rebase owner either merges the new frontend unit harness or hands
   off an approved feature-test config. Current `frontend/vite.config.ts` points into a
   protected legacy testframework path; this phase must not read/change that protected
   suite or claim its own tests executed through it.
5. No new MSW dependency is added by default. Repository-native typed client/fetch fakes
   are used unless founder approves a reviewed dependency/provenance change.

## Route contract

```text
/mail                                  -> redirect /mail/inbox
/mail/:folderKey                       -> folder + thread list, no selected thread
/mail/:folderKey/thread/:threadId      -> folder + thread + reader
/mail/search                           -> search state in URL
/mail/search/thread/:threadId          -> search + reader
```

Calendar remains `/calendar` with additive leaf/detail routes coordinated in its module.
Generated `frontend/src/routeTree.gen.ts` is never manually edited; generation is a
separate accepted Docker/build step.

## Atomic packets

| ID | Single owned behavior | Exact primary files | Oracle / acceptance | Dependency |
|---|---|---|---|---|
| 146-01A | Safe request options and recovery response headers | `frontend/src/api/client.ts`, `frontend/tests/mail/api-client-mail-headers.test.ts` | idempotency/If-Match survive one auth retry; caller cannot override auth/content type; same-origin Location only; ApiError has code/etag/location | runtime + test harness |
| 146-01B | OpenAPI 0.1.2 mailbox types/client/query keys | `frontend/src/api/mail.ts`, `frontend/src/modules/mail/mail-query-keys.ts`, focused unit | exact paths/body/result; opaque cursor; submission GET; no account/org input | 146-01A + Phase 144 |
| 146-02A | Mail route files and shell layout only | `frontend/src/routes/_authenticated/mail*.tsx` as router convention requires, `frontend/src/modules/mail/mail-shell.tsx`, focused route unit | exact redirect/deep links/back-forward; URL truth; no routeTree manual edit | 146-01B |
| 146-02B | Responsive folder pane + counts | `mail-folder-pane.tsx`, `mail-folder-tree.ts`, focused unit | system/custom hierarchy; keyboard/focus; mobile sheet; loading/error/empty | 146-02A |
| 146-02C | Virtualized thread list + selection | `mail-thread-list.tsx`, `mail-thread-row.tsx`, focused unit | cursor pages, stable keys, unread/flag/attachment, no duplicate cache | 146-02A |
| 146-03A | Isolated safe message reader boundary | `mail-message-reader.tsx`, `mail-safe-content.ts`, focused security unit | sanitized server HTML only; sandbox/CSP boundary; no direct remote fetch; plain fallback | 146-02A + Phase 144 |
| 146-03B | Attachment states then calendar invite card as two sequential variants | `mail-attachments.tsx`/unit, then `mail-invite-card.tsx`/unit | pending/rejected/clean; then refetch-on-409 and occurrence/series | 146-03A + Phase 143 |
| 146-04A | Local composer reducer then Plate wrapper as two sequential variants | `mail-composer-state.ts`/unit, then `mail-composer-editor.tsx`/unit | deterministic dirty/revision state; then rich/plain editor boundary | 146-01B |
| 146-04B | Address chips then composition policy as two sequential variants | `mail-address-field.tsx`/unit, then `mail-compose-policy.ts`/unit | user-first suggestions/aliases; then CC/BCC/priority/signature/quoting/bounds | 146-04A |
| 146-04C | Autosave conflict/unknown recovery | `use-mail-draft.ts`, focused hook unit | debounce cancellation; one in-flight save; If-Match; unknown refetch/reconcile; no blind overwrite | 146-04A/B + Phase 144 draft API |
| 146-04D | Send receipt and unknown polling/retry | `use-mail-submit.ts`, focused hook unit | stable key per intent; duplicate clicks one request; unknown polls receipt; retry same key; no second send | 146-04C + Phase 144 submission API |
| 146-05 | Bulk actions/toolbar then search as sequential variants | action hook/toolbar units, then `mail-search.tsx`/unit | optimistic rollback/partial failures/bounds; then URL filters/cursors | 146-02B/C |
| 146-06A-Settings | User mailbox settings page and exact settings-route entry | `frontend/src/modules/mail/mail-settings-page.tsx`, `frontend/src/modules/mail/mail-user-settings.tsx`, `frontend/src/routes/_authenticated/settings/mail.tsx`, `frontend/src/routes/_authenticated/settings.tsx`, `frontend/tests/mail/mail-user-settings.test.tsx`, `frontend/tests/mail/mail-settings-route.test.tsx` | defaults/options/errors; one `/settings/mail` entry; legacy settings items preserved | 146-01B |
| 146-06A-Signatures | Signature CRUD and exact mail-settings page integration | `frontend/src/modules/mail/mail-signatures.tsx`, `frontend/src/modules/mail/mail-settings-page.tsx`, `frontend/tests/mail/mail-signatures.test.tsx` | sanitized bounded CRUD/default; existing user settings remain mounted | 146-06A-Settings |
| 146-06B-API | Separate mailbox admin client | `frontend/src/api/admin-mailbox.ts`, `frontend/tests/mail/admin-mailbox-api.test.ts` | new policy/domain/health shapes cannot alias legacy flat Resend client | Phase 144 admin API |
| 146-06B-UI | Mailbox admin panel and exact `/admin/mail` tab integration | `frontend/src/components/admin/mailbox-admin-settings.tsx`, `frontend/src/routes/_authenticated/admin/mail.tsx`, `frontend/tests/mail/mailbox-admin-settings.test.tsx` | distinct Alerts/Mailbox tabs; redacted health/domains/policy; legacy panel untouched | 146-06B-API |
| 146-07 | WebSocket/query invalidation mapping | `use-mail-realtime.ts`, focused unit | invalidates smallest keys; duplicate event safe; no payload body cache | 146-01B/02 |
| 146-08-I18n | Register Mail namespace and exact EN/RU navigation/user strings | `frontend/src/i18n/index.ts`, `frontend/src/i18n/locales/en/mail.json`, `frontend/src/i18n/locales/ru/mail.json`, `frontend/src/i18n/locales/en/common.json`, `frontend/src/i18n/locales/ru/common.json`, `frontend/tests/mail/mail-i18n.test.ts` | identical keys; escaped interpolation; no hard-coded Mail UI strings | 146-02A |
| 146-08-Nav | Desktop/mobile Mail navigation and mailbox unread selector | `frontend/src/components/layout/sidebar-nav.tsx`, `frontend/src/components/layout/sidebar-link.tsx`, `frontend/src/components/mobile-nav.tsx`, `frontend/src/modules/mail/use-mail-unread-count.ts`, `frontend/tests/mail/mail-navigation.test.tsx` | Mail between Messenger/Calendar; query-derived unread; no Messenger count reuse | 146-02A/07 + 146-08-I18n |
| 146-09 | Route generation, Docker-only build/unit/a11y verification | generated route tree + task report only via accepted launcher | current SHA evidence; keyboard/resize/reduced motion; exact cleanup | all + harness/launcher |

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
- Draft bodies live in local composer state and authorized API responses, not TanStack
  mutation metadata persisted beyond need, Zustand, localStorage or logs.
- Autosave and send are separate state machines. Send waits for or reconciles the current
  draft revision; unknown send polls/retries the same receipt key.
- Accessibility includes logical focus on pane changes, keyboard list/reader navigation,
  visible focus, semantic labels/status announcements and mobile escape/back behavior.
- No manual edit of generated route tree and no host test/build/lint/validator.

## Gate status

| Gate | Status |
|---|---|
| UI design contract | READY |
| Mail/Calendar backend implementation | NOT_IMPLEMENTED |
| Exact Spark runtime | UNAVAILABLE |
| Frontend unit harness | PENDING_PROTECTED_REBASE |
| Accepted launcher | NOT_READY |
| Build/unit/a11y/runtime | NOT_RUN |
