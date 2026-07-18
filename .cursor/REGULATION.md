# Регламент AI-команды Hub

> Полный цикл разработки: **30 агентов** (лидерство/продукт, аналитика/архитектура/дизайн с верификаторами, backend/frontend/QA в параллельных потоках с кросс-ревью и верификаторами, cross-cutting специалисты). Ростер: `.cursor/playbooks/README.md`.  
> **Ключевой принцип:** единая база знаний в `docs/` + practices-playbooks в `.cursor/playbooks/`, верификация каждого действия, docs sync после каждого изменения.

*Обновлено: 2026-07-18*

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

**Модели:** с **августа 2026** — `team-lead` = **GPT-5.6 Sol Pro**; все остальные роли = **GPT-5.6 Sol**; micro = **Codex (latest)** (`codex-micro`). До cutover — Claude: Fable 5 / Opus 4.8 escalate / Sonnet 5. Канон: `.planning/decisions/ADR-MODELS-GPT56-2026.md`.

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
| Product score | **87/100** (`.planning/audits/2026-06/AUDIT-SCORECARD.md`) |
| Active work | FA-06 manual ops · TF-CLEANUP-01 |
| Branch | `master` |
| Blockers | PG-03 rotation (manual) |

> **Актуализация 2026-07-18:** `.planning/` реорганизован — живой канон (`PROJECT.md`, `ROADMAP.md`, `STATE.md`, `REQUIREMENTS.md`, `COVERAGE.md`, `config.json`) остался в корне, ~42 разовых артефакта разложены по темам. Навигация: [`.planning/README.md`](../.planning/README.md). Все пути ниже (§12) — актуальные после реорганизации.

---

## 12. Карта путей `.planning/` (после реорганизации 2026-07-18)

> Полная навигация с описанием каждой папки: [`.planning/README.md`](../.planning/README.md). Правило: **ничего не удалять**, новые разовые отчёты — в тематическую папку по дате, не в корень.

### Живой канон (корень `.planning/`)

| Файл | Назначение |
|------|------------|
| `.planning/STATE.md` | Текущая позиция, session notes, Pending Todos — читать **первым** в сессии |
| `.planning/PROJECT.md` | Полный список требований проекта |
| `.planning/ROADMAP.md` | Дорожная карта фаз (GSD, исторический — 125 фаз) |
| `.planning/REQUIREMENTS.md` | Требования (FR/NFR ID) |
| `.planning/COVERAGE.md` | Матрица requirement × код × тест |
| `.planning/config.json` | GSD config |
| `.planning/README.md` | Индекс-навигация по всем тематическим папкам |

### Тематические папки

