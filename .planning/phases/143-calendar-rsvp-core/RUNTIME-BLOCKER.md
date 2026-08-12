# Phase 143 code-runtime blocker

> **Detected:** 2026-08-13
> **Scope:** atomic production/test code only
> **Planning/review:** continues

## Required route

Founder lock requires every atomic code task to run as:

- requested model: `gpt-5.3-codex` (5.3 Codex Spark)
- reasoning effort: `xhigh`
- fallback: none

Root Sol and Terra are not permitted substitutes.

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

## Result

- actual model: `UNAVAILABLE` (no inference or fallback)
- provider: OpenAI Responses API reached on the second attempt
- code changes from dispatch: none
- tests/build/lint/validators: NOT_RUN
- Task 143-01 code status: `BLOCKED_BY_EXACT_RUNTIME`

## Resume condition

Resume code only when either:

1. the native dispatcher exposes exact `gpt-5.3-codex` with `xhigh`; or
2. the ChatGPT-account Codex catalog supports the exact locked slug; or
3. the founder explicitly amends the model-routing lock with another exact route.

Until then, continue bounded Terra analysis/review, documentation, dispatch preparation,
and other non-code work. Do not silently use Sol/Terra for implementation.
