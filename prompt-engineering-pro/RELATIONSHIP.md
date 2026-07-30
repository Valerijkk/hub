# Standard ↔ Pro

```
┌─────────────────────────────────────────────────────────┐
│                 Touchstone Pro                          │
│  GSD max · prompt-contract · waves · mutation · planning│
│─────────────────────────────────────────────────────────│
│              Touchstone Standard (required)             │
│  context · agents · recipes · rules · evals · packager  │
└─────────────────────────────────────────────────────────┘
```

## Правила загрузки контекста агентом

1. Всегда грузить Standard `rules/00` + `rules/20` + `rules/40` + `rules/50`.
2. Для test-задач — Standard agents `test-*` и recipes.
3. Для оркестрации/фаз — Pro extracts `prompt-contract`, `wave-protocol`,
   `master-delegate`.
4. Для «закрыли баг/фикс» — Pro `mutation-evidence` + Standard `test-verification`.
5. Для постановки работ — Pro `backlog-task` (мастер-контракт).

## Что НЕ дублируем

| Артефакт | Где канон | В Pro |
|----------|-----------|-------|
| agents/*.md | Standard | ссылка |
| scripts/*.py | Standard | ссылка |
| prompts/01–13 | Standard | ссылка |
| Hub TZ / Fiber / Ent | `.cursor/rules/60` | **не переносится** |
| Hub REGULATION полный | `.cursor/` | сжатый extract практик |

## Версии

Pro `X.Y.*` требует Standard `X.Y.*` той же линии. Минорный drift запрещён
без записи в обоих `VERSION.md`.
