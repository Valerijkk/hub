# AI Expansion — mid-August 2026

> **Горизонт:** 22.07.2026 → **15.08.2026**
> **Суть:** портировать лучшее из GSD (Get Shit Done) на русский (практичнее оригинала), расширить `.cursor/agents/` + `pro-qa-ai/`, собрать **переносимый kit** для bootstrap корпоративного хаба (урезанный GitHub-аналог + микросервисы + AI-команда).
> **Код продукта:** не трогаем (July freeze). Только plans / agents / prompts / skills.
>
> **Продуктовая упаковка (2026-07-30):** линейка **Touchstone** —
> Standard = [`pro-qa-ai/TOUCHSTONE.md`](../../../../pro-qa-ai/TOUCHSTONE.md) ·
> Pro = [`prompt-engineering-pro/`](../../../../prompt-engineering-pro/README.md).
> Этот каталог — исторический план и источники решений; поставка — соседние папки репо.

## Текущий фокус (2026-07-30)

**Touchstone Standard** (`pro-qa-ai/`) — portable **AI quality layer для test automation**:

| Принцип | Описание |
|---------|----------|
| Context bootstrap | Явные roots → `context/projects/<profile>/`; без project-хардкода в core |
| Human-in-the-loop | После bootstrap человек добавляет good/bad code examples + комментарии коллег |
| Без MR-scanning | **Нет** автоматического сканирования закрытых MR/PR |
| Eval | Harness implemented; effectiveness/ROI **NOT_RUN** |
| Pro overlay | `prompt-engineering-pro/` — GSD max + Cursor contract + mutation + planning |

Repo-scan provenance: [`.planning/audits/2026-07/repo-scan/PROVENANCE.md`](../../../audits/2026-07/repo-scan/PROVENANCE.md) — Must PASS здесь **qualified**, не erase historical pass.

## Индекс артефактов

| # | Файл | Владелец | Статус |
|---|------|----------|--------|
| 01 | [01-PRODUCT-SCOPE.md](01-PRODUCT-SCOPE.md) | product | ✅ |
| 02 | [02-TECH-PLAN.md](02-TECH-PLAN.md) | tech-lead | ✅ (updated intent) |
| 03 | [03-GSD-PORT-INVENTORY.md](03-GSD-PORT-INVENTORY.md) | research | ✅ 79 · P0=21 / P1=35 |
| 04 | [04-PORTABLE-HUB-ARCH.md](04-PORTABLE-HUB-ARCH.md) | arch | ✅ |
| 05 | [05-PORTABLE-KIT.md](05-PORTABLE-KIT.md) | docs | ✅ |
| 06 | [06-REDUCED-GITHUB-MODULE.md](06-REDUCED-GITHUB-MODULE.md) | product+arch | ✅ |
| 07 | [07-ACCEPTANCE.md](07-ACCEPTANCE.md) | product-ver | **QUALIFIED PASS** · structural PASS, eval pending |
| 08 | [08-PORTABLE-KIT-DRY-RUN.md](08-PORTABLE-KIT-DRY-RUN.md) | docs | ✅ dry-run (docs) |
| 09 | [09-AI-QUALITY-LAYER-REWORK.md](09-AI-QUALITY-LAYER-REWORK.md) | team-lead | ✅ implementation contract |

## Волна maximal expand (2026-07-23)

Углубление после Must PASS (параллельные агенты + fallback оркестратора):

| Результат | Факт на диске |
|-----------|----------------|
| Corporate prompts | `prompts/13-corporate-delivery.md` **#1–112** |
| PE corporate + QA/review/debug | +`qa-manual`, `qa-auto`, `code-reviewer`, `debugger` (рядом с orchestrator…devops) |
| Dry-run portable kit | [`08-PORTABLE-KIT-DRY-RUN.md`](08-PORTABLE-KIT-DRY-RUN.md) |
| P0/P1 маршрутизация | [`.cursor/GSD-INTEGRATION.md`](../../../../.cursor/GSD-INTEGRATION.md) § **P0/P1 port status mid-Aug** |
| GSD-RU | **6** skills (ядро P0: plan/wave/verify/handoff + portable/strangler) |
| Оставшийся P0 | **15/21** — agents/practice/route (не отдельный gsd-ru skill) → polish до 15.08 |

Приёмка: [`07-ACCEPTANCE.md`](07-ACCEPTANCE.md) — Must **QUALIFIED PASS** (2026-07-23 historical PASS сохранён); maximal expand **PARTIAL**; structural gaps → remediation.

## Связанные пути в репо

| Что | Где |
|-----|-----|
| GSD-RU skills (Hub-lite) | `.cursor/skills/gsd-ru/` |
| Upstream GSD маршрутизация + P0/P1 mid-Aug | `.cursor/GSD-INTEGRATION.md` |
| Prompt Engineering toolkit | `pro-qa-ai/` |
| PE → GSD port index | `prompt-engineering-pro/gsd-port/README.md` |
| Portable kit user guide | `prompt-engineering-pro/PORTABLE-HUB-KIT.md` |
| Agent prompts (32) | `.cursor/agents/*.md` — секция **GSD-паттерны (Hub-усиленные)** |
| Контракт промпта | `.cursor/playbooks/prompt-contract.md` |
| Microservices ADR | `.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md` |
| Dynamic model routing (applied 26.07) | `.planning/decisions/ADR-MODELS-GPT56-2026.md` |
| Repo-scan provenance | `.planning/audits/2026-07/repo-scan/PROVENANCE.md` |

## Быстрый старт (перенос в другой репо)

См. [`prompt-engineering-pro/PORTABLE-HUB-KIT.md`](../../../../prompt-engineering-pro/PORTABLE-HUB-KIT.md), [`05-PORTABLE-KIT.md`](05-PORTABLE-KIT.md) и dry-run [`08-PORTABLE-KIT-DRY-RUN.md`](08-PORTABLE-KIT-DRY-RUN.md).

Bootstrap: создать `context/projects/<profile>/` из явных roots → человек при
необходимости вручную добавляет примеры кода и review-комментарии → **никакого**
closed-MR mining.

## Принципы порта GSD

1. **Паттерны, не копипаста upstream** — извлекаем lifecycle, wave, handoff, goal-backward, atomic plans.
2. **Русский + ворота Hub** — rule 56 (цикл качество+регрессия), prompt-contract, шаблон team-lead §5.2.
3. **Обезличенность PE** — project facts только в generated profile.
4. **Переносимость** — один kit → любой корпоративный monorepo.
5. **Concise agents** — GSD-секции: progressive disclosure, не line-count inflation (см. `02-TECH-PLAN.md`).

*Создано: 2026-07-22 · обновлено: 2026-07-26 (provenance + PE test-automation intent + qualified acceptance).*
