---
name: qa-gate
description: Полный чек качества перед коммитом — lint + typecheck + unit тесты + API smoke + security scan. Запускай перед каждым PR.
disable-model-invocation: true
---

# /qa-gate — Pre-commit Quality Gate

Запусти все проверки последовательно. Останови при первой FAIL. При FAIL — не чини сразу: примени `.cursor/skills/test-triage/SKILL.md`, затем повтори **с начала упавшего шага** (rule `56-regression-verify.mdc`, цикл до green).

> **Два режима:** быстрый локальный (Step 1-2, разработка) и **канонический Docker `-tags hubtest`** (Step 2a, обязателен перед handoff/PR — в `backend/testframework/` 495+ тестов, в `internal/` — 0).

## Step 1: Go Build + Vet

```bash
cd backend
go build -tags release ./...
go vet ./...
```

## Step 2: Go Tests локально (race detector, быстрый цикл)

```bash
cd backend
go test -race -count=1 -timeout=120s ./...
```

Expected: `ok` для всех пакетов (в основном `internal/` без бизнес-тестов — они в `testframework/`).

## Step 2a: Go Tests canonical — Docker `-tags hubtest` (обязателен перед handoff)

```bash
docker compose -f deployments/docker-compose.test.yml up -d
docker run --rm --network deployments_default \
  -v "${PWD}/backend:/app" -w /app \
  -e POSTGRES_TEST_URL="postgresql://hub:hub@postgres:5432/hub?sslmode=disable" \
  -e MONGODB_TEST_URL="mongodb://hub:hub@mongo:27017/?authSource=admin" \
  golang:1.25-bookworm \
  sh -c "go test -tags hubtest -race -count=1 -timeout=600s ./testframework/..."
```

Точечный прогон одного домена (быстрее полного, для итерации на фиче — из той же Docker-сессии):

```bash
go test -tags hubtest -race -count=1 ./testframework/auth/...
go test -tags hubtest -race -count=1 ./testframework/tracker/...
go test -tags hubtest -race -count=1 ./testframework/bases/...
go test -tags hubtest -race -count=1 ./testframework/tms/...
go test -tags hubtest -race -count=1 ./testframework/messenger/...
go test -tags hubtest -race -count=1 ./testframework/ws/...
go test -tags hubtest -race -count=1 ./testframework/calls/...
go test -tags hubtest -race -count=1 ./testframework/audit/...
```

Expected: `ok` для всех пакетов. Если `FAIL` — **не чинить сразу**. См. `.cursor/skills/test-triage/SKILL.md`:
1. Воспроизвести функционал по docs
2. Код верен → fix test; код неверен → fix code
3. Повторить тест только после triage (домен + regression)

## Step 3: TypeScript Check

```bash
cd frontend
node scripts/link-testframework-node-modules.mjs
npx tsc --noEmit
```

Expected: 0 errors.

## Step 4: Frontend Lint

```bash
cd frontend
npm run lint
```

## Step 5: Frontend Tests (vitest)

```bash
cd frontend
npx vitest run
```

Точечный домен (реальные пути `testframework/stages/unit/frontend-test/domains/`):

```bash
npx vitest run testframework/stages/unit/frontend-test/domains/auth/
npx vitest run testframework/stages/unit/frontend-test/domains/tracker/
npx vitest run testframework/stages/unit/frontend-test/domains/tms/
npx vitest run testframework/stages/unit/frontend-test/domains/v4/
npx vitest run testframework/stages/unit/frontend-test/domains/ai/
npx vitest run testframework/stages/unit/frontend-test/domains/platform/
npx vitest run testframework/stages/unit/frontend-test/domains/admin/
npx vitest run testframework/stages/unit/frontend-test/domains/messenger/
npx vitest run testframework/stages/unit/frontend-test/domains/dashboard/
```

При FAIL — `.cursor/skills/test-triage/SKILL.md` (сначала UI/контракт, потом тест).

## Step 6: Security Scan (secrets in staged files)

```bash
git diff --staged --name-only | xargs grep -l -E "(secret|password|devkey|api_key|token)" 2>/dev/null
```

Если нашёл файлы — проверить вручную что это не hardcoded credentials.

Также проверить:
```bash
git diff --staged | grep -E "DEV_MODE=true|devkey: secret|livekit_api_secret"
```

## Step 7: API Smoke (если есть запущенный сервер)

```bash
curl -sf http://localhost:3000/api/health | jq .status
```

Expected: `"ok"`

## Step 8: E2E smoke (CI job `e2e-smoke`, 3 spec, обязателен на PR)

```bash
npx playwright test testframework/stages/e2e/domains/auth/login.spec.ts \
  testframework/stages/e2e/domains/auth/logout.spec.ts \
  testframework/stages/e2e/domains/platform/health.spec.ts
```

Expected: `3 passed`.

## Step 9: E2E full (CI job `e2e-full`, 9 доменов ~20 spec — перед merge волны/релизом, не на каждый PR)

```bash
npx playwright test testframework/stages/e2e/domains/
```

Домены: `auth` (login, logout, onboarding-profile), `tracker` (project-lifecycle), `tms` (tms), `messenger` (messenger), `ai` (ai-hub, automation), `admin` (admin-ai, admin-audit, admin-metrics), `navigation` (command-palette, shortcuts, navigation), `platform` (health, metrics, notifications, a11y, critical-gate), `v4` (v4-superapp).

## Пример прогона для фичи в одном домене (tracker automation)

```bash
# 1. Точечный backend
go test -tags hubtest -race -count=1 ./testframework/tracker/... -run TestAutomationEngine

# 2. Полный домен backend (регрессия)
go test -tags hubtest -race -count=1 ./testframework/tracker/...

# 3. Смежный домен (automation публикует события, читает dashboard)
go test -tags hubtest -race -count=1 ./testframework/dashboard/...

# 4. Frontend, если менялся UI автоматизации
npx vitest run testframework/stages/unit/frontend-test/domains/tracker/

# 5. e2e-smoke перед PR, e2e-full перед merge волны
```

## Результат

```
✅ Step 1: build+vet — PASS
✅ Step 2a: hubtest race (домен + full) — PASS
✅ Step 3-5: tsc+lint+vitest — PASS
✅ Step 6: security scan — clean
✅ Step 8: e2e-smoke (3/3) — PASS
❌ GATE FAILED at step N: <причина> → test-triage → fix → повтор с начала упавшего шага
```

## Связанные документы

- Цикличное качество (правило + пример): `.cursor/REGULATION.md` §15-16
- Test triage: `.cursor/skills/test-triage/SKILL.md`
- GSD-эквивалент: `.cursor/GSD-INTEGRATION.md` §Hub slash-skills vs GSD skills
