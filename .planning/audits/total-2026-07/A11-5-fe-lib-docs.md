# W11 агент A5 — fe lib+docs part 5 (remaining docs/wiki)

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: remaining docs/wiki drift vs code («Чему верить нельзя» candidates).  
> ID: `W11-A5-001`…  
> Сверка: метрики репозитория (testframework/swagger/routes), CI `ci.yml`, TZ phases, auth/calls/wiring.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `docs/security_guidelines.md` | 351 | 2 | замечания |
| `docs/stack_and_tools.md` | 236 | 1 | замечания |
| `docs/system_specification.md` | 692 | 2 | замечания |
| `docs/TECHNICAL.md` | 1899 | 2 | замечания |
| `docs/testframework_strategy.md` | 353 | 1 | замечания |
| `docs/TESTING.md` | 552 | 3 | замечания |
| `docs/test_plan.md` | 215 | 1 | замечания |
| `docs/TROUBLESHOOTING.md` | 246 | 2 | замечания |
| `docs/tz_constraints.md` | 153 | 1 | замечания |
| `docs/user_needs_analysis_2026.md` | 263 | 3 | замечания |
| `docs/vision_and_scope.md` | 636 | 3 | замечания |
| `docs/wiki/API.md` | 1164 | 2 | замечания |
| `docs/wiki/ARCHITECTURE.md` | 192 | 1 | замечания |
| `docs/wiki/DEPLOY.md` | 258 | 0 | чисто |
| `docs/wiki/HOME.md` | 87 | 1 | замечания |
| `docs/wiki/QUICKSTART.md` | 119 | 2 | замечания |
| `docs/wiki/SHORTCUTS.md` | 93 | 0 | чисто |

Файлов в списке: 17. Прочитано: 17. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W11-A5-001 | P1 | `TESTING.md:95-102` | Команды регрессии: `go test … ./internal/...` и `./internal/auth/...` | В `backend/internal/` **0** `*_test.go` (TF-M08). Путь вводит в заблуждение; CI гоняет только `./testframework/...` (`.github/workflows/ci.yml:61`) | Заменить на `./testframework/...` / `./testframework/auth/...`; удалить `./internal/...` |
| W11-A5-002 | P1 | `TESTING.md:43-79` | Инвентарь «33 тестовых файла / handler_test.go / Telegram auth» | Пост-миграционный артефакт: тесты в `testframework/`, Magic Link only; «Telegram auth» противоречит ADR-AUTH-01 и § выше в том же файле | Переписать инвентарь под `backend/testframework/**` (факт **580** файлов) или удалить устаревшую таблицу |
| W11-A5-003 | P1 | `TECHNICAL.md:12` | CI job `backend-test`: `go test … ./internal/... ./testframework/...` | Не совпадает с реальным CI; `./internal/...` пуст по тестам | Синхронизировать с `ci.yml`: только `./testframework/...` |
| W11-A5-004 | P1 | `wiki/QUICKSTART.md:95` | Минимальная регрессия включает `./internal/... ./testframework/...` | Та же ловушка, что W11-A5-001; «зелёный» локальный прогон без смысла на internal | Как в `docs/test_plan.md:69` — только `./testframework/...` |
| W11-A5-005 | P1 | `user_needs_analysis_2026.md:23` | Calls: «webcams» в статусе **Production** | TZ + `vision` US-M06: webcam → **Phase 16**; в `frontend/src/modules/calls` нет webcam UI | Убрать webcams из Production **или** пометить Phase 16 / Partial |
| W11-A5-006 | P1 | `vision_and_scope.md:28` vs `:428-431` | G-07: «US-WIKI реализованы и приняты» ↔ ниже KPI wiki — **«планируется»** | Внутренний drift: цель отмечена done, US-WIKI-01..05 ещё planned; wiki-модуля в routes нет как shipped G-07 | G-07 → Partial/planned; не утверждать «реализованы» до кода |
| W11-A5-007 | P2 | `test_plan.md:52-56`; `testframework_strategy.md:76-78` | BE tests **515**, Vitest **37 files / 99 tests** | Факт 2026-07-28: **580** `*_test.go`; Vitest **65** файлов / ~**169** `it\|test` | Пересчитать и обновить оба файла (+ TROUBLESHOOTING) |
| W11-A5-008 | P2 | `TROUBLESHOOTING.md:21`; `:48` | «Все **39** export_hooks»; «**515** тестовых файлов» | Факт: **36** `export_hooks.go`; **580** `*_test.go` | Обновить числа; не ссылаться на 515/39 |
| W11-A5-009 | P2 | `TESTING.md:33` | Backend test files = **499** | Устарело даже относительно соседних docs (515) и факта (580) | Одна каноническая цифра → `test_plan` / COVERAGE |
| W11-A5-010 | P2 | `security_guidelines.md:44,54,71-72` | Фиксы привязаны к `cmd/server/main.go:265/544/117…` | `main.go` = **34** строки; wiring в `wiring_auth.go` / `wiring_*.go` | Обновить file:line на актуальный wiring |
| W11-A5-011 | P2 | `security_guidelines.md:80,227,315` | ARCH-H1 всё ещё **PARTIAL** / «прямой DB в handler» | В `auth/handler.go` комментарий ARCH-H1 + вызов `h.service.GetMe` — layering закрыт | Статус → FIXED; убрать Phase 19 backlog |
| W11-A5-012 | P2 | `system_specification.md:159`; `wiki/API.md:11,720` | Infra/audit «в `main.go`»; inventory `/docs`,`/health`,`/metrics` | Health: `wiring_messenger.go` (`/live`,`/ready`,`/health`); metrics: `wiring_fiber.go:58`; audit MW — `wiring_fiber.go` | Ссылки на wiring_*; inventory: +`/live`/`/ready` |
| W11-A5-013 | P2 | `system_specification.md:27,160`; `wiki/API.md:5-12`; `wiki/HOME.md:44` | REST «**~286**» / «286 endpoints» | `backend/docs/swagger.json` paths ≈ **252**; расхождение ~34 без пояснения методики | Зафиксировать метод подсчёта (RegisterRoutes vs swagger) и переснять |
| W11-A5-014 | P2 | `vision_and_scope.md:35`; `wiki/ARCHITECTURE.md:83` | UI-маршруты «**60+**» | Факт: **77** `.tsx` в `frontend/src/routes/` | Обновить KPI / wiki |
| W11-A5-015 | P2 | `vision_and_scope.md:40` | Security 6/6: «auth, ws, **storage**, rbac, audit, calls» | Канон COVERAGE: 6-й пакет = **`platform/middleware`**, не storage (storage ≥80% отдельно) | Заменить storage → platform/middleware |
| W11-A5-016 | P2 | `user_needs_analysis_2026.md:4,57` | Product score **93/100** | `AGENTS.md` / актуальный статус: **89/100** | Синхронизировать со scorecard / AGENTS |
| W11-A5-017 | P2 | `TECHNICAL.md:61` | `auth/` — «Magic Link, **Telegram**, JWT» | Telegram не login; читается как auth-метод | «Magic Link + JWT; Telegram — notify bind (notification)» |
| W11-A5-018 | P2 | `tz_constraints.md:120-124` | Gantt / AI BYOK+Ollama → Phase 16/17 | Gantt route+`GanttChart` уже в FE; AI Hub v3.0 shipped (user_needs/AGENTS) | Обновить таблицу фаз под факт shipped / уточнить «остаток scope» |
| W11-A5-019 | P2 | `wiki/QUICKSTART.md:51-57` | Онбординг «**3 шага**» (профиль / обзор / настройки) | FE: **4** шага — profile, tour, theme, project (`onboarding.lazy.tsx` STEP_KEYS) | Описать 4 шага как в коде |
| W11-A5-020 | P3 | `stack_and_tools.md:60,67,73` | Zustand «11+»; motion `^12.38.0`; PWA `^1.2.0` | Stores: **12** файлов; package.json: motion `^12.40.0`, vite-plugin-pwa `^1.3.0` | Подтянуть pin’ы и «12 stores» |
| W11-A5-021 | P3 | `wiki/API.md:68-75` | Примеры curl на `localhost:3000` | Локальный UX-путь — Caddy `:8080` (QUICKSTART/DEPLOY); 3000 = прямой backend | Уточнить «backend :3000 / через Caddy :8080» |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `docs/wiki/DEPLOY.md` | Образы PG17/Mongo8/Redis7.4/NATS2.11/SeaweedFS3.97/LiveKit v1.8/Caddy2 совпадают с `docker-compose.prod.yml`; `/api/live`+`/api/ready`; `.env.example`; local compose + DEV_MODE — ок |
| `docs/wiki/SHORTCUTS.md` | `Ctrl+K` → `command-palette.tsx` + e2e path существует; `Ctrl+N` / `?` в `app-shell.tsx`; raise-hand `POST /api/calls/:id/raise-hand` в handler — ок |
| `docs/tz_constraints.md` (кроме фаз) | Magic Link, JWT 15m/30d, rate limits, UUID/version, Fiber/NATS/SeaweedFS/Caddy, forbid-паттерны — согласованы с правилом 60 и кодом |
| `docs/stack_and_tools.md` (ядро) | Go 1.25.0, Fiber v3.0.0-beta.3, Ent 0.14.1, Node 22 CI, основные go.mod pins — ок |
| Auth contract (API.md / ARCHITECTURE magic TTL) | Magic Link only; Redis TTL кода **5 min** (`magiclink.go`) — ок |
| Ent schemas | **82** entity + `helpers.go` — формулировка «82 schemas» в docs ок |

