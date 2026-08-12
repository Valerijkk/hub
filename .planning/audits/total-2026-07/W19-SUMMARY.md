# W19 SUMMARY — перекрёстная проверка

> 2026-07-28 · TASK-008 Stage C · после W18

## Контекст после W18

| Метрика | Значение |
|---------|---------:|
| Mutation PASS → «Исправлена» | **39** |
| Mutation FAIL → «Открыта» | **9** |
| FAIL IDs | **014, 019, 041, 042, 044, 045, 046, 052, 054** |
| Catch-up | TASK-079, TASK-080 |

**TASK-008 не закрывать.**

## Orphans P0/P1

- P0 orphans: **0**
- P1 orphans: **10** (все U-section / ADR/PO; missed = 0)
- См. `W19-A1-ORPHAN-P0P1.md`

## Evidence gaps

- **39** PASS без SHA в теле задачи (evidence в `W18-*-MUTATION.md`)
- См. `W19-A2-EVIDENCE-GAPS.md`

## Combos

- **16** комбинаций: CRITICAL **2** · HIGH **6** · MEDIUM **5** · LOW **3**
- Top: C-001 GDPR Mongo eraser × HardDelete; C-002 dual StartSprint × no unique active
- См. `W19-A3-COMBOS.md`

## Red team top

1. AddTeamToProject без team ACL (014)
2. ConfirmUpload ghost (045)
3. GDPR Mongo residual PII (079)
4. SSRF To4/Meili residual (019/044)
5. AI nil-access fail-open (044)

См. `W19-A4-REDTEAM.md`

## Before W20

1. Не закрывать TASK-008
2. Приоритет reopen-фиксов: **014 → 079 → 045**, затем 044/019/041+046
3. U\* без ADR не трогать
4. W20: FINAL + полный hubtest + coverage vs 81.9%
