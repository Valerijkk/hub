# Layer 05 — Verification & Mutation

Источники: Standard `rules/40`, `skills/test-verification`, `skills/verify-goal`,
`.cursor/rules/56-regression-verify.mdc`, `.cursor/playbooks/verification.md`,
практика Stage C mutation (Hub audits).

## Пирамида доказательств

| Уровень | Что считается evidence |
|---------|------------------------|
| L0 Claim | Текст агента без команды — **не evidence** |
| L1 Command | Команда + exit code / ключевой stdout |
| L2 Artifact | Путь к логу, отчёту, diff |
| L3 Mutation | Сломали prod-guard → тест RED → restore → GREEN |
| L4 Independent | Verifier в отдельном контексте (critical/security) |

## Mutation contract (Touchstone)

Для фикса/закрытия задачи с тестом обязательны три поля:

```
что ломать      файл + фрагмент / условие
какой тест      имя теста, обязан упасть
слой            handler | service | repository | schema | ui | other
```

Вердикты:

| Вердикт | Значение |
|---------|----------|
| `PROVEN` | RED на мутации + GREEN после restore |
| `mutation_failed_to_build` | Мутация сломала сборку — переформулировать, не засчитывать |
| `layer_mismatch` | Тест бьёт не в тот слой, где guard |
| `NOT_PROVEN` | Нет RED / нет спецификации мутации |

Extract: [`../extracts/mutation-evidence.md`](../extracts/mutation-evidence.md).

## Регрессионный цикл

Новое поведение: test → (fix) → re-test до green.  
Старое: тот же цикл на затронутом контуре.  
Без green — не baseline.

## DoD слоя

Отчёт исполнителя без L1+ не принимается.

**Для security, доступа, денег и разрушающих операций — L3 и L4 вместе**,
а не на выбор. Мутация доказывает, что проверка способна упасть;
независимый верификатор доказывает, что её оценивал не автор. Одно без
другого оставляет дыру: доказанный тест, который никто не перечитал, либо
перечитанный тест, который не умеет краснеть.

Для остального critical — L3 либо L4 по классу риска, таблица соответствия
в [`../skills/verify-goal/SKILL.md`](../skills/verify-goal/SKILL.md) §Доказательства.