| Папка | Что внутри | Примеры файлов |
|-------|------------|-----------------|
| `.planning/july-2026/` | July Rebuild / QA-HARDENING | `INDEX-JULY-2026.md`, `JULY-2026-REBUILD-BRIEF.md`, `ROADMAP-FINAL-2026-07-03.md`, `ROADMAP-DETAILED-JULY-RETURN.md`, `PROJECT-STATUS-2026-07-10.md`, `CODE-QUALITY-CHARTER.md`, `EDGE-CORNER-CATALOG.md`, `EXPANSION-ANALYSIS-2026-07-02.md` |
| `.planning/audits/2026-06/` | Аудиты июня | `AUDIT-2026-06-23.md`, `AUDIT-2026-06-23-full.md`, `AUDIT-SCORECARD.md`, `FA-01-AUDIT.md`, `FA-02-SECURITY-REAUDIT.md`, `FA-10-QA-GATE.md`, `FA-RBAC-FIBER3-AUDIT.md`, `TECH-DEBT-AUDIT-2026-06-23.md`, `WAVE-C-VERIFICATION.md`, `RU-V-REPORT.md`, `AUDIT-2026-06-22-quality-processes.md` |
| `.planning/production/` | Prod ops | `PG-03-SECRET-ROTATION.md`, `PRODUCTION-HANDOFF.md`, `FA-06-EXECUTION-LOG.md`, `PROD-01-SPIKE.md`, `PR-feat-v2.5-trust.md` |
| `.planning/gray-zones/` | Серые зоны + зафиксированные решения | `GRAY-ZONES-DECISIONS-2026-07-03.md`, `AREA-quality.md`, `AREA-release-process.md`, `AREA-security.md`, `AREA-security-ops.md`, `AREA-v4-parity.md`, `AREA-arch-services.md` |
| `.planning/roadmaps/` | Исторические roadmap/backlog | `EXECUTIVE-REPORT-2026.md`, `FEATURE-BACKLOG-2026.md`, `IMPROVEMENT-MASTER-PLAN.md`, `MICROSERVICES-ROADMAP.md`, `MILESTONE-2.md`, `MILESTONE-3.md`, `DISCUSS-ROADMAP-LOCKED.md`, `DISCUSS-ROADMAP-2026-CONTEXT.md` |
| `.planning/specs/` | Технические спеки | `AI-PROVIDER-HUB-SPEC.md`, `NOTIF-TG-EMAIL-SPEC.md`, `PLATFORM-VISION-v4.md` |
| `.planning/tasks/` | Закрытие gaps / финальный аудит фаз | `TASKS-CLOSE-GAPS.md`, `PLAN-CLOSE-GAPS.md`, `TASKS-RU-COMMENTS-DOCS.md`, `TASKS-PHASES-121-FINAL-AUDIT.md` |
| `.planning/decisions/` | ADR | `ADR-AUTH-01-magic-link-only.md`, `ADR-AUTH-SSO-2026.md`, `ADR-TZ-REWRITE-2026.md`, `ADR-ARCH-MICROSERVICES-2026.md`, `ADR-AI-AGENTS-MCP-2026.md`, `ADR-NOTIF-01-transactional-email.md` |
| `.planning/seeds/` | Forward-looking идеи (GSD `gsd-plant-seed`) | `SEED-001-mediasoup-migration.md` (rejected), `SEED-002-calendar-calls-email-sync.md` (active) |
| `.planning/edge-expansions/` | Edge/corner каталог по волнам | `00-CORNER-PARENT.md`, `wave2..5/edge/E0{1-5}-*.md`, `wave2..5/corners/C0{2-5}-*.md` |
| `.planning/milestones/` | Архив milestone-фаз (v1.0–v4) | `v4-super-app/SUMMARY.md`, `v2.0-phases/*`, `v1.0-phases/*` |
| `.planning/phases/` | Актуальные phase-планы (GSD) | `16-v26-automate/PLAN.md`, `63-production-gate/PLAN.md`, `67-production-gate/PLAN.md` |
| `.planning/code-review-final/` | Итоговый code review v2.0.1 | `SUMMARY.md` (12 CRITICAL findings, статус) |
| `.planning/ab-tests/`, `.planning/ab-microservices/` | A/B планы двух моделей (fable/opus) | `PLAN-bases-fable.md` / `-opus.md`, `PLAN-schedule-*`, `PLAN-team-lead-*` |
| `.planning/research/` | Research notes | `MARKET-AND-NEEDS-2026-07-02.md` |

### Быстрые ссылки (частые запросы)

| Нужно | Путь |
|-------|------|
| Что сейчас происходит | `.planning/STATE.md` → `.planning/README.md` |
| Актуальный roadmap (не исторический) | `.planning/july-2026/ROADMAP-FINAL-2026-07-03.md` |
| Решения по серым зонам | `.planning/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` |
| Секреты / прод | `.planning/production/PG-03-SECRET-ROTATION.md`, `.planning/production/FA-06-EXECUTION-LOG.md` |
| Полный аудит проекта | `.planning/audits/2026-06/AUDIT-2026-06-23.md` |
| Coverage | `.planning/COVERAGE.md` |
| Vision v4 | `.planning/specs/PLATFORM-VISION-v4.md` |

---

## 13. Docs-sync — расширенная матрица с реальными путями

> Правило: `.cursor/rules/55-docs-sync.mdc`. Ниже — детализация «тип изменения → конкретный файл» на основе реальных модулей Hub.

