# Milestone 2: Technology Research

**Researched:** 2026-03-22
**Overall confidence:** HIGH (most technologies already in the stack or have official Go SDKs)

---

## 1. GitHub App vs GitHub OAuth App

**Recommendation: GitHub App** -- use a GitHub App as the primary integration method, with user-level OAuth as a supplement for personal actions.

### Why GitHub App (not OAuth App)

| Criterion | GitHub App | OAuth App |
|-----------|-----------|-----------|
| Token scope | Fine-grained per-repository permissions | Broad user-level scopes |
| Token lifetime | Short-lived (1h installation tokens) | Long-lived (until revoked) |
| Rate limits | Scales with repos + org users (5000+/hr) | Fixed 5000/hr per user |
| Tied to user? | No -- survives employee departure | Yes -- dies with user account |
| Repository access | Org owner selects repos during install | Whatever user can access |
| Webhook delivery | Built-in, per-installation | Manual webhook setup |
| Acts independently | Yes (bot operations without user) | No -- always on behalf of user |

**GitHub is actively deprecating OAuth Apps in favor of GitHub Apps.** All new integrations should default to GitHub Apps.

### Implementation Architecture

```
Hub Admin Settings -> "Install GitHub App" button
    -> Redirects to github.com/apps/hub-dev/installations/new
    -> User selects org + repos
    -> GitHub redirects back with installation_id
    -> Hub stores installation_id + generates installation access token

For user-level actions (commenting as the user, not as "hub-bot"):
    User Settings -> "Connect GitHub Account"
    -> OAuth flow (separate from GitHub App)
    -> Stores user's OAuth token for personal actions
```

### Required GitHub App Permissions

| Permission | Access | Used For |
|------------|--------|----------|
| Issues | Read & Write | Bidirectional issue sync |
| Pull Requests | Read | PR status, review status, diffs |
| Actions | Read | Workflow run status, artifacts |
| Contents | Read | File diffs, branch info |
| Checks | Read | Check run status |
| Metadata | Read | Repository info |

### Webhook Events to Subscribe

- `push` -- commit linking to tasks
- `pull_request` -- PR status, review status
- `pull_request_review` -- review approved/changes_requested
- `workflow_run` -- CI/CD pipeline status
- `issues` -- bidirectional sync
- `issue_comment` -- comment sync
- `check_run` -- status checks

### Go Implementation

The project already uses `go-github/v68`. GitHub App authentication requires:

```go
// Already in go.mod: github.com/google/go-github/v68
// Add: github.com/bradleyfalzon/ghinstallation/v2

// Installation-level client (for bot operations)
itr, _ := ghinstallation.New(http.DefaultTransport, appID, installationID, privateKey)
client := github.NewClient(&http.Client{Transport: itr})

// User-level client (for personal operations)
ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: userToken})
client := github.NewClient(oauth2.NewClient(ctx, ts))
```

**New dependency:** `github.com/bradleyfalzon/ghinstallation/v2` -- widely used, MIT license, 1.5k stars.

**Confidence:** HIGH -- based on official GitHub documentation and go-github examples.

