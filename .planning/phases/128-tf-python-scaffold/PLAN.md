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
envelope and static receipt; then a separately approved and sealed host
PowerShell TCB, a renewed envelope/static receipt after wrapper hardening, and
one independently reviewed, one-time read-only guard receipt.

## Tasks

<task id="128-E">
<requirements>TASK-202, TASK-203, TASK-204; architecture sections 1,2,5,6,12,14.</requirements><owner>tech-lead</owner><verifier>arch-ver; security-auditor</verifier>
<read_first>Accepted G128-R receipts; exact Git tracking state; wrapper plus every executable/read authority it consumes; agent safety rule; for renewal only, accepted G128-H receipt and powershell-host-lock.json.</read_first>
<consumes>Initially G128-R commit fd3a5671 and immutable architecture, ledger, founder and Phase 129 identities; on the mandatory renewal, accepted G128-H and its committed host lock.</consumes>
<produces>Initially one committed full-HEAD source envelope; after G128-H, one renewed envelope commit/review binding the hardened wrapper and host lock. Neither pass produces a runtime receipt or authority transition.</produces>
<files>.planning/audits/test-rebase/validate-contract.ps1; .planning/audits/test-rebase/apply_cutover_batch.py; .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json; .planning/audits/test-rebase/inventory.py; .planning/audits/test-rebase/inventory.json; .planning/audits/test-rebase/inventory.schema.json; .planning/audits/test-rebase/test-architecture-contract.json; .planning/audits/test-rebase/validate_test_architecture_contract.py; .planning/audits/test-rebase/build_cutover_manifest.py; .planning/audits/test-rebase/validate_cutover_manifest.py; .planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md; .planning/audits/test-rebase/powershell-host-lock.json</files>
<failing_sentinel>Fail while any of the first five writable envelope paths is dirty/untracked, any of the eleven paths lacks a same-HEAD blob, or an immutable identity differs from 1049 / 507-245-297+28 / 173-1120-max56.</failing_sentinel>
<verify><automated>Static read-only source and Git-tree review only; no Python, validator, builder, Docker, SUT, publish, apply, rollback or report. Initial review accepts the exact five-path writable diff plus six read-only same-HEAD dependencies. Mandatory post-G128-H renewal accepts the separately committed host lock and commits/reviews only the hardened validate-contract.ps1; it proves fail-closed equality for current interpreter absolute path, SHA256, file/product version and Authenticode identity.</automated></verify>
<done>The initial eleven-path envelope is accepted; after G128-H, the twelve-path envelope is rebound to the new committed wrapper and host lock, architecture/security reviews have zero HIGH, and runtime remains NOT_RUN.</done>
<atomic_allowlist>Initial commit: only validate-contract.ps1, apply_cutover_batch.py, founder-greenfield-cutover-manifest.json, inventory.py and inventory.json. Renewal commit: only validate-contract.ps1. Exclude VERIFICATION, baseline/debug, reports, cutover-manifest, global planning/product dirt and pro-qa-ai.</atomic_allowlist>
</task>

<task id="128-S">
<requirements>TASK-202, TASK-203, TASK-204; architecture sections 1–15.</requirements><owner>tech-lead</owner><verifier>gsd-plan-checker</verifier>
<read_first>Accepted G128-E commit/tree and reviews; Phase 128 seal, validate-contract.ps1, current manifest and G-META receipt; for renewal only, accepted G128-H host lock and renewed G128-E.</read_first>
<consumes>Initial or renewed G128-E committed envelope; G-META; immutable Phase 128 authority. The initial receipt can authorize only G128-H; G128-D consumes only the renewed receipt.</consumes>
<produces>Initial static review of the read-only wrapper boundary, then a mandatory renewed review after wrapper hardening that seals the literal absolute pwsh path and exact G128-D command.</produces>
<files>.planning/audits/test-rebase/validate-contract.ps1; .planning/audits/test-rebase/apply_cutover_batch.py; .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json; .planning/audits/test-rebase/inventory.py; .planning/audits/test-rebase/inventory.json; .planning/audits/test-rebase/inventory.schema.json; .planning/audits/test-rebase/test-architecture-contract.json; .planning/audits/test-rebase/validate_test_architecture_contract.py; .planning/audits/test-rebase/build_cutover_manifest.py; .planning/audits/test-rebase/validate_cutover_manifest.py; .planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md; .planning/audits/test-rebase/powershell-host-lock.json; .planning/audits/test-rebase/VERIFICATION.md</files>
<failing_sentinel>Fail if the wrapper exposes apply, publish, cutover, SUT-test, or unsealed write authority.</failing_sentinel>
<verify><automated>NOT_READY until the applicable G128-E. Initial review authorizes G128-H only. After G128-H changes the trust boundary and renewed G128-E commits the hardened wrapper, repeat the review and bind the literal absolute interpreter path/hash/version/signature plus exact invocation; PATH pwsh, powershell.exe and every fallback are forbidden.</automated></verify>
<done>Initial and renewed independent reviews each record zero HIGH; only the renewed G128-S proves the hardened wrapper is read-only and may precede G128-D.</done>
<atomic_allowlist>Read-only authority review and its bounded receipt; no implementation mutation.</atomic_allowlist>
</task>

