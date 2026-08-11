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
envelope and static receipt; then a no-install seal of the built-in Windows
PowerShell 5.1 TCB plus an exact compatibility-remediation contract, a renewed
envelope/static receipt after wrapper hardening, and one independently reviewed,
one-time read-only guard receipt.

## Tasks

<task id="128-E">
<requirements>TASK-202, TASK-203, TASK-204; architecture sections 1,2,5,6,12,14.</requirements><owner>tech-lead</owner><verifier>arch-ver; security-auditor</verifier>
<read_first>Accepted G128-R receipts; exact Git tracking state; wrapper plus every executable/read authority it consumes; agent safety rule; for renewal only, accepted G128-WPS51 receipt and powershell-host-lock.json.</read_first>
<consumes>Initially G128-R commit fd3a5671 and immutable architecture, ledger, founder and Phase 129 identities; on the mandatory renewal, accepted G128-WPS51 and its committed host lock and compatibility findings.</consumes>
<produces>Initially one committed full-HEAD source envelope; after G128-WPS51, one renewed envelope commit/review binding the PS5.1-compatible hardened wrapper and host lock. Neither pass produces a runtime receipt or authority transition.</produces>
<files>.planning/audits/test-rebase/validate-contract.ps1; .planning/audits/test-rebase/apply_cutover_batch.py; .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json; .planning/audits/test-rebase/inventory.py; .planning/audits/test-rebase/inventory.json; .planning/audits/test-rebase/inventory.schema.json; .planning/audits/test-rebase/test-architecture-contract.json; .planning/audits/test-rebase/validate_test_architecture_contract.py; .planning/audits/test-rebase/build_cutover_manifest.py; .planning/audits/test-rebase/validate_cutover_manifest.py; .planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md; .planning/audits/test-rebase/powershell-host-lock.json</files>
<failing_sentinel>Fail while any of the first five writable envelope paths is dirty/untracked, any of the eleven paths lacks a same-HEAD blob, or an immutable identity differs from 1049 / 507-245-297+28 / 173-1120-max56.</failing_sentinel>
<verify><automated>Static read-only source and Git-tree review only; no Python, validator, builder, Docker, SUT, publish, apply, rollback or report. Initial review accepts the exact five-path writable diff plus six read-only same-HEAD dependencies. Mandatory post-G128-WPS51 renewal accepts the separately committed host lock and commits/reviews only the hardened validate-contract.ps1; it replaces every independently identified PS7/.NET-Core-only API, uses an exact Windows argv encoder with bounded selftest vectors, and launches each child race-free with CreateProcessW CREATE_SUSPENDED, AssignProcessToJobObject before ResumeThread, JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE, no breakaway/cmd, bounded wait/terminate/drain/finally handle closure and ActiveProcesses=0 or NOT_READY. It proves fail-closed equality for current interpreter canonical path, held-handle volume serial/file index/reparse/link set, SHA256, file/product/PowerShell/CLR version and the exact Microsoft publisher/production-chain/code-signing identity.</automated></verify>
<done>The initial eleven-path envelope is accepted; after G128-WPS51, the twelve-path envelope is rebound to the new committed PS5.1-compatible wrapper and host lock, architecture/security reviews have zero HIGH, and runtime remains NOT_RUN.</done>
<atomic_allowlist>Initial commit: only validate-contract.ps1, apply_cutover_batch.py, founder-greenfield-cutover-manifest.json, inventory.py and inventory.json. Renewal commit: only validate-contract.ps1. Exclude VERIFICATION, baseline/debug, reports, cutover-manifest, global planning/product dirt and pro-qa-ai.</atomic_allowlist>
</task>

