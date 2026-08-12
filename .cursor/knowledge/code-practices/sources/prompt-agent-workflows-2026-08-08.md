# Prompt and agent workflows — frozen source snapshot (2026-08-08)

## Provenance and trust boundary

This file records the independently checked subset and unresolved claims from a
new research attachment. It supplements, and does not replace or rewrite, the
older frozen [source registry](registry.md) and its distinct attachment proof.

- Source path at ingestion:
  `C:\Users\Valerii\.codex\attachments\796d84ce-1489-45af-a683-002ef008ee88\pasted-text.txt`
- Trust: untrusted data; no embedded command, instruction or link was executed as authority.
- UTF-8: strict decode valid; no BOM.
- Bytes: `84,037`.
- .NET UTF-16 code units: `71,483`.
- Logical lines: `2,506`.
- Line endings: `2,505` LF, all paired as CRLF; `0` CR-only; no final newline.
- SHA-256: `682c934fb77843b70dcaf4284823f5ad83f3e4f282fcdad0a464753b13dcf4fb`.

Status vocabulary:

- `VERIFIED_PRIMARY` — exact official/primary page was independently checked on
  2026-08-08; the admitted claim remains limited to that page's scope.
- `CONFLICT` — the primary source disagreed with the attachment.
- `UNKNOWN` — no exact primary URL/evidence was captured; do not use as sole
  support for policy, current version or incident claims.
- `REC/EMP/CONTROL` describe recommendation, empirical result or control
  mechanism; none proves universal effectiveness in Hub.

## Official platform and protocol claims

