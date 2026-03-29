---
phase: 06-notifications-search-dashboards
plan: 01
subsystem: notification
tags: [notification, nats, jetstream, telegram, websocket, zustand, ent]
dependency_graph:
  requires: [auth, tracker, messenger, tms, calls, ws-bridge]
  provides: [notification-service, notification-consumer, telegram-sender, notification-api, notification-frontend]
  affects: [ws/bridge.go, cmd/server/main.go, ent/schema/user.go, frontend/topbar]
tech_stack:
  added: [telego]
  patterns: [jetstream-durable-consumer, debounce-dedup, nil-safe-sender, quiet-hours, ws-notification-delivery]
key_files:
  created:
    - backend/ent/schema/notification.go
    - backend/ent/schema/notificationpreference.go
    - backend/internal/notification/events.go
    - backend/internal/notification/dto.go
    - backend/internal/notification/repository.go
    - backend/internal/notification/service.go
    - backend/internal/notification/consumer.go
    - backend/internal/notification/telegram.go
    - backend/internal/notification/handler.go
    - backend/internal/notification/service_test.go
    - frontend/src/api/notifications.ts
    - frontend/src/stores/notification-store.ts
    - frontend/src/components/notifications/notification-bell.tsx
    - frontend/src/components/notifications/notification-panel.tsx
    - frontend/src/components/notifications/notification-settings.tsx
  modified:
    - backend/ent/schema/user.go
    - backend/internal/ws/bridge.go
    - backend/cmd/server/main.go
    - backend/go.mod
    - backend/go.sum
    - frontend/src/api/client.ts
    - frontend/src/components/layout/topbar.tsx
decisions:
  - "JetStream durable consumer 'notification-dispatcher' with AckExplicit, MaxDeliver=5, AckWait=30s for reliable notification delivery"
  - "2-second in-memory debounce window keyed by userID:type:entityID to prevent duplicate notifications"
  - "Nil-safe TelegramSender: nil botToken returns nil sender, Send() on nil is no-op"
  - "Quiet hours: HH:MM format stored on User, midnight crossover handled, urgent notifications bypass"
  - "Reuse existing TelegramConfig.BotToken for notification bot (same bot for auth and notifications)"
  - "DM-only message notifications to avoid channel noise; channel mentions via comment events"
  - "patch() method added to ApiClient for PATCH HTTP requests"
  - "useRef pattern for WS notification handler to avoid stale closures"
metrics:
  duration: 12min
  completed: "2026-03-21T04:10:00Z"
  tasks_completed: 2
  tasks_total: 2
  files_created: 15
  files_modified: 7
---

# Phase 06 Plan 01: Notification System Summary

Cross-cutting notification system with JetStream durable consumer, Ent schemas, in-app delivery via WS bridge, Telegram bot sender with rate limiting, per-user preferences with quiet hours, and React frontend with bell icon, panel, and settings.

## Completed Tasks

| Task | Name | Commit | Key Files |
|------|------|--------|-----------|
| 1 | Notification backend (Ent schemas, consumer, service, handler, telegram, tests) | 71e0147 | notification/*.go, ent/schema/notification*.go, ws/bridge.go, main.go |
| 2 | Notification frontend (bell, panel, settings, WS integration) | 4c9723d | api/notifications.ts, notification-store.ts, notification-bell.tsx, notification-panel.tsx, notification-settings.tsx, topbar.tsx |

## Architecture

### Backend

- **Ent schemas**: `Notification` (user_id, type, title, body, entity_type, entity_id, project_id, actor_id, is_read, is_urgent) with composite indexes; `NotificationPreference` (user_id, notification_type, channel, enabled, project_id) with unique constraint
- **User schema extended**: quiet_hours_start/end (HH:MM, optional, nillable), edges to notifications and notification_preferences
- **JetStream consumer**: Durable consumer "notification-dispatcher" on stream "NOTIFICATIONS" with subjects project.>, messenger.channel.>, tms.>, calls.channel.>. WorkQueue retention, 7-day max age. Processes: assigned, status_changed, comment_added (with mentions), blocker_added, message_received (DM only), call_started, run_completed
- **Service**: Dispatch checks preferences (per-project then global fallback, default enabled), quiet hours (timezone-aware, midnight crossover), saves to DB, publishes to NATS notification.user.{userID} for WS, queues to Telegram sender
- **Telegram sender**: telego bot wrapper with buffered channel (cap 100), worker goroutine with 35ms rate limiting, MarkdownV2 formatting, nil-safe lifecycle
- **WS bridge**: Added notification.user.> subscription routing to SendToUser with event "notification.new"
- **Handler**: 6 REST endpoints under /api/notifications (list, unread-count, mark-read, mark-all-read, preferences, update-preference)

### Frontend

- **API client**: 6 functions wrapping REST endpoints; added patch() to ApiClient
- **Zustand store**: notifications[], unreadCount, page/total/hasMore state; WS listener for notification.new events; sonner toast on new notification
- **NotificationBell**: Bell icon with red badge (unread count), Popover trigger, WS listener setup with useRef pattern
- **NotificationPanel**: Scrollable list with type icons, relative timestamps, unread indicator (blue dot), load-more on scroll, mark-all-read button, empty state
- **NotificationSettings**: Checkbox table (9 types x 2 channels), quiet hours time inputs
- **Topbar**: Placeholder replaced with live NotificationBell component

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] ProjectMember field access via Edges**
- **Found during:** Task 1
- **Issue:** ProjectMember struct uses edge-based relations (not direct ProjectID/UserID fields)
- **Fix:** Changed handleRunCompleted to use project.QueryMembers().WithUser().All() edge traversal
- **Files modified:** backend/internal/notification/consumer.go

**2. [Rule 3 - Blocking] Channel.Type field name is ChannelType**
- **Found during:** Task 1
- **Issue:** Ent generated Channel struct uses `ChannelType` enum field, not `Type`
- **Fix:** Changed to `string(channel.ChannelType) != "dm"` comparison
- **Files modified:** backend/internal/notification/consumer.go

**3. [Rule 3 - Blocking] telego.SendMessage requires context parameter**
- **Found during:** Task 1
- **Issue:** telego v1.7.0 SendMessage() requires context.Context as first argument
- **Fix:** Added context.Background() to SendMessage call
- **Files modified:** backend/internal/notification/telegram.go

**4. [Rule 2 - Missing] ApiClient.patch() method**
- **Found during:** Task 2
- **Issue:** Frontend ApiClient had no patch() method for PATCH requests
- **Fix:** Added patch() method to ApiClient class
- **Files modified:** frontend/src/api/client.ts

### Cross-agent Impact

Backend notification files (Task 1) were committed by the parallel audit agent (06-04) in commit 71e0147 due to shared staging area. The code content is correct and unmodified. Frontend (Task 2) committed independently in 4c9723d.

## Verification

- `go build ./internal/notification/...` -- PASSED
- `go test ./internal/notification/... -v -count=1` -- 6 tests PASSED (SubjectToType, ParseTimeMinutes, EscapeMarkdownV2, FormatTelegramMessage, MatchSuffix, ShouldDispatch_Debounce)
- `npx tsc --noEmit` -- 0 notification-related errors (other errors from parallel agents)
- Notification Ent schemas generated properly
- WS bridge has notification.user.> subscription
- main.go initializes notification module with consumer, telegram sender, handler
