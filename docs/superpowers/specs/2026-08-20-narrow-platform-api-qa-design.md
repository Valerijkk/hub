# Narrow Platform API QA Contour Design

**Date:** 2026-08-20
**Status:** architecture approved in chat; written specification awaiting review
**Runtime truth:** `NOT_RUN`

## Decision

Implement the smallest real backend-backed QA contour that can execute the two
existing platform-abuse API cases:

- `TC-PLATFORM-0001`: three invalid magic-link verifications return `401`, the
  fourth request is rejected by the real global auth limiter with `429` and
  `Retry-After: 600`;
- `TC-PLATFORM-0002`: a valid one-time magic-link code returns a real token
  pair, and replay of the same code returns `401`.

The contour reuses the production auth handler, auth service, Ent repository,
session service, Redis rate-limit middleware, and Redis magic-link logic. It
does not start the ordinary `cmd/server`, does not fake HTTP responses, and
does not expose test provisioning through a product route.

The repository keeps exactly the current eight public root-Compose QA service
names. The existing `python-api` service remains the public API-suite command;
this slice only makes its two Platform nodes real. Until the other API harness
providers exist, the aggregate `python-api` target remains `PARTIAL_FAIL` even
when both Platform nodes pass.

## Why This Is Architectural

The change introduces a new internal runtime boundary: a QA-tagged SUT, a
schema initializer, a narrow state-control service, launcher-side provider
composition, and their root-Compose dependency graph. It changes neither the
public product API nor the public QA command list, but it changes how multiple
components fit together and therefore requires a versioned design and an
implementation plan.

## Considered Approaches

### 1. QA-tagged narrow SUT and control plane — selected

Build dedicated `qa`-tagged binaries that wire only the product behavior under
test. Provision and cleanup happen through a separate internal control service
whose API is closed to the two Platform scenarios.

Benefits:

- exercises the real public Fiber route and production auth/Redis/Ent code;
- keeps `DEV_MODE=false` and avoids `/api/dev/*`;
- has a small dependency graph: PostgreSQL and Redis only;
- cannot be linked into the normal release binary without the `qa` tag;
- permits exact, ledger-owned cleanup without giving pytest raw DB or Redis
  credentials.

Cost: four small support binaries/services and a versioned Compose-contract
amendment are required.

### 2. Start the ordinary product server — rejected

The ordinary server wires unrelated MongoDB, NATS/JetStream, S3, outbox,
messenger, and other modules. Its automatic Ent schema creation is currently
coupled to `DEV_MODE=true`, while these cases require `DEV_MODE=false` so the
real Redis limiter is active. This approach is not narrow and would either
overprovision the contour or weaken the behavior being tested.

### 3. Fake transport or direct pytest datastore access — rejected

A fake transport cannot prove real routing, middleware order, rate limiting,
one-time deletion, session creation, or response mapping. Giving pytest raw
PostgreSQL/Redis access would move privileged provisioning into the test suite,
break the opaque-harness boundary, and make broad cleanup possible.

## Goals

- Execute the two existing Platform tests against a real internal HTTP SUT.
- Preserve the current public route and response semantics.
- Keep the server in non-development mode with no OTP disclosure or delivery.
- Isolate every scenario by a server-generated nonce, email, user, source IP,
  and exact ownership ledger.
- Delete only resources owned by that ledger; never scan, truncate, flush, or
  delete by caller-supplied datastore identifiers.
- Keep secrets out of environment variables, URLs, logs, reports, exception
  messages, object representations, and the repository.
- Keep all build/test/validator execution Docker-only.

## Non-goals

- Making the full `python-api` target green in this slice.
- Adding a ninth public Compose test service or arbitrary profile arguments.
- Starting the complete Hub backend.
- Exercising email delivery, MongoDB, NATS, S3, outbox, cron, frontend BFF, or
  unrelated API domains.
