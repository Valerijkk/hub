# Hub ↔ GSD Skills Integration

> GSD skills: `~/.claude/skills/gsd-*/SKILL.md` · Hub slash-skills: `.cursor/skills/*/SKILL.md`.
> Ростер команды: `.cursor/playbooks/README.md` · Регламент: `.cursor/REGULATION.md`.
>
> **Этот файл отвечает «какая ситуация → какой скил» (маршрутизация).** Вопрос «где в промпте объявить скил и что агент с ним делает» — [`playbooks/prompt-contract.md`](playbooks/prompt-contract.md) §13 (в т.ч. обязательные скилы-ворота: `/test-triage`, `/qa-gate`, `/handoff-sync`, `/style-check`, `/coverage-audit`).

*Обновлено: 2026-07-19 — ссылка на контракт промпта (§13 prompt-contract) / 2026-07-18 — ростер 30 агентов + playbooks.*

---

## Orchestrator (team-lead / пользователь)

| Intent | GSD Skill | Hub Slash |
|--------|-----------|-----------|
| Не знаю какую команду | `gsd-do` | — |
| Статус проекта | `gsd-progress` | — |
| Следующий шаг | `gsd-next` | — |
| Волны/исполнение | `gsd-execute-phase`, `gsd-workstreams` | — |
| Мелкая задача | `gsd-quick`, `gsd-fast` | — |
| Синхронизация STATE | `gsd-handoff` via | `/handoff-sync` |
| Полный quality gate | — | `/qa-gate` |
| Матрица покрытия | `gsd-coverage-audit` | `/coverage-audit` |

---

## Ростер 30 агентов → GSD + playbook

| Agent | Primary GSD | Secondary GSD | Playbook |
|-------|-------------|---------------|----------|
| `team-lead` | `gsd-do`, `gsd-execute-phase`, `gsd-autonomous` | `gsd-workstreams`, `gsd-progress` | `team-workflow.md` |
| `tech-lead` | `gsd-plan-phase`, `gsd-analyze-dependencies` | `gsd-insert-phase`, `gsd-execute-phase` | `team-workflow.md` |
| `product` | `gsd-new-milestone`, `gsd-review-backlog` | `gsd-explore`, `gsd-plant-seed` | `analysis.md`, `team-workflow.md` |
| `product-ver` | `gsd-verify-work` | `gsd-audit-milestone` | `verification.md` |
| `ba1`, `ba2` | `gsd-discuss-phase`, `gsd-spec-phase` | `gsd-review-backlog` | `analysis.md` |
| `ba-ver` | `gsd-verify-work` | — | `verification.md` |
| `sa1`, `sa2` | `gsd-coverage-audit`, `gsd-intel` | `gsd-audit-uat`, `gsd-map-codebase` | `analysis.md` |
| `sa-ver` | `gsd-verify-work` | `gsd-audit-uat` | `verification.md` |
| `arch` | `gsd-ai-integration-phase`, `gsd-ui-phase` | `gsd-research-phase`, `gsd-select-framework` | `analysis.md`, `backend.md`, `frontend.md` |
| `arch-ver` | `gsd-verify-work` | `gsd-secure-phase` | `verification.md` |
| `designer` | `gsd-sketch`, `gsd-ui-phase` | `gsd-ui-review`, `gsd-sketch-wrap-up` | `design.md` |
| `designer-ver` | `gsd-ui-review` | `gsd-verify-work` | `verification.md`, `design.md` |
| `be1`, `be2` | `gsd-quick`, `gsd-debug` | `gsd-fast`, `gsd-add-tests` | `backend.md`, `observability.md` |
| `be-ver` | `gsd-verify-work`, `gsd-code-review` | `gsd-secure-phase` | `verification.md`, `backend.md` |
| `fe1`, `fe2` | `gsd-quick`, `gsd-ui-review` | `gsd-sketch`, `gsd-add-tests` | `frontend.md`, `design.md` |
| `fe-ver` | `gsd-verify-work`, `gsd-ui-review` | — | `verification.md`, `frontend.md` |
| `qa-manual` | `gsd-audit-uat`, `gsd-verify-work` | — | `testing.md` |
| `qa-manual-ver` | `gsd-verify-work`, `gsd-audit-uat` | — | `verification.md` |
| `qa-auto1`, `qa-auto2` | `gsd-add-tests`, `gsd-validate-phase` | `gsd-audit-fix` | `testing.md` |
| `qa-auto-ver` | `gsd-verify-work`, `gsd-validate-phase` | — | `verification.md`, `testing.md` |
| `devops-engineer` | `gsd-ship` | — | `observability.md`, `git-workflow.md` |
| `security-auditor` | `gsd-secure-phase`, `gsd-audit-fix` | — | `verification.md`, `backend.md` |
| `debugger` | `gsd-debug` | — | `observability.md`, `testing.md` |
| `code-reviewer` | `gsd-code-review`, `gsd-code-review-fix` | — | `verification.md`, `code-comments.md`, `git-workflow.md` |
| `docs-writer` | `gsd-docs-update`, `gsd-milestone-summary` | — | `code-comments.md` |

