# A3 W4 агент A4 — platform config/db/crypto/core (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline: [A2W4-4-platform-core.md](A2W4-4-platform-core.md) · Stage A: [A4-4-platform-core.md](A4-4-platform-core.md)  
> Assignment: [W4-ASSIGNMENTS.md](W4-ASSIGNMENTS.md) § A4 · Периметр: **20** · Прочитано: **20/20** · Пропущено: **0**  
> Фокус: secrets in config defaults · crypto misuse · DB connect fail-closed · probe auth · lifecycle cancel · Telegram markdown injection  
> Remediation evidence: TASK-112 (A2W4-4-001 AAD) · W4-A4-001…012 cascade

> **Вне списка assignment (evidence FIXED-OK, не в таблице чтения):**  
> `testframework/platform/config/derive_ssl_task043_test.go` (DeriveDevKey / sslmode),  
> `testframework/platform/cron/timeout_task043_test.go` (tick budget),  
> `testframework/platform/database/postgres_pool_p1be25_test.go` + `automigrate_task043_test.go`,  
> `testframework/platform/telegram/backslash_task043_test.go`,  
> `health/export_hooks.go` (`ExportProbeParentContext`).

## Сверка с A2 (A2W4-4-*)

| A2 ID | Tag | Статус в HEAD | Комментарий |
|-------|-----|---------------|-------------|
| A2W4-4-001 | FIXED-OK | закрыт | TASK-112: `EncryptWithAAD`/`DecryptWithAAD` + `TestEncryptDecryptWithAAD_RoundtripAndMismatch` (`aesgcm.go:122–176`, `aesgcm_test.go:101–126`). Soft `MaybeDecryptString` оставлен documented migration window |

**REGRESS:** 0 · **A2 open residual:** 0

### Cascade FIXED-OK (Stage A W4-A4-* / A2 FIXED-OK)

