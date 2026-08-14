# Terminal Codex: продолжение Hub после Desktop handoff

> **CURRENT BANNER:** This historical terminal text is subordinate to `.continue-here.md`.
> Folder source/oracle `f8cb7643986fc591cae16764575337b72ef3a692` is narrowly static accepted;
> the 20-path docs candidate awaits three exact-SHA reviews. After PASS, root only freezes and
> adjudicates `144-03B-Rule`; no Rule TDD or production authorization follows.

> **CONSUMED / HISTORICAL PRE-01A RESUME PROMPT:** не выполняй старое следующее действие из
> исторического текста ниже. Текущая authority — `.continue-here.md`: `144-03A-Address` уже
> `STATIC_ACCEPTED` на `41564d003946ddb78df18efca166ca86e4055405`; callable/encoding/oracle
> `144-03A-Thread` имеет статус `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` на
> `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` (C/I/M `0/0/0`). Текущее
> Следующая строка историческая и не является текущей authority: Thread уже narrowly
> `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`; current action определяет banner выше —
> historical Folder freeze/adjudication text; current authority is the banner above. Текст
> ниже исторический и подчинён `.continue-here.md`.

Ты продолжаешь работу как root-orchestrator Hub. Не полагайся на этот чат и не считай любую неуказанную проверку выполненной.

## Старт и обязательный preflight

```powershell
Set-Location 'C:\Users\Valerii\Desktop\hub\.worktrees\codex-mail-module-20260813'
git branch --show-current
git status --short
git rev-parse HEAD
```

Ожидаемая ветка: `codex/mail-module-20260813`. `1c0299bd720ad6f8521314d83e478dedc666eccd` — pre-handoff baseline, а не ожидаемый конечный HEAD. На resume убедись, что текущий HEAD — его descendant, содержит три committed handoff-файла (`.planning/HANDOFF.json`, `.planning/phases/143-calendar-rsvp-core/.continue-here.md`, `.planning/phases/143-calendar-rsvp-core/TERMINAL-RESUME-PROMPT-RU.md`), а worktree чист перед новыми действиями. До запуска `$gsd-resume-work` файл `.planning/HANDOFF.json` должен существовать. Он одноразовый: после успешного восстановления контекста `$gsd-resume-work` может удалить его; committed history остаётся recoverable. Если ветка, ancestry, файлы или чистота не соответствуют — немедленно сообщи `BLOCKED` с точными фактами; не переключай ветку, не откатывай и не исправляй расхождение предположениями.

Сначала полностью прочитай, в указанном порядке:

1. `AGENTS.md`.
1. `.planning/STATE.md`.
1. `docs/README.md`.
1. `.planning/HANDOFF.json`.
1. `.planning/phases/143-calendar-rsvp-core/.continue-here.md`.
1. `.planning/phases/143-calendar-rsvp-core/DISPATCH-INDEX.md`, `IMPLEMENTATION-143-03D.md`, `IMPLEMENTATION-143-04L.md`, `RUNTIME-BLOCKER.md`.
1. `.planning/phases/144-mailbox-domain/DISPATCH-INDEX.md`, `PLAN.md`, `REVIEW-144-0.1.8-SOURCE-ONLY-2026-08-13.md`, `REVIEW-144-02C-SCHEMA-READINESS.md`.
1. `.planning/decisions/gray-zones/HUB-FOUNDER-ALL-A-2026-08-13.md`.
1. `.planning/decisions/specs/MAIL-SPEC.md`, `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`, `docs/mail/openapi.yaml`, `docs/mail/CONTRACT-CHANGELOG.md`, `docs/superpowers/plans/2026-08-13-mail-engine-0.1.8.md`.
1. `C:\Users\Valerii\.codex\gsd-core\workflows\pause-work.md`.

Не читай и не меняй `pro-qa-ai/**`.

## Зафиксированная фактическая картина

Текущие принятые пакеты Phase 143 имеют только статический статус:

- `143-01`: `e7d5f450`.
- `143-02A`: через `d73a9380`.
- `143-02B`: через `efa7814d`.
- `143-03A`: `31bfff8f58ba83d12b59600135503f98a6c92fbd`.
- `143-03B`: `STATIC_ACCEPTED` на чистом `dd346e31`, исходная граница `61c19f9f`.
- `143-03C`: `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`.
- `143-03D`: `10b4ceead119cb74a16b325b33f9121f60d7b1b5`, финальный independent C/I/M `0/0/0`.
- `143-03F`: `846cd00ecb243402877e18239c89a62bf07e8d14`.
- `143-04L`: `4525c559da22729aa8bc51ff5336c29dad4f27e9`, финальный Terra/high C/I/M `0/0/0`.

`143-03E` — `BLOCKED/NOT_RUN`: требуются и accepted launcher, и отдельный founder-approved production migration mechanism/gate. `143-04A` заблокирован принятым `03E`; зависимость `04L` уже удовлетворена. Ни Phase 143, ни Mail, ни продукт, ни migration, ни runtime не приняты.

