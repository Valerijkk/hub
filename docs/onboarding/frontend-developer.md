# Онбординг — Frontend-разработчик

> Роль: разработка frontend Hub на **React 19 + TypeScript 5.7 + Vite + TanStack (Router/Query) + Zustand + Tailwind CSS 4 + Shadcn/ui**.
> Главные playbooks: `.cursor/playbooks/frontend.md` + `.cursor/playbooks/design.md` · Индекс онбординга: [README.md](README.md)

---

## Кто ты и за что отвечаешь

- Реализуешь UI-фичи: компоненты, хуки, stores, роуты (`frontend/src/`: api/, components/, hooks/, modules/, routes/, stores/).
- **Unit-тесты (Vitest + RTL) — твоя обязанность**, пишешь вместе с фичей в `testframework/stages/unit/frontend-test/domains/{domain}/`. QA-конвейер покрывает E2E и компонентные тесты по TC — не твои юниты.
- Работаешь строго в дизайн-системе: токены `--hub-*`, Shadcn-компоненты (менять только `className`, Radix-поведение не трогать), доступность WCAG.
- Доводишь ветку до merge-ready; мерж в `master` делает только владелец.

Соседи по роли в AI-команде — агенты `fe1`/`fe2` (параллельные потоки), верификатор `fe-ver`, дизайн — `designer`/`designer-ver`.

## День 1

**Доступы и окружение** — общий чек-лист в [README.md](README.md). Дополнительно для FE:

- [ ] Node 22 (как в CI; локально 20+ допустимо), Docker Compose v2.
- [ ] Поднят полный стек: `docker compose --env-file .env.local -f docker-compose.local.yml up --build` → UI `http://localhost:8080`, dev-login (`DEV_MODE=true`) работает.
- [ ] Frontend-проверки зелёные:

```bash
cd frontend
npm install
node scripts/link-testframework-node-modules.mjs   # перед Vitest/E2E — обязательно
npx tsc --noEmit
npm run lint          # Biome — gate
npm run test:run      # Vitest
```

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
- [ ] Изучи один модуль насквозь: роут в `src/routes/` → компоненты → хуки/запросы (TanStack Query) → store (Zustand) → его Vitest-тесты в `testframework/stages/unit/frontend-test/domains/`.
- [ ] Разберись с общими элементами: `render.tsx` (TestRenderer), Shadcn-обёртки, токены темы.
- [ ] **Первая задача (good-first):** мелкий UI-фикс или тест — например, состояние empty/skeleton для существующего списка или a11y-фикс (aria-метки, focus). Согласуй с тимлидом; если задача видима пользователю — сначала проверь, есть ли дизайн-спека.
- [ ] Первый PR: фича-ветка → tsc + lint + vitest зелёные → handoff.

## Инструменты и команды

```bash
cd frontend

# Быстрая регрессия слоя (перед каждым handoff)
npx tsc --noEmit && npm run lint && npx vitest run

# Один домен Vitest
npx vitest run ../testframework/stages/unit/frontend-test/domains/auth

# Полная сборка (проверяет импорты, Tailwind 4, Router codegen)
npm run build

# E2E (нужен backend для full)
npm run test:e2e:smoke   # 3 spec, быстрые
npm run test:e2e         # 20 spec
```

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
- **Кросс-ревью:** твой PR смотрит второй FE-поток (`fe1`/`fe2`), затем `fe-ver` (Rules of Hooks, a11y, дизайн-система, типобезопасность, Lighthouse). `verified: true` — только при зелёной регрессии с логом.
- **Упал тест — triage сначала** (`.cursor/skills/test-triage/SKILL.md`): функционал по спеке → ROOT=code|test → потом фикс.
- **API-контракты** фиксируют `sa1`/`sa2` до старта — не меняй типы ответов в одностороннем порядке; расхождение с backend решается через спеку, а не хардкодом на клиенте.
- **Docs-sync:** новый UI-flow/hotkey → `docs/wiki/SHORTCUTS.md` + `QUICKSTART.md`; любой код → session note в `.planning/STATE.md` (rule 55).
- **Git:** фича-ветки + Conventional Commits (`feat(frontend): ...`); **мерж в master — только владелец**. Параллельный поток не трогает твои файлы и наоборот.

## Чек-лист готовности

- [ ] Стек поднят, dev-login работает, продукт прощупан руками.
- [ ] `tsc` + Biome + Vitest зелёные локально; `link-testframework-node-modules.mjs` понятен.
- [ ] Прочитаны: frontend.md, design.md, code_style_guide, SHORTCUTS.
- [ ] Понимаю дизайн-систему `--hub-*` и границы Shadcn.
- [ ] Знаю Rules of Hooks и таблицу «нельзя/нужно».
- [ ] Первый PR прошёл кросс-ревью + fe-ver и доведён до merge-ready.
- [ ] Знаю git-политику: фича-ветки — можно, master — только владелец.
