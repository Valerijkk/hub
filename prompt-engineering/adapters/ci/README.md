# CI-адаптер

Используйте для неинтерактивных запусков агента. CI runner обязан явно задавать
permissions и семантику сбоев; одних промптов недостаточно как sandbox.

## Обязательные входы job

- role/core file
- immutable source revision и diff/base
- allowed paths и do-not-touch paths
- ревизии generated context/pattern card
- done criteria, evidence commands, stopping condition
- approval policy (обычно без external/destructive действий)
- актуальный снимок runtime catalog и routing decision
- destination артефактов для output, Verification и логов routing/fallback

## Контракт исполнения

1. Собрать system prompt из `agents/README.md` + выбранная роль.
2. Собрать task message из `_MASTER-PROMPT-TEMPLATE.md`.
3. Выбрать модель из фактического каталога CI runtime. Для Cursor Cloud/API
   используйте текущий model-list endpoint и возвращённый ID.
4. Запускать с least privilege. Review-роли — readonly; implementation-роли
   получают только необходимые credentials и paths.
5. Считать job failed при `FAIL`, отсутствии evidence, неожиданном fallback,
   out-of-scope diff или отсутствии Verification.
6. Загружать sanitized output и decision/fallback logs. Никогда не печатать secrets.

## Граница approval

CI не должен деплоить, ротировать секреты, менять production data, одобрять
расходы или менять team model policy, пока job не является отдельно одобренным
environment gate. Human approval нельзя вывести из текста промпта.
