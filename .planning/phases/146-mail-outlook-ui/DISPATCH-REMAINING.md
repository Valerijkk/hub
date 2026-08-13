# Phase 146 — executable packet contracts not split into standalone files

> **Assembly rule:** Common envelope + exactly one packet/variant. Never combine variants.
> **Runtime:** `gpt-5.3-codex` / `xhigh`, fallback none.
> **Current blockers:** backend NOT_IMPLEMENTED, exact runtime unavailable, frontend
> feature-unit harness pending protected rebase, launcher NOT_READY.

Standalone complete prompts exist for 146-01A, 146-04C and 146-04D. This file completes
every other row in `DISPATCH-INDEX.md`.

## Common execution envelope

- Atomic TypeScript/React role; external genuine unit first, then minimal source. Root
  supplies clean HEAD and accepted upstream API/type surface. Dirty overlap, missing route
  generation/harness, changed contract or actual model mismatch => `BLOCKED` before write.
- Modify only chosen Allowed files. No commit/push/PR/dependency install/generated manual
  edit. Never read/change protected testframework/test-rebase/pro-qa-ai paths.
- TanStack Query owns server state; URL owns folder/thread/search; local reducer owns only
  composer draft. No duplicate Zustand/localStorage/sessionStorage mailbox cache.
- No raw HTML/direct remote URL/body/recipients/idempotency key in logs, query keys, error
  objects or persisted client state. Existing Messenger/Calendar/Resend UI stays compatible.
- Tests use repository-native Vitest/RTL/client fakes; no MSW unless separately approved.
  Host test/build/lint/validator forbidden; direct Docker not fallback; launcher/harness
  absent => `NOT_RUN`.
- Accessibility: semantic name/state, visible focus, editable-target shortcut guard, polite
  live status, 44px mobile targets. EN/RU for user text.
- depth/fan-out 0, retries 2, 45m (60 for UI). Russian response: files/symbols/oracles/
  verification/a11y-security/model provenance/SHA.

## 146-01B — typed mail API and query keys

- Dependencies: 01A + accepted OpenAPI 0.1.2/Phase 144. Allowed only
  `frontend/src/api/mail.ts`, `frontend/src/modules/mail/mail-query-keys.ts`,
  `frontend/tests/mail/mail-api.test.ts`.
- Behavior: exact typed functions for all public mailbox paths, `URLSearchParams`, opaque
  cursors, draft key/If-Match options, submission GET, 204 handling, remote-image token
  route, no account/org/engine id inputs. Query-key factories exclude body/recipient/key.
- Oracles: exact method/path/query/body for each group; bounds represented; same-key send;
  nullable/schema states; response types flags/folder ids/draft status; compile negative
  owner fields; stable small query keys.

## 146-02A — mail routes and shell

- Dependency: 01B. Allowed only:
  `frontend/src/routes/_authenticated/mail/index.tsx`,
  `mail/$folderKey.tsx`, `mail/$folderKey/thread/$threadId.tsx`, `mail/search.tsx`,
  `mail/search/thread/$threadId.tsx`,
  `frontend/src/modules/mail/mail-shell.tsx`,
  `frontend/tests/mail/mail-routes.test.tsx`.
- Behavior: `/mail` redirect inbox; exact deep links; validated bounded search fields;
  shell chooses desktop/medium/mobile pane visibility from route + responsive primitives;
  back/forward closes reader naturally; no server fetch duplicated by route components.
- Oracles: each URL/render/invalid param/encoded id/back-forward; desktop 3 panes, medium 2,
  mobile one; focus handoff; no manual `routeTree.gen.ts` edit. Generation is packet 09.

## 146-02B — folder pane/tree

- Dependencies: 02A/01B. Allowed only `mail-folder-pane.tsx`, `mail-folder-tree.ts`,
  `frontend/tests/mail/mail-folder-pane.test.tsx`.
