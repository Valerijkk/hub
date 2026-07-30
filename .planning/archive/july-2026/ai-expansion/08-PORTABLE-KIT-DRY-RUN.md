# Portable Kit — Dry-run переноса в `corp-hub-starter`

> **Дата:** 2026-07-23 · **Цель:** полный бумажный (и воспроизводимый) dry-run переноса AI kit из Hub в гипотетический **пустой** monorepo `corp-hub-starter`.
> **Код продукта не создаём** — только дерево AI Ops / PE / planning.
> **Канон слоёв:** [`05-PORTABLE-KIT.md`](05-PORTABLE-KIT.md) · **User guide:** [`prompt-engineering-pro/PORTABLE-HUB-KIT.md`](../../../../prompt-engineering-pro/PORTABLE-HUB-KIT.md)
>
> **SUPERSEDED EXECUTION PROCEDURE (2026-07-26):** этот файл сохраняет
> исторический paper dry-run. Не выполнять старые raw-copy команды ниже.
> Канонический перенос PE: `scripts/package_portable.py` → packaged validator →
> `/bootstrap-test-context`. Physical empty-repo installation остаётся `NOT_RUN`.

## 0. Гипотеза и границы

| Параметр | Значение |
|----------|----------|
| Целевой репо | `corp-hub-starter` (пустой git, один README) |
| Источник | Hub monorepo на диске (этот репозиторий) |
| Режим | Historical paper dry-run; реальные шаги заменены packager workflow |
| Scope | L1–L5 Must/Should из `05`; L6 Could — out of scope dry-run |
| Стиль агентов | **Corporate PE** (`prompt-engineering/agents/*`) — не полный Hub `.cursor/agents/` (32 роли) |
| Не копируем | `.env`, секреты, `livekit-prod.yaml` с ключами, coverage dumps, prod hosts runbooks |

### Предусловия PASS/FAIL

| # | Условие | PASS | FAIL |
|---|---------|------|------|
| P0 | Источник содержит `prompt-engineering/`, `.cursor/skills/gsd-ru/`, `ai-expansion/05–06` | есть | отсутствует слой Must |
| P1 | Целевой репо пуст (только `.git` + README) или явная папка-песочница | чистое дерево | конфликт с чужим `.cursor/` |
| P2 | Выбран один стиль агентов (PE corporate **или** Hub agents) | зафиксирован в STATE | оба стиля смешаны без RACI |

---

## 1. Expected tree после L1–L5

```
corp-hub-starter/
├── README.md                          # продукт-заглушка starter
├── AGENTS.md                          # краткий указатель на PE + .planning
├── prompt-engineering/                # L2 — validated packager output
│   ├── PORTABLE-HUB-KIT.md
│   ├── INSTALL.md
│   ├── README.md
│   ├── agents/                        # corporate + test + templates
│   ├── context/                       # templates + generated profiles после bootstrap
│   ├── gsd-port/
│   ├── playbooks/
│   ├── prompts/
│   ├── rules/
│   └── skills/
├── .cursor/
│   ├── agents/                        # опц.: symlink/copy из PE agents (corporate)
│   ├── playbooks/
│   │   ├── prompt-contract.md         # L1 — из Hub или PE-эквивалент
│   │   └── team-workflow.md           # адаптированные роли
│   ├── skills/
│   │   └── gsd-ru/                    # L1
│   │       ├── phase-plan/
│   │       ├── wave-run/
│   │       ├── goal-verify/
│   │       ├── handoff/
│   │       ├── portable-kit/
│   │       ├── microservices-strangler/
│   │       └── README.md
│   └── rules/                         # L1 минимум + PE rules зеркало
│       ├── 00-style.mdc               # адаптировать под starter
│       ├── 10-verification.mdc        # regression / 10-этап
│       └── 20-docs-sync.mdc
└── .planning/                         # L3 + L4 + L5 refs
    ├── STATE.md
    ├── ROADMAP.md
    ├── README.md
    ├── decisions/
    │   ├── ADR-ARCH-MICROSERVICES-STUB.md   # L4
    │   └── ADR-MODELS-STUB.md               # L4
    └── ai-expansion/                  # reference copies
        ├── 04-PORTABLE-HUB-ARCH.md
        ├── 05-PORTABLE-KIT.md
        ├── 06-REDUCED-GITHUB-MODULE.md
        └── 08-PORTABLE-KIT-DRY-RUN.md  # этот файл (как инструкция)
```