- Adding a generic QA command/control API.
- Changing product auth, magic-link, rate-limit, or session behavior merely to
  simplify the harness.
- Creating a host-side runner, additional Compose file, Docker socket
  orchestration, or a direct host fallback.

## Authority Amendment

The current `hub.qa.root-compose-testframework/v1` policy closes the service
set before internal SUT-support services existed. Implementation replaces it
atomically with `hub.qa.root-compose-testframework/v2`:

- `public_services` stays exactly the existing eight names;
- `support_services` is exactly `qa-secret-init`, `qa-schema`, `qa-backend`,
  and `qa-control`;
- `infrastructure_services` remains the existing disposable QA infrastructure
  list; the narrow contour depends only on `qa-postgres` and `qa-redis`;
- retained Docker paths remain explicit and closed;
- the root-Compose selftest verifies exact service membership, dependency
  edges, build targets, mounts, secrets, networks, security settings, and
  absence of public ports.

The v1 policy is removed in the same atomic change after every consumer has
moved to v2. Two policy files may not coexist as competing authority.

## Runtime Topology

```text
qa-secret-init ───────────────┐
       │                      │
       ├─> qa-postgres ─> qa-schema ─┐
       │                             ├─> qa-backend ─┐
       └─────────────────────────────┤              ├─> python-api
                         qa-redis ───┴─> qa-control ┘
```

`qa-backend`, `qa-control`, and `python-api` share only the private
`qa-internal` network. No new service publishes a host port. `qa-secret-init`
has no network. The normal development services, databases, and volumes are
not dependencies.

The minimum dependency order is:

1. `qa-secret-init` completes successfully;
2. `qa-postgres` and `qa-redis` become healthy;
3. `qa-schema` completes successfully;
4. `qa-backend` and `qa-control` become healthy;
5. `python-api` starts pytest.

No NATS dependency is permitted for this contour. The production session
service already treats a nil event publisher as a no-op; token issuance and
session persistence still use real PostgreSQL and Redis.

## Components

### `qa-secret-init`

A `//go:build qa` one-shot binary creates three project-local random secrets:

- PostgreSQL password;
- JWT signing secret;
- QA-control bearer token.

Each secret lives in a separate non-external, project-local named Docker
volume. The initializer is the sole narrow root exception in this design: it
runs once as UID/GID `0:0` with `cap_drop: [ALL]`, `no-new-privileges`, no
network, a read-only root filesystem, and only the three secret volumes mounted
read-write. It needs no `chown`: it creates root-owned regular files directly
in the root-owned volumes with mode `0444`. The files contain bounded
lower-case hexadecimal bytes plus one final LF and are never symlinks. Every
other service is non-root and mounts only its listed secret volume read-only.

On a repeated Compose invocation the initializer validates and reuses an
existing valid project-local secret; it never silently replaces a secret while
consumers may still be running. Each one-shot service uses `restart: "no"`.
Named volumes are disposable with the QA Compose project but are not claimed
to be in-memory or task-scoped.

These are disposable QA credentials, not production credentials. They are not
committed, passed as build arguments, or exported through Compose environment
variables. Docker support for the selected named-volume creation, reuse, and
read-only distribution semantics is a runtime acceptance condition, not a
source-review fact.

The closed mount matrix is:

| Service | Secret mounts |
|---|---|
| `qa-secret-init` | PostgreSQL, JWT, and control-token volumes, read-write |
| `qa-postgres` | PostgreSQL volume, read-only |
| `qa-schema` | PostgreSQL volume, read-only |
| `qa-backend` | PostgreSQL and JWT volumes, read-only |
| `qa-control` | PostgreSQL and control-token volumes, read-only |
| `python-api` | control-token volume, read-only |

No other service may mount any of the three volumes. In particular, Python and
`qa-control` may not mount the JWT volume, and `qa-backend` may not mount the
control-token volume.

