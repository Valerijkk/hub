# TASK-105 — 📐 Docs G-07: wiki KPI planned vs shipped

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P1** · Статус: Закрыта  
> Источник: A2W11-5-001 · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008

## [OUTCOME]

vision_and_scope / wiki KPI G-07 не врёт: planned vs shipped согласованы с PO.

## [ROLE]

- Role: product→docs-writer
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - docs/vision_and_scope.md
  - docs/wiki/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - менять product code
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: **N/A для кода** (docs-only). PO reconcile выполнен product→docs-writer 2026-07-28 без отдельного ADR (нет кодовых/migrate изменений).

## [REFERENCES]

- Finding IDs: A2W11-5-001 (1)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - PO reconcile записан
  - docs updated
  - Каждый ID из списка — FIXED с evidence (тест или docs+PO)
- Stop when: green + статус «Исправлена» в README

## [OUTPUT SCHEMA]

- Checklist ID → evidence
- Test commands + PASS
- Mutation note (для code-фиксов)
- README status update

## [OPERATING RULES]

- Не оставлять ID «WONTFIX» без явного owner cancel в задаче
- ADR-first: код только после approve в Approval record

---

## PO reconcile (2026-07-28)

| Решение | Содержание |
|---------|------------|
| Вердикт G-07 | **Partial** — не done |
| Shipped | Базовый Intranet + Calendar (v4): `backend/internal/{intranet,calendar}`, FE routes `/intranet`, `/calendar`, modules `intranet`/`calendar` |
| Planned | US-INTR/US-CAL v2 enhancements (KPI секций — частично/планируется); **весь US-WIKI-01..05** и product wiki-модуль |
| Evidence кода | Нет `backend/internal/wiki`; нет FE module/routes `wiki`; `docs/wiki/API.md` секция Wiki — **planned/spec** |
| Anti-drift | `docs/wiki/` = документация репозитория ≠ продуктовый wiki (spaces/pages) |
| Finding | **A2W11-5-001 → FIXED** (docs+PO) |

## Evidence checklist

| ID | Статус | Evidence |
|----|--------|----------|
| A2W11-5-001 | **FIXED** | `docs/vision_and_scope.md` G-07 → Partial; KPI Wiki остаются «планируется» + PO note; `docs/wiki/API.md` + `HOME.md` уточняют planned vs docs-wiki |

## Tests / mutation

- N/A: docs-only; product code не менялся.
- Проверка факта (shell): `Test-Path backend/internal/wiki` → `False`; modules FE без `wiki`.

## README

Статус в `.planning/backlog/README.md` → **Исправлена**.
