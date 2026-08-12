# Tests / Coverage / Assert-free / Mocks

> READ-ONLY аудит зоны A12 · 2026-07-26 · ветка `cursor-grok-4.5`  
> Разрешена запись только этого отчёта; product code не менялся.  
> Прогоны: `go test -tags hubtest` (репрезентативные пакеты), `npx vitest run`, инвентарь `testframework/` + сверка `.planning/COVERAGE.md`.

## Что проверено

| Слой | Факт (2026-07-26) |
|------|-------------------|
| Backend `testframework/**/*_test.go` | **499** файлов · **~2741** `func Test*` |
| Backend `internal/*_test.go` | **0** (TF-M08 держится) |
| TF-домены | 36 dirs (`auth`…`ws` + `env`/`factory`/`mongotest`/`steps`) |
| Vitest | **37** файлов · **99** tests · **99/99 PASS** (~6.4s) |
| Playwright E2E | **20** spec · **~53** `test(` в доменах |
| `go test -tags hubtest` | После `go generate ./ent` — auth/platform/status **ok**; без generate на stale tree возможен локальный `ent.init` panic (A12-001 P2 DX) |
| Postgres/Mongo localhost | **недоступны** (5432/27017) — integration Skip-пути не прогонялись live |
| Assert-free Go | **21** `Test*` без `assert`/`require`/`t.Fatal`/`t.Error` |
| Tracker mock vs ent | ~9 mock-файлов / ~65 Test* vs ~11 enttest-файлов / ~43 Test* |

