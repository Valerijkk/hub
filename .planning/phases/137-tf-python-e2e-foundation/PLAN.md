---
phase: 137-tf-python-e2e-foundation
runtime_status: NOT_READY
activation_precondition: G-ACTIVATE-137
---
# Phase 137 — on-demand browser foundation
Every task is **NOT_READY** until `G-ACTIVATE-137` names one feature, risk, files and rollback boundary.
<read_first>G130-H; sections 3,4,6,9–12; TASK-194, TASK-217; G-ACTIVATE-137.</read_first>
<tasks>
<task id="137-A"><requirements>TASK-194; sections 3,4,6.</requirements><read_first>G-ACTIVATE-137 and browser lifecycle boundary.</read_first><consumes>G130-H;G-ACTIVATE-137.</consumes><produces>Browser lifecycle and sanitized artifacts.</produces><files>testframework/src/hub_testkit/browser/**;testframework/selftests/browser/**;testframework/docker/Dockerfile.browser;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail browser leak, unredacted artifact or mutable image.</failing_sentinel><verify profile="browser-lifecycle">NOT_READY until G-ACTIVATE-137.</verify><done>Activated browser lifecycle is independently verified.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="137-B"><requirements>TASK-217; sections 3,4,9.</requirements><read_first>137-A receipt and activated smoke journey.</read_first><consumes>137-A;G-ACTIVATE-137.</consumes><produces>Critical smoke journey cases.</produces><files>testframework/tests/e2e/smoke/**;testframework/cases/e2e/**;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail login/journey oracle, artifact privacy or cleanup.</failing_sentinel><verify profile="e2e-smoke">NOT_READY until 137-A and G-ACTIVATE-137.</verify><done>One activated smoke journey is verified.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="137-C"><requirements>TASK-194,TASK-217; sections 10–15.</requirements><read_first>137-A/B receipts and activated boundary.</read_first><consumes>137-A,137-B;G-ACTIVATE-137.</consumes><produces>Single E2E docs closure.</produces><files>testframework/docs/e2e.md;testframework/cases/e2e/**;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail closure without feature/risk/file/rollback.</failing_sentinel><verify profile="e2e-docs">NOT_READY until 137-B and G-ACTIVATE-137.</verify><done>No journey range is authorized.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
</tasks>
