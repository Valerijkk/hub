# Phase 1: Foundation & Infrastructure - Research

**Researched:** 2026-03-20
**Domain:** Auth, Projects, RBAC, Real-time infra, Storage, Dev environment (Go + Fiber v3 + Ent + PostgreSQL + MongoDB + NATS + Redis + SeaweedFS)
**Confidence:** HIGH

## Summary

Phase 1 establishes every foundational subsystem that later phases depend on: authentication (Magic Link + Telegram), user profiles, project CRUD with membership, RBAC middleware (system + project roles with custom permissions), WebSocket gateway bridged to NATS, SeaweedFS file storage with presigned URLs, database schemas and migrations (PostgreSQL via Ent + MongoDB), Redis session/cache layer, Docker Compose dev environment, and the CI pipeline. This phase also addresses 6 of the 9 critical pitfalls identified in project research -- Fiber RequestCtx reuse conventions, PG+MongoDB outbox pattern, RBAC role hierarchy design, NATS core-vs-JetStream separation, WebSocket distributed delivery via NATS bridge, and Ent query conventions.

The project is greenfield -- no source code exists yet. Every directory, configuration file, Go module, and npm package must be created from scratch. The monorepo structure (backend/ + frontend/ + cli/ + deployments/) must be established first. The backend follows a modular monolith pattern with handler/service/repository layers per module, communicating through NATS events. The frontend is a React 19 SPA with TanStack Router (file-based routes), TanStack Query (server state), Zustand (client state), and shadcn/ui + Tailwind CSS 4 for components.

**Primary recommendation:** Build infrastructure layers bottom-up: dev environment (Docker Compose) -> database schemas (Ent + MongoDB) -> platform services (config, DB connections, NATS, Redis) -> auth module (Magic Link + Telegram + JWT + sessions) -> RBAC middleware -> project module -> WebSocket gateway + NATS bridge -> storage service (SeaweedFS) -> frontend shell (routing, auth, layout) -> onboarding UI. Establish Fiber RequestCtx copy conventions and test with `-race` flag from the very first handler.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| AUTH-01 | Magic Link login via email (no password) | Resend Go SDK v3 for email delivery, Redis for token storage (5min TTL, single-use), rate limit 3/10min |
| AUTH-02 | Telegram login (no password) | telego library for Bot API, Telegram Login Widget for web auth, HMAC-SHA256 hash verification |
| AUTH-03 | Admin configurable registration mode (open or invite-only) | System settings table in PG, admin-only API endpoint, enforced in auth middleware |
| AUTH-04 | User profile editing (name, avatar, email, Telegram, etc.) | Ent User schema with all fields, avatar upload via SeaweedFS presigned URL, timezone as string field |
| AUTH-05 | System roles: Admin and Member | Ent Role schema, system_role field on User, enforced via RBAC middleware |
| AUTH-06 | Multiple active sessions per user, session list, session termination | JWT access (15min) + refresh token in Redis (30 days), session metadata stored per token |
| AUTH-07 | Notification on new session creation | NATS event auth.session.created, delivered via WebSocket gateway to existing sessions |
| AUTH-08 | Onboarding flow for new users | Frontend multi-step wizard (profile -> tour -> theme/lang -> project), onboarding_completed flag on User |
| PROJ-01 | Project creation with auto-generated tracker, TMS, chat, channel, dashboard | Ent Project schema, NATS event triggers downstream resource creation (outbox pattern for PG+Mongo) |
| PROJ-02 | Project Owner manages roles, workflows, teams, access | Project-scoped RBAC with owner permission set, Ent privacy rules |
| PROJ-03 | Fixed project roles (Owner/Developer/Tester/Viewer) + custom RBAC roles | Permission bitmap per role, predefined roles as seed data, custom roles with same permission set |
| PROJ-04 | Invite by link + team bulk add | Invite token in Redis with TTL, team_members join table, bulk membership creation |
| PROJ-06 | Project archiving (read-only mode) | archived_at timestamp on Project, Ent privacy rule denies mutations when archived |
| PROJ-07 | Soft-delete with 30-day trash + export before deletion | deleted_at + scheduled_purge_at fields, NATS JetStream durable job for cleanup |
| UI-01 | Collapsible left sidebar + top bar (search, notifications, profile) | React shell layout with Zustand for sidebar state, shadcn/ui sidebar + header components |
| UI-07 | Error handling: toast, request ID, retry, offline mode, error pages | Fiber request ID middleware, frontend error boundary, TanStack Query retry, toast via sonner |
| UI-08 | File attachments everywhere with previews and drag-and-drop | SeaweedFS presigned URL flow, frontend dropzone component, image/PDF preview components |
</phase_requirements>

## Standard Stack

