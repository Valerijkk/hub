# Cursor-адаптер

## Зачем нужен адаптер

Файлы в `prompt-engineering/agents/` — портативные контракты ролей, а не
определения Cursor custom-agent. Копировать их «как есть» в `.cursor/agents/`
нельзя: Cursor обнаруживает один markdown-файл на custom agent и интерпретирует
YAML frontmatter (`name`, `description`, `model`, `readonly`).

Этот адаптер отделяет runtime metadata от core-поведения:

- `agents.map.yaml` — маппинг role→core и permissions
- `wrapper.template.md` — валидная форма Cursor wrapper
- core prompt — загружается из `prompt-engineering/agents/<role>.md`

Официальный формат: https://cursor.com/docs/subagents

## Политика моделей

1. Активный orchestrator/main run стартует на **текущей записи каталога**,
   отображаемой как **GPT-5.6 Sol Pro**.
2. Актуальный ID берите из Cursor model selector или `GET /v1/models`;
   не сохраняйте запомненный provider slug в core-файлах.
3. Сгенерированные custom-agent wrappers по умолчанию используют `model: inherit`.
4. Orchestrator может явно переопределить downstream run только после применения
   `playbooks/model-routing.md` к фактическому каталогу и логирования
   resolved ID/fallback.
5. Если запрошенная запись orchestrator недоступна или ограничена, применяйте
   залогированную fallback-политику; не утверждайте молча, что использован default.

`inherit` означает, что subagent использует модель родителя — это default Cursor.
Конкретный ID — runtime pin и может быть заблокирован plan/team restrictions.

## Установка

1. Держите `prompt-engineering/` в репозитории.
2. Для каждой выбранной записи в `agents.map.yaml` отрендерьте
   `wrapper.template.md` в `.cursor/agents/<name>.md`.
3. Замените каждый `{{...}}`; не копируйте шаблон неотрендеренным.
4. Убедитесь, что `core_file` обёртки существует и `readonly` совпадает с map.
5. Выберите/разрешите модель orchestrator, как описано выше.
6. Smoke-тест: одна readonly и одна write-capable роль с явным scope.

Обёртка намеренно указывает runtime читать и общий core-контракт, и файл роли.
У Cursor нет портативной markdown-директивы `include`, поэтому голого указателя
без этой инструкции недостаточно.

## Правила frontmatter

- `name`: lowercase kebab-case; уникальный и не совпадающий с built-in agent name.
- `description`: конкретная подсказка для делегирования; когда использовать роль.
- `model`: `inherit`, пока текущее решение по каталогу явно не закрепит ID.
- `readonly`: `true` для reviewers/auditors; `false` для implementation-ролей.

Не добавляйте provider-specific model IDs в core agents. Не считайте модель
доступной только потому, что она есть в этом репозитории или в старых логах.

Offline toolkit validator проверяет только статический контракт адаптера: unique
names, существующие core files, boolean permissions, `inherit`, readonly
reviewer/auditor роли и обязательные ссылки policy в wrapper. Он **не**
проверяет live catalog availability, account/team entitlements, provider,
processing region, retention, training use или data-policy approval. Это
остаются runtime gates и должны подтверждаться evidence из текущего окружения.

## Запуск по каталогу

- IDE: смотрите live model selector и team restrictions.
- API/SDK: используйте model list endpoint, возьмите возвращённый ID и передайте
  его parent run. Inline subagent definitions также могут использовать `inherit`.
- Фиксируйте snapshot и decision через `models/registry.template.yaml`.

Официальные ссылки:

- https://cursor.com/docs/subagents
- https://cursor.com/docs/cloud-agent/api/endpoints#list-models
- https://cursor.com/docs/sdk/typescript

## Smoke acceptance

- [ ] Cursor обнаруживает отрендеренный wrapper
- [ ] Агент читает `agents/README.md` и свой core role перед действиями
- [ ] Readonly-роль не может редактировать
- [ ] Write-роль остаётся в delegated files
- [ ] Фактические orchestrator/resolved downstream model IDs залогированы
- [ ] Fallbacks видны, не silent
