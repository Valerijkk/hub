# 01 — Evidence model

## Классы

| Label | Значение | Что разрешает |
|---|---|---|
| `CONTROL` | Механизм непосредственно ограничивает определённый класс ошибок | Кандидат на hard/hybrid gate, если проверка точна |
| `NORM` | Стандарт, спецификация или platform contract | Gate в заявленной области применимости, не универсальный design law |
| `EMP` | Эмпирическое исследование | Калибровка и warning; causal gate требует более сильного основания |
| `REC` | Официальная или широко применяемая рекомендация | Default или review question |
| `OPINION` | Мнение автора/школы | Discussion input, не gate |
| `POLICY` | Осознанная локальная политика Hub | Enforceable после явного принятия, ownership и rollout |
| `DISPUTED` | Спорная/context-dependent практика | Decision matrix с alternatives и trade-offs |

## Enforcement ladder

| Уровень | Когда выбирать | Требование |
|---|---|---|
| `HARD` | Failure class точен, нарушение опасно, detector достаточно надёжен | Механизм, scope, owner, remediation, exception path |
| `HYBRID` | Часть свойства проверяется машиной, семантика требует review/test | Machine evidence + named reviewer question |
| `WARNING` | Сигнал коррелирует с риском, но false positive/context значимы | Не блокирует; собирает локальные данные |
| `REVIEW` | Решение зависит от domain/change model | Конкретный вопрос, не лозунг |

## Gate admission test

Hard gate допускается только если на все вопросы дан ответ:

1. Какой конкретный failure предотвращается?
2. Как detector отличает нарушение от допустимого случая?
3. Каков scope: язык, boundary, runtime, тип артефакта?
4. Как исправить нарушение без ритуального workaround?
5. Кто владеет rule и как обновляется source/version?
6. Как оформить временное исключение с expiry и compensating control?

Если ответа нет, правило остаётся `WARNING` или `REVIEW`.

## Confidence

- `HIGH`: источник прямо поддерживает claim в сопоставимом scope либо mechanism очевидно ограничивает failure.
- `MEDIUM`: recommendation/empirical result полезен, но перенос зависит от контекста.
- `LOW`: opinion, слабая operationalization или непрямой источник.
- `UNKNOWN`: нет достаточного evidence; запрещено маскировать как certainty.

## Source hygiene

- Нормативное и эмпирическое evidence не смешиваются.
- Engineering blog подтверждает практику конкретной организации, но не causal universal law.
- Observational correlation не становится magic threshold.
- Ссылка без точной страницы/ревизии не поддерживает hard control.
- Текущий frozen snapshot и ограничения перечислены в [registry](sources/registry.md) и [integrity status](sources/evidence-status.md).

## Пример классификации

**BAD:** «Метод на 26 строк нарушает научно доказанный Clean Code standard».

**GOOD:** `EMP` S12 обнаружил Java-specific correlation; Hub может выдать `WARNING` после 25 logical lines и задать review question о cohesion/nesting, но не блокировать CI только по размеру.
