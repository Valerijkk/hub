# Troubleshooting — типовые проблемы разработки

> Грабли, на которые наступают при работе с Hub локально: сборка, тесты, окружение, Docker. Для проблем **продакшна** — [`runbooks/`](runbooks/README.md).
>
> Каждый пункт: симптом → причина → решение. Все команды и факты проверены на текущем состоянии репозитория.

*Создано: 2026-07-19.*

---

## 1. Тесты backend

### 1.1 ⚠️ `undefined: <pkg>.Export...` при прогоне тестов

**Симптом:**

```
vet.exe: testframework\accounting\fixtures_test.go:58:20: undefined: accounting.ExportMapServiceError
```

**Причина.** Тесты в `backend/testframework/` — белый ящик: они обращаются к внутренностям модулей через хуки `backend/internal/<module>/export_hooks.go`. Все **39** таких файлов (`export_hooks.go`, measured 2026-07-29) помечены `//go:build hubtest`, то есть **компилируются только с этим build-тегом**. Без тега символы не существуют.

**Решение — всегда прогонять backend-тесты с `-tags hubtest`:**

```bash
cd backend
go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...
```

Точечно по домену:

```bash
go test -tags hubtest -race -count=1 ./testframework/tracker/...
```

### 1.2 ⚠️ `make test` не прогоняет backend-тесты (известный дефект)

**Симптом.** `make test` завершается ошибкой компиляции backend-тестов либо не выполняет ни одного backend-теста.

**Причина.** Цель в `Makefile`:

```make
test:
	cd backend && go test -race -count=1 ./...
	cd frontend && npm test -- --run
```

Здесь **нет `-tags hubtest`**, поэтому пакеты `testframework/*` не компилируются (см. §1.1). Дополнительно: в `backend/internal/` **ноль** файлов `*_test.go` — все **804** тестовых файлов лежат в `backend/testframework/` (measured 2026-07-29; was 515 → 580 → 652). То есть даже без ошибки компиляции `./...` не нашёл бы там тестов.

**Решение сейчас** — не полагаться на `make test` для backend, использовать канонические команды:

```bash
# Backend (канон)
cd backend && go build -tags release ./... && go vet ./... \
  && go test -tags hubtest -race -count=1 -timeout=600s ./testframework/...

# Frontend
cd frontend && npx tsc --noEmit && npm run lint && npx vitest run
```

Либо слэш-скилл `/qa-gate` (`.cursor/skills/qa-gate/SKILL.md`) — он останавливается на первой ошибке.

> `TODO`: починить цель `test` в `Makefile` (добавить `-tags hubtest` и путь `./testframework/...`). Пока не починено — цель вводит в заблуждение: «зелёный make test» не означает, что backend-тесты прошли.

### 1.3 Интеграционные тесты падают на подключении к БД

**Симптом.** `connection refused` к Postgres/Mongo при прогоне тестов.

**Причина.** Интеграционные тесты подключаются к контейнерам из `deployments/docker-compose.test.yml` через переменные окружения. Testcontainers в проекте **не используются** — окружение поднимается отдельно.

**Решение:**

```bash
docker compose -f deployments/docker-compose.test.yml up -d

cd backend
POSTGRES_TEST_URL="postgresql://hub:hub@localhost:5432/hub?sslmode=disable" \
MONGODB_TEST_URL="mongodb://hub:hub@localhost:27017/?authSource=admin" \
go test -tags hubtest -race -count=1 ./testframework/...
```

Канонический прогон (как в CI) — в контейнере, чтобы имена хостов совпадали с сетью compose:

```bash
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."
```

> Разница в хостах: изнутри контейнера — `postgres`/`mongo` (имена сервисов), с хоста — `localhost` (опубликованные порты).

### 1.4 Тест проходит локально, падает в CI (или наоборот)

Проверять по порядку:

| Гипотеза | Как проверить |
|----------|---------------|
| Разные флаги | CI гоняет с `-race -count=1`; локально без них тест может «повезти» |
| Зависимость от порядка | Прогнать сьют целиком, а не один тест: `go test -tags hubtest ./testframework/<domain>/...` |
| Общие данные | Прогнать дважды подряд без очистки БД — упадёт, если тест не изолирован |
| Реальное время/таймзона | Тест зависит от `time.Now()` или локали машины |
| Кэш результатов | `-count=1` отключает кэш; без него Go может вернуть старый результат |

Протокол разбора — `.cursor/skills/test-triage/SKILL.md`: сначала сверить функционал со спецификацией, потом решать, чинить код или тест.

