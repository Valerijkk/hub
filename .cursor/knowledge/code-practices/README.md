# Hub code practices

Компактная база решений о качестве кода. Её цель — снижать стоимость безопасного изменения: делать контракт локально понятным, ограничивать радиус изменения и быстро получать надёжную обратную связь.

Это не новый линтер и не норматив «чистоты». Здесь всегда разделены:

1. принцип;
2. класс evidence;
3. hard/hybrid control либо warning/review heuristic;
4. BAD/GOOD пример;
5. источник и его статус.

## Как применять

- Начните с [принципов](00-principles.md) и [модели evidence](01-evidence-model.md).
- Машинный блок допустим только для точно определённого класса ошибок и с явным механизмом проверки.
- Метрики размера, сложности, duplication и количества mocks — только сигналы для review, пока Hub не накопил локальную калибровку.
- Любое исключение оформляется по [шаблону rationale](controls/exceptions-and-rationale.md), а не обходом проверки.
- Статус `candidate` означает рекомендацию к внедрению, а не утверждение, что control уже реализован в CI Hub.

## Навигация

### Основа

- [00 — Принципы](00-principles.md)
- [01 — Evidence model](01-evidence-model.md)
- [07 — Ошибки и resilience](07-errors-and-resilience.md)
- [15 — Test automation](15-test-automation.md)
- [18 — Security](18-security.md)
- [21 — AI-agent engineering](21-ai-agent-engineering.md)
- [22 — AI workflow decisioning](22-ai-workflow-decisioning.md)
- [23 — AI workflow evals](23-ai-workflow-evals.md)

### Controls

- [Hard и hybrid controls C01–C53](controls/hard-controls.md)
- [Review guidelines](controls/review-guidelines.md)
- [Warning signals](controls/warning-signals.md)
- [Negative sentinels NEG-01–NEG-27](controls/negative-sentinels.md)
- [Exceptions and rationale](controls/exceptions-and-rationale.md)

### Examples

- [Python](examples/python.md)
- [Go](examples/go.md)
- [TypeScript / React](examples/typescript-react.md)
- [Tests](examples/tests.md)

### Decisions

- [Decision matrix](decisions/decision-matrix.md)
- [Contradictions](decisions/contradictions.md)
- [Cargo-cults](decisions/cargo-cults.md)

### Sources

- [Frozen source registry](sources/registry.md)
- [Evidence and integrity status](sources/evidence-status.md)
- [Prompt/agent workflow snapshot — 2026-08-08](sources/prompt-agent-workflows-2026-08-08.md)

## Не превращать в gate

`≤25 lines`, SOLID, DRY, Rule of Three, POM, mocks, Pyramid/Trophy/Honeycomb, число assertions, 100% coverage, mutation-score threshold, zero duplication, «никаких comments/inheritance/Singleton» и фиксированное значение Cognitive Complexity не являются Boolean-определениями качества.

## Формат новой карточки

```markdown
## <Название>

- Принцип: <какой риск или стоимость изменения уменьшается>
- Evidence: `<CONTROL|NORM|EMP|REC|OPINION|POLICY|DISPUTED>`; <S-ID>; confidence <уровень>
- Режим: `<HARD|HYBRID|WARNING|REVIEW>`; <конкретный механизм или вопрос>
- BAD: <нарушенный контракт/failure mode>
- GOOD: <минимальный sufficient design>
- Исключение: <условия и evidence>
- Источник: [S-ID](sources/registry.md#<anchor>)
```

Новая карточка не может ссылаться только на имя школы или паттерна. Она должна назвать failure mode и область применимости.
