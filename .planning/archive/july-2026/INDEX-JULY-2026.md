# Hub — Index: July 2026 planning

> Оглавление planning-артефактов июля 2026.  
> Все пути ниже — на диске (QA-HARDENING P0, 2026-07-10).  
> Обновлено: 2026-07-26 · Test Automation AI Quality Layer rework · файлы июля лежат в `.planning/archive/july-2026/` · навигация всего `.planning/`: [`../README.md`](../../README.md)

---

## AI-expansion (22.07 → 15.08.2026)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Index | [ai-expansion/README.md](ai-expansion/README.md) | Порт GSD + portable kit + agents/PE |
| Product scope | [ai-expansion/01-PRODUCT-SCOPE.md](ai-expansion/01-PRODUCT-SCOPE.md) | MoSCoW, metrics |
| Tech plan | [ai-expansion/02-TECH-PLAN.md](ai-expansion/02-TECH-PLAN.md) | 26 задач, W1–W4 + structural/package/eval gates |
| GSD inventory | [ai-expansion/03-GSD-PORT-INVENTORY.md](ai-expansion/03-GSD-PORT-INVENTORY.md) | ~79 skills → P0/P1/P2 |
| Portable arch | [ai-expansion/04-PORTABLE-HUB-ARCH.md](ai-expansion/04-PORTABLE-HUB-ARCH.md) | C4 + microservices + AI layer |
| Portable kit | [ai-expansion/05-PORTABLE-KIT.md](ai-expansion/05-PORTABLE-KIT.md) | Что копировать |
| Reduced GitHub | [ai-expansion/06-REDUCED-GITHUB-MODULE.md](ai-expansion/06-REDUCED-GITHUB-MODULE.md) | Blueprint issue/PR/checks |
| Acceptance | [ai-expansion/07-ACCEPTANCE.md](ai-expansion/07-ACCEPTANCE.md) | Исторический Must PASS; rework-gates перепроверяются |
| Portable kit dry-run | [ai-expansion/08-PORTABLE-KIT-DRY-RUN.md](ai-expansion/08-PORTABLE-KIT-DRY-RUN.md) | L1–L5 → `corp-hub-starter`, PASS/FAIL |
| AI quality layer rework | [ai-expansion/09-AI-QUALITY-LAYER-REWORK.md](ai-expansion/09-AI-QUALITY-LAYER-REWORK.md) | Context bootstrap, patterns, manual colleague examples, recipes/evals |
| GSD-RU skills | `.cursor/skills/gsd-ru/` | RU lite |
| PE portable guide | `pro-qa-ai/PORTABLE-HUB-KIT.md` | 10 шагов переноса |
| July repo-scan index | [`../audits/2026-07/repo-scan/INDEX.md`](../../audits/2026-07/repo-scan/INDEX.md) | 40 report slots, not 40 independent runs |
| Repo-scan provenance | [`../audits/2026-07/repo-scan/PROVENANCE.md`](../../audits/2026-07/repo-scan/PROVENANCE.md) | Confidence classes; SCORECARD non-authoritative |

---

## Статус и позиция

| Артефакт | Путь | Назначение |
|----------|------|------------|
| STATE | `.planning/STATE.md` | Текущая позиция, session notes, Pending Todos |
| Project status snapshot | `.planning/archive/july-2026/PROJECT-STATUS-2026-07-10.md` | Сводка статуса на 2026-07-10 |
| July return ops plan | `.planning/archive/july-2026/ROADMAP-DETAILED-JULY-RETURN.md` | W0–N, RACI, DoR/DoD, чек-лист D1 |
| AGENTS (кратко) | `AGENTS.md` § «Текущий статус» | Метрики + next steps для агентов |
| COVERAGE | `.planning/COVERAGE.md` | Матрица покрытия (baseline COV 83.0%, 2026-06-22 Docker) |

---

## July Rebuild (возврат ~конец июля)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Rebuild brief | `.planning/archive/july-2026/JULY-2026-REBUILD-BRIEF.md` | Снимок команды, as-is calls/calendar, kickoff; **LiveKit stay** |
| Roadmap (pointer) | `.planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md` § «Отложено: July Rebuild» | Эпики CAL-SYNC / QA-HARDENING / Redesign; волны 1–11 |
| Gray zones | `.planning/decisions/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` | Locked решения (в т.ч. calendar iCal без OAuth) |

### Seeds

| Seed | Путь | Статус |
|------|------|--------|
| SEED-001 mediasoup / SFU | `.planning/reference/seeds/SEED-001-mediasoup-migration.md` | **REJECTED** — остаёмся на LiveKit |
| SEED-002 calendar + RSVP + mail | `.planning/reference/seeds/SEED-002-calendar-calls-email-sync.md` | **Active** — триггер конец июля |

---

## QA-HARDENING (planning wave 2026-07-10)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Code Quality Charter | `.planning/archive/july-2026/CODE-QUALITY-CHARTER.md` | Correctness = maintainability; DoD be/fe; антипаттерны |
| Edge / corner catalog | `.planning/archive/july-2026/EDGE-CORNER-CATALOG.md` | База + индекс · **≈4696 ECC** |
| Edge expansions 01–05 | `.planning/reference/edge-expansions/01…05-*.md` | Wave1 **+427 Edge** |
| Corner parent + C01–C05 | `.planning/reference/edge-expansions/00-CORNER-PARENT.md`, `corners/C01…C05` | Wave1 **+588 Corner** |
| Wave2 edge+corners | `.planning/reference/edge-expansions/wave2/` | **+771** |
| Wave3–5 | `.planning/reference/edge-expansions/wave{3,4,5}/` | **+2459** (781+848+830) |
| Testing Guide | `testframework/TESTING_GUIDE.md` | Канон автотестов: SOLID/DRY/KISS/POM |
| Guides 01–06 | `testframework/guides/` | Principles, POM, Go, Vitest, E2E, edge coverage |
---

## Команда и процесс (июльский контекст)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| Playbooks index | `.cursor/playbooks/README.md` | 32 agent files, dynamic models, RACI |
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
| Expansion / market research | `.planning/archive/july-2026/EXPANSION-ANALYSIS-2026-07-02.md`, `.planning/research/MARKET-AND-NEEDS-2026-07-02.md` | Контекст продукта |

---

## Ops / prod blockers (не July-feature, но в Pending)

| Артефакт | Путь | Назначение |
|----------|------|------------|
| FA-06 runbook | `docs/runbooks/fa06-secret-rotation.md` | Ручная ротация секретов |
| FA-06 execution log | `.planning/ops/production/FA-06-EXECUTION-LOG.md` | Phase 0 scan |
| Production handoff | `.planning/ops/production/PRODUCTION-HANDOFF.md` | Cutover checklist |

---

## Как пользоваться индексом

1. Старт сессии июля → `STATE.md` + этот INDEX.  
2. Rebuild / CAL-SYNC → brief + SEED-002 (SEED-001 не поднимать).  
3. Тесты / edge → Testing Guide + EDGE-CORNER-CATALOG (не копировать сюда).  
4. Код PR → CODE-QUALITY-CHARTER + playbooks + `*-ver` / code-reviewer.  
5. Мерж в `master` — **только пользователь**; Волна 1 merged пользователем 2026-07-11.

---

*Не переписывать содержимое соседних артефактов — только ссылки.*
