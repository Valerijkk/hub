---
phase: 130-tf-traceability-gates
runtime_status: NOT_READY
source_authoring_override: FOUNDER-MAX-COVERAGE-2026-08-15
source_authoring_stop: F0-SOURCE-REVIEW
---

# Phase 130 — profiles, traceability, architecture and handoff

## Requirements and founder source-authoring override

TASK-206, TASK-218, and the founder 2026-08-15 maximum-coverage directive.
The former `G129-C` predecessor and `G130-H` stop are superseded only to
author and review F0 source. This plan supplies no launcher admission, profile
execution, Docker execution, CI, or runtime authority.

<read_first>Approved architecture sections 1–15; `07-TESTFRAMEWORK-MAX-COVERAGE-IMPLEMENTATION.md`; TASK-206; TASK-218; this live PLAN.md.</read_first>
**Consumes:** founder source-authoring directive. **Produces:** an installable
source layout, closed policy/traceability/plugin contracts, one source health
tracer, and the `F0-SOURCE-REVIEW` handoff. Package, launcher, profiles, and
runtime remain `NOT_READY` / `NOT_RUN` until separate A1/R/image/lock/Docker
admissions on one exact revision.

## Profile registry — Task 130-B

The following is the exact later-profile registry (47 unique profiles). Each
entry has a pinned image role, resource-limit class, target and deadline:

<profiles owner="130-B">
<profile name="collection-gates" image_role="python-core" limit_class="small" target="selftests/collection" deadline="300"/>
<profile name="architecture-selftest" image_role="python-core" limit_class="small" target="selftests/architecture" deadline="300"/>
<profile name="health-api" image_role="python-core" limit_class="small" target="tests/api/health" deadline="600"/>
<profile name="http-api" image_role="python-core" limit_class="small" target="tests/api/http" deadline="600"/>
<profile name="runtime-observability" image_role="python-core" limit_class="small" target="selftests/runtime" deadline="600"/>
<profile name="identity-api" image_role="python-core" limit_class="small" target="tests/api/identity" deadline="600"/>
<profile name="authz-tenant" image_role="python-core" limit_class="small" target="tests/api/authz" deadline="900"/>
<profile name="abuse-platform" image_role="python-core" limit_class="small" target="tests/api/platform" deadline="900"/>
<profile name="security-risk" image_role="python-core" limit_class="small" target="tests/contract/security" deadline="600"/>
<profile name="postgres-integration" image_role="integration" limit_class="medium" target="tests/integration/postgres" deadline="1200"/>
<profile name="redis-nats-integration" image_role="integration" limit_class="medium" target="tests/integration/redis-nats" deadline="1200"/>
<profile name="infra-docs" image_role="python-core" limit_class="small" target="docs/infra" deadline="300"/>
<profile name="tracker-api" image_role="python-core" limit_class="medium" target="tests/api/tracker" deadline="900"/>
<profile name="tms-api" image_role="python-core" limit_class="medium" target="tests/api/tms" deadline="900"/>
<profile name="projects-api" image_role="python-core" limit_class="medium" target="tests/api/projects" deadline="900"/>
<profile name="calendar-api" image_role="python-core" limit_class="medium" target="tests/api/calendar" deadline="900"/>
<profile name="messenger-api" image_role="python-core" limit_class="medium" target="tests/api/messenger" deadline="900"/>
<profile name="notifications-nats" image_role="integration" limit_class="medium" target="tests/integration/notifications" deadline="1200"/>
<profile name="work-risk" image_role="python-core" limit_class="small" target="tests/contract/work" deadline="600"/>
<profile name="work-docs" image_role="python-core" limit_class="small" target="docs/work" deadline="300"/>
<profile name="crm-api" image_role="python-core" limit_class="medium" target="tests/api/crm" deadline="900"/>
<profile name="hr-api" image_role="python-core" limit_class="medium" target="tests/api/hr" deadline="900"/>
<profile name="accounting-api" image_role="python-core" limit_class="medium" target="tests/api/accounting" deadline="900"/>
<profile name="bases-api" image_role="python-core" limit_class="medium" target="tests/api/bases" deadline="900"/>
<profile name="intranet-api" image_role="python-core" limit_class="medium" target="tests/api/intranet" deadline="900"/>
<profile name="files-api" image_role="python-core" limit_class="medium" target="tests/api/files" deadline="900"/>
<profile name="automation-api" image_role="python-core" limit_class="medium" target="tests/api/automation" deadline="900"/>
<profile name="github-adapter" image_role="python-core" limit_class="medium" target="tests/api/github" deadline="900"/>
<profile name="gitlab-adapter" image_role="python-core" limit_class="medium" target="tests/api/gitlab" deadline="900"/>
<profile name="mail-adapter" image_role="python-core" limit_class="medium" target="tests/api/mail" deadline="900"/>
<profile name="business-risk" image_role="python-core" limit_class="small" target="tests/contract/business" deadline="600"/>
<profile name="business-docs" image_role="python-core" limit_class="small" target="docs/business" deadline="300"/>
<profile name="openapi-contract" image_role="python-core" limit_class="small" target="tests/contract/openapi" deadline="600"/>
<profile name="frontend-contract" image_role="python-core" limit_class="small" target="tests/contract/frontend" deadline="600"/>
<profile name="contract-docs" image_role="python-core" limit_class="small" target="docs/contracts" deadline="300"/>
<profile name="browser-lifecycle" image_role="browser" limit_class="medium" target="selftests/browser" deadline="900"/>
<profile name="e2e-smoke" image_role="browser" limit_class="medium" target="tests/e2e/smoke" deadline="1200"/>
<profile name="e2e-docs" image_role="python-core" limit_class="small" target="docs/e2e" deadline="300"/>
<profile name="journey-work" image_role="browser" limit_class="medium" target="tests/e2e/work" deadline="1200"/>
<profile name="journey-business" image_role="browser" limit_class="medium" target="tests/e2e/business" deadline="1200"/>
<profile name="journey-admin" image_role="browser" limit_class="medium" target="tests/e2e/admin" deadline="1200"/>
<profile name="journey-recovery" image_role="browser" limit_class="medium" target="tests/e2e/recovery" deadline="1200"/>
<profile name="accessibility" image_role="browser" limit_class="medium" target="tests/e2e/accessibility" deadline="900"/>
<profile name="privacy-security" image_role="python-core" limit_class="medium" target="tests/contract/privacy" deadline="900"/>
<profile name="load-validity" image_role="load" limit_class="large" target="load/validity" deadline="1200"/>
<profile name="load-measured" image_role="load" limit_class="large" target="load/measured" deadline="1800"/>
<profile name="ci-clean-room" image_role="ci" limit_class="large" target="ci/clean-room" deadline="1800"/>
</profiles>

