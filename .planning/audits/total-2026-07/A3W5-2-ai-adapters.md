# A3 W5 агент 2 — ai adapters/stream (FINAL)

> TASK-008 · Stage A3 FINAL · READ-ONLY · 2026-07-29  
> HEAD: `ee353287fc7bdb3ba4ab294adb20753893ac9558` (`ee353287`) · branch `cursor-grok-4.5`  
> Baseline A2: [A2W5-2-ai-adapters.md](A2W5-2-ai-adapters.md)  
> Stage A: [A5-2-ai-adapters.md](A5-2-ai-adapters.md)  
> Периметр: 20 файлов (W5-ASSIGNMENTS A2) · Прочитано: 20 · Пропущено: 0  
> Фокус: SSRF SafeDialer на user URLs · streamClient cancel/leak · OfflineAI · sanitize · secrets in logs

## Сверка с A2 (A2W5-A2-*)

| A2 ID | Tag (A2) | Статус в HEAD | Доказательство |
|-------|----------|---------------|----------------|
| A2W5-A2-001 | REGRESS → FIXED-OK | закрыт | `provider_rpm.go:52-54` Redis err → `ErrProviderRPMUnavailable` (не `return nil`); `task085_regress_test.go` / `task113_policy_rpm_test.go` (вне слота) |
| A2W5-A2-002 | REGRESS → FIXED-OK | закрыт | `provider_rpm.go:56-59` ZRem member при reject; ZSET card не раздувается (`task113`) |
| A2W5-A2-003 | REGRESS → **частично** | cloud FIXED / offline RESIDUAL | cloud SSE: `streamClient` = `newSafeStreamHTTPClient()` без `Client.Timeout` (`adapter_openai.go:43-44`, `http_stream_client.go` вне списка); **offline** SSE всё ещё `NewOfflineAIHTTPClient(90s)` → A3W5-2-001 |
| A2W5-A2-004 | REGRESS → FIXED-OK | закрыт | `fallback.go:176-198` `emittedContent` → no provider switch; `task085_regress_test.go` |
| A2W5-A2-005 | FIXED → FIXED-OK | закрыт | `ValidateOfflineAIURL` → `offlineClient` (`adapter_openai.go:49-63`); `ollama_loopback_task101_test.go` |
| A2W5-A2-006 | FIXED → FIXED-OK | закрыт | NFKC+ZW+RU denylist (`content_policy.go:15-66`); `TestCheckContentPolicy_ZWAndNFKCBypass_TASK113` |
| A2W5-A2-007 | FIXED → FIXED-OK | закрыт | = 001+002 (fail-closed + no inflate) |
| A2W5-A2-008 | FIXED → FIXED-OK | закрыт | link-local → invalid URL; loopback → OfflineAI path (`task113`) |
| W5-A2-001…017 (Stage A via A2) | FIXED-OK / см. residual | без регресса P1 | all-roles policy; host allowlist offline; ctx cancel mid-stream; Giga ReadAll err; pref nil→opt-in false; Yandex host map; stream publisher nil-safe |

**REGRESS:** 0 · **RESIDUAL (активные):** 2 · **NEW:** 1

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/adapter_anthropic.go` | 151 | 0 | чисто (ValidateOutboundURL + SafeHTTPClient; fake-stream) |
| `backend/internal/ai/adapter_gigachat.go` | 162 | 0 | чисто (OAuth+base Validate; ReadAll err; token в sync.Map) |
| `backend/internal/ai/adapter_openai.go` | 133 | 1 | замечания (offline stream Timeout → A3W5-2-001) |
| `backend/internal/ai/adapter_stream.go` | 99 | 1 | замечания (swallow ReadAll err → A3W5-2-002); cancel FIXED-OK |
| `backend/internal/ai/adapter_yandex.go` | 180 | 0 | чисто (hard-coded URL Validate; host allowlist) |
| `backend/internal/ai/content_policy.go` | 67 | 1 | замечания (слабый denylist residual → A3W5-2-003) |
| `backend/internal/ai/fallback.go` | 208 | 0 | чисто (offline host allowlist; no switch after delta) |
| `backend/internal/ai/privacy.go` | 55 | 0 | чисто (`pref==nil` → allowExternal=false) |
| `backend/internal/ai/provider_rpm.go` | 67 | 0 | чисто (fail-closed + ZRem) |
| `backend/internal/ai/stream.go` | 37 | 0 | чисто (nil-safe; без ключей в payload) |
| `backend/testframework/ai/adapter_anthropic_test.go` | 44 | 0 | чисто |
| `backend/testframework/ai/adapter_http_test.go` | 329 | 0 | чисто (SSE/headers/Giga/Yandex) |
| `backend/testframework/ai/adapter_openai_test.go` | 36 | 0 | чисто (`://bad`; stream reject; truncate) |
| `backend/testframework/ai/adapter_wave110_test.go` | 76 | 0 | чисто |
| `backend/testframework/ai/adapter_yandex_test.go` | 32 | 0 | чисто (host allowlist negatives) |
| `backend/testframework/ai/content_policy_test.go` | 92 | 0 | чисто (system/assistant/empty/tool roles) |
| `backend/testframework/ai/fallback_test.go` | 70 | 0 | чисто (`/ollama` public host blocked) |
| `backend/testframework/ai/privacy_test.go` | 68 | 0 | чисто |
| `backend/testframework/ai/provider_rpm_test.go` | 32 | 0 | чисто на nil/maxRPM=0; mutation fail-closed **вне** слота |
| `backend/testframework/ai/stream_test.go` | 26 | 0 | чисто (nil publisher); cancel-тест вне слота |

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.

