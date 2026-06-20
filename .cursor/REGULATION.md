# Регламент AI-команды Hub

> Полный цикл разработки: 6 ролей + 6 персональных верификаторов.
> Ключевой принцип: единая база знаний в `docs/`, верификация каждого действия.

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

Дополнительно (не дублировать): `.planning/ROADMAP.md`, `.planning/STATE.md`, `docs/wiki/`.

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

### Вспомогательные (вне регламента)

| Агент | Назначение |
|-------|------------|
| `devops-engineer` | Docker, CI/CD, K3s |
| `debugger` | Root cause analysis |
| `refactor-specialist` | Массовые rename/move |

Legacy aliases: `business-analyst`→`ba`, `system-analyst`→`sa`, `architect`→`arch`, `backend-dev`→`be`, `frontend-dev`→`fe`, `qa-auto`→`qa`.

---

## 3. Цикл верификации

```
Агент → изменение в docs/ или коде
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
3. **Регрессия обязательна после КАЖДОГО изменения** — см. §8; без green `/qa-gate` артефакт не в baseline
4. **Полный контекст** — загрузка всех `docs/` перед задачей
5. **Атомарность** — ≤200–300 строк за итерацию
6. **Security** — проверка по security_guidelines.md
7. **Понятное описание** — 2–5 предложений к каждому артефакту

## 8. Обязательная регрессия после каждого изменения

Любое изменение кода, конфигурации или docs, затрагивающее поведение — **сразу** прогон полного регрессионного набора:

```bash
# Backend (обязательно)
cd backend && go build ./... && go vet ./... && go test -race -count=1 -timeout=120s ./...

# Frontend (обязательно при изменениях frontend/)
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run

# Security scan diff (обязательно)
git diff --staged | findstr /I "secret password devkey token DEV_MODE"
```

Или slash-команда: **`/qa-gate`** — останавливается на первой FAIL.

**Правило для агентов (BE, FE, QA, devops):**
1. Сделал атомарное изменение
2. Запустил регрессию (минимум затронутый пакет + полный gate перед handoff)
3. Записал результат в skill-лог (`PASS` / `FAIL`)
4. Только после green — передача `*-ver`

**Правило для верификаторов:** не ставить `verified: true`, если регрессия не зелёная или не приложен лог прогона.

---

## 5. Workflow типовой задачи

```
1. BA  → vision_and_scope (новая фича)
2. SA  → system_specification
3. ARCH → architecture (если нужно)
4. BE/FE → код + тесты (атомарно)
5. **Docs sync** → обновить docs по `.cursor/rules/55-docs-sync.mdc` (STATE + domain docs)
6. QA  → test_plan, регрессия
7. *-ver → верификация каждого шага
8. /qa-gate → финальный чек
```

Hotfix (v2.0.1): `be`/`fe` → `be-ver`/`fe-ver` → `/qa-gate`

---

## 6. Slash-команды

| Команда | Назначение |
|---------|------------|
| `/qa-gate` | build + test + lint + security scan |
| `/coverage-audit` | матрица требований × код × тесты |
| `/handoff-sync` | обновить STATE.md, ROADMAP.md, COVERAGE.md + docs matrix (rule 55) |
| `/style-check` | anti-AI + deprecated + a11y |
| `/test-triage` | протокол: функционал → код vs тест (см. skills/test-triage) |

---

## 9. GSD Skills Integration

Hub использует **GSD** (Get Shit Done) skills из `~/.claude/skills/gsd-*/` параллельно с Hub slash-skills.

**Полная матрица:** `.cursor/GSD-INTEGRATION.md`

| Ситуация | GSD Skill |
|----------|-----------|
| Декомпозиция фазы | `gsd-plan-phase`, `gsd-execute-phase` |
| Мелкий фикс | `gsd-quick`, `gsd-fast` |
| Добавить тесты | `gsd-add-tests` |
| Падают тесты / баг | `gsd-debug` + `/test-triage` |
| Аудит → фикс | `gsd-audit-fix` |
| Верификация фазы | `gsd-verify-work`, `gsd-validate-phase` |
| Не знаю команду | `gsd-do` |
| Статус / handoff | `gsd-progress`, `/handoff-sync` |

Каждый агент в `.cursor/agents/*.md` и `.cursor/project_skills/*_skill.md` содержит свои GSD-навыки.

---

## 10. Test Triage (обязательно при FAIL)

**Перед исправлением упавшего теста или «бага из теста»:**

1. Воспроизвести **функционал** по `docs/` (не только assert в тесте)
2. Определить: код неверен **или** тест устарел/неверен
3. Чинить **минимальный** слой (код или тест — не оба «наугад»)
4. Записать triage в skill-лог: `ROOT=code|test`, файл, обоснование

Skill: `.cursor/skills/test-triage/SKILL.md` — обязателен для `be`, `fe`, `qa`, `debugger`, `*-ver`.

**Запрещено:** делать тесты зелёными без проверки соответствия спецификации.

---

## 11. Documentation Sync (обязательно)

После **любого** функционального изменения (код, API, CI, UX) агент **в той же сессии** обновляет затронутую документацию.

**Правило:** `.cursor/rules/55-docs-sync.mdc` (alwaysApply)

**Минимум:** `.planning/STATE.md` + доменный doc из матрицы (API → `system_specification` + `wiki/API.md`; UI → `SHORTCUTS`; deploy → `stack_and_tools` + `DEPLOY.md`; tests → `COVERAGE.md`).

**Handoff:** `/handoff-sync` — полная синхронизация planning + wiki.

**Запрещено:** handoff/commit user-visible изменений без обновления docs.

---

*Hub AI Team Regulation — 2026-06-20 (docs-sync §11)*
