# Регламент AI-команды Hub

> Полный цикл разработки: **30 агентов** (лидерство/продукт, аналитика/архитектура/дизайн с верификаторами, backend/frontend/QA в параллельных потоках с кросс-ревью и верификаторами, cross-cutting специалисты). Ростер: `.cursor/playbooks/README.md`.  
> **Ключевой принцип:** единая база знаний в `docs/` + practices-playbooks в `.cursor/playbooks/`, верификация каждого действия, docs sync после каждого изменения.

*Обновлено: 2026-07-11*

---

## 1. Единая база знаний (`docs/`)

Каждый агент **перед задачей** загружает все общие файлы. Дублирование в skill-файлах запрещено.

| Файл | Содержание | Владелец |
|------|------------|----------|
| `docs/vision_and_scope.md` | SMART-цели, KPI, персоны, User Stories, MoSCoW | BA |
| `docs/system_specification.md` | FR/NFR, ERD, API, DFD | SA |
| `docs/architecture.md` | C4, deployment, patterns, security architecture | ARCH |
| `docs/test_plan.md` | Пирамида, критерии, матрица, регрессия | QA |
| `docs/stack_and_tools.md` | Стек с версиями, CI/CD, линтеры | ARCH |
| `docs/code_style_guide.md` | Именование, форматирование, запреты | ARCH |
| `docs/security_guidelines.md` | OWASP, секреты, auth, чек-лист | ARCH/BE |
| `docs/tz_constraints.md` | Mandatory/forbidden (tracked ADR) | ARCH |
| `docs/competitive_analysis_2026.md` | Конкуренты, TCO, gaps | BA |
| `docs/competitive_audit.md` | Best practices, feature backlog | BA |
| `docs/user_needs_analysis_2026.md` | Персоны, JTBD, MoSCoW | BA |

Дополнительно (не дублировать): `.planning/STATE.md`, `.planning/COVERAGE.md`, `.planning/ROADMAP.md`, `docs/wiki/`.

---

## 2. Команда агентов

> **2026-07-02.** Команда реорганизована под параллельную разработку: **30 агентов**. Полный ростер, модели и RACI — `.cursor/playbooks/README.md` и `.cursor/playbooks/team-workflow.md`. Практики ремесла вынесены в `.cursor/playbooks/*` (файлы агентов ссылаются, не дублируют). Прежние `project_skills/*_skill.md` удалены (миграция: `.cursor/project_skills/README.md`).

**Модели:** `team-lead` — Fable 5 (thinking-high) дефолт; `team-lead-opus` — Opus 4.8 @ 1M эскалация (тяжёлые декомпозиции/приёмка). Все остальные — Sonnet 5, thinking-high (max effort), ~200–300k. (Fable выбран дефолтом по итогам 5 A/B-задач.)

### Лидерство и продукт

| Агент | Cursor agent | Роль |
|-------|--------------|------|
| Team Lead | `team-lead` | Оркестрация волн, делегирование, финальная приёмка |
| Tech Lead | `tech-lead` | Декомпозиция на фазы/планы, зависимости |
| Product | `product` | Приоритеты, scope, KPI |
| Product Ver | `product-ver` | Соответствие vision/KPI/scope |

### Аналитика, архитектура, дизайн (параллельные потоки + верификаторы)

| Роль | Агенты | Верификатор |
|------|--------|-------------|
| Business Analyst | `ba1`, `ba2` | `ba-ver` |
| System Analyst | `sa1`, `sa2` | `sa-ver` |
| Architect | `arch` | `arch-ver` |
| Designer | `designer` | `designer-ver` |

### Разработка и QA (параллельные потоки + кросс-ревью + верификаторы)

| Роль | Агенты | Верификатор |
|------|--------|-------------|
| Backend | `be1`, `be2` (кросс-ревью) | `be-ver` |
| Frontend | `fe1`, `fe2` (кросс-ревью) | `fe-ver` |
| Manual QA | `qa-manual` | `qa-manual-ver` |
| Auto QA | `qa-auto1`, `qa-auto2` (кросс-ревью) | `qa-auto-ver` |

### Cross-cutting специалисты

| Агент | Назначение |
|-------|------------|
| `devops-engineer` | Docker, CI/CD, K3s, Caddy, compose |
| `security-auditor` | OWASP/ASVS, security-инварианты TZ |
| `debugger` | Root-cause анализ |
| `code-reviewer` | Pre-merge diff review |
| `docs-writer` | docs/wiki sync, Swagger |

**Разделение параллельной работы (be1/be2, fe1/fe2, ba1/ba2, sa1/sa2, qa-auto1/qa-auto2):** разные модули/файлы в волне, общие контракты фиксируют `sa`/`arch` заранее, частый rebase, кросс-ревью парного агента перед персональным верификатором. Детали — `.cursor/playbooks/team-workflow.md` §4 и `git-workflow.md`.

---

## 3. Цикл верификации

```
Агент → изменение в docs/ или коде
     → docs sync (STATE + domain docs)
     → запись в skill-лог
     → описание простым языком (2–5 предложений)
     → персональный верификатор
     → замечания? → исправление → повтор
     → verified: true, YYYY-MM-DD HH:MM
     → baseline проекта
```

---

## 4. Принципы качества