## Counts

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 6 |
| P2 | 13 |
| P3 | 2 |
| **Σ** | **21** |

Файлов с находками: **15** / 17. Чисто: **2** (`DEPLOY.md`, `SHORTCUTS.md`).

## Топ-3 («Чему верить нельзя»)

1. **Команды тестов с `./internal/...`** — `TESTING.md`, `TECHNICAL.md`, `wiki/QUICKSTART.md` (W11-A5-001…004): копипаст сломает/обессмыслит локальную регрессию после TF-M08.  
2. **Матрица зрелости фич** — `user_needs` (webcams Production), `vision` G-07 US-WIKI done↔planned, `tz_constraints` фазы AI/Gantt vs shipped код (W11-A5-005,006,018).  
3. **Метрики тестов/API** — 499/515 vs **580** BE tests; Vitest 37/99 vs **65/~169**; REST **286** vs swagger **~252** (W11-A5-007…009,013).

## Evidence (команды сверки)

```
# counts 2026-07-28
(Get-ChildItem -Recurse backend/testframework -Filter *_test.go).Count          # 580
(Get-ChildItem -Recurse testframework/stages/unit/frontend-test -Filter *.test.ts*).Count  # 65
(Get-ChildItem -Recurse backend/internal -Filter export_hooks.go).Count         # 36
(Get-ChildItem -Recurse frontend/src/routes -Filter *.tsx).Count                # 77
(Get-ChildItem backend/ent/schema/*.go).Count                                   # 83 (=82+helpers)
(Get-Content backend/cmd/server/main.go).Count                                  # 34
# swagger paths
(ConvertFrom-Json (Get-Content backend/docs/swagger.json -Raw)).paths.PSObject.Properties.Name.Count  # 252
# CI
rg "go test -tags hubtest" .github/workflows/ci.yml   # ./testframework/... only
```
