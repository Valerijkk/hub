# 06 — Edge / corner coverage → автотесты

> Часть [Testing Guide](../TESTING_GUIDE.md).  
> Источник кейсов: `.planning/EDGE-CORNER-CATALOG.md` (каталог happy / edge / corner → TC backlog).  
> Этот гайд **не копирует** каталог — описывает, как превращать записи в тесты.

---

## Определения

| Тип | Смысл | Пример Hub |
|-----|-------|------------|
| **Happy** | Основной успешный путь | Magic Link → сессия; CreateIssue → `TST-1` |
| **Negative** | Ожидаемый отказ / ошибка контракта | invalid hash; 403 read-only channel; 409 version mismatch |
| **Edge** | Граница допустимого | BodyLimit 25 MB OK / +1 байт reject; пустой список; max title length |
| **Corner** | Редкая комбинация / гонка / порядок | double-submit invite; WS subscribe без membership; concurrent optimistic lock |

---

## Откуда брать кейсы

1. **Каталог** `.planning/EDGE-CORNER-CATALOG.md` — систематический backlog.
2. **TC-матрица** `docs/TESTING.md` — уже формализованные шаги (после `qa-manual-ver`).
3. **Спека** `docs/system_specification.md` / NFR / security guidelines (SEC-*, NF-*).
4. **Аудиты** (например corner/race findings) — только как вход для TC, не для «тихого» автотеста в обход конвейера.

**Порядок:** каталог/спека → `qa-manual` пишет TC → `qa-manual-ver` → `qa-auto*` автоматизирует.

---

## Выбор слоя (куда класть тест)

```
Corner/race логики сервиса     → Go unit (mockRepo) / targeted integration
Edge лимитов API               → Go handler (httptest) table-driven
Negative authz/RBAC            → Go unit + точечный E2E только если user-visible
Happy критический путь UI      → 1 E2E smoke + unit на детали
UI validation / store          → Vitest
```

**Правило стоимости:** экзотика → ниже в пирамиде. E2E на каждый corner = «мороженое» и сожжённый CI.

| Кейс | Слой | Почему |
|------|------|--------|
| version mismatch 409 | `backend/testframework/tracker` | детерминированно, быстро |
| login branding | E2E `auth/login.spec.ts` | user-visible |
| auth-store logout clears tokens | Vitest `auth-store.test.ts` | нет смысла в браузере |
| Telegram expired auth | Go `TestVerifyTelegramAuth_ExpiredAuth` | clock/boundary |
| WS без membership (NF-02) | Go ws + security | не E2E flaky сеть |

---

## Алгоритм: из строки каталога в автотест

1. **Прочитай кейс** — preconditions, steps, expected, модуль, severity.
2. **Найди / создай TC** в `docs/TESTING.md` (`TC-<MODULE>-NN`); дождись `qa-manual-ver`.
3. **Выбери слой** по таблице выше; проверь владельца (`qa-auto1` vs `qa-auto2`).
4. **Напиши тест** по паттерну слоя (guides 03–05):
   - Go: table-driven для границ; отдельный `Test…_…` для corner.
   - Vitest: `it("rejects …")` / empty state.
   - E2E: только если без UI нельзя подтвердить риск.
5. **Привяжи:** комментарий TC-ID + колонка «Автотест» + Allure `AS_ID` при TMS.
6. **Прогони** `-race` / vitest / e2e; зафиксируй время; cleanup артефактов.
7. **Отметь в каталоге/STATE** (если процесс команды требует) — покрытие кейса закрыто автотестом.

---

## Шаблоны покрытия

### Edge — table-driven (Go)

```go
tests := []struct {
	name    string
	size    int
	wantErr bool
}{
	{name: "at_limit", size: 25 << 20, wantErr: false},
	{name: "over_limit", size: (25 << 20) + 1, wantErr: true},
}
```

### Negative — явная доменная ошибка

```go
// Good
if err != auth.ErrTelegramAuthExpired {
	t.Fatalf("expected ErrTelegramAuthExpired, got: %v", err)
}

// Bad
if err == nil { t.Fatal("expected error") } // какая именно — неизвестно
```

### Corner — изоляция гонки

- Уникальные сущности на goroutine/тест.
- `t.Parallel` только с изолированным env.
- Не полагаться на `Sleep` для «успеть до гонки» — синхронизация через каналы / детерминированный mock.

### E2E — минимальный happy + один user-visible negative

Не переносить все 409/429 в Playwright. Один representative negative (например logout clears session) — достаточно.

---

## Что не автоматизировать сразу

- Исследовательские / UX-субъективные кейсы → ручной TC.
- Инфра-only (K3s misconfig) → runbook / devops check, не product E2E.
- Кейсы без наблюдаемого oracle → сначала уточнить AC у BA/SA.

---

## DoD для edge/corner automation

- [ ] Есть TC с expected result (не «проверить что работает»)
- [ ] Слой выбран осознанно (стоимость / стабильность)
- [ ] Границы в table-driven или явных `_AtLimit` / `_OverLimit` тестах
- [ ] Corner не использует sleep; при FAIL — triage
- [ ] Traceability TC ↔ тест ↔ (опционально) строка каталога

Связанные: [01-principles.md](01-principles.md) · playbook §3 (техники) · §7 (traceability).
