---
fixture_id: EVAL-REV-001
category: review
status: template
owner: automation-quality
oracle_status: TODO
---

# Review fixture

## Goal

Оценить качество defect-first review, дисциплину evidence, калибровку severity
и контроль false-positive на изменении автоматизации.

## Instantiate before use

- Frozen diff или patch: `{{review_diff}}`
- Источники requirement и test-case: `{{requirements_and_test_cases}}`
- Источники framework и naming: `{{framework_sources}}`
- Allowed review scope: `{{review_scope}}`
- Known valid findings у evaluator: `{{gold_findings}}`
- Known tempting non-issues: `{{negative_controls}}`

## Prompt given to both arms

Review the supplied change for defects that can alter test meaning, hide product regressions, introduce flaky behavior, break isolation, violate framework conventions, or escape scope. Return only actionable findings with severity, exact location, evidence, impact, and a concrete correction. Separate open questions from defects. If no actionable finding exists, say so.

Assisted arm дополнительно получает versioned toolkit material, определённый
eval manifest.

## Oracle

- Каждый accepted finding идентифицирует реальный defect в frozen diff.
- Severity следует impact и likelihood, а не интенсивности формулировок.
- Style-only preferences не повышаются до correctness defects.
- Reviewer проверяет false-green assertions, missing waits, shared mutable
  state, cleanup, test data collisions и requirement drift, когда применимо.
- Никакой file вне supplied review scope не трактуется как changed.
- Suggested commands и paths существуют в frozen context.
- Отчёт не содержит praise padding перед material findings.

## Seeded mutations

1. Замени assertion на всегда проходящее.
2. Введи fixed-delay synchronization вокруг asynchronous condition.
3. Добавь harmless formatting change как negative control.
4. Измени test name, сохраняя documented naming convention, как второй
   negative control.

## Scoring emphasis

Defect recall by severity, false positives, traceability to diff и
requirements, first-pass acceptance, review minutes, hallucinated paths или
commands, scope escapes, tool calls, tokens и latency.