### Core (Phase 1 specific)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.26.x | Backend language | Goroutines for concurrent real-time handlers, single binary deployment |
| Fiber | v3.x | HTTP framework | Express-style API on fasthttp, v3 adds context.Context on fiber.Ctx, requires Go 1.25+ |
| Ent ORM | 0.14.x | PostgreSQL ORM | Code-gen type-safe queries, graph traversal for entity relations, Privacy Layer for RBAC |
| Atlas | latest | DB migrations | Declarative migrations from Ent schemas, versioned migration files |
| go-redis/v9 | 9.x | Redis client | Sessions, magic link tokens, RBAC cache, rate limiting |
| nats.go | 1.49.x | NATS client | Core NATS for real-time, JetStream for durable jobs |
| mongo-driver/v2 | 2.5.x | MongoDB driver | v2 current line for new projects |
| aws-sdk-go-v2 | latest | S3 client for SeaweedFS | Standard S3 presigned URL generation, endpoint-agnostic |
| resend-go/v3 | 3.x | Email sending | Magic Link delivery, 3000/month free tier |
| telego | 1.0.x | Telegram Bot API | 100% Bot API coverage, type-safe, fasthttp default |
| golang-jwt/jwt/v5 | 5.x | JWT tokens | Access token creation/validation, RS256 signing |
| React | 19.2.x | Frontend framework | Stable, useTransition/useOptimistic for optimistic UI |
| Vite | 8.x | Build tool | Rolldown bundler (10-30x faster), shared config with Vitest |
| TypeScript | 5.7.x | Type safety | Strict mode, end-to-end type safety with generated API types |
| TanStack Router | 1.x | Routing | File-based routes, type-safe search params, auto code-splitting |
| TanStack Query | 5.x | Server state | Cache, optimistic updates, background refetch |
| Zustand | 5.x | Client state | Sidebar state, auth state, WebSocket connection state |
| shadcn/ui | CLI v4 | UI components | Copy-paste Radix primitives, Tailwind-native |
| Tailwind CSS | 4.2.x | Styling | CSS-first config via @theme, no JS config file |
| sonner | latest | Toast notifications | shadcn/ui integrated toast library |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| slog (stdlib) | Go stdlib | Structured logging | JSON output for Loki, context-aware, every handler |
| Air | latest | Go hot reload | Development file watcher, rebuilds on save |
| golangci-lint | latest | Go linting | CI + pre-commit, 50+ linters in parallel |
| testify | 1.10.x | Go test assertions | assert/require/suite for all Go tests |
| testcontainers-go | 0.36.x | Integration tests | Real PG/Mongo/Redis/NATS in Docker for tests |
| mockery | 2.x | Mock generation | Generate mocks from Go interfaces |
| Vitest | 4.1.x | Frontend tests | Vite-native, browser mode stable |
| @testing-library/react | 16.x | Component tests | DOM-based behavioral testing |
| MSW | 2.x | API mocking | Service Worker intercepts for frontend tests |
| Biome | latest | Frontend formatting | Faster than Prettier |
| Docker Compose | 2.x | Dev environment | PG + Mongo + Redis + NATS + SeaweedFS |
| Caddy | 2.11.x | Reverse proxy | Auto-HTTPS, WebSocket proxying |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Fiber v3 | Chi | Chi uses net/http (broader middleware compatibility) but loses fasthttp performance |
| Ent | GORM | GORM has larger ecosystem but no graph traversal or Privacy Layer for RBAC |
| Resend | SMTP directly | More control but setup complexity, no DKIM/SPF out of box |
| telego | go-telegram-bot-api | More popular but incomplete Bot API, no codegen |

**Installation:**

```bash
# Backend (Go modules - run from backend/)
go mod init github.com/hub/backend
go get github.com/gofiber/fiber/v3
go get github.com/gofiber/contrib/v3/websocket
go get entgo.io/ent
go get ariga.io/atlas
go get go.mongodb.org/mongo-driver/v2
go get github.com/redis/go-redis/v9
go get github.com/nats-io/nats.go
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/aws/aws-sdk-go-v2/feature/s3/manager
go get github.com/resend/resend-go/v3
go get github.com/mymmrac/telego
go get github.com/golang-jwt/jwt/v5
go get github.com/stretchr/testify
go get github.com/testcontainers/testcontainers-go

# Frontend (npm - run from frontend/)
npm create vite@latest . -- --template react-ts
npm install react@19 react-dom@19
npm install @tanstack/react-query @tanstack/react-router
npm install @tanstack/router-plugin
npm install zustand
npm install sonner
npm install -D typescript vite@8 vitest@4
npm install -D tailwindcss@4 @tailwindcss/vite
npm install -D @testing-library/react msw
npm install -D biome
npx shadcn@latest init
```

## Architecture Patterns

### Recommended Monorepo Structure

