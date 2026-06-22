# Phase 74 — AI Module Coverage (M3 / COV-01)

**Branch:** `feat/v2.5-trust`  
**Status:** COMPLETE  
**Date:** 2026-06-21

## Goal

`internal/ai/` **24.8% → ≥50%** — handlers, repository, settings, privacy, adapters (unit paths without live LLM).

## Deliverables

### Repository & settings
- `repository_ent_test.go` — CRUD providers, models, preferences, usage, empty preference
- `settings_store_test.go` — EntSettingsStore Get/Upsert

### HTTP handlers
- `handler_extended_test.go` — RegisterRoutes, bot/models/usage/preferences, invalid bodies
- `admin_handler_test.go` — admin providers/settings/usage/test
- `user_handler_test.go` — BYOK `/ai/my/providers` CRUD

### Service & privacy
- `service_test.go` (existing) + fallback/stream paths
- `privacy_test.go` — consent filter, resolveProviders, consentErrorForModel
- `issue_lookup_test.go` — nil tracker, OfflineOnlyEnabled
- `provider_rpm_test.go` — Redis/Noop RPM limiter, SetProviderRPMLimiter

### Adapters & bot helpers
- `adapter_openai_test.go` — stream rejected, invalid URL, truncate
- `bot_user_test.go`, `bot_commands_test.go`, `stream_test.go`

## Results

| Metric | Before | After |
|--------|--------|-------|
| `internal/ai/` | 24.8% | **50.7%** |
| Test files | ~11 | **21** |

## Deferred (out of scope)

- `messenger_bot.go` OnMessageSent (needs messenger mocks)
- `FetchMRForReview` / live adapter HTTP (httptest GitLab/GitHub)
- `audit.go` LogAIChat (full audit.Repository mock too heavy)

## Success criteria

- [x] ai/ ≥50%
- [x] `go test -race ./internal/ai/...` PASS
- [x] Handler routes + admin/user BYOK covered
- [x] EntRepository preferences/usage covered
