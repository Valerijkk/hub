# TASK-123 — 🧹 FE lib dead export + docs drift NF-11

> Заведено: 2026-07-28 · Автор: team-lead · Приоритет: **P2** · Статус: **Закрыта**  
> Источник: A2W11 residual · **Owner lock: правим ВСЕ 254 A2 open** · TASK-008  
> Закрыта: 2026-07-28 (code) · Mutation PROVEN 2026-07-30 W8-fe1 · Исполнитель: fe2+docs · ветка `fix/frontend/task-123-dead-export-docs`

## [OUTCOME]

useRooms удалён; BUSINESS.md webcam consistency; NF-11 + REST count headers синхронизированы.

## [ROLE]

- Role: fe2+docs
- Responsibility: закрыть **все** перечисленные finding ID + тесты/мутация
- Handoff to: *-ver → code-reviewer

## [SCOPE AND CONSTRAINTS]

- Files allowed:
  - frontend/src/**
  - docs/**
  - `.planning/backlog/README.md` + этот файл
  - audit refs under `.planning/audits/total-2026-07/`
- Do not touch:
  - ломать livekit calls UI
- Constraints: Conventional Commits RU; no Co-authored-by; Ent migrate только с owner approve
- Approval record: N/A

## [REFERENCES]

- Finding IDs: A2W11-4-001/002, A2W11-5-003/004 (4)
- Map: `.planning/audits/total-2026-07/A2-DEDUP-2026-07-28.md`
- Source precedence: код

## [ACCEPTANCE]

- Done criteria:
  - 4 ID closed
  - docs+code match
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

## [EVIDENCE — 2026-07-28]

| ID | Status | Evidence |
|----|--------|----------|
| A2W11-4-001 | FIXED | `docs/BUSINESS.md`: голос + Phase 16+ + Out of Scope — webcam UI shipped; residual JQL/hub-mail |
| A2W11-4-002 | FIXED | удалён `useRooms` из `frontend/src/hooks/use-websocket.ts`; Vitest asserts no export |
| A2W11-5-003 | FIXED | `docs/security_guidelines.md` NF-11 → `cmd/server/wiring_fiber.go` |
| A2W11-5-004 | FIXED | `docs/system_specification.md` summary REST **~252** OpenAPI / ~286 inventory |

**Mutation:** вернуть `export function useRooms` → `i18n-push-rooms-p2` FAIL; вернуть `REST **286** endpoints` в summary → `docs-drift-task123` FAIL.

## [MUTATION CRITERIA] — verifier-ready

Проверено по коду HEAD (2026-07-30, fe2 W7 slot 5). Слой **frontend** / **docs**.

| ID | что ломать (file:line + фрагмент) | какой тест (`it` / команда) | слой |
|----|-----------------------------------|-----------------------------|------|
| A2W11-4-002 | `frontend/src/hooks/use-websocket.ts` — вернуть `export function useRooms(...)` | describe `useRooms dead export removed (A2W11-4-002 / TASK-123)` → `it("does not export useRooms (0 call sites → deleted)")` · файл `domains/platform/i18n-push-rooms-p2.test.ts` | frontend |
| A2W11-4-001 | `docs/BUSINESS.md` — Phase 16+ снова `Gantt, webcam,…` без «shipped»; убрать `webcam (UI shipped` / `CallControls` / `setCameraEnabled` | describe `docs drift NF-11 / REST / webcam (A2W11 / TASK-123)` → `it("A2W11-4-001: BUSINESS.md webcam wording…")` · `docs-drift-task123.test.ts` | docs |
| A2W11-5-003 | `docs/security_guidelines.md` NF-11 evidence → `cmd/server/main.go` (вместо `wiring_fiber.go`) | `it("A2W11-5-003: NF-11 evidence points to wiring_fiber.go")` | docs |
| A2W11-5-004 | `docs/system_specification.md` summary — `REST **286** endpoints` вместо `~252` | `it("A2W11-5-004: system_specification summary uses OpenAPI ~252 methodology")` | docs |

### Gaps (mutation-proof)

| Guard | Ловит тест на том же слое? | Примечание |
|-------|----------------------------|------------|
| useRooms deleted (source-guard) | ✅ Vitest readFile | frontend |
| BUSINESS / NF-11 / REST docs | ✅ Vitest readDoc | docs (не FE runtime) |
| layer_mismatch | нет | source/docs guards = intended layer |
| FE runtime webcam UI | N/A | docs-only ID |

**Confidence:** CONFIRMED (4/4 runnable).

### Mutation prove — 2026-07-30 W8 slot 4 (fe1)

Evidence: `.planning/audits/total-2026-07/MUTATION-2026-07-30-W8-fe1.md`  
**4/4 PROVEN** (Break→RED named `it`→restore→GREEN). Статус задачи → **Закрыта**.
