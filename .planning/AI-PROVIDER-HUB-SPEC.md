# AI Provider Hub — Technical Specification

> **Module:** `hub-ai-gateway` · **Milestones:** v3.0 (MVP) → v3.5 (OpenRouter) → v4.0 (enterprise)  
> **Goal:** Любая модель из top providers через единый UI и OpenAI-compatible API **без VPN** (where possible).

---

## 1. Problem

- Teams want ChatGPT, Claude, DeepSeek, GigaChat in one place
- MILESTONE-3 lists RU providers — extend to **pluggable registry**
- Keys must not leak to frontend; audit required; rate limits mandatory

---

## 2. Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────────────┐
│ React SPA   │────►│  hub-api (Fiber) │────►│  internal/ai/           │
│ ModelPicker │     │  /api/ai/*       │     │  GatewayService         │
│ AI Panel    │     │  RBAC + rate     │     │  ├─ ProviderRegistry    │
└─────────────┘     └──────────────────┘     │  ├─ OpenAIAdapter       │
                                              │  ├─ AnthropicAdapter    │
                                              │  ├─ GigaChatAdapter     │
                                              │  └─ OllamaAdapter       │
                                              └───────────┬─────────────┘
                                                          │ TLS + SSRF guard
                                              ┌───────────▼─────────────┐
                                              │ External provider APIs   │
                                              └─────────────────────────┘
```

**v3.5+ optional split:** `cmd/ai-gateway/` separate process, same repo, shares `internal/ai/`.

---

## 3. Provider registry

```go
type ProviderType string
const (
    ProviderOpenAICompatible ProviderType = "openai_compatible" // DeepSeek, OpenRouter, local
    ProviderGigaChat         ProviderType = "gigachat"
    ProviderYandexGPT        ProviderType = "yandexgpt"
    ProviderAnthropic        ProviderType = "anthropic"
    ProviderOllama           ProviderType = "ollama"
)

type ProviderConfig struct {
    ID          uuid.UUID
    Name        string        // "DeepSeek", "OpenRouter"
    Type        ProviderType
    BaseURL     string        // validated via ValidateOutboundURL
    APIKeyEnc   []byte        // AES-256 encrypted
    Models      []string      // allowlist
    Enabled     bool
    Priority    int           // fallback order
    MaxRPM      int           // per instance
}
```

### Pre-built presets (admin wizard)

| Preset | Base URL | VPN needed | Notes |
|--------|----------|------------|-------|
| DeepSeek | `https://api.deepseek.com/v1` | Usually no | OpenAI format |
| OpenRouter | `https://openrouter.ai/api/v1` | Sometimes | 100+ models |
| GigaChat | Sber OAuth | No | RU |
| YandexGPT | Yandex Cloud | No | RU |
| Ollama | `http://ollama:11434/v1` | No | Self-host |
| OpenAI | `https://api.openai.com/v1` | Often yes | Admin discretion |
| Anthropic | `https://api.anthropic.com` | Often yes | Messages API adapter |

---

## 4. API surface

### 4.1 User chat (messenger integration)

- Existing M3-009..014: `@hub-ai` bot, streaming WS
- **Enhancement:** `model` field in request; user default from profile

### 4.2 OpenAI-compatible proxy (developers)

```
POST /api/ai/v1/chat/completions
Authorization: Bearer <hub-jwt>
Content-Type: application/json

{
  "model": "deepseek-chat",
  "messages": [...],
  "stream": true
}
```

Hub resolves `model` → provider via registry; injects provider key server-side.

### 4.3 Admin

```
GET/POST   /api/admin/ai/providers
PATCH      /api/admin/ai/providers/:id
POST       /api/admin/ai/providers/:id/test
GET        /api/admin/ai/usage?from=&to=
```

---

## 5. Frontend UX

| Component | Path | Description |
|-----------|------|-------------|
| `AIProviderSettings` | `/admin/settings/ai` | Admin CRUD providers |
| `ModelPicker` | AI Panel, composer | Dropdown grouped by provider |
| `AIChatPanel` | Ctrl+Shift+A | Side panel, model + history |
| `ai-store.ts` | zustand | Selected model, usage quota |

**i18n:** `ai.json` — provider names, quota messages, errors

---

## 6. Security checklist

| # | Requirement | Implementation |
|---|-------------|----------------|
| S1 | No API keys in FE | Admin API only |
| S2 | SSRF | `ValidateOutboundURL` on BaseURL |
| S3 | Rate limit | Redis: user/day + RPM |
| S4 | Audit | `audit_log`: provider_id, model, tokens, latency — **no prompt/response body** |
| S5 | RBAC | `PermAIUse`, `PermAIManageProviders` |
| S6 | Secret compare | `subtle.ConstantTimeCompare` for webhook shared secrets |
| S7 | Content policy | Optional admin blocklist regex (v3.5) |
| S8 | Air-gap | Ollama-only mode flag `AI_OFFLINE_ONLY=true` |

---

## 7. Ent schemas (new)

- `ai_provider` — provider configs
- `ai_usage_daily` — user_id, date, tokens_in, tokens_out, request_count
- `ai_user_preference` — default_model_id

---

## 8. Phased delivery

| Phase | Scope | Effort |
|-------|-------|--------|
| **AI-1** | Registry + DeepSeek + Ollama + OpenAI adapter | L |
| **AI-2** | GigaChat + YandexGPT | M |
| **AI-3** | @hub-ai streaming + ModelPicker | L |
| **AI-4** | OpenRouter preset + 100 models | M |
| **AI-5** | `/api/ai/v1/chat/completions` proxy | M |
| **AI-6** | Anthropic native adapter | S |
| **AI-7** | Usage dashboards + quotas | M |

---

## 9. Acceptance (v3.0 MVP)

1. Admin adds DeepSeek key → test ping OK
2. User selects DeepSeek-V3 in AI Panel → streaming response
3. User switches to Ollama llama3 → works on LAN
4. Disabled provider returns 503 with i18n message
5. Audit row created without message content
6. Rate limit 429 with Retry-After

---

*Parent: [PLATFORM-VISION-v4.md](./PLATFORM-VISION-v4.md) · MILESTONE-3 M3-001..030*
