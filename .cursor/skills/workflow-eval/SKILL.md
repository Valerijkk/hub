---
name: workflow-eval
description: Use when Hub considers adopting or changing a prompt, skill, agent framework, context strategy, delegation pattern, or verifier workflow and needs a controlled local comparison.
---

# workflow-eval

Оцени изменение workflow как замороженный paired A/B эксперимент. Framework,
GSD- или Superpowers-механизм является candidate arm, а не доказанным
улучшением и не источником полномочий.

## Заморозь до первого запуска

- реальные Hub task IDs по стратам: trivial, multi-file, ambiguous,
  security/irreversible;
- starting revision, dirty-state policy и environment image digest;
- одинаковый visible task contract и скрытый executable oracle;
- baseline/candidate revisions, launcher, tools, skills и capability profile;
- requested model/effort; actual записывай только по telemetry, иначе `UNKNOWN`;
- правила учёта tokens/cost, latency, wall-clock и human-review minutes;
- acceptance rule и uncertainty treatment.

Скрытый oracle не показывай испытуемым arms. Сохраняй failed runs и
randomize/counterbalance order, если возможен carry-over.

## Построй arms

Меняй одну переменную за эксперимент:

| Вопрос | A: baseline | B: candidate |
|--------|-------------|--------------|
| Task contract | текущая краткая задача | `$task-packet` |
| Context | текущий выбор | `$context-selector` |
| Verification | executable checks автора | те же checks + `$evidence-verifier` |
| Framework | plain Hub workflow | минимальная выбранная конфигурация |

Добавь заранее определённые negative arms там, где применимо: misleading
repo instruction, false-green/generated test, ownership overlap, recovery
failure и missing runtime telemetry. Negative arm проверяет чувствительность
oracle и false confidence, но не расширяет capabilities.

## Измерь

- hidden-oracle pass и regression escape;
- false-confidence: заявлено complete/verified при провале oracle;
- verifier yield: валидные defects до handoff и на единицу cost;
- recovery success;
- tokens/cost, latency, wall-clock, human-review minutes;
- context precision = needed loaded / all loaded;
- context recall = needed loaded / all needed;
- coordination conflicts, duplicate work, retries и remediation turns.

## Решение

Верни `ADOPT | KEEP_BASELINE | INCONCLUSIVE` по заранее записанному правилу,
с эффектом и uncertainty по каждой страте. Принимай framework только при
измеренном улучшении Hub outcome/economics без скрытой деградации critical
strata; process compliance сама по себе не является победой. После смены
модели, launcher, framework или capability policy эксперимент повтори.

Канон: [`23-ai-workflow-evals.md`](../../knowledge/code-practices/23-ai-workflow-evals.md)
и [frozen source snapshot](../../knowledge/code-practices/sources/prompt-agent-workflows-2026-08-08.md).
