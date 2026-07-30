# Extract — Wave Protocol (portable)

> Сжатие GSD execute-phase + Hub team-workflow + Standard wave-execution.

## 1. Подготовка

1. Outcome фазы одной фразой (goal-backward).  
2. Atomic units: каждая — самостоятельный PASS/FAIL.  
3. Граф зависимостей: что строго до чего.  
4. Волны: внутри волны — только disjoint file sets.  
5. Общие контракты (API/схема/типы) — **до** параллели.

## 2. Делегирование слота

На каждый слот — [master-delegate.md](master-delegate.md) с уникальным
allowlist файлов. Пересечение allowlist двух слотов одной волны = FAIL планирования.

## 3. Во время волны

- Не расширять scope «заодно»
- Блокер → STOP слота, эскалация orchestrator, не тихий workaround
- Docs sync в той же сессии, если менялось поведение

## 4. Сборка волны

1. Каждый слот: verdict + evidence  
2. Кросс-ревью при парных ролях  
3. Independent verifier на critical/security  
4. Goal-verify фазы (не «все задачи отмечены»)  
5. Handoff/STATE note

## 5. Verify goal-backward

Спроси: *если бы задач не было, а был только outcome — доказан ли он?*  
CI green без outcome evidence = недостаточно.

## 6. Anti-patterns

| Антипаттерн | Почему плохо |
|-------------|--------------|
| Одна волна на весь монорепо | Конфликты, неатомарность |
| «Сделай как в чате» | Субагент чата не видит |
| PASS без команды | Нет evidence |
| Мутация после «Закрыта» | Поздно; спецификация должна быть в ACCEPTANCE |
