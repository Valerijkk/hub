# BE — Backend Developer

## Роль

Реализация серверной части Hub по спецификации и архитектуре.

## Обязательные docs/

- `docs/system_specification.md`
- `docs/architecture.md`
- `docs/code_style_guide.md`
- `docs/security_guidelines.md`
- `docs/test_plan.md`

## Skill-файл Cursor

`.cursor/project_skills/BE_skill.md` (этот файл)

## Шаги работы

1. Загрузить все файлы из `docs/`
2. Реализовать атомарный модуль (≤200–300 строк за итерацию)
3. Соблюдать handler → service → repository layering
4. Unit-тесты в `backend/internal/<module>/`
5. **Docs sync** — обновить затронутые docs по `.cursor/rules/55-docs-sync.mdc` (API → `system_specification.md` + `wiki/API.md`; tests → `COVERAGE.md`)
6. Запустить регрессию: `go build ./... && go test -race -count=1 ./...` (**обязательно после каждого изменения**, см. `.cursor/REGULATION.md` §8)
7. При FAIL теста → **test triage** (§10): сначала проверить функционал по docs, потом чинить код или тест
8. Приложить описание простым языком (2–5 предложений)
9. Передать BE_VER

## GSD Skills

| Задача | Skill |
|--------|-------|
| Мелкий фикс | `gsd-quick`, `gsd-fast` |
| Отладка | `gsd-debug` |
| Тесты | `gsd-add-tests` |
| Test triage | `.cursor/skills/test-triage/SKILL.md` |

Матрица: `.cursor/GSD-INTEGRATION.md`

## Модульная структура

```
backend/internal/<module>/
├── handler.go, service.go, repository.go, ent_repository.go, dto.go
└── <module>_test.go
```

## Критические правила Hub

- Fiber v3, не net/http
- UUID v4, не strings
- `*ent.Client` только в ent_repository.go
- `safehttp` для user URLs
- `subtle.ConstantTimeCompare` для secrets
- Context первым параметром

## Верификация

| Статус | Дата | Верификатор |
|--------|------|-------------|
| verified: true | 2026-06-17 | BE_VER |

## Лог изменений

| Дата | Изменение | Файл |
|------|-----------|------|
| 2026-06-17 | NF-12 refresh SHA-256 at rest, SEC-06 pg_advisory lock, NEW-02 SafeHTTP OG, NF-09 filename sanitize, NEW-03 backup LimitReader, NF-11 trusted proxy | auth, storage, github, backup, main | regression PASS |
| 2026-06-17 | SEC audit v2.0.1: verified SEC-01/02/05/06/07 + INF-04 already fixed in tree | `main.go`, `handler.go`, `hub.go`, `magiclink.go`, `service_registration.go` |
| 2026-06-17 | SEC-04: `ValidateOutboundURL` at GitLab `CreateIntegration` (SafeHTTPClient was already wired) | `gitlab/service.go:69-72` |
| 2026-06-17 | Updated CRITICAL blocker status table | `docs/security_guidelines.md` |
| 2026-06-17 | NEW-01: `MessageSendRateLimit` 300→60/min per spec | `middleware/ratelimit.go`, `middleware_test.go` |
| 2026-06-17 | NF-08: `SecurityHeaders()` middleware (CSP, X-Frame-Options, etc.) | `middleware/security_headers.go`, `main.go:190` |

### SEC audit summary (2026-06-17)

**Fixed / verified in codebase:**
- SEC-01 `main.go:250` — `APIKeyAuth` before first `RegisterRoutes`
- SEC-02 `calls/handler.go:261-264` — mandatory `sha256` JWT claim
- SEC-04 `gitlab/service.go:55,69-72` — `NewSafeHTTPClient` + URL validate on create
- SEC-05 `ws/hub.go:295-302` — close `done` only; producers check `<-done` before send
- SEC-06 `service_registration.go:108` — `firstUserMu` serializes first-admin path
- SEC-07 `magiclink.go:135` — `subtle.ConstantTimeCompare`
- INF-04 `main.go:531-537` — `backupS3Client = storageSvc.S3Client()`
- INF-02 `deploy.yml:63-65` — no `DEV_MODE=true` sed

**Remains (not BE or needs ops):**
- SEC-03 — rotate LiveKit keys if old `livekit-prod.yaml` was deployed; example file only
- INF-01 — verify `departure_timeout: 120` in prod compose heredoc
- INF-03 — rotate Resend/Telegram tokens in local compose
- FE-01 — frontend Rules of Hooks

**Fixed this session (HIGH security):**
- NEW-01 `ratelimit.go:30` — `MessageSendRateLimit` 60/min (was 300)
- NF-08 `security_headers.go` + `main.go:190` — global security headers after recover

**Tests:** `go build ./...` OK; `go test -race -count=1 ./internal/platform/middleware/...` OK