<task id="128-H">
<requirements>TASK-204; architecture sections 1,5,6,12,14; host-interpreter TCB incident repair.</requirements><owner>devops-engineer</owner><verifier>security-auditor; arch-ver</verifier>
<read_first>Accepted initial G128-S receipt; Windows host-mutation approval; Microsoft WinGet Microsoft.PowerShell package metadata; agent safety rule; validate-contract.ps1.</read_first>
<consumes>Accepted initial G128-S as review evidence only. It does not consume or spend a G128-D invocation.</consumes>
<produces>A separately committed .planning/audits/test-rebase/powershell-host-lock.json that seals the only permitted host interpreter TCB. The later renewed G128-E task, not this planning repair, hardens and commits validate-contract.ps1 against that lock.</produces>
<files>.planning/audits/test-rebase/powershell-host-lock.json; .planning/audits/test-rebase/validate-contract.ps1</files>
<failing_sentinel>Fail closed before host mutation without explicit approval, or on any package other than exact Microsoft.PowerShell from the Microsoft-supported WinGet source, non-stable/pre-release selection, unresolved/non-absolute/reparse interpreter path, missing SHA256/version/signature/package provenance, non-Valid Authenticode status, or signer/thumbprint drift. PATH pwsh, powershell.exe, aliases, shims and fallback interpreters are forbidden.</failing_sentinel>
<verify><automated>Execution remains NOT_READY now. After explicit host-mutation approval only, provision official stable PowerShell 7 with WinGet exact package id Microsoft.PowerShell and source winget; independently inspect the installed executable without invoking the guard. Create and commit the exact lock with absolute_path, sha256, file_version, product_version, authenticode_status, authenticode_signer, authenticode_thumbprint, package_id, package_version, package_source_url and install_method. Security and architecture review the package provenance and executable identity. No Python, Docker, SUT, validator, builder or G128-D run.</automated></verify>
<done>The exact host lock is committed and independently accepted; runtime is still NOT_RUN. G128-D remains NOT_READY until validate-contract.ps1 is hardened in a renewed G128-E envelope commit and a renewed G128-S receipt seals its new bytes and literal absolute invocation.</done>
<atomic_allowlist>Commit only .planning/audits/test-rebase/powershell-host-lock.json. The renewed G128-E task later commits only validate-contract.ps1; no installer artifacts, reports, manifest or unrelated host/repository paths.</atomic_allowlist>
</task>

<task id="128-D">
<requirements>TASK-204; architecture sections 1–15.</requirements><owner>qa-auto-ver</owner><verifier>gsd-plan-checker</verifier>
<read_first>Accepted G128-H host lock, renewed G128-E envelope, renewed G128-S review, sealed authority, wrapper source and G-META receipt.</read_first>
<consumes>G128-H plus the renewed G128-E and renewed G128-S receipts, including the independently sealed literal absolute pwsh path and hardened wrapper identity.</consumes>
<produces>One conclusion-blind G128-D receipt.</produces>
<files>testframework/reports/&lt;task-run&gt;/ only; all guard sources are read-only.</files>
<failing_sentinel>Fail on mutable authority, interpreter path/hash/version/signature drift, PATH resolution, powershell.exe or fallback use, missing cleanup proof, timeout ambiguity, or a second consumed historical guard run.</failing_sentinel>
<verify><automated>NOT_READY until G128-H and renewed G128-E/G128-S. Invoke only the literal absolute pwsh.exe path copied verbatim from the accepted host lock into renewed G128-S; no PATH lookup, variable fallback, powershell.exe, alias or shim is permitted. One Docker-owned historical guard is permitted only by that sealed receipt contract and is forbidden after G129-L.</automated></verify>
<done>Independent receipt proves determinism, timeout, exact cleanup, and no SUT/cutover/apply/publish path.</done>
<atomic_allowlist>One task-owned report receipt only; no authority materialization.</atomic_allowlist>
</task>

## Founder gate

The initial G128-E precedes initial G128-S. Initial G128-S authorizes only
G128-H. Accepted G128-H requires renewed G128-E and renewed G128-S because the
wrapper bytes and trust boundary change; only that renewed G128-S may precede
G128-D, and G128-D is a predecessor of G129-L only. The current prelaunch
attempt is non-consumed: G128-H is `NOT_READY`, runtime is `NOT_RUN`, and the
one permitted G128-D invocation remains unused.
`STATIC_REVIEW_PASS` is not runtime
`PASS`; all executable states remain `NOT_READY` until their named gates.
