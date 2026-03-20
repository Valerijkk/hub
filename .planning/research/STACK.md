# Stack Research

**Domain:** Team collaboration hub (task tracker + messenger + TMS + voice calls)
**Researched:** 2026-03-20
**Confidence:** HIGH (all core technologies verified via official sources)

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.x | Backend language | Goroutines for real-time (messenger, calls, notifications). Compiles to single binary. First-class concurrency for 50+ simultaneous WebSocket connections per server. |
| Fiber | v3.x | HTTP framework | Express-style API, built on fasthttp (fastest Go HTTP engine). v3 adds `context.Context` interface on `fiber.Ctx`, eliminating v2's `UserContext()` bridging. Requires Go 1.25+. CLI migration tool available. |
| Ent | v0.14.x | ORM (PostgreSQL) | Code-generated type-safe queries, graph-based entity traversal (critical for task hierarchies and parent/child relations). Privacy layer maps directly to RBAC. Atlas migrations for schema versioning. |
| React | 19.2.x | Frontend framework | Stable since Dec 2024, 19.2.4 is latest (Jan 2026). Server Components stable but not needed here (SPA). useTransition and useOptimistic useful for optimistic UI in task/chat updates. |
| Vite | 8.x | Build tool | Ships Rolldown (Rust bundler): 10-30x faster builds vs Vite 7/webpack. Single unified bundler. Oxc integration for faster transforms. |
| TypeScript | 5.7.x | Frontend type safety | Strict mode mandatory. Pairs with Ent codegen for end-to-end type safety (generate TypeScript types from Ent schemas via openapi). |
| PostgreSQL | 17.x | Primary database | Relational data: users, projects, tasks, TMS, sprints, RBAC. JSONB for flexible custom fields. Full-text search for JQL-like queries. Row-level security possible for multi-project isolation. |
| MongoDB | 8.x | Document database | Chat messages (append-heavy, denormalized), Allure report JSON blobs, audit logs (write-heavy, time-series). Go driver v2.5.x. |
| Redis | 7.4.x | Cache + sessions | Session store, magic link tokens (TTL), online/offline status, rate limiting. NOT for pub/sub (NATS handles that). |
| NATS + JetStream | Server 2.11.x, Go client 1.49.x | Messaging + event bus | Replaces both RabbitMQ (task queues) and Redis pub/sub (real-time). JetStream adds persistence for at-least-once delivery (notifications, Allure results ingestion). 10MB RAM footprint. Go-native. |
| LiveKit | Server 1.8.x | Voice calls + screen share | Go-native WebRTC SFU. Self-hosted single container. Egress service for call recording. React SDK (@livekit/components-react 2.9.x) for frontend. Up to 50 participants per room. |
| Caddy | 2.11.x | Reverse proxy | Auto-HTTPS with Let's Encrypt, native WebSocket proxying, ~15 lines config. Post-quantum TLS support. Simpler than Nginx for this use case. |

### Database Drivers & Data Layer

| Library | Version | Purpose | Notes |
|---------|---------|---------|-------|
| entgo.io/ent | 0.14.x | PostgreSQL ORM | `go get entgo.io/ent`. Atlas for migrations (`ariga.io/atlas`). |
| go.mongodb.org/mongo-driver/v2 | 2.5.x | MongoDB driver | v2 is current line; v1.17 is final v1. Use v2 for new projects. |
| github.com/redis/go-redis/v9 | 9.x | Redis client | Supports Redis 7.x features, context-aware, pipelining. |
| github.com/nats-io/nats.go | 1.49.x | NATS client | JetStream API in `github.com/nats-io/nats.go/jetstream` subpackage. |
| github.com/livekit/server-sdk-go/v2 | 2.x | LiveKit server SDK | Room management, token generation, webhook verification. |

### Object Storage -- CRITICAL CHANGE

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| **SeaweedFS** | 3.x | S3-compatible object storage | **MinIO is DEAD** (archived Feb 13, 2026). SeaweedFS is the recommended replacement: Apache 2.0 license, production-proven since 2016, O(1) disk reads, K8s support, S3 API compatible. Presigned URLs work the same as MinIO. |

**Why NOT MinIO:** Repository archived Feb 2026. No new features, no PRs accepted. Company pushed everyone to paid AIStor. Community edition is dead.

**Why SeaweedFS over alternatives:**
- **vs Garage:** SeaweedFS has better throughput for file-heavy workloads (attachments, recordings, Allure reports). Garage is better for geo-distributed setups but overkill for single-cluster K3s.
- **vs RustFS:** Still in alpha (v1.0.0-alpha). Not production-ready. Impressive benchmarks but too early.
- **vs Ceph:** Massive operational overhead. Designed for petabyte scale, not a 10-50 person team.

