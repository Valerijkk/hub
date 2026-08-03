# Skills — Touchstone Pro

Три скила жизненного цикла фазы. Скилы работы с тестовым фреймворком
(bootstrap/refresh context, test-verification, naming-check, anti-ai-check,
curate-*) остались в Standard.

| Skill | Что делает | Когда |
|-------|------------|-------|
| [phase-plan](phase-plan/SKILL.md) | Атомарные планы фазы (порт GSD) | Перед волной |
| [wave-orchestrate](wave-orchestrate/SKILL.md) | Оркестрация волны агентов | Во время execute |
| [verify-goal](verify-goal/SKILL.md) | Goal-backward приёмка фазы | После волны/фазы |

Порядок фиксированный: план → волна → приёмка. `verify-goal` проверяет
достижение цели фазы, а не факт закрытия задач — список галочек и
достигнутый outcome это разные вещи.

Протокол верификации, на который опираются все три, живёт в Standard:
[`rules/40-verification-protocol.md`](../../pro-qa-ai/rules/40-verification-protocol.md).
Усиление Pro поверх него — пирамида L0–L4 и мутационный контракт:
[`../layers/05-verification-mutation.md`](../layers/05-verification-mutation.md).
