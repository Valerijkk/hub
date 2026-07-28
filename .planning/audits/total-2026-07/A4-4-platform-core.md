# W4 агент A4 — platform config/db/crypto/core

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: secrets in config defaults, AES-GCM misuse, DB pool, listen SO_REUSE, crypto.  
> Сверка: A14-* (config fail-closed JWT/S3/LiveKit/NATS/metrics), A8-007 (mongo disconnect), A8-008 (listen cancel), HI-05 (AES-GCM), INFRA.md §4 (Schema.Create).

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/platform/config/config.go` | 356 | 3 | замечания |
| `backend/internal/platform/config/export_hooks.go` | 18 | 0 | чисто |
| `backend/internal/platform/cron/cron.go` | 94 | 1 | замечания |
| `backend/internal/platform/crypto/aesgcm.go` | 165 | 1 | замечания |
| `backend/internal/platform/database/mongodb.go` | 38 | 0 | чисто |
| `backend/internal/platform/database/postgres.go` | 42 | 3 | замечания |
| `backend/internal/platform/errors/errors.go` | 61 | 1 | замечания |
| `backend/internal/platform/errors/messages.go` | 17 | 0 | чисто |
| `backend/internal/platform/health/probes.go` | 65 | 2 | замечания |
| `backend/internal/platform/lifecycle/listen.go` | 16 | 0 | чисто |
| `backend/internal/platform/telegram/markdown.go` | 27 | 1 | замечания |
| `backend/testframework/platform/config/config_test.go` | 279 | 1 | замечания |
| `backend/testframework/platform/cron/cron_test.go` | 66 | 0 | чисто |
| `backend/testframework/platform/crypto/aesgcm_test.go` | 99 | 0 | чисто |
| `backend/testframework/platform/database/mongo_disconnect_a8007_test.go` | 48 | 0 | чисто |
| `backend/testframework/platform/errors/errors_test.go` | 149 | 0 | чисто |
| `backend/testframework/platform/errors/errors_wave101_test.go` | 21 | 0 | чисто |
| `backend/testframework/platform/health/probes_test.go` | 84 | 0 | чисто |
| `backend/testframework/platform/lifecycle/listen_a8008_test.go` | 31 | 0 | чисто |
| `backend/testframework/platform/telegram/markdown_test.go` | 29 | 0 | чисто |

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W4-A4-001 | P1 | `postgres.go:21-35` | После `sql.Open` / `ent.Driver` **нет** `SetMaxOpenConns` / `SetMaxIdleConns` / `SetConnMaxLifetime` / `SetConnMaxIdleTime`. По репо `SetMaxOpenConns` нигде не вызывается | `database/sql` при `MaxOpenConns=0` = безлимит. Всплеск нагрузки / утечка conn → исчерпание `max_connections` Postgres → каскадный отказ всего API (health/outbox/Ent) | После `sql.Open` (или на `drv.DB()`): лимиты из env (напр. open=25–50, idle=10, lifetime=30m). Тест: после N параллельных запросов `Stats().OpenConnections ≤ MaxOpen` |
| W4-A4-002 | P1 | `config.go:159-170`; `config.go:223-282` | Prod `validate` fail-closed для JWT/S3/LiveKit/INTEGRATION_KEY/NATS/metrics, но **не** для `DATABASE_URL` / `MONGODB_URL` / `REDIS_URL`: дефолты `hub:hub@…` остаются валидны при `DEV_MODE=false` + `ENVIRONMENT=production`, если env не задан | Симметрия с A14-002 нарушена: процесс стартует с известными weak DB credentials (как SeaweedFS `hub/hubsecret` до фикса). Риск misconfig / accidental local bind к open Postgres | В `validate` при `!DevMode` (или `ENVIRONMENT=production`): отвергать URL с user/pass `hub`/`hub`, localhost-only без явного allow; тест table-driven как A14-002 |
| W4-A4-003 | P2 | `postgres.go:28-32` | `client.Schema.Create(ctx)` **всегда** (включая production). Комментарий «Atlas для prod» не отражён в коде | Auto-migrate на каждом boot: drift, гонки двух реплик при деплое, риск silent schema change (INFRA.md §4 residual) | `Schema.Create` только при `DevMode`; prod — `atlas migrate apply` в deploy. Тест: mock/spy или build-tag assert |
| W4-A4-004 | P2 | `config.go:160`; `config.go:223-282` | Дефолт `sslmode=disable`; production validate **не** требует TLS к Postgres | Cleartext PG credentials/traffic вне docker-network (K3s/multi-host). Compose single-host ранее accepted (M-08), но config не гейтит ENVIRONMENT=production | При `ENVIRONMENT=production` требовать `sslmode=require|verify-ca|verify-full` (исключение только явный `ALLOW_INSECURE_DB_SSL=true`) |
| W4-A4-005 | P2 | `errors.go:7-13` | `AppError` JSON: `code` + `message` (+ `request_id`). TZ / `tz_constraints.md`: `{"error","request_id"}` | Контракт ошибок расходится с обязательным форматом ТЗ; клиенты/wiki могут ожидать другое поле | Либо сменить теги/`ErrorHandler` на `{"error": Message, "request_id"}`, либо ADR + правка `tz_constraints.md` (осознанный drift). Не оставлять молча |
| W4-A4-006 | P2 | `markdown.go:12-15` | В `markdownV2Special` **нет** `\` | Telegram MarkdownV2 требует экранировать backslash первым; user-input с `\` → reject Bot API / сломанная разметка (auth/notification delivery) | Добавить `"\\"` **первым** в список; тест `"a\\b" → "a\\\\b"` |
| W4-A4-007 | P2 | `probes.go:39-41`; `probes.go:75-83` (тест фиксирует поведение) | `Ready`: `fn == nil` → skip; все nil → **200** `status=ok`, пустой `services` | Miswiring readiness = K8s/Compose считает под healthy без Postgres/Redis | Fail-closed: если ни одного ping не задано → 503; или требовать минимум postgres. Тест: `Deps{}` → 503 |
| W4-A4-008 | P2 | `cron.go:61-63`; `cron.go:72-94` | `runTick` вызывает `fn(ctx)` без deadline; зависший tick блокирует горутину до возврата (отмена `Schedule` не прерывает CPU-bound/игнор ctx) | Один bad job (backup/cleanup) стопорит дальнейшие tick'и этой задачи навсегда при hang | `context.WithTimeout(ctx, tickBudget)` вокруг `fn`; лог timeout. Тест: fn sleep > budget → error path, следующий tick идёт |
| W4-A4-009 | P2 | `config.go:260-264`; `aesgcm.go:87-89` | Prod требует непустой `INTEGRATION_ENCRYPTION_KEY`, но **не** запрещает известный `DeriveDevKey()` (hex/base64 SHA-256 от `hub-dev-integration-encryption-key-v1`) | Dev-ключ at-rest для GitHub/GitLab tokens в «production» конфиге — эквивалент hardcoded secret | В validate: decode key → `subtle.ConstantTimeCompare` с `DeriveDevKey()` → reject. Тест на hex(dev key) |
| W4-A4-010 | P3 | `postgres.go:38-41` | `OpenRawDB` — второй `sql.Open` без pool limits; **ни одного caller** в репо | Мёртвый опасный API: будущий вызов = второй unbounded pool рядом с Ent | Удалить или обернуть shared `*sql.DB` + те же лимиты, что W4-A4-001 |
| W4-A4-011 | P3 | `config_test.go:276-279` | `TestLoad_clearsEnvBetweenTests` — только `os.Unsetenv`, **без assert** | Assert-free: зелёный при любом коде; не ловит регрессию | Удалить или заменить на Load()+assert DevMode/err |
| W4-A4-012 | P3 | `probes.go:33-34` | Ready пинги на `context.Background()+timeout`, не `c.Context()` | Отмена HTTP-клиента не отменяет Ping; лишняя нагрузка при probe storm | `context.WithTimeout(c.Context(), timeout)` |
| W4-A4-013 | P3 | `aesgcm.go:102`; `aesgcm.go:147-165` | `Seal/Open` с `additionalData=nil`; `MaybeDecryptString` soft-passthrough legacy | Без AAD ciphertext не привязан к resource-id (перестановка blob между rows не детектится на crypto-слое). Soft decrypt — ок для migration, но callers часто игнорируют `ok` (вне списка: github/gitlab/ai) | Опционально AAD=`entityID`; сузить `MaybeDecrypt` / deprecate после миграции. Не баг nonce: `rand.Reader` + 12 байт — корректно |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| `config.go` A14-* | JWT min 32; `devDefaults` JWT/S3/LiveKit; `DEV_MODE`+`ENVIRONMENT=production` conflict; NATS user/pass в production; empty `METRICS_ALLOWED_CIDRS` fail-closed; `INTEGRATION_ENCRYPTION_KEY` обязателен при `!DevMode`; TTL access 15m / refresh 30d |
| `config/export_hooks.go` | hubtest re-export `isDevDefault` / getEnv helpers — без prod-логики |
| `crypto/aesgcm.go` (ядро) | AES-256 key len check; hex/base64 decode; **unique random nonce** (`io.ReadFull(rand.Reader)`); GCM Open rejects tamper; nil cipher → `ErrCipherKeyNotConfigured` (no panic); `DeriveDevKey` задокументирован DEV-only |
| `database/mongodb.go` | Connect+Ping; `MongoDisconnectCleanup` 5s timeout, nil-safe (A8-007 закрыт) |
| `lifecycle/listen.go` **SO_REUSE** | Только `CancelOnListenError` (A8-008): log + `cancel()`, **не** `os.Exit`. Prefork/`reuseport` **не** используются — `app.Listen` в wiring = exclusive `net.Listen`, SO_REUSEPORT не включён → нет Windows SO_REUSEADDR multi-bind сюрприза |
| `errors/messages.go` | Константы Msg* — без логики |
| `cron.go` (базово) | ctx cancel stop; recover+stack на panic; error tick не убивает loop |
| `config_test.go` (кроме assert-free) | A14-002 per-key, A14-003 NATS, A14-005 metrics, A14-007 stacktrace, encryption required |
| `aesgcm_test.go` | Roundtrip; nonce uniqueness; tamper; hex/b64; MaybeDecrypt legacy |
| `cron_test.go` | Tick / cancel / error-continues — asserts есть |
| `mongo_disconnect_a8007_test.go` | Ping fail after Disconnect; nil cleanup |
| `errors_*_test.go` | Constructors + RetryAfter min 1s |
| `probes_test.go` | Live 200; Ready ok/503; nil-skip (фиксирует W4-A4-007) |
| `listen_a8008_test.go` | cancel invoked; nil cancel safe |
| `markdown_test.go` | Набор спецсимволов кроме `\` — mutation на missing `\` не краснеет (см. W4-A4-006) |

**AES-GCM вердикт:** misuse классический (reuse nonce / ECB / без auth) **не** найден. Остаточный риск — nil AAD + soft `MaybeDecrypt` + отсутствие ban на DeriveDevKey в prod config.

**DB pool вердикт:** shared Ent/`rawDB` pool для health — хорошо (BE-H5); **лимиты пула отсутствуют** — главный gap.

**SO_REUSE вердикт:** в scope A4 нет включения reuseport; listen helper корректен.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **20 / 20 / 0** |
| Находок | **13** |
| P0 | 0 |
| P1 | 2 |
| P2 | 7 |
| P3 | 4 |

### Топ-3

1. **W4-A4-001 (P1)** — Postgres/Ent pool без `SetMaxOpenConns` (и idle/lifetime): безлимитные conn → outage под нагрузкой.
2. **W4-A4-002 (P1)** — production validate не отвергает weak `DATABASE_URL`/`MONGODB_URL`/`REDIS_URL` (`hub:hub`), в отличие от JWT/S3/LiveKit.
3. **W4-A4-003 (P2)** — `Schema.Create` на каждом старте в production (auto-migrate vs заявленный Atlas).
