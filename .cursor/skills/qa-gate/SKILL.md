---
name: qa-gate
description: Полный чек качества перед коммитом — lint + typecheck + unit тесты + API smoke + security scan. Запускай перед каждым PR.
disable-model-invocation: true
---

# /qa-gate — Pre-commit Quality Gate

Запусти все проверки последовательно. Останови при первой FAIL.

## Step 1: Go Build + Vet

```bash
cd backend
go build -tags release ./...
go vet ./...
```

## Step 2: Go Tests (race detector)

```bash
cd backend
go test -race -count=1 -timeout=120s ./...
```

Expected: `ok` для всех пакетов. Если `FAIL` — **не чинить сразу**. См. `.cursor/skills/test-triage/SKILL.md`:
1. Воспроизвести функционал по docs
2. Код верен → fix test; код неверен → fix code
3. Повторить тест только после triage

## Step 3: TypeScript Check

```bash
cd frontend
npx tsc --noEmit
```

Expected: 0 errors.

## Step 4: Frontend Lint

```bash
cd frontend
npm run lint
```

## Step 5: Frontend Tests

```bash
cd frontend
npx vitest run
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

## Результат

```
✅ GATE PASSED — ready to commit
❌ GATE FAILED at step N: <причина>
```