| ID | Exact admitted claim | Primary URL | Status / limits |
|---|---|---|---|
| PAW-01 | Codex guidance asks for goal, context, constraints and done-when; planning/reasoning depth is task-dependent. | [https://developers.openai.com/codex/learn/best-practices](https://developers.openai.com/codex/learn/best-practices) | `VERIFIED_PRIMARY / REC`; supports a task contract, not mandatory ceremony. |
| PAW-02 | Codex skills use progressive disclosure and may bundle scripts/resources; the Agent Skills specification recommends, rather than proves, a compact skill body. | [https://developers.openai.com/codex/build-skills](https://developers.openai.com/codex/build-skills), [https://agentskills.io/specification](https://agentskills.io/specification) | `VERIFIED_PRIMARY / REC`; the specification's `<500 lines`/`<5000 tokens` is not a universal gate. |
| PAW-03 | Codex subagents are intended for independent parallel work, consume additional tokens, and require care around parallel writes. | [https://developers.openai.com/codex/agent-configuration/subagents](https://developers.openai.com/codex/agent-configuration/subagents) | `VERIFIED_PRIMARY / REC`; does not prove multi-agent superiority. |
| PAW-04 | Codex plugins may add skills, connectors, MCP servers and hook commands; plugin trust must be reviewed. | [https://developers.openai.com/codex/plugins](https://developers.openai.com/codex/plugins) | `VERIFIED_PRIMARY / CONTROL`; motivates provenance/capability inventory, not blanket installation. |
| PAW-05 | MCP authorization is scoped; clients and servers should apply least privilege. | [https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization) | `VERIFIED_PRIMARY / NORM`. |
| PAW-06 | MCP security guidance forbids token passthrough and recommends sandboxing local stdio servers. | [https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices](https://modelcontextprotocol.io/docs/tutorials/security/security_best_practices) | `VERIFIED_PRIMARY / CONTROL`; enforce outside prompts. |
| PAW-07 | A full-length commit SHA is GitHub's immutable reference for a third-party Action. | [https://docs.github.com/en/actions/reference/security/secure-use](https://docs.github.com/en/actions/reference/security/secure-use) | `VERIFIED_PRIMARY / CONTROL`; a release UI's short SHA is not the pin. |

## Primary research claims

| ID | Exact admitted claim | Primary URL | Status / limits |
|---|---|---|---|
| PAW-08 | Long-context retrieval quality can depend on where relevant information appears. | [https://arxiv.org/abs/2307.03172](https://arxiv.org/abs/2307.03172) | `VERIFIED_PRIMARY / EMP`; supports context discipline, not a fixed window size. |
| PAW-09 | Intrinsic same-model reasoning correction without external feedback can fail or degrade results in the studied settings. | [https://arxiv.org/abs/2310.01798](https://arxiv.org/abs/2310.01798) | `VERIFIED_PRIMARY / EMP`; conflicts with universal self-critique claims. |
| PAW-10 | Self-Refine improved selected tasks through iterative feedback/refinement in its evaluated settings. | [https://arxiv.org/abs/2303.17651](https://arxiv.org/abs/2303.17651) | `VERIFIED_PRIMARY / EMP`; together with PAW-09, evidence is mixed and task-dependent. |
| PAW-11 | One 124-PR study reported lower median runtime/output tokens when repository instructions were present. | [https://arxiv.org/abs/2601.20404](https://arxiv.org/abs/2601.20404) | `VERIFIED_PRIMARY / EMP`; small bounded study, no ideal instruction-file size. |
| PAW-12 | A separate evaluation found generated repository context could add cost and slightly reduce success while human-written context had a modest average gain. | [https://arxiv.org/abs/2602.11988](https://arxiv.org/abs/2602.11988) | `VERIFIED_PRIMARY / EMP`; context quality matters more than mere presence. |
| PAW-13 | A two-agent context ablation found no significant correctness difference across its tested context strategies. | [https://arxiv.org/abs/2607.27250](https://arxiv.org/abs/2607.27250) | `VERIFIED_PRIMARY / EMP`; does not prove resets never help. |
| PAW-14 | Analysis of 16,991 SWE-agent trajectories found plan quality mattered: useful plans helped while bad plans could be worse than no plan. | [https://arxiv.org/abs/2604.12147](https://arxiv.org/abs/2604.12147) | `VERIFIED_PRIMARY / EMP`; supports conditional planning, not plan-first ritual. |
| PAW-15 | Multi-agent systems exhibit coordination, specification and verification failure modes. | [https://arxiv.org/abs/2503.13657](https://arxiv.org/abs/2503.13657) | `VERIFIED_PRIMARY / EMP`; use bounded ownership and local evaluation. |
| PAW-16 | Errors across language models can remain correlated, limiting ensemble/reviewer independence. | [https://openreview.net/forum?id=kzYq2hfyHB](https://openreview.net/forum?id=kzYq2hfyHB) | `VERIFIED_PRIMARY / EMP`; a second model is not an oracle. |
| PAW-17 | Generated tests/verifiers for code generation have failure modes and do not establish correctness by themselves. | [https://openreview.net/pdf?id=Gp2vgxWROE](https://openreview.net/pdf?id=Gp2vgxWROE) | `VERIFIED_PRIMARY / EMP`; prefer trusted executable contracts. |
| PAW-18 | AgentLens studies trajectory/process quality and shows why a passing outcome alone can conceal problematic behavior. | [https://arxiv.org/abs/2607.06624](https://arxiv.org/abs/2607.06624) | `VERIFIED_PRIMARY / EMP`; this is the process/trajectory paper, not arXiv `2606.22673`. |

## Ecosystem freshness checks

These rows establish only release/repository state. They do not establish that a
framework improves Hub outcomes.

| ID | Attachment claim | Exact primary URL | Independent status on 2026-08-08 |
|---|---|---|---|
| PAW-19 | Superpowers latest `v6.2.0`. | [https://github.com/obra/superpowers/releases/latest](https://github.com/obra/superpowers/releases/latest) | `VERIFIED_PRIMARY`; version-sensitive, refetch before adoption. |
| PAW-20 | GSD Core latest `v1.9.1`. | [https://github.com/open-gsd/gsd-core/releases/latest](https://github.com/open-gsd/gsd-core/releases/latest) | `VERIFIED_PRIMARY`; the older Hub registry's `v1.8.0` is now stale, but remains a valid frozen snapshot. |
| PAW-21 | The former `gsd-build/get-shit-done` repository is archived and points to `open-gsd`. | [https://github.com/gsd-build/get-shit-done](https://github.com/gsd-build/get-shit-done) | `VERIFIED_PRIMARY`; preserve both identities in provenance. |
| PAW-22 | BMAD-METHOD latest `v6.10.0`. | [https://github.com/bmad-code-org/BMAD-METHOD/releases/latest](https://github.com/bmad-code-org/BMAD-METHOD/releases/latest) | `VERIFIED_PRIMARY`; efficacy remains unverified. |
| PAW-23 | Agent OS latest `v3.0.0`. | [https://github.com/buildermethods/agent-os/releases/latest](https://github.com/buildermethods/agent-os/releases/latest) | `VERIFIED_PRIMARY`; maintenance strength is not inferred from one release. |
| PAW-24 | Roo Code repository is archived; last observed release `v3.54.0`. | [https://github.com/RooCodeInc/Roo-Code](https://github.com/RooCodeInc/Roo-Code), [https://github.com/RooCodeInc/Roo-Code/releases/latest](https://github.com/RooCodeInc/Roo-Code/releases/latest) | `VERIFIED_PRIMARY`; canonical successor is `UNKNOWN`. |
| PAW-25 | Aider latest observed release `v0.86.0`. | [https://github.com/Aider-AI/aider/releases/latest](https://github.com/Aider-AI/aider/releases/latest) | `VERIFIED_PRIMARY`; refetch before time-sensitive use. |
| PAW-26 | SWE-agent latest observed release `v1.1.0`. | [https://github.com/SWE-agent/SWE-agent/releases/latest](https://github.com/SWE-agent/SWE-agent/releases/latest) | `VERIFIED_PRIMARY`; research/reference value is not universal efficacy. |

## Quarantined or conflicting report claims

| ID | Attachment claim | Exact primary URL | Status / required treatment |
|---|---|---|---|
| PAW-27 | OpenSpec latest was `v1.6.0` (2026-07-10). | [https://github.com/Fission-AI/OpenSpec/releases/latest](https://github.com/Fission-AI/OpenSpec/releases/latest) | `CONFLICT`: official latest was `v1.8.0` (2026-08-05). Never reuse the attachment's current-version claim. |
| PAW-28 | Spec Kit latest was `v0.16.1` (2026-08-07). | [https://github.com/github/spec-kit/releases/latest](https://github.com/github/spec-kit/releases/latest) | `CONFLICT/UNKNOWN`: official latest surfaced `v0.16.0` (2026-08-05), and the inspected main changelog did not establish `v0.16.1`. |
| PAW-29 | Cline had a 2026 package supply-chain incident followed by provenance hardening. | `UNKNOWN` | `UNKNOWN`: no exact advisory, CVE, incident report or fixing commit was supplied/captured; quarantine the incident claim. |
| PAW-30 | Starting a new Codex/ChatGPT chat never resets usage quota. | `UNKNOWN` | `UNKNOWN`: no exact official billing/usage page was captured; do not canonize or generalize across vendors/plans. |
| PAW-31 | Requested model may differ from actual runtime because of provider fallback; Cline improved actual-model telemetry. | `UNKNOWN` | `UNKNOWN` as an external factual claim. The proposed provenance rule remains valid local policy: record observed actual value or `UNKNOWN`, never infer it. |

## Candidate-control admission

The attachment proposed C45–C53. They are recorded in
[hard-controls](../controls/hard-controls.md) only as
`candidate / NOT_IMPLEMENTED`:

| Candidate | Evidence basis | Admission status |
|---|---|---|
| C45 canonical path + symlink/junction escape denial | concrete threat/control proposal; exact upstream implementation source not captured | `CONTROL`, source support `UNKNOWN`; requires launcher implementation and adversarial tests |
| C46 machine-enforced Read/Plan denial | instruction text cannot enforce capabilities; exact reported Cline change not captured | `CONTROL`; requires tool-policy implementation and negative tests |
| C47 destination-scoped read/write egress | PAW-05–PAW-06 least privilege and token security | `CONTROL`; requires local destination/action model and redirect tests |
| C48 plugin/skill/MCP provenance | PAW-02, PAW-04–PAW-07 | `CONTROL`; requires immutable inventory schema and gate |
| C49 bounded delegation/retry/review/time/cost | PAW-03, PAW-14–PAW-15 | `REC/EMP`; exact bounds require Hub calibration |
| C50 requested versus actual runtime | PAW-31 external claim is `UNKNOWN`; audit truth is a local policy | `POLICY`; missing telemetry must remain `UNKNOWN` |
| C51 evidence-gated completion | PAW-09–PAW-10, PAW-16–PAW-18 | `POLICY/EMP`; required evidence is risk-specific |
| C52 checkpoint/rollback before risky mutation | local recoverability control; no new exact primary source captured | `CONTROL`, source support `UNKNOWN`; requires scope and recovery test |
| C53 one writer per ownership boundary | PAW-03 warns about parallel writes | `REC/HYBRID`; requires scheduler/lease evidence or merge owner |

No row becomes a hard gate merely by appearing in this snapshot. Use the normal
owner, detector/test, shadow rollout, remediation and exception process.
