# MUTATION — 2026-07-29 · W1 · fe1 (независимый от авторов)

> Задача 2 / TASK-008 · Проверяющий: **fe1** · Авторы: TASK-052=fe2, TASK-054=docs-writer  
> Протокол: `.planning/audits/total-2026-07/W18-C5-MUTATION.md`  
> Rework после W18-C5 FAIL учтён (tsc fixture + `scripts/check-docs-inventory.mjs`)

## Итог

| ID | Class | Verdict | README |
|----|-------|---------|--------|
| TASK-052 | P2-FE-08 | **PASS** | Закрыта |
| TASK-054 | P2-FE-12 | **PASS** | Закрыта |

Production после restore: без diff в `frontend/src/api/bases.ts` и 4 docs inventory.

## TASK-052 — PASS

### ACCEPTANCE
`updateBase` без `version` → TS fail (не runtime `expectTypeOf` alone).

### Baseline GREEN
```
cd frontend
node scripts/link-testframework-node-modules.mjs
npx vitest run ../testframework/stages/unit/frontend-test/domains/platform/api-dto-gaps-p2.test.ts
→ 6/6 passed (tsc fixture ~1.5s)
```

### Мутация
`frontend/src/api/bases.ts`: `version: number` → `version?: number` в `updateBase` и `updateRecord`.

### RED
```
Tests  1 failed | 5 passed (6)
× tsc fixture: omit version is @ts-expect-error (mutation optional → unused)
  → tsc -p tsconfig.task052.json failed (unused @ts-expect-error)
✓ updateBase/updateRecord require version (expectTypeOf) — остался зелёным
```

**Важно:** `expectTypeOf` runtime **no-op** (остался GREEN при optional) — канонический guard = **tsc fixture** (`updateBase-version-required.fixture.ts` + spawn `tsc`), не type-only expect.

### Restore → GREEN
Вернул `version: number`. Vitest снова **6/6**. `git diff` bases.ts пуст.

### Mock-only?
Нет — fixture импортирует реальный `@/api/bases` и гоняет `tsc` по production-типам.

---

## TASK-054 — PASS

### ACCEPTANCE
Вернуть устаревший count (515) → расхождение; автопроверка краснеет.

### Guard
`scripts/check-docs-inventory.mjs` (TASK-080 rework): measured `_test.go` в `backend/testframework/` vs `**N**` в:
- `docs/TROUBLESHOOTING.md`
- `docs/test_plan.md`
- `docs/TESTING.md`
- `docs/testframework_strategy.md`

Measured на прогоне: **821**.

### Мутация
Во всех 4 файлах `**812**` → `**515**` (6 вхождений).

### RED
```
FAIL docs/TROUBLESHOOTING.md: claimed 515 _test.go, measured 821
FAIL docs/test_plan.md: claimed 515 _test.go, measured 821
FAIL docs/TESTING.md: claimed 515 _test.go, measured 821
FAIL docs/testframework_strategy.md: claimed 515 _test.go, measured 821
docs inventory drift (measured=821)
exit 1
```

### Restore
`**515**` → `**812**` (история `was 515` без `**` не тронута). Diff docs пуст относительно pre-mutation.

### GREEN path (логический)
Скрипт: `claim === measured` → OK / exit 0. Measured=821; при claim=821 был бы GREEN. Pre-existing drift docs `**812**`≠821 **до** мутации — не регрессия этой проверки; gate на неверный count работает.

### Mock-only?
Нет — script читает реальные docs + дерево `backend/testframework/`.

---

## Таблица PASS/FAIL

| ID | Мутация | RED | GREEN restore | Verdict |
|----|---------|-----|---------------|---------|
| TASK-052 | `version?` в bases.ts | tsc fixture FAIL | Vitest 6/6 | **PASS** |
| TASK-054 | docs `812`→`515` | inventory script exit 1 | files restored | **PASS** |