---

## Планы Hub в GSD-терминах

| Hub Artifact | GSD Equivalent |
|--------------|----------------|
| `.planning/ROADMAP.md` | Milestone phases |
| `.planning/STATE.md` | `gsd-progress` state |
| `.planning/COVERAGE.md` | `gsd-coverage-audit` output |
| `.planning/EXPANSION-ANALYSIS-*.md` | backlog для `gsd-plan-phase` |
| `.planning/quick/` | `gsd-quick` tasks |
| `.planning/debug/` | `gsd-debug` sessions |

---

## Правило: Test Triage (обязательно)

Перед фиксом упавшего теста — `.cursor/skills/test-triage/SKILL.md`:
1. Воспроизвести функционал (не тест). 2. Сверить с `docs/`. 3. Код верен → чинить тест; код неверен → чинить код. 4. Запрет «зелёных» без шагов 1–2.

---

## Wave execution (gsd-execute-phase)

```
Волна N: team-lead запускает ≤4 параллельных агента (разные модули/файлы)
  → каждый: реализация → регрессия → test-triage на FAIL → персональный *-ver
  → кросс-ревью парного агента (be1↔be2 / fe1↔fe2 / qa-auto1↔qa-auto2)
  → team-lead: сборка + /handoff-sync
Волна N+1: зависит от артефактов Волны N
```

Детали и RACI: `.cursor/playbooks/team-workflow.md`

---

## Когда какую GSD-команду — расширенная матрица по фазам жизненного цикла

> Дублирует часть `.cursor/skills/gsd-index/SKILL.md`, но с привязкой к реальным ситуациям Hub. Полный список GSD-команд: `~/.claude/skills/gsd-*/SKILL.md` (более 60 skills), здесь — только те, что реально используются в Hub-практике.

### 1. Старт / ориентация в сессии

| Ситуация | Команда | Комментарий |
|----------|---------|-------------|
| Новая сессия, не знаю с чего начать | `gsd-do` | Роутит freeform-текст в нужную команду |
| «Где мы сейчас в проекте?» | `gsd-progress` | Сверяет `.planning/STATE.md` + `ROADMAP.md`, есть флаг `--forensic` для integrity-аудита |
| «Что делать дальше?» | `gsd-next` | Логический шаг после завершения фазы/плана |
| Хочу список todo | `gsd-check-todos` | Перед стартом новой волны |
| Резюме предыдущей сессии | `gsd-resume-work` | После `gsd-pause-work` в прошлой сессии |

### 2. Планирование фичи/фазы

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Крупная новая фича, нужна декомпозиция | `gsd-discuss-phase` → `gsd-plan-phase` | CAL-SYNC (email+calendar+calls sync) |
| Нужно понять зависимости фаз | `gsd-analyze-dependencies` | Перед вставкой новой фазы в roadmap |
| Срочная задача между фазами | `gsd-insert-phase` (напр. `72.1`) | Хотфикс безопасности вне очереди |
| Фича на конец roadmap | `gsd-add-phase` | Добавление v4 AI hub фазы |
| Нужен research перед планом | `gsd-research-phase`, `gsd-phase-researcher` | Выбор AI framework для gateway BYOK |
| AI-система (провайдер/агент/RAG) | `gsd-ai-integration-phase` | AI Provider Hub Spec (`.planning/specs/AI-PROVIDER-HUB-SPEC.md`) |
| UI-контракт перед реализацией | `gsd-ui-phase` → `gsd-ui-review` | Bases inline-edit экран |
| Проверка плана до исполнения | `gsd-plan-checker` | team-lead перед запуском волны |