**Sources:**
- [GitHub Docs: Differences between GitHub Apps and OAuth Apps](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/differences-between-github-apps-and-oauth-apps)
- [GitHub Docs: Deciding when to build a GitHub App](https://docs.github.com/en/apps/creating-github-apps/about-creating-github-apps/deciding-when-to-build-a-github-app)
- [Nango: GitHub App vs GitHub OAuth](https://nango.dev/blog/github-app-vs-github-oauth)

---

## 2. AI Integration: Claude API via Anthropic Go SDK

**Recommendation: Official `anthropic-sdk-go`** -- Anthropic's own Go SDK, actively maintained, type-safe.

### SDK Selection

| Option | Recommendation | Notes |
|--------|---------------|-------|
| `anthropics/anthropic-sdk-go` (official) | **USE THIS** | Official, maintained by Anthropic, full API coverage |
| `liushuangls/go-anthropic` | No | Community wrapper, will lag behind official |
| Direct HTTP calls | No | No type safety, manual retry/streaming handling |
| Local models (Ollama) | No for now | Latency is fine for Hub's use cases, Claude quality >> local |

### Model Selection Strategy

| Use Case | Model | Why | Est. Cost |
|----------|-------|-----|-----------|
| Test case generation | Claude Haiku 4.5 | Fast (< 2s), cheap, good enough for structured generation | ~$0.001/request |
| Code review suggestions | Claude Sonnet 4.5 | Needs deeper code understanding | ~$0.01/request |
| Natural language search | Claude Haiku 4.5 | JQL translation is simple pattern matching | ~$0.0005/request |
| Sprint retrospective | Claude Sonnet 4.5 | Needs synthesis of complex data | ~$0.02/request |
| Bug enrichment | Claude Haiku 4.5 | Structured extraction from text | ~$0.001/request |
| Semantic embeddings | Claude Haiku 4.5 | Embedding-like similarity via prompt | ~$0.0005/request |

**Monthly cost estimate (10-person team, moderate usage):** ~$15-30/month. Well within budget constraints.

### Architecture Pattern

```
User Action (e.g., "Generate Tests")
    -> Backend Handler validates + authorizes
    -> Publishes to NATS JetStream "AI.requests" stream
    -> AI Worker consumes from stream
        -> Check Redis cache (prompt hash -> response)
        -> If miss: call Anthropic API via SDK
        -> Store response in Redis (1h TTL)
        -> Publish result to NATS "AI.responses.{requestID}"
    -> WebSocket delivers result to frontend
```

**Why NATS queue, not direct API call:**
- Rate limiting (Anthropic has per-minute limits)
- Retry with exponential backoff on 429/500
- Cost tracking (log every request with tokens used)
- Decouple request from response (AI calls can take 5-30s)

### Semantic Search: pgvector

For semantic search (M2-103), use PostgreSQL `pgvector` extension rather than a separate vector database:

```sql
-- Already using PostgreSQL, just add extension
CREATE EXTENSION vector;

-- Add embedding column to relevant tables
ALTER TABLE issues ADD COLUMN embedding vector(1024);
ALTER TABLE test_cases ADD COLUMN embedding vector(1024);

-- Create index for fast similarity search
CREATE INDEX ON issues USING ivfflat (embedding vector_cosine_ops);
```

Generate embeddings via Claude API (or switch to a dedicated embedding model if cost becomes an issue). Store in pgvector. Query with cosine similarity.

**Confidence:** HIGH for SDK choice (official Anthropic SDK), MEDIUM for pgvector (need to verify CloudForBox PostgreSQL supports extensions).

**Sources:**
- [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)
- [Claude API Docs: Client SDKs](https://platform.claude.com/docs/en/api/client-sdks)
- [Claude API Docs: Go SDK](https://platform.claude.com/docs/en/api/sdks/go)

---

## 3. K3s Deployment Patterns for Go + React

**Recommendation: Kustomize bases + overlays** -- no Helm charts needed for a single-tenant app. Kustomize is built into kubectl.

### Deployment Architecture

```
K3s Cluster (CloudForBox)
├── Namespace: hub
│   ├── Deployment: hub-backend (2 replicas, Go binary)
│   ├── Deployment: hub-frontend (1 replica, Caddy serving static)
│   ├── StatefulSet: postgresql (1 replica, PVC)
│   ├── StatefulSet: mongodb (1 replica, PVC)
│   ├── Deployment: redis (1 replica)
│   ├── Deployment: nats (1 replica, JetStream enabled)
│   ├── Deployment: seaweedfs (1 replica, PVC)
│   ├── Deployment: livekit (1 replica)
│   ├── Deployment: loki (1 replica, PVC)
│   ├── Deployment: prometheus (1 replica, PVC)
│   ├── Deployment: grafana (1 replica)
│   ├── ConfigMap: hub-config
│   ├── Secret: hub-secrets
│   ├── Ingress: hub-ingress (Traefik)
│   └── Job: hub-migrate (runs before deployment)
```

### Kustomize Structure

```
deployments/k3s/
├── base/
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── backend/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── hpa.yaml
│   ├── frontend/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── postgresql/
│   │   ├── statefulset.yaml
│   │   ├── service.yaml
│   │   └── pvc.yaml
│   ├── mongodb/
│   ├── redis/
│   ├── nats/
│   ├── seaweedfs/
│   ├── livekit/
│   ├── monitoring/
│   │   ├── prometheus.yaml
│   │   ├── loki.yaml
│   │   └── grafana.yaml
│   ├── ingress.yaml
│   └── migrate-job.yaml
└── overlays/
    ├── dev/
    │   └── kustomization.yaml  (1 replica, debug logging)
    └── prod/
        └── kustomization.yaml  (2 replicas, resource limits)
```

### Resource Budget (CloudForBox)

| Service | CPU Request | CPU Limit | RAM Request | RAM Limit |
|---------|-------------|-----------|-------------|-----------|
| backend (x2) | 250m | 500m | 256Mi | 512Mi |
| frontend | 50m | 100m | 64Mi | 128Mi |
| postgresql | 250m | 1000m | 512Mi | 1Gi |
| mongodb | 250m | 500m | 256Mi | 512Mi |
| redis | 100m | 200m | 128Mi | 256Mi |
| nats | 100m | 200m | 64Mi | 128Mi |
| seaweedfs | 100m | 500m | 256Mi | 512Mi |
| livekit | 250m | 500m | 256Mi | 512Mi |
| prometheus | 100m | 200m | 256Mi | 512Mi |
| loki | 100m | 200m | 128Mi | 256Mi |
| grafana | 50m | 100m | 64Mi | 128Mi |
| **Total** | **~1.85 CPU** | **~4 CPU** | **~2.4Gi** | **~4.9Gi** |

K3s itself needs ~500MB RAM. **Minimum server: 4 CPU, 8GB RAM.** Recommended: 6 CPU, 12GB RAM.

### Zero-Downtime Deploy Strategy

```yaml
# Rolling update
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # create 1 new pod
      maxUnavailable: 0   # never kill old pod before new is ready
  template:
    spec:
      terminationGracePeriodSeconds: 30
      containers:
      - name: hub-backend
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health/live
            port: 3000
          initialDelaySeconds: 10
          periodSeconds: 10
```

**Confidence:** HIGH -- K3s + Kustomize is a well-documented pattern for single-tenant apps.

**Sources:**
- [K3s Official Docs](https://docs.k3s.io/)
- [Kustomize Official](https://kustomize.io/)

---

## 4. CI/CD for Monorepo (Go + React + CLI)

**Recommendation: GitHub Actions with path-filtered jobs** -- no Turborepo or Nx needed (those are JS-only). Path filters in GitHub Actions handle polyglot monorepos well.

### Workflow Structure

```yaml
# .github/workflows/ci.yml
name: CI
on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  changes:
    runs-on: ubuntu-latest
    outputs:
      backend: ${{ steps.filter.outputs.backend }}
      frontend: ${{ steps.filter.outputs.frontend }}
      cli: ${{ steps.filter.outputs.cli }}
    steps:
      - uses: dorny/paths-filter@v3
        id: filter
        with:
          filters: |
            backend:
              - 'backend/**'
            frontend:
              - 'frontend/**'
            cli:
              - 'cli/**'

  backend:
    needs: changes
    if: needs.changes.outputs.backend == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
      - run: go vet ./...
      - run: go test ./... -race -coverprofile=coverage.out
      - run: golangci-lint run

  frontend:
    needs: changes
    if: needs.changes.outputs.frontend == 'true'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/setup-node@v4
        with:
          node-version: '22'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
      - run: npx tsc --noEmit
      - run: npx vitest run
      - run: npx vite build

  docker:
    needs: [backend, frontend]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v6
        with:
          context: ./backend
          push: true
          tags: ghcr.io/${{ github.repository }}/backend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### Key Practices

1. **Path filtering with `dorny/paths-filter`** -- only run backend tests when backend changes, saves ~3000 CI minutes/month
2. **Go module caching** -- `actions/setup-go` with `cache: true` caches `~/go/pkg/mod`
3. **npm caching** -- `actions/setup-node` with `cache: 'npm'` caches `~/.npm`
4. **Docker layer caching** -- `cache-from: type=gha` uses GitHub Actions cache for Docker layers
5. **Required status checks** -- all three jobs (backend, frontend, cli) must pass before merge
6. **Coverage reporting** -- upload to Codecov for PR coverage comments

### Budget Consideration

With ~3000 free CI minutes (GitHub student pack), path filtering is critical. Without it, every PR runs all three jobs (~6 min total). With filtering, most PRs run 1 job (~2 min). That is 3x more PRs per minute budget.

**Confidence:** HIGH -- standard GitHub Actions patterns, well-documented.

**Sources:**
- [GitHub Actions Monorepo Guide 2026](https://dev.to/pockit_tools/github-actions-in-2026-the-complete-guide-to-monorepo-cicd-and-self-hosted-runners-1jop)
- [dorny/paths-filter](https://github.com/dorny/paths-filter)

---

## 5. Code Splitting for Large React App with Vite

**Recommendation: TanStack Router automatic code splitting + manual vendor chunks** -- the project already uses TanStack Router, which has built-in code splitting support.

### Strategy 1: TanStack Router Lazy Routes (Primary)

TanStack Router supports `.lazy.tsx` files that split route components from route configuration. This is the highest-impact optimization.

**Before (current -- everything in one chunk):**
```
src/routes/_authenticated/projects/$projectId/board/index.tsx  (340 lines)
```

**After (code split):**
```
src/routes/_authenticated/projects/$projectId/board/index.tsx      (route config, loader)
src/routes/_authenticated/projects/$projectId/board/index.lazy.tsx  (component code)
```

The `.lazy.tsx` file contains the heavy component code. TanStack Router automatically loads it only when the route is navigated to. The route file keeps only `createFileRoute` with `beforeLoad`/`loader`.

**Impact estimate:** With 15+ route files averaging 300-500 lines each, this alone could split ~60% of the 2.6MB bundle into on-demand chunks.

### Strategy 2: Vite manualChunks (Vendor Splitting)

```typescript
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react/jsx-runtime'],
          'vendor-tanstack': ['@tanstack/react-router', '@tanstack/react-query'],
          'vendor-plate': ['@udecode/plate', '@udecode/plate/react'],
          'vendor-livekit': ['@livekit/components-react', 'livekit-client'],
          'vendor-charts': ['recharts'],
          'vendor-ui': ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu',
                         '@radix-ui/react-popover', '@radix-ui/react-select'],
          'vendor-dnd': ['@atlaskit/pragmatic-drag-and-drop'],
          'vendor-i18n': ['i18next', 'react-i18next'],
        },
      },
    },
  },
});
```

**Why manual chunks matter:** Without this, Vite/Rollup puts all vendor code into a single huge chunk. Splitting vendors means:
- `vendor-react` (~150KB) is cached forever (changes rarely)
- `vendor-plate` (~200KB) only loads on pages with editors
- `vendor-livekit` (~300KB) only loads when joining calls
- `vendor-charts` (~150KB) only loads on dashboard

### Strategy 3: Dynamic Imports for Heavy Components

```typescript
// Lazy-load Plate editor
const PlateEditor = React.lazy(() => import('@/components/editor/plate-editor'));

