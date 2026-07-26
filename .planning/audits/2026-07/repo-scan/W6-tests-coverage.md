# W6 — Tests / coverage-gate / COVERAGE drift (readonly)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> **Агент:** `qa-auto-ver` (Grok 4.5) · **Дата:** 2026-07-23 01:19 UTC+3
> **Scope:** `.planning/COVERAGE.md`, CI `coverage-gate`, security **6×80%**, inventory/docs drift, детерминизм (sleep/timeout), TC↔автотест.
> **Код не менялся.** Живой `go test -cover` / полный `/qa-gate` в этом слоте **не гонялись** (scan-only).
> **Примечание:** восстановлено оркестратором — субагент был в Ask mode.

## Вердикт (кратко)

| Контроль | Статус | Evidence |
|----------|--------|----------|
| Overall gate ≥82% в CI | PASS (конфиг) | `ci.yml:117-129` · `coverage-gate.yaml` `threshold_percent: 82` |
| Security 6×80% в CI | **FAIL (не enforced)** | В workflow **нет** per-package порога; только total |
| COVERAGE.md = live inventory | **DRIFT** | docs: **495** vs диск: **499** `_test.go` |
| COV % freshness | STALE | Verified **2026-06-22/23**; пересчёт июль **не приложен** |
| Vitest / E2E counts | PASS | 37 files · 20 specs (совпадает с COVERAGE/TESTING) |
| `internal/*_test.go` = 0 | PASS | TF-M08 |
| Детерминизм E2E | **FAIL** | `waitForTimeout(3000)` в `ai-hub.spec.ts:228` |
| Детерминизм Go | PARTIAL | **16** `time.Sleep` в 15 файлах `backend/testframework/` |
| TC↔автотест (рубрика 2.4) | PARTIAL | Часть AUTH — без точного имени функции |
| Coverage ≠ authz quality | **GAP** | W2 CRITICAL IDOR при COV tms/project ≥82% |

**verified: false** — 2026-07-23 01:19 UTC+3
Блокеры: нет CI-энфорса 6×80%; inventory/docs drift; flaky-паттерн E2E; нет свежего cover-лога.

---

## Baseline (канон vs диск)

| Метрика | COVERAGE.md / AGENTS | Диск / TESTING.md (2026-07-23) | Drift |
|---------|----------------------|--------------------------------|-------|
| BE `testframework/**/*_test.go` | **495** (`COVERAGE` · `AGENTS` таблица) | **499** | **+4** HIGH docs |
| BE `internal/*_test.go` | 0 | 0 | OK |
| Vitest domain files | 37 / 99 tests | 37 | OK |
| Playwright specs | 20 | 20 | OK |
| Overall `internal/` | **83.0%** (2026-06-22 Docker baseline) | не переизмерено | STALE |
| Security 6×80% | **6/6 ✅** (auth 82.4, ws 87.7, rbac 84.7, audit 80.5, middleware 84.0, calls 80.2) | не переизмерено; **CI не проверяет** | PROCESS GAP |

Примечание: `AGENTS.md` внутри себя противоречит — таблица **495**, блок команд «все **499**». `docs/TESTING.md` §счётчики = **499** (ближе к диску).

---

## coverage-gate

**Манифест:** `testframework/pipeline/jobs/coverage-gate.yaml`
- `when: github.event_name != pull_request` (только master / non-PR)
- `threshold_percent: 82`
- `go test -tags hubtest -coverprofile=/tmp/cov -coverpkg=./internal/... ./testframework/...`

**CI:** `.github/workflows/ci.yml` job `coverage-gate`
- Порог: `awk … >= 82.0`
- **Нет** `-race` в этом job (race — в `backend-test`)
- **Нет** проверки пакетов auth/ws/rbac/audit/middleware/calls ≥80%

**README:** `testframework/stages/quality/coverage-gate/README.md` — совпадает с манифестом (≥82%, master-only).

### Security 6×80%

Документировано как MANDATORY в `COVERAGE.md` / `AGENTS.md` / `docs/TESTING.md`, но **ни один CI step не валит билд** при падении security-пакета ниже 80%, пока overall ≥82%.
Риск: регрессия `audit`/`calls` (у порога 80.2–80.5%) незаметна на PR и даже на master coverage-gate.

---

## Findings

