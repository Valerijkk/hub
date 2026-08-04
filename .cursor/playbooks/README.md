# Hub — AI Team Playbooks (индекс)

> Единый справочник команды. Агент начинает с STATE/индекса и загружает только
> профильные docs/playbooks, необходимые для текущего outcome.
> Регламент: `.cursor/REGULATION.md` · GSD: `.cursor/GSD-INTEGRATION.md` · Ростер файлов агентов: `.cursor/agents/*.md` · Planning-навигация: [`.planning/README.md`](../../.planning/README.md)
>
> **Хребет цикла — [`phase-loop.md`](phase-loop.md):** пять шагов на фазу и
> ворота между ними. Остальные плейбуки раскрывают отдельные шаги.

*Обновлено: 2026-07-26 — evidence-first prompt contract, dynamic model routing (ADR-MODELS-GPT56 amendment), agent frontmatter synced.*

---

## Как читать этот индекс (для новой сессии/агента)

Порядок progressive disclosure:

1. **`.planning/STATE.md`** — что происходит в проекте прямо сейчас (session notes, freeze-статус July Rebuild). 30 секунд на весь файл.
2. **`.planning/README.md`** — если нужен конкретный planning-артефакт (аудит, ADR, roadmap, prod-runbook) — навигация по папкам, не гадать путь руками.
3. **`docs/README.md`** → нужные файлы «Единой базы знаний» (`vision_and_scope.md`, `system_specification.md`, `architecture.md`, `tz_constraints.md`, ...) — факты о продукте.
4. **Этот файл** → §2/§2.1 «Карта playbook-ов» / обратная карта агент→playbook — найди свой playbook по роли; §0 — актуальная runtime-routing policy.
5. **Свой файл агента** (`.cursor/agents/<role>.md`) — конкретные обязанности, DoD, с кем взаимодействуешь; ссылается на playbooks, не дублирует их.
6. **Rules** (`.cursor/rules/*.mdc`). Always-applied: `00`, `50`, `60`.
   Glob/intelligent: `10`, `20`, `30`, `40`, `45`, `55`, `56`, `70`.

Не знаешь, какой документ обновлять после своего изменения — смотри §6 «Что обновить когда» в конце этого файла.

---

## 0. Runtime routing моделей

> **Policy (ADR-MODELS-GPT56, amended 2026-07-26):** `team-lead` оркестрирует на **GPT-5.6 Sol Pro** (`model: gpt-5.6-sol`; Pro = picker/catalog mode). Downstream-работу он маршрутизирует на любую реально доступную Cursor-модель по capability/cost/risk/context. Portable/role prompts не владеют model params — downstream agents используют `model: inherit`.

| Слот | Frontmatter | Runtime policy |
|------|-------------|----------------|
| Оркестратор `team-lead` | `gpt-5.6-sol` | **GPT-5.6 Sol Pro** (catalog resolves Pro mode) |
| Ролевые исполнители и `*-ver` | `inherit` | Dynamic routing; parent override только после catalog-backed choice + logging |
| Micro `codex-micro` | `inherit` | Lowest sufficient coding model for completion bar |
| Portable agents | `inherit` | Launcher/model-router owns slug and sampling |
| `team-lead-opus` | `inherit` | Deprecated compatibility redirect; not an escalation slot |

Parent/orchestrator may explicitly override downstream model at launch only after catalog-backed routing and logging in handoff.

---

## 1. Как устроена документация команды

Три уровня, без дублирования:

| Уровень | Где | Что содержит |
|---------|-----|--------------|
| **База знаний** | `docs/*.md` | Требования, спецификация, архитектура, тест-план, стиль, security. Единый источник правды. |
| **Playbooks** | `.cursor/playbooks/*.md` | Глубокие практики по ремеслу: как писать код/тесты/логи/аналитику. Общие для всех агентов роли. |
| **Файлы агентов** | `.cursor/agents/*.md` | Кто агент, что делает, когда, с кем, DoD. Ссылается на playbooks — не дублирует. |

**Правило:** агент читает свой файл → грузит указанные там playbooks и docs → работает. Playbook меняем один раз — эффект на всю команду.

---

## 2. Карта playbook-ов

