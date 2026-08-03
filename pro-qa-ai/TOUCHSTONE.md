# Touchstone — AI Quality Layer for Test Frameworks

> **Product line:** Touchstone
> **Edition:** Standard (foundation)
> **Source tree:** `pro-qa-ai/`
> **Pro edition:** sibling folder `prompt-engineering-pro/` (Hub monorepo; not inside portable package)

## Что это

**Touchstone** — переносимая AI-прослойка для инженеров автоматизации. Она не
заменяет тестовый фреймворк: изучает уже существующий, даёт моделям
проверяемый контекст и паттерны, ведёт агентов по lifecycle и измеряет
качество относительно bare-prompt baseline.

Это фундамент продуктовой линейки Touchstone для тестовых фреймворков.

Деление линейки идёт по **объекту работы**, а не по уровню «базовый /
продвинутый»:

| Издание | Папка | Объект работы | Для кого |
|---------|-------|---------------|----------|
| **Standard** | `pro-qa-ai/` | Тестовый фреймворк | Инженер автоматизации тестирования |
| **Pro** | `prompt-engineering-pro/` (sibling) | Продукт, который разрабатывают | Команда разработки целиком |

Автоматизатору достаточно Standard. Команде разработки нужны оба: Pro
ставится поверх Standard и переиспользует его context, rules, recipes и
validator.

## Границы Standard

**Входит**

- Bootstrap/refresh контекста тестового фреймворка (provenance, `UNKNOWN`)
- Четыре агента `test-*`: генерация TC, ревью кода фреймворка, отладка, документация
- Rules, skills, playbooks, recipes, prompt fragments 01–12
- Adapters (Cursor / CI / generic-chat) и model routing
- Structural validator, packager, inventory, task-packet builder
- Eval protocol и fixtures (effectiveness — отдельный gate)

**Не входит**

- Scaffold нового тестового фреймворка (pytest/Playwright и т.п.)
- Роли разработки продукта и процессы поставки — это Pro
- **Мутация продуктового кода** — требует доступа к продукту, которого у
  автоматизатора нет; живёт в Pro. В Standard её место занимает проверка
  чувствительности по управляемому входу теста (`rules/40`): слабее на
  уровень, но исполняемая, а не умозрительная
- Hub-only TZ / RBAC / Fiber / Ent инварианты
- Автоскан закрытых MR
- Заявления ROI без paired eval evidence

## Производственный статус

| Gate | Статус |
|------|--------|
| Structural validation | PASS (см. `VERSION.md`) |
| Portable package (без project profiles) | PASS |
| Unit tests toolkit scripts | PASS (+ environmental SKIP на symlink) |
| Paired effectiveness / ROI | `NOT_RUN` — честно, без маркетинга |
| Empty-repo physical install | pilot gate |
| Release eligibility | foundation-ready · full GA после eval + owner sign-off |

«Production вид» здесь означает: ясная идентичность, чистый вход, воспроизводимый
package, governance и отсутствие выдуманных метрик. Это не заявление
«уже измерили N× ускорение».

## Быстрый вход

1. [START-HERE.md](START-HERE.md) — сценарии A–D
2. [README.md](README.md) — философия и слои
3. [INSTALL.md](INSTALL.md) — перенос в другой репозиторий
4. Pro (в Hub monorepo): открой sibling `prompt-engineering-pro/START-HERE.md`
   — роли разработки, волны, GSD-порт, portable hub kit

## Связь с Hub

В репозитории Hub Standard живёт рядом с продуктовым кодом как эталон и
источник portable package. Hub-каноны (rule 56, TZ, REGULATION) остаются в
`.cursor/` и **не** копируются в Standard целиком — в Pro они портированы
как project-agnostic практики, где это безопасно.

Спека происхождения: `.planning/archive/july-2026/ai-expansion/`.