Структура `testframework/PIPELINE.md` ↔ диск: backend физически в `backend/testframework/`, FE unit в `stages/unit/frontend-test/domains/`, e2e в `stages/e2e/domains/` — **совпадает**.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это баг | Как чинить |
|----|----------|-------------|----------|----------------|------------|
| A12-001 | P2 | `backend/ent/runtime.go` (gitignored) + `schema/outboxevent.go:28-34` | Локальный stale `ent` generate после добавления `trace_context` → `ent.init` cast panic | Симптом реален на протухшем дереве, но **не** в git и **не** на CI: `.gitignore:51` игнорирует `backend/ent/*`; CI/`Dockerfile` всегда `go generate ./ent` | DX: `make test` → `generate`; не путать с prod P0 |
| A12-002 | P1 | `backend/testframework/notification/coverage_wave79_test.go:45-48` | Assert-free тест с именем `…UnknownSubjectAcks` | Зовёт `ExportProcessEvent` со `stubJetStreamMsg`, **не** проверяет `msg.acked` (хотя stub умеет флаг — `stubs_test.go:23`, образец assert в `consumer_wave121_test.go:68`) | `assert.True(t, msg.acked)` после вызова |
| A12-003 | P1 | `backend/testframework/github/service_mock_test.go:310-314` | `TestMarkTestCasesOutdated_PublishesWhenFilesChanged` без assert | Имя обещает publish при смене файлов; тело только вызывает export — ложный «зелёный» | Замокать publisher/repo side-effect и `assert` на вызов/флаг |
| A12-004 | P1 | `frontend/src/modules/settings/{api-keys,appearance,ai-privacy,webhooks,project-templates}.tsx:1` | **0** Vitest/E2E на settings | 5 UI-страниц (API keys / webhooks / AI privacy / …) вне `testframework/stages/**` | Минимум: unit на формы + e2e smoke `/settings` |
| A12-005 | P1 | `frontend/src/modules/calls/` (10 файлов; e2e touch `testframework/stages/e2e/domains/platform/a11y.spec.ts:212-228`) | Нет Vitest-домена `calls`; нет functional E2E | Security-critical модуль (LiveKit); покрытие UI = axe-визит `/calls`, не join/schedule/recording | Vitest на `create-call-dialog` / calendar + e2e smoke create/join |
| A12-006 | P1 | `testframework/stages/unit/frontend-test/domains/messenger/channel-icon.test.tsx:1-31` | Единственный messenger Vitest = иконка канала | COVERAGE/матрица дают ощущение FE-покрытия messenger; реально нет composer/threads/reactions unit | Добавить RTL на inbox/composer; не считать icon-suite покрытием модуля |
| A12-007 | P1 | `backend/testframework/tracker/service_test.go:51-77` (+ wave113–121 mock-репозитории) | Критичный tracker service слой на `mockRepo`, не Ent/SQL | ~половина tracker Test* не ловит FK/unique/optimistic-lock/SQL semantics; мок подменяет реальную проверку | Критичные CRUD/409/RBAC — на `enttest`/PG; mock оставить для pure engine |
| A12-008 | P2 | `backend/testframework/notification/consumer_wave118_test.go:94-99`, `:102-107`, `:156-160` | ShortSubject / InvalidJSON — вызов без assert side-effects | Рядом `SelfMentionSkipped` (`:134-136`) проверяет `Count==0`; эти — нет → регрессия «всё же создали notification» не поймается | После вызова: `assert.Zero` на Notification.Query().Count |
| A12-009 | P2 | `backend/testframework/ai/provider_rpm_test.go:29-32` | `SetProviderRPMLimiter` — setter без проверки | Тест всегда green; не фиксирует, что limiter применился | Getter/export или последующий `Allow` через установленный limiter |
| A12-010 | P2 | `backend/testframework/github/service_mock_test.go` + `tracker/automation_*` stubs | Широкое использование mock/stub вместо реальных NATS/HTTP/DB | Допустимо для unit, но имена вроде `PublishesWhen…` маскируют отсутствие контрактной проверки | Явный контракт: mock assert **или** integration с miniredis/nats-server/enttest |
| A12-011 | P2 | E2E inventory vs product modules | Нет e2e-доменов: `calls`, `ws`, `sprint`, `apigateway`, `github`/`gitlab`, `settings` | COVERAGE для ws/sprint/apigateway/gh уже `❌` — ок; **calls** в Module Summary помечен `⚠️`, фактически только a11y | Либо поднять calls до отдельного e2e, либо явно `❌` в COVERAGE |
| A12-012 | P2 | `backend/testframework/search/service_wave116_postgres_test.go:46` · `calls/chat_test.go:33-37` | Integration Skip при отсутствии PG/Mongo | На хосте без `docker-compose.test.yml` куски «integration» silently skip — ложное чувство полного прогона | Канон только Docker job; в отчётах CI явно считать Skip; не смешивать с unit green |
| A12-013 | P2 | `backend/testframework/calendar/` — только `handler_test.go` + `service_wave98_test.go` (**11** Test*) | Заявлено calendar **86.7%** stmt | Тонкий suite относительно %; stmt-% = CI baseline 2026-06-22 | Coverage refresh на CI; расширить CRUD/RBAC calendar API |
| A12-014 | P2 | `backend/testframework/status/` — **3** Test* (`service_test.go:12`, `:24`, `handler_test.go:15`) | COVERAGE: status **100.0%** | Для крошечного пакета возможно, но хрупко: любой новый probe без теста уронит claim | Держать % свежим; добавить fail-path probes |
| A12-015 | P3 | `backend/testframework/notification/telegram_bind_wave95_test.go:115-128`, `telegram_lifecycle_test.go:33-47`, `:93` | Кластер NilSafe/StartHelp/QueueOverflow без assert | Часть — intentional smoke; без `assert.NotPanics`/`require` регрессия = тихий no-op | Минимум `assert.NotPanics` + проверка состояния очереди/флагов |
| A12-016 | P3 | `backend/testframework/platform/config/config_test.go:125` · `project/service_wave101_test.go:150` · `backup/service_extended_test.go:181` · `audit/consumer_lifecycle_test.go:48` · `tracker/automation_engine_wave100_test.go:27` · `apigateway/webhook_dispatcher_wave97_test.go:115` · `messenger/auto_channels_service_test.go:122` · `calls/chat_mem_test.go:115` · `notification/telegram_bind_wave102_test.go:33` | Остальные assert-free (итого **21**) | Coverage-wave leftover: «вызвали код» ≠ проверка | Добавить assert или пометить `// smoke: panic-only` + `assert.NotPanics` |
| A12-017 | P3 | `backend/Makefile:6-7` (+ отсутствие hook в `go test`) | Нет защиты от протухшего generated ent | После `git pull` со сменой `ent/schema` прямой `go test` даёт непонятную panic; `make generate` / CI генерируют, голый `go test` — нет | Одной строкой: `test: generate` в Makefile |

## COVERAGE.md drift

Источник: `.planning/COVERAGE.md` (inventory refresh 2026-07-26; **stmt % = baseline 2026-06-22**).