| Тип изменения | Домен-пример | Обязательные документы |
|---|---|---|
| Новый API endpoint | `POST /api/tracker/issues/:id/automation` | `docs/system_specification.md` §API, `docs/wiki/API.md`, `.planning/STATE.md` |
| Новая ERD-сущность | `bases_view`, `calendar_event` | `docs/system_specification.md` §ERD, соответствующий ADR если архитектурное решение |
| Изменение auth-потока | Magic Link rate-limit, refresh rotation | `docs/security_guidelines.md`, `docs/tz_constraints.md` (если меняет mandatory), `.planning/STATE.md` |
| Новый UI flow / хоткей | Ctrl+K палитра, bases inline-edit | `docs/wiki/SHORTCUTS.md`, `docs/wiki/QUICKSTART.md`, `docs/vision_and_scope.md` (если новый US) |
| Новые тесты / coverage | `backend/testframework/tracker/service_wave121_test.go` | `.planning/COVERAGE.md`, `docs/test_plan.md`, `docs/TESTING.md` |
| CI/CD, Docker, K3s | `docker-compose.test.yml`, `.github/workflows/ci.yml` | `docs/stack_and_tools.md`, `deployments/.env.example`, `docs/wiki/DEPLOY.md` |
| Security fix | webhook signature, SSRF guard | `docs/security_guidelines.md` + запись в `.planning/audits/2026-06/AUDIT-SCORECARD.md` если закрывает finding |
| Закрытие gap/wave | Волна интеграции | `.planning/tasks/TASKS-CLOSE-GAPS.md`, `.planning/STATE.md` |
| Архитектурное решение | Микросервисы, SSO | `.planning/decisions/ADR-*.md` + `docs/architecture.md` |
| Best practice / backlog идея | Конкурентная фича | `docs/competitive_audit.md` |
| Статус проекта / milestone | Завершение фазы | `AGENTS.md` (кратко), `.planning/STATE.md`, при завершении milestone — `.planning/milestones/` |
| Roadmap-решение (in/out scope) | Feature toggle по фазам | `.cursor/rules/60-tz-constraints.mdc` §IN SCOPE, `docs/vision_and_scope.md` |

**Правило DoD:** ни один PR с user-visible или prod-behavior изменением не уходит на `code-reviewer`/team-lead без строки в `.planning/STATE.md` и обновлённого доменного doc из таблицы выше.

---

## 14. Test Triage — сценарии по реальным доменам Hub

> Полный протокол: `.cursor/skills/test-triage/SKILL.md`, `.cursor/GSD-INTEGRATION.md` §Test Triage. Ниже — конкретные примеры ROOT=code vs ROOT=test на действующих модулях `backend/internal/{auth,tracker,bases}`.

### Домен `auth` (`backend/internal/auth/`: `magiclink.go`, `service_login.go`, `service_session.go`, `middleware.go`, `telegram.go`)

| # | Симптом | Функционал по docs | Вердикт | Действие |
|---|---------|---------------------|---------|----------|
| A1 | `TestMagicLink_RateLimit_Exceeded` FAIL: получил 6-ю попытку без 429 | `docs/tz_constraints.md`: Rate limiting auth 3 попытки/10 мин (mandatory) | **ROOT=code** | Добавить счётчик в `service_login.go`, тест не трогать |
| A2 | `TestRefreshToken_TTL` FAIL: ожидает TTL=7 дней | `tz_constraints.md`: Refresh token Redis, 30 дней TTL | **ROOT=test** | Тест устарел (старое требование) — исправить ожидание на 30д |
| A3 | `TestTelegramBind_AsLogin` PASS, но по ADR-AUTH-01 Telegram — только notification bind, не login | Спецификация: `ADR-AUTH-01-magic-link-only.md` | **ROOT=code** (баг скрыт зелёным тестом!) | Тест недостаточно строгий — добавить negative-кейс + фикс `telegram.go` |
| A4 | `TestSessionMiddleware_ExpiredJWT` flaky (иногда PASS/FAIL) | JWT TTL 15 мин, зависит от `time.Now()` | **Flaky** | Заменить на fake clock в `middleware.go` тесте, не увеличивать TTL в проде |

