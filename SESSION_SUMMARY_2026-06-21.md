# Hub — Сводка работ (2026-06-21)

> Branch: `feat/v2.5-trust` · Обновлено после завершения всех агентов Wave RU-1

---

## 1. Общий прогресс

| Направление | Статус | Прогресс |
|-------------|--------|----------|
| **Wave A** — Coverage фазы 121–125 | ✅ Готово | COV **82.6%**, security ≥80% |
| **Wave B** — Test Framework Migration | 🔄 В работе | **52%** Go-тестов в testframework |
| **Wave RU-1** — Русификация comments + docs | ✅ Готово | 30 пакетов / 9 агентов |
| **Wave RU-V** — Верификация 100% RU | ⏳ Не начато | 10 верификаторов |
| **Wave C** — Финальный аудит FA-01…12 | ⏳ Заблокировано | После Wave B |

---

## 2. Wave A — Coverage (фазы 121–125) ✅

| Phase | Результат |
|-------|-----------|
| 121 | notification **82.5%**, tracker **82.4%** |
| 122–124 | messenger **82.3%**, storage **86.8%**; tms **81.8%**, project **81.7%** (−0.2% до 82%) |
| 125 | COV **82.6%**; auth **82.4%**, ws **87.7%**, rbac **84.7%**, audit **80.5%** |

---

## 3. Wave B — Test Framework

### Архитектура

```
backend/internal/<mod>/          production + export_hooks.go (hubtest)
backend/testframework/<mod>/     package <mod>_test
testframework/e2e/               Playwright (корень монорепо)
frontend/testframework/unit/       Vitest (25 файлов)
deployments/docker-compose.test.yml  PG + Mongo + Redis + NATS
```

**Docker для Go-тестов:** контейнер `golang:1.25-bookworm` + mount `backend:/app` + сеть compose. Отдельных Dockerfile на модуль нет.

### Счётчики тестов

| Слой | Расположение | Файлов |
|------|--------------|--------|
| Go testframework | `backend/testframework/` | **251** |
| Go internal (остаток) | `backend/internal/*/_test.go` | **227** |
| Vitest unit | `frontend/testframework/unit/` | **25** |
| Vitest в src | `frontend/src/**/*.test.*` | **12** |
| Playwright E2E | `testframework/e2e/specs/` | **20** |

**Миграция Go:** 251 / (251+227) = **52.5%**

### Модули в testframework

| Модуль | Файлов | Docker `-tags hubtest` |
|--------|--------|-------------------------|
| notification | 36 | ✅ PASS |
| auth | 26 | ⚠️ compile (jwtSecret unexported) |
| ws | 8 | ✅ PASS |
| rbac | 6 | ✅ PASS |
| audit | 18 | ✅ PASS |
| tracker | 46 | ❌ compile (export_hooks) |
| tms | 45 | ❌ compile |
| project | 26 | ❌ compile |
| messenger | 39 | ❌ compile |

### Осталось в internal (~227 файлов)

ai(30), search(25), calls(21), github(16), apigateway(10), gitlab(9), storage(9), backup(8), imex(8), bases(8), crm(8), intranet(7), sprint(7), dashboard(6), hr(6), marketplace(6), accounting(6), mail(5), dev(3), calendar(2), status(2)

### Инструменты

- `scripts/migrate-go-tests.ps1` — перенос internal → testframework
- `scripts/fix-testframework-prefix.ps1` — префиксы `mod.Symbol`
- `export_hooks.go` — auth, ws, rbac, audit, notification, tracker, messenger

### Следующие шаги Wave B

1. TF-M04 — починить compile tracker/tms/project/messenger + auth regression
2. TF-M05 — мигрировать 227 файлов
3. TF-M06 — 12 Vitest из `frontend/src` → testframework
4. TF-M08 — `internal/*/_test.go` = 0
5. Wave C FA-01…12

---

## 4. Wave RU-1 — Русификация ✅ (9 агентов)

### Docs (~38 markdown)

| Агент | Файлы | Ключевое |
|-------|-------|----------|
| RU-D01–D03 | 5 core docs | v4 shipped, 82 Ent schemas, 31 модуль, Magic Link fix |
| RU-D04–D08 | 14 docs + wiki | CI/Biome gate, E2E path, rate 60/min |
| RU-D09–D10 | 9 planning docs | STATE, COVERAGE, AGENTS, REGULATION, README |

### Backend production (~255 `.go`)

| Агент | Область | Файлов |
|-------|---------|--------|
| RU-B01–B03 | auth, ws, rbac, audit, notification | 45 |
| RU-B04–B07 | tracker, tms, project, sprint, messenger, calls | ~70 |
| RU-B08–B14 | platform, ent, cmd, ai, crm, github, … | ~140 |

`go build -tags release ./...` — OK, логика не менялась.

### Frontend (~113 файлов)

| Агент | Область |
|-------|---------|
| RU-F01–F03 | api, hooks, stores, modules, components, routes |

~440 комментариев на русский; `tsc --noEmit` OK.

### Testframework + E2E (289 файлов)

| Агент | Область |
|-------|---------|
| RU-B15 | 255 Go + 34 e2e TS — intent-комментарии на русском |

### Infra (23 файла)

| Агент | Область |
|-------|---------|
| RU-F05 | deployments/*, cli/*, новые README |

---

## 5. Известные gaps (для RU-V и Wave C)

| Gap | Описание |
|-----|----------|
| RU-V не запущена | 10 верификаторов — grep англ. комментариев, doc-verifier |
| Messenger test literals | Битая кодировка в assert-строках (логику не трогали) |
| frontend/README.md | Отсутствует |
| Swagger | ~286 REST vs ~100 paths в swagger.yaml |
| TF-M04 compile | 4 core-модуля + auth regression |
| tms/project coverage | 81.7–81.8% (−0.2% до порога 82%, FA-11) |
| Webcam | Out of Scope в BUSINESS vs Phase 16 в roadmap |

---

## 6. Docker — каноническая команда

```powershell
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default `
  -v "${PWD}/backend:/app" -v hub-go-mod-cache:/go/pkg/mod -w /app `
  -e CGO_ENABLED=1 `
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" `
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" `
  golang:1.25-bookworm `
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."
```

---

## 7. Приоритеты дальше

1. **TF-M04 + TF-M05** — завершить testframework migration
2. **RU-V01…V10** — верификация 100% русского
3. **Wave C FA-01…12** — аудит, best practices, qa-gate

---

*Детали: `.planning/TASKS-PHASES-121-FINAL-AUDIT.md`, `.planning/TASKS-RU-COMMENTS-DOCS.md`*
