# Phase 143 code-runtime routing record

> **Detected:** 2026-08-13
> **Scope:** atomic production/test code only
> **Planning/review:** continues
> **Current status:** ROUTE_AUTHORIZED · DISPATCHER_ENFORCED_PROVENANCE_ACCEPTED · RUNTIME_NOT_RUN

## Current route — task-local founder amendment

On 2026-08-13 the founder explicitly instructed this Mail branch to use Terra instead of
Codex for implementation. For Phases 143–148 on `codex/mail-module-20260813`:

- requested model: `gpt-5.6-terra`
- reasoning effort: `xhigh`
- fallback: none
- root `gpt-5.6-sol` remains orchestrator/reviewer and does not write atomic product/test
  code

This is a branch/workstream-specific amendment, not a global rewrite of Hub routing. Every
dispatch records its exact requested model/effort and provenance source. An explicit native
dispatch accepted without rejection records `dispatcher-enforced`; actual telemetry, when
exposed, must match. A mismatch, rejection, omitted model/effort or fallback remains
`BLOCKED`. The instruction does not authorize protected-path work, direct Docker fallback,
migration apply, push, PR, merge, deployment, DNS, license purchase or secrets.
The mandatory first-read `AGENTS.md` now records this exact narrow exception so a
fail-closed dispatcher does not have to choose between contradictory routes.

## Superseded route at detection

The prior founder lock required every atomic code task to run as:

- requested model: `gpt-5.3-codex` (5.3 Codex Spark)
- reasoning effort: `xhigh`
- fallback: none

At detection, Root Sol and Terra were not permitted substitutes.

## Attempts and evidence

1. Native collaboration dispatch explicitly requested `gpt-5.3-codex` / `xhigh` for
   Task 143-01. The dispatcher rejected it before execution:
   `Unknown model gpt-5.3-codex`; available models were only `gpt-5.6-sol` and
   `gpt-5.6-terra`.
2. Local exact CLI route used `codex-cli 0.147.0-alpha.6.6`, explicit
   `-m gpt-5.3-codex` and `model_reasoning_effort="xhigh"` with dispatch artifact
   `DISPATCH-143-01.md`.
3. The sandboxed attempt could not reach the API. Per execution policy it was retried with
   approved network access.
4. The approved-network attempt reached the API and returned HTTP 400:
   `The 'gpt-5.3-codex' model is not supported when using Codex with a ChatGPT account.`

CLI thread ids for provenance: `019ff842-5c0f-7782-a8af-adc755392634` (network-blocked)
and `019ff843-2b21-72e0-b91f-0c66590c1cba` (catalog rejection).

## Historical result

- prior requested model: `UNAVAILABLE` (no inference or fallback)
- provider: OpenAI Responses API reached on the second attempt
- code changes from dispatch: none
- tests/build/lint/validators: NOT_RUN
- Task 143-01 code status at that attempt: `BLOCKED_BY_EXACT_RUNTIME`

## Resolution

The prior blocker could be resolved only when either:

1. the native dispatcher exposes exact `gpt-5.3-codex` with `xhigh`; or
2. the ChatGPT-account Codex catalog supports the exact locked slug; or
3. the founder explicitly amends the model-routing lock with another exact route.

Condition 3 was satisfied by the explicit founder directive above. Code dispatch is now
authorized only as exact `gpt-5.6-terra` / `xhigh`. Several authoring and review packets
subsequently reported exact provenance; their evidence is recorded in the matching
`IMPLEMENTATION-*.md` files. Docker-backed verification remains separately `NOT_READY`
until the canonical launcher is accepted.

## 2026-08-13 per-packet telemetry regression and founder resolution

Several earlier Terra packets reported requested and actual model/effort and were reviewed
under that evidence. Later 143-03C and 143-03F review runtimes did not expose actual
telemetry. A clean repeat 143-03C provenance review, the first 144-01B authoring attempt and
the exact 143-03A hash-pair remediation attempt also explicitly requested the required
Terra route, reported actual `UNAVAILABLE`, used no fallback and stopped before reading or
writing files.

Founder question 29 is now answered. For this branch and Phases 143–148, an explicit native
collaboration dispatch with the exact Terra route and effort that the dispatcher accepts and
runs without rejection is sufficient provenance, recorded as `dispatcher-enforced`. Actual
telemetry, if exposed, must match; a mismatch, rejection, omitted model/effort or fallback
remains `BLOCKED`.

Consequences:

- 143-03A/C remediation and 143-03C/F review may be dispatched under that exact rule;
- 143-03D remains blocked by completion of its hash-pair and independent-review dependencies;
- 144-01B remains subject to its independent contract/source-review gates, but not to a
  telemetry-only provenance block;
- an exact Terra/xhigh worker may author atomic production/test code only through an accepted
  native dispatch; root only orchestrates, reviews, stages and commits; and
- Docker availability does not make the separately missing launcher ready, and all executable
  evidence remains `NOT_RUN`.
