# Best Practices 2026 — Backend (сверка с playbook)

> Дата: 2026-07-03 · Исследование: WebSearch (go.dev release notes 1.24/1.25, gofiber docs/releases v3.0.0–v3.3.0, entgo.io/atlasgo.io CI-доки, OTel 2026 semantic conventions, Go supply-chain hardening guides, PostgreSQL 17/18 release notes)
> Сверялось с: `.cursor/playbooks/backend.md` (19 разделов), `.cursor/playbooks/observability.md`, `docs/stack_and_tools.md`, `backend/go.mod`, `.github/workflows/ci.yml`, `backend/ent/migrate/`
> Статус: **отчёт-дельта, код и playbook не менялись**

---

## 1. Резюме

Найдено **22 дельты**: 10 — «принять», 9 — «обсудить», 3 — «отклонить/наблюдать». Playbook в целом очень зрелый: ключевые инварианты Hub (outbox, идемпотентность, SSRF-защита с анти-DNS-rebinding, constant-time compare, optimistic locking, slog-контекст) полностью соответствуют состоянию индустрии 2026. Основные дельты — не про «плохие практики», а про (а) устаревшие зависимости/API, (б) отсутствующие инструменты в CI, (в) новые возможности stdlib Go 1.24/1.25, которых не было на момент написания playbook.

**Топ-5 важных дельт:**

1. **Fiber v3 вышел в stable (v3.0.0, текущая v3.3.0), Hub сидит на `v3.0.0-beta.3`.** В stable переименован proxy-конфиг (`EnableTrustedProxyCheck`/`TrustedProxies` → `TrustProxy`/`TrustProxyConfig`), часть `fiber.Config` переехала в `ListenConfig`, middleware перешли на пакет `extractors`. Playbook §6.2 документирует beta-API — при апгрейде примеры перестанут компилироваться. Beta в production — риск сам по себе (в v3.0.0–v3.3.0 закрыто много security-фиксов: host normalization, cookie handling, oversized bodies).
2. **Supply-chain безопасность в CI отсутствует.** В `ci.yml` нет `govulncheck`, `gosec`, `go mod verify`, secrets-скана. Это стандартный минимум 2026 (govulncheck — официальный инструмент Go team с call-graph анализом, почти нулевой шум). Playbook §7 покрывает код, но не цепочку поставки.
3. **Миграции: drift между playbook и кодом + отсутствие Atlas lint.** Playbook §5.4 описывает versioned-миграции Atlas («миграции идут в контроль версий»), но в `backend/ent/migrate/` нет каталога миграций — фактически используется авто-миграция `Schema.Create`. Индустрия-2026: versioned migrations + `atlas migrate lint` в CI (ловит destructive/data-dependent изменения до мержа). Для монолита с одной БД авто-миграция терпима, для курса на микросервисы (schema-per-service) — блокер.
4. **Новые тестовые инструменты stdlib: `testing/synctest` (stable в 1.25), `testing.B.Loop` (1.24), `T.Context()` (1.24).** Playbook §10.8 полагается только на `-race`; `synctest` даёт детерминированное тестирование конкурентного кода с виртуальными часами (идеально для outbox-worker, rate-limiter cleanup, WS hub — сейчас такие тесты либо flaky, либо со `sleep`). Не отражены в §10/§13.
5. **Outbox/JetStream: три уточнения к готовности масштабироваться.** (а) Worker без `SELECT ... FOR UPDATE SKIP LOCKED` нельзя запускать в 2+ экземплярах — будут дубли; (б) JetStream publish-side дедуп через заголовок `Nats-Msg-Id` (окно дедупликации стрима) дополняет consumer-side SETNX из §8.6; (в) trace context (traceparent) не пропагируется через outbox-таблицу и NATS-заголовки — трейс рвётся на границе события. Всё три — стандартные практики 2026 для event-driven Go.

---

## 2. Таблица дельт

