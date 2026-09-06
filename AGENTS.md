# Hub — AI Agent Workspace

Маршрутизация моделей: `.cursor/playbooks/model-routing.md`; регламент и предметные правила — `.cursor/REGULATION.md` и `.cursor/rules/`.
Начинай с этого файла, актуального header `.planning/STATE.md` и `docs/README.md`, затем читай только контекст задачи.

`team-lead`/root работает на `gpt-6-astra` с `max`; `ultra` не используется.
Для каждой самостоятельной подзадачи оркестратор явно выбирает модель и effort по `.cursor/playbooks/model-routing.md`.
Роль описывает ответственность, но не фиксирует runtime; requested runtime и observed runtime/`UNKNOWN` передаются в handoff.
Один active writer владеет границей файлов; reviewer/verifier остаётся read-only. Сохраняй dirty work и следуй явному запросу на ветку/worktree.

`pro-qa-ai/**` — защищённая founder-зона: не читать, не менять и не запускать без явно переданного артефакта.
Hub tests, build, lint и validators запускаются только через task-owned Docker Compose exact service; host — только чтение, редактирование и Git.
Если проверка не выполнена, фиксируй `NOT_READY` или `NOT_RUN`; не выдавай это за `PASS`. Hub остаётся незавершённым MVP.
Глубина workflow, тестов и review соответствует риску и размеру задачи; для docs/config не имитируй полный product cycle.
Ограничивай поиск, запись и проверки нужными путями; переиспользуй кэши и артефакты, избегай лишних полных сборок, переиндексаций, массовых копирований/перегенераций, но не пропускай обязательные проверки.

Используй `codex/<task>` для новых Codex-веток и не чисти чужой checkout. Не коммить секреты или AI attribution/co-author trailers.
Новые commit subject/scope — Conventional Commit на английском, описание и body — на русском.
Security/auth, DB migrations, deploy/CI и ADR требуют founder review перед интеграцией.
