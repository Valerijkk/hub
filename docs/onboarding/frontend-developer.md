# Онбординг — Frontend-разработчик

> Роль: разработка frontend Hub на **React 19 + TypeScript 5.7 + Vite + TanStack (Router/Query) + Zustand + Tailwind CSS 4 + Shadcn/ui**.
> Главные playbooks: `.cursor/playbooks/frontend.md` + `.cursor/playbooks/design.md` · Индекс онбординга: [README.md](README.md)

---

## Кто ты и за что отвечаешь

- Реализуешь UI-фичи: компоненты, хуки, stores, роуты (`frontend/src/`: api/, components/, hooks/, modules/, routes/, stores/).
- **Genuine unit-тесты (Vitest + RTL) — твоя обязанность**, пишешь вместе с фичей в `frontend/tests/{domain}/` и support в `frontend/tests/support/`. Python QA-конвейер покрывает contract/API/integration/E2E по verified TC — не твои юниты.
- Работаешь строго в дизайн-системе: токены `--hub-*`, Shadcn-компоненты (менять только `className`, Radix-поведение не трогать), доступность WCAG.
- Работаешь в `codex/<task>`, созданной от refreshed `master`; не реализуешь напрямую в `master`. После Docker-only проверки и независимого review `team-lead` интегрирует routine accepted branch и удаляет её; security/auth, DB migrations, deploy/CI и ADR требуют founder review до integration.

Соседи по роли в AI-команде — агенты `fe1`/`fe2` (параллельные потоки), верификатор `fe-ver`, дизайн — `designer`/`designer-ver`.

## День 1

**Доступы и окружение** — общий чек-лист в [README.md](README.md). Дополнительно для FE:

- [ ] Docker CLI/Compose доступен; Node/Vite/tsc/Biome/Vitest/Playwright на Windows host не запускаются.
- [ ] Поднят полный стек: `docker compose --env-file .env.local -f docker-compose.local.yml up --build` → UI `http://localhost:8080`, dev-login (`DEV_MODE=true`) работает.
- [ ] Знаешь текущую runtime-правду: canonical `/qa-gate` target `frontend-unit`
  должен исполняться только через task-owned limited Docker outer launcher.
  Launcher пока `NOT_READY`, поэтому runtime остаётся `NOT_RUN`; `npm install`,
  `tsc`, Biome, Vitest, Vite и Playwright на host или через legacy runner не fallback.

**Что прочитать (в этом порядке):**

- [ ] `docs/vision_and_scope.md` — что за продукт и User Stories.
- [ ] `docs/architecture.md` — модули и как frontend общается с backend (REST + WebSocket).
- [ ] `docs/code_style_guide.md` — TS/React-правила: Rules of Hooks (все хуки до любого return), стабильные `key`, `type="button"`, optional chaining вместо `!`.
- [ ] **`.cursor/playbooks/frontend.md`** — главный рабочий документ: TanStack Router/Query, Zustand, производительность, a11y.
- [ ] **`.cursor/playbooks/design.md`** — дизайн-система `--hub-*`, формат дизайн-спеки, WCAG.
- [ ] `docs/wiki/SHORTCUTS.md` — горячие клавиши продукта (ты их поддерживаешь).
- [ ] `.cursor/playbooks/code-comments.md`, `git-workflow.md`.

## Неделя 1

- [ ] Пройди продукт как пользователь: логин → онбординг → трекер (Kanban, задачи) → мессенджер → TMS → дашборды → Ctrl+K поиск.
- [ ] Изучи один модуль насквозь: роут в `src/routes/` → компоненты → хуки/запросы (TanStack Query) → store (Zustand) → его genuine units в `frontend/tests/{domain}/`.
- [ ] Разберись с общими элементами: `render.tsx` (TestRenderer), Shadcn-обёртки, токены темы.
- [ ] **Первая задача (good-first):** мелкий UI-фикс или тест — например, состояние empty/skeleton для существующего списка или a11y-фикс (aria-метки, focus). Согласуй с тимлидом; если задача видима пользователю — сначала проверь, есть ли дизайн-спека.
- [ ] Первый handoff: task-ветка → `frontend-unit` evidence из canonical Docker launcher; пока launcher `NOT_READY`, честно передай `NOT_READY`/`NOT_RUN`, не заявляя green.

## Инструменты и команды

Единственный принимаемый execution path для genuine frontend units — target
`frontend-unit` через sole versioned `/qa-gate` outer launcher в task-owned
limited Docker. Python E2E будет запускаться отдельными targets
`python-e2e-smoke`/`python-e2e-full`, а TypeScript Playwright не портируется.
Пока launcher `NOT_READY`, runtime `NOT_RUN`; host npm/npx, прямой Docker/Compose
и legacy runners не допускаются как fallback.

**Правила, за которые заворачивает ревью:**

| Нельзя | Нужно |
|--------|-------|
| Хук после условного return | Все хуки до любого return |
| `key={index}` в списках | Стабильный `key={item.id}` |
| `user!.profile!.name` | `user?.profile?.name ?? '...'` |
| `!important` в CSS | Токены/утилиты Tailwind |
| Кастомный CSS вместо токенов | `--hub-*` переменные |
| Правка Radix-поведения Shadcn | Только `className` |

## Как работает AI-команда и как с ней взаимодействовать

Ростер: `.cursor/playbooks/README.md`. Что важно тебе:

- **Дизайн до кода:** новые экраны/компоненты делаются по дизайн-спеке агента `designer` (проверяет `designer-ver`). Нет спеки — запроси через тимлида, не выдумывай UI сам.
- **Кросс-ревью:** твой PR смотрит второй FE-поток (`fe1`/`fe2`), затем `fe-ver` (Rules of Hooks, a11y, дизайн-система, типобезопасность, Lighthouse). `verified: true` допустим только при `PASS` canonical Docker target с логом; `NOT_READY`/`NOT_RUN` не green.
- **Упал тест — triage сначала** (`.cursor/skills/test-triage/SKILL.md`): функционал по спеке → ROOT=code|test → потом фикс.
- **API-контракты** фиксируют `sa1`/`sa2` до старта — не меняй типы ответов в одностороннем порядке; расхождение с backend решается через спеку, а не хардкодом на клиенте.
- **Docs-sync:** новый UI-flow/hotkey → `docs/wiki/SHORTCUTS.md` + `QUICKSTART.md`; любой код → session note в `.planning/STATE.md` (rule 55).
- **Git:** назначенная `codex/<task>` от refreshed master + Conventional Commits (`feat(frontend): ...`); direct master implementation запрещён. После Docker-only evidence и независимого review routine merge/push/cleanup выполняет `team-lead`; security/auth, DB migrations, deploy/CI и ADR требуют founder review до integration. Параллельный поток не трогает твои файлы и наоборот.

## Чек-лист готовности

- [ ] Стек поднят, dev-login работает, продукт прощупан руками.
- [ ] Понимаю status boundary: launcher `NOT_READY`, runtime `NOT_RUN`; host/legacy fallback запрещён, green без canonical evidence не заявляется.
- [ ] Прочитаны: frontend.md, design.md, code_style_guide, SHORTCUTS.
- [ ] Понимаю дизайн-систему `--hub-*` и границы Shadcn.
- [ ] Знаю Rules of Hooks и таблицу «нельзя/нужно».
- [ ] Первый PR прошёл кросс-ревью + fe-ver и доведён до merge-ready.
- [ ] Знаю git-политику: `codex/<task>` от refreshed master; direct master implementation запрещён; routine integration выполняет team-lead после ворот, а security/auth, DB migrations, deploy/CI и ADR требуют founder review.
