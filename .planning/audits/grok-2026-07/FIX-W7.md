# FIX-W7 — Authz remainder A1-005/006, A7-001, A4-003/004, A9-012

**Ветка:** `cursor-grok-4.5`  
**Дата:** 2026-07-27  
**Предшественник:** FIX-W6

## Закрыто

| ID | SHA | Фикс |
|----|-----|------|
| A1-005 | `428865c` | founder в system-admin bypass (`rbac/service`) |
| A1-006 | `1d92f12` | founder в `RequireAdmin` |
| A4-003 | `5d2e590` | CreateSavedFilter → membership |
| A4-004 | `4f9b2bb` | ListSavedFilters → membership |
| A7-001 | `c5edd70` | dashboard project stats → membership |
| A9-012 | `c7a4589` | RequireAuth принимает `userID` от API key (*сообщение коммита ошибочно говорит A1-005; содержимое — A9-012*) |

Мутации: снять membership/bypass/early-return → denial-тесты краснеют.

## Baseline/after
См. общий прогон W6+W7 (`w67-after-hubtest.txt`).

*Next: W8.*