### Домен `tracker` (`backend/internal/tracker/`: `handler_issues.go`, `automation_engine.go`, `repo_time.go`, `service_workflow.go`)

| # | Симптом | Функционал по docs | Вердикт | Действие |
|---|---------|---------------------|---------|----------|
| T1 | `TestAutomationEngine_TriggerOnStatusChange` FAIL: правило не срабатывает | `docs/system_specification.md` FR-TRACKER JQL automation | **ROOT=code** | `automation_engine.go` не подписан на событие `hub.tracker.issue.updated` — добавить consumer |
| T2 | `TestIssueHandler_Pagination` FAIL: ожидает cursor-пагинацию | `tz_constraints.md`: задачи — **offset/limit**, чат — cursor | **ROOT=test** | Тест написан по неверному паттерну — исправить на offset/limit assertions |
| T3 | `TestRepoTime_Overlap` FAIL: два time-log перекрываются без ошибки | Спецификация не запрещает overlap явно — уточнить у `sa` | **Stop → clarify** | Не чинить ни код ни тест, эскалировать в `sa1`/`sa2`, зафиксировать в `.planning/STATE.md` |
| T4 | `TestWorkflowTransition_SkipGuard` FAIL после рефакторинга `service_workflow.go` | Guard-функция переименована, тест дёргает старое имя мока | **ROOT=test** | Обновить mock/stub на новый интерфейс, поведение не менялось |

### Домен `bases` (no-code БД, `backend/internal/bases/`: `service.go`, `repository.go`)

| # | Симптом | Функционал по docs | Вердикт | Действие |
|---|---------|---------------------|---------|----------|
| B1 | `TestBasesRBAC_ViewerCannotEdit` FAIL: viewer смог изменить row | `docs/security_guidelines.md` + RBAC-инвариант GZ-SEC | **ROOT=code** | `service.go` не проверяет role перед `UpdateRow` — добавить guard |
| B2 | `TestBasesService_SoftDelete` FAIL: запись удалена hard | `tz_constraints.md`: Soft delete через корзину 30 дней, hard delete — только явный Admin/GDPR запрос | **ROOT=code** | `repository.go` использует `Delete()` вместо `SetDeletedAt()` — критичный фикс |
| B3 | `TestBasesService_OptimisticLock` FAIL: два конкурентных PATCH — второй не увидел конфликт | `tz_constraints.md`: Optimistic locking через `version` | **ROOT=code** | Добавить проверку `version` в `service.go` перед update, тест уже корректен |
| B4 | `TestBasesService_ColumnTypeCast` FAIL: assertion ждёт `error`, а код возвращает `nil, nil` при невалидном типе | Спецификация неочевидна: что должно быть при попытке привести `text` в `number` | **Stop → clarify** | Эскалация `sa`, обновить `docs/system_specification.md` перед фиксом |

**Вывод:** ни один из этих кейсов не решается «поменять assert, чтобы стало зелёным» — сначала матчинг с `docs/`, `tz_constraints.md` или ADR, потом классификация code/test/stop.

---

## 15. `/qa-gate` — детальный прогон с реальными командами

> Canonical backend regression — **Docker `-tags hubtest`** (не голый `go test ./...` — там 0 internal тестов, всё в `backend/testframework/`).

### Backend: build + vet + race (локально, быстрый цикл разработчика)

```bash
cd backend
go build -tags release ./...
go vet ./...
go test -race -count=1 -timeout=120s ./...
```

### Backend: canonical Docker hubtest (обязателен перед handoff/PR)

```bash
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."
```

Точечный прогон одного домена (быстрее полного gate при итерации на фиче):

```bash
# только tracker
go test -tags hubtest -race -count=1 ./testframework/tracker/...

# только auth
go test -tags hubtest -race -count=1 ./testframework/auth/...

# только bases
go test -tags hubtest -race -count=1 ./testframework/bases/...
```

### Frontend: tsc + lint + vitest

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
npm run lint
npx vitest run
```

Точечный домен (Vitest, пути реальные):

```bash
# auth-компоненты
npx vitest run testframework/stages/unit/frontend-test/domains/auth/