---

## 2. Тесты frontend

### 2.1 Vitest/Playwright не находят модули из `testframework/`

**Симптом.** Ошибки резолва импортов при прогоне тестов фронтенда.

**Причина.** Тесты фронтенда живут в `testframework/stages/unit/frontend-test/`, вне `frontend/`, и им нужны слинкованные `node_modules`.

**Решение — обязательный шаг перед прогоном:**

```bash
node frontend/scripts/link-testframework-node-modules.mjs
cd frontend && npx vitest run
```

Пропуск этого шага — самая частая причина «у меня не запускаются фронтовые тесты».

### 2.2 Playwright: браузеры не установлены

```bash
cd frontend && npx playwright install --with-deps
```

Smoke-набор (быстрый гейт) против полного:

```bash
npm run test:e2e:smoke   # 3 спеки
npm run test:e2e         # все домены
```

---

## 3. Локальное окружение

### 3.1 Запуск стека

```bash
make dev-up          # docker compose -f deployments/docker-compose.yml up -d
make dev-backend     # cd backend && air   (hot reload)
make dev-frontend    # cd frontend && npm run dev
```

`make dev` поднимает только инфраструктуру и печатает подсказку — backend и frontend запускаются отдельными терминалами.

### 3.2 Порт занят

**Симптом:** `bind: address already in use`.

```bash
docker compose -f deployments/docker-compose.yml ps      # что уже поднято
docker compose -f deployments/docker-compose.yml down    # остановить прошлый стек
```

Частая причина — оставшийся контейнер от предыдущей сессии или локально запущенный Postgres, конфликтующий с контейнерным.

### 3.3 `.env.local` отсутствует

```bash
cp .env.local.example .env.local
```

Далее заполнить значения. Секреты **не коммитить** — проверка есть в `.cursor/rules/45-artifacts-cleanup.mdc` и `50-git.mdc`.

### 3.4 Ent: изменил схему, код не видит изменений

Схема Ent требует кодогенерации:

```bash
cd backend && go generate ./ent
```

Без этого шага новые поля/сущности не появятся в сгенерированном API.

### 3.5 Swagger-документация не обновилась

```bash
make swagger    # swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

---

## 4. Гигиена репозитория

### 4.1 После прогона покрытия остались артефакты

Профили покрытия и дампы **не коммитятся** (`.cursor/rules/45-artifacts-cleanup.mdc`). Писать их следует во временный каталог:

```bash
go test -tags hubtest -coverprofile="$TMPDIR/coverage.out" ./testframework/...
```

Если артефакты всё же появились в репозитории — удалить до коммита:

```powershell
Get-ChildItem backend -File | Where-Object {
  $_.Name -match '^(cover|coverage|cov_|tmp|.*_cov)' -or $_.Extension -eq '.out'
} | Remove-Item -Force
```

### 4.2 Проверка перед PR

```bash
# Полный гейт
/qa-gate                       # .cursor/skills/qa-gate/SKILL.md

# Стиль и anti-AI паттерны
/style-check                   # .cursor/skills/style-check/SKILL.md
```

Ручной эквивалент — команды из §1.2.

---

## 5. Когда проблема оказалась багом

Если разбор показал дефект кода, а не окружения:

1. **Test triage** (`.cursor/skills/test-triage/SKILL.md`) — зафиксировать `ROOT=code` или `ROOT=test` с обоснованием от спецификации, **до** правки.
2. Регрессионный тест писать на **класс** проблемы, а не на конкретный воспроизведённый случай.
3. Прогнать три круга: изолированно → сьют домена → смежные домены.
4. Docs-sync по матрице `.cursor/rules/55-docs-sync.mdc`.

Для инцидента в продакшне — [`runbooks/incident-response.md`](runbooks/incident-response.md), включая шаблон постмортема (§7).

---

## 6. Куда идти дальше

| Нужно | Документ |
|-------|----------|
| Термин/префикс ID непонятен | [GLOSSARY.md](GLOSSARY.md) |
| Онбординг в роль | [onboarding/](onboarding/README.md) |
| Стратегия тестирования | [test_plan.md](test_plan.md), [testframework_strategy.md](testframework_strategy.md) |
| Карта тестов по стадиям CI | `testframework/PIPELINE.md` |
| Прод лежит | [runbooks/incident-response.md](runbooks/incident-response.md) |
| Правила стиля и запреты | [code_style_guide.md](code_style_guide.md), `.cursor/rules/10-style.mdc` |
