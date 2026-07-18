# Hub — Index: July 2026 planning

> Оглавление planning-артефактов июля 2026.  
> Все пути ниже — на диске (QA-HARDENING P0, 2026-07-10).  
> Обновлено: 2026-07-18 · файлы июля лежат в `.planning/july-2026/` · навигация всего `.planning/`: [`../README.md`](../README.md)

---

## Статус и позиция

| Артефакт | Путь | Назначение |
|----------|------|------------|
| STATE | `.planning/STATE.md` | Текущая позиция, session notes, Pending Todos |
| Project status snapshot | `.planning/july-2026/PROJECT-STATUS-2026-07-10.md` | Сводка статуса на 2026-07-10 |
| July return ops plan | `.planning/july-2026/ROADMAP-DETAILED-JULY-RETURN.md` | W0–N, RACI, DoR/DoD, чек-лист D1 |
| AGENTS (кратко) | `AGENTS.md` § «Текущий статус» | Метрики + next steps для агентов |
| COVERAGE | `.planning/COVERAGE.md` | Матрица покрытия (baseline COV 83.0%) |

---

## July Rebuild (возврат ~конец июля)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Rebuild brief | `.planning/july-2026/JULY-2026-REBUILD-BRIEF.md` | Снимок команды, as-is calls/calendar, kickoff; **LiveKit stay** |
| Roadmap (pointer) | `.planning/july-2026/ROADMAP-FINAL-2026-07-03.md` § «Отложено: July Rebuild» | Эпики CAL-SYNC / QA-HARDENING / Redesign; волны 1–11 |
| Gray zones | `.planning/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` | Locked решения (в т.ч. calendar iCal без OAuth) |

### Seeds

| Seed | Путь | Статус |
|------|------|--------|
| SEED-001 mediasoup / SFU | `.planning/seeds/SEED-001-mediasoup-migration.md` | **REJECTED** — остаёмся на LiveKit |
| SEED-002 calendar + RSVP + mail | `.planning/seeds/SEED-002-calendar-calls-email-sync.md` | **Active** — триггер конец июля |

---

## QA-HARDENING (planning wave 2026-07-10)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Code Quality Charter | `.planning/july-2026/CODE-QUALITY-CHARTER.md` | Correctness = maintainability; DoD be/fe; антипаттерны |
| Edge / corner catalog | `.planning/july-2026/EDGE-CORNER-CATALOG.md` | База + индекс · **≈4696 ECC** |
| Edge expansions 01–05 | `.planning/edge-expansions/01…05-*.md` | Wave1 **+427 Edge** |
| Corner parent + C01–C05 | `.planning/edge-expansions/00-CORNER-PARENT.md`, `corners/C01…C05` | Wave1 **+588 Corner** |
| Wave2 edge+corners | `.planning/edge-expansions/wave2/` | **+771** |
| Wave3–5 | `.planning/edge-expansions/wave{3,4,5}/` | **+2459** (781+848+830) |
| Testing Guide | `testframework/TESTING_GUIDE.md` | Канон автотестов: SOLID/DRY/KISS/POM |
| Guides 01–06 | `testframework/guides/` | Principles, POM, Go, Vitest, E2E, edge coverage |
---

## Команда и процесс (июльский контекст)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Playbooks index | `.cursor/playbooks/README.md` | 30 агентов, модели, RACI |
| Team workflow | `.cursor/playbooks/team-workflow.md` | Волны, кросс-ревью, merge policy |
| Backend / Frontend / Comments | `.cursor/playbooks/backend.md`, `frontend.md`, `code-comments.md` | Практики кода |
| Docs sync rule | `.cursor/rules/55-docs-sync.mdc` | Матрица docs после фич |
| Regression gate | `.cursor/rules/56-regression-verify.mdc` | Доменная регрессия + `/qa-gate` |
| TZ constraints | `.cursor/rules/60-tz-constraints.mdc` · `docs/tz_constraints.md` | Mandatory / forbidden |
| Style rules | `.cursor/rules/10-style.mdc`, `20-backend.mdc`, `30-frontend.mdc` | Стиль + слои |

---

## Архитектура / ADR (релевантно July)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Microservices ADR | `.planning/decisions/ADR-ARCH-MICROSERVICES-2026.md` | Strangler-fig; hub-calls = LiveKit-домен |
| TZ rewrite ADR | `.planning/decisions/ADR-TZ-REWRITE-2026.md` | Scope by phase |
| Expansion / market research | `.planning/july-2026/EXPANSION-ANALYSIS-2026-07-02.md`, `.planning/research/MARKET-AND-NEEDS-2026-07-02.md` | Контекст продукта |

---

## Ops / prod blockers (не July-feature, но в Pending)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| FA-06 runbook | `docs/runbooks/fa06-secret-rotation.md` | Ручная ротация секретов |
| FA-06 execution log | `.planning/production/FA-06-EXECUTION-LOG.md` | Phase 0 scan |
| Production handoff | `.planning/production/PRODUCTION-HANDOFF.md` | Cutover checklist |

---

## Как пользоваться индексом

1. Старт сессии июля → `STATE.md` + этот INDEX.  
2. Rebuild / CAL-SYNC → brief + SEED-002 (SEED-001 не поднимать).  
3. Тесты / edge → Testing Guide + EDGE-CORNER-CATALOG (не копировать сюда).  
4. Код PR → CODE-QUALITY-CHARTER + playbooks + `*-ver` / code-reviewer.  
5. Мерж в `master` — **только пользователь**; Волна 1 уже в фича-ветках.

---

*Не переписывать содержимое соседних артефактов — только ссылки.*
