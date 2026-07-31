# Layer 02 — GSD Max

Сжатый max-порт паттернов GSD, уже отфильтрованных Hub-инвентарём
(`.planning/archive/july-2026/ai-expansion/03-GSD-PORT-INVENTORY.md`).

## Обязательный цикл (P0)

```
discuss / clarify → plan (atomic) → execute waves → verify goal-backward → handoff
```

| Паттерн | Суть | Portable в Pro |
|---------|------|----------------|
| Goal-backward | Проверяй outcome фазы, не список галочек задач | `extracts/wave-protocol.md` §verify |
| Atomic work | Одна проверяемая единица → один коммит/артефакт | layer 04 |
| Waves | Параллель только без пересечения файлов + общий контракт | `extracts/wave-protocol.md` |
| Handoff/STATE | Следующий агент не видит чат — только артефакт | layer 06 |
| Discuss gray zones | Неясное — до plan, не во время code | layer 06 |
| Verify ≠ CI green | CI — необходимое, не достаточное; goal-verify отдельно | layer 05 |

## Где это живёт (не дублировать)

Всё перечисленное — в **Pro**: слой разработки переехал сюда из Standard
при разделении изданий по объекту работы.

- [`../playbooks/gsd-lifecycle.md`](../playbooks/gsd-lifecycle.md)
- [`../playbooks/wave-execution.md`](../playbooks/wave-execution.md)
- [`../playbooks/handoff-state.md`](../playbooks/handoff-state.md)
- [`../playbooks/discuss-and-research.md`](../playbooks/discuss-and-research.md) — этап до плана
- skills: [`phase-plan`](../skills/phase-plan/SKILL.md), [`wave-orchestrate`](../skills/wave-orchestrate/SKILL.md), [`verify-goal`](../skills/verify-goal/SKILL.md)

В Standard остаётся то, что относится к тестовому фреймворку: context,
rules, recipes, evals, validator.

## Hub-only усиления (в Pro — как опция)

| Hub skill | Когда звать в Hub | Portable аналог |
|-----------|-------------------|-----------------|
| `.cursor/skills/gsd-ru/phase-plan` | План фазы с TZ | Standard `skills/phase-plan` |
| `gsd-ru/wave-run` | Волна с rule 56 | `extracts/wave-protocol.md` |
| `gsd-ru/goal-verify` | Приёмка фазы | Standard `skills/verify-goal` |
| `gsd-ru/handoff` | Конец сессии | layer 06 |
| `gsd-ru/portable-kit` | Перенос kit | Standard `PORTABLE-HUB-KIT.md` |

## SKIP из upstream GSD (не тащим)

Полный autonomous режим, Discord, workspace meta, vendor-copy skill bodies,
автоскан закрытых MR.

## DoD слоя

Оркестратор умеет описать фазу goal-backward и разрезать на волны без пересечения
scope, сославшись на extracts, а не на «как в чате обсудили».
