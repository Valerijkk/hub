# MUTATION W8 slot 4 — TASK-123 (fe1) · 2026-07-30

> Checker: **fe1** · Env: local Vitest · Criteria: `.planning/backlog/TASK-123-2026-07-28-fe-lib-dead-export-docs-drift.md` `[MUTATION CRITERIA]` (author: fe2 W7)  
> Коммит: **не создавался** (инструкция слота).

## Verdict

**4/4 PROVEN** · TASK-123 → **Закрыта**

| ID | Verdict | RED test (`it`) | Mutation (BREAK) | Layer |
|----|---------|-----------------|------------------|-------|
| A2W11-4-002 | **PROVEN** | `does not export useRooms (0 call sites → deleted)` | restore `export function useRooms` + `roomsKey` in `use-websocket.ts` | frontend |
| A2W11-4-001 | **PROVEN** | `A2W11-4-001: BUSINESS.md webcam wording…` | Phase 16+ → `Gantt, webcam,…`; drop `webcam (UI shipped` / CallControls / setCameraEnabled | docs |
| A2W11-5-003 | **PROVEN** | `A2W11-5-003: NF-11 evidence points to wiring_fiber.go` | NF-11 evidence → `cmd/server/main.go` | docs |
| A2W11-5-004 | **PROVEN** | `A2W11-5-004: system_specification summary uses OpenAPI ~252 methodology` | summary `REST **286** endpoints` вместо `~252` | docs |

Production restored (`git checkout --` на 4 файла). Mock-gap: нет (source/docs readFile guards).

## Per-ID evidence

### A2W11-4-002 — useRooms dead export

```text
BASELINE GREEN → PASS does not export useRooms…
BREAK: frontend/src/hooks/use-websocket.ts — append roomsKey + export function useRooms
→ FAIL does not export useRooms (expected not to match /\bexport\s+function\s+useRooms\b/)
RESTORE: git checkout -- frontend/src/hooks/use-websocket.ts
→ PASS same it
```

### A2W11-4-001 — BUSINESS.md webcam wording

```text
BREAK: docs/BUSINESS.md
  - Голос row: drop «webcam (UI shipped… CallControls / setCameraEnabled»
  - Phase 16+: «Gantt, webcam, full JQL…» (без shipped)
→ FAIL A2W11-4-001 (expected to match /webcam \(UI shipped/i)
RESTORE: git checkout -- docs/BUSINESS.md
→ PASS same it
```

### A2W11-5-003 — NF-11 → wiring_fiber.go

```text
BREAK: docs/security_guidelines.md NF-11 evidence `cmd/server/wiring_fiber.go` → `cmd/server/main.go`
→ FAIL A2W11-5-003 (expected wiring_fiber.go row)
RESTORE: git checkout -- docs/security_guidelines.md
→ PASS same it
```

### A2W11-5-004 — REST ~252 methodology

```text
BREAK: docs/system_specification.md summary API row → `REST **286** endpoints`
→ FAIL A2W11-5-004 (expected /REST\s*\*\*~?252\*\*/)
RESTORE: git checkout -- docs/system_specification.md
→ PASS same it
```

## Commands

```text
cd frontend
node scripts/link-testframework-node-modules.mjs
# baseline + post-restore:
npx vitest run \
  ../testframework/stages/unit/frontend-test/domains/platform/i18n-push-rooms-p2.test.ts \
  ../testframework/stages/unit/frontend-test/domains/platform/docs-drift-task123.test.ts
→ Test Files 2 passed · Tests 7 passed (~2.5s) after full restore
```

## Assumptions / unresolved

- Нет: criteria CONFIRMED 4/4; все ID ловятся named `it` на заявленном слое.
- Leftover MUTATED: нет (`git status` clean на 4 prod-файлах).

## Handoff

Evidence: этот файл · TASK-123 **Закрыта** · fe-ver → code-reviewer (docs+dead-export already FIXED).
