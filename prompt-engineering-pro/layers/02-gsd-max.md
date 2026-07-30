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

## Уже в Standard (не дублировать)

- `prompt-engineering/playbooks/gsd-lifecycle.md`
- `prompt-engineering/playbooks/wave-execution.md`
- `prompt-engineering/playbooks/handoff-state.md`
- skills: `phase-plan`, `wave-orchestrate`, `verify-goal`

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