**PASS дерева:** все пути L1–L3 существуют; L4 stubs + L5 refs на месте; нет `backend/`, `frontend/` продукта (starter пустой по коду).
**FAIL:** отсутствует packager/validator, `/bootstrap-test-context` или optional
`gsd-ru/portable-kit`.

---

## 2. Пошагово L1 → L5

Каждый шаг: действие → критерий PASS → критерий FAIL → команда (PowerShell / bash).

### L1 — AI Ops (Must)

| Шаг | Действие | PASS | FAIL |
|-----|----------|------|------|
| L1.1 | Скопировать `.cursor/skills/gsd-ru/` | ≥6 skills + README | пустая папка / только 1 skill |
| L1.2 | Скопировать `prompt-contract` + `team-workflow` в `.cursor/playbooks/` | оба файла; роли в team-workflow = PE corporate имена **или** явная карта alias | Hub-only имена без alias (`be1`/`fe1` без маппинга) |
| L1.3 | Agents: скопировать `prompt-engineering/agents/{orchestrator,product-owner,architect,backend-dev,frontend-dev,security-reviewer,devops}.md` (+ test-*) в `.cursor/agents/` **или** оставить только PE path | один канон путей в AGENTS.md | дубликаты с разными DoD |
| L1.4 | Минимальные rules: verification + naming/severity из `prompt-engineering/rules/` (+ docs-sync аналог) | rules читаются Cursor | rules ссылаются на Hub `backend/internal/...` без адаптации |
| L1.5 | Grep обезличенности PE agents | 0 хардкод-путей `Desktop/hub` / `backend/internal/auth` в PE agents | найдены Hub-only paths |

Historical raw-copy commands were removed. `.cursor` Hub extensions are an
optional pack and require an explicit reviewed manifest; they are not silently
included in the project-neutral PE package.

### L2 — Prompt Engineering (Must)

| Шаг | Действие | PASS | FAIL |
|-----|----------|------|------|
| L2.1 | Собрать `package_portable.py` | packaged validator 0/0 + inventory | raw source-tree copy |
| L2.2 | Исключить generated/local evidence | Hub profile/smoke отсутствуют | project data попали в output |
| L2.3 | `INSTALL.md` + `PORTABLE-HUB-KIT.md` на месте | ссылки относительные живые | битые ссылки на `.planning/archive/july-2026` без копии |

```powershell
py -3 -B .\prompt-engineering\scripts\package_portable.py `
  --root .\prompt-engineering `
  --output <empty-repo>\prompt-engineering
py -3 -B <empty-repo>\prompt-engineering\scripts\validate_toolkit.py `
  --root <empty-repo>\prompt-engineering
```

### L3 — Planning (Must)

| Шаг | Действие | PASS | FAIL |
|-----|----------|------|------|
| L3.1 | Создать `.planning/STATE.md`, `ROADMAP.md`, `README.md` | STATE с session note «kit imported» | пустой `.planning/` |
| L3.2 | Скопировать reference: `05`, `06`, `04`, `08` в `.planning/ai-expansion/` | 4 файла | только README без спек |
| L3.3 | Зафиксировать версию kit + дату в STATE | таблица версий из `05` | нет версии |

Минимальный `STATE.md` starter:

```markdown
---
status: kit-bootstrap
last_updated: "2026-07-23"
kit_version: "2026-07-22"
source: hub-ai-expansion
---
# corp-hub-starter

