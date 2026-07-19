# Hub — `.planning/` index

> Навигация по planning-артефактам. **Ничего не удаляем** — только раскладка по смыслу и датам.  
> Обновлено: 2026-07-18 (reorg root clutter)

---

## Живой канон (корень)

| Файл | Назначение |
|------|------------|
| [`STATE.md`](STATE.md) | Текущая позиция, session notes, Pending Todos |
| [`PROJECT.md`](PROJECT.md) | Проект, цели, контекст |
| [`ROADMAP.md`](ROADMAP.md) | Дорожная карта фаз (GSD) |
| [`REQUIREMENTS.md`](REQUIREMENTS.md) | Требования |
| [`COVERAGE.md`](COVERAGE.md) | Матрица покрытия тестами |
| [`config.json`](config.json) | GSD config |

Старт сессии: **STATE** → этот README → нужная тема ниже.

---

## По смыслу

| Папка | Что внутри |
|-------|------------|
| [`july-2026/`](july-2026/) | July Rebuild / QA-HARDENING: INDEX, brief, status, roadmaps return/final, charter, EDGE catalog, expansion |
| [`audits/2026-06/`](audits/2026-06/) | Аудиты июня: AUDIT-*, FA-01/02/10/RBAC, TECH-DEBT, WAVE-C, RU-V, SCORECARD |
| [`production/`](production/) | Prod ops: PG-03, handoff, FA-06 log, PROD-01, PR v2.5 trust |
| [`specs/`](specs/) | Спеки: AI Provider Hub, Notif TG/Email, Platform Vision v4 |
| [`roadmaps/`](roadmaps/) | Исторические roadmap/backlog: DISCUSS-*, FEATURE-BACKLOG, MICROSERVICES, MILESTONE-2/3, IMPROVEMENT, EXECUTIVE |
| [`tasks/`](tasks/) | Закрытие gaps / Wave C tasks: TASKS-*, PLAN-CLOSE-GAPS |
| [`gray-zones/`](gray-zones/) | Gray-zone areas + locked decisions (`GRAY-ZONES-DECISIONS-2026-07-03.md`) |
| [`decisions/`](decisions/) | ADR |
| [`seeds/`](seeds/) | Seeds (SEED-001 rejected / SEED-002 active) |
| [`edge-expansions/`](edge-expansions/) | Edge/corner expansions (wave1–5) |
| [`research/`](research/) | Research notes |
| [`milestones/`](milestones/) | Архивы milestone phases (v1–v4) |
| [`phases/`](phases/) | Phase plans (GSD) |
| [`ab-tests/`](ab-tests/), [`ab-microservices/`](ab-microservices/) | A/B планы |
| [`code-review/`](code-review/), [`code-review-final/`](code-review-final/) | Code review артефакты |
| [`quick/`](quick/) | Задачи `gsd-quick` — папка на задачу `YYYYMMDD-<slug>/` с `PLAN.md` + `SUMMARY.md` |
| [`tooling/`](tooling/) | biome dump / parse script (не продукт) |
| [`debug/`](debug/), [`ui-reviews/`](ui-reviews/) | Прочие артефакты |

---

## Быстрые входы (июль 2026)

| Нужно | Путь |
|-------|------|
| Индекс июля | [`july-2026/INDEX-JULY-2026.md`](july-2026/INDEX-JULY-2026.md) |
| Rebuild brief | [`july-2026/JULY-2026-REBUILD-BRIEF.md`](july-2026/JULY-2026-REBUILD-BRIEF.md) |
| Gray zones locked | [`gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md`](gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md) |
| FA-06 / secrets | [`production/FA-06-EXECUTION-LOG.md`](production/FA-06-EXECUTION-LOG.md), [`production/PG-03-SECRET-ROTATION.md`](production/PG-03-SECRET-ROTATION.md) |
| Coverage | [`COVERAGE.md`](COVERAGE.md) |

---

## Правила

1. **Не удалять** артефакты — только перенос / архив по папкам.
2. Живой канон GSD остаётся в **корне** `.planning/`.
3. Новые разовые отчёты → в тему (`audits/YYYY-MM/`, `production/`, `july-YYYY/` и т.д.), не в корень.
4. После переноса обновлять ссылки в `docs/`, `AGENTS.md`, playbooks.
