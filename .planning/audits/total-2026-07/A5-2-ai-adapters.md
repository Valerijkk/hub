# W5 агент A2 — ai adapters/stream

> TASK-008 · READ-ONLY · 2026-07-28  
> Фокус: SSRF `SafeHTTPClient` / `ValidateOutboundURL` на provider URLs, stream cancel, secret headers, content policy bypass.  
> Сверка: TZ SSRF (`http.DefaultClient` запрещён), prior `grok-2026-07` A4-005 (Ollama ↔ SafeDial), SEC-R8 / AI-SPEC content policy.

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/adapter_anthropic.go` | 151 | 1 | замечания |
| `backend/internal/ai/adapter_gigachat.go` | 158 | 1 | замечания |
| `backend/internal/ai/adapter_openai.go` | 82 | 1 | замечания |
| `backend/internal/ai/adapter_stream.go` | 96 | 1 | замечания |
| `backend/internal/ai/adapter_yandex.go` | 162 | 1 | замечания |
| `backend/internal/ai/content_policy.go` | 38 | 2 | замечания |
| `backend/internal/ai/fallback.go` | 171 | 3 | замечания |
| `backend/internal/ai/privacy.go` | 54 | 1 | замечания |
| `backend/internal/ai/provider_rpm.go` | 59 | 2 | замечания |
| `backend/internal/ai/stream.go` | 37 | 0 | чисто |
| `backend/testframework/ai/adapter_anthropic_test.go` | 44 | 0 | чисто |
| `backend/testframework/ai/adapter_http_test.go` | 329 | 0 | чисто |
| `backend/testframework/ai/adapter_openai_test.go` | 36 | 1 | замечания |
| `backend/testframework/ai/adapter_wave110_test.go` | 76 | 0 | чисто |
| `backend/testframework/ai/adapter_yandex_test.go` | 28 | 0 | чисто |
| `backend/testframework/ai/content_policy_test.go` | 52 | 1 | замечания |
| `backend/testframework/ai/fallback_test.go` | 40 | 1 | замечания |
| `backend/testframework/ai/privacy_test.go` | 68 | 1 | замечания |
| `backend/testframework/ai/provider_rpm_test.go` | 32 | 1 | замечания |
| `backend/testframework/ai/stream_test.go` | 26 | 1 | замечания |

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.

## Находки

| ID | Severity | Файл:строка | Проблема | Почему это дефект | Как чинить |
|----|----------|-------------|---------|-------------------|------------|
| W5-A2-001 | P1 | `content_policy.go:22-36` | **Content policy сканирует только `Role == "user"`.** `system` / `assistant` / пустой role / `User` пропускаются | Клиент шлёт jailbreak в `role:"system"` (или не-`user`) → SEC-R8 не срабатывает; адаптеры (Anthropic/Yandex) поднимают это в system prompt. `validateChatRequest` (`service.go`) зовёт только `CheckContentPolicy` — allowlist ролей нет | Сканировать **все** роли (или хотя бы system+user); allowlist `user\|assistant\|system`; отклонять прочие; тест на system/empty-role jailbreak |
| W5-A2-002 | P1 | `fallback.go:51-61`; `privacy.go:33-43` | **`isOfflineProvider`:** `ProviderTypeOllama` **или** substring `127.0.0.1`/`localhost`/`ollama` в BaseURL | `https://evil.example/ollama` / `…?x=localhost` проходит `filterProvidersByConsent(…, false)` как «offline» → **обход AllowExternalAI** при admin/BYOK URL с подстрокой. Тесты покрывают только явный cloud vs `ProviderTypeOllama` | Offline = type Ollama **и** host из allowlist (`127.0.0.1`/`localhost`/unix) после parse URL; запретить `Contains("ollama")` на публичном host; regression на `https://api.deepseek.com/v1/ollama` → filtered out |
| W5-A2-003 | P1 | `fallback.go:51-61`; `adapter_openai.go:32`; (SafeDial вне списка) | **A4-005 жив:** offline-путь помечает loopback как OK, `NewSafeHTTPClient` **блокирует** dial на `127.0.0.1`/`localhost` | `AI_OFFLINE_ONLY` / local Ollama: create URL проходит `ValidateOutboundURL`, runtime → `ErrBlockedAddress`. Consent/offline и SSRF-периметр противоречат | Trusted offline dialer/allowlist **или** sidecar с публичным DNS; не ослаблять SSRF для произвольного BYOK |
| W5-A2-004 | P2 | `adapter_stream.go:59-91` | SSE-цикл **не проверяет `ctx.Err()`** между чанками; cancel только через `NewRequestWithContext` → `Body.Read` | При буфере scanner / медленном `onDelta` отмена клиента (WS disconnect) может продолжать слать deltas и держать upstream до следующего Read/Timeout | В начале итерации: `if err := ctx.Err(); err != nil { return err }`; тест cancel mid-stream |
| W5-A2-005 | P2 | `fallback.go:139-166` | Stream fallback при retryable ошибке **после частичных `onDelta`** переключает провайдера без сброса/маркера | Клиент получает склейку ответов двух провайдеров (битый stream UX + возможные дубли) | Не fallback после первого успешного delta; или новый `stream_id`/событие `restart`; считать partial stream non-retryable |
| W5-A2-006 | P2 | `provider_rpm.go:47-49` | Redis `pipe.Exec` error → **`return nil` (fail-open)** | При недоступном Redis MaxRPM провайдера не enforced → обход AI-SPEC S3 / cost spike | Fail-closed (`return err` / `ErrProviderRateLimited`) или in-memory fallback с лимитом (как API RL); не silent allow |
| W5-A2-007 | P2 | `provider_rpm.go:41-51` | **ZAdd до проверки `ZCard > maxRPM`** (тот же паттерн, что W4-A1-004) | Отклонённые запросы раздувают окно → ложные 429 после burst | ZCARD→ZADD условно, или ZADD+ZREM при reject; Lua |
| W5-A2-008 | P2 | `adapter_openai.go:32`; `adapter_anthropic.go:25`; `adapter_gigachat.go:36`; `adapter_yandex.go:25`; `adapter_stream.go` (тот же client) | `NewSafeHTTPClient(90 * time.Second)` — **Timeout на весь request включая body** | Длинный SSE (>90s) обрывается клиентским deadline даже при живом upstream | Отдельный client для stream: без общего `Timeout` / только dial+header timeouts; cancel через ctx |
| W5-A2-009 | P3 | `content_policy.go:12-18` | Denylist — узкие EN-фразы; опечатки/`іgnore`/zero-width/`IGNORE…` с вставками обходят | SEC-R8 как единственный слой легко обходится; не дефект «забыли check», а слабая политика | Нормализация (NFKC, strip ZW*), расширить паттерны; не полагаться только на regex |
| W5-A2-010 | P3 | `adapter_gigachat.go:76` | `raw, _ := io.ReadAll(...)` — ошибка чтения OAuth body проглочена | При truncated/ошибке чтения возможен misleading `oauth parse` / пустой token path | Проверять `err` от `ReadAll` |
| W5-A2-011 | P3 | `privacy.go:27-29` | `pref == nil` → **`allowExternal = true`** (opt-out) | Новый user без записи preference сразу шлёт в cloud; privacy-by-default слабее | Default `false` или явная запись preference при регистрации; зафиксировать в SPEC |
| W5-A2-012 | P3 | `adapter_yandex.go:159-161` | `IsYandexProvider`: `Contains("yandex")` на любом BaseURL | Ложное срабатывание на URL с подстрокой → маршрутизация на hard-coded Yandex API, игнор BaseURL | Match host `llm.api.cloud.yandex.net` (и документированные aliases), не substring |
| W5-A2-013 | P3 | `content_policy_test.go:19-29` (нет кейсов) | Нет тестов: jailbreak в `system` / empty role / `User` | Мутация «сканировать только user» / W5-A2-001 **не краснеет** | Table-driven bypass cases → `ErrContentPolicyViolation` или 400 на invalid role |
| W5-A2-014 | P3 | `stream_test.go:10-25`; `adapter_http_test.go:29-58` | Нет теста cancel mid-SSE; `stream_test` только nil-safe publisher | Регресс W5-A2-004 не ловится | `ChatCompletionStream` + cancel ctx + assert `context.Canceled` / stop deltas |
| W5-A2-015 | P3 | `provider_rpm_test.go:12-22` | Dead Redis тестируется с `maxRPM==0` (short-circuit); **нет** miniredis enforce / fail-open при `maxRPM>0` | W5-A2-006/007 не покрыты; тест «пройдёт на сломанном лимитере» | miniredis: N+1 → `ErrProviderRateLimited`; unreachable Redis + maxRPM>0 → ожидаемый fail-closed/open по политике |
| W5-A2-016 | P3 | `fallback_test.go:22-34`; `privacy_test.go:14-28` | Нет кейса cloud URL с `/ollama` или `ProviderTypeOllama`+публичный host | Мутация/регресс W5-A2-002 не ловится | Assert filtered out при `allowExternal=false` |
| W5-A2-017 | P3 | `adapter_openai_test.go:21-28` | SSRF-тест только `://bad`; нет loopback/`file:`/`169.254.169.254` | Регресс SafeDial+Validate на adapter path тонкий | `http://127.0.0.1:9/v1` → error (Validate или dial); `ErrorIs` где применимо |

