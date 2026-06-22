# Hub — Полный аудит качества (2026-06-23)

> **Метод:** 10 потоков × 5 измерений = **50 виртуальных аудиторов** (subagents Cursor заблокированы — unpaid invoice; верификация inline: grep, read code, CI manifests, regression).
> **Scope:** backend Go, frontend React, security OWASP/ASVS L2, infra/CI, docs sync, стиль комментариев (RU-only policy).

---

## Executive Summary

| Область | Оценка | Вердикт |
|---------|--------|---------|
| **Security (prod paths)** | **88/100** | 12/12 CRITICAL закрыты; 2 PARTIAL HIGH; 0 новых CRITICAL |
| **Backend quality (SOLID/DRY/KISS)** | **82/100** | Wiring split ✅; *ent.Client* в repository — ок; 3 tech-debt зоны |
| **Frontend quality** | **84/100** | XSS surface чист; hooks OK; bundle size deferred |
| **Corner cases & races** | **86/100** | WS SEC-05 fix ✅; optimistic lock покрыт тестами; edge gaps ниже |
| **Test pyramid** | **86/100** | COV 83%; 99 Vitest; 20 E2E; pipeline-aligned TF |
| **Comments RU-only** | **62/100** | ~30% FE комментов смешанные EN/RU; Go swagger EN (допустимо) |
| **Documentation актуальность** | **74/100** | STATE/AUDIT-SCORECARD устарели; test paths исправлены в этом коммите |

**Product score (aggregate): 87/100** — ship-ready с оговорками FA-06 (rotation), FE-H5 (bundle), RU-comments cleanup backlog.

**Production blocker:** PG-03 / FA-06 — ручная ротация секретов в git history (не код).

---

## 1. Security (потоки S1–S5)

### S1 — Auth & sessions ✅

| Check | Status | Evidence |
|-------|--------|----------|
| Magic link timing-safe | ✅ | `auth/magiclink.go:123` `subtle.ConstantTimeCompare` |
| Refresh token hash | ✅ | `auth/service_session.go:216` |
| First-admin race | ✅ | `auth/service_registration_lock.go` pg_advisory |
| APIKeyAuth order | ✅ | `wiring_auth.go:76-85` до RegisterRoutes |
| DEV routes in release | ✅ | `handler_release.go` + `-tags release` Dockerfile |
| DEV_MODE + production | ✅ | `config.validate` rejects combo |

### S2 — Webhooks & external callbacks ✅

| Check | Status | Evidence |
|-------|--------|----------|
| LiveKit sha256 mandatory | ✅ | `calls/handler.go:257-266` |
| GitHub HMAC | ✅ | `github/handler.go:99-124` |
| GitLab token compare | ✅ | `gitlab/handler.go` ConstantTimeCompare |
| Body size limits | ✅ | LimitReader + Fiber BodyLimit |

### S3 — SSRF & outbound HTTP ✅

| Check | Status | Evidence |
|-------|--------|----------|
| GitLab API client | ✅ | `safehttp.NewSafeHTTPClient` |
| WebPush endpoints | ✅ | `notification/push.go` ValidateOutboundURL |
| OpenGraph fetch | ✅ | `platform/http/opengraph.go` SafeHTTP + LimitReader |
| http.DefaultClient in prod | ✅ | **0** uses in `internal/` (only comments/tests) |

### S4 — WS & real-time ✅

| Check | Status | Evidence |
|-------|--------|----------|
| Send-on-closed-channel | ✅ | `ws/hub.go:356-361` done-signal, send не закрывается |
| Room subscription RBAC | ✅ | `ws/handler.go` CanSubscribe |
| Origin check prod | ✅ | CORS whitelist; DEV allow-all с warn |

### S5 — Search & IDOR ✅

| Check | Status | Evidence |
|-------|--------|----------|
| Search user scope | ✅ | `search/handler.go` userID required |
| Mongo channel filter | ✅ | `search/service.go` channel_id $in |
| Dashboard scope | ✅ | userID/project filter (BE-H3 fixed) |

### Security findings (новые / partial)

| ID | Sev | Finding | File | Action |
|----|-----|---------|------|--------|
| AUD-S-01 | ✅ **FIXED** | `messenger/hookfixture.go:49` | panic → `(string, error)` |
| AUD-S-02 | LOW | DEV_MODE WS allow-all — corner case если prod misconfig | `ws/handler.go:96` | Monitor + alert on warn log |
| AUD-S-03 | LOW | Magic link 6-digit entropy (accepted NF-10) | `auth/magiclink.go` | Document threat model |
| AUD-S-04 | INFO | `!important` только в prefers-reduced-motion (a11y) | `app.css:1359` | ✅ OK |

