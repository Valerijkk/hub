# Hub — `.planning/` index

> Навигация по planning-артефактам. **Ничего не удаляем** — только раскладка по смыслу и датам.
> Обновлено: 2026-08-04 — детальный план августа; уровни проверки артефакта

---

## Живой канон (корень)

| Файл | Назначение |
|------|------------|
| [`STATE.md`](STATE.md) | **Live** позиция, session notes, Pending Todos, blockers, deploy_blockers |
| [`PROJECT.md`](PROJECT.md) | Проект, цели, контекст |
| [`ROADMAP.md`](ROADMAP.md) | Дорожная карта фаз (GSD) |
| [`REQUIREMENTS.md`](REQUIREMENTS.md) | Требования |
| [`COVERAGE.md`](COVERAGE.md) | Матрица покрытия тестами |
| [`VERIFICATION-LEVELS.md`](VERIFICATION-LEVELS.md) | Четыре уровня проверки артефакта при приёмке фазы |
| [`EXTRACTION-CANDIDATES.md`](EXTRACTION-CANDIDATES.md) | Что обкатано в Hub и готово к выносу в продукт |
| [`TOUCHSTONE-IDEAS.md`](TOUCHSTONE-IDEAS.md) | Копилка идей продукта: режимы, вход с нуля, механизмы экономии контекста |
| [`config.json`](config.json) | GSD config |

Старт сессии: **STATE** → этот README → нужная тема ниже.

---

## Структура

Одиннадцать папок вместо прежних двадцати четырёх. Деление — по **назначению**,
а не по дате: активное отдельно от исторического, решения отдельно от аудитов.

### Управляется GSD — пути зашиты в скилах, не переименовывать

| Папка | Что внутри |
|-------|------------|
| [`phases/`](phases/) | Phase plans (GSD), 124 фазы + backlog `999-*` |
| [`milestones/`](milestones/) | Архивы завершённых milestone: v1.0 (01–07), v2.0 (08–13), v4 |
| [`research/`](research/) | Research notes перед планированием фаз |
| [`debug/`](debug/) | Сессии `gsd-debug` |
| [`ui-reviews/`](ui-reviews/) | Отчёты `gsd-ui-review` |

> ⚠️ Эти пути жёстко прописаны в `~/.claude/agents/gsd-*` и `~/.claude/skills/gsd-*`.
> Перемещение сломает GSD-команды. Остальные папки ниже — свободны.

### Активная работа

| Папка | Что внутри |
|-------|------------|
| [`backlog/`](backlog/) | **Задачи исполнителям** — [таблица со ссылками](backlog/README.md), формат по мастер-промпту |
| [`active/august-2026/`](active/august-2026/) | Планы на август: [действующий детальный](active/august-2026/05-AUGUST-DETAILED.md), [пересмотр приоритетов](active/august-2026/04-REVISED-PRIORITIES.md), [микросервисы отложены](active/august-2026/01-MICROSERVICES-SPLIT.md) |

Задачи пишут **team-lead и architect**, по
[мастер-промпту](../pro-qa-ai/agents/templates/_MASTER-PROMPT-TEMPLATE.md).
Исполнитель задачу не сочиняет — выполняет и отчитывается.
Правила: [`.cursor/rules/57-backlog-tasks.mdc`](../.cursor/rules/57-backlog-tasks.mdc).

### Решения и спецификации

| Папка | Что внутри |
|-------|------------|
| [`decisions/`](decisions/) | ADR — архитектурные решения |
| [`decisions/gray-zones/`](decisions/gray-zones/) | 6 областей серых зон + [зафиксированные решения](decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md) |
| [`decisions/specs/`](decisions/specs/) | Спеки модулей: AI Provider Hub, Notif TG/Email, Platform Vision v4 |
| [`decisions/ab-microservices/`](decisions/ab-microservices/) | A/B-планы выноса в микросервисы (fable / opus) |

### Аудиты и ревью

| Папка | Что внутри |
|-------|------------|
| [`audits/2026-06/`](audits/2026-06/) | Аудиты июня: FA-01/02/10, RBAC, TECH-DEBT, WAVE-C, SCORECARD |
| [`audits/2026-07/`](audits/2026-07/) | July repo-scan + provenance policy |
| [`audits/grok-2026-07/`](audits/grok-2026-07/) | Аудит кода: 261 находка, волны фиксов W1–W20 |
| [`audits/planning-audit-2026-07/`](audits/planning-audit-2026-07/) | Аудит планирования: 11 находок, матрица фаз, классификации NFR |
| [`audits/code-review/`](audits/code-review/), [`audits/code-review-final/`](audits/code-review-final/) | Code review артефакты |

### Референс

| Папка | Что внутри |
|-------|------------|
| [`reference/edge-expansions/`](reference/edge-expansions/) | Edge/corner-кейсы по доменам — источник TC и материал для Swagger |
| [`reference/seeds/`](reference/seeds/) | Отложенные идеи (SEED-001 rejected, SEED-002 active) |

### Слой качества тестов (вне `.planning/`, связано)

AI-слой для тестовых фреймворков живёт в [`pro-qa-ai/`](../pro-qa-ai/TOUCHSTONE.md).
Planning хранит историю решений по нему (`archive/july-2026/ai-expansion/`).

