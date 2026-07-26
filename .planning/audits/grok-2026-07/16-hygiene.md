# Repo Hygiene — Dead Code / Deps / Docs Drift / Naming

> READ-ONLY аудит зоны A16 · 2026-07-26 · ветка `cursor-grok-4.5`  
> Объём: мёртвый код / unused exports · дубликаты между доменами · TODO/FIXME/HACK · go.mod + package.json · нейминг/структура · расхождения docs/README с фактом.  
> Product code не менялся; единственный артефакт — этот отчёт.

## Что проверено

| Область | Метод |
|---------|--------|
| Dead / unused FE | grep импортов `VelocityChart` / `modules/tracker/velocity`; API-клиенты `@/api/*`; `jsdom` vs `happy-dom` в `vite.config.ts` |
| Дубликаты логики | `IsSystemAdmin` ×N репозиториев; maroto PDF `imex/pdf.go` ↔ `tms/exporter.go`; burndown/velocity в `modules/tracker` + `components/dashboard/widgets` |
| TODO/FIXME/HACK | `backend/internal` (0), `frontend/src` (1), `docs/runbooks`, `.planning/edge-expansions` (каталог ECC) |
| Deps | чтение `backend/go.mod`, `frontend/package.json` + lock; `npm outdated` (локально); go list -m -u — **не завершён** (скачивание toolchain go1.25) |
| Naming / структура | `backend/internal/*` (33) ↔ `frontend/src/modules` / `api` / `stores`; layering gaps (`ws`, `mail`, `status`, `apigateway`) |
| Docs drift | `README.md`, `AGENTS.md`, `docs/{vision,architecture,stack_and_tools}`, `.cursor/rules/{00-project,30-frontend}.mdc` vs filesystem counts |