<task id="128-S">
<requirements>TASK-202, TASK-203, TASK-204; architecture sections 1–15.</requirements><owner>tech-lead</owner><verifier>gsd-plan-checker</verifier>
<read_first>Accepted G128-E commit/tree and reviews; Phase 128 seal, validate-contract.ps1, current manifest and G-META receipt; for renewal only, accepted G128-WPS51 host lock and renewed G128-E.</read_first>
<consumes>Initial or renewed G128-E committed envelope; G-META; immutable Phase 128 authority. The initial receipt can authorize only G128-WPS51; G128-D consumes only the renewed receipt.</consumes>
<produces>Initial static review of the read-only wrapper boundary, then a mandatory renewed review after PS5.1 compatibility hardening that seals the literal absolute Windows PowerShell path and exact G128-D command.</produces>
<files>.planning/audits/test-rebase/validate-contract.ps1; .planning/audits/test-rebase/apply_cutover_batch.py; .planning/audits/test-rebase/founder-greenfield-cutover-manifest.json; .planning/audits/test-rebase/inventory.py; .planning/audits/test-rebase/inventory.json; .planning/audits/test-rebase/inventory.schema.json; .planning/audits/test-rebase/test-architecture-contract.json; .planning/audits/test-rebase/validate_test_architecture_contract.py; .planning/audits/test-rebase/build_cutover_manifest.py; .planning/audits/test-rebase/validate_cutover_manifest.py; .planning/decisions/ADR-TEST-ARCHITECTURE-REBASE-2026.md; .planning/audits/test-rebase/powershell-host-lock.json; .planning/audits/test-rebase/VERIFICATION.md</files>
<failing_sentinel>Fail if the wrapper exposes apply, publish, cutover, SUT-test, or unsealed write authority.</failing_sentinel>
<verify><automated>NOT_READY until the applicable G128-E. Initial review authorizes G128-WPS51 only. After G128-WPS51 establishes the no-install trust boundary and renewed G128-E commits the PS5.1-compatible hardened wrapper, repeat architecture and security review and bind literal C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe, its held-handle identity/hash/version/publisher-chain signature plus the exact invocation. Unqualified powershell.exe, pwsh, PATH/PATHEXT discovery and every fallback are forbidden.</automated></verify>
<done>Initial and renewed independent reviews each record zero HIGH; only the renewed G128-S proves the hardened wrapper is read-only and may precede G128-D.</done>
<atomic_allowlist>Read-only authority review and its bounded receipt; no implementation mutation.</atomic_allowlist>
</task>

<task id="128-WPS51">
<requirements>TASK-204; architecture sections 1,5,6,12,14; host-interpreter TCB incident repair.</requirements><owner>devops-engineer</owner><verifier>security-auditor; arch-ver</verifier>
<read_first>Accepted initial G128-S receipt; founder no-install decision; independent Windows PowerShell 5.1 compatibility audit; agent safety rule; validate-contract.ps1; exact built-in C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe.</read_first>
<consumes>Accepted initial G128-S as review evidence only. It does not consume or spend a G128-D invocation.</consumes>
<produces>A separately committed .planning/audits/test-rebase/powershell-host-lock.json that seals the preinstalled Windows inbox interpreter TCB and exact accepted hardlink set, plus an independently accepted PS5.1 remediation contract. The later renewed G128-E task, not this planning repair, hardens and commits validate-contract.ps1 against that lock.</produces>
<files>.planning/audits/test-rebase/powershell-host-lock.json; .planning/audits/test-rebase/validate-contract.ps1</files>
<failing_sentinel>Fail closed on any install/download/package-manager action; any interpreter except exact C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe; unresolved/non-absolute identity; held-handle reparse flag not false; volume serial/file index/hardlink count or exact System32/WinSxS link-set drift; missing SHA256/file/product/PowerShell/CLR/build/signature provenance; Authenticode not Valid; publisher outside Microsoft Windows Publisher; leaf/intermediate/root outside the sealed Microsoft Windows production chain; missing code-signing EKU OID 1.3.6.1.5.5.7.3.3; signer/issuer/thumbprint/chain drift; an incomplete PS5.1 remediation inventory; unqualified powershell.exe, pwsh, PATH/PATHEXT lookup, alias, shim or fallback.</failing_sentinel>
<verify><automated>Execution remains NOT_READY. Perform bounded read-only identity inspection of the already installed Windows PowerShell 5.1 inbox component; install nothing and do not invoke the guard. Open the canonical executable once with no-follow/no-reparse semantics and keep that handle held while GetFileInformationByHandle captures and later rechecks volume_serial, file_index, link count and reparse_point=false and while every enumerated System32/WinSxS hardlink is reopened no-follow and matched to the same file identity. Create and commit the exact lock with schema_version, absolute_path, volume_serial, file_index, reparse_point, sha256, file_version, product_version, powershell_version, ps_edition, clr_version, authenticode_status, authenticode_publisher, authenticode_signer, authenticode_issuer, authenticode_thumbprint, authenticode_chain, authenticode_eku_oids, provenance_type, windows_build, component_target, hardlink_count, accepted_hardlink_paths and inspected_at, with no extras. Require Microsoft Windows Publisher, an exact pinned Microsoft Windows production leaf/intermediate/root chain and code-signing EKU 1.3.6.1.5.5.7.3.3; arbitrary Authenticode Valid is insufficient. The hardlink exception is scoped only to this signed Windows inbox component at that held exact file identity with hardlink_count=2 and exact System32 plus WinSxS paths; no generic nlink relaxation is allowed. Independently review the exact compatibility remediation: RNG.Create/GetBytes/Dispose plus BitConverter hex; ProcessStartInfo.EnvironmentVariables; null/NUL-rejecting Windows argv encoding with quote/backslash/empty/non-ASCII selftest vectors; CreateProcessW CREATE_SUSPENDED then AssignProcessToJobObject before ResumeThread, JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE, no breakaway/cmd, bounded wait/TerminateJobObject/drain/finally closure and ActiveProcesses=0 or NOT_READY; current-process identity via Diagnostics.Process.GetCurrentProcess().MainModule.FileName. No Python, Docker, SUT, validator, builder or G128-D run.</automated></verify>
<done>The exact host lock is committed and independently accepted; runtime is still NOT_RUN. G128-D remains NOT_READY until validate-contract.ps1 is hardened in a renewed G128-E envelope commit and a renewed G128-S receipt seals its new bytes and literal absolute invocation.</done>
<atomic_allowlist>Commit only .planning/audits/test-rebase/powershell-host-lock.json. The renewed G128-E task later commits only validate-contract.ps1; no installer/download/package artifacts, reports, manifest or unrelated host/repository paths.</atomic_allowlist>
</task>

