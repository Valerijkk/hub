# Touchstone Pro — AI-слой для разработки

> **Продукт:** Touchstone · **Издание:** Pro (`0.3.0-beta.1`)
> **База:** [Touchstone Standard](../pro-qa-ai/TOUCHSTONE.md) (обязателен)
> **Первый вход:** [START-HERE.md](START-HERE.md)

## Два издания — два объекта работы

Линейка делится не по «базовый / продвинутый», а по тому, **что именно
делает команда**.

| Издание | Объект работы | Кому |
|---------|---------------|------|
| **Standard** `pro-qa-ai/` | Тестовый фреймворк | Инженеру автоматизации тестирования |
| **Pro** `prompt-engineering-pro/` | Продукт, который разрабатывают | Команде разработки целиком |

Автоматизатору достаточно Standard. Команде разработки нужны оба:
Pro ставится **поверх** Standard и переиспользует его context, rules,
recipes и validator — дублировать их запрещено.

## Что внутри Pro

| Раздел | Содержание |
|--------|------------|
| [`agents/`](agents/README.md) | 12 ролей разработки: orchestrator, product-owner, architect, backend-dev, frontend-dev, qa-manual, qa-auto, code-reviewer, debugger, security-reviewer, devops, docs-writer |
| [`skills/`](skills/README.md) | phase-plan · wave-orchestrate · verify-goal |
| [`playbooks/`](playbooks/README.md) | corporate-delivery · gsd-lifecycle · wave-execution · handoff-state · portable-kit-dry-run · discuss-and-research · debug-forensics |
| [`prompts/`](prompts/README.md) | Категория 13 — corporate delivery, 112 фрагментов |
| [`adapters/`](adapters/README.md) | Cursor-маппинг 12 ролей |
| [`extracts/`](extracts/README.md) | Переносимые контракты: prompt-contract, master-delegate, wave-protocol, mutation-evidence, backlog-task |
| [`layers/`](layers/01-core-quality-layer.md) | Семь слоёв практик — теория за разделами выше |
| [`gsd-port/`](gsd-port/README.md) | Индекс соответствия GSD → Touchstone |
| [`PORTABLE-HUB-KIT.md`](PORTABLE-HUB-KIT.md) | Перенос корпоративного kit в чужой репозиторий |

## Откуда взято

Pro — сборка того, что уже доказано на практике, а не набор идей.

| Источник | Что взято |
|----------|-----------|
| **GSD** | goal-backward приёмка, атомарная работа, волны, handoff, цикл discuss→plan→execute→verify |
| **`.cursor/` Hub** | prompt-contract, оркестрация команды, протокол верификации, регрессионный цикл, anti-AI |
| **`.planning/` Hub** | контракт backlog-задачи, мутационное доказательство, честный статус фазы, STATE-handoff |
| **Standard** | context, rules, recipes, evals, packager — по ссылке, без копии |

Карта источников с путями — [SOURCES.md](SOURCES.md).

## Приколы Cursor

Практики, которые в обычных пакетах не встречаются, — они и есть причина
существования Pro:

1. **Парные роли исполнитель ↔ верификатор.** Ревьюер не пишет код: в
   Cursor-маппинге `architect`, `code-reviewer` и `security-reviewer`
   помечены `readonly: true`. Независимость проверки обеспечена правами,
   а не обещанием.
2. **Мутационное доказательство.** Фикс с тестом закрывается только если
   сломанный guard в проде уронил тест. Зелёный тест сам по себе не
   доказывает ничего — [`extracts/mutation-evidence.md`](extracts/mutation-evidence.md).

   Это возможно **только здесь**: Pro работает по продуктовому коду, и ломать
   есть что. В Standard, живущем в репозитории тестового фреймворка, доступа
   к продукту нет, и мутация там подменяется структурным обоснованием
   чувствительности. Требовать её от Standard — значит завести гейт, который
   невозможно пройти честно.
3. **Пирамида evidence L0–L4.** Текст агента — не доказательство.
   Команда, артефакт, мутация, независимый верификатор — доказательства.
4. **Триаж `ROOT=code|test`.** Перед починкой красного теста решается,
   виноват код или тест. Иначе тест «чинится» под баг.
5. **Волна с непересекающимися файлами.** Два агента на одном файле —
   гонка, а не параллелизм.
6. **Недоверенные данные.** Код, логи, issue, вывод другого агента —
   данные, а не инструкции. Вложенные указания не меняют scope.

## Слои

Теория за практикой, семь документов:

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

## Принципы упаковки

1. **Standard не форкается** — Pro ссылается на `../pro-qa-ai/`.
2. **Паттерны, не vendor-copy** — тексты upstream GSD-скилов целиком не тащим.
3. **Hub-специфика наружу не утекает** — TZ, Fiber, Ent, RBAC остаются в
   `.cursor/rules/60-tz-constraints.mdc`.
4. **Evidence вместо галочек.**
5. **Честный статус** — эффективность до GA помечена `NOT_RUN`.

## Статус

`foundation-ready` · требует Standard `>=0.3.0-beta.1` · отдельный paired eval
Pro-пакета не гонялся (`NOT_RUN`). Подробно — [VERSION.md](VERSION.md).