```
hub/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # Entry point, wiring
│   ├── internal/
│   │   ├── auth/                    # Auth module (Phase 1)
│   │   │   ├── handler.go           # Fiber route handlers
│   │   │   ├── service.go           # Business logic
│   │   │   ├── repository.go        # Data access
│   │   │   ├── middleware.go        # JWT validation middleware
│   │   │   ├── magiclink.go         # Magic link token logic
│   │   │   ├── telegram.go          # Telegram auth verification
│   │   │   ├── dto.go               # Request/response types
│   │   │   └── events.go            # NATS event definitions
│   │   ├── project/                 # Project module (Phase 1)
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   ├── invite.go            # Invite link logic
│   │   │   ├── dto.go
│   │   │   └── events.go
│   │   ├── rbac/                    # RBAC engine (Phase 1)
│   │   │   ├── middleware.go        # Permission checking middleware
│   │   │   ├── service.go           # Permission computation
│   │   │   ├── permissions.go       # Permission constants/bitmaps
│   │   │   └── cache.go             # Redis permission cache
│   │   ├── storage/                 # File storage (Phase 1)
│   │   │   ├── handler.go           # Upload/download URL endpoints
│   │   │   ├── service.go           # Presigned URL generation
│   │   │   └── cleanup.go           # Orphaned file cleanup cron
│   │   ├── ws/                      # WebSocket gateway (Phase 1)
│   │   │   ├── hub.go               # Connection registry, rooms
│   │   │   ├── handler.go           # Upgrade handler
│   │   │   ├── bridge.go            # NATS-to-WebSocket bridge
│   │   │   └── protocol.go          # WS message types
│   │   └── platform/                # Shared infrastructure
│   │       ├── database/
│   │       │   ├── postgres.go      # Ent client setup
│   │       │   └── mongodb.go       # MongoDB client setup
│   │       ├── nats/
│   │       │   └── client.go        # NATS + JetStream setup
│   │       ├── redis/
│   │       │   └── client.go        # Redis client
│   │       ├── config/
│   │       │   └── config.go        # Env-based configuration
│   │       ├── middleware/
│   │       │   ├── requestid.go     # Request ID generation
│   │       │   ├── logging.go       # Slog request logging
│   │       │   ├── ratelimit.go     # Rate limiting
│   │       │   └── cors.go          # CORS configuration
│   │       ├── errors/
│   │       │   └── errors.go        # Standard error types
│   │       └── outbox/
│   │           ├── publisher.go     # Outbox event writer
│   │           └── worker.go        # Outbox relay worker
│   ├── ent/
│   │   ├── schema/
│   │   │   ├── user.go
│   │   │   ├── session.go
│   │   │   ├── project.go
│   │   │   ├── projectmember.go
│   │   │   ├── role.go
│   │   │   ├── permission.go
│   │   │   ├── team.go
│   │   │   ├── invite.go
│   │   │   ├── filemetadata.go
│   │   │   ├── outboxevent.go
│   │   │   └── systemsetting.go
│   │   └── generate.go
│   ├── migrations/                  # Atlas versioned migrations
│   ├── api/                         # OpenAPI/Swagger specs
│   ├── Makefile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── routes/                  # TanStack Router file-based routes
│   │   │   ├── __root.tsx           # Root layout (sidebar + topbar)
│   │   │   ├── index.tsx            # Home/dashboard redirect
│   │   │   ├── login.tsx            # Login page
│   │   │   ├── onboarding.tsx       # Onboarding wizard
│   │   │   ├── _authenticated.tsx   # Auth guard layout
│   │   │   └── _authenticated/
│   │   │       ├── projects/
│   │   │       │   ├── index.tsx    # Project list
│   │   │       │   ├── new.tsx      # Create project
│   │   │       │   └── $projectId.tsx # Project detail
│   │   │       ├── settings/
│   │   │       │   └── profile.tsx  # User profile
│   │   │       └── trash.tsx        # Trash / deleted items
│   │   ├── components/              # Shared UI components
│   │   │   ├── ui/                  # shadcn/ui generated
│   │   │   ├── layout/
│   │   │   │   ├── sidebar.tsx
│   │   │   │   ├── topbar.tsx
│   │   │   │   └── app-shell.tsx
│   │   │   ├── auth/
│   │   │   │   ├── magic-link-form.tsx
│   │   │   │   └── telegram-button.tsx
│   │   │   └── files/
│   │   │       ├── file-upload.tsx
│   │   │       ├── file-preview.tsx
│   │   │       └── dropzone.tsx
│   │   ├── hooks/
│   │   │   ├── use-auth.ts
│   │   │   ├── use-websocket.ts
│   │   │   └── use-file-upload.ts
│   │   ├── api/
│   │   │   ├── client.ts            # Fetch wrapper with auth
│   │   │   ├── auth.ts              # Auth API calls
│   │   │   ├── projects.ts          # Project API calls
│   │   │   └── files.ts             # File upload/download
│   │   ├── stores/
│   │   │   ├── auth-store.ts        # User, session, permissions
│   │   │   ├── sidebar-store.ts     # Sidebar collapsed state
│   │   │   └── ws-store.ts          # WebSocket connection
│   │   ├── lib/
│   │   │   ├── utils.ts
│   │   │   └── permissions.ts       # Client-side permission helpers
│   │   ├── i18n/
│   │   │   ├── ru.json
│   │   │   └── en.json
│   │   ├── routeTree.gen.ts         # Auto-generated by TanStack Router
│   │   └── main.tsx
│   ├── public/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tailwind.css                 # Tailwind v4 CSS-first config
│   └── package.json
├── cli/                             # Placeholder for hub-cli (Phase 7)
├── deployments/
│   ├── docker-compose.yml           # Full dev stack
│   ├── docker-compose.test.yml      # Test dependencies only
│   ├── Caddyfile                    # Reverse proxy config
│   └── k3s/                         # Production manifests (Phase 7)
├── .github/
│   └── workflows/
│       ├── backend.yml              # Go CI: lint + test + build
│       └── frontend.yml             # Node CI: lint + test + build
├── .gitignore
├── Makefile                         # Top-level orchestration
└── README.md
```

### Pattern 1: JWT Access + Redis Refresh Token

**What:** Short-lived JWT access tokens (15 min, stateless) paired with long-lived refresh tokens stored in Redis (30 days). Refresh tokens are opaque UUIDs mapped to user/session metadata in Redis.

**When to use:** Every authenticated request.

**Example:**
```go
// auth/service.go
type TokenPair struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int64
}

func (s *AuthService) CreateSession(ctx context.Context, userID int, deviceInfo string) (*TokenPair, error) {
    sessionID := uuid.New().String()

    // JWT access token - 15 min, contains userID + sessionID
    accessToken, err := s.signJWT(userID, sessionID, 15*time.Minute)
    if err != nil {
        return nil, err
    }

    // Refresh token - opaque UUID stored in Redis
    refreshToken := uuid.New().String()
    sessionData := SessionData{
        UserID:    userID,
        SessionID: sessionID,
        Device:    deviceInfo,
        CreatedAt: time.Now(),
    }
    // Store in Redis with 30-day TTL
    err = s.redis.Set(ctx,
        "refresh:"+refreshToken,
        sessionData,
        30*24*time.Hour,
    ).Err()
    if err != nil {
        return nil, err
    }

    // Add to user's session set for listing/revocation
    s.redis.SAdd(ctx, "sessions:"+strconv.Itoa(userID), sessionID)

    // Publish session created event for AUTH-07
    s.nats.Publish("auth.session.created", SessionCreatedEvent{
        UserID:    userID,
        SessionID: sessionID,
        Device:    deviceInfo,
    })

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    900, // 15 min in seconds
    }, nil
}
```

### Pattern 2: Outbox Pattern for PG+MongoDB Consistency

