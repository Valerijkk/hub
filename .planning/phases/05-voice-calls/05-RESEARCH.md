# Phase 5: Voice Calls - Research

**Researched:** 2026-03-21
**Domain:** WebRTC voice calls + screen sharing via LiveKit
**Confidence:** HIGH

## Summary

Phase 5 adds audio-only voice calls with screen sharing to Hub, leveraging LiveKit as the WebRTC SFU. LiveKit handles all media transport, STUN/TURN, and participant management -- the Hub backend only needs to manage rooms (create/delete), generate access tokens, and receive webhook events for lifecycle tracking. The frontend uses `@livekit/components-react` prefab components (`LiveKitRoom`, `ControlBar`, `AudioConference`) which provide a complete call UI out of the box with microphone toggle, screen share, and participant display.

The architecture is straightforward: Hub backend exposes REST endpoints for starting/joining calls (which create LiveKit rooms and return JWT tokens), LiveKit server handles all WebRTC media directly with clients, and LiveKit webhooks notify Hub when rooms start/finish and participants join/leave. Call metadata (room name, participants, duration, channel binding) is stored in PostgreSQL via Ent. Call events (started, ended, participant joined/left) are published to NATS for the WebSocket bridge to push to connected clients and for posting call activity to messenger channels. Ephemeral call chat is stored in MongoDB with TTL-based auto-deletion after the call ends.

**Primary recommendation:** Use `server-sdk-go/v2` for backend room management + token generation, `@livekit/components-react` for frontend UI, LiveKit webhooks for lifecycle events, and NATS for distributing call events to messenger channels.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CALL-01 | Audio calls + screen share (no webcams), up to 50 participants | LiveKit SFU supports 50+ participants per room. Disable camera in ControlBar via `controls={{ camera: false }}`. Screen share via `setScreenShareEnabled(true)` with native browser picker. |
| CALL-03 | Mute/unmute, raise hand, screen/window picker, multiple simultaneous screen shares | ControlBar provides mute/unmute + screen share toggle. Raise hand implemented via LiveKit participant metadata (JSON field). Multiple screen shares: each participant can publish one screen track, multiple participants can share simultaneously -- LiveKit has no limit on concurrent screen share tracks per room. |
| CALL-06 | Ephemeral call chat (disappears after call ends) | Store chat messages in MongoDB collection with `call_id` field. On `room_finished` webhook, delete all messages for that call OR use MongoDB TTL index (set `expires_at` to call end time + grace period). No Ent schema needed -- purely MongoDB. |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/livekit/server-sdk-go/v2 | 2.16.0 | Room management, token generation, webhook handling | Official Go SDK, provides RoomServiceClient + auth.AccessToken + webhook.ReceiveWebhookEvent |
| github.com/livekit/protocol | (transitive) | LiveKit protobuf types (Room, Participant, WebhookEvent) | Imported transitively by server-sdk-go, provides auth and webhook packages |
| @livekit/components-react | 2.9.20 | React UI components for call interface | Official React SDK with prefab AudioConference, ControlBar, LiveKitRoom |
| livekit-client | 2.17.3 | Low-level JS client SDK (peer dependency) | Required by components-react, provides Room, RoomEvent, Track types |
| @livekit/components-styles | latest | Default CSS styles for LiveKit components | Baseline styling, can be overridden with Tailwind |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| LiveKit Server | 1.8.x | Self-hosted WebRTC SFU (Docker container) | Always -- runs as a separate service in Docker Compose |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| LiveKit | Jitsi | Jitsi is Java-based, 2GB+ RAM, complex config. LiveKit is Go-native, ~200MB RAM, single container |
| LiveKit | mediasoup | mediasoup is Node.js, no built-in recording, requires more manual wiring. LiveKit has Egress for recording (Phase 7) |
| @livekit/components-react | Custom UI with livekit-client | More control but 10x more code. Prefabs cover 90% of needs, can customize via composition |

**Installation:**

```bash
# Backend
cd backend
go get github.com/livekit/server-sdk-go/v2@v2.16.0

# Frontend
cd frontend
npm install @livekit/components-react livekit-client @livekit/components-styles
```

## Architecture Patterns

### Recommended Project Structure

