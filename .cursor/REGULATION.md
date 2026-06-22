# Регламент AI-команды Hub

> Полный цикл разработки: 6 ролей + 6 персональных верификаторов + вспомогательные агенты.  
> **Ключевой принцип:** единая база знаний в `docs/`, верификация каждого действия, docs sync после каждого изменения.

*Обновлено: 2026-06-22*

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

### Основные роли

| Агент | Skill | Cursor agent | Модель |
|-------|-------|--------------|--------|
| BA | `project_skills/BA_skill.md` | `ba` | sonnet (readonly) |
| SA | `project_skills/SA_skill.md` | `sa` | sonnet (readonly) |
| ARCH | `project_skills/ARCH_skill.md` | `arch` | sonnet (readonly) |
| BE | `project_skills/BE_skill.md` | `be` | composer |
| FE | `project_skills/FE_skill.md` | `fe` | composer |
| QA | `project_skills/QA_skill.md` | `qa` | composer |

### Верификаторы

| Агент | Skill | Cursor agent |
|-------|-------|--------------|
| BA_VER | `BA_VER_skill.md` | `ba-ver` |
| SA_VER | `SA_VER_skill.md` | `sa-ver` |
| ARCH_VER | `ARCH_VER_skill.md` | `arch-ver` |
| BE_VER | `BE_VER_skill.md` | `be-ver` |
| FE_VER | `FE_VER_skill.md` | `fe-ver` |
| QA_VER | `QA_VER_skill.md` | `qa-ver` |

### Вспомогательные (вне основного цикла)

| Агент | Назначение |
|-------|------------|
| `devops-engineer` | Docker, CI/CD, K3s, compose |
| `debugger` | Root cause analysis |
| `refactor-specialist` | Массовые rename/move |
| `docs-writer` | README, wiki, Swagger, phase docs |

Legacy aliases: `business-analyst`→`ba`, `system-analyst`→`sa`, `architect`→`arch`, `backend-dev`→`be`, `frontend-dev`→`fe`, `qa-auto`→`qa`.

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
3. **Регрессия обязательна** — см. §8; без green `/qa-gate` артефакт не в baseline
4. **Полный контекст** — загрузка всех `docs/` перед задачей
5. **Атомарность** — ≤200–300 строк за итерацию
6. **Security** — проверка по `security_guidelines.md` и `tz_constraints.md`
7. **Понятное описание** — 2–5 предложений к каждому артефакту
8. **Docker для тестов** — канонический прогон: `-tags hubtest` в `deployments/docker-compose.test.yml`

---

## 5. Workflow типовой задачи

```
1. BA  → vision_and_scope (новая фича)
2. SA  → system_specification
3. ARCH → architecture (если нужно)
4. BE/FE → код + тесты (атомарно)
5. Docs sync → .cursor/rules/55-docs-sync.mdc (STATE + domain docs)
6. QA  → test_plan, регрессия
7. *-ver → верификация каждого шага
8. /qa-gate → финальный чек
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

## 7. Обязательная регрессия (§8)

Любое изменение кода, конфигурации или docs, затрагивающее поведение — **сразу** прогон регрессии:

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

Или slash-команда: **`/qa-gate`** — останавливается на первой FAIL.

**Правило для агентов:** сделал изменение → регрессия → skill-лог PASS/FAIL → только после green — передача `*-ver`.

**Правило для верификаторов:** не ставить `verified: true` без green regression + лог прогона.

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

*Обновлено: 2026-06-22*

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
