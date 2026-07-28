# ADR-AI-OLLAMA-LOOPBACK-2026 — Ollama localhost consent vs SafeDial

- Статус: **Accepted** (owner 2026-07-28: отдельный approve ADR не требуется — TASK-101)
- Дата: 2026-07-28
- Автор: be1 · Finding: P1-BE-U01 / A2W5-A2-005 / W5-A2-003
- Связанные: TASK-101, `backend/internal/platform/http/safedialer.go`, `backend/internal/ai/adapter_*.go`

## Контекст

Consent-слой AI (`isOfflineBaseURL` / `filterProvidersByConsent`) считает
`http://127.0.0.1` / `localhost` / `::1` допустимыми **без** `AllowExternalAI`
(локальный LLM / Ollama). При этом `ValidateOutboundURL` + `SafeDialContext` /
`NewSafeHTTPClient` блокируют loopback как SSRF — offline-провайдер проходит
consent и падает на dial. Противоречие P1-BE-U01.

## Рассмотренные варианты

1. **Trusted offline dialer (выбрано)** — отдельный client/dialer только для
   loopback allowlist; глобальный SafeDial не ослаблять.
2. Sidecar / reverse-proxy на публичный bind внутри сети — отклонено: лишняя
   infra для self-hosted Ollama на том же хосте.
3. Запрет loopback BYOK / только remote Ollama — отклонено: ломает AI-SPEC
   offline / `AI_OFFLINE_ONLY` и consent без external.

## Решение

1. **`ValidateOutboundURL` / `SafeDialContext` / `NewSafeHTTPClient` без изменений
   семантики** — loopback, RFC1918, link-local, metadata остаются blocked для
   webhook / OpenGraph / Git* / cloud BYOK / любых user-controlled URL.
2. **Явный consent на dial** = BaseURL проходит `ValidateOfflineAIURL`:
   схема `http`/`https`, host ∈ `{localhost, 127.0.0.1, ::1}`, без userinfo.
   Это согласовано с `ai.isOfflineBaseURL` (HTTP-ветка).
3. **`SafeDialContextAllowLoopback` / `NewOfflineAIHTTPClient`** — после resolve
   **все** IP должны быть loopback; иначе `ErrBlockedAddress`. Не разрешает
   RFC1918 / `169.254.169.254` / публичные IP.
4. **OpenAI-compatible AI adapter** выбирает offline-client только если
   `ValidateOfflineAIURL` OK; иначе обычный SafeDial path.
5. Cloudflare / DeepSeek / любой non-loopback URL **никогда** не получает
   AllowLoopback dialer.

## Consequences

### Positive

- Ollama на `127.0.0.1:11434` работает при offline consent.
- SSRF-инвариант для остального периметра сохранён (мутационные тесты
  `ValidateOutboundURL` / `SafeDialContext_BlocksLoopback` остаются зелёными).

### Negative

- Два HTTP-клиента в AI adapter; нужно не перепутать path.
- DNS rebind `localhost` → non-loopback по-прежнему режется AllowLoopback dialer.

## Finding IDs (TASK-101)

| ID | Вердикт |
|----|---------|
| P1-BE-U01 | FIXED — ADR + offline dialer |
| A2W5-A2-005 | FIXED — evidence: tests + код |
| W5-A2-003 | FIXED — то же |
