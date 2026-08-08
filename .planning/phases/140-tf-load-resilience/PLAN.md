---
phase: 140-tf-load-resilience
runtime_status: NOT_READY
activation_precondition: G-ACTIVATE-140
---
# Phase 140 — on-demand measured load and resilience
Every task is **NOT_READY** until `G-ACTIVATE-140` names one feature, risk, files and rollback boundary.
<read_first>G130-H; sections 6,8–12; TASK-200; G-ACTIVATE-140.</read_first>
<tasks>
<task id="140-A"><requirements>TASK-200; sections 6,8,9.</requirements><read_first>G-ACTIVATE-140 and load safety boundary.</read_first><consumes>131-D;G-ACTIVATE-140.</consumes><produces>Load validity guard.</produces><files>testframework/load/validity/**;testframework/policies/load.toml;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail production target, missing cap or unsafe cleanup.</failing_sentinel><verify profile="load-validity">NOT_READY until G-ACTIVATE-140.</verify><done>Activated load safety validated.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="140-B"><requirements>TASK-200; sections 6,8,9.</requirements><read_first>140-A receipt and measured profile boundary.</read_first><consumes>140-A;G-ACTIVATE-140.</consumes><produces>Measured Locust profiles.</produces><files>testframework/load/profiles/**;testframework/load/scenarios/**;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail unbounded rate, baseline ambiguity or data leak.</failing_sentinel><verify profile="load-measured">NOT_READY until 140-A and G-ACTIVATE-140.</verify><done>Activated measured load profile verified.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
<task id="140-C"><requirements>TASK-200; sections 10–15.</requirements><read_first>140-A/B receipts and activated boundary.</read_first><consumes>140-A,140-B;G-ACTIVATE-140.</consumes><produces>Single load docs closure.</produces><files>testframework/docs/load.md;testframework/cases/load/**;testframework/reports/&lt;task-run&gt;/</files><failing_sentinel>Fail closure without feature/risk/file/rollback.</failing_sentinel><verify profile="load-measured">NOT_READY until 140-B and G-ACTIVATE-140.</verify><done>No automatic production claim.</done><atomic_allowlist>Listed paths only.</atomic_allowlist></task>
</tasks>
