# Narrow Platform API QA Contour Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Execute `TC-PLATFORM-0001` and `TC-PLATFORM-0002` against a real,
QA-tagged Fiber/Auth/Ent/Redis SUT while preserving the exact eight public
root-Compose QA commands.

**Architecture:** Four private QA binaries provide secret initialization,
schema creation, the narrow public SUT, and a nonce-ledger control plane. A
fixed Python provider composes the existing opaque Platform harness and talks
only to those internal services; PostgreSQL and Redis are the only product
infrastructure dependencies.

**Tech Stack:** Go 1.25, Fiber v3, Ent, go-redis, PostgreSQL 17, Redis 7.4,
Python 3.12, pytest 9, httpx, Docker Compose, multi-stage Docker builds.

**Spec:**
`docs/superpowers/specs/2026-08-20-narrow-platform-api-qa-design.md`

## Global Constraints

- Read the spec before starting every task; the spec wins over shortcuts in
  this plan.
- Only the root orchestrator may coordinate. Atomic code implementation uses
  `gpt-5.3-codex` with `xhigh`; if that model is unavailable, return `BLOCKED`
  rather than falling back. Independent review uses `gpt-5.6-terra` with
  `high` effort.
- The numbered tasks are root-owned integration envelopes, not permission to
  hand a whole multi-module slice to one worker. Within each task, dispatch one
  Spark worker per production module plus its focused test file, sequentially
  where a shared contract file is involved. The root performs the envelope
  integration check and exact-path commit only after all module reviews pass.
- Preserve the dirty worktree. Own only files listed in the current task,
  never revert another worker, never use broad `git add`, and commit with
  an explicit exact-path list, for example
  `git commit --only -m "message" -- path/one path/two`.
- Never read or modify `pro-qa-ai/**`.
- Do not change `backend/cmd/server`, product routes, auth semantics, limiter
  values, magic-link behavior, or session behavior.
- Every new Go implementation file under `backend/cmd/qa-*` and
  `backend/internal/qa/**` uses `//go:build qa`. Go tests under
  `backend/tests/qa/**` use `//go:build hubtest` and are made executable by the
  runner's combined `hubtest qa` tags.
- Compile QA binaries with exact tags `release qa`. Pass a nil
  `*natsclient.Client`; do not create a NATS fake or dependency.
- Public Compose services remain exactly eight. The only new service names are
  the private support services `qa-secret-init`, `qa-schema`, `qa-backend`,
  and `qa-control`.
- The narrow contour depends on `qa-postgres` and `qa-redis` only. No MongoDB,
  NATS, SeaweedFS, Resend, Docker socket, host port, host test process, or
  arbitrary command is allowed.
- Secrets are random, file-backed, absent from environment values and the
  repository, and mounted according to the spec's closed matrix.
- All test/build/vet/lint/validator execution is Docker-only through root
  `docker-compose.local.yml`. Host work is limited to reading, editing, Git,
  and static text inspection.
- A `python-contract` command may remain non-zero because of the already-known
  product contract failures. The framework-selftest phase must be completely
  green before its second SUT phase starts.
- The aggregate `python-api` command remains `PARTIAL_FAIL` until other domain
  providers exist. Acceptance for this slice is two exact Platform nodes PASS
  twice sequentially, not a green aggregate exit.
- No runtime PASS, cleanup guarantee, or Docker compatibility claim may be
  inferred from source review.

## File Structure

### New backend files

- `backend/cmd/qa-secret-init/main_qa.go` — one-shot secret initializer CLI.
- `backend/cmd/qa-schema/main_qa.go` — one-shot Ent schema CLI.
- `backend/cmd/qa-backend/main_qa.go` — narrow SUT CLI and health subcommand.
- `backend/cmd/qa-control/main_qa.go` — control-plane CLI and health subcommand.
- `backend/internal/qa/config_qa.go` — closed runtime gates and topology.
- `backend/internal/qa/secret_files_qa.go` — held/no-follow secret read and
  atomic create/reuse.
- `backend/internal/qa/database_qa.go` — credential-file PostgreSQL/Redis open.
- `backend/internal/qa/schema_qa.go` — explicit Ent schema application.
- `backend/internal/qa/backend_app_qa.go` — Fiber app and auth-specific wiring.
- `backend/internal/qa/control_models_qa.go` — closed flows, scenario, ledger
  models, and canonical keys.
- `backend/internal/qa/control_ent_qa.go` — exact Ent user/session operations.
- `backend/internal/qa/control_ledger_qa.go` — Redis ledger transitions.
- `backend/internal/qa/control_service_qa.go` — provision/observe/cleanup.
- `backend/internal/qa/control_http_qa.go` — closed authenticated control API.

### New backend tests

- `backend/tests/qa/config_test.go`
- `backend/tests/qa/secret_files_test.go`
- `backend/tests/qa/schema_test.go`
- `backend/tests/qa/backend_app_test.go`
- `backend/tests/qa/control_models_test.go`
- `backend/tests/qa/control_ledger_test.go`
- `backend/tests/qa/control_ent_test.go`
- `backend/tests/qa/control_service_test.go`
- `backend/tests/qa/control_http_test.go`
- `backend/tests/qa/build_tags_test.go`

### New Python files

- `testframework/src/hub_testkit/runtime/secret_file.py` — held, bounded ASCII
  secret reader.
- `testframework/src/hub_testkit/harness/platform_abuse_provider.py` — fixed
  operational provider and delegate.
- `testframework/src/hub_testkit/pytest_plugin/platform_abuse_provider.py` —
  composition-only pytest hook.
- `testframework/selftests/runtime/test_secret_file.py`
- `testframework/selftests/collection/pytest_plugin/test_plugin_state_access.py`
- `testframework/selftests/harness/test_platform_abuse_provider.py`
- `testframework/selftests/collection/pytest_plugin/test_platform_abuse_provider.py`
- `testframework/selftests/architecture/imports/test_platform_abuse_provider_boundary.py`