## Чисто

| Файл / зона | Что проверено |
|-------------|----------------|
| SSRF baseline (все adapters) | `NewSafeHTTPClient(90s)` + `ValidateOutboundURL(base)` перед Do; Anthropic default URL тоже Validate; Yandex — hard-coded `yandexCompletionURL` + Validate; GigaChat OAuth URL Validate; body `io.LimitReader` ≤ `maxResponseBytes` (4 MiB) / stream LimitReader; **нет** `http.DefaultClient` в списке |
| Secret headers | API key только в `Authorization` / `x-api-key` / BasicAuth; в slog этих файлов ключи не пишутся; HTTP error body truncate ≤512 без echo headers; Giga token в `sync.Map` (не лог) |
| `adapter_openai.go` non-stream | `req.Stream==true` → явный отказ («use stream_id…»); empty apiKey допускается (Ollama) — осознанно |
| `adapter_anthropic.go` / Giga / Yandex stream | Fake-stream через полный completion + ctx на HTTP; `onDelta` error прерывает |
| `stream.go` | Nil-safe publisher; payload копирует `stream_id`; события `ai.stream.*` через `SendToUser` — без утечки ключей |
| `fallback.go` retry taxonomy | 401/content-policy non-retryable; 429/5xx/timeout retryable; RPM skip provider |
| `adapter_*_test` / `adapter_http_test` | Happy-path + HTTP error + SSE parse + header asserts (`Bearer`/`x-api-key`/`Api-Key`); credential parse Giga/Yandex |
| `adapter_anthropic_test` | split messages, empty API key reject |
| `adapter_yandex_test` | credentials + IsYandexProvider basics |
| `adapter_wave110_test` | OpenAI success/429/bad JSON |

**Периметр SSRF (кроме W5-A2-003):** user-controlled BaseURL на chat path закрыт Validate+SafeDial; IPv4-mapped обход — в W4-A1-001 (safedialer), здесь не дублировать как новый root-cause.

**Secret headers:** в зоне A2 утечек ключа в логи/ошибки не найдено.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **20 / 20 / 0** |
| Находок | **17** |
| P0 | 0 |
| P1 | 3 |
| P2 | 5 |
| P3 | 9 |

### Топ-3

1. **W5-A2-001 (P1)** — content policy только на `role=user` → jailbreak через `system`/не-user role.  
2. **W5-A2-002 (P1)** — substring/`ProviderTypeOllama` в `isOfflineProvider` обходит `AllowExternalAI`.  
3. **W5-A2-003 (P1)** — offline Ollama (localhost) vs SafeDial: local AI фактически неработоспособен (A4-005).
