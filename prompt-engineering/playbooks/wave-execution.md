# Playbook: Wave Execution

## Цель

Параллельный запуск независимых агентов с жёсткими границами файлов.

## Шаги

1. Взять планы волны из phase-plan.
2. Проверить: пересечение файлов = 0 внутри волны.
3. Для каждого агента заполнить `agents/templates/wave-delegate.md`: outcome,
   files, references, approval, evidence, done и stop.
4. Выбрать модель по `model-routing.md` из текущего каталога; записать decision/fallback.
5. Запустить ≤4 параллельно (или больше при явной карте).
6. Собрать outputs, Verification и routing logs; конфликты → стоп и переплан.
7. `skills/verify-goal/SKILL.md` по success criteria волны.

## Шаблон делегирования (минимум)

```
Роль: {{role}}
Outcome: {{outcome}}
Файлы: {{files}}
Не трогать: {{out_of_scope}}
References: {{context_cards}} + {{pattern_cards}} + {{inputs}}
Approval: {{approval_boundary}}
Done/stop: {{done_definition}} / {{stopping_condition}}
Формат: outcome / paths / evidence / assumptions / risks / handoff / Verification
Условия: новое поведение + регрессия {{regression_scope}} до green или явный FAIL
```

## Чек-лист

- [ ] Нет свободного «сделай X»
- [ ] Есть regression_scope
- [ ] Handoff к verifier указан
- [ ] Есть approval/stop и model decision log

## Антипаттерны

- Два агента на один файл
- Волна без sync point
- Модель выбрана по памяти, а не по текущему каталогу
- Runtime fallback не отражён в evidence