### Modified execution/contract files

- `testframework/docker/entrypoints/backend-unit.sh`
- `testframework/docker/entrypoints/python-qa.sh`
- `testframework/docker/Dockerfile.qa`
- `docker-compose.local.yml`
- `testframework/selftests/architecture/docker/test_root_compose_contract.py`
- replace `testframework/policies/root-compose-testframework.v1.json` with
  `testframework/policies/root-compose-testframework.v2.json`

### Modified package/docs files

- `testframework/src/hub_testkit/runtime/__init__.py`
- `testframework/src/hub_testkit/harness/platform_abuse.py`
- `testframework/src/hub_testkit/pytest_plugin/plugin.py`
- `testframework/README.md`
- `testframework/PACKAGE-CONTRACT.md`
- `testframework/PIPELINE.md`
- `testframework/TESTING_GUIDE.md`
- `testframework/docker/README.md`
- `docs/TESTING.md`
- `docs/test_plan.md`
- `docs/stack_and_tools.md`
- `docs/onboarding/qa-automation.md`
- `testframework/cases/platform/TC-PLATFORM-0001.md`
- `testframework/cases/platform/TC-PLATFORM-0002.md`
- `.planning/STATE.md` only after Docker evidence exists.

---

### Task 1: Admit QA-tagged Go units and implement runtime/secret primitives

**Files:**

- Create: `backend/internal/qa/config_qa.go`
- Create: `backend/internal/qa/secret_files_qa.go`
- Create: `backend/tests/qa/config_test.go`
- Create: `backend/tests/qa/secret_files_test.go`
- Modify: `backend/go.mod`
- Modify: `testframework/docker/entrypoints/backend-unit.sh`
- Modify: `testframework/selftests/architecture/docker/test_root_compose_contract.py`

**Atomic dispatches:** config+config tests; secret reader/writer+secret tests and
the pinned direct dependency; exact runner-tag contract. The contract-file
dispatch runs last to avoid overlapping edits.

**Interfaces:**

- Produces:

```go
type Service string

const (
    ServiceSecretInit Service = "qa-secret-init"
    ServiceSchema     Service = "qa-schema"
    ServiceBackend    Service = "qa-backend"
    ServiceControl    Service = "qa-control"
)

type Config struct {
    Service              Service
    PostgresHost         string
    PostgresPort         string
    PostgresUser         string
    PostgresDatabase     string
    PostgresPasswordFile string
    RedisHost            string
    RedisPort            string
    JWTSecretFile        string
    ControlTokenFile     string
    SchemaApply          bool
}

func LoadConfig(service Service, lookup func(string) (string, bool)) (Config, error)
func ReadHexSecret(path string) (string, error)
func EnsureHexSecret(path string, random io.Reader) error
```

The only admitted secret paths are
`/run/hub-qa/postgres/secret`, `/run/hub-qa/jwt/secret`, and
`/run/hub-qa/control/secret` for the services named by the spec's mount matrix.

- `LoadConfig` requires `HUB_QA_MODE=1`, `ENVIRONMENT=qa`, and
  `DEV_MODE=false`; it admits only `qa-postgres` and `qa-redis`, exact file
  paths, and no secret/credential-bearing DSN environment values.
- `ReadHexSecret` accepts exactly 64 lower-case hexadecimal characters plus
  one LF from a regular `0444` file opened with no-follow semantics. It reads
  at most 66 bytes and verifies stable file identity before/after.
- `EnsureHexSecret` creates atomically with `O_EXCL`, validates existing files,
  and returns only constant generic errors.

- [ ] **Step 1: Write the failing architecture and Go tests**

Add source-contract expectations that both unit and race commands use exact
`-tags "hubtest qa"`. Add table tests covering every service gate, external
hosts, `DEV_MODE=true`, missing schema apply, environment passwords/DSNs,
symlink/directory/FIFO/oversize/uppercase/CR/invalid-hex secrets, hostile
canaries, create, reuse, and unexpected pre-existing content.

The core test shapes are:

```go
func TestLoadConfigRejectsDevelopmentMode(t *testing.T) {
    env := validBackendEnv()
    env["DEV_MODE"] = "true"
    _, err := qa.LoadConfig(qa.ServiceBackend, lookup(env))
    require.EqualError(t, err, "qa runtime configuration is invalid")
}

func TestReadHexSecretRejectsSymlinkWithoutCanaryLeak(t *testing.T) {
    path, canary := hostileSecretFixture(t)
    _, err := qa.ReadHexSecret(path)
    require.EqualError(t, err, "qa secret is invalid")
    require.NotContains(t, fmt.Sprintf("%+v", err), canary)
}
```

- [ ] **Step 2: Run the Docker RED check**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: unit/race fail because `backend/tests/qa/**` imports a `qa`-tagged
package while the current runner supplies only `hubtest`.

Run:

```text
docker compose -f docker-compose.local.yml build python-contract
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected: framework phase fails the exact backend-entrypoint source contract
before product contracts execute.

- [ ] **Step 3: Implement the minimal runtime and secret code**

Use `golang.org/x/sys/unix` for `O_NOFOLLOW`/`Fstat`; move the existing pinned
`golang.org/x/sys v0.42.0` line from the indirect block to the direct require
block in `backend/go.mod`. Do not change its version or run a host resolver.

Change the existing runner commands exactly:

```sh
go vet -tags "hubtest qa" ./...
go test -tags "hubtest qa" -json -count=1 -timeout=150s ./tests/...
go test -tags "hubtest qa" -json -race -count=1 -timeout=150s ./tests/...
```

- [ ] **Step 4: Run the Docker GREEN checks**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit python-contract
docker compose -f docker-compose.local.yml run --rm backend-unit
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected:

- `backend-unit`: build, vet, unit, and race all PASS;
- `python-contract`: every framework selftest passes, then the known product
  contract phase may retain its independent failures.

- [ ] **Step 5: Independent review and commit**

Request Terra/high review for secret races, symlink handling, generic errors,
env/topology closure, and exact runner collection. Fix all HIGH/MEDIUM findings.

Commit only owned files:

```text
git commit --only -m "feat(qa): add closed runtime and secret primitives" -- backend/go.mod backend/internal/qa/config_qa.go backend/internal/qa/secret_files_qa.go backend/tests/qa/config_test.go backend/tests/qa/secret_files_test.go testframework/docker/entrypoints/backend-unit.sh testframework/selftests/architecture/docker/test_root_compose_contract.py
```

### Task 2: Add secret initialization and explicit schema binaries

**Files:**

- Create: `backend/internal/qa/database_qa.go`
- Create: `backend/internal/qa/schema_qa.go`
- Create: `backend/cmd/qa-secret-init/main_qa.go`
- Create: `backend/cmd/qa-schema/main_qa.go`
- Create: `backend/tests/qa/schema_test.go`

**Atomic dispatches:** database/schema helpers+schema tests; secret-init
command; schema command. The commands consume only the reviewed helper API.

**Interfaces:**

```go
type Database struct {
    Ent *ent.Client
    SQL *sql.DB
}

func OpenDatabase(ctx context.Context, cfg Config) (*Database, error)
func OpenRedis(ctx context.Context, cfg Config) (*redis.Client, error)
func EnsureRuntimeSecrets(cfg Config, random io.Reader) error
func ApplySchema(ctx context.Context, client *ent.Client) error
```

`OpenDatabase` reads the password from the approved file, constructs the DSN
inside the process, and never includes it in an error. `ApplySchema` calls
`client.Schema.Create(ctx)` explicitly while `DEV_MODE=false`.

- [ ] **Step 1: Write failing one-shot behavior tests**

Cover exact three-file initialization order, create/reuse, partial-set failure,
`HUB_QA_SCHEMA_APPLY=1`, no implicit dev-mode migration, context cancellation,
and constant errors. Use injected functions for orchestration tests; do not
fake product auth behavior.

```go
func TestApplySchemaRequiresExplicitGate(t *testing.T) {
    env := validSchemaEnv()
    delete(env, "HUB_QA_SCHEMA_APPLY")
    _, err := qa.LoadConfig(qa.ServiceSchema, lookup(env))
    require.EqualError(t, err, "qa runtime configuration is invalid")
}
```

- [ ] **Step 2: Run Docker RED**

Run `docker compose -f docker-compose.local.yml build backend-unit`, then
`docker compose -f docker-compose.local.yml run --rm backend-unit`. Expected:
the new QA unit package fails before implementation.

- [ ] **Step 3: Implement internal orchestration and thin commands**

Both command files remain composition-only. `qa-secret-init` calls
`EnsureRuntimeSecrets` and exits. `qa-schema` opens PostgreSQL, defers both Ent
and SQL closes, calls `ApplySchema`, and exits. Both use signal-aware contexts
and print only a fixed failure line.

- [ ] **Step 4: Run Docker GREEN**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: all four backend steps currently present in the entrypoint's result
model remain green; no runtime service is started yet.

- [ ] **Step 5: Review and commit**

Request Terra/high review for DSN redaction, explicit schema gating, close
order, rootless consumer assumptions, and command thinness. Fix H/M findings.

```text
git commit --only -m "feat(qa): add secret and schema one-shots" -- backend/internal/qa/database_qa.go backend/internal/qa/schema_qa.go backend/cmd/qa-secret-init/main_qa.go backend/cmd/qa-schema/main_qa.go backend/tests/qa/schema_test.go
```

### Task 3: Build the narrow production-auth Fiber SUT

**Files:**

- Create: `backend/internal/qa/backend_app_qa.go`
- Create: `backend/cmd/qa-backend/main_qa.go`
- Create: `backend/tests/qa/backend_app_test.go`

**Atomic dispatches:** Fiber app+app tests; thin backend command. The command is
written only after the app contract is accepted.

**Interfaces:**

```go
type Readiness func(context.Context) error

type BackendDeps struct {
    Auth      *auth.AuthService
    Redis     *redis.Client
    RootCtx   context.Context
    Readiness Readiness
}