**What:** Write domain events to a PostgreSQL outbox table within the same transaction as the primary write. A background worker polls the outbox and applies changes to MongoDB (or publishes to NATS). This ensures at-least-once delivery without distributed transactions.

**When to use:** Any operation that must write to both PostgreSQL and MongoDB (e.g., creating a project creates PG record + MongoDB channel document).

**Example:**
```go
// platform/outbox/publisher.go
func (p *OutboxPublisher) PublishInTx(tx *ent.Tx, eventType string, payload any) error {
    data, _ := json.Marshal(payload)
    _, err := tx.OutboxEvent.Create().
        SetEventType(eventType).
        SetPayload(data).
        SetStatus("pending").
        SetCreatedAt(time.Now()).
        Save(context.Background())
    return err
}

// platform/outbox/worker.go -- runs as background goroutine
func (w *OutboxWorker) ProcessLoop(ctx context.Context) {
    ticker := time.NewTicker(100 * time.Millisecond)
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            events, _ := w.client.OutboxEvent.Query().
                Where(outboxevent.StatusEQ("pending")).
                Order(ent.Asc(outboxevent.FieldCreatedAt)).
                Limit(50).
                All(ctx)
            for _, evt := range events {
                if err := w.nats.Publish(evt.EventType, evt.Payload); err != nil {
                    continue // retry next tick
                }
                evt.Update().SetStatus("published").Save(ctx)
            }
        }
    }
}
```

### Pattern 3: RBAC Permission Model

**What:** Three-tier permission hierarchy: System roles (Admin/Member) > Project roles (Owner/Developer/Tester/Viewer/Custom) > Resource permissions. Permissions stored as bitmaps. Effective permissions precomputed on role change and cached in Redis.

**When to use:** Every API endpoint that accesses project-scoped resources.

**Example:**
```go
// rbac/permissions.go
type Permission uint64

const (
    PermProjectView     Permission = 1 << iota // 1
    PermProjectEdit                             // 2
    PermProjectDelete                           // 4
    PermProjectManageMembers                    // 8
    PermProjectManageRoles                      // 16
    PermTaskCreate                              // 32
    PermTaskEdit                                // 64
    PermTaskDelete                              // 128
    PermTaskAssign                              // 256
    PermFileUpload                              // 512
    PermFileDelete                              // 1024
    // ... up to 64 permissions with uint64
)

// Predefined role permission sets
var RolePermissions = map[string]Permission{
    "owner":     0xFFFFFFFFFFFFFFFF, // all permissions
    "developer": PermProjectView | PermTaskCreate | PermTaskEdit | PermTaskAssign | PermFileUpload,
    "tester":    PermProjectView | PermTaskCreate | PermTaskEdit | PermFileUpload,
    "viewer":    PermProjectView,
}

// rbac/middleware.go
func RequirePermission(perm Permission) fiber.Handler {
    return func(c fiber.Ctx) error {
        userID := auth.UserIDFromCtx(c)
        projectID := c.Params("projectId")

        // Check Redis cache first
        effective, err := rbacCache.GetEffective(c.Context(), userID, projectID)
        if err != nil {
            // Cache miss: compute from DB, cache result
            effective, _ = computeEffectivePermissions(c.Context(), userID, projectID)
            rbacCache.Set(c.Context(), userID, projectID, effective, 5*time.Minute)
        }

        if effective&perm == 0 {
            return c.Status(403).JSON(fiber.Map{"error": "insufficient permissions"})
        }
        return c.Next()
    }
}
```

### Pattern 4: WebSocket Gateway with NATS Bridge

**What:** Single WebSocket connection per client. Fiber WebSocket middleware handles upgrade. Hub manages connections and room subscriptions. Bridge subscribes to NATS subjects and fans out to relevant clients.

**When to use:** All real-time features.

**Example:**
```go
// ws/hub.go
type Hub struct {
    clients    sync.Map // userID -> *Client
    rooms      sync.Map // roomKey -> map[userID]bool
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    UserID    string
    Conn      *websocket.Conn
    Rooms     []string
    SendCh    chan []byte
}

// ws/handler.go -- CRITICAL: copy ctx values before upgrade
func (h *Hub) HandleUpgrade(c fiber.Ctx) error {
    // Copy values from Fiber ctx BEFORE WebSocket upgrade
    userID := utils.CopyString(c.Locals("userID").(string))
    sessionID := utils.CopyString(c.Locals("sessionID").(string))

    return websocket.New(func(conn *websocket.Conn) {
        client := &Client{
            UserID: userID, // safe: copied before upgrade
            Conn:   conn,
            SendCh: make(chan []byte, 256),
        }
        h.register <- client
        defer func() { h.unregister <- client }()

        go client.writePump()
        client.readPump(h)
    })(c)
}

// ws/bridge.go
func (b *Bridge) Start(ctx context.Context) {
    // Subscribe to project-scoped events
    b.nats.Subscribe("project.>", func(msg *nats.Msg) {
        projectID := extractSegment(msg.Subject, 1)
        roomKey := "project:" + projectID
        b.hub.BroadcastToRoom(roomKey, msg.Data)
    })
    // Subscribe to user-specific events (session notifications)
    b.nats.Subscribe("user.>", func(msg *nats.Msg) {
        userID := extractSegment(msg.Subject, 1)
        b.hub.SendToUser(userID, msg.Data)
    })
}
```

### Pattern 5: Presigned URL File Upload/Download

**What:** Backend generates short-lived presigned URLs. Browser uploads/downloads directly to SeaweedFS. Backend only stores metadata.

**When to use:** All file operations -- avatars, attachments, any file.

