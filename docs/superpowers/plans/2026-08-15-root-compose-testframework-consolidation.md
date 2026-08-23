# Root Compose Test Framework Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the unfinished PowerShell/G0 runner and duplicate Compose
manifests with eight fixed test services in root `docker-compose.local.yml`
while preserving all greenfield framework source, cases, policies, schemas,
and genuine units.

**Architecture:** One multi-stage `testframework/docker/Dockerfile.qa` produces
separate Go, Node, and Python images. Root Compose is the sole test execution
manifest; each service has the public command shape
`docker compose run --rm <service>`. QA-only infrastructure is isolated from
development data and reusable Python behavior remains in
`testframework/src/hub_testkit/**`.

**Tech Stack:** Docker Compose v2, multi-stage Dockerfile, POSIX shell, Go 1.25,
Node/Playwright base, Python 3.13, uv, pytest, Vitest, PostgreSQL 17, MongoDB
8.0, Redis 7.4, NATS 2.11 JetStream, SeaweedFS 3.97.

## Global Constraints

- Windows host may read/edit files and use Git only; no host tests, validators,
  linters, builds, Python, Go, Node, Vite, pytest, or PowerShell test targets.
- Do not read or modify `pro-qa-ai/**`.
- Preserve current uncommitted project-outbox source and accept it before any
  consolidation deletion.
- No Docker socket, nested Compose, arbitrary shell command, host fallback,
  direct cloud egress, or shared development database/volume in a QA service.
- Root `docker-compose.local.yml` is the only test execution manifest.
- Public services are exactly `backend-unit`, `frontend-unit`,
  `python-contract`, `python-api`, `python-integration`,
  `python-e2e-smoke`, `python-e2e-full`, and `python-load`.
- All tests are authored before their implementation boundary. If no accepted
  Docker image exists yet, record `NOT_RUN_BOOTSTRAP`; never claim an executed
  RED or PASS from static review.
- Runtime readiness remains `NOT_READY` / `NOT_RUN` until the Docker-only
  verification task completes.
- Every commit is atomic and contains no AI attribution or `Co-authored-by`.

---

### Task 1: Preserve and accept the project-outbox draft

**Files:**
- Review/commit: `testframework/selftests/architecture/imports/test_project_outbox_boundary.py`
- Review/commit: `testframework/src/hub_testkit/integration/project_outbox/{__init__,models,builders,api,oracles}.py`
- Review/commit: `testframework/src/hub_testkit/harness/{project_outbox,pytest_project_outbox}.py`
- Modify/commit: `testframework/src/hub_testkit/harness/__init__.py`
- Review/commit: `testframework/tests/integration/redis-nats/project_outbox/{conftest,test_project_outbox}.py`
- Review/commit: `testframework/cases/outbox/TC-OUTBOX-0001.md` through `TC-OUTBOX-0005.md`

**Interfaces:**
- Consumes: accepted receipt `f0023702` and current backend project/outbox
  source facts.
- Produces: five accepted source-only `TASK-211` cases and a clean worktree for
  consolidation.

- [ ] **Step 1: Reconcile the exact 16-path allowlist**

Run only read-only Git/status and verify that the draft changes exactly the 16
paths listed by `f0023702`. Any extra path is a blocker; do not reset or delete
it.

- [ ] **Step 2: Perform independent source review**

Review exact function-to-TC mapping, nonconstructible opaque models, bounded
duplicate-aware JSON, `201/409/401/400` source facts, InterestPolicy durable
observer precreation, suppressed errors, and quiescent LIFO cleanup. Runtime
remains `NOT_RUN_BOOTSTRAP`.

- [ ] **Step 3: Fix every HIGH/MEDIUM finding test-first**

Change only the 16 owned paths. Each fix begins in
`test_project_outbox_boundary.py` or the five literal suite/case files before
the corresponding source boundary.

- [ ] **Step 4: Commit the accepted source slice**

```text
git add -- <the exact 16 paths>
git diff --cached --check
git commit -m "test(testframework): add project outbox integration cases"
```

Expected: one 16-path commit and no remaining outbox worktree changes.

---

### Task 2: Author the root-Compose contract before implementation

**Files:**
- Create: `testframework/policies/root-compose-testframework.v1.json`
- Create: `testframework/selftests/architecture/docker/test_root_compose_contract.py`
- Modify: `testframework/pyproject.toml`

**Interfaces:**
- Consumes: repository root, the approved design, and the exact service/asset
  names below.