| Sev | file:line / артефакт | Наблюдение | Рекомендация |
|-----|----------------------|------------|--------------|
| **HIGH** | `.github/workflows/ci.yml:117-129` · `coverage-gate.yaml` | Security **6×80%** заявлены как gate, в CI **не enforced** — только overall ≥82%. | Добавить post-step `go tool cover -func` + awk по 6 пакетам; FAIL если любой <80. Документировать в README coverage-gate. |
| **HIGH** | `.planning/COVERAGE.md:36` · `AGENTS.md:143` vs диск **499** | Inventory drift: канон coverage matrix отстаёт на 4 файла; `AGENTS` 495 vs 499. | Синхронизировать COVERAGE + AGENTS → **499**; одна цифра в каноне. |
| **HIGH** | `COVERAGE.md` footer Verified 2026-06-23 | Package % / 83.0% (2026-06-22 Docker baseline) / 6/6 — **месяц без remeasure** в этом scan. Нельзя считать COV-01 live PASS. | Перегнать Docker cover → обновить таблицу; зафиксировать commit SHA. |
| **HIGH** | `testframework/stages/e2e/domains/ai/ai-hub.spec.ts:228` | `await page.waitForTimeout(3_000)` — синхронизация «на удачу» (CRITICAL по рубрике qa-auto-ver §2.1). | Web-first assert / `expect.poll` на содержимое composer; убрать timeout. |
| **MEDIUM** | `backend/testframework/**` (15 файлов, 16 Sleep) | `time.Sleep` для lifecycle/rate-window (ws, notification consumers, middleware rate, webhook dispatcher, cron…). Часть оправдана (окно rate limit), часть — polling без канала. | Rate-limit: оставить с комментарием WHY; lifecycle → channel/`Eventually`/fake clock. |
| **MEDIUM** | `docs/TESTING.md:43` vs `:33` | В том же файле: актуальные **499** файлов и устаревшая секция «**33** тестовых файла / 272 функции» (pre-TF migration). | Удалить/переписать legacy-таблицу или пометить HISTORICAL. |
| **MEDIUM** | `docs/TESTING.md:209-216` (TC-AUTH) | Слабая привязка: «handler validation tests», «ConstantTimeCompare path», без `func Test…`. `TestVerifyMagicLink_*` в коде в основном rune-edge, не happy-path verify. | Колонка «Автотест» = точное имя; комментарий `// TC-AUTH-NN` над функцией. |
| **MEDIUM** | `testframework/stages/e2e/**` | `AS_ID` / allure labels — **0** совпадений. Матчинг TMS только по fullName хрупок. | Добавить `AS_ID` / tmsLink `HUB-TC-*` для specs, льющихся в Hub TMS. |
| **MEDIUM** | Cross-wave: `W2-tms.md`, `W2-project-dashboard.md` | Statement coverage ≥82% **не ловит** project-scope IDOR; denial-suite отсутствует. | Coverage-gate ≠ security acceptance; обязательны RBAC/IDOR suites (эталон `bases/rbac_test.go`). |
| **LOW** | `ci.yml` coverage-gate без `-race` | Осознанная экономия минут; race в `backend-test`. | Оставить; не смешивать с claim «race-clean coverage». |
| **LOW** | `COVERAGE.md` ai **78.9%**, bases **78.6%**, gitlab **68.5%** | Ниже overall floor по пакету; не ломают total 83%. | Бэклог raise-to-82 или явный ADR «v4/integration floor 50/70». |
| **LOW** | `a11y.spec.ts:53` | `page.locator(".login-form-animate")` — CSS-класс вместо role. | `getByRole` / стабильный testid. |
| **INFO** | coverage-gate skip PR | Сознательно (CI minutes ~3000). | OK; security package check можно добавить только на master вместе с overall. |

### CRITICAL/HIGH по product-багам кода

В этом слоте **не** искались IDOR в прод-коде (это W1/W2). Для coverage-слота: **0 CRITICAL** в test harness; **HIGH** — процесс/docs/flaky.

---

## Чек-лист G.1 (сжато)

| # | Критерий | Результат |
|---|----------|-----------|
| 1 | Детерминизм | **FAIL** — E2E `waitForTimeout`; Go Sleep PARTIAL |
| 2 | Race | Не переизмерялся; конфиг `backend-test` имеет `-race` |
| 3 | Привязка TC | PARTIAL — AUTH пробелы |
| 4 | Осмысленность | N/A (новых тестов в слоте нет); gap: нет IDOR-тестов при высоком % |
| 5 | Покрытие gate | Конфиг overall OK; 6×80% **не в CI**; % STALE |
| 6 | Границы/негатив | DRIFT vs W2 (IDOR не покрыты) |
| 7 | Triage | N/A (scan) |
| 8 | Время | Не замерено |
| 9 | Артефакты | Scan-only, dumps не создавались |
| 10 | Docs sync | **FAIL** — COVERAGE/AGENTS vs 499; TESTING legacy § |

---

## Команды для закрытия (следующий Agent-проход)

```bash
# Inventory
# expect 499
Get-ChildItem -Recurse backend/testframework -Filter '*_test.go' | Measure-Object

# Cover + security packages (Docker/test URLs)
cd backend
go test -tags hubtest -coverprofile="$env:TEMP/hub-cov.out" -covermode=atomic -coverpkg=./internal/... ./testframework/...
go tool cover -func="$env:TEMP/hub-cov.out" | Select-Object -Last 1
go tool cover -func="$env:TEMP/hub-cov.out" | Select-String "internal/(auth|ws|rbac|audit|calls|platform/middleware)"

# Flaky scan
rg "time\.Sleep\(" backend/testframework
rg "waitForTimeout" testframework/stages/e2e
```

---

## Handoff

- **docs-writer / team-lead:** синхронизировать 495→499; освежить COVERAGE % после cover-прогона.
- **devops-engineer:** CI step security 6×80% на master coverage-gate.
- **qa-auto2:** убрать `waitForTimeout` в `ai-hub.spec.ts`.
- **qa-auto1:** IDOR denial suites (после fix be*) — см. W2.
- **code-reviewer:** не требуется на этот docs-only scan.

**Регрессия:** не запускалась (scan-only).
**Вердикт:** `verified: false` — 2026-07-23 01:19 UTC+3