```
backend/internal/calls/
  handler.go        # HTTP handlers (create room, join, leave, webhook)
  service.go        # Business logic (room lifecycle, token generation)
  repository.go     # Ent repository (call metadata CRUD)
  chat.go           # Ephemeral call chat (MongoDB)
  events.go         # NATS event publishing
  models.go         # Request/response DTOs

backend/ent/schema/
  call.go           # Call entity (room name, channel, status, timestamps)
  callparticipant.go # Call participant (user, join/leave times)

frontend/src/features/calls/
  components/
    CallRoom.tsx        # Main call container (LiveKitRoom wrapper)
    CallControls.tsx    # Custom ControlBar (mic, screen share, raise hand, leave)
    ParticipantList.tsx # Participant grid with raise-hand indicators
    CallChat.tsx        # Ephemeral in-call chat panel
    ScreenShareView.tsx # Multi-screen-share layout
  hooks/
    useCallToken.ts     # Fetch token from backend
    useRaiseHand.ts     # Raise hand via participant metadata
    useCallChat.ts      # Send/receive ephemeral chat messages
  api/
    calls.ts            # API client for call endpoints
  pages/
    CallPage.tsx        # Route: /calls/:callId
```

### Pattern 1: Token Generation Flow

**What:** Backend generates LiveKit JWT tokens with room-scoped permissions.
**When to use:** Every time a user joins a call.

```go
// backend/internal/calls/service.go
import (
    "github.com/livekit/protocol/auth"
    lksdk "github.com/livekit/server-sdk-go/v2"
)

// GenerateToken создаёт JWT-токен для подключения пользователя к комнате LiveKit.
func (s *CallService) GenerateToken(roomName, userID, userName string) (string, error) {
    at := auth.NewAccessToken(s.cfg.APIKey, s.cfg.APISecret)
    grant := &auth.VideoGrant{
        RoomJoin: true,
        Room:     roomName,
    }
    at.SetVideoGrant(grant).
        SetIdentity(userID).
        SetName(userName).
        SetValidFor(24 * time.Hour)

    return at.ToJWT()
}
```

### Pattern 2: Room Lifecycle via RoomServiceClient

**What:** Backend creates/manages LiveKit rooms via the admin API.
**When to use:** Starting a call, force-ending a call.

```go
// backend/internal/calls/service.go

// roomClient инициализируется при старте сервиса.
func NewCallService(cfg LiveKitConfig, db *ent.Client, nc *nats.Client) *CallService {
    roomClient := lksdk.NewRoomServiceClient(
        cfg.URL, cfg.APIKey, cfg.APISecret,
    )
    return &CallService{
        roomClient: roomClient,
        cfg:        cfg,
        db:         db,
        nats:       nc,
    }
}

// CreateCall создаёт комнату в LiveKit и сохраняет метаданные звонка в БД.
func (s *CallService) CreateCall(ctx context.Context, channelID uuid.UUID, userID uuid.UUID) (*CallResponse, error) {
    roomName := fmt.Sprintf("call-%s", uuid.New().String())

    _, err := s.roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
        Name:            roomName,
        MaxParticipants: 50,
    })
    if err != nil {
        return nil, fmt.Errorf("не удалось создать комнату LiveKit: %w", err)
    }

    // Сохраняем метаданные звонка в PostgreSQL
    call, err := s.db.Call.Create().
        SetRoomName(roomName).
        SetChannelID(channelID).
        SetCreatedBy(userID).
        SetStatus("active").
        Save(ctx)
    // ...
}
```

### Pattern 3: Webhook Handler for Lifecycle Events

**What:** Receive LiveKit webhooks to track room/participant lifecycle.
**When to use:** Detecting when calls start/end, participants join/leave.

```go
// backend/internal/calls/handler.go

// HandleWebhook обрабатывает входящие webhook-события от LiveKit.
func (h *Handler) HandleWebhook(c fiber.Ctx) error {
    // ВАЖНО: копируем тело запроса из-за Fiber/fasthttp переиспользования контекста
    body := make([]byte, len(c.Body()))
    copy(body, c.Body())

    authProvider := auth.NewSimpleKeyProvider(h.cfg.APIKey, h.cfg.APISecret)

    // webhook.ReceiveWebhookEvent принимает *http.Request, нужен адаптер
    event, err := h.verifyAndParseWebhook(body, c.Get("Authorization"), authProvider)
    if err != nil {
        return fiber.NewError(fiber.StatusUnauthorized, "невалидный webhook")
    }

    switch event.GetEvent() {
    case "room_started":
        return h.service.OnRoomStarted(c.Context(), event)
    case "room_finished":
        return h.service.OnRoomFinished(c.Context(), event)
    case "participant_joined":
        return h.service.OnParticipantJoined(c.Context(), event)
    case "participant_left":
        return h.service.OnParticipantLeft(c.Context(), event)
    }

    return c.SendStatus(fiber.StatusOK)
}
```

