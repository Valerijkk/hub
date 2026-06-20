# SA_VER — Верификатор System Analyst

## Роль

Независимая проверка технической спецификации.

## Чек-лист

- [x] Все User Stories из vision_and_scope покрыты технически
- [x] ERD корректна: UUID, version, UTC, soft delete
- [x] API-контракты непротиворечивы с architecture.md — **285 REST подтверждено**
- [x] Описание от SA соответствует спецификации — **counts исправлены 2026-06-17**

## Верификация

| Статус | Дата |
|--------|------|
| verified: true | 2026-06-17 |

## Лог проверок

| Дата | Артефакт | Результат | Замечания |
|------|----------|-----------|-----------|
| 2026-06-17 | system_specification.md init | pending | — |
| 2026-06-17 | Phase 0 SA audit (system_specification.md, COVERAGE.md) | **FAIL** | Handler routes 282≠290; total REST 285≠293; wiki 270 ✓; 6 module counts wrong |

### Phase 0 Reconciliation — 2026-06-17

**Method:** Independent recount via `RegisterRoutes` block parser on all handler files + cross-check `docs/wiki/API.md` table rows + spot-check CRITICAL fixes in code.

**API counts (verified):**

| Источник | SA claimed | Verified | Status |
|----------|------------|----------|--------|
| `docs/wiki/API.md` | 270 | **270** | ✅ |
| Handler `RegisterRoutes` | 290 | **282** | ❌ (−8) |
| Infra (`/docs`, `/docs/doc.json`, `/health`) | 3 | **3** | ✅ |
| **Total REST `/api/*`** | 293 | **285** | ❌ (−8) |
| WebSocket `GET /ws` | 1 | **1** | ✅ |
| Dev `/api/dev/*` | 3 | **3** | ✅ (non-release only) |

**Per-module route errors (SA → verified):**

| Module | SA | Verified |
|--------|-----|----------|
| search | 7 | **5** |
| notification | 11 | **10** |
| dashboard | 8 | **7** |
| github | 11 | **9** |
| gitlab | 9 | **7** |
| calls | 22 | **21** |

**NFR (NFR-01..30):** PASS with minor notes — mandatory constraints from `.cursor/rules/60-tz-constraints.md` covered; optional infra/frontend stack constraints (Fiber WS, Caddy) not listed (acceptable).

**Gap matrix:** PASS (reasonable) — Top-20 mapping, CRITICAL reconciliation, test baseline (42 files, 11.5%, 0 Vitest, no Playwright) verified; SEC-01/02/04/05/06/07, INF-02/04, FE-01 fixes confirmed in code.

**Action for SA:** ~~Correct totals~~ **DONE** — 282 + 3 = 285 REST, module breakdown updated, wiki drift ~15.
