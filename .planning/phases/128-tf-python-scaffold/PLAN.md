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

**Produces:** first a committed, independently reviewed historical-guard source
envelope; then one static receipt and one independently reviewed, one-time
read-only guard receipt.

## Tasks

<task id="128-E">
<requirements>TASK-202, TASK-203, TASK-204; architecture sections 1,2,5,6,12,14.</requirements><owner>tech-lead</owner><verifier>arch-ver; security-auditor</verifier>
<read_first>Accepted G128-R receipts; exact Git tracking state; wrapper plus every executable/read authority it consumes; agent safety rule.</read_first>
<consumes>G128-R commit fd3a5671 and immutable architecture, ledger, founder and Phase 129 identities.</consumes>
<produces>One committed full-HEAD source envelope for the historical guard; no runtime receipt or authority transition.</produces>
<files>.planning/audits/test-rebase/validate-contract.ps1; .planning/audits/test-rebase/apply_cutover_batch.py; .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json; .planning/audits/test-rebase/inventory.py; .planning/audits/test-rebase/inventory.json; .planning/audits/test-rebase/inventory.schema.json; .planning/audits/test-rebase/test-architecture-contract.json; .planning/audits/test-rebase/validate_test_architecture_contract.py; .planning/audits/test-rebase/build_cutover_manifest.py; .planning/audits/test-rebase/validate_cutover_manifest.py; .planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md</files>
<failing_sentinel>Fail while any of the first five writable envelope paths is dirty/untracked, any of the eleven paths lacks a same-HEAD blob, or an immutable identity differs from 1049 / 507-245-297+28 / 173-1120-max56.</failing_sentinel>
<verify><automated>Static read-only source and Git-tree review only; no Python, validator, builder, Docker, SUT, publish, apply, rollback or report. Arch and security reviewers must accept the exact five-path writable diff plus six read-only same-HEAD dependencies.</automated></verify>
<done>All eleven envelope paths are bound to one committed HEAD; the five writable paths are clean; architecture/security reviews have zero HIGH; runtime remains NOT_RUN.</done>
<atomic_allowlist>Commit only validate-contract.ps1, apply_cutover_batch.py, founder-greenfield-cutover-manifest.json, inventory.py and inventory.json; exclude VERIFICATION, baseline/debug, reports, cutover-manifest, global planning/product dirt and pro-qa-ai.</atomic_allowlist>
</task>

<task id="128-S">
<requirements>TASK-202, TASK-203, TASK-204; architecture sections 1–15.</requirements><owner>tech-lead</owner><verifier>gsd-plan-checker</verifier>
<read_first>Accepted G128-E commit/tree and reviews; Phase 128 seal, validate-contract.ps1, current manifest and G-META receipt.</read_first>
<consumes>G128-E committed envelope; G-META; immutable Phase 128 authority.</consumes>
<produces>Static review of the read-only wrapper boundary.</produces>
<files>.planning/audits/test-rebase/validate-contract.ps1; .planning/audits/test-rebase/apply_cutover_batch.py; .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json; .planning/audits/test-rebase/inventory.py; .planning/audits/test-rebase/inventory.json; .planning/audits/test-rebase/inventory.schema.json; .planning/audits/test-rebase/test-architecture-contract.json; .planning/audits/test-rebase/validate_test_architecture_contract.py; .planning/audits/test-rebase/build_cutover_manifest.py; .planning/audits/test-rebase/validate_cutover_manifest.py; .planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md; .planning/audits/test-rebase/VERIFICATION.md</files>
<failing_sentinel>Fail if the wrapper exposes apply, publish, cutover, SUT-test, or unsealed write authority.</failing_sentinel>
<verify><automated>NOT_READY until G128-E; the historical wrapper may be invoked once only after G128-S through its named Docker guard, with no direct host or Docker fallback.</automated></verify>
<done>Independent review records zero HIGH and proves the wrapper is read-only.</done>
<atomic_allowlist>Read-only authority review and its bounded receipt; no implementation mutation.</atomic_allowlist>
</task>

<task id="128-D">
<requirements>TASK-204; architecture sections 1–15.</requirements><owner>qa-auto-ver</owner><verifier>gsd-plan-checker</verifier>
<read_first>Accepted G128-S review, sealed authority, wrapper source and G-META receipt.</read_first>
<consumes>G128-S and the immutable historical guard contract.</consumes>
<produces>One conclusion-blind G128-D receipt.</produces>
<files>testframework/reports/&lt;task-run&gt;/ only; all guard sources are read-only.</files>
<failing_sentinel>Fail on mutable authority, missing cleanup proof, timeout ambiguity, or a second historical guard run.</failing_sentinel>
<verify><automated>NOT_READY until G128-S; one Docker-owned historical guard is permitted only by its sealed receipt contract and is forbidden after G129-L.</automated></verify>
<done>Independent receipt proves determinism, timeout, exact cleanup, and no SUT/cutover/apply/publish path.</done>
<atomic_allowlist>One task-owned report receipt only; no authority materialization.</atomic_allowlist>
</task>

## Founder gate

G128-E precedes G128-S; G128-D is a predecessor of G129-L only.
`STATIC_REVIEW_PASS` is not runtime
`PASS`; all executable states remain `NOT_READY` until their named gates.
