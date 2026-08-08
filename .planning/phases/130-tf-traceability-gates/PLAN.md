---
phase: 130-tf-traceability-gates
runtime_status: NOT_READY
stop_gate: G130-H
---

# Phase 130 — profiles, traceability, architecture and handoff

## Requirements

TASK-206 and TASK-218; approved architecture sections 1–15. Phase 130 is the
only current-scope implementation continuation after G129-C and stops at
G130-H. It creates no Phase 131–141 execution authority.

<read_first>Approved architecture sections 1–15; G129-C receipt; TASK-206; TASK-218; this live PLAN.md.</read_first>
**Consumes:** G129-C and accepted launcher v1. **Produces:** one bounded profile
registry, framework selftests, a health TC and the G130-H handoff receipt.

## Profile registry — Task 130-A

The following is the exact later-profile registry (47 unique profiles). Each
entry has a pinned image role, resource-limit class, target and deadline:

<profiles owner="130-A">
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
<task id="130-A"><requirements>TASK-206; architecture sections 3, 6, 9.</requirements><read_first>Profile registry above, launcher lock, policies and G129-C receipt.</read_first><consumes>G129-C.</consumes><produces>Read-only 47-row registry in testframework/policies/profiles.toml.</produces><files>testframework/policies/profiles.toml; testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail duplicate, unknown, incomplete image/limit/target/deadline profile.</failing_sentinel><verify profile="collection-gates">NOT_READY until G129-C; accepted launcher only.</verify><done>Registry has exactly 47 rows and no runner profile writer exists after this task.</done><atomic_allowlist>testframework/policies/profiles.toml; task-owned report.</atomic_allowlist></task>
<task id="130-B"><requirements>TASK-218; architecture sections 3, 4, 9.</requirements><read_first>130-A profile registry and traceability schema.</read_first><consumes>130-A.</consumes><produces>Public-hook-only pytest plugin and traceability gate.</produces><files>testframework/src/hub_testkit/pytest_plugin/**; testframework/src/hub_testkit/traceability/**; testframework/selftests/traceability/**; testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail orphan TC, unknown marker, missing risk/oracle or unexpected skip.</failing_sentinel><verify profile="collection-gates">NOT_READY until 130-A and G129-C.</verify><done>Negative fixtures fail closed and restore without shared artifacts.</done><atomic_allowlist>Listed task files only.</atomic_allowlist></task>
<task id="130-C"><requirements>TASK-218; architecture sections 3–6.</requirements><read_first>130-A registry, 130-B plugin and architecture rules.</read_first><consumes>130-A, 130-B.</consumes><produces>Framework architecture selftests and one health-domain API TC.</produces><files>testframework/src/hub_testkit/config/**; testframework/src/hub_testkit/runtime/**; testframework/src/hub_testkit/domains/health/**; testframework/tests/api/health/**; testframework/cases/health/**; testframework/selftests/architecture/**; testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail forbidden imports, nonzero collection loss, unredacted diagnostics or missing TC metadata.</failing_sentinel><verify profile="architecture-selftest">NOT_READY until 130-B and G129-C.</verify><done>One traceable health TC and selftests provide the current-scope evidence.</done><atomic_allowlist>Listed task files only.</atomic_allowlist></task>
<task id="130-D"><requirements>TASK-206, TASK-218; architecture sections 10–15.</requirements><read_first>130-A through 130-C receipts and current docs closure.</read_first><consumes>130-A, 130-B, 130-C.</consumes><produces>G130-H handoff and exact documentation/traceability closure.</produces><files>testframework/docs/authoring.md; testframework/docs/triage.md; testframework/docs/onboarding.md; testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail missing feature, risk, file or rollback boundary in handoff evidence.</failing_sentinel><verify profile="health-api">NOT_READY until 130-C; accepted launcher only.</verify><done>G130-H independently records framework selftests, one health TC, docs closure and STOP.</done><atomic_allowlist>Listed documentation and task-owned report only.</atomic_allowlist></task>
</tasks>
<docs_mutation_contracts><docs_mutation task_id="130-A" paths="none"/><docs_mutation task_id="130-B" paths="none"/><docs_mutation task_id="130-C" paths="none"/><docs_mutation task_id="130-D" paths="testframework/docs/authoring.md,testframework/docs/triage.md,testframework/docs/onboarding.md"/></docs_mutation_contracts>

## Stop

G130-H ends the current implementation run. Registration of a later profile is
not activation. Phases 131–141 require a separate exact G-ACTIVATE-&lt;phase&gt;
receipt that names one feature, risk, files and rollback boundary.
