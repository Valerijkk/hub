# Демо для руководителя — 15 минут

## Цель

Показать не «красивый промпт», а управляемый цикл:

1. AI получает факты конкретного тестового фреймворка.
2. Использует одобренные project patterns.
3. Работает в явных границах.
4. Возвращает evidence.
5. Эффект можно измерить относительно baseline.

## Подготовка

- Structural validator PASS.
- Generated profile выбранного проекта просмотрен framework owner.
- Сохранены live Cursor catalog snapshot, routing decision, фактический model ID
  и fallback/tool trace; self-reported routing явно помечен как PARTIAL.
- Используются только synthetic или разрешённые данные.
- При наличии реального примера коллеги он добавлен человеком и имеет approved
  redaction/reuse status.
- Eval metrics до запуска остаются `NOT_RUN`.

## Сценарий

### 0–2 мин — проблема

Показать bare fragment из `prompts/`: без project context он не знает реальные
пути, команды, фикстуры и принятый стиль.

### 2–5 мин — bootstrap

Показать вызов:

```text
/bootstrap-test-context
project_root=<repo>
test_roots=<tests>
profile=<name>
```

Открыть generated:

- `CONTEXT-MANIFEST.yaml`;
- `TEST-ARCHITECTURE.md`;
- `COMMANDS.md` со статусом `not-run`;
- `SOURCE-MAP.md`;
- `PATTERN-INDEX.md`.

Акцент: bootstrap read-only, не читает MR/сеть/secrets и не запускает найденные
команды.

### 5–7 мин — знания коллег

Открыть `standards/team-review/threads/THREAD-TEMPLATE.md`. Объяснить:

- автоматизатор вручную вставляет bad code, комментарий коллеги и good code;
- curator не выдумывает пропущенное;
- project-only — default;
- только approved card становится нормативной.

Если реального разрешённого примера нет, оставить template — не подменять его
искусственным «успехом».

### 7–10 мин — production recipe

Выбрать `workflows/recipes/review-ai-test-code.md` или `generate-api-tests.md`.
Показать:

- outcome;
- allowed files/tools;
- context/pattern refs;
- output schema;
- evidence/completion bar;
- stop/approval conditions.

Sol Pro orchestrator выбирает downstream model по текущему Cursor catalog и
фиксирует fallback; prompt body не содержит устаревающий model slug.

### 10–12 мин — проверки

```powershell
py -3 -B .\pro-qa-ai\scripts\validate_toolkit.py `
  --root .\pro-qa-ai

py -3 -B -m unittest discover `
  -s .\pro-qa-ai\scripts\tests `
  -p "test_*.py" -v
```

Показать фактический exit code и output. Не использовать скриншот старого PASS.

### 12–15 мин — измерение и решение

Открыть `standards/practices/coverage.md` и `docs/ROI-METRICS.md`.
Объяснить paired baseline/assisted pilot и decision gates.

## Вопросы, к которым быть готовым

**Это заменяет ревьюера?** Нет. Слой уменьшает повторяемый шум и делает evidence
структурированным; high-impact решения остаются за человеком.

**Куда уходят наши исходники?** Никуда автоматически. Network/provider use
подчиняется adapter/model/data policy; project examples не входят в portable
baseline без разрешения.

**Почему не обещаем 10x?** Потому что multiplier зависит от проекта, модели,
задач и качества baseline. Его можно утверждать только после paired pilot.

**Что произойдёт при смене модели?** Registry/routing обновляется из фактического
catalog, после чего affected evals запускаются повторно.

## Успешное демо

- Руководитель понимает, что слой управляет контекстом и качеством, а не только
  формулировками prompts.
- Все показанные PASS имеют свежий output.
- Не показаны секреты, закрытый код без разрешения или выдуманные KPI.
- Следующее решение сформулировано: pilot scope, owner, дата, acceptance gates.