| Playbook | Для кого | Тема |
|----------|----------|------|
| `backend.md` | be1, be2, be-ver | Go/Fiber/Ent, layering, ошибки, context, security, NATS, Mongo, конкурентность |
| `frontend.md` | fe1, fe2, fe-ver | React 19/TS, TanStack Router/Query, Zustand, Shadcn, a11y, производительность |
| `design.md` | designer, designer-ver, fe* | Дизайн-система `--hub-*`, Shadcn, UX, WCAG, формат дизайн-спеки |
| `testing.md` | qa-manual, qa-auto1/2, qa-*-ver, be*, fe* | Пирамида, тест-дизайн, Go/Vitest/Playwright, Allure, замер времени, MD↔автотесты |
| `observability.md` | be*, fe*, qa*, devops | Логирование (slog), метрики, тайминги, трейсинг, что не логировать |
| `code-comments.md` | все, кто пишет код/тесты | Комментарии объясняют ПОЧЕМУ; godoc/TSDoc; запреты |
| `git-workflow.md` | be*, fe*, qa-auto*, code-reviewer | Ветки, Conventional Commits, параллельная работа, PR, запреты |
| `analysis.md` | ba1/ba2, sa1/sa2, arch, ba-ver/sa-ver/arch-ver | US/AC/SMART, FR/NFR/ERD/API, C4/ADR, трассируемость |
| `verification.md` | все `*-ver`, code-reviewer, security-auditor | Протокол верификации, чек-листы по ролям, severity |
| `team-workflow.md` | team-lead, tech-lead, product, все | Состав, RACI, потоки фич, волны, параллелизм, handoff, GSD |
| `prompt-contract.md` | **все агенты (32 файла)** | Natural-language intake → внутренний task contract; concise rationale/assumptions/evidence; prompt-injection boundary; risk-scaled verification; строгий §5.2 только для delegation/handoff |

### 2.1 Обратная карта: агент → playbook-и

> §2 отвечает «playbook → кто использует»; здесь обратный индекс «агент → что грузить перед задачей», без пролистывания §2/§3.

> **`prompt-contract.md` — обязателен для всех ролей** (нормализация задачи, evidence-first output, untrusted-data boundary и handoff). В таблице ниже не повторяется построчно: считай его нулевой строкой для каждого агента.

| Агент(ы) | Playbooks к загрузке |
|----------|-----------------------|
| `team-lead`, `team-lead-opus`, `tech-lead` | `team-workflow.md` (обязателен) + playbook домена, который делегируется |
| `product`, `product-ver` | `team-workflow.md`, `analysis.md` |
| `ba1`, `ba2`, `ba-ver` | `analysis.md` |
| `sa1`, `sa2`, `sa-ver` | `analysis.md` |
| `arch`, `arch-ver` | `analysis.md`, `backend.md` (реализуемость стеком), `observability.md` |
| `designer`, `designer-ver` | `design.md` |
| `be1`, `be2`, `be-ver` | `backend.md`, `testing.md`, `observability.md`, `code-comments.md`, `git-workflow.md` |
| `fe1`, `fe2`, `fe-ver` | `frontend.md`, `design.md`, `testing.md`, `observability.md`, `code-comments.md`, `git-workflow.md` |
| `qa-manual`, `qa-manual-ver` | `testing.md` |
| `qa-auto1`, `qa-auto2`, `qa-auto-ver` | `testing.md`, `observability.md`, `git-workflow.md` |
| `devops-engineer` | `observability.md`, `git-workflow.md` |
| `security-auditor` | `backend.md` (security-секции), `verification.md` |
| `debugger` | `observability.md` + `backend.md`/`frontend.md` по домену бага |
| `code-reviewer` | `verification.md`, `git-workflow.md`, `code-comments.md` |
| `docs-writer` | `code-comments.md` (стиль примеров) + `.cursor/rules/55-docs-sync.mdc` |
| `codex-micro` | не грузит playbooks — только rule-файлы + свой agent-файл (задачи слишком малы для полного контекста) |

**Правило:** если роль выходит за штатный сценарий (например, `fe-ver` делает разовое backend-ревью), грузить playbook **целевого домена задачи**, а не playbook своей роли по умолчанию.

### 2.2 Карта «тонкий rule → толстый playbook»

`.cursor/rules/*.mdc` — краткие карточки (что нельзя/обязательно + 1 пример); часть always-applied, часть glob-scoped (см. §«Как читать индекс»). Playbook — глубина (почему, как тестировать, антипаттерны). Если правило не помещается в 5–10 строк с примером — это уже playbook, не rule.