Граница по **объекту работы**: `.cursor/` + `.planning/` — про продукт Hub,
`pro-qa-ai/` — про тестовый фреймворк как таковой.

### Операционное

| Папка | Что внутри |
|-------|------------|
| [`ops/production/`](ops/production/) | Prod ops: PG-03, FA-06, handoff, PR v2.5 trust |
| [`ops/tooling/`](ops/tooling/) | biome dump и парсер — не продукт |

### Архив — историческое, завершённое

| Папка | Что внутри |
|-------|------------|
| [`archive/july-2026/`](archive/july-2026/) | July Rebuild / QA-HARDENING: индекс, brief, charter, EDGE catalog, AI expansion |
| [`archive/roadmaps/`](archive/roadmaps/) | Исторические roadmap и backlog: DISCUSS-*, FEATURE-BACKLOG, MICROSERVICES, EXECUTIVE |
| [`archive/tasks/`](archive/tasks/) | Закрытие gaps, Wave C tasks |
| [`archive/ab-tests/`](archive/ab-tests/) | A/B-планы оркестрации (bases, incident) |
| [`archive/quick/`](archive/quick/) | Задачи `gsd-quick`, папка на задачу `YYYYMMDD-<slug>/` |
| [`archive/history/`](archive/history/) | Session notes и completion logs ≤2026-07-23 — **не** live STATE |

---

## Быстрые входы

| Нужно | Путь |
|-------|------|
| **Планы на август — действующий** | [`active/august-2026/05-AUGUST-DETAILED.md`](active/august-2026/05-AUGUST-DETAILED.md) |
| Почему порядок такой | [`active/august-2026/04-REVISED-PRIORITIES.md`](active/august-2026/04-REVISED-PRIORITIES.md) |
| Что не так с планированием | [`audits/planning-audit-2026-07/00-PLANNING-AUDIT.md`](audits/planning-audit-2026-07/00-PLANNING-AUDIT.md) |
| Реальный статус фаз | [`audits/planning-audit-2026-07/PHASE-STATUS-MATRIX.md`](audits/planning-audit-2026-07/PHASE-STATUS-MATRIX.md) |
| Решения, ждущие владельца | [`audits/planning-audit-2026-07/DECISIONS-PENDING.md`](audits/planning-audit-2026-07/DECISIONS-PENDING.md) |
| Свежее покрытие | [`audits/planning-audit-2026-07/cov-2026-07-27-REPORT.md`](audits/planning-audit-2026-07/cov-2026-07-27-REPORT.md) |
| Gray zones locked | [`decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`](decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md) |
| FA-06 / секреты | [`ops/production/PG-03-SECRET-ROTATION.md`](ops/production/PG-03-SECRET-ROTATION.md) |
| Индекс июля | [`archive/july-2026/INDEX-JULY-2026.md`](archive/july-2026/INDEX-JULY-2026.md) |
| AI-слой для тестового фреймворка | [`../pro-qa-ai/TOUCHSTONE.md`](../pro-qa-ai/TOUCHSTONE.md) |
| Порядок фаз и ворота между ними | [`../.cursor/playbooks/phase-loop.md`](../.cursor/playbooks/phase-loop.md) |
| Уровни проверки артефакта | [`VERIFICATION-LEVELS.md`](VERIFICATION-LEVELS.md) |
| Что готово к выносу в продукт | [`EXTRACTION-CANDIDATES.md`](EXTRACTION-CANDIDATES.md) |
| Идеи Touchstone | [`TOUCHSTONE-IDEAS.md`](TOUCHSTONE-IDEAS.md) |
| AI-expansion планы (история) | [`archive/july-2026/ai-expansion/README.md`](archive/july-2026/ai-expansion/README.md) |
| Активный backlog задач | [`backlog/README.md`](backlog/README.md) |

---

## Правила

1. **Не удалять** артефакты — только перенос и архив по папкам.
2. Живой канон GSD остаётся в **корне** `.planning/`.
3. Папки из раздела «Управляется GSD» **не переименовывать и не перемещать** —
   их пути зашиты в скилах.
4. Новые разовые отчёты — в тему (`audits/YYYY-MM/`, `ops/production/`), не в корень.
5. Завершённое переносить в `archive/`, а не оставлять рядом с активным.
6. После любого переноса — прогнать проверку ссылок и обновить этот README.

### Проверка ссылок после переноса

```bash
python - <<'PY'
import os,re
broken=[]
for root,dirs,fs in os.walk("."):
    dirs[:]=[d for d in dirs if d not in {".git","node_modules","ent",".venv","dist"}]
    for f in fs:
        if not f.endswith(".md"): continue
        p=os.path.join(root,f)
        try: s=open(p,encoding="utf-8").read()
        except Exception: continue
        for m in re.finditer(r'\]\(([^)#\s]+\.md)(#[^)]*)?\)', s):
            t=m.group(1)
            if t.startswith(("http","mailto")): continue
            if not os.path.exists(os.path.normpath(os.path.join(root,t))):
                broken.append(f"{p} -> {t}")
print("битых ссылок:", len(broken))
for b in broken[:20]: print("  ", b)
PY
```