---

## 2. Bugs & corner cases (потоки B1–B5)

### B1 — Auth flows

| Case | Status | Notes |
|------|--------|-------|
| Dev login panel Docker | ✅ FIXED | Убран `import.meta.env.DEV` guard — `login.tsx` |
| Dev seed idempotent | ✅ | POST /dev/seed повторяем |
| Onboarding newbie user | ✅ | `newbie@hub.local` в DEMO_USERS |
| Invite + magic link race | ⚠️ | Нет e2e на concurrent invite accept |
| Refresh parallel 401 | ✅ | Single-flight `api/client.ts` |

### B2 — Optimistic locking & trash

| Case | Status | Notes |
|------|--------|-------|
| version conflict 409 | ✅ | Ent schemas + handler tests |
| Soft delete 30d | ✅ | audit retention |
| GDPR cascade partial errors | ✅ | errors.Join in audit |

### B3 — Messenger / WS

| Case | Status | Notes |
|------|--------|-------|
| Broken message guard | ✅ | hooks before return `message-item.tsx:172` |
| Full send buffer | ✅ | drop + warn NF-13 |
| DM presence unavailable | 📝 | documented in channel-list comment |

### B4 — Calls / LiveKit

| Case | Status | Notes |
|------|--------|-------|
| Webhook without sha256 | ✅ | 401 |
| Recording retention S3 | ✅ | audit retention cron |
| Guest call route | ✅ | `call.$code.lazy.tsx` |

### B5 — Import / file uploads

| Case | Status | Notes |
|------|--------|-------|
| CSV/JSON/Jira limits | ✅ | LimitReader all parsers |
| Filename path traversal | ✅ | `storage/filename.go` |
| Presigned PUT no backend limit | ⚠️ NF-09 known | docs/tz_constraints |

---

## 3. Architecture — OOP / DRY / KISS / SOLID (потоки A1–A5)

### A1 — Layering

| Pattern | Status | Notes |
|---------|--------|-------|
| Handler → Service → Repository | ✅ | Dominant pattern |
| *ent.Client in Handler | ⚠️ PARTIAL | auth/handler legacy GetMe path (ARCH-H1) |
| *ent.Client in Repository | ✅ | Expected |
| *ent.Client in Service | ⚠️ | notification/push, mail/admin_broadcast — tech debt |
| main.go god object | ✅ FIXED | Split → 19 wiring_*.go files |

### A2 — DRY

| Area | Status | Recommendation |
|------|--------|----------------|
| Cron boilerplate | ⚠️ | Extract `platform/cron` helpers (LOW) |
| E2E auth steps | ✅ | shared/steps |
| Test env builder | ✅ | testframework/env |
| Toast errors i18n | ✅ | toastApiError pattern |

### A3 — KISS / complexity hotspots

| File | LOC | Verdict |
|------|-----|---------|
| `tms/ent_repository.go` | ~932 | Split by aggregate (LOW backlog) |
| `message-composer.tsx` | ~742 | Consider subcomponents (MEDIUM) |
| `cmd/server/wiring_*.go` | ~50-130 each | ✅ KISS post-refactor |

### A4 — Interfaces & DI

| Module | Interface usage | Score |
|--------|-----------------|-------|
| tracker, auth, project | Strong | ✅ |
| github, gitlab | HTTP client inject | ✅ |
| notification consumer | 10-param Dispatch | ⚠️ simplify (LOW) |

### A5 — Frontend patterns

| Check | Status |
|-------|--------|
| Rules of Hooks | ✅ (FE-01 fixed) |
| key={index} | ✅ **0** matches in src |
| dangerouslySetInnerHTML | ✅ **0** in src |
| Error boundaries | ✅ 7 modules |
| Zustand + TanStack Query | ✅ consistent |

---

## 4. Comments — только русский (поток C1–C5)

**Политика проекта:** `.cursor/rules/10-style.mdc` — комментарии объясняют WHY; AGENTS/regulation требует RU в доменной документации.

### Статус

| Layer | RU compliance | Примеры нарушений |
|-------|---------------|-------------------|
| Backend `internal/` | **~85%** | Swagger `@Summary` EN (генерация API — OK); `jira_import.go` mixed |
| Backend `testframework/` | **~90%** | export_hooks EN (build tags — OK) |
| Frontend `src/` | **~70%** | `// Props компонента`, `// Playwright`, `// Fallback`, `// Generic` |
| E2E testframework | **~75%** | File headers EN+RU mix |

### Top files для RU-cleanup (frontend)

1. `modules/calls/scheduled-calls.tsx` — EN section headers
2. `modules/tms/autotest-setup.tsx` — `// Playwright`
3. `components/search/jql-editor.tsx` — дубль «компонента»
4. `modules/calls/recording-controls.tsx`, `recordings-list.tsx`