### 3. Исполнение

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Параллельное исполнение планов волны | `gsd-execute-phase` | Волна be1/be2/fe1/fe2 на 4 независимых модуля |
| Несколько долгоживущих параллельных потоков | `gsd-workstreams` | Track F (frontend) и Track P (platform) одновременно |
| Мелкая точечная задача с трекингом | `gsd-quick` | Добавить один RBAC-guard в `bases/service.go` |
| Тривиальный inline-фикс без оверхеда | `gsd-fast` | Опечатка в error message, переименование поля DTO |
| Всё подряд автономно | `gsd-autonomous` | Только team-lead/team-lead-opus, финальные волны |

### 4. Качество / тесты / баги

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Нужны тесты после фичи | `gsd-add-tests` | После WIP-лимитов в tracker — `TestBoards_WipLimit*` |
| Nyquist-gaps после фазы | `gsd-validate-phase` | Retroactive проверка фазы 121 |
| UAT / acceptance проверка | `gsd-verify-work` | Приёмка Wave C аудита |
| Сложный баг, нужна persistent-сессия | `gsd-debug` | Flaky `TestSessionMiddleware_ExpiredJWT`, race в `ws/hub.go` |
| Массовые findings → фикс | `gsd-audit-fix` | После `AUDIT-2026-06-23-full.md` (сначала `--dry-run`) |
| Cross-phase UAT-долги | `gsd-audit-uat` | Ревизия открытых UAT перед релизом |
| Root-cause post-mortem неудачного workflow | `gsd-forensics` | Волна упала на интеграции — разбор что пошло не так |
| **Hub:** упавший тест — код vs тест | `.cursor/skills/test-triage/SKILL.md` | См. `REGULATION.md` §14, §18 |
| **Hub:** полный gate | `.cursor/skills/qa-gate/SKILL.md` | См. `REGULATION.md` §15 |

### 5. Документация / аудит покрытия

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Bulk-обновление docs с верификацией по коду | `gsd-docs-update` | После `.planning/` reorg — проверка ссылок в `docs/README.md` |
| Матрица требование×код×тест | `gsd-coverage-audit` | Обновление `.planning/COVERAGE.md` |
| Codebase intelligence refresh | `gsd-intel` | После крупного рефакторинга tracker automation |
| Глубокий анализ кодовой базы | `gsd-map-codebase` | Перед v4 super-app декомпозицией |
| Milestone-резюме для онбординга | `gsd-milestone-summary` | `.planning/milestones/v4-super-app/SUMMARY.md` |
| Peer review плана/кода | `gsd-code-review` → `gsd-code-review-fix` | Pre-merge diff review |
| Threat-mitigation verification | `gsd-secure-phase` | После изменений auth/RBAC |
| Аудит eval-покрытия AI-фазы | `gsd-eval-review` | AI gateway BYOK evaluation strategy |

### 6. Завершение / handoff

| Ситуация | Команда | Пример Hub |
|----------|---------|------------|
| Конец значимой сессии | `/handoff-sync` (Hub) + `gsd-progress` | Всегда перед закрытием сессии с кодом |
| Пауза посреди фазы | `gsd-pause-work` | Context handoff на следующую сессию |
| Завершение milestone | `gsd-complete-milestone`, `gsd-audit-milestone` | v4 super-app → архивация |
| PR готов к ревью | `gsd-ship` | После green `/qa-gate` |
| Разбить на несколько PR | `gsd-pr-branch` | Волна с 4 модулями → 4 отдельных PR |

---

## Ростер 30 агентов → GSD + playbook (детализация по use-case)