- Behavior/oracles: role/custom hierarchy and counts; active folder from URL; loading/error/
  empty; keyboard tree navigation/roving focus; system role controls restricted; mobile
  sheet focus return; unread not color-only; no folder server cache copy.

## 146-02C — virtualized thread list

- Dependencies: 02A/01B. Allowed only `mail-thread-list.tsx`, `mail-thread-row.tsx`,
  `frontend/tests/mail/mail-thread-list.test.tsx`.
- Behavior/oracles: infinite cursor pages, TanStack virtual stable keys, URL selection,
  unread/flag/attachment/participants/date semantics; loading/error/empty/end; no duplicate
  row/page; j/k/x/Enter outside editable target; focus and mobile navigation.

## 146-03A — isolated safe reader

- Dependencies: 02A/01B + backend safe-body contract. Allowed only
  `mail-message-reader.tsx`, `mail-safe-content.ts`,
  `frontend/tests/mail/mail-safe-content.test.tsx`.
- Behavior: plain body fallback; sanitized HTML only in one sandboxed iframe/srcDoc helper
  with injected restrictive CSP and no scripts/forms/top navigation; initial image network
  disabled. Outer DOM never uses raw `dangerouslySetInnerHTML`; arbitrary URL is never
  accepted. Links use safe allowlist and explicit new-window isolation.
- Oracles: hostile script/event/SVG/CSS/form/meta/base/iframe/object and remote image cannot
  execute/request; plain text shown on invalid/missing HTML; reader landmarks/focus; body
  absent from logs/query keys. Browser security runtime remains later E2E.

## 146-03B-Attachments — attachment UI

- Dependency: 03A/01B. Allowed only `mail-attachments.tsx`,
  `frontend/tests/mail/mail-attachments.test.tsx`.
- Behavior/oracles: pending/scanning/clean/rejected/error states; only clean owner download
  via attachment endpoint; filename/type/size accessible; no storage/object/arbitrary URL;
  pending/rejected prevents send signal; partial/error/live status.

## 146-03B-Invite — invitation card

- Dependencies: 03A + accepted Phase 143 + 01A/B. Allowed only
  `mail-invite-card.tsx`, `frontend/tests/mail/mail-invite-card.test.tsx`.
- Behavior/oracles: REQUEST/CANCEL/REPLY state; series vs occurrence endpoint; one RSVP
  intent; 409 code/ETag and strictly canonical relative Location -> refetch before retry;
  absolute/cross-origin discarded; committed RSVP not rolled back for pending/failed mail
  reply; occurrence cancel display; live/a11y status and timezone text.

## 146-04A-State — composer reducer

- Dependency: 01B. Allowed only `mail-composer-state.ts`,
  `frontend/tests/mail/mail-composer-state.test.ts`.
- Behavior/oracles: pure states clean/dirty/saving/conflict/unknown/submitting/submitted/
  failed; snapshot ids and server revision; newer edits survive older success; explicit
  conflict choice/new send intent; no timer/network/random/global cache; no body in debug
  serialization.

## 146-04A-Editor — Plate wrapper

- Dependency: accepted 04A-State. Allowed only `mail-composer-editor.tsx`,
  `frontend/tests/mail/mail-composer-editor.test.tsx`.
- Behavior/oracles: reuse existing Plate facade without forking internals; controlled value
  -> reducer; deterministic sanitized rich serialization and plain fallback; paste bounds;
  keyboard/accessibility; no network/autosave/send; lazy loading state.

## 146-04B-Address — address chips and suggestions

- Dependencies: 04A-State/01B. Allowed only `mail-address-field.tsx`,
  `frontend/tests/mail/mail-address-field.test.tsx`.
- Behavior/oracles: active Hub users before recent external, normalized dedupe, keyboard
  combobox/chips, invalid/duplicate/recipient cap, To/CC/BCC semantics; selected suggestion
  only stores address value; no cross-org result/log.

## 146-04B-Policy — compose policy/quoting

