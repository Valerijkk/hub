---
name: portable-kit
description: Упаковка и перенос AI-операционного слоя (PE, gsd-ru, agents) в другой репозиторий. Используй для Corporate Hub Starter или dry-run переноса mid-Aug.
disable-model-invocation: true
---

# portable-kit

Упаковка и перенос AI-операционного слоя (agents, PE, GSD-RU, playbooks, planning stubs) в **новый корпоративный репозиторий**.
Исполнитель: **`docs-writer`** / **`team-lead`** / devops по задаче переноса.

## Когда

- Нужен «Corporate Hub Starter» в другом репо
- Dry-run переноса (проверка обезличенности PE)
- Обновление kit после AI-expansion mid-Aug

**Не использовать** для копирования product-кода Hub (`backend/`, `frontend/`) «на всякий».

## Спеки (канон)

| Документ | Роль |
|----------|------|
| `.planning/archive/july-2026/ai-expansion/05-PORTABLE-KIT.md` | слои L1–L6, DoD переноса |
| `.planning/archive/july-2026/ai-expansion/08-PORTABLE-KIT-DRY-RUN.md` | полный dry-run L1–L5 + PASS/FAIL (`corp-hub-starter`) |
| `prompt-engineering/PORTABLE-HUB-KIT.md` | user-facing 10 шагов · **Touchstone Standard** |
| `prompt-engineering/TOUCHSTONE.md` | продукт Standard / связь с Pro |
| `prompt-engineering/INSTALL.md` | установка Standard |
| `prompt-engineering-pro/START-HERE.md` | **Touchstone Pro** (GSD+Cursor+planning max) |
| `ai-expansion/04-PORTABLE-HUB-ARCH.md` | архитектурный контекст kit |
| Playbook dry-run | `prompt-engineering/playbooks/portable-kit-dry-run.md` |

## Слои kit

| Слой | Содержимое | Must/Should |
|------|------------|-------------|
| L1 AI Ops | agents, playbooks, prompt-contract, gsd-ru, rules | Must |
| L2 PE | output `package_portable.py` (не source tree целиком) | Must |
| L3 Planning | STATE/ROADMAP/phases templates, ссылки ai-expansion | Must |
| L4 ADR stubs | microservices, models | Should |
| L5 Blueprints | reduced-GitHub, strangler checklist | Should |
| L6 Stack map | skeleton architecture docs | Could |

## Шаги переноса (10)

1. Выбрать целевой слой L1–L5 (минимум L1+L2+L3).
2. Собрать PE через `prompt-engineering/scripts/package_portable.py`; raw copy
   source tree запрещён после появления generated profiles/local evidence.
3. Copy `.cursor/skills/gsd-ru/` (этот пакет).
4. Copy `prompt-contract.md` + `team-workflow.md` (адаптировать имена ролей).
5. Agents: `.cursor/agents/` **или** только `prompt-engineering/agents/*` (corporate) — **один** стиль, не оба конфликтующих канона.
6. Rules: verification (`40`), naming, severity из PE `rules/`; Hub-only TZ не копировать слепо.
7. Создать `.planning/STATE.md` + `ROADMAP.md` (+ README навигация).
8. Прогнать `/bootstrap-test-context` с явными project/test roots → получить
   `context/projects/<profile>/`.
9. **Grep обезличенности:** в PE agents/prompts нет путей старого проекта (`hub/`, `backend/internal/...` вне явных examples).
10. Пробная волна: orchestrator → phase-kickoff / docs-задача по шаблону делегирования → проверить Verification блок.

Зафиксировать в STATE целевого репо: версия kit + дата переноса.

## Полный чек-лист копирования

