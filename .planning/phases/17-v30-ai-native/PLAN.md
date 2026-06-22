# Phase 17 — AI Native (v3.0)

> **Branch:** `feat/v3.0-ai` · **PO LOCK:** `DISCUSS-ROADMAP-LOCKED.md` · **Spec:** `AI-PROVIDER-HUB-SPEC.md`

## Exit criteria

AI gateway (Ollama + BYOK) · admin provider registry · chat completions proxy · @hub-ai bot · ModelPicker · JQL automation

## Verification loop

```
be/fe → unit tests → qa-gate slice → *-ver → fix → re-verify → commit
```

---

## Wave 1 — AI-1 Provider registry + gateway MVP

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W1-01 Ent `ai_provider`, `ai_usage_daily`, `ai_user_preference` | be | arch-ver | ✓ |
| W1-02 Provider interface + OpenAI-compatible adapter | be | be-ver | ✓ |
| W1-03 Gateway service + admin CRUD + test ping | be | be-ver | ✓ |
| W1-04 `POST /api/ai/v1/chat/completions` + audit metadata | be | security-auditor | ✓ |
| W1-05 FE admin AI providers + `GET /api/ai/models` | fe | fe-ver | ✓ |
| W1-06 Unit tests + rate limit | qa | qa-ver | ✓ |

## Wave 2 — AI-03 @hub-ai messenger bot

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W2-01 System bot user + mention handler | be | be-ver | ✓ |
| W2-02 Thread context (last 20 msgs) + reply | be | qa-ver | ✓ |
| W2-03 FE markdown + code highlight in bot messages | fe | fe-ver | ✓ |
| W2-04 M3-010 Streaming bot replies (typing + incremental edit) | be | qa-ver | ✓ |

## Wave 3 — AI Panel + ModelPicker

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W3-01 `ai-store.ts` + ModelPicker component | fe | fe-ver | ✓ |
| W3-02 AI side panel Ctrl+Shift+A | fe | fe-ver | ✓ |
| W3-03 Streaming WS responses | be/fe | qa-ver | ✓ |

## Wave 4 — BYOK + RU providers

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W4-01 User-scoped providers (BYOK) | be | security-auditor | ✓ |
| W4-02 GigaChat adapter stub + ru_resident label | be | fe-ver | ✓ |

## Wave 5 — JQL-01 Full automation

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W5-01 Natural language → JQL assistant endpoint | be/fe | sa-ver | ✓ |

## Wave 6 — Bot commands + inline AI (M3-015..016, M3-023, M3-025)

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W6-01 `/ai summarize\|test-cases\|explain\|translate` | be | qa-ver | ✓ |
| W6-02 `GET /api/ai/usage/today` + panel indicator | be/fe | fe-ver | ✓ |
| W6-03 Inline «Спросить AI» on issue description | fe | fe-ver | ✓ |

## Wave 7 — v3.5 hardening (preferences, quota, GigaChat, OpenRouter)

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W7-01 User default model preference API + ModelPicker persist | be/fe | fe-ver | ✓ |
| W7-02 Daily quota 429 + content policy filter (SEC-R8) | be | security-auditor | ✓ |
| W7-03 GigaChat OAuth adapter (client_id:client_secret) | be | be-ver | ✓ |
| W7-04 OpenRouter preset + admin usage summary | fe/be | qa-ver | ✓ |

## Wave 8 — M3-017/020/021 + Anthropic (v3.5)

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W8-01 `/ai review` + GitHub PR fetch | be | be-ver | ✓ |
| W8-02 `/ai docs` RAG-lite via global search | be | sa-ver | ✓ |
| W8-03 `/ai sql` + `/ai help` commands | be | qa-ver | ✓ |
| W8-04 Anthropic native adapter + admin preset | be/fe | security-auditor | ✓ |

## Wave 9 — Multi-VCS review + YandexGPT + RPM + usage history (v3.5)

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W9-01 Unified `/ai review` GitHub PR + GitLab MR | be | be-ver | ✓ |
| W9-02 Per-provider MaxRPM via Redis | be | security-auditor | ✓ |
| W9-03 YandexGPT adapter + admin preset | be/fe | fe-ver | ✓ |
| W9-04 Admin usage history `GET /admin/ai/usage/history` | be/fe | qa-ver | ✓ |

## Wave 10 — Fallback chain + admin prompt + air-gap (M3-008, M3-014, S8)

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W10-01 Provider fallback chain on 5xx/timeout (M3-008) | be | be-ver | ✓ |
| W10-02 Admin bot system prompt API + UI (M3-014) | be/fe | fe-ver | ✓ |
| W10-03 `AI_OFFLINE_ONLY` Ollama-only routing (S8) | be | security-auditor | ✓ |
| W10-04 Mistral preset + fallback tests | qa | qa-ver | ✓ |

## Wave 11 — Quality gate + docs (Phase 17 exit)

| Task | Agent | Verifier | Done |
|------|-------|----------|------|
| W11-01 Vitest: `ai-store`, `ModelPicker` | fe | fe-ver | ✓ |
| W11-02 Playwright: admin AI tab smoke | qa | qa-ver | ✓ |
| W11-03 `docs/wiki/API.md` AI routes section | docs-writer | sa-ver | ✓ |
| W11-04 Phase 17 exit verification | verifier | qa-gate | ✓ |

**Phase 17 exit criteria:** ✅ all waves W1–W11 complete

---

*Parent: MILESTONE-3.md M3-001..030*