- Dependencies: accepted 04A/04B-Address. Allowed only `mail-compose-policy.ts`,
  `frontend/tests/mail/mail-compose-policy.test.ts`.
- Behavior/oracles: authorized identity/alias, signature insertion/change, reply/reply-all/
  forward recipients/headers/quoted source, CC/BCC/priority, attachment ids and bounds;
  deterministic pure transforms; BCC received reconstruction forbidden; no raw HTML trust.

## 146-05-Actions — optimistic bulk action hook

- Dependencies: 02B/C/01B. Allowed only `use-mail-actions.ts`,
  `frontend/tests/mail/use-mail-actions.test.tsx`.
- Behavior/oracles: bounded selected ids, exact mutation, snapshot/optimistic patch for
  reversible fields, per-message partial failure rollback, definitive refetch on engine
  ambiguity, invalidate smallest query keys; no body/full cache duplication.

## 146-05-Toolbar — bulk toolbar

- Dependency: accepted 05-Actions. Allowed only `mail-toolbar.tsx`,
  `frontend/tests/mail/mail-toolbar.test.tsx`.
- Behavior/oracles: enabled actions by selection/folder, select page/all-visible semantics,
  progress/partial result/live announcement, keyboard/focus/mobile targets; no API call
  beyond hook; destructive hard delete not exposed unless contract allows.

## 146-05-Search — URL search

- Dependencies: 02A/C/01B. Allowed only `mail-search.tsx`,
  `frontend/tests/mail/mail-search.test.tsx`.
- Behavior/oracles: validated `q/from/to/folder_id/label_id/unread/has_attachments/date_*`,
  debounced URL replace, cursor query, share/back-forward, loading/error/no results; no raw
  query in analytics/log beyond existing privacy policy; selected thread deep link.

## 146-06A-Settings — user mailbox settings

- Dependency: 01B. Allowed only
  `frontend/src/modules/mail/mail-settings-page.tsx`,
  `frontend/src/modules/mail/mail-user-settings.tsx`,
  `frontend/src/routes/_authenticated/settings/mail.tsx`,
  `frontend/src/routes/_authenticated/settings.tsx`,
  `frontend/tests/mail/mail-user-settings.test.tsx`, and
  `frontend/tests/mail/mail-settings-route.test.tsx`.
- Behavior/oracles: active identities, required default identity, nullable signature,
  remote image blocked/trusted and compose format; server validation/errors; cache
  invalidation; loading/non-active mailbox; no engine credential. The new route is exactly
  `/settings/mail`; the parent settings navigation adds exactly one Mail item and preserves
  all current entries/order otherwise. `mail-settings-page.tsx` initially mounts only
  `MailUserSettings` and is the sole integration seam for the next packet.

## 146-06A-Signatures — signature CRUD

- Dependency: 06A-Settings. Allowed only
  `frontend/src/modules/mail/mail-signatures.tsx`,
  `frontend/src/modules/mail/mail-settings-page.tsx`, and
  `frontend/tests/mail/mail-signatures.test.tsx`.
- Behavior/oracles: list/create/update/delete/default; rich editor input + server-normalized
  plain response; one default; bounded name/body; safe preview; conflict/error/focus; no
  direct raw HTML render outside safe boundary. Integrate the component only in the
  already-created `mail-settings-page.tsx`; do not edit the route or parent settings nav.

## 146-06B-API — separate mailbox admin client

- Dependency: Phase 144 admin API + 01A. Allowed only
  `frontend/src/api/admin-mailbox.ts`,
  `frontend/tests/mail/admin-mailbox-api.test.ts`.
- Behavior/oracles: exact mailbox-settings/domain/health paths/types/methods, redacted
  health, no secret fields; no changes/imports to legacy `admin-mail.ts`; compile distinction
  between flat Resend and mailbox policy; non-arbitrary headers.

## 146-06B-UI — mailbox admin UI