## Находки (активные A3)

| ID | Tag | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|-----|----------|-------------|---------|-------------------|------------|
| A3W5-2-001 | RESIDUAL | P2 | `adapter_openai.go:44,66-82` (+ `NewOfflineAIHTTPClient(90s)`) | Offline/Ollama SSE **не** получает streamClient без `Client.Timeout`: `resolveStreamHTTPClient` отдаёт `streamClient` только если `c == a.client`; при Offline возвращает `offlineClient` с **Timeout=90s** | Cloud stream живёт до service `WithTimeout(5m)` (`service.go` вне слота); локальный Ollama SSE >90s рвётся клиентским deadline — частичный fix W5-A2-008 / TASK-085 | `newOfflineAIStreamHTTPClient()` = AllowLoopback dial + TLS/header timeouts **без** `Client.Timeout`; в `resolveStreamHTTPClient` при `c == a.offlineClient` отдавать его; тест stream >90s offline + cancel |
| A3W5-2-002 | NEW | P3 | `adapter_stream.go:52-54` | `raw, _ := io.ReadAll(LimitReader(...))` — ошибка чтения error-body проглочена | Тот же класс, что W5-A2-010 (GigaChat уже чинили): truncated/read err → misleading HTTP status message | `if err != nil { return fmt.Errorf("read error body: %w", err) }` |
| A3W5-2-003 | RESIDUAL | P3 | `content_policy.go:15-25` | Denylist по-прежнему узкий EN/RU regex; homoglyph/вставки вне NFKC+ZW обходят SEC-R8 | W5-A2-009 не закрыт целиком: NFKC/ZW/RU (TASK-113) закрыли обход нормализацией, но политика остаётся weak denylist как единственный слой | Расширить паттерны / второй слой; не полагаться только на regex; mutation на известные bypass |

## FIXED-OK (сверка A2 + Stage A фокус)

| Зона | Проверено в HEAD |
|------|------------------|
| SSRF SafeDial (user BaseURL) | OpenAI: `ValidateOutboundURL` **или** `ValidateOfflineAIURL`+offline dialer; Anthropic/Giga/Yandex: Validate + `NewSafeHTTPClient`; **нет** `http.DefaultClient`; metadata/`169.254` → invalid URL |
| OfflineAI (ADR-AI-OLLAMA) | Host allowlist `localhost`/`127.0.0.1`/`::1` (+unix в consent); **не** substring `ollama` / `ProviderTypeOllama` alone; SafeDial глобально не ослаблен |
| streamClient cancel | `adapter_stream.go:59-62` `ctx.Err()` в цикле; `defer Body.Close`; cloud stream без whole-request Timeout; service stream ≤5m (вне слота) — cancel/leak cloud FIXED-OK |
| Stream fallback | После первого content-delta — no provider switch |
| RPM | Fail-closed + ZRem; nil/maxRPM≤0 short-circuit |
| Sanitize | Error bodies `truncate` ≤512; LimitReader 4MiB / stream; Giga OAuth ReadAll err checked |
| Secrets in logs | API key только в headers (`Authorization` / `x-api-key` / BasicAuth / Api-Key); slog в `fallback.go` — provider name + err (не ключ); `stream.go` payload без credentials; Giga token в `sync.Map` |
| Content policy roles | Allowlist `user\|assistant\|system`; invalid/empty role → violation |
| Privacy default | `pref==nil` → `allowExternal=false` (opt-in) |
| Yandex routing | Host map `llm.api.cloud.yandex.net`, не substring |

## Сводка counts (активные)

| Severity | Count |
|----------|------:|
| P0 | 0 |
| P1 | 0 |
| P2 | 1 |
| P3 | 2 |
| **Total** | **3** |

| Метрика | Значение |
|---------|----------|
| Файлов | **20 / 20 / 0** |
| Активных | **3** (NEW 1 · REGRESS 0 · RESIDUAL 2) |
| FIXED-OK (A2 actives 001–008) | **7/8** полных; **003** partial → A3W5-2-001 |
| FIXED-OK (Stage A P1 via A2) | **все** без регресса |

### Топ-3

1. **A3W5-2-001 (P2, RESIDUAL)** — OfflineAI SSE всё ещё под `Client.Timeout=90s`, пока cloud stream уже без Timeout.
2. **A3W5-2-003 (P3, RESIDUAL)** — content-policy denylist остаётся слабым слоем после NFKC/ZW.
3. **A3W5-2-002 (P3, NEW)** — swallow `ReadAll` err на SSE HTTP-error path.

---

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.  
Counts: P0=0 · P1=0 · P2=1 · P3=2 · Total=3.  
Топ: A3W5-2-001 (offline stream 90s).