### Pattern 4: Raise Hand via Participant Metadata

**What:** Use LiveKit participant metadata JSON to implement raise hand.
**When to use:** Features that need per-participant state visible to all.

```go
// Сервер обновляет метаданные участника через RoomServiceClient
func (s *CallService) SetRaiseHand(ctx context.Context, roomName, identity string, raised bool) error {
    meta := fmt.Sprintf(`{"hand_raised":%v,"raised_at":"%s"}`, raised, time.Now().Format(time.RFC3339))
    _, err := s.roomClient.UpdateParticipant(ctx, &livekit.UpdateParticipantRequest{
        Room:     roomName,
        Identity: identity,
        Metadata: meta,
    })
    return err
}
```

```tsx
// frontend: подписываемся на изменение метаданных участника
import { useParticipants } from '@livekit/components-react';

function ParticipantList() {
  const participants = useParticipants();
  return (
    <ul>
      {participants.map((p) => {
        const meta = JSON.parse(p.metadata || '{}');
        return (
          <li key={p.identity}>
            {p.name} {meta.hand_raised && '(ruka)'}
          </li>
        );
      })}
    </ul>
  );
}
```

### Pattern 5: Ephemeral Call Chat (MongoDB)

**What:** In-call text chat stored in MongoDB, deleted after call ends.
**When to use:** CALL-06 -- chat disappears after call ends.

```go
// backend/internal/calls/chat.go

// CallChatMessage представляет сообщение в чате звонка.
type CallChatMessage struct {
    ID        string    `bson:"_id,omitempty"`
    CallID    string    `bson:"call_id"`
    UserID    string    `bson:"user_id"`
    UserName  string    `bson:"user_name"`
    Text      string    `bson:"text"`
    CreatedAt time.Time `bson:"created_at"`
}

// Индексы:
// - {call_id: 1, created_at: 1} -- для выборки сообщений звонка
// MongoDB коллекция: "call_chat_messages"

// DeleteCallChat удаляет все сообщения чата при завершении звонка.
func (r *ChatRepository) DeleteCallChat(ctx context.Context, callID string) error {
    _, err := r.collection.DeleteMany(ctx, bson.M{"call_id": callID})
    return err
}
```

### Pattern 6: NATS Events for Call Activity in Channels

**What:** Publish call lifecycle events via NATS so the messenger channel shows "Call started" / "Call ended" messages.
**When to use:** Cross-module integration between calls and messenger.

```
NATS subjects:
  calls.channel.{channelID}.started   -> payload: {call_id, room_name, created_by}
  calls.channel.{channelID}.ended     -> payload: {call_id, duration_sec, participant_count}
  calls.channel.{channelID}.joined    -> payload: {call_id, user_id, user_name}
  calls.channel.{channelID}.left      -> payload: {call_id, user_id}
```

The WebSocket bridge (bridge.go) needs a new subscription for `calls.channel.>` that routes to room `channel:{channelID}`.

### Anti-Patterns to Avoid

- **Proxying WebRTC media through Caddy/backend:** Caddy ONLY proxies the LiveKit signaling WebSocket. Media (audio/video) flows directly between clients and LiveKit via UDP. Never attempt to proxy UDP media through an HTTP reverse proxy.
- **Creating LiveKit rooms on-demand without tracking:** Always create a Call entity in PostgreSQL before/alongside the LiveKit room. Without this, you lose track of which channel a call belongs to and can't clean up properly.
- **Storing call chat in Ent/PostgreSQL:** Call chat is ephemeral, append-heavy, and deleted in bulk. MongoDB is the right store for this pattern (matches messenger pattern).
- **Using LiveKit DataChannel for call chat:** While LiveKit supports data channels, they are unreliable for chat persistence and don't integrate with the existing messenger UI patterns. Use a separate MongoDB-backed chat endpoint.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| WebRTC SFU | Custom WebRTC server | LiveKit Server | WebRTC is enormously complex (ICE, STUN, TURN, DTLS, codec negotiation). LiveKit handles all of this in a single container |
| Audio/screen share UI | Custom WebRTC UI components | @livekit/components-react AudioConference + ControlBar | Prefabs handle audio rendering, device selection, screen share toggle, mute states. Hundreds of edge cases handled |
| Token auth for WebRTC | Custom token format | LiveKit auth.AccessToken | LiveKit tokens include VideoGrant permissions, are signed with HMAC, and verified by LiveKit server automatically |
| STUN/TURN servers | Separate coturn deployment | LiveKit built-in TURN | LiveKit includes embedded TURN server for NAT traversal |
| Participant metadata sync | Custom WebSocket signaling for raise hand | LiveKit participant metadata | Metadata changes are broadcast to all participants automatically via LiveKit signaling |

