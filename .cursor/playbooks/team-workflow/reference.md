# Командный процесс: справочник

> Часть плейбука. Канон и оглавление — [`../team-workflow.md`](../team-workflow.md).

Правила загрузки контекста, приложения, развёрнутые примеры.

---

## 11.1 Правило загрузки

Каждый агент **перед задачей** загружает все общие файлы `docs/`. **Дублирование** содержимого в
skill-файлах и планах **запрещено** — ссылаться, а не копировать.

| Файл | Содержание | Владелец |
|---|---|---|
| `docs/vision_and_scope.md` | SMART-цели, KPI, персоны, US, MoSCoW | BA |
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

Дополнительно (не дублировать): `.planning/STATE.md`, `.planning/COVERAGE.md`, `.planning/ROADMAP.md`,
`docs/wiki/`.
## 11.2 Кто владеет и кто обновляет

- **Владелец** файла (BA/SA/ARCH/QA) отвечает за его актуальность и структуру.
- **Любой** агент, чьё изменение затрагивает содержимое, обязан обновить файл (docs-sync §8.2) —
  через владельца или напрямую с уведомлением владельца.
- Конфликт версий docs решает `team-lead` (Accountable).
## 11.3 Запрет дублирования

- Skill-файлы (`.cursor/project_skills/*`) ссылаются на `docs/`, не копируют определения.
- Планы (`PLAN.md`) ссылаются на AC/требования по ID (US-Mxx, FR-xx), не переписывают их.
- Если факт есть в `docs/` — используй ссылку, а не копию (копия устареет).

### Good / Bad — база знаний

**Good:** `sa` в спеке пишет «реализует US-M03 (см. `vision_and_scope.md`)» и добавляет только новые
FR/API. Единый источник истины.

**Bad:** `sa` копирует текст US-M03 в спеку, `ba` позже правит оригинал — две расходящиеся версии,
никто не знает, какая верна.

### Чек-лист базы знаний (§11)

- [ ] Все `docs/` загружены перед задачей.
- [ ] Изменения синхронизированы с владельцем файла.
- [ ] Нет дублирования — только ссылки по ID.
- [ ] `STATE.md` отражает актуальный статус.

---

# 12. Антипаттерны командной работы

| # | Антипаттерн | Почему плохо | Правильно |
|---|---|---|---|
| 1 | Конфликтующие правки (два агента — один файл) | Постоянные merge-конфликты, потеря работы | Развести по файлам/модулям; общий контракт первым коммитом (§4) |
| 2 | Коммит без регрессии / one-shot без цикла | В baseline попадает поломка | Цикл качества новой + регрессия старого до green (§8.1, rule 56) |
| 2a | Делегирование без шаблона §5.3.1 | Субагент без контекста, drift scope | Только шаблон роли/контекста/формата/условий (§5.3) |
| 3 | Коммит без docs-sync | Docs расходятся с кодом, drift | Обновить STATE + доменный doc в той же сессии (§8.2) |
| 4 | Handoff без `verified: true` | Непроверенный артефакт в baseline | Персональный верификатор → verified → дальше (§7) |
| 5 | Широкий несогласованный scope | Агент трогает лишнее, ломает соседей | `team-lead` задаёт границы; расширение = эскалация (§5) |
| 6 | «Зелёные тесты» без test-triage | Маскируется реальный баг | test-triage: функционал → код vs тест (§8.3) |
| 7 | Дублирование знаний в skill/план | Расходящиеся версии требований | Ссылка по ID на `docs/` (§11) |
| 8 | Верификатор ставит verified на слово | Теряется независимость проверки | Воспроизвести сценарий + лог регрессии (§7) |
| 9 | Огромная неатомарная задача (>300 строк) | Невозможно верифицировать/откатить атомарно | Дробить на атомарные планы (§6) |
| 10 | Прямой мерж в `master` без code-reviewer | Непроверенный diff в prod-ветке | code-reviewer → team-lead → merge-ready → пользовательский rebase-merge (§3, `50-git.mdc`) |
| 11 | Пропуск `team-lead`-приёмки | Нет единой ответственности за фазу | Финальная приёмка `team-lead` обязательна (§5) |
| 12 | Секреты/DEV_MODE в коммите | Security-инцидент, prod-риск | git-scan diff; запреты `60-tz-constraints.mdc`, `50-git.mdc` |
| 13 | Параллельная правка точки композиции (`main.go`) | Гарантированный конфликт | Wiring отдаётся одному агенту (§4) |
| 14 | Handoff без «что проверить» | Верификатор буксует, догадки | Полный формат handoff (§9) |
| 15 | Запуск >3 субагентов одновременно | Превышен runtime cap, хаос контекста | root + ≤3 субагента, остальное — следующая волна (§4) |

### Финальный чек-лист team-workflow.md

- [ ] `team-lead` делегирует **только** по шаблону §5.3.1; в условиях — цикл качества + регрессия.
- [ ] `team-lead` (GPT-5.6 Sol / Max) оркестрирует и принимает; downstream — explicit Terra с task-routed effort по `model-routing.md`.
- [ ] Поток фичи: product → ba → sa → (arch) → (designer) → be∥fe → qa → review → team-lead → docs-sync.
- [ ] Волны root + ≤3 субагента, разведены по файлам, sync между волнами.
- [ ] Каждый артефакт → персональный верификатор → `verified: true` при green + лог.
- [ ] Ворота: цикл качества новой фичи, регрессия старого, docs-sync, test-triage, root Compose service.
- [ ] Handoff полный; `STATE.md` обновлён; знания не дублируются.

