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

### Layering violations (rule `10-style.mdc` — критично для be-ver)

```bash
# *ent.Client в Service (запрещено — нарушение layering)
rg -n "\*ent\.Client" ./internal/*/service*.go --type go

# *ent.Client в Handler (запрещено)
rg -n "\*ent\.Client" ./internal/*/handler*.go --type go

# Пример реального нарушения (искать по доменам с недавним рефакторингом)
rg -n "\*ent\.Client" ./internal/tracker/ ./internal/bases/ ./internal/tms/ --type go
```

### Security-инварианты TZ (обязательный grep перед PR)

```bash
# http.DefaultClient (SSRF risk) — вместо safehttp.NewSafeHTTPClient
rg -n "http\.DefaultClient" ./internal/ --type go

# io.ReadAll без LimitReader (memory DoS)
rg -n "io\.ReadAll\(" ./internal/ --type go

# != для секретов вместо subtle.ConstantTimeCompare
rg -n '!= "' ./internal/auth/ ./internal/ws/ --type go

# panic() в production paths (handler/service, не тесты)
rg -n "panic\(" ./internal/ --type go --glob '!*_test.go'
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
Phase 2: Go Style (+layering +security) — N findings
Phase 3: TypeScript Style — N findings
Phase 4: A11y — N findings
Phase 5: CSS — N findings

Total: N findings
🔴 FAIL — fix before PR | ✅ PASS
```

## Примеры реальных находок Hub (для калибровки серьёзности)

| Находка | Домен | Серьёзность | Действие |
|---------|-------|-------------|----------|
| `*ent.Client` протащен в `service_boards.go` после быстрого хотфикса | tracker | 🔴 CRITICAL (layering) | Вынести через `repo_boards.go` interface, немедленный фикс до merge |
| `console.log` оставлен в `use-composer-ai-suggest.ts` после дебага | ai (frontend) | 🟡 MEDIUM | Удалить перед PR, не критично для прод (dev-only лог), но правило style-check ловит |
| `!= "` в сравнении webhook-подписи (`export_hooks.go`) | tracker/bases | 🔴 CRITICAL (SEC-02) | `subtle.ConstantTimeCompare`, эскалация `security-auditor` |
| `<button onClick=...>` без `type="button"` в bases inline-edit | v4/bases (frontend) | 🟡 MEDIUM | Implicit submit риск в форме — добавить `type="button"` |
| `key={index}` в списке кейсов TMS (`case-table-row.tsx`) | tms (frontend) | 🟡 MEDIUM | React reconciliation bug при пересортировке — заменить на `case.id` |

## Связь с REGULATION / GSD

- Полный gate (сборка+тесты+lint) — отдельно, `.cursor/skills/qa-gate/SKILL.md`
- Anti-AI паттерны — базовый набор в `.cursor/skills/anti-ai-check/SKILL.md` (Phase 1 здесь его переиспользует)
- GSD-эквивалент peer review: `gsd-code-review` (более широкий, не построчный grep) — см. `.cursor/GSD-INTEGRATION.md` §Hub slash-skills vs GSD skills