# tracker-компоненты
npx vitest run testframework/stages/unit/frontend-test/domains/tracker/

# v4-модули (bases, crm, hr, accounting)
npx vitest run testframework/stages/unit/frontend-test/domains/v4/
```

### E2E: smoke (CI на каждый PR) vs full (перед релизом/волной)

```bash
# e2e-smoke — 3 spec, домены auth + platform (быстрый гейт)
npx playwright test testframework/stages/e2e/domains/auth/login.spec.ts \
  testframework/stages/e2e/domains/auth/logout.spec.ts \
  testframework/stages/e2e/domains/platform/health.spec.ts

# e2e-full — 9 доменов, ~20 spec (перед merge волны / релизом)
npx playwright test testframework/stages/e2e/domains/
```

Реальные домены e2e: `auth` (login, logout, onboarding-profile), `tracker` (project-lifecycle), `tms` (tms), `messenger` (messenger), `ai` (ai-hub, automation), `admin` (admin-ai, admin-audit, admin-metrics), `navigation` (command-palette, shortcuts, navigation), `platform` (health, metrics, notifications, a11y, critical-gate), `v4` (v4-superapp).

### Security scan (git diff перед коммитом)

```bash
git diff --staged --name-only | xargs grep -l -E "(secret|password|devkey|api_key|token)" 2>/dev/null
git diff --staged | grep -E "DEV_MODE=true|devkey: secret|livekit_api_secret"
```

### Результат gate

```
✅ Backend build+vet+hubtest race — PASS
✅ Frontend tsc+lint+vitest — PASS
✅ e2e-smoke (3/3) — PASS
✅ Security scan — clean
❌ GATE FAILED at step N: <причина> → test-triage → fix → повтор с начала упавшего шага
```

Slash: `.cursor/skills/qa-gate/SKILL.md`.

---

## 16. Цикличное качество — сквозной пример (rule 56)

Пример полного цикла для фичи «WIP-лимиты в Kanban-колонках» (`tracker`):

```
1. BE: implement WipLimit в repo_boards.go + service_boards.go
2. Новое поведение:
   go test -tags hubtest -race -count=1 ./testframework/tracker/... -run TestBoards_WipLimit
   → FAIL: limit не применяется при drag-n-drop через API
   → test-triage: ROOT=code (guard не вызван в handler_boards.go)
   → fix → повтор → PASS
3. Регрессия домена tracker:
   go test -tags hubtest -race -count=1 ./testframework/tracker/...
   → FAIL: TestAutomationEngine_TriggerOnStatusChange (WIP limit блокирует переход, автоматизация ожидала успех)
   → test-triage: ROOT=test (сценарий теста не учитывал новый лимит — устарел)
   → fix теста → повтор → PASS
4. Смежный домен (dashboard читает board stats):
   go test -tags hubtest -race -count=1 ./testframework/dashboard/...
   → PASS (нет regression)