| Тонкое правило (`.cursor/rules/`) | Толстый playbook | Связь |
|---|---|---|
| `10-style.mdc`, `20-backend.mdc` | `backend.md` | Rule = запрещённые паттерны + пример; playbook = layering, конкурентность, NATS, Mongo целиком |
| `10-style.mdc`, `30-frontend.mdc` | `frontend.md` | Rule = Rules of Hooks/design tokens кратко; playbook = TanStack Query, a11y, перф глубоко |
| — (нет отдельного rule) | `design.md` | Дизайн-система `--hub-*` — слишком объёмно для rule, живёт только в playbook |
| `40-testing.mdc` | `testing.md` | Rule = pipeline map + команды; playbook = пирамида, техники тест-дизайна, Allure, MD↔автотест |
| — (нет отдельного rule) | `observability.md` | slog-паттерны, метрики, трейсинг — детально только в playbook |
| `10-style.mdc` (кратко) | `code-comments.md` | Rule = «объясняй ПОЧЕМУ» одной строкой; playbook = godoc/TSDoc, примеры |
| `50-git.mdc` | `git-workflow.md` | Rule = naming/commit format/что нельзя коммитить; playbook = параллельная работа, PR-флоу |
| — (нет отдельного rule) | `analysis.md` | US/AC/SMART/FR/NFR/ERD/C4/ADR методология — только в playbook |
| `56-regression-verify.mdc` | `verification.md`, `testing.md` §11.1 | Rule = обязательный цикл кратко; playbook = чек-листы по ролям, severity |
| `00-project.mdc` | `team-workflow.md` | Rule = стек/статус кратко; playbook = RACI, волны, handoff, шаблон делегирования §5.3.1 |
| `70-pro-qa-ai.mdc` | `prompt-contract.md` | Rule = карта тулкита `pro-qa-ai/`; playbook = task normalization, concise rationale/assumptions/evidence, risk-scaled verification и handoff |

Практический вывод: не хватает контекста из rule — ожидаемо, иди в playbook по этой таблице.

---

## 3. Ростер команды (32 файла агентов)

**Runtime:** `team-lead` — GPT-5.6 Sol Pro (`gpt-5.6-sol` + Pro via catalog); downstream — `inherit`, routing по capability/cost/risk среди реально доступных Cursor-моделей (§0).

### Тир 1 — Лидерство
| Агент | Файл | Runtime | Роль |
|-------|------|-----------------|------|
| Team Lead | `team-lead.md` | **Sol Pro** | Оркестратор волн, делегирование, финальная приёмка |
| Team Lead (esc.) | `team-lead-opus.md` | deprecated | Исторический профиль; routing теперь внутри `team-lead` |
| Tech Lead | `tech-lead.md` | routed | Декомпозиция на фазы/планы, зависимости, риски |
| Codex Micro | `codex-micro.md` | `inherit` (routed micro) | Супер-мелкие правки (не волны) |

### Тир 2 — Продукт
| Агент | Файл | Роль |
|-------|------|------|
| Product | `product.md` | Приоритеты, scope, KPI, roadmap-решения |
| Product Verifier | `product-ver.md` | Соответствие vision/KPI/scope |

### Тир 3 — Аналитика
| Агент | Файл | Роль |
|-------|------|------|
| Business Analyst 1 | `ba1.md` | Требования/US/AC — домены A (auth, tracker, sprint, messenger) |
| Business Analyst 2 | `ba2.md` | Требования/US/AC — домены B (tms, calls, dashboard, v4-модули) |
| BA Verifier | `ba-ver.md` | SMART, полнота US/AC, MoSCoW |
| System Analyst 1 | `sa1.md` | Спецификация/ERD/API — домены A |
| System Analyst 2 | `sa2.md` | Спецификация/ERD/API — домены B |
| SA Verifier | `sa-ver.md` | Покрытие требований, ERD, API-контракты |

### Тир 4 — Архитектура
| Агент | Файл | Роль |
|-------|------|------|
| Architect | `arch.md` | C4, deployment, выбор библиотек, ADR, TZ-инварианты |
| Arch Verifier | `arch-ver.md` | SOLID, KISS, реализуемость стеком, соответствие TZ |

### Тир 5 — Дизайн
| Агент | Файл | Роль |
|-------|------|------|
| Designer | `designer.md` | Дизайн-система, UX-спеки, токены, макеты |
| Designer Verifier | `designer-ver.md` | Консистентность токенов, a11y, отзывчивость |

### Тир 6 — Backend
| Агент | Файл | Роль |
|-------|------|------|
| Backend Dev 1 | `be1.md` | Реализация бэкенда — поток A (параллельно с be2) |
| Backend Dev 2 | `be2.md` | Реализация бэкенда — поток B; кросс-ревью be1 |
| BE Verifier | `be-ver.md` | Layering, security, ошибки, тесты, race |

### Тир 7 — Frontend
| Агент | Файл | Роль |
|-------|------|------|
| Frontend Dev 1 | `fe1.md` | Реализация фронта — поток A (параллельно с fe2) |
| Frontend Dev 2 | `fe2.md` | Реализация фронта — поток B; кросс-ревью fe1 |
| FE Verifier | `fe-ver.md` | Rules of Hooks, a11y, дизайн-система, Lighthouse |

### Тир 8 — QA
| Агент | Файл | Роль |
|-------|------|------|
| Manual QA | `qa-manual.md` | Тест-кейсы (TC-*), тест-дизайн, ручные чек-листы, MD-документация |
| Manual QA Verifier | `qa-manual-ver.md` | Полнота TC, техники, трассируемость |
| Auto QA 1 | `qa-auto1.md` | Автотесты — поток A (backend/unit) |
| Auto QA 2 | `qa-auto2.md` | Автотесты — поток B (frontend/e2e); кросс-ревью qa-auto1 |
| Auto QA Verifier | `qa-auto-ver.md` | Детерминизм, flaky, TC↔автотест, покрытие |