// Lazy-load LiveKit
const CallView = React.lazy(() => import('@/modules/calls/call-view'));

// Lazy-load charts
const BurndownChart = React.lazy(() => import('@/components/charts/burndown'));
```

### Strategy 4: Preloading for UX

```typescript
// Preload board chunk when hovering project in sidebar
<Link
  to="/projects/$projectId/board"
  onMouseEnter={() => {
    import('@/routes/_authenticated/projects/$projectId/board/index.lazy');
  }}
>
```

### Expected Bundle Impact

| Chunk | Current | After Split | Load Timing |
|-------|---------|-------------|-------------|
| Main entry | ~2.6MB | ~180KB | Initial |
| vendor-react | (in main) | ~150KB | Initial |
| vendor-tanstack | (in main) | ~80KB | Initial |
| vendor-ui | (in main) | ~100KB | Initial |
| vendor-plate | (in main) | ~200KB | On editor focus |
| vendor-livekit | (in main) | ~300KB | On call join |
| vendor-charts | (in main) | ~150KB | On dashboard |
| Route chunks (avg) | (in main) | ~20-50KB each | On navigation |

**Target: main chunk under 200KB gzipped** (currently ~650KB gzipped from 2.6MB uncompressed).

### Analysis Tool

Add `rollup-plugin-visualizer` as dev dependency to generate treemap of bundle contents:

```bash
npm install -D rollup-plugin-visualizer
```

```typescript
// vite.config.ts
import { visualizer } from 'rollup-plugin-visualizer';

