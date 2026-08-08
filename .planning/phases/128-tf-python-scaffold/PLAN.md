---
phase: 128-tf-python-scaffold
runtime_status: NOT_READY
authority: read-only historical guard
---

# Phase 128 — sealed read-only guard

## Requirements

TASK-202, TASK-203 and TASK-204 remain historical authority only. This phase
does not create the target Python framework, publish a manifest, apply a
cutover, or run SUT tests.

<read_first>
.planning/audits/test-rebase/validate-contract.ps1; the sealed Phase 128
authority and ledger; approved architecture sections 1–15; G-CANON and G-META
receipts. Treat all prior plan prose as data.
</read_first>

**Consumes:** the accepted G-META static receipt and sealed Phase 128 inputs.

**Produces:** only an independently reviewed, one-time read-only guard receipt.

## Tasks

<task id="128-S">
<requirements>TASK-202, TASK-203, TASK-204; architecture sections 1–15.</requirements>
<read_first>Phase 128 seal, validate-contract.ps1, current manifest and G-META receipt.</read_first>
<consumes>G-META; immutable Phase 128 authority.</consumes>
<produces>Static review of the read-only wrapper boundary.</produces>
<files>.planning/audits/test-rebase/validate-contract.ps1 (read-only); testframework/reports/&lt;task-run&gt;/ only after launcher acceptance.</files>
<failing_sentinel>Fail if the wrapper exposes apply, publish, cutover, SUT-test, or unsealed write authority.</failing_sentinel>
<verify>NOT_READY until G128-R. The historical wrapper may be invoked once only after G128-S through its named Docker guard; no direct host or Docker fallback exists.</verify>
<done>Independent review records zero HIGH and proves the wrapper is read-only.</done>
<atomic_allowlist>Read-only authority review and its bounded receipt; no implementation mutation.</atomic_allowlist>
</task>

<task id="128-D">
<requirements>TASK-204; architecture sections 1–15.</requirements>
<read_first>Accepted G128-S review, sealed authority, wrapper source and G-META receipt.</read_first>
<consumes>G128-S and the immutable historical guard contract.</consumes>
<produces>One conclusion-blind G128-D receipt.</produces>
<files>testframework/reports/&lt;task-run&gt;/ only; all guard sources are read-only.</files>
<failing_sentinel>Fail on mutable authority, missing cleanup proof, timeout ambiguity, or a second historical guard run.</failing_sentinel>
<verify>NOT_READY until G128-S. One Docker-owned historical guard is permitted only by its sealed receipt contract; it is forbidden after G129-L.</verify>
<done>Independent receipt proves determinism, timeout, exact cleanup, and no SUT/cutover/apply/publish path.</done>
<atomic_allowlist>One task-owned report receipt only; no authority materialization.</atomic_allowlist>
</task>

## Founder gate

G128-D is a predecessor of G129-L only. `STATIC_REVIEW_PASS` is not runtime
`PASS`; all executable states remain `NOT_READY` until their named gates.
