# Hub — AI Team Playbooks (индекс)

> Единый справочник команды. Каждый агент **перед задачей** грузит `docs/*` (база знаний) и профильные playbook-и из этого каталога.
> Регламент: `.cursor/REGULATION.md` · GSD: `.cursor/GSD-INTEGRATION.md` · Ростер файлов агентов: `.cursor/agents/*.md`

*Обновлено: 2026-07-02 — реорганизация команды под параллельную разработку.*

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

---

## 3. Ростер команды (30 агентов)

**Модели:** `team-lead` — **Fable 5 (thinking-high)** дефолтный оркестратор; `team-lead-opus` — **Opus 4.8 @ 1M** эскалация для самых тяжёлых декомпозиций/приёмок. Все остальные — **Sonnet 5, thinking-high (max effort), ~200–300k контекст**. Выбор Fable дефолтом — по итогам 5 A/B-задач (паритет качества + лучшее grounding + дешевле).

> Примечание: размер контекста и «effort» задаются профилем модели в Cursor; во frontmatter фиксируется slug. `team-lead` = `claude-fable-5-thinking-high`, `team-lead-opus` = `claude-opus-4-8-thinking-high`, остальные = `claude-sonnet-5-thinking-high`.

### Тир 1 — Лидерство
| Агент | Файл | Модель | Роль |
|-------|------|--------|------|
| Team Lead | `team-lead.md` | Fable 5 | Оркестратор волн, делегирование субагентам, финальная приёмка |
| Team Lead (esc.) | `team-lead-opus.md` | Opus 4.8 (1M) | Эскалация: тяжёлые декомпозиции/приёмка milestone |
| Tech Lead | `tech-lead.md` | Sonnet 5 | Декомпозиция на фазы/планы, зависимости, риски |

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
| Регрессия | После изменения кода/поведения — сразу прогон; без green нет handoff | REGULATION §7 |
| Docs-sync | После user-visible/prod изменения — обновить docs в той же сессии | `.cursor/rules/55-docs-sync.mdc` |
| Test-triage | Упал тест → сначала проверить функционал по docs, потом чинить | `.cursor/skills/test-triage/SKILL.md` |
| Верификация | Каждый артефакт → персональный `*-ver`; `verified:true` только при green + лог | `verification.md` |
| /qa-gate | Финальный чек перед handoff/merge | `.cursor/skills/qa-gate/SKILL.md` |
| Cleanup | Удалить coverage-dumps/temp до коммита | `.cursor/rules/45-artifacts-cleanup.mdc` |
