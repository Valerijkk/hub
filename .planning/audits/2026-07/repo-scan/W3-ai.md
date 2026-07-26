# W3-ai — `backend/internal/ai`

> **Provenance class:** FALLBACK_SUMMARY · **Confidence:** PRELIMINARY · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Роль:** security-auditor + be-ver (readonly) · **Дата:** 2026-07-23  
> **Scope:** adapters, repository, settings_store, bot_user · **Код не менялся**

## Summary

Исходящие LLM/BYOK URL проходят **`safehttp.ValidateOutboundURL`** + `NewSafeHTTPClient` — **SSRF PASS**. Layering: `EntRepository` с `*ent.Client` — ок; **`EntSettingsStore` и `EnsureBotUser` держат/принимают `*ent.Client` вне Repository-интерфейса** — note (нарушение духа TZ layering, не HTTP-bypass).

**Вердикт:** `verified: false` из-за layering NOTE (MEDIUM) · CRITICAL/HIGH security = 0.

## Findings

| Severity | file:line | Issue | Fix |
|----------|-----------|-------|-----|
| **PASS** | `adapter_openai.go:32,41` · anthropic/gigachat/yandex/stream | SafeHTTPClient + ValidateOutboundURL на base/URL | Держать на каждом adapter |
| **PASS** | `repository_ent.go:64,101` | ValidateOutboundURL на user BaseURL при save provider | — |
| **MEDIUM** | `settings_store.go:21-26` | `EntSettingsStore.db *ent.Client` — store вне Repository, прямой Ent в AI-пакете | Вынести в `EntRepository` / SystemSetting repo |
| **MEDIUM** | `bot_user.go:20-40` | `EnsureBotUser(ctx, *ent.Client)` — bootstrap с прямым Ent | Wiring → repo method; не экспортировать client в ai API |
| **INFO** | `repository_ent.go:22-27` | `*ent.Client` в EntRepository — канон | — |
| **LOW** | adapters timeout 90s | Длинный timeout на LLM — DoS resource при злоупотреблении admin BYOK | Rate-limit admin AI + max concurrent |

## Checklist

| Инвариант | Статус |
|-----------|--------|
| SSRF ValidateOutboundURL | **PASS** |
| SafeHTTPClient (не DefaultClient) | **PASS** |
| Layering без `*ent.Client` в service/handler | **PARTIAL** (settings_store / bot_user) |
| Secrets в adapters | PASS (из settings/cipher) |