Все executable tests/build/lint/generation/migrations/runtime — `NOT_RUN`. Docker может быть запущен внешне, но это не проверено и не заменяет task-owned launcher. Launcher — `NOT_READY`. Не запускай тесты, сборку, линтеры, генерацию, валидаторы, сервисы на Windows host и не используй прямой Docker/Compose как обход. Не меняй этот статус без допустимого current-SHA результата официального launcher.

## Founder и маршрутизация

Founder выбрал Option A для всех 209 вопросов; каноническая запись `.planning/decisions/gray-zones/HUB-FOUNDER-ALL-A-2026-08-13.md` в commit `628f1dea`. Не задавай эти 209 вопросов снова. Только FQ11 и FQ29 получили прямой ответ; all-A — продуктовая, а не blanket technical/runtime/legal/migration/execution acceptance.

Только root применяет `gpt-5.6-sol/max`. Анализ, декомпозиция, документация и review используют `gpt-5.6-terra` с task-routed effort. Для этой ветки и Phases 143–148 atomic production/test-code пакеты используют строго `gpt-5.6-terra/xhigh`, без fallback; сохраняй действующее правило `dispatcher-enforced` provenance. При доступной telemetry она должна совпасть; mismatch, rejection, пропущенные model/effort или fallback означают `BLOCKED`. Если telemetry не выдана, фиксируй `UNKNOWN`, не выдумывай PASS. Твоя роль root — оркестрация, preflight, review, разрешение этапов и точечная фиксация; не обходи frozen contract.

## Точный следующий шаг — до любого кода

Независимый readiness audit определил единственный лучший следующий шаг: docs-only contract amendment, устраняющий FILE-05 versus Mail 0.1.8. После preflight измени согласованно только следующие документы:

1. `.planning/decisions/specs/MAIL-SPEC.md`.
1. `docs/mail/PERSISTENCE-CONTRACT-0.1.8.md`.
1. `docs/mail/openapi.yaml`.
1. `docs/mail/CONTRACT-CHANGELOG.md`.
1. `docs/superpowers/plans/2026-08-13-mail-engine-0.1.8.md`.

Зафиксируй однозначно:

- при создании `MailboxPolicy`: `max_message_bytes=25_000_000` и `max_attachment_bytes=25_000_000`, именно decimal bytes;
- hard ceilings остаются: message `52_428_800`, attachment `26_214_400`;
- администратор может менять policy только внутри соответствующих bounds;
- product defaults не равны и не подменяют adapter caps.

Так как plan входит в эти пять exact paths, обнови его устаревший pin `c0f20498` внутри текущего пакета. Остальные известные stale-записи — telemetry blocker у `144-01B`, dependency у `144-03D` и maxima wording в `REVIEW-144-02C-SCHEMA-READINESS.md` — являются read-time non-authoritative notes. Не редактируй их в этом пятидокументном пакете и не расширяй его scope; поставь их в очередь отдельного exact-path docs status-sync только после independent `PASS` по пятидокументному кандидату либо как отдельный independently reviewed packet.

Не изобретай дополнительные лимиты, migration или launcher. После amendment закажи независимый exact-candidate review строго `gpt-5.6-terra/high` с явными model и effort. Только `PASS` этого review может авторизовать один точный кодовый пакет `144-01A`. До PASS не изменяй production/test code для 144-01A.

После правок выполни только допустимые статические проверки для изменённых путей (включая `git diff --check`), сделай независимый review и затем действуй по его настоящему вердикту. Коммит допускается только после требуемого review и только с exact-path staging; никогда не используй `git add -A`, не добавляй AI attribution/`Co-authored-by`, не пушь/мерджь/создавай PR без отдельной авторизации.

## Критические запреты и определения

- Для JSON schema checks нужны оба условия: `utf8.Valid` и `json.Valid`.
- Иммутабельность не допускает exported или aliased mutable references; для envelope допустимы только snapshot и copy-returning значения/bytes.
- При проверке хэшей Git blob bytes каноничны; CRLF делает filesystem hash непригодным как acceptance evidence.
- Founder product approval не равен технической, runtime, legal, migration или production acceptance.
- Не трогай protected zone `pro-qa-ai/**`.
- Не перезаписывай и не откатывай чужие изменения; при overlap остановись и сообщи.

Заверши каждый этап честным статусом: отдельно static acceptance и отдельно `NOT_READY`/`NOT_RUN`. При любом mismatch, отсутствующем документе, конфликте границ, неявном fallback или ином blocker сообщи его с точными фактами и не импровизируй.
> **144-03B Folder post-source lifecycle (2026-08-14):** historical subordinate to `.continue-here.md`. Source/oracle `f8cb7643986fc591cae16764575337b72ef3a692` is `STATIC_ACCEPTED`, static-only, after Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Compile-shape Terra/high `EXACT_SHA_PASS` reviews, C/I/M `0/0/0`. The 20-path docs candidate awaits its own three reviews; then root only freezes/adjudicates `144-03B-Rule`. Launcher `NOT_READY`, executable evidence `NOT_RUN`; 04B reparent, 144-01B, Phase 142 and Phase 148 remain blocked. Author Terra/medium, dispatcher-enforced, no fallback, telemetry `UNKNOWN`.