### Тир 9 — Cross-cutting специалисты
| Агент | Файл | Роль |
|-------|------|------|
| DevOps | `devops-engineer.md` | Docker, K3s, CI/CD, Caddy, compose |
| Security Auditor | `security-auditor.md` | OWASP/ASVS, security-инварианты TZ |
| Debugger | `debugger.md` | Root-cause по трейсам/логам |
| Code Reviewer | `code-reviewer.md` | Pre-merge diff review |
| Docs Writer | `docs-writer.md` | Синхронизация docs/ и wiki, Swagger |

---

## 4. Разделение параллельной работы (без конфликтов)

Дубли ролей (be1/be2, fe1/fe2, ba1/ba2, sa1/sa2, qa-auto1/qa-auto2) работают одновременно. Правила, чтобы не мешать друг другу:

1. **Границы по модулям.** Один агент = один набор модулей/файлов на волну. Не редактируют одни и те же файлы одновременно.
2. **Общие интерфейсы — заранее.** Общие типы/контракты/DTO фиксирует SA/ARCH до старта; разработчики их не меняют в одностороннем порядке.
3. **Кросс-ревью.** be1 ревьюит PR be2 и наоборот (аналогично fe, qa-auto) — до передачи персональному верификатору.
4. **Частый rebase на `master`, маленькие атомарные PR** (≤200–300 строк).
5. **Порядок мержа** назначает tech-lead/team-lead; конфликтующие изменения сериализуются.

Детали — `team-workflow.md` и `git-workflow.md`.

---

## 5. Обязательные ворота (для всех)

| Ворота | Правило | Ссылка |
|--------|---------|--------|
| Приём задачи | Принять natural-language запрос, внутренне нормализовать outcome/constraints/evidence/completion bar; спросить только при material ambiguity/destructive choice | `prompt-contract.md` §2 |
| Evidence | Перед handoff — risk-scaled evidence; `verified: true` только независимым verifier | `prompt-contract.md` §6 |
| Регрессия | После изменения кода/поведения — сразу прогон; без green нет handoff | REGULATION §7 |
| Docs-sync | После user-visible/prod изменения — обновить docs в той же сессии | `.cursor/rules/55-docs-sync.mdc` |
| Test-triage | Упал тест → сначала проверить функционал по docs, потом чинить | `.cursor/skills/test-triage/SKILL.md` |
| Верификация | Каждый артефакт → персональный `*-ver`; `verified:true` только при green + лог | `verification.md` |
| /qa-gate | Финальный чек перед handoff/merge | `.cursor/skills/qa-gate/SKILL.md` |
| Cleanup | Удалить coverage-dumps/temp до коммита | `.cursor/rules/45-artifacts-cleanup.mdc` |

---

## 6. Что обновить когда (шпаргалка агента)

После своего изменения — сверься с этой таблицей, чтобы не забыть doc-sync (полная матрица — `.cursor/rules/55-docs-sync.mdc`):

| Что сделал | Что обновить |
|---|---|
| Изменил код/фичу/фикс | `.planning/STATE.md` (session note) |
| Добавил/сломал API-контракт | `docs/system_specification.md`, `docs/wiki/API.md` |
| Новый UI-flow/hotkey | `docs/wiki/SHORTCUTS.md`, `docs/wiki/QUICKSTART.md`, `docs/vision_and_scope.md` (если US) |
| Написал тесты / поменял coverage | `.planning/COVERAGE.md`, `docs/test_plan.md`, `docs/TESTING.md` |
| Тронул CI/deploy/Docker | `docs/stack_and_tools.md`, `deployments/.env.example`, `docs/wiki/DEPLOY.md` |
| Security-фикс/аудит | `docs/security_guidelines.md`, `.planning/audits/2026-06/` (новый файл по дате) |
| Закрыл gap/wave-задачу | `.planning/archive/tasks/TASKS-CLOSE-GAPS.md` |
| Prod-ops / секрет-ротация | `.planning/ops/production/`, `docs/runbooks/` |
| Изменил роль/модель агента | этот файл (§0, §3), сам файл агента `.cursor/agents/<role>.md` |
| Добавил/поменял playbook | этот файл §2 + §2.1 (обратная карта агент→playbook) + §2.2, если поменялась связь с rule |
| Переложил папки в `.planning/` | `.planning/README.md` + все ссылки на старые пути в `docs/`, `AGENTS.md`, playbooks |
| Любое из вышеперечисленного, критично для прод/security/деплой | плюс личное ревью пользователя (`.cursor/rules/50-git.mdc`) |
