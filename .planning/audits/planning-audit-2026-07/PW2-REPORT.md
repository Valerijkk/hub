# PW2-REPORT — 51 фаза без верификации

**Дата:** 2026-07-27  
**Источник списка:** `PHASE-STATUS-MATRIX.md` (52 без VERIFICATION на момент скана; плюс новые plan-only фазы)

## Сделано

Артефакты `VERIFICATION.md` / `VERIFICATION-SKIPPED.md` созданы батчами:

| Batch | Коммит (пример) |
|-------|-----------------|
| 14–21 | `f7b9c07` / `1b8ba4f` |
| 38–45 | `85dce92` |
| 46–62 | `1c2d699` |
| 64–999 | inside `c6aa14a` (staged with A4) |
| 127–136 plan-only | `9fa2f5b` SKIPPED |
| 22–37 | **агент в полёте** — добить до нуля «без артефакта» |

## Правило

- PASS только с замером
- FAIL/PARTIAL — нормально (122/124 coverage goals historically missed)
- 126/127–136 — SKIPPED (plan-only, not executed)

## Хуже ожидаемого

- Параллельные git-коммиты иногда склеивали чужие файлы в один commit message (A4 + PW2 batch) — содержимое артефактов валидно, сообщения шумные.
- **`.gitignore` содержит `.planning/**/VERIFICATION.md`** — часть артефактов PW2 лежит на диске, но **не в git**, пока не `git add -f`. На 2026-07-27: on-disk без артефакта = 0; tracked VERIFICATION ≈ 51 + SKIPPED 11. **Предложение владельцу:** убрать ignore или заменить на `VERIFICATION.local.md` pattern — иначе планирование снова «врёт» в remote.

## Владельцу

1. Снять ignore `VERIFICATION.md` или утвердить `git add -f` policy.
2. Пересчитать `goals_verified` после force-track всех файлов.
