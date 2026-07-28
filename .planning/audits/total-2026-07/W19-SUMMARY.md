# W19 SUMMARY — перекрёстная проверка

> 2026-07-28 · TASK-008 Stage C · после W18

## Orphans P0/P1

- **P0 orphans: 0**
- **P1 orphans: 10** — все U-section (ADR/PO), missed mutation-ready = **0**
- См. `W19-A1-ORPHAN-P0P1.md`

## Evidence gaps

- **39** PASS-задач без SHA в теле файла задачи (evidence есть в `W18-*-MUTATION.md`)
- См. `W19-A2-EVIDENCE-GAPS.md`

## Combos

- **13** комбинаций: CRITICAL **1** · HIGH **5** · MEDIUM **5** · LOW **2**
- Top: **C-001** GDPR Mongo eraser no-op + HardDelete
- См. `W19-A3-COMBOS.md`

## Red team (top residual)

1. AddTeamToProject without team ACL (TASK-014)
2. ConfirmUpload weak entity/object proof (TASK-045)
3. Dual StartSprint race unproven (TASK-041)
4. SSRF push/link-preview + To4 residual (TASK-019)
5. AI privacy/fail-open + stream cancel (TASK-044)

См. `W19-A4-REDTEAM.md`

## Before W20

1. Не закрывать TASK-008
2. Приоритет reopen-фиксов: **014, 041, 045**, затем 019/042/044/046
3. Backfill SHA в тела PASS-задач (optional hygiene)
4. U\* не разблокировать без ADR/PO