- Produces:
  `load_root_compose_contract(repo_root: Path) -> Mapping[str, object]` in the
  selftest module and a closed migration policy consumed only by the
  source-contract test.

- [ ] **Step 1: Write the closed migration policy**

Create canonical JSON with exact ordered top-level keys
`schema`, `root_compose`, `public_services`, `retained_docker_paths`,
`retired_paths`, and `forbidden_tokens`. Freeze:

```json
{
  "schema": "hub.qa.root-compose-testframework/v1",
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
  "retained_docker_paths": [
    "testframework/docker/Dockerfile.qa",
    "testframework/docker/README.md",
    "testframework/docker/entrypoints/backend-unit.sh",
    "testframework/docker/entrypoints/frontend-unit.sh",
    "testframework/docker/entrypoints/python-qa.sh",
    "testframework/docker/vitest.frontend-unit.config.ts"
  ]
}
```

`retired_paths` contains every exact path deleted in Task 6. Forbidden active
tokens are `/qa-gate`, `invoke-qa-bootstrap`, `invoke-qa-gate`,
`compose.unit.yaml`, `compose.g0-bootstrap.yaml`, and `docker.sock`.

- [ ] **Step 2: Write the failing duplicate-aware YAML/JSON source contract**

Add `PyYAML>=6,<7` to the `framework` dependency group. In
`test_root_compose_contract.py`, define a `yaml.SafeLoader` subclass whose
mapping constructor raises a constant error on duplicate keys. The tests must
assert:

```python
EXPECTED_SERVICES = (
    "backend-unit",
    "frontend-unit",
    "python-contract",
    "python-api",
    "python-integration",
    "python-e2e-smoke",
    "python-e2e-full",
    "python-load",
)

def test_root_compose_exposes_exact_fixed_qa_services() -> None: ...
def test_qa_services_have_fixed_commands_and_security_options() -> None: ...
def test_qa_infrastructure_is_unpublished_and_disposable() -> None: ...
def test_docker_directory_contains_only_manifested_assets() -> None: ...
def test_retired_control_plane_is_absent_and_unreferenced() -> None: ...
def test_duplicate_yaml_keys_are_rejected() -> None: ...
```

The parser accepts only the fixed root Compose and policy paths, rejects
symlinks/non-regular files, caps each read at 1 MiB, decodes strict UTF-8, and
wraps all errors in constant `from None` messages.

- [ ] **Step 3: Record the source-level RED**

Do not execute Python on the host. Static review must confirm that the tests
currently fail their stated path/service/absence predicates. Record status
`NOT_RUN_BOOTSTRAP`; the executable RED occurs in Task 5 after the Python image
exists.

- [ ] **Step 4: Commit tests and policy only**

```text
git add -- testframework/policies/root-compose-testframework.v1.json testframework/selftests/architecture/docker/test_root_compose_contract.py testframework/pyproject.toml
git commit -m "test(testframework): define root compose contract"
```

---

### Task 3: Build the consolidated Docker execution assets

**Files:**
- Create: `testframework/docker/Dockerfile.qa`
- Modify: `testframework/docker/entrypoints/backend-unit.sh`
- Modify: `testframework/docker/entrypoints/frontend-unit.sh`
- Create: `testframework/docker/entrypoints/python-qa.sh`
- Retain: `testframework/docker/vitest.frontend-unit.config.ts`
- Modify: `testframework/docker/README.md`
- Create: `testframework/src/hub_testkit/py.typed`

**Interfaces:**
- Consumes: fixed service names and pytest targets from Task 2.
- Produces: Docker build targets `backend-unit`, `frontend-unit`,
  `python-tests`, and maintenance target `python-lock`; entrypoint
  `/opt/hub/bin/python-qa <contract|api|integration|e2e-smoke|e2e-full|load>`.

- [ ] **Step 1: Extend the source-contract tests first**

Require exactly one Dockerfile, three shell entrypoints, the Vitest config, and
the Docker README. Reject `curl | sh`, unpinned `latest`, Docker socket mounts,
shell `eval`, `sh -c` command injection, host package directories, and arbitrary
Python target arguments.

- [ ] **Step 2: Create `Dockerfile.qa` with separate final targets**

Use the existing locked bases as repository digests:

```dockerfile
FROM golang@sha256:bbb255b0e131db500cf0520adc97441d2260cf629c7fa7e39e025ddf53995a24 AS backend-unit
FROM mcr.microsoft.com/playwright@sha256:57b65fdc9ceabe0ef613124c7bbe2babcf9362c4d85e382fe3b03604e84b428a AS frontend-unit
FROM python@sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36 AS python-base
```

