---
status: fixing
trigger: "EMERGENCY: Demo in 10 hours. The entire app is crashing. Fix ALL bugs."
created: 2026-03-21T00:00:00Z
updated: 2026-03-21T00:00:00Z
---

## Current Focus

hypothesis: Multiple bugs identified across 7 files
test: Fix all and verify build
expecting: App loads without crashes
next_action: Apply all fixes

## Symptoms

expected: App works - messenger, dashboard, calls, notifications all functional
actual: React error #185, WS failures, undefined call IDs, infinite redirects, missing channels, notification routing broken
errors: React error #185, WebSocket closed before established, /api/calls/undefined/join 400
reproduction: Open messenger, click notification, create channel, join call
started: Current state of codebase

## Evidence

- timestamp: 2026-03-21
  checked: All files in files_to_read
  found: 9 distinct bugs across notification-panel, channel-list, call-store, Dockerfile, profile page, dashboard route
  implication: All fixable with targeted changes

## Bugs Found

1. **notification-panel.tsx line 121**: `case "call"` returns `/dashboard` - should navigate to channel, but currently navigates to dashboard which starts call cycle. The `entity_type: "call"` notification uses `entity_id` as call ID and navigates to `/dashboard` -- but user expects navigation to the channel. Since call notifications likely have `channel_id` in metadata, the route should go to the channel. However the bigger issue is that the notification click handler for `call_started` type navigates to `/dashboard`, and there's no explicit call-starting on dashboard -- so bug #9 is likely from a different mechanism. Actually re-reading: the notification `resolveNotificationRoute` for `case "call"` returns `"/dashboard"` -- this IS the bug. It should NOT navigate to dashboard for call notifications. Should return null or navigate to channel if available.

2. **channel-list.tsx**: `createChannel` and `CreateChannelDialogContent` don't invalidate the channels query after creation -- new channel doesn't appear in list (bug #7).

3. **Dockerfile nginx**: WebSocket location `/ws` doesn't have `proxy_read_timeout` and needs proper headers -- WS connection drops (bug #2).

4. **profile.tsx**: Custom status has no emoji picker, only text input (bug #6 -- "only 30+ statuses, can't set custom"). The status input only allows text.

5. **call-store.ts / CallButton.tsx**: When `existingCall` is null and user clicks, `startCall(channelId)` is called which calls `createCall(channelId)`. This works. But `joinCall(existingCall.id)` -- if `existingCall` is stale or the check returned a call without ID, it calls `/calls/undefined/join` (bug #3, #8).

6. **Dashboard redirect loop (bug #5)**: Looking at sidebar.tsx -- there's NO `/dashboard` link in sidebar. The route exists at `/_authenticated/dashboard` but sidebar links go to `/projects`, `/messenger`, `/settings/profile`, `/trash`. Once user navigates to `/dashboard` (e.g., from notification click for call), they can't get back easily. But the actual "can't open dashboard" issue -- the dashboard route imports from `@/stores/dashboard-store` and `@/components/dashboard/widget-grid` and `@/components/dashboard/dashboard-templates`. If these don't exist, the page crashes.

7. **React error #185 (bug #1)**: GenericPreview in link-preview.tsx returns `null` on error/empty -- that's fine in JSX. The real crash source: looking at the project $projectId route -- it's just `<Outlet />`. If no child route matches, Outlet renders nothing (undefined). Actually TanStack Router handles that. Let me reconsider -- the crash is in messenger. The ChannelList's `ChannelIcon` function uses a switch with 4 cases but no `default` -- if channel_type is something unexpected, it returns `undefined` (bug #1 - React error #185!).

## Resolution

root_cause: Multiple bugs - ChannelIcon missing default return, no query invalidation on channel create, WS nginx config incomplete, notification routing wrong, no emoji in status, call join with undefined ID
fix: See individual fixes below
verification: pending
files_changed: []
