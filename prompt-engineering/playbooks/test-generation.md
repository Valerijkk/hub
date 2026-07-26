# Playbook: от документации до автотестов

> Конвейер: спека → TC → независимая верификация → код автотестов → ревью. Каждый переход — через артефакт.

## Конвейер

```
┌─ документация фичи (спека/требования/API-док)
│
1. test-case-generator (templates/test-case-generation.md)
│    выход: TC-набор + матрица трассируемости + вопросы к доке
│    самоверификация пройдена (§2 шаг 4 агента)
│
2. ГЕЙТ: skills/test-verification/SKILL.md (templates/test-case-verification.md)
│    независимый исполнитель (≠ автор набора)
│    APPROVED → дальше | NEEDS WORK → возврат в шаг 1 (адресно: REJECT/FIX/GAP)
│    вопросы к документации → владельцу спеки ПАРАЛЛЕЛЬНО (не блокируют
│    approved-часть, блокируют только зависящие кейсы)
│
3. PLAN (для нетривиального пакета): skills/phase-plan/SKILL.md
│    files / depends / new checks / regression / done / stop
│
4. Автоматизация: qa-auto (или явно назначенный test coder)
│    вход: APPROVED TC; каждый тест ссылается на TC-ID (active profile TRACEABILITY)
│    уровень — из TC («Уровень автоматизации»), инфраструктура — из
│    `TEST-ARCHITECTURE.md`; недостающая инфраструктура — отдельной задачей ДО тестов
│
5. ЦИКЛ: targeted test → affected regression до green
│    FAIL → triage code|test|infra|flaky до любого изменения expectation
│
6. АВТОГЕЙТЫ: skills/anti-ai-check/SKILL.md +
│    skills/naming-check/SKILL.md
│
7. ГЕЙТ: test-code-reviewer (templates/review-request.md)
│    сверка код ↔ TC: каждый approved-TC реализован, каждый тест трассируется
│    + полный чек-лист ревью (паттерны, нейминг, гигиена)
│
8. Прогон в CI под флагами канона → зелёный
│
9. skills/verify-goal/SKILL.md: evidence по criteria автоматизации
│
└─ 10. test-documenter (templates/docs-sync.md): TC-каталог (TC ↔ автотест),
       карта сьютов, coverage-док
```

## Правила конвейера

1. **Нет спеки — нет конвейера.** Exploratory-режим (TC из кода) — только явным решением, кейсы помечены `SOURCE=code` и идут на ревью аналитика.
2. **TC-ID сквозной:** спека R → TC → имя/комментарий теста → отчёт CI. Обрыв цепочки в любом звене — находка ревью.
3. **Матрица живёт:** после шага 6 матрица «R × TC × автотест» — часть TC-каталога; новые требования начинают новый прогон конвейера.
4. **Частичная сдача допустима:** approved-подмножество TC автоматизируется, пока спорные кейсы ждут ответов — но матрица помечает их OPEN.

## Роли и независимость

| Шаг | Исполнитель | Запрет |
|-----|-------------|--------|
| 1 | tc-generator | не пишет код тестов |
| 2 | любой агент/человек ≠ автор шага 1 | не правит кейсы сам |
| 3 | orchestrator / qa-auto | не смешивает независимые files в один plan |
| 4–6 | qa-auto | не меняет TC (несогласие → возврат на шаг 2) |
| 7 | test-code-reviewer | readonly |
| 9 | verifier ≠ author | goal-backward, не «tasks done» |
| 10 | test-documenter | не меняет тесты/TC |
