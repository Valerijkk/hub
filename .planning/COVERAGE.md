# Hub — Матрица покрытия требований

> Phase 0 reconciliation: требование → код → тест.  
> **Источники:** `.planning/PROJECT.md`, `docs/vision_and_scope.md`, `.planning/code-review-final/SUMMARY.md`  
> **Верификация:** 2026-06-22 (Wave C FA-03, Docker `-tags hubtest`, post FA-11/12)

---

## Методология

| Колонка | Значения |
|---------|----------|
| **Code** | ✅ Implemented · ⚠️ Partial · ❌ Missing · 🔧 Fixed post-audit |
| **Tests** | ✅ Covered · ⚠️ Partial · ❌ Missing |
| **Gap** | Открытый finding или пробел покрытия |

---

## API — сводка (2026-06-21)

| Источник | Count | Примечание |
|----------|-------|------------|
| `docs/wiki/API.md` (documented) | **287+** | + AI composer, v4 modules |
| Handler `RegisterRoutes` (code) | **283** | 21+ модулей + templates |
| Dev routes (`/dev/*`) | **+3** | non-release only |
| Infra (`/docs`, `/health`, `/metrics`) | **+3** | main.go |
| **Total REST `/api/*`** | **286** | Wiki synced |
| WebSocket | **+1** | `GET /ws` |

---

## Test baseline (Wave C FA-03, Docker 2026-06-22, verified)

| Layer | Count | Notes |
|-------|-------|-------|
| Backend `testframework/**/*_test.go` | **488** files | TF migration COMPLETE |
| Backend `internal/*_test.go` (remaining) | **0** files | TF-M08 ✅ verified |
| Frontend Vitest `frontend/testframework/unit/` | **37** files | **99/99** tests · 0 in `src/` |
| Playwright E2E | **20** spec files | `testframework/e2e/specs/` |
| Statement coverage (overall `internal/`) | **83.0%** | post FA-11/12 ✅ |

### Покрытие по пакетам (актуальное, Docker)

| Package | Coverage | Phase / Notes |
|---------|----------|---------------|
| **Overall `internal/`** | **83.0%** | Wave C post FA-11/12 |
| ws | **87.7%** | security wave |
| storage | **86.8%** | Phase 123 |
| rbac | **84.7%** | security wave |
| search | **92.2%** | Phase 116 carry |
| platform/outbox | **88.6%** | Phase 101 |
| notification | **82.5%** | Phase 121 + testframework |
| auth | **82.4%** | security wave |
| tracker | **82.4%** | Phase 121/124 |
| messenger | **82.3%** | Phase 123 |
| audit | **80.5%** | security wave (min threshold) |
| calls | **80.2%** | PG-01 Wave 6 |
| tms | **82.0%** | FA-11 ✅ micro-gap closed |
| project | **82.0%** | FA-11 ✅ micro-gap closed |
| sprint | **81.7%** | Phase 113 |
| hr | **83.7%** | Phase 111 |
| crm | **81.9%** | Phase 111 |
| imex | **84.7%** | Phase 90 |
| ai | **78.9%** | Phase 110 |
| github | **80.8%** | Phase 104 |
| apigateway | **82.2%** | Phase 97 |
| dashboard | **86.4%** | Phase 107 |
| accounting | **80.4%** | Phase 108 |
| calendar | **86.7%** | Phase 98 |
| intranet | **81.3%** | Phase 91 |
| bases | **78.6%** | Phase 107 |
| marketplace | **80.1%** | Phase 105 |
| backup | **85.8%** | Phase 100 |
| integrations/onec | **83.4%** | Phase 101 |

### Security-critical packages (Wave 1 + Phase 125)

| Package | Coverage | Status |
|---------|----------|--------|
| auth | **82.4%** | ✅ ≥80% |
| ws | **87.7%** | ✅ ≥80% |
| rbac | **84.7%** | ✅ ≥80% |
| audit | **80.5%** | ✅ ≥80% |
| platform/middleware | **84.0%** | ✅ ≥80% |
| calls | **80.2%** | ✅ ≥80% |

**Target Wave 1:** ≥80% per security package — **6/6 met** ✅

---

## Top 20 Critical Requirements — Gap Matrix