plugins: [
  visualizer({ open: true, gzipSize: true }),
]
```

**Confidence:** HIGH -- TanStack Router docs explicitly support this, Vite's Rollup integration is mature.

**Sources:**
- [TanStack Router: Code Splitting](https://tanstack.com/router/latest/docs/guide/code-splitting)
- [TanStack Router: Automatic Code Splitting](https://tanstack.com/router/v1/docs/framework/react/guide/automatic-code-splitting)
- [Vite + React bundle optimization](https://www.mykolaaleksandrov.dev/posts/2025/11/taming-large-chunks-vite-react/)
- [Optimizing Bundle Splitting: 82% reduction](https://briandouglas.me/posts/2025/08/23/optimizing-bundle-splitting/)

---

## 6. Atlas Migration Strategy for Production

**Recommendation: Atlas versioned migrations** -- the project already uses Ent + Atlas. Move from dev auto-migrate to versioned migration files for production safety.

### Current State (v1)

In development, Ent's `client.Schema.Create()` auto-migrates on startup. This is unsafe for production because:
- No migration history
- Cannot roll back
- Destructive changes applied without review

### Production Strategy

```bash
# Generate migration from Ent schema diff
atlas migrate diff add_ai_embeddings \
  --dir "file://backend/migrations" \
  --to "ent://backend/internal/ent/schema" \
  --dev-url "docker://postgres/15/dev"

