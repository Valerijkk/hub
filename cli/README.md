# hub-cli — CLI для интеграции Hub с CI/CD

Утилита командной строки для пайплайнов: загрузка результатов Allure в TMS, создание багов и обновление задач в трекере Hub.

## Установка

Из корня репозитория:

```bash
cd cli
go build -o hub-cli .
# или go install ./...
```

Бинарник: `hub-cli` (Windows: `hub-cli.exe`).

## Авторизация

API-ключ выдаётся в настройках Hub (или админом). Три способа передать ключ:

1. **Конфиг** (рекомендуется для CI после однократной настройки агента):

```bash
hub-cli auth --token YOUR_API_KEY
# сохраняет ~/.hub/config.json (права 0600)
```

2. **Переменная окружения:**

```bash
export HUB_API_KEY=YOUR_API_KEY
```

3. **Флаг каждой команды:**

```bash
hub-cli --api-key YOUR_API_KEY ...
```

Базовый URL API (по умолчанию `http://localhost:3000`):

```bash
export HUB_API_URL=https://your-hub.example.com
# или
hub-cli --api-url https://your-hub.example.com ...
```

## Глобальные флаги

| Флаг | Env | Описание |
|------|-----|----------|
| `--api-url` | `HUB_API_URL` | URL сервера Hub |
| `--api-key` | `HUB_API_KEY` | API-ключ |
| `--json` | — | Вывод ответа в JSON |

## Команды

### `auth`

Сохраняет API-ключ в `~/.hub/config.json`.

```bash
hub-cli auth --token YOUR_API_KEY
hub-cli auth --token YOUR_API_KEY --api-url https://hub.example.com
```

### `push-results`

Загружает JSON-файлы Allure из директории в TMS проекта.

```bash
hub-cli push-results --project HUB --file ./allure-results/
hub-cli push-results --project HUB --run-id "Build-42" --file ./allure-results/
```

| Флаг | Обязательный | Описание |
|------|--------------|----------|
| `--project` | да | Ключ проекта (например, `HUB`) |
| `--file` | да | Путь к каталогу с `*.json` |
| `--run-id` / `--run` | нет | ID прогона для привязки к существующему run |

### `create-bug`

Создаёт задачу типа `bug` в трекере.

```bash
hub-cli create-bug --project HUB --title "Кнопка не работает" --priority high
hub-cli create-bug --project HUB --title "Ошибка API" \
  --description "Шаги воспроизведения..." \
  --assignee dev@example.com \
  --labels "regression,api"
```

Приоритеты: `critical`, `high`, `medium`, `low`, `trivial` (по умолчанию `medium`).

### `update-task`

Обновляет задачу по ключу (`HUB-123`). Проект определяется из ключа автоматически.

```bash
hub-cli update-task HUB-123 --status done
hub-cli update-task HUB-123 --status in_progress --priority high
hub-cli update-task HUB-123 --assignee dev@example.com --comment "Исправлено в CI"
```

## Пример GitHub Actions

```yaml
- name: Upload Allure to Hub TMS
  env:
    HUB_API_URL: https://hub.example.com
    HUB_API_KEY: ${{ secrets.HUB_API_KEY }}
  run: |
    hub-cli push-results --project HUB --file ./allure-results/
```

## Код

| Файл | Содержание |
|------|------------|
| `main.go` | Точка входа |
| `cmd/root.go` | Корневая команда, HTTP-клиент, глобальные флаги |
| `cmd/auth.go` | Сохранение конфига |
| `cmd/push_results.go` | Импорт Allure |
| `cmd/create_bug.go` | Создание бага |
| `cmd/update_task.go` | Обновление задачи |

Зависимости: [Cobra](https://github.com/spf13/cobra). Модуль: `github.com/valerii/hub/cli`.

## Ограничения

- Таймаут HTTP-запросов: 30 с.
- Авторизация: заголовок `X-API-Key` (не OAuth/SSO).
- При ошибке API код выхода — `1`, сообщение в stderr.

API-контракты: `docs/system_specification.md`, `docs/wiki/API.md`.