> Таблица выше (§«Ростер 30 агентов») даёт primary/secondary skill. Ниже — конкретные ситуации, когда агент реально вызывает каждый skill.

| Agent | Skill | Реальная ситуация Hub |
|-------|-------|------------------------|
| `team-lead` | `gsd-execute-phase` | Запуск волны 4 параллельных be/fe агентов на разные модули (auth+tracker+bases+tms) |
| `team-lead` | `gsd-workstreams` | Одновременное ведение Track F (frontend redesign) и Track P (platform/otel) |
| `tech-lead` | `gsd-plan-phase` | Декомпозиция CAL-SYNC на фазы email L0-L3 + calendar + calls RSVP |
| `ba1`/`ba2` | `gsd-discuss-phase` | Уточнение серых зон перед US для intranet v2 (poll-in-post, kudos) |
| `sa1`/`sa2` | `gsd-coverage-audit` | Трассировка FR-CAL-01…08 → `calendar_event` ERD → тесты |
| `arch` | `gsd-ai-integration-phase` | Framework selection для AI gateway (BYOK vs Ollama) |
| `designer` | `gsd-sketch` | Throwaway HTML-мокапы bases inline-edit перед `gsd-ui-phase` |
| `be1`/`be2` | `gsd-debug` | Race condition в `ws/hub_shutdown_test.go` — persistent debug session |
| `fe1`/`fe2` | `gsd-ui-review` | Retroactive 6-pillar аудит v4 bases-page.tsx после реализации |
| `qa-manual` | `gsd-audit-uat` | Свежая ревизия TC для tms после wave96 testrail import |
| `qa-auto1`/`qa-auto2` | `gsd-add-tests` | Новые тесты `service_wave12x_test.go` после каждой волны tracker |
| `devops-engineer` | `gsd-ship` | PR для CI job `coverage-gate` после изменения threshold |
| `security-auditor` | `gsd-secure-phase` | Retroactive проверка threat model после RBAC-фикса в bases |
| `debugger` | `gsd-debug` | Root-cause паники в `automation_consumer.go` под нагрузкой |
| `code-reviewer` | `gsd-code-review` | Pre-merge review волны из 4 be/fe веток перед интеграцией |
| `docs-writer` | `gsd-docs-update` | Синхронизация `docs/system_specification.md` после Волны 2 (Intranet/Calendar/Wiki спеки) |

---

## Планы Hub в GSD-терминах (актуализировано после reorg `.planning/` 2026-07-18)

| Hub Artifact | GSD Equivalent | Путь |
|--------------|-----------------|------|
| `.planning/ROADMAP.md` | Milestone phases (исторический, 125 фаз) | корень |
| `.planning/july-2026/ROADMAP-FINAL-2026-07-03.md` | Актуальный roadmap (волны Track F/P, TF-PY, DOCS-EXCELLENCE) | `july-2026/` |
| `.planning/STATE.md` | `gsd-progress` state | корень |
| `.planning/COVERAGE.md` | `gsd-coverage-audit` output | корень |
| `.planning/july-2026/EXPANSION-ANALYSIS-2026-07-02.md` | backlog для `gsd-plan-phase` | `july-2026/` |
| `.planning/quick/` | `gsd-quick` tasks | корень (если создаётся) |
| `.planning/debug/` | `gsd-debug` sessions | корень |
| `.planning/gray-zones/GRAY-ZONES-DECISIONS-2026-07-03.md` | Зафиксированные gray-area решения (аналог `gsd-list-phase-assumptions` output) | `gray-zones/` |
| `.planning/audits/2026-06/AUDIT-SCORECARD.md` | `gsd-audit-uat` / `gsd-validate-phase` output | `audits/2026-06/` |
| `.planning/decisions/ADR-*.md` | ADR (аналог decisions log GSD) | `decisions/` |
| `.planning/seeds/SEED-*.md` | `gsd-plant-seed` output | `seeds/` |
| `.planning/milestones/v4-super-app/SUMMARY.md` | `gsd-milestone-summary` output | `milestones/` |
| `.planning/production/PRODUCTION-HANDOFF.md` | `gsd-ship` / prod handoff | `production/` |

