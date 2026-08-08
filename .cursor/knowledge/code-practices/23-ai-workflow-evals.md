# 23 — AI workflow evals

## Goal

Adopt workflow changes only when a controlled Hub-local comparison shows lower
cost of safe change, fewer regressions or lower false confidence than a simpler
baseline. Upstream demos and a framework's internal benchmark are hypotheses,
not Hub evidence.

## Frozen cohort

Before running any arm, freeze:

- task IDs and risk strata: trivial, multi-file, ambiguous, security/irreversible;
- starting repository revision and environment image digest;
- visible task contract and hidden executable oracle;
- requested and actual model/effort (`UNKNOWN` when not observable), launcher,
  tool/workflow/skill revisions and capability profile;
- wall-clock, token/cost and human-review accounting rules.

Keep the hidden oracle unavailable to the agent and workflow author. Randomize or
counterbalance arm order where carry-over is possible, and retain failed runs.

## Minimal paired experiments

Each experiment changes one decision while holding the cohort and runtime
configuration constant:

| Question | Baseline | Candidate |
|---|---|---|
| Does a structured contract help? | current concise task | TaskContract with outcome/non-goals/ownership/evidence/stop fields |
| Does a framework help? | plain current workflow | smallest selected framework configuration |
| Does delegation help? | single agent | bounded multi-agent only on independent task strata |
| Does context reset help? | continuous context + explicit state | fresh context + identical explicit handoff |
| Does verification help? | generator + executable checks | conclusion-blind verifier + same checks |

Do not combine all candidates into one arm; an aggregate win would not identify
which added ceremony caused the result.

## Metrics

Primary outcome metrics:

- hidden-oracle pass rate and regression escape rate;
- valid defects found before handoff and valid defect yield per unit cost;
- false-confidence rate = runs claiming `complete/verified` while the hidden
  oracle fails;
- recovery success for injected failure and rollback scenarios.

Efficiency and context metrics:

- model/tool cost, tokens, latency, wall-clock and human-review minutes;
- context precision = loaded items actually needed / all loaded items;
- context recall = needed items loaded / all needed items identified after the run;
- coordination failures, duplicate work, ownership conflicts and retry/remediation turns.

Process metrics diagnose failures but do not override outcome. A ritual-compliant
run that passes by luck remains unsafe; a process violation is not automatically
a product defect unless its stated risk materializes or policy requires it.

## Decision record

Pre-register the acceptance rule per experiment; this catalog does not invent a
universal numeric threshold. Record effect by task stratum, uncertainty, cost
change, failures, and the smallest configuration that preserves the gain.
Promote a candidate control only through the normal admission process. Retest
after model, launcher, framework or capability-policy changes.

Evidence and limitations: [prompt/agent workflow snapshot](sources/prompt-agent-workflows-2026-08-08.md).

