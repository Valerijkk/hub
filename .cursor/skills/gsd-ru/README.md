# GSD-RU — русский GSD-lite для Hub

> Hub-адаптация GSD с воротами rule 56, prompt-contract и team-lead §5.2.
> Не заменяет `~/.claude/skills/gsd-*`: закрывает ежедневный цикл
> plan→wave→verify→handoff и portable kit.
> Язык skills: **RU**. Основной `SKILL.md` остаётся сфокусированным, детали
> загружаются по ссылкам (progressive disclosure); line-count не является KPI.

## Skills

| Skill | Путь | Когда | Исполнитель |
|-------|------|-------|-------------|
| phase-plan | [phase-plan/SKILL.md](phase-plan/SKILL.md) | Декомпозиция фазы на атомарные планы | `tech-lead` (docs-only — `team-lead`) |
| wave-run | [wave-run/SKILL.md](wave-run/SKILL.md) | Запуск волны агентов по §5.2 | `team-lead` |
| goal-verify | [goal-verify/SKILL.md](goal-verify/SKILL.md) | Goal-backward приёмка | `team-lead` / `*-ver` / `product-ver` |
| handoff | [handoff/SKILL.md](handoff/SKILL.md) | Role/wave outcome + evidence + передача | любой агент → следующий |
| portable-kit | [portable-kit/SKILL.md](portable-kit/SKILL.md) | Упаковка/перенос AI kit | `docs-writer` / `team-lead` / devops |
| microservices-strangler | [microservices-strangler/SKILL.md](microservices-strangler/SKILL.md) | Чек-лист выноса сервиса (planning) | `arch` + `tech-lead` |

## Отличие от upstream GSD

| | Upstream GSD | GSD-RU Hub |
|--|--------------|------------|
| Язык | EN | RU |
| Оркестрация | часто solo-agent | `team-lead` + 32 роли |
| Quality | phase verify | + цикл новое/старое ([rule 56](../../rules/56-regression-verify.mdc)) |
| Делегирование | свободно | **только** шаблон [team-lead.md §5.2](../../agents/team-lead.md) |
| Portable | project GSD | PE + agents kit ([05-PORTABLE-KIT](../../../.planning/archive/july-2026/ai-expansion/05-PORTABLE-KIT.md)) |
| Microservices | generic | [ADR-ARCH-MICROSERVICES](../../../.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md) + strangler checklist |

## Типовой день

```
phase-plan → wave-run (×N) → goal-verify → handoff
```

Параллельно (по запросу):

- `portable-kit` — перенос AI-слоя в другой репо
- `microservices-strangler` — planning-only вынос домена (код freeze продукта не мешает планировать)

## Канонические якоря Hub

| Якорь | Путь |
|-------|------|
| Шаблон делегирования §5.2 | `.cursor/agents/team-lead.md` §5.2 · `team-workflow.md` §5.3.1 |
| Заполненный пример §5.3 | bases security → be1 (`team-lead.md`) |
| Цикл качества | `.cursor/rules/56-regression-verify.mdc` |
| Docs-sync | `.cursor/rules/55-docs-sync.mdc` |
| Artifacts cleanup | `.cursor/rules/45-artifacts-cleanup.mdc` |
| Контракт промпта | `.cursor/playbooks/prompt-contract.md` |
| Microservices ADR | `.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md` |
| AI-expansion | `.planning/archive/july-2026/ai-expansion/` |
| PE-зеркало | `prompt-engineering/skills/{phase-plan,wave-orchestrate,verify-goal}/` |

## Worked example (сквозной): bases security

Исторический кейс Волны 1 (старт 2026-07-03, merged пользователем
2026-07-11), эталон rule 56 кейс 1 + `team-lead.md` §5.3.

1. **phase-plan** — цель: «чужая база недоступна (403/404), владелец без регрессии»; задачи be1 → be-ver → docs; ветка отдельно от эпика.
2. **wave-run** — промпт be1 строго по §5.2: границы `backend/internal/bases/*` + `testframework/bases/*`, не трогать `ent/schema`, оба контура rule 56.
3. **goal-verify** — evidence: `authorizeBase`, `rbac_test.go`, лог `go test -tags hubtest ./testframework/bases/...`, `docs/wiki/API.md`.
4. **handoff** — outcome/evidence → be-ver по §5.2; `/handoff-sync` закрыл
   STATE/docs, после чего пользователь смержил подготовленные ветки 2026-07-11.

Коммиты-якоря: `98fb3df` (fix), `8ff10d7` (RBAC denial suite).

## Progressive disclosure

`SKILL.md` содержит только триггер, обязательный рабочий цикл, минимальный DoD
и прямые ссылки на канонические правила. Подробные справочники выносятся в
`references/` и читаются только по необходимости. Полный шаблон, worked
example и таблица anti-patterns не копируются в каждый skill, если уже есть
канонический playbook или rule. Длина файла определяется задачей, не числом
строк.

## Маршрутизация

| Нужно | Куда |
|-------|------|
| Полная матрица агент→GSD | `.cursor/GSD-INTEGRATION.md` |
| Выбор команды | `.cursor/GSD-INTEGRATION.md` (канон), `gsd-help` / `gsd-do` |
| FAIL теста | `.cursor/skills/test-triage/SKILL.md` |
| Планы mid-Aug | `.planning/archive/july-2026/ai-expansion/README.md` |
| Приёмка Must AI-expansion | `ai-expansion/07-ACCEPTANCE.md` |

## DoD пакета GSD-RU

- [ ] Каждый исполняемый каталог содержит валидный `SKILL.md`
- [ ] Frontmatter `name` совпадает с именем каталога
- [ ] Основной workflow краток; детали подключаются progressive disclosure
- [ ] Ссылки на §5.2 / rule 56 / ADR не битые
- [ ] README отражает все 6 nested skills

## Anti-patterns пакета

| Антипаттерн | Что вместо |
|-------------|------------|
| Дублировать полный upstream GSD | lite-цикл 4 skills + 2 спец |
| Хардкодить секреты/prod-хосты в skills | только пути docs/ADR |
| Обещать UI/API без проверки кода | evidence в goal-verify |
| wave-run без phase-plan при >2 агентах | сначала PLAN |
| Свободный «сделай X» | шаблон §5.2 |
| Копировать product-код в portable-kit | слои L1–L3 AI Ops |

## Версия

- Пакет GSD-RU: **2026-07-26** (progressive disclosure, canonical routing)
- Согласован с: prompt-contract 2026-07-19, ADR microservices accepted 2026-07-02, AI-expansion Must PASS