---

## Пример: полный GSD-маршрут для новой крупной фичи (CAL-SYNC)

```
1. gsd-discuss-phase --power   → ba1/ba2 собирают серые зоны email L0-L3, calendar RSVP
2. gsd-research-phase          → arch исследует iCal RFC 5545 constraints
3. gsd-plan-phase               → tech-lead: PLAN.md с task breakdown + Depends on
4. gsd-plan-checker              → team-lead верифицирует план до исполнения
5. gsd-execute-phase             → be1 (calendar_event ERD) + be2 (ical sync) + fe1 (UI) параллельно
   → каждый: /test-triage на FAIL → персональный *-ver
6. gsd-add-tests                 → qa-auto1/2 добивают coverage gaps
7. gsd-secure-phase               → security-auditor: ValidateOutboundURL для ical_subscription (SSRF)
8. gsd-code-review                → code-reviewer pre-merge
9. gsd-docs-update                → docs-writer: system_specification.md + wiki/API.md
10. /qa-gate (Hub)                → team-lead финальный чек
11. gsd-milestone-summary          → архив после релиза
```

---

## Пример: полный GSD-маршрут для мелкого бага (auth rate-limit)

```
1. gsd-do "падает тест rate limit в auth"  → роутит в gsd-debug
2. gsd-debug                                → воспроизвести, найти root cause в service_login.go
3. .cursor/skills/test-triage/SKILL.md      → классификация ROOT=code (см. REGULATION.md §14, A1)
4. gsd-fast (или gsd-quick, если >1 файл)   → минимальный фикс
5. /qa-gate --domain auth (точечный прогон) → testframework/auth/...
6. gsd-docs-update (если менялся контракт)  → security_guidelines.md
7. be-ver                                    → verified: true
```

---

## Hub slash-skills vs GSD skills — когда что

> Частая ошибка маршрутизации: пытаться использовать GSD `gsd-verify-work` там, где нужен доменный Hub-гейт, или наоборот.

| Нужно | Используй | НЕ используй |
|-------|-----------|---------------|
| Полный build+test+lint+security гейт перед PR | `.cursor/skills/qa-gate/SKILL.md` | `gsd-verify-work` (это UAT, не CI-гейт) |
| Упавший тест — код или тест виноват | `.cursor/skills/test-triage/SKILL.md` | `gsd-debug` напрямую (сначала триаж, `gsd-debug` — только если триаж показал сложный root cause) |
| STATE/COVERAGE/ROADMAP синхронизация в конце сессии | `.cursor/skills/handoff-sync/SKILL.md` | `gsd-milestone-summary` (это для завершённого milestone, не рутинный handoff) |
| Требование×код×тест матрица | `.cursor/skills/coverage-audit/SKILL.md` | `gsd-map-codebase` (это структурный анализ, не coverage) |
| Anti-AI паттерны + стиль | `.cursor/skills/style-check/SKILL.md`, `anti-ai-check/SKILL.md` | `gsd-code-review` (более общий peer review, не построчный grep) |
| Новая крупная фаза с research+planning | `gsd-discuss-phase` → `gsd-plan-phase` | Hub-skills (их нет для планирования — это зона GSD) |
| Retroactive verify реализованной фазы | `gsd-verify-work`, `gsd-validate-phase` | Hub-skills (Hub-гейты — pre-merge, GSD verify — post-hoc) |

## Частые ошибки маршрутизации (наблюдались в практике)

1. **Запуск `gsd-audit-fix` без `--dry-run`** на первом проходе — приводит к неконтролируемым правкам сразу в 10+ файлах. Правило: всегда dry-run → ревью списка → apply.
2. **`gsd-quick` для многофайловой фичи** — превышает scope skill, нужен `gsd-plan-phase`. Признак: если задача трогает >3 модуля или требует ADR — это `plan-phase`, не `quick`.
3. **Пропуск `test-triage` перед `gsd-debug`** — `gsd-debug` дороже (persistent session), для простых кейсов (см. REGULATION §14 таблицы A1-A4/T1-T4/B1-B4) достаточно триажа за 2 минуты.
4. **`/handoff-sync` вместо `gsd-docs-update`** для bulk-верификации всех doc против кода — `handoff-sync` обновляет память сессии, не верифицирует факты по коду построчно.
5. **Забытый `gsd-secure-phase`** после auth/RBAC/webhook изменений — security-инварианты TZ (webhook signature, WS RBAC, search user-scope, WebPush SSRF) требуют явной retroactive-проверки, не покрываются обычным `be-ver`.