**Example:**
```go
// storage/service.go
func (s *StorageService) GenerateUploadURL(ctx context.Context, req UploadRequest) (*UploadResponse, error) {
    key := fmt.Sprintf("%s/%s/%s", req.EntityType, req.EntityID, uuid.New().String()+filepath.Ext(req.Filename))

    presignClient := s3.NewPresignClient(s.s3Client)
    presigned, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        ContentType: aws.String(req.ContentType),
    }, s3.WithPresignExpires(5*time.Minute))
    if err != nil {
        return nil, err
    }

    // Save metadata to PG (pending status until confirmed)
    fileMeta, _ := s.entClient.FileMetadata.Create().
        SetKey(key).
        SetFilename(req.Filename).
        SetContentType(req.ContentType).
        SetSize(req.Size).
        SetStatus("pending").
        SetUploadedBy(req.UserID).
        Save(ctx)

    return &UploadResponse{
        FileID:    fileMeta.ID,
        UploadURL: presigned.URL,
    }, nil
}
```

### Anti-Patterns to Avoid

- **Direct module imports:** Never import project/ from auth/ or vice versa. Use NATS events for cross-module communication.
- **Fiber ctx in goroutines:** Never pass `fiber.Ctx` to goroutines. Copy all needed values first. Set `Immutable: true` in dev config.
- **Ent auto-migration in production:** Use Atlas versioned migrations. Auto-migration only in tests.
- **In-memory WebSocket hub without NATS:** Even single-instance must use NATS bridge pattern for future scalability.
- **Proxying files through backend:** Always use presigned URLs. Backend never touches file bytes.
- **String-based permission checks:** Never `if role == "admin"`. Use permission bitmaps.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| JWT creation/validation | Custom JWT parser | golang-jwt/jwt/v5 | Signature validation edge cases, algorithm confusion attacks |
| Email delivery | SMTP client | Resend Go SDK | DKIM/SPF/DMARC handled, deliverability, bounce handling |
| Telegram auth verification | Custom HMAC | telego + Telegram Login Widget | Hash format changes, bot token handling |
| S3 presigned URLs | Custom signature generation | aws-sdk-go-v2 presign client | SigV4 signing is complex, region/service scoping |
| Rate limiting | Custom token bucket | Fiber built-in limiter middleware | Sliding window, Redis-backed for distributed |
| Request ID generation | Custom UUID per request | Fiber requestid middleware | Propagation to logs, standardized header |
| WebSocket upgrade | Raw net/http hijack | Fiber contrib/websocket | Handles fasthttp upgrade correctly, integrates with Fiber middleware |
| CORS handling | Manual headers | Fiber CORS middleware | Preflight handling, credential cookies |
| Password hashing | (N/A - no passwords) | - | Magic Link + Telegram = passwordless by design |
| Database migrations | Manual SQL files | Atlas (from Ent team) | Declarative from Ent schema, diff-based, rollback support |

**Key insight:** Phase 1 is infrastructure -- nearly every component has a well-tested library. The value is in correct wiring and conventions, not custom implementations.

## Common Pitfalls

### Pitfall 1: Fiber RequestCtx Reuse (CRITICAL)
**What goes wrong:** fasthttp reuses RequestCtx objects. Values from `fiber.Ctx` become invalid after handler returns. Goroutines referencing ctx get corrupted data or panic.
**Why it happens:** Zero-allocation performance optimization. Silent until load increases.
**How to avoid:**
- `utils.CopyString()` for all params/headers/query values used outside handler
- Copy body bytes: `body := make([]byte, len(c.Body())); copy(body, c.Body())`
- Set `app.Config{Immutable: true}` in development
- Run all tests with `-race` flag
- Extract all values BEFORE WebSocket upgrade in ws/handler.go
**Warning signs:** Intermittent wrong data in responses, data race warnings, panics in background goroutines.

### Pitfall 2: PG+MongoDB Cross-Database Inconsistency
**What goes wrong:** Creating project writes PG record but MongoDB channel creation fails. Orphaned entities with no cleanup path.
**Why it happens:** No distributed transactions. "It rarely fails" until it does.
**How to avoid:** Outbox pattern (see Architecture Patterns). PG is source of truth. MongoDB is derived/eventual. Reconciliation cron for safety.
**Warning signs:** Orphaned channels, missing data, inconsistent counts.

### Pitfall 3: RBAC Permission Scattering
**What goes wrong:** Permission checks scattered across handlers as `if role == "admin"`. Inconsistent enforcement, security holes.
**Why it happens:** Incremental development adds checks ad-hoc instead of centrally.
**How to avoid:** Centralized RBAC middleware with permission bitmaps. Ent Privacy Layer as second defense. Never hardcode role names in business logic.
**Warning signs:** Different endpoints checking permissions differently, users seeing data they shouldn't.

### Pitfall 4: NATS JetStream Overuse
**What goes wrong:** Using JetStream consumers for everything (real-time chat, presence, typing indicators) leads to consumer proliferation.
**Why it happens:** JetStream feels like the "right" abstraction for guaranteed delivery.
**How to avoid:** Core NATS for ephemeral real-time (typing, presence, WebSocket events). JetStream ONLY for durable jobs (outbox relay, audit, scheduled cleanup).
**Warning signs:** Consumer count growing with user count, NATS memory climbing.

### Pitfall 5: SeaweedFS Presigned URL with Internal Docker Hostname
**What goes wrong:** Presigned URLs generated with internal hostname (e.g., `seaweedfs:8333`) that browsers cannot reach.
**Why it happens:** S3 client configured with Docker service name for backend access.
**How to avoid:** Configure separate internal endpoint (for backend) and external endpoint (for presigned URLs). External endpoint goes through Caddy.
**Warning signs:** Upload URLs returning network errors in browser, CORS failures.

### Pitfall 6: Magic Link Token Security
**What goes wrong:** Long TTL tokens, no rate limiting, tokens readable in Telegram message history.
**Why it happens:** Development convenience, skipping security hardening.
**How to avoid:** 5-minute TTL in Redis, single-use (delete on first use), rate limit 3 requests per email per 10 minutes. For Telegram: use inline keyboards with callback queries, delete token message after use.
**Warning signs:** Token reuse working, no rate limit errors during testing.