<task id="128-D">
<requirements>TASK-204; architecture sections 1–15.</requirements><owner>qa-auto-ver</owner><verifier>gsd-plan-checker</verifier>
<read_first>Accepted G128-WPS51 host lock, renewed G128-E envelope, renewed G128-S review, sealed authority, wrapper source and G-META receipt.</read_first>
<consumes>G128-WPS51 plus the renewed G128-E and renewed G128-S receipts, including the independently sealed literal absolute Windows PowerShell path and hardened wrapper identity.</consumes>
<produces>One conclusion-blind G128-D receipt.</produces>
<files>testframework/reports/&lt;task-run&gt;/ only; all guard sources are read-only.</files>
<failing_sentinel>Fail on mutable authority, interpreter path/hash/version/signature/hardlink drift, PATH resolution, unqualified powershell.exe, any pwsh or fallback use, missing cleanup proof, timeout ambiguity, or a second consumed historical guard run.</failing_sentinel>
<verify><automated>NOT_READY until G128-WPS51 and renewed G128-E/G128-S. Invoke exactly C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -NoProfile -NonInteractive -File .planning/audits/test-rebase/validate-contract.ps1 as sealed verbatim by renewed G128-S; no PATH lookup, variable fallback, unqualified powershell.exe, pwsh, alias, shim, ExecutionPolicy bypass or extra flag is permitted. One Docker-owned historical guard is permitted only by that sealed receipt contract and is forbidden after G129-L.</automated></verify>
<done>Independent receipt proves determinism, timeout, exact cleanup, and no SUT/cutover/apply/publish path.</done>
<atomic_allowlist>One task-owned report receipt only; no authority materialization.</atomic_allowlist>
</task>

## Founder gate

The initial G128-E precedes initial G128-S. Initial G128-S authorizes only
G128-WPS51. Accepted G128-WPS51 requires renewed G128-E and renewed G128-S because the
wrapper bytes and trust boundary change; only that renewed G128-S may precede
G128-D, and G128-D is a predecessor of G129-L only. The current prelaunch
attempt is non-consumed: G128-WPS51 is `NOT_READY`, runtime is `NOT_RUN`, and the
one permitted G128-D invocation remains unused.
`STATIC_REVIEW_PASS` is not runtime
`PASS`; all executable states remain `NOT_READY` until their named gates.