| Заявлено в COVERAGE.md | Факт аудита 2026-07-26 | Вердикт |
|------------------------|------------------------|---------|
| BE test files **499** | **499** | ✅ match |
| internal tests **0** | **0** | ✅ match |
| Vitest **37** files · **99/99** | **37** · **99/99 PASS** | ✅ match |
| E2E **20** specs | **20** spec files (~53 tests) | ✅ match (files) |
| Overall stmt **83.0%** · security **6/6 ≥80%** | CI baseline 2026-06-22; CI всегда `go generate` перед gate — локальный stale ent **не** опровергает цифры | ⚠️ возраст baseline, не «gate broken» |
| Package table (auth 82.4% … bases 78.6%, ai 78.9%) | Не пересчитано локально; документ сам пишет «until new approved coverage gate» | ⚠️ drift risk — цифры выглядят «текущими» в таблицах |
| Module Summary: gitlab **68.5%** | В package-% таблице gitlab **отсутствует**; противоречие внутри файла | ❌ internal drift |
| Module Summary: calls E2E **⚠️** | Functional e2e/vitest **нет**; только a11y | ⚠️ understated gap |
| Module Summary: messenger E2E **⚠️** | 1 e2e (send+assert в POM) + Vitest только icon | ⚠️ FE unit overstated if read as «messenger covered» |
| Module Summary: ws / sprint / apigateway / gh E2E **❌** | Доменов e2e нет | ✅ match |
| v4 E2E ✅ `v4-superapp` | `v4-superapp.spec.ts:26-54` — route smoke (heading) | ✅ partial-but-real |
| PROJ-01 Tests **⚠️** E2E partial | `project-lifecycle.spec.ts` есть; полный tracker/TMS board flow — нет | ✅ still valid |
| BACK-01 Tests **⚠️** unit only | Backup — unit в TF, не e2e restore | ✅ still valid |
| status **100%** | 3 Test*, no-ent пакет проходит `go test` | ⚠️ claim plausible, suite thin |
| calendar **86.7%** | 2 файла / 11 Test* | ⚠️ thin vs % |
| COV-01 gate **PASS** | CI runnable (generate+hubtest); локальный stale generate — DX, не опровержение CI | ✅ CI claim holds |

**Итог drift:** file-inventory claims **синхронны**; statement-coverage — **возрастной baseline**, не «сломано из‑за ent». Локальный A12-001 = P2 DX.

## Прогоны (evidence)

```
# Vitest (frontend)
Test Files  37 passed (37)
Tests       99 passed (99)

# Go hubtest (после go generate ./ent — FIX-W0)
ok   testframework/auth
ok   testframework/status
ok   testframework/platform/...
# На stale generated tree (без generate) — локальный ent.init panic (A12-001 P2 DX)
```

Flakes: не наблюдались. Skip-integration не исполнялись (нет PG/Mongo на хосте аудита).

## Чисто (проверено, проблем нет / ок)

- Инвентарь TF: **499 / 0 internal** — совпадает с COVERAGE и AGENTS.md
- Vitest green **99/99**, домены v4 покрывают accounting/crm/hr/bases/… page smoke
- E2E messenger POM (`messenger.page.ts:21`) реально `expect(getByText)` после send
- E2E project-lifecycle — полноценный create/board/trash/restore
- Assert-free доля мала (~21 / ~2741 ≈ **0.8%**); большинство suite — с `require`/`assert`
- `testframework/PIPELINE.md` карта стадий соответствует дереву на диске
- CI/Dockerfile: `go generate ./ent` перед тестами — hubtest **operational** на CI

## Оценка зоны: 6/10

Инвентарь и Vitest зрелые; backend hubtest **не** заблокирован на CI (A12-001 downgrade: локальный stale generate). Остаются assert-free «coverage filler», mock-heavy tracker, дыры FE (`settings`, `calls`) и возрастной stmt-% в COVERAGE.md. До ~7.5/10 после закрытия P1 assert-free и FE gaps.

### Сводка counts

| Severity | Count |
|----------|------:|
| P0 | 1 |
| P1 | 6 |
| P2 | 7 |
| P3 | 2 |
| **Total** | **16** |
| Assert-free Go tests | **21** |
| Vitest | **99/99 PASS** |
| Go hubtest (sampled) | **3 ok / ent-pkgs FAIL init** |
| **Score** | **4/10** |