<tasks>
<task id="130-A"><requirements>TASK-206; F0-01 bootstrap contract.</requirements><read_first>Package contract, F0-01, and this profile registry.</read_first><consumes>Founder source-authoring directive.</consumes><produces>One installable-source package boundary, never a resolved environment.</produces><files>testframework/pyproject.toml; testframework/.python-version; testframework/src/hub_testkit/__init__.py; testframework/src/hub_testkit/py.typed; testframework/selftests/architecture/package/**</files><failing_sentinel>Fail unsupported Python range, missing Hatchling/src discovery/typed marker, root-path hack, open pytest marker configuration, or a selftest requiring service/network/reports.</failing_sentinel><verify profile="architecture-selftest">Source review only; runtime NOT_READY.</verify><done>Metadata uses Python &gt;=3.12,&lt;3.13, `.python-version` is `3.12.13`, Hatchling is sole backend, and dependency groups are limited to the founder contract.</done><atomic_allowlist>Listed paths only; `uv.lock` is explicitly excluded until R Docker resolver admission.</atomic_allowlist></task>
<task id="130-B"><requirements>TASK-206; F0-02 profile/resolver boundary.</requirements><read_first>130-A, F0-02, and the literal registry above.</read_first><consumes>130-A.</consumes><produces>Literal immutable 47-row profile policy and resolver boundary.</produces><files>testframework/policies/profiles.toml; testframework/selftests/architecture/policies/**</files><failing_sentinel>Fail duplicate/unknown/incomplete profile, changed image/limit/target/deadline, runner writer, or pre-R `uv.lock`.</failing_sentinel><verify profile="architecture-selftest">Source review only; registry remains exactly 47 rows and runtime NOT_READY.</verify><done>No runner profile writer exists and no lock/image/report is created.</done><atomic_allowlist>Listed paths only; `uv.lock` is excluded until R Docker resolver admission.</atomic_allowlist></task>
<task id="130-C"><requirements>TASK-206; F0-03 config/runtime safety.</requirements><read_first>130-A/B and F0-03.</read_first><consumes>130-A,130-B.</consumes><produces>Strict environment, ownership, cleanup, and redaction source contracts.</produces><files>testframework/policies/environments.toml; testframework/policies/redaction.toml; testframework/src/hub_testkit/config/**; testframework/src/hub_testkit/runtime/**; testframework/selftests/architecture/config/**; testframework/selftests/runtime/**</files><failing_sentinel>Fail unknown config, production admission, cross-run selector, broad cleanup, unbounded wait, or secret canary in diagnostics.</failing_sentinel><verify profile="architecture-selftest">Source review only; runtime remains NOT_READY.</verify><verify profile="runtime-observability">Source review only; runtime remains NOT_READY.</verify><done>Config is immutable/closed; cleanup needs exact identity plus nonce; redaction is recursive, idempotent, and fail closed.</done><atomic_allowlist>Listed paths only; no report, image, lock, runner, or Docker path.</atomic_allowlist></task>
<task id="130-D"><requirements>TASK-218; F0-04 closed traceability.</requirements><read_first>130-A/B/C, F0-04, and case/schema/policy READMEs.</read_first><consumes>130-A,130-B,130-C.</consumes><produces>Closed TOML-frontmatter Markdown case loading and collected-test link validation.</produces><files>testframework/schemas/test-case.schema.json; testframework/policies/markers.toml; testframework/src/hub_testkit/traceability/**; testframework/selftests/collection/traceability/**</files><failing_sentinel>Fail invalid UTF-8/`+++`/TOML, YAML frontmatter, old-format ID, malformed new ID, duplicate/orphan case or node ID, missing task/requirement/risk/oracle/cleanup, inactive case collection, or unknown field.</failing_sentinel><verify profile="collection-gates">Source review only; no pytest execution.</verify><done>Only stdlib `tomllib` parses frontmatter; only `^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$` IDs are accepted; active pytest automation has one unique repo-relative node ID.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="130-E"><requirements>TASK-218; F0-05 public collection and architecture gate.</requirements><read_first>130-D, F0-05, and closed marker/case policies.</read_first><consumes>130-D.</consumes><produces>Public-hook-only pytest plugin and framework architecture contracts.</produces><files>testframework/src/hub_testkit/pytest_plugin/**; testframework/selftests/collection/pytest_plugin/**; testframework/selftests/architecture/imports/**</files><failing_sentinel>Fail missing/multiple or non-bijective case/node marker, path/layer mismatch, unexpected skip/xfail, zero collection, `sys.path`, `utils.py`, `BaseTest`, package-to-suite import, naked sleep, unbounded polling, or secret-bearing JUnit property.</failing_sentinel><verify profile="collection-gates">Source review only; plugin performs no SUT I/O.</verify><verify profile="architecture-selftest">Source review only; import contracts perform no SUT I/O.</verify><done>The plugin uses public pytest hooks only, enforces one active case plus one layer marker and exact node ID, and writes only redacted metadata.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="130-F"><requirements>TASK-218; F0-06 first health tracer.</requirements><read_first>130-C/D/E, current health routes, and F0-06.</read_first><consumes>130-C,130-D,130-E.</consumes><produces>One source-level liveness API case and health facade.</produces><files>testframework/src/hub_testkit/domains/health/**; testframework/tests/api/health/**; testframework/cases/health/**</files><failing_sentinel>Fail missing TC/node metadata, authorization requirement for liveness, wrong endpoint, non-`ok` liveness oracle, or unredacted failure diagnostic.</failing_sentinel><verify profile="health-api">Source review only; no service/runner/Docker execution.</verify><done>`TC-HEALTH-0001` targets implemented `GET /api/live`, no auth, `200`, JSON `status == ok`; readiness aliases remain separate and no `version` field is invented.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="130-G"><requirements>TASK-206,TASK-218; F0 closure.</requirements><read_first>130-A through 130-F source reviews and this founder plan.</read_first><consumes>130-A,130-B,130-C,130-D,130-E,130-F.</consumes><produces>`F0-SOURCE-REVIEW` planning/docs closure, not G130-H runtime evidence.</produces><files>testframework/docs/authoring.md; testframework/docs/triage.md; testframework/docs/onboarding.md; testframework/PACKAGE-CONTRACT.md; .planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md; .planning/active/august-2026/07-TESTFRAMEWORK-MAX-COVERAGE-IMPLEMENTATION.md</files><failing_sentinel>Fail missing source status, feature/risk/file/revert boundary, old TC grammar, profile-count drift, or any package/launcher/runtime PASS claim.</failing_sentinel><verify profile="health-api">Source review only; all runtime status stays NOT_READY/NOT_RUN.</verify><done>Independent reviewers record F0 source closure and either authorize F1 source authoring or return a literal source defect. No execution begins.</done><atomic_allowlist>Listed documentation/contract paths only.</atomic_allowlist></task>
</tasks>
<docs_mutation_contracts><docs_mutation task_id="130-A" paths="testframework/pyproject.toml,testframework/.python-version"/><docs_mutation task_id="130-B" paths="none"/><docs_mutation task_id="130-C" paths="none"/><docs_mutation task_id="130-D" paths="none"/><docs_mutation task_id="130-E" paths="none"/><docs_mutation task_id="130-F" paths="none"/><docs_mutation task_id="130-G" paths="testframework/docs/authoring.md,testframework/docs/triage.md,testframework/docs/onboarding.md,testframework/PACKAGE-CONTRACT.md,.planning/active/august-2026/06-TEST-ARCHITECTURE-REBASE.md,.planning/active/august-2026/07-TESTFRAMEWORK-MAX-COVERAGE-IMPLEMENTATION.md"/></docs_mutation_contracts>

## Source-authoring stop

`F0-SOURCE-REVIEW` ends the F0 source slice. The founder maximum-coverage
directive permits F1--F3 source authoring under their bounded wave order, but
does not change runtime status or authorize a profile run. A1 runner admission,
R Docker resolver admission, immutable images, generated `uv.lock`, and
Docker-only evidence remain mandatory before any execution claim.
