# Hub — AI Agent Workspace

> **Читай первым.** Регламент: `.cursor/REGULATION.md` | База знаний: `docs/README.md`

---

## Проект

**Hub** — self-hosted **corporate super-app** для команд 10–50 человек: Task Tracker + Messenger + TMS + Voice Calls + Dashboards **+** интранет, календарь, CRM-lite, HR, учёт, no-code базы, AI hub.

**Roadmap v4:** `.planning/milestones/v4-super-app/SUMMARY.md` · Vision: `.planning/PLATFORM-VISION-v4.md`

Монорепо: `backend/` (Go, **82 Ent schemas**) + `frontend/` (React) + `cli/` + `deployments/` + `testframework/`

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
| `docs/tz_constraints.md` | TZ mandatory/forbidden |
| `docs/competitive_analysis_2026.md` | Конкуренты + Mighty/Bitrix/Lark |
| `docs/competitive_audit.md` | Best practices + backlog |
| `docs/user_needs_analysis_2026.md` | Персоны, JTBD, MoSCoW |

Планирование: `.planning/STATE.md`, `.planning/COVERAGE.md`, `.planning/ROADMAP.md`, `.planning/DISCUSS-ROADMAP-LOCKED.md`

**Active:** Wave C ✅ · COV **83.0%** · ветка `master` · HEAD `bbdaf16` · FA-06 manual — единственный prod-блокер

GSD + агенты: `.cursor/GSD-INTEGRATION.md`, `.cursor/REGULATION.md` §9–§11

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

`devops-engineer`, `debugger`, `refactor-specialist`, `docs-writer`

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

---

## Текущий статус (2026-06-23)

| Milestone | Статус |
|-----------|--------|
| v1.0 MVP | ✅ SHIPPED 2026-03-21 |
| v2.0 Hub Redesign | ✅ COMPLETE 2026-05-15 |
| v2.5 Trust | ✅ COMPLETE 2026-06-22 |
| v2.6 Automate + Arch + OBS | ✅ COMPLETE 2026-06-24 |
| v3.0 AI Native | ✅ COMPLETE 2026-06-22 |
| v4 Super-app | ✅ COMPLETE 2026-06-25 |
| Coverage wave (Ph 59–125) | ✅ COV **83.0%** |
| Test framework migration | ✅ Wave B COMPLETE (0 internal tests) |
| Final audit (FA-01–FA-12) | ✅ Wave C COMPLETE (кроме FA-06 manual) |
| Lighthouse CI (FB-06) | ✅ artifact job PR #27 (non-blocking) |

### Метрики качества

| Метрика | Значение |
|---------|----------|
| COV overall | **83.0%** |
| Security packages ≥80% | **6/6** ✅ |
| auth / ws / rbac / audit / calls | 82.4% / 87.7% / 84.7% / 80.5% / 80.2% |
| tracker / messenger / notification / tms / project | 82.4% / 82.3% / 82.5% / 82.0% / 82.0% |
| Product score | **93/100** |
| REST API | **286+** endpoints |
| Ent schemas | **82** |
| Backend tests (`testframework/`) | **488** · 0 in `internal/` |
| Vitest | **99/99** pass |
| Playwright E2E | **20** specs |
| qa-gate (FA-10) | build ✅ · hubtest ✅ · race ✅ · Biome ✅ · e2e-smoke **3/3** ✅ |

### Следующие шаги

| Gap | Статус |
|-----|--------|
| FA-01 audit + remediation backlog | ✅ closed 2026-06-22 |
| Test framework migration TF-M04–M08 | ✅ Wave B COMPLETE |
| Wave C final audit FA-01–FA-12 | ✅ COMPLETE (кроме FA-06 manual) |
| FA-10 e2e-smoke | ✅ 3/3 PASS |
| PG-03 Secret rotation (FA-06) | ⏳ manual ops — **единственный prod-блокер** |
| SSH deploy + v2.5.0 tag | ⏳ после FA-06 + CTO sign-off |
| NFR-16 WCAG full sign-off | ⚠️ login axe PASS |
| NFR-17 Lighthouse blocking gate | ⚠️ artifact only (PR #27) |

Полный аудит: `.planning/AUDIT-2026-06-23.md` · матрица: `.planning/COVERAGE.md`

---

## Workflow

```
1. Загрузить docs/*
2. ba → sa → arch (для новых фич)
3. be/fe → атомарная реализация + тесты
4. Docs sync (STATE + domain docs)
5. *-ver → верификация (verified: true в skill)
6. qa → регрессия (Docker: -tags hubtest)
7. /qa-gate → финальный чек
```

---

## Команды

```bash
# Dev
docker compose -f docker-compose.local.yml up -d

# Backend regression (local)
cd backend && go build -tags release ./... && go test -race -count=1 ./...

# Backend regression (Docker, canonical)
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./internal/... ./testframework/..."

# Frontend regression
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run

# Full gate
/qa-gate
```

---

*Обновлено: 2026-06-23 — Wave C complete, master bbdaf16, AUDIT-2026-06-23*
