---
name: coverage-audit
description: Матрица требований/код/тесты. Делегирует system-analyst + verifier. Создаёт/обновляет .planning/COVERAGE.md с полной картиной покрытия.
disable-model-invocation: true
---

# /coverage-audit — Coverage Audit

Создаёт матрицу: требование → реализован ли → есть ли тест.

## Step 1: Измерить фактическое test coverage

```bash
cd C:/Users/Valerii/Desktop/hub/backend
go test -race -count=1 -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -5   # total coverage
go tool cover -func=coverage.out | grep -E "0\.0%"  # zero coverage packages
```

```bash
cd C:/Users/Valerii/Desktop/hub/frontend
npx vitest run --coverage --reporter=json 2>/dev/null | \
  node -e "const d=JSON.parse(require('fs').readFileSync('/dev/stdin','utf8')); console.log('Coverage:', d.total?.lines?.pct + '%')"
```

## Step 2: Прочитать требования

```
Читай:
- .planning/PROJECT.md §Активные требования
- .planning/REQUIREMENTS.md
- .planning/ROADMAP.md §Success Criteria для каждой фазы
```

## Step 3: Трассировать требования к коду

Для каждого requirement ID найди:

```bash
# Например для AUTH-01 (Magic Link)
rg -n "MagicLink\|magic.link\|magic_link" backend/internal/auth/ --type go

# Для MSG-01 (personal chats)
rg -n "DirectMessage\|DM\|direct_message" backend/internal/messenger/ --type go
```

## Step 4: Проверить наличие тестов

```bash
# Есть ли тест для auth/magiclink.go?
ls backend/internal/auth/*_test.go

# Покрытие конкретного файла
go tool cover -func=coverage.out | grep "auth/magiclink"
```

## Step 5: Особые gaps из аудита

Всегда проверить статус 12 CRITICAL findings:
```
Читай: .planning/code-review-final/SUMMARY.md §CRITICAL
```

Для каждого: OPEN / FIXED / PARTIAL?

## Step 6: Создать/обновить COVERAGE.md

Создать `.planning/COVERAGE.md`:

```markdown
# Coverage Matrix — Hub

**Date:** YYYY-MM-DD
**Go test coverage:** X.X%
**Frontend coverage:** X.X%

## Summary
| Status | Count | % |
|--------|-------|---|
| ✅ Implemented + Tested | N | % |
| ⚠️ Implemented, No Tests | N | % |
| 🔴 Not Implemented | N | % |

## Critical Security Findings Status
| ID | Finding | Status | Fixed in |
|----|---------|--------|---------|
| SEC-01 | APIKeyAuth dead code | OPEN/FIXED | — |
...

## Requirements Matrix  
| ID | Requirement | Code Location | Tests | Status |
|----|-------------|---------------|-------|--------|
| AUTH-01 | Magic Link login | auth/magiclink.go | TestMagicLink | ✅ |

## Pending Tasks (Gaps)
| Priority | Gap | Est. Effort |
|----------|-----|-------------|
```

## Step 7: Commit

```bash
git add .planning/COVERAGE.md
git commit -m "chore(planning): update coverage matrix $(date +%Y-%m-%d)"
```