The Go stage copies `backend/`, resolves Go modules during the approved image
build, installs the fixed backend entrypoint, and runs as a non-root user. The
Node stage copies `frontend/package*.json`, runs `npm ci` during image build,
copies frontend source and the Vitest config, and installs the fixed frontend
entrypoint. The Python stages copy `pyproject.toml`/`uv.lock`, perform
`uv sync --frozen` into `/opt/hub/.venv`, copy package/tests/cases/policies, and
install `python-qa.sh`. No test runs during image build.

- [ ] **Step 3: Make the Go and Node entrypoints fixed and bounded**

Remove A1/bootstrap arguments and accept no user command. `backend-unit.sh`
runs exactly build, vet, unit, and race with `-tags hubtest`; the container
wall-clock timeout is enforced by Compose. `frontend-unit.sh` runs exactly
TypeScript, Biome, Vite build, and Vitest using the fixed config. Both use
`set -eu`, write reports below `/reports/<service>/`, preserve first non-zero
exit, and emit no secret-bearing environment dump.

- [ ] **Step 4: Add the closed Python dispatcher**

Implement a literal `case "$1" in` with only:

```sh
contract)    exec pytest selftests tests/contract --hub-case-root cases --strict-markers ;;
api)         exec pytest tests/api --hub-case-root cases --strict-markers ;;
integration) exec pytest tests/integration --hub-case-root cases --strict-markers ;;
e2e-smoke)   exec pytest tests/e2e -m smoke --hub-case-root cases --strict-markers ;;
e2e-full)    exec pytest tests/e2e --hub-case-root cases --strict-markers ;;
load)        exec locust -f load/locustfile.py --headless ;;
*)           exit 64 ;;
esac
```

Use absolute venv executables, a fixed `/opt/hub/testframework` working
directory, no `eval`, no command concatenation, and `PYTHONSAFEPATH=1`.

- [ ] **Step 5: Static review and commit**

Runtime remains `NOT_RUN_BOOTSTRAP`. Commit only the seven execution assets:

```text
git commit -m "build(testframework): consolidate Docker execution assets"
```

---

### Task 4: Make root Compose the only public test manifest

**Files:**
- Modify: `docker-compose.local.yml`
- Test: `testframework/selftests/architecture/docker/test_root_compose_contract.py`

**Interfaces:**
- Consumes: Dockerfile targets and entrypoints from Task 3.
- Produces: the exact eight public QA services and QA-only infrastructure
  services `qa-postgres`, `qa-mongo`, `qa-redis`, `qa-nats`, and
  `qa-seaweedfs`.

- [ ] **Step 1: Add failing Compose security assertions first**

For every public service require profile `qa`, `read_only: true`,
`security_opt: [no-new-privileges:true]`, `cap_drop: [ALL]`, `init: true`,
`pids_limit`, CPU/memory limits, a fixed command, and a `/reports` bind. Reject
published ports and `/var/run/docker.sock`.

- [ ] **Step 2: Add reusable Compose extensions**

Define `x-qa-common`, `x-qa-python`, and `x-qa-infrastructure`. Python services
share build target `python-tests` but have fixed commands `contract`, `api`,
`integration`, `e2e-smoke`, `e2e-full`, or `load`.

- [ ] **Step 3: Add the exact public services**

Use this shape without arbitrary command substitution:

```yaml
  backend-unit:
    profiles: [qa]
    build:
      context: .
      dockerfile: testframework/docker/Dockerfile.qa
      target: backend-unit
    command: []

  python-contract:
    profiles: [qa]
    build:
      context: .
      dockerfile: testframework/docker/Dockerfile.qa
      target: python-tests
    command: [contract]
```

Repeat with the exact target/command mapping for all eight services.

- [ ] **Step 4: Add isolated QA infrastructure**

Use the same pinned product versions as the development stack but distinct
service names, no host ports, tmpfs/disposable storage, health checks, limits,
and QA-only credentials. Only `python-integration`, `python-e2e-smoke`, and
`python-e2e-full` depend on healthy QA infrastructure. Unit/contract services
have no SUT infrastructure dependency.

- [ ] **Step 5: Source-review and commit root Compose**

The duplicate-aware contract must now accept all service/security predicates
except retired-path absence. Runtime remains `NOT_RUN_BOOTSTRAP`.

```text
git commit -m "build(testframework): expose root compose test services"
```

---

### Task 5: Resolve the Python lock inside Docker and execute the first RED