func NewBackendApp(deps BackendDeps) (*fiber.App, error)
```

The command constructs exactly:

```go
repo := auth.NewEntRepository(database.Ent)
magic := auth.NewMagicLinkService(redisClient, nil, "")
service := auth.NewAuthService(
    repo,
    redisClient,
    nil,
    jwtSecret,
    15*time.Minute,
    30*24*time.Hour,
    magic,
    database.SQL,
)
handler := auth.NewHandler(service, false)
```

`NewBackendApp` mounts verify-email `10/10m`, then auth-IP `3/10m`, then
`handler.VerifyMagicLink`; it never calls `RegisterRoutes`.

- [ ] **Step 1: Write the failing app tests**

Use Fiber `app.Test`, miniredis, and a test-local `auth.Repository` mock. Assert:

- only `/livez`, `/readyz`, and `/api/auth/magic-link/verify` exist;
- send-code, refresh, sessions, admin, dev-login, and arbitrary paths are 404;
- trusted XFF produces the expected `rl:auth:ip:<sha256[:16]>` key;
- three invalid requests are 401 and the fourth is 429 with `Retry-After: 600`;
- verify-email limiter executes before auth-IP;
- `DEV_MODE=false` and nil delivery never return `code`;
- readiness hides dependency errors.

```go
func TestBackendAppUsesProductAuthLimiterOrder(t *testing.T) {
    app, redisClient := newBackendUnitApp(t)
    statuses := sendFourInvalidVerifyRequests(t, app, "198.18.0.10")
    require.Equal(t, []int{401, 401, 401, 429}, statuses)
    require.Equal(t, int64(4), zcard(t, redisClient,
        "rl:auth-verify-email:email:rate@example.test"))
}
```

- [ ] **Step 2: Run Docker RED**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: new backend-app tests fail because `NewBackendApp` is absent.

- [ ] **Step 3: Implement the app and command**

Use the production Fiber error handler, struct validator, RequestID middleware,
and exact trusted-proxy CIDRs from `cmd/server/wiring_fiber.go`. The command
supports only no-argument serve mode and literal `healthcheck`; healthcheck
performs one bounded localhost `/readyz` request.

- [ ] **Step 4: Run Docker GREEN**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: build/vet/unit/race PASS, including non-vacuous 401/429 and XFF
assertions.

- [ ] **Step 5: Review and commit**

Request Terra/high review for route closure, middleware ordering, trusted proxy
behavior, NATS absence, and real auth semantics. Fix H/M findings.

```text
git commit --only -m "feat(qa): add narrow platform auth SUT" -- backend/internal/qa/backend_app_qa.go backend/cmd/qa-backend/main_qa.go backend/tests/qa/backend_app_test.go
```

### Task 4: Implement nonce-ledger provisioning, observation, and cleanup

**Files:**

- Create: `backend/internal/qa/control_models_qa.go`
- Create: `backend/internal/qa/control_ent_qa.go`
- Create: `backend/internal/qa/control_ledger_qa.go`
- Create: `backend/internal/qa/control_service_qa.go`
- Create: `backend/internal/qa/control_http_qa.go`
- Create: `backend/cmd/qa-control/main_qa.go`
- Create: `backend/tests/qa/control_models_test.go`
- Create: `backend/tests/qa/control_ledger_test.go`
- Create: `backend/tests/qa/control_ent_test.go`
- Create: `backend/tests/qa/control_service_test.go`
- Create: `backend/tests/qa/control_http_test.go`

**Atomic dispatches:** models+model tests; Redis ledger+ledger tests; Ent
adapter+compile/source tests; service+service tests; HTTP app+HTTP tests; thin
command. These are sequential because they share the frozen types below.

**Interfaces:**

```go
type ScenarioFlow string

const (
    UnknownCodeRateLimit ScenarioFlow = "unknown-code-rate-limit"
    OneTimeCodeReplay    ScenarioFlow = "one-time-code-replay"
)

type Scenario struct {
    Flow                     ScenarioFlow `json:"flow"`
    Email                    string       `json:"email"`
    Code                     string       `json:"code"`
    Nonce                    string       `json:"nonce"`
    ServerObservedClient     string       `json:"server_observed_client"`
    ServerObservedClientHash string       `json:"server_observed_client_hash"`
    IdentityActive           bool         `json:"identity_active"`
    DeliveryAttempted        bool         `json:"delivery_attempted"`
    DevelopmentLoginUsed     bool         `json:"development_login_used"`
    DevMode                  bool         `json:"dev_mode"`
    RedisHealthy             bool         `json:"redis_healthy"`
    DatabaseHealthy          bool         `json:"database_healthy"`
    RedisExclusive           bool         `json:"redis_exclusive"`
    DatabaseExclusive        bool         `json:"database_exclusive"`
    OwnedStateEntries        [4]string    `json:"owned_state_entries"`
}

type ControlStore interface {
    CreateActiveUser(context.Context, uuid.UUID, string) error
    SessionRefreshHashes(context.Context, uuid.UUID) ([]string, error)
    DeleteUserSessions(context.Context, uuid.UUID) error
    DeleteUser(context.Context, uuid.UUID) error
    UserExists(context.Context, uuid.UUID) (bool, error)
}

type Control struct {
    store       ControlStore
    redis       *redis.Client
    random      io.Reader
    now         func() time.Time
    mu          sync.Mutex
    activeNonce string
}

