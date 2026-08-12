# E3 — Playwright TS → Python: per-spec plan

**Дата:** 2026-07-27  
**Источник:** [`02-TEST-FRAMEWORK-MIGRATION.md`](02-TEST-FRAMEWORK-MIGRATION.md) §Э2  
**Вход:** `testframework/stages/e2e/domains/**/*.spec.ts` (20 файлов)  
**Целевая папка (после Э1):** `/tests/e2e/` — **не создавать в этой задаче**

---

## Классификация

| Класс | Смысл | Типичные маркеры |
|-------|--------|------------------|
| **mechanical** | Перенос 1:1 после порта shared-слоя: `goto` / `get_by_*` / `expect` / клики | Локаторы + короткие timeout; auth только через `loginAsDev`/`loginAsAdmin` |
| **rewrite** | Нужна пересборка на Python API / другие фикстуры | `wait_for_request`, мутация session, axe-core, API+UI mix, длинные async waits, сложный multi-step |

**Правило порядка:** сначала shared (fixtures / auth / POM / a11y), потом mechanical specs, потом rewrite. Паритет сценариев обязателен; имена test-case сохраняем.

---

## Shared layer (обязательный rewrite до specs)

| TS-источник | Python-цель (план) | Почему rewrite |
|-------------|-------------------|----------------|
| `shared/fixtures.ts` | `tests/conftest.py` + `tests/e2e/conftest.py` | `base.extend` → pytest fixtures; `skipUnless*` → `pytest.mark.skipif` |
| `shared/helpers/auth.ts` | `tests/fixtures/auth.py` | `fetch` + `addInitScript` → `httpx` + `page.add_init_script`; seed `/api/dev/seed` |
| `shared/helpers/{backend,frontend}.ts` | `tests/fixtures/env.py` | Reachability probes для skip |
| `shared/helpers/a11y.ts` | `tests/e2e/helpers/a11y.py` | `@axe-core/playwright` → `axe-playwright-python` / inject axe |
| `shared/helpers/ai.ts` | `tests/e2e/helpers/ai.py` | Bearer API helpers |
| `shared/pages/*.ts` | `tests/e2e/pages/` | POM: механический порт API, но классы пишутся заново |
| `shared/steps/auth.steps.ts` | `tests/e2e/steps/auth.py` | Обёртка над fixtures |
| `playwright.config.ts` | `pytest.ini` / `pyproject` + `playwright` install | baseURL, webServer, timeouts |

Smoke CI сегодня: `login` + `a11y` + `critical-gate` (3 spec). После миграции тот же набор должен быть зелёным первым.

---

## Per-spec matrix (20)

