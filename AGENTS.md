# Hub — AI Agent Workspace

> **Читай первым.** Регламент: `.cursor/REGULATION.md` | База знаний: `docs/README.md`

---

## Проект

**Hub** — self-hosted **corporate super-app** для команд 10–50 человек: Task Tracker + Messenger + TMS + Voice Calls + Dashboards **+** интранет, календарь, CRM-lite, HR, учёт, no-code базы, AI hub.

**Roadmap v4:** `.planning/milestones/v4-super-app/SUMMARY.md` · Vision: `.planning/specs/PLATFORM-VISION-v4.md`

Монорепо: `backend/` (Go, **82 Ent schemas**) + `frontend/` (React) + `cli/` + `deployments/` + **`testframework/`** (E2E + Vitest)

### Расположение тестов (pipeline-aligned)

| CI Job | Стадия | Путь | Домены |
|--------|--------|------|--------|
| `backend-test` | unit | `backend/testframework/{domain}/` | auth, tracker, tms, … |
| `frontend-test` | unit | `testframework/stages/unit/frontend-test/domains/{domain}/` | auth, admin, ai, v4, … |
| `e2e-smoke` | e2e | `testframework/stages/e2e/domains/` | auth, platform (3 spec) |
| `e2e-full` | e2e | `testframework/stages/e2e/domains/` | все 9 доменов, 20 spec |
| `coverage-gate` | quality | `backend/testframework/` | ≥82% |

Манифесты: `testframework/pipeline/jobs/*.yaml` · Карта: `testframework/PIPELINE.md`

Перед Vitest/E2E: `node frontend/scripts/link-testframework-node-modules.mjs`

---

## Маршрутизация контекста (`docs/`)

Всегда начинай с `AGENTS.md`, `.planning/STATE.md` и `docs/README.md`.
Затем загружай **только документы, влияющие на текущую задачу**, по таблице
ниже. Полный набор нужен для cross-domain аудита/архитектурной приёмки, но
вреден для точечной правки: раздувает контекст и увеличивает drift.

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
| `docs/GLOSSARY.md` | Термины, ID-конвенции (TC-/FR-/GZ-/ECC-…), модули backend |
| `docs/TROUBLESHOOTING.md` | Типовые грабли разработки (сборка, тесты, окружение) |
| `docs/runbooks/README.md` | Эксплуатация прода: инцидент, алерты, деплой/откат, БД, бэкапы |

Планирование: [`.planning/README.md`](.planning/README.md) (навигация) · `.planning/STATE.md` · `.planning/COVERAGE.md` · `.planning/ROADMAP.md` · `.planning/roadmaps/DISCUSS-ROADMAP-LOCKED.md`

**Active (2026-07-26):** Test Automation AI Quality Layer
`0.1.0-alpha.1` — structural PASS, paired effectiveness/ROI **NOT_RUN**;
[ai-expansion](.planning/july-2026/ai-expansion/README.md) qualified · July
Rebuild product-code freeze · LiveKit stay · FA-06 manual.

GSD + агенты: `.cursor/GSD-INTEGRATION.md`, `.cursor/REGULATION.md` §8 (GSD), §9 (test triage), §10 (docs sync)

---

## AI-команда (32 файла агентов)

> Реорганизована 2026-07-02 под параллельную разработку. Ростер: **`.cursor/playbooks/README.md`** · RACI: `.cursor/playbooks/team-workflow.md` · модели: **`.planning/decisions/ADR-MODELS-GPT56-2026.md`**.
> **Модели (ADR-MODELS-GPT56, amended 2026-07-26):** `team-lead` — **GPT-5.6 Sol Pro** (`model: gpt-5.6-sol`; Pro via catalog). Все downstream agents (включая `*-ver` и `codex-micro`) — `model: inherit`; orchestrator выбирает runtime по capability/cost/risk/context. `team-lead-opus` — deprecated redirect, не escalation slot.

### Лидерство / продукт
`team-lead` (оркестратор, Sol Pro), `tech-lead` (декомпозиция), `product` + `product-ver`, `codex-micro` (мелочь)

### Аналитика / архитектура / дизайн (потоки + верификаторы)

| Роль | Агенты | Verifier |
|------|--------|----------|
| Business Analyst | `ba1`, `ba2` | `ba-ver` |
| System Analyst | `sa1`, `sa2` | `sa-ver` |
| Architect | `arch` | `arch-ver` |
| Designer | `designer` | `designer-ver` |

### Разработка / QA (потоки + кросс-ревью + верификаторы)

| Роль | Агенты | Verifier |
|------|--------|----------|
| Backend | `be1`, `be2` | `be-ver` |
| Frontend | `fe1`, `fe2` | `fe-ver` |
| Manual QA | `qa-manual` | `qa-manual-ver` |
| Auto QA | `qa-auto1`, `qa-auto2` | `qa-auto-ver` |

