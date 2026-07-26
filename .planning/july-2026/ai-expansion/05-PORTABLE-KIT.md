# Portable Kit — спецификация содержимого

## Назначение

Два слоя: project-neutral Test Automation AI Quality Layer (собирается безопасным
packager-ом) и опциональный Hub corporate/GSD pack.

## Слои kit

| Слой | Содержимое | Обязателен? |
|------|------------|-------------|
| L1 AI Ops | agents, playbooks, prompt-contract, gsd-ru, rules | Must |
| L2 PE | output `package_portable.py`, не raw source tree | Must |
| L3 Planning | STATE/ROADMAP/phases templates, `ai-expansion/` | Must |
| L4 ADR stubs | microservices, models | Should |
| L5 Product blueprints | reduced-GitHub, strangler checklist | Should |
| L6 Stack map | опционально docs architecture skeleton | Could |

## Порядок копирования (в новый репо)

1. Собрать `prompt-engineering/` через `scripts/package_portable.py`; проверить
   packaged output его `scripts/validate_toolkit.py`.
2. `.cursor/playbooks/prompt-contract.md` + `team-workflow.md` (или PE corporate equivalents).
3. `.cursor/skills/gsd-ru/` + нужные slash-skills (`test-triage`, `handoff-sync` аналоги).
4. `.cursor/agents/` **или** только `prompt-engineering/agents/*` (corporate) — выбрать один стиль.
5. `.cursor/rules/` минимальный набор: style, regression-verify, docs-sync, tz-constraints (адаптировать).
6. `.planning/` шаблоны: `STATE.md`, `ROADMAP.md`, `README.md`, `decisions/`.
7. Скопировать `ai-expansion/04`, `05`, `06`, `08` как reference.
8. Прогнать `/bootstrap-test-context` с явными project/test roots.
9. Проверить `context/projects/<profile>/` и все `UNKNOWN`.
10. Пробная волна: orchestrator → backend-dev (dry doc task).

User-facing: [`prompt-engineering/PORTABLE-HUB-KIT.md`](../../../prompt-engineering/PORTABLE-HUB-KIT.md).

## Версии / совместимость

| Компонент | Версия kit | Примечание |
|-----------|------------|------------|
| Portable PE | `prompt-engineering/VERSION.md` | canonical semver/status |
| prompt-contract | 2026-07-26 | evidence-first + routing |
| PE base | 2026-07-26 | context profiles, recipes, validator, packager |
| GSD-RU | 2026-07-22 | lite, не full upstream |
| ADR microservices | accepted 2026-07-02 | |

## DoD переноса

- [ ] Packaged validator PASS
- [ ] `/bootstrap-test-context` создал generated profile
- [ ] Агент отвечает по шаблону делегирования
- [ ] Risk-scaled evidence/independent gate выполнены
- [ ] Нет Hub-only путей в PE agents
- [ ] STATE.md ведётся

## Что не копировать

- Секреты, `.env`, `livekit-prod.yaml` с ключами
- Coverage dumps
- Prod-only runbooks с внутренними хостами (опционально обезличить)
- Generated `prompt-engineering/context/projects/*`, local smoke/run evidence,
  project-only pattern/review content
- `.planning/audits/`, `production/`, edge-expansions — не часть kit
- Полный roster `.cursor/agents/` (32), если выбран стиль **PE corporate** (дублирование RACI)

---

## Манифест файлов по слоям (L1–L5)

| Слой | Источник в Hub | Назначение в целевом репо |
|------|----------------|---------------------------|
| L1 | `.cursor/skills/gsd-ru/**` | RU GSD-lite skills |
| L1 | `.cursor/playbooks/prompt-contract.md`, `team-workflow.md` | контракт + волны (адаптировать роли) |
| L1 | `prompt-engineering/agents/{orchestrator,product-owner,architect,backend-dev,frontend-dev,security-reviewer,devops,test-*}.md` **или** `.cursor/agents/*` | один стиль агентов |
| L1 | `prompt-engineering/rules/*.md` (+ минимальные `.cursor/rules`) | verification / naming / severity |
| L2 | output `prompt-engineering/scripts/package_portable.py` | project-neutral PE toolkit |
| L3 | шаблоны `.planning/{STATE,ROADMAP,README}.md` + `ai-expansion/{04,05,06,08}` | planning + reference |
| L4 | stubs на базе `ADR-ARCH-MICROSERVICES-2026`, `ADR-MODELS-GPT56-2026` | обезличенные ADR |
| L5 | `06-REDUCED-GITHUB-MODULE.md`, skill `microservices-strangler` | blueprints процессов |

---

## Выбор стиля агентов (XOR)

| Стиль | Когда | Плюс | Минус |
|-------|-------|------|-------|
| **A · PE corporate** | новый репо / внешняя команда | обезличен, 7+4 ролей | нет be1/fe1 параллели |
| **B · Hub agents** | форк/продолжение Hub | полный RACI 32 | тащит Hub-paths в правилах |

В `AGENTS.md` целевого репо явно: `agent_style: pe-corporate | hub-roster`. Смешивать без карты alias — **запрещено**.

---

## Dry-run переноса

Полный исторический сценарий с PASS/FAIL; актуальный маршрут использует packager
и `/bootstrap-test-context`:

→ [`08-PORTABLE-KIT-DRY-RUN.md`](08-PORTABLE-KIT-DRY-RUN.md) (цель: гипотетический monorepo `corp-hub-starter`).

Краткий маршрут: package → packaged validator → bootstrap profile → physical
empty-repo install → docs-only smoke → DoD.

---

## Версионирование kit при обновлении

1. Не включать/перезаписывать `context/projects/<profile>/` целевого проекта.
2. Обновлять L1/L2/L5 skills атомарно; фиксировать `kit_version` + дату в STATE.
3. При breaking change prompt-contract — bump в таблице «Версии / совместимость» и note в ROADMAP.
4. Регрессия переноса: повторить пробную волну из `08` §4 на docs-задаче.

---

## Анти-паттерны переноса

| Антипаттерн | Почему плохо | Вместо |
|-------------|--------------|--------|
| Copy всего `.planning/` | шум, секреты ops, drift | только шаблоны + 04/05/06/08 |
| Filled context из Hub | ложный стек | template + intake |
| Оба стиля агентов | конфликт DoD | XOR § выше |
| Обещать код reduced-GitHub | scope kit = процесс | blueprint `06` |
| Пропуск Verification | «вроде работает» | rules/40 обязателен |