func NewControl(store ControlStore, rdb *redis.Client, random io.Reader) (*Control, error)
func (c *Control) Provision(context.Context, string, ScenarioFlow) (Scenario, error)
func (c *Control) Observe(context.Context, string) error
func (c *Control) Cleanup(context.Context, string) error
func NewControlApp(control *Control, token []byte, readiness Readiness) (*fiber.App, error)
```

`NewControlApp` exposes only:

```text
GET    /__hub_qa/v1/livez
GET    /__hub_qa/v1/readyz
POST   /__hub_qa/v1/platform-abuse/scenarios
POST   /__hub_qa/v1/platform-abuse/scenarios/{nonce}/observe
DELETE /__hub_qa/v1/platform-abuse/scenarios/{nonce}
```

Provision accepts only the closed JSON object `{"label": <label>, "flow":
<fixed-flow>}` with a 4 KiB body limit. Observe and cleanup accept no body.
Only liveness/readiness are unauthenticated.

- [ ] **Step 1: Write failing model/ledger/service/HTTP tests**

Cover:

- 32-char nonce, unique lower-case email, six-digit code, RFC 2544 IP derived
  from the nonce, canonical SHA-256 prefix, and serial one-live-scenario
  admission for the fixed non-xdist profile;
- exact four keys and internal `hubqa:platform-abuse:<nonce>` ledger;
- `SETNX`, 30-minute active TTL, five-minute cleaned tombstone;
- unknown flow creates no user/OTP; replay creates one active member user and
  one five-minute magic-link value;
- observe checks the exact global IP key after the first SUT request;
- cleanup order refresh keys → sessions → user → four keys → absence check →
  tombstone;
- partial cleanup retains active ledger and is retryable;
- missing/expired ledger fails; cleaned nonce is idempotent;
- no `SCAN`, `KEYS`, `FLUSHDB`, truncate, prefix delete, or caller-provided IDs;
- exact closed JSON, duplicate/unknown fields rejected, body capped;
- missing/wrong bearer tokens return the same `404` and fixed JSON body;
  comparison is constant-time; the only admitted credential header is exact
  `Authorization: Bearer <file-token>`; all responses use
  `Cache-Control: no-store`;
- `var _ ControlStore = (*EntControlStore)(nil)` binds the real Ent adapter at
  compile time, while service tests use a closed in-memory fake.

```go
func TestCleanupDeletesOnlyLedgerDerivedResources(t *testing.T) {
    control, calls := provisionReplayControl(t)
    require.NoError(t, control.Cleanup(context.Background(), calls.Nonce))
    require.Equal(t, []string{
        "refresh:owned-hash",
        "sessions:owned-user",
        "user:owned-user",
        "four-sut-keys",
        "verify-absence",
        "cleaned-tombstone",
    }, calls.Order())
}
```

- [ ] **Step 2: Run Docker RED**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: new control tests fail because the service and app do not exist.

- [ ] **Step 3: Implement exact Ent/Redis behavior**

The Ent adapter creates an active member without an organization. Cleanup
queries sessions only through `session.HasUserWith(user.ID(ownedUserID))`,
deletes `refresh:` plus the stored hash, deletes those sessions, and finally
deletes the exact user. Never log the ledger or scenario.

`Provision` admits at most one active scenario per provider process. It derives
the RFC 2544 address deterministically from the cryptographic nonce and rejects
a second live provision attempt until cleanup finalizes the first ledger. This
keeps the finite documentation subnet collision-free without `SCAN`, a shared
counter, or a caller-controlled address.

The command opens the real Ent and Redis clients, reads only the control-token
file, constructs `Control`, and supports only serve and literal healthcheck.

- [ ] **Step 4: Run Docker GREEN**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit
docker compose -f docker-compose.local.yml run --rm backend-unit
```

Expected: every exact ownership, auth, cleanup, and negative test passes under
unit and race.

- [ ] **Step 5: Review and commit**

Request two reviews: Terra/high functional correctness and Terra/high security.
Fix every H/M finding before commit.

```text
git commit --only -m "feat(qa): add platform scenario control plane" -- backend/internal/qa/control_models_qa.go backend/internal/qa/control_ent_qa.go backend/internal/qa/control_ledger_qa.go backend/internal/qa/control_service_qa.go backend/internal/qa/control_http_qa.go backend/cmd/qa-control/main_qa.go backend/tests/qa/control_models_test.go backend/tests/qa/control_ledger_test.go backend/tests/qa/control_ent_test.go backend/tests/qa/control_service_test.go backend/tests/qa/control_http_test.go
```

### Task 5: Implement the fixed Python Platform provider

**Files:**

- Create: `testframework/src/hub_testkit/runtime/secret_file.py`
- Create: `testframework/src/hub_testkit/harness/platform_abuse_provider.py`
- Create: `testframework/src/hub_testkit/pytest_plugin/platform_abuse_provider.py`
- Create: `testframework/selftests/runtime/test_secret_file.py`
- Create: `testframework/selftests/collection/pytest_plugin/test_plugin_state_access.py`
- Create: `testframework/selftests/harness/test_platform_abuse_provider.py`
- Create: `testframework/selftests/collection/pytest_plugin/test_platform_abuse_provider.py`
- Create: `testframework/selftests/architecture/imports/test_platform_abuse_provider_boundary.py`
- Modify: `testframework/src/hub_testkit/runtime/__init__.py`
- Modify: `testframework/src/hub_testkit/harness/platform_abuse.py`
- Modify: `testframework/src/hub_testkit/pytest_plugin/plugin.py`

**Atomic dispatches:** secret reader+runtime export+secret tests; public
redactor-state accessor+state-access tests; operational provider+harness tests;
thin pytest composition plugin+plugin tests; import-boundary test. The root
integrates them in that order.

**Interfaces:**

- `read_regular_ascii_secret(path: Path, *, maximum_bytes: int) -> str`
  returns one validated secret without its final LF.
- `require_redactor(config: pytest.Config) -> Redactor` returns the already
  validated immutable redactor or raises the existing generic framework
  configuration error.
- `build_platform_abuse_harness_factory(redactor: Redactor) ->
  PlatformAbuseHarnessFactory` returns the fixed callable factory.
- `pytest_configure(config: pytest.Config) -> None` in the thin provider plugin
  runs with `@pytest.hookimpl(trylast=True)` and installs that factory once.

The operational module owns fixed constants:

```python
_CONTROL_BASE_URL = "http://qa-control:3001"
_SUT_BASE_URL = "http://qa-backend:3000"
_CONTROL_TOKEN_PATH = Path("/run/hub-qa/control/secret")
_PROFILE_POLICY_PATH = Path("/opt/hub/testframework/policies/profiles.toml")
_PLATFORM_DEADLINE_SECONDS = 900
_CLEANUP_ALLOWANCE_SECONDS = 120
_ACTIVE_LEDGER_TTL_SECONDS = 1800
_MAX_RESPONSE_BYTES = 65_536
```

- [ ] **Step 1: Write failing secret/provider/plugin/boundary selftests**

Test the fixed URL/token/profile paths, no-follow token read, exact
`abuse-platform` profile record, monotonic 900-second send deadline, 120-second
cleanup allowance, `1800 >= 900 + 120`, `trust_env=False`, no redirects, one
provision per delegate, exact closed JSON, request reconstruction equality,
scenario XFF, mandatory observe after first request, one cleanup, response
caps, `Retry-After` `None|"600"`, and canary-free repr/error/traceback. Bind the
one-live-scenario assumption by rejecting `pytest-xdist`, `-n`, or another
parallel API invocation in the package/entrypoint architecture contract.

