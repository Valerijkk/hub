# 05 — E2E Playwright

> Часть [Testing Guide](../TESTING_GUIDE.md). POM-канон: [02-page-object.md](02-page-object.md).  
> Playbook: [`.cursor/playbooks/testing.md`](../../.cursor/playbooks/testing.md) §6.  
> Корень: `testframework/stages/e2e/`.

---

## Карта

```
stages/e2e/
├── playwright.config.ts
├── shared/
│   ├── fixtures.ts          # skipUnless*, extend test
│   ├── pages/               # POM
│   ├── steps/               # бизнес-шаги
│   └── helpers/             # auth, backend, frontend, a11y, ai
└── domains/
    ├── auth/                # login, logout, onboarding — qa-auto1
    ├── navigation/          # qa-auto1
    ├── tracker/             # qa-auto1
    ├── tms/                 # qa-auto1
    ├── messenger/           # qa-auto1
    ├── admin/               # qa-auto2
    ├── ai/                  # qa-auto2
    ├── platform/            # qa-auto2 (health, a11y, critical-gate, …)
    └── v4/                  # qa-auto2
```

**Smoke vs full:** `e2e-smoke` ≈ 3 spec (auth + platform gate); `e2e-full` ≈ 20 spec. Не раздувать smoke unit-логикой.

---

## Архитектура сценария

```
fixtures (backendAvailable…) 
  → AuthSteps / NavigationSteps 
    → LoginPage / AppShellPage 
      → expect (web-first)
```

Spec **не** содержит CSS-селекторов страницы — только Steps/POM.

---

## Fixtures и skip

Реальный код: `shared/fixtures.ts`.

```typescript
export const test = base.extend<HubFixtures>({
	backendAvailable: async ({}, use) => { await use(await isBackendReachable()); },
	devModeAvailable: async ({}, use) => { await use(await isDevModeAvailable()); },
	frontendAvailable: async ({}, use) => { await use(await isFrontendReachable()); },
});

export function skipUnlessBackend(backendAvailable: boolean): void {
	test.skip(!backendAvailable, "Backend not running — start with: docker compose -f docker-compose.local.yml up -d");
}
```

| Good | Bad |
|------|-----|
| `skipUnlessDevMode` с понятной причиной | `test.skip(true)` без текста |
| Dev-login через helpers (`devLogin` + `authenticatePage`) | Хардкод JWT в spec |
| Уникальные имена сущностей (`Date.now().toString(36)`) | Фиксированный `project-1` при parallel workers |

---

## Ожидания — без sleep

```typescript
// Good
await expect(page).toHaveURL(/\/projects/, { timeout: 15_000 });
await expect(loc).toBeVisible();

// Bad
await page.waitForTimeout(3000);
await new Promise((r) => setTimeout(r, 1000));
```

Playwright auto-wait + web-first assertions. Таймаут — параметр `expect`, не sleep.

---

## Auth в E2E

1. **UI dev-panel** — `AuthSteps.devLoginAsAdminUI()` (нужны backend + `DEV_MODE`).
2. **API seed + inject session** — `loginAsAdmin` / `loginAsDev` через `helpers/auth.ts` (быстрее для большинства сценариев).

Не тестировать Magic Link end-to-end с реальным SMTP в каждом spec — это integration/backend + точечный E2E branding/redirect.

---

## Allure в E2E

```typescript
import * as allure from "allure-js-commons";

await allure.epic("Auth");
await allure.feature("Login");
await allure.severity("critical");
await allure.label("AS_ID", "HUB-TC-AUTH-01");
```

Бизнес-шаги — `test.step(…)` (конвертируется allure-playwright). Секреты в attachment запрещены. Playbook §8.

---

## Изоляция и parallel

- Каждый тест сам логинится / чистит данные, которые создал.
- Не шарить один project/channel между workers.
- `test.describe.configure({ mode: "parallel" })` — только при независимых данных.

---

## Calls / LiveKit

Голосовые/видео E2E и backend-тесты calls остаются на **LiveKit**. Не писать mediasoup-сценарии (миграция SFU отклонена).

---

## Прогон

```bash
cd frontend && npm run test:e2e:smoke
cd frontend && npm run test:e2e
```

Конфиг: `stages/e2e/playwright.config.ts`. Backend: `docker compose -f docker-compose.local.yml up -d`.

---

## Чек-лист E2E

- [ ] Spec в своём `domains/{domain}/`, владелец qa-auto1/2 соблюдён
- [ ] POM + Steps; нет сырых селекторов в spec
- [ ] Нет `waitForTimeout`; есть `skipUnless*` где нужно
- [ ] Уникальные данные; стабилен при повторном прогоне
- [ ] TC-ID + Allure `AS_ID` при TMS; колонка в `TESTING.md`
- [ ] Не дублирует то, что уже закрыто unit/integration
