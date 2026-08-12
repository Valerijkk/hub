# FIX-W0 — A12-001 ent.init panic (`trace_context`)

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-26  
**Scope:** только регенерация Ent; схемы/security/version не трогались.

## Что сломалось и почему

В `ent/schema/outboxevent.go` поле `trace_context` (JSON, optional) вставлено третьим по счёту, а локально сгенерированный `ent/runtime.go` остался от схемы без него: `DefaultAttempts` брал `outboxeventFields[4]` (фактически enum `status` → string `"pending"`) и падал cast’ом в `int` в `ent.init` (`runtime.go:1737`). Любой пакет, импортирующий `ent`, не поднимался. CI это маскирует: в `.github/workflows/ci.yml` перед тестами всегда `go generate ./ent`.

## Действие

```text
cd backend && go generate ./ent
```

Индексы после регенерации (проверено в `ent/runtime.go`):

| Default | Было | Стало | Ожидание |
|---------|------|-------|----------|
| `DefaultAttempts` | `[4]` | **`[5]`** | attempts |
| `DefaultCreatedAt` | `[6]` | **`[7]`** | created_at |

`runtime.go` руками не правился.

## go test ДО / ПОСЛЕ

### ДО (`go generate`)

```text
go test -tags hubtest ./testframework/auth/... -count=1
panic: interface conversion: interface {} is string, not int
github.com/valerii/hub/backend/ent.init.0()
  .../backend/ent/runtime.go:1737
FAIL  github.com/valerii/hub/backend/testframework/auth  0.897s
FAIL
```

**Пакеты:** 1/1 прогнанных (auth) — **FAIL** из‑за `ent.init`. По смыслу A12: любой пакет с импортом `ent` падает так же на init; без ent (часть `status` / `platform/*`) могут жить.

### ПОСЛЕ

```text
go generate ./ent                         # exit 0
go build ./ent/...                        # exit 0 (GOPROXY=off, без сети)

go test -tags hubtest ./testframework/auth/... -count=1
ok  github.com/valerii/hub/backend/testframework/auth  2.582s   # повтор 2.710s

go test -tags hubtest ./testframework/status/... -count=1
ok  .../testframework/status  0.561s

go test -tags hubtest ./testframework/platform/... -count=1
ok  crypto, errors, http, middleware, nats, observability, outbox, telegram
```

| Метрика | ДО | ПОСЛЕ |
|---------|----|-------|
| `testframework/auth` | FAIL (`ent.init` panic) | **PASS** |
| `ent.init` блокирует импортёров | да | **нет** |

Полный `go test -tags hubtest ./...` и полный `go build ./...` **не гонялись до конца**: Docker Desktop недоступен (нет PG/Mongo для доменов), а `go build ./cmd/...` / `./...` тянет PDF-стек `imex`/`tms` → `maroto` → `pdfcpu` (не относится к A12-001; в sandbox-кэше модуль битый/неполный). Это **не** регрессия фикса.

## Какие схемы затронула регенерация

`go generate ./ent` переписывает **всё** дерево `backend/ent/*` кроме tracked `schema/` + `generate.go` (стандартное поведение ent v0.14.1). Отдельного «только outboxevent» режима нет.

Подтверждённый логический drift был **только у OutboxEvent** (panic + неверные индексы). Сравнительный snapshot «до» по остальным схемам не сохранялся (generated gitignored) — других index-cast паник на прогнанных пакетах после generate нет.

Политика репо (`.gitignore:51-53`):

```gitignore
backend/ent/*
!backend/ent/schema
!backend/ent/generate.go
```

Схема `outboxevent.go` с `trace_context` уже в git с 2026-07-11; **tracked diff для `fix(ent)` отсутствует** — чинится локальный/CI generate, не исходник схемы. Force-add сгенерированного дерева против политики проекта не делался.

## Оставшиеся падения (не чинились)

| Что | Причина | Связь с A12-001 |
|-----|---------|-----------------|
| `go build ./cmd/...` / `./...` в этой среде | sandbox/modcache: `pdfcpu` / транзитивы (`imex`, `tms` PDF) | нет |
| Полный `go test -tags hubtest ./...` | нет Docker → нет `POSTGRES_TEST_URL` / Mongo для доменов | нет |
| Security / IDOR / FE contract из audit | вне scope W0 | нет |

Security-находки, `version`, soft-delete, правки схем — **не трогались**.

## Коммиты

1. ~~`fix(ent): regenerate...`~~ — **нет tracked файлов** (см. gitignore). Локально generate выполнен; CI сделает то же на job.
2. `docs(audit): W0 fix report` — этот файл.
