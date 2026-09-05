# Hub — Матрица покрытия требований

> Phase 0 reconciliation: требование → код → тест.  
> **Источники:** `.planning/PROJECT.md`, `docs/vision_and_scope.md`, `.planning/audits/code-review-final/SUMMARY.md`  
> **Founder cutover decision (2026-08-07):** statement coverage below is
> historical discovery evidence only. It is **not** a migration gate, parity
> target, or KPI for the greenfield Python test architecture. New acceptance is
> fail-closed execution plus task/FR/NFR/verified-TC and risk traceability.
> **Last legacy coverage observation:** **2026-08-07** (Docker Go 1.25.11,
> zero-failure, overall **82.0373%**)
> Previous valid gate: 2026-07-27 Docker **81.9%** (see ## Historical baselines)
> Contaminated 2026-07-29 local attempt **80.2%** is historical, not baseline.
> **Current minimal gate (2026-08-23):** root Compose exact services are the QA
> authority. `backend-unit`, `frontend-unit` and framework-only
> `python-contract` have Docker PASS evidence; Python API/integration/e2e/load
> are `NOT_RUN` and outside this merge gate. Host test/validator commands and
> removed legacy entrypoints are forbidden.
> Historical package table: 2026-06-22 (see ## Historical baselines)  
> **Источники:** `.planning/PROJECT.md`, `docs/vision_and_scope.md`, `.planning/audits/code-review-final/SUMMARY.md`

---

## Методология

### Дополнение ветки calls-feature-codex — 2026-09-05

MSG151-FOL-01/02 + MSG151-CHN-02: управление подпиской в ThreadPanel принято
на unit-уровне. 29 новых изолированных RTL/Vitest сценариев покрывают режимы,
capabilities, identity/revoke, single-flight, точные server-refetch ключи,
RU/EN ошибки, фоновые гонки и focus-return регрессию. Docker frontend-unit:
TypeScript PASS; 190 файлов / 630 тестов PASS, 54.44s, 2 GiB/no-swap. Это не процент
готовности модуля или новый statement coverage; API/e2e/browser/media и
real visible-read browser receipt остаются NOT_RUN/UNBOUND. [Evidence](../docs/superpowers/plans/2026-09-05-thread-follow-control.md).

Дополнение 2026-09-06: MSG151-UNR read-through + FOL/CHN gates — автоматическая
отметка по видимому концу ответа принята на unit-уровне. 36 новых genuine
сценариев: геометрический witness, identity/root/revoke, follow/capabilities,
ошибка/retry, pending coalescing и watermark, focus/refetch recovery, exact
unread refresh без арифметики. Последний Docker frontend-unit: TypeScript PASS,
191 файлов / 666 тестов PASS,58.20s,2GiB/no-swap; независимое review C/I/M0/0/0.
Это не процент готовности продукта. Browser/API/storage/media NOT_RUN.
[Evidence](../docs/superpowers/plans/2026-09-05-thread-read-receipts.md).

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

## Last legacy coverage observation (measured **2026-08-07**, Docker)

This section records what existed before the founder-approved cutover. It does
not block retaining genuine native units, deleting legacy non-unit suites, or
building new Python suites from requirements and risks.

| Metric | Value | Measured |
|--------|------:|----------|
| Overall `internal/` statements | **82.0373%** (21,269 / 25,926) | **2026-08-07** |
| Historical retired-job floor (>=82%) | **PASS (+0.0373 pp over floor)** | 2026-08-07 |
| vs previous valid gate (2026-07-27) | **+0.1373 pp** (was 81.9%) | - |
| vs 2026-06-22 Docker baseline | **-0.9627 pp** (was 83.0%) | - |
| Backend package results | **50 PASS / 0 FAIL / 3 no-test-files** | 2026-08-07 |
| Backend Go tests in this exact coverage-profile source snapshot | **830** files / **3,667** declared tests: legacy root **829 / 3,665** + server **1 / 2** | historical profile input |
| Backend `internal/*_test.go` | **0** | 2026-08-07 |

**Method:** TASK-225 source at HEAD `c71a213b` plus the cron repair later
committed as `ee6d696c`,
`golang:1.25-bookworm` (Go 1.25.11), Postgres/Mongo/Redis/NATS from
`deployments/docker-compose.test.yml`;
`go test -tags hubtest -race -count=1 -timeout=600s -covermode=atomic -coverpkg=./internal/... -coverprofile=<temp> ./testframework/...`;
exact source-range merge of the atomic profile. Exit `0`, wall time 119 s,
profile SHA-256
`7C20BF7452AD9387D96F548368830282FD2A50361A0D0106E7B8069FC4C99856`.

Current cutover inventory is newer and deliberately does not rewrite this
historical profile: pinned revision `b9f690da`, ledger **1,049 / 4,246**, Go
**832 / 3,674**, SHA-256 `83b6a311…`. It is the path/action authority; the
82.0373% profile remains discovery evidence only.

### Security packages — same valid full profile

Repeated blocks emitted by multi-package `-coverpkg` are merged by source
range; a statement is covered when any matching counter is non-zero.

| Package | Covered / total | Coverage | Historical floor >=80% |
|---|---:|---:|:---:|
| auth | 686 / 851 | **80.6110%** | ✅ |
| ws | 408 / 443 | **92.0993%** | ✅ |
| rbac | 337 / 387 | **87.0801%** | ✅ |
| audit | 928 / 1,083 | **85.6879%** | ✅ |
| calls | 1,059 / 1,281 | **82.6698%** | ✅ |
| platform/middleware | 207 / 245 | **84.4898%** | ✅ |

Current full-profile observation: **6/6** security packages >=80%.

**Evidence:** `.planning/audits/test-rebase/baseline/TASK-203-BASELINE-2026-08-07.md`.

### Покрытие по пакетам — два метода измерения

**Важно перед чтением чисел.** Один и тот же пакет даёт разное покрытие
в зависимости от того, как считали. Ниже обе колонки, и они не
противоречат друг другу.

| Метод | Что считает | Когда применим |
|-------|-------------|----------------|
| **Сквозной** (`-coverpkg=./internal/...` на всём наборе) | Долю строк пакета, задетых **любыми** тестами репозитория | Общая цифра по проекту |
| **Пакетный** (`-coverpkg=./internal/<pkg>/...` + только его тесты) | Долю строк, покрытых **своими** тестами пакета | Метод старого, выведенного из эксплуатации job |

Сквозной метод занижал историческое значение: строки, покрытые только тестами
своего пакета, размывались в общем прогоне. Это описание старого job, не
текущий gate и не инструкция к запуску.

#### Пакетный замер старого job (2026-07-30; historical/non-gating)

| Пакет | Пакетный | Сквозной | Исторический результат |
|-------|---------:|---------:|:----:|
| rbac | **86.8%** | 77.9% | ✅ |
| storage | **83.9%** | 80.3% | ✅ |
| messenger | **82.8%** | 69.7% | ✅ |

Расхождение до 13 пунктов между методами было нормой для прежней методики.
Исторический threshold по всем трём был достигнут; для cutover это не gate.

**Historical inner command (не запускать напрямую и не на host):**

Ниже только provenance старого evidence. До готовности versioned outer
launcher runtime остаётся `NOT_READY`; прямой вызов не является fallback.

```bash
go test -tags hubtest ./testframework/<pkg>/... -count=1 \
  -coverpkg=./internal/<pkg>/...
```

#### Требуют внимания (сквозной замер 2026-07-29)

Пакетный замер по ним не делался — цифры сквозные, реальное покрытие
своими тестами выше.

| Пакет | Сквозной | vs 2026-06-22 |
|-------|---------:|--------------:|
| intranet | 76.0% | −5.3 п.п. |
| github | 77.4% | −3.4 п.п. |
| platform/redis | 66.7% | нет прежней строки |

Полная таблица и список FAIL: `COV-2026-07-29.md`.
Доказательство лифта messenger и rbac:
`.planning/audits/total-2026-07/COV-2026-07-30-messenger-rbac.md`.
Доказательство лифта storage: коммит `bdabd58c`.

The 2026-07-29 overall value (**80.2%**) is a contaminated historical attempt;
the current zero-failure result is **82.0373%** from 2026-08-07.

### Historical baselines

#### Contaminated local attempt **2026-07-29** — not a baseline

| Metric | Value |
|---|---:|
| Overall `internal/` statements | **80.2%** |
| Package failures | **11** |
| Valid for current gate | **No** |

Method: local Windows host with `-tags hubtest`, `-covermode=atomic` and
`-coverpkg=./internal/...`. The `ai` build failure excluded `internal/ai` from
the profile, so the sample is contaminated. Evidence:
`.planning/audits/total-2026-07/COV-2026-07-29.md`.

#### Coverage gate **2026-07-27** (Docker) — superseded 2026-08-07

| Metric | Value | Measured |
|--------|------:|----------|
| Overall `internal/` statements | **81.9%** | **2026-07-27** |
| Historical retired-job floor (>=82%) | **MISS (-0.1 pp under floor)** | 2026-07-27 |
| vs previous baseline | **-1.1 pp** (was 83.0% on 2026-06-22) | - |

**Method:** `golang:1.25-bookworm` on `hub_default`,  
`go test -tags hubtest -count=1 -timeout=900s -covermode=atomic -coverpkg=./internal/... ./testframework/...`,  
`go tool cover -func` -> overall statements.

**Evidence:** `.planning/audits/planning-audit-2026-07/cov-2026-07-27-REPORT.md`, `cov-2026-07-27-total.txt`.

Per-package **statement** table for 2026-07-27 was **pending** (coverprofile export hit `proxy.golang.org` EOF).  
Overall-only comparable decline at that time: **83.0% -> 81.9% (-1.1 pp)**.


Moved prior “current” package table here. Do not present as live without re-run date.

#### Test baseline (coverage **2026-06-22**; file inventory later refreshed)

| Layer | Count | Notes |
|-------|-------|-------|
| Backend `testframework/**/*_test.go` | **515** files | Inventory 2026-07-27 (was 499 on 2026-07-26) |
| Backend `internal/*_test.go` (remaining) | **0** files | TF-M08 ✅ verified |
| Frontend Vitest `testframework/stages/unit/frontend-test/domains/` | **37** files | **99/99** tests · 0 in `src/` |
| Playwright E2E | **20** spec files | `testframework/stages/e2e/domains/` |
| E2E shared (POM/steps) | pages, steps, helpers | `testframework/stages/e2e/shared/` |
| Statement coverage (overall `internal/`) | **83.0%** | **2026-06-22** Docker baseline; post FA-11/12 |

#### Покрытие по пакетам (**2026-06-22** Docker baseline — historical)

| Package | Coverage | Phase / Notes |
|---------|----------|---------------|
| **Overall `internal/`** | **83.0%** | 2026-06-22 Docker baseline; Wave C post FA-11/12 |
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

#### Security-critical packages (Wave 1 + Phase 125; **2026-06-22** Docker baseline — historical)

| Package | Coverage | Status |
|---------|----------|--------|
| auth | **82.4%** (2026-06-22) | ✅ ≥80% |
| ws | **87.7%** (2026-06-22) | ✅ ≥80% |
| rbac | **84.7%** (2026-06-22) | ✅ ≥80% |
| audit | **80.5%** (2026-06-22) | ✅ ≥80% |
| platform/middleware | **84.0%** (2026-06-22) | ✅ ≥80% |
| calls | **80.2%** (2026-06-22) | ✅ ≥80% |

**Target Wave 1:** ≥80% per security package — **6/6 met** on 2026-06-22 baseline ✅  
(2026-07-27 security package statement re-check: pending coverprofile.)

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
| **COV-01 (historical)** | Retired legacy coverage floor | G-04 | historical | historical | **82.0373%** overall (**2026-08-07**); discovery evidence only, not a cutover gate or KPI |

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

**Historical verdict:** legacy COV-01 gate **PASS** (82.0373% >=82%,
zero-failure Docker baseline 2026-08-07); current full-profile security observation is **6/6**
with `audit` at 85.6879%. Historical 2026-06-22 verdict was PASS at 83.0% and
6/6 security packages; Wave C FA-02 re-audit remains historical DONE — 0
CRITICAL regression (`.planning/audits/2026-06/FA-02-SECURITY-REAUDIT.md`).

---

## Historical module coverage summary (2026-06-22; non-gating)

This table predates the founder cutover. Its “tests” and E2E cells are legacy
observations, not native-unit taxonomy or acceptance of the future runner.

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
| 1 | Greenfield test architecture cutover (Phases 128–141) | Minimal package/framework slice Docker PASS; API/integration/e2e/load `NOT_RUN` and tracked separately |
| 2 | Historical FA-11 tms/project legacy floor | archived/non-gating |
| 3 | NFR-16 full WCAG AA sign-off | ⚠️ partial |
| 4 | NFR-17 Lighthouse ≥90 in CI | ⚠️ artifact job (non-blocking) |
| 5 | FA-06 / PG-03 secret rotation execute | ⏳ manual ops |
| 6 | FA-04 github http.DefaultClient fallback | ✅ safehttp |
| 7 | FA-05 hub-mail prod compose | ✅ 2026-06-22 |
| 8 | FA-02-02 TMS webhook RBAC (MEDIUM) | ✅ FA-RBAC-FIBER3 123 routes |
| 9 | Squash PR v2.5-trust → master | ✅ PR #1 merged 2026-06-22 + hotfixes |
| 10 | FA-10 full qa-gate Docker | historical legacy PASS only; not acceptance of the future runner |

---

## Next Actions

1. **DevOps:** FA-06 PG-03 Phase 1–3 (manual) — scripts: `bash scripts/fa06.sh p1-compose` or `p1-k3s` → `p2-smoke` → CTO sign-off
2. **DevOps:** SSH deploy — `bash scripts/fa06.sh deploy` · GitHub Secrets SSH_HOST/SSH_KEY
3. **Release:** `bash scripts/fa06.sh release-gate` → `git tag v2.5.0` after FA-06 complete
4. **Backlog:** NFR-17 Lighthouse blocking gate · NFR-16 WCAG AA sign-off · `.planning/audits/2026-06/AUDIT-2026-06-23.md`

---

*Generated: 2026-06-17 · Verified: 2026-06-23 (AUDIT-2026-06-23: COV 83.0% = 2026-06-22 Docker baseline, master bbdaf16, CI green)*