The boundary test must prove the thin pytest module imports no `httpx`, `os`,
`socket`, file API, or network API; all operational imports stay under
`harness/platform_abuse_provider.py`.

```python
def test_send_reconstructs_exact_public_request_before_network() -> None:
    factory, scripted = provider_factory_with_scripted_http()
    delegate = factory()
    scenario = delegate.provision_scenario(
        PlatformAbuseSeed(label="platform-rate-limit", flow="unknown-code-rate-limit")
    )
    admitted = verify_magic_link_request(scenario, "platform-rate-one")
    exchange = delegate.send(admitted)
    assert exchange.exchange.response.status_code == 401
    assert scripted.events == ("provision", "sut", "observe")
```

- [ ] **Step 2: Run Docker RED**

Run:

```text
docker compose -f docker-compose.local.yml build python-contract
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected: framework phase fails on missing provider/secret reader and boundary
contract; product phase does not start.

- [ ] **Step 3: Implement the provider**

Use two private `httpx.Client` instances with fixed base URLs, bounded
timeouts, `trust_env=False`, and `follow_redirects=False`. Do not extend the
general `HttpxTransport`: the provider must add XFF and preserve Retry-After.
Parse JSON with duplicate-key rejection and fixed byte caps. Build the existing
`OwnedVerifyScenario`, `HttpExchange`, `HttpDiagnostic`, and
`PlatformAbuseExchange` types.

Use `httpx` streaming for every response and stop after
`_MAX_RESPONSE_BYTES + 1`; calling `.content` on an unbounded response is not
admitted. Provision, observe, and cleanup each have one attempt only. The real
public verify call also has one attempt so a limiter or replay result cannot be
masked by transport retries.

Update harness docstrings so cleanup owns four public SUT keys plus only
server-derived user/session/refresh resources from the nonce ledger. Expose no
URL, token, IP, header, datastore ID, or transport constructor through the
suite-facing harness.

- [ ] **Step 4: Run Docker GREEN**

Run:

```text
docker compose -f docker-compose.local.yml build python-contract
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected: all framework selftests pass; known product contract failures remain
separately classified.

- [ ] **Step 5: Review and commit**

Request Terra/high reviews for Python type/collection correctness and secret/
network/redaction boundaries. Fix H/M findings.

```text
git commit --only -m "feat(qa): add fixed platform abuse provider" -- testframework/src/hub_testkit/runtime/secret_file.py testframework/src/hub_testkit/runtime/__init__.py testframework/src/hub_testkit/harness/platform_abuse.py testframework/src/hub_testkit/harness/platform_abuse_provider.py testframework/src/hub_testkit/pytest_plugin/plugin.py testframework/src/hub_testkit/pytest_plugin/platform_abuse_provider.py testframework/selftests/runtime/test_secret_file.py testframework/selftests/collection/pytest_plugin/test_plugin_state_access.py testframework/selftests/harness/test_platform_abuse_provider.py testframework/selftests/collection/pytest_plugin/test_platform_abuse_provider.py testframework/selftests/architecture/imports/test_platform_abuse_provider_boundary.py
```

### Task 6: Add immutable QA binary images and fixed runner bindings

**Files:**

- Modify: `testframework/docker/Dockerfile.qa`
- Modify: `testframework/docker/entrypoints/backend-unit.sh`
- Modify: `testframework/docker/entrypoints/python-qa.sh`
- Modify: `testframework/selftests/architecture/docker/test_root_compose_contract.py`
- Modify: `testframework/docker/README.md`
- Create: `backend/tests/qa/build_tags_test.go`

**Atomic dispatches:** Docker QA stages+build-tag contract; backend entrypoint
`qa-build` step+its exact contract; Python API plugin binding+its exact
contract; Docker README facts. Shared architecture-test edits are serialized.

**Interfaces:**

- Docker targets: `qa-secret-init`, `qa-schema`, `qa-backend`, `qa-control`.
- Backend result steps: `build`, `qa-build`, `vet`, `unit`, `race`.
- Only the literal `api)` branch loads
  `hub_testkit.pytest_plugin.platform_abuse_provider`.

- [ ] **Step 1: Write failing exact Docker/entrypoint contracts**

Extend the canonical architecture test first. Require one `qa-go-build` stage,
four literal `CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false
-tags="release qa"` commands, four scratch finals, numeric users, fixed binary
entrypoints, no package manager/shell/CA bundle, and no broad source copy.

Require this exact backend build command as the new step:

```sh
go build -trimpath -buildvcs=false -tags "release qa" \
  ./cmd/qa-secret-init ./cmd/qa-schema ./cmd/qa-backend ./cmd/qa-control
```

Require the provider plugin token only in the API branch.

Add a Go source contract that enumerates every `backend/cmd/qa-*` and
`backend/internal/qa/*_qa.go` path, requires `//go:build qa` as its first line,
and proves no non-test `cmd/server` source imports `backend/internal/qa`.

- [ ] **Step 2: Run Docker RED**

Run:

```text
docker compose -f docker-compose.local.yml build python-contract
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected: framework architecture phase fails against the current Dockerfile
and entrypoints.

- [ ] **Step 3: Implement Docker stages and runner commands**

Start `qa-go-build` from the generated-Ent `backend-unit` stage, switch to root
only for compiling `/out`, then copy one static binary into each scratch final.
Set `USER 0:0` only in `qa-secret-init`; set `USER 10001:10001` in the other
three. Every final has one fixed binary `ENTRYPOINT`.

Add `qa-build` before vet. Keep vet/unit/race tagged `hubtest qa`. Add the thin
provider plugin only to `python-qa.sh api)`.

- [ ] **Step 4: Run Docker GREEN**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit python-api
docker compose -f docker-compose.local.yml run --rm backend-unit
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected: images build; backend has five PASS steps; all framework selftests
pass before known product-contract results.

- [ ] **Step 5: Review and commit**

Request Terra/high review for build-tag reachability, scratch runtime needs,
binary health subcommands, exact runner calls, and absence of dynamic command
fallbacks. Fix H/M findings.

```text
git commit --only -m "build(qa): add narrow backend runtime images" -- backend/tests/qa/build_tags_test.go testframework/docker/Dockerfile.qa testframework/docker/entrypoints/backend-unit.sh testframework/docker/entrypoints/python-qa.sh testframework/selftests/architecture/docker/test_root_compose_contract.py testframework/docker/README.md
```

### Task 7: Replace policy v1 and wire the private root-Compose topology

**Files:**

- Create: `testframework/policies/root-compose-testframework.v2.json`
- Delete: `testframework/policies/root-compose-testframework.v1.json`
- Modify: `testframework/selftests/architecture/docker/test_root_compose_contract.py`
- Modify: `docker-compose.local.yml`

**Atomic dispatches:** v2 policy+closed parser/mutation tests; root Compose
topology+materialized-mapping tests. The second dispatch starts only after the
policy RED state is captured.

**Interfaces:**

Policy v2 adds these exact closed arrays while preserving the v1 retained,
retired, and forbidden arrays byte-for-byte:

```json
{
  "schema": "hub.qa.root-compose-testframework/v2",
  "root_compose": "docker-compose.local.yml",
  "public_services": [
    "backend-unit",
    "frontend-unit",
    "python-contract",
    "python-api",
    "python-integration",
    "python-e2e-smoke",
    "python-e2e-full",
    "python-load"
  ],
  "support_services": [
    "qa-secret-init",
    "qa-schema",
    "qa-backend",
    "qa-control"
  ],
  "infrastructure_services": [
    "qa-postgres",
    "qa-mongo",
    "qa-redis",
    "qa-nats",
    "qa-seaweedfs"
  ],
  "narrow_platform_infrastructure": ["qa-postgres", "qa-redis"],
  "secret_volumes": [
    "qa-secret-postgres",
    "qa-secret-jwt",
    "qa-secret-control"
  ]
}
```

The real file also contains the three preserved arrays in their existing
order; no second policy remains.

- [ ] **Step 1: Write failing v2 topology tests and policy**

Change the policy path/schema/closed-key parser, add exact service/volume
arrays, and add hostile mutations for every mount, dependency, public port,
user, network, security option, secret env, and ninth public target. Require
`network_mode: none` only for `qa-secret-init`; reject it elsewhere.

Run:

```text
docker compose -f docker-compose.local.yml build python-contract
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected: framework phase fails because root Compose has no support topology.

- [ ] **Step 2: Implement root Compose topology**

Add one reusable `x-qa-go-support` extension for the three non-root support
services. Add the four exact services and three named non-external volumes.
Every support service has `profiles: [qa]`; ordinary development startup cannot
select it.

Freeze these resource/security fields in the architecture test:

```text
qa-secret-init: read_only, user 0:0, cap_drop ALL, no-new-privileges,
  pids 64, cpu 0.50, memory 128m, restart no, stop grace 10s,
  /tmp noexec/nosuid/nodev 16m mode 0700
qa-schema: read_only, user 10001:10001, cap_drop ALL, no-new-privileges,
  init, pids 128, cpu 1.0, memory 512m, restart no, stop grace 30s,
  /tmp noexec/nosuid/nodev 64m uid/gid 10001 mode 0700
qa-backend and qa-control: the same non-root bounds as qa-schema,
  restart no, and a 5s/3s/20-retry health check with 5s start period
```

Use the exact mount matrix:

```text
qa-secret-init  postgres:rw jwt:rw control:rw
qa-postgres     postgres:ro
qa-schema       postgres:ro
qa-backend      postgres:ro jwt:ro
qa-control      postgres:ro control:ro
python-api      control:ro
```

The literal volume targets are:

```text
qa-secret-postgres -> /run/hub-qa/postgres
qa-secret-jwt      -> /run/hub-qa/jwt
qa-secret-control  -> /run/hub-qa/control
```

Every consumer points at the final `secret` regular file under its admitted
directory; no service mounts `/run/hub-qa` broadly.

`qa-postgres` must use `POSTGRES_PASSWORD_FILE` and must not contain
`POSTGRES_PASSWORD`. `qa-schema`, `qa-backend`, and `qa-control` get only
non-secret hosts/users/database plus secret-file paths.

Use dependency conditions:

```text
qa-secret-init completed -> qa-postgres
qa-postgres healthy + qa-redis healthy -> qa-schema
qa-schema completed + qa-redis healthy -> qa-backend
qa-schema completed + qa-redis healthy -> qa-control
qa-backend healthy + qa-control healthy -> python-api
```

Support services have no ports, reports bind, host bind, Docker socket, or
external network. `qa-secret-init` is the sole root service and has
`network_mode: none`; all long-running support services are UID/GID 10001.

- [ ] **Step 3: Run static and image GREEN checks**

Run:

```text
docker compose -f docker-compose.local.yml build qa-secret-init qa-schema qa-backend qa-control python-api python-contract
docker compose -f docker-compose.local.yml config
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected: Compose config succeeds; all framework selftests pass; all five
contour images build. Product contract failures remain separate.

- [ ] **Step 4: Independent security/topology review**

Request Terra/high review of the fully materialized Compose mapping, exact
volume matrix, secret initializer root exception, dependency lifecycle,
absence of NATS edge, and no public port. Fix all H/M findings.

- [ ] **Step 5: Commit**

```text
git commit --only -m "build(qa): wire narrow platform runtime" -- docker-compose.local.yml testframework/policies/root-compose-testframework.v1.json testframework/policies/root-compose-testframework.v2.json testframework/selftests/architecture/docker/test_root_compose_contract.py
```

### Task 8: Synchronize source-status documentation

**Files:**

- Modify: `testframework/README.md`
- Modify: `testframework/PACKAGE-CONTRACT.md`
- Modify: `testframework/PIPELINE.md`
- Modify: `testframework/TESTING_GUIDE.md`
- Modify: `testframework/docker/README.md`
- Modify: `docs/TESTING.md`
- Modify: `docs/test_plan.md`
- Modify: `docs/stack_and_tools.md`
- Modify: `docs/onboarding/qa-automation.md`
- Modify: `testframework/cases/platform/TC-PLATFORM-0001.md`
- Modify: `testframework/cases/platform/TC-PLATFORM-0002.md`

**Interfaces:** Documentation must state `SOURCE_AUTHORED` / `NOT_RUN`, v2
policy, exact eight public services, four private support services, two real
Platform cases, PostgreSQL+Redis dependency, and dynamic session/refresh
cleanup beyond the four public Redis keys.

- [ ] **Step 1: Update facts without runtime promotion**

Remove statements that the Platform provider is absent. Do not say PASS,
ready, stable, or green. Keep other providers absent and the aggregate API
target `PARTIAL_FAIL`/pending runtime.

- [ ] **Step 2: Run static documentation checks**

Run `git diff --check -- <eleven owned docs>` and scoped `rg` checks for v1,
provider-absent, ninth public targets, NATS dependency, and four-key-only
cleanup wording. Exclude the superseded 2026-08-15 historical plan from the v1
scan. Expected: no stale active claim in owned docs or the two active cases.

- [ ] **Step 3: Review and commit**

Request Terra/medium code-to-doc fact review. Fix all H/M findings.

```text
git commit --only -m "docs(qa): document narrow platform contour" -- testframework/README.md testframework/PACKAGE-CONTRACT.md testframework/PIPELINE.md testframework/TESTING_GUIDE.md testframework/docker/README.md docs/TESTING.md docs/test_plan.md docs/stack_and_tools.md docs/onboarding/qa-automation.md testframework/cases/platform/TC-PLATFORM-0001.md testframework/cases/platform/TC-PLATFORM-0002.md
```

### Task 9: Execute Docker runtime twice and promote only proven status

**Files:**

- Modify after evidence only: `testframework/README.md`
- Modify after evidence only: `testframework/PACKAGE-CONTRACT.md`
- Modify after evidence only: `testframework/PIPELINE.md`
- Modify after evidence only: `testframework/TESTING_GUIDE.md`
- Modify after evidence only: `testframework/docker/README.md`
- Modify after evidence only: `docs/TESTING.md`
- Modify after evidence only: `docs/test_plan.md`
- Modify after evidence only: `docs/stack_and_tools.md`
- Modify after evidence only: `docs/onboarding/qa-automation.md`
- Modify after evidence only: `.planning/STATE.md`

**Interfaces:** Runtime evidence must distinguish framework health, the two
Platform node results, aggregate API status, cleanup result, and the second-run
isolation result.

- [ ] **Step 1: Verify clean task-owned Docker state**

Resolve the exact Compose project from the current worktree. Obtain approval
before destructive cleanup, then remove only that project's containers and
volumes. Do not target another Compose project, development volume, or broad
Docker inventory.

- [ ] **Step 2: Build and run backend/framework gates**

Run:

```text
docker compose -f docker-compose.local.yml build backend-unit qa-secret-init qa-schema qa-backend qa-control python-api python-contract
docker compose -f docker-compose.local.yml run --rm backend-unit
docker compose -f docker-compose.local.yml run --rm python-contract
```

Expected:

- backend build/qa-build/vet/unit/race all PASS;
- framework selftests all PASS;
- product contract failures, if still present, match independently classified
  Swagger/product defects and are not framework/provider failures.

- [ ] **Step 3: Run the full fixed API target first time**

Run:

```text
docker compose -f docker-compose.local.yml run --rm python-api
```

Expected aggregate exit: non-zero until other providers exist. Required
sub-results:

- `TC-PLATFORM-0001` PASS;
- `TC-PLATFORM-0002` PASS;
- no Platform setup, config, network, observe, cleanup, secret, or traceback
  failure;
- every other failure is separately attributed to an absent provider or an
  already-classified domain issue.

- [ ] **Step 4: Run the full fixed API target second time**

Run the identical command again. Required: both exact Platform nodes PASS with
fresh nonces; no 429 leakage into replay, stale session, duplicate email,
missing ledger, or cleanup collision.

- [ ] **Step 5: Independent evidence and security review**

Dispatch Terra/high reviewers to inspect the exact Docker outputs and live
Compose configuration. They must confirm the two node results, secret mount
matrix, no public port, no NATS edge, exact cleanup, and honest aggregate
status. Any H/M finding returns to the owning task using systematic debugging;
do not patch around a failing oracle.

- [ ] **Step 6: Promote only proven documentation**

If and only if both runs satisfy the required sub-results, update the ten
status files to say the two Platform nodes are Docker `PASS` twice and the
aggregate API target remains `PARTIAL_FAIL`. Include exact run timestamps or
run identifiers available from the captured output. Do not promote unrelated
targets.

- [ ] **Step 7: Final verification and commit**

Run scoped `git diff --check`, source-contract scans, and one final independent
review of all task commits. Commit only evidence-backed docs:

```text
git commit --only -m "docs(qa): record platform API runtime evidence" -- testframework/README.md testframework/PACKAGE-CONTRACT.md testframework/PIPELINE.md testframework/TESTING_GUIDE.md testframework/docker/README.md docs/TESTING.md docs/test_plan.md docs/stack_and_tools.md docs/onboarding/qa-automation.md .planning/STATE.md
```

Do not merge or push. Hand the branch and exact remaining non-Platform blockers
back to the user.