# Lint migration for safety
atlas migrate lint \
  --dir "file://backend/migrations" \
  --dev-url "docker://postgres/15/dev"

# Apply in production (as K8s Job)
atlas migrate apply \
  --dir "file://backend/migrations" \
  --url "postgres://user:pass@postgres:5432/hub?sslmode=disable"
```

### CI Integration

```yaml
# In CI pipeline
- name: Lint migrations
  run: |
    atlas migrate lint \
      --dir "file://backend/migrations" \
      --dev-url "docker://postgres/15/dev" \
      --git-dir "." \
      --git-base "origin/main"
```

Atlas lint catches: destructive column drops, table renames without data migration, index creation on large tables without CONCURRENTLY, constraint additions that may fail on existing data.

**Confidence:** HIGH -- Atlas is the official migration tool for Ent, well-documented integration path.

**Sources:**
- [Atlas: Versioned Migrations](https://atlasgo.io/versioned/diff)
- [Ent: Versioned Migrations](https://entgo.io/docs/versioned-migrations/)
- [Atlas: Handling Migration Errors](https://atlasgo.io/blog/2025/04/06/atlas-and-golang-migrate)

---

## 7. New Dependencies Summary

### Backend (Go)

| Package | Version | Purpose | Phase |
|---------|---------|---------|-------|
| `github.com/bradleyfalzon/ghinstallation/v2` | latest | GitHub App JWT authentication | 10 |
| `github.com/anthropics/anthropic-sdk-go` | latest | Claude API client | 11 |
| `github.com/ansrivas/fiberprometheus/v3` | latest | Prometheus metrics middleware for Fiber v3 | 9 |
| `github.com/pgvector/pgvector-go` | latest | pgvector support for semantic search | 11 |
| `ariga.io/atlas-go-sdk` | latest | Atlas migration SDK (for programmatic apply) | 9 |

### Frontend (npm)

| Package | Version | Purpose | Phase |
|---------|---------|---------|-------|
| `rollup-plugin-visualizer` | latest (dev) | Bundle analysis | 8 |
| `@tanstack/react-virtual` | latest | Virtual scrolling (may already be present) | 8 |
| `react-diff-viewer-continued` | latest | Syntax-highlighted diff viewer for PRs | 10 |

### Infrastructure

| Tool | Purpose | Phase |
|------|---------|-------|
| Atlas CLI | Migration management | 9 |
| `dorny/paths-filter` GitHub Action | Monorepo path detection | 9 |
| `docker/build-push-action` GitHub Action | Docker image builds | 9 |

---

*Researched: 2026-03-22*
*Sources verified against official documentation where available*
