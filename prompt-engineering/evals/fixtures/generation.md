---
fixture_id: EVAL-GEN-001
category: generation
status: template
owner: automation-quality
oracle_status: TODO
---

# Generation fixture

## Goal

Оценить, может ли модель превратить approved requirement в runnable
автоматизированный тест без выдумывания behavior, commands, paths или
framework conventions.

## Instantiate before use

- Requirement IDs и точные source excerpts: `{{requirement_ids_and_quotes}}`
- Accepted test-case IDs: `{{test_case_ids}}`
- Frozen repository revision: `{{repository_revision}}`
- Canonical test path: `{{allowed_test_path}}`
- Canonical focused command: `{{focused_test_command}}`
- Canonical regression command: `{{regression_command}}`
- Existing pattern file: `{{pattern_source_path}}`
- Explicit out of scope: `{{out_of_scope}}`

## Prompt given to both arms

Create the smallest automated test artifact that proves the accepted test case. Change only the allowed test path. Use only supplied commands and repository patterns. If the requirement does not determine an expected result, return an open question instead of choosing one. Report changed paths, requirement-to-assertion traceability, focused evidence, regression evidence, and residual flaky risks.

Assisted arm дополнительно получает versioned toolkit material, определённый
eval manifest.

## Oracle

- Артефакт компилируется или собирается canonical runner.
- Каждое assertion отображается на accepted requirement или test-case ID.
- Не создаётся production file или framework scaffold.
- Не цитируется path или command, отсутствующие в frozen context.
- Negative и boundary behavior включаются только когда указаны или явно
  запрошены.
- Time, randomness, shared state, network и cleanup детерминированы или
  контролируемы.
- Focused и regression commands в точности совпадают с approved commands.

## Seeded mutations

1. Инвертируй один specified expected result в disposable mutant.
2. Замени canonical test path в decoy note на правдоподобный, но
   несуществующий path.
3. Убери cleanup из pattern example, когда cleanup требуется fixture.

Запиши, сопротивляется ли output decoy и обнаруживает ли его тест behavioral
mutant.

## Scoring emphasis

First-pass acceptance, traceability, mutation sensitivity, hallucinated paths
или commands, scope escapes, defects, flaky risk, review minutes, tokens, tool
calls и latency.
