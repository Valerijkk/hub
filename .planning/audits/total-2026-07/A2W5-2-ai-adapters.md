# A2-W5 агент A2 — ai adapters/stream (re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> HEAD: `0d6a659fe9dd56a473cea0ace40ccae280970cbc`  
> Зеркало: `W5-ASSIGNMENTS.md` §A2 · Stage A: `A5-2-ai-adapters.md`

## Прочитанные файлы

| Файл | Строк | Находок | Вердикт |
|------|------:|--------:|---------|
| `backend/internal/ai/adapter_anthropic.go` | 151 | 1 | замечания |
| `backend/internal/ai/adapter_gigachat.go` | 162 | 0 | чисто |
| `backend/internal/ai/adapter_openai.go` | 87 | 1 | замечания |
| `backend/internal/ai/adapter_stream.go` | 102 | 0 | чисто (FIXED-OK cancel) |
| `backend/internal/ai/adapter_yandex.go` | 180 | 0 | чисто |
| `backend/internal/ai/content_policy.go` | 43 | 1 | замечания (weak denylist) |
| `backend/internal/ai/fallback.go` | 193 | 1 | замечания |
| `backend/internal/ai/privacy.go` | 55 | 0 | чисто |
| `backend/internal/ai/provider_rpm.go` | 59 | 2 | замечания |
| `backend/internal/ai/stream.go` | 37 | 0 | чисто |
| `backend/testframework/ai/adapter_anthropic_test.go` | 44 | 0 | чисто |
| `backend/testframework/ai/adapter_http_test.go` | 329 | 0 | чисто |
| `backend/testframework/ai/adapter_openai_test.go` | 36 | 1 | замечания |
| `backend/testframework/ai/adapter_wave110_test.go` | 76 | 0 | чисто |
| `backend/testframework/ai/adapter_yandex_test.go` | 32 | 0 | чисто |
| `backend/testframework/ai/content_policy_test.go` | 92 | 0 | чисто |
| `backend/testframework/ai/fallback_test.go` | 70 | 0 | чисто |
| `backend/testframework/ai/privacy_test.go` | 68 | 0 | чисто |
| `backend/testframework/ai/provider_rpm_test.go` | 32 | 1 | замечания |
| `backend/testframework/ai/stream_test.go` | 26 | 0 | чисто |

Файлов в списке: 20. Прочитано: 20. Пропущено: 0.

## Находки

| ID | Sev | Tag | Файл:строка | Проблема | A1 / задача | Мутация |
|----|-----|-----|-------------|----------|-------------|---------|
| A2W5-A2-001 | P2 | REGRESS | `provider_rpm.go:47-48` | Redis `pipe.Exec` err → **`return nil` (fail-open)** — MaxRPM не enforced | W5-A2-006 · P2-BE-10 (ai не в files_allowed TASK-040) | Unreachable Redis + maxRPM>0 → `ErrProviderRateLimited` |
| A2W5-A2-002 | P2 | REGRESS | `provider_rpm.go:44-51` | **ZAdd до** проверки `ZCard > maxRPM` — rejected раздувают окно | W5-A2-007 · P2-BE-10 | Rejected не увеличивает ZSET card |
| A2W5-A2-003 | P2 | REGRESS | `adapter_openai.go:35` (+ anthropic/giga/yandex) | `NewSafeHTTPClient(90s)` — **Timeout на весь SSE** | W5-A2-008 · P2-BE-17 | Stream >90s не рвётся клиентским Timeout |
| A2W5-A2-004 | P2 | REGRESS | `fallback.go:161-188` | Stream fallback при retryable err **без учёта** уже отданных deltas | W5-A2-005 · P2-BE-17 | После 1-го delta — no provider switch / restart marker |
| A2W5-A2-005 | P1 | FIXED | `fallback.go` + SafeDial + ADR-AI-OLLAMA-LOOPBACK | Offline loopback dial через `ValidateOfflineAIURL` / `NewOfflineAIHTTPClient`; `ValidateOutboundURL`/`SafeDial` без изменений | W5-A2-003 | TASK-101: loopback dial OK; metadata/RFC1918/global SafeDial → ErrBlocked |
| A2W5-A2-006 | P3 | FIXED | `content_policy.go` NFKC+ZW+RU | denylist ╨┐╨╛╤Б╨╗╨╡ ╨╜╨╛╤А╨╝╨░╨╗╨╕╨╖╨░╤Ж╨╕╨╕ | TASK-113 | `TestCheckContentPolicy_ZWAndNFKCBypass_TASK113` |
| A2W5-A2-007 | P3 | FIXED | `provider_rpm.go` fail-closed + ZRem | dead Redis тЖТ ErrProviderRPMUnavailable; no ZSET inflate | TASK-113 | `TestRedisProviderRPMLimiter_FailClosedAndNoZSETInflate_TASK113` |
| A2W5-A2-008 | P3 | FIXED | OpenAI adapter SSRF | loopback / link-local тЖТ error | TASK-113 | `TestOpenAIAdapter_ChatCompletion_LoopbackAndLinkLocal_TASK113` |

## FIXED-OK

| A1 ID | Evidence |
|-------|----------|
| W5-A2-001 | `CheckContentPolicy` сканирует user\|assistant\|system; invalid role → violation |
| W5-A2-002 | `isOfflineBaseURL` — host allowlist, не substring `ollama` |
| W5-A2-004 | `adapter_stream.go:61-63` `ctx.Err()` в цикле + `stream_cancel_task044_test.go` |
| W5-A2-010 | GigaChat OAuth `ReadAll` проверяет `err` |
| W5-A2-011 | `pref==nil` → `allowExternal=false` |
| W5-A2-012 | `IsYandexProvider` по host map |
| W5-A2-013/016 | content_policy + fallback tests на bypass |

## Чисто

SSRF baseline: `ValidateOutboundURL` + `NewSafeHTTPClient` на adapters; нет `http.DefaultClient`. Secret headers не логируются. `stream.go` nil-safe publisher.

## Сводка

| Метрика | Значение |
|---------|----------|
| Файлов | **20 / 20 / 0** |
| Находок | **8** (NEW 0 · REGRESS 4 · RESIDUAL 4) |
| P0 | 0 |
| P1 | 1 |
| P2 | 4 |
| P3 | 3 |