**Files:**
- Create: `testframework/uv.lock`
- Modify: `testframework/docker/Dockerfile.qa`
- Modify: `docker-compose.local.yml`

**Interfaces:**
- Consumes: `python-lock` Docker target and `pyproject.toml` with PyYAML.
- Produces: a committed cross-platform uv lock and buildable `python-tests`
  image.

- [ ] **Step 1: Add the internal maintenance service**

Add non-public profile `qa-maintenance` service `qa-lock`, build target
`python-lock`, fixed command `uv lock --project /workspace/testframework`, and
the single writable bind for `testframework/`. It has no Docker socket and no
test execution authority.

- [ ] **Step 2: Generate the lock in Docker**

Run with explicit approval for dependency network access:

```text
docker compose --profile qa-maintenance run --rm qa-lock
```

Expected: only `testframework/uv.lock` changes; package resolution succeeds.

- [ ] **Step 3: Build the Python test image through Compose**

```text
docker compose --profile qa build python-contract
```

Expected: `uv sync --frozen` succeeds; no test runs during build.

- [ ] **Step 4: Execute the root-Compose contract and confirm RED**

```text
docker compose run --rm python-contract
```

Expected: FAIL only because Task 6 retired paths still exist. Any parser,
import, lock, image, or Compose failure is triaged before proceeding.

- [ ] **Step 5: Commit the lock/bootstrap closure**

```text
git commit -m "build(testframework): lock Python QA image"
```

---

### Task 6: Delete the retired runner and duplicate Docker control plane

**Files:**
- Delete: `testframework/runner/contracts/hub.qa-bootstrap.v1.json`
- Delete: `testframework/runner/{invoke-qa-bootstrap-v1,invoke-qa-gate-v2,new-repository-snapshot,new-run-envelope,run-task,run-task-v2,write-g0-activation}.ps1`
- Delete: `testframework/runner/README.md`
- Delete: `testframework/selftests/runner/{fixtures,g0-bootstrap-attestor-contract,g0-image-contract,g0-runner-v2-contract,private-adapter,runner-contract}.ps1`
- Delete: `testframework/docker/admission/{admit_local_dependency_assets,verify_local_dependency_bundle}.py`
- Delete: `testframework/docker/{admit-local-dependency-assets,source-image-attestor}.ps1`
- Delete: `testframework/docker/{bootstrap-control.v1,build-manifest.v2,dependency-bundle.schema.v1,source-policy.v1}.json`
- Delete: `testframework/docker/{compose.g0-bootstrap,compose.unit}.yaml`
- Delete: `testframework/docker/Dockerfile.{backend-openapi,backend-unit,frontend-unit,g0-admission-helper,g0-planning-selftest}`
- Delete: `testframework/docker/entrypoints/{backend-openapi,g0-admission-helper,g0-planning-selftest}.sh`
- Delete: `testframework/docker/runner-fixtures/g0-planning-selftest.py`
- Delete: `.cursor/skills/qa-gate/SKILL.md`
- Delete: `deployments/docker-compose.test.yml`

**Interfaces:**
- Consumes: root Compose replacement and closed retirement policy.
- Produces: no old public launcher, bootstrap, activation, or nested Compose
  path.

- [ ] **Step 1: Verify every delete candidate is in the policy**

Compare the literal deletion list with
`root-compose-testframework.v1.json`. Any mismatch blocks deletion.

- [ ] **Step 2: Perform a retained-reference preflight**

Use `rg` to classify every active reference. Historical plans/specs remain
immutable history but must be excluded explicitly; active docs/CI are handled
in Task 7. No production or retained test source may import a deleted module.

- [ ] **Step 3: Delete only the exact paths with `apply_patch`**

Do not use recursive filesystem deletion, glob deletion, reset, checkout, or
clean. Preserve `src`, `tests`, `cases`, `policies`, `schemas`, `load`, and all
non-runner selftests.

- [ ] **Step 4: Run the Docker source contract to GREEN**

```text
docker compose run --rm python-contract
```

Expected: PASS for exact services, Docker asset set, security options, and
retired-path absence. Active documentation references may still fail only if
the test intentionally includes Task 7 paths; fix the contract boundary rather
than hiding a real active reference.

- [ ] **Step 5: Commit the deletion atomically**

```text
git commit -m "refactor(testframework): remove retired runner control plane"
```

---

