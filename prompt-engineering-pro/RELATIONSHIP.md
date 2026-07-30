# Standard ↔ Pro

```
┌──────────────────────────────────────────────────────────┐
│  Touchstone Pro — объект работы: ПРОДУКТ                 │
│  12 ролей разработки · phase/wave/verify · corporate     │
│  delivery · GSD max · mutation · Cursor readonly-пары    │
│──────────────────────────────────────────────────────────│
│  Touchstone Standard — объект работы: ТЕСТОВЫЙ ФРЕЙМВОРК │
│  test-* агенты · context · recipes · rules · evals ·     │
│  validator · packager        (обязателен для Pro)        │
└──────────────────────────────────────────────────────────┘
```

Pro не «продвинутый Standard». Это соседний слой про другой объект
работы, который переиспользует инфраструктуру Standard вместо своей.

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
| `test-*` агенты | Standard | ссылка |
| Роли разработки | **Pro** | канон здесь |
| scripts/*.py | Standard | ссылка |
| prompts 01–12 | Standard | ссылка |
| prompts 13 | **Pro** | канон здесь |
| Hub TZ / Fiber / Ent | `.cursor/rules/60` | **не переносится** |
| Hub REGULATION полный | `.cursor/` | сжатый extract практик |

## Версии

Pro объявляет **минимальную** версию Standard, против которой проверен
(`depends_on` в [VERSION.md](VERSION.md)), а не жёсткое равенство
major/minor. Издания покрывают разные объекты работы и будут расходиться
в темпе: правка ростера разработки не обязана двигать версию тестового
слоя.

Требование одно: `depends_on` Pro должен указывать на реально
существующую версию Standard, и обновление Standard ниже этой границы
считается несовместимым.
