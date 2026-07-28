# Phase 132 — Backend units colocated (`internal/`)

**Status:** PLANNED  
**Migration stage:** Э3  
**Estimate:** 10–15 person-days (волнами по доменам)  
**Branch (suggested):** `feat/qa/tf-backend-units-colocated`

## Goal

Раскидать ~508 Go-тестов из `backend/testframework/**` в
`backend/internal/<domain>/` (рядом с кодом). Убрать ненужные `export_hooks.go`
там, где тесты в том же пакете получают доступ к unexported API.

Источник решения: [`02-TEST-FRAMEWORK-MIGRATION.md`](../../audits/planning-audit-2026-07/02-TEST-FRAMEWORK-MIGRATION.md) §Э3

## Depends on

- **128** (мягкая — для будущего integration extract)
- **A3 coverage baseline** (hard для начала любой миграции; сравнить cov до/после)

## Deliverables

- [ ] Попакетный перенос: auth → tracker → … (чеклист доменов)
- [ ] `go test ./internal/...` green (`-tags hubtest` только где ещё нужен)
- [ ] Удаление опустевших `backend/testframework/<pkg>` по мере переноса
- [ ] Сценарии «с БД/NATS» помечены кандидатами в `tests/integration` (phase 134)
- [ ] Coverage delta vs A3 baseline задокументирован

## Out of scope

- Frontend Vitest
- Playwright
- Полный CI cutover

## Success criteria

- [ ] `backend/testframework/` пуст или только re-export stubs на время перехода
- [ ] Coverage overall не ниже floor (открытый вопрос: 82% — см. миграцию)
- [ ] Race: `go test -race` на затронутых пакетах green

## Risks

| Risk | Mitigation |
|------|------------|
| Потеря покрытия | A3 baseline + per-package cover diff |
| hubtest/export_hooks | Удалять hooks только после green same-package tests |
| Огромный diff | Волны по 2–3 домена / PR |