| Stage A / A2 | Evidence (файл:строка) |
|--------------|------------------------|
| W4-A4-001 | `postgres.go:20–33`, `applyPoolLimits` MaxOpen=25 / idle / lifetime |
| W4-A4-002 | `config.go:299–309` `weakDataStoreURLProblem` hub:hub@; tests `config_test.go:225–278` |
| W4-A4-003 | `postgres.go:51–56` `Schema.Create` только `DEV_MODE` |
| W4-A4-004 | `config.go:287–296`, `326–344` sslmode require\|verify-*; override `ALLOW_INSECURE_DB_SSL` |
| W4-A4-005 | `errors.go:8–16` JSON `error` + `request_id`; Code/Details `json:"-"` |
| W4-A4-006 | `markdown.go:13–16` `\` первым; `markdown_test.go:19` + backslash_task043 (вне слота) |
| W4-A4-007 | `probes.go:60–66` checked==0 → 503; `probes_test.go:75–88` |
| W4-A4-008 | `cron.go:96–97` `WithTimeout`; timeout_task043 (вне слота) |
| W4-A4-009 | `config.go:275–278` + `aesgcm.go:104–112` `IsDevDerivedKey`; derive_ssl_task043 (вне слота) |
| W4-A4-010 | `OpenRawDB` удалён (pool-тест комментарий) |
| W4-A4-011 | `config_test.go:344–358` Load+assert DevMode |
| W4-A4-012 | `probes.go:35`, `79–81` parent = `c.Context()`; `probes_test.go:90–106` |

## Прочитанные файлы

| Файл | Строк | Находок (open) | Вердикт |
|------|------:|---------------:|---------|
| `backend/internal/platform/config/config.go` | 418 | 0 | FIXED-OK secrets/validate A14 + weak URL + TLS + DeriveDevKey |
| `backend/internal/platform/config/export_hooks.go` | 18 | 0 | чисто — hubtest re-export |
| `backend/internal/platform/cron/cron.go` | 109 | 0 | FIXED-OK tick timeout + recover + ctx stop |
| `backend/internal/platform/crypto/aesgcm.go` | 215 | 0 | FIXED-OK AAD API; nonce/rand; nil→Err; soft MaybeDecrypt documented |
| `backend/internal/platform/database/mongodb.go` | 38 | 0 | чисто — Connect+Ping fail-closed; disconnect cleanup |
| `backend/internal/platform/database/postgres.go` | 69 | 1 | NEW — нет Ping после Open (асимметрия Mongo/Redis) |
| `backend/internal/platform/errors/errors.go` | 64 | 0 | FIXED-OK TZ `{"error","request_id"}` |
| `backend/internal/platform/errors/messages.go` | 17 | 0 | чисто — Msg* константы |
| `backend/internal/platform/health/probes.go` | 82 | 0 | FIXED-OK Ready fail-closed + RequestCtx parent; Live unauth intentional |
| `backend/internal/platform/lifecycle/listen.go` | 16 | 0 | FIXED-OK cancel-only, не os.Exit |
| `backend/internal/platform/telegram/markdown.go` | 28 | 0 | FIXED-OK MarkdownV2 `\` first |
| `backend/testframework/platform/config/config_test.go` | 359 | 0 | чисто — A14/P1-BE-26; DeriveDevKey/ssl вне слота |
| `backend/testframework/platform/cron/cron_test.go` | 66 | 1 | NEW — mutation gap: нет assert на tick timeout |
| `backend/testframework/platform/crypto/aesgcm_test.go` | 127 | 0 | FIXED-OK AAD mismatch + tamper + MaybeDecrypt |
| `backend/testframework/platform/database/mongo_disconnect_a8007_test.go` | 48 | 0 | чисто — A8-007 |
| `backend/testframework/platform/errors/errors_test.go` | 149 | 0 | чисто |
| `backend/testframework/platform/errors/errors_wave101_test.go` | 21 | 0 | чисто — RetryAfter min 1s |
| `backend/testframework/platform/health/probes_test.go` | 106 | 0 | FIXED-OK nil Deps→503 + parent context |
| `backend/testframework/platform/lifecycle/listen_a8008_test.go` | 31 | 0 | чисто — cancel + nil-safe |
| `backend/testframework/platform/telegram/markdown_test.go` | 30 | 0 | FIXED-OK спецсимволы + `\` |

Файлов в списке: **20**. Прочитано: **20**. Пропущено: **0**.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему дефект | Как чинить |
|----|-----|----------|-------------|----------|---------------|------------|
| A3W4-4-001 | NEW | P2 | `postgres.go:42–59` (vs `mongodb.go:14–20`, redis `Ping`) | `NewPostgresClient`: `sql.Open` + pool + optional `Schema.Create`, **без** `PingContext`; лог «connected to PostgreSQL» при lazy connect | Фокус DB connect fail-closed: Mongo/Redis отказывают на старте; PG в prod (`Schema.Create` off) стартует с мёртвым DSN → outbox/Ent error-storm до Ready 503. Misconfig URL не fail-start | После `applyPoolLimits`: `rawDB.PingContext(ctx)` → err return (+ Close drv); тест unreachable DSN → error, без ложного «connected» |
| A3W4-4-002 | NEW | P3 | `cron_test.go:15–66` (prod `cron.go:96–97`) | Слот-тест: tick / cancel / error-continues; **нет** кейса hang > budget | Мутация «убрать `WithTimeout` из `runTick`» оставляет assigned suite зелёным; сильный regression только в `timeout_task043_test.go` (вне слота) | Перенести/дублировать `TestRunTick_TimeoutCancelsHang` в `cron_test.go` или явно связать assignment со слот-файлом timeout |

## FIXED-OK (сводка)

| ID | Evidence |
|----|----------|
| A2W4-4-001 | TASK-112 AAD Seal/Open + mismatch test |
| W4-A4-001…012 | см. cascade — без регрессии на HEAD |

## Чисто (фокус)

- **Secrets / config defaults:** `devDefaults` JWT/S3/LiveKit; `!DevMode` reject + JWT≥32; `hub:hub@` data-stores; prod NATS userinfo; empty `METRICS_ALLOWED_CIDRS` fail-closed; `INTEGRATION_ENCRYPTION_KEY` + DeriveDevKey ban; `DEV_MODE`⊥`ENVIRONMENT=production`.
- **Crypto:** AES-256-GCM, 12B random nonce, tamper reject, nil cipher → Err (no panic); AAD entity bind; MaybeDecrypt soft — только migration (callers вне слота).
- **Probe auth:** `/live`/`/ready` без JWT — канон K8s/Compose; Ready fail-closed при 0 probes; детальный status — auth (wiring вне слота). Утечка degraded services через Ready — accepted ops surface.
- **Lifecycle:** `CancelOnListenError` → cancel, не `os.Exit` (A8-008).
- **Telegram markdown:** `\` первым; user-input escape path в notification/auth (потребители вне слота).
- **Mongo:** Connect+Ping + `MongoDisconnectCleanup` nil-safe.

### Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 1 |
| **Total** | **2** |

### Теги vs A2

| Tag | Count |
|-----|------:|
| FIXED-OK (A2 active → closed) | 1 (`A2W4-4-001`) |
| FIXED-OK (Stage A cascade) | 12 |
| NEW | 2 |
| REGRESS | 0 |
| RESIDUAL | 0 |

### Топ ID

1. **A3W4-4-001 (P2)** — Postgres connect без Ping (fail-open vs Mongo/Redis).
2. **A3W4-4-002 (P3)** — assigned `cron_test` не мутирует tick timeout.
