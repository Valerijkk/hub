# Review guidelines

Эти вопросы не являются Boolean gates. Reviewer оценивает change model, behavioral contract и failure mode.

## Functions and modules

- Какое единое обещание выполняет unit?
- Смешаны ли policy, I/O, authorization, persistence и notification с разными failure semantics?
- Создаст ли extraction meaningful name/boundary или только навигационные прыжки?
- Скрывает ли module реально изменчивое решение?

`>25 logical lines` — повод задать вопросы, а не требование дробить cohesive linear algorithm. Evidence: `EMP`, `REC`; [S02, S12–S14](../sources/registry.md).

## SOLID as questions

- `SRP`: какие change reasons действительно независимы?
- `OCP`: variation доказана или speculative?
- `LSP`: сохраняет ли subtype observable preconditions, postconditions, effects и lifetime?
- `ISP`: какой capability реально нужен consumer?
- `DIP`: captive ли stable policy у volatile detail?

BAD: interface/class quota. GOOD: smallest boundary around demonstrated variation. Evidence: `NORM`, `REC`, `DISPUTED`; [S02, S03](../sources/registry.md).

## DRY, abstraction, YAGNI

- Дублируется одно knowledge, обязанное меняться вместе, или лишь похожий syntax?
- Не накопились ли flags, потому что abstraction связывает разные concepts?
- Security/compatibility/failure recovery — известное требование или speculative feature?

BAD: объединить login email и invoice display email из-за одинаковых двух строк. GOOD: сохранить разные domain names, пока инварианты не доказаны общими. Evidence: `REC`; [S15, S16](../sources/registry.md).

## Composition, inheritance, DI, lifetimes

- Есть ли стабильный behavioral subtype? Если нет, composition обычно яснее.
- Видны ли mutable/external dependencies и lifetime?
- Нужен ли interface consumer-у или он создан ради SOLID score?
- Процесс-wide immutable value допустим; mutable globally reachable state требует особого rationale.

Evidence: `NORM`, `REC`; [S03, S20, S21](../sources/registry.md).

## Tests

- Проверяется observable outcome или private call order?
- Assertions выражают один behavioral concept?
- Помогает ли fixture/factory/POM читать scenario или скрывает его?
- Выбран ли scope по риску и feedback, а не по требуемому проценту pyramid?

Evidence: `REC`, `DISPUTED`; [S23, S25, S26, S28, S31, S32](../sources/registry.md).

## Rationale outcome

Review завершается одним из решений:

- оставить простой design;
- выделить boundary с названным change/failure reason;
- добавить targeted test/control;
- оформить временное [исключение](exceptions-and-rationale.md);
- собрать telemetry, прежде чем вводить новый threshold.