### Pitfall 7: Ent N+1 Queries
**What goes wrong:** Loading project list with members, roles, and settings generates hundreds of queries.
**Why it happens:** Ent lazy-loads edges by default. Must explicitly eager-load with `.With*()`.
**How to avoid:** Always use `.WithMembers()`, `.WithRoles()` etc. Enable query logging in development. Profile all list endpoints.
**Warning signs:** Slow list pages, high query count in logs.

## Code Examples

### Docker Compose Dev Environment
```yaml
# deployments/docker-compose.yml
services:
  postgres:
    image: postgres:17
    environment:
      POSTGRES_DB: hub
      POSTGRES_USER: hub
      POSTGRES_PASSWORD: hub_dev
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data

  mongodb:
    image: mongo:8
    environment:
      MONGO_INITDB_ROOT_USERNAME: hub
      MONGO_INITDB_ROOT_PASSWORD: hub_dev
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db

  redis:
    image: redis:7.4-alpine
    ports:
      - "6379:6379"

  nats:
    image: nats:2.11-alpine
    command: ["-js", "-sd", "/data"]
    ports:
      - "4222:4222"
      - "8222:8222"  # monitoring
    volumes:
      - nats_data:/data

  seaweedfs:
    image: chrislusf/seaweedfs
    command: "server -s3 -dir=/data"
    ports:
      - "8333:8333"  # S3 API
      - "9333:9333"  # master
    volumes:
      - seaweed_data:/data

  caddy:
    image: caddy:2.11-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile

volumes:
  pg_data:
  mongo_data:
  nats_data:
  seaweed_data:
```

### Caddyfile for Development
```
:80 {
    # API
    handle /api/* {
        reverse_proxy localhost:3000
    }

    # WebSocket
    handle /ws {
        reverse_proxy localhost:3000
    }

    # SeaweedFS (for presigned URL access from browser)
    handle /storage/* {
        reverse_proxy seaweedfs:8333
    }

    # Frontend (Vite dev server)
    handle {
        reverse_proxy localhost:5173
    }
}
```

### Ent User Schema
```go
// ent/schema/user.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/index"
)

type User struct {
    ent.Schema
}

func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("email").Unique().NotEmpty(),
        field.String("name").Optional(),
        field.String("avatar_url").Optional(),
        field.String("telegram_id").Optional().Unique(),
        field.String("telegram_username").Optional(),
        field.String("position").Optional(),
        field.String("department").Optional(),
        field.String("phone").Optional(),
        field.String("timezone").Default("UTC"),
        field.String("language").Default("en"),
        field.Time("birthday").Optional().Nillable(),
        field.Enum("system_role").Values("admin", "member").Default("member"),
        field.Bool("onboarding_completed").Default(false),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
        field.Time("deleted_at").Optional().Nillable(),
    }
}

func (User) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("project_memberships", ProjectMember.Type),
        edge.To("sessions", Session.Type),
        edge.To("uploaded_files", FileMetadata.Type),
    }
}

func (User) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("email"),
        index.Fields("telegram_id"),
        index.Fields("deleted_at"),
    }
}
```

### Fiber App Setup with Conventions
```go
// cmd/server/main.go
func setupApp(cfg *config.Config) *fiber.App {
    app := fiber.New(fiber.Config{
        // CRITICAL: Enable in development to catch RequestCtx reuse bugs
        Immutable: cfg.IsDev(),
        ErrorHandler: platform.ErrorHandler,
    })

    // Global middleware
    app.Use(middleware.RequestID())
    app.Use(middleware.Logger(cfg.Logger))
    app.Use(middleware.CORS(cfg.CORSOrigins))
    app.Use(middleware.RateLimiter(cfg.Redis, 100, time.Second)) // 100 req/sec global

    // Auth routes (no auth middleware)
    authGroup := app.Group("/api/auth")
    authGroup.Post("/magic-link", ratelimit(3, 10*time.Minute), authHandler.SendMagicLink)
    authGroup.Post("/magic-link/verify", authHandler.VerifyMagicLink)
    authGroup.Post("/telegram", authHandler.TelegramLogin)
    authGroup.Post("/refresh", authHandler.RefreshToken)
    authGroup.Post("/logout", authMiddleware.Required(), authHandler.Logout)

    // Protected routes
    api := app.Group("/api", authMiddleware.Required())

    // User profile
    api.Get("/me", authHandler.GetProfile)
    api.Put("/me", authHandler.UpdateProfile)
    api.Get("/me/sessions", authHandler.ListSessions)
    api.Delete("/me/sessions/:sessionId", authHandler.RevokeSession)

    // Projects
    projects := api.Group("/projects")
    projects.Get("/", projectHandler.List)
    projects.Post("/", projectHandler.Create)
    projects.Get("/:projectId", rbac.Require(PermProjectView), projectHandler.Get)
    projects.Put("/:projectId", rbac.Require(PermProjectEdit), projectHandler.Update)
    projects.Post("/:projectId/archive", rbac.Require(PermProjectEdit), projectHandler.Archive)
    projects.Delete("/:projectId", rbac.Require(PermProjectDelete), projectHandler.Delete)
    projects.Post("/:projectId/members", rbac.Require(PermProjectManageMembers), projectHandler.AddMember)
    projects.Delete("/:projectId/members/:userId", rbac.Require(PermProjectManageMembers), projectHandler.RemoveMember)
    projects.Get("/:projectId/roles", rbac.Require(PermProjectView), projectHandler.ListRoles)
    projects.Post("/:projectId/roles", rbac.Require(PermProjectManageRoles), projectHandler.CreateRole)

    // File storage
    files := api.Group("/files")
    files.Post("/upload-url", storageHandler.GenerateUploadURL)
    files.Post("/:fileId/confirm", storageHandler.ConfirmUpload)
    files.Get("/:fileId/url", storageHandler.GenerateDownloadURL)

    // WebSocket
    app.Use("/ws", authMiddleware.RequiredWS()) // validate JWT before upgrade
    app.Get("/ws", websocket.New(wsHub.HandleConnection))

    return app
}
```