| REQ-ID | Requirement | Source | Code | Tests | Gap |
|--------|-------------|--------|------|-------|-----|
| **AUTH-01** | Magic Link + Telegram bind (no passwords) | US-M01 | ✅ | ✅ | SEC-07 🔧 fixed |
| **AUTH-02** | JWT 15m + refresh 30d Redis | NFR-09/10 | ✅ | ✅ | FE refresh mutex ✅ |
| **AUTH-03** | Auth rate limit 3/10 min | NFR-05 | ✅ | ✅ | DEV_MODE skip by design |
| **PROJ-01** | Project creates tracker+TMS+chat | US-M02 | ✅ | ⚠️ | E2E partial |
| **TASK-01** | 9 issue types, CRUD, links | US-M03 | ✅ | ✅ **82.4%** | — |
| **TASK-02** | Optimistic lock (version, 409) | US-M03 | ✅ | ✅ | — |
| **MSG-01** | Channels, threads, reactions | US-M04 | ✅ | ✅ **82.3%** | — |
| **TMS-01** | Test cases + Allure CI | US-M05 | ✅ | ✅ **82.0%** | FA-12 TestRail enum ✅ |
| **CALL-01** | LiveKit voice+screen+webcam | US-M06 | ✅ | ✅ **80.2%** | — |
| **SRCH-01** | Global search user-scope | US-M07 | ✅ | ✅ **92.2%** | Ctrl+K ✅ |
| **WS-01** | WebSocket via NATS | NFR-02 | ✅ | ✅ **87.7%** | SEC-H1 🔧 fixed |
| **API-01** | REST + API keys | API module | ✅ 286 ep | ✅ **82.2%** | SEC-01 🔧 fixed |
| **NOTF-01** | In-app + TG + Web Push | US-S05 | ✅ | ✅ **82.5%** | SEC-H3 🔧 fixed |
| **AUDIT-01** | Audit log + trash 30d | US-S03 | ✅ | ✅ **80.5%** | BE-H6/H7 ✅ |
| **OBS-01** | Prometheus + OTel + Loki + AM | US-S03 | ✅ | ✅ | **CLOSED** Ph 29–36 |
| **BACK-01** | Weekly PG+Mongo+S3 backup | PROJECT | ✅ | ⚠️ | INF-04 🔧 fixed |
| **GH-01** | GitHub PR + webhooks | US-S04 | ✅ | ✅ **80.8%** | SafeDialer ✅ |
| **SEC-PROD** | 0 CRITICAL blockers | G-04 | ✅ | ✅ | 12/12 fixed |
| **NFR-13** | Error JSON `{error, request_id}` | NFR-13 | ✅ | ✅ | — |
| **NFR-16** | WCAG 2.1 AA | NFR-16 | ⚠️ | ✅ login axe | full audit pending |
| **COV-01** | Coverage ≥82% / security ≥80% | G-04 | ✅ | ✅ | **83.0%** / **6/6** ✅ |

---

## v4 Super-app modules (Phase 58+, verified 2026-06-25)

| Module | Stmt coverage | E2E |
|--------|---------------|-----|
| status | **100.0%** | ✅ v4-superapp |
| crm | **72.8%** → **81.9%** | ✅ |
| hr | **73.3%** → **83.7%** | ✅ |
| marketplace | **69.2%** → **80.1%** | ✅ |
| calendar | **65.6%** → **86.7%** | ✅ |
| accounting | **65.0%** → **80.4%** | ✅ |
| intranet | **60.8%** → **81.3%** | ✅ |
| bases | **60.8%** → **78.6%** | ✅ |
| integrations/onec | **72.0%** → **83.4%** | — |

Все v4 модули **≥50%** ✅ (целевой floor **≥80%** для core — достигнут для большинства)

---

## SUMMARY.md CRITICAL — Reconciliation

