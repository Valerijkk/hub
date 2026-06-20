# Hub — AI Agent Workspace

> **Читай первым.** Регламент: `.cursor/REGULATION.md` | База знаний: `docs/README.md`

---

## Проект

**Hub** — единая платформа для команд 10–50 человек: Task Tracker + Messenger + TMS + Voice Calls + Dashboards.

Монорепо: `backend/` (Go) + `frontend/` (React) + `cli/` + `deployments/`

---

## Единая база знаний (`docs/`)

Перед любой задачей загрузи **все** файлы:

| Файл | Когда нужен |
|------|-------------|
| `docs/vision_and_scope.md` | Бизнес-контекст, User Stories |
| `docs/system_specification.md` | API, модели данных |
| `docs/architecture.md` | Модули, потоки, deployment |
| `docs/test_plan.md` | Тесты, регрессия |
| `docs/stack_and_tools.md` | Версии, инструменты |
| `docs/code_style_guide.md` | Стиль Go/TS |
| `docs/security_guidelines.md` | OWASP, blockers |
| `docs/competitive_audit.md` | Best practices, backlog |

Планирование: `.planning/STATE.md`, `.planning/ROADMAP.md`, `.planning/PLAN-CLOSE-GAPS.md`

GSD + агенты: `.cursor/GSD-INTEGRATION.md`, `.cursor/REGULATION.md` §9–§10

---

## AI-команда (12 агентов)

### Основные роли → Верификаторы

| Роль | Agent | Skill | Verifier |
|------|-------|-------|----------|
| Business Analyst | `ba` | `BA_skill.md` | `ba-ver` |
| System Analyst | `sa` | `SA_skill.md` | `sa-ver` |
| Architect | `arch` | `ARCH_skill.md` | `arch-ver` |
| Backend Dev | `be` | `BE_skill.md` | `be-ver` |
| Frontend Dev | `fe` | `FE_skill.md` | `fe-ver` |
| QA | `qa` | `QA_skill.md` | `qa-ver` |

Skills: `.cursor/project_skills/`

### Вспомогательные

`devops-engineer`, `debugger`, `refactor-specialist`

Legacy aliases: `business-analyst`→`ba`, `backend-dev`→`be`, `frontend-dev`→`fe`, `qa-auto`→`qa`

---

## GSD Integration

Hub использует **GSD skills** (`~/.claude/skills/gsd-*`) + Hub slash-skills (`.cursor/skills/`).

| Документ | Назначение |
|----------|------------|
| `.cursor/GSD-INTEGRATION.md` | Матрица агент → GSD skills |
| `.cursor/skills/gsd-index/SKILL.md` | Быстрый индекс команд |
| `.cursor/skills/test-triage/SKILL.md` | FAIL теста: функционал → код vs тест |

**Test triage (обязательно):** при падении теста сначала проверить корректность функционала по `docs/`, только потом чинить код или тест. REGULATION §10.

**Docs sync (обязательно):** после любого изменения кода обновляй затронутую документацию в той же сессии. Правило: `.cursor/rules/55-docs-sync.mdc` · skill: `/handoff-sync`.

Текущий план закрытия пробелов: `.planning/PLAN-CLOSE-GAPS.md`

---

## Текущий статус

- **v1.0** SHIPPED 2026-03-21
- **v2.0** Hub Redesign COMPLETE 2026-05-15
- **v2.0.1** hotfix — **12 CRITICAL закрыты в коде** (см. `docs/security_guidelines.md`)
- **v2.0.2** **SHIP-READY** — score **93/100** (2026-06-20, uncommitted)
- **PLAN-CLOSE-GAPS** Waves 1–6 ✅ · cleanup cycle 12 ✅
- Production ready: **✅ v2.0.2 gate** — COV-01 PASS; PG-03 secrets manual; v2.6 = PROD-01

### Оставшиеся gaps (v2.6 only)

| Gap | Статус |
|-----|--------|
| PROD-01 workflow automation | Spike ready — `.planning/PROD-01-SPIKE.md` |
| Vitest / Playwright | **49 tests** / **7 e2e specs** + CI ✅ |
| Lighthouse ≥90 | manual — `npm run lighthouse:help` |
| Secret rotation (git history) | PG-03 manual ops |
| Biome lint baseline | not gated |
| RHF+Zod forms | v2.6 backlog |

Полный список: `.planning/COVERAGE.md` · задачи: `.planning/TASKS-CLOSE-GAPS.md`

---

## Workflow

```
1. Загрузить docs/*
2. ba → sa → arch (для новых фич)
3. be/fe → атомарная реализация + тесты
4. *-ver → верификация (verified: true в skill)
5. qa → регрессия
6. /qa-gate → финальный чек
```

---

## Команды

```bash
# Dev
docker compose -f docker-compose.local.yml up -d

# Backend regression
cd backend && go build ./... && go test -race -count=1 ./...

# Frontend regression
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run

# Full gate
/qa-gate
```

---

*2026-06-17 — Реструктуризация под регламент AI-команды*