1. **Контекст задачи** — верификатор воспроизводит сценарий из требований
2. **Одна нейросеть проверяет другую** — верификатор независим
3. **Качество циклично** — новое поведение: test→fix→re-test до green; старый функционал — регрессия тем же циклом; без green `/qa-gate` артефакт не в baseline (см. §7, rule 56)
4. **Полный контекст** — загрузка всех `docs/` перед задачей
5. **Атомарность** — ≤200–300 строк за итерацию
6. **Security** — проверка по `security_guidelines.md` и `tz_constraints.md`
7. **Понятное описание** — 2–5 предложений к каждому артефакту
8. **Docker для тестов** — канонический прогон: `-tags hubtest` в `deployments/docker-compose.test.yml`
9. **Делегирование team-lead** — только по шаблону `.cursor/agents/team-lead.md` §5 / `team-workflow.md` §5.3.1

---

## 5. Workflow типовой задачи

```
1. BA  → vision_and_scope (новая фича) → ba-ver
2. SA  → system_specification → sa-ver
3. ARCH → architecture (если нужно) → arch-ver
4. BE/FE → код + UNIT-тесты (атомарно; юниты — обязанность разработчика) → be-ver/fe-ver
5. QA-конвейер (строго последовательно, решение 2026-07-03):
   qa-manual (TC из бизнес-доки) → qa-manual-ver (верификация тестовой доки)
   → qa-auto1/2 (автотесты по верифицированным TC) → qa-auto-ver
6. Docs sync → .cursor/rules/55-docs-sync.mdc (STATE + domain docs)
7. code-reviewer → team-lead приёмка
8. /qa-gate → финальный чек (на интеграции волны)
```

Hotfix: `be`/`fe` → `be-ver`/`fe-ver` → `/qa-gate`

---

## 6. Slash-команды

| Команда | Назначение |
|---------|------------|
| `/qa-gate` | build + test + lint + security scan |
| `/coverage-audit` | матрица требований × код × тесты |
| `/handoff-sync` | STATE + ROADMAP + COVERAGE + docs matrix |
| `/style-check` | anti-AI + deprecated + a11y |
| `/test-triage` | протокол: функционал → код vs тест |

---

## 7. Обязательная регрессия и цикличное качество (§8)

Любое изменение кода, конфигурации или docs, затрагивающее поведение — **сразу** два циклических контура:

1. **Новая фича / новое поведение** — проверить → FAIL? → triage → fix → повтор до green.
2. **Старый функционал (регрессия)** — домен/смежное → FAIL? → triage → fix → повтор; на интеграции волны — `/qa-gate`.

Одноразовой проверки недостаточно.

```bash
# Backend local
cd backend && go build -tags release ./... && go vet ./... && go test -race -count=1 -timeout=600s ./...

# Backend canonical (Docker)
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."

# Frontend (при изменениях frontend/)
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

Или slash-команда: **`/qa-gate`** — останавливается на первой FAIL; после фикса — повтор (цикл).

**Правило для агентов:** сделал изменение → цикл качества нового → цикл регрессии старого → skill-лог PASS/FAIL → только после green — передача `*-ver`.

**Правило для верификаторов:** не ставить `verified: true` без green обоих циклов + лог прогона.

**Правило для team-lead:** делегирование субагентам — только шаблон §5 (`team-lead.md` / `team-workflow.md` §5.3.1), с явным требованием обоих циклов.

---

## 8. GSD Skills Integration (§9)

Hub использует **GSD** skills из `~/.claude/skills/gsd-*/` + Hub slash-skills.

**Матрица:** `.cursor/GSD-INTEGRATION.md`

| Ситуация | GSD Skill |
|----------|-----------|
| Декомпозиция фазы | `gsd-plan-phase`, `gsd-execute-phase` |
| Мелкий фикс | `gsd-quick`, `gsd-fast` |
| Добавить тесты | `gsd-add-tests` |
| Падают тесты / баг | `gsd-debug` + `/test-triage` |
| Аудит → фикс | `gsd-audit-fix` |
| Верификация фазы | `gsd-verify-work`, `gsd-validate-phase` |
| Docs bulk update | `gsd-docs-update` |
| Статус / handoff | `gsd-progress`, `/handoff-sync` |

---

## 9. Test Triage (§10)

**Перед исправлением упавшего теста:**

1. Воспроизвести **функционал** по `docs/` (не только assert)
2. Определить: код неверен **или** тест устарел
3. Чинить **минимальный** слой
4. Записать triage: `ROOT=code|test`, файл, обоснование

Skill: `.cursor/skills/test-triage/SKILL.md`

**Запрещено:** делать тесты зелёными без проверки спецификации.

---

## 10. Documentation Sync (§11)

После **любого** функционального изменения агент **в той же сессии** обновляет docs.

**Правило:** `.cursor/rules/55-docs-sync.mdc` (alwaysApply)

**Минимум:** `.planning/STATE.md` + доменный doc (API → `system_specification` + `wiki/API.md`; UI → `SHORTCUTS`; tests → `COVERAGE.md`; competitive → `competitive_*.md`).

**Handoff:** `/handoff-sync`

**Запрещено:** handoff/commit user-visible изменений без docs sync.

---

*Обновлено: 2026-07-11*

---

## 11. Текущий статус проекта (справка для агентов)

| Показатель | Значение (2026-06-22) |
|------------|----------------------|
| Milestone | v4 Super-app ✅ + Wave C audit ✅ |
| COV overall | **83.0%** |
| Security 6/6 | ✅ ≥80% |
| Product score | **87/100** (`.planning/AUDIT-SCORECARD.md`) |
| Active work | FA-06 manual ops · TF-CLEANUP-01 |
| Branch | `master` |
| Blockers | PG-03 rotation (manual) |

---

*Hub AI Team Regulation — 2026-06-22*