| # | Тема | Что в playbook сейчас | Свежая практика 2026 | Рекомендация | Куда вписать |
|---|------|----------------------|----------------------|--------------|--------------|
| 1 | Fiber v3 stable | §6 документирует API `v3.0.0-beta.3`: `EnableTrustedProxyCheck` + `TrustedProxies` в `fiber.Config` | Fiber v3.0.0 stable (v3.3.0 — текущая, требует Go 1.25): `TrustProxy` + `TrustProxyConfig`, `ListenConfig` (`DisableStartupMessage`, `EnablePrefork` и др. переехали из `fiber.Config`), пакет `extractors` для middleware, `Bind().All()` с фиксированным приоритетом источников, 17 сигнатур handler'ов, CLI `fiber migrate --to v3`. В stable закрыты security-фиксы (host normalization, oversized bodies, cookie encryption fallback) | **Принять** (запланировать апгрейд beta.3 → v3.x; после апгрейда обновить примеры) | §6.1, §6.2, §6.3 |
| 2 | govulncheck в CI | Не упоминается; §16.1 — только `build` + `vet` + `test -race` | `govulncheck ./...` — официальный сканер Go team; call-graph анализ отсекает недостижимые CVE (низкий шум). Стандарт: шаг в CI + scheduled run для новых advisories | **Принять** | §16.1 + `ci.yml` |
| 3 | Supply chain: verify/tidy | Не упоминается | `go mod verify` + `go mod tidy -diff` (Go 1.23+) ранними шагами CI: целостность module cache и отсутствие подмешанных зависимостей | **Принять** | §16.1 + `ci.yml` |
| 4 | Secrets scanning | §15.3/`50-git.mdc`: ручной `git diff --staged \| grep -E "(secret\|password\|...)"` | Автоматизированный скан (gitleaks/trufflehog) в CI + pre-commit; grep по diff даёт и пропуски, и шум | **Принять** | §15.4 + `ci.yml` |
| 5 | gosec / SAST | §7 — инварианты грепаются be-ver вручную (`rg`-набор §17.3) | `gosec` (SSA + taint-анализ: SQLi, SSRF, path traversal, слабое крипто) как CI-gate; SARIF-отчёт в GitHub code scanning. Дополняет, не заменяет грепы be-ver | **Принять** | §17.3 + `ci.yml` |
| 6 | Atlas versioned migrations | §5.4: «миграции применяются через Atlas, идут в контроль версий» — но в коде `Schema.Create` (авто-миграция), каталога миграций нет | Versioned migrations + `atlas migrate lint` в CI (dev-database в services): ловит destructive/data-dependent изменения, защищает историю от конкурентных правок. Обязательный шаг перед schema-per-service | **Принять** (как минимум устранить drift: либо признать авто-миграцию в §5.4, либо мигрировать на versioned) | §5.4 + `ci.yml` |
| 7 | testing/synctest | §10.8: только `go test -race`; конкурентные тесты без управления временем | `testing/synctest` — stable в Go 1.25: изолированный «пузырь» с виртуальными часами, `synctest.Wait()`. Убирает `time.Sleep` и flaky из тестов воркеров/тикеров/таймаутов | **Принять** | §10.8, §13 |
| 8 | testing.B.Loop | Бенчмарки не покрыты playbook'ом | `for b.Loop()` вместо `for range b.N`: авто-таймер setup/cleanup, защита от dead-code elimination, один прогон функции на замер. Официально предпочтительный способ с Go 1.24 | **Принять** (короткий пункт: «если пишете бенчмарк — только `b.Loop()`») | §13 (новый подраздел) |
| 9 | T.Context() | В тестах используется `context.Background()` (§13.3 пример) | `t.Context()` (Go 1.24) — контекст, отменяемый по завершении теста; корректно гасит горутины/подключения теста | **Принять** | §13.2–13.3 |
| 10 | Trace context через события | observability §5.4 — пропагация только в исходящих HTTP-вызовах; outbox/NATS не несут trace context | Стандарт 2026: сериализовать span context в колонку outbox-таблицы, worker инжектит `traceparent`/`tracestate` в NATS headers (`TextMapPropagator` + header carrier); консьюмер извлекает → единый трейс producer → outbox → consumer | **Принять** | §8.4–8.5 + observability §5.4 |
| 11 | Outbox: SKIP LOCKED | §8.5: worker поллит `StatusPending` без блокировок — корректно только для одного экземпляра | `SELECT ... FOR UPDATE SKIP LOCKED` в выборке батча — безопасное горизонтальное масштабирование relay-воркеров. Критично для v2.6 arch/микросервисов | **Обсудить** (сейчас один инстанс — не баг; вписать как предусловие масштабирования) | §8.5 |
| 12 | JetStream publish-дедуп | §8.6: только consumer-side дедуп (Redis SETNX) | Заголовок `Nats-Msg-Id` (= outbox event ID) + `Duplicates` window стрима: JetStream сам отбрасывает повторную публикацию после ретрая worker'а. «Exactly-once» = Msg-Id дедуп + double-ack + идемпотентный консьюмер | **Принять** | §8.5–8.6 |
| 13 | Go 1.25 container-aware GOMAXPROCS | Не упоминается | Рантайм Go 1.25 сам читает cgroup CPU-лимиты (v1/v2) и выставляет GOMAXPROCS, перечитывает при изменении. `automaxprocs` больше не нужен; но контейнерам нужны явные CPU limits в compose/K3s, иначе поведение прежнее | **Принять** (заметка в §11 + чек devops: CPU limits у backend-контейнера) | §11.5 + deploy docs |
| 14 | CSRF: Sec-Fetch-Site | CSRF-защита не описана (JWT в header — классический CSRF неприменим, но cookie-flow WS/`/ws?token=` и будущие SSO-редиректы — зона риска) | Go 1.25 `http.CrossOriginProtection`: header-based CSRF (Sec-Fetch-Site/Origin), без токенов. Для Fiber — аналогичная проверка в middleware. Актуально при ADR-AUTH-SSO (cookie-редиректы OIDC) | **Обсудить** (при реализации SSO — обязательный пункт security-ревью) | §7 (новый подраздел при SSO) |
| 15 | Iterators / range-over-func | Не упоминается | `iter.Seq`/`iter.Seq2` — стандартный протокол итераторов; stdlib уже перешёл (`strings.SplitSeq`, `bytes.Lines`). Generic type aliases (1.24) упрощают сигнатуры. Уместно для стриминга больших выборок (экспорт, backup) вместо загрузки слайса в память | **Обсудить** (принять для новых пайплайн-API; не рефакторить существующее) | §11 или §2 (соглашения) |
| 16 | Ent interceptors для soft-delete | §5.8: ручной `DeletedAtIsNil()` в каждом Query — человеческий фактор (одно забытое место = утечка удалённых данных) | Ent Interceptors/Privacy: глобальный traversal-фильтр soft-delete + tenant-scope на уровне ORM (fail-closed). §5.9 упоминает privacy «осознанно», но паттерн не применён | **Обсудить** (сильный выигрыш по безопасности; требует аккуратного внедрения + обхода для корзины/GDPR) | §5.8–5.9 |
| 17 | Contract testing | Не упоминается (монолит) | Pact (+ pact-go v2) жив и является стандартом consumer-driven contracts; для NATS-событий — Pact async message contracts либо schema registry. Предусловие распила на сервисы из `MICROSERVICES-ROADMAP` | **Обсудить** (вписать в roadmap v2.6, не в текущий playbook-минимум) | §8 или отдельный arch-doc |
| 18 | OTel semantic conventions v1 | observability §5: свои имена span'ов/метрик; конвенции не зафиксированы | SemConv v1 заморожены: `http.request.method`, `http.response.status_code`, `db.system.name`, `messaging.system` и т.д. Использование стандартных имён = совместимость с готовыми Grafana-дашбордами и предсказуемость | **Обсудить** (аудит текущих имён метрик/span'ов на соответствие; переименование — breaking для дашбордов) | observability §4–5 |
| 19 | Trace sampling | observability §5: сэмплирование не описано (по умолчанию AlwaysSample) | Прод-стандарт: `ParentBased(TraceIDRatioBased(N))` + tail-based sampling на коллекторе для error-трейсов. 100% трейсинг под нагрузкой дорог | **Принять** (одна строка конфига + абзац в playbook) | observability §5.6 |
| 20 | PostgreSQL 17 JSON_TABLE | Не упоминается; JSONB-поля (`description_json`) обрабатываются на стороне Go | PG17: `JSON_TABLE()` разворачивает JSONB в реляционные строки прямо в SQL — заменяет `jsonb_array_elements` и app-level парсинг для аналитики/дашбордов | **Обсудить** (полезно для dashboard-запросов; через Ent — только raw SQL/Modify) | §5.10 |
| 21 | UUIDv7 для PK | §5.1/TZ: UUID v4 обязателен (`uuid.New`) | Индустрия смещается к UUIDv7 (time-ordered, RFC 9562): меньше фрагментация B-tree на insert-heavy таблицах; `google/uuid` имеет `uuid.NewV7()`, PG18 — встроенный `uuidv7()` | **Отклонить** сейчас (TZ-инвариант «UUID v4» — менять только через ADR; выигрыш заметен на десятках млн строк, Hub далёк от этого) | — (ADR-кандидат) |
| 22 | GOEXPERIMENT=jsonv2 | Не упоминается | Go 1.25: экспериментальный `encoding/json/v2` (быстрее, строже). Ещё experimental | **Отклонить/наблюдать** (не вносить в правила до stable) | — |