`qa-postgres` depends on successful secret initialization and uses only
`POSTGRES_PASSWORD_FILE`; `POSTGRES_PASSWORD` is forbidden. `qa-schema`,
`qa-backend`, and `qa-control` accept only an exact
`HUB_QA_POSTGRES_PASSWORD_FILE` path and assemble their credential-bearing DSN
inside the process. They reject a password or credential-bearing DSN supplied
through the environment.

### `qa-schema`

A `//go:build qa` one-shot binary connects only to `qa-postgres` and runs the
generated Ent schema creation directly. It is not the ordinary server with
`DEV_MODE=true`.

Startup requires all of:

- `HUB_QA_MODE=1`;
- `ENVIRONMENT=qa`;
- `DEV_MODE=false`;
- `HUB_QA_SCHEMA_APPLY=1`;
- exact PostgreSQL host `qa-postgres`;
- a valid regular secret file.

Any missing, malformed, or conflicting input terminates non-zero with a
constant redacted error.

### `qa-backend`

A `//go:build qa` binary exposes only:

- `GET /livez`;
- `GET /readyz`;
- `POST /api/auth/magic-link/verify`.

The public endpoint is registered by constructing the production auth handler
but binding only its exported `VerifyMagicLink` method directly. The broad
`RegisterRoutes` method is not called because it would expose unrelated auth
routes. The narrow SUT reproduces the production auth-specific order:

1. the verify-email limiter, `10` requests per `10` minutes;
2. the global `/api/auth` IP limiter, `3` requests per `10` minutes;
3. `auth.Handler.VerifyMagicLink`;
4. real auth/session/magic-link services using Ent and Redis.

It does not claim to reproduce unrelated ordinary-server middleware such as
maintenance mode, HTTP audit, the global 100 requests/second limiter, or API-key
authentication; those dependencies are outside the two frozen cases.

It runs with `DEV_MODE=false`, no Resend client, and a nil
`*natsclient.Client` as the production session service's optional publisher.
There is no raw NATS connection, in-memory substitute, or fake. Nil delivery
never exposes an OTP because response disclosure is controlled by the
handler's explicit development-mode flag. The service does not register
send-code, dev-login, session-management, admin, or any other product route.

The app uses the production Fiber error handler, struct validation, and trusted
private-proxy rules. The Python provider sends one server-issued stable
`X-Forwarded-For` value per scenario; the global limiter therefore uses a
scenario-specific real `c.IP()` identity.

### `qa-control`

A separate `//go:build qa` binary owns privileged setup, observation, and exact
cleanup. It is not imported by `cmd/server` and exposes only these internal
routes:

- `GET /__hub_qa/v1/livez`;
- `GET /__hub_qa/v1/readyz`;
- `POST /__hub_qa/v1/platform-abuse/scenarios`;
- `POST /__hub_qa/v1/platform-abuse/scenarios/{nonce}/observe`;
- `DELETE /__hub_qa/v1/platform-abuse/scenarios/{nonce}`.

Mutation and observation routes require the file-backed control token. Token
comparison is constant-time. Missing and wrong tokens produce the same
redacted response and reveal no route details. Responses use
`Cache-Control: no-store`; request bodies are closed, size-bounded JSON with
unknown and duplicate fields rejected.

The API accepts only the two fixed flow names. It has no endpoint for SQL,
Redis commands, arbitrary keys, arbitrary IDs, arbitrary URLs, shell commands,
or generic fixture creation.

### Python provider

`hub_testkit.pytest_plugin.platform_abuse_provider` is loaded only by the fixed
`python-api` entrypoint. It installs the existing
`PlatformAbuseHarnessFactory` exactly once and uses:

- fixed control URL `http://qa-control:3001`;
- fixed SUT URL `http://qa-backend:3000`;
- fixed token file path under `/run/hub-qa`;
- fixed profile-policy path
  `/opt/hub/testframework/policies/profiles.toml`, selecting only the existing
  `abuse-platform` deadline record;