---

# Приложение A. Быстрый справочник ролей (шпаргалка)

| Роль | Триггер подключения | Отдаёт | Кому |
|---|---|---|---|
| `product` | Новая инициатива/приоритизация | scope + MoSCoW | ba |
| `ba1/ba2` | Нужны требования | US + AC | sa (через ba-ver) |
| `sa1/sa2` | Нужна спека/API | FR/NFR + API + ERD | arch/be/fe (через sa-ver) |
| `arch` | Архитектурно значимо | C4 + паттерны | be/fe (через arch-ver) |
| `designer` | Есть UI | DESIGN-SPEC | fe (через designer-ver) |
| `be1/be2` | Backend-код | код + тесты | be-ver → qa |
| `fe1/fe2` | Frontend-код | UI + тесты | fe-ver → qa |
| `qa-manual` | Приёмка/TC | тест-кейсы | qa-manual-ver |
| `qa-auto1/2` | Автотесты | тесты + coverage | qa-auto-ver |
| `code-reviewer` | Pre-merge | ревью diff | team-lead |
| `team-lead` | Всегда (оркестр.) | приёмка фазы | docs-sync/пользователь |
| `devops-engineer` | Инфра/CI/deploy | pipeline/деплой | team-lead |
| `security-auditor` | Security | находки + fix-план | be/fe/team-lead |
| `debugger` | Неясный баг | root cause + план | be/fe |
| `docs-writer` | Docs bulk | обновлённые docs | team-lead |

# Приложение B. Онбординг новой сессии (первые шаги любого агента)

1. Прочитай `.planning/STATE.md` — где сейчас проект, что pending, блокеры.
2. Загрузи `docs/*` (§11) — единая база знаний.
3. Прочитай свой skill (`.cursor/project_skills/<ROLE>_skill.md`) и этот playbook.
4. Уточни у `team-lead` границы задачи (файлы, контракт, критерии), если делегируют.
5. Работай атомарно → регрессия → docs-sync → верификатор → handoff.

# Приложение C. Глоссарий

| Термин | Значение |
|---|---|
| Волна (Wave) | Root + до 3 параллельных субагентов (≤4 active slots total) |
| Атомарный коммит | Изменение ≤200–300 строк, целиком проходящее регрессию |
| Верификатор | Независимый агент, проверяющий артефакт исполнителя |
| Ворота (gate) | Обязательный чек перед handoff (регрессия/docs/root Compose service) |
| Handoff | Передача работы следующему агенту (формат §9) |
| Baseline | Проверенное состояние проекта после verified + green gate |
| Docs-sync | Обновление docs в той же сессии после изменения |
| Test-triage | Протокол: функционал → код vs тест перед фиксом |
| Accountable (A) | Роль, отвечающая за результат (всегда team-lead на приёмке) |

# Приложение D. Шаблоны handoff для каждого перехода

Готовые шаблоны — копируй и заполняй. Все содержат 5 полей формата §9.
## D.1 product → ba

```
Что сделано: приоритизирован scope фичи "Sprint Backlog" (Must Have, US-M11 расширение).
  Бизнес-ценность: тимлид Дмитрий планирует спринт за минуты вместо часов.
Что проверить (ba): достаточно ли контекста для US/AC; покрыты ли персоны Дмитрий/Алексей.
Артефакты: запись в vision_and_scope.md (раздел MoSCoW), ссылка на конкурентный анализ.
Следующий шаг: ba1 пишет US + AC, ba2 — edge-cases; затем ba-ver.
Статус ворот: docs-sync done (vision_and_scope), verified: n/a (product-ver проверит scope).
```
## D.2 ba → sa

```
Что сделано: US-M11.1 "Backlog с drag-reorder" + 6 AC (создание, сорт, WIP, права).
Что проверить (sa): все ли AC тестируемы; нужны ли новые сущности/эндпоинты.
Артефакты: vision_and_scope.md §Must, таблица AC.
Следующий шаг: sa1 — API-контракт, sa2 — ERD (Sprint/BacklogItem); затем sa-ver.
Статус ворот: docs-sync done, verified: true (ba-ver 2026-07-02 21:10).
```
## D.3 sa → arch / designer

```
Что сделано: API /projects/:id/backlog (GET/PATCH reorder), ERD BacklogItem(version, order).
Что проверить (arch): паттерн reorder (optimistic lock), нет layering-нарушений в дизайне API.
  (designer): UX drag-reorder, состояния, a11y drag с клавиатуры.
Артефакты: system_specification.md §Backlog, wiki/API.md обновлён.
Следующий шаг: arch — решение по reorder-стратегии; designer — DESIGN-SPEC.
Статус ворот: docs-sync done, verified: true (sa-ver 2026-07-02 21:40).
```
## D.4 designer → fe