```markdown
## Portable kit transfer — checklist
Источник: <hub path> → Назначение: <new repo>
Слои: L1 [ ] L2 [ ] L3 [ ] L4 [ ] L5 [ ] L6 [ ]

### Copy
- [ ] packaged prompt-engineering/ (validator PASS)
- [ ] .cursor/skills/gsd-ru/
- [ ] playbooks: prompt-contract, team-workflow (adapted)
- [ ] agents: cursor XOR pe-corporate
- [ ] PE rules 00–40 (min 40-verification)
- [ ] .planning stubs: STATE, ROADMAP, README
- [ ] Reference: 05-PORTABLE-KIT, 06-REDUCED-GITHUB (optional), strangler skill

### Sanitize
- [ ] Нет .env / ключей / livekit-prod.yaml с секретами
- [ ] Нет coverage dumps
- [ ] Prod hostnames обезличены или удалены
- [ ] Grep: старые project paths в PE agents = 0 (кроме examples)

### Activate
- [ ] generated context profile заполнен через `/bootstrap-test-context`
- [ ] Пробная волна §5.2-эквивалент (_MASTER-PROMPT-TEMPLATE)
- [ ] Verification блок в ответе агента
- [ ] STATE note о переносе
```

## Шаблон STATE note (целевой репо)

```markdown
## Session note — YYYY-MM-DD (portable-kit import)
- Источник kit: Hub ai-expansion / PE версия <date>
- Слои: L1–L3 (… )
- context profile bootstrap: PASS
- Пробная волна: orchestrator → … PASS
- Отклонения от канона Hub: <роли переименованы / …>
- Next: первая продуктовая фаза по phase-plan
```

## Worked example (Hub → новый репо)

**Сценарий:** перенос AI Ops для внутреннего «Project X» без кода Hub.

1. Build validated PE package + copy optional gsd-ru/prompt-contract.
2. Роли: `orchestrator` вместо `team-lead` (corporate PE agents).
3. `context/projects/project-x/PROJECT.md` — стек Project X; **не** копировать
   Hub TZ Magic Link как обязательный.
4. Grep: удалить случайные `c:\Users\Valerii\Desktop\hub` из черновиков.
5. Dry-run: `phase-plan` docs-задачи → `wave-run` (1 агент) → `goal-verify` → `handoff`.
6. DoD: context заполнен, Verification в ответах, секретов нет.

Спека слоёв: `05-PORTABLE-KIT.md`. User guide: `PORTABLE-HUB-KIT.md`.

## Что НЕ копировать

| Запрещено | Почему |
|-----------|--------|
| `.env`, реальные токены | секреты |
| `livekit-prod.yaml` с ключами | credentials |
| `backend/`, `frontend/` целиком | не AI kit |
| Coverage dumps / `*.out` | мусор |
| Prod-only runbooks с внутренними хостами | без обезличивания |
| Хардкод Hub-путей в PE agents | ломает переносимость |

## DoD

- [ ] Выбранные слои собраны packager-ом и проверены validator-ом
- [ ] `/bootstrap-test-context` создал profile
- [ ] Делегирование (wave-delegate / §5.2-эквивалент) работает
- [ ] Risk-scaled evidence block и independent gate для high-impact
- [ ] Нет секретов в kit
- [ ] Нет Hub-only путей в PE agents (вне examples)
- [ ] STATE.md целевого репо ведётся
- [ ] Пробная волна PASS

## Anti-patterns

| Антипаттерн | Почему | Что вместо |
|-------------|--------|------------|
| Копировать `.env` | утечка | только `.env.example` placeholders |
| Тащить весь backend | не portable kit | L1–L3 AI Ops |
| Два канона агентов сразу | конфликт инструкций | cursor **или** PE corporate |
| Оставить пути Hub в PE | kit не переносится | grep + context |
| Пропустить пробную волну | «скопировали = работает» | dry-run docs task |
| Копировать TZ Hub как универсальный закон | другой продукт | адаптировать constraints |

## Связь

- Спеки: `05-PORTABLE-KIT.md`, `08-PORTABLE-KIT-DRY-RUN.md`, `PORTABLE-HUB-KIT.md`
- Playbook: `prompt-engineering/playbooks/portable-kit-dry-run.md`
- Смежный skill: [microservices-strangler](../microservices-strangler/SKILL.md) (blueprint в L5)
- Индекс: [gsd-ru/README.md](../README.md) · AI-expansion README