- HTTP clients with environment proxy inheritance disabled;
- the framework's mandatory redactor for all diagnostics.

The delegate rejects any method, path, body shape, or header request outside
the existing Platform builders. Tests cannot supply a base URL, source IP,
control token, datastore identifier, or cleanup key. The provider adds the
scenario's stable `X-Forwarded-For` and a nonce-bound correlation ID only to
the real SUT request.

Other API harness factories remain absent and fail closed. Loading this
provider does not make identity, authorization, files, or health-fake cases
backend-backed.

## Scenario Provisioning

For either flow the control service generates:

- a 32-character lower-case hexadecimal nonce;
- a unique lower-case QA email;
- a six-digit code;
- a stable RFC 2544 documentation-network source IP;
- `sha256(source_ip)[:16]` as the expected limiter hash.

It stores an internal bounded ledger under the exact Redis key
`hubqa:platform-abuse:<nonce>`. The ledger is created with `SETNX`, has a
bounded TTL, contains only server-generated ownership data, and is never
accepted from the caller.

For `unknown-code-rate-limit`, no user and no valid OTP are created. The code
returned to the opaque provider is deliberately invalid for the absent
`magiclink:<email>` entry.

For `one-time-code-replay`, the control service creates one active Ent user
with no organization dependency and writes one real `magiclink:<email>` value
with the production TTL. It records the generated user ID only in the internal
ledger.

The existing `OwnedVerifyScenario.owned_state_entries` remains exactly these
four SUT Redis names, in order:

1. `rl:auth:ip:<client-hash>`;
2. `rl:auth-verify-email:email:<email>`;
3. `magiclink:fails:<email>`;
4. `magiclink:<email>`.

This four-key tuple describes the fixed Redis state under direct test. It is
not the complete cleanup ledger: a successful login also creates a dynamic Ent
session and `refresh:<hash>` Redis key, which remain private to the control
service.

Implementation updates the existing harness protocol and factory docstrings to
state this boundary explicitly: public ownership remains the four fixed SUT
keys, while cleanup additionally owns only the user/session/refresh artifacts
derived by the server from its nonce ledger. The docstrings may not continue to
promise four-key-only cleanup.

## Request and Observation Flow

1. The fixture asks the provider for one fresh harness.
2. The harness provisions a scenario through `qa-control`.
3. The existing builder creates the public verify request.
4. The provider validates the request and sends it to `qa-backend` with the
   scenario's stable source IP.
5. After the first public request, the provider calls `observe` once.
6. `observe` proves that the exact global limiter key for the expected
   server-observed IP exists. A client assertion alone is insufficient.
7. Existing oracles validate public statuses, `Retry-After`, token-pair shape,
   replay denial, and redaction.
8. Fixture `finally` invokes cleanup even after an assertion failure.

The real public request never passes through `qa-control` and cannot be
answered by it.

## Exact Cleanup

Cleanup is driven only by the server-held nonce ledger. The caller supplies
the nonce capability but never supplies user IDs, session IDs, refresh hashes,
or Redis keys to delete.

For a replay scenario the control service:

1. loads the exact ledger;
2. queries sessions only for the ledger-owned user ID;
3. deletes each exact `refresh:<stored-hash>` key;
4. deletes those exact session rows;
5. deletes the exact owned user row;
6. deletes the four exact SUT Redis keys;
7. verifies absence of those exact resources;
8. replaces the active ledger with a `cleaned` tombstone whose TTL is exactly
   five minutes.

For the unknown-code flow only the four exact SUT Redis keys and ledger state
are involved. Repeated cleanup of a valid cleaned nonce is idempotent. The
active ledger TTL is exactly 30 minutes. The provider loads the existing
`abuse-platform` policy deadline of 900 seconds, enforces it with monotonic
time and bounded HTTP calls, and reserves a separate 120-second cleanup
deadline. The active ledger therefore outlives the maximum admitted scenario
lifetime of 17 minutes. This does not invent a 20-minute outer deadline for the
whole `python-api` command. Cleanup fails closed if an active ledger is missing
or expired. Partial cleanup preserves the active ledger and remains retryable;
it returns non-zero to pytest and may not report success or replace the ledger
with a tombstone before exact absence verification succeeds.