| # | Spec | Domain | Tests (approx) | Class | Drivers | Notes |
|---|------|--------|----------------:|-------|---------|-------|
| 1 | `auth/login.spec.ts` | auth | 3 | **mechanical** | — | Branding + UI dev-login + sessionStorage assert; зависит от AuthSteps |
| 2 | `auth/logout.spec.ts` | auth | 1 | **rewrite** | **waits** | `page.waitForRequest` на `POST /api/auth/logout` — в Python `expect_request` / `wait_for_event` |
| 3 | `auth/onboarding-profile.spec.ts` | auth | 1 | **rewrite** | **auth** | `injectSession` с подменой `onboarding_completed`; form enable/disable |
| 4 | `platform/health.spec.ts` | platform | 3 | **mechanical** | — | Почти API-only (`request` / httpx); браузер минимален |
| 5 | `platform/metrics.spec.ts` | platform | 1 | **mechanical** | — | `GET /metrics` Prometheus text |
| 6 | `platform/notifications.spec.ts` | platform | 1 | **mechanical** | — | Settings UI после `loginAsDev` |
| 7 | `platform/critical-gate.spec.ts` | platform | 5 | **mechanical** | — | Smoke 5 paths; auth + shell + health |
| 8 | `platform/a11y.spec.ts` | platform | 5 | **rewrite** | **fixtures** | axe-core scan + exclude login badges; smoke-критичен |
| 9 | `admin/admin-ai.spec.ts` | admin | 2 | **mechanical** | — | Admin tab + non-admin redirect |
| 10 | `admin/admin-audit.spec.ts` | admin | 3 | **mechanical** | — | Audit table + filter options + RBAC redirect |
| 11 | `admin/admin-metrics.spec.ts` | admin | 1 | **mechanical** | — | Metrics tab load |
| 12 | `ai/ai-hub.spec.ts` | ai | 6 | **rewrite** | **auth + waits** | Mix API (`devLogin`) + UI DM `@hub-ai`; bot reply timeout до 45s |
| 13 | `ai/automation.spec.ts` | ai | 1 | **mechanical** | — | Workflow settings empty-state; нужен seeded `projectId` |
| 14 | `messenger/messenger.spec.ts` | messenger | 1 | **mechanical** | — | POM `MessengerPage`; composer timeout 30s — порт POM, сценарий простой |
| 15 | `navigation/navigation.spec.ts` | navigation | 2 | **mechanical** | — | Sidebar + unauth redirect |
| 16 | `navigation/shortcuts.spec.ts` | navigation | 1 | **mechanical** | — | `?` dialog |
| 17 | `navigation/command-palette.spec.ts` | navigation | 2 | **mechanical** | — | Ctrl+K + action navigation |
| 18 | `tms/tms.spec.ts` | tms | 1 | **mechanical** | — | Create case in DEMO project |
| 19 | `tracker/project-lifecycle.spec.ts` | tracker | 1 | **rewrite** | **waits** | Create → board → trash → restore; много URL/timeout 15–30s |
| 20 | `v4/v4-superapp.spec.ts` | v4 | 1 | **mechanical** | — | Loop routes + headings (intranet/crm/hr/…) |

### Сводка

| Class | Count | Specs |
|-------|------:|-------|
| mechanical | **15** | login, health, metrics, notifications, critical-gate, admin-ai, admin-audit, admin-metrics, automation, messenger, navigation, shortcuts, command-palette, tms, v4 |
| rewrite | **5** | logout, onboarding-profile, a11y, ai-hub, project-lifecycle |

mechanical = #1,4–7,9–11,13–18,20; rewrite = #2,3,8,12,19. Итого 20.

---

## Порядок переноса (рекомендуемый)

```
Wave A — shared rewrite
  fixtures → auth/env → POM/steps → a11y helper

Wave B — smoke mechanical (+ critical parity)
  login → critical-gate → health → metrics

Wave C — domain mechanical
  navigation×3 → admin×3 → notifications → automation → messenger → tms → v4

Wave D — rewrite
  logout → onboarding-profile → a11y → project-lifecycle → ai-hub
```

**Критерий паритета Э2:** каждый TS-сценарий имеет Python-аналог с тем же intent; `e2e-smoke` 3/3; `e2e-full` 20/20; затем удаление `testframework/stages/e2e` node_modules / TS specs (в фазе CI cleanup).

---

## Риски именно E2E-порта

| Риск | Митигация |
|------|-----------|
| Auth через sessionStorage Zustand persist | Один канонический `authenticate_page()` в fixtures; покрыть login+logout regression |
| axe API drift | Зафиксировать версию axe + те же impact filters (`critical`/`serious` как в TS) |
| Flaky long waits (ai-hub 45s, messenger 30s) | Явные `expect(...).to_be_visible(timeout=)` + retry mark; не `time.sleep` |
| Seeded DEMO `projectId` | Фабрика в `/tests/fixtures` или parse из UI после seed (как сейчас) |
| Два раннера в переходный период | CI: старый TS job + новый Python job до cleanup; паритет по имени сценария |

---

## Mapping → GSD phases

См. [`TF-MIGRATE-E6-PHASES.md`](TF-MIGRATE-E6-PHASES.md):

- **129** — shared E2E layer  
- **130** — mechanical 15 specs  
- **131** — rewrite 5 specs  

Зависит от **128** (`/tests` scaffold). Hard gate на старт всей миграции: **A3 coverage baseline** ([`01-PLANNING-FIXES.md`](01-PLANNING-FIXES.md) §A3).
