# Local task inputs

Кладите один UTF-8 task-файл на запуск, например `auth-unit-scope.txt`.

Правила:

- без secrets, tokens, passwords, private keys или сырого содержимого `.env`;
- описывайте outcome, allowed paths и forbidden paths простым языком;
- передавайте файл в `scripts/build_task_packet.py --task-file …`;
- не вставляйте задачу shell-inline.

Эти файлы — локальные рабочие входы и git-ignored, кроме этого README.
