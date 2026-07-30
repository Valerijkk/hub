# Touchstone Pro — Prompt Engineering Pro

> **Продукт:** Touchstone · **Издание:** Pro (`0.2.0-rc.1`)  
> **База:** [Touchstone Standard](../prompt-engineering/TOUCHSTONE.md)  
> **Первый вход:** [START-HERE.md](START-HERE.md)

## Зачем Pro

Standard закрывает слой «AI для автоматизатора вокруг уже существующего
тестового фреймворка». Pro добавляет **всё лучшее**, что Hub уже доказал на
практике в GSD, Cursor-команде и planning — в **project-agnostic** виде, чтобы
лечь в основу Touchstone для любых тестовых фреймворков.

| Источник | Что берём в Pro |
|----------|-----------------|
| **GSD** | goal-backward, atomic commits, waves, handoff, discuss→plan→execute→verify, Nyquist/verify patterns |
| **`prompt-engineering/`** | Standard целиком: context, agents, recipes, evals, packager |
| **`.cursor/`** | prompt-contract, team orchestration, verification/evidence, regression cycle, anti-AI |
| **`.planning/`** | backlog task contract, mutation evidence, phase status honesty, STATE handoff |

## Принцип упаковки

1. **Standard не форкается** — Pro надстраивается и ссылается на `../prompt-engineering/`.
2. **Паттерны, не vendor-copy** — не копируем upstream GSD skill-тексты целиком;
   фиксируем контракты и чеклисты (как в `03-GSD-PORT-INVENTORY`).
3. **Hub-only TZ наружу не утекает** — Fiber/Ent/JWT инварианты остаются в Hub;
   в Pro — обобщённые security/verification практики.
4. **Evidence > галочки** — mutation, layer match, independent verifier.
5. **Честный status** — structural foundation-ready; effectiveness до GA = `NOT_RUN`.

## Слои Pro

| # | Слой | Файл |
|---|------|------|
| 0 | Отношение Standard ↔ Pro | [RELATIONSHIP.md](RELATIONSHIP.md) |
| 1 | Core Quality Layer (= Standard) | [layers/01-core-quality-layer.md](layers/01-core-quality-layer.md) |
| 2 | GSD Max | [layers/02-gsd-max.md](layers/02-gsd-max.md) |
| 3 | Prompt Contract | [layers/03-prompt-contract.md](layers/03-prompt-contract.md) |
| 4 | Team Orchestration | [layers/04-team-orchestration.md](layers/04-team-orchestration.md) |
| 5 | Verification & Mutation | [layers/05-verification-mutation.md](layers/05-verification-mutation.md) |
| 6 | Planning Ops | [layers/06-planning-ops.md](layers/06-planning-ops.md) |
| 7 | Cursor / runtime adapters | [layers/07-cursor-adapters.md](layers/07-cursor-adapters.md) |

Практические extracts (готовые к переносу тексты): [`extracts/`](extracts/).  
Индекс источников: [SOURCES.md](SOURCES.md) · манифест: [MANIFEST.md](MANIFEST.md).

## Кому какое издание

| Сценарий | Издание |
|----------|---------|
| Подключить AI-слой к чужому testframework | **Standard** |
| Корпоративный delivery + волны агентов + planning discipline | **Pro** |
| Только рецепты генерации API/E2E тестов | Standard + `recipes/` |
| Полный цикл как в Hub (plan→wave→verify→mutation) | **Pro** |

## Статус

`foundation-ready` · зависит от Standard `0.2.0-rc.1` · paired eval Pro-пакета
отдельно не гонялся (`NOT_RUN`). См. [VERSION.md](VERSION.md).