## Session note — kit import dry-run
- Импортированы L1–L5 по 05-PORTABLE-KIT / 08-DRY-RUN
- context profile bootstrap: pending
- Пробная волна: pending
```

### L4 — ADR stubs (Should)

| Шаг | Действие | PASS | FAIL |
|-----|----------|------|------|
| L4.1 | Stub `ADR-ARCH-MICROSERVICES` (strangler order, NATS+outbox, schema-per-service) | файл в `decisions/`, статус `proposed`/`accepted-template` | копипаста полного Hub ADR с внутренними хостами |
| L4.2 | Stub `ADR-MODELS` (оркестратор / workers / micro) | роли ↔ модели таблица | привязка к unpaid Cursor invoice Hub |

### L5 — Product blueprints (Should)

| Шаг | Действие | PASS | FAIL |
|-----|----------|------|------|
| L5.1 | Reference `06-REDUCED-GITHUB-MODULE.md` | файл + ссылка из ROADMAP | обещание «клон GitHub UI» |
| L5.2 | Skill `microservices-strangler` доступен | путь `.cursor/skills/gsd-ru/microservices-strangler/` | skill отсутствует при L4 stub |

---

## 3. Bootstrap context profile на starter

Пустой monorepo даст много `UNKNOWN` — это ожидаемо, если источники/границы
зафиксированы и факты не выдуманы.

### 3.1 Промпт агенту

```text
/bootstrap-test-context
project_root=<empty-repo>
test_roots=<explicit-test-root>
profile=corp-hub-starter
```

Если test root ещё не существует, bootstrap останавливается и возвращает
material blocker; он не придумывает framework.

### 3.2 Ожидаемый отчёт (скелет)

| Артефакт | Ожидание | PASS | FAIL |
|----------|----------|------|------|
| CONTEXT-MANIFEST.yaml | explicit roots/policy/status | complete/partial честно | выдуманный root |
| PROJECT.md | имя + evidence; стек UNKNOWN | sourced/UNKNOWN | скопирован Hub stack |
| TEST-ARCHITECTURE.md | уровни только по evidence | отсутствие явно | придуман framework |
| COMMANDS.md | discovered commands `not-run` | source+status | команда «по привычке» |
| PATTERN-INDEX.md | пуст до human curation | UNKNOWN/candidate table | Hub code snippets |

### 3.3 Команды проверки после intake

```powershell
# Незаполненные плейсхолдеры вне примеров
Select-String -Path "$dst\prompt-engineering\context\*.md" -Pattern "\{\{[A-Za-z0-9_]+\}\}"
# Hub-paths не должны появиться в context как «факт starter»
Select-String -Path "$dst\prompt-engineering\context\*.md" -Pattern "Valerii\\Desktop\\hub|backend/internal/"
```

---

## 4. Пробная волна (smoke)

Цель: проверить **делегирование + Verification**, не писать прод-код.

### 4.1 Сценарий

1. **Orchestrator** читает `agents/templates/phase-kickoff.md` + `wave-delegate.md`.
2. Задача docs-only: «Добавить в README corp-hub-starter секцию AI Ops (ссылки на PE и STATE)».
3. Границы файлов: только `README.md`, `.planning/STATE.md`.
4. Исполнитель: `backend-dev` **или** `docs` (если есть) — для starter допустим `orchestrator` → self docs; предпочтительно роль, которая умеет markdown.
5. Приёмка: блок `Verification` (10 этапов) в ответе; STATE session note обновлён.

### 4.2 Шаблон делегирования (минимум)

```
[РОЛЬ] backend-dev (docs-only dry-run)
[ЗАДАЧА] Обновить README.md: секция «AI Ops» со ссылками на prompt-engineering/ и .planning/STATE.md
[КОНТЕКСТ] corp-hub-starter после kit import; не трогать prompt-engineering/context
[ФОРМАТ] diff-описание + Verification
[ЦИКЛ КАЧЕСТВА] новое: ссылки открываются; регрессия: не удалить существующий README intro
```

### 4.3 PASS/FAIL пробной волны

| Критерий | PASS | FAIL |
|----------|------|------|
| Промпт по шаблону (роль/контекст/формат) | да | свободный «сделай README» |
| Границы файлов соблюдены | только разрешённые | правка agents/rules |
| Verification блок | есть, этап 6 = факты проверены | «вроде ок» без блока |
| Handoff | STATE note | нет STATE |
| Секреты | не добавлены | `.env` example с реальным токеном |

---

## 5. Сводный чек-лист DoD переноса

| # | Критерий (из 05 + PORTABLE-HUB-KIT) | PASS/FAIL |
|---|-------------------------------------|-----------|
| D1 | L1 AI Ops на месте (gsd-ru, playbooks, agents XOR PE, rules) | ☐ |
| D2 | L2 packaged/validated PE output | ☐ |
| D3 | L3 `.planning` STATE/ROADMAP + ai-expansion refs | ☐ |
| D4 | L4 ADR stubs (Should) | ☐ |
| D5 | L5 reduced-GitHub + strangler skill (Should) | ☐ |
| D6 | `/bootstrap-test-context` создал evidence-backed profile | ☐ |
| D7 | Агент отвечает по шаблону делегирования | ☐ |
| D8 | Risk-scaled evidence + required independent gate | ☐ |
| D9 | Нет Hub-only путей в PE agents | ☐ |
| D10 | Нет секретов в kit | ☐ |
| D11 | STATE ведётся; версия kit зафиксирована | ☐ |
| D12 | Expected tree совпадает с §1 (допуск: без L6) | ☐ |

**Вердикт dry-run:** все Must (D1–D3, D6–D11) = PASS → kit переносим; Should (D4–D5) можно отложить с записью в ROADMAP.

---

## 6. Риски и митигации

| Risk | Вероятность | Impact | Mitigation |
|------|-------------|--------|------------|
| Смешение Hub 32-agent roster с PE corporate | высокая | путаница RACI | выбрать **один** стиль; карта alias в team-workflow |
| Перенос Hub generated profile | средняя | ложные факты | packager исключает profiles; bootstrap на месте |
| Слишком толстый kit (весь `.cursor/agents` + audits) | средняя | шум | слои L1–L5; не копировать `audits/`, `production/` |
| ADR microservices не применим к стеку starter | средняя | мёртвый stub | stub = инварианты границ, стек «адаптируй» |
| Битые ссылки на Hub `.planning/archive/july-2026` | высокая | doc drift | копировать 04/05/06/08 в целевой `.planning/ai-expansion/` |
| Пробная волна на пустом коде «ломает» ожидания | низкая | ложный FAIL | docs-only задача; TODO в context = норма |
| Секреты/runbooks с хостами | низкая | leak | deny-list из `05` «Что не копировать» |
| GSD-RU vs full upstream GSD | низкая | ожидание 79 skills | явно: lite pack; полный inventory только в источнике `03` |

---

## 7. Матрица «шаг PORTABLE-HUB-KIT ↔ слой L»

| # в PORTABLE-HUB-KIT | Слой | Шаг dry-run |
|----------------------|------|-------------|
| 1 Package PE | L2 | §2 L2.1 |
| 2 gsd-ru | L1 | §2 L1.1 |
| 3 prompt-contract + team-workflow | L1 | §2 L1.2 |
| 4 agents XOR | L1 | §2 L1.3 |
| 5 rules | L1 | §2 L1.4 |
| 6 STATE/ROADMAP | L3 | §2 L3.* |
| 7 context bootstrap | post-L | §3 |
| 8 grep обезличенность | L1/L2 | L1.5 |
| 9 пробная волна | smoke | §4 |
| 10 версия в STATE | L3 | L3.3 |

---

## 8. Как повторить физически (опционально)

```powershell
# 1) Создать песочницу
New-Item -ItemType Directory -Path "C:\Users\Valerii\Desktop\corp-hub-starter" -Force
Set-Location "C:\Users\Valerii\Desktop\corp-hub-starter"
git init
"# corp-hub-starter`n" | Set-Content README.md