### Task 7: Rewrite CI and active documentation to the Compose API

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `.gitlab-ci.yml`
- Modify: `AGENTS.md`
- Delete: `.cursor/skills/qa-gate/SKILL.md` if not already deleted in Task 6
- Modify: `README.md`
- Modify: `docs/{README,TESTING,TECHNICAL,TROUBLESHOOTING,code_style_guide,stack_and_tools,test_plan,testframework_strategy,user_needs_analysis_2026}.md`
- Modify: `docs/architecture.md`
- Modify: `docs/onboarding/{backend-developer,frontend-developer,pm-teamlead,qa-automation}.md`
- Modify: `docs/runbooks/deploy-and-rollback.md`
- Modify: `docs/wiki/{ARCHITECTURE,QUICKSTART}.md`
- Modify: `deployments/README.md`
- Modify: `testframework/{PACKAGE-CONTRACT,README,TESTING_GUIDE}.md`
- Modify: `testframework/docker/README.md`

**Interfaces:**
- Consumes: exact eight root Compose service names.
- Produces: no active `/qa-gate`, runner, nested Compose, or host test command.

- [ ] **Step 1: Write CI source assertions first**

Extend `test_root_compose_contract.py` to require the GitHub and GitLab test
jobs use only `docker compose run --rm <exact-service>` and to reject `go test`,
`vitest`, `pytest`, `npm test`, direct Docker run, and removed runner names in
active workflow command nodes.

- [ ] **Step 2: Replace quarantined CI jobs**

Map backend test to `backend-unit`, frontend test to `frontend-unit`, and
contract/API/integration jobs to their exact Compose services. Keep build jobs
dependent on successful test jobs. Do not add a host setup or fallback.

- [ ] **Step 3: Rewrite active docs consistently**

Every command example uses one of the eight exact Compose commands. Explain
that `docker compose up` remains the development stack, QA services are behind
profiles/explicit run, integration data is disposable and isolated, and the
PowerShell/G0 runner is retired. Remove readiness claims unsupported by Task 8.

- [ ] **Step 4: Scan active references**

Exclude `.planning/archive/**`, historical `docs/superpowers/plans/**`, and the
approved consolidation spec/plan. All other references to retired tokens are
failures.

- [ ] **Step 5: Commit CI/docs closure**

```text
git commit -m "docs(testframework): publish root compose test interface"
```

---

### Task 8: Docker-only verification and readiness truth

**Files:**
- Modify after evidence: `testframework/PACKAGE-CONTRACT.md`
- Modify after evidence: `testframework/README.md`
- Modify after evidence: `.planning/STATE.md`
- Create: `testframework/reports/<task-run>/` artifacts only through containers

**Interfaces:**
- Consumes: complete root Compose topology and locked images.
- Produces: actual per-service PASS/FAIL/NOT_READY evidence without inferred
  status.

- [ ] **Step 1: Validate Compose rendering inside Docker-capable environment**

```text
docker compose config --quiet
```

Expected: zero exit, no duplicate/unresolved service/config error.

- [ ] **Step 2: Run source and package contracts first**

```text
docker compose run --rm python-contract
```

Expected: architecture, config, traceability, contract, and root-Compose tests
execute. Triage every failure against product/docs before changing code/tests.

- [ ] **Step 3: Run genuine unit services**

```text
docker compose run --rm backend-unit
docker compose run --rm frontend-unit
```

Expected: build/vet/hubtest/race and TypeScript/Biome/Vite/Vitest execute in
their respective images. Record exact failing packages/tests; do not lower
coverage or delete genuine tests to obtain green.

- [ ] **Step 4: Run Python behavior stages in risk order**

```text
docker compose run --rm python-api
docker compose run --rm python-integration
docker compose run --rm python-e2e-smoke
docker compose run --rm python-e2e-full
docker compose run --rm python-load
```

Expected: each stage is recorded independently as PASS, FAIL, or NOT_READY.
Missing launcher delegates or unfinished cases remain NOT_READY rather than
being skipped or replaced with fakes.

- [ ] **Step 5: Verify cleanup and residual state**

Confirm reports contain no secrets, QA services publish no ports, development
volumes/databases are unchanged, and per-test ledgers leave no owned database,
bucket, stream, consumer, key, session, or principal residue.

- [ ] **Step 6: Update readiness documentation from evidence only**

Set each service to its observed status and include the exact Docker command,
commit, image digest, exit code, report path, and timestamp. Do not mark the
whole framework ready unless all required stages and cleanup checks pass.

- [ ] **Step 7: Commit verified status**

```text
git commit -m "test(testframework): verify root compose matrix"
```
