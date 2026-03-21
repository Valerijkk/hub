---
phase: 07-polish-extended
plan: 03
subsystem: calls
tags: [recording, egress, livekit, scheduled-calls, rrule, s3]
dependency_graph:
  requires: [calls-module, livekit-sdk, seaweedfs-s3]
  provides: [recording-service, scheduled-call-service, recording-ui, scheduled-call-ui]
  affects: [calls-handler, calls-repository, main.go, call-store]
tech_stack:
  added: [rrule-go]
  patterns: [LiveKit Egress, S3Upload, RRULE recurrence, cron ticker]
key_files:
  created:
    - backend/ent/schema/recording.go
    - backend/ent/schema/scheduledcall.go
    - backend/internal/calls/recording.go
    - backend/internal/calls/scheduled.go
    - backend/internal/calls/dto.go
    - backend/internal/calls/recording_test.go
    - frontend/src/modules/calls/recording-controls.tsx
    - frontend/src/modules/calls/recordings-list.tsx
    - frontend/src/modules/calls/scheduled-calls.tsx
  modified:
    - backend/ent/schema/call.go
    - backend/ent/schema/project.go
    - backend/ent/schema/user.go
    - backend/ent/schema/channel.go
    - backend/ent/schema/environment.go
    - backend/internal/calls/handler.go
    - backend/internal/calls/repository.go
    - backend/internal/calls/service.go
    - backend/cmd/server/main.go
    - frontend/src/api/calls.ts
    - frontend/src/stores/call-store.ts
decisions:
  - "EgressClient from lksdk.NewEgressClient for recording management"
  - "rrule-go library for RFC 5545 RRULE parsing and next occurrence calculation"
  - "60-second ticker goroutine for ProcessDueScheduledCalls cron"
  - "Recording limit: max 2 active recordings per call"
  - "S3 path: recordings/{projectID}/{callID}.mp4"
  - "Environment schema field renamed label -> env_label to avoid Ent codegen conflict (Rule 3)"
metrics:
  duration: 12min
  completed: "2026-03-21T05:13:18Z"
---

# Phase 07 Plan 03: Recording + Scheduled Calls Summary

Recording via LiveKit Egress with S3 upload to SeaweedFS, scheduled/recurring calls with RFC 5545 RRULE support, project recordings section with download, frontend recording controls with REC indicator.

## Task Results

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Recording + Scheduled calls backend | b2a654e | 15 files (schemas, services, handler, repo, dto, tests, main.go) |
| 2 | Recording + Scheduled calls frontend | 4cc4b08 | 5 files (api, components, store) |

## What Was Built

### Backend (Task 1)

**Ent Schemas:**
- `Recording`: egress_id, status (recording/processing/ready/failed), file_path, file_size, duration_seconds, started_at, ended_at; edges to Call, Project, User; indexes on project_id+created_at, call_id, egress_id
- `ScheduledCall`: title, description, scheduled_at, rrule (RFC 5545), channel_id, project_id, is_active; edges to Project, Channel, User; indexes on channel_id+scheduled_at, is_active+scheduled_at

**RecordingService (recording.go):**
- `StartRecording`: checks active recordings limit (max 2), creates RoomCompositeEgress with S3 output, saves Recording in DB, publishes NATS event
- `StopRecording`: calls egressClient.StopEgress, updates status to processing
- `OnEgressEnded`: webhook handler updates file_size, duration, status=ready, publishes ready event
- `ListProjectRecordings`: paginated list by project
- `GetRecordingDownloadURL`: returns presigned-style URL

**ScheduledCallService (scheduled.go):**
- `CreateScheduledCall`: validates RRULE, computes next occurrence
- `ListScheduledCalls`, `UpdateScheduledCall`, `DeleteScheduledCall`
- `ProcessDueScheduledCalls`: cron job finds due calls, publishes NATS event, computes next RRULE occurrence or deactivates

**Handler endpoints:**
- POST /api/calls/:callId/recording/start
- POST /api/calls/:callId/recording/stop
- GET /api/projects/:projectId/recordings
- GET /api/recordings/:recordingId/download
- POST /api/projects/:projectId/scheduled-calls
- GET /api/channels/:channelId/scheduled-calls
- PUT /api/scheduled-calls/:id
- DELETE /api/scheduled-calls/:id
- Webhook egress_ended handling

**main.go:** EgressClient, RecordingService, ScheduledCallService wiring + 60s ticker goroutine

### Frontend (Task 2)

**API (calls.ts):** startRecording, stopRecording, listProjectRecordings, getRecordingDownloadURL, createScheduledCall, listScheduledCalls, updateScheduledCall, deleteScheduledCall

**RecordingControls:** Start/stop recording button, REC indicator with timer visible to all participants, only initiator can stop

**RecordingsList:** Project recordings table with date filter, pagination, status badge, download button

**ScheduledCalls:** List with create form, recurrence presets (daily/weekly/weekdays/custom RRULE), human-readable labels, delete

**call-store:** isRecording, recordingId, recordingStartedBy state + setRecordingState action

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Environment schema field name conflict**
- **Found during:** Task 1 (go generate)
- **Issue:** `label` field in Environment schema conflicted with Ent generated `Label` constant
- **Fix:** Renamed field to `env_label`
- **Files modified:** backend/ent/schema/environment.go
- **Commit:** b2a654e

## Verification

- `go build ./cmd/server` -- passes
- `go test ./internal/calls/... -v -count=1` -- 10 tests pass (RRULE parsing, next occurrence, COUNT limit, status values, token generation)
- `npx vite build` -- passes (pre-existing TS errors in unrelated modules)