# 2) Собрать PE package через package_portable.py
# 3) Проверить packaged validator
# 4) Установить только явно выбранный optional .cursor pack
# 5) Открыть Cursor и выполнить /bootstrap-test-context
# 6) Пробная docs-only волна (§4)
# 7) Заполнить чек-лист §5 → приложить к STATE
```

Не коммитить в Hub артефакты песочницы. Результат dry-run (этот отчёт + чек-лист) живёт в Hub `ai-expansion/08`.

---

## 9. Связанные артефакты

| Артефакт | Роль |
|----------|------|
| [`05-PORTABLE-KIT.md`](05-PORTABLE-KIT.md) | спека слоёв + манифест |
| [`04-PORTABLE-HUB-ARCH.md`](04-PORTABLE-HUB-ARCH.md) | C4 / AI layer |
| [`06-REDUCED-GITHUB-MODULE.md`](06-REDUCED-GITHUB-MODULE.md) | L5 blueprint |
| [`07-ACCEPTANCE.md`](07-ACCEPTANCE.md) | gap «dry-run» → закрыт этим файлом |
| [`PORTABLE-HUB-KIT.md`](../../../../prompt-engineering-pro/PORTABLE-HUB-KIT.md) | user-facing 10 шагов |
| `.cursor/skills/gsd-ru/portable-kit/` | skill-обёртка |
| `prompt-engineering/skills/bootstrap-test-context/` | бутстрап context |

---

## Verification (docs-writer)

| Этап | Результат |
|------|-----------|
| 3 Факты | Historical paper flow; current execution = packager + validator + bootstrap skill |
| 6 Критич. | Не утверждаем физический copy в `corp-hub-starter` — это **инструкция dry-run**; дерево = expected |
| 9 Воспроизводимость | Другой агент по §2–§5 повторит тот же чек-лист |

*Создано: 2026-07-23 · docs-writer · AI-expansion W2/P2 gap close.*
