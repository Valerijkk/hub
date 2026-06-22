# frontend-test

CI job: `frontend-test` · Stage: `unit`

Vitest unit-тесты по доменам функционала.

```bash
cd frontend && node scripts/link-testframework-node-modules.mjs && npm run test:run

# один домен
cd frontend && npx vitest run ../testframework/stages/unit/frontend-test/domains/auth
```

Манифест: [`../../pipeline/jobs/frontend-test.yaml`](../../pipeline/jobs/frontend-test.yaml)

| Домен | Что тестирует |
|-------|---------------|
| `auth` | login, session, client API |
| `admin` | audit, 1C, metrics |
| `ai` | model picker, store, privacy |
| `tracker` | issue card/form, user picker |
| `messenger` | channel icon |
| `tms` | case table row |
| `dashboard` | system health widget |
| `platform` | UI shell, i18n, lib helpers |
| `v4` | super-app modules (CRM, HR, intranet, …) |
