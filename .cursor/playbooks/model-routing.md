# Hub — Model Routing

> **FOUNDER LOCK — 2026-08-22 (supersedes every conflicting row below):**
> root `team-lead` is the sole `gpt-5.6-sol` runtime and always uses `max`.
> Every downstream dispatch explicitly requests `gpt-5.6-terra` and an explicit,
> task-routed `reasoning_effort`. `inherit`, omitted fields, Sol subagents and
> silent fallback are forbidden. Atomic implementation begins only after
> root/Terra freezes the contract and always uses Terra / `xhigh`.

This is the canonical runtime-routing policy for the Hub AI team. Role names are
organizational labels, not evidence of an observed runtime.

## Mandatory dispatch guard

Before every downstream spawn, the orchestrator validates all facts below:

1. `model` is present and exactly `gpt-5.6-terra`.
2. `reasoning_effort` is present and matches the task class: `low` for
   deterministic mechanics/hash/provenance; `medium` for ordinary analysis or
   authoring; `high` for independent review; `xhigh` for unusually complex
   downstream audits and every atomic implementation.
3. Atomic implementation has a frozen contract, exact file/function ownership,
   tests or static acceptance evidence, and no unresolved architecture,
   security, or product decision.
4. The request uses a bounded/no-history fork and never inherits the root model.
5. The handoff records requested and observed runtime; without telemetry,
   `actual_model: UNKNOWN`.
6. Exact-runtime unavailability or mismatch returns `UNKNOWN/BLOCKED`; no
   automatic model or effort substitution is allowed.

Failure of any guard means **do not dispatch**. No downstream subagent uses Sol.

## Routing matrix

| Work class | Runtime | Responsibility |
|---|---|---|
| Root orchestration, architecture, final acceptance | Sol / Max | root `team-lead` only |
| Mechanical identity/hash/provenance | Terra / Low | deterministic downstream mechanics |
| Ordinary analysis, decomposition, coordination, authoring | Terra / Medium | bounded role work |
| Independent review | Terra / High | conclusion-blind review |
| Complex cross-domain or security audit | Terra / XHigh | bounded downstream audit |
| Atomic implementation/test code after frozen contract | Terra / XHigh | one function, small module, test, or behaviour; exact-path bounded |

Architecture, security, and product choices stay with root/Terra. An atomic
implementer does not broaden ownership; an incomplete contract returns
`BLOCKED` for further root/Terra decomposition.

## Role defaults

| Roles | Default | Escalate to | Scope |
|---|---|---|---|
| `team-lead` | Sol / Max | — | sole root/orchestrator runtime |
| `tech-lead`, `product`, `arch`, security roles | Terra / High | Terra / XHigh for bounded complex work | analysis, decisions prepared for root |
| `ba*`, `sa*`, `be*`, `fe*`, `qa-*`, docs/devops roles | Terra / Medium | Terra / High for independent review or complex analysis | owned bounded work |
| `*-ver`, `code-reviewer` | Terra / High | Terra / XHigh for complex bounded audit | independent verification |
| `codex-micro` | Terra / XHigh | STOP → root/Terra decomposition | legacy role ID; atomic implementation only |

`codex-micro` is a stable legacy role identifier, not a Codex-runtime
requirement. Its dispatch is always `gpt-5.6-terra` / `xhigh` with exact-path
ownership.

## Dispatch and handoff provenance

```text
requested_model: gpt-5.6-terra  # REQUIRED; never inherit/omit
actual_model: <observed provider/model | UNKNOWN>
actual_model_evidence: <launcher telemetry/run id/log ref | none>
provider: <observed | UNKNOWN>
reasoning_effort: <low|medium|high|xhigh>  # REQUIRED; atomic=xhigh
reason: <task class + why this exact effort>
fallback: none; unavailable runtime => UNKNOWN/BLOCKED; Sol subagent forbidden
catalog_status: available | unavailable | not_checked
workflow: <name@revision | UNKNOWN>
skills: <name@revision | none | UNKNOWN>
plugins: <name@revision | none | UNKNOWN>
MCP: <server@revision + capability scope | none | UNKNOWN>
revision: <branch + full HEAD or artifact digest>
```

`requested_model` is intent, not telemetry. Never infer `actual_model` from
requested configuration, role name, frontmatter, or an expected fallback.

## Quick choice

```text
Orchestration/architecture/final acceptance?  root Sol / Max
Mechanical provenance?                         Terra / Low
Ordinary analysis or authoring?                 Terra / Medium
Independent review?                             Terra / High
Complex bounded audit?                          Terra / XHigh
Atomic code after frozen contract?              Terra / XHigh
```

*Founder amendment 2026-08-22 · explicit Terra downstream; no inherit or silent fallback*
