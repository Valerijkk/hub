---
phase: 136-tf-python-contracts
runtime_status: NOT_READY
activation_precondition: G-ACTIVATE-136
---
# Phase 136 — on-demand contracts
Every task is **NOT_READY** until `G-ACTIVATE-136` names one feature, risk, files and rollback boundary.
<read_first>G130-H; architecture sections 3,4,7,9–12; TASK-216, TASK-230; G-ACTIVATE-136.</read_first>
<tasks>
<task id="136-A"><requirements>TASK-216; sections 3,4,7.</requirements><read_first>G-ACTIVATE-136 and current OpenAPI/schema/event boundary.</read_first><consumes>G130-H;G-ACTIVATE-136.</consumes><produces>API/schema/event contract cases.</produces><files>testframework/src/hub_testkit/contract/**;testframework/tests/contract/openapi/**;testframework/cases/contracts/**;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail incompatible schema/event or unbounded snapshot.</failing_sentinel><verify profile="openapi-contract">NOT_READY until G-ACTIVATE-136.</verify><done>Activated contract boundary is verified.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="136-B"><requirements>TASK-230; sections 3,4,7.</requirements><read_first>G-ACTIVATE-136 and frontend contract boundary.</read_first><consumes>G130-H;G-ACTIVATE-136.</consumes><produces>Frontend BFF contract cases.</produces><files>testframework/tests/contract/frontend/**;testframework/cases/frontend-contract/**;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail incompatible public BFF schema or redaction boundary.</failing_sentinel><verify profile="frontend-contract">NOT_READY until G-ACTIVATE-136.</verify><done>Activated frontend contract is verified.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="136-C"><requirements>TASK-216,TASK-230; sections 10–15.</requirements><read_first>136-A/B receipts and activated boundary.</read_first><consumes>136-A,136-B;G-ACTIVATE-136.</consumes><produces>Single contract docs closure.</produces><files>testframework/docs/contracts.md;testframework/cases/contracts/**;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail closure without feature/risk/file/rollback.</failing_sentinel><verify profile="contract-docs">NOT_READY until 136-A/B and G-ACTIVATE-136.</verify><done>No automatic activation occurs.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
</tasks>