**Migration from MinIO code:** S3 API compatible. Change endpoint URL and credentials. No code changes needed if using standard S3 SDK (`aws-sdk-go-v2`).

### Frontend UI & Components

| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| shadcn/ui | CLI v4 | Component system | Copy-paste components (full control), Radix UI primitives, Tailwind CSS native. CLI v4 (March 2026) adds shadcn/skills for AI coding agents. Now supports choosing between Radix and Base UI engines. |
| Tailwind CSS | 4.2.x | Styling | v4 is a full rewrite: CSS-first config, faster builds, new color palettes. No more `tailwind.config.js` -- uses CSS `@theme` directive. |
| Plate | 49.x | Rich-text editor | Shadcn-native (uses same Radix primitives). Built on Slate.js. @mentions, slash commands, tables -- all free (Tiptap charges for tables). Headless + styled components. |
| @atlaskit/pragmatic-drag-and-drop | 1.7.x | Drag and drop | From Atlassian (powers Jira/Trello). Tree adapter for task hierarchies. Kanban board columns. Zero unnecessary re-renders. Framework-agnostic core with React optional package. |
| motion (ex framer-motion) | 12.x | Animations | Renamed from `framer-motion` to `motion`. Import from `motion/react`. Layout animations for Kanban card moves, list reordering, modal transitions. |
| Tremor | npm 3.18.x / copy-paste | Dashboard charts | Acquired by Vercel, now MIT licensed. Tailwind-native. 35+ components for KPIs, charts, metrics. Use copy-paste approach (like shadcn) for maximum control. Built on Recharts internally. |
| @livekit/components-react | 2.9.x | Voice call UI | Pre-built React components for LiveKit rooms: participant tiles, audio controls, screen share. Tree-shakable. |
| TanStack Query | 5.x | Server state | Cache management for API calls. Optimistic updates for task mutations. Automatic background refetching. Pairs with REST API. |
| TanStack Router | 1.x | Routing | Type-safe routing with search params. File-based route generation. Better than React Router for type safety. |
| Zustand | 5.x | Client state | Lightweight global state for UI state (sidebar open, active filters, theme). NOT for server data (TanStack Query handles that). |
| react-virtual (TanStack Virtual) | 3.x | Virtualization | Virtual scrolling for chat message lists, large task tables, Kanban boards with many cards. Critical for performance at scale. |

### Telegram Bot

| Library | Version | Purpose | Notes |
|---------|---------|---------|-------|
| github.com/mymmrac/telego | 1.0.x | Telegram Bot API | 100% Bot API coverage via codegen. Uses fasthttp by default. Type-safe. Middleware support. Handles magic link auth + notifications. |

### Infrastructure & DevOps

| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| Docker | 27.x | Containerization | Multi-stage builds for Go (scratch/distroless final image ~15MB). |
| K3s | 1.35.x (v1.35.2+k3s1) | Production orchestration | Lightweight K8s (<70MB binary, ~500MB RAM). Containerd v2.1.5. Go 1.25.7 runtime. |
| Docker Compose | 2.x | Local development | Full stack locally: PG, Mongo, Redis, NATS, SeaweedFS, LiveKit, Caddy. |
| GitHub Actions | N/A | CI/CD | Student pack ~3000 min. Self-hosted runner as fallback. |

### Monitoring & Observability

| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| log/slog | stdlib | Structured logging | Go stdlib since 1.21. JSON output. Context-aware. No external dependency. |
| Grafana Loki | 3.x | Log aggregation | Pairs with slog JSON. Lightweight vs ELK. Same query lang as Prometheus. |
| Prometheus | 2.x | Metrics | Scrapes Go runtime + custom metrics. Fiber middleware exposes /metrics. |
| Grafana | 11.x | Dashboards | Unified UI for logs (Loki) + metrics (Prometheus). Admin health dashboard. |

### Testing

| Technology | Version | Purpose | Notes |
|------------|---------|---------|-------|
| testing (stdlib) | Go stdlib | Unit tests | Go built-in. Table-driven tests pattern. |
| testify | 1.10.x | Assertions + mocks | `assert`, `require`, `suite` packages. De facto standard. |
| testcontainers-go | 0.36.x | Integration tests | Spin up real PG, Mongo, Redis, NATS in Docker for tests. |
| mockery | 2.x | Mock generation | Generates mocks from interfaces. Pairs with testify. |
| Vitest | 4.1.x | Frontend unit tests | Browser mode stable in v4. Vite-native (shares config). 10x faster than Jest. |
| @testing-library/react | 16.x | Component testing | DOM-based testing. Tests behavior, not implementation. |
| Playwright | 1.58.x | E2E tests | Cross-browser. HTML report with timeline. Chrome for Testing builds. |
| MSW | 2.x | API mocking | Service Worker-based. Intercepts fetch/XHR. Shared handlers between tests and Storybook. |

## Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| Air | Hot reload (Go) | File watcher, rebuilds on save. Essential for DX. |
| golangci-lint | Go linting | Runs 50+ linters in parallel. Configure via `.golangci.yml`. |
| ESLint + Biome | Frontend linting | ESLint for React rules, Biome for formatting (faster than Prettier). |
| OpenAPI Generator | API types | Generate TypeScript client from Go-served OpenAPI spec. End-to-end type safety. |
| Atlas | DB migrations | From Ent team. Declarative migrations from Ent schemas. Versioned migration files. |

## Installation

```bash
# Backend (Go modules)
go get github.com/gofiber/fiber/v3
go get entgo.io/ent
go get ariga.io/atlas
go get go.mongodb.org/mongo-driver/v2
go get github.com/redis/go-redis/v9
go get github.com/nats-io/nats.go
go get github.com/livekit/server-sdk-go/v2
go get github.com/mymmrac/telego
go get github.com/stretchr/testify

# Frontend
npm install react@19.2 react-dom@19.2
npm install @tanstack/react-query @tanstack/react-router zustand
npm install @livekit/components-react livekit-client
npm install motion @tanstack/react-virtual
npm install @udecode/plate

# Frontend dev
npm install -D typescript vite@8 vitest@4 @playwright/test
npm install -D tailwindcss@4 @tailwindcss/vite
npm install -D msw @testing-library/react
npx shadcn@latest init
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Fiber v3 | Chi/Echo | Chi if you want stdlib `net/http` compatibility (easier middleware ecosystem). Echo if team knows it already. Fiber's fasthttp means some stdlib middleware won't work directly. |
| Ent | GORM | GORM if team is familiar and project is CRUD-only. Ent wins for graph queries (task trees), Privacy layer (RBAC), and codegen. GORM has larger ecosystem but weaker type safety. |
| Ent | sqlc | sqlc if you want raw SQL control and maximum performance. Loses graph traversal and codegen convenience. Better for read-heavy microservices. |
| PostgreSQL+MongoDB | PostgreSQL-only | PG-only if you store chat messages in PG with JSONB. Simpler ops. MongoDB justified here by Allure JSON blobs and append-heavy chat patterns. |
| NATS | RabbitMQ | RabbitMQ if you need complex routing (topic exchanges, dead letter queues). NATS is simpler and lighter for pub/sub + basic queues. |
| SeaweedFS | Garage | Garage if deploying across multiple geographic locations. SeaweedFS better for single-cluster with high file throughput. |
| TanStack Router | React Router 7 | React Router if team knows it. TanStack Router has better TypeScript inference for search params and loaders. |
| Tremor | Recharts | Recharts if you only need basic charts without KPI cards and dashboard layouts. Tremor is a dashboard framework, not just a chart library. |
| Plate | Tiptap | Tiptap if you don't use shadcn/ui. Tiptap has larger ecosystem but tables/mentions are paid extensions. Plate is fully free. |
| motion | None | motion (ex framer-motion) is the standard. No real alternative at this level of polish for React layout animations. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| MinIO | Archived Feb 2026. Dead project. No security patches. | SeaweedFS |
| framer-motion (package name) | Renamed to `motion`. Old package still works but deprecated. | `motion` (import from `motion/react`) |
| dnd-kit | Abandoned by maintainer since 2023. No updates. | @atlaskit/pragmatic-drag-and-drop |
| GORM | Implicit magic, poor type safety, N+1 query traps. | Ent (for this project's graph-heavy data model) |
| Ant Design / MUI | Heavyweight, opinionated styling, hard to customize with Tailwind. | shadcn/ui (copy-paste, Tailwind-native) |
| Jest | Slow, requires babel transforms, doesn't share Vite config. | Vitest (Vite-native, 10x faster) |
| Webpack | Superseded by Vite 8 with Rolldown. No reason for new projects. | Vite 8 |
| `tailwind.config.js` | Deprecated in Tailwind v4. CSS-first configuration now. | `@theme` directive in CSS |
| Jitsi | Java-based, heavy (2GB+ RAM), complex config. | LiveKit (Go-native, single container, 200MB RAM) |
| Socket.IO | Not needed. NATS handles pub/sub. Fiber has native WebSocket. | Fiber WebSocket + NATS pub/sub |
| Nginx | More complex config for same result. No auto-HTTPS. | Caddy |
| go-telegram-bot-api | Incomplete Bot API coverage, manual type definitions. | telego (100% codegen coverage) |

## Stack Patterns

**For real-time features (chat, notifications, online status):**
- Client connects via WebSocket to Fiber
- Fiber subscribes to NATS subjects per user/channel
- NATS fans out events across server instances
- JetStream persists messages for offline delivery

**For file uploads (attachments, recordings, avatars):**
- Client gets presigned upload URL from Go API
- Client uploads directly to SeaweedFS (S3 API)
- Go API stores metadata in PostgreSQL
- Frontend uses presigned download URLs for display

**For voice calls:**
- Client gets LiveKit token from Go API (server-sdk-go)
- Client connects directly to LiveKit server (WebRTC)
- LiveKit Egress records to SeaweedFS (S3 API)
- Call metadata stored in PostgreSQL

**For Allure integration:**
- CI generates allure-results JSON
- CLI tool POSTs results to Hub API
- Go API parses JSON, matches test case IDs
- Results stored in MongoDB (raw JSON) + PostgreSQL (structured metadata)
- NATS publishes notification events

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| Fiber v3 | Go >= 1.25 | v3 requires Go 1.25+. Use Go 1.26 for latest features. |
| Ent | Go >= 1.22 | Works with Go 1.26. Atlas CLI needed for migrations. |
| Vite 8 | React 19.2.x | Rolldown bundler. Vitest 4.x compatible. |
| Tailwind CSS 4.2 | Vite 8 | Use `@tailwindcss/vite` plugin. No PostCSS config needed. |
| shadcn/ui CLI v4 | Tailwind CSS 4.x, React 19.x | Generates components using Tailwind v4 syntax. |
| Plate 49 | React 19.x, shadcn/ui | Uses same Radix UI primitives as shadcn. |
| LiveKit components-react 2.9.x | React 19.x, livekit-client 2.x | Requires livekit-client as peer dependency. |
| Vitest 4.x | Vite 8.x | Shares Vite config. Browser mode stable. |
| TanStack Query 5 | React 19.x | Supports React 19 `use()` hook. |
| motion 12.x | React 19.x | Import from `motion/react`. |

## Sources

- [Go 1.26 release](https://go.dev/blog/go1.26) -- Go version verified
- [Fiber v3 docs](https://docs.gofiber.io/next/whats_new/) -- v3 breaking changes, Go 1.25+ requirement
- [Fiber v3 release](https://github.com/gofiber/fiber/releases/tag/v3.0.0) -- Release notes
- [Ent official site](https://entgo.io/) -- ORM features, Atlas integration
- [React 19.2 blog](https://react.dev/blog/2025/10/01/react-19-2) -- Stable features confirmed
- [Vite 8 announcement](https://voidzero.dev/posts/announcing-vite-plus) -- Rolldown integration
- [Tailwind CSS v4.2](https://tailwindcss.com/blog/tailwindcss-v4) -- CSS-first config
- [shadcn/ui CLI v4](https://ui.shadcn.com/docs/changelog/2026-03-cli-v4) -- Latest CLI, Radix/Base UI choice
- [Plate editor](https://platejs.org/) -- Shadcn-native rich text
- [pragmatic-drag-and-drop](https://atlassian.design/components/pragmatic-drag-and-drop/) -- Atlassian DnD
- [motion upgrade guide](https://motion.dev/docs/react-upgrade-guide) -- framer-motion renamed to motion
- [Tremor acquisition by Vercel](https://vercel.com/blog/vercel-acquires-tremor) -- MIT license, active
- [LiveKit self-hosting](https://docs.livekit.io/transport/self-hosting/) -- Deployment docs
- [LiveKit components-react npm](https://www.npmjs.com/package/@livekit/components-react) -- v2.9.20
- [NATS releases](https://github.com/nats-io/nats-server/releases) -- Server versions
- [nats.go package](https://pkg.go.dev/github.com/nats-io/nats.go/jetstream) -- JetStream Go API
- [MinIO archived](https://github.com/minio/minio) -- Repository archived Feb 13, 2026
- [MinIO maintenance mode (InfoQ)](https://www.infoq.com/news/2025/12/minio-s3-api-alternatives/) -- Context on MinIO death
- [SeaweedFS vs alternatives](https://rilavek.com/resources/self-hosted-s3-compatible-object-storage-2026) -- Comparison
- [Caddy releases](https://github.com/caddyserver/caddy/releases) -- v2.11.1
- [K3s v1.35 release](https://docs.k3s.io/blog/2026/01/15/K3s-1.35-release) -- Latest K3s
- [telego](https://github.com/mymmrac/telego) -- Telegram bot library
- [MongoDB Go driver v2](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2) -- Latest driver
- [Vitest 4 announcement](https://voidzero.dev/posts/announcing-vitest-4) -- Browser mode stable
- [Playwright 1.58](https://playwright.dev/docs/release-notes) -- Chrome for Testing

---
*Stack research for: Team collaboration hub (task tracker + messenger + TMS + voice calls)*
*Researched: 2026-03-20*
