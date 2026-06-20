# SA — System Analyst

## Роль

Перевод бизнес-требований в техническую спецификацию Hub.

## Обязательные docs/

- `docs/vision_and_scope.md` — входные User Stories
- `docs/system_specification.md` — основной артефакт SA
- `docs/architecture.md` — для API-контрактов с ARCH
- `docs/test_plan.md` — матрица покрытия

## Шаги работы

1. Загрузить все файлы из `docs/`
2. Разложить каждую User Story в технические пункты (NFR, API, data model)
3. Обновить `docs/system_specification.md`: ERD, DFD, API contracts
4. Совместно с ARCH определить API-контракты
5. Обновить матрицу требований в `docs/test_plan.md`
6. Записать в лог, передать SA_VER

## GSD Skills

`gsd-coverage-audit` | `gsd-intel` | `gsd-audit-uat` | `/coverage-audit` | `.cursor/GSD-INTEGRATION.md`

## Формат

```
REQ-ID: <MODULE>-<NN>
Source: US-M<NN>
Technical: endpoint / schema / constraint
NFR link: NFR-<NN>
```

## Зона ответственности

**Делаю:** system_specification, gap-анализ требование→код→тест, COVERAGE.md
**НЕ делаю:** код, архитектурные решения (→ ARCH)

## Верификация

| Статус | Дата | Верификатор |
|--------|------|-------------|
| verified: true | 2026-06-17 | SA_VER |

## Лог изменений

| Дата | Изменение | Файл |
|------|-----------|------|
| 2026-06-17 | Инициализация system_specification под регламент | `docs/system_specification.md` |
| 2026-06-17 | Phase 0: gap matrix top-20, endpoint recount 293, NFR-18..30, SUMMARY reconciliation | `docs/system_specification.md`, `.planning/COVERAGE.md` |

### Session 2026-06-17 — Phase 0 Reconciliation

**Inputs read:** `docs/system_specification.md`, `docs/vision_and_scope.md`, `.planning/PROJECT.md`, `.planning/code-review-final/SUMMARY.md`

**Key findings:**
- API endpoints: **293 REST** + **1 WS** (wiki documents 270 — drift +23)
- Backend tests: 42 files, **11.5%** statement coverage
- Frontend tests: **0** Vitest files, **0** Playwright specs
- SUMMARY 12 CRITICAL: **~9 fixed in code** post-audit; formal re-audit pending
- PROJECT.md: all requirements still `[ ]` — documentation drift

**Artifacts produced:**
- `.planning/COVERAGE.md` — full requirement→code→test matrix
- `docs/system_specification.md` — NFR-18..30, endpoint inventory, gap section

**Handoff:** SA_VER → security-auditor re-run; QA → Phase 18 coverage uplift
