---
phase: 03-messenger
verified: 2026-03-21T00:00:00Z
status: passed
score: 5/5 must-haves verified
re_verification: false
---

# Phase 03: Messenger Verification Report

**Phase Goal:** Пользователи могут общаться в реальном времени через личные сообщения, каналы и треды с форматированным текстом, реакциями, голосовыми сообщениями и индикаторами присутствия
**Verified:** 2026-03-21
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (from ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Пользователь может отправлять и получать сообщения в реальном времени через личные сообщения, групповые чаты, публичные и приватные каналы | VERIFIED | `service.go` SendMessage + IsMember check; `handler.go` POST /channels/:id/messages; NATS `messenger.channel.{id}.message_sent`; WS bridge `messenger.channel.>` -> room `channel:{id}`; frontend `use-messages.ts` subscribes to `channel:{channelId}` room and handles `message_sent` events |
| 2 | Пользователь может отвечать в тредах, ставить реакции эмодзи, форматировать текст (жирный/курсив/код/списки) и записывать голосовые сообщения | VERIFIED | `thread-panel.tsx` renders replies with `MessageComposer` wired to `threadId`; `reaction-picker.tsx` + `toggleReaction` API; `message-composer.tsx` uses Plate editor (`usePlateEditor`, `Plate`, `PlateContent`); `voice-recorder.tsx` uses MediaRecorder API with opus codec and uploads via `uploadFile` |
| 3 | Каналы проекта создаются автоматически при создании проекта; каналы только для чтения работают для сообщений ботов | VERIFIED | `auto_channels.go` subscribes to NATS `project.created` (matches outbox worker publish subject); creates `#project-{key}` public channel and adds project members; `service.go` SendMessage checks `ch.IsReadonly` and blocks non-owner/non-admin |
| 4 | Пользователь может редактировать/удалять свои сообщения, закреплять сообщения, добавлять в закладки и искать по всем беседам | VERIFIED | `service.go` EditMessage (authorship check), DeleteMessage (authorship + admin check); TogglePin with Ent PinnedMessage; AddBookmark/ListBookmarks via Ent Bookmark; SearchMessages via MongoDB text index; all endpoints registered in handler.go; `search-dialog.tsx`, `pinned-messages.tsx`, `bookmarks-panel.tsx` all substantive |
| 5 | Статус онлайн/офлайн/отошёл виден в реальном времени, и счётчики непрочитанных корректно обновляются по каналам | VERIFIED | `presence.go` Redis sorted set with Heartbeat/GetStatus (online/away/offline thresholds); `unread.go` Redis hash with Increment (skip sender), MarkRead, GetAll; `presence-indicator.tsx` renders green/gray/yellow dots; `channel-list.tsx` shows unread badges and PresenceIndicator for DMs; StartCleanupLoop wired in main.go |

**Score:** 5/5 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `backend/ent/schema/channel.go` | Ent schema for channels | VERIFIED | Exists with channel_type enum, dm_key, is_readonly, project edge |
| `backend/internal/messenger/mongo_repository.go` | MongoDB CRUD with EnsureIndexes | VERIFIED | EnsureIndexes creates 3 indexes: (channel_id, created_at DESC), (channel_id, thread_id, created_at DESC), text on content_text |
| `backend/internal/messenger/service.go` | Business logic with 19 methods | VERIFIED | 745 lines, SendMessage/EditMessage/DeleteMessage/ToggleReaction/ListMessages/JoinChannel/CreateOrGetDM and 12 more |
| `backend/internal/messenger/handler.go` | RegisterRoutes with all endpoints | VERIFIED | 731 lines, RegisterRoutes registers channels CRUD, messages CRUD, reactions, pin, forward, replies, presence, unread, pins, search, bookmarks, DM |
| `backend/internal/messenger/service_test.go` | Unit tests >= 200 lines | VERIFIED | 923 lines, 20 test functions |
| `backend/internal/messenger/presence.go` | Redis presence tracking | VERIFIED | 134 lines; Heartbeat, GetStatus, GetBulkStatus, SetCustomStatus, CleanupStale, StartCleanupLoop |
| `backend/internal/messenger/unread.go` | Redis unread counts | VERIFIED | 61 lines; Increment (pipeline + skip sender), MarkRead, GetAll |
| `backend/internal/messenger/auto_channels.go` | Auto-channel on project creation | VERIFIED | HandleProjectCreated subscribes to `project.created`, creates `#project-{key}`, seeds project members |
| `backend/cmd/server/main.go` | Module wiring | VERIFIED | MongoDB client, CombinedRepository, PresenceService, UnreadService, MessengerService, Handler.RegisterRoutes, AutoChannelService.Start, StartCleanupLoop goroutine |
| `frontend/src/api/messenger.ts` | REST API client | VERIFIED | 346 lines; getChannels, getMessages, sendMessage, editMessage, deleteMessage, toggleReaction, togglePin, forwardMessage, getThreadReplies, createOrGetDM, searchMessages, getPresence, sendHeartbeat, getUnreadCounts, markChannelRead, getChannelPins, getBookmarks, addBookmark, removeBookmark |
| `frontend/src/components/messenger/chat-view.tsx` | Virtualized chat | VERIFIED | 183 lines; useVirtualizer with estimateSize 80px, overscan 5; infinite scroll up; auto-scroll down; markRead on channel open; SearchDialog + PinnedMessages in header; ThreadPanel conditional render |
| `frontend/src/components/messenger/voice-recorder.tsx` | MediaRecorder voice recording | VERIFIED | 187 lines; navigator.mediaDevices.getUserMedia; MediaRecorder with audio/webm;codecs=opus; 5-minute MAX_DURATION_MS; uploadFile integration; error handling for mic unavailable |
| `frontend/src/components/messenger/thread-panel.tsx` | Thread view | VERIFIED | 100 lines; useThreadMessages hook; parent message display; MessageComposer with threadId prop |
| `frontend/src/components/messenger/presence-indicator.tsx` | Online/offline/away indicator | VERIFIED | green/gray/yellow dots via Tailwind bg-green-500/bg-gray-400/bg-yellow-500 |
| `frontend/src/components/messenger/search-dialog.tsx` | Message search dialog | VERIFIED | Uses searchMessages API; scope toggle (current channel / all); Dialog with Input and results list |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `service.go` | NATS `messenger.channel.{id}.message_sent` | `nats.Publish` in publishEvent | VERIFIED | Line 360: `s.publishEvent(fmt.Sprintf(SubjectMessageSent, channelID.String()), ...)` |
| `backend/internal/ws/bridge.go` | WS hub BroadcastToRoom | NATS subscription `messenger.channel.>` | VERIFIED | Line 50: `nc.Subscribe("messenger.channel.>", ...)` -> room `channel:{channelID}` |
| `backend/cmd/server/main.go` | messenger handler RegisterRoutes | module wiring | VERIFIED | Line 153: `messengerHandler.RegisterRoutes(api, requireAuth, rbacSvc)` |
| `auto_channels.go` | NATS `project.created` | NATS subscription | VERIFIED | Line 50: `nc.Subscribe("project.created", ...)` matching outbox worker publish |
| `frontend/src/hooks/use-messages.ts` | WebSocket room `channel:{channelId}` | subscribe + addListener | VERIFIED | Line 34: `room = channel:${channelId}`; subscribes on mount, handles message_sent/message_edited/message_deleted events |
| `frontend/src/components/messenger/chat-view.tsx` | @tanstack/react-virtual | useVirtualizer | VERIFIED | Line 2: `import { useVirtualizer }`, line 50: `useVirtualizer({count, getScrollElement, estimateSize: () => 80, overscan: 5})` |
| `frontend/src/api/messenger.ts` | `/api/messenger` endpoints | fetch via api client | VERIFIED | All 19 functions call `/messenger/*` paths |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| MSG-01 | 03-01 | Личные сообщения (1-на-1) и групповые чаты | SATISFIED | CreateOrGetDM with dm_key dedup; channel_type dm/group; handler POST /dm |
| MSG-02 | 03-02 | Публичные и приватные каналы | SATISFIED | JoinChannel differentiates public (allowed) vs private (forbidden); service.go switch on channel_type |
| MSG-03 | 03-02 | Автоматически создаваемые каналы проекта | SATISFIED | auto_channels.go subscribes to project.created, creates #project-{key} with member seeding |
| MSG-04 | 03-02 | Каналы только для чтения | SATISFIED | SendMessage checks ch.IsReadonly; only owner/admin can post; chat-view.tsx hides composer for is_read_only |
| MSG-05 | 03-01, 03-03 | Ответы в тредах | SATISFIED | IncrementReplyCount in SendMessage for thread replies; thread-panel.tsx; useThreadMessages hook |
| MSG-06 | 03-01, 03-03 | Emoji-реакции | SATISFIED | ToggleReaction in service; reaction-picker.tsx; toggleReaction API; NATS ReactionToggledEvent |
| MSG-07 | 03-01, 03-03 | Форматирование текста | SATISFIED | Plate editor in message-composer.tsx (usePlateEditor with bold/italic/code); RenderNode in message-item.tsx |
| MSG-08 | 03-01 | Редактирование и удаление сообщений | SATISFIED | EditMessage checks authorship; DeleteMessage checks authorship OR admin/owner role; soft delete in MongoDB |
| MSG-09 | 03-03, 03-04 | Пересылка, закрепление и закладки | SATISFIED | ForwardMessage creates copy with forwarded_from_id; TogglePin/ListPins via Ent PinnedMessage; AddBookmark/RemoveBookmark; forward-dialog.tsx, pinned-messages.tsx, bookmarks-panel.tsx |
| MSG-10 | 03-01, 03-03 | Голосовые сообщения | SATISFIED | SendMessage accepts voice_url/voice_duration_ms; voice-recorder.tsx uses MediaRecorder + uploadFile; message-item.tsx renders audio player |
| MSG-12 | 03-04 | Глобальный поиск и поиск по каналу | SATISFIED | SearchMessages uses MongoDB text index; search-dialog.tsx with scope toggle; handler GET /search |
| MSG-13 | 03-04 | Статус онлайн/офлайн/отошёл | SATISFIED | PresenceService with Redis sorted set; GetBulkStatus; Heartbeat; presence-indicator.tsx; usePresence hook with 30s polling |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| None found | - | - | - | No stubs, placeholders, or empty implementations detected in verified files |

### Human Verification Required

#### 1. Real-time Message Delivery End-to-End

**Test:** Open two browser tabs logged in as different users. Join the same channel. Send a message from tab 1.
**Expected:** Message appears in tab 2 within ~100ms without page refresh.
**Why human:** Cannot verify NATS -> WebSocket -> React state update chain programmatically without running the full stack.

#### 2. Voice Message Record and Playback

**Test:** Click voice message button, grant microphone access, record a short message, send it. Verify the audio player appears in the channel and audio plays back.
**Expected:** Audio player renders with duration, audio plays correctly.
**Why human:** MediaRecorder API and audio playback require real browser environment.

#### 3. Auto-Channel Creation on Project Creation

**Test:** Create a new project. Navigate to the Messenger section.
**Expected:** A `#project-{key}` channel appears automatically in the channel list with all project members already added.
**Why human:** Requires full running stack (outbox worker + NATS + auto_channels subscription) to verify end-to-end.

#### 4. Presence Status Display Accuracy

**Test:** Log in as user A. In another tab, check user A's presence indicator.
**Expected:** Green dot when active, yellow dot after 5 minutes of inactivity, gray after 90 seconds more.
**Why human:** Requires real-time observation with time passage. Cannot simulate Redis TTL behavior in static analysis.

### Gaps Summary

No gaps found. All 5 observable truths are verified, all 12 required requirements have implementation evidence, all key links are wired, and no blocker anti-patterns were detected.

The messenger module is fully implemented across backend (Go/Fiber/Ent/MongoDB/Redis/NATS) and frontend (React/TypeScript/Plate/TanStack). The 4 human verification items are all positive-path checks rather than gaps — they confirm runtime behavior that the static code analysis already supports.

---
_Verified: 2026-03-21_
_Verifier: Claude (gsd-verifier)_