## Каталог GSD skills, релевантных Hub (полный, по категориям)

> `~/.claude/skills/gsd-*/SKILL.md` — более 60 skills в системе; ниже — те, что реально применимы в контексте Hub (монорепо Go+React с 30-агентной командой). Остальные (напр. `gsd-new-project`, `gsd-new-workspace`) — не актуальны, т.к. проект уже существует и не мультиворкспейс.

| Категория | Skills |
|-----------|--------|
| Оркестрация сессии | `gsd-do`, `gsd-progress`, `gsd-next`, `gsd-help`, `gsd-check-todos`, `gsd-resume-work`, `gsd-pause-work` |
| Планирование | `gsd-discuss-phase`, `gsd-spec-phase`, `gsd-plan-phase`, `gsd-plan-checker`, `gsd-analyze-dependencies`, `gsd-insert-phase`, `gsd-add-phase`, `gsd-list-phase-assumptions`, `gsd-research-phase` |
| AI-фичи | `gsd-ai-integration-phase`, `gsd-domain-researcher`, `gsd-eval-planner`, `gsd-eval-review`, `gsd-framework-selector` |
| UI/дизайн | `gsd-ui-phase`, `gsd-ui-review`, `gsd-sketch`, `gsd-sketch-wrap-up` |
| Исполнение | `gsd-execute-phase`, `gsd-workstreams`, `gsd-quick`, `gsd-fast`, `gsd-autonomous` |
| Качество/тесты | `gsd-add-tests`, `gsd-validate-phase`, `gsd-verify-work`, `gsd-audit-fix`, `gsd-audit-uat`, `gsd-debug`, `gsd-forensics`, `gsd-nyquist-auditor` (внутренний, вызывается из `gsd-validate-phase`) |
| Документация/анализ | `gsd-docs-update`, `gsd-coverage-audit`, `gsd-intel`, `gsd-map-codebase`, `gsd-code-review`, `gsd-code-review-fix`, `gsd-milestone-summary`, `gsd-extract_learnings` |
| Безопасность | `gsd-secure-phase` |
| Завершение | `gsd-complete-milestone`, `gsd-audit-milestone`, `gsd-ship`, `gsd-pr-branch`, `gsd-undo` |
| Идеи/backlog | `gsd-explore`, `gsd-plant-seed`, `gsd-add-backlog`, `gsd-review-backlog`, `gsd-note`, `gsd-add-todo` |
| Служебное | `gsd-health`, `gsd-settings`, `gsd-set-profile`, `gsd-stats`, `gsd-graphify` |

---

## Типовые многошаговые сценарии Hub (skill-цепочки)

### Сценарий: security finding из аудита → закрытие

```
1. security-auditor читает .planning/audits/2026-06/FA-02-SECURITY-REAUDIT.md
2. gsd-audit-fix --dry-run           → список кандидатов на фикс с файлами
3. Ревью списка (team-lead/security-auditor)
4. gsd-audit-fix (apply)              → be1/be2 фиксят по одному finding = 1 коммит
5. test-triage на каждый FAIL после фикса
6. gsd-secure-phase                   → retroactive verify threat model
7. docs-writer: docs/security_guidelines.md + AUDIT-SCORECARD.md статус
8. /qa-gate (Hub)                      → team-lead финальная приёмка
```

### Сценарий: coverage gap найден для домена `tms`