Forbidden cleanup operations include `SCAN`, `KEYS`, `FLUSHDB`, database
truncate, prefix deletion, table-wide queries, bucket deletion, and any delete
target supplied by pytest.

An ungraceful container kill can bypass fixture cleanup. The data stores are
therefore disposable, private, and nonce-isolated; such a run is failed and
requires destruction of the QA Compose project before evidence can be
accepted. The implementation must not claim crash recovery it does not prove.

## Startup and Isolation Gates

Every QA binary is excluded without the `qa` build tag and is compiled for
acceptance with `-tags "release qa"`. Every long-running QA binary also
requires exact runtime values `HUB_QA_MODE=1`, `ENVIRONMENT=qa`, and
`DEV_MODE=false`.

Configuration accepts only the closed internal hosts and ports required by the
service. It rejects loopback aliases, host-gateway names, external DNS names,
URLs with userinfo, proxy environment fallbacks, and secret values supplied in
environment variables. Regular-file checks hold the file handle while reading,
reject symlinks/reparse points and oversized or malformed content, and expose
only constant error text.

All new services are:

- behind the `qa` profile;
- connected only to `qa-internal` when networking is required;
- without `ports`, host bind mounts, Docker socket, or external network;
- read-only at the root filesystem;
- non-root, except for the isolated one-shot `qa-secret-init` described above;
- `cap_drop: [ALL]` with `no-new-privileges:true`;
- PID, CPU, memory, timeout, and health-check bounded;
- limited to explicit UID-owned tmpfs paths;
- independent of the Python report bind, except `python-api` itself.

## Docker Assets

No new Dockerfile or shell entrypoint is added. The existing
`testframework/docker/Dockerfile.qa` gains four named build/final targets for
the four QA binaries, each with a fixed binary `ENTRYPOINT`. Root Compose owns
their fixed service configuration.

The existing `python-qa.sh` gains only an exact API-target plugin registration;
it still accepts no arbitrary command string or arbitrary pytest path. The
retained Docker-path list therefore does not grow.

The backend unit entrypoint must explicitly compile and test the QA-tagged
packages with `-tags "hubtest qa"` or equivalent closed commands. Untagged
`go build ./...` is not evidence that the QA binaries compile.

## Error and Redaction Contract

- Public SUT errors retain the product wire shape.
- Internal health responses contain only liveness/readiness state.
- Internal auth/config/provision/cleanup failures use constant generic text.
- Go and Python errors suppress wrapped hostile values from message, cause,
  context, traceback, logs, and representations.
- Control token, JWT secret, PostgreSQL password, OTP, email, access token,
  refresh token, session hash, DSN, source IP, nonce, and raw response bodies
  are treated as sensitive.
- Provision and cleanup are not automatically retried by a generic HTTP
  retry policy. Idempotency is explicit at the nonce-ledger boundary.
- A cleanup or observation failure makes the test run non-zero even if public
  assertions passed.

## Tests-first Implementation Contract

Implementation begins only after this spec and its implementation plan are
approved. Tests are authored before each production slice.

### Static and architecture tests

- Replace the root-Compose v1 policy with the closed v2 policy.
- Assert the exact eight public services and exact four support services.
- Assert only PostgreSQL and Redis dependencies for this contour.
- Assert exact build tags, Docker targets, fixed binary entrypoints, internal
  network, secret mounts, dependency conditions, health checks, and resource
  bounds.
- Reject public ports, host binds, Docker socket, broad source copies, proxy
  inheritance, arbitrary commands, alternate URLs, and ninth public targets.