### Cross-cutting

`devops-engineer`, `security-auditor`, `debugger`, `code-reviewer`, `docs-writer`

### Practices playbooks (`.cursor/playbooks/`)

`backend.md` · `frontend.md` · `design.md` · `testing.md` (Allure + замер времени + MD↔автотесты) · `observability.md` · `code-comments.md` · `git-workflow.md` · `analysis.md` · `verification.md` · `team-workflow.md` · **`prompt-contract.md`** (контракт промпта: приём задачи, формат, техники, температура, 10-этапная верификация, матрица ролей — обязателен всем)

---

## GSD Integration

Hub использует **GSD skills** (`~/.claude/skills/gsd-*`) + Hub slash-skills (`.cursor/skills/`).

| Документ | Назначение |
|----------|------------|
| `.cursor/GSD-INTEGRATION.md` | Матрица агент → GSD skills |
| `.cursor/skills/gsd-index/SKILL.md` | Быстрый индекс команд |
| `.cursor/skills/test-triage/SKILL.md` | FAIL теста: функционал → код vs тест |

**Test triage (обязательно):** при падении теста сначала проверить корректность функционала по `docs/`, только потом чинить код или тест. REGULATION §9 (детальные сценарии — §14, §18).

**Docs sync (обязательно):** после любого изменения кода обновляй затронутую документацию в той же сессии. Правило: `.cursor/rules/55-docs-sync.mdc` · skill: `/handoff-sync`.

---

## Текущий статус (2026-07-26)

> July Rebuild planning · LiveKit канон · метрики ниже = baseline master (без регрессии).

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
| Волна 1 (roadmap 2026-07-03) | ✅ merged to `master` by user 2026-07-11 |
| July Rebuild / QA-HARDENING | 📋 planning · kickoff ~конец июля |
| AI quality layer / GSD / portable kit | Structural PASS · Hub profile complete · assisted smoke PARTIAL · paired eval NOT_RUN |

### Метрики качества

| Метрика | Значение |
|---------|----------|
| COV overall | **83.0%** |
| Security packages ≥80% | **6/6** ✅ |
| auth / ws / rbac / audit / calls | 82.4% / 87.7% / 84.7% / 80.5% / 80.2% |
| tracker / messenger / notification / tms / project | 82.4% / 82.3% / 82.5% / 82.0% / 82.0% |
| Product score | **89/100** |
| REST API | **286+** endpoints |
| Ent schemas | **82** |
| Backend tests (`testframework/`) | **499** files · 0 in `internal/` (filesystem count 2026-07-26) |
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
| Merge Волна 1 → master | ✅ user merged 2026-07-11 |
| July return / CAL-SYNC / QA-HARDENING / Redesign | 📋 planning → kickoff ~конец июля |
| AI-layer paired pilot / physical empty-repo installation | ⏳ pre-release acceptance; package build/validation PASS |
| TF-CLEANUP-01 legacy e2e/specs | ✅ closed 2026-06-22 |
| RU-COMMENTS-01 top files | ✅ closed 2026-06-22 |
| DOC-02 historical SUMMARY banner | ✅ closed 2026-06-22 |
| SSH deploy + v2.5.0 tag | ⏳ после FA-06 + CTO sign-off |
| NFR-16 WCAG full sign-off | ⚠️ login axe PASS |
| NFR-17 Lighthouse blocking gate | ⚠️ artifact only (PR #27) |

Полный аудит: `.planning/audits/2026-06/AUDIT-2026-06-23.md` · матрица: `.planning/COVERAGE.md` · July index: `.planning/july-2026/INDEX-JULY-2026.md`

---

## Workflow

```
1. Загрузить docs/*
2. ba → sa → arch (для новых фич)
3. be/fe → атомарная реализация + тесты
4. Цикл качества нового + регрессия старого (rule 56) — до green
5. Docs sync (STATE + domain docs)
6. *-ver → верификация (verified: true в skill)
7. qa → регрессия (Docker: -tags hubtest)
8. /qa-gate → финальный чек
```

`team-lead` делегирует субагентам **только** по шаблону §5 (`team-lead.md` / `team-workflow.md` §5.3.1).
---

## Команды

```bash
# Dev
docker compose -f docker-compose.local.yml up -d

# Backend regression (local)
# ВАЖНО: -tags hubtest обязателен — export_hooks.go (39 файлов) под этим тегом,
# без него testframework не компилируется. Тестов в internal/ нет (все 499 в testframework/).
cd backend && go build -tags release ./... && go vet ./... \
  && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Backend regression (Docker, canonical)
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."

# Frontend regression
cd frontend && node scripts/link-testframework-node-modules.mjs && npx tsc --noEmit && npm run lint && npm run test:run

# Full gate
/qa-gate
```

---

*Обновлено: 2026-07-11 — team-lead шаблон делегирования + цикличное QA/регрессия*