```
Что сделано: DESIGN-SPEC-backlog.md — раскладка, токены, 5 состояний, a11y drag-keyboard.
Что проверить (fe): реализовать по спеке; только className для Shadcn; drag доступен с клавиатуры.
Артефакты: DESIGN-SPEC-backlog.md (8 разделов, ASCII-раскладка, дерево компонентов).
Следующий шаг: fe1 — список + drag, fe2 — панель фильтров (разные файлы).
Статус ворот: docs-sync done (SHORTCUTS.md — hotkeys), verified: true (designer-ver 2026-07-02 22:05).
```
## D.5 be → be-ver / fe → fe-ver

```
Что сделано: BacklogService.Reorder с version-lock; handler + route; unit+integration тесты.
Что проверить (be-ver): layering (repo interface), 409-обработка, security (RBAC reorder),
  регрессия backend green.
Артефакты: backlog/service.go, handler.go, testframework/tracker/backlog_test.go, лог go test.
Следующий шаг: be-ver верифицирует; при verified → fe берёт контракт.
Статус ворот: regression PASS (race 600s), docs-sync done, verified: нет.
```
## D.6 qa → team-lead (финальная приёмка)

```
Что сделано: 12 ручных TC (все PASS), 8 backend integration + 1 e2e Playwright spec (все green).
Что проверить (team-lead): соответствие AC US-M11.1, покрытие edge-cases, coverage не упал.
Артефакты: test_plan.md §Backlog, COVERAGE.md (tracker 82.4%, 2026-06-22 Docker baseline), лог green root Compose service.
Следующий шаг: team-lead финальная приёмка → docs-sync → merge-ready → пользовательский rebase-merge.
Статус ворот: archived legacy gate green (build/test/lint/security/e2e-smoke), verified: true (qa-*-ver 2026-07-02 23:30).
```

### Чек-лист шаблонов (Прил. D)

- [ ] Каждый handoff содержит 5 полей.
- [ ] Статус ворот заполнен реальными данными (лог, дата verified).
- [ ] Следующий шаг называет конкретного агента.

---

# Приложение E. Полный worked-example: фича «Sprint Backlog»

Сквозной пример от идеи до мержа, чтобы показать взаимодействие всех ролей и волн.
## E.1 Таймлайн волн

```
День 1
  Wave 1 (аналитика/дизайн, параллельно, root + ≤3 субагента):
    ├─ ba1: US-M11.1 + AC ──────────────► ba-ver ✅
    ├─ sa1: API /backlog ───────────────► (ждёт ba-ver)
    └─ product: уточняет приоритет
  Sync: team-lead проверяет ba-ver verified → запускает sa

День 1-2
  Wave 2 (спека + арх + дизайн):
    ├─ sa1: API-контракт ───► sa-ver ✅
    ├─ sa2: ERD BacklogItem ► sa-ver ✅
    ├─ arch: reorder-стратегия ► arch-ver ✅
    └─ designer: DESIGN-SPEC ► designer-ver ✅
  Sync: контракт зафиксирован (types.ts + API.md)

День 2-3
  Wave 3 (реализация, разные файлы):
    ├─ be1: service.go + handler.go ──► be-ver ✅
    ├─ be2: ent/schema + repository ──► be-ver ✅
    ├─ fe1: List.tsx + drag ──────────► fe-ver ✅
    └─ fe2: Filters.tsx ──────────────► fe-ver ✅
  Sync: green root Compose services на объединении

День 3
  Wave 4 (качество):
    ├─ qa-auto1: backend integration ─► qa-auto-ver ✅
    ├─ qa-auto2: e2e spec ────────────► qa-auto-ver ✅
    └─ qa-manual: 12 TC ──────────────► qa-manual-ver ✅

День 3
  Wave 5 (приёмка):
    code-reviewer ✅ → team-lead приёмка ✅ → docs-sync → merge-ready
      → пользователь делает rebase-merge в master
```
## E.2 Кто какой файл трогает (карта без конфликтов)

| Агент | Файлы | Волна |
|---|---|---|
| ba1 | `docs/vision_and_scope.md` | 1 |
| sa1 | `docs/system_specification.md`, `docs/wiki/API.md` | 2 |
| sa2 | `backend/ent/schema/backlog_item.go` (спека, не код) | 2 |
| arch | `docs/architecture.md` (раздел reorder) | 2 |
| designer | `DESIGN-SPEC-backlog.md`, (при токенах) `app.css` | 2 |
| be1 | `backend/internal/backlog/service.go`, `handler.go` | 3 |
| be2 | `backend/ent/schema/backlog_item.go`, `backlog/ent_repository.go` | 3 |
| fe1 | `frontend/src/components/backlog/List.tsx` | 3 |
| fe2 | `frontend/src/components/backlog/Filters.tsx` | 3 |
| qa-auto1 | `testframework/tests/api/tracker/test_backlog.py` | 4 |
| qa-auto2 | `testframework/tests/e2e/tracker/test_backlog.py` | 4 |
| qa-manual | `docs/test_plan.md` | 4 |

> Единственная общая точка — `types.ts` (frontend) и API-контракт: фиксируются **первым** коммитом
> волны 3, дальше конфликтов нет.
## E.3 Что происходит при провале верификации

