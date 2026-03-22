# API Reference

Hub предоставляет REST API с ~230 endpoints. Все запросы идут через `/api/` префикс.

## Аутентификация

### JWT Bearer Token
```
Authorization: Bearer <access_token>
```

### API Key
```
X-API-Key: <api_key>
```

## Основные модули

### Auth (`/api/auth/`)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/auth/magic-link` | Отправить magic link код |
| POST | `/auth/magic-link/verify` | Верифицировать код |
| POST | `/auth/telegram` | Вход через Telegram |
| POST | `/auth/refresh` | Обновить токены |
| POST | `/auth/logout` | Выход |
| GET | `/auth/me` | Текущий пользователь |
| PUT | `/auth/me` | Обновить профиль |
| GET | `/auth/sessions` | Список сессий |
| DELETE | `/auth/sessions/:id` | Завершить сессию |

### Projects (`/api/projects/`)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/projects/` | Создать проект |
| GET | `/projects/` | Список проектов |
| GET | `/projects/:id` | Получить проект |
| PUT | `/projects/:id` | Обновить проект |
| DELETE | `/projects/:id` | Удалить (в корзину) |
| POST | `/projects/:id/archive` | Архивировать |
| POST | `/projects/:id/members` | Добавить участника |
| POST | `/projects/:id/invites` | Создать приглашение |

### Issues (`/api/projects/:id/issues/`)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/issues/` | Создать задачу |
| GET | `/issues/` | Список задач (pagination) |
| GET | `/issues/:key` | Получить задачу по ключу |
| PUT | `/issues/:key` | Обновить задачу |
| DELETE | `/issues/:key` | Удалить задачу |
| PUT | `/issues/:key/position` | Переместить (Kanban) |
| POST | `/issues/jql` | JQL поиск |
| POST | `/issues/:key/comments` | Добавить комментарий |
| POST | `/issues/:key/time` | Записать время |
| POST | `/issues/:key/timer/start` | Старт таймера |

### Messenger (`/api/messenger/`)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/channels/` | Создать канал |
| GET | `/channels/` | Мои каналы |
| POST | `/channels/:id/messages/` | Отправить сообщение |
| GET | `/channels/:id/messages/` | Список сообщений |
| POST | `/dm` | Создать/открыть DM |
| GET | `/presence` | Онлайн-статусы |
| GET | `/search` | Поиск сообщений |

### TMS (`/api/projects/:id/tms/`)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/cases/` | Создать тест-кейс |
| GET | `/cases/` | Список тест-кейсов |
| POST | `/runs/` | Создать прогон |
| POST | `/runs/:id/start` | Запустить прогон |
| POST | `/runs/:id/complete` | Завершить прогон |
| POST | `/allure/import` | Импорт Allure results |
| GET | `/reports/trends` | Тренды тестов |
| GET | `/reports/traceability` | Матрица трассировки |

### Calls (`/api/calls/`)
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/calls/` | Создать звонок |
| POST | `/calls/:id/join` | Присоединиться |
| POST | `/calls/:id/leave` | Покинуть |
| POST | `/calls/:id/end` | Завершить |
| POST | `/calls/:id/recording/start` | Начать запись |

### WebSocket (`/ws`)
```
ws://host/ws?token=<jwt>
```

## Формат ответа

Успешный:
```json
{
  "id": "uuid",
  "title": "Название",
  "created_at": "2026-03-22T12:00:00Z"
}
```

Ошибка:
```json
{
  "code": 400,
  "message": "Описание ошибки",
  "request_id": "req-uuid"
}
```

Пагинация:
```json
{
  "items": [...],
  "next_cursor": "HUB-10",
  "has_more": true,
  "total": 42
}
```