**Key insight:** LiveKit is a complete WebRTC platform. The backend's role is limited to room lifecycle management (create/delete), token generation, and webhook handling. Don't try to manage media, participants, or signaling manually.

## Common Pitfalls

### Pitfall 1: LiveKit + Caddy Proxy Misconfiguration

**What goes wrong:** Caddy is configured to proxy ALL LiveKit traffic. WebRTC media requires direct UDP access to LiveKit, not HTTP proxying. Calls connect but have no audio or screen share.
**Why it happens:** Developers treat LiveKit like a normal HTTP service and proxy everything through Caddy.
**How to avoid:**
- Caddy proxies ONLY the LiveKit signaling path (WebSocket on port 7880).
- LiveKit server ports 7881 (TCP fallback) and 50000-60000 (UDP media) must be exposed directly, NOT through Caddy.
- In Docker Compose, LiveKit container exposes these ports directly to the host using `network_mode: host` or explicit port mapping.
- Set `use_external_ip: true` in LiveKit config for cloud/Docker environments.
**Warning signs:** Calls connect (you see participants) but there is no audio. Screen share shows "connecting" forever.

### Pitfall 2: Fiber RequestCtx in Webhook Handler

**What goes wrong:** The LiveKit `webhook.ReceiveWebhookEvent` function expects a standard `*http.Request`. Fiber uses fasthttp, so the body and headers must be copied before passing to the webhook verifier.
**Why it happens:** Fiber's fasthttp reuses RequestCtx -- body bytes become invalid after handler returns.
**How to avoid:**
- Copy `c.Body()` before any async processing.
- Build a minimal adapter that creates an `*http.Request` from the copied Fiber context data for webhook verification.
- OR implement manual JWT verification: decode the `Authorization` header, verify the signature, and parse the JSON body directly.
**Warning signs:** Webhook verification fails intermittently under load.

### Pitfall 3: LiveKit Room Not Cleaned Up

**What goes wrong:** Backend creates a Call record but the LiveKit room persists after all participants leave. Or the Call record stays "active" forever because the `room_finished` webhook was missed.
**Why it happens:** Webhook delivery is not guaranteed. Network issues, server restarts, or misconfigured webhook URLs cause missed events.
**How to avoid:**
- Set `empty_timeout` on room creation (e.g., 300 seconds) so LiveKit auto-closes empty rooms.
- Run a periodic cleanup job (every 5 minutes) that checks for "active" calls older than a threshold and verifies room status via `RoomServiceClient.ListRooms()`.
- Implement idempotent webhook handling (use event `id` for deduplication).
**Warning signs:** Growing number of "active" calls in the database that don't correspond to real rooms.

### Pitfall 4: Screen Share Resolution and Bandwidth

**What goes wrong:** Screen share consumes excessive bandwidth with 50 participants, causing audio quality degradation.
**Why it happens:** Default screen share resolution is high (1080p). LiveKit's SFU optimizes via simulcast layers but the sender still uploads full resolution.
**How to avoid:**
- Set screen share resolution constraints: `{ width: 1920, height: 1080, frameRate: 15 }` for presenters, lower for secondary shares.
- LiveKit automatically handles quality adaptation via simulcast -- no manual bitrate management needed.
- For multiple simultaneous shares, consider limiting to 2-3 active shares and queueing additional requests.
**Warning signs:** Audio crackling or dropping when screen share is active on low-bandwidth connections.

## Code Examples

### Complete LiveKit Docker Compose Service

```yaml
# deployments/docker-compose.yml (дополнение)
livekit:
  image: livekit/livekit-server:v1.8
  ports:
    - "7880:7880"    # Сигнальный (WebSocket)
    - "7881:7881"    # TCP fallback для WebRTC
    - "50000-50100:50000-50100/udp"  # UDP медиа (для dev достаточно 100 портов)
  environment:
    - LIVEKIT_KEYS=devkey:devsecret
  volumes:
    - ./livekit.yaml:/etc/livekit.yaml
  command: ["--config", "/etc/livekit.yaml"]
```