Пример: `be-ver` находит `*ent.Client` в `service.go` (нарушение layering).

```
be-ver: verified: false
  Findings: [CRITICAL] service.go импортирует ent.Client напрямую (layering bypass).
  Вердикт: вынести доступ к данным в repository interface.
     │
     ▼ возврат be1 (не дальше по потоку)
be1: рефакторит на BacklogRepository interface → регрессия green → повторный handoff
     │
     ▼
be-ver: перепроверяет → verified: true 2026-07-02 22:50
     │
     ▼ поток продолжается (fe берёт контракт)
```

---

# Приложение F. RACI для дополнительных сценариев
## F.1 Миграция БД (Ent schema change)

| Этап | be | arch | devops | qa | team-lead |
|---|---|---|---|---|---|
| Изменение схемы Ent | R | C | I | I | A |
| Atlas миграция | R | C | C | I | A |
| Проверка UTC/version/UUID (tz_constraints) | C | R | I | I | A |
| Регрессия + бэкап-совместимость | C | I | C | R | A |
## F.2 Обновление зависимостей (deps)

| Этап | be/fe | devops | security-auditor | qa | team-lead |
|---|---|---|---|---|---|
| Bump версии | R | C | I | I | A |
| Проверка CVE/lock | I | C | R | I | A |
| Регрессия (build+test) | C | C | I | R | A |
| Docs (stack_and_tools) | R | C | I | I | A |
## F.3 i18n / текстовые изменения

| Этап | fe | ba | qa | team-lead |
|---|---|---|---|---|
| Правка строк ru/en | R | C | I | A |
| Проверка соответствия US | I | R | I | A |
| Регрессия UI (нет сломанной раскладки) | C | I | R | A |
## F.4 Документация (без кода)

| Этап | docs-writer | владелец doc | team-lead |
|---|---|---|---|
| Обновление wiki/README | R | C | A |
| Сверка с кодом (verify) | R | C | A |
| Публикация | R | I | A |

### Чек-лист доп-сценариев (Прил. F)

- [ ] Миграции проверены на tz_constraints (UTC/version/UUID/soft-delete).
- [ ] Deps: CVE-скан + lock + регрессия.
- [ ] i18n: соответствие US + отсутствие сломанной раскладки.
- [ ] Docs: сверка с кодом перед публикацией.

---

# Приложение G. Эскалационная матрица

Когда агент **обязан** остановиться и эскалировать, а не решать сам.

| Ситуация | Кому эскалировать | Действие |
|---|---|---|
| Изменение scope (задача выросла) | `team-lead` → пользователь | Стоп, согласовать перед продолжением |
| Деструктивное действие (drop, force-push, hard delete) | пользователь | Явное подтверждение обязательно |
| Конфликт требований в `docs/` | владелец doc → `team-lead` | Зафиксировать одну версию |
| Провал верификации (CRITICAL/HIGH) | исполнитель | Возврат с findings, не мержить |
| Падение регрессии | `debugger` + `/test-triage` | Root cause до фикса |
| Security-находка | `security-auditor` | Приоритет, fix-план, проверка запретов |
| Prod-блокер (PG-03 rotation и т.п.) | пользователь/ops | Не деплоить до разрешения |
| Конфликт файлов между агентами | `team-lead` | Определить базовый коммит, второй ребейзит |
| Нехватка контекста для решения | `team-lead` → пользователь | Задать уточняющий вопрос, не домысливать |
| Запрос коммита/пуша | пользователь | Только по явному запросу (rule 50-git) |

**Правило:** эскалация — это не «слабость», а защита baseline. Лучше остановиться и спросить, чем
внести несогласованное изменение в prod-ветку.

---

# Приложение H. Definition of Done по типам артефактов

| Артефакт | Done, когда |
|---|---|
| **User Story (ba)** | Есть US + тестируемые AC, покрыты персоны, MoSCoW-приоритет; `ba-ver` verified; docs-sync (vision_and_scope) |
| **Спека/API (sa)** | FR/NFR + API-контракт + ERD; покрывают все AC; `sa-ver` verified; docs-sync (system_specification + wiki/API) |
| **Архитектура (arch)** | C4/паттерн/выбор библиотек; SOLID/KISS; реализуемо стеком; `arch-ver` verified; docs-sync (architecture) |
| **Дизайн (designer)** | DESIGN-SPEC полная (8 разделов), токены реальны, a11y AA; `designer-ver` verified; см. `design.md` §8 |
| **Backend-код (be)** | Layering соблюдён, security по guidelines, тесты, регрессия green; `be-ver` verified; docs-sync |
| **Frontend-код (fe)** | По DESIGN-SPEC, Shadcn className-only, a11y (axe) + Lighthouse, регрессия green; `fe-ver` verified; docs-sync |
| **Тесты (qa)** | TC покрывают AC; автотесты green; coverage не упал; test-triage при фиксах; `qa-*-ver` verified; COVERAGE.md |
| **Фаза (team-lead)** | Все артефакты verified, применимые root Compose services green, AC/success criteria выполнены, docs-sync, code-reviewer прошёл |

### Универсальный DoD-гейт

