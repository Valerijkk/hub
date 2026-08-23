# ADR-MODELS-GPT56-2026: Sol orchestrator → Terra downstream

- **Status:** accepted, amended 2026-08-22
- **Decision date:** 2026-07-18
- **Current founder lock:** 2026-08-22
- **Author:** founder + docs-sync · **Verifiers:** team-lead, product-ver
- **Related:** `.cursor/playbooks/model-routing.md`, `team-workflow.md`,
  `.cursor/REGULATION.md`, `AGENTS.md`, `.cursor/agents/*`

---

## Context

The prior atomic-implementation routing was erroneous. The founder amendment
of 2026-08-22 replaces it: all downstream work uses one explicit Terra runtime.
Root Sol / Max remains unchanged.

## Decision

| Layer | Exact runtime | Who | Responsibility |
|---|---|---|---|
| Orchestrator | `gpt-5.6-sol` / `max` | root `team-lead` | goal, architecture, security/product decisions, waves, final acceptance |
| Downstream | `gpt-5.6-terra` / task-routed effort | all downstream role profiles | analysis, decomposition, coordination, authoring, independent review, bounded implementation |
| Atomic implementation | `gpt-5.6-terra` / `xhigh` | exact-path bounded downstream owner | one function, small module, focused test, or one behaviour after a frozen contract |

### Founder locks

1. Root is the sole Sol runtime; Sol subagents are forbidden.
2. Every downstream frontmatter and dispatch contains literal
   `model: gpt-5.6-terra` and an explicit `reasoning_effort`.
3. `inherit`, omitted model/effort, full-history model inheritance, and silent
   fallback are forbidden.
4. Effort is task-routed: `low` for deterministic mechanics/hash/provenance;
   `medium` for ordinary analysis or authoring; `high` for independent review;
   `xhigh` for unusually complex downstream audits and every atomic
   implementation.
5. Atomic implementation begins only after root/Terra freezes the contract. It
   is limited to one function, small module, focused test, or one behaviour,
   with exact paths and acceptance evidence. Architecture, security, and
   product decisions remain with root/Terra.
6. Unavailable exact runtime, model mismatch, or missing required telemetry
   returns `UNKNOWN/BLOCKED`; no automatic model or effort substitution exists.
7. Requested runtime is not observed runtime. Handoff records observed model
   evidence or `UNKNOWN`, effort, and exact revision.
8. A model does not replace an independent verifier, Docker-only evidence, test
   triage, docs sync, or security/authority gates.

## Effort routing

| Class | Runtime |
|---|---|
| Root orchestration, architecture, final acceptance | Sol / Max |
| Mechanical identity/hash/provenance | Terra / Low |
| Ordinary analysis, decomposition, authoring | Terra / Medium |
| Independent review | Terra / High |
| Complex cross-domain/security audit | Terra / XHigh |
| Atomic implementation/test code after frozen contract | Terra / XHigh |

If a brief is incomplete or the task is not atomic, the downstream owner
returns `BLOCKED`; root/Terra clarifies the contract or decomposes further.

## Frontmatter (applied 2026-08-22)

| Profiles | `model:` |
|---|---|
| `team-lead` | `gpt-5.6-sol` |
| every downstream profile, including `codex-micro` and deprecated redirects | `gpt-5.6-terra` |

`codex-micro` remains a stable legacy role ID; it is not a Codex-runtime
requirement. Its atomic dispatch remains Terra / `xhigh` and exact-path bounded.

Frontmatter does not encode effort, so every launch contract explicitly passes
`reasoning_effort`. A runtime that cannot accept it is not silently compliant.

## Dispatch examples

| Task | Route |
|---|---|
| Milestone, architecture, conflict resolution, final seal | root Sol / Max |
| Decompose a feature/API contract into executable backlog | Terra / High–XHigh |
| Review implementation or evidence conclusion-blind | Terra / High |
| Write one Go function plus regression test by frozen contract | Terra / XHigh |
| Implement a small React hook/component by DESIGN-SPEC | Terra / XHigh |
| Write a Python helper or one test behaviour | Terra / XHigh |

Anti-patterns: Sol downstream; atomic implementation without a frozen contract;
automatic fallback; inherited model/effort; or treating the legacy
`codex-micro` name as a runtime request.

## Protected external zone

`pro-qa-ai/**` remains the founder-operated Claude Code Opus 5 zone. Internal
Hub agents do not read or change it without a literal founder handoff. This ADR
does not expand any model's data or process authority.

## Applied checklist

1. `[done]` root `team-lead` → `gpt-5.6-sol`; effort `max`.
2. `[done]` every downstream profile, including `codex-micro` →
   `gpt-5.6-terra`.
3. `[done]` all atomic implementation dispatches → Terra / `xhigh`, with a
   frozen contract and exact-path boundary.
4. `[done]` `inherit`, silent fallback, and Sol subagents prohibited.
5. `[done]` active obsolete pricing/source rationale removed from routing.
6. `[done]` `pro-qa-ai/**` remains an external protected zone.
7. `[pending]` Representative measured routing/eval pilot; documentation review
   does not substitute for runtime-quality evidence.

---

*Accepted 2026-07-18 · current amendment 2026-08-22 · explicit Terra downstream; no inherit*