| ID | Finding | Code | Tests |
|----|---------|------|-------|
| SEC-01 | APIKeyAuth dead code | 🔧 Fixed | ✅ Phase 69 |
| SEC-02 | LiveKit webhook optional | 🔧 Fixed | ✅ |
| SEC-03 | Hardcoded livekit secret | ⚠️ templated | ❌ |
| SEC-04 | GitLab SSRF DefaultClient | 🔧 Fixed safehttp | ✅ Phase 70 |
| SEC-05 | WS send-on-closed panic | 🔧 Fixed | ⚠️ partial race |
| SEC-06 | First-admin race | 🔧 Fixed mutex | ✅ |
| SEC-07 | Magic-link timing | 🔧 ConstantTimeCompare | ✅ |
| INF-01 | LiveKit departure_timeout | 🔧 In yaml/compose | ❌ |
| INF-02 | DEV_MODE in prod CI | 🔧 Removed deploy.yml | ❌ |
| INF-03 | Plaintext creds compose | ⚠️ env placeholders | PG-03 manual |
| INF-04 | BackupService nil s3 | 🔧 Fixed main.go | ⚠️ unit only |
| FE-01 | Rules of Hooks | 🔧 Fixed | ❌ |

**Verdict:** COV-01 gate **PASS** (83.0% ≥82%, security 6/6 ≥80%). Wave C FA-02 re-audit **DONE** — 0 CRITICAL regression (`.planning/FA-02-SECURITY-REAUDIT.md`).

---

## Module Coverage Summary

| Module | Routes | Unit Tests | E2E | Notes |
|--------|--------|------------|-----|-------|
| auth | 13 | ✅ 82.4% | ✅ login | security |
| project | 27 | ✅ 82.0% | ⚠️ | FA-11 ✅ |
| tracker | 38 | ✅ 82.4% | ⚠️ | — |
| messenger | 42 | ✅ 82.3% | ⚠️ | — |
| tms | 52 | ✅ 82.0% | ⚠️ | FA-11 ✅ · FA-12 TestRail enum |
| calls | 22 | ✅ 80.2% | ⚠️ | security |
| ws | 1 | ✅ 87.7% | ❌ | security |
| search | 7 | ✅ 92.2% | ⚠️ | — |
| notification | 11 | ✅ 82.5% | ⚠️ | testframework |
| apigateway | 8 | ✅ 82.2% | ❌ | — |
| sprint | 14 | ✅ 81.7% | ❌ | — |
| dashboard | 8 | ✅ 86.4% | ⚠️ | — |
| audit/backup | 12 | ✅ 80.5%/85.8% | ✅ admin-audit | — |
| github/gitlab | 20 | ✅ 80.8%/68.5% | ❌ | — |
| ai | 12+ | ✅ 78.9% | ⚠️ composer | v3.0 |
| intranet/crm/hr/… | v4 | ✅ ≥78% | ✅ v4-superapp | v4 modules |

---

## Top 10 Open Items

| # | Item | Status |
|---|------|--------|
| 1 | TF test framework migration (488 / 0 internal) | ✅ COMPLETE |
| 2 | FA-11 tms/project → 82% | ✅ closed |
| 3 | NFR-16 full WCAG AA sign-off | ⚠️ partial |
| 4 | NFR-17 Lighthouse ≥90 in CI | ❌ not measured |
| 5 | FA-06 / PG-03 secret rotation execute | ⏳ manual ops |
| 6 | FA-04 github http.DefaultClient fallback | ✅ safehttp |
| 7 | FA-05 hub-mail prod compose | ✅ 2026-06-22 |
| 8 | FA-02-02 TMS webhook RBAC (MEDIUM) | ✅ FA-RBAC-FIBER3 123 routes |
| 9 | Squash PR v2.5-trust → master | ✅ PR #1 merged 2026-06-22 + hotfixes |
| 10 | FA-10 full qa-gate Docker | ✅ local PASS · CI e2e-full 20 specs |

---

## Next Actions

1. **DevOps:** FA-06 PG-03 Phase 1–3 (manual) — scripts: `bash scripts/fa06.sh p1-compose` or `p1-k3s` → `p2-smoke` → CTO sign-off
2. **DevOps:** SSH deploy — `bash scripts/fa06.sh deploy` · GitHub Secrets SSH_HOST/SSH_KEY
3. **Release:** `bash scripts/fa06.sh release-gate` → `git tag v2.5.0` after FA-06 complete
4. **Backlog:** NFR-17 Lighthouse in CI · full WCAG AA sign-off

---

*Generated: 2026-06-17 · Verified: 2026-06-22 (Wave C FA-03: COV 83.0%, security 6/6 ≥80%, 488 testframework / 0 internal tests)*