- [ ] Артефакт verified персональным верификатором.
- [ ] Регрессия/root Compose service green (лог приложен).
- [ ] Docs-sync выполнен (STATE + доменный doc).
- [ ] Handoff передан следующему агенту.

---

# Приложение I. Контекст-бюджет и промптинг субагентов (для `team-lead`)

`team-lead` (GPT-5.6 Sol / Max) держит фазу через `STATE.md`/handoff-сводки и выбирает downstream
runtime по capability/cost/risk. Субагенты **не видят** историю чата. Правила промптинга:

1. **Самодостаточность.** В промпт субагента вложи: задачу, ссылки на нужные `docs/` (по разделам),
   границы файлов, API-контракт/типы, критерии приёмки, ожидаемый формат handoff.
2. **Границы.** Явно скажи «трогай только файл X и его тест», чтобы избежать конфликтов с параллельным
   агентом.
3. **Ожидаемый вывод.** Проси вернуть handoff по формату §9 + лог регрессии.
4. **Не перегружай.** Не вкладывай весь `docs/` целиком в каждый промпт — ссылайся на разделы. Экономь
   контекст субагента для работы, а не для чтения лишнего.
5. **Параллельность.** Запускай независимые субагенты одним батчем (до 3 вместе с root = 4 active slots). Зависимые — следующей волной.

### Пример промпта субагенту (be1)