### TanStack Router Vite Config
```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { TanStackRouterVite } from '@tanstack/router-plugin/vite'

export default defineConfig({
  plugins: [
    TanStackRouterVite({
      target: 'react',
      autoCodeSplitting: true,
    }),
    react(),
    tailwindcss(),
  ],
  server: {
    proxy: {
      '/api': 'http://localhost:3000',
      '/ws': { target: 'ws://localhost:3000', ws: true },
      '/storage': 'http://localhost:8333',
    },
  },
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `tailwind.config.js` | CSS `@theme` directive | Tailwind v4 (2025) | No JS config, CSS-first setup |
| framer-motion package | `motion` package | 2025 rename | Import from `motion/react` |
| MinIO for S3 storage | SeaweedFS | Feb 2026 | MinIO archived, SeaweedFS is drop-in replacement |
| dnd-kit for drag-and-drop | pragmatic-drag-and-drop | 2024 (dnd-kit abandoned) | Atlassian-backed, tree adapter |
| React Router v6 | TanStack Router v1 | 2024-2025 adoption | Type-safe routes, file-based, auto code-split |
| Fiber v2 `UserContext()` | Fiber v3 native `context.Context` on Ctx | 2025-2026 | No more bridging, direct context access |
| Ent auto-migration | Atlas versioned migrations | Atlas GA 2024 | Declarative, diff-based, rollback support |
| GORM for Go ORMs | Ent (for graph-heavy models) | Growing 2024-2026 | Privacy layer, codegen, graph traversal |
| Redis pub/sub + RabbitMQ | NATS (core + JetStream) | Growing 2024-2026 | Single system for both real-time and durable |
| minio-go SDK | aws-sdk-go-v2 | Feb 2026 (MinIO death) | Standard S3, provider-agnostic |

**Deprecated/outdated:**
- **MinIO:** Archived Feb 2026, replaced by SeaweedFS
- **dnd-kit:** Abandoned by maintainer since 2023
- **framer-motion package name:** Renamed to `motion`
- **tailwind.config.js:** Replaced by CSS-first config in v4
- **Fiber v2:** v3 is current, adds context.Context interface

## Open Questions

1. **Telegram Login Widget vs Bot Commands for auth**
   - What we know: Telegram Login Widget uses an HTML widget that verifies via HMAC-SHA256. Bot commands require user to message the bot.
   - What's unclear: Whether Login Widget works well embedded in React SPA (it injects a script tag). May need a popup flow.
   - Recommendation: Start with Login Widget (better UX for web), fall back to bot deep-link if widget integration proves problematic.

2. **SeaweedFS CORS configuration for presigned uploads**
   - What we know: SeaweedFS supports S3 API including presigned URLs. aws-sdk-go-v2 works with custom endpoints.
   - What's unclear: Exact CORS configuration method for SeaweedFS S3 gateway (different from MinIO's bucket policy). Documentation is sparse.
   - Recommendation: Test CORS in Docker Compose early. May need Caddy to proxy uploads if SeaweedFS CORS is limited.

3. **Ent Privacy Layer performance for high-frequency permission checks**
   - What we know: Privacy Layer evaluates rules on every query. This adds overhead per DB call.
   - What's unclear: Whether the overhead is acceptable for list endpoints returning 100+ entities with privacy checks.
   - Recommendation: Use Privacy Layer for entity-level security, but rely on middleware RBAC (with Redis cache) for route-level checks. Profile and optimize if Privacy Layer adds >50ms to list queries.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework (backend) | Go testing + testify 1.10.x + testcontainers-go 0.36.x |
| Framework (frontend) | Vitest 4.1.x + @testing-library/react 16.x |
| Config file (backend) | None -- Go uses convention (files ending _test.go) |
| Config file (frontend) | `frontend/vitest.config.ts` (extends vite.config.ts) |
| Quick run (backend) | `cd backend && go test ./internal/auth/... -race -count=1` |
| Quick run (frontend) | `cd frontend && npx vitest run --reporter=verbose` |
| Full suite (backend) | `cd backend && go test ./... -race -count=1 -timeout=5m` |
| Full suite (frontend) | `cd frontend && npx vitest run` |

### Phase Requirements -> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| AUTH-01 | Magic Link send + verify flow | integration | `go test ./internal/auth/... -run TestMagicLink -race` | Wave 0 |
| AUTH-02 | Telegram login hash verification | unit | `go test ./internal/auth/... -run TestTelegramAuth -race` | Wave 0 |
| AUTH-03 | Registration mode enforcement | unit | `go test ./internal/auth/... -run TestRegistrationMode -race` | Wave 0 |
| AUTH-04 | Profile CRUD + avatar upload | integration | `go test ./internal/auth/... -run TestProfile -race` | Wave 0 |
| AUTH-05 | System role enforcement | unit | `go test ./internal/rbac/... -run TestSystemRoles -race` | Wave 0 |
| AUTH-06 | Multi-session + revocation | integration | `go test ./internal/auth/... -run TestSessions -race` | Wave 0 |
| AUTH-07 | New session notification via WS | integration | `go test ./internal/ws/... -run TestSessionNotification -race` | Wave 0 |
| AUTH-08 | Onboarding flow | frontend | `npx vitest run src/routes/onboarding.test.tsx` | Wave 0 |
| PROJ-01 | Project creation with auto-resources | integration | `go test ./internal/project/... -run TestProjectCreate -race` | Wave 0 |
| PROJ-02 | Owner manages roles/access | integration | `go test ./internal/project/... -run TestOwnerManagement -race` | Wave 0 |
| PROJ-03 | Fixed + custom RBAC roles | unit | `go test ./internal/rbac/... -run TestProjectRoles -race` | Wave 0 |
| PROJ-04 | Invite link + team add | integration | `go test ./internal/project/... -run TestInvite -race` | Wave 0 |
| PROJ-06 | Project archiving (read-only) | integration | `go test ./internal/project/... -run TestArchive -race` | Wave 0 |
| PROJ-07 | Soft-delete + 30d trash | integration | `go test ./internal/project/... -run TestSoftDelete -race` | Wave 0 |
| UI-01 | Sidebar + topbar layout | frontend | `npx vitest run src/components/layout/` | Wave 0 |
| UI-07 | Error handling (toast, 404/500) | frontend | `npx vitest run src/components/error/` | Wave 0 |
| UI-08 | File upload dropzone | frontend | `npx vitest run src/components/files/` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/{module}/... -race -count=1` (backend) or `npx vitest run src/{module}/` (frontend)
- **Per wave merge:** `go test ./... -race -count=1 && cd ../frontend && npx vitest run`
- **Phase gate:** Full backend + frontend suite green before verify