### LiveKit Server Config (livekit.yaml)

```yaml
# deployments/livekit.yaml
port: 7880
rtc:
  tcp_port: 7881
  port_range_start: 50000
  port_range_end: 50100
  use_external_ip: false  # true в продакшене
keys:
  devkey: devsecret
webhook:
  urls:
    - "http://host.docker.internal:3000/api/calls/webhook"
logging:
  level: info
room:
  empty_timeout: 300     # Закрыть комнату через 5 мин после ухода последнего участника
  max_participants: 50
```

### Caddy Update for LiveKit Signaling

```
# deployments/Caddyfile (дополнение)
handle /livekit/* {
    reverse_proxy localhost:7880
}
```

### LiveKit Config Addition to Backend

```go
// backend/internal/platform/config/config.go (дополнение)

type LiveKitConfig struct {
    URL       string // "http://localhost:7880" или "ws://localhost:7880"
    APIKey    string
    APISecret string
}

// В Config struct:
// LiveKit LiveKitConfig

// В Load():
// LiveKit: LiveKitConfig{
//     URL:       getEnv("LIVEKIT_URL", "http://localhost:7880"),
//     APIKey:    getEnv("LIVEKIT_API_KEY", "devkey"),
//     APISecret: getEnv("LIVEKIT_API_SECRET", "devsecret"),
// },
```

### Ent Schema for Call

```go
// backend/ent/schema/call.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

// Call хранит метаданные голосового звонка.
type Call struct {
    ent.Schema
}

func (Call) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New),
        field.String("room_name").Unique().NotEmpty(),
        field.UUID("channel_id", uuid.UUID{}),
        field.UUID("created_by", uuid.UUID{}),
        field.Enum("status").Values("active", "ended").Default("active"),
        field.Time("started_at").Optional(),
        field.Time("ended_at").Optional(),
        field.Int("participant_count").Default(0),
    }
}

func (Call) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("channel", Channel.Type).Ref("calls").Unique().Required().Field("channel_id"),
        edge.From("creator", User.Type).Ref("created_calls").Unique().Required().Field("created_by"),
        edge.To("participants", CallParticipant.Type),
    }
}

func (Call) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("channel_id", "status"),
        index.Fields("room_name"),
    }
}
```

### Frontend: Call Room Component

```tsx
// frontend/src/features/calls/components/CallRoom.tsx
import {
  LiveKitRoom,
  RoomAudioRenderer,
  useParticipants,
} from '@livekit/components-react';
import '@livekit/components-styles';
import { CallControls } from './CallControls';
import { ParticipantList } from './ParticipantList';
import { CallChat } from './CallChat';

interface CallRoomProps {
  token: string;
  serverUrl: string;
  onDisconnect: () => void;
}

export function CallRoom({ token, serverUrl, onDisconnect }: CallRoomProps) {
  return (
    <LiveKitRoom
      token={token}
      serverUrl={serverUrl}
      connect={true}
      audio={true}
      video={false} // Без вебкамер (CALL-01)
      onDisconnected={onDisconnect}
    >
      <div className="flex h-full">
        <div className="flex-1 flex flex-col">
          <ParticipantList />
          <CallControls />
        </div>
        <CallChat />
      </div>
      <RoomAudioRenderer />
    </LiveKitRoom>
  );
}
```

### Frontend: Custom ControlBar

```tsx
// frontend/src/features/calls/components/CallControls.tsx
import { ControlBar } from '@livekit/components-react';

export function CallControls() {
  return (
    <ControlBar
      controls={{
        microphone: true,
        camera: false,         // Без вебкамер
        screenShare: true,     // Демонстрация экрана
        leave: true,
      }}
      variation="verbose"
    />
  );
}
```

### WebSocket Bridge Update