```
1. gsd-coverage-audit (или /coverage-audit Hub) → .planning/COVERAGE.md обновлён, виден gap tms 78%<82%
2. qa-auto1/qa-auto2 назначаются на gap
3. gsd-add-tests                        → новые testframework/tms/service_waveNNN_test.go
4. Цикл: новое поведение → регрессия домена tms → /qa-gate точечный (testframework/tms/...)
5. qa-auto-ver                           → verified: true
6. docs-writer: .planning/COVERAGE.md, docs/test_plan.md
```

### Сценарий: эскалация team-lead → team-lead-opus

```
Условие эскалации: декомпозиция затрагивает >4 модуля ОДНОВРЕМЕННО,
  ИЛИ архитектурное решение без готового ADR,
  ИЛИ финальная приёмка milestone (не фазы).

1. team-lead формулирует запрос эскалации с контекстом (что не может решить Fable-дефолтом)
2. team-lead-opus (Opus 4.8 @ 1M) : gsd-plan-phase / gsd-audit-milestone на полном контексте
3. Результат передаётся обратно team-lead для исполнения волнами
```

### Сценарий: UI-фича с нуля (bases inline-edit)

```
1. gsd-sketch                    → designer: throwaway HTML варианты (2-3)
2. Выбор варианта пользователем/PM
3. gsd-ui-phase                  → designer: UI-SPEC.md (токены --hub-*, состояния, a11y)
4. gsd-ui-checker                 → designer-ver: BLOCK/FLAG/PASS до реализации
5. fe1/fe2 implement по UI-SPEC.md
6. gsd-ui-review                  → fe-ver / designer-ver: 6-pillar аудит после реализации
7. Vitest: testframework/stages/unit/frontend-test/domains/v4/bases-page.test.tsx
8. e2e: testframework/stages/e2e/domains/v4/v4-superapp.spec.ts
9. docs-writer: wiki/SHORTCUTS.md если новый хоткей
```

### Сценарий: паника в production path (ws hub)

```
1. Симптом: паника в логах ws/hub.go под нагрузкой
2. .cursor/skills/test-triage/SKILL.md — воспроизвести НЕ через "почему тест не падает", а через docs (Fiber WS — не менять на net/http, канон)
3. gsd-debug                       → debugger: систематический root-cause (race, nil pointer, closed channel)
4. Фикс минимальный (rule 10-style.mdc: no panic() in production paths)
5. Regression: testframework/ws/... (hub_test.go, hub_shutdown_test.go, bridge_lifecycle_test.go) — race detector обязателен
6. be-ver + security-auditor (если паника эксплуатируема — DoS)
7. docs-writer: STATE.md + при необходимости observability.md (алерты/метрики на этот паттерн)
```

---

## Глоссарий: термины GSD ↔ термины Hub

| GSD термин | Эквивалент в Hub | Комментарий |
|------------|------------------|-------------|
| Phase | Фаза (`.planning/phases/NN-*/PLAN.md`) | В Hub фазы нумеруются сквозно (1-125+), не сбрасываются на milestone |
| Plan | План внутри фазы, атомарные коммиты | PR = фаза, коммиты внутри = планы (rule `50-git.mdc`) |
| Milestone | v1.0 / v2.0 / v2.5 / v2.6 / v3.0 / v4 | Архивируется в `.planning/milestones/` |
| Gray area | Серая зона | `.planning/gray-zones/AREA-*.md` + `GRAY-ZONES-DECISIONS-*.md` |
| Verification loop | Цикл верификации | REGULATION.md §3, §16 (цикличное качество) |
| UAT | Приёмка пользователем | `qa-manual` + `gsd-verify-work` |
| Seed | Отложенная идея с триггером | `.planning/seeds/SEED-*.md` |
| Workstream | Параллельный поток разработки | Track F / Track P в `july-2026/ROADMAP-FINAL-2026-07-03.md` |
| Backlog parking lot | Бэклог 999.x | `.planning/roadmaps/FEATURE-BACKLOG-2026.md` |

---

*Updated: 2026-07-18 — docs-writer: расширение матриц (Когда какая GSD-команда, use-case таблица агентов, актуальные пути `.planning/`, сквозные примеры маршрута, каталог skills, частые ошибки маршрутизации, 5 многошаговых сценариев, глоссарий)*
