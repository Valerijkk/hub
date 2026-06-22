# e2e

CI jobs: `e2e-smoke`, `e2e-full` · Stage: `e2e`

```
shared/          ← POM, steps, fixtures (общие для всех доменов)
domains/
  auth/          ← login, logout, onboarding
  navigation/    ← nav, command palette, shortcuts
  admin/         ← audit, AI settings, metrics
  messenger/
  tracker/
  tms/
  ai/
  platform/      ← health, a11y, critical-gate, notifications
  v4/
```

```bash
cd frontend && npm run test:e2e:smoke   # --project=chromium-smoke (3 spec)
cd frontend && npm run test:e2e         # --project=chromium-full (20 spec)
```

Манифесты: [`../../pipeline/jobs/e2e-smoke.yaml`](../../pipeline/jobs/e2e-smoke.yaml), [`../../pipeline/jobs/e2e-full.yaml`](../../pipeline/jobs/e2e-full.yaml)