### Wave 0 Gaps
- [ ] `backend/internal/auth/auth_test.go` -- covers AUTH-01 through AUTH-07
- [ ] `backend/internal/auth/magiclink_test.go` -- covers AUTH-01 specifically
- [ ] `backend/internal/auth/telegram_test.go` -- covers AUTH-02 specifically
- [ ] `backend/internal/rbac/rbac_test.go` -- covers AUTH-05, PROJ-03
- [ ] `backend/internal/project/project_test.go` -- covers PROJ-01, PROJ-02, PROJ-04, PROJ-06, PROJ-07
- [ ] `backend/internal/ws/ws_test.go` -- covers AUTH-07
- [ ] `backend/internal/storage/storage_test.go` -- covers UI-08 (backend side)
- [ ] `backend/internal/platform/outbox/outbox_test.go` -- covers PG+MongoDB consistency
- [ ] `frontend/vitest.config.ts` -- Vitest configuration
- [ ] `frontend/src/test/setup.ts` -- test setup with MSW
- [ ] Go test infrastructure: `go get github.com/stretchr/testify && go get github.com/testcontainers/testcontainers-go`
- [ ] Frontend test infrastructure: `npm install -D vitest @testing-library/react msw`

## Sources

### Primary (HIGH confidence)
- [Fiber v3 What's New](https://docs.gofiber.io/next/whats_new/) -- RequestCtx changes, Go Context interface, Immutable config
- [Fiber v3 WebSocket contrib](https://github.com/gofiber/contrib/tree/main/v3/websocket) -- WebSocket middleware for Fiber v3
- [Fiber JWT middleware](https://docs.gofiber.io/contrib/jwt/) -- JWT validation middleware
- [Ent Privacy Layer](https://entgo.io/docs/privacy/) -- Privacy rules for RBAC
- [Ent multi-tenant example](https://atlasgo.io/blog/2025/05/26/gophercon-scalable-multi-tenant-apps-in-go) -- Privacy layer with Atlas
- [NATS JetStream docs](https://docs.nats.io/nats-concepts/jetstream) -- Core vs JetStream patterns
- [Synadia JetStream anti-patterns](https://www.synadia.com/blog/jetstream-design-patterns-for-scale) -- Consumer proliferation warning
- [SeaweedFS S3 API](https://github.com/seaweedfs/seaweedfs) -- S3 compatibility, presigned URLs
- [aws-sdk-go-v2 S3 presign](https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html) -- Presigned URL generation
- [Resend Go SDK](https://resend.com/docs/send-with-go) -- Email sending API
- [telego](https://github.com/mymmrac/telego) -- Telegram Bot API library
- [Telegram Login Widget](https://core.telegram.org/widgets/login) -- Web auth widget
- [TanStack Router Vite setup](https://tanstack.com/router/latest/docs/installation/with-vite) -- File-based routing config

### Secondary (MEDIUM confidence)
- [JWT + Redis refresh pattern](https://codevoweb.com/how-to-properly-use-jwt-for-authentication-in-golang/) -- Access/refresh token architecture
- [NATS WebSocket bridge patterns](https://dasroot.net/posts/2026/02/building-event-driven-systems-go-nats/) -- Event-driven Go + NATS
- [TanStack Router SaaS template 2026](https://dev.to/kiran_ravi_092a2cfcf60389/tanstack-router-setup-in-our-react-saas-template-2026-4b67) -- Modern React routing setup
- [Ent ORM comparison 2026](https://encore.cloud/resources/go-orms) -- Ent vs alternatives assessment

### Tertiary (LOW confidence)
- SeaweedFS CORS configuration -- documentation sparse, needs hands-on validation
- Ent Privacy Layer performance at scale -- no benchmarks found, needs profiling
- Telegram Login Widget in React SPA -- few examples, may need popup approach

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all libraries verified via official docs and release notes, versions confirmed current
- Architecture: HIGH -- modular monolith + NATS bridge is well-documented pattern, project structure follows Go conventions
- Pitfalls: HIGH -- Fiber ctx reuse, outbox pattern, RBAC explosion, NATS consumer anti-patterns all confirmed by official sources
- Auth patterns: MEDIUM-HIGH -- JWT + Redis refresh pattern is standard but Telegram Login Widget in React SPA needs validation
- SeaweedFS integration: MEDIUM -- S3 API compatibility confirmed but CORS and presigned URL behavior needs hands-on testing

**Research date:** 2026-03-20
**Valid until:** 2026-04-20 (30 days -- stable technologies, no fast-moving changes expected)