**Факт-сверка (workspace):** Ent entity schemas **82** (+ `helpers.go` → 83 файлов); internal pkgs **33**; Zustand stores **13**; route `.tsx` **72**; `createFileRoute`/`createLazyFileRoute` hits **142**; backend `*_test.go` в `testframework/` **499**; Vitest claim **99/99** (не перегонялся).

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг / долг | Как чинить |
|----|----------|-------------|---------|----------------------|------------|
| A16-001 | P2 | `frontend/src/modules/tracker/velocity-chart.tsx:25` | Unused export `VelocityChart` — ни одного импорта вне файла | Sprint UI использует только `BurndownChart` (`sprints/$sprintId.tsx`); dashboard — отдельный `VelocityChartWidget`. Мёртвый код / дрейф SCRM-02 UI | Подключить на sprint page **или** удалить файл; один канон графика |
| A16-002 | P2 | `frontend/package.json:75` (+ `vite.config.ts:78`) | `jsdom` в devDependencies, runtime тестов — `happy-dom` | Лишняя тяжёлая зависимость; путаница окружения; `npm outdated` тянет оба | Удалить `jsdom` из `package.json` / lock, оставить happy-dom |
| A16-003 | P2 | `frontend/package.json:44,58,59,65,68` | Пять пакетов на `"latest"`: `clsx`, `sonner`, `tailwind-merge`, `@biomejs/biome`, `@tanstack/router-plugin` | Невоспроизводимые сборки; CI drift; lock ≠ манифест-интент | Запинить semver (как у остальных deps); обновить lock |
| A16-004 | P1 | `backend/go.mod:14` | Prod HTTP-стек на `github.com/gofiber/fiber/v3 v3.0.0-beta.3` | Pre-stable API/security surface в единственном gateway; известные Fiber v3 middleware gotchas | Трекать stable Fiber v3; pin + changelog review; не игнорировать beta в security-gate |
| A16-005 | P2 | `backend/internal/auth/service_organization.go:29` (+ копии: `accounting/repository.go:217`, `audit/repository.go:512`, `hr/repository.go:147`, `intranet/repository.go:133`, `marketplace/repository.go:170`, `search/repository.go:56`) | `IsSystemAdmin` / founder\|\|admin скопирован в ≥6 пакетах | Расхождение семантики при правке одной копии (уже ловили founder-bypass в RBAC/dashboard аудитах) | Один shared helper (`auth.IsSystemAdmin` / platform) + тонкие repo-обёртки |
| A16-006 | P2 | `backend/internal/imex/pdf.go:22` ↔ `backend/internal/tms/exporter.go:178` | Два независимых maroto PDF-экспорта «отчёт по прогону» | Дубли стилей/колонок; правки отчёта в одном месте не попадают в другое | Общий `imex`/`report` builder; TMS/imex вызывают один API |
| A16-007 | P2 | `README.md:24` (+ `:154`, `:233`) | README: **39+** routes, **495** backend-тестов | Факт: **72** route `.tsx`, vision **60+** (`docs/vision_and_scope.md:36`), тесты **499** (`AGENTS.md:149`) | Синхронизировать badge/таблицы с AGENTS/vision; один source of truth |
| A16-008 | P2 | `.cursor/rules/00-project.mdc:15,34,44,45` | Rules: Go **1.24**, **39** routes, **11** stores | `go.mod:3` → **1.25.0**; stores **13**; routes ≫39; `docs/stack_and_tools.md:12` уже 1.25 | Обновить 00-project (+ зеркало в AGENTS layout) |
| A16-009 | P2 | `.cursor/rules/30-frontend.mdc:124` vs `:135` | «11 stores», но перечень в том же файле — **13** имён | Внутренний drift правила → агенты считают неверный канон | Исправить на 13; сверить с `frontend/src/stores/*.ts` |
| A16-010 | P2 | `docs/architecture.md:197` (+ `:204` TD-01, `:206` TD-03) | Top-5/TD ссылаются на `main.go:540`, `rbac/... *ent.Client`, «main 630 LOC» | `main.go` = **34** строки (wiring_*); `rbac/middleware.go` **без** `*ent.Client` (Repository); `GetDB` не на `:153` | Переписать Top-5/TD под текущий код; закрыть FIXED пункты явно |
| A16-011 | P3 | `frontend/src/routes/_authenticated/messenger/$channelId.tsx:9` | `TODO(FE-01)` после закрытого hooks-бага | Комментарий сам говорит: триггер закрыт; TODO = шум для hygiene-сканов | Переименовать в NOTE / убрать после стабилизации WS |
| A16-012 | P2 | `docs/runbooks/incident-response.md:38` (+ `:165`, `:177`, `:183`) | Активные `TODO(ops)`: хост, log-driver, домен, лимиты Resend | Runbook нельзя выполнить «как написано» в инциденте | Заполнить ops-факты или ссылки на secrets/inventory |
| A16-013 | P3 | `frontend/src/api/notifications.ts:1` / `files.ts` / `projects.ts` vs `backend/internal/{notification,storage,project}` | FE API plural / backend package singular | Поиск по домену ломается; onboarding агентов путает storage↔files | Зафиксировать конвенцию в code_style; постепенно выровнять имена |
| A16-014 | P3 | `frontend/src/components/tracker/` (13 files) + `frontend/src/modules/tracker/` (17 files) | Tracker UI размазан по `components/` и `modules/` | Дубли паттернов (charts), неясный канон feature-module | Свести tracker UI под `modules/tracker` (как CRM/HR) |
| A16-015 | P3 | `frontend/package.json:41` | `@types/react-grid-layout` в `dependencies` | Types не нужны в prod bundle metadata; засоряет runtime deps | Перенести в `devDependencies` |
| A16-016 | P3 | `README.md:74-80` | Скриншоты → `docs/assets/screenshots/*.png`, каталога **нет** | Битые картинки на главной репо | Добавить ассеты или убрать секцию до появления |
| A16-017 | P2 | `frontend/package.json` (npm outdated: 30+ pkgs, в т.ч. livekit, radix, tanstack, biome, vite-ecosystem) | Массовый outdated относительно registry | Накопление security/compat lag; нет регулярного bump-цикла | Dependabot/календарь bump; приоритет LiveKit/vite/biome |
| A16-018 | P3 | `.planning/edge-expansions/**` (сотни строк `\| TODO \|`) | ECC-каталог помечен TODO как gap-статус тестов | Не code-debt, но засоряет grep TODO по репо | Отдельный статус-маркер (`GAP`/`OPEN`) ≠ `TODO` в code sense |
| A16-019 | P3 | `backend/internal/{ws,mail,status,apigateway}` | Отклонение от канона handler/service/repository | `ws`: H без S/R; `mail`: S без H/R; `status`: без repository; `apigateway`: без service.go | Документировать исключения **или** выровнять имена файлов |
| A16-020 | P3 | `AGENTS.md:13` vs `docs/vision_and_scope.md:38` | «82 Ent» ок, но «31 пакет» (vision) vs **33** dir в `internal/` | Мелкий drift модульного счёта (platform/testutil/integrations) | Унифицировать формулу подсчёта в KPI-таблицах |

## Чисто (проверено, проблем нет / низкий шум)

- **TODO в `backend/internal`:** 0 (`FIXME`/`HACK`/`XXX` тоже 0).
- **Тесты не в `internal/`:** 0 `*_test.go` — миграция в `testframework/` соблюдена (499 файлов).
- **Ent KPI «82 без helpers»:** совпадает с vision (`helpers.go` отдельно).
- **`docs/stack_and_tools.md`:** Go **1.25.0**, Fiber **beta.3** — честно отражены.
- **go-sqlite3 / nats-server / miniredis:** используются testframework’ом (не orphan в go.mod).
- **FE API-клиенты:** все `@/api/<name>` имеют ≥1 потребителя (нет полностью orphan api-файлов).
- **Дубликат burndown:** `modules/tracker` (sprint page) и dashboard widget — разные контракты (OK, не мёртвый код).

## Оценка зоны: 7/10

**Обоснование:** Product-код гигиеничен по TODO и тестовой раскладке; явного «кладбища» unused API почти нет. Оценка снижена **системным docs/rules drift** (Go 1.24, 39 routes, 11 vs 13 stores, README 495, architecture Top-5 на мёртвые `main.go:540` / `*ent.Client` в rbac), **deps** (`latest` pins, unused jsdom, Fiber **beta** как P1), и **дубликатами** (`IsSystemAdmin`, PDF run-report, orphan `VelocityChart`). После sync docs+rules, pin deps, удаления/подключения velocity и консолидации IsSystemAdmin/PDF — зона ~8.5/10.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 1 |
| P2 | 11 |
| P3 | 8 |
| **Total** | **20** |

**Score: 7/10**
