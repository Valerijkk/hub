# Generic chat-адаптер

Используйте, когда хост поддерживает system/user messages, но не
repository custom agents.

Для **слабой локальной модели** предпочитайте детерминированный packet path —
не вставляйте весь тулкит целиком.

## Основной локальный путь (рекомендуется)

1. Следуйте [START-HERE.md](../../START-HERE.md), Scenario D.
2. Положите задачу (без secrets) в `local-task-inputs/<task>.txt`.
3. Соберите один bounded packet через
   [`build_task_packet.py`](../../scripts/build_task_packet.py).
   Контракт и флаги:
   [`LOCAL-TASK-PACKET.template.md`](LOCAL-TASK-PACKET.template.md).
4. Дайте модели **только** сгенерированный
   `generated-task-packets/<packet>.md`.
5. Не заполняйте шаблон вручную; builder enforced schema, approvals и
   approved-only patterns.

Русский walkthrough:
[`docs/LOCAL-AGENT-GUIDE-RU.md`](../../docs/LOCAL-AGENT-GUIDE-RU.md).

## Ручной mapping (способный хост с доступом к файлам)

1. System message: конкатенация `agents/README.md` и одного выбранного
   `agents/<role>.md`, плюс rules `00` и `50`.
2. Добавляйте только релевантные context/pattern cards; не вставляйте весь repo.
3. User message: natural-language запрос или строгий шаблон из
   `agents/templates/` для delegation/handoff.
4. Tool permissions, write scope, approval boundary и stopping condition должен
   enforce хост — markdown сам их не обеспечивает.
5. Выбирайте модель из текущего каталога хоста по
   `playbooks/model-routing.md`; не переиспользуйте Cursor IDs у другого провайдера.

Если хост не может читать файлы репозитория, используйте builder packet path выше,
чтобы referenced content был inlined с hashes.

## Output

Требуйте schema роли плюс evidence, assumptions, confidence, handoff и
Verification. Generic chat без tools должен помечать ненаблюдаемые file/runtime
утверждения как `NOT VERIFIED` / `NOT_RUN`.
