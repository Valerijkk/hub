---
name: coverage-audit
description: Матрица требований/код/тесты. Делегирует sa1/sa2 + qa-auto-ver. Создаёт/обновляет .planning/COVERAGE.md с полной картиной покрытия.
disable-model-invocation: true
---

# /coverage-audit — Coverage Audit

Создаёт матрицу: требование → реализован ли → есть ли тест.

## Step 1: Измерить фактическое test coverage

> **Важно:** внутри `backend/internal/` тестов больше нет (миграция Wave B завершена) — все юниты и интеграционные тесты живут в `backend/testframework/{domain}/`. Профиль cover надо снимать по этому пути, иначе получишь 0% ложно.

```bash
cd C:/Users/Valerii/Desktop/hub/backend
go test -tags hubtest -race -count=1 -coverpkg=./internal/... -coverprofile="${TMPDIR:-/tmp}/hub-cov.out" ./testframework/...
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | Select-String -Pattern "total:"   # total coverage (PowerShell)
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | Select-String -Pattern "0\.0%"    # zero coverage packages
```

Профиль писать **вне репозитория** (`${TMPDIR}` / системный temp) — правило `.cursor/rules/45-artifacts-cleanup.mdc`, никогда `coverage.out` в `backend/`.

Домен-специфичный прогон (быстрее для точечного аудита):

```bash
go test -tags hubtest -race -count=1 -coverpkg=./internal/tracker/... -coverprofile="${TMPDIR:-/tmp}/tracker-cov.out" ./testframework/tracker/...
go test -tags hubtest -race -count=1 -coverpkg=./internal/bases/...   -coverprofile="${TMPDIR:-/tmp}/bases-cov.out"   ./testframework/bases/...
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
- .planning/archive/july-2026/ROADMAP-FINAL-2026-07-03.md §Success Criteria (актуальные волны)
- .planning/decisions/specs/*.md — если модуль в стадии спеки (Intranet/Calendar/AI Provider Hub/Notif)
```

## Step 3: Трассировать требования к коду

Для каждого requirement ID найди:

```bash
# AUTH-01 (Magic Link)
rg -n "MagicLink|magic.link|magic_link" backend/internal/auth/ --type go

# MSG-01 (personal chats)
rg -n "DirectMessage|DM|direct_message" backend/internal/messenger/ --type go

# TRACKER-AUTOMATION (JQL automation, фаза 17)
rg -n "AutomationEngine|automation_engine" backend/internal/tracker/ --type go

# BASES-RBAC (no-code DB row-level RBAC)
rg -n "RBAC|role.*guard" backend/internal/bases/ --type go

# CAL-SYNC (calendar_event ERD, если фаза уже в коде)
rg -n "calendar_event|CalendarEvent" backend/ent/schema/ backend/internal/ --type go
```

## Step 4: Проверить наличие тестов

```bash
# Тесты домена auth (в testframework/, не в internal/)
ls backend/testframework/auth/*_test.go

# Тесты домена tracker (много волновых файлов service_waveNNN_test.go)
ls backend/testframework/tracker/*_test.go

# Покрытие конкретного пакета из снятого профиля
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | Select-String -Pattern "internal/auth/magiclink"
go tool cover -func="${TMPDIR:-/tmp}/hub-cov.out" | Select-String -Pattern "internal/bases"

# Frontend: домен-специфичные unit-тесты
ls testframework/stages/unit/frontend-test/domains/tracker/
ls testframework/stages/unit/frontend-test/domains/v4/   # bases, crm, hr, accounting

# E2E: покрыт ли домен хотя бы одним spec
ls testframework/stages/e2e/domains/tracker/
```

## Step 5: Особые gaps из аудита

Всегда проверить статус 12 CRITICAL findings:
```
Читай: .planning/audits/code-review-final/SUMMARY.md §CRITICAL
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

## Step 7: Cleanup + Commit

Удалить временные cover-профили (правило `.cursor/rules/45-artifacts-cleanup.mdc` — они писались вне репо, но проверить `git status` на всякий случай):

```bash
git status  # убедиться, что нет cover*/cov_*/*.out/tmp_* в diff
```

```bash
git add .planning/COVERAGE.md
git commit -m "chore(planning): update coverage matrix $(date +%Y-%m-%d)"
```

## Пример итоговой записи (реальные цифры Hub, ориентир)

```markdown
## Requirements Matrix (фрагмент)
| ID | Requirement | Code Location | Tests | Status |
|----|-------------|----------------|-------|--------|
| AUTH-01 | Magic Link login | auth/magiclink.go | testframework/auth/service_test.go | ✅ |
| TRACKER-WIP | WIP-лимиты Kanban | tracker/service_boards.go | testframework/tracker/service_wave121_test.go | ✅ |
| BASES-RBAC | Row-level RBAC | bases/service.go | testframework/bases/rbac_test.go | ✅ |
| CAL-SYNC | Calendar/email sync | — (planning) | — | 🔴 not implemented (seed SEED-002) |
```