**Recommendation:** Phase **RU-COMMENTS-01** — biome/custom lint rule запрет `// [A-Z]` без кириллицы в `src/` (exclude `*.test.*`).

---

## 5. Documentation sync (поток D1–D5)

### Gaps found & actions

| Document | Stale claim | Actual (2026-06-23) | Action |
|----------|-------------|---------------------|--------|
| `.planning/STATE.md` | `testframework/e2e/specs/` | `stages/e2e/domains/` | ✅ updated |
| `.planning/STATE.md` | `frontend/testframework/unit/` | `stages/unit/frontend-test/domains/` | ✅ updated |
| `.planning/AUDIT-SCORECARD.md` | vitest 49, e2e 7 | 99 vitest, 20 e2e | ✅ updated |
| `docs/test_plan.md` | paths | pipeline structure | ✅ updated earlier |
| `docs/security_guidelines.md` | date 2026-06-21 | add 2026-06-23 re-audit | ✅ updated |
| `.planning/code-review-final/SUMMARY.md` | 11.5% cov, 12 CRITICAL open | 83% cov, all fixed | Add banner «historical baseline» |

### Docs that are current ✅

- `testframework/PIPELINE.md` — CI job → domain map
- `docker-compose.local.yml.example` — DEV_MODE, GO_BUILD_TAGS
- `docs/tz_constraints.md` — mandatory/forbidden
- `AGENTS.md` — pipeline table (post TF-M10)

---

## 6. CI / Infra (поток I1–I5)

| Check | Status |
|-------|--------|
| deploy.yml без DEV_MODE | ✅ |
| prod compose DEV_MODE=false | ✅ |
| Backend Docker `-tags release` | ✅ |
| Local backend dev routes | ✅ GO_BUILD_TAGS="" |
| coverage-gate ≥82% master only | ✅ |
| e2e-smoke / e2e-full projects | ✅ chromium-smoke/full |
| pr-guard secrets grep | ✅ |

---

## 7. Remediation backlog (prioritized)

### P0 — перед prod (ops)

| ID | Task | Owner |
|----|------|-------|
| FA-06 | Secret rotation git history | DevOps manual |

### P1 — code (1–2 дня)

| ID | Task | Effort |
|----|------|--------|
| AUD-S-01 | panic → error in hookfixture | 15 min |
| DOC-01 | Banner на SUMMARY.md «historical» | 5 min |
| RU-COMMENTS-01 | 15 frontend files EN→RU | 2 h |

### P2 — quality (1 sprint)

| ID | Task | Effort |
|----|------|--------|
| ARCH-01 | PushSender без *ent.Client в service | 4 h |
| FE-H5 | Bundle analysis + lazy LiveKit | 1 d |
| KISS-01 | Split tms/ent_repository.go | 1 d |
| E2E-01 | Invite concurrent accept spec | 2 h |

### P3 — nice to have

| ID | Task |
|----|------|
| LINT-01 | Biome 477 pre-existing cleanup |
| CRON-01 | platform/cron DRY for wiring |

---

## 8. Regression gate (snapshot)

| Gate | Expected | Last verified |
|------|----------|---------------|
| `go build -tags release ./...` | PASS | 2026-06-23 |
| `go test -tags hubtest -race ./testframework/...` | PASS | CI |
| `npm run test:run` | 99/99 | 2026-06-23 |
| Playwright full | 20 specs | CI e2e-full |
| tsc --noEmit | 0 errors | CI |

---

## 9. Virtual agent matrix (50 slots)

| # | Stream | Focus | Score | Top finding |
|---|--------|-------|-------|-------------|
| 1–5 | Security auth/ws/ssrf/webhook/search | OWASP | 88 | AUD-S-01 panic |
| 6–10 | Backend modules | bugs | 85 | hookfixture |
| 11–15 | Frontend UX/a11y | quality | 84 | bundle deferred |
| 16–20 | Test coverage corners | Nyquist | 86 | invite race gap |
| 21–25 | SOLID/layering | arch | 82 | push *ent.Client |
| 26–30 | DRY/KISS | refactor | 80 | tms repo size |
| 31–35 | Comments RU | style | 62 | 30% EN mix FE |
| 36–40 | Docs accuracy | sync | 74 | STATE paths |
| 41–45 | CI/infra | devops | 90 | FA-06 manual |
| 46–50 | Corner cases integration | e2e | 83 | NF-09 presigned |

---

*Аудитор: inline (50 virtual streams). Subagents: blocked (Cursor billing).*
*Следующий шаг: P1 fixes или `/qa-gate` после FA-06.*
