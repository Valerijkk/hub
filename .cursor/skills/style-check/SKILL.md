---
name: style-check
description: Расширенный стиль-чек — anti-AI фразы + deprecated паттерны + echo-comments + a11y проблемы. Запускай перед PR review.
disable-model-invocation: true
---

# /style-check — Extended Style Check

Расширенный аудит кода на стилевые нарушения. Включает всё из `/anti-ai-check` плюс дополнительные проверки.

## Phase 1: Anti-AI Check

Запусти `/anti-ai-check` (все шаги).

## Phase 2: Go Style Violations

```bash
cd C:/Users/Valerii/Desktop/hub/backend

# log.Printf / log.Println вместо slog
rg -n "log\.(Printf|Println|Print|Fatal|Fatalf|Panic|Panicf)\(" . --type go

# fmt.Println в production paths (не в тестах)
rg -n "fmt\.Println\|fmt\.Printf" ./internal/ --type go

# Голые return без значения в функции с returns
rg -n "^\s+return$" ./internal/ --type go

# Неиспользуемые параметры context (должен быть первым)
rg -n "func.*\(.*context\.Context" ./internal/ --type go -l | head -20

# Нет defer для cleanup
rg -n "resp\, err := " ./internal/ --type go -A2 | grep -v "defer resp.Body.Close"
```

## Phase 3: TypeScript Style

```bash
cd C:/Users/Valerii/Desktop/hub/frontend/src

# any без комментария
rg -n ": any\b" . --type ts --type tsx | grep -v "// "

# as any (type assertion escape)
rg -n " as any\b" . --type ts --type tsx | grep -v "// "

# TODO без тикета
rg -n "TODO:" . --type ts --type tsx | grep -v "TODO: HUB-"

# console.log
rg -n "console\.(log|warn|error)\(" . --type ts --type tsx
```

## Phase 4: A11y Check

```bash
cd C:/Users/Valerii/Desktop/hub/frontend/src

# Кнопки без aria-label или видимого текста
rg -n "<button[^>]*>" . --type tsx | grep -v "aria-label\|<button.*>[^<]"

# img без alt
rg -n "<img[^>]*>" . --type tsx | grep -v "alt="

# Интерактивные div без role
rg -n "<div.*onClick" . --type tsx | grep -v "role="
```

## Phase 5: CSS / Design System

```bash
cd C:/Users/Valerii/Desktop/hub/frontend

# !important (запрещены кроме @media prefers-reduced-motion)
rg -n "!important" src/ --type css

# Inline styles (должны быть через CSS variables)
rg -n 'style=\{' src/ --type tsx | grep -v "style=\{.*var(--" | head -20

# Hardcoded hex colors (должны быть через design tokens)
rg -n "#[0-9a-fA-F]{3,6}" src/ --type tsx --type ts | grep -v "// " | head -20
```

## Результат

```
Phase 1: Anti-AI — N findings
Phase 2: Go Style — N findings
Phase 3: TypeScript Style — N findings
Phase 4: A11y — N findings
Phase 5: CSS — N findings

Total: N findings
🔴 FAIL — fix before PR | ✅ PASS
```
