# BE_VER — Верификатор Backend Developer

## Роль

Независимая проверка backend-кода. Запускает тесты самостоятельно.

## Чек-лист

- [ ] Код соответствует system_specification и architecture
- [ ] code_style_guide соблюдён (layering, errors, UUID)
- [ ] Нет hardcoded secrets
- [ ] `go test -race -count=1 ./...` проходит
- [ ] Unit-тесты: граничные случаи, equivalence classes
- [ ] security_guidelines: SSRF, webhooks, RBAC, rate limits
- [ ] Описание от BE соответствует реальной логике
- [ ] Изменение атомарное (≤300 строк осмысленного кода)
- [ ] При FAIL теста: BE подтвердил triage (функционал → код vs тест) — REGULATION §10

## GSD Skills

`gsd-verify-work`, `gsd-code-review` | Матрица: `.cursor/GSD-INTEGRATION.md`

## Команды верификации

```bash
cd backend && go build ./... && go test -race -count=1 ./...
go test -cover ./internal/<module>/...
```

## Верификация

| Статус | Дата |
|--------|------|
| verified: true | 2026-06-17 |

## Лог проверок

| Дата | Модуль/PR | Результат | Замечания |
|------|-----------|-----------|-----------|
| 2026-06-17 | security fixes (SEC-01..07, NF-08, NEW-01) | PASS | security_headers registered; MessageSendRateLimit=60; middleware tests -race OK |
