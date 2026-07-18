---
name: anti-ai-check
description: Детектор AI-говнокода в тестах и коде — комментарии-нарраторы, плейсхолдеры, фейковые тесты, мёртвые ассерты, generic AI-язык. Отчёт с severity.
---

# /anti-ai-check — детектор AI-паттернов

> AI-сгенерированный код имеет узнаваемые болезни. Скил грепает их масс-паттернами и классифицирует по severity. Пути для скана — из `context/PROJECT-CONTEXT.md` §3 (директории кода и тестов). Все команды — `rg` (ripgrep); подставь свои расширения из стека проекта.

## Step 1. Комментарии-нарраторы (пересказ кода словами)

```bash
# "// increment counter" над counter++ — комментарий, повторяющий код
rg -n "// ?(Handle|Return|Define|Import|Initialize|Create|Set|Get|Check|Update|Delete|Loop|Call|Increment|Assign|Declare) (the |a )?" {{src_dirs}}
rg -n "# ?(handle|return|define|import|initialize|create|set|get|check|update|delete) " {{src_dirs_python}}
rg -n "// ?(увеличиваем|возвращаем|создаём|проверяем|получаем|обновляем|удаляем|инициализируем) " {{src_dirs}}
```

**Severity: LOW** (шум) / **MEDIUM** если комментарий врёт (противоречит коду рядом).

## Step 2. AI-шаблонные фразы и мета-комментарии

```bash
rg -in "(as an ai|note that this|it's worth noting|for demonstration purposes|in a real (application|project|scenario)|you (may|might) want to|this is a (simple|basic) (example|implementation)|for simplicity)" {{src_dirs}}
rg -n "// ?(Здесь|Тут) (мы|можно|нужно|должны)" {{src_dirs}}
# Комментарии-обращения к ревьюеру (объясняют изменение, а не код)
rg -in "(fixed the|updated to|changed from|as requested|per the requirements|new implementation)" {{src_dirs}} -g '*.{{ext}}'
```

**Severity: MEDIUM** — мета-комментарии удаляются всегда, им место в commit message.

## Step 3. Плейсхолдеры и заглушки

```bash
rg -in "(lorem ipsum|placeholder|your.?(api.?key|token|url|email)|example\.com|test@test|foo@bar|dummy|fake.?data|sample.?data|changeme|xxx+)" {{src_dirs}} --glob '!**/*test*' --glob '!**/fixtures/**'
rg -n "(TODO|FIXME|HACK)(?!\()" {{src_dirs}}   # TODO без владельца/тикета
```

**Severity: HIGH** в продакшн-коде, LOW в тестовых фикстурах (там допустимо осознанно).

## Step 4. Фейковые и мёртвые тесты (главная AI-болезнь в тестах)

```bash
# Тест без единого ассерта
rg -L -l "func Test|it\(|test\(" {{test_dirs}} | xargs rg -L "assert|expect|require|should"
# Всегда-зелёные ассерты
rg -n "(assert\.True\(t, true\)|expect\(true\)\.toBe\(true\)|assertTrue\(true\)|expect\(1\)\.toBe\(1\))" {{test_dirs}}
# Отключённые тесты без причины
rg -n "(\.skip\(|\.only\(|t\.Skip\(\)|@Disabled|xit\(|xdescribe\()" {{test_dirs}}
# Пустое тело теста / только заглушка
rg -n -A2 "(func Test\w+\(t \*testing\.T\) \{\s*\}|it\(['\"].*['\"], ?\(\) ?=> ?\{\s*\}\))" {{test_dirs}}
# Catch/recover, глотающий ошибку в тесте
rg -n "catch ?\(\w*\) ?\{ ?\}" {{test_dirs}}
```

**Severity: CRITICAL** — фейковый тест хуже отсутствующего: даёт ложную уверенность и зелёный CI.

## Step 5. Копипаст и мёртвый код

```bash
# Закомментированные блоки кода (3+ строки подряд)
rg -n -U "(^\s*//.*\n){4,}" {{src_dirs}} | rg -v "(Copyright|License|eslint|prettier)"
# Дубли тестовых функций с суффиксами-порядковыми
rg -n "func Test\w+(2|Copy|New|Old|Backup)\(|it\(['\"].*\(copy\)" {{test_dirs}}
# Неиспользуемые импорты-хвосты (для языков без компилятора это проверит линтер из PROJECT-CONTEXT §4)
```

**Severity: MEDIUM** — мёртвый код запрещён, история есть в git.

## Step 6. Языко-специфичные запреты проекта

Возьми список запрещённых паттернов из `context/PROJECT-CONTEXT.md` §7 и `context/PATTERNS-LIBRARY.md` §7 и прогрепай каждый. Примеры типовых (замени на свои):

```bash
rg -n "{{forbidden_pattern_1}}" {{src_dirs}}   # напр. println-debug вместо логгера
rg -n "{{forbidden_pattern_2}}" {{src_dirs}}   # напр. deprecated http-клиент
rg -n "sleep\(|Sleep\(|setTimeout\(.*, ?\d{4,}" {{test_dirs}}   # sleep-синхронизация в тестах
```

## Формат отчёта

```
## Anti-AI Check — {{scope}} — {{date}}

🔴 CRITICAL (N) — фейковые тесты, всегда-зелёные ассерты
  - file:line — <паттерн> → <что сделать>
🟠 HIGH (N) — плейсхолдеры в проде, TODO без владельца
🟡 MEDIUM (N) — мета-комментарии, мёртвый код, врущие комментарии
⚪ LOW (N) — нарраторы

Итого: N находок в M файлах. Чисто: ✅/❌
```

## Политика

- CRITICAL блокирует мерж безусловно.
- Скил только репортит; фиксы — отдельным шагом (агент или человек), чтобы отчёт оставался воспроизводимым.
- Прогонять: после каждой AI-сгенерированной волны кода, перед ревью, периодически по всему репо.
