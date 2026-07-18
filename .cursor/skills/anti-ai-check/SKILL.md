---
name: anti-ai-check
description: Грепнуть запрещённые AI-фразы и паттерны по всему проекту — narrating comments, placeholder text, generic AI-sounding language.
disable-model-invocation: true
---

# /anti-ai-check — Anti-AI Pattern Scanner

Проверяет кодовую базу на типичные AI-generated антипаттерны.

## Step 1: Narrating Comments (комментарии-нарраторы)

```bash
cd C:/Users/Valerii/Desktop/hub

# Go файлы
rg -n "// (Handle|Return|Define|Import|Initialize|Create|Set|Get|Check|Update|Delete) the " backend/ --type go

# TypeScript файлы  
rg -n "// (Handle|Return|Define|Import|Initialize|Create|Set|Get|Check|Update|Delete) the " frontend/src/ --type ts
rg -n "// (Handle|Return|Define|Import|Initialize|Create|Set|Get|Check|Update|Delete) the " frontend/src/ --type tsx
```

## Step 2: Echo Comments (повторяют имя функции)

```bash
# "// increment counter" над `counter++`
# "// return result" над `return result`
rg -n "// (increment|decrement|return result|return error|log error|parse request|send response)" backend/ frontend/src/ --type go --type ts
```

## Step 3: Placeholder Text

```bash
rg -rn "(TODO:|FIXME:|HACK:|placeholder|lorem ipsum|example\.com|test@test\.com|dummy|fake data)" \
    backend/internal/ frontend/src/ \
    --type go --type ts --type tsx
```

## Step 4: Deprecated Patterns

```bash
# minio-go вместо aws-sdk-go-v2
rg -n "minio-go|minio\.New\(" backend/ --type go

# Gorilla WebSocket вместо Fiber
rg -n "gorilla/websocket" backend/ --type go

# log.Printf вместо slog
rg -n "log\.Printf\|log\.Println\|fmt\.Println" backend/ --type go

# console.log в production TS
rg -n "console\.log\(" frontend/src/ --type ts --type tsx
```

## Step 5: Security Antipatterns

```bash
# http.DefaultClient (SSRF risk)
rg -n "http\.DefaultClient\." backend/ --type go

# io.ReadAll without limit
rg -n "io\.ReadAll\(" backend/ --type go

# Timing-unsafe secret compare
rg -n '!= "' backend/internal/auth/ backend/internal/ws/ --type go
```

## Результат

Для каждой найденной проблемы:
```
🔴 [TYPE] file:line — description
```

Если 0 результатов: `✅ No anti-AI patterns found`

## Примеры реальных находок и фиксов (Hub)

| Тип | Пример находки | Фикс |
|-----|-----------------|------|
| Narrating comment | `// Return the result` над `return issue, nil` в `service_issues.go` | Удалить комментарий, код самодостаточен |
| Echo comment | `// increment counter` над `counter++` в `automation_engine.go` | Удалить; если логика неочевидна — заменить на объяснение WHY, не WHAT |
| Placeholder | `dummy data` в seed-скрипте `provisioner.go` (не тестовый файл) | Заменить на реальный fixture-генератор или перенести в `_test.go` |
| Deprecated pattern | `minio.New(` в старом PoC-файле calls-хранилища | Заменить на `aws-sdk-go-v2` (mandatory, `tz_constraints.md`) |
| Security antipattern | `http.DefaultClient.Get(userURL)` в `ical_subscription` фиче (calendar) | `safehttp.NewSafeHTTPClient(10*time.Second)` + `ValidateOutboundURL` (SSRF, NF-07 аналог) |
| Security antipattern | `io.ReadAll(r.Body)` в file upload handler без лимита | `io.LimitReader(r.Body, maxBytes)` — memory DoS |

## Когда запускать

- Перед `code-reviewer` pre-merge review (обязательный шаг)
- После больших AI-сгенерированных волн (be1/be2/fe1/fe2 параллельные потоки) — команда часто оставляет narrating-комментарии по инерции
- Входит как Phase 1 в `.cursor/skills/style-check/SKILL.md` (расширенная версия)

## Связь с REGULATION

`.cursor/rules/10-style.mdc` — источник правды по запрещённым паттернам (Universal Rules + Prohibited Patterns table).