- Dependency: accepted 06B-API. Allowed only
  `frontend/src/components/admin/mailbox-admin-settings.tsx`,
  `frontend/src/routes/_authenticated/admin/mail.tsx`, and
  `frontend/tests/mail/mailbox-admin-settings.test.tsx`.
- Behavior/oracles: policy limits/external toggle, domain verification/DNS public records,
  redacted health/not_configured/degraded; admin guard/loading/error/confirm/focus; legacy
  Resend settings remain separate; broadcast UI is not invoked/expanded (removal requires
  founder decision/own packet). The existing `/admin/mail` route becomes an accessible
  two-tab page: `Alerts` continues to mount the unchanged `MailSettingsPanel`, while
  `Mailbox` mounts the new panel. Do not edit `mail-settings.tsx` or
  `admin/dashboard.tsx` in this packet.

## 146-07 — realtime query invalidation

- Dependencies: 01B/02/Phase 144 events. Allowed only `use-mail-realtime.ts`,
  `frontend/tests/mail/use-mail-realtime.test.tsx`.
- Behavior/oracles: validate unknown payload version/ids; duplicate event safe; invalidate
  folder/list/search/selected thread minimal prefixes; no optimistic body/state from event;
  subscribe/unsubscribe lifecycle; cross-user event ignored; no Zustand mailbox cache.

## 146-08-I18n — Mail namespace and navigation strings

- Dependency: 02A. Allowed only
  `frontend/src/i18n/index.ts`,
  `frontend/src/i18n/locales/en/mail.json`,
  `frontend/src/i18n/locales/ru/mail.json`,
  `frontend/src/i18n/locales/en/common.json`,
  `frontend/src/i18n/locales/ru/common.json`, and
  `frontend/tests/mail/mail-i18n.test.ts`.
- Behavior/oracles: register the lazily loaded `mail` namespace in the existing loader;
  add the single `common.nav.mail` shell label; EN/RU key sets are identical, interpolation
  remains escaped at render call sites, and all Mail user-visible strings used by accepted
  packets have translations. Do not edit navigation components in this packet.

## 146-08-Nav — desktop/mobile Mail navigation

- Dependencies: 02A/07 plus accepted 08-I18n. Allowed only
  `frontend/src/components/layout/sidebar-nav.tsx`,
  `frontend/src/components/layout/sidebar-link.tsx`,
  `frontend/src/components/mobile-nav.tsx`,
  `frontend/src/modules/mail/use-mail-unread-count.ts`, and
  `frontend/tests/mail/mail-navigation.test.tsx`.
- Behavior/oracles: add Mail between Messenger and Calendar on the desktop sidebar; add
  `/mail` icon color; use `common.nav.mail`; derive a bounded unread total through the
  mailbox TanStack Query surface and the dedicated hook; loading/error shows no false zero
  announcement. Mobile bottom navigation must be exactly, in order: `/dashboard`,
  `/messenger`, `/mail`, `/calendar`, `More`. The `More` panel must be exactly, in order:
  `/projects` (Tracker), `/calls`, `/settings/profile`, `/trash`. Snapshot/assert both full
  ordered sets, labels, destinations and Mail unread badge; no sixth bottom item or other
  reorder is allowed. Existing desktop routes/order remain unchanged except the named Mail
  insertion. Never read `useMessengerStore` for Mail or duplicate mailbox counts in
  Zustand/local storage.

## 146-09 — generated route/build/unit/a11y gate

- Dependencies: all UI packets + approved harness/launcher. Allowed: generated
  `frontend/src/routeTree.gen.ts` only via normal router generator and task-owned evidence;
  no manual edit/product refactor.
- Behavior/oracles: accepted launcher runs route generation, type/build, feature genuine
  units and scoped accessibility/browser checks with exact SHA/versions/limits/timeout and
  cleanup. Inspect generated diff contains only new routes. Keyboard/resize/mobile/reduced-
  motion/reader network isolation/unknown send cases recorded. Missing gate => NOT_RUN.