```
Роль: be1. Реализуй BacklogService.Reorder.
Контекст: docs/system_specification.md §Backlog (API PATCH /backlog reorder),
  docs/code_style_guide.md (Go-стиль, error wrapping), docs/security_guidelines.md (RBAC).
Границы: только backend/internal/backlog/service.go + service_test.go. НЕ трогай handler.go (это be2? нет — твой),
  НЕ трогай ent/schema (это be2), НЕ трогай cmd/server/main.go.
Контракт: тип BacklogReorderRequest уже в types (первый коммит волны).
Критерии: optimistic lock по version, 409 при конфликте, repository через интерфейс (не ent.Client).
Регрессия: task-owned Docker `backend-unit` по `backend/tests/backlog/**` + применимый Python boundary job.
Верни: handoff по формату §9 + лог go test.
```

---

# Приложение J. Git-интеграция workflow

Из `.cursor/rules/50-git.mdc`. Команда следует этим правилам при мерже результатов.
## J.1 Ветки

```
main/master — protected, production-ready
feat/<module>/<desc> — фичи
fix/<module>/<desc> — багфиксы
refactor/<module>/<desc> — рефакторинг
chore/<desc> — deps/tooling
hotfix/<id>/<desc> — критичные prod-фиксы
```
## J.2 Коммиты (Conventional Commits)

```
<type>(<scope>): <description>
type: feat|fix|refactor|test|docs|chore|perf|ci
scope: auth|tracker|messenger|tms|calls|dashboard|notification|search|ws|frontend|infra|ent|...
```

Один атомарный коммит = одна задача плана (§6). Пример: `feat(tracker): add backlog reorder with version lock`.
## J.3 Порядок и правила

- Агенты могут атомарно коммитить и пушить фича-ветки после ворот; push/merge в
  `master` выполняет только пользователь.
- Pre-commit: build + test + lint + security-scan diff (`grep secret|password|token|key`).
- PR ссылается на задачу PLAN; все CI-чеки green; `code-reviewer` approve;
  атомарные коммиты сохраняются для пользовательского rebase-merge в `master`.
- Никогда: `git config`, force-push в master, `--no-verify`, коммит `.env`/секретов, `DEV_MODE=true`.

### Чек-лист git (Прил. J)

- [ ] Ветка названа по конвенции; коммит — Conventional Commits.
- [ ] Один атомарный коммит на задачу; pre-commit проверки прошли.
- [ ] Нет секретов/DEV_MODE в diff; PR ссылается на PLAN.
- [ ] Ветка merge-ready; пользователь лично выполняет rebase-merge в `master`.

---

# Приложение K. FAQ командной работы

**Кто принимает финальное решение по фазе?** Всегда `team-lead` на GPT-5.6 Sol / Max. Для высокорисковых
частей он назначает отдельный сильный downstream-проход/verifier, но Accountable остаётся сам.

**Можно ли пропустить верификатора для мелочи?** Нет для user-visible/prod-поведения. Для чистого
рефакторинга без изменения поведения — минимум `STATE.md` one-liner + регрессия (rule 55 исключения).

**be1 закончил, be2 ещё нет — можно мержить be1?** Да, если be1 атомарен, verified и не ломает регрессию.
Фундаментальные артефакты мержатся раньше зависимых (§4.4).

**Что если два верификатора расходятся?** Эскалация `team-lead` — он финальная инстанция.

**Сколько агентов запускать?** До 3 субагентов одновременно плюс root. Остальные — следующей волной (§4).

**Где смотреть «что дальше»?** `.planning/STATE.md` (текущий статус) + `gsd-next`/`gsd-progress`.

**Нужно ли грузить все `docs/` каждый раз?** Да, перед задачей (§11). Но в промпт субагента вкладывай
ссылки на разделы, а не полные копии (Прил. I).

**Кто обновляет `STATE.md`?** Тот, кто закрывает волну/задачу, через `/handoff-sync`. Владелец
статуса — `team-lead`.

**Можно коммитить без docs-sync?** Нет для user-visible/prod. Docs-sync — в той же сессии (§8.2).

---

# Приложение L. Детальные чек-листы верификаторов

Расширение §7.3. Каждый верификатор проходит **все** пункты; `verified: true` только при 0 FAIL по
CRITICAL/HIGH + приложенный лог регрессии.
## L.1 `product-ver`

- [ ] Scope соответствует бизнес-цели (G-xx) и не выходит за границы (tz_constraints FORBIDDEN).
- [ ] Приоритет MoSCoW обоснован; фича не дублирует существующую.
- [ ] Учтены персоны и их JTBD (user_needs_analysis).
- [ ] Нет out-of-scope фич без ADR (SSO, multi-org, billing и т.п.).
## L.2 `ba-ver`

- [ ] Каждая US имеет тестируемые AC (проверяемо, не «удобно/красиво»).
- [ ] AC покрывают happy path + edge-cases + права (RBAC).
- [ ] US привязаны к персонам и целям.
- [ ] MoSCoW-приоритет проставлен; нет противоречий с уже принятыми US.
- [ ] docs-sync: vision_and_scope обновлён.
## L.3 `sa-ver`

- [ ] Все AC покрыты FR/NFR (матрица покрытия).
- [ ] API-контракт полон: методы, коды ошибок (`{error, request_id}`), пагинация (offset/cursor).
- [ ] ERD корректен: UUID PK, `version` для optimistic lock, UTC, soft-delete.
- [ ] Rate limiting/размеры файлов соответствуют tz_constraints.
- [ ] docs-sync: system_specification + wiki/API обновлены.
## L.4 `arch-ver`

- [ ] Решение соответствует модульному монолиту и Clean Architecture (handler→service→repo).
- [ ] SOLID/KISS; нет over-engineering.
- [ ] Реализуемо текущим стеком (Go/Fiber/Ent/React/TanStack) без запрещённых замен.
- [ ] Выбор библиотек обоснован; нет запрещённых (minio-go, gorilla, net/http для WS и т.п.).
- [ ] Security-архитектура учтена (SSRF, secrets, WS RBAC).
- [ ] docs-sync: architecture обновлён.
## L.5 `designer-ver`

См. полный чек-лист в `.cursor/playbooks/design.md` §7. Кратко:

- [ ] Токены реальны (`app.css`), нет HEX-хардкода.
- [ ] Контраст AA (4.5:1 / 3:1) рассчитан; focus/keyboard/44px/aria заданы.
- [ ] 5 состояний + loading/empty/error; responsive breakpoints.
- [ ] Shadcn className-only, Radix не тронут, нет `!important`/inline.
## L.6 `be-ver`

- [ ] Layering: нет `*ent.Client` в service/handler; доступ к данным через repository interface.
- [ ] Error handling: wrap `%w`, нет `panic()` в prod-путях, `ent.IsNotFound` обработан.
- [ ] Security: нет `http.DefaultClient` для user-URL, `io.LimitReader`, `subtle.ConstantTimeCompare`.
- [ ] Context первым аргументом; `slog` с контекстом, не `fmt.Println`.
- [ ] Тесты покрывают happy+error+edge; регрессия backend green (лог).
- [ ] docs-sync выполнен.
## L.7 `fe-ver`

- [ ] Реализация соответствует DESIGN-SPEC; токены/состояния на месте.
- [ ] a11y: axe без critical; focus виден; keyboard-навигация; aria/labels.
- [ ] Lighthouse (perf/a11y) в норме; нет XSS (dangerouslySetInnerHTML без sanitize).
- [ ] Shadcn className-only; hooks-правила; стабильные keys; нет `key={index}`.
- [ ] Регрессия frontend green (tsc/vitest/lint, лог).
- [ ] docs-sync (SHORTCUTS/QUICKSTART при hotkey/flow).
## L.8 `qa-manual-ver`

- [ ] TC покрывают все AC (traceability AC → TC).
- [ ] Есть негативные и граничные TC, проверка прав.
- [ ] Ручная приёмка воспроизводима по шагам.
- [ ] docs-sync: test_plan обновлён.
## L.9 `qa-auto-ver`

- [ ] Выбор unit/contract/API/integration/E2E/load следует changed risk и observable boundary; форма пирамиды не quota.
- [ ] Тесты детерминированы; historical coverage/parity не migration gate.
- [ ] Test-triage применён при фиксах (ROOT зафиксирован).
- [ ] Genuine units в `backend/tests/**`/`frontend/tests/**`; non-unit QA в root Python `testframework/**`.
- [ ] Применимые target services воспроизведены Docker-only через `docker compose -f docker-compose.local.yml run --rm <service>` на exact revision;
      отсутствующий launcher/scaffold = `NOT_READY`, не legacy fallback.

---

# Приложение M. Исторические метрики качества (не текущие ворота)

Снимок 2026-06-22..24 сохранён только для истории. Он не определяет current
acceptance и не разрешает legacy suite, parity или coverage migration gate.
Источник текущего статуса — `.planning/STATE.md`; current gate — root Compose service.

| Метрика | Значение |
|---|---|
| Milestone | v4 Super-app ✅ + Wave C audit ✅ |
| COV overall | 83.0% (2026-06-22 Docker baseline) |
| Security-пакеты ≥80% | 6/6 (auth 82.4 / ws 87.7 / rbac 84.7 / audit 80.5 / calls 80.2 / storage; 2026-06-22 Docker baseline) |
| Product score | 87–89/100 |
| REST API | 286+ endpoints |
| Ent schemas | 82 |
| Backend tests (testframework) | 515 files (measured 2026-07-27; was 499 on 2026-07-26), 0 в internal/ |
| Vitest | 99/99 pass |
| Playwright E2E | 20 spec |
| Branch | master |
| Prod-блокер | PG-03 secret rotation (manual ops, FA-06) |

**Current acceptance:** только применимые founder-greenfield target jobs с
traceability/risk oracle, exact Docker provenance и cleanup PASS. Пока target
launcher/scaffold отсутствует, verdict `NOT_READY`. Числа выше не являются воротами.

---

# Приложение N. Жизненный цикл артефакта

```
[draft] ──исполнитель делает──► [regression] ──green──► [docs-synced]
                                     │FAIL                    │
                                     ▼                        ▼
                                 test-triage            [in-verification] ──верификатор──►
                                                              │
                              ┌───────────────────────────────┴───────────┐
                              │CRITICAL/HIGH findings          0 findings │
                              ▼                                            ▼
                          [returned] ──fix──► (назад в regression)    [verified:true]
                                                                           │
                                                                           ▼
                                                              [in-root-compose-service] ──green──►
                                                                           │
                                                                           ▼
                                                              [reviewed] (code-reviewer)
                                                                           │
                                                                           ▼
                                                              [accepted] (team-lead)
                                                                           │
                                                                           ▼
                                                              [merged → baseline]
```

**Правило перехода:** артефакт не переходит в следующее состояние, пока текущее не закрыто полностью.
Возврат (`returned`) всегда идёт к исполнителю, а не «дальше по потоку».

---

# Приложение O. Сквозные ответственности (cross-cutting)

Некоторые аспекты касаются всех ролей. Владелец задаёт стандарт, все соблюдают.

| Аспект | Владелец стандарта | Соблюдают | Проверяют |
|---|---|---|---|
| Security (OWASP, secrets, RBAC) | `security-auditor` / arch | be, fe, devops | be-ver, arch-ver, security-auditor |
| Доступность (WCAG AA) | `designer` | fe | designer-ver, fe-ver |
| Производительность (Lighthouse, запросы) | arch | be, fe | fe-ver, be-ver |
| i18n (ru/en) | ba | fe | ba-ver, fe-ver |
| Docs-sync | все (rule 55) | все | верификаторы + team-lead |
| Регрессия | все (rule 56) | все | верификаторы |
| Код-стиль | arch (`code_style_guide`) | be, fe | be-ver, fe-ver, code-reviewer |
| tz_constraints (mandatory/forbidden) | arch | все | arch-ver, code-reviewer |

**Правило:** сквозной аспект не «чья-то чужая забота» — каждый агент отвечает за него в своём слое.
Например, security — не только `security-auditor`: `be` не должен писать `http.DefaultClient` для
user-URL, даже если аудит ещё не пришёл.

### Чек-лист сквозных аспектов (Прил. O)

- [ ] Security соблюдён в своём слое (не ждать аудита).
- [ ] a11y учтён в UI (не «потом добавим»).
- [ ] Производительность (lazy-load тяжёлых модулей, staleTime запросов).
- [ ] i18n: строки вынесены, ru/en покрыты.
- [ ] Код-стиль и tz_constraints соблюдены.

---

# Приложение P. Матрица «кто с кем общается» (communication)

Прямые каналы handoff между ролями (кто кому передаёт артефакт):

```
product ──► ba ──► sa ──► arch ──► designer ──┐
                    │        │                 │
                    └────────┴─────────────────┼──► be1/be2 ──► be-ver ──┐
                                               └──► fe1/fe2 ──► fe-ver ──┤
                                                                         ▼
                                       qa-manual ──► qa-manual-ver ──► qa-auto1/2 ──► qa-auto-ver
                                        (TC из бизнес-доки; юниты пишут be/fe сами)
                                                                         │
                                                                  code-reviewer
                                                                         │
                                                                   team-lead (приёмка)
                                                                         │
                                                                    docs-writer / docs-sync

support (в любой момент по эскалации):
  debugger ◄── (падение/баг) ── be/fe/qa
  security-auditor ◄── (security) ── be/fe/arch/team-lead
  devops-engineer ◄── (инфра/deploy) ── team-lead
```

**Правило:** общение идёт через артефакты и handoff, а не «мимо потока». Если `fe` нужен другой
API-контракт — это возврат к `sa` через `team-lead`, а не тихая правка бэкенда самим `fe`.

---

# Приложение Q. Антипаттерны — детальные примеры

Дополнение к таблице §12. Конкретные good/bad-сценарии.
## Q.1 Конфликтующие правки

**Bad:** `team-lead` даёт be1 и be2 задачу «доделать спринты», не разграничив файлы. Оба правят
`sprint/service.go` — постоянные конфликты, be2 перезаписывает be1.

**Good:** be1 → `service.go`, be2 → `ent_repository.go` + `ent/schema/sprint.go`. Общий `SprintDTO` —
первым коммитом. Ветки мержатся чисто.
## Q.2 Коммит без регрессии

**Bad:** fe1 «быстро поправил» компонент и передал fe-ver без прогона `tsc`/`vitest`. Верификатор
находит сломанный тип, цикл откатывается.

**Good:** fe1 прогнал `tsc --noEmit && vitest run` (green), приложил лог в handoff. fe-ver сразу
проверяет по существу.
## Q.3 «Зелёные тесты» без triage

**Bad:** упал тест `backlog_reorder` — qa-auto2 меняет assert, чтобы «прошло». Маскирует реальный баг
в reorder-логике.

**Good:** qa-auto2 применяет test-triage: воспроизводит по `docs/`, видит что код нарушает AC (порядок
не сохраняется), эскалирует be1 фикс кода, тест оставляет строгим. ROOT=code зафиксирован.
## Q.4 Широкий scope

**Bad:** be1 в задаче про backlog «заодно» рефакторит auth-модуль. Ломает регрессию auth, задевает
работу другого агента.

**Good:** be1 замечает проблему в auth, но не трогает — создаёт заметку/todo и эскалирует `team-lead`.
Scope волны остаётся чистым.
## Q.5 Пропуск верификатора

**Bad:** be1 сам себе «проверил» и передал в qa. Layering-нарушение просочилось в baseline, всплыло
на релизе.

**Good:** be1 → be-ver (независимый) → findings → fix → verified. Нарушение поймано до baseline.
## Q.6 Дублирование знаний

**Bad:** sa скопировал текст US-M11 в спеку. ba позже уточнил US — две версии, разработчик читает
устаревшую копию.

**Good:** sa ссылается «реализует US-M11 (vision_and_scope.md)». Единый источник, нет дрейфа.

### Чек-лист антипаттернов (Прил. Q)

- [ ] Файлы разграничены; общий контракт первым коммитом.
- [ ] Регрессия перед handoff; test-triage перед фиксом.
- [ ] Scope не расползается; расширение — эскалация.
- [ ] Верификатор независим; знания по ссылке, не копией.

---

# Приложение R. Ритуал сессии (start / end)
## R.1 Начало сессии (любой агент)

1. Прочитать `.planning/STATE.md` — где проект, pending, блокеры.
2. Загрузить `docs/*` (§11).
3. Прочитать свой skill + этот playbook (+ `design.md` для дизайн/fe-ролей).
4. Уточнить границы задачи у `team-lead` (файлы, контракт, критерии).
5. `gsd-progress` / `gsd-next` — понять следующий логический шаг.
## R.2 Во время работы

- Атомарные изменения (≤200–300 строк).
- После каждого изменения — регрессия слоя.
- При падении теста — test-triage до фикса.
- Держаться границ файлов; не расширять scope.
## R.3 Конец сессии / волны

1. Регрессия green (лог).
2. Docs-sync: `STATE.md` + доменный doc.
3. Handoff по формату §9 следующему агенту/верификатору.
4. `/handoff-sync` — синхронизировать STATE/ROADMAP/COVERAGE.
5. Артефакты-мусор (coverage dumps) удалить (rule 45).
6. `team-lead`: собрать волну, применимые root Compose services, решить о мерже/следующей волне.

### Чек-лист ритуала (Прил. R)

- [ ] Старт: STATE + docs + skill + границы.
- [ ] Работа: атомарно + регрессия + triage + границы.
- [ ] Финиш: регрессия green + docs-sync + handoff + `/handoff-sync` + cleanup.

---

# Приложение S. Сводная памятка (одна страница)

```
РОЛИ:      team-lead(GPT-5.6 Sol/Max) | Terra(task-routed downstream; atomic code=XHigh)
ПОТОК:     product → ba → sa → (arch) → (designer) → be∥fe → qa → review → team-lead → docs-sync
ВОЛНЫ:     root + ≤3 субагента (4 slots total), разные файлы, sync между волнами
ВЕРИФ:     каждый артефакт → персональный -ver → verified:true при green+лог
ВОРОТА:    регрессия → docs-sync → test-triage → root Compose service (перед handoff)
HANDOFF:   что сделано / что проверить / артефакты / след.шаг / статус ворот
ДЕКОМП:    фаза ≤1-3д, план ≤200-300 строк = атомарный коммит, Depends on + success criteria
ЗНАНИЯ:    грузить docs/* перед задачей; ссылки по ID, не копии
ЭСКАЛАЦИЯ: scope/destructive/конфликт/prod-блокер → team-lead → пользователь
ЗАПРЕТЫ:   коммит без регрессии/docs; verified на слово; >300 строк; секреты; >3 субагента одновременно
```

---

> Связанные документы: `.cursor/REGULATION.md`, `.cursor/GSD-INTEGRATION.md`, `AGENTS.md`,
> `.cursor/rules/50-git.mdc`, `55-docs-sync.mdc`, `56-regression-verify.mdc`, `.cursor/playbooks/design.md`,
> `.planning/STATE.md`, `docs/README.md`.
