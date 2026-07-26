---
fixture_id: EVAL-DBG-001
category: debug-flaky
status: template
owner: automation-quality
oracle_status: TODO
---

# Debug and flaky fixture

## Goal

Оценить evidence-led triage падающего или intermittently failing теста, включая
разделение причин product-code, test-code, environment и nondeterminism.

## Instantiate before use

- Failure log без sensitive values: `{{failure_log}}`
- Test и релевантные production paths: `{{allowed_paths}}`
- Requirement source: `{{requirement_source}}`
- Reproduction command: `{{reproduction_command}}`
- Repetition command или approved repeat method: `{{repeat_command}}`
- Known root cause у evaluator: `{{gold_root_cause}}`
- Explicitly forbidden shortcuts: `{{forbidden_shortcuts}}`

## Prompt given to both arms

Diagnose the failure using a hypothesis-and-experiment sequence. First establish the expected behavior from the approved source, then classify the root as product code, test code, infrastructure, flaky behavior, or insufficient evidence. Propose the minimum discriminating experiments before a fix. Do not mask the symptom by increasing a timeout, adding retries, or inserting a fixed delay. Return evidence, root classification, confidence, smallest justified correction, and regression scope.

Assisted arm дополнительно получает versioned toolkit material, определённый
eval manifest.

## Oracle

- Expected behavior опирается на requirement, а не выводится из текущего
  output.
- Initial failure воспроизводим или явно помечен как not reproduced.
- Competing hypotheses различимы предложенными experiments.
- Root classification соответствует evidence и по умолчанию не винит тест.
- Correction устраняет cause и сохраняет diagnostic signal.
- Fixed delay, blanket retry, quarantine или timeout inflation отклоняются,
  пока approved contract сам этого не требует.
- Focused и regression commands берутся из supplied context.

## Seeded mutations

1. Сделай так, чтобы порядок тестов влиял на shared state.
2. Сдвинь событие чуть за пределы fixed polling deadline.
3. Дай misleading log line, temporally adjacent, но не causal.
4. Измени requirement так, чтобы старое assertion стало genuinely obsolete.

## Scoring emphasis

Root-cause accuracy, experiment quality, flaky-risk reduction, defects, false
positives, hallucinated commands, traceability, first-pass acceptance, review
minutes, tool calls, tokens и latency.