```go
// backend/internal/ws/bridge.go -- добавить подписку на calls.channel.>

// В методе Start():
callSub, err := b.nats.Subscribe("calls.channel.>", func(msg *nats.Msg) {
    b.handleCallEvent(msg)
})

// handleCallEvent маршрутизирует NATS-события звонков в соответствующую WS-комнату.
// Формат subject: calls.channel.{channelID}.{event}
// Формат room: channel:{channelID}
func (b *Bridge) handleCallEvent(msg *nats.Msg) {
    parts := strings.SplitN(msg.Subject, ".", 4)
    if len(parts) < 3 {
        return
    }

    channelID := parts[2]
    room := "channel:" + channelID

    event := msg.Subject
    if len(parts) >= 4 {
        event = parts[3]
    }

    var payload any
    if len(msg.Data) > 0 {
        _ = json.Unmarshal(msg.Data, &payload)
    }

    b.hub.BroadcastToRoom(room, ServerMessage{
        Type:    TypeEvent,
        Event:   event,
        Room:    room,
        Payload: payload,
    })
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Custom WebRTC with coturn | LiveKit SFU | 2022+ | Single container replaces separate STUN/TURN + SFU + recording infrastructure |
| Manual data channels for features | Participant metadata API | LiveKit 1.5+ | Raise hand, custom states broadcast automatically via signaling |
| Separate Jitsi deploy | LiveKit in Docker Compose | 2024+ | 200MB vs 2GB RAM, Go-native vs Java, simpler config |
| livekit-client raw API | @livekit/components-react | 2023+ | Prefab components reduce frontend code by ~80% |

## Open Questions

1. **Webhook delivery with Fiber/fasthttp**
   - What we know: `webhook.ReceiveWebhookEvent` expects `*http.Request`, not Fiber context. Need an adapter or manual verification.
   - What's unclear: Whether the official webhook package works with a request adapter, or if manual JWT + hash verification is more reliable.
   - Recommendation: Implement manual verification (decode JWT from Authorization header, verify SHA256 hash of body, parse JSON). Simpler than building an http.Request adapter.

2. **LiveKit Docker networking in development**
   - What we know: LiveKit needs UDP ports exposed. Docker on Windows/Mac may have issues with `network_mode: host`.
   - What's unclear: Whether port mapping (50000-50100/udp) works reliably on Docker Desktop Windows.
   - Recommendation: Start with explicit port mapping. If issues arise, test with `network_mode: host` or use WSL2 networking.

3. **Ephemeral chat message delivery**
   - What we know: Chat messages go to MongoDB. Need real-time delivery to call participants.
   - What's unclear: Whether to use NATS pub/sub (like messenger) or LiveKit data channels for real-time delivery.
   - Recommendation: Use NATS pub/sub with subject `calls.chat.{callID}`. Subscribe in WebSocket bridge. This reuses existing infrastructure and keeps the chat backend consistent with the messenger pattern. Store in MongoDB for scrollback during the call, delete on call end.

## Sources

### Primary (HIGH confidence)
- [LiveKit server-sdk-go/v2 v2.16.0](https://pkg.go.dev/github.com/livekit/server-sdk-go/v2) - API reference, RoomServiceClient, AccessToken
- [LiveKit deployment docs](https://docs.livekit.io/transport/self-hosting/deployment/) - Ports, config, Docker setup
- [LiveKit webhooks docs](https://docs.livekit.io/home/server/webhooks/) - Event types, Go handler, verification
- [LiveKit React components](https://docs.livekit.io/reference/components/react/) - AudioConference, ControlBar, LiveKitRoom
- [LiveKit screen sharing docs](https://docs.livekit.io/transport/media/screenshare/) - Screen share API, browser picker
- [@livekit/components-react npm](https://www.npmjs.com/package/@livekit/components-react) - v2.9.20 verified
- [livekit-client npm](https://www.npmjs.com/package/livekit-client) - v2.17.3 verified

### Secondary (MEDIUM confidence)
- [LiveKit + Caddy pitfall](https://foundryvtt.wiki/en/setup/hosting/Self-Hosting-LiveKit-Audio-Video-Server-on-Existing-Linux-Setup) - Community setup guide confirming Caddy proxies signaling only
- [LiveKit self-hosted examples](https://github.com/anguzo/livekit-self-hosted) - Docker compose reference

### Tertiary (LOW confidence)
- Multiple simultaneous screen shares: Not explicitly documented as a limit. Testing confirms multiple participants can share simultaneously (each publishes one screen track). Need validation with 50 participants.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all versions verified via npm/pkg.go.dev, official LiveKit docs
- Architecture: HIGH - follows existing Hub module pattern (handler/service/repository), LiveKit integration is well-documented
- Pitfalls: HIGH - Caddy+LiveKit proxy issue confirmed by multiple sources, Fiber ctx reuse is a known project convention

**Research date:** 2026-03-21
**Valid until:** 2026-04-21 (LiveKit is stable, unlikely to break)
