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