5. На интеграции волны: /qa-gate (полный, включая e2e-smoke)
6. Docs sync: system_specification.md (FR-TRACKER WIP), COVERAGE.md, STATE.md
7. be-ver: verified: true + лог обоих циклов (шаги 2–3)
```

**Анти-паттерн (запрещён):** «прогнал `go test ./testframework/tracker/... -run TestBoards_WipLimit`, зелёный → handoff» — это только контур 1 (новое), контур 2 (регрессия домена) пропущен.

---

## 17. GSD — когда какая команда (расширенная шпаргалка)

> Полная агент→skill матрица: `.cursor/GSD-INTEGRATION.md`. Ниже — быстрый выбор по ситуации.

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Не знаю с чего начать сессию | `gsd-do` | «Разберись с багом в bases RBAC» → роутится в `gsd-debug` |
| Хочу статус проекта | `gsd-progress` | Сверка с `.planning/STATE.md` перед новой задачей |
| Что делать следующим | `gsd-next` | После завершения фазы 121 |
| Декомпозировать крупную фичу | `gsd-plan-phase` → `gsd-execute-phase` | CAL-SYNC (email/calendar/calls) |
| Мелкий таргетированный фикс | `gsd-quick` | Добавить один RBAC-guard в `bases/service.go` |
| Тривиальная правка без оверхеда | `gsd-fast` | Опечатка в error message |
| Нужны тесты после фичи | `gsd-add-tests` | После WIP-лимитов — `TestBoards_WipLimit*` |
| Тест/прод падает | `gsd-debug` + `/test-triage` | Flaky `TestSessionMiddleware_ExpiredJWT` |
| Аудит → массовый фикс | `gsd-audit-fix` | После `AUDIT-2026-06-23-full.md` |
| Верификация фазы завершена | `gsd-verify-work`, `gsd-validate-phase` | Приёмка Wave C |
| Bulk docs update | `gsd-docs-update` | После миграции `.planning/` reorg (проверка ссылок) |
| Handoff в конце сессии | `gsd-progress` / `/handoff-sync` | Всегда перед завершением значимой сессии |
| Threat model проверка | `gsd-secure-phase` | После auth/RBAC изменений |
| UI-дизайн контракт | `gsd-ui-phase` → `gsd-ui-review` | Новый экран bases inline-edit |
| AI-фича (провайдер/агент) | `gsd-ai-integration-phase` | AI gateway BYOK + Ollama (фаза 17) |

---

## 18. Test Triage — дополнительные домены (messenger / tms / calls)

### Домен `messenger` (`backend/internal/messenger/`)

| # | Симптом | Функционал по docs | Вердикт | Действие |
|---|---------|---------------------|---------|----------|
| M1 | `TestChannelPagination_Cursor` FAIL: тест ждёт offset/limit | `tz_constraints.md`: чат — cursor (infinite scroll), задачи — offset/limit | **ROOT=test** | Тест написан по чужому паттерну (скопирован из tracker) — исправить на cursor |
| M2 | `TestPresence_WSRoomJoin_NoMembershipCheck` PASS, но должен FAIL | `tz_constraints.md`: WS room subscription без membership/RBAC — forbidden (NF-02) | **ROOT=code** | Дырка в безопасности скрыта отсутствием negative-теста — добавить проверку в `bridge.go` + тест |

### Домен `tms` (`backend/internal/tms/`)

| # | Симптом | Функционал по docs | Вердикт | Действие |
|---|---------|---------------------|---------|----------|
| TM1 | `TestTestRailImport_DuplicateCase` FAIL после wave96 | Спецификация: импорт должен делать upsert по external_id | **ROOT=code** | `testrail_import.go` создаёт дубликат вместо upsert |
| TM2 | `TestRunsExtra_StatusTransition` FAIL: mock возвращает старую сигнатуру сервиса | Поведение не менялось, изменился только интерфейс `service.go` | **ROOT=test** | Обновить mock/stub под новую сигнатуру |

### Домен `calls` (`backend/internal/calls/`, LiveKit)

| # | Симптом | Функционал по docs | Вердикт | Действие |
|---|---------|---------------------|---------|----------|
| C1 | `TestChatMongo_MessagePersist` FAIL: сообщение не сохраняется после reconnect | `docs/architecture.md`: LiveKit-домен — SEED-001 mediasoup REJECTED, канон LiveKit | **ROOT=code** | Consumer не переподписывается на JetStream после reconnect — фикс в `chat_mongo.go` |
| C2 | `TestSearch_Wave103Mongo_NoUserScope` PASS, должен FAIL | `tz_constraints.md`: Search без user-scope фильтра — forbidden (NF-03) | **ROOT=code** | Добавить user-scope guard + negative-тест до фикса самого поиска |

---

## 19. Аудит-цикл — примеры закрытия findings (§ ссылка на §правила аудит-цикла выше)

Примеры реально закрытых findings из `.planning/audits/2026-06/`:

| Finding | Файл-источник | Домен | Закрыто через |
|---------|----------------|-------|----------------|
| SEC-01 APIKeyAuth dead code | `AUDIT-2026-06-23-full.md` | auth | `be1` fix + `be-ver` + regression `testframework/auth/` |
| NF-02 WS room без RBAC | `FA-RBAC-FIBER3-AUDIT.md` | ws/messenger | `security-auditor` finding → `be2` fix `bridge.go` + `testframework/ws/authorizer_test.go` |
| FA-02 Security re-audit gaps | `FA-02-SECURITY-REAUDIT.md` | platform | `security-auditor` → `gsd-audit-fix` → cross-domain fix wave |
| TECH-DEBT tracker automation coupling | `TECH-DEBT-AUDIT-2026-06-23.md` | tracker | `arch` рефакторинг-план → `be1`/`be2` волна |

**Правило:** при работе с аудитом всегда сверять статус finding в `.planning/audits/2026-06/AUDIT-SCORECARD.md` **до** и **после** фикса — не помечать closed без re-verify по коду.

---

## 20. Prompt Engineering Toolkit (`prompt-engineering/`)

Переносимый (project-agnostic) тулкит промпт-инжиниринга для работы с тестовым фреймворком. Правило: [`.cursor/rules/70-prompt-engineering.mdc`](rules/70-prompt-engineering.mdc). Полный индекс: [`prompt-engineering/README.md`](../prompt-engineering/README.md).

### Назначение

Готовый набор агентов, скилов, правил, контекст-шаблонов, плейбуков и **1000 промптов** для тестирования, который можно скопировать в любой проект и оживить одним скилом `context-intake` (сканирует репо → заполняет `context/*.md` фактами). Обезличен: Hub-специфика живёт только в `context/*.md` и rule 70, не в самих агентах/промптах.

### Состав

| Блок | Путь | Содержание |
|------|------|------------|
| Агенты | `prompt-engineering/agents/` | `test-code-reviewer` (паттерны фреймворка + нейминг функций/файлов), `test-case-generator` (TC из документации + верификация), `test-debugger` (triage code/test/flaky), `test-documenter` (синхрон доков) |
| Скилы | `prompt-engineering/skills/` | `context-intake` (бутстрап контекста), `anti-ai-check` (говнокод + AI-комментарии + фейковые тесты), `naming-check` (аудит нейминга), `test-verification` (гейт TC) |
| Правила | `prompt-engineering/rules/` | `00` глобальные, `10` нейминг, `20` качество теста, `30` severity, **`40` 10-этапная верификация** |
| Плейбуки | `prompt-engineering/playbooks/` | onboarding, test-generation, review-cycle, debugging, documentation |
| Промпты | `prompt-engineering/prompts/` | 12 категорий × ~85 = **1000** промптов со `{{слотами}}` |
| Шаблоны запуска | `prompt-engineering/agents/templates/` | `_MASTER-PROMPT-TEMPLATE.md` (кто/что/где + best practices) + 9 сценарных |

### Ключевые механизмы

1. **Единый шаблон промпта** — `_MASTER-PROMPT-TEMPLATE.md` (role prompting, контекст-инъекция, формат вывода, few-shot, chain-of-thought, step-back/ReAct, параметры генерации, верификация). Обезличенное зеркало канона `team-lead.md` §5.2 / `team-workflow.md` §5.3.1.
2. **10-этапная верификация** (`rules/40`) — каждый агент/скил перед сдачей прогоняет self-check по 10 этапам (контекст→требование→факты→негатив→точность→критич.класс→регрессия→формат→воспроизводимость→врата); для CRITICAL — independent-проход. Согласуется с §3, §4, rule 56.

### Правила для агентов Hub

- Тестовые задачи (ревью тестов, генерация TC, дебаг тестов, документация фреймворка) — через агентов тулкита как инструмент **внутри** Hub QA-конвейера (§5), не вместо него.
- Не хардкодить Hub-факты в агентов/промпты тулкита — только в `context/*.md`. Иначе теряется переносимость.
- `anti-ai-check` тулкита — для общих AI-паттернов и фейковых тестов; Hub-специфику (minio/gorilla/SSRF) держит Hub `.cursor/skills/anti-ai-check`.

---

*Hub AI Team Regulation — 2026-07-19 (§20 Prompt Engineering Toolkit) / 2026-07-18 (§12–§19 примеры по доменам)*