Дополнительно (микро-дельты, без строки в таблице): `testing.AllocsPerRun` теперь паникует при параллельных тестах (Go 1.25) — учитывать в бенчмарк-тестах; `T.Attr`/`T.Output` (1.25) — полезно для Allure-конвейера testframework; экспериментальный GC greenteagc (1.25) — можно попробовать в staging, в прод не включать.

---

## 3. Что в playbook УЖЕ соответствует состоянию 2026 (подтверждение)

| Область | Подтверждение |
|---------|---------------|
| **Transactional outbox** (§8.4–8.5) | Паттерн, изоляция сбоев через recover, at-least-once, ретраи с attempts-лимитом — ровно то, что рекомендуют свежие гайды (open-outbox spec, BackendBytes 2026). Публикация в той же транзакции через `PublishInTx` — каноничное решение dual-write |
| **Идемпотентность консьюмеров** (§8.6) | Дедуп по event ID / upsert / естественная идемпотентность — текущий стандарт для at-least-once |
| **SSRF-защита** (§7.1) | `SafeDialContext` с резолвом перед подключением (анти-DNS-rebinding) + блок приватных диапазонов включая 169.254.169.254 — уровень выше среднего по индустрии; gosec taint-анализ подтверждает те же паттерны |
| **`io.LimitReader`, `subtle.ConstantTimeCompare`, fail-closed webhook** (§7.2–7.4) | Без изменений — актуальные инварианты OWASP-класса |
| **slog: структурные пары, `*Context`-варианты, уровни, запрет секретов** (§12, observability §2) | Полностью соответствует best practices slog 2026 (контекстное логирование, correlation id, сэмплирование шумных логов) |
| **`errors.Join` для агрегирования частичных ошибок** (§5.8, GDPR-каскад) | Уже применяется — соответствует идиоме |
| **Optimistic locking `VersionEQ`+`AddVersion`+`affected==0`** (§5.6) | Каноничный паттерн для Ent, изменений в индустрии нет |
| **Eager loading `.WithX()` против N+1, жёсткие лимиты пагинации, cursor для лент** (§5.7, §11) | Актуально без правок |
| **NATS: core для эфемерного, JetStream durable для критичного, InterestPolicy, иерархия subjects** (§8.1–8.3, 8.7) | Соответствует рекомендациям NATS-гайдов 2026 (queue groups, durable consumers, иерархические subjects) |
| **`ParentBased`-совместимая пропагация trace в HTTP, child-span'ы бизнес-операций** (observability §5.2–5.4) | Архитектура верная; дельты только в sampling (#19) и событиях (#10) |
| **`-race` + `go vet` + `build -tags release` перед handoff** (§16.1) | База осталась базой; дельты — дополнение (#2–5, #7), не замена |
| **Trusted proxy check в Fiber** (§6.2) | Концепт обязателен и в stable v3 (там это `TrustProxy`) — Hub уже делает это правильно, меняется только имя поля |
| **Table-driven тесты, testify, black-box `_test`-пакеты, моки через Repository-интерфейс** (§13) | Актуально; synctest/B.Loop/T.Context — дополнения |
| **Dependabot weekly** (`stack_and_tools.md`) | Есть и соответствует; govulncheck дополняет его reachability-анализом |

---

## 4. Новые инструменты для CI (с обоснованием)

Приоритет с учётом лимита GitHub Actions (~3000 мин/мес — экономить):

| Инструмент | Что даёт | Цена в CI | Приоритет |
|------------|----------|-----------|-----------|
| **govulncheck** (`golang.org/x/vuln/cmd/govulncheck`) | Официальный сканер уязвимостей Go team; call-graph — репортит только реально достижимые CVE (почти нет false positives). Закрывает окно «уязвимость опубликована → dependabot ещё не поднял версию» | ~1–2 мин; кэшируется вместе с go-кэшем | **P0** — добавить в `backend-test` job |
| **go mod verify + go mod tidy -diff** | Целостность module cache; выявление подмешанных/устаревших записей go.sum. Две строки, нулевая стоимость | Секунды | **P0** — первые шаги перед build |
| **gitleaks** | Автоматический secrets-скан diff'а вместо ручного grep из `50-git.mdc`; PG-03 (ротация секретов) показал, что секреты в git — реальный риск Hub | ~30 сек | **P1** |
| **atlas migrate lint** (`ariga/atlas-action`) | Ловит destructive/data-dependent миграции до мержа. Требует сначала перейти на versioned migrations (#6) | ~1–2 мин + dev-database container | **P1** (после решения по #6) |
| **gosec** | SAST: taint-анализ SQLi/SSRF/path traversal/крипто. Автоматизирует часть грепов be-ver; SARIF → GitHub code scanning | ~1–3 мин | **P2** (частично дублирует be-ver-грепы и ручной аудит) |
| **osv-scanner** | Второе мнение по зависимостям (OSV-база шире Go vulndb, но без reachability) | ~30 сек | **P3** — опционально, scheduled weekly, не в каждый PR |
| **slsa-github-generator (SLSA L3 provenance) + SBOM** | Подписанная провенанс-цепочка релизных образов GHCR | Заметная (отдельный job на релиз) | **P3** — только на release-тег, не на PR; для self-hosted продукта ценность средняя |

Рекомендуемая связка на PR: `go mod verify` → `go mod tidy -diff` → build/vet/test (как сейчас) → `govulncheck ./...` → gitleaks. Scheduled (cron, раз в сутки): govulncheck по master (новые advisories появляются без изменения кода) + osv-scanner.

---

## 5. Источники

- go.dev/doc/go1.24, go.dev/doc/go1.25 — release notes (B.Loop, synctest, T.Context, generic aliases, GOMAXPROCS, CrossOriginProtection, jsonv2)
- go.dev/blog/testing-b-loop — обоснование B.Loop
- github.com/gofiber/fiber releases v3.0.0–v3.3.0; docs.gofiber.io «What's new in v3» (2026-02-11) — stable API, TrustProxy, ListenConfig, extractors, migration CLI
- entgo.io/docs/ci, entgo.io/docs/versioned/verifying-safety, atlasgo.io/versioned/lint — Atlas lint в CI; entgo hooks/privacy docs
- go.dev/doc/security/vuln, systemshardening.com (Go module supply chain), jlevy/supply-chain-hardening — govulncheck, go mod verify/tidy -diff, GOPROXY/GOSUMDB
- securego/gosec — SAST-возможности 2026 (taint-анализ)
- OTel 2026 deep dive (youngju.dev), dev.to distributed tracing guide 2026, oneuptime exemplars — semconv v1 lock, sampling, exemplars, NATS header carrier
- BackendBytes «Event-Driven Microservices in Go» 2026, open-outbox/relay, oneuptime NATS guide 2026 — SKIP LOCKED, Msg-Id дедуп, trace через outbox
- postgresql.org/docs/17/release-17, обзоры PG17/PG18 2026 — JSON_TABLE, MERGE RETURNING, vacuum, uuidv7 (PG18)