- Prove normal release builds cannot contain QA command/control symbols.

### Go unit tests

- strict runtime gates and topology allowlist;
- held regular-file secret parsing and redacted hostile failures;
- secret initializer create/reuse/reject behavior;
- exact verify-email then auth-IP middleware order, direct
  `VerifyMagicLink`-only registration, and `DEV_MODE=false` handler
  construction;
- nonce, email, code, IP, and client-hash generation;
- flow-specific provisioning;
- ledger `SETNX`, TTL, state transitions, and tombstone idempotency;
- exact binding of the `abuse-platform` 900-second policy deadline, monotonic
  enforcement, bounded requests, and 120-second cleanup allowance;
- exact session/refresh/user/four-key cleanup plan;
- rejection of arbitrary keys, IDs, commands, fields, routes, and broad
  datastore operations;
- constant-time control authentication and equal wrong/missing-token response;
- compile assertions for the real repository/service interfaces used.

### Python selftests

- provider installs exactly once only for the API target;
- fixed control/SUT URLs and token path;
- `trust_env=False` and no caller-controlled headers or base URL;
- exact request/path/body admission;
- first-request observation is mandatory and non-vacuous;
- exact four-key scenario tuple remains stable;
- all secrets/canaries are absent from repr, exception, traceback, and report
  diagnostics;
- cleanup runs in `finally`, is invoked once, and cleanup failure fails the
  fixture;
- protocol and factory documentation says cleanup owns the four fixed SUT keys
  plus only server-derived session/refresh/user artifacts from the nonce
  ledger;
- other missing domain providers continue to fail closed.

### Docker-only runtime acceptance

1. Build the amended images through root Compose.
2. Run the framework selftests in their existing Docker target.
3. Start the narrow dependency graph and prove every readiness gate.
4. Run the full fixed `python-api` target once.
5. Verify the two exact Platform nodes pass against `qa-backend`; record the
   expected aggregate non-zero result caused only by still-missing providers.
6. Run the same target a second time and prove both Platform nodes pass again
   with fresh nonces and no cross-run limiter/session interference.
7. Independently review the generated evidence and classify every non-Platform
   failure before any readiness statement.

No source/static review is runtime evidence. Until this sequence is captured,
the contour remains `NOT_RUN`.

## Documentation and Status Updates

The implementation updates the root Compose policy, package contract, Docker
README, testing guide, active test-plan status, and case wording that currently
describes the four Redis entries as the whole cleanup boundary. Status language
must distinguish:

- `SOURCE_AUTHORED` / `STATIC_ACCEPT` for reviewed code;
- two Platform nodes `PASS` only after Docker evidence;
- aggregate `python-api` `PARTIAL_FAIL` until every active API provider exists;
- no claim that the complete API framework or product is ready.

## Acceptance Criteria

- The public Compose QA service list is still exactly eight names.
- The four internal support services are versioned, closed, unpublished, and
  unavailable to normal development startup.
- Only PostgreSQL and Redis are required for the narrow contour.
- Both existing Platform nodes exercise the real public Fiber endpoint and
  production auth/Ent/Redis behavior.
- `DEV_MODE=false` is enforced and no OTP is delivered, logged, or disclosed.
- The fourth unknown-code request is blocked by the real global IP limiter and
  reports `Retry-After: 600`.
- Valid verification creates a real session/token pair; replay is denied.
- Server observation proves the rate-limit key corresponds to the SUT-observed
  client IP.
- Cleanup removes the dynamic session/refresh/user state and exact four SUT
  Redis keys without a broad operation or caller-supplied delete target.
- Secrets and sensitive scenario data do not appear in diagnostics or reports.
- Normal release artifacts contain no QA control-plane code or route.
- Static contracts, QA-tagged units, two sequential Docker runs, and an
  independent evidence review complete before the two nodes are called PASS.
- The aggregate `python-api` status remains honest and non-green until its
  other concrete providers are implemented.
