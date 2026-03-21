---
phase: 05-voice-calls
plan: 02
subsystem: calls
tags: [livekit, webrtc, react, zustand, voice-calls, screen-share]

requires:
  - phase: 05-voice-calls
    provides: Backend call lifecycle API (create/join/leave/end/raise-hand), LiveKit Docker service
  - phase: 01-foundation
    provides: API client pattern, auth store, UI components (Button, Card, Tooltip)
  - phase: 03-messenger
    provides: Channel concept, messenger store pattern
provides:
  - API client for calls (createCall, joinCall, leaveCall, endCall, getActiveCall, raiseHand)
  - Zustand store for call state management (useCallStore)
  - CallRoom component wrapping LiveKitRoom with audio-only mode
  - CallControls with mic toggle, screen share, raise hand, leave, end call
  - ParticipantList with mute/hand/screenshare indicators
  - ScreenShareView supporting multiple simultaneous screen shares
  - CallButton for starting/joining calls from channel UI
  - CallPanel overlay for active call display
affects: [05-voice-calls, 06-dashboards]

tech-stack:
  added: ["@livekit/components-react", "livekit-client", "@livekit/components-styles"]
  patterns: [livekit-room-wrapper, custom-call-controls, participant-metadata-parsing]

key-files:
  created:
    - frontend/src/api/calls.ts
    - frontend/src/stores/call-store.ts
    - frontend/src/components/calls/CallRoom.tsx
    - frontend/src/components/calls/CallControls.tsx
    - frontend/src/components/calls/ParticipantList.tsx
    - frontend/src/components/calls/ScreenShareView.tsx
    - frontend/src/components/calls/CallButton.tsx
    - frontend/src/components/calls/CallPanel.tsx
    - frontend/src/vite-env.d.ts
  modified: []

key-decisions:
  - "Custom CallControls instead of ControlBar: needed raise hand button not available in built-in ControlBar"
  - "Participant metadata JSON parse for hand_raised indicator from LiveKit metadata"
  - "Fixed overlay panel (420x500px) for call UI instead of full-screen modal"

patterns-established:
  - "LiveKit room wrapper pattern: LiveKitRoom with RoomAudioRenderer + custom controls"
  - "Participant metadata parsing: JSON.parse(p.metadata) with safe fallback for hand_raised"

requirements-completed: [CALL-01, CALL-03]

duration: 4min
completed: 2026-03-21
---

# Phase 5 Plan 2: Call UI Summary

**React call UI with LiveKit components: audio-only rooms, mic/screen share controls, raise hand, participant list with indicators, and channel call button**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-21T03:22:30Z
- **Completed:** 2026-03-21T03:26:31Z
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments
- API client covering all 6 call endpoints with proper error handling (404 for getActiveCall returns null)
- Zustand store managing full call lifecycle (start, join, leave, end, raise hand)
- 6 React components: CallRoom, CallControls, ParticipantList, ScreenShareView, CallButton, CallPanel
- Audio-only mode enforced (video=false) per CALL-01 requirement
- Multiple simultaneous screen shares supported with responsive grid layout per CALL-03

## Task Commits

Each task was committed atomically:

1. **Task 1: npm install LiveKit, API client, Zustand store** - `e4cec78` (feat)
2. **Task 2: React components (CallRoom, CallControls, ParticipantList, ScreenShareView, CallButton, CallPanel)** - `9b1c582` (feat)

## Files Created/Modified
- `frontend/src/api/calls.ts` - API client with createCall, joinCall, leaveCall, endCall, getActiveCall, raiseHand
- `frontend/src/stores/call-store.ts` - Zustand store for call state (activeCall, token, isInCall, handRaised)
- `frontend/src/components/calls/CallRoom.tsx` - LiveKitRoom wrapper with audio=true, video=false
- `frontend/src/components/calls/CallControls.tsx` - Custom controls: mic, screen share, raise hand, leave, end
- `frontend/src/components/calls/ParticipantList.tsx` - Sorted participant list with mute/hand/screenshare indicators
- `frontend/src/components/calls/ScreenShareView.tsx` - Grid of VideoTrack components for screen shares
- `frontend/src/components/calls/CallButton.tsx` - Channel button to start/join calls
- `frontend/src/components/calls/CallPanel.tsx` - Fixed overlay panel for active call
- `frontend/src/vite-env.d.ts` - Vite client types for import.meta.env
- `frontend/package.json` - Added LiveKit dependencies
- `frontend/package-lock.json` - Lock file updated

## Decisions Made
- Custom CallControls instead of built-in ControlBar: needed raise hand button which ControlBar does not provide
- Participant metadata parsed as JSON with safe fallback for hand_raised field from LiveKit metadata
- CallPanel positioned as fixed overlay (420x500px) in bottom-right corner rather than full-screen modal for less intrusive UX

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added missing vite-env.d.ts for import.meta.env types**
- **Found during:** Task 2 (CallRoom component)
- **Issue:** TypeScript error: 'env' does not exist on type 'ImportMeta' -- project was missing Vite client type reference
- **Fix:** Created frontend/src/vite-env.d.ts with `/// <reference types="vite/client" />`
- **Files modified:** frontend/src/vite-env.d.ts
- **Verification:** npx tsc --noEmit passes
- **Committed in:** 9b1c582 (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Standard Vite setup file was missing. No scope creep.

## Issues Encountered
None beyond the auto-fixed deviation above.

## User Setup Required
None - LiveKit URL defaults to ws://localhost:7880, configurable via VITE_LIVEKIT_URL env var.

## Next Phase Readiness
- All call UI components ready for integration into messenger channel views
- Ready for ephemeral call chat (Phase 5 Plan 3)
- Ready for call recording UI (Phase 5 Plan 4 if applicable)
- CallButton can be placed in channel topbar by importing from components/calls/CallButton
- CallPanel should be rendered at app root level for global call overlay

## Self-Check: PASSED

All 9 created files verified. Both task commits (e4cec78, 9b1c582) confirmed in git log.

---
*Phase: 05-voice-calls*
*Completed: 2026-03-21*